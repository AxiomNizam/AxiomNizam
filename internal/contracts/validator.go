package contracts

// =====================================================
// WS-2.2 — Data Contract Breaking Change Validator
//
// Detects breaking changes between contract versions based on
// the configured compatibility mode.
//
// Breaking change rules:
//   Backward: new schema can read old data
//   Forward:  old schema can read new data
//   Full:     both directions
// =====================================================

import (
	"fmt"
	"strings"
)

// BreakingChange represents a detected breaking change between contract versions.
type BreakingChange struct {
	Type        string `json:"type"`        // added_required, removed_column, type_narrowed, etc.
	Column      string `json:"column"`
	Description string `json:"description"`
	Severity    string `json:"severity"`    // breaking, warning
}

// ContractValidator validates schema changes between contract versions.
type ContractValidator struct{}

// NewContractValidator creates a new validator.
func NewContractValidator() *ContractValidator {
	return &ContractValidator{}
}

// ValidateChange checks if a schema change is compatible with the contract's compatibility mode.
func (v *ContractValidator) ValidateChange(
	oldSchema ContractSchema,
	newSchema ContractSchema,
	mode CompatibilityMode,
) []BreakingChange {
	switch mode {
	case CompatBackward:
		return v.checkBackward(oldSchema, newSchema)
	case CompatForward:
		return v.checkForward(oldSchema, newSchema)
	case CompatFull:
		backward := v.checkBackward(oldSchema, newSchema)
		forward := v.checkForward(oldSchema, newSchema)
		return append(backward, forward...)
	case CompatNone:
		return nil
	default:
		return nil
	}
}

// checkBackward ensures new schema can read data written with old schema.
// Forbidden: remove required column, add required column without default, narrow type.
func (v *ContractValidator) checkBackward(oldSchema, newSchema ContractSchema) []BreakingChange {
	var changes []BreakingChange

	newMap := columnMap(newSchema.Columns)
	oldMap := columnMap(oldSchema.Columns)

	// Check for removed columns.
	for _, oldCol := range oldSchema.Columns {
		if _, exists := newMap[strings.ToLower(oldCol.Name)]; !exists {
			if oldCol.Required {
				changes = append(changes, BreakingChange{
					Type:        "removed_required_column",
					Column:      oldCol.Name,
					Description: fmt.Sprintf("required column '%s' was removed (backward incompatible)", oldCol.Name),
					Severity:    "breaking",
				})
			}
		}
	}

	// Check for new required columns (old data won't have them).
	for _, newCol := range newSchema.Columns {
		if _, existed := oldMap[strings.ToLower(newCol.Name)]; !existed {
			if newCol.Required && !newCol.Nullable {
				changes = append(changes, BreakingChange{
					Type:        "added_required_column",
					Column:      newCol.Name,
					Description: fmt.Sprintf("new required non-nullable column '%s' added (old data won't have it)", newCol.Name),
					Severity:    "breaking",
				})
			}
		}
	}

	// Check for type narrowing.
	for _, oldCol := range oldSchema.Columns {
		if newCol, exists := newMap[strings.ToLower(oldCol.Name)]; exists {
			if oldCol.Type != "" && newCol.Type != "" && isTypeNarrowed(oldCol.Type, newCol.Type) {
				changes = append(changes, BreakingChange{
					Type:        "type_narrowed",
					Column:      oldCol.Name,
					Description: fmt.Sprintf("column '%s' type narrowed from '%s' to '%s'", oldCol.Name, oldCol.Type, newCol.Type),
					Severity:    "breaking",
				})
			}
		}
	}

	// Check for nullability change (nullable -> not nullable).
	for _, oldCol := range oldSchema.Columns {
		if newCol, exists := newMap[strings.ToLower(oldCol.Name)]; exists {
			if oldCol.Nullable && !newCol.Nullable {
				changes = append(changes, BreakingChange{
					Type:        "nullability_restricted",
					Column:      oldCol.Name,
					Description: fmt.Sprintf("column '%s' changed from nullable to required", oldCol.Name),
					Severity:    "breaking",
				})
			}
		}
	}

	return changes
}

// checkForward ensures old schema can read data written with new schema.
// Forbidden: add column without default (old reader fails), remove required column.
func (v *ContractValidator) checkForward(oldSchema, newSchema ContractSchema) []BreakingChange {
	var changes []BreakingChange

	newMap := columnMap(newSchema.Columns)
	oldRequired := make(map[string]bool)
	for _, col := range oldSchema.RequiredColumns {
		oldRequired[strings.ToLower(col)] = true
	}
	for _, col := range oldSchema.Columns {
		if col.Required {
			oldRequired[strings.ToLower(col.Name)] = true
		}
	}

	// Check for removed columns that old readers expect.
	for _, oldCol := range oldSchema.Columns {
		if _, exists := newMap[strings.ToLower(oldCol.Name)]; !exists {
			if oldRequired[strings.ToLower(oldCol.Name)] {
				changes = append(changes, BreakingChange{
					Type:        "removed_expected_column",
					Column:      oldCol.Name,
					Description: fmt.Sprintf("column '%s' removed but old readers expect it (forward incompatible)", oldCol.Name),
					Severity:    "breaking",
				})
			}
		}
	}

	// Check for type widening (old reader may not handle wider type).
	for _, oldCol := range oldSchema.Columns {
		if newCol, exists := newMap[strings.ToLower(oldCol.Name)]; exists {
			if oldCol.Type != "" && newCol.Type != "" && isTypeWidened(oldCol.Type, newCol.Type) {
				changes = append(changes, BreakingChange{
					Type:        "type_widened",
					Column:      oldCol.Name,
					Description: fmt.Sprintf("column '%s' type widened from '%s' to '%s' (old readers may not handle)", oldCol.Name, oldCol.Type, newCol.Type),
					Severity:    "warning",
				})
			}
		}
	}

	return changes
}

// columnMap creates a lookup map from column name (lowercase) to column.
func columnMap(columns []ContractColumn) map[string]ContractColumn {
	m := make(map[string]ContractColumn, len(columns))
	for _, col := range columns {
		m[strings.ToLower(col.Name)] = col
	}
	return m
}

// isTypeNarrowed checks if a type change represents narrowing.
func isTypeNarrowed(oldType, newType string) bool {
	old := strings.ToLower(oldType)
	new_ := strings.ToLower(newType)

	narrowings := map[string]string{
		"bigint":  "int",
		"double":  "float",
		"text":    "varchar",
		"float64": "float32",
		"int64":   "int32",
		"number":  "integer",
	}

	for wide, narrow := range narrowings {
		if (old == wide || strings.HasPrefix(old, wide)) &&
			(new_ == narrow || strings.HasPrefix(new_, narrow)) {
			return true
		}
	}
	return false
}

// isTypeWidened checks if a type change represents widening.
func isTypeWidened(oldType, newType string) bool {
	// Widening is the reverse of narrowing.
	return isTypeNarrowed(newType, oldType)
}
