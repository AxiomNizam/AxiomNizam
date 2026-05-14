// Package admission implements a Kubernetes-style admission chain for
// AxiomNizam resource writes.  Resource handlers invoke the chain
// before persisting a create/update/delete request; the chain runs
// registered mutating webhooks (which may rewrite the object) followed
// by validating webhooks (which may reject the operation).  Rejections
// are surfaced to the caller as a typed error carrying the offending
// webhook's identifier and message.
//
// # Design notes
//
// This implementation intentionally keeps the "webhook" concept
// in-process: each Validator/Mutator is a Go type registered with a
// Chain.  Out-of-process admission (the HTTPS webhook model used by
// apiserver) can be layered on top by providing an implementation of
// Validator/Mutator that marshals the request, POSTs it to the remote
// URL, and returns the response verbatim — but the in-process model
// avoids the latency, PKI, and availability concerns that make
// apiserver's webhook path notoriously operationally expensive.
//
// The chain is schema-agnostic: it operates on *Request values whose
// Object field is an opaque map[string]interface{}.  Resource-specific
// plugins are responsible for inspecting / mutating that map.
package admission

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Operation enumerates the mutation verbs that may be admitted.
type Operation string

const (
	// OperationCreate is a POST / create-resource request.
	OperationCreate Operation = "CREATE"
	// OperationUpdate is a PATCH/PUT / update-resource request.
	OperationUpdate Operation = "UPDATE"
	// OperationDelete is a DELETE request.
	OperationDelete Operation = "DELETE"
)

// Request is the payload passed to every plugin in the chain.  It is
// modelled on k8s admission.v1 AdmissionRequest but trimmed to the
// fields AxiomNizam actually uses.
type Request struct {
	// UID is a random token used to correlate chain entries in logs.
	UID string

	// Group / Version / Kind identify the resource's schema.  The
	// built-in resources use Group "axiomnizam.io".
	Group, Version, Kind string

	// Namespace and Name identify the individual resource instance.
	// Namespace is empty for cluster-scoped kinds.
	Namespace, Name string

	// Operation is the verb being admitted.
	Operation Operation

	// Object is the incoming object as a map.  Mutating plugins SHOULD
	// modify this map in place; plugins MUST NOT replace the map
	// reference itself (callers hold a reference to the original).
	Object map[string]interface{}

	// OldObject is populated for Update and Delete operations.  It is
	// read-only from the plugin's perspective.
	OldObject map[string]interface{}

	// UserInfo carries the authenticated caller — plugins that enforce
	// RBAC-style checks consult this field.
	UserInfo UserInfo
}

// UserInfo describes the caller of the admission request.
type UserInfo struct {
	Username string
	UID      string
	Groups   []string
	Extra    map[string][]string
}

// Validator is a plugin that may reject a request.  Name is used in
// error messages and in the chain's ordering algorithm.
type Validator interface {
	Name() string
	Validate(ctx context.Context, req *Request) error
}

// Mutator is a plugin that may modify a request's Object in place.
// Implementations SHOULD be idempotent so that running the mutator twice
// on the same object yields the same result (kube-apiserver reinvokes
// mutators when another mutator upstream rewrote the object).
type Mutator interface {
	Name() string
	Mutate(ctx context.Context, req *Request) error
}

// RejectionError is returned by Chain.Admit when a validator refuses a
// request.  Callers can type-assert to extract the failing plugin.
type RejectionError struct {
	PluginName string
	Reason     string
}

// Error implements the error interface.
func (e *RejectionError) Error() string {
	return fmt.Sprintf("admission rejected by %q: %s", e.PluginName, e.Reason)
}

// IsRejection reports whether err is an admission rejection.
func IsRejection(err error) bool {
	var r *RejectionError
	return errors.As(err, &r)
}

// Chain holds the ordered list of mutators and validators.  It is safe
// for concurrent Admit calls; Register* calls should be made at startup.
type Chain struct {
	mu         sync.RWMutex
	mutators   []Mutator
	validators []Validator
	timeout    time.Duration
}

// NewChain constructs an empty Chain.  The supplied per-plugin timeout
// is applied independently to each Mutate / Validate invocation.
func NewChain(perPluginTimeout time.Duration) *Chain {
	if perPluginTimeout <= 0 {
		perPluginTimeout = 3 * time.Second
	}
	return &Chain{timeout: perPluginTimeout}
}

// RegisterMutator appends m to the mutator list.  Mutators run in
// registration order; callers that care about ordering (e.g. "defaulter
// must run before injector") should register in the desired sequence.
func (c *Chain) RegisterMutator(m Mutator) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mutators = append(c.mutators, m)
}

// RegisterValidator appends v to the validator list.
func (c *Chain) RegisterValidator(v Validator) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.validators = append(c.validators, v)
}

// Admit runs the full chain for req.  The mutators run first in order;
// the first mutator error aborts the chain with a wrapped error.  The
// validators run next; the first validator rejection aborts the chain
// with the corresponding RejectionError.  When all plugins succeed the
// caller may proceed to persist req.Object.
func (c *Chain) Admit(ctx context.Context, req *Request) error {
	c.mu.RLock()
	mutators := append([]Mutator(nil), c.mutators...)
	validators := append([]Validator(nil), c.validators...)
	timeout := c.timeout
	c.mu.RUnlock()

	for _, m := range mutators {
		pCtx, cancel := context.WithTimeout(ctx, timeout)
		err := m.Mutate(pCtx, req)
		cancel()
		if err != nil {
			return fmt.Errorf("mutator %q failed: %w", m.Name(), err)
		}
	}

	for _, v := range validators {
		pCtx, cancel := context.WithTimeout(ctx, timeout)
		err := v.Validate(pCtx, req)
		cancel()
		if err != nil {
			// Wrap non-RejectionError results so callers get a stable
			// type to test against.
			if !IsRejection(err) {
				err = &RejectionError{PluginName: v.Name(), Reason: err.Error()}
			}
			return err
		}
	}
	return nil
}
