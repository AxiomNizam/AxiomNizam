package webauthn

// AuthenticationRequest contains data for completing WebAuthn authentication.
type AuthenticationRequest struct {
	UserID   string `json:"user_id"`
	Response []byte `json:"response"`
}

// AuthenticationResponse contains the result of a successful authentication.
type AuthenticationResponse struct {
	Verified    bool   `json:"verified"`
	CredentialID []byte `json:"credential_id"`
}
