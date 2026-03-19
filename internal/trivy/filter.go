package trivy

import "strings"

type FilterOptions struct {
	AllowedSeverities map[string]struct{}
	IgnoreUnfixed     bool
	PolicyHooks       []PolicyHook
}

func BuildFilterOptions(severity []string, ignoreUnfixed bool, hooks []PolicyHook) FilterOptions {
	allowed := make(map[string]struct{}, len(severity))
	for _, item := range severity {
		normalized := normalizeSeverity(item)
		if normalized == "" {
			continue
		}
		allowed[normalized] = struct{}{}
	}

	if len(allowed) == 0 {
		allowed[SeverityHigh] = struct{}{}
		allowed[SeverityCritical] = struct{}{}
	}

	return FilterOptions{
		AllowedSeverities: allowed,
		IgnoreUnfixed:     ignoreUnfixed,
		PolicyHooks:       hooks,
	}
}

func FilterFindings(findings []Finding, options FilterOptions) []Finding {
	filtered := make([]Finding, 0, len(findings))

	for _, finding := range findings {
		if !passesSeverityFilter(finding, options.AllowedSeverities) {
			continue
		}
		if options.IgnoreUnfixed && finding.Category == "vulnerability" && finding.Unfixed {
			continue
		}
		if !passesPolicyHooks(finding, options.PolicyHooks) {
			continue
		}
		filtered = append(filtered, finding)
	}

	return filtered
}

func passesSeverityFilter(finding Finding, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return true
	}
	severity := normalizeSeverity(finding.Severity)
	_, ok := allowed[severity]
	return ok
}

func passesPolicyHooks(finding Finding, hooks []PolicyHook) bool {
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if !hook(finding) {
			return false
		}
	}
	return true
}

func ParseSeverityCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{SeverityHigh, SeverityCritical}
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := normalizeSeverity(part)
		if normalized == "" {
			continue
		}
		result = append(result, normalized)
	}

	if len(result) == 0 {
		return []string{SeverityHigh, SeverityCritical}
	}

	return result
}

func normalizeSeverity(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "CRITICAL":
		return SeverityCritical
	case "HIGH":
		return SeverityHigh
	case "MEDIUM":
		return SeverityMedium
	case "LOW":
		return SeverityLow
	case "UNKNOWN":
		return SeverityUnknown
	default:
		return ""
	}
}
