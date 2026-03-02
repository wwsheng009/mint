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
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type SetEventPhaseIntent struct{}
func (SetEventPhaseIntent) IntentType() string { return "SetEventPhase" }
func (SetEventPhaseIntent) StayPressed() bool  { return true }

type SetSetStatePhaseIntent struct{}
func (SetSetStatePhaseIntent) IntentType() string { return "SetSetStatePhase" }
func (SetSetStatePhaseIntent) StayPressed() bool  { return true }

type SetSchedulerPhaseIntent struct{}
func (SetSchedulerPhaseIntent) IntentType() string { return "SetSchedulerPhase" }
func (SetSchedulerPhaseIntent) StayPressed() bool  { return true }

type SetRenderPhaseIntent struct{}
func (SetRenderPhaseIntent) IntentType() string { return "SetRenderPhase" }
func (SetRenderPhaseIntent) StayPressed() bool  { return true }

type SetReconcilePhaseIntent struct{}
func (SetReconcilePhaseIntent) IntentType() string { return "SetReconcilePhase" }
func (SetReconcilePhaseIntent) StayPressed() bool  { return true }

type SetLayoutPhaseIntent struct{}
func (SetLayoutPhaseIntent) IntentType() string { return "SetLayoutPhase" }
func (SetLayoutPhaseIntent) StayPressed() bool  { return true }

type SetPaintPhaseIntent struct{}
func (SetPaintPhaseIntent) IntentType() string { return "SetPaintPhase" }
func (SetPaintPhaseIntent) StayPressed() bool  { return true }

type SetIdlePhaseIntent struct{}
func (SetIdlePhaseIntent) IntentType() string { return "SetIdlePhase" }
func (SetIdlePhaseIntent) StayPressed() bool  { return true }

// Global setters for control panel
var (
	globalSetCurrentPhase       func(string)
	globalSetEventCount         func(interface{})
	globalSetRenderCount        func(interface{})
	globalSetBufferUpdates      func(interface{})
)

func main() {
	// Check if layout debug mode is enabled
	if os.Getenv("TUI_UI_DEBUG_LAYOUT") == "true" || os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
		TestLayoutInfo()
		return
	}

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

	// Update global setters
	globalSetCurrentPhase = setCurrentPhase
	globalSetEventCount = setEventCount
	globalSetRenderCount = setRenderCount
	globalSetBufferUpdates = setBufferUpdates

	// Register intent handlers
	ui.On(SetEventPhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("Event")
		}
		if globalSetEventCount != nil {
			globalSetEventCount(func(c int) int { return c + 1 })
		}
	})
	ui.On(SetSetStatePhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("setState")
		}
	})
	ui.On(SetSchedulerPhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("Scheduler")
		}
		if globalSetRenderCount != nil {
			globalSetRenderCount(func(c int) int { return c + 1 })
		}
	})
	ui.On(SetRenderPhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("Render")
		}
	})
	ui.On(SetReconcilePhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("Reconcile")
		}
	})
	ui.On(SetLayoutPhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("Layout")
		}
	})
	ui.On(SetPaintPhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("Paint")
		}
		if globalSetBufferUpdates != nil {
			globalSetBufferUpdates(func(c int) int { return c + 1 })
		}
	})
	ui.On(SetIdlePhaseIntent{}, func() {
		if globalSetCurrentPhase != nil {
			globalSetCurrentPhase("idle")
		}
	})

	return ui.VStack(
		HeaderPanel(),
		PipelineVisualization(currentPhase),
		StatisticsPanel(eventCount, renderCount, bufferUpdates),
		ControlPanel(),
		ExplanationPanel(currentPhase),
	)
}

// HeaderPanel shows the title with border
func HeaderPanel() ui.VNode {
	// Use HStackBuilder with AlignCenter for true center alignment
	headerContent := ui.HStackBuilder(
		ui.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
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
	return ui.NewTextBuilder(result).
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
	return ui.NewTextBuilder(result).
		Style(style.Foreground(theme.Muted())).
		Build()
}

// StatisticsPanel shows runtime statistics
func StatisticsPanel(eventCount, renderCount, bufferUpdates int) ui.VNode {
	content := ui.HStack(
		ui.NewTextBuilder("Events:").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("%6d", eventCount)).
			Style(style.FgBgBold(theme.Error(), theme.BG())).
			Build(),
		ui.NewTextBuilder("  Renders:").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("%6d", renderCount)).
			Style(style.FgBgBold(theme.Info(), theme.BG())).
			Build(),
		ui.NewTextBuilder("  Buffers:").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("%6d", bufferUpdates)).
			Style(style.FgBgBold(theme.Success(), theme.BG())).
			Build(),
	)

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		Build()
}

// ControlPanel provides buttons to trigger each phase
// Uses Wrap component for automatic wrapping based on screen width
func ControlPanel() ui.VNode {
	// Create all buttons as a slice
	allButtons := []ui.VNode{
		app.ButtonBuilder("[1] Event").
			Variant(app.ButtonVariantDanger).
			OnPress(SetEventPhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[2]setState").
			Variant(app.ButtonVariantSecondary).
			OnPress(SetSetStatePhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[3]Scheduler").
			Variant(app.ButtonVariantSuccess).
			OnPress(SetSchedulerPhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[4] Render").
			Variant(app.ButtonVariantPrimary).
			OnPress(SetRenderPhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[5]Reconcile").
			OnPress(SetReconcilePhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[6] Layout").
			OnPress(SetLayoutPhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[7] Paint").
			OnPress(SetPaintPhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[0] Idle").
			OnPress(SetIdlePhaseIntent{}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
	}

	// Use Wrap component for automatic wrapping
	// ScreenWidth: 78 = container width (80) - border (2)
	// This ensures buttons fill the entire available width
	wrappedButtons := app.WrapBuilder(allButtons...).
		Gap(1).                    // 1 space gap between buttons
		RowGap(0).                 // No extra gap between rows
		Width(78).                 // Container width (80) - borders (2) = 78 (ScreenWidth renamed to Width)
		Align(ui.AlignCenter).     // Center each button within its allocated space
		FillWidth().               // Stretch each row to fill width
		Build()

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(wrappedButtons).
		FillWidth(). // Stretch to full width
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

	content := ui.NewTextBuilder(text).
		Style(style.Foreground(theme.Text())).
		Build()

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		Build()
}
