package ui

import (
	"errors"
	"os"

	"github.com/wwsheng009/mint/framework"
	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/scheduler"
	"github.com/wwsheng009/mint/runtime/statemachine"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// App Entry Point - Simplified version
// =============================================================================
// For full component support, use app.Run() instead
// The app package provides component shortcuts and enhanced functionality

// Option configures the app
type Option func(*Options)

// Options holds app configuration
type Options struct {
	Width             int
	Height            int
	Title             string
	FPS               int
	EnableDevTools    bool
	NoAlternateScreen bool   // Don't use alternate screen mode - allows copying/scrolling
	InitFunc          func() // Initialization function called after Intent Runtime is created
	PluginSetupFunc   func(*framework.App)
	AIConfig          *framework.AIConfig

	// Lane Scheduler Options
	UseLaneScheduler bool   // Enable priority-based lane scheduler
	DefaultLane      uint32 // Default lane for updates (1=Sync, 2=Input, 4=Default, 8=Transition, 16=Idle)
	InteractionMode  framework.InteractionMode
}

// InteractionMode controls runtime mouse/selection behavior.
type InteractionMode = framework.InteractionMode

const (
	InteractionModeInteractive       = framework.InteractionModeInteractive
	InteractionModeAppSelection      = framework.InteractionModeAppSelection
	InteractionModeTerminalSelection = framework.InteractionModeTerminalSelection
)

// WithWidth sets the window width
func WithWidth(width int) Option {
	return func(o *Options) {
		o.Width = width
	}
}

// WithHeight sets the window height
func WithHeight(height int) Option {
	return func(o *Options) {
		o.Height = height
	}
}

// WithSize sets both width and height
func WithSize(width, height int) Option {
	return func(o *Options) {
		o.Width = width
		o.Height = height
	}
}

// WithTitle sets the window title
func WithTitle(title string) Option {
	return func(o *Options) {
		o.Title = title
	}
}

// WithFPS sets the frame rate limit
func WithFPS(fps int) Option {
	return func(o *Options) {
		o.FPS = fps
	}
}

// WithNoAlternateScreen disables alternate screen mode
// This allows:
// - Copying text from the terminal with mouse
// - Scrolling through previous output
// - Content persists after the app exits
func WithNoAlternateScreen() Option {
	return func(o *Options) {
		o.NoAlternateScreen = true
	}
}

// WithInteractionMode sets initial runtime interaction mode.
func WithInteractionMode(mode InteractionMode) Option {
	return func(o *Options) {
		o.InteractionMode = framework.InteractionMode(mode)
	}
}

// WithInit sets an initialization function that will be called after the Intent Runtime is created
// This is useful for registering Intent Handlers that depend on the Intent Runtime.
//
// Example:
//
//	ui.Run(App,
//		ui.WithInit(func() {
//			ui.RegisterIntent(func(ctx *intent.ActionContext, i MyIntent) intent.IntentResult {
//				// Handle intent
//				return intent.HandledResult()
//			})
//		}),
//	)
func WithInit(initFunc func()) Option {
	return func(o *Options) {
		o.InitFunc = initFunc
	}
}

// WithPluginSetup sets a function that will be called after the framework app is created
// This allows registering UI component plugins like modal's global handlers and middleware.
//
// Example:
//
//	import (
//	    "github.com/wwsheng009/mint/framework"
//	    "github.com/wwsheng009/mint/ui/components/modal"
//	)
//
//	ui.Run(App,
//	    ui.WithPluginSetup(func(app *framework.App) {
//	        // Register modal support
//	        app.RegisterGlobalHandler(modal.NewModalGlobalHandler())
//	        app.AddMiddleware(modal.NewModalClickMiddleware())
//	    }),
//	)
func WithPluginSetup(pluginSetup func(*framework.App)) Option {
	return func(o *Options) {
		o.PluginSetupFunc = pluginSetup
	}
}

// WithLaneScheduler enables priority-based lane scheduling for rendering.
// This allows:
//   - User input to be processed with higher priority
//   - Background tasks to be processed during idle time
//   - Interruptible rendering for large updates
//
// Example:
//
//	ui.Run(App, ui.WithLaneScheduler())
func WithLaneScheduler() Option {
	return func(o *Options) {
		o.UseLaneScheduler = true
	}
}

// WithDefaultLane sets the default lane for updates when lane scheduler is enabled.
// Lane values: 1=Sync, 2=Input, 4=Default, 8=Transition, 16=Idle
//
// Example:
//
//	// Use Input lane as default (high priority)
//	ui.Run(App, ui.WithLaneScheduler(), ui.WithDefaultLane(2))
func WithDefaultLane(lane uint32) Option {
	return func(o *Options) {
		o.DefaultLane = lane
	}
}

// appInstance holds the framework app for quit functionality
var appInstance *framework.App

// Run starts the declarative UI application
func Run(app ComponentFunc, opts ...Option) error {
	options := &Options{
		Width:           80,
		Height:          24,
		Title:           "Mint UI App",
		FPS:             60,
		InteractionMode: framework.InteractionModeInteractive,
	}

	log.UILogger.IfEnabled().Debug("ui.Run: Starting")

	for _, opt := range opts {
		opt(options)
	}

	// Set environment variable for NoAlternateScreen mode
	// The framework will check this to skip screen clearing
	if options.NoAlternateScreen {
		os.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")
	} else {
		os.Unsetenv("MINT_NO_ALTERNATE_SCREEN")
	}

	// Create the framework app
	log.UILogger.IfEnabled().Debug("ui.Run: Creating framework app")
	fwApp := framework.NewApp()
	// IMPORTANT: SetConfigSize sets the LAYOUT constraints (user's intended size)
	// Resize() only sets the BUFFER size (actual terminal size)
	fwApp.SetConfigSize(options.Width, options.Height)
	fwApp.Resize(options.Width, options.Height) // Initial terminal size
	fwApp.SetFPS(options.FPS)
	appInstance = fwApp
	if err := fwApp.SetInteractionMode(options.InteractionMode); err != nil {
		return err
	}
	if options.AIConfig != nil {
		if err := fwApp.EnableAI(*options.AIConfig); err != nil {
			return err
		}
	}

	// Initialize theme
	log.UILogger.IfEnabled().Debug("ui.Run: Initializing theme")
	if err := fwApp.InitTheme(fwtheme.DefaultThemeName); err != nil {
		log.UILogger.IfEnabled().Debug("Failed to initialize theme: %v", err)
	}

	// Initialize Intent Runtime (declarative UI layer)
	intentRuntime := intent.NewRuntime()
	intent.SetupBuiltinHandlers(intentRuntime) // Register built-in intent handlers (Increment, Decrement, etc.)
	rtui.SetGlobalIntentRuntime(intentRuntime)
	log.UILogger.IfEnabled().Debug("ui.Run: Intent Runtime initialized with built-in handlers")

	// Call plugin setup function if provided (e.g., for registering UI component extensions like Modal)
	if options.PluginSetupFunc != nil {
		options.PluginSetupFunc(fwApp)
	}

	// Call initialization function if provided (e.g., for registering Intent Handlers)
	if options.InitFunc != nil {
		options.InitFunc()
	}

	// Create declarative node from the component function with Fiber reconciler enabled
	// Fiber is now the default and required for persistent component instances and event handlers
	declarativeRoot := render.NewDeclarativeNodeFromFuncWithFiber(wrapWithDefaultPortalRoots(app))

	// Set app on the declarative node (this sets the scheduler for frame scheduling)
	declarativeRoot.SetApp(fwApp)

	// Initialize Lane Scheduler if enabled
	var fiberScheduler *rtui.FiberScheduler
	if options.UseLaneScheduler {
		log.UILogger.IfEnabled().Debug("ui.Run: Lane Scheduler enabled")
		fiberScheduler = rtui.NewFiberScheduler(
			rtui.WithOnCommit(func() {
				// Commit callback - can be used for logging/metrics
				log.UILogger.IfEnabled().Debug("ui.Run: FiberScheduler commit")
			}),
		)

		// Set default lane if specified
		if options.DefaultLane > 0 {
			defaultLane := scheduler.Lane(options.DefaultLane)
			log.UILogger.IfEnabled().Debug("ui.Run: Default lane set to %s", defaultLane)
		}

		// Store scheduler reference for global access
		rtui.SetGlobalFiberScheduler(fiberScheduler)
	}

	// CRITICAL: Sync FocusManager from DeclarativeNode to framework.App
	// This ensures keyboard navigation (Tab/Shift+Tab) works correctly
	// The Reconciler collects focusable Fibers into DeclarativeNode.focusMgr,
	// but App.focusManager is used for event routing in processMsg()
	if fm := declarativeRoot.GetFocusManager(); fm != nil {
		fwApp.SetFocusManagerFromDeclarativeNode(fm)
	}

	// Pass Intent Runtime to declarative node for component context
	render.SetDeclarativeNodeIntentRuntime(declarativeRoot, intentRuntime)

	log.UILogger.IfEnabled().Debug("ui.Run: Fiber mode enabled (default)")
	// Set as the root of the framework app
	fwApp.SetRoot(declarativeRoot)
	log.UILogger.IfEnabled().Debug("ui.Run: declarative root set to %T", declarativeRoot)
	// Run the app
	log.UILogger.IfEnabled().Debug("ui.Run: Calling fwApp.Run()")
	return fwApp.Run()
}

// Quit exits the application
func Quit() {
	if appInstance != nil {
		appInstance.Quit()
	}
}

// SetInteractionMode updates runtime interaction mode of the running app.
func SetInteractionMode(mode InteractionMode) error {
	if appInstance == nil {
		return errors.New("app is not running")
	}
	return appInstance.SetInteractionMode(framework.InteractionMode(mode))
}

// GetInteractionMode returns current interaction mode of the running app.
func GetInteractionMode() (InteractionMode, error) {
	if appInstance == nil {
		return InteractionModeInteractive, errors.New("app is not running")
	}
	return InteractionMode(appInstance.GetInteractionMode()), nil
}

// CycleInteractionMode cycles interaction mode of the running app.
func CycleInteractionMode() (InteractionMode, error) {
	if appInstance == nil {
		return InteractionModeInteractive, errors.New("app is not running")
	}
	mode, err := appInstance.CycleInteractionMode()
	return InteractionMode(mode), err
}

// RunApp starts a declarative UI application using Store + Reducer architecture.
// This is the recommended entry point for applications using AppRuntime.
//
// RunApp[T] integrates with the AppRuntime to provide:
//   - Automatic state subscription and re-rendering
//   - Type-safe state management
//   - Time-travel debugging support (via AppRuntime's history)
//
// IMPORTANT: You must register Intent handlers using ui.WithInit
//
// Example:
//
//	type AppState struct {
//		Count int
//	}
//
//	// Option 1: Direct return (less type-safe)
//	func AppView(state AppState) any {
//		return ui.VStack(
//			ui.Text(fmt.Sprintf("Count: %d", state.Count)),
//			ui.NewButtonBuilder("+").OnPress(IncrementIntent{}).Build(),
//		)
//	}
//
//	// Option 2: Type-safe wrapper (recommended)
//	func renderAppView(state AppState) ui.VNode {
//		return ui.VStack(
//			ui.Text(fmt.Sprintf("Count: %d", state.Count)),
//			ui.NewButtonBuilder("+").OnPress(IncrementIntent{}).Build(),
//		)
//	}
//	func AppView(state AppState) any { return renderAppView(state) }
//
//	var appReducerBuilder = reducer.NewBuilder[AppState]().
//	    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
//	        s.Count++
//	        return s
//	    })
//
//	func main() {
//		rt := statemachine.NewAppRuntime(AppState{}, AppView, appReducerBuilder.Build())
//		ui.RunApp(rt,
//			ui.WithWidth(60),
//			ui.WithTitle("My App"),
//			// IMPORTANT: Register Intent handlers
//			ui.WithInit(func() {
//				appReducerBuilder.RegisterToGlobal(rt.GetStore())
//			}),
//		)
//	}
func RunApp[T any](rt *statemachine.AppRuntime[T], opts ...Option) error {
	options := &Options{
		Width:           80,
		Height:          24,
		Title:           "Mint UI App",
		FPS:             60,
		InteractionMode: framework.InteractionModeInteractive,
	}

	for _, opt := range opts {
		opt(options)
	}

	log.UILogger.IfEnabled().Debug("ui.RunApp: Starting Store + Reducer app")

	// Set environment variable for NoAlternateScreen mode
	if options.NoAlternateScreen {
		os.Setenv("MINT_NO_ALTERNATE_SCREEN", "true")
	} else {
		os.Unsetenv("MINT_NO_ALTERNATE_SCREEN")
	}

	// Create the framework app
	log.UILogger.IfEnabled().Debug("ui.RunApp: Creating framework app")
	fwApp := framework.NewApp()
	fwApp.SetConfigSize(options.Width, options.Height)
	fwApp.Resize(options.Width, options.Height)
	fwApp.SetFPS(options.FPS)
	appInstance = fwApp
	if err := fwApp.SetInteractionMode(options.InteractionMode); err != nil {
		return err
	}
	if options.AIConfig != nil {
		if err := fwApp.EnableAI(*options.AIConfig); err != nil {
			return err
		}
	}

	// Initialize theme
	log.UILogger.IfEnabled().Debug("ui.RunApp: Initializing theme")
	if err := fwApp.InitTheme(fwtheme.DefaultThemeName); err != nil {
		log.UILogger.IfEnabled().Debug("Failed to initialize theme: %v", err)
	}

	// Initialize Intent Runtime
	intentRuntime := intent.NewRuntime()
	intent.SetupBuiltinHandlers(intentRuntime)
	rtui.SetGlobalIntentRuntime(intentRuntime)
	log.UILogger.IfEnabled().Debug("ui.RunApp: Intent Runtime initialized with built-in handlers")

	// Set up state change callback to trigger re-render
	// This creates the automatic integration between Store and Fiber rendering
	var declarativeRoot *render.DeclarativeNode
	rt.OnStateChange(func(state T) {
		if declarativeRoot != nil {
			log.UILogger.IfEnabled().Debug("ui.RunApp: State changed, attempting re-render")
			// Try multiple methods to trigger re-render (for compatibility)
			if r := declarativeRoot.GetReconciler(); r != nil {
				// Try type assertion for ScheduleUpdate method (internal reconciler)
				if sched, ok := r.(interface{ ScheduleUpdate(lane interface{}) }); ok {
					sched.ScheduleUpdate(rtui.LaneSyncLane)
					log.UILogger.IfEnabled().Debug("ui.RunApp: Re-render triggered via ScheduleUpdate")
					return
				}
				// Try type assertion for MarkDirty method
				if dirty, ok := r.(interface{ MarkDirty() }); ok {
					dirty.MarkDirty()
					log.UILogger.IfEnabled().Debug("ui.RunApp: Re-render triggered via MarkDirty")
					return
				}
			}
			log.UILogger.IfEnabled().Debug("ui.RunApp: Could not trigger re-render - reconciler does not support ScheduleUpdate or MarkDirty")
		}
	})

	// Call plugin setup function if provided
	if options.PluginSetupFunc != nil {
		options.PluginSetupFunc(fwApp)
	}

	// Call initialization function if provided
	if options.InitFunc != nil {
		options.InitFunc()
	}

	// Create declarative node from a component that wraps AppRuntime's view function
	// The component function captures the AppRuntime and calls View() on each render
	app := func() VNode {
		// Call AppRuntime's View function with current state
		result := rt.View()
		if vnode, ok := result.(VNode); ok {
			return vnode
		}
		// Fallback for non-VNode types (shouldn't happen with well-typed views)
		return Text("Error: View function must return ui.VNode")
	}

	// Create declarative node with Fiber reconciler
	declarativeRoot = render.NewDeclarativeNodeFromFuncWithFiber(wrapWithDefaultPortalRoots(app))

	// Set app on the declarative node
	declarativeRoot.SetApp(fwApp)

	// Initialize Lane Scheduler if enabled
	var fiberScheduler *rtui.FiberScheduler
	if options.UseLaneScheduler {
		log.UILogger.IfEnabled().Debug("ui.RunApp: Lane Scheduler enabled")
		fiberScheduler = rtui.NewFiberScheduler(
			rtui.WithOnCommit(func() {
				log.UILogger.IfEnabled().Debug("ui.RunApp: FiberScheduler commit")
			}),
		)

		if options.DefaultLane > 0 {
			defaultLane := scheduler.Lane(options.DefaultLane)
			log.UILogger.IfEnabled().Debug("ui.RunApp: Default lane set to %s", defaultLane)
		}

		rtui.SetGlobalFiberScheduler(fiberScheduler)
	}

	// Sync FocusManager from DeclarativeNode to framework.App
	if fm := declarativeRoot.GetFocusManager(); fm != nil {
		fwApp.SetFocusManagerFromDeclarativeNode(fm)
	}

	// Pass Intent Runtime to declarative node
	render.SetDeclarativeNodeIntentRuntime(declarativeRoot, intentRuntime)

	log.UILogger.IfEnabled().Debug("ui.RunApp: Fiber mode enabled (default)")
	fwApp.SetRoot(declarativeRoot)
	log.UILogger.IfEnabled().Debug("ui.RunApp: declarative root set to %T", declarativeRoot)

	// Cleanup AppRuntime when app exits
	defer rt.Close()

	// Run the app
	log.UILogger.IfEnabled().Debug("ui.RunApp: Calling fwApp.Run()")
	return fwApp.Run()
}

func wrapWithDefaultPortalRoots(app ComponentFunc) ComponentFunc {
	return func() VNode {
		content := app()
		if content == nil {
			content = Text("")
		}
		return rtui.VStack(
			newDefaultPortalRoot(rtui.DefaultOverlayPortalRootID, rtui.LayerOverlay),
			newDefaultPortalRoot(rtui.DefaultModalPortalRootID, rtui.LayerModal),
			newDefaultPortalRoot(rtui.DefaultTooltipPortalRootID, rtui.LayerTooltip),
			content,
		)
	}
}

func newDefaultPortalRoot(rootID string, layer rtui.Layer) rtui.VNode {
	return rtui.NewElement("box").SetProps(rtui.Props{
		"portalRootId": rootID,
		"_layer":       layer,
		"position":     "absolute",
		"left":         0,
		"top":          0,
		"width":        1,
		"height":       1,
	})
}
