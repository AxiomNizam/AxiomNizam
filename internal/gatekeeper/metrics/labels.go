package metrics

// Metric label constants for consistent Prometheus label usage.
const (
	LabelFactorType    = "factor_type"    // totp, sms, email, webauthn
	LabelChallengeType = "challenge_type" // enrollment, verification, recovery
	LabelOutcome       = "outcome"        // success, failure, timeout, expired
	LabelRiskLevel     = "risk_level"     // low, medium, high, critical
	LabelUserAction    = "user_action"    // enroll, activate, disable, verify, trust_device
)

// Common label value sets.
var (
	FactorTypes    = []string{"totp", "sms", "email", "webauthn", "backup"}
	ChallengeTypes = []string{"enrollment", "verification", "recovery"}
	Outcomes       = []string{"success", "failure", "timeout", "expired"}
	RiskLevels     = []string{"low", "medium", "high", "critical"}
	UserActions    = []string{"enroll", "activate", "disable", "verify", "trust_device", "revoke_device"}
)
