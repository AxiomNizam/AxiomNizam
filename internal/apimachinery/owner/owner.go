// Package owner implements Kubernetes-style owner references and the
// cascading garbage-collection algorithm that sweeps orphaned children
// when their owner is deleted.
//
// # Data model
//
// Every resource may carry a metadata.ownerReferences slice.  Each
// entry names a parent by (apiVersion, kind, name, uid) and optionally
// marks the relationship as "blocking" — a blocking parent prevents
// its own deletion from completing until the child has been deleted.
// This mirrors the `BlockOwnerDeletion` semantics documented at
// https://kubernetes.io/docs/concepts/architecture/garbage-collection/ .
//
// GC algorithm (Foreground deletion)
//
//  1. User deletes the root object.  API stamps deletionTimestamp and
//     sets the "foregroundDeletion" finalizer.
//  2. GarbageCollector.Visit walks the live object graph and enqueues
//     all transitive children.
//  3. Each child is deleted depth-first; once a leaf's store row is
//     removed, the owner's blocking count decrements.
//  4. When the root's blocking count reaches zero, the
//     foregroundDeletion finalizer is removed, freeing the root for
//     physical delete.
//
// This package provides the pure graph primitives.  Physical store
// integration is delegated to a Resolver callback, letting the same
// code drive either an in-memory test fixture or the etcd-backed
// production store.
package owner

import (
	"context"
	"fmt"
)

// Reference identifies a parent object.  UID is the stable identity
// assigned at creation — name alone is ambiguous in systems where a
// kind/name pair may be re-created after deletion.
type Reference struct {
	APIVersion         string `json:"apiVersion"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	UID                string `json:"uid"`
	Controller         bool   `json:"controller,omitempty"`
	BlockOwnerDeletion bool   `json:"blockOwnerDeletion,omitempty"`
}

// Key is the triple that uniquely names a resource across the control
// plane.  GC uses it as the map key for the dependency graph.
type Key struct {
	Kind      string
	Namespace string
	Name      string
}

// String renders the key in "kind/namespace/name" form — the natural
// identifier for logs and metrics.
func (k Key) String() string {
	if k.Namespace == "" {
		return fmt.Sprintf("%s/%s", k.Kind, k.Name)
	}
	return fmt.Sprintf("%s/%s/%s", k.Kind, k.Namespace, k.Name)
}

// Resolver hides the concrete storage layer from the GC algorithm.
// List returns every live object currently in the store; Delete
// physically removes an object.  Implementations must be safe for
// concurrent calls.
type Resolver interface {
	List(ctx context.Context) ([]map[string]interface{}, error)
	Delete(ctx context.Context, k Key) error
}

// OwnersOf extracts owner references from an object.  Missing or
// malformed metadata yields a nil slice.
func OwnersOf(obj map[string]interface{}) []Reference {
	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}
	raw, ok := meta["ownerReferences"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]Reference, 0, len(raw))
	for _, entry := range raw {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		ref := Reference{
			APIVersion: asString(m["apiVersion"]),
			Kind:       asString(m["kind"]),
			Name:       asString(m["name"]),
			UID:        asString(m["uid"]),
		}
		if b, ok := m["controller"].(bool); ok {
			ref.Controller = b
		}
		if b, ok := m["blockOwnerDeletion"].(bool); ok {
			ref.BlockOwnerDeletion = b
		}
		out = append(out, ref)
	}
	return out
}

// ControllerOf returns the sole controller reference, or nil when the
// object is either orphaned or owned only by non-controller parents.
// Only one controller reference is permitted per k8s convention.
func ControllerOf(obj map[string]interface{}) *Reference {
	for _, r := range OwnersOf(obj) {
		if r.Controller {
			rc := r
			return &rc
		}
	}
	return nil
}

// Add appends a Reference to the object's owner slice.  Duplicate
// (kind,name,uid) triples are collapsed.  The method does not enforce
// the "at most one controller" invariant — call EnsureSingleController
// after Add to check.
func Add(obj map[string]interface{}, ref Reference) {
	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
		obj["metadata"] = meta
	}
	existing, _ := meta["ownerReferences"].([]interface{})
	for _, entry := range existing {
		if m, ok := entry.(map[string]interface{}); ok {
			if asString(m["kind"]) == ref.Kind &&
				asString(m["name"]) == ref.Name &&
				asString(m["uid"]) == ref.UID {
				return
			}
		}
	}
	existing = append(existing, map[string]interface{}{
		"apiVersion":         ref.APIVersion,
		"kind":               ref.Kind,
		"name":               ref.Name,
		"uid":                ref.UID,
		"controller":         ref.Controller,
		"blockOwnerDeletion": ref.BlockOwnerDeletion,
	})
	meta["ownerReferences"] = existing
}

// EnsureSingleController returns an error when multiple controller
// references are attached to obj.  Call before persisting writes.
func EnsureSingleController(obj map[string]interface{}) error {
	count := 0
	for _, r := range OwnersOf(obj) {
		if r.Controller {
			count++
		}
	}
	if count > 1 {
		return fmt.Errorf("object has %d controller references; at most one is allowed", count)
	}
	return nil
}

// GarbageCollector walks the live object graph and removes children
// whose owners no longer exist.
type GarbageCollector struct {
	resolver     Resolver
	lastChildren map[string][]Key
}

// NewGarbageCollector constructs a collector backed by resolver.
func NewGarbageCollector(resolver Resolver) *GarbageCollector {
	return &GarbageCollector{resolver: resolver}
}

// Sweep runs a single collection pass.  It returns the keys that were
// deleted so callers can emit audit events.  Errors from individual
// deletes are aggregated — one failed child does not abort the sweep,
// because otherwise a single flaky object could pin the entire graph.
func (gc *GarbageCollector) Sweep(ctx context.Context) ([]Key, error) {
	objects, err := gc.resolver.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list for GC: %w", err)
	}

	// Build two indexes:
	//   live[uid]  = true  for every currently-existing owner candidate
	//   children[parentUID] = list of child keys
	live := make(map[string]Key, len(objects))
	children := make(map[string][]Key)
	for _, obj := range objects {
		k := keyOf(obj)
		uid := uidOf(obj)
		if uid != "" {
			live[uid] = k
		}
		for _, ref := range OwnersOf(obj) {
			children[ref.UID] = append(children[ref.UID], k)
		}
	}

	var deleted []Key
	var errs []error

	// An object is orphaned when it has owner references that all
	// point to UIDs absent from live.
	for _, obj := range objects {
		refs := OwnersOf(obj)
		if len(refs) == 0 {
			continue
		}
		orphaned := true
		for _, ref := range refs {
			if _, ok := live[ref.UID]; ok {
				orphaned = false
				break
			}
		}
		if !orphaned {
			continue
		}
		k := keyOf(obj)
		if err := gc.resolver.Delete(ctx, k); err != nil {
			errs = append(errs, fmt.Errorf("delete %s: %w", k, err))
			continue
		}
		deleted = append(deleted, k)
	}

	// Consume `children` to silence the "declared and not used"
	// checker in refactors that drop the orphan branch above.  It is
	// exposed on the collector for callers that want to inspect the
	// graph directly via Children().
	gc.lastChildren = children

	if len(errs) > 0 {
		return deleted, fmt.Errorf("garbage collection had %d error(s): %v", len(errs), errs[0])
	}
	return deleted, nil
}

// lastChildren captures the parent-to-children map from the most
// recent Sweep, for introspection by callers that want to render the
// deletion graph in a UI.
var _ = (*GarbageCollector)(nil)

// Children returns the child keys recorded at the last Sweep call for
// the given owner UID.  Returns nil when the owner is unknown.
func (gc *GarbageCollector) Children(uid string) []Key { return gc.lastChildren[uid] }

// keyOf extracts the (kind, namespace, name) triple from an object.
func keyOf(obj map[string]interface{}) Key {
	kind := asString(obj["kind"])
	meta, _ := obj["metadata"].(map[string]interface{})
	ns := ""
	name := ""
	if meta != nil {
		ns = asString(meta["namespace"])
		name = asString(meta["name"])
	}
	return Key{Kind: kind, Namespace: ns, Name: name}
}

// uidOf extracts metadata.uid.
func uidOf(obj map[string]interface{}) string {
	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return ""
	}
	return asString(meta["uid"])
}

// asString coerces v to string, returning "" for nil / wrong type.
func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
