package trivy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

const (
	colorRed   = "\u001b[31m"
	colorReset = "\u001b[0m"
)

func RenderOutput(result ScanResult, format OutputFormat) (string, error) {
	switch format {
	case FormatJSON:
		return formatJSON(result)
	case FormatTable:
		return formatTable(result), nil
	default:
		return "", fmt.Errorf("unsupported output format %q", format)
	}
}

func formatJSON(result ScanResult) (string, error) {
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal scan result: %w", err)
	}
	return string(encoded), nil
}

func formatTable(result ScanResult) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	fmt.Fprintf(writer, "SCANNER:\t%s\n", result.Scanner)
	fmt.Fprintf(writer, "TARGET KIND:\t%s\n", result.TargetKind)
	fmt.Fprintf(writer, "TARGET:\t%s\n", result.Target)
	if result.ArtifactName != "" {
		fmt.Fprintf(writer, "ARTIFACT:\t%s\n", result.ArtifactName)
	}
	if result.ArtifactType != "" {
		fmt.Fprintf(writer, "ARTIFACT TYPE:\t%s\n", result.ArtifactType)
	}
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "SEVERITY\tTYPE\tID\tTARGET\tPACKAGE\tINSTALLED\tFIXED\tTITLE")
	fmt.Fprintln(writer, "--------\t----\t--\t------\t-------\t---------\t-----\t-----")

	findings := append([]Finding(nil), result.Findings...)
	sort.SliceStable(findings, func(i, j int) bool {
		return severityWeight(findings[i].Severity) > severityWeight(findings[j].Severity)
	})

	for _, finding := range findings {
		severity := highlightSeverity(finding.Severity)
		title := trimText(firstNonEmpty(finding.Title, finding.Description), 72)
		fmt.Fprintf(writer,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			severity,
			trimText(finding.Category, 16),
			trimText(finding.ID, 24),
			trimText(firstNonEmpty(finding.Resource, finding.Target), 28),
			trimText(finding.PackageName, 18),
			trimText(finding.InstalledVersion, 12),
			trimText(finding.FixedVersion, 12),
			title,
		)
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "SUMMARY")
	fmt.Fprintf(writer, "TOTAL\t%d\n", result.Summary.Total)
	fmt.Fprintf(writer, "CRITICAL\t%d\n", result.Summary.Critical)
	fmt.Fprintf(writer, "HIGH\t%d\n", result.Summary.High)
	fmt.Fprintf(writer, "MEDIUM\t%d\n", result.Summary.Medium)
	fmt.Fprintf(writer, "LOW\t%d\n", result.Summary.Low)
	fmt.Fprintf(writer, "UNKNOWN\t%d\n", result.Summary.Unknown)

	_ = writer.Flush()
	return strings.TrimRight(buffer.String(), "\n")
}

func highlightSeverity(severity string) string {
	normalized := normalizeSeverity(severity)
	if normalized == SeverityCritical {
		return colorRed + "CRITICAL" + colorReset
	}
	if normalized == "" {
		return SeverityUnknown
	}
	return normalized
}

func severityWeight(severity string) int {
	switch normalizeSeverity(severity) {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	default:
		return 1
	}
}

func trimText(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	if maxLen < 4 {
		return value[:maxLen]
	}
	return value[:maxLen-3] + "..."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
