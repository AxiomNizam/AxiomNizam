// Package uuid generates RFC 4122 version-4 UUIDs using crypto/rand.
// AxiomNizam uses UUIDs as the stable identity of every resource —
// the name may be reused after deletion, the UID may not.
//
// The implementation is deliberately tiny (no external deps) because
// the upstream google/uuid and k8s.io/apimachinery/pkg/util/uuid
// packages weigh too much for a primitive every resource-create path
// touches.
package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// NewUUID returns a random v4 UUID in canonical 8-4-4-4-12 hex form.
// It panics on crypto/rand failure because the alternative is to
// return the zero UUID, which would collapse every resource's
// identity to the same value — a far worse failure mode.
func NewUUID() string {
	var b [16]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		panic(fmt.Sprintf("uuid: crypto/rand failure: %v", err))
	}
	// Set version (4) and variant (10xx) bits per RFC 4122 §4.4.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return formatUUID(b)
}

// formatUUID renders the 16-byte buffer in canonical form.
func formatUUID(b [16]byte) string {
	// Allocate once with exact length — 32 hex digits + 4 hyphens.
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf)
}
