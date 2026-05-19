package challenge

import "errors"

// ValidateBeginRequest checks the begin request for common errors.
func ValidateBeginRequest(userID, factorID string) error {
	if userID == "" {
		return errors.New("user_id is required")
	}
	if factorID == "" {
		return errors.New("factor_id is required")
	}
	return nil
}
