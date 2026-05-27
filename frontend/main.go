package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// HealthResponse represents the health endpoint response
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StatusResponse represents the status endpoint response
type StatusResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
}

var backendURL string
var backendProxyURL string

func trimTrailingSlash(raw string) string {
	value := strings.TrimSpace(raw)
	if len(value) > 1 && strings.HasSuffix(value, "/") {
		return value[:len(value)-1]
	}
	return value
}

func normalizeFrontendRole(role string) string {
	value := strings.ToLower(strings.TrimSpace(role))
	switch value {
	case "sysadmin", "system-admin", "system_admin":
		return "system-manager"
	case "superadmin", "super-admin":
		return "admin"
	case "api-manager", "api_manager":
		return "manager"
	case "admin", "manager", "system-manager":
		return value
	default:
		return "user"
	}
}

func defaultPathForRole(role string) string {
	switch normalizeFrontendRole(role) {
	case "system-manager":
		return "/system-manager"
	case "admin":
		return "/admin"
	case "manager":
		return "/manager"
	default:
		return "/"
	}
}

func setNoCacheHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}

func requireFrontendRoles(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]bool, len(allowed))
	for _, role := range allowed {
		allowedSet[normalizeFrontendRole(role)] = true
	}

	return func(c *gin.Context) {
		authToken := c.GetHeader("Authorization")
		if authToken == "" {
			authToken, _ = c.Cookie("authToken")
		}
		if authToken == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		role := c.GetHeader("X-User-Role")
		if role == "" {
			role, _ = c.Cookie("userRole")
		}
		normalized := normalizeFrontendRole(role)
		if !allowedSet[normalized] {
			c.Redirect(http.StatusFound, defaultPathForRole(normalized))
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using defaults")
	}

	// Get browser-facing backend URL from environment or use default.
	backendURL = trimTrailingSlash(os.Getenv("BACKEND_URL"))
	if backendURL == "" {
		backendURL = "http://127.0.0.1:8000"
	}

	// Backend proxy URL is used by frontend server-side routes (/api/health, /api/status).
	// This can point to an internal service address even when BACKEND_URL is public.
	backendProxyURL = trimTrailingSlash(os.Getenv("BACKEND_PROXY_URL"))
	if backendProxyURL == "" {
		backendProxyURL = backendURL
	}

	router := gin.Default()

	// Add custom template functions
	router.SetFuncMap(template.FuncMap{
		"safeHTML": func(html string) template.HTML {
			return template.HTML(html)
		},
	})

	// Serve static files and HTML templates
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "templates/")

	// Routes
	router.GET("/", dashboardHandler)
	router.GET("/signup", signupHandler)
	router.GET("/login", loginHandler)
	router.GET("/admin", adminHandler)
	router.GET("/system-manager", systemManagerHandler)
	router.GET("/manager", managerHandler)
	router.GET("/gis", gisHandler)
	router.GET("/analytics", analyticsHandler)
	router.GET("/cdc-etl", cdcEtlHandler)
	router.GET("/netintel", netintelHandler)
	router.GET("/conductor", conductorHandler)
	router.GET("/governance", requireFrontendRoles("admin", "system-manager"), governanceHandler)
	router.GET("/operations-center", requireFrontendRoles("admin", "system-manager", "manager"), operationsCenterHandler)
	router.GET("/lineage-version", requireFrontendRoles("admin", "system-manager"), versionLineageHandler)
	router.GET("/iam-admin", requireFrontendRoles("system-manager"), iamAdminHandler)
	router.GET("/object-storage", requireFrontendRoles("system-manager"), objectStorageHandler)
	router.GET("/two-factor", twoFactorHandler)
	router.GET("/favicon.ico", faviconHandler)
	router.GET("/api/health", apiHealthHandler)
	router.GET("/api/status", apiStatusHandler)

	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "7000"
	}

	fmt.Printf("\n🌐 Frontend Server running on http://localhost:%s\n", port)
	fmt.Printf("📊 Dashboard: http://localhost:%s\n", port)
	fmt.Printf("🔧 Admin: http://localhost:%s/admin\n", port)
	fmt.Printf("🖥️  System Manager: http://localhost:%s/system-manager\n", port)
	fmt.Printf("📋 Manager Portal: http://localhost:%s/manager\n", port)
	fmt.Printf("🌍 GIS Dashboard: http://localhost:%s/gis\n", port)
	fmt.Printf("📊 Analytics Dashboard: http://localhost:%s/analytics\n", port)
	fmt.Printf("🔄 CDC/ETL Dashboard: http://localhost:%s/cdc-etl\n", port)
	fmt.Printf("📡 Network Intelligence: http://localhost:%s/netintel\n", port)
	fmt.Printf("� Conductor: http://localhost:%s/conductor\n", port)
	fmt.Printf("�🏛️ Governance Console: http://localhost:%s/governance\n", port)
	fmt.Printf("🛠️ Operations Center: http://localhost:%s/operations-center\n", port)
	fmt.Printf("🧭 Version & Lineage: http://localhost:%s/lineage-version\n", port)
	fmt.Printf("🔐 IAM Admin Console: http://localhost:%s/iam-admin\n", port)
	fmt.Printf("🔒 Two-Factor Auth: http://localhost:%s/two-factor\n", port)
	fmt.Printf("📝 Signup: http://localhost:%s/signup\n", port)
	fmt.Printf("📡 Backend (browser): %s\n", backendURL)
	fmt.Printf("🔁 Backend (proxy): %s\n\n", backendProxyURL)

	router.Run(fmt.Sprintf(":%s", port))
}

// dashboardHandler serves the public landing page (no auth required)
func dashboardHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}
	isAuth := authToken != ""

	userName := "Guest"
	if fromCookie, _ := c.Cookie("userName"); strings.TrimSpace(fromCookie) != "" {
		userName = fromCookie
	}

	health, _ := fetchHealth()

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":       "AxiomNizam - Enterprise Data Control Plane",
		"pageName":    "public-dashboard",
		"page":        "public-dashboard",
		"isAuth":      isAuth,
		"userName":    userName,
		"backendURL":  backendURL,
		"frontendURL": fmt.Sprintf("http://localhost:%s", os.Getenv("FRONTEND_PORT")),
		"health":      health,
	})
}

// signupHandler serves the signup page
func signupHandler(c *gin.Context) {
	setNoCacheHeaders(c)

	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}
	if strings.TrimSpace(authToken) != "" {
		role := c.GetHeader("X-User-Role")
		if role == "" {
			role, _ = c.Cookie("userRole")
		}
		c.Redirect(http.StatusFound, defaultPathForRole(role))
		return
	}

	health, _ := fetchHealth()
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":       "AxiomNizam - Sign Up",
		"pageName":    "signup",
		"page":        "signup",
		"isAuth":      false,
		"userName":    "Guest",
		"backendURL":  backendURL,
		"frontendURL": fmt.Sprintf("http://localhost:%s", os.Getenv("FRONTEND_PORT")),
		"health":      health,
		"hideChrome":  true,
	})
}

// loginHandler serves login-focused entrypoint for unauthenticated users.
func loginHandler(c *gin.Context) {
	setNoCacheHeaders(c)

	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}
	if strings.TrimSpace(authToken) != "" {
		role := c.GetHeader("X-User-Role")
		if role == "" {
			role, _ = c.Cookie("userRole")
		}
		c.Redirect(http.StatusFound, defaultPathForRole(role))
		return
	}

	health, _ := fetchHealth()
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":       "AxiomNizam - Login",
		"pageName":    "login",
		"page":        "login",
		"isAuth":      false,
		"userName":    "Guest",
		"backendURL":  backendURL,
		"frontendURL": fmt.Sprintf("http://localhost:%s", os.Getenv("FRONTEND_PORT")),
		"health":      health,
		"hideChrome":  true,
	})
}

// adminHandler serves the admin dashboard
func adminHandler(c *gin.Context) {
	// Check authentication
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Admin"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Admin",
		"pageName":   "admin",
		"page":       "admin",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// managerHandler serves the manager portal
func managerHandler(c *gin.Context) {
	setNoCacheHeaders(c)

	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Manager"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Manager Portal",
		"pageName":   "manager",
		"page":       "manager",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// systemManagerHandler serves the system manager dashboard
func systemManagerHandler(c *gin.Context) {
	setNoCacheHeaders(c)

	// Check authentication
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "System Manager"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - System Manager",
		"pageName":   "system-manager",
		"page":       "system-manager",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// analyticsHandler serves the analytics dashboard
func analyticsHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	embedded := c.Query("embed") == "1" || strings.EqualFold(c.Query("embed"), "true")
	if embedded {
		setNoCacheHeaders(c)
	}
	userName := "User"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Analytics Dashboard",
		"pageName":   "analytics-dashboard",
		"page":       "analytics-dashboard",
		"isAuth":     isAuth,
		"embedded":   embedded,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// cdcEtlHandler serves the CDC/ETL dashboard
func cdcEtlHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "User"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - CDC/ETL Dashboard",
		"pageName":   "cdc-etl-dashboard",
		"page":       "cdc-etl-dashboard",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// netintelHandler serves the Network Intelligence dashboard
func netintelHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	embedded := c.Query("embed") == "1" || strings.EqualFold(c.Query("embed"), "true")
	if embedded {
		setNoCacheHeaders(c)
	}
	userName := "User"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Network Intelligence",
		"pageName":   "netintel-dashboard",
		"page":       "netintel-dashboard",
		"isAuth":     isAuth,
		"embedded":   embedded,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// governanceHandler serves the governance dashboard
func governanceHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Governance"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Governance Console",
		"pageName":   "governance-dashboard",
		"page":       "governance-dashboard",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// operationsCenterHandler serves incidents and operations center
func operationsCenterHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Operations"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Operations Center",
		"pageName":   "operations-center",
		"page":       "operations-center",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// versionLineageHandler serves version and lineage explorer
func versionLineageHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Explorer"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Version & Lineage Explorer",
		"pageName":   "version-lineage-dashboard",
		"page":       "version-lineage-dashboard",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// gisHandler serves the GIS dashboard
func gisHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	embedded := c.Query("embed") == "1" || strings.EqualFold(c.Query("embed"), "true")
	if embedded {
		setNoCacheHeaders(c)
	}
	userName := "User"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - GIS Dashboard",
		"pageName":   "gis-dashboard",
		"page":       "gis-dashboard",
		"isAuth":     isAuth,
		"embedded":   embedded,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// iamAdminHandler serves the IAM admin console
func iamAdminHandler(c *gin.Context) {
	setNoCacheHeaders(c)

	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "IAM Admin"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - IAM Admin Console",
		"pageName":   "iam-admin",
		"page":       "iam-admin",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// objectStorageHandler serves the Object Storage console
func objectStorageHandler(c *gin.Context) {
	setNoCacheHeaders(c)

	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Storage Admin"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Object Storage",
		"pageName":   "object-storage",
		"page":       "object-storage",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// twoFactorHandler serves the 2FA management page
func twoFactorHandler(c *gin.Context) {
	setNoCacheHeaders(c)

	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "User"
	if fromCookie, _ := c.Cookie("userName"); strings.TrimSpace(fromCookie) != "" {
		userName = fromCookie
	}

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Two-Factor Authentication",
		"pageName":   "two-factor",
		"page":       "two-factor",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// conductorHandler serves the conductor dashboard
func conductorHandler(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Conductor"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - Conductor",
		"pageName":   "conductor-dashboard",
		"page":       "conductor-dashboard",
		"isAuth":     isAuth,
		"userName":   userName,
		"backendURL": backendURL,
	})
}

// faviconHandler serves a favicon
func faviconHandler(c *gin.Context) {
	c.Header("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>⚙️</text></svg>`)
}

// apiHealthHandler fetches and returns health status as JSON.
// Returns 200 with degraded status when backend is unreachable,
// so the frontend itself is never blamed for backend startup delays.
func apiHealthHandler(c *gin.Context) {
	health, err := fetchHealth()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":         "degraded",
			"message":        fmt.Sprintf("Backend unreachable at %s: %v", backendProxyURL, err),
			"frontend":       "ok",
			"backend":        "unreachable",
		})
		return
	}
	c.JSON(http.StatusOK, health)
}

// apiStatusHandler fetches and returns status as JSON
func apiStatusHandler(c *gin.Context) {
	status, err := fetchStatus()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Failed to fetch status from %s: %v", backendProxyURL, err),
		})
		return
	}
	c.JSON(http.StatusOK, status)
}

// fetchHealth makes a request to the backend health endpoint.
// Tries backendProxyURL first, falls back to backendURL, then to 127.0.0.1.
func fetchHealth() (*HealthResponse, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Build candidate URLs with automatic fallbacks for Docker networking.
	seen := map[string]bool{}
	var candidates []string
	fallbacks := []string{"axiomnizam", "host.docker.internal", "127.0.0.1"}
	for _, u := range []string{backendProxyURL, backendURL} {
		if u != "" && !seen[u] {
			seen[u] = true
			candidates = append(candidates, u)
		}
		// Try Docker service name / host fallbacks when URL uses localhost.
		if strings.Contains(u, "localhost") || strings.Contains(u, "127.0.0.1") {
			for _, host := range fallbacks {
				fallback := u
				fallback = strings.Replace(fallback, "localhost", host, 1)
				fallback = strings.Replace(fallback, "127.0.0.1", host, 1)
				if !seen[fallback] {
					seen[fallback] = true
					candidates = append(candidates, fallback)
				}
			}
		}
	}

	var lastErr error
	for _, baseURL := range candidates {
		resp, err := client.Get(fmt.Sprintf("%s/health", baseURL))
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("status %d from %s", resp.StatusCode, baseURL)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var health HealthResponse
		if err := json.Unmarshal(body, &health); err != nil {
			return nil, err
		}
		return &health, nil
	}
	return nil, lastErr
}

// fetchStatus makes a request to the backend status endpoint
func fetchStatus() (*StatusResponse, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("%s/status", backendProxyURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var status StatusResponse
	err = json.Unmarshal(body, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
