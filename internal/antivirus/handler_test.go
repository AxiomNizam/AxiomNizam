package antivirus

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestEngine creates a minimal engine for handler tests.
func newTestEngine() *Engine {
	cfg := &Config{
		Enabled:          true,
		Workers:          2,
		QueueSize:        100,
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		CacheSize:        1000,
		CacheTTL:         time.Hour,
		HashDBEnabled:    true,
		PatternEnabled:   true,
		HeuristicEnabled: true,
		EntropyEnabled:   true,
		YARAEnabled:      true,
	}
	return NewEngine(cfg)
}

func setupRouter(engine *Engine) *gin.Engine {
	router := gin.New()
	handler := NewAPIHandler(engine)
	handler.RegisterRoutes(router.Group("/api"))
	return router
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/status
// ─────────────────────────────────────────────────────────────────────────────

func TestAPIHandler_Status_Disabled(t *testing.T) {
	cfg := &Config{Enabled: false}
	engine := NewEngine(cfg)
	router := setupRouter(engine)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/antivirus/status", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "disabled" {
		t.Errorf("expected 'disabled', got %v", resp["status"])
	}
}

func TestAPIHandler_Status_Running(t *testing.T) {
	engine := newTestEngine()
	engine.Start()
	defer engine.Shutdown(context.Background())

	router := setupRouter(engine)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/antivirus/status", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "running" {
		t.Errorf("expected 'running', got %v", resp["status"])
	}
	if resp["engineVersion"] != EngineVersion {
		t.Errorf("expected engine version %s, got %v", EngineVersion, resp["engineVersion"])
	}

	// Check nested features object.
	features, ok := resp["features"].(map[string]interface{})
	if !ok {
		t.Fatal("expected features object")
	}
	if features["hashDB"] != true {
		t.Error("hashDB should be enabled")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/stats
// ─────────────────────────────────────────────────────────────────────────────

func TestAPIHandler_Stats(t *testing.T) {
	engine := newTestEngine()
	engine.Start()
	defer engine.Shutdown(context.Background())

	router := setupRouter(engine)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/antivirus/stats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Should have zero values initially.
	if resp["totalScanned"] != float64(0) {
		t.Errorf("expected 0 totalScanned, got %v", resp["totalScanned"])
	}
	if resp["threatsFound"] != float64(0) {
		t.Errorf("expected 0 threatsFound, got %v", resp["threatsFound"])
	}

	// Cache should be present.
	cache, ok := resp["cache"].(map[string]interface{})
	if !ok {
		t.Fatal("expected cache object")
	}
	if cache["hits"] != float64(0) {
		t.Errorf("expected 0 cache hits, got %v", cache["hits"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /antivirus/scan
// ─────────────────────────────────────────────────────────────────────────────

func TestAPIHandler_ManualScan_Clean(t *testing.T) {
	engine := newTestEngine()
	engine.Start()
	defer engine.Shutdown(context.Background())

	router := setupRouter(engine)

	// Create multipart upload with clean content.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "clean.txt")
	part.Write([]byte("This is a perfectly clean file with no malicious content."))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/antivirus/scan", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result ScanResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result.Verdict != VerdictClean {
		t.Errorf("expected clean verdict, got %s", result.Verdict)
	}
	if result.Filename != "clean.txt" {
		t.Errorf("expected filename 'clean.txt', got %q", result.Filename)
	}
	if result.SHA256 == "" {
		t.Error("expected SHA256 to be set")
	}
}

func TestAPIHandler_ManualScan_NoFile(t *testing.T) {
	engine := newTestEngine()
	engine.Start()
	defer engine.Shutdown(context.Background())

	router := setupRouter(engine)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/antivirus/scan", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPIHandler_ManualScan_EngineNotRunning(t *testing.T) {
	engine := newTestEngine()
	// Not started — engine is not running.

	router := setupRouter(engine)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/antivirus/scan", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/threats
// ─────────────────────────────────────────────────────────────────────────────

func TestAPIHandler_Threats_Empty(t *testing.T) {
	engine := newTestEngine()
	engine.Start()
	defer engine.Shutdown(context.Background())

	router := setupRouter(engine)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/antivirus/threats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["count"] != float64(0) {
		t.Errorf("expected 0 threats, got %v", resp["count"])
	}
}

func TestAPIHandler_Threats_WithDetection(t *testing.T) {
	engine := newTestEngine()
	engine.Start()
	defer engine.Shutdown(context.Background())

	// Manually record a threat.
	engine.recordThreat(ScanResult{
		Verdict:  VerdictMalware,
		Filename: "evil.exe",
		SHA256:   "abc123",
		Threats: []ThreatInfo{
			{Name: "Trojan.Test", Severity: SeverityCritical},
		},
		ScannedAt: time.Now(),
	})

	router := setupRouter(engine)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/antivirus/threats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["count"] != float64(1) {
		t.Errorf("expected 1 threat, got %v", resp["count"])
	}

	threats := resp["threats"].([]interface{})
	first := threats[0].(map[string]interface{})
	if first["filename"] != "evil.exe" {
		t.Errorf("expected 'evil.exe', got %v", first["filename"])
	}
	if first["severity"] != "critical" {
		t.Errorf("expected 'critical', got %v", first["severity"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/config
// ─────────────────────────────────────────────────────────────────────────────

func TestAPIHandler_Config(t *testing.T) {
	engine := newTestEngine()
	router := setupRouter(engine)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/antivirus/config", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["enabled"] != true {
		t.Error("expected enabled=true")
	}
	if resp["workers"] != float64(2) {
		t.Errorf("expected 2 workers, got %v", resp["workers"])
	}

	layers := resp["layers"].(map[string]interface{})
	if layers["hashDB"] != true {
		t.Error("hashDB should be enabled")
	}
	if layers["yara"] != true {
		t.Error("yara should be enabled")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// redactURL
// ─────────────────────────────────────────────────────────────────────────────

func TestRedactURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "(not configured)"},
		{"https://updates.example.com/v1/sigs", "https://***" + "/v1/sigs"},
		{"http://localhost:8080/update", "http://***" + "/update"},
		{"ftp://host", "ftp://host"}, // no path segment
	}

	for _, tt := range tests {
		got := redactURL(tt.input)
		if got != tt.expected {
			t.Errorf("redactURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Nil handler
// ─────────────────────────────────────────────────────────────────────────────

func TestAPIHandler_Nil(t *testing.T) {
	// Should not panic.
	var h *APIHandler
	h.RegisterRoutes(nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// Engine threat log
// ─────────────────────────────────────────────────────────────────────────────

func TestEngine_ThreatLog_Cap(t *testing.T) {
	engine := newTestEngine()

	// Record more than maxThreatLogSize.
	for i := 0; i < maxThreatLogSize+50; i++ {
		engine.recordThreat(ScanResult{
			Verdict:  VerdictMalware,
			Filename: "test.exe",
			SHA256:   "hash",
		})
	}

	threats := engine.RecentThreats()
	if len(threats) != maxThreatLogSize {
		t.Errorf("expected %d threats, got %d", maxThreatLogSize, len(threats))
	}
}

func TestEngine_RecentThreats_Order(t *testing.T) {
	engine := newTestEngine()

	engine.recordThreat(ScanResult{Filename: "first.exe"})
	engine.recordThreat(ScanResult{Filename: "second.exe"})
	engine.recordThreat(ScanResult{Filename: "third.exe"})

	threats := engine.RecentThreats()
	if len(threats) != 3 {
		t.Fatalf("expected 3 threats, got %d", len(threats))
	}
	// Should be reverse chronological (newest first).
	if threats[0].Filename != "third.exe" {
		t.Errorf("expected 'third.exe' first, got %q", threats[0].Filename)
	}
	if threats[2].Filename != "first.exe" {
		t.Errorf("expected 'first.exe' last, got %q", threats[2].Filename)
	}
}
