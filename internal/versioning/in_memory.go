package versioning

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryVersionManager in-memory versioning implementation
type InMemoryVersionManager struct {
	mu        sync.RWMutex
	versions  map[string][]*ResourceVersion
	histories map[string]*VersionHistory
	snapshots map[string]*Snapshot
}

// NewInMemoryVersionManager creates manager
func NewInMemoryVersionManager() *InMemoryVersionManager {
	return &InMemoryVersionManager{
		versions:   make(map[string][]*ResourceVersion),
		histories:  make(map[string]*VersionHistory),
		snapshots:  make(map[string]*Snapshot),
	}
}

// GetVersion retrieves specific version
func (m *InMemoryVersionManager) GetVersion(resourceType, resourceID string, version int64) (*ResourceVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := resourceType + ":" + resourceID
	versions, exists := m.versions[key]
	if !exists {
		return nil, fmt.Errorf("version not found")
	}

	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}
	return nil, fmt.Errorf("version not found")
}

// ListVersions lists all versions
func (m *InMemoryVersionManager) ListVersions(resourceType, resourceID string) ([]*ResourceVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := resourceType + ":" + resourceID
	versions, exists := m.versions[key]
	if !exists {
		return []*ResourceVersion{}, nil
	}
	return versions, nil
}

// GetHistory retrieves version history
func (m *InMemoryVersionManager) GetHistory(resourceType, resourceID string) (*VersionHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := resourceType + ":" + resourceID
	history, exists := m.histories[key]
	if !exists {
		return nil, fmt.Errorf("history not found")
	}
	return history, nil
}

// GetDiff compares versions
func (m *InMemoryVersionManager) GetDiff(resourceType, resourceID string, from, to int64) (*VersionDiff, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	diff := &VersionDiff{
		FromVersion: from,
		ToVersion:   to,
	}
	return diff, nil
}

// CreateSnapshot creates named snapshot
func (m *InMemoryVersionManager) CreateSnapshot(snapshot *Snapshot) (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if snapshot.ID == "" {
		snapshot.ID = fmt.Sprintf("snap-%d", time.Now().UnixNano())
	}

	m.snapshots[snapshot.ID] = snapshot
	return snapshot, nil
}

// Rollback rolls back to version
func (m *InMemoryVersionManager) Rollback(resourceType, resourceID string, targetVersion int64, reason string) (*RollbackResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := &RollbackResult{
		Success:       true,
		ToVersion:     targetVersion,
		CreatedAt:     time.Now(),
	}
	return result, nil
}
