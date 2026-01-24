package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Scheduler implements job scheduling with cron-like expressions
type SimpleScheduler struct {
	mu        sync.RWMutex
	scheduled map[string]*ScheduledJob
	queue     Queue
	logger    *log.Logger
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewSimpleScheduler creates a new scheduler
func NewSimpleScheduler() *SimpleScheduler {
	return &SimpleScheduler{
		scheduled: make(map[string]*ScheduledJob),
		logger:    log.New(log.Writer(), "[SCHEDULER] ", log.LstdFlags),
		running:   false,
	}
}

// Schedule schedules a recurring job (simplified cron: every X minutes/hours)
func (ss *SimpleScheduler) Schedule(jobType JobType, interval string, data map[string]interface{}) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Parse interval (simplified: "5m", "1h", "30s")
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("invalid interval format: %s", interval)
	}

	key := fmt.Sprintf("%s_%s", jobType, interval)
	ss.scheduled[key] = &ScheduledJob{
		ID:       key,
		Type:     jobType,
		CronExpr: interval,
		Data:     data,
		Enabled:  true,
		NextRun:  time.Now().Add(duration),
	}

	ss.logger.Printf("Job scheduled: %s (interval: %s)", jobType, interval)
	return nil
}

// Unschedule removes a scheduled job
func (ss *SimpleScheduler) Unschedule(jobType JobType, cronExpr string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	key := fmt.Sprintf("%s_%s", jobType, cronExpr)
	if _, exists := ss.scheduled[key]; !exists {
		return fmt.Errorf("scheduled job not found: %s", key)
	}

	delete(ss.scheduled, key)
	ss.logger.Printf("Scheduled job removed: %s", key)
	return nil
}

// Start starts the scheduler
func (ss *SimpleScheduler) Start(ctx context.Context, queue Queue) error {
	ss.mu.Lock()
	if ss.running {
		ss.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}

	ss.running = true
	ss.queue = queue
	ss.ctx, ss.cancel = context.WithCancel(ctx)
	ss.mu.Unlock()

	// Start scheduler loop
	go ss.run()

	ss.logger.Printf("Scheduler started")
	return nil
}

// run executes the scheduler loop
func (ss *SimpleScheduler) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ss.ctx.Done():
			ss.logger.Printf("Scheduler stopping")
			return
		case <-ticker.C:
			ss.checkAndSubmitJobs()
		}
	}
}

// checkAndSubmitJobs checks for jobs that need to be scheduled
func (ss *SimpleScheduler) checkAndSubmitJobs() {
	ss.mu.Lock()
	scheduled := make([]*ScheduledJob, 0)
	for _, job := range ss.scheduled {
		if job.Enabled && time.Now().After(job.NextRun) {
			scheduled = append(scheduled, job)
		}
	}
	ss.mu.Unlock()

	now := time.Now()
	for _, scheduled := range scheduled {
		job := CreateJobWithPriority(
			scheduled.Type,
			scheduled.Data,
			PriorityNormal,
		)

		if err := ss.queue.Submit(ss.ctx, job); err != nil {
			ss.logger.Printf("Error submitting scheduled job: %v", err)
		} else {
			ss.logger.Printf("Scheduled job submitted: %s", scheduled.Type)

			// Update next run time
			duration, _ := time.ParseDuration(scheduled.CronExpr)
			ss.mu.Lock()
			scheduled.LastRun = now
			scheduled.NextRun = now.Add(duration)
			ss.mu.Unlock()
		}
	}
}

// Stop stops the scheduler
func (ss *SimpleScheduler) Stop() error {
	ss.mu.Lock()
	if !ss.running {
		ss.mu.Unlock()
		return fmt.Errorf("scheduler not running")
	}

	ss.running = false
	ss.mu.Unlock()

	ss.cancel()
	ss.logger.Printf("Scheduler stopped")
	return nil
}

// ListScheduled lists all scheduled jobs
func (ss *SimpleScheduler) ListScheduled() []ScheduledJob {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	var result []ScheduledJob
	for _, job := range ss.scheduled {
		result = append(result, *job)
	}

	return result
}

// JobManager manages all job operations
type JobManager struct {
	queue      Queue
	processor  Processor
	scheduler  Scheduler
	logger     *log.Logger
	config     *JobConfig
	emailQueue chan *Job
	resultChan chan *JobResult
}

// NewJobManager creates a new job manager
func NewJobManager(config *JobConfig) *JobManager {
	if config == nil {
		config = DefaultJobConfig()
	}

	return &JobManager{
		queue:      NewMemoryQueue(config.MaxQueueSize),
		processor:  NewMemoryProcessor(NewMemoryQueue(config.MaxQueueSize)),
		scheduler:  NewSimpleScheduler(),
		logger:     log.New(log.Writer(), "[JOB_MANAGER] ", log.LstdFlags),
		config:     config,
		emailQueue: make(chan *Job, 100),
		resultChan: make(chan *JobResult, 100),
	}
}

// Submit submits a job to the queue
func (jm *JobManager) Submit(ctx context.Context, job *Job) error {
	return jm.queue.Submit(ctx, job)
}

// SubmitEmail submits an email job
func (jm *JobManager) SubmitEmail(ctx context.Context, to string, subject string, body string) error {
	job := CreateJobWithPriority(JobTypeEmail, map[string]interface{}{
		"to":      to,
		"subject": subject,
		"body":    body,
	}, PriorityHigh)

	return jm.Submit(ctx, job)
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(ctx context.Context, jobID string) (*Job, error) {
	return jm.queue.Get(ctx, jobID)
}

// GetJobStats retrieves job statistics
func (jm *JobManager) GetJobStats(ctx context.Context) (*QueueStats, error) {
	return jm.queue.GetStats(ctx)
}

// StartWorkers starts job processing
func (jm *JobManager) StartWorkers(ctx context.Context, numWorkers int) error {
	if numWorkers <= 0 {
		numWorkers = jm.config.NumWorkers
	}

	return jm.processor.Start(ctx, numWorkers)
}

// StopWorkers stops job processing
func (jm *JobManager) StopWorkers() error {
	return jm.processor.Stop()
}

// RegisterHandler registers a job handler
func (jm *JobManager) RegisterHandler(jobType JobType, handler JobHandler) {
	jm.processor.Register(jobType, handler)
}

// ScheduleJob schedules a recurring job
func (jm *JobManager) ScheduleJob(jobType JobType, interval string, data map[string]interface{}) error {
	return jm.scheduler.Schedule(jobType, interval, data)
}

// StartScheduler starts the job scheduler
func (jm *JobManager) StartScheduler(ctx context.Context) error {
	return jm.scheduler.Start(ctx, jm.queue)
}

// StopScheduler stops the job scheduler
func (jm *JobManager) StopScheduler() error {
	return jm.scheduler.Stop()
}

// GetProcessorStats returns processor statistics
func (jm *JobManager) GetProcessorStats() *ProcessorStats {
	return jm.processor.GetStats()
}

// Health checks the health of job system
func (jm *JobManager) Health() error {
	stats, err := jm.queue.GetStats(context.Background())
	if err != nil {
		return err
	}

	if stats.Failed > 0 && float64(stats.Failed)/float64(stats.Total) > 0.5 {
		return fmt.Errorf("high failure rate: %d/%d", stats.Failed, stats.Total)
	}

	return nil
}

// GetResults returns a channel for job results
func (jm *JobManager) GetResults() <-chan *JobResult {
	if mp, ok := jm.processor.(*MemoryProcessor); ok {
		return mp.GetResults()
	}
	return jm.resultChan
}
