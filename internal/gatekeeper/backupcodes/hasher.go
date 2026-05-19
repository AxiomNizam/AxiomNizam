package backupcodes

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// hashBackupCode creates a SHA-256 hash of the normalized backup code.
// The code is normalized by removing dashes and converting to lowercase.
func hashBackupCode(code string) []byte {
	normalized := strings.ToLower(strings.ReplaceAll(code, "-", ""))
	hash := sha256.Sum256([]byte(normalized))
	return hash[:]
}

// hashBackupCodeHex returns the hex-encoded hash of a backup code.
func hashBackupCodeHex(code string) string {
	return hex.EncodeToString(hashBackupCode(code))
}
