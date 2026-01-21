package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	elastic "github.com/elastic/go-elasticsearch/v8"
	etcdclient "go.etcd.io/etcd/client/v3"
)

// Global connections
var (
	mysqlDB      *gorm.DB
	mysqlDB2     *gorm.DB
	postgresDB   *gorm.DB
	mongoClient  *mongo.Client
	valkeyClient *redis.Client
	elasticClient *elastic.Client
	etcdClient   *etcdclient.Client
)

// Initialize all connections
func initConnections() error {
	var err error

	// MySQL
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("MYSQL_USER", "root"),
		getEnv("MYSQL_PASSWORD", "root"),
		getEnv("MYSQL_HOST", "localhost"),
		getEnv("MYSQL_PORT", "3306"),
		getEnv("MYSQL_DATABASE", "app_db"),
	)
	mysqlDB, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		log.Printf("❌ MySQL connection failed: %v", err)
	} else {
		log.Printf("✅ MySQL connected")
	}

	// MariaDB
	mariadbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("MARIADB_USER", "root"),
		getEnv("MARIADB_PASSWORD", "root"),
		getEnv("MARIADB_HOST", "localhost"),
		getEnv("MARIADB_PORT", "3306"),
		getEnv("MARIADB_DATABASE", "app_db"),
	)
	mysqlDB2, err = gorm.Open(mysql.Open(mariadbDSN), &gorm.Config{})
	if err != nil {
		log.Printf("❌ MariaDB connection failed: %v", err)
	} else {
		log.Printf("✅ MariaDB connected")
	}

	// PostgreSQL
	postgresDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_USER", "postgres"),
		getEnv("POSTGRES_PASSWORD", "postgres"),
		getEnv("POSTGRES_DATABASE", "app_db"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_SSLMODE", "disable"),
	)
	postgresDB, err = gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Printf("❌ PostgreSQL connection failed: %v", err)
	} else {
		log.Printf("✅ PostgreSQL connected")
	}

	// MongoDB
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		getEnv("MONGODB_USER", "root"),
		getEnv("MONGODB_PASSWORD", "root"),
		getEnv("MONGODB_HOST", "localhost"),
		getEnv("MONGODB_PORT", "27017"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	cancel()
	if err != nil {
		log.Printf("❌ MongoDB connection failed: %v", err)
	} else {
		log.Printf("✅ MongoDB connected")
	}

	// Valkey
	valkeyClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", getEnv("VALKEY_HOST", "localhost"), getEnv("VALKEY_PORT", "6379")),
		Password: getEnv("VALKEY_PASSWORD", ""),
	})
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	_, err = valkeyClient.Ping(ctx).Result()
	cancel()
	if err != nil {
		log.Printf("❌ Valkey connection failed: %v", err)
	} else {
		log.Printf("✅ Valkey connected")
	}

	// Elasticsearch
	elasticClient, err = elastic.NewClient(
		elastic.Config{
			Addresses: []string{fmt.Sprintf("%s://%s:%s",
				getEnv("ELASTICSEARCH_SCHEME", "http"),
				getEnv("ELASTICSEARCH_HOST", "localhost"),
				getEnv("ELASTICSEARCH_PORT", "9200"),
			)},
		},
	)
	if err != nil {
		log.Printf("❌ Elasticsearch connection failed: %v", err)
	} else {
		log.Printf("✅ Elasticsearch connected")
	}

	// etcd
	etcdClient, err = etcdclient.New(etcdclient.Config{
		Endpoints:   []string{fmt.Sprintf("%s:%s", getEnv("ETCD_HOST", "localhost"), getEnv("ETCD_PORT", "2379"))},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("❌ etcd connection failed: %v", err)
	} else {
		log.Printf("✅ etcd connected")
	}

	return nil
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Response structure
type Response struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    interface{}       `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "AxiomNizam API is running",
	})
}

// MySQL endpoint
func queryMySQL(c *gin.Context) {
	if mysqlDB == nil {
		c.JSON(http.StatusServiceUnavailable, Response{
			Status: "error",
			Error:  "MySQL not connected",
		})
		return
	}

	var result map[string]interface{}
	mysqlDB.Raw("SELECT NOW() as current_time").Scan(&result)
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "MySQL query successful",
		Data:    result,
	})
}

// PostgreSQL endpoint
func queryPostgres(c *gin.Context) {
	if postgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, Response{
			Status: "error",
			Error:  "PostgreSQL not connected",
		})
		return
	}

	var result map[string]interface{}
	postgresDB.Raw("SELECT NOW() as current_time").Scan(&result)
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "PostgreSQL query successful",
		Data:    result,
	})
}

// MongoDB endpoint
func queryMongo(c *gin.Context) {
	if mongoClient == nil {
		c.JSON(http.StatusServiceUnavailable, Response{
			Status: "error",
			Error:  "MongoDB not connected",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mongoClient.Ping(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  fmt.Sprintf("MongoDB ping failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "MongoDB ping successful",
		Data:    map[string]string{"status": "connected"},
	})
}

// Valkey endpoint
func queryValkey(c *gin.Context) {
	if valkeyClient == nil {
		c.JSON(http.StatusServiceUnavailable, Response{
			Status: "error",
			Error:  "Valkey not connected",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set a test value
	valkeyClient.Set(ctx, "test_key", "test_value", 1*time.Minute)

	// Get the value
	val, err := valkeyClient.Get(ctx, "test_key").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  fmt.Sprintf("Valkey operation failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "Valkey operation successful",
		Data:    map[string]string{"test_key": val},
	})
}

// Elasticsearch endpoint
func queryElasticsearch(c *gin.Context) {
	if elasticClient == nil {
		c.JSON(http.StatusServiceUnavailable, Response{
			Status: "error",
			Error:  "Elasticsearch not connected",
		})
		return
	}

	resp, err := elasticClient.Info()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  fmt.Sprintf("Elasticsearch query failed: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "Elasticsearch query successful",
		Data:    result,
	})
}

// etcd endpoint
func queryEtcd(c *gin.Context) {
	if etcdClient == nil {
		c.JSON(http.StatusServiceUnavailable, Response{
			Status: "error",
			Error:  "etcd not connected",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set a test value
	etcdClient.Put(ctx, "test_key", "test_value")

	// Get the value
	resp, err := etcdClient.Get(ctx, "test_key")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  fmt.Sprintf("etcd operation failed: %v", err),
		})
		return
	}

	var data interface{}
	if len(resp.Kvs) > 0 {
		data = map[string]string{"test_key": string(resp.Kvs[0].Value)}
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "etcd operation successful",
		Data:    data,
	})
}

// Status endpoint - check all connections
func statusCheck(c *gin.Context) {
	status := map[string]string{
		"mysql":        "disconnected",
		"mariadb":      "disconnected",
		"postgres":     "disconnected",
		"mongodb":      "disconnected",
		"valkey":       "disconnected",
		"elasticsearch": "disconnected",
		"etcd":         "disconnected",
	}

	if mysqlDB != nil {
		status["mysql"] = "connected"
	}
	if mysqlDB2 != nil {
		status["mariadb"] = "connected"
	}
	if postgresDB != nil {
		status["postgres"] = "connected"
	}
	if mongoClient != nil {
		status["mongodb"] = "connected"
	}
	if valkeyClient != nil {
		status["valkey"] = "connected"
	}
	if elasticClient != nil {
		status["elasticsearch"] = "connected"
	}
	if etcdClient != nil {
		status["etcd"] = "connected"
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "System status",
		Data:    status,
	})
}

func main() {	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}
	fmt.Println("� Starting AxiomNizam API Server...\n")

	// Initialize all connections
	initConnections()

	// Create Gin router
	router := gin.Default()

	// Health check
	router.GET("/health", healthCheck)

	// Status check
	router.GET("/status", statusCheck)

	// Database query endpoints
	router.GET("/query/mysql", queryMySQL)
	router.GET("/query/postgres", queryPostgres)
	router.GET("/query/mongodb", queryMongo)

	// Cache endpoint
	router.GET("/query/valkey", queryValkey)

	// Elasticsearch endpoint
	router.GET("/query/elasticsearch", queryElasticsearch)

	// etcd endpoint
	router.GET("/query/etcd", queryEtcd)

	apiPort := getEnv("API_PORT", "8000")
	apiHost := getEnv("API_HOST", "0.0.0.0")

	fmt.Printf("📡 API Server running on http://%s:%s\n", apiHost, apiPort)
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health              - Health check")
	fmt.Println("  GET  /status              - Check all connections")
	fmt.Println("  GET  /query/mysql         - Query MySQL")
	fmt.Println("  GET  /query/postgres      - Query PostgreSQL")
	fmt.Println("  GET  /query/mongodb       - Query MongoDB")
	fmt.Println("  GET  /query/valkey        - Query Valkey/Redis")
	fmt.Println("  GET  /query/elasticsearch - Query Elasticsearch")
	fmt.Println("  GET  /query/etcd          - Query etcd")
	fmt.Println()

	router.Run(fmt.Sprintf("%s:%s", apiHost, apiPort))
}
