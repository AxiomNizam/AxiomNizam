package controlplane
import (
	"fmt"
	"regexp"
	"sync"
)

// SchemaValidator validates resource schemas (Kubernetes-style CRD validation)
type SchemaValidator struct {
	mu      sync.RWMutex
	schemas map[string]*ResourceSchema
}

// ResourceSchema defines validation rules for a resource type
type ResourceSchema struct {
	Kind     string
	Version  string
	Fields   map[string]*FieldSchema
	Required []string
}

// FieldSchema defines validation for a specific field
type FieldSchema struct {
	Type        string // string, integer, boolean, array, object
	Required    bool
	Pattern     string // regex pattern
	MinLength   int
	MaxLength   int
	MinValue    int
	MaxValue    int
	Enum        []string
	Default     interface{}
	Description string
	Items       *FieldSchema            // for arrays
	Properties  map[string]*FieldSchema // for objects
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		schemas: make(map[string]*ResourceSchema),
	}
}

// RegisterSchema registers a resource schema
func (sv *SchemaValidator) RegisterSchema(schema *ResourceSchema) error {
	if schema.Kind == "" {
		return fmt.Errorf("schema kind cannot be empty")
	}

	sv.mu.Lock()
	defer sv.mu.Unlock()

	key := fmt.Sprintf("%s/%s", schema.Kind, schema.Version)
	sv.schemas[key] = schema
	return nil
}

// ValidateResource validates a resource against its schema
func (sv *SchemaValidator) ValidateResource(kind, version string, resource map[string]interface{}) error {
	sv.mu.RLock()
	key := fmt.Sprintf("%s/%s", kind, version)
	schema := sv.schemas[key]
	sv.mu.RUnlock()

	if schema == nil {
		return fmt.Errorf("schema not found for %s/%s", kind, version)
	}

	// Check required fields
	for _, field := range schema.Required {
		if _, ok := resource[field]; !ok {
			return fmt.Errorf("required field missing: %s", field)
		}
	}

	// Validate each field
	for field, value := range resource {
		if fieldSchema, ok := schema.Fields[field]; ok {
			if err := sv.validateField(fieldSchema, value); err != nil {
				return fmt.Errorf("field %s validation failed: %w", field, err)
			}
		}
	}

	return nil
}

// validateField validates a single field
func (sv *SchemaValidator) validateField(schema *FieldSchema, value interface{}) error {
	if value == nil {
		if schema.Required {
			return fmt.Errorf("required field is nil")
		}
		return nil
	}

	switch schema.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

		if len(str) < schema.MinLength {
			return fmt.Errorf("string too short (min: %d)", schema.MinLength)
		}
		if schema.MaxLength > 0 && len(str) > schema.MaxLength {
			return fmt.Errorf("string too long (max: %d)", schema.MaxLength)
		}

		if schema.Pattern != "" {
			matched, _ := regexp.MatchString(schema.Pattern, str)
			if !matched {
				return fmt.Errorf("string does not match pattern: %s", schema.Pattern)
			}
		}

		if len(schema.Enum) > 0 {
			found := false
			for _, e := range schema.Enum {
				if e == str {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("value not in enum: %v", schema.Enum)
			}
		}

	case "integer":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected integer, got %T", value)
		}

		intVal := int(num)
		if schema.MinValue > 0 && intVal < schema.MinValue {
			return fmt.Errorf("integer too small (min: %d)", schema.MinValue)
		}
		if schema.MaxValue > 0 && intVal > schema.MaxValue {
			return fmt.Errorf("integer too large (max: %d)", schema.MaxValue)
		}

	case "boolean":
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}

	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
		if schema.Items != nil {
			for i, item := range arr {
				if err := sv.validateField(schema.Items, item); err != nil {
					return fmt.Errorf("array item %d: %w", i, err)
				}
			}
		}

	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
		if schema.Properties != nil {
			for field, fieldSchema := range schema.Properties {
				if fieldVal, ok := obj[field]; ok {
					if err := sv.validateField(fieldSchema, fieldVal); err != nil {
						return fmt.Errorf("property %s: %w", field, err)
					}
				} else if fieldSchema.Required {
					return fmt.Errorf("required property missing: %s", field)
				}
			}
		}
	}

	return nil
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid    bool
	Errors   map[string][]string
	Warnings []string
}

// ValidateResourceWithDetails validates and returns detailed errors
func (sv *SchemaValidator) ValidateResourceWithDetails(kind, version string, resource map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make(map[string][]string),
		Warnings: []string{},
	}

	sv.mu.RLock()
	key := fmt.Sprintf("%s/%s", kind, version)
	schema := sv.schemas[key]
	sv.mu.RUnlock()

	if schema == nil {
		result.Valid = false
		result.Errors["_schema"] = []string{fmt.Sprintf("schema not found for %s/%s", kind, version)}
		return result
	}

	// Check required fields
	for _, field := range schema.Required {
		if _, ok := resource[field]; !ok {
			result.Valid = false
			result.Errors[field] = append(result.Errors[field], "required field missing")
		}
	}

	// Validate each field
	for field, value := range resource {
		if fieldSchema, ok := schema.Fields[field]; ok {
			if err := sv.validateField(fieldSchema, value); err != nil {
				result.Valid = false
				result.Errors[field] = append(result.Errors[field], err.Error())
			}
		}
	}

	return result
}
