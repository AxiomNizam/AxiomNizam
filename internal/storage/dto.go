package storage

import "time"

// --- Bucket DTOs ---

// BucketListItem is a typed entry in the bucket list.
type BucketListItem struct {
	Name        string            `json:"name"`
	Region      string            `json:"region"`
	Versioning  bool              `json:"versioning"`
	Tags        map[string]string `json:"tags,omitempty"`
	Status      string            `json:"status"`
	AccessCount int               `json:"access_count"`
	LastAccess  *time.Time        `json:"last_access,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// BucketListResponse is the typed response for GET /api/v1/storage.
type BucketListResponse struct {
	Status  string           `json:"status"`
	Buckets []BucketListItem `json:"buckets"`
	Total   int              `json:"total"`
}

// BucketResponse is the typed response for a single bucket.
type BucketResponse struct {
	Status string  `json:"status"`
	Bucket *Bucket `json:"bucket"`
}

// --- Object DTOs ---

// ObjectListItem is a typed entry in the object list.
type ObjectListItem struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	LastModified time.Time         `json:"last_modified"`
}

// ObjectListResponse is the typed response for listing objects in a bucket.
type ObjectListResponse struct {
	Status  string           `json:"status"`
	Objects []ObjectListItem `json:"objects"`
	Total   int              `json:"total"`
}

// ObjectResponse is the typed response for a single object.
type ObjectResponse struct {
	Status string  `json:"status"`
	Object *Object `json:"object"`
}

// ObjectUploadResponse is the typed response for uploading an object.
type ObjectUploadResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Bucket  string `json:"bucket"`
	Key     string `json:"key"`
	Size    int64  `json:"size"`
	ETag    string `json:"etag,omitempty"`
}

// --- Policy DTOs ---

// AccessPolicyListItem is a typed entry in the policy list.
type AccessPolicyListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Bucket    string    `json:"bucket"`
	Principal string    `json:"principal"`
	Actions   []string  `json:"actions"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AccessPolicyListResponse is the typed response for listing access policies.
type AccessPolicyListResponse struct {
	Status   string               `json:"status"`
	Policies []AccessPolicyListItem `json:"policies"`
	Total    int                  `json:"total"`
}

// AccessPolicyResponse is the typed response for a single policy.
type AccessPolicyResponse struct {
	Status string        `json:"status"`
	Policy *AccessPolicy `json:"policy"`
}

// --- Metrics DTOs ---

// StorageMetricsResponse is the typed response for GET /api/v1/storage/metrics.
type StorageMetricsResponse struct {
	Status  string          `json:"status"`
	Metrics MetricsSnapshot `json:"metrics"`
}

// MetricsSnapshot is a point-in-time storage metrics snapshot.
type MetricsSnapshot struct {
	TotalBuckets   int   `json:"total_buckets"`
	TotalObjects   int64 `json:"total_objects"`
	TotalSizeBytes int64 `json:"total_size_bytes"`
	TotalUploads   int64 `json:"total_uploads"`
	TotalDownloads int64 `json:"total_downloads"`
	TotalDeletes   int64 `json:"total_deletes"`
	UptimeSeconds  int64 `json:"uptime_seconds"`
}

// --- Audit DTOs ---

// AuditListResponse is the typed response for listing audit events.
type AuditListResponse struct {
	Status string          `json:"status"`
	Events []AuditEvent    `json:"events"`
	Total  int             `json:"total"`
}

// AuditEvent represents a single audit event for the API layer.
type AuditEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	EventType  string    `json:"event_type"`
	BucketName string    `json:"bucket_name,omitempty"`
	ObjectKey  string    `json:"object_key,omitempty"`
	Principal  string    `json:"principal,omitempty"`
	SourceIP   string    `json:"source_ip,omitempty"`
	Message    string    `json:"message"`
}

// --- Stats DTOs ---

// BucketStatsResponse is the typed response for GET /api/v1/storage/:bucket/stats.
type BucketStatsResponse struct {
	Status      string `json:"status"`
	Bucket      string `json:"bucket"`
	ObjectCount int    `json:"object_count"`
	TotalSize   int64  `json:"total_size"`
}

// --- Generic DTOs ---

// MessageResponse is a generic ack/error response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
