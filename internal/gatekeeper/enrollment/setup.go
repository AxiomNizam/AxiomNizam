package enrollment

import "errors"

// ValidateSetupRequest checks the setup request for common errors.
func ValidateSetupRequest(req *SetupRequest) error {
	if req.UserID.String() == "00000000-0000-0000-0000-000000000000" {
		return errors.New("user_id is required")
	}
	if req.FactorType == "" {
		return errors.New("factor_type is required")
	}
	if req.Issuer == "" {
		req.Issuer = "AxiomNizam"
	}
	return nil
}
