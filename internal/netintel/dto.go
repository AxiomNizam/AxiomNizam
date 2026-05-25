package netintel

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// --- NetIntel Response DTOs ---

type SummaryResponse struct {
	Status  string      `json:"status"`
	Summary interface{} `json:"summary"`
}

type ObservabilityResponse struct {
	Status        string      `json:"status"`
	Observability interface{} `json:"observability"`
}

type LogTypesResponse struct {
	Status   string      `json:"status"`
	LogTypes interface{} `json:"log_types"`
	Total    int         `json:"total"`
}

type ParserListResponse struct {
	Status  string      `json:"status"`
	Parsers interface{} `json:"parsers"`
	Total   int         `json:"total"`
}

type ParserResponse struct {
	Status string      `json:"status"`
	Parser interface{} `json:"parser"`
}

type EntryListResponse struct {
	Status  string      `json:"status"`
	Entries interface{} `json:"entries"`
	Total   int         `json:"total"`
}

type StatsResponse struct {
	Status string      `json:"status"`
	Stats  interface{} `json:"stats"`
}

type TopologyResponse struct {
	Status   string      `json:"status"`
	Topology interface{} `json:"topology"`
}

type TopologyNodeResponse struct {
	Status string      `json:"status"`
	Node   interface{} `json:"node"`
}

type HeatmapResponse struct {
	Status  string      `json:"status"`
	Heatmap interface{} `json:"heatmap"`
}

type TrendsResponse struct {
	Status string      `json:"status"`
	Metric string      `json:"metric"`
	Hours  int         `json:"hours"`
	Trend  interface{} `json:"trend"`
}

type PredictionsResponse struct {
	Status      string      `json:"status"`
	Predictions interface{} `json:"predictions"`
	Total       int         `json:"total"`
}

type TrackListResponse struct {
	Status string      `json:"status"`
	Tracks interface{} `json:"tracks"`
	Total  int         `json:"total"`
}

type TrackResponse struct {
	Status string      `json:"status"`
	Track  interface{} `json:"track"`
}

type AnomalyListResponse struct {
	Status    string      `json:"status"`
	Anomalies interface{} `json:"anomalies"`
	Total     int         `json:"total"`
}

type AlertListResponse struct {
	Status string      `json:"status"`
	Alerts interface{} `json:"alerts"`
	Total  int         `json:"total"`
}

type ForecastListResponse struct {
	Status    string      `json:"status"`
	Forecasts interface{} `json:"forecasts"`
}

type ForecastResponse struct {
	Status   string      `json:"status"`
	Forecast interface{} `json:"forecast"`
}
