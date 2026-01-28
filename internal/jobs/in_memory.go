package jobs

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryJobManager in-memory job implementation
type InMemoryJobManager struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	logs map[string][]*JobLog
}

// NewInMemoryJobManager creates manager
func NewInMemoryJobManager() *InMemoryJobManager {
	return &InMemoryJobManager{
		jobs: make(map[string]*Job),
		logs: make(map[string][]*JobLog),
	}
}

// SubmitJob submits new job
func (m *InMemoryJobManager) SubmitJob(job *Job) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%d", time.Now().UnixNano())
	}

	job.Status = "Pending"
	job.CreatedAt = time.Now()
	m.jobs[job.ID] = job
	m.logs[job.ID] = []*JobLog{}
	return job, nil
}

// GetJob retrieves job
func (m *InMemoryJobManager) GetJob(id string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}
	return job, nil
}

// ListJobs lists jobs matching filter
func (m *InMemoryJobManager) ListJobs(filter *JobFilter) ([]*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Job
	for _, job := range m.jobs {
		// Note: Job doesn't have TenantID, filtering by other fields
		if filter.Status != "" && JobStatus(filter.Status) != job.Status {
			continue
		}
		if filter.Type != "" && job.Type != JobType(filter.Type) {
			continue
		}
		result = append(result, job)
	}
	return result, nil
}

// CancelJob cancels job
func (m *InMemoryJobManager) CancelJob(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[id]
	if !exists {
		return fmt.Errorf("job not found")
	}

	job.Status = "Cancelled"
	job.CompletedAt = time.Now()
	return nil
}

// RetryJob retries failed job
func (m *InMemoryJobManager) RetryJob(id string) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}

	job.Status = JobStatus("Queued")
	job.Retries++
	return job, nil
}

// GetJobLogs retrieves job logs
func (m *InMemoryJobManager) GetJobLogs(id string) ([]*JobLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs, exists := m.logs[id]
	if !exists {
		return nil, fmt.Errorf("logs not found")
	}
	return logs, nil
}
