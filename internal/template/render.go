// Package template wraps text/template with a sprig-like subset of
// helper functions used by consul-template / nomad-template style
// configuration rendering.  It covers the 90% case without pulling
// in the full sprig dependency tree.
//
// Provided functions:
//
//	env       - read an environment variable
//	default   - return second arg if first is zero/empty
//	required  - fail rendering if the value is empty
//	toYaml    - best-effort YAML-ish encoding (for simple maps)
//	toJson    - JSON encoding with indent
//	split     - strings.Split
//	join      - strings.Join
//	upper     - strings.ToUpper
//	lower     - strings.ToLower
//	trim      - strings.TrimSpace
//	replace   - strings.ReplaceAll
package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	ttmpl "text/template"
)

// FuncMap returns the standard function set.  Callers may extend it
// before passing to Render.
func FuncMap() ttmpl.FuncMap {
	return ttmpl.FuncMap{
		"env":      os.Getenv,
		"default":  tplDefault,
		"required": tplRequired,
		"toYaml":   tplToYaml,
		"toJson":   tplToJSON,
		"split":    func(sep, s string) []string { return strings.Split(s, sep) },
		"join":     func(sep string, xs []string) string { return strings.Join(xs, sep) },
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"trim":     strings.TrimSpace,
		"replace":  strings.ReplaceAll,
	}
}

// Render parses src and executes it against data.  funcs extends
// FuncMap; later entries override.
func Render(name, src string, data interface{}, funcs ttmpl.FuncMap) (string, error) {
	fm := FuncMap()
	for k, v := range funcs {
		fm[k] = v
	}
	t, err := ttmpl.New(name).Funcs(fm).Parse(src)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTo is the streaming variant — useful for very large outputs
// where buffering would be wasteful.
func RenderTo(w io.Writer, name, src string, data interface{}, funcs ttmpl.FuncMap) error {
	fm := FuncMap()
	for k, v := range funcs {
		fm[k] = v
	}
	t, err := ttmpl.New(name).Funcs(fm).Parse(src)
	if err != nil {
		return err
	}
	return t.Execute(w, data)
}

// tplDefault returns v if non-empty, else fallback.  Semantics match
// sprig: empty string, zero numeric, nil, and zero-length collections
// all count as "empty".
func tplDefault(fallback, v interface{}) interface{} {
	if isEmpty(v) {
		return fallback
	}
	return v
}

// tplRequired blocks rendering when the value is empty.
func tplRequired(msg string, v interface{}) (interface{}, error) {
	if isEmpty(v) {
		return nil, fmt.Errorf("template: required: %s", msg)
	}
	return v, nil
}

// isEmpty implements the sprig emptiness predicate.
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Bool:
		return !rv.Bool()
	}
	return false
}

// tplToJSON encodes v as indented JSON.
func tplToJSON(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// tplToYaml is a deliberately minimal YAML encoder — it handles
// scalar / slice / map[string]interface{} / struct-via-JSON.  Real
// YAML features (anchors, multi-doc, custom tags) are out of scope;
// call out to a real YAML library when they are needed.
func tplToYaml(v interface{}) (string, error) {
	var sb strings.Builder
	if err := yamlEncode(&sb, v, 0); err != nil {
		return "", err
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}

// yamlEncode is a naive recursive encoder.
func yamlEncode(sb *strings.Builder, v interface{}, depth int) error {
	indent := strings.Repeat("  ", depth)
	switch x := v.(type) {
	case nil:
		sb.WriteString("null\n")
	case string:
		fmt.Fprintf(sb, "%q\n", x)
	case bool, int, int64, float64:
		fmt.Fprintf(sb, "%v\n", x)
	case []interface{}:
		if len(x) == 0 {
			sb.WriteString("[]\n")
			return nil
		}
		sb.WriteByte('\n')
		for _, item := range x {
			sb.WriteString(indent)
			sb.WriteString("- ")
			if err := yamlEncode(sb, item, depth+1); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		if len(x) == 0 {
			sb.WriteString("{}\n")
			return nil
		}
		sb.WriteByte('\n')
		for k, val := range x {
			sb.WriteString(indent)
			fmt.Fprintf(sb, "%s: ", k)
			if err := yamlEncode(sb, val, depth+1); err != nil {
				return err
			}
		}
	default:
		// Fallback: JSON round-trip.
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		sb.Write(b)
		sb.WriteByte('\n')
	}
	return nil
}
