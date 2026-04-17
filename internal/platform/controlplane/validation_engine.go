package controlplane
import (
	"context"
	"fmt"
	"regexp"
	"sync"
)

// ValidationEngine provides comprehensive resource validation
type ValidationEngine struct {
	mu           sync.RWMutex
	rules        map[string]*ValidationRule
	validators   map[string]CustomValidator
	transformers map[string]ResourceTransformer
}

// ValidationRule defines validation for a resource type
type ValidationRule struct {
	Kind         string
	Version      string
	Required     []string
	Rules        map[string]*FieldValidation
	CustomRules  []CustomValidationFn
	BeforeCreate []TransformFn
	BeforeUpdate []TransformFn
	AfterCreate  []PostProcessFn
	AfterUpdate  []PostProcessFn
}

// FieldValidation defines field-level validation
type FieldValidation struct {
	Type       string // string, number, integer, boolean, array, object
	Required   bool
	MinLength  int
	MaxLength  int
	Pattern    string // regex
	Enum       []string
	Minimum    *float64
	Maximum    *float64
	MinItems   int
	MaxItems   int
	Unique     bool
	ReadOnly   bool
	WriteOnly  bool
	Immutable  bool
	Format     string                      // email, uri, uuid, date-time, etc.
	Properties map[string]*FieldValidation // for objects
}

// CustomValidator validates a resource
type CustomValidator interface {
	Validate(ctx context.Context, resource *ManagedResource) error
}

// ResourceTransformer transforms a resource
type ResourceTransformer interface {
	Transform(ctx context.Context, resource *ManagedResource) (*ManagedResource, error)
}

// CustomValidationFn is a validation function
type CustomValidationFn func(context.Context, *ManagedResource) error

// TransformFn is a transform function
type TransformFn func(context.Context, *ManagedResource) error

// PostProcessFn is a post-process function
type PostProcessFn func(context.Context, *ManagedResource) error

// ValidationResult is defined in schema.go
// NewValidationEngine creates a new validation engine
func NewValidationEngine() *ValidationEngine {
	return &ValidationEngine{
		rules:        make(map[string]*ValidationRule),
		validators:   make(map[string]CustomValidator),
		transformers: make(map[string]ResourceTransformer),
	}
}

// RegisterRule registers validation rules for a resource type
func (ve *ValidationEngine) RegisterRule(key string, rule *ValidationRule) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.rules[key] = rule
}

// RegisterValidator registers a custom validator
func (ve *ValidationEngine) RegisterValidator(kind string, validator CustomValidator) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.validators[kind] = validator
}

// RegisterTransformer registers a resource transformer
func (ve *ValidationEngine) RegisterTransformer(kind string, transformer ResourceTransformer) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.transformers[kind] = transformer
}

// Validate validates a resource
func (ve *ValidationEngine) Validate(ctx context.Context, resource *ManagedResource, operation string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make(map[string][]string),
		Warnings: make([]string, 0),
	}

	ve.mu.RLock()
	rule, exists := ve.rules[resource.Kind]
	ve.mu.RUnlock()

	if !exists {
		return result
	}

	// Check required fields
	spec := resource.Spec
	for _, field := range rule.Required {
		if _, exists := spec[field]; !exists {
			result.Valid = false
			result.Errors[field] = append(result.Errors[field], "field is required")
		}
	}

	// Validate field types and values
	for field, value := range spec {
		if fieldRule, exists := rule.Rules[field]; exists {
			ve.validateField(field, value, fieldRule, result)
		}
	}

	// Check immutable fields on update
	if operation == "update" {
		for field, fieldRule := range rule.Rules {
			if fieldRule.Immutable {
				if _, exists := spec[field]; exists {
					result.Warnings = append(result.Warnings, fmt.Sprintf("field %s is immutable", field))
				}
			}
		}
	}

	// Run custom validations
	for _, customRule := range rule.CustomRules {
		if err := customRule(ctx, resource); err != nil {
			result.Valid = false
			result.Errors["_custom"] = append(result.Errors["_custom"], err.Error())
		}
	}

	// Run custom validator
	ve.mu.RLock()
	validator, exists := ve.validators[resource.Kind]
	ve.mu.RUnlock()

	if exists {
		if err := validator.Validate(ctx, resource); err != nil {
			result.Valid = false
			result.Errors["_validator"] = append(result.Errors["_validator"], err.Error())
		}
	}

	return result
}

// validateField validates a field value
func (ve *ValidationEngine) validateField(field string, value interface{}, rule *FieldValidation, result *ValidationResult) {
	switch rule.Type {
	case "string":
		if str, ok := value.(string); ok {
			if rule.MinLength > 0 && len(str) < rule.MinLength {
				result.Valid = false
				result.Errors[field] = append(result.Errors[field], fmt.Sprintf("minimum length is %d", rule.MinLength))
			}
			if rule.MaxLength > 0 && len(str) > rule.MaxLength {
				result.Valid = false
				result.Errors[field] = append(result.Errors[field], fmt.Sprintf("maximum length is %d", rule.MaxLength))
			}
			if rule.Pattern != "" {
				if re, err := regexp.Compile(rule.Pattern); err == nil {
					if !re.MatchString(str) {
						result.Valid = false
						result.Errors[field] = append(result.Errors[field], fmt.Sprintf("does not match pattern %s", rule.Pattern))
					}
				}
			}
			if len(rule.Enum) > 0 {
				found := false
				for _, e := range rule.Enum {
					if e == str {
						found = true
						break
					}
				}
				if !found {
					result.Valid = false
					result.Errors[field] = append(result.Errors[field], fmt.Sprintf("must be one of %v", rule.Enum))
				}
			}
		} else if rule.Required {
			result.Valid = false
			result.Errors[field] = append(result.Errors[field], "must be a string")
		}

	case "number":
		if num, ok := toFloat64(value); ok {
			if rule.Minimum != nil && num < *rule.Minimum {
				result.Valid = false
				result.Errors[field] = append(result.Errors[field], fmt.Sprintf("minimum value is %v", *rule.Minimum))
			}
			if rule.Maximum != nil && num > *rule.Maximum {
				result.Valid = false
				result.Errors[field] = append(result.Errors[field], fmt.Sprintf("maximum value is %v", *rule.Maximum))
			}
		} else if rule.Required {
			result.Valid = false
			result.Errors[field] = append(result.Errors[field], "must be a number")
		}

	case "array":
		if arr, ok := value.([]interface{}); ok {
			if rule.MinItems > 0 && len(arr) < rule.MinItems {
				result.Valid = false
				result.Errors[field] = append(result.Errors[field], fmt.Sprintf("minimum items is %d", rule.MinItems))
			}
			if rule.MaxItems > 0 && len(arr) > rule.MaxItems {
				result.Valid = false
				result.Errors[field] = append(result.Errors[field], fmt.Sprintf("maximum items is %d", rule.MaxItems))
			}
		} else if rule.Required {
			result.Valid = false
			result.Errors[field] = append(result.Errors[field], "must be an array")
		}
	}
}

// Transform transforms a resource
func (ve *ValidationEngine) Transform(ctx context.Context, resource *ManagedResource, operation string) (*ManagedResource, error) {
	ve.mu.RLock()
	rule, exists := ve.rules[resource.Kind]
	ve.mu.RUnlock()

	if !exists {
		return resource, nil
	}

	// Run before transforms
	var transforms []TransformFn
	if operation == "create" {
		transforms = rule.BeforeCreate
	} else if operation == "update" {
		transforms = rule.BeforeUpdate
	}

	for _, transform := range transforms {
		if err := transform(ctx, resource); err != nil {
			return nil, err
		}
	}

	// Run transformer
	ve.mu.RLock()
	transformer, exists := ve.transformers[resource.Kind]
	ve.mu.RUnlock()

	if exists {
		var err error
		resource, err = transformer.Transform(ctx, resource)
		if err != nil {
			return nil, err
		}
	}

	// Run after transforms (post-process)
	var postProcess []PostProcessFn
	if operation == "create" {
		postProcess = rule.AfterCreate
	} else if operation == "update" {
		postProcess = rule.AfterUpdate
	}

	for _, process := range postProcess {
		if err := process(ctx, resource); err != nil {
			return nil, err
		}
	}

	return resource, nil
}

// DefaultTransformer provides default transformation
type DefaultTransformer struct {
	// Custom transform logic
}

func (dt *DefaultTransformer) Transform(ctx context.Context, resource *ManagedResource) (*ManagedResource, error) {
	// Add custom logic here
	return resource, nil
}

// MutationEngine provides resource mutation control
type MutationEngine struct {
	mu       sync.RWMutex
	mutators map[string][]ResourceMutator
}

// ResourceMutator mutates resources
type ResourceMutator interface {
	Mutate(ctx context.Context, resource *ManagedResource, operation string) error
}

// NewMutationEngine creates a new mutation engine
func NewMutationEngine() *MutationEngine {
	return &MutationEngine{
		mutators: make(map[string][]ResourceMutator),
	}
}

// RegisterMutator registers a mutator for a resource kind
func (me *MutationEngine) RegisterMutator(kind string, mutator ResourceMutator) {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.mutators[kind] = append(me.mutators[kind], mutator)
}

// Mutate applies mutations
func (me *MutationEngine) Mutate(ctx context.Context, resource *ManagedResource, operation string) error {
	me.mu.RLock()
	mutators, exists := me.mutators[resource.Kind]
	me.mu.RUnlock()

	if !exists {
		return nil
	}

	for _, mutator := range mutators {
		if err := mutator.Mutate(ctx, resource, operation); err != nil {
			return err
		}
	}

	return nil
}

// DefaultMutators provides common mutations
type DefaultMutators struct{}

// AddMissingLabels adds missing labels to resources
type AddMissingLabels struct {
	DefaultLabels map[string]string
}

func (aml *AddMissingLabels) Mutate(ctx context.Context, resource *ManagedResource, operation string) error {
	if operation != "create" {
		return nil
	}

	if resource.Metadata.Labels == nil {
		resource.Metadata.Labels = make(map[string]string)
	}

	for k, v := range aml.DefaultLabels {
		if _, exists := resource.Metadata.Labels[k]; !exists {
			resource.Metadata.Labels[k] = v
		}
	}

	return nil
}

// EnforceNamespaceQuota enforces namespace quotas
type EnforceNamespaceQuota struct {
	QuotaLimits map[string]int
	Current     map[string]int
}

func (enq *EnforceNamespaceQuota) Mutate(ctx context.Context, resource *ManagedResource, operation string) error {
	if operation != "create" {
		return nil
	}

	kind := resource.Kind
	if limit, exists := enq.QuotaLimits[kind]; exists {
		if current, _ := enq.Current[kind]; current >= limit {
			return fmt.Errorf("quota exceeded for %s", kind)
		}
		enq.Current[kind]++
	}

	return nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
