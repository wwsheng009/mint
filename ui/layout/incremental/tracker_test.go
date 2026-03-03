package incremental

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
	"github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Mock VNode for Testing
// =============================================================================

type mockVNode struct {
	key  string
	tag  string
	props ui.Props
}

func (m *mockVNode) Key() string {
	if m != nil {
		return m.key
	}
	return ""
}

func (m *mockVNode) SetKey(k string) ui.VNode {
	if m != nil {
		m.key = k
		return m
	}
	return nil
}

func (m *mockVNode) Tag() string {
	if m != nil {
		return m.tag
	}
	return ""
}

func (m *mockVNode) Type() ui.VNodeType {
	return ui.VNodeElement
}

func (m *mockVNode) Props() ui.Props {
	if m != nil {
		return m.props
	}
	return nil
}

func (m *mockVNode) SetProps(p ui.Props) ui.VNode {
	if m != nil {
		m.props = p
		return m
	}
	return nil
}

func (m *mockVNode) Children() []ui.VNode {
	return nil
}

func (m *mockVNode) SetChildren(children []ui.VNode) ui.VNode {
	return nil
}

func (m *mockVNode) Style() style.Style {
	return style.Style{}
}

func (m *mockVNode) SetStyle(s style.Style) ui.VNode {
	return nil
}

func (m *mockVNode) GetLayer() types.Layer {
	return types.LayerBase
}

func (m *mockVNode) SetLayer(l types.Layer) ui.VNode {
	return m
}

func newMockVNode(key string) *mockVNode {
	return &mockVNode{
		key:   key,
		tag:   "mock",
		props: make(ui.Props),
	}
}

// =============================================================================
// IncrementalLayout Core Tests
// =============================================================================

func TestIncrementalLayout_New(t *testing.T) {
	il := NewIncrementalLayout()

	if il == nil {
		t.Fatal("NewIncrementalLayout() should not return nil")
	}

	stats := il.Stats()
	if stats.TotalNodes != 0 {
		t.Errorf("expected 0 total nodes, got %d", stats.TotalNodes)
	}
}

func TestIncrementalLayout_MarkDirty(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("dirty-test-node")

	il.MarkDirty(node, Dirty, LayoutChange{})

	if !il.IsDirty(node) {
		t.Error("node should be marked as dirty")
	}

	flag := il.GetDirtyFlag(node)
	if flag != Dirty {
		t.Errorf("expected Dirty flag, got %d", flag)
	}
}

func TestIncrementalLayout_MarkDirty_Propagate(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("propagate-node")

	il.MarkDirty(node, Propagate, LayoutChange{})

	flag := il.GetDirtyFlag(node)
	if flag != Propagate {
		t.Errorf("expected Propagate flag, got %d", flag)
	}
}

func TestIncrementalLayout_IsDirty_False(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("clean-node")

	if il.IsDirty(node) {
		t.Error("node should not be dirty initially")
	}
}

func TestIncrementalLayout_MarkClean(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("clean-test-node")

	// Mark dirty first
	il.MarkDirty(node, Dirty, LayoutChange{})
	if !il.IsDirty(node) {
		t.Fatal("node should be dirty after markDirty")
	}

	// Mark clean
	il.MarkClean(node)
	if il.IsDirty(node) {
		t.Error("node should be clean after markClean")
	}
}

func TestIncrementalLayout_PropagateDirty(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("propagate-test-node")
	size := layout.Size{Width: 30, Height: 10}

	il.PropagateDirty(node, size)

	if !il.IsDirty(node) {
		t.Error("node should be dirty after PropagateDirty")
	}

	flag := il.GetDirtyFlag(node)
	if flag != Propagate {
		t.Errorf("expected Propagate flag, got %d", flag)
	}

	changes := il.GetChanges(node)
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Type != ChangeDimension {
		t.Errorf("expected ChangeDimension, got %d", changes[0].Type)
	}
}

// =============================================================================
// Change Tracking Tests
// =============================================================================

func TestIncrementalLayout_GetChanges(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("changes-node")

	change1 := LayoutChange{
		Node:    node,
		Type:    ChangeProps,
		OldSize: layout.Size{Width: 10, Height: 5},
		NewSize: layout.Size{Width: 20, Height: 10},
	}
	change2 := LayoutChange{
		Node:    node,
		Type:    ChangeContent,
		OldSize: layout.Size{Width: 20, Height: 10},
		NewSize: layout.Size{Width: 25, Height: 12},
	}

	il.MarkDirty(node, Dirty, change1)
	il.MarkDirty(node, Dirty, change2)

	changes := il.GetChanges(node)
	if len(changes) != 2 {
		t.Errorf("expected 2 changes, got %d", len(changes))
	}
	if changes[0].Type != ChangeProps {
		t.Errorf("expected first change to be ChangeProps, got %d", changes[0].Type)
	}
	if changes[1].Type != ChangeContent {
		t.Errorf("expected second change to be ChangeContent, got %d", changes[1].Type)
	}
}

func TestIncrementalLayout_ClearChanges(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("clear-changes-node")

	il.MarkDirty(node, Dirty, LayoutChange{Type: ChangeProps})

	// Verify change exists
	changes := il.GetChanges(node)
	if len(changes) != 1 {
		t.Fatal("expected 1 change before clear")
	}

	// Clear changes
	il.ClearChanges(node)

	// Verify changes cleared
	changes = il.GetChanges(node)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes after clear, got %d", len(changes))
	}
}

// =============================================================================
// Version Tests
// =============================================================================

func TestIncrementalLayout_GetVersion(t *testing.T) {
	il := NewIncrementalLayout()
	node := newMockVNode("version-node")

	// Initial version should be 0
	if v := il.GetVersion(node); v != 0 {
		t.Errorf("expected initial version 0, got %d", v)
	}

	// After marking dirty, version should increase
	il.MarkDirty(node, Dirty, LayoutChange{})
	if v := il.GetVersion(node); v != 1 {
		t.Errorf("expected version 1 after first markDirty, got %d", v)
	}

	// Second markDirty should increase version again
	il.MarkDirty(node, Dirty, LayoutChange{})
	if v := il.GetVersion(node); v != 2 {
		t.Errorf("expected version 2 after second markDirty, got %d", v)
	}
}

// =============================================================================
// Dirty Node Listing Tests
// =============================================================================

func TestIncrementalLayout_GetDirtyNodes(t *testing.T) {
	il := NewIncrementalLayout()

	node1 := newMockVNode("dirty-node-1")
	node2 := newMockVNode("dirty-node-2")
	_ = newMockVNode("clean-node")

	// Mark two nodes as dirty
	il.MarkDirty(node1, Dirty, LayoutChange{})
	il.MarkDirty(node2, Dirty, LayoutChange{})

	// Get dirty nodes
	dirtyNodes := il.GetDirtyNodes()

	if len(dirtyNodes) != 2 {
		t.Errorf("expected 2 dirty nodes, got %d", len(dirtyNodes))
	}
}

func TestIncrementalLayout_GetDirtyNodesByFlag(t *testing.T) {
	il := NewIncrementalLayout()

	node1 := newMockVNode("dirty-flag-node-1")
	node2 := newMockVNode("propagate-flag-node")
	node3 := newMockVNode("dirty-flag-node-2")

	il.MarkDirty(node1, Dirty, LayoutChange{})
	il.MarkDirty(node2, Propagate, LayoutChange{})
	il.MarkDirty(node3, Dirty, LayoutChange{})

	// Get Dirty nodes
	dirtyNodes := il.GetDirtyNodesByFlag(Dirty)
	if len(dirtyNodes) != 2 {
		t.Errorf("expected 2 Dirty nodes, got %d", len(dirtyNodes))
	}

	// Get Propagate nodes
	propagateNodes := il.GetDirtyNodesByFlag(Propagate)
	if len(propagateNodes) != 1 {
		t.Errorf("expected 1 Propagate node, got %d", len(propagateNodes))
	}
}

// =============================================================================
// Stats Tests
// =============================================================================

func TestIncrementalLayout_Stats(t *testing.T) {
	il := NewIncrementalLayout()

	// Initial stats
	stats := il.Stats()
	if stats.TotalNodes != 0 || stats.DirtyCount != 0 || stats.TotalChanges != 0 {
		t.Errorf("expected empty stats, got %v", stats)
	}

	// Add some dirty nodes
	il.MarkDirty(newMockVNode("stats-node-1"), Dirty, LayoutChange{Type: ChangeProps})
	il.MarkDirty(newMockVNode("stats-node-2"), Propagate, LayoutChange{Type: ChangeChildren})
	il.MarkDirty(newMockVNode("stats-node-3"), Dirty, LayoutChange{Type: ChangeDimension})

	// Check stats
	stats = il.Stats()
	if stats.TotalNodes != 3 {
		t.Errorf("expected 3 total nodes, got %d", stats.TotalNodes)
	}
	if stats.DirtyCount != 2 {
		t.Errorf("expected 2 dirty nodes, got %d", stats.DirtyCount)
	}
	if stats.PropagateCount != 1 {
		t.Errorf("expected 1 propagate node, got %d", stats.PropagateCount)
	}
	if stats.TotalChanges != 3 {
		t.Errorf("expected 3 total changes, got %d", stats.TotalChanges)
	}
	if stats.MaxVersion != 1 {
		t.Errorf("expected max version 1, got %d", stats.MaxVersion)
	}
}

func TestLayoutStats_String(t *testing.T) {
	stats := LayoutStats{
		TotalNodes:    10,
		DirtyCount:    5,
		PropagateCount: 2,
		TotalChanges:  8,
		MaxVersion:    15,
	}

	str := stats.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}
}

// =============================================================================
// Clear Tests
// =============================================================================

func TestIncrementalLayout_Clear(t *testing.T) {
	il := NewIncrementalLayout()

	// Add some data
	il.MarkDirty(newMockVNode("clear-node-1"), Dirty, LayoutChange{})
	il.MarkDirty(newMockVNode("clear-node-2"), Propagate, LayoutChange{})

	// Verify not empty
	stats := il.Stats()
	if stats.TotalNodes == 0 {
		t.Fatal("expected non-empty stats before clear")
	}

	// Clear
	il.Clear()

	// Verify empty
	stats = il.Stats()
	if stats.TotalNodes != 0 || stats.DirtyCount != 0 {
		t.Errorf("expected empty stats after clear, got %v", stats)
	}
}

// =============================================================================
// Nil Handling Tests
// =============================================================================

func TestIncrementalLayout_Nil_Handling(t *testing.T) {
	var il *IncrementalLayout = nil
	var node ui.VNode = nil

	// Should not panic
	il.MarkDirty(node, Dirty, LayoutChange{})
	il.IsDirty(node)
	il.GetDirtyFlag(node)
	il.MarkClean(node)
	il.PropagateDirty(node, layout.Size{})
	il.GetChanges(node)
	il.ClearChanges(node)
	il.GetVersion(node)
	il.GetDirtyNodes()
	il.GetDirtyNodesByFlag(Dirty)
	il.Clear()
	il.Stats()
}

// =============================================================================
// LayoutContext Tests
// =============================================================================

func TestLayoutContext_New(t *testing.T) {
	lc := NewLayoutContext()

	if lc == nil {
		t.Fatal("NewLayoutContext() should not return nil")
	}
	if lc.Incremental == nil {
		t.Error("Incremental should not be nil")
	}
}

func TestLayoutContext_NeedsLayout(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("needs-layout-node")

	if lc.NeedsLayout(node) {
		t.Error("node should not need layout initially")
	}

	lc.MarkPropsChanged(node)

	if !lc.NeedsLayout(node) {
		t.Error("node should need layout after markPropsChanged")
	}
}

func TestLayoutContext_MarkNodeChanged(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("mark-node-changed")

	oldSize := layout.Size{Width: 10, Height: 5}
	newSize := layout.Size{Width: 20, Height: 10}

	lc.MarkNodeChanged(node, ChangeDimension, oldSize, newSize)

	changes := lc.Incremental.GetChanges(node)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	change := changes[0]
	if change.Type != ChangeDimension {
		t.Errorf("expected ChangeDimension, got %d", change.Type)
	}
	if change.OldSize != oldSize {
		t.Errorf("old size mismatch")
	}
	if change.NewSize != newSize {
		t.Errorf("new size mismatch")
	}
}

func TestLayoutContext_MarkChildrenChanged(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("mark-children-changed")

	lc.MarkChildrenChanged(node)

	if !lc.NeedsLayout(node) {
		t.Error("node should need layout after MarkChildrenChanged")
	}

	flag := lc.Incremental.GetDirtyFlag(node)
	if flag != Dirty {
		t.Errorf("expected Dirty flag, got %d", flag)
	}
}

func TestLayoutContext_MarkPropsChanged(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("mark-props-changed")

	lc.MarkPropsChanged(node)

	changes := lc.Incremental.GetChanges(node)
	if len(changes) != 1 {
		t.Fatal("expected 1 change")
	}
	if changes[0].Type != ChangeProps {
		t.Errorf("expected ChangeProps, got %d", changes[0].Type)
	}
}

func TestLayoutContext_MarkContentChanged(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("mark-content-changed")

	lc.MarkContentChanged(node)

	changes := lc.Incremental.GetChanges(node)
	if len(changes) != 1 {
		t.Fatal("expected 1 change")
	}
	if changes[0].Type != ChangeContent {
		t.Errorf("expected ChangeContent, got %d", changes[0].Type)
	}
}

func TestLayoutContext_MarkSizeChanged(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("mark-size-changed")

	oldSize := layout.Size{Width: 10, Height: 5}
	newSize := layout.Size{Width: 20, Height: 10}

	lc.MarkSizeChanged(node, oldSize, newSize)

	changes := lc.Incremental.GetChanges(node)
	if len(changes) != 1 {
		t.Fatal("expected 1 change")
	}
	if changes[0].Type != ChangeDimension {
		t.Errorf("expected ChangeDimension, got %d", changes[0].Type)
	}
}

func TestLayoutContext_FinishLayout(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("finish-layout-node")

	lc.MarkPropsChanged(node)
	if !lc.NeedsLayout(node) {
		t.Fatal("node should need layout after mark")
	}

	lc.FinishLayout(node)
	if lc.NeedsLayout(node) {
		t.Error("node should not need layout after FinishLayout")
	}
}

func TestLayoutContext_GetNodeVersion(t *testing.T) {
	lc := NewLayoutContext()
	node := newMockVNode("get-version-node")

	if v := lc.GetNodeVersion(node); v != 0 {
		t.Errorf("expected initial version 0, got %d", v)
	}

	lc.MarkPropsChanged(node)
	if v := lc.GetNodeVersion(node); v != 1 {
		t.Errorf("expected version 1, got %d", v)
	}
}

func TestLayoutContext_GetStats(t *testing.T) {
	lc := NewLayoutContext()

	lc.MarkPropsChanged(newMockVNode("stats-node-1"))
	lc.MarkContentChanged(newMockVNode("stats-node-2"))

	stats := lc.GetStats()
	if stats.TotalNodes != 2 {
		t.Errorf("expected 2 total nodes, got %d", stats.TotalNodes)
	}
}

func TestLayoutContext_Clear(t *testing.T) {
	lc := NewLayoutContext()

	lc.MarkPropsChanged(newMockVNode("clear-node-1"))

	// Verify not empty
	stats := lc.GetStats()
	if stats.TotalNodes == 0 {
		t.Fatal("expected non-empty stats before clear")
	}

	// Clear
	lc.Clear()

	// Verify empty
	stats = lc.GetStats()
	if stats.TotalNodes != 0 {
		t.Errorf("expected 0 nodes after clear, got %d", stats.TotalNodes)
	}
}

func TestLayoutContext_Nil_Handling(t *testing.T) {
	var lc *LayoutContext = nil
	var node ui.VNode = nil

	// Should not panic
	lc.NeedsLayout(node)
	lc.MarkNodeChanged(node, ChangeProps, layout.Size{}, layout.Size{})
	lc.MarkChildrenChanged(node)
	lc.MarkPropsChanged(node)
	lc.MarkContentChanged(node)
	lc.MarkSizeChanged(node, layout.Size{}, layout.Size{})
	lc.FinishLayout(node)
	lc.GetNodeVersion(node)
	lc.GetStats()
	lc.Clear()
}
