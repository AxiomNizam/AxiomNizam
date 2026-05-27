package cdc

import "time"

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// --- ETL Response DTOs ---

type ETLPipelineListResponse struct {
	Status    string      `json:"status"`
	Pipelines interface{} `json:"pipelines"`
	Total     int         `json:"total"`
}

type ETLPipelineResponse struct {
	Status   string      `json:"status"`
	Pipeline interface{} `json:"pipeline"`
}

type ETLRunResponse struct {
	Status string      `json:"status"`
	Run    interface{} `json:"run"`
}

type ETLRunListResponse struct {
	Status string      `json:"status"`
	Runs   interface{} `json:"runs"`
	Total  int         `json:"total"`
}

type ETLConnectorListResponse struct {
	Status     string         `json:"status"`
	Connectors interface{}    `json:"connectors"`
	Total      int            `json:"total"`
	Categories map[string]int `json:"categories"`
}

type ETLConnectorResponse struct {
	Status    string      `json:"status"`
	Connector interface{} `json:"connector"`
}

type ETLConnectorCatalogResponse struct {
	Status     string      `json:"status"`
	Connectors interface{} `json:"connectors"`
	ByCategory interface{} `json:"by_category"`
	Total      int         `json:"total"`
}

type ETLCapabilitiesResponse struct {
	Status       string      `json:"status"`
	Capabilities interface{} `json:"capabilities"`
}

type ETLBlueprintsResponse struct {
	Status    string      `json:"status"`
	Blueprints interface{} `json:"blueprints"`
}

type ETLObservabilityResponse struct {
	Status        string      `json:"status"`
	Observability interface{} `json:"observability"`
}

// --- CDC Response DTOs ---

type CDCPipelineListResponse struct {
	Status    string      `json:"status"`
	Pipelines interface{} `json:"pipelines"`
	Total     int         `json:"total"`
}

type CDCPipelineResponse struct {
	Status   string      `json:"status"`
	Pipeline interface{} `json:"pipeline"`
}

type CDCPipelineActionResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Pipeline interface{} `json:"pipeline"`
}

type CDCSourceTypesResponse struct {
	Status  string      `json:"status"`
	Sources interface{} `json:"sources"`
}

type CDCSinkTypesResponse struct {
	Status string      `json:"status"`
	Sinks  interface{} `json:"sinks"`
}

type CDCObservabilityResponse struct {
	Status        string      `json:"status"`
	Observability interface{} `json:"observability"`
}

// --- Platform Overview DTO ---

type PlatformOverviewResponse struct {
	Status  string           `json:"status"`
	Overview PlatformOverview `json:"overview"`
}

type PlatformOverview struct {
	ETL        ETLOverview        `json:"etl"`
	CDC        CDCOverview        `json:"cdc"`
	Connectors ConnectorSummary   `json:"connectors"`
}

type ETLOverview struct {
	PipelinesTotal    int         `json:"pipelines_total"`
	RunsTotal         int         `json:"runs_total"`
	RunsSuccess       int         `json:"runs_success"`
	RunsFailed        int         `json:"runs_failed"`
	RunsRunning       int         `json:"runs_running"`
	TotalRowsRead     int64       `json:"total_rows_read"`
	TotalRowsWritten  int64       `json:"total_rows_written"`
	AvgDurationSeconds float64    `json:"avg_duration_seconds"`
	Pipelines         interface{} `json:"pipelines"`
}

type CDCOverview struct {
	PipelinesTotal  int         `json:"pipelines_total"`
	PipelinesActive int         `json:"pipelines_active"`
	PipelinesPaused int         `json:"pipelines_paused"`
	PipelinesFailed int         `json:"pipelines_failed"`
	TotalEvents     int64       `json:"total_events"`
	EventsPerSecond float64     `json:"events_per_second"`
	TotalErrors     int64       `json:"total_errors"`
	ErrorRate       float64     `json:"error_rate"`
	AvgLagMs        float64     `json:"avg_lag_ms"`
	Pipelines       interface{} `json:"pipelines"`
}

type ConnectorSummary struct {
	ETLConnectors int `json:"etl_connectors"`
	CDCSources    int `json:"cdc_sources"`
	CDCSinks      int `json:"cdc_sinks"`
}

// --- CDC Stream DTOs ---

type CDCStreamChangeResponse struct {
	ID        string    `json:"id"`
	Table     string    `json:"table"`
	Operation string    `json:"operation"`
	Timestamp time.Time `json:"timestamp"`
}

type CDCStreamHistoryResponse struct {
	Table     string                   `json:"table"`
	Events    []map[string]interface{} `json:"events"`
	Count     int                      `json:"count"`
	Timestamp time.Time                `json:"timestamp"`
}

type CDCStreamCreateResponse struct {
	ID        string    `json:"id"`
	Table     string    `json:"table"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type CDCSubscribeResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type CDCStatsResponse struct {
	TotalEvents      int64     `json:"total_events"`
	TotalStreams      int       `json:"total_streams"`
	ActiveStreams     int       `json:"active_streams"`
	TotalWebhooks     int       `json:"total_webhooks"`
	InsertEvents      int64     `json:"insert_events"`
	UpdateEvents      int64     `json:"update_events"`
	DeleteEvents      int64     `json:"delete_events"`
	BufferUtilization float64   `json:"buffer_utilization"`
	Timestamp         time.Time `json:"timestamp"`
}
