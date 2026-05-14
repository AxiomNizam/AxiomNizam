// Package sets — typed sets over integers and bytes.
//
// These mirror String one-to-one.  We deliberately copy the code
// rather than inheriting through an interface: a generic wrapper
// would slow the hot path (set operations are often in tight inner
// loops), and the amount of code is small.
package sets

import "sort"

// Int is a set of ints.
type Int map[int]struct{}

// NewInt constructs an Int populated with items.
func NewInt(items ...int) Int {
	s := make(Int, len(items))
	s.Insert(items...)
	return s
}

// Insert adds items.
func (s Int) Insert(items ...int) Int {
	for _, it := range items {
		s[it] = struct{}{}
	}
	return s
}

// Delete removes items.
func (s Int) Delete(items ...int) Int {
	for _, it := range items {
		delete(s, it)
	}
	return s
}

// Has reports membership.
func (s Int) Has(item int) bool { _, ok := s[item]; return ok }

// Len returns cardinality.
func (s Int) Len() int { return len(s) }

// Union returns a new set.
func (s Int) Union(other Int) Int {
	out := make(Int, len(s)+len(other))
	for k := range s {
		out[k] = struct{}{}
	}
	for k := range other {
		out[k] = struct{}{}
	}
	return out
}

// Intersection returns a new set.
func (s Int) Intersection(other Int) Int {
	small, big := s, other
	if len(other) < len(s) {
		small, big = other, s
	}
	out := make(Int, len(small))
	for k := range small {
		if _, ok := big[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}

// Difference returns s \ other.
func (s Int) Difference(other Int) Int {
	out := make(Int, len(s))
	for k := range s {
		if _, ok := other[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return out
}

// List returns sorted members.
func (s Int) List() []int {
	out := make([]int, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// Byte is a set of byte values — useful when iterating over raw IDs.
type Byte map[byte]struct{}

// NewByte constructs a Byte set.
func NewByte(items ...byte) Byte {
	s := make(Byte, len(items))
	for _, it := range items {
		s[it] = struct{}{}
	}
	return s
}

// Has reports membership.
func (s Byte) Has(item byte) bool { _, ok := s[item]; return ok }

// Insert adds items.
func (s Byte) Insert(items ...byte) Byte {
	for _, it := range items {
		s[it] = struct{}{}
	}
	return s
}

// Len returns cardinality.
func (s Byte) Len() int { return len(s) }
