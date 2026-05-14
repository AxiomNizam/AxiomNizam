// Package errors provides the Aggregate type — a single error that
// carries an ordered list of underlying errors.  This pattern is
// pervasive in k8s: when ten validators run against one object or
// ten stores run in parallel, callers want every failure in one
// response rather than the first one only.
//
// The shape mirrors k8s.io/apimachinery/pkg/util/errors.Aggregate so
// that code written against that interface can be ported by swapping
// the import path.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// Aggregate represents a bundle of errors.  It implements the error
// interface so it can be returned anywhere an error is expected.
type Aggregate interface {
	error
	Errors() []error
	Is(target error) bool
}

// aggregate is the concrete implementation.  A nil or empty slice is
// equivalent to no error — see NewAggregate.
type aggregate []error

// NewAggregate returns nil for an empty / all-nil input list; it
// returns the single error when only one non-nil entry is present
// (avoiding a superfluous "[...]" wrap); otherwise it returns a true
// Aggregate.
func NewAggregate(errs []error) Aggregate {
	if len(errs) == 0 {
		return nil
	}
	filtered := make([]error, 0, len(errs))
	for _, e := range errs {
		if e != nil {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return aggregate(filtered)
}

// Error renders "[a; b; c]".  Duplicate messages are collapsed so
// that fan-out operations which surface the same error from every
// replica do not bury the signal.
func (agg aggregate) Error() string {
	if len(agg) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(agg))
	parts := make([]string, 0, len(agg))
	for _, e := range agg {
		m := e.Error()
		if _, dup := seen[m]; dup {
			continue
		}
		seen[m] = struct{}{}
		parts = append(parts, m)
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "[" + strings.Join(parts, "; ") + "]"
}

// Errors returns the underlying slice.  Callers may not mutate it.
func (agg aggregate) Errors() []error { return []error(agg) }

// Is walks the aggregate tree so errors.Is works transitively.  This
// matches the behaviour of the fmt.Errorf("%w") chain — errors.Is
// returns true when target appears anywhere in the bundle.
func (agg aggregate) Is(target error) bool {
	for _, e := range agg {
		if errors.Is(e, target) {
			return true
		}
	}
	return false
}

// Flatten walks the aggregate tree and returns every leaf error as a
// single flat list.  Useful when aggregates nest several layers deep
// (e.g. each sub-reconciler returns its own Aggregate, and the
// controller aggregates *those*).
func Flatten(agg Aggregate) Aggregate {
	if agg == nil {
		return nil
	}
	var flat []error
	var walk func(e error)
	walk = func(e error) {
		if e == nil {
			return
		}
		if inner, ok := e.(Aggregate); ok {
			for _, sub := range inner.Errors() {
				walk(sub)
			}
			return
		}
		flat = append(flat, e)
	}
	for _, e := range agg.Errors() {
		walk(e)
	}
	return NewAggregate(flat)
}

// Matcher returns the subset of agg for which predicate(err) is true.
// Useful for separating recoverable vs fatal errors from a batch.
func Matcher(agg Aggregate, predicate func(error) bool) Aggregate {
	if agg == nil {
		return nil
	}
	var matched []error
	for _, e := range agg.Errors() {
		if predicate(e) {
			matched = append(matched, e)
		}
	}
	return NewAggregate(matched)
}

// FilterOut returns agg with every error matching predicate removed.
// Complement of Matcher.
func FilterOut(agg Aggregate, predicate func(error) bool) Aggregate {
	if agg == nil {
		return nil
	}
	var kept []error
	for _, e := range agg.Errors() {
		if !predicate(e) {
			kept = append(kept, e)
		}
	}
	return NewAggregate(kept)
}

// Reduce returns nil if agg is empty, the single inner error if agg
// has one member, or agg otherwise.  Use at the end of a batch to
// present the simplest possible value to the caller.
func Reduce(agg Aggregate) error {
	if agg == nil {
		return nil
	}
	errs := agg.Errors()
	if len(errs) == 1 {
		return errs[0]
	}
	return agg
}

// AggregateGoroutines runs fn N times in parallel goroutines and
// collects their errors.  The classic k8s helper — here so callers
// don't reinvent the fan-out + sync.WaitGroup pattern.
func AggregateGoroutines(fn func() error, n int) Aggregate {
	if n <= 0 {
		return nil
	}
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic: %v", r)
				}
			}()
			errCh <- fn()
		}()
	}
	errs := make([]error, 0, n)
	for i := 0; i < n; i++ {
		if e := <-errCh; e != nil {
			errs = append(errs, e)
		}
	}
	return NewAggregate(errs)
}
