package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// Counter 示例计数器应用
// 这是一个简单的计数器应用，演示如何使用 Sandbox 进行测试
func Counter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	name, setName := ui.UseStateString("Guest")

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Sandbox Demo: Counter     ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Hello, "),
			ui.NewTextBuilder(name).
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text("!"),
		),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.ButtonBuilder("  [ - ]  ").
				OnClick(func() {
					setCount(func(c int) int { return c - 1 })
				}).
				Build(),
			ui.Text(" "),
			ui.ButtonBuilder("  [ + ]  ").
				OnClick(func() {
					setCount(func(c int) int { return c + 1 })
				}).
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Name: "),
			ui.InputBuilder().
				Value(name).
				Placeholder("Enter name").
				MaxLength(15).
				OnChange(setName).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("Tab: focus").
				FgColor("bright-black").
				Build(),
			ui.Text("  "),
			ui.NewTextBuilder("Enter: click").
				FgColor("bright-black").
				Build(),
			ui.Text("  "),
			ui.NewTextBuilder("q: quit").
				FgColor("bright-black").
				Build(),
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
