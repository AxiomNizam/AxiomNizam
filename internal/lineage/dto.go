package lineage

// NodeListResponse is the API response for listing nodes.
type NodeListResponse struct {
	Nodes []*LineageNode `json:"nodes"`
	Count int            `json:"count"`
}

// PathListResponse is the API response for listing lineage paths.
type PathListResponse struct {
	Paths []*LineagePath `json:"paths"`
	Count int            `json:"count"`
}

// MessageResponse is a generic error response.
type MessageResponse struct {
	Error string `json:"error"`
}
