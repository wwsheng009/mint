// Package main provides sandbox demo application - updated for new architecture
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type DecrementSandboxDemoIntent struct{}
func (DecrementSandboxDemoIntent) IntentType() string { return "Decrement" }
func (DecrementSandboxDemoIntent) StayPressed() bool  { return true }

type IncrementSandboxDemoIntent struct{}
func (IncrementSandboxDemoIntent) IntentType() string { return "Increment" }
func (IncrementSandboxDemoIntent) StayPressed() bool  { return true }

// Counter 示例计数器应用
// 演示如何使用 Sandbox 进行交互式组件测试
func Counter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	name, _ := ui.UseStateString("Guest")

	// Register intent handlers for buttons
	ui.On(DecrementSandboxDemoIntent{}, func() {
		setCount(func(c int) int { return c - 1 })
	})
	ui.On(IncrementSandboxDemoIntent{}, func() {
		setCount(func(c int) int { return c + 1 })
	})

	// Style builders
	cyanStyle := style.NewStyle().Foreground(style.Color("cyan"))
	yellowBoldStyle := style.NewStyle().Foreground(style.Color("yellow")).Bold(true)
	greenBoldStyle := style.NewStyle().Foreground(style.Color("green")).Bold(true)
	grayStyle := style.NewStyle().Foreground(style.Color("bright-black"))

	// Build VNode tree using builder pattern
	return app.VStack(
		ui.TextWithStyle("╔══════════════════════════════╗", cyanStyle),
		ui.TextWithStyle("║     Sandbox Demo: Counter     ║", cyanStyle),
		ui.TextWithStyle("╚══════════════════════════════╝", cyanStyle),
		ui.Text(""),

		// Greeting
		app.HStack(
			ui.Text("Hello, "),
			ui.TextWithStyle(name, yellowBoldStyle),
			ui.Text("!"),
		),
		ui.Text(""),

		// Counter display
		ui.TextWithStyle(fmt.Sprintf("Count: %d", count), greenBoldStyle),
		ui.Text(""),

		// Buttons
		app.HStack(
			app.ButtonBuilder("  [ - ]  ").
				OnPress(DecrementSandboxDemoIntent{}).
				Build(),
			ui.Text(" "),
			app.ButtonBuilder("  [ + ]  ").
				OnPress(IncrementSandboxDemoIntent{}).
				Build(),
		),
		ui.Text(""),

		// Input field (name input - TODO: integrate with FieldChangeIntent)
		app.HStack(
			ui.Text("Name: "),
			app.InputBuilder().
				Value(name).
				Placeholder("Enter name").
				Build(),
		),
		ui.Text(""),
		ui.TextWithStyle("──────────────────────────────────", grayStyle),
		ui.Text(""),

		// Instructions
		app.HStack(
			ui.TextWithStyle("Tab: focus", grayStyle),
			ui.Text("  "),
			ui.TextWithStyle("Enter: click", grayStyle),
			ui.Text("  "),
			ui.TextWithStyle("q: quit", grayStyle),
		),
	)
}

// main 正常运行应用
func main() {
	err := ui.Run(Counter,
		ui.WithWidth(40),
		ui.WithHeight(18),
		ui.WithTitle("Sandbox Demo"),
	)
	if err != nil {
		panic(err)
	}
}

