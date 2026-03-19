package apiscanner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testDomain = "example.com"

func TestNormalizeDomainTargetAcceptsURL(t *testing.T) {
	value, err := normalizeDomainTarget("https://api.example.com/v1/users")
	if err != nil {
		t.Fatalf("normalizeDomainTarget returned error: %v", err)
	}
	if value != "api.example.com" {
		t.Fatalf("unexpected normalized domain: %s", value)
	}
}

func TestParseDiscoverySchemesRejectsUnknown(t *testing.T) {
	_, err := parseDiscoverySchemes([]string{"https", "ftp"})
	if err == nil {
		t.Fatal("expected parseDiscoverySchemes to reject unsupported scheme")
	}
}

func TestBuildSubdomainCandidatesRespectsLimit(t *testing.T) {
	candidates := buildSubdomainCandidates(testDomain, 3)
	if len(candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(candidates))
	}
	if candidates[0] != testDomain {
		t.Fatalf("expected first candidate to be root domain, got %s", candidates[0])
	}
}

func TestProbeDomainAPIHints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Powered-By", "test-suite")
		switch r.URL.Path {
		case "/openapi.json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"openapi":"3.0.0"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	hints, fingerprints := probeDomainAPIHints(context.Background(), DiscoverDomainRequest{
		Timeout:  3 * time.Second,
		MaxHints: 10,
		Schemes:  []string{"http"},
	}, []ResolvedSubdomain{{Host: host}})

	if len(hints) == 0 {
		t.Fatalf("expected at least one API hint, got 0")
	}
	if hints[0].Category != discoveryCategoryAPIHint {
		t.Fatalf("unexpected hint category: %s", hints[0].Category)
	}
	if hints[0].CheckID == "" {
		t.Fatal("expected hint check ID to be populated")
	}
	if len(fingerprints) == 0 {
		t.Fatalf("expected at least one fingerprint, got 0")
	}
}

func TestNormalizeDiscoverDomainRequestRejectsUnknownScanID(t *testing.T) {
	_, err := normalizeDiscoverDomainRequest(DiscoverDomainRequest{
		Target:     testDomain,
		IncludeIDs: []string{"hint.unknown"},
	})
	if err == nil {
		t.Fatal("expected unknown scan ID validation error")
	}
}

func TestProbeHintsForHostSchemeIncludeScanID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Powered-By", "test-suite")
		switch r.URL.Path {
		case "/openapi.json", "/graphql":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	filter := newDiscoveryCheckFilter([]string{"hint.openapi"}, nil)
	hints, fingerprints := probeHintsForHostScheme(
		context.Background(),
		newDiscoveryClient(3*time.Second, false),
		DiscoverDomainRequest{Headers: map[string]string{}},
		filter,
		host,
		"http",
		10,
	)

	if len(hints) != 1 {
		t.Fatalf("expected exactly one hint, got %d", len(hints))
	}
	if hints[0].CheckID != "hint.openapi" {
		t.Fatalf("expected hint.openapi check ID, got %s", hints[0].CheckID)
	}
	if len(fingerprints) != 0 {
		t.Fatalf("expected no fingerprints when include IDs omit fingerprint check, got %d", len(fingerprints))
	}
}

func TestRenderDomainDiscoveryOutputJSON(t *testing.T) {
	result := DomainDiscoveryResult{
		Scanner:    "domain-discovery",
		Target:     testDomain,
		Candidates: []string{testDomain, "api.example.com"},
		Summary: DomainDiscoveryStats{
			Candidates: 2,
			Resolved:   1,
			APIHints:   1,
		},
	}

	output, err := RenderDomainDiscoveryOutput(result, FormatJSON)
	if err != nil {
		t.Fatalf("RenderDomainDiscoveryOutput returned error: %v", err)
	}
	if !strings.Contains(output, `"scanner": "domain-discovery"`) {
		t.Fatalf("expected JSON scanner field in output, got: %s", output)
	}
}
