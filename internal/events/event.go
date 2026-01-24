package events

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

// Common event types
const (
	EventTypeUserCreated   EventType = "user.created"
	EventTypeUserUpdated   EventType = "user.updated"
	EventTypeUserDeleted   EventType = "user.deleted"
	EventTypeUserLoggedIn  EventType = "user.logged_in"
	EventTypeUserLoggedOut EventType = "user.logged_out"
	EventTypeDataExported  EventType = "data.exported"
	EventTypeJobStarted    EventType = "job.started"
	EventTypeJobCompleted  EventType = "job.completed"
	EventTypeJobFailed     EventType = "job.failed"
	EventTypeErrorOccurred EventType = "error.occurred"
)

// Event represents a domain event
type Event struct {
	ID            string                 `json:"id"`
	Type          EventType              `json:"type"`
	Source        string                 `json:"source"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
	UserID        string                 `json:"user_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// EventHandler is a function that handles events
type EventHandler func(ctx context.Context, event *Event) error

// Bus defines the event bus interface
type Bus interface {
	// Publish publishes an event
	Publish(ctx context.Context, event *Event) error

	// Subscribe subscribes to events of a type
	Subscribe(eventType EventType, handler EventHandler) error

	// Unsubscribe removes a subscription
	Unsubscribe(eventType EventType, handler EventHandler) error

	// SubscribeAll subscribes to all events
	SubscribeAll(handler EventHandler) error

	// GetEventHistory retrieves event history
	GetEventHistory(ctx context.Context, eventType EventType, limit int) ([]*Event, error)

	// GetStats returns bus statistics
	GetStats() *BusStats
}

// BusStats contains event bus statistics
type BusStats struct {
	TotalEvents     int64
	EventsByType    map[EventType]int64
	SubscriberCount int
	HandlerErrors   int64
}

// MemoryBus implements Bus interface using in-memory storage
type MemoryBus struct {
	mu           sync.RWMutex
	handlers     map[EventType][]EventHandler
	allHandlers  []EventHandler
	history      []*Event
	maxHistory   int
	logger       *log.Logger
	stats        *BusStats
	asyncMode    bool
	errorHandler func(error)
}

// NewMemoryBus creates a new in-memory event bus
func NewMemoryBus(maxHistory int) *MemoryBus {
	if maxHistory <= 0 {
		maxHistory = 10000
	}

	return &MemoryBus{
		handlers:    make(map[EventType][]EventHandler),
		allHandlers: make([]EventHandler, 0),
		history:     make([]*Event, 0),
		maxHistory:  maxHistory,
		logger:      log.New(log.Writer(), "[EVENT_BUS] ", log.LstdFlags),
		stats: &BusStats{
			EventsByType: make(map[EventType]int64),
		},
		asyncMode: true,
	}
}

// SetAsyncMode sets whether to handle events asynchronously
func (mb *MemoryBus) SetAsyncMode(async bool) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.asyncMode = async
}

// SetErrorHandler sets a handler for errors
func (mb *MemoryBus) SetErrorHandler(handler func(error)) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.errorHandler = handler
}

// Publish publishes an event
func (mb *MemoryBus) Publish(ctx context.Context, event *Event) error {
	if event == nil {
		return errors.New("event cannot be nil")
	}

	if event.ID == "" {
		event.ID = generateEventID()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add to history
	mb.mu.Lock()
	mb.history = append(mb.history, event)
	if len(mb.history) > mb.maxHistory {
		mb.history = mb.history[1:]
	}

	mb.stats.TotalEvents++
	mb.stats.EventsByType[event.Type]++

	// Get handlers
	typeHandlers := mb.handlers[event.Type]
	handlers := make([]EventHandler, len(typeHandlers)+len(mb.allHandlers))
	copy(handlers, typeHandlers)
	copy(handlers[len(typeHandlers):], mb.allHandlers)

	mb.mu.Unlock()

	mb.logger.Printf("Event published: %s (id: %s)", event.Type, event.ID)

	// Execute handlers
	if mb.asyncMode {
		go mb.executeHandlers(ctx, event, handlers)
	} else {
		mb.executeHandlers(ctx, event, handlers)
	}

	return nil
}

// executeHandlers executes all handlers for an event
func (mb *MemoryBus) executeHandlers(ctx context.Context, event *Event, handlers []EventHandler) {
	for _, handler := range handlers {
		if handler == nil {
			continue
		}

		if err := handler(ctx, event); err != nil {
			mb.logger.Printf("Handler error for event %s: %v", event.Type, err)

			mb.mu.Lock()
			mb.stats.HandlerErrors++
			mb.mu.Unlock()

			if mb.errorHandler != nil {
				mb.errorHandler(err)
			}
		}
	}
}

// Subscribe subscribes to events of a type
func (mb *MemoryBus) Subscribe(eventType EventType, handler EventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.handlers[eventType] = append(mb.handlers[eventType], handler)
	mb.logger.Printf("Handler subscribed to event type: %s", eventType)

	return nil
}

// Unsubscribe removes a subscription
func (mb *MemoryBus) Unsubscribe(eventType EventType, handler EventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	mb.mu.Lock()
	defer mb.mu.Unlock()

	handlers := mb.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			mb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			mb.logger.Printf("Handler unsubscribed from event type: %s", eventType)
			return nil
		}
	}

	return errors.New("handler not found")
}

// SubscribeAll subscribes to all events
func (mb *MemoryBus) SubscribeAll(handler EventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.allHandlers = append(mb.allHandlers, handler)
	mb.logger.Printf("Handler subscribed to all events")

	return nil
}

// GetEventHistory retrieves event history
func (mb *MemoryBus) GetEventHistory(ctx context.Context, eventType EventType, limit int) ([]*Event, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	var results []*Event
	for i := len(mb.history) - 1; i >= 0; i-- {
		event := mb.history[i]
		if event.Type == eventType {
			results = append(results, event)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// GetStats returns bus statistics
func (mb *MemoryBus) GetStats() *BusStats {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	statsCopy := *mb.stats
	statsCopy.SubscriberCount = len(mb.allHandlers)

	for _, handlers := range mb.handlers {
		statsCopy.SubscriberCount += len(handlers)
	}

	return &statsCopy
}

// CreateEvent creates a new event
func CreateEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	if data == nil {
		data = make(map[string]interface{})
	}

	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    source,
		Data:      data,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// CreateEventWithUser creates an event with user context
func CreateEventWithUser(eventType EventType, source string, userID string, data map[string]interface{}) *Event {
	event := CreateEvent(eventType, source, data)
	event.UserID = userID
	return event
}

// SetCorrelationID sets the correlation ID for event tracing
func (e *Event) SetCorrelationID(id string) {
	e.CorrelationID = id
}

// AddMetadata adds metadata to the event
func (e *Event) AddMetadata(key, value string) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return "evt_" + time.Now().Format("20060102150405") + "_" + randString(8)
}

// randString generates a random string
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// EventFilter represents an event filter
type EventFilter struct {
	EventType EventType
	Source    string
	UserID    string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// Filter filters events based on criteria
func (mb *MemoryBus) Filter(filter *EventFilter) []*Event {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	var results []*Event

	for _, event := range mb.history {
		if filter.EventType != "" && event.Type != filter.EventType {
			continue
		}
		if filter.Source != "" && event.Source != filter.Source {
			continue
		}
		if filter.UserID != "" && event.UserID != filter.UserID {
			continue
		}
		if !filter.StartTime.IsZero() && event.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && event.Timestamp.After(filter.EndTime) {
			continue
		}

		results = append(results, event)
		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}

	return results
}
