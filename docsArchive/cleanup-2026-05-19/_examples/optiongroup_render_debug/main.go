// OptionGroup Render Debug - Static rendering test for OptionGroup
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
	ui_btn "github.com/wwsheng009/mint/ui/components/button"
)

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║ OptionGroup Render Debug - Form UI Testing                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	testCompleteForm()

	fmt.Println("\n=== Test Complete ===")
}

func testCompleteForm() {
	// Build the form view with initial state
	state := FormState{
		Username:  "",
		Email:     "",
		Age:       0,
		Active:    false,
		City:      "bj",      // Default: Beijing
		Interests: "dev",     // Default: Development
	}

	formNode := renderFormView(state)

	fmt.Printf("\nForm State:\n")
	fmt.Printf("  Username: %q\n", state.Username)
	fmt.Printf("  Email: %q\n", state.Email)
	fmt.Printf("  Age: %d\n", state.Age)
	fmt.Printf("  Active: %v\n", state.Active)
	fmt.Printf("  City: %q (Expected: bj = Beijing selected)\n", state.City)
	fmt.Printf("  Interests: %q (Expected: dev = Development selected)\n\n", state.Interests)

	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Println("Rendering Complete Form UI...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	renderOptionGroup(formNode, 80, 40)
}

// renderOptionGroup renders an OptionGroup and prints the buffer
func renderOptionGroup(node rtui.VNode, width, height int) {
	fwApp := framework.NewApp()
	renderNode := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return node
	})
	renderNode.SetApp(fwApp)
	renderNode.SetRenderMode(render.RenderModeFiberFirst)

	// Create buffer
	buf := paint.NewBuffer(width, height)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: width, Height: height},
		AvailableWidth:  width,
		AvailableHeight: height,
	}

	// Render
	renderNode.Paint(ctx, buf)

	// Print buffer
	printBuffer(buf, width, height)

	// Get layout boxes
	fmt.Println("\n=== Layout Boxes (All) ===")
	boxes := renderNode.GetLayoutBoxes()
	if boxes != nil {
		for _, box := range boxes {
			// Filter for OptionGroup-related boxes only
			if box.ID == "city-group" || box.ID == "interests-group" ||
				strings.Contains(box.ID, "city-group") || strings.Contains(box.ID, "interests-group") ||
				box.X <= 50 { // Include boxes in left area
				fmt.Printf("  [%s] Pos: (%d,%d), AbsPos: (%d,%d), Size: %dx%d\n",
					box.ID, box.X, box.Y, box.AbsX, box.AbsY, box.Width, box.Height)
			}
		}
	}

	// Get layout tree string
	fmt.Println("\n=== Layout Tree ===")
	if layoutTree := renderNode.GetLayoutTreeString(); layoutTree != "" {
		fmt.Println(layoutTree)
	}

	// Get paintable tree string
	fmt.Println("\n=== Paintable Tree ===")
	if paintableTree := renderNode.GetPaintableTreeString(); paintableTree != "" {
		fmt.Println(paintableTree)
	}
}

// =============================================================================
// Application State
// =============================================================================

type FormState struct {
	Username  string
	Email     string
	Age       int
	Active    bool
	City      string   // OptionGroup single-select
	Interests string   // OptionGroup multi-select (comma-separated)
}

// =============================================================================
// View Components
// =============================================================================

// renderFormView is the actual view function (from typed_intent_demo)
func renderFormView(state FormState) rtui.VNode {
	// Build form components
	usernameInput := input.NewBuilder().
		Placeholder("Type username").
		Value(state.Username).
		Width(30).
		Build()

	emailInput := input.NewBuilder().
		Placeholder("Enter email").
		Value(state.Email).
		Width(30).
		Build()

	ageInput := input.NewBuilder().
		Placeholder("Enter age").
		Value(formatInt(state.Age)).
		Width(10).
		Build()

	activeCheckbox := checkbox.NewBuilder().
		Label("Active").
		Checked(state.Active).
		Build()

	submitButton := ui_btn.NewBuilder("Submit").Build()

	resetButton := ui_btn.NewBuilder("Reset").Build()

	incAgeButton := ui_btn.NewBuilder(" + ").Build()

	decAgeButton := ui_btn.NewBuilder(" - ").Build()

	// Build the form layout
	var layout []rtui.VNode

	layout = append(layout,
		// Header
		ui.NewTextBuilder("📝 Typed Intent Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Store + Reducer Architecture").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HDivider(),
	)

	// Form fields
	layout = append(layout,
		ui.HStack(
			ui.Text("Username: "),
			ui.Text("  "),
			usernameInput,
		),

		ui.HStack(
			ui.Text("Email:    "),
			ui.Text("  "),
			emailInput,
		),

		ui.HStack(
			ui.Text("Age:      "),
			ui.Text("  "),
			ageInput,
			ui.Text(" "),
			decAgeButton,
			ui.Text(" "),
			incAgeButton,
		),

		ui.HStack(
			ui.Text("  "),
			ui.Text("  "),
			activeCheckbox,
		),

		ui.Text(""),

		// OptionGroup - Single Select (City)
		optiongroup.NewBuilder([]optiongroup.Option{
			{Value: "bj", Label: "Beijing"},
			{Value: "sh", Label: "Shanghai"},
			{Value: "gz", Label: "Guangzhou"},
			{Value: "sz", Label: "Shenzhen"},
		}).
			Key("city-group").
			Label("City:").
			Mode(optiongroup.ModeSingle).
			Selected(state.City).
			Vertical().
			Build(),

		ui.Text(""),

		// OptionGroup - Multiple Select (Interests)
		optiongroup.NewBuilder([]optiongroup.Option{
			{Value: "dev", Label: "Development"},
			{Value: "design", Label: "Design"},
			{Value: "test", Label: "Testing"},
			{Value: "pm", Label: "Product Management"},
		}).
			Key("interests-group").
			Label("Interests:").
			Mode(optiongroup.ModeMultiple).
			Selecteds(strings.Split(state.Interests, ",")).
			Vertical().
			Build(),

		ui.Text(""),

		// Action buttons
		ui.HStack(
			ui.Text("  "),
			resetButton,
			ui.Text(" "),
			submitButton,
		),
	)

	// State display
	layout = append(layout,
		ui.Text(""),
		ui.HDivider(),
		ui.NewTextBuilder("Current State:").
			FgColor("gray").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("  Username: ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(formatString(state.Username)).
				FgColor("white").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("  Email:    ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(formatString(state.Email)).
				FgColor("white").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("  Age:      ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(formatInt(state.Age)).
				FgColor("white").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("  Active:   ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(formatBool(state.Active)).
				FgColor("white").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("  City:     ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(formatString(state.City)).
				FgColor("white").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("  Interests:").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(formatString(state.Interests)).
				FgColor("white").
				Build(),
		),
	)

	layout = append(layout,
		ui.Text(""),
		ui.HDivider(),
		ui.NewTextBuilder("Type-Safe Features:").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("✓ FieldMap for automatic field updates").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("✓ Centralized logic in Reducer").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("✓ Compile-time type checking").
			FgColor("gray").
			Build(),
		ui.Text(""),
	)

	return ui.VStack(layout...)
}

// Helper functions (simplified from typed_intent_demo)
func formatInt(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

func formatString(s string) string {
	return fmt.Sprintf("%q", s)
}

func formatBool(b bool) string {
	return fmt.Sprintf("%v", b)
}

// printBuffer prints the rendered buffer content
func printBuffer(buf *paint.Buffer, w, h int) {
	fmt.Println("┌" + strings.Repeat("─", w) + "┐")
	for y := 0; y < h; y++ {
		line := "|"
		for x := 0; x < w; x++ {
			if y < len(buf.Cells) && x < len(buf.Cells[y]) {
				cell := buf.Cells[y][x]
				if len(cell.Cluster) == 0 || cell.Cluster == " " {
					line += " "
				} else {
					for _, r := range cell.Cluster {
						line += string(r)
						break
					}
				}
			} else {
				line += " "
			}
		}
		line += "|"
		fmt.Println(line)
	}
	fmt.Println("└" + strings.Repeat("─", w) + "┘")
}
