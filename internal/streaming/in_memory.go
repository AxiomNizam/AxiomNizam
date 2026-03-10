package streaming

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryStreamManager in-memory streaming implementation
type InMemoryStreamManager struct {
	mu            sync.RWMutex
	streams       map[string]*StreamSession
	subscriptions map[string]*StreamSubscription
}

// NewInMemoryStreamManager creates manager
func NewInMemoryStreamManager() *InMemoryStreamManager {
	return &InMemoryStreamManager{
		streams:       make(map[string]*StreamSession),
		subscriptions: make(map[string]*StreamSubscription),
	}
}

// CreateStream creates streaming session
func (m *InMemoryStreamManager) CreateStream(req *StreamRequest) (*StreamSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &StreamSession{
		ID:        fmt.Sprintf("stream-%d", time.Now().UnixNano()),
		CreatedAt: time.Now(),
		Active:    true,
	}

	m.streams[session.ID] = session
	return session, nil
}

// GetStream retrieves stream session
func (m *InMemoryStreamManager) GetStream(id string) (*StreamSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stream, exists := m.streams[id]
	if !exists {
		return nil, fmt.Errorf("stream not found")
	}
	return stream, nil
}

// ListStreams lists active streams
func (m *InMemoryStreamManager) ListStreams(tenantID, status string) ([]*StreamSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*StreamSession
	for _, stream := range m.streams {
		if status == "" || (status == "active" && stream.Active) {
			result = append(result, stream)
		}
	}
	return result, nil
}

// CancelStream cancels stream
func (m *InMemoryStreamManager) CancelStream(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stream, exists := m.streams[id]
	if !exists {
		return fmt.Errorf("stream not found")
	}

	stream.Active = false
	stream.LastActivity = time.Now()
	return nil
}

// Subscribe creates subscription
func (m *InMemoryStreamManager) Subscribe(sub *StreamSubscription) (*StreamSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub.ID == "" {
		sub.ID = fmt.Sprintf("sub-%d", time.Now().UnixNano())
	}

	m.subscriptions[sub.ID] = sub
	return sub, nil
}

// Unsubscribe cancels subscription
func (m *InMemoryStreamManager) Unsubscribe(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.subscriptions, id)
	return nil
}
