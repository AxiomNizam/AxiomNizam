package eventbus

// =====================================================
// P2.1 — Typed-topic facade for the canonical event bus.
//
// Rationale: the workspace today has three overlapping event APIs:
//
//   1. `internal/events`        — in-process `Bus.Publish(*Event)` with
//                                 an `EventType` enum.
//   2. `internal/eventbus`      — this package, a broker-style manager
//                                 with topics / subscriptions / DLQ.
//   3. `internal/apiserver`     — `ResourceStore.notifyWatchers` which
//                                 calls registered `ResourceWatcher`s.
//
// Unifying all three in one pass would churn dozens of call sites.
// Instead we pick `eventbus` as the canonical broker and add a small
// generic typed-topic façade so new code can publish/subscribe with
// compile-time types.  Legacy APIs continue to work unchanged; the
// follow-up is to migrate callers one package at a time.
//
// The façade is intentionally thin: it marshals the typed payload into
// an `EventBusEvent.Data` map using the canonical JSON shape and
// delegates to whatever `EventBusManager` implementation is in play.
// =====================================================

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Topic declares a strongly-typed event topic.  The zero value is not
// usable — always construct with `NewTopic[T]`.
type Topic[T any] struct {
	name    string
	version string
}

// NewTopic returns a typed topic descriptor.  `name` must match the
// logical topic known to the underlying manager; `version` is written
// into `EventBusEvent.Version` so subscribers can evolve payloads
// safely.
func NewTopic[T any](name, version string) Topic[T] {
	return Topic[T]{name: name, version: version}
}

// Name returns the canonical topic name.
func (t Topic[T]) Name() string { return t.name }

// Version returns the declared schema version.
func (t Topic[T]) Version() string { return t.version }

// Publish serialises `payload` and hands it to the given manager on
// the configured topic.  Returns the manager's publish response so the
// caller can observe the assigned event ID / sequence.
func Publish[T any](mgr EventBusManager, topic Topic[T], tenantID string, payload T) (*EventPublishResponse, error) {
	if mgr == nil {
		return nil, fmt.Errorf("eventbus.Publish: manager is nil")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("eventbus.Publish: marshal payload: %w", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		// Payload is not a JSON object — wrap it in a single-key envelope.
		data = map[string]interface{}{"value": payload}
	}
	evt := &EventBusEvent{
		ID:              uuid.NewString(),
		TenantID:        tenantID,
		Type:            topic.name,
		Version:         topic.version,
		Source:          "axiomnizam",
		Timestamp:       time.Now().UTC(),
		DataContentType: "application/json",
		Data:            data,
	}
	return mgr.PublishEvent(evt)
}

// TypedHandler is the signature a typed subscriber provides.
type TypedHandler[T any] func(payload T, raw *EventBusEvent) error

// typedRegistry routes a single EventBusEvent to all typed handlers
// registered for its Type/Version pair.  It exists so the in-memory
// façade can deliver events synchronously without the subscription
// model forcing callers to write untyped decoders.
type typedRegistry struct {
	mu       sync.RWMutex
	handlers map[string][]func(*EventBusEvent) error
}

// Globals are intentional: the façade is a compatibility bridge, not a
// long-lived managed component.  Callers that want full DI should use
// `PublishEvent` / manager subscriptions directly.
var (
	globalRegistryOnce sync.Once
	globalRegistry     *typedRegistry
)

func registry() *typedRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = &typedRegistry{
			handlers: make(map[string][]func(*EventBusEvent) error),
		}
	})
	return globalRegistry
}

// Subscribe registers a typed handler for `topic`.  The façade is
// deliver-synchronously: publishers that want guaranteed async delivery
// should use `EventBusManager.CreateSubscription` directly.
func Subscribe[T any](topic Topic[T], h TypedHandler[T]) {
	key := topic.name + "@" + topic.version
	wrapped := func(e *EventBusEvent) error {
		var payload T
		buf, err := json.Marshal(e.Data)
		if err != nil {
			return fmt.Errorf("eventbus.Subscribe[%s]: marshal: %w", topic.name, err)
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return fmt.Errorf("eventbus.Subscribe[%s]: unmarshal: %w", topic.name, err)
		}
		return h(payload, e)
	}
	r := registry()
	r.mu.Lock()
	r.handlers[key] = append(r.handlers[key], wrapped)
	r.mu.Unlock()
}

// Dispatch is the hook the concrete manager calls after persisting an
// event.  Current in-memory implementation can call it directly; for
// remote brokers it would be invoked by the subscriber goroutine.
func Dispatch(e *EventBusEvent) {
	if e == nil {
		return
	}
	r := registry()
	r.mu.RLock()
	hs := append([]func(*EventBusEvent) error(nil), r.handlers[e.Type+"@"+e.Version]...)
	r.mu.RUnlock()
	for _, h := range hs {
		_ = h(e)
	}
}
