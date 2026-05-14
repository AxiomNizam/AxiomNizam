package catalog

// Metadata enrichment: column profiling, PII detection, and auto-classification.
//
// The enrichment reconciler runs on a schedule and profiles catalog assets,
// detecting PII patterns, computing column statistics, and auto-classifying
// data sensitivity levels.

import (
	"regexp"
	"strings"
)

// PIIPattern defines a regex pattern for detecting PII in column names or values.
type PIIPattern struct {
	Name        string
	Category    string // PII, PHI, Financial, Credentials
	ColumnRegex *regexp.Regexp
	ValueRegex  *regexp.Regexp
}

// DefaultPIIPatterns returns the built-in PII detection patterns.
func DefaultPIIPatterns() []PIIPattern {
	return []PIIPattern{
		{Name: "email", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(email|e_mail|email_address|mail)`)},
		{Name: "phone", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(phone|telephone|mobile|cell|fax)`)},
		{Name: "ssn", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(ssn|social_security|sin|national_id)`)},
		{Name: "credit_card", Category: "Financial", ColumnRegex: regexp.MustCompile(`(?i)(credit_card|card_number|cc_num|pan)`)},
		{Name: "ip_address", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(ip_address|ip_addr|client_ip|remote_ip)`)},
		{Name: "address", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(address|street|city|zip|postal|zip_code)`)},
		{Name: "name", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(first_name|last_name|full_name|surname|given_name)`)},
		{Name: "dob", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(date_of_birth|dob|birth_date|birthday)`)},
		{Name: "passport", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(passport|passport_number|passport_no)`)},
		{Name: "driver_license", Category: "PII", ColumnRegex: regexp.MustCompile(`(?i)(driver_license|license_number|dl_number)`)},
		{Name: "bank_account", Category: "Financial", ColumnRegex: regexp.MustCompile(`(?i)(bank_account|account_number|iban|routing)`)},
		{Name: "salary", Category: "Financial", ColumnRegex: regexp.MustCompile(`(?i)(salary|wage|compensation|income|pay_rate)`)},
		{Name: "password", Category: "Credentials", ColumnRegex: regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|api_key)`)},
		{Name: "medical", Category: "PHI", ColumnRegex: regexp.MustCompile(`(?i)(diagnosis|medication|treatment|medical|health|patient)`)},
		{Name: "genetic", Category: "PHI", ColumnRegex: regexp.MustCompile(`(?i)(genetic|dna|genome|biometric)`)},
	}
}

// DetectPII checks a column name against PII patterns and returns detected categories.
func DetectPII(columnName string, patterns []PIIPattern) (bool, []string) {
	var categories []string
	isPII := false

	for _, pattern := range patterns {
		if pattern.ColumnRegex != nil && pattern.ColumnRegex.MatchString(columnName) {
			isPII = true
			categories = append(categories, pattern.Category)
		}
	}

	return isPII, uniqueStrings(categories)
}

// AutoClassify determines the data classification level based on detected PII.
func AutoClassify(columns []CatalogColumn, patterns []PIIPattern) DataClassification {
	classification := DataClassification{
		Level: "public",
	}

	var allCategories []string
	hasPII := false
	hasPHI := false
	hasFinancial := false
	hasCredentials := false

	for i := range columns {
		isPII, categories := DetectPII(columns[i].Name, patterns)
		if isPII {
			columns[i].IsPII = true
			columns[i].Classification = strings.Join(categories, ",")
			hasPII = true
			allCategories = append(allCategories, categories...)

			for _, cat := range categories {
				switch cat {
				case "PHI":
					hasPHI = true
				case "Financial":
					hasFinancial = true
				case "Credentials":
					hasCredentials = true
				}
			}
		}
	}

	classification.Categories = uniqueStrings(allCategories)
	classification.PII = hasPII

	// Determine level based on what was found.
	if hasCredentials || hasPHI {
		classification.Level = "restricted"
		classification.Sensitive = true
	} else if hasFinancial || hasPII {
		classification.Level = "confidential"
		classification.Sensitive = true
	} else if len(allCategories) > 0 {
		classification.Level = "internal"
	}

	return classification
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
