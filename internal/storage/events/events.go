package events

import (
	"log"
	"sync"
	"time"

	"example.com/axiomnizam/internal/storage/models"
	"github.com/google/uuid"
)

// AuditLog records and stores storage operation events.
type AuditLog struct {
	mu     sync.RWMutex
	events []models.StorageEvent
	max    int
}

// NewAuditLog creates a new audit log with max event capacity.
func NewAuditLog(maxEvents int) *AuditLog {
	if maxEvents <= 0 {
		maxEvents = 10000
	}
	return &AuditLog{
		events: make([]models.StorageEvent, 0, 256),
		max:    maxEvents,
	}
}

// Record adds a new storage event to the audit log.
func (a *AuditLog) Record(eventType, tenantID, userID, bucket, key string, size int64, details string) {
	ev := models.StorageEvent{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Type:      eventType,
		TenantID:  tenantID,
		UserID:    userID,
		Bucket:    bucket,
		Key:       key,
		Size:      size,
		Details:   details,
	}

	a.mu.Lock()
	if len(a.events) >= a.max {
		// Evict oldest 10%
		evict := a.max / 10
		if evict < 1 {
			evict = 1
		}
		a.events = a.events[evict:]
	}
	a.events = append(a.events, ev)
	a.mu.Unlock()

	log.Printf("📝 Storage audit: %s tenant=%s bucket=%s key=%s", eventType, tenantID, bucket, key)
}

// List returns events filtered by optional tenant and event type.
// Returns most recent first, up to limit.
func (a *AuditLog) List(tenantID, eventType string, limit int) []models.StorageEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	var result []models.StorageEvent
	// Iterate backwards for most recent first
	for i := len(a.events) - 1; i >= 0 && len(result) < limit; i-- {
		ev := a.events[i]
		if tenantID != "" && ev.TenantID != tenantID {
			continue
		}
		if eventType != "" && ev.Type != eventType {
			continue
		}
		result = append(result, ev)
	}
	return result
}

// Count returns the total number of stored events.
func (a *AuditLog) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.events)
}

// ListByBucket returns events for a specific bucket.
func (a *AuditLog) ListByBucket(bucket string, limit int) []models.StorageEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	var result []models.StorageEvent
	for i := len(a.events) - 1; i >= 0 && len(result) < limit; i-- {
		ev := a.events[i]
		if ev.Bucket == bucket {
			result = append(result, ev)
		}
	}
	return result
}

// Event type constants
const (
	EventBucketCreated    = "bucket.created"
	EventBucketDeleted    = "bucket.deleted"
	EventObjectUploaded   = "object.uploaded"
	EventObjectDownloaded = "object.downloaded"
	EventObjectDeleted    = "object.deleted"
	EventObjectCopied     = "object.copied"
	EventPolicyCreated    = "policy.created"
	EventPolicyDeleted    = "policy.deleted"
	EventPresignGenerated = "presign.generated"
	EventMultiDelete      = "object.multi-deleted"
)
