package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration
type Config struct {
	API        APIConfig
	MySQL      DatabaseConfig
	MariaDB    DatabaseConfig
	Percona    DatabaseConfig
	PostgreSQL DatabaseConfig
	MongoDB    DatabaseConfig
	Oracle     DatabaseConfig
	Firebase   FirebaseConfig
	Valkey     ValKeyConfig
	Elasticsearch ElasticsearchConfig
	Etcd       EtcdConfig
}

type APIConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	SID      string
}

type ValKeyConfig struct {
	Host     string
	Port     string
	Password string
}

type ElasticsearchConfig struct {
	Host   string
	Port   string
	Scheme string
}

type EtcdConfig struct {
	Host string
	Port string
}

type FirebaseConfig struct {
	ProjectID     string
	PrivateKeyID  string
	PrivateKey    string
	ClientEmail   string
	ClientID      string
	AuthURI       string
	TokenURI      string
	DatabaseURL   string
}

// LoadConfig loads configuration from environment
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	return &Config{
		API: APIConfig{
			Port: getEnv("API_PORT", "8000"),
			Host: getEnv("API_HOST", "0.0.0.0"),
		},
		MySQL: DatabaseConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "root"),
			Database: getEnv("MYSQL_DATABASE", "app_db"),
		},
		MariaDB: DatabaseConfig{
			Host:     getEnv("MARIADB_HOST", "localhost"),
			Port:     getEnv("MARIADB_PORT", "3306"),
			User:     getEnv("MARIADB_USER", "root"),
			Password: getEnv("MARIADB_PASSWORD", "root"),
			Database: getEnv("MARIADB_DATABASE", "app_db"),
		},
		Percona: DatabaseConfig{
			Host:     getEnv("PERCONA_HOST", "localhost"),
			Port:     getEnv("PERCONA_PORT", "3306"),
			User:     getEnv("PERCONA_USER", "root"),
			Password: getEnv("PERCONA_PASSWORD", "root"),
			Database: getEnv("PERCONA_DATABASE", "app_db"),
		},
		PostgreSQL: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			Database: getEnv("POSTGRES_DATABASE", "app_db"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		MongoDB: DatabaseConfig{
			Host:     getEnv("MONGODB_HOST", "localhost"),
			Port:     getEnv("MONGODB_PORT", "27017"),
			User:     getEnv("MONGODB_USER", "root"),
			Password: getEnv("MONGODB_PASSWORD", "root"),
			Database: getEnv("MONGODB_DATABASE", "app_db"),
		},
		Oracle: DatabaseConfig{
			Host:     getEnv("ORACLE_HOST", "localhost"),
			Port:     getEnv("ORACLE_PORT", "1521"),
			User:     getEnv("ORACLE_USER", "system"),
			Password: getEnv("ORACLE_PASSWORD", "oracle123"),
			SID:      getEnv("ORACLE_SID", "XE"),
			Database: getEnv("ORACLE_DATABASE", "app_db"),
		},
		Firebase: FirebaseConfig{
			ProjectID:    getEnv("FIREBASE_PROJECT_ID", ""),
			PrivateKeyID: getEnv("FIREBASE_PRIVATE_KEY_ID", ""),
			PrivateKey:   getEnv("FIREBASE_PRIVATE_KEY", ""),
			ClientEmail:  getEnv("FIREBASE_CLIENT_EMAIL", ""),
			ClientID:     getEnv("FIREBASE_CLIENT_ID", ""),
			AuthURI:      getEnv("FIREBASE_AUTH_URI", ""),
			TokenURI:     getEnv("FIREBASE_TOKEN_URI", ""),
			DatabaseURL:  getEnv("FIREBASE_DATABASE_URL", ""),
		},
		Valkey: ValKeyConfig{
			Host:     getEnv("VALKEY_HOST", "localhost"),
			Port:     getEnv("VALKEY_PORT", "6379"),
			Password: getEnv("VALKEY_PASSWORD", ""),
		},
		Elasticsearch: ElasticsearchConfig{
			Host:   getEnv("ELASTICSEARCH_HOST", "localhost"),
			Port:   getEnv("ELASTICSEARCH_PORT", "9200"),
			Scheme: getEnv("ELASTICSEARCH_SCHEME", "http"),
		},
		Etcd: EtcdConfig{
			Host: getEnv("ETCD_HOST", "localhost"),
			Port: getEnv("ETCD_PORT", "2379"),
		},
	}
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetMySQLDSN returns MySQL DSN
func (c *Config) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQL.User, c.MySQL.Password, c.MySQL.Host, c.MySQL.Port, c.MySQL.Database)
}

// GetMariaDBDSN returns MariaDB DSN
func (c *Config) GetMariaDBDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MariaDB.User, c.MariaDB.Password, c.MariaDB.Host, c.MariaDB.Port, c.MariaDB.Database)
}

// GetPerconaDSN returns Percona DSN
func (c *Config) GetPerconaDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Percona.User, c.Percona.Password, c.Percona.Host, c.Percona.Port, c.Percona.Database)
}

// GetPostgresDSN returns PostgreSQL DSN
func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.PostgreSQL.Host, c.PostgreSQL.User, c.PostgreSQL.Password,
		c.PostgreSQL.Database, c.PostgreSQL.Port, c.PostgreSQL.SSLMode)
}

// GetMongoDBURI returns MongoDB URI
func (c *Config) GetMongoDBURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s",
		c.MongoDB.User, c.MongoDB.Password, c.MongoDB.Host, c.MongoDB.Port)
}

// GetOracleConnectionString returns Oracle connection string
func (c *Config) GetOracleConnectionString() string {
	return fmt.Sprintf("%s/%s@%s:%s/%s",
		c.Oracle.User, c.Oracle.Password, c.Oracle.Host, c.Oracle.Port, c.Oracle.SID)
}

// GetElasticsearchURL returns Elasticsearch URL
func (c *Config) GetElasticsearchURL() string {
	return fmt.Sprintf("http://%s:%s", c.Elasticsearch.Host, c.Elasticsearch.Port)
}

// GetEtcdEndpoint returns etcd endpoint
func (c *Config) GetEtcdEndpoint() string {
	return fmt.Sprintf("%s:%s", c.Etcd.Host, c.Etcd.Port)
}

// GetValkeyAddr returns Valkey address
func (c *Config) GetValkeyAddr() string {
	return fmt.Sprintf("%s:%s", c.Valkey.Host, c.Valkey.Port)
}
