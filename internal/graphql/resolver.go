package graphql

import (
	"context"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"gorm.io/gorm"
)

// QueryResolver resolves GraphQL queries against database
type QueryResolver struct {
	db *gorm.DB
}

// NewQueryResolver creates a new query resolver
func NewQueryResolver(db *gorm.DB) *QueryResolver {
	return &QueryResolver{db: db}
}

// ResolveQuery executes a GraphQL query
func (qr *QueryResolver) ResolveQuery(ctx context.Context, query string, variables map[string]interface{}, operationName string) (interface{}, error) {
	schema, err := BuildDatabaseSchema(qr.db)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema: %w", err)
	}

	result := graphql.Do(graphql.Params{
		Schema:         *schema,
		RequestString:  query,
		VariableValues: variables,
		OperationName:  operationName,
		Context:        ctx,
	})

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	return result.Data, nil
}

// BuildSchema builds a GraphQL schema from the configured database connection.
func (qr *QueryResolver) BuildSchema() (*graphql.Schema, error) {
	return BuildDatabaseSchema(qr.db)
}

// BuildDatabaseSchema builds a GraphQL schema from database
func BuildDatabaseSchema(db *gorm.DB) (*graphql.Schema, error) {
	if db == nil {
		return nil, fmt.Errorf("no SQL database connection available for GraphQL schema generation")
	}

	builder := NewSchemaBuilder()

	tables, err := listTables(db)
	if err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("no database tables available for GraphQL schema generation")
	}

	// Add schemas for each table
	for _, table := range tables {
		columns := getTableColumns(db, table)
		builder.AddTableSchema(table, columns)
	}

	return builder.BuildSchema()
}

func listTables(db *gorm.DB) ([]string, error) {
	var (
		query string
		args  []interface{}
	)

	switch strings.ToLower(db.Dialector.Name()) {
	case "postgres":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'"
	case "mysql", "mariadb":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'"
	default:
		query = "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE'"
	}

	rows, err := db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// getTableColumns retrieves columns for a table
func getTableColumns(db *gorm.DB, tableName string) map[string]string {
	columns := make(map[string]string)

	type ColumnInfo struct {
		ColumnName string
		DataType   string
	}

	var colInfos []ColumnInfo
	switch strings.ToLower(db.Dialector.Name()) {
	case "postgres":
		db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ?", tableName).Scan(&colInfos)
	case "mysql", "mariadb":
		db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&colInfos)
	default:
		db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = ?", tableName).Scan(&colInfos)
	}

	for _, col := range colInfos {
		columns[col.ColumnName] = strings.ToUpper(col.DataType)
	}

	return columns
}

// QueryMetrics tracks query performance
type QueryMetrics struct {
	Query        string
	Duration     int64 // milliseconds
	RowsScanned  int64
	RowsReturned int64
	CacheHit     bool
}

// MetricsCollector collects GraphQL query metrics
type MetricsCollector struct {
	metrics []QueryMetrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make([]QueryMetrics, 0),
	}
}

// Record records a query metric
func (mc *MetricsCollector) Record(metric QueryMetrics) {
	mc.metrics = append(mc.metrics, metric)
}

// GetMetrics returns all metrics
func (mc *MetricsCollector) GetMetrics() []QueryMetrics {
	return mc.metrics
}
