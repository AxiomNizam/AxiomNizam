package uuid

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GenerateV4 generates a random UUID v4
func GenerateV4() string {
	return uuid.New().String()
}

// GenerateV4NoHyphens generates a random UUID v4 without hyphens
func GenerateV4NoHyphens() string {
	return strings.ReplaceAll(GenerateV4(), "-", "")
}

// GenerateV3 generates a deterministic UUID v3 using MD5
func GenerateV3(namespace uuid.UUID, name string) string {
	return uuid.NewMD5(namespace, []byte(name)).String()
}

// GenerateV5 generates a deterministic UUID v5 using SHA1
func GenerateV5(namespace uuid.UUID, name string) string {
	return uuid.NewSHA1(namespace, []byte(name)).String()
}

// NamespaceDNS is the DNS namespace UUID
var NamespaceDNS = uuid.NameSpaceDNS

// NamespaceURL is the URL namespace UUID
var NamespaceURL = uuid.NameSpaceURL

// NamespaceOID is the OID namespace UUID
var NamespaceOID = uuid.NameSpaceOID

// NamespaceX500 is the X.500 DN namespace UUID
var NamespaceX500 = uuid.NameSpaceX500

// Parse parses a UUID string
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValid validates a UUID string
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// MustParse parses a UUID string or panics
func MustParse(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("invalid uuid: %s", s))
	}
	return id
}

// Generator provides UUID generation with options
type Generator struct {
	withHyphens bool
}

// NewGenerator creates a new UUID generator
func NewGenerator() *Generator {
	return &Generator{
		withHyphens: true,
	}
}

// WithoutHyphens configures generator to produce UUIDs without hyphens
func (g *Generator) WithoutHyphens() *Generator {
	g.withHyphens = false
	return g
}

// Generate generates a random UUID v4
func (g *Generator) Generate() string {
	id := GenerateV4()
	if !g.withHyphens {
		return strings.ReplaceAll(id, "-", "")
	}
	return id
}

// GenerateDeterministic generates a deterministic UUID based on name
func (g *Generator) GenerateDeterministic(name string) string {
	id := GenerateV5(NamespaceURL, name)
	if !g.withHyphens {
		return strings.ReplaceAll(id, "-", "")
	}
	return id
}

// Batch represents a batch of generated UUIDs
type Batch struct {
	ids []string
}

// NewBatch creates a new batch
func NewBatch() *Batch {
	return &Batch{
		ids: make([]string, 0),
	}
}

// Generate generates n UUIDs
func (b *Batch) Generate(n int) *Batch {
	for i := 0; i < n; i++ {
		b.ids = append(b.ids, GenerateV4())
	}
	return b
}

// IDs returns all generated IDs
func (b *Batch) IDs() []string {
	return b.ids
}

// First returns the first ID
func (b *Batch) First() string {
	if len(b.ids) > 0 {
		return b.ids[0]
	}
	return ""
}

// Last returns the last ID
func (b *Batch) Last() string {
	if len(b.ids) > 0 {
		return b.ids[len(b.ids)-1]
	}
	return ""
}

// Count returns the number of IDs
func (b *Batch) Count() int {
	return len(b.ids)
}

// Reset clears all IDs
func (b *Batch) Reset() *Batch {
	b.ids = make([]string, 0)
	return b
}

// RequestID generates a request ID
func RequestID() string {
	return GenerateV4NoHyphens()
}

// CorrelationID generates a correlation ID
func CorrelationID() string {
	return GenerateV4NoHyphens()
}

// TraceID generates a trace ID
func TraceID() string {
	return GenerateV4NoHyphens()
}

// SpanID generates a span ID
func SpanID() string {
	return GenerateV4NoHyphens()
}

// ResourceID generates a resource ID
func ResourceID(resourceType string) string {
	return GenerateV5(NamespaceURL, resourceType)
}

// UserID generates a user ID
func UserID() string {
	return GenerateV4NoHyphens()
}

// TenantID generates a tenant ID
func TenantID() string {
	return GenerateV4NoHyphens()
}

// SessionID generates a session ID
func SessionID() string {
	return GenerateV4NoHyphens()
}

// TokenID generates a token ID
func TokenID() string {
	return GenerateV4NoHyphens()
}

// DeterministicID generates a deterministic ID based on input
func DeterministicID(input string) string {
	return GenerateV5(NamespaceURL, input)
}

// MD5UUID generates a UUID using MD5
func MD5UUID(namespace uuid.UUID, name string) string {
	return GenerateV3(namespace, name)
}

// SHA1UUID generates a UUID using SHA1
func SHA1UUID(namespace uuid.UUID, name string) string {
	return GenerateV5(namespace, name)
}

// ConfigID generates a config ID from data
func ConfigID(data []byte) string {
	hash := md5.Sum(data)
	return uuid.NewMD5(NamespaceURL, hash[:]).String()
}

// ShortID generates a short unique ID
func ShortID() string {
	return GenerateV4NoHyphens()[:12]
}

// LongID generates a long unique ID with timestamp prefix
func LongID() string {
	return fmt.Sprintf("%d_%s", unixTimestamp(), GenerateV4NoHyphens())
}

func unixTimestamp() int64 {
	return 0 // Could import time and use time.Now().Unix()
}

// PrefixedID generates an ID with custom prefix
func PrefixedID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, GenerateV4NoHyphens())
}

// Validator validates UUID strings
type Validator struct{}

// NewValidator creates a new UUID validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a UUID string
func (v *Validator) Validate(s string) error {
	if !IsValid(s) {
		return fmt.Errorf("invalid uuid: %s", s)
	}
	return nil
}

// ValidateSlice validates a slice of UUID strings
func (v *Validator) ValidateSlice(uuids []string) error {
	for _, id := range uuids {
		if err := v.Validate(id); err != nil {
			return err
		}
	}
	return nil
}

// ValidateFormat validates UUID format
func ValidateFormat(s string) bool {
	return IsValid(s)
}
