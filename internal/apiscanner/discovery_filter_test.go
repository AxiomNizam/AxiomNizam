package apiscanner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testWellKnownHealthID = "well-known.health"
	testHintOpenAPIID     = "hint.openapi"
)

func TestNormalizeDiscoveryIDs(t *testing.T) {
	ids := normalizeDiscoveryIDs([]string{" well-known.health , well-known.healthz", discoveryCheckFingerprintHeaders, ""})
	if len(ids) != 3 {
		t.Fatalf("expected 3 normalized IDs, got %d", len(ids))
	}
	if ids[0] != discoveryCheckFingerprintHeaders {
		t.Fatalf("expected sorted normalized IDs, got %v", ids)
	}
}

func TestValidateDiscoveryIDSelectionRejectsUnknown(t *testing.T) {
	err := validateDiscoveryIDSelection([]string{"unknown.id"}, nil, supportedAPIDiscoveryCheckIDs())
	if err == nil {
		t.Fatal("expected validation error for unknown include ID")
	}
}

func TestGetSupportedDiscoveryScanIDs(t *testing.T) {
	supported := GetSupportedDiscoveryScanIDs()
	if len(supported.API) == 0 {
		t.Fatal("expected API discovery scan IDs")
	}
	if len(supported.Domain) == 0 {
		t.Fatal("expected domain discovery scan IDs")
	}

	if !containsString(supported.API, testWellKnownHealthID) {
		t.Fatalf("expected API scan IDs to include %s, got %v", testWellKnownHealthID, supported.API)
	}
	if !containsString(supported.API, discoveryCheckFingerprintHeaders) {
		t.Fatalf("expected API scan IDs to include %s", discoveryCheckFingerprintHeaders)
	}
	if !containsString(supported.Domain, testHintOpenAPIID) {
		t.Fatalf("expected domain scan IDs to include %s, got %v", testHintOpenAPIID, supported.Domain)
	}
	if !containsString(supported.Domain, discoveryCheckFingerprintHeaders) {
		t.Fatalf("expected domain scan IDs to include %s", discoveryCheckFingerprintHeaders)
	}
}

func TestDiscoverAPIIncludeScanID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	result, err := DiscoverAPI(context.Background(), DiscoverRequest{
		Target:     server.URL,
		Timeout:    3 * time.Second,
		MaxPaths:   64,
		IncludeIDs: []string{testWellKnownHealthID},
	})
	if err != nil {
		t.Fatalf("DiscoverAPI failed: %v", err)
	}

	if len(result.Discovered) != 1 {
		t.Fatalf("expected 1 discovered path, got %d", len(result.Discovered))
	}
	if result.Discovered[0].CheckID != testWellKnownHealthID {
		t.Fatalf("expected check ID %s, got %s", testWellKnownHealthID, result.Discovered[0].CheckID)
	}
	if len(result.Fingerprints) != 0 {
		t.Fatalf("expected no fingerprints when include-scan-id omits fingerprint check, got %d", len(result.Fingerprints))
	}
}

func TestDiscoverAPIExcludeFingerprint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Powered-By", "test-suite")
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	result, err := DiscoverAPI(context.Background(), DiscoverRequest{
		Target:     server.URL,
		Timeout:    3 * time.Second,
		MaxPaths:   64,
		ExcludeIDs: []string{discoveryCheckFingerprintHeaders},
	})
	if err != nil {
		t.Fatalf("DiscoverAPI failed: %v", err)
	}

	if len(result.Discovered) == 0 {
		t.Fatal("expected discovered paths to still run when excluding fingerprint check")
	}
	if len(result.Fingerprints) != 0 {
		t.Fatalf("expected 0 fingerprints when fingerprint check is excluded, got %d", len(result.Fingerprints))
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
