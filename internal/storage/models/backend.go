package models

import (
	"context"
	"io"
	"time"
)

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

	// ---- Object operations ----

	PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) error
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

	// ---- Utility ----

	GetBucketSize(ctx context.Context, bucket string) (int64, int64, error)
}
