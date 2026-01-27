package lineage

import (
	"fmt"
	"sync"
	"time"
)

// DataLineageNode represents a data entity in lineage
type DataLineageNode struct {
	ID            string
	Type          string // table, column, view, procedure
	Name          string
	Schema        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Metadata      map[string]interface{}
}

// DataLineageEdge represents a relationship between nodes
type DataLineageEdge struct {
	ID            string
	SourceID      string
	TargetID      string
	RelationType  string // reads, writes, transforms, aggregates
	CreatedAt     time.Time
	Metadata      map[string]string
}

// LineageTracer traces data lineage
type LineageTracer struct {
	ID        string
	Timestamp time.Time
	QueryID   string
	Operation string
	Source    string
	Target    string
	DataFlow  []string // chain of data flow
}

// DataLineageTracker manages data lineage
type DataLineageTracker struct {
	mu                sync.RWMutex
	nodes             map[string]*DataLineageNode
	edges             map[string]*DataLineageEdge
	lineageTraces     []*LineageTracer
	dependencyGraph   map[string][]string
	impactAnalysis    map[string]*ImpactAnalysis
	maxTraceSize      int
	transformLog      []*DataTransformation
	maxTransformSize  int
}

// ImpactAnalysis analyzes data impact
type ImpactAnalysis struct {
	SourceNodeID      string
	AffectedNodes     []string
	AffectedEdges     []string
	ImpactLevel       string // low, medium, high, critical
	ChangeDescription string
	AnalyzedAt        time.Time
}

// DataTransformation logs data transformations
type DataTransformation struct {
	ID              string
	Timestamp       time.Time
	SourceTable     string
	TargetTable     string
	SourceColumns   []string
	TargetColumns   []string
	TransformType   string // join, aggregate, filter, map
	Query           string
	ExecutionTime   int64 // milliseconds
}

// NewDataLineageTracker creates lineage tracker
func NewDataLineageTracker() *DataLineageTracker {
	return &DataLineageTracker{
		nodes:           make(map[string]*DataLineageNode),
		edges:           make(map[string]*DataLineageEdge),
		lineageTraces:   make([]*LineageTracer, 0),
		dependencyGraph: make(map[string][]string),
		impactAnalysis:  make(map[string]*ImpactAnalysis),
		transformLog:    make([]*DataTransformation, 0),
		maxTraceSize:    50000,
		maxTransformSize: 10000,
	}
}

// RegisterDataNode registers a data entity node
func (dlt *DataLineageTracker) RegisterDataNode(node *DataLineageNode) error {
	dlt.mu.Lock()
	defer dlt.mu.Unlock()

	if node.ID == "" {
		node.ID = fmt.Sprintf("node-%d", time.Now().UnixNano())
	}

	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	dlt.nodes[node.ID] = node
	dlt.dependencyGraph[node.ID] = make([]string, 0)

	return nil
}

// CreateLineageEdge creates a lineage relationship
func (dlt *DataLineageTracker) CreateLineageEdge(sourceID, targetID, relationType string) (*DataLineageEdge, error) {
	dlt.mu.Lock()
	defer dlt.mu.Unlock()

	edge := &DataLineageEdge{
		ID:           fmt.Sprintf("edge-%d", time.Now().UnixNano()),
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]string),
	}

	dlt.edges[edge.ID] = edge

	// Update dependency graph
	if _, exists := dlt.dependencyGraph[sourceID]; exists {
		dlt.dependencyGraph[sourceID] = append(dlt.dependencyGraph[sourceID], targetID)
	}

	return edge, nil
}

// TraceDataFlow traces the flow of data
func (dlt *DataLineageTracker) TraceDataFlow(source, target string) (*LineageTracer, error) {
	dlt.mu.Lock()
	defer dlt.mu.Unlock()

	trace := &LineageTracer{
		ID:       fmt.Sprintf("trace-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Source:    source,
		Target:    target,
		DataFlow:  make([]string, 0),
	}

	// Build flow path using BFS
	path := dlt.findDataFlowPath(source, target)
	trace.DataFlow = path

	dlt.lineageTraces = append(dlt.lineageTraces, trace)
	if len(dlt.lineageTraces) > dlt.maxTraceSize {
		dlt.lineageTraces = dlt.lineageTraces[1:]
	}

	return trace, nil
}

// findDataFlowPath finds path from source to target
func (dlt *DataLineageTracker) findDataFlowPath(source, target string) []string {
	path := []string{source}

	// BFS implementation
	queue := []string{source}
	visited := make(map[string]bool)
	parent := make(map[string]string)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		if current == target {
			// Rebuild path
			var fullPath []string
			node := target
			for node != "" {
				fullPath = append([]string{node}, fullPath...)
				node = parent[node]
			}
			return fullPath
		}

		if deps, exists := dlt.dependencyGraph[current]; exists {
			for _, dep := range deps {
				if !visited[dep] {
					queue = append(queue, dep)
					parent[dep] = current
				}
			}
		}
	}

	return path
}

// GetUpstreamLineage gets upstream data sources
func (dlt *DataLineageTracker) GetUpstreamLineage(nodeID string) []*DataLineageNode {
	dlt.mu.RLock()
	defer dlt.mu.RUnlock()

	upstream := make([]*DataLineageNode, 0)
	visited := make(map[string]bool)

	queue := []string{nodeID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		// Find edges where current is target
		for _, edge := range dlt.edges {
			if edge.TargetID == current {
				sourceNode := dlt.nodes[edge.SourceID]
				if sourceNode != nil {
					upstream = append(upstream, sourceNode)
					queue = append(queue, edge.SourceID)
				}
			}
		}
	}

	return upstream
}

// GetDownstreamLineage gets downstream data consumers
func (dlt *DataLineageTracker) GetDownstreamLineage(nodeID string) []*DataLineageNode {
	dlt.mu.RLock()
	defer dlt.mu.RUnlock()

	downstream := make([]*DataLineageNode, 0)
	visited := make(map[string]bool)

	if deps, exists := dlt.dependencyGraph[nodeID]; exists {
		queue := deps
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if visited[current] {
				continue
			}
			visited[current] = true

			targetNode := dlt.nodes[current]
			if targetNode != nil {
				downstream = append(downstream, targetNode)
			}

			if curDeps, exists := dlt.dependencyGraph[current]; exists {
				queue = append(queue, curDeps...)
			}
		}
	}

	return downstream
}

// AnalyzeImpact analyzes impact of changes
func (dlt *DataLineageTracker) AnalyzeImpact(nodeID string) *ImpactAnalysis {
	dlt.mu.Lock()
	defer dlt.mu.Unlock()

	affected := dlt.GetDownstreamLineage(nodeID)
	impactLevel := "low"

	if len(affected) > 10 {
		impactLevel = "high"
	} else if len(affected) > 5 {
		impactLevel = "medium"
	}

	affectedNodeIDs := make([]string, 0)
	affectedEdgeIDs := make([]string, 0)

	for _, node := range affected {
		affectedNodeIDs = append(affectedNodeIDs, node.ID)
	}

	// Find affected edges
	for _, edge := range dlt.edges {
		for _, nodeID := range affectedNodeIDs {
			if edge.TargetID == nodeID || edge.SourceID == nodeID {
				affectedEdgeIDs = append(affectedEdgeIDs, edge.ID)
			}
		}
	}

	analysis := &ImpactAnalysis{
		SourceNodeID:      nodeID,
		AffectedNodes:     affectedNodeIDs,
		AffectedEdges:     affectedEdgeIDs,
		ImpactLevel:       impactLevel,
		ChangeDescription: fmt.Sprintf("Change affects %d nodes", len(affectedNodeIDs)),
		AnalyzedAt:        time.Now(),
	}

	dlt.impactAnalysis[nodeID] = analysis
	return analysis
}

// LogTransformation logs a data transformation
func (dlt *DataLineageTracker) LogTransformation(transform *DataTransformation) {
	dlt.mu.Lock()
	defer dlt.mu.Unlock()

	if transform.ID == "" {
		transform.ID = fmt.Sprintf("trans-%d", time.Now().UnixNano())
	}

	if transform.Timestamp.IsZero() {
		transform.Timestamp = time.Now()
	}

	dlt.transformLog = append(dlt.transformLog, transform)

	if len(dlt.transformLog) > dlt.maxTransformSize {
		dlt.transformLog = dlt.transformLog[1:]
	}
}

// GetLineageTraces gets lineage traces
func (dlt *DataLineageTracker) GetLineageTraces(limit int) []*LineageTracer {
	dlt.mu.RLock()
	defer dlt.mu.RUnlock()

	if limit > len(dlt.lineageTraces) {
		limit = len(dlt.lineageTraces)
	}
	if limit == 0 {
		return make([]*LineageTracer, 0)
	}

	return dlt.lineageTraces[len(dlt.lineageTraces)-limit:]
}

// GetTransformationLog gets transformation history
func (dlt *DataLineageTracker) GetTransformationLog(limit int) []*DataTransformation {
	dlt.mu.RLock()
	defer dlt.mu.RUnlock()

	if limit > len(dlt.transformLog) {
		limit = len(dlt.transformLog)
	}
	if limit == 0 {
		return make([]*DataTransformation, 0)
	}

	return dlt.transformLog[len(dlt.transformLog)-limit:]
}

// GetLineageGraph returns full lineage information
func (dlt *DataLineageTracker) GetLineageGraph() map[string]interface{} {
	dlt.mu.RLock()
	defer dlt.mu.RUnlock()

	return map[string]interface{}{
		"nodes":           len(dlt.nodes),
		"edges":           len(dlt.edges),
		"traces":          len(dlt.lineageTraces),
		"transformations": len(dlt.transformLog),
		"dependency_graph": dlt.dependencyGraph,
	}
}

// GetLineageStats returns lineage statistics
func (dlt *DataLineageTracker) GetLineageStats() map[string]interface{} {
	dlt.mu.RLock()
	defer dlt.mu.RUnlock()

	avgTransformTime := float64(0)
	if len(dlt.transformLog) > 0 {
		totalTime := int64(0)
		for _, t := range dlt.transformLog {
			totalTime += t.ExecutionTime
		}
		avgTransformTime = float64(totalTime) / float64(len(dlt.transformLog))
	}

	return map[string]interface{}{
		"total_nodes":               len(dlt.nodes),
		"total_edges":               len(dlt.edges),
		"total_traces":              len(dlt.lineageTraces),
		"total_transformations":     len(dlt.transformLog),
		"avg_transformation_time":   avgTransformTime,
		"impact_analyses":           len(dlt.impactAnalysis),
	}
}
