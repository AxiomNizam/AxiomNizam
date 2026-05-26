package testutil

import (
	"time"

	"example.com/axiomnizam/internal/storage/models"
)

// TestTenantID is a fixed tenant ID for testing.
const TestTenantID = "test-tenant"

// TestBucketName is a fixed bucket name for testing.
const TestBucketName = "test-bucket"

// TestObjectKey is a fixed object key for testing.
const TestObjectKey = "test-object.txt"

// NewTestBucket creates a test BucketResource with sensible defaults.
func NewTestBucket() *models.BucketResource {
	now := time.Now().UTC()
	return &models.BucketResource{
		Spec: models.BucketSpec{
			TenantID:    TestTenantID,
			Name:        TestBucketName,
			Region:      "us-east-1",
			Versioning:  false,
			Encryption:  models.EncryptionConfig{Enabled: false},
		},
		Status: models.BucketResourceStatus{
			Phase:      models.BucketPhaseReady,
			Endpoint:   "http://localhost:8000",
			ObjectCount: 0,
			TotalSize:   0,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewTestBucketWithObjects creates a test bucket with object count set.
func NewTestBucketWithObjects(count int, size int64) *models.BucketResource {
	b := NewTestBucket()
	b.Status.ObjectCount = int64(count)
	b.Status.TotalSize = size
	return b
}
