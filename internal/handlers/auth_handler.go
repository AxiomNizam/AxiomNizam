package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	iamBaseURL    string
	rateLimiter   *auth.RateLimiter
	platformUsers *PlatformUserHandler
	httpClient    *http.Client
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler() *AuthHandler {
	iamBaseURL := strings.TrimSpace(getEnv("IAM_ISSUER_URL", ""))
	if iamBaseURL == "" {
		host := getEnv("API_HOST", "localhost")
		port := getEnv("API_PORT", "8000")
		iamBaseURL = fmt.Sprintf("http://%s:%s", host, port)
	}
	iamBaseURL = strings.TrimRight(iamBaseURL, "/")

	return &AuthHandler{
		iamBaseURL:  iamBaseURL,
		rateLimiter: nil, // Will be set via SetRateLimiter
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetRateLimiter sets the rate limiter for the auth handler
func (h *AuthHandler) SetRateLimiter(limiter *auth.RateLimiter) {
	h.rateLimiter = limiter
}

// SetPlatformUserHandler wires the platform user store into the auth handler
// so that users created via the sysadmin UI can log in.
func (h *AuthHandler) SetPlatformUserHandler(puh *PlatformUserHandler) {
	h.platformUsers = puh
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
	return token.SignedString([]byte(auth.DemoJWTSecret()))
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
		Roles       []string `json:"roles"`
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
		ResourceAccess map[string]struct {
			Roles []string `json:"roles"`
		} `json:"resource_access"`
	}
	if err := json.Unmarshal(decoded, &payload2); err != nil {
		return "user"
	}

	allRoles := make([]string, 0, len(payload2.Roles)+len(payload2.RealmAccess.Roles)+8)
	allRoles = append(allRoles, payload2.Roles...)
	allRoles = append(allRoles, payload2.RealmAccess.Roles...)
	for _, access := range payload2.ResourceAccess {
		allRoles = append(allRoles, access.Roles...)
	}

	for _, r := range allRoles {
		rl := strings.ToLower(strings.TrimSpace(r))
		if rl == "system-manager" || rl == "system_manager" || rl == "system-admin" || rl == "sysadmin" {
			return "system-manager"
		}
	}
	for _, r := range allRoles {
		rl := strings.ToLower(strings.TrimSpace(r))
		if strings.Contains(rl, "admin") && !strings.Contains(rl, "account") {
			return "admin"
		}
	}
	for _, r := range allRoles {
		rl := strings.ToLower(strings.TrimSpace(r))
		if rl == "manager" || rl == "api-manager" || rl == "api_manager" {
			return "manager"
		}
	}
	return "user"
}

func resolvePrimaryRole(roles []string) string {
	resolved := "user"
	for _, role := range roles {
		r := strings.ToLower(strings.TrimSpace(role))
		switch {
		case r == "system-manager" || r == "system_manager" || r == "system-admin" || r == "sysadmin":
			return "system-manager"
		case strings.Contains(r, "admin") && !strings.Contains(r, "account"):
			resolved = "admin"
		case (r == "manager" || r == "api-manager" || r == "api_manager") && resolved == "user":
			resolved = "manager"
		}
	}
	return resolved
}

// TokenResponse is the response payload used by IAM token endpoints.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
	User         struct {
		ID          string   `json:"id,omitempty"`
		Email       string   `json:"email,omitempty"`
		DisplayName string   `json:"display_name,omitempty"`
		Roles       []string `json:"roles,omitempty"`
	} `json:"user,omitempty"`
}

// Login handles POST /auth/login
// This endpoint proxies authentication to built-in IAM endpoints.
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

	iamOnlyAuth := strings.EqualFold(getEnv("IAM_ONLY_AUTH", "true"), "true")

	if !iamOnlyAuth {
		// Optional local demo login path for development.
		demoEnabled := getEnv("ENABLE_DEMO_ACCOUNTS", "false") == "true"
		if demo, ok := demoAccounts[req.Username]; ok && demo.password == req.Password {
			if demoEnabled {
				demoToken, err := generateDemoToken(req.Username, demo.role)
				if err == nil {
					if h.rateLimiter != nil {
						h.rateLimiter.RegisterToken(demoToken, req.Username)
					}
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
			}
		}

		// Optional platform user fallback.
		if h.platformUsers != nil {
			if platformUser, ok := h.platformUsers.ValidateCredentials(req.Username, req.Password); ok {
				platformToken, err := generateDemoToken(platformUser.Username, platformUser.Role)
				if err == nil {
					if h.rateLimiter != nil {
						h.rateLimiter.RegisterToken(platformToken, platformUser.Username)
					}
					c.JSON(http.StatusOK, gin.H{
						"status":        "ok",
						"access_token":  platformToken,
						"expires_in":    28800,
						"refresh_token": "",
						"token_type":    "Bearer",
						"username":      platformUser.Username,
						"role":          platformUser.Role,
						"demo_mode":     true,
					})
					return
				}
			}
		}
	}

	loginID := strings.TrimSpace(req.Username)
	if loginID == "" {
		c.JSON(http.StatusBadRequest, models.Response{Status: "error", Error: "username is required"})
		return
	}

	if !strings.Contains(loginID, "@") {
		defaultDomain := strings.TrimSpace(getEnv("IAM_DEFAULT_EMAIL_DOMAIN", ""))
		if defaultDomain != "" {
			loginID = loginID + "@" + strings.TrimPrefix(defaultDomain, "@")
		}
	}

	tokenURL := h.iamBaseURL + "/iam/auth/login"
	log.Printf("📝 IAM login attempt for identifier: %s", loginID)

	body, _ := json.Marshal(map[string]string{
		"email":    loginID,
		"password": req.Password,
	})

	resp, err := h.httpClient.Post(tokenURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to connect to IAM authentication service: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to read IAM authentication response: " + err.Error(),
		})
		return
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(responseBody, &tokenResp); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to parse IAM authentication response: " + err.Error(),
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := strings.TrimSpace(tokenResp.Error)
		if errMsg == "" {
			errMsg = strings.TrimSpace(tokenResp.ErrorDesc)
		}
		if errMsg == "" {
			errMsg = strings.TrimSpace(string(responseBody))
		}
		if errMsg == "" {
			errMsg = "authentication failed"
		}

		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  errMsg,
		})
		return
	}

	resolvedRole := resolvePrimaryRole(tokenResp.User.Roles)
	if resolvedRole == "user" {
		resolvedRole = extractRoleFromToken(tokenResp.AccessToken)
	}

	resolvedUsername := strings.TrimSpace(tokenResp.User.DisplayName)
	if resolvedUsername == "" {
		resolvedUsername = strings.TrimSpace(tokenResp.User.Email)
	}
	if resolvedUsername == "" {
		resolvedUsername = req.Username
	}

	if h.rateLimiter != nil {
		h.rateLimiter.RegisterToken(tokenResp.AccessToken, resolvedUsername)
		log.Printf("✅ Token registered in rate limiter for user: %s (500 calls available)", resolvedUsername)
	}

	c.JSON(http.StatusOK, gin.H{
		"role":          resolvedRole,
		"status":        "ok",
		"access_token":  tokenResp.AccessToken,
		"expires_in":    tokenResp.ExpiresIn,
		"refresh_token": tokenResp.RefreshToken,
		"token_type":    tokenResp.TokenType,
		"username":      resolvedUsername,
		"user":          tokenResp.User,
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

	tokenURL := h.iamBaseURL + "/iam/auth/refresh"

	body, _ := json.Marshal(map[string]string{
		"refresh_token": req.RefreshToken,
	})

	resp, err := h.httpClient.Post(
		tokenURL,
		"application/json",
		strings.NewReader(string(body)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to connect to IAM authentication service: " + err.Error(),
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
			Error:  "Failed to parse IAM authentication response: " + err.Error(),
		})
		return
	}

	// Check if IAM returned an error
	if tokenResp.Error != "" {
		errText := tokenResp.ErrorDesc
		if strings.TrimSpace(errText) == "" {
			errText = tokenResp.Error
		}
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "Token refresh failed: " + errText,
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
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.Response{
			Status: "error",
			Error:  "No token provided",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	req, err := http.NewRequest(http.MethodGet, h.iamBaseURL+"/iam/auth/whoami", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Status: "error", Error: "Failed to build validation request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Response{Status: "error", Error: "IAM validation endpoint unreachable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, models.Response{Status: "error", Error: "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, models.Response{Status: "ok", Message: "Token is valid"})
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
