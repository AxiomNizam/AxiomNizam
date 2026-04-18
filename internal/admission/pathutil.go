// Package admission — tiny helpers for dot-delimited path lookup and
// equality over map[string]interface{} trees.  Kept separate from the
// public plugin API so callers can substitute their own traversal logic
// if they prefer to work with typed structs.
package admission

import (
	"reflect"
	"strings"
)

// splitPath turns "metadata.labels.app" into ["metadata","labels","app"].
// A nil or empty input yields a nil slice.
func splitPath(p string) []string {
	if p == "" {
		return nil
	}
	return strings.Split(p, ".")
}

// lookup walks m along path and returns the final value plus whether
// the full path existed.  Intermediate non-map values cause ok=false.
func lookup(m map[string]interface{}, path []string) (interface{}, bool) {
	if len(path) == 0 {
		return nil, false
	}
	var cur interface{} = m
	for _, seg := range path {
		asMap, ok := cur.(map[string]interface{})
		if !ok {
			return nil, false
		}
		cur, ok = asMap[seg]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// isPresent reports whether path resolves to a non-empty value.  Empty
// string, nil, zero-length slice, and zero-length map all count as
// "absent" — this mirrors k8s's treatment of required fields.
func isPresent(m map[string]interface{}, path []string) bool {
	v, ok := lookup(m, path)
	if !ok || v == nil {
		return false
	}
	switch x := v.(type) {
	case string:
		return x != ""
	case []interface{}:
		return len(x) > 0
	case map[string]interface{}:
		return len(x) > 0
	default:
		return true
	}
}

// setIfAbsent writes value at path when the path currently resolves to
// "absent" (as per isPresent).  Intermediate maps are created as
// needed; intermediate non-map values cause the call to be a no-op.
func setIfAbsent(m map[string]interface{}, path []string, value interface{}) {
	if len(path) == 0 {
		return
	}
	cur := m
	for i, seg := range path {
		if i == len(path)-1 {
			if !isPresent(cur, []string{seg}) {
				cur[seg] = value
			}
			return
		}
		next, ok := cur[seg].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			cur[seg] = next
		}
		cur = next
	}
}

// deepEqual compares two opaque values using reflect.DeepEqual but with
// nil-map tolerance — reflect treats a nil map and an empty map as
// unequal, which is a footgun when admission requests are unmarshalled
// from JSON.
func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if am, ok := a.(map[string]interface{}); ok && len(am) == 0 {
		if bm, ok := b.(map[string]interface{}); ok && len(bm) == 0 {
			return true
		}
		if b == nil {
			return true
		}
	}
	if bm, ok := b.(map[string]interface{}); ok && len(bm) == 0 {
		if a == nil {
			return true
		}
	}
	return reflect.DeepEqual(a, b)
}
