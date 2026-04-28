package catalog

// Scanner provides auto-discovery capabilities for the catalog.
// It connects to registered datasources and discovers tables, views,
// columns, and their metadata automatically.

import (
	"context"
	"fmt"
	"time"
)

// ScanRequest represents a request to scan a datasource for assets.
type ScanRequest struct {
	DataSourceRef string   `json:"dataSourceRef"`
	Database      string   `json:"database,omitempty"`
	Schemas       []string `json:"schemas,omitempty"`
	IncludeViews  bool     `json:"includeViews"`
	ExcludePatterns []string `json:"excludePatterns,omitempty"`
}

// ScanResult holds the results of a datasource scan.
type ScanResult struct {
	DataSourceRef  string          `json:"dataSourceRef"`
	AssetsFound    int             `json:"assetsFound"`
	AssetsCreated  int             `json:"assetsCreated"`
	AssetsUpdated  int             `json:"assetsUpdated"`
	Errors         []ScanError     `json:"errors,omitempty"`
	Duration       time.Duration   `json:"duration"`
	ScannedAt      time.Time       `json:"scannedAt"`
	DiscoveredAssets []DiscoveredAsset `json:"discoveredAssets,omitempty"`
}

// ScanError represents an error during scanning.
type ScanError struct {
	Asset   string `json:"asset"`
	Message string `json:"message"`
}

// DiscoveredAsset represents a table/view found during scanning.
type DiscoveredAsset struct {
	Database  string          `json:"database"`
	Schema    string          `json:"schema"`
	Name      string          `json:"name"`
	Type      AssetType       `json:"type"`
	Columns   []CatalogColumn `json:"columns"`
	RowCount  int64           `json:"rowCount,omitempty"`
	SizeBytes int64           `json:"sizeBytes,omitempty"`
}

// Scanner discovers data assets from connected datasources.
type Scanner struct {
	connector DataSourceConnector
	lister    DataSourceLister
}

// DataSourceLister lists tables/views in a datasource.
type DataSourceLister interface {
	// ListTables returns all tables in the given database/schema.
	ListTables(ctx context.Context, dsRef, database, schema string) ([]TableInfo, error)

	// ListSchemas returns all schemas in the given database.
	ListSchemas(ctx context.Context, dsRef, database string) ([]string, error)

	// ListDatabases returns all databases in the datasource.
	ListDatabases(ctx context.Context, dsRef string) ([]string, error)
}

// TableInfo holds basic table metadata from listing.
type TableInfo struct {
	Name      string    `json:"name"`
	Schema    string    `json:"schema"`
	Database  string    `json:"database"`
	Type      AssetType `json:"type"` // table or view
	RowCount  int64     `json:"rowCount,omitempty"`
	SizeBytes int64     `json:"sizeBytes,omitempty"`
}

// NewScanner creates a new catalog scanner.
func NewScanner(connector DataSourceConnector, lister DataSourceLister) *Scanner {
	return &Scanner{
		connector: connector,
		lister:    lister,
	}
}

// Scan performs a full discovery scan of a datasource.
func (s *Scanner) Scan(ctx context.Context, req ScanRequest) (*ScanResult, error) {
	if s.lister == nil {
		return nil, fmt.Errorf("catalog scanner: no datasource lister configured")
	}

	start := time.Now()
	result := &ScanResult{
		DataSourceRef: req.DataSourceRef,
		ScannedAt:     start,
	}

	// Determine schemas to scan.
	schemas := req.Schemas
	if len(schemas) == 0 && req.Database != "" {
		discovered, err := s.lister.ListSchemas(ctx, req.DataSourceRef, req.Database)
		if err != nil {
			return nil, fmt.Errorf("catalog scanner: failed to list schemas: %w", err)
		}
		schemas = discovered
	}

	// If no database specified, discover databases first.
	databases := []string{req.Database}
	if req.Database == "" {
		discovered, err := s.lister.ListDatabases(ctx, req.DataSourceRef)
		if err != nil {
			return nil, fmt.Errorf("catalog scanner: failed to list databases: %w", err)
		}
		databases = discovered
	}

	// Scan each database/schema combination.
	for _, db := range databases {
		if db == "" {
			continue
		}
		dbSchemas := schemas
		if len(dbSchemas) == 0 {
			discovered, err := s.lister.ListSchemas(ctx, req.DataSourceRef, db)
			if err != nil {
				result.Errors = append(result.Errors, ScanError{
					Asset:   db,
					Message: fmt.Sprintf("failed to list schemas: %v", err),
				})
				continue
			}
			dbSchemas = discovered
		}

		for _, schema := range dbSchemas {
			tables, err := s.lister.ListTables(ctx, req.DataSourceRef, db, schema)
			if err != nil {
				result.Errors = append(result.Errors, ScanError{
					Asset:   fmt.Sprintf("%s.%s", db, schema),
					Message: fmt.Sprintf("failed to list tables: %v", err),
				})
				continue
			}

			for _, table := range tables {
				// Skip views if not requested.
				if table.Type == AssetTypeView && !req.IncludeViews {
					continue
				}

				// Skip excluded patterns.
				if s.isExcluded(table.Name, req.ExcludePatterns) {
					continue
				}

				// Introspect columns.
				var columns []CatalogColumn
				if s.connector != nil {
					introspection, err := s.connector.IntrospectTable(ctx, req.DataSourceRef, db, schema, table.Name)
					if err != nil {
						result.Errors = append(result.Errors, ScanError{
							Asset:   fmt.Sprintf("%s.%s.%s", db, schema, table.Name),
							Message: fmt.Sprintf("introspection failed: %v", err),
						})
					} else {
						columns = introspection.Columns
						table.RowCount = introspection.RowCount
						table.SizeBytes = introspection.SizeBytes
					}
				}

				result.DiscoveredAssets = append(result.DiscoveredAssets, DiscoveredAsset{
					Database:  db,
					Schema:    schema,
					Name:      table.Name,
					Type:      table.Type,
					Columns:   columns,
					RowCount:  table.RowCount,
					SizeBytes: table.SizeBytes,
				})
				result.AssetsFound++
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// isExcluded checks if a table name matches any exclusion pattern.
func (s *Scanner) isExcluded(name string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple prefix/suffix matching. Could be extended to glob.
		if matchSimplePattern(name, pattern) {
			return true
		}
	}
	return false
}

// matchSimplePattern does basic wildcard matching (prefix*, *suffix, *contains*).
func matchSimplePattern(name, pattern string) bool {
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}

	// Prefix match: "pg_*"
	if pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(name) >= len(prefix) && name[:len(prefix)] == prefix
	}

	// Suffix match: "*_backup"
	if pattern[0] == '*' {
		suffix := pattern[1:]
		return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
	}

	// Exact match.
	return name == pattern
}
