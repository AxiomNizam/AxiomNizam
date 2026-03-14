package reviewflow

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type Stage string

const (
	StageDraft    Stage = "draft"
	StageReview   Stage = "review"
	StageApproved Stage = "approved"
	StageRejected Stage = "rejected"
	StageMerged   Stage = "merged"
)

type ReviewItem struct {
	ID          string
	Title       string
	Description string
	Author      string
	Tags        []string
	Score       float64
	Stage       Stage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Pipeline struct {
	mu    sync.RWMutex
	items map[string]ReviewItem
}

func NewPipeline() *Pipeline {
	return &Pipeline{items: make(map[string]ReviewItem)}
}

func (p *Pipeline) Upsert(item ReviewItem) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	if item.Stage == "" {
		item.Stage = StageDraft
	}
	p.items[item.ID] = item
}

func (p *Pipeline) Get(id string) (ReviewItem, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	item, ok := p.items[id]
	return item, ok
}

func (p *Pipeline) ListByStage(stage Stage) []ReviewItem {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]ReviewItem, 0, len(p.items))
	for _, item := range p.items {
		if stage == "" || item.Stage == stage {
			out = append(out, item)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].UpdatedAt.After(out[j].UpdatedAt)
		}
		return out[i].Score > out[j].Score
	})
	return out
}

func (p *Pipeline) Advance(id string, target Stage) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	item, ok := p.items[id]
	if !ok {
		return false
	}
	item.Stage = target
	item.UpdatedAt = time.Now().UTC()
	p.items[id] = item
	return true
}

func ScoreBySignals(title, desc string, tags []string) float64 {
	score := 0.0
	score += float64(len(strings.Fields(title))) * 0.2
	score += float64(len(strings.Fields(desc))) * 0.05
	score += float64(len(tags)) * 0.7
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		switch t {
		case "security", "reliability", "breaking-change":
			score += 2.0
		case "bugfix", "performance":
			score += 1.3
		case "docs", "chore":
			score += 0.2
		}
	}
	return score
}
