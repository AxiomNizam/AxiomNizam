package data

import (
	"fmt"
	"time"
)

// DataPolicy defines data governance policies
type DataPolicy struct {
	ID                   string
	Name                 string
	Type                 string
	Version              string
	Enabled              bool
	DataClassifications  []DataClassification
	DataHandlingRules    []DataHandlingRule
	RetentionPolicy      RetentionPolicy
	AuditingRequirements AuditingRequirements
	EncryptionPolicy     EncryptionPolicy
	Description          string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// DataClassification defines data sensitivity levels
type DataClassification struct {
	Level       string // "public", "internal", "confidential", "restricted"
	Description string
	Tags        []string
	Handlers    []string // roles allowed to handle this data
}

// DataHandlingRule defines how data should be handled
type DataHandlingRule struct {
	DataLevel      string
	AllowedActions []string // "read", "write", "delete", "export", "share"
	RequiredApprovals int
	NotificationRequired bool
	LoggingRequired      bool
	MaskingRequired      bool
}

// RetentionPolicy defines data retention rules
type RetentionPolicy struct {
	MinRetentionDays  int
	MaxRetentionDays  int
	AutoDeleteAfter   int
	ArchiveAfter      int
	Purpose           string
	ApplicableTo      []string
}

// AuditingRequirements defines audit requirements
type AuditingRequirements struct {
	Required              bool
	LogAllAccess          bool
	LogModifications      bool
	LogDeletions          bool
	RetentionDays         int
	SendAlertsOn          []string // "access", "modification", "deletion", "export"
}

// EncryptionPolicy defines encryption requirements
type EncryptionPolicy struct {
	AtRestRequired       bool
	InTransitRequired    bool
	Algorithm            string // "AES-256", "RSA", etc
	KeyManagement        string // "system", "customer"
	ComplianceStandard   string // "HIPAA", "GDPR", "SOC2"
}

// GetID returns policy ID
func (dp *DataPolicy) GetID() string {
	return dp.ID
}

// GetName returns policy name
func (dp *DataPolicy) GetName() string {
	return dp.Name
}

// GetType returns policy type
func (dp *DataPolicy) GetType() string {
	return dp.Type
}

// GetVersion returns version
func (dp *DataPolicy) GetVersion() string {
	return dp.Version
}

// GetEnabled returns if enabled
func (dp *DataPolicy) GetEnabled() bool {
	return dp.Enabled
}

// Validate validates the policy
func (dp *DataPolicy) Validate() error {
	if dp.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if dp.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if len(dp.DataClassifications) == 0 {
		return fmt.Errorf("at least one data classification must be defined")
	}
	return nil
}

// DataGovernanceEngine manages data governance
type DataGovernanceEngine struct {
	policies        []*DataPolicy
	dataInventory   map[string]DataAsset
	complianceChecks map[string]ComplianceCheck
}

// DataAsset represents a data asset
type DataAsset struct {
	ID               string
	Name             string
	Classification   string
	Owner            string
	Sensitivity      string
	CreatedAt        time.Time
	LastAccessedAt   time.Time
	AccessLog        []AccessRecord
	Handlers         []string
	Metadata         map[string]interface{}
}

// AccessRecord records data access
type AccessRecord struct {
	Timestamp time.Time
	UserID    string
	Action    string
	IPAddress string
	Status    string
}

// ComplianceCheck tracks compliance status
type ComplianceCheck struct {
	ID            string
	CheckName     string
	DataAssetID   string
	Status        string // "pass", "fail", "warning"
	LastCheckedAt time.Time
	Issues        []string
	Remediation   string
}

// NewDataGovernanceEngine creates a new data governance engine
func NewDataGovernanceEngine() *DataGovernanceEngine {
	return &DataGovernanceEngine{
		policies:         make([]*DataPolicy, 0),
		dataInventory:    make(map[string]DataAsset),
		complianceChecks: make(map[string]ComplianceCheck),
	}
}

// RegisterPolicy registers a data policy
func (dge *DataGovernanceEngine) RegisterPolicy(policy *DataPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	dge.policies = append(dge.policies, policy)
	return nil
}

// RegisterDataAsset registers a data asset
func (dge *DataGovernanceEngine) RegisterDataAsset(asset DataAsset) {
	dge.dataInventory[asset.ID] = asset
}

// ClassifyData assigns classification to data
func (dge *DataGovernanceEngine) ClassifyData(assetID, classification string) error {
	asset, exists := dge.dataInventory[assetID]
	if !exists {
		return fmt.Errorf("asset not found: %s", assetID)
	}

	// Validate classification exists in policy
	classificationFound := false
	for _, policy := range dge.policies {
		for _, dc := range policy.DataClassifications {
			if dc.Level == classification {
				classificationFound = true
				break
			}
		}
		if classificationFound {
			break
		}
	}

	if !classificationFound {
		return fmt.Errorf("invalid classification: %s", classification)
	}

	asset.Classification = classification
	asset.Sensitivity = classification
	dge.dataInventory[assetID] = asset
	return nil
}

// CanAccess checks if data access is allowed
func (dge *DataGovernanceEngine) CanAccess(userID, assetID, action string) (bool, string) {
	asset, exists := dge.dataInventory[assetID]
	if !exists {
		return false, "asset not found"
	}

	// Find applicable policy
	for _, policy := range dge.policies {
		if !policy.Enabled {
			continue
		}

		for _, rule := range policy.DataHandlingRules {
			if rule.DataLevel == asset.Classification {
				// Check if action is allowed
				actionAllowed := false
				for _, allowed := range rule.AllowedActions {
					if allowed == action || allowed == "*" {
						actionAllowed = true
						break
					}
				}

				if !actionAllowed {
					return false, fmt.Sprintf("action %s not allowed for %s data", action, asset.Classification)
				}

				// Record access
				dge.recordAccess(assetID, userID, action)
				return true, ""
			}
		}
	}

	return true, ""
}

func (dge *DataGovernanceEngine) recordAccess(assetID, userID, action string) {
	asset, exists := dge.dataInventory[assetID]
	if !exists {
		return
	}

	record := AccessRecord{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Status:    "allowed",
	}

	asset.AccessLog = append(asset.AccessLog, record)
	asset.LastAccessedAt = time.Now()
	dge.dataInventory[assetID] = asset
}

// RunComplianceCheck runs compliance checks on data
func (dge *DataGovernanceEngine) RunComplianceCheck(assetID string) ComplianceCheck {
	asset, exists := dge.dataInventory[assetID]
	if !exists {
		return ComplianceCheck{
			Status: "fail",
			Issues: []string{"asset not found"},
		}
	}

	check := ComplianceCheck{
		ID:            fmt.Sprintf("check-%s-%d", assetID, time.Now().Unix()),
		CheckName:     fmt.Sprintf("Compliance Check for %s", assetID),
		DataAssetID:   assetID,
		Status:        "pass",
		LastCheckedAt: time.Now(),
		Issues:        make([]string, 0),
	}

	// Find applicable policies
	for _, policy := range dge.policies {
		// Check encryption requirements
		if policy.EncryptionPolicy.AtRestRequired {
			// In production, actually check if data is encrypted
			if !dge.isDataEncrypted(asset) {
				check.Issues = append(check.Issues, "Data not encrypted at rest")
				check.Status = "fail"
			}
		}

		// Check access controls
		if len(asset.Handlers) == 0 && len(policy.DataHandlingRules) > 0 {
			check.Issues = append(check.Issues, "No handlers assigned to data asset")
			check.Status = "fail"
		}

		// Check audit logging
		if policy.AuditingRequirements.Required && len(asset.AccessLog) == 0 {
			check.Issues = append(check.Issues, "Audit logging not enabled")
			check.Status = "fail"
		}
	}

	if len(check.Issues) == 0 {
		check.Status = "pass"
	}

	dge.complianceChecks[check.ID] = check
	return check
}

func (dge *DataGovernanceEngine) isDataEncrypted(asset DataAsset) bool {
	// In production, actually check encryption status
	return true
}

// DataMaskingPolicy defines data masking rules
type DataMaskingPolicy struct {
	ID          string
	Name        string
	Patterns    []MaskingPattern
	Enabled     bool
}

// MaskingPattern defines a masking pattern
type MaskingPattern struct {
	DataType    string // "ssn", "email", "phone", "credit_card"
	MaskType    string // "hash", "redact", "partial", "shuffle"
	MaskFormat  string // format string for masked output
}

// MaskData masks sensitive data based on policy
func (dmp *DataMaskingPolicy) MaskData(data string, dataType string) string {
	for _, pattern := range dmp.Patterns {
		if pattern.DataType == dataType {
			switch pattern.MaskType {
			case "hash":
				return hashData(data)
			case "redact":
				return "***REDACTED***"
			case "partial":
				return partialMask(data)
			case "shuffle":
				return shuffleData(data)
			}
		}
	}
	return data
}

func hashData(data string) string {
	// In production, use actual hash function
	return fmt.Sprintf("HASH_%d", len(data))
}

func partialMask(data string) string {
	if len(data) <= 4 {
		return "****"
	}
	return data[:2] + "****" + data[len(data)-2:]
}

func shuffleData(data string) string {
	// In production, actually shuffle
	return data
}

// DataLineage tracks data relationships
type DataLineage struct {
	AssetID  string
	Sources  []string // upstream data assets
	Targets  []string // downstream data assets
	Transform string   // transformation applied
	CreatedAt time.Time
}

// TraceLineage traces data lineage
func (dl *DataLineage) TraceUpstream() []string {
	var lineage []string
	lineage = append(lineage, dl.AssetID)
	lineage = append(lineage, dl.Sources...)
	return lineage
}

// TraceDownstream traces downstream impact
func (dl *DataLineage) TraceDownstream() []string {
	var lineage []string
	lineage = append(lineage, dl.AssetID)
	lineage = append(lineage, dl.Targets...)
	return lineage
}
