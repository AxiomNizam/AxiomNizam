package audit

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// InMemoryAuditLogger in-memory implementation
type InMemoryAuditLogger struct {
	mu    sync.RWMutex
	logs  map[string]*AuditLog
	index map[string][]string // Tenant -> log IDs
}

// NewInMemoryAuditLogger creates logger
func NewInMemoryAuditLogger() *InMemoryAuditLogger {
	return &InMemoryAuditLogger{
		logs:  make(map[string]*AuditLog),
		index: make(map[string][]string),
	}
}

// LogAction records audit action
func (l *InMemoryAuditLogger) LogAction(log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if log.ID == "" {
		log.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}

	// Calculate hash for immutability
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v%v%v", log.User, log.Action, log.Timestamp)))
	log.Hash = fmt.Sprintf("%x", hash)

	l.logs[log.ID] = log
	l.index[log.TenantID] = append(l.index[log.TenantID], log.ID)
	return nil
}

// QueryLogs retrieves logs matching filter
func (l *InMemoryAuditLogger) QueryLogs(filter *AuditFilter) ([]*AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var results []*AuditLog

	ids := l.index[filter.TenantID]
	for _, id := range ids {
		log := l.logs[id]
		if l.matchesFilter(log, filter) {
			results = append(results, log)
		}
	}

	return results, nil
}

// GetReport generates audit report
func (l *InMemoryAuditLogger) GetReport(tenantID string) (*AuditReport, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	report := &AuditReport{
		TenantID:    tenantID,
		GeneratedAt: time.Now(),
	}

	ids := l.index[tenantID]
	for _, id := range ids {
		log := l.logs[id]
		report.TotalActions++
		if log.Result == "SUCCESS" {
			report.SuccessfulActions++
		} else if log.Result == "FAILURE" {
			report.FailedActions++
		}
	}

	return report, nil
}

// DeleteOldLogs removes logs older than days
func (l *InMemoryAuditLogger) DeleteOldLogs(tenantID, days string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var ids []string
	for _, id := range l.index[tenantID] {
		if log, exists := l.logs[id]; exists {
			ids = append(ids, id)
			delete(l.logs, id)
		}
	}

	l.index[tenantID] = ids
	return nil
}

// VerifyIntegrity verifies log integrity
func (l *InMemoryAuditLogger) VerifyIntegrity(logID string) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	log, exists := l.logs[logID]
	if !exists {
		return false, fmt.Errorf("log not found")
	}

	hash := sha256.Sum256([]byte(fmt.Sprintf("%v%v%v", log.User, log.Action, log.Timestamp)))
	return log.Hash == fmt.Sprintf("%x", hash), nil
}

func (l *InMemoryAuditLogger) matchesFilter(log *AuditLog, filter *AuditFilter) bool {
	if filter.User != "" && log.User != filter.User {
		return false
	}
	if filter.Resource != "" && log.Resource != filter.Resource {
		return false
	}
	if filter.Action != "" && log.Action != AuditAction(filter.Action) {
		return false
	}
	return true
}
