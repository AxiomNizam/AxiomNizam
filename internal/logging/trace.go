package logging

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const traceIDKey contextKey = "trace_id"
const requestIDKey contextKey = "request_id"

// WithTraceID returns a new context containing the given trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from the context.
// Returns empty string if not present.
func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// WithRequestID returns a new context containing the given request ID.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from the context.
// Returns empty string if not present.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// EnsureTraceID returns the trace ID from context, generating one if absent.
func EnsureTraceID(ctx context.Context) string {
	if tid := TraceIDFromContext(ctx); tid != "" {
		return tid
	}
	return uuid.New().String()
}

// TraceLogger returns a zap.Logger enriched with trace_id and request_id from context.
func TraceLogger(ctx context.Context) *zap.Logger {
	logger := Z()
	if tid := TraceIDFromContext(ctx); tid != "" {
		logger = logger.With(zap.String("trace_id", tid))
	}
	if rid := RequestIDFromContext(ctx); rid != "" {
		logger = logger.With(zap.String("request_id", rid))
	}
	return logger
}
