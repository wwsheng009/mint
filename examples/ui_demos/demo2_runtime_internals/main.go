// Demo 2: Runtime Internals Visualization
//
// This demo visualizes the internal runtime scheduling flow from setState
// to terminal buffer output - the "Total Assembly Diagram" of the engine.
//
// Pipeline:
//   Event → setState → Scheduler → Render → Reconcile → Layout
//   → Layer Merge → Paint → Buffer Diff → Terminal Output
//
// Based on: framework/docs/ui/demo/demo2_inside.md

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(RuntimeDemo,
		ui.WithWidth(100),
		ui.WithHeight(35),
		ui.WithTitle("Mint TUI - Runtime Internals"),
	)
	if err != nil {
		panic(err)
	}
}

// RuntimeDemo visualizes the runtime pipeline
func RuntimeDemo() ui.VNode {
	currentPhase, setCurrentPhase := ui.UseStateString("idle")
	eventCount, setEventCount, _ := ui.UseStateInt(0)
	renderCount, setRenderCount, _ := ui.UseStateInt(0)
	bufferUpdates, setBufferUpdates, _ := ui.UseStateInt(0)

	return ui.VStack(
		HeaderPanel(),
		PipelineVisualization(currentPhase),
		StatisticsPanel(eventCount, renderCount, bufferUpdates),
		ControlPanel(setCurrentPhase, setEventCount, setRenderCount, setBufferUpdates),
		ExplanationPanel(currentPhase),
	)
}

// HeaderPanel shows the title
func HeaderPanel() ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("╔══════════════════════════════════════════════════════════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.HStack(
			app.NewTextBuilder("║ ").
				FgColor("cyan").
				Build(),
			app.NewTextBuilder("                    Runtime Scheduling Pipeline Visualization").
				Bold(true).
				FgColor("white").
				Build(),
			app.NewTextBuilder("                          ║").
				FgColor("cyan").
				Build(),
		),
		app.NewTextBuilder("╚══════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
	)
}

// PipelineVisualization shows the runtime pipeline flow
func PipelineVisualization(currentPhase string) ui.VNode {
	phases := []struct {
		name     string
		color    string
		position int
	}{
		{"Event", "red", 0},
		{"setState", "yellow", 15},
		{"Scheduler", "green", 30},
		{"Render", "blue", 45},
		{"Reconcile", "magenta", 60},
		{"Layout", "cyan", 75},
		{"Paint", "white", 90},
	}

	activeIndex := -1
	for i, p := range phases {
		if p.name == currentPhase {
			activeIndex = i
			break
		}
	}

	return ui.VStack(
		app.NewTextBuilder("┌─ Runtime Pipeline ─────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.HStack(
			app.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			buildPipelineLine(phases, activeIndex),
			app.NewTextBuilder(" │").
				FgColor("gray").
				Build(),
		),
		app.NewTextBuilder("│                                                                                              │").
			FgColor("gray").
			Build(),
		ui.HStack(
			app.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			buildPipelineArrows(phases, activeIndex),
			app.NewTextBuilder(" │").
				FgColor("gray").
				Build(),
		),
		app.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// buildPipelineLine creates the phase boxes
func buildPipelineLine(phases []struct{ name string; color string; position int }, activeIndex int) ui.VNode {
	var result string
	for _, p := range phases {
		spaces := p.position - len(result)
		for i := 0; i < spaces; i++ {
			result += " "
		}
		result += "[" + p.name + "]"
		_ = p // unused but kept for clarity
	}
	return app.NewTextBuilder(result).Build()
}

// buildPipelineArrows creates the flow arrows
func buildPipelineArrows(phases []struct{ name string; color string; position int }, activeIndex int) ui.VNode {
	var result string
	for i := range phases {
		if i > 0 {
			result += "       "
		}
		if i < len(phases)-1 {
			result += "  ↓  "
		}
	}
	return app.NewTextBuilder(result).
		FgColor("gray").
		Build()
}

// StatisticsPanel shows runtime statistics
func StatisticsPanel(eventCount, renderCount, bufferUpdates int) ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("┌─ Runtime Statistics ─────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.HStack(
			app.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			app.NewTextBuilder("Events Processed: ").
				FgColor("white").
				Build(),
			app.NewTextBuilder(fmt.Sprintf("%6d", eventCount)).
				BgColor("red").
				FgColor("white").
				Bold(true).
				Build(),
			app.NewTextBuilder("    Renders: ").
				FgColor("white").
				Build(),
			app.NewTextBuilder(fmt.Sprintf("%6d", renderCount)).
				BgColor("blue").
				FgColor("white").
				Bold(true).
				Build(),
			app.NewTextBuilder("    Buffer Updates: ").
				FgColor("white").
				Build(),
			app.NewTextBuilder(fmt.Sprintf("%6d", bufferUpdates)).
				BgColor("green").
				FgColor("white").
				Bold(true).
				Build(),
			app.NewTextBuilder(" │").
				FgColor("gray").
				Build(),
		),
		app.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// ControlPanel provides buttons to trigger each phase
func ControlPanel(
	setCurrentPhase func(string),
	setEventCount func(interface{}),
	setRenderCount func(interface{}),
	setBufferUpdates func(interface{}),
) ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("┌─ Phase Triggers ──────────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.HStack(
			app.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.VStack(
				ui.HStack(
					app.ButtonBuilder("[1] Event").
						FgColor("red").
						OnClick(func() {
							setCurrentPhase("Event")
							setEventCount(func(c int) int { return c + 1 })
						}).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[2] setState").
						FgColor("yellow").
						OnClick(func() {
							setCurrentPhase("setState")
						}).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[3] Scheduler").
						FgColor("green").
						OnClick(func() {
							setCurrentPhase("Scheduler")
							setRenderCount(func(c int) int { return c + 1 })
						}).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[4] Render").
						FgColor("blue").
						OnClick(func() {
							setCurrentPhase("Render")
						}).
						Build(),
				),
				ui.Text(""),
				ui.HStack(
					app.ButtonBuilder("[5] Reconcile").
						FgColor("magenta").
						OnClick(func() {
							setCurrentPhase("Reconcile")
						}).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[6] Layout").
						FgColor("cyan").
						OnClick(func() {
							setCurrentPhase("Layout")
						}).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[7] Paint").
						FgColor("white").
						OnClick(func() {
							setCurrentPhase("Paint")
							setBufferUpdates(func(c int) int { return c + 1 })
						}).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[0] Idle").
						FgColor("gray").
						OnClick(func() {
							setCurrentPhase("idle")
						}).
						Build(),
				),
			),
			app.NewTextBuilder(" │").
				FgColor("gray").
				Build(),
		),
		app.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// ExplanationPanel shows detailed explanation of each phase
func ExplanationPanel(currentPhase string) ui.VNode {
	explanations := map[string]string{
		"idle":      "System idle, waiting for events...",
		"Event":     "🔁 Event Phase: Keyboard/Mouse/Resize events are captured from the terminal and queued for processing.",
		"setState":  "⚡ setState: Component state changes are queued, components are marked dirty for re-rendering.",
		"Scheduler": "🗓️ Scheduler: Batches dirty components and schedules them for rendering. Implements time-slicing.",
		"Render":    "🌲 Render: Component functions are called to generate new VNode trees based on current state.",
		"Reconcile": "🔄 Reconcile: Diff algorithm compares old VNode tree with new VNode tree, computes minimal changes.",
		"Layout":    "📐 Layout: Constraint-based layout computes position (x,y) and size (w,h) for each visible node.",
		"Paint":     "🎨 Paint: Nodes render their content to the back buffer. Only dirty regions are updated.",
	}

	text := explanations[currentPhase]
	if text == "" {
		text = "Select a phase to see detailed explanation..."
	}

	return ui.VStack(
		app.NewTextBuilder("┌─ Phase Details ────────────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.HStack(
			app.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			app.NewTextBuilder(fmt.Sprintf("%-100s", text)).
				FgColor("white").
				Build(),
			app.NewTextBuilder("│").
				FgColor("gray").
				Build(),
		),
		app.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}
