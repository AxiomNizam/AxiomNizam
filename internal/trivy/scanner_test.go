package trivy

import (
	"testing"
	"time"
)

func TestBuildSummary(t *testing.T) {
	findings := []Finding{
		{Severity: SeverityCritical},
		{Severity: SeverityHigh},
		{Severity: SeverityHigh},
		{Severity: SeverityMedium},
		{Severity: SeverityLow},
		{Severity: "unexpected"},
	}

	summary := buildSummary(findings)
	if summary.Total != 6 {
		t.Fatalf("expected total 6, got %d", summary.Total)
	}
	if summary.Critical != 1 || summary.High != 2 || summary.Medium != 1 || summary.Low != 1 || summary.Unknown != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestExponentialBackoff(t *testing.T) {
	base := 200 * time.Millisecond
	d1 := exponentialBackoff(base, 1)
	d2 := exponentialBackoff(base, 2)
	d3 := exponentialBackoff(base, 3)

	if d1 != 200*time.Millisecond || d2 != 400*time.Millisecond || d3 != 800*time.Millisecond {
		t.Fatalf("unexpected backoff durations: %s, %s, %s", d1, d2, d3)
	}
}
