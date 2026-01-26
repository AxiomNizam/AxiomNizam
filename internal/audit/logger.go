package audit

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
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

// InMemoryAuditLogger stores logs in memory (for testing)
type InMemoryAuditLogger struct {
	mu   sync.RWMutex
	logs []*AuditLog
	cfg  *AuditConfig
}

// NewInMemoryAuditLogger creates in-memory logger
func NewInMemoryAuditLogger(cfg *AuditConfig) *InMemoryAuditLogger {
	if cfg == nil {
		cfg = &AuditConfig{
			Enabled:       true,
			RetentionDays: 90,
		}
	}
	return &InMemoryAuditLogger{
		logs: make([]*AuditLog, 0),
		cfg:  cfg,
	}
}

// LogAction logs action
func (ial *InMemoryAuditLogger) LogAction(ctx context.Context, log *AuditLog) error {
	if !ial.cfg.Enabled {
		return nil
	}

	ial.mu.Lock()
	defer ial.mu.Unlock()

	// Generate immutable hash
	log.ID = fmt.Sprintf("%s-%d", log.ResourceID, time.Now().UnixNano())
	log.ImmutableHash = ial.calculateHash(log)

	ial.logs = append(ial.logs, log)
	return nil
}

// QueryLogs queries logs
func (ial *InMemoryAuditLogger) QueryLogs(ctx context.Context, filter *AuditFilter) ([]*AuditLog, error) {
	ial.mu.RLock()
	defer ial.mu.RUnlock()

	result := make([]*AuditLog, 0)

	for _, log := range ial.logs {
		if ial.matches(log, filter) {
			result = append(result, log)
		}
	}

	// Apply limit/offset
	start := filter.Offset
	if start > len(result) {
		start = len(result)
	}
	end := start + filter.Limit
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// GetReport generates report
func (ial *InMemoryAuditLogger) GetReport(ctx context.Context, filter *AuditFilter) (*AuditReport, error) {
	logs, err := ial.QueryLogs(ctx, &AuditFilter{
		TenantID:     filter.TenantID,
		UserID:       filter.UserID,
		Action:       filter.Action,
		Result:       filter.Result,
		ResourceType: filter.ResourceType,
		StartTime:    filter.StartTime,
		EndTime:      filter.EndTime,
		Limit:        10000,
	})
	if err != nil {
		return nil, err
	}

	report := &AuditReport{
		TotalRecords:      int64(len(logs)),
		DateRange:         DateRange{Start: filter.StartTime, End: filter.EndTime},
		ActionBreakdown:   make(map[string]int64),
		ResultBreakdown:   make(map[string]int64),
		UserBreakdown:     make(map[string]int64),
		ResourceBreakdown: make(map[string]int64),
		HighRiskActions:   make([]AuditLog, 0),
	}

	successCount := int64(0)
	failureCount := int64(0)

	for _, log := range logs {
		report.ActionBreakdown[string(log.Action)]++
		report.ResultBreakdown[string(log.Result)]++
		report.UserBreakdown[log.Username]++
		report.ResourceBreakdown[log.ResourceType]++

		if log.Result == ResultSuccess {
			successCount++
		} else {
			failureCount++
		}

		// Track high-risk actions
		if ial.isHighRisk(log) {
			report.HighRiskActions = append(report.HighRiskActions, *log)
		}
	}

	if report.TotalRecords > 0 {
		report.FailureRate = float64(failureCount) / float64(report.TotalRecords) * 100
	}

	return report, nil
}

// DeleteOldLogs removes old logs
func (ial *InMemoryAuditLogger) DeleteOldLogs(ctx context.Context, olderThanDays int) error {
	ial.mu.Lock()
	defer ial.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	filtered := make([]*AuditLog, 0)

	for _, log := range ial.logs {
		if log.Timestamp.After(cutoff) {
			filtered = append(filtered, log)
		}
	}

	ial.logs = filtered
	return nil
}

// VerifyIntegrity verifies hash
func (ial *InMemoryAuditLogger) VerifyIntegrity(ctx context.Context, logID string) (bool, error) {
	ial.mu.RLock()
	defer ial.mu.RUnlock()

	for _, log := range ial.logs {
		if log.ID == logID {
			expectedHash := ial.calculateHash(log)
			return log.ImmutableHash == expectedHash, nil
		}
	}

	return false, fmt.Errorf("log not found: %s", logID)
}

// Helper functions
func (ial *InMemoryAuditLogger) matches(log *AuditLog, filter *AuditFilter) bool {
	if filter.TenantID != "" && log.TenantID != filter.TenantID {
		return false
	}
	if filter.UserID != "" && log.UserID != filter.UserID {
		return false
	}
	if filter.Username != "" && log.Username != filter.Username {
		return false
	}
	if filter.Action != "" && log.Action != filter.Action {
		return false
	}
	if filter.Result != "" && log.Result != filter.Result {
		return false
	}
	if filter.ResourceType != "" && log.ResourceType != filter.ResourceType {
		return false
	}
	if filter.ResourceID != "" && log.ResourceID != filter.ResourceID {
		return false
	}
	if filter.Namespace != "" && log.Namespace != filter.Namespace {
		return false
	}
	if filter.SourceIP != "" && log.SourceIP != filter.SourceIP {
		return false
	}
	if !filter.StartTime.IsZero() && log.Timestamp.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && log.Timestamp.After(filter.EndTime) {
		return false
	}
	return true
}

func (ial *InMemoryAuditLogger) isHighRisk(log *AuditLog) bool {
	highRiskActions := []AuditAction{
		ActionDelete,
		ActionPolicyChange,
	}
	for _, action := range highRiskActions {
		if log.Action == action {
			return true
		}
	}
	return false
}

func (ial *InMemoryAuditLogger) calculateHash(log *AuditLog) string {
	// Simple hash of key fields for immutability verification
	hashInput := fmt.Sprintf("%s:%s:%s:%s:%d", log.ID, log.UserID, log.Action, log.ResourceID, log.Timestamp.Unix())
	hash := sha256.Sum256([]byte(hashInput))
	return fmt.Sprintf("%x", hash)
}
