package catalog

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// --- Catalog Response DTOs ---

type AssetListResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
}

type CatalogSearchResponse struct {
	Query   string      `json:"query"`
	Results interface{} `json:"results"`
	Total   int         `json:"total"`
}

type ScanResultResponse struct {
	DataSourceRef  string      `json:"dataSourceRef"`
	AssetsFound    int         `json:"assetsFound"`
	AssetsCreated  int         `json:"assetsCreated"`
	AssetsUpdated  int         `json:"assetsUpdated"`
	Duration       string      `json:"duration"`
	Errors         interface{} `json:"errors,omitempty"`
	PartialFailure bool        `json:"partialFailure,omitempty"`
	ScanErrors     interface{} `json:"scanErrors,omitempty"`
}

type DomainListResponse struct {
	Domains interface{} `json:"domains"`
	Total   int         `json:"total"`
}

type CollectionListResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
}
