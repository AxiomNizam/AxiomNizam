package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with request context support
type Logger struct {
	*zap.Logger
}

// RequestIDKey is the key for request IDs in context
type RequestIDKey struct{}

// CorrelationIDKey is the key for correlation IDs in context
type CorrelationIDKey struct{}

// New creates a new logger with specified environment
func New(env string) (*Logger, error) {
	var config zap.Config

	switch env {
	case "production", "prod":
		config = zap.NewProductionConfig()
	case "development", "dev":
		config = zap.NewDevelopmentConfig()
	default:
		config = zap.NewDevelopmentConfig()
	}

	// Set encoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{Logger: zapLogger}, nil
}

// NewDevelopment creates a development logger
func NewDevelopment() (*Logger, error) {
	return New("development")
}

// NewProduction creates a production logger
func NewProduction() (*Logger, error) {
	return New("production")
}

// WithRequestID returns a logger with request ID
func (l *Logger) WithRequestID(ctx context.Context, requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("request_id", requestID)),
	}
}

// WithCorrelationID returns a logger with correlation ID
func (l *Logger) WithCorrelationID(ctx context.Context, correlationID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("correlation_id", correlationID)),
	}
}

// WithContext returns a logger with context metadata
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	// Extract request ID if present
	if requestID, ok := ctx.Value(RequestIDKey{}).(string); ok {
		logger = logger.With(zap.String("request_id", requestID))
	}

	// Extract correlation ID if present
	if correlationID, ok := ctx.Value(CorrelationIDKey{}).(string); ok {
		logger = logger.With(zap.String("correlation_id", correlationID))
	}

	return &Logger{Logger: logger}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
	}
}

// FromContext extracts logger from context or returns a default logger
func FromContext(ctx context.Context, defaultLogger *Logger) *Logger {
	if logger, ok := ctx.Value("logger").(*Logger); ok {
		return logger
	}
	return defaultLogger
}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}

// ContextWithRequestID adds request ID to context
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey{}, requestID)
}

// ContextWithCorrelationID adds correlation ID to context
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey{}, correlationID)
}

// ContextWithUserID adds user ID to context for logging
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}

// ContextWithTraceID adds trace ID to context
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

// LogrusToZapField converts common fields to zap fields
func LogrusToZapField(key string, value interface{}) zap.Field {
	switch v := value.(type) {
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case float64:
		return zap.Float64(key, v)
	case bool:
		return zap.Bool(key, v)
	case error:
		return zap.Error(v)
	default:
		return zap.Any(key, v)
	}
}

// SetLogLevel updates the log level for logger
func SetLogLevel(level string) zap.AtomicLevel {
	atomicLevel := zap.NewAtomicLevel()

	switch level {
	case "debug":
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case "info":
		atomicLevel.SetLevel(zapcore.InfoLevel)
	case "warn", "warning":
		atomicLevel.SetLevel(zapcore.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zapcore.ErrorLevel)
	case "fatal":
		atomicLevel.SetLevel(zapcore.FatalLevel)
	default:
		atomicLevel.SetLevel(zapcore.InfoLevel)
	}

	return atomicLevel
}

// StructuredFields represents common structured logging fields
type StructuredFields struct {
	RequestID     string
	CorrelationID string
	UserID        string
	TraceID       string
	SpanID        string
	Action        string
	Resource      string
	Status        string
	Duration      int64 // milliseconds
	Error         error
	StatusCode    int
	Message       string
}

// ToZapFields converts StructuredFields to zap.Field slice
func (sf StructuredFields) ToZapFields() []zap.Field {
	var fields []zap.Field

	if sf.RequestID != "" {
		fields = append(fields, zap.String("request_id", sf.RequestID))
	}
	if sf.CorrelationID != "" {
		fields = append(fields, zap.String("correlation_id", sf.CorrelationID))
	}
	if sf.UserID != "" {
		fields = append(fields, zap.String("user_id", sf.UserID))
	}
	if sf.TraceID != "" {
		fields = append(fields, zap.String("trace_id", sf.TraceID))
	}
	if sf.SpanID != "" {
		fields = append(fields, zap.String("span_id", sf.SpanID))
	}
	if sf.Action != "" {
		fields = append(fields, zap.String("action", sf.Action))
	}
	if sf.Resource != "" {
		fields = append(fields, zap.String("resource", sf.Resource))
	}
	if sf.Status != "" {
		fields = append(fields, zap.String("status", sf.Status))
	}
	if sf.Duration > 0 {
		fields = append(fields, zap.Int64("duration_ms", sf.Duration))
	}
	if sf.Error != nil {
		fields = append(fields, zap.Error(sf.Error))
	}
	if sf.StatusCode > 0 {
		fields = append(fields, zap.Int("status_code", sf.StatusCode))
	}
	if sf.Message != "" {
		fields = append(fields, zap.String("message", sf.Message))
	}

	return fields
}

// LogAuditEvent logs an audit event
func LogAuditEvent(logger *Logger, action, resource, userID, status string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("user_id", userID),
		zap.String("status", status),
	}

	if len(details) > 0 {
		fields = append(fields, zap.Any("details", details))
	}

	logger.Info("audit_event", fields...)
}

// LogSecurityEvent logs a security event
func LogSecurityEvent(logger *Logger, eventType, severity, userID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("severity", severity),
		zap.String("user_id", userID),
	}

	if len(details) > 0 {
		fields = append(fields, zap.Any("details", details))
	}

	if severity == "critical" || severity == "high" {
		logger.Error("security_event", fields...)
	} else {
		logger.Warn("security_event", fields...)
	}
}

// LogPerformanceMetric logs performance metrics
func LogPerformanceMetric(logger *Logger, operation string, durationMs int64, success bool, metadata map[string]interface{}) {
	level := "info"
	if !success {
		level = "warn"
	}

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Int64("duration_ms", durationMs),
		zap.Bool("success", success),
	}

	if len(metadata) > 0 {
		fields = append(fields, zap.Any("metadata", metadata))
	}

	if level == "warn" {
		logger.Warn("performance_metric", fields...)
	} else {
		logger.Info("performance_metric", fields...)
	}
}

// AtomicLevel wrapper for log level changes
type LogLevel struct {
	level zap.AtomicLevel
}

// NewLogLevel creates a new log level manager
func NewLogLevel(initialLevel string) LogLevel {
	return LogLevel{
		level: SetLogLevel(initialLevel),
	}
}

// Set changes the log level
func (ll LogLevel) Set(level string) error {
	l := SetLogLevel(level)
	ll.level.SetLevel(l.Level())
	return nil
}

// Get returns current log level
func (ll LogLevel) Get() string {
	return ll.level.Level().String()
}
