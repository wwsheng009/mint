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
//
// Architecture: Store + Reducer + Custom Intent (Single Source of Truth)

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	ui "github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the demo state.
type AppState struct {
	CurrentPhase  string
	EventCount    int
	RenderCount   int
	BufferUpdates int
	ShowInspector bool
}

// =============================================================================
// Custom Intent Types
// =============================================================================

// SetPhaseIntent sets the current pipeline phase.
type SetPhaseIntent struct {
	Phase string
}

func (SetPhaseIntent) IntentType() string { return "SetPhase" }
func (SetPhaseIntent) StayPressed() bool  { return true }

// SetPhaseWithRenderCountIntent sets phase and increments render count.
type SetPhaseWithRenderCountIntent struct {
	Phase string
}

func (SetPhaseWithRenderCountIntent) IntentType() string { return "SetPhaseWithRenderCount" }
func (SetPhaseWithRenderCountIntent) StayPressed() bool  { return true }

// SetPhaseWithBufferCountIntent sets phase and increments buffer count.
type SetPhaseWithBufferCountIntent struct {
	Phase string
}

func (SetPhaseWithBufferCountIntent) IntentType() string { return "SetPhaseWithBufferCount" }
func (SetPhaseWithBufferCountIntent) StayPressed() bool  { return true }

// SetPhaseWithEventCountIntent sets phase and increments event count.
type SetPhaseWithEventCountIntent struct {
	Phase string
}

func (SetPhaseWithEventCountIntent) IntentType() string { return "SetPhaseWithEventCount" }
func (SetPhaseWithEventCountIntent) StayPressed() bool  { return true }

// ToggleInspectorIntent toggles the inspector visibility.
type ToggleInspectorIntent struct{}

func (ToggleInspectorIntent) IntentType() string { return "ToggleInspector" }
func (ToggleInspectorIntent) StayPressed() bool  { return true }

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer.
func init() {
	// Handle SetPhaseIntent
	appReducer.On(SetPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
		spi := i.(SetPhaseIntent)
		s.CurrentPhase = spi.Phase
		return s
	})

	// Handle SetPhaseWithEventCountIntent
	appReducer.On(SetPhaseWithEventCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithEventCountIntent)
		s.CurrentPhase = pi.Phase
		s.EventCount++
		return s
	})

	// Handle SetPhaseWithRenderCountIntent
	appReducer.On(SetPhaseWithRenderCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithRenderCountIntent)
		s.CurrentPhase = pi.Phase
		s.RenderCount++
		return s
	})

	// Handle SetPhaseWithBufferCountIntent
	appReducer.On(SetPhaseWithBufferCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithBufferCountIntent)
		s.CurrentPhase = pi.Phase
		s.BufferUpdates++
		return s
	})

	// Handle ToggleInspectorIntent
	appReducer.On(ToggleInspectorIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ShowInspector = !s.ShowInspector
		// Toggle global inspector visibility
		globalInspector.ToggleVisibility()
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the demo state.
var appStore = store.NewStore(AppState{
	CurrentPhase:  "idle",
	EventCount:    0,
	RenderCount:   0,
	BufferUpdates: 0,
	ShowInspector: false,
})

// =============================================================================
// Global instances
// =============================================================================

var globalInspector *inspector.StandaloneInspector
var FwApp *framework.App
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
		// Update store state
		appStore.Update(func(s AppState) AppState {
			s.ShowInspector = true
			return s
		})
		fmt.Println("UI Inspector auto-enabled - Press F12 or Ctrl+D to toggle")
	} else {
		fmt.Println("UI Inspector ready - Press [I] button or F12/Ctrl+D to toggle")
	}

	// Initialize theme
	_ = theme.SetTheme("nord")

	// Create framework app to enable F12 shortcut
	FwApp = framework.NewApp()

	// Create declarative root WITH FIBER RECONCILER
	declarativeRoot = render.NewDeclarativeNodeFromFuncWithFiber(RuntimeDemoWithInspectorOverlay)
	declarativeRoot.SetApp(FwApp)

	// Set as root FIRST (before registering Inspector)
	FwApp.SetRoot(declarativeRoot)

	// THEN register Inspector (after root is set, so hooks can be registered)
	FwApp.SetInspector(globalInspector)
	FwApp.SetupInspectorShortcut() // Enable F12 and Ctrl+D shortcuts

	FwApp.Resize(120, 40)

	// Initialize theme
	if err := FwApp.InitTheme("nord"); err != nil {
		log.UILogger.IfEnabled().Debug("Failed to initialize theme: %v", err)
	}

	// Register reducer handlers to store
	appReducer.RegisterToGlobal(appStore)

	// Run the app
	fmt.Println("Starting Mint TUI Demo - Press F12 or Ctrl+D to toggle Inspector")

	// Debug info
	log.UILogger.IfEnabled().Debug("[DEMO2] Inspector enabled: %v", globalInspector.IsEnabled())
	log.UILogger.IfEnabled().Debug("[DEMO2] Inspector visible: %v", globalInspector.IsVisible())

	if err := FwApp.Run(); err != nil {
		panic(err)
	}
}

// RuntimeDemoWithInspectorOverlay demonstrates framework-level inspector overlay
func RuntimeDemoWithInspectorOverlay() ui.VNode {
	// Get current state snapshot from Store
	state := appStore.Get()

	// Track performance
	globalInspector.StartFrame()
	defer globalInspector.EndFrame()

	// Build main application content
	appContent := buildDemoContent(state)

	// Just return app content - Inspector is handled by framework hooks
	return appContent
}

// buildDemoContent builds the original demo2 content
func buildDemoContent(state AppState) ui.VNode {
	return ui.VStack(
		HeaderPanel(),
		PipelineVisualization(state.CurrentPhase),
		StatisticsPanel(state.EventCount, state.RenderCount, state.BufferUpdates),
		ControlPanel(),
		ExplanationPanel(state.CurrentPhase),
	)
}

// HeaderPanel shows the title with border
func HeaderPanel() ui.VNode {
	headerContent := ui.HStackBuilder(
		ui.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
			Style(style.FgBold(theme.Text())).
			Build(),
	).
		Gap(0).
		Align(ui.AlignCenter).
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{headerContent})
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

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{
			ui.VStack(
				buildPipelineLine(phases, activeIndex),
				ui.Text(""),
				buildPipelineArrows(phases, activeIndex),
			),
		})
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

	return ui.NewTextBuilder(result).
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

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{content})
}

// ControlPanel provides buttons to trigger each phase
func ControlPanel() ui.VNode {
	allButtons := []ui.VNode{
		ui.NewButtonBuilder("[1] Event").
			Key("btn-event").
			Variant(ui.ButtonVariantDanger).
			OnPress(SetPhaseWithEventCountIntent{Phase: "Event"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[2]setState").
			Key("btn-setstate").
			Variant(ui.ButtonVariantSecondary).
			OnPress(SetPhaseIntent{Phase: "setState"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[3]Scheduler").
			Key("btn-scheduler").
			Variant(ui.ButtonVariantSuccess).
			OnPress(SetPhaseWithRenderCountIntent{Phase: "Scheduler"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[4] Render").
			Key("btn-render").
			Variant(ui.ButtonVariantPrimary).
			OnPress(SetPhaseIntent{Phase: "Render"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[5]Reconcile").
			Key("btn-reconcile").
			OnPress(SetPhaseIntent{Phase: "Reconcile"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[6] Layout").
			Key("btn-layout").
			OnPress(SetPhaseIntent{Phase: "Layout"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[7] Paint").
			Key("btn-paint").
			OnPress(SetPhaseWithBufferCountIntent{Phase: "Paint"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[0] Idle").
			Key("btn-idle").
			OnPress(SetPhaseIntent{Phase: "idle"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		// Toggle Inspector button - works with framework-level overlay
		ui.NewButtonBuilder("[I] Inspector").
			Key("btn-inspector").
			Variant(ui.ButtonVariantSecondary).
			OnPress(ToggleInspectorIntent{}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
	}

	wrappedButtons := ui.NewWrapBuilder().Children(allButtons...).
		Gap(1).
		RowGap(0).
		Width(98).
		Align(ui.AlignCenter).
		FillWidth().
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{wrappedButtons})
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

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{content})
}
