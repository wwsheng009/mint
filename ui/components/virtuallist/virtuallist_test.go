package virtuallist

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	vnode := New()
	if vnode == nil {
		t.Fatal("New() returned nil")
	}
	if vnode.Tag() != "virtuallist" {
		t.Errorf("Expected tag 'virtuallist', got '%s'", vnode.Tag())
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
	items := []string{"Item 1", "Item 2", "Item 3"}

	vnode := New().
		SetItems(items).
		SetItemCount(3).
		SetItemHeight(2).
		SetVisibleCount(10).
		SetHeight(15).
		SetWidth(50).
		SetScrollOffset(0).
		SetSelectedIndex(1).
		SetAllowScroll(true)

	if vnode.itemCount != 3 {
		t.Errorf("Expected itemCount 3, got %d", vnode.itemCount)
	}
	if vnode.itemHeight != 2 {
		t.Errorf("Expected itemHeight 2, got %d", vnode.itemHeight)
	}
	if vnode.visibleCount != 10 {
		t.Errorf("Expected visibleCount 10, got %d", vnode.visibleCount)
	}
	if vnode.height != 15 {
		t.Errorf("Expected height 15, got %d", vnode.height)
	}
	if vnode.width != 50 {
		t.Errorf("Expected width 50, got %d", vnode.width)
	}
	if vnode.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset 0, got %d", vnode.scrollOffset)
	}
	if vnode.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", vnode.selectedIndex)
	}
	if !vnode.allowScroll {
		t.Error("Expected allowScroll to be true")
	}
}

func TestVNode_AddItem(t *testing.T) {
	vnode := New()

	if len(vnode.items) != 0 {
		t.Errorf("Expected empty items initially, got %d", len(vnode.items))
	}

	vnode.AddItem("Item 1")
	if len(vnode.items) != 1 {
		t.Errorf("Expected 1 item after AddItem, got %d", len(vnode.items))
	}

	vnode.AddItem("Item 2")
	vnode.AddItem("Item 3")
	if len(vnode.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(vnode.items))
	}

	if vnode.items[0] != "Item 1" {
		t.Errorf("Expected 'Item 1', got '%s'", vnode.items[0])
	}
	if vnode.items[2] != "Item 3" {
		t.Errorf("Expected 'Item 3', got '%s'", vnode.items[2])
	}
}

func TestVNode_SetScrollOffset(t *testing.T) {
	vnode := New()

	vnode.SetScrollOffset(5)
	if vnode.scrollOffset != 5 {
		t.Errorf("Expected scrollOffset 5, got %d", vnode.scrollOffset)
	}

	vnode.SetScrollOffset(0)
	if vnode.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset 0, got %d", vnode.scrollOffset)
	}
}

func TestVNode_SetSelectedIndex(t *testing.T) {
	vnode := New()

	vnode.SetSelectedIndex(2)
	if vnode.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2, got %d", vnode.selectedIndex)
	}

	vnode.SetSelectedIndex(-1)
	if vnode.selectedIndex != -1 {
		t.Errorf("Expected selectedIndex -1, got %d", vnode.selectedIndex)
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

	// Test ComponentInstance interface
	var _ rtui.ComponentInstance = inst

	// Test PaintableInstance interface
	var _ rtui.PaintableInstance = inst

	// Test ActionHandlerInstance interface
	var _ rtui.ActionHandlerInstance = inst
}

func TestInstance_DefaultProps(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	if inst.scrollOffset != 0 {
		t.Errorf("Expected default scrollOffset 0, got %d", inst.scrollOffset)
	}
	if inst.selectedIndex != -1 {
		t.Errorf("Expected default selectedIndex -1, got %d", inst.selectedIndex)
	}
	if inst.itemHeight != 1 {
		t.Errorf("Expected default itemHeight 1, got %d", inst.itemHeight)
	}
	if inst.visibleCount != 10 {
		t.Errorf("Expected default visibleCount 10, got %d", inst.visibleCount)
	}
	if inst.height != 10 {
		t.Errorf("Expected default height 10, got %d", inst.height)
	}
	if inst.width != 40 {
		t.Errorf("Expected default width 40, got %d", inst.width)
	}
	if !inst.allowScroll {
		t.Error("Expected default allowScroll true")
	}
}

func TestInstance_Props(t *testing.T) {
	props := rtui.Props{
		"items":         []string{"A", "B", "C"},
		"itemCount":     3,
		"itemHeight":    2,
		"visibleCount":  5,
		"height":        12,
		"width":         40,
		"scrollOffset":  0,
		"selectedIndex": 1,
		"allowScroll":   false,
	}

	inst := NewInstance(props)

	if len(inst.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(inst.items))
	}
	if inst.itemCount != 3 {
		t.Errorf("Expected itemCount 3, got %d", inst.itemCount)
	}
	if inst.itemHeight != 2 {
		t.Errorf("Expected itemHeight 2, got %d", inst.itemHeight)
	}
	if inst.visibleCount != 5 {
		t.Errorf("Expected visibleCount 5, got %d", inst.visibleCount)
	}
	if inst.height != 12 {
		t.Errorf("Expected height 12, got %d", inst.height)
	}
	if inst.width != 40 {
		t.Errorf("Expected width 40, got %d", inst.width)
	}
	if inst.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", inst.selectedIndex)
	}
	if inst.allowScroll {
		t.Error("Expected allowScroll false")
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":  40,
		"height": 15,
	})

	constraints := layout.Constraints{
		MinWidth:  30,
		MaxWidth:  60,
		MinHeight: 10,
		MaxHeight: 20,
	}

	size := inst.Measure(constraints)
	if size.Width != 40 {
		t.Errorf("Expected width 40, got %d", size.Width)
	}
	if size.Height != 15 {
		t.Errorf("Expected height 15, got %d", size.Height)
	}
}

func TestInstance_MeasureWithConstraints(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width":  25,
		"height": 8,
	})

	constraints := layout.Constraints{
		MinWidth:  30,
		MaxWidth:  60,
		MinHeight: 10,
		MaxHeight: 20,
	}

	size := inst.Measure(constraints)
	if size.Width != 30 {
		t.Errorf("Expected constrained width 30, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("Expected constrained height 10, got %d", size.Height)
	}
}

func TestInstance_NavigateUp(t *testing.T) {
	props := rtui.Props{
		"items":         []string{"A", "B", "C", "D", "E"},
		"itemCount":     5,
		"selectedIndex": 3,
	}
	inst := NewInstance(props)

	result := inst.navigateUp()
	if !result {
		t.Error("navigateUp() should return true")
	}
	if inst.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2 after navigateUp, got %d", inst.selectedIndex)
	}
}

func TestInstance_NavigateUpAtStart(t *testing.T) {
	props := rtui.Props{
		"items":         []string{"A", "B", "C"},
		"itemCount":     3,
		"selectedIndex": 0,
	}
	inst := NewInstance(props)

	result := inst.navigateUp()
	if result {
		t.Error("navigateUp() should return false when already at start")
	}
	if inst.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", inst.selectedIndex)
	}
}

func TestInstance_NavigateDown(t *testing.T) {
	props := rtui.Props{
		"items":         []string{"A", "B", "C", "D", "E"},
		"itemCount":     5,
		"selectedIndex": 1,
	}
	inst := NewInstance(props)

	result := inst.navigateDown()
	if !result {
		t.Error("navigateDown() should return true")
	}
	if inst.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2 after navigateDown, got %d", inst.selectedIndex)
	}
}

func TestInstance_NavigateDownAtEnd(t *testing.T) {
	props := rtui.Props{
		"items":         []string{"A", "B", "C"},
		"itemCount":     3,
		"selectedIndex": 2,
	}
	inst := NewInstance(props)

	result := inst.navigateDown()
	if result {
		t.Error("navigateDown() should return false when already at end")
	}
	if inst.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2, got %d", inst.selectedIndex)
	}
}

func TestInstance_ScrollBy(t *testing.T) {
	props := rtui.Props{
		"items":        make([]string, 20),
		"itemCount":    20,
		"visibleCount": 10,
	}
	inst := NewInstance(props)

	inst.scrollBy(5)
	if inst.scrollOffset != 5 {
		t.Errorf("Expected scrollOffset 5, got %d", inst.scrollOffset)
	}

	inst.scrollBy(3)
	if inst.scrollOffset != 8 {
		t.Errorf("Expected scrollOffset 8, got %d", inst.scrollOffset)
	}
}

func TestInstance_ClampOffset(t *testing.T) {
	props := rtui.Props{
		"items":        make([]string, 5),
		"itemCount":    5,
		"visibleCount": 10,
	}
	inst := NewInstance(props)

	// Should clamp to 0 since visibleCount > itemCount
	inst.scrollTo(5)
	if inst.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset clamped to 0, got %d", inst.scrollOffset)
	}
}

func TestInstance_ClampSelectedIndex(t *testing.T) {
	props := rtui.Props{
		"items":     make([]string, 5),
		"itemCount": 5,
	}
	inst := NewInstance(props)

	// Set to valid index
	inst.selectedIndex = 3
	inst.clampSelectedIndex()
	if inst.selectedIndex != 3 {
		t.Errorf("Expected selectedIndex 3, got %d", inst.selectedIndex)
	}

	// Set to invalid - should clamp to max - 1
	inst.selectedIndex = 10
	inst.clampSelectedIndex()
	if inst.selectedIndex != 4 {
		t.Errorf("Expected selectedIndex clamped to 4, got %d", inst.selectedIndex)
	}

	// Set to invalid - should clamp to -1
	inst.selectedIndex = -5
	inst.clampSelectedIndex()
	if inst.selectedIndex != -1 {
		t.Errorf("Expected selectedIndex clamped to -1, got %d", inst.selectedIndex)
	}
}

func TestInstance_GetVisibleRange(t *testing.T) {
	props := rtui.Props{
		"items":        make([]string, 20),
		"itemCount":    20,
		"visibleCount": 10,
		"scrollOffset": 5,
	}
	inst := NewInstance(props)

	start, end := inst.GetVisibleRange()
	if start != 5 {
		t.Errorf("Expected start 5, got %d", start)
	}
	if end != 15 {
		t.Errorf("Expected end 15, got %d", end)
	}
}

func TestInstance_IsAtEnd(t *testing.T) {
	props := rtui.Props{
		"items":        make([]string, 10),
		"itemCount":    10,
		"visibleCount": 10,
	}
	inst := NewInstance(props)

	if !inst.IsItemAtEnd() {
		t.Error("Expected to be at end when itemCount equals visibleCount")
	}

	props2 := rtui.Props{
		"items":        make([]string, 20),
		"itemCount":    20,
		"visibleCount": 10,
		"scrollOffset": 10,
	}
	inst2 := NewInstance(props2)

	if !inst2.IsItemAtEnd() {
		t.Error("Expected to be at end when scrolled to bottom")
	}
}

func TestInstance_GetItem(t *testing.T) {
	props := rtui.Props{
		"items": []string{"A", "B", "C"},
	}
	inst := NewInstance(props)

	if inst.GetItem(0) != "A" {
		t.Errorf("Expected 'A', got '%s'", inst.GetItem(0))
	}
	if inst.GetItem(1) != "B" {
		t.Errorf("Expected 'B', got '%s'", inst.GetItem(1))
	}
	if inst.GetItem(2) != "C" {
		t.Errorf("Expected 'C', got '%s'", inst.GetItem(2))
	}
	if inst.GetItem(3) != "" {
		t.Errorf("Expected empty string for out-of-range index, got '%s'", inst.GetItem(3))
	}
	if inst.GetItem(-1) != "" {
		t.Errorf("Expected empty string for negative index, got '%s'", inst.GetItem(-1))
	}
}

func TestInstance_HandleAction(t *testing.T) {
	props := rtui.Props{
		"items":       []string{"A", "B", "C", "D", "E"},
		"itemCount":   5,
		"allowScroll": true,
	}
	inst := NewInstance(props)

	// Test navigate_up with item selected
	inst.selectedIndex = 2
	result := inst.HandleAction(action.NewAction(action.ActionNavigateUp))
	if !result {
		t.Error("HandleAction navigate_up should return true")
	}
	if inst.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", inst.selectedIndex)
	}

	// Test navigate_down
	inst.HandleAction(action.NewAction(action.ActionNavigateDown))
	if inst.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2, got %d", inst.selectedIndex)
	}

	// Test scroll action - virtuallist doesn't handle generic scroll action
	result = inst.HandleAction(action.NewActionWithPayload(action.ActionScroll, 5))
	if result {
		t.Error("HandleAction scroll should return false (not implemented)")
	}

	// Test scroll_up - requires canScrollUp to be true
	// First set up scroll offset to allow scrolling up
	inst.scrollOffset = 5
	result = inst.HandleAction(action.NewAction(action.ActionNavigatePageUp))
	if !result {
		t.Error("HandleAction page_up should return true when scrollOffset > 0")
	}

	// Test scroll_down - requires canScrollDown to be true
	// Set up to allow scrolling down
	inst.scrollOffset = 0
	inst.visibleCount = 3 // Can show 3 items
	inst.itemCount = 5    // Has 5 items, so can scroll down
	result = inst.HandleAction(action.NewAction(action.ActionNavigatePageDown))
	if !result {
		t.Error("HandleAction page_down should return true when can scroll down")
	}

	// Test navigate_home - requires scrollOffset > 0
	inst.scrollOffset = 5
	inst.visibleCount = 3
	inst.itemCount = 10
	result = inst.HandleAction(action.NewAction(action.ActionNavigateHome))
	if !result {
		t.Error("HandleAction navigate_home should return true when scrollOffset > 0")
	}
	if inst.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset 0 after navigate_home, got %d", inst.scrollOffset)
	}

	// Test navigate_end - requires not at end
	inst.scrollOffset = 0
	inst.itemCount = 10
	inst.visibleCount = 3
	result = inst.HandleAction(action.NewAction(action.ActionNavigateEnd))
	if !result {
		t.Error("HandleAction navigate_end should return true when not at end")
	}
}

func TestInstance_HandleAction_DisabledScroll(t *testing.T) {
	props := rtui.Props{
		"items":       []string{"A", "B", "C"},
		"itemCount":   3,
		"allowScroll": false,
	}
	inst := NewInstance(props)

	result := inst.HandleAction(action.NewAction(action.ActionNavigateUp))
	if result {
		t.Error("HandleAction should return false when allowScroll is false")
	}
}

func TestInstance_DefaultsItemCountToItemsLength(t *testing.T) {
	props := rtui.Props{
		"items":         []string{"A", "B", "C"},
		"visibleCount":  2,
		"height":        4,
		"width":         12,
		"selectedIndex": 1,
	}
	inst := NewInstance(props)

	if inst.itemCount != 3 {
		t.Fatalf("itemCount = %d, want 3", inst.itemCount)
	}
	if inst.selectedIndex != 1 {
		t.Fatalf("selectedIndex = %d, want 1", inst.selectedIndex)
	}
	start, end := inst.GetVisibleRange()
	if start != 0 || end != 2 {
		t.Fatalf("visible range = (%d, %d), want (0, 2)", start, end)
	}
}

func TestInstance_Paint_LongRowDoesNotPanic(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"items":        []string{"this is a very long virtual list row that previously overflowed"},
		"visibleCount": 1,
		"height":       3,
		"width":        12,
	})

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Paint should not panic for long rows, got %v", recovered)
		}
	}()

	cmds := inst.Paint(0, 0)
	if len(cmds) != 3 {
		t.Fatalf("command count = %d, want 3", len(cmds))
	}
	if cmds[1].Text != "│ this i.. │" {
		t.Fatalf("row text = %q, want %q", cmds[1].Text, "│ this i.. │")
	}
}

func TestInstance_HandleAction_UnknownAction(t *testing.T) {
	props := rtui.Props{
		"items":       []string{"A", "B"},
		"itemCount":   2,
		"allowScroll": true,
	}
	inst := NewInstance(props)

	result := inst.HandleAction(action.NewActionWithPayload("unknown_action", nil))
	if result {
		t.Error("HandleAction should return false for unknown action")
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
	items := []string{"Item 1", "Item 2", "Item 3"}

	vnode := NewBuilder().
		Key("test-list").
		Items(items).
		ItemCount(3).
		ItemHeight(2).
		VisibleCount(10).
		Height(15).
		Width(50).
		ScrollOffset(0).
		SelectedIndex(1).
		AllowScroll(true).
		BuildVNode()

	if vnode.key != "test-list" {
		t.Errorf("Expected key 'test-list', got '%s'", vnode.key)
	}
	if vnode.itemCount != 3 {
		t.Errorf("Expected itemCount 3, got %d", vnode.itemCount)
	}
	if vnode.itemHeight != 2 {
		t.Errorf("Expected itemHeight 2, got %d", vnode.itemHeight)
	}
	if vnode.visibleCount != 10 {
		t.Errorf("Expected visibleCount 10, got %d", vnode.visibleCount)
	}
	if vnode.height != 15 {
		t.Errorf("Expected height 15, got %d", vnode.height)
	}
	if vnode.width != 50 {
		t.Errorf("Expected width 50, got %d", vnode.width)
	}
	if vnode.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", vnode.selectedIndex)
	}
	if !vnode.allowScroll {
		t.Error("Expected allowScroll true")
	}
}

func TestBuilder_Size(t *testing.T) {
	vnode := NewBuilder().
		Size(40, 10).
		BuildVNode()

	if vnode.width != 40 {
		t.Errorf("Expected width 40, got %d", vnode.width)
	}
	if vnode.height != 10 {
		t.Errorf("Expected height 10, got %d", vnode.height)
	}
}

func TestBuilder_AddItem(t *testing.T) {
	vnode := NewBuilder().
		AddItem("A").
		AddItem("B").
		AddItem("C").
		BuildVNode()

	if len(vnode.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(vnode.items))
	}
	if vnode.items[0] != "A" {
		t.Errorf("Expected 'A', got '%s'", vnode.items[0])
	}
}

func TestBuilder_ListStyle(t *testing.T) {
	listStyle := style.Style{FG: style.Red, BG: style.Blue}
	selectedStyle := style.Style{FG: style.Green, BG: style.Yellow}

	vnode := NewBuilder().
		ListStyle(listStyle).
		SelectedStyle(selectedStyle).
		BuildVNode()

	if vnode.listStyle.FG != style.Red {
		t.Errorf("Expected listStyle.FG red, got %s", vnode.listStyle.FG)
	}
	if vnode.selectedStyle.FG != style.Green {
		t.Errorf("Expected selectedStyle.FG green, got %s", vnode.selectedStyle.FG)
	}
}

func TestBuilder_ColorMethods(t *testing.T) {
	vnode := NewBuilder().
		FgColor(style.Blue).
		BgColor(style.Cyan).
		SelectedFgColor(style.Green).
		SelectedBgColor(style.Yellow).
		BuildVNode()

	if vnode.listStyle.FG != style.Blue {
		t.Errorf("Expected listStyle.FG blue, got %s", vnode.listStyle.FG)
	}
	if vnode.listStyle.BG != style.Cyan {
		t.Errorf("Expected listStyle.BG cyan, got %s", vnode.listStyle.BG)
	}
	if vnode.selectedStyle.FG != style.Green {
		t.Errorf("Expected selectedStyle.FG green, got %s", vnode.selectedStyle.FG)
	}
	if vnode.selectedStyle.BG != style.Yellow {
		t.Errorf("Expected selectedStyle.BG yellow, got %s", vnode.selectedStyle.BG)
	}
}

func TestBuilder_Build(t *testing.T) {
	items := []string{"A", "B"}

	vnode := NewBuilder().Items(items).Build()

	if vnode == nil {
		t.Fatal("Build() returned nil")
	}

	// Test that it returns a VNode
	var _ rtui.VNode = vnode
}

func TestBuilder_BuildInstance(t *testing.T) {
	items := []string{"A", "B"}

	inst := NewBuilder().Items(items).BuildInstance()

	if inst == nil {
		t.Fatal("BuildInstance() returned nil")
	}

	// Test that it returns an Instance
	var _ *Instance = inst
}

func TestBuilder_BuildVNode(t *testing.T) {
	items := []string{"A", "B"}

	vnode := NewBuilder().Items(items).BuildVNode()

	if vnode == nil {
		t.Fatal("BuildVNode() returned nil")
	}

	// Test that it returns a *VNode
	var _ *VNode = vnode
}

// =============================================================================
// Convenience Functions
// =============================================================================

func TestOf(t *testing.T) {
	items := []string{"A", "B", "C"}

	vnode := Of(items)

	if vnode == nil {
		t.Fatal("Of() returned nil")
	}

	inst, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Of() did not return a *VNode")
	}

	if len(inst.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(inst.items))
	}
}

func TestOfSize(t *testing.T) {
	items := []string{"A", "B", "C"}

	vnode := OfSize(items, 40, 15).(*VNode)

	if vnode == nil {
		t.Fatal("OfSize() returned nil")
	}

	if vnode.width != 40 {
		t.Errorf("Expected width 40, got %d", vnode.width)
	}
	if vnode.height != 15 {
		t.Errorf("Expected height 15, got %d", vnode.height)
	}
}

func TestOfItems(t *testing.T) {
	items := []string{"A", "B", "C", "D", "E"}

	vnode := OfItems(items, 10, 5).(*VNode)

	if vnode == nil {
		t.Fatal("OfItems() returned nil")
	}

	if vnode.itemCount != 10 {
		t.Errorf("Expected itemCount 10, got %d", vnode.itemCount)
	}
	if vnode.visibleCount != 5 {
		t.Errorf("Expected visibleCount 5, got %d", vnode.visibleCount)
	}
}

func TestVirtualList(t *testing.T) {
	builder := VirtualList()

	if builder == nil {
		t.Fatal("VirtualList() returned nil")
	}

	vnode := builder.Items([]string{"A", "B"}).Build()

	if vnode == nil {
		t.Fatal("VirtualList().Build() returned nil")
	}
}

func TestWithItems(t *testing.T) {
	items := []string{"A", "B", "C"}

	builder := WithItems(items)

	if builder == nil {
		t.Fatal("WithItems() returned nil")
	}

	vnode := builder.BuildVNode()

	if len(vnode.items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(vnode.items))
	}
}
