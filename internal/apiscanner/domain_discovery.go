package apiscanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

const discoveryCategoryAPIHint = "api-hint"

var commonSubdomainPrefixes = []string{
	"api",
	"www",
	"dev",
	"staging",
	"test",
	"sandbox",
	"gateway",
	"graphql",
	"auth",
	"admin",
	"internal",
	"edge",
	"proxy",
	"mobile",
	"partner",
	"public",
	"v1",
	"v2",
	"rest",
	"backend",
	"service",
	"platform",
	"demo",
}

var domainHintPaths = []discoverPath{
	{ID: "hint.openapi", Category: discoveryCategoryAPIHint, Name: "OpenAPI", Path: "/openapi.json"},
	{ID: "hint.swagger", Category: discoveryCategoryAPIHint, Name: "Swagger", Path: "/swagger.json"},
	{ID: "hint.graphql", Category: discoveryCategoryAPIHint, Name: "GraphQL", Path: "/graphql"},
	{ID: "hint.health", Category: discoveryCategoryAPIHint, Name: "Health", Path: "/health"},
	{ID: "hint.ready", Category: discoveryCategoryAPIHint, Name: "Ready", Path: "/ready"},
	{ID: "hint.actuator", Category: discoveryCategoryAPIHint, Name: "Actuator", Path: "/actuator"},
	{ID: "hint.versioned-api", Category: discoveryCategoryAPIHint, Name: "Versioned API", Path: "/v1"},
	{ID: "hint.api-root", Category: discoveryCategoryAPIHint, Name: "API Root", Path: "/api"},
}

type DiscoverDomainRequest struct {
	Target             string
	Headers            map[string]string
	Timeout            time.Duration
	InsecureSkipVerify bool
	MaxSubdomains      int
	MaxHints           int
	Schemes            []string
	IncludeIDs         []string
	ExcludeIDs         []string
}

func supportedDomainDiscoveryCheckIDs() map[string]struct{} {
	supported := make(map[string]struct{}, len(domainHintPaths)+1)
	for _, item := range domainHintPaths {
		supported[normalizeDiscoveryID(item.ID)] = struct{}{}
	}
	supported[discoveryCheckFingerprintHeaders] = struct{}{}
	return supported
}

type ResolvedSubdomain struct {
	Name      string   `json:"name"`
	Host      string   `json:"host"`
	Addresses []string `json:"addresses"`
}

type DomainDiscoveryStats struct {
	Candidates   int `json:"candidates"`
	Resolved     int `json:"resolved"`
	APIHints     int `json:"apiHints"`
	Fingerprints int `json:"fingerprints"`
}

type DomainDiscoveryResult struct {
	Scanner      string               `json:"scanner"`
	Target       string               `json:"target"`
	ScannedAt    time.Time            `json:"scannedAt"`
	Candidates   []string             `json:"candidates"`
	Resolved     []ResolvedSubdomain  `json:"resolved"`
	APIHints     []Discovered         `json:"apiHints"`
	Fingerprints []Fingerprint        `json:"fingerprints"`
	Summary      DomainDiscoveryStats `json:"summary"`
}

func DiscoverDomain(ctx context.Context, req DiscoverDomainRequest) (DomainDiscoveryResult, error) {
	normalized, err := normalizeDiscoverDomainRequest(req)
	if err != nil {
		return DomainDiscoveryResult{}, err
	}

	candidates := buildSubdomainCandidates(normalized.Target, normalized.MaxSubdomains)
	resolved := resolveCandidateHosts(ctx, candidates, normalized.Timeout)

	hints, fingerprints := probeDomainAPIHints(ctx, normalized, resolved)
	stats := DomainDiscoveryStats{
		Candidates:   len(candidates),
		Resolved:     len(resolved),
		APIHints:     len(hints),
		Fingerprints: len(fingerprints),
	}

	result := DomainDiscoveryResult{
		Scanner:      "domain-discovery",
		Target:       normalized.Target,
		ScannedAt:    time.Now().UTC(),
		Candidates:   candidates,
		Resolved:     resolved,
		APIHints:     hints,
		Fingerprints: fingerprints,
		Summary:      stats,
	}

	return result, nil
}

func normalizeDiscoverDomainRequest(req DiscoverDomainRequest) (DiscoverDomainRequest, error) {
	clone := req

	target, err := normalizeDomainTarget(clone.Target)
	if err != nil {
		return DiscoverDomainRequest{}, err
	}
	clone.Target = target

	if clone.Headers == nil {
		clone.Headers = map[string]string{}
	}
	if clone.Timeout <= 0 {
		clone.Timeout = 20 * time.Second
	}
	if clone.MaxSubdomains <= 0 {
		clone.MaxSubdomains = 32
	}
	if clone.MaxHints <= 0 {
		clone.MaxHints = 48
	}

	schemes, err := parseDiscoverySchemes(clone.Schemes)
	if err != nil {
		return DiscoverDomainRequest{}, err
	}
	clone.Schemes = schemes
	clone.IncludeIDs = normalizeDiscoveryIDs(clone.IncludeIDs)
	clone.ExcludeIDs = normalizeDiscoveryIDs(clone.ExcludeIDs)
	if err := validateDiscoveryIDSelection(clone.IncludeIDs, clone.ExcludeIDs, supportedDomainDiscoveryCheckIDs()); err != nil {
		return DiscoverDomainRequest{}, err
	}

	return clone, nil
}

func normalizeDomainTarget(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("discovery target is required")
	}

	host, err := extractTargetHost(trimmed)
	if err != nil {
		return "", err
	}

	host = strings.TrimSpace(strings.Trim(host, "."))
	if host == "" {
		return "", fmt.Errorf("discovery target is required")
	}

	if ip := net.ParseIP(host); ip != nil {
		return host, nil
	}
	if !isValidDomainTarget(host) {
		return "", fmt.Errorf("invalid domain target %q", host)
	}

	return strings.ToLower(host), nil
}

func extractTargetHost(value string) (string, error) {
	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("invalid discovery target: %w", err)
		}
		return parsed.Hostname(), nil
	}

	host := strings.Split(value, "/")[0]
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost, nil
	}

	return host, nil
}

func isValidDomainTarget(value string) bool {
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			continue
		}
		return false
	}

	return true
}

func parseDiscoverySchemes(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return []string{"https", "http"}, nil
	}

	seen := map[string]struct{}{}
	schemes := make([]string, 0, len(raw))
	for _, item := range raw {
		for _, part := range strings.Split(item, ",") {
			s := strings.ToLower(strings.TrimSpace(part))
			if s == "" {
				continue
			}
			if s != "http" && s != "https" {
				return nil, fmt.Errorf("unsupported scheme %q (allowed: http, https)", s)
			}
			if _, exists := seen[s]; exists {
				continue
			}
			seen[s] = struct{}{}
			schemes = append(schemes, s)
		}
	}

	if len(schemes) == 0 {
		return []string{"https", "http"}, nil
	}

	return schemes, nil
}

func buildSubdomainCandidates(domain string, maxSubdomains int) []string {
	result := make([]string, 0, len(commonSubdomainPrefixes)+1)
	seen := map[string]struct{}{}
	add := func(value string) {
		v := strings.ToLower(strings.TrimSpace(value))
		if v == "" {
			return
		}
		if _, exists := seen[v]; exists {
			return
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}

	add(domain)
	for _, prefix := range commonSubdomainPrefixes {
		add(prefix + "." + domain)
		if maxSubdomains > 0 && len(result) >= maxSubdomains {
			break
		}
	}

	if maxSubdomains > 0 && len(result) > maxSubdomains {
		return result[:maxSubdomains]
	}
	return result
}

func resolveCandidateHosts(ctx context.Context, candidates []string, timeout time.Duration) []ResolvedSubdomain {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	resolver := net.DefaultResolver
	perLookupTimeout := timeout
	if perLookupTimeout > 5*time.Second {
		perLookupTimeout = 5 * time.Second
	}

	resolved := make([]ResolvedSubdomain, 0)
	for _, candidate := range candidates {
		lookupCtx, cancel := context.WithTimeout(ctx, perLookupTimeout)
		addresses, err := resolver.LookupHost(lookupCtx, candidate)
		cancel()
		if err != nil || len(addresses) == 0 {
			continue
		}

		unique := normalizeLookupAddresses(addresses)
		if len(unique) == 0 {
			continue
		}

		resolved = append(resolved, ResolvedSubdomain{
			Name:      candidate,
			Host:      candidate,
			Addresses: unique,
		})
	}

	sort.SliceStable(resolved, func(i, j int) bool {
		return resolved[i].Host < resolved[j].Host
	})

	return resolved
}

func normalizeLookupAddresses(addresses []string) []string {
	addrSet := map[string]struct{}{}
	unique := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		trimmed := strings.TrimSpace(addr)
		if trimmed == "" {
			continue
		}
		if _, exists := addrSet[trimmed]; exists {
			continue
		}
		addrSet[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}
	sort.Strings(unique)
	return unique
}

func probeDomainAPIHints(ctx context.Context, req DiscoverDomainRequest, resolved []ResolvedSubdomain) ([]Discovered, []Fingerprint) {
	filter := newDiscoveryCheckFilter(req.IncludeIDs, req.ExcludeIDs)
	client := newDiscoveryClient(req.Timeout, req.InsecureSkipVerify)
	hints := make([]Discovered, 0)
	fingerprintList := make([]Fingerprint, 0)
	seenHint := map[string]struct{}{}

	for _, host := range resolved {
		for _, scheme := range req.Schemes {
			schemeHints, schemeFingerprints := probeHintsForHostScheme(ctx, client, req, filter, host.Host, scheme, req.MaxHints-len(hints))
			fingerprintList = append(fingerprintList, schemeFingerprints...)
			hints = appendUniqueHints(hints, schemeHints, seenHint, req.MaxHints)

			if len(hints) >= req.MaxHints {
				break
			}
		}
		if len(hints) >= req.MaxHints {
			break
		}
	}

	sort.SliceStable(hints, func(i, j int) bool {
		return hints[i].URL < hints[j].URL
	})

	return hints, dedupeFingerprints(fingerprintList)
}

func probeHintsForHostScheme(
	ctx context.Context,
	client *http.Client,
	req DiscoverDomainRequest,
	filter discoveryCheckFilter,
	host string,
	scheme string,
	remaining int,
) ([]Discovered, []Fingerprint) {
	if remaining <= 0 {
		return nil, nil
	}

	hints := make([]Discovered, 0)
	fingerprints := make([]Fingerprint, 0)
	base := (&url.URL{Scheme: scheme, Host: host}).String()

	if filter.Allows(discoveryCheckFingerprintHeaders) {
		if fps, err := collectFingerprints(ctx, client, base, req.Headers); err == nil && len(fps) > 0 {
			fingerprints = append(fingerprints, fps...)
		}
	}

	for _, hint := range domainHintPaths {
		if !filter.Allows(hint.ID) {
			continue
		}

		if len(hints) >= remaining {
			break
		}

		fullURL := (&url.URL{Scheme: scheme, Host: host, Path: hint.Path}).String()
		statusCode, err := probeDiscoveryPath(ctx, client, fullURL, req.Headers)
		if err != nil || statusCode == 0 || statusCode == http.StatusNotFound {
			continue
		}

		hints = append(hints, Discovered{
			CheckID:    hint.ID,
			Category:   discoveryCategoryAPIHint,
			Name:       hint.Name,
			URL:        fullURL,
			StatusCode: statusCode,
		})
	}

	return hints, fingerprints
}

func appendUniqueHints(existing []Discovered, incoming []Discovered, seen map[string]struct{}, max int) []Discovered {
	for _, item := range incoming {
		if max > 0 && len(existing) >= max {
			break
		}

		key := strings.ToLower(strings.TrimSpace(item.URL))
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		existing = append(existing, item)
	}

	return existing
}

func RenderDomainDiscoveryOutput(result DomainDiscoveryResult, format OutputFormat) (string, error) {
	switch format {
	case FormatJSON:
		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal domain discovery output: %w", err)
		}
		return string(encoded), nil
	case FormatTable:
		return formatDomainDiscoveryTable(result), nil
	default:
		return "", fmt.Errorf("unsupported output format %q", format)
	}
}

func formatDomainDiscoveryTable(result DomainDiscoveryResult) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	fmt.Fprintf(writer, "SCANNER:\t%s\n", result.Scanner)
	fmt.Fprintf(writer, "TARGET:\t%s\n", result.Target)
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "RESOLVED SUBDOMAINS")
	fmt.Fprintln(writer, "HOST\tADDRESSES")
	fmt.Fprintln(writer, "----\t---------")
	for _, item := range result.Resolved {
		fmt.Fprintf(writer, "%s\t%s\n", item.Host, strings.Join(item.Addresses, ", "))
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "API HINTS")
	fmt.Fprintln(writer, "CHECK ID\tNAME\tSTATUS\tURL")
	fmt.Fprintln(writer, "--------\t----\t------\t---")
	for _, hint := range result.APIHints {
		fmt.Fprintf(writer, "%s\t%s\t%d\t%s\n", hint.CheckID, hint.Name, hint.StatusCode, hint.URL)
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
	fmt.Fprintf(writer, "CANDIDATES\t%d\n", result.Summary.Candidates)
	fmt.Fprintf(writer, "RESOLVED\t%d\n", result.Summary.Resolved)
	fmt.Fprintf(writer, "API HINTS\t%d\n", result.Summary.APIHints)
	fmt.Fprintf(writer, "FINGERPRINTS\t%d\n", result.Summary.Fingerprints)

	_ = writer.Flush()
	return strings.TrimRight(buffer.String(), "\n")
}
