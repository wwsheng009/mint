package ui

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
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
	Width          int
	Height         int
	Title          string
	FPS            int
	EnableDevTools bool
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

	for _, opt := range opts {
		opt(options)
	}

	// Create the framework app
	fwApp := framework.NewApp()
	fwApp.Resize(options.Width, options.Height)
	appInstance = fwApp

	// Initialize theme
	if err := fwApp.InitTheme("dark"); err != nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "Failed to initialize theme: %v\n", err)
		}
	}

	// Create declarative node from the component function
	declarativeRoot := render.NewDeclarativeNodeFromFunc(app)

	// Set as the root of the framework app
	fwApp.SetRoot(declarativeRoot)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "ui.Run: declarative root set to %T\n", declarativeRoot)
	}

	// Run the app
	return fwApp.Run()
}

// Quit exits the application
func Quit() {
	if appInstance != nil {
		appInstance.Quit()
	}
}
