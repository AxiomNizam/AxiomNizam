package apiscanner

import "example.com/axiomnizam/internal/apiscanner/models"

// Re-export domain types from models subpackage.
type OutputFormat = models.OutputFormat
type VulnerabilityType = models.VulnerabilityType
type Endpoint = models.Endpoint
type ScanRequest = models.ScanRequest
type ScanResult = models.ScanResult
type ScanCheckStatus = models.ScanCheckStatus
type Finding = models.Finding
type Summary = models.Summary

// Re-export resource types from models subpackage.
type APIScanResource = models.APIScanResource
type APIScanSpec = models.APIScanSpec
type APIScanResourceStatus = models.APIScanResourceStatus

const (
	APIScanKind       = models.APIScanKind
	APIScanAPIVersion = models.APIScanAPIVersion
)

const (
	FormatTable = models.FormatTable
	FormatJSON  = models.FormatJSON
)

const (
	SeverityCritical = models.SeverityCritical
	SeverityHigh     = models.SeverityHigh
	SeverityMedium   = models.SeverityMedium
	SeverityLow      = models.SeverityLow
	SeverityInfo     = models.SeverityInfo
)

const (
	VulnAuthBypass      = models.VulnAuthBypass
	VulnSQLInjection    = models.VulnSQLInjection
	VulnNoSQLInjection  = models.VulnNoSQLInjection
	VulnHTTPMethod      = models.VulnHTTPMethod
	VulnSecurityHeaders = models.VulnSecurityHeaders
	VulnParameterTamper = models.VulnParameterTamper
	VulnXSS             = models.VulnXSS
)

const (
	CheckAuthBypassDetection = models.CheckAuthBypassDetection
	CheckAuthBypassTesting   = models.CheckAuthBypassTesting
	CheckSQLInjection        = models.CheckSQLInjection
	CheckNoSQLInjection      = models.CheckNoSQLInjection
	CheckHTTPMethod          = models.CheckHTTPMethod
	CheckSecurityHeaders     = models.CheckSecurityHeaders
	CheckParameterTampering  = models.CheckParameterTampering
	CheckXSS                 = models.CheckXSS
)
