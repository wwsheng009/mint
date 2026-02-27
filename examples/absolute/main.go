package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
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
		ui.On(IncrementIntent{}, func() {
			setCount(count + 1)
		})

		return app.VStack(
			app.NewTextBuilder("Absolute Positioning Demo").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.Text("Button with notification badge:"),
			app.Text(""),
			app.HStack(
				app.ButtonBuilder("  Messages  ").
					OnPress(IncrementIntent{}).
					Variant(button.VariantPrimary).
					Build(),
				// Badge positioned absolutely relative to parent
				app.AbsoluteBuilder(
					app.NewTextBuilder("New!").
						FgColor("red").
						Bold(true).
						Build(),
				).
					Left(absolute.AbsolutePos(15)).
					Top(absolute.AbsolutePos(0)).
					Build(),
			),
			app.Text(""),
			app.NewTextBuilder("Stacked Elements").FgColor("yellow").Build(),
			app.Text(""),
			app.VStack(
				app.Text("Background layer"),
				app.HStack(
					app.Text("Middle layer"),
					app.AbsoluteBuilder(
						app.NewTextBuilder("OVERLAY").FgColor("white").BgColor("red").Build(),
					).
						Left(absolute.AbsolutePos(20)).
						Top(absolute.AbsolutePos(0)).
						ZIndex(10).
						Build(),
				),
			),
			app.Text(""),
			app.NewTextBuilder(fmt.Sprintf("Click count: %d", count)).Build(),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Absolute Demo"),
	)
}
