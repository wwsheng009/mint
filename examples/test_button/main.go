package main

import (
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func SimpleTest() ui.VNode {
	return ui.VStack(
		ui.Text("Button Test"),
		app.ButtonBuilder("Button1").Build(),
		app.ButtonBuilder("Button2").Build(),
	)
}

func main() {
	// Enable debug
	os.Setenv("TUI_DEBUG_UI", "true")

	err := ui.Run(SimpleTest,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Button Test"),
	)
	if err != nil {
		panic(err)
	}
}
