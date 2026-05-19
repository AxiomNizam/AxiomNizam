package enrollment

import "errors"

// ValidateActivateRequest checks the activation request for common errors.
func ValidateActivateRequest(req *ActivateRequest) error {
	if req.FactorID.String() == "00000000-0000-0000-0000-000000000000" {
		return errors.New("factor_id is required")
	}
	if req.Code == "" {
		return errors.New("code is required")
	}
	if len(req.Code) < 4 || len(req.Code) > 8 {
		return errors.New("code must be between 4 and 8 digits")
	}
	return nil
}
