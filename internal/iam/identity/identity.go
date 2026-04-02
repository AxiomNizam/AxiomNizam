package identity

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents an IAM identity.
type User struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Email         string    `json:"email" gorm:"uniqueIndex;type:varchar(255);not null"`
	PasswordHash  string    `json:"-" gorm:"type:varchar(255);not null"` // never serialised
	DisplayName   string    `json:"display_name" gorm:"type:varchar(255)"`
	Roles         []string  `json:"roles" gorm:"-"` // managed via role-bindings, not this table
	Active        bool      `json:"active" gorm:"default:true"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CreateUserRequest is the payload for user registration.
type CreateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
}

// UpdateUserRequest allows partial updates.
type UpdateUserRequest struct {
	Email         *string `json:"email"`
	Password      *string `json:"password"`
	DisplayName   *string `json:"display_name"`
	Active        *bool   `json:"active"`
	EmailVerified *bool   `json:"email_verified"`
}

const (
	bcryptCost      = 12
	minPasswordLen  = 8
	maxPasswordLen  = 128
	recoveryCodeLen = 32 // bytes → 64 hex chars
)

// HashPassword securely hashes a plaintext password using bcrypt.
func HashPassword(plain string) (string, error) {
	if len(plain) < minPasswordLen {
		return "", errors.New("password must be at least 8 characters")
	}
	if len(plain) > maxPasswordLen {
		return "", errors.New("password too long")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// NewUserID generates a UUID for a new user.
func NewUserID() string {
	return uuid.New().String()
}

// GenerateRecoveryCode creates a cryptographically random recovery token.
func GenerateRecoveryCode() (string, error) {
	buf := make([]byte, recoveryCodeLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// NormaliseEmail lower-cases and trims an email address.
func NormaliseEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
