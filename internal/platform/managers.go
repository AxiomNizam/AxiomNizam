package platform

import (
	"example.com/axiomnizam/internal/bulk"
	"example.com/axiomnizam/internal/database"
	"example.com/axiomnizam/internal/eventbus"
	exportpkg "example.com/axiomnizam/internal/export"
	"example.com/axiomnizam/internal/lineage"
	"example.com/axiomnizam/internal/rbac"
	"example.com/axiomnizam/internal/streaming"
	"example.com/axiomnizam/internal/tenant"
	"example.com/axiomnizam/internal/tracing"
	"example.com/axiomnizam/internal/versioning"
	"example.com/axiomnizam/internal/webhooks"
)

// Managers bundles persistent platform manager implementations used by API handlers.
type Managers struct {
	Bulk     bulk.BulkManager
	EventBus eventbus.EventBusManager
	Export   exportpkg.ExportManager
	Stream   streaming.StreamManager
	Webhook  webhooks.WebhookManager
	Tenant   tenant.TenantManager
	RBAC     rbac.RBACManager
	Version  versioning.VersionManager
	Lineage  lineage.LineageManager
	Tracing  tracing.TracingManager
}

// NewManagers creates persistent platform managers.
// When etcd is available, uses etcd-backed persistence.
// When etcd is nil (e.g. STORAGE_BACKEND=raft), uses in-memory-only
// persistence (state is not persisted across restarts but managers
// still function).
func NewManagers(conns *database.Connections) (*Managers, error) {
	store := newPlatformStateStore(conns, "axiomnizam")

	return &Managers{
		Bulk:     newPersistentBulkManager(store),
		EventBus: newPersistentEventBusManager(store),
		Export:   &exportManagerAdapter{base: newPersistentExportCoreManager(store)},
		Stream:   newPersistentStreamManager(store),
		Webhook:  newPersistentWebhookManager(store),
		Tenant:   newPersistentTenantManager(store),
		RBAC:     &rbacManagerAdapter{base: newPersistentRBACCoreManager(store)},
		Version:  newPersistentVersionManager(store),
		Lineage:  &lineageManagerAdapter{base: newPersistentLineageCoreManager(store)},
		Tracing:  &tracingManagerAdapter{base: newPersistentTracingCoreManager(store)},
	}, nil
}
