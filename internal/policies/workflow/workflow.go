package workflow

import (
	"fmt"
	"sync"
	"time"
)

// WorkflowPolicy defines workflow and orchestration policies
type WorkflowPolicy struct {
	ID            string
	Name          string
	Type          string
	Version       string
	Enabled       bool
	Workflows     []WorkflowDefinition
	ApprovalRules []ApprovalRule
	SLAPolicy     SLAPolicy
	ErrorHandling ErrorHandlingPolicy
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// WorkflowDefinition defines a workflow
type WorkflowDefinition struct {
	ID         string
	Name       string
	Steps      []WorkflowStep
	Triggers   []WorkflowTrigger
	Conditions []WorkflowCondition
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
	Enabled    bool
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID         string
	Name       string
	Type       string // "action", "decision", "approval", "notification", "wait"
	Handler    string
	Parameters map[string]interface{}
	NextSteps  []string // IDs of next steps
	OnError    string   // ID of error handling step
	Timeout    time.Duration
}

// WorkflowTrigger defines what triggers a workflow
type WorkflowTrigger struct {
	Type      string // "manual", "scheduled", "event", "webhook"
	Event     string
	Schedule  string // cron expression for scheduled
	Condition string
}

// WorkflowCondition defines branching conditions
type WorkflowCondition struct {
	Expression string
	TrueStep   string
	FalseStep  string
}

// ApprovalRule defines approval requirements
type ApprovalRule struct {
	ID                string
	AppliesTo         []string // workflow IDs
	RequiredApprovals int
	Approvers         []string
	TimeLimit         time.Duration
	NotifyBefore      time.Duration
}

// SLAPolicy defines service level agreements for workflows
type SLAPolicy struct {
	TargetCompletionTime time.Duration
	WarningThreshold     float64 // percentage
	AlertOnViolation     bool
	Remediation          string
}

// ErrorHandlingPolicy defines error handling behavior
type ErrorHandlingPolicy struct {
	RetryPolicy     RetryPolicy
	FailoverEnabled bool
	NotifyOnError   bool
	LogLevel        string // "debug", "info", "warn", "error"
	CircuitBreaker  CircuitBreakerConfig
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	RetryableErrors   []string
}

// CircuitBreakerConfig defines circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	HalfOpenRequests int
}

// GetID returns policy ID
func (wp *WorkflowPolicy) GetID() string {
	return wp.ID
}

// GetName returns policy name
func (wp *WorkflowPolicy) GetName() string {
	return wp.Name
}

// GetType returns policy type
func (wp *WorkflowPolicy) GetType() string {
	return wp.Type
}

// GetVersion returns version
func (wp *WorkflowPolicy) GetVersion() string {
	return wp.Version
}

// GetEnabled returns if enabled
func (wp *WorkflowPolicy) GetEnabled() bool {
	return wp.Enabled
}

// Validate validates the policy
func (wp *WorkflowPolicy) Validate() error {
	if wp.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if wp.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if len(wp.Workflows) == 0 {
		return fmt.Errorf("at least one workflow must be defined")
	}
	return nil
}

// WorkflowEngine orchestrates workflow execution
type WorkflowEngine struct {
	mu           sync.RWMutex
	policies     []*WorkflowPolicy
	executions   map[string]WorkflowExecution
	approvals    map[string]ApprovalRequest
	slaMonitor   *SLAMonitor
	errorHandler *ErrorHandler
}

// WorkflowExecution tracks workflow execution
type WorkflowExecution struct {
	ID             string
	WorkflowID     string
	Status         string // "pending", "running", "completed", "failed", "paused"
	StartTime      time.Time
	EndTime        time.Time
	CurrentStep    string
	CompletedSteps []string
	FailedSteps    []string
	Context        map[string]interface{}
	Approvals      map[string]bool
	Errors         []WorkflowError
}

// WorkflowError represents a workflow error
type WorkflowError struct {
	StepID    string
	Error     string
	Timestamp time.Time
	Retried   int
}

// ApprovalRequest represents an approval request
type ApprovalRequest struct {
	ID           string
	ExecutionID  string
	ApprovalRule ApprovalRule
	Requestor    string
	CreatedAt    time.Time
	DueDate      time.Time
	Status       string // "pending", "approved", "rejected"
	Approvers    map[string]bool
	Comments     []ApprovalComment
}

// ApprovalComment represents a comment on approval request
type ApprovalComment struct {
	Approver  string
	Comment   string
	Status    string // "approve", "reject", "comment"
	Timestamp time.Time
}

// SLAMonitor monitors SLA compliance
type SLAMonitor struct {
	executions map[string]SLAStatus
}

// SLAStatus tracks SLA status for an execution
type SLAStatus struct {
	ExecutionID string
	TargetTime  time.Time
	CurrentTime time.Duration
	Status      string // "on-track", "at-risk", "violated"
	AlertSent   bool
}

// ErrorHandler handles workflow errors
type ErrorHandler struct {
	handlers map[string]func(error) error
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		policies:     make([]*WorkflowPolicy, 0),
		executions:   make(map[string]WorkflowExecution),
		approvals:    make(map[string]ApprovalRequest),
		slaMonitor:   &SLAMonitor{executions: make(map[string]SLAStatus)},
		errorHandler: &ErrorHandler{handlers: make(map[string]func(error) error)},
	}
}

// RegisterPolicy registers a workflow policy
func (we *WorkflowEngine) RegisterPolicy(policy *WorkflowPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	we.mu.Lock()
	defer we.mu.Unlock()
	we.policies = append(we.policies, policy)
	return nil
}

// ExecuteWorkflow executes a workflow
func (we *WorkflowEngine) ExecuteWorkflow(workflowID string, context map[string]interface{}) (string, error) {
	we.mu.Lock()
	defer we.mu.Unlock()

	// Find workflow
	var workflow *WorkflowDefinition
	for _, policy := range we.policies {
		for i, w := range policy.Workflows {
			if w.ID == workflowID {
				workflow = &policy.Workflows[i]
				break
			}
		}
	}

	if workflow == nil {
		return "", fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Create execution
	executionID := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	execution := WorkflowExecution{
		ID:             executionID,
		WorkflowID:     workflowID,
		Status:         "running",
		StartTime:      time.Now(),
		Context:        context,
		CompletedSteps: make([]string, 0),
		FailedSteps:    make([]string, 0),
		Errors:         make([]WorkflowError, 0),
	}

	we.executions[executionID] = execution

	// Execute first step
	if len(workflow.Steps) > 0 {
		we.executeStep(executionID, workflow.Steps[0].ID, workflow)
	}

	return executionID, nil
}

func (we *WorkflowEngine) executeStep(executionID, stepID string, workflow *WorkflowDefinition) {
	exec, exists := we.executions[executionID]
	if !exists {
		return
	}

	// Find step
	var step *WorkflowStep
	for i, s := range workflow.Steps {
		if s.ID == stepID {
			step = &workflow.Steps[i]
			break
		}
	}

	if step == nil {
		return
	}

	exec.CurrentStep = stepID

	// Execute based on step type
	switch step.Type {
	case "action":
		// Execute action
		exec.CompletedSteps = append(exec.CompletedSteps, stepID)
	case "approval":
		// Create approval request
		approvalReq := ApprovalRequest{
			ID:          fmt.Sprintf("apr-%d", time.Now().UnixNano()),
			ExecutionID: executionID,
			Requestor:   "system",
			CreatedAt:   time.Now(),
			DueDate:     time.Now().Add(24 * time.Hour),
			Status:      "pending",
			Approvers:   make(map[string]bool),
			Comments:    make([]ApprovalComment, 0),
		}
		we.approvals[approvalReq.ID] = approvalReq
		exec.Status = "paused"
	case "decision":
		// Evaluate condition and continue
		exec.CompletedSteps = append(exec.CompletedSteps, stepID)
	}

	we.executions[executionID] = exec
}

// GetExecution returns execution details
func (we *WorkflowEngine) GetExecution(executionID string) (WorkflowExecution, error) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	exec, exists := we.executions[executionID]
	if !exists {
		return WorkflowExecution{}, fmt.Errorf("execution not found: %s", executionID)
	}

	return exec, nil
}

// ApproveStep approves a step in workflow
func (we *WorkflowEngine) ApproveStep(executionID string, approver string) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	exec, exists := we.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found")
	}

	exec.Approvals[approver] = true
	we.executions[executionID] = exec

	return nil
}

// CancelWorkflow cancels a workflow execution
func (we *WorkflowEngine) CancelWorkflow(executionID string) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	exec, exists := we.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found")
	}

	exec.Status = "cancelled"
	exec.EndTime = time.Now()
	we.executions[executionID] = exec

	return nil
}

// PauseWorkflow pauses a workflow execution
func (we *WorkflowEngine) PauseWorkflow(executionID string) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	exec, exists := we.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found")
	}

	exec.Status = "paused"
	we.executions[executionID] = exec

	return nil
}

// ResumeWorkflow resumes a paused workflow
func (we *WorkflowEngine) ResumeWorkflow(executionID string) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	exec, exists := we.executions[executionID]
	if !exists {
		return fmt.Errorf("execution not found")
	}

	if exec.Status != "paused" {
		return fmt.Errorf("workflow is not paused")
	}

	exec.Status = "running"
	we.executions[executionID] = exec

	return nil
}

// OrchestrationEngine provides advanced orchestration capabilities
type OrchestrationEngine struct {
	dependencies map[string][]string // task -> dependencies
	execOrder    []string
}

// NewOrchestrationEngine creates a new orchestration engine
func NewOrchestrationEngine() *OrchestrationEngine {
	return &OrchestrationEngine{
		dependencies: make(map[string][]string),
	}
}

// AddDependency adds a dependency between tasks
func (oe *OrchestrationEngine) AddDependency(task string, dependsOn []string) {
	oe.dependencies[task] = dependsOn
}

// ComputeExecutionOrder computes the execution order based on dependencies
func (oe *OrchestrationEngine) ComputeExecutionOrder() ([]string, error) {
	// Simple topological sort
	visited := make(map[string]bool)
	var order []string

	for task := range oe.dependencies {
		if !visited[task] {
			err := oe.visit(task, visited, &order)
			if err != nil {
				return nil, err
			}
		}
	}

	oe.execOrder = order
	return order, nil
}

func (oe *OrchestrationEngine) visit(task string, visited map[string]bool, order *[]string) error {
	visited[task] = true

	for _, dep := range oe.dependencies[task] {
		if !visited[dep] {
			oe.visit(dep, visited, order)
		}
	}

	*order = append(*order, task)
	return nil
}
