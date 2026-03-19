package apiscanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

func RenderOutput(result ScanResult, format OutputFormat) (string, error) {
	switch format {
	case FormatJSON:
		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal scan result: %w", err)
		}
		return string(encoded), nil
	case FormatTable:
		return formatTable(result), nil
	default:
		return "", fmt.Errorf("unsupported output format %q", format)
	}
}

func formatTable(result ScanResult) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	fmt.Fprintf(writer, "SCANNER:\t%s\n", result.Scanner)
	fmt.Fprintf(writer, "TARGET:\t%s\n", result.Target)
	fmt.Fprintf(writer, "METHOD:\t%s\n", result.Method)
	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "SEVERITY\tTYPE\tTITLE\tMETHOD\tEVIDENCE")
	fmt.Fprintln(writer, "--------\t----\t-----\t------\t--------")

	findings := append([]Finding(nil), result.Findings...)
	sort.SliceStable(findings, func(i, j int) bool {
		return severityWeight(findings[i].Severity) > severityWeight(findings[j].Severity)
	})

	for _, finding := range findings {
		fmt.Fprintf(writer,
			"%s\t%s\t%s\t%s\t%s\n",
			normalizeSeverity(finding.Severity),
			string(finding.Type),
			trimText(finding.Title, 56),
			finding.Method,
			trimText(firstNonEmpty(finding.Evidence, finding.Payload), 72),
		)
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "SUMMARY")
	fmt.Fprintf(writer, "TOTAL\t%d\n", result.Summary.Total)
	fmt.Fprintf(writer, "CRITICAL\t%d\n", result.Summary.Critical)
	fmt.Fprintf(writer, "HIGH\t%d\n", result.Summary.High)
	fmt.Fprintf(writer, "MEDIUM\t%d\n", result.Summary.Medium)
	fmt.Fprintf(writer, "LOW\t%d\n", result.Summary.Low)
	fmt.Fprintf(writer, "INFO\t%d\n", result.Summary.Info)

	_ = writer.Flush()
	return strings.TrimRight(buffer.String(), "\n")
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

func normalizeSeverity(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return SeverityInfo
	}
	return normalized
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
