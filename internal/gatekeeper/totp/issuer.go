package totp

// IssuerProvider builds the otpauth:// URI for TOTP provisioning.
type IssuerProvider interface {
	// BuildOTPAuthURI creates the QR code content URI.
	BuildOTPAuthURI(secret, accountName, issuer string) string
}

// DefaultIssuer provides standard TOTP URI building.
type DefaultIssuer struct{}

// BuildOTPAuthURI creates the otpauth:// TOTP URI.
func (d *DefaultIssuer) BuildOTPAuthURI(secret, accountName, issuer string) string {
	return "otpauth://totp/" + issuer + ":" + accountName + "?secret=" + secret + "&issuer=" + issuer + "&algorithm=SHA1&digits=6&period=30"
}

// NewIssuerProvider returns the default issuer.
func NewIssuerProvider() IssuerProvider {
	return &DefaultIssuer{}
}