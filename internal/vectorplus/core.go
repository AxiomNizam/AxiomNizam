package vectorplus

import (
	"math"
	"sort"
	"sync"
)

type Vector []float64

type Record struct {
	ID     string
	Vec    Vector
	Labels map[string]string
}

type SearchResult struct {
	ID       string
	Score    float64
	Distance float64
}

type Index struct {
	mu      sync.RWMutex
	dim     int
	records map[string]Record
}

func NewIndex(dim int) *Index {
	if dim < 1 {
		dim = 1
	}
	return &Index{dim: dim, records: make(map[string]Record)}
}

func (idx *Index) Upsert(r Record) bool {
	if len(r.Vec) != idx.dim || r.ID == "" {
		return false
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if r.Labels == nil {
		r.Labels = map[string]string{}
	}
	idx.records[r.ID] = r
	return true
}

func (idx *Index) Delete(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.records, id)
}

func dot(a, b Vector) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var out float64
	for i := 0; i < n; i++ {
		out += a[i] * b[i]
	}
	return out
}

func l2(a, b Vector) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var out float64
	for i := 0; i < n; i++ {
		d := a[i] - b[i]
		out += d * d
	}
	return math.Sqrt(out)
}

func (idx *Index) Search(query Vector, k int) []SearchResult {
	if len(query) != idx.dim || k < 1 {
		return nil
	}
	idx.mu.RLock()
	list := make([]Record, 0, len(idx.records))
	for _, r := range idx.records {
		list = append(list, r)
	}
	idx.mu.RUnlock()

	out := make([]SearchResult, 0, len(list))
	for _, r := range list {
		d := l2(query, r.Vec)
		s := dot(query, r.Vec)
		out = append(out, SearchResult{ID: r.ID, Score: s, Distance: d})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Distance == out[j].Distance {
			return out[i].Score > out[j].Score
		}
		return out[i].Distance < out[j].Distance
	})
	if len(out) > k {
		out = out[:k]
	}
	return out
}
