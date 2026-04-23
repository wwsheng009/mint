package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// ============================================================================
// AppState - 定义应用状态
// ============================================================================

type AppState struct {
	SelectedIndex int // 选中的主题索引
}

// ============================================================================
// Intent Types
// ============================================================================

type SetThemeIntent struct {
	Index int
}
func (SetThemeIntent) IntentType() string { return "SetTheme" }
func (SetThemeIntent) StayPressed() bool  { return true }

// ============================================================================
// Store 初始化
// ============================================================================

var selectStore = store.NewStore(AppState{
	SelectedIndex: 0,
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(SetThemeIntent{}, func(s AppState, i intent.Intent) AppState {
			s.SelectedIndex = i.(SetThemeIntent).Index
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), selectStore)
}

// ============================================================================
// 主题映射
// ============================================================================

// 索引到主题值的映射
var indexToTheme = map[int]string{
	0: "dark",
	1: "light",
	2: "dracula",
	3: "nord",
}

// 主题值到显示名称的映射
var themeNames = map[string]string{
	"dark":    "Dark Theme",
	"light":   "Light Theme",
	"dracula": "Dracula Theme",
	"nord":    "Nord Theme",
}

// ============================================================================
// SelectDemo demonstrates the select dropdown component (Store 模式)
// ============================================================================

func SelectDemo() ui.VNode {
	// ✅ 订阅 selectedIndex 状态
	selectedIndex := ui.UseStoreSelector(selectStore, func(s AppState) int { return s.SelectedIndex })

	// 获取当前选中的主题值
	currentThemeValue := indexToTheme[selectedIndex]
	currentThemeName := themeNames[currentThemeValue]

	return ui.VStack(
		ui.NewTextBuilder("Settings Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("─────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Theme:").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		// Select 组件 - 当前只是显示选中项，实际选择逻辑需要进一步集成
		ui.NewSelectBuilder().
			AddOption("dark", "Dark Theme").
			AddOption("light", "Light Theme").
			AddOption("dracula", "Dracula Theme").
			AddOption("nord", "Nord Theme").
			Selected(selectedIndex).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Selected: %s", currentThemeName)).
			FgColor("green").
			Build(),
		ui.Text(""),
		// 添加手动选择按钮（演示 Store 模式）
		ui.HStack(
			ui.NewTextBuilder("Manual: ").
				FgColor("gray").
				Build(),
			ui.NewButtonBuilder("Dark").
				OnPress(SetThemeIntent{Index: 0}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Light").
				OnPress(SetThemeIntent{Index: 1}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Dracula").
				OnPress(SetThemeIntent{Index: 2}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Nord").
				OnPress(SetThemeIntent{Index: 3}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("─────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Up/Down/Enter: select | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// ============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(24),
		ui.WithTitle("Select & Table Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
