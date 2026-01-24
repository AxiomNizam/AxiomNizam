package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsValidPhone validates phone number format (basic)
func IsValidPhone(phone string) bool {
	pattern := `^\+?[1-9]\d{1,14}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsValidURL validates URL format
func IsValidURL(url string) bool {
	pattern := `^https?://[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// IsValidIPAddress validates IPv4 address
func IsValidIPAddress(ip string) bool {
	pattern := `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, _ := regexp.MatchString(pattern, ip)
	return matched
}

// IsValidUUID validates UUID format
func IsValidUUID(uuid string) bool {
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(uuid))
	return matched
}

// IsValidHexColor validates hex color format
func IsValidHexColor(color string) bool {
	pattern := `^#?([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`
	matched, _ := regexp.MatchString(pattern, color)
	return matched
}

// IsValidJSON checks if string is valid JSON
func IsValidJSON(jsonStr string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// IsValidPassword validates password strength
// Returns true if password has at least 8 chars, 1 uppercase, 1 lowercase, 1 digit, 1 special char
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsValidUsername validates username format
// Allows alphanumeric, underscore, hyphen (3-20 chars)
func IsValidUsername(username string) bool {
	pattern := `^[a-zA-Z0-9_-]{3,20}$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// IsValidDomain validates domain name format
func IsValidDomain(domain string) bool {
	pattern := `^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(domain))
	return matched
}

// IsValidPort validates port number (1-65535)
func IsValidPort(port string) bool {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return portNum > 0 && portNum <= 65535
}

// IsValidInt checks if string is valid integer
func IsValidInt(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// IsValidFloat checks if string is valid float
func IsValidFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsValidBoolean checks if string is valid boolean
func IsValidBoolean(s string) bool {
	_, err := strconv.ParseBool(s)
	return err == nil
}

// IsValidLength checks if string length is within range
func IsValidLength(s string, minLength, maxLength int) bool {
	length := StringLength(s)
	return length >= minLength && length <= maxLength
}

// IsValidAlpha checks if string contains only alphabetic characters
func IsValidAlpha(s string) bool {
	pattern := `^[a-zA-Z]+$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// IsValidAlphaNumeric checks if string contains only alphanumeric characters
func IsValidAlphaNumeric(s string) bool {
	pattern := `^[a-zA-Z0-9]+$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// IsValidNumeric checks if string contains only digits
func IsValidNumeric(s string) bool {
	pattern := `^[0-9]+$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// IsValidSlug checks if string is valid URL slug
func IsValidSlug(slug string) bool {
	pattern := `^[a-z0-9]+(?:-[a-z0-9]+)*$`
	matched, _ := regexp.MatchString(pattern, slug)
	return matched
}

// IsValidCreditCard validates credit card number using Luhn algorithm
func IsValidCreditCard(cardNumber string) bool {
	// Remove spaces and hyphens
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	if !IsValidNumeric(cardNumber) || len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}

	// Luhn algorithm
	sum := 0
	isEven := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

// ValidateStruct validates a struct with basic checks
type ValidatorFunc func(interface{}) error

// ValidateRequired checks if value is not empty
func ValidateRequired(value interface{}) error {
	switch v := value.(type) {
	case string:
		if IsEmpty(v) {
			return fmt.Errorf("field is required")
		}
	case int, int64:
		if v == 0 {
			return fmt.Errorf("field is required")
		}
	}
	return nil
}

// ValidateMinLength validates minimum string length
func ValidateMinLength(minLength int) ValidatorFunc {
	return func(value interface{}) error {
		if s, ok := value.(string); ok {
			if StringLength(s) < minLength {
				return fmt.Errorf("must be at least %d characters", minLength)
			}
		}
		return nil
	}
}

// ValidateMaxLength validates maximum string length
func ValidateMaxLength(maxLength int) ValidatorFunc {
	return func(value interface{}) error {
		if s, ok := value.(string); ok {
			if StringLength(s) > maxLength {
				return fmt.Errorf("must be at most %d characters", maxLength)
			}
		}
		return nil
	}
}

// ValidateEmail validates email format
func ValidateEmail(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidEmail(s) {
			return fmt.Errorf("invalid email format")
		}
	}
	return nil
}

// ValidateURL validates URL format
func ValidateURL(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidURL(s) {
			return fmt.Errorf("invalid URL format")
		}
	}
	return nil
}

// SanitizeInput removes potentially dangerous characters from user input
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters (except newlines in some cases)
	for i := 0; i < 32; i++ {
		if i != 9 && i != 10 && i != 13 { // Keep tab, newline, carriage return
			input = strings.ReplaceAll(input, string(rune(i)), "")
		}
	}

	return input
}

// SanitizeHTMLInput removes HTML/JavaScript that could cause XSS
func SanitizeHTMLInput(input string) string {
	// Remove potentially dangerous HTML tags
	dangerous := []string{
		"<script", "</script>",
		"<iframe", "</iframe>",
		"<object", "</object>",
		"<embed", "</embed>",
		"<link", "</link>",
		"<meta", "</meta>",
		"<style", "</style>",
		"<form", "</form>",
	}

	result := input
	for _, tag := range dangerous {
		pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(tag))
		result = pattern.ReplaceAllString(result, "")
	}

	// Remove dangerous event handlers
	eventHandlers := []string{
		"onclick", "onload", "onerror", "onmouseover",
		"onmouseout", "onkeydown", "onkeyup", "onfocus",
		"onblur", "onchange", "onsubmit",
	}

	for _, handler := range eventHandlers {
		pattern := regexp.MustCompile(`(?i)` + handler + `\s*=`)
		result = pattern.ReplaceAllString(result, "")
	}

	return result
}

// ValidateJSONInput validates if input is valid JSON
func ValidateJSONInput(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidJSON(s) {
			return fmt.Errorf("invalid JSON format")
		}
	}
	return nil
}

// ValidateIPInput validates IP address input
func ValidateIPInput(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidIPAddress(s) {
			return fmt.Errorf("invalid IP address format")
		}
	}
	return nil
}

// ValidateUUIDInput validates UUID format
func ValidateUUIDInput(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidUUID(s) {
			return fmt.Errorf("invalid UUID format")
		}
	}
	return nil
}

// ValidateDomainInput validates domain name
func ValidateDomainInput(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidDomain(s) {
			return fmt.Errorf("invalid domain format")
		}
	}
	return nil
}

// ValidateNumericInput validates numeric string
func ValidateNumericInput(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidNumeric(s) {
			return fmt.Errorf("input must contain only digits")
		}
	}
	return nil
}

// ValidateAlphaNumericInput validates alphanumeric input
func ValidateAlphaNumericInput(value interface{}) error {
	if s, ok := value.(string); ok {
		if !IsValidAlphaNumeric(s) {
			return fmt.Errorf("input must be alphanumeric only")
		}
	}
	return nil
}

// ValidateNoSpecialChars validates that input has no special characters
func ValidateNoSpecialChars(value interface{}) error {
	if s, ok := value.(string); ok {
		pattern := `[^a-zA-Z0-9\s_-]`
		matched, _ := regexp.MatchString(pattern, s)
		if matched {
			return fmt.Errorf("input contains invalid special characters")
		}
	}
	return nil
}

// ValidateFileNameInput validates filename format
func ValidateFileNameInput(filename string) error {
	if IsEmpty(filename) {
		return fmt.Errorf("filename cannot be empty")
	}

	// Disallow directory traversal attempts
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename - directory traversal detected")
	}

	// Allow alphanumeric, dots, hyphens, underscores
	pattern := `^[a-zA-Z0-9._-]+$`
	matched, _ := regexp.MatchString(pattern, filename)
	if !matched {
		return fmt.Errorf("filename contains invalid characters")
	}

	return nil
}

// ValidateDatabaseIdentifier validates table/column names
func ValidateDatabaseIdentifier(identifier string) error {
	if IsEmpty(identifier) {
		return fmt.Errorf("identifier cannot be empty")
	}

	// Only allow alphanumeric and underscores, must start with letter or underscore
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	matched, _ := regexp.MatchString(pattern, identifier)
	if !matched {
		return fmt.Errorf("invalid database identifier format")
	}

	return nil
}

// ValidatePath validates file path for security
func ValidatePath(path string) error {
	if IsEmpty(path) {
		return fmt.Errorf("path cannot be empty")
	}

	// Prevent directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	return nil
}

// ValidateAPIKey validates API key format
func ValidateAPIKey(key string) error {
	if IsEmpty(key) {
		return fmt.Errorf("API key cannot be empty")
	}

	// API keys are typically alphanumeric with some special chars like -_
	pattern := `^[a-zA-Z0-9_-]{20,}$`
	matched, _ := regexp.MatchString(pattern, key)
	if !matched {
		return fmt.Errorf("invalid API key format")
	}

	return nil
}

// ValidateURLPath validates URL path component
func ValidateURLPath(path string) error {
	if IsEmpty(path) {
		return fmt.Errorf("URL path cannot be empty")
	}

	// Allow alphanumeric, forward slashes, hyphens, underscores, dots
	pattern := `^[a-zA-Z0-9/_\-\.]*$`
	matched, _ := regexp.MatchString(pattern, path)
	if !matched {
		return fmt.Errorf("invalid URL path format")
	}

	return nil
}

// SanitizeAndValidate performs both sanitization and validation
func SanitizeAndValidate(input string, validators ...ValidatorFunc) error {
	// First sanitize the input
	sanitized := SanitizeInput(input)

	// Then run validation functions
	for _, validator := range validators {
		if err := validator(sanitized); err != nil {
			return err
		}
	}

	return nil
}
