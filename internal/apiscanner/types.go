package apiscanner

import "time"

type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

const (
	SeverityCritical = "CRITICAL"
	SeverityHigh     = "HIGH"
	SeverityMedium   = "MEDIUM"
	SeverityLow      = "LOW"
	SeverityInfo     = "INFO"
)

type VulnerabilityType string

const (
	VulnAuthBypass      VulnerabilityType = "auth_bypass"
	VulnSQLInjection    VulnerabilityType = "sql_injection"
	VulnNoSQLInjection  VulnerabilityType = "nosql_injection"
	VulnHTTPMethod      VulnerabilityType = "http_method_validation"
	VulnSecurityHeaders VulnerabilityType = "security_headers"
	VulnParameterTamper VulnerabilityType = "parameter_tampering"
	VulnXSS             VulnerabilityType = "xss"
)

type Endpoint struct {
	URL     string
	Method  string
	Body    string
	Headers map[string]string
}

type ScanRequest struct {
	Endpoint           Endpoint
	Timeout            time.Duration
	RetryCount         int
	RetryBackoff       time.Duration
	InsecureSkipVerify bool
	AuthHeader         string
	AuthValue          string
	Format             OutputFormat
}

type ScanResult struct {
	Scanner   string            `json:"scanner"`
	Target    string            `json:"target"`
	Method    string            `json:"method"`
	ScannedAt time.Time         `json:"scannedAt"`
	Findings  []Finding         `json:"findings"`
	Checks    []ScanCheckStatus `json:"checks,omitempty"`
	Summary   Summary           `json:"summary"`
}

const (
	CheckAuthBypassDetection = "auth_bypass_detection"
	CheckAuthBypassTesting   = "auth_bypass_testing"
	CheckSQLInjection        = "sql_injection"
	CheckNoSQLInjection      = "nosql_injection"
	CheckHTTPMethod          = "http_method_validation"
	CheckSecurityHeaders     = "security_header_analysis"
	CheckParameterTampering  = "parameter_tampering"
	CheckXSS                 = "xss"
)

type ScanCheckStatus struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Executed bool   `json:"executed"`
	Findings int    `json:"findings"`
}

type Finding struct {
	Type           VulnerabilityType `json:"type"`
	Severity       string            `json:"severity"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Endpoint       string            `json:"endpoint"`
	Method         string            `json:"method"`
	Evidence       string            `json:"evidence,omitempty"`
	Payload        string            `json:"payload,omitempty"`
	Recommendation string            `json:"recommendation,omitempty"`
}

type Summary struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
}
