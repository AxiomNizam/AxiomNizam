// Package rand wraps math/rand with the generator-safety and
// character-set conveniences k8s controllers reach for: "give me a
// 5-character suffix for a generated name" without each caller
// rolling its own seed management.
package rand

import (
	"math/rand"
	"sync"
	"time"
)

// source is a process-global *rand.Rand guarded by a mutex so that
// callers from many goroutines can share one seed without racing.
// math/rand's top-level functions are safe for concurrent use in
// modern Go, but we keep a dedicated source so callers can Seed for
// tests.
var (
	src   = rand.New(rand.NewSource(time.Now().UnixNano()))
	srcMu sync.Mutex
)

// Seed replaces the global source — used only by tests that need
// reproducibility.
func Seed(seed int64) {
	srcMu.Lock()
	defer srcMu.Unlock()
	src = rand.New(rand.NewSource(seed))
}

// Intn returns a non-negative pseudo-random int in [0, n).
func Intn(n int) int {
	srcMu.Lock()
	defer srcMu.Unlock()
	return src.Intn(n)
}

// Int63nRange returns an int64 in [min, max).
func Int63nRange(min, max int64) int64 {
	if max <= min {
		return min
	}
	srcMu.Lock()
	defer srcMu.Unlock()
	return src.Int63n(max-min) + min
}

// alphabet is the lowercase-alphanumeric charset k8s uses for
// generated-name suffixes.  `g` and `o` are excluded because they
// are visually similar to `q` and `0` respectively in some fonts —
// matches upstream.
const alphabet = "bcdfghjklmnpqrstvwxz0123456789"

// String returns a random string of n characters drawn from
// alphabet.  Suitable for metadata.name suffixes, correlation IDs,
// and anything else that just needs "probably unique".
func String(n int) string {
	if n <= 0 {
		return ""
	}
	srcMu.Lock()
	defer srcMu.Unlock()
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[src.Intn(len(alphabet))]
	}
	return string(b)
}

// SafeEncodeString is the inverse of SafeEncode used by k8s when a
// random value might otherwise begin with a digit (which is illegal
// as a DNS1123 label's first character).  We prepend a letter when
// needed.
func SafeEncodeString(s string) string {
	if s == "" {
		return ""
	}
	if s[0] >= '0' && s[0] <= '9' {
		return "a" + s
	}
	return s
}
