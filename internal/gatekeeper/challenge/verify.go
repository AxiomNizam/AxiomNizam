package challenge

import "errors"

// ValidateVerifyRequest checks the verify request for common errors.
func ValidateVerifyRequest(challengeID, code string) error {
	if challengeID == "" {
		return errors.New("challenge_id is required")
	}
	if code == "" {
		return errors.New("code is required")
	}
	if len(code) < 4 || len(code) > 8 {
		return errors.New("code must be between 4 and 8 digits")
	}
	return nil
}
