package featurestore

// =====================================================
// WS-7.1 — Feature Metadata Registry
//
// Centralized registry for feature group metadata, feature
// definitions, lineage tracking, and discovery. Provides
// search, dependency resolution, and feature reuse tracking.
// =====================================================

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// FeatureMetadata describes a registered feature with lineage and usage info.
type FeatureMetadata struct {
	GroupName    string    `json:"groupName"`
	FeatureName  string   `json:"featureName"`
	Type         string   `json:"type"`
	Description  string   `json:"description,omitempty"`
	Owner        string   `json:"owner,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Transform    string   `json:"transform,omitempty"`
	Source       string   `json:"source,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	// Lineage
	UpstreamAssets  []string `json:"upstreamAssets,omitempty"`
	DownstreamModels []string `json:"downstreamModels,omitempty"`
	// Usage stats
	ServingRequests int64 `json:"servingRequests"`
	TrainingJobs    int64 `json:"trainingJobs"`
}

// FeatureSearchResult represents a search result from the registry.
type FeatureSearchResult struct {
	Features  []FeatureMetadata `json:"features"`
	Total     int               `json:"total"`
	Query     string            `json:"query"`
}

// RegistryStats exposes registry metrics.
type RegistryStats struct {
	TotalGroups   int   `json:"totalGroups"`
	TotalFeatures int   `json:"totalFeatures"`
	TotalSearches int64 `json:"totalSearches"`
}

// Registry provides centralized feature metadata management.
type Registry struct {
	mu       sync.RWMutex
	features map[string]*FeatureMetadata // "group/feature" -> metadata
	groups   map[string][]string          // group -> feature names
	stats    RegistryStats
	searches int64
}

// NewRegistry creates a new feature metadata registry.
func NewRegistry() *Registry {
	return &Registry{
		features: make(map[string]*FeatureMetadata),
		groups:   make(map[string][]string),
	}
}

// Register adds or updates feature metadata in the registry.
func (r *Registry) Register(meta *FeatureMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s/%s", meta.GroupName, meta.FeatureName)
	now := time.Now()

	if existing, ok := r.features[key]; ok {
		meta.CreatedAt = existing.CreatedAt
		meta.ServingRequests = existing.ServingRequests
		meta.TrainingJobs = existing.TrainingJobs
	} else {
		meta.CreatedAt = now
		// Track group membership.
		r.groups[meta.GroupName] = appendUniqueStr(r.groups[meta.GroupName], meta.FeatureName)
	}
	meta.UpdatedAt = now
	r.features[key] = meta
}

// RegisterGroup registers all features from a FeatureGroupResource.
func (r *Registry) RegisterGroup(fg *FeatureGroupResource) {
	for _, f := range fg.Spec.Features {
		r.Register(&FeatureMetadata{
			GroupName:   fg.Name,
			FeatureName: f.Name,
			Type:        f.Type,
			Description: f.Description,
			Owner:       fg.Spec.Owner,
			Tags:        fg.Spec.Tags,
			Transform:   f.Transform,
			Source:      fg.Spec.Source.DataSourceRef,
			UpstreamAssets: []string{fg.Spec.Source.DataSourceRef},
		})
	}
}

// Get retrieves metadata for a specific feature.
func (r *Registry) Get(group, feature string) (*FeatureMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := fmt.Sprintf("%s/%s", group, feature)
	meta, ok := r.features[key]
	return meta, ok
}

// ListGroup returns all features in a group.
func (r *Registry) ListGroup(group string) []FeatureMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names, ok := r.groups[group]
	if !ok {
		return nil
	}

	var result []FeatureMetadata
	for _, name := range names {
		key := fmt.Sprintf("%s/%s", group, name)
		if meta, ok := r.features[key]; ok {
			result = append(result, *meta)
		}
	}
	return result
}

// Search finds features matching a text query across names, descriptions, and tags.
func (r *Registry) Search(query string) *FeatureSearchResult {
	r.mu.Lock()
	r.searches++
	r.mu.Unlock()

	r.mu.RLock()
	defer r.mu.RUnlock()

	lower := strings.ToLower(query)
	var matches []FeatureMetadata

	for _, meta := range r.features {
		if strings.Contains(strings.ToLower(meta.FeatureName), lower) ||
			strings.Contains(strings.ToLower(meta.Description), lower) ||
			strings.Contains(strings.ToLower(meta.GroupName), lower) ||
			containsAnyStr(meta.Tags, lower) {
			matches = append(matches, *meta)
		}
	}

	return &FeatureSearchResult{
		Features: matches,
		Total:    len(matches),
		Query:    query,
	}
}

// RecordUsage tracks feature usage for lineage.
func (r *Registry) RecordUsage(group, feature, usageType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s/%s", group, feature)
	meta, ok := r.features[key]
	if !ok {
		return
	}

	switch usageType {
	case "serving":
		meta.ServingRequests++
	case "training":
		meta.TrainingJobs++
	}
}

// AddDownstreamModel records that a model depends on a feature.
func (r *Registry) AddDownstreamModel(group, feature, modelName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s/%s", group, feature)
	meta, ok := r.features[key]
	if !ok {
		return
	}
	meta.DownstreamModels = appendUniqueStr(meta.DownstreamModels, modelName)
}

// Stats returns registry statistics.
func (r *Registry) Stats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return RegistryStats{
		TotalGroups:   len(r.groups),
		TotalFeatures: len(r.features),
		TotalSearches: r.searches,
	}
}

// --- Helpers ---

func appendUniqueStr(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

func containsAnyStr(slice []string, substr string) bool {
	for _, v := range slice {
		if strings.Contains(strings.ToLower(v), substr) {
			return true
		}
	}
	return false
}
