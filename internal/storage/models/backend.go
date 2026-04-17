package models

import (
	"context"
	"io"
	"time"
)

// PutObjectOptions provides optional settings for PutObject.
type PutObjectOptions struct {
	ContentType  string
	UserMetadata map[string]string
	Encryption   bool // apply server-side encryption
	LegalHold    bool // place legal hold on object
	RetainUntil  *time.Time
	StorageClass string
}

// Backend is the interface every object-storage backend must satisfy.
// Both the native (filesystem) backend and the S3-compatible HTTP client
// implement this interface so they can be swapped transparently.
type Backend interface {
	// Ping checks whether the backend is reachable.
	Ping(ctx context.Context) error
	// Endpoint returns a human-readable address of the backend.
	Endpoint() string

	// ---- Bucket operations ----

	CreateBucket(ctx context.Context, name string) error
	DeleteBucket(ctx context.Context, name string) error
	BucketExists(ctx context.Context, name string) (bool, error)
	ListBuckets(ctx context.Context) ([]string, error)

	// ---- Versioning ----

	SetBucketVersioning(ctx context.Context, bucket string, enabled bool) error
	GetBucketVersioning(ctx context.Context, bucket string) (bool, error)

	// ---- Lifecycle ----

	SetBucketLifecycle(ctx context.Context, bucket string, rules []LifecycleRule) error
	GetBucketLifecycle(ctx context.Context, bucket string) ([]LifecycleRule, error)

	// ---- Encryption ----

	SetBucketEncryption(ctx context.Context, bucket string, cfg BucketEncryption) error
	GetBucketEncryption(ctx context.Context, bucket string) (*BucketEncryption, error)
	DeleteBucketEncryption(ctx context.Context, bucket string) error

	// ---- Object Lock / Retention ----

	SetObjectLockConfig(ctx context.Context, bucket string, cfg ObjectLockConfig) error
	GetObjectLockConfig(ctx context.Context, bucket string) (*ObjectLockConfig, error)
	PutObjectRetention(ctx context.Context, bucket, key string, until time.Time, mode string) error
	GetObjectRetention(ctx context.Context, bucket, key string) (*time.Time, string, error)
	PutObjectLegalHold(ctx context.Context, bucket, key string, hold bool) error
	GetObjectLegalHold(ctx context.Context, bucket, key string) (bool, error)

	// ---- CORS ----

	SetBucketCORS(ctx context.Context, bucket string, rules []CORSRule) error
	GetBucketCORS(ctx context.Context, bucket string) ([]CORSRule, error)
	DeleteBucketCORS(ctx context.Context, bucket string) error

	// ---- Notifications ----

	SetBucketNotification(ctx context.Context, bucket string, cfg BucketNotificationConfig) error
	GetBucketNotification(ctx context.Context, bucket string) (*BucketNotificationConfig, error)

	// ---- Bucket Policy ----

	SetBucketPolicy(ctx context.Context, bucket string, policy S3BucketPolicy) error
	GetBucketPolicy(ctx context.Context, bucket string) (*S3BucketPolicy, error)
	DeleteBucketPolicy(ctx context.Context, bucket string) error

	// ---- Object operations ----

	PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) error
	PutObjectWithOptions(ctx context.Context, bucket, key string, data io.Reader, size int64, opts PutObjectOptions) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)
	StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error)

	// ---- Batch / copy ----

	MultiDeleteObjects(ctx context.Context, bucket string, keys []string) (int, []string, error)
	CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error

	// ---- Pre-signed URLs ----

	PresignGetObject(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
	PresignPutObject(ctx context.Context, bucket, key string, expires time.Duration) (string, error)

	// ---- Tagging ----

	GetBucketTagging(ctx context.Context, bucket string) ([]BucketTag, error)
	PutBucketTagging(ctx context.Context, bucket string, tags []BucketTag) error
	DeleteBucketTagging(ctx context.Context, bucket string) error

	// ---- Object Metadata ----

	GetObjectMetadata(ctx context.Context, bucket, key string) (map[string]string, error)
	PutObjectMetadata(ctx context.Context, bucket, key string, meta map[string]string) error

	// ---- Utility ----

	GetBucketSize(ctx context.Context, bucket string) (int64, int64, error)
	GetBucketQuota(ctx context.Context, bucket string) (*QuotaInfo, error)
}
