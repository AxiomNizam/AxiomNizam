package notification

// NotifResponse is the API response for sending a notification.
type NotifResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Title   string `json:"title,omitempty"`
}

// HealthNotifResponse is the API response for health notification.
type HealthNotifResponse struct {
	Status     string           `json:"status"`
	Message    string           `json:"message"`
	HealthData HealthStatusData `json:"health_data,omitempty"`
}

// StatusNotifResponse is the API response for status notification.
type StatusNotifResponse struct {
	Status     string           `json:"status"`
	Message    string           `json:"message"`
	StatusData HealthStatusData `json:"status_data,omitempty"`
}

// ServiceStatusResponse is the API response for notification service status.
type ServiceStatusResponse struct {
	Status            string   `json:"status"`
	WebhookURL        string   `json:"webhook_url"`
	NotificationTypes []string `json:"notification_types"`
	SupportedTypes    []string `json:"supported_types"`
}
