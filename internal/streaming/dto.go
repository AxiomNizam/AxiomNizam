package streaming

// StreamCreatedResponse is the API response for creating a stream.
type StreamCreatedResponse struct {
	StreamID string `json:"streamId"`
}

// StreamListResponse is the API response for listing streams.
type StreamListResponse struct {
	Streams []*StreamSession `json:"streams"`
	Count   int              `json:"count"`
}

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}
