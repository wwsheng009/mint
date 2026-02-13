package ui

import (
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/render"
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
	NoAlternateScreen bool // Don't use alternate screen mode - allows copying/scrolling
}

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

// appInstance holds the framework app for quit functionality
var appInstance *framework.App

// Run starts the declarative UI application
func Run(app ComponentFunc, opts ...Option) error {
	options := &Options{
		Width:  80,
		Height: 24,
		Title:  "Mint UI App",
		FPS:    60,
	}

	if log.UILogger.Enabled() {
		log.UILogger.Debug("ui.Run: Starting")
	}

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
	if log.UILogger.Enabled() {
		log.UILogger.Debug("ui.Run: Creating framework app")
	}
	fwApp := framework.NewApp()
	// IMPORTANT: SetConfigSize sets the LAYOUT constraints (user's intended size)
	// Resize() only sets the BUFFER size (actual terminal size)
	fwApp.SetConfigSize(options.Width, options.Height)
	fwApp.Resize(options.Width, options.Height) // Initial terminal size
	appInstance = fwApp

	// Initialize theme
	if log.UILogger.Enabled() {
		log.UILogger.Debug("ui.Run: Initializing theme")
	}
	if err := fwApp.InitTheme("dark"); err != nil {
		if log.UILogger.Enabled() {
			log.UILogger.Debug("Failed to initialize theme: %v", err)
		}
	}

	// Create declarative node from the component function with Fiber reconciler enabled
	// Fiber is now the default and required for persistent component instances and event handlers
	declarativeRoot := render.NewDeclarativeNodeFromFuncWithFiber(app, fwApp)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("ui.Run: Fiber mode enabled (default)")
	}

	// Set as the root of the framework app
	fwApp.SetRoot(declarativeRoot)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("ui.Run: declarative root set to %T", declarativeRoot)
	}

	// Run the app
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("ui.Run: Calling fwApp.Run()")
	}
	return fwApp.Run()
}

// Quit exits the application
func Quit() {
	if appInstance != nil {
		appInstance.Quit()
	}
}
