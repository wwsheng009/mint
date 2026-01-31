package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// ControlledInputDemo demonstrates controlled input with real-time updates
func ControlledInputDemo() ui.VNode {
	text, setText := ui.UseStateString("")

	return ui.VStack(
		ui.NewTextBuilder("Input Demo").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Type something:").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("> "),
			ui.InputBuilder().
				Value(text).
				Placeholder("Type here...").
				MaxLength(20).
				OnChange(setText).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Length: %d/20", len(text))).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Value:").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("\"%s\"", text)).
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Type: add text | Backspace: delete | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(ControlledInputDemo,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Input Demo"),
	)
	if err != nil {
		panic(err)
	}
}
