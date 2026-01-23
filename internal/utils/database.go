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

// QueryBuilder helps construct SQL queries
type QueryBuilder struct {
	selectClauses []string
	fromClause    string
	whereClauses  []string
	params        []interface{}
	orderBy       []string
	groupBy       []string
	havingClauses []string
	limitValue    int
	offsetValue   int
	joinClauses   []string
	distinct      bool
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		selectClauses: []string{},
		whereClauses:  []string{},
		params:        []interface{}{},
		orderBy:       []string{},
		groupBy:       []string{},
		havingClauses: []string{},
		joinClauses:   []string{},
	}
}

// Select adds columns to SELECT clause
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectClauses = append(qb.selectClauses, columns...)
	return qb
}

// From sets the FROM table
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.fromClause = table
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, condition)
	qb.params = append(qb.params, args...)
	return qb
}

// AndWhere adds an AND WHERE condition
func (qb *QueryBuilder) AndWhere(condition string, args ...interface{}) *QueryBuilder {
	if len(qb.whereClauses) > 0 {
		qb.whereClauses = append(qb.whereClauses, "AND "+condition)
	} else {
		qb.whereClauses = append(qb.whereClauses, condition)
	}
	qb.params = append(qb.params, args...)
	return qb
}

// OrWhere adds an OR WHERE condition
func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	if len(qb.whereClauses) > 0 {
		qb.whereClauses = append(qb.whereClauses, "OR "+condition)
	} else {
		qb.whereClauses = append(qb.whereClauses, condition)
	}
	qb.params = append(qb.params, args...)
	return qb
}

// Join adds an INNER JOIN clause
func (qb *QueryBuilder) Join(table, condition string) *QueryBuilder {
	qb.joinClauses = append(qb.joinClauses, fmt.Sprintf("INNER JOIN %s ON %s", table, condition))
	return qb
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	qb.joinClauses = append(qb.joinClauses, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

// RightJoin adds a RIGHT JOIN clause
func (qb *QueryBuilder) RightJoin(table, condition string) *QueryBuilder {
	qb.joinClauses = append(qb.joinClauses, fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition))
	return qb
}

// GroupBy adds a GROUP BY clause
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// Having adds a HAVING clause
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.havingClauses = append(qb.havingClauses, condition)
	qb.params = append(qb.params, args...)
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(orderSpec ...string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, orderSpec...)
	return qb
}

// Distinct adds DISTINCT to SELECT
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.distinct = true
	return qb
}

// Limit sets the LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitValue = limit
	return qb
}

// Offset sets the OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetValue = offset
	return qb
}

// Build constructs the SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	if qb.distinct {
		query.WriteString("DISTINCT ")
	}
	if len(qb.selectClauses) > 0 {
		query.WriteString(strings.Join(qb.selectClauses, ", "))
	} else {
		query.WriteString("*")
	}

	// FROM clause
	if qb.fromClause == "" {
		return "", qb.params // Error: no table specified
	}
	query.WriteString(" FROM " + qb.fromClause)

	// JOIN clauses
	for _, join := range qb.joinClauses {
		query.WriteString(" " + join)
	}

	// WHERE clause
	if len(qb.whereClauses) > 0 {
		query.WriteString(" WHERE " + strings.Join(qb.whereClauses, " "))
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY " + strings.Join(qb.groupBy, ", "))
	}

	// HAVING clause
	if len(qb.havingClauses) > 0 {
		query.WriteString(" HAVING " + strings.Join(qb.havingClauses, " AND "))
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY " + strings.Join(qb.orderBy, ", "))
	}

	// LIMIT clause
	if qb.limitValue > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limitValue))
	}

	// OFFSET clause
	if qb.offsetValue > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offsetValue))
	}

	return query.String(), qb.params
}

// String returns the built query as string
func (qb *QueryBuilder) String() string {
	query, _ := qb.Build()
	return query
}

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
