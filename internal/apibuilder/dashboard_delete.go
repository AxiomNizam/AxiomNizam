package apibuilder

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteDashboard removes a generated analytics dashboard
func (h *APIBuilderHandler) DeleteDashboard(c *gin.Context) {
	dashID := c.Param("id")

	h.analyticsHandler.mu.Lock()
	_, ok := h.analyticsHandler.dashboards[dashID]
	if !ok {
		h.analyticsHandler.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	delete(h.analyticsHandler.dashboards, dashID)
	h.analyticsHandler.mu.Unlock()

	// Clear references in CSV uploads
	h.mu.Lock()
	for _, u := range h.csvUploads {
		if u.DashboardID == dashID {
			u.DashboardID = ""
			if u.Status == "dashboard_created" {
				u.Status = "analyzed"
			}
		}
	}
	delete(h.generatedDashboards, dashID)
	h.persistStateLocked()
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "dashboard deleted", "id": dashID})
}
