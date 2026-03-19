package apiscanner

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type DiscoveryResult struct {
	Scanner      string         `json:"scanner"`
	Target       string         `json:"target"`
	ScannedAt    time.Time      `json:"scannedAt"`
	Discovered   []Discovered   `json:"discovered"`
	Fingerprints []Fingerprint  `json:"fingerprints"`
	Summary      DiscoveryStats `json:"summary"`
}

type Discovered struct {
	CheckID    string `json:"checkId,omitempty"`
	Category   string `json:"category"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	StatusCode int    `json:"statusCode"`
}

type Fingerprint struct {
	Type         string `json:"type"`
	Value        string `json:"value"`
	SourceHeader string `json:"sourceHeader"`
}

type DiscoveryStats struct {
	TotalPaths   int `json:"totalPaths"`
	WellKnown    int `json:"wellKnown"`
	ExposedFiles int `json:"exposedFiles"`
	Fingerprints int `json:"fingerprints"`
}

type DiscoverRequest struct {
	Target             string
	Headers            map[string]string
	Timeout            time.Duration
	InsecureSkipVerify bool
	MaxPaths           int
	IncludeIDs         []string
	ExcludeIDs         []string
}

type discoverPath struct {
	ID       string
	Category string
	Name     string
	Path     string
}

const (
	discoveryCategoryWellKnown   = "well-known"
	discoveryCategoryExposedFile = "exposed-file"
)

var wellKnownPaths = []discoverPath{
	{ID: "well-known.openapi", Category: discoveryCategoryWellKnown, Name: "OpenAPI", Path: "/openapi.json"},
	{ID: "well-known.openapi-well-known", Category: discoveryCategoryWellKnown, Name: "OpenAPI Well-Known", Path: "/.well-known/openapi.json"},
	{ID: "well-known.swagger", Category: discoveryCategoryWellKnown, Name: "Swagger", Path: "/swagger.json"},
	{ID: "well-known.swagger-ui", Category: discoveryCategoryWellKnown, Name: "Swagger UI", Path: "/swagger"},
	{ID: "well-known.graphql", Category: discoveryCategoryWellKnown, Name: "GraphQL", Path: "/graphql"},
	{ID: "well-known.api-graphql", Category: discoveryCategoryWellKnown, Name: "API GraphQL", Path: "/api/graphql"},
	{ID: "well-known.jwks", Category: discoveryCategoryWellKnown, Name: "JWKS", Path: "/.well-known/jwks.json"},
	{ID: "well-known.health", Category: discoveryCategoryWellKnown, Name: "Health", Path: "/health"},
	{ID: "well-known.healthz", Category: discoveryCategoryWellKnown, Name: "Healthz", Path: "/healthz"},
	{ID: "well-known.readyz", Category: discoveryCategoryWellKnown, Name: "Readyz", Path: "/readyz"},
}

var exposedPaths = []discoverPath{
	{ID: "exposed-file.dotenv", Category: discoveryCategoryExposedFile, Name: "DotEnv", Path: "/.env"},
	{ID: "exposed-file.dotenv-local", Category: discoveryCategoryExposedFile, Name: "DotEnv Local", Path: "/.env.local"},
	{ID: "exposed-file.git-config", Category: discoveryCategoryExposedFile, Name: "Git Config", Path: "/.git/config"},
	{ID: "exposed-file.docker-compose", Category: discoveryCategoryExposedFile, Name: "Docker Compose", Path: "/docker-compose.yml"},
	{ID: "exposed-file.config", Category: discoveryCategoryExposedFile, Name: "Config", Path: "/config.php"},
	{ID: "exposed-file.backup", Category: discoveryCategoryExposedFile, Name: "Backup", Path: "/backup.zip"},
	{ID: "exposed-file.actuator-env", Category: discoveryCategoryExposedFile, Name: "Actuator Env", Path: "/actuator/env"},
}

func supportedAPIDiscoveryCheckIDs() map[string]struct{} {
	supported := make(map[string]struct{}, len(wellKnownPaths)+len(exposedPaths)+1)
	for _, item := range wellKnownPaths {
		supported[normalizeDiscoveryID(item.ID)] = struct{}{}
	}
	for _, item := range exposedPaths {
		supported[normalizeDiscoveryID(item.ID)] = struct{}{}
	}
	supported[discoveryCheckFingerprintHeaders] = struct{}{}
	return supported
}

func DiscoverAPI(ctx context.Context, req DiscoverRequest) (DiscoveryResult, error) {
	normalized, err := normalizeDiscoverRequest(req)
	if err != nil {
		return DiscoveryResult{}, err
	}
	filter := newDiscoveryCheckFilter(normalized.IncludeIDs, normalized.ExcludeIDs)

	client := newDiscoveryClient(normalized.Timeout, normalized.InsecureSkipVerify)
	targetURL, err := url.Parse(normalized.Target)
	if err != nil {
		return DiscoveryResult{}, fmt.Errorf("invalid target URL: %w", err)
	}

	paths := buildDiscoveryPathList(normalized.MaxPaths)
	discovered := make([]Discovered, 0)

	for _, candidate := range paths {
		if !filter.Allows(candidate.ID) {
			continue
		}

		resolved := targetURL.ResolveReference(&url.URL{Path: candidate.Path})
		statusCode, probeErr := probeDiscoveryPath(ctx, client, resolved.String(), normalized.Headers)
		if probeErr != nil {
			continue
		}

		if statusCode == http.StatusNotFound {
			continue
		}

		discovered = append(discovered, Discovered{
			CheckID:    candidate.ID,
			Category:   candidate.Category,
			Name:       candidate.Name,
			URL:        resolved.String(),
			StatusCode: statusCode,
		})
	}

	fingerprints := make([]Fingerprint, 0)
	if filter.Allows(discoveryCheckFingerprintHeaders) {
		fingerprints, _ = collectFingerprints(ctx, client, normalized.Target, normalized.Headers)
	}
	stats := buildDiscoveryStats(discovered, fingerprints)

	result := DiscoveryResult{
		Scanner:      "api-discovery",
		Target:       normalized.Target,
		ScannedAt:    time.Now().UTC(),
		Discovered:   discovered,
		Fingerprints: fingerprints,
		Summary:      stats,
	}

	return result, nil
}

func normalizeDiscoverRequest(req DiscoverRequest) (DiscoverRequest, error) {
	clone := req
	clone.Target = strings.TrimSpace(clone.Target)
	if clone.Target == "" {
		return DiscoverRequest{}, fmt.Errorf("discovery target is required")
	}
	if _, err := url.ParseRequestURI(clone.Target); err != nil {
		return DiscoverRequest{}, fmt.Errorf("invalid discovery target URL: %w", err)
	}

	if clone.Headers == nil {
		clone.Headers = map[string]string{}
	}
	if clone.Timeout <= 0 {
		clone.Timeout = 20 * time.Second
	}
	if clone.MaxPaths <= 0 {
		clone.MaxPaths = 64
	}
	clone.IncludeIDs = normalizeDiscoveryIDs(clone.IncludeIDs)
	clone.ExcludeIDs = normalizeDiscoveryIDs(clone.ExcludeIDs)
	if err := validateDiscoveryIDSelection(clone.IncludeIDs, clone.ExcludeIDs, supportedAPIDiscoveryCheckIDs()); err != nil {
		return DiscoverRequest{}, err
	}

	return clone, nil
}

func newDiscoveryClient(timeout time.Duration, insecureSkipVerify bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify} //nolint:gosec

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func buildDiscoveryPathList(maxPaths int) []discoverPath {
	all := make([]discoverPath, 0, len(wellKnownPaths)+len(exposedPaths))
	all = append(all, wellKnownPaths...)
	all = append(all, exposedPaths...)

	if maxPaths <= 0 || len(all) <= maxPaths {
		return all
	}
	return all[:maxPaths]
}

func probeDiscoveryPath(ctx context.Context, client *http.Client, path string, headers map[string]string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 8*1024))

	return resp.StatusCode, nil
}

func collectFingerprints(ctx context.Context, client *http.Client, target string, headers map[string]string) ([]Fingerprint, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 8*1024))

	fingerprints := make([]Fingerprint, 0)
	fingerprints = appendIfHeader(fingerprints, resp.Header, "Server", "server")
	fingerprints = appendIfHeader(fingerprints, resp.Header, "X-Powered-By", "framework")
	fingerprints = appendIfHeader(fingerprints, resp.Header, "X-AspNet-Version", "framework")
	fingerprints = appendIfHeader(fingerprints, resp.Header, "Via", "proxy")
	fingerprints = appendIfHeader(fingerprints, resp.Header, "CF-Ray", "cdn")

	if cookies := resp.Header.Values("Set-Cookie"); len(cookies) > 0 {
		for _, cookie := range cookies {
			trimmed := strings.TrimSpace(cookie)
			if trimmed == "" {
				continue
			}
			value := trimmed
			if len(value) > 96 {
				value = value[:96] + "..."
			}
			fingerprints = append(fingerprints, Fingerprint{
				Type:         "cookie",
				Value:        value,
				SourceHeader: "Set-Cookie",
			})
			break
		}
	}

	return dedupeFingerprints(fingerprints), nil
}

func appendIfHeader(existing []Fingerprint, headers http.Header, headerKey string, fpType string) []Fingerprint {
	value := strings.TrimSpace(headers.Get(headerKey))
	if value == "" {
		return existing
	}
	if len(value) > 96 {
		value = value[:96] + "..."
	}
	return append(existing, Fingerprint{Type: fpType, Value: value, SourceHeader: headerKey})
}

func dedupeFingerprints(values []Fingerprint) []Fingerprint {
	if len(values) <= 1 {
		return values
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]Fingerprint, 0, len(values))
	for _, item := range values {
		key := strings.ToLower(item.Type + "|" + item.Value + "|" + item.SourceHeader)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Type == result[j].Type {
			return result[i].SourceHeader < result[j].SourceHeader
		}
		return result[i].Type < result[j].Type
	})

	return result
}

func buildDiscoveryStats(discovered []Discovered, fingerprints []Fingerprint) DiscoveryStats {
	stats := DiscoveryStats{TotalPaths: len(discovered), Fingerprints: len(fingerprints)}
	for _, item := range discovered {
		switch item.Category {
		case discoveryCategoryWellKnown:
			stats.WellKnown++
		case discoveryCategoryExposedFile:
			stats.ExposedFiles++
		}
	}
	return stats
}

func RenderDiscoveryOutput(result DiscoveryResult, format OutputFormat) (string, error) {
	switch format {
	case FormatJSON:
		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal discovery output: %w", err)
		}
		return string(encoded), nil
	case FormatTable:
		return formatDiscoveryTable(result), nil
	default:
		return "", fmt.Errorf("unsupported output format %q", format)
	}
}

func formatDiscoveryTable(result DiscoveryResult) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	fmt.Fprintf(writer, "SCANNER:\t%s\n", result.Scanner)
	fmt.Fprintf(writer, "TARGET:\t%s\n", result.Target)
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "DISCOVERED PATHS")
	fmt.Fprintln(writer, "CHECK ID\tCATEGORY\tNAME\tSTATUS\tURL")
	fmt.Fprintln(writer, "--------\t--------\t----\t------\t---")
	for _, item := range result.Discovered {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%s\n", item.CheckID, item.Category, item.Name, item.StatusCode, item.URL)
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "FINGERPRINTS")
	fmt.Fprintln(writer, "TYPE\tVALUE\tSOURCE")
	fmt.Fprintln(writer, "----\t-----\t------")
	for _, fp := range result.Fingerprints {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", fp.Type, fp.Value, fp.SourceHeader)
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "SUMMARY")
	fmt.Fprintf(writer, "TOTAL PATHS\t%d\n", result.Summary.TotalPaths)
	fmt.Fprintf(writer, "WELL-KNOWN\t%d\n", result.Summary.WellKnown)
	fmt.Fprintf(writer, "EXPOSED FILES\t%d\n", result.Summary.ExposedFiles)
	fmt.Fprintf(writer, "FINGERPRINTS\t%d\n", result.Summary.Fingerprints)

	_ = writer.Flush()
	return strings.TrimRight(buffer.String(), "\n")
}
