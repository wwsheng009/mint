// Demo 4: Complex Layout System
//
// This demo demonstrates the complex layout capabilities:
// - Flex (线性分配)
// - Grid (二维分区)
// - Absolute (覆盖层)
// - Scroll containers
// - Constraint propagation (约束传播)
//
// Based on: framework/docs/ui/demo/demo4_layout.md

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type SetComplexLayoutTabIntent struct {
	TabID string
}
func (SetComplexLayoutTabIntent) IntentType() string { return "SetComplexLayoutTab" }
func (SetComplexLayoutTabIntent) StayPressed() bool  { return true }

// Global setter for tab navigation
var globalSetComplexLayoutTab func(string)

func main() {
	err := ui.Run(LayoutDemo,
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Mint TUI - Complex Layout"),
	)
	if err != nil {
		panic(err)
	}
}

// LayoutDemo is the root component
func LayoutDemo() ui.VNode {
	currentDemo, setCurrentDemo := ui.UseStateString("flex")

	// Update global setter
	globalSetComplexLayoutTab = setCurrentDemo

	// Register tab change handler
	ui.On(SetComplexLayoutTabIntent{TabID: currentDemo}, func() {
		if globalSetComplexLayoutTab != nil {
			globalSetComplexLayoutTab(currentDemo)
		}
	})

	return ui.VStack(
		HeaderPanel(),
		TabNavigation(currentDemo),
		ui.Text(""),
		renderDemoContent(currentDemo),
	)
}

// HeaderPanel shows the title
func HeaderPanel() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════════════════════════════════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("cyan").
				Build(),
			ui.NewTextBuilder("                          Complex Layout System Demo").
				Bold(true).
				FgColor("white").
				Build(),
			ui.NewTextBuilder("                                   ║").
				FgColor("cyan").
				Build(),
		),
		ui.NewTextBuilder("╚══════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
	)
}

// TabNavigation provides tab buttons
func TabNavigation(currentDemo string) ui.VNode {
	tabs := []struct {
		id    string
		label string
	}{
		{"flex", "Flex"},
		{"grid", "Grid"},
		{"absolute", "Absolute"},
		{"scroll", "Scroll"},
		{"complex", "Complex"},
	}

	var children []ui.VNode
	for _, tab := range tabs {
		isActive := currentDemo == tab.id
		var btn ui.VNode
		if isActive {
			btn = ui.NewButtonBuilder("[" + tab.label + "]").
				BgColor("blue").
				FgColor("white").
				OnPress(SetComplexLayoutTabIntent{TabID: tab.id}).
				Build()
		} else {
			btn = ui.NewButtonBuilder(" " + tab.label + " ").
				FgColor("blue").
				OnPress(SetComplexLayoutTabIntent{TabID: tab.id}).
				Build()
		}
		children = append(children, btn, ui.Text(" "))
	}

	return ui.HStack(children...)
}

// renderDemoContent renders the selected demo content
func renderDemoContent(currentDemo string) ui.VNode {
	switch currentDemo {
	case "flex":
		return FlexDemo()
	case "grid":
		return GridDemo()
	case "absolute":
		return AbsoluteDemo()
	case "scroll":
		return ScrollDemo()
	case "complex":
		return ComplexLayoutDemo()
	default:
		return ui.NewTextBuilder("Unknown demo").Build()
	}
}

// FlexDemo demonstrates Flex layout
func FlexDemo() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("┌─ Flex Layout (Row) ─────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("[Fixed]").
				BgColor("red").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder("[Flex=1]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder("[Flex=2]").
				BgColor("green").
				FgColor("white").
				Build(),
			ui.NewTextBuilder("                                               │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("│ Fixed width elements take their space, remaining is divided by Flex ratio                │").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("┌─ Flex Layout (Column) ────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.VStack(
			ui.NewTextBuilder("│ [Fixed Height=3]").
				FgColor("gray").
				BgColor("red").
				FgColor("white").
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("│ [Flex=1 - takes remaining space]").
				FgColor("gray").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("│ [Fixed Height=2]").
				FgColor("gray").
				BgColor("green").
				FgColor("white").
				Build(),
		),
		ui.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// GridDemo demonstrates Grid layout
func GridDemo() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("╔═════════════════ Grid Layout Demo ══════════════════════════════════════════════════════════════╗").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.VStack(
				ui.HStack(
					ui.NewTextBuilder("┌───────┐").
						FgColor("yellow").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("┌───────┐").
						FgColor("cyan").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("┌───────┐").
						FgColor("green").
						Build(),
				),
				ui.HStack(
					ui.NewTextBuilder("│ CPU   │").
						FgColor("yellow").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("│ RAM   │").
						FgColor("cyan").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("│ NET   │").
						FgColor("green").
						Build(),
				),
				ui.HStack(
					ui.NewTextBuilder("│  32%  │").
						FgColor("yellow").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("│  68%  │").
						FgColor("cyan").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("│ 120MB │").
						FgColor("green").
						Build(),
				),
				ui.HStack(
					ui.NewTextBuilder("└───────┘").
						FgColor("yellow").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("└───────┘").
						FgColor("cyan").
						Build(),
					ui.Text("  "),
					ui.NewTextBuilder("└───────┘").
						FgColor("green").
						Build(),
				),
			),
			ui.NewTextBuilder("                                                                      ║").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("Grid allows 2D positioning with row/column spans. Ideal for dashboard layouts.    ").
				FgColor("white").
				Build(),
			ui.NewTextBuilder("║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("╚════════════════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("gray").
			Build(),
	)
}

// AbsoluteDemo demonstrates Absolute positioning
func AbsoluteDemo() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("╔═══════════════════ Absolute Positioning Demo ════════════════════════════════════════════════════╗").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.Text("Base Layer Content "),
			ui.Text("                  "),
			ui.NewTextBuilder("TOP").
				BgColor("red").
				FgColor("white").
				Bold(true).
				Build(),
			ui.NewTextBuilder("                                          ║").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.Text("                  "),
			ui.NewTextBuilder("[Overlay]").
				BgColor("yellow").
				FgColor("black").
				Build(),
			ui.Text("                                               "),
			ui.NewTextBuilder("RIGHT").
				BgColor("blue").
				FgColor("white").
				Bold(true).
				Build(),
			ui.NewTextBuilder("    ║").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("Absolute positioning:").
				Bold(true).
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder("Top").
				BgColor("magenta").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder("Bottom").
				BgColor("magenta").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder("Left").
				BgColor("magenta").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder("Right").
				BgColor("magenta").
				FgColor("white").
				Build(),
			ui.NewTextBuilder("                     ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("╚════════════════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("gray").
			Build(),
	)
}

// ScrollDemo demonstrates Scroll containers
func ScrollDemo() ui.VNode {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("Item #%02d", i+1)
	}

	return ui.VStack(
		ui.NewTextBuilder("╔═══════════════════ Scroll Container Demo ══════════════════════════════════════════════════════╗").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("Scrollable List (8 visible of 20 total):").
				FgColor("cyan").
				Bold(true).
				Build(),
			ui.NewTextBuilder("                                    ║").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.VStack(
				ui.NewTextBuilder("┌─ Scroll Area ─────────────────────┐").
					FgColor("gray").
					Build(),
				ui.HStack(
					ui.NewTextBuilder("│ ").
						FgColor("gray").
						Build(),
					ui.VStack(renderScrollItems(items, 0, 8)...),
					ui.NewTextBuilder(" │").
						FgColor("gray").
						Build(),
				),
				ui.NewTextBuilder("└────────────────────────────────────┘").
					FgColor("gray").
					Build(),
			),
			ui.NewTextBuilder("                                      ║").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.Text("Use arrow keys or Page Up/Down to scroll"),
			ui.NewTextBuilder("                                    ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("╚════════════════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("gray").
			Build(),
	)
}

// renderScrollItems renders scrollable items
func renderScrollItems(items []string, offset int, visibleCount int) []ui.VNode {
	end := offset + visibleCount
	if end > len(items) {
		end = len(items)
	}

	var result []ui.VNode
	for i := offset; i < end; i++ {
		result = append(result,
			ui.NewTextBuilder(items[i]).
				FgColor("white").
				Build(),
		)
	}
	return result
}

// ComplexLayoutDemo demonstrates a complex IDE-like layout
func ComplexLayoutDemo() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════════════════════════════════════════════════════════════════════════╗").
			FgColor("gray").
			Build(),
		// Header row (fixed height)
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("[File] [Edit] [View] [Run]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.NewTextBuilder("                                                                  ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("╠══════════════════════════════════════════════════════════════════════════════════════════════════╣").
			FgColor("gray").
			Build(),
		// Main content row (flex height)
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			// Sidebar (fixed width)
			ui.VStack(
				ui.NewTextBuilder("┌─ Explorer ─┐").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│ > src      │").
					FgColor("yellow").
					Build(),
				ui.NewTextBuilder("│   > ui     │").
					FgColor("yellow").
					Build(),
				ui.NewTextBuilder("│   > core   │").
					FgColor("white").
					Build(),
				ui.NewTextBuilder("│ > pkg      │").
					FgColor("yellow").
					Build(),
				ui.NewTextBuilder("│   main.go  │").
					FgColor("white").
					Build(),
				ui.NewTextBuilder("└────────────┘").
					FgColor("gray").
					Build(),
			),
			ui.Text(" "),
			// Main content area (flex width)
			ui.VStack(
				// Tab bar (fixed height)
				ui.NewTextBuilder("┌─ main.go ────────────────────────────────────────┐").
					FgColor("gray").
					Build(),
				// Editor area (flex height)
				ui.NewTextBuilder("│ func main() {                                     │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│     ui.Run(App)                                   │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│ }                                                 │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│                                                   │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│ (scroll)                                          │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("└───────────────────────────────────────────────────┘").
					FgColor("gray").
					Build(),
				// Console (fixed height)
				ui.NewTextBuilder("┌─ Console ─────────────────────────────────────────┐").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│ > Building...                                    │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│ > Done in 1.2s                                   │").
					FgColor("green").
					Build(),
				ui.NewTextBuilder("└───────────────────────────────────────────────────┘").
					FgColor("gray").
					Build(),
			),
			ui.Text(" "),
			// Problems panel (fixed width)
			ui.VStack(
				ui.NewTextBuilder("┌─ Problems ─┐").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│            │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("│ 0 errors  │").
					FgColor("green").
					Build(),
				ui.NewTextBuilder("│ 0 warnings│").
					FgColor("yellow").
					Build(),
				ui.NewTextBuilder("│            │").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("└────────────┘").
					FgColor("gray").
					Build(),
			),
			ui.NewTextBuilder("  ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("╠══════════════════════════════════════════════════════════════════════════════════════════════════╣").
			FgColor("gray").
			Build(),
		// Status bar (fixed height)
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("Ln 1, Col 1").
				FgColor("white").
				Build(),
			ui.NewTextBuilder("  ").
				Build(),
			ui.NewTextBuilder("UTF-8").
				FgColor("green").
				Build(),
			ui.NewTextBuilder("  ").
				Build(),
			ui.NewTextBuilder("Go").
				BgColor("cyan").
				FgColor("black").
				Build(),
			ui.NewTextBuilder("                                                               ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("╚══════════════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("gray").
			Build(),
	)
}
