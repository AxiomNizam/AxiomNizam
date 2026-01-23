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
