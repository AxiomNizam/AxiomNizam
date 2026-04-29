package mlpipeline

import (
	"net/http"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"github.com/gin-gonic/gin"
)

type MLPipelineHandlers struct {
	pipelineStore   store.ResourceStore[*MLPipelineResource]
	deploymentStore store.ResourceStore[*ModelDeploymentResource]
}

func NewMLPipelineHandlers(
	pipelineStore store.ResourceStore[*MLPipelineResource],
	deploymentStore store.ResourceStore[*ModelDeploymentResource],
) *MLPipelineHandlers {
	return &MLPipelineHandlers{pipelineStore: pipelineStore, deploymentStore: deploymentStore}
}

func (h *MLPipelineHandlers) RegisterRoutes(rg *gin.RouterGroup) {
	ml := rg.Group("/ml")
	{
		// Pipelines
		ml.GET("/pipelines", h.ListPipelines)
		ml.GET("/pipelines/:name", h.GetPipeline)
		ml.POST("/pipelines", h.CreatePipeline)
		ml.PUT("/pipelines/:name", h.UpdatePipeline)
		ml.DELETE("/pipelines/:name", h.DeletePipeline)
		ml.POST("/pipelines/:name/run", h.TriggerRun)

		// Deployments
		ml.GET("/deployments", h.ListDeployments)
		ml.GET("/deployments/:name", h.GetDeployment)
		ml.POST("/deployments", h.CreateDeployment)
		ml.DELETE("/deployments/:name", h.DeleteDeployment)
	}
}

func (h *MLPipelineHandlers) ListPipelines(c *gin.Context) {
	pipelines, err := h.pipelineStore.List(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pipelines": pipelines, "count": len(pipelines)})
}

func (h *MLPipelineHandlers) GetPipeline(c *gin.Context) {
	name := c.Param("name")
	pipeline, err := h.pipelineStore.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, pipeline)
}

func (h *MLPipelineHandlers) CreatePipeline(c *gin.Context) {
	var pipeline MLPipelineResource
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pipeline.Kind = MLPipelineKind
	pipeline.APIVersion = MLPipelineAPIVersion
	now := time.Now()
	pipeline.CreatedAt = now
	pipeline.Generation = 1
	pipeline.Status.Phase = "Pending"
	pipeline.Status.PipelineStatus = "pending"
	if err := h.pipelineStore.Create(c.Request.Context(), &pipeline); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pipeline)
}

func (h *MLPipelineHandlers) UpdatePipeline(c *gin.Context) {
	name := c.Param("name")
	existing, err := h.pipelineStore.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found", "name": name})
		return
	}
	var updated MLPipelineResource
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated.ObjectMeta = existing.ObjectMeta
	updated.Generation = existing.Generation + 1
	updated.Status = existing.Status
	if err := h.pipelineStore.Update(c.Request.Context(), &updated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *MLPipelineHandlers) DeletePipeline(c *gin.Context) {
	name := c.Param("name")
	if err := h.pipelineStore.Delete(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": name})
}

func (h *MLPipelineHandlers) TriggerRun(c *gin.Context) {
	name := c.Param("name")
	pipeline, err := h.pipelineStore.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found", "name": name})
		return
	}
	pipeline.Generation++
	// Reset step statuses for a fresh run.
	pipeline.Status.StepStatuses = nil
	pipeline.Status.PipelineStatus = "pending"
	_ = h.pipelineStore.Update(c.Request.Context(), pipeline)
	c.JSON(http.StatusAccepted, gin.H{"message": "pipeline run triggered", "pipeline": name})
}

// --- Deployment Handlers ---

func (h *MLPipelineHandlers) ListDeployments(c *gin.Context) {
	deployments, err := h.deploymentStore.List(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deployments": deployments, "count": len(deployments)})
}

func (h *MLPipelineHandlers) GetDeployment(c *gin.Context) {
	name := c.Param("name")
	deployment, err := h.deploymentStore.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, deployment)
}

func (h *MLPipelineHandlers) CreateDeployment(c *gin.Context) {
	var deployment ModelDeploymentResource
	if err := c.ShouldBindJSON(&deployment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	deployment.Kind = ModelDeploymentKind
	deployment.APIVersion = ModelDeploymentAPIVersion
	now := time.Now()
	deployment.CreatedAt = now
	deployment.Generation = 1
	deployment.Status.Phase = "Pending"
	deployment.Status.DeploymentStatus = "pending"
	if err := h.deploymentStore.Create(c.Request.Context(), &deployment); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, deployment)
}

func (h *MLPipelineHandlers) DeleteDeployment(c *gin.Context) {
	name := c.Param("name")
	if err := h.deploymentStore.Delete(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": name})
}
