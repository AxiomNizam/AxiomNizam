package modes

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type Mode string

const (
	ModeRealtime     Mode = "realtime"
	ModeForensics    Mode = "forensics"
	ModeThreatHunt   Mode = "threat-hunt"
	ModeCapacityPlan Mode = "capacity-planning"
	ModeAnomalyLab   Mode = "anomaly-lab"
)

type ModeConfig struct {
	Name           Mode
	Enabled        bool
	SamplingRate   float64
	RetentionHours int
	AlertThreshold float64
	Labels         map[string]string
}

type ModeEvent struct {
	Mode      Mode
	EventType string
	Source    string
	Payload   map[string]any
	Timestamp time.Time
}

type Manager struct {
	mu      sync.RWMutex
	configs map[Mode]ModeConfig
	events  []ModeEvent
}

func NewManager() *Manager {
	return &Manager{
		configs: map[Mode]ModeConfig{},
		events:  make([]ModeEvent, 0, 1024),
	}
}

func (m *Manager) Upsert(cfg ModeConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cfg.Labels == nil {
		cfg.Labels = map[string]string{}
	}
	m.configs[cfg.Name] = cfg
}

func (m *Manager) Get(name Mode) (ModeConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, ok := m.configs[name]
	return cfg, ok
}

func (m *Manager) List() []ModeConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ModeConfig, 0, len(m.configs))
	for _, v := range m.configs {
		out = append(out, v)
	}
	sort.SliceStable(out, func(i, j int) bool { return string(out[i].Name) < string(out[j].Name) })
	return out
}

func (m *Manager) Record(ev ModeEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, ev)
	if len(m.events) > 50000 {
		m.events = m.events[len(m.events)-50000:]
	}
}

func (m *Manager) FindByMode(mode Mode) []ModeEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ModeEvent, 0, 128)
	for _, ev := range m.events {
		if strings.EqualFold(string(ev.Mode), string(mode)) {
			out = append(out, ev)
		}
	}
	return out
}
