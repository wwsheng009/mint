package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/absolute"
	"github.com/wwsheng009/mint/ui/components/button"
)

// IncrementIntent - 自定义 Intent 用于增加计数
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

func main() {
	ui.Run(func() ui.VNode {
		// 使用 UseState 管理计数状态
		count, setCount, _ := ui.UseStateInt(0)

		// 使用 ui.On 注册 Intent 处理器（带 sync.Map 去重，每次渲染不会重复注册）
		// 重要：使用函数式更新 setCount(func(c int) int { return c + 1 }) 避免闭包捕获旧值
		ui.On(IncrementIntent{}, func() {
			setCount(func(c int) int {
				return c + 1
			})
		})

		return ui.VStack(
			ui.NewTextBuilder("Absolute Positioning Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.Text("Button with notification badge:"),
			ui.Text(""),
			ui.HStack(
				ui.NewButtonBuilder("  Messages  ").
					OnPress(IncrementIntent{}).
					Variant(button.VariantPrimary).
					Build(),
				// Badge positioned absolutely relative to parent
				ui.NewAbsoluteBuilder(
					ui.NewTextBuilder("New!").
						FgColor("red").
						BgColor("white").
						Bold(true).
						Build(),
				).
					Left(absolute.AbsolutePos(16)).
					Top(absolute.AbsolutePos(10)).
					Build(),
			),
			ui.Text(""),
			ui.NewTextBuilder("Stacked Elements").FgColor("yellow").Build(),
			ui.Text(""),
			ui.VStack(
				ui.Text("Background layer"),
				ui.HStack(
					ui.Text("Middle layer"),
					ui.NewAbsoluteBuilder(
						ui.NewTextBuilder("OVERLAY").FgColor("white").BgColor("red").Build(),
					).
						Left(absolute.AbsolutePos(10)).
						Top(absolute.AbsolutePos(5)).
						// ZIndex(10).
						Build(),
				),
			),
			ui.Text(""),
			ui.NewTextBuilder(fmt.Sprintf("Click count: %d", count)).Build(),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(15),
		ui.WithTitle("Absolute Demo"),
	)
}
