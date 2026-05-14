// Package labels implements a label selector parser compatible with
// the Kubernetes label-selection grammar defined at
// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors.
//
// Two dialects are supported:
//
//   - Equality-based:  "env=prod", "tier!=debug"
//   - Set-based:       "env in (prod,staging)", "!beta", "tier"
//
// Selectors may combine any number of requirements with commas; a
// selector matches a label set iff every requirement matches.  Empty
// selectors ("") match all labels, following k8s semantics.
//
// # Design notes
//
// This package is deliberately self-contained: it does not depend on
// k8s.io/apimachinery, so AxiomNizam does not pull in the full k8s
// import graph.  The grammar implemented is a subset of the upstream
// grammar — the comparatively rare "value <n" / ">n" operators are
// omitted.  Upstream's operator names (In, NotIn, Exists, DoesNotExist,
// Equals, NotEquals) are preserved so operators familiar with k8s can
// read the code without translation.
package labels

import (
	"fmt"
	"sort"
	"strings"
)

// Set is an alias for the common map representation of labels.  It is
// defined as a named type so that Set-typed receivers can provide
// convenience methods without polluting the map type.
type Set map[string]string

// Get returns the value of key or the empty string when absent.
func (s Set) Get(key string) string { return s[key] }

// Has reports whether key exists, regardless of value.
func (s Set) Has(key string) bool { _, ok := s[key]; return ok }

// String renders the set in the canonical "k=v,k=v" form with keys in
// sorted order — useful for debug logging and for generating stable
// fingerprints of a label set.
func (s Set) String() string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, s[k]))
	}
	return strings.Join(parts, ",")
}

// Operator enumerates the comparison kinds supported by a Requirement.
type Operator string

const (
	// OpEquals matches when labels[key] == values[0].
	OpEquals Operator = "="
	// OpNotEquals matches when labels[key] != values[0].
	OpNotEquals Operator = "!="
	// OpIn matches when labels[key] is in values.
	OpIn Operator = "in"
	// OpNotIn matches when labels[key] is not in values.
	OpNotIn Operator = "notin"
	// OpExists matches when labels[key] is present (any value).
	OpExists Operator = "exists"
	// OpDoesNotExist matches when labels[key] is absent.
	OpDoesNotExist Operator = "!"
)

// Requirement is a single term in a selector: (key, operator, values).
type Requirement struct {
	Key      string
	Operator Operator
	Values   []string
}

// Matches reports whether the requirement holds for the given label
// set.  Operator semantics follow the k8s definition.
func (r Requirement) Matches(labels Set) bool {
	switch r.Operator {
	case OpEquals:
		if len(r.Values) != 1 {
			return false
		}
		v, ok := labels[r.Key]
		return ok && v == r.Values[0]
	case OpNotEquals:
		if len(r.Values) != 1 {
			return false
		}
		v, ok := labels[r.Key]
		return !ok || v != r.Values[0]
	case OpIn:
		v, ok := labels[r.Key]
		if !ok {
			return false
		}
		for _, want := range r.Values {
			if v == want {
				return true
			}
		}
		return false
	case OpNotIn:
		v, ok := labels[r.Key]
		if !ok {
			return true
		}
		for _, want := range r.Values {
			if v == want {
				return false
			}
		}
		return true
	case OpExists:
		_, ok := labels[r.Key]
		return ok
	case OpDoesNotExist:
		_, ok := labels[r.Key]
		return !ok
	}
	return false
}

// Selector is an ordered list of Requirements joined by logical AND.
// The zero value matches any label set — Parse("") returns this.
type Selector struct {
	requirements []Requirement
}

// Everything returns a selector that matches every label set.
func Everything() Selector { return Selector{} }

// Nothing returns a selector that matches no label set.  Useful as a
// sentinel when constructing filters from user input that may be empty.
func Nothing() Selector {
	return Selector{requirements: []Requirement{{Key: "__nothing__", Operator: OpExists}}}
}

// Add appends a requirement and returns the updated selector.
func (s Selector) Add(r Requirement) Selector {
	out := Selector{requirements: make([]Requirement, len(s.requirements)+1)}
	copy(out.requirements, s.requirements)
	out.requirements[len(s.requirements)] = r
	return out
}

// Requirements returns the underlying requirement list.  The returned
// slice aliases the selector's internal storage; callers that mutate
// it risk corrupting subsequent Matches calls.
func (s Selector) Requirements() []Requirement { return s.requirements }

// Matches reports whether every requirement in s holds for labels.  An
// empty selector matches any label set — including nil.
func (s Selector) Matches(labels Set) bool {
	for _, r := range s.requirements {
		if !r.Matches(labels) {
			return false
		}
	}
	return true
}

// String re-emits the selector in a form that Parse can round-trip.
// Values appearing in set-based requirements are rendered inside
// parentheses with comma separation; no whitespace is emitted.
func (s Selector) String() string {
	parts := make([]string, 0, len(s.requirements))
	for _, r := range s.requirements {
		switch r.Operator {
		case OpEquals:
			parts = append(parts, fmt.Sprintf("%s=%s", r.Key, r.Values[0]))
		case OpNotEquals:
			parts = append(parts, fmt.Sprintf("%s!=%s", r.Key, r.Values[0]))
		case OpIn:
			parts = append(parts, fmt.Sprintf("%s in (%s)", r.Key, strings.Join(r.Values, ",")))
		case OpNotIn:
			parts = append(parts, fmt.Sprintf("%s notin (%s)", r.Key, strings.Join(r.Values, ",")))
		case OpExists:
			parts = append(parts, r.Key)
		case OpDoesNotExist:
			parts = append(parts, "!"+r.Key)
		}
	}
	return strings.Join(parts, ",")
}

// Parse converts an expression to a Selector.  The empty string yields
// Everything().  Syntax errors are reported as a non-nil error whose
// message identifies the offending fragment.
//
// The parser is hand-rolled rather than regex-driven so that error
// messages can include the exact failing term.  It treats a comma as
// the top-level separator, then classifies each term by the first
// operator-ish substring it finds.
func Parse(expr string) (Selector, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return Everything(), nil
	}

	terms := splitTopLevel(expr, ',')
	sel := Selector{requirements: make([]Requirement, 0, len(terms))}
	for _, raw := range terms {
		term := strings.TrimSpace(raw)
		if term == "" {
			return Selector{}, fmt.Errorf("empty term in selector %q", expr)
		}
		req, err := parseTerm(term)
		if err != nil {
			return Selector{}, err
		}
		sel.requirements = append(sel.requirements, req)
	}
	return sel, nil
}

// parseTerm classifies a single requirement expression.  The order of
// checks matters — "!=" must be matched before "=" and "notin" before
// "in", to avoid misclassification.
func parseTerm(term string) (Requirement, error) {
	// Set-based "DoesNotExist": leading '!' followed by a key.
	if strings.HasPrefix(term, "!") {
		key := strings.TrimSpace(term[1:])
		if key == "" {
			return Requirement{}, fmt.Errorf("selector term %q: '!' must be followed by a key", term)
		}
		return Requirement{Key: key, Operator: OpDoesNotExist}, nil
	}

	// Set-based "in" / "notin".  The regex for these operators is
	// " notin " and " in " — note the leading and trailing spaces.
	if idx := indexOfWord(term, " notin "); idx >= 0 {
		return parseSetTerm(term, idx, len(" notin "), OpNotIn)
	}
	if idx := indexOfWord(term, " in "); idx >= 0 {
		return parseSetTerm(term, idx, len(" in "), OpIn)
	}

	// Equality operators.  "!=" first to avoid matching "=" inside it.
	if idx := strings.Index(term, "!="); idx >= 0 {
		return Requirement{
			Key:      strings.TrimSpace(term[:idx]),
			Operator: OpNotEquals,
			Values:   []string{strings.TrimSpace(term[idx+2:])},
		}, nil
	}
	if idx := strings.Index(term, "="); idx >= 0 {
		return Requirement{
			Key:      strings.TrimSpace(term[:idx]),
			Operator: OpEquals,
			Values:   []string{strings.TrimSpace(term[idx+1:])},
		}, nil
	}

	// Bare key ⇒ Exists.
	return Requirement{Key: term, Operator: OpExists}, nil
}

// parseSetTerm extracts key and parenthesised value list for "in" /
// "notin" requirements.
func parseSetTerm(term string, opStart, opLen int, op Operator) (Requirement, error) {
	key := strings.TrimSpace(term[:opStart])
	rest := strings.TrimSpace(term[opStart+opLen:])
	if !strings.HasPrefix(rest, "(") || !strings.HasSuffix(rest, ")") {
		return Requirement{}, fmt.Errorf("selector term %q: %s value list must be parenthesised", term, op)
	}
	inner := strings.TrimSpace(rest[1 : len(rest)-1])
	if inner == "" {
		return Requirement{}, fmt.Errorf("selector term %q: %s value list is empty", term, op)
	}
	rawValues := strings.Split(inner, ",")
	values := make([]string, 0, len(rawValues))
	for _, v := range rawValues {
		v = strings.TrimSpace(v)
		if v == "" {
			return Requirement{}, fmt.Errorf("selector term %q: empty value in list", term)
		}
		values = append(values, v)
	}
	return Requirement{Key: key, Operator: op, Values: values}, nil
}

// splitTopLevel splits s on sep while honouring parenthesis nesting so
// that "k in (a,b),other=c" splits into two terms, not three.
func splitTopLevel(s string, sep rune) []string {
	var out []string
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case sep:
			if depth == 0 {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
	}
	out = append(out, s[start:])
	return out
}

// indexOfWord returns the index of word in s, or -1.  Wraps
// strings.Index but centralises the name so the parser reads more
// clearly.
func indexOfWord(s, word string) int { return strings.Index(s, word) }
