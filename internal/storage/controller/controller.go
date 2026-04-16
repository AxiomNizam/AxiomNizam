package controller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"example.com/axiomnizam/internal/storage/models"
	"example.com/axiomnizam/internal/storage/store"
)

// BucketController implements the reconciliation loop for Bucket resources.
// Follows the Kubernetes controller pattern: observe → diff → act.
type BucketController struct {
	store    *store.BucketStore
	client   models.Backend
	endpoint string

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewBucketController creates a new controller that reconciles BucketResources
// against the storage backend.
func NewBucketController(s *store.BucketStore, client models.Backend, endpoint string) *BucketController {
	return &BucketController{
		store:    s,
		client:   client,
		endpoint: endpoint,
	}
}

// Start begins the reconciliation loop. It is safe to call multiple times.
func (bc *BucketController) Start(ctx context.Context) {
	bc.mu.Lock()
	if bc.running {
		bc.mu.Unlock()
		return
	}
	bc.running = true
	bc.stopCh = make(chan struct{})
	bc.mu.Unlock()

	log.Println("✅ Storage: BucketController started")
	go bc.run(ctx)
}

// Stop halts the reconciliation loop.
func (bc *BucketController) Stop() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if !bc.running {
		return
	}
	close(bc.stopCh)
	bc.running = false
	log.Println("✅ Storage: BucketController stopped")
}

func (bc *BucketController) run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial reconciliation
	bc.reconcileAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-bc.stopCh:
			return
		case <-ticker.C:
			bc.reconcileAll(ctx)
		}
	}
}

func (bc *BucketController) reconcileAll(ctx context.Context) {
	buckets := bc.store.ListAll()
	for _, b := range buckets {
		if err := bc.Reconcile(ctx, b); err != nil {
			log.Printf("⚠️  Storage: reconcile error for %s/%s: %v", b.Metadata.TenantID, b.Metadata.Name, err)
		}
	}
}

// Reconcile performs a single reconciliation pass for a bucket resource.
func (bc *BucketController) Reconcile(ctx context.Context, bucket *models.BucketResource) error {
	tenantBucket := bucket.Spec.Name
	log.Printf("✅ Storage: reconciling bucket %s/%s (phase=%s)", bucket.Metadata.TenantID, tenantBucket, bucket.Status.Phase)

	// Step 1: Ensure the bucket exists in the storage backend.
	exists, err := bc.client.BucketExists(ctx, tenantBucket)
	if err != nil {
		bc.setCondition(bucket, "BackendReachable", "False", "PingFailed", err.Error())
		bc.setPhase(bucket, models.BucketPhaseError)
		return fmt.Errorf("bucket exists check: %w", err)
	}

	if !exists {
		if err := bc.client.CreateBucket(ctx, tenantBucket); err != nil {
			bc.setCondition(bucket, "BucketCreated", "False", "CreateFailed", err.Error())
			bc.setPhase(bucket, models.BucketPhaseError)
			return fmt.Errorf("create bucket: %w", err)
		}
		bc.setCondition(bucket, "BucketCreated", "True", "Created", "Bucket created in storage backend")
	} else {
		bc.setCondition(bucket, "BucketCreated", "True", "AlreadyExists", "Bucket already exists")
	}

	// Step 2: Reconcile versioning.
	if bucket.Spec.Versioning == models.VersioningEnabled {
		if err := bc.client.SetBucketVersioning(ctx, tenantBucket, true); err != nil {
			bc.setCondition(bucket, "VersioningConfigured", "False", "VersioningFailed", err.Error())
			bc.setPhase(bucket, models.BucketPhaseError)
			return fmt.Errorf("set versioning: %w", err)
		}
		bc.setCondition(bucket, "VersioningConfigured", "True", "Enabled", "Versioning enabled")
	} else if bucket.Spec.Versioning == models.VersioningDisabled {
		if err := bc.client.SetBucketVersioning(ctx, tenantBucket, false); err != nil {
			log.Printf("⚠️  Storage: could not suspend versioning on %s: %v", tenantBucket, err)
		}
		bc.setCondition(bucket, "VersioningConfigured", "True", "Disabled", "Versioning suspended")
	}

	// Step 3: Apply lifecycle policy if defined.
	if len(bucket.Spec.LifecyclePolicy) > 0 {
		if err := bc.client.SetBucketLifecycle(ctx, tenantBucket, bucket.Spec.LifecyclePolicy); err != nil {
			bc.setCondition(bucket, "LifecycleApplied", "False", "LifecycleFailed", err.Error())
			log.Printf("⚠️  Storage: lifecycle policy apply failed on %s: %v", tenantBucket, err)
		} else {
			bc.setCondition(bucket, "LifecycleApplied", "True", "Applied", fmt.Sprintf("%d rules applied", len(bucket.Spec.LifecyclePolicy)))
		}
	}

	// Step 4: Gather stats.
	objects, err := bc.client.ListObjects(ctx, tenantBucket, "")
	if err == nil {
		var totalSize int64
		for _, o := range objects {
			totalSize += o.Size
		}
		bucket.Status.ObjectCount = int64(len(objects))
		bucket.Status.TotalSize = totalSize
	}

	// Step 5: Update status to Ready.
	bucket.Status.Endpoint = bc.endpoint
	bucket.Status.ObservedGeneration = bucket.Generation
	bc.setPhase(bucket, models.BucketPhaseReady)
	bc.setCondition(bucket, "Ready", "True", "Reconciled", "Bucket fully reconciled")

	return nil
}

// ReconcileOne triggers immediate reconciliation for a single bucket.
func (bc *BucketController) ReconcileOne(ctx context.Context, tenantID, name string) error {
	bucket, err := bc.store.Get(tenantID, name)
	if err != nil {
		return err
	}
	return bc.Reconcile(ctx, bucket)
}

func (bc *BucketController) setPhase(bucket *models.BucketResource, phase models.BucketPhase) {
	bucket.Status.Phase = phase
	_ = bc.store.UpdateStatus(bucket.Metadata.TenantID, bucket.Metadata.Name, bucket.Status)
}

func (bc *BucketController) setCondition(bucket *models.BucketResource, condType, status, reason, message string) {
	now := time.Now().UTC()
	for i, c := range bucket.Status.Conditions {
		if c.Type == condType {
			bucket.Status.Conditions[i] = models.Condition{
				Type:               condType,
				Status:             status,
				Reason:             reason,
				Message:            message,
				LastTransitionTime: now,
			}
			return
		}
	}
	bucket.Status.Conditions = append(bucket.Status.Conditions, models.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	})
}
