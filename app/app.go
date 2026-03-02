package app

import (
	"github.com/wwsheng009/mint/framework"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// App Entry Point
// =============================================================================

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
func Run(appFunc rtui.ComponentFunc, opts ...func(*ui.Options)) error {
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
