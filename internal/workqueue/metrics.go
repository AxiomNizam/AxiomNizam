// Package workqueue — Prometheus metrics provider.
//
// This mirrors the MetricsProvider contract from client-go.  A queue
// instrumented with a provider produces the standard k8s workqueue
// metrics, which operators can then scrape via the prometheus client:
//
//	workqueue_depth{name="controller-foo"}
//	workqueue_adds_total{name="controller-foo"}
//	workqueue_queue_duration_seconds{name="controller-foo"}
//	workqueue_work_duration_seconds{name="controller-foo"}
//	workqueue_retries_total{name="controller-foo"}
//	workqueue_longest_running_processor_seconds{name="controller-foo"}
//	workqueue_unfinished_work_seconds{name="controller-foo"}
//
// A single PrometheusProvider instance is safe to share across many
// queues; the "name" label distinguishes them.  Register the provider
// at process start with prometheus.DefaultRegisterer (or any custom
// registry) via the Register method.
package workqueue

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// GaugeMetric is the minimal interface a queue needs for depth /
// unfinished-work metrics.  It is an interface (rather than a
// concrete Gauge) so tests can plug in an in-memory double.
type GaugeMetric interface {
	Inc()
	Dec()
	Set(float64)
}

// CounterMetric tracks a monotonically-increasing value.
type CounterMetric interface {
	Inc()
}

// HistogramMetric records a distribution of observations.
type HistogramMetric interface {
	Observe(float64)
}

// SettableGaugeMetric is a gauge that the queue only writes via Set —
// used for the "longest-running processor" gauge which is sampled on
// a schedule rather than incremented per event.
type SettableGaugeMetric interface {
	Set(float64)
}

// MetricsProvider yields per-queue metrics handles.  Implementations
// are expected to memoise by name so that re-creating a queue with
// the same name reuses its existing Prometheus collectors.
type MetricsProvider interface {
	NewDepthMetric(name string) GaugeMetric
	NewAddsMetric(name string) CounterMetric
	NewLatencyMetric(name string) HistogramMetric
	NewWorkDurationMetric(name string) HistogramMetric
	NewUnfinishedWorkSecondsMetric(name string) SettableGaugeMetric
	NewLongestRunningProcessorSecondsMetric(name string) SettableGaugeMetric
	NewRetriesMetric(name string) CounterMetric
}

// NoopMetricsProvider discards every observation — the default when
// callers do not wire up monitoring.  Returning real (nil-safe) stubs
// rather than nil lets queues call methods unconditionally without
// nil checks on the hot path.
type NoopMetricsProvider struct{}

type noopGauge struct{}

func (noopGauge) Inc()          {}
func (noopGauge) Dec()          {}
func (noopGauge) Set(_ float64) {}

type noopCounter struct{}

func (noopCounter) Inc() {}

type noopHistogram struct{}

func (noopHistogram) Observe(_ float64) {}

// NewDepthMetric returns a no-op gauge.
func (NoopMetricsProvider) NewDepthMetric(_ string) GaugeMetric { return noopGauge{} }

// NewAddsMetric returns a no-op counter.
func (NoopMetricsProvider) NewAddsMetric(_ string) CounterMetric { return noopCounter{} }

// NewLatencyMetric returns a no-op histogram.
func (NoopMetricsProvider) NewLatencyMetric(_ string) HistogramMetric { return noopHistogram{} }

// NewWorkDurationMetric returns a no-op histogram.
func (NoopMetricsProvider) NewWorkDurationMetric(_ string) HistogramMetric { return noopHistogram{} }

// NewUnfinishedWorkSecondsMetric returns a no-op gauge.
func (NoopMetricsProvider) NewUnfinishedWorkSecondsMetric(_ string) SettableGaugeMetric {
	return noopGauge{}
}

// NewLongestRunningProcessorSecondsMetric returns a no-op gauge.
func (NoopMetricsProvider) NewLongestRunningProcessorSecondsMetric(_ string) SettableGaugeMetric {
	return noopGauge{}
}

// NewRetriesMetric returns a no-op counter.
func (NoopMetricsProvider) NewRetriesMetric(_ string) CounterMetric { return noopCounter{} }

// -----------------------------------------------------------------------------
// PrometheusProvider — real implementation
// -----------------------------------------------------------------------------

// PrometheusProvider creates collectors lazily and registers each one
// at most once with the supplied registerer.  Callers create the
// provider once at process start and reuse it across every queue.
type PrometheusProvider struct {
	registerer prometheus.Registerer

	// Each vec is created up-front but collectors for specific
	// queue-name labels are instantiated lazily.  We cache the
	// per-name handles behind a mutex to amortise the label lookup.
	depth          *prometheus.GaugeVec
	adds           *prometheus.CounterVec
	latency        *prometheus.HistogramVec
	workDuration   *prometheus.HistogramVec
	unfinishedWork *prometheus.GaugeVec
	longestRunning *prometheus.GaugeVec
	retries        *prometheus.CounterVec

	once sync.Once
	err  error
}

// NewPrometheusProvider constructs a provider that registers with r.
// Pass prometheus.DefaultRegisterer for the global registry.  The
// collectors are lazily created and registered on first access to
// avoid registering unused metrics.
func NewPrometheusProvider(r prometheus.Registerer) *PrometheusProvider {
	return &PrometheusProvider{registerer: r}
}

// init constructs the vectors and registers them with the registerer.
// It runs at most once; subsequent failures are cached in p.err so
// callers can surface them via RegistrationError.
func (p *PrometheusProvider) init() {
	p.once.Do(func() {
		p.depth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "workqueue_depth",
			Help: "Current number of items queued for processing.",
		}, []string{"name"})
		p.adds = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "workqueue_adds_total",
			Help: "Total number of Add calls observed on the queue.",
		}, []string{"name"})
		// Latency = time spent sitting in the queue before Get.
		// Histogram bucket choice mirrors client-go's ExponentialBuckets(10e-9, 10, 10)
		// expressed in seconds so the exported distribution is
		// compatible with off-the-shelf k8s dashboards.
		p.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "workqueue_queue_duration_seconds",
			Help:    "Time items spent in the queue before being picked up, in seconds.",
			Buckets: prometheus.ExponentialBuckets(10e-9, 10, 10),
		}, []string{"name"})
		// Work duration = time from Get to Done.
		p.workDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "workqueue_work_duration_seconds",
			Help:    "Time spent processing an item after dequeue, in seconds.",
			Buckets: prometheus.ExponentialBuckets(10e-9, 10, 10),
		}, []string{"name"})
		p.unfinishedWork = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "workqueue_unfinished_work_seconds",
			Help: "Sum of seconds that in-flight items have been processed so far.",
		}, []string{"name"})
		p.longestRunning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "workqueue_longest_running_processor_seconds",
			Help: "Age of the oldest currently-processing item, in seconds.",
		}, []string{"name"})
		p.retries = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "workqueue_retries_total",
			Help: "Total number of retries issued by the rate limiter.",
		}, []string{"name"})

		for _, c := range []prometheus.Collector{
			p.depth, p.adds, p.latency, p.workDuration,
			p.unfinishedWork, p.longestRunning, p.retries,
		} {
			if err := p.registerer.Register(c); err != nil {
				// AlreadyRegisteredError means another provider beat
				// us to the punch — that's fine as long as the
				// existing collector has the same type, which is the
				// usual case when a process constructs two providers
				// by mistake.  We tolerate it silently.
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					p.err = err
					return
				}
			}
		}
	})
}

// RegistrationError returns a non-nil error if any collector failed
// to register with the registerer.  Call after init() runs (i.e.
// after the first New*Metric call) to surface configuration bugs.
func (p *PrometheusProvider) RegistrationError() error {
	return p.err
}

// NewDepthMetric returns a gauge bound to name.
func (p *PrometheusProvider) NewDepthMetric(name string) GaugeMetric {
	p.init()
	return p.depth.WithLabelValues(name)
}

// NewAddsMetric returns a counter bound to name.
func (p *PrometheusProvider) NewAddsMetric(name string) CounterMetric {
	p.init()
	return p.adds.WithLabelValues(name)
}

// NewLatencyMetric returns the queue-duration histogram for name.
func (p *PrometheusProvider) NewLatencyMetric(name string) HistogramMetric {
	p.init()
	return p.latency.WithLabelValues(name)
}

// NewWorkDurationMetric returns the processing-duration histogram for name.
func (p *PrometheusProvider) NewWorkDurationMetric(name string) HistogramMetric {
	p.init()
	return p.workDuration.WithLabelValues(name)
}

// NewUnfinishedWorkSecondsMetric returns the in-flight work gauge.
func (p *PrometheusProvider) NewUnfinishedWorkSecondsMetric(name string) SettableGaugeMetric {
	p.init()
	return p.unfinishedWork.WithLabelValues(name)
}

// NewLongestRunningProcessorSecondsMetric returns the oldest-in-flight gauge.
func (p *PrometheusProvider) NewLongestRunningProcessorSecondsMetric(name string) SettableGaugeMetric {
	p.init()
	return p.longestRunning.WithLabelValues(name)
}

// NewRetriesMetric returns the retry counter.
func (p *PrometheusProvider) NewRetriesMetric(name string) CounterMetric {
	p.init()
	return p.retries.WithLabelValues(name)
}

// -----------------------------------------------------------------------------
// QueueMetrics — the per-queue bundle
// -----------------------------------------------------------------------------

// QueueMetrics bundles the per-name handles for one queue.  A queue
// constructs one at startup (via provider.Bind) and updates it from
// Add / Get / Done / retry paths.
type QueueMetrics struct {
	Depth          GaugeMetric
	Adds           CounterMetric
	Latency        HistogramMetric
	WorkDuration   HistogramMetric
	UnfinishedWork SettableGaugeMetric
	LongestRunning SettableGaugeMetric
	Retries        CounterMetric

	// processing tracks when each in-flight item was Get'd, so that
	// periodic sampling can compute UnfinishedWork and LongestRunning.
	// A workqueue that wants those gauges to tick should call
	// UpdateInFlight on a timer.
	mu         sync.Mutex
	processing map[string]time.Time
}

// Bind constructs a QueueMetrics for a queue named n using provider p.
func Bind(p MetricsProvider, n string) *QueueMetrics {
	if p == nil {
		p = NoopMetricsProvider{}
	}
	return &QueueMetrics{
		Depth:          p.NewDepthMetric(n),
		Adds:           p.NewAddsMetric(n),
		Latency:        p.NewLatencyMetric(n),
		WorkDuration:   p.NewWorkDurationMetric(n),
		UnfinishedWork: p.NewUnfinishedWorkSecondsMetric(n),
		LongestRunning: p.NewLongestRunningProcessorSecondsMetric(n),
		Retries:        p.NewRetriesMetric(n),
		processing:     map[string]time.Time{},
	}
}

// OnAdd records an enqueue.
func (q *QueueMetrics) OnAdd() { q.Adds.Inc(); q.Depth.Inc() }

// OnGet records a dequeue of key after it spent queueDuration waiting.
func (q *QueueMetrics) OnGet(key string, queueDuration time.Duration) {
	q.Depth.Dec()
	q.Latency.Observe(queueDuration.Seconds())
	q.mu.Lock()
	q.processing[key] = time.Now()
	q.mu.Unlock()
}

// OnDone records completion of key after workDuration of processing.
func (q *QueueMetrics) OnDone(key string, workDuration time.Duration) {
	q.WorkDuration.Observe(workDuration.Seconds())
	q.mu.Lock()
	delete(q.processing, key)
	q.mu.Unlock()
}

// OnRetry records a rate-limited requeue.
func (q *QueueMetrics) OnRetry() { q.Retries.Inc() }

// UpdateInFlight recomputes UnfinishedWork and LongestRunning from the
// tracked processing map.  Callers should invoke it on a timer
// (typically every 500ms) so dashboards observe the gauge without
// requiring queue traffic to refresh it.
func (q *QueueMetrics) UpdateInFlight() {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	var total float64
	var longest float64
	for _, started := range q.processing {
		age := now.Sub(started).Seconds()
		total += age
		if age > longest {
			longest = age
		}
	}
	q.UnfinishedWork.Set(total)
	q.LongestRunning.Set(longest)
}
