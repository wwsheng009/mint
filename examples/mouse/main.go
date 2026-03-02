package main

import (
	"fmt"

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
			ui.NewButtonBuilder(" [-] ").
				OnPress(DecrementMouseIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder(fmt.Sprintf(" Count: %d ", count)).
				Bold(true).
				FgColor("green").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" [+] ").
				OnPress(IncrementMouseIntent{}).
				Build(),
		),
		ui.Text(""),

		// Checkbox Section
		ui.NewTextBuilder("☑️ CHECKBOXES - Click to toggle").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewCheckboxBuilder().
			Label("Enable notifications").
			Checked(checked1).
			Build(), // TODO: integrate with FieldChangeIntent
		ui.NewCheckboxBuilder().
			Label("Accept terms and conditions").
			Checked(checked2).
			Build(), // TODO: integrate with FieldChangeIntent
		ui.Text(""),

		// Input Section
		ui.NewTextBuilder("📝 INPUT - Click to focus, type to edit").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Name: "),
			ui.NewInputBuilder().
				Value(text).
				Placeholder("Hover and click here...").
				Build(), // TODO: integrate with FieldChangeIntent
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
			ui.NewSelectBuilder().
				Options([]ui.NewSelectOption{
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
		ui.NewTextBuilder("📄 TEXTAREA - Click to focus").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextareaBuilder().
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
}

func main() {
	ui.Run(MouseInteractionDemo,
		ui.WithWidth(50),
		ui.WithHeight(28),
		ui.WithTitle("Mouse Interaction Demo"),
	)
}
