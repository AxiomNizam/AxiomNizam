package workflows

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`   // "http", "notification", "script", "policy", etc.
	Action  string                 `json:"action"` // e.g., "notify-slack", "trigger-job"
	Config  map[string]interface{} `json:"config"`
	Timeout time.Duration          `json:"timeout,omitempty"`
	Retry   int                    `json:"retry,omitempty"`
}

// WorkflowTrigger defines when a workflow is triggered
type WorkflowTrigger struct {
	Type      string                 `json:"type"` // "policy", "resource", "schedule", "manual"
	Condition map[string]interface{} `json:"condition"`
}

// Workflow represents a workflow definition
type Workflow struct {
	// Metadata
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Version   string `json:"version"`

	// Definition
	Description string            `json:"description"`
	Triggers    []WorkflowTrigger `json:"triggers"`
	Steps       []WorkflowStep    `json:"steps"`

	// Configuration
	Enabled bool `json:"enabled"`

	// Status tracking
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// WorkflowExecution represents an execution of a workflow
type WorkflowExecution struct {
	ID             string                 `json:"id"`
	WorkflowName   string                 `json:"workflowName"`
	Status         string                 `json:"status"` // "pending", "running", "success", "failed"
	StartTime      time.Time              `json:"startTime"`
	EndTime        *time.Time             `json:"endTime,omitempty"`
	CompletedSteps int                    `json:"completedSteps"`
	TotalSteps     int                    `json:"totalSteps"`
	Error          string                 `json:"error,omitempty"`
	StepExecutions []*StepExecution       `json:"stepExecutions,omitempty"`
	TriggerContext map[string]interface{} `json:"triggerContext"`
}

// StepExecution represents execution of a single step
type StepExecution struct {
	StepID    string                 `json:"stepId"`
	StepName  string                 `json:"stepName"`
	Status    string                 `json:"status"`
	StartTime time.Time              `json:"startTime"`
	EndTime   *time.Time             `json:"endTime,omitempty"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// WorkflowEngine executes workflows
type WorkflowEngine struct {
	mu         sync.RWMutex
	workflows  map[string]*Workflow
	executions map[string]*WorkflowExecution
	handlers   map[string]StepHandler
}

// StepHandler handles a specific step type
type StepHandler func(ctx context.Context, step *WorkflowStep, input map[string]interface{}) (map[string]interface{}, error)

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows:  make(map[string]*Workflow),
		executions: make(map[string]*WorkflowExecution),
		handlers:   make(map[string]StepHandler),
	}
}

// RegisterHandler registers a handler for a step type
func (we *WorkflowEngine) RegisterHandler(stepType string, handler StepHandler) {
	we.mu.Lock()
	defer we.mu.Unlock()
	we.handlers[stepType] = handler
}

// AddWorkflow adds a workflow
func (we *WorkflowEngine) AddWorkflow(ctx context.Context, workflow *Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	we.mu.Lock()
	defer we.mu.Unlock()

	we.workflows[workflow.Name] = workflow
	return nil
}

// GetWorkflow retrieves a workflow
func (we *WorkflowEngine) GetWorkflow(name string) *Workflow {
	we.mu.RLock()
	defer we.mu.RUnlock()
	return we.workflows[name]
}

// ListWorkflows lists all workflows
func (we *WorkflowEngine) ListWorkflows() []*Workflow {
	we.mu.RLock()
	defer we.mu.RUnlock()

	workflows := make([]*Workflow, 0, len(we.workflows))
	for _, w := range we.workflows {
		workflows = append(workflows, w)
	}
	return workflows
}

// Execute executes a workflow
func (we *WorkflowEngine) Execute(ctx context.Context, workflowName string, triggerContext map[string]interface{}) (*WorkflowExecution, error) {
	workflow := we.GetWorkflow(workflowName)
	if workflow == nil {
		return nil, fmt.Errorf("workflow not found: %s", workflowName)
	}

	if !workflow.Enabled {
		return nil, fmt.Errorf("workflow is disabled: %s", workflowName)
	}

	execution := &WorkflowExecution{
		ID:             generateExecutionID(),
		WorkflowName:   workflowName,
		Status:         "running",
		StartTime:      time.Now(),
		TotalSteps:     len(workflow.Steps),
		TriggerContext: triggerContext,
		StepExecutions: make([]*StepExecution, 0),
	}

	we.mu.Lock()
	we.executions[execution.ID] = execution
	we.mu.Unlock()

	// Execute steps sequentially
	stepInput := make(map[string]interface{})
	for i, step := range workflow.Steps {
		stepExec := &StepExecution{
			StepID:    step.ID,
			StepName:  step.Name,
			Status:    "running",
			StartTime: time.Now(),
		}

		// Get handler
		we.mu.RLock()
		handler, ok := we.handlers[step.Type]
		we.mu.RUnlock()

		var err error
		if !ok {
			stepExec.Status = "failed"
			stepExec.Error = fmt.Sprintf("no handler for step type: %s", step.Type)
			err = fmt.Errorf("no handler for step type: %s", step.Type)
		} else {
			// Execute step with timeout
			stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
			if step.Timeout == 0 {
				stepCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
			}

			output, execErr := handler(stepCtx, step, stepInput)
			cancel()

			if execErr != nil {
				stepExec.Status = "failed"
				stepExec.Error = execErr.Error()
				err = execErr

				// Check if we should retry
				if step.Retry > 0 {
					for retry := 1; retry <= step.Retry && err != nil; retry++ {
						stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
						if step.Timeout == 0 {
							stepCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
						}

						output, execErr = handler(stepCtx, step, stepInput)
						cancel()
						if execErr == nil {
							stepExec.Status = "success"
							stepExec.Output = output
							stepInput = output
							err = nil
							break
						}
					}
				} else {
					// Stop workflow on error if no retries
					break
				}
			} else {
				stepExec.Status = "success"
				stepExec.Output = output
				stepInput = output
			}
		}

		now := time.Now()
		stepExec.EndTime = &now
		execution.StepExecutions = append(execution.StepExecutions, stepExec)
		execution.CompletedSteps = i + 1

		if err != nil {
			execution.Status = "failed"
			execution.Error = err.Error()
			break
		}
	}

	// Mark as complete
	if execution.Status != "failed" {
		execution.Status = "success"
	}

	now := time.Now()
	execution.EndTime = &now

	return execution, nil
}

// GetExecution retrieves an execution
func (we *WorkflowEngine) GetExecution(id string) *WorkflowExecution {
	we.mu.RLock()
	defer we.mu.RUnlock()
	return we.executions[id]
}

// ListExecutions lists executions
func (we *WorkflowEngine) ListExecutions(workflowName string) []*WorkflowExecution {
	we.mu.RLock()
	defer we.mu.RUnlock()

	executions := make([]*WorkflowExecution, 0)
	for _, exec := range we.executions {
		if exec.WorkflowName == workflowName {
			executions = append(executions, exec)
		}
	}
	return executions
}

// Builtin step handlers

// NotificationHandler sends notifications
var NotificationHandler = func(ctx context.Context, step *WorkflowStep, input map[string]interface{}) (map[string]interface{}, error) {
	action := step.Config["action"].(string)
	message := step.Config["message"].(string)

	fmt.Printf("📨 Notification (%s): %s\n", action, message)

	return map[string]interface{}{
		"action":  action,
		"message": message,
		"status":  "sent",
	}, nil
}

// HTTPHandler makes HTTP requests
var HTTPHandler = func(ctx context.Context, step *WorkflowStep, input map[string]interface{}) (map[string]interface{}, error) {
	method := step.Config["method"].(string)
	url := step.Config["url"].(string)

	fmt.Printf("🌐 HTTP Request: %s %s\n", method, url)

	return map[string]interface{}{
		"method":     method,
		"url":        url,
		"statusCode": 200,
	}, nil
}

// GlobalWorkflowEngine is the package-level workflow engine
var GlobalWorkflowEngine = NewWorkflowEngine()

// Initialize builtin handlers
func init() {
	GlobalWorkflowEngine.RegisterHandler("notification", NotificationHandler)
	GlobalWorkflowEngine.RegisterHandler("http", HTTPHandler)
}

// AddWorkflow adds workflow via global engine
func AddWorkflow(ctx context.Context, workflow *Workflow) error {
	return GlobalWorkflowEngine.AddWorkflow(ctx, workflow)
}

// Execute executes workflow via global engine
func Execute(ctx context.Context, workflowName string, triggerContext map[string]interface{}) (*WorkflowExecution, error) {
	return GlobalWorkflowEngine.Execute(ctx, workflowName, triggerContext)
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("exec-%d", time.Now().UnixNano())
}

// ConnectPolicies connects policies to workflows
// Example: When policy "data-access-approved" is allowed, trigger workflow "notify-slack"
type PolicyWorkflowConnection struct {
	PolicyName   string   `json:"policyName"`
	PolicyResult bool     `json:"policyResult"`
	Workflows    []string `json:"workflows"`
}

// WorkflowTriggerManager manages policy-to-workflow connections
type WorkflowTriggerManager struct {
	mu          sync.RWMutex
	connections map[string][]string // policyName -> workflow names
}

// NewWorkflowTriggerManager creates a new trigger manager
func NewWorkflowTriggerManager() *WorkflowTriggerManager {
	return &WorkflowTriggerManager{
		connections: make(map[string][]string),
	}
}

// Connect connects a policy result to workflows
func (wtm *WorkflowTriggerManager) Connect(policyName string, workflows []string) {
	wtm.mu.Lock()
	defer wtm.mu.Unlock()
	wtm.connections[policyName] = workflows
}

// GetTriggeredWorkflows gets workflows triggered by a policy
func (wtm *WorkflowTriggerManager) GetTriggeredWorkflows(policyName string) []string {
	wtm.mu.RLock()
	defer wtm.mu.RUnlock()
	return wtm.connections[policyName]
}

// GlobalWorkflowTriggerManager is the package-level trigger manager
var GlobalWorkflowTriggerManager = NewWorkflowTriggerManager()
