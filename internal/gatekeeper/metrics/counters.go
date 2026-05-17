package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector gathers metrics for MFA operations.
type Collector struct {
	mu sync.RWMutex

	// Counter metrics
	EnrollmentsTotal      prometheus.Counter
	VerificationsTotal    prometheus.Counter
	VerificationFailures  prometheus.Counter
	BackupCodesUsed       prometheus.Counter
	TrustedDevicesCreated prometheus.Counter

	// Gauge metrics
	ActiveFactorsTotal prometheus.Gauge
	HighRiskEvents     prometheus.Gauge

	// Histogram metrics (latency)
	VerificationDuration prometheus.Histogram
	EnrollmentDuration   prometheus.Histogram
}

// NewCollector creates a new metrics collector with Prometheus.
func NewCollector() *Collector {
	return &Collector{
		EnrollmentsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "mfa",
			Name:      "enrollments_total",
			Help:      "Total number of factor enrollments",
		}),
		VerificationsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "mfa",
			Name:      "verifications_total",
			Help:      "Total number of successful MFA verifications",
		}),
		VerificationFailures: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "mfa",
			Name:      "verification_failures_total",
			Help:      "Total number of failed MFA verifications",
		}),
		BackupCodesUsed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "mfa",
			Name:      "backup_codes_used_total",
			Help:      "Total number of backup codes consumed",
		}),
		TrustedDevicesCreated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "mfa",
			Name:      "trusted_devices_total",
			Help:      "Total number of trusted devices created",
		}),
		ActiveFactorsTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "mfa",
			Name:      "active_factors",
			Help:      "Number of active MFA factors",
		}),
		HighRiskEvents: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "mfa",
			Name:      "high_risk_events",
			Help:      "Number of high-risk authentication events",
		}),
		VerificationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "mfa",
			Name:      "verification_duration_seconds",
			Help:      "Time taken to verify MFA code",
			Buckets:   prometheus.DefBuckets,
		}),
		EnrollmentDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "mfa",
			Name:      "enrollment_duration_seconds",
			Help:      "Time taken to complete MFA enrollment",
			Buckets:   prometheus.DefBuckets,
		}),
	}
}

// RecordEnrollment increments enrollment counter.
func (c *Collector) RecordEnrollment() {
	c.EnrollmentsTotal.Inc()
}

// RecordVerification increments verification counter.
func (c *Collector) RecordVerification() {
	c.VerificationsTotal.Inc()
}

// RecordVerificationFailure increments failure counter.
func (c *Collector) RecordVerificationFailure() {
	c.VerificationFailures.Inc()
}

// RecordBackupCodeUsed increments backup code usage counter.
func (c *Collector) RecordBackupCodeUsed() {
	c.BackupCodesUsed.Inc()
}

// RecordTrustedDeviceCreated increments trusted device counter.
func (c *Collector) RecordTrustedDeviceCreated() {
	c.TrustedDevicesCreated.Inc()
}

// SetActiveFactors sets the current number of active factors.
func (c *Collector) SetActiveFactors(count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ActiveFactorsTotal.Set(count)
}

// SetHighRiskEvents sets the current number of high-risk events.
func (c *Collector) SetHighRiskEvents(count float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HighRiskEvents.Set(count)
}

// RecordVerificationDuration records verification latency.
func (c *Collector) RecordVerificationDuration(duration time.Duration) {
	c.VerificationDuration.Observe(duration.Seconds())
}

// RecordEnrollmentDuration records enrollment latency.
func (c *Collector) RecordEnrollmentDuration(duration time.Duration) {
	c.EnrollmentDuration.Observe(duration.Seconds())
}

// SimpleMetrics provides a non-Prometheus implementation for testing.
type SimpleMetrics struct {
	mu                      sync.RWMutex
	Enrollments             int64
	Verifications           int64
	VerificationFailures    int64
	BackupCodesUsed         int64
	TrustedDevicesCreated   int64
	ActiveFactorsCount      int64
	HighRiskEventsCount     int64
	VerificationDurationSum time.Duration
	EnrollmentDurationSum   time.Duration
}

// RecordEnrollment increments enrollment counter.
func (s *SimpleMetrics) RecordEnrollment() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Enrollments++
}

// RecordVerification increments verification counter.
func (s *SimpleMetrics) RecordVerification() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Verifications++
}

// RecordVerificationFailure increments failure counter.
func (s *SimpleMetrics) RecordVerificationFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.VerificationFailures++
}

// RecordBackupCodeUsed increments backup code counter.
func (s *SimpleMetrics) RecordBackupCodeUsed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BackupCodesUsed++
}

// RecordTrustedDeviceCreated increments trusted device counter.
func (s *SimpleMetrics) RecordTrustedDeviceCreated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TrustedDevicesCreated++
}

// SetActiveFactors sets active factors gauge.
func (s *SimpleMetrics) SetActiveFactors(count float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActiveFactorsCount = int64(count)
}

// SetHighRiskEvents sets high-risk events gauge.
func (s *SimpleMetrics) SetHighRiskEvents(count float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HighRiskEventsCount = int64(count)
}

// RecordVerificationDuration records verification latency.
func (s *SimpleMetrics) RecordVerificationDuration(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.VerificationDurationSum += duration
}

// RecordEnrollmentDuration records enrollment latency.
func (s *SimpleMetrics) RecordEnrollmentDuration(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.EnrollmentDurationSum += duration
}
