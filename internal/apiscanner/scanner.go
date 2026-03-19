package apiscanner

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	contentTypeHeader = "Content-Type"
	maxReadBytes      = 1024 * 1024
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

type probeResult struct {
	statusCode int
	body       string
	headers    http.Header
	err        error
}

func (e *Engine) Scan(ctx context.Context, req ScanRequest) (ScanResult, error) {
	if e == nil {
		return ScanResult{}, fmt.Errorf("api scanner engine is not configured")
	}

	normalized, err := normalizeRequest(req)
	if err != nil {
		return ScanResult{}, err
	}

	client := newHTTPClient(normalized.Timeout, normalized.InsecureSkipVerify)

	baselineHeaders := cloneHeaderMap(normalized.Endpoint.Headers)
	if normalized.AuthHeader != "" && normalized.AuthValue != "" {
		baselineHeaders[normalized.AuthHeader] = normalized.AuthValue
	}

	baseline := e.probe(ctx, client, normalized, normalized.Endpoint.URL, normalized.Endpoint.Method, normalized.Endpoint.Body, baselineHeaders)

	findings := make([]Finding, 0)
	checks := make([]ScanCheckStatus, 0, 8)

	runCheck := func(id string, name string, fn func() ([]Finding, bool)) {
		checkFindings, executed := fn()
		findings = append(findings, checkFindings...)
		checks = append(checks, ScanCheckStatus{
			ID:       id,
			Name:     name,
			Executed: executed,
			Findings: len(checkFindings),
		})
	}

	runCheck(CheckSecurityHeaders, "Security Header Analysis", func() ([]Finding, bool) {
		return e.checkSecurityHeaders(ctx, client, normalized, baseline), true
	})
	runCheck(CheckHTTPMethod, "HTTP Method Validation", func() ([]Finding, bool) {
		return e.checkHTTPMethodValidation(ctx, client, normalized, baseline), true
	})
	runCheck(CheckAuthBypassDetection, "Authentication Bypass Detection", func() ([]Finding, bool) {
		return e.checkAuthBypassDetection(ctx, client, normalized, baseline)
	})
	runCheck(CheckAuthBypassTesting, "Authentication Bypass Testing", func() ([]Finding, bool) {
		return e.checkAuthBypassTesting(ctx, client, normalized, baseline)
	})
	runCheck(CheckSQLInjection, "SQL Injection Vulnerabilities", func() ([]Finding, bool) {
		return e.checkSQLInjection(ctx, client, normalized, baseline), true
	})
	runCheck(CheckNoSQLInjection, "NoSQL Injection Vulnerabilities", func() ([]Finding, bool) {
		return e.checkNoSQLInjection(ctx, client, normalized, baseline), true
	})
	runCheck(CheckXSS, "Cross-Site Scripting (XSS) Vulnerabilities", func() ([]Finding, bool) {
		return e.checkXSS(ctx, client, normalized, baseline), true
	})
	runCheck(CheckParameterTampering, "Parameter Tampering Detection", func() ([]Finding, bool) {
		return e.checkParameterTampering(ctx, client, normalized, baseline), true
	})

	result := ScanResult{
		Scanner:   "api-scanner",
		Target:    normalized.Endpoint.URL,
		Method:    normalized.Endpoint.Method,
		ScannedAt: time.Now().UTC(),
		Findings:  findings,
		Checks:    checks,
		Summary:   buildSummary(findings),
	}

	return result, nil
}

func normalizeRequest(req ScanRequest) (ScanRequest, error) {
	clone := req
	clone.Endpoint.URL = strings.TrimSpace(clone.Endpoint.URL)
	if clone.Endpoint.URL == "" {
		return ScanRequest{}, fmt.Errorf("scan target is required")
	}
	if _, err := url.ParseRequestURI(clone.Endpoint.URL); err != nil {
		return ScanRequest{}, fmt.Errorf("invalid scan target URL: %w", err)
	}

	clone.Endpoint.Method = strings.ToUpper(strings.TrimSpace(clone.Endpoint.Method))
	if clone.Endpoint.Method == "" {
		clone.Endpoint.Method = http.MethodGet
	}

	if clone.Timeout <= 0 {
		clone.Timeout = 30 * time.Second
	}
	if clone.RetryCount < 0 {
		clone.RetryCount = 0
	}
	if clone.RetryBackoff <= 0 {
		clone.RetryBackoff = time.Second
	}
	if clone.AuthHeader == "" && clone.AuthValue != "" {
		clone.AuthHeader = "Authorization"
	}
	if clone.AuthValue == "" {
		if value, ok := lookupHeaderIgnoreCase(clone.Endpoint.Headers, "Authorization"); ok {
			clone.AuthHeader = "Authorization"
			clone.AuthValue = strings.TrimSpace(value)
		}
	}
	if clone.Format == "" {
		clone.Format = FormatTable
	}

	if clone.Endpoint.Headers == nil {
		clone.Endpoint.Headers = map[string]string{}
	}

	return clone, nil
}

func newHTTPClient(timeout time.Duration, insecureSkipVerify bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify} //nolint:gosec

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func (e *Engine) probe(
	ctx context.Context,
	client *http.Client,
	req ScanRequest,
	targetURL, method, body string,
	headers map[string]string,
) probeResult {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}

	attempts := req.RetryCount + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		result, err := e.probeOnce(ctx, client, targetURL, method, body, headers)
		if err == nil {
			return result
		}

		lastErr = err
		if waitErr := waitBeforeRetry(ctx, req.RetryBackoff, attempt, attempts); waitErr != nil {
			return probeResult{err: waitErr}
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("request failed")
	}

	return probeResult{err: lastErr}
}

func (e *Engine) probeOnce(
	ctx context.Context,
	client *http.Client,
	targetURL, method, body string,
	headers map[string]string,
) (probeResult, error) {
	reqBody := io.Reader(nil)
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, targetURL, reqBody)
	if err != nil {
		return probeResult{}, err
	}
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return probeResult{}, err
	}
	defer resp.Body.Close()

	respBody, err := readBody(resp)
	if err != nil {
		return probeResult{}, err
	}

	return probeResult{
		statusCode: resp.StatusCode,
		body:       respBody,
		headers:    resp.Header.Clone(),
	}, nil
}

func waitBeforeRetry(ctx context.Context, base time.Duration, attempt, attempts int) error {
	if attempt >= attempts {
		return nil
	}

	delay := exponentialBackoff(base, attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func readBody(resp *http.Response) (string, error) {
	if resp == nil || resp.Body == nil {
		return "", nil
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxReadBytes))
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func (e *Engine) checkSecurityHeaders(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) []Finding {
	if baseline.err != nil {
		baseline = e.probe(ctx, client, req, req.Endpoint.URL, req.Endpoint.Method, req.Endpoint.Body, req.Endpoint.Headers)
	}
	if baseline.err != nil {
		return nil
	}

	expected := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Content-Security-Policy",
		"Referrer-Policy",
	}
	if strings.HasPrefix(strings.ToLower(req.Endpoint.URL), "https://") {
		expected = append(expected, "Strict-Transport-Security")
	}

	missing := make([]string, 0)
	for _, header := range expected {
		if strings.TrimSpace(baseline.headers.Get(header)) == "" {
			missing = append(missing, header)
		}
	}

	findings := make([]Finding, 0)
	if len(missing) > 0 {
		findings = append(findings, Finding{
			Type:        VulnSecurityHeaders,
			Severity:    SeverityMedium,
			Title:       "Missing recommended security headers",
			Description: "One or more recommended headers are missing from the API response.",
			Endpoint:    req.Endpoint.URL,
			Method:      req.Endpoint.Method,
			Evidence:    strings.Join(missing, ", "),
			Recommendation: "Add missing headers such as CSP, X-Frame-Options, and X-Content-Type-Options " +
				"at the API gateway or application layer.",
		})
	}

	insecureHeaders := []string{"Server", "X-Powered-By"}
	presentInsecure := make([]string, 0)
	for _, header := range insecureHeaders {
		if strings.TrimSpace(baseline.headers.Get(header)) != "" {
			presentInsecure = append(presentInsecure, header)
		}
	}
	if len(presentInsecure) > 0 {
		findings = append(findings, Finding{
			Type:           VulnSecurityHeaders,
			Severity:       SeverityLow,
			Title:          "Potential information disclosure headers present",
			Description:    "Server fingerprint headers may expose implementation details.",
			Endpoint:       req.Endpoint.URL,
			Method:         req.Endpoint.Method,
			Evidence:       strings.Join(presentInsecure, ", "),
			Recommendation: "Suppress or generalize server-identifying headers in production.",
		})
	}

	return findings
}

func (e *Engine) checkHTTPMethodValidation(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) []Finding {
	methods := []string{http.MethodTrace, "TRACK", http.MethodConnect}
	findings := make([]Finding, 0)

	for _, method := range methods {
		probe := e.probe(ctx, client, req, req.Endpoint.URL, method, "", req.Endpoint.Headers)
		if probe.err != nil {
			continue
		}
		if probe.statusCode < 400 {
			findings = append(findings, Finding{
				Type:           VulnHTTPMethod,
				Severity:       SeverityMedium,
				Title:          "Potentially unsafe HTTP method accepted",
				Description:    "Endpoint accepted a method that is usually disabled for APIs.",
				Endpoint:       req.Endpoint.URL,
				Method:         method,
				Evidence:       fmt.Sprintf("status=%d", probe.statusCode),
				Recommendation: "Restrict unsupported methods at application and gateway layers.",
			})
		}
	}

	return findings
}

func (e *Engine) checkAuthBypassDetection(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) ([]Finding, bool) {
	if !hasAuthContext(req) {
		return nil, false
	}
	if baseline.err != nil || baseline.statusCode >= 400 {
		return nil, false
	}

	findings := make([]Finding, 0)

	noAuthHeaders := cloneHeaderMap(req.Endpoint.Headers)
	removeHeaderIgnoreCase(noAuthHeaders, req.AuthHeader)
	probeNoAuth := e.probe(ctx, client, req, req.Endpoint.URL, req.Endpoint.Method, req.Endpoint.Body, noAuthHeaders)
	if probeNoAuth.err == nil && probeNoAuth.statusCode < 400 && responseLooksEquivalent(baseline, probeNoAuth) {
		findings = append(findings, Finding{
			Type:           VulnAuthBypass,
			Severity:       SeverityHigh,
			Title:          "Endpoint accessible without authentication",
			Description:    "Authenticated and unauthenticated responses appear equivalent.",
			Endpoint:       req.Endpoint.URL,
			Method:         req.Endpoint.Method,
			Evidence:       fmt.Sprintf("authenticated_status=%d unauthenticated_status=%d", baseline.statusCode, probeNoAuth.statusCode),
			Recommendation: "Enforce auth checks before business logic and return 401/403 for unauthenticated requests.",
		})
	}

	invalidAuthHeaders := cloneHeaderMap(req.Endpoint.Headers)
	invalidAuthHeaders[req.AuthHeader] = req.AuthValue + "-invalid"
	probeInvalid := e.probe(ctx, client, req, req.Endpoint.URL, req.Endpoint.Method, req.Endpoint.Body, invalidAuthHeaders)
	if probeInvalid.err == nil && probeInvalid.statusCode < 400 && responseLooksEquivalent(baseline, probeInvalid) {
		findings = append(findings, Finding{
			Type:           VulnAuthBypass,
			Severity:       SeverityHigh,
			Title:          "Endpoint accepts invalid authentication token",
			Description:    "Invalid auth credential returned a successful and equivalent response.",
			Endpoint:       req.Endpoint.URL,
			Method:         req.Endpoint.Method,
			Evidence:       fmt.Sprintf("authenticated_status=%d invalid_auth_status=%d", baseline.statusCode, probeInvalid.statusCode),
			Recommendation: "Validate token integrity/signature and reject malformed or modified credentials.",
		})
	}

	return findings, true
}

func (e *Engine) checkAuthBypassTesting(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) ([]Finding, bool) {
	if !hasAuthContext(req) {
		return nil, false
	}
	if baseline.err != nil || baseline.statusCode >= 400 {
		return nil, false
	}

	const loopback = "127.0.0.1"

	bypassHeaders := cloneHeaderMap(req.Endpoint.Headers)
	bypassHeaders["X-Forwarded-For"] = loopback
	bypassHeaders["X-Original-URL"] = req.Endpoint.URL
	bypassHeaders["X-Rewrite-URL"] = req.Endpoint.URL
	bypassHeaders["X-Originating-IP"] = loopback
	bypassHeaders["X-Remote-IP"] = loopback
	bypassHeaders["X-Forwarded-Host"] = "localhost"

	probeBypass := e.probe(ctx, client, req, req.Endpoint.URL, req.Endpoint.Method, req.Endpoint.Body, bypassHeaders)
	if probeBypass.err == nil && probeBypass.statusCode < 400 && responseLooksEquivalent(baseline, probeBypass) {
		return []Finding{{
			Type:           VulnAuthBypass,
			Severity:       SeverityHigh,
			Title:          "Endpoint appears vulnerable to header-based auth bypass",
			Description:    "Bypass-oriented headers produced a successful response equivalent to authenticated access.",
			Endpoint:       req.Endpoint.URL,
			Method:         req.Endpoint.Method,
			Evidence:       fmt.Sprintf("authenticated_status=%d bypass_status=%d", baseline.statusCode, probeBypass.statusCode),
			Recommendation: "Ignore spoofable forwarding headers for trust decisions unless set by a trusted proxy chain.",
		}}, true
	}

	return nil, true
}

func (e *Engine) checkSQLInjection(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) []Finding {
	findings := make([]Finding, 0)
	for _, payload := range defaultSQLPayloads {
		mutatedURL, mutatedBody, mutatedHeaders := applyMutation(req.Endpoint.URL, req.Endpoint.Body, req.Endpoint.Headers, payload)
		probe := e.probe(ctx, client, req, mutatedURL, req.Endpoint.Method, mutatedBody, mutatedHeaders)
		if probe.err != nil {
			continue
		}

		if indicatorsOfInjection(probe.body, sqlErrorSignatures) || isServerErrorDelta(baseline, probe) {
			findings = append(findings, Finding{
				Type:           VulnSQLInjection,
				Severity:       SeverityCritical,
				Title:          "Potential SQL injection vulnerability",
				Description:    "Response patterns indicate possible SQL query manipulation.",
				Endpoint:       req.Endpoint.URL,
				Method:         req.Endpoint.Method,
				Payload:        payload,
				Evidence:       buildEvidence(probe),
				Recommendation: "Use parameterized queries, strict validation, and central query builders.",
			})
			break
		}
	}

	return findings
}

func (e *Engine) checkNoSQLInjection(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) []Finding {
	findings := make([]Finding, 0)
	for _, payload := range defaultNoSQLPayloads {
		mutatedURL, mutatedBody, mutatedHeaders := applyMutationWithKey(req.Endpoint.URL, req.Endpoint.Body, req.Endpoint.Headers, "filter", payload)
		probe := e.probe(ctx, client, req, mutatedURL, req.Endpoint.Method, mutatedBody, mutatedHeaders)
		if probe.err != nil {
			continue
		}

		if indicatorsOfInjection(probe.body, noSQLErrorSignatures) || isServerErrorDelta(baseline, probe) {
			findings = append(findings, Finding{
				Type:           VulnNoSQLInjection,
				Severity:       SeverityCritical,
				Title:          "Potential NoSQL injection vulnerability",
				Description:    "NoSQL error signatures or anomaly patterns were detected.",
				Endpoint:       req.Endpoint.URL,
				Method:         req.Endpoint.Method,
				Payload:        payload,
				Evidence:       buildEvidence(probe),
				Recommendation: "Validate query operators strictly and sanitize document-based query input.",
			})
			break
		}
	}

	return findings
}

func (e *Engine) checkXSS(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) []Finding {
	findings := make([]Finding, 0)
	for _, payload := range defaultXSSPayloads {
		mutatedURL, mutatedBody, mutatedHeaders := applyMutation(req.Endpoint.URL, req.Endpoint.Body, req.Endpoint.Headers, payload)
		probe := e.probe(ctx, client, req, mutatedURL, req.Endpoint.Method, mutatedBody, mutatedHeaders)
		if probe.err != nil {
			continue
		}

		if strings.Contains(probe.body, payload) {
			findings = append(findings, Finding{
				Type:           VulnXSS,
				Severity:       SeverityHigh,
				Title:          "Potential reflected XSS vulnerability",
				Description:    "An injected payload appears reflected in the response body.",
				Endpoint:       req.Endpoint.URL,
				Method:         req.Endpoint.Method,
				Payload:        payload,
				Evidence:       buildEvidence(probe),
				Recommendation: "Apply output encoding and input validation for all user-controlled values.",
			})
			break
		}
	}

	return findings
}

func (e *Engine) checkParameterTampering(ctx context.Context, client *http.Client, req ScanRequest, baseline probeResult) []Finding {
	tampered := []struct {
		key   string
		value string
	}{
		{key: "id", value: "999999999"},
		{key: "id", value: "-1"},
		{key: "user_id", value: "1"},
		{key: "account_id", value: "1"},
		{key: "role", value: "admin"},
		{key: "permission", value: "admin"},
		{key: "isAdmin", value: "true"},
		{key: "tenant", value: "root"},
	}

	for _, mutation := range tampered {
		mutatedURL, mutatedBody, mutatedHeaders := applyMutationWithKey(req.Endpoint.URL, req.Endpoint.Body, req.Endpoint.Headers, mutation.key, mutation.value)
		probe := e.probe(ctx, client, req, mutatedURL, req.Endpoint.Method, mutatedBody, mutatedHeaders)
		if probe.err != nil {
			continue
		}

		if looksLikeTamperingSuccess(baseline, probe) {
			return []Finding{{
				Type:           VulnParameterTamper,
				Severity:       SeverityMedium,
				Title:          "Potential parameter tampering vulnerability",
				Description:    "Tampered parameters changed access behavior in a suspicious way.",
				Endpoint:       req.Endpoint.URL,
				Method:         req.Endpoint.Method,
				Payload:        mutation.key + "=" + mutation.value,
				Evidence:       buildEvidence(probe),
				Recommendation: "Enforce server-side authorization checks for IDs and role-related parameters.",
			}}
		}
	}

	return nil
}

func cloneHeaderMap(values map[string]string) map[string]string {
	cloned := make(map[string]string, len(values))
	for k, v := range values {
		cloned[k] = v
	}
	return cloned
}

func lookupHeaderIgnoreCase(headers map[string]string, key string) (string, bool) {
	for existingKey, value := range headers {
		if strings.EqualFold(strings.TrimSpace(existingKey), strings.TrimSpace(key)) {
			return value, true
		}
	}
	return "", false
}

func removeHeaderIgnoreCase(headers map[string]string, key string) {
	for existingKey := range headers {
		if strings.EqualFold(strings.TrimSpace(existingKey), strings.TrimSpace(key)) {
			delete(headers, existingKey)
		}
	}
}

func hasAuthContext(req ScanRequest) bool {
	return strings.TrimSpace(req.AuthHeader) != "" && strings.TrimSpace(req.AuthValue) != ""
}

func applyMutation(rawURL, body string, headers map[string]string, payload string) (string, string, map[string]string) {
	return applyMutationWithKey(rawURL, body, headers, "q", payload)
}

func applyMutationWithKey(rawURL, body string, headers map[string]string, key string, value string) (string, string, map[string]string) {
	mutatedHeaders := cloneHeaderMap(headers)

	parsed, err := url.Parse(rawURL)
	if err == nil {
		query := parsed.Query()
		query.Set(key, value)
		parsed.RawQuery = query.Encode()
		rawURL = parsed.String()
	}

	if strings.TrimSpace(body) == "" {
		return rawURL, body, mutatedHeaders
	}

	if looksLikeJSON(body) {
		mutatedHeaders[contentTypeHeader] = firstNonEmpty(mutatedHeaders[contentTypeHeader], "application/json")
		return rawURL, injectJSONField(body, key, value), mutatedHeaders
	}

	if strings.Contains(body, "=") {
		values, err := url.ParseQuery(body)
		if err == nil {
			values.Set(key, value)
			mutatedHeaders[contentTypeHeader] = firstNonEmpty(mutatedHeaders[contentTypeHeader], "application/x-www-form-urlencoded")
			return rawURL, values.Encode(), mutatedHeaders
		}
	}

	return rawURL, body + " " + value, mutatedHeaders
}

func looksLikeJSON(body string) bool {
	trimmed := strings.TrimSpace(body)
	return strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
}

func injectJSONField(raw string, key string, value string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "{}" {
		return fmt.Sprintf("{\"%s\":%q}", key, value)
	}
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return raw
	}
	body := strings.TrimSuffix(strings.TrimPrefix(trimmed, "{"), "}")
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Sprintf("{\"%s\":%q}", key, value)
	}
	return fmt.Sprintf("{%s,\"%s\":%q}", body, key, value)
}

func indicatorsOfInjection(body string, signatures []string) bool {
	lower := strings.ToLower(body)
	for _, signature := range signatures {
		if strings.Contains(lower, signature) {
			return true
		}
	}
	return false
}

func responseLooksEquivalent(a probeResult, b probeResult) bool {
	if a.statusCode != b.statusCode {
		return false
	}
	if a.body == b.body {
		return true
	}
	la := len(a.body)
	lb := len(b.body)
	if la == 0 && lb == 0 {
		return true
	}
	maxLen := maxInt(la, lb)
	minLen := minInt(la, lb)
	if maxLen == 0 {
		return true
	}
	return float64(minLen)/float64(maxLen) >= 0.85
}

func isServerErrorDelta(baseline probeResult, mutated probeResult) bool {
	if mutated.statusCode >= 500 {
		if baseline.statusCode == 0 {
			return true
		}
		return baseline.statusCode < 500
	}
	return false
}

func looksLikeTamperingSuccess(baseline probeResult, mutated probeResult) bool {
	if baseline.err != nil || mutated.err != nil {
		return false
	}
	if baseline.statusCode >= 400 && mutated.statusCode < 400 {
		return true
	}
	if baseline.statusCode < 400 && mutated.statusCode < 400 {
		a := len(strings.TrimSpace(baseline.body))
		b := len(strings.TrimSpace(mutated.body))
		if a == 0 || b == 0 {
			return false
		}
		ratio := float64(maxInt(a, b)) / float64(minInt(a, b))
		return ratio >= 1.7
	}
	return false
}

func buildEvidence(probe probeResult) string {
	if probe.err != nil {
		return probe.err.Error()
	}
	compact := strings.TrimSpace(probe.body)
	compact = strings.ReplaceAll(compact, "\n", " ")
	compact = strings.ReplaceAll(compact, "\r", " ")
	compact = strings.TrimSpace(compact)
	if len(compact) > 96 {
		compact = compact[:96] + "..."
	}
	return fmt.Sprintf("status=%d body=%q", probe.statusCode, compact)
}

func buildSummary(findings []Finding) Summary {
	summary := Summary{Total: len(findings)}
	for _, finding := range findings {
		switch strings.ToUpper(strings.TrimSpace(finding.Severity)) {
		case SeverityCritical:
			summary.Critical++
		case SeverityHigh:
			summary.High++
		case SeverityMedium:
			summary.Medium++
		case SeverityLow:
			summary.Low++
		default:
			summary.Info++
		}
	}
	return summary
}

func exponentialBackoff(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		base = time.Second
	}
	if attempt < 1 {
		attempt = 1
	}

	factor := math.Pow(2, float64(attempt-1))
	delay := float64(base) * factor
	if delay > float64(math.MaxInt64) {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(delay)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
