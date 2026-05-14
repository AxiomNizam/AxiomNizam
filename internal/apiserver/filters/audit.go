// Package filters — audit middleware.
//
// Every request produces one Event describing (who, what, when,
// from-where, to-what, outcome).  The middleware captures the
// request-side fields before the handler runs, buffers the response
// status via a responseWriter wrapper, and emits the Event to a
// Sink after the handler returns.
//
// The filter deliberately does not log request/response bodies —
// that's a separate concern handled by the body-audit policy engine.
// Keeping this filter body-agnostic means every request pays the
// same fixed CPU cost regardless of payload size.
package filters

import (
	"net/http"
	"time"
)

// Event is the audit record.
type Event struct {
	Timestamp         time.Time
	UserName          string
	UserGroups        []string
	Verb              string
	APIGroup          string
	APIVersion        string
	Resource          string
	Subresource       string
	Namespace         string
	Name              string
	RequestURI        string
	Method            string
	SourceIP          string
	UserAgent         string
	ResponseStatus    int
	ResponseSizeBytes int64
	Duration          time.Duration
}

// Sink is the audit destination — a file, a Kafka topic, a SIEM.
type Sink interface {
	Record(e Event)
}

// SinkFunc adapts a function.
type SinkFunc func(e Event)

// Record satisfies Sink.
func (f SinkFunc) Record(e Event) { f(e) }

// Audit returns the middleware.  sink may not be nil.
func Audit(sink Sink) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			h.ServeHTTP(rw, r)
			evt := Event{
				Timestamp:         start,
				RequestURI:        r.RequestURI,
				Method:            r.Method,
				SourceIP:          clientIP(r),
				UserAgent:         r.UserAgent(),
				ResponseStatus:    rw.status,
				ResponseSizeBytes: rw.bytes,
				Duration:          time.Since(start),
			}
			if user := UserFrom(r.Context()); user != nil {
				evt.UserName = user.Name
				evt.UserGroups = user.Groups
			}
			if ri := FromContext(r.Context()); ri != nil {
				evt.Verb = ri.Verb
				evt.APIGroup = ri.APIGroup
				evt.APIVersion = ri.APIVersion
				evt.Resource = ri.Resource
				evt.Subresource = ri.Subresource
				evt.Namespace = ri.Namespace
				evt.Name = ri.Name
			}
			sink.Record(evt)
		})
	}
}

// statusRecorder captures the response status code and byte count
// without buffering the body.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int64
	// wroteHeader tracks whether WriteHeader was called explicitly;
	// implicit 200 on first Write still counts, matching the
	// net/http contract.
	wroteHeader bool
}

// WriteHeader records the status.
func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.wroteHeader = true
	s.ResponseWriter.WriteHeader(code)
}

// Write tallies bytes and forwards to the underlying writer.
func (s *statusRecorder) Write(p []byte) (int, error) {
	if !s.wroteHeader {
		s.wroteHeader = true // implicit 200
	}
	n, err := s.ResponseWriter.Write(p)
	s.bytes += int64(n)
	return n, err
}

// clientIP extracts the remote IP honouring X-Forwarded-For if the
// proxy in front of the apiserver is trusted.  Callers that do not
// run behind a trusted proxy should configure the apiserver to strip
// this header; this helper just returns whatever is there.
func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		// First entry is the original client.
		for i := 0; i < len(xf); i++ {
			if xf[i] == ',' {
				return xf[:i]
			}
		}
		return xf
	}
	return r.RemoteAddr
}
