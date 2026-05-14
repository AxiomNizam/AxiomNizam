package etl

// =====================================================
// P1.1 — ETL Pipeline as a declarative resource
//
// Historically `Pipeline` was an imperative config object managed by
// `Engine`.  For the declarative control-plane we wrap it in a proper
// `resources.Resource` (Spec/Status/Conditions/ObservedGeneration) so a
// dedicated controller can reconcile it like any other platform resource.
//
// The in-memory `Pipeline` type is preserved untouched (it is still the
// wire/exec shape used by the existing Engine).  `PipelineResource` is the
// declarative envelope and is what the API server / controller operate on.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

// PipelineSpec is the *desired* state of an ETL Pipeline.
//
// It is intentionally a superset that mirrors the imperative `Pipeline`
// definition so the controller can deterministically produce / update an
// underlying `*Pipeline` inside `Engine` from the spec alone.
type PipelineSpec struct {
	Description   string                 `json:"description,omitempty"`
	Steps         []Step                 `json:"steps"`
	Schedule      string                 `json:"schedule,omitempty"`
	Orchestration OrchestrationConfig    `json:"orchestration,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	Tags          []string               `json:"tags,omitempty"`

	// Paused, when true, asks the controller to not execute scheduled runs.
	// It maps to `PipelineStatus = paused` on the underlying Pipeline.
	Paused bool `json:"paused,omitempty"`
}

// PipelineResourceStatus extends the canonical object status with
// ETL-specific telemetry.  All fields are controller-owned.
type PipelineResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	// PipelineStatus mirrors the engine's per-pipeline status
	// (`created/running/paused/succeeded/failed/stopped`).
	PipelineStatus PipelineStatus `json:"pipelineStatus,omitempty"`

	// LastRunAt is the most recent run timestamp.
	LastRunAt *time.Time `json:"lastRunAt,omitempty"`

	// RunCount is the total number of runs observed by the controller.
	RunCount int `json:"runCount"`

	// LastRunID is the identifier of the most recent `PipelineRun`.
	LastRunID string `json:"lastRunId,omitempty"`
}

// PipelineResource is the declarative resource for an ETL Pipeline.
type PipelineResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   PipelineSpec           `json:"spec"`
	Status PipelineResourceStatus `json:"status"`
}

// Kind / APIVersion constants used by the controller + API server.
const (
	PipelineKind       = "ETLPipeline"
	PipelineAPIVersion = "etl.axiomnizam.io/v1"
)

// --- resources.Resource implementation ---

func (p *PipelineResource) GetObjectMeta() *resources.ObjectMeta { return &p.ObjectMeta }
func (p *PipelineResource) GetTypeMeta() *resources.TypeMeta     { return &p.TypeMeta }

func (p *PipelineResource) GetStatus() *resources.ObjectStatus {
	return &p.Status.ObjectStatus
}

func (p *PipelineResource) SetStatus(status *resources.ObjectStatus) {
	if status == nil {
		return
	}
	p.Status.ObjectStatus = *status
}

func (p *PipelineResource) DeepCopy() resources.Resource {
	cp := *p
	// Steps slice is the only deep field we realistically need copied.
	if len(p.Spec.Steps) > 0 {
		cp.Spec.Steps = make([]Step, len(p.Spec.Steps))
		copy(cp.Spec.Steps, p.Spec.Steps)
	}
	return &cp
}

// --- reconciler.Resource implementation ---

// GetKey returns the canonical namespace/name key.  If namespace is empty
// we fall back to just the name to remain consistent with global
// resources.
func (p *PipelineResource) GetKey() string {
	if p.Namespace == "" {
		return p.Name
	}
	return p.Namespace + "/" + p.Name
}

func (p *PipelineResource) GetGeneration() int64         { return p.Generation }
func (p *PipelineResource) GetObservedGeneration() int64 { return p.Status.ObservedGeneration }

// ToPipeline converts the declarative resource into the imperative
// `*Pipeline` shape consumed by `Engine`.  The returned Pipeline reuses
// the resource UID as its ID so the two stay linked.
func (p *PipelineResource) ToPipeline() *Pipeline {
	now := time.Now()
	id := p.UID
	if id == "" {
		id = p.Name
	}
	status := PipelineCreated
	if p.Spec.Paused {
		status = PipelinePaused
	}
	if p.Status.PipelineStatus != "" {
		status = p.Status.PipelineStatus
	}
	return &Pipeline{
		ID:            id,
		Name:          p.Name,
		Description:   p.Spec.Description,
		Steps:         append([]Step(nil), p.Spec.Steps...),
		Schedule:      p.Spec.Schedule,
		Orchestration: p.Spec.Orchestration,
		Config:        p.Spec.Config,
		Status:        status,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     now,
		LastRunAt:     p.Status.LastRunAt,
		RunCount:      p.Status.RunCount,
		Tags:          p.Spec.Tags,
	}
}
