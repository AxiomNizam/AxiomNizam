package agents

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// AgentStatus represents the status of a cloud agent
type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusUnknown AgentStatus = "unknown"
)

// CloudAgent represents a registered remote cloud agent
type CloudAgent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Endpoint     string            `json:"endpoint"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Status       AgentStatus       `json:"status"`
	RegisteredAt time.Time         `json:"registeredAt"`
	LastSeenAt   *time.Time        `json:"lastSeenAt,omitempty"`
}

// DelegatedTask represents a task delegated to a cloud agent
type DelegatedTask struct {
	ID           string                 `json:"id"`
	AgentID      string                 `json:"agentId"`
	Status       string                 `json:"status"` // "pending", "running", "succeeded", "failed"
	Payload      map[string]interface{} `json:"payload"`
	Result       map[string]interface{} `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
	DelegatedAt  time.Time              `json:"delegatedAt"`
	CompletedAt  *time.Time             `json:"completedAt,omitempty"`
	RemoteTaskID string                 `json:"remoteTaskId,omitempty"`
}

// Common errors
var (
	ErrAgentNotFound    = errors.New("cloud agent not found")
	ErrAgentExists      = errors.New("cloud agent already exists")
	ErrTaskNotFound     = errors.New("delegated task not found")
	ErrAgentOffline     = errors.New("cloud agent is offline")
	ErrInvalidAgent     = errors.New("invalid cloud agent")
)

// Registry manages registered cloud agents
type Registry struct {
	mu     sync.RWMutex
	agents map[string]*CloudAgent // id -> agent
	tasks  map[string]*DelegatedTask // task id -> task
}

// NewRegistry creates a new agent registry
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]*CloudAgent),
		tasks:  make(map[string]*DelegatedTask),
	}
}

// Register adds a cloud agent to the registry
func (r *Registry) Register(agent *CloudAgent) error {
	if agent == nil || agent.Name == "" || agent.Endpoint == "" {
		return ErrInvalidAgent
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if agent.ID == "" {
		agent.ID = generateAgentID(agent.Name)
	}

	if _, exists := r.agents[agent.ID]; exists {
		return fmt.Errorf("%w: %s", ErrAgentExists, agent.ID)
	}

	if agent.Status == "" {
		agent.Status = AgentStatusUnknown
	}
	agent.RegisteredAt = time.Now()

	r.agents[agent.ID] = agent
	return nil
}

// Get retrieves a cloud agent by ID
func (r *Registry) Get(id string) (*CloudAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[id]
	if !ok {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

// List returns all registered cloud agents
func (r *Registry) List() []*CloudAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*CloudAgent, 0, len(r.agents))
	for _, a := range r.agents {
		list = append(list, a)
	}
	return list
}

// Unregister removes a cloud agent from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.agents[id]; !ok {
		return ErrAgentNotFound
	}
	delete(r.agents, id)
	return nil
}

// UpdateStatus updates the status of a cloud agent
func (r *Registry) UpdateStatus(id string, status AgentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, ok := r.agents[id]
	if !ok {
		return ErrAgentNotFound
	}

	agent.Status = status
	now := time.Now()
	agent.LastSeenAt = &now
	return nil
}

// AddTask records a delegated task
func (r *Registry) AddTask(task *DelegatedTask) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
}

// GetTask retrieves a delegated task by ID
func (r *Registry) GetTask(id string) (*DelegatedTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// UpdateTask updates a delegated task
func (r *Registry) UpdateTask(task *DelegatedTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[task.ID]; !ok {
		return ErrTaskNotFound
	}
	r.tasks[task.ID] = task
	return nil
}

// ListTasksByAgent returns all tasks for a given agent
func (r *Registry) ListTasksByAgent(agentID string) []*DelegatedTask {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*DelegatedTask, 0)
	for _, t := range r.tasks {
		if t.AgentID == agentID {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

// generateAgentID produces a unique agent ID
func generateAgentID(name string) string {
	return fmt.Sprintf("agent-%s-%d", sanitize(name), time.Now().UnixNano())
}

// sanitize replaces spaces and special characters with hyphens
func sanitize(s string) string {
	result := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result[i] = c
		} else if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = '-'
		}
	}
	return string(result)
}

// GlobalRegistry is the package-level agent registry
var GlobalRegistry = NewRegistry()
