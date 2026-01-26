package eventbus

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryEventBusManager in-memory event bus implementation
type InMemoryEventBusManager struct {
	mu            sync.RWMutex
	events        map[string]*Event
	topics        map[string]*Topic
	subscriptions map[string]*Subscription
	dlq           map[string]*DeadLetterEvent
}

// NewInMemoryEventBusManager creates manager
func NewInMemoryEventBusManager() *InMemoryEventBusManager {
	return &InMemoryEventBusManager{
		events:        make(map[string]*Event),
		topics:        make(map[string]*Topic),
		subscriptions: make(map[string]*Subscription),
		dlq:           make(map[string]*DeadLetterEvent),
	}
}

// PublishEvent publishes event
func (m *InMemoryEventBusManager) PublishEvent(event *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	m.events[event.ID] = event

	// Deliver to subscriptions
	topic := m.topics[event.Topic]
	if topic != nil {
		topic.EventCount++
	}

	return nil
}

// GetEvent retrieves event
func (m *InMemoryEventBusManager) GetEvent(id string) (*Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	event, exists := m.events[id]
	if !exists {
		return nil, fmt.Errorf("event not found")
	}
	return event, nil
}

// ListEvents lists events
func (m *InMemoryEventBusManager) ListEvents(topic string, limit int) ([]*Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Event
	for _, e := range m.events {
		if topic != "" && e.Topic != topic {
			continue
		}
		result = append(result, e)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

// CreateTopic creates topic
func (m *InMemoryEventBusManager) CreateTopic(topic *Topic) (*Topic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if topic.Name == "" {
		return nil, fmt.Errorf("topic name required")
	}

	m.topics[topic.Name] = topic
	return topic, nil
}

// ListTopics lists topics
func (m *InMemoryEventBusManager) ListTopics() ([]*Topic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Topic
	for _, t := range m.topics {
		result = append(result, t)
	}
	return result, nil
}

// CreateSubscription creates subscription
func (m *InMemoryEventBusManager) CreateSubscription(sub *Subscription) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub.ID == "" {
		sub.ID = fmt.Sprintf("sub-%d", time.Now().UnixNano())
	}
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now()
	}

	m.subscriptions[sub.ID] = sub
	return sub, nil
}

// GetSubscription retrieves subscription
func (m *InMemoryEventBusManager) GetSubscription(id string) (*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sub, exists := m.subscriptions[id]
	if !exists {
		return nil, fmt.Errorf("subscription not found")
	}
	return sub, nil
}

// ListSubscriptions lists subscriptions
func (m *InMemoryEventBusManager) ListSubscriptions(topic string) ([]*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Subscription
	for _, s := range m.subscriptions {
		if topic != "" && s.Topic != topic {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

// DeleteSubscription deletes subscription
func (m *InMemoryEventBusManager) DeleteSubscription(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.subscriptions, id)
	return nil
}

// ListDLQ lists dead letter queue
func (m *InMemoryEventBusManager) ListDLQ(topic string) ([]*DeadLetterEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*DeadLetterEvent
	for _, e := range m.dlq {
		if topic != "" && e.Topic != topic {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

// PurgeDeadLetterEvent purges DLQ entry
func (m *InMemoryEventBusManager) PurgeDeadLetterEvent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.dlq, id)
	return nil
}
