package scheduler

import (
	"sort"
	"sync"
	"time"
)

type Node struct {
	Name           string
	Zone           string
	Labels         map[string]string
	CapacityCPU    int
	CapacityMemory int
	UsedCPU        int
	UsedMemory     int
	Taints         []string
}

type Workload struct {
	Name          string
	Namespace     string
	Labels        map[string]string
	RequestCPU    int
	RequestMemory int
	NodeSelector  map[string]string
	Tolerations   []string
	Priority      int
	CreatedAt     time.Time
}

type ScheduleDecision struct {
	NodeName string
	Score    int
	Reasons  []string
}

type Scheduler struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

func NewScheduler() *Scheduler {
	return &Scheduler{nodes: make(map[string]*Node)}
}

func (s *Scheduler) UpsertNode(n Node) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := n
	if cp.Labels == nil {
		cp.Labels = map[string]string{}
	}
	s.nodes[n.Name] = &cp
}

func (s *Scheduler) ListNodes() []Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		out = append(out, *n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Scheduler) Score(w Workload) []ScheduleDecision {
	s.mu.RLock()
	nodes := make([]*Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		nodes = append(nodes, n)
	}
	s.mu.RUnlock()

	out := make([]ScheduleDecision, 0, len(nodes))
	for _, n := range nodes {
		cpuFree := n.CapacityCPU - n.UsedCPU
		memFree := n.CapacityMemory - n.UsedMemory
		if cpuFree < w.RequestCPU || memFree < w.RequestMemory {
			continue
		}
		score := cpuFree*2 + memFree
		reasons := []string{"fit:resources"}
		if n.Zone != "" {
			score += 5
			reasons = append(reasons, "zone:present")
		}
		out = append(out, ScheduleDecision{NodeName: n.Name, Score: score, Reasons: reasons})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

func (s *Scheduler) PickBest(w Workload) (ScheduleDecision, bool) {
	list := s.Score(w)
	if len(list) == 0 {
		return ScheduleDecision{}, false
	}
	return list[0], true
}
