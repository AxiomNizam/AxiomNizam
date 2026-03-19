package apiscanner

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type openAPISpec struct {
	Servers []struct {
		URL string `json:"url" yaml:"url"`
	} `json:"servers" yaml:"servers"`
	Paths map[string]map[string]openAPIOperation `json:"paths" yaml:"paths"`
}

type openAPIOperation struct {
	RequestBody struct {
		Content map[string]openAPIContent `json:"content" yaml:"content"`
	} `json:"requestBody" yaml:"requestBody"`
}

type openAPIContent struct {
	Example  interface{}                    `json:"example" yaml:"example"`
	Examples map[string]openAPIExampleValue `json:"examples" yaml:"examples"`
}

type openAPIExampleValue struct {
	Value interface{} `json:"value" yaml:"value"`
}

func LoadOpenAPIEndpoints(
	ctx context.Context,
	target string,
	baseURL string,
	headers map[string]string,
	timeout time.Duration,
	insecureSkipVerify bool,
) ([]Endpoint, error) {
	content, err := loadOpenAPIContent(ctx, target, headers, timeout, insecureSkipVerify)
	if err != nil {
		return nil, err
	}

	spec, err := parseOpenAPISpec(content)
	if err != nil {
		return nil, err
	}

	resolvedBase := resolveOpenAPIBaseURL(spec, target, baseURL)
	if resolvedBase == "" {
		return nil, fmt.Errorf("unable to resolve base URL from OpenAPI spec; pass --base-url")
	}

	endpoints, err := extractOpenAPIEndpoints(spec, resolvedBase)
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no scannable operations found in OpenAPI spec")
	}

	return endpoints, nil
}

func loadOpenAPIContent(
	ctx context.Context,
	target string,
	headers map[string]string,
	timeout time.Duration,
	insecureSkipVerify bool,
) ([]byte, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, fmt.Errorf("OpenAPI target is required")
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	if isHTTPURL(target) {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify} //nolint:gosec

		client := &http.Client{Timeout: timeout, Transport: transport}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAPI request: %w", err)
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch OpenAPI target: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("failed to fetch OpenAPI target: status=%d", resp.StatusCode)
		}

		payload, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
		if err != nil {
			return nil, fmt.Errorf("failed to read OpenAPI response body: %w", err)
		}
		return payload, nil
	}

	path := target
	if !filepath.IsAbs(path) {
		path = filepath.Clean(path)
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI file: %w", err)
	}
	return payload, nil
}

func parseOpenAPISpec(payload []byte) (openAPISpec, error) {
	var spec openAPISpec
	if err := json.Unmarshal(payload, &spec); err == nil && len(spec.Paths) > 0 {
		return spec, nil
	}

	if err := yaml.Unmarshal(payload, &spec); err != nil {
		return openAPISpec{}, fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}
	if len(spec.Paths) == 0 {
		return openAPISpec{}, fmt.Errorf("OpenAPI document does not contain paths")
	}

	return spec, nil
}

func resolveOpenAPIBaseURL(spec openAPISpec, target string, userBase string) string {
	if trimmed := strings.TrimSpace(userBase); trimmed != "" {
		return trimmed
	}

	for _, server := range spec.Servers {
		if strings.TrimSpace(server.URL) != "" {
			return strings.TrimSpace(server.URL)
		}
	}

	if parsed, err := url.Parse(target); err == nil {
		scheme := strings.ToLower(parsed.Scheme)
		if scheme == "http" || scheme == "https" {
			return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
		}
	}

	return ""
}

func extractOpenAPIEndpoints(spec openAPISpec, baseURL string) ([]Endpoint, error) {
	paths := make([]string, 0, len(spec.Paths))
	for path := range spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	endpoints := make([]Endpoint, 0)
	for _, path := range paths {
		operations := spec.Paths[path]
		methods := make([]string, 0, len(operations))
		for method := range operations {
			methods = append(methods, method)
		}
		sort.Strings(methods)

		for _, method := range methods {
			upperMethod := strings.ToUpper(strings.TrimSpace(method))
			if !isScannableMethod(upperMethod) {
				continue
			}

			operation := operations[method]
			resolvedURL, err := resolveEndpointURL(baseURL, path)
			if err != nil {
				return nil, err
			}

			endpoint := Endpoint{
				URL:     resolvedURL,
				Method:  upperMethod,
				Headers: map[string]string{},
			}

			if upperMethod == http.MethodPost || upperMethod == http.MethodPut || upperMethod == http.MethodPatch {
				if body, contentType := extractRequestExample(operation); body != "" {
					endpoint.Body = body
					if contentType != "" {
						endpoint.Headers[contentTypeHeader] = contentType
					}
				}
			}

			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints, nil
}

func resolveEndpointURL(baseURL string, path string) (string, error) {
	path = strings.TrimSpace(path)
	if isHTTPURL(path) {
		return path, nil
	}

	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("invalid base URL %q: %w", baseURL, err)
	}

	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid OpenAPI path %q: %w", path, err)
	}

	return base.ResolveReference(ref).String(), nil
}

func extractRequestExample(operation openAPIOperation) (string, string) {
	if len(operation.RequestBody.Content) == 0 {
		return "", ""
	}

	contentTypes := make([]string, 0, len(operation.RequestBody.Content))
	for contentType := range operation.RequestBody.Content {
		contentTypes = append(contentTypes, contentType)
	}
	sort.Strings(contentTypes)

	for _, contentType := range contentTypes {
		content := operation.RequestBody.Content[contentType]
		if encoded := encodeExampleValue(content.Example); encoded != "" {
			return encoded, contentType
		}

		exampleKeys := make([]string, 0, len(content.Examples))
		for key := range content.Examples {
			exampleKeys = append(exampleKeys, key)
		}
		sort.Strings(exampleKeys)

		for _, key := range exampleKeys {
			if encoded := encodeExampleValue(content.Examples[key].Value); encoded != "" {
				return encoded, contentType
			}
		}
	}

	return "", ""
}

func encodeExampleValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(encoded)
	}
}

func isScannableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func isHTTPURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed == nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	return scheme == "http" || scheme == "https"
}
