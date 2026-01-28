package migrations

import (
	"AxiomNizam/internal/models"

	"gorm.io/gorm"
)

// RunMigrations executes all database migrations
func RunMigrations(db *gorm.DB) error {
	// Audit migrations
	if err := db.AutoMigrate(&models.AuditLogModel{}); err != nil {
		return err
	}

	// Tenant migrations
	if err := db.AutoMigrate(&models.TenantModel{}, &models.TenantMemberModel{}, &models.TenantQuotaModel{}); err != nil {
		return err
	}

	// Job migrations
	if err := db.AutoMigrate(&models.JobModel{}, &models.JobLogModel{}); err != nil {
		return err
	}

	// Streaming migrations
	if err := db.AutoMigrate(&models.StreamModel{}, &models.StreamSubscriptionModel{}); err != nil {
		return err
	}

	// Bulk operation migrations
	if err := db.AutoMigrate(&models.BulkOperationModel{}, &models.BulkResultModel{}); err != nil {
		return err
	}

	// Versioning migrations
	if err := db.AutoMigrate(&models.ResourceVersionModel{}, &models.VersionSnapshotModel{}); err != nil {
		return err
	}

	// Webhook migrations
	if err := db.AutoMigrate(&models.WebhookModel{}, &models.WebhookDeliveryLogModel{}); err != nil {
		return err
	}

	// Event bus migrations
	if err := db.AutoMigrate(
		&models.EventModel{},
		&models.TopicModel{},
		&models.SubscriptionModel{},
		&models.DeadLetterEventModel{},
	); err != nil {
		return err
	}

	// Tracing migrations
	if err := db.AutoMigrate(&models.TraceModel{}, &models.SpanModel{}, &models.ServiceMetricsModel{}); err != nil {
		return err
	}

	// Export migrations
	if err := db.AutoMigrate(&models.ExportJobModel{}, &models.ExportResultModel{}, &models.ExportTemplateModel{}); err != nil {
		return err
	}

	// Lineage migrations
	if err := db.AutoMigrate(
		&models.LineageNodeModel{},
		&models.LineageEdgeModel{},
		&models.LineageProcessModel{},
	); err != nil {
		return err
	}

	// Encryption migrations
	if err := db.AutoMigrate(
		&models.EncryptionKeyModel{},
		&models.EncryptionPolicyModel{},
		&models.KeyRotationModel{},
		&models.EncryptionAuditLogModel{},
	); err != nil {
		return err
	}

	// RBAC migrations
	if err := db.AutoMigrate(
		&models.RoleModel{},
		&models.RoleBindingModel{},
		&models.PermissionModel{},
		&models.AccessRequestModel{},
	); err != nil {
		return err
	}

	// Create indexes for performance
	if err := createIndexes(db); err != nil {
		return err
	}

	return nil
}

// createIndexes creates additional performance indexes
func createIndexes(db *gorm.DB) error {
	// Audit indexes
	if err := db.Model(&models.AuditLogModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_audit_tenant_action ON audit_logs(tenant_id, action_type, created_at)").Error; err != nil {
		return err
	}

	// Job indexes
	if err := db.Model(&models.JobModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_job_tenant_status ON jobs(tenant_id, status, created_at)").Error; err != nil {
		return err
	}

	// Stream indexes
	if err := db.Model(&models.StreamModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_stream_tenant_status ON streams(tenant_id, status, created_at)").Error; err != nil {
		return err
	}

	// Bulk indexes
	if err := db.Model(&models.BulkOperationModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_bulk_tenant_status ON bulk_operations(tenant_id, status, created_at)").Error; err != nil {
		return err
	}

	// Version indexes
	if err := db.Model(&models.ResourceVersionModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_version_resource ON resource_versions(tenant_id, resource_type, resource_id, version_number DESC)").Error; err != nil {
		return err
	}

	// Webhook indexes
	if err := db.Model(&models.WebhookModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_webhook_tenant ON webhooks(tenant_id, active, created_at)").Error; err != nil {
		return err
	}

	// Event indexes
	if err := db.Model(&models.EventModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_event_topic ON events(tenant_id, topic, created_at DESC)").Error; err != nil {
		return err
	}

	// Subscription indexes
	if err := db.Model(&models.SubscriptionModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_subscription_topic ON subscriptions(tenant_id, topic, status)").Error; err != nil {
		return err
	}

	// Trace indexes
	if err := db.Model(&models.TraceModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_trace_service ON traces(tenant_id, service, start_time DESC)").Error; err != nil {
		return err
	}

	// Span indexes
	if err := db.Model(&models.SpanModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_span_trace ON spans(trace_id, start_time ASC)").Error; err != nil {
		return err
	}

	// Export indexes
	if err := db.Model(&models.ExportJobModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_export_tenant_status ON export_jobs(tenant_id, status, created_at DESC)").Error; err != nil {
		return err
	}

	// Lineage indexes
	if err := db.Model(&models.LineageEdgeModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_lineage_edges ON lineage_edges(tenant_id, source_node_id, target_node_id)").Error; err != nil {
		return err
	}

	// Encryption indexes
	if err := db.Model(&models.EncryptionKeyModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_encryption_key_tenant ON encryption_keys(tenant_id, status, created_at DESC)").Error; err != nil {
		return err
	}

	// RBAC indexes
	if err := db.Model(&models.RoleModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_role_tenant_level ON roles(tenant_id, level DESC)").Error; err != nil {
		return err
	}

	if err := db.Model(&models.RoleBindingModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_binding_subject ON role_bindings(tenant_id, subject_id)").Error; err != nil {
		return err
	}

	if err := db.Model(&models.AccessRequestModel{}).
		Exec("CREATE INDEX IF NOT EXISTS idx_access_request_subject ON access_requests(tenant_id, subject_id, status)").Error; err != nil {
		return err
	}

	return nil
}
