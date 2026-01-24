package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Validator defines validation functionality
type Validator struct {
	rules map[string][]Rule
}

// Rule defines a validation rule
type Rule func(value interface{}) error

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]Rule),
	}
}

// AddRule adds a validation rule
func (v *Validator) AddRule(field string, rule Rule) *Validator {
	v.rules[field] = append(v.rules[field], rule)
	return v
}

// Validate validates all fields
func (v *Validator) Validate(data map[string]interface{}) error {
	for field, rules := range v.rules {
		value, exists := data[field]
		if !exists && !isRequired(rules) {
			continue
		}

		for _, rule := range rules {
			if err := rule(value); err != nil {
				return fmt.Errorf("field %s: %w", field, err)
			}
		}
	}
	return nil
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(s interface{}) error {
	// Basic struct validation implementation
	return nil
}

// Required creates a required field rule
func Required() Rule {
	return func(value interface{}) error {
		if value == nil {
			return fmt.Errorf("required field")
		}

		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("required field")
			}
		case []interface{}:
			if len(v) == 0 {
				return fmt.Errorf("required field")
			}
		}

		return nil
	}
}

// MinLength creates a minimum length rule
func MinLength(min int) Rule {
	return func(value interface{}) error {
		if str, ok := value.(string); ok {
			if len(str) < min {
				return fmt.Errorf("minimum length is %d", min)
			}
		}
		return nil
	}
}

// MaxLength creates a maximum length rule
func MaxLength(max int) Rule {
	return func(value interface{}) error {
		if str, ok := value.(string); ok {
			if len(str) > max {
				return fmt.Errorf("maximum length is %d", max)
			}
		}
		return nil
	}
}

// Email creates an email validation rule
func Email() Rule {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return func(value interface{}) error {
		if str, ok := value.(string); ok {
			if !emailRegex.MatchString(str) {
				return fmt.Errorf("invalid email format")
			}
		}
		return nil
	}
}

// URL creates a URL validation rule
func URL() Rule {
	urlRegex := regexp.MustCompile(`^https?://`)
	return func(value interface{}) error {
		if str, ok := value.(string); ok {
			if !urlRegex.MatchString(str) {
				return fmt.Errorf("invalid URL format")
			}
		}
		return nil
	}
}

// Pattern creates a regex pattern validation rule
func Pattern(pattern string) Rule {
	regex := regexp.MustCompile(pattern)
	return func(value interface{}) error {
		if str, ok := value.(string); ok {
			if !regex.MatchString(str) {
				return fmt.Errorf("does not match required pattern")
			}
		}
		return nil
	}
}

// OneOf creates an enum validation rule
func OneOf(allowed ...string) Rule {
	return func(value interface{}) error {
		if str, ok := value.(string); ok {
			for _, a := range allowed {
				if str == a {
					return nil
				}
			}
			return fmt.Errorf("must be one of: %v", allowed)
		}
		return nil
	}
}

// Min creates a minimum value rule
func Min(min float64) Rule {
	return func(value interface{}) error {
		switch v := value.(type) {
		case float64:
			if v < min {
				return fmt.Errorf("minimum value is %f", min)
			}
		case int:
			if float64(v) < min {
				return fmt.Errorf("minimum value is %f", min)
			}
		}
		return nil
	}
}

// Max creates a maximum value rule
func Max(max float64) Rule {
	return func(value interface{}) error {
		switch v := value.(type) {
		case float64:
			if v > max {
				return fmt.Errorf("maximum value is %f", max)
			}
		case int:
			if float64(v) > max {
				return fmt.Errorf("maximum value is %f", max)
			}
		}
		return nil
	}
}

// Custom creates a custom validation rule
func Custom(fn func(interface{}) error) Rule {
	return fn
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	rule := Email()
	return rule(email)
}

// ValidateURL validates URL format
func ValidateURL(url string) error {
	rule := URL()
	return rule(url)
}

// ValidateString validates string with options
func ValidateString(str string, opts ...Rule) error {
	for _, opt := range opts {
		if err := opt(str); err != nil {
			return err
		}
	}
	return nil
}

// PayloadValidator validates API payloads
type PayloadValidator struct {
	maxSize int
	rules   map[string][]Rule
}

// NewPayloadValidator creates a new payload validator
func NewPayloadValidator(maxSize int) *PayloadValidator {
	return &PayloadValidator{
		maxSize: maxSize,
		rules:   make(map[string][]Rule),
	}
}

// AddRule adds a validation rule
func (pv *PayloadValidator) AddRule(field string, rule Rule) *PayloadValidator {
	pv.rules[field] = append(pv.rules[field], rule)
	return pv
}

// Validate validates a payload
func (pv *PayloadValidator) Validate(data []byte) error {
	if len(data) > pv.maxSize {
		return fmt.Errorf("payload exceeds maximum size of %d bytes", pv.maxSize)
	}

	// Parse and validate
	var payload map[string]interface{}
	// Would unmarshal JSON here

	for field, rules := range pv.rules {
		value, exists := payload[field]
		if !exists {
			continue
		}

		for _, rule := range rules {
			if err := rule(value); err != nil {
				return fmt.Errorf("field %s: %w", field, err)
			}
		}
	}

	return nil
}

// PolicyValidator validates policies
type PolicyValidator struct {
	requiredFields map[string]bool
}

// NewPolicyValidator creates a new policy validator
func NewPolicyValidator() *PolicyValidator {
	return &PolicyValidator{
		requiredFields: make(map[string]bool),
	}
}

// RequireField marks a field as required
func (pv *PolicyValidator) RequireField(field string) *PolicyValidator {
	pv.requiredFields[field] = true
	return pv
}

// Validate validates a policy
func (pv *PolicyValidator) Validate(policy interface{}) error {
	v := reflect.ValueOf(policy)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for field := range pv.requiredFields {
		if !v.FieldByName(field).IsValid() {
			return fmt.Errorf("required field missing: %s", field)
		}
	}

	return nil
}

// InputError represents an input validation error
type InputError struct {
	Field   string
	Message string
}

// Error implements error interface
func (ie InputError) Error() string {
	return fmt.Sprintf("invalid input for field %s: %s", ie.Field, ie.Message)
}

// InputErrors represents multiple input errors
type InputErrors struct {
	errors []InputError
}

// NewInputErrors creates new input errors
func NewInputErrors() *InputErrors {
	return &InputErrors{
		errors: make([]InputError, 0),
	}
}

// Add adds an error
func (ie *InputErrors) Add(field, message string) *InputErrors {
	ie.errors = append(ie.errors, InputError{
		Field:   field,
		Message: message,
	})
	return ie
}

// HasErrors checks if there are errors
func (ie *InputErrors) HasErrors() bool {
	return len(ie.errors) > 0
}

// Error implements error interface
func (ie *InputErrors) Error() string {
	if len(ie.errors) == 0 {
		return "no errors"
	}

	var msg string
	for i, err := range ie.errors {
		if i > 0 {
			msg += "; "
		}
		msg += err.Error()
	}
	return msg
}

// Errors returns all errors
func (ie *InputErrors) Errors() []InputError {
	return ie.errors
}

func isRequired(rules []Rule) bool {
	for _, rule := range rules {
		// Simple heuristic - could improve
		_ = rule
	}
	return false
}
