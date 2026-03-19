package waitx

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	RetryLinear      = "linear"
	RetryExponential = "exponential"
	RetryFibonacci   = "fibonacci"
	RetryCustom      = "custom"
)

// LinearRetry uses a fixed delay.
type LinearRetry struct{}

func (LinearRetry) NextDelay(_ int, baseInterval, maxInterval time.Duration, _ error) time.Duration {
	if baseInterval <= 0 {
		baseInterval = time.Second
	}
	return clampDuration(baseInterval, maxInterval)
}

// ExponentialRetry uses interval * coefficient^(attempt-1).
type ExponentialRetry struct {
	Coefficient float64
}

func (e ExponentialRetry) NextDelay(attempt int, baseInterval, maxInterval time.Duration, _ error) time.Duration {
	if baseInterval <= 0 {
		baseInterval = time.Second
	}
	if attempt < 1 {
		attempt = 1
	}
	coef := e.Coefficient
	if coef <= 1 {
		coef = 2
	}

	delay := float64(baseInterval) * math.Pow(coef, float64(attempt-1))
	if delay > float64(math.MaxInt64) {
		return clampDuration(time.Duration(math.MaxInt64), maxInterval)
	}
	return clampDuration(time.Duration(delay), maxInterval)
}

// FibonacciRetry grows delay using fibonacci numbers.
type FibonacciRetry struct{}

func (FibonacciRetry) NextDelay(attempt int, baseInterval, maxInterval time.Duration, _ error) time.Duration {
	if baseInterval <= 0 {
		baseInterval = time.Second
	}
	if attempt < 1 {
		attempt = 1
	}

	fib := fibonacci(attempt)
	delay := time.Duration(fib) * baseInterval
	return clampDuration(delay, maxInterval)
}

// CustomSequenceRetry uses a caller-defined sequence and then a fallback strategy.
type CustomSequenceRetry struct {
	Sequence []time.Duration
	Fallback RetryStrategy
}

func (c CustomSequenceRetry) NextDelay(attempt int, baseInterval, maxInterval time.Duration, lastErr error) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	idx := attempt - 1
	if idx < len(c.Sequence) {
		return clampDuration(c.Sequence[idx], maxInterval)
	}

	if c.Fallback != nil {
		return c.Fallback.NextDelay(attempt, baseInterval, maxInterval, lastErr)
	}

	if len(c.Sequence) > 0 {
		return clampDuration(c.Sequence[len(c.Sequence)-1], maxInterval)
	}

	return clampDuration(baseInterval, maxInterval)
}

// NewRetryStrategy builds a strategy from CLI-style options.
func NewRetryStrategy(policy string, coefficient float64, customSequence []time.Duration) (RetryStrategy, error) {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "", RetryLinear:
		return LinearRetry{}, nil
	case RetryExponential:
		return ExponentialRetry{Coefficient: coefficient}, nil
	case RetryFibonacci:
		return FibonacciRetry{}, nil
	case RetryCustom:
		if len(customSequence) == 0 {
			return nil, fmt.Errorf("retry strategy %q requires a non-empty custom sequence", RetryCustom)
		}
		return CustomSequenceRetry{
			Sequence: customSequence,
			Fallback: ExponentialRetry{Coefficient: coefficient},
		}, nil
	default:
		return nil, fmt.Errorf("unknown retry strategy %q", policy)
	}
}

// ParseDurationSequence parses comma-separated durations (e.g. 500ms,1s,2s).
func ParseDurationSequence(raw string) ([]time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	sequence := make([]time.Duration, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		d, err := time.ParseDuration(item)
		if err != nil {
			return nil, fmt.Errorf("invalid duration %q in retry sequence: %w", item, err)
		}
		sequence = append(sequence, d)
	}
	if len(sequence) == 0 {
		return nil, fmt.Errorf("retry sequence is empty")
	}
	return sequence, nil
}

func clampDuration(delay, maxInterval time.Duration) time.Duration {
	if delay < 0 {
		delay = 0
	}
	if maxInterval > 0 && delay > maxInterval {
		return maxInterval
	}
	return delay
}

func fibonacci(n int) int64 {
	if n <= 2 {
		return 1
	}

	var a int64 = 1
	var b int64 = 1
	for i := 3; i <= n; i++ {
		a, b = b, a+b
		if b < 0 {
			return math.MaxInt64
		}
	}
	return b
}
