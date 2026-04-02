package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// OAuthClient represents a registered OAuth2 client application.
type OAuthClient struct {
	ID           string    `json:"id"`
	Secret       string    `json:"secret,omitempty"` // bcrypt hash stored; raw returned at creation only
	Name         string    `json:"name"`
	RedirectURIs []string  `json:"redirect_uris"`
	Scopes       []string  `json:"scopes"`
	GrantTypes   []string  `json:"grant_types"` // authorization_code, refresh_token, client_credentials
	ServiceRoles []string  `json:"service_roles,omitempty"`
	Public       bool      `json:"public"` // public clients (SPA, mobile)
	CreatedAt    time.Time `json:"created_at"`
	Active       bool      `json:"active"`
}

// AuthorizationCode is a short-lived code exchanged for tokens (Authorization Code flow).
type AuthorizationCode struct {
	Code                string    `json:"code"`
	ClientID            string    `json:"client_id"`
	UserID              string    `json:"user_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method"` // S256
	ExpiresAt           time.Time `json:"expires_at"`
	Used                bool      `json:"used"`
}

// RefreshTokenRecord persists issued refresh tokens for rotation/revocation.
type RefreshTokenRecord struct {
	ID        string    `json:"id"` // jti
	UserID    string    `json:"user_id"`
	ClientID  string    `json:"client_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// AuthorizeRequest represents the /oauth/authorize query.
type AuthorizeRequest struct {
	ResponseType        string `form:"response_type" binding:"required"`
	ClientID            string `form:"client_id" binding:"required"`
	RedirectURI         string `form:"redirect_uri" binding:"required"`
	Scope               string `form:"scope"`
	State               string `form:"state"`
	CodeChallenge       string `form:"code_challenge" binding:"required"`
	CodeChallengeMethod string `form:"code_challenge_method" binding:"required"`
}

// TokenRequest represents the /oauth/token body.
type TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type" binding:"required"`
	Code         string `json:"code" form:"code"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	CodeVerifier string `json:"code_verifier" form:"code_verifier"`
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
	Scope        string `json:"scope" form:"scope"`
}

// ClientRepository manages OAuth client persistence.
type ClientRepository interface {
	GetClient(clientID string) (*OAuthClient, error)
	CreateClient(client *OAuthClient) error
	UpdateClient(client *OAuthClient) error
	DeleteClient(clientID string) error
	ListClients() ([]*OAuthClient, error)
}

// CodeRepository manages authorization codes (short TTL).
type CodeRepository interface {
	StoreCode(code *AuthorizationCode) error
	GetCode(code string) (*AuthorizationCode, error)
	InvalidateCode(code string) error
}

// RefreshTokenRepository manages refresh token state.
type RefreshTokenRepository interface {
	StoreRefreshToken(rt *RefreshTokenRecord) error
	GetRefreshToken(jti string) (*RefreshTokenRecord, error)
	RevokeRefreshToken(jti string) error
	RevokeAllForUser(userID string) error
}

const (
	codeLength = 32 // bytes → 64-char hex
	codeTTL    = 5 * time.Minute
)

// GenerateAuthorizationCode creates a cryptographically secure code.
func GenerateAuthorizationCode() (string, error) {
	buf := make([]byte, codeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// GenerateClientSecret creates a random client secret.
func GenerateClientSecret() string {
	buf := make([]byte, 32)
	rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

// GenerateClientID creates a random client ID.
func GenerateClientID() string {
	return uuid.New().String()
}

// ValidateRedirectURI checks that the redirect URI is registered and uses HTTPS (or localhost).
func ValidateRedirectURI(registered []string, requested string) error {
	if requested == "" {
		return errors.New("redirect_uri is required")
	}
	parsed, err := url.Parse(requested)
	if err != nil {
		return fmt.Errorf("invalid redirect_uri: %w", err)
	}
	// Require HTTPS in production; allow http for localhost
	if parsed.Scheme != "https" {
		host := parsed.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			return errors.New("redirect_uri must use HTTPS")
		}
	}
	// Strict exact match against registered URIs (no wildcards, no subpath matching)
	for _, r := range registered {
		if r == requested {
			return nil
		}
	}
	return errors.New("redirect_uri not registered for this client")
}

// ValidateScopes ensures requested scopes are allowed for the client.
func ValidateScopes(allowed, requested []string) error {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, s := range allowed {
		allowedSet[s] = struct{}{}
	}
	for _, s := range requested {
		if _, ok := allowedSet[s]; !ok {
			return fmt.Errorf("scope %q not allowed for this client", s)
		}
	}
	return nil
}

// ParseScopes splits a space-delimited scope string.
func ParseScopes(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Fields(raw)
	seen := make(map[string]struct{}, len(parts))
	unique := make([]string, 0, len(parts))
	for _, p := range parts {
		if _, dup := seen[p]; !dup {
			seen[p] = struct{}{}
			unique = append(unique, p)
		}
	}
	return unique
}

// VerifyPKCE validates the code_verifier against the stored challenge.
func VerifyPKCE(challenge, verifier, method string) error {
	if method != "S256" {
		return errors.New("unsupported code_challenge_method; only S256 is allowed")
	}
	if verifier == "" {
		return errors.New("code_verifier is required")
	}
	h := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	if computed != challenge {
		return errors.New("PKCE verification failed")
	}
	return nil
}
