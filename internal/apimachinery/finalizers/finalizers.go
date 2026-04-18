// Package finalizers implements the Kubernetes finalizer pattern:
// controllers attach a finalizer token to metadata.finalizers before
// allocating out-of-band resources (external records, cloud assets),
// then delete that token once cleanup completes.  Deletion does not
// physically remove the object until the finalizer set is empty — so
// a crashed controller can resume cleanup on restart without losing
// track of what was pending.
//
// The canonical reference is
// https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/ .
//
// This package operates on the opaque
// map[string]interface{} object model used by AxiomNizam's resource
// store.  Typed resource implementations can build on top of these
// helpers by exposing methods that forward to them.
package finalizers

import (
	"time"
)

// FinalizersPath is the JSON path of the slice this package manages.
// Exposed as a constant because multiple subsystems (audit, admission,
// patch) need to refer to it.
const FinalizersPath = "metadata.finalizers"

// DeletionTimestampPath is the JSON path of the RFC3339 timestamp set
// by the API server when a Delete request is received.  Controllers
// branch on its presence to enter the cleanup path.
const DeletionTimestampPath = "metadata.deletionTimestamp"

// Contains reports whether name is already present in the finalizer
// slice of obj.  A nil / missing metadata or finalizer list is treated
// as an empty set.
func Contains(obj map[string]interface{}, name string) bool {
	for _, existing := range current(obj) {
		if existing == name {
			return true
		}
	}
	return false
}

// Add inserts name into the finalizer slice when not already present.
// Returns true when the object was modified.  Intermediate metadata
// maps are created as needed so first-time callers don't have to pre-
// populate the tree.
func Add(obj map[string]interface{}, name string) bool {
	if Contains(obj, name) {
		return false
	}
	list := current(obj)
	list = append(list, name)
	setList(obj, list)
	return true
}

// Remove deletes name from the finalizer slice.  Returns true when the
// object was modified.  A missing finalizer is a silent no-op — this
// mirrors k8s semantics where Remove is idempotent.
func Remove(obj map[string]interface{}, name string) bool {
	list := current(obj)
	out := make([]string, 0, len(list))
	removed := false
	for _, existing := range list {
		if existing == name {
			removed = true
			continue
		}
		out = append(out, existing)
	}
	if !removed {
		return false
	}
	setList(obj, out)
	return true
}

// IsBeingDeleted reports whether metadata.deletionTimestamp is set.
// Controllers MUST test this before every reconcile: a resource under
// deletion must only have its finalizers executed, never its normal
// reconcile logic, or it will be re-created out from under the user.
func IsBeingDeleted(obj map[string]interface{}) bool {
	_, ok := lookupString(obj, "metadata", "deletionTimestamp")
	return ok
}

// MarkForDeletion stamps metadata.deletionTimestamp with now().  The
// returned bool is false when the stamp was already present (callers
// generally emit a "delete already in progress" event in that case).
func MarkForDeletion(obj map[string]interface{}, now func() time.Time) bool {
	if IsBeingDeleted(obj) {
		return false
	}
	if now == nil {
		now = time.Now
	}
	meta := metadataOrCreate(obj)
	meta["deletionTimestamp"] = now().UTC().Format(time.RFC3339Nano)
	return true
}

// IsReadyForPhysicalDelete reports whether the object can safely be
// removed from the underlying store — that is, it is marked for
// deletion AND its finalizer set is empty.
func IsReadyForPhysicalDelete(obj map[string]interface{}) bool {
	return IsBeingDeleted(obj) && len(current(obj)) == 0
}

// List returns a copy of the current finalizer slice.  Returned slice
// is safe to mutate without affecting obj.
func List(obj map[string]interface{}) []string {
	src := current(obj)
	out := make([]string, len(src))
	copy(out, src)
	return out
}

// --- internal helpers ---------------------------------------------------

// current returns the live finalizer slice as []string.  An empty
// slice is returned for nil / missing data or when the underlying
// value is not a list of strings.  The returned slice aliases the
// metadata map — callers that need ownership should copy it.
func current(obj map[string]interface{}) []string {
	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}
	raw, ok := meta["finalizers"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// setList writes list back to obj, round-tripping through
// []interface{} so that subsequent JSON marshalling produces the
// expected shape.
func setList(obj map[string]interface{}, list []string) {
	meta := metadataOrCreate(obj)
	iface := make([]interface{}, len(list))
	for i, s := range list {
		iface[i] = s
	}
	meta["finalizers"] = iface
}

// metadataOrCreate returns the metadata map for obj, creating the slot
// if necessary.  A non-map metadata value is overwritten: the API
// contract says metadata must be a map, and a corrupt value cannot be
// preserved without violating later invariants.
func metadataOrCreate(obj map[string]interface{}) map[string]interface{} {
	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
		obj["metadata"] = meta
	}
	return meta
}

// lookupString returns the nested string at path or ("", false).
func lookupString(obj map[string]interface{}, path ...string) (string, bool) {
	var cur interface{} = obj
	for _, seg := range path {
		asMap, ok := cur.(map[string]interface{})
		if !ok {
			return "", false
		}
		cur, ok = asMap[seg]
		if !ok {
			return "", false
		}
	}
	s, ok := cur.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}
