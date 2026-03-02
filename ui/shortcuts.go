// Package ui provides convenient shortcuts for common UI components
//
// These shortcuts redirect to Fiber-first components in ui/components/.
// For full-featured Builder patterns, use the component packages directly:
//   - ui/components/text.NewBuilder(content)
//   - ui/components/button.NewBuilder(label)
//   - ui/components/input.NewBuilder()
//
// Or use the app package for convenience re-exports:
//   - app.Text(content) / app.Button(label)
//   - app.Input() / app.Checkbox(label)
package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"

	"github.com/wwsheng009/mint/ui/components/absolute"
	"github.com/wwsheng009/mint/ui/components/border"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/ui/components/grid"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/list"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/progress"
	"github.com/wwsheng009/mint/ui/components/scrollview"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
	"github.com/wwsheng009/mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/table"
	"github.com/wwsheng009/mint/ui/components/tabs"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/textarea"
	"github.com/wwsheng009/mint/ui/components/tooltip"
	"github.com/wwsheng009/mint/ui/components/treeview"
	"github.com/wwsheng009/mint/ui/components/virtuallist"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

// =============================================================================
// Form Components shortcuts
// =============================================================================

// Input shortcuts - redirect to ui/components/input

// Input creates an input field with placeholder
func Input(placeholder string) rtui.VNode {
	return input.NewBuilder().Placeholder(placeholder).Build()
}

// InputWithValue creates an input field with initial value and placeholder
func InputWithValue(placeholder, value string) rtui.VNode {
	return input.NewBuilder().Placeholder(placeholder).Value(value).Build()
}

// Textarea shortcuts - redirect to ui/components/textarea

// Textarea creates a textarea field with placeholder
func Textarea(placeholder string) rtui.VNode {
	return textarea.NewBuilder().Placeholder(placeholder).Build()
}

// TextareaWithValue creates a textarea with initial value
func TextareaWithValue(placeholder, value string) rtui.VNode {
	return textarea.NewBuilder().Placeholder(placeholder).Value(value).Build()
}

// Checkbox shortcuts - redirect to ui/components/checkbox

// Checkbox creates a checkbox
func Checkbox(label string, checked bool) rtui.VNode {
	return checkbox.NewBuilder().Label(label).Checked(checked).Build()
}

// Select shortcuts - redirect to ui/components/selectcomp

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

// =============================================================================
// Button shortcuts - redirect to ui/components/button (Intent-based)
// =============================================================================

// Button creates a button with label (no click handler)
// Note: This version does not support onClick. Use ButtonWithIntent for actions.
func Button(label string) rtui.VNode {
	return button.NewBuilder(label).Build()
}

// ButtonWithIntent creates a button with Intent (Fiber-first pattern)
// This is the recommended way to create buttons with actions.
func ButtonWithIntent(label string, pressIntent intent.Intent) rtui.VNode {
	return button.NewBuilder(label).OnPress(pressIntent).Build()
}

// =============================================================================
// Display Components shortcuts
// =============================================================================

// Text shortcuts - redirect to ui/components/text

// Text creates a simple text VNode
func Text(content string) rtui.VNode {
	return text.T(content)
}

// Textf creates a formatted text VNode
// Note: actual formatting should be done by caller with fmt.Sprintf
func Textf(format string, args ...interface{}) rtui.VNode {
	// Note: actual formatting should be done by caller with fmt.Sprintf
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

// TextAlign creates a text VNode with horizontal alignment
// Supported align values: "left", "center", "right"
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

// Progress shortcuts - redirect to ui/components/progress

// Progress creates a progress bar
func Progress(value, max int) rtui.VNode {
	return progress.NewBuilder().Value(value).Max(max).Build()
}

// ProgressPercent creates a progress bar with percentage
func ProgressPercent(percent int) rtui.VNode {
	return progress.NewBuilder().Value(percent).Max(100).Build()
}

// =============================================================================
// Layout Components shortcuts (some are already in ui/layout.go)
// =============================================================================

// H creates a horizontal stack (HStack) - redirects to ui/components/stack
func H(children ...rtui.VNode) rtui.VNode {
	return stack.HBox(children...)
}

// V creates a vertical stack (VStack) - redirects to ui/components/stack
func V(children ...rtui.VNode) rtui.VNode {
	return stack.VBox(children...)
}

// HBox creates an HStack with children (alias for H)
func HBox(children ...rtui.VNode) rtui.VNode {
	return H(children...)
}

// VBox creates a VStack with children (alias for V)
func VBox(children ...rtui.VNode) rtui.VNode {
	return V(children...)
}

// RowStack creates a row layout with gap
func RowStack(gap int, children ...rtui.VNode) rtui.VNode {
	return stack.RowStack(gap, children...)
}

// ColStack creates a column layout with gap
func ColStack(gap int, children ...rtui.VNode) rtui.VNode {
	return stack.ColStack(gap, children...)
}

// Wrap shortcuts - redirect to ui/components/wrap
// (Note: ui/layout.go also has Wrap(), this provides additional variants)

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

// Grid shortcuts - redirect to ui/components/grid

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

// Absolute shortcuts - redirect to ui/components/absolute

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

// Center places a child at center of container
func CenterAbs(child rtui.VNode) rtui.VNode {
	return absolute.Center(child)
}

// =============================================================================
// Container Components shortcuts
// =============================================================================

// Border shortcuts - redirect to ui/components/border

// Border creates a box with single border
func Border(child rtui.VNode) rtui.VNode {
	return border.B(child)
}

// B creates a single-line border (alias for Border, avoid collision with rtui.Bordered())
func Bc(child rtui.VNode) rtui.VNode {
	return Border(child)
}

// Single creates a box with single border
func Single(child rtui.VNode) rtui.VNode {
	return border.Single(child)
}

// Double creates a box with double border
func Double(child rtui.VNode) rtui.VNode {
	return border.Double(child)
}

// Rounded creates a box with rounded corners
func Rounded(child rtui.VNode) rtui.VNode {
	return border.Rounded(child)
}

// Dashed creates a box with dashed border
func Dashed(child rtui.VNode) rtui.VNode {
	return border.Dashed(child)
}

// WithLabel creates a border with an optional label
func WithLabel(label string, child rtui.VNode) rtui.VNode {
	return border.WithLabel(label, child)
}

// WithColorBorder creates a border with custom color (renamed to avoid collision)
func WithColorBorder(color string, child rtui.VNode) rtui.VNode {
	return border.WithColor(color, child)
}

// Panel shortcuts - redirect to ui/components/panel

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

// ScrollView shortcuts - redirect to ui/components/scrollview

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

// List shortcuts - redirect to ui/components/list

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

// Table shortcuts - redirect to ui/components/table
// (Note: ui/layout.go also has Table(), this is for the full table component)

// TableOf creates a table with columns and rows
func TableOf(columns []string, rows [][]string) rtui.VNode {
	// Convert columns to TableColumn type
	cols := make([]table.TableColumn, len(columns))
	for i, col := range columns {
		cols[i] = table.TableColumn{Title: col}
	}
	return table.Of(cols, rows)
}

// TreeView shortcuts - redirect to ui/components/treeview

// TreeView creates a tree view
func TreeView() *treeview.Builder {
	return treeview.NewBuilder()
}

// TreeViewOf creates a tree view with nodes
func TreeViewOf(nodes []treeview.TreeNode) rtui.VNode {
	return treeview.Of(nodes)
}

// =============================================================================
// VirtualList shortcuts - redirect to ui/components/virtuallist

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

// Tabs shortcuts - redirect to ui/components/tabs

// Tabs creates a tabs component with tab items
func Tabs(tabItems []tabs.TabItem) rtui.VNode {
	return tabs.Of(tabItems)
}

// =============================================================================
// Divider shortcuts - redirect to ui/components/divider
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
func WithAnchorTo(anchorID string, anchor types.Anchor) func(rtui.VNode) rtui.VNode {
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
func FallbackText(txt string) rtui.VNode {
	return rtui.FallbackText(txt)
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
// Toast Notifications shortcuts (from ui/components/tooltip)
// =============================================================================

// ToastInfo creates an info toast notification
func ToastInfo(message string) rtui.VNode {
	return tooltip.Info(message)
}

// ToastSuccess creates a success toast notification
func ToastSuccess(message string) rtui.VNode {
	return tooltip.Success(message)
}

// ToastWarning creates a warning toast notification
func ToastWarning(message string) rtui.VNode {
	return tooltip.Warning(message)
}

// ToastError creates an error toast notification
func ToastError(message string) rtui.VNode {
	return tooltip.Error(message)
}

// Tooltip shortcut - redirect to ui/components/tooltip

// TooltipFor creates a tooltip for a content element
func TooltipFor(content rtui.VNode, tooltipText string) rtui.VNode {
	return tooltip.Tooltip(content, tooltipText)
}
