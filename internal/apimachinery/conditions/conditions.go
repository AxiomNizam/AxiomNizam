// Package conditions implements the Kubernetes condition-set algebra
// described at
// https://github.com/kubernetes/apimachinery/blob/master/pkg/api/meta/conditions.go
// and standardised as `metav1.Condition` across the k8s ecosystem.
//
// A Condition is a short, machine-readable statement about the current
// state of a resource along a single axis.  Controllers publish one
// condition per axis they observe — typical axes are:
//
//   - Ready       : "does this object represent the intended state?"
//   - Available   : "is the underlying runtime serving traffic?"
//   - Progressing : "is a rollout currently advancing?"
//   - Degraded    : "is the object partially functional?"
//
// The invariants enforced here mirror the upstream k8s implementation:
//
//  1. A condition Type appears at most once in the slice.
//  2. LastTransitionTime changes only when Status changes.
//  3. ObservedGeneration records the spec generation at which the
//     condition was last evaluated — consumers use it to detect stale
//     conditions that predate the latest spec change.
//
// This package is consumed by status trackers, API handlers, and by
// `kubectl describe`-equivalent frontends.
package conditions

import (
	"sort"
	"time"
)

// Status is the tri-valued health of a Condition.
type Status string

const (
	// StatusTrue asserts the condition holds.
	StatusTrue Status = "True"
	// StatusFalse asserts the condition is violated.
	StatusFalse Status = "False"
	// StatusUnknown indicates the controller could not determine the
	// condition.  Used during startup or when observing a dependency
	// that has not yet reported.
	StatusUnknown Status = "Unknown"
)

// Condition is a single observation about a resource along one axis.
// It mirrors k8s.io/apimachinery/pkg/apis/meta/v1.Condition field-for-
// field so that JSON documents are cross-compatible with kubectl.
type Condition struct {
	// Type is a short PascalCase token — "Ready", "Available".
	Type string `json:"type"`

	// Status is True / False / Unknown.
	Status Status `json:"status"`

	// ObservedGeneration records the resource generation observed
	// when this condition was last set.  Consumers compare against
	// metadata.generation to detect stale conditions.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LastTransitionTime is updated only when Status changes.
	LastTransitionTime time.Time `json:"lastTransitionTime"`

	// Reason is a short machine-readable CamelCase token —
	// "ReconcileSucceeded", "BackendUnreachable".
	Reason string `json:"reason"`

	// Message is a human-readable explanation.
	Message string `json:"message,omitempty"`
}

// Set inserts or updates cond in conditions.  The returned slice is the
// new canonical form, sorted by Type for deterministic JSON output.
// LastTransitionTime is updated only when the Status value changes —
// re-asserting the same Status preserves the original transition time
// so "has been Ready for 42m" telemetry remains accurate.
func Set(conditions []Condition, cond Condition) []Condition {
	if cond.LastTransitionTime.IsZero() {
		cond.LastTransitionTime = time.Now().UTC()
	}
	if cond.Status == "" {
		cond.Status = StatusUnknown
	}

	out := make([]Condition, 0, len(conditions)+1)
	found := false
	for _, existing := range conditions {
		if existing.Type != cond.Type {
			out = append(out, existing)
			continue
		}
		found = true
		// Preserve transition time when status has not flipped.
		if existing.Status == cond.Status {
			cond.LastTransitionTime = existing.LastTransitionTime
		}
		out = append(out, cond)
	}
	if !found {
		out = append(out, cond)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Type < out[j].Type })
	return out
}

// Remove deletes the condition with the given Type.  No-op when not
// found.  The returned slice is newly allocated and safe to store.
func Remove(conditions []Condition, condType string) []Condition {
	out := make([]Condition, 0, len(conditions))
	for _, c := range conditions {
		if c.Type != condType {
			out = append(out, c)
		}
	}
	return out
}

// Find returns the condition with the given Type, or nil when absent.
// The returned pointer aliases the slice element; callers that intend
// to mutate it must copy first to avoid races with concurrent readers.
func Find(conditions []Condition, condType string) *Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

// IsTrue reports whether the condition of the given Type is Status=True.
// A missing condition is treated as Unknown, not True.
func IsTrue(conditions []Condition, condType string) bool {
	c := Find(conditions, condType)
	return c != nil && c.Status == StatusTrue
}

// IsFalse reports whether the condition of the given Type is Status=False.
func IsFalse(conditions []Condition, condType string) bool {
	c := Find(conditions, condType)
	return c != nil && c.Status == StatusFalse
}

// IsUnknown reports whether the condition is missing or StatusUnknown.
func IsUnknown(conditions []Condition, condType string) bool {
	c := Find(conditions, condType)
	return c == nil || c.Status == StatusUnknown
}

// IsStatusEqualTo is the general form of IsTrue / IsFalse.
func IsStatusEqualTo(conditions []Condition, condType string, status Status) bool {
	c := Find(conditions, condType)
	return c != nil && c.Status == status
}

// IsStale returns true when the condition's ObservedGeneration is
// behind the supplied current generation.  Callers use this to decide
// whether to trust a condition for admission / rollout gating.
func IsStale(conditions []Condition, condType string, currentGeneration int64) bool {
	c := Find(conditions, condType)
	if c == nil {
		return true
	}
	return c.ObservedGeneration < currentGeneration
}

// Standard condition-type constants used by AxiomNizam controllers.
// Controllers may introduce additional domain-specific types, but the
// names below have agreed semantics and must not be redefined.
const (
	// TypeReady asserts the object represents its intended state and
	// is usable by consumers.
	TypeReady = "Ready"

	// TypeAvailable asserts the underlying runtime resource is online.
	// An object may be Ready=True but Available=False briefly during
	// leader-election failover.
	TypeAvailable = "Available"

	// TypeProgressing asserts a rollout or reconciliation is currently
	// advancing.  Flips to False when steady state is reached OR when
	// progress stalls.
	TypeProgressing = "Progressing"

	// TypeDegraded asserts the object is partially functional.  Used
	// when some replicas / partitions have failed but aggregate
	// availability remains.
	TypeDegraded = "Degraded"

	// TypeReconcileSucceeded asserts the most recent reconcile loop
	// completed without error.
	TypeReconcileSucceeded = "ReconcileSucceeded"
)
