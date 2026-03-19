package apiscanner

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEngineDetectsSecurityHeadersAndMethodIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodTrace {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("trace enabled"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	engine := NewEngine()
	result, err := engine.Scan(context.Background(), ScanRequest{
		Endpoint: Endpoint{URL: server.URL, Method: http.MethodGet},
		Timeout:  3 * time.Second,
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !hasFinding(result.Findings, VulnSecurityHeaders) {
		t.Fatalf("expected security header finding, got %+v", result.Findings)
	}
	if !hasFinding(result.Findings, VulnHTTPMethod) {
		t.Fatalf("expected HTTP method finding, got %+v", result.Findings)
	}
}

func TestEngineDetectsSQLiAndXSS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		switch {
		case strings.Contains(q, "' OR '1'='1"):
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("You have an error in your SQL syntax"))
		case strings.Contains(q, "<script>alert('XSS')</script>"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(q))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("safe"))
		}
	}))
	defer server.Close()

	engine := NewEngine()
	result, err := engine.Scan(context.Background(), ScanRequest{
		Endpoint: Endpoint{URL: server.URL, Method: http.MethodGet},
		Timeout:  3 * time.Second,
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !hasFinding(result.Findings, VulnSQLInjection) {
		t.Fatalf("expected SQL injection finding, got %+v", result.Findings)
	}
	if !hasFinding(result.Findings, VulnXSS) {
		t.Fatalf("expected XSS finding, got %+v", result.Findings)
	}
}

func TestEngineDetectsNoSQLInjection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if strings.Contains(filter, "{$ne:null}") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("MongoError: cannot deserialize expression"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	engine := NewEngine()
	result, err := engine.Scan(context.Background(), ScanRequest{
		Endpoint: Endpoint{URL: server.URL, Method: http.MethodGet},
		Timeout:  3 * time.Second,
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !hasFinding(result.Findings, VulnNoSQLInjection) {
		t.Fatalf("expected NoSQL injection finding, got %+v", result.Findings)
	}
}

func TestEngineDetectsAuthBypass(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		hasBypassHeader := r.Header.Get("X-Original-URL") != ""

		if auth == "Bearer good-token" || hasBypassHeader {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("sensitive data"))
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	engine := NewEngine()
	result, err := engine.Scan(context.Background(), ScanRequest{
		Endpoint:   Endpoint{URL: server.URL, Method: http.MethodGet},
		AuthHeader: "Authorization",
		AuthValue:  "Bearer good-token",
		Timeout:    3 * time.Second,
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !hasFinding(result.Findings, VulnAuthBypass) {
		t.Fatalf("expected auth bypass finding, got %+v", result.Findings)
	}
}

func TestEngineDetectsParameterTampering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") == "999999999" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(strings.Repeat("x", 240)))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	engine := NewEngine()
	result, err := engine.Scan(context.Background(), ScanRequest{
		Endpoint: Endpoint{URL: server.URL, Method: http.MethodGet},
		Timeout:  3 * time.Second,
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if !hasFinding(result.Findings, VulnParameterTamper) {
		t.Fatalf("expected parameter tampering finding, got %+v", result.Findings)
	}
}

func TestRenderOutputJSON(t *testing.T) {
	result := ScanResult{
		Scanner: "api-scanner",
		Target:  "https://api.example.com/users",
		Method:  http.MethodGet,
		Findings: []Finding{
			{Type: VulnXSS, Severity: SeverityHigh, Title: "Potential reflected XSS", Endpoint: "https://api.example.com/users", Method: http.MethodGet},
		},
		Summary: Summary{Total: 1, High: 1},
	}

	output, err := RenderOutput(result, FormatJSON)
	if err != nil {
		t.Fatalf("render output failed: %v", err)
	}

	if !strings.Contains(output, fmt.Sprintf("\"scanner\": %q", result.Scanner)) {
		t.Fatalf("expected JSON output with scanner field, got: %s", output)
	}
}

func hasFinding(findings []Finding, vulnType VulnerabilityType) bool {
	for _, finding := range findings {
		if finding.Type == vulnType {
			return true
		}
	}
	return false
}
