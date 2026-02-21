package incremental

import (
	"fmt"
	"sync"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Incremental Layout Tracker
// =============================================================================

// DirtyFlag indicates whether a node needs layout recalculation.
type DirtyFlag int

const (
	// Clean means the node doesn't need relayout
	Clean DirtyFlag = iota
	// Dirty means the node needs layout recalculation
	Dirty
	// Propagate means the node's children may be affected
	Propagate
)

// ChangeType indicates the type of change that caused dirty state.
type ChangeType int

const (
	// ChangeNone means no change
	ChangeNone ChangeType = iota
	// ChangeProps means properties changed
	ChangeProps
	// ChangeChildren means children changed
	ChangeChildren
	// ChangeContent means content changed
	ChangeContent
	// ChangeDimension means dimensions changed
	ChangeDimension
)

// LayoutChange represents a change in the layout tree.
type LayoutChange struct {
	Node    ui.VNode
	Type    ChangeType
	OldSize layout.Size
	NewSize layout.Size
}

// IncrementalLayout tracks dirty nodes for incremental layout updates.
type IncrementalLayout struct {
	mu         sync.RWMutex
	dirty      map[string]DirtyFlag
	changes    map[string][]LayoutChange
	versions   map[string]int
	nextID     int
}

// NewIncrementalLayout creates a new incremental layout tracker.
func NewIncrementalLayout() *IncrementalLayout {
	return &IncrementalLayout{
		dirty:    make(map[string]DirtyFlag),
		changes:  make(map[string][]LayoutChange),
		versions: make(map[string]int),
		nextID:   1,
	}
}

// MarkDirty marks a node as needing layout recalculation.
func (il *IncrementalLayout) MarkDirty(node ui.VNode, flag DirtyFlag, change LayoutChange) {
	if il == nil || node == nil {
		return
	}

	key := node.Key()
	if key == "" {
		key = il.generateKey(node)
	}

	il.mu.Lock()
	defer il.mu.Unlock()

	// Update dirty flag
	// If already Propagate, keep it
	if il.dirty[key] != Propagate {
		il.dirty[key] = flag
	}

	// Record change
	il.changes[key] = append(il.changes[key], change)

	// Increment version for this node
	il.versions[key]++
}

// IsDirty checks if a node needs layout recalculation.
func (il *IncrementalLayout) IsDirty(node ui.VNode) bool {
	if il == nil || node == nil {
		return false
	}

	key := node.Key()
	if key == "" {
		key = il.generateKey(node)
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	flag, exists := il.dirty[key]
	return exists && flag != Clean
}

// GetDirtyFlag returns the dirty flag for a node.
func (il *IncrementalLayout) GetDirtyFlag(node ui.VNode) DirtyFlag {
	if il == nil || node == nil {
		return Clean
	}

	key := node.Key()
	if key == "" {
		key = il.generateKey(node)
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	flag, exists := il.dirty[key]
	if !exists {
		return Clean
	}
	return flag
}

// MarkClean marks a node as clean (no relayout needed).
func (il *IncrementalLayout) MarkClean(node ui.VNode) {
	if il == nil || node == nil {
		return
	}

	key := node.Key()
	if key == "" {
		key = il.generateKey(node)
	}

	il.mu.Lock()
	defer il.mu.Unlock()

	il.dirty[key] = Clean
	// Note: we keep the changes and versions for history/debugging
}

// PropagateDirty marks ancestors as dirty when a child changes size.
// This is needed because a size change in a child may affect parent layout.
func (il *IncrementalLayout) PropagateDirty(child ui.VNode, childSize layout.Size) {
	if il == nil || child == nil {
		return
	}

	// For now, we'll just mark the child as Propagate
	// In a full implementation, this would walk up the tree
	il.MarkDirty(child, Propagate, LayoutChange{
		Node:    child,
		Type:    ChangeDimension,
		OldSize: layout.Size{}, // Unknown
		NewSize: childSize,
	})
}

// GetChanges returns all changes recorded for a node.
func (il *IncrementalLayout) GetChanges(node ui.VNode) []LayoutChange {
	if il == nil || node == nil {
		return nil
	}

	key := node.Key()
	if key == "" {
		key = il.generateKey(node)
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	changes := il.changes[key]
	result := make([]LayoutChange, len(changes))
	copy(result, changes)
	return result
}

// ClearChanges clears all changes for a node.
func (il *IncrementalLayout) ClearChanges(node ui.VNode) {
	if il == nil || node == nil {
		return
	}

	key := node.Key()
	if key == "" {
		key = il.generateKey(node)
	}

	il.mu.Lock()
	defer il.mu.Unlock()

	il.changes[key] = nil
}

// GetVersion returns the current version of a node.
func (il *IncrementalLayout) GetVersion(node ui.VNode) int {
	if il == nil || node == nil {
		return 0
	}

	key := node.Key()
	if key == "" {
		key = il.generateKey(node)
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	return il.versions[key]
}

// GetDirtyNodes returns a list of all nodes that are dirty.
func (il *IncrementalLayout) GetDirtyNodes() []string {
	if il == nil {
		return nil
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	dirtyNodes := make([]string, 0, len(il.dirty))
	for key, flag := range il.dirty {
		if flag != Clean {
			dirtyNodes = append(dirtyNodes, key)
		}
	}
	return dirtyNodes
}

// GetDirtyNodesByFlag returns dirty nodes with a specific flag.
func (il *IncrementalLayout) GetDirtyNodesByFlag(flag DirtyFlag) []string {
	if il == nil {
		return nil
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	dirtyNodes := make([]string, 0)
	for key, f := range il.dirty {
		if f == flag {
			dirtyNodes = append(dirtyNodes, key)
		}
	}
	return dirtyNodes
}

// Clear clears all tracking state.
func (il *IncrementalLayout) Clear() {
	if il == nil {
		return
	}

	il.mu.Lock()
	defer il.mu.Unlock()

	il.dirty = make(map[string]DirtyFlag)
	il.changes = make(map[string][]LayoutChange)
	il.versions = make(map[string]int)
}

// Stats returns statistics about the incremental layout state.
func (il *IncrementalLayout) Stats() LayoutStats {
	if il == nil {
		return LayoutStats{}
	}

	il.mu.RLock()
	defer il.mu.RUnlock()

	dirtyCount := 0
	propagateCount := 0
	totalChanges := 0
	maxVersion := 0

	for key, flag := range il.dirty {
		if flag == Dirty {
			dirtyCount++
		} else if flag == Propagate {
			propagateCount++
		}
		totalChanges += len(il.changes[key])
		if il.versions[key] > maxVersion {
			maxVersion = il.versions[key]
		}
	}

	return LayoutStats{
		TotalNodes:   len(il.dirty),
		DirtyCount:   dirtyCount,
		PropagateCount: propagateCount,
		TotalChanges: totalChanges,
		MaxVersion:   maxVersion,
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func (il *IncrementalLayout) generateKey(node ui.VNode) string {
	if node == nil {
		return fmt.Sprintf("gen-%d", il.nextID)
	}
	key := fmt.Sprintf("%s-%s", node.Type(), node.Tag())
	il.nextID++
	return key
}

// =============================================================================
// Layout Stats
// =============================================================================

// LayoutStats contains statistics about the incremental layout tracker.
type LayoutStats struct {
	TotalNodes    int
	DirtyCount    int
	PropagateCount int
	TotalChanges  int
	MaxVersion    int
}

// String returns a string representation of the stats.
func (s LayoutStats) String() string {
	return fmt.Sprintf("nodes=%d dirty=%d propagate=%d changes=%d max_version=%d",
		s.TotalNodes, s.DirtyCount, s.PropagateCount, s.TotalChanges, s.MaxVersion)
}

// =============================================================================
// LayoutContext - Combines Incremental Layout with Cache
// =============================================================================

// LayoutContext provides a unified context for incremental layout operations.
type LayoutContext struct {
	Incremental *IncrementalLayout
}

// NewLayoutContext creates a new layout context.
func NewLayoutContext() *LayoutContext {
	return &LayoutContext{
		Incremental: NewIncrementalLayout(),
	}
}

// NeedsLayout checks if a node needs layout recalculation.
func (lc *LayoutContext) NeedsLayout(node ui.VNode) bool {
	if lc == nil || lc.Incremental == nil {
		return true // Default to full layout
	}
	return lc.Incremental.IsDirty(node)
}

// MarkNodeChanged marks a node as changed.
func (lc *LayoutContext) MarkNodeChanged(node ui.VNode, changeType ChangeType, oldSize, newSize layout.Size) {
	if lc == nil || lc.Incremental == nil {
		return
	}

	flag := Dirty
	if changeType == ChangeDimension {
		flag = Propagate
	}

	lc.Incremental.MarkDirty(node, flag, LayoutChange{
		Node:    node,
		Type:    changeType,
		OldSize: oldSize,
		NewSize: newSize,
	})
}

// MarkChildrenChanged marks that children of a node changed.
func (lc *LayoutContext) MarkChildrenChanged(node ui.VNode) {
	lc.MarkNodeChanged(node, ChangeChildren, layout.Size{}, layout.Size{})
}

// MarkPropsChanged marks that properties of a node changed.
func (lc *LayoutContext) MarkPropsChanged(node ui.VNode) {
	lc.MarkNodeChanged(node, ChangeProps, layout.Size{}, layout.Size{})
}

// MarkContentChanged marks that content of a node changed.
func (lc *LayoutContext) MarkContentChanged(node ui.VNode) {
	lc.MarkNodeChanged(node, ChangeContent, layout.Size{}, layout.Size{})
}

// MarkSizeChanged marks that the size of a node changed.
func (lc *LayoutContext) MarkSizeChanged(node ui.VNode, oldSize, newSize layout.Size) {
	lc.MarkNodeChanged(node, ChangeDimension, oldSize, newSize)
}

// FinishLayout marks a node as done with layout.
func (lc *LayoutContext) FinishLayout(node ui.VNode) {
	if lc == nil || lc.Incremental == nil {
		return
	}
	lc.Incremental.MarkClean(node)
}

// GetNodeVersion returns the version of a node.
func (lc *LayoutContext) GetNodeVersion(node ui.VNode) int {
	if lc == nil || lc.Incremental == nil {
		return 0
	}
	return lc.Incremental.GetVersion(node)
}

// GetStats returns statistics about the layout state.
func (lc *LayoutContext) GetStats() LayoutContextStats {
	if lc == nil || lc.Incremental == nil {
		return LayoutContextStats{}
	}

	iStats := lc.Incremental.Stats()
	return LayoutContextStats{
		LayoutStats: iStats,
	}
}

// Clear resets the layout context.
func (lc *LayoutContext) Clear() {
	if lc == nil || lc.Incremental == nil {
		return
	}
	lc.Incremental.Clear()
}

// =============================================================================
// Layout Context Stats
// =============================================================================

// LayoutContextStats contains statistics about the layout context.
type LayoutContextStats struct {
	LayoutStats
}

// String returns a string representation of the stats.
func (s LayoutContextStats) String() string {
	return s.LayoutStats.String()
}
