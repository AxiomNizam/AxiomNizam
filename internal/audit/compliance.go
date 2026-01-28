package audit

import (
	"fmt"
	"sync"
	"time"
)

// ComplianceRule defines a compliance requirement
type ComplianceRule struct {
	ID          string
	Framework   string // GDPR, HIPAA, SOC2, PCI-DSS
	Requirement string
	Description string
	IsActive    bool
	CreatedAt   time.Time
}

// ComplianceViolation represents compliance violation
type ComplianceViolation struct {
	ID               string
	Timestamp        time.Time
	RuleID           string
	Framework        string
	ViolationType    string
	Severity         string // low, medium, high, critical
	AffectedResource string
	Description      string
	RemediationSteps string
	Status           string // open, acknowledged, resolved
}

// ComplianceReport represents generated compliance report
type ComplianceReport struct {
	ID              string
	ReportType      string // GDPR, HIPAA, SOC2, etc
	GeneratedAt     time.Time
	StartDate       time.Time
	EndDate         time.Time
	TotalEvents     int64
	ComplianceScore float64 // 0-100
	ViolationCount  int64
	RemediatedCount int64
	Frameworks      []string
	Findings        []*ComplianceFinding
	RiskAssessment  *RiskAssessment
}

// ComplianceFinding represents compliance finding
type ComplianceFinding struct {
	ID          string
	Framework   string
	Category    string
	Title       string
	Description string
	Severity    string
	Evidence    []string
}

// RiskAssessment represents risk assessment
type RiskAssessment struct {
	OverallRisk        string // low, medium, high, critical
	RiskScore          float64
	TopRisks           []string
	MitigationSteps    []string
	RecommendedActions []string
}

// AuditComplianceManager manages audit and compliance
type AuditComplianceManager struct {
	mu               sync.RWMutex
	auditLogs        []*AuditLog
	complianceRules  map[string]*ComplianceRule
	violations       []*ComplianceViolation
	reports          []*ComplianceReport
	auditMetrics     *AuditMetrics
	maxAuditLogSize  int
	maxViolationSize int
	maxReportSize    int
	retentionDays    int
}

// AuditMetrics tracks audit statistics
type AuditMetrics struct {
	TotalAuditLogs      int64
	ActionsTracked      int64
	ViolationsFound     int64
	ReportsGenerated    int64
	AverageResponseTime float64
	LastAuditTime       time.Time
}

// NewAuditComplianceManager creates audit manager
func NewAuditComplianceManager() *AuditComplianceManager {
	return &AuditComplianceManager{
		auditLogs:        make([]*AuditLog, 0),
		complianceRules:  make(map[string]*ComplianceRule),
		violations:       make([]*ComplianceViolation, 0),
		reports:          make([]*ComplianceReport, 0),
		auditMetrics:     &AuditMetrics{},
		maxAuditLogSize:  100000,
		maxViolationSize: 50000,
		maxReportSize:    1000,
		retentionDays:    365,
	}
}

// LogAuditEvent logs an audit event
func (acm *AuditComplianceManager) LogAuditEvent(event *AuditLog) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	acm.auditLogs = append(acm.auditLogs, event)
	acm.auditMetrics.TotalAuditLogs++
	acm.auditMetrics.ActionsTracked++
	acm.auditMetrics.LastAuditTime = time.Now()

	if len(acm.auditLogs) > acm.maxAuditLogSize {
		acm.auditLogs = acm.auditLogs[1:]
	}

	return nil
}

// RegisterComplianceRule registers a compliance rule
func (acm *AuditComplianceManager) RegisterComplianceRule(rule *ComplianceRule) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule-%s-%d", rule.Framework, time.Now().UnixNano())
	}

	rule.CreatedAt = time.Now()
	acm.complianceRules[rule.ID] = rule
	return nil
}

// RecordViolation records a compliance violation
func (acm *AuditComplianceManager) RecordViolation(violation *ComplianceViolation) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	if violation.ID == "" {
		violation.ID = fmt.Sprintf("vio-%d", time.Now().UnixNano())
	}

	if violation.Timestamp.IsZero() {
		violation.Timestamp = time.Now()
	}

	violation.Status = "open"
	acm.violations = append(acm.violations, violation)
	acm.auditMetrics.ViolationsFound++

	if len(acm.violations) > acm.maxViolationSize {
		acm.violations = acm.violations[1:]
	}

	return nil
}

// GenerateComplianceReport generates a compliance report
func (acm *AuditComplianceManager) GenerateComplianceReport(reportType string, startDate, endDate time.Time) (*ComplianceReport, error) {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	report := &ComplianceReport{
		ID:          fmt.Sprintf("report-%s-%d", reportType, time.Now().UnixNano()),
		ReportType:  reportType,
		GeneratedAt: time.Now(),
		StartDate:   startDate,
		EndDate:     endDate,
		Findings:    make([]*ComplianceFinding, 0),
		Frameworks:  []string{reportType},
	}

	// Count events in range
	for _, log := range acm.auditLogs {
		if log.Timestamp.After(startDate) && log.Timestamp.Before(endDate) {
			report.TotalEvents++
		}
	}

	// Find violations
	for _, vio := range acm.violations {
		if vio.Timestamp.After(startDate) && vio.Timestamp.Before(endDate) && vio.Framework == reportType {
			report.ViolationCount++
			if vio.Status == "resolved" {
				report.RemediatedCount++
			}

			finding := &ComplianceFinding{
				ID:          vio.ID,
				Framework:   vio.Framework,
				Title:       vio.ViolationType,
				Description: vio.Description,
				Severity:    vio.Severity,
			}
			report.Findings = append(report.Findings, finding)
		}
	}

	// Calculate compliance score
	if report.TotalEvents > 0 {
		report.ComplianceScore = (float64(report.RemediatedCount) / float64(report.ViolationCount)) * 100
	}

	// Generate risk assessment
	report.RiskAssessment = acm.assessRisk(report)

	acm.reports = append(acm.reports, report)
	acm.auditMetrics.ReportsGenerated++

	if len(acm.reports) > acm.maxReportSize {
		acm.reports = acm.reports[1:]
	}

	return report, nil
}

// assessRisk generates risk assessment
func (acm *AuditComplianceManager) assessRisk(report *ComplianceReport) *RiskAssessment {
	assessment := &RiskAssessment{
		TopRisks:           make([]string, 0),
		MitigationSteps:    make([]string, 0),
		RecommendedActions: make([]string, 0),
	}

	criticalCount := 0
	highCount := 0
	mediumCount := 0

	for _, finding := range report.Findings {
		switch finding.Severity {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		case "medium":
			mediumCount++
		}
	}

	// Determine overall risk
	if criticalCount > 0 {
		assessment.OverallRisk = "critical"
		assessment.RiskScore = 95.0
	} else if highCount > 3 {
		assessment.OverallRisk = "high"
		assessment.RiskScore = 75.0
	} else if mediumCount > 5 {
		assessment.OverallRisk = "medium"
		assessment.RiskScore = 50.0
	} else {
		assessment.OverallRisk = "low"
		assessment.RiskScore = 25.0
	}

	// Add recommended actions
	if criticalCount > 0 {
		assessment.RecommendedActions = append(assessment.RecommendedActions,
			"Immediate remediation required for critical findings",
			"Security review recommended",
			"Consider incident response",
		)
	}

	if report.ComplianceScore < 50 {
		assessment.RecommendedActions = append(assessment.RecommendedActions,
			"Implement missing controls",
			"Review compliance procedures",
			"Enhanced monitoring required",
		)
	}

	return assessment
}

// GetAuditLogs gets audit logs
func (acm *AuditComplianceManager) GetAuditLogs(limit int) []*AuditLog {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	if limit > len(acm.auditLogs) {
		limit = len(acm.auditLogs)
	}
	if limit == 0 {
		return make([]*AuditLog, 0)
	}

	return acm.auditLogs[len(acm.auditLogs)-limit:]
}

// GetViolations gets compliance violations
func (acm *AuditComplianceManager) GetViolations(status string) []*ComplianceViolation {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	violations := make([]*ComplianceViolation, 0)

	for _, vio := range acm.violations {
		if status == "" || vio.Status == status {
			violations = append(violations, vio)
		}
	}

	return violations
}

// ResolveViolation marks violation as resolved
func (acm *AuditComplianceManager) ResolveViolation(violationID string) error {
	acm.mu.Lock()
	defer acm.mu.Unlock()

	for _, vio := range acm.violations {
		if vio.ID == violationID {
			vio.Status = "resolved"
			return nil
		}
	}

	return fmt.Errorf("violation not found")
}

// GetComplianceReports gets generated reports
func (acm *AuditComplianceManager) GetComplianceReports(limit int) []*ComplianceReport {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	if limit > len(acm.reports) {
		limit = len(acm.reports)
	}
	if limit == 0 {
		return make([]*ComplianceReport, 0)
	}

	return acm.reports[len(acm.reports)-limit:]
}

// GetAuditMetrics returns audit metrics
func (acm *AuditComplianceManager) GetAuditMetrics() *AuditMetrics {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	return &AuditMetrics{
		TotalAuditLogs:   acm.auditMetrics.TotalAuditLogs,
		ActionsTracked:   acm.auditMetrics.ActionsTracked,
		ViolationsFound:  acm.auditMetrics.ViolationsFound,
		ReportsGenerated: acm.auditMetrics.ReportsGenerated,
		LastAuditTime:    acm.auditMetrics.LastAuditTime,
	}
}

// SearchAuditLogs searches audit logs
func (acm *AuditComplianceManager) SearchAuditLogs(userID, resourceType string) []*AuditLog {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	results := make([]*AuditLog, 0)

	for _, log := range acm.auditLogs {
		if (userID == "" || log.UserID == userID) &&
			(resourceType == "" || log.ResourceType == resourceType) {
			results = append(results, log)
		}
	}

	return results
}

// GetComplianceStatus returns overall compliance status
func (acm *AuditComplianceManager) GetComplianceStatus() map[string]interface{} {
	acm.mu.RLock()
	defer acm.mu.RUnlock()

	activeViolations := 0
	for _, vio := range acm.violations {
		if vio.Status == "open" {
			activeViolations++
		}
	}

	latestReport := &ComplianceReport{}
	if len(acm.reports) > 0 {
		latestReport = acm.reports[len(acm.reports)-1]
	}

	return map[string]interface{}{
		"total_audit_logs":        acm.auditMetrics.TotalAuditLogs,
		"total_violations":        acm.auditMetrics.ViolationsFound,
		"active_violations":       activeViolations,
		"reports_generated":       acm.auditMetrics.ReportsGenerated,
		"compliance_rules":        len(acm.complianceRules),
		"latest_compliance_score": latestReport.ComplianceScore,
		"last_audit_time":         acm.auditMetrics.LastAuditTime,
	}
}
