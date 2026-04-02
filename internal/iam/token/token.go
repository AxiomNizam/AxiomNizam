package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// IAMClaims are the JWT claims issued by the IAM system.
type IAMClaims struct {
	Sub         string   `json:"sub"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Scope       string   `json:"scope,omitempty"`
	ClientID    string   `json:"client_id,omitempty"`
	SessionID   string   `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair holds both access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// AccessTokenResponse is used for client_credentials and other access-token-only flows.
type AccessTokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Scope       string    `json:"scope,omitempty"`
}

// JWKSResponse is the /.well-known/jwks.json payload.
type JWKSResponse struct {
	Keys []JWKEntry `json:"keys"`
}

// JWKEntry represents a single JWK public key.
type JWKEntry struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// Issuer handles JWT signing, verification and JWKS exposure.
type Issuer struct {
	mu         sync.RWMutex
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	kid        string
	issuerURL  string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

const (
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 24 * time.Hour * 7 // 7 days
	rsaKeyBits        = 2048
)

// NewIssuer creates a token issuer. It loads an RSA key from the environment
// or generates an ephemeral one for development.
func NewIssuer(issuerURL string) (*Issuer, error) {
	iss := &Issuer{
		issuerURL:       issuerURL,
		AccessTokenTTL:  defaultAccessTTL,
		RefreshTokenTTL: defaultRefreshTTL,
	}

	if keyPEM := os.Getenv("IAM_RSA_PRIVATE_KEY"); keyPEM != "" {
		block, _ := pem.Decode([]byte(keyPEM))
		if block == nil {
			return nil, errors.New("IAM_RSA_PRIVATE_KEY: invalid PEM")
		}
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("IAM_RSA_PRIVATE_KEY: %w", err)
		}
		iss.privateKey = priv
		iss.publicKey = &priv.PublicKey
	} else if keyPath := os.Getenv("IAM_RSA_PRIVATE_KEY_FILE"); keyPath != "" {
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("IAM_RSA_PRIVATE_KEY_FILE: %w", err)
		}
		block, _ := pem.Decode(data)
		if block == nil {
			return nil, errors.New("IAM_RSA_PRIVATE_KEY_FILE: invalid PEM")
		}
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("IAM_RSA_PRIVATE_KEY_FILE: %w", err)
		}
		iss.privateKey = priv
		iss.publicKey = &priv.PublicKey
	} else {
		// Ephemeral key for development
		priv, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		iss.privateKey = priv
		iss.publicKey = &priv.PublicKey
	}

	iss.kid = uuid.New().String()[:8]
	return iss, nil
}

// IssueTokenPair creates a signed access and refresh token pair.
func (iss *Issuer) IssueTokenPair(sub, email, displayName, scope, clientID, sessionID string, roles []string) (*TokenPair, error) {
	now := time.Now().UTC()
	accessExp := now.Add(iss.AccessTokenTTL)
	refreshExp := now.Add(iss.RefreshTokenTTL)

	accessClaims := IAMClaims{
		Sub:         sub,
		Email:       email,
		DisplayName: displayName,
		Roles:       roles,
		Scope:       scope,
		ClientID:    clientID,
		SessionID:   sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    iss.issuerURL,
			Subject:   sub,
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessToken.Header["kid"] = iss.kid

	signedAccess, err := accessToken.SignedString(iss.privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshClaims := jwt.RegisteredClaims{
		Issuer:    iss.issuerURL,
		Subject:   sub,
		Audience:  jwt.ClaimStrings{clientID},
		ExpiresAt: jwt.NewNumericDate(refreshExp),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshToken.Header["kid"] = iss.kid

	signedRefresh, err := refreshToken.SignedString(iss.privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    int(iss.AccessTokenTTL.Seconds()),
		ExpiresAt:    accessExp,
		Scope:        scope,
	}, nil
}

// IssueAccessToken creates a signed access token without issuing a refresh token.
func (iss *Issuer) IssueAccessToken(sub, email, displayName, scope, clientID, sessionID string, roles []string) (*AccessTokenResponse, error) {
	now := time.Now().UTC()
	accessExp := now.Add(iss.AccessTokenTTL)

	accessClaims := IAMClaims{
		Sub:         sub,
		Email:       email,
		DisplayName: displayName,
		Roles:       roles,
		Scope:       scope,
		ClientID:    clientID,
		SessionID:   sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    iss.issuerURL,
			Subject:   sub,
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessToken.Header["kid"] = iss.kid

	signedAccess, err := accessToken.SignedString(iss.privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	return &AccessTokenResponse{
		AccessToken: signedAccess,
		TokenType:   "Bearer",
		ExpiresIn:   int(iss.AccessTokenTTL.Seconds()),
		ExpiresAt:   accessExp,
		Scope:       scope,
	}, nil
}

// ValidateAccessToken parses and validates an access token.
func (iss *Issuer) ValidateAccessToken(raw string) (*IAMClaims, error) {
	claims := &IAMClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return iss.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ValidateRefreshToken parses a refresh token to extract the subject.
func (iss *Issuer) ValidateRefreshToken(raw string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return iss.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token validation failed: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	return claims, nil
}

// JWKS returns the JSON Web Key Set containing the public key.
func (iss *Issuer) JWKS() *JWKSResponse {
	iss.mu.RLock()
	defer iss.mu.RUnlock()

	return &JWKSResponse{
		Keys: []JWKEntry{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: iss.kid,
				Alg: "RS256",
				N:   base64.RawURLEncoding.EncodeToString(iss.publicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(iss.publicKey.E)).Bytes()),
			},
		},
	}
}

// ServeJWKS is an http.HandlerFunc for /.well-known/jwks.json
func (iss *Issuer) ServeJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(iss.JWKS())
}

// IssuerURL returns the configured issuer base URL.
func (iss *Issuer) IssuerURL() string {
	return iss.issuerURL
}

// OpenIDConfigurationWithEndpoints builds an OIDC discovery document for custom endpoint paths.
func (iss *Issuer) OpenIDConfigurationWithEndpoints(issuerURL, authorizationEndpoint, tokenEndpoint, jwksURI string) map[string]interface{} {
	return map[string]interface{}{
		"issuer":                                issuerURL,
		"authorization_endpoint":                authorizationEndpoint,
		"token_endpoint":                        tokenEndpoint,
		"jwks_uri":                              jwksURI,
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email", "roles"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
		"code_challenge_methods_supported":      []string{"S256"},
	}
}

// OpenIDConfiguration returns the OIDC discovery document.
func (iss *Issuer) OpenIDConfiguration() map[string]interface{} {
	base := strings.TrimRight(iss.issuerURL, "/")
	return iss.OpenIDConfigurationWithEndpoints(
		base,
		base+"/oauth/authorize",
		base+"/oauth/token",
		base+"/.well-known/jwks.json",
	)
}
