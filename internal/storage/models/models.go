package models

import (
	"time"
)

// BucketPhase represents the lifecycle phase of a bucket.
type BucketPhase string

const (
	BucketPhasePending BucketPhase = "Pending"
	BucketPhaseReady   BucketPhase = "Ready"
	BucketPhaseError   BucketPhase = "Error"
)

// VersioningStatus represents the versioning state.
type VersioningStatus string

const (
	VersioningEnabled  VersioningStatus = "Enabled"
	VersioningDisabled VersioningStatus = "Disabled"
)

// Condition represents a status condition on a resource.
type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"` // "True", "False", "Unknown"
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

// LifecycleRule defines an object lifecycle policy.
type LifecycleRule struct {
	ID                       string `json:"id"`
	Prefix                   string `json:"prefix"`
	ExpirationDays           int    `json:"expirationDays,omitempty"`
	NoncurrentExpirationDays int    `json:"noncurrentExpirationDays,omitempty"`
	TransitionDays           int    `json:"transitionDays,omitempty"`
	TransitionStorageClass   string `json:"transitionStorageClass,omitempty"`
}

// BucketMetadata holds identifying information for a bucket.
type BucketMetadata struct {
	Name      string            `json:"name"`
	TenantID  string            `json:"tenantId"`
	UID       string            `json:"uid,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// BucketSpec defines the desired state of a bucket.
type BucketSpec struct {
	Name            string           `json:"name"`
	Versioning      VersioningStatus `json:"versioning"`
	LifecyclePolicy []LifecycleRule  `json:"lifecyclePolicy,omitempty"`
	Region          string           `json:"region,omitempty"`
	Quota           int64            `json:"quota,omitempty"` // bytes, 0 = unlimited
}

// BucketStatus represents the observed state of a bucket.
type BucketStatus struct {
	Phase              BucketPhase `json:"phase"`
	Endpoint           string      `json:"endpoint,omitempty"`
	Conditions         []Condition `json:"conditions,omitempty"`
	ObjectCount        int64       `json:"objectCount"`
	TotalSize          int64       `json:"totalSize"` // bytes
	ObservedGeneration int64       `json:"observedGeneration"`
}

// BucketResource is the CRD-style resource for object storage buckets.
type BucketResource struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   BucketMetadata `json:"metadata"`
	Spec       BucketSpec     `json:"spec"`
	Status     BucketStatus   `json:"status"`
	Generation int64          `json:"generation"`
}

// ObjectInfo contains metadata about a stored object.
type ObjectInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"contentType"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"lastModified"`
	VersionID    string    `json:"versionId,omitempty"`
}

// StorageStats holds aggregate storage statistics.
type StorageStats struct {
	TotalBuckets   int   `json:"totalBuckets"`
	TotalObjects   int64 `json:"totalObjects"`
	TotalSizeBytes int64 `json:"totalSizeBytes"`
	TenantCount    int   `json:"tenantCount"`
}

// PreSignedURLResponse contains the generated pre-signed URL.
type PreSignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// S3Action represents a permitted S3 operation.
type S3Action string

const (
	S3ActionGetObject    S3Action = "s3:GetObject"
	S3ActionPutObject    S3Action = "s3:PutObject"
	S3ActionDeleteObject S3Action = "s3:DeleteObject"
	S3ActionListBucket   S3Action = "s3:ListBucket"
	S3ActionGetBucketLoc S3Action = "s3:GetBucketLocation"
	S3ActionAll          S3Action = "s3:*"
)

// S3PolicyStatement is a single statement in an S3 bucket policy.
type S3PolicyStatement struct {
	Sid       string     `json:"Sid"`
	Effect    string     `json:"Effect"` // "Allow" or "Deny"
	Principal string     `json:"Principal"`
	Action    []S3Action `json:"Action"`
	Resource  []string   `json:"Resource"`
}

// S3BucketPolicy is a complete S3 bucket policy document.
type S3BucketPolicy struct {
	Version   string              `json:"Version"`
	Statement []S3PolicyStatement `json:"Statement"`
}

// StorageRole defines a role that maps to S3 permissions.
type StorageRole string

const (
	StorageRoleAdmin    StorageRole = "storage-admin"    // full access
	StorageRoleWriter   StorageRole = "storage-writer"   // read + write
	StorageRoleReader   StorageRole = "storage-reader"   // read only
	StorageRoleUploader StorageRole = "storage-uploader" // write only
)

// TenantPolicy tracks the IAM-to-storage policy mapping for a tenant.
type TenantPolicy struct {
	TenantID   string      `json:"tenantId"`
	UserID     string      `json:"userId"`
	Role       StorageRole `json:"role"`
	BucketName string      `json:"bucketName"`
	Prefix     string      `json:"prefix,omitempty"` // optional path restriction
	PolicyJSON string      `json:"policyJson"`
}

// ===== Enhanced Features =====

// BucketTag is a key-value tag on a bucket.
type BucketTag struct {
	Key   string `json:"key" xml:"Key"`
	Value string `json:"value" xml:"Value"`
}

// BucketEncryption defines server-side encryption settings.
type BucketEncryption struct {
	Enabled   bool   `json:"enabled"`
	Algorithm string `json:"algorithm"` // "AES256" or "aws:kms"
	KMSKeyID  string `json:"kmsKeyId,omitempty"`
}

// CORSRule defines a CORS rule for a bucket.
type CORSRule struct {
	AllowedOrigins []string `json:"allowedOrigins" xml:"AllowedOrigin"`
	AllowedMethods []string `json:"allowedMethods" xml:"AllowedMethod"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty" xml:"AllowedHeader"`
	ExposeHeaders  []string `json:"exposeHeaders,omitempty" xml:"ExposeHeader"`
	MaxAgeSeconds  int      `json:"maxAgeSeconds,omitempty" xml:"MaxAgeSeconds"`
}

// ReplicationRule defines bucket-to-bucket replication.
type ReplicationRule struct {
	ID                string `json:"id"`
	DestinationBucket string `json:"destinationBucket"`
	Prefix            string `json:"prefix,omitempty"`
	Enabled           bool   `json:"enabled"`
}

// StorageEvent represents a storage operation event for audit.
type StorageEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "bucket.created", "object.uploaded", "object.deleted", etc.
	TenantID  string    `json:"tenantId"`
	UserID    string    `json:"userId,omitempty"`
	Bucket    string    `json:"bucket"`
	Key       string    `json:"key,omitempty"`
	Size      int64     `json:"size,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// BucketMetrics holds per-bucket performance metrics.
type BucketMetrics struct {
	BucketName     string    `json:"bucketName"`
	TenantID       string    `json:"tenantId"`
	RequestCount   int64     `json:"requestCount"`
	GetRequests    int64     `json:"getRequests"`
	PutRequests    int64     `json:"putRequests"`
	DeleteRequests int64     `json:"deleteRequests"`
	BytesIn        int64     `json:"bytesIn"`
	BytesOut       int64     `json:"bytesOut"`
	ErrorCount     int64     `json:"errorCount"`
	AvgLatencyMs   float64   `json:"avgLatencyMs"`
	LastAccessed   time.Time `json:"lastAccessed"`
	CollectedAt    time.Time `json:"collectedAt"`
}

// SystemMetrics holds aggregate storage system metrics.
type SystemMetrics struct {
	Uptime         string  `json:"uptime"`
	TotalBuckets   int     `json:"totalBuckets"`
	TotalObjects   int64   `json:"totalObjects"`
	TotalSizeBytes int64   `json:"totalSizeBytes"`
	TenantCount    int     `json:"tenantCount"`
	TotalRequests  int64   `json:"totalRequests"`
	TotalBytesIn   int64   `json:"totalBytesIn"`
	TotalBytesOut  int64   `json:"totalBytesOut"`
	TotalErrors    int64   `json:"totalErrors"`
	ActivePolicies int     `json:"activePolicies"`
	BackendHealthy bool    `json:"backendHealthy"`
	AvgLatencyMs   float64 `json:"avgLatencyMs"`
}

// MultiDeleteRequest represents a batch delete request.
type MultiDeleteRequest struct {
	Objects []DeleteObjectEntry `json:"objects"`
}

// DeleteObjectEntry is a single object key in a batch delete.
type DeleteObjectEntry struct {
	Key       string `json:"key" xml:"Key"`
	VersionID string `json:"versionId,omitempty" xml:"VersionId"`
}

// CopyObjectRequest represents a server-side copy operation.
type CopyObjectRequest struct {
	SourceBucket string `json:"sourceBucket"`
	SourceKey    string `json:"sourceKey"`
	DestBucket   string `json:"destBucket"`
	DestKey      string `json:"destKey"`
}

// ObjectVersion represents a specific version of an object.
type ObjectVersion struct {
	Key          string    `json:"key"`
	VersionID    string    `json:"versionId"`
	IsLatest     bool      `json:"isLatest"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
}

// BucketNotification configures event notifications for a bucket.
type BucketNotification struct {
	ID     string   `json:"id"`
	Events []string `json:"events"` // e.g., "s3:ObjectCreated:*", "s3:ObjectRemoved:*"
	Prefix string   `json:"prefix,omitempty"`
	Suffix string   `json:"suffix,omitempty"`
}
