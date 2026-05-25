package audit

// Storage event type constants.
const (
	EventBucketCreated        = "bucket.created"
	EventBucketDeleted        = "bucket.deleted"
	EventObjectUploaded       = "object.uploaded"
	EventObjectDownloaded     = "object.downloaded"
	EventObjectDeleted        = "object.deleted"
	EventObjectCopied         = "object.copied"
	EventPolicyCreated        = "policy.created"
	EventPolicyDeleted        = "policy.deleted"
	EventPresignGenerated     = "presign.generated"
	EventMultiDelete          = "object.multi-deleted"
	EventObjectScanClean      = "object.scan.clean"
	EventObjectThreatDetected = "object.scan.threat"
)

// Severity levels for storage audit events.
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Event category constants.
const (
	CategoryBucket = "bucket"
	CategoryObject = "object"
	CategoryPolicy = "policy"
	CategoryScan   = "scan"
)
