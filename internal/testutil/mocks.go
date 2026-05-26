package testutil

import (
	"context"
	"sync"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// MockKVStore — test double for platform/store.KVStore
// ─────────────────────────────────────────────────────────────────────────────

// MockKVStore is an in-memory mock of the KVStore interface.
type MockKVStore struct {
	mu   sync.RWMutex
	data map[string]string
	Err  error
}

// NewMockKVStore creates a new mock KV store.
func NewMockKVStore() *MockKVStore {
	return &MockKVStore{
		data: make(map[string]string),
	}
}

func (m *MockKVStore) Get(_ context.Context, key string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key], nil
}

func (m *MockKVStore) Put(_ context.Context, key, value string) error {
	if m.Err != nil {
		return m.Err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MockKVStore) PutWithTTL(_ context.Context, key, value string, _ time.Duration) error {
	return m.Put(nil, key, value)
}

func (m *MockKVStore) Delete(_ context.Context, key string) error {
	if m.Err != nil {
		return m.Err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MockKVStore) List(_ context.Context, prefix string) (map[string]string, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			result[k] = v
		}
	}
	return result, nil
}

func (m *MockKVStore) CAS(_ context.Context, key, value string) (string, bool, error) {
	if m.Err != nil {
		return "", false, m.Err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	old := m.data[key]
	m.data[key] = value
	return old, true, nil
}

// Keys returns all keys in the mock store (for test assertions).
func (m *MockKVStore) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Count returns the number of entries in the mock store.
func (m *MockKVStore) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// Clear removes all entries from the mock store.
func (m *MockKVStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]string)
}
