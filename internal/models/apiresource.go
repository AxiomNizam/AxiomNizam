package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// APIResource represents a Kubernetes-style API resource definition
type APIResource struct {
	APIVersion string      `json:"apiVersion" yaml:"apiVersion"`
	Kind       string      `json:"kind" yaml:"kind"`
	Metadata   Metadata    `json:"metadata" yaml:"metadata"`
	Spec       interface{} `json:"spec" yaml:"spec"`
	Status     *Status     `json:"status,omitempty" yaml:"status,omitempty"`
}

// Metadata contains resource metadata
type Metadata struct {
	Name                       string            `json:"name" yaml:"name"`
	Namespace                  string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	UID                        string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	ResourceVersion            int64             `json:"resourceVersion,omitempty" yaml:"resourceVersion,omitempty"`
	Generation                 int64             `json:"generation,omitempty" yaml:"generation,omitempty"`
	CreatedAt                  time.Time         `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	UpdatedAt                  time.Time         `json:"updatedAt,omitempty" yaml:"updatedAt,omitempty"`
	DeletedAt                  *time.Time        `json:"deletedAt,omitempty" yaml:"deletedAt,omitempty"`
	DeletionGracePeriodSeconds *int32            `json:"deletionGracePeriodSeconds,omitempty" yaml:"deletionGracePeriodSeconds,omitempty"`
	Labels                     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations                map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	OwnerReferences            []OwnerReference  `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Finalizers                 []string          `json:"finalizers,omitempty" yaml:"finalizers,omitempty"`
}

// Status contains resource status
type Status struct {
	Phase              string                 `json:"phase" yaml:"phase"`
	Conditions         []Condition            `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Message            string                 `json:"message,omitempty" yaml:"message,omitempty"`
	ObservedGeneration int64                  `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	LastUpdateTime     time.Time              `json:"lastUpdateTime,omitempty" yaml:"lastUpdateTime,omitempty"`
	Details            map[string]interface{} `json:"details,omitempty" yaml:"details,omitempty"`
}

// Condition represents resource condition
type Condition struct {
	Type               string    `json:"type" yaml:"type"`
	Status             string    `json:"status" yaml:"status"` // True, False, Unknown
	LastUpdateTime     time.Time `json:"lastUpdateTime,omitempty" yaml:"lastUpdateTime,omitempty"`
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
	Reason             string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message            string    `json:"message,omitempty" yaml:"message,omitempty"`
}

// OwnerReference indicates controller ownership
type OwnerReference struct {
	APIVersion         string `json:"apiVersion" yaml:"apiVersion"`
	Kind               string `json:"kind" yaml:"kind"`
	Name               string `json:"name" yaml:"name"`
	UID                string `json:"uid" yaml:"uid"`
	Controller         *bool  `json:"controller,omitempty" yaml:"controller,omitempty"`
	BlockOwnerDeletion *bool  `json:"blockOwnerDeletion,omitempty" yaml:"blockOwnerDeletion,omitempty"`
}

// CustomResourceDefinition defines a custom resource schema
type CustomResourceDefinition struct {
	APIVersion string     `json:"apiVersion" yaml:"apiVersion"`
	Kind       string     `json:"kind" yaml:"kind"`
	Metadata   Metadata   `json:"metadata" yaml:"metadata"`
	Spec       CRDSpec    `json:"spec" yaml:"spec"`
	Status     *CRDStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// CRDSpec defines CRD specification
type CRDSpec struct {
	Group                 string              `json:"group" yaml:"group"`
	Names                 CRDNames            `json:"names" yaml:"names"`
	Scope                 string              `json:"scope" yaml:"scope"` // Namespaced or Cluster
	Versions              []CRDVersion        `json:"versions" yaml:"versions"`
	Conversion            *ConversionStrategy `json:"conversion,omitempty" yaml:"conversion,omitempty"`
	PreserveUnknownFields *bool               `json:"preserveUnknownFields,omitempty" yaml:"preserveUnknownFields,omitempty"`
	Categories            []string            `json:"categories,omitempty" yaml:"categories,omitempty"`
}

// CRDNames defines names for a custom resource
type CRDNames struct {
	Plural     string   `json:"plural" yaml:"plural"`
	Singular   string   `json:"singular,omitempty" yaml:"singular,omitempty"`
	Kind       string   `json:"kind" yaml:"kind"`
	ListKind   string   `json:"listKind,omitempty" yaml:"listKind,omitempty"`
	ShortNames []string `json:"shortNames,omitempty" yaml:"shortNames,omitempty"`
	Categories []string `json:"categories,omitempty" yaml:"categories,omitempty"`
}

// CRDVersion defines a version of a custom resource
type CRDVersion struct {
	Name                     string            `json:"name" yaml:"name"`
	Served                   bool              `json:"served" yaml:"served"`
	Storage                  bool              `json:"storage" yaml:"storage"`
	Deprecated               *bool             `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	DeprecationWarning       *string           `json:"deprecationWarning,omitempty" yaml:"deprecationWarning,omitempty"`
	Schema                   *ValidationSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
	Subresources             *Subresources     `json:"subresources,omitempty" yaml:"subresources,omitempty"`
	AdditionalPrinterColumns []PrinterColumn   `json:"additionalPrinterColumns,omitempty" yaml:"additionalPrinterColumns,omitempty"`
}

// ValidationSchema defines JSON schema for validation
type ValidationSchema struct {
	OpenAPIV3Schema *JSONSchema `json:"openAPIV3Schema,omitempty" yaml:"openAPIV3Schema,omitempty"`
}

// JSONSchema represents JSON Schema properties
type JSONSchema struct {
	Type                        string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Title                       string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Description                 string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Default                     interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	Example                     interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
	Format                      string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Pattern                     string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Minimum                     *float64               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum                     *float64               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MinLength                   *int                   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength                   *int                   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinItems                    *int                   `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems                    *int                   `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	Required                    []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                        []interface{}          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Items                       *JSONSchema            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties                  map[string]*JSONSchema `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties        *JSONSchema            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Ref                         string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	XPreserveUnknownFields      *bool                  `json:"x-kubernetes-preserve-unknown-fields,omitempty" yaml:"x-kubernetes-preserve-unknown-fields,omitempty"`
	XKubernetesEmbeddedResource *bool                  `json:"x-kubernetes-embedded-resource,omitempty" yaml:"x-kubernetes-embedded-resource,omitempty"`
	XKubernetesValidation       []ValidationRule       `json:"x-kubernetes-validations,omitempty" yaml:"x-kubernetes-validations,omitempty"`
}

// ValidationRule defines custom validation rules
type ValidationRule struct {
	Rule    string `json:"rule" yaml:"rule"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// Subresources defines subresources for a custom resource
type Subresources struct {
	Status *StatusSubresource `json:"status,omitempty" yaml:"status,omitempty"`
	Scale  *ScaleSubresource  `json:"scale,omitempty" yaml:"scale,omitempty"`
}

// StatusSubresource defines status subresource
type StatusSubresource struct{}

// ScaleSubresource defines scale subresource
type ScaleSubresource struct {
	SpecReplicasPath   string `json:"specReplicasPath" yaml:"specReplicasPath"`
	StatusReplicasPath string `json:"statusReplicasPath" yaml:"statusReplicasPath"`
	LabelSelectorPath  string `json:"labelSelectorPath,omitempty" yaml:"labelSelectorPath,omitempty"`
}

// PrinterColumn defines additional columns for kubectl output
type PrinterColumn struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"` // string, integer, number, boolean, date
	Format      string `json:"format,omitempty" yaml:"format,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	JSONPath    string `json:"jsonPath" yaml:"jsonPath"`
	Priority    int    `json:"priority,omitempty" yaml:"priority,omitempty"`
}

// ConversionStrategy defines conversion between API versions
type ConversionStrategy struct {
	Strategy string             `json:"strategy" yaml:"strategy"` // None, Webhook
	Webhook  *WebhookConversion `json:"webhook,omitempty" yaml:"webhook,omitempty"`
}

// WebhookConversion defines webhook-based conversion
type WebhookConversion struct {
	URL          string        `json:"url" yaml:"url"`
	CABundle     []byte        `json:"caBundle,omitempty" yaml:"caBundle,omitempty"`
	ClientConfig *ClientConfig `json:"clientConfig,omitempty" yaml:"clientConfig,omitempty"`
}

// ClientConfig defines webhook client configuration
type ClientConfig struct {
	URL      string            `json:"url,omitempty" yaml:"url,omitempty"`
	Service  *ServiceReference `json:"service,omitempty" yaml:"service,omitempty"`
	CABundle []byte            `json:"caBundle,omitempty" yaml:"caBundle,omitempty"`
}

// ServiceReference references a service
type ServiceReference struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	Port      *int32 `json:"port,omitempty" yaml:"port,omitempty"`
}

// CRDStatus tracks CRD status
type CRDStatus struct {
	Conditions     []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	AcceptedNames  CRDNames    `json:"acceptedNames,omitempty" yaml:"acceptedNames,omitempty"`
	StoredVersions []string    `json:"storedVersions,omitempty" yaml:"storedVersions,omitempty"`
}

// APIResourceValidator validates API resources against schema
type APIResourceValidator struct {
	schema *ValidationSchema
}

// NewAPIResourceValidator creates a new validator
func NewAPIResourceValidator(schema *ValidationSchema) *APIResourceValidator {
	return &APIResourceValidator{
		schema: schema,
	}
}

// Validate validates a resource against the schema
func (arv *APIResourceValidator) Validate(resource *APIResource) error {
	if arv.schema == nil || arv.schema.OpenAPIV3Schema == nil {
		return nil
	}

	return arv.validateValue(resource.Spec, arv.schema.OpenAPIV3Schema, "spec")
}

// validateValue validates a value against a schema
func (arv *APIResourceValidator) validateValue(value interface{}, schema *JSONSchema, path string) error {
	if schema == nil {
		return nil
	}

	// Type validation
	if schema.Type != "" {
		if !arv.typeMatches(value, schema.Type) {
			return fmt.Errorf("invalid type at %s: expected %s, got %T", path, schema.Type, value)
		}
	}

	// String validations
	if str, ok := value.(string); ok {
		if schema.MinLength != nil && len(str) < *schema.MinLength {
			return fmt.Errorf("string too short at %s: minimum length is %d", path, *schema.MinLength)
		}
		if schema.MaxLength != nil && len(str) > *schema.MaxLength {
			return fmt.Errorf("string too long at %s: maximum length is %d", path, *schema.MaxLength)
		}
		if schema.Pattern != "" {
			regex, err := regexp.Compile(schema.Pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern at %s: %w", path, err)
			}
			if !regex.MatchString(str) {
				return fmt.Errorf("value at %s does not match pattern: %s", path, schema.Pattern)
			}
		}
		if len(schema.Enum) > 0 {
			found := false
			for _, e := range schema.Enum {
				if e == str {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("value at %s not in enum", path)
			}
		}
	}

	// Numeric validations
	if num, ok := value.(float64); ok {
		if schema.Minimum != nil && num < *schema.Minimum {
			return fmt.Errorf("number too small at %s: minimum is %f", path, *schema.Minimum)
		}
		if schema.Maximum != nil && num > *schema.Maximum {
			return fmt.Errorf("number too large at %s: maximum is %f", path, *schema.Maximum)
		}
	}

	// Object validations
	if obj, ok := value.(map[string]interface{}); ok && schema.Properties != nil {
		for key, propSchema := range schema.Properties {
			if val, exists := obj[key]; exists {
				if err := arv.validateValue(val, propSchema, fmt.Sprintf("%s.%s", path, key)); err != nil {
					return err
				}
			} else if propSchema.Default != nil {
				obj[key] = propSchema.Default
			}
		}

		// Check required fields
		for _, required := range schema.Required {
			if _, exists := obj[required]; !exists {
				return fmt.Errorf("required field missing: %s.%s", path, required)
			}
		}
	}

	// Array validations
	if arr, ok := value.([]interface{}); ok {
		if schema.MinItems != nil && len(arr) < *schema.MinItems {
			return fmt.Errorf("array too small at %s: minimum items is %d", path, *schema.MinItems)
		}
		if schema.MaxItems != nil && len(arr) > *schema.MaxItems {
			return fmt.Errorf("array too large at %s: maximum items is %d", path, *schema.MaxItems)
		}
	}

	return nil
}

// typeMatches checks if value matches schema type
func (arv *APIResourceValidator) typeMatches(value interface{}, schemaType string) bool {
	switch schemaType {
	case "string":
		_, ok := value.(string)
		return ok
	case "integer", "number":
		_, ok := value.(float64)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "null":
		return value == nil
	}
	return true
}

// APIResourceFactory creates resources with defaults
type APIResourceFactory struct {
	kind       string
	apiVersion string
	schema     *ValidationSchema
}

// NewAPIResourceFactory creates a factory
func NewAPIResourceFactory(kind, apiVersion string, schema *ValidationSchema) *APIResourceFactory {
	return &APIResourceFactory{
		kind:       kind,
		apiVersion: apiVersion,
		schema:     schema,
	}
}

// Create creates a new resource with defaults
func (arf *APIResourceFactory) Create(name, namespace string, spec interface{}) *APIResource {
	resource := &APIResource{
		APIVersion: arf.apiVersion,
		Kind:       arf.kind,
		Metadata: Metadata{
			Name:      name,
			Namespace: namespace,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Spec: spec,
		Status: &Status{
			Phase:          "Creating",
			LastUpdateTime: time.Now(),
			Conditions:     make([]Condition, 0),
		},
	}

	return resource
}

// AsJSON converts resource to JSON
func (ar *APIResource) AsJSON() ([]byte, error) {
	return json.MarshalIndent(ar, "", "  ")
}

// FromJSON unmarshals JSON into resource
func (ar *APIResource) FromJSON(data []byte) error {
	return json.Unmarshal(data, ar)
}
