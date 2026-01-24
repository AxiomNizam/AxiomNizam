package resources

import (
	"fmt"
	"time"
)

// WorkloadResource represents a workload that needs to be executed
type WorkloadResource struct {
	BaseResource `json:"baseResource"`

	// Spec defines the desired state
	Spec WorkloadSpec `json:"spec"`
}

// WorkloadSpec defines the desired state of a workload
type WorkloadSpec struct {
	// Parallelism controls max concurrent executions
	Parallelism int32 `json:"parallelism,omitempty"`

	// Completions desired number of successful completions
	Completions int32 `json:"completions,omitempty"`

	// ActiveDeadlineSeconds max seconds to keep running
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// Template defines execution template
	Template WorkloadTemplate `json:"template"`

	// Suspend pauses execution
	Suspend bool `json:"suspend,omitempty"`

	// RetryStrategy for failed executions
	RetryStrategy *RetryStrategy `json:"retryStrategy,omitempty"`
}

// WorkloadTemplate defines how to execute the workload
type WorkloadTemplate struct {
	// Image to execute
	Image string `json:"image"`

	// Command to run
	Command []string `json:"command,omitempty"`

	// Args to command
	Args []string `json:"args,omitempty"`

	// Environment variables
	Env map[string]string `json:"env,omitempty"`

	// Timeout in seconds
	Timeout int64 `json:"timeout,omitempty"`
}

// RetryStrategy defines retry behavior
type RetryStrategy struct {
	// MaxRetries maximum attempt count
	MaxRetries int32 `json:"maxRetries,omitempty"`

	// BackoffLimit exponential backoff limit
	BackoffLimit int32 `json:"backoffLimit,omitempty"`
}

// DeepCopy creates a deep copy
func (w *WorkloadResource) DeepCopy() Resource {
	return &WorkloadResource{
		BaseResource: w.BaseResource,
		Spec:         w.Spec,
	}
}

// PipelineResource represents a pipeline of workloads
type PipelineResource struct {
	BaseResource `json:"baseResource"`

	// Spec defines the desired state
	Spec PipelineSpec `json:"spec"`
}

// PipelineSpec defines the desired state of a pipeline
type PipelineSpec struct {
	// Stages are executed sequentially
	Stages []PipelineStage `json:"stages"`

	// Parallelism for stage execution
	Parallelism int32 `json:"parallelism,omitempty"`
}

// PipelineStage represents a stage in a pipeline
type PipelineStage struct {
	// Name of the stage
	Name string `json:"name"`

	// Tasks to execute in parallel
	Tasks []PipelineTask `json:"tasks"`

	// DependsOn other stages
	DependsOn []string `json:"dependsOn,omitempty"`
}

// PipelineTask represents a task in a stage
type PipelineTask struct {
	// Name of the task
	Name string `json:"name"`

	// WorkloadRef references a workload to execute
	WorkloadRef string `json:"workloadRef"`
}

// DeepCopy creates a deep copy
func (p *PipelineResource) DeepCopy() Resource {
	return &PipelineResource{
		BaseResource: p.BaseResource,
		Spec:         p.Spec,
	}
}

// ScheduleResource represents a scheduled execution
type ScheduleResource struct {
	BaseResource `json:"baseResource"`

	// Spec defines the desired state
	Spec ScheduleSpec `json:"spec"`
}

// ScheduleSpec defines when and what to execute
type ScheduleSpec struct {
	// Cron expression (e.g., "0 9 * * 1-5" = 9 AM weekdays)
	Cron string `json:"cron"`

	// Timezone for cron evaluation
	Timezone string `json:"timezone,omitempty"`

	// WorkloadRef references workload to execute
	WorkloadRef string `json:"workloadRef"`

	// Suspend pauses scheduling
	Suspend bool `json:"suspend,omitempty"`

	// SuccessfulExecutionsHistoryLimit how many success records to keep
	SuccessfulExecutionsHistoryLimit *int32 `json:"successfulExecutionsHistoryLimit,omitempty"`

	// FailedExecutionsHistoryLimit how many failure records to keep
	FailedExecutionsHistoryLimit *int32 `json:"failedExecutionsHistoryLimit,omitempty"`
}

// DeepCopy creates a deep copy
func (s *ScheduleResource) DeepCopy() Resource {
	return &ScheduleResource{
		BaseResource: s.BaseResource,
		Spec:         s.Spec,
	}
}

// ExecutionResource represents an execution result
type ExecutionResource struct {
	BaseResource `json:"baseResource"`

	// Spec what was executed
	Spec ExecutionSpec `json:"spec"`
}

// ExecutionSpec represents what was executed
type ExecutionSpec struct {
	// WorkloadRef the workload that was executed
	WorkloadRef string `json:"workloadRef"`

	// StartTime when execution started
	StartTime time.Time `json:"startTime"`

	// CompletionTime when execution completed
	CompletionTime *time.Time `json:"completionTime,omitempty"`

	// Duration how long execution took
	Duration *time.Duration `json:"duration,omitempty"`

	// ExitCode of the execution
	ExitCode *int32 `json:"exitCode,omitempty"`

	// Stdout output
	Stdout string `json:"stdout,omitempty"`

	// Stderr output
	Stderr string `json:"stderr,omitempty"`
}

// DeepCopy creates a deep copy
func (e *ExecutionResource) DeepCopy() Resource {
	return &ExecutionResource{
		BaseResource: e.BaseResource,
		Spec:         e.Spec,
	}
}

// NewWorkloadResource creates a new workload resource
func NewWorkloadResource(name, namespace string) *WorkloadResource {
	now := time.Now()
	return &WorkloadResource{
		BaseResource: BaseResource{
			TypeMeta: TypeMeta{
				APIVersion: "axiom.dev/v1",
				Kind:       "Workload",
			},
			ObjectMeta: ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				UID:         fmt.Sprintf("%d-%s", now.Unix(), name),
				CreatedAt:   now,
				UpdatedAt:   now,
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
			Status: ObjectStatus{
				Phase:              "Pending",
				LastTransitionTime: now,
			},
		},
	}
}

// NewPipelineResource creates a new pipeline resource
func NewPipelineResource(name, namespace string) *PipelineResource {
	now := time.Now()
	return &PipelineResource{
		BaseResource: BaseResource{
			TypeMeta: TypeMeta{
				APIVersion: "axiom.dev/v1",
				Kind:       "Pipeline",
			},
			ObjectMeta: ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				UID:         fmt.Sprintf("%d-%s", now.Unix(), name),
				CreatedAt:   now,
				UpdatedAt:   now,
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
			Status: ObjectStatus{
				Phase:              "Pending",
				LastTransitionTime: now,
			},
		},
	}
}

// NewScheduleResource creates a new schedule resource
func NewScheduleResource(name, namespace string) *ScheduleResource {
	now := time.Now()
	return &ScheduleResource{
		BaseResource: BaseResource{
			TypeMeta: TypeMeta{
				APIVersion: "axiom.dev/v1",
				Kind:       "Schedule",
			},
			ObjectMeta: ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				UID:         fmt.Sprintf("%d-%s", now.Unix(), name),
				CreatedAt:   now,
				UpdatedAt:   now,
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
			Status: ObjectStatus{
				Phase:              "Active",
				LastTransitionTime: now,
			},
		},
	}
}

// NewExecutionResource creates a new execution resource
func NewExecutionResource(name, namespace, workloadRef string) *ExecutionResource {
	now := time.Now()
	return &ExecutionResource{
		BaseResource: BaseResource{
			TypeMeta: TypeMeta{
				APIVersion: "axiom.dev/v1",
				Kind:       "Execution",
			},
			ObjectMeta: ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				UID:         fmt.Sprintf("%d-%s", now.Unix(), name),
				CreatedAt:   now,
				UpdatedAt:   now,
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
			Status: ObjectStatus{
				Phase:              "Running",
				LastTransitionTime: now,
			},
		},
		Spec: ExecutionSpec{
			WorkloadRef: workloadRef,
			StartTime:   now,
		},
	}
}
