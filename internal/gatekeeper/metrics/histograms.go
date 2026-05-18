package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ChallengeDuration tracks time to complete challenge verification.
var ChallengeDuration = promauto.NewHistogram(prometheus.HistogramOpts{
	Namespace: "mfa",
	Name:      "challenge_duration_seconds",
	Help:      "Time taken to complete an MFA challenge (begin to verify)",
	Buckets:   prometheus.DefBuckets,
})

// BackupCodeRegenerationDuration tracks time to regenerate backup codes.
var BackupCodeRegenerationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
	Namespace: "mfa",
	Name:      "backup_code_regen_duration_seconds",
	Help:      "Time taken to regenerate backup codes",
	Buckets:   prometheus.DefBuckets,
})

// RiskScoringDuration tracks time to calculate risk scores.
var RiskScoringDuration = promauto.NewHistogram(prometheus.HistogramOpts{
	Namespace: "mfa",
	Name:      "risk_scoring_duration_seconds",
	Help:      "Time taken to calculate risk scores",
	Buckets:   prometheus.DefBuckets,
})

// RecordChallengeDuration records a challenge completion time.
func RecordChallengeDuration(d time.Duration) {
	ChallengeDuration.Observe(d.Seconds())
}

// RecordBackupCodeRegenerationDuration records backup code regeneration time.
func RecordBackupCodeRegenerationDuration(d time.Duration) {
	BackupCodeRegenerationDuration.Observe(d.Seconds())
}

// RecordRiskScoringDuration records risk scoring time.
func RecordRiskScoringDuration(d time.Duration) {
	RiskScoringDuration.Observe(d.Seconds())
}
