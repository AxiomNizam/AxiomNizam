// Package metav1 — RFC3339 Time + ISO-8601 Duration wrapper types.
//
// time.Time and time.Duration JSON-encode to forms that are
// inconvenient for kubectl and YAML consumers.  These wrappers emit
// the formats k8s has standardised on.
package metav1

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Time is a time.Time with RFC3339 JSON encoding.  The zero value
// encodes to JSON null so that optional timestamp fields elide
// cleanly when unset.
type Time struct {
	time.Time
}

// NewTime constructs a Time from a time.Time.
func NewTime(t time.Time) Time { return Time{Time: t} }

// Now returns the current wall-clock wrapped in Time.
func Now() Time { return Time{Time: time.Now()} }

// IsZero reports whether t is unset.
func (t *Time) IsZero() bool { return t == nil || t.Time.IsZero() }

// MarshalJSON writes "null" for the zero value, else an RFC3339 string.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	// Use UTC and strip sub-second precision for stable diffs.
	return []byte("\"" + t.Time.UTC().Format(time.RFC3339) + "\""), nil
}

// UnmarshalJSON accepts an RFC3339 string, null, or an empty string.
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" || s == "\"\"" || s == "" {
		t.Time = time.Time{}
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// Duration is a time.Duration with the Go stdlib string form ("5s",
// "1h30m") rather than a nanosecond integer.
type Duration struct {
	time.Duration
}

// MarshalJSON emits "5s" / "1h30m" style literals.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte("\"" + d.Duration.String() + "\""), nil
}

// UnmarshalJSON accepts "5s" or a bare number of seconds.
func (d *Duration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(strings.TrimSpace(string(b)), "\"")
	if s == "" {
		d.Duration = 0
		return nil
	}
	// Bare number → seconds.
	if sec, err := strconv.ParseFloat(s, 64); err == nil && !strings.ContainsAny(s, "hmsun") {
		d.Duration = time.Duration(sec * float64(time.Second))
		return nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	return nil
}

// MicroTime is Time with microsecond precision — used for event
// timestamps where sub-second ordering matters.
type MicroTime struct {
	time.Time
}

// MarshalJSON preserves microseconds via the "2006-01-02T15:04:05.000000Z07:00"
// layout; the compact k8s form is close enough.
func (mt MicroTime) MarshalJSON() ([]byte, error) {
	if mt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte("\"" + mt.Time.UTC().Format("2006-01-02T15:04:05.000000Z07:00") + "\""), nil
}

// iso8601Re is ISO-8601 duration ("P1Y2M3DT4H5M6S") — not supported
// by time.ParseDuration.  k8s usually uses the Go form; we keep this
// pattern here in case we wire it in later.
var iso8601Re = regexp.MustCompile(`^P((\d+)Y)?((\d+)M)?((\d+)D)?(T((\d+)H)?((\d+)M)?((\d+(\.\d+)?)S)?)?$`)

// ParseISO8601Duration accepts the XSD-style P-form.  Exposed so
// callers interoperating with non-Go clients have an option.
func ParseISO8601Duration(s string) (time.Duration, error) {
	if !iso8601Re.MatchString(s) {
		return 0, fmt.Errorf("not an ISO-8601 duration: %q", s)
	}
	// Deliberately minimal — years/months are ambiguous without a
	// reference date.  We only support days/hours/minutes/seconds.
	var total time.Duration
	m := iso8601Re.FindStringSubmatch(s)
	if m[4] != "" {
		days, _ := strconv.Atoi(m[6])
		total += time.Duration(days) * 24 * time.Hour
	}
	if m[9] != "" {
		hours, _ := strconv.Atoi(m[9])
		total += time.Duration(hours) * time.Hour
	}
	if m[11] != "" {
		mins, _ := strconv.Atoi(m[11])
		total += time.Duration(mins) * time.Minute
	}
	if m[13] != "" {
		secs, _ := strconv.ParseFloat(m[13], 64)
		total += time.Duration(secs * float64(time.Second))
	}
	return total, nil
}
