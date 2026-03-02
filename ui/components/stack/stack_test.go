package stack

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestNew(t *testing.T) {
	h := NewHStack()
	if h.Direction() != Row {
		t.Error("HStack should have Row direction")
	}
	if h.Tag() != "hstack" {
		t.Errorf("Tag() = %q, want %q", h.Tag(), "hstack")
	}

	v := NewVStack()
	if v.Direction() != Column {
		t.Error("VStack should have Column direction")
	}
	if v.Tag() != "vstack" {
		t.Errorf("Tag() = %q, want %q", v.Tag(), "vstack")
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	h := NewHStack()

	// Test VNode interface
	var _ rtui.VNode = h

	// Test InstanceFactory interface
	var _ rtui.InstanceFactory = h
}

func TestVNode_CreateInstance(t *testing.T) {
	h := NewHStack()
	h.SetKey("test-stack")
	h.SetGap(2)
	h.SetPadding(1, 2, 1, 2)

	inst := h.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}

	stackInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("CreateInstance() did not return *Instance")
	}

	if stackInst.key != "test-stack" {
		t.Errorf("key = %q, want %q", stackInst.key, "test-stack")
	}

	if stackInst.gap != 2 {
		t.Errorf("gap = %d, want 2", stackInst.gap)
	}

	if stackInst.padding != [4]int{1, 2, 1, 2} {
		t.Errorf("padding = %v, want [1, 2, 1, 2]", stackInst.padding)
	}
}

func TestVNode_FluentAPI(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{width: 10, height: 1},
		&mockVNode{width: 20, height: 1},
	}

	h := NewHStack()
	h.SetKey("test")
	h.SetGap(1)
	h.SetAlign(AlignCenter)
	h.SetCrossAlign(AlignCenter)
	h.SetStretchCross(true)
	h.SetWidth(100)
	h.SetHeight(10)
	h.SetChildrenList(children)

	if h.Key() != "test" {
		t.Errorf("Key() = %q, want %q", h.Key(), "test")
	}

	if h.Gap() != 1 {
		t.Errorf("Gap() = %d, want 1", h.Gap())
	}

	if h.Align() != AlignCenter {
		t.Errorf("Align() = %v, want AlignCenter", h.Align())
	}

	if h.CrossAlign() != AlignCenter {
		t.Errorf("CrossAlign() = %v, want AlignCenter", h.CrossAlign())
	}

	if !h.StretchCross() {
		t.Error("StretchCross() should be true")
	}

	if h.Width() != 100 {
		t.Errorf("Width() = %d, want 100", h.Width())
	}

	if h.Height() != 10 {
		t.Errorf("Height() = %d, want 10", h.Height())
	}

	if len(h.Children()) != 2 {
		t.Errorf("len(Children()) = %d, want 2", len(h.Children()))
	}
}

func TestVNode_ConvenienceMethods(t *testing.T) {
	h := New(Row).Horizontal()
	if h.Direction() != Row {
		t.Error("Horizontal() should set Row direction")
	}

	v := New(Row).Vertical()
	if v.Direction() != Column {
		t.Error("Vertical() should set Column direction")
	}

	s := NewHStack().Stretch()
	if !s.StretchCross() {
		t.Error("Stretch() should enable stretchCross")
	}

	c := NewHStack().Center()
	if c.Align() != AlignCenter {
		t.Error("Center() should set AlignCenter")
	}

	cc := NewHStack().CenterCross()
	if cc.CrossAlign() != AlignCenter {
		t.Error("CenterCross() should set AlignCenter for cross axis")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_Measure_Empty(t *testing.T) {
	h := NewHStack()
	inst := h.CreateInstance().(*Instance)

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 20,
	}

	size := inst.Measure(constraints)

	if size.Width != 0 {
		t.Errorf("Empty stack width = %d, want 0", size.Width)
	}

	if size.Height != 0 {
		t.Errorf("Empty stack height = %d, want 0", size.Height)
	}
}

func TestInstance_Measure_HStack(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{width: 10, height: 2},
		&mockVNode{width: 20, height: 3},
		&mockVNode{width: 15, height: 1},
	}

	h := NewHStack().SetGap(1).SetChildrenList(children)
	inst := h.CreateInstance().(*Instance)

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 20,
	}

	size := inst.Measure(constraints)

	// Total width = 10 + 1(gap) + 20 + 1(gap) + 15 = 47
	expectedWidth := 10 + 1 + 20 + 1 + 15
	if size.Width != expectedWidth {
		t.Errorf("Width = %d, want %d", size.Width, expectedWidth)
	}

	// Height should be max child height = 3
	if size.Height != 3 {
		t.Errorf("Height = %d, want 3", size.Height)
	}
}

func TestInstance_Measure_VStack(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{width: 10, height: 2},
		&mockVNode{width: 20, height: 3},
		&mockVNode{width: 15, height: 1},
	}

	v := NewVStack().SetGap(1).SetChildrenList(children)
	inst := v.CreateInstance().(*Instance)

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 20,
	}

	size := inst.Measure(constraints)

	// Width should be max child width = 20
	if size.Width != 20 {
		t.Errorf("Width = %d, want 20", size.Width)
	}

	// Total height = 2 + 1(gap) + 3 + 1(gap) + 1 = 8
	expectedHeight := 2 + 1 + 3 + 1 + 1
	if size.Height != expectedHeight {
		t.Errorf("Height = %d, want %d", size.Height, expectedHeight)
	}
}

func TestInstance_Measure_WithPadding(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{width: 10, height: 1},
	}

	h := NewHStack().
		SetPadding(1, 2, 3, 4). // top=1, right=2, bottom=3, left=4
		SetChildrenList(children)
	inst := h.CreateInstance().(*Instance)

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 20,
	}

	size := inst.Measure(constraints)

	// Width = child(10) + padding_left(4) + padding_right(2) = 16
	if size.Width != 16 {
		t.Errorf("Width = %d, want 16", size.Width)
	}

	// Height = child(1) + padding_top(1) + padding_bottom(3) = 5
	if size.Height != 5 {
		t.Errorf("Height = %d, want 5", size.Height)
	}
}

func TestInstance_Measure_ExplicitSize(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{width: 10, height: 1},
	}

	h := NewHStack().
		SetWidth(50).
		SetHeight(10).
		SetChildrenList(children)
	inst := h.CreateInstance().(*Instance)

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 20,
	}

	size := inst.Measure(constraints)

	if size.Width != 50 {
		t.Errorf("Width = %d, want 50 (explicit)", size.Width)
	}

	if size.Height != 10 {
		t.Errorf("Height = %d, want 10 (explicit)", size.Height)
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"gap": 1,
	})

	// Update props
	changed := inst.SetProps(rtui.Props{
		"gap": 2,
		"direction": Column,
	})

	if !changed {
		t.Error("SetProps() returned false, want true")
	}

	if inst.gap != 2 {
		t.Errorf("gap = %d, want 2", inst.gap)
	}

	if inst.direction != Column {
		t.Errorf("direction = %v, want Column", inst.direction)
	}

	// Set same props again
	changed = inst.SetProps(rtui.Props{
		"gap": 2,
		"direction": Column,
	})

	if changed {
		t.Error("SetProps() returned true for unchanged props, want false")
	}
}

func TestInstance_Bounds(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	inst.SetBounds(10, 20, 100, 50)
	x, y, w, h := inst.GetBounds()

	if x != 10 || y != 20 || w != 100 || h != 50 {
		t.Errorf("GetBounds() = (%d, %d, %d, %d), want (10, 20, 100, 50)", x, y, w, h)
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder(t *testing.T) {
	children := []rtui.VNode{
		&mockVNode{width: 10, height: 1},
	}

	s := NewBuilder(Row).
		Key("test-key").
		Gap(2).
		Padding(1, 1, 1, 1).
		Stretch().
		Children(children...).
		Build()

	vnode, ok := s.(*VNode)
	if !ok {
		t.Fatal("Build() did not return *VNode")
	}

	if vnode.Key() != "test-key" {
		t.Errorf("Key() = %q, want %q", vnode.Key(), "test-key")
	}

	if vnode.Gap() != 2 {
		t.Errorf("Gap() = %d, want 2", vnode.Gap())
	}

	if !vnode.StretchCross() {
		t.Error("StretchCross() should be true")
	}
}

func TestHStackBuilder(t *testing.T) {
	h := NewHStackBuilder().
		Gap(1).
		Center().
		Build()

	vnode, ok := h.(*VNode)
	if !ok {
		t.Fatal("Build() did not return *VNode")
	}

	if vnode.Direction() != Row {
		t.Error("HStackBuilder should create Row direction")
	}

	if vnode.Align() != AlignCenter {
		t.Error("Center() should set AlignCenter")
	}
}

func TestVStackBuilder(t *testing.T) {
	v := NewVStackBuilder().
		Gap(2).
		Build()

	vnode, ok := v.(*VNode)
	if !ok {
		t.Fatal("Build() did not return *VNode")
	}

	if vnode.Direction() != Column {
		t.Error("VStackBuilder should create Column direction")
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestH(t *testing.T) {
	h := H(&mockVNode{}, &mockVNode{})

	if h.Direction() != Row {
		t.Error("H() should create Row direction")
	}

	if len(h.Children()) != 2 {
		t.Errorf("len(Children()) = %d, want 2", len(h.Children()))
	}
}

func TestV(t *testing.T) {
	v := V(&mockVNode{}, &mockVNode{})

	if v.Direction() != Column {
		t.Error("V() should create Column direction")
	}
}

func TestRowStack(t *testing.T) {
	rs := RowStack(5, &mockVNode{}, &mockVNode{})

	vnode := rs.(*VNode)
	if vnode.Direction() != Row {
		t.Error("RowStack should create Row direction")
	}

	if vnode.Gap() != 5 {
		t.Errorf("Gap() = %d, want 5", vnode.Gap())
	}
}

func TestColStack(t *testing.T) {
	cs := ColStack(3, &mockVNode{}, &mockVNode{})

	vnode := cs.(*VNode)
	if vnode.Direction() != Column {
		t.Error("ColStack should create Column direction")
	}

	if vnode.Gap() != 3 {
		t.Errorf("Gap() = %d, want 3", vnode.Gap())
	}
}

func TestSpacer(t *testing.T) {
	sp := Spacer(2)

	if sp.GetLayoutInfo().Flex != 2 {
		t.Errorf("Spacer flex = %d, want 2", sp.GetLayoutInfo().Flex)
	}
}

// =============================================================================
// Mock VNode
// =============================================================================

type mockVNode struct {
	width, height int
	content       string
	id            string
}

func (m *mockVNode) Type() rtui.VNodeType                      { return rtui.VNodeElement }
func (m *mockVNode) ID() string                                { return m.id }
func (m *mockVNode) SetID(id string) rtui.VNode                { m.id = id; return m }
func (m *mockVNode) Key() string                               { return "" }
func (m *mockVNode) SetKey(string) rtui.VNode                  { return m }
func (m *mockVNode) Tag() string                               { return "mock" }
func (m *mockVNode) Style() style.Style                        { return style.Style{} }
func (m *mockVNode) SetStyle(style.Style) rtui.VNode           { return m }
func (m *mockVNode) Children() []rtui.VNode                    { return nil }
func (m *mockVNode) SetChildren([]rtui.VNode) rtui.VNode       { return m }
func (m *mockVNode) Props() rtui.Props {
	return rtui.Props{"content": m.content, "width": m.width, "height": m.height}
}
func (m *mockVNode) SetProps(rtui.Props) rtui.VNode             { return m }
func (m *mockVNode) GetLayer() rtui.Layer                       { return rtui.LayerBase }
func (m *mockVNode) SetLayer(rtui.Layer) rtui.VNode            { return m }
func (m *mockVNode) SetPortalRoot(portalRootID string) rtui.VNode { return m }
func (m *mockVNode) SetAnchorTo(anchorID string, anchor types.Anchor) rtui.VNode { return m }
func (m *mockVNode) SetPortalPosition(position types.PositionType) rtui.VNode { return m }
func (m *mockVNode) SetPortalPriority(priority int) rtui.VNode { return m }
func (m *mockVNode) SetPortalRootId(portalRootId string) rtui.VNode { return m }
func (m *mockVNode) CreateInstance() rtui.ComponentInstance {
	return &mockInstance{width: m.width, height: m.height}
}
func (m *mockVNode) GetLayoutInfo() rtui.LayoutInfo              { return rtui.LayoutInfo{} }

type mockInstance struct {
	width, height int
	bounds        [4]int
}

func (m *mockInstance) Key() string                       { return "" }
func (m *mockInstance) SetKey(string)                     {}
func (m *mockInstance) Init(rtui.Props)                   {}
func (m *mockInstance) Destroy()                          {}
func (m *mockInstance) OnMount()                          {}
func (m *mockInstance) OnUnmount()                        {}
func (m *mockInstance) SetProps(rtui.Props) bool          { return false }
func (m *mockInstance) GetProps() rtui.Props              { return nil }
func (m *mockInstance) MarkDirty()                        {}
func (m *mockInstance) IsDirty() bool                     { return false }
func (m *mockInstance) GetContext() *rtui.ComponentContext { return nil }
func (m *mockInstance) GetBounds() (int, int, int, int)    { return m.bounds[0], m.bounds[1], m.bounds[2], m.bounds[3] }
func (m *mockInstance) SetBounds(x, y, w, h int)           { m.bounds = [4]int{x, y, w, h} }
func (m *mockInstance) Measure(c layout.Constraints) layout.Size {
	return layout.Size{Width: m.width, Height: m.height}
}
func (m *mockInstance) Paint(x, y int) []paint.DrawCmd { return nil }
