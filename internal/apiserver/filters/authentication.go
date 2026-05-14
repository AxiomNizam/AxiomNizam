// Package filters — authentication middleware.
//
// The filter delegates the actual credential validation to a
// pluggable Authenticator (bearer token, mTLS, OIDC, HMAC, etc).  On
// success the resolved User is stashed in the request context; on
// failure the response is 401 with a WWW-Authenticate header so
// clients know which scheme to retry with.
package filters

import (
	"context"
	"errors"
	"net/http"
)

// User is the caller identity every authenticator produces.  Kept
// intentionally small — groups + extras cover the downstream authz
// needs without committing to a richer identity model.
type User struct {
	// Name is the stable principal identifier (e.g. email, SPIFFE ID).
	Name string
	// UID is an opaque stable ID — useful for audit correlation
	// when Name can be re-assigned.
	UID string
	// Groups are the authorization groups this user belongs to.
	Groups []string
	// Extra carries authenticator-specific claims (scopes, tenant IDs).
	Extra map[string][]string
}

// Authenticator is the delegate contract.  Returning (nil, false, nil)
// means "no credentials presented" — the filter responds 401 with
// the challenge.  Returning (nil, false, err) is a hard auth failure
// (bad signature, expired token) — same response, err logged.
// Returning (user, true, nil) is success.
type Authenticator interface {
	AuthenticateRequest(r *http.Request) (*User, bool, error)
}

// AuthenticatorFunc adapts a function.
type AuthenticatorFunc func(r *http.Request) (*User, bool, error)

// AuthenticateRequest satisfies Authenticator.
func (f AuthenticatorFunc) AuthenticateRequest(r *http.Request) (*User, bool, error) {
	return f(r)
}

// userKey is the context key type.
type userKey struct{}

// WithUser attaches the authenticated user to ctx.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

// UserFrom returns the authenticated user, or nil.
func UserFrom(ctx context.Context) *User {
	u, _ := ctx.Value(userKey{}).(*User)
	return u
}

// Authentication returns a middleware configured with the supplied
// authenticator and WWW-Authenticate challenge value.  Anonymous is
// permitted when AllowAnonymous=true — useful for /healthz.
type AuthNOptions struct {
	Authenticator  Authenticator
	Challenge      string          // e.g. `Bearer realm="axiomnizam"`.
	AllowAnonymous bool            // routes that don't need auth.
	AnonymousPaths map[string]bool // exact-match allow list.
}

// Authentication returns the middleware.
func Authentication(opts AuthNOptions) func(http.Handler) http.Handler {
	if opts.Challenge == "" {
		opts.Challenge = `Bearer realm="axiomnizam"`
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.AllowAnonymous || opts.AnonymousPaths[r.URL.Path] {
				h.ServeHTTP(w, r)
				return
			}
			if opts.Authenticator == nil {
				unauthorized(w, opts.Challenge, errors.New("no authenticator configured"))
				return
			}
			user, ok, err := opts.Authenticator.AuthenticateRequest(r)
			if err != nil || !ok {
				unauthorized(w, opts.Challenge, err)
				return
			}
			r = r.WithContext(WithUser(r.Context(), user))
			h.ServeHTTP(w, r)
		})
	}
}

// unauthorized writes a consistent 401 response.  The err argument
// is deliberately ignored for the client body — we only emit a
// generic message to avoid leaking authenticator internals.
func unauthorized(w http.ResponseWriter, challenge string, _ error) {
	w.Header().Set("WWW-Authenticate", challenge)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
