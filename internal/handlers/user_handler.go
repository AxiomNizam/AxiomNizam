package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// PlatformUser represents a platform user with role-based access
type PlatformUser struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`      // never expose in JSON
	Role      string    `json:"role"`   // admin, manager, user
	Status    string    `json:"status"` // active, disabled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePlatformUserRequest is the request body for creating a platform user
type CreatePlatformUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

// UpdatePlatformUserRequest is the request body for updating a platform user
type UpdatePlatformUserRequest struct {
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// PlatformUserHandler manages platform user CRUD operations
type PlatformUserHandler struct {
	mu    sync.RWMutex
	users map[string]*PlatformUser
}

// NewPlatformUserHandler creates a new platform user handler
func NewPlatformUserHandler() *PlatformUserHandler {
	return &PlatformUserHandler{
		users: make(map[string]*PlatformUser),
	}
}

func generateUserID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

var validPlatformRoles = map[string]bool{"admin": true, "manager": true, "user": true}

// ListPlatformUsers returns all platform users
func (h *PlatformUserHandler) ListPlatformUsers(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]*PlatformUser, 0, len(h.users))
	for _, u := range h.users {
		users = append(users, u)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"users":  users,
		"count":  len(users),
	})
}

// GetPlatformUser returns a single platform user by ID
func (h *PlatformUserHandler) GetPlatformUser(c *gin.Context) {
	id := c.Param("id")

	h.mu.RLock()
	user, exists := h.users[id]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "user": user})
}

// CreatePlatformUser creates a new platform user
func (h *PlatformUserHandler) CreatePlatformUser(c *gin.Context) {
	var req CreatePlatformUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))

	if !validPlatformRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid role. Must be admin, manager, or user"})
		return
	}

	// Check for duplicate username
	h.mu.RLock()
	for _, u := range h.users {
		if strings.EqualFold(u.Username, req.Username) {
			h.mu.RUnlock()
			c.JSON(http.StatusConflict, gin.H{"status": "error", "error": "Username already exists"})
			return
		}
	}
	h.mu.RUnlock()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Failed to process password"})
		return
	}

	now := time.Now()
	user := &PlatformUser{
		ID:        generateUserID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      req.Role,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.mu.Lock()
	h.users[user.ID] = user
	h.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("User '%s' created successfully", user.Username),
		"user":    user,
	})
}

// UpdatePlatformUser updates an existing platform user
func (h *PlatformUserHandler) UpdatePlatformUser(c *gin.Context) {
	id := c.Param("id")

	var req UpdatePlatformUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	user, exists := h.users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "User not found"})
		return
	}

	if req.Email != "" {
		user.Email = strings.TrimSpace(req.Email)
	}
	if req.Role != "" {
		role := strings.ToLower(strings.TrimSpace(req.Role))
		if !validPlatformRoles[role] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid role. Must be admin, manager, or user"})
			return
		}
		user.Role = role
	}
	if req.Status != "" {
		status := strings.ToLower(strings.TrimSpace(req.Status))
		if status != "active" && status != "disabled" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid status. Must be active or disabled"})
			return
		}
		user.Status = status
	}
	user.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("User '%s' updated successfully", user.Username),
		"user":    user,
	})
}

// DeletePlatformUser deletes a platform user
func (h *PlatformUserHandler) DeletePlatformUser(c *gin.Context) {
	id := c.Param("id")

	h.mu.Lock()
	defer h.mu.Unlock()

	user, exists := h.users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "User not found"})
		return
	}

	username := user.Username
	delete(h.users, id)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("User '%s' deleted successfully", username),
	})
}
