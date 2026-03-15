package eventbus

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryEventBusManager in-memory event bus implementation
type InMemoryEventBusManager struct {
	mu            sync.RWMutex
	events        map[string]*EventBusEvent
	topics        map[string]*EventTopic
	subscriptions map[string]*EventSubscription
	dlq           map[string]*DLQEvent
}

// NewInMemoryEventBusManager creates manager
func NewInMemoryEventBusManager() *InMemoryEventBusManager {
	return &InMemoryEventBusManager{
		events:        make(map[string]*EventBusEvent),
		topics:        make(map[string]*EventTopic),
		subscriptions: make(map[string]*EventSubscription),
		dlq:           make(map[string]*DLQEvent),
	}
}

// PublishEvent publishes event
func (m *InMemoryEventBusManager) PublishEvent(event *EventBusEvent) (*EventPublishResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	m.events[event.ID] = event

	// Update topic message count
	topic := m.topics[event.Type]
	if topic != nil {
		topic.MessageCount++
	}

	return &EventPublishResponse{
		EventID:   event.ID,
		Timestamp: event.Timestamp,
		Topic:     event.Type,
	}, nil
}

// GetEvent retrieves event
func (m *InMemoryEventBusManager) GetEvent(id string) (*EventBusEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	event, exists := m.events[id]
	if !exists {
		return nil, fmt.Errorf("event not found")
	}
	return event, nil
}

// ListEvents lists events filtered by tenantID, eventType, and processed status
func (m *InMemoryEventBusManager) ListEvents(tenantID, eventType, processed string) ([]*EventBusEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*EventBusEvent
	for _, e := range m.events {
		if tenantID != "" && e.TenantID != tenantID {
			continue
		}
		if eventType != "" && e.Type != eventType {
			continue
		}
		if processed == "true" && !e.IsProcessed {
			continue
		}
		if processed == "false" && e.IsProcessed {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

// AckEvent marks an event as processed and optionally updates subscription progress.
func (m *InMemoryEventBusManager) AckEvent(eventID, subscriptionID, acknowledgedBy, message string) (*EventBusEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	event, exists := m.events[eventID]
	if !exists {
		return nil, fmt.Errorf("event not found")
	}

	now := time.Now()
	event.IsProcessed = true
	event.ProcessedAt = now

	if event.Metadata == nil {
		event.Metadata = map[string]string{}
	}
	if acknowledgedBy != "" {
		event.Metadata["acknowledgedBy"] = acknowledgedBy
	}
	event.Metadata["acknowledgedAt"] = now.UTC().Format(time.RFC3339)
	if message != "" {
		event.Metadata["ackMessage"] = message
	}
	if subscriptionID != "" {
		event.Metadata["subscriptionId"] = subscriptionID
		if sub, ok := m.subscriptions[subscriptionID]; ok {
			sub.ProcessedCount++
			sub.LastProcessed = now
			sub.UpdatedAt = now
			if event.EventSequence > sub.Offset {
				sub.Offset = event.EventSequence
			}
		}
	}

	return event, nil
}

// CreateTopic creates topic
func (m *InMemoryEventBusManager) CreateTopic(topic *EventTopic) (*EventTopic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if topic.Name == "" {
		return nil, fmt.Errorf("topic name required")
	}

	m.topics[topic.Name] = topic
	return topic, nil
}

// ListTopics lists topics
func (m *InMemoryEventBusManager) ListTopics() ([]*EventTopic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*EventTopic
	for _, t := range m.topics {
		result = append(result, t)
	}
	return result, nil
}

// CreateSubscription creates subscription
func (m *InMemoryEventBusManager) CreateSubscription(sub *EventSubscription) (*EventSubscription, error) {
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
func (m *InMemoryEventBusManager) GetSubscription(id string) (*EventSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sub, exists := m.subscriptions[id]
	if !exists {
		return nil, fmt.Errorf("subscription not found")
	}
	return sub, nil
}

// ListSubscriptions lists subscriptions filtered by tenantID
func (m *InMemoryEventBusManager) ListSubscriptions(tenantID string) ([]*EventSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*EventSubscription
	for _, s := range m.subscriptions {
		if tenantID != "" && s.TenantID != tenantID {
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

// ListDLQEvents lists dead letter queue events filtered by tenantID
func (m *InMemoryEventBusManager) ListDLQEvents(tenantID string) ([]*DLQEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*DLQEvent
	for _, e := range m.dlq {
		if tenantID != "" && e.TenantID != tenantID {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

// ReplayDLQEvent replays a dead-lettered event back into the bus.
func (m *InMemoryEventBusManager) ReplayDLQEvent(dlqID, replayToTopic, replayedBy string) (*EventPublishResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dlqEvent, exists := m.dlq[dlqID]
	if !exists {
		return nil, fmt.Errorf("dlq event not found")
	}

	topic := replayToTopic
	if topic == "" {
		topic = dlqEvent.ReplayToTopic
	}
	if topic == "" {
		topic = dlqEvent.Topic
	}
	if topic == "" {
		topic = dlqEvent.Event.Type
	}
	if topic == "" {
		return nil, fmt.Errorf("replay topic is required")
	}

	now := time.Now()
	replayEvent := dlqEvent.Event
	replayEvent.ID = fmt.Sprintf("event-%d", now.UnixNano())
	replayEvent.Type = topic
	replayEvent.Timestamp = now
	replayEvent.IsProcessed = false
	replayEvent.ProcessedAt = time.Time{}
	replayEvent.DeadLettered = false
	replayEvent.RetryCount = dlqEvent.FailureCount + 1
	if replayEvent.Metadata == nil {
		replayEvent.Metadata = map[string]string{}
	}
	replayEvent.Metadata["replayedFromDLQ"] = dlqID
	if replayedBy != "" {
		replayEvent.Metadata["replayedBy"] = replayedBy
	}

	m.events[replayEvent.ID] = &replayEvent
	if topicState := m.topics[topic]; topicState != nil {
		topicState.MessageCount++
	}

	dlqEvent.ManuallyResolved = true
	dlqEvent.ResolutionAction = "retry"
	dlqEvent.ResolutionTime = now
	dlqEvent.ReplayToTopic = topic

	return &EventPublishResponse{
		EventID:   replayEvent.ID,
		Timestamp: replayEvent.Timestamp,
		Topic:     replayEvent.Type,
	}, nil
}

// PurgeDLQEvent purges DLQ entry
func (m *InMemoryEventBusManager) PurgeDLQEvent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.dlq, id)
	return nil
}
