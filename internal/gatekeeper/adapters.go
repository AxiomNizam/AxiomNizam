package gatekeeper

import (
	"context"

	"github.com/google/uuid"
	"example.com/axiomnizam/internal/gatekeeper/contracts"
	"example.com/axiomnizam/internal/gatekeeper/enrollment"
	"example.com/axiomnizam/internal/gatekeeper/challenge"
	"example.com/axiomnizam/internal/gatekeeper/models"
	"example.com/axiomnizam/internal/gatekeeper/repositories"
)

// enrollmentServiceWrapper wraps enrollment.Service to match contracts.EnrollmentService.
type enrollmentServiceWrapper struct {
	svc *enrollment.Service
}

func wrapEnrollmentService(svc *enrollment.Service) contracts.EnrollmentService {
	return &enrollmentServiceWrapper{svc: svc}
}

func (w *enrollmentServiceWrapper) SetupFactor(ctx context.Context, userID models.UserID, factorType models.FactorType) (string, error) {
	resp, err := w.svc.SetupFactor(ctx, &enrollment.SetupRequest{
		UserID:     userID,
		FactorType: factorType,
		Issuer:     "AxiomNizam",
	})
	if err != nil {
		return "", err
	}
	return resp.Secret, nil
}

func (w *enrollmentServiceWrapper) ActivateFactor(ctx context.Context, factorID models.FactorID, code string) ([]string, error) {
	resp, err := w.svc.ActivateFactor(ctx, &enrollment.ActivateRequest{
		FactorID: factorID,
		Code:     code,
	})
	if err != nil {
		return nil, err
	}
	return resp.BackupCodes, nil
}

func (w *enrollmentServiceWrapper) DisableFactor(ctx context.Context, factorID models.FactorID) error {
	return w.svc.DisableFactor(ctx, factorID)
}

// challengeServiceWrapper wraps challenge.Service to match contracts.ChallengeService.
type challengeServiceWrapper struct {
	svc *challenge.Service
}

func wrapChallengeService(svc *challenge.Service) contracts.ChallengeService {
	return &challengeServiceWrapper{svc: svc}
}

func (w *challengeServiceWrapper) BeginChallenge(ctx context.Context, userID models.UserID, factorID models.FactorID) (string, error) {
	challenge, err := w.svc.BeginChallenge(ctx, &challenge.BeginRequest{
		UserID:   userID,
		FactorID: factorID,
	})
	if err != nil {
		return "", err
	}
	return challenge.ID.String(), nil
}

func (w *challengeServiceWrapper) VerifyChallenge(ctx context.Context, challengeID string, code string) (bool, error) {
	id, err := uuid.Parse(challengeID)
	if err != nil {
		return false, err
	}
	challenge, err := w.svc.VerifyChallenge(ctx, &challenge.VerifyRequest{
		ChallengeID: id,
		Code:        code,
	})
	if err != nil {
		return false, err
	}
	return challenge.Phase == models.ChallengePhaseVerified, nil
}

func (w *challengeServiceWrapper) ExpireChallenge(ctx context.Context, challengeID string) error {
	id, err := uuid.Parse(challengeID)
	if err != nil {
		return err
	}
	return w.svc.ExpireChallenge(ctx, id)
}

// factorServiceWrapper wraps FactorRepository to match contracts.FactorService.
type factorServiceWrapper struct {
	repo repositories.FactorRepository
}

func wrapFactorService(repo repositories.FactorRepository) contracts.FactorService {
	return &factorServiceWrapper{repo: repo}
}

func (w *factorServiceWrapper) GetFactor(ctx context.Context, factorID models.FactorID) (*models.Factor, error) {
	return w.repo.Get(ctx, factorID)
}

func (w *factorServiceWrapper) ListFactors(ctx context.Context, userID models.UserID) ([]*models.Factor, error) {
	return w.repo.GetByUserID(ctx, userID)
}

func (w *factorServiceWrapper) DeleteFactor(ctx context.Context, factorID models.FactorID) error {
	return w.repo.Delete(ctx, factorID)
}

func (w *factorServiceWrapper) GetActiveFactorCount(ctx context.Context, userID models.UserID) (int, error) {
	factors, err := w.repo.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, f := range factors {
		if f.IsActive() {
			count++
		}
	}
	return count, nil
}