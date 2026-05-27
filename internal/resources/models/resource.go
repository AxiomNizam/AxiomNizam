package models

import (
	"fmt"
	"time"
)

// ObjectMeta contains metadata common to all resources
type ObjectMeta struct {
	// Name of the object
	Name string `json:"name"`

	// Namespace the object belongs to
	Namespace string `json:"namespace,omitempty"`

	// UID unique identifier
	UID string `json:"uid"`

	// Generation is incremented when spec changes
	Generation int64 `json:"generation"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt timestamp
	UpdatedAt time.Time `json:"updatedAt"`

	// DeletedAt timestamp (set when being deleted)
	DeletedAt *time.Time `json:"deletedAt,omitempty"`

	// Labels key-value pairs for organization
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for arbitrary metadata
	Annotations map[string]string `json:"annotations,omitempty"`

	// OwnerReferences for resource hierarchy
	OwnerReferences []OwnerReference `json:"ownerReferences,omitempty"`

	// Finalizers prevent deletion until cleared
	Finalizers []string `json:"finalizers,omitempty"`
}

// OwnerReference represents a reference to an owner object
type OwnerReference struct {
	// APIVersion of the owner
	APIVersion string `json:"apiVersion"`

	// Kind of the owner
	Kind string `json:"kind"`

	// Name of the owner
	Name string `json:"name"`

	// UID of the owner
	UID string `json:"uid"`

	// Controller indicates if this owner is a controller
	Controller *bool `json:"controller,omitempty"`
}

// TypeMeta describes a resource type
type TypeMeta struct {
	// APIVersion is the version of the API
	APIVersion string `json:"apiVersion"`

	// Kind is the name of the object schema
	Kind string `json:"kind"`
}

// ObjectStatus contains status information
type ObjectStatus struct {
	// Phase indicates the state (Pending, Active, Failed, etc)
	Phase string `json:"phase"`

	// Conditions represent the latest state
	Conditions []Condition `json:"conditions,omitempty"`

	// LastTransitionTime when status last changed
	LastTransitionTime time.Time `json:"lastTransitionTime"`

	// ObservedGeneration reflects the generation the controller has seen
	ObservedGeneration int64 `json:"observedGeneration"`
}

// Condition represents a condition of a resource
type Condition struct {
	// Type of the condition (Ready, Error, etc)
	Type string `json:"type"`

	// Status of the condition (True, False, Unknown)
	Status string `json:"status"`

	// LastTransitionTime when status changed
	LastTransitionTime time.Time `json:"lastTransitionTime"`

	// Reason for status change
	Reason string `json:"reason"`

	// Message describing the state
	Message string `json:"message"`
}

// Resource is the interface all resources must implement
type Resource interface {
	// GetObjectMeta returns the object metadata
	GetObjectMeta() *ObjectMeta

	// GetTypeMeta returns the type metadata
	GetTypeMeta() *TypeMeta

	// GetStatus returns the status
	GetStatus() *ObjectStatus

	// SetStatus sets the status
	SetStatus(status *ObjectStatus)

	// DeepCopy creates a deep copy
	DeepCopy() Resource
}

// BaseResource provides common implementation for all resources
type BaseResource struct {
	TypeMeta   `json:"typeMetadata"`
	ObjectMeta `json:"metadata"`
	Status     ObjectStatus `json:"status"`
}

// GetObjectMeta returns object metadata
func (br *BaseResource) GetObjectMeta() *ObjectMeta {
	return &br.ObjectMeta
}

// GetTypeMeta returns type metadata
func (br *BaseResource) GetTypeMeta() *TypeMeta {
	return &br.TypeMeta
}

// GetStatus returns status
func (br *BaseResource) GetStatus() *ObjectStatus {
	return &br.Status
}

// SetStatus sets status
func (br *BaseResource) SetStatus(status *ObjectStatus) {
	br.Status = *status
}

// HasFinalizer checks if a finalizer exists
func (om *ObjectMeta) HasFinalizer(finalizer string) bool {
	for _, f := range om.Finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}

// AddFinalizer adds a finalizer
func (om *ObjectMeta) AddFinalizer(finalizer string) {
	if !om.HasFinalizer(finalizer) {
		om.Finalizers = append(om.Finalizers, finalizer)
	}
}

// RemoveFinalizer removes a finalizer
func (om *ObjectMeta) RemoveFinalizer(finalizer string) {
	finalizers := []string{}
	for _, f := range om.Finalizers {
		if f != finalizer {
			finalizers = append(finalizers, f)
		}
	}
	om.Finalizers = finalizers
}

// MatchesLabels checks if all query labels match
func (om *ObjectMeta) MatchesLabels(selector map[string]string) bool {
	for k, v := range selector {
		if om.Labels[k] != v {
			return false
		}
	}
	return true
}

// GetKey returns the canonical key for the resource in the form "namespace/name".
// When Namespace is empty the bare name is returned.
func (br *BaseResource) GetKey() string {
	if br.Namespace == "" {
		return br.Name
	}
	return fmt.Sprintf("%s/%s", br.Namespace, br.Name)
}

// GetGeneration returns the spec generation of the resource.
func (br *BaseResource) GetGeneration() int64 {
	return br.ObjectMeta.Generation
}

// GetObservedGeneration returns the generation last observed by the controller.
func (br *BaseResource) GetObservedGeneration() int64 {
	return br.Status.ObservedGeneration
}
