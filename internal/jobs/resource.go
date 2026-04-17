package jobs

// =====================================================
// P1.3 — Job as a declarative resource
//
// `Job` historically was an imperative runtime record owned by the
// `JobManager`.  `JobResource` exposes the same concept as a declarative
// resource (Spec/Status/Conditions/ObservedGeneration) so a controller
// can reconcile it on the shared workqueue with rate-limited retries.
//
// The imperative `Job` type is left untouched and is still the wire /
// storage shape.  `JobResource` is what the API server presents.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

const (
	JobKind       = "Job"
	JobAPIVersion = "jobs.axiomnizam.io/v1"
)

// JobSpec is the *desired* state of a background job.
type JobSpec struct {
	Type        JobType                `json:"type"`
	Priority    JobPriority            `json:"priority"`
	Data        map[string]interface{} `json:"data"`
	MaxRetries  int                    `json:"maxRetries"`
	Timeout     time.Duration          `json:"timeout"`
	Tags        []string               `json:"tags,omitempty"`
	CallbackURL string                 `json:"callbackUrl,omitempty"`
	DeadlineAt  time.Time              `json:"deadlineAt,omitempty"`

	// Schedule, if non-empty, asks the controller to register this job
	// with a cron-style scheduler instead of submitting it once.
	Schedule string `json:"schedule,omitempty"`

	// Suspend, when true, prevents the controller from dispatching new
	// runs of a scheduled job.
	Suspend bool `json:"suspend,omitempty"`
}

// JobResourceStatus extends the canonical ObjectStatus with job-specific
// lifecycle telemetry.
type JobResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	JobStatus   JobStatus              `json:"jobStatus,omitempty"`
	Retries     int                    `json:"retries"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	StartedAt   time.Time              `json:"startedAt,omitempty"`
	CompletedAt time.Time              `json:"completedAt,omitempty"`
}

// JobResource is the declarative resource for a background job.
type JobResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   JobSpec           `json:"spec"`
	Status JobResourceStatus `json:"status"`
}

// --- resources.Resource ---

func (j *JobResource) GetObjectMeta() *resources.ObjectMeta { return &j.ObjectMeta }
func (j *JobResource) GetTypeMeta() *resources.TypeMeta     { return &j.TypeMeta }
func (j *JobResource) GetStatus() *resources.ObjectStatus   { return &j.Status.ObjectStatus }
func (j *JobResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		j.Status.ObjectStatus = *s
	}
}
func (j *JobResource) DeepCopy() resources.Resource {
	cp := *j
	return &cp
}

// --- reconciler.Resource ---

func (j *JobResource) GetKey() string {
	if j.Namespace == "" {
		return j.Name
	}
	return j.Namespace + "/" + j.Name
}
func (j *JobResource) GetGeneration() int64         { return j.Generation }
func (j *JobResource) GetObservedGeneration() int64 { return j.Status.ObservedGeneration }

// ToJob converts the declarative resource into the imperative `*Job`
// shape accepted by `JobManager.Submit`.
func (j *JobResource) ToJob() *Job {
	id := j.UID
	if id == "" {
		id = j.Name
	}
	return &Job{
		ID:          id,
		Type:        j.Spec.Type,
		Status:      JobStatusPending,
		Priority:    j.Spec.Priority,
		Data:        j.Spec.Data,
		MaxRetries:  j.Spec.MaxRetries,
		CreatedAt:   j.CreatedAt,
		Timeout:     j.Spec.Timeout,
		Tags:        j.Spec.Tags,
		CallbackURL: j.Spec.CallbackURL,
		DeadlineAt:  j.Spec.DeadlineAt,
	}
}
