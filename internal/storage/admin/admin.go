package admin

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/axiomnizam/internal/storage/access"
	"example.com/axiomnizam/internal/storage/controller"
	"example.com/axiomnizam/internal/storage/events"
	storageMetrics "example.com/axiomnizam/internal/storage/metrics"
	"example.com/axiomnizam/internal/storage/models"
	"example.com/axiomnizam/internal/storage/store"
	"example.com/axiomnizam/internal/storage/tenant"
	"github.com/gin-gonic/gin"
)

// Handler exposes the object storage API endpoints with IAM-integrated access control.
type Handler struct {
	store      *store.BucketStore
	client     models.Backend
	tenant     *tenant.Manager
	controller *controller.BucketController
	access     *access.Controller
	metrics    *storageMetrics.Collector
	audit      *events.AuditLog
	endpoint   string
}

// NewHandler creates a new storage API handler.
func NewHandler(
	s *store.BucketStore,
	client models.Backend,
	t *tenant.Manager,
	ctrl *controller.BucketController,
	ac *access.Controller,
	m *storageMetrics.Collector,
	a *events.AuditLog,
	endpoint string,
) *Handler {
	return &Handler{
		store:      s,
		client:     client,
		tenant:     t,
		controller: ctrl,
		access:     ac,
		metrics:    m,
		audit:      a,
		endpoint:   endpoint,
	}
}

// RegisterRoutes registers all object storage API routes on the given router group.
// Routes are grouped by access level:
//   - Public:        health check
//   - Authenticated: all storage operations (requires IAM JWT or access key)
//   - Admin:         system metrics, policies, access keys, shares
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	sg := rg.Group("/storage")
	{
		// Public (no auth)
		sg.GET("/health", h.Health)

		// Authenticated
		// IAM auth is applied at the system level (storage.go injects middleware).
		// Here we apply storage-level role checks per route group.
		auth := sg.Group("", h.access.RequireStorageAuth())
		{
			// Stats & Monitoring (read access)
			auth.GET("/stats", h.Stats)
			auth.GET("/events", h.ListEvents)
			auth.GET("/events/:bucket", h.ListBucketEvents)

			// Metrics (authenticated alias — frontend uses /storage/metrics)
			auth.GET("/metrics", h.SystemMetrics)
			auth.GET("/metrics/:bucket", h.BucketMetricsHandler)

			// Bucket CRUD
			auth.POST("/buckets", h.access.RequireStorageRole(models.StorageRoleWriter, models.StorageRoleAdmin), h.CreateBucket)
			auth.GET("/buckets", h.ListBuckets)
			auth.GET("/buckets/:bucket", h.GetBucket)
			auth.DELETE("/buckets/:bucket", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.DeleteBucket)

			// Bucket Tagging
			auth.GET("/buckets/:bucket/tags", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetBucketTags)
			auth.PUT("/buckets/:bucket/tags", h.access.RequireBucketAccess(models.StorageRoleWriter), h.SetBucketTags)
			auth.DELETE("/buckets/:bucket/tags", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.DeleteBucketTags)

			// Bucket Encryption
			auth.GET("/buckets/:bucket/encryption", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetBucketEncryption)
			auth.PUT("/buckets/:bucket/encryption", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.SetBucketEncryption)
			auth.DELETE("/buckets/:bucket/encryption", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.DeleteBucketEncryption)

			// Object Lock / Retention
			auth.GET("/buckets/:bucket/object-lock", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetObjectLockConfig)
			auth.PUT("/buckets/:bucket/object-lock", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.SetObjectLockConfig)
			auth.GET("/buckets/:bucket/object-retention", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetObjectRetention)
			auth.PUT("/buckets/:bucket/object-retention", h.access.RequireBucketAccess(models.StorageRoleWriter), h.SetObjectRetention)
			auth.GET("/buckets/:bucket/object-legal-hold", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetObjectLegalHold)
			auth.PUT("/buckets/:bucket/object-legal-hold", h.access.RequireBucketAccess(models.StorageRoleWriter), h.SetObjectLegalHold)

			// CORS
			auth.GET("/buckets/:bucket/cors", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetBucketCORS)
			auth.PUT("/buckets/:bucket/cors", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.SetBucketCORS)
			auth.DELETE("/buckets/:bucket/cors", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.DeleteBucketCORS)

			// Notifications
			auth.GET("/buckets/:bucket/notifications", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetBucketNotifications)
			auth.PUT("/buckets/:bucket/notifications", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.SetBucketNotifications)

			// Bucket Policy
			auth.GET("/buckets/:bucket/policy", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetBucketPolicy)
			auth.PUT("/buckets/:bucket/policy", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.SetBucketPolicy)
			auth.DELETE("/buckets/:bucket/policy", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.DeleteBucketPolicy)

			// Quota
			auth.GET("/buckets/:bucket/quota", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetBucketQuota)

			// Object operations
			auth.PUT("/buckets/:bucket/objects/*key", h.access.RequireBucketAccess(models.StorageRoleWriter), h.PutObject)
			auth.GET("/buckets/:bucket/objects/*key", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetObject)
			auth.DELETE("/buckets/:bucket/objects/*key", h.access.RequireBucketAccess(models.StorageRoleWriter), h.DeleteObject)
			auth.GET("/buckets/:bucket/objects", h.access.RequireBucketAccess(models.StorageRoleReader), h.ListObjects)
			auth.HEAD("/buckets/:bucket/objects/*key", h.access.RequireBucketAccess(models.StorageRoleReader), h.HeadObject)

			// Object metadata
			auth.GET("/buckets/:bucket/object-metadata", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetObjectMetadata)
			auth.PUT("/buckets/:bucket/object-metadata", h.access.RequireBucketAccess(models.StorageRoleWriter), h.PutObjectMetadata)

			// Batch operations
			auth.POST("/buckets/:bucket/multi-delete", h.access.RequireBucketAccess(models.StorageRoleWriter), h.MultiDeleteObjects)
			auth.POST("/copy", h.access.RequireStorageRole(models.StorageRoleWriter, models.StorageRoleAdmin), h.CopyObject)

			// Pre-signed URLs
			auth.POST("/buckets/:bucket/presign", h.access.RequireBucketAccess(models.StorageRoleReader), h.PresignURL)

			// Bucket Sharing
			auth.POST("/buckets/:bucket/shares", h.access.RequireBucketAccess(models.StorageRoleAdmin), h.CreateBucketShare)
			auth.GET("/buckets/:bucket/shares", h.access.RequireBucketAccess(models.StorageRoleReader), h.ListBucketShares)
			auth.DELETE("/shares/:shareId", h.RevokeShare)
			auth.GET("/my/shares", h.ListMyShares)

			// Object Sharing (shareable pre-signed URLs)
			auth.POST("/buckets/:bucket/share-object", h.access.RequireBucketAccess(models.StorageRoleReader), h.ShareObject)

			// Access Policies
			auth.POST("/policies", h.CreatePolicy)
			auth.GET("/policies", h.ListPolicies)
			auth.DELETE("/policies/:tenantId/:userId/:bucket", h.DeletePolicy)

			// Access Keys
			auth.POST("/access-keys", h.CreateAccessKey)
			auth.GET("/access-keys", h.ListAccessKeys)
			auth.DELETE("/access-keys/:keyId", h.RevokeAccessKey)

			// Bucket Lifecycle
			auth.GET("/buckets/:bucket/lifecycle", h.access.RequireBucketAccess(models.StorageRoleReader), h.GetBucketLifecycle)

			// Admin-only
			admin := auth.Group("", h.access.RequireStorageRole(models.StorageRoleAdmin))
			{
				admin.GET("/system/metrics", h.SystemMetrics)
				admin.GET("/system/metrics/:bucket", h.BucketMetricsHandler)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Health & Monitoring
// ---------------------------------------------------------------------------

func (h *Handler) Health(c *gin.Context) {
	err := h.client.Ping(context.Background())
	status := "healthy"
	if err != nil {
		status = "unhealthy"
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"backend":  "native",
		"endpoint": h.endpoint,
		"features": gin.H{
			"versioning":     true,
			"lifecycle":      true,
			"encryption":     true,
			"objectLock":     true,
			"cors":           true,
			"notifications":  true,
			"bucketPolicy":   true,
			"presignedUrls":  true,
			"multiDelete":    true,
			"serverSideCopy": true,
			"tagging":        true,
			"accessKeys":     true,
			"bucketSharing":  true,
			"quotas":         true,
			"iamIntegrated":  true,
		},
	})
}

func (h *Handler) Stats(c *gin.Context) {
	sc := access.GetStorageContext(c)
	tenantID := ""
	if sc != nil {
		tenantID = sc.TenantID
	}

	buckets := h.store.List(tenantID)
	var totalObjects int64
	var totalSize int64
	for _, b := range buckets {
		totalObjects += b.Status.ObjectCount
		totalSize += b.Status.TotalSize
	}

	tenantCount := 1
	if tenantID == "" {
		seen := map[string]struct{}{}
		for _, b := range h.store.ListAll() {
			seen[b.Metadata.TenantID] = struct{}{}
		}
		tenantCount = len(seen)
	}

	c.JSON(http.StatusOK, models.StorageStats{
		TotalBuckets:   len(buckets),
		TotalObjects:   totalObjects,
		TotalSizeBytes: totalSize,
		TenantCount:    tenantCount,
	})
}

func (h *Handler) SystemMetrics(c *gin.Context) {
	all := h.store.ListAll()
	var totalObjects int64
	var totalSize int64
	for _, b := range all {
		totalObjects += b.Status.ObjectCount
		totalSize += b.Status.TotalSize
	}

	seen := map[string]struct{}{}
	for _, b := range all {
		seen[b.Metadata.TenantID] = struct{}{}
	}

	err := h.client.Ping(context.Background())
	m := h.metrics.GetSystemMetrics(
		len(all),
		int(totalObjects),
		totalSize,
		len(seen),
		h.access.PolicyCount(),
		err == nil,
	)
	m.ActiveAccessKeys = h.access.ActiveAccessKeyCount()
	m.ActiveShares = h.access.ActiveShareCount()

	c.JSON(http.StatusOK, m)
}

func (h *Handler) BucketMetricsHandler(c *gin.Context) {
	sc := access.GetStorageContext(c)
	tenantID := ""
	if sc != nil {
		tenantID = sc.TenantID
	}
	bucket := c.Param("bucket")
	c.JSON(http.StatusOK, h.metrics.GetBucketMetrics(tenantID, bucket))
}

// ---------------------------------------------------------------------------
// Audit Events
// ---------------------------------------------------------------------------

func (h *Handler) ListEvents(c *gin.Context) {
	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	eventType := c.Query("type")
	c.JSON(http.StatusOK, h.audit.List("", eventType, limit))
}

func (h *Handler) ListBucketEvents(c *gin.Context) {
	bucket := c.Param("bucket")
	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	c.JSON(http.StatusOK, h.audit.ListByBucket(bucket, limit))
}

// ---------------------------------------------------------------------------
// Bucket CRUD
// ---------------------------------------------------------------------------

func (h *Handler) CreateBucket(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req struct {
		Name       string                   `json:"name" binding:"required"`
		Versioning models.VersioningStatus  `json:"versioning,omitempty"`
		Quota      int64                    `json:"quota,omitempty"`
		Encryption *models.BucketEncryption `json:"encryption,omitempty"`
		ObjectLock *models.ObjectLockConfig `json:"objectLock,omitempty"`
		Region     string                   `json:"region,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spec := models.BucketSpec{
		Versioning: req.Versioning,
		Quota:      req.Quota,
		Region:     req.Region,
	}
	if req.Encryption != nil {
		spec.Encryption = *req.Encryption
	}
	if req.ObjectLock != nil {
		spec.ObjectLock = *req.ObjectLock
	}

	bucket, err := h.tenant.CreateTenantBucket(sc.TenantID, req.Name, spec)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	bucket.Metadata.CreatedBy = sc.UserID

	// Trigger immediate reconciliation.
	h.controller.ReconcileOne(context.Background(), sc.TenantID, req.Name)

	// Set encryption on backend if requested.
	if req.Encryption != nil && req.Encryption.Enabled {
		_ = h.client.SetBucketEncryption(context.Background(), bucket.Spec.Name, *req.Encryption)
	}
	// Set object lock on backend if requested.
	if req.ObjectLock != nil && req.ObjectLock.Enabled {
		_ = h.client.SetObjectLockConfig(context.Background(), bucket.Spec.Name, *req.ObjectLock)
	}

	h.audit.Record(models.StorageEvent{
		Type:     "bucket.created",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   req.Name,
		Details:  fmt.Sprintf("bucket %q created by %s", req.Name, sc.UserID),
		SourceIP: c.ClientIP(),
	})

	log.Printf("Storage: bucket %q created for tenant %q by %s", req.Name, sc.TenantID, sc.UserID)
	c.JSON(http.StatusCreated, bucket)
}

func (h *Handler) ListBuckets(c *gin.Context) {
	sc := access.GetStorageContext(c)
	tenantID := ""
	if sc != nil {
		tenantID = sc.TenantID
	}
	c.JSON(http.StatusOK, h.store.List(tenantID))
}

func (h *Handler) GetBucket(c *gin.Context) {
	sc := access.GetStorageContext(c)
	tenantID := ""
	if sc != nil {
		tenantID = sc.TenantID
	}
	name := c.Param("bucket")

	b, err := h.store.Get(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) DeleteBucket(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	name := c.Param("bucket")

	b, err := h.store.Get(sc.TenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.client.DeleteBucket(context.Background(), b.Spec.Name); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	if err := h.tenant.DeleteTenantBucket(sc.TenantID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.audit.Record(models.StorageEvent{
		Type:     "bucket.deleted",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   name,
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{"deleted": name})
}

// ---------------------------------------------------------------------------
// Bucket Tagging
// ---------------------------------------------------------------------------

func (h *Handler) GetBucketTags(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	tags, err := h.client.GetBucketTagging(context.Background(), b.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "tags": tags})
}

func (h *Handler) SetBucketTags(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		Tags []models.BucketTag `json:"tags" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.PutBucketTagging(context.Background(), b.Spec.Name, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "tags": req.Tags})
}

func (h *Handler) DeleteBucketTags(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.DeleteBucketTagging(context.Background(), b.Spec.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "bucket": bucket})
}

// ---------------------------------------------------------------------------
// Bucket Encryption
// ---------------------------------------------------------------------------

func (h *Handler) GetBucketEncryption(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	enc, err := h.client.GetBucketEncryption(context.Background(), b.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "encryption": enc})
}

func (h *Handler) SetBucketEncryption(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req models.BucketEncryption
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.SetBucketEncryption(context.Background(), b.Spec.Name, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.audit.Record(models.StorageEvent{
		Type: "bucket.encryption.set", TenantID: sc.TenantID, UserID: sc.UserID, Bucket: bucket,
		Details: fmt.Sprintf("encryption %s enabled=%v", req.Algorithm, req.Enabled), SourceIP: c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "encryption": req})
}

func (h *Handler) DeleteBucketEncryption(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.DeleteBucketEncryption(context.Background(), b.Spec.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "bucket": bucket})
}

// ---------------------------------------------------------------------------
// Object Lock / Retention
// ---------------------------------------------------------------------------

func (h *Handler) GetObjectLockConfig(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	cfg, err := h.client.GetObjectLockConfig(context.Background(), b.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "objectLock": cfg})
}

func (h *Handler) SetObjectLockConfig(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req models.ObjectLockConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.SetObjectLockConfig(context.Background(), b.Spec.Name, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.audit.Record(models.StorageEvent{
		Type: "bucket.objectlock.set", TenantID: sc.TenantID, UserID: sc.UserID, Bucket: bucket,
		Details: fmt.Sprintf("object lock mode=%s enabled=%v", req.Mode, req.Enabled), SourceIP: c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "objectLock": req})
}

func (h *Handler) GetObjectRetention(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
		return
	}
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	until, mode, err := h.client.GetObjectRetention(context.Background(), b.Spec.Name, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "key": key, "retainUntil": until, "mode": mode})
}

func (h *Handler) SetObjectRetention(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
		return
	}
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		RetainUntil time.Time `json:"retainUntil" binding:"required"`
		Mode        string    `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Mode != "GOVERNANCE" && req.Mode != "COMPLIANCE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be GOVERNANCE or COMPLIANCE"})
		return
	}
	if err := h.client.PutObjectRetention(context.Background(), b.Spec.Name, key, req.RetainUntil, req.Mode); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "key": key, "retainUntil": req.RetainUntil, "mode": req.Mode})
}

func (h *Handler) GetObjectLegalHold(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
		return
	}
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	hold, err := h.client.GetObjectLegalHold(context.Background(), b.Spec.Name, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "key": key, "legalHold": hold})
}

func (h *Handler) SetObjectLegalHold(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
		return
	}
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		LegalHold bool `json:"legalHold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.PutObjectLegalHold(context.Background(), b.Spec.Name, key, req.LegalHold); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "key": key, "legalHold": req.LegalHold})
}

// ---------------------------------------------------------------------------
// CORS
// ---------------------------------------------------------------------------

func (h *Handler) GetBucketCORS(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	rules, err := h.client.GetBucketCORS(context.Background(), b.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "cors": rules})
}

func (h *Handler) SetBucketCORS(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		Rules []models.CORSRule `json:"rules" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.SetBucketCORS(context.Background(), b.Spec.Name, req.Rules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "cors": req.Rules})
}

func (h *Handler) DeleteBucketCORS(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.DeleteBucketCORS(context.Background(), b.Spec.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "bucket": bucket})
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

func (h *Handler) GetBucketNotifications(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	cfg, err := h.client.GetBucketNotification(context.Background(), b.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "notifications": cfg})
}

func (h *Handler) SetBucketNotifications(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req models.BucketNotificationConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.SetBucketNotification(context.Background(), b.Spec.Name, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.audit.Record(models.StorageEvent{
		Type: "bucket.notifications.set", TenantID: sc.TenantID, UserID: sc.UserID, Bucket: bucket,
		Details: fmt.Sprintf("%d notification rules", len(req.Rules)), SourceIP: c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "notifications": req})
}

// ---------------------------------------------------------------------------
// Bucket Policy
// ---------------------------------------------------------------------------

func (h *Handler) GetBucketPolicy(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	pol, err := h.client.GetBucketPolicy(context.Background(), b.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "policy": pol})
}

func (h *Handler) SetBucketPolicy(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req models.S3BucketPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.SetBucketPolicy(context.Background(), b.Spec.Name, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.audit.Record(models.StorageEvent{
		Type: "bucket.policy.set", TenantID: sc.TenantID, UserID: sc.UserID, Bucket: bucket,
		SourceIP: c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "policy": req})
}

func (h *Handler) DeleteBucketPolicy(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.DeleteBucketPolicy(context.Background(), b.Spec.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "bucket": bucket})
}

// ---------------------------------------------------------------------------
// Quota
// ---------------------------------------------------------------------------

func (h *Handler) GetBucketQuota(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	q, err := h.client.GetBucketQuota(context.Background(), b.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	q.TenantID = sc.TenantID
	c.JSON(http.StatusOK, q)
}

// ---------------------------------------------------------------------------
// Object Operations
// ---------------------------------------------------------------------------

func (h *Handler) PutObject(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "object key is required"})
		return
	}

	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	start := time.Now()

	// Check for enhanced options in headers.
	opts := models.PutObjectOptions{
		ContentType:  contentType,
		StorageClass: c.GetHeader("X-Storage-Class"),
	}

	// User metadata from X-Axiom-Meta-* headers.
	userMeta := make(map[string]string)
	for _, key := range []string{} {
		_ = key
	}
	for k, vals := range c.Request.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-axiom-meta-") {
			metaKey := strings.TrimPrefix(strings.ToLower(k), "x-axiom-meta-")
			if len(vals) > 0 {
				userMeta[metaKey] = vals[0]
			}
		}
	}
	if len(userMeta) > 0 {
		opts.UserMetadata = userMeta
	}

	if c.GetHeader("X-Storage-Encrypt") == "true" {
		opts.Encryption = true
	}

	err = h.client.PutObjectWithOptions(context.Background(), b.Spec.Name, key, c.Request.Body, c.Request.ContentLength, opts)
	latency := time.Since(start)
	if err != nil {
		h.metrics.RecordRequest(sc.TenantID, bucket, "PUT", 0, latency, true)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordRequest(sc.TenantID, bucket, "PUT", c.Request.ContentLength, latency, false)

	h.audit.Record(models.StorageEvent{
		Type:     "object.uploaded",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   bucket,
		Key:      key,
		Size:     c.Request.ContentLength,
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"bucket": bucket,
		"key":    key,
		"size":   c.Request.ContentLength,
	})
}

func (h *Handler) GetObject(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()
	reader, err := h.client.GetObject(context.Background(), b.Spec.Name, key)
	if err != nil {
		h.metrics.RecordRequest(sc.TenantID, bucket, "GET", 0, time.Since(start), true)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Set content type from metadata if available.
	info, _ := h.client.StatObject(context.Background(), b.Spec.Name, key)
	if info != nil && info.ContentType != "" {
		c.Header("Content-Type", info.ContentType)
		c.Header("Content-Length", strconv.FormatInt(info.Size, 10))
		c.Header("ETag", info.ETag)
		c.Header("Last-Modified", info.LastModified.UTC().Format(http.TimeFormat))
		if info.StorageClass != "" {
			c.Header("X-Storage-Class", info.StorageClass)
		}
	}

	n, _ := io.Copy(c.Writer, reader)
	latency := time.Since(start)
	h.metrics.RecordRequest(sc.TenantID, bucket, "GET", n, latency, false)

	h.audit.Record(models.StorageEvent{
		Type:     "object.downloaded",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   bucket,
		Key:      key,
		Size:     n,
		SourceIP: c.ClientIP(),
	})
}

func (h *Handler) HeadObject(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	info, err := h.client.StatObject(context.Background(), b.Spec.Name, key)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Content-Type", info.ContentType)
	c.Header("Content-Length", strconv.FormatInt(info.Size, 10))
	c.Header("ETag", info.ETag)
	c.Header("Last-Modified", info.LastModified.UTC().Format(http.TimeFormat))
	if info.StorageClass != "" {
		c.Header("X-Storage-Class", info.StorageClass)
	}
	if info.RetainUntil != nil {
		c.Header("X-Object-Retain-Until", info.RetainUntil.Format(time.RFC3339))
	}
	if info.LegalHold {
		c.Header("X-Object-Legal-Hold", "true")
	}
	c.Status(http.StatusOK)
}

func (h *Handler) DeleteObject(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	key := strings.TrimPrefix(c.Param("key"), "/")

	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()
	if err := h.client.DeleteObject(context.Background(), b.Spec.Name, key); err != nil {
		h.metrics.RecordRequest(sc.TenantID, bucket, "DELETE", 0, time.Since(start), true)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordRequest(sc.TenantID, bucket, "DELETE", 0, time.Since(start), false)

	h.audit.Record(models.StorageEvent{
		Type:     "object.deleted",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   bucket,
		Key:      key,
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{"deleted": key, "bucket": bucket})
}

func (h *Handler) ListObjects(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	prefix := c.Query("prefix")

	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	objects, err := h.client.ListObjects(context.Background(), b.Spec.Name, prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if objects == nil {
		objects = []models.ObjectInfo{}
	}

	c.JSON(http.StatusOK, gin.H{
		"bucket":  bucket,
		"prefix":  prefix,
		"count":   len(objects),
		"objects": objects,
	})
}

// ---------------------------------------------------------------------------
// Object Metadata
// ---------------------------------------------------------------------------

func (h *Handler) GetObjectMetadata(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
		return
	}
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	meta, err := h.client.GetObjectMetadata(context.Background(), b.Spec.Name, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "key": key, "metadata": meta})
}

func (h *Handler) PutObjectMetadata(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key query parameter is required"})
		return
	}
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		Metadata map[string]string `json:"metadata" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.client.PutObjectMetadata(context.Background(), b.Spec.Name, key, req.Metadata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "key": key, "metadata": req.Metadata})
}

// ---------------------------------------------------------------------------
// Batch Operations
// ---------------------------------------------------------------------------

func (h *Handler) MultiDeleteObjects(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req models.MultiDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keys := make([]string, len(req.Objects))
	for i, o := range req.Objects {
		keys[i] = o.Key
	}

	deleted, errors, _ := h.client.MultiDeleteObjects(context.Background(), b.Spec.Name, keys)

	h.audit.Record(models.StorageEvent{
		Type:     "object.multi-deleted",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   bucket,
		Details:  fmt.Sprintf("deleted %d objects, %d errors", deleted, len(errors)),
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"deleted": deleted,
		"errors":  errors,
	})
}

func (h *Handler) CopyObject(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req models.CopyObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	srcB, err := h.store.Get(sc.TenantID, req.SourceBucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "source bucket: " + err.Error()})
		return
	}
	dstB, err := h.store.Get(sc.TenantID, req.DestBucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dest bucket: " + err.Error()})
		return
	}

	if err := h.client.CopyObject(context.Background(), srcB.Spec.Name, req.SourceKey, dstB.Spec.Name, req.DestKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.audit.Record(models.StorageEvent{
		Type:     "object.copied",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   req.SourceBucket,
		Key:      req.SourceKey,
		Details:  fmt.Sprintf("to %s/%s", req.DestBucket, req.DestKey),
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"source": fmt.Sprintf("%s/%s", req.SourceBucket, req.SourceKey),
		"dest":   fmt.Sprintf("%s/%s", req.DestBucket, req.DestKey),
	})
}

// ---------------------------------------------------------------------------
// Pre-signed URLs
// ---------------------------------------------------------------------------

func (h *Handler) PresignURL(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Key     string `json:"key" binding:"required"`
		Method  string `json:"method"`  // "GET" or "PUT", default GET
		Expires int    `json:"expires"` // seconds, default 3600
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "GET"
	}
	expires := time.Duration(req.Expires) * time.Second
	if expires <= 0 {
		expires = time.Hour
	}

	var presignURL string
	switch method {
	case "GET":
		presignURL, err = h.client.PresignGetObject(context.Background(), b.Spec.Name, req.Key, expires)
	case "PUT":
		presignURL, err = h.client.PresignPutObject(context.Background(), b.Spec.Name, req.Key, expires)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "method must be GET or PUT"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	presignURL = absoluteURL(c, presignURL)

	h.audit.Record(models.StorageEvent{
		Type:     "presign.generated",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   bucket,
		Key:      req.Key,
		Details:  fmt.Sprintf("method=%s expires=%v", method, expires),
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, models.PreSignedURLResponse{
		URL:       presignURL,
		ExpiresAt: time.Now().Add(expires),
		Method:    method,
		Bucket:    bucket,
		Key:       req.Key,
	})
}

// ---------------------------------------------------------------------------
// Bucket Sharing
// ---------------------------------------------------------------------------

func (h *Handler) CreateBucketShare(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	// Verify bucket exists.
	if _, err := h.store.Get(sc.TenantID, bucket); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		GranteeType string             `json:"granteeType" binding:"required"` // "user", "application", "service-account"
		GranteeID   string             `json:"granteeId" binding:"required"`
		GranteeName string             `json:"granteeName,omitempty"`
		Role        models.StorageRole `json:"role" binding:"required"`
		Prefix      string             `json:"prefix,omitempty"`
		ExpiresAt   *time.Time         `json:"expiresAt,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	share := models.BucketShare{
		BucketName:  bucket,
		TenantID:    sc.TenantID,
		GranteeType: req.GranteeType,
		GranteeID:   req.GranteeID,
		GranteeName: req.GranteeName,
		Role:        req.Role,
		Prefix:      req.Prefix,
		ExpiresAt:   req.ExpiresAt,
		SharedBy:    sc.UserID,
	}

	result, err := h.access.ShareBucket(share)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) ListBucketShares(c *gin.Context) {
	sc := access.GetStorageContext(c)
	bucket := c.Param("bucket")
	shares := h.access.ListBucketShares(sc.TenantID, bucket)
	if shares == nil {
		shares = []*models.BucketShare{}
	}
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "shares": shares})
}

func (h *Handler) ListMyShares(c *gin.Context) {
	sc := access.GetStorageContext(c)
	shares := h.access.ListUserShares(sc.UserID)
	if shares == nil {
		shares = []*models.BucketShare{}
	}
	c.JSON(http.StatusOK, gin.H{"shares": shares})
}

func (h *Handler) RevokeShare(c *gin.Context) {
	sc := access.GetStorageContext(c)
	shareID := c.Param("shareId")
	if err := h.access.RevokeShare(shareID, sc.UserID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"revoked": shareID})
}

// ---------------------------------------------------------------------------
// Access Keys
// ---------------------------------------------------------------------------

func (h *Handler) CreateAccessKey(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req struct {
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description,omitempty"`
		Role        models.StorageRole `json:"role" binding:"required"`
		BucketScope []string           `json:"bucketScope,omitempty"`
		ExpiresAt   *time.Time         `json:"expiresAt,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ak, err := h.access.CreateAccessKey(sc.UserID, sc.TenantID, req.Name, req.Description, req.Role, req.BucketScope, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the full key including secret ONLY on creation.
	c.JSON(http.StatusCreated, ak)
}

func (h *Handler) ListAccessKeys(c *gin.Context) {
	sc := access.GetStorageContext(c)
	keys := h.access.ListAccessKeys("", sc.TenantID)
	if keys == nil {
		keys = []*models.AccessKey{}
	}
	c.JSON(http.StatusOK, gin.H{"accessKeys": keys})
}

func (h *Handler) RevokeAccessKey(c *gin.Context) {
	keyID := c.Param("keyId")
	// Admin can revoke any key (userID="" bypasses ownership check).
	if err := h.access.RevokeAccessKey(keyID, ""); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"revoked": keyID})
}

// ---------------------------------------------------------------------------
// Access Policies
// ---------------------------------------------------------------------------

func (h *Handler) CreatePolicy(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req models.TenantPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	req.BucketName = strings.TrimSpace(req.BucketName)
	if req.UserID == "" || req.BucketName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and bucketName are required"})
		return
	}

	// Non-admin users can only create policies in their own tenant.
	if hasAnyRole(sc.Roles, "sysadmin", "system-manager", "admin") {
		if strings.TrimSpace(req.TenantID) == "" {
			req.TenantID = sc.TenantID
		}
	} else {
		req.TenantID = sc.TenantID
	}

	req.GrantedBy = sc.UserID

	if err := h.access.SetPolicy(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.audit.Record(models.StorageEvent{
		Type:     "policy.created",
		TenantID: req.TenantID,
		UserID:   sc.UserID,
		Bucket:   req.BucketName,
		Details:  fmt.Sprintf("role=%s for user=%s", req.Role, req.UserID),
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusCreated, req)
}

func (h *Handler) ListPolicies(c *gin.Context) {
	sc := access.GetStorageContext(c)
	tenantFilter := sc.TenantID
	if hasAnyRole(sc.Roles, "sysadmin", "system-manager", "admin") {
		tenantFilter = strings.TrimSpace(c.Query("tenantId")) // empty => list all
	}

	policies := h.access.ListPolicies(tenantFilter)
	if policies == nil {
		policies = []*models.TenantPolicy{}
	}
	c.JSON(http.StatusOK, gin.H{"policies": policies})
}

func (h *Handler) DeletePolicy(c *gin.Context) {
	sc := access.GetStorageContext(c)
	tenantID := c.Param("tenantId")
	userID := c.Param("userId")
	bucket := c.Param("bucket")

	// Only allow deleting own-tenant policies unless sysadmin.
	if tenantID != sc.TenantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete cross-tenant policy"})
		return
	}

	if err := h.access.DeletePolicy(tenantID, userID, bucket); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.audit.Record(models.StorageEvent{
		Type:     "policy.deleted",
		TenantID: tenantID,
		UserID:   sc.UserID,
		Bucket:   bucket,
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ---------------------------------------------------------------------------
// Object Sharing (shareable pre-signed download URLs)
// ---------------------------------------------------------------------------

func (h *Handler) ShareObject(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Key     string `json:"key" binding:"required"`
		Expires int    `json:"expires"` // seconds, default 3600
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expires := time.Duration(req.Expires) * time.Second
	if expires <= 0 {
		expires = time.Hour
	}
	if expires > 7*24*time.Hour {
		expires = 7 * 24 * time.Hour
	}

	shareURL, err := h.client.PresignGetObject(context.Background(), b.Spec.Name, req.Key, expires)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	shareURL = absoluteURL(c, shareURL)

	h.audit.Record(models.StorageEvent{
		Type:     "object.shared",
		TenantID: sc.TenantID,
		UserID:   sc.UserID,
		Bucket:   bucket,
		Key:      req.Key,
		Details:  fmt.Sprintf("expires=%v", expires),
		SourceIP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"url":       shareURL,
		"bucket":    bucket,
		"key":       req.Key,
		"expiresAt": time.Now().Add(expires).Format(time.RFC3339),
		"expiresIn": int(expires.Seconds()),
	})
}

func hasAnyRole(roles []string, allowed ...string) bool {
	if len(roles) == 0 || len(allowed) == 0 {
		return false
	}
	for _, role := range roles {
		r := strings.ToLower(strings.TrimSpace(role))
		for _, a := range allowed {
			if r == strings.ToLower(strings.TrimSpace(a)) {
				return true
			}
		}
	}
	return false
}

func absoluteURL(c *gin.Context, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}

	scheme := "http"
	if fp := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); fp != "" {
		scheme = strings.ToLower(strings.Split(fp, ",")[0])
	} else if c.Request != nil && c.Request.TLS != nil {
		scheme = "https"
	}

	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" && c.Request != nil {
		host = c.Request.Host
	}
	if host == "" {
		return raw
	}

	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, raw)
}

// ---------------------------------------------------------------------------
// Bucket Lifecycle
// ---------------------------------------------------------------------------

func (h *Handler) GetBucketLifecycle(c *gin.Context) {
	sc := access.GetStorageContext(c)
	if sc == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	bucket := c.Param("bucket")
	b, err := h.store.Get(sc.TenantID, bucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bucket":      b.Spec.Name,
		"versioning":  b.Spec.Versioning,
		"encryption":  b.Spec.Encryption,
		"objectLock":  b.Spec.ObjectLock,
		"quota":       b.Spec.Quota,
		"labels":      b.Metadata.Labels,
		"createdAt":   b.Metadata.CreatedAt,
		"phase":       b.Status.Phase,
		"objectCount": b.Status.ObjectCount,
		"totalSize":   b.Status.TotalSize,
	})
}
