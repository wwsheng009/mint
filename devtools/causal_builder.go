// Package devtools provides the causal builder for DevTools.
//
// This file implements the CausalBuilder that builds causal graphs
// by linking events, mutations, layout changes, and repaints.
package devtools

import (
	"sync"
	"sync/atomic"
)

// CausalBuilder builds causal graphs by tracking relationships between
// events, mutations, layout changes, and repaints.
type CausalBuilder struct {
	enabled     atomic.Uint32
	currentGraph atomic.Pointer[CausalGraph]

	// Tracking current event context
	currentEvent atomic.Value // EventID

	// Tracking mutations for current frame
	pendingMutations []MutationID
	pendingMu        sync.Mutex

	// Tracking layout changes for current frame
	pendingLayouts []LayoutID
	layoutMu       sync.Mutex

	// Completed graphs for querying
	completedGraphs []*CausalGraph
	completedMu     sync.RWMutex

	// Config
	maxGraphs int
}

// NewCausalBuilder creates a new causal builder.
func NewCausalBuilder() *CausalBuilder {
	cb := &CausalBuilder{
		pendingMutations: make([]MutationID, 0, 32),
		pendingLayouts:   make([]LayoutID, 0, 32),
		completedGraphs:  make([]*CausalGraph, 0, 100),
		maxGraphs:        100, // Keep last 100 frames
	}
	cb.enabled.Store(0)
	cb.currentEvent.Store(EventID(0))
	return cb
}

// Enable enables the causal builder.
func (cb *CausalBuilder) Enable() {
	cb.enabled.Store(1)
}

// Disable disables the causal builder.
func (cb *CausalBuilder) Disable() {
	cb.enabled.Store(0)
}

// IsEnabled returns whether the causal builder is enabled.
func (cb *CausalBuilder) IsEnabled() bool {
	return cb.enabled.Load() == 1
}

// BeginFrame starts a new frame's causal graph.
func (cb *CausalBuilder) BeginFrame(frameID FrameID) *CausalGraph {
	if !cb.IsEnabled() {
		return nil
	}

	// Finalize previous graph if exists
	if prev := cb.currentGraph.Load(); prev != nil {
		prev.Finalize()
		cb.addCompletedGraph(prev)
	}

	// Create new graph for this frame
	graph := NewCausalGraph(frameID)
	cb.currentGraph.Store(graph)

	// Reset pending tracking
	cb.pendingMu.Lock()
	cb.pendingMutations = cb.pendingMutations[:0]
	cb.pendingMu.Unlock()

	cb.layoutMu.Lock()
	cb.pendingLayouts = cb.pendingLayouts[:0]
	cb.layoutMu.Unlock()

	return graph
}

// EndFrame ends the current frame and keeps the graph active for potential additions.
func (cb *CausalBuilder) EndFrame() {
	if !cb.IsEnabled() {
		return
	}

	// Clear current event context
	cb.currentEvent.Store(EventID(0))
}

// RecordEvent records an event and sets it as the current causal context.
func (cb *CausalBuilder) RecordEvent(eventType string, targetID NodeID, phase string) EventID {
	if !cb.IsEnabled() {
		return 0
	}

	graph := cb.currentGraph.Load()
	if graph == nil {
		return 0
	}

	eventID := graph.AddEvent(eventType, targetID, phase)
	cb.currentEvent.Store(eventID)
	return eventID
}

// RecordMutation records a mutation caused by the current event.
func (cb *CausalBuilder) RecordMutation(component string, kind MutationKind, field string, oldValue, newValue interface{}) MutationID {
	if !cb.IsEnabled() {
		return 0
	}

	graph := cb.currentGraph.Load()
	if graph == nil {
		return 0
	}

	// Get current event as cause
	causedBy, _ := cb.currentEvent.Load().(EventID)
	if causedBy == 0 {
		// No current event, use zero event ID
		causedBy = 0
	}

	mutationID := graph.AddMutation(component, kind, field, oldValue, newValue, causedBy)

	// Track mutation for layout linking
	cb.pendingMu.Lock()
	cb.pendingMutations = append(cb.pendingMutations, mutationID)
	cb.pendingMu.Unlock()

	return mutationID
}

// RecordLayoutChange records a layout change caused by pending mutations.
func (cb *CausalBuilder) RecordLayoutChange(nodeID NodeID, changeMask ChangeMask, oldRect, newRect *Rect) LayoutID {
	if !cb.IsEnabled() {
		return 0
	}

	graph := cb.currentGraph.Load()
	if graph == nil {
		return 0
	}

	// Get pending mutations as causes
	cb.pendingMu.Lock()
	causedBy := make([]MutationID, len(cb.pendingMutations))
	copy(causedBy, cb.pendingMutations)
	cb.pendingMutations = cb.pendingMutations[:0] // Clear after use
	cb.pendingMu.Unlock()

	layoutID := graph.AddLayoutChange(nodeID, changeMask, oldRect, newRect, causedBy)

	// Track layout for repaint linking
	cb.layoutMu.Lock()
	cb.pendingLayouts = append(cb.pendingLayouts, layoutID)
	cb.layoutMu.Unlock()

	return layoutID
}

// RecordRepaint records a repaint caused by pending layout changes.
func (cb *CausalBuilder) RecordRepaint(dirtyRegions []Rect, changedCells, totalCells int) RepaintID {
	if !cb.IsEnabled() {
		return 0
	}

	graph := cb.currentGraph.Load()
	if graph == nil {
		return 0
	}

	// Get pending layouts as causes
	cb.layoutMu.Lock()
	causedBy := make([]LayoutID, len(cb.pendingLayouts))
	copy(causedBy, cb.pendingLayouts)
	cb.pendingLayouts = cb.pendingLayouts[:0] // Clear after use
	cb.layoutMu.Unlock()

	return graph.AddRepaint(dirtyRegions, changedCells, totalCells, causedBy)
}

// GetCurrentGraph returns the current frame's causal graph.
func (cb *CausalBuilder) GetCurrentGraph() *CausalGraph {
	return cb.currentGraph.Load()
}

// GetGraph returns a completed causal graph by frame ID.
func (cb *CausalBuilder) GetGraph(frameID FrameID) *CausalGraph {
	cb.completedMu.RLock()
	defer cb.completedMu.RUnlock()

	for _, graph := range cb.completedGraphs {
		if graph.FrameID == frameID {
			return graph
		}
	}
	return nil
}

// GetAllGraphs returns all completed causal graphs.
func (cb *CausalBuilder) GetAllGraphs() []*CausalGraph {
	cb.completedMu.RLock()
	defer cb.completedMu.RUnlock()

	graphs := make([]*CausalGraph, len(cb.completedGraphs))
	copy(graphs, cb.completedGraphs)
	return graphs
}

// GetLastNGraphs returns the last N completed causal graphs.
func (cb *CausalBuilder) GetLastNGraphs(n int) []*CausalGraph {
	cb.completedMu.RLock()
	defer cb.completedMu.RUnlock()

	count := len(cb.completedGraphs)
	if n > count {
		n = count
	}

	graphs := make([]*CausalGraph, n)
	copy(graphs, cb.completedGraphs[count-n:])
	return graphs
}

// addCompletedGraph adds a completed graph to the history.
func (cb *CausalBuilder) addCompletedGraph(graph *CausalGraph) {
	cb.completedMu.Lock()
	defer cb.completedMu.Unlock()

	cb.completedGraphs = append(cb.completedGraphs, graph)

	// Keep only maxGraphs
	if len(cb.completedGraphs) > cb.maxGraphs {
		// Remove oldest
		cb.completedGraphs = cb.completedGraphs[1:]
	}
}

// Clear clears all completed graphs.
func (cb *CausalBuilder) Clear() {
	cb.completedMu.Lock()
	defer cb.completedMu.Unlock()

	cb.completedGraphs = make([]*CausalGraph, 0, cb.maxGraphs)
}

// SetMaxGraphs sets the maximum number of graphs to keep.
func (cb *CausalBuilder) SetMaxGraphs(n int) {
	cb.completedMu.Lock()
	defer cb.completedMu.Unlock()

	cb.maxGraphs = n

	// Trim if necessary
	if len(cb.completedGraphs) > n {
		cb.completedGraphs = cb.completedGraphs[len(cb.completedGraphs)-n:]
	}
}

// GetStats returns statistics about the causal builder.
func (cb *CausalBuilder) GetStats() *CausalBuilderStats {
	current := cb.currentGraph.Load()
	summary := (*FrameSummary)(nil)
	if current != nil {
		summary = current.GetFrameSummary()
	}

	cb.completedMu.RLock()
	completedCount := len(cb.completedGraphs)
	cb.completedMu.RUnlock()

	return &CausalBuilderStats{
		Enabled:         cb.IsEnabled(),
		CurrentSummary:  summary,
		CompletedGraphs: completedCount,
		MaxGraphs:       cb.maxGraphs,
	}
}

// CausalBuilderStats contains statistics about the causal builder.
type CausalBuilderStats struct {
	Enabled         bool
	CurrentSummary  *FrameSummary
	CompletedGraphs int
	MaxGraphs       int
}
