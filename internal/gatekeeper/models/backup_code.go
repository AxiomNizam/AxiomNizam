package models

import (
	"time"

	"github.com/google/uuid"
)

// BackupCode is a one-time recovery credential.
// Backup codes are generated together as a set during enrollment.
// Each code can be used exactly once if MFA verification fails.
type BackupCode struct {
	ID        uuid.UUID  `db:"id"        json:"id"`
	UserID    UserID     `db:"user_id"   json:"user_id"`
	FactorID  FactorID   `db:"factor_id" json:"factor_id"`
	CodeHash  []byte     `db:"code_hash" json:"-"` // Argon2id / bcrypt hash
	UsedAt    *time.Time `db:"used_at"   json:"used_at,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// IsUsed returns true if the backup code has been consumed.
func (b *BackupCode) IsUsed() bool {
	return b.UsedAt != nil
}

// MarshalJSON is intentionally omitted — struct tags handle field exclusion:
// - CodeHash uses json:"-" to exclude from JSON output
// - All other fields use explicit json tags
