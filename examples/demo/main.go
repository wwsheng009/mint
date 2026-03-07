package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================

// AppState - 定义应用状态
// =============================================================================

type AppState struct {
	CurrentTab string  // 当前选中的 tab: "counter", "input", "tasks"
	Text       string  // 输入框文本
	Checked1   bool    // Checkbox 1 状态
	Checked2   bool    // Checkbox 2 状态
	Checked3   bool    // Checkbox 3 状态
}

// =============================================================================
// Intent Types
// =============================================================================

type SetTabIntent struct {
	Tab string
}

func (SetTabIntent) IntentType() string { return "SetTab" }
func (SetTabIntent) StayPressed() bool  { return true }

type SetDemoTextIntent struct {
	Text string
}

func (SetDemoTextIntent) IntentType() string { return "SetDemoText" }
func (SetDemoTextIntent) StayPressed() bool  { return false }

type SetDemoChecked1Intent struct {
	Checked bool
}

func (SetDemoChecked1Intent) IntentType() string { return "SetDemoChecked1" }
func (SetDemoChecked1Intent) StayPressed() bool  { return false }

type SetDemoChecked2Intent struct {
	Checked bool
}

func (SetDemoChecked2Intent) IntentType() string { return "SetDemoChecked2" }
func (SetDemoChecked2Intent) StayPressed() bool  { return false }

type SetDemoChecked3Intent struct {
	Checked bool
}

func (SetDemoChecked3Intent) IntentType() string { return "SetDemoChecked3" }
func (SetDemoChecked3Intent) StayPressed() bool  { return false }

// =============================================================================
// Store 初始化
// ============================================================================

var demoAppStore = store.NewStore(AppState{
	CurrentTab: "counter",
	Text:       "",
	Checked1:   false,
	Checked2:   false,
	Checked3:   false,
})

// =============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(SetTabIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentTab = i.(SetTabIntent).Tab
			return s
		}).
		On(SetDemoTextIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Text = i.(SetDemoTextIntent).Text
			return s
		}).
		On(SetDemoChecked1Intent{}, func(s AppState, i intent.Intent) AppState {
			s.Checked1 = i.(SetDemoChecked1Intent).Checked
			return s
		}).
		On(SetDemoChecked2Intent{}, func(s AppState, i intent.Intent) AppState {
			s.Checked2 = i.(SetDemoChecked2Intent).Checked
			return s
		}).
		On(SetDemoChecked3Intent{}, func(s AppState, i intent.Intent) AppState {
			s.Checked3 = i.(SetDemoChecked3Intent).Checked
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), demoAppStore)
}

// =============================================================================
// DemoApp - 演示所有 UI 组件
// ============================================================================

func DemoApp() ui.VNode {
	// ✅ 订阅存储的状态
	currentTab := ui.UseStoreSelector(demoAppStore, func(s AppState) string { return s.CurrentTab })
	text := ui.UseStoreSelector(demoAppStore, func(s AppState) string { return s.Text })
	checked1 := ui.UseStoreSelector(demoAppStore, func(s AppState) bool { return s.Checked1 })
	checked2 := ui.UseStoreSelector(demoAppStore, func(s AppState) bool { return s.Checked2 })
	checked3 := ui.UseStoreSelector(demoAppStore, func(s AppState) bool { return s.Checked3 })

	return ui.VStack(
		ui.NewTextBuilder("╔═══════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Mint UI Declarative Framework     ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚═══════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		// Tab navigation
		ui.HStack(
			ui.NewButtonBuilder(" [1] Counter ").
				OnPress(SetTabIntent{Tab: "counter"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" [2] Input ").
				OnPress(SetTabIntent{Tab: "input"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" [3] Tasks ").
				OnPress(SetTabIntent{Tab: "tasks"}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("───────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		// Content based on selected tab
		func() ui.VNode {
			if currentTab == "counter" {
				return ui.Fragment(
					ui.NewTextBuilder("📊 Counter Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.HStack(
						ui.NewButtonBuilder("  [ - ]  ").
							OnPress(intent.ClickIntent{}).
							Build(),
						ui.Text(" "),
						ui.NewButtonBuilder("  [ + ]  ").
							OnPress(intent.ClickIntent{}).
							Build(),
					),
				)
			} else if currentTab == "input" {
				return ui.Fragment(
					ui.NewTextBuilder("📝 Input Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.NewInputBuilder().
						Placeholder("Type here...").
						Build(),
					ui.Text(""),
					ui.NewTextBuilder("You typed: "+text).
						FgColor("magenta").
						Build(),
				)
			} else {
				return ui.Fragment(
					ui.NewTextBuilder("✓ Tasks Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.NewCheckboxBuilder().
						Label("Review documentation").
						Checked(checked1).
						Build(),
					ui.NewCheckboxBuilder().
						Label("Write tests").
						Checked(checked2).
						Build(),
					ui.NewCheckboxBuilder().
						Label("Build release").
						Checked(checked3).
						Build(),
				)
			}
		}(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Space/Enter: select | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(DemoApp,
		ui.WithWidth(50),
		ui.WithHeight(24),
		ui.WithTitle("Mint UI Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
