package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"net/url"
	"strings"
	"unicode"
)

// SecurityHeaders defines security headers for HTTP responses
type SecurityHeaders struct {
	ContentSecurityPolicy  string
	XFrameOptions          string
	XContentTypeOptions    string
	StrictTransportSecurity string
	XSSProtection          string
	ReferrerPolicy         string
	PermissionsPolicy      string
}

// DefaultSecurityHeaders returns the default security headers
func DefaultSecurityHeaders() SecurityHeaders {
	return SecurityHeaders{
		ContentSecurityPolicy:  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		XFrameOptions:          "DENY",
		XContentTypeOptions:    "nosniff",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
		XSSProtection:          "1; mode=block",
		ReferrerPolicy:         "strict-origin-when-cross-origin",
		PermissionsPolicy:      "camera=(), microphone=(), geolocation=()",
	}
}

// HMACValidator handles HMAC-based request validation
type HMACValidator struct {
	secret []byte
	algo   string
}

// NewHMACValidator creates a new HMAC validator
func NewHMACValidator(secret string) *HMACValidator {
	return &HMACValidator{
		secret: []byte(secret),
		algo:   "sha256",
	}
}

// GenerateSignature generates an HMAC signature for data
func (hv *HMACValidator) GenerateSignature(data string) string {
	h := hmac.New(sha256.New, hv.secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies an HMAC signature
func (hv *HMACValidator) VerifySignature(data, signature string) bool {
	expected := hv.GenerateSignature(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// PasswordValidator performs comprehensive password validation
type PasswordValidator struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigits  bool
	RequireSpecial bool
	ForbiddenWords []string
}

// NewPasswordValidator creates a new password validator with defaults
func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigits:  true,
		RequireSpecial: true,
		ForbiddenWords: []string{"password", "admin", "root", "test", "user"},
	}
}

// ValidatePassword validates password strength
func (pv *PasswordValidator) ValidatePassword(password string) (bool, []string) {
	var errors []string

	if len(password) < pv.MinLength {
		errors = append(errors, fmt.Sprintf("Password must be at least %d characters", pv.MinLength))
	}

	if pv.RequireUpper && !hasUpperCase(password) {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}

	if pv.RequireLower && !hasLowerCase(password) {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}

	if pv.RequireDigits && !hasDigit(password) {
		errors = append(errors, "Password must contain at least one digit")
	}

	if pv.RequireSpecial && !hasSpecialChar(password) {
		errors = append(errors, "Password must contain at least one special character")
	}

	for _, word := range pv.ForbiddenWords {
		if strings.Contains(strings.ToLower(password), strings.ToLower(word)) {
			errors = append(errors, fmt.Sprintf("Password cannot contain '%s'", word))
		}
	}

	return len(errors) == 0, errors
}

func hasUpperCase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasLowerCase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSpecialChar(s string) bool {
	special := "!@#$%^&*()_+-=[]{}|;:,.<>?/~`"
	for _, r := range s {
		for _, sp := range special {
			if r == sp {
				return true
			}
		}
	}
	return false
}

// CORSValidator validates CORS requests
type CORSValidator struct {
	AllowedOrigins  []string
	AllowedMethods  []string
	AllowedHeaders  []string
	ExposedHeaders  []string
	AllowCredentials bool
	MaxAge          int
}

// NewCORSValidator creates a new CORS validator
func NewCORSValidator() *CORSValidator {
	return &CORSValidator{
		AllowedOrigins:   []string{"http://localhost:7000", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

// ValidateCORSOrigin validates if the origin is allowed
func (cv *CORSValidator) ValidateCORSOrigin(origin string) bool {
	for _, allowed := range cv.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// ValidateCORSMethod validates if the method is allowed
func (cv *CORSValidator) ValidateCORSMethod(method string) bool {
	for _, allowed := range cv.AllowedMethods {
		if allowed == "*" || allowed == method {
			return true
		}
	}
	return false
}

// RateLimitValidator validates rate limit compliance
type RateLimitValidator struct {
	RequestsPerSecond int
	BurstSize        int
}

// NewRateLimitValidator creates a new rate limit validator
func NewRateLimitValidator(rps, burst int) *RateLimitValidator {
	return &RateLimitValidator{
		RequestsPerSecond: rps,
		BurstSize:        burst,
	}
}

// CryptographicHelper provides cryptographic utilities
type CryptographicHelper struct {
	hashAlgo string
}

// NewCryptographicHelper creates a new cryptographic helper
func NewCryptographicHelper() *CryptographicHelper {
	return &CryptographicHelper{
		hashAlgo: "sha256",
	}
}

// GenerateSecureRandom generates cryptographically secure random bytes
func (ch *CryptographicHelper) GenerateSecureRandom(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSecureToken generates a secure token
func (ch *CryptographicHelper) GenerateSecureToken(length int) (string, error) {
	return ch.GenerateSecureRandom(length)
}

// HashData hashes data using SHA256
func (ch *CryptographicHelper) HashData(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HashDataWithSalt hashes data with salt using SHA256
func (ch *CryptographicHelper) HashDataWithSalt(data, salt string) string {
	h := sha256.New()
	h.Write([]byte(data + salt))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHash verifies a hash matches the data
func (ch *CryptographicHelper) VerifyHash(data, hash string) bool {
	return ch.HashData(data) == hash
}

// ComputeChecksum computes a checksum for data
func (ch *CryptographicHelper) ComputeChecksum(data string) string {
	h := sha512.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// InputSanitizer sanitizes user input
type InputSanitizer struct {
	maxLength int
}

// NewInputSanitizer creates a new input sanitizer
func NewInputSanitizer(maxLength int) *InputSanitizer {
	if maxLength <= 0 {
		maxLength = 1000
	}
	return &InputSanitizer{maxLength: maxLength}
}

// SanitizeURL sanitizes and validates a URL
func (is *InputSanitizer) SanitizeURL(rawURL string) (string, error) {
	if len(rawURL) > is.maxLength {
		return "", fmt.Errorf("URL exceeds maximum length of %d", is.maxLength)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https schemes
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme: %s", u.Scheme)
	}

	return u.String(), nil
}

// SanitizeString removes potentially dangerous characters
func (is *InputSanitizer) SanitizeString(s string) string {
	if len(s) > is.maxLength {
		s = s[:is.maxLength]
	}

	// Remove control characters
	filtered := ""
	for _, r := range s {
		if r >= 32 && r != 127 {
			filtered += string(r)
		}
	}

	return filtered
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	CertFile         string
	KeyFile          string
	MinVersion       string
	PreferServerCipherSuites bool
	InsecureSkipVerify bool
}

// DefaultTLSConfig returns default TLS configuration
func DefaultTLSConfig() TLSConfig {
	return TLSConfig{
		MinVersion:      "1.2",
		PreferServerCipherSuites: true,
		InsecureSkipVerify: false,
	}
}

// AuditLogger logs security-related events
type AuditLogger struct {
	events []SecurityEvent
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Timestamp string
	EventType string
	UserID    string
	Resource  string
	Action    string
	Result    string
	IPAddress string
	Details   map[string]interface{}
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		events: make([]SecurityEvent, 0),
	}
}

// LogEvent logs a security event
func (al *AuditLogger) LogEvent(event SecurityEvent) {
	al.events = append(al.events, event)
}

// GetEvents returns all logged events
func (al *AuditLogger) GetEvents() []SecurityEvent {
	return al.events
}

// HashFunction returns the appropriate hash function
func HashFunction(algo string) hash.Hash {
	switch algo {
	case "sha512":
		return sha512.New()
	case "sha256":
		fallthrough
	default:
		return sha256.New()
	}
}
