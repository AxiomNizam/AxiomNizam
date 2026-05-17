# Gatekeeper Module Cleanup Plan

> **Generated:** 2026-05-17
> **Scope:** `internal/gatekeeper/` — 88 .go files, 22 packages

---

## Current State

| Category | Count |
|----------|-------|
| Total .go files | 88 |
| Broken (missing `package` declaration) | 27 |
| Stub/placeholder files | ~32 |
| Compilable packages | 16 |
| Broken packages | 6 (`events`, `middleware`, `sms`, `email`, `webauthn`, `testutil`) |

---

## Phase 1: Fix Build-Blocking Issues

**Goal:** Module compiles without errors.

### 1.1 Fix 27 files with missing `package` declaration

All these files have `package ` with no package name — Go syntax errors:

| Package | Files |
|---------|-------|
| `events/` | `verified.go`, `enrolled.go`, `failed.go`, `disabled.go`, `backup_code_used.go` |
| `middleware/` | `context.go`, `grpc.go`, `http.go` |
| `sms/` | `provider.go`, `service.go`, `errors.go` |
| `email/` | `provider.go`, `service.go`, `errors.go` |
| `webauthn/` | `service.go`, `authentication.go`, `registration.go`, `errors.go` |
| `testutil/` | `fake_clock.go`, `fixtures.go`, `mocks.go` |
| `bootstrap/` | `providers.go`, `routes.go` |

**Action:** Add correct `package <name>` declarations to each file. For files that are pure stubs, either add minimal compilable content or delete the file.

### 1.2 Fix Raft FSM constant mismatches

**File:** `raft/fsm.go`

The FSM `Apply()` switch references constants that don't exist in `models/enums.go`:

| FSM references | Actual constant in `enums.go` |
|---|---|
| `models.CmdEnrollFactor` | `models.RaftCmdEnrollFactor` |
| `models.CmdActivateFactor` | `models.RaftCmdActivateFactor` |
| `models.CmdDisableFactor` | `models.RaftCmdDisableFactor` |
| `models.CmdRevokeFactor` | (does not exist — add or remove) |
| `models.CmdBeginChallenge` | (does not exist — add or remove) |
| `models.CmdVerifyChallenge` | `models.RaftCmdVerifyChallenge` |
| `models.CmdExpireChallenge` | `models.RaftCmdExpireChallenge` |
| `models.CmdFailChallenge` | (does not exist — add or remove) |
| `models.CmdUseBackupCode` | (does not exist — add or remove) |
| `models.CmdTrustDevice` | `models.RaftCmdTrustDevice` |
| `models.CmdRevokeDevice` | `models.RaftCmdRevokeDevice` |

Also fix:
- `models.ConditionReady` → `models.ConditionTypeReady`
- `models.ChallengePhasePending` → `models.ChallengePhaseWaiting`

**Action:** Update FSM to use the actual constant names from `enums.go`. Add missing constants if they are needed.

### 1.3 Fix Bootstrap `challenge.NewService()` arg count

**File:** `bootstrap/module.go:86-90`

Current (broken):
```go
m.challengeSvc = challenge.NewService(m.challengeRepo, m.factorRepo, challenge.NewRealClock())
```

Required (4 args):
```go
m.challengeSvc = challenge.NewService(m.challengeRepo, m.factorRepo, m.totpSvc, challenge.NewRealClock())
```

**Action:** Pass `m.totpSvc` as the `TOTPValidator` argument.

### 1.4 Fix Bootstrap nil FactorService

**File:** `bootstrap/module.go:129`

`FactorService` is passed as `nil` to `NewHTTPHandler` — will nil-pointer panic on `ListFactors`, `GetFactor`, `DeleteFactor` endpoints.

**Action:** Create a proper `factorServiceWrapper` via `wrapFactorService(m.factorRepo)` and pass it.

---

## Phase 2: Fix Runtime-Critical Issues

**Goal:** Module runs without panics or data corruption.

### 2.1 Fix `BackupCode.MarshalJSON()` returning `nil, nil`

**File:** `models/backup_code.go:27-29`

Returns `nil, nil` which corrupts JSON responses.

**Action:** Implement proper JSON marshaling or remove the custom method (let default encoding handle it).

### 2.2 Fix `StartControllers()` no-op

**File:** `system.go:342-347`

Logs "started" but never actually starts any controller loop. The `controller/manager.go` has proper `Start()`/`Stop()` methods but is never used.

**Action:** Wire `controller.Manager` in `StartControllers()` to actually run the reconcile loop.

### 2.3 Fix `ConsumeBackupCode` stub

**File:** `backupcodes/service.go:78-92`

Always returns `errors.New("backup code not found or already used")`. The hash lookup is unimplemented.

**Action:** Implement actual hash-based lookup using `backupCodeRepo`.

### 2.4 Fix `VerifyDeviceToken` adapter stub

**File:** `adapters.go:215`

Always returns `false, nil` — trusted device bypass never works.

**Action:** Implement actual token verification by delegating to the service with fingerprint lookup.

### 2.5 Fix `ListTrustedDevices` handler stub

**File:** `handlers/http.go:310-313`

Returns hardcoded empty array, never calls device service.

**Action:** Implement by calling `deviceSvc.ListTrustedDevices()` or querying the repo directly.

### 2.6 Fix `loadFactorsFromKV()` and `loadPoliciesFromKV()` dead code

**File:** `system.go:202-239`

Iterates keys but does nothing with them (`_ = key`).

**Action:** Implement actual deserialization and loading, or remove these methods.

---

## Phase 3: Security Fixes

**Goal:** Eliminate known security vulnerabilities.

### 3.1 Replace hardcoded encryption keys

**Files:**
- `system.go:287` — `[]byte("encryption-key-placeholder")`
- `bootstrap/module.go:97` — `[]byte("encryption-key-todo")`

**Action:** Load encryption key from environment variable or config. Fail startup if not set.

### 3.2 Encrypt TOTP secrets at rest

**File:** `enrollment/service.go:85,140`

Currently stores `EncryptedSecret: []byte(secret)` in plaintext with `// TODO: Encrypt this`.

**Action:** Use AES-GCM encryption with the configured key before persisting, decrypt on read.

### 3.3 Fix backup code hashing

**File:** `enrollment/service.go:215-217`

`hashCode()` is just `base64.StdEncoding.EncodeToString([]byte(code))` — trivially reversible.

**Action:** Use `bcrypt` or `argon2id` for hashing backup codes.

### 3.4 Fix device token hashing

**File:** `trusteddevices/service.go:179-182`

`hashDeviceToken()` is also just base64 encoding.

**Action:** Use `bcrypt` or `sha256` with salt.

### 3.5 Fix non-constant-time OTP comparison

**File:** `totp/validator.go:37`

Uses `code == expectedCode` — enables timing attacks.

**Action:** Use `crypto/subtle.ConstantTimeCompare()`.

---

## Phase 4: Remove Dead Code

**Goal:** Reduce noise, remove unused files and duplicate definitions.

### 4.1 Delete empty stub files that add no value

Files that contain only `"Placeholder - implementation pending."` and a package declaration:

- `contracts/types.go` — types already in `contracts/service.go`
- `contracts/provider.go` — `Provider` interface already in `contracts/service.go`
- `config/defaults.go` — `DefaultConfig()` already in `config/config.go`
- `config/validation.go` — `Validate()` already in `config/config.go`
- `challenge/begin.go`, `verify.go`, `session.go`, `state.go` — logic already in `challenge/service.go`
- `enrollment/setup.go`, `activate.go`, `disable.go`, `status.go` — logic already in `enrollment/service.go`
- `handlers/grpc.go`, `dto.go`, `mapper.go` — no gRPC handler implemented
- `metrics/histograms.go`, `labels.go` — metrics already in `metrics/counters.go`
- `policy/evaluator.go`, `rules.go` — logic already in `policy/engine.go`
- `risk/scorer.go`, `signals.go` — logic already in `risk/engine.go`
- `backupcodes/generator.go`, `hasher.go`, `validator.go` — logic already in `backupcodes/service.go`
- `totp/errors.go`, `qrcode.go` — stubs
- `trusteddevices/cookie.go`, `fingerprint.go`, `token.go` — stubs
- `raft/store.go` — empty placeholder
- `audit/event_types.go` — placeholder comment only
- `cache/rate_limit.go` — placeholder

**Action:** Delete all listed files. Consolidate any actual logic into the parent service file.

### 4.2 Delete broken packages that are never imported

These packages have zero compilable code and are never imported anywhere:

| Package | Files | Status |
|---------|-------|--------|
| `events/` | 5 files | All broken, never imported |
| `middleware/` | 3 files | All broken, never imported |
| `sms/` | 3 files | All broken, never imported |
| `email/` | 3 files | All broken, never imported |
| `webauthn/` | 4 files | All broken, never imported |
| `testutil/` | 3 files | All broken, never imported |

**Action:** Delete entire directories. Re-create with proper implementations when actually needed.

### 4.3 Remove duplicate `Clock` interface

Defined identically in 3 places:
- `challenge/service.go:29-31`
- `totp/clock.go:6-8`
- `trusteddevices/service.go:23-25`

**Action:** Keep `totp/clock.go` as the canonical `Clock` interface. Have `challenge` and `trusteddevices` import it from there.

### 4.4 Remove duplicate repository interfaces

`contracts/repository.go` and `repositories/*.go` define overlapping interfaces with different signatures. The `pgstore/` implementations target `repositories/`. The `contracts/` versions are only used by adapter wrappers.

**Action:** Keep `repositories/` as the canonical interface layer. Remove `contracts/repository.go`. Update `adapters.go` to use `repositories/` interfaces directly.

### 4.5 Consolidate `system.go` and `bootstrap/module.go`

These two files wire up the same components in ~90% identical code.

**Action:** Keep `system.go` as the primary wiring (it has KVStore integration). Delete `bootstrap/module.go` or make it a thin wrapper that delegates to `system.go`.

---

## Phase 5: Architectural Improvements

**Goal:** Clean structure, proper patterns.

### 5.1 Implement `controller.Manager` reconcile loop

**File:** `controller/manager.go`

The `Manager` type exists with `Start()`/`Stop()` but is never instantiated.

**Action:** Wire `Manager` in `system.go` with proper goroutine-based reconcile loop, similar to how `internal/storage/controller/` works.

### 5.2 Implement Raft store

**File:** `raft/store.go` (currently empty)

The `contracts.RaftStore` interface and `raft/fsm.go` exist but there's no store implementation.

**Action:** Implement `raft.Store` that wraps the FSM, manages Raft log replication, and provides the `Apply`/`IsLeader`/`ReadFactor`/`ReadChallenge` API.

### 5.3 Wire audit logger to PostgreSQL

**File:** `audit/logger.go`

Currently in-memory only with `// TODO: PostgreSQL` comment.

**Action:** Add PG-backed audit log persistence, similar to `internal/storage/events/events.go`.

### 5.4 Implement policy rules

**File:** `policy/engine.go:120-125`

`GetPolicy()` returns a hardcoded default policy. `rules` is always `[]policy.Rule{}`.

**Action:** Define actual policy rules and load from config/KVStore.

---

## Execution Order

```
Phase 1 (Build)     ──►  Phase 2 (Runtime)  ──►  Phase 3 (Security)
                                                        │
                                                        ▼
                                                  Phase 4 (Dead Code)
                                                        │
                                                        ▼
                                                  Phase 5 (Architecture)
```

Each phase should be a separate PR/branch to keep reviews manageable.

---

## File Count After Cleanup

| Category | Before | After (est.) |
|----------|--------|--------------|
| Total .go files | 88 | ~45 |
| Broken files | 27 | 0 |
| Stub files | ~32 | 0 |
| Packages | 22 | ~14 |

---

*Last updated: 2026-05-17*
