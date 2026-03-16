package tabs

import (
	"testing"

	"github.com/wwsheng009/mint/framework/styling"
	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type TestCloseCustomIntent struct{}

func (TestCloseCustomIntent) IntentType() string { return "Tabs:TestClose" }

type TestReorderCustomIntent struct{}

func (TestReorderCustomIntent) IntentType() string { return "Tabs:TestReorder" }

func withTabsTestStyleGetter(t *testing.T, getter func(string, string) style.Style) {
	t.Helper()

	style.RegisterStyleGetter(func() func(string, string) style.Style {
		return getter
	})
	t.Cleanup(func() {
		style.RegisterStyleGetter(func() func(string, string) style.Style { return nil })
	})
}

func withTabsTestTheme(t *testing.T, theme *fwtheme.Theme) {
	t.Helper()

	mgr := fwtheme.GlobalManager()
	oldTheme := fwtheme.GetTheme()
	mgr.Register(theme)
	if err := mgr.Set(theme.Name); err != nil {
		t.Fatalf("Set(%q) error = %v", theme.Name, err)
	}

	t.Cleanup(func() {
		if oldTheme != nil {
			_ = mgr.Set(oldTheme.Name)
		}
		mgr.Unregister(theme.Name)
	})
}

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

func TestTabItem_WithClosable(t *testing.T) {
	tab := Item("tab1", "Tab 1").WithClosable(true)
	if !tab.Closable {
		t.Fatal("Expected WithClosable(true) to mark tab as closable")
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

func TestVNode_CardVariant(t *testing.T) {
	vnode := New().Card()
	if vnode.TabVariant() != TabVariantCard {
		t.Fatalf("TabVariant = %v, want %v", vnode.TabVariant(), TabVariantCard)
	}

	built := NewBuilder().Card().BuildVNode()
	if built.TabVariant() != TabVariantCard {
		t.Fatalf("builder TabVariant = %v, want %v", built.TabVariant(), TabVariantCard)
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

func TestBuilder_OnClose(t *testing.T) {
	vnode := NewBuilder().
		OnClose(TestCloseCustomIntent{}).
		BuildVNode()

	if _, ok := vnode.CloseIntent().(TestCloseCustomIntent); !ok {
		t.Fatalf("Expected close intent to be TestCloseCustomIntent, got %T", vnode.CloseIntent())
	}

	inst := vnode.CreateInstance().(*Instance)
	if _, ok := inst.closeIntent.(TestCloseCustomIntent); !ok {
		t.Fatalf("Expected instance close intent to be TestCloseCustomIntent, got %T", inst.closeIntent)
	}
}

func TestBuilder_OnReorder(t *testing.T) {
	vnode := NewBuilder().
		Reorderable(true).
		OnReorder(TestReorderCustomIntent{}).
		BuildVNode()

	if !vnode.Reorderable() {
		t.Fatal("Expected vnode to be reorderable")
	}
	if _, ok := vnode.ReorderIntent().(TestReorderCustomIntent); !ok {
		t.Fatalf("Expected reorder intent to be TestReorderCustomIntent, got %T", vnode.ReorderIntent())
	}

	inst := vnode.CreateInstance().(*Instance)
	if !inst.reorderable {
		t.Fatal("Expected instance to be reorderable")
	}
	if _, ok := inst.reorderIntent.(TestReorderCustomIntent); !ok {
		t.Fatalf("Expected instance reorder intent to be TestReorderCustomIntent, got %T", inst.reorderIntent)
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

func TestInstance_Paint_CardVariant(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":       tabs,
		"tabVariant": TabVariantCard,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) < 3 {
		t.Fatalf("expected tab bar and divider commands, got %d", len(cmds))
	}
	if cmds[0].Text != "╭ Tab 1 ╮" {
		t.Fatalf("active card text = %q, want %q", cmds[0].Text, "╭ Tab 1 ╮")
	}
	if cmds[2].Text != "│ Tab 2 │" {
		t.Fatalf("inactive card text = %q, want %q", cmds[2].Text, "│ Tab 2 │")
	}
}

func TestInstance_Paint_CardVariantClosable(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1", Closable: true},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":       tabs,
		"tabVariant": TabVariantCard,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) < 3 {
		t.Fatalf("expected tab bar and divider commands, got %d", len(cmds))
	}
	if cmds[0].Text != "╭ Tab 1 × ╮" {
		t.Fatalf("active closable card text = %q, want %q", cmds[0].Text, "╭ Tab 1 × ╮")
	}
	if inst.tabBarBounds[0].closeW == 0 {
		t.Fatal("Expected closable tab to expose a close hitbox")
	}
}

func TestInstance_ResolveTabStyle_ActiveTabPreservesThemeSelectForegroundForContrast(t *testing.T) {
	style.RegisterStyleGetter(func() func(string, string) style.Style {
		return func(componentID, state string) style.Style {
			if componentID == "tabs" && state == "select" {
				return style.NewStyle().
					Foreground(style.Red).
					Background(style.Yellow).
					Underline(true)
			}
			return style.Style{}
		}
	})
	t.Cleanup(func() {
		style.RegisterStyleGetter(func() func(string, string) style.Style { return nil })
	})

	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
		},
		"activeTabStyle": style.NewStyle().Foreground(style.Cyan),
	})

	active := inst.resolveTabStyle(0)
	if active.FG != style.Red {
		t.Fatalf("Expected theme select FG to be preserved for contrast, got %s", active.FG)
	}
	if active.BG != style.Yellow {
		t.Fatalf("Expected theme select BG to be preserved, got %s", active.BG)
	}
	if !active.IsUnderline() {
		t.Fatal("Expected theme select underline to be applied to active tab")
	}

	inactive := inst.resolveTabStyle(1)
	if inactive.BG != style.NoColor {
		t.Fatalf("Inactive tab should not inherit theme select BG, got %s", inactive.BG)
	}
	if inactive.IsUnderline() {
		t.Fatal("Inactive tab should not inherit theme select underline")
	}
}

func TestInstance_ResolveTabStyle_ActiveTabFallsBackToThemeSelectBackground(t *testing.T) {
	mgr, err := fwtheme.InitThemes("nord")
	if err != nil {
		t.Fatalf("InitThemes(nord) error = %v", err)
	}

	oldProvider := styling.GetProvider()
	styling.SetProvider(fwtheme.NewThemeStyleProvider(mgr))
	t.Cleanup(func() {
		styling.SetProvider(oldProvider)
	})

	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
		},
		"activeTabStyle": style.NewStyle().Foreground(style.Cyan),
	})

	active := inst.resolveTabStyle(0)
	if active.BG != fwtheme.Select() {
		t.Fatalf("Expected active tab fallback BG %s from theme select color, got %s", fwtheme.Select(), active.BG)
	}
	if active.FG != fwtheme.BG() {
		t.Fatalf("Expected active tab fallback FG %s for contrast, got %s", fwtheme.BG(), active.FG)
	}
}

func TestInstance_ResolveTabStyle_ReverseFallbackClearsCallerColors(t *testing.T) {
	withTabsTestStyleGetter(t, nil)

	theme := fwtheme.NewTheme("tabs-reverse-fallback")
	theme.Colors.Select = fwtheme.NoColor
	theme.Colors.BG = fwtheme.NoColor
	withTabsTestTheme(t, theme)

	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
		},
		"tabStyle":       style.NewStyle().Foreground(style.Green).Background(style.Yellow),
		"activeTabStyle": style.NewStyle().Foreground(style.Cyan).Background(style.Red).Underline(true),
	})

	active := inst.resolveTabStyle(0)
	if active.FG != style.NoColor {
		t.Fatalf("Expected reverse fallback to clear FG, got %s", active.FG)
	}
	if active.BG != style.NoColor {
		t.Fatalf("Expected reverse fallback to clear BG, got %s", active.BG)
	}
	if !active.IsReverse() {
		t.Fatal("Expected reverse fallback to set reverse")
	}
	if !active.IsBold() {
		t.Fatal("Expected reverse fallback to keep active tab bold")
	}
	if !active.IsUnderline() {
		t.Fatal("Expected non-color active flags to survive reverse fallback")
	}
}

func TestInstance_ResolveTabStyle_DisabledTabFallsBackToItalicWhenSemanticColorMissing(t *testing.T) {
	withTabsTestStyleGetter(t, nil)

	theme := fwtheme.NewTheme("tabs-disabled-fallback")
	theme.Colors.DisabledFG = fwtheme.NoColor
	withTabsTestTheme(t, theme)

	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Disabled: true},
			{ID: "tab2", Label: "Tab 2"},
		},
	})

	disabled := inst.resolveTabStyle(0)
	if disabled.FG != style.NoColor {
		t.Fatalf("Expected disabled fallback FG to remain unset, got %s", disabled.FG)
	}
	if !disabled.IsItalic() {
		t.Fatal("Expected disabled fallback to use italic when theme disabled color is missing")
	}
}

func TestInstance_ResolveTabStyle_BGOnlySelectStyleUsesThemeBGForContrast(t *testing.T) {
	withTabsTestStyleGetter(t, func(componentID, state string) style.Style {
		if componentID == "tabs" && state == "select" {
			return style.NewStyle().Background(style.Yellow).Underline(true)
		}
		return style.Style{}
	})

	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
		},
		"tabStyle":       style.NewStyle().Foreground(style.Green),
		"activeTabStyle": style.NewStyle().Foreground(style.Cyan),
	})

	active := inst.resolveTabStyle(0)
	if active.BG != style.Yellow {
		t.Fatalf("Expected BG-only select style to set BG yellow, got %s", active.BG)
	}
	if active.FG != fwtheme.BG() {
		t.Fatalf("Expected BG-only select style to protect FG with theme BG %s, got %s", fwtheme.BG(), active.FG)
	}
	if !active.IsUnderline() {
		t.Fatal("Expected BG-only select style flags to survive")
	}
}

func TestInstance_ResolveTabStyle_ReverseOnlySelectStyleClearsCallerColors(t *testing.T) {
	withTabsTestStyleGetter(t, func(componentID, state string) style.Style {
		if componentID == "tabs" && state == "select" {
			return style.NewStyle().Reverse(true)
		}
		return style.Style{}
	})

	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
		},
		"tabStyle":       style.NewStyle().Foreground(style.Green).Background(style.Yellow),
		"activeTabStyle": style.NewStyle().Foreground(style.Cyan).Background(style.Red).Underline(true),
	})

	active := inst.resolveTabStyle(0)
	if active.FG != style.NoColor {
		t.Fatalf("Expected reverse-only select style to clear FG, got %s", active.FG)
	}
	if active.BG != style.NoColor {
		t.Fatalf("Expected reverse-only select style to clear BG, got %s", active.BG)
	}
	if !active.IsReverse() {
		t.Fatal("Expected reverse-only select style to keep reverse")
	}
	if !active.IsUnderline() {
		t.Fatal("Expected non-color active flags to survive reverse-only select style")
	}
}

func TestInstance_ResolveTabStyle_NonProtectableSelectStyleFallsBackToSemanticTheme(t *testing.T) {
	withTabsTestStyleGetter(t, func(componentID, state string) style.Style {
		if componentID == "tabs" && state == "select" {
			return style.NewStyle().Underline(true)
		}
		return style.Style{}
	})

	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
		},
		"tabStyle":       style.NewStyle().Foreground(style.Green),
		"activeTabStyle": style.NewStyle().Foreground(style.Cyan),
	})

	active := inst.resolveTabStyle(0)
	if active.BG != fwtheme.Select() {
		t.Fatalf("Expected non-protectable select style to fall back to semantic BG %s, got %s", fwtheme.Select(), active.BG)
	}
	if active.FG != fwtheme.BG() {
		t.Fatalf("Expected non-protectable select style to fall back to semantic FG %s, got %s", fwtheme.BG(), active.FG)
	}
	if active.IsUnderline() {
		t.Fatal("Expected non-protectable select style flags to be ignored in favor of semantic fallback")
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

func TestInstance_HandleAction_ClickUsesMousePayload(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Home"},
		{ID: "tab2", Label: "Settings"},
	}

	inst := NewInstance(rtui.Props{"tabs": tabs})
	inst.Paint(12, 4)

	mouse := runtimemsg.NewMouseMsgWithTarget(21, 4, 9, 0, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	handled := inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(mouse))
	if !handled {
		t.Fatal("Expected ActionClick with MouseMsg payload to activate second tab")
	}
	if inst.activeTab != 1 {
		t.Errorf("Expected activeTab 1, got %d", inst.activeTab)
	}
}

func TestInstance_HandleMouseMessage_DragReordersTabs(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
			{ID: "tab3", Label: "Tab 3"},
		},
		"reorderable": true,
	})

	inst.Paint(0, 0)
	first := inst.tabBarBounds[0]
	last := inst.tabBarBounds[2]

	if !inst.HandleMouseMessage(&runtimemsg.MouseMsg{
		LocalX: first.x,
		LocalY: first.y,
		Action: runtimemsg.MouseActionPress,
		Button: runtimemsg.MouseLeft,
	}) {
		t.Fatal("Expected press to start drag on active tab when reorderable")
	}
	if !inst.dragging {
		t.Fatal("Expected instance to enter dragging state after press")
	}

	if !inst.HandleMouseMessage(&runtimemsg.MouseMsg{
		LocalX: last.x + last.width - 1,
		LocalY: last.y,
		Action: runtimemsg.MouseActionMove,
		Button: runtimemsg.MouseLeft,
	}) {
		t.Fatal("Expected move to reorder tabs while dragging")
	}

	if !inst.HandleMouseMessage(&runtimemsg.MouseMsg{
		LocalX: last.x + last.width - 1,
		LocalY: last.y,
		Action: runtimemsg.MouseActionRelease,
		Button: runtimemsg.MouseLeft,
	}) {
		t.Fatal("Expected release to finish drag")
	}

	if inst.dragging {
		t.Fatal("Expected drag state to clear after release")
	}

	got := []string{inst.tabs[0].ID, inst.tabs[1].ID, inst.tabs[2].ID}
	want := []string{"tab2", "tab3", "tab1"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tabs[%d] = %q, want %q; full order=%v", i, got[i], want[i], got)
		}
	}
	if inst.activeTab != 2 || inst.GetActiveTabID() != "tab1" {
		t.Fatalf("Expected dragged active tab to remain active at index 2, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_HandleAction_DragReordersTabs(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
			{ID: "tab3", Label: "Tab 3"},
		},
		"reorderable": true,
	})

	inst.Paint(10, 4)
	first := inst.tabBarBounds[0]
	last := inst.tabBarBounds[2]

	press := runtimemsg.NewMouseMsgWithTarget(10+first.x, 4+first.y, first.x, first.y, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	if !inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(press)) {
		t.Fatal("Expected ActionClick to start drag when reorderable")
	}

	move := runtimemsg.NewMouseMsgWithTarget(10+last.x+last.width-1, 4+last.y, last.x+last.width-1, last.y, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionMove)
	if !inst.HandleAction(action.NewAction(action.ActionHover).WithPayload(move)) {
		t.Fatal("Expected ActionHover to drive drag reordering")
	}

	release := runtimemsg.NewMouseMsgWithTarget(10+last.x+last.width-1, 4+last.y, last.x+last.width-1, last.y, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	if !inst.HandleAction(action.NewAction(action.ActionMouseRelease).WithPayload(release)) {
		t.Fatal("Expected ActionMouseRelease to finish drag")
	}

	if inst.GetActiveTabID() != "tab1" || inst.activeTab != 2 {
		t.Fatalf("Expected dragged tab1 to end active at index 2, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
	if inst.tabs[2].ID != "tab1" {
		t.Fatalf("Expected tab1 to move to the end, got order=%v", []string{inst.tabs[0].ID, inst.tabs[1].ID, inst.tabs[2].ID})
	}
}

func TestInstance_DragReorder_PreservesNonDraggedActiveTabByID(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
			{ID: "tab3", Label: "Tab 3"},
		},
		"activeTab":   1,
		"reorderable": true,
	})

	inst.Paint(0, 0)
	first := inst.tabBarBounds[0]
	last := inst.tabBarBounds[2]

	inst.HandleMouseMessage(&runtimemsg.MouseMsg{LocalX: first.x, LocalY: first.y, Action: runtimemsg.MouseActionPress, Button: runtimemsg.MouseLeft})
	inst.HandleMouseMessage(&runtimemsg.MouseMsg{LocalX: last.x + last.width - 1, LocalY: last.y, Action: runtimemsg.MouseActionMove, Button: runtimemsg.MouseLeft})
	inst.HandleMouseMessage(&runtimemsg.MouseMsg{LocalX: last.x + last.width - 1, LocalY: last.y, Action: runtimemsg.MouseActionRelease, Button: runtimemsg.MouseLeft})

	if inst.GetActiveTabID() != "tab2" || inst.activeTab != 0 {
		t.Fatalf("Expected previously active tab2 to remain active after reorder, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_DragReorder_CloseHitboxStillCloses(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2"},
		},
		"reorderable": true,
	})

	inst.Paint(0, 0)
	tb := inst.tabBarBounds[0]

	if !inst.HandleMouseMessage(&runtimemsg.MouseMsg{
		LocalX: tb.closeX,
		LocalY: tb.y,
		Action: runtimemsg.MouseActionPress,
		Button: runtimemsg.MouseLeft,
	}) {
		t.Fatal("Expected close hitbox press to close tab even when reorderable")
	}
	if inst.dragging {
		t.Fatal("Expected close hitbox not to leave drag state behind")
	}
	if inst.GetTabCount() != 1 || inst.GetActiveTabID() != "tab2" {
		t.Fatalf("Expected tab1 to close, got count=%d active=%q", inst.GetTabCount(), inst.GetActiveTabID())
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

func TestInstance_CloseTab_ClosesActiveAndSelectsNext(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
			{ID: "tab3", Label: "Tab 3", Closable: true},
		},
		"activeTab": 1,
	})

	if !inst.CloseTab(1) {
		t.Fatal("Expected CloseTab(1) to succeed")
	}
	if inst.GetTabCount() != 2 {
		t.Fatalf("Expected 2 tabs after close, got %d", inst.GetTabCount())
	}
	if inst.activeTab != 1 || inst.GetActiveTabID() != "tab3" {
		t.Fatalf("Expected active tab to move to tab3 at index 1, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_CloseTab_ClosesLastActiveAndSelectsPrevious(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
			{ID: "tab3", Label: "Tab 3", Closable: true},
		},
		"activeTab": 2,
	})

	if !inst.CloseTab(2) {
		t.Fatal("Expected CloseTab(2) to succeed")
	}
	if inst.activeTab != 1 || inst.GetActiveTabID() != "tab2" {
		t.Fatalf("Expected active tab to move to previous tab2, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_CloseTab_ClosesBeforeActiveAndPreservesActiveTabID(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
			{ID: "tab3", Label: "Tab 3", Closable: true},
		},
		"activeTab": 2,
	})

	if !inst.CloseTab(0) {
		t.Fatal("Expected CloseTab(0) to succeed")
	}
	if inst.activeTab != 1 || inst.GetActiveTabID() != "tab3" {
		t.Fatalf("Expected tab3 to remain active at index 1, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_CloseTab_OnlyTabLeavesNoActiveTab(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{{ID: "tab1", Label: "Tab 1", Closable: true}},
	})

	if !inst.CloseTab(0) {
		t.Fatal("Expected CloseTab(0) to succeed")
	}
	if inst.GetTabCount() != 0 {
		t.Fatalf("Expected 0 tabs after close, got %d", inst.GetTabCount())
	}
	if inst.activeTab != -1 || inst.GetActiveTabID() != "" {
		t.Fatalf("Expected no active tab after closing the last tab, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_HandleClick_CloseHitboxClosesTab(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
		},
		"tabVariant": TabVariantCard,
	})

	inst.Paint(0, 0)
	tb := inst.tabBarBounds[0]
	if tb.closeW == 0 {
		t.Fatal("Expected close hitbox for the first tab")
	}

	if !inst.handleClick(map[string]interface{}{"localX": tb.closeX, "localY": tb.y}) {
		t.Fatal("Expected click on close hitbox to close the tab")
	}
	if inst.GetTabCount() != 1 || inst.GetActiveTabID() != "tab2" {
		t.Fatalf("Expected first tab to be removed and tab2 to become active, count=%d active=%q", inst.GetTabCount(), inst.GetActiveTabID())
	}
}

func TestInstance_SetProps_RemovedSiblingPreservesActiveTabByID(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
			{ID: "tab3", Label: "Tab 3", Closable: true},
		},
		"activeTab": 2,
	})

	changed := inst.SetProps(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab3", Label: "Tab 3", Closable: true},
		},
		"activeTab": 2,
	})
	if !changed {
		t.Fatal("Expected SetProps to report changed when tabs shrink")
	}
	if inst.activeTab != 1 || inst.GetActiveTabID() != "tab3" {
		t.Fatalf("Expected tab3 to remain active at index 1, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_SetProps_RemovedActiveSelectsClosestTab(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
			{ID: "tab3", Label: "Tab 3", Closable: true},
		},
		"activeTab": 2,
	})

	inst.SetProps(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
		},
		"activeTab": 2,
	})
	if inst.activeTab != 1 || inst.GetActiveTabID() != "tab2" {
		t.Fatalf("Expected closest surviving tab2 to become active, got index=%d id=%q", inst.activeTab, inst.GetActiveTabID())
	}
}

func TestInstance_IsFocusable(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{{ID: "tab1", Label: "Tab 1"}},
	})

	focusable, ok := interface{}(inst).(rtui.FocusableInstance)
	if !ok {
		t.Fatal("tabs instance should implement FocusableInstance")
	}

	focusable.SetFocus(true)
	if !focusable.HasFocus() {
		t.Fatal("tabs instance should track focus state")
	}

	focusable.SetFocus(false)
	if focusable.HasFocus() {
		t.Fatal("tabs instance should clear focus state")
	}
}

func TestInstance_HandleAction_KeyboardNavigationAndHotkeys(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1", Hotkey: 'h'},
		{ID: "tab2", Label: "Tab 2", Hotkey: 'p'},
		{ID: "tab3", Label: "Tab 3", Hotkey: 's'},
	}

	inst := NewInstance(rtui.Props{
		"tabs":           tabs,
		"loopNavigation": true,
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateRight)) {
		t.Fatal("ActionNavigateRight should switch to the next tab")
	}
	if inst.activeTab != 1 {
		t.Fatalf("Expected activeTab 1 after ActionNavigateRight, got %d", inst.activeTab)
	}

	if !inst.HandleAction(action.NewAction(action.ActionInputText).WithPayload(
		runtimemsg.NewKeyMsg('s', platform.KeyUnknown, runtimemsg.Modifiers{}),
	)) {
		t.Fatal("ActionInputText with hotkey payload should select matching tab")
	}
	if inst.activeTab != 2 {
		t.Fatalf("Expected activeTab 2 after hotkey action, got %d", inst.activeTab)
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

func withTabsTestIntentRuntime(t *testing.T) *intent.Runtime {
	t.Helper()

	oldRuntime := rtui.GetGlobalIntentRuntime()
	rt := intent.NewRuntimeWithNewRegistry()
	rtui.SetGlobalIntentRuntime(rt)
	t.Cleanup(func() {
		rtui.SetGlobalIntentRuntime(oldRuntime)
	})
	return rt
}

func TestInstance_ComponentIDEmitsTabChangeIntentToGlobalRuntimeWhenHandlerRegistered(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":        tabs,
		"componentID": "workspace-tabs",
	})

	rt := withTabsTestIntentRuntime(t)

	var emitted []TabChangeIntent
	unregister := intent.RegisterTypedRuntime(rt, func(_ *intent.ActionContext, i TabChangeIntent) intent.IntentResult {
		emitted = append(emitted, i)
		return intent.HandledResult()
	})
	defer unregister()

	bubbleCount := 0
	intent.SetBubbleTestHook(func(component interface{}, i intent.Intent) bool {
		if _, ok := i.(TabChangeIntent); ok {
			bubbleCount++
		}
		return false
	})
	defer intent.SetBubbleTestHook(nil)

	if !inst.NextTab() {
		t.Fatal("Expected NextTab to succeed")
	}

	if len(emitted) != 1 {
		t.Fatalf("Expected exactly 1 globally emitted intent, got %d", len(emitted))
	}

	change := emitted[0]
	if change.ComponentID != "workspace-tabs" || change.ActiveTab != 1 || change.TabID != "tab2" || change.TabLabel != "Tab 2" {
		t.Fatalf("Unexpected TabChangeIntent payload: %+v", change)
	}
	if bubbleCount != 1 {
		t.Fatalf("Expected one bubbled TabChangeIntent, got %d", bubbleCount)
	}
}

func TestInstance_ComponentIDAndFieldIntentEmitTabChangeGloballyAndFieldIntentLocally(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":              tabs,
		"componentID":       "workspace-tabs",
		"changeIntentField": intent.FieldChangeIntent{Field: "currentTab"},
	})

	rt := withTabsTestIntentRuntime(t)

	var tabChanges []TabChangeIntent
	unregister := intent.RegisterTypedRuntime(rt, func(_ *intent.ActionContext, i TabChangeIntent) intent.IntentResult {
		tabChanges = append(tabChanges, i)
		return intent.HandledResult()
	})
	defer unregister()

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.NextTab() {
		t.Fatal("Expected NextTab to succeed")
	}

	if len(tabChanges) != 1 {
		t.Fatalf("Expected 1 global TabChangeIntent, got %d", len(tabChanges))
	}
	change := tabChanges[0]
	if change.ComponentID != "workspace-tabs" || change.ActiveTab != 1 || change.TabID != "tab2" || change.TabLabel != "Tab 2" {
		t.Fatalf("Unexpected TabChangeIntent payload: %+v", change)
	}

	if len(emitted) != 1 {
		t.Fatalf("Expected 1 locally emitted intent, got %d", len(emitted))
	}

	fieldIntent, ok := emitted[0].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("Expected locally emitted intent to be FieldChangeIntent, got %T", emitted[0])
	}
	if fieldIntent.Field != "currentTab" || fieldIntent.Value != "1" {
		t.Fatalf("Unexpected FieldChangeIntent payload: %+v", fieldIntent)
	}
}

func TestInstance_ComponentIDAndCustomIntentEmitTabChangeGloballyAndCustomIntentLocally(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":         tabs,
		"componentID":  "workspace-tabs",
		"changeIntent": TestCustomIntent{},
	})

	rt := withTabsTestIntentRuntime(t)

	var tabChanges []TabChangeIntent
	unregister := intent.RegisterTypedRuntime(rt, func(_ *intent.ActionContext, i TabChangeIntent) intent.IntentResult {
		tabChanges = append(tabChanges, i)
		return intent.HandledResult()
	})
	defer unregister()

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.NextTab() {
		t.Fatal("Expected NextTab to succeed")
	}

	if len(tabChanges) != 1 {
		t.Fatalf("Expected 1 global TabChangeIntent, got %d", len(tabChanges))
	}
	if change := tabChanges[0]; change.ComponentID != "workspace-tabs" || change.ActiveTab != 1 || change.TabID != "tab2" || change.TabLabel != "Tab 2" {
		t.Fatalf("Unexpected TabChangeIntent payload: %+v", change)
	}

	if len(emitted) != 1 {
		t.Fatalf("Expected 1 locally emitted intent, got %d", len(emitted))
	}
	if _, ok := emitted[0].(TestCustomIntent); !ok {
		t.Fatalf("Expected locally emitted intent to be TestCustomIntent, got %T", emitted[0])
	}
}

func TestInstance_ComponentIDWithoutEmitterStillBubblesWithoutPanic(t *testing.T) {
	tabs := []TabItem{
		{ID: "tab1", Label: "Tab 1"},
		{ID: "tab2", Label: "Tab 2"},
	}

	inst := NewInstance(rtui.Props{
		"tabs":              tabs,
		"componentID":       "workspace-tabs",
		"changeIntentField": intent.FieldChangeIntent{Field: "currentTab"},
	})

	bubbleCount := 0
	intent.SetBubbleTestHook(func(component interface{}, i intent.Intent) bool {
		if _, ok := i.(TabChangeIntent); ok {
			bubbleCount++
		}
		return false
	})
	defer intent.SetBubbleTestHook(nil)

	if !inst.NextTab() {
		t.Fatal("Expected NextTab to succeed without emitter")
	}

	if bubbleCount != 1 {
		t.Fatalf("Expected one bubbled TabChangeIntent, got %d", bubbleCount)
	}
}

func TestInstance_CloseTab_EmitsLocalCloseIntentAndFieldChangeIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
			{ID: "tab3", Label: "Tab 3", Closable: true},
		},
		"activeTab":         1,
		"closeIntent":       TestCloseCustomIntent{},
		"changeIntentField": intent.FieldChangeIntent{Field: "currentTab"},
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.CloseTab(1) {
		t.Fatal("Expected CloseTab(1) to succeed")
	}
	if len(emitted) != 2 {
		t.Fatalf("Expected 2 emitted intents, got %d", len(emitted))
	}
	if _, ok := emitted[0].(TestCloseCustomIntent); !ok {
		t.Fatalf("Expected first emitted intent to be TestCloseCustomIntent, got %T", emitted[0])
	}
	fieldIntent, ok := emitted[1].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("Expected second emitted intent to be FieldChangeIntent, got %T", emitted[1])
	}
	if fieldIntent.Field != "currentTab" || fieldIntent.Value != "1" {
		t.Fatalf("Unexpected field intent payload after close: %+v", fieldIntent)
	}
}

func TestInstance_CloseTab_EmitsTabCloseIntentToGlobalRuntimeWhenHandlerRegistered(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1", Closable: true},
			{ID: "tab2", Label: "Tab 2", Closable: true},
		},
		"componentID": "workspace-tabs",
	})

	rt := withTabsTestIntentRuntime(t)

	var emitted []TabCloseIntent
	unregister := intent.RegisterTypedRuntime(rt, func(_ *intent.ActionContext, i TabCloseIntent) intent.IntentResult {
		emitted = append(emitted, i)
		return intent.HandledResult()
	})
	defer unregister()

	bubbleCount := 0
	intent.SetBubbleTestHook(func(component interface{}, i intent.Intent) bool {
		if _, ok := i.(TabCloseIntent); ok {
			bubbleCount++
		}
		return false
	})
	defer intent.SetBubbleTestHook(nil)

	if !inst.CloseTabByID("tab1") {
		t.Fatal("Expected CloseTabByID(tab1) to succeed")
	}

	if len(emitted) != 1 {
		t.Fatalf("Expected exactly 1 globally emitted close intent, got %d", len(emitted))
	}

	closeIntent := emitted[0]
	if closeIntent.ComponentID != "workspace-tabs" ||
		closeIntent.ClosedTabIndex != 0 ||
		closeIntent.ClosedTabID != "tab1" ||
		closeIntent.ClosedTabLabel != "Tab 1" ||
		closeIntent.ActiveTab != 0 ||
		closeIntent.ActiveTabID != "tab2" ||
		closeIntent.ActiveTabLabel != "Tab 2" {
		t.Fatalf("Unexpected TabCloseIntent payload: %+v", closeIntent)
	}
	if bubbleCount != 1 {
		t.Fatalf("Expected one bubbled TabCloseIntent, got %d", bubbleCount)
	}
}

func TestInstance_DragReorder_EmitsLocalReorderIntentAndGlobalTabReorderIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"tabs": []TabItem{
			{ID: "tab1", Label: "Tab 1"},
			{ID: "tab2", Label: "Tab 2"},
			{ID: "tab3", Label: "Tab 3"},
		},
		"componentID":   "workspace-tabs",
		"reorderable":   true,
		"reorderIntent": TestReorderCustomIntent{},
	})

	rt := withTabsTestIntentRuntime(t)

	var reordered []TabReorderIntent
	unregister := intent.RegisterTypedRuntime(rt, func(_ *intent.ActionContext, i TabReorderIntent) intent.IntentResult {
		reordered = append(reordered, i)
		return intent.HandledResult()
	})
	defer unregister()

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	inst.Paint(0, 0)
	first := inst.tabBarBounds[0]
	last := inst.tabBarBounds[2]

	inst.HandleMouseMessage(&runtimemsg.MouseMsg{LocalX: first.x, LocalY: first.y, Action: runtimemsg.MouseActionPress, Button: runtimemsg.MouseLeft})
	inst.HandleMouseMessage(&runtimemsg.MouseMsg{LocalX: last.x + last.width - 1, LocalY: last.y, Action: runtimemsg.MouseActionMove, Button: runtimemsg.MouseLeft})
	inst.HandleMouseMessage(&runtimemsg.MouseMsg{LocalX: last.x + last.width - 1, LocalY: last.y, Action: runtimemsg.MouseActionRelease, Button: runtimemsg.MouseLeft})

	if len(emitted) != 1 {
		t.Fatalf("Expected 1 local reorder intent, got %d", len(emitted))
	}
	if _, ok := emitted[0].(TestReorderCustomIntent); !ok {
		t.Fatalf("Expected local reorder intent to be TestReorderCustomIntent, got %T", emitted[0])
	}

	if len(reordered) != 1 {
		t.Fatalf("Expected 1 global TabReorderIntent, got %d", len(reordered))
	}
	reorderIntent := reordered[0]
	if reorderIntent.ComponentID != "workspace-tabs" ||
		reorderIntent.FromIndex != 0 ||
		reorderIntent.ToIndex != 2 ||
		reorderIntent.TabID != "tab1" ||
		reorderIntent.ActiveTab != 2 ||
		reorderIntent.ActiveTabID != "tab1" {
		t.Fatalf("Unexpected TabReorderIntent payload: %+v", reorderIntent)
	}
	orderWant := []string{"tab2", "tab3", "tab1"}
	if len(reorderIntent.TabOrder) != len(orderWant) {
		t.Fatalf("Unexpected TabOrder length: %+v", reorderIntent.TabOrder)
	}
	for i := range orderWant {
		if reorderIntent.TabOrder[i] != orderWant[i] {
			t.Fatalf("TabOrder[%d] = %q, want %q; full=%v", i, reorderIntent.TabOrder[i], orderWant[i], reorderIntent.TabOrder)
		}
	}
}
