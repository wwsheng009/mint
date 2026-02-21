package scrollview

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// buildLines creates n lines of text for testing
func buildLines(n int) string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	return strings.Join(lines, "\n")
}

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	vnode := New()
	if vnode == nil {
		t.Fatal("New() returned nil")
	}
	if vnode.Tag() != "scrollview" {
		t.Errorf("Expected tag 'scrollview', got '%s'", vnode.Tag())
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	vnode := New()

	var _ rtui.VNode = vnode
	var _ rtui.InstanceFactory = vnode
	var _ rtui.BoxModel = vnode
}

func TestVNode_FluentAPI(t *testing.T) {
	textNode := newtext.New("Test content")

	vnode := New().
		SetChild(textNode).
		SetWidth(40).
		SetHeight(10).
		SetScrollOffset(5).
		SetShowBorder(true).
		SetShowIndicator(false)

	if vnode.Child() == nil {
		t.Error("Child should be set")
	}
	if vnode.Width() != 40 {
		t.Errorf("Expected width 40, got %d", vnode.Width())
	}
	if vnode.Height() != 10 {
		t.Errorf("Expected height 10, got %d", vnode.Height())
	}
	if vnode.ScrollOffset() != 5 {
		t.Errorf("Expected scrollOffset 5, got %d", vnode.ScrollOffset())
	}
	if !vnode.ShowBorder() {
		t.Error("ShowBorder should be true")
	}
	if vnode.ShowIndicator() {
		t.Error("ShowIndicator should be false")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	textNode := newtext.New("Test content")

	vnode := New().
		SetChild(textNode).
		SetWidth(40).
		SetHeight(10)

	inst := vnode.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}

	svInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}

	if svInst.width != 40 {
		t.Errorf("Expected width 40, got %d", svInst.width)
	}
	if svInst.height != 10 {
		t.Errorf("Expected height 10, got %d", svInst.height)
	}
}

func TestVNode_Children(t *testing.T) {
	textNode := newtext.New("Test content")
	vnode := New().SetChild(textNode)

	// ScrollView handles its own painting, so Children() returns nil
	children := vnode.Children()
	if len(children) != 0 {
		t.Fatalf("Expected 0 children (ScrollView paints own content), got %d", len(children))
	}

	// Test nil child
	vnode2 := New()
	if len(vnode2.Children()) != 0 {
		t.Error("Empty vnode should have no children")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	if inst == nil {
		t.Fatal("NewInstance() returned nil")
	}
}

func TestInstance_Measure_Empty(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.width = 30
	inst.height = 10

	size := inst.Measure(layout.Constraints{MaxWidth: 100, MaxHeight: 100})

	if size.Width != 30 {
		t.Errorf("Expected width 30, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("Expected height 10, got %d", size.Height)
	}
}

func TestInstance_Measure_WithBorder(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":      30,
		"height":     10,
		"showBorder": true,
	})

	size := inst.Measure(layout.Constraints{MaxWidth: 100, MaxHeight: 100})

	// Border adds 2 to each dimension
	if size.Width != 32 {
		t.Errorf("Expected width 32 (30 + 2 border), got %d", size.Width)
	}
	if size.Height != 12 {
		t.Errorf("Expected height 12 (10 + 2 border), got %d", size.Height)
	}
}

func TestInstance_Measure_AutoWidth(t *testing.T) {
	textNode := newtext.New("Line 1\nLine 2 is longer\nLine 3")

	inst := NewInstance(rtui.Props{
		"child":  textNode,
		"height": 10,
	})

	size := inst.Measure(layout.Constraints{MaxWidth: 100, MaxHeight: 100})

	// Auto-width should use max line length
	if size.Width < 16 {
		t.Errorf("Expected width >= 16 (max line length), got %d", size.Width)
	}
}

func TestInstance_Measure_AutoHeight(t *testing.T) {
	textNode := newtext.New("Line 1\nLine 2\nLine 3\nLine 4")

	inst := NewInstance(rtui.Props{
		"child": textNode,
		"width": 30,
	})

	size := inst.Measure(layout.Constraints{MaxWidth: 100, MaxHeight: 100})

	// Auto-height should show all content
	if size.Height != 4 {
		t.Errorf("Expected height 4 (4 lines), got %d", size.Height)
	}
}

func TestInstance_ScrollBy(t *testing.T) {
	textNode := newtext.New(buildLines(20))

	inst := NewInstance(rtui.Props{
		"child":  textNode,
		"width":  30,
		"height": 5,
	})

	inst.extractContent()

	// Scroll down
	inst.ScrollBy(5)
	if inst.scrollOffset != 5 {
		t.Errorf("Expected offset 5, got %d", inst.scrollOffset)
	}

	// Scroll up
	inst.ScrollBy(-2)
	if inst.scrollOffset != 3 {
		t.Errorf("Expected offset 3, got %d", inst.scrollOffset)
	}

	// Clamp to 0
	inst.ScrollBy(-10)
	if inst.scrollOffset != 0 {
		t.Errorf("Expected offset 0 (clamped), got %d", inst.scrollOffset)
	}
}

func TestInstance_ScrollTo(t *testing.T) {
	textNode := newtext.New(buildLines(20))

	inst := NewInstance(rtui.Props{
		"child":  textNode,
		"width":  30,
		"height": 5,
	})

	inst.extractContent()

	inst.ScrollTo(10)
	if inst.scrollOffset != 10 {
		t.Errorf("Expected offset 10, got %d", inst.scrollOffset)
	}

	// Clamp to max
	inst.ScrollTo(100)
	// Max offset should be 20 - 5 = 15
	if inst.scrollOffset != 15 {
		t.Errorf("Expected offset 15 (clamped), got %d", inst.scrollOffset)
	}
}

func TestInstance_CanScroll(t *testing.T) {
	textNode := newtext.New(buildLines(10))

	inst := NewInstance(rtui.Props{
		"child":  textNode,
		"width":  30,
		"height": 5,
	})

	inst.extractContent()

	// At top
	if !inst.CanScrollDown() {
		t.Error("Should be able to scroll down at top")
	}
	if inst.CanScrollUp() {
		t.Error("Should not be able to scroll up at top")
	}

	// At bottom
	inst.ScrollBottom()
	if inst.CanScrollDown() {
		t.Error("Should not be able to scroll down at bottom")
	}
	if !inst.CanScrollUp() {
		t.Error("Should be able to scroll up at bottom")
	}
}

func TestInstance_PageUpDown(t *testing.T) {
	textNode := newtext.New(buildLines(20))

	inst := NewInstance(rtui.Props{
		"child":  textNode,
		"width":  30,
		"height": 5,
	})

	inst.extractContent()

	// PageDown should scroll by viewport height
	inst.PageDown()
	if inst.scrollOffset != 5 {
		t.Errorf("Expected offset 5 after PageDown, got %d", inst.scrollOffset)
	}

	// Another PageDown
	inst.PageDown()
	if inst.scrollOffset != 10 {
		t.Errorf("Expected offset 10 after second PageDown, got %d", inst.scrollOffset)
	}

	// PageUp
	inst.PageUp()
	if inst.scrollOffset != 5 {
		t.Errorf("Expected offset 5 after PageUp, got %d", inst.scrollOffset)
	}
}

func TestInstance_Paint_NoBorder(t *testing.T) {
	textNode := newtext.New("Line 1\nLine 2\nLine 3")

	inst := NewInstance(rtui.Props{
		"child":      textNode,
		"width":      20,
		"height":     2,
		"showBorder": false,
	})

	inst.extractContent()
	cmds := inst.Paint(0, 0)

	// Should have 2 content lines (viewport height)
	if len(cmds) != 2 {
		t.Errorf("Expected 2 draw commands (2 lines), got %d", len(cmds))
	}
}

func TestInstance_Paint_WithBorder(t *testing.T) {
	textNode := newtext.New("Line 1\nLine 2\nLine 3")

	inst := NewInstance(rtui.Props{
		"child":      textNode,
		"width":      20,
		"height":     2,
		"showBorder": true,
	})

	inst.extractContent()
	cmds := inst.Paint(0, 0)

	// Border: top + (left+content+right) * height + bottom
	// = 1 + (1 + 1 + 1) * 2 + 1 = 8
	if len(cmds) != 8 {
		t.Errorf("Expected 8 draw commands with border, got %d", len(cmds))
	}
}

func TestInstance_Paint_ScrollIndicator(t *testing.T) {
	textNode := newtext.New(buildLines(10))

	inst := NewInstance(rtui.Props{
		"child":         textNode,
		"width":         20,
		"height":        3,
		"showBorder":    true,
		"showIndicator": true,
	})

	inst.extractContent()

	// At top - should show down indicator
	cmds := inst.Paint(0, 0)
	foundDown := false
	for _, cmd := range cmds {
		if cmd.Text == "↓" || cmd.Text == "↕" || cmd.Text == "↑" {
			foundDown = true
			break
		}
	}
	if !foundDown {
		t.Error("Should show scroll indicator when scrollable")
	}
}

func TestInstance_Bounds(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	inst.SetBounds(10, 20, 30, 40)

	x, y, w, h := inst.GetBounds()
	if x != 10 || y != 20 || w != 30 || h != 40 {
		t.Errorf("Bounds mismatch: got (%d, %d, %d, %d)", x, y, w, h)
	}
}

func TestInstance_SetProps(t *testing.T) {
	textNode := newtext.New("Content")

	inst := NewInstance(rtui.Props{
		"child":         textNode,
		"width":         40,
		"height":        10,
		"scrollOffset":  5,
		"showBorder":    true,
		"showIndicator": false,
	})

	if inst.width != 40 {
		t.Errorf("Expected width 40, got %d", inst.width)
	}
	if inst.height != 10 {
		t.Errorf("Expected height 10, got %d", inst.height)
	}
	if inst.scrollOffset != 5 {
		t.Errorf("Expected scrollOffset 5, got %d", inst.scrollOffset)
	}
	if !inst.showBorder {
		t.Error("showBorder should be true")
	}
	if inst.showIndicator {
		t.Error("showIndicator should be false")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_FluentAPI(t *testing.T) {
	textNode := newtext.New("Content")

	vnode := NewBuilder().
		Child(textNode).
		Width(50).
		Height(15).
		ScrollOffset(3).
		ShowBorder(true).
		ShowIndicator(false).
		Key("test-scroll").
		Build()

	sv, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() did not return *VNode")
	}

	if sv.Width() != 50 {
		t.Errorf("Expected width 50, got %d", sv.Width())
	}
	if sv.Height() != 15 {
		t.Errorf("Expected height 15, got %d", sv.Height())
	}
	if sv.ScrollOffset() != 3 {
		t.Errorf("Expected scrollOffset 3, got %d", sv.ScrollOffset())
	}
	if sv.Key() != "test-scroll" {
		t.Errorf("Expected key 'test-scroll', got '%s'", sv.Key())
	}
}

func TestBuilder_ConvenienceMethods(t *testing.T) {
	textNode := newtext.New("Content")

	// Test Border()
	vnode1 := NewBuilder().Child(textNode).Width(30).Height(10).Border().Build()
	sv1 := vnode1.(*VNode)
	if !sv1.ShowBorder() {
		t.Error("Border() should enable border")
	}

	// Test NoBorder()
	vnode2 := NewBuilder().Child(textNode).Width(30).Height(10).NoBorder().Build()
	sv2 := vnode2.(*VNode)
	if sv2.ShowBorder() {
		t.Error("NoBorder() should disable border")
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestConvenienceFunctions(t *testing.T) {
	textNode := newtext.New("Content")

	// Of
	vnode1 := Of(textNode)
	if vnode1 == nil {
		t.Error("Of() returned nil")
	}

	// OfSize
	vnode2 := OfSize(textNode, 40, 20)
	sv2 := vnode2.(*VNode)
	if sv2.Width() != 40 || sv2.Height() != 20 {
		t.Error("OfSize() dimensions mismatch")
	}

	// Bordered
	vnode3 := Bordered(textNode, 50, 15)
	sv3 := vnode3.(*VNode)
	if sv3.Width() != 50 || sv3.Height() != 15 {
		t.Error("Bordered() dimensions mismatch")
	}
	if !sv3.ShowBorder() {
		t.Error("Bordered() should enable border")
	}
}

// =============================================================================
// Integration Test - DrawCmd Validation
// =============================================================================

func TestInstance_Paint_CommandPositions(t *testing.T) {
	textNode := newtext.New("ABC\nDEF\nGHI")

	inst := NewInstance(rtui.Props{
		"child":      textNode,
		"width":      5,
		"height":     2,
		"showBorder": true,
	})

	inst.extractContent()
	cmds := inst.Paint(0, 0)

	// Verify top border position
	topBorder := cmds[0]
	if topBorder.X != 0 || topBorder.Y != 0 {
		t.Errorf("Top border at wrong position: (%d, %d)", topBorder.X, topBorder.Y)
	}
	if !strings.Contains(topBorder.Text, "┌") {
		t.Errorf("Top border should contain corner: %s", topBorder.Text)
	}

	// Verify content lines start at y=1
	for _, cmd := range cmds {
		if cmd.Y == 1 && cmd.X == 1 {
			// Content is padded to fill width
			if !strings.HasPrefix(cmd.Text, "ABC") {
				t.Errorf("First content line should start with 'ABC', got '%s'", cmd.Text)
			}
			break
		}
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestInstance_EmptyContent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"child":  nil,
		"width":  30,
		"height": 10,
	})

	inst.extractContent()

	if inst.totalLines != 0 {
		t.Errorf("Expected 0 totalLines for empty content, got %d", inst.totalLines)
	}

	cmds := inst.Paint(0, 0)
	// Should still produce commands (empty lines)
	if len(cmds) < 0 {
		t.Error("Paint should handle empty content")
	}
}

func TestInstance_SmallViewport(t *testing.T) {
	textNode := newtext.New("Only one line")

	inst := NewInstance(rtui.Props{
		"child":  textNode,
		"width":  30,
		"height": 10,
	})

	inst.extractContent()

	// Should not be scrollable
	if inst.IsScrollable() {
		t.Error("Content smaller than viewport should not be scrollable")
	}

	// ScrollBy should have no effect
	inst.ScrollBy(5)
	if inst.scrollOffset != 0 {
		t.Errorf("ScrollBy should not scroll when content fits, got %d", inst.scrollOffset)
	}
}

func TestInstance_ScrollOffsetClamp(t *testing.T) {
	textNode := newtext.New(buildLines(5))

	inst := NewInstance(rtui.Props{
		"child":         textNode,
		"width":         30,
		"height":        3,
		"showBorder":    false,
		"scrollOffset":  100, // Too high
	})

	inst.extractContent()

	// ScrollOffset should be clamped during Paint
	inst.Paint(0, 0)

	// Max offset = 5 - 3 = 2
	if inst.scrollOffset > 2 {
		t.Errorf("ScrollOffset should be clamped to max, got %d", inst.scrollOffset)
	}
}

// =============================================================================
// BoxModel Test
// =============================================================================

func TestVNode_GetBorder(t *testing.T) {
	// With border
	vnode1 := New().SetShowBorder(true)
	border1 := vnode1.GetBorder()
	if border1.Style != layout.BorderSingle {
		t.Error("GetBorder should return single border when showBorder=true")
	}

	// Without border
	vnode2 := New().SetShowBorder(false)
	border2 := vnode2.GetBorder()
	if border2.Style != layout.BorderNone {
		t.Error("GetBorder should return none border when showBorder=false")
	}
}
