package trusteddevices

import (
	"fmt"
	"net/http"
	"time"
)

const (
	// CookieName is the default cookie name for trusted device tokens.
	CookieName = "axiomnizam_device_token"
	// CookiePath is the cookie path.
	CookiePath = "/"
	// CookieSameSite is the SameSite policy.
	CookieSameSite = http.SameSiteStrictMode
)

// CookieConfig holds cookie configuration.
type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

// DefaultCookieConfig returns sensible defaults for device trust cookies.
func DefaultCookieConfig(ttlDays int) *CookieConfig {
	return &CookieConfig{
		Name:     CookieName,
		Path:     CookiePath,
		MaxAge:   ttlDays * 24 * 60 * 60,
		Secure:   true,
		HTTPOnly: true,
		SameSite: CookieSameSite,
	}
}

// SetCookie writes a trusted device cookie to the response.
func SetCookie(w http.ResponseWriter, cfg *CookieConfig, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    token,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		MaxAge:   cfg.MaxAge,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HTTPOnly,
		SameSite: cfg.SameSite,
	})
}

// GetCookie reads the trusted device cookie from the request.
func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearCookie removes the trusted device cookie.
func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     CookiePath,
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// CookieExpiresAt returns the expiration time for a cookie with the given TTL.
func CookieExpiresAt(ttlDays int) time.Time {
	return time.Now().UTC().AddDate(0, 0, ttlDays)
}

// BuildCookieValue constructs a cookie value from device ID and token.
func BuildCookieValue(deviceID, token string) string {
	return fmt.Sprintf("%s:%s", deviceID, token)
}
