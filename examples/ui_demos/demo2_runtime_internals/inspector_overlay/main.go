// Demo 2: Runtime Internals with Framework-Level Inspector Overlay
//
// This demo demonstrates the UI Inspector as a framework-level overlay
// using the Layer system (LayerInspector).
//
// Key features:
//   - Inspector renders as an overlay layer (z-index: 4)
//   - Application remains fully interactive
//   - Press F12 or Ctrl+D to toggle inspector
//   - No interference with application layout
//
// Usage:
//   Run the program
//   Press F12 or Ctrl+D to toggle inspector overlay
//   Both app and inspector remain visible and interactive

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/style"
	ui "github.com/wwsheng009/mint/ui"
)

// Global inspector instance
var globalInspector *inspector.StandaloneInspector

// Global framework app instance
var FwApp *framework.App

// Global declarative root (to access reconciler)
var declarativeRoot *render.DeclarativeNode

func main() {
	// Initialize standalone inspector
	globalInspector = inspector.NewStandaloneInspector()
	globalInspector.Enable() // ALWAYS enable inspector (so F12 and buttons work)

	// Enable verbose inspector output from environment
	if log.InspectorLogger.Enabled() {
		fmt.Println("UI Inspector verbose mode enabled")
	}

	// Auto-show inspector from environment
	if os.Getenv("TUI_INSPECTOR") == "true" {
		globalInspector.ToggleVisibility()
		fmt.Println("UI Inspector auto-enabled - Press F12 or Ctrl+D to toggle")
	} else {
		fmt.Println("UI Inspector ready - Press [I] button or F12/Ctrl+D to toggle")
	}

	// Initialize theme
	_ = theme.SetTheme("nord")

	// Create framework app to enable F12 shortcut
	FwApp = framework.NewApp()

	// Create declarative root WITH FIBER RECONCILER
	// This enables VNodeComponentInstance for persistent event handlers
	declarativeRoot = render.NewDeclarativeNodeFromFuncWithFiber(RuntimeDemoWithInspectorOverlay, FwApp)
	declarativeRoot.SetFrameworkApp(FwApp)

	// Set as root FIRST (before registering Inspector)
	FwApp.SetRoot(declarativeRoot)

	// THEN register Inspector (after root is set, so hooks can be registered)
	FwApp.SetInspector(globalInspector)
	FwApp.SetupInspectorShortcut() // Enable F12 and Ctrl+D shortcuts

	FwApp.Resize(120, 40)

	// Initialize theme
	if err := FwApp.InitTheme("nord"); err != nil {
		log.UILogger.Debug("Failed to initialize theme: %v", err)
	}

	// Run the app
	fmt.Println("Starting Mint TUI Demo - Press F12 or Ctrl+D to toggle Inspector")

	// Debug info
	if log.UILogger.Enabled() {
		log.UILogger.Debug("[DEMO2] Inspector enabled: %v", globalInspector.IsEnabled())
		log.UILogger.Debug("[DEMO2] Inspector visible: %v", globalInspector.IsVisible())
	}

	if err := FwApp.Run(); err != nil {
		panic(err)
	}
}

// RuntimeDemoWithInspectorOverlay demonstrates framework-level inspector overlay
func RuntimeDemoWithInspectorOverlay() ui.VNode {
	currentPhase, setCurrentPhase := ui.UseStateString("idle")
	eventCount, setEventCount, _ := ui.UseStateInt(0)
	renderCount, setRenderCount, _ := ui.UseStateInt(0)
	bufferUpdates, setBufferUpdates, _ := ui.UseStateInt(0)

	// Debug: check inspector visibility before hook
	if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" {
		log.UILogger.Debug("[DEMO] globalInspector.IsVisible() = %v", globalInspector.IsVisible())
	}

	showInspector, setShowInspector := ui.UseStateBool(globalInspector.IsVisible())

	// Debug: log state
	if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" {
		log.UILogger.Debug("[DEMO] showInspector (from hook) = %v", showInspector)
	}

	// Track performance
	globalInspector.StartFrame()
	defer globalInspector.EndFrame()

	// Build main application content
	// NOTE: The reconciler will add Fiber keys to these VNodes during reconciliation
	// NOTE: Inspector is already attached in main() to declarativeRoot
	// declarativeRoot will be updated by reconciliation with Fiber keys
	appContent := buildDemoContent(
		currentPhase, eventCount, renderCount, bufferUpdates,
		setCurrentPhase, setEventCount, setRenderCount, setBufferUpdates,
		setShowInspector,
	)

	// IMPORTANT: Check inspector visibility on every render
	// This fixes the state synchronization issue between imperative Inspector API
	// and React-like declarative hooks
	inspectorVisible := globalInspector.IsVisible()

	// Debug: log visibility check
	if os.Getenv("TUI_DEBUG_INSPECTOR") == "true" {
		log.UILogger.Debug("[DEMO] Inspector visible: %v (showInspector state: %v)",
			inspectorVisible, showInspector)
	}

	// NOTE: Inspector overlay is now automatically injected by the hook system
	// The framework's PipelineRenderer will automatically wrap the app content with
	// Inspector overlay when Inspector is visible. Application code no longer needs
	// to manually handle Fragment wrapping or SetLayer() calls.
	//
	// If inspector is enabled, the hook system will:
	// 1. Call inspector.RenderContent()
	// 2. Set LayerInspector on the overlay
	// 3. Wrap in Fragment with app content
	// 4. Render using multi-layer pipeline

	// Just return app content - Inspector is handled by hooks
	return appContent
}

// buildDemoContent builds the original demo2 content
func buildDemoContent(
	currentPhase string,
	eventCount, renderCount, bufferUpdates int,
	setCurrentPhase func(string),
	setEventCount, setRenderCount, setBufferUpdates func(interface{}),
	setShowInspector func(bool),
) ui.VNode {
	return ui.VStack(
		HeaderPanel(),
		PipelineVisualization(currentPhase),
		StatisticsPanel(eventCount, renderCount, bufferUpdates),
		ControlPanel(setCurrentPhase, setEventCount, setRenderCount, setBufferUpdates, setShowInspector),
		ExplanationPanel(currentPhase),
	)
}

// HeaderPanel shows the title with border
func HeaderPanel() ui.VNode {
	headerContent := ui.HStackBuilder(
		app.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
			Style(style.FgBold(theme.Text())).
			Build(),
	).
		Gap(0).
		Align(ui.AlignCenter).
		Build()

	return ui.Bordered().
		Style(string(theme.Primary())).
		Child(headerContent).
		FillWidth().
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
func buildPipelineLine(phases []struct {
	name     string
	color    string
	position int
}, activeIndex int) ui.VNode {
	var result string
	for _, p := range phases {
		spaces := p.position - len(result)
		for i := 0; i < spaces; i++ {
			result += " "
		}
		result += "[" + p.name + "]"
	}

	return app.NewTextBuilder(result).
		Style(style.Foreground(theme.Muted())).
		Build()
}

// buildPipelineArrows creates the flow arrows
func buildPipelineArrows(phases []struct {
	name     string
	color    string
	position int
}, activeIndex int) ui.VNode {
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
	setEventCount, setRenderCount, setBufferUpdates func(interface{}),
	setShowInspector func(bool),
) ui.VNode {
	allButtons := []ui.VNode{
		app.ButtonBuilder("[1] Event").
			Key("btn-event").
			Variant(app.ButtonVariantDanger).
			OnClick(func() {
				setCurrentPhase("Event")
				setEventCount(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[2]setState").
			Key("btn-setstate").
			Variant(app.ButtonVariantSecondary).
			OnClick(func() {
				setCurrentPhase("setState")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[3]Scheduler").
			Key("btn-scheduler").
			Variant(app.ButtonVariantSuccess).
			OnClick(func() {
				setCurrentPhase("Scheduler")
				setRenderCount(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[4] Render").
			Key("btn-render").
			Variant(app.ButtonVariantPrimary).
			OnClick(func() {
				setCurrentPhase("Render")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[5]Reconcile").
			Key("btn-reconcile").
			OnClick(func() {
				setCurrentPhase("Reconcile")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[6] Layout").
			Key("btn-layout").
			OnClick(func() {
				setCurrentPhase("Layout")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[7] Paint").
			Key("btn-paint").
			OnClick(func() {
				setCurrentPhase("Paint")
				setBufferUpdates(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[0] Idle").
			Key("btn-idle").
			OnClick(func() {
				setCurrentPhase("idle")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		// Toggle Inspector button - works with framework-level overlay
		app.ButtonBuilder("[I] Inspector").
			Key("btn-inspector").
			Variant(app.ButtonVariantSecondary).
			OnClick(func() {
				// Debug output
				if os.Getenv("TUI_DEBUG") == "true" || os.Getenv("TUI_DEBUG_UI") == "true" {
					log.UILogger.Debug("[DEMO2] [I] button clicked, Inspector enabled=%v, visible=%v",
						globalInspector.IsEnabled(), globalInspector.IsVisible())
				}

				// Toggle inspector visibility
				globalInspector.ToggleVisibility()

				// Debug output after toggle
				if os.Getenv("TUI_DEBUG") == "true" || os.Getenv("TUI_DEBUG_UI") == "true" {
					log.UILogger.Debug("[DEMO2] After toggle, Inspector visible=%v",
						globalInspector.IsVisible())
				}

				// Trigger re-render to show/hide overlay
				setShowInspector(globalInspector.IsVisible())
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
	}

	wrappedButtons := app.WrapBuilder(allButtons...).
		Gap(1).
		RowGap(0).
		ScreenWidth(98).
		Align(ui.AlignCenter).
		FillWidth().
		Build()

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(wrappedButtons).
		FillWidth().
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
