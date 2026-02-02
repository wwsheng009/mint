package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// MouseInteractionDemo showcases all mouse-supported components
func MouseInteractionDemo() ui.VNode {
	// Track state for various components - must be inside component
	count, setCount, _ := ui.UseStateInt(0)
	text, setText := ui.UseStateString("")
	checked1, setChecked1 := ui.UseStateBool(false)
	checked2, setChecked2 := ui.UseStateBool(false)
	selectedIndex, setSelectedIndex, _ := ui.UseStateInt(0)

	return ui.VStack(
		// Header
		app.NewTextBuilder("╔══════════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║     Mouse Interaction Demo               ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║     🖱️ Hover & Click to interact          ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("╚══════════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.Text(""),

		// Button Section
		app.NewTextBuilder("🔘 BUTTONS - Click to interact").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			app.ButtonBuilder(" [-] ").
				OnClick(func() {
					setCount(count - 1)
				}).
				Build(),
			ui.Text(" "),
			app.NewTextBuilder(fmt.Sprintf(" Count: %d ", count)).
				Bold(true).
				FgColor("green").
				Build(),
			ui.Text(" "),
			app.ButtonBuilder(" [+] ").
				OnClick(func() {
					setCount(count + 1)
				}).
				Build(),
		),
		ui.Text(""),

		// Checkbox Section
		app.NewTextBuilder("☑️ CHECKBOXES - Click to toggle").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		app.CheckboxBuilder().
			Label("Enable notifications").
			Checked(checked1).
			OnChange(setChecked1).
			Build(),
		app.CheckboxBuilder().
			Label("Accept terms and conditions").
			Checked(checked2).
			OnChange(setChecked2).
			Build(),
		ui.Text(""),

		// Input Section
		app.NewTextBuilder("📝 INPUT - Click to focus, type to edit").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Name: "),
			app.InputBuilder().
				Value(text).
				Placeholder("Hover and click here...").
				MaxLength(20).
				OnChange(setText).
				Build(),
		),
		ui.Text(""),

		// Select Section
		app.NewTextBuilder("📋 SELECT - Click to cycle options").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Theme: "),
			app.SelectBuilder().
				Options([]app.SelectOption{
					{Value: "dark", Label: "Dark"},
					{Value: "light", Label: "Light"},
					{Value: "blue", Label: "Blue"},
					{Value: "green", Label: "Green"},
				}).
				Selected(selectedIndex).
				OnChange(func(value string) {
					// Find index by value
					for i, opt := range []app.SelectOption{
						{Value: "dark", Label: "Dark"},
						{Value: "light", Label: "Light"},
						{Value: "blue", Label: "Blue"},
						{Value: "green", Label: "Green"},
					} {
						if opt.Value == value {
							setSelectedIndex(i)
							break
						}
					}
				}).
				Build(),
		),
		ui.Text(""),

		// Textarea Section
		app.NewTextBuilder("📄 TEXTAREA - Click to focus").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		app.TextareaBuilder().
			Placeholder("Hover and click to edit multi-line text...").
			Rows(3).
			Cols(40).
			Build(),
		ui.Text(""),

		// Info Section
		app.NewTextBuilder("─────────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		app.NewTextBuilder("💡 TIP: Hover over controls highlights them").
			FgColor("gray").
			Build(),
		app.NewTextBuilder("💡 TIP: Click buttons/checkboxes to interact").
			FgColor("gray").
			Build(),
		app.NewTextBuilder("💡 TIP: Use Tab to navigate, Enter to select").
			FgColor("gray").
			Build(),
	)
}

func main() {
	ui.Run(MouseInteractionDemo,
		ui.WithWidth(50),
		ui.WithHeight(28),
		ui.WithTitle("Mouse Interaction Demo"),
	)
}
