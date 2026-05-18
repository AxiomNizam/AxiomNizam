package trusteddevices

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// FingerprintData contains browser/device signals for fingerprinting.
type FingerprintData struct {
	UserAgent  string
	AcceptLang string
	ScreenRes  string
	Timezone   string
	Platform   string
}

// GenerateFingerprint creates a stable device fingerprint from browser signals.
func GenerateFingerprint(data *FingerprintData) string {
	// Combine signals into a stable string
	parts := []string{
		data.UserAgent,
		data.AcceptLang,
		data.ScreenRes,
		data.Timezone,
		data.Platform,
	}
	combined := strings.Join(parts, "|")

	// Hash to create a stable fingerprint
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// NormalizeUserAgent normalizes a user agent string for consistent fingerprinting.
func NormalizeUserAgent(ua string) string {
	// Trim whitespace and lowercase
	ua = strings.TrimSpace(ua)
	// Remove version numbers for stability
	// (simplified - production would use a proper UA parser)
	return ua
}

// ExtractBrowser extracts the browser name from a user agent string.
func ExtractBrowser(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "firefox"):
		return "firefox"
	case strings.Contains(ua, "edg"):
		return "edge"
	case strings.Contains(ua, "chrome"):
		return "chrome"
	case strings.Contains(ua, "safari"):
		return "safari"
	default:
		return "unknown"
	}
}

// ExtractOS extracts the OS name from a user agent string.
func ExtractOS(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "windows"):
		return "windows"
	case strings.Contains(ua, "mac os"):
		return "macos"
	case strings.Contains(ua, "linux"):
		return "linux"
	case strings.Contains(ua, "android"):
		return "android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "ios"
	default:
		return "unknown"
	}
}

// FingerprintMatch checks if two fingerprints are similar enough to be the same device.
func FingerprintMatch(fp1, fp2 string) bool {
	return fp1 == fp2
}

// FingerprintSummary returns a human-readable summary of the fingerprint.
func FingerprintSummary(data *FingerprintData) string {
	return fmt.Sprintf("%s on %s", ExtractBrowser(data.UserAgent), ExtractOS(data.UserAgent))
}
