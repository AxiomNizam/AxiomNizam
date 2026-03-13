package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DelegateRequest is the payload sent to a cloud agent when delegating a task
type DelegateRequest struct {
	TaskID  string                 `json:"taskId"`
	Payload map[string]interface{} `json:"payload"`
}

// DelegateResponse is the response received from the cloud agent after accepting a task
type DelegateResponse struct {
	RemoteTaskID string `json:"remoteTaskId"`
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
}

// TaskStatusResponse is the response from the cloud agent when polling task status
type TaskStatusResponse struct {
	RemoteTaskID string                 `json:"remoteTaskId"`
	Status       string                 `json:"status"` // "pending", "running", "succeeded", "failed"
	Result       map[string]interface{} `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// Executor delegates workflow steps to remote cloud agents
type Executor struct {
	registry   *Registry
	httpClient *http.Client
	pollInterval time.Duration
}

// NewExecutor creates a new cloud agent executor
func NewExecutor(registry *Registry) *Executor {
	return &Executor{
		registry: registry,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		pollInterval: 5 * time.Second,
	}
}

// Delegate sends a task to the specified cloud agent and waits for completion.
// It returns the task result on success or an error on failure.
func (e *Executor) Delegate(ctx context.Context, agentID string, payload map[string]interface{}) (*DelegatedTask, error) {
	agent, err := e.registry.Get(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent lookup failed: %w", err)
	}

	task := &DelegatedTask{
		ID:          generateTaskID(),
		AgentID:     agentID,
		Status:      "pending",
		Payload:     payload,
		DelegatedAt: time.Now(),
	}
	e.registry.AddTask(task)

	// Send task to the remote agent
	remoteID, err := e.sendTask(ctx, agent, task)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		now := time.Now()
		task.CompletedAt = &now
		_ = e.registry.UpdateTask(task)
		return task, fmt.Errorf("failed to delegate task to agent %s: %w", agentID, err)
	}

	task.RemoteTaskID = remoteID
	task.Status = "running"
	_ = e.registry.UpdateTask(task)

	// Update agent last seen timestamp
	_ = e.registry.UpdateStatus(agentID, AgentStatusOnline)

	// Poll until the remote task completes or context is cancelled
	result, pollErr := e.pollUntilDone(ctx, agent, remoteID)

	now := time.Now()
	task.CompletedAt = &now

	if pollErr != nil {
		task.Status = "failed"
		task.Error = pollErr.Error()
		_ = e.registry.UpdateTask(task)
		return task, pollErr
	}

	task.Status = result.Status
	task.Result = result.Result
	if result.Error != "" {
		task.Error = result.Error
	}
	_ = e.registry.UpdateTask(task)

	if task.Status == "failed" {
		return task, fmt.Errorf("remote task failed: %s", task.Error)
	}

	return task, nil
}

// DelegateAsync sends a task to the specified cloud agent without waiting for completion.
// It returns the local task record immediately.
func (e *Executor) DelegateAsync(ctx context.Context, agentID string, payload map[string]interface{}) (*DelegatedTask, error) {
	agent, err := e.registry.Get(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent lookup failed: %w", err)
	}

	task := &DelegatedTask{
		ID:          generateTaskID(),
		AgentID:     agentID,
		Status:      "pending",
		Payload:     payload,
		DelegatedAt: time.Now(),
	}
	e.registry.AddTask(task)

	remoteID, err := e.sendTask(ctx, agent, task)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		now := time.Now()
		task.CompletedAt = &now
		_ = e.registry.UpdateTask(task)
		return task, fmt.Errorf("failed to delegate task to agent %s: %w", agentID, err)
	}

	task.RemoteTaskID = remoteID
	task.Status = "running"
	_ = e.registry.UpdateTask(task)
	_ = e.registry.UpdateStatus(agentID, AgentStatusOnline)

	return task, nil
}

// PollTaskStatus fetches the latest status of a delegated task from the remote agent
// and updates the local task record.
func (e *Executor) PollTaskStatus(ctx context.Context, taskID string) (*DelegatedTask, error) {
	task, err := e.registry.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	if task.Status == "succeeded" || task.Status == "failed" {
		return task, nil
	}

	agent, err := e.registry.Get(task.AgentID)
	if err != nil {
		return task, fmt.Errorf("agent lookup failed: %w", err)
	}

	status, err := e.fetchTaskStatus(ctx, agent, task.RemoteTaskID)
	if err != nil {
		return task, fmt.Errorf("status poll failed: %w", err)
	}

	task.Status = status.Status
	task.Result = status.Result
	if status.Error != "" {
		task.Error = status.Error
	}
	if task.Status == "succeeded" || task.Status == "failed" {
		now := time.Now()
		task.CompletedAt = &now
	}
	_ = e.registry.UpdateTask(task)

	return task, nil
}

// sendTask posts a task to the remote cloud agent and returns the remote task ID
func (e *Executor) sendTask(ctx context.Context, agent *CloudAgent, task *DelegatedTask) (string, error) {
	reqBody := DelegateRequest{
		TaskID:  task.ID,
		Payload: task.Payload,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/tasks", agent.Endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("agent returned status %d", resp.StatusCode)
	}

	var delegateResp DelegateResponse
	if err := json.NewDecoder(resp.Body).Decode(&delegateResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return delegateResp.RemoteTaskID, nil
}

// fetchTaskStatus polls the remote agent for the current task status
func (e *Executor) fetchTaskStatus(ctx context.Context, agent *CloudAgent, remoteTaskID string) (*TaskStatusResponse, error) {
	url := fmt.Sprintf("%s/tasks/%s", agent.Endpoint, remoteTaskID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("agent returned status %d", resp.StatusCode)
	}

	var statusResp TaskStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &statusResp, nil
}

// pollUntilDone polls the remote agent until the task completes or context is cancelled
func (e *Executor) pollUntilDone(ctx context.Context, agent *CloudAgent, remoteTaskID string) (*TaskStatusResponse, error) {
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := e.fetchTaskStatus(ctx, agent, remoteTaskID)
			if err != nil {
				return nil, err
			}
			if status.Status == "succeeded" || status.Status == "failed" {
				return status, nil
			}
		}
	}
}

// generateTaskID produces a unique task ID
func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

// GlobalExecutor is the package-level executor backed by GlobalRegistry
var GlobalExecutor = NewExecutor(GlobalRegistry)
