// Package patch — RFC 6902 pointer traversal helpers.
//
// RFC 6901 defines JSON Pointers as a slash-delimited path with two
// escape sequences: "~1" for '/' and "~0" for '~'.  These helpers
// implement that grammar and walk arbitrary map/slice trees to read,
// insert, or remove values.  They are intentionally permissive in
// error messages to ease debugging of hand-authored patches.
package patch

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// jpUnescape turns a JSON-pointer reference token into its literal
// form per RFC 6901 §4.
func jpUnescape(tok string) string {
	// "~1" decodes before "~0" — the order matters when a pointer
	// contains "~01" (= "~1" literal): decoding "~0" first would turn
	// it into "~1" and then the second pass would wrongly produce "/".
	tok = strings.ReplaceAll(tok, "~1", "/")
	tok = strings.ReplaceAll(tok, "~0", "~")
	return tok
}

// jpTokens splits a "/a/b/c" pointer into ["a","b","c"].  The empty
// pointer "" refers to the document root and yields nil.
func jpTokens(ptr string) ([]string, error) {
	if ptr == "" {
		return nil, nil
	}
	if !strings.HasPrefix(ptr, "/") {
		return nil, fmt.Errorf("json pointer %q must start with '/'", ptr)
	}
	parts := strings.Split(ptr[1:], "/")
	for i, p := range parts {
		parts[i] = jpUnescape(p)
	}
	return parts, nil
}

// jpGet reads the value at ptr.  Returns (nil, error) when the path
// does not resolve.
func jpGet(doc interface{}, ptr string) (interface{}, error) {
	tokens, err := jpTokens(ptr)
	if err != nil {
		return nil, err
	}
	cur := doc
	for _, tok := range tokens {
		switch c := cur.(type) {
		case map[string]interface{}:
			v, ok := c[tok]
			if !ok {
				return nil, fmt.Errorf("path %s: key %q not found", ptr, tok)
			}
			cur = v
		case []interface{}:
			idx, err := strconv.Atoi(tok)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, fmt.Errorf("path %s: bad index %q", ptr, tok)
			}
			cur = c[idx]
		default:
			return nil, fmt.Errorf("path %s: cannot traverse into %T", ptr, cur)
		}
	}
	return cur, nil
}

// jpAdd inserts value at ptr.  When the parent is a slice, "-" means
// "append" (RFC 6902 §4.1).  The returned value is the updated root —
// callers must replace the original reference to observe changes that
// re-assign the root itself (e.g. ptr == "").
func jpAdd(doc interface{}, ptr string, value interface{}) (interface{}, error) {
	tokens, err := jpTokens(ptr)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return value, nil
	}
	parent, err := jpGet(doc, pointerParent(ptr))
	if err != nil {
		return nil, err
	}
	last := tokens[len(tokens)-1]
	switch p := parent.(type) {
	case map[string]interface{}:
		p[last] = value
	case []interface{}:
		if last == "-" {
			// Append: because jpGet returned p by value, we must
			// write back through the document to observe the extended
			// slice.  Use a helper that descends with index awareness.
			return jpAppend(doc, pointerParent(ptr), value)
		}
		idx, err := strconv.Atoi(last)
		if err != nil || idx < 0 || idx > len(p) {
			return nil, fmt.Errorf("path %s: bad insert index %q", ptr, last)
		}
		newSlice := make([]interface{}, 0, len(p)+1)
		newSlice = append(newSlice, p[:idx]...)
		newSlice = append(newSlice, value)
		newSlice = append(newSlice, p[idx:]...)
		return jpReplaceAt(doc, pointerParent(ptr), newSlice)
	default:
		return nil, fmt.Errorf("path %s: parent %T is not addressable", ptr, parent)
	}
	return doc, nil
}

// jpAppend appends value to the slice at ptr and writes the new slice
// back to the document.  Splitting this from jpAdd keeps the latter's
// map-case fast path simple.
func jpAppend(doc interface{}, ptr string, value interface{}) (interface{}, error) {
	cur, err := jpGet(doc, ptr)
	if err != nil {
		return nil, err
	}
	list, ok := cur.([]interface{})
	if !ok {
		return nil, fmt.Errorf("path %s: target is not an array", ptr)
	}
	return jpReplaceAt(doc, ptr, append(list, value))
}

// jpReplaceAt replaces the value at ptr with newVal and returns the
// updated root.  Internally it rewrites the containing map / slice so
// that slice growth is visible to the caller.
func jpReplaceAt(doc interface{}, ptr string, newVal interface{}) (interface{}, error) {
	tokens, err := jpTokens(ptr)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return newVal, nil
	}
	parent, err := jpGet(doc, pointerParent(ptr))
	if err != nil {
		return nil, err
	}
	last := tokens[len(tokens)-1]
	switch p := parent.(type) {
	case map[string]interface{}:
		p[last] = newVal
		return doc, nil
	case []interface{}:
		idx, err := strconv.Atoi(last)
		if err != nil || idx < 0 || idx >= len(p) {
			return nil, fmt.Errorf("path %s: bad index %q", ptr, last)
		}
		p[idx] = newVal
		// Slices are reference types only for the underlying array —
		// length changes require writing the slice header back to the
		// parent.  Recurse once more to put `p` in place of itself,
		// which is a no-op unless we actually grew the slice in a
		// caller above.
		return jpReplaceAt(doc, pointerParent(ptr), p)
	default:
		return nil, fmt.Errorf("path %s: parent %T is not addressable", ptr, parent)
	}
}

// jpRemove deletes the value at ptr.  Removing from a slice shifts
// subsequent elements left; removing a map key deletes it outright.
func jpRemove(doc interface{}, ptr string) (interface{}, error) {
	tokens, err := jpTokens(ptr)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("cannot remove document root")
	}
	parent, err := jpGet(doc, pointerParent(ptr))
	if err != nil {
		return nil, err
	}
	last := tokens[len(tokens)-1]
	switch p := parent.(type) {
	case map[string]interface{}:
		if _, ok := p[last]; !ok {
			return nil, fmt.Errorf("path %s: key %q not found", ptr, last)
		}
		delete(p, last)
		return doc, nil
	case []interface{}:
		idx, err := strconv.Atoi(last)
		if err != nil || idx < 0 || idx >= len(p) {
			return nil, fmt.Errorf("path %s: bad index %q", ptr, last)
		}
		newSlice := make([]interface{}, 0, len(p)-1)
		newSlice = append(newSlice, p[:idx]...)
		newSlice = append(newSlice, p[idx+1:]...)
		return jpReplaceAt(doc, pointerParent(ptr), newSlice)
	default:
		return nil, fmt.Errorf("path %s: parent %T is not addressable", ptr, parent)
	}
}

// pointerParent returns the pointer to ptr's parent.  "/a/b" → "/a",
// "/a" → "", "" → "".
func pointerParent(ptr string) string {
	i := strings.LastIndex(ptr, "/")
	if i <= 0 {
		return ""
	}
	return ptr[:i]
}

// deepEqual defers to reflect for structural equality — used by the
// RFC 6902 "test" op to compare arbitrary JSON values.
func deepEqual(a, b interface{}) bool { return reflect.DeepEqual(a, b) }
