package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/ui"
)

func SimpleTest() ui.VNode {
	return ui.VStack(
		ui.Text("Button Test"),
		ui.ButtonBuilder("Button1").Build(),
		ui.ButtonBuilder("Button2").Build(),
	)
}

func main() {
	// Enable debug
	os.Setenv("TUI_DEBUG_UI", "true")

	app := ui.NewApp(
		SimpleTest,
		ui.WithWidth(30),
		ui.WithHeight(10),
		ui.WithTitle("Button Test"),
	)

	// Get the root to check button collection
	if root, ok := app.(*ui.DeclarativeRoot); ok {
		fmt.Fprintf(os.Stderr, "[DEBUG] Before run, buttons: %d\n", len(root.(*ui.declarativeRoot).buttons))
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
