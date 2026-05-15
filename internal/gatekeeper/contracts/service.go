package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/org/project/internal/twofactor/models"
)

// ─── Enrollment ───────────────────────────────────────────────────────────────

// EnrollmentService manages the Factor lifecycle: enroll → activate → disable.
type EnrollmentService interface {
	// Enroll creates a Factor in Pending phase and returns a provisioning URI
	// (otpauth://) for the authenticator app to scan.
	Enroll(ctx context.Context, req EnrollRequest) (*EnrollResult, error)

	// Activate transitions a Pending factor to Active after the user proves
	// possession by submitting a valid OTP.
	Activate(ctx context.Context, req ActivateRequest) error

	// Disable soft-disables an Active factor.  The factor is preserved for audit.
	Disable(ctx context.Context, userID, factorID uuid.UUID) error

	// Status returns the current phase and conditions for a factor.
	Status(ctx context.Context, userID, factorID uuid.UUID) (*models.Factor, error)

	// List returns all factors for a user.
	List(ctx context.Context, userID uuid.UUID) ([]*models.Factor, error)
}

// EnrollRequest carries caller input for a new factor.
type EnrollRequest struct {
	UserID      uuid.UUID
	Type        models.FactorType
	Issuer      string
	AccountName string // shown in the authenticator app (usually email)
	PhoneNumber string // SMS only
	Email       string // email-OTP only
}

// EnrollResult carries the provisioning data returned to the caller.
type EnrollResult struct {
	Factor          *models.Factor
	ProvisioningURI string   // otpauth://
	QRCodeSVG       string   // base64-encoded SVG for QR display
	BackupCodes     []string // plaintext — shown once, then hashed
}

// ActivateRequest carries the OTP the user submits to prove possession.
type ActivateRequest struct {
	UserID   uuid.UUID
	FactorID uuid.UUID
	OTP      string
}

// ─── Challenge ────────────────────────────────────────────────────────────────

// ChallengeService orchestrates the runtime "prove it" flow.
type ChallengeService interface {
	// Begin creates a new Challenge for the user's active factor.
	Begin(ctx context.Context, req BeginRequest) (*models.Challenge, error)

	// Verify submits an OTP against an open challenge.
	Verify(ctx context.Context, req VerifyRequest) (*VerifyResult, error)

	// Get returns a challenge by ID.
	Get(ctx context.Context, challengeID uuid.UUID) (*models.Challenge, error)
}

// BeginRequest carries metadata for creating a challenge.
type BeginRequest struct {
	UserID    uuid.UUID
	FactorID  uuid.UUID
	IPAddress string
	UserAgent string
	TTL       time.Duration // defaults to config.ChallengeTTL if zero
}

// VerifyRequest carries the user's OTP submission.
type VerifyRequest struct {
	ChallengeID uuid.UUID
	UserID      uuid.UUID
	OTP         string
	IPAddress   string
}

// VerifyResult indicates the outcome of a verification attempt.
type VerifyResult struct {
	Success   bool
	Challenge *models.Challenge
	// SessionToken is minted on success; empty on failure.
	SessionToken string
}

// ─── TOTP Provider ────────────────────────────────────────────────────────────

// TOTPProvider is the low-level TOTP primitive used by enrollment and challenge.
type TOTPProvider interface {
	GenerateSecret() (secret string, err error)
	ProvisioningURI(secret, issuer, account string) string
	Validate(secret, otp string, at time.Time) (bool, error)
	GenerateQRCode(uri string) (svgData string, err error)
}

// ─── Reconciler ───────────────────────────────────────────────────────────────

// Reconciler is the K8s-style controller loop interface.
// Each subsystem (factor, challenge, device) implements its own Reconciler.
type Reconciler interface {
	Reconcile(ctx context.Context, req ReconcileRequest) (ReconcileResult, error)
}

// ReconcileRequest carries the object identity to reconcile.
type ReconcileRequest struct {
	// ID is the primary key of the object being reconciled.
	ID uuid.UUID
	// Generation is the object's current generation counter.
	Generation int64
}

// ReconcileResult instructs the controller loop what to do next.
type ReconcileResult struct {
	// Requeue immediately.
	Requeue bool
	// RequeueAfter schedules a re-enqueue after a delay (e.g. TTL expiry).
	RequeueAfter time.Duration
}

// ─── Raft Store ───────────────────────────────────────────────────────────────

// RaftStore is the distributed state machine interface.
// Only the Raft leader may apply commands; followers serve reads from local state.
type RaftStore interface {
	Apply(cmd RaftCommand) error
	IsLeader() bool
	LeaderAddr() string
	// ReadFactor returns the in-memory factor state (may lag by one Raft tick).
	ReadFactor(id uuid.UUID) (*models.Factor, bool)
	// ReadChallenge returns the in-memory challenge state.
	ReadChallenge(id uuid.UUID) (*models.Challenge, bool)
}

// RaftCommand is the opaque payload serialised into the Raft log entry.
type RaftCommand struct {
	Type    models.RaftCommandType `json:"type"`
	Payload []byte                 `json:"payload"` // msgpack-encoded per-command struct
}
