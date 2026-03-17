package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// RealmAccess contains realm-level roles from Keycloak
type RealmAccess struct {
	Roles []string `json:"roles"`
}

// Claims represents JWT claims from Keycloak
type Claims struct {
	Sub               string                 `json:"sub"`
	PreferredUsername string                 `json:"preferred_username"`
	Email             string                 `json:"email"`
	Name              string                 `json:"name"`
	RealmAccess       RealmAccess            `json:"realm_access"`
	ResourceAccess    map[string]interface{} `json:"resource_access"`
	jwt.RegisteredClaims
}

// HasRole checks if the claims contain a specific role
func (c *Claims) HasRole(role string) bool {
	if c == nil {
		return false
	}
	targetRole := strings.ToLower(strings.TrimSpace(role))
	if targetRole == "" {
		return false
	}

	for _, r := range c.RealmAccess.Roles {
		if strings.ToLower(strings.TrimSpace(r)) == targetRole {
			return true
		}
	}

	// Keycloak often stores app-specific roles under resource_access.<client>.roles.
	for _, clientAccessRaw := range c.ResourceAccess {
		clientAccess, ok := clientAccessRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rolesRaw, ok := clientAccess["roles"]
		if !ok {
			continue
		}

		switch typedRoles := rolesRaw.(type) {
		case []interface{}:
			for _, rr := range typedRoles {
				if roleStr, ok := rr.(string); ok && strings.ToLower(strings.TrimSpace(roleStr)) == targetRole {
					return true
				}
			}
		case []string:
			for _, roleStr := range typedRoles {
				if strings.ToLower(strings.TrimSpace(roleStr)) == targetRole {
					return true
				}
			}
		}
	}

	return false
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

var demoJWTSecret = loadDemoJWTSecret()

func loadDemoJWTSecret() string {
	if secret := strings.TrimSpace(os.Getenv("DEMO_JWT_SECRET")); secret != "" {
		return secret
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err == nil {
		generated := base64.RawURLEncoding.EncodeToString(b)
		log.Printf("⚠️  DEMO_JWT_SECRET is not set, using generated ephemeral demo token secret")
		return generated
	}

	fallback := fmt.Sprintf("ephemeral-demo-secret-%d", time.Now().UnixNano())
	log.Printf("⚠️  DEMO_JWT_SECRET generation failed, falling back to process-ephemeral secret")
	return fallback
}

// DemoJWTSecret returns the HMAC secret used for demo account tokens.
func DemoJWTSecret() string {
	return demoJWTSecret
}

// ValidateDemoToken validates an HMAC-signed demo token
func (tv *TokenValidator) ValidateDemoToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(DemoJWTSecret()), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid demo token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid demo token claims")
	}

	claims := &Claims{}
	if v, ok := mapClaims["preferred_username"].(string); ok {
		claims.PreferredUsername = v
	}
	if v, ok := mapClaims["email"].(string); ok {
		claims.Email = v
	}
	if ra, ok := mapClaims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := ra["roles"].([]interface{}); ok {
			for _, r := range roles {
				if s, ok := r.(string); ok {
					claims.RealmAccess.Roles = append(claims.RealmAccess.Roles, s)
				}
			}
		}
	}
	return claims, nil
}

// ValidateToken validates a JWT token (Keycloak RSA first, demo HMAC fallback)
func (tv *TokenValidator) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Get the key ID from the token header
			kid, ok := token.Header["kid"].(string)
			if !ok {
				// No kid — try demo HMAC validation
				if claims, demoErr := tv.ValidateDemoToken(tokenString); demoErr == nil {
					return nil, fmt.Errorf("demo:ok:%s", claims.PreferredUsername)
				}
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
		// Check if it's a demo token (no kid header)
		if demoClaims, demoErr := tv.ValidateDemoToken(tokenString); demoErr == nil {
			log.Printf("✅ Demo token validated for user: %s (roles: %v)", demoClaims.PreferredUsername, demoClaims.RealmAccess.Roles)
			return demoClaims, nil
		}
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
