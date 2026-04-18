// Package health — common probe implementations used across the code
// base.  These helpers are opinionated wrappers around the generic
// Probe interface so subsystems do not each have to reinvent timeout
// handling or error-message formatting.
package health

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// SQLPingProbe probes a *sql.DB.  Works for any driver registered with
// database/sql — PostgreSQL, MySQL, Oracle, SQLite, etc.
type SQLPingProbe struct {
	ProbeName string
	DB        *sql.DB
}

// Name implements Probe.
func (p *SQLPingProbe) Name() string { return p.ProbeName }

// Check implements Probe by issuing a PingContext against the pool.
func (p *SQLPingProbe) Check(ctx context.Context) error {
	if p.DB == nil {
		return fmt.Errorf("sql handle is nil")
	}
	return p.DB.PingContext(ctx)
}

// TCPProbe opens a plain TCP connection to Addr and closes it
// immediately.  Useful for bare network deps such as Kafka brokers or
// legacy services that do not expose an HTTP health endpoint.
type TCPProbe struct {
	ProbeName string
	Addr      string
}

// Name implements Probe.
func (p *TCPProbe) Name() string { return p.ProbeName }

// Check implements Probe.
func (p *TCPProbe) Check(ctx context.Context) error {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", p.Addr)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

// HTTPProbe performs an HTTP GET against URL and treats any 2xx or 3xx
// response as success.  It reuses a package-level http.Client so that
// probes do not leak sockets when invoked frequently.
type HTTPProbe struct {
	ProbeName string
	URL       string
	Client    *http.Client
}

// Name implements Probe.
func (p *HTTPProbe) Name() string { return p.ProbeName }

// Check implements Probe.
func (p *HTTPProbe) Check(ctx context.Context) error {
	cli := p.Client
	if cli == nil {
		cli = &http.Client{Timeout: 5 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
	if err != nil {
		return err
	}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// ToggleProbe is a probe whose healthiness is controlled by the
// application.  It is useful for readiness gates that flip to
// "unhealthy" during graceful shutdown so load balancers drain traffic
// before the process exits.
//
// Usage:
//
//	t := health.NewToggleProbe("shutdown-gate")
//	registry.Register(health.KindReadiness, t)
//	// … later, when ctx is cancelled:
//	t.Set(false)
type ToggleProbe struct {
	ProbeName string
	healthy   atomic.Bool
}

// NewToggleProbe creates a toggle probe that starts in the healthy
// state.  Callers flip it to unhealthy via Set(false).
func NewToggleProbe(name string) *ToggleProbe {
	p := &ToggleProbe{ProbeName: name}
	p.healthy.Store(true)
	return p
}

// Name implements Probe.
func (p *ToggleProbe) Name() string { return p.ProbeName }

// Check implements Probe.
func (p *ToggleProbe) Check(_ context.Context) error {
	if p.healthy.Load() {
		return nil
	}
	return fmt.Errorf("%s is not healthy", p.ProbeName)
}

// Set flips the probe's healthiness flag.  Safe for concurrent use.
func (p *ToggleProbe) Set(healthy bool) { p.healthy.Store(healthy) }
