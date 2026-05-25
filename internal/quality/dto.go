package quality

import "time"

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// ValidateDataResponse is the response for data validation.
type ValidateDataResponse struct {
	Status    string                 `json:"status"`
	Table     string                 `json:"table"`
	Record    map[string]interface{} `json:"record"`
	Timestamp time.Time              `json:"timestamp"`
}

// DetectAnomaliesResponse is the response for anomaly detection.
type DetectAnomaliesResponse struct {
	Table     string                   `json:"table"`
	Field     string                   `json:"field"`
	Anomalies []map[string]interface{} `json:"anomalies"`
	Timestamp time.Time                `json:"timestamp"`
}

// QualityMetricsResponse is the response for quality metrics.
type QualityMetricsResponse struct {
	TotalChecks    int       `json:"total_checks"`
	PassedChecks   int       `json:"passed_checks"`
	FailedChecks   int       `json:"failed_checks"`
	AnomaliesFound int       `json:"anomalies_found"`
	QualityScore   float64   `json:"quality_score"`
	Timestamp      time.Time `json:"timestamp"`
}
