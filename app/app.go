package app

import (
	"github.com/wwsheng009/mint/components/basic"
	"github.com/wwsheng009/mint/components/button"
	"github.com/wwsheng009/mint/components/data"
	"github.com/wwsheng009/mint/components/feedback"
	"github.com/wwsheng009/mint/components/form"
	"github.com/wwsheng009/mint/components/layout"
	"github.com/wwsheng009/mint/components/navigation"
	"github.com/wwsheng009/mint/components/overlay"
	"github.com/wwsheng009/mint/framework"
	ui "github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Component Shortcuts - Re-export for convenience
// =============================================================================

// Basic components
var (
	Text            = basic.Text
	NewText         = basic.NewText
	NewTextBuilder  = basic.NewTextBuilder
	Divider         = basic.Divider
	NewDivider      = basic.NewDivider
	DividerBuilder  = basic.DividerBuilder
)

// DividerStyle type and constants
type DividerStyle = basic.DividerStyle

const (
	DividerSolid   = basic.DividerSolid
	DividerDashed  = basic.DividerDashed
	DividerDotted  = basic.DividerDotted
)

// Button components
var (
	Button       = button.Button
	NewButton    = button.NewButton
	ButtonBuilder = button.ButtonBuilder
)

// ButtonFocusStyle type and constants
type ButtonFocusStyle = button.ButtonFocusStyle

const (
	FocusStyleReverse   = button.FocusStyleReverse
	FocusStyleUnderline = button.FocusStyleUnderline
	FocusStyleBracket   = button.FocusStyleBracket
	FocusStyleBold      = button.FocusStyleBold
)

// ButtonVariant type and constants
type ButtonVariant = button.ButtonVariant

const (
	ButtonVariantDefault   = button.ButtonVariantDefault
	ButtonVariantPrimary   = button.ButtonVariantPrimary
	ButtonVariantSecondary = button.ButtonVariantSecondary
	ButtonVariantDanger    = button.ButtonVariantDanger
	ButtonVariantSuccess   = button.ButtonVariantSuccess
)

// Form components
var (
	Input        = form.Input
	NewInput     = form.NewInput
	InputBuilder = form.InputBuilder

	Textarea        = form.Textarea
	NewTextarea     = form.NewTextarea
	TextareaBuilder = form.TextareaBuilder

	Checkbox        = form.Checkbox
	NewCheckbox     = form.NewCheckbox
	CheckboxBuilder = form.CheckboxBuilder

	Select        = form.Select
	NewSelect     = form.NewSelect
	SelectBuilder = form.SelectBuilder
)

// SelectOption type alias for convenience
type SelectOption = form.SelectOption

// Layout components
var (
	HStack        = layout.HStack
	VStack        = layout.VStack
	Box           = layout.Box
	Spacer        = layout.Spacer
	Absolute      = layout.Absolute
	AbsoluteBuilder = layout.AbsoluteBuilder
	Grid          = layout.Grid
	GridBuilder   = layout.GridBuilder
	Center        = layout.Center
	Wrap          = layout.Wrap
	WrapBuilder   = layout.NewWrapBuilder
)

// Grid dimension types
type (
	Fixed = layout.Fixed
	Flex  = layout.Flex
	Auto  = layout.Auto
)

// Absolute position type
type AbsolutePosition = layout.AbsolutePosition

// Feedback components
var (
	Progress       = feedback.Progress
	NewProgress    = feedback.NewProgress
	ProgressBuilder = feedback.ProgressBuilder

	Spinner       = feedback.Spinner
	NewSpinner    = feedback.NewSpinner
	SpinnerBuilder = feedback.SpinnerBuilder
)

// Data components
var (
	Table         = data.Table
	NewTable      = data.NewTable
	TableBuilder  = data.TableBuilder

	VirtualList         = data.VirtualList
	NewVirtualList      = data.NewVirtualList
	VirtualListBuilder  = data.VirtualListBuilder
)

// TableColumn type alias for convenience
type TableColumn = data.TableColumn

// Navigation components
var (
	Tabs         = navigation.Tabs
	NewTabs      = navigation.NewTabs
	TabsBuilder  = navigation.TabsBuilder
)

// Overlay components
var (
	Modal         = overlay.Modal
	NewModal      = overlay.NewModal
	ModalBuilder  = overlay.ModalBuilder

	Tooltip         = overlay.Tooltip
	TooltipBuilder  = overlay.TooltipBuilder

	Toast         = overlay.Toast
	ToastBuilder  = overlay.ToastBuilder
	NewToast      = overlay.NewToast
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
func Run(app ui.ComponentFunc, opts ...Option) error {
	// Convert app.Option to ui.Option
	uiOpts := make([]ui.Option, len(opts))
	for i, opt := range opts {
		uiOpts[i] = ui.Option(opt)
	}

	// Use ui.Run which handles the framework integration
	return ui.Run(app, uiOpts...)
}

// Quit exits the application
func Quit() {
	if appInstance != nil {
		appInstance.Quit()
	}
}
