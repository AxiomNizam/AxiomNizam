package contracts

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Module is the lifecycle contract every AxiomNizam module must implement.
// It provides a uniform way to initialize, configure, start, and stop modules.
//
// Modules that don't support routes or KV persistence can leave those methods
// as no-ops (return nil).
type Module interface {
	// Name returns the module identifier (e.g., "gatekeeper", "storage", "iam").
	Name() string

	// RegisterRoutes registers the module's HTTP routes on the given router group.
	// Modules without HTTP endpoints can return nil.
	RegisterRoutes(rg *gin.RouterGroup) error

	// Start initializes the module — starts controllers, schedulers, background workers.
	// The context controls the module's lifetime; canceling it should trigger shutdown.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the module — drains queues, persists state, closes connections.
	Stop() error
}

// KVStoreProvider is implemented by modules that need Raft KV persistence.
// The platform calls SetKVStore during bootstrap when STORAGE_BACKEND=raft.
type KVStoreProvider interface {
	SetKVStore(kv interface{})
}

// RoutesRegistrar is a convenience interface for modules that only need route registration
// without the full lifecycle. Use this for lightweight modules (scanners, utilities).
type RoutesRegistrar interface {
	RegisterRoutes(rg *gin.RouterGroup)
}

// Startable is a convenience interface for modules that need background work
// but don't expose HTTP routes (e.g., schedulers, workers).
type Startable interface {
	Start(ctx context.Context) error
	Stop() error
}
