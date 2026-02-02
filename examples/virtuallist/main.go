package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// Generate a large list of items
	items := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = fmt.Sprintf("Item #%d - This is a long description text", i+1)
	}

	ui.Run(func() ui.VNode {
		// Track scroll position and selected item
		offset, setOffset, _ := ui.UseStateInt(0)
		selected, setSelected, _ := ui.UseStateInt(-1)

		return app.VStack(
			app.NewTextBuilder("Virtual List Demo").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.NewTextBuilder(fmt.Sprintf("Items: %d | Offset: %d | Selected: %d",
				len(items), offset, selected)).FgColor("gray").Build(),
			app.Text(""),
			app.HStack(
				app.ButtonBuilder(" Scroll Up ").
					OnClick(func() {
						newOffset := offset - 5
						if newOffset < 0 {
							newOffset = 0
						}
						setOffset(newOffset)
					}).
					Build(),
				app.ButtonBuilder(" Scroll Down ").
					OnClick(func() {
						newOffset := offset + 5
						maxOffset := len(items) - 10
						if maxOffset < 0 {
							maxOffset = 0
						}
						if newOffset > maxOffset {
							newOffset = maxOffset
						}
						setOffset(newOffset)
					}).
					Build(),
			),
			app.Text(""),
			app.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			app.Text(""),
			// Virtual list - only renders visible items
			app.VirtualListBuilder().
				Items(items).
				RenderItem(func(item interface{}) ui.VNode {
					text := item.(string)
					return app.Text(text)
				}).
				ItemHeight(1).
				VisibleCount(10).
				Height(10).
				ScrollOffset(offset).
				SelectedIndex(selected).
				OnItemSelect(func(index int) {
					setSelected(index)
				}).
				PrimaryKey(func(item interface{}) string {
					return fmt.Sprintf("%v", item)
				}).
				Build(),
			app.Text(""),
			app.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			app.Text(""),
			app.NewTextBuilder("Tab to buttons, Enter to scroll").FgColor("gray").Build(),
		)
	},
		ui.WithWidth(60),
		ui.WithHeight(25),
		ui.WithTitle("Virtual List Demo"),
	)
}
