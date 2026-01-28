package handlers

import (
	"net/http"
	"os/exec"
	"strings"

	"example.com/axiomnizam/internal/database"
	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler handles user CRUD operations
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	result := h.db.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Status:  "ok",
		Message: "User created successfully",
		Data:    user,
	})
}

// GetAllUsers handles GET /users
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	var users []models.User
	result := h.db.Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

// GetUserByID handles GET /users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	id := c.Param("id")
	var user models.User
	result := h.db.First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Status: "error",
			Error:  "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	id := c.Param("id")
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	result := h.db.Where("id = ?", id).Updates(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "User updated successfully",
		Data:    user,
	})
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	id := c.Param("id")
	result := h.db.Delete(&models.User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "User deleted successfully",
	})
}

// HealthHandler handles health check
type HealthHandler struct {
	conns *database.Connections
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(conns *database.Connections) *HealthHandler {
	return &HealthHandler{conns: conns}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "AxiomNizam API is running",
	})
}

// Status handles GET /status
func (h *HealthHandler) Status(c *gin.Context) {
	status := map[string]string{}
	connected := h.conns.IsConnected()

	for db, isConnected := range connected {
		if isConnected {
			status[db] = "connected"
		} else {
			status[db] = "disconnected"
		}
	}

	// Firebase and Oracle are emulated services, always show as available
	status["firebase"] = "connected"
	status["oracle"] = "connected"

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "System status",
		Data:    status,
	})
}

// Distributed handles GET /distributed - Check if system is running in distributed mode
func (h *HealthHandler) Distributed(c *gin.Context) {
	distributedStatus := map[string]interface{}{
		"is_distributed": false,
		"members":        []string{},
		"leader":         "",
		"healthy":        false,
		"error":          nil,
	}

	cmd := exec.Command("etcdctl", "--endpoints=localhost:2379", "member", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		distributedStatus["error"] = "etcdctl not available or etcd not running"
		c.JSON(http.StatusOK, models.Response{
			Status:  "ok",
			Message: "Distributed status check",
			Data:    distributedStatus,
		})
		return
	}

	members := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			members = append(members, line)
		}
	}

	if len(members) > 0 {
		distributedStatus["is_distributed"] = true
		distributedStatus["members"] = members
		distributedStatus["member_count"] = len(members)
	}

	healthCmd := exec.Command("etcdctl", "--endpoints=localhost:2379", "endpoint", "health")
	healthOutput, healthErr := healthCmd.CombinedOutput()

	if healthErr == nil {
		distributedStatus["healthy"] = true
		distributedStatus["health_info"] = strings.TrimSpace(string(healthOutput))
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "Distributed status check",
		Data:    distributedStatus,
	})
}
