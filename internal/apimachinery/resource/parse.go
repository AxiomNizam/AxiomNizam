// Package resource — Quantity parser.
//
// The parser is hand-written rather than driven by big.Rat.SetString
// because the latter does not understand the SI / binary suffixes
// that Kubernetes uses, and the whole point of Quantity is that "1Gi"
// is a valid literal.
package resource

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

// ParseQuantity is the inverse of Quantity.String.  It accepts:
//
//	123         → 123
//	123m        → 123 / 1000         (milli)
//	1.5k        → 1500               (decimal SI)
//	2Gi         → 2 * 2^30           (binary SI)
//	1e3         → 1000               (scientific notation)
//	-1.5M       → -1500000
//
// Whitespace is trimmed.  An empty string is not valid — upstream k8s
// returns zero for an empty string during YAML decoding, but callers
// should use the zero value of Quantity{} explicitly.
func ParseQuantity(s string) (Quantity, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Quantity{}, fmt.Errorf("empty quantity string")
	}

	numStr, suffix := splitSuffix(s)
	if numStr == "" {
		return Quantity{}, fmt.Errorf("quantity %q has no numeric part", s)
	}

	// Scientific notation: "1e3", "1.5E-2".  The exponent forms part
	// of the number, not the suffix.
	if eIdx := strings.IndexAny(numStr, "eE"); eIdx >= 0 {
		return parseScientific(numStr, suffix)
	}

	rat, ok := new(big.Rat).SetString(numStr)
	if !ok {
		return Quantity{}, fmt.Errorf("invalid numeric part %q", numStr)
	}

	mult, format, err := suffixMultiplier(suffix)
	if err != nil {
		return Quantity{}, err
	}
	rat.Mul(rat, mult)
	return Quantity{value: rat, Format: format}, nil
}

// splitSuffix separates the numeric prefix from the SI / binary suffix.
// The numeric prefix may contain a sign, digits, a decimal point, and
// an optional exponent (e/E).  Everything after is the suffix.
func splitSuffix(s string) (num, suffix string) {
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	// Consume digits, dot, and scientific-notation e/E+digits.
	for i < len(s) {
		c := rune(s[i])
		if unicode.IsDigit(c) || c == '.' {
			i++
			continue
		}
		if (c == 'e' || c == 'E') && i+1 < len(s) {
			// e/E only starts an exponent if followed by digits or sign.
			next := rune(s[i+1])
			if unicode.IsDigit(next) || next == '+' || next == '-' {
				i++
				if next == '+' || next == '-' {
					i++
				}
				continue
			}
		}
		break
	}
	return s[:i], s[i:]
}

// parseScientific handles literals such as "1.5e3Ki" — rare but valid.
func parseScientific(numStr, suffix string) (Quantity, error) {
	rat, ok := new(big.Rat).SetString(numStr)
	if !ok {
		return Quantity{}, fmt.Errorf("invalid scientific literal %q", numStr)
	}
	mult, format, err := suffixMultiplier(suffix)
	if err != nil {
		return Quantity{}, err
	}
	if suffix == "" {
		format = DecimalExponent
	}
	rat.Mul(rat, mult)
	return Quantity{value: rat, Format: format}, nil
}

// suffixMultiplier maps a textual suffix to its numeric multiplier.
// Unrecognised suffixes are a parse error; the empty suffix yields 1.
func suffixMultiplier(suffix string) (*big.Rat, Format, error) {
	switch suffix {
	case "":
		return big.NewRat(1, 1), DecimalSI, nil
	case "m":
		return big.NewRat(1, 1000), DecimalSI, nil
	case "k":
		return big.NewRat(1_000, 1), DecimalSI, nil
	case "M":
		return big.NewRat(1_000_000, 1), DecimalSI, nil
	case "G":
		return big.NewRat(1_000_000_000, 1), DecimalSI, nil
	case "T":
		return big.NewRat(1_000_000_000_000, 1), DecimalSI, nil
	case "P":
		return big.NewRat(1_000_000_000_000_000, 1), DecimalSI, nil
	case "E":
		return big.NewRat(1_000_000_000_000_000_000, 1), DecimalSI, nil
	case "Ki":
		return big.NewRat(1<<10, 1), BinarySI, nil
	case "Mi":
		return big.NewRat(1<<20, 1), BinarySI, nil
	case "Gi":
		return big.NewRat(1<<30, 1), BinarySI, nil
	case "Ti":
		return big.NewRat(1<<40, 1), BinarySI, nil
	case "Pi":
		return big.NewRat(1<<50, 1), BinarySI, nil
	case "Ei":
		return big.NewRat(1<<60, 1), BinarySI, nil
	}
	return nil, "", fmt.Errorf("unrecognised quantity suffix %q", suffix)
}

// formatQuantity is the inverse of ParseQuantity.  For round-trip
// fidelity with the original input it would need to remember the
// exact suffix used; instead we pick the smallest suffix that renders
// the value as an integer (or fall back to a decimal literal).
func formatQuantity(v *big.Rat, f Format) string {
	if v.Sign() == 0 {
		return "0"
	}
	switch f {
	case BinarySI:
		for _, try := range []struct {
			suffix string
			mult   *big.Rat
		}{
			{"Ei", big.NewRat(1<<60, 1)},
			{"Pi", big.NewRat(1<<50, 1)},
			{"Ti", big.NewRat(1<<40, 1)},
			{"Gi", big.NewRat(1<<30, 1)},
			{"Mi", big.NewRat(1<<20, 1)},
			{"Ki", big.NewRat(1<<10, 1)},
		} {
			q := new(big.Rat).Quo(v, try.mult)
			if q.IsInt() {
				return q.RatString() + try.suffix
			}
		}
	case DecimalSI:
		for _, try := range []struct {
			suffix string
			mult   *big.Rat
		}{
			{"E", big.NewRat(1_000_000_000_000_000_000, 1)},
			{"P", big.NewRat(1_000_000_000_000_000, 1)},
			{"T", big.NewRat(1_000_000_000_000, 1)},
			{"G", big.NewRat(1_000_000_000, 1)},
			{"M", big.NewRat(1_000_000, 1)},
			{"k", big.NewRat(1_000, 1)},
		} {
			q := new(big.Rat).Quo(v, try.mult)
			if q.IsInt() {
				return q.RatString() + try.suffix
			}
		}
		// milli fallback for sub-unit values (e.g. CPU "100m").
		milli := new(big.Rat).Mul(v, big.NewRat(1000, 1))
		if milli.IsInt() {
			return milli.RatString() + "m"
		}
	}
	// Fallback: render as a plain decimal with up to 12 fractional digits.
	return strings.TrimRight(strings.TrimRight(v.FloatString(12), "0"), ".")
}
