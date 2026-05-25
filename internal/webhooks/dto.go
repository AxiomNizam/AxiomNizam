package webhooks

// WebhookListResponse is the API response for listing webhooks.
type WebhookListResponse struct {
	Webhooks []*Webhook `json:"webhooks"`
	Count    int        `json:"count"`
}

// DeliveryLogListResponse is the API response for listing delivery logs.
type DeliveryLogListResponse struct {
	Deliveries []*WebhookDeliveryLog `json:"deliveries"`
}

// ResourceCreatedResponse is the accepted response for reconciler-authoritative creation.
type ResourceCreatedResponse struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MessageResponse is a generic action acknowledgment.
type MessageResponse struct {
	Message string `json:"message"`
}
