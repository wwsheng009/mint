package treeview

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type testTreeReorderIntent struct{}

func (testTreeReorderIntent) IntentType() string { return "treeview.test.reorder" }

func withTreeIntentRuntime(t *testing.T) *intent.Runtime {
	t.Helper()

	oldRuntime := rtui.GetGlobalIntentRuntime()
	rt := intent.NewRuntimeWithNewRegistry()
	rtui.SetGlobalIntentRuntime(rt)
	t.Cleanup(func() {
		rtui.SetGlobalIntentRuntime(oldRuntime)
	})
	return rt
}

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

func TestBuilder_ReorderableAndOnReorder(t *testing.T) {
	vnode := NewBuilder().
		Reorderable(true).
		OnReorder(testTreeReorderIntent{}).
		BuildVNode()

	if !vnode.reorderable {
		t.Fatal("Expected vnode to be reorderable")
	}
	if _, ok := vnode.reorderIntent.(testTreeReorderIntent); !ok {
		t.Fatalf("Expected vnode reorder intent to be testTreeReorderIntent, got %T", vnode.reorderIntent)
	}

	inst := vnode.CreateInstance().(*Instance)
	if !inst.reorderable {
		t.Fatal("Expected instance to be reorderable")
	}
	if _, ok := inst.reorderIntent.(testTreeReorderIntent); !ok {
		t.Fatalf("Expected instance reorder intent to be testTreeReorderIntent, got %T", inst.reorderIntent)
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
	var _ rtui.FocusableInstance = inst
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
		"expandLevel":    2,
		"showIcons":      false,
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
		MinWidth:  40,
		MaxWidth:  80,
		MinHeight: 10,
		MaxHeight: 30,
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

func TestInstance_SearchQueryFilters(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "child", NodeID: 1, NodeType: "file"},
			{Indent: 4, Content: "other", NodeID: 2, NodeType: "file"},
		},
		"searchQuery": "child",
	}
	inst := NewInstance(props)

	visible, _ := inst.visibleEntries()
	if len(visible) != 2 {
		t.Fatalf("expected 2 visible nodes (root + child), got %d", len(visible))
	}
	if !visible[1].Match {
		t.Fatalf("expected child to be marked as match")
	}
}

func TestInstance_ToggleChecked(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "child", NodeID: 1, NodeType: "file"},
		},
		"selectionMode": SelectionMultiple,
	}
	inst := NewInstance(props)
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		t.Fatal("expected visible entries")
	}

	inst.toggleChecked(visible[1])
	if len(inst.checkedKeys) != 1 {
		t.Fatalf("expected 1 checked key, got %d", len(inst.checkedKeys))
	}

	inst.toggleChecked(visible[1])
	if len(inst.checkedKeys) != 0 {
		t.Fatalf("expected 0 checked keys after toggle off, got %d", len(inst.checkedKeys))
	}
}

func TestInstance_ToggleCheckedCascadesDescendants(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", Path: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "child", Path: "root/child", NodeID: 1, NodeType: "file"},
			{Indent: 4, Content: "branch", Path: "root/branch", NodeID: 2, NodeType: "folder"},
			{Indent: 8, Content: "leaf", Path: "root/branch/leaf", NodeID: 3, NodeType: "file"},
		},
		"selectionMode": SelectionMultiple,
		"expandLevel":   3,
	}
	inst := NewInstance(props)
	visible, _ := inst.visibleEntries()

	if !inst.toggleChecked(visible[0]) {
		t.Fatalf("expected cascade select on parent to succeed")
	}
	if got := len(inst.GetCheckedKeys()); got != 4 {
		t.Fatalf("expected all descendants to be selected, got %d", got)
	}

	if !inst.toggleChecked(visible[0]) {
		t.Fatalf("expected cascade deselect on parent to succeed")
	}
	if got := len(inst.GetCheckedKeys()); got != 0 {
		t.Fatalf("expected all descendants to be deselected, got %d", got)
	}
}

func TestInstance_LazyLoadRequested(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 0, NodeType: "folder", Lazy: true},
		},
		"selectedIndex": 0,
		"expandLevel":   0,
	}
	inst := NewInstance(props)
	inst.toggleExpand()

	key := nodeKey(inst.nodes[0], 0)
	if !inst.lazyRequested[key] {
		t.Fatalf("expected lazy load request for key %s", key)
	}
}

func TestInstance_LazyLoadChildren(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 0, NodeType: "folder", Lazy: true},
		},
		"selectedIndex":      0,
		"expandLevel":        0,
		"lazyLoadChildrenFn": func(node TreeNode) []TreeNode { return []TreeNode{{Content: "child", NodeID: 1}} },
	}
	inst := NewInstance(props)
	inst.toggleExpand()

	if len(inst.nodes) != 2 {
		t.Fatalf("expected 2 nodes after lazy load, got %d", len(inst.nodes))
	}
	if inst.nodes[1].Indent <= inst.nodes[0].Indent {
		t.Fatalf("expected child indent greater than parent (parent=%d child=%d)", inst.nodes[0].Indent, inst.nodes[1].Indent)
	}
}

func TestInstance_LazyLoadNotRepeatedAfterCollapse(t *testing.T) {
	loadCount := 0
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 0, NodeType: "folder", Lazy: true},
		},
		"selectedIndex": 0,
		"expandLevel":   0,
		"lazyLoadFn": func(node TreeNode) {
			loadCount++
		},
	}
	inst := NewInstance(props)

	if !inst.toggleExpand() {
		t.Fatalf("expected first expand to succeed")
	}
	if loadCount != 1 {
		t.Fatalf("expected first expand to request lazy load once, got %d", loadCount)
	}
	if inst.toggleExpand() {
		t.Fatalf("expected collapse to return false")
	}
	if !inst.toggleExpand() {
		t.Fatalf("expected re-expand to succeed")
	}
	if loadCount != 1 {
		t.Fatalf("expected lazy load request to be deduplicated after collapse, got %d", loadCount)
	}
}

func TestInstance_HandleIntentExpandTriggersLazyLoad(t *testing.T) {
	loadCount := 0
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 7, Path: "root/lazy", NodeType: "folder", Lazy: true},
		},
		"expandLevel": 0,
		"lazyLoadFn": func(node TreeNode) {
			loadCount++
		},
	}
	inst := NewInstance(props)

	if !inst.HandleIntent(NodeExpand(0, "root/lazy", 7)) {
		t.Fatalf("expected expand intent to be handled")
	}
	if loadCount != 1 {
		t.Fatalf("expected expand intent to trigger lazy load, got %d", loadCount)
	}
	key := nodeKey(inst.nodes[0], 0)
	if !inst.expandState[key] {
		t.Fatalf("expected node to be expanded after intent")
	}
}

func TestInstance_HandleIntentLazyLoadSuccess(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 7, Path: "root/lazy", NodeType: "folder", Lazy: true, Loading: true},
		},
		"expandLevel": 0,
	}
	inst := NewInstance(props)

	if !inst.HandleIntent(LazyLoadSuccess(0, "root/lazy", 7, []TreeNode{{Content: "child", NodeID: 8}})) {
		t.Fatalf("expected lazy load success intent to be handled")
	}
	if len(inst.nodes) != 2 {
		t.Fatalf("expected child to be inserted, got %d nodes", len(inst.nodes))
	}
	if inst.nodes[0].Loading || inst.nodes[0].Lazy {
		t.Fatalf("expected parent loading/lazy flags to be cleared")
	}
	if inst.nodes[1].Path != "root/lazy/child" {
		t.Fatalf("expected child path to be normalized, got %q", inst.nodes[1].Path)
	}
}

func TestInstance_HandleIntentLazyLoadSuccessReplace(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 7, Path: "root/lazy", NodeType: "folder", Lazy: true, Loading: true},
			{Indent: 4, Content: "old", NodeID: 8, Path: "root/lazy/old", NodeType: "file"},
		},
		"expandLevel": 2,
	}
	inst := NewInstance(props)

	intent := LazyLoadSuccess(0, "root/lazy", 7, []TreeNode{{Content: "fresh", NodeID: 9}})
	intent.Replace = true
	if !inst.HandleIntent(intent) {
		t.Fatalf("expected lazy load replace intent to be handled")
	}
	if len(inst.nodes) != 2 {
		t.Fatalf("expected descendants to be replaced, got %d nodes", len(inst.nodes))
	}
	if inst.nodes[1].Path != "root/lazy/fresh" {
		t.Fatalf("expected replacement child path, got %q", inst.nodes[1].Path)
	}
}

func TestInstance_HandleIntentLazyLoadFailure(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 7, Path: "root/lazy", NodeType: "folder", Lazy: true, Loading: true},
		},
		"expandLevel": 0,
	}
	inst := NewInstance(props)

	if !inst.HandleIntent(LazyLoadFailure(0, "root/lazy", 7, "network error")) {
		t.Fatalf("expected lazy load failure intent to be handled")
	}
	if inst.nodes[0].Loading {
		t.Fatalf("expected loading flag to be cleared after failure")
	}
	if inst.nodes[0].LoadError != "network error" {
		t.Fatalf("expected load error to be recorded, got %q", inst.nodes[0].LoadError)
	}
}

func TestInstance_ControlledExpandedPathsReplaceLocalState(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", Path: "root", NodeID: 1, NodeType: "folder"},
			{Indent: 4, Content: "child", Path: "root/child", NodeID: 2, NodeType: "file"},
		},
		"expandLevel": 0,
		"expandedKeys": map[string]bool{
			"root": true,
		},
		"expandedKeysControlled": true,
	}
	inst := NewInstance(props)

	if visible := inst.GetVisibleNodes(); len(visible) != 2 {
		t.Fatalf("expected controlled expanded tree to show child, got %d visible nodes", len(visible))
	}

	updated := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", Path: "root", NodeID: 1, NodeType: "folder"},
			{Indent: 4, Content: "child", Path: "root/child", NodeID: 2, NodeType: "file"},
		},
		"expandLevel":            0,
		"expandedKeys":           map[string]bool{},
		"expandedKeysControlled": true,
	}
	inst.SetProps(updated)

	if visible := inst.GetVisibleNodes(); len(visible) != 1 {
		t.Fatalf("expected controlled collapse to hide child, got %d visible nodes", len(visible))
	}
}

func TestInstance_SearchNextPrev(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "alpha", NodeID: 1, NodeType: "file"},
			{Indent: 4, Content: "beta", NodeID: 2, NodeType: "file"},
		},
		"searchQuery":   "beta",
		"selectedIndex": 0,
	}
	inst := NewInstance(props)

	actNext := action.NewAction(action.ActionSearch).WithPayload("next")
	if !inst.HandleAction(actNext) {
		t.Fatalf("expected search next to be handled")
	}
	if inst.selectedIndex < 0 {
		t.Fatalf("expected selected index to move to match")
	}

	actPrev := action.NewAction(action.ActionSearch).WithPayload("prev")
	if !inst.HandleAction(actPrev) {
		t.Fatalf("expected search prev to be handled")
	}
	if inst.selectedIndex < 0 {
		t.Fatalf("expected selected index to remain valid")
	}
}

func TestInstance_PaintSearchStats(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "child", NodeID: 1, NodeType: "file"},
		},
		"searchQuery":     "child",
		"showSearchStats": true,
	}
	inst := NewInstance(props)

	cmds := inst.Paint(0, 0)
	foundStats := false
	foundRoot := false
	for _, cmd := range cmds {
		if cmd.Y == 1 && strings.Contains(cmd.Text, `Search: "child" 1/1`) {
			foundStats = true
		}
		if cmd.Y == 2 && strings.Contains(cmd.Text, "root") {
			foundRoot = true
		}
	}
	if !foundStats {
		t.Fatalf("expected search stats row to be painted")
	}
	if !foundRoot {
		t.Fatalf("expected content rows to start after search stats row")
	}
}

func TestInstance_PaintLazyLoadHint(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 0, NodeType: "folder", Lazy: true},
		},
		"expandLevel": 0,
	}
	inst := NewInstance(props)

	cmds := inst.Paint(0, 0)
	for _, cmd := range cmds {
		if strings.Contains(cmd.Text, "[load:R]") {
			return
		}
	}
	t.Fatalf("expected lazy load shortcut hint to be painted")
}

func TestInstance_FocusedBorderUsesHighlightStyle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, NodeType: "folder"},
		},
	})
	inst.SetFocus(true)

	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatalf("expected paint commands")
	}
	if cmds[0].Style.BG != "" {
		t.Fatalf("expected focused border to preserve empty background, got %q", cmds[0].Style.BG)
	}
	if !strings.HasPrefix(cmds[0].Text, "╔") || !strings.HasSuffix(cmds[0].Text, "╗") || !strings.Contains(cmds[0].Text, "[FOCUS]") {
		t.Fatalf("expected focused border to use double-line focus title, got %q", cmds[0].Text)
	}
	if !cmds[0].Style.IsBold() {
		t.Fatalf("expected focused border to be bold")
	}
}

func TestInstance_InputShortcutRefreshLazy(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "lazy", NodeID: 0, NodeType: "folder", Lazy: true},
		},
		"selectedIndex":      0,
		"expandLevel":        0,
		"lazyLoadChildrenFn": func(node TreeNode) []TreeNode { return []TreeNode{{Content: "child", NodeID: 1}} },
	}
	inst := NewInstance(props)

	if !inst.HandleAction(action.NewAction(action.ActionInputText).WithPayload("r")) {
		t.Fatalf("expected input shortcut to refresh selected lazy node")
	}
	if len(inst.nodes) != 2 {
		t.Fatalf("expected lazy children to be inserted via shortcut, got %d nodes", len(inst.nodes))
	}
}

func TestInstance_SelectRangeUsesAnchor(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, Path: "root", NodeType: "folder"},
			{Indent: 4, Content: "a", NodeID: 1, Path: "root/a", NodeType: "file"},
			{Indent: 4, Content: "b", NodeID: 2, Path: "root/b", NodeType: "file"},
			{Indent: 4, Content: "c", NodeID: 3, Path: "root/c", NodeType: "file"},
		},
		"selectionMode": SelectionMultiple,
		"expandLevel":   2,
	}
	inst := NewInstance(props)

	visible, _ := inst.visibleEntries()
	inst.selectVisibleIndex(1, visible, true)
	if !inst.HandleAction(action.NewAction(action.ActionToggleSelect)) {
		t.Fatalf("expected initial toggle select to succeed")
	}
	visible, _ = inst.visibleEntries()
	inst.selectVisibleIndex(3, visible, true)
	if !inst.HandleAction(action.NewAction(action.ActionSelectRange)) {
		t.Fatalf("expected range select with anchor to succeed")
	}

	keys := inst.GetCheckedKeys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 checked nodes from anchor range, got %d", len(keys))
	}
}

func TestInstance_ToggleSelectInvertMatches(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", NodeID: 0, Path: "root", NodeType: "folder"},
			{Indent: 4, Content: "alpha", NodeID: 1, Path: "root/alpha", NodeType: "file"},
			{Indent: 4, Content: "beta", NodeID: 2, Path: "root/beta", NodeType: "file"},
		},
		"selectionMode": SelectionMultiple,
		"searchQuery":   "beta",
		"expandLevel":   2,
	}
	inst := NewInstance(props)

	if !inst.HandleAction(action.NewAction(action.ActionToggleSelect).WithPayload("invert")) {
		t.Fatalf("expected invert toggle to succeed")
	}
	keys := inst.GetCheckedKeys()
	if len(keys) != 1 || keys[0] != "path:root/beta" {
		t.Fatalf("expected only matched node to be inverted, got %v", keys)
	}
}

func TestInstance_ActionRefreshTargetsNodeByPath(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", Path: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "lazy", Path: "root/lazy", NodeID: 1, NodeType: "folder", Lazy: true},
		},
		"selectedIndex": 0,
		"expandLevel":   2,
		"lazyLoadChildrenFn": func(node TreeNode) []TreeNode {
			if node.Path == "root/lazy" {
				return []TreeNode{{Content: "child", NodeID: 2}}
			}
			return nil
		},
	}
	inst := NewInstance(props)

	act := action.NewAction(action.ActionRefresh).WithPayload(map[string]string{"path": "root/lazy"})
	if !inst.HandleAction(act) {
		t.Fatalf("expected refresh action with path payload to be handled")
	}
	if len(inst.nodes) != 3 {
		t.Fatalf("expected targeted refresh to insert child, got %d nodes", len(inst.nodes))
	}
	if inst.nodes[2].Path != "root/lazy/child" {
		t.Fatalf("expected inserted child path to be normalized, got %q", inst.nodes[2].Path)
	}
}

func TestInstance_ActionSelectAllVisibleScope(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", Path: "root", NodeID: 0, NodeType: "folder"},
			{Indent: 4, Content: "alpha", Path: "root/alpha", NodeID: 1, NodeType: "file"},
			{Indent: 4, Content: "beta", Path: "root/beta", NodeID: 2, NodeType: "file"},
		},
		"selectionMode": SelectionMultiple,
		"searchQuery":   "beta",
		"expandLevel":   2,
	}
	inst := NewInstance(props)

	if !inst.HandleAction(action.NewAction(action.ActionSelectAll).WithPayload("visible")) {
		t.Fatalf("expected visible-scope select all to succeed")
	}
	keys := inst.GetCheckedKeys()
	if len(keys) != 2 {
		t.Fatalf("expected visible scope to select ancestor and match, got %v", keys)
	}
}

func TestInstance_DragReorder_RootSiblings(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "alpha", Path: "alpha", NodeType: "folder"},
			{Indent: 4, Content: "alpha-child", Path: "alpha/child", NodeType: "file"},
			{Content: "beta", Path: "beta", NodeType: "folder"},
			{Indent: 4, Content: "beta-child", Path: "beta/child", NodeType: "file"},
			{Content: "gamma", Path: "gamma", NodeType: "folder"},
		},
		"expandLevel":   2,
		"reorderable":   true,
		"selectedIndex": 0,
	}
	inst := NewInstance(props)

	start := runtimemsg.NewMouseMsgWithTarget(0, 1, 0, 1, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	move := runtimemsg.NewMouseMsgWithTarget(0, 5, 0, 5, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionMove)
	release := runtimemsg.NewMouseMsgWithTarget(0, 5, 0, 5, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)

	if !inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(start)) {
		t.Fatal("expected ActionClick press to start drag")
	}
	if !inst.HandleAction(action.NewAction(action.ActionHover).WithPayload(move)) {
		t.Fatal("expected ActionHover to reorder dragged node")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseRelease).WithPayload(release)) {
		t.Fatal("expected ActionMouseRelease to finish drag")
	}

	wantOrder := []string{"beta", "beta/child", "gamma", "alpha", "alpha/child"}
	for i, want := range wantOrder {
		if inst.nodes[i].Path != want {
			t.Fatalf("nodes[%d].Path = %q, want %q", i, inst.nodes[i].Path, want)
		}
	}
	selected, ok := inst.GetSelectedNode()
	if !ok || selected.Path != "alpha" {
		t.Fatalf("selected node = (%+v,%v), want alpha,true", selected, ok)
	}
}

func TestInstance_DragReorder_ChildSiblings(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", Path: "root", NodeType: "folder"},
			{Indent: 4, Content: "alpha", Path: "root/alpha", NodeType: "folder"},
			{Indent: 8, Content: "alpha-child", Path: "root/alpha/child", NodeType: "file"},
			{Indent: 4, Content: "beta", Path: "root/beta", NodeType: "folder"},
			{Indent: 8, Content: "beta-child", Path: "root/beta/child", NodeType: "file"},
			{Indent: 4, Content: "gamma", Path: "root/gamma", NodeType: "folder"},
		},
		"expandLevel": 3,
		"reorderable": true,
	}
	inst := NewInstance(props)

	start := runtimemsg.NewMouseMsgWithTarget(0, 2, 0, 2, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	move := runtimemsg.NewMouseMsgWithTarget(0, 6, 0, 6, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionMove)
	release := runtimemsg.NewMouseMsgWithTarget(0, 6, 0, 6, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)

	inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(start))
	inst.HandleAction(action.NewAction(action.ActionHover).WithPayload(move))
	inst.HandleAction(action.NewAction(action.ActionMouseRelease).WithPayload(release))

	wantOrder := []string{"root", "root/beta", "root/beta/child", "root/gamma", "root/alpha", "root/alpha/child"}
	for i, want := range wantOrder {
		if inst.nodes[i].Path != want {
			t.Fatalf("nodes[%d].Path = %q, want %q", i, inst.nodes[i].Path, want)
		}
	}
}

func TestInstance_Reorderable_ExpanderClickStillExpands(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "root", Path: "root", NodeType: "folder"},
			{Indent: 4, Content: "child", Path: "root/child", NodeType: "file"},
		},
		"expandLevel": 0,
		"reorderable": true,
	}
	inst := NewInstance(props)

	mouse := runtimemsg.NewMouseMsgWithTarget(0, 1, 2, 1, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	if !inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(mouse)) {
		t.Fatal("expected expander click to be handled")
	}
	if !inst.isExpanded(0) {
		t.Fatal("expected root to expand on expander click even when reorderable")
	}
}

func TestInstance_DragReorder_EmitsIntents(t *testing.T) {
	props := rtui.Props{
		"nodes": []TreeNode{
			{Content: "alpha", Path: "alpha", NodeType: "folder"},
			{Content: "beta", Path: "beta", NodeType: "folder"},
			{Content: "gamma", Path: "gamma", NodeType: "folder"},
		},
		"componentID":   "tree.orders",
		"reorderable":   true,
		"reorderIntent": testTreeReorderIntent{},
	}
	inst := NewInstance(props)

	rt := withTreeIntentRuntime(t)
	var emittedGlobal []NodeReorderIntent
	unregister := intent.RegisterTypedRuntime(rt, func(_ *intent.ActionContext, i NodeReorderIntent) intent.IntentResult {
		emittedGlobal = append(emittedGlobal, i)
		return intent.HandledResult()
	})
	defer unregister()

	var emittedLocal []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emittedLocal = append(emittedLocal, i)
	})

	start := runtimemsg.NewMouseMsgWithTarget(0, 1, 0, 1, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	move := runtimemsg.NewMouseMsgWithTarget(0, 3, 0, 3, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionMove)
	release := runtimemsg.NewMouseMsgWithTarget(0, 3, 0, 3, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)

	inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(start))
	inst.HandleAction(action.NewAction(action.ActionHover).WithPayload(move))
	inst.HandleAction(action.NewAction(action.ActionMouseRelease).WithPayload(release))

	if len(emittedLocal) != 1 {
		t.Fatalf("local emitted len = %d, want 1", len(emittedLocal))
	}
	if _, ok := emittedLocal[0].(testTreeReorderIntent); !ok {
		t.Fatalf("local emitted intent = %T, want testTreeReorderIntent", emittedLocal[0])
	}
	if len(emittedGlobal) != 1 {
		t.Fatalf("global emitted len = %d, want 1", len(emittedGlobal))
	}
	if emittedGlobal[0].ComponentID != "tree.orders" || emittedGlobal[0].Path != "alpha" || emittedGlobal[0].ToVisibleIndex != 2 {
		t.Fatalf("unexpected reorder intent payload: %+v", emittedGlobal[0])
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
		ShowSearchStats(true).
		SearchStatsStyle(style.Style{FG: style.Green}).
		SearchQueryControlled("child").
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
	if !vnode.showSearchStats {
		t.Error("Expected showSearchStats true")
	}
	if vnode.searchStatsStyle.FG != style.Green {
		t.Errorf("Expected searchStatsStyle.FG green, got %s", vnode.searchStatsStyle.FG)
	}
	if vnode.searchQuery != "child" || !vnode.searchQueryControlled {
		t.Errorf("Expected controlled search query to be set, got query=%q controlled=%t", vnode.searchQuery, vnode.searchQueryControlled)
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
