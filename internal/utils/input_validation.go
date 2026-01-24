package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// InputValidator provides comprehensive input validation utilities
type InputValidator struct {
	maxStringLength int
	maxIntValue     int64
	minIntValue     int64
}

// NewInputValidator creates a new input validator with default settings
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxStringLength: 5000,
		maxIntValue:     9223372036854775807,  // Max int64
		minIntValue:     -9223372036854775808, // Min int64
	}
}

// ValidateString validates a string input with optional constraints
func (iv *InputValidator) ValidateString(input string, opts ...ValidationOption) error {
	if IsEmpty(input) {
		return fmt.Errorf("string cannot be empty")
	}

	// Apply default option
	vo := &validationOptions{
		maxLength: iv.maxStringLength,
		minLength: 0,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(vo)
	}

	// Check length
	length := StringLength(input)
	if length < vo.minLength {
		return fmt.Errorf("string must be at least %d characters", vo.minLength)
	}
	if length > vo.maxLength {
		return fmt.Errorf("string exceeds maximum length of %d characters", vo.maxLength)
	}

	// Check pattern if specified
	if vo.pattern != "" {
		matched, _ := regexp.MatchString(vo.pattern, input)
		if !matched {
			return fmt.Errorf("string does not match required pattern")
		}
	}

	// Check allowed characters
	if len(vo.allowedChars) > 0 {
		pattern := fmt.Sprintf("^[%s]+$", regexp.QuoteMeta(vo.allowedChars))
		matched, _ := regexp.MatchString(pattern, input)
		if !matched {
			return fmt.Errorf("string contains disallowed characters")
		}
	}

	// Check forbidden characters
	if len(vo.forbiddenChars) > 0 {
		pattern := fmt.Sprintf("[%s]", regexp.QuoteMeta(vo.forbiddenChars))
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return fmt.Errorf("string contains forbidden characters")
		}
	}

	return nil
}

// ValidateInteger validates an integer value
func (iv *InputValidator) ValidateInteger(input interface{}, opts ...ValidationOption) error {
	var intVal int64

	switch v := input.(type) {
	case int:
		intVal = int64(v)
	case int64:
		intVal = v
	case string:
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer format: %v", input)
		}
		intVal = val
	default:
		return fmt.Errorf("unsupported type for integer validation: %T", input)
	}

	// Apply options
	vo := &validationOptions{
		minValue: iv.minIntValue,
		maxValue: iv.maxIntValue,
	}

	for _, opt := range opts {
		opt(vo)
	}

	// Check range
	if intVal < vo.minValue {
		return fmt.Errorf("value must be at least %d", vo.minValue)
	}
	if intVal > vo.maxValue {
		return fmt.Errorf("value must be at most %d", vo.maxValue)
	}

	return nil
}

// ValidateEmail validates email format with optional domain check
func (iv *InputValidator) ValidateEmail(email string, opts ...ValidationOption) error {
	if IsEmpty(email) {
		return fmt.Errorf("email cannot be empty")
	}

	if !IsValidEmail(email) {
		return fmt.Errorf("invalid email format")
	}

	// Check length
	if StringLength(email) > 254 {
		return fmt.Errorf("email exceeds maximum length of 254 characters")
	}

	return nil
}

// ValidatePassword validates password strength
func (iv *InputValidator) ValidatePassword(password string, opts ...ValidationOption) error {
	if IsEmpty(password) {
		return fmt.Errorf("password cannot be empty")
	}

	vo := &validationOptions{
		minLength:           8,
		requireUppercase:    true,
		requireLowercase:    true,
		requireNumbers:      true,
		requireSpecialChars: true,
	}

	for _, opt := range opts {
		opt(vo)
	}

	length := StringLength(password)
	if length < vo.minLength {
		return fmt.Errorf("password must be at least %d characters", vo.minLength)
	}

	if vo.requireUppercase && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if vo.requireLowercase && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if vo.requireNumbers && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if vo.requireSpecialChars && !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateUsername validates username format
func (iv *InputValidator) ValidateUsername(username string, opts ...ValidationOption) error {
	if IsEmpty(username) {
		return fmt.Errorf("username cannot be empty")
	}

	vo := &validationOptions{
		minLength: 3,
		maxLength: 20,
		pattern:   `^[a-zA-Z0-9_-]+$`,
	}

	for _, opt := range opts {
		opt(vo)
	}

	length := StringLength(username)
	if length < vo.minLength {
		return fmt.Errorf("username must be at least %d characters", vo.minLength)
	}
	if length > vo.maxLength {
		return fmt.Errorf("username must be at most %d characters", vo.maxLength)
	}

	matched, _ := regexp.MatchString(vo.pattern, username)
	if !matched {
		return fmt.Errorf("username contains invalid characters")
	}

	return nil
}

// ValidateURL validates and sanitizes URL
func (iv *InputValidator) ValidateURL(url string, opts ...ValidationOption) error {
	if IsEmpty(url) {
		return fmt.Errorf("URL cannot be empty")
	}

	if !IsValidURL(url) {
		return fmt.Errorf("invalid URL format")
	}

	if StringLength(url) > 2048 {
		return fmt.Errorf("URL exceeds maximum length of 2048 characters")
	}

	return nil
}

// ValidateIPAddress validates IP address
func (iv *InputValidator) ValidateIPAddress(ip string, opts ...ValidationOption) error {
	if IsEmpty(ip) {
		return fmt.Errorf("IP address cannot be empty")
	}

	if !IsValidIPAddress(ip) {
		return fmt.Errorf("invalid IP address format")
	}

	return nil
}

// ValidatePhoneNumber validates phone number
func (iv *InputValidator) ValidatePhoneNumber(phone string, opts ...ValidationOption) error {
	if IsEmpty(phone) {
		return fmt.Errorf("phone number cannot be empty")
	}

	if !IsValidPhone(phone) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// ValidateDate validates date format (YYYY-MM-DD)
func (iv *InputValidator) ValidateDate(date string, opts ...ValidationOption) error {
	if IsEmpty(date) {
		return fmt.Errorf("date cannot be empty")
	}

	pattern := `^\d{4}-\d{2}-\d{2}$`
	matched, _ := regexp.MatchString(pattern, date)
	if !matched {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}

	return nil
}

// ValidateTime validates time format (HH:MM:SS)
func (iv *InputValidator) ValidateTime(time string, opts ...ValidationOption) error {
	if IsEmpty(time) {
		return fmt.Errorf("time cannot be empty")
	}

	pattern := `^([0-1]?[0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$`
	matched, _ := regexp.MatchString(pattern, time)
	if !matched {
		return fmt.Errorf("invalid time format, expected HH:MM:SS")
	}

	return nil
}

// ValidateJSON validates JSON format
func (iv *InputValidator) ValidateJSON(jsonStr string, opts ...ValidationOption) error {
	if IsEmpty(jsonStr) {
		return fmt.Errorf("JSON string cannot be empty")
	}

	if !IsValidJSON(jsonStr) {
		return fmt.Errorf("invalid JSON format")
	}

	return nil
}

// ValidateArray validates array input
func (iv *InputValidator) ValidateArray(arr []string, opts ...ValidationOption) error {
	if len(arr) == 0 {
		return fmt.Errorf("array cannot be empty")
	}

	vo := &validationOptions{
		minLength: 1,
		maxLength: 1000,
	}

	for _, opt := range opts {
		opt(vo)
	}

	if len(arr) < vo.minLength {
		return fmt.Errorf("array must have at least %d items", vo.minLength)
	}
	if len(arr) > vo.maxLength {
		return fmt.Errorf("array must have at most %d items", vo.maxLength)
	}

	return nil
}

// ValidateMap validates map input
func (iv *InputValidator) ValidateMap(m map[string]interface{}, opts ...ValidationOption) error {
	if len(m) == 0 {
		return fmt.Errorf("map cannot be empty")
	}

	vo := &validationOptions{
		maxLength: 1000,
	}

	for _, opt := range opts {
		opt(vo)
	}

	if len(m) > vo.maxLength {
		return fmt.Errorf("map exceeds maximum size of %d items", vo.maxLength)
	}

	return nil
}

// Batch validation for multiple inputs
type ValidationBatch struct {
	validator *InputValidator
	errors    []ValidationError
}

// NewValidationBatch creates a new validation batch
func (iv *InputValidator) NewValidationBatch() *ValidationBatch {
	return &ValidationBatch{
		validator: iv,
		errors:    []ValidationError{},
	}
}

// AddStringValidation adds a string validation to the batch
func (vb *ValidationBatch) AddStringValidation(fieldName, value string, opts ...ValidationOption) *ValidationBatch {
	if err := vb.validator.ValidateString(value, opts...); err != nil {
		vb.errors = append(vb.errors, ValidationError{
			Field:   fieldName,
			Message: err.Error(),
		})
	}
	return vb
}

// AddEmailValidation adds an email validation to the batch
func (vb *ValidationBatch) AddEmailValidation(fieldName, value string) *ValidationBatch {
	if err := vb.validator.ValidateEmail(value); err != nil {
		vb.errors = append(vb.errors, ValidationError{
			Field:   fieldName,
			Message: err.Error(),
		})
	}
	return vb
}

// AddPasswordValidation adds a password validation to the batch
func (vb *ValidationBatch) AddPasswordValidation(fieldName, value string, opts ...ValidationOption) *ValidationBatch {
	if err := vb.validator.ValidatePassword(value, opts...); err != nil {
		vb.errors = append(vb.errors, ValidationError{
			Field:   fieldName,
			Message: err.Error(),
		})
	}
	return vb
}

// AddIntegerValidation adds an integer validation to the batch
func (vb *ValidationBatch) AddIntegerValidation(fieldName string, value interface{}, opts ...ValidationOption) *ValidationBatch {
	if err := vb.validator.ValidateInteger(value, opts...); err != nil {
		vb.errors = append(vb.errors, ValidationError{
			Field:   fieldName,
			Message: err.Error(),
		})
	}
	return vb
}

// GetErrors returns all validation errors
func (vb *ValidationBatch) GetErrors() []ValidationError {
	return vb.errors
}

// HasErrors checks if there are any validation errors
func (vb *ValidationBatch) HasErrors() bool {
	return len(vb.errors) > 0
}

// Error returns formatted error message
func (vb *ValidationBatch) Error() string {
	if !vb.HasErrors() {
		return ""
	}

	var messages []string
	for _, err := range vb.errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	return strings.Join(messages, "; ")
}

// ValidationOption is a functional option for validation
type ValidationOption func(*validationOptions)

// validationOptions holds validation configuration
type validationOptions struct {
	minLength           int
	maxLength           int
	minValue            int64
	maxValue            int64
	pattern             string
	allowedChars        string
	forbiddenChars      string
	requireUppercase    bool
	requireLowercase    bool
	requireNumbers      bool
	requireSpecialChars bool
}

// WithMinLength sets minimum length
func WithMinLength(minLength int) ValidationOption {
	return func(vo *validationOptions) {
		vo.minLength = minLength
	}
}

// WithMaxLength sets maximum length
func WithMaxLength(maxLength int) ValidationOption {
	return func(vo *validationOptions) {
		vo.maxLength = maxLength
	}
}

// WithMinValue sets minimum value
func WithMinValue(minValue int64) ValidationOption {
	return func(vo *validationOptions) {
		vo.minValue = minValue
	}
}

// WithMaxValue sets maximum value
func WithMaxValue(maxValue int64) ValidationOption {
	return func(vo *validationOptions) {
		vo.maxValue = maxValue
	}
}

// WithPattern sets validation pattern
func WithPattern(pattern string) ValidationOption {
	return func(vo *validationOptions) {
		vo.pattern = pattern
	}
}

// WithAllowedChars sets allowed characters
func WithAllowedChars(chars string) ValidationOption {
	return func(vo *validationOptions) {
		vo.allowedChars = chars
	}
}

// WithForbiddenChars sets forbidden characters
func WithForbiddenChars(chars string) ValidationOption {
	return func(vo *validationOptions) {
		vo.forbiddenChars = chars
	}
}

// WithRequireUppercase requires uppercase letters
func WithRequireUppercase(require bool) ValidationOption {
	return func(vo *validationOptions) {
		vo.requireUppercase = require
	}
}

// WithRequireLowercase requires lowercase letters
func WithRequireLowercase(require bool) ValidationOption {
	return func(vo *validationOptions) {
		vo.requireLowercase = require
	}
}

// WithRequireNumbers requires numbers
func WithRequireNumbers(require bool) ValidationOption {
	return func(vo *validationOptions) {
		vo.requireNumbers = require
	}
}

// WithRequireSpecialChars requires special characters
func WithRequireSpecialChars(require bool) ValidationOption {
	return func(vo *validationOptions) {
		vo.requireSpecialChars = require
	}
}
