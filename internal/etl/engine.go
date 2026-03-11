package etl

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// =====================================================
// ETL Engine — Extract, Transform, Load
// Dynamic pipelines with observability
// =====================================================

// --- Pipeline Models ---

type PipelineStatus string

const (
	PipelineCreated PipelineStatus = "created"
	PipelineRunning PipelineStatus = "running"
	PipelinePaused  PipelineStatus = "paused"
	PipelineSuccess PipelineStatus = "succeeded"
	PipelineFailed  PipelineStatus = "failed"
	PipelineStopped PipelineStatus = "stopped"
)

type StepType string

const (
	StepExtract   StepType = "extract"
	StepTransform StepType = "transform"
	StepLoad      StepType = "load"
	StepFilter    StepType = "filter"
	StepMap       StepType = "map"
	StepAggregate StepType = "aggregate"
	StepJoin      StepType = "join"
	StepValidate  StepType = "validate"
	StepEnrich    StepType = "enrich"
	StepDedupe    StepType = "deduplicate"
)

type Pipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []Step                 `json:"steps"`
	Schedule    string                 `json:"schedule,omitempty"` // cron expression
	Config      map[string]interface{} `json:"config,omitempty"`
	Status      PipelineStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastRunAt   *time.Time             `json:"last_run_at,omitempty"`
	RunCount    int                    `json:"run_count"`
	Tags        []string               `json:"tags,omitempty"`
}

type Step struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       StepType               `json:"type"`
	Connector  string                 `json:"connector"` // e.g. mysql, postgres, csv, api, kafka
	Config     map[string]interface{} `json:"config"`
	Order      int                    `json:"order"`
	DependsOn  []string               `json:"depends_on,omitempty"`
	RetryCount int                    `json:"retry_count,omitempty"`
	Timeout    string                 `json:"timeout,omitempty"` // duration string
}

// --- Run / Execution ---

type PipelineRun struct {
	ID          string         `json:"id"`
	PipelineID  string         `json:"pipeline_id"`
	Status      PipelineStatus `json:"status"`
	StartedAt   time.Time      `json:"started_at"`
	FinishedAt  *time.Time     `json:"finished_at,omitempty"`
	Duration    string         `json:"duration,omitempty"`
	StepResults []StepResult   `json:"step_results"`
	RowsRead    int64          `json:"rows_read"`
	RowsWritten int64          `json:"rows_written"`
	RowsFailed  int64          `json:"rows_failed"`
	ErrorMsg    string         `json:"error_msg,omitempty"`
	Trigger     string         `json:"trigger"` // manual, schedule, api, event
}

type StepResult struct {
	StepID     string      `json:"step_id"`
	StepName   string      `json:"step_name"`
	StepType   StepType    `json:"step_type"`
	Status     string      `json:"status"` // running, succeeded, failed, skipped
	StartedAt  time.Time   `json:"started_at"`
	FinishedAt *time.Time  `json:"finished_at,omitempty"`
	Duration   string      `json:"duration,omitempty"`
	RowsIn     int64       `json:"rows_in"`
	RowsOut    int64       `json:"rows_out"`
	RowsError  int64       `json:"rows_error"`
	ErrorMsg   string      `json:"error_msg,omitempty"`
	Metrics    StepMetrics `json:"metrics"`
}

type StepMetrics struct {
	BytesProcessed int64   `json:"bytes_processed"`
	ThroughputRPS  float64 `json:"throughput_rps"` // rows per second
	AvgRowSizeB    int64   `json:"avg_row_size_bytes"`
	MemoryUsedMB   float64 `json:"memory_used_mb"`
	CPUPercent     float64 `json:"cpu_percent"`
}

// --- Connector Registry ---

type ConnectorType struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`     // database, file, api, queue, stream
	SupportedAs []string `json:"supported_as"` // extract, load, both
	ConfigKeys  []string `json:"config_keys"`
	Icon        string   `json:"icon"`
}

// --- Observability ---

type ETLObservability struct {
	mu             sync.RWMutex
	PipelinesTotal int               `json:"pipelines_total"`
	RunsTotal      int               `json:"runs_total"`
	RunsSuccess    int               `json:"runs_success"`
	RunsFailed     int               `json:"runs_failed"`
	RunsRunning    int               `json:"runs_running"`
	TotalRowsRead  int64             `json:"total_rows_read"`
	TotalRowsWrite int64             `json:"total_rows_written"`
	AvgDuration    float64           `json:"avg_duration_seconds"`
	StepTypeStats  map[string]int    `json:"step_type_stats"`
	ErrorsByType   map[string]int    `json:"errors_by_type"`
	ThroughputLog  []ThroughputPoint `json:"throughput_log"`
	LastUpdated    time.Time         `json:"last_updated"`
}

type ThroughputPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	RowsPerSec float64   `json:"rows_per_sec"`
	Pipeline   string    `json:"pipeline"`
}

// --- Engine ---

type Engine struct {
	mu            sync.RWMutex
	pipelines     map[string]*Pipeline
	runs          map[string]*PipelineRun
	connectors    []ConnectorType
	observability *ETLObservability
	sequence      int64
}

func NewEngine() *Engine {
	e := &Engine{
		pipelines: make(map[string]*Pipeline),
		runs:      make(map[string]*PipelineRun),
		observability: &ETLObservability{
			StepTypeStats: make(map[string]int),
			ErrorsByType:  make(map[string]int),
			ThroughputLog: make([]ThroughputPoint, 0),
			LastUpdated:   time.Now(),
		},
	}
	e.registerConnectors()
	e.seedPipelines()
	return e
}

// --- Connector Registry ---

func (e *Engine) registerConnectors() {
	e.connectors = []ConnectorType{
		{ID: "mysql", Name: "MySQL", Category: "database", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"host", "port", "database", "table", "query"}, Icon: "🐬"},
		{ID: "postgres", Name: "PostgreSQL", Category: "database", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"host", "port", "database", "table", "query"}, Icon: "🐘"},
		{ID: "mariadb", Name: "MariaDB", Category: "database", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"host", "port", "database", "table", "query"}, Icon: "🦭"},
		{ID: "mongodb", Name: "MongoDB", Category: "database", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"uri", "database", "collection", "filter"}, Icon: "🍃"},
		{ID: "oracle", Name: "Oracle", Category: "database", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"host", "port", "service_name", "table", "query"}, Icon: "🔴"},
		{ID: "csv", Name: "CSV File", Category: "file", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"path", "delimiter", "encoding", "has_header"}, Icon: "📄"},
		{ID: "json", Name: "JSON File", Category: "file", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"path", "json_path"}, Icon: "📋"},
		{ID: "api", Name: "REST API", Category: "api", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"url", "method", "headers", "body", "auth"}, Icon: "🌐"},
		{ID: "kafka", Name: "Apache Kafka", Category: "stream", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"brokers", "topic", "group_id", "offset"}, Icon: "📨"},
		{ID: "redis", Name: "Redis/Valkey", Category: "queue", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"host", "port", "key", "type"}, Icon: "🔴"},
		{ID: "elasticsearch", Name: "Elasticsearch", Category: "search", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"url", "index", "query"}, Icon: "🔍"},
		{ID: "s3", Name: "S3/MinIO", Category: "storage", SupportedAs: []string{"extract", "load"}, ConfigKeys: []string{"endpoint", "bucket", "key", "region"}, Icon: "☁️"},
	}
}

func (e *Engine) GetConnectors() []ConnectorType {
	return e.connectors
}

// --- Pipeline CRUD ---

func (e *Engine) CreatePipeline(p *Pipeline) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if p.ID == "" {
		e.sequence++
		p.ID = fmt.Sprintf("etl-pipe-%d", e.sequence)
	}
	p.Status = PipelineCreated
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	e.pipelines[p.ID] = p
	e.observability.mu.Lock()
	e.observability.PipelinesTotal++
	e.observability.mu.Unlock()
	return nil
}

func (e *Engine) GetPipeline(id string) (*Pipeline, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	p, ok := e.pipelines[id]
	return p, ok
}

func (e *Engine) ListPipelines() []*Pipeline {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]*Pipeline, 0, len(e.pipelines))
	for _, p := range e.pipelines {
		result = append(result, p)
	}
	return result
}

func (e *Engine) UpdatePipeline(id string, updates map[string]interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.pipelines[id]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", id)
	}
	if name, ok := updates["name"].(string); ok && name != "" {
		p.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		p.Description = desc
	}
	if sched, ok := updates["schedule"].(string); ok {
		p.Schedule = sched
	}
	p.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) DeletePipeline(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.pipelines[id]; !ok {
		return fmt.Errorf("pipeline not found: %s", id)
	}
	delete(e.pipelines, id)
	e.observability.mu.Lock()
	e.observability.PipelinesTotal--
	e.observability.mu.Unlock()
	return nil
}

// --- Execution ---

func (e *Engine) RunPipeline(ctx context.Context, pipelineID string, trigger string) (*PipelineRun, error) {
	e.mu.Lock()
	p, ok := e.pipelines[pipelineID]
	if !ok {
		e.mu.Unlock()
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}
	e.sequence++
	runID := fmt.Sprintf("run-%d-%d", time.Now().UnixMilli(), e.sequence)
	now := time.Now()

	run := &PipelineRun{
		ID:          runID,
		PipelineID:  pipelineID,
		Status:      PipelineRunning,
		StartedAt:   now,
		StepResults: make([]StepResult, 0),
		Trigger:     trigger,
	}
	e.runs[runID] = run
	p.Status = PipelineRunning
	p.LastRunAt = &now
	p.RunCount++
	e.mu.Unlock()

	// Update observability
	e.observability.mu.Lock()
	e.observability.RunsTotal++
	e.observability.RunsRunning++
	e.observability.mu.Unlock()

	// Execute steps in order (simulated)
	go e.executeSteps(ctx, p, run)

	return run, nil
}

func (e *Engine) executeSteps(ctx context.Context, p *Pipeline, run *PipelineRun) {
	var totalRead, totalWritten, totalFailed int64
	allSuccess := true

	for _, step := range p.Steps {
		select {
		case <-ctx.Done():
			e.mu.Lock()
			run.Status = PipelineStopped
			run.ErrorMsg = "cancelled"
			finTime := time.Now()
			run.FinishedAt = &finTime
			p.Status = PipelineStopped
			e.mu.Unlock()
			return
		default:
		}

		stepStart := time.Now()
		// Simulate step execution based on type
		rowsIn, rowsOut, rowsErr, err := e.simulateStep(step)

		stepEnd := time.Now()
		dur := stepEnd.Sub(stepStart)
		status := "succeeded"
		errMsg := ""
		if err != nil {
			status = "failed"
			errMsg = err.Error()
			allSuccess = false
		}

		throughput := float64(0)
		if dur.Seconds() > 0 {
			throughput = float64(rowsOut) / dur.Seconds()
		}

		result := StepResult{
			StepID:     step.ID,
			StepName:   step.Name,
			StepType:   step.Type,
			Status:     status,
			StartedAt:  stepStart,
			FinishedAt: &stepEnd,
			Duration:   dur.String(),
			RowsIn:     rowsIn,
			RowsOut:    rowsOut,
			RowsError:  rowsErr,
			ErrorMsg:   errMsg,
			Metrics: StepMetrics{
				BytesProcessed: rowsOut * 256,
				ThroughputRPS:  throughput,
				AvgRowSizeB:    256,
				MemoryUsedMB:   float64(rowsOut) * 0.001,
				CPUPercent:     math.Min(float64(rowsOut)*0.002, 95),
			},
		}

		e.mu.Lock()
		run.StepResults = append(run.StepResults, result)
		e.mu.Unlock()

		totalRead += rowsIn
		totalWritten += rowsOut
		totalFailed += rowsErr

		// Track step type stats
		e.observability.mu.Lock()
		e.observability.StepTypeStats[string(step.Type)]++
		if err != nil {
			errorType := "step_failure"
			if strings.Contains(errMsg, "timeout") {
				errorType = "timeout"
			} else if strings.Contains(errMsg, "connection") {
				errorType = "connection"
			}
			e.observability.ErrorsByType[errorType]++
		}
		// Track throughput
		if len(e.observability.ThroughputLog) > 200 {
			e.observability.ThroughputLog = e.observability.ThroughputLog[1:]
		}
		e.observability.ThroughputLog = append(e.observability.ThroughputLog, ThroughputPoint{
			Timestamp:  time.Now(),
			RowsPerSec: throughput,
			Pipeline:   p.Name,
		})
		e.observability.mu.Unlock()

		if !allSuccess {
			break
		}
	}

	finTime := time.Now()
	dur := finTime.Sub(run.StartedAt)

	e.mu.Lock()
	run.FinishedAt = &finTime
	run.Duration = dur.String()
	run.RowsRead = totalRead
	run.RowsWritten = totalWritten
	run.RowsFailed = totalFailed
	if allSuccess {
		run.Status = PipelineSuccess
		p.Status = PipelineSuccess
	} else {
		run.Status = PipelineFailed
		p.Status = PipelineFailed
	}
	e.mu.Unlock()

	e.observability.mu.Lock()
	e.observability.RunsRunning--
	if allSuccess {
		e.observability.RunsSuccess++
	} else {
		e.observability.RunsFailed++
	}
	e.observability.TotalRowsRead += totalRead
	e.observability.TotalRowsWrite += totalWritten
	n := float64(e.observability.RunsTotal)
	if n > 0 {
		e.observability.AvgDuration = (e.observability.AvgDuration*(n-1) + dur.Seconds()) / n
	}
	e.observability.LastUpdated = time.Now()
	e.observability.mu.Unlock()
}

func (e *Engine) simulateStep(step Step) (rowsIn, rowsOut, rowsErr int64, err error) {
	// Simulate different row counts based on step type
	switch step.Type {
	case StepExtract:
		rowsIn = 10000
		rowsOut = 10000
	case StepTransform:
		rowsIn = 10000
		rowsOut = 9800
		rowsErr = 200
	case StepFilter:
		rowsIn = 9800
		rowsOut = 7500
	case StepMap:
		rowsIn = 7500
		rowsOut = 7500
	case StepAggregate:
		rowsIn = 7500
		rowsOut = 150
	case StepValidate:
		rowsIn = 7500
		rowsOut = 7200
		rowsErr = 300
	case StepEnrich:
		rowsIn = 7200
		rowsOut = 7200
	case StepDedupe:
		rowsIn = 7200
		rowsOut = 6800
	case StepLoad:
		rowsIn = 6800
		rowsOut = 6800
	default:
		rowsIn = 1000
		rowsOut = 1000
	}
	// Simulate small delay
	time.Sleep(50 * time.Millisecond)
	return
}

// --- Run Queries ---

func (e *Engine) GetRun(runID string) (*PipelineRun, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	r, ok := e.runs[runID]
	return r, ok
}

func (e *Engine) ListRuns(pipelineID string) []*PipelineRun {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]*PipelineRun, 0)
	for _, r := range e.runs {
		if pipelineID == "" || r.PipelineID == pipelineID {
			result = append(result, r)
		}
	}
	return result
}

func (e *Engine) GetObservability() *ETLObservability {
	e.observability.mu.RLock()
	defer e.observability.mu.RUnlock()
	// Return a snapshot
	snap := *e.observability
	return &snap
}

// --- Seed ---

func (e *Engine) seedPipelines() {
	now := time.Now()
	pastRun := now.Add(-2 * time.Hour)

	// Pipeline 1: MySQL to PostgreSQL sync
	p1 := &Pipeline{
		ID:          "etl-mysql-pg-sync",
		Name:        "MySQL → PostgreSQL User Sync",
		Description: "Extracts user data from MySQL, transforms and loads into PostgreSQL",
		Steps: []Step{
			{ID: "s1", Name: "Extract from MySQL", Type: StepExtract, Connector: "mysql", Order: 1, Config: map[string]interface{}{"host": "localhost", "port": 3306, "database": "axiomnizam", "table": "users", "query": "SELECT * FROM users WHERE updated_at > ?"}},
			{ID: "s2", Name: "Validate Schema", Type: StepValidate, Connector: "", Order: 2, Config: map[string]interface{}{"rules": []string{"email_format", "not_null:name"}}},
			{ID: "s3", Name: "Transform Fields", Type: StepTransform, Connector: "", Order: 3, Config: map[string]interface{}{"operations": []string{"rename:fullname->name", "lowercase:email", "hash:password"}}},
			{ID: "s4", Name: "Deduplicate", Type: StepDedupe, Connector: "", Order: 4, Config: map[string]interface{}{"key": "email"}},
			{ID: "s5", Name: "Load to PostgreSQL", Type: StepLoad, Connector: "postgres", Order: 5, Config: map[string]interface{}{"host": "localhost", "port": 5432, "database": "axiomnizam_dw", "table": "users_synced"}},
		},
		Schedule:  "*/30 * * * *",
		Status:    PipelineSuccess,
		CreatedAt: now.Add(-72 * time.Hour),
		UpdatedAt: now,
		LastRunAt: &pastRun,
		RunCount:  145,
		Tags:      []string{"sync", "users", "critical"},
	}
	e.pipelines[p1.ID] = p1

	// Pipeline 2: API → Data Warehouse
	p2 := &Pipeline{
		ID:          "etl-api-warehouse",
		Name:        "REST API → Data Warehouse",
		Description: "Fetches order data from external API, enriches, and loads to warehouse",
		Steps: []Step{
			{ID: "s1", Name: "Fetch from API", Type: StepExtract, Connector: "api", Order: 1, Config: map[string]interface{}{"url": "https://api.example.com/orders", "method": "GET", "headers": map[string]string{"Authorization": "Bearer ****"}}},
			{ID: "s2", Name: "Filter Active", Type: StepFilter, Connector: "", Order: 2, Config: map[string]interface{}{"condition": "status == 'active'"}},
			{ID: "s3", Name: "Enrich with Customer Data", Type: StepEnrich, Connector: "mysql", Order: 3, Config: map[string]interface{}{"lookup_table": "customers", "join_key": "customer_id"}},
			{ID: "s4", Name: "Aggregate by Region", Type: StepAggregate, Connector: "", Order: 4, Config: map[string]interface{}{"group_by": "region", "functions": []string{"SUM(amount)", "COUNT(*)"}}},
			{ID: "s5", Name: "Load to Warehouse", Type: StepLoad, Connector: "postgres", Order: 5, Config: map[string]interface{}{"host": "localhost", "port": 5432, "database": "axiomnizam_dw", "table": "order_aggregates"}},
		},
		Schedule:  "0 */6 * * *",
		Status:    PipelineSuccess,
		CreatedAt: now.Add(-48 * time.Hour),
		UpdatedAt: now,
		LastRunAt: &pastRun,
		RunCount:  24,
		Tags:      []string{"orders", "warehouse", "aggregation"},
	}
	e.pipelines[p2.ID] = p2

	// Pipeline 3: CSV → MongoDB Analytics
	p3 := &Pipeline{
		ID:          "etl-csv-mongo-analytics",
		Name:        "CSV Import → MongoDB Analytics",
		Description: "Imports CSV reports, validates, transforms, and stores in MongoDB for analytics",
		Steps: []Step{
			{ID: "s1", Name: "Read CSV Files", Type: StepExtract, Connector: "csv", Order: 1, Config: map[string]interface{}{"path": "/data/reports/*.csv", "delimiter": ",", "has_header": true}},
			{ID: "s2", Name: "Validate Data", Type: StepValidate, Connector: "", Order: 2, Config: map[string]interface{}{"rules": []string{"numeric:amount", "date:created_at", "not_empty:product_id"}}},
			{ID: "s3", Name: "Map Fields", Type: StepMap, Connector: "", Order: 3, Config: map[string]interface{}{"mappings": map[string]string{"prod_id": "product_id", "amt": "amount", "dt": "created_at"}}},
			{ID: "s4", Name: "Load to MongoDB", Type: StepLoad, Connector: "mongodb", Order: 4, Config: map[string]interface{}{"uri": "mongodb://localhost:27017", "database": "analytics", "collection": "reports"}},
		},
		Schedule:  "0 2 * * *",
		Status:    PipelineCreated,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
		RunCount:  0,
		Tags:      []string{"csv", "import", "analytics"},
	}
	e.pipelines[p3.ID] = p3

	// Pipeline 4: Kafka Stream → Elasticsearch
	p4 := &Pipeline{
		ID:          "etl-kafka-elastic",
		Name:        "Kafka Events → Elasticsearch",
		Description: "Consumes events from Kafka, transforms, and indexes in Elasticsearch",
		Steps: []Step{
			{ID: "s1", Name: "Consume Kafka Topic", Type: StepExtract, Connector: "kafka", Order: 1, Config: map[string]interface{}{"brokers": "localhost:9092", "topic": "user-events", "group_id": "etl-consumer"}},
			{ID: "s2", Name: "Filter by Type", Type: StepFilter, Connector: "", Order: 2, Config: map[string]interface{}{"condition": "event_type IN ('login', 'purchase', 'signup')"}},
			{ID: "s3", Name: "Transform to ES Doc", Type: StepTransform, Connector: "", Order: 3, Config: map[string]interface{}{"operations": []string{"flatten_json", "add_timestamp", "geo_resolve:ip"}}},
			{ID: "s4", Name: "Index in Elasticsearch", Type: StepLoad, Connector: "elasticsearch", Order: 4, Config: map[string]interface{}{"url": "http://localhost:9200", "index": "user-events-2026"}},
		},
		Status:    PipelineRunning,
		CreatedAt: now.Add(-96 * time.Hour),
		UpdatedAt: now,
		LastRunAt: &now,
		RunCount:  1200,
		Tags:      []string{"streaming", "events", "search"},
	}
	e.pipelines[p4.ID] = p4

	// Seed some historical runs
	e.seedRuns(p1, now)

	e.observability.PipelinesTotal = len(e.pipelines)
}

func (e *Engine) seedRuns(p *Pipeline, now time.Time) {
	statuses := []PipelineStatus{PipelineSuccess, PipelineSuccess, PipelineSuccess, PipelineFailed, PipelineSuccess}
	for i, status := range statuses {
		startTime := now.Add(-time.Duration(5-i) * time.Hour)
		finTime := startTime.Add(45 * time.Second)
		dur := finTime.Sub(startTime)

		run := &PipelineRun{
			ID:          fmt.Sprintf("run-seed-%d", i+1),
			PipelineID:  p.ID,
			Status:      status,
			StartedAt:   startTime,
			FinishedAt:  &finTime,
			Duration:    dur.String(),
			Trigger:     "schedule",
			RowsRead:    10000,
			RowsWritten: 9500,
			StepResults: []StepResult{
				{StepID: "s1", StepName: "Extract from MySQL", StepType: StepExtract, Status: "succeeded", StartedAt: startTime, FinishedAt: &finTime, Duration: "12s", RowsIn: 10000, RowsOut: 10000, Metrics: StepMetrics{BytesProcessed: 2560000, ThroughputRPS: 833}},
				{StepID: "s2", StepName: "Validate Schema", StepType: StepValidate, Status: "succeeded", StartedAt: startTime, FinishedAt: &finTime, Duration: "5s", RowsIn: 10000, RowsOut: 9800, RowsError: 200, Metrics: StepMetrics{BytesProcessed: 2500000, ThroughputRPS: 1960}},
				{StepID: "s3", StepName: "Transform Fields", StepType: StepTransform, Status: "succeeded", StartedAt: startTime, FinishedAt: &finTime, Duration: "8s", RowsIn: 9800, RowsOut: 9800, Metrics: StepMetrics{BytesProcessed: 2500000, ThroughputRPS: 1225}},
				{StepID: "s4", StepName: "Deduplicate", StepType: StepDedupe, Status: "succeeded", StartedAt: startTime, FinishedAt: &finTime, Duration: "6s", RowsIn: 9800, RowsOut: 9500, Metrics: StepMetrics{BytesProcessed: 2430000, ThroughputRPS: 1583}},
				{StepID: "s5", StepName: "Load to PostgreSQL", StepType: StepLoad, Status: "succeeded", StartedAt: startTime, FinishedAt: &finTime, Duration: "14s", RowsIn: 9500, RowsOut: 9500, Metrics: StepMetrics{BytesProcessed: 2430000, ThroughputRPS: 678}},
			},
		}
		if status == PipelineFailed {
			run.ErrorMsg = "connection timeout: PostgreSQL host unreachable"
			run.RowsWritten = 0
			run.RowsFailed = 9500
			run.StepResults[4].Status = "failed"
			run.StepResults[4].ErrorMsg = "connection timeout"
		}
		e.runs[run.ID] = run
	}

	// Update observability
	e.observability.RunsTotal = len(statuses)
	e.observability.RunsSuccess = 4
	e.observability.RunsFailed = 1
	e.observability.TotalRowsRead = 50000
	e.observability.TotalRowsWrite = 38000
	e.observability.AvgDuration = 45.0
}
