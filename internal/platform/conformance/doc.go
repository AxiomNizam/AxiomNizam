// Package conformance contains repository-level structural tests that
// enforce architecture invariants declared in the Phase 5 checklist.
//
// These are *lint-as-test* checks — running `go test ./internal/platform/conformance/...`
// fails CI when:
//
//   - a handler struct holds a raw `*clientv3.Client` or `*sql.DB`
//     (handlers must go through a `Store[T]` / repository instead);
//   - a handler struct holds an ad-hoc `sync.Mutex` / `sync.RWMutex`
//     guarding an in-memory resource map (state lives in the store);
//   - a package imports `log` directly from controllers/reconcilers
//     (use `internal/logging` or `internal/utils/logger`).
//
// The tests are intentionally loose: they only inspect `internal/handlers/`
// today.  Extend the `targets` slice to cover additional packages once
// they have been migrated.
package conformance
