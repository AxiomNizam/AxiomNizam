// Package validation — structured field-path errors.
//
// field.Path + field.Error + field.ErrorList mirror the upstream
// k8s.io/apimachinery/pkg/util/validation/field package.  Every
// validator in AxiomNizam should emit ErrorList so responses can be
// concatenated and returned as a single 422 body.
package validation

import (
	"fmt"
	"strconv"
	"strings"
)

// Path builds a dotted reference to a field in a resource.  Immutable
// — Child, Index, and Key return a new Path that prepends to the
// parent chain.
type Path struct {
	parent *Path
	name   string
	index  int
	key    string
	kind   pathKind
}

type pathKind uint8

const (
	kindRoot pathKind = iota
	kindField
	kindIndex
	kindKey
)

// NewPath constructs a root Path.  The first arg is the field name;
// additional args become successive Child calls for brevity.
func NewPath(root string, rest ...string) *Path {
	p := &Path{name: root, kind: kindField}
	for _, r := range rest {
		p = p.Child(r)
	}
	return p
}

// Child returns a subfield reference (e.g. "spec.containers").
func (p *Path) Child(name string) *Path {
	return &Path{parent: p, name: name, kind: kindField}
}

// Index returns a slice-element reference (e.g. "spec.containers[0]").
func (p *Path) Index(i int) *Path {
	return &Path{parent: p, index: i, kind: kindIndex}
}

// Key returns a map-entry reference (e.g. `metadata.labels["app"]`).
func (p *Path) Key(k string) *Path {
	return &Path{parent: p, key: k, kind: kindKey}
}

// String walks back to the root and renders the dotted/bracketed form.
func (p *Path) String() string {
	if p == nil {
		return ""
	}
	var parts []string
	for cur := p; cur != nil; cur = cur.parent {
		switch cur.kind {
		case kindField:
			parts = append([]string{cur.name}, parts...)
		case kindIndex:
			parts = append([]string{"[" + strconv.Itoa(cur.index) + "]"}, parts...)
		case kindKey:
			parts = append([]string{"[" + strconv.Quote(cur.key) + "]"}, parts...)
		}
	}
	// Elide the leading "." between a field name and a bracket.
	var b strings.Builder
	for i, p := range parts {
		if i > 0 && !strings.HasPrefix(p, "[") {
			b.WriteByte('.')
		}
		b.WriteString(p)
	}
	return b.String()
}

// ErrorType classifies the failure mode; useful for selecting the
// right HTTP status code or response body shape.
type ErrorType string

const (
	// ErrorTypeRequired means the field was missing entirely.
	ErrorTypeRequired ErrorType = "FieldValueRequired"
	// ErrorTypeInvalid means the value was present but malformed.
	ErrorTypeInvalid ErrorType = "FieldValueInvalid"
	// ErrorTypeDuplicate means a uniqueness constraint was violated.
	ErrorTypeDuplicate ErrorType = "FieldValueDuplicate"
	// ErrorTypeNotSupported means the value was not among an enum.
	ErrorTypeNotSupported ErrorType = "FieldValueNotSupported"
	// ErrorTypeForbidden means the field may not be set in this context.
	ErrorTypeForbidden ErrorType = "FieldValueForbidden"
	// ErrorTypeTooLong means a length cap was exceeded.
	ErrorTypeTooLong ErrorType = "FieldValueTooLong"
	// ErrorTypeInternal means the validator itself failed unexpectedly.
	ErrorTypeInternal ErrorType = "InternalError"
)

// Error is a structured validation failure.
type Error struct {
	Type     ErrorType
	Field    string
	BadValue interface{}
	Detail   string
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.BadValue != nil {
		return fmt.Sprintf("%s: %s: %v: %s", e.Field, e.Type, e.BadValue, e.Detail)
	}
	return fmt.Sprintf("%s: %s: %s", e.Field, e.Type, e.Detail)
}

// ErrorList aggregates multiple Errors into one slice.  It implements
// `error` via ToAggregate for callers that want a single value.
type ErrorList []*Error

// Required constructs an ErrorTypeRequired entry.
func Required(p *Path, detail string) *Error {
	return &Error{Type: ErrorTypeRequired, Field: p.String(), Detail: detail}
}

// Invalid constructs an ErrorTypeInvalid entry.
func Invalid(p *Path, badValue interface{}, detail string) *Error {
	return &Error{Type: ErrorTypeInvalid, Field: p.String(), BadValue: badValue, Detail: detail}
}

// Duplicate constructs an ErrorTypeDuplicate entry.
func Duplicate(p *Path, badValue interface{}) *Error {
	return &Error{Type: ErrorTypeDuplicate, Field: p.String(), BadValue: badValue}
}

// NotSupported constructs an ErrorTypeNotSupported entry.  supported
// is the list of acceptable values.
func NotSupported(p *Path, badValue interface{}, supported []string) *Error {
	return &Error{
		Type:     ErrorTypeNotSupported,
		Field:    p.String(),
		BadValue: badValue,
		Detail:   "supported values: " + strings.Join(supported, ", "),
	}
}

// TooLong constructs an ErrorTypeTooLong entry.
func TooLong(p *Path, badValue interface{}, maxLen int) *Error {
	return &Error{
		Type:     ErrorTypeTooLong,
		Field:    p.String(),
		BadValue: badValue,
		Detail:   fmt.Sprintf("must be no more than %d characters", maxLen),
	}
}

// ToAggregate returns a single error whose Error() concatenates every
// entry with a semicolon.  Returns nil for an empty list.
func (list ErrorList) ToAggregate() error {
	if len(list) == 0 {
		return nil
	}
	msgs := make([]string, len(list))
	for i, e := range list {
		msgs[i] = e.Error()
	}
	return fmt.Errorf("%s", strings.Join(msgs, "; "))
}
