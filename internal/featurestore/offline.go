package featurestore

// =====================================================
// WS-7.1 — Offline Feature Store Backend
//
// Provides batch feature retrieval for model training.
// Supports point-in-time joins, historical feature retrieval,
// and dataset generation for ML pipelines.
// =====================================================

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HistoricalFeatureRequest defines a request for historical features.
type HistoricalFeatureRequest struct {
	GroupName     string            `json:"groupName"`
	EntityKeys   []string          `json:"entityKeys"`
	Features     []string          `json:"features,omitempty"` // Empty = all features
	StartTime    *time.Time        `json:"startTime,omitempty"`
	EndTime      *time.Time        `json:"endTime,omitempty"`
	PointInTime  *time.Time        `json:"pointInTime,omitempty"` // For point-in-time joins
	Limit        int               `json:"limit,omitempty"`
}

// TrainingDataset represents a generated training dataset.
type TrainingDataset struct {
	GroupName    string                   `json:"groupName"`
	Features     []string                `json:"features"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int64                   `json:"rowCount"`
	GeneratedAt  time.Time               `json:"generatedAt"`
	TimeRange    *TimeRange              `json:"timeRange,omitempty"`
}

// TimeRange represents a time window for historical queries.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OfflineStore provides batch feature retrieval for ML training.
type OfflineStore interface {
	// GetHistoricalFeatures retrieves historical feature values.
	GetHistoricalFeatures(ctx context.Context, req HistoricalFeatureRequest) (*TrainingDataset, error)

	// GenerateTrainingDataset creates a training dataset with point-in-time correctness.
	GenerateTrainingDataset(ctx context.Context, groups []string, entityEvents []EntityEvent) (*TrainingDataset, error)

	// WriteFeatures writes materialized features to the offline store.
	WriteFeatures(ctx context.Context, group string, vectors []*FeatureVector) error

	// Stats returns store statistics.
	Stats() OfflineStoreStats
}

// EntityEvent represents a labeled event for point-in-time joins.
type EntityEvent struct {
	EntityKey string    `json:"entityKey"`
	EventTime time.Time `json:"eventTime"`
	Label     string    `json:"label,omitempty"`
}

// OfflineStoreStats tracks offline store metrics.
type OfflineStoreStats struct {
	Backend       string `json:"backend"`
	TotalRows     int64  `json:"totalRows"`
	TotalGroups   int    `json:"totalGroups"`
	TotalQueries  int64  `json:"totalQueries"`
	LastWriteAt   *time.Time `json:"lastWriteAt,omitempty"`
}

// MemoryOfflineStore is an in-memory offline store for development.
type MemoryOfflineStore struct {
	mu      sync.RWMutex
	// group -> entityKey -> []FeatureVector (time-ordered)
	data    map[string]map[string][]FeatureVector
	stats   OfflineStoreStats
}

// NewMemoryOfflineStore creates a new in-memory offline store.
func NewMemoryOfflineStore() *MemoryOfflineStore {
	return &MemoryOfflineStore{
		data:  make(map[string]map[string][]FeatureVector),
		stats: OfflineStoreStats{Backend: "memory"},
	}
}

func (s *MemoryOfflineStore) GetHistoricalFeatures(ctx context.Context, req HistoricalFeatureRequest) (*TrainingDataset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.stats.TotalQueries++

	groupData, ok := s.data[req.GroupName]
	if !ok {
		return &TrainingDataset{GroupName: req.GroupName, GeneratedAt: time.Now()}, nil
	}

	var rows []map[string]interface{}
	for _, ek := range req.EntityKeys {
		vectors, ok := groupData[ek]
		if !ok {
			continue
		}

		for _, v := range vectors {
			// Apply time filters.
			if req.StartTime != nil && v.Timestamp.Before(*req.StartTime) {
				continue
			}
			if req.EndTime != nil && v.Timestamp.After(*req.EndTime) {
				continue
			}

			row := map[string]interface{}{
				"entity_key": v.EntityKey,
				"timestamp":  v.Timestamp,
			}
			for k, val := range v.Features {
				// Filter features if specified.
				if len(req.Features) > 0 && !containsStr(req.Features, k) {
					continue
				}
				row[k] = val
			}
			rows = append(rows, row)

			if req.Limit > 0 && len(rows) >= req.Limit {
				break
			}
		}
	}

	features := req.Features
	if len(features) == 0 && len(rows) > 0 {
		for k := range rows[0] {
			if k != "entity_key" && k != "timestamp" {
				features = append(features, k)
			}
		}
	}

	return &TrainingDataset{
		GroupName:   req.GroupName,
		Features:   features,
		Rows:       rows,
		RowCount:   int64(len(rows)),
		GeneratedAt: time.Now(),
	}, nil
}

func (s *MemoryOfflineStore) GenerateTrainingDataset(ctx context.Context, groups []string, entityEvents []EntityEvent) (*TrainingDataset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.stats.TotalQueries++

	var rows []map[string]interface{}

	for _, event := range entityEvents {
		row := map[string]interface{}{
			"entity_key": event.EntityKey,
			"event_time": event.EventTime,
			"label":      event.Label,
		}

		// Point-in-time join: for each group, get the latest feature vector before event time.
		for _, group := range groups {
			groupData, ok := s.data[group]
			if !ok {
				continue
			}
			vectors, ok := groupData[event.EntityKey]
			if !ok {
				continue
			}

			// Find the latest vector before event time.
			var best *FeatureVector
			for i := range vectors {
				if vectors[i].Timestamp.Before(event.EventTime) || vectors[i].Timestamp.Equal(event.EventTime) {
					if best == nil || vectors[i].Timestamp.After(best.Timestamp) {
						best = &vectors[i]
					}
				}
			}

			if best != nil {
				for k, v := range best.Features {
					row[fmt.Sprintf("%s__%s", group, k)] = v
				}
			}
		}

		rows = append(rows, row)
	}

	return &TrainingDataset{
		Rows:        rows,
		RowCount:    int64(len(rows)),
		GeneratedAt: time.Now(),
	}, nil
}

func (s *MemoryOfflineStore) WriteFeatures(ctx context.Context, group string, vectors []*FeatureVector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[group]; !ok {
		s.data[group] = make(map[string][]FeatureVector)
		s.stats.TotalGroups++
	}

	for _, v := range vectors {
		s.data[group][v.EntityKey] = append(s.data[group][v.EntityKey], *v)
		s.stats.TotalRows++
	}

	now := time.Now()
	s.stats.LastWriteAt = &now
	return nil
}

func (s *MemoryOfflineStore) Stats() OfflineStoreStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
