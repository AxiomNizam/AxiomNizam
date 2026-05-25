package schemaregistry

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// SubjectNotFoundResponse is returned when a subject cannot be found.
type SubjectNotFoundResponse struct {
	Error   string `json:"error"`
	Subject string `json:"subject"`
}

// SchemaDetailResponse is the full schema version detail including ID and references.
type SchemaDetailResponse struct {
	Subject    string            `json:"subject"`
	Version    int               `json:"version"`
	ID         int64             `json:"id"`
	SchemaType string            `json:"schemaType"`
	Schema     string            `json:"schema"`
	References []SchemaReference `json:"references"`
}

// DetailErrorResponse is an error response that includes a detail field.
type DetailErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// SchemaRegisteredResponse is returned when a schema is accepted for registration.
type SchemaRegisteredResponse struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// VersionDeletedResponse is returned when a schema version is soft-deleted.
type VersionDeletedResponse struct {
	Version int  `json:"version"`
	Deleted bool `json:"deleted"`
}

// SchemaByIDResponse is returned when looking up a schema by its global ID.
type SchemaByIDResponse struct {
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType"`
	Subject    string `json:"subject"`
	Version    int    `json:"version"`
}

// SchemaIDNotFoundResponse is returned when no schema matches the given ID.
type SchemaIDNotFoundResponse struct {
	Error string `json:"error"`
	ID    int64  `json:"id"`
}

// CompatibilityResponse is returned when getting or setting compatibility mode.
type CompatibilityResponse struct {
	Compatibility string `json:"compatibility"`
}

// CompatibilityCheckResponse is returned from a compatibility check operation.
type CompatibilityCheckResponse struct {
	IsCompatible bool     `json:"is_compatible"`
	Errors       []string `json:"errors"`
}
