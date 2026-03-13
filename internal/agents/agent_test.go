package agents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Registry tests
// ---------------------------------------------------------------------------

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()

	agent := &CloudAgent{
		Name:     "test-agent",
		Endpoint: "http://localhost:9000",
	}
	if err := r.Register(agent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.ID == "" {
		t.Fatal("expected ID to be populated after Register")
	}

	got, err := r.Get(agent.ID)
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}
	if got.Name != "test-agent" {
		t.Errorf("expected name %q, got %q", "test-agent", got.Name)
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	agent := &CloudAgent{ID: "dup-id", Name: "a", Endpoint: "http://x"}
	_ = r.Register(agent)

	err := r.Register(&CloudAgent{ID: "dup-id", Name: "b", Endpoint: "http://y"})
	if err == nil {
		t.Fatal("expected error registering duplicate agent, got nil")
	}
}

func TestRegistry_RegisterInvalid(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(nil); err != ErrInvalidAgent {
		t.Fatalf("expected ErrInvalidAgent, got %v", err)
	}
	if err := r.Register(&CloudAgent{Name: "", Endpoint: "http://x"}); err != ErrInvalidAgent {
		t.Fatalf("expected ErrInvalidAgent for empty name, got %v", err)
	}
	if err := r.Register(&CloudAgent{Name: "a", Endpoint: ""}); err != ErrInvalidAgent {
		t.Fatalf("expected ErrInvalidAgent for empty endpoint, got %v", err)
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	if err != ErrAgentNotFound {
		t.Fatalf("expected ErrAgentNotFound, got %v", err)
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&CloudAgent{Name: "a", Endpoint: "http://a"})
	_ = r.Register(&CloudAgent{Name: "b", Endpoint: "http://b"})

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(list))
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	agent := &CloudAgent{Name: "c", Endpoint: "http://c"}
	_ = r.Register(agent)

	if err := r.Unregister(agent.ID); err != nil {
		t.Fatalf("Unregister: unexpected error: %v", err)
	}
	if _, err := r.Get(agent.ID); err != ErrAgentNotFound {
		t.Fatalf("expected ErrAgentNotFound after unregister, got %v", err)
	}
}

func TestRegistry_UpdateStatus(t *testing.T) {
	r := NewRegistry()
	agent := &CloudAgent{Name: "d", Endpoint: "http://d"}
	_ = r.Register(agent)

	if err := r.UpdateStatus(agent.ID, AgentStatusOnline); err != nil {
		t.Fatalf("UpdateStatus: unexpected error: %v", err)
	}
	got, _ := r.Get(agent.ID)
	if got.Status != AgentStatusOnline {
		t.Errorf("expected status %q, got %q", AgentStatusOnline, got.Status)
	}
	if got.LastSeenAt == nil {
		t.Error("expected LastSeenAt to be set")
	}
}

func TestRegistry_Tasks(t *testing.T) {
	r := NewRegistry()

	task := &DelegatedTask{
		ID:          "task-1",
		AgentID:     "agent-1",
		Status:      "pending",
		DelegatedAt: time.Now(),
	}
	r.AddTask(task)

	got, err := r.GetTask("task-1")
	if err != nil {
		t.Fatalf("GetTask: unexpected error: %v", err)
	}
	if got.ID != "task-1" {
		t.Errorf("expected task ID %q, got %q", "task-1", got.ID)
	}

	_, err = r.GetTask("nonexistent")
	if err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestRegistry_ListTasksByAgent(t *testing.T) {
	r := NewRegistry()
	r.AddTask(&DelegatedTask{ID: "t1", AgentID: "ag1", Status: "pending", DelegatedAt: time.Now()})
	r.AddTask(&DelegatedTask{ID: "t2", AgentID: "ag1", Status: "running", DelegatedAt: time.Now()})
	r.AddTask(&DelegatedTask{ID: "t3", AgentID: "ag2", Status: "pending", DelegatedAt: time.Now()})

	tasks := r.ListTasksByAgent("ag1")
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks for ag1, got %d", len(tasks))
	}
}

// ---------------------------------------------------------------------------
// Executor tests (using an in-process fake agent server)
// ---------------------------------------------------------------------------

// newFakeAgentServer creates a test HTTP server that simulates a cloud agent.
// On POST /tasks it returns a remoteTaskId; on GET /tasks/:id it returns "succeeded".
func newFakeAgentServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(DelegateResponse{
			RemoteTaskID: "remote-task-123",
			Status:       "pending",
		})
	})

	mux.HandleFunc("/tasks/remote-task-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TaskStatusResponse{
			RemoteTaskID: "remote-task-123",
			Status:       "succeeded",
			Result:       map[string]interface{}{"output": "done"},
		})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return httptest.NewServer(mux)
}

func TestExecutor_DelegateSync(t *testing.T) {
	srv := newFakeAgentServer(t)
	defer srv.Close()

	r := NewRegistry()
	agent := &CloudAgent{Name: "fake", Endpoint: srv.URL}
	_ = r.Register(agent)

	exec := NewExecutor(r)
	exec.pollInterval = 100 * time.Millisecond

	ctx := context.Background()
	task, err := exec.Delegate(ctx, agent.ID, map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("Delegate: unexpected error: %v", err)
	}
	if task.Status != "succeeded" {
		t.Errorf("expected status %q, got %q", "succeeded", task.Status)
	}
	if task.Result["output"] != "done" {
		t.Errorf("expected result output %q, got %v", "done", task.Result["output"])
	}
}

func TestExecutor_DelegateAsync(t *testing.T) {
	srv := newFakeAgentServer(t)
	defer srv.Close()

	r := NewRegistry()
	agent := &CloudAgent{Name: "fake-async", Endpoint: srv.URL}
	_ = r.Register(agent)

	exec := NewExecutor(r)

	ctx := context.Background()
	task, err := exec.DelegateAsync(ctx, agent.ID, map[string]interface{}{})
	if err != nil {
		t.Fatalf("DelegateAsync: unexpected error: %v", err)
	}
	if task.Status != "running" {
		t.Errorf("expected status %q, got %q", "running", task.Status)
	}
	if task.RemoteTaskID != "remote-task-123" {
		t.Errorf("expected remoteTaskID %q, got %q", "remote-task-123", task.RemoteTaskID)
	}
}

func TestExecutor_DelegateAgentNotFound(t *testing.T) {
	r := NewRegistry()
	exec := NewExecutor(r)

	_, err := exec.Delegate(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown agent, got nil")
	}
}
