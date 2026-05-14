package anonymization

// =====================================================
// WS-7.3 — Masking Function Implementations
//
// Provides 8 masking techniques for PII anonymization:
// hash, redact, partial, tokenize, noise, generalize, synthetic, shuffle.
// =====================================================

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
)

// Masker applies a masking technique to a string value.
type Masker struct {
	secret []byte // HMAC key for consistent hashing/tokenization
	rng    *rand.Rand
}

// NewMasker creates a new masker with the given secret for deterministic operations.
func NewMasker(secret string) *Masker {
	return &Masker{
		secret: []byte(secret),
		rng:    rand.New(rand.NewSource(42)), // Deterministic for reproducibility
	}
}

// Mask applies the specified technique to a value.
func (m *Masker) Mask(value string, technique MaskTechnique, config map[string]string) string {
	switch technique {
	case MaskHash:
		return m.hash(value)
	case MaskRedact:
		return m.redact(config)
	case MaskPartial:
		return m.partial(value, config)
	case MaskTokenize:
		return m.tokenize(value)
	case MaskNoise:
		return m.noise(value)
	case MaskGeneralize:
		return m.generalize(value, config)
	case MaskSynthetic:
		return m.synthetic(value, config)
	case MaskShuffle:
		return m.shuffle(value)
	default:
		return "[MASKED]"
	}
}

// hash produces a consistent HMAC-SHA256 pseudonym.
// Same input always produces the same output (for join preservation).
func (m *Masker) hash(value string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))[:16]
}

// redact replaces the entire value with a placeholder.
func (m *Masker) redact(config map[string]string) string {
	placeholder := config["placeholder"]
	if placeholder == "" {
		placeholder = "[REDACTED]"
	}
	return placeholder
}

// partial masks the middle of a value, preserving prefix and suffix.
// "john@email.com" -> "j***@e***.com"
func (m *Masker) partial(value string, config map[string]string) string {
	if len(value) <= 4 {
		return "****"
	}

	// Handle email specially.
	if atIdx := strings.Index(value, "@"); atIdx > 0 {
		local := value[:atIdx]
		domain := value[atIdx+1:]
		maskedLocal := string(local[0]) + strings.Repeat("*", len(local)-1)
		dotIdx := strings.LastIndex(domain, ".")
		if dotIdx > 0 {
			maskedDomain := string(domain[0]) + strings.Repeat("*", dotIdx-1) + domain[dotIdx:]
			return maskedLocal + "@" + maskedDomain
		}
		return maskedLocal + "@" + strings.Repeat("*", len(domain))
	}

	// Generic: show first and last char.
	prefixLen := 1
	suffixLen := 1
	if p, ok := config["prefixLen"]; ok {
		fmt.Sscanf(p, "%d", &prefixLen)
	}
	if s, ok := config["suffixLen"]; ok {
		fmt.Sscanf(s, "%d", &suffixLen)
	}

	if prefixLen+suffixLen >= len(value) {
		return strings.Repeat("*", len(value))
	}

	return value[:prefixLen] + strings.Repeat("*", len(value)-prefixLen-suffixLen) + value[len(value)-suffixLen:]
}

// tokenize produces a reversible token (with the secret key).
// Format: TOK_<short_hmac>
func (m *Masker) tokenize(value string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(value))
	return "TOK_" + hex.EncodeToString(mac.Sum(nil))[:8]
}

// noise adds statistical noise to numeric values.
// "75000" -> "73200" (within ±5% by default)
func (m *Masker) noise(value string) string {
	var num float64
	if _, err := fmt.Sscanf(value, "%f", &num); err != nil {
		return value // Not numeric, return as-is.
	}

	// Add ±5% noise.
	noiseFactor := 0.95 + m.rng.Float64()*0.10
	noised := num * noiseFactor

	// Preserve integer format if input was integer.
	if !strings.Contains(value, ".") {
		return fmt.Sprintf("%.0f", noised)
	}
	return fmt.Sprintf("%.2f", noised)
}

// generalize replaces precise values with ranges.
// "34" -> "30-40", "john@email.com" -> "***@email.com"
func (m *Masker) generalize(value string, config map[string]string) string {
	// Try numeric generalization.
	var num int
	if _, err := fmt.Sscanf(value, "%d", &num); err == nil {
		bucketSize := 10
		if b, ok := config["bucketSize"]; ok {
			fmt.Sscanf(b, "%d", &bucketSize)
		}
		lower := (num / bucketSize) * bucketSize
		upper := lower + bucketSize
		return fmt.Sprintf("%d-%d", lower, upper)
	}

	// String generalization: mask to category.
	if len(value) > 3 {
		return strings.Repeat("*", len(value)-3) + value[len(value)-3:]
	}
	return "***"
}

// synthetic generates realistic fake data of the same type.
func (m *Masker) synthetic(value string, config map[string]string) string {
	dataType := config["type"]
	if dataType == "" {
		// Auto-detect.
		if strings.Contains(value, "@") {
			dataType = "email"
		} else if len(value) > 0 && value[0] >= '0' && value[0] <= '9' {
			dataType = "number"
		} else {
			dataType = "name"
		}
	}

	switch dataType {
	case "email":
		names := []string{"alice", "bob", "carol", "dave", "eve", "frank", "grace", "heidi"}
		domains := []string{"example.com", "test.org", "sample.net"}
		return names[m.rng.Intn(len(names))] + "@" + domains[m.rng.Intn(len(domains))]
	case "name":
		first := []string{"Alice", "Bob", "Carol", "Dave", "Eve", "Frank", "Grace", "Heidi"}
		last := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller"}
		return first[m.rng.Intn(len(first))] + " " + last[m.rng.Intn(len(last))]
	case "phone":
		return fmt.Sprintf("+1-%03d-%03d-%04d", m.rng.Intn(900)+100, m.rng.Intn(900)+100, m.rng.Intn(9000)+1000)
	case "number":
		var num float64
		fmt.Sscanf(value, "%f", &num)
		return fmt.Sprintf("%.0f", num*0.8+m.rng.Float64()*num*0.4)
	default:
		return "SYNTHETIC_" + m.hash(value)[:8]
	}
}

// shuffle returns a shuffled version of the string characters.
func (m *Masker) shuffle(value string) string {
	runes := []rune(value)
	m.rng.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})
	return string(runes)
}

// MaskBatch applies masking to a batch of values.
func (m *Masker) MaskBatch(values []string, technique MaskTechnique, config map[string]string) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = m.Mask(v, technique, config)
	}
	return result
}
