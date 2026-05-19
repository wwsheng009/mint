package main

import (
	"github.com/wwsheng009/mint/ui"
)

// TextChangeIntent defines custom intent for input text changes
type TextChangeIntent struct{}

func (t TextChangeIntent) IntentType() string { return "TextChange" }
func (t TextChangeIntent) StayPressed() bool  { return false }

// ControlledInputDemo demonstrates controlled input with real-time updates
func ControlledInputDemo() ui.VNode {
	// Input demo - simple uncontrolled input
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
			ui.NewInputBuilder().
				Placeholder("Type here...").
				OnChange(TextChangeIntent{}).
				Build(),
		),
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
