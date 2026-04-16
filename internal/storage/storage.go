package storage

import (
	"context"
	"log"
	"os"

	"example.com/axiomnizam/internal/storage/admin"
	"example.com/axiomnizam/internal/storage/controller"
	"example.com/axiomnizam/internal/storage/events"
	storageMetrics "example.com/axiomnizam/internal/storage/metrics"
	"example.com/axiomnizam/internal/storage/policy"
	"example.com/axiomnizam/internal/storage/s3client"
	"example.com/axiomnizam/internal/storage/store"
	"example.com/axiomnizam/internal/storage/tenant"
	"github.com/gin-gonic/gin"
)

// System holds the fully initialised object storage system and exposes
// the router registration. Follows the IAM System struct pattern.
type System struct {
	Client     *s3client.Client
	Store      *store.BucketStore
	Controller *controller.BucketController
	Tenant     *tenant.Manager
	Policy     *policy.Controller
	Handler    *admin.Handler
	Metrics    *storageMetrics.Collector
	AuditLog   *events.AuditLog
	Config     Config
}

// Config holds configuration for connecting to an S3-compatible storage backend.
type Config struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	UseSSL          bool   `json:"useSsl"`
	Region          string `json:"region"`
	BucketPrefix    string `json:"bucketPrefix"` // e.g., "axiom-"
}

// DefaultConfig returns configuration populated from environment variables
// with sensible defaults for a local MinIO instance.
func DefaultConfig() Config {
	return Config{
		Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:          getEnv("MINIO_USE_SSL", "false") == "true",
		Region:          getEnv("MINIO_REGION", "us-east-1"),
		BucketPrefix:    getEnv("MINIO_BUCKET_PREFIX", "axiom-"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// NewSystem initialises the complete object storage system.
// Creates the native S3 HTTP client, in-memory store, reconciliation controller,
// tenant manager, policy controller, and admin handler.
func NewSystem(cfg Config) (*System, error) {
	client, err := s3client.NewClient(cfg.Endpoint, cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Region, cfg.UseSSL)
	if err != nil {
		return nil, err
	}

	bucketStore := store.NewBucketStore()
	tenantMgr := tenant.NewManager(cfg.BucketPrefix, bucketStore)
	bucketCtrl := controller.NewBucketController(bucketStore, client, cfg.Endpoint)
	policyCtrl := policy.NewController()
	metricsCollector := storageMetrics.NewCollector()
	auditLog := events.NewAuditLog(10000)
	handler := admin.NewHandler(bucketStore, client, tenantMgr, bucketCtrl, policyCtrl, metricsCollector, auditLog, cfg.Endpoint)

	return &System{
		Client:     client,
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
	log.Println("✅ Storage: module started")
}

// Stop gracefully shuts down the storage module.
func (s *System) Stop() {
	s.Controller.Stop()
	log.Println("✅ Storage: module stopped")
}
