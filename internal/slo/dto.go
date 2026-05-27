package slo

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// SLOListResponse is the API response for listing SLOs.
type SLOListResponse struct {
	SLOs  []*SLOResource `json:"slos"`
	Count int            `json:"count"`
}

// BudgetResponse is the API response for SLO budget status.
type BudgetResponse struct {
	Name           string  `json:"name"`
	Target         float64 `json:"target"`
	CurrentSLI     float64 `json:"currentSli"`
	ErrorBudget    float64 `json:"errorBudget"`
	BudgetConsumed float64 `json:"budgetConsumed"`
	BurnRate       float64 `json:"burnRate"`
	IsBreaching    bool    `json:"isBreaching"`
	TimeToExhaust  string  `json:"timeToExhaust"`
	GoodEvents     int64   `json:"goodEvents"`
	TotalEvents    int64   `json:"totalEvents"`
	Window         string  `json:"window"`
}

// AllStatusResponse is the API response for all SLO statuses.
type AllStatusResponse struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	AtRisk    int `json:"atRisk"`
	Breaching int `json:"breaching"`
}
