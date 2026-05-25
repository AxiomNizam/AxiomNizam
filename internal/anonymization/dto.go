package anonymization

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Policy  string `json:"policy,omitempty"`
}

// ListPoliciesResponse is the typed response for ListPolicies.
type ListPoliciesResponse struct {
	Policies []*AnonymizationPolicyResource `json:"policies"`
	Count    int                            `json:"count"`
}
