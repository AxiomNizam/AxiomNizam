// Package fields mirrors k8s.io/apimachinery/pkg/fields — the
// field-selector grammar is intentionally a strict subset of the
// label-selector grammar (only `=`, `==`, `!=`, and comma-joined
// terms; no set operators).  Field selectors address the immutable
// *fields* of an object (spec.nodeName, status.phase) rather than
// user-assigned labels.
//
// We keep this package separate from labels/ because the two
// grammars differ subtly — using the labels parser here would accept
// `status.phase in (Running, Pending)` which the k8s API server
// rejects.  Coherence with upstream matters for kubectl compatibility.
package fields

import (
	"fmt"
	"sort"
	"strings"
)

// Fields is the "key → value" projection of an object's scalar fields.
// Callers build one per object via a FieldsFunc registered with the
// indexer, then feed it to Selector.Matches.
type Fields map[string]string

// Has returns true when key is present (any value).
func (f Fields) Has(key string) bool { _, ok := f[key]; return ok }

// Get returns the value for key, or "".
func (f Fields) Get(key string) string { return f[key] }

// Selector is the compiled form of a field-selector string.
type Selector struct {
	terms []term
}

// term is one equality/inequality test.
type term struct {
	field string
	equal bool
	value string
}

// Everything returns a selector that matches any object.
func Everything() *Selector { return &Selector{} }

// Nothing returns a selector that matches no object.
func Nothing() *Selector {
	return &Selector{terms: []term{{field: "__never__", equal: true, value: "__never__"}}}
}

// Empty reports whether the selector has no terms (= Everything).
func (s *Selector) Empty() bool { return s == nil || len(s.terms) == 0 }

// Matches reports whether fields satisfies every term in s.
func (s *Selector) Matches(fields Fields) bool {
	if s == nil {
		return true
	}
	for _, t := range s.terms {
		v := fields.Get(t.field)
		if t.equal && v != t.value {
			return false
		}
		if !t.equal && v == t.value {
			return false
		}
	}
	return true
}

// Requirements exposes the compiled terms for inspection (e.g. for
// pushing selected predicates down into the storage layer).
type Requirement struct {
	Field    string
	Operator string // "=" or "!="
	Value    string
}

// Requirements returns the selector's terms.
func (s *Selector) Requirements() []Requirement {
	out := make([]Requirement, len(s.terms))
	for i, t := range s.terms {
		op := "="
		if !t.equal {
			op = "!="
		}
		out[i] = Requirement{Field: t.field, Operator: op, Value: t.value}
	}
	return out
}

// String renders the selector back to its canonical form, with terms
// in field-name order for stable comparison.
func (s *Selector) String() string {
	if s.Empty() {
		return ""
	}
	sorted := append([]term(nil), s.terms...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].field < sorted[j].field })
	parts := make([]string, len(sorted))
	for i, t := range sorted {
		op := "="
		if !t.equal {
			op = "!="
		}
		parts[i] = t.field + op + t.value
	}
	return strings.Join(parts, ",")
}

// ParseSelector parses the selector grammar.  Comma separates terms;
// each term is "<field><op><value>" where op is one of "=", "==", "!=".
// Whitespace around separators is tolerated.
func ParseSelector(s string) (*Selector, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Everything(), nil
	}
	sel := &Selector{}
	for _, raw := range strings.Split(s, ",") {
		t, err := parseTerm(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		sel.terms = append(sel.terms, t)
	}
	return sel, nil
}

// parseTerm splits a single term, preferring the longer operator
// "!=" over "=" when both would match.
func parseTerm(s string) (term, error) {
	if i := strings.Index(s, "!="); i >= 0 {
		return term{field: strings.TrimSpace(s[:i]), equal: false, value: strings.TrimSpace(s[i+2:])}, nil
	}
	if i := strings.Index(s, "=="); i >= 0 {
		return term{field: strings.TrimSpace(s[:i]), equal: true, value: strings.TrimSpace(s[i+2:])}, nil
	}
	if i := strings.Index(s, "="); i >= 0 {
		return term{field: strings.TrimSpace(s[:i]), equal: true, value: strings.TrimSpace(s[i+1:])}, nil
	}
	return term{}, fmt.Errorf("field selector term %q has no '=' or '!='", s)
}

// OneTermEqualSelector constructs a Selector matching a single "k=v".
// Convenience for programmatic use.
func OneTermEqualSelector(field, value string) *Selector {
	return &Selector{terms: []term{{field: field, equal: true, value: value}}}
}

// AndSelectors combines several selectors into a single conjunction.
func AndSelectors(selectors ...*Selector) *Selector {
	out := &Selector{}
	for _, sel := range selectors {
		if sel == nil {
			continue
		}
		out.terms = append(out.terms, sel.terms...)
	}
	return out
}
