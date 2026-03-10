package export

import (
	"fmt"
	"sync"
	"time"
)

const errExportNotFound = "export not found"

// InMemoryExportManager in-memory export implementation
type InMemoryExportManager struct {
	mu        sync.RWMutex
	exports   map[string]*ExportJob
	history   map[string][]*ExportHistory
	templates map[string]*ExportTemplate
}

// NewInMemoryExportManager creates manager
func NewInMemoryExportManager() *InMemoryExportManager {
	return &InMemoryExportManager{
		exports:   make(map[string]*ExportJob),
		history:   make(map[string][]*ExportHistory),
		templates: make(map[string]*ExportTemplate),
	}
}

// SubmitExport submits export job
func (m *InMemoryExportManager) SubmitExport(export *ExportJob) (*ExportJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if export.ID == "" {
		export.ID = fmt.Sprintf("export-%d", time.Now().UnixNano())
	}
	if export.CreatedAt.IsZero() {
		export.CreatedAt = time.Now()
	}

	export.Status = ExportPending
	export.Progress = 0
	m.exports[export.ID] = export
	m.history[export.ID] = []*ExportHistory{}

	return export, nil
}

// GetExport retrieves export job
func (m *InMemoryExportManager) GetExport(id string) (*ExportJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	export, exists := m.exports[id]
	if !exists {
		return nil, fmt.Errorf(errExportNotFound)
	}
	return export, nil
}

// ListExports lists export jobs
func (m *InMemoryExportManager) ListExports(tenantID string) ([]*ExportJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*ExportJob
	for _, e := range m.exports {
		if tenantID != "" && e.TenantID != tenantID {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

// UpdateProgress updates export progress
func (m *InMemoryExportManager) UpdateProgress(id string, progress float64, status ExportStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	export, exists := m.exports[id]
	if !exists {
		return fmt.Errorf(errExportNotFound)
	}

	export.Progress = progress
	export.Status = status

	return nil
}

// CancelExport cancels export
func (m *InMemoryExportManager) CancelExport(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	export, exists := m.exports[id]
	if !exists {
		return fmt.Errorf(errExportNotFound)
	}

	export.Status = ExportCancelled

	return nil
}

// GetHistory retrieves export history
func (m *InMemoryExportManager) GetHistory(exportID string) ([]*ExportHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hist, exists := m.history[exportID]
	if !exists {
		return nil, fmt.Errorf("history not found")
	}
	return hist, nil
}

// GetDownloadURL gets download URL
func (m *InMemoryExportManager) GetDownloadURL(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	export, exists := m.exports[id]
	if !exists {
		return "", fmt.Errorf(errExportNotFound)
	}

	if export.Status != ExportCompleted {
		return "", fmt.Errorf("export not ready")
	}

	return fmt.Sprintf("/downloads/%s.%s", export.ID, export.Format), nil
}

// CreateTemplate creates export template
func (m *InMemoryExportManager) CreateTemplate(template *ExportTemplate) (*ExportTemplate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if template.ID == "" {
		template.ID = fmt.Sprintf("template-%d", time.Now().UnixNano())
	}

	m.templates[template.ID] = template
	return template, nil
}

// ListTemplates lists templates
func (m *InMemoryExportManager) ListTemplates(tenantID string) ([]*ExportTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*ExportTemplate
	for _, t := range m.templates {
		if tenantID != "" && t.TenantID != tenantID {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

// GetTemplate retrieves template
func (m *InMemoryExportManager) GetTemplate(id string) (*ExportTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	template, exists := m.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found")
	}
	return template, nil
}

// DeleteTemplate deletes template
func (m *InMemoryExportManager) DeleteTemplate(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.templates, id)
	return nil
}
