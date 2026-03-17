package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	mu           sync.RWMutex
	users        map[string]*PlatformUser
	keycloakSync *keycloakUserSync
	etcd         *clientv3.Client
	stateKey     string
}

// NewPlatformUserHandler creates a new platform user handler
func NewPlatformUserHandler(etcd *clientv3.Client) *PlatformUserHandler {
	h := &PlatformUserHandler{
		users:        make(map[string]*PlatformUser),
		keycloakSync: newKeycloakUserSync(),
		etcd:         etcd,
		stateKey:     "axiomnizam:platform:users",
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
		log.Printf("platform-users: failed to load persisted state from etcd: %v", err)
		return
	}
	if len(resp.Kvs) == 0 {
		return
	}

	var users map[string]*PlatformUser
	if err := json.Unmarshal(resp.Kvs[0].Value, &users); err != nil {
		log.Printf("platform-users: failed to decode persisted state: %v", err)
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
		log.Printf("platform-users: failed to encode state: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.etcd.Put(ctx, h.stateKey, string(payload)); err != nil {
		log.Printf("platform-users: failed to persist state to etcd: %v", err)
	}
}

type keycloakSyncError struct {
	StatusCode int
	Message    string
}

func (e *keycloakSyncError) Error() string {
	return e.Message
}

type keycloakUserSync struct {
	baseURL       string
	targetRealm   string
	adminRealm    string
	adminClientID string
	adminUsername string
	adminPassword string
	httpClient    *http.Client
}

func newKeycloakUserSync() *keycloakUserSync {
	enabled := strings.EqualFold(getEnv("KEYCLOAK_USER_SYNC_ENABLED", "true"), "true")
	if !enabled {
		return nil
	}

	host := getEnv("KEYCLOAK_HOST", "keycloak")
	port := getEnv("KEYCLOAK_PORT", "8080")
	adminUsername := getEnv("KEYCLOAK_ADMIN_USERNAME", getEnv("KEYCLOAK_ADMIN", "admin"))
	adminPassword := getEnv("KEYCLOAK_ADMIN_PASSWORD", "")

	if strings.TrimSpace(adminPassword) == "" {
		log.Printf("⚠️  KEYCLOAK_ADMIN_PASSWORD is empty; Keycloak user sync disabled")
		return nil
	}

	return &keycloakUserSync{
		baseURL:       fmt.Sprintf("http://%s:%s", host, port),
		targetRealm:   getEnv("KEYCLOAK_REALM", "axiomnizam"),
		adminRealm:    getEnv("KEYCLOAK_ADMIN_REALM", "master"),
		adminClientID: getEnv("KEYCLOAK_ADMIN_CLIENT_ID", "admin-cli"),
		adminUsername: adminUsername,
		adminPassword: adminPassword,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (k *keycloakUserSync) getAdminAccessToken() (string, error) {
	form := url.Values{}
	form.Set("client_id", k.adminClientID)
	form.Set("grant_type", "password")
	form.Set("username", k.adminUsername)
	form.Set("password", k.adminPassword)

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.baseURL, url.PathEscape(k.adminRealm))
	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get keycloak admin token: %s", strings.TrimSpace(string(body)))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", fmt.Errorf("keycloak admin token missing in response")
	}

	return tokenResp.AccessToken, nil
}

func (k *keycloakUserSync) doJSON(method, endpoint, bearerToken string, payload interface{}) (int, []byte, http.Header, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, resp.Header, nil
}

func extractKeycloakUserIDFromLocation(location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return ""
	}
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func (k *keycloakUserSync) findUserID(token, username string) (string, error) {
	endpoint := fmt.Sprintf("%s/admin/realms/%s/users?username=%s&exact=true", k.baseURL, url.PathEscape(k.targetRealm), url.QueryEscape(username))
	status, body, _, err := k.doJSON(http.MethodGet, endpoint, token, nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("failed to query keycloak user: %s", strings.TrimSpace(string(body)))
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(body, &users); err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", fmt.Errorf("user not found in keycloak after create")
	}
	id, _ := users[0]["id"].(string)
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("invalid keycloak user id")
	}
	return id, nil
}

func (k *keycloakUserSync) ensureRealmRole(token, role string) error {
	endpoint := fmt.Sprintf("%s/admin/realms/%s/roles", k.baseURL, url.PathEscape(k.targetRealm))
	status, body, _, err := k.doJSON(http.MethodPost, endpoint, token, map[string]interface{}{"name": role})
	if err != nil {
		return err
	}
	if status == http.StatusCreated || status == http.StatusConflict || status == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("failed to ensure role '%s': %s", role, strings.TrimSpace(string(body)))
}

func (k *keycloakUserSync) getRealmRole(token, role string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/admin/realms/%s/roles/%s", k.baseURL, url.PathEscape(k.targetRealm), url.PathEscape(role))
	status, body, _, err := k.doJSON(http.MethodGet, endpoint, token, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch role '%s': %s", role, strings.TrimSpace(string(body)))
	}

	var roleRep map[string]interface{}
	if err := json.Unmarshal(body, &roleRep); err != nil {
		return nil, err
	}
	return roleRep, nil
}

func (k *keycloakUserSync) assignRealmRole(token, userID, role string) error {
	if strings.TrimSpace(role) == "" {
		return nil
	}
	if err := k.ensureRealmRole(token, role); err != nil {
		return err
	}

	roleRep, err := k.getRealmRole(token, role)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", k.baseURL, url.PathEscape(k.targetRealm), url.PathEscape(userID))
	status, body, _, err := k.doJSON(http.MethodPost, endpoint, token, []map[string]interface{}{roleRep})
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("failed to assign role '%s': %s", role, strings.TrimSpace(string(body)))
	}
	return nil
}

func (k *keycloakUserSync) listUserRealmRoles(token, userID string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", k.baseURL, url.PathEscape(k.targetRealm), url.PathEscape(userID))
	status, body, _, err := k.doJSON(http.MethodGet, endpoint, token, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("failed to list user realm roles: %s", strings.TrimSpace(string(body)))
	}

	var roles []map[string]interface{}
	if len(body) == 0 {
		return roles, nil
	}
	if err := json.Unmarshal(body, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (k *keycloakUserSync) removeRealmRoles(token, userID string, roles []map[string]interface{}) error {
	if len(roles) == 0 {
		return nil
	}

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", k.baseURL, url.PathEscape(k.targetRealm), url.PathEscape(userID))
	status, body, _, err := k.doJSON(http.MethodDelete, endpoint, token, roles)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("failed to remove realm roles: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func isManagedPlatformRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "manager", "user", "system-manager", "sysadmin", "system_admin", "system-admin":
		return true
	default:
		return false
	}
}

func (k *keycloakUserSync) UpdateUserRole(username, role string) error {
	if k == nil {
		return nil
	}

	normalizedRole := strings.ToLower(strings.TrimSpace(role))
	if normalizedRole == "" {
		return nil
	}

	token, err := k.getAdminAccessToken()
	if err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "failed to authenticate with keycloak admin API"}
	}

	userID, err := k.findUserID(token, username)
	if err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "failed to locate keycloak user for role update"}
	}

	existingRoles, err := k.listUserRealmRoles(token, userID)
	if err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "failed to read keycloak user roles"}
	}

	toRemove := make([]map[string]interface{}, 0)
	for _, roleRep := range existingRoles {
		name, _ := roleRep["name"].(string)
		if isManagedPlatformRole(name) {
			toRemove = append(toRemove, roleRep)
		}
	}

	if err := k.removeRealmRoles(token, userID, toRemove); err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "failed to remove previous keycloak role mappings"}
	}

	if err := k.assignRealmRole(token, userID, normalizedRole); err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "failed to assign updated keycloak role mapping"}
	}

	return nil
}

func (k *keycloakUserSync) CreateUser(username, email, password, role string) error {
	if k == nil {
		return nil
	}

	firstName, lastName := splitDisplayName(username)

	token, err := k.getAdminAccessToken()
	if err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "failed to authenticate with keycloak admin API"}
	}

	payload := map[string]interface{}{
		"username":      username,
		"email":         email,
		"firstName":     firstName,
		"lastName":      lastName,
		"enabled":       true,
		"emailVerified": true,
		"credentials": []map[string]interface{}{
			{
				"type":      "password",
				"value":     password,
				"temporary": false,
			},
		},
	}

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users", k.baseURL, url.PathEscape(k.targetRealm))
	status, body, headers, err := k.doJSON(http.MethodPost, endpoint, token, payload)
	if err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "failed to connect to keycloak admin API"}
	}

	if status == http.StatusConflict {
		return &keycloakSyncError{StatusCode: http.StatusConflict, Message: "user already exists in keycloak"}
	}
	if status != http.StatusCreated {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = "keycloak user creation failed"
		}
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: message}
	}

	userID := extractKeycloakUserIDFromLocation(headers.Get("Location"))
	if userID == "" {
		foundID, findErr := k.findUserID(token, username)
		if findErr != nil {
			return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "user created in keycloak but user id lookup failed"}
		}
		userID = foundID
	}

	if err := k.assignRealmRole(token, userID, strings.ToLower(strings.TrimSpace(role))); err != nil {
		return &keycloakSyncError{StatusCode: http.StatusBadGateway, Message: "user created in keycloak but role assignment failed: " + err.Error()}
	}

	return nil
}

func splitDisplayName(username string) (string, string) {
	clean := strings.TrimSpace(username)
	if clean == "" {
		return "User", "Account"
	}

	parts := strings.FieldsFunc(clean, func(r rune) bool {
		switch r {
		case '.', '-', '_', ' ':
			return true
		default:
			return false
		}
	})

	if len(parts) == 0 {
		return clean, "Account"
	}
	if len(parts) == 1 {
		return parts[0], "Account"
	}

	first := strings.TrimSpace(parts[0])
	last := strings.TrimSpace(strings.Join(parts[1:], " "))
	if first == "" {
		first = "User"
	}
	if last == "" {
		last = "Account"
	}
	return first, last
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
	if h.keycloakSync != nil {
		if err := h.keycloakSync.CreateUser(req.Username, req.Email, req.Password, req.Role); err != nil {
			var syncErr *keycloakSyncError
			if errors.As(err, &syncErr) {
				status := http.StatusBadGateway
				if syncErr.StatusCode == http.StatusConflict {
					status = http.StatusConflict
				}
				c.JSON(status, gin.H{"status": "error", "error": syncErr.Message})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": "failed to provision keycloak user: " + err.Error()})
			return
		}
	}

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
	h.persistStateLocked()
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

	normalizedEmail := strings.TrimSpace(req.Email)
	normalizedRole := ""
	if req.Role != "" {
		role := strings.ToLower(strings.TrimSpace(req.Role))
		if !validPlatformRoles[role] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid role. Must be admin, manager, or user"})
			return
		}
		normalizedRole = role
	}

	normalizedStatus := ""
	if req.Status != "" {
		status := strings.ToLower(strings.TrimSpace(req.Status))
		if status != "active" && status != "disabled" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid status. Must be active or disabled"})
			return
		}
		normalizedStatus = status
	}

	h.mu.RLock()
	user, exists := h.users[id]
	h.mu.RUnlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "User not found"})
		return
	}

	if normalizedRole != "" && h.keycloakSync != nil {
		if err := h.keycloakSync.UpdateUserRole(user.Username, normalizedRole); err != nil {
			var syncErr *keycloakSyncError
			if errors.As(err, &syncErr) {
				c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": syncErr.Message})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": "failed to sync updated role to keycloak: " + err.Error()})
			return
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	user, exists = h.users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "User not found"})
		return
	}

	if req.Email != "" {
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
	h.persistStateLocked()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("User '%s' deleted successfully", username),
	})
}

// ValidateCredentials checks username+password against platform users.
// Returns the matched user and true on success, nil and false otherwise.
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
