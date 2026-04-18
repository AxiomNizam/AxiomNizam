// Package lineage — OpenLineage-compatible event emitter.
//
// AxiomNizam's internal lineage store (tracker.go) is optimised for
// in-process graph queries.  To interoperate with external catalogues
// such as Marquez, DataHub, OpenMetadata, or Dagster, lineage events
// also need to be emitted in the OpenLineage JSON format described at
// https://openlineage.io/docs/spec/object-model .  This file provides
// that translation plus a pluggable transport.
//
// Usage
//
//	e := lineage.NewOpenLineageEmitter(lineage.HTTPTransport("https://marquez.example/api/v1/lineage"))
//	e.EmitStart(ctx, lineage.RunEvent{
//	    Job:     lineage.Job{Namespace: "axiomnizam", Name: "cdc.orders"},
//	    RunID:   "a-uuid",
//	    Inputs:  []lineage.Dataset{{Namespace: "postgres://prod", Name: "public.orders"}},
//	    Outputs: []lineage.Dataset{{Namespace: "kafka://prod", Name: "orders.cdc"}},
//	})
//	// … job runs …
//	e.EmitComplete(ctx, runEvent)
//
// The emitter is intentionally small: it does not retry failed sends
// (that is the transport's responsibility) and does not batch events
// (OpenLineage consumers handle deduplication via the (runID,eventType)
// tuple).
package lineage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// EventType enumerates the OpenLineage RunEvent types we emit.  "OTHER"
// is deliberately unsupported: catalog consumers typically ignore it.
type EventType string

const (
	// EventStart indicates a run has begun executing.
	EventStart EventType = "START"
	// EventComplete indicates a run finished successfully.
	EventComplete EventType = "COMPLETE"
	// EventFail indicates a run terminated with an error.
	EventFail EventType = "FAIL"
	// EventAbort indicates a run was cancelled before completion.
	EventAbort EventType = "ABORT"
)

// Job identifies the unit of work producing lineage.  Namespace is a
// stable identifier for the system running the job ("axiomnizam", or a
// cluster-id for multi-tenant deployments); Name is the job path
// ("cdc.orders", "etl.customer_ingest").
type Job struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// Dataset identifies a logical data asset.  Namespace uses the scheme
// URIs defined by the OpenLineage naming spec — "postgres://host",
// "kafka://cluster", "s3://bucket".
type Dataset struct {
	Namespace string                 `json:"namespace"`
	Name      string                 `json:"name"`
	Facets    map[string]interface{} `json:"facets,omitempty"`
}

// RunEvent is AxiomNizam's internal representation of an OpenLineage
// run event.  The emitter translates it to the on-wire format.
type RunEvent struct {
	Job     Job
	RunID   string
	Inputs  []Dataset
	Outputs []Dataset
	// Facets attached at the run level (e.g. nominalTime, parent).
	Facets map[string]interface{}
	// Producer overrides the emitter's default producer string.  Leave
	// empty to use the emitter default.
	Producer string
}

// Transport delivers OpenLineage payloads.  HTTPTransport is the
// default implementation; tests and in-process consumers can supply a
// custom transport that appends to a slice for assertion.
type Transport interface {
	Send(ctx context.Context, payload []byte) error
}

// TransportFunc adapts a function to the Transport interface.
type TransportFunc func(ctx context.Context, payload []byte) error

// Send implements Transport.
func (f TransportFunc) Send(ctx context.Context, payload []byte) error { return f(ctx, payload) }

// HTTPTransport POSTs payloads to the configured OpenLineage endpoint.
// It reuses a single http.Client with a 10s timeout; callers needing
// custom TLS / auth should substitute their own TransportFunc.
func HTTPTransport(endpoint string) Transport {
	client := &http.Client{Timeout: 10 * time.Second}
	return TransportFunc(func(ctx context.Context, payload []byte) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("openlineage endpoint returned %d", resp.StatusCode)
		}
		return nil
	})
}

// OpenLineageEmitter serialises RunEvents to the OpenLineage wire
// format and forwards them via the configured Transport.
type OpenLineageEmitter struct {
	transport Transport
	producer  string

	mu        sync.Mutex
	schemaURL string
}

// NewOpenLineageEmitter constructs an emitter.  The producer value is
// included in every event per the OpenLineage spec — it should be a
// URL identifying the emitting software (e.g. a GitHub repo URL).
func NewOpenLineageEmitter(transport Transport) *OpenLineageEmitter {
	return &OpenLineageEmitter{
		transport: transport,
		producer:  "https://github.com/shafiunmiraz0/AxiomNizam",
		schemaURL: "https://openlineage.io/spec/2-0-2/OpenLineage.json",
	}
}

// WithProducer overrides the default producer string.
func (e *OpenLineageEmitter) WithProducer(p string) *OpenLineageEmitter {
	e.mu.Lock()
	e.producer = p
	e.mu.Unlock()
	return e
}

// EmitStart emits a START event for the supplied run.
func (e *OpenLineageEmitter) EmitStart(ctx context.Context, ev RunEvent) error {
	return e.emit(ctx, EventStart, ev)
}

// EmitComplete emits a COMPLETE event for the supplied run.
func (e *OpenLineageEmitter) EmitComplete(ctx context.Context, ev RunEvent) error {
	return e.emit(ctx, EventComplete, ev)
}

// EmitFail emits a FAIL event carrying the supplied error as an
// errorMessage facet per the OpenLineage 2.0 spec.
func (e *OpenLineageEmitter) EmitFail(ctx context.Context, ev RunEvent, cause error) error {
	if cause != nil {
		if ev.Facets == nil {
			ev.Facets = map[string]interface{}{}
		}
		ev.Facets["errorMessage"] = map[string]interface{}{
			"_producer":           e.producer,
			"_schemaURL":          "https://openlineage.io/spec/facets/1-0-0/ErrorMessageRunFacet.json",
			"message":             cause.Error(),
			"programmingLanguage": "Go",
		}
	}
	return e.emit(ctx, EventFail, ev)
}

// EmitAbort emits an ABORT event for the supplied run.
func (e *OpenLineageEmitter) EmitAbort(ctx context.Context, ev RunEvent) error {
	return e.emit(ctx, EventAbort, ev)
}

// emit is the shared serialisation+transport path.
func (e *OpenLineageEmitter) emit(ctx context.Context, kind EventType, ev RunEvent) error {
	e.mu.Lock()
	producer := e.producer
	if ev.Producer != "" {
		producer = ev.Producer
	}
	schema := e.schemaURL
	e.mu.Unlock()

	payload := map[string]interface{}{
		"eventType": string(kind),
		"eventTime": time.Now().UTC().Format(time.RFC3339Nano),
		"producer":  producer,
		"schemaURL": schema,
		"run":       map[string]interface{}{"runId": ev.RunID, "facets": ev.Facets},
		"job":       map[string]interface{}{"namespace": ev.Job.Namespace, "name": ev.Job.Name},
		"inputs":    translateDatasets(ev.Inputs),
		"outputs":   translateDatasets(ev.Outputs),
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal OpenLineage event: %w", err)
	}
	return e.transport.Send(ctx, buf)
}

// translateDatasets converts the Go struct form to the wire form.  A
// nil slice is emitted as an empty JSON array, not null, per spec.
func translateDatasets(in []Dataset) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(in))
	for _, d := range in {
		entry := map[string]interface{}{
			"namespace": d.Namespace,
			"name":      d.Name,
		}
		if len(d.Facets) > 0 {
			entry["facets"] = d.Facets
		}
		out = append(out, entry)
	}
	return out
}
