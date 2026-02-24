package treeview

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	vnode := New()
	if vnode == nil {
		t.Fatal("New() returned nil")
	}
	if vnode.Tag() != "treeview" {
		t.Errorf("Expected tag 'treeview', got '%s'", vnode.Tag())
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	vnode := New()

	var _ rtui.VNode = vnode
	var _ rtui.InstanceFactory = vnode
}

func TestVNode_FluentAPI(t *testing.T) {
	nodes := []TreeNode{
		{Content: "root", NodeID: 0, NodeType: "folder"},
		{Indent: 4, Content: "child", NodeID: 1, NodeType: "file"},
	}

	vnode := New()
	vnode.key = "test-tree"
	vnode.SetNodes(nodes).
		SetExpandLevel(1).
		SetShowIcons(true).
		SetShowLineNums(false).
		SetCompact(false).
		SetSelectedIndex(0).
		SetViewportHeight(10).
		SetAllowScroll(true).
		SetAllowExpand(true)

	if vnode.key != "test-tree" {
		t.Errorf("Expected key 'test-tree', got '%s'", vnode.key)
	}
	if len(vnode.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(vnode.nodes))
	}
	if vnode.expandLevel != 1 {
		t.Errorf("Expected expandLevel 1, got %d", vnode.expandLevel)
	}
	if !vnode.showIcons {
		t.Error("Expected showIcons true")
	}
	if vnode.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", vnode.selectedIndex)
	}
}

func TestVNode_FromLines(t *testing.T) {
	lines := []string{
		"root",
		"    child1",
		"    child2",
		"        grandchild",
	}

	vnode := New().FromLines(lines)

	if len(vnode.nodes) != 4 {
		t.Errorf("Expected 4 nodes, got %d", len(vnode.nodes))
	}
	if vnode.nodes[0].Content != "root" {
		t.Errorf("Expected 'root', got '%s'", vnode.nodes[0].Content)
	}
	if vnode.nodes[1].Indent != 4 {
		t.Errorf("Expected indent 4 for child1, got %d", vnode.nodes[1].Indent)
	}
}

func TestVNode_AddNode(t *testing.T) {
	vnode := New()

	if len(vnode.nodes) != 0 {
		t.Errorf("Expected empty nodes initially, got %d", len(vnode.nodes))
	}

	vnode.AddNode(TreeNode{Content: "root", NodeID: 0})
	if len(vnode.nodes) != 1 {
		t.Errorf("Expected 1 node after AddNode, got %d", len(vnode.nodes))
	}

	vnode.AddNode(TreeNode{Indent: 4, Content: "child", NodeID: 1})
	if len(vnode.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(vnode.nodes))
	}
}

func TestVNode_Setters(t *testing.T) {
	treeStyle := style.Style{FG: style.Cyan}
	selectedStyle := style.Style{FG: style.White, BG: style.Blue}
	iconStyle := style.Style{FG: style.Yellow}

	vnode := New().
		SetTreeStyle(treeStyle).
		SetSelectedStyle(selectedStyle).
		SetIconStyle(iconStyle).
		SetScrollOffset(5)

	if vnode.treeStyle.FG != style.Cyan {
		t.Errorf("Expected treeStyle.FG cyan, got %s", vnode.treeStyle.FG)
	}
	if vnode.selectedStyle.FG != style.White {
		t.Errorf("Expected selectedStyle.FG white, got %s", vnode.selectedStyle.FG)
	}
	if vnode.scrollOffset != 5 {
		t.Errorf("Expected scrollOffset 5, got %d", vnode.scrollOffset)
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

func TestInstance_ImplementsInterfaces(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	var _ rtui.ComponentInstance = inst
	var _ rtui.PaintableInstance = inst
	var _ rtui.ActionHandlerInstance = inst
}

func TestInstance_DefaultProps(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	if inst.expandLevel != 1 {
		t.Errorf("Expected default expandLevel 1, got %d", inst.expandLevel)
	}
	if !inst.showIcons {
		t.Error("Expected default showIcons true")
	}
	if inst.viewportHeight != 10 {
		t.Errorf("Expected default viewportHeight 10, got %d", inst.viewportHeight)
	}
	if inst.selectedIndex != -1 {
		t.Errorf("Expected default selectedIndex -1, got %d", inst.selectedIndex)
	}
	if !inst.allowScroll {
		t.Error("Expected default allowScroll true")
	}
	if !inst.allowExpand {
		t.Error("Expected default allowExpand true")
	}
}

func TestInstance_Props(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0},
			{Indent: 4, Content: "child", NodeID: 1},
		},
		"expandLevel":   2,
		"showIcons":     false,
		"viewportHeight": 15,
		"selectedIndex":  1,
	}

	inst := NewInstance(props)

	if len(inst.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(inst.nodes))
	}
	if inst.expandLevel != 2 {
		t.Errorf("Expected expandLevel 2, got %d", inst.expandLevel)
	}
	if inst.showIcons {
		t.Error("Expected showIcons false")
	}
	if inst.viewportHeight != 15 {
		t.Errorf("Expected viewportHeight 15, got %d", inst.viewportHeight)
	}
	if inst.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", inst.selectedIndex)
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0},
		},
	})

	constraints := layout.Constraints{
		MinWidth:   40,
		MaxWidth:   80,
		MinHeight:  10,
		MaxHeight:  30,
	}

	size := inst.Measure(constraints)
	if size.Width < 40 {
		t.Errorf("Expected width >= 40, got %d", size.Width)
	}
	if size.Height < 10 {
		t.Errorf("Expected height >= 10, got %d", size.Height)
	}
}

func TestInstance_NavigateUp(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0},
			{Indent: 4, Content: "child1", NodeID: 1},
			{Indent: 4, Content: "child2", NodeID: 2},
		},
		"selectedIndex": 2,
	}
	inst := NewInstance(props)

	result := inst.navigateUp()
	if !result {
		t.Error("navigateUp() should return true")
	}
	if inst.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after navigateUp, got %d", inst.selectedIndex)
	}
}

func TestInstance_NavigateDown(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0},
			{Indent: 4, Content: "child1", NodeID: 1},
			{Indent: 4, Content: "child2", NodeID: 2},
		},
		"selectedIndex": 0,
	}
	inst := NewInstance(props)

	result := inst.navigateDown()
	if !result {
		t.Error("navigateDown() should return true")
	}
	if inst.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after navigateDown, got %d", inst.selectedIndex)
	}
}

func TestInstance_ToggleExpand(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, NodeType: "folder"},
		},
		"selectedIndex": 0,
	}
	inst := NewInstance(props)

	// Initially expanded by default (expandLevel=1)
	if !inst.isExpanded(0) {
		t.Error("Expected node 0 to be expanded initially")
	}

	// Toggle to collapse
	result := inst.toggleExpand()
	if result {
		t.Error("toggleExpand() should return false when collapsing")
	}

	// Toggle to expand
	result = inst.toggleExpand()
	if !result {
		t.Error("toggleExpand() should return true when expanding")
	}
}

func TestInstance_GetVisibleNodes(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "child1", NodeID: 1},
			{Indent: 8, Content: "grandchild", NodeID: 2},
			{Indent: 4, Content: "child2", NodeID: 3},
		},
		"expandLevel": 1,
		"allowExpand": true,
	}
	inst := NewInstance(props)

	visible := inst.getVisibleNodes()

	// With expandLevel=1, root is expanded, showing only root and its direct children
	if len(visible) < 2 {
		t.Errorf("Expected at least 2 visible nodes (root + child1), got %d", len(visible))
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_New(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("NewBuilder() returned nil")
	}
	if builder.vnode == nil {
		t.Error("Builder vnode is nil")
	}
}

func TestBuilder_FluentAPI(t *testing.T) {
	nodes := []TreeNode{
		{Content: "root", NodeID: 0, NodeType: "folder"},
		{Indent: 4, Content: "child", NodeID: 1},
	}

	vnode := NewBuilder().
		Key("test").
		Nodes(nodes).
		ExpandLevel(1).
		ShowIcons(true).
		ViewportHeight(10).
		SelectedIndex(0).
		BuildVNode()

	if vnode.key != "test" {
		t.Errorf("Expected key 'test', got '%s'", vnode.key)
	}
	if len(vnode.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(vnode.nodes))
	}
	if vnode.viewportHeight != 10 {
		t.Errorf("Expected viewportHeight 10, got %d", vnode.viewportHeight)
	}
}

func TestBuilder_FromLines(t *testing.T) {
	lines := []string{
		"root",
		"    child",
	}

	vnode := NewBuilder().FromLines(lines).BuildVNode()

	if len(vnode.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(vnode.nodes))
	}
}

func TestBuilder_Build(t *testing.T) {
	vnode := NewBuilder().Build()

	if vnode == nil {
		t.Fatal("Build() returned nil")
	}

	var _ rtui.VNode = vnode
}

func TestBuilder_BuildInstance(t *testing.T) {
	inst := NewBuilder().BuildInstance()

	if inst == nil {
		t.Fatal("BuildInstance() returned nil")
	}

	var _ *Instance = inst
}

// =============================================================================
// Convenience Functions
// =============================================================================

func TestOf(t *testing.T) {
	nodes := []TreeNode{
		{Content: "test", NodeID: 0},
	}

	vnode := Of(nodes)

	if vnode == nil {
		t.Fatal("Of() returned nil")
	}

	inst, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Of() did not return a *VNode")
	}

	if len(inst.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(inst.nodes))
	}
}

func TestOfLines(t *testing.T) {
	lines := []string{
		"root",
		"    child",
	}

	vnode := OfLines(lines)

	if vnode == nil {
		t.Fatal("OfLines() returned nil")
	}

	inst, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("OfLines() did not return a *VNode")
	}

	if len(inst.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(inst.nodes))
	}
}

func TestTreeView(t *testing.T) {
	builder := TreeView()

	if builder == nil {
		t.Fatal("TreeView() returned nil")
	}

	vnode := builder.Build()

	if vnode == nil {
		t.Fatal("TreeView().Build() returned nil")
	}
}

func TestWithNodes(t *testing.T) {
	nodes := []TreeNode{
		{Content: "test", NodeID: 0},
	}

	builder := WithNodes(nodes)

	if builder == nil {
		t.Fatal("WithNodes() returned nil")
	}

	vnode := builder.BuildVNode()

	if len(vnode.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(vnode.nodes))
	}
}

func TestWithLines(t *testing.T) {
	lines := []string{"root"}

	builder := WithLines(lines)

	if builder == nil {
		t.Fatal("WithLines() returned nil")
	}

	vnode := builder.BuildVNode()

	if len(vnode.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(vnode.nodes))
	}
}
