package tabs

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/platform"
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
		{ID: "tab1", Label: "Tab 1", Icon: "H", Badge: "1", Hotkey: '1'},
		{ID: "tab2", Label: "Tab 2"},
	}
	tabs2 := []TabItem{
		{ID: "tab1", Label: "Tab 1", Icon: "H", Badge: "1", Hotkey: '1'},
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

func TestTabItem_HelperMethods(t *testing.T) {
	tab := Item("home", "Home").
		WithIcon("H").
		WithBadge("3").
		WithHotkey('h').
		WithDisabled(true).
		WithHidden(true)

	if tab.ID != "home" || tab.Label != "Home" {
		t.Fatal("Item helper should initialize ID and Label")
	}
	if tab.Icon != "H" || tab.Badge != "3" || tab.Hotkey != 'h' {
		t.Fatal("Tab helper methods should configure icon, badge, and hotkey")
	}
	if !tab.Disabled || !tab.Hidden {
		t.Fatal("Tab helper methods should configure disabled/hidden flags")
	}
}

func TestBuilder_EnhancedOptions(t *testing.T) {
	disabledStyle := style.Style{FG: style.Color("bright-black")}

	vnode := NewBuilder().
		ActiveTab(1).
		ActiveTabID("settings").
		LoopNavigation(true).
		ShowHotkeys(true).
		Divider(" / ").
		DisabledTabStyle(disabledStyle).
		BuildVNode()

	if vnode.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", vnode.activeTab)
	}
	if vnode.activeTabID != "settings" {
		t.Errorf("Expected activeTabID 'settings', got %q", vnode.activeTabID)
	}
	if !vnode.loopNavigation || !vnode.showHotkeys {
		t.Fatal("Expected loopNavigation and showHotkeys to be true")
	}
	if vnode.divider != " / " {
		t.Errorf("Expected divider ' / ', got %q", vnode.divider)
	}
	if vnode.disabledTabStyle != disabledStyle {
		t.Fatal("Expected disabledTabStyle to be preserved")
	}
}

func TestInstance_NewInstance_UsesRequestedActiveTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
		{ID: "tab3", Label: "Tab 3", Disabled: true},
	}

	inst := NewInstance(rtui.Props{
		"tabs":      tabs,
		"activeTab": 1,
	})
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}

	inst = NewInstance(rtui.Props{
		"tabs":        tabs,
		"activeTabID": "tab2",
	})
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1 from activeTabID, got %d", inst.activeTab)
	}

	inst = NewInstance(rtui.Props{
		"tabs":        tabs,
		"activeTabID": "tab3",
	})
	if inst.activeTab != 0 {
		t.Errorf("Expected fallback to first enabled tab, got %d", inst.activeTab)
	}
}

func TestInstance_HandleMouseMessage_UsesLocalCoordinates(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Home"},
		{ID: "tab2", Label: "Settings"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.Paint(12, 4)

	handled := inst.HandleMouseMessage(&runtimemsg.MouseMsg{
		LocalX: 9,
		LocalY: 0,
		Action: runtimemsg.MouseActionPress,
	})
	if !handled {
		t.Fatal("Expected local click to activate second tab even when painted at non-zero origin")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}
}

func TestInstance_Paint_WrapTabsCreatesMultipleRows(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Alpha"},
		{ID: "tab2", Label: "Beta"},
		{ID: "tab3", Label: "Gamma"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":     tabs,
		"width":    12,
		"wrapTabs": true,
	})

	inst.Paint(0, 0)

	if len(inst.tabBarBounds) != 3 {
		t.Fatalf("Expected 3 clickable tab bounds, got %d", len(inst.tabBarBounds))
	}
	if inst.tabBarBounds[1].y == 0 && inst.tabBarBounds[2].y == 0 {
		t.Fatal("Expected wrapped tabs to occupy multiple rows")
	}
}

func TestInstance_Paint_RightPositionOffsetsBounds(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "One"},
		{ID: "tab2", Label: "Two"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":     tabs,
		"position": TabPositionRight,
		"width":    12,
	})

	inst.Paint(0, 0)

	if len(inst.tabBarBounds) != 2 {
		t.Fatalf("Expected 2 clickable tab bounds, got %d", len(inst.tabBarBounds))
	}
	if inst.tabBarBounds[0].x <= 0 {
		t.Fatalf("Expected right-positioned tabs to be offset horizontally, got x=%d", inst.tabBarBounds[0].x)
	}
}

func TestInstance_HandleKeyMessage_HotkeysAndCtrlTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "home", Label: "Home", Hotkey: 'h'},
		{ID: "settings", Label: "Settings", Hotkey: 's'},
		{ID: "logs", Label: "Logs"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":           tabs,
		"loopNavigation": true,
	})

	if !inst.HandleKeyMessage(runtimemsg.NewKeyMsg('s', platform.KeyUnknown, runtimemsg.Modifiers{})) {
		t.Fatal("Expected hotkey to select the settings tab")
	}
	if inst.activeTab != 1 {
		t.Fatalf("Expected activeTab 1 after hotkey, got %d", inst.activeTab)
	}

	if !inst.HandleKeyMessage(runtimemsg.NewKeyMsg(0, platform.KeyTab, runtimemsg.Modifiers{Ctrl: true})) {
		t.Fatal("Expected Ctrl+Tab to move to the next tab")
	}
	if inst.activeTab != 2 {
		t.Fatalf("Expected activeTab 2 after Ctrl+Tab, got %d", inst.activeTab)
	}

	if !inst.HandleKeyMessage(runtimemsg.NewKeyMsg(0, platform.KeyTab, runtimemsg.Modifiers{Ctrl: true, Shift: true})) {
		t.Fatal("Expected Ctrl+Shift+Tab to move to the previous tab")
	}
	if inst.activeTab != 1 {
		t.Fatalf("Expected activeTab 1 after Ctrl+Shift+Tab, got %d", inst.activeTab)
	}
}

func TestInstance_LoopNavigation_DoesNotCycleSingleTab(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs":           []TabItem{{ID: "only", Label: "Only"}},
		"loopNavigation": true,
	})

	if inst.CanGoNext() {
		t.Fatal("Expected CanGoNext to be false when only one tab is selectable")
	}
	if inst.NextTab() {
		t.Fatal("Expected NextTab to return false when only one tab is selectable")
	}
}

func TestInstance_SetTabEnabled_RehomesActiveTab(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
		{ID: "tab3", Label: "Tab 3"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs, "activeTab": 1})
	if !inst.SetTabEnabled(1, false) {
		t.Fatal("Expected SetTabEnabled to succeed")
	}
	if inst.activeTab != 0 {
		t.Fatalf("Expected active tab to fall back to the first enabled tab, got %d", inst.activeTab)
	}
}

func TestInstance_FieldIntentPropExtraction(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":              tabs,
		"changeIntentField": intent.FieldChangeIntent{Field: "currentTab"},
	})

	if inst.changeIntentField == nil {
		t.Fatal("Expected changeIntentField to be extracted from props")
	}

	var emitted intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = i
	})

	if !inst.NextTab() {
		t.Fatal("Expected NextTab to succeed")
	}

	fieldIntent, ok := emitted.(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("Expected emitted FieldChangeIntent, got %T", emitted)
	}
	if fieldIntent.Field != "currentTab" || fieldIntent.Value != "1" {
		t.Fatalf("Unexpected field intent payload: %+v", fieldIntent)
	}
}
