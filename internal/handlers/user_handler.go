package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/logging"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// PlatformUser represents a platform user with role-based access.
type PlatformUser struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePlatformUserRequest is the request body for creating a platform user.
type CreatePlatformUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

// UpdatePlatformUserRequest is the request body for updating a platform user.
type UpdatePlatformUserRequest struct {
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// PlatformUserHandler manages platform user CRUD operations.
type PlatformUserHandler struct {
	mu       sync.RWMutex
	users    map[string]*PlatformUser
	etcd     *clientv3.Client
	stateKey string
}

var validPlatformRoles = map[string]bool{"admin": true, "manager": true, "user": true}

// NewPlatformUserHandler creates a new platform user handler.
func NewPlatformUserHandler(etcd *clientv3.Client) *PlatformUserHandler {
	h := &PlatformUserHandler{
		users:    make(map[string]*PlatformUser),
		etcd:     etcd,
		stateKey: "axiomnizam:platform:users",
	}
	h.loadState()
	return h
}

func (h *PlatformUserHandler) loadState() {
	if h.etcd == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.etcd.Get(ctx, h.stateKey)
	if err != nil {
		logging.Z().Warn("platform-users: failed to load persisted state", zap.Error(err))
		return
	}
	if len(resp.Kvs) == 0 {
		return
	}

	var users map[string]*PlatformUser
	if err := json.Unmarshal(resp.Kvs[0].Value, &users); err != nil {
		logging.Z().Warn("platform-users: failed to decode persisted state", zap.Error(err))
		return
	}
	if users == nil {
		users = make(map[string]*PlatformUser)
	}
	h.users = users
}

func (h *PlatformUserHandler) persistStateLocked() {
	if h.etcd == nil {
		return
	}

	payload, err := json.Marshal(h.users)
	if err != nil {
		logging.Z().Warn("platform-users: failed to encode state", zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.etcd.Put(ctx, h.stateKey, string(payload)); err != nil {
		logging.Z().Warn("platform-users: failed to persist state", zap.Error(err))
	}
}

func generateUserID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// ListPlatformUsers returns all platform users.
func (h *PlatformUserHandler) ListPlatformUsers(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]*PlatformUser, 0, len(h.users))
	for _, u := range h.users {
		users = append(users, u)
	}

	c.JSON(200, gin.H{
		"status": "success",
		"users":  users,
		"count":  len(users),
	})
}

// GetPlatformUser returns a single platform user by ID.
func (h *PlatformUserHandler) GetPlatformUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	h.mu.RLock()
	user, exists := h.users[id]
	h.mu.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"status": "error", "error": "User not found"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "user": user})
}

// CreatePlatformUser creates a new platform user.
func (h *PlatformUserHandler) CreatePlatformUser(c *gin.Context) {
	var req CreatePlatformUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"status": "error", "error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))

	if !validPlatformRoles[req.Role] {
		c.JSON(400, gin.H{"status": "error", "error": "Invalid role. Must be admin, manager, or user"})
		return
	}

	h.mu.RLock()
	for _, u := range h.users {
		if strings.EqualFold(u.Username, req.Username) {
			h.mu.RUnlock()
			c.JSON(409, gin.H{"status": "error", "error": "Username already exists"})
			return
		}
		if strings.EqualFold(u.Email, req.Email) {
			h.mu.RUnlock()
			c.JSON(409, gin.H{"status": "error", "error": "Email already exists"})
			return
		}
	}
	h.mu.RUnlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "error": "Failed to process password"})
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
	h.persistStateLocked()
	h.mu.Unlock()

	c.JSON(201, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("User '%s' created successfully", user.Username),
		"user":    user,
	})
}

// UpdatePlatformUser updates an existing platform user.
func (h *PlatformUserHandler) UpdatePlatformUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	var req UpdatePlatformUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"status": "error", "error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	normalizedEmail := strings.TrimSpace(req.Email)
	normalizedRole := ""
	if req.Role != "" {
		role := strings.ToLower(strings.TrimSpace(req.Role))
		if !validPlatformRoles[role] {
			c.JSON(400, gin.H{"status": "error", "error": "Invalid role. Must be admin, manager, or user"})
			return
		}
		normalizedRole = role
	}

	normalizedStatus := ""
	if req.Status != "" {
		status := strings.ToLower(strings.TrimSpace(req.Status))
		switch status {
		case "active", "enabled":
			normalizedStatus = "active"
		case "disabled", "inactive", "deactive", "deactivated":
			normalizedStatus = "disabled"
		default:
			c.JSON(400, gin.H{"status": "error", "error": "Invalid status. Must be active/disabled (inactive/deactive also accepted)"})
			return
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	user, exists := h.users[id]
	if !exists {
		c.JSON(404, gin.H{"status": "error", "error": "User not found"})
		return
	}

	if normalizedEmail != "" {
		for uid, u := range h.users {
			if uid != id && strings.EqualFold(u.Email, normalizedEmail) {
				c.JSON(409, gin.H{"status": "error", "error": "Email already exists"})
				return
			}
		}
		user.Email = normalizedEmail
	}
	if normalizedRole != "" {
		user.Role = normalizedRole
	}
	if normalizedStatus != "" {
		user.Status = normalizedStatus
	}
	user.UpdatedAt = time.Now()
	h.persistStateLocked()

	c.JSON(200, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("User '%s' updated successfully", user.Username),
		"user":    user,
	})
}

// DeletePlatformUser deletes a platform user.
func (h *PlatformUserHandler) DeletePlatformUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	h.mu.Lock()
	defer h.mu.Unlock()

	user, exists := h.users[id]
	if !exists {
		c.JSON(404, gin.H{"status": "error", "error": "User not found"})
		return
	}

	username := user.Username
	delete(h.users, id)
	h.persistStateLocked()

	c.JSON(200, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("User '%s' deleted successfully", username),
	})
}

// ValidateCredentials checks username+password against platform users.
func (h *PlatformUserHandler) ValidateCredentials(username, password string) (*PlatformUser, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, u := range h.users {
		if strings.EqualFold(u.Username, username) && u.Status == "active" {
			if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err == nil {
				return u, true
			}
			return nil, false
		}
	}
	return nil, false
}

func normalizePlatformUserStatus(status string) string {
	v := strings.ToLower(strings.TrimSpace(status))
	switch v {
	case "active", "enabled":
		return "active"
	case "disabled", "inactive", "deactive", "deactivated":
		return "disabled"
	default:
		return "active"
	}
}

func normalizePlatformUserRole(role string) string {
	v := strings.ToLower(strings.TrimSpace(role))
	if validPlatformRoles[v] {
		return v
	}
	if strings.Contains(v, "admin") {
		return "admin"
	}
	if strings.Contains(v, "manager") {
		return "manager"
	}
	return "user"
}

func randomFederatedPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(b)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// EnsureFederatedUser creates or updates a user record for OAuth/federated login.
// It ensures the user appears in the platform user list so roles can be managed.
func (h *PlatformUserHandler) EnsureFederatedUser(username, email, defaultRole string) (*PlatformUser, error) {
	uname := strings.TrimSpace(username)
	mail := strings.TrimSpace(email)
	if uname == "" && mail == "" {
		return nil, fmt.Errorf("federated user requires username or email")
	}

	if uname == "" && strings.Contains(mail, "@") {
		uname = strings.TrimSpace(strings.Split(mail, "@")[0])
	}
	role := normalizePlatformUserRole(defaultRole)

	h.mu.Lock()
	defer h.mu.Unlock()

	var existing *PlatformUser
	for _, u := range h.users {
		if uname != "" && strings.EqualFold(strings.TrimSpace(u.Username), uname) {
			existing = u
			break
		}
		if mail != "" && strings.EqualFold(strings.TrimSpace(u.Email), mail) {
			existing = u
			break
		}
	}

	now := time.Now()
	if existing != nil {
		changed := false
		if strings.TrimSpace(existing.Username) == "" && uname != "" {
			existing.Username = uname
			changed = true
		}
		if strings.TrimSpace(existing.Email) == "" && mail != "" {
			existing.Email = mail
			changed = true
		}
		if strings.TrimSpace(existing.Role) == "" {
			existing.Role = role
			changed = true
		}
		normalizedStatus := normalizePlatformUserStatus(existing.Status)
		if normalizedStatus != existing.Status {
			existing.Status = normalizedStatus
			changed = true
		}
		if changed {
			existing.UpdatedAt = now
			h.persistStateLocked()
		}

		copyUser := *existing
		return &copyUser, nil
	}

	hashedPassword, err := randomFederatedPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare federated password: %w", err)
	}

	user := &PlatformUser{
		ID:        generateUserID(),
		Username:  uname,
		Email:     mail,
		Password:  hashedPassword,
		Role:      role,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.users[user.ID] = user
	h.persistStateLocked()

	copyUser := *user
	return &copyUser, nil
}
