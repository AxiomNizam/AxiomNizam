// Package periodic — Dispatcher drives many Schedules, firing a
// caller-supplied callback at each scheduled instant.  Modelled on
// Nomad's periodic-dispatcher, simplified to a single goroutine
// using a min-heap of next-fire times.
//
// The dispatcher intentionally has no persistence: a process restart
// skips any fires that were due during the outage.  Callers that
// care about missed fires should checkpoint Last() to durable
// storage and rebuild state at boot.
package periodic

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// FireFunc is the user-supplied callback.
type FireFunc func(ctx context.Context, id string, scheduled time.Time)

// entry pairs a Schedule with its next-fire time.
type entry struct {
	id    string
	sched *Schedule
	next  time.Time
	index int // heap index
}

// jobHeap is a min-heap by next.
type jobHeap []*entry

// Len implements heap.Interface.
func (h jobHeap) Len() int { return len(h) }

// Less implements heap.Interface — earlier times first.
func (h jobHeap) Less(i, j int) bool { return h[i].next.Before(h[j].next) }

// Swap implements heap.Interface, tracking the mutated indices.
func (h jobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

// Push implements heap.Interface.
func (h *jobHeap) Push(x interface{}) {
	e := x.(*entry)
	e.index = len(*h)
	*h = append(*h, e)
}

// Pop implements heap.Interface.
func (h *jobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	e := old[n-1]
	e.index = -1
	*h = old[:n-1]
	return e
}

// Dispatcher fires schedules.
type Dispatcher struct {
	// Fire is the callback invoked in its own goroutine per fire.
	Fire FireFunc

	mu     sync.Mutex
	jobs   jobHeap
	byID   map[string]*entry
	wake   chan struct{}
	stopCh chan struct{}
	now    func() time.Time // overridable for tests
}

// NewDispatcher returns an empty dispatcher.  Callers must call
// Start before Adding schedules.
func NewDispatcher(fire FireFunc) *Dispatcher {
	return &Dispatcher{
		Fire:   fire,
		byID:   map[string]*entry{},
		wake:   make(chan struct{}, 1),
		stopCh: make(chan struct{}),
		now:    time.Now,
	}
}

// Start spins the dispatcher goroutine.
func (d *Dispatcher) Start() { go d.run() }

// Stop halts the goroutine; idempotent.
func (d *Dispatcher) Stop() {
	select {
	case <-d.stopCh:
	default:
		close(d.stopCh)
	}
}

// Add registers a schedule.  Replaces any prior registration under
// the same id — useful for "job spec updated, replan".
func (d *Dispatcher) Add(id string, s *Schedule) {
	d.mu.Lock()
	defer d.mu.Unlock()
	next := s.Next(d.now())
	if existing, ok := d.byID[id]; ok {
		heap.Remove(&d.jobs, existing.index)
	}
	e := &entry{id: id, sched: s, next: next}
	heap.Push(&d.jobs, e)
	d.byID[id] = e
	d.signal()
}

// Remove deregisters a schedule.
func (d *Dispatcher) Remove(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if e, ok := d.byID[id]; ok {
		heap.Remove(&d.jobs, e.index)
		delete(d.byID, id)
		d.signal()
	}
}

// signal non-blockingly pokes the run loop to re-evaluate.
func (d *Dispatcher) signal() {
	select {
	case d.wake <- struct{}{}:
	default:
	}
}

// run is the single goroutine driving fires.
func (d *Dispatcher) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		d.mu.Lock()
		var sleep time.Duration
		if d.jobs.Len() == 0 {
			sleep = time.Hour // idle until Add or Stop
		} else {
			sleep = time.Until(d.jobs[0].next)
			if sleep < 0 {
				sleep = 0
			}
		}
		d.mu.Unlock()

		timer := time.NewTimer(sleep)
		select {
		case <-d.stopCh:
			timer.Stop()
			return
		case <-d.wake:
			timer.Stop()
			continue
		case <-timer.C:
		}

		// Fire every job whose next time has arrived.
		d.mu.Lock()
		now := d.now()
		var fires []*entry
		for d.jobs.Len() > 0 && !d.jobs[0].next.After(now) {
			top := d.jobs[0]
			fires = append(fires, &entry{id: top.id, next: top.next})
			top.next = top.sched.Next(now)
			if top.next.IsZero() {
				// Schedule exhausted — drop from heap.
				heap.Remove(&d.jobs, top.index)
				delete(d.byID, top.id)
			} else {
				heap.Fix(&d.jobs, top.index)
			}
		}
		d.mu.Unlock()

		if d.Fire != nil {
			for _, f := range fires {
				go d.Fire(ctx, f.id, f.next)
			}
		}
	}
}
