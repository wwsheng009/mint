package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// 自定义 Intent 类型 - 用于计数器控制
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

type DecrementIntent struct{}
func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool  { return true }

func RefDemo() ui.VNode {
	// 1. 定义状态（hooks 必须在顶部）
	count, setCount, _ := ui.UseStateInt(0)

	// 2. 注册 Intent handler（闭包捕获 setter）
	ui.On(IncrementIntent{}, func() {
		setCount(func(c int) int { return c + 1 })
	})
	ui.On(DecrementIntent{}, func() {
		setCount(func(c int) int { return c - 1 })
	})

	// 3. 返回 VNode（使用 OnPress 绑定 Intent）
	return ui.VStack(
		app.NewTextBuilder("State Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		app.Text(""),
		ui.HStack(
			app.ButtonBuilder("  -  ").OnPress(DecrementIntent{}).Build(),
			app.Text("   "),
			app.ButtonBuilder("  +  ").OnPress(IncrementIntent{}).Build(),
		),
		app.Text(""),
		app.NewTextBuilder("Tab: focus | Enter: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(RefDemo,
		ui.WithWidth(40),
		ui.WithHeight(12),
		ui.WithTitle("State Demo"),
	)
	if err != nil {
		panic(err)
	}
}
