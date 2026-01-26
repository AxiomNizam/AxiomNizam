package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// IndexKey represents a field to index
type IndexKey struct {
	Name      string
	Extractor func(interface{}) string // Extract value from resource
}

// Index provides fast lookups by indexed fields
type Index struct {
	mu         sync.RWMutex
	name       string
	indexedBy  string
	values     map[string][]string // indexed value -> resource IDs
	resources  map[string]interface{}
}

// NewIndex creates a new index
func NewIndex(name, indexedBy string, extractor func(interface{}) string) *Index {
	return &Index{
		name:      name,
		indexedBy: indexedBy,
		values:    make(map[string][]string),
		resources: make(map[string]interface{}),
	}
}

// Add adds a resource to the index
func (idx *Index) Add(resourceID string, resource interface{}) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.resources[resourceID] = resource
	// Would extract and index value here
}

// Remove removes a resource from the index
func (idx *Index) Remove(resourceID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.resources, resourceID)
}

// Get returns resources by indexed value
func (idx *Index) Get(value string) []interface{} {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	resourceIDs := idx.values[value]
	results := make([]interface{}, len(resourceIDs))

	for i, id := range resourceIDs {
		results[i] = idx.resources[id]
	}

	return results
}

// ProjectionField represents a field to project
type ProjectionField struct {
	Name      string
	Extractor func(interface{}) interface{}
}

// Projection projects resources to specific fields (like SELECT in SQL)
type Projection struct {
	mu     sync.RWMutex
	fields []ProjectionField
	cache  map[string]map[string]interface{} // resourceID -> projected fields
}

// NewProjection creates a new projection
func NewProjection(fields ...ProjectionField) *Projection {
	return &Projection{
		fields: fields,
		cache:  make(map[string]map[string]interface{}),
	}
}

// Project projects a resource to specified fields
func (p *Projection) Project(resourceID string, resource interface{}) map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	projected := make(map[string]interface{})

	for _, field := range p.fields {
		if field.Extractor != nil {
			projected[field.Name] = field.Extractor(resource)
		}
	}

	p.cache[resourceID] = projected
	return projected
}

// Get returns cached projection
func (p *Projection) Get(resourceID string) map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.cache[resourceID]
}

// ResourceLabelSelector matches resources by labels
type LabelSelector struct {
	MatchLabels map[string]string
	MatchExpressions []LabelSelectorRequirement
}

// LabelSelectorRequirement represents a label selector requirement
type LabelSelectorRequirement struct {
	Key      string
	Operator string // In, NotIn, Exists, DoesNotExist
	Values   []string
}

// LabelMatcher matches labels against selectors
type LabelMatcher struct {
	mu       sync.RWMutex
	resources map[string]map[string]string // resourceID -> labels
}

// NewLabelMatcher creates a new label matcher
func NewLabelMatcher() *LabelMatcher {
	return &LabelMatcher{
		resources: make(map[string]map[string]string),
	}
}

// SetLabels sets labels for a resource
func (lm *LabelMatcher) SetLabels(resourceID string, labels map[string]string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.resources[resourceID] = labels
}

// SelectByLabels returns resources matching label selector
func (lm *LabelMatcher) SelectByLabels(selector *LabelSelector) []string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	results := make([]string, 0)

	for resourceID, labels := range lm.resources {
		if matchesSelector(selector, labels) {
			results = append(results, resourceID)
		}
	}

	return results
}

// FieldSelector filters resources by field values
type FieldSelector struct {
	Requirements []FieldSelectorRequirement
}

// FieldSelectorRequirement represents a field filter
type FieldSelectorRequirement struct {
	Field    string
	Operator string // =, !=, ==, in, notin
	Value    string
}

// FieldMatcher matches fields against selectors
type FieldMatcher struct {
	mu        sync.RWMutex
	resources map[string]map[string]interface{} // resourceID -> fields
}

// NewFieldMatcher creates a new field matcher
func NewFieldMatcher() *FieldMatcher {
	return &FieldMatcher{
		resources: make(map[string]map[string]interface{}),
	}
}

// SetFields sets fields for a resource
func (fm *FieldMatcher) SetFields(resourceID string, fields map[string]interface{}) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.resources[resourceID] = fields
}

// SelectByFields returns resources matching field selector
func (fm *FieldMatcher) SelectByFields(selector *FieldSelector) []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	results := make([]string, 0)

	for resourceID, fields := range fm.resources {
		if matchesFieldSelector(selector, fields) {
			results = append(results, resourceID)
		}
	}

	return results
}

// Aggregation performs aggregation on resources
type Aggregation struct {
	Type     string // count, sum, avg, min, max, group
	Field    string // field to aggregate
	GroupBy  []string // fields to group by
}

// Aggregator performs aggregations
type Aggregator struct {
	mu        sync.RWMutex
	resources map[string]map[string]interface{}
}

// NewAggregator creates a new aggregator
func NewAggregator() *Aggregator {
	return &Aggregator{
		resources: make(map[string]map[string]interface{}),
	}
}

// Add adds a resource for aggregation
func (agg *Aggregator) Add(resourceID string, fields map[string]interface{}) {
	agg.mu.Lock()
	defer agg.mu.Unlock()

	agg.resources[resourceID] = fields
}

// Aggregate performs aggregation
func (agg *Aggregator) Aggregate(aggregation *Aggregation) interface{} {
	agg.mu.RLock()
	defer agg.mu.RUnlock()

	switch aggregation.Type {
	case "count":
		return len(agg.resources)

	case "sum":
		sum := 0.0
		for _, fields := range agg.resources {
			if val, ok := fields[aggregation.Field].(float64); ok {
				sum += val
			}
		}
		return sum

	case "group":
		groups := make(map[string][]string)
		for resourceID, fields := range agg.resources {
			key := agg.getGroupKey(fields, aggregation.GroupBy)
			groups[key] = append(groups[key], resourceID)
		}
		return groups
	}

	return nil
}

// getGroupKey creates a group key from fields
func (agg *Aggregator) getGroupKey(fields map[string]interface{}, groupBy []string) string {
	parts := make([]string, len(groupBy))
	for i, field := range groupBy {
		if val, ok := fields[field]; ok {
			parts[i] = fmt.Sprintf("%v", val)
		}
	}
	return strings.Join(parts, ":")
}

// Watch watches for resource changes
type Watch struct {
	ResourceType string
	Selector     *LabelSelector
	FieldSelector *FieldSelector
	Channel      chan WatchEvent
}

// WatchEvent represents a change event
type WatchEvent struct {
	Type      string // ADDED, MODIFIED, DELETED
	Object    interface{}
	Timestamp time.Time
}

// WatchManager manages resource watchers
type WatchManager struct {
	mu       sync.RWMutex
	watches  map[string]*Watch // watchID -> watch
	nextID   int
}

// NewWatchManager creates a new watch manager
func NewWatchManager() *WatchManager {
	return &WatchManager{
		watches: make(map[string]*Watch),
	}
}

// Watch creates a new watch
func (wm *WatchManager) Watch(resourceType string, selector *LabelSelector, fieldSelector *FieldSelector) *Watch {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.nextID++
	watchID := fmt.Sprintf("watch-%d", wm.nextID)

	watch := &Watch{
		ResourceType:  resourceType,
		Selector:      selector,
		FieldSelector: fieldSelector,
		Channel:       make(chan WatchEvent, 100),
	}

	wm.watches[watchID] = watch
	return watch
}

// NotifyChange notifies all watchers of a change
func (wm *WatchManager) NotifyChange(resourceType string, eventType string, object interface{}) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for _, watch := range wm.watches {
		if watch.ResourceType == resourceType {
			select {
			case watch.Channel <- WatchEvent{
				Type:      eventType,
				Object:    object,
				Timestamp: time.Now(),
			}:
			default:
				// Channel full, skip
			}
		}
	}
}

// Helper functions
func matchesSelector(selector *LabelSelector, labels map[string]string) bool {
	for key, value := range selector.MatchLabels {
		if labels[key] != value {
			return false
		}
	}

	for _, req := range selector.MatchExpressions {
		switch req.Operator {
		case "In":
			found := false
			for _, val := range req.Values {
				if labels[req.Key] == val {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "NotIn":
			for _, val := range req.Values {
				if labels[req.Key] == val {
					return false
				}
			}
		case "Exists":
			if _, ok := labels[req.Key]; !ok {
				return false
			}
		case "DoesNotExist":
			if _, ok := labels[req.Key]; ok {
				return false
			}
		}
	}

	return true
}

func matchesFieldSelector(selector *FieldSelector, fields map[string]interface{}) bool {
	for _, req := range selector.Requirements {
		val, ok := fields[req.Field]
		if !ok {
			return false
		}

		valStr := fmt.Sprintf("%v", val)
		switch req.Operator {
		case "=", "==":
			if valStr != req.Value {
				return false
			}
		case "!=":
			if valStr == req.Value {
				return false
			}
		}
	}

	return true
}
