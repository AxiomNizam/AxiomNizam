package storage

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	iamMiddleware "example.com/axiomnizam/internal/iam/middleware"
	iamStorage "example.com/axiomnizam/internal/iam/storage"
	"example.com/axiomnizam/internal/iam/token"
	"example.com/axiomnizam/internal/storage/access"
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
	clientv3 "go.etcd.io/etcd/client/v3"
)

// System holds the fully initialised object storage system and exposes
// the router registration. Follows the IAM System struct pattern.
type System struct {
	Backend    models.Backend
	Store      *store.BucketStore
	Controller *controller.BucketController
	Tenant     *tenant.Manager
	Policy     *policy.Controller
	Access     *access.Controller
	Handler    *admin.Handler
	Metrics    *storageMetrics.Collector
	AuditLog   *events.AuditLog
	Config     Config

	// IAM references (nil when IAM is not available).
	iamIssuer       *token.Issuer
	iamRevokedStore *iamStorage.EtcdRevokedTokenStore
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
// IAM issuer and revokedStore are optional; when provided, routes are protected
// by IAM JWT middleware so that the access controller can extract user identity.
func NewSystem(cfg Config, issuer *token.Issuer, revokedStore *iamStorage.EtcdRevokedTokenStore, etcdClient *clientv3.Client) (*System, error) {
	backend, err := native.New(cfg.DataDir, cfg.PresignSecret)
	if err != nil {
		return nil, err
	}

	endpoint := backend.Endpoint()
	bucketStore := store.NewBucketStore()
	bucketStore.ConfigurePersistence(etcdClient)
	tenantMgr := tenant.NewManager(cfg.BucketPrefix, bucketStore)
	bucketCtrl := controller.NewBucketController(bucketStore, backend, endpoint)
	policyCtrl := policy.NewController()
	metricsCollector := storageMetrics.NewCollector()
	auditLog := events.NewAuditLog(10000)
	accessCtrl := access.NewController(auditLog)
	accessCtrl.ConfigurePersistence(etcdClient)
	handler := admin.NewHandler(bucketStore, backend, tenantMgr, bucketCtrl, accessCtrl, metricsCollector, auditLog, endpoint)

	return &System{
		Backend:         backend,
		Store:           bucketStore,
		Controller:      bucketCtrl,
		Tenant:          tenantMgr,
		Policy:          policyCtrl,
		Access:          accessCtrl,
		Handler:         handler,
		Metrics:         metricsCollector,
		AuditLog:        auditLog,
		Config:          cfg,
		iamIssuer:       issuer,
		iamRevokedStore: revokedStore,
	}, nil
}

// RegisterRoutes mounts all object storage endpoints on the provided router group.
// When an IAM issuer is configured, the storage route group is wrapped with JWTAuth
// middleware so that downstream handlers can extract iam_claims for access control.
func (s *System) RegisterRoutes(rg *gin.RouterGroup) {
	if s.iamIssuer != nil {
		jwtAuth := iamMiddleware.JWTAuth(s.iamIssuer, s.iamRevokedStore)
		rg.Use(func(c *gin.Context) {
			if isPresignedObjectRequest(c) {
				c.Next()
				return
			}
			jwtAuth(c)
		})
		log.Println("✅ Storage: IAM JWT middleware attached to storage routes")
	}
	s.Handler.RegisterRoutes(rg)
}

func isPresignedObjectRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}
	if strings.TrimSpace(c.Query("X-Axiom-Token")) == "" {
		return false
	}
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodPut {
		return false
	}
	path := c.Request.URL.Path
	if !strings.Contains(path, "/storage/buckets/") {
		return false
	}
	return strings.Contains(path, "/objects/")
}

// Start begins the reconciliation controller.
func (s *System) Start(ctx context.Context) {
	s.Controller.Start(ctx)
	log.Println("✅ Storage: module started (native backend, IAM-integrated)")
}

// Stop gracefully shuts down the storage module.
func (s *System) Stop() {
	s.Controller.Stop()
	log.Println("✅ Storage: module stopped")
}
