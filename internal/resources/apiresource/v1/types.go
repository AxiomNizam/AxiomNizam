package v1

import (
	"time"
)

// APIVersion and Kind define the resource type
const (
	APIVersion = "axiom.io/v1"
	Kind       = "APIResource"
)

// Phase represents the lifecycle phase of a resource
type Phase string

const (
	PhasePending    Phase = "Pending"
	PhaseValidated  Phase = "Validated"
	PhaseApplied    Phase = "Applied"
	PhaseFailed     Phase = "Failed"
	PhaseTerminated Phase = "Terminated"
)

// ConditionStatus represents the status of a condition
type ConditionStatus string

const (
	ConditionStatusTrue    ConditionStatus = "True"
	ConditionStatusFalse   ConditionStatus = "False"
	ConditionStatusUnknown ConditionStatus = "Unknown"
)

// Standard condition types
const (
	ConditionTypeReady       = "Ready"
	ConditionTypeValidated   = "Validated"
	ConditionTypeSynced      = "Synced"
	ConditionTypeFailed      = "Failed"
	ConditionTypeReconciling = "Reconciling"
)

// Condition describes an aspect of the resource's state
type Condition struct {
	// Type of condition
	Type string `json:"type"`

	// Status of the condition
	Status ConditionStatus `json:"status"`

	// Reason is a short machine-readable explanation
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable explanation
	Message string `json:"message,omitempty"`

	// FirstObservedTime when the condition was first observed
	FirstObservedTime time.Time `json:"firstObservedTime"`

	// LastTransitionTime when the condition last changed
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

// ObjectMetadata contains standard object metadata
type ObjectMetadata struct {
	// Name of the resource
	Name string `json:"name"`

	// Namespace where the resource exists
	Namespace string `json:"namespace"`

	// UID is a unique identifier
	UID string `json:"uid,omitempty"`

	// Generation is incremented on each spec change
	Generation int64 `json:"generation,omitempty"`

	// CreatedAt is when the resource was created
	CreatedAt time.Time `json:"createdAt,omitempty"`

	// UpdatedAt is when the resource was last updated
	UpdatedAt time.Time `json:"updatedAt,omitempty"`

	// Labels are key-value pairs for resource identification
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations store arbitrary metadata
	Annotations map[string]string `json:"annotations,omitempty"`

	// FinalizedAt is when the resource was finalized (added in v1)
	FinalizedAt *time.Time `json:"finalizedAt,omitempty"`

	// DeletionTimestamp is set when the resource is being deleted (added in v1)
	DeletionTimestamp *time.Time `json:"deletionTimestamp,omitempty"`
}

// APIResourceSpec defines the desired state of an APIResource
type APIResourceSpec struct {
	// BasePath is the API base path
	BasePath string `json:"basePath"`

	// Title is the API title
	Title string `json:"title"`

	// Description is a human-readable description
	Description string `json:"description,omitempty"`

	// Version is the API version
	Version string `json:"version"`

	// Tags for resource identification
	Tags map[string]string `json:"tags,omitempty"`

	// Timeout in seconds
	Timeout int `json:"timeout,omitempty"`

	// Validation rules
	ValidationRules []ValidationRule `json:"validationRules,omitempty"`

	// Policies to apply (added in v1)
	Policies []PolicyReference `json:"policies,omitempty"`

	// Custom data
	Data map[string]interface{} `json:"data,omitempty"`
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Pattern   string   `json:"pattern,omitempty"`
	MinLength int      `json:"minLength,omitempty"`
	MaxLength int      `json:"maxLength,omitempty"`
	Values    []string `json:"values,omitempty"`
	Message   string   `json:"message,omitempty"`
	Required  bool     `json:"required,omitempty"`
}

// PolicyReference references a policy to be applied (added in v1)
type PolicyReference struct {
	Name       string                 `json:"name"`
	Version    string                 `json:"version,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// APIResourceStatus represents the observed state of an APIResource
type APIResourceStatus struct {
	// Phase is the current lifecycle phase
	Phase Phase `json:"phase"`

	// Ready indicates if the resource is operational
	Ready bool `json:"ready"`

	// Message is a human-readable status message
	Message string `json:"message,omitempty"`

	// LastUpdateTime is when the status was last updated
	LastUpdateTime time.Time `json:"lastUpdateTime,omitempty"`

	// Conditions are detailed status conditions
	Conditions []Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation this status was observed for
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ReconcileCount is the number of reconciliation attempts
	ReconcileCount int64 `json:"reconcileCount,omitempty"`

	// LastReconcileTime is when reconciliation last ran
	LastReconcileTime time.Time `json:"lastReconcileTime,omitempty"`
}

// APIResource is the v1 stable APIResource type
type APIResource struct {
	// APIVersion is the version of the API
	APIVersion string `json:"apiVersion"`

	// Kind is the resource kind
	Kind string `json:"kind"`

	// Metadata is the resource metadata
	Metadata ObjectMetadata `json:"metadata"`

	// Spec is the desired state
	Spec APIResourceSpec `json:"spec"`

	// Status is the observed state
	Status APIResourceStatus `json:"status,omitempty"`
}

// GetKey returns namespace/name for work queue
func (a *APIResource) GetKey() string {
	return a.Metadata.Namespace + "/" + a.Metadata.Name
}

// GetGeneration returns the spec generation
func (a *APIResource) GetGeneration() int64 {
	return a.Metadata.Generation
}

// GetObservedGeneration returns the observed spec generation
func (a *APIResource) GetObservedGeneration() int64 {
	return a.Status.ObservedGeneration
}

// GetPhase returns the current phase
func (a *APIResource) GetPhase() Phase {
	return a.Status.Phase
}

// SetPhase sets the phase and updates timestamp
func (a *APIResource) SetPhase(phase Phase) {
	a.Status.Phase = phase
	a.Status.LastUpdateTime = time.Now()
}

// IsReady returns true if resource is operational
func (a *APIResource) IsReady() bool {
	return a.Status.Ready && a.Status.Phase == PhaseApplied
}

// SetReady marks the resource as ready
func (a *APIResource) SetReady(ready bool) {
	a.Status.Ready = ready
	if ready {
		a.Status.Phase = PhaseApplied
	}
	a.Status.LastUpdateTime = time.Now()
}

// SetMessage sets the status message
func (a *APIResource) SetMessage(msg string) {
	a.Status.Message = msg
	a.Status.LastUpdateTime = time.Now()
}

// IncrementReconcileCount increments the reconciliation count
func (a *APIResource) IncrementReconcileCount() {
	a.Status.ReconcileCount++
	a.Status.LastReconcileTime = time.Now()
}

// MarkReady marks the resource as Ready with reason
func (a *APIResource) MarkReady(reason, message string) {
	a.SetReady(true)
	a.SetPhase(PhaseApplied)
	a.AddCondition(ConditionTypeReady, ConditionStatusTrue, reason, message)
}

// MarkNotReady marks the resource as not Ready with reason
func (a *APIResource) MarkNotReady(reason, message string) {
	a.SetReady(false)
	a.AddCondition(ConditionTypeReady, ConditionStatusFalse, reason, message)
}

// MarkValidated marks the resource as validated
func (a *APIResource) MarkValidated(reason, message string) {
	a.SetPhase(PhaseValidated)
	a.AddCondition(ConditionTypeValidated, ConditionStatusTrue, reason, message)
}

// MarkFailed marks the resource as failed
func (a *APIResource) MarkFailed(reason, message string) {
	a.SetPhase(PhaseFailed)
	a.SetReady(false)
	a.SetMessage(message)
	a.AddCondition(ConditionTypeFailed, ConditionStatusTrue, reason, message)
}

// MarkTerminated marks the resource as terminated
func (a *APIResource) MarkTerminated(reason, message string) {
	a.SetPhase(PhaseTerminated)
	a.SetReady(false)
	a.AddCondition("Terminated", ConditionStatusTrue, reason, message)
}

// MarkReconciling marks the resource as currently reconciling
func (a *APIResource) MarkReconciling(reason, message string) {
	a.AddCondition(ConditionTypeReconciling, ConditionStatusTrue, reason, message)
}

// MarkSynced marks the resource as synced with desired state
func (a *APIResource) MarkSynced(reason, message string) {
	a.AddCondition(ConditionTypeSynced, ConditionStatusTrue, reason, message)
}

// AddCondition adds or updates a condition
func (a *APIResource) AddCondition(condType string, status ConditionStatus, reason, message string) {
	now := time.Now()

	// Check if condition already exists
	for i, cond := range a.Status.Conditions {
		if cond.Type == condType {
			// Update existing condition
			if a.Status.Conditions[i].Status != status {
				a.Status.Conditions[i].LastTransitionTime = now
			}
			a.Status.Conditions[i].Status = status
			a.Status.Conditions[i].Reason = reason
			a.Status.Conditions[i].Message = message
			return
		}
	}

	// Add new condition
	a.Status.Conditions = append(a.Status.Conditions, Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		FirstObservedTime:  now,
		LastTransitionTime: now,
	})
}

// GetCondition returns a specific condition by type
func (a *APIResource) GetCondition(condType string) *Condition {
	for i := range a.Status.Conditions {
		if a.Status.Conditions[i].Type == condType {
			return &a.Status.Conditions[i]
		}
	}
	return nil
}

// HasCondition returns true if a condition exists and has the given status
func (a *APIResource) HasCondition(condType string, status ConditionStatus) bool {
	cond := a.GetCondition(condType)
	return cond != nil && cond.Status == status
}

// IsDeleting returns true if the resource is being deleted
func (a *APIResource) IsDeleting() bool {
	return a.Metadata.DeletionTimestamp != nil
}

// New creates a new APIResource with initial state
func New(namespace, name string, spec APIResourceSpec) *APIResource {
	now := time.Now()

	return &APIResource{
		APIVersion: APIVersion,
		Kind:       Kind,
		Metadata: ObjectMetadata{
			Name:        name,
			Namespace:   namespace,
			UID:         now.Format("20060102150405-") + name,
			Generation:  1,
			CreatedAt:   now,
			UpdatedAt:   now,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Spec: spec,
		Status: APIResourceStatus{
			Phase:              PhasePending,
			Ready:              false,
			Message:            "Pending validation and application",
			LastUpdateTime:     now,
			Conditions:         []Condition{},
			ObservedGeneration: 0,
			ReconcileCount:     0,
			LastReconcileTime:  time.Time{},
		},
	}
}
