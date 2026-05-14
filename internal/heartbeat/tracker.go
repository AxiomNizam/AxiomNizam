// Package heartbeat implements the Nomad-style TTL tracker for
// liveness signals.  Callers register an entity with a TTL; every
// subsequent Beat extends the deadline; if the deadline passes
// without a Beat, the configured Expire callback fires.
//
// The canonical use is "node liveness": every node calls Beat on a
// schedule; the server fires Expire when a node is silent for longer
// than its TTL, allowing the scheduler to reschedule its allocations.
// AxiomNizam also uses this for session lifetimes and client-agent
// liveness.
package heartbeat

import (
	"sync"
	"time"
)

// ExpireFunc runs when an entity's TTL lapses.  Called from the
// reaper goroutine — must not block; fan out to another goroutine
// if expensive work is required.
type ExpireFunc func(id string)

// Tracker is the TTL registry.
type Tracker struct {
	// OnExpire is invoked for each entity whose deadline has passed.
	// Can be nil if callers poll Expired() instead.
	OnExpire ExpireFunc
	// ReapInterval is how often the reaper goroutine runs.  Defaults
	// to 1 second when zero.
	ReapInterval time.Duration

	mu        sync.Mutex
	deadlines map[string]time.Time
	expired   []string
	stopCh    chan struct{}
	started   bool
}

// New returns a Tracker ready for use.  Callers must call Start to
// begin the reaper goroutine; Stop releases it.
func New(onExpire ExpireFunc) *Tracker {
	return &Tracker{
		OnExpire:  onExpire,
		deadlines: map[string]time.Time{},
		stopCh:    make(chan struct{}),
	}
}

// Start begins the reaper.  Calling Start twice panics because
// double-start would leak the second goroutine.
func (t *Tracker) Start() {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		panic("heartbeat: Start called twice")
	}
	t.started = true
	interval := t.ReapInterval
	if interval <= 0 {
		interval = time.Second
	}
	t.mu.Unlock()
	go t.loop(interval)
}

// Stop halts the reaper.  Safe to call even if Start was never called
// (makes Start no-op thereafter).
func (t *Tracker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.started {
		t.started = true // prevent future Start
		return
	}
	select {
	case <-t.stopCh:
		// already closed
	default:
		close(t.stopCh)
	}
}

// Set records or replaces an entity's deadline to now + ttl.  Used
// for both initial registration and subsequent heartbeats — the
// caller does not need to distinguish.
func (t *Tracker) Set(id string, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.deadlines[id] = time.Now().Add(ttl)
}

// Beat is an alias for Set — more readable at call sites that want
// to emphasise "this is a renewal, not a registration".
func (t *Tracker) Beat(id string, ttl time.Duration) { t.Set(id, ttl) }

// Delete removes an entity.  Used when the caller explicitly wants
// to forget a node (for example, graceful deregistration).
func (t *Tracker) Delete(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.deadlines, id)
}

// IsAlive reports whether id has a non-expired entry.
func (t *Tracker) IsAlive(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	dl, ok := t.deadlines[id]
	return ok && time.Now().Before(dl)
}

// Expired returns and clears the list of entities whose TTL has
// passed since the last Expired call.  Used by callers that prefer
// pulling to a callback.
func (t *Tracker) Expired() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := t.expired
	t.expired = nil
	return out
}

// loop is the reaper body.
func (t *Tracker) loop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.reap()
		}
	}
}

// reap scans deadlines for expiries and fires callbacks outside the lock.
func (t *Tracker) reap() {
	now := time.Now()
	t.mu.Lock()
	var expired []string
	for id, dl := range t.deadlines {
		if now.After(dl) {
			expired = append(expired, id)
			delete(t.deadlines, id)
		}
	}
	t.expired = append(t.expired, expired...)
	cb := t.OnExpire
	t.mu.Unlock()

	if cb == nil {
		return
	}
	for _, id := range expired {
		cb(id)
	}
}
