package graphql

import (
	"context"
	"fmt"

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
func (qr *QueryResolver) ResolveQuery(ctx context.Context, query string) (interface{}, error) {
	schema, err := BuildDatabaseSchema(qr.db)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema: %w", err)
	}

	result := graphql.Do(graphql.Params{
		Schema:        *schema,
		RequestString: query,
		Context:       ctx,
	})

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	return result.Data, nil
}

// BuildDatabaseSchema builds a GraphQL schema from database
func BuildDatabaseSchema(db *gorm.DB) (*graphql.Schema, error) {
	builder := NewSchemaBuilder()

	// Get all tables
	tables := []string{}
	rows, err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	// Add schemas for each table
	for _, table := range tables {
		columns := getTableColumns(db, table)
		builder.AddTableSchema(table, columns)
	}

	return builder.BuildSchema()
}

// getTableColumns retrieves columns for a table
func getTableColumns(db *gorm.DB, tableName string) map[string]string {
	columns := make(map[string]string)

	type ColumnInfo struct {
		ColumnName string
		DataType   string
	}

	var colInfos []ColumnInfo
	db.Raw(fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s'", tableName)).Scan(&colInfos)

	for _, col := range colInfos {
		columns[col.ColumnName] = col.DataType
	}

	return columns
}

// QueryMetrics tracks query performance
type QueryMetrics struct {
	Query       string
	Duration    int64 // milliseconds
	RowsScanned int64
	RowsReturned int64
	CacheHit    bool
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
