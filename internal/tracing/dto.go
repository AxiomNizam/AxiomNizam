package tracing

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	TraceID string `json:"traceId,omitempty"`
}

// IngestionAuditListResponse is the API response for listing ingestion audits.
type IngestionAuditListResponse struct {
	Logs  []*TraceIngestionAuditLog `json:"logs"`
	Count int                       `json:"count"`
}

// TraceSearchListResponse is the API response for trace search results.
type TraceSearchListResponse struct {
	Traces []*TraceSearchResult `json:"traces"`
	Count  int                  `json:"count"`
}

// ServiceListResponse is the API response for listing services.
type ServiceListResponse struct {
	Services []*ServiceInfo `json:"services"`
	Count    int            `json:"count"`
}
