// Package filters — RequestInfo derives the REST tuple (APIGroup,
// Version, Resource, Namespace, Name, Verb) from the URL path and
// stashes it in the request context for every later filter and the
// handler to consume.  Matches the k8s apiserver grammar:
//
//	/api/v1/namespaces/{ns}/{resource}[/{name}[/{subresource}]]
//	/apis/{group}/{version}/namespaces/{ns}/{resource}[/{name}...]
//	/apis/{group}/{version}/{cluster-scoped-resource}[/{name}...]
//
// This extraction is done once at the outermost layer so authz and
// audit filters don't re-parse the URL.
package filters

import (
	"context"
	"net/http"
	"strings"
)

// RequestInfo is the structured request descriptor.
type RequestInfo struct {
	// IsResourceRequest is false for non-resource URLs (/healthz etc).
	IsResourceRequest bool
	// Verb is the REST verb: get, list, create, update, patch, delete, watch.
	Verb string
	// APIPrefix is "api" or "apis".
	APIPrefix string
	// APIGroup is "" for the core group, else e.g. "apps".
	APIGroup string
	// APIVersion is the group version, e.g. "v1", "v1beta1".
	APIVersion string
	// Resource is the resource name: pods, services, etc.
	Resource string
	// Subresource is e.g. "status", "scale", or "" for root.
	Subresource string
	// Namespace is "" for cluster-scoped or non-namespaced verbs.
	Namespace string
	// Name is the object name, or "" for list/create.
	Name string
	// Path is the raw URL path (for logs/audit).
	Path string
}

// requestInfoKey is the context key type.  Using a package-private
// type prevents clashes with other packages' context keys.
type requestInfoKey struct{}

// WithRequestInfo attaches ri to ctx.
func WithRequestInfo(ctx context.Context, ri *RequestInfo) context.Context {
	return context.WithValue(ctx, requestInfoKey{}, ri)
}

// FromContext returns the RequestInfo stored by the middleware, or
// nil if the middleware did not run.
func FromContext(ctx context.Context) *RequestInfo {
	ri, _ := ctx.Value(requestInfoKey{}).(*RequestInfo)
	return ri
}

// RequestInfoMiddleware parses the path and attaches the result.
func RequestInfoMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ri := parseRequestInfo(r)
			r = r.WithContext(WithRequestInfo(r.Context(), ri))
			h.ServeHTTP(w, r)
		})
	}
}

// parseRequestInfo walks the path segments.
func parseRequestInfo(r *http.Request) *RequestInfo {
	ri := &RequestInfo{Path: r.URL.Path}
	segs := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(segs) == 0 || (segs[0] != "api" && segs[0] != "apis") {
		ri.Verb = verbFromMethod(r.Method, false)
		return ri
	}
	ri.IsResourceRequest = true
	ri.APIPrefix = segs[0]
	i := 1
	if segs[0] == "api" {
		if len(segs) > i {
			ri.APIVersion = segs[i]
			i++
		}
	} else {
		if len(segs) > i {
			ri.APIGroup = segs[i]
			i++
		}
		if len(segs) > i {
			ri.APIVersion = segs[i]
			i++
		}
	}
	if i < len(segs) && segs[i] == "namespaces" {
		i++
		if i < len(segs) {
			ri.Namespace = segs[i]
			i++
		}
	}
	if i < len(segs) {
		ri.Resource = segs[i]
		i++
	}
	if i < len(segs) {
		ri.Name = segs[i]
		i++
	}
	if i < len(segs) {
		ri.Subresource = segs[i]
	}
	hasName := ri.Name != ""
	ri.Verb = verbFromMethod(r.Method, hasName)
	// `watch=true` query param overrides get/list into watch.
	if r.URL.Query().Get("watch") == "true" {
		ri.Verb = "watch"
	}
	return ri
}

// verbFromMethod maps HTTP methods to REST verbs.  Distinguishes
// get-by-name from list based on whether the path had a name segment.
func verbFromMethod(method string, hasName bool) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		if hasName {
			return "get"
		}
		return "list"
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodPatch:
		return "patch"
	case http.MethodDelete:
		if hasName {
			return "delete"
		}
		return "deletecollection"
	}
	return strings.ToLower(method)
}
