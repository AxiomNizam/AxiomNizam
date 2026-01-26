package policies

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ComplianceEngine manages compliance and audit requirements
type ComplianceEngine struct {
	mu                sync.RWMutex
	requirements      map[string]*ComplianceRequirement
	violations        []*ComplianceViolation
	auditTrail        []*AuditEntry
	complianceReports map[string]*ComplianceReport
	remediationPlans  map[string]*RemediationPlan
	maxAuditEntries   int
	retentionDays     int
}

// ComplianceRequirement defines a compliance requirement
type ComplianceRequirement struct {
	ID          string
	Name        string
	Description string
	Framework   string // HIPAA, PCI-DSS, SOC2, GDPR, etc.
	Rules       []*ComplianceRule
	Controls    []string
	CreatedAt   time.Time
}

// ComplianceRule defines a rule for compliance
type ComplianceRule struct {
	ID          string
	Name        string
	Description string
	Severity    string // Critical, High, Medium, Low
	Check       ComplianceCheckFn
	Remediation string
}

// ComplianceCheckFn checks compliance
type ComplianceCheckFn func(ctx context.Context, resource map[string]interface{}) (bool, string)

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID              string
	RequirementID   string
	RuleID          string
	ResourceName    string
	ResourceKind    string
	Severity        string
	Description     string
	DetectedAt      time.Time
	ResolvedAt      *time.Time
	Evidence        map[string]interface{}
	RemediationPlan *RemediationPlan
}

// RemediationPlan defines how to remediate a violation
type RemediationPlan struct {
	ID          string
	ViolationID string
	Steps       []*RemediationStep
	Status      string // Pending, InProgress, Completed, Failed
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// RemediationStep defines a remediation step
type RemediationStep struct {
	Order       int
	Description string
	Action      RemediationActionFn
	Status      string // Pending, InProgress, Completed, Failed
	ExecutedAt  *time.Time
	Error       string
}

// RemediationActionFn executes a remediation action
type RemediationActionFn func(ctx context.Context, violation *ComplianceViolation) error

// AuditEntry represents an audit entry
type AuditEntry struct {
	ID           string
	Timestamp    time.Time
	Actor        string // user, system, service
	Action       string
	ResourceKind string
	ResourceName string
	Namespace    string
	Changes      map[string]interface{}
	Result       string // Success, Failure
	Reason       string
	IPAddress    string
	UserAgent    string
	RequestID    string
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ID              string
	RequirementID   string
	GeneratedAt     time.Time
	Scope           string // namespace, all
	TotalResources  int
	ComplianceScore float64 // 0-100
	ViolationCount  int
	CriticalCount   int
	HighCount       int
	MediumCount     int
	LowCount        int
	Trends          map[string]interface{}
	Recommendations []string
	NextReviewDate  time.Time
}

// NewComplianceEngine creates a new compliance engine
func NewComplianceEngine(maxAuditEntries int, retentionDays int) *ComplianceEngine {
	return &ComplianceEngine{
		requirements:      make(map[string]*ComplianceRequirement),
		violations:        make([]*ComplianceViolation, 0),
		auditTrail:        make([]*AuditEntry, 0, maxAuditEntries),
		complianceReports: make(map[string]*ComplianceReport),
		remediationPlans:  make(map[string]*RemediationPlan),
		maxAuditEntries:   maxAuditEntries,
		retentionDays:     retentionDays,
	}
}

// RegisterRequirement registers a compliance requirement
func (ce *ComplianceEngine) RegisterRequirement(ctx context.Context, req *ComplianceRequirement) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if req.ID == "" {
		return fmt.Errorf("requirement ID is required")
	}

	if _, exists := ce.requirements[req.ID]; exists {
		return fmt.Errorf("requirement %s already exists", req.ID)
	}

	req.CreatedAt = time.Now()
	ce.requirements[req.ID] = req

	return nil
}

// CheckCompliance checks compliance for a resource
func (ce *ComplianceEngine) CheckCompliance(ctx context.Context, resource map[string]interface{}, requirementID string) ([]*ComplianceViolation, error) {
	ce.mu.RLock()
	req, exists := ce.requirements[requirementID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("requirement %s not found", requirementID)
	}

	var violations []*ComplianceViolation

	for _, rule := range req.Rules {
		compliant, reason := rule.Check(ctx, resource)

		if !compliant {
			violation := &ComplianceViolation{
				ID:            fmt.Sprintf("%s-%d", rule.ID, time.Now().Unix()),
				RequirementID: req.ID,
				RuleID:        rule.ID,
				Severity:      rule.Severity,
				Description:   reason,
				DetectedAt:    time.Now(),
				Evidence:      resource,
			}

			violations = append(violations, violation)

			// Store violation
			ce.mu.Lock()
			ce.violations = append(ce.violations, violation)
			ce.mu.Unlock()
		}
	}

	return violations, nil
}

// RecordAuditEntry records an audit entry
func (ce *ComplianceEngine) RecordAuditEntry(ctx context.Context, entry *AuditEntry) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	entry.Timestamp = time.Now()

	ce.auditTrail = append(ce.auditTrail, entry)

	// Keep only maxAuditEntries
	if len(ce.auditTrail) > ce.maxAuditEntries {
		ce.auditTrail = ce.auditTrail[len(ce.auditTrail)-ce.maxAuditEntries:]
	}
}

// SearchAuditTrail searches audit trail
func (ce *ComplianceEngine) SearchAuditTrail(ctx context.Context, filters map[string]interface{}) []*AuditEntry {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	var results []*AuditEntry

	for _, entry := range ce.auditTrail {
		if matches(entry, filters) {
			results = append(results, entry)
		}
	}

	return results
}

// GetViolations gets violations
func (ce *ComplianceEngine) GetViolations(ctx context.Context, filters map[string]interface{}) []*ComplianceViolation {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	var results []*ComplianceViolation

	for _, violation := range ce.violations {
		if matches(violation, filters) {
			results = append(results, violation)
		}
	}

	return results
}

// CreateRemediationPlan creates a remediation plan
func (ce *ComplianceEngine) CreateRemediationPlan(ctx context.Context, plan *RemediationPlan) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if plan.ID == "" {
		return fmt.Errorf("plan ID is required")
	}

	plan.Status = "Pending"
	plan.CreatedAt = time.Now()

	ce.remediationPlans[plan.ID] = plan

	return nil
}

// ExecuteRemediationPlan executes a remediation plan
func (ce *ComplianceEngine) ExecuteRemediationPlan(ctx context.Context, planID string) error {
	ce.mu.Lock()
	plan, exists := ce.remediationPlans[planID]
	ce.mu.Unlock()

	if !exists {
		return fmt.Errorf("plan %s not found", planID)
	}

	plan.Status = "InProgress"

	// Find violation
	ce.mu.RLock()
	var violation *ComplianceViolation
	for _, v := range ce.violations {
		if v.ID == plan.ViolationID {
			violation = v
			break
		}
	}
	ce.mu.RUnlock()

	if violation == nil {
		return fmt.Errorf("violation %s not found", plan.ViolationID)
	}

	// Execute steps
	for i, step := range plan.Steps {
		if step.Status == "Completed" {
			continue
		}

		step.Status = "InProgress"
		now := time.Now()
		step.ExecutedAt = &now

		if err := step.Action(ctx, violation); err != nil {
			step.Status = "Failed"
			step.Error = err.Error()
			plan.Status = "Failed"
			return err
		}

		step.Status = "Completed"
	}

	now := time.Now()
	plan.CompletedAt = &now
	plan.Status = "Completed"

	// Mark violation as resolved
	ce.mu.Lock()
	violation.ResolvedAt = &now
	ce.mu.Unlock()

	return nil
}

// GenerateComplianceReport generates a compliance report
func (ce *ComplianceEngine) GenerateComplianceReport(ctx context.Context, requirementID string) (*ComplianceReport, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	req, exists := ce.requirements[requirementID]
	if !exists {
		return nil, fmt.Errorf("requirement %s not found", requirementID)
	}

	report := &ComplianceReport{
		ID:              fmt.Sprintf("%s-%d", requirementID, time.Now().Unix()),
		RequirementID:   requirementID,
		GeneratedAt:     time.Now(),
		NextReviewDate:  time.Now().AddDate(0, 1, 0),
		Trends:          make(map[string]interface{}),
		Recommendations: make([]string, 0),
	}

	// Count violations
	for _, violation := range ce.violations {
		if violation.RequirementID != requirementID {
			continue
		}

		if violation.ResolvedAt == nil {
			report.ViolationCount++

			switch violation.Severity {
			case "Critical":
				report.CriticalCount++
			case "High":
				report.HighCount++
			case "Medium":
				report.MediumCount++
			case "Low":
				report.LowCount++
			}
		}
	}

	// Calculate compliance score
	totalRules := len(req.Rules)
	if totalRules > 0 {
		compliantRules := totalRules - report.ViolationCount
		report.ComplianceScore = (float64(compliantRules) / float64(totalRules)) * 100
	}

	// Store report
	ce.complianceReports[report.ID] = report

	return report, nil
}

// GetComplianceReports gets compliance reports
func (ce *ComplianceEngine) GetComplianceReports(ctx context.Context) []*ComplianceReport {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	reports := make([]*ComplianceReport, 0, len(ce.complianceReports))
	for _, report := range ce.complianceReports {
		reports = append(reports, report)
	}

	return reports
}

// CleanupExpiredEntries cleans up expired audit entries and old violations
func (ce *ComplianceEngine) CleanupExpiredEntries(ctx context.Context) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	cutoffDate := time.Now().AddDate(0, 0, -ce.retentionDays)

	// Clean audit trail
	var newAuditTrail []*AuditEntry
	for _, entry := range ce.auditTrail {
		if entry.Timestamp.After(cutoffDate) {
			newAuditTrail = append(newAuditTrail, entry)
		}
	}
	ce.auditTrail = newAuditTrail

	// Clean resolved violations
	var newViolations []*ComplianceViolation
	for _, violation := range ce.violations {
		if violation.ResolvedAt == nil || violation.ResolvedAt.After(cutoffDate) {
			newViolations = append(newViolations, violation)
		}
	}
	ce.violations = newViolations

	return nil
}

// matches checks if object matches filters
func matches(obj interface{}, filters map[string]interface{}) bool {
	for key, expectedValue := range filters {
		var objValue interface{}

		// Extract value from object
		switch o := obj.(type) {
		case *AuditEntry:
			switch key {
			case "actor":
				objValue = o.Actor
			case "action":
				objValue = o.Action
			case "resource_kind":
				objValue = o.ResourceKind
			case "result":
				objValue = o.Result
			}
		case *ComplianceViolation:
			switch key {
			case "requirement_id":
				objValue = o.RequirementID
			case "severity":
				objValue = o.Severity
			case "status":
				if o.ResolvedAt == nil {
					objValue = "Open"
				} else {
					objValue = "Resolved"
				}
			}
		}

		if objValue != expectedValue {
			return false
		}
	}

	return true
}
