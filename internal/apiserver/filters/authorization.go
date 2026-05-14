// Package filters — authorization middleware.
//
// The filter reads the authenticated User and the parsed RequestInfo
// from the context and asks an Authorizer whether the action is
// permitted.  Decisions are (Allow, Deny, NoOpinion); chains of
// authorizers combine via the Union helper — the first non-NoOpinion
// decision wins.
package filters

import (
	"net/http"
)

// Decision is the authorizer verdict.
type Decision int

const (
	// DecisionNoOpinion means "pass to the next authorizer in the chain".
	DecisionNoOpinion Decision = iota
	// DecisionAllow means "permit the request".
	DecisionAllow
	// DecisionDeny means "reject the request with 403".
	DecisionDeny
)

// Attributes describes the action being authorized.  Built from the
// User + RequestInfo + request headers.
type Attributes struct {
	User            *User
	Verb            string
	APIGroup        string
	APIVersion      string
	Resource        string
	Subresource     string
	Namespace       string
	Name            string
	ResourceRequest bool
	Path            string
}

// Authorizer is the decision delegate.  Returning a non-empty reason
// is encouraged — it shows up in 403 bodies and audit logs.
type Authorizer interface {
	Authorize(a Attributes) (Decision, string, error)
}

// AuthorizerFunc adapts a function.
type AuthorizerFunc func(a Attributes) (Decision, string, error)

// Authorize satisfies Authorizer.
func (f AuthorizerFunc) Authorize(a Attributes) (Decision, string, error) { return f(a) }

// Union chains authorizers — first non-NoOpinion wins.  If every
// authorizer abstains, the union returns NoOpinion and the middleware
// denies by default.
func Union(authzs ...Authorizer) Authorizer {
	return AuthorizerFunc(func(a Attributes) (Decision, string, error) {
		for _, az := range authzs {
			d, reason, err := az.Authorize(a)
			if err != nil {
				return DecisionDeny, reason, err
			}
			if d != DecisionNoOpinion {
				return d, reason, nil
			}
		}
		return DecisionNoOpinion, "no authorizer expressed an opinion", nil
	})
}

// Authorization returns the middleware.  Authorizer may not be nil.
func Authorization(az Authorizer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFrom(r.Context())
			ri := FromContext(r.Context())
			attrs := Attributes{User: user}
			if ri != nil {
				attrs.Verb = ri.Verb
				attrs.APIGroup = ri.APIGroup
				attrs.APIVersion = ri.APIVersion
				attrs.Resource = ri.Resource
				attrs.Subresource = ri.Subresource
				attrs.Namespace = ri.Namespace
				attrs.Name = ri.Name
				attrs.ResourceRequest = ri.IsResourceRequest
				attrs.Path = ri.Path
			}
			decision, reason, err := az.Authorize(attrs)
			if err != nil || decision != DecisionAllow {
				msg := reason
				if msg == "" {
					msg = "forbidden"
				}
				http.Error(w, msg, http.StatusForbidden)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
