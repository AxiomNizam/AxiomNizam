package admission

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type ObjectRef struct {
	Namespace string
	Name      string
	Kind      string
}

type AdmissionRequest struct {
	Object      ObjectRef
	Kind        string
	Operation   string
	User        string
	TenantID    string
	Labels      map[string]string
	Annotations map[string]string
	Spec        map[string]any
	Timestamp   time.Time
}

type AdmissionDecision struct {
	Allowed   bool
	Severity  string
	Reason    string
	Mutations map[string]string
}

type PolicyFunc func(AdmissionRequest) AdmissionDecision

type policyEntry struct {
	name     string
	priority int
	eval     PolicyFunc
}

type Engine struct {
	mu       sync.RWMutex
	policies []policyEntry
}

func NewEngine() *Engine {
	return &Engine{policies: make([]policyEntry, 0, 32)}
}

func (e *Engine) RegisterPolicy(name string, priority int, fn PolicyFunc) {
	if strings.TrimSpace(name) == "" || fn == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies = append(e.policies, policyEntry{name: name, priority: priority, eval: fn})
	sort.SliceStable(e.policies, func(i, j int) bool { return e.policies[i].priority > e.policies[j].priority })
}

func (e *Engine) ListPolicies() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]string, 0, len(e.policies))
	for _, p := range e.policies {
		out = append(out, p.name)
	}
	return out
}

func (e *Engine) Evaluate(req AdmissionRequest) AdmissionDecision {
	e.mu.RLock()
	policies := append([]policyEntry(nil), e.policies...)
	e.mu.RUnlock()

	merged := AdmissionDecision{Allowed: true, Severity: "info", Mutations: map[string]string{}}
	for _, p := range policies {
		d := p.eval(req)
		if len(d.Mutations) > 0 {
			for k, v := range d.Mutations {
				merged.Mutations[k] = v
			}
		}
		if !d.Allowed {
			return d
		}
		if d.Reason != "" {
			merged.Reason = d.Reason
		}
		if d.Severity != "" {
			merged.Severity = d.Severity
		}
	}
	return merged
}
