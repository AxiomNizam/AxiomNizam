// Package health provides a Kubernetes-style health/readiness/liveness
// framework for AxiomNizam.  The control plane and data-plane
// subsystems register named probes with a Registry; HTTP handlers then
// expose /healthz, /livez and /readyz endpoints that aggregate probe
// results.  This mirrors the contract documented at
// https://kubernetes.io/docs/reference/using-api/health-checks/ and the
// pattern used internally by kube-apiserver.
//
// Probe taxonomy:
//
//   - Liveness: "is this process still functional?"  A failing liveness
//     probe is a signal to a supervisor (systemd / k8s / docker) that
//     the process should be restarted.  Transient dependency failures
//     MUST NOT cause liveness to fail.
//
//   - Readiness: "should this instance receive traffic right now?"  A
//     failing readiness probe removes the instance from load balancers
//     but does not cause a restart.  Dependency failures (database
//     unreachable, cache degraded, leader-election lost) MUST cause
//     readiness to fail.
//
//   - Health (/healthz): legacy combined endpoint.  Implemented here as
//     "liveness AND readiness" for backward compatibility with dashboards
//     and ingress controllers that still look for /healthz.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Probe is the unit of health registration.  Implementations MUST be
// safe for concurrent use and MUST honour the provided context deadline.
// A nil return value indicates success; any non-nil error is reported to
// the caller verbatim (truncated when the endpoint is queried with the
// ?verbose=false form).
type Probe interface {
	// Name returns the stable identifier used in JSON output and in the
	// text form of the /healthz body.  It MUST be a short snake_case or
	// kebab-case token — e.g. "etcd", "postgres-primary".
	Name() string

	// Check performs the probe.  Implementations should complete well
	// within the supplied context deadline; a probe that blocks past
	// the deadline is treated as failed.
	Check(ctx context.Context) error
}

// ProbeFunc adapts a bare function to the Probe interface.
type ProbeFunc struct {
	N string
	F func(ctx context.Context) error
}

// Name implements Probe.
func (p ProbeFunc) Name() string { return p.N }

// Check implements Probe.
func (p ProbeFunc) Check(ctx context.Context) error { return p.F(ctx) }

// Kind selects which aggregate endpoint a probe contributes to.
type Kind int

const (
	// KindLiveness probes contribute to /livez.  Register only probes
	// that fail when the process itself is wedged (deadlock, panic
	// recovery, exhausted goroutine pool).
	KindLiveness Kind = iota
	// KindReadiness probes contribute to /readyz.  Register probes for
	// external dependencies whose failure should drain traffic.
	KindReadiness
)

// Registry is the central coordination point.  A single Registry
// instance is typically constructed in main() and injected into
// subsystems that wish to publish a probe.
type Registry struct {
	mu             sync.RWMutex
	live           map[string]Probe
	ready          map[string]Probe
	defaultTimeout time.Duration
	startupGrace   time.Duration
	startedAt      time.Time
}

// New constructs an empty Registry.  startupGrace is the window after
// process start during which readiness probes report success regardless
// of dependency state — useful to avoid flapping before the first
// reconcile loop has completed.  Pass 0 to disable.
func New(defaultTimeout, startupGrace time.Duration) *Registry {
	if defaultTimeout <= 0 {
		defaultTimeout = 5 * time.Second
	}
	return &Registry{
		live:           make(map[string]Probe),
		ready:          make(map[string]Probe),
		defaultTimeout: defaultTimeout,
		startupGrace:   startupGrace,
		startedAt:      time.Now(),
	}
}

// Register adds a probe under the given kind.  Re-registering the same
// name replaces the prior probe (useful when subsystems restart).
func (r *Registry) Register(kind Kind, p Probe) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch kind {
	case KindLiveness:
		r.live[p.Name()] = p
	case KindReadiness:
		r.ready[p.Name()] = p
	}
}

// Deregister removes a probe.  It is a no-op if the probe was never
// registered.
func (r *Registry) Deregister(kind Kind, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch kind {
	case KindLiveness:
		delete(r.live, name)
	case KindReadiness:
		delete(r.ready, name)
	}
}

// result is the per-probe outcome emitted in JSON responses.
type result struct {
	Name     string `json:"name"`
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

// aggregate is the envelope emitted by /livez, /readyz and /healthz.
type aggregate struct {
	Status    string    `json:"status"`
	Kind      string    `json:"kind"`
	CheckedAt time.Time `json:"checkedAt"`
	Uptime    string    `json:"uptime"`
	Checks    []result  `json:"checks"`
}

// runProbes executes the named set in parallel with the registry's
// default timeout, returning an envelope and an overall-ok bool.
func (r *Registry) runProbes(ctx context.Context, kind Kind, set map[string]Probe) (aggregate, bool) {
	// Snapshot the set under read lock so callers can register probes
	// concurrently without blocking health checks.
	r.mu.RLock()
	snapshot := make([]Probe, 0, len(set))
	for _, p := range set {
		snapshot = append(snapshot, p)
	}
	timeout := r.defaultTimeout
	r.mu.RUnlock()

	results := make([]result, len(snapshot))
	var wg sync.WaitGroup
	for i, p := range snapshot {
		wg.Add(1)
		go func(i int, p Probe) {
			defer wg.Done()
			pCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			start := time.Now()
			err := p.Check(pCtx)
			res := result{
				Name:     p.Name(),
				OK:       err == nil,
				Duration: time.Since(start).String(),
			}
			if err != nil {
				res.Error = err.Error()
			}
			results[i] = res
		}(i, p)
	}
	wg.Wait()

	// Deterministic ordering aids humans reading the /healthz body.
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })

	allOK := true
	for _, r := range results {
		if !r.OK {
			allOK = false
			break
		}
	}

	kindName := "liveness"
	if kind == KindReadiness {
		kindName = "readiness"
	}
	env := aggregate{
		Status:    statusString(allOK),
		Kind:      kindName,
		CheckedAt: time.Now().UTC(),
		Uptime:    time.Since(r.startedAt).String(),
		Checks:    results,
	}
	return env, allOK
}

func statusString(ok bool) string {
	if ok {
		return "ok"
	}
	return "fail"
}

// LivenessHandler returns an http.Handler for /livez.
func (r *Registry) LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.RLock()
		set := r.live
		r.mu.RUnlock()
		env, ok := r.runProbes(req.Context(), KindLiveness, set)
		writeProbeResponse(w, req, env, ok)
	})
}

// ReadinessHandler returns an http.Handler for /readyz.  During the
// startup-grace window the handler reports success even when individual
// probes have not yet transitioned to healthy, matching the behaviour
// of kube-apiserver's readyz shutdown-delay.
func (r *Registry) ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.RLock()
		set := r.ready
		grace := r.startupGrace
		started := r.startedAt
		r.mu.RUnlock()

		env, ok := r.runProbes(req.Context(), KindReadiness, set)

		if !ok && grace > 0 && time.Since(started) < grace {
			// Within the grace window: flip the envelope status to ok
			// but leave the underlying check failures visible.
			env.Status = "ok-startup-grace"
			ok = true
		}
		writeProbeResponse(w, req, env, ok)
	})
}

// HealthHandler returns an http.Handler for the legacy combined /healthz
// endpoint.  It succeeds only when BOTH liveness and readiness succeed.
func (r *Registry) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.RLock()
		live := r.live
		ready := r.ready
		r.mu.RUnlock()

		liveEnv, liveOK := r.runProbes(req.Context(), KindLiveness, live)
		readyEnv, readyOK := r.runProbes(req.Context(), KindReadiness, ready)

		combined := aggregate{
			Status:    statusString(liveOK && readyOK),
			Kind:      "health",
			CheckedAt: time.Now().UTC(),
			Uptime:    liveEnv.Uptime,
			Checks:    append(liveEnv.Checks, readyEnv.Checks...),
		}
		writeProbeResponse(w, req, combined, liveOK && readyOK)
	})
}

// writeProbeResponse emits either JSON (default) or a kube-style text
// body when ?verbose is passed, matching kubectl's `-v` expectations.
func writeProbeResponse(w http.ResponseWriter, req *http.Request, env aggregate, ok bool) {
	status := http.StatusOK
	if !ok {
		status = http.StatusServiceUnavailable
	}

	if req.URL.Query().Has("verbose") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		for _, c := range env.Checks {
			mark := "[+]"
			if !c.OK {
				mark = "[-]"
			}
			_, _ = fmt.Fprintf(w, "%s %s %s %s\n", mark, c.Name, statusString(c.OK), c.Duration)
			if c.Error != "" {
				_, _ = fmt.Fprintf(w, "    error: %s\n", c.Error)
			}
		}
		_, _ = fmt.Fprintf(w, "%s check %s\n", env.Kind, env.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// Mount installs the three standard endpoints on the given mux.  The
// path prefix is applied verbatim — pass "" for the Kubernetes defaults
// ("/livez", "/readyz", "/healthz") or a prefix such as "/axiom" to
// namespace them.
func (r *Registry) Mount(mux *http.ServeMux, prefix string) {
	mux.Handle(prefix+"/livez", r.LivenessHandler())
	mux.Handle(prefix+"/readyz", r.ReadinessHandler())
	mux.Handle(prefix+"/healthz", r.HealthHandler())
}
