package contracts

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
	Name     string `json:"name,omitempty"`
	Contract string `json:"contract,omitempty"`
}

// ContractListResponse is the API response for listing contracts.
type ContractListResponse struct {
	Contracts []*DataContractResource `json:"contracts"`
	Count     int                     `json:"count"`
}

// ViolationsResponse is the API response for contract violations.
type ViolationsResponse struct {
	Contract    string      `json:"contract"`
	Compliant   bool        `json:"compliant"`
	Violations  interface{} `json:"violations"`
	ValidatedAt interface{} `json:"validatedAt"`
}
