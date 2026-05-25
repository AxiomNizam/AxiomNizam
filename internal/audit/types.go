package audit

import "example.com/axiomnizam/internal/audit/models"

// Re-export domain Resource types from models subpackage.
type AuditPolicyResource = models.AuditPolicyResource
type AuditPolicySpec = models.AuditPolicySpec
type AuditPolicyResourceStatus = models.AuditPolicyResourceStatus

const (
	AuditPolicyKind       = models.AuditPolicyKind
	AuditPolicyAPIVersion = models.AuditPolicyAPIVersion
)
