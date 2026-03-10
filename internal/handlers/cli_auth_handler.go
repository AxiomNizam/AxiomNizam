package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// CLIAuthHandler handles authentication requests from the CLI
type CLIAuthHandler struct {
	mu     sync.RWMutex
	users  map[string]*CLIUser // username -> user
	tokens map[string]string   // token -> username
	jwtKey []byte
}

// CLIUser represents a user for CLI auth
type CLIUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

// NewCLIAuthHandler creates a new CLI auth handler with default admin user
func NewCLIAuthHandler() *CLIAuthHandler {
	h := &CLIAuthHandler{
		users:  make(map[string]*CLIUser),
		tokens: make(map[string]string),
		jwtKey: []byte("axiom-nizam-secret-key-change-in-production"),
	}

	// Register default admin user
	h.users["admin"] = &CLIUser{
		ID:       uuid.New().String(),
		Name:     "Admin",
		Email:    "admin@axiom-nizam.io",
		Username: "admin",
		Password: "admin",
		Role:     "admin",
	}

	return h
}

// Login handles CLI login requests
func (h *CLIAuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	h.mu.RLock()
	user, exists := h.users[req.Username]
	h.mu.RUnlock()

	if !exists || user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	h.mu.Lock()
	h.tokens[tokenString] = user.Username
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"expiresAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"user": gin.H{
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// Verify verifies an API key or token
func (h *CLIAuthHandler) Verify(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	h.mu.RLock()
	username, exists := h.tokens[tokenStr]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	h.mu.RLock()
	user := h.users[username]
	h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// WhoAmI returns current user info
func (h *CLIAuthHandler) WhoAmI(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	h.mu.RLock()
	username, exists := h.tokens[tokenStr]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	h.mu.RLock()
	user := h.users[username]
	h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}
