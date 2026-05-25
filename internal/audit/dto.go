package audit

// LogActionResponse is the API response for creating an audit log.
type LogActionResponse struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
}

// LogListResponse is the API response for listing audit logs.
type LogListResponse struct {
	Logs  []AuditLog `json:"logs"`
	Count int        `json:"count"`
}

// ResourceCreatedResponse is the accepted response for reconciler-authoritative creation.
type ResourceCreatedResponse struct {
	Status   string `json:"status"`
	TenantID string `json:"tenantId"`
	Message  string `json:"message"`
}

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
