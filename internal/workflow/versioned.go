package workflow

import (
	"fmt"
	"sync"
	"time"
)

// WorkflowDefinition defines a workflow with versioning
type WorkflowDefinition struct {
	ID              string
	Name            string
	Version         string
	Status          string // draft, published, deprecated
	Steps           []*WorkflowStep
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       string
	Description     string
	Tags            []string
}

// WorkflowStep represents a step in workflow
type WorkflowStep struct {
	ID              string
	Name            string
	StepNumber      int
	StepType        string // action, decision, approval, notification
	Configuration   map[string]interface{}
	NextStepID      string
	ErrorHandler    string
	Timeout         int64 // milliseconds
}

// WorkflowInstance represents an execution instance
type WorkflowInstance struct {
	ID              string
	WorkflowID      string
	WorkflowVersion string
	Status          string // running, completed, failed, paused
	StartedAt       time.Time
	CompletedAt     *time.Time
	CurrentStepID   string
	ExecutionHistory []*ExecutionRecord
	ContextData     map[string]interface{}
	Variables       map[string]interface{}
}

// ExecutionRecord records step execution
type ExecutionRecord struct {
	StepID        string
	StepName      string
	ExecutedAt    time.Time
	Duration      int64 // milliseconds
	Status        string // success, failed, skipped
	Output        map[string]interface{}
	Error         string
}

// WorkflowVersion represents versioned workflow
type WorkflowVersion struct {
	WorkflowID      string
	Version         string
	VersionNumber   int
	PreviousVersion string
	ChangeSummary   string
	BreakingChanges []string
	CreatedAt       time.Time
	CreatedBy       string
	IsActive        bool
}

// MultiVersionWorkflowManager manages versioned workflows
type MultiVersionWorkflowManager struct {
	mu                sync.RWMutex
	workflows         map[string]*WorkflowDefinition
	versions          map[string][]*WorkflowVersion
	instances         map[string]*WorkflowInstance
	versionHistory    map[string][]*WorkflowDefinition
	executionLogs     []*ExecutionRecord
	migrations        map[string]*MigrationStrategy
	maxInstanceSize   int
	maxExecutionSize  int
}

// MigrationStrategy defines workflow migration
type MigrationStrategy struct {
	FromVersion string
	ToVersion   string
	Steps       []*MigrationStep
	Automatic   bool
	RequiresApproval bool
}

// MigrationStep defines a migration action
type MigrationStep struct {
	Step        int
	Description string
	Action      string // transform_data, update_reference, notify
	Mapping     map[string]interface{}
}

// NewMultiVersionWorkflowManager creates workflow manager
func NewMultiVersionWorkflowManager() *MultiVersionWorkflowManager {
	return &MultiVersionWorkflowManager{
		workflows:       make(map[string]*WorkflowDefinition),
		versions:        make(map[string][]*WorkflowVersion),
		instances:       make(map[string]*WorkflowInstance),
		versionHistory:  make(map[string][]*WorkflowDefinition),
		executionLogs:   make([]*ExecutionRecord, 0),
		migrations:      make(map[string]*MigrationStrategy),
		maxInstanceSize: 100000,
		maxExecutionSize: 50000,
	}
}

// CreateWorkflow creates a new workflow
func (mwm *MultiVersionWorkflowManager) CreateWorkflow(workflow *WorkflowDefinition) (*WorkflowVersion, error) {
	mwm.mu.Lock()
	defer mwm.mu.Unlock()

	if workflow.ID == "" {
		workflow.ID = fmt.Sprintf("wf-%d", time.Now().UnixNano())
	}

	if workflow.Version == "" {
		workflow.Version = "1.0.0"
	}

	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	mwm.workflows[workflow.ID] = workflow

	// Create version record
	version := &WorkflowVersion{
		WorkflowID:    workflow.ID,
		Version:       workflow.Version,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
		CreatedBy:     workflow.CreatedBy,
		IsActive:      true,
	}

	mwm.versions[workflow.ID] = []*WorkflowVersion{version}
	mwm.versionHistory[workflow.ID] = []*WorkflowDefinition{workflow}

	return version, nil
}

// PublishWorkflowVersion publishes a new version
func (mwm *MultiVersionWorkflowManager) PublishWorkflowVersion(workflowID string, newDef *WorkflowDefinition) (*WorkflowVersion, error) {
	mwm.mu.Lock()
	defer mwm.mu.Unlock()

	oldDef, exists := mwm.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found")
	}

	// Determine new version
	oldVersions := mwm.versions[workflowID]
	versionNumber := len(oldVersions) + 1
	newVersion := fmt.Sprintf("%d.0.0", versionNumber)

	newDef.ID = workflowID
	newDef.Version = newVersion
	newDef.UpdatedAt = time.Now()

	// Create version record
	version := &WorkflowVersion{
		WorkflowID:      workflowID,
		Version:         newVersion,
		VersionNumber:   versionNumber,
		PreviousVersion: oldDef.Version,
		CreatedAt:       time.Now(),
		CreatedBy:       newDef.CreatedBy,
		IsActive:        true,
	}

	// Mark old version inactive
	for _, v := range oldVersions {
		v.IsActive = false
	}

	mwm.workflows[workflowID] = newDef
	mwm.versions[workflowID] = append(mwm.versions[workflowID], version)
	mwm.versionHistory[workflowID] = append(mwm.versionHistory[workflowID], newDef)

	return version, nil
}

// StartWorkflowInstance starts a workflow execution
func (mwm *MultiVersionWorkflowManager) StartWorkflowInstance(workflowID, version string, contextData map[string]interface{}) (*WorkflowInstance, error) {
	mwm.mu.Lock()
	defer mwm.mu.Unlock()

	workflow, exists := mwm.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found")
	}

	// If no version specified, use current
	if version == "" {
		version = workflow.Version
	}

	instance := &WorkflowInstance{
		ID:              fmt.Sprintf("inst-%d", time.Now().UnixNano()),
		WorkflowID:      workflowID,
		WorkflowVersion: version,
		Status:          "running",
		StartedAt:       time.Now(),
		ExecutionHistory: make([]*ExecutionRecord, 0),
		ContextData:     contextData,
		Variables:       make(map[string]interface{}),
	}

	// Set first step
	if len(workflow.Steps) > 0 {
		instance.CurrentStepID = workflow.Steps[0].ID
	}

	mwm.instances[instance.ID] = instance

	if len(mwm.instances) > mwm.maxInstanceSize {
		// Remove oldest instance
		var oldestID string
		var oldestTime time.Time
		for id, inst := range mwm.instances {
			if oldestTime.IsZero() || inst.StartedAt.Before(oldestTime) {
				oldestID = id
				oldestTime = inst.StartedAt
			}
		}
		if oldestID != "" {
			delete(mwm.instances, oldestID)
		}
	}

	return instance, nil
}

// RecordStepExecution records step execution
func (mwm *MultiVersionWorkflowManager) RecordStepExecution(instanceID string, execution *ExecutionRecord) error {
	mwm.mu.Lock()
	defer mwm.mu.Unlock()

	instance, exists := mwm.instances[instanceID]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	instance.ExecutionHistory = append(instance.ExecutionHistory, execution)
	mwm.executionLogs = append(mwm.executionLogs, execution)

	if len(mwm.executionLogs) > mwm.maxExecutionSize {
		mwm.executionLogs = mwm.executionLogs[1:]
	}

	return nil
}

// CompleteWorkflowInstance marks workflow as completed
func (mwm *MultiVersionWorkflowManager) CompleteWorkflowInstance(instanceID string) error {
	mwm.mu.Lock()
	defer mwm.mu.Unlock()

	instance, exists := mwm.instances[instanceID]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	now := time.Now()
	instance.CompletedAt = &now
	instance.Status = "completed"

	return nil
}

// GetWorkflowVersions gets all versions of a workflow
func (mwm *MultiVersionWorkflowManager) GetWorkflowVersions(workflowID string) []*WorkflowVersion {
	mwm.mu.RLock()
	defer mwm.mu.RUnlock()

	if versions, exists := mwm.versions[workflowID]; exists {
		return versions
	}
	return make([]*WorkflowVersion, 0)
}

// GetWorkflowVersion gets specific workflow version
func (mwm *MultiVersionWorkflowManager) GetWorkflowVersion(workflowID, version string) (*WorkflowDefinition, error) {
	mwm.mu.RLock()
	defer mwm.mu.RUnlock()

	history, exists := mwm.versionHistory[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found")
	}

	for _, wf := range history {
		if wf.Version == version {
			return wf, nil
		}
	}

	return nil, fmt.Errorf("version not found")
}

// CreateMigrationStrategy creates a migration plan
func (mwm *MultiVersionWorkflowManager) CreateMigrationStrategy(fromVersion, toVersion string) (*MigrationStrategy, error) {
	mwm.mu.Lock()
	defer mwm.mu.Unlock()

	strategy := &MigrationStrategy{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Steps:       make([]*MigrationStep, 0),
		Automatic:   true,
	}

	key := fmt.Sprintf("%s-%s", fromVersion, toVersion)
	mwm.migrations[key] = strategy

	return strategy, nil
}

// GetInstanceHistory gets workflow instance history
func (mwm *MultiVersionWorkflowManager) GetInstanceHistory(workflowID string, limit int) []*WorkflowInstance {
	mwm.mu.RLock()
	defer mwm.mu.RUnlock()

	instances := make([]*WorkflowInstance, 0)

	for _, inst := range mwm.instances {
		if inst.WorkflowID == workflowID {
			instances = append(instances, inst)
		}
	}

	if limit > 0 && len(instances) > limit {
		instances = instances[len(instances)-limit:]
	}

	return instances
}

// GetWorkflowMetrics gets workflow execution metrics
func (mwm *MultiVersionWorkflowManager) GetWorkflowMetrics(workflowID string) map[string]interface{} {
	mwm.mu.RLock()
	defer mwm.mu.RUnlock()

	instances := mwm.GetInstanceHistory(workflowID, 0)
	completed := 0
	failed := 0
	running := 0
	totalDuration := int64(0)

	for _, inst := range instances {
		switch inst.Status {
		case "completed":
			completed++
			if inst.CompletedAt != nil {
				totalDuration += inst.CompletedAt.Sub(inst.StartedAt).Milliseconds()
			}
		case "failed":
			failed++
		case "running":
			running++
		}
	}

	avgDuration := float64(0)
	if completed > 0 {
		avgDuration = float64(totalDuration) / float64(completed)
	}

	successRate := 0.0
	if completed+failed > 0 {
		successRate = float64(completed) / float64(completed+failed) * 100
	}

	return map[string]interface{}{
		"total_instances":   len(instances),
		"completed":         completed,
		"failed":            failed,
		"running":           running,
		"success_rate":      successRate,
		"avg_duration_ms":   avgDuration,
		"total_versions":    len(mwm.versions[workflowID]),
	}
}

// GetWorkflowStatus gets overall workflow status
func (mwm *MultiVersionWorkflowManager) GetWorkflowStatus() map[string]interface{} {
	mwm.mu.RLock()
	defer mwm.mu.RUnlock()

	totalInstances := len(mwm.instances)
	activeVersions := 0
	totalVersions := 0

	for _, versions := range mwm.versions {
		for _, v := range versions {
			totalVersions++
			if v.IsActive {
				activeVersions++
			}
		}
	}

	return map[string]interface{}{
		"total_workflows":   len(mwm.workflows),
		"total_versions":    totalVersions,
		"active_versions":   activeVersions,
		"running_instances": totalInstances,
		"execution_logs":    len(mwm.executionLogs),
		"migrations":        len(mwm.migrations),
	}
}
