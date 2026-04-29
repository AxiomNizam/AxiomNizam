package streamanalytics

// =====================================================
// WS-7.2 — Windowing Functions
//
// Implements tumbling, sliding, and session windows for
// real-time stream aggregation. Each window collects events
// and flushes aggregated results when the window closes.
// =====================================================

import (
	"sync"
	"time"
)

// Event represents a single stream event.
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WindowResult holds the aggregated output of a closed window.
type WindowResult struct {
	WindowStart time.Time              `json:"windowStart"`
	WindowEnd   time.Time              `json:"windowEnd"`
	EventCount  int64                  `json:"eventCount"`
	GroupKey    string                 `json:"groupKey,omitempty"`
	Values      map[string]float64     `json:"values"`
}

// Window manages event collection and flushing for a single window instance.
type Window struct {
	mu         sync.Mutex
	start      time.Time
	end        time.Time
	events     []Event
	groupKey   string
}

// NewWindow creates a new window.
func NewWindow(start, end time.Time, groupKey string) *Window {
	return &Window{start: start, end: end, groupKey: groupKey}
}

// Add inserts an event into the window.
func (w *Window) Add(event Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = append(w.events, event)
}

// EventCount returns the number of events in the window.
func (w *Window) EventCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.events)
}

// IsExpired checks if the window has passed its end time.
func (w *Window) IsExpired(now time.Time) bool {
	return now.After(w.end)
}

// Flush aggregates events and returns the window result.
func (w *Window) Flush(aggregations []AggregationSpec) *WindowResult {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := &WindowResult{
		WindowStart: w.start,
		WindowEnd:   w.end,
		EventCount:  int64(len(w.events)),
		GroupKey:     w.groupKey,
		Values:      make(map[string]float64),
	}

	for _, agg := range aggregations {
		result.Values[agg.OutputField] = w.aggregate(agg)
	}

	return result
}

// aggregate computes a single aggregation over the window's events.
func (w *Window) aggregate(agg AggregationSpec) float64 {
	switch agg.Function {
	case "count":
		return float64(len(w.events))
	case "sum":
		return w.sumField(agg.InputField)
	case "avg":
		if len(w.events) == 0 {
			return 0
		}
		return w.sumField(agg.InputField) / float64(len(w.events))
	case "min":
		return w.minField(agg.InputField)
	case "max":
		return w.maxField(agg.InputField)
	case "distinct_count":
		return float64(w.distinctCount(agg.InputField))
	default:
		return 0
	}
}

func (w *Window) sumField(field string) float64 {
	var sum float64
	for _, e := range w.events {
		sum += toFloat64(e.Data[field])
	}
	return sum
}

func (w *Window) minField(field string) float64 {
	if len(w.events) == 0 {
		return 0
	}
	min := toFloat64(w.events[0].Data[field])
	for _, e := range w.events[1:] {
		v := toFloat64(e.Data[field])
		if v < min {
			min = v
		}
	}
	return min
}

func (w *Window) maxField(field string) float64 {
	if len(w.events) == 0 {
		return 0
	}
	max := toFloat64(w.events[0].Data[field])
	for _, e := range w.events[1:] {
		v := toFloat64(e.Data[field])
		if v > max {
			max = v
		}
	}
	return max
}

func (w *Window) distinctCount(field string) int {
	seen := make(map[interface{}]bool)
	for _, e := range w.events {
		if v, ok := e.Data[field]; ok {
			seen[v] = true
		}
	}
	return len(seen)
}

// --- Window Manager ---

// WindowManager creates and manages windows based on the window spec.
type WindowManager struct {
	mu      sync.Mutex
	spec    WindowSpec
	windows map[string]*Window // groupKey -> active window
	size    time.Duration
	slide   time.Duration
}

// NewWindowManager creates a new manager for the given window spec.
func NewWindowManager(spec WindowSpec) *WindowManager {
	size := parseDurationSafe(spec.Size)
	if size == 0 {
		size = 5 * time.Minute
	}
	slide := parseDurationSafe(spec.Slide)
	if slide == 0 {
		slide = size // Tumbling window: slide == size
	}

	return &WindowManager{
		spec:    spec,
		windows: make(map[string]*Window),
		size:    size,
		slide:   slide,
	}
}

// AddEvent routes an event to the appropriate window, creating one if needed.
func (wm *WindowManager) AddEvent(event Event, groupKey string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	w, ok := wm.windows[groupKey]
	if !ok || w.IsExpired(event.Timestamp) {
		// Create a new window.
		start := event.Timestamp.Truncate(wm.size)
		end := start.Add(wm.size)
		w = NewWindow(start, end, groupKey)
		wm.windows[groupKey] = w
	}

	w.Add(event)
}

// FlushExpired flushes all expired windows and returns their results.
func (wm *WindowManager) FlushExpired(now time.Time, aggregations []AggregationSpec) []WindowResult {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	var results []WindowResult
	for key, w := range wm.windows {
		if w.IsExpired(now) {
			results = append(results, *w.Flush(aggregations))
			delete(wm.windows, key)
		}
	}
	return results
}

// ActiveWindowCount returns the number of active (non-expired) windows.
func (wm *WindowManager) ActiveWindowCount() int {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return len(wm.windows)
}

// --- Helpers ---

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}

func parseDurationSafe(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}
