package streamanalytics

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// ListJobsResponse is the typed response for ListJobs.
type ListJobsResponse struct {
	StreamJobs []*StreamJobResource `json:"streamJobs"`
	Count      int                  `json:"count"`
}
