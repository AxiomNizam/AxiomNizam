package handlers

import (
	"fmt"
	"net/http"
	"time"

	"axiom/internal/audit"
	"axiom/internal/encryption"
	"axiom/internal/lineage"
	"axiom/internal/workflow"

	"github.com/gin-gonic/gin"
)

// Phase3Handlers manages all Phase 3 endpoints
type Phase3Handlers struct {
	encryptionMgr *encryption.FieldLevelEncryption
	lineageMgr    *lineage.DataLineageTracker
	auditMgr      *audit.AuditComplianceManager
	workflowMgr   *workflow.MultiVersionWorkflowManager
}

// NewPhase3Handlers creates Phase3Handlers
func NewPhase3Handlers(
	encMgr *encryption.FieldLevelEncryption,
	linMgr *lineage.DataLineageTracker,
	audMgr *audit.AuditComplianceManager,
	wfMgr *workflow.MultiVersionWorkflowManager,
) *Phase3Handlers {
	return &Phase3Handlers{
		encryptionMgr: encMgr,
		lineageMgr:    lineMgr,
		auditMgr:      audMgr,
		workflowMgr:   wfMgr,
	}
}

// ========== ENCRYPTION ENDPOINTS ==========

// RegisterEncryptionKey registers new encryption key
func (h *Phase3Handlers) RegisterEncryptionKey(c *gin.Context) {
	type Request struct {
		KeyID      string    `json:"key_id" binding:"required"`
		Key        string    `json:"key" binding:"required"`
		ExpiresAt  time.Time `json:"expires_at"`
		EncryptType string   `json:"encrypt_type"` // deterministic or searchable
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encKey := &encryption.EncryptionKey{
		KeyID:      req.KeyID,
		Key:        req.Key,
		CreatedAt:  time.Now(),
		ExpiresAt:  req.ExpiresAt,
		EncryptType: req.EncryptType,
	}

	if err := h.encryptionMgr.RegisterKey(encKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "encryption key registered",
		"key_id":  req.KeyID,
	})
}

// AddEncryptionPolicy adds field encryption policy
func (h *Phase3Handlers) AddEncryptionPolicy(c *gin.Context) {
	type Request struct {
		TableName   string `json:"table_name" binding:"required"`
		ColumnName  string `json:"column_name" binding:"required"`
		KeyID       string `json:"key_id" binding:"required"`
		Searchable  bool   `json:"searchable"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy := &encryption.FieldEncryptionPolicy{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
		KeyID:      req.KeyID,
		Searchable: req.Searchable,
	}

	if err := h.encryptionMgr.AddEncryptionPolicy(policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "encryption policy added",
		"table":   req.TableName,
		"column":  req.ColumnName,
	})
}

// EncryptFieldValue encrypts a field value
func (h *Phase3Handlers) EncryptFieldValue(c *gin.Context) {
	type Request struct {
		TableName  string `json:"table_name" binding:"required"`
		ColumnName string `json:"column_name" binding:"required"`
		Value      string `json:"value" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encrypted, err := h.encryptionMgr.EncryptField(req.TableName, req.ColumnName, []byte(req.Value))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"encrypted_data": fmt.Sprintf("%x", encrypted.CipherText),
		"iv":             fmt.Sprintf("%x", encrypted.IV),
		"key_id":         encrypted.KeyID,
	})
}

// DecryptFieldValue decrypts a field value
func (h *Phase3Handlers) DecryptFieldValue(c *gin.Context) {
	type Request struct {
		TableName    string `json:"table_name" binding:"required"`
		ColumnName   string `json:"column_name" binding:"required"`
		EncryptedData string `json:"encrypted_data" binding:"required"`
		IV           string `json:"iv" binding:"required"`
		KeyID        string `json:"key_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encField := &encryption.EncryptedField{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
		CipherText: []byte(req.EncryptedData),
		IV:         []byte(req.IV),
		KeyID:      req.KeyID,
	}

	decrypted, err := h.encryptionMgr.DecryptField(encField)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"decrypted_value": string(decrypted),
	})
}

// RotateEncryptionKey rotates encryption key
func (h *Phase3Handlers) RotateEncryptionKey(c *gin.Context) {
	keyID := c.Param("key_id")

	newKey := c.PostForm("new_key")
	if newKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_key required"})
		return
	}

	if err := h.encryptionMgr.RotateKey(keyID, newKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "encryption key rotated",
		"key_id":  keyID,
	})
}

// GetEncryptionMetrics gets encryption metrics
func (h *Phase3Handlers) GetEncryptionMetrics(c *gin.Context) {
	metrics := h.encryptionMgr.GetEncryptionMetrics()

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetEncryptionStatus gets encryption status
func (h *Phase3Handlers) GetEncryptionStatus(c *gin.Context) {
	status := h.encryptionMgr.GetEncryptionStatus()

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// ========== LINEAGE ENDPOINTS ==========

// RegisterDataNode registers a data node
func (h *Phase3Handlers) RegisterDataNode(c *gin.Context) {
	type Request struct {
		NodeID      string `json:"node_id" binding:"required"`
		NodeName    string `json:"node_name" binding:"required"`
		NodeType    string `json:"node_type" binding:"required"` // table, column, view
		Schema      string `json:"schema"`
		Description string `json:"description"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node := &lineage.DataLineageNode{
		NodeID:      req.NodeID,
		NodeName:    req.NodeName,
		NodeType:    req.NodeType,
		Schema:      req.Schema,
		Description: req.Description,
	}

	if err := h.lineageMgr.RegisterDataNode(node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "data node registered",
		"node_id": req.NodeID,
	})
}

// CreateLineageEdge creates data lineage edge
func (h *Phase3Handlers) CreateLineageEdge(c *gin.Context) {
	type Request struct {
		SourceNodeID string `json:"source_node_id" binding:"required"`
		TargetNodeID string `json:"target_node_id" binding:"required"`
		RelationType string `json:"relation_type" binding:"required"` // reads, writes, transforms
		Metadata     map[string]interface{} `json:"metadata"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	edge := &lineage.DataLineageEdge{
		SourceNodeID: req.SourceNodeID,
		TargetNodeID: req.TargetNodeID,
		RelationType: req.RelationType,
		CreatedAt:    time.Now(),
		Metadata:     req.Metadata,
	}

	if err := h.lineageMgr.CreateLineageEdge(edge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "lineage edge created",
		"source_node":  req.SourceNodeID,
		"target_node":  req.TargetNodeID,
	})
}

// GetUpstreamLineage gets upstream lineage
func (h *Phase3Handlers) GetUpstreamLineage(c *gin.Context) {
	nodeID := c.Query("node_id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id required"})
		return
	}

	lineage, err := h.lineageMgr.GetUpstreamLineage(nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upstream_lineage": lineage,
	})
}

// GetDownstreamLineage gets downstream lineage
func (h *Phase3Handlers) GetDownstreamLineage(c *gin.Context) {
	nodeID := c.Query("node_id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id required"})
		return
	}

	lineage, err := h.lineageMgr.GetDownstreamLineage(nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"downstream_lineage": lineage,
	})
}

// AnalyzeImpact analyzes change impact
func (h *Phase3Handlers) AnalyzeImpact(c *gin.Context) {
	type Request struct {
		SourceNodeID string `json:"source_node_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	impact, err := h.lineageMgr.AnalyzeImpact(req.SourceNodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"impact_analysis": impact,
	})
}

// GetLineageGraph gets lineage graph
func (h *Phase3Handlers) GetLineageGraph(c *gin.Context) {
	graph := h.lineageMgr.GetLineageGraph()

	c.JSON(http.StatusOK, gin.H{
		"lineage_graph": graph,
	})
}

// GetLineageStats gets lineage statistics
func (h *Phase3Handlers) GetLineageStats(c *gin.Context) {
	stats := h.lineageMgr.GetLineageStats()

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// ========== AUDIT & COMPLIANCE ENDPOINTS ==========

// LogAuditEvent logs an audit event
func (h *Phase3Handlers) LogAuditEvent(c *gin.Context) {
	type Request struct {
		UserID       string                 `json:"user_id" binding:"required"`
		Action       string                 `json:"action" binding:"required"`
		ResourceType string                 `json:"resource_type"`
		ResourceID   string                 `json:"resource_id"`
		IPAddress    string                 `json:"ip_address"`
		Changes      map[string]interface{} `json:"changes"`
		Status       string                 `json:"status"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auditLog := &audit.AuditLog{
		UserID:       req.UserID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		IPAddress:    req.IPAddress,
		Timestamp:    time.Now(),
		Changes:      req.Changes,
		Status:       req.Status,
	}

	if err := h.auditMgr.LogAuditEvent(auditLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "audit event logged",
	})
}

// RegisterComplianceRule registers compliance rule
func (h *Phase3Handlers) RegisterComplianceRule(c *gin.Context) {
	type Request struct {
		RuleID       string `json:"rule_id" binding:"required"`
		RuleName     string `json:"rule_name" binding:"required"`
		Framework    string `json:"framework" binding:"required"` // GDPR, HIPAA, SOC2, PCI-DSS
		Description  string `json:"description"`
		Severity     string `json:"severity"` // low, medium, high
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &audit.ComplianceRule{
		RuleID:      req.RuleID,
		RuleName:    req.RuleName,
		Framework:   req.Framework,
		Description: req.Description,
		Severity:    req.Severity,
		CreatedAt:   time.Now(),
	}

	if err := h.auditMgr.RegisterComplianceRule(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "compliance rule registered",
		"rule_id": req.RuleID,
	})
}

// GenerateComplianceReport generates compliance report
func (h *Phase3Handlers) GenerateComplianceReport(c *gin.Context) {
	framework := c.Query("framework")
	if framework == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "framework required"})
		return
	}

	report, err := h.auditMgr.GenerateComplianceReport(framework)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"compliance_report": report,
	})
}

// GetComplianceStatus gets compliance status
func (h *Phase3Handlers) GetComplianceStatus(c *gin.Context) {
	status := h.auditMgr.GetComplianceStatus()

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// SearchAuditLogs searches audit logs
func (h *Phase3Handlers) SearchAuditLogs(c *gin.Context) {
	userID := c.Query("user_id")
	resourceType := c.Query("resource_type")

	logs := h.auditMgr.SearchAuditLogs(userID, resourceType)

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": logs,
	})
}

// RecordViolation records compliance violation
func (h *Phase3Handlers) RecordViolation(c *gin.Context) {
	type Request struct {
		RuleID      string `json:"rule_id" binding:"required"`
		Description string `json:"description"`
		Severity    string `json:"severity"` // low, medium, high, critical
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	violation := &audit.ComplianceViolation{
		RuleID:      req.RuleID,
		Description: req.Description,
		Severity:    req.Severity,
		DetectedAt:  time.Now(),
	}

	if err := h.auditMgr.RecordViolation(violation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "violation recorded",
	})
}

// ========== WORKFLOW ENDPOINTS ==========

// CreateWorkflow creates new workflow
func (h *Phase3Handlers) CreateWorkflow(c *gin.Context) {
	type Request struct {
		Name        string                   `json:"name" binding:"required"`
		Description string                   `json:"description"`
		Steps       []map[string]interface{} `json:"steps"`
		CreatedBy   string                   `json:"created_by"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wfDef := &workflow.WorkflowDefinition{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
		Status:      "draft",
	}

	version, err := h.workflowMgr.CreateWorkflow(wfDef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "workflow created",
		"workflow_id": wfDef.ID,
		"version":     version.Version,
	})
}

// PublishWorkflowVersion publishes workflow version
func (h *Phase3Handlers) PublishWorkflowVersion(c *gin.Context) {
	type Request struct {
		WorkflowID  string                   `json:"workflow_id" binding:"required"`
		ChangeSummary string                 `json:"change_summary"`
		Steps       []map[string]interface{} `json:"steps"`
		CreatedBy   string                   `json:"created_by"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newDef := &workflow.WorkflowDefinition{
		CreatedBy: req.CreatedBy,
		Status:    "published",
	}

	version, err := h.workflowMgr.PublishWorkflowVersion(req.WorkflowID, newDef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "workflow version published",
		"version": version.Version,
	})
}

// StartWorkflowInstance starts workflow execution
func (h *Phase3Handlers) StartWorkflowInstance(c *gin.Context) {
	type Request struct {
		WorkflowID  string                 `json:"workflow_id" binding:"required"`
		Version     string                 `json:"version"`
		ContextData map[string]interface{} `json:"context_data"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	instance, err := h.workflowMgr.StartWorkflowInstance(req.WorkflowID, req.Version, req.ContextData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "workflow instance started",
		"instance_id": instance.ID,
		"status":      instance.Status,
	})
}

// GetWorkflowMetrics gets workflow metrics
func (h *Phase3Handlers) GetWorkflowMetrics(c *gin.Context) {
	workflowID := c.Query("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id required"})
		return
	}

	metrics := h.workflowMgr.GetWorkflowMetrics(workflowID)

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetWorkflowStatus gets workflow status
func (h *Phase3Handlers) GetWorkflowStatus(c *gin.Context) {
	status := h.workflowMgr.GetWorkflowStatus()

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// GetInstanceHistory gets workflow instance history
func (h *Phase3Handlers) GetInstanceHistory(c *gin.Context) {
	workflowID := c.Query("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id required"})
		return
	}

	limit := 100
	history := h.workflowMgr.GetInstanceHistory(workflowID, limit)

	c.JSON(http.StatusOK, gin.H{
		"instance_history": history,
	})
}
