package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims from Keycloak
type Claims struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	jwt.RegisteredClaims
}

// KeycloakConfig holds Keycloak configuration
type KeycloakConfig struct {
	ServerURL string
	Realm     string
	ClientID  string
}

// JWKS represents the JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// TokenValidator validates JWT tokens from Keycloak
type TokenValidator struct {
	config       *KeycloakConfig
	publicKeys   map[string]*rsa.PublicKey
	publicKeysMu sync.RWMutex
	lastFetch    time.Time
}

// NewTokenValidator creates a new token validator
func NewTokenValidator(config *KeycloakConfig) (*TokenValidator, error) {
	tv := &TokenValidator{
		config:     config,
		publicKeys: make(map[string]*rsa.PublicKey),
	}

	// Fetch JWKS from Keycloak
	if err := tv.refreshPublicKeys(); err != nil {
		return nil, err
	}

	return tv, nil
}

// refreshPublicKeys fetches the public keys from Keycloak
func (tv *TokenValidator) refreshPublicKeys() error {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs",
		tv.config.ServerURL, tv.config.Realm)

	resp, err := http.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	tv.publicKeysMu.Lock()
	defer tv.publicKeysMu.Unlock()

	tv.publicKeys = make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty == "RSA" {
			pubKey, err := decodeRSAPublicKey(key)
			if err != nil {
				log.Printf("⚠️  Failed to decode RSA key %s: %v", key.Kid, err)
				continue
			}
			tv.publicKeys[key.Kid] = pubKey
		}
	}

	tv.lastFetch = time.Now()
	log.Printf("✅ Loaded %d public keys from Keycloak", len(tv.publicKeys))
	return nil
}

// ValidateToken validates a JWT token
func (tv *TokenValidator) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Get the key ID from the token header
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid not found in token header")
			}

			tv.publicKeysMu.RLock()
			pubKey, exists := tv.publicKeys[kid]
			tv.publicKeysMu.RUnlock()

			if !exists {
				// Try refreshing keys if key not found
				if err := tv.refreshPublicKeys(); err != nil {
					return nil, fmt.Errorf("failed to refresh public keys: %w", err)
				}

				tv.publicKeysMu.RLock()
				pubKey, exists = tv.publicKeys[kid]
				tv.publicKeysMu.RUnlock()

				if !exists {
					return nil, fmt.Errorf("key not found in JWKS: %s", kid)
				}
			}

			return pubKey, nil
		})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check token expiration
	if claims.ExpiresAt != nil && time.Now().Unix() > claims.ExpiresAt.Unix() {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// ExtractBearerToken extracts the JWT token from Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}

// decodeRSAPublicKey decodes a JWK to RSA public key
func decodeRSAPublicKey(key JWK) (*rsa.PublicKey, error) {
	// Decode base64url encoded values
	nBytes, err := decodeBase64URL(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	eBytes, err := decodeBase64URL(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	// Convert bytes to big integers
	n := bytesToBigInt(nBytes)
	e := bytesToInt(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

// decodeBase64URL decodes base64url encoded string
func decodeBase64URL(str string) ([]byte, error) {
	// Add padding if needed
	switch len(str) % 4 {
	case 2:
		str += "=="
	case 3:
		str += "="
	}

	// Replace URL-safe characters
	str = strings.ReplaceAll(str, "-", "+")
	str = strings.ReplaceAll(str, "_", "/")

	return decodeBase64Standard(str)
}

// decodeBase64Standard decodes standard base64
func decodeBase64Standard(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

// bytesToBigInt converts bytes to big.Int
func bytesToBigInt(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}

// bytesToInt converts bytes to int
func bytesToInt(b []byte) int {
	result := 0
	for _, byte := range b {
		result = (result << 8) | int(byte)
	}
	return result
}
