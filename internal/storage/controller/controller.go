package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/storage/models"
	"example.com/axiomnizam/internal/storage/store"
)

const defaultResyncInterval = 7 * time.Minute

type reconcileRequest struct {
	tenantID string
	name     string
	reason   string
}

func (r reconcileRequest) key() string {
	return r.tenantID + "/" + r.name
}

// BucketController implements the reconciliation loop for Bucket resources.
// Follows the Kubernetes controller pattern: observe → diff → act.
type BucketController struct {
	store    *store.BucketStore
	client   models.Backend
	endpoint string

	mu             sync.Mutex
	running        bool
	stopCh         chan struct{}
	workCh         chan reconcileRequest
	watchID        int
	watchCh        <-chan store.BucketEvent
	pending        map[string]struct{}
	resyncInterval time.Duration
	debugEnabled   bool
}

// NewBucketController creates a new controller that reconciles BucketResources
// against the storage backend.
func NewBucketController(s *store.BucketStore, client models.Backend, endpoint string) *BucketController {
	return &BucketController{
		store:          s,
		client:         client,
		endpoint:       endpoint,
		pending:        make(map[string]struct{}),
		resyncInterval: resyncIntervalFromEnv(),
		debugEnabled:   debugEnabledFromEnv(),
	}
}

// Start begins the reconciliation loop. It is safe to call multiple times.
func (bc *BucketController) Start(ctx context.Context) {
	bc.mu.Lock()
	if bc.running {
		bc.mu.Unlock()
		return
	}
	watchID, watchCh := bc.store.Subscribe(256)
	bc.running = true
	bc.stopCh = make(chan struct{})
	bc.workCh = make(chan reconcileRequest, 512)
	bc.pending = make(map[string]struct{})
	bc.watchID = watchID
	bc.watchCh = watchCh
	bc.mu.Unlock()

	log.Printf("✅ Storage: BucketController started (event-driven, resync=%s)", bc.resyncInterval)
	go bc.worker(ctx)
	go bc.run(ctx)
}

// Stop halts the reconciliation loop.
func (bc *BucketController) Stop() {
	bc.mu.Lock()
	if !bc.running {
		bc.mu.Unlock()
		return
	}
	close(bc.stopCh)
	if bc.watchID != 0 {
		bc.store.Unsubscribe(bc.watchID)
		bc.watchID = 0
		bc.watchCh = nil
	}
	bc.running = false
	bc.pending = make(map[string]struct{})
	bc.mu.Unlock()
	log.Println("✅ Storage: BucketController stopped")
}

func (bc *BucketController) run(ctx context.Context) {
	ticker := time.NewTicker(bc.resyncInterval)
	defer ticker.Stop()

	// Initial reconciliation enqueue.
	bc.enqueueAll("startup")

	for {
		select {
		case <-ctx.Done():
			return
		case <-bc.stopCh:
			return
		case ev, ok := <-bc.watchCh:
			if !ok {
				return
			}
			switch ev.Type {
			case store.BucketEventCreate, store.BucketEventUpdate:
				bc.enqueue(ev.TenantID, ev.Name, string(ev.Type))
			case store.BucketEventDelete:
				bc.debugf("skip enqueue for deleted bucket %s/%s", ev.TenantID, ev.Name)
			}
		case <-ticker.C:
			bc.enqueueAll("resync")
		}
	}
}

func (bc *BucketController) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-bc.stopCh:
			return
		case req := <-bc.workCh:
			bc.dequeue(req.key())
			if err := bc.reconcileByKey(ctx, req); err != nil {
				log.Printf("⚠️  Storage: reconcile error for %s/%s: %v", req.tenantID, req.name, err)
			}
		}
	}
}

func (bc *BucketController) reconcileByKey(ctx context.Context, req reconcileRequest) error {
	bucket, err := bc.store.Get(req.tenantID, req.name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			bc.debugf("bucket removed before reconcile: %s/%s", req.tenantID, req.name)
			return nil
		}
		return err
	}
	return bc.Reconcile(ctx, bucket)
}

func (bc *BucketController) enqueueAll(reason string) {
	buckets := bc.store.ListAll()
	for _, b := range buckets {
		bc.enqueue(b.Metadata.TenantID, b.Metadata.Name, reason)
	}
}

func (bc *BucketController) enqueue(tenantID, name, reason string) {
	if tenantID == "" || name == "" {
		return
	}

	req := reconcileRequest{tenantID: tenantID, name: name, reason: reason}
	key := req.key()

	bc.mu.Lock()
	if !bc.running {
		bc.mu.Unlock()
		return
	}
	if _, exists := bc.pending[key]; exists {
		bc.mu.Unlock()
		bc.debugf("skip duplicate enqueue for %s (reason=%s)", key, reason)
		return
	}
	bc.pending[key] = struct{}{}
	workCh := bc.workCh
	stopCh := bc.stopCh
	bc.mu.Unlock()

	select {
	case workCh <- req:
	case <-stopCh:
		bc.dequeue(key)
	}
}

func (bc *BucketController) dequeue(key string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if bc.pending == nil {
		return
	}
	delete(bc.pending, key)
}

// Reconcile performs a single reconciliation pass for a bucket resource.
func (bc *BucketController) Reconcile(ctx context.Context, bucket *models.BucketResource) error {
	if bucket == nil {
		return fmt.Errorf("bucket is nil")
	}

	specChanged := bucket.Status.ObservedGeneration != bucket.Generation
	if bucket.Status.Phase == models.BucketPhaseReady && !specChanged {
		bc.debugf("skip reconcile for %s/%s (ready and unchanged generation=%d)", bucket.Metadata.TenantID, bucket.Metadata.Name, bucket.Generation)
		return nil
	}

	if bucket.Status.ObservedGeneration == 0 {
		log.Printf("Storage: reconciling new resource %s/%s", bucket.Metadata.TenantID, bucket.Metadata.Name)
	} else if specChanged {
		log.Printf("Storage: reconciling spec change for %s/%s (observedGeneration=%d generation=%d)", bucket.Metadata.TenantID, bucket.Metadata.Name, bucket.Status.ObservedGeneration, bucket.Generation)
	}

	tenantBucket := bucket.Spec.Name

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
		log.Printf("Storage: created backend bucket for %s/%s", bucket.Metadata.TenantID, bucket.Metadata.Name)
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

	if bc.isRunning() {
		bc.enqueue(tenantID, name, "manual")
		return nil
	}

	return bc.Reconcile(ctx, bucket)
}

func (bc *BucketController) isRunning() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.running
}

func (bc *BucketController) setPhase(bucket *models.BucketResource, phase models.BucketPhase) {
	bucket.Status.Phase = phase
	_ = bc.store.UpdateStatus(bucket.Metadata.TenantID, bucket.Metadata.Name, bucket.Status)
}

func (bc *BucketController) setCondition(bucket *models.BucketResource, condType, status, reason, message string) {
	now := time.Now().UTC()
	for i, c := range bucket.Status.Conditions {
		if c.Type == condType {
			if c.Status == status && c.Reason == reason && c.Message == message {
				return
			}
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

func (bc *BucketController) debugf(format string, args ...interface{}) {
	if !bc.debugEnabled {
		return
	}
	log.Printf("Storage[debug]: "+format, args...)
}

func debugEnabledFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_CONTROLLER_DEBUG")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func resyncIntervalFromEnv() time.Duration {
	raw := strings.TrimSpace(os.Getenv("STORAGE_CONTROLLER_RESYNC_INTERVAL"))
	if raw == "" {
		return defaultResyncInterval
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("⚠️  Storage: invalid STORAGE_CONTROLLER_RESYNC_INTERVAL=%q, using default %s", raw, defaultResyncInterval)
		return defaultResyncInterval
	}
	if parsed < 5*time.Minute {
		return 5 * time.Minute
	}
	if parsed > 10*time.Minute {
		return 10 * time.Minute
	}
	return parsed
}
