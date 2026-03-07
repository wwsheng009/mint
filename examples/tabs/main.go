package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// Tab 枚举
type Tab int

const (
	TabHome Tab = iota
	TabProfile
	TabSettings
)

// =============================================================================
// AppState - Tabs 状态
// =============================================================================

type AppState struct {
	ActiveTab Tab // 0 = Home, 1 = Profile, 2 = Settings
}

// =============================================================================
// Intent Types
// =============================================================================

type SetHomeTabIntent struct{}
func (SetHomeTabIntent) IntentType() string    { return "SetHomeTab" }
func (SetHomeTabIntent) StayPressed() bool     { return true }

type SetProfileTabIntent struct{}
func (SetProfileTabIntent) IntentType() string { return "SetProfileTab" }
func (SetProfileTabIntent) StayPressed() bool  { return true }

type SetSettingsTabIntent struct{}
func (SetSettingsTabIntent) IntentType() string { return "SetSettingsTab" }
func (SetSettingsTabIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var tabsStore = store.NewStore(AppState{
	ActiveTab: TabHome,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(SetHomeTabIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ActiveTab = TabHome
			return s
		}).
		On(SetProfileTabIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ActiveTab = TabProfile
			return s
		}).
		On(SetSettingsTabIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ActiveTab = TabSettings
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), tabsStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// ✅ 无需 ui.On，Reducer 已在 init 中注册
	ui.Run(MainComponent,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Tabs Demo (Store 模式)"),
	)
}

// =============================================================================
// Main Component
// =============================================================================

func MainComponent() ui.VNode {
	// ✅ 订阅活动标签页
	activeTab := ui.UseStoreSelector(tabsStore, func(s AppState) Tab { return s.ActiveTab })

	// 根据活动标签页选择内容
	content := TabContent(activeTab)

	return ui.VStack(
		ui.NewTextBuilder("Tabs Demo").Bold(true).FgColor("cyan").Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder(" Home ").
				OnPress(SetHomeTabIntent{}).
				Build(),
			ui.NewButtonBuilder(" Profile ").
				OnPress(SetProfileTabIntent{}).
				Build(),
			ui.NewButtonBuilder(" Settings ").
				OnPress(SetSettingsTabIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.Text("─────────────────────────────────────"),
		ui.Text(""),
		content,
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
			ui.NewTextBuilder("This is the main content area.").FgColor("gray").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Navigate using the buttons below:").FgColor("bright-black").Build(),
		)
	case TabProfile:
		return ui.VStack(
			ui.NewTextBuilder("User Profile").FgColor("cyan").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Name:   John Doe").Build(),
			ui.NewTextBuilder("Email:  john@example.com").Build(),
			ui.NewTextBuilder("Role:   Administrator").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Member since: Jan 2025").FgColor("gray").Build(),
		)
	case TabSettings:
		return ui.VStack(
			ui.NewTextBuilder("System Settings").FgColor("yellow").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Theme:     Dark").Build(),
			ui.NewTextBuilder("Language:  English").Build(),
			ui.NewTextBuilder("Auto-save:  Enabled").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Notifications: On").FgColor("green").Build(),
		)
	default:
		return ui.Text("")
	}
}
