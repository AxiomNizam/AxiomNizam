package models

// User model for database operations
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Username string `json:"username" gorm:"uniqueIndex"`
	Email    string `json:"email" gorm:"uniqueIndex"`
	Password string `json:"password"`
	Age      int    `json:"age"`
}

// Response structure for API responses
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthResponse for health check
type HealthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime,omitempty"`
}

// DatabaseStatus for status check
type DatabaseStatus struct {
	MySQL         string `json:"mysql"`
	MariaDB       string `json:"mariadb"`
	PostgreSQL    string `json:"postgres"`
	MongoDB       string `json:"mongodb"`
	Valkey        string `json:"valkey"`
	Elasticsearch string `json:"elasticsearch"`
	Etcd          string `json:"etcd"`
	Oracle        string `json:"oracle"`
	Firebase      string `json:"firebase"`
}
