// Package diff produces human-readable diffs of arbitrary Go values.
// It is primarily used by admission webhooks and audit logs to show
// what a mutating plugin changed or what a user's PATCH altered.
//
// The output format is a minimal unified-ish diff: one line per key
// that differs, prefixed with "-" for the old side, "+" for the new.
// Deeply-nested changes are rendered under their dotted path so the
// diff stays readable even when the input is a large nested object.
//
// This is not a drop-in for `google/go-cmp` — it's an order of
// magnitude smaller and only tries to be useful, not minimal.
package diff

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ObjectDiff returns a string describing how a differs from b.  An
// empty return value means they are equal.
func ObjectDiff(a, b interface{}) string {
	var lines []string
	walk("", a, b, &lines)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// walk recursively compares a and b, emitting diff lines for each
// leaf difference.  path is the dotted/bracketed accessor so callers
// can locate the change in the original document.
func walk(path string, a, b interface{}, lines *[]string) {
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			emit(path, a, b, lines)
			return
		}
		keys := mergedKeys(av, bv)
		for _, k := range keys {
			sub := path
			if sub == "" {
				sub = k
			} else {
				sub = path + "." + k
			}
			walk(sub, av[k], bv[k], lines)
		}
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok {
			emit(path, a, b, lines)
			return
		}
		// Lists are compared positionally — this does not attempt to
		// detect inserts / reorders.  For k8s-style list-with-key
		// semantics, callers should normalise lists before diffing.
		max := len(av)
		if len(bv) > max {
			max = len(bv)
		}
		for i := 0; i < max; i++ {
			sub := path + "[" + strconv.Itoa(i) + "]"
			var left, right interface{}
			if i < len(av) {
				left = av[i]
			}
			if i < len(bv) {
				right = bv[i]
			}
			walk(sub, left, right, lines)
		}
	default:
		if !reflect.DeepEqual(a, b) {
			emit(path, a, b, lines)
		}
	}
}

// emit formats a single difference.  Missing values (when one side
// was absent in a parent map) render as "<absent>" rather than "nil"
// to distinguish them from an explicit JSON null.
func emit(path string, a, b interface{}, lines *[]string) {
	render := func(v interface{}) string {
		if v == nil {
			return "<absent>"
		}
		return fmt.Sprintf("%v", v)
	}
	*lines = append(*lines, fmt.Sprintf("  %s: %s -> %s", path, render(a), render(b)))
}

// mergedKeys returns the union of both maps' keys in sorted order.
func mergedKeys(a, b map[string]interface{}) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
