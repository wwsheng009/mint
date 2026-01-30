// Package devtools provides the causal graph engine for DevTools.
//
// This file implements the causal graph that tracks the causal relationships
// between events, mutations, layout changes, and repaints.
package devtools

import (
	"sync"
	"time"
)

// CausalGraph represents a causal graph of a single frame.
type CausalGraph struct {
	mu        sync.RWMutex

	// Frame identification
	FrameID   FrameID
	StartTime time.Time
	EndTime   time.Time

	// Nodes
	Events    []*CausalEvent
	Mutations []*CausalMutation
	Layouts   []*CausalLayout
	Repaints  []*CausalRepaint

	// Causal edges
	Edges    []*CausalEdge

	// Indexes for fast lookup
	eventIndex    map[EventID]int
	mutationIndex map[MutationID]int
	layoutIndex   map[NodeID]int
	repaintIndex  map[RepaintID]int
}

// NewCausalGraph creates a new causal graph for the given frame.
func NewCausalGraph(frameID FrameID) *CausalGraph {
	return &CausalGraph{
		FrameID:       frameID,
		StartTime:     time.Now(),
		Events:        make([]*CausalEvent, 0, 16),
		Mutations:     make([]*CausalMutation, 0, 32),
		Layouts:        make([]*CausalLayout, 0, 32),
		Repaints:       make([]*CausalRepaint, 0, 16),
		Edges:         make([]*CausalEdge, 0, 64),
		eventIndex:    make(map[EventID]int),
		mutationIndex: make(map[MutationID]int),
		layoutIndex:   make(map[NodeID]int),
		repaintIndex:  make(map[RepaintID]int),
	}
}

// AddEvent adds an event node to the causal graph.
func (cg *CausalGraph) AddEvent(eventType string, targetID NodeID, phase string) EventID {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	id := EventID(len(cg.Events) + 1)
	node := &CausalEvent{
		ID:       id,
		Type:     eventType,
		TargetID: targetID,
		Phase:    phase,
		Time:     time.Now(),
	}

	cg.Events = append(cg.Events, node)
	cg.eventIndex[id] = len(cg.Events) - 1

	return id
}

// AddMutation adds a mutation node to the causal graph.
func (cg *CausalGraph) AddMutation(component string, kind MutationKind, field string, oldValue, newValue interface{}, causedBy EventID) MutationID {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	id := MutationID(len(cg.Mutations) + 1)
	node := &CausalMutation{
		ID:        id,
		Component: component,
		Kind:      kind,
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		Time:      time.Now(),
	}

	cg.Mutations = append(cg.Mutations, node)
	cg.mutationIndex[id] = len(cg.Mutations) - 1

	// Add causal edge from event to mutation
	if causedBy != 0 {
		cg.addEdge(&CausalEdge{
			From: uint64(causedBy),
			To:   uint64(id),
			Type: EdgeEventToMutation,
		})
	}

	return id
}

// AddLayoutChange adds a layout change node to the causal graph.
func (cg *CausalGraph) AddLayoutChange(nodeID NodeID, changeMask ChangeMask, oldRect, newRect *Rect, causedBy []MutationID) LayoutID {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	id := LayoutID(len(cg.Layouts) + 1)
	node := &CausalLayout{
		ID:         id,
		NodeID:     nodeID,
		ChangeMask: changeMask,
		OldRect:    oldRect,
		NewRect:    newRect,
		Time:       time.Now(),
	}

	cg.Layouts = append(cg.Layouts, node)
	cg.layoutIndex[nodeID] = len(cg.Layouts) - 1

	// Add causal edges from mutations to layout
	for _, mutID := range causedBy {
		cg.addEdge(&CausalEdge{
			From: uint64(mutID),
			To:   uint64(id),
			Type: EdgeMutationToLayout,
		})
	}

	return id
}

// AddRepaint adds a repaint node to the causal graph.
func (cg *CausalGraph) AddRepaint(dirtyRegions []Rect, changedCells, totalCells int, causedBy []LayoutID) RepaintID {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	id := RepaintID(len(cg.Repaints) + 1)
	node := &CausalRepaint{
		ID:           id,
		DirtyRegions: dirtyRegions,
		ChangedCells: changedCells,
		TotalCells:   totalCells,
		Time:         time.Now(),
	}

	cg.Repaints = append(cg.Repaints, node)
	cg.repaintIndex[id] = len(cg.Repaints) - 1

	// Add causal edges from layout to repaint
	for _, layoutID := range causedBy {
		cg.addEdge(&CausalEdge{
			From: uint64(layoutID),
			To:   uint64(id),
			Type: EdgeLayoutToRepaint,
		})
	}

	return id
}

// addEdge adds a causal edge to the graph.
func (cg *CausalGraph) addEdge(edge *CausalEdge) {
	cg.Edges = append(cg.Edges, edge)
}

// Finalize finalizes the causal graph for the frame.
func (cg *CausalGraph) Finalize() {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	cg.EndTime = time.Now()
}

// GetEvent retrieves an event node by ID.
func (cg *CausalGraph) GetEvent(id EventID) *CausalEvent {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	if idx, ok := cg.eventIndex[id]; ok {
		return cg.Events[idx]
	}
	return nil
}

// GetMutation retrieves a mutation node by ID.
func (cg *CausalGraph) GetMutation(id MutationID) *CausalMutation {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	if idx, ok := cg.mutationIndex[id]; ok {
		return cg.Mutations[idx]
	}
	return nil
}

// GetLayout retrieves a layout node by node ID.
func (cg *CausalGraph) GetLayout(nodeID NodeID) *CausalLayout {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	if idx, ok := cg.layoutIndex[nodeID]; ok {
		return cg.Layouts[idx]
	}
	return nil
}

// GetRepaint retrieves a repaint node by ID.
func (cg *CausalGraph) GetRepaint(id RepaintID) *CausalRepaint {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	if idx, ok := cg.repaintIndex[id]; ok {
		return cg.Repaints[idx]
	}
	return nil
}

// GetFrameSummary returns a summary of the frame.
func (cg *CausalGraph) GetFrameSummary() *FrameSummary {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	return &FrameSummary{
		FrameID:      cg.FrameID,
		StartTime:    cg.StartTime,
		EndTime:      cg.EndTime,
		Duration:     cg.EndTime.Sub(cg.StartTime),
		EventCount:   len(cg.Events),
		MutationCount: len(cg.Mutations),
		LayoutCount:  len(cg.Layouts),
		RepaintCount: len(cg.Repaints),
		EdgeCount:    len(cg.Edges),
	}
}

// CausalEvent represents an event in the causal graph.
type CausalEvent struct {
	ID       EventID
	Type     string
	TargetID NodeID
	Phase    string
	Time     time.Time
}

// CausalMutation represents a state mutation in the causal graph.
type CausalMutation struct {
	ID        MutationID
	Component string
	Kind      MutationKind
	Field     string
	OldValue  interface{}
	NewValue  interface{}
	Time      time.Time
}

// CausalLayout represents a layout change in the causal graph.
type CausalLayout struct {
	ID         LayoutID
	NodeID     NodeID
	ChangeMask ChangeMask
	OldRect    *Rect
	NewRect    *Rect
	Time       time.Time
}

// CausalRepaint represents a repaint operation in the causal graph.
type CausalRepaint struct {
	ID           RepaintID
	DirtyRegions []Rect
	ChangedCells int
	TotalCells   int
	Time         time.Time
}

// CausalEdge represents a causal relationship between two nodes.
type CausalEdge struct {
	From uint64    // Source node ID
	To   uint64    // Target node ID
	Type EdgeType  // Type of causal relationship
}

// FrameSummary provides a summary of a frame's causal graph.
type FrameSummary struct {
	FrameID      FrameID
	StartTime    time.Time
	EndTime      time.Time
	Duration    time.Duration
	EventCount   int
	MutationCount int
	LayoutCount  int
	RepaintCount int
	EdgeCount    int
}
