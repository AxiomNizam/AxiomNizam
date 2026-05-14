// Package stream — NDJSON HTTP handler that streams a Subscription
// out over chunked transfer encoding.  This is the shape Nomad uses
// for its /v1/event/stream endpoint and AxiomNizam exposes via
// /v1/stream to give operators a live tail of resource changes.
package stream

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// HTTPHandler returns an http.Handler that opens a Subscription
// against broker and flushes each matching event as a JSON line.
// Query parameters:
//
//	topic=<name>     may repeat
//	key=<ns>/<name>  may repeat
//	index=<uint>     resume index (0 = from-now)
//
// The handler disables HTTP buffering where possible and sends
// heartbeats only via events themselves — callers who want dead-
// connection detection should layer TCP keepalives at the listener.
func HTTPHandler(broker *Broker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		req := SubscribeReq{
			Topics: dropEmpty(q["topic"]),
			Keys:   dropEmpty(q["key"]),
		}
		if s := q.Get("index"); s != "" {
			if idx, err := strconv.ParseUint(s, 10, 64); err == nil {
				req.StartIndex = idx
			}
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported by transport", http.StatusInternalServerError)
			return
		}

		// Hint to intermediaries (nginx, ELB) that this is a streaming
		// response so they don't buffer.
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		sub := broker.Subscribe(req)
		defer sub.Close()

		enc := json.NewEncoder(bufio.NewWriter(w))
		for {
			evt, err := sub.Next(r.Context())
			if err != nil {
				return
			}
			if err := enc.Encode(evt); err != nil {
				return
			}
			flusher.Flush()
		}
	})
}

// dropEmpty strips blank strings from a repeated query param — users
// who pass ?topic=&topic=Job should be treated as if they only sent
// ?topic=Job rather than getting an implicit "match empty topic" term.
func dropEmpty(ss []string) []string {
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}
