package versioning

import "time"

// VersionListResponse is the API response for listing API versions.
type VersionListResponse struct {
	Versions       []string  `json:"versions"`
	CurrentVersion string    `json:"current_version"`
	DefaultVersion string    `json:"default_version"`
	Count          int       `json:"count"`
	Timestamp      time.Time `json:"timestamp"`
}

// VersionInfoResponse is the API response for version info.
type VersionInfoResponse struct {
	Version             string    `json:"version"`
	Title               string    `json:"title"`
	Status              string    `json:"status"`
	EndpointCount       int       `json:"endpoint_count"`
	DeprecationWarnings []string  `json:"deprecation_warnings"`
	Timestamp           time.Time `json:"timestamp"`
}

// DeprecationWarningsResponse is the API response for deprecation warnings.
type DeprecationWarningsResponse struct {
	Version   string    `json:"version"`
	Warnings  []string  `json:"warnings"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// MigrationGuideResponse is the API response for migration guide.
type MigrationGuideResponse struct {
	FromVersion string                   `json:"from_version"`
	ToVersion   string                   `json:"to_version"`
	Steps       []map[string]interface{} `json:"steps"`
	Timestamp   time.Time                `json:"timestamp"`
}

// VersionUsageResponse is the API response for version usage stats.
type VersionUsageResponse struct {
	Usage         map[string]int `json:"usage"`
	TotalRequests int            `json:"total_requests"`
	Timestamp     time.Time      `json:"timestamp"`
}

// TransformResponse is the API response for request transformation.
type TransformResponse struct {
	Original    interface{} `json:"original"`
	FromVersion string      `json:"from_version"`
	ToVersion   string      `json:"to_version"`
	Transformed interface{} `json:"transformed"`
	Timestamp   time.Time   `json:"timestamp"`
}

// MessageResponse is a generic error response.
type MessageResponse struct {
	Error string `json:"error"`
}
