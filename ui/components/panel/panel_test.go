package panel

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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

	// Composed should be a VStack with native border properties (no longer Border VNode)
	if composed.Tag() != "vstack" {
		t.Errorf("Expected composed tag 'vstack', got '%s'", composed.Tag())
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

func TestOperationsPanel(t *testing.T) {
	node := OperationsPanel("Runtime Operations", []rtui.VNode{
		newtext.New("Effective Sections"),
		newtext.New("Pending Restart"),
	}, "runtime diagnostics unavailable", 126)
	panelNode := node.(*VNode)
	if panelNode.title != "Runtime Operations" {
		t.Fatalf("title = %q, want Runtime Operations", panelNode.title)
	}
	if panelNode.width != 126 {
		t.Fatalf("width = %d, want 126", panelNode.width)
	}
	if panelNode.borderStyle != layout.BorderSingle {
		t.Fatalf("borderStyle = %v, want single", panelNode.borderStyle)
	}
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	if len(content.Children()) != 2 {
		t.Fatalf("content children = %d, want 2", len(content.Children()))
	}
}

func TestStackPanelWithScope(t *testing.T) {
	node := StackPanelWithScope("Distribution", []rtui.VNode{
		newtext.New("chart"),
	}, "source=analytics 15m top 5", "distribution unavailable", 126)
	panelNode := node.(*VNode)
	if panelNode.title != "Distribution" {
		t.Fatalf("title = %q, want Distribution", panelNode.title)
	}
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("content children = %d, want content node and scope", len(children))
	}
	if got := children[1].Props()["content"]; got != "Scope: source=analytics 15m top 5" {
		t.Fatalf("scope content = %v, want Scope line", got)
	}
}

func TestTablePanelWithScope(t *testing.T) {
	node := TablePanelWithScope("Provider Summary", newtext.New("table"), "provider=all / sort=rows tokens desc", "summary unavailable", 72)
	panelNode := node.(*VNode)
	if panelNode.title != "Provider Summary" {
		t.Fatalf("title = %q, want Provider Summary", panelNode.title)
	}
	if panelNode.width != 72 {
		t.Fatalf("width = %d, want 72", panelNode.width)
	}
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("content children = %d, want table node and scope", len(children))
	}
	if got := children[1].Props()["content"]; got != "Scope: provider=all / sort=rows tokens desc" {
		t.Fatalf("scope content = %v, want Scope line", got)
	}
}

func TestTablePanelWithScopeKeepsEmptyStateWhenContentIsNil(t *testing.T) {
	node := TablePanelWithScope("Provider Summary", nil, "provider=all", "summary unavailable", 72)
	panelNode := node.(*VNode)
	if panelNode.content == nil || panelNode.content.Tag() != "empty" {
		t.Fatalf("content = %T/%q, want empty", panelNode.content, panelNode.content.Tag())
	}
	if got := panelNode.content.Props()["description"]; got != "summary unavailable" {
		t.Fatalf("empty description = %v, want summary unavailable", got)
	}
}

func TestStackPanelWithScopeKeepsEmptyStateWhenNodesAreEmpty(t *testing.T) {
	node := StackPanelWithScope("Distribution", nil, "source=analytics", "distribution unavailable", 126)
	panelNode := node.(*VNode)
	children := panelNode.content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want empty state only", len(children))
	}
	if children[0].Tag() != "empty" {
		t.Fatalf("empty child tag = %q, want empty", children[0].Tag())
	}
	if got := children[0].Props()["description"]; got != "distribution unavailable" {
		t.Fatalf("empty description = %v, want distribution unavailable", got)
	}
}

func TestOperationsPanelEmptyState(t *testing.T) {
	node := OperationsPanel("Runtime Operations", nil, "runtime diagnostics unavailable", 126)
	panelNode := node.(*VNode)
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want empty state", len(children))
	}
	if children[0].Tag() != "empty" {
		t.Fatalf("empty child tag = %q, want empty", children[0].Tag())
	}
	if got := children[0].Props()["description"]; got != "runtime diagnostics unavailable" {
		t.Fatalf("empty description = %v, want runtime diagnostics unavailable", got)
	}
}

func TestOperationsPanelWithScope(t *testing.T) {
	node := OperationsPanelWithScope("Runtime Operations", []rtui.VNode{
		newtext.New("Effective Sections"),
	}, "runtime=available / concurrency=unavailable", "runtime diagnostics unavailable", 126)
	panelNode := node.(*VNode)
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("content children = %d, want progress node and scope", len(children))
	}
	if got := children[1].Props()["content"]; got != "Scope: runtime=available / concurrency=unavailable" {
		t.Fatalf("scope content = %v, want Scope line", got)
	}
}

func TestOperationsPanelWithScopeSkipsBlankScope(t *testing.T) {
	node := OperationsPanelWithScope("Runtime Operations", []rtui.VNode{
		newtext.New("Effective Sections"),
	}, "  ", "runtime diagnostics unavailable", 126)
	panelNode := node.(*VNode)
	children := panelNode.content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want original nodes only", len(children))
	}
}

func TestOperationsPanelWithScopeKeepsEmptyStateWhenNodesAreEmpty(t *testing.T) {
	node := OperationsPanelWithScope("Runtime Operations", nil, "runtime=available", "runtime diagnostics unavailable", 126)
	panelNode := node.(*VNode)
	children := panelNode.content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want empty state only", len(children))
	}
	if children[0].Tag() != "empty" {
		t.Fatalf("empty child tag = %q, want empty", children[0].Tag())
	}
}

func TestContentPanel(t *testing.T) {
	node := ContentPanel("Actions", newtext.New("buttons"), "actions unavailable", 126)
	panelNode := node.(*VNode)
	if panelNode.title != "Actions" {
		t.Fatalf("title = %q, want Actions", panelNode.title)
	}
	if panelNode.width != 126 {
		t.Fatalf("width = %d, want 126", panelNode.width)
	}
	if panelNode.borderStyle != layout.BorderSingle {
		t.Fatalf("borderStyle = %v, want single", panelNode.borderStyle)
	}
	if panelNode.content == nil || panelNode.content.Tag() != "text" {
		t.Fatalf("content = %T, want text", panelNode.content)
	}
}

func TestContentPanelEmptyState(t *testing.T) {
	node := ContentPanel("Actions", nil, "actions unavailable", 126)
	panelNode := node.(*VNode)
	if panelNode.content == nil || panelNode.content.Tag() != "empty" {
		t.Fatalf("content = %T, want empty", panelNode.content)
	}
	if got := panelNode.content.Props()["description"]; got != "actions unavailable" {
		t.Fatalf("empty description = %v, want actions unavailable", got)
	}
}

func TestStackPanel(t *testing.T) {
	node := StackPanel("Alert History", []rtui.VNode{
		newtext.New("Table"),
		newtext.New("Pagination"),
		newtext.New("Selection"),
	}, "history unavailable", 62)
	panelNode := node.(*VNode)
	if panelNode.title != "Alert History" {
		t.Fatalf("title = %q, want Alert History", panelNode.title)
	}
	if panelNode.width != 62 {
		t.Fatalf("width = %d, want 62", panelNode.width)
	}
	if panelNode.borderStyle != layout.BorderSingle {
		t.Fatalf("borderStyle = %v, want single", panelNode.borderStyle)
	}
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	if len(content.Children()) != 3 {
		t.Fatalf("content children = %d, want 3", len(content.Children()))
	}
}

func TestStackPanelSkipsNilOptionalNodes(t *testing.T) {
	node := StackPanel("Alert History", []rtui.VNode{
		newtext.New("Table"),
		nil,
		newtext.New("Selection"),
	}, "history unavailable", 62)
	panelNode := node.(*VNode)
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("content children = %d, want nil optional node skipped", len(children))
	}
	if children[0].Tag() != "text" || children[1].Tag() != "text" {
		t.Fatalf("children tags = %q/%q, want text/text", children[0].Tag(), children[1].Tag())
	}
}

func TestStackPanelEmptyState(t *testing.T) {
	node := StackPanel("Alert Rules", nil, "rules unavailable", 62)
	panelNode := node.(*VNode)
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want empty state", len(children))
	}
	if children[0].Tag() != "empty" {
		t.Fatalf("empty child tag = %q, want empty", children[0].Tag())
	}
	if got := children[0].Props()["description"]; got != "rules unavailable" {
		t.Fatalf("empty description = %v, want rules unavailable", got)
	}
}

func TestStackPanelAllNilUsesEmptyState(t *testing.T) {
	node := StackPanel("Alert Rules", []rtui.VNode{nil, nil}, "rules unavailable", 62)
	panelNode := node.(*VNode)
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 1 || children[0].Tag() != "empty" {
		t.Fatalf("children = %d first tag = %q, want empty state", len(children), children[0].Tag())
	}
}

func TestPaddedStackPanel(t *testing.T) {
	node := PaddedStackPanel("Login", []rtui.VNode{
		newtext.New("Gateway Base URL"),
		newtext.New("Captcha"),
		newtext.New("Login"),
	}, "login form unavailable", 92, 1)
	panelNode := node.(*VNode)
	if panelNode.title != "Login" {
		t.Fatalf("title = %q, want Login", panelNode.title)
	}
	if panelNode.width != 92 {
		t.Fatalf("width = %d, want 92", panelNode.width)
	}
	if panelNode.padding != 1 {
		t.Fatalf("padding = %d, want 1", panelNode.padding)
	}
	if panelNode.borderStyle != layout.BorderSingle {
		t.Fatalf("borderStyle = %v, want single", panelNode.borderStyle)
	}
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	if len(content.Children()) != 3 {
		t.Fatalf("content children = %d, want 3", len(content.Children()))
	}
}

func TestPaddedStackPanelEmptyState(t *testing.T) {
	node := PaddedStackPanel("Login", nil, "login form unavailable", 92, 1)
	panelNode := node.(*VNode)
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want empty state", len(children))
	}
	if children[0].Tag() != "empty" {
		t.Fatalf("empty child tag = %q, want empty", children[0].Tag())
	}
	if got := children[0].Props()["description"]; got != "login form unavailable" {
		t.Fatalf("empty description = %v, want login form unavailable", got)
	}
}

func TestPanelRow(t *testing.T) {
	node := PanelRow(
		TablePanel("Left", newtext.New("left"), 40),
		nil,
		TablePanel("Right", newtext.New("right"), 60),
	)
	if node.Tag() != "hstack" {
		t.Fatalf("tag = %q, want hstack", node.Tag())
	}
	layoutNode, ok := node.(*rtui.LayoutNode)
	if !ok {
		t.Fatalf("node = %T, want *ui.LayoutNode", node)
	}
	if layoutNode.Direction() != rtui.DirectionRow {
		t.Fatalf("direction = %v, want row", layoutNode.Direction())
	}
	if layoutNode.Gap() != 1 {
		t.Fatalf("gap = %d, want 1", layoutNode.Gap())
	}
	if len(layoutNode.Children()) != 2 {
		t.Fatalf("children = %d, want 2 non-nil panels", len(layoutNode.Children()))
	}
	for index, child := range layoutNode.Children() {
		if child.Tag() != "panel" {
			t.Fatalf("child %d tag = %q, want panel", index, child.Tag())
		}
	}
}

func TestLinesPanel(t *testing.T) {
	node := LinesPanel("Capabilities", []Line{
		SuccessLine("ops-monitor/overview: available"),
		MutedLine("applied filters: group=default"),
		WarningLine("unsupported filters: provider"),
		TextLine("", "red"),
	}, "No capabilities.", 126)
	panelNode := node.(*VNode)
	if panelNode.title != "Capabilities" {
		t.Fatalf("title = %q, want Capabilities", panelNode.title)
	}
	if panelNode.width != 126 {
		t.Fatalf("width = %d, want 126", panelNode.width)
	}
	if panelNode.borderStyle != layout.BorderSingle {
		t.Fatalf("borderStyle = %v, want single", panelNode.borderStyle)
	}
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 3 {
		t.Fatalf("content children = %d, want 3 non-empty lines", len(children))
	}
	wantColors := []style.Color{"green", "gray", "yellow"}
	for i, want := range wantColors {
		lineStyle, ok := children[i].Props()["style"].(style.Style)
		if !ok {
			t.Fatalf("line %d style = %T, want style.Style", i, children[i].Props()["style"])
		}
		if lineStyle.FG != want {
			t.Fatalf("line %d fg = %q, want %q", i, lineStyle.FG, want)
		}
	}
}

func TestLinesPanelFallback(t *testing.T) {
	node := LinesPanel("Recommended Drilldown", []Line{
		TextLine("  ", "green"),
	}, "No drilldown links returned.", 126)
	panelNode := node.(*VNode)
	content := panelNode.content
	children := content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want fallback line", len(children))
	}
	if got := children[0].Props()["content"]; got != "No drilldown links returned." {
		t.Fatalf("fallback line = %v, want No drilldown links returned.", got)
	}
	lineStyle, ok := children[0].Props()["style"].(style.Style)
	if !ok {
		t.Fatalf("fallback style = %T, want style.Style", children[0].Props()["style"])
	}
	if lineStyle.FG != "gray" {
		t.Fatalf("fallback fg = %q, want gray", lineStyle.FG)
	}
}

func TestNoticePanel(t *testing.T) {
	node := NoticePanel("Boundary", []string{
		"This page is read-only in this phase.",
		"",
		"Follow-up write flows require reason and confirmation.",
	}, 62)
	panelNode := node.(*VNode)
	if panelNode.title != "Boundary" {
		t.Fatalf("title = %q, want Boundary", panelNode.title)
	}
	if panelNode.width != 62 {
		t.Fatalf("width = %d, want 62", panelNode.width)
	}
	content := panelNode.content
	if content == nil || content.Tag() != "vstack" {
		t.Fatalf("content = %T, want vstack", content)
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("content children = %d, want 2 non-empty notice lines", len(children))
	}
	lineStyle, ok := children[0].Props()["style"].(style.Style)
	if !ok {
		t.Fatalf("first notice line style = %T, want style.Style", children[0].Props()["style"])
	}
	if children[0].Tag() != "text" || lineStyle.FG != "bright-black" {
		t.Fatalf("first notice line = %s %+v, want muted text", children[0].Tag(), children[0].Props())
	}
}

func TestNoticePanelFallback(t *testing.T) {
	node := NoticePanel("Boundary", nil, 62)
	panelNode := node.(*VNode)
	content := panelNode.content
	children := content.Children()
	if len(children) != 1 {
		t.Fatalf("content children = %d, want fallback notice", len(children))
	}
	if got := children[0].Props()["content"]; got != "No notice." {
		t.Fatalf("fallback notice = %v, want No notice.", got)
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

	// BorderNone causes Props() to delete the "borderStyle" key (no border = absent key)
	bs, exists := props["borderStyle"]
	if exists && bs != "none" && bs != nil {
		t.Errorf("NoBorder should not set a non-none border style, got %v", bs)
	}
}

func TestVNode_CompositionPassesDimensions(t *testing.T) {
	vnode := New().
		SetWidth(40).
		SetHeight(10).
		SetFlex(2).
		SetContent(newtext.New("Content"))

	// Test Panel's own props (not composed Stack's props)
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

	// Composed VStack now gets the same dimensions as Panel
	// Stack's native border properties handle border padding internally
	composed := vnode.getComposed()
	stackProps := composed.Props()

	if stackProps["width"] != 40 {
		t.Errorf("Stack width should be 40, got %v", stackProps["width"])
	}
	if stackProps["height"] != 10 {
		t.Errorf("Stack height should be 10, got %v", stackProps["height"])
	}
	if stackProps["flex"] != 2 {
		t.Errorf("Stack flex should be 2, got %v", stackProps["flex"])
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
}
