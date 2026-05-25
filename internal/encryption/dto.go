package encryption

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

// --- Encryption Handler Response DTOs ---

type KeyListResponse struct {
	Keys  interface{} `json:"keys"`
	Count int         `json:"count"`
}

type PolicyListResponse struct {
	Policies interface{} `json:"policies"`
	Count    int         `json:"count"`
}

// --- Phase3 Response DTOs ---

type AckResponse struct {
	Message string `json:"message"`
}

type KeyRegisteredResponse struct {
	Message string `json:"message"`
	KeyID   string `json:"key_id"`
}

type PolicyAddedResponse struct {
	Message string `json:"message"`
	Table   string `json:"table"`
	Column  string `json:"column"`
}

type EncryptedFieldResponse struct {
	EncryptedData string `json:"encrypted_data"`
	IV            string `json:"iv"`
	KeyID         string `json:"key_id"`
}

type DecryptedFieldResponse struct {
	DecryptedValue interface{} `json:"decrypted_value"`
}

type KeyRotatedResponse struct {
	Message string `json:"message"`
	KeyID   string `json:"key_id"`
}

type MetricsResponse struct {
	Metrics interface{} `json:"metrics"`
}

type StatusResponse struct {
	Status interface{} `json:"status"`
}

type NodeRegisteredResponse struct {
	Message string `json:"message"`
	NodeID  string `json:"node_id"`
}

type LineageEdgeCreatedResponse struct {
	Message  string `json:"message"`
	EdgeID   string `json:"edge_id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

type UpstreamLineageResponse struct {
	UpstreamLineage interface{} `json:"upstream_lineage"`
}

type DownstreamLineageResponse struct {
	DownstreamLineage interface{} `json:"downstream_lineage"`
}

type ImpactAnalysisResponse struct {
	ImpactAnalysis interface{} `json:"impact_analysis"`
}

type LineageGraphResponse struct {
	LineageGraph interface{} `json:"lineage_graph"`
}

type StatisticsResponse struct {
	Statistics interface{} `json:"statistics"`
}

type ComplianceRuleRegisteredResponse struct {
	Message string `json:"message"`
	RuleID  string `json:"rule_id"`
}

type ComplianceReportResponse struct {
	ComplianceReport interface{} `json:"compliance_report"`
}

type AuditLogsResponse struct {
	AuditLogs interface{} `json:"audit_logs"`
}

type WorkflowCreatedResponse struct {
	Message    string `json:"message"`
	WorkflowID string `json:"workflow_id"`
	Version    string `json:"version"`
}

type WorkflowVersionPublishedResponse struct {
	Message string `json:"message"`
	Version string `json:"version"`
}

type WorkflowInstanceStartedResponse struct {
	Message    string      `json:"message"`
	InstanceID string      `json:"instance_id"`
	Status     interface{} `json:"status"`
}

type InstanceHistoryResponse struct {
	InstanceHistory interface{} `json:"instance_history"`
}
