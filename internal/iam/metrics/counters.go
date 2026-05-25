package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ─────────────────────────────────────────────────────────────────────────────
// Counter metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	AuthAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "auth_attempts_total",
		Help:      "Total number of authentication attempts",
	})

	AuthSuccesses = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "auth_successes_total",
		Help:      "Total number of successful authentications",
	})

	AuthFailures = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "auth_failures_total",
		Help:      "Total number of failed authentications",
	})

	TokensIssued = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "tokens_issued_total",
		Help:      "Total number of tokens issued",
	})

	TokensRevoked = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "tokens_revoked_total",
		Help:      "Total number of tokens revoked",
	})

	TokenRefreshes = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "token_refreshes_total",
		Help:      "Total number of token refresh operations",
	})

	PermissionChecks = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "permission_checks_total",
		Help:      "Total number of permission/authorization checks",
	})

	PermissionDenied = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "permission_denied_total",
		Help:      "Total number of permission denials",
	})

	SessionsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "sessions_created_total",
		Help:      "Total number of SSO sessions created",
	})

	SessionsRevoked = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "sessions_revoked_total",
		Help:      "Total number of SSO sessions revoked",
	})

	UsersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "users_created_total",
		Help:      "Total number of users created",
	})

	UsersDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "iam",
		Name:      "users_deleted_total",
		Help:      "Total number of users deleted",
	})
)

// ─────────────────────────────────────────────────────────────────────────────
// Gauge metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	ActiveSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "iam",
		Name:      "active_sessions",
		Help:      "Current number of active SSO sessions",
	})

	ActiveUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "iam",
		Name:      "active_users",
		Help:      "Current number of active users",
	})
)

// ─────────────────────────────────────────────────────────────────────────────
// Histogram metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	AuthDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "iam",
		Name:      "auth_duration_seconds",
		Help:      "Time taken to complete authentication",
		Buckets:   prometheus.DefBuckets,
	})

	TokenIssueDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "iam",
		Name:      "token_issue_duration_seconds",
		Help:      "Time taken to issue a token",
		Buckets:   prometheus.DefBuckets,
	})

	PermissionCheckDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "iam",
		Name:      "permission_check_duration_seconds",
		Help:      "Time taken to evaluate permissions",
		Buckets:   prometheus.DefBuckets,
	})
)
