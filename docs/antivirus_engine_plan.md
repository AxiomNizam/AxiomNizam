# 🛡️ AxiomNizam Native Antivirus Engine — Implementation Plan

## Problem Statement

ClamAV consumes **~1–2 GB RAM** at idle (signature database loaded into memory) and requires a separate Docker container. We need a **lightweight, Go-native antivirus engine** built directly into the AxiomNizam platform, following the same module pattern as `internal/iam/`, `internal/storage/`, and `internal/trivy/`.

---

## How ClamAV Works (What We're Replacing)

ClamAV's `libclamav` is a ~80K-line C engine. Its core components are:

| ClamAV Component | What It Does | RAM Cost |
|-----------------|-------------|----------|
| **Signature Database** (`.cvd` files) | ~12M+ malware signatures loaded into memory | **~800MB–1.5GB** |
| **Aho-Corasick Automaton** | Compiles all byte-pattern signatures into a single finite-state automaton for single-pass matching | **~400MB** |
| **Hash Lookup Tables** | SHA256/MD5 hash sets for known-malware instant lookup | ~50MB |
| **File Type Identification** | Magic bytes → file type → determines which scanners to run | Negligible |
| **Unpackers / Decoders** | ZIP, RAR, 7z, PE/ELF unpacking, PDF/OLE parsing, UPX decompression | On-demand |
| **YARA Engine** | Pattern-matching rule engine for advanced signatures | ~50MB |
| **Bytecode VM** | Sandboxed VM to run detection scripts | Negligible |

> [!IMPORTANT]
> We are **NOT** trying to replicate ClamAV's full 12M-signature database. That would require the same RAM. Instead, we build a **focused, multi-layer detection engine** optimized for object storage security — targeting the threats most relevant to file uploads.

---

## Architecture: What We Build Instead

Our engine uses a **layered defense** strategy. Each layer is lightweight and catches a different class of threat. Combined, they provide strong coverage with **<50MB RAM**.

```mermaid
graph TB
    subgraph "internal/antivirus"
        direction TB
        A["engine.go<br/>Engine Orchestrator"] --> B["hashdb.go<br/>Known-Malware Hash DB"]
        A --> C["signature.go<br/>Byte-Pattern Matcher<br/>(Aho-Corasick)"]
        A --> D["heuristic.go<br/>Behavioral Heuristics"]
        A --> E["yara.go<br/>YARA Rule Engine"]
        A --> F["entropy.go<br/>Entropy Analysis"]
        
        G["sigdb/loader.go<br/>Signature Database Loader"] --> B
        G --> C
        G --> E
        
        H["updater.go<br/>Signature Auto-Updater"] --> G
    end
    
    subgraph "Existing Scanners (internal/scanner)"
        I["mime.go — MIME validation"]
        J["archive.go — Zip bomb detection"]
        K["macro.go — Office/PDF macros"]
        L["svg.go — SVG XSS"]
        M["metadata.go — Filename/size checks"]
    end
    
    A -.-> I
    A -.-> J
    A -.-> K
    A -.-> L
    A -.-> M
    
    style A fill:#ff6b6b,color:#fff
    style B fill:#4ecdc4,color:#fff
    style C fill:#45b7d1,color:#fff
    style D fill:#f9ca24,color:#333
    style E fill:#a29bfe,color:#fff
    style F fill:#fd79a8,color:#fff
```

---

## Detection Layers (ClamAV Feature Mapping)

| Layer | ClamAV Equivalent | Our Implementation | RAM | Detection Target |
|-------|-------------------|-------------------|-----|-----------------|
| **1. Hash DB** | `.hdb` / `.hsb` hash sigs | SHA-256 hash set (bloom filter + map) | ~5MB for 500K hashes | Known malware (exact match) |
| **2. Byte Patterns** | `.ndb` body sigs + AC automaton | Aho-Corasick automaton (Go-native) | ~10–20MB for ~50K patterns | Malware byte signatures |
| **3. Heuristic Rules** | `cli_scan_*` heuristic checks | Go functions (PE header analysis, shellcode patterns, script obfuscation) | Negligible | Zero-day / obfuscated threats |
| **4. YARA Rules** | YARA engine integration | Go YARA library or lightweight rule matcher | ~5MB for ~1K rules | Community threat rules |
| **5. Entropy Analysis** | Used internally for packed EXE detection | Shannon entropy calculator | Negligible | Packed/encrypted malware |
| **6. Existing Scanners** | N/A (we already have these) | `internal/scanner/*` (MIME, archive, macro, SVG, metadata) | Negligible | File-type-specific threats |

**Total estimated RAM: ~20–40MB** (vs ClamAV's ~1–1.5GB)

---

## Package Structure

Following the existing AxiomNizam module pattern (`internal/iam/`, `internal/storage/`, `internal/trivy/`):

```
internal/antivirus/
├── engine.go              # Engine struct, Scan() method, orchestrator
├── engine_test.go         # Engine integration tests
├── config.go              # Configuration (env vars, feature flags)
├── types.go               # ScanResult, ThreatInfo, ScanVerdict types
│
├── hashdb/                # Layer 1: Known-malware hash lookup
│   ├── hashdb.go          # Bloom filter + SHA-256 hash set
│   ├── hashdb_test.go
│   └── loader.go          # Load .hdb/.hsb hash files
│
├── matcher/               # Layer 2: Byte-pattern matching (Aho-Corasick)
│   ├── ahocorasick.go     # Go Aho-Corasick implementation
│   ├── ahocorasick_test.go
│   ├── signature.go       # Signature format parser
│   └── patterns.go        # Built-in critical patterns (PE, ELF, shellcode)
│
├── heuristic/             # Layer 3: Behavioral heuristics
│   ├── pe.go              # PE/EXE header analysis
│   ├── elf.go             # ELF binary analysis
│   ├── script.go          # Script obfuscation detection (JS, PS1, VBS)
│   ├── shellcode.go       # Shellcode/NOP sled detection
│   └── heuristic_test.go
│
├── yara/                  # Layer 4: YARA rule engine
│   ├── engine.go          # YARA rule compiler and matcher
│   ├── rules.go           # Built-in YARA rules
│   └── engine_test.go
│
├── entropy/               # Layer 5: Entropy analysis
│   ├── entropy.go         # Shannon entropy calculator
│   ├── packer.go          # Packer/crypter detection
│   └── entropy_test.go
│
├── sigdb/                 # Signature database management
│   ├── database.go        # Unified signature DB (load, query, stats)
│   ├── updater.go         # Auto-update from remote source
│   ├── format.go          # Signature file format parsers
│   └── builtin.go         # Compiled-in essential signatures
│
├── cache/                 # Scan result cache
│   ├── cache.go           # LRU cache keyed by SHA-256
│   └── cache_test.go
│
└── handler.go             # HTTP API handler (GET /antivirus/status, etc.)
```

---

## Implementation Phases

### Phase 1: Foundation — Types, Config, Engine Shell
**Files:** `types.go`, `config.go`, `engine.go`

Create the core types and engine orchestrator:

```go
// types.go
type ScanVerdict string

const (
    VerdictClean       ScanVerdict = "clean"
    VerdictMalware     ScanVerdict = "malware"
    VerdictSuspicious  ScanVerdict = "suspicious"
    VerdictError       ScanVerdict = "error"
)

type ThreatInfo struct {
    Name        string      `json:"name"`
    Category    string      `json:"category"`    // trojan, worm, ransomware, exploit, etc.
    Severity    string      `json:"severity"`    // critical, high, medium, low
    Layer       string      `json:"layer"`       // hashdb, pattern, heuristic, yara, entropy
    Description string      `json:"description"`
    Signature   string      `json:"signature,omitempty"`
    Confidence  float64     `json:"confidence"`  // 0.0–1.0
}

type ScanResult struct {
    Verdict     ScanVerdict  `json:"verdict"`
    Threats     []ThreatInfo `json:"threats"`
    SHA256      string       `json:"sha256"`
    FileSize    int64        `json:"fileSize"`
    FileType    string       `json:"fileType"`
    ScannedAt   time.Time    `json:"scannedAt"`
    DurationMs  int64        `json:"durationMs"`
    LayersRun   []string     `json:"layersRun"`
    CacheHit    bool         `json:"cacheHit"`
    EngineVer   string       `json:"engineVersion"`
    SigDBVer    string       `json:"sigDbVersion"`
}
```

**Estimated effort:** 1–2 hours

---

### Phase 2: Hash Database — Known Malware Lookup
**Files:** `hashdb/hashdb.go`, `hashdb/loader.go`

This is the fastest detection layer. A SHA-256 hash of the uploaded file is checked against a set of known-malware hashes.

**How ClamAV does it:** ClamAV loads ~5M hashes from `.hdb`/`.hsb` files into hash tables. Each hash maps to a malware name.

**Our implementation:**
- **Bloom filter** for fast negative lookups (99.99% of files are clean → skip in ~1μs)
- **`map[string]string`** for confirmed positives (SHA-256 → malware name)
- Load from ClamAV-compatible `.hdb` format OR our own JSON format
- Support incremental updates (add/remove hashes without full reload)

**RAM:** ~5MB for 500K hashes (bloom filter + map)

**Estimated effort:** 2–3 hours

---

### Phase 3: Byte-Pattern Matcher (Aho-Corasick)
**Files:** `matcher/ahocorasick.go`, `matcher/signature.go`, `matcher/patterns.go`

This is the **core detection engine** — equivalent to ClamAV's AC automaton. Scans file content for malware byte patterns in a single pass.

**How ClamAV does it:** Compiles ~8M signatures into a deterministic finite-state automaton. Each byte of input transitions the automaton; matches trigger alerts. Uses ~400MB RAM for the full DB.

**Our implementation:**
- Pure Go Aho-Corasick with **curated, high-value signatures only** (~10K–50K patterns)
- Focus on: PE malware, shellcode, exploit payloads, ransomware markers, cryptominers
- Support ClamAV `.ndb` format for signature import
- Built-in patterns for the most critical threats (compiled into the binary)
- **Streaming scan** — process file in chunks, no need to load entire file

**RAM:** ~10–20MB for 50K patterns (vs ClamAV's ~400MB for 8M patterns)

**Why this works:** The top ~50K signatures cover ~95% of real-world threats. The long tail of 8M+ signatures catches extremely rare variants that are less relevant for object storage security.

**Estimated effort:** 4–6 hours

---

### Phase 4: Behavioral Heuristics
**Files:** `heuristic/pe.go`, `heuristic/elf.go`, `heuristic/script.go`, `heuristic/shellcode.go`

Zero-day detection that doesn't rely on signatures. Analyzes file **structure and behavior** patterns.

**Detections:**

| Heuristic | What It Catches | How |
|-----------|----------------|-----|
| **PE Header Analysis** | Packed/modified Windows executables | Section entropy, import table anomalies, header corruption |
| **ELF Analysis** | Malicious Linux binaries | Suspicious sections, stripped symbols + executable stack |
| **Script Obfuscation** | Obfuscated JS/PS1/VBS | Char code arrays, eval chains, base64 nesting depth |
| **Shellcode Detection** | NOP sleds, syscall patterns | Known shellcode byte sequences, syscall instruction patterns |
| **Polyglot Files** | Files valid as 2+ types | Conflicting magic bytes (e.g., valid JPEG that's also valid JS) |
| **Embedded Executables** | EXE inside images/docs | PE/ELF magic bytes at non-standard offsets |

**RAM:** Negligible (pure logic, no data structures)

**Estimated effort:** 4–6 hours

---

### Phase 5: Entropy Analysis
**Files:** `entropy/entropy.go`, `entropy/packer.go`

Packed/encrypted malware has high entropy (close to 8.0 bits/byte for random data). Clean files have lower, predictable entropy.

**Detections:**
- Overall file entropy (flag if > 7.5 for executables)
- Per-section entropy for PE files (flag packed sections)
- Entropy distribution anomalies (legitimate encrypted files vs packed malware)
- Known packer signatures (UPX, Themida, VMProtect headers)

**RAM:** Negligible

**Estimated effort:** 2–3 hours

---

### Phase 6: YARA Rule Engine
**Files:** `yara/engine.go`, `yara/rules.go`

Lightweight YARA-compatible rule matcher for community threat intelligence rules.

**Implementation options:**
1. **Pure Go YARA parser** — parse YARA rule syntax, compile to our AC automaton
2. **Simplified subset** — support `strings` + `condition` blocks only (covers 90% of YARA rules)

**Built-in rule categories:**
- Ransomware detection (file markers, ransom notes)
- Cryptominer detection (mining pool URLs, XMRig strings)
- Webshell detection (PHP/JSP/ASPX backdoors)
- Exploit kit detection (known exploit payloads)

**RAM:** ~5MB for ~1K rules

**Estimated effort:** 4–6 hours

---

### Phase 7: Signature Database & Auto-Updater
**Files:** `sigdb/database.go`, `sigdb/updater.go`, `sigdb/format.go`, `sigdb/builtin.go`

**Database structure:**
```
/data/antivirus/
├── axiom-main.sig          # Primary signature database (hash + patterns)
├── axiom-heuristic.rules   # Heuristic rule configuration
├── axiom-yara.yar          # YARA rules
├── version.json            # DB version, last update timestamp
└── custom/                 # User-uploaded custom rules
    ├── custom-hashes.txt
    └── custom-yara.yar
```

**Auto-updater:**
- Configurable update URL (`ANTIVIRUS_UPDATE_URL`)
- Check for updates every N hours (`ANTIVIRUS_UPDATE_INTERVAL`, default: 6h)
- Download delta updates (only new signatures)
- Hot-reload signatures without engine restart (RWMutex swap)
- Fallback to built-in signatures if update fails

**Built-in signatures:** ~5K essential patterns compiled into the Go binary, so the engine works even without any external database.

**Estimated effort:** 3–4 hours

---

### Phase 8: Scan Cache
**Files:** `cache/cache.go`

- LRU cache keyed by SHA-256 hash
- If a file was scanned recently and is clean, skip re-scanning
- Configurable cache size (`ANTIVIRUS_CACHE_SIZE`, default: 100K entries)
- Cache TTL (`ANTIVIRUS_CACHE_TTL`, default: 24h)
- Invalidate on signature DB update

**RAM:** ~10MB for 100K cached results

**Estimated effort:** 1–2 hours

---

### Phase 9: Wire Into Storage System
**Files:** Modify `internal/storage/storage.go`, `internal/storage/admin/admin.go`

Same wiring pattern as the previous file scanner plan:
- Add `antivirus.Engine` to `storage.System`
- Initialize in `NewSystem()`
- Async scan on `PutObject()` (non-blocking)
- Scan result → object metadata tagging + notifications
- API endpoints for scan status

**Estimated effort:** 2–3 hours

---

### Phase 10: API & Admin Dashboard
**Files:** `handler.go`, frontend updates

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /antivirus/status` | GET | Engine status, signature DB version, stats |
| `GET /antivirus/stats` | GET | Scan statistics (total scanned, threats found, cache hit rate) |
| `POST /antivirus/scan` | POST | Manual scan of a specific object |
| `GET /antivirus/threats` | GET | List recent threat detections |
| `POST /antivirus/update` | POST | Trigger signature database update |
| `PUT /antivirus/config` | PUT | Update engine configuration |

**Estimated effort:** 2–3 hours

---

## ClamAV vs AxiomNizam Engine Comparison

| Metric | ClamAV | AxiomNizam Engine |
|--------|--------|-------------------|
| **Language** | C/Rust (~100K lines) | Go (~3–5K lines) |
| **RAM at idle** | ~1–1.5 GB | ~20–40 MB |
| **Signatures** | ~12M (full virus DB) | ~50K curated + heuristics |
| **Detection approach** | Primarily signature-based | Multi-layer (hash + pattern + heuristic + YARA + entropy) |
| **Zero-day detection** | Limited (bytecode VM) | Strong (heuristic + entropy analysis) |
| **Startup time** | ~30–60s (load DB) | ~1–2s |
| **Deployment** | Separate container | Built into main binary |
| **Update mechanism** | `freshclam` daemon | Built-in HTTP updater |
| **Scan speed** | ~50MB/s | ~100MB/s (fewer patterns, Go-optimized) |
| **Object storage optimized** | No (general purpose) | Yes (focused on upload threats) |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ANTIVIRUS_ENABLED` | `true` | Enable/disable the engine |
| `ANTIVIRUS_WORKERS` | `4` | Concurrent scan workers |
| `ANTIVIRUS_QUEUE_SIZE` | `10000` | Async scan queue size |
| `ANTIVIRUS_MAX_FILE_SIZE` | `104857600` | Max file size to scan (100MB) |
| `ANTIVIRUS_CACHE_SIZE` | `100000` | Scan result cache entries |
| `ANTIVIRUS_CACHE_TTL` | `24h` | Cache entry TTL |
| `ANTIVIRUS_UPDATE_URL` | _(empty)_ | Signature update URL |
| `ANTIVIRUS_UPDATE_INTERVAL` | `6h` | Auto-update check interval |
| `ANTIVIRUS_SIG_DIR` | `/data/antivirus` | Signature database directory |
| `ANTIVIRUS_HASH_DB_ENABLED` | `true` | Enable hash DB layer |
| `ANTIVIRUS_PATTERN_ENABLED` | `true` | Enable byte-pattern layer |
| `ANTIVIRUS_HEURISTIC_ENABLED` | `true` | Enable heuristic layer |
| `ANTIVIRUS_YARA_ENABLED` | `true` | Enable YARA layer |
| `ANTIVIRUS_ENTROPY_ENABLED` | `true` | Enable entropy layer |
| `ANTIVIRUS_QUARANTINE_ACTION` | `tag` | What to do on detection: `tag`, `delete`, `move` |
| `ANTIVIRUS_WEBHOOK_URL` | _(empty)_ | Webhook for threat notifications |

---

## Execution Order

| Phase | Component | Files | Effort | Dependency | Status |
|-------|-----------|-------|--------|------------|--------|
| **1** | Foundation (types, config, engine shell) | 4 files | 1–2h | None | ✅ Done |
| **2** | Hash Database | 3 files | 2–3h | Phase 1 | ✅ Done |
| **3** | Byte-Pattern Matcher (Aho-Corasick) | 4 files | 4–6h | Phase 1 | ✅ Done |
| **4** | Behavioral Heuristics | 6 files | 4–6h | Phase 1 | ✅ Done |
| **5** | Entropy Analysis | 2 files | 2–3h | Phase 1 | ⬜ Pending |
| **6** | YARA Rule Engine | 2 files | 4–6h | Phase 3 | ⬜ Pending |
| **7** | Signature DB & Updater | 4 files | 3–4h | Phase 2, 3, 6 | ⬜ Pending |
| **8** | Scan Cache | 1 file | 1–2h | Phase 1 | ⬜ Pending |
| **9** | Storage System Integration | 2 modified | 2–3h | Phase 1–8 | ⬜ Pending |
| **10** | API & Dashboard | 1+ files | 2–3h | Phase 9 | ⬜ Pending |

> [!TIP]
> **Phases 1–5 are independently testable** and provide value on their own. Phase 3 (Aho-Corasick) is the highest-impact single phase. Phase 4 (heuristics) provides zero-day coverage that ClamAV itself struggles with.

> [!NOTE]
> The existing `internal/scanner/` package (MIME, archive, macro, SVG, metadata scanners) remains untouched and continues to run alongside. The new `internal/antivirus/` engine adds **malware-specific** detection on top.

---

## What We're NOT Building (And Why)

| ClamAV Feature | Why We Skip It |
|---------------|---------------|
| Full 12M signature DB | Requires ~1GB RAM — defeats the purpose |
| Bytecode VM | Complex, niche — heuristics cover the same ground |
| RAR/7z deep unpacking | Already handled by `internal/scanner/archive.go` |
| Email MIME parsing | Not relevant for object storage uploads |
| Network stream scanning | Not needed — we scan at-rest objects |
| Phishing URL detection | Out of scope for file scanning |

---

## File Manifest (Complete)

| File | Action | Phase | Status | Description |
|------|--------|-------|--------|-------------|
| `internal/antivirus/types.go` | CREATE | 1 | ✅ | Core types: ScanVerdict, ThreatInfo, ScanResult, ScanLayer interface |
| `internal/antivirus/config.go` | CREATE | 1 | ✅ | Configuration from 17 env vars with validation |
| `internal/antivirus/engine.go` | CREATE | 1 | ✅ | Engine orchestrator, Scan(), lifecycle, atomic stats |
| `internal/antivirus/engine_test.go` | CREATE | 1 | ✅ | 25 tests — engine lifecycle, scan flows, types, config |
| `internal/antivirus/hashdb/hashdb.go` | CREATE | 2 | ✅ | Bloom filter (Kirsch-Mitzenmacker) + SHA-256 map, ~1.1MB for 500K hashes |
| `internal/antivirus/hashdb/loader.go` | CREATE | 2 | ✅ | ClamAV .hdb/.hsb, JSON, plain text format loaders |
| `internal/antivirus/hashdb/hashdb_test.go` | CREATE | 2 | ✅ | 27 tests — bloom FP rate, concurrency, all 3 formats, category inference |
| `internal/antivirus/matcher/ahocorasick.go` | CREATE | 3 | ✅ | Aho-Corasick automaton + Layer + Builder |
| `internal/antivirus/matcher/signature.go` | CREATE | 3 | ✅ | ClamAV .ndb + JSON format loaders |
| `internal/antivirus/matcher/patterns.go` | CREATE | 3 | ✅ | 25 built-in signatures (EICAR, ransomware, cryptominer, webshell, exploits) |
| `internal/antivirus/matcher/ahocorasick_test.go` | CREATE | 3 | ✅ | 30 tests — AC correctness, layer, builtins, loaders, concurrency |
| `internal/antivirus/heuristic/heuristic.go` | CREATE | 4 | ✅ | Layer orchestrator dispatching to 4 sub-analyzers |
| `internal/antivirus/heuristic/pe.go` | CREATE | 4 | ✅ | PE header analysis: packers, W+X sections, entry point, entropy |
| `internal/antivirus/heuristic/elf.go` | CREATE | 4 | ✅ | ELF analysis: executable stack, stripped binaries, UPX, entropy |
| `internal/antivirus/heuristic/script.go` | CREATE | 4 | ✅ | Script obfuscation: JS, PowerShell, VBS, Bash (weighted scoring) |
| `internal/antivirus/heuristic/shellcode.go` | CREATE | 4 | ✅ | NOP sleds, syscall patterns, XOR decode loops |
| `internal/antivirus/heuristic/heuristic_test.go` | CREATE | 4 | ✅ | 25 tests — PE, ELF, script, shellcode, entropy, concurrency |
| `internal/antivirus/entropy/entropy.go` | CREATE | 5 | ⬜ | Shannon entropy calculator |
| `internal/antivirus/entropy/packer.go` | CREATE | 5 | ⬜ | Packer detection |
| `internal/antivirus/entropy/entropy_test.go` | CREATE | 5 | ⬜ | Entropy tests |
| `internal/antivirus/yara/engine.go` | CREATE | 6 | ⬜ | YARA rule engine |
| `internal/antivirus/yara/rules.go` | CREATE | 6 | ⬜ | Built-in YARA rules |
| `internal/antivirus/yara/engine_test.go` | CREATE | 6 | ⬜ | YARA tests |
| `internal/antivirus/sigdb/database.go` | CREATE | 7 | ⬜ | Unified signature DB |
| `internal/antivirus/sigdb/updater.go` | CREATE | 7 | ⬜ | Auto-updater |
| `internal/antivirus/sigdb/format.go` | CREATE | 7 | ⬜ | Signature file parsers |
| `internal/antivirus/sigdb/builtin.go` | CREATE | 7 | ⬜ | Compiled-in essential sigs |
| `internal/antivirus/cache/cache.go` | CREATE | 8 | ⬜ | LRU scan cache |
| `internal/antivirus/cache/cache_test.go` | CREATE | 8 | ⬜ | Cache tests |
| `internal/antivirus/handler.go` | CREATE | 10 | ⬜ | HTTP API endpoints |
| `internal/storage/storage.go` | MODIFY | 9 | ⬜ | Wire antivirus into storage |
| `internal/storage/admin/admin.go` | MODIFY | 9 | ⬜ | PutObject scan hook + routes |
| `internal/storage/events/events.go` | MODIFY | 9 | ⬜ | New scan event constants |
| `.env.example` | MODIFY | 1 | ✅ | Added 17 ANTIVIRUS_* env vars |

---

## Implementation Progress Log

### ✅ Phase 1 — Foundation (Completed: 2026-05-13)

**Files created:** `types.go` (362 lines), `config.go` (233 lines), `engine.go` (492 lines), `engine_test.go` (350 lines)  
**Tests:** 25/25 passing | `go vet`: clean  
**Key decisions:**
- `ScanLayer` interface mirrors `scanner.Scanner` but purpose-built for AV detection
- Layer list frozen after `Start()` — prevents data races, panics on late registration
- Atomic stats counters — lock-free concurrent updates, zero overhead
- Layer errors are non-fatal — a broken YARA rule won't block uploads
- Confidence-based verdicts: ≥0.8 → malware, <0.8 → suspicious
- Config auto-corrects invalid values and returns warnings instead of crashing

### ✅ Phase 2 — Hash Database (Completed: 2026-05-13)

**Files created:** `hashdb/hashdb.go` (306 lines), `hashdb/loader.go` (296 lines), `hashdb/hashdb_test.go` (410 lines)  
**Tests:** 27/27 passing | `go vet`: clean  
**Key metrics:**
- Bloom filter: **1.14 MB** for 500K hashes at 0.01% false positive rate (k=14 hash functions)
- Actual measured FP rate: **1.007%** (at 1% target, 10K items, 100K test lookups)
- Kirsch-Mitzenmacker optimisation: g_i(x) = h1(x) + i·h2(x) — only 1 SHA-256 hash needed
- Three format loaders: ClamAV `.hdb`/`.hsb`, AxiomNizam JSON, plain text
- `Reload()` builds new bloom filter outside lock, swaps atomically — zero-downtime updates
- Category inference from ClamAV naming conventions (e.g. `Trojan.Win32.Emotet.A` → trojan)
- Thread-safe: RWMutex for all operations, concurrent read test passes with 50 goroutines

### ✅ Phase 3 — Byte-Pattern Matcher / Aho-Corasick (Completed: 2026-05-13)

**Files created:** `matcher/ahocorasick.go` (347 lines), `matcher/signature.go` (289 lines), `matcher/patterns.go` (305 lines), `matcher/ahocorasick_test.go` (390 lines)  
**Tests:** 30/30 passing | `go vet`: clean  
**Key design decisions:**
- Pure Go Aho-Corasick — zero dependencies, zero cgo, fully portable
- Classical trie with BFS failure links + output chains (not DFA — saves memory for sparse binary patterns)
- Nodes use `map[byte]int` children — trades speed for lower memory when alphabet utilisation is sparse
- Automaton is immutable after `Build()` — no locking needed during concurrent scans
- Layer deduplicates: same signature matching multiple offsets produces only 1 ThreatInfo
- `Reload()` swaps atomically via RWMutex for hot signature updates
- 25 built-in patterns compiled into binary: EICAR, WannaCry, Log4Shell, ShellShock, Stratum/XMRig/CoinHive miners, PHP/JSP webshells, bash/Python reverse shells, PE/ELF embedded executable detection, Meterpreter/Cobalt Strike markers
- ClamAV `.ndb` loader skips wildcard patterns (`??`, `{n}`, `(a|b)`) — exact hex only for now
- Minimum 4-byte pattern length enforced to prevent false positives

### ✅ Phase 4 — Behavioral Heuristics (Completed: 2026-05-13)

**Files created:** `heuristic/heuristic.go` (120 lines), `heuristic/pe.go` (225 lines), `heuristic/elf.go` (228 lines), `heuristic/script.go` (256 lines), `heuristic/shellcode.go` (148 lines), `heuristic/heuristic_test.go` (350 lines)  
**Tests:** 25/25 passing | `go vet`: clean  
**Key design decisions:**
- **PE analysis**: Raw binary parsing of MZ/PE/COFF/Optional/Section headers (no external deps). Detects 16 packer section names (UPX, Themida, VMProtect, ASPack, etc.), writable+executable sections, entry point anomalies, per-section Shannon entropy >7.2, and virtual-vs-raw size mismatches >10x
- **ELF analysis**: Full ELF32/ELF64 little-endian header parsing. Detects executable GNU_STACK (exploit indicator), stripped+exec-stack combo (evasion), UPX on Linux, high-entropy LOAD segments
- **Script obfuscation**: Weighted scoring system — each indicator carries weight + confidence; triggers require ≥3 score points AND ≥2 distinct indicators. Covers JavaScript (eval, fromCharCode, unescape, Function constructor), PowerShell (-EncodedCommand, IEX, WebClient, -w hidden), VBScript (Execute, Chr()), Bash (eval+base64, /dev/tcp)
- **Shellcode detection**: Skips PE/ELF files (have dedicated analyzers). Detects variable-length NOP sleds (≥16 bytes), x86/x64 syscall instructions, shellcode prologues, XOR decode loops, Windows API hash rotation (Metasploit). Requires ≥2 pattern matches to avoid false positives
- **Zero RAM overhead**: All analyzers are pure functions — no data structures, no state, fully concurrent-safe
