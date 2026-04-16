package storage

import (
	"context"
	"log"
	"os"

	"example.com/axiomnizam/internal/storage/admin"
	"example.com/axiomnizam/internal/storage/controller"
	"example.com/axiomnizam/internal/storage/events"
	storageMetrics "example.com/axiomnizam/internal/storage/metrics"
	"example.com/axiomnizam/internal/storage/models"
	"example.com/axiomnizam/internal/storage/native"
	"example.com/axiomnizam/internal/storage/policy"
	"example.com/axiomnizam/internal/storage/store"
	"example.com/axiomnizam/internal/storage/tenant"
	"github.com/gin-gonic/gin"
)

// System holds the fully initialised object storage system and exposes
// the router registration. Follows the IAM System struct pattern.
type System struct {
	Backend    models.Backend
	Store      *store.BucketStore
	Controller *controller.BucketController
	Tenant     *tenant.Manager
	Policy     *policy.Controller
	Handler    *admin.Handler
	Metrics    *storageMetrics.Collector
	AuditLog   *events.AuditLog
	Config     Config
}

// Config holds configuration for the native object storage backend.
type Config struct {
	DataDir       string `json:"dataDir"`       // filesystem root for object data
	BucketPrefix  string `json:"bucketPrefix"`  // e.g., "axiom-"
	PresignSecret string `json:"presignSecret"` // HMAC key for presign tokens
}

// DefaultConfig returns configuration populated from environment variables
// with sensible defaults for a local native storage backend.
func DefaultConfig() Config {
	return Config{
		DataDir:       getEnv("STORAGE_DATA_DIR", "/data/storage"),
		BucketPrefix:  getEnv("STORAGE_BUCKET_PREFIX", "axiom-"),
		PresignSecret: getEnv("STORAGE_PRESIGN_SECRET", "axiom-native-storage-default-key"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// NewSystem initialises the complete object storage system.
// Uses the built-in native filesystem backend — no external service required.
func NewSystem(cfg Config) (*System, error) {
	backend, err := native.New(cfg.DataDir, cfg.PresignSecret)
	if err != nil {
		return nil, err
	}

	endpoint := backend.Endpoint()
	bucketStore := store.NewBucketStore()
	tenantMgr := tenant.NewManager(cfg.BucketPrefix, bucketStore)
	bucketCtrl := controller.NewBucketController(bucketStore, backend, endpoint)
	policyCtrl := policy.NewController()
	metricsCollector := storageMetrics.NewCollector()
	auditLog := events.NewAuditLog(10000)
	handler := admin.NewHandler(bucketStore, backend, tenantMgr, bucketCtrl, policyCtrl, metricsCollector, auditLog, endpoint)

	return &System{
		Backend:    backend,
		Store:      bucketStore,
		Controller: bucketCtrl,
		Tenant:     tenantMgr,
		Policy:     policyCtrl,
		Handler:    handler,
		Metrics:    metricsCollector,
		AuditLog:   auditLog,
		Config:     cfg,
	}, nil
}

// RegisterRoutes mounts all object storage endpoints on the provided router group.
func (s *System) RegisterRoutes(rg *gin.RouterGroup) {
	s.Handler.RegisterRoutes(rg)
}

// Start begins the reconciliation controller.
func (s *System) Start(ctx context.Context) {
	s.Controller.Start(ctx)
	log.Println("✅ Storage: module started (native backend)")
}

// Stop gracefully shuts down the storage module.
func (s *System) Stop() {
	s.Controller.Stop()
	log.Println("✅ Storage: module stopped")
}
