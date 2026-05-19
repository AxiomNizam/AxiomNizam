package totp

import (
	"fmt"
	"net/url"
)

// QRCodeData contains the data needed to generate a QR code.
type QRCodeData struct {
	URI     string // otpauth:// URI
	Secret  string // Base32-encoded secret
	Issuer  string
	Account string
}

// BuildOTPAuthURI constructs an otpauth:// URI for QR code generation.
func BuildOTPAuthURI(secret, accountName, issuer string) string {
	params := url.Values{}
	params.Set("secret", secret)
	params.Set("issuer", issuer)
	params.Set("algorithm", "SHA1")
	params.Set("digits", fmt.Sprintf("%d", Digits))
	params.Set("period", fmt.Sprintf("%d", TimeStepSeconds))

	return fmt.Sprintf("otpauth://totp/%s:%s?%s",
		url.PathEscape(issuer),
		url.PathEscape(accountName),
		params.Encode(),
	)
}

// ParseOTPAuthURI parses an otpauth:// URI into its components.
func ParseOTPAuthURI(uri string) (*QRCodeData, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}
	if parsed.Scheme != "otpauth" {
		return nil, fmt.Errorf("expected otpauth scheme, got %s", parsed.Scheme)
	}
	if parsed.Host != "totp" {
		return nil, fmt.Errorf("expected totp host, got %s", parsed.Host)
	}

	params := parsed.Query()
	return &QRCodeData{
		URI:     uri,
		Secret:  params.Get("secret"),
		Issuer:  params.Get("issuer"),
		Account: parsed.Path,
	}, nil
}
