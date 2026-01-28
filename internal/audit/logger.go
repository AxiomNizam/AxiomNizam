package audit

import (
	"context"
)

// AuditLogger logs audit events
type AuditLogger interface {
	// LogAction logs an action with full context
	LogAction(ctx context.Context, log *AuditLog) error

	// QueryLogs searches audit logs with filter
	QueryLogs(ctx context.Context, filter *AuditFilter) ([]*AuditLog, error)

	// GetReport generates audit report
	GetReport(ctx context.Context, filter *AuditFilter) (*AuditReport, error)

	// DeleteOldLogs removes logs older than retention period
	DeleteOldLogs(ctx context.Context, olderThanDays int) error

	// VerifyIntegrity verifies audit log immutability
	VerifyIntegrity(ctx context.Context, logID string) (bool, error)
}
