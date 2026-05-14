package antivirus

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Mock Scan Layer
// ─────────────────────────────────────────────────────────────────────────────

// mockLayer is a configurable test double for ScanLayer.
type mockLayer struct {
	name    string
	threats []ThreatInfo
	err     error
	called  int
}

func (m *mockLayer) Name() string { return m.name }

func (m *mockLayer) Scan(file *ScanTarget) ([]ThreatInfo, error) {
	m.called++
	if m.err != nil {
		return nil, m.err
	}
	return m.threats, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Engine Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestNewEngine_DefaultConfig(t *testing.T) {
	engine := NewEngine(nil)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
	if engine.cfg == nil {
		t.Fatal("engine config is nil")
	}
	if engine.LayerCount() != 0 {
		t.Errorf("expected 0 layers, got %d", engine.LayerCount())
	}
}

func TestNewEngine_CustomConfig(t *testing.T) {
	cfg := &Config{
		Enabled:     true,
		Workers:     2,
		QueueSize:   500,
		MaxFileSize: 1024,
	}
	engine := NewEngine(cfg)
	if engine.cfg.Workers != 2 {
		t.Errorf("expected 2 workers, got %d", engine.cfg.Workers)
	}
	if engine.cfg.QueueSize != 500 {
		t.Errorf("expected queue size 500, got %d", engine.cfg.QueueSize)
	}
}

func TestRegisterLayer(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true})

	layer1 := &mockLayer{name: "layer1"}
	layer2 := &mockLayer{name: "layer2"}

	engine.RegisterLayer(layer1)
	engine.RegisterLayer(layer2)

	if engine.LayerCount() != 2 {
		t.Errorf("expected 2 layers, got %d", engine.LayerCount())
	}
}

func TestRegisterLayer_DuplicateIgnored(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true})

	layer1 := &mockLayer{name: "same_name"}
	layer2 := &mockLayer{name: "same_name"}

	engine.RegisterLayer(layer1)
	engine.RegisterLayer(layer2)

	if engine.LayerCount() != 1 {
		t.Errorf("expected 1 layer (duplicate ignored), got %d", engine.LayerCount())
	}
}

func TestRegisterLayer_PanicsAfterStart(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true})
	engine.Start()
	defer engine.Shutdown(context.Background())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when registering layer after Start()")
		}
	}()

	engine.RegisterLayer(&mockLayer{name: "late"})
}

func TestScan_DisabledEngine(t *testing.T) {
	engine := NewEngine(&Config{Enabled: false})
	result, err := engine.Scan(context.Background(), []byte("test"), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictClean {
		t.Errorf("expected VerdictClean, got %s", result.Verdict)
	}
}

func TestScan_NoLayers_Clean(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), []byte("hello world"), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictClean {
		t.Errorf("expected VerdictClean, got %s", result.Verdict)
	}
	if result.SHA256 == "" {
		t.Error("SHA256 should be populated")
	}
	if result.FileSize != 11 {
		t.Errorf("expected file size 11, got %d", result.FileSize)
	}
}

func TestScan_CleanFile(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.RegisterLayer(&mockLayer{name: "clean_layer", threats: nil})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), []byte("safe content"), "safe.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictClean {
		t.Errorf("expected VerdictClean, got %s", result.Verdict)
	}
	if len(result.LayersRun) != 1 {
		t.Errorf("expected 1 layer run, got %d", len(result.LayersRun))
	}
}

func TestScan_MalwareDetected(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.RegisterLayer(&mockLayer{
		name: "hashdb",
		threats: []ThreatInfo{
			{
				Name:       "Trojan.Win32.Emotet.A",
				Category:   CategoryTrojan,
				Severity:   SeverityCritical,
				Layer:      LayerHashDB,
				Confidence: 1.0,
			},
		},
	})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), []byte("malicious"), "evil.exe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictMalware {
		t.Errorf("expected VerdictMalware, got %s", result.Verdict)
	}
	if len(result.Threats) != 1 {
		t.Fatalf("expected 1 threat, got %d", len(result.Threats))
	}
	if result.Threats[0].Name != "Trojan.Win32.Emotet.A" {
		t.Errorf("unexpected threat name: %s", result.Threats[0].Name)
	}
}

func TestScan_SuspiciousFile(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.RegisterLayer(&mockLayer{
		name: "heuristic",
		threats: []ThreatInfo{
			{
				Name:       "Heuristic.Packed.Unknown",
				Category:   CategoryPacker,
				Severity:   SeverityMedium,
				Layer:      LayerHeuristic,
				Confidence: 0.6,
			},
		},
	})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), []byte("packed"), "packed.exe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictSuspicious {
		t.Errorf("expected VerdictSuspicious, got %s", result.Verdict)
	}
}

func TestScan_MultipleLayerThreats(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.RegisterLayer(&mockLayer{
		name:    "hashdb",
		threats: []ThreatInfo{{Name: "Hash.Match", Confidence: 1.0, Layer: LayerHashDB}},
	})
	engine.RegisterLayer(&mockLayer{
		name:    "pattern",
		threats: []ThreatInfo{{Name: "Pattern.Match", Confidence: 0.9, Layer: LayerPattern}},
	})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), []byte("data"), "file.bin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Threats) != 2 {
		t.Errorf("expected 2 threats, got %d", len(result.Threats))
	}
	if result.Verdict != VerdictMalware {
		t.Errorf("expected VerdictMalware, got %s", result.Verdict)
	}
}

func TestScan_LayerError_ContinuesScanning(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.RegisterLayer(&mockLayer{name: "broken", err: fmt.Errorf("db connection failed")})
	engine.RegisterLayer(&mockLayer{
		name:    "working",
		threats: []ThreatInfo{{Name: "Found.It", Confidence: 0.9, Layer: LayerPattern}},
	})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), []byte("data"), "file.bin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LayersRun) != 2 {
		t.Errorf("expected 2 layers run, got %d", len(result.LayersRun))
	}
	if len(result.Threats) != 1 {
		t.Errorf("expected 1 threat from working layer, got %d", len(result.Threats))
	}
}

func TestScan_FileTooLarge(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: 10})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), make([]byte, 20), "big.bin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != VerdictClean {
		t.Errorf("oversized files should return VerdictClean (skipped), got %s", result.Verdict)
	}
}

func TestScan_ContextCancelled(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	slowLayer := &mockLayer{name: "slow"}
	engine.RegisterLayer(slowLayer)
	engine.Start()
	defer engine.Shutdown(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	result, err := engine.Scan(ctx, []byte("data"), "file.bin")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if result.Verdict != VerdictError {
		t.Errorf("expected VerdictError, got %s", result.Verdict)
	}
}

func TestScan_SHA256Computed(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.Start()
	defer engine.Shutdown(context.Background())

	result, err := engine.Scan(context.Background(), []byte("test"), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// SHA-256 of "test" is well-known
	expected := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if result.SHA256 != expected {
		t.Errorf("expected SHA256 %s, got %s", expected, result.SHA256)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Stats Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestStats_AfterScans(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true, MaxFileSize: DefaultMaxFileSize})
	engine.RegisterLayer(&mockLayer{name: "clean_layer"})
	engine.Start()
	defer engine.Shutdown(context.Background())

	for i := 0; i < 5; i++ {
		engine.Scan(context.Background(), []byte("data"), "file.txt")
	}

	stats := engine.Stats()
	if stats.TotalScanned != 5 {
		t.Errorf("expected 5 scans, got %d", stats.TotalScanned)
	}
	if stats.CleanFiles != 5 {
		t.Errorf("expected 5 clean, got %d", stats.CleanFiles)
	}
	if stats.EngineVersion != EngineVersion {
		t.Errorf("expected version %s, got %s", EngineVersion, stats.EngineVersion)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Type Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestScanVerdict_IsThreat(t *testing.T) {
	tests := []struct {
		verdict ScanVerdict
		threat  bool
	}{
		{VerdictClean, false},
		{VerdictMalware, true},
		{VerdictSuspicious, true},
		{VerdictError, false},
	}
	for _, tt := range tests {
		if got := tt.verdict.IsThreat(); got != tt.threat {
			t.Errorf("%s.IsThreat() = %v, want %v", tt.verdict, got, tt.threat)
		}
	}
}

func TestThreatSeverity_Weight(t *testing.T) {
	if SeverityCritical.Weight() <= SeverityHigh.Weight() {
		t.Error("critical should outweigh high")
	}
	if SeverityHigh.Weight() <= SeverityMedium.Weight() {
		t.Error("high should outweigh medium")
	}
	if SeverityMedium.Weight() <= SeverityLow.Weight() {
		t.Error("medium should outweigh low")
	}
}

func TestScanResult_HighestSeverity(t *testing.T) {
	r := &ScanResult{
		Threats: []ThreatInfo{
			{Severity: SeverityLow},
			{Severity: SeverityCritical},
			{Severity: SeverityMedium},
		},
	}
	if got := r.HighestSeverity(); got != SeverityCritical {
		t.Errorf("expected critical, got %s", got)
	}
}

func TestScanResult_ThreatNames_Deduplicated(t *testing.T) {
	r := &ScanResult{
		Threats: []ThreatInfo{
			{Name: "Trojan.A"},
			{Name: "Trojan.B"},
			{Name: "Trojan.A"}, // duplicate
		},
	}
	names := r.ThreatNames()
	if len(names) != 2 {
		t.Errorf("expected 2 unique names, got %d", len(names))
	}
}

func TestDetermineVerdict(t *testing.T) {
	tests := []struct {
		name    string
		threats []ThreatInfo
		want    ScanVerdict
	}{
		{"no threats", nil, VerdictClean},
		{"high confidence", []ThreatInfo{{Confidence: 0.95}}, VerdictMalware},
		{"threshold confidence", []ThreatInfo{{Confidence: 0.8}}, VerdictMalware},
		{"medium confidence", []ThreatInfo{{Confidence: 0.6}}, VerdictSuspicious},
		{"low confidence", []ThreatInfo{{Confidence: 0.3}}, VerdictSuspicious},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineVerdict(tt.threats); got != tt.want {
				t.Errorf("determineVerdict() = %s, want %s", got, tt.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Config Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestConfig_Validate_FixesInvalid(t *testing.T) {
	cfg := &Config{
		Enabled:   true,
		Workers:   -1,
		QueueSize: 0,
	}
	warnings := cfg.Validate()
	if len(warnings) == 0 {
		t.Error("expected validation warnings")
	}
	if cfg.Workers != 1 {
		t.Errorf("expected Workers corrected to 1, got %d", cfg.Workers)
	}
	if cfg.QueueSize != 100 {
		t.Errorf("expected QueueSize corrected to 100, got %d", cfg.QueueSize)
	}
}

func TestConfig_EnabledLayers(t *testing.T) {
	cfg := &Config{
		HashDBEnabled:    true,
		PatternEnabled:   false,
		HeuristicEnabled: true,
		YARAEnabled:      false,
		EntropyEnabled:   true,
	}
	layers := cfg.EnabledLayers()
	if len(layers) != 3 {
		t.Errorf("expected 3 enabled layers, got %d: %v", len(layers), layers)
	}
}

func TestParseQuarantineAction(t *testing.T) {
	tests := []struct {
		input string
		want  QuarantineAction
	}{
		{"tag", QuarantineTag},
		{"delete", QuarantineDelete},
		{"move", QuarantineMove},
		{"DELETE", QuarantineDelete},
		{"  Move  ", QuarantineMove},
		{"unknown", QuarantineTag},
		{"", QuarantineTag},
	}
	for _, tt := range tests {
		if got := ParseQuarantineAction(tt.input); got != tt.want {
			t.Errorf("ParseQuarantineAction(%q) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Lifecycle Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestEngine_StartShutdown(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true})
	engine.Start()
	if !engine.IsRunning() {
		t.Error("engine should be running after Start()")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := engine.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}

func TestEngine_DoubleStart(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true})
	engine.Start()
	engine.Start() // should not panic
	defer engine.Shutdown(context.Background())
}

func TestEngine_ShutdownWithoutStart(t *testing.T) {
	engine := NewEngine(&Config{Enabled: true})
	err := engine.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("shutdown without start should not error: %v", err)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{500, "500B"},
		{1500, "1.5KB"},
		{1500000, "1.4MB"},
		{1500000000, "1.4GB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.input); got != tt.want {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.input, got, tt.want)
		}
	}
}
