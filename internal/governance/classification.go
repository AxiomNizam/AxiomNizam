package governance

// =====================================================
// WS-6.1 — Auto-Classification Rules Engine
//
// Automatically classifies catalog assets based on column
// names, data patterns, and metadata tags. Maps classifications
// to compliance frameworks (PII, PHI, Financial, Public).
// =====================================================

import (
	"regexp"
	"strings"
)

// ClassificationLevel represents a data sensitivity level.
type ClassificationLevel string

const (
	ClassPublic       ClassificationLevel = "public"
	ClassInternal     ClassificationLevel = "internal"
	ClassConfidential ClassificationLevel = "confidential"
	ClassRestricted   ClassificationLevel = "restricted"
)

// ClassificationTag represents a specific data type classification.
type ClassificationTag string

const (
	TagPII       ClassificationTag = "PII"
	TagPHI       ClassificationTag = "PHI"
	TagFinancial ClassificationTag = "Financial"
	TagAuth      ClassificationTag = "Authentication"
	TagPublic    ClassificationTag = "Public"
)

// ClassificationResult holds the result of classifying a single column or asset.
type ClassificationResult struct {
	ColumnName      string              `json:"columnName,omitempty"`
	AssetRef        string              `json:"assetRef,omitempty"`
	Level           ClassificationLevel `json:"level"`
	Tags            []ClassificationTag `json:"tags"`
	Confidence      float64             `json:"confidence"` // 0.0 - 1.0
	MatchedPatterns []string            `json:"matchedPatterns,omitempty"`
}

// ClassificationRule defines a single auto-classification pattern.
type ClassificationRule struct {
	Name        string              `json:"name"`
	Pattern     *regexp.Regexp      `json:"-"`
	PatternStr  string              `json:"pattern"`
	Level       ClassificationLevel `json:"level"`
	Tags        []ClassificationTag `json:"tags"`
	Confidence  float64             `json:"confidence"`
	Description string              `json:"description,omitempty"`
}

// Classifier applies auto-classification rules to assets and columns.
type Classifier struct {
	rules []ClassificationRule
}

// NewClassifier creates a classifier with the default PII/PHI/Financial detection rules.
func NewClassifier() *Classifier {
	c := &Classifier{}
	c.loadDefaultRules()
	return c
}

// ClassifyColumn classifies a single column by name.
func (c *Classifier) ClassifyColumn(columnName string) *ClassificationResult {
	result := &ClassificationResult{
		ColumnName: columnName,
		Level:      ClassPublic,
		Confidence: 0,
	}

	for _, rule := range c.rules {
		if rule.Pattern.MatchString(strings.ToLower(columnName)) {
			result.Tags = appendUniqueTags(result.Tags, rule.Tags...)
			result.MatchedPatterns = append(result.MatchedPatterns, rule.Name)
			if rule.Confidence > result.Confidence {
				result.Confidence = rule.Confidence
			}
			if isHigherLevel(rule.Level, result.Level) {
				result.Level = rule.Level
			}
		}
	}

	return result
}

// ClassifyColumns classifies a batch of columns and returns only those with non-public results.
func (c *Classifier) ClassifyColumns(columns []string) []ClassificationResult {
	var results []ClassificationResult
	for _, col := range columns {
		result := c.ClassifyColumn(col)
		if result.Level != ClassPublic || len(result.Tags) > 0 {
			results = append(results, *result)
		}
	}
	return results
}

// ClassifyAsset classifies an asset based on its columns and metadata.
func (c *Classifier) ClassifyAsset(assetRef string, columns []string, tags []string) *ClassificationResult {
	result := &ClassificationResult{
		AssetRef:   assetRef,
		Level:      ClassPublic,
		Confidence: 0,
	}

	// Classify each column and aggregate.
	for _, col := range columns {
		colResult := c.ClassifyColumn(col)
		result.Tags = appendUniqueTags(result.Tags, colResult.Tags...)
		if isHigherLevel(colResult.Level, result.Level) {
			result.Level = colResult.Level
		}
		if colResult.Confidence > result.Confidence {
			result.Confidence = colResult.Confidence
		}
		result.MatchedPatterns = append(result.MatchedPatterns, colResult.MatchedPatterns...)
	}

	// Check asset-level tags.
	for _, tag := range tags {
		lower := strings.ToLower(tag)
		if strings.Contains(lower, "pii") {
			result.Tags = appendUniqueTags(result.Tags, TagPII)
			result.Level = maxLevel(result.Level, ClassRestricted)
		}
		if strings.Contains(lower, "phi") || strings.Contains(lower, "health") {
			result.Tags = appendUniqueTags(result.Tags, TagPHI)
			result.Level = maxLevel(result.Level, ClassRestricted)
		}
		if strings.Contains(lower, "financial") || strings.Contains(lower, "payment") {
			result.Tags = appendUniqueTags(result.Tags, TagFinancial)
			result.Level = maxLevel(result.Level, ClassConfidential)
		}
	}

	return result
}

// AddRule adds a custom classification rule.
func (c *Classifier) AddRule(name, pattern string, level ClassificationLevel, tags []ClassificationTag, confidence float64) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	c.rules = append(c.rules, ClassificationRule{
		Name:       name,
		Pattern:    re,
		PatternStr: pattern,
		Level:      level,
		Tags:       tags,
		Confidence: confidence,
	})
	return nil
}

// --- Default classification rules ---

func (c *Classifier) loadDefaultRules() {
	defaults := []struct {
		name       string
		pattern    string
		level      ClassificationLevel
		tags       []ClassificationTag
		confidence float64
	}{
		// PII - Names
		{"name", `(first|last|full|middle|display|user)[-_]?name`, ClassRestricted, []ClassificationTag{TagPII}, 0.90},
		// PII - Contact
		{"email", `e[-_]?mail|email[-_]?addr`, ClassRestricted, []ClassificationTag{TagPII}, 0.95},
		{"phone", `phone|mobile|tel|fax|cell`, ClassRestricted, []ClassificationTag{TagPII}, 0.90},
		{"address", `(street|postal|zip|city|state|country|addr)`, ClassConfidential, []ClassificationTag{TagPII}, 0.80},
		// PII - IDs
		{"ssn", `ssn|social[-_]?sec|national[-_]?id|tax[-_]?id|nin`, ClassRestricted, []ClassificationTag{TagPII}, 0.99},
		{"passport", `passport`, ClassRestricted, []ClassificationTag{TagPII}, 0.95},
		{"driver_license", `driver|license[-_]?num`, ClassRestricted, []ClassificationTag{TagPII}, 0.90},
		// PII - Dates
		{"dob", `(date[-_]?of[-_]?birth|dob|birth[-_]?date)`, ClassConfidential, []ClassificationTag{TagPII}, 0.90},
		// PHI - Health
		{"diagnosis", `diagnos|icd|condition|symptom`, ClassRestricted, []ClassificationTag{TagPHI}, 0.85},
		{"medication", `medication|prescription|drug|dosage`, ClassRestricted, []ClassificationTag{TagPHI}, 0.85},
		{"medical_record", `medical[-_]?record|mrn|patient[-_]?id`, ClassRestricted, []ClassificationTag{TagPHI}, 0.95},
		{"insurance", `insurance|policy[-_]?num|claim`, ClassRestricted, []ClassificationTag{TagPHI, TagFinancial}, 0.80},
		// Financial
		{"credit_card", `(credit|debit)[-_]?card|card[-_]?num|pan|cvv|cvc`, ClassRestricted, []ClassificationTag{TagFinancial}, 0.99},
		{"bank_account", `(bank|account)[-_]?(num|no)|iban|swift|routing`, ClassRestricted, []ClassificationTag{TagFinancial}, 0.95},
		{"salary", `salary|compensation|wage|income|revenue`, ClassConfidential, []ClassificationTag{TagFinancial}, 0.80},
		// Authentication
		{"password", `password|passwd|pwd|secret|token|api[-_]?key`, ClassRestricted, []ClassificationTag{TagAuth}, 0.99},
		{"ip_address", `ip[-_]?addr|remote[-_]?addr|client[-_]?ip`, ClassInternal, []ClassificationTag{TagPII}, 0.70},
	}

	for _, d := range defaults {
		re := regexp.MustCompile(d.pattern)
		c.rules = append(c.rules, ClassificationRule{
			Name:       d.name,
			Pattern:    re,
			PatternStr: d.pattern,
			Level:      d.level,
			Tags:       d.tags,
			Confidence: d.confidence,
		})
	}
}

// --- Helpers ---

func appendUniqueTags(existing []ClassificationTag, tags ...ClassificationTag) []ClassificationTag {
	seen := make(map[ClassificationTag]bool)
	for _, t := range existing {
		seen[t] = true
	}
	result := existing
	for _, t := range tags {
		if !seen[t] {
			result = append(result, t)
			seen[t] = true
		}
	}
	return result
}

func isHigherLevel(a, b ClassificationLevel) bool {
	order := map[ClassificationLevel]int{ClassPublic: 0, ClassInternal: 1, ClassConfidential: 2, ClassRestricted: 3}
	return order[a] > order[b]
}

func maxLevel(a, b ClassificationLevel) ClassificationLevel {
	if isHigherLevel(a, b) {
		return a
	}
	return b
}
