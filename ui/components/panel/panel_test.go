package panel

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newborder "github.com/wwsheng009/mint/ui/components/border"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	vnode := New()
	if vnode == nil {
		t.Fatal("New() returned nil")
	}
	if vnode.Tag() != "panel" {
		t.Errorf("Expected tag 'panel', got '%s'", vnode.Tag())
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	vnode := New()

	// Test VNode interface
	var _ rtui.VNode = vnode

	// Test InstanceFactory interface
	var _ rtui.InstanceFactory = vnode
}

func TestVNode_FluentAPI(t *testing.T) {
	textNode := newtext.New("Content")

	vnode := New().
		SetTitle("Test Panel").
		SetHeader(newtext.New("Header")).
		SetContent(textNode).
		SetFooter(newtext.New("Footer")).
		SetWidth(40).
		SetHeight(10).
		SetFlex(1).
		SetPadding(1).
		Rounded()

	if vnode.title != "Test Panel" {
		t.Errorf("Expected title 'Test Panel', got '%s'", vnode.title)
	}
	if vnode.width != 40 {
		t.Errorf("Expected width 40, got %d", vnode.width)
	}
	if vnode.height != 10 {
		t.Errorf("Expected height 10, got %d", vnode.height)
	}
	if vnode.flex != 1 {
		t.Errorf("Expected flex 1, got %d", vnode.flex)
	}
	if vnode.padding != 1 {
		t.Errorf("Expected padding 1, got %d", vnode.padding)
	}
	if vnode.borderStyle != layout.BorderRounded {
		t.Errorf("Expected rounded border style")
	}
	if vnode.content != textNode {
		t.Error("Content mismatch")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	vnode := New().
		SetTitle("Test").
		SetWidth(30).
		SetHeight(10).
		SetContent(newtext.New("Content"))

	inst := vnode.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}
	// Panel delegates to Border instance
}

func TestVNode_Children(t *testing.T) {
	header := newtext.New("Header")
	content := newtext.New("Content")
	footer := newtext.New("Footer")

	vnode := New().
		SetHeader(header).
		SetContent(content).
		SetFooter(footer)

	children := vnode.Children()
	// Panel delegates to Border which has VStack as child
	// The children should come from the composed structure
	if children == nil {
		t.Error("Children should not be nil (delegated from composed Border)")
	}
}

func TestVNode_GetComposed(t *testing.T) {
	vnode := New().
		SetTitle("Test Panel").
		SetContent(newtext.New("Content")).
		Rounded()

	composed := vnode.getComposed()
	if composed == nil {
		t.Fatal("getComposed() returned nil")
	}

	// Composed should be a Border VNode
	if composed.Tag() != "bordered" {
		t.Errorf("Expected composed tag 'bordered', got '%s'", composed.Tag())
	}
}

func TestVNode_TitleToBorderLabel(t *testing.T) {
	vnode := New().SetTitle("Test Panel")

	composed := vnode.getComposed()
	props := composed.Props()

	// Border should have label derived from title
	if props["borderLabel"] != " Test Panel " {
		t.Errorf("Expected border label ' Test Panel ', got '%s'", props["borderLabel"])
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_FluentAPI(t *testing.T) {
	vnode := NewBuilder().
		Title("Panel Title").
		Header(newtext.New("Header")).
		Content(newtext.New("Content")).
		Footer(newtext.New("Footer")).
		Width(50).
		Height(15).
		Flex(2).
		Padding(1).
		Rounded().
		BorderColor("green").
		Build()

	if vnode == nil {
		t.Fatal("Build() returned nil")
	}

	panelVNode, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("VNode is not *VNode")
	}

	if panelVNode.title != "Panel Title" {
		t.Errorf("Expected title 'Panel Title', got '%s'", panelVNode.title)
	}
	if panelVNode.width != 50 {
		t.Errorf("Expected width 50, got %d", panelVNode.width)
	}
	if panelVNode.height != 15 {
		t.Errorf("Expected height 15, got %d", panelVNode.height)
	}
	if panelVNode.flex != 2 {
		t.Errorf("Expected flex 2, got %d", panelVNode.flex)
	}
	if panelVNode.borderStyle != layout.BorderRounded {
		t.Error("Expected rounded border")
	}
	if panelVNode.borderColor != "green" {
		t.Errorf("Expected border color 'green', got '%s'", panelVNode.borderColor)
	}
}

func TestBuilder_ConvenienceMethods(t *testing.T) {
	// Test Double()
	vnode1 := NewBuilder().Double().Build()
	if vnode1.(*VNode).borderStyle != layout.BorderDouble {
		t.Error("Double() should set double border")
	}

	// Test NoBorder()
	vnode2 := NewBuilder().NoBorder().Build()
	if vnode2.(*VNode).borderStyle != layout.BorderNone {
		t.Error("NoBorder() should remove border")
	}

	// Test Size()
	vnode3 := NewBuilder().Size(60, 20).Build()
	if vnode3.(*VNode).width != 60 || vnode3.(*VNode).height != 20 {
		t.Error("Size() should set width and height")
	}

	// Test Label()
	vnode4 := NewBuilder().Label(" Custom Label ").Build()
	if vnode4.(*VNode).borderLabel != " Custom Label " {
		t.Error("Label() should set border label")
	}
}

// =============================================================================
// Convenience Functions Tests
// =============================================================================

func TestConvenienceFunctions(t *testing.T) {
	content := newtext.New("Content")

	// Test Of()
	vnode1 := Of(content)
	if vnode1 == nil {
		t.Error("Of() returned nil")
	}

	// Test OfSize()
	vnode2 := OfSize(content, 40, 15)
	panel2 := vnode2.(*VNode)
	if panel2.width != 40 || panel2.height != 15 {
		t.Error("OfSize() should set dimensions")
	}

	// Test Titled()
	vnode3 := Titled("My Panel", content)
	panel3 := vnode3.(*VNode)
	if panel3.title != "My Panel" {
		t.Error("Titled() should set title")
	}
	if panel3.borderStyle != layout.BorderRounded {
		t.Error("Titled() should use rounded border")
	}

	// Test Bordered()
	vnode4 := Bordered(content, 50, 20)
	panel4 := vnode4.(*VNode)
	if panel4.width != 50 || panel4.height != 20 {
		t.Error("Bordered() should set dimensions")
	}
}

// =============================================================================
// Border Info Tests
// =============================================================================

func TestVNode_GetBorder(t *testing.T) {
	vnode := New().
		Rounded().
		SetBorderColor("red").
		SetBorderLabel(" Test ")

	border := vnode.GetBorder()

	// Panel is a composition container that delegates to Border component
	// GetBorder() returns BorderNone to avoid double border calculation
	if border.Style != layout.BorderNone {
		t.Error("Expected BorderNone style (Panel delegates to internal Border)")
	}
	// Label is not exposed through GetBorder() - it's set on the internal Border
}

func TestVNode_GetPadding(t *testing.T) {
	vnode := New().SetPadding(2)

	padding := vnode.GetPadding()

	if padding.Top != 2 || padding.Right != 2 || padding.Bottom != 2 || padding.Left != 2 {
		t.Errorf("Expected padding 2 on all sides, got %+v", padding)
	}
}

// =============================================================================
// Style Tests
// =============================================================================

func TestVNode_Style(t *testing.T) {
	s := style.Style{FG: "white", BG: "blue"}
	vnode := New().SetStyle(s)

	if vnode.Style().FG != "white" {
		t.Error("Style FG should be white")
	}
	if vnode.Style().BG != "blue" {
		t.Error("Style BG should be blue")
	}
}

// =============================================================================
// Composition Tests
// =============================================================================

func TestVNode_CompositionWithNoBorder(t *testing.T) {
	vnode := New().
		SetContent(newtext.New("Content")).
		NoBorder()

	composed := vnode.getComposed()
	props := composed.Props()

	if props["borderStyle"] != layout.BorderNone {
		t.Error("NoBorder should set BorderNone style")
	}
}

func TestVNode_CompositionPassesDimensions(t *testing.T) {
	vnode := New().
		SetWidth(40).
		SetHeight(10).
		SetFlex(2).
		SetContent(newtext.New("Content"))

	// Test Panel's own props (not composed Border's props)
	panelProps := vnode.Props()

	if panelProps["width"] != 40 {
		t.Errorf("Panel width should be 40, got %v", panelProps["width"])
	}
	if panelProps["height"] != 10 {
		t.Errorf("Panel height should be 10, got %v", panelProps["height"])
	}
	if panelProps["flex"] != 2 {
		t.Errorf("Panel flex should be 2, got %v", panelProps["flex"])
	}

	// Composed Border gets inner dimensions (minus border padding)
	composed := vnode.getComposed()
	borderProps := composed.Props()

	borderPadding := 2 * newborder.GetBorderWidth(vnode.borderStyle)
	expectedInnerWidth := 40 - borderPadding // 40 - 2 = 38 for single border
	expectedInnerHeight := 10 - borderPadding // 10 - 2 = 8 for single border

	if borderProps["width"] != expectedInnerWidth {
		t.Errorf("Border inner width should be %d (40 - %d), got %v", expectedInnerWidth, borderPadding, borderProps["width"])
	}
	if borderProps["height"] != expectedInnerHeight {
		t.Errorf("Border inner height should be %d (10 - %d), got %v", expectedInnerHeight, borderPadding, borderProps["height"])
	}
}

func TestVNode_CustomBorderLabel(t *testing.T) {
	vnode := New().
		SetTitle("Title").
		SetBorderLabel("Custom Label").
		SetContent(newtext.New("Content"))

	composed := vnode.getComposed()
	props := composed.Props()

	// Custom label should override title-derived label
	if props["borderLabel"] != "Custom Label" {
		t.Errorf("Expected custom label 'Custom Label', got '%s'", props["borderLabel"])
	}
}

func TestVNode_DefaultContent(t *testing.T) {
	// Panel without content should use empty text
	vnode := New().SetTitle("Empty Panel")

	composed := vnode.getComposed()
	children := composed.Children()

	if len(children) == 0 {
		t.Error("Composed Border should have VStack child")
	}

	// Title should appear in border label, not as header child
	props := composed.Props()
	if props["borderLabel"] != " Empty Panel " {
		t.Errorf("Expected border label ' Empty Panel ', got '%s'", props["borderLabel"])
	}
}

// TestVNode_CreateInstanceFromPanelWithLabel tests full creation flow
func TestVNode_CreateInstanceFromPanelWithLabel(t *testing.T) {
	vnode := New().
		SetTitle("My Panel").
		SetContent(newtext.New("Content")).
		Rounded()

	// Get composed Border
	composed := vnode.getComposed()
	borderProps := composed.Props()

	// Check Border props
	if borderProps["borderLabel"] != " My Panel " {
		t.Errorf("Expected borderLabel ' My Panel ', got '%v'", borderProps["borderLabel"])
	}

	// Create instance
	inst := vnode.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	// Check if it's a border Instance
	props := inst.GetProps()
	if props["borderLabel"] != " My Panel " {
		t.Errorf("Instance props borderLabel should be ' My Panel ', got '%v'", props["borderLabel"])
	}
}
