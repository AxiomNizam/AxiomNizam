package metrics

// Metric label constants for consistent Prometheus label usage.
const (
	LabelAuthMethod  = "auth_method"  // password, token, sso, api_key
	LabelGrantType   = "grant_type"   // authorization_code, client_credentials, refresh_token
	LabelOutcome     = "outcome"      // success, failure, denied, error
	LabelResource    = "resource"     // user, client, role, group, session
	LabelRealm       = "realm"        // IAM realm identifier
)

// Common label value sets.
var (
	AuthMethods = []string{"password", "token", "sso", "api_key"}
	GrantTypes  = []string{"authorization_code", "client_credentials", "refresh_token", "password"}
	Outcomes    = []string{"success", "failure", "denied", "error"}
	Resources   = []string{"user", "client", "role", "group", "session", "realm"}
)
