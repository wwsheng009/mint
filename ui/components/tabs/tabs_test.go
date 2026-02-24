package tabs

import (
	"testing"

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
	if vnode.Tag() != "tabs" {
		t.Errorf("Expected tag 'tabs', got '%s'", vnode.Tag())
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
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	vnode := New().
		SetTabs(tabs).
		SetPosition(TabPositionTop).
		SetWrapTabs(true).
		SetTabGap(2).
		SetWidth(80).
		SetHeight(20).
		SetFlex(1)

	if len(vnode.tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(vnode.tabs))
	}
	if vnode.position != TabPositionTop {
		t.Error("Expected TabPositionTop")
	}
	if !vnode.wrapTabs {
		t.Error("Expected wrapTabs to be true")
	}
	if vnode.tabGap != 2 {
		t.Errorf("Expected tabGap 2, got %d", vnode.tabGap)
	}
	if vnode.width != 80 {
		t.Errorf("Expected width 80, got %d", vnode.width)
	}
	if vnode.height != 20 {
		t.Errorf("Expected height 20, got %d", vnode.height)
	}
	if vnode.flex != 1 {
		t.Errorf("Expected flex 1, got %d", vnode.flex)
	}
}

func TestVNode_AddTab(t *testing.T) {
	vnode := New().
		AddTab("tab1", "Tab 1").
		AddTab("tab2", "Tab 2")

	if len(vnode.tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(vnode.tabs))
	}
	if vnode.tabs[0].ID != "tab1" {
		t.Errorf("Expected ID 'tab1', got '%s'", vnode.tabs[0].ID)
	}
	if vnode.tabs[1].Label != "Tab 2" {
		t.Errorf("Expected label 'Tab 2', got '%s'", vnode.tabs[1].Label)
	}
}

func TestVNode_AddTabWithOptions(t *testing.T) {
	vnode := New().
		AddTabWithOptions("tab1", "Tab 1", false).
		AddTabWithOptions("tab2", "Tab 2", true)

	if len(vnode.tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(vnode.tabs))
	}
	if vnode.tabs[0].Disabled {
		t.Error("Tab 1 should not be disabled")
	}
	if !vnode.tabs[1].Disabled {
		t.Error("Tab 2 should be disabled")
	}
}

func TestVNode_TabPosition(t *testing.T) {
	tests := []struct {
		name     string
		position TabPosition
	}{
		{"Top", TabPositionTop},
		{"Bottom", TabPositionBottom},
		{"Left", TabPositionLeft},
		{"Right", TabPositionRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := New().SetPosition(tt.position)
			if vnode.Position() != tt.position {
				t.Errorf("Expected position %v, got %v", tt.position, vnode.Position())
			}
		})
	}
}

func TestVNode_Props(t *testing.T) {
	tabs := []TabItem{{ID: "tab1", Label: "Tab 1"}}
	tabStyle := style.Style{FG: style.Color("blue")}

	vnode := New().
		SetTabs(tabs).
		SetPosition(TabPositionBottom).
		SetWrapTabs(true).
		SetTabStyle(tabStyle)

	props := vnode.Props()
	if !tabsEqual(props["tabs"].([]TabItem), tabs) {
		t.Error("Tabs mismatch")
	}
	if props["position"] != TabPositionBottom {
		t.Error("Position mismatch")
	}
	if props["wrapTabs"] != true {
		t.Error("WrapTabs mismatch")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_FluentAPI(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	vnode := NewBuilder().
		Tabs(tabs).
		Position(TabPositionTop).
		WrapTabs(false).
		TabGap(1).
		Width(100).
		Height(30).
		Flex(2).
		BuildVNode()

	if len(vnode.tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(vnode.tabs))
	}
	if vnode.width != 100 {
		t.Errorf("Expected width 100, got %d", vnode.width)
	}
	if vnode.height != 30 {
		t.Errorf("Expected height 30, got %d", vnode.height)
	}
}

func TestBuilder_AddTab(t *testing.T) {
	vnode := NewBuilder().
		AddTab("home", "Home").
		AddTab("about", "About").
		AddTab("contact", "Contact").
		BuildVNode()

	if len(vnode.tabs) != 3 {
		t.Errorf("Expected 3 tabs, got %d", len(vnode.tabs))
	}
}

func TestBuilder_Size(t *testing.T) {
	vnode := NewBuilder().
		Size(80, 25).
		BuildVNode()

	if vnode.width != 80 {
		t.Errorf("Expected width 80, got %d", vnode.width)
	}
	if vnode.height != 25 {
		t.Errorf("Expected height 25, got %d", vnode.height)
	}
}

func TestBuilder_PositionConvenience(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*Builder) *Builder
		expected TabPosition
	}{
		{"Top", func(b *Builder) *Builder { return b.Top() }, TabPositionTop},
		{"Bottom", func(b *Builder) *Builder { return b.Bottom() }, TabPositionBottom},
		{"Left", func(b *Builder) *Builder { return b.Left() }, TabPositionLeft},
		{"Right", func(b *Builder) *Builder { return b.Right() }, TabPositionRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.method(NewBuilder()).BuildVNode()
			if vnode.Position() != tt.expected {
				t.Errorf("Expected position %v, got %v", tt.expected, vnode.Position())
			}
		})
	}
}

func TestBuilder_ConvenienceFunctions(t *testing.T) {
	tabs := []TabItem{{ID: "t1", Label: "Test"}}

	// Test Of
	vnode := Of(tabs)
	if vnode.(*VNode).tabs == nil {
		t.Error("Of() should set tabs")
	}

	// Test OfWidth
	vnode = OfWidth(tabs, 60)
	if vnode.(*VNode).width != 60 {
		t.Error("OfWidth() should set width")
	}

	// Test OfSize
	vnode = OfSize(tabs, 80, 20)
	vn := vnode.(*VNode)
	if vn.width != 80 || vn.height != 20 {
		t.Error("OfSize() should set dimensions")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_NewInstance(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	props := rtui.Props{
		"tabs":     tabs,
		"position": TabPositionTop,
		"wrapTabs": false,
		"tabGap":   1,
	}

	inst := NewInstance(props)
	if inst == nil {
		t.Fatal("NewInstance() returned nil")
	}
	if len(inst.tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(inst.tabs))
	}
	if inst.position != TabPositionTop {
		t.Error("Expected TabPositionTop")
	}
	if inst.activeTab != 0 {
		t.Errorf("Expected activeTab 0, got %d", inst.activeTab)
	}
}

func TestInstance_SetActiveTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
		{ID: "tab3", Label: "Tab 3"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})

	// Set to valid index
	if !inst.SetActiveTab(1) {
		t.Error("SetActiveTab(1) should return true")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}

	// Set to invalid index
	if inst.SetActiveTab(5) {
		t.Error("SetActiveTab(5) should return false for invalid index")
	}
}

func TestInstance_SetActiveTabByID(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})

	if !inst.SetActiveTabByID("tab2") {
		t.Error("SetActiveTabByID('tab2') should return true")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}

	if inst.SetActiveTabByID("tab3") {
		t.Error("SetActiveTabByID('tab3') should return false for non-existent ID")
	}
}

func TestInstance_NextTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
		{ID: "tab3", Label: "Tab 3"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.SetActiveTab(0)

	// Move to next
	if !inst.NextTab() {
		t.Error("NextTab() should return true")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}

	// Move to last
	if !inst.NextTab() {
		t.Error("NextTab() should return true")
	}
	if inst.activeTab != 2 {
		t.Errorf("Expected activeTab 2, got %d", inst.activeTab)
	}

	// No more tabs
	if inst.NextTab() {
		t.Error("NextTab() should return false when at last tab")
	}
}

func TestInstance_PreviousTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.SetActiveTab(1)

	// Move to previous
	if !inst.PreviousTab() {
		t.Error("PreviousTab() should return true")
	}
	if inst.activeTab != 0 {
		t.Errorf("Expected activeTab 0, got %d", inst.activeTab)
	}

	// No more tabs
	if inst.PreviousTab() {
		t.Error("PreviousTab() should return false when at first tab")
	}
}

func TestInstance_FirstTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.SetActiveTab(1)

	if !inst.FirstTab() {
		t.Error("FirstTab() should return true")
	}
	if inst.activeTab != 0 {
		t.Errorf("Expected activeTab 0, got %d", inst.activeTab)
	}
}

func TestInstance_LastTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.SetActiveTab(0)

	if !inst.LastTab() {
		t.Error("LastTab() should return true")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}
}

func TestInstance_DisabledTabs(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1", Disabled: false},
		{ID: "tab2", Label: "Tab 2", Disabled: true},
		{ID: "tab3", Label: "Tab 3", Disabled: false},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.SetActiveTab(0)

	// Next should skip disabled tab 2
	if !inst.NextTab() {
		t.Error("NextTab() should return true")
	}
	if inst.activeTab != 2 {
		t.Errorf("Expected activeTab 2 (skipping disabled tab), got %d", inst.activeTab)
	}
}

func TestInstance_Measure(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab One"},
		{ID: "tab2", Label: "Tab Two"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs, "width": 80, "height": 20})

	size := inst.Measure(layout.Constraints{})
	if size.Width != 80 {
		t.Errorf("Expected width 80, got %d", size.Width)
	}
	if size.Height != 20 { // Explicit height includes tab bar
		t.Errorf("Expected height 20, got %d", size.Height)
	}
}

func TestInstance_Paint(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":           tabs,
		"tabStyle":       style.Style{FG: style.Color("white")},
		"activeTabStyle": style.Style{FG: style.Color("cyan")},
	})

	cmds := inst.Paint(0, 0)
	if cmds == nil {
		t.Error("Paint() should return commands")
	}
	if len(cmds) == 0 {
		t.Error("Paint() should return non-empty commands")
	}
	if len(inst.tabBarBounds) != 2 {
		t.Errorf("Expected 2 tab bounds, got %d", len(inst.tabBarBounds))
	}
}

func TestInstance_GetActiveTabInfo(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.SetActiveTab(1)

	if inst.GetActiveTab() != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.GetActiveTab())
	}
	if inst.GetActiveTabID() != "tab2" {
		t.Errorf("Expected ID 'tab2', got '%s'", inst.GetActiveTabID())
	}
	if inst.GetActiveTabLabel() != "Tab 2" {
		t.Errorf("Expected label 'Tab 2', got '%s'", inst.GetActiveTabLabel())
	}
}

func TestTabsEqual(t *testing.T) {
	tabs1 := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}
	tabs2 := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}
	tabs3 := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
	}

	if !tabsEqual(tabs1, tabs2) {
		t.Error("Expected equal tabs")
	}
	if tabsEqual(tabs1, tabs3) {
		t.Error("Expected unequal tabs")
	}
}
