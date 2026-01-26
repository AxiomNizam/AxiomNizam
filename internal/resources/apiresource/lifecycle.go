package apiresource

import (
	"time"
)

// APIResource is the actual resource instance with lifecycle management
type APIResource struct {
	Metadata MetadataSpec  `json:"metadata" yaml:"metadata"`
	Spec     SpecSection   `json:"spec" yaml:"spec"`
	Status   StatusSection `json:"status" yaml:"status"`
}

// MetadataSpec contains resource metadata
type MetadataSpec struct {
	Name       string            `json:"name" yaml:"name"`
	Namespace  string            `json:"namespace" yaml:"namespace"`
	UID        string            `json:"uid"`
	Generation int64             `json:"generation"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// SpecSection is the desired state
type SpecSection struct {
	BasePath    string            `json:"basePath" yaml:"basePath"`
	Title       string            `json:"title" yaml:"title"`
	Description string            `json:"description" yaml:"description"`
	Version     string            `json:"version" yaml:"version"`
	Tags        map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Timeout     int               `json:"timeout" yaml:"timeout"` // seconds
}

// StatusSection is the observed state (controlled by controller)
type StatusSection struct {
	Phase      string      `json:"phase" yaml:"phase"`     // Pending, Creating, Ready, Failed
	Ready      bool        `json:"ready" yaml:"ready"`     // true if operational
	Message    string      `json:"message" yaml:"message"` // human-readable status
	LastUpdate time.Time   `json:"lastUpdate" yaml:"lastUpdate"`
	Conditions []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

// GetKey returns namespace/name for work queue
func (api *APIResource) GetKey() string {
	return api.Metadata.Namespace + "/" + api.Metadata.Name
}

// SetPhase updates the phase and timestamp
func (api *APIResource) SetPhase(phase string) {
	api.Status.Phase = phase
	api.Status.LastUpdate = time.Now()
}

// SetReady marks resource as ready or not
func (api *APIResource) SetReady(ready bool) {
	api.Status.Ready = ready
	if ready {
		api.Status.Phase = "Ready"
	}
}

// SetMessage updates status message
func (api *APIResource) SetMessage(message string) {
	api.Status.Message = message
	api.Status.LastUpdate = time.Now()
}

// AddCondition adds a condition to track status
func (api *APIResource) AddCondition(condType, status, message string) {
	api.Status.Conditions = append(api.Status.Conditions, Condition{
		Type:      condType,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// MarkFailed marks resource as failed with reason
func (api *APIResource) MarkFailed(reason string) {
	api.Status.Phase = "Failed"
	api.Status.Ready = false
	api.Status.Message = reason
	api.Status.LastUpdate = time.Now()
	api.AddCondition("Failed", "True", reason)
}

// MarkCreating marks resource as being created
func (api *APIResource) MarkCreating(message string) {
	api.Status.Phase = "Creating"
	api.Status.Ready = false
	api.Status.Message = message
	api.Status.LastUpdate = time.Now()
	api.AddCondition("Creating", "True", message)
}

// New creates a new APIResource with initial state
func New(namespace, name string, spec map[string]interface{}) *APIResource {
	now := time.Now()

	// Convert map to SpecSection
	specSection := SpecSection{
		BasePath:    getStringValue(spec, "basePath", ""),
		Title:       getStringValue(spec, "title", ""),
		Description: getStringValue(spec, "description", ""),
		Version:     getStringValue(spec, "version", "1.0"),
		Tags:        make(map[string]string),
		Timeout:     getIntValue(spec, "timeout", 30),
	}

	return &APIResource{
		Metadata: MetadataSpec{
			Name:       name,
			Namespace:  namespace,
			UID:        now.Format("20060102150405"),
			Generation: 1,
			CreatedAt:  now,
			UpdatedAt:  now,
			Labels:     make(map[string]string),
		},
		Spec: specSection,
		Status: StatusSection{
			Phase:      "Pending",
			Ready:      false,
			Message:    "Resource created, waiting for reconciliation",
			LastUpdate: now,
			Conditions: []Condition{},
		},
	}
}

// Helper functions for spec conversion
func getStringValue(spec map[string]interface{}, key string, defaultVal string) string {
	if val, ok := spec[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func getIntValue(spec map[string]interface{}, key string, defaultVal int) int {
	if val, ok := spec[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
		if num, ok := val.(int); ok {
			return num
		}
	}
	return defaultVal
}
