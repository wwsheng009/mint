package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type ScrollUpIntent struct{}
func (ScrollUpIntent) IntentType() string { return "ScrollUp" }
func (ScrollUpIntent) StayPressed() bool  { return true }

type ScrollDownIntent struct{}
func (ScrollDownIntent) IntentType() string { return "ScrollDown" }
func (ScrollDownIntent) StayPressed() bool  { return true }

func main() {
	// Generate a large list of items
	items := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = fmt.Sprintf("Item #%d - This is a long description text", i+1)
	}

	ui.Run(func() ui.VNode {
		// Track scroll position and selected item
		offset, setOffset, _ := ui.UseStateInt(0)
		selected, _, _ := ui.UseStateInt(-1)

		// 将状态保存到 GlobalState，供 handler 从 ActionContext 读取
		ctx := ui.GetCurrentContext()
		if ctx != nil {
			ctx.GlobalState["offset"] = offset
			ctx.GlobalState["setOffset"] = setOffset
		}

		// Register intent handlers using ui.On (从 ActionContext 读取状态)
		ui.On(ScrollUpIntent{}, func(actx *intent.ActionContext) {
			currentOffset := actx.GetIntState("offset", 0)
			newOffset := currentOffset - 5
			if newOffset < 0 {
				newOffset = 0
			}
			if fn, ok := actx.GetState("setOffset"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(newOffset)
				}
			}
		})
		ui.On(ScrollDownIntent{}, func(actx *intent.ActionContext) {
			currentOffset := actx.GetIntState("offset", 0)
			newOffset := currentOffset + 5
			maxOffset := len(items) - 10
			if maxOffset < 0 {
				maxOffset = 0
			}
			if newOffset > maxOffset {
				newOffset = maxOffset
			}
			if fn, ok := actx.GetState("setOffset"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(newOffset)
				}
			}
		})

		return ui.VStack(
			ui.NewTextBuilder("Virtual List Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.NewTextBuilder(fmt.Sprintf("Items: %d | Offset: %d | Selected: %d",
				len(items), offset, selected)).FgColor("gray").Build(),
			ui.Text(""),
			ui.HStack(
				ui.NewButtonBuilder(" Scroll Up ").
					OnPress(ScrollUpIntent{}).
					Build(),
				ui.NewButtonBuilder(" Scroll Down ").
					OnPress(ScrollDownIntent{}).
					Build(),
			),
			ui.Text(""),
			ui.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			ui.Text(""),
			// Virtual list - only renders visible items
			ui.NewVirtualListBuilder().
				Items(items).
				ItemHeight(1).
				VisibleCount(10).
				Height(10).
				ScrollOffset(offset).
				SelectedIndex(selected).
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Tab to buttons, Enter to scroll").FgColor("gray").Build(),
		)
	},
		ui.WithWidth(60),
		ui.WithHeight(25),
		ui.WithTitle("Virtual List Demo"),
	)
}
