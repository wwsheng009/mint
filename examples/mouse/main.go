package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// MouseInteractionDemo showcases all mouse-supported components
func main() {
	// Track state for various components
	count, setCount, _ := ui.UseStateInt(0)
	text, setText := ui.UseStateString("")
	checked1, setChecked1 := ui.UseStateBool(false)
	checked2, setChecked2 := ui.UseStateBool(false)
	selectedIndex, setSelectedIndex, _ := ui.UseStateInt(0)

	ui.Run(func() ui.VNode {
		return ui.VStack(
			// Header
			ui.NewTextBuilder("╔══════════════════════════════════════════╗").
				FgColor("cyan").
				Build(),
			ui.NewTextBuilder("║     Mouse Interaction Demo               ║").
				FgColor("cyan").
				Build(),
			ui.NewTextBuilder("║     🖱️ Hover & Click to interact          ║").
				FgColor("cyan").
				Build(),
			ui.NewTextBuilder("╚══════════════════════════════════════════╝").
				FgColor("cyan").
				Build(),
			ui.Text(""),
			ui.Text(""),

			// Button Section
			ui.NewTextBuilder("🔘 BUTTONS - Click to interact").
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(""),
			ui.HStack(
				ui.ButtonBuilder(" [-] ").
					OnClick(func() {
						setCount(count - 1)
					}).
					Build(),
				ui.Text(" "),
				ui.NewTextBuilder(fmt.Sprintf(" Count: %d ", count)).
					Bold(true).
					FgColor("green").
					Build(),
				ui.Text(" "),
				ui.ButtonBuilder(" [+] ").
					OnClick(func() {
						setCount(count + 1)
					}).
					Build(),
			),
			ui.Text(""),

			// Checkbox Section
			ui.NewTextBuilder("☑️ CHECKBOXES - Click to toggle").
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(""),
			ui.CheckboxBuilder().
				Label("Enable notifications").
				Checked(checked1).
				OnChange(setChecked1).
				Build(),
			ui.CheckboxBuilder().
				Label("Accept terms and conditions").
				Checked(checked2).
				OnChange(setChecked2).
				Build(),
			ui.Text(""),

			// Input Section
			ui.NewTextBuilder("📝 INPUT - Click to focus, type to edit").
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(""),
			ui.HStack(
				ui.Text("Name: "),
				ui.InputBuilder().
					Value(text).
					Placeholder("Hover and click here...").
					MaxLength(20).
					OnChange(setText).
					Build(),
			),
			ui.Text(""),

			// Select Section
			ui.NewTextBuilder("📋 SELECT - Click to cycle options").
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(""),
			ui.HStack(
				ui.Text("Theme: "),
				ui.SelectBuilder().
					Options([]ui.SelectOption{
						{Value: "dark", Label: "Dark"},
						{Value: "light", Label: "Light"},
						{Value: "blue", Label: "Blue"},
						{Value: "green", Label: "Green"},
					}).
					Selected(selectedIndex).
					OnChange(func(value string) {
						// Find index by value
						for i, opt := range []ui.SelectOption{
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
			ui.NewTextBuilder("📄 TEXTAREA - Click to focus").
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(""),
			ui.TextareaBuilder().
				Placeholder("Hover and click to edit multi-line text...").
				Rows(3).
				Cols(40).
				Build(),
			ui.Text(""),

			// Info Section
			ui.NewTextBuilder("─────────────────────────────────────────").
				FgColor("bright-black").
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("💡 TIP: Hover over controls highlights them").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("💡 TIP: Click buttons/checkboxes to interact").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("💡 TIP: Use Tab to navigate, Enter to select").
				FgColor("gray").
				Build(),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(28),
		ui.WithTitle("Mouse Interaction Demo"),
	)
}
