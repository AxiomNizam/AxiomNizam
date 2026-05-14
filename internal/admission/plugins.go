// Package admission — bundled plugins that are useful across most
// resource types.  They are optional: applications may register them on
// their chain or substitute their own.
package admission

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

// -----------------------------------------------------------------------------
// Mutators
// -----------------------------------------------------------------------------

// DefaultValueMutator sets a default for a single field when the field
// is absent or explicitly nil.  The field path is dot-delimited ("spec.
// replicas") and traverses nested maps — lists are not supported.
type DefaultValueMutator struct {
	NameStr string
	Path    string
	Value   interface{}
}

// Name implements Mutator.
func (m *DefaultValueMutator) Name() string { return m.NameStr }

// Mutate implements Mutator.
func (m *DefaultValueMutator) Mutate(_ context.Context, req *Request) error {
	setIfAbsent(req.Object, splitPath(m.Path), m.Value)
	return nil
}

// TimestampMutator stamps CreationTimestamp on CREATE and a
// LastModified timestamp on UPDATE.  Fields are placed under
// metadata.creationTimestamp / metadata.lastModified as RFC3339 strings
// so that JSON round-trips preserve them exactly.
type TimestampMutator struct {
	NameStr string
	Now     func() time.Time // overridable for tests
}

// Name implements Mutator.
func (m *TimestampMutator) Name() string { return m.NameStr }

// Mutate implements Mutator.
func (m *TimestampMutator) Mutate(_ context.Context, req *Request) error {
	now := time.Now
	if m.Now != nil {
		now = m.Now
	}
	ts := now().UTC().Format(time.RFC3339Nano)
	meta, ok := req.Object["metadata"].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
		req.Object["metadata"] = meta
	}
	switch req.Operation {
	case OperationCreate:
		if _, exists := meta["creationTimestamp"]; !exists {
			meta["creationTimestamp"] = ts
		}
		meta["lastModified"] = ts
	case OperationUpdate:
		meta["lastModified"] = ts
	}
	return nil
}

// -----------------------------------------------------------------------------
// Validators
// -----------------------------------------------------------------------------

// RequiredFieldsValidator rejects a request whose object lacks any of
// the listed dot-delimited paths.  Empty string / nil / empty map /
// empty slice all count as "missing".
type RequiredFieldsValidator struct {
	NameStr string
	Fields  []string
}

// Name implements Validator.
func (v *RequiredFieldsValidator) Name() string { return v.NameStr }

// Validate implements Validator.
func (v *RequiredFieldsValidator) Validate(_ context.Context, req *Request) error {
	var missing []string
	for _, f := range v.Fields {
		if !isPresent(req.Object, splitPath(f)) {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		return &RejectionError{
			PluginName: v.NameStr,
			Reason:     fmt.Sprintf("required field(s) missing: %v", missing),
		}
	}
	return nil
}

// NameFormatValidator enforces a regex on metadata.name.  This is a
// common requirement for kubectl-style kinds where names must be DNS
// label compatible.
type NameFormatValidator struct {
	NameStr string
	Pattern *regexp.Regexp
}

// Name implements Validator.
func (v *NameFormatValidator) Name() string { return v.NameStr }

// Validate implements Validator.
func (v *NameFormatValidator) Validate(_ context.Context, req *Request) error {
	if req.Name == "" {
		return &RejectionError{PluginName: v.NameStr, Reason: "metadata.name is empty"}
	}
	if v.Pattern != nil && !v.Pattern.MatchString(req.Name) {
		return &RejectionError{
			PluginName: v.NameStr,
			Reason:     fmt.Sprintf("name %q does not match %s", req.Name, v.Pattern.String()),
		}
	}
	return nil
}

// ImmutableFieldsValidator rejects updates that attempt to change any of
// the listed paths after creation.  Delete and create operations are
// not affected.
type ImmutableFieldsValidator struct {
	NameStr string
	Fields  []string
}

// Name implements Validator.
func (v *ImmutableFieldsValidator) Name() string { return v.NameStr }

// Validate implements Validator.
func (v *ImmutableFieldsValidator) Validate(_ context.Context, req *Request) error {
	if req.Operation != OperationUpdate || req.OldObject == nil {
		return nil
	}
	for _, f := range v.Fields {
		path := splitPath(f)
		before, _ := lookup(req.OldObject, path)
		after, _ := lookup(req.Object, path)
		if !deepEqual(before, after) {
			return &RejectionError{
				PluginName: v.NameStr,
				Reason:     fmt.Sprintf("field %q is immutable", f),
			}
		}
	}
	return nil
}
