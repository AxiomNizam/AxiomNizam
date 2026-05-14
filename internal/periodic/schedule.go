// Package periodic implements a lightweight cron-style parser and
// dispatcher modelled on Nomad's periodic-job scheduler.
//
// The parser supports the standard 5-field cron grammar:
//
//	minute  hour  day-of-month  month  day-of-week
//
// Each field accepts: a number, a list (1,2,3), a range (1-5), a
// step (*/15), or "*" for wildcard.  Month and day-of-week also
// accept three-letter English abbreviations (mon, feb, ...).
//
// The dispatcher is not in this file — see dispatcher.go — because
// the parser is independently useful (for example, validating user
// input in an admission webhook before storage).
package periodic

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule is a parsed cron expression.  Zero value is invalid;
// always obtain one via Parse.
type Schedule struct {
	minute fieldMask // 0-59
	hour   fieldMask // 0-23
	dom    fieldMask // 1-31
	month  fieldMask // 1-12
	dow    fieldMask // 0-6, Sunday=0
	expr   string    // original text for String / error messages
}

// fieldMask is a bitset where bit i = 1 means "value i is permitted".
// 64 bits is enough for every cron field (max 60 values in minute).
type fieldMask uint64

// set marks value v in the mask.
func (f *fieldMask) set(v int) { *f |= 1 << v }

// matches tests membership.
func (f fieldMask) matches(v int) bool { return f&(1<<v) != 0 }

// Parse returns a validated Schedule or an error.
func Parse(expr string) (*Schedule, error) {
	expr = strings.TrimSpace(expr)
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("periodic: %q: expected 5 fields, got %d", expr, len(fields))
	}
	s := &Schedule{expr: expr}
	var err error
	if s.minute, err = parseField(fields[0], 0, 59, nil); err != nil {
		return nil, fmt.Errorf("periodic: minute: %w", err)
	}
	if s.hour, err = parseField(fields[1], 0, 23, nil); err != nil {
		return nil, fmt.Errorf("periodic: hour: %w", err)
	}
	if s.dom, err = parseField(fields[2], 1, 31, nil); err != nil {
		return nil, fmt.Errorf("periodic: day-of-month: %w", err)
	}
	if s.month, err = parseField(fields[3], 1, 12, monthAliases); err != nil {
		return nil, fmt.Errorf("periodic: month: %w", err)
	}
	if s.dow, err = parseField(fields[4], 0, 6, dowAliases); err != nil {
		return nil, fmt.Errorf("periodic: day-of-week: %w", err)
	}
	return s, nil
}

// monthAliases maps three-letter month names to their ordinal.
var monthAliases = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

// dowAliases — Sunday is 0 to match POSIX crontabs.
var dowAliases = map[string]int{
	"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
}

// parseField handles one cron field.
func parseField(s string, lo, hi int, aliases map[string]int) (fieldMask, error) {
	var mask fieldMask
	for _, token := range strings.Split(s, ",") {
		step := 1
		if slashIdx := strings.Index(token, "/"); slashIdx != -1 {
			var err error
			if step, err = strconv.Atoi(token[slashIdx+1:]); err != nil || step <= 0 {
				return 0, fmt.Errorf("invalid step %q", token)
			}
			token = token[:slashIdx]
		}
		a, b := lo, hi
		switch {
		case token == "*":
			// whole range
		case strings.Contains(token, "-"):
			parts := strings.SplitN(token, "-", 2)
			var err error
			if a, err = resolve(parts[0], aliases); err != nil {
				return 0, err
			}
			if b, err = resolve(parts[1], aliases); err != nil {
				return 0, err
			}
		default:
			v, err := resolve(token, aliases)
			if err != nil {
				return 0, err
			}
			a, b = v, v
		}
		if a < lo || b > hi || a > b {
			return 0, fmt.Errorf("value %d-%d out of range [%d,%d]", a, b, lo, hi)
		}
		for v := a; v <= b; v += step {
			mask.set(v)
		}
	}
	return mask, nil
}

// resolve parses a token as a number or alias.
func resolve(s string, aliases map[string]int) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if v, ok := aliases[s]; ok {
		return v, nil
	}
	return strconv.Atoi(s)
}

// String returns the original expression.
func (s *Schedule) String() string { return s.expr }

// Next returns the first time strictly after `after` that satisfies
// the schedule.  Returns the zero time if no such time exists
// within 5 years (defensive cap against infinite loops on
// pathological expressions like "31 * * 2 *").
func (s *Schedule) Next(after time.Time) time.Time {
	// Cron fires at the start of the minute — truncate and step once.
	t := after.Truncate(time.Minute).Add(time.Minute)
	deadline := after.Add(5 * 365 * 24 * time.Hour)
	for t.Before(deadline) {
		if !s.month.matches(int(t.Month())) {
			// Jump to the first day of next month.
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
			continue
		}
		if !s.dom.matches(t.Day()) || !s.dow.matches(int(t.Weekday())) {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
			continue
		}
		if !s.hour.matches(t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
			continue
		}
		if !s.minute.matches(t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}
		return t
	}
	return time.Time{}
}
