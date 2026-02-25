package app

import (
	"github.com/wwsheng009/mint/framework"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/absolute"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/grid"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/modal"
	"github.com/wwsheng009/mint/ui/components/progress"
	"github.com/wwsheng009/mint/ui/components/scrollview"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
	"github.com/wwsheng009/mint/ui/components/table"
	"github.com/wwsheng009/mint/ui/components/tabs"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/textarea"
	"github.com/wwsheng009/mint/ui/components/tooltip"
	"github.com/wwsheng009/mint/ui/components/virtuallist"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

// =============================================================================
// Component Shortcuts - Re-export for convenience
// =============================================================================

// Basic components (text and divider)
func Text(content string) rtui.VNode              { return text.Text(content) }
func NewText(content string) *text.VNode          { return text.NewText(content) }
func NewTextBuilder(content string) *text.Builder { return text.NewBuilder(content) }
func DividerCmp() *divider.VNode                  { return divider.Divider() }
func NewDivider() *divider.VNode                  { return divider.NewDivider() }
func DividerBuilder() *divider.Builder            { return divider.NewBuilder() }

// DividerStyle type and constants
type DividerStyle = divider.DividerStyle

const (
	DividerSolid   = divider.DividerSolid
	DividerDashed  = divider.DividerDashed
	DividerDotted  = divider.DividerDotted
)

// Button components
func Button(label string) *button.VNode          { return button.Button(label) }
func NewButton(label string) *button.VNode       { return button.NewButton(label) }
func ButtonBuilder(label string) *button.Builder { return button.NewBuilder(label) }

// ButtonFocusStyle type and constants
type ButtonFocusStyle = button.FocusStyle

const (
	FocusStyleReverse   = button.FocusStyleReverse
	FocusStyleUnderline = button.FocusStyleUnderline
	FocusStyleBracket   = button.FocusStyleBracket
	FocusStyleBold      = button.FocusStyleBold
)

// ButtonVariant type and constants
type ButtonVariant = button.Variant

const (
	ButtonVariantDefault   = button.VariantDefault
	ButtonVariantPrimary   = button.VariantPrimary
	ButtonVariantSecondary = button.VariantSecondary
	ButtonVariantDanger    = button.VariantDanger
	ButtonVariantSuccess   = button.VariantSuccess
)

// Form components
func Input() *input.VNode            { return input.Input() }
func NewInput() *input.VNode         { return input.NewInput() }
func InputBuilder() *input.Builder   { return input.NewBuilder() }

func Textarea() *textarea.VNode          { return textarea.Textarea() }
func NewTextarea() *textarea.VNode       { return textarea.NewTextarea() }
func TextareaBuilder() *textarea.Builder { return textarea.NewBuilder() }

func Checkbox(label string) *checkbox.VNode     { return checkbox.Checkbox(label) }
func NewCheckbox(label string) *checkbox.VNode  { return checkbox.NewCheckbox(label) }
func CheckboxBuilder() *checkbox.Builder        { return checkbox.NewBuilder() }

func Select() *selectcomp.VNode          { return selectcomp.Select() }
func NewSelect() *selectcomp.VNode       { return selectcomp.NewSelect() }
func SelectBuilder() *selectcomp.Builder { return selectcomp.NewBuilder() }

// SelectOption type alias for convenience
type SelectOption = selectcomp.Option

// Layout components
func HStack(children ...rtui.VNode) rtui.VNode { return rtui.HStack(children...) }
func VStack(children ...rtui.VNode) rtui.VNode { return rtui.VStack(children...) }
func Box() *rtui.BoxLayoutBuilder              { return rtui.Box() }
func Spacer() *rtui.SpacerBuilder              { return rtui.Spacer() }
func Wrap(children ...rtui.VNode) *wrap.VNode  { return wrap.Wrap(children...) }
func WrapBuilder(children ...rtui.VNode) *wrap.Builder {
	return wrap.NewBuilder().Children(children...)
}

// Grid and Absolute
func Grid() *grid.Builder                        { return grid.NewBuilder() }
func GridBuilder() *grid.Builder                 { return grid.NewBuilder() }
func Absolute(child rtui.VNode) *absolute.Builder { return absolute.NewBuilder(child) }
func AbsoluteBuilder(child rtui.VNode) *absolute.Builder {
	return absolute.NewBuilder(child)
}
func Center(child rtui.VNode) rtui.VNode { return absolute.Center(child) }
func ScrollView() *scrollview.Builder    { return scrollview.ScrollView() }

// Grid dimension types
type (
	Fixed = grid.Fixed
	Flex  = grid.Flex
	Auto  = grid.Auto
)

// Absolute position types
type AbsolutePosition = absolute.PositionValue

// Feedback components
func Progress() *progress.VNode          { return progress.Progress() }
func NewProgress() *progress.VNode       { return progress.NewProgress() }
func ProgressBuilder() *progress.Builder { return progress.NewBuilder() }

// Data components
func Table() *table.Builder               { return table.Table() }
func NewTable() *table.VNode              { return table.NewTable() }
func TableBuilder() *table.Builder        { return table.NewBuilder() }

func VirtualList() *virtuallist.Builder        { return virtuallist.VirtualList() }
func NewVirtualList() *virtuallist.VNode       { return virtuallist.NewVirtualList() }
func VirtualListBuilder() *virtuallist.Builder { return virtuallist.NewBuilder() }

// TableColumn type alias for convenience
type TableColumn = table.TableColumn

// Navigation components
func Tabs() *tabs.Builder        { return tabs.Tabs() }
func NewTabs() *tabs.VNode       { return tabs.NewTabs() }
func TabsBuilder() *tabs.Builder { return tabs.NewBuilder() }

// Overlay components
func Modal() *modal.Builder        { return modal.Modal() }
func NewModal() *modal.VNode       { return modal.NewModal() }
func ModalBuilder() *modal.Builder { return modal.NewBuilder() }

func Tooltip(content rtui.VNode, text string) *tooltip.VNode {
	return tooltip.Tooltip(content, text)
}
func TooltipBuilder(content rtui.VNode, text string) *tooltip.Builder {
	return tooltip.NewBuilder(content, text)
}

func Toast(message string) *tooltip.ToastVNode     { return tooltip.Toast(message) }
func NewToast(message string) *tooltip.ToastVNode  { return tooltip.NewToast(message) }
func ToastBuilder(message string) *tooltip.ToastBuilder {
	return tooltip.NewToastBuilder(message)
}

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
func Run(appFunc rtui.ComponentFunc, opts ...Option) error {
	// Convert app.Option to rtui.Option
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
