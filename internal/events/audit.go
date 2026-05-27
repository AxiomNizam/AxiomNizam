package events

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditLevel represents audit severity level
type AuditLevel string

const (
	AuditLevelNormal  AuditLevel = "Normal"
	AuditLevelWarning AuditLevel = "Warning"
	AuditLevelError   AuditLevel = "Error"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	// Core fields
	ID           string     `json:"id"`
	Timestamp    time.Time  `json:"timestamp"`
	Level        AuditLevel `json:"level"`
	EventType    string     `json:"eventType"`    // e.g., "PolicyApplied", "ValidationFailed"
	Action       string     `json:"action"`       // e.g., "apply", "delete", "update"
	Resource     string     `json:"resource"`     // e.g., "api/production"
	ResourceKind string     `json:"resourceKind"` // e.g., "API", "Policy", "Workflow"

	// User/Source information
	User      string `json:"user"`
	UserID    string `json:"userId,omitempty"`
	SourceIP  string `json:"sourceIp,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`

	// Outcome
	Status  string `json:"status"`  // "success", "failure"
	Message string `json:"message"` // Human-readable message
	Error   string `json:"error,omitempty"`

	// Additional context
	Namespace        string                 `json:"namespace,omitempty"`
	Generation       int64                  `json:"generation,omitempty"`
	OldValue         map[string]interface{} `json:"oldValue,omitempty"`
	NewValue         map[string]interface{} `json:"newValue,omitempty"`
	Reason           string                 `json:"reason,omitempty"`
	AuditAnnotations map[string]string      `json:"auditAnnotations,omitempty"`
}

// AuditLogger handles audit logging
type AuditLogger struct {
	mu         sync.RWMutex
	handlers   []AuditHandler
	eventChan  chan *AuditEvent
	stopChan   chan struct{}
	wg         sync.WaitGroup
	fileHandle *os.File
	enabled    bool
}

// AuditHandler processes audit events
type AuditHandler func(ctx context.Context, event *AuditEvent) error

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	al := &AuditLogger{
		handlers:  make([]AuditHandler, 0),
		eventChan: make(chan *AuditEvent, 1000),
		stopChan:  make(chan struct{}),
		enabled:   true,
	}

	// Start event processor
	al.wg.Add(1)
	go al.processEvents()

	return al
}

// AddHandler adds an audit handler
func (al *AuditLogger) AddHandler(handler AuditHandler) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.handlers = append(al.handlers, handler)
}

// LogApply logs a resource apply action
func (al *AuditLogger) LogApply(ctx context.Context, user string, kind, name, namespace string, status string, err error) {
	if !al.enabled {
		return
	}

	event := &AuditEvent{
		ID:           generateAuditID(),
		Timestamp:    time.Now(),
		EventType:    "ResourceApplied",
		Action:       "apply",
		Resource:     fmt.Sprintf("%s/%s", kind, name),
		ResourceKind: kind,
		User:         user,
		Namespace:    namespace,
		Status:       status,
	}

	if err != nil {
		event.Level = AuditLevelError
		event.Error = err.Error()
	} else {
		event.Level = AuditLevelNormal
	}

	// Try to extract source info from context
	if userID, ok := ctx.Value("user_id").(string); ok {
		event.UserID = userID
	}
	if sourceIP, ok := ctx.Value("source_ip").(string); ok {
		event.SourceIP = sourceIP
	}

	al.eventChan <- event
}

// LogDelete logs a resource deletion
func (al *AuditLogger) LogDelete(ctx context.Context, user string, kind, name, namespace string, reason string) {
	if !al.enabled {
		return
	}

	event := &AuditEvent{
		ID:           generateAuditID(),
		Timestamp:    time.Now(),
		Level:        AuditLevelNormal,
		EventType:    "ResourceDeleted",
		Action:       "delete",
		Resource:     fmt.Sprintf("%s/%s", kind, name),
		ResourceKind: kind,
		User:         user,
		Namespace:    namespace,
		Status:       "success",
		Reason:       reason,
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		event.UserID = userID
	}
	if sourceIP, ok := ctx.Value("source_ip").(string); ok {
		event.SourceIP = sourceIP
	}

	al.eventChan <- event
}

// LogPolicyEvaluation logs policy evaluation
func (al *AuditLogger) LogPolicyEvaluation(ctx context.Context, user string, policyName string, result bool, reason string) {
	if !al.enabled {
		return
	}

	level := AuditLevelNormal
	status := "allowed"
	if !result {
		level = AuditLevelWarning
		status = "denied"
	}

	event := &AuditEvent{
		ID:           generateAuditID(),
		Timestamp:    time.Now(),
		Level:        level,
		EventType:    "PolicyEvaluated",
		Action:       "evaluate",
		Resource:     policyName,
		ResourceKind: "Policy",
		User:         user,
		Status:       status,
		Reason:       reason,
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		event.UserID = userID
	}

	al.eventChan <- event
}

// LogReconciliation logs controller reconciliation
func (al *AuditLogger) LogReconciliation(ctx context.Context, kind, name, namespace string, status string, err error) {
	if !al.enabled {
		return
	}

	level := AuditLevelNormal
	eventStatus := "success"
	if err != nil {
		level = AuditLevelError
		eventStatus = "failure"
	}

	event := &AuditEvent{
		ID:           generateAuditID(),
		Timestamp:    time.Now(),
		Level:        level,
		EventType:    "ReconciliationCompleted",
		Action:       "reconcile",
		Resource:     fmt.Sprintf("%s/%s", kind, name),
		ResourceKind: kind,
		Namespace:    namespace,
		Status:       eventStatus,
		Message:      status,
	}

	if err != nil {
		event.Error = err.Error()
	}

	al.eventChan <- event
}

// LogAuthenticationFailure logs authentication failures
func (al *AuditLogger) LogAuthenticationFailure(ctx context.Context, user string, sourceIP string, reason string) {
	if !al.enabled {
		return
	}

	event := &AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now(),
		Level:     AuditLevelWarning,
		EventType: "AuthenticationFailed",
		Action:    "authenticate",
		User:      user,
		SourceIP:  sourceIP,
		Status:    "failure",
		Reason:    reason,
	}

	al.eventChan <- event
}

// LogAuthorizationFailure logs authorization failures
func (al *AuditLogger) LogAuthorizationFailure(ctx context.Context, user string, action string, resource string, reason string) {
	if !al.enabled {
		return
	}

	event := &AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now(),
		Level:     AuditLevelWarning,
		EventType: "AuthorizationFailed",
		Action:    action,
		Resource:  resource,
		User:      user,
		Status:    "denied",
		Reason:    reason,
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		event.UserID = userID
	}

	al.eventChan <- event
}

// GetAuditLog retrieves audit log entries
func (al *AuditLogger) GetAuditLog(ctx context.Context, filters map[string]interface{}, limit int) ([]*AuditEvent, error) {
	// This would query a persistent store in a real implementation
	// For now, return empty slice - persistence handled by handlers
	return []*AuditEvent{}, nil
}

// SetFileOutput sets file output for audit logs
func (al *AuditLogger) SetFileOutput(filepath string) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit file: %w", err)
	}

	al.fileHandle = f

	// Add file handler
	al.handlers = append(al.handlers, func(ctx context.Context, event *AuditEvent) error {
		// JSON encode and write to file
		data := fmt.Sprintf("%s [%s] %s: %s (user=%s, status=%s)\n",
			event.Timestamp.Format(time.RFC3339),
			event.Level,
			event.EventType,
			event.Resource,
			event.User,
			event.Status,
		)
		_, err := al.fileHandle.WriteString(data)
		return err
	})

	return nil
}

// Disable disables audit logging
func (al *AuditLogger) Disable() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.enabled = false
}

// Enable enables audit logging
func (al *AuditLogger) Enable() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.enabled = true
}

// Close closes the audit logger
func (al *AuditLogger) Close() error {
	close(al.stopChan)
	al.wg.Wait()

	if al.fileHandle != nil {
		return al.fileHandle.Close()
	}
	return nil
}

// processEvents processes audit events in background
func (al *AuditLogger) processEvents() {
	defer al.wg.Done()

	for {
		select {
		case event := <-al.eventChan:
			if event == nil {
				continue
			}

			al.mu.RLock()
			handlers := al.handlers
			al.mu.RUnlock()

			// Call all handlers
			ctx := context.Background()
			for _, handler := range handlers {
				if err := handler(ctx, event); err != nil {
					// Log handler errors silently to avoid recursion
					fmt.Fprintf(os.Stderr, "audit handler error: %v\n", err)
				}
			}

		case <-al.stopChan:
			return
		}
	}
}

// generateAuditID generates a unique audit event ID
func generateAuditID() string {
	return fmt.Sprintf("audit-%d", time.Now().UnixNano())
}

// LogApply creates a temporary audit logger and logs an apply event.
// For persistent logging, create an AuditLogger instance and use it directly.
func LogApply(ctx context.Context, user string, kind, name, namespace string, status string, err error) {
	logger := NewAuditLogger()
	logger.LogApply(ctx, user, kind, name, namespace, status, err)
}

// LogDelete creates a temporary audit logger and logs a delete event.
func LogDelete(ctx context.Context, user string, kind, name, namespace string, reason string) {
	logger := NewAuditLogger()
	logger.LogDelete(ctx, user, kind, name, namespace, reason)
}

// LogAuthenticationFailure creates a temporary audit logger and logs an auth failure.
func LogAuthenticationFailure(ctx context.Context, user string, sourceIP string, reason string) {
	logger := NewAuditLogger()
	logger.LogAuthenticationFailure(ctx, user, sourceIP, reason)
}
