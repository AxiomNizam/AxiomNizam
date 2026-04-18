// Package rpcpool is a modest connection pool in the style of
// Nomad's yamux-based RPC pool: it keeps at most N connections per
// target, hands them out to callers, and reuses them until they
// fail or go idle too long.
//
// The package does not pick a wire format — callers Dial a
// transport-specific connection and the pool tracks its liveness.
// In AxiomNizam this sits under the TLS RPC path to the conductor
// service.
package rpcpool

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"
)

// DialFunc creates a fresh connection to addr.
type DialFunc func(ctx context.Context, addr string) (io.Closer, error)

// Config tunes the pool.
type Config struct {
	// MaxPerTarget caps concurrent connections per address.
	MaxPerTarget int
	// IdleTimeout evicts connections unused for this long.
	IdleTimeout time.Duration
	// Dial creates new connections on demand.
	Dial DialFunc
}

// Pool manages connections.
type Pool struct {
	cfg Config

	mu    sync.Mutex
	store map[string][]*entry // key = target address
}

// entry holds one pooled connection.
type entry struct {
	conn     io.Closer
	lastUsed time.Time
}

// New constructs a Pool.
func New(cfg Config) *Pool {
	if cfg.MaxPerTarget <= 0 {
		cfg.MaxPerTarget = 4
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 2 * time.Minute
	}
	return &Pool{cfg: cfg, store: map[string][]*entry{}}
}

// ErrPoolClosed is returned when the pool has been Closed.
var ErrPoolClosed = errors.New("rpcpool: closed")

// Get borrows a connection.  Callers must call Put or Discard when done.
func (p *Pool) Get(ctx context.Context, addr string) (io.Closer, error) {
	p.mu.Lock()
	entries := p.store[addr]
	// Evict idle entries from the head.
	cutoff := time.Now().Add(-p.cfg.IdleTimeout)
	for len(entries) > 0 && entries[0].lastUsed.Before(cutoff) {
		_ = entries[0].conn.Close()
		entries = entries[1:]
	}
	if len(entries) > 0 {
		e := entries[len(entries)-1]
		entries = entries[:len(entries)-1]
		p.store[addr] = entries
		p.mu.Unlock()
		return e.conn, nil
	}
	p.store[addr] = entries
	p.mu.Unlock()

	if p.cfg.Dial == nil {
		return nil, errors.New("rpcpool: no Dial func configured")
	}
	return p.cfg.Dial(ctx, addr)
}

// Put returns a connection to the pool.  If the pool is full, the
// connection is closed instead of retained.
func (p *Pool) Put(addr string, c io.Closer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.store[addr]) >= p.cfg.MaxPerTarget {
		_ = c.Close()
		return
	}
	p.store[addr] = append(p.store[addr], &entry{conn: c, lastUsed: time.Now()})
}

// Discard closes a connection without returning it to the pool.
// Used after a known-bad error (I/O failure, protocol violation).
func (p *Pool) Discard(c io.Closer) { _ = c.Close() }

// Close drains every pooled connection.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var firstErr error
	for addr, list := range p.store {
		for _, e := range list {
			if err := e.conn.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		delete(p.store, addr)
	}
	return firstErr
}
