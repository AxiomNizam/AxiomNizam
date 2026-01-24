package ctxutils

import (
	"context"
	"time"
)

// ContextKey defines a type for context keys
type ContextKey string

const (
	// User context keys
	UserIDKey       ContextKey = "user_id"
	UsernameKey     ContextKey = "username"
	UserRoleKey     ContextKey = "user_role"
	UserEmailKey    ContextKey = "user_email"
	UserPermissionsKey ContextKey = "user_permissions"

	// Request context keys
	RequestIDKey      ContextKey = "request_id"
	CorrelationIDKey  ContextKey = "correlation_id"
	TraceIDKey        ContextKey = "trace_id"
	SpanIDKey         ContextKey = "span_id"
	RequestPathKey    ContextKey = "request_path"
	RequestMethodKey  ContextKey = "request_method"

	// Metadata keys
	StartTimeKey      ContextKey = "start_time"
	DeadlineKey       ContextKey = "deadline"
	TimeoutKey        ContextKey = "timeout"
	ClientIPKey       ContextKey = "client_ip"
	UserAgentKey      ContextKey = "user_agent"

	// Authorization keys
	AuthTokenKey      ContextKey = "auth_token"
	AuthTypeKey       ContextKey = "auth_type"
	AuthExpiresKey    ContextKey = "auth_expires"

	// Application keys
	AppVersionKey     ContextKey = "app_version"
	EnvironmentKey    ContextKey = "environment"
	TenantIDKey       ContextKey = "tenant_id"
)

// RequestMetadata holds request-level metadata
type RequestMetadata struct {
	RequestID      string
	CorrelationID  string
	TraceID        string
	SpanID         string
	Method         string
	Path           string
	ClientIP       string
	UserAgent      string
	StartTime      time.Time
	UserID         string
	Username       string
	UserRole       string
	TenantID       string
	AuthToken      string
	AuthType       string
}

// NewContext creates a new context with metadata
func NewContext(parent context.Context, metadata RequestMetadata) context.Context {
	ctx := parent
	
	// Add request IDs
	if metadata.RequestID != "" {
		ctx = context.WithValue(ctx, RequestIDKey, metadata.RequestID)
	}
	if metadata.CorrelationID != "" {
		ctx = context.WithValue(ctx, CorrelationIDKey, metadata.CorrelationID)
	}
	if metadata.TraceID != "" {
		ctx = context.WithValue(ctx, TraceIDKey, metadata.TraceID)
	}
	if metadata.SpanID != "" {
		ctx = context.WithValue(ctx, SpanIDKey, metadata.SpanID)
	}

	// Add request metadata
	if metadata.Method != "" {
		ctx = context.WithValue(ctx, RequestMethodKey, metadata.Method)
	}
	if metadata.Path != "" {
		ctx = context.WithValue(ctx, RequestPathKey, metadata.Path)
	}
	if metadata.ClientIP != "" {
		ctx = context.WithValue(ctx, ClientIPKey, metadata.ClientIP)
	}
	if metadata.UserAgent != "" {
		ctx = context.WithValue(ctx, UserAgentKey, metadata.UserAgent)
	}

	// Add user information
	if metadata.UserID != "" {
		ctx = context.WithValue(ctx, UserIDKey, metadata.UserID)
	}
	if metadata.Username != "" {
		ctx = context.WithValue(ctx, UsernameKey, metadata.Username)
	}
	if metadata.UserRole != "" {
		ctx = context.WithValue(ctx, UserRoleKey, metadata.UserRole)
	}
	if metadata.TenantID != "" {
		ctx = context.WithValue(ctx, TenantIDKey, metadata.TenantID)
	}

	// Add auth information
	if metadata.AuthToken != "" {
		ctx = context.WithValue(ctx, AuthTokenKey, metadata.AuthToken)
	}
	if metadata.AuthType != "" {
		ctx = context.WithValue(ctx, AuthTypeKey, metadata.AuthType)
	}

	// Add timing
	if !metadata.StartTime.IsZero() {
		ctx = context.WithValue(ctx, StartTimeKey, metadata.StartTime)
	}

	return ctx
}

// GetString extracts a string value from context
func GetString(ctx context.Context, key ContextKey) (string, bool) {
	val, ok := ctx.Value(key).(string)
	return val, ok
}

// GetStringOrDefault extracts a string value or returns default
func GetStringOrDefault(ctx context.Context, key ContextKey, defaultVal string) string {
	if val, ok := GetString(ctx, key); ok {
		return val
	}
	return defaultVal
}

// GetInt extracts an int value from context
func GetInt(ctx context.Context, key ContextKey) (int, bool) {
	val, ok := ctx.Value(key).(int)
	return val, ok
}

// GetInt64 extracts an int64 value from context
func GetInt64(ctx context.Context, key ContextKey) (int64, bool) {
	val, ok := ctx.Value(key).(int64)
	return val, ok
}

// GetBool extracts a bool value from context
func GetBool(ctx context.Context, key ContextKey) (bool, bool) {
	val, ok := ctx.Value(key).(bool)
	return val, ok
}

// GetTime extracts a time.Time value from context
func GetTime(ctx context.Context, key ContextKey) (time.Time, bool) {
	val, ok := ctx.Value(key).(time.Time)
	return val, ok
}

// GetAny extracts an interface{} value from context
func GetAny(ctx context.Context, key ContextKey) (interface{}, bool) {
	val := ctx.Value(key)
	return val, val != nil
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithUsername adds username to context
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, UsernameKey, username)
}

// WithUserRole adds user role to context
func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRoleKey, role)
}

// WithUserEmail adds user email to context
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, UserEmailKey, email)
}

// WithPermissions adds user permissions to context
func WithPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, UserPermissionsKey, permissions)
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithSpanID adds span ID to context
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// WithClientIP adds client IP to context
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ClientIPKey, ip)
}

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// WithAuthToken adds auth token to context
func WithAuthToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, AuthTokenKey, token)
}

// WithAuthType adds auth type to context
func WithAuthType(ctx context.Context, authType string) context.Context {
	return context.WithValue(ctx, AuthTypeKey, authType)
}

// WithStartTime adds start time to context
func WithStartTime(ctx context.Context, startTime time.Time) context.Context {
	return context.WithValue(ctx, StartTimeKey, startTime)
}

// WithTimeout adds timeout to context
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// WithDeadline adds deadline to context
func WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, deadline)
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	return GetString(ctx, UserIDKey)
}

// GetUsername extracts username from context
func GetUsername(ctx context.Context) (string, bool) {
	return GetString(ctx, UsernameKey)
}

// GetUserRole extracts user role from context
func GetUserRole(ctx context.Context) (string, bool) {
	return GetString(ctx, UserRoleKey)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) (string, bool) {
	return GetString(ctx, RequestIDKey)
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) (string, bool) {
	return GetString(ctx, CorrelationIDKey)
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) (string, bool) {
	return GetString(ctx, TraceIDKey)
}

// GetStartTime extracts start time from context
func GetStartTime(ctx context.Context) (time.Time, bool) {
	return GetTime(ctx, StartTimeKey)
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) (string, bool) {
	return GetString(ctx, TenantIDKey)
}

// MustGetUserID extracts user ID or panics
func MustGetUserID(ctx context.Context) string {
	userID, ok := GetUserID(ctx)
	if !ok {
		panic("user_id not found in context")
	}
	return userID
}

// MustGetRequestID extracts request ID or panics
func MustGetRequestID(ctx context.Context) string {
	requestID, ok := GetRequestID(ctx)
	if !ok {
		panic("request_id not found in context")
	}
	return requestID
}

// GetMetadata extracts all available metadata from context
func GetMetadata(ctx context.Context) RequestMetadata {
	metadata := RequestMetadata{
		StartTime: time.Now(),
	}

	if val, ok := GetString(ctx, RequestIDKey); ok {
		metadata.RequestID = val
	}
	if val, ok := GetString(ctx, CorrelationIDKey); ok {
		metadata.CorrelationID = val
	}
	if val, ok := GetString(ctx, TraceIDKey); ok {
		metadata.TraceID = val
	}
	if val, ok := GetString(ctx, SpanIDKey); ok {
		metadata.SpanID = val
	}
	if val, ok := GetString(ctx, RequestMethodKey); ok {
		metadata.Method = val
	}
	if val, ok := GetString(ctx, RequestPathKey); ok {
		metadata.Path = val
	}
	if val, ok := GetString(ctx, ClientIPKey); ok {
		metadata.ClientIP = val
	}
	if val, ok := GetString(ctx, UserAgentKey); ok {
		metadata.UserAgent = val
	}
	if val, ok := GetString(ctx, UserIDKey); ok {
		metadata.UserID = val
	}
	if val, ok := GetString(ctx, UsernameKey); ok {
		metadata.Username = val
	}
	if val, ok := GetString(ctx, UserRoleKey); ok {
		metadata.UserRole = val
	}
	if val, ok := GetString(ctx, TenantIDKey); ok {
		metadata.TenantID = val
	}
	if val, ok := GetString(ctx, AuthTokenKey); ok {
		metadata.AuthToken = val
	}
	if val, ok := GetString(ctx, AuthTypeKey); ok {
		metadata.AuthType = val
	}
	if val, ok := GetTime(ctx, StartTimeKey); ok {
		metadata.StartTime = val
	}

	return metadata
}
