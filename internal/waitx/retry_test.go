package waitx

import (
	"testing"
	"time"
)

func TestExponentialRetryNextDelay(t *testing.T) {
	strategy := ExponentialRetry{Coefficient: 2}

	d1 := strategy.NextDelay(1, 100*time.Millisecond, 3*time.Second, nil)
	d2 := strategy.NextDelay(2, 100*time.Millisecond, 3*time.Second, nil)
	d3 := strategy.NextDelay(3, 100*time.Millisecond, 3*time.Second, nil)

	if d1 != 100*time.Millisecond || d2 != 200*time.Millisecond || d3 != 400*time.Millisecond {
		t.Fatalf("unexpected exponential sequence: %s %s %s", d1, d2, d3)
	}
}

func TestParseDurationSequence(t *testing.T) {
	seq, err := ParseDurationSequence("100ms, 1s,2s")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(seq) != 3 {
		t.Fatalf("expected 3 durations, got %d", len(seq))
	}
	if seq[0] != 100*time.Millisecond || seq[1] != time.Second || seq[2] != 2*time.Second {
		t.Fatalf("unexpected sequence values: %v", seq)
	}
}

func TestNewRetryStrategyCustomRequiresSequence(t *testing.T) {
	_, err := NewRetryStrategy(RetryCustom, 2, nil)
	if err == nil {
		t.Fatal("expected error for custom retry strategy without sequence")
	}
}
