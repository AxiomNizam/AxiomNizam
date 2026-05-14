// Package intstr provides the IntOrString type — a field that may be
// expressed as either an integer or a string in JSON/YAML.  k8s uses
// it for scale targets ("replicas: 5" or "replicas: 50%"), pod
// disruption budgets, and service ports ("port: 80" or "port: http").
package intstr

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Type discriminates which half of the union carries the value.
type Type int

const (
	// Int means IntVal is authoritative.
	Int Type = iota
	// String means StrVal is authoritative.
	String
)

// IntOrString is the union type.  Zero value is the integer 0.
type IntOrString struct {
	Type   Type
	IntVal int32
	StrVal string
}

// FromInt constructs the int variant.
func FromInt(v int) IntOrString { return IntOrString{Type: Int, IntVal: int32(v)} }

// FromString constructs the string variant.
func FromString(v string) IntOrString { return IntOrString{Type: String, StrVal: v} }

// IntValue returns the numeric value.  String values that happen to
// parse as ints are returned as their numeric equivalent; percentage
// values are stripped of the trailing "%" and parsed.
func (v *IntOrString) IntValue() int {
	if v.Type == Int {
		return int(v.IntVal)
	}
	s := strings.TrimSuffix(v.StrVal, "%")
	i, _ := strconv.Atoi(s)
	return i
}

// String renders a textual form for logs.
func (v *IntOrString) String() string {
	if v == nil {
		return "<nil>"
	}
	if v.Type == String {
		return v.StrVal
	}
	return strconv.Itoa(int(v.IntVal))
}

// MarshalJSON emits the active half as its native JSON form.
func (v IntOrString) MarshalJSON() ([]byte, error) {
	switch v.Type {
	case Int:
		return []byte(strconv.Itoa(int(v.IntVal))), nil
	case String:
		return json.Marshal(v.StrVal)
	}
	return nil, fmt.Errorf("IntOrString has unknown Type %d", v.Type)
}

// UnmarshalJSON sniffs the first non-space byte: `"` → string, else int.
func (v *IntOrString) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) == 0 {
		return fmt.Errorf("empty IntOrString")
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		v.Type = String
		v.StrVal = s
		return nil
	}
	var i int32
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	v.Type = Int
	v.IntVal = i
	return nil
}

// GetValueFromPercent returns the rounded-down result of
// (percentage * total / 100) for IntOrString values like "25%".
// For Int values it simply returns the integer.  This is the exact
// arithmetic the k8s scheduler uses for PodDisruptionBudget.
func GetValueFromPercent(v IntOrString, total int) (int, error) {
	if v.Type == Int {
		return int(v.IntVal), nil
	}
	s := strings.TrimSpace(v.StrVal)
	if !strings.HasSuffix(s, "%") {
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("IntOrString %q is neither int nor percentage", v.StrVal)
		}
		return n, nil
	}
	pct, err := strconv.Atoi(strings.TrimSuffix(s, "%"))
	if err != nil {
		return 0, fmt.Errorf("invalid percentage %q: %w", v.StrVal, err)
	}
	return pct * total / 100, nil
}
