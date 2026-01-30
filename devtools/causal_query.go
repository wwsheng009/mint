// Package devtools provides the causal query API for DevTools.
//
// This file implements the CausalQuery API that allows querying
// the causal graph to trace cause and effect relationships.
package devtools

// CausalQuery provides query capabilities for causal graphs.
type CausalQuery struct {
	graph *CausalGraph
}

// NewCausalQuery creates a new causal query for the given graph.
func NewCausalQuery(graph *CausalGraph) *CausalQuery {
	return &CausalQuery{graph: graph}
}

// FindRootCauses finds all root events that ultimately caused the given repaint.
func (q *CausalQuery) FindRootCauses(repaintID RepaintID) []*CausalEvent {
	if q.graph == nil {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	// Traverse backwards from repaint to find all events
	events := make([]*CausalEvent, 0)
	visited := make(map[uint64]bool)

	q.findCauses(uint64(repaintID), visited, &events)

	return events
}

// findCauses recursively finds all causes of a node.
func (q *CausalQuery) findCauses(nodeID uint64, visited map[uint64]bool, events *[]*CausalEvent) {
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	// Find incoming edges
	for _, edge := range q.graph.Edges {
		if edge.To == nodeID {
			// Check if source is an event (EventID starts from 1, events are in order)
			if edge.From <= uint64(len(q.graph.Events)) {
				event := q.graph.Events[int(edge.From)-1]
				if event != nil {
					*events = append(*events, event)
				}
			}
			// Recursively find causes of the source
			q.findCauses(edge.From, visited, events)
		}
	}
}

// FindEffects finds all effects caused by the given event.
func (q *CausalQuery) FindEffects(eventID EventID) *EffectChain {
	if q.graph == nil {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	chain := &EffectChain{
		EventID: eventID,
	}

	visited := make(map[uint64]bool)
	q.findEffects(uint64(eventID), visited, chain)

	return chain
}

// findEffects recursively finds all effects of a node.
func (q *CausalQuery) findEffects(nodeID uint64, visited map[uint64]bool, chain *EffectChain) {
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	// Find outgoing edges
	for _, edge := range q.graph.Edges {
		if edge.From == nodeID {
			// Categorize the effect
			switch edge.Type {
			case EdgeEventToMutation:
				if int(edge.To) <= len(q.graph.Mutations) {
					chain.Mutations = append(chain.Mutations, q.graph.Mutations[int(edge.To)-1])
				}
			case EdgeMutationToLayout:
				if int(edge.To) <= len(q.graph.Layouts) {
					chain.Layouts = append(chain.Layouts, q.graph.Layouts[int(edge.To)-1])
				}
			case EdgeLayoutToRepaint:
				if int(edge.To) <= len(q.graph.Repaints) {
					chain.Repaints = append(chain.Repaints, q.graph.Repaints[int(edge.To)-1])
				}
			}
			// Recursively find effects of the target
			q.findEffects(edge.To, visited, chain)
		}
	}
}

// FindMutationChain finds the mutation chain between an event and a layout change.
func (q *CausalQuery) FindMutationChain(eventID EventID, layoutID LayoutID) []*CausalMutation {
	if q.graph == nil {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	// Find all mutations caused by this event that caused this layout
	var mutations []*CausalMutation

	for _, edge := range q.graph.Edges {
		if edge.From == uint64(eventID) && edge.Type == EdgeEventToMutation {
			// This mutation was caused by the event
			mutationIdx := int(edge.To) - 1
			if mutationIdx >= 0 && mutationIdx < len(q.graph.Mutations) {
				mutation := q.graph.Mutations[mutationIdx]

				// Check if this mutation caused the target layout
				for _, layoutEdge := range q.graph.Edges {
					if layoutEdge.From == edge.To && layoutEdge.To == uint64(layoutID) {
						mutations = append(mutations, mutation)
						break
					}
				}
			}
		}
	}

	return mutations
}

// FindLayoutsForNode finds all layout changes for a specific node.
func (q *CausalQuery) FindLayoutsForNode(nodeID NodeID) []*CausalLayout {
	if q.graph == nil {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	var layouts []*CausalLayout
	for _, layout := range q.graph.Layouts {
		if layout.NodeID == nodeID {
			layouts = append(layouts, layout)
		}
	}

	return layouts
}

// FindMutationsByComponent finds all mutations for a specific component.
func (q *CausalQuery) FindMutationsByComponent(component string) []*CausalMutation {
	if q.graph == nil {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	var mutations []*CausalMutation
	for _, mutation := range q.graph.Mutations {
		if mutation.Component == component {
			mutations = append(mutations, mutation)
		}
	}

	return mutations
}

// GetCausalPath returns a causal path from an event to a repaint.
func (q *CausalQuery) GetCausalPath(eventID EventID, repaintID RepaintID) *CausalPath {
	if q.graph == nil {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	path := &CausalPath{
		StartEvent: q.graph.GetEvent(eventID),
		EndRepaint: q.graph.GetRepaint(repaintID),
	}

	// Build path using BFS
	visited := make(map[uint64]bool)
	queue := []causalPathNode{{id: uint64(eventID), path: []interface{}{}}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.id == uint64(repaintID) {
			// Found path
			path.Path = current.path
			return path
		}

		if visited[current.id] {
			continue
		}
		visited[current.id] = true

		// Find neighbors
		for _, edge := range q.graph.Edges {
			if edge.From == current.id {
				nextPath := make([]interface{}, len(current.path)+1)
				copy(nextPath, current.path)

				// Add node to path based on type
				switch edge.Type {
				case EdgeEventToMutation:
					if int(edge.To) <= len(q.graph.Mutations) {
						nextPath[len(current.path)] = q.graph.Mutations[int(edge.To)-1]
					}
				case EdgeMutationToLayout:
					if int(edge.To) <= len(q.graph.Layouts) {
						nextPath[len(current.path)] = q.graph.Layouts[int(edge.To)-1]
					}
				case EdgeLayoutToRepaint:
					if int(edge.To) <= len(q.graph.Repaints) {
						nextPath[len(current.path)] = q.graph.Repaints[int(edge.To)-1]
					}
				}

				queue = append(queue, causalPathNode{
					id:   edge.To,
					path: nextPath,
				})
			}
		}
	}

	return nil
}

// causalPathNode is a helper for BFS path finding.
type causalPathNode struct {
	id   uint64
	path []interface{}
}

// EffectChain represents the chain of effects caused by an event.
type EffectChain struct {
	EventID  EventID
	Mutations []*CausalMutation
	Layouts   []*CausalLayout
	Repaints  []*CausalRepaint
}

// CausalPath represents a causal path from an event to a repaint.
type CausalPath struct {
	StartEvent *CausalEvent
	EndRepaint *CausalRepaint
	Path       []interface{} // Mixed CausalMutation, CausalLayout, CausalRepaint
}

// TraceResult represents a complete trace result for debugging.
type TraceResult struct {
	FrameID      FrameID
	RootEvent    *CausalEvent
	EffectChain  *EffectChain
	AffectedNodes []NodeID
	DirtyRegions []Rect
}

// TraceFromEvent performs a complete trace from an event to all its effects.
func (q *CausalQuery) TraceFromEvent(eventID EventID) *TraceResult {
	if q.graph == nil {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	event := q.graph.GetEvent(eventID)
	if event == nil {
		return nil
	}

	result := &TraceResult{
		FrameID:   q.graph.FrameID,
		RootEvent: event,
	}

	// Find all effects
	chain := &EffectChain{EventID: eventID}
	visited := make(map[uint64]bool)
	q.findEffects(uint64(eventID), visited, chain)
	result.EffectChain = chain

	// Collect affected nodes
	nodeSet := make(map[NodeID]bool)
	for _, layout := range chain.Layouts {
		nodeSet[layout.NodeID] = true
	}
	result.AffectedNodes = make([]NodeID, 0, len(nodeSet))
	for nodeID := range nodeSet {
		result.AffectedNodes = append(result.AffectedNodes, nodeID)
	}

	// Collect dirty regions
	for _, repaint := range chain.Repaints {
		result.DirtyRegions = append(result.DirtyRegions, repaint.DirtyRegions...)
	}

	return result
}

// GetCriticalPath returns the critical path (longest path) in the causal graph.
func (q *CausalQuery) GetCriticalPath() *CausalPath {
	if q.graph == nil || len(q.graph.Events) == 0 {
		return nil
	}

	q.graph.mu.RLock()
	defer q.graph.mu.RUnlock()

	var longestPath *CausalPath
	maxLength := 0

	// Try each event as a starting point
	for _, event := range q.graph.Events {
		// Find furthest repaint
		for _, repaint := range q.graph.Repaints {
			path := q.getBFSPath(uint64(event.ID), uint64(repaint.ID))
			if path != nil && len(path) > maxLength {
				maxLength = len(path)
				longestPath = &CausalPath{
					StartEvent: event,
					EndRepaint: repaint,
					Path:       path,
				}
			}
		}
	}

	return longestPath
}

// getBFSPath finds a path between two nodes using BFS.
func (q *CausalQuery) getBFSPath(from, to uint64) []interface{} {
	if from == to {
		return []interface{}{}
	}

	visited := make(map[uint64]bool)
	parent := make(map[uint64]causalParent)
	queue := []uint64{from}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == to {
			// Reconstruct path
			return q.reconstructPath(from, to, parent)
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		// Find neighbors
		for _, edge := range q.graph.Edges {
			if edge.From == current && !visited[edge.To] {
				if _, exists := parent[edge.To]; !exists {
					parent[edge.To] = causalParent{
						parent: current,
						edge:   edge,
					}
					queue = append(queue, edge.To)
				}
			}
		}
	}

	return nil
}

// causalParent stores parent information for path reconstruction.
type causalParent struct {
	parent uint64
	edge   *CausalEdge
}

// reconstructPath reconstructs a path from parent pointers.
func (q *CausalQuery) reconstructPath(from, to uint64, parent map[uint64]causalParent) []interface{} {
	var path []interface{}
	current := to

	for current != from {
		p, exists := parent[current]
		if !exists {
			return nil
		}

		// Add node based on edge type
		switch p.edge.Type {
		case EdgeEventToMutation:
			if int(current) <= len(q.graph.Mutations) {
				path = append([]interface{}{q.graph.Mutations[int(current)-1]}, path...)
			}
		case EdgeMutationToLayout:
			if int(current) <= len(q.graph.Layouts) {
				path = append([]interface{}{q.graph.Layouts[int(current)-1]}, path...)
			}
		case EdgeLayoutToRepaint:
			if int(current) <= len(q.graph.Repaints) {
				path = append([]interface{}{q.graph.Repaints[int(current)-1]}, path...)
			}
		}

		current = p.parent
	}

	return path
}
