package handlers

import (
	"encoding/base64"
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
	gojwt "github.com/golang-jwt/jwt/v5"
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
	// Build Keycloak URL from host and port environment variables
	keycloakHost := getEnv("KEYCLOAK_HOST", "keycloak")
	keycloakPort := getEnv("KEYCLOAK_PORT", "8080")
	keycloakURL := fmt.Sprintf("http://%s:%s", keycloakHost, keycloakPort)

	return &AuthHandler{
		keycloakURL:    keycloakURL,
		keycloakRealm:  getEnv("KEYCLOAK_REALM", "axiomnizam"),
		keycloakClient: getEnv("KEYCLOAK_CLIENT_ID", "axiomnizam-backend"),
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

// demoAccounts maps username → {password, role} for local dev fallback
// admin   / admin    → /admin        (full admin access)
// sysadmin/ sysadmin → /system-manager (user management, system ops)
// manager / manager  → /manager      (view+edit APIs/dashboards, no delete/create)
// user    / user     → /             (view only)
var demoAccounts = map[string]struct {
	password string
	role     string
}{
	"admin":    {password: "admin", role: "admin"},
	"sysadmin": {password: "sysadmin", role: "system-manager"},
	"manager":  {password: "manager", role: "manager"},
	"user":     {password: "user", role: "user"},
}

// generateDemoToken creates an HMAC-HS256 JWT for a demo account
func generateDemoToken(username, role string) (string, error) {
	now := time.Now()
	claims := gojwt.MapClaims{
		"preferred_username": username,
		"email":              username + "@demo.local",
		"realm_access": map[string]interface{}{
			"roles": []string{role, "uma_authorization"},
		},
		"demo": true,
		"iat":  now.Unix(),
		"exp":  now.Add(8 * time.Hour).Unix(),
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(auth.DemoJWTSecret))
}

// extractRoleFromToken decodes the JWT payload and determines the user role
func extractRoleFromToken(tokenString string) string {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "user"
	}
	payload := parts[1]
	// Add padding
	padded := payload + strings.Repeat("=", (4-len(payload)%4)%4)
	padded = strings.ReplaceAll(padded, "-", "+")
	padded = strings.ReplaceAll(padded, "_", "/")
	decoded, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		return "user"
	}
	var payload2 struct {
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
	}
	if err := json.Unmarshal(decoded, &payload2); err != nil {
		return "user"
	}
	for _, r := range payload2.RealmAccess.Roles {
		rl := strings.ToLower(r)
		if strings.Contains(rl, "admin") && !strings.Contains(rl, "account") {
			return "admin"
		}
		if rl == "system-manager" || rl == "system_manager" || rl == "system-admin" {
			return "system-manager"
		}
		if rl == "manager" || rl == "api-manager" || rl == "api_manager" {
			return "manager"
		}
	}
	return "user"
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

	// Check demo accounts FIRST — always takes priority over Keycloak for known demo credentials.
	// This ensures local dev accounts always work regardless of Keycloak state.
	// Set ENABLE_DEMO_ACCOUNTS=false to disable (e.g. in production when you want to remove these accounts).
	demoEnabled := getEnv("ENABLE_DEMO_ACCOUNTS", "true") == "true"
	if demo, ok := demoAccounts[req.Username]; ok && demo.password == req.Password {
		if !demoEnabled {
			log.Printf("⚠️  Demo account '%s' matched but ENABLE_DEMO_ACCOUNTS=false — falling through to Keycloak\n", req.Username)
		} else {
			demoToken, err := generateDemoToken(req.Username, demo.role)
			if err == nil {
				log.Printf("✅ Demo login for user: %s (role: %s)\n", req.Username, demo.role)
				c.JSON(http.StatusOK, gin.H{
					"status":        "ok",
					"access_token":  demoToken,
					"expires_in":    28800,
					"refresh_token": "",
					"token_type":    "Bearer",
					"username":      req.Username,
					"role":          demo.role,
					"demo_mode":     true,
				})
				return
			}
			log.Printf("⚠️  Demo token generation failed for %s: %v — falling through to Keycloak\n", req.Username, err)
		}
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

	// Determine role from the Keycloak JWT payload (server-side) for reliable redirect
	keycloakRole := extractRoleFromToken(tokenResp.AccessToken)

	// Success - return token info with rate limit info
	c.JSON(http.StatusOK, gin.H{
		"role":          keycloakRole,
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
