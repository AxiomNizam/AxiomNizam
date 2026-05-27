package authn

import "strings"

// resolveIAMLoginIdentifier keeps username-style login compatible with IAM.
func resolveIAMLoginIdentifier(rawIdentifier string) string {
	identifier := strings.TrimSpace(rawIdentifier)
	if identifier == "" {
		return ""
	}

	// Preserve explicit email logins.
	if strings.Contains(identifier, "@") {
		return strings.ToLower(identifier)
	}

	// Keep sysadmin shorthand compatible with configured bootstrap email.
	sysadminEmail := strings.TrimSpace(getEnv("IAM_SYSADMIN_EMAIL", ""))
	if sysadminEmail != "" {
		normalizedSysadmin := strings.ToLower(sysadminEmail)
		parts := strings.SplitN(normalizedSysadmin, "@", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], identifier) {
			return normalizedSysadmin
		}
	}

	// Defer non-email usernames to IAM-side identifier lookup.
	return strings.ToLower(identifier)
}
