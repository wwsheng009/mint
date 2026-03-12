package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

type Tab int

const (
	TabHome Tab = iota
	TabProfile
	TabSettings
)

const tabsComponentID = "tabs-demo-main"

// =============================================================================
// AppState
// =============================================================================

type AppState struct {
	ActiveTab Tab
}

var tabsStore = store.NewStore(AppState{
	ActiveTab: TabHome,
})

// =============================================================================
// Reducer Registration
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(ui.TabChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			change, ok := i.(ui.TabChangeIntent)
			if !ok || change.ComponentID != tabsComponentID {
				return s
			}
			s.ActiveTab = Tab(change.ActiveTab)
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), tabsStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	ui.Run(MainComponent,
		ui.WithWidth(56),
		ui.WithHeight(20),
		ui.WithTitle("Tabs Demo"),
	)
}

// =============================================================================
// Main Component
// =============================================================================

func MainComponent() ui.VNode {
	activeTab := ui.UseStoreSelector(tabsStore, func(s AppState) Tab { return s.ActiveTab })

	return ui.VStack(
		ui.NewTextBuilder("Tabs Demo").Bold(true).FgColor("cyan").Build(),
		ui.NewTextBuilder("Real tabs component + store-driven content").FgColor("bright-black").Build(),
		ui.Text(""),
		ui.NewTabsBuilder().
			ComponentID(tabsComponentID).
			Tabs([]ui.TabItem{
				ui.NewTabItem("home", "Home").WithIcon("H").WithHotkey('h'),
				ui.NewTabItem("profile", "Profile").WithIcon("P").WithHotkey('p'),
				ui.NewTabItem("settings", "Settings").WithIcon("S").WithBadge("2").WithHotkey('s'),
				ui.NewTabItem("admin", "Admin").WithIcon("X").WithDisabled(true),
			}).
			ActiveTab(int(activeTab)).
			ShowHotkeys(true).
			LoopNavigation(true).
			Width(52).
			ActiveTabStyle(style.NewStyle().Foreground(style.Cyan).Bold(true)).
			DisabledTabStyle(style.NewStyle().Foreground(style.BrightBlack)).
			Build(),
		ui.NewTextBuilder("Use arrows, Ctrl+Tab, or H/P/S to switch tabs.").FgColor("bright-black").Build(),
		ui.Text(""),
		ui.Text("────────────────────────────────────────────────────"),
		ui.Text(""),
		TabContent(activeTab),
	)
}

// =============================================================================
// Tab Content
// =============================================================================

func TabContent(tab Tab) ui.VNode {
	switch tab {
	case TabHome:
		return ui.VStack(
			ui.NewTextBuilder("Welcome to the Home tab!").FgColor("green").Build(),
			ui.Text(""),
			ui.NewTextBuilder("This example now uses the actual tabs component.").FgColor("gray").Build(),
			ui.Text(""),
			ui.NewTextBuilder("State still lives in the store, not inside the view.").FgColor("bright-black").Build(),
		)
	case TabProfile:
		return ui.VStack(
			ui.NewTextBuilder("User Profile").FgColor("cyan").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Name:   John Doe").Build(),
			ui.NewTextBuilder("Email:  john@example.com").Build(),
			ui.NewTextBuilder("Role:   Administrator").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Member since: January 2025").FgColor("gray").Build(),
		)
	case TabSettings:
		return ui.VStack(
			ui.NewTextBuilder("System Settings").FgColor("yellow").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Theme:         Dark").Build(),
			ui.NewTextBuilder("Language:      English").Build(),
			ui.NewTextBuilder("Auto-save:     Enabled").Build(),
			ui.NewTextBuilder("Notifications: On").FgColor("green").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Badge on this tab hints there are pending changes.").FgColor("bright-black").Build(),
		)
	default:
		return ui.Text("")
	}
}
