package conductor

// ProducerListResponse is the API response for listing producers.
type ProducerListResponse struct {
	Producers []*Producer `json:"producers"`
}

// ConsumerListResponse is the API response for listing consumers.
type ConsumerListResponse struct {
	Consumers []*Consumer `json:"consumers"`
}

// MessageListResponse is the API response for listing messages.
type MessageListResponse struct {
	Messages []*Message `json:"messages"`
}

// DLQListResponse is the API response for listing DLQ entries.
type DLQListResponse struct {
	DLQ []*DLQEntry `json:"dlq"`
}

// ReplayResponse is the API response for replaying a DLQ message.
type ReplayResponse struct {
	Message string   `json:"message"`
	Msg     *Message `json:"msg"`
}

// MessageResponse is a generic action acknowledgment.
type MessageResponse struct {
	Message string `json:"message"`
}
