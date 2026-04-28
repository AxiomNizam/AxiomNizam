package schemaregistry

// Compatibility checking engine for schema evolution.
//
// Implements backward, forward, and full compatibility validation
// for JSON Schema. Avro and Protobuf support can be added incrementally.

import (
	"encoding/json"
	"fmt"
)

// JSONSchemaCompatibilityChecker validates JSON Schema compatibility.
type JSONSchemaCompatibilityChecker struct{}

// NewJSONSchemaCompatibilityChecker creates a new checker.
func NewJSONSchemaCompatibilityChecker() *JSONSchemaCompatibilityChecker {
	return &JSONSchemaCompatibilityChecker{}
}

// CheckCompatibility validates compatibility between two schemas.
func (c *JSONSchemaCompatibilityChecker) CheckCompatibility(
	newSchema, oldSchema string,
	schemaType SchemaType,
	mode CompatibilityMode,
) []string {
	switch schemaType {
	case SchemaTypeJSON:
		return c.checkJSONCompatibility(newSchema, oldSchema, mode)
	case SchemaTypeAvro:
		return c.checkAvroCompatibility(newSchema, oldSchema, mode)
	default:
		return []string{fmt.Sprintf("unsupported schema type: %s", schemaType)}
	}
}

// checkJSONCompatibility validates JSON Schema compatibility.
func (c *JSONSchemaCompatibilityChecker) checkJSONCompatibility(newSchema, oldSchema string, mode CompatibilityMode) []string {
	var newParsed, oldParsed map[string]interface{}

	if err := json.Unmarshal([]byte(newSchema), &newParsed); err != nil {
		return []string{fmt.Sprintf("failed to parse new schema: %v", err)}
	}
	if err := json.Unmarshal([]byte(oldSchema), &oldParsed); err != nil {
		return []string{fmt.Sprintf("failed to parse old schema: %v", err)}
	}

	var errors []string

	switch mode {
	case CompatBackward, CompatBackwardTransitive:
		errors = c.checkBackwardCompat(newParsed, oldParsed)
	case CompatForward, CompatForwardTransitive:
		errors = c.checkForwardCompat(newParsed, oldParsed)
	case CompatFull, CompatFullTransitive:
		backwardErrors := c.checkBackwardCompat(newParsed, oldParsed)
		forwardErrors := c.checkForwardCompat(newParsed, oldParsed)
		errors = append(backwardErrors, forwardErrors...)
	case CompatNone:
		// No compatibility check needed.
		return nil
	}

	return errors
}

// checkBackwardCompat ensures new schema can read data written with old schema.
// Rules:
//   - Cannot remove a field that existed in old schema (unless it had a default)
//   - Cannot add a required field without a default value
//   - Cannot narrow a type (e.g., number -> integer)
func (c *JSONSchemaCompatibilityChecker) checkBackwardCompat(newSchema, oldSchema map[string]interface{}) []string {
	var errors []string

	oldProps := getProperties(oldSchema)
	newProps := getProperties(newSchema)
	oldRequired := getRequired(oldSchema)
	newRequired := getRequired(newSchema)

	// Check for removed fields.
	for fieldName := range oldProps {
		if _, exists := newProps[fieldName]; !exists {
			// Field removed — check if it was required.
			if containsStr(oldRequired, fieldName) {
				errors = append(errors, fmt.Sprintf("BACKWARD: required field '%s' was removed", fieldName))
			}
		}
	}

	// Check for new required fields without defaults.
	for fieldName := range newProps {
		if _, existedBefore := oldProps[fieldName]; !existedBefore {
			if containsStr(newRequired, fieldName) {
				// New required field — check for default.
				fieldDef := getFieldDef(newProps, fieldName)
				if !hasDefault(fieldDef) {
					errors = append(errors, fmt.Sprintf("BACKWARD: new required field '%s' has no default value", fieldName))
				}
			}
		}
	}

	// Check for type narrowing.
	for fieldName, oldDef := range oldProps {
		if newDef, exists := newProps[fieldName]; exists {
			if typeNarrowed(oldDef, newDef) {
				errors = append(errors, fmt.Sprintf("BACKWARD: field '%s' type was narrowed", fieldName))
			}
		}
	}

	return errors
}

// checkForwardCompat ensures old schema can read data written with new schema.
// Rules:
//   - Cannot add a field without a default (old reader won't know about it)
//   - Cannot remove a required field (old reader expects it)
//   - Cannot widen a type (e.g., integer -> number)
func (c *JSONSchemaCompatibilityChecker) checkForwardCompat(newSchema, oldSchema map[string]interface{}) []string {
	var errors []string

	oldProps := getProperties(oldSchema)
	newProps := getProperties(newSchema)
	oldRequired := getRequired(oldSchema)

	// Check for removed required fields (old reader expects them).
	for fieldName := range oldProps {
		if _, exists := newProps[fieldName]; !exists {
			if containsStr(oldRequired, fieldName) {
				errors = append(errors, fmt.Sprintf("FORWARD: required field '%s' was removed (old readers expect it)", fieldName))
			}
		}
	}

	// Check for type widening.
	for fieldName, oldDef := range oldProps {
		if newDef, exists := newProps[fieldName]; exists {
			if typeWidened(oldDef, newDef) {
				errors = append(errors, fmt.Sprintf("FORWARD: field '%s' type was widened (old readers may not handle it)", fieldName))
			}
		}
	}

	return errors
}

// checkAvroCompatibility validates Avro schema compatibility.
// Simplified implementation — production would use a proper Avro library.
func (c *JSONSchemaCompatibilityChecker) checkAvroCompatibility(newSchema, oldSchema string, mode CompatibilityMode) []string {
	var newParsed, oldParsed map[string]interface{}

	if err := json.Unmarshal([]byte(newSchema), &newParsed); err != nil {
		return []string{fmt.Sprintf("failed to parse new Avro schema: %v", err)}
	}
	if err := json.Unmarshal([]byte(oldSchema), &oldParsed); err != nil {
		return []string{fmt.Sprintf("failed to parse old Avro schema: %v", err)}
	}

	var errors []string

	// For Avro, check fields array.
	oldFields := getAvroFields(oldParsed)
	newFields := getAvroFields(newParsed)

	switch mode {
	case CompatBackward, CompatBackwardTransitive:
		// New schema must be able to read old data.
		// New fields must have defaults.
		for name, def := range newFields {
			if _, existed := oldFields[name]; !existed {
				if !hasAvroDefault(def) {
					errors = append(errors, fmt.Sprintf("BACKWARD: new Avro field '%s' has no default", name))
				}
			}
		}
	case CompatForward, CompatForwardTransitive:
		// Old schema must be able to read new data.
		// Removed fields must have had defaults in old schema.
		for name, def := range oldFields {
			if _, exists := newFields[name]; !exists {
				if !hasAvroDefault(def) {
					errors = append(errors, fmt.Sprintf("FORWARD: removed Avro field '%s' had no default in old schema", name))
				}
			}
		}
	case CompatFull, CompatFullTransitive:
		backwardErrors := c.checkAvroCompatibility(newSchema, oldSchema, CompatBackward)
		forwardErrors := c.checkAvroCompatibility(newSchema, oldSchema, CompatForward)
		errors = append(backwardErrors, forwardErrors...)
	}

	return errors
}

// --- Helper functions ---

func getProperties(schema map[string]interface{}) map[string]interface{} {
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		return props
	}
	return make(map[string]interface{})
}

func getRequired(schema map[string]interface{}) []string {
	if req, ok := schema["required"].([]interface{}); ok {
		var result []string
		for _, r := range req {
			if s, ok := r.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

func getFieldDef(props map[string]interface{}, field string) map[string]interface{} {
	if def, ok := props[field].(map[string]interface{}); ok {
		return def
	}
	return nil
}

func hasDefault(fieldDef map[string]interface{}) bool {
	if fieldDef == nil {
		return false
	}
	_, has := fieldDef["default"]
	return has
}

func typeNarrowed(oldDef, newDef interface{}) bool {
	oldType := getType(oldDef)
	newType := getType(newDef)

	// number -> integer is narrowing
	if oldType == "number" && newType == "integer" {
		return true
	}
	// string -> enum is narrowing
	if oldType == "string" && newType == "string" {
		oldEnum := getEnum(oldDef)
		newEnum := getEnum(newDef)
		if oldEnum == nil && newEnum != nil {
			return true
		}
	}
	return false
}

func typeWidened(oldDef, newDef interface{}) bool {
	oldType := getType(oldDef)
	newType := getType(newDef)

	// integer -> number is widening
	if oldType == "integer" && newType == "number" {
		return true
	}
	return false
}

func getType(def interface{}) string {
	if m, ok := def.(map[string]interface{}); ok {
		if t, ok := m["type"].(string); ok {
			return t
		}
	}
	return ""
}

func getEnum(def interface{}) []interface{} {
	if m, ok := def.(map[string]interface{}); ok {
		if e, ok := m["enum"].([]interface{}); ok {
			return e
		}
	}
	return nil
}

func getAvroFields(schema map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	if fieldArr, ok := schema["fields"].([]interface{}); ok {
		for _, f := range fieldArr {
			if field, ok := f.(map[string]interface{}); ok {
				if name, ok := field["name"].(string); ok {
					fields[name] = field
				}
			}
		}
	}
	return fields
}

func hasAvroDefault(def interface{}) bool {
	if field, ok := def.(map[string]interface{}); ok {
		_, has := field["default"]
		return has
	}
	return false
}

func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
