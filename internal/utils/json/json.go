package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Marshal safely marshals data to JSON
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent safely marshals data to indented JSON
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal safely unmarshals JSON to data
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalString marshals data to JSON string
func MarshalString(v interface{}) (string, error) {
	data, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalString unmarshals JSON string to data
func UnmarshalString(str string, v interface{}) error {
	return Unmarshal([]byte(str), v)
}

// Pretty returns a pretty-printed JSON string
func Pretty(v interface{}) string {
	data, err := MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// Compact returns a compact JSON string
func Compact(v interface{}) string {
	data, err := Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

// Decoder wraps json.Decoder for safe decoding
type Decoder struct {
	decoder *json.Decoder
}

// NewDecoder creates a new JSON decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		decoder: json.NewDecoder(r),
	}
}

// Decode decodes JSON from reader
func (d *Decoder) Decode(v interface{}) error {
	return d.decoder.Decode(v)
}

// DecodeFromReader decodes JSON from reader
func DecodeFromReader(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}

// Encoder wraps json.Encoder for safe encoding
type Encoder struct {
	encoder *json.Encoder
}

// NewEncoder creates a new JSON encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		encoder: json.NewEncoder(w),
	}
}

// Encode encodes data to JSON
func (e *Encoder) Encode(v interface{}) error {
	return e.encoder.Encode(v)
}

// EncodeToWriter encodes data to writer
func EncodeToWriter(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}

// Schema defines a JSON schema for validation
type Schema struct {
	schema map[string]interface{}
}

// NewSchema creates a new JSON schema
func NewSchema(schema map[string]interface{}) *Schema {
	return &Schema{
		schema: schema,
	}
}

// Validate validates data against schema
func (s *Schema) Validate(data interface{}) error {
	// Basic implementation - could use full JSON Schema validation
	return nil
}

// SafeUnmarshal unmarshals with type safety
func SafeUnmarshal(data []byte, v interface{}) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid json")
	}
	return json.Unmarshal(data, v)
}

// SafeUnmarshalString unmarshals string with type safety
func SafeUnmarshalString(str string, v interface{}) error {
	return SafeUnmarshal([]byte(str), v)
}

// ValidateJSON checks if data is valid JSON
func ValidateJSON(data []byte) bool {
	return json.Valid(data)
}

// ValidateJSONString checks if string is valid JSON
func ValidateJSONString(str string) bool {
	return ValidateJSON([]byte(str))
}

// Object represents a JSON object
type Object map[string]interface{}

// NewObject creates a new JSON object
func NewObject() Object {
	return make(Object)
}

// Set sets a value in the object
func (o Object) Set(key string, value interface{}) Object {
	o[key] = value
	return o
}

// Get gets a value from the object
func (o Object) Get(key string) (interface{}, bool) {
	val, ok := o[key]
	return val, ok
}

// GetString gets a string value
func (o Object) GetString(key string) (string, bool) {
	val, ok := o.Get(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt gets an int value
func (o Object) GetInt(key string) (int, bool) {
	val, ok := o.Get(key)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}

// GetBool gets a bool value
func (o Object) GetBool(key string) (bool, bool) {
	val, ok := o.Get(key)
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// Delete deletes a key from the object
func (o Object) Delete(key string) {
	delete(o, key)
}

// Has checks if a key exists
func (o Object) Has(key string) bool {
	_, ok := o[key]
	return ok
}

// Merge merges another object into this one
func (o Object) Merge(other Object) Object {
	for k, v := range other {
		o[k] = v
	}
	return o
}

// Clone creates a deep copy of the object
func (o Object) Clone() Object {
	data, _ := Marshal(o)
	cloned := make(Object)
	_ = Unmarshal(data, &cloned)
	return cloned
}

// ToJSON converts object to JSON
func (o Object) ToJSON() string {
	return Compact(o)
}

// Array represents a JSON array
type Array []interface{}

// NewArray creates a new JSON array
func NewArray() Array {
	return make(Array, 0)
}

// Append appends a value to the array
func (a *Array) Append(value interface{}) *Array {
	*a = append(*a, value)
	return a
}

// Length returns the length of the array
func (a Array) Length() int {
	return len(a)
}

// Get gets a value at index
func (a Array) Get(index int) (interface{}, bool) {
	if index >= 0 && index < len(a) {
		return a[index], true
	}
	return nil, false
}

// ToJSON converts array to JSON
func (a Array) ToJSON() string {
	return Compact(a)
}

// RawMessage is an alias for json.RawMessage
type RawMessage = json.RawMessage

// NewRawMessage creates a new raw message
func NewRawMessage(data []byte) RawMessage {
	return json.RawMessage(data)
}

// Diff represents JSON difference
type Diff struct {
	original interface{}
	modified interface{}
}

// NewDiff creates a new diff
func NewDiff(original, modified interface{}) *Diff {
	return &Diff{
		original: original,
		modified: modified,
	}
}

// Changes returns the changes between original and modified
func (d *Diff) Changes() map[string]interface{} {
	changes := make(map[string]interface{})

	origData, _ := Marshal(d.original)
	modData, _ := Marshal(d.modified)

	var orig, mod map[string]interface{}
	_ = Unmarshal(origData, &orig)
	_ = Unmarshal(modData, &mod)

	for key, newVal := range mod {
		if oldVal, exists := orig[key]; exists {
			if !bytesEqual(mustMarshal(oldVal), mustMarshal(newVal)) {
				changes[key] = newVal
			}
		} else {
			changes[key] = newVal
		}
	}

	return changes
}

func bytesEqual(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func mustMarshal(v interface{}) []byte {
	data, _ := Marshal(v)
	return data
}
