// Package filters — per-request timeout.
//
// The middleware attaches a context.WithTimeout to the request so any
// handler doing context-aware work (DB queries, downstream RPCs) gets
// cancelled when the budget is exceeded.  If the handler returns
// before the timeout, the context is cancelled immediately on the way
// out — preventing goroutine leaks in downstream fan-out logic.
package filters

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Timeout returns a middleware that bounds every request to d.  The
// configurable error message lets callers distinguish apiserver
// timeouts from downstream service timeouts in client logs.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return TimeoutWithMessage(d, fmt.Sprintf("request exceeded the %s apiserver timeout", d))
}

// TimeoutWithMessage is Timeout with a custom 503 body.
func TimeoutWithMessage(d time.Duration, msg string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if d <= 0 {
				h.ServeHTTP(w, r)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			r = r.WithContext(ctx)
			done := make(chan struct{})
			go func() {
				defer close(done)
				h.ServeHTTP(w, r)
			}()
			select {
			case <-done:
				return
			case <-ctx.Done():
				// The handler goroutine may still be running — net/http
				// does not give us a way to abort it.  We surface the
				// timeout to the client and leave the goroutine to
				// complete (its writes to w will be discarded because
				// the client connection is torn down).
				http.Error(w, msg, http.StatusServiceUnavailable)
				<-done
			}
		})
	}
}
