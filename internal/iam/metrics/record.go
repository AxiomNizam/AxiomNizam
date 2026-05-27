package metrics

import "time"

// RecordAuthAttempt records an authentication attempt.
func RecordAuthAttempt() { AuthAttempts.Inc() }

// RecordAuthSuccess records a successful authentication.
func RecordAuthSuccess() { AuthSuccesses.Inc() }

// RecordAuthFailure records a failed authentication.
func RecordAuthFailure() { AuthFailures.Inc() }

// RecordTokenIssued records a token issuance.
func RecordTokenIssued() { TokensIssued.Inc() }

// RecordTokenRevoked records a token revocation.
func RecordTokenRevoked() { TokensRevoked.Inc() }

// RecordTokenRefresh records a token refresh.
func RecordTokenRefresh() { TokenRefreshes.Inc() }

// RecordPermissionCheck records a permission check.
func RecordPermissionCheck() { PermissionChecks.Inc() }

// RecordPermissionDenied records a permission denial.
func RecordPermissionDenied() { PermissionDenied.Inc() }

// RecordSessionCreated records a session creation.
func RecordSessionCreated() { SessionsCreated.Inc() }

// RecordSessionRevoked records a session revocation.
func RecordSessionRevoked() { SessionsRevoked.Inc() }

// RecordUserCreated records a user creation.
func RecordUserCreated() { UsersCreated.Inc() }

// RecordUserDeleted records a user deletion.
func RecordUserDeleted() { UsersDeleted.Inc() }

// SetActiveSessions sets the current active session count.
func SetActiveSessions(count float64) { ActiveSessions.Set(count) }

// SetActiveUsers sets the current active user count.
func SetActiveUsers(count float64) { ActiveUsers.Set(count) }

// RecordAuthDuration records authentication latency.
func RecordAuthDuration(d time.Duration) { AuthDuration.Observe(d.Seconds()) }

// RecordTokenIssueDuration records token issuance latency.
func RecordTokenIssueDuration(d time.Duration) { TokenIssueDuration.Observe(d.Seconds()) }

// RecordPermissionCheckDuration records permission check latency.
func RecordPermissionCheckDuration(d time.Duration) { PermissionCheckDuration.Observe(d.Seconds()) }
