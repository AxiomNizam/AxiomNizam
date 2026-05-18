package webauthn

// RegistrationRequest contains data for completing WebAuthn registration.
type RegistrationRequest struct {
	UserID    string `json:"user_id"`
	Response  []byte `json:"response"`
}

// RegistrationResponse contains the result of a successful registration.
type RegistrationResponse struct {
	CredentialID []byte `json:"credential_id"`
	PublicKey    []byte `json:"public_key"`
}
