package enrollment

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shafiunmiraz0/AxiomNizam/internal/gatekeeper/models"
	"github.com/shafiunmiraz0/AxiomNizam/internal/gatekeeper/repositories"
	"github.com/shafiunmiraz0/AxiomNizam/internal/gatekeeper/totp"
)

// Service manages the 2FA enrollment workflow.
type Service struct {
	factorRepo     repositories.FactorRepository
	backupCodeRepo repositories.BackupCodeRepository
	totpSvc        *totp.Service
	encryptionKey  []byte
}

// NewService creates a new enrollment service.
func NewService(
	fr repositories.FactorRepository,
	bcr repositories.BackupCodeRepository,
	ts *totp.Service,
	encKey []byte,
) *Service {
	return &Service{
		factorRepo:     fr,
		backupCodeRepo: bcr,
		totpSvc:        ts,
		encryptionKey:  encKey,
	}
}

// SetupRequest contains parameters for starting factor setup.
type SetupRequest struct {
	UserID     models.UserID
	FactorType models.FactorType
	Email      string // For email OTP
	Phone      string // For SMS OTP
	Issuer     string
}

// SetupResponse contains setup instructions for the user.
type SetupResponse struct {
	FactorID  models.FactorID
	Secret    string // Base32-encoded TOTP secret (for TOTP only)
	QRCodeURI string // otpauth:// URI for QR code generation
}

// SetupFactor initiates factor enrollment by creating a Pending factor and generating a secret.
func (s *Service) SetupFactor(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}

	if req.FactorType == "" {
		return nil, errors.New("factor_type is required")
	}

	// Generate secret for TOTP
	var secret, qrCodeURI string
	var err error

	if req.FactorType == models.FactorTypeTOTP {
		secret, qrCodeURI, err = s.totpSvc.GenerateSecret(ctx, req.UserID, req.UserID.String(), req.Issuer)
		if err != nil {
			return nil, err
		}
	}

	// Create factor in Pending phase
	factor := &models.Factor{
		ID:     uuid.New(),
		UserID: req.UserID,
		Spec: models.FactorSpec{
			Type:            req.FactorType,
			PhoneNumber:     req.Phone,
			Email:           req.Email,
			EncryptedSecret: []byte(secret), // TODO: Encrypt this
			Issuer:          req.Issuer,
		},
		Status: models.FactorStatus{
			Phase:      models.FactorPhasePending,
			Conditions: []models.Condition{},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	factor, err = s.factorRepo.Create(ctx, factor)
	if err != nil {
		return nil, err
	}

	return &SetupResponse{
		FactorID:  factor.ID,
		Secret:    secret,
		QRCodeURI: qrCodeURI,
	}, nil
}

// ActivateRequest contains parameters for completing factor activation.
type ActivateRequest struct {
	FactorID models.FactorID
	Code     string // OTP code to verify
}

// ActivateResponse contains confirmation and backup codes.
type ActivateResponse struct {
	FactorID    models.FactorID
	BackupCodes []string // Plaintext backup codes (shown once)
}

// ActivateFactor completes enrollment by verifying the OTP and transitioning to Active phase.
// Generates backup codes as fallback recovery mechanism.
func (s *Service) ActivateFactor(ctx context.Context, req *ActivateRequest) (*ActivateResponse, error) {
	// Retrieve the factor
	factor, err := s.factorRepo.Get(ctx, req.FactorID)
	if err != nil {
		return nil, err
	}
	if factor == nil {
		return nil, errors.New("factor not found")
	}

	if factor.Status.Phase != models.FactorPhasePending {
		return nil, errors.New("factor is not in Pending phase")
	}

	// Verify the OTP code
	var valid bool
	switch factor.Spec.Type {
	case models.FactorTypeTOTP:
		secret := string(factor.Spec.EncryptedSecret) // TODO: Decrypt
		valid, err = s.totpSvc.ValidateCode(ctx, secret, req.Code)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported factor type for OTP validation")
	}

	if !valid {
		return nil, errors.New("invalid OTP code")
	}

	// Generate backup codes
	plainCodes, err := s.totpSvc.GenerateRecoveryCodes(ctx, 10)
	if err != nil {
		return nil, err
	}

	// Hash and persist backup codes
	now := time.Now().UTC()
	backupCodes := make([]*models.BackupCode, len(plainCodes))
	for i, code := range plainCodes {
		backupCodes[i] = &models.BackupCode{
			ID:        uuid.New(),
			UserID:    factor.UserID,
			FactorID:  req.FactorID,
			CodeHash:  hashCode(code), // TODO: Use bcrypt or argon2
			CreatedAt: now,
		}
	}

	err = s.backupCodeRepo.Create(ctx, backupCodes)
	if err != nil {
		return nil, err
	}

	// Transition factor to Active phase
	factor.Status.Phase = models.FactorPhaseActive
	factor.Status.ActivatedAt = &now
	factor.Status.LastVerifiedAt = &now

	_, err = s.factorRepo.Update(ctx, factor)
	if err != nil {
		return nil, err
	}

	return &ActivateResponse{
		FactorID:    req.FactorID,
		BackupCodes: plainCodes,
	}, nil
}

// DisableFactor removes an active factor (user-initiated).
func (s *Service) DisableFactor(ctx context.Context, factorID models.FactorID) error {
	factor, err := s.factorRepo.Get(ctx, factorID)
	if err != nil {
		return err
	}
	if factor == nil {
		return errors.New("factor not found")
	}

	now := time.Now().UTC()
	factor.Status.Phase = models.FactorPhaseDisabled
	factor.Status.DisabledAt = &now

	// Clean up backup codes
	_ = s.backupCodeRepo.DeleteByFactorID(ctx, factorID)

	_, err = s.factorRepo.Update(ctx, factor)
	return err
}

// hashCode creates a simple hash of a backup code (TODO: use bcrypt or argon2).
func hashCode(code string) []byte {
	return []byte(base64.StdEncoding.EncodeToString([]byte(code)))
}
