package schemaregistry

// =====================================================
// WS-3.1 — Schema Serializer (Protobuf Support)
//
// Parses and validates Protobuf schema definitions alongside
// JSON Schema and Avro. Provides descriptor extraction,
// field enumeration, and compatibility checking for Proto schemas.
// =====================================================

import (
	"fmt"
	"regexp"
	"strings"
)

// SchemaFormat enumerates supported serialization formats.
type SchemaFormat string

const (
	FormatJSON     SchemaFormat = "json"
	FormatAvro     SchemaFormat = "avro"
	FormatProtobuf SchemaFormat = "protobuf"
)

// Serializer handles parsing and introspection for all schema formats.
type Serializer struct{}

// NewSerializer creates a new schema serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// FieldDescriptor represents a parsed field from any schema format.
type FieldDescriptor struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Number   int    `json:"number,omitempty"` // Protobuf field number
	Required bool   `json:"required"`
	Repeated bool   `json:"repeated"`
	Optional bool   `json:"optional"`
	Comment  string `json:"comment,omitempty"`
}

// MessageDescriptor represents a parsed message/record from any schema format.
type MessageDescriptor struct {
	Name     string            `json:"name"`
	Fields   []FieldDescriptor `json:"fields"`
	Package  string            `json:"package,omitempty"`
	Syntax   string            `json:"syntax,omitempty"` // proto2, proto3
	Options  map[string]string `json:"options,omitempty"`
	Nested   []MessageDescriptor `json:"nested,omitempty"`
}

// ParseSchema parses a schema definition based on its format.
func (s *Serializer) ParseSchema(schema string, format SchemaFormat) (*MessageDescriptor, error) {
	switch format {
	case FormatProtobuf:
		return s.parseProtobuf(schema)
	case FormatAvro:
		return s.parseAvro(schema)
	case FormatJSON:
		return s.parseJSONSchema(schema)
	default:
		return nil, fmt.Errorf("serializer: unsupported schema format %q", format)
	}
}

// ExtractFields returns the list of fields from any schema format.
func (s *Serializer) ExtractFields(schema string, format SchemaFormat) ([]FieldDescriptor, error) {
	desc, err := s.ParseSchema(schema, format)
	if err != nil {
		return nil, err
	}
	return desc.Fields, nil
}

// Fingerprint returns a deterministic content hash for a schema.
func (s *Serializer) Fingerprint(schema string, format SchemaFormat) string {
	// Normalize whitespace and compute a simple hash.
	normalized := strings.Join(strings.Fields(schema), " ")
	return fmt.Sprintf("%x", hashString(normalized))
}

// --- Protobuf Parser ---

var (
	protoSyntaxRe  = regexp.MustCompile(`syntax\s*=\s*"(proto[23])"\s*;`)
	protoPackageRe = regexp.MustCompile(`package\s+([\w.]+)\s*;`)
	protoMessageRe = regexp.MustCompile(`message\s+(\w+)\s*\{`)
	protoFieldRe   = regexp.MustCompile(`(?m)^\s*(optional|required|repeated)?\s*([\w.]+)\s+(\w+)\s*=\s*(\d+)\s*;`)
	protoOptionRe  = regexp.MustCompile(`option\s+(\w+)\s*=\s*"?([^";]+)"?\s*;`)
)

func (s *Serializer) parseProtobuf(schema string) (*MessageDescriptor, error) {
	desc := &MessageDescriptor{
		Options: make(map[string]string),
	}

	// Extract syntax.
	if m := protoSyntaxRe.FindStringSubmatch(schema); len(m) > 1 {
		desc.Syntax = m[1]
	}

	// Extract package.
	if m := protoPackageRe.FindStringSubmatch(schema); len(m) > 1 {
		desc.Package = m[1]
	}

	// Extract message name.
	if m := protoMessageRe.FindStringSubmatch(schema); len(m) > 1 {
		desc.Name = m[1]
	} else {
		return nil, fmt.Errorf("protobuf: no message definition found")
	}

	// Extract options.
	for _, m := range protoOptionRe.FindAllStringSubmatch(schema, -1) {
		if len(m) > 2 {
			desc.Options[m[1]] = m[2]
		}
	}

	// Extract fields.
	for _, m := range protoFieldRe.FindAllStringSubmatch(schema, -1) {
		if len(m) < 5 {
			continue
		}
		modifier := m[1]
		fieldType := m[2]
		fieldName := m[3]
		fieldNum := 0
		fmt.Sscanf(m[4], "%d", &fieldNum)

		field := FieldDescriptor{
			Name:     fieldName,
			Type:     fieldType,
			Number:   fieldNum,
			Required: modifier == "required",
			Repeated: modifier == "repeated",
			Optional: modifier == "optional" || (desc.Syntax == "proto3" && modifier == ""),
		}
		desc.Fields = append(desc.Fields, field)
	}

	return desc, nil
}

// --- Avro Parser (simplified) ---

func (s *Serializer) parseAvro(schema string) (*MessageDescriptor, error) {
	// Simplified Avro parsing using regex for field extraction.
	desc := &MessageDescriptor{
		Options: make(map[string]string),
	}

	// Extract name.
	nameRe := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
	if m := nameRe.FindStringSubmatch(schema); len(m) > 1 {
		desc.Name = m[1]
	}

	// Extract namespace.
	nsRe := regexp.MustCompile(`"namespace"\s*:\s*"([^"]+)"`)
	if m := nsRe.FindStringSubmatch(schema); len(m) > 1 {
		desc.Package = m[1]
	}

	// Extract fields.
	fieldRe := regexp.MustCompile(`\{"name"\s*:\s*"([^"]+)"\s*,\s*"type"\s*:\s*"?([^",}\]]+)"?`)
	for _, m := range fieldRe.FindAllStringSubmatch(schema, -1) {
		if len(m) > 2 {
			desc.Fields = append(desc.Fields, FieldDescriptor{
				Name: m[1],
				Type: m[2],
			})
		}
	}

	return desc, nil
}

// --- JSON Schema Parser (simplified) ---

func (s *Serializer) parseJSONSchema(schema string) (*MessageDescriptor, error) {
	desc := &MessageDescriptor{
		Options: make(map[string]string),
	}

	// Extract title as name.
	titleRe := regexp.MustCompile(`"title"\s*:\s*"([^"]+)"`)
	if m := titleRe.FindStringSubmatch(schema); len(m) > 1 {
		desc.Name = m[1]
	}

	// Extract properties.
	propRe := regexp.MustCompile(`"(\w+)"\s*:\s*\{\s*"type"\s*:\s*"([^"]+)"`)
	for _, m := range propRe.FindAllStringSubmatch(schema, -1) {
		if len(m) > 2 && m[1] != "properties" {
			desc.Fields = append(desc.Fields, FieldDescriptor{
				Name: m[1],
				Type: m[2],
			})
		}
	}

	// Check required fields.
	reqRe := regexp.MustCompile(`"required"\s*:\s*\[([^\]]+)\]`)
	if m := reqRe.FindStringSubmatch(schema); len(m) > 1 {
		for _, f := range strings.Split(m[1], ",") {
			name := strings.Trim(strings.TrimSpace(f), `"`)
			for i := range desc.Fields {
				if desc.Fields[i].Name == name {
					desc.Fields[i].Required = true
				}
			}
		}
	}

	return desc, nil
}

// --- Helper ---

func hashString(s string) uint64 {
	var h uint64 = 14695981039346656037
	for _, c := range s {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}
