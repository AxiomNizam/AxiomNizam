package datasourceresource

import "time"

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// --- DataSource Response DTOs ---

type DataSourceCreatedResponse struct {
	Status      string             `json:"status"`
	Message     string             `json:"message"`
	Datasource  DataSourceResource `json:"datasource"`
}

type DataSourceListResponse struct {
	Status      string               `json:"status"`
	Datasources []*DataSourceResource `json:"datasources"`
	Total       int                  `json:"total"`
}

type DataSourceGetResponse struct {
	Status     string             `json:"status"`
	Datasource *DataSourceResource `json:"datasource"`
}

type DataSourceUpdatedResponse struct {
	Status      string             `json:"status"`
	Message     string             `json:"message"`
	Datasource  DataSourceResource `json:"datasource"`
}

type DataSourceTestResponse struct {
	Status   string    `json:"status"`
	Message  string    `json:"message"`
	Driver   string    `json:"driver"`
	Endpoint string    `json:"endpoint"`
	TestedAt time.Time `json:"tested_at"`
}
