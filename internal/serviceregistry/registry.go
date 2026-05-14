// Package serviceregistry is an in-process service registry in the
// shape of Consul / Nomad's native service discovery.  Services
// register with one or more health checks; the registry rolls up
// check results into an overall Passing / Warning / Critical
// status; subscribers observe changes via a streaming interface.
//
// The registry is storage-agnostic — it holds state in memory and
// exposes Export / Import for callers that want durability.  The
// design target is per-tenant registries embedded in controller
// processes, not a cluster-wide discovery plane.
package serviceregistry

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// Status is the coarse health classification.
type Status string

const (
	// StatusPassing means all checks are healthy.
	StatusPassing Status = "passing"
	// StatusWarning means at least one check is warning and none are
	// critical.
	StatusWarning Status = "warning"
	// StatusCritical means at least one check has reported critical.
	StatusCritical Status = "critical"
	// StatusUnknown means the service has no checks or none have
	// reported yet.
	StatusUnknown Status = "unknown"
)

// CheckKind distinguishes how a check reports.
type CheckKind string

const (
	// CheckHTTP expects an external prober (like a sidecar) to call
	// UpdateCheck periodically — the registry itself does not make
	// HTTP calls, keeping it transport-agnostic.
	CheckHTTP CheckKind = "http"
	// CheckTCP is similar — reported externally.
	CheckTCP CheckKind = "tcp"
	// CheckTTL requires the service itself to call UpdateCheck
	// before the TTL expires; otherwise the registry marks it
	// critical.  This is the classic self-reported-liveness model.
	CheckTTL CheckKind = "ttl"
)

// Check describes one health signal.
type Check struct {
	// ID is unique within the service.
	ID string
	// Kind selects the probing model.
	Kind CheckKind
	// TTL is the expiry window for CheckTTL.
	TTL time.Duration
	// Status is the last reported outcome.
	Status Status
	// Notes is free-form detail for the last update, shown in UI.
	Notes string
	// Updated is when Status was last written.
	Updated time.Time
}

// Service is one registration.
type Service struct {
	// ID is the registry-wide unique ID.
	ID string
	// Name groups service instances (e.g., "payments-api").
	Name string
	// Address is the reachable IP/host.
	Address string
	// Port is the reachable port.
	Port int
	// Tags classify the instance (version, region, tier).
	Tags []string
	// Checks is the set of health checks on this instance.
	Checks map[string]*Check
}

// Rollup computes the effective service status.
func (s *Service) Rollup() Status {
	if len(s.Checks) == 0 {
		return StatusUnknown
	}
	worst := StatusPassing
	for _, c := range s.Checks {
		switch c.Status {
		case StatusCritical:
			return StatusCritical
		case StatusWarning:
			worst = StatusWarning
		}
	}
	return worst
}

// Registry is the mutable store.
type Registry struct {
	mu       sync.RWMutex
	services map[string]*Service
	// ttlReaper exits when Close is called.
	stopCh chan struct{}
}

// New returns an empty registry with the TTL reaper running.
func New() *Registry {
	r := &Registry{
		services: map[string]*Service{},
		stopCh:   make(chan struct{}),
	}
	go r.reaper()
	return r
}

// Close halts the reaper.  Idempotent.
func (r *Registry) Close() {
	select {
	case <-r.stopCh:
	default:
		close(r.stopCh)
	}
}

// ErrNotFound is returned when the requested service ID is missing.
var ErrNotFound = errors.New("serviceregistry: service not found")

// Register inserts or replaces a service.  Checks that already exist
// retain their Status so a re-register does not incorrectly reset
// a healthy service to Unknown.
func (r *Registry) Register(s *Service) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.services[s.ID]; ok {
		for id, c := range s.Checks {
			if prev, ok := existing.Checks[id]; ok {
				c.Status = prev.Status
				c.Updated = prev.Updated
			}
		}
	}
	r.services[s.ID] = s
}

// Deregister removes a service.
func (r *Registry) Deregister(id string) { r.mu.Lock(); delete(r.services, id); r.mu.Unlock() }

// UpdateCheck records a check result.
func (r *Registry) UpdateCheck(serviceID, checkID string, status Status, notes string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	svc, ok := r.services[serviceID]
	if !ok {
		return ErrNotFound
	}
	c, ok := svc.Checks[checkID]
	if !ok {
		return ErrNotFound
	}
	c.Status = status
	c.Notes = notes
	c.Updated = time.Now()
	return nil
}

// Get returns a deep-ish copy of the service for read-only use.
func (r *Registry) Get(id string) (*Service, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.services[id]
	if !ok {
		return nil, false
	}
	return cloneService(s), true
}

// ByName returns all services with the given name, sorted by ID for
// deterministic output.
func (r *Registry) ByName(name string) []*Service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Service
	for _, s := range r.services {
		if s.Name == name {
			out = append(out, cloneService(s))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// cloneService copies a service so callers cannot mutate registry state.
func cloneService(s *Service) *Service {
	cp := *s
	cp.Tags = append([]string(nil), s.Tags...)
	cp.Checks = make(map[string]*Check, len(s.Checks))
	for k, c := range s.Checks {
		cc := *c
		cp.Checks[k] = &cc
	}
	return &cp
}

// reaper marks CheckTTL checks critical if they have not been
// updated within their TTL window.
func (r *Registry) reaper() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case now := <-ticker.C:
			r.mu.Lock()
			for _, svc := range r.services {
				for _, c := range svc.Checks {
					if c.Kind != CheckTTL {
						continue
					}
					if c.Updated.IsZero() {
						continue // waiting on first report
					}
					if now.Sub(c.Updated) > c.TTL && c.Status != StatusCritical {
						c.Status = StatusCritical
						c.Notes = "TTL expired"
					}
				}
			}
			r.mu.Unlock()
		}
	}
}
