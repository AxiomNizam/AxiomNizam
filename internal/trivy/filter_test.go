package trivy

import "testing"

func TestParseSeverityCSV(t *testing.T) {
	values := ParseSeverityCSV("critical, high, bogus")
	if len(values) != 2 {
		t.Fatalf("expected 2 severities, got %d", len(values))
	}
	if values[0] != SeverityCritical || values[1] != SeverityHigh {
		t.Fatalf("unexpected severity parse result: %v", values)
	}
}

func TestFilterFindingsIgnoreUnfixed(t *testing.T) {
	findings := []Finding{
		{Category: "vulnerability", Severity: SeverityCritical, ID: "CVE-1", Unfixed: true},
		{Category: "vulnerability", Severity: SeverityCritical, ID: "CVE-2", Unfixed: false},
		{Category: "misconfiguration", Severity: SeverityCritical, ID: "MIS-1", Unfixed: false},
	}

	filtered := FilterFindings(findings, BuildFilterOptions([]string{SeverityCritical}, true, nil))
	if len(filtered) != 2 {
		t.Fatalf("expected 2 findings after filter, got %d", len(filtered))
	}
	if filtered[0].ID == "CVE-1" || filtered[1].ID == "CVE-1" {
		t.Fatal("unfixed vulnerability should have been removed")
	}
}
