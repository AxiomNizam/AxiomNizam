// Package resource — arithmetic helpers for Quantity maps.
//
// k8s controllers repeatedly sum and compare ResourceLists (a
// map[ResourceName]Quantity).  This file collects those operations
// so call sites stay concise.
package resource

// Name is the typed key in a ResourceList.  Standard values include
// "cpu", "memory", "storage", "ephemeral-storage", or any vendor-
// defined extended resource (e.g. "nvidia.com/gpu").
type Name string

const (
	// NameCPU represents CPU, in cores.
	NameCPU Name = "cpu"
	// NameMemory represents memory, in bytes.
	NameMemory Name = "memory"
	// NameStorage represents volume-backed storage, in bytes.
	NameStorage Name = "storage"
	// NameEphemeralStorage represents node-local ephemeral storage.
	NameEphemeralStorage Name = "ephemeral-storage"
)

// List is the typed map used for requests, limits, and capacity.
type List map[Name]Quantity

// Add produces a new List equal to a+b for every key present in either.
func Add(a, b List) List {
	out := make(List, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if existing, ok := out[k]; ok {
			sum := existing
			sum.Add(v)
			out[k] = sum
		} else {
			out[k] = v
		}
	}
	return out
}

// Sub produces a-b.  Missing keys in b are treated as zero.
func Sub(a, b List) List {
	out := make(List, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if existing, ok := out[k]; ok {
			diff := existing
			diff.Sub(v)
			out[k] = diff
		} else {
			z := NewQuantity(0, v.Format)
			z.Sub(v)
			out[k] = z
		}
	}
	return out
}

// LessThanOrEqual reports whether every entry of a is ≤ the matching
// entry of b.  Keys in a but not in b are considered exceeded; keys
// in b but not in a do not fail the check.  This matches the k8s
// quota-admission semantics.
func LessThanOrEqual(a, b List) bool {
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if av.Cmp(bv) > 0 {
			return false
		}
	}
	return true
}

// Equal reports whether a and b have the same key set and every value
// compares equal.
func Equal(a, b List) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok || av.Cmp(bv) != 0 {
			return false
		}
	}
	return true
}
