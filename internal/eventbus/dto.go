package eventbus

// EventListResponse is the API response for listing events.
type EventListResponse struct {
	Events []*EventBusEvent `json:"events"`
	Count  int              `json:"count"`
}

// TopicListResponse is the API response for listing topics.
type TopicListResponse struct {
	Topics []*EventTopic `json:"topics"`
	Count  int           `json:"count"`
}

// ResourceCreatedResponse is the accepted response for reconciler-authoritative creation.
type ResourceCreatedResponse struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// SubscriptionListResponse is the API response for listing subscriptions.
type SubscriptionListResponse struct {
	Subscriptions []*EventSubscription `json:"subscriptions"`
	Count         int                  `json:"count"`
}

// DLQListResponse is the API response for listing DLQ events.
type DLQListResponse struct {
	Events []*DLQEvent `json:"events"`
	Count  int         `json:"count"`
}

// AckResponse is the API response for acknowledging an event.
type AckResponse struct {
	Message string        `json:"message"`
	Event   *EventBusEvent `json:"event"`
}

// ReplayResponse is the API response for replaying a DLQ event.
type ReplayResponse struct {
	Message string               `json:"message"`
	Replay  *EventPublishResponse `json:"replay"`
}

// MessageResponse is a generic error response.
type MessageResponse struct {
	Error string `json:"error"`
}
