// Package main provides sandbox demo application - updated for new architecture
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// ============================================================================
// AppState - 定义应用状态
// ============================================================================

type AppState struct {
	Count int // 计数器值
	Name  string // 用户名称（只读展示）
}

// ============================================================================
// Intent Types
// ============================================================================

type DecrementSandboxDemoIntent struct{}
func (DecrementSandboxDemoIntent) IntentType() string { return "Decrement" }
func (DecrementSandboxDemoIntent) StayPressed() bool  { return true }

type IncrementSandboxDemoIntent struct{}
func (IncrementSandboxDemoIntent) IntentType() string { return "Increment" }
func (IncrementSandboxDemoIntent) StayPressed() bool  { return true }

// ============================================================================
// Store 初始化
// ============================================================================

var sandboxStore = store.NewStore(AppState{
	Count: 0,
	Name:  "Guest",
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(DecrementSandboxDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			if s.Count < 0 {
				s.Count = 0
			}
			return s
		}).
		On(IncrementSandboxDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), sandboxStore)
}

// ============================================================================
// Counter - 计数器组件
// ============================================================================

// Counter 示例计数器应用
// 演示如何使用 Sandbox 进行交互式组件测试
func Counter() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(sandboxStore, func(s AppState) int { return s.Count })
	name := ui.UseStoreSelector(sandboxStore, func(s AppState) string { return s.Name })

	// Style builders
	cyanStyle := style.NewStyle().Foreground(style.Color("cyan"))
	yellowBoldStyle := style.NewStyle().Foreground(style.Color("yellow")).Bold(true)
	greenBoldStyle := style.NewStyle().Foreground(style.Color("green")).Bold(true)
	grayStyle := style.NewStyle().Foreground(style.Color("bright-black"))

	// Build VNode tree using builder pattern
	return ui.VStack(
		ui.TextWithStyle("╔══════════════════════════════╗", cyanStyle),
		ui.TextWithStyle("║     Sandbox Demo: Counter     ║", cyanStyle),
		ui.TextWithStyle("╚══════════════════════════════╝", cyanStyle),
		ui.Text(""),

		// Greeting
		ui.HStack(
			ui.Text("Hello, "),
			ui.TextWithStyle(name, yellowBoldStyle),
			ui.Text("!"),
		),
		ui.Text(""),

		// Counter display
		ui.TextWithStyle(fmt.Sprintf("Count: %d", count), greenBoldStyle),
		ui.Text(""),

		// Buttons
		ui.HStack(
			ui.NewButtonBuilder("  [ - ]  ").
				OnPress(DecrementSandboxDemoIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  [ + ]  ").
				OnPress(IncrementSandboxDemoIntent{}).
				Build(),
		),
		ui.Text(""),

		// Input field (name input - TODO: integrate with FieldChangeIntent)
		ui.HStack(
			ui.Text("Name: "),
			ui.NewInputBuilder().
				Value(name).
				Placeholder("Enter name").
				Build(),
		),
		ui.Text(""),
		ui.TextWithStyle("──────────────────────────────────", grayStyle),
		ui.Text(""),

		// Instructions
		ui.HStack(
			ui.TextWithStyle("Tab: focus", grayStyle),
			ui.Text("  "),
			ui.TextWithStyle("Enter: click", grayStyle),
			ui.Text("  "),
			ui.TextWithStyle("q: quit", grayStyle),
		),
	)
}

// ============================================================================
// Main
// ============================================================================

// main 正常运行应用
func main() {
	err := ui.Run(Counter,
		ui.WithWidth(40),
		ui.WithHeight(18),
		ui.WithTitle("Sandbox Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
