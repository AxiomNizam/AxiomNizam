package federation

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// --- Federation Response DTOs ---

type VirtualTableListResponse struct {
	VirtualTables interface{} `json:"virtualTables"`
	Count         int         `json:"count"`
}

type QueryErrorResponse struct {
	Error   string `json:"error"`
	QueryID string `json:"queryId"`
}

type QueryExecResponse struct {
	QueryID  string      `json:"queryId"`
	Status   string      `json:"status"`
	Plan     interface{} `json:"plan"`
	Duration string      `json:"duration"`
	Sources  interface{} `json:"sources"`
}

type ExplainResponse struct {
	Plan    interface{} `json:"plan"`
	Sources interface{} `json:"sources"`
}

type QueryListResponse struct {
	Queries interface{} `json:"queries"`
	Count   int         `json:"count"`
}
