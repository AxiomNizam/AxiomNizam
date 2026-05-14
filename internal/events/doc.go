// Package events is the legacy domain-event and audit-event layer.
//
// Phase 1 consolidation status
// ----------------------------
//
// The Phase 1 checklist calls for collapsing `events/` and `eventbus/`
// into a single package with typed topics.  The typed-topic facade now
// lives in `internal/eventbus/typed.go` (see `eventbus.Topic[T]`,
// `Publish[T]`, `Subscribe[T]`).
//
// This package is kept for now because it carries significantly more
// than event transport:
//
//   - `Event` / `EventBus` / `EventRecorder` domain event model.
//   - `ReconciliationEvent` audit log.
//   - `EventLifecycleManager` hooks wired into `controllers/`.
//
// New code SHOULD depend on `eventbus` instead.  Code in this package
// is frozen except for bug fixes; any new event type should be declared
// as an `eventbus.Topic[T]`.  A follow-up pass will migrate the
// recorder/audit surface onto `eventbus` and delete this package.
package events
