package absolute

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	child := newtext.New("Test")
	a := New(child)

	if a == nil {
		t.Fatal("New() returned nil")
	}

	if a.Tag() != "absolute" {
		t.Errorf("Expected tag 'absolute', got '%s'", a.Tag())
	}

	if a.Child() != child {
		t.Error("Child should be set")
	}

	if a.AnchorPoint() != AnchorTopLeft {
		t.Errorf("Default anchor should be AnchorTopLeft, got %d", a.AnchorPoint())
	}
}

func TestVNode_Position(t *testing.T) {
	child := newtext.New("Test")

	a := New(child).
		SetLeft(AbsolutePos(10)).
		SetTop(AbsolutePos(5))

	if a.LeftPos() == nil {
		t.Error("Left should not be nil")
	}

	if a.TopPos() == nil {
		t.Error("Top should not be nil")
	}

	// Test position calculation
	x, y := a.CalculatePosition(100, 50)
	if x != 10 {
		t.Errorf("X: expected 10, got %d", x)
	}
	if y != 5 {
		t.Errorf("Y: expected 5, got %d", y)
	}
}

func TestVNode_RelativePosition(t *testing.T) {
	child := newtext.New("Test")

	a := New(child).
		SetLeft(RelativePos(50)).
		SetTop(RelativePos(25)).
		Size(20, 10)

	// Test position calculation (50% of 100 = 50, 25% of 50 = 12)
	x, y := a.CalculatePosition(100, 50)
	if x != 50 {
		t.Errorf("X: expected 50, got %d", x)
	}
	if y != 12 {
		t.Errorf("Y: expected 12, got %d", y)
	}
}

func TestVNode_RightBottomPosition(t *testing.T) {
	child := newtext.New("Test")

	a := New(child).
		SetRight(AbsolutePos(10)).
		SetBottom(AbsolutePos(5)).
		Size(20, 10)

	// Test position calculation
	// Right: 100 - 10 = 90
	// Bottom: 50 - 5 = 45
	x, y := a.CalculatePosition(100, 50)
	if x != 90 {
		t.Errorf("X: expected 90, got %d", x)
	}
	if y != 45 {
		t.Errorf("Y: expected 45, got %d", y)
	}
}

func TestVNode_Anchor(t *testing.T) {
	child := newtext.New("Test")

	// Test center anchor
	a := New(child).
		SetLeft(RelativePos(50)).
		SetTop(RelativePos(50)).
		SetAnchor(AnchorCenter).
		Size(20, 10)

	x, y := a.CalculatePosition(100, 50)
	// Center: x - width/2, y - height/2
	if x != 40 { // 50 - 20/2
		t.Errorf("X: expected 40, got %d", x)
	}
	if y != 20 { // 25 - 10/2
		t.Errorf("Y: expected 20, got %d", y)
	}
}

func TestVNode_AnchorBottomRight(t *testing.T) {
	child := newtext.New("Test")

	a := New(child).
		SetLeft(AbsolutePos(100)).
		SetTop(AbsolutePos(50)).
		SetAnchor(AnchorBottomRight).
		Size(20, 10)

	x, y := a.CalculatePosition(200, 100)
	// BottomRight: x - width, y - height
	if x != 80 { // 100 - 20
		t.Errorf("X: expected 80, got %d", x)
	}
	if y != 40 { // 50 - 10
		t.Errorf("Y: expected 40, got %d", y)
	}
}

func TestVNode_Size(t *testing.T) {
	child := newtext.New("Test")

	a := New(child).
		SetWidth(50).
		SetHeight(10)

	if a.AbsWidth() != 50 {
		t.Errorf("Width: expected 50, got %d", a.AbsWidth())
	}
	if a.AbsHeight() != 10 {
		t.Errorf("Height: expected 10, got %d", a.AbsHeight())
	}
}

func TestVNode_ZIndex(t *testing.T) {
	child := newtext.New("Test")

	a := New(child).SetZIndex(5)

	if a.ZIndex() != 5 {
		t.Errorf("ZIndex: expected 5, got %d", a.ZIndex())
	}
}

func TestVNode_InstanceFactory(t *testing.T) {
	child := newtext.New("Test")

	a := New(child).SetWidth(50)

	inst := a.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	absInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not a *Instance")
	}

	if absInst.width != 50 {
		t.Errorf("Instance width: expected 50, got %d", absInst.width)
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":  30,
		"height": 5,
	})

	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := inst.Measure(constraints)

	if size.Width != 30 {
		t.Errorf("Width: expected 30, got %d", size.Width)
	}
	if size.Height != 5 {
		t.Errorf("Height: expected 5, got %d", size.Height)
	}
}

func TestInstance_Measure_Default(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := inst.Measure(constraints)

	// Default size: 20x1
	if size.Width != 20 {
		t.Errorf("Width: expected 20, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height: expected 1, got %d", size.Height)
	}
}

func TestInstance_Measure_Constraints(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":  100,
		"height": 50,
	})

	constraints := layout.TightConstraints(30, 10)
	size := inst.Measure(constraints)

	if size.Width != 30 {
		t.Errorf("Width: expected 30, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("Height: expected 10, got %d", size.Height)
	}
}

func TestInstance_Bounds(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	inst.SetBounds(5, 10, 30, 5)
	x, y, w, h := inst.GetBounds()

	if x != 5 || y != 10 || w != 30 || h != 5 {
		t.Errorf("Bounds: expected (5,10,30,5), got (%d,%d,%d,%d)", x, y, w, h)
	}
}

func TestInstance_CalculatePosition(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"left":   AbsolutePos(10),
		"top":    AbsolutePos(5),
		"width":  20,
		"height": 10,
	})

	x, y := inst.CalculatePosition(100, 50)
	if x != 10 {
		t.Errorf("X: expected 10, got %d", x)
	}
	if y != 5 {
		t.Errorf("Y: expected 5, got %d", y)
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width": 20,
	})

	changed := inst.SetProps(rtui.Props{
		"width":  50,
		"height": 10,
	})

	if !changed {
		t.Error("SetProps should return true when props change")
	}

	if inst.width != 50 {
		t.Errorf("Width: expected 50, got %d", inst.width)
	}
	if inst.height != 10 {
		t.Errorf("Height: expected 10, got %d", inst.height)
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_Full(t *testing.T) {
	child := newtext.New("Test")

	a := NewBuilder(child).
		Key("test-abs").
		Left(AbsolutePos(10)).
		Top(AbsolutePos(5)).
		Width(30).
		Height(10).
		ZIndex(3).
		Build()

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}

	if abs.Key() != "test-abs" {
		t.Errorf("Key: expected 'test-abs', got '%s'", abs.Key())
	}
	if abs.AbsWidth() != 30 {
		t.Errorf("Width: expected 30, got %d", abs.AbsWidth())
	}
	if abs.AbsHeight() != 10 {
		t.Errorf("Height: expected 10, got %d", abs.AbsHeight())
	}
	if abs.ZIndex() != 3 {
		t.Errorf("ZIndex: expected 3, got %d", abs.ZIndex())
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestTopLeft(t *testing.T) {
	child := newtext.New("Test")
	a := TopLeft(child)

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("TopLeft should return *VNode")
	}

	x, y := abs.CalculatePosition(100, 50)
	if x != 0 || y != 0 {
		t.Errorf("Position: expected (0,0), got (%d,%d)", x, y)
	}
}

func TestTopRight(t *testing.T) {
	child := newtext.New("Test")
	a := TopRight(child)

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("TopRight should return *VNode")
	}

	if abs.AnchorPoint() != AnchorTopRight {
		t.Errorf("Anchor: expected AnchorTopRight, got %d", abs.AnchorPoint())
	}
}

func TestBottomLeft(t *testing.T) {
	child := newtext.New("Test")
	a := BottomLeft(child)

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("BottomLeft should return *VNode")
	}

	if abs.AnchorPoint() != AnchorBottomLeft {
		t.Errorf("Anchor: expected AnchorBottomLeft, got %d", abs.AnchorPoint())
	}
}

func TestBottomRight(t *testing.T) {
	child := newtext.New("Test")
	a := BottomRight(child)

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("BottomRight should return *VNode")
	}

	if abs.AnchorPoint() != AnchorBottomRight {
		t.Errorf("Anchor: expected AnchorBottomRight, got %d", abs.AnchorPoint())
	}
}

func TestCenter(t *testing.T) {
	child := newtext.New("Test")
	a := Center(child)

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("Center should return *VNode")
	}

	if abs.AnchorPoint() != AnchorCenter {
		t.Errorf("Anchor: expected AnchorCenter, got %d", abs.AnchorPoint())
	}

	// Test center calculation
	abs.SetWidth(20).SetHeight(10)
	x, y := abs.CalculatePosition(100, 50)
	// 50% - 20/2 = 40, 50% - 10/2 = 20
	if x != 40 {
		t.Errorf("X: expected 40, got %d", x)
	}
	if y != 20 {
		t.Errorf("Y: expected 20, got %d", y)
	}
}

func TestAt(t *testing.T) {
	child := newtext.New("Test")
	a := At(child, 15, 25)

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("At should return *VNode")
	}

	x, y := abs.CalculatePosition(100, 50)
	if x != 15 {
		t.Errorf("X: expected 15, got %d", x)
	}
	if y != 25 {
		t.Errorf("Y: expected 25, got %d", y)
	}
}

func TestAtPercent(t *testing.T) {
	child := newtext.New("Test")
	a := AtPercent(child, 25, 75)

	abs, ok := a.(*VNode)
	if !ok {
		t.Fatal("AtPercent should return *VNode")
	}

	// 25% of 100 = 25, 75% of 50 = 37
	x, y := abs.CalculatePosition(100, 50)
	if x != 25 {
		t.Errorf("X: expected 25, got %d", x)
	}
	if y != 37 { // 50 * 75 / 100 = 37
		t.Errorf("Y: expected 37, got %d", y)
	}
}

// =============================================================================
// Position Type Tests
// =============================================================================

func TestPositionTypes(t *testing.T) {
	// AbsolutePos
	abs := AbsolutePos(10)
	if int(abs) != 10 {
		t.Errorf("AbsolutePos: expected 10, got %d", int(abs))
	}

	// RelativePos
	rel := RelativePos(50)
	if int(rel) != 50 {
		t.Errorf("RelativePos: expected 50, got %d", int(rel))
	}
}

// =============================================================================
// Anchor Tests
// =============================================================================

func TestAllAnchors(t *testing.T) {
	anchors := []Anchor{
		AnchorTopLeft,
		AnchorTop,
		AnchorTopRight,
		AnchorLeft,
		AnchorCenter,
		AnchorRight,
		AnchorBottomLeft,
		AnchorBottom,
		AnchorBottomRight,
	}

	for i, anchor := range anchors {
		if int(anchor) != i {
			t.Errorf("Anchor %d: expected value %d, got %d", i, i, int(anchor))
		}
	}
}
