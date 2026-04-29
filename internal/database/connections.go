package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"example.com/axiomnizam/internal/config"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	elastic "github.com/elastic/go-elasticsearch/v8"
	etcdclient "go.etcd.io/etcd/client/v3"
)

// Connections holds all database connections
type Connections struct {
	MySQL         *gorm.DB
	MariaDB       *gorm.DB
	Percona       *gorm.DB
	PostgreSQL    *gorm.DB
	MongoDB       *mongo.Client
	Valkey        *redis.Client
	Elasticsearch *elastic.Client
	Etcd          *etcdclient.Client
	Oracle        *gorm.DB
	Firebase      interface{} // Placeholder for Firebase connection
}

// InitConnections initializes all database connections
func InitConnections(cfg *config.Config) *Connections {
	conns := &Connections{}

	// MySQL
	if db, err := gorm.Open(mysql.Open(cfg.GetMySQLDSN()), &gorm.Config{}); err == nil {
		conns.MySQL = db
		log.Println("✅ MySQL connected")
	} else {
		log.Printf("❌ MySQL connection failed: %v", err)
	}

	// MariaDB
	if db, err := gorm.Open(mysql.Open(cfg.GetMariaDBDSN()), &gorm.Config{}); err == nil {
		conns.MariaDB = db
		log.Println("✅ MariaDB connected")
	} else {
		log.Printf("❌ MariaDB connection failed: %v", err)
	}

	// Percona
	if db, err := gorm.Open(mysql.Open(cfg.GetPerconaDSN()), &gorm.Config{}); err == nil {
		conns.Percona = db
		log.Println("✅ Percona connected")
	} else {
		log.Printf("❌ Percona connection failed: %v", err)
	}

	// PostgreSQL
	if db, err := gorm.Open(postgres.Open(cfg.GetPostgresDSN()), &gorm.Config{}); err == nil {
		conns.PostgreSQL = db
		log.Println("✅ PostgreSQL connected")
	} else {
		log.Printf("❌ PostgreSQL connection failed: %v", err)
	}

	// MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.GetMongoDBURI())); err == nil {
		conns.MongoDB = client
		log.Println("✅ MongoDB connected")
	} else {
		log.Printf("❌ MongoDB connection failed: %v", err)
	}

	// Valkey
	conns.Valkey = redis.NewClient(&redis.Options{
		Addr:     cfg.GetValkeyAddr(),
		Password: cfg.Valkey.Password,
	})
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := conns.Valkey.Ping(ctx).Result(); err == nil {
		log.Println("✅ Valkey connected")
	} else {
		log.Printf("❌ Valkey connection failed: %v", err)
	}

	// Elasticsearch
	if client, err := elastic.NewClient(elastic.Config{
		Addresses: []string{cfg.GetElasticsearchURL()},
	}); err == nil {
		conns.Elasticsearch = client
		log.Println("✅ Elasticsearch connected")
	} else {
		log.Printf("❌ Elasticsearch connection failed: %v", err)
	}

	// etcd — skip when using embedded Raft storage backend
	storageBackend := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_BACKEND")))
	if storageBackend == "raft" {
		log.Println("ℹ️  etcd skipped (STORAGE_BACKEND=raft — using embedded Raft storage)")
	} else if client, err := etcdclient.New(etcdclient.Config{
		Endpoints:   []string{fmt.Sprintf("%s:%s", cfg.Etcd.Host, cfg.Etcd.Port)},
		DialTimeout: 5 * time.Second,
	}); err == nil {
		conns.Etcd = client
		log.Println("✅ etcd connected")
	} else {
		log.Printf("❌ etcd connection failed: %v", err)
	}

	// Oracle
	if db, err := gorm.Open(postgres.Open(cfg.GetOracleDSN()), &gorm.Config{}); err == nil {
		conns.Oracle = db
		log.Println("✅ Oracle connected")
	} else {
		log.Printf("⚠️  Oracle connection failed: %v", err)
	}

	// Firebase (placeholder)
	log.Println("⚠️  Firebase connection - placeholder (requires Firebase credentials)")

	return conns
}

// Close closes all database connections
func (c *Connections) Close() {
	if c.MongoDB != nil {
		c.MongoDB.Disconnect(context.Background())
	}
	if c.Valkey != nil {
		c.Valkey.Close()
	}
	if c.Etcd != nil {
		c.Etcd.Close()
	}
}

// IsConnected returns connection status for all databases
func (c *Connections) IsConnected() map[string]bool {
	status := map[string]bool{
		"mysql":         c.MySQL != nil,
		"mariadb":       c.MariaDB != nil,
		"percona":       c.Percona != nil,
		"postgres":      c.PostgreSQL != nil,
		"mongodb":       c.MongoDB != nil,
		"valkey":        c.Valkey != nil,
		"elasticsearch": c.Elasticsearch != nil,
		"etcd":          c.Etcd != nil,
		"oracle":        c.Oracle != nil,
	}
	return status
}
