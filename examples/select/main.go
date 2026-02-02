package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// SelectDemo demonstrates the select dropdown component
func SelectDemo() ui.VNode {
	selectedTheme, setSelectedTheme := ui.UseStateString("dark")

	// Map values to display names
	themes := map[string]string{
		"dark":    "Dark Theme",
		"light":   "Light Theme",
		"dracula": "Dracula Theme",
		"nord":    "Nord Theme",
	}
	// Map values to indices
	themeToIndex := map[string]int{
		"dark":    0,
		"light":   1,
		"dracula": 2,
		"nord":    3,
	}

	return ui.VStack(
		app.NewTextBuilder("Settings Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		app.Text(""),
		app.NewTextBuilder("─────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		app.Text(""),
		app.NewTextBuilder("Theme:").
			FgColor("bright-black").
			Build(),
		app.Text(""),
		app.SelectBuilder().
			AddOption("dark", "Dark Theme").
			AddOption("light", "Light Theme").
			AddOption("dracula", "Dracula Theme").
			AddOption("nord", "Nord Theme").
			Selected(themeToIndex[selectedTheme]).
			OnChange(setSelectedTheme).
			Build(),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Selected: %s", themes[selectedTheme])).
			FgColor("green").
			Build(),
		app.Text(""),
		app.TableBuilder().
			Columns([]app.TableColumn{
				{Title: "ID", Width: 5},
				{Title: "Name", Width: 12},
				{Title: "Status", Width: 10},
			}).
			AddRow("1", "Alice", "Active").
			AddRow("2", "Bob", "Active").
			AddRow("3", "Charlie", "Inactive").
			Build(),
		app.Text(""),
		app.NewTextBuilder("Tab: focus | Up/Down/Enter: select | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(22),
		ui.WithTitle("Select & Table Demo"),
	)
	if err != nil {
		panic(err)
	}
}
