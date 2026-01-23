package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"example.com/axiomnizam/internal/auth"
	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	keycloakURL    string
	keycloakRealm  string
	keycloakClient string
	clientSecret   string
	rateLimiter    *auth.RateLimiter
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		keycloakURL:    getEnv("KEYCLOAK_URL", "http://keycloak:8080"),
		keycloakRealm:  getEnv("KEYCLOAK_REALM", "axiomnizam"),
		keycloakClient: getEnv("KEYCLOAK_CLIENT", "axiomnizam-backend"),
		clientSecret:   getEnv("KEYCLOAK_CLIENT_SECRET", "6rFrY3rcyfEma3C5Vj7xCELT7uxFtk72"),
		rateLimiter:    nil, // Will be set via SetRateLimiter
	}
}

// SetRateLimiter sets the rate limiter for the auth handler
func (h *AuthHandler) SetRateLimiter(limiter *auth.RateLimiter) {
	h.rateLimiter = limiter
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// LoginRequest is the request payload for login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse is the response from Keycloak
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// Login handles POST /auth/login
// This endpoint proxies the authentication request to Keycloak
// so the client secret never leaves the backend
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	// Prepare the Keycloak token request
	tokenURL := h.keycloakURL + "/realms/" + h.keycloakRealm + "/protocol/openid-connect/token"
	log.Printf("📝 Login attempt for user: %s, token URL: %s\n", req.Username, tokenURL)

	// Use form-urlencoded body for token request
	body := url.Values{}
	body.Add("client_id", h.keycloakClient)
	body.Add("client_secret", h.clientSecret)
	body.Add("grant_type", "password")
	body.Add("username", req.Username)
	body.Add("password", req.Password)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request to Keycloak
	resp, err := client.Post(
		tokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		log.Printf("❌ Keycloak connection error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to connect to authentication service: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to read authentication response: " + err.Error(),
		})
		return
	}

	log.Printf("📋 Keycloak response status: %d\n", resp.StatusCode)
	log.Printf("📋 Keycloak response body: %s\n", string(responseBody))

	// Parse token response
	var tokenResp TokenResponse
	if err := json.Unmarshal(responseBody, &tokenResp); err != nil {
		log.Printf("❌ Failed to parse JSON: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to parse authentication response: " + err.Error(),
		})
		return
	}

	// Check if Keycloak returned an error
	if tokenResp.Error != "" {
		log.Printf("❌ Keycloak auth error: %s - %s\n", tokenResp.Error, tokenResp.ErrorDesc)
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Authentication failed: %s", tokenResp.ErrorDesc),
		})
		return
	}

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Keycloak returned status %d\n", resp.StatusCode)
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "Authentication failed with status: " + resp.Status,
		})
		return
	}

	log.Printf("✅ Login successful for user: %s\n", req.Username)

	// Register token in rate limiter
	if h.rateLimiter != nil {
		h.rateLimiter.RegisterToken(tokenResp.AccessToken, req.Username)
		log.Printf("✅ Token registered in rate limiter for user: %s (500 calls available)", req.Username)
	}

	// Success - return token info with rate limit info
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"access_token":  tokenResp.AccessToken,
		"expires_in":    tokenResp.ExpiresIn,
		"refresh_token": tokenResp.RefreshToken,
		"token_type":    tokenResp.TokenType,
		"username":      req.Username,
		"rate_limit": gin.H{
			"max_calls":    500,
			"validity_min": 10,
			"expires_at":   time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05"),
			"message":      "You have 500 API calls available with this token. Token expires in 10 minutes.",
		},
	})
}

// RefreshToken handles POST /auth/refresh
// This endpoint refreshes an expired token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	// Prepare the Keycloak token refresh request
	tokenURL := h.keycloakURL + "/realms/" + h.keycloakRealm + "/protocol/openid-connect/token"

	body := url.Values{}
	body.Add("client_id", h.keycloakClient)
	body.Add("client_secret", h.clientSecret)
	body.Add("grant_type", "refresh_token")
	body.Add("refresh_token", req.RefreshToken)

	// Make request to Keycloak
	resp, err := http.Post(
		tokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to connect to authentication service: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to read authentication response: " + err.Error(),
		})
		return
	}

	// Parse token response
	var tokenResp TokenResponse
	if err := json.Unmarshal(responseBody, &tokenResp); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to parse authentication response: " + err.Error(),
		})
		return
	}

	// Check if Keycloak returned an error
	if tokenResp.Error != "" {
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "Token refresh failed: " + tokenResp.ErrorDesc,
		})
		return
	}

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "Token refresh failed with status: " + resp.Status,
		})
		return
	}

	// Success - return new token info
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"access_token":  tokenResp.AccessToken,
		"expires_in":    tokenResp.ExpiresIn,
		"refresh_token": tokenResp.RefreshToken,
		"token_type":    tokenResp.TokenType,
	})
}

// ValidateToken handles GET /auth/validate
// This endpoint validates if a token is still valid
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "No token provided",
		})
		return
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// For now, just return success if token is present
	// In production, you would validate the token signature/expiry
	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "Token is valid",
	})
}

// GetTokenStatus handles GET /auth/token-status
// Returns the current rate limit status and token validity information
func (h *AuthHandler) GetTokenStatus(c *gin.Context) {
	if h.rateLimiter == nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Rate limiter not initialized",
		})
		return
	}

	// Get token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "missing authorization header",
		})
		return
	}

	// Extract Bearer token
	token, err := auth.ExtractBearerToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "invalid authorization header",
		})
		return
	}

	// Get token stats
	stats, err := h.rateLimiter.GetTokenStats(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "token not found or invalid",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   stats,
	})
}

// GetAllTokensStatus handles GET /auth/admin/tokens-status (admin only)
// Returns stats for all active tokens
func (h *AuthHandler) GetAllTokensStatus(c *gin.Context) {
	if h.rateLimiter == nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Rate limiter not initialized",
		})
		return
	}

	stats := h.rateLimiter.GetAllTokenStats()

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   stats,
	})
}
