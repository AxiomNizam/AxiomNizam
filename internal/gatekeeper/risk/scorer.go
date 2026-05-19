package risk

import (
	"context"
	"strings"
)

// CompositeScorer combines multiple scoring strategies.
type CompositeScorer struct {
	scorers []Scorer
	weights []float64
}

// NewCompositeScorer creates a scorer that combines multiple weighted scorers.
func NewCompositeScorer(scorers []Scorer, weights []float64) *CompositeScorer {
	return &CompositeScorer{scorers: scorers, weights: weights}
}

// Score calculates a weighted average across all scorers.
func (c *CompositeScorer) Score(ctx context.Context, signals *Signals) (int, error) {
	if len(c.scorers) == 0 {
		return 0, nil
	}

	var totalWeight, weightedScore float64
	for i, scorer := range c.scorers {
		score, err := scorer.Score(ctx, signals)
		if err != nil {
			continue
		}
		weight := 1.0
		if i < len(c.weights) {
			weight = c.weights[i]
		}
		weightedScore += float64(score) * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0, nil
	}
	return int(weightedScore / totalWeight), nil
}

// GeoScorer scores based on geographic signals only.
type GeoScorer struct{}

// Score calculates risk from geographic indicators.
func (g *GeoScorer) Score(ctx context.Context, signals *Signals) (int, error) {
	score := 0

	if signals.VPNDetected {
		score += 25
	}
	if signals.DatacenterIP {
		score += 15
	}
	if signals.ASNChange {
		score += 20
	}
	if signals.GeoDifference > 1000 {
		score += 25
	} else if signals.GeoDifference > 500 {
		score += 15
	}
	if signals.IPReputation > 70 {
		score += 20
	} else if signals.IPReputation > 50 {
		score += 10
	}

	if score > 100 {
		score = 100
	}
	return score, nil
}

// BehavioralScorer scores based on behavioral signals only.
type BehavioralScorer struct{}

// Score calculates risk from behavioral indicators.
func (b *BehavioralScorer) Score(ctx context.Context, signals *Signals) (int, error) {
	score := 0

	if signals.UnusualActivity {
		score += 25
	}
	if signals.SuspiciousLogin {
		score += 30
	}
	if signals.FrequentFailures {
		score += 20
	}
	if signals.FailureCount > 5 {
		score += 25
	} else if signals.FailureCount > 2 {
		score += 15
	}
	if signals.UnusualTimeOfDay {
		score += 10
	}

	if score > 100 {
		score = 100
	}
	return score, nil
}

// IsKnownGoodIP checks if an IP is from a known-good range (corporate VPN, etc).
func IsKnownGoodIP(ip string, knownRanges []string) bool {
	for _, r := range knownRanges {
		if strings.HasPrefix(ip, r) {
			return true
		}
	}
	return false
}
