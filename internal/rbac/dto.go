package rbac

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// --- RBAC Response DTOs ---

type RoleCreatedResponse struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RoleListResponse struct {
	Roles interface{} `json:"roles"`
	Count int         `json:"count"`
}

type BindingListResponse struct {
	Bindings interface{} `json:"bindings"`
	Count    int         `json:"count"`
}

type PermissionListResponse struct {
	Permissions interface{} `json:"permissions"`
	Count       int         `json:"count"`
}

type AccessRequestListResponse struct {
	AccessRequests interface{} `json:"accessRequests"`
	Count          int         `json:"count"`
}
