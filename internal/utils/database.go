package utils

import (
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// ConnectionPool manages database connections
type ConnectionPool struct {
	connections map[string]*gorm.DB
	mu          sync.RWMutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*gorm.DB),
	}
}

// Add adds a connection to the pool
func (cp *ConnectionPool) Add(name string, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("cannot add nil connection")
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if _, exists := cp.connections[name]; exists {
		return fmt.Errorf("connection '%s' already exists", name)
	}

	cp.connections[name] = db
	return nil
}

// Get retrieves a connection from the pool
func (cp *ConnectionPool) Get(name string) (*gorm.DB, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	db, exists := cp.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}

	return db, nil
}

// Remove removes a connection from the pool
func (cp *ConnectionPool) Remove(name string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if _, exists := cp.connections[name]; !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}

	delete(cp.connections, name)
	return nil
}

// GetAll returns all connections
func (cp *ConnectionPool) GetAll() map[string]*gorm.DB {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	connections := make(map[string]*gorm.DB)
	for name, db := range cp.connections {
		connections[name] = db
	}
	return connections
}

// Size returns the number of connections in the pool
func (cp *ConnectionPool) Size() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.connections)
}

// Close closes all connections
func (cp *ConnectionPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	var errors []string
	for name, db := range cp.connections {
		if db != nil {
			sqlDB, err := db.DB()
			if err == nil {
				if err := sqlDB.Close(); err != nil {
					errors = append(errors, fmt.Sprintf("failed to close connection '%s': %v", name, err))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %s", strings.Join(errors, "; "))
	}

	return nil
}

// QueryBuilder implementation is in query_builder.go

// SQLBuilder for INSERT, UPDATE, DELETE operations
type SQLBuilder struct {
	operation string
	table     string
	columns   []string
	values    []interface{}
	set       map[string]interface{}
	where     []string
	params    []interface{}
}

// NewInsertBuilder creates a builder for INSERT
func NewInsertBuilder(table string) *SQLBuilder {
	return &SQLBuilder{
		operation: "INSERT",
		table:     table,
		columns:   []string{},
		values:    []interface{}{},
		set:       make(map[string]interface{}),
		where:     []string{},
		params:    []interface{}{},
	}
}

// NewUpdateBuilder creates a builder for UPDATE
func NewUpdateBuilder(table string) *SQLBuilder {
	return &SQLBuilder{
		operation: "UPDATE",
		table:     table,
		set:       make(map[string]interface{}),
		where:     []string{},
		params:    []interface{}{},
	}
}

// NewDeleteBuilder creates a builder for DELETE
func NewDeleteBuilder(table string) *SQLBuilder {
	return &SQLBuilder{
		operation: "DELETE",
		table:     table,
		where:     []string{},
		params:    []interface{}{},
	}
}

// Columns sets columns for INSERT
func (sb *SQLBuilder) Columns(cols ...string) *SQLBuilder {
	sb.columns = append(sb.columns, cols...)
	return sb
}

// Values sets values for INSERT
func (sb *SQLBuilder) Values(vals ...interface{}) *SQLBuilder {
	sb.values = append(sb.values, vals...)
	return sb
}

// Set sets column=value for UPDATE
func (sb *SQLBuilder) Set(column string, value interface{}) *SQLBuilder {
	sb.set[column] = value
	return sb
}

// Where adds WHERE condition
func (sb *SQLBuilder) Where(condition string, args ...interface{}) *SQLBuilder {
	sb.where = append(sb.where, condition)
	sb.params = append(sb.params, args...)
	return sb
}

// BuildInsert builds INSERT query
func (sb *SQLBuilder) BuildInsert() (string, []interface{}) {
	if len(sb.columns) != len(sb.values) {
		return "", []interface{}{} // Error: columns and values mismatch
	}

	var query strings.Builder
	query.WriteString("INSERT INTO " + sb.table + " (")
	query.WriteString(strings.Join(sb.columns, ", "))
	query.WriteString(") VALUES (")

	placeholders := make([]string, len(sb.values))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query.WriteString(strings.Join(placeholders, ", "))
	query.WriteString(")")

	return query.String(), sb.values
}

// BuildUpdate builds UPDATE query
func (sb *SQLBuilder) BuildUpdate() (string, []interface{}) {
	var query strings.Builder
	query.WriteString("UPDATE " + sb.table + " SET ")

	setClauses := []string{}
	setParams := []interface{}{}

	for column, value := range sb.set {
		setClauses = append(setClauses, column+" = ?")
		setParams = append(setParams, value)
	}

	query.WriteString(strings.Join(setClauses, ", "))

	if len(sb.where) > 0 {
		query.WriteString(" WHERE " + strings.Join(sb.where, " AND "))
	}

	allParams := append(setParams, sb.params...)
	return query.String(), allParams
}

// BuildDelete builds DELETE query
func (sb *SQLBuilder) BuildDelete() (string, []interface{}) {
	var query strings.Builder
	query.WriteString("DELETE FROM " + sb.table)

	if len(sb.where) > 0 {
		query.WriteString(" WHERE " + strings.Join(sb.where, " AND "))
	}

	return query.String(), sb.params
}

// Build constructs the appropriate query based on operation
func (sb *SQLBuilder) Build() (string, []interface{}) {
	switch sb.operation {
	case "INSERT":
		return sb.BuildInsert()
	case "UPDATE":
		return sb.BuildUpdate()
	case "DELETE":
		return sb.BuildDelete()
	default:
		return "", []interface{}{}
	}
}
