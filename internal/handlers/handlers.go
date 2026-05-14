package handlers

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"time"

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
	conns      *database.Connections
	backendMgr interface{} // *platformstore.BackendManager or nil
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(conns *database.Connections) *HealthHandler {
	return &HealthHandler{conns: conns}
}

// SetBackendManager wires the storage backend manager for /distributed endpoint.
func (h *HealthHandler) SetBackendManager(bm interface{}) {
	h.backendMgr = bm
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

// raftBackendInfo is a local interface to avoid importing platform/store.
type raftBackendInfo interface {
	IsRaft() bool
	IsEtcd() bool
}

// Distributed handles GET /distributed - Check cluster status (Raft or etcd)
func (h *HealthHandler) Distributed(c *gin.Context) {
	// Try Raft backend first.
	if h.backendMgr != nil {
		if bm, ok := h.backendMgr.(raftBackendInfo); ok && bm.IsRaft() {
			h.distributedRaft(c)
			return
		}
	}

	// Fallback: etcd via etcdctl.
	h.distributedEtcd(c)
}

// distributedRaft reports Raft cluster status (leader, state, peers).
//
// Leader/state are read from non-blocking atomics (instant).  Full stats
// and peer list are fetched in a goroutine with a 3-second safety timeout.
// If the Raft main loop is saturated, falls back to QuickStatus (atomic).
func (h *HealthHandler) distributedRaft(c *gin.Context) {
	status := map[string]interface{}{
		"backend":        "raft",
		"is_distributed": true,
		"healthy":        true,
	}

	type raftInfo interface {
		GetRaftLeader() (string, string)
		GetRaftIsLeader() bool
		GetRaftQuickStatus() map[string]string
		GetRaftStats() map[string]string
		GetRaftPeers() ([]map[string]string, error)
	}

	if rf, ok := h.backendMgr.(raftInfo); ok {
		// Leader info — non-blocking atomic reads, always instant.
		leaderAddr, leaderID := rf.GetRaftLeader()
		isLeader := rf.GetRaftIsLeader()
		if isLeader {
			status["node_state"] = "leader"
		} else {
			status["node_state"] = "follower"
		}
		status["leader_addr"] = leaderAddr
		status["leader_id"] = leaderID
		status["is_leader"] = isLeader

		// Full stats + peers in a goroutine with safety timeout.
		// Now that the Redis middleware bug is fixed, this should
		// complete in milliseconds.  The timeout is a safety net.
		type raftResult struct {
			stats map[string]string
			peers []map[string]string
		}
		done := make(chan raftResult, 1)
		go func() {
			var r raftResult
			r.stats = rf.GetRaftStats()
			if p, err := rf.GetRaftPeers(); err == nil {
				r.peers = p
			}
			done <- r
		}()

		timer := time.NewTimer(3 * time.Second)
		select {
		case r := <-done:
			timer.Stop()
			status["stats"] = r.stats
			if r.peers != nil {
				status["peers"] = r.peers
				status["member_count"] = len(r.peers)
			}
		case <-timer.C:
			// Fallback: non-blocking atomic stats only.
			status["stats"] = rf.GetRaftQuickStatus()
			status["stats_note"] = "full stats timed out, showing atomic snapshot"
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "Raft cluster status",
		Data:    status,
	})
}

// distributedEtcd reports etcd cluster status (legacy).
func (h *HealthHandler) distributedEtcd(c *gin.Context) {
	distributedStatus := map[string]interface{}{
		"backend":        "etcd",
		"is_distributed": false,
		"members":        []string{},
		"leader":         "",
		"healthy":        false,
		"error":          nil,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "etcdctl", "--endpoints=localhost:2379", "member", "list")
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

	healthCmd := exec.CommandContext(ctx, "etcdctl", "--endpoints=localhost:2379", "endpoint", "health")
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

