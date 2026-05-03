package materialized

// =====================================================
// WS-5.2 — Materialized Views
//
// Pre-computes and caches cross-source query results as
// declarative materialized views. The manager handles
// scheduled refresh, staleness tracking, and incremental
// invalidation when source data changes.
// =====================================================

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MaterializedView represents a pre-computed query result.
type MaterializedView struct {
	Name          string                   `json:"name"`
	Query         string                   `json:"query"`
	RefreshSchedule string                 `json:"refreshSchedule,omitempty"` // "5m", "1h"
	Columns       []string                 `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int64                    `json:"rowCount"`
	Sources       []string                 `json:"sources"` // Datasources this view depends on
	CreatedAt     time.Time                `json:"createdAt"`
	RefreshedAt   *time.Time               `json:"refreshedAt,omitempty"`
	NextRefreshAt *time.Time               `json:"nextRefreshAt,omitempty"`
	Stale         bool                     `json:"stale"`
	RefreshCount  int64                    `json:"refreshCount"`
	AvgRefreshMs  float64                  `json:"avgRefreshMs"`
}

// ViewStats exposes manager-level metrics.
type ViewStats struct {
	TotalViews     int   `json:"totalViews"`
	StaleViews     int   `json:"staleViews"`
	TotalRefreshes int64 `json:"totalRefreshes"`
	TotalHits      int64 `json:"totalHits"`
}

// QueryRefresher abstracts executing a query to refresh a materialized view.
type QueryRefresher interface {
	Execute(ctx context.Context, sql string) ([]string, []map[string]interface{}, error)
}

// ViewManager manages the lifecycle of materialized views.
type ViewManager struct {
	mu        sync.RWMutex
	views     map[string]*MaterializedView
	refresher QueryRefresher
	hits      int64
}

// NewViewManager creates a new materialized view manager.
func NewViewManager(refresher QueryRefresher) *ViewManager {
	return &ViewManager{
		views:     make(map[string]*MaterializedView),
		refresher: refresher,
	}
}

// Create registers a new materialized view.
func (m *ViewManager) Create(name, query, refreshSchedule string, sources []string) (*MaterializedView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.views[name]; exists {
		return nil, fmt.Errorf("materialized view %q already exists", name)
	}

	view := &MaterializedView{
		Name:            name,
		Query:           query,
		RefreshSchedule: refreshSchedule,
		Sources:         sources,
		CreatedAt:       time.Now(),
		Stale:           true,
	}

	m.views[name] = view
	return view, nil
}

// Get retrieves a materialized view's cached data.
func (m *ViewManager) Get(name string) (*MaterializedView, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	view, ok := m.views[name]
	if ok {
		m.hits++
	}
	return view, ok
}

// Refresh re-executes a view's query and updates the cached data.
func (m *ViewManager) Refresh(ctx context.Context, name string) error {
	m.mu.Lock()
	view, ok := m.views[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("materialized view %q not found", name)
	}
	m.mu.Unlock()

	if m.refresher == nil {
		return fmt.Errorf("no query refresher configured")
	}

	start := time.Now()
	columns, rows, err := m.refresher.Execute(ctx, view.Query)
	if err != nil {
		return fmt.Errorf("refresh of %q failed: %w", name, err)
	}
	duration := time.Since(start)

	m.mu.Lock()
	defer m.mu.Unlock()

	view.Columns = columns
	view.Rows = rows
	view.RowCount = int64(len(rows))
	now := time.Now()
	view.RefreshedAt = &now
	view.Stale = false
	view.RefreshCount++

	// Update rolling average refresh time.
	if view.RefreshCount == 1 {
		view.AvgRefreshMs = float64(duration.Milliseconds())
	} else {
		view.AvgRefreshMs = (view.AvgRefreshMs*float64(view.RefreshCount-1) + float64(duration.Milliseconds())) / float64(view.RefreshCount)
	}

	// Schedule next refresh.
	if view.RefreshSchedule != "" {
		if d, err := time.ParseDuration(view.RefreshSchedule); err == nil {
			next := now.Add(d)
			view.NextRefreshAt = &next
		}
	}

	return nil
}

// InvalidateBySource marks all views depending on a source as stale.
func (m *ViewManager) InvalidateBySource(source string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, view := range m.views {
		for _, s := range view.Sources {
			if s == source {
				view.Stale = true
				count++
				break
			}
		}
	}
	return count
}

// RefreshStale refreshes all views that are marked as stale.
func (m *ViewManager) RefreshStale(ctx context.Context) (int, []error) {
	m.mu.RLock()
	var staleNames []string
	for name, view := range m.views {
		if view.Stale {
			staleNames = append(staleNames, name)
		}
	}
	m.mu.RUnlock()

	var errs []error
	refreshed := 0
	for _, name := range staleNames {
		if err := m.Refresh(ctx, name); err != nil {
			errs = append(errs, err)
		} else {
			refreshed++
		}
	}
	return refreshed, errs
}

// RefreshDue refreshes all views whose next refresh time has passed.
func (m *ViewManager) RefreshDue(ctx context.Context) (int, []error) {
	now := time.Now()

	m.mu.RLock()
	var dueNames []string
	for name, view := range m.views {
		if view.NextRefreshAt != nil && now.After(*view.NextRefreshAt) {
			dueNames = append(dueNames, name)
		}
	}
	m.mu.RUnlock()

	var errs []error
	refreshed := 0
	for _, name := range dueNames {
		if err := m.Refresh(ctx, name); err != nil {
			errs = append(errs, err)
		} else {
			refreshed++
		}
	}
	return refreshed, errs
}

// Delete removes a materialized view.
func (m *ViewManager) Delete(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.views[name]; ok {
		delete(m.views, name)
		return true
	}
	return false
}

// List returns all materialized views.
func (m *ViewManager) List() []*MaterializedView {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*MaterializedView
	for _, view := range m.views {
		result = append(result, view)
	}
	return result
}

// Stats returns manager-level statistics.
func (m *ViewManager) Stats() ViewStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stale := 0
	var totalRefreshes int64
	for _, view := range m.views {
		if view.Stale {
			stale++
		}
		totalRefreshes += view.RefreshCount
	}

	return ViewStats{
		TotalViews:     len(m.views),
		StaleViews:     stale,
		TotalRefreshes: totalRefreshes,
		TotalHits:      m.hits,
	}
}
