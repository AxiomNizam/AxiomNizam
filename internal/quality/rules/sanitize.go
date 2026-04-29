package rules

// =====================================================
// P1-1 — SQL Identifier Validation
//
// Prevents SQL injection in the quality rules engine by validating
// that all user-provided identifiers (table names, column names,
// asset references) contain only safe characters.
// =====================================================

import (
	"fmt"
	"regexp"
	"strings"
)

// validIdentifier matches safe SQL identifiers:
// letters, digits, underscores, dots (for schema.table), hyphens (for kebab-case names).
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.\-]*$`)

// maxIdentifierLen caps identifier length to prevent abuse.
const maxIdentifierLen = 128

// sqlKeywords that must never appear as bare identifiers.
var sqlKeywords = map[string]bool{
	"drop": true, "delete": true, "insert": true, "update": true,
	"alter": true, "create": true, "truncate": true, "exec": true,
	"execute": true, "grant": true, "revoke": true, "union": true,
	"into": true, "merge": true, "call": true,
}

// ValidateIdentifier checks that a string is a safe SQL identifier.
// Returns an error if the identifier contains SQL metacharacters,
// is too long, or matches a dangerous SQL keyword.
func ValidateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("empty identifier")
	}
	if len(name) > maxIdentifierLen {
		return fmt.Errorf("identifier too long: %d chars (max %d)", len(name), maxIdentifierLen)
	}
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("invalid identifier %q: must match [a-zA-Z_][a-zA-Z0-9_.\\-]*", name)
	}

	// Check each dot-separated part against SQL keywords.
	parts := strings.Split(strings.ToLower(name), ".")
	for _, part := range parts {
		if sqlKeywords[part] {
			return fmt.Errorf("identifier %q contains reserved SQL keyword %q", name, part)
		}
	}

	// Check for comment sequences.
	if strings.Contains(name, "--") || strings.Contains(name, "/*") || strings.Contains(name, "*/") {
		return fmt.Errorf("identifier %q contains SQL comment sequence", name)
	}

	// Check for semicolons (statement terminators).
	if strings.Contains(name, ";") {
		return fmt.Errorf("identifier %q contains semicolon", name)
	}

	return nil
}

// ValidateAssetRef validates a table/asset reference (may contain schema.table notation).
func ValidateAssetRef(ref string) error {
	if ref == "" {
		return fmt.Errorf("empty asset reference")
	}
	return ValidateIdentifier(ref)
}

// ValidateColumn validates a column name.
func ValidateColumn(col string) error {
	if col == "" {
		return fmt.Errorf("empty column name")
	}
	return ValidateIdentifier(col)
}

// ValidateRuleInputs validates all user-provided identifiers in a quality rule spec.
// Call this before executing any SQL queries.
func ValidateRuleInputs(rule *QualityRuleResource) error {
	if err := ValidateAssetRef(rule.Spec.AssetRef); err != nil {
		return fmt.Errorf("invalid assetRef: %w", err)
	}
	if err := ValidateIdentifier(rule.Spec.DataSourceRef); err != nil {
		return fmt.Errorf("invalid dataSourceRef: %w", err)
	}

	// Validate rule-type-specific fields.
	switch rule.Spec.RuleType {
	case RuleTypeFreshness:
		if rule.Spec.Freshness != nil {
			if err := ValidateColumn(rule.Spec.Freshness.TimestampColumn); err != nil {
				return fmt.Errorf("invalid freshness.timestampColumn: %w", err)
			}
		}
	case RuleTypeNotNull:
		if rule.Spec.NotNull != nil {
			if err := ValidateColumn(rule.Spec.NotNull.Column); err != nil {
				return fmt.Errorf("invalid notNull.column: %w", err)
			}
		}
	case RuleTypeUnique:
		if rule.Spec.Unique != nil {
			if err := ValidateColumn(rule.Spec.Unique.Column); err != nil {
				return fmt.Errorf("invalid unique.column: %w", err)
			}
		}
	case RuleTypeRange:
		if rule.Spec.Range != nil {
			if err := ValidateColumn(rule.Spec.Range.Column); err != nil {
				return fmt.Errorf("invalid range.column: %w", err)
			}
		}
	case RuleTypeCompleteness:
		if rule.Spec.Completeness != nil {
			if err := ValidateColumn(rule.Spec.Completeness.Column); err != nil {
				return fmt.Errorf("invalid completeness.column: %w", err)
			}
		}
	case RuleTypeStatistical:
		if rule.Spec.Statistical != nil {
			if err := ValidateColumn(rule.Spec.Statistical.Column); err != nil {
				return fmt.Errorf("invalid statistical.column: %w", err)
			}
		}
	case RuleTypeCustomSQL:
		// Custom SQL is inherently user-provided — we can't validate the full query,
		// but we can reject obvious injection patterns.
		if rule.Spec.CustomSQL != nil {
			q := strings.ToLower(rule.Spec.CustomSQL.Query)
			for _, dangerous := range []string{"drop ", "delete ", "insert ", "update ", "alter ", "truncate ", "grant ", "revoke "} {
				if strings.Contains(q, dangerous) {
					return fmt.Errorf("custom SQL contains forbidden keyword %q", strings.TrimSpace(dangerous))
				}
			}
		}
	}

	return nil
}
