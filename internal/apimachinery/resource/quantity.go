// Package resource implements the Quantity type: a serializable
// fixed-point number with a unit suffix, used throughout Kubernetes
// to express CPU, memory, storage, and other resource requests.
//
// The canonical grammar (from the k8s api/resource/quantity.go) is:
//
//	<quantity>  ::= <signedNumber><suffix>
//	<suffix>    ::= <binarySI> | <decimalSI> | <decimalExponent> | ""
//	<binarySI>  ::= Ki | Mi | Gi | Ti | Pi | Ei
//	<decimalSI> ::= "" | m | "" | k | M | G | T | P | E
//	<decimalExponent> ::= "e" <signedNumber> | "E" <signedNumber>
//
// Examples: "100m" (100 milli-units), "1Gi" (2^30), "2.5G" (2.5*10^9),
// "1e3" (1000).  The parser is tolerant of whitespace around the number.
//
// This implementation is not bit-for-bit identical to upstream but is
// sufficient for AxiomNizam's resource-request workflows: quotas,
// limits, autoscaler thresholds.
package resource

import (
	"fmt"
	"math/big"
	"strings"
)

// Format controls the preferred string rendering of a Quantity.
type Format string

const (
	// DecimalExponent is "<num>e<exp>"  — useful for very large values.
	DecimalExponent Format = "DecimalExponent"
	// BinarySI is Ki/Mi/Gi/Ti/Pi/Ei — memory-style.
	BinarySI Format = "BinarySI"
	// DecimalSI is k/M/G/T/P/E or m — CPU/SI-style.
	DecimalSI Format = "DecimalSI"
)

// Quantity is the fixed-point value.  The internal representation
// uses math/big.Rat so that a "100m" CPU request plus a "50m" request
// round-trips exactly.  Format is remembered so MarshalJSON preserves
// the caller's stylistic intent.
type Quantity struct {
	value  *big.Rat
	Format Format
}

// NewQuantity constructs a Quantity representing value * 10^0 in the
// requested format.
func NewQuantity(value int64, format Format) Quantity {
	return Quantity{value: big.NewRat(value, 1), Format: format}
}

// NewMilliQuantity constructs value * 10^-3, matching client-go helper.
func NewMilliQuantity(value int64, format Format) Quantity {
	return Quantity{value: big.NewRat(value, 1000), Format: format}
}

// Sign returns -1, 0, or 1.
func (q Quantity) Sign() int {
	if q.value == nil {
		return 0
	}
	return q.value.Sign()
}

// Value returns the floor of the rational value.  Fractional units are
// truncated towards zero — mirrors the upstream behaviour.
func (q Quantity) Value() int64 {
	if q.value == nil {
		return 0
	}
	f, _ := q.value.Float64()
	return int64(f)
}

// MilliValue returns the value in thousandths of a unit.
func (q Quantity) MilliValue() int64 {
	if q.value == nil {
		return 0
	}
	milli := new(big.Rat).Mul(q.value, big.NewRat(1000, 1))
	f, _ := milli.Float64()
	return int64(f)
}

// Cmp returns -1, 0, or 1 for q < other, q == other, q > other.
func (q Quantity) Cmp(other Quantity) int {
	if q.value == nil && other.value == nil {
		return 0
	}
	if q.value == nil {
		return -other.value.Sign()
	}
	if other.value == nil {
		return q.value.Sign()
	}
	return q.value.Cmp(other.value)
}

// Add mutates q to q+other and returns q for chaining.
func (q *Quantity) Add(other Quantity) *Quantity {
	if q.value == nil {
		q.value = new(big.Rat)
	}
	if other.value == nil {
		return q
	}
	q.value.Add(q.value, other.value)
	return q
}

// Sub mutates q to q-other and returns q for chaining.
func (q *Quantity) Sub(other Quantity) *Quantity {
	if q.value == nil {
		q.value = new(big.Rat)
	}
	if other.value == nil {
		return q
	}
	q.value.Sub(q.value, other.value)
	return q
}

// String renders the quantity using the remembered Format.
func (q Quantity) String() string {
	if q.value == nil {
		return "0"
	}
	return formatQuantity(q.value, q.Format)
}

// MarshalJSON produces a JSON string ("2Gi").  Quantities serialise
// as strings so numeric precision is preserved across YAML/JSON round
// trips — matching the k8s API convention.
func (q Quantity) MarshalJSON() ([]byte, error) {
	s := q.String()
	// Quoted JSON string, no escaping needed for SI suffixes.
	return []byte("\"" + s + "\""), nil
}

// UnmarshalJSON parses either a quoted string or a bare number.
func (q *Quantity) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	s = strings.Trim(s, "\"")
	parsed, err := ParseQuantity(s)
	if err != nil {
		return err
	}
	*q = parsed
	return nil
}

// MustParse is the panicking variant of ParseQuantity.  Use only for
// constants in tests or package-init blocks.
func MustParse(s string) Quantity {
	q, err := ParseQuantity(s)
	if err != nil {
		panic(fmt.Sprintf("MustParse(%q): %v", s, err))
	}
	return q
}
