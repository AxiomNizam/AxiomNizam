package security

import (
	"fmt"
	"time"
)

// SecurityPolicy defines security and compliance policies
type SecurityPolicy struct {
	ID                   string
	Name                 string
	Type                 string
	Version              string
	Enabled              bool
	AuthenticationPolicy  AuthenticationPolicy
	AuthorizationPolicy   AuthorizationPolicy
	EncryptionPolicy      EncryptionPolicy
	AuditPolicy           AuditPolicy
	VulnerabilityPolicy   VulnerabilityPolicy
	ComplianceFrameworks  []ComplianceFramework
	Description          string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// AuthenticationPolicy defines authentication requirements
type AuthenticationPolicy struct {
	MFARequired          bool
	MinPasswordLength    int
	PasswordExpiration   int // days
	HistoryCheck         int // remember last N passwords
	LockoutPolicy        LockoutPolicy
	SessionTimeout       time.Duration
	AllowedMethods       []string // "password", "mfa", "oauth", "saml"
}

// LockoutPolicy defines account lockout rules
type LockoutPolicy struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
	ResetAfter        time.Duration
}

// AuthorizationPolicy defines authorization requirements
type AuthorizationPolicy struct {
	DefaultDeny          bool
	RequireApproval      bool
	SeparationOfDuties   bool
	PrivilegeEscalation  PrivilegePolicy
	ResourceBasedPolicy  bool
}

// PrivilegePolicy defines privilege escalation rules
type PrivilegePolicy struct {
	Allowed              bool
	RequireApproval      bool
	TimeLimit            time.Duration
	NotifyOnEscalation   bool
	MaxElevationDuration time.Duration
}

// EncryptionPolicy defines encryption requirements
type EncryptionPolicy struct {
	Algorithm              string
	MinKeyLength           int
	RequireAtRest          bool
	RequireInTransit       bool
	TLSMinVersion          string // "1.2", "1.3"
	CertificateValidation  bool
	CertificatePinning     bool
	PerfectForwardSecrecy  bool
}

// AuditPolicy defines auditing requirements
type AuditPolicy struct {
	LogAllAccess         bool
	LogModifications     bool
	LogDeletions         bool
	LogAuthentication    bool
	LogAuthorizationFailures bool
	RetentionDays        int
	AlertOnSuspiciousActivity bool
	SuspiciousPatterns   []string
}

// VulnerabilityPolicy defines vulnerability management
type VulnerabilityPolicy struct {
	ScanFrequency        string // "daily", "weekly", "monthly"
	MaxCriticalVulnerabilities int
	MaxHighVulnerabilities int
	PatchingTimeLimit    time.Duration
	ZeroDayResponse      string // immediate action required
}

// ComplianceFramework defines compliance frameworks
type ComplianceFramework struct {
	Name      string // "HIPAA", "GDPR", "SOC2", "PCI-DSS", "ISO27001"
	Enabled   bool
	Controls  []ComplianceControl
}

// ComplianceControl defines a compliance control
type ComplianceControl struct {
	ID          string
	Name        string
	Description string
	Evidence    []string
	Status      string // "compliant", "non-compliant", "in-progress"
	LastChecked time.Time
}

// GetID returns policy ID
func (sp *SecurityPolicy) GetID() string {
	return sp.ID
}

// GetName returns policy name
func (sp *SecurityPolicy) GetName() string {
	return sp.Name
}

// GetType returns policy type
func (sp *SecurityPolicy) GetType() string {
	return sp.Type
}

// GetVersion returns version
func (sp *SecurityPolicy) GetVersion() string {
	return sp.Version
}

// GetEnabled returns if enabled
func (sp *SecurityPolicy) GetEnabled() bool {
	return sp.Enabled
}

// Validate validates the policy
func (sp *SecurityPolicy) Validate() error {
	if sp.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if sp.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	return nil
}

// SecurityComplianceEngine manages security and compliance
type SecurityComplianceEngine struct {
	policies             []*SecurityPolicy
	vulnerabilities      map[string]Vulnerability
	incidents            map[string]SecurityIncident
	complianceStatus     map[string]ComplianceStatus
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID           string
	CVE          string
	Severity     string // "critical", "high", "medium", "low"
	Description  string
	AffectedSystems []string
	DiscoveredAt time.Time
	PatchedAt    time.Time
	Status       string // "open", "patched", "acknowledged"
}

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID              string
	Type            string // "intrusion", "data_breach", "malware", "unauthorized_access"
	Severity        string
	Description     string
	DiscoveredAt    time.Time
	ReportedAt      time.Time
	ResolvedAt      time.Time
	AffectedAssets  []string
	RootCause       string
	Remediation     string
	Status          string // "open", "investigating", "resolved"
}

// ComplianceStatus tracks compliance status
type ComplianceStatus struct {
	Framework         string
	Status            string // "compliant", "non-compliant", "partial"
	LastAuditDate     time.Time
	NextAuditDate     time.Time
	Issues            []string
	Remediation       map[string]string
	CompliancePercent float64
}

// NewSecurityComplianceEngine creates a new security compliance engine
func NewSecurityComplianceEngine() *SecurityComplianceEngine {
	return &SecurityComplianceEngine{
		policies:         make([]*SecurityPolicy, 0),
		vulnerabilities:  make(map[string]Vulnerability),
		incidents:        make(map[string]SecurityIncident),
		complianceStatus: make(map[string]ComplianceStatus),
	}
}

// RegisterPolicy registers a security policy
func (sce *SecurityComplianceEngine) RegisterPolicy(policy *SecurityPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	sce.policies = append(sce.policies, policy)
	return nil
}

// ReportVulnerability reports a vulnerability
func (sce *SecurityComplianceEngine) ReportVulnerability(vuln Vulnerability) {
	sce.vulnerabilities[vuln.ID] = vuln
}

// ReportIncident reports a security incident
func (sce *SecurityComplianceEngine) ReportIncident(incident SecurityIncident) {
	sce.incidents[incident.ID] = incident
}

// IsCompliant checks if system is compliant with policies
func (sce *SecurityComplianceEngine) IsCompliant(framework string) (bool, ComplianceStatus) {
	status, exists := sce.complianceStatus[framework]
	if !exists {
		status = ComplianceStatus{
			Framework: framework,
			Status:    "unknown",
		}
	}

	for _, policy := range sce.policies {
		for _, cf := range policy.ComplianceFrameworks {
			if cf.Name == framework {
				// Check controls
				compliantControls := 0
				for _, control := range cf.Controls {
					if control.Status == "compliant" {
						compliantControls++
					}
				}

				total := len(cf.Controls)
				if total > 0 {
					status.CompliancePercent = float64(compliantControls) / float64(total) * 100
					if status.CompliancePercent == 100 {
						status.Status = "compliant"
					} else if status.CompliancePercent >= 80 {
						status.Status = "partial"
					} else {
						status.Status = "non-compliant"
					}
				}
			}
		}
	}

	return status.Status == "compliant", status
}

// RunSecurityAudit runs a security audit
func (sce *SecurityComplianceEngine) RunSecurityAudit() SecurityAuditResult {
	result := SecurityAuditResult{
		Timestamp:      time.Now(),
		PolicyFindings: make([]AuditFinding, 0),
	}

	// Check vulnerabilities
	critical := 0
	high := 0
	for _, vuln := range sce.vulnerabilities {
		if vuln.Status == "open" {
			if vuln.Severity == "critical" {
				critical++
			} else if vuln.Severity == "high" {
				high++
			}
		}
	}

	if critical > 0 {
		result.PolicyFindings = append(result.PolicyFindings, AuditFinding{
			Severity:    "critical",
			Description: fmt.Sprintf("Found %d critical vulnerabilities", critical),
			Recommendation: "Patch immediately",
		})
	}

	if high > 0 {
		result.PolicyFindings = append(result.PolicyFindings, AuditFinding{
			Severity:    "high",
			Description: fmt.Sprintf("Found %d high-severity vulnerabilities", high),
			Recommendation: "Patch within 30 days",
		})
	}

	// Check incidents
	if len(sce.incidents) > 0 {
		result.PolicyFindings = append(result.PolicyFindings, AuditFinding{
			Severity:    "high",
			Description: fmt.Sprintf("Found %d unresolved security incidents", countUnresolved(sce.incidents)),
			Recommendation: "Investigate and resolve",
		})
	}

	return result
}

// SecurityAuditResult represents audit results
type SecurityAuditResult struct {
	Timestamp      time.Time
	PolicyFindings []AuditFinding
	Score          int // 0-100
}

// AuditFinding represents an audit finding
type AuditFinding struct {
	Severity       string
	Description    string
	Recommendation string
}

func countUnresolved(incidents map[string]SecurityIncident) int {
	count := 0
	for _, incident := range incidents {
		if incident.Status != "resolved" {
			count++
		}
	}
	return count
}

// ThreatModel represents a threat model
type ThreatModel struct {
	ID              string
	Name            string
	Assets          []Asset
	Threats         []Threat
	Vulnerabilities []Vulnerability
	Controls        []SecurityControl
}

// Asset represents a system asset
type Asset struct {
	ID    string
	Name  string
	Value string // "critical", "high", "medium", "low"
}

// Threat represents a threat
type Threat struct {
	ID          string
	Name        string
	Source      string
	Probability float64 // 0-1
	Impact      string // "critical", "high", "medium", "low"
	MitigationControls []string
}

// SecurityControl represents a security control
type SecurityControl struct {
	ID          string
	Name        string
	Type        string // "preventive", "detective", "corrective"
	Effectiveness float64
	CostBenefit float64
}

// CreateThreatModel creates a new threat model
func CreateThreatModel(id, name string) *ThreatModel {
	return &ThreatModel{
		ID:              id,
		Name:            name,
		Assets:          make([]Asset, 0),
		Threats:         make([]Threat, 0),
		Vulnerabilities: make([]Vulnerability, 0),
		Controls:        make([]SecurityControl, 0),
	}
}

// IdentifyRisks identifies risks in the threat model
func (tm *ThreatModel) IdentifyRisks() []Risk {
	var risks []Risk

	for _, threat := range tm.Threats {
		riskScore := threat.Probability * riskScoreForImpact(threat.Impact)
		risk := Risk{
			ID:        fmt.Sprintf("risk-%s", threat.ID),
			Threat:    threat.Name,
			Score:     riskScore,
			Mitigation: findMitigation(tm.Controls, threat.MitigationControls),
		}
		risks = append(risks, risk)
	}

	return risks
}

// Risk represents an identified risk
type Risk struct {
	ID         string
	Threat     string
	Score      float64 // 0-1
	Mitigation []SecurityControl
}

func riskScoreForImpact(impact string) float64 {
	switch impact {
	case "critical":
		return 1.0
	case "high":
		return 0.75
	case "medium":
		return 0.5
	case "low":
		return 0.25
	default:
		return 0
	}
}

func findMitigation(controls []SecurityControl, controlIDs []string) []SecurityControl {
	var mitigations []SecurityControl
	for _, control := range controls {
		for _, id := range controlIDs {
			if control.ID == id {
				mitigations = append(mitigations, control)
			}
		}
	}
	return mitigations
}
