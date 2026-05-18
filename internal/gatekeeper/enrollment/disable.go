package enrollment

import "errors"

// ValidateDisableRequest checks the disable request for common errors.
func ValidateDisableRequest(factorID string) error {
	if factorID == "" {
		return errors.New("factor_id is required")
	}
	return nil
}
