package netintel

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
