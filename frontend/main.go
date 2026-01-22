package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
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

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using defaults")
	}

	// Get backend URL from environment or use default
	backendURL = os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8000"
	}
	// Ensure backendURL doesn't have trailing slash
	if len(backendURL) > 1 && backendURL[len(backendURL)-1] == '/' {
		backendURL = backendURL[:len(backendURL)-1]
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
	router.GET("/admin", adminHandler)
	router.GET("/system-manager", systemManagerHandler)
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
	fmt.Printf("📡 Backend: %s\n\n", backendURL)

	router.Run(fmt.Sprintf(":%s", port))
}

// dashboardHandler serves the public dashboard
func dashboardHandler(c *gin.Context) {
	isAuth := c.GetBool("isAuthenticated")
	userName := c.GetString("userName")

	health, _ := fetchHealth()

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":       "AxiomNizam - Dashboard",
		"pageName":    "public-dashboard",
		"page":        "public-dashboard",
		"isAuth":      isAuth,
		"userName":    userName,
		"backendURL":  "http://localhost:8000",
		"frontendURL": fmt.Sprintf("http://localhost:%s", os.Getenv("FRONTEND_PORT")),
		"health":      health,
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

// systemManagerHandler serves the system manager dashboard
func systemManagerHandler(c *gin.Context) {
	// Check authentication
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		authToken, _ = c.Cookie("authToken")
	}

	isAuth := authToken != ""
	userName := "Manager"

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":      "AxiomNizam - System Manager",
		"pageName":   "system-manager",
		"page":       "system-manager",
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

// apiHealthHandler fetches and returns health status as JSON
func apiHealthHandler(c *gin.Context) {
	health, err := fetchHealth()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Failed to fetch health: %v", err),
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
			"message": fmt.Sprintf("Failed to fetch status: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, status)
}

// fetchHealth makes a request to the backend health endpoint
func fetchHealth() (*HealthResponse, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("%s/health", backendURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var health HealthResponse
	err = json.Unmarshal(body, &health)
	if err != nil {
		return nil, err
	}

	return &health, nil
}

// fetchStatus makes a request to the backend status endpoint
func fetchStatus() (*StatusResponse, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("%s/status", backendURL))
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
