package backupcodes

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	// CodeLength is the number of digits in a backup code.
	CodeLength = 8
	// CodeCharset is the character set for backup codes.
	CodeCharset = "0123456789"
)

// GenerateCodes creates cryptographically secure backup codes.
func GenerateCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = generateSingleCode(CodeLength)
	}
	return codes
}

// GenerateFormattedCodes creates backup codes with dash formatting (XXXX-XXXX).
func GenerateFormattedCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code := generateSingleCode(CodeLength)
		codes[i] = fmt.Sprintf("%s-%s", code[:4], code[4:])
	}
	return codes
}

func generateSingleCode(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(CodeCharset))))
		if err != nil {
			// Fallback to index 0 on error (should never happen)
			bytes[i] = CodeCharset[0]
			continue
		}
		bytes[i] = CodeCharset[n.Int64()]
	}
	return string(bytes)
}
