// Package sets provides typed set implementations patterned after
// k8s.io/apimachinery/pkg/util/sets.  Go's built-in map[T]struct{} is
// a fine set, but the domain-specific helpers (Insert, Union, Diff,
// HasAll, List, PopAny) are wordy to rewrite at every call site.
//
// Three concrete types are provided — String, Int, and Byte — because
// they cover the common workload and keep the package free of generics
// for compatibility with older toolchains that may be pinned for
// build reproducibility.  Generic sets can trivially wrap these.
package sets

import "sort"

// String is a set of strings.
type String map[string]struct{}

// NewString constructs a String populated with items.
func NewString(items ...string) String {
	s := make(String, len(items))
	s.Insert(items...)
	return s
}

// Insert adds one or more items.
func (s String) Insert(items ...string) String {
	for _, it := range items {
		s[it] = struct{}{}
	}
	return s
}

// Delete removes one or more items.
func (s String) Delete(items ...string) String {
	for _, it := range items {
		delete(s, it)
	}
	return s
}

// Has reports membership.
func (s String) Has(item string) bool { _, ok := s[item]; return ok }

// HasAll reports whether every item is present.
func (s String) HasAll(items ...string) bool {
	for _, it := range items {
		if !s.Has(it) {
			return false
		}
	}
	return true
}

// HasAny reports whether at least one item is present.
func (s String) HasAny(items ...string) bool {
	for _, it := range items {
		if s.Has(it) {
			return true
		}
	}
	return false
}

// Len returns the cardinality.
func (s String) Len() int { return len(s) }

// Union returns s ∪ other as a new set.
func (s String) Union(other String) String {
	out := make(String, len(s)+len(other))
	for k := range s {
		out[k] = struct{}{}
	}
	for k := range other {
		out[k] = struct{}{}
	}
	return out
}

// Intersection returns s ∩ other as a new set.
func (s String) Intersection(other String) String {
	// Iterate the smaller set for O(min) membership checks.
	small, big := s, other
	if len(other) < len(s) {
		small, big = other, s
	}
	out := make(String, len(small))
	for k := range small {
		if _, ok := big[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}

// Difference returns s \ other.
func (s String) Difference(other String) String {
	out := make(String, len(s))
	for k := range s {
		if _, ok := other[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return out
}

// Equal reports set equality.
func (s String) Equal(other String) bool {
	if len(s) != len(other) {
		return false
	}
	for k := range s {
		if _, ok := other[k]; !ok {
			return false
		}
	}
	return true
}

// IsSuperset reports whether s contains every member of other.
func (s String) IsSuperset(other String) bool {
	for k := range other {
		if _, ok := s[k]; !ok {
			return false
		}
	}
	return true
}

// List returns the members in ascending order — useful for stable
// rendering in error messages and diff output.
func (s String) List() []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// UnsortedList returns the members in arbitrary order.  Faster when
// ordering doesn't matter.
func (s String) UnsortedList() []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

// PopAny removes and returns an arbitrary member, or ("", false) when
// empty.  Useful for draining work queues.
func (s String) PopAny() (string, bool) {
	for k := range s {
		delete(s, k)
		return k, true
	}
	return "", false
}
