# 🛡️ SafeGate Scanner — Restructuring Plan

## Overview

Restructure the `internal/scanner/` module from a flat, untested package into a
well-organized, modular system following the same architectural patterns
established by `internal/antivirus/`.

**Module**: `internal/scanner/` (SafeGate File Scanner Pipeline)  
**Started**: 2026-05-13  
**Status**: Phase 1 complete (1/6)

---

## Phase Summary

| Phase | Title | Status | Date |
|-------|-------|--------|------|
| 1 | Foundation — Types, Config & Tests | ✅ Complete | 2026-05-13 |
| 2 | Orchestrator Rewrite (context + parallel + metrics) | ⏳ Pending | — |
| 3 | Subpackage Extraction | ⏳ Pending | — |
| 4 | Individual Scanner Enhancements | ⏳ Pending | — |
| 5 | Observability & Metrics | ⏳ Pending | — |
| 6 | Full Test Suite & Docs | ⏳ Pending | — |

---

## Phase 1: Foundation — Types, Config & Tests ✅

**Goal**: Extract shared types into dedicated file, create centralized
configuration with environment variable support, establish comprehensive
test baseline. Zero breaking changes.

### Files Created

| File | Lines | Purpose |
|------|-------|---------|
| `types.go` | 152 | Extracted and enhanced types: `Severity`, `Finding`, `ScanResult`, `FileInfo` |
| `config.go` | 199 | Centralized `Config` struct with `DefaultConfig()` and `LoadConfigFromEnv()` |
| `scanner_test.go` | 430+ | 27 tests covering types, config, and orchestrator |

### Files Modified

| File | Change |
|------|--------|
| `scanner.go` | Rewritten as slim orchestrator core. Types moved to `types.go`. Added `NewOrchestratorWithConfig()`, `Config()`, `ScannerCount()`. Uses `Severity.IsThreat()` instead of inline severity check. |

### Types Enhancements (types.go)

**Severity type gains utility methods:**
- `Weight() int` — Numeric severity weight (5=critical → 1=info) for comparison/sorting
- `IsThreat() bool` — Returns true for severity ≥ medium (threat threshold)

**Finding gains new optional fields:**
- `Offset int64` — Byte offset where issue was found (for future scanner precision)
- `Category string` — Classification category (e.g. "xss", "macro", "bomb")

**ScanResult gains query methods:**
- `Summary() string` — Human-readable one-line scan summary (CLEAN/THREAT format)
- `FindingsBySeverity(sev)` — Filter findings by severity level
- `ThreatFindings()` — Returns only medium+ findings
- `ScannerRan(name)` — Check if a specific scanner was executed (case-insensitive)

### Configuration (config.go)

**10 environment variables** for all previously-hardcoded thresholds:

| Variable | Default | Description |
|----------|---------|-------------|
| `SCANNER_TIMEOUT` | `2m` | Max orchestrator scan duration |
| `SCANNER_PARALLEL` | `true` | Run scanners concurrently |
| `SCANNER_MAX_FILE_SIZE` | `104857600` (100MB) | Max file size |
| `SCANNER_NULL_BYTE_SAMPLE_SIZE` | `8192` | Bytes to sample for null check |
| `SCANNER_MAX_FILENAME_LENGTH` | `255` | Max filename length |
| `SCANNER_ALLOWED_MIME_TYPES` | (built-in list) | Comma-separated MIME types |
| `SCANNER_ARCHIVE_MAX_DEPTH` | `5` | Max archive nesting depth |
| `SCANNER_ARCHIVE_MAX_DECOMPRESS` | `1073741824` (1GB) | Max decompressed size |
| `SCANNER_ARCHIVE_MAX_FILES` | `10000` | Max entries in archive |
| `SCANNER_ARCHIVE_RATIO_LIMIT` | `100` | Max compression ratio before bomb alert |

**Key design decisions:**
- `DefaultConfig()` returns production-safe values — zero-config deployment
- `LoadConfigFromEnv()` overrides only fields with valid env vars — invalid values silently use defaults
- Built-in MIME type list covers text, documents, archives, Office (legacy + OOXML), images, and media
- Config is stored on `Orchestrator` for downstream scanner access in Phase 2+

### Test Coverage (scanner_test.go)

**27 tests, all passing:**
- Severity: Weight values, strict ordering, IsThreat boundary
- ScanResult: Summary (clean + threat format), FindingsBySeverity, ThreatFindings, ScannerRan, edge cases (nil, empty)
- Config: DefaultConfig values, AllowedMIMETypes content, LoadConfigFromEnv overrides, invalid env values
- Orchestrator: Clean scan, threat scan, error handling, info/low safety, medium unsafety, multiple findings, file metadata propagation, empty orchestrator, WithConfig

### Backward Compatibility

✅ **Zero breaking changes:**
- `scanner.FileInfo`, `scanner.Finding`, `scanner.ScanResult`, `scanner.Severity` all work identically
- `scanner.NewOrchestrator(scanners...)` has identical signature and behavior
- `scanner.Orchestrator.Scan(file)` returns identical `*ScanResult`
- `scanner.ScannerNames()` returns identical `[]string`
- All existing callers in `api_builder_handler.go` and `main.go` compile without changes

### Build & Test Verification

```
go build -o NUL .            → ✅ Clean
go vet ./internal/scanner/...  → ✅ Clean
go test ./internal/scanner/... → ✅ 27/27 PASS (0.48s)
```

---

## Phase 2: Orchestrator Rewrite ⏳

**Goal**: Add `context.Context` to Scanner interface, implement parallel
execution, add per-scanner timeout and timing metrics.

**Interface change:**
```go
type Scanner interface {
    Name() string
    Scan(ctx context.Context, file *FileInfo) ([]Finding, error)
}
```

**Key changes planned:**
- All 6 scanners + NativeAVScanner gain `ctx` parameter
- Orchestrator uses `errgroup` for parallel fan-out
- Per-scanner timing recorded in result metadata
- Config.Timeout enforced via context deadline
- Config.Parallel controls concurrent vs sequential mode

---

## Phase 3: Subpackage Extraction ⏳

**Goal**: Move each scanner into its own subpackage for independent
testability and cleaner dependency graph.

**Target structure:**
```
internal/scanner/
├── scanner.go          (interface + orchestrator)
├── types.go            (shared types)
├── config.go           (configuration)
├── metadata/metadata.go
├── mime/mime.go + compat.go
├── svg/svg.go
├── macro/macro.go
├── archive/archive.go
├── native/native_av.go
└── scanner_test.go
```

---

## Phase 4: Individual Scanner Enhancements ⏳

| Scanner | Enhancements |
|---------|-------------|
| Metadata | Sample-based null byte detection, Unicode/control char detection, path traversal in filenames |
| MIME | Configurable compat map, WebAssembly/Java class magic bytes |
| SVG | CSS `@import`/`url()` detection, `<use>` external reference detection |
| Macro | OLE2 stream-based analysis, DDE link detection, ActiveX detection |
| Archive | TAR/GZIP/BZ2 analysis, symlink bomb detection |

---

## Phase 5: Observability & Metrics ⏳

- Per-scanner execution time tracking
- Scan throughput counters
- Finding severity distribution
- Health endpoint enhancements

---

## Phase 6: Full Test Suite & Documentation ⏳

- Per-scanner test files with real file samples
- Benchmark tests for performance regression
- Architecture documentation update
