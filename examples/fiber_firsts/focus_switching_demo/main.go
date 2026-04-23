// Package main demonstrates focus switching between multiple components
// This demo shows Tab-based keyboard navigation across focusable components
//
// Architecture: Store + Reducer (Single Source of Truth)
//
// Data Flow:
//   User Input → FieldChangeIntent → Reducer → Store → View Re-render
//
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	buttonComp "github.com/wwsheng009/mint/ui/components/button"
	checkboxComp "github.com/wwsheng009/mint/ui/components/checkbox"
	inputComp "github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// Application State (Single Source of Truth)
// =============================================================================

// AppState 是应用的单一状态源，所有状态都存储在这里
type AppState struct {
	// 输入框字段
	Input1 string
	Input2 string
	Input3 string

	// Checkbox 状态（使用 string 统一存储，"true"/"false"）
	Checked1 string
	Checked2 string
	Checked3 string

	// 按钮点击计数
	ClickCount int

	// UI 状态
	ActiveTab int
}

// =============================================================================
// Intent Definitions
// =============================================================================

// ClickButtonIntent 按钮点击 intent
type ClickButtonIntent struct{}

func (ClickButtonIntent) IntentType() string { return "ClickButton" }
func (ClickButtonIntent) StayPressed() bool  { return true }

// =============================================================================
// Reducer State Machine (Pure Functions)
// =============================================================================

// reducerBuilder 定义所有状态转换逻辑
// 所有的状态变更都必须通过 Reducer，保证可预测性
var appReducer = reducer.NewBuilder[AppState]().
	// 处理按钮点击
	On(ClickButtonIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ClickCount++
		return s
	}).
	// 处理输入框的字段变更（FieldChangeIntent 由组件自动发射）
	On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
		if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
			switch fieldChange.Field {
			case "input1-value", "input2-value", "input3-value":
				// 直接更新 State，不需要类型断言
				s.Input1 = fieldChange.Value
				s.Input2 = fieldChange.Value
				s.Input3 = fieldChange.Value
				switch fieldChange.Field {
				case "input1-value":
					s.Input1 = fieldChange.Value
				case "input2-value":
					s.Input2 = fieldChange.Value
				case "input3-value":
					s.Input3 = fieldChange.Value
				}
			case "checked1", "checked2", "checked3":
				// Checkbox 的 boolean 值传递为 "true"/"false" 字符串
				switch fieldChange.Field {
				case "checked1":
					s.Checked1 = fieldChange.Value
				case "checked2":
					s.Checked2 = fieldChange.Value
				case "checked3":
					s.Checked3 = fieldChange.Value
				}
			}
		}
		return s
	})

// =============================================================================
// Global Store
// =============================================================================

// appStore 是全局单一状态源
var appStore *store.Store[AppState]

// initStore 初始化 Store
func initStore() {
	appStore = store.NewStore(AppState{
		Input1:     "",
		Input2:     "",
		Input3:     "Disabled Input",
		Checked1:   "false",
		Checked2:   "false",
		Checked3:   "false",
		ClickCount: 0,
		ActiveTab:  0,
	})
}

// registerHandlers 注册所有 Intent handlers
func registerHandlers() {
	// BuildAndRegister 会自动注册 handlers 到全局 Registry
	// 每个 handler 会：
	//   1. 运行 Reducer 计算新 State
	//   2. 更新 Store
	//   3. 调用 ScheduleUpdate() 触发重新渲染
	appReducer.RegisterToGlobal(appStore)
}

// =============================================================================
// View Layer (Pure Function)
// =============================================================================

// FocusApp 渲染应用，从 Store 读取状态
func FocusApp() ui.VNode {
	// 从 Store 获取当前状态（每次渲染时获取最新值）
	state := appStore.Get()

	return ui.VStack(
		ui.NewTextBuilder("─────────────────────────────────────────").FgColor("cyan").Build(),
		ui.NewTextBuilder("Focus Switching Demo (Store + Reducer)").Bold(true).FgColor("yellow").Build(),
		ui.Text(""),
		ui.NewTextBuilder("Use TAB to navigate between components").FgColor("gray").Build(),
		ui.NewTextBuilder("Press ENTER to activate button/checkbox").FgColor("gray").Build(),
		ui.Text(""),
		ui.NewTextBuilder("─────────────────────────────────────────").FgColor("cyan").Build(),
		ui.Text(""),

		// Button 1
		buttonComp.NewBuilder("Button 1 - First").
			OnPress(ClickButtonIntent{}).
			Build().
			SetKey("btn1"),

		// Input 1 - 显示值从 Store 读取
		ui.VStack(
			ui.NewTextBuilder("Input 1:").FgColor("blue").Build(),
			inputComp.NewBuilder().
				ForField(intent.BindField("input1-value")).
				Value(state.Input1).
				Placeholder("Enter name...").
				Width(25).
				Build().
				SetKey("input1"),
			// 显示当前输入值（调试）
			ui.NewTextBuilder(fmt.Sprintf("Value: %s", state.Input1)).
				FgColor("bright-black").
				Build(),
		),

		// Checkbox 1
		checkboxComp.NewBuilder().
			Label("Option A").
			ForField(intent.BindField("checked1")).
			Checked(state.Checked1 == "true").
			Build().
			SetKey("chk1"),

		// Button 2
		buttonComp.NewBuilder("Button 2 - Middle").
			OnPress(ClickButtonIntent{}).
			Build().
			SetKey("btn2"),

		// Input 2
		ui.VStack(
			ui.NewTextBuilder("Input 2:").FgColor("blue").Build(),
			inputComp.NewBuilder().
				ForField(intent.BindField("input2-value")).
				Value(state.Input2).
				Placeholder("Enter email...").
				Width(25).
				Build().
				SetKey("input2"),
			// 显示当前输入值（调试）
			ui.NewTextBuilder(fmt.Sprintf("Value: %s", state.Input2)).
				FgColor("bright-black").
				Build(),
		),

		// Checkbox 2
		checkboxComp.NewBuilder().
			Label("Option B").
			ForField(intent.BindField("checked2")).
			Checked(state.Checked2 == "true").
			Build().
			SetKey("chk2"),

		// Button 3 - Disabled
		buttonComp.NewBuilder("Button 3 - Last").
			Disabled(true).
			OnPress(ClickButtonIntent{}).
			Build().
			SetKey("btn3"),

		// Disabled Input 3
		ui.VStack(
			ui.NewTextBuilder("Input 3 (disabled):").FgColor("blue").Build(),
			inputComp.NewBuilder().
				ForField(intent.BindField("input3-value")).
				Value(state.Input3).
				Placeholder("Disabled Input").
				Disabled(true).
				Build().
				SetKey("input3"),
		),

		// Disabled Checkbox
		checkboxComp.NewBuilder().
			Label("Disabled Checkbox").
			Disabled(true).
			ForField(intent.BindField("checked3")).
			Checked(state.Checked3 == "true").
			Build().
			SetKey("chk3"),

		ui.Text(""),
		ui.NewTextBuilder("─────────────────────────────────────────").FgColor("cyan").Build(),

		// 显示按钮点击计数
		ui.NewTextBuilder(fmt.Sprintf("Button Click Count: %d", state.ClickCount)).
			FgColor("green").
			Bold(true).
			Build(),

		// 显示 checkbox 状态
		ui.NewTextBuilder(fmt.Sprintf("Checkbox A: %v  |  Checkbox B: %v",
			state.Checked1 == "true", state.Checked2 == "true")).
			FgColor("yellow").
			Build(),
	)
}

// =============================================================================
// Main Entry Point
// =============================================================================

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Focus Switching Demo - Store + Reducer Architecture   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("This demo showcases focus management using:")
	fmt.Println("  - Store[T]: Single source of truth for application state")
	fmt.Println("  - Reducer[T]: Pure functions for state transformations")
	fmt.Println("  - ForField() API: Automatic FieldChangeIntent emission")
	fmt.Println("")
	fmt.Println("Data Flow:")
	fmt.Println("  User Input → FieldChangeIntent → Reducer → Store → View")
	fmt.Println("")
	fmt.Println("Button:   OnPress(ClickButtonIntent{})")
	fmt.Println("Input:    ForField(intent.BindField(\"key\")) + Value(state)")
	fmt.Println("Checkbox: ForField(intent.BindField(\"key\")) + Checked(state)")
	fmt.Println("")
	fmt.Println("Using ui.Run() to start the application...")
	fmt.Println("")
	fmt.Println("Press TAB to move focus between components.")
	fmt.Println("Press ESC or CTRL+C to exit.")
	fmt.Println("")

	// 初始化 Store
	initStore()

	err := ui.Run(FocusApp,
		ui.WithWidth(60),
		ui.WithHeight(35),
		ui.WithTitle("Focus Switching Demo"),
		ui.WithInit(registerHandlers), // 注册 Reducer handlers
	)
	if err != nil {
		panic(err)
	}
}
