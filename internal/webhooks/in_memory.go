package webhooks

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryWebhookManager in-memory webhook implementation
type InMemoryWebhookManager struct {
	mu    sync.RWMutex
	hooks  map[string]*Webhook
	logs   map[string][]*WebhookDeliveryLog
}

// NewInMemoryWebhookManager creates manager
func NewInMemoryWebhookManager() *InMemoryWebhookManager {
	return &InMemoryWebhookManager{
		hooks: make(map[string]*Webhook),
		logs:  make(map[string][]*WebhookDeliveryLog),
	}
}

// CreateWebhook creates webhook
func (m *InMemoryWebhookManager) CreateWebhook(webhook *Webhook) (*Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if webhook.ID == "" {
		webhook.ID = fmt.Sprintf("webhook-%d", time.Now().UnixNano())
	}

	m.hooks[webhook.ID] = webhook
	m.logs[webhook.ID] = []*WebhookDeliveryLog{}
	return webhook, nil
}

// GetWebhook retrieves webhook
func (m *InMemoryWebhookManager) GetWebhook(id string) (*Webhook, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	webhook, exists := m.hooks[id]
	if !exists {
		return nil, fmt.Errorf("webhook not found")
	}
	return webhook, nil
}

// ListWebhooks lists webhooks
func (m *InMemoryWebhookManager) ListWebhooks(tenantID, eventType string) ([]*Webhook, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Webhook
	for _, w := range m.hooks {
		if tenantID != "" && w.TenantID != tenantID {
			continue
		}
		result = append(result, w)
	}
	return result, nil
}

// UpdateWebhook updates webhook
func (m *InMemoryWebhookManager) UpdateWebhook(webhook *Webhook) (*Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.hooks[webhook.ID]; !exists {
		return nil, fmt.Errorf("webhook not found")
	}

	webhook.UpdatedAt = time.Now()
	m.hooks[webhook.ID] = webhook
	return webhook, nil
}

// DeleteWebhook deletes webhook
func (m *InMemoryWebhookManager) DeleteWebhook(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.hooks, id)
	return nil
}

// TestWebhook tests webhook
func (m *InMemoryWebhookManager) TestWebhook(webhook *Webhook) (interface{}, error) {
	return map[string]interface{}{"status": "ok", "latency": 100}, nil
}

// GetDeliveryLogs retrieves delivery logs
func (m *InMemoryWebhookManager) GetDeliveryLogs(webhookID string) ([]*WebhookDeliveryLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs, exists := m.logs[webhookID]
	if !exists {
		return nil, fmt.Errorf("logs not found")
	}
	return logs, nil
}
