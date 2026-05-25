package alerting

import "time"

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// RuleListResponse is the API response for listing alert rules.
type RuleListResponse struct {
	Rules []*AlertRuleResource `json:"rules"`
	Count int                  `json:"count"`
}

// IncidentListResponse is the API response for listing incidents.
type IncidentListResponse struct {
	Incidents []*AlertIncidentResource `json:"incidents"`
	Count     int                      `json:"count"`
}

// SilenceResponse is the API response for silencing a rule.
type SilenceResponse struct {
	Silenced     bool       `json:"silenced"`
	SilenceUntil *time.Time `json:"silenceUntil,omitempty"`
	Reason       string     `json:"reason,omitempty"`
}

// UnsilenceResponse is the API response for unsilencing a rule.
type UnsilenceResponse struct {
	Silenced bool `json:"silenced"`
}

// ActionResponse is the API response for acknowledge/resolve actions.
type ActionResponse struct {
	Action   bool   `json:"action"`
	Incident string `json:"incident"`
}

// ChannelListResponse is the API response for listing channels.
type ChannelListResponse struct {
	Channels []*NotificationChannelResource `json:"channels"`
	Count    int                            `json:"count"`
}

// TestChannelResponse is the API response for testing a channel.
type TestChannelResponse struct {
	Success bool   `json:"tested"`
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// SummaryResponse is the API response for alerting summary.
type SummaryResponse struct {
	TotalRules            int `json:"totalRules"`
	SilencedRules         int `json:"silencedRules"`
	FiringIncidents       int `json:"firingIncidents"`
	AcknowledgedIncidents int `json:"acknowledgedIncidents"`
	ResolvedIncidents     int `json:"resolvedIncidents"`
	TotalIncidents        int `json:"totalIncidents"`
}
