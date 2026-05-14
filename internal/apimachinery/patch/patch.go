// Package patch implements the three patch content-types accepted by
// a Kubernetes-style API server:
//
//   - application/merge-patch+json          (RFC 7396)
//   - application/json-patch+json           (RFC 6902)
//   - application/strategic-merge-patch+json  (k8s-specific)
//
// A patch request is a partial document describing modifications to
// apply to an existing resource.  Handlers select the content-type via
// the HTTP Content-Type header; this package provides a pure-Go Apply
// entry point that accepts a Type and returns the resulting document.
//
// # Implementation notes
//
// The strategic merge variant in upstream k8s is driven by struct tags
// (`patchStrategy:"merge" patchMergeKey:"name"`) to identify list
// identity keys.  Because AxiomNizam's objects are schemaless
// map[string]interface{} values, we accept those hints at call time
// through a StrategicKeys map: path → mergeKey.  Absent a hint, list
// fields are replaced wholesale — matching the "replace" default of
// strategic merge.
package patch

import (
	"encoding/json"
	"fmt"
)

// Type enumerates the accepted patch content-types.
type Type string

const (
	// TypeMerge is RFC 7396 JSON Merge Patch.  Nil values delete keys;
	// all other values replace them.  Lists are always replaced.
	TypeMerge Type = "application/merge-patch+json"

	// TypeJSON is RFC 6902 JSON Patch — an explicit op list.  More
	// expressive than merge but more verbose to author.
	TypeJSON Type = "application/json-patch+json"

	// TypeStrategicMerge is the k8s-specific dialect that treats lists
	// of objects as merge-indexed on a named key.
	TypeStrategicMerge Type = "application/strategic-merge-patch+json"
)

// Apply produces a new document that is the result of applying patch
// to original using the semantics associated with t.  Neither input is
// mutated.  keys is consulted only for TypeStrategicMerge.
func Apply(t Type, original, patch []byte, keys StrategicKeys) ([]byte, error) {
	var doc interface{}
	if err := json.Unmarshal(original, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal original: %w", err)
	}
	var patchDoc interface{}
	if err := json.Unmarshal(patch, &patchDoc); err != nil {
		return nil, fmt.Errorf("unmarshal patch: %w", err)
	}

	var result interface{}
	switch t {
	case TypeMerge:
		result = mergePatch(doc, patchDoc)
	case TypeStrategicMerge:
		result = strategicMerge(doc, patchDoc, keys, "")
	case TypeJSON:
		ops, ok := patchDoc.([]interface{})
		if !ok {
			return nil, fmt.Errorf("json-patch requires an array body")
		}
		var err error
		result, err = jsonPatch(doc, ops)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported patch type %q", t)
	}

	return json.Marshal(result)
}

// StrategicKeys identifies which field of an object list functions as
// the merge key.  Keys are dotted paths ("spec.containers"); values
// are the field name within each list element ("name").
type StrategicKeys map[string]string

// -----------------------------------------------------------------------------
// RFC 7396 merge patch
// -----------------------------------------------------------------------------

// mergePatch implements the recursive algorithm from RFC 7396.
func mergePatch(target, patch interface{}) interface{} {
	pm, ok := patch.(map[string]interface{})
	if !ok {
		// Non-object patches replace target wholesale (RFC 7396 §2).
		return patch
	}
	tm, ok := target.(map[string]interface{})
	if !ok {
		tm = map[string]interface{}{}
	}
	for k, v := range pm {
		if v == nil {
			delete(tm, k)
			continue
		}
		tm[k] = mergePatch(tm[k], v)
	}
	return tm
}

// -----------------------------------------------------------------------------
// RFC 6902 JSON patch (subset: add / remove / replace / copy / move / test)
// -----------------------------------------------------------------------------

// jsonPatch applies a sequence of ops.  Ops are atomic: the first
// failing op aborts and returns the original document unchanged —
// callers that need best-effort semantics should split the op list
// themselves.
func jsonPatch(target interface{}, ops []interface{}) (interface{}, error) {
	for i, rawOp := range ops {
		op, ok := rawOp.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("op %d: not an object", i)
		}
		name, _ := op["op"].(string)
		path, _ := op["path"].(string)
		value := op["value"]
		from, _ := op["from"].(string)

		var err error
		switch name {
		case "add":
			target, err = jpAdd(target, path, value)
		case "remove":
			target, err = jpRemove(target, path)
		case "replace":
			if _, err = jpGet(target, path); err != nil {
				return nil, fmt.Errorf("op %d replace: %w", i, err)
			}
			target, err = jpAdd(target, path, value)
		case "copy":
			v, gerr := jpGet(target, from)
			if gerr != nil {
				return nil, fmt.Errorf("op %d copy: %w", i, gerr)
			}
			target, err = jpAdd(target, path, v)
		case "move":
			v, gerr := jpGet(target, from)
			if gerr != nil {
				return nil, fmt.Errorf("op %d move: %w", i, gerr)
			}
			target, err = jpRemove(target, from)
			if err == nil {
				target, err = jpAdd(target, path, v)
			}
		case "test":
			v, gerr := jpGet(target, path)
			if gerr != nil {
				return nil, fmt.Errorf("op %d test: %w", i, gerr)
			}
			if !deepEqual(v, value) {
				return nil, fmt.Errorf("op %d test: value mismatch at %s", i, path)
			}
		default:
			return nil, fmt.Errorf("op %d: unsupported op %q", i, name)
		}
		if err != nil {
			return nil, fmt.Errorf("op %d %s: %w", i, name, err)
		}
	}
	return target, nil
}

// -----------------------------------------------------------------------------
// Strategic merge patch
// -----------------------------------------------------------------------------

// strategicMerge is mergePatch plus list-keyed merging.  When the
// currently-traversed path is registered in keys, list elements are
// merged by identity rather than replaced.
func strategicMerge(target, patch interface{}, keys StrategicKeys, path string) interface{} {
	pm, ok := patch.(map[string]interface{})
	if !ok {
		return patch
	}
	tm, ok := target.(map[string]interface{})
	if !ok {
		tm = map[string]interface{}{}
	}
	for k, v := range pm {
		childPath := k
		if path != "" {
			childPath = path + "." + k
		}
		if v == nil {
			delete(tm, k)
			continue
		}
		if patchList, isList := v.([]interface{}); isList {
			if mergeKey, hasKey := keys[childPath]; hasKey {
				tm[k] = mergeListsByKey(tm[k], patchList, mergeKey, keys, childPath)
				continue
			}
			// No hint: replace the list wholesale.
			tm[k] = patchList
			continue
		}
		tm[k] = strategicMerge(tm[k], v, keys, childPath)
	}
	return tm
}

// mergeListsByKey merges patchList into targetList using mergeKey to
// identify elements.  Elements present only in patchList are appended;
// elements present only in targetList are preserved; elements present
// in both are recursively strategic-merged.
func mergeListsByKey(target interface{}, patchList []interface{}, mergeKey string, keys StrategicKeys, path string) interface{} {
	var targetList []interface{}
	if t, ok := target.([]interface{}); ok {
		targetList = t
	}
	byKey := make(map[string]int, len(targetList))
	for i, entry := range targetList {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		k, _ := m[mergeKey].(string)
		byKey[k] = i
	}
	for _, entry := range patchList {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		k, _ := m[mergeKey].(string)
		if idx, exists := byKey[k]; exists {
			targetList[idx] = strategicMerge(targetList[idx], entry, keys, path)
		} else {
			targetList = append(targetList, entry)
			byKey[k] = len(targetList) - 1
		}
	}
	return targetList
}
