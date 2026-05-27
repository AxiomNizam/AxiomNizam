package featurestore

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Group   string `json:"group,omitempty"`
}

// FeatureGroupListResponse is the response for listing feature groups.
type FeatureGroupListResponse struct {
	FeatureGroups []*FeatureGroupResource `json:"featureGroups"`
	Count         int                     `json:"count"`
}

// OnlineServingResponse is the response for online feature serving.
type OnlineServingResponse struct {
	FeatureGroup string `json:"featureGroup"`
	Entities     int    `json:"entities"`
	Features     int    `json:"features"`
	Freshness    string `json:"freshness"`
	Message      string `json:"message"`
}
