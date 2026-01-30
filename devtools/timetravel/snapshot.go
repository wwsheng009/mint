// Package timetravel provides time travel capabilities for DevTools.
//
// This file implements frame snapshots that capture complete state
// for time travel debugging.
package timetravel

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// SnapshotManager manages frame snapshots for time travel.
type SnapshotManager struct {
	mu        sync.RWMutex
	snapshots []*FrameSnapshot
	maxCount  int

	// Index for fast lookup
	frameIndex map[devtools.FrameID]int
}

// FrameSnapshot represents a complete snapshot of a frame's state.
type FrameSnapshot struct {
	// Metadata
	FrameID    devtools.FrameID
	Timestamp  time.Time
	PrevFrame  *FrameSnapshot // Linked list for traversal

	// Causal data
	CausalGraph *devtools.CausalGraph

	// Component state
	ComponentStates map[uint32]*ComponentState

	// Layout state
	LayoutState *LayoutSnapshot

	// Repaint state
	RepaintState *RepaintSnapshot

	// Event data
	Events []devtools.CausalEvent
}

// ComponentState represents the complete state of a component.
type ComponentState struct {
	ComponentID   uint32
	ComponentName string
	State         map[string]interface{}
	Props         map[string]interface{}
	Style         map[string]interface{}
	Children      []uint32
}

// LayoutSnapshot represents the complete layout state.
type LayoutSnapshot struct {
	Nodes map[string]*NodeLayout
	Root  string
}

// NodeLayout represents layout information for a single node.
type NodeLayout struct {
	ID          string
	Type        string
	X, Y        int
	Width, Height int
	ZIndex      int
	Visible     bool
	FlexGrow    float32
	FlexShrink  float32
	Parent      string
	Children    []string
}

// RepaintSnapshot represents repaint information.
type RepaintSnapshot struct {
	DirtyRegions []devtools.Rect
	ChangedCells int
	TotalCells   int
	Buffer       []byte // Optional: full buffer snapshot
}

// NewSnapshotManager creates a new snapshot manager.
func NewSnapshotManager(maxCount int) *SnapshotManager {
	return &SnapshotManager{
		snapshots:  make([]*FrameSnapshot, 0, maxCount),
		maxCount:   maxCount,
		frameIndex: make(map[devtools.FrameID]int),
	}
}

// AddSnapshot adds a new frame snapshot.
func (sm *SnapshotManager) AddSnapshot(snapshot *FrameSnapshot) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Link to previous frame
	if len(sm.snapshots) > 0 {
		snapshot.PrevFrame = sm.snapshots[len(sm.snapshots)-1]
	}

	sm.snapshots = append(sm.snapshots, snapshot)
	sm.frameIndex[snapshot.FrameID] = len(sm.snapshots) - 1

	// Trim to max count
	if len(sm.snapshots) > sm.maxCount {
		removed := sm.snapshots[0]
		delete(sm.frameIndex, removed.FrameID)
		sm.snapshots = sm.snapshots[1:]
	}
}

// GetSnapshot returns a snapshot by frame ID.
func (sm *SnapshotManager) GetSnapshot(frameID devtools.FrameID) *FrameSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if idx, ok := sm.frameIndex[frameID]; ok {
		return sm.snapshots[idx]
	}
	return nil
}

// GetSnapshotByIndex returns a snapshot by index.
func (sm *SnapshotManager) GetSnapshotByIndex(index int) *FrameSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if index < 0 || index >= len(sm.snapshots) {
		return nil
	}
	return sm.snapshots[index]
}

// GetLatestSnapshot returns the most recent snapshot.
func (sm *SnapshotManager) GetLatestSnapshot() *FrameSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.snapshots) == 0 {
		return nil
	}
	return sm.snapshots[len(sm.snapshots)-1]
}

// GetAllSnapshots returns all snapshots.
func (sm *SnapshotManager) GetAllSnapshots() []*FrameSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	snapshots := make([]*FrameSnapshot, len(sm.snapshots))
	copy(snapshots, sm.snapshots)
	return snapshots
}

// GetSnapshotsInRange returns snapshots in a frame ID range.
func (sm *SnapshotManager) GetSnapshotsInRange(startID, endID devtools.FrameID) []*FrameSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*FrameSnapshot
	for _, snapshot := range sm.snapshots {
		if snapshot.FrameID >= startID && snapshot.FrameID <= endID {
			result = append(result, snapshot)
		}
	}
	return result
}

// Clear removes all snapshots.
func (sm *SnapshotManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.snapshots = make([]*FrameSnapshot, 0, sm.maxCount)
	sm.frameIndex = make(map[devtools.FrameID]int)
}

// Count returns the number of snapshots.
func (sm *SnapshotManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.snapshots)
}

// GetFrameIDs returns all frame IDs.
func (sm *SnapshotManager) GetFrameIDs() []devtools.FrameID {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]devtools.FrameID, len(sm.snapshots))
	for i, snapshot := range sm.snapshots {
		ids[i] = snapshot.FrameID
	}
	return ids
}

// ToJSON converts a snapshot to JSON for serialization.
func (fs *FrameSnapshot) ToJSON() ([]byte, error) {
	return json.Marshal(fs)
}

// FromJSON creates a snapshot from JSON.
func FromJSON(data []byte) (*FrameSnapshot, error) {
	var snapshot FrameSnapshot
	err := json.Unmarshal(data, &snapshot)
	return &snapshot, err
}

// SnapshotBuilder helps build frame snapshots.
type SnapshotBuilder struct {
	current *FrameSnapshot
	mgr     *SnapshotManager
}

// NewSnapshotBuilder creates a new snapshot builder.
func NewSnapshotBuilder(mgr *SnapshotManager) *SnapshotBuilder {
	return &SnapshotBuilder{mgr: mgr}
}

// BeginSnapshot starts building a new snapshot.
func (sb *SnapshotBuilder) BeginSnapshot(frameID devtools.FrameID) *SnapshotBuilder {
	sb.current = &FrameSnapshot{
		FrameID:         frameID,
		Timestamp:       time.Now(),
		ComponentStates: make(map[uint32]*ComponentState),
	}
	return sb
}

// WithCausalGraph sets the causal graph for the snapshot.
func (sb *SnapshotBuilder) WithCausalGraph(graph *devtools.CausalGraph) *SnapshotBuilder {
	if sb.current != nil {
		sb.current.CausalGraph = graph
	}
	return sb
}

// WithComponentState adds a component state to the snapshot.
func (sb *SnapshotBuilder) WithComponentState(state *ComponentState) *SnapshotBuilder {
	if sb.current != nil {
		sb.current.ComponentStates[state.ComponentID] = state
	}
	return sb
}

// WithLayoutState sets the layout state for the snapshot.
func (sb *SnapshotBuilder) WithLayoutState(layout *LayoutSnapshot) *SnapshotBuilder {
	if sb.current != nil {
		sb.current.LayoutState = layout
	}
	return sb
}

// WithRepaintState sets the repaint state for the snapshot.
func (sb *SnapshotBuilder) WithRepaintState(repaint *RepaintSnapshot) *SnapshotBuilder {
	if sb.current != nil {
		sb.current.RepaintState = repaint
	}
	return sb
}

// WithEvents adds events to the snapshot.
func (sb *SnapshotBuilder) WithEvents(events []devtools.CausalEvent) *SnapshotBuilder {
	if sb.current != nil {
		sb.current.Events = events
	}
	return sb
}

// Build finalizes and adds the snapshot to the manager.
func (sb *SnapshotBuilder) Build() *FrameSnapshot {
	if sb.current != nil && sb.mgr != nil {
		sb.mgr.AddSnapshot(sb.current)
		result := sb.current
		sb.current = nil
		return result
	}
	return nil
}

// GetCurrent returns the current snapshot being built.
func (sb *SnapshotBuilder) GetCurrent() *FrameSnapshot {
	return sb.current
}

// SnapshotDiff represents differences between two snapshots.
type SnapshotDiff struct {
	FromFrame   devtools.FrameID
	ToFrame     devtools.FrameID

	// Changed components
	AddedComponents   []uint32
	RemovedComponents []uint32
	ChangedComponents []ComponentDiff

	// Layout changes
	LayoutChanges []LayoutDiff

	// Causal changes
	NewEvents     []*devtools.CausalEvent
	NewMutations  []*devtools.CausalMutation
	NewLayouts    []*devtools.CausalLayout
	NewRepaints   []*devtools.CausalRepaint
}

// ComponentDiff represents differences in a component.
type ComponentDiff struct {
	ComponentID uint32
	Changes     StateChanges
}

// StateChanges represents state changes.
type StateChanges struct {
	Added    map[string]interface{}
	Removed  []string
	Modified map[string]ValueChange
}

// ValueChange represents a value change.
type ValueChange struct {
	OldValue interface{}
	NewValue interface{}
}

// LayoutDiff represents layout changes for a node.
type LayoutDiff struct {
	NodeID     string
	ChangeMask devtools.ChangeMask
	OldRect    *devtools.Rect
	NewRect    *devtools.Rect
}

// Diff computes the difference between two snapshots.
func (fs *FrameSnapshot) Diff(other *FrameSnapshot) *SnapshotDiff {
	if other == nil {
		return nil
	}

	diff := &SnapshotDiff{
		FromFrame: fs.FrameID,
		ToFrame:   other.FrameID,
	}

	// Compare component states
	diff.computeComponentDiff(fs, other)

	// Compare layout states
	diff.computeLayoutDiff(fs, other)

	// Compare causal data
	diff.computeCausalDiff(fs, other)

	return diff
}

// computeComponentDiff computes component state differences.
func (sd *SnapshotDiff) computeComponentDiff(from, to *FrameSnapshot) {
	// Find added and removed components
	fromComponents := make(map[uint32]bool)
	for id := range from.ComponentStates {
		fromComponents[id] = true
	}

	for id := range to.ComponentStates {
		if !fromComponents[id] {
			sd.AddedComponents = append(sd.AddedComponents, id)
		}
	}

	for id := range from.ComponentStates {
		if _, exists := to.ComponentStates[id]; !exists {
			sd.RemovedComponents = append(sd.RemovedComponents, id)
		}
	}

	// Find changed components
	for id, toState := range to.ComponentStates {
		if fromState, exists := from.ComponentStates[id]; exists {
			if compDiff := compareComponentStates(fromState, toState); compDiff != nil {
				sd.ChangedComponents = append(sd.ChangedComponents, *compDiff)
			}
		}
	}
}

// compareComponentStates compares two component states.
func compareComponentStates(from, to *ComponentState) *ComponentDiff {
	changes := StateChanges{
		Added:    make(map[string]interface{}),
		Modified: make(map[string]ValueChange),
	}

	// Compare state
	for key, fromVal := range from.State {
		if toVal, exists := to.State[key]; exists {
			if !valuesEqual(fromVal, toVal) {
				changes.Modified[key] = ValueChange{
					OldValue: fromVal,
					NewValue: toVal,
				}
			}
		} else {
			changes.Removed = append(changes.Removed, key)
		}
	}

	for key, toVal := range to.State {
		if _, exists := from.State[key]; !exists {
			changes.Added[key] = toVal
		}
	}

	if len(changes.Added) == 0 && len(changes.Removed) == 0 && len(changes.Modified) == 0 {
		return nil
	}

	return &ComponentDiff{
		ComponentID: from.ComponentID,
		Changes:     changes,
	}
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b interface{}) bool {
	// Simple equality check
	// For more complex types, use reflection or custom comparators
	return a == b
}

// computeLayoutDiff computes layout differences.
func (sd *SnapshotDiff) computeLayoutDiff(from, to *FrameSnapshot) {
	if from.LayoutState == nil || to.LayoutState == nil {
		return
	}

	// Compare all nodes in the target layout
	for nodeID, toNode := range to.LayoutState.Nodes {
		if fromNode, exists := from.LayoutState.Nodes[nodeID]; exists {
			// Check for changes
			changeMask := devtools.ChangeMask(0)
			oldRect := &devtools.Rect{
				X:      fromNode.X,
				Y:      fromNode.Y,
				Width:  fromNode.Width,
				Height: fromNode.Height,
			}
			newRect := &devtools.Rect{
				X:      toNode.X,
				Y:      toNode.Y,
				Width:  toNode.Width,
				Height: toNode.Height,
			}

			if fromNode.X != toNode.X || fromNode.Y != toNode.Y ||
				fromNode.Width != toNode.Width || fromNode.Height != toNode.Height {
				changeMask |= devtools.ChangeRect
			}

			if fromNode.Visible != toNode.Visible {
				changeMask |= devtools.ChangeVisibility
			}

			if changeMask != 0 {
				sd.LayoutChanges = append(sd.LayoutChanges, LayoutDiff{
					NodeID:     nodeID,
					ChangeMask: changeMask,
					OldRect:    oldRect,
					NewRect:    newRect,
				})
			}
		}
	}
}

// computeCausalDiff computes causal graph differences.
func (sd *SnapshotDiff) computeCausalDiff(from, to *FrameSnapshot) {
	if from.CausalGraph == nil || to.CausalGraph == nil {
		return
	}

	// New events in the target snapshot
	sd.NewEvents = to.CausalGraph.Events
	sd.NewMutations = to.CausalGraph.Mutations
	sd.NewLayouts = to.CausalGraph.Layouts
	sd.NewRepaints = to.CausalGraph.Repaints
}
