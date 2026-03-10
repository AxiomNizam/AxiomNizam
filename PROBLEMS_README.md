# AxiomNizam — Problems Found & Resolved

This document catalogs all problems (compile errors, lint warnings, and code quality issues) that were identified and fixed across the project. All **146+ issues** have been resolved.

---

## Summary

| Category | Count | Status |
|---|---|---|
| Undefined types / interface mismatches | ~52 | **Fixed** |
| Duplicate string literals (3+ usages) | ~45 | **Fixed** |
| `fmt.Println` redundant newline warnings | ~30 | **Fixed** |
| Cognitive complexity exceeds threshold | 4 | **Fixed** |
| Redundant type assertions | 2 | **Fixed** |
| Missing dependency (graphql) | 1 | **Fixed** |
| Dockerfile build issues | 1 | **Fixed** |
| Miscellaneous compile errors | ~12 | **Fixed** |

**Total: ~147 issues — all resolved. `get_errors` now returns 0.**

---

## Detailed Fix Log

### 1. `internal/eventbus/in_memory.go` — Complete Rewrite

**Problems (23 errors):** All type references were wrong — the implementation used `Event`, `Topic`, `Subscription`, `DeadLetterEvent` which don't exist in `models.go`.

**Fixes:**
- `Event` → `EventBusEvent`
- `Topic` → `EventTopic`
- `Subscription` → `EventSubscription`
- `DeadLetterEvent` → `DLQEvent`
- `PublishEvent` return type → `(*EventPublishResponse, error)`
- `ListEvents` signature → `(tenantID, eventType, processed string)`
- `ListSubscriptions` signature → `(tenantID string)`
- `ListDLQ` → `ListDLQEvents(tenantID string)`
- `event.CreatedAt` → `event.Timestamp`
- `event.Topic` → `event.Type`
- `topic.EventCount` → `topic.MessageCount`

---

### 2. `internal/tracing/in_memory.go` — Complete Rewrite

**Problems (21 errors):** Types and field names mismatched `models.go` definitions.

**Fixes:**
- `ServiceMetrics` → `TraceMetrics`
- `ServiceDependency` → `DependencyMetrics`
- `OperationMetrics` → `SpanMetrics`
- `trace.Service` → helper `traceService()` that returns `Services[0]` (Trace has `Services []string`, not `Service string`)
- `span.Operation` → `span.OperationName`
- `span.Duration` used as `time.Duration` → fixed to `int64` arithmetic
- `span.Error` is `bool` not `string` → use `span.ErrorMessage` for error text
- `ErrorAnalysis.Error` → `ErrorAnalysis.ErrorType`
- `ErrorAnalysis.Count` type `int` → `int64`
- Removed `TraceSearchRequest.Status` filter (field doesn't exist)

---

### 3. `internal/export/in_memory.go` — Complete Rewrite

**Problems (8 errors):** `ExportResult` type doesn't exist; field types wrong.

**Fixes:**
- `results map[string][]*ExportResult` → `history map[string][]*ExportHistory`
- `GetResults` → `GetHistory`
- `UpdateProgress(id string, progress int, status string)` → `UpdateProgress(id string, progress float64, status ExportStatus)`
- `export.Status = "submitted"` → `ExportPending`
- `"cancelled"` → `ExportCancelled`
- `"completed"` → `ExportCompleted`
- Removed `export.UpdatedAt = time.Now()` (ExportJob has no `UpdatedAt` field)
- Added `const errExportNotFound` for 4 duplicate literal usages

---

### 4. `internal/rbac/in_memory.go`

**Problems:** Undefined types `RBACEvent`, `AccessDecision`, `PolicyViolation` + duplicate literal.

**Fixes:**
- `RBACEvent` → `AuditEvent`
- `AccessDecision` → `AccessReview`
- `PolicyViolation` → `PolicyEvaluation`
- Field fixes: `Decision` → `Status`, `Resource` → correct field
- Added `const errRequestNotFound = "request not found"` (3 usages)

---

### 5. `internal/streaming/in_memory.go`

**Problems:** Undefined types and wrong field names.

**Fixes:**
- `StreamEvent` → `Message`
- `StreamConsumer` → `Consumer`
- Various field name corrections to match `models.go`

---

### 6. `internal/streaming/handlers.go`

**Problems:** Interface method signatures didn't match implementation.

**Fixes:** Aligned handler interface definitions with actual model types.

---

### 7. `internal/tenant/in_memory.go` — Deleted

**Problem:** File had extensive errors with undefined types and was a duplicate/conflicting implementation of `manager.go`.

**Fix:** Deleted the file entirely. `manager.go` serves as the correct implementation.

---

### 8. `internal/tenant/manager.go`

**Problems:** Duplicate string literals.

**Fixes:**
- Added `const errTenantNotFound = "tenant not found: %s"` (4 usages)
- Added `const errQuotaNotFound = "quota not found: %s"` (3 usages)

---

### 9. `internal/tenant/handlers.go`

**Problems:** Referenced types from deleted `in_memory.go`.

**Fixes:** Updated to use types defined in `models.go`.

---

### 10. `internal/apiserver/server.go`

**Problems:** 3 duplicate string literals + cognitive complexity of `ApplyResource` (16 > 15).

**Fixes:**
- Added `const errResourceNotFound = "resource not found: %s/%s"` (6 usages)
- Added `const errUnknownKind = "unknown resource kind: %s"` (3 usages)
- Added `const routeNsKindName` for route pattern (3 usages)
- Extracted `parseResourceByKind()` helper to reduce `ApplyResource` complexity

---

### 11. `internal/bulk/in_memory.go`

**Problems:** Duplicate string literal.

**Fix:** Added `const errOperationNotFound = "operation not found"` (3 usages)

---

### 12. `internal/integration/integration_test.go`

**Problems:** Duplicate string literals.

**Fixes:**
- Added `const testOwnerUser = "test-user"` (4 usages)
- Added `const testOwnerInteg = "integration-test"` (4 usages)

---

### 13. `internal/utils/retry/retry.go`

**Problems:** Duplicate string literal.

**Fix:** Added `const errMaxAttemptsExceeded = "max attempts (%d) exceeded"` (3 usages)

---

### 14. `internal/utils/uuid/uuid.go`

**Problems:** Compile error — wrong import/usage of UUID library.

**Fix:** Corrected to use `google/uuid` v1.6.0 API properly.

---

### 15. `internal/utils/metrics/metrics.go`

**Problems:** Compile errors with metric types.

**Fix:** Corrected type usage to match actual package API.

---

### 16. `internal/policies/access/access.go`

**Problems:** Compile errors with undefined types.

**Fix:** Corrected to match defined types in the package.

---

### 17. `internal/migrations/migrations.go`

**Problems:** Cognitive complexity of `createIndexes` (16 > 15 allowed).

**Fix:** Refactored 16 sequential `if err != nil` blocks into a data-driven loop using `[]indexDef` slice, reducing complexity to 2.

---

### 18. `cmd/axiomnizam-server/main.go`

**Problems:** `fmt.Println` with redundant trailing newline.

**Fix:** Removed trailing `\n` from Println argument.

---

### 19. `cmd/axiomnizamctl/lifecycle_test.go`

**Problems:** Duplicate string literal + cognitive complexity 64 (max 15).

**Fixes:**
- Added `const testAPIName = "test-api"` (3 usages)
- Extracted 5 helper functions: `verifyResourceExists`, `verifyPendingStatus`, `waitForReady`, `verifyFinalStatus`, `verifyListQueryable` — reducing `TestReconciliationLoopComplete` from complexity 64 to under 15

---

### 20. `cmd/axiomnizamctl/apibank.go`

**Problems:** 2 `fmt.Println` with redundant newlines.

**Fix:** Split `fmt.Println("...\n")` into `fmt.Println("...")` + `fmt.Println()`.

---

### 21. `cmd/axiomnizamctl/integration.go`

**Problems:** 2 redundant type assertions + 2 redundant newlines + duplicate `"%s: %v\n"` (5 usages).

**Fixes:**
- Removed `banks.(interface{})` type assertions (asserting to same type)
- Fixed `fmt.Println` redundant newlines
- Added `const fmtKeyValue = "%s: %v\n"` (5 usages)

---

### 22. `cmd/axiomnizamctl/mesh.go`

**Problems:** 1 redundant newline + duplicate `"Domain name"` (6 usages) + duplicate `"Product name"` (4 usages).

**Fixes:**
- Fixed `fmt.Println` redundant newline
- Added `const flagDomainName = "Domain name"` (6 usages)
- Added `const flagProductName = "Product name"` (4 usages)

---

### 23. `internal/integration/demo.go`

**Problems:** 1 redundant newline + duplicate `"Finance/TransactionData"` (4 usages) + 8 more `fmt.Println` redundant newlines.

**Fixes:**
- Added `const resFinanceTransaction = "Finance/TransactionData"` (4 usages)
- Fixed all 9 `fmt.Println` redundant newline warnings (split to separate calls)

---

### 24. `internal/integration/phase1_examples.go`

**Problems:** 5 `fmt.Println` with raw string backtick arguments ending in `\n` + 4 `fmt.Println("\n")`.

**Fixes:**
- Changed `fmt.Println(` raw strings `)` → `fmt.Print(` to avoid double newlines
- Changed `fmt.Println("\n")` → `fmt.Println()` for single blank line

---

### 25. `internal/integration/setup_guide.go`

**Problems:** 2 `fmt.Println` with raw string backtick arguments ending in `\n`.

**Fix:** Changed `fmt.Println(` → `fmt.Print(` for raw strings that already end with newline.

---

### 26. `internal/graphql/` — Dependency

**Problem:** Missing `graphql-go/graphql` dependency.

**Fix:** Added `github.com/graphql-go/graphql v0.8.1` to `go.mod`.

---

### 27. `Dockerfile`

**Problem:** Build stage referenced incorrect binary path.

**Fix:** Corrected binary output path in the `go build` command.

---

## Verification

After all fixes, running error diagnostics returns:

```
No errors found.
```

All 146+ problems have been resolved across 25+ files.
