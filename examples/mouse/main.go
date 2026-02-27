package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type DecrementMouseIntent struct{}
func (DecrementMouseIntent) IntentType() string { return "DecrementMouse" }
func (DecrementMouseIntent) StayPressed() bool  { return true }

type IncrementMouseIntent struct{}
func (IncrementMouseIntent) IntentType() string { return "IncrementMouse" }
func (IncrementMouseIntent) StayPressed() bool  { return true }

// MouseInteractionDemo showcases all mouse-supported components
func MouseInteractionDemo() ui.VNode {
	// Track state for various components - must be inside component
	count, setCount, _ := ui.UseStateInt(0)
	text, _ := ui.UseStateString("")
	checked1, _ := ui.UseStateBool(false)
	checked2, _ := ui.UseStateBool(false)
	selectedIndex, _, _ := ui.UseStateInt(0)

	// Register intent handlers for buttons
	ui.On(DecrementMouseIntent{}, func() {
		setCount(count - 1)
	})
	ui.On(IncrementMouseIntent{}, func() {
		setCount(count + 1)
	})

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
				OnPress(DecrementMouseIntent{}).
				Build(),
			ui.Text(" "),
			app.NewTextBuilder(fmt.Sprintf(" Count: %d ", count)).
				Bold(true).
				FgColor("green").
				Build(),
			ui.Text(" "),
			app.ButtonBuilder(" [+] ").
				OnPress(IncrementMouseIntent{}).
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
			Build(), // TODO: integrate with FieldChangeIntent
		app.CheckboxBuilder().
			Label("Accept terms and conditions").
			Checked(checked2).
			Build(), // TODO: integrate with FieldChangeIntent
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
				Build(), // TODO: integrate with FieldChangeIntent
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
				Build(), // TODO: integrate with FieldChangeIntent
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
