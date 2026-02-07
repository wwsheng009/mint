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
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// Initialize theme
	_ = theme.SetTheme("nord")

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

// HeaderPanel shows the title with border
func HeaderPanel() ui.VNode {
	// Use HStackBuilder with AlignCenter for true center alignment
	headerContent := ui.HStackBuilder(
		app.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
			Style(style.FgBold(theme.Text())).
			Build(),
	).
		Gap(0).
		Align(ui.AlignCenter).
		Build()

	// Use FillWidth() to stretch horizontally WITHOUT affecting vertical direction
	// This is the new layout system feature for single-component stretching
	return ui.Bordered().
		Style(string(theme.Primary())).
		Child(headerContent).
		FillWidth(). // Horizontal stretch ONLY (doesn't affect height)
		Build()
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

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(
			ui.VStack(
				buildPipelineLine(phases, activeIndex),
				ui.Text(""),
				buildPipelineArrows(phases, activeIndex),
			),
		).
		Build()
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
	return app.NewTextBuilder(result).
		Style(style.Foreground(theme.Muted())).
		Build()
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
		Style(style.Foreground(theme.Muted())).
		Build()
}

// StatisticsPanel shows runtime statistics
func StatisticsPanel(eventCount, renderCount, bufferUpdates int) ui.VNode {
	content := ui.HStack(
		app.NewTextBuilder("Events:").
			Style(style.Foreground(theme.Text())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("%6d", eventCount)).
			Style(style.FgBgBold(theme.Error(), theme.BG())).
			Build(),
		app.NewTextBuilder("  Renders:").
			Style(style.Foreground(theme.Text())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("%6d", renderCount)).
			Style(style.FgBgBold(theme.Info(), theme.BG())).
			Build(),
		app.NewTextBuilder("  Buffers:").
			Style(style.Foreground(theme.Text())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("%6d", bufferUpdates)).
			Style(style.FgBgBold(theme.Success(), theme.BG())).
			Build(),
	)

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		Build()
}

// ControlPanel provides buttons to trigger each phase
func ControlPanel(
	setCurrentPhase func(string),
	setEventCount func(interface{}),
	setRenderCount func(interface{}),
	setBufferUpdates func(interface{}),
) ui.VNode {
	buttonsTop := ui.HStackBuilder(
		app.ButtonBuilder("[1] Event").
			Variant(app.ButtonVariantDanger).
			OnClick(func() {
				setCurrentPhase("Event")
				setEventCount(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[2] setState").
			Variant(app.ButtonVariantSecondary).
			OnClick(func() {
				setCurrentPhase("setState")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[3] Scheduler").
			Variant(app.ButtonVariantSuccess).
			OnClick(func() {
				setCurrentPhase("Scheduler")
				setRenderCount(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[4] Render").
			Variant(app.ButtonVariantPrimary).
			OnClick(func() {
				setCurrentPhase("Render")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
	).
		Gap(1). // 使用 Gap 而不是手动空格
		Align(ui.AlignStart). // 从左到右排列，对应标题
		Build()

	buttonsBottom := ui.HStackBuilder(
		app.ButtonBuilder("[5] Reconcile").
			OnClick(func() {
				setCurrentPhase("Reconcile")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[6] Layout").
			OnClick(func() {
				setCurrentPhase("Layout")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[7] Paint").
			OnClick(func() {
				setCurrentPhase("Paint")
				setBufferUpdates(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[0] Idle").
			OnClick(func() {
				setCurrentPhase("idle")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
	).
		Gap(1). // 使用 Gap 而不是手动空格
		Align(ui.AlignStart). // 从左到右排列，对应标题
		Build()

	content := ui.VStack(
		buttonsTop,
		ui.Text(""),
		buttonsBottom,
	)

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		FillWidth(). // 让按钮面板横向拉伸
		Build()
}

// ExplanationPanel shows detailed explanation of each phase
func ExplanationPanel(currentPhase string) ui.VNode {
	explanations := map[string]string{
		"idle":      "System idle, waiting for events...",
		"Event":     "Event captured from terminal, queued for processing.",
		"setState":  "State changes queued, components marked dirty for re-render.",
		"Scheduler": "Batches dirty components, schedules rendering with time-slicing.",
		"Render":    "Component functions called to generate VNode trees from state.",
		"Reconcile": "Diff algorithm compares old/new VNode trees, computes minimal changes.",
		"Layout":    "Constraint-based layout computes position (x,y) and size (w,h).",
		"Paint":     "Nodes render to back buffer. Only dirty regions are updated.",
	}

	text := explanations[currentPhase]
	if text == "" {
		text = "Select a phase to see detailed explanation..."
	}

	content := app.NewTextBuilder(text).
		Style(style.Foreground(theme.Text())).
		Build()

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		Build()
}
