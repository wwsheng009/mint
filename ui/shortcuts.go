// Package ui provides convenient shortcuts for common UI components
//
// This package re-exports:
// 1. Quick shortcut functions for common use cases (e.g., Text(), Button(), Input())
// 2. Builder factory functions (e.g., NewButtonBuilder(), NewInputBuilder())
//
// Usage examples:
//
// Quick shortcuts:
//
//	ui.Text("Hello")
//	ui.Button("Click Me")
//	ui.Input("Placeholder")
//
// Full Builder patterns:
//
//	ui.NewTextBuilder("Hello").Bold().FgColor("red").Build()
//	ui.NewButtonBuilder("Click").Primary().Large().OnPress(intent).Build()
//	ui.NewInputBuilder().Placeholder("...").Value("x").OnChange(intent).Build()
package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"

	"github.com/wwsheng009/mint/ui/components/absolute"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/grid"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/list"
	"github.com/wwsheng009/mint/ui/components/modal"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/progress"
	"github.com/wwsheng009/mint/ui/components/scrollview"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
	"github.com/wwsheng009/mint/ui/components/statusbar"
	"github.com/wwsheng009/mint/ui/components/table"
	"github.com/wwsheng009/mint/ui/components/tabs"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/textarea"
	"github.com/wwsheng009/mint/ui/components/toast"
	"github.com/wwsheng009/mint/ui/components/tooltip"
	"github.com/wwsheng009/mint/ui/components/treeview"
	"github.com/wwsheng009/mint/ui/components/virtuallist"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

// =============================================================================
// Builder Factory Functions - Re-exported from component packages
// =============================================================================

// Form Components
func NewInputBuilder() *input.Builder {
	return input.NewBuilder()
}

func NewTextareaBuilder() *textarea.Builder {
	return textarea.NewBuilder()
}

func NewCheckboxBuilder() *checkbox.Builder {
	return checkbox.NewBuilder()
}

// Button Components
func NewButtonBuilder(label string) *button.Builder {
	return button.NewBuilder(label)
}

// Text Components
func NewTextBuilder(content string) *text.Builder {
	return text.NewBuilder(content)
}

func NewWrapBuilder() *wrap.Builder {
	return wrap.NewBuilder()
}

func NewGridBuilder() *grid.Builder {
	return grid.NewBuilder()
}

func NewAbsoluteBuilder(child rtui.VNode) *absolute.Builder {
	return absolute.NewBuilder(child)
}

// Container Components

// NewBorderBuilder creates a border builder.
func NewPanelBuilder() *panel.Builder {
	return panel.NewBuilder()
}

func NewScrollViewBuilder() *scrollview.Builder {
	return scrollview.NewBuilder()
}

func NewStatusBarBuilder() *statusbar.Builder {
	return statusbar.NewBuilder()
}

func StatusBar(left, center, right []statusbar.Section) rtui.VNode {
	return statusbar.NewBuilder().
		LeftSections(left...).
		CenterSections(center...).
		RightSections(right...).
		Build()
}

func StatusBarWithHelp(theme statusbar.Theme, helpFallback string, left, center, right []statusbar.Section) rtui.VNode {
	return StatusBarWithHelpMode(theme, helpFallback, statusbar.HelpDisplayInline, left, center, right)
}

func StatusBarWithHelpMode(theme statusbar.Theme, helpFallback string, mode statusbar.HelpDisplayMode, left, center, right []statusbar.Section) rtui.VNode {
	return statusbar.NewBuilder().
		Theme(theme).
		HelpFallback(helpFallback).
		HelpDisplayMode(mode).
		LeftSections(left...).
		CenterSections(center...).
		RightSections(right...).
		BuildWithHelp()
}

// Data Display Components
func NewListBuilder() *list.Builder {
	return list.NewBuilder()
}

func NewTableBuilder() *table.Builder {
	return table.NewBuilder()
}

func NewTreeViewBuilder() *treeview.Builder {
	return treeview.NewBuilder()
}

func NewVirtualListBuilder() *virtuallist.Builder {
	return virtuallist.NewBuilder()
}
func NewTabsBuilder() *tabs.Builder {
	return tabs.NewBuilder()
}
func NewSelectBuilder() *selectcomp.Builder {
	return selectcomp.NewBuilder()
}

// Navigation Components
// Note: NewTabsBuilder() is declared above

// Divider Components
func NewDividerBuilder() *divider.Builder {
	return divider.NewBuilder()
}

// Progress Components
func NewProgressBuilder() *progress.Builder {
	return progress.NewBuilder()
}

// Modal Components
func NewModalBuilder() *modal.Builder {
	return modal.NewBuilder()
}

// Tooltip Components
func NewTooltipBuilder(content rtui.VNode, tooltipText string) *tooltip.Builder {
	return tooltip.NewBuilder(content, tooltipText)
}

func NewToastBuilder(message string) *toast.ToastBuilder {
	return toast.NewToastBuilder(message)
}

// =============================================================================
// Common Type Re-exports
// =============================================================================

// Button Types
type (
	ButtonVariant    = button.Variant
	ButtonSize       = button.Size
	ButtonFocusStyle = button.FocusStyle
)

const (
	ButtonVariantDefault   = button.VariantDefault
	ButtonVariantPrimary   = button.VariantPrimary
	ButtonVariantSecondary = button.VariantSecondary
	ButtonVariantDanger    = button.VariantDanger
	ButtonVariantSuccess   = button.VariantSuccess

	ButtonSmall  = button.SizeSmall
	ButtonMedium = button.SizeMedium
	ButtonLarge  = button.SizeLarge

	FocusStyleReverse   = button.FocusStyleReverse
	FocusStyleUnderline = button.FocusStyleUnderline
	FocusStyleBracket   = button.FocusStyleBracket
	FocusStyleBold      = button.FocusStyleBold
)

// Divider Types
type (
	DividerStyle       = divider.Style
	DividerOrientation = divider.Orientation
)

const (
	DividerSolid  = divider.StyleSolid
	DividerDashed = divider.StyleDashed
	DividerDotted = divider.StyleDotted
	DividerDouble = divider.StyleDouble

	HorizontalDivider = divider.Horizontal
	VerticalDivider   = divider.Vertical
)

// StatusBar Types
type (
	StatusBarSection           = statusbar.Section
	StatusBarTheme             = statusbar.Theme
	StatusBarOverflow          = statusbar.OverflowMode
	StatusBarHelpDisplay       = statusbar.HelpDisplayMode
	StatusBarTooltipPlacement  = statusbar.TooltipPlacement
	StatusBarTooltipArrowStyle = statusbar.TooltipArrowStyle
)

const (
	StatusBarOverflowEllipsis = statusbar.OverflowEllipsis
	StatusBarOverflowClip     = statusbar.OverflowClip

	StatusBarHelpInline  = statusbar.HelpDisplayInline
	StatusBarHelpOverlay = statusbar.HelpDisplayOverlay
	StatusBarHelpBoth    = statusbar.HelpDisplayBoth

	StatusBarTooltipAuto   = statusbar.TooltipPlacementAuto
	StatusBarTooltipTop    = statusbar.TooltipPlacementTop
	StatusBarTooltipBottom = statusbar.TooltipPlacementBottom

	StatusBarTooltipArrowDefault = statusbar.TooltipArrowStyleDefault
	StatusBarTooltipArrowSharp   = statusbar.TooltipArrowStyleSharp
	StatusBarTooltipArrowRounded = statusbar.TooltipArrowStyleRounded
)

// Grid Types
type (
	GridDimension = grid.Dimension
	GridFixed     = grid.Fixed
	GridFlex      = grid.Flex
	GridAuto      = grid.Auto
)

// Grid dimension helper functions - avoid conflict with ui.Flex function
func FixedDim(size int) grid.Fixed { return grid.Fixed(size) }
func FlexDim(factor int) grid.Flex { return grid.Flex{Factor: factor} }
func AutoDim() grid.Auto           { return grid.Auto{} }

// Absolute Position Types
type PositionValue = absolute.PositionValue
type Anchor = absolute.Anchor

// Tab Types
type TabPosition = tabs.TabPosition
type TabItem = tabs.TabItem

const (
	TabPositionTop    = tabs.TabPositionTop
	TabPositionBottom = tabs.TabPositionBottom
	TabPositionLeft   = tabs.TabPositionLeft
	TabPositionRight  = tabs.TabPositionRight
)

// Table Types
type TableColumn = table.TableColumn

// Tree Types
type TreeNode = treeview.TreeNode
type TreeSelectionMode = treeview.SelectionMode

const (
	TreeSelectionNone     = treeview.SelectionNone
	TreeSelectionSingle   = treeview.SelectionSingle
	TreeSelectionMultiple = treeview.SelectionMultiple
)

// Toast Types
type ToastType = toast.ToastType

const (
	ToastTypeInfo    = toast.ToastInfo
	ToastTypeSuccess = toast.ToastSuccess
	ToastTypeWarning = toast.ToastWarning
	ToastTypeError   = toast.ToastError
)

// Select Types
type SelectOption = selectcomp.Option

// NewSelectOption creates a new select option with value and label
func NewSelectOption(value, label string) selectcomp.Option {
	return selectcomp.Option{Value: value, Label: label}
}

// =============================================================================
// NOTE: BorderStyle constants are in ui/layout.go (re-exported from rtui)
// =============================================================================
// NOTE: HStackBuilder, VStackBuilder are in ui/layout.go (runtime/ui.LayoutBuilder)
// NOTE: ModalBuilder, TooltipBuilder are in ui/layer.go (ui-specific implementations)

// =============================================================================
// Quick Shortcut Functions (for common use cases)
// =============================================================================

// Form Components shortcuts

// Input creates an input field with placeholder
func Input(placeholder string) rtui.VNode {
	return input.NewBuilder().Placeholder(placeholder).Build()
}

// InputWithValue creates an input field with initial value and placeholder
func InputWithValue(placeholder, value string) rtui.VNode {
	return input.NewBuilder().Placeholder(placeholder).Value(value).Build()
}

// Textarea creates a textarea field with placeholder
func Textarea(placeholder string) rtui.VNode {
	return textarea.NewBuilder().Placeholder(placeholder).Build()
}

// TextareaWithValue creates a textarea with initial value
func TextareaWithValue(placeholder, value string) rtui.VNode {
	return textarea.NewBuilder().Placeholder(placeholder).Value(value).Build()
}

// Checkbox creates a checkbox
func Checkbox(label string, checked bool) rtui.VNode {
	return checkbox.NewBuilder().Label(label).Checked(checked).Build()
}

// Select creates a select dropdown with options
func Select(options []map[string]interface{}) rtui.VNode {
	// Convert options to selectcomp.Option format
	opts := make([]selectcomp.Option, len(options))
	for i, opt := range options {
		value, _ := opt["value"].(string)
		label, _ := opt["label"].(string)
		opts[i] = selectcomp.Option{Value: value, Label: label}
	}
	return selectcomp.NewBuilder().Options(opts).Build()
}

// Button shortcuts

// Button creates a button with label (no click handler)
// Note: This version does not support onClick. Use ButtonWithIntent for actions.
func Button(label string) rtui.VNode {
	return button.NewBuilder(label).Build()
}

// ButtonWithIntent creates a button with Intent (Fiber-first pattern)
func ButtonWithIntent(label string, pressIntent intent.Intent) rtui.VNode {
	return button.NewBuilder(label).OnPress(pressIntent).Build()
}

// Display Components shortcuts

// Text shortcuts

// Text creates a simple text VNode
func Text(content string) rtui.VNode {
	return text.T(content)
}

// Textf creates a formatted text VNode
// Note: actual formatting should be done by caller with fmt.Sprintf
func Textf(format string, args ...interface{}) rtui.VNode {
	return text.NewBuilder(format).Build()
}

// TextWithStyle creates a styled text VNode
func TextWithStyle(content string, s style.Style) rtui.VNode {
	return text.Styled(content, s)
}

// TextBold creates a bold text VNode
func TextBold(content string) rtui.VNode {
	return text.Bold(content)
}

// TextColored creates a colored text VNode
func TextColored(content string, fg style.Color) rtui.VNode {
	return text.Colored(content, fg)
}

// StatusBarText creates a plain status bar section.
func StatusBarText(content string) statusbar.Section {
	return statusbar.Text(content)
}

// StatusBarActionText creates a clickable plain status bar section.
func StatusBarActionText(content string, pressIntent intent.Intent) statusbar.Section {
	return statusbar.ActionText(content, pressIntent)
}

func StatusBarSections(sections ...statusbar.Section) []statusbar.Section {
	return statusbar.Sections(sections...)
}

// StatusBarBadge creates a highlighted status bar section.
func StatusBarBadge(content, fgColor, bgColor string) statusbar.Section {
	return statusbar.Badge(content, fgColor, bgColor)
}

// StatusBarActionBadge creates a clickable highlighted status bar section.
func StatusBarActionBadge(content, fgColor, bgColor string, pressIntent intent.Intent) statusbar.Section {
	return statusbar.ActionBadge(content, fgColor, bgColor, pressIntent)
}

// StatusBarHelp creates a help/tooltip text for a section.
func StatusBarHelp(section statusbar.Section, helpText string) statusbar.Section {
	return section.WithHelp(helpText)
}

// StatusBarThemeDefault returns the default status bar theme.
func StatusBarThemeDefault() statusbar.Theme {
	return statusbar.DefaultTheme()
}

// StatusBarThemeMuted returns the muted status bar theme.
func StatusBarThemeMuted() statusbar.Theme {
	return statusbar.MutedTheme()
}

// StatusBarThemeContrast returns the contrast status bar theme.
func StatusBarThemeContrast() statusbar.Theme {
	return statusbar.ContrastTheme()
}

func StatusBarWithTheme(theme statusbar.Theme, left, center, right []statusbar.Section) rtui.VNode {
	return statusbar.NewBuilder().
		Theme(theme).
		LeftSections(left...).
		CenterSections(center...).
		RightSections(right...).
		Build()
}

// TextAlign creates a text VNode with horizontal alignment
func TextAlign(content string, align string) rtui.VNode {
	var a rtui.Align
	switch align {
	case "center":
		a = rtui.AlignCenter
	case "right":
		a = rtui.AlignEnd
	default:
		a = rtui.AlignStart
	}
	return text.NewBuilder(content).TextAlign(a).Build()
}

// TextCenter creates a centered text VNode
func TextCenter(content string) rtui.VNode {
	return TextAlign(content, "center")
}

// TextRight creates a right-aligned text VNode
func TextRight(content string) rtui.VNode {
	return TextAlign(content, "right")
}

// Progress shortcuts

// Progress creates a progress bar
func Progress(value, max int) rtui.VNode {
	return progress.NewBuilder().Value(value).Max(max).Build()
}

// ProgressPercent creates a progress bar with percentage
func ProgressPercent(percent int) rtui.VNode {
	return progress.NewBuilder().Value(percent).Max(100).Build()
}

// Wrap shortcuts

// WrapWithWidth creates a wrap layout with specified width
func WrapWithWidth(width int, children ...rtui.VNode) rtui.VNode {
	w := wrap.W()
	w.Width(width)
	return w.Children(children...).Build()
}

// WrapWithGap creates a wrap layout with specified gap
func WrapWithGap(gap int, children ...rtui.VNode) rtui.VNode {
	return wrap.W().Gap(gap).Children(children...).Build()
}

// Grid shortcuts

// Grid creates a simple grid with specified number of columns
func Grid(numCols int, children ...rtui.VNode) rtui.VNode {
	return grid.SimpleGrid(numCols, children...)
}

// TwoColumnGrid creates a two-column grid layout
func TwoColumnGrid(children ...rtui.VNode) rtui.VNode {
	return grid.TwoColumnGrid(children...)
}

// ThreeColumnGrid creates a three-column grid layout
func ThreeColumnGrid(children ...rtui.VNode) rtui.VNode {
	return grid.ThreeColumnGrid(children...)
}

// Absolute shortcuts

// At places a child at absolute coordinates (x, y)
func At(child rtui.VNode, x, y int) rtui.VNode {
	return absolute.At(child, x, y)
}

// TopLeft places a child at top-left corner
func TopLeft(child rtui.VNode) rtui.VNode {
	return absolute.TopLeft(child)
}

// TopRight places a child at top-right corner
func TopRight(child rtui.VNode) rtui.VNode {
	return absolute.TopRight(child)
}

// BottomLeft places a child at bottom-left corner
func BottomLeft(child rtui.VNode) rtui.VNode {
	return absolute.BottomLeft(child)
}

// BottomRight places a child at bottom-right corner
func BottomRight(child rtui.VNode) rtui.VNode {
	return absolute.BottomRight(child)
}

// CenterAbs places a child at center of container
func CenterAbs(child rtui.VNode) rtui.VNode {
	return absolute.Center(child)
}

// =============================================================================
// Container Components shortcuts
// =============================================================================

// Panel shortcuts

// Panel creates a panel with content
func Panel(content rtui.VNode) rtui.VNode {
	return panel.Of(content)
}

// PanelOfSize creates a panel with specified size
func PanelOfSize(content rtui.VNode, width, height int) rtui.VNode {
	return panel.OfSize(content, width, height)
}

// PanelTitled creates a panel with title
func PanelTitled(title string, content rtui.VNode) rtui.VNode {
	return panel.Titled(title, content)
}

// PanelBordered creates a panel with border and specified size
func PanelBordered(content rtui.VNode, width, height int) rtui.VNode {
	return panel.Bordered(content, width, height)
}

// ScrollView shortcuts

// ScrollView creates a scrollable view
func ScrollView(child rtui.VNode) rtui.VNode {
	return scrollview.Scroll(child)
}

// Scroll creates a scrollable view (alias for ScrollView)
func Scroll(child rtui.VNode) rtui.VNode {
	return scrollview.Scroll(child)
}

// ScrollSize creates a scrollable view with specified size
func ScrollSize(child rtui.VNode, width, height int) rtui.VNode {
	return scrollview.ScrollSize(child, width, height)
}

// ScrollBordered creates a bordered scrollable view with specified size
func ScrollBordered(child rtui.VNode, width, height int) rtui.VNode {
	return scrollview.Bordered(child, width, height)
}

// =============================================================================
// Data Display Components shortcuts
// =============================================================================

// List shortcuts

// List creates a list component
func List() *list.Builder {
	return list.NewBuilder()
}

// ListOf creates a list with given rows
func ListOf(rows []string) rtui.VNode {
	return list.Of(rows)
}

// ListWithHeader creates a list with header and rows
func ListWithHeader(header string, rows []string) rtui.VNode {
	return list.WithHeader(header).Rows(rows).Build()
}

// Table shortcuts

// TableOf creates a table with columns and rows
func TableOf(columns []string, rows [][]string) rtui.VNode {
	// Convert columns to TableColumn type
	cols := make([]table.TableColumn, len(columns))
	for i, col := range columns {
		cols[i] = table.TableColumn{Title: col}
	}
	return table.Of(cols, rows)
}

// TreeView shortcuts

// TreeView creates a tree view
func TreeView() *treeview.Builder {
	return treeview.NewBuilder()
}

// TreeViewOf creates a tree view with nodes
func TreeViewOf(nodes []treeview.TreeNode) rtui.VNode {
	return treeview.Of(nodes)
}

// VirtualList shortcuts

// VirtualList creates a virtual list
func VirtualList() *virtuallist.Builder {
	return virtuallist.NewBuilder()
}

// VirtualListOfSize creates a virtual list with items
func VirtualListOfSize(items []string, width, height int) rtui.VNode {
	return virtuallist.Of(items)
}

// =============================================================================
// Navigation Components shortcuts
// =============================================================================

// Tabs creates a tabs component with tab items
func Tabs(tabItems []tabs.TabItem) rtui.VNode {
	return tabs.Of(tabItems)
}

// =============================================================================
// Divider shortcuts
// =============================================================================

// Divider creates a simple horizontal divider
func Divider() rtui.VNode {
	return divider.D()
}

// DividerWithLabel creates a horizontal divider with label
func DividerWithLabel(label string) rtui.VNode {
	return divider.H(label)
}

// VDivider creates a vertical divider
func VDivider() rtui.VNode {
	return divider.V()
}

// HDivider creates a horizontal divider (alias for Divider)
func HDivider() rtui.VNode {
	return divider.D()
}

// DividerSection creates a section divider with title
func DividerSection(title string) rtui.VNode {
	return divider.Section(title)
}

// =============================================================================
// Modal shortcuts
// =============================================================================

// Note: Modal() function exists in ui/layer.go (returns *ModalBuilder)
// The following are shortcuts for ui/components/modal.Of():

// ModalOfSize creates a modal with specified size
func ModalOfSize(content rtui.VNode, width, height int) rtui.VNode {
	return modal.OfSize(content, width, height)
}

// ModalTitled creates a modal with title
func ModalTitled(title string, content rtui.VNode) rtui.VNode {
	return modal.Titled(title, content)
}

// ModalAlert creates an alert modal dialog
func ModalAlert(title, message string) rtui.VNode {
	return modal.Alert(title, message)
}

// ModalConfirm creates a confirm modal dialog
func ModalConfirm(title, message string) rtui.VNode {
	return modal.Confirm(title, message)
}

// =============================================================================
// Style Helpers - VNode manipulation helpers
// =============================================================================

// Styled applies a style to a VNode (only works for ElementVNode)
func Styled(vnode rtui.VNode, s style.Style) rtui.VNode {
	if elem, ok := vnode.(*rtui.ElementVNode); ok {
		elem.SetStyle(s)
		return elem
	}
	return vnode
}

// WithStyle is a fluent wrapper for Styled
func WithStyle(s style.Style) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return Styled(vnode, s)
	}
}

// =============================================================================
// Key Helpers
// =============================================================================

// WithKey adds a key to a VNode for reconciliation
func WithKey(key string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		if elem, ok := vnode.(*rtui.ElementVNode); ok {
			elem.SetKey(key)
			return elem
		}
		if comp, ok := vnode.(*rtui.ComponentVNode); ok {
			comp.SetKey(key)
			return comp
		}
		return vnode
	}
}

// =============================================================================
// ID Helpers
// =============================================================================

// WithID adds an ID to a VNode for Portal anchoring
func WithID(id string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		vnode.SetID(id)
		return vnode
	}
}

// =============================================================================
// Props Helpers
// =============================================================================

// WithProp adds a single property to a VNode
func WithProp(key string, value interface{}) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		if elem, ok := vnode.(*rtui.ElementVNode); ok {
			elem.Props().Set(key, value)
			return elem
		}
		return vnode
	}
}

// WithProps adds multiple properties to a VNode
func WithProps(props map[string]interface{}) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		if elem, ok := vnode.(*rtui.ElementVNode); ok {
			for k, v := range props {
				elem.Props().Set(k, v)
			}
			return elem
		}
		return vnode
	}
}

// =============================================================================
// Portal Helpers - Functional-style helpers for Portal configuration
// =============================================================================

// WithPortalRoot is a functional helper that sets the portalRoot property
func WithPortalRoot(portalRootID string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalRoot(portalRootID)
	}
}

// WithAnchorTo is a functional helper that sets anchorId and anchor properties
func WithAnchorTo(anchorID string, anchor Anchor) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetAnchorTo(anchorID, anchor)
	}
}

// WithPortalPosition is a functional helper that sets the position property
func WithPortalPosition(position types.PositionType) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalPosition(position)
	}
}

// WithPortalPriority is a functional helper that sets the priority property
func WithPortalPriority(priority int) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalPriority(priority)
	}
}

// WithPortalRootId is a functional helper that sets the portalRootId property
func WithPortalRootId(portalRootId string) func(rtui.VNode) rtui.VNode {
	return func(vnode rtui.VNode) rtui.VNode {
		return vnode.SetPortalRootId(portalRootId)
	}
}

// =============================================================================
// Error Boundary Shortcuts - Runtime feature
// =============================================================================

// ErrorBoundary creates a new error boundary wrapper
func ErrorBoundary(name string, component rtui.ComponentFunc, fallback rtui.VNode) rtui.VNode {
	return rtui.ErrorBoundary(name, component, fallback)
}

// FallbackText creates a simple text fallback
func FallbackText(text string) rtui.VNode {
	return rtui.FallbackText(text)
}

// FallbackError creates an error message fallback with details
func FallbackError(prefix string) rtui.VNode {
	return rtui.FallbackError(prefix)
}

// FallbackBox creates a boxed error message
func FallbackBox(title, message string) rtui.VNode {
	return rtui.FallbackBox(title, message)
}

// =============================================================================
// Memo Shortcuts - Runtime optimization
// =============================================================================

// Memo wraps a component to memoize its output
func Memo(component rtui.VNode) rtui.VNode {
	return rtui.NewMemo(component)
}

// MemoWithCompare wraps a component with a custom comparison function
func MemoWithCompare(component rtui.VNode, compare rtui.PropsEqual) rtui.VNode {
	return rtui.NewMemoWithCompare(component, compare)
}

// MemoComponent wraps a component function with memoization
func MemoizedComponent(name string, fn rtui.ComponentFunc) rtui.VNode {
	return rtui.MemoComponent(name, fn)
}

// ShallowPropsEqual performs shallow comparison of two Props objects
func ShallowPropsEqual(oldProps, newProps rtui.Props) bool {
	return rtui.ShallowPropsEqual(oldProps, newProps)
}

// PropsEqualExcept creates a comparison function that ignores specific keys
func PropsEqualExcept(exceptKeys ...string) rtui.PropsEqual {
	return rtui.PropsEqualExcept(exceptKeys...)
}

// PropsEqualOnly creates a comparison function that only checks specific keys
func PropsEqualOnly(keys ...string) rtui.PropsEqual {
	return rtui.PropsEqualOnly(keys...)
}

// PureComponent creates a memoized component that only re-renders when props change
func PureComponent(name string, fn rtui.ComponentFunc) rtui.VNode {
	return rtui.MemoComponent(name, fn)
}

// PureComponentWithProps creates a memoized component with props
func PureComponentWithProps(name string, fn rtui.ComponentFuncWithProps) rtui.VNode {
	component := rtui.NewComponentWithProps(name, fn)
	return rtui.NewMemo(component)
}

// =============================================================================
// Toast Notifications shortcuts
// =============================================================================

// ToastInfo creates an info toast notification
func ToastInfo(message string) rtui.VNode {
	return toast.Info(message)
}

// ToastSuccess creates a success toast notification
func ToastSuccess(message string) rtui.VNode {
	return toast.Success(message)
}

// ToastWarning creates a warning toast notification
func ToastWarning(message string) rtui.VNode {
	return toast.Warning(message)
}

// ToastError creates an error toast notification
func ToastError(message string) rtui.VNode {
	return toast.Error(message)
}

// Tooltip shortcut

// TooltipFor creates a tooltip for a content element
func TooltipFor(content rtui.VNode, tooltipText string) rtui.VNode {
	return tooltip.Tooltip(content, tooltipText)
}
