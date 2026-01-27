package lineage

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryLineageManager in-memory lineage implementation
type InMemoryLineageManager struct {
	mu        sync.RWMutex
	nodes     map[string]*LineageNode
	edges     map[string]*LineageEdge
	processes map[string]*LineageProcess
	graphs    map[string]*LineageGraph
}

// NewInMemoryLineageManager creates manager
func NewInMemoryLineageManager() *InMemoryLineageManager {
	return &InMemoryLineageManager{
		nodes:     make(map[string]*LineageNode),
		edges:     make(map[string]*LineageEdge),
		processes: make(map[string]*LineageProcess),
		graphs:    make(map[string]*LineageGraph),
	}
}

// GetNode retrieves node
func (m *InMemoryLineageManager) GetNode(id string) (*LineageNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node not found")
	}
	return node, nil
}

// CreateNode creates node
func (m *InMemoryLineageManager) CreateNode(node *LineageNode) (*LineageNode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if node.ID == "" {
		node.ID = fmt.Sprintf("node-%d", time.Now().UnixNano())
	}

	m.nodes[node.ID] = node
	return node, nil
}

// ListNodes lists nodes
func (m *InMemoryLineageManager) ListNodes(nodeType string) ([]*LineageNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*LineageNode
	for _, n := range m.nodes {
		if nodeType != "" && string(n.NodeType) != nodeType {
			continue
		}
		result = append(result, n)
	}
	return result, nil
}

// CreateEdge creates edge
func (m *InMemoryLineageManager) CreateEdge(edge *LineageEdge) (*LineageEdge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if edge.ID == "" {
		edge.ID = fmt.Sprintf("edge-%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
	}

	m.edges[edge.ID] = edge
	return edge, nil
}

// ListEdges lists edges
func (m *InMemoryLineageManager) ListEdges(source, target string) ([]*LineageEdge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*LineageEdge
	for _, e := range m.edges {
		if source != "" && e.SourceNodeID != source {
			continue
		}
		if target != "" && e.TargetNodeID != target {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

// GetUpstream gets upstream nodes
func (m *InMemoryLineageManager) GetUpstream(nodeID string, depth int) ([]*LineageNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*LineageNode, 0)
	visited := make(map[string]bool)

	var traverse func(string, int)
	traverse = func(id string, d int) {
		if d <= 0 || visited[id] {
			return
		}
		visited[id] = true

		for _, e := range m.edges {
			if e.TargetNodeID == id {
				if node, exists := m.nodes[e.SourceNodeID]; exists {
					result = append(result, node)
					traverse(e.SourceNodeID, d-1)
				}
			}
		}
	}

	traverse(nodeID, depth)
	return result, nil
}

// GetDownstream gets downstream nodes
func (m *InMemoryLineageManager) GetDownstream(nodeID string, depth int) ([]*LineageNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*LineageNode, 0)
	visited := make(map[string]bool)

	var traverse func(string, int)
	traverse = func(id string, d int) {
		if d <= 0 || visited[id] {
			return
		}
		visited[id] = true

		for _, e := range m.edges {
			if e.SourceNodeID == id {
				if node, exists := m.nodes[e.TargetNodeID]; exists {
					result = append(result, node)
					traverse(e.TargetNodeID, d-1)
				}
			}
		}
	}

	traverse(nodeID, depth)
	return result, nil
}

// AnalyzeImpact analyzes change impact
func (m *InMemoryLineageManager) AnalyzeImpact(nodeID string) ([]*ImpactAnalysis, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	impacts := make([]*ImpactAnalysis, 0)
	downstream, _ := m.getDownstreamUnlocked(nodeID, 100)

	affectedNodes := make([]string, 0)
	for _, node := range downstream {
		affectedNodes = append(affectedNodes, node.ID)
	}

	impacts = append(impacts, &ImpactAnalysis{
		SourceNodeID:      nodeID,
		AffectedNodeCount: len(affectedNodes),
		AffectedNodes:     affectedNodes,
		EstimatedImpact:   "high",
	})

	return impacts, nil
}

// getDownstreamUnlocked helper for downstream traversal
func (m *InMemoryLineageManager) getDownstreamUnlocked(nodeID string, depth int) ([]*LineageNode, error) {
	result := make([]*LineageNode, 0)
	visited := make(map[string]bool)

	var traverse func(string, int)
	traverse = func(id string, d int) {
		if d <= 0 || visited[id] {
			return
		}
		visited[id] = true

		for _, e := range m.edges {
			if e.SourceNodeID == id {
				if node, exists := m.nodes[e.TargetNodeID]; exists {
					result = append(result, node)
					traverse(e.TargetNodeID, d-1)
				}
			}
		}
	}

	traverse(nodeID, depth)
	return result, nil
}

// GetColumnLineage gets column-level lineage
func (m *InMemoryLineageManager) GetColumnLineage(nodeID, columnName string) ([]*ColumnLineage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ColumnLineage, 0)
	node, exists := m.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found")
	}

	// Find column transformations
	for _, edge := range m.edges {
		if edge.TargetNodeID == nodeID {
			result = append(result, &ColumnLineage{
				ID:           fmt.Sprintf("col-lineage-%d", time.Now().UnixNano()),
				SourceColumn: fmt.Sprintf("%s.%s", edge.SourceNodeID, columnName),
				TargetColumn: fmt.Sprintf("%s.%s", nodeID, columnName),
				LastModified: time.Now(),
			})
		}
	}

	_ = node
	return result, nil
}

// CreateProcess creates process
func (m *InMemoryLineageManager) CreateProcess(process *LineageProcess) (*LineageProcess, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if process.ID == "" {
		process.ID = fmt.Sprintf("process-%d", time.Now().UnixNano())
	}

	m.processes[process.ID] = process
	return process, nil
}

// ListProcesses lists processes
func (m *InMemoryLineageManager) ListProcesses(nodeID string) ([]*LineageProcess, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*LineageProcess
	for _, p := range m.processes {
		if nodeID != "" && !contains(p.SourceNodes, nodeID) && !contains(p.TargetNodes, nodeID) {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// BuildGraph builds lineage graph
func (m *InMemoryLineageManager) BuildGraph(startNodeID string, direction string, depth int) (*LineageGraph, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	graph := &LineageGraph{
		ID:        fmt.Sprintf("graph-%d", time.Now().UnixNano()),
		RootNodes: []string{startNodeID},
		Depth:     depth,
		CreatedAt: time.Now(),
	}

	// Add root node
	if root, exists := m.nodes[startNodeID]; exists {
		graph.Nodes = append(graph.Nodes, *root)
	}

	// Add related nodes
	for _, edge := range m.edges {
		if edge.SourceNodeID == startNodeID || edge.TargetNodeID == startNodeID {
			graph.Edges = append(graph.Edges, *edge)
		}
	}

	m.graphs[graph.ID] = graph
	return graph, nil
}

// GetGraph retrieves graph
func (m *InMemoryLineageManager) GetGraph(id string) (*LineageGraph, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	graph, exists := m.graphs[id]
	if !exists {
		return nil, fmt.Errorf("graph not found")
	}
	return graph, nil
}

// GetLineageStatistics retrieves statistics
func (m *InMemoryLineageManager) GetLineageStatistics() (*LineageStatistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodeTypeCount := make(map[string]int)
	for _, n := range m.nodes {
		nodeTypeCount[string(n.NodeType)]++
	}

	return &LineageStatistics{
		TotalNodes: int64(len(m.nodes)),
		TotalEdges: int64(len(m.edges)),
	}, nil
}
