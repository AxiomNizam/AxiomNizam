package trusteddevices

import (
	"crypto/rand"
	"encoding/base64"
)

const (
	// TokenLength is the byte length of generated device tokens.
	TokenLength = 32
)

// GenerateDeviceToken creates a cryptographically secure random token.
func GenerateDeviceToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// TokenParts represents the components of a device trust token.
type TokenParts struct {
	DeviceID string
	Token    string
}

// SplitToken splits a "deviceID:token" format string.
func SplitToken(value string) *TokenParts {
	for i := 0; i < len(value); i++ {
		if value[i] == ':' {
			return &TokenParts{
				DeviceID: value[:i],
				Token:    value[i+1:],
			}
		}
	}
	return &TokenParts{Token: value}
}

// JoinToken combines device ID and token into a cookie value.
func JoinToken(deviceID, token string) string {
	return deviceID + ":" + token
}
