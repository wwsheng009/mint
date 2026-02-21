package border

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// mockChildVNode is a mock VNode for testing
type mockChildVNode struct {
	width  int
	height int
}

func (m *mockChildVNode) Props() rtui.Props {
	return rtui.Props{
		"width":  m.width,
		"height": m.height,
	}
}
func (m *mockChildVNode) SetProps(p rtui.Props) rtui.VNode { return m }
func (m *mockChildVNode) Type() rtui.VNodeType             { return rtui.VNodeElement }
func (m *mockChildVNode) Key() string                      { return "" }
func (m *mockChildVNode) SetKey(key string) rtui.VNode     { return m }
func (m *mockChildVNode) Children() []rtui.VNode           { return nil }
func (m *mockChildVNode) SetChildren(children []rtui.VNode) rtui.VNode { return m }
func (m *mockChildVNode) Style() style.Style               { return style.Style{} }
func (m *mockChildVNode) SetStyle(s style.Style) rtui.VNode { return m }
func (m *mockChildVNode) Tag() string                      { return "mock" }
func (m *mockChildVNode) GetLayer() rtui.Layer             { return rtui.LayerBase }
func (m *mockChildVNode) SetLayer(l rtui.Layer) rtui.VNode { return m }

// TestMeasureWithExplicitDimensions tests that explicit dimensions are used when set
func TestMeasureWithExplicitDimensions(t *testing.T) {
	child := &mockChildVNode{width: 100, height: 10}
	
	inst := NewInstance(rtui.Props{
		"width":  20,
		"height": 5,
		"child":  child,
	})

	constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
	size := inst.Measure(constraints)

	// Expected: 20 + 2 (border) = 22, 5 + 2 = 7
	if size.Width != 22 {
		t.Errorf("Expected width 22, got %d", size.Width)
	}
	if size.Height != 7 {
		t.Errorf("Expected height 7, got %d", size.Height)
	}
}

// TestMeasureAutoWidth tests auto-measuring width from child
func TestMeasureAutoWidth(t *testing.T) {
	child := &mockChildVNode{width: 50, height: 5}
	
	inst := NewInstance(rtui.Props{
		"height": 5, // explicit height, auto width
		"child":  child,
	})

	constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
	size := inst.Measure(constraints)

	// Expected width: 50 (from child props) + 2 (border) = 52
	if size.Width != 52 {
		t.Errorf("Expected width 52 (auto-measured), got %d", size.Width)
	}
	// Height should be explicit: 5 + 2 = 7
	if size.Height != 7 {
		t.Errorf("Expected height 7, got %d", size.Height)
	}
}

// TestMeasureAutoHeight tests auto-measuring height from child
func TestMeasureAutoHeight(t *testing.T) {
	child := &mockChildVNode{width: 30, height: 8}
	
	inst := NewInstance(rtui.Props{
		"width": 30, // explicit width, auto height
		"child":  child,
	})

	constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
	size := inst.Measure(constraints)

	// Width should be explicit: 30 + 2 = 32
	if size.Width != 32 {
		t.Errorf("Expected width 32, got %d", size.Width)
	}
	// Expected height: 8 (from child props) + 2 (border) = 10
	if size.Height != 10 {
		t.Errorf("Expected height 10 (auto-measured), got %d", size.Height)
	}
}

// TestMeasureAutoBoth tests auto-measuring both dimensions from child
func TestMeasureAutoBoth(t *testing.T) {
	child := &mockChildVNode{width: 40, height: 6}
	
	inst := NewInstance(rtui.Props{
		"child": child, // no explicit dimensions
	})

	constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
	size := inst.Measure(constraints)

	// Expected: 40 + 2 = 42, 6 + 2 = 8
	if size.Width != 42 {
		t.Errorf("Expected width 42 (auto-measured), got %d", size.Width)
	}
	if size.Height != 8 {
		t.Errorf("Expected height 8 (auto-measured), got %d", size.Height)
	}
}

// TestMeasureNoChild tests behavior when no child is present
func TestMeasureNoChild(t *testing.T) {
	inst := NewInstance(rtui.Props{
		// no child, no dimensions
	})

	constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
	size := inst.Measure(constraints)

	// With no child and no dimensions, should be minimal (border only)
	if size.Width != 2 {
		t.Errorf("Expected width 2 (border only), got %d", size.Width)
	}
	if size.Height != 2 {
		t.Errorf("Expected height 2 (border only), got %d", size.Height)
	}
}

// TestMeasureWithConstraints tests that constraints are respected
func TestMeasureWithConstraints(t *testing.T) {
	child := &mockChildVNode{width: 100, height: 20}
	
	inst := NewInstance(rtui.Props{
		"child": child,
	})

	// Tight constraints
	constraints := layout.Constraints{MaxWidth: 30, MaxHeight: 10}
	size := inst.Measure(constraints)

	// Should be constrained
	if size.Width > 30 {
		t.Errorf("Width %d exceeds MaxWidth 30", size.Width)
	}
	if size.Height > 10 {
		t.Errorf("Height %d exceeds MaxHeight 10", size.Height)
	}
}

// TestMeasuredChildSizeCached tests that measured size is cached
func TestMeasuredChildSizeCached(t *testing.T) {
	child := &mockChildVNode{width: 50, height: 10}
	
	inst := NewInstance(rtui.Props{
		"child": child,
	})

	constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
	inst.Measure(constraints)

	// Check that measuredChildSize is cached
	if inst.measuredChildSize.Width != 50 {
		t.Errorf("Expected cached child width 50, got %d", inst.measuredChildSize.Width)
	}
	if inst.measuredChildSize.Height != 10 {
		t.Errorf("Expected cached child height 10, got %d", inst.measuredChildSize.Height)
	}
}

// TestBorderStyles tests different border styles
func TestBorderStyles(t *testing.T) {
	tests := []struct {
		style         BorderStyle
		expectedW     int
		expectedH     int
	}{
		{BorderNone, 10, 5},       // No border
		{BorderSingle, 12, 7},     // 1 + 10 + 1 = 12, 1 + 5 + 1 = 7
		{BorderDouble, 12, 7},     // Same as single (all glyphs are 1-char wide)
		{BorderRounded, 12, 7},    // Same as single
		{BorderDashed, 12, 7},     // Same as single
	}

	for _, tt := range tests {
		inst := NewInstance(rtui.Props{
			"borderStyle": tt.style,
			"width":       10,
			"height":      5,
		})

		constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
		size := inst.Measure(constraints)

		if size.Width != tt.expectedW || size.Height != tt.expectedH {
			t.Errorf("BorderStyle %v: expected %dx%d, got %dx%d",
				tt.style, tt.expectedW, tt.expectedH, size.Width, size.Height)
		}
	}
}

// TestBorderLabel tests that border label is correctly rendered
func TestBorderLabel(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":       30,
		"height":      5,
		"borderStyle": BorderSingle,
		"borderLabel": " Test ",
	})

	inst.SetBounds(0, 0, 32, 7)
	constraints := layout.Constraints{MaxWidth: 1000, MaxHeight: 100}
	inst.Measure(constraints)

	cmds := inst.Paint(0, 0)

	if len(cmds) == 0 {
		t.Fatal("Expected border commands")
	}

	// First command should be top border with label
	topBorder := cmds[0]
	if topBorder.Y != 0 {
		t.Errorf("Top border should be at y=0, got y=%d", topBorder.Y)
	}

	// Check that label appears in top border
	if !strings.Contains(topBorder.Text, "Test") {
		t.Errorf("Top border should contain 'Test', got: %s", topBorder.Text)
	}
}

// TestBorderLabelFromVNode tests that label is passed from VNode to Instance
func TestBorderLabelFromVNode(t *testing.T) {
	vnode := New().
		SetBorderLabel(" My Label ").
		SetWidth(30).
		SetHeight(5)

	inst := vnode.CreateInstance()
	borderInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Expected *Instance")
	}

	if borderInst.borderLabel != " My Label " {
		t.Errorf("Expected borderLabel ' My Label ', got '%s'", borderInst.borderLabel)
	}
}
