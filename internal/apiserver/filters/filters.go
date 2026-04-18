// Package filters implements the apiserver's HTTP middleware chain.
// The order matches the upstream kube-apiserver: outermost filter
// runs first, innermost runs last before the handler.  The canonical
// chain, outside in:
//
//  1. Panic recovery (so every later filter's failure is logged).
//  2. Request info extraction (derives verb, resource, namespace
//     from the URL path).
//  3. Timeout  — attach a per-request deadline.
//  4. Rate limit — reject at the door when overloaded.
//  5. Authentication — identify the caller.
//  6. Authorization — check the authenticated caller is permitted.
//  7. Audit     — record the request/response tuple.
//  8. Handler   — the actual route.
//
// Each filter is a net/http middleware and composes freely with any
// router AxiomNizam chooses.  We deliberately avoid a single
// monolithic chain constructor — callers compose what they need.
package filters

import "net/http"

// Chain wraps h with the provided middlewares, applying them in
// reverse order so the first argument is the outermost filter.
// Equivalent to calling each filter as f1(f2(f3(handler))).
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
