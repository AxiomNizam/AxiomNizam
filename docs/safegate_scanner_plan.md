# 🛡️ SafeGate Scanner — Restructuring Plan

## Overview

Restructure the `internal/scanner/` module from a flat, untested package into a
well-organized, modular system following the same architectural patterns
established by `internal/antivirus/`.

**Module**: `internal/scanner/` (SafeGate File Scanner Pipeline)  
**Started**: 2026-05-13  
**Status**: Phase 5 complete (5/6)

---

## Phase Summary

| Phase | Title | Status | Date |
|-------|-------|--------|------|
| 1 | Foundation — Types, Config & Tests | ✅ Complete | 2026-05-13 |
| 2 | Orchestrator Rewrite (context + parallel + metrics) | ✅ Complete | 2026-05-13 |
| 3 | Subpackage Extraction | ✅ Complete | 2026-05-13 |
| 4 | Individual Scanner Enhancements | ✅ Complete | 2026-05-13 |
| 5 | Observability & Metrics | ✅ Complete | 2026-05-13 |
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

## Phase 2: Orchestrator Rewrite (context + parallel + metrics) ✅

**Goal**: Add `context.Context` to Scanner interface, implement parallel
execution, add per-scanner timeout and timing metrics.

### Interface Change

Scanner interface updated from:
```go
Scan(file *FileInfo) ([]Finding, error)
```
to:
```go
Scan(ctx context.Context, file *FileInfo) ([]Finding, error)
```

### Files Modified

| File | Change |
|------|--------|
| `scanner.go` | Full rewrite — parallel + sequential execution paths, `ScanWithContext()`, context deadline enforcement, per-scanner timing |
| `types.go` | Added `ScannerTiming` struct and `Timings` field on `ScanResult` |
| `metadata.go` | `Scan()` now accepts `context.Context` |
| `mime.go` | `Scan()` now accepts `context.Context` |
| `svg.go` | `Scan()` now accepts `context.Context` |
| `macro.go` | `Scan()` now accepts `context.Context` |
| `archive.go` | `Scan()` now accepts `context.Context` |
| `native_av.go` | Propagates caller ctx to AV engine; adds safety-net timeout only when caller has no deadline |
| `scanner_test.go` | Updated mock, added `slowMockScanner`, +8 new tests |

### New Orchestrator Capabilities

| Feature | Detail |
|---------|--------|
| **Parallel execution** | `Config.Parallel=true` (default): scanners run concurrently via `sync.WaitGroup`. Output merged in registration order for determinism. |
| **Sequential fallback** | `Config.Parallel=false`: scanners run in order (useful for debugging). |
| **Single scanner optimization** | When only 1 scanner is registered, parallel mode is skipped (no goroutine overhead). |
| **Context propagation** | `ScanWithContext(ctx, file)` passes ctx to all scanners. `Scan(file)` wraps with `Config.Timeout` deadline. |
| **Timeout enforcement** | `Config.Timeout` creates deadline context. Shorter of caller deadline vs config timeout wins. |
| **Per-scanner timing** | `ScanResult.Timings[]` records execution time, finding count, error/timeout status per scanner. |
| **Error isolation** | Scanner errors produce info-level findings. In sequential mode, context cancellation stops remaining scanners. |

### New Type: ScannerTiming

```go
type ScannerTiming struct {
    Scanner      string  // Scanner name.
    DurationMs   int64   // Execution time in milliseconds.
    FindingCount int     // Number of findings produced.
    Error        bool    // True if the scanner returned an error.
    TimedOut     bool    // True if cancelled due to timeout.
}
```

### Test Coverage (8 new tests)

| Test | Validates |
|------|----------|
| `TestOrchestrator_Parallel_AllScannersRun` | All scanners execute and findings merge in parallel mode |
| `TestOrchestrator_Parallel_Timings` | Per-scanner timing recorded correctly in parallel mode |
| `TestOrchestrator_Sequential_Timings` | Error timings tracked in sequential mode |
| `TestOrchestrator_ScanWithContext` | ScanWithContext API works correctly |
| `TestOrchestrator_ScanWithContext_Cancellation` | Pre-cancelled context stops slow scanners immediately |
| `TestOrchestrator_Timeout_Sequential` | Config.Timeout aborts slow scanners within deadline |
| `TestOrchestrator_Parallel_FasterThanSequential` | Parallel mode runs 3 slow scanners faster than sequential |
| `TestOrchestrator_SingleScanner_NoParallel` | Single scanner skips parallel overhead |

### Backward Compatibility

✅ **Zero breaking changes for callers:**
- `Scan(file)` still works identically (internal context created)
- `NewOrchestrator(scanners...)` signature unchanged
- `ScanResult` gains `Timings` field (omitempty — invisible in JSON when empty)
- All handler code in `api_builder_handler.go` compiles without changes

### Build & Test Verification

```
go build -o NUL .              → ✅ Clean
go vet ./internal/scanner/...  → ✅ Clean
go test ./internal/scanner/... → ✅ 35/35 PASS (0.73s)
```

---

## Phase 3: Subpackage Extraction ✅

**Goal**: Move each scanner into its own subpackage for independent
testability and cleaner dependency graph.

### Final Directory Structure

```
internal/scanner/
├── scanner.go            (Scanner interface + Orchestrator)
├── types.go              (FileInfo, Finding, ScanResult, ScannerTiming, Severity)
├── config.go             (Config, DefaultConfig, LoadConfigFromEnv)
├── scanner_test.go       (35 tests: types, config, orchestrator, parallel)
├── metadata/metadata.go  (MetadataScanner → metadata.Scanner)
├── mimetype/mimetype.go  (MIMEScanner → mimetype.Scanner)
├── svg/svg.go            (SVGScanner → svg.Scanner)
├── macro/macro.go        (MacroScanner → macro.Scanner)
├── archivescan/archivescan.go  (ArchiveScanner → archivescan.Scanner)
└── native/native.go      (NativeAVScanner → native.Scanner)
```

### Deleted Files (replaced by subpackages)

| Old File | New Location |
|----------|-------------|
| `scanner/metadata.go` | `scanner/metadata/metadata.go` |
| `scanner/mime.go` | `scanner/mimetype/mimetype.go` |
| `scanner/svg.go` | `scanner/svg/svg.go` |
| `scanner/macro.go` | `scanner/macro/macro.go` |
| `scanner/archive.go` | `scanner/archivescan/archivescan.go` |
| `scanner/native_av.go` | `scanner/native/native.go` |

### Naming Decisions

| Subpackage | Rationale |
|------------|----------|
| `metadata` | Clear, no conflicts |
| `mimetype` | Avoids shadowing Go stdlib `mime` package |
| `svg` | Clear, no conflicts |
| `macro` | Clear, no conflicts |
| `archivescan` | Avoids shadowing Go stdlib `archive` package |
| `native` | Clear, no conflicts |

### Constructor Migration

```diff
- scanner.NewMetadataScanner(100*1024*1024)
+ metadata.NewScanner(cfg.MaxFileSize)

- scanner.NewMIMEScanner([]string{...})
+ mimetype.NewScanner(cfg.AllowedMIMETypes)

- &scanner.SVGScanner{}
+ svg.NewScanner()

- &scanner.MacroScanner{}
+ macro.NewScanner()

- scanner.NewArchiveScanner(5, 1024*1024*1024)
+ archivescan.NewScanner(cfg.ArchiveMaxDepth, cfg.ArchiveMaxDecompressedSize)

- scanner.NewNativeAVScanner(engine)
+ native.NewScanner(engine)
```

### Handler Integration Updated

`api_builder_handler.go` now:
- Imports 6 scanner subpackages alongside root `scanner` package
- Uses `scanner.LoadConfigFromEnv()` for centralized configuration
- Passes config values to subpackage constructors (no more hardcoded thresholds)
- Both `NewAPIBuilderHandler()` and `SetAVEngine()` updated

### Architecture Benefits

| Benefit | Detail |
|---------|--------|
| **Independent testability** | Each subpackage can have focused unit tests |
| **Cleaner dependency graph** | Subpackages only import root scanner for types — no circular deps |
| **Decoupled namespace** | Each scanner owns its helpers and regex patterns |
| **Config-driven** | Thresholds from centralized Config, no inline magic numbers |
| **Extensibility** | New scanners = new subpackage, no root package changes |

### Build & Test Verification

```
go build -o NUL .              → ✅ Clean
go test ./internal/scanner/... → ✅ 35/35 PASS (0.53s)
└─ 6 subpackages detected (no test files yet — Phase 6)
```

---

## Phase 4: Individual Scanner Enhancements ✅

**Goal**: Deep-dive security improvements for each scanner, adding detection
capabilities that address real-world attack vectors.

### Metadata Scanner Enhancements

| Enhancement | Detail |
|-------------|--------|
| **Sample-based null byte detection** | Configurable `NullByteSampleSize` (default 8192) — avoids O(n) scan on large files |
| **Path traversal detection** | Detects `../`, absolute paths, Windows drive letters (`C:\`), bare `..` |
| **Unicode/control char detection** | Bidi override chars (U+202A-E), zero-width spaces (U+200B-D), control chars |
| **Null bytes in filename** | Detects null byte truncation attacks in filenames |
| **Expanded text file recognition** | Added `.css`, `.js`, `.ts`, `.md`, `.yaml`, `.py`, `.go`, `.rs`, etc. |
| **Configurable filename length** | Uses `Config.MaxFilenameLength` instead of hardcoded 255 |
| **`NewScannerWithConfig()`** | New constructor accepting all 3 config parameters |

### MIME Scanner Enhancements

| Enhancement | Detail |
|-------------|--------|
| **WebAssembly detection** | `\x00asm` magic bytes — WASM can execute native-speed code in browsers |
| **Java class detection** | `0xCAFEBABE` magic bytes — compiled bytecode for JVM exploitation |
| **Shell script detection** | `#!` shebang — script files can execute system commands |
| **Table-driven format checks** | Refactored from monolithic `isExecutableSignature()` to `dangerousFormatChecks[]` |
| **Individual matchers** | `isPE()`, `isELF()`, `isMachO()`, `isWebAssembly()`, `isJavaClass()`, `isShellScript()` |
| **Extended CompatMap** | Added `application/wasm` and `application/java-archive` entries |

### SVG Scanner Enhancements

| Enhancement | Detail |
|-------------|--------|
| **CSS `@import` detection** | External stylesheet loading via `@import url()` |
| **CSS `url()` detection** | External resource loading via `url(https://...)` |
| **`<style>` element detection** | Inline CSS blocks that can contain injection vectors |
| **`<use>` external refs** | `<use href="https://...">` can load arbitrary SVG fragments |
| **`<use>` data: URIs** | `<use href="data:...">` can embed arbitrary SVG content |
| **`<iframe>` detection** | Iframes embedded in SVG via foreignObject |
| **`<embed>`/`<object>` detection** | Plugin/Flash/arbitrary content loading |

### Macro Scanner Enhancements

| Enhancement | Detail |
|-------------|--------|
| **DDE field code detection** | `DDE`/`DDEAUTO` field codes that execute commands without VBA macros |
| **DDE command execution** | `DDEAUTO cmd.exe` / `powershell` / `mshta` patterns |
| **ActiveX control detection** | `activeX*.xml`, `ActiveXData` patterns |
| **COM/CLSID detection** | `CLSID` and `classid=` references to COM objects |
| **Shell automation objects** | `Shell.Application`, `WScript.Shell`, `Scripting.FileSystemObject` |
| **OLE embedded objects** | OLE stream names (`\x01Ole`, `ObjectPool`, `\x01CompObj`) |
| **`isAnyOffice()` helper** | DDE/ActiveX checks apply to both legacy and modern Office |

### Archive Scanner Enhancements

| Enhancement | Detail |
|-------------|--------|
| **TAR analysis** | Full `archive/tar` parsing — path traversal, executables, size limits |
| **GZIP bomb detection** | Limited decompression read via `io.LimitReader` |
| **BZ2 bomb detection** | Limited decompression read via `io.LimitReader` |
| **TAR-in-GZIP** | `.tar.gz` recursive analysis (decompress GZIP, then analyze TAR) |
| **TAR-in-BZ2** | `.tar.bz2` recursive analysis |
| **Symlink detection (ZIP)** | File mode bit check for symlinks in ZIP entries |
| **Symlink detection (TAR)** | `tar.TypeSymlink` / `tar.TypeLink` detection |
| **Symlink bomb** | >50 symlinks triggers high-severity finding |
| **Symlink target traversal** | Validates symlink targets for path traversal |
| **Extended executable list** | Added `.dll`, `.sys`, `.cpl`, `.hta`, `.inf`, `.reg` |
| **`.tgz` support** | Recognized as GZIP archive |
| **BZ2 magic bytes** | `0x42 0x5A 0x68` ("BZh") detection |
| **TAR magic bytes** | `ustar` at offset 257 |

### Build & Test Verification

```
go build -o NUL .              → ✅ Clean
go test ./internal/scanner/... → ✅ 35/35 PASS (0.53s)
```

---

## Phase 5: Observability & Metrics ✅

**Goal**: Add pipeline-wide metrics collection, scan throughput tracking,
finding severity distribution, and a structured health endpoint.

### New Files

| File | Purpose |
|------|--------|
| `metrics.go` | Thread-safe `Metrics` collector, `MetricsSnapshot`, `HealthStatus`, `ScannerMetrics`, `AtomicCounter` |
| `metrics_test.go` | 22 tests covering all metrics/health functionality |

### Metrics Collector (`Metrics`)

| Capability | Detail |
|------------|--------|
| **Scan throughput** | `totalScans`, `totalSafe`, `totalUnsafe` counters |
| **Finding distribution** | `bySeverity` map tracking counts per `Severity` level |
| **Per-scanner stats** | `scannerScans`, `scannerFindings`, `scannerErrors`, `scannerTimeouts`, `scannerTotalMs` |
| **Timing aggregation** | `totalDurationMs`, `maxDurationMs`, `minDurationMs` across all scans |
| **Thread safety** | `sync.RWMutex` — concurrent reads via `RLock`, exclusive writes via `Lock` |
| **Auto-recording** | `Record(result)` called automatically by orchestrator after every scan |

### MetricsSnapshot (JSON-serializable)

```json
{
  "total_scans": 1542,
  "total_safe": 1500,
  "total_unsafe": 42,
  "safety_rate": "97.3%",
  "total_findings": 87,
  "findings_by_severity": { "critical": 3, "high": 12, "medium": 30, "low": 42 },
  "avg_duration_ms": 45,
  "max_duration_ms": 1200,
  "min_duration_ms": 5,
  "uptime_seconds": 86400,
  "scanners": [
    { "name": "metadata_scanner", "total_runs": 1542, "avg_ms": 2 },
    { "name": "native_antivirus", "total_runs": 1542, "avg_ms": 35 }
  ]
}
```

### Health Endpoint (`HealthStatus`)

| Field | Detail |
|-------|--------|
| `status` | `"healthy"` / `"degraded"` (>50% errors) / `"unavailable"` (>90% errors or no scanners) |
| `scanner_count` | Number of registered scanners |
| `scanners` | Names of all registered scanners |
| `total_scans` | Scans since startup |
| `error_rate` | Formatted percentage (e.g. `"0.5%"`) |
| `metrics` | Full `MetricsSnapshot` when `includeMetrics=true` |

### Orchestrator Integration

| Change | Detail |
|--------|--------|
| `metrics` field added to `Orchestrator` struct | Auto-initialized in both constructors |
| `ScanWithContext()` calls `metrics.Record(result)` | Every scan automatically collected |
| `Metrics()` method | Returns the metrics collector for snapshot access |
| `Health(includeMetrics)` method | Returns structured health status with optional full metrics |

### Test Coverage (22 new tests)

| Test Category | Tests |
|---------------|-------|
| **Metrics.Record** | SafeScan, UnsafeScan, DurationTracking, PerScanner, Timeouts |
| **Metrics.Snapshot** | Full snapshot, NoScans edge case |
| **Metrics.Reset** | All fields cleared |
| **Metrics.TotalScans** | Thread-safe accessor |
| **Orchestrator integration** | AutoRecord, FindingsTracked, PerScannerTimings, WithContext |
| **Health endpoint** | Healthy, WithMetrics, NoScanners, ErrorRate, ScannerList |
| **AtomicCounter** | Inc, Add, Reset |

### Build & Test Verification

```
go build -o NUL .              → ✅ Clean
go test ./internal/scanner/... → ✅ 57/57 PASS (0.54s)
└─ 35 existing + 22 new metrics/health tests
```

---

## Phase 6: Full Test Suite & Documentation ⏳

- Per-scanner test files with real file samples
- Benchmark tests for performance regression
- Architecture documentation update
