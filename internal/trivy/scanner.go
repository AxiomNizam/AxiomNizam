package trivy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type Engine struct {
	wrapper *ExternalWrapper
}

func NewEngine(binaryPath string) *Engine {
	return &Engine{wrapper: NewExternalWrapper(binaryPath)}
}

func (e *Engine) Scan(ctx context.Context, req ScanRequest) (ScanResult, error) {
	if e == nil {
		return ScanResult{}, fmt.Errorf("trivy engine is not configured")
	}

	normalizedReq, err := normalizeRequest(req)
	if err != nil {
		return ScanResult{}, err
	}

	if !normalizedReq.UseExternal {
		return ScanResult{}, fmt.Errorf("embedded trivy scanner is disabled; use --external")
	}

	args, err := e.buildArgs(normalizedReq)
	if err != nil {
		return ScanResult{}, err
	}

	raw, err := e.runWithRetry(ctx, args, normalizedReq.RetryCount, normalizedReq.RetryBackoff)
	if err != nil {
		return ScanResult{}, err
	}

	parsed, err := parseTrivyJSON(raw)
	if err != nil {
		return ScanResult{}, err
	}

	findings := normalizeFindings(parsed)
	filters := BuildFilterOptions(normalizedReq.Severity, normalizedReq.IgnoreUnfixed, normalizedReq.PolicyHooks)
	findings = FilterFindings(findings, filters)

	result := ScanResult{
		Scanner:      "trivy",
		TargetKind:   string(normalizedReq.Kind),
		Target:       normalizedReq.Target,
		ArtifactName: parsed.ArtifactName,
		ArtifactType: parsed.ArtifactType,
		ScannedAt:    time.Now().UTC(),
		Findings:     findings,
		Summary:      buildSummary(findings),
		Metadata: map[string]string{
			"source": "external-binary",
		},
	}

	return result, nil
}

func normalizeRequest(req ScanRequest) (ScanRequest, error) {
	clone := req
	clone.Target = strings.TrimSpace(clone.Target)
	if clone.Target == "" {
		return ScanRequest{}, fmt.Errorf("scan target is required")
	}

	if clone.Kind != TargetImage && clone.Kind != TargetFS && clone.Kind != TargetK8s && clone.Kind != TargetRepo {
		return ScanRequest{}, fmt.Errorf("unsupported scan target kind %q", clone.Kind)
	}

	if clone.Timeout <= 0 {
		clone.Timeout = 5 * time.Minute
	}
	if clone.RetryCount < 0 {
		clone.RetryCount = 0
	}
	if clone.RetryBackoff <= 0 {
		clone.RetryBackoff = time.Second
	}
	if len(clone.Severity) == 0 {
		clone.Severity = []string{SeverityHigh, SeverityCritical}
	}

	if clone.Format == "" {
		clone.Format = FormatTable
	}

	if !clone.UseExternal {
		clone.UseExternal = false
	}

	return clone, nil
}

func (e *Engine) buildArgs(req ScanRequest) ([]string, error) {
	baseCmd, err := trivySubcommand(req.Kind)
	if err != nil {
		return nil, err
	}

	args := []string{baseCmd}
	args = append(args, "--quiet", "--format", "json", "--timeout", req.Timeout.String())

	if len(req.Severity) > 0 {
		args = append(args, "--severity", strings.Join(req.Severity, ","))
	}
	if req.IgnoreUnfixed {
		args = append(args, "--ignore-unfixed")
	}

	if req.Kind == TargetK8s {
		args = append(args, "--scanners", "misconfig,secret")
	}

	args = append(args, req.Target)
	return args, nil
}

func trivySubcommand(kind TargetKind) (string, error) {
	switch kind {
	case TargetImage:
		return "image", nil
	case TargetFS:
		return "fs", nil
	case TargetK8s:
		return "config", nil
	case TargetRepo:
		return "repo", nil
	default:
		return "", fmt.Errorf("unsupported scan target kind %q", kind)
	}
}

func (e *Engine) runWithRetry(ctx context.Context, args []string, retries int, baseBackoff time.Duration) ([]byte, error) {
	attempts := retries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		output, err := e.wrapper.RunJSON(ctx, args)
		if err == nil {
			return output, nil
		}
		lastErr = err

		if attempt == attempts {
			break
		}

		delay := exponentialBackoff(baseBackoff, attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return nil, fmt.Errorf("trivy scan failed after %d attempt(s): %w", attempts, lastErr)
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

type trivyReport struct {
	ArtifactName string        `json:"ArtifactName"`
	ArtifactType string        `json:"ArtifactType"`
	Results      []trivyResult `json:"Results"`
}

type trivyResult struct {
	Target            string                  `json:"Target"`
	Class             string                  `json:"Class"`
	Type              string                  `json:"Type"`
	Vulnerabilities   []trivyVulnerability    `json:"Vulnerabilities"`
	Misconfigurations []trivyMisconfiguration `json:"Misconfigurations"`
	Secrets           []trivySecret           `json:"Secrets"`
	Licenses          []trivyLicense          `json:"Licenses"`
}

type trivyVulnerability struct {
	ID               string `json:"VulnerabilityID"`
	PkgName          string `json:"PkgName"`
	InstalledVersion string `json:"InstalledVersion"`
	FixedVersion     string `json:"FixedVersion"`
	Severity         string `json:"Severity"`
	Title            string `json:"Title"`
	Description      string `json:"Description"`
	PrimaryURL       string `json:"PrimaryURL"`
	Status           string `json:"Status"`
}

type trivyMisconfiguration struct {
	ID          string `json:"ID"`
	AVDID       string `json:"AVDID"`
	Type        string `json:"Type"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Message     string `json:"Message"`
	Severity    string `json:"Severity"`
	Status      string `json:"Status"`
	Resolution  string `json:"Resolution"`
	PrimaryURL  string `json:"PrimaryURL"`
}

type trivySecret struct {
	RuleID     string `json:"RuleID"`
	Category   string `json:"Category"`
	Severity   string `json:"Severity"`
	Title      string `json:"Title"`
	StartLine  int    `json:"StartLine"`
	EndLine    int    `json:"EndLine"`
	Match      string `json:"Match"`
	PrimaryURL string `json:"PrimaryURL"`
}

type trivyLicense struct {
	Name     string `json:"Name"`
	Severity string `json:"Severity"`
	Category string `json:"Category"`
}

func parseTrivyJSON(raw []byte) (trivyReport, error) {
	if len(raw) == 0 {
		return trivyReport{}, errors.New("trivy output is empty")
	}

	var report trivyReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return trivyReport{}, fmt.Errorf("failed to parse trivy JSON: %w", err)
	}

	return report, nil
}

func normalizeFindings(report trivyReport) []Finding {
	findings := make([]Finding, 0)

	for _, result := range report.Results {
		for _, vuln := range result.Vulnerabilities {
			findings = append(findings, Finding{
				Category:         "vulnerability",
				ID:               strings.TrimSpace(vuln.ID),
				Severity:         normalizeSeverityFallback(vuln.Severity),
				Title:            strings.TrimSpace(vuln.Title),
				Description:      strings.TrimSpace(vuln.Description),
				Target:           strings.TrimSpace(result.Target),
				Resource:         strings.TrimSpace(result.Target),
				PackageName:      strings.TrimSpace(vuln.PkgName),
				InstalledVersion: strings.TrimSpace(vuln.InstalledVersion),
				FixedVersion:     strings.TrimSpace(vuln.FixedVersion),
				Reference:        strings.TrimSpace(vuln.PrimaryURL),
				Status:           strings.TrimSpace(vuln.Status),
				Unfixed:          strings.TrimSpace(vuln.FixedVersion) == "",
			})
		}

		for _, misconf := range result.Misconfigurations {
			id := firstDefined(misconf.ID, misconf.AVDID)
			title := firstDefined(misconf.Title, misconf.Message)
			description := firstDefined(misconf.Description, misconf.Resolution)
			findings = append(findings, Finding{
				Category:    "misconfiguration",
				ID:          strings.TrimSpace(id),
				Severity:    normalizeSeverityFallback(misconf.Severity),
				Title:       strings.TrimSpace(title),
				Description: strings.TrimSpace(description),
				Target:      strings.TrimSpace(result.Target),
				Resource:    strings.TrimSpace(result.Target),
				Reference:   strings.TrimSpace(misconf.PrimaryURL),
				Status:      strings.TrimSpace(misconf.Status),
			})
		}

		for _, secret := range result.Secrets {
			title := strings.TrimSpace(secret.Title)
			if title == "" {
				title = strings.TrimSpace(secret.Category)
			}
			findings = append(findings, Finding{
				Category:    "secret",
				ID:          strings.TrimSpace(secret.RuleID),
				Severity:    normalizeSeverityFallback(secret.Severity),
				Title:       title,
				Description: strings.TrimSpace(secret.Match),
				Target:      strings.TrimSpace(result.Target),
				Resource:    strings.TrimSpace(result.Target),
				Reference:   strings.TrimSpace(secret.PrimaryURL),
			})
		}

		for _, license := range result.Licenses {
			findings = append(findings, Finding{
				Category: "license",
				ID:       strings.TrimSpace(license.Name),
				Severity: normalizeSeverityFallback(license.Severity),
				Title:    strings.TrimSpace(license.Category),
				Target:   strings.TrimSpace(result.Target),
				Resource: strings.TrimSpace(result.Target),
			})
		}
	}

	return findings
}

func normalizeSeverityFallback(raw string) string {
	normalized := normalizeSeverity(raw)
	if normalized == "" {
		return SeverityUnknown
	}
	return normalized
}

func buildSummary(findings []Finding) Summary {
	summary := Summary{Total: len(findings)}

	for _, finding := range findings {
		switch normalizeSeverityFallback(finding.Severity) {
		case SeverityCritical:
			summary.Critical++
		case SeverityHigh:
			summary.High++
		case SeverityMedium:
			summary.Medium++
		case SeverityLow:
			summary.Low++
		default:
			summary.Unknown++
		}
	}

	return summary
}

func firstDefined(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
