package bulk

import (
	"fmt"
	"sync"
	"time"
)

const errOperationNotFound = "operation not found"

// InMemoryBulkManager in-memory bulk operations implementation
type InMemoryBulkManager struct {
	mu         sync.RWMutex
	operations map[string]*BulkOperation
	results    map[string]*BulkOperationResponse
}

// NewInMemoryBulkManager creates manager
func NewInMemoryBulkManager() *InMemoryBulkManager {
	return &InMemoryBulkManager{
		operations: make(map[string]*BulkOperation),
		results:    make(map[string]*BulkOperationResponse),
	}
}

// SubmitOperation submits bulk operation
func (m *InMemoryBulkManager) SubmitOperation(op *BulkOperation) (*BulkOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if op.ID == "" {
		op.ID = fmt.Sprintf("bulk-%d", time.Now().UnixNano())
	}

	op.Status = "Pending"
	op.CreatedAt = time.Now()
	op.TotalItems = int64(len(op.Items))
	m.operations[op.ID] = op
	return op, nil
}

// GetOperation retrieves operation
func (m *InMemoryBulkManager) GetOperation(id string) (*BulkOperation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	op, exists := m.operations[id]
	if !exists {
		return nil, fmt.Errorf(errOperationNotFound)
	}
	return op, nil
}

// ListOperations lists operations
func (m *InMemoryBulkManager) ListOperations(tenantID, status string) ([]*BulkOperation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*BulkOperation
	for _, op := range m.operations {
		if tenantID != "" && op.TenantID != tenantID {
			continue
		}
		if status != "" && string(op.Status) != status {
			continue
		}
		result = append(result, op)
	}
	return result, nil
}

// CancelOperation cancels operation
func (m *InMemoryBulkManager) CancelOperation(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	op, exists := m.operations[id]
	if !exists {
		return fmt.Errorf(errOperationNotFound)
	}

	op.Status = "Cancelled"
	now := time.Now()
	op.CompletedAt = &now
	return nil
}

// RetryFailed retries failed items
func (m *InMemoryBulkManager) RetryFailed(id string) (*BulkOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	op, exists := m.operations[id]
	if !exists {
		return nil, fmt.Errorf(errOperationNotFound)
	}

	op.Status = "Running"
	return op, nil
}

// GetResults retrieves operation results
func (m *InMemoryBulkManager) GetResults(id string) (*BulkOperationResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result, exists := m.results[id]
	if !exists {
		return nil, fmt.Errorf("results not found")
	}
	return result, nil
}
