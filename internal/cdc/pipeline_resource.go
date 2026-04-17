package cdc

// =====================================================
// P1.2 — CDC Pipeline as a declarative resource
//
// The existing `CDCPipeline` (owned by `PipelineEngine`) and `CDCStream`
// (owned by `ChangeDataCapture`) are imperative runtime objects.  Here
// we expose `CDCPipelineResource` as a declarative envelope with Spec /
// Status / Conditions so a controller can reconcile it.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

const (
	CDCPipelineKind       = "CDCPipeline"
	CDCPipelineAPIVersion = "cdc.axiomnizam.io/v1"
)

// CDCPipelineSpec is the *desired* state of a CDC Pipeline.
type CDCPipelineSpec struct {
	Description string                 `json:"description,omitempty"`
	Source      CDCSource              `json:"source"`
	Sink        CDCSink                `json:"sink"`
	Filters     CDCFilters             `json:"filters"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Tags        []string               `json:"tags,omitempty"`

	// Paused asks the controller to keep the underlying CDCPipeline in
	// the `paused` state.
	Paused bool `json:"paused,omitempty"`
}

// CDCPipelineResourceStatus extends canonical status with CDC telemetry.
type CDCPipelineResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	CDCStatus   PipelineStatus `json:"cdcStatus,omitempty"`
	EventCount  int64          `json:"eventCount"`
	ErrorCount  int64          `json:"errorCount"`
	LastEventAt *time.Time     `json:"lastEventAt,omitempty"`
	Lag         string         `json:"lag,omitempty"`
}

// CDCPipelineResource is the declarative resource for a CDC pipeline.
type CDCPipelineResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   CDCPipelineSpec           `json:"spec"`
	Status CDCPipelineResourceStatus `json:"status"`
}

// --- resources.Resource ---

func (c *CDCPipelineResource) GetObjectMeta() *resources.ObjectMeta { return &c.ObjectMeta }
func (c *CDCPipelineResource) GetTypeMeta() *resources.TypeMeta     { return &c.TypeMeta }
func (c *CDCPipelineResource) GetStatus() *resources.ObjectStatus   { return &c.Status.ObjectStatus }
func (c *CDCPipelineResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		c.Status.ObjectStatus = *s
	}
}
func (c *CDCPipelineResource) DeepCopy() resources.Resource {
	cp := *c
	return &cp
}

// --- reconciler.Resource ---

func (c *CDCPipelineResource) GetKey() string {
	if c.Namespace == "" {
		return c.Name
	}
	return c.Namespace + "/" + c.Name
}
func (c *CDCPipelineResource) GetGeneration() int64         { return c.Generation }
func (c *CDCPipelineResource) GetObservedGeneration() int64 { return c.Status.ObservedGeneration }

// ToCDCPipeline projects the declarative resource onto the imperative
// `*CDCPipeline` shape consumed by `PipelineEngine`.
func (c *CDCPipelineResource) ToCDCPipeline() *CDCPipeline {
	id := c.UID
	if id == "" {
		id = c.Name
	}
	status := CDCCreated
	if c.Spec.Paused {
		status = CDCPaused
	}
	if c.Status.CDCStatus != "" {
		status = c.Status.CDCStatus
	}
	return &CDCPipeline{
		ID:          id,
		Name:        c.Name,
		Description: c.Spec.Description,
		Source:      c.Spec.Source,
		Sink:        c.Spec.Sink,
		Filters:     c.Spec.Filters,
		Status:      status,
		Config:      c.Spec.Config,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   time.Now(),
		EventCount:  c.Status.EventCount,
		ErrorCount:  c.Status.ErrorCount,
		LastEventAt: c.Status.LastEventAt,
		Lag:         c.Status.Lag,
		Tags:        c.Spec.Tags,
	}
}
