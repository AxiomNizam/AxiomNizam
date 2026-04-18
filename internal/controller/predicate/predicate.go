// Package predicate filters events before they reach the controller's
// workqueue.  The canonical use is "ignore status-only updates" so a
// controller that writes back its own status observations does not
// endlessly re-trigger itself.
//
// A Predicate is a tuple of four functions — one per event kind.
// Returning false from any of them drops the event.  Predicates can
// be chained with And/Or for composite behaviour.
package predicate

// EventKind discriminates the four event types a controller cares
// about.  Delete events arrive after the object has already been
// removed from the cache, hence carry the prior object snapshot.
type EventKind int

const (
	// Create fires once when the informer first sees an object.
	Create EventKind = iota
	// Update fires whenever the cache observes a change.
	Update
	// Delete fires when the object is removed.
	Delete
	// Generic fires for external signals (channel sources).
	Generic
)

// Object is the minimal surface a Predicate sees.  Controllers pass
// the schemaless map[string]interface{} form; matching is done on
// labels, annotations, and generation just like upstream k8s.
type Object interface {
	// GetLabels returns metadata.labels (or nil if unset).
	GetLabels() map[string]string
	// GetAnnotations returns metadata.annotations.
	GetAnnotations() map[string]string
	// GetGeneration returns metadata.generation — server-set integer
	// that increments each time spec changes.
	GetGeneration() int64
	// GetResourceVersion returns metadata.resourceVersion.
	GetResourceVersion() string
}

// Predicate is the filter contract.  Old is nil for Create, non-nil
// and distinct from New for Update, non-nil for Delete, nil for
// Generic.
type Predicate interface {
	// Accept decides whether the event should flow downstream.
	Accept(kind EventKind, old, new Object) bool
}

// Funcs is a four-field Predicate usable as a value literal.  Unset
// fields default to "accept" so callers only override the kinds they
// care about.
type Funcs struct {
	CreateFunc  func(Object) bool
	UpdateFunc  func(old, new Object) bool
	DeleteFunc  func(Object) bool
	GenericFunc func(Object) bool
}

// Accept dispatches to the corresponding field.
func (f Funcs) Accept(kind EventKind, old, newObj Object) bool {
	switch kind {
	case Create:
		if f.CreateFunc != nil {
			return f.CreateFunc(newObj)
		}
	case Update:
		if f.UpdateFunc != nil {
			return f.UpdateFunc(old, newObj)
		}
	case Delete:
		if f.DeleteFunc != nil {
			return f.DeleteFunc(old)
		}
	case Generic:
		if f.GenericFunc != nil {
			return f.GenericFunc(newObj)
		}
	}
	return true
}

// GenerationChangedPredicate drops Update events whose generation is
// unchanged — the standard guard against status-only self-triggers.
// CreateFunc and DeleteFunc accept by design; Generic events have no
// generation data so also accept.
type GenerationChangedPredicate struct{}

// Accept implements Predicate.
func (GenerationChangedPredicate) Accept(kind EventKind, old, newObj Object) bool {
	if kind != Update {
		return true
	}
	if old == nil || newObj == nil {
		return true
	}
	return old.GetGeneration() != newObj.GetGeneration()
}

// LabelChangedPredicate drops Update events whose labels map is
// byte-identical.  Cheaper than a full reconcile when the controller
// only reacts to label changes.
type LabelChangedPredicate struct{}

// Accept implements Predicate.
func (LabelChangedPredicate) Accept(kind EventKind, old, newObj Object) bool {
	if kind != Update || old == nil || newObj == nil {
		return true
	}
	return !sameStringMap(old.GetLabels(), newObj.GetLabels())
}

// sameStringMap is a local helper — avoids a dependency on reflect
// for a very hot path.
func sameStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// And returns a Predicate that accepts only when every member does.
func And(ps ...Predicate) Predicate { return andPredicate(ps) }

// Or returns a Predicate that accepts when any member does.
func Or(ps ...Predicate) Predicate { return orPredicate(ps) }

type andPredicate []Predicate

// Accept iterates until the first rejection.
func (a andPredicate) Accept(kind EventKind, old, newObj Object) bool {
	for _, p := range a {
		if !p.Accept(kind, old, newObj) {
			return false
		}
	}
	return true
}

type orPredicate []Predicate

// Accept iterates until the first acceptance.
func (o orPredicate) Accept(kind EventKind, old, newObj Object) bool {
	for _, p := range o {
		if p.Accept(kind, old, newObj) {
			return true
		}
	}
	return len(o) == 0 // empty Or is vacuously true, matching And's semantics.
}
