package events

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	platformstore "example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/storage/models"
	"github.com/google/uuid"
)

const (
	auditKVKey   = "storage:audit:log"
	auditTimeout = 5 * time.Second
)

// AuditLog records and stores storage operation events.
type AuditLog struct {
	mu     sync.RWMutex
	events []models.StorageEvent
	max    int

	kvStore platformstore.KVStore
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
// Accepts a pre-built StorageEvent struct.
func (a *AuditLog) Record(ev models.StorageEvent) {
	if ev.ID == "" {
		ev.ID = uuid.New().String()
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
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

	log.Printf("📝 Storage audit: %s tenant=%s bucket=%s key=%s", ev.Type, ev.TenantID, ev.Bucket, ev.Key)

	// Async save to persistent store.
	go a.save()
}

// ConfigureKVPersistence enables KVStore-backed persistence for the audit log.
func (a *AuditLog) ConfigureKVPersistence(kv platformstore.KVStore) {
	a.mu.Lock()
	a.kvStore = kv
	a.mu.Unlock()
	a.load()
}

func (a *AuditLog) load() {
	a.mu.Lock()
	kv := a.kvStore
	a.mu.Unlock()
	if kv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), auditTimeout)
	defer cancel()

	val, err := kv.Get(ctx, auditKVKey)
	if err != nil {
		return // not found
	}

	var events []models.StorageEvent
	if err := json.Unmarshal([]byte(val), &events); err != nil {
		log.Printf("⚠️  storage audit: failed to unmarshal events: %v", err)
		return
	}

	a.mu.Lock()
	a.events = events
	a.mu.Unlock()
	log.Printf("✅ storage audit: loaded %d persistent events", len(events))
}

func (a *AuditLog) save() {
	a.mu.RLock()
	kv := a.kvStore
	if kv == nil {
		a.mu.RUnlock()
		return
	}
	// Take a small snapshot of recent events to avoid giant KV values.
	// We only persist up to 1000 most recent events to Raft.
	const maxPersistentEvents = 1000
	persistEvents := a.events
	if len(persistEvents) > maxPersistentEvents {
		persistEvents = persistEvents[len(persistEvents)-maxPersistentEvents:]
	}
	a.mu.RUnlock()

	data, err := json.Marshal(persistEvents)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), auditTimeout)
	defer cancel()
	_ = kv.Put(ctx, auditKVKey, string(data))
}

// List returns events filtered by optional tenant and event type.
// Returns most recent first, up to limit.
func (a *AuditLog) List(tenantID, eventType string, limit int) []models.StorageEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	result := make([]models.StorageEvent, 0)
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

	result := make([]models.StorageEvent, 0)
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
	EventBucketCreated         = "bucket.created"
	EventBucketDeleted         = "bucket.deleted"
	EventObjectUploaded        = "object.uploaded"
	EventObjectDownloaded      = "object.downloaded"
	EventObjectDeleted         = "object.deleted"
	EventObjectCopied          = "object.copied"
	EventPolicyCreated         = "policy.created"
	EventPolicyDeleted         = "policy.deleted"
	EventPresignGenerated      = "presign.generated"
	EventMultiDelete           = "object.multi-deleted"
	EventObjectScanClean       = "object.scan.clean"
	EventObjectThreatDetected  = "object.scan.threat"
)
