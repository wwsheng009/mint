package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// ControlledInputDemo demonstrates controlled input with real-time updates
func ControlledInputDemo() ui.VNode {
	text, setText := ui.UseStateString("")

	return app.VStack(
		app.NewTextBuilder("Input Demo").
			Bold(true).
			Build(),
		app.Text(""),
		app.NewTextBuilder("Type something:").
			Build(),
		app.Text(""),
		app.HStack(
			app.Text("> "),
			app.InputBuilder().
				Value(text).
				Placeholder("Type here...").
				MaxLength(20).
				OnChange(setText).
				Build(),
		),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Length: %d/20", len(text))).
			Build(),
		app.Text(""),
		app.NewTextBuilder("Value:").
			Build(),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("\"%s\"", text)).
			FgColor("cyan").
			Build(),
		app.Text(""),
		app.NewTextBuilder("Tab: focus | Type: add text | Backspace: delete | q: quit").
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
