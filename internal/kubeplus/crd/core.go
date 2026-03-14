package crd

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type VersionSchema struct {
	Version string
	Fields  map[string]string
}

type Definition struct {
	Group    string
	Kind     string
	Plural   string
	Versions []VersionSchema
}

type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

type Registry struct {
	mu   sync.RWMutex
	defs map[string]Definition
}

func NewRegistry() *Registry {
	return &Registry{defs: make(map[string]Definition)}
}

func key(group, kind string) string {
	return strings.ToLower(strings.TrimSpace(group) + "/" + strings.TrimSpace(kind))
}

func (r *Registry) Register(def Definition) error {
	if strings.TrimSpace(def.Group) == "" || strings.TrimSpace(def.Kind) == "" {
		return fmt.Errorf("group and kind are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defs[key(def.Group, def.Kind)] = def
	return nil
}

func (r *Registry) Get(group, kind string) (Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.defs[key(group, kind)]
	return d, ok
}

func (r *Registry) List() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Definition, 0, len(r.defs))
	for _, d := range r.defs {
		out = append(out, d)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Group == out[j].Group {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Group < out[j].Group
	})
	return out
}

func ValidateSpec(fields map[string]string, spec map[string]any) ValidationResult {
	res := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}
	for k, typ := range fields {
		v, ok := spec[k]
		if !ok {
			res.Valid = false
			res.Errors = append(res.Errors, "missing required field: "+k)
			continue
		}
		switch typ {
		case "string":
			if _, ok := v.(string); !ok {
				res.Valid = false
				res.Errors = append(res.Errors, "field "+k+" must be string")
			}
		case "int":
			switch v.(type) {
			case int, int32, int64, float64:
			default:
				res.Valid = false
				res.Errors = append(res.Errors, "field "+k+" must be numeric")
			}
		}
	}
	return res
}
