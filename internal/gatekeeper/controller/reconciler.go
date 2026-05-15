package controller

import (
	"context"

	"yourapp/internal/gatekeeper/models"
	"yourapp/internal/gatekeeper/repositories"
)

type Reconciler struct {
	Factors repositories.FactorRepository
}

func (r *Reconciler) Reconcile(ctx context.Context, userID string) error {

	factor, err := r.Factors.GetByUser(ctx, userID)
	if err != nil {
		return err
	}

	// Desired state = ACTIVE MFA
	if factor.Status != models.StatusActive {

		// reconcile action
		factor.Status = models.StatusActive

		if factor.Secret == "" {
			factor.Secret = generateSecret()
		}

		return r.Factors.Update(ctx, factor)
	}

	return nil
}

func generateSecret() string {
	return "AUTO-GENERATED-SECRET"
}
