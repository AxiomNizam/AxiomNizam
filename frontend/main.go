package main

import (
	"encoding/json"
	"fmt"
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

	router := gin.Default()

	// Serve static files and HTML templates
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "templates/")

	// Routes
	router.GET("/", dashboardHandler)
	router.GET("/api/health", apiHealthHandler)
	router.GET("/api/status", apiStatusHandler)

	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🌐 Frontend Dashboard running on http://localhost:%s\n", port)
	fmt.Printf("📡 Monitoring backend at %s\n", backendURL)

	router.Run(fmt.Sprintf(":%s", port))
}

// dashboardHandler serves the main dashboard
func dashboardHandler(c *gin.Context) {
	health, _ := fetchHealth()
	status, _ := fetchStatus()

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"health":     health,
		"status":     status,
		"backendURL": backendURL,
	})
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
