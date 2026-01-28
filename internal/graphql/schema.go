package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

// SchemaBuilder generates GraphQL schema from database models
type SchemaBuilder struct {
	types   map[string]*graphql.Object
	fields  map[string]graphql.Fields
	queries graphql.Fields
}

// NewSchemaBuilder creates a new schema builder
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		types:   make(map[string]*graphql.Object),
		fields:  make(map[string]graphql.Fields),
		queries: make(graphql.Fields),
	}
}

// AddTableSchema adds a table schema to the GraphQL schema
func (sb *SchemaBuilder) AddTableSchema(tableName string, columns map[string]string) error {
	fields := graphql.Fields{}

	// Add ID field
	fields["id"] = &graphql.Field{
		Type: graphql.NewNonNull(graphql.ID),
	}

	// Map SQL types to GraphQL types
	for colName, colType := range columns {
		graphqlType := sqlTypeToGraphQLType(colType)
		if graphqlType == nil {
			continue
		}
		fields[colName] = &graphql.Field{
			Type: graphqlType,
		}
	}

	// Create object type
	objectType := graphql.NewObject(graphql.ObjectConfig{
		Name:   tableName,
		Fields: fields,
	})

	sb.types[tableName] = objectType
	sb.fields[tableName] = fields

	// Add query fields
	sb.queries[tableName] = &graphql.Field{
		Type: graphql.NewList(objectType),
		Args: graphql.FieldConfigArgument{
			"limit": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"offset": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"where": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return nil, nil
		},
	}

	sb.queries[fmt.Sprintf("%sById", tableName)] = &graphql.Field{
		Type: objectType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return nil, nil
		},
	}

	return nil
}

// BuildSchema builds the final GraphQL schema
func (sb *SchemaBuilder) BuildSchema() (*graphql.Schema, error) {
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: sb.queries,
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})

	if err != nil {
		return nil, fmt.Errorf("error building schema: %w", err)
	}

	return &schema, nil
}

// sqlTypeToGraphQLType converts SQL types to GraphQL types
func sqlTypeToGraphQLType(sqlType string) graphql.Type {
	switch sqlType {
	case "INT", "INTEGER", "SMALLINT", "BIGINT", "TINYINT":
		return graphql.Int
	case "FLOAT", "DOUBLE", "DECIMAL", "NUMERIC":
		return graphql.Float
	case "BOOLEAN", "BOOL":
		return graphql.Boolean
	case "TEXT", "VARCHAR", "CHAR", "STRING":
		return graphql.String
	case "TIMESTAMP", "DATETIME", "DATE", "TIME":
		return graphql.String
	case "JSON", "JSONB":
		return graphql.String
	default:
		return graphql.String
	}
}

// GetType returns a GraphQL type for a table
func (sb *SchemaBuilder) GetType(tableName string) *graphql.Object {
	return sb.types[tableName]
}

// GetSchema returns the GraphQL schema object
func (sb *SchemaBuilder) GetSchema() (*graphql.Schema, error) {
	return sb.BuildSchema()
}
