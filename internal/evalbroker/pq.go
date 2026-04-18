// Package evalbroker — priority-queue implementation backing
// Broker.ready.  Uses container/heap with a two-level ordering
// (priority desc, create-time asc) so older evals of the same
// priority win ties — prevents starvation of low-churn tenants
// sharing a busy broker with hot-loop workloads.
package evalbroker

// pq is a max-heap of Evaluation by priority, tiebreaking on older
// CreateTime.  container/heap is sad about generic types so we use
// the concrete-slice pattern.
type pq []Evaluation

// Len implements heap.Interface.
func (p pq) Len() int { return len(p) }

// Less orders highest priority first, then oldest create-time.
func (p pq) Less(i, j int) bool {
	if p[i].Priority != p[j].Priority {
		return p[i].Priority > p[j].Priority
	}
	return p[i].CreateTime.Before(p[j].CreateTime)
}

// Swap implements heap.Interface.
func (p pq) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Push appends; heap semantics re-order.
func (p *pq) Push(x interface{}) { *p = append(*p, x.(Evaluation)) }

// Pop removes the tail element — heap has already rotated the
// target into that slot before calling Pop.
func (p *pq) Pop() interface{} {
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[:n-1]
	return x
}
