// Package filters — panic recovery.
//
// Any panic in a downstream handler is caught, logged via the
// injected logger, and surfaced to the client as HTTP 500.  Without
// this filter a single handler bug takes the entire listener down.
package filters

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Logger is the narrow logging interface filters use.  Callers wire
// their structured logger of choice by adapting it to this shape.
type Logger interface {
	Errorf(format string, args ...interface{})
}

// stdoutLogger is the fallback when the caller does not supply one.
type stdoutLogger struct{}

// Errorf writes to os.Stderr via fmt.
func (stdoutLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("ERROR: "+format+"\n", args...)
}

// Recovery returns a middleware that catches panics from h.
func Recovery(log Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = stdoutLogger{}
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					// Capture stack for debuggability; the client
					// only ever sees a generic 500 to avoid leaking
					// stack frames to untrusted callers.
					log.Errorf("panic handling %s %s: %v\n%s", r.Method, r.URL.Path, v, debug.Stack())
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}
