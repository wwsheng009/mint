package app

import (
	"github.com/wwsheng009/mint/framework"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// App Entry Point
// =============================================================================

// Option configures the app
type Option func(*ui.Options)

// Options holds app configuration (re-exported from ui)
type Options = ui.Options

// WithWidth sets the window width
var WithWidth = ui.WithWidth

// WithHeight sets the window height
var WithHeight = ui.WithHeight

// WithTitle sets the window title
var WithTitle = ui.WithTitle

// WithFPS sets the frame rate limit
var WithFPS = ui.WithFPS

// appInstance holds the framework app for quit functionality
var appInstance *framework.App

// Run starts the declarative UI application
// This is the main entry point for Mint UI applications
//
// For UI components, use the ui package directly:
//   import "github.com/wwsheng009/mint/ui"
//
//   ui.Text("Hello World")
//   ui.Button("Click Me")
//   ui.Input("Placeholder")
//   ui.HBox(ui.Text("A"), ui.Text("B"))
func Run(appFunc rtui.ComponentFunc, opts ...Option) error {
	// Convert app.Option to ui.Option
	uiOpts := make([]ui.Option, len(opts))
	for i, opt := range opts {
		uiOpts[i] = ui.Option(opt)
	}

	// Use rtui.Run which handles the framework integration
	return ui.Run(appFunc, uiOpts...)
}

// Quit exits the application
func Quit() {
	if appInstance != nil {
		appInstance.Quit()
	}
}
