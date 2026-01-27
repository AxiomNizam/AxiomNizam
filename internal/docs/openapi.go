package docs

import (
	"fmt"
	"reflect"
	"strings"
)

// OpenAPIInfo contains OpenAPI metadata
type OpenAPIInfo struct {
	Title       string
	Version     string
	Description string
}

// OpenAPIEndpoint represents an API endpoint
type OpenAPIEndpoint struct {
	Path        string
	Method      string
	Summary     string
	Description string
	Tags        []string
	Parameters  []OpenAPIParameter
	RequestBody OpenAPIRequestBody
	Response    OpenAPIResponse
	Security    []string
}

// OpenAPIParameter represents an API parameter
type OpenAPIParameter struct {
	Name        string
	In          string // path, query, header
	Required    bool
	Type        string
	Description string
	Example     interface{}
}

// OpenAPIRequestBody represents request body
type OpenAPIRequestBody struct {
	Required    bool
	ContentType string
	Schema      map[string]interface{}
}

// OpenAPIResponse represents API response
type OpenAPIResponse struct {
	Status      int
	Description string
	Schema      map[string]interface{}
	Example     interface{}
}

// OpenAPIGenerator generates OpenAPI specifications
type OpenAPIGenerator struct {
	info      OpenAPIInfo
	endpoints []OpenAPIEndpoint
	schemas   map[string]interface{}
}

// NewOpenAPIGenerator creates a new OpenAPI generator
func NewOpenAPIGenerator(info OpenAPIInfo) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		info:      info,
		endpoints: make([]OpenAPIEndpoint, 0),
		schemas:   make(map[string]interface{}),
	}
}

// AddEndpoint adds an endpoint to the OpenAPI spec
func (og *OpenAPIGenerator) AddEndpoint(endpoint OpenAPIEndpoint) {
	og.endpoints = append(og.endpoints, endpoint)
}

// AddSchema adds a schema to the OpenAPI spec
func (og *OpenAPIGenerator) AddSchema(name string, schema map[string]interface{}) {
	og.schemas[name] = schema
}

// GenerateFromStruct generates schema from Go struct
func (og *OpenAPIGenerator) GenerateFromStruct(name string, data interface{}) map[string]interface{} {
	schema := og.structToSchema(data)
	og.AddSchema(name, schema)
	return schema
}

// structToSchema converts a Go struct to OpenAPI schema
func (og *OpenAPIGenerator) structToSchema(data interface{}) map[string]interface{} {
	t := reflect.TypeOf(data)
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	properties := schema["properties"].(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		if jsonTag == "-" || jsonTag == "" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		fieldType := og.goTypeToOpenAPIType(field.Type)

		properties[fieldName] = map[string]interface{}{
			"type": fieldType,
		}
	}

	return schema
}

// goTypeToOpenAPIType converts Go type to OpenAPI type
func (og *OpenAPIGenerator) goTypeToOpenAPIType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}

// BuildOpenAPI builds the complete OpenAPI spec
func (og *OpenAPIGenerator) BuildOpenAPI() map[string]interface{} {
	paths := make(map[string]interface{})

	for _, endpoint := range og.endpoints {
		pathKey := endpoint.Path

		if _, exists := paths[pathKey]; !exists {
			paths[pathKey] = make(map[string]interface{})
		}

		methodKey := strings.ToLower(endpoint.Method)
		pathObj := paths[pathKey].(map[string]interface{})

		pathObj[methodKey] = og.buildEndpointSpec(endpoint)
	}

	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       og.info.Title,
			"version":     og.info.Version,
			"description": og.info.Description,
		},
		"paths":   paths,
		"schemas": og.schemas,
	}
}

// buildEndpointSpec builds individual endpoint spec
func (og *OpenAPIGenerator) buildEndpointSpec(endpoint OpenAPIEndpoint) map[string]interface{} {
	spec := map[string]interface{}{
		"summary":     endpoint.Summary,
		"description": endpoint.Description,
		"tags":        endpoint.Tags,
	}

	// Add parameters
	if len(endpoint.Parameters) > 0 {
		params := make([]map[string]interface{}, len(endpoint.Parameters))
		for i, param := range endpoint.Parameters {
			params[i] = map[string]interface{}{
				"name":        param.Name,
				"in":          param.In,
				"required":    param.Required,
				"description": param.Description,
				"schema": map[string]interface{}{
					"type": param.Type,
				},
			}
		}
		spec["parameters"] = params
	}

	// Add request body
	if endpoint.RequestBody.ContentType != "" {
		spec["requestBody"] = map[string]interface{}{
			"required": endpoint.RequestBody.Required,
			"content": map[string]interface{}{
				endpoint.RequestBody.ContentType: map[string]interface{}{
					"schema": endpoint.RequestBody.Schema,
				},
			},
		}
	}

	// Add responses
	spec["responses"] = map[string]interface{}{
		fmt.Sprintf("%d", endpoint.Response.Status): map[string]interface{}{
			"description": endpoint.Response.Description,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": endpoint.Response.Schema,
				},
			},
		},
	}

	// Add security
	if len(endpoint.Security) > 0 {
		spec["security"] = endpoint.Security
	}

	return spec
}

// GetEndpointMarkdown returns markdown documentation for endpoints
func (og *OpenAPIGenerator) GetEndpointMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", og.info.Title))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", og.info.Version))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", og.info.Description))

	sb.WriteString("## Endpoints\n\n")

	for _, endpoint := range og.endpoints {
		sb.WriteString(fmt.Sprintf("### %s %s\n", endpoint.Method, endpoint.Path))
		sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", endpoint.Summary))

		if endpoint.Description != "" {
			sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", endpoint.Description))
		}

		if len(endpoint.Parameters) > 0 {
			sb.WriteString("**Parameters:**\n\n")
			for _, param := range endpoint.Parameters {
				required := "false"
				if param.Required {
					required = "true"
				}
				sb.WriteString(fmt.Sprintf("- `%s` (%s): %s - Required: %s\n", param.Name, param.Type, param.Description, required))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// ExportYAML exports OpenAPI spec as YAML string
func (og *OpenAPIGenerator) ExportYAML() string {
	spec := og.BuildOpenAPI()
	return fmt.Sprintf("%+v", spec)
}

// GetSwaggerUIHTML returns Swagger UI HTML
func GetSwaggerUIHTML(specURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <redoc spec-url='%s'></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"></script>
</body>
</html>
`, specURL)
}
