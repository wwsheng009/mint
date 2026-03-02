package wrap

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	w := New()

	if w.Tag() != "wrap" {
		t.Errorf("Tag() = %q, want %q", w.Tag(), "wrap")
	}
	if w.Gap() != 1 {
		t.Errorf("Gap() = %d, want 1", w.Gap())
	}
	if w.Align() != AlignStart {
		t.Errorf("Align() = %v, want AlignStart", w.Align())
	}
	if w.Width() != 80 {
		t.Errorf("Width() = %d, want 80", w.Width())
	}
}

func TestVNode_FluentAPI(t *testing.T) {
	w := New()
	w.SetKey("test-wrap")
	w.SetGap(2)
	w.SetRowGap(1)
	w.SetAlign(AlignCenter)
	w.SetWidth(100)
	w.SetPadding(1, 2, 1, 2)
	w.SetFillWidth(true)
	w.SetFillHeight(true)

	if w.Key() != "test-wrap" {
		t.Errorf("Key() = %q, want %q", w.Key(), "test-wrap")
	}
	if w.Gap() != 2 {
		t.Errorf("Gap() = %d, want 2", w.Gap())
	}
	if w.RowGap() != 1 {
		t.Errorf("RowGap() = %d, want 1", w.RowGap())
	}
	if w.Align() != AlignCenter {
		t.Errorf("Align() = %v, want AlignCenter", w.Align())
	}
	if w.Width() != 100 {
		t.Errorf("Width() = %d, want 100", w.Width())
	}
	if !w.FillWidth() {
		t.Error("FillWidth() = false, want true")
	}
	if !w.FillHeight() {
		t.Error("FillHeight() = false, want true")
	}

	padding := w.Padding()
	if padding != [4]int{1, 2, 1, 2} {
		t.Errorf("Padding() = %v, want [1, 2, 1, 2]", padding)
	}
}

func TestVNode_Children(t *testing.T) {
	child1 := &mockVNode{id: "child1"}
	child2 := &mockVNode{id: "child2"}

	w := New()
	w.SetChildrenList([]rtui.VNode{child1, child2})

	children := w.Children()
	if len(children) != 2 {
		t.Errorf("Children() length = %d, want 2", len(children))
	}

	// Test AddChild
	child3 := &mockVNode{id: "child3"}
	w.AddChild(child3)
	if len(w.Children()) != 3 {
		t.Errorf("Children() after AddChild = %d, want 3", len(w.Children()))
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	w := New()
	w.SetGap(3)
	w.SetWidth(60)

	inst := w.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}

	// Verify instance implements interfaces
	if _, ok := inst.(rtui.PaintableInstance); !ok {
		t.Error("Instance should implement PaintableInstance")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	props := rtui.Props{
		"key":   "test",
		"gap":   2,
		"width": 50,
	}

	inst := NewInstance(props)

	if inst.Key() != "test" {
		t.Errorf("Key() = %q, want %q", inst.Key(), "test")
	}
	if inst.gap != 2 {
		t.Errorf("gap = %d, want 2", inst.gap)
	}
	if inst.width != 50 {
		t.Errorf("width = %d, want 50", inst.width)
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{"gap": 1})

	changed := inst.SetProps(rtui.Props{"gap": 5})
	if !changed {
		t.Error("SetProps should return true when props change")
	}
	if inst.gap != 5 {
		t.Errorf("gap = %d, want 5", inst.gap)
	}

	// No change
	changed = inst.SetProps(rtui.Props{"gap": 5})
	if changed {
		t.Error("SetProps should return false when props don't change")
	}
}

func TestInstance_Measure_Empty(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":  80,
		"gap":    1,
		"padding": [4]int{0, 0, 0, 0},
	})

	size := inst.Measure(layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	})

	// Empty wrap should have minimal size (just padding)
	if size.Width < 0 {
		t.Errorf("Width = %d, should be >= 0", size.Width)
	}
	if size.Height < 0 {
		t.Errorf("Height = %d, should be >= 0", size.Height)
	}
}

func TestInstance_Measure_SingleChild(t *testing.T) {
	child := &mockVNode{id: "child", width: 10, height: 1}

	inst := NewInstance(rtui.Props{
		"width":    80,
		"gap":      1,
		"children": []rtui.VNode{child},
	})

	size := inst.Measure(layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	})

	// Should fit in one row
	if size.Width != 80 {
		t.Errorf("Width = %d, want 80", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}

	// Check rows calculation
	if len(inst.rows) != 1 {
		t.Fatalf("rows length = %d, want 1", len(inst.rows))
	}
	if len(inst.rows[0]) != 1 {
		t.Errorf("first row length = %d, want 1", len(inst.rows[0]))
	}
}

func TestInstance_Measure_MultipleChildren(t *testing.T) {
	// Create children that total 30 chars width
	// With width=20, gap=1: [child1:10] + gap + [child2:10] = 21 > 20, should wrap
	children := []rtui.VNode{
		&mockVNode{id: "child1", width: 10, height: 1},
		&mockVNode{id: "child2", width: 10, height: 1},
	}

	inst := NewInstance(rtui.Props{
		"width":    20,
		"gap":      1,
		"rowGap":   0, // rowGap=0 means use gap
		"children": children,
	})

	size := inst.Measure(layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	})

	// Should be 2 rows
	if len(inst.rows) != 2 {
		t.Errorf("rows count = %d, want 2", len(inst.rows))
	}

	// Height should be 3: 1 (row1) + 1 (rowGap, using gap) + 1 (row2) = 3
	if size.Height != 3 {
		t.Errorf("Height = %d, want 3", size.Height)
	}
}

func TestInstance_Measure_WithRowGap(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{id: "child1", width: 15, height: 1},
		&mockVNode{id: "child2", width: 15, height: 1},
	}

	inst := NewInstance(rtui.Props{
		"width":    20,
		"gap":      1,
		"rowGap":   2,
		"children": children,
	})

	size := inst.Measure(layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	})

	// 2 rows with rowGap=2: 1 + 2 + 1 = 4
	if size.Height != 4 {
		t.Errorf("Height = %d, want 4", size.Height)
	}
}

func TestInstance_Measure_WithPadding(t *testing.T) {
	child := &mockVNode{id: "child", width: 10, height: 1}

	inst := NewInstance(rtui.Props{
		"width":    80,
		"gap":      1,
		"padding":  [4]int{1, 2, 1, 2}, // top, right, bottom, left
		"children": []rtui.VNode{child},
	})

	size := inst.Measure(layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	})

	// Height should include top + bottom padding: 1 + 1 + 1 = 3
	if size.Height != 3 {
		t.Errorf("Height = %d, want 3", size.Height)
	}
}

func TestInstance_Bounds(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	inst.SetBounds(10, 20, 80, 10)
	x, y, w, h := inst.GetBounds()
	if x != 10 || y != 20 || w != 80 || h != 10 {
		t.Errorf("GetBounds() = (%d, %d, %d, %d), want (10, 20, 80, 10)", x, y, w, h)
	}
}

func TestInstance_ChildBounds(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{id: "child1"},
		&mockVNode{id: "child2"},
	}
	inst := NewInstance(rtui.Props{"children": children})

	inst.SetChildBounds(0, 0, 0, 10, 1)
	inst.SetChildBounds(1, 11, 0, 10, 1)

	x, y, w, h := inst.GetChildBounds(0)
	if x != 0 || y != 0 || w != 10 || h != 1 {
		t.Errorf("GetChildBounds(0) = (%d, %d, %d, %d), want (0, 0, 10, 1)", x, y, w, h)
	}

	x, y, w, h = inst.GetChildBounds(1)
	if x != 11 || y != 0 || w != 10 || h != 1 {
		t.Errorf("GetChildBounds(1) = (%d, %d, %d, %d), want (11, 0, 10, 1)", x, y, w, h)
	}

	// Invalid index
	x, y, w, h = inst.GetChildBounds(99)
	if x != 0 || y != 0 || w != 0 || h != 0 {
		t.Errorf("GetChildBounds(99) should return (0, 0, 0, 0), got (%d, %d, %d, %d)", x, y, w, h)
	}
}

func TestInstance_Dirty(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	if !inst.IsDirty() {
		t.Error("New instance should be dirty")
	}

	inst.ClearDirty()
	if inst.IsDirty() {
		t.Error("After ClearDirty, instance should not be dirty")
	}

	inst.MarkDirty()
	if !inst.IsDirty() {
		t.Error("After MarkDirty, instance should be dirty")
	}
}

func TestInstance_Paint(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	cmds := inst.Paint(0, 0)
	if cmds != nil {
		t.Errorf("Paint() should return nil for pure layout container, got %d commands", len(cmds))
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_FluentAPI(t *testing.T) {
	vnode := NewBuilder().
		Key("test-wrap").
		Gap(2).
		RowGap(1).
		Align(AlignCenter).
		Width(60).
		Padding(1, 1, 1, 1).
		FillWidth().
		Build()

	w, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}

	if w.Key() != "test-wrap" {
		t.Errorf("Key() = %q, want %q", w.Key(), "test-wrap")
	}
	if w.Gap() != 2 {
		t.Errorf("Gap() = %d, want 2", w.Gap())
	}
	if w.RowGap() != 1 {
		t.Errorf("RowGap() = %d, want 1", w.RowGap())
	}
	if w.Align() != AlignCenter {
		t.Errorf("Align() = %v, want AlignCenter", w.Align())
	}
	if w.Width() != 60 {
		t.Errorf("Width() = %d, want 60", w.Width())
	}
	if !w.FillWidth() {
		t.Error("FillWidth() = false, want true")
	}
}

func TestBuilder_ConvenienceMethods(t *testing.T) {
	// Test Center()
	w1 := NewBuilder().Center().Build().(*VNode)
	if w1.Align() != AlignCenter {
		t.Errorf("Center() should set AlignCenter, got %v", w1.Align())
	}

	// Test End()
	w2 := NewBuilder().End().Build().(*VNode)
	if w2.Align() != AlignEnd {
		t.Errorf("End() should set AlignEnd, got %v", w2.Align())
	}

	// Test FillHeight()
	w3 := NewBuilder().FillHeight().Build().(*VNode)
	if !w3.FillHeight() {
		t.Error("FillHeight() should set fillHeight to true")
	}
}

func TestBuilder_BuildInstance(t *testing.T) {
	inst := NewBuilder().
		Gap(3).
		Width(50).
		BuildInstance()

	if inst == nil {
		t.Fatal("BuildInstance() returned nil")
	}

	wrapInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("BuildInstance() should return *Instance")
	}

	if wrapInst.gap != 3 {
		t.Errorf("gap = %d, want 3", wrapInst.gap)
	}
	if wrapInst.width != 50 {
		t.Errorf("width = %d, want 50", wrapInst.width)
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestConvenienceFunctions(t *testing.T) {
	child := &mockVNode{id: "child"}

	// Test W()
	b := W()
	if b == nil {
		t.Error("W() returned nil")
	}

	// Test Wrap()
	w1 := Wrap(child)
	if w1 == nil {
		t.Error("Wrap() returned nil")
	}

	// Test WrapWithWidth()
	w2 := WrapWithWidth(60, child)
	if w2 == nil {
		t.Error("WrapWithWidth() returned nil")
	}

	// Test WrapWithGap()
	w3 := WrapWithGap(3, child)
	if w3 == nil {
		t.Error("WrapWithGap() returned nil")
	}

	// Test WrapConfig()
	w4 := WrapConfig(50, 2, AlignCenter, child)
	if w4 == nil {
		t.Error("WrapConfig() returned nil")
	}
}

// =============================================================================
// Mock VNode for Testing
// =============================================================================

type mockVNode struct {
	id     string
	width  int
	height int
}

func (m *mockVNode) Type() rtui.VNodeType                           { return rtui.VNodeElement }
func (m *mockVNode) Key() string                                    { return m.id }
func (m *mockVNode) SetKey(string) rtui.VNode                       { return m }
func (m *mockVNode) ID() string                                     { return m.id }
func (m *mockVNode) SetID(id string) rtui.VNode                     { m.id = id; return m }
func (m *mockVNode) Tag() string                                    { return "mock" }
func (m *mockVNode) Style() style.Style                             { return style.Style{} }
func (m *mockVNode) SetStyle(style.Style) rtui.VNode                { return m }
func (m *mockVNode) Children() []rtui.VNode                         { return nil }
func (m *mockVNode) SetChildren([]rtui.VNode) rtui.VNode            { return m }
func (m *mockVNode) Props() rtui.Props {
	return rtui.Props{
		"width":  m.width,
		"height": m.height,
	}
}
func (m *mockVNode) SetProps(rtui.Props) rtui.VNode { return m }
func (m *mockVNode) GetLayer() rtui.Layer           { return rtui.LayerBase }
func (m *mockVNode) SetLayer(rtui.Layer) rtui.VNode { return m }
func (m *mockVNode) CreateInstance() rtui.ComponentInstance {
	return &mockInstance{width: m.width, height: m.height}
}
func (m *mockVNode) SetPortalRoot(portalRootID string) rtui.VNode { return m }
func (m *mockVNode) SetAnchorTo(anchorID string, anchor types.Anchor) rtui.VNode { return m }
func (m *mockVNode) SetPortalPosition(position types.PositionType) rtui.VNode { return m }
func (m *mockVNode) SetPortalPriority(priority int) rtui.VNode { return m }
func (m *mockVNode) SetPortalRootId(portalRootId string) rtui.VNode { return m }

type mockInstance struct {
	width  int
	height int
}

func (m *mockInstance) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{Width: m.width, Height: m.height}
}
func (m *mockInstance) Key() string                          { return "" }
func (m *mockInstance) SetKey(string)                        {}
func (m *mockInstance) Init(rtui.Props)                      {}
func (m *mockInstance) Destroy()                             {}
func (m *mockInstance) OnMount()                             {}
func (m *mockInstance) OnUnmount()                           {}
func (m *mockInstance) SetProps(rtui.Props) bool             { return false }
func (m *mockInstance) GetProps() rtui.Props                 { return nil }
func (m *mockInstance) MarkDirty()                           {}
func (m *mockInstance) IsDirty() bool                        { return false }
func (m *mockInstance) GetContext() *rtui.ComponentContext   { return nil }
