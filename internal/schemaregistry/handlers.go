package schemaregistry

// HTTP handlers for the Schema Registry API.
// Provides a Confluent Schema Registry-compatible wire format.
//
// Routes:
//   GET    /api/v1/schemas/subjects                              — List subjects
//   GET    /api/v1/schemas/subjects/:subject/versions            — List versions
//   GET    /api/v1/schemas/subjects/:subject/versions/:version   — Get schema
//   POST   /api/v1/schemas/subjects/:subject/versions            — Register schema
//   POST   /api/v1/schemas/compatibility/subjects/:subject/versions/:version — Check compat
//   GET    /api/v1/schemas/ids/:id                               — Get by global ID
//   DELETE /api/v1/schemas/subjects/:subject/versions/:version   — Delete version
//   PUT    /api/v1/schemas/config/:subject                       — Set compatibility
//   GET    /api/v1/schemas/config/:subject                       — Get compatibility

import (
	"net/http"
	"strconv"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/resources"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SchemaRegistryHandlers provides HTTP handlers for schema registry operations.
type SchemaRegistryHandlers struct {
	schemaStore  store.ResourceStore[*SchemaResource]
	subjectStore store.ResourceStore[*SchemaSubjectResource]
	checker      CompatibilityChecker
}

// NewSchemaRegistryHandlers creates handlers.
func NewSchemaRegistryHandlers(
	schemaStore store.ResourceStore[*SchemaResource],
	subjectStore store.ResourceStore[*SchemaSubjectResource],
	checker CompatibilityChecker,
) *SchemaRegistryHandlers {
	return &SchemaRegistryHandlers{
		schemaStore:  schemaStore,
		subjectStore: subjectStore,
		checker:      checker,
	}
}

// RegisterRoutes mounts schema registry routes.
func (h *SchemaRegistryHandlers) RegisterRoutes(rg *gin.RouterGroup) {
	schemas := rg.Group("/schemas")
	{
		schemas.GET("/subjects", h.ListSubjects)
		schemas.GET("/subjects/:subject/versions", h.ListVersions)
		schemas.GET("/subjects/:subject/versions/:version", h.GetSchemaByVersion)
		schemas.POST("/subjects/:subject/versions", h.RegisterSchema)
		schemas.DELETE("/subjects/:subject/versions/:version", h.DeleteSchemaVersion)
		schemas.GET("/ids/:id", h.GetSchemaByID)
		schemas.PUT("/config/:subject", h.SetSubjectCompatibility)
		schemas.GET("/config/:subject", h.GetSubjectCompatibility)
		schemas.POST("/compatibility/subjects/:subject/versions/:version", h.CheckCompatibility)
	}
}

// ListSubjects returns all registered subjects.
func (h *SchemaRegistryHandlers) ListSubjects(c *gin.Context) {
	ctx := c.Request.Context()
	subjects, err := h.subjectStore.List(ctx, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subjects"})
		return
	}

	var names []string
	for _, s := range subjects {
		names = append(names, s.Name)
	}

	c.JSON(http.StatusOK, names)
}

// ListVersions returns all versions for a subject.
func (h *SchemaRegistryHandlers) ListVersions(c *gin.Context) {
	subject := c.Param("subject")
	ctx := c.Request.Context()

	schemas, err := h.schemaStore.List(ctx, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schemas"})
		return
	}

	var versions []int
	for _, s := range schemas {
		if s.Spec.Subject == subject && s.Status.Version > 0 {
			versions = append(versions, s.Status.Version)
		}
	}

	if len(versions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "subject not found", "subject": subject})
		return
	}

	c.JSON(http.StatusOK, versions)
}

// GetSchemaByVersion returns a specific schema version.
func (h *SchemaRegistryHandlers) GetSchemaByVersion(c *gin.Context) {
	subject := c.Param("subject")
	versionStr := c.Param("version")
	ctx := c.Request.Context()

	var targetVersion int
	if versionStr == "latest" {
		targetVersion = -1 // Will find latest
	} else {
		v, err := strconv.Atoi(versionStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version number"})
			return
		}
		targetVersion = v
	}

	schemas, err := h.schemaStore.List(ctx, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schemas"})
		return
	}

	var found *SchemaResource
	for _, s := range schemas {
		if s.Spec.Subject != subject {
			continue
		}
		if targetVersion == -1 {
			if found == nil || s.Status.Version > found.Status.Version {
				found = s
			}
		} else if s.Status.Version == targetVersion {
			found = s
			break
		}
	}

	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subject":    found.Spec.Subject,
		"version":    found.Status.Version,
		"id":         found.Status.SchemaID,
		"schemaType": found.Spec.SchemaType,
		"schema":     found.Spec.Schema,
		"references": found.Spec.References,
	})
}

// RegisterSchema registers a new schema version for a subject.
func (h *SchemaRegistryHandlers) RegisterSchema(c *gin.Context) {
	subject := c.Param("subject")

	var req struct {
		SchemaType string            `json:"schemaType"`
		Schema     string            `json:"schema"`
		References []SchemaReference `json:"references,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
		return
	}

	if req.Schema == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schema field is required"})
		return
	}

	schemaType := SchemaType(req.SchemaType)
	if schemaType == "" {
		schemaType = SchemaTypeJSON
	}

	now := time.Now()
	schemaName := subject + "-" + uuid.New().String()[:8]

	schema := &SchemaResource{
		TypeMeta: resources.TypeMeta{
			APIVersion: SchemaAPIVersion,
			Kind:       SchemaKind,
		},
		ObjectMeta: resources.ObjectMeta{
			Name:       schemaName,
			UID:        uuid.New().String(),
			Generation: 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		Spec: SchemaSpec{
			Subject:    subject,
			SchemaType: schemaType,
			Schema:     req.Schema,
			References: req.References,
		},
	}

	ctx := c.Request.Context()

	// Ensure subject exists.
	_, err := h.subjectStore.Get(ctx, subject)
	if err != nil {
		// Auto-create subject.
		subj := &SchemaSubjectResource{
			TypeMeta: resources.TypeMeta{
				APIVersion: SubjectAPIVersion,
				Kind:       SubjectKind,
			},
			ObjectMeta: resources.ObjectMeta{
				Name:       subject,
				UID:        uuid.New().String(),
				Generation: 1,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			Spec: SubjectSpec{
				Compatibility: CompatBackward,
				SchemaType:    schemaType,
			},
		}
		_ = h.subjectStore.Create(ctx, subj)
	}

	// Create schema resource — reconciler will validate compatibility.
	if err := h.schemaStore.Create(ctx, schema); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register schema", "detail": err.Error()})
		return
	}

	// Return 202 — reconciler will process compatibility.
	c.JSON(http.StatusAccepted, gin.H{
		"name":    schemaName,
		"subject": subject,
		"message": "schema submitted for compatibility validation",
	})
}

// DeleteSchemaVersion soft-deletes a schema version.
func (h *SchemaRegistryHandlers) DeleteSchemaVersion(c *gin.Context) {
	subject := c.Param("subject")
	versionStr := c.Param("version")
	ctx := c.Request.Context()

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version number"})
		return
	}

	schemas, err := h.schemaStore.List(ctx, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schemas"})
		return
	}

	for _, s := range schemas {
		if s.Spec.Subject == subject && s.Status.Version == version {
			if err := h.schemaStore.Delete(ctx, s.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete schema"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"version": version, "deleted": true})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "schema version not found"})
}

// GetSchemaByID returns a schema by its global ID.
func (h *SchemaRegistryHandlers) GetSchemaByID(c *gin.Context) {
	idStr := c.Param("id")
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema ID"})
		return
	}

	schemas, err := h.schemaStore.List(ctx, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search schemas"})
		return
	}

	for _, s := range schemas {
		if s.Status.SchemaID == id {
			c.JSON(http.StatusOK, gin.H{
				"schema":     s.Spec.Schema,
				"schemaType": s.Spec.SchemaType,
				"subject":    s.Spec.Subject,
				"version":    s.Status.Version,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "schema not found", "id": id})
}

// SetSubjectCompatibility updates the compatibility mode for a subject.
func (h *SchemaRegistryHandlers) SetSubjectCompatibility(c *gin.Context) {
	subject := c.Param("subject")
	ctx := c.Request.Context()

	var req struct {
		Compatibility string `json:"compatibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	subj, err := h.subjectStore.Get(ctx, subject)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subject not found"})
		return
	}

	subj.Spec.Compatibility = CompatibilityMode(req.Compatibility)
	subj.Generation++
	subj.UpdatedAt = time.Now()

	if err := h.subjectStore.Update(ctx, subj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subject"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"compatibility": req.Compatibility})
}

// GetSubjectCompatibility returns the compatibility mode for a subject.
func (h *SchemaRegistryHandlers) GetSubjectCompatibility(c *gin.Context) {
	subject := c.Param("subject")
	ctx := c.Request.Context()

	subj, err := h.subjectStore.Get(ctx, subject)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subject not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"compatibility": subj.Spec.Compatibility,
	})
}

// CheckCompatibility tests if a schema is compatible without registering it.
func (h *SchemaRegistryHandlers) CheckCompatibility(c *gin.Context) {
	subject := c.Param("subject")
	versionStr := c.Param("version")
	ctx := c.Request.Context()

	var req struct {
		SchemaType string `json:"schemaType"`
		Schema     string `json:"schema"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Find the target version to check against.
	var targetVersion int
	if versionStr == "latest" {
		targetVersion = -1
	} else {
		v, err := strconv.Atoi(versionStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}
		targetVersion = v
	}

	schemas, err := h.schemaStore.List(ctx, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schemas"})
		return
	}

	var target *SchemaResource
	for _, s := range schemas {
		if s.Spec.Subject != subject {
			continue
		}
		if targetVersion == -1 {
			if target == nil || s.Status.Version > target.Status.Version {
				target = s
			}
		} else if s.Status.Version == targetVersion {
			target = s
			break
		}
	}

	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target schema version not found"})
		return
	}

	// Get compatibility mode.
	compatMode := CompatBackward
	subj, err := h.subjectStore.Get(ctx, subject)
	if err == nil {
		compatMode = subj.Spec.Compatibility
	}

	// Run check.
	schemaType := SchemaType(req.SchemaType)
	if schemaType == "" {
		schemaType = SchemaTypeJSON
	}

	errors := h.checker.CheckCompatibility(req.Schema, target.Spec.Schema, schemaType, compatMode)

	c.JSON(http.StatusOK, gin.H{
		"is_compatible": len(errors) == 0,
		"errors":        errors,
	})
}
