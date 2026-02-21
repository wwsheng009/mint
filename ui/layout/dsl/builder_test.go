package dsl

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Node Basics Tests
// =============================================================================

func TestNode_Tag_Props_Children(t *testing.T) {
	props := ui.Props{"title": "Test"}
	children := []Node{Text("Child")}
	node := Node{tag: "panel", props: props, children: children}

	if node.Tag() != "panel" {
		t.Errorf("expected tag 'panel', got '%s'", node.Tag())
	}
	if node.Props()["title"] != "Test" {
		t.Errorf("expected title 'Test', got '%v'", node.Props()["title"])
	}
	if len(node.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(node.Children()))
	}
}

// =============================================================================
// Factory Functions Tests
// =============================================================================

func TestPanel(t *testing.T) {
	props := ui.Props{"width": 30, "height": 10}
	node := Panel(props, Text("Content"))

	if node.Tag() != "panel" {
		t.Errorf("expected tag 'panel', got '%s'", node.Tag())
	}
	if node.Props()["width"] != 30 {
		t.Errorf("expected width 30, got %d", node.Props()["width"])
	}
	if len(node.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(node.Children()))
	}
}

func TestText(t *testing.T) {
	node := Text("Hello World")

	if node.Tag() != "text" {
		t.Errorf("expected tag 'text', got '%s'", node.Tag())
	}
	if node.Props()["content"] != "Hello World" {
		t.Errorf("expected content 'Hello World', got '%v'", node.Props()["content"])
	}
}

func TestRow(t *testing.T) {
	props := ui.Props{"flex": 1}
	node := Row(props, Text("Item1"), Text("Item2"))

	if node.Tag() != "row" {
		t.Errorf("expected tag 'row', got '%s'", node.Tag())
	}
	if node.Props()["_layout"] != "row" {
		t.Errorf("expected _layout 'row', got '%v'", node.Props()["_layout"])
	}
	if len(node.Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(node.Children()))
	}
}

func TestColumn(t *testing.T) {
	props := ui.Props{}
	node := Column(props, Text("Item1"), Text("Item2"))

	if node.Tag() != "column" {
		t.Errorf("expected tag 'column', got '%s'", node.Tag())
	}
	if node.Props()["_layout"] != "column" {
		t.Errorf("expected _layout 'column', got '%v'", node.Props()["_layout"])
	}
	if len(node.Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(node.Children()))
	}
}

// =============================================================================
// PropsBuilder Tests
// =============================================================================

func TestPropsBuilder_WidthHeight(t *testing.T) {
	props := NewProps().Width(30).Height(10).Build()

	if props.GetInt("width") != 30 {
		t.Errorf("expected width 30, got %d", props.GetInt("width"))
	}
	if props.GetInt("height") != 10 {
		t.Errorf("expected height 10, got %d", props.GetInt("height"))
	}
}

func TestPropsBuilder_Flex(t *testing.T) {
	props := NewProps().Flex(1).Build()

	if props.GetInt("flex") != 1 {
		t.Errorf("expected flex 1, got %d", props.GetInt("flex"))
	}
}

func TestPropsBuilder_Padding(t *testing.T) {
	props := NewProps().Padding(2).Build()

	if props.GetInt("padding") != 2 {
		t.Errorf("expected padding 2, got %d", props.GetInt("padding"))
	}
}

func TestPropsBuilder_Title(t *testing.T) {
	props := NewProps().Title("My Panel").Build()

	if props.GetString("title") != "My Panel" {
		t.Errorf("expected title 'My Panel', got '%s'", props.GetString("title"))
	}
}

func TestPropsBuilder_Complex(t *testing.T) {
	props := NewProps().
		Width(30).
		Height(10).
		Flex(1).
		Padding(2).
		Title("Title").
		Build()

	if props.GetInt("width") != 30 {
		t.Errorf("expected width 30")
	}
	if props.GetInt("height") != 10 {
		t.Errorf("expected height 10")
	}
	if props.GetInt("flex") != 1 {
		t.Errorf("expected flex 1")
	}
	if props.GetInt("padding") != 2 {
		t.Errorf("expected padding 2")
	}
	if props.GetString("title") != "Title" {
		t.Errorf("expected title 'Title'")
	}
}

func TestPropsBuilder_Set(t *testing.T) {
	props := NewProps().Set("custom", "value").Build()

	if props.Get("custom") != "value" {
		t.Errorf("expected custom 'value', got '%v'", props.Get("custom"))
	}
}

// =============================================================================
// Layout Shortcut Tests
// =============================================================================

func TestFlexWidth(t *testing.T) {
	props := FlexWidth(1)

	if props.GetInt("flex") != 1 {
		t.Errorf("expected flex 1, got %d", props.GetInt("flex"))
	}
}

func TestFixedWidth(t *testing.T) {
	props := FixedWidth(20)

	if props.GetInt("width") != 20 {
		t.Errorf("expected width 20, got %d", props.GetInt("width"))
	}
}

func TestFixedSize(t *testing.T) {
	props := FixedSize(30, 10)

	if props.GetInt("width") != 30 {
		t.Errorf("expected width 30")
	}
	if props.GetInt("height") != 10 {
		t.Errorf("expected height 10")
	}
}

func TestAutoSize(t *testing.T) {
	props := AutoSize()

	if len(props) != 0 {
		t.Errorf("expected empty props, got %v", props)
	}
}

// =============================================================================
// Component Shortcut Tests
// =============================================================================

func TestInfoBox(t *testing.T) {
	node := InfoBox("Info", "This is info")

	if node.Tag() != "panel" {
		t.Errorf("expected tag 'panel', got '%s'", node.Tag())
	}
	if node.Props()["title"] != "Info" {
		t.Errorf("expected title 'Info'")
	}
	if len(node.Children()) != 1 {
		t.Errorf("expected 1 child")
	}
}

func TestErrorBox(t *testing.T) {
	node := ErrorBox("Error", "Something went wrong")

	if node.Tag() != "panel" {
		t.Errorf("expected tag 'panel', got '%s'", node.Tag())
	}
	if node.Props()["title"] != "Error" {
		t.Errorf("expected title 'Error'")
	}
}

func TestSuccessBox(t *testing.T) {
	node := SuccessBox("Success", "Operation completed")

	if node.Tag() != "panel" {
		t.Errorf("expected tag 'panel', got '%s'", node.Tag())
	}
	if node.Props()["title"] != "Success" {
		t.Errorf("expected title 'Success'")
	}
}

func TestWarningBox(t *testing.T) {
	node := WarningBox("Warning", "Caution advised")

	if node.Tag() != "panel" {
		t.Errorf("expected tag 'panel', got '%s'", node.Tag())
	}
	if node.Props()["title"] != "Warning" {
		t.Errorf("expected title 'Warning'")
	}
}

// =============================================================================
// ToVNode Tests
// =============================================================================

func TestNode_ToVNode_Text(t *testing.T) {
	node := Text("Hello")
	vnode := node.ToVNode()

	if vnode == nil {
		t.Fatal("ToVNode should not return nil")
	}

	textNode, ok := vnode.(*text.VNode)
	if !ok {
		t.Errorf("expected *text.VNode, got %T", vnode)
	}
	if textNode == nil {
		t.Error("textNode should not be nil")
	}
}

func TestNode_ToVNode_Panel(t *testing.T) {
	props := NewProps().
		Title("My Panel").
		Width(30).
		Height(10).
		Build()
	node := Panel(props, Text("Content"))
	vnode := node.ToVNode()

	if vnode == nil {
		t.Fatal("ToVNode should not return nil")
	}
	if vnode.Tag() != "panel" {
		t.Errorf("expected tag 'panel', got '%s'", vnode.Tag())
	}
}

// =============================================================================
// String Representation Tests
// =============================================================================

func TestNode_String_Simple(t *testing.T) {
	node := Text("Hello")

	str := node.String()
	if str == "" {
		t.Error("String should not be empty")
	}
	// Check for escaped quote in Go format
}

func TestNode_String_Complex(t *testing.T) {
	props := NewProps().Width(30).Height(10).Title("Panel").Build()
	node := Panel(props, Text("Content"))

	str := node.String()
	if str == "" {
		t.Error("String should not be empty")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestDSL_ComplexLayout(t *testing.T) {
	layout := Column(
		NewProps().Flex(1).Build(),
		Panel(
			NewProps().Title("Header").Height(3).Build(),
			Text("Header Content"),
		),
		Row(
			NewProps().Flex(1).Build(),
			Panel(
				NewProps().Title("Sidebar").Width(20).Build(),
				Text("Sidebar Content"),
			),
			Panel(
				NewProps().Title("Main").Flex(1).Build(),
				Text("Main Content"),
			),
		),
		Panel(
			NewProps().Title("Footer").Height(3).Build(),
			Text("Footer Content"),
		),
	)

	if layout.Tag() != "column" {
		t.Errorf("expected tag 'column', got '%s'", layout.Tag())
	}
	if len(layout.Children()) != 3 {
		t.Errorf("expected 3 children, got %d", len(layout.Children()))
	}
}

func TestDSL_NestedPanels(t *testing.T) {
	node := Panel(
		NewProps().Title("Outer").Width(40).Height(20).Build(),
		Panel(
			NewProps().Title("Inner").Width(30).Height(10).Build(),
			Text("Inner Content"),
		),
	)

	if node.Tag() != "panel" {
		t.Errorf("expected tag 'panel'")
	}
	if len(node.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(node.Children()))
	}
}

// =============================================================================
// Style Props Tests
// =============================================================================

func TestPropsBuilder_Border(t *testing.T) {
	props := NewProps().
		BorderStyle(layout.BorderSingle).
		BorderColor(style.Color("red")).
		Build()

	if props.Get("borderStyle") == nil {
		t.Error("borderStyle should be set")
	}
	if props.Get("borderColor") == nil {
		t.Error("borderColor should be set")
	}
}

func TestPropsBuilder_Colors(t *testing.T) {
	props := NewProps().
		Color(style.Color("blue")).
		Background(style.Color("white")).
		Build()

	if props.Get("color") == nil {
		t.Error("color should be set")
	}
	if props.Get("background") == nil {
		t.Error("background should be set")
	}
}
