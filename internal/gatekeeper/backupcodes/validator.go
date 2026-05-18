package backupcodes

import (
	"regexp"
	"strings"
)

var (
	// codePattern matches 8-digit codes with optional dash separator.
	codePattern = regexp.MustCompile(`^[0-9]{4}-?[0-9]{4}$`)
)

// CodeValidator validates backup code format.
type CodeValidator struct{}

// IsValid checks if a backup code has the correct format.
func (v *CodeValidator) IsValid(code string) bool {
	if code == "" {
		return false
	}
	code = strings.TrimSpace(code)
	return codePattern.MatchString(code)
}

// Normalize removes dashes and whitespace from a backup code.
func Normalize(code string) string {
	code = strings.TrimSpace(code)
	code = strings.ReplaceAll(code, "-", "")
	return strings.ToLower(code)
}

// FormatDisplay adds dashes for display (XXXX-XXXX).
func FormatDisplay(code string) string {
	code = Normalize(code)
	if len(code) != CodeLength {
		return code
	}
	return code[:4] + "-" + code[4:]
}
