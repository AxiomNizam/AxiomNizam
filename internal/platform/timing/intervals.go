// Package timing centralises the interval / resync / backoff constants
// that controllers, reconcilers, and schedulers use.
//
// P3.4 — previously `5 * time.Minute` / `1 * time.Minute` / etc. were
// scattered across controllers, reconcilers, and jobs.  Anything long-
// lived should now come from here so we can tune them in one place.
package timing

import "time"

const (
	// DefaultResyncPeriod is the cadence at which a controller should
	// re-list and diff its backing store against runtime, catching
	// missed events.  5m is the Kubernetes default and works well for
	// resources whose spec changes in minutes, not seconds.
	DefaultResyncPeriod = 5 * time.Minute

	// DefaultRequeueAfter is the default "try again later" delay a
	// reconciler returns when no error happened but steady-state has
	// not been reached (e.g. waiting for an external dependency).
	DefaultRequeueAfter = 1 * time.Minute

	// DefaultMaxBackoff caps exponential-backoff requeues inside the
	// reconciler.  The workqueue rate limiter has its own base/max
	// knobs; this is the upper bound for per-item backoff.
	DefaultMaxBackoff = 1 * time.Minute

	// DefaultBucketResyncInterval is the resync cadence used by the
	// storage-bucket controller when not configured otherwise.
	DefaultBucketResyncInterval = 2 * time.Minute

	// DefaultSchedulerTick is the cadence of the jobs scheduler's
	// "are any scheduled jobs due?" sweep.  1m mirrors the legacy
	// SimpleScheduler behaviour.
	DefaultSchedulerTick = 1 * time.Minute

	// DefaultPersistenceFlushInterval is the cadence used by the
	// persistent job queue to mirror the in-memory queue to the
	// underlying repository.
	DefaultPersistenceFlushInterval = 30 * time.Second

	// DefaultAdmissionCacheTTL is how long admission decisions are
	// cached before being re-evaluated.
	DefaultAdmissionCacheTTL = 5 * time.Minute

	// DefaultDLQSweepInterval is the cadence at which the dead-letter
	// queue scans for items to expire.
	DefaultDLQSweepInterval = 24 * time.Hour
)
