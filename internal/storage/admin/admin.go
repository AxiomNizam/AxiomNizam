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

	"example.com/axiomnizam/internal/storage/controller"
	"example.com/axiomnizam/internal/storage/events"
	storageMetrics "example.com/axiomnizam/internal/storage/metrics"
	"example.com/axiomnizam/internal/storage/models"
	"example.com/axiomnizam/internal/storage/policy"
	"example.com/axiomnizam/internal/storage/store"
	"example.com/axiomnizam/internal/storage/tenant"
	"github.com/gin-gonic/gin"
)

// Handler exposes the object storage API endpoints.
type Handler struct {
	store      *store.BucketStore
	client     models.Backend
	tenant     *tenant.Manager
	controller *controller.BucketController
	policy     *policy.Controller
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
	p *policy.Controller,
	m *storageMetrics.Collector,
	a *events.AuditLog,
	endpoint string,
) *Handler {
	return &Handler{
		store:      s,
		client:     client,
		tenant:     t,
		controller: ctrl,
		policy:     p,
		metrics:    m,
		audit:      a,
		endpoint:   endpoint,
	}
}

// RegisterRoutes registers all object storage API routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	sg := rg.Group("/storage")
	{
		// Health & Monitoring
		sg.GET("/health", h.Health)
		sg.GET("/stats", h.Stats)
		sg.GET("/metrics", h.SystemMetrics)
		sg.GET("/metrics/:bucket", h.BucketMetricsHandler)

		// Audit Events
		sg.GET("/events", h.ListEvents)
		sg.GET("/events/:bucket", h.ListBucketEvents)

		// Bucket CRUD
		sg.POST("/buckets", h.CreateBucket)
		sg.GET("/buckets", h.ListBuckets)
		sg.GET("/buckets/:name", h.GetBucket)
		sg.DELETE("/buckets/:name", h.DeleteBucket)

		// Bucket Tagging
		sg.GET("/buckets/:name/tags", h.GetBucketTags)
		sg.PUT("/buckets/:name/tags", h.SetBucketTags)
		sg.DELETE("/buckets/:name/tags", h.DeleteBucketTags)

		// Object operations
		sg.PUT("/buckets/:name/objects/*key", h.PutObject)
		sg.GET("/buckets/:name/objects/*key", h.GetObject)
		sg.DELETE("/buckets/:name/objects/*key", h.DeleteObject)
		sg.GET("/buckets/:name/objects", h.ListObjects)

		// Batch operations
		sg.POST("/buckets/:name/multi-delete", h.MultiDeleteObjects)
		sg.POST("/copy", h.CopyObject)

		// Pre-signed URLs
		sg.POST("/buckets/:name/presign", h.PresignURL)

		// Policies
		sg.POST("/policies", h.CreatePolicy)
		sg.GET("/policies", h.ListPolicies)
		sg.DELETE("/policies/:tenantId/:userId/:bucket", h.DeletePolicy)
	}
}

// ---------- Health & Stats ----------

func (h *Handler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	checkedAt := time.Now().UTC()

	if err := h.client.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"error":     err.Error(),
			"endpoint":  h.endpoint,
			"checkedAt": checkedAt,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"endpoint":  h.endpoint,
		"checkedAt": checkedAt,
	})
}

func (h *Handler) Stats(c *gin.Context) {
	buckets := h.store.ListAll()
	tenants := make(map[string]struct{})
	var totalObjects, totalSize int64

	for _, b := range buckets {
		tenants[b.Metadata.TenantID] = struct{}{}
		totalObjects += b.Status.ObjectCount
		totalSize += b.Status.TotalSize
	}

	c.JSON(http.StatusOK, models.StorageStats{
		TotalBuckets:   len(buckets),
		TotalObjects:   totalObjects,
		TotalSizeBytes: totalSize,
		TenantCount:    len(tenants),
	})
}

// SystemMetrics returns detailed system metrics including request counts, throughput, etc.
func (h *Handler) SystemMetrics(c *gin.Context) {
	buckets := h.store.ListAll()
	tenants := make(map[string]struct{})
	var totalObjects, totalSize int64

	for _, b := range buckets {
		tenants[b.Metadata.TenantID] = struct{}{}
		totalObjects += b.Status.ObjectCount
		totalSize += b.Status.TotalSize
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	healthy := h.client.Ping(ctx) == nil

	sm := h.metrics.GetSystemMetrics(
		len(buckets), int(totalObjects), totalSize,
		len(tenants), len(h.policy.ListPolicies("")),
		healthy,
	)
	c.JSON(http.StatusOK, sm)
}

// BucketMetricsHandler returns metrics for a specific bucket.
func (h *Handler) BucketMetricsHandler(c *gin.Context) {
	bucket := c.Param("bucket")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}
	bm := h.metrics.GetBucketMetrics(tenantID, bucket)
	c.JSON(http.StatusOK, bm)
}

// ---------- Audit Events ----------

func (h *Handler) ListEvents(c *gin.Context) {
	tenantID := c.Query("tenantId")
	eventType := c.Query("type")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 100
	}
	evts := h.audit.List(tenantID, eventType, limit)
	if evts == nil {
		evts = []models.StorageEvent{}
	}
	c.JSON(http.StatusOK, gin.H{
		"events": evts,
		"count":  len(evts),
		"total":  h.audit.Count(),
	})
}

func (h *Handler) ListBucketEvents(c *gin.Context) {
	bucket := c.Param("bucket")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 100
	}
	evts := h.audit.ListByBucket(bucket, limit)
	if evts == nil {
		evts = []models.StorageEvent{}
	}
	c.JSON(http.StatusOK, gin.H{
		"events": evts,
		"count":  len(evts),
		"bucket": bucket,
	})
}

// ---------- Bucket CRUD ----------

type createBucketRequest struct {
	Name            string                  `json:"name" binding:"required"`
	TenantID        string                  `json:"tenantId" binding:"required"`
	Versioning      models.VersioningStatus `json:"versioning"`
	LifecyclePolicy []models.LifecycleRule  `json:"lifecyclePolicy,omitempty"`
	Region          string                  `json:"region,omitempty"`
	Quota           int64                   `json:"quota,omitempty"`
}

func (h *Handler) CreateBucket(c *gin.Context) {
	start := time.Now()
	var req createBucketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Versioning == "" {
		req.Versioning = models.VersioningDisabled
	}

	spec := models.BucketSpec{
		Name:            h.tenant.ResolveBucketName(req.TenantID, req.Name),
		Versioning:      req.Versioning,
		LifecyclePolicy: req.LifecyclePolicy,
		Region:          req.Region,
		Quota:           req.Quota,
	}

	bucket, err := h.tenant.CreateTenantBucket(req.TenantID, req.Name, spec)
	if err != nil {
		h.metrics.RecordRequest(req.TenantID, req.Name, "PUT", 0, time.Since(start), true)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.controller.ReconcileOne(ctx, req.TenantID, req.Name); err != nil {
			log.Printf("⚠️  Storage: async reconcile failed for %s/%s: %v", req.TenantID, req.Name, err)
		}
	}()

	h.metrics.RecordRequest(req.TenantID, req.Name, "PUT", 0, time.Since(start), false)
	h.audit.Record(events.EventBucketCreated, req.TenantID, "", req.Name, "", 0, "versioning="+string(req.Versioning))

	c.JSON(http.StatusCreated, bucket)
}

func (h *Handler) ListBuckets(c *gin.Context) {
	tenantID := c.Query("tenantId")
	buckets := h.store.List(tenantID)
	if buckets == nil {
		buckets = []*models.BucketResource{}
	}
	c.JSON(http.StatusOK, buckets)
}

func (h *Handler) GetBucket(c *gin.Context) {
	name := c.Param("name")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bucket)
}

func (h *Handler) DeleteBucket(c *gin.Context) {
	start := time.Now()
	name := c.Param("name")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	if err := h.client.DeleteBucket(ctx, bucket.Spec.Name); err != nil {
		h.metrics.RecordRequest(tenantID, name, "DELETE", 0, time.Since(start), true)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("backend delete failed: %v", err)})
		return
	}

	if err := h.store.Delete(tenantID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordRequest(tenantID, name, "DELETE", 0, time.Since(start), false)
	h.audit.Record(events.EventBucketDeleted, tenantID, "", name, "", 0, "")

	c.JSON(http.StatusOK, gin.H{"message": "bucket deleted", "name": name})
}

// ---------- Object Operations ----------

func (h *Handler) PutObject(c *gin.Context) {
	start := time.Now()
	bucketName := c.Param("name")
	key := strings.TrimPrefix(c.Param("key"), "/")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	size := c.Request.ContentLength

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	if err := h.client.PutObject(ctx, bucket.Spec.Name, key, c.Request.Body, size, contentType); err != nil {
		h.metrics.RecordRequest(tenantID, bucketName, "PUT", size, time.Since(start), true)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("upload failed: %v", err)})
		return
	}

	h.metrics.RecordRequest(tenantID, bucketName, "PUT", size, time.Since(start), false)
	h.audit.Record(events.EventObjectUploaded, tenantID, "", bucketName, key, size, contentType)

	c.JSON(http.StatusOK, gin.H{
		"message": "object uploaded",
		"bucket":  bucketName,
		"key":     key,
		"size":    size,
	})
}

func (h *Handler) GetObject(c *gin.Context) {
	start := time.Now()
	bucketName := c.Param("name")
	key := strings.TrimPrefix(c.Param("key"), "/")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	obj, err := h.client.GetObject(ctx, bucket.Spec.Name, key)
	if err != nil {
		h.metrics.RecordRequest(tenantID, bucketName, "GET", 0, time.Since(start), true)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer obj.Close()

	var objSize int64
	info, err := h.client.StatObject(ctx, bucket.Spec.Name, key)
	if err == nil {
		c.Header("Content-Type", info.ContentType)
		c.Header("Content-Length", strconv.FormatInt(info.Size, 10))
		c.Header("ETag", info.ETag)
		objSize = info.Size
	}

	c.Status(http.StatusOK)
	io.Copy(c.Writer, obj)

	h.metrics.RecordRequest(tenantID, bucketName, "GET", objSize, time.Since(start), false)
	h.audit.Record(events.EventObjectDownloaded, tenantID, "", bucketName, key, objSize, "")
}

func (h *Handler) DeleteObject(c *gin.Context) {
	start := time.Now()
	bucketName := c.Param("name")
	key := strings.TrimPrefix(c.Param("key"), "/")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	if err := h.client.DeleteObject(ctx, bucket.Spec.Name, key); err != nil {
		h.metrics.RecordRequest(tenantID, bucketName, "DELETE", 0, time.Since(start), true)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordRequest(tenantID, bucketName, "DELETE", 0, time.Since(start), false)
	h.audit.Record(events.EventObjectDeleted, tenantID, "", bucketName, key, 0, "")

	c.JSON(http.StatusOK, gin.H{"message": "object deleted", "key": key})
}

func (h *Handler) ListObjects(c *gin.Context) {
	bucketName := c.Param("name")
	prefix := c.Query("prefix")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	objects, err := h.client.ListObjects(ctx, bucket.Spec.Name, prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if objects == nil {
		objects = []models.ObjectInfo{}
	}

	c.JSON(http.StatusOK, gin.H{
		"bucket":  bucketName,
		"prefix":  prefix,
		"objects": objects,
		"count":   len(objects),
	})
}

// ---------- Pre-signed URLs ----------

type presignRequest struct {
	Key       string `json:"key" binding:"required"`
	Method    string `json:"method"`
	ExpiresIn int    `json:"expiresIn,omitempty"`
}

func (h *Handler) PresignURL(c *gin.Context) {
	bucketName := c.Param("name")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	var req presignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucket, err := h.store.Get(tenantID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.ExpiresIn <= 0 {
		req.ExpiresIn = 900
	}
	expires := time.Duration(req.ExpiresIn) * time.Second

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var presignedURL string
	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "GET"
	}

	switch method {
	case "GET":
		presignedURL, err = h.client.PresignGetObject(ctx, bucket.Spec.Name, req.Key, expires)
	case "PUT":
		presignedURL, err = h.client.PresignPutObject(ctx, bucket.Spec.Name, req.Key, expires)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "method must be GET or PUT"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.PreSignedURLResponse{
		URL:       presignedURL,
		ExpiresAt: time.Now().Add(expires),
	})

	h.audit.Record(events.EventPresignGenerated, tenantID, "", bucketName, req.Key, 0, method)
}

// ---------- Policies ----------

type createPolicyRequest struct {
	TenantID   string `json:"tenantId" binding:"required"`
	UserID     string `json:"userId" binding:"required"`
	BucketName string `json:"bucketName" binding:"required"`
	Role       string `json:"role" binding:"required"`
	Prefix     string `json:"prefix,omitempty"`
}

func (h *Handler) CreatePolicy(c *gin.Context) {
	var req createPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.StorageRole(req.Role)
	storageBucket := h.tenant.ResolveBucketName(req.TenantID, req.BucketName)

	tp, err := h.policy.GeneratePolicy(req.TenantID, req.UserID, storageBucket, role, req.Prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.audit.Record(events.EventPolicyCreated, req.TenantID, req.UserID, req.BucketName, "", 0, string(role))
	c.JSON(http.StatusCreated, tp)
}

func (h *Handler) ListPolicies(c *gin.Context) {
	tenantID := c.Query("tenantId")
	policies := h.policy.ListPolicies(tenantID)
	if policies == nil {
		policies = []*models.TenantPolicy{}
	}
	c.JSON(http.StatusOK, policies)
}

func (h *Handler) DeletePolicy(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.Param("userId")
	bucket := c.Param("bucket")

	h.policy.DeletePolicy(tenantID, userID, bucket)
	h.audit.Record(events.EventPolicyDeleted, tenantID, userID, bucket, "", 0, "")
	c.JSON(http.StatusOK, gin.H{"message": "policy deleted"})
}

// ---------- Bucket Tagging ----------

func (h *Handler) GetBucketTags(c *gin.Context) {
	name := c.Param("name")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tags, err := h.client.GetBucketTagging(ctx, bucket.Spec.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tags == nil {
		tags = []models.BucketTag{}
	}
	c.JSON(http.StatusOK, gin.H{"bucket": name, "tags": tags})
}

type setTagsRequest struct {
	Tags []models.BucketTag `json:"tags" binding:"required"`
}

func (h *Handler) SetBucketTags(c *gin.Context) {
	name := c.Param("name")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	var req setTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucket, err := h.store.Get(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.client.PutBucketTagging(ctx, bucket.Spec.Name, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tags updated", "bucket": name, "tags": req.Tags})
}

func (h *Handler) DeleteBucketTags(c *gin.Context) {
	name := c.Param("name")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	bucket, err := h.store.Get(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.client.DeleteBucketTagging(ctx, bucket.Spec.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tags deleted", "bucket": name})
}

// ---------- Multi-Delete ----------

type multiDeleteRequest struct {
	Keys []string `json:"keys" binding:"required"`
}

func (h *Handler) MultiDeleteObjects(c *gin.Context) {
	start := time.Now()
	bucketName := c.Param("name")
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		tenantID = "default"
	}

	var req multiDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucket, err := h.store.Get(tenantID, bucketName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	deleted, errs, err := h.client.MultiDeleteObjects(ctx, bucket.Spec.Name, req.Keys)
	if err != nil {
		h.metrics.RecordRequest(tenantID, bucketName, "DELETE", 0, time.Since(start), true)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordRequest(tenantID, bucketName, "DELETE", 0, time.Since(start), len(errs) > 0)
	h.audit.Record(events.EventMultiDelete, tenantID, "", bucketName, "",
		0, fmt.Sprintf("deleted=%d errors=%d", deleted, len(errs)))

	c.JSON(http.StatusOK, gin.H{
		"deleted": deleted,
		"errors":  errs,
		"total":   len(req.Keys),
	})
}

// ---------- Copy Object ----------

type copyRequest struct {
	SourceBucket string `json:"sourceBucket" binding:"required"`
	SourceKey    string `json:"sourceKey" binding:"required"`
	DestBucket   string `json:"destBucket" binding:"required"`
	DestKey      string `json:"destKey" binding:"required"`
	TenantID     string `json:"tenantId"`
}

func (h *Handler) CopyObject(c *gin.Context) {
	start := time.Now()
	var req copyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = "default"
	}

	srcBucket, err := h.store.Get(tenantID, req.SourceBucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "source bucket not found: " + err.Error()})
		return
	}

	dstBucket, err := h.store.Get(tenantID, req.DestBucket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "destination bucket not found: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.client.CopyObject(ctx, srcBucket.Spec.Name, req.SourceKey, dstBucket.Spec.Name, req.DestKey); err != nil {
		h.metrics.RecordRequest(tenantID, req.DestBucket, "PUT", 0, time.Since(start), true)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordRequest(tenantID, req.DestBucket, "PUT", 0, time.Since(start), false)
	h.audit.Record(events.EventObjectCopied, tenantID, "",
		req.SourceBucket, req.SourceKey, 0,
		fmt.Sprintf("→ %s/%s", req.DestBucket, req.DestKey))

	c.JSON(http.StatusOK, gin.H{
		"message":     "object copied",
		"source":      req.SourceBucket + "/" + req.SourceKey,
		"destination": req.DestBucket + "/" + req.DestKey,
	})
}
