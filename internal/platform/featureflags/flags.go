// Package featureflags provides per-module feature flag helpers for the
// migration plan. Flags are read from environment variables and cached
// at first access.
//
// Phase 2: DUAL_WRITE_<MODULE>=true enables dual-write for a module.
// Phase 3: RECONCILER_AUTHORITATIVE_<MODULE>=true makes the reconciler authoritative.
package featureflags

import (
	"os"
	"strings"
	"sync"
)

var (
	mu    sync.RWMutex
	cache map[string]*bool
)

func init() {
	cache = make(map[string]*bool)
}

func envBool(key string) bool {
	mu.RLock()
	if v, ok := cache[key]; ok {
		mu.RUnlock()
		return *v
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	// Double-check after acquiring write lock.
	if v, ok := cache[key]; ok {
		return *v
	}
	val := strings.EqualFold(strings.TrimSpace(os.Getenv(key)), "true")
	cache[key] = &val
	return val
}

// DualWriteEnabled returns true when DUAL_WRITE_<MODULE>=true.
// Module names are uppercased and hyphens replaced with underscores.
func DualWriteEnabled(module string) bool {
	key := "DUAL_WRITE_" + normalizeModule(module)
	return envBool(key)
}

// ReconcilerAuthoritative returns true when RECONCILER_AUTHORITATIVE_<MODULE>=true.
func ReconcilerAuthoritative(module string) bool {
	key := "RECONCILER_AUTHORITATIVE_" + normalizeModule(module)
	return envBool(key)
}

// ShadowMode returns true when RECONCILER_SHADOW_MODE is not explicitly "false".
func ShadowMode() bool {
	raw := strings.TrimSpace(os.Getenv("RECONCILER_SHADOW_MODE"))
	if raw == "" {
		return true // default: shadow mode on
	}
	return !strings.EqualFold(raw, "false")
}

func normalizeModule(module string) string {
	s := strings.ToUpper(module)
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

// StorageBackendIsRaft returns true when STORAGE_BACKEND=raft.
// When true, the platform uses embedded Raft + go-memdb instead of
// etcd for resource persistence.  Default is false (etcd).
func StorageBackendIsRaft() bool {
	return envBool("STORAGE_BACKEND_RAFT")
}

// StorageBackend returns the raw STORAGE_BACKEND env var value.
// Valid values: "etcd" (default), "raft".
func StorageBackend() string {
	raw := strings.TrimSpace(os.Getenv("STORAGE_BACKEND"))
	if raw == "" {
		return "etcd"
	}
	return strings.ToLower(raw)
}

// Reset clears the cache — used in tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	cache = make(map[string]*bool)
}
