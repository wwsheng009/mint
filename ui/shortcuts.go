// Package ui provides convenient shortcuts for common UI components
//
// For full-featured components with Builder patterns, use the app package:
//   - app.Text() / app.NewText()
//   - app.Button() / app.NewButton()
//   - app.Input() / app.NewInput()
//   - app.HStack() / app.VStack()
//
// The shortcuts below are provided for convenience and basic use cases.
package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Basic Component Shortcuts
// =============================================================================

// Text creates a simple text VNode
func Text(content string) VNode {
	return rtui.Element("text").Prop("content", content).Build()
}

// Textf creates a formatted text VNode (wrapper, use fmt.Sprintf for actual formatting)
func Textf(format string, args ...interface{}) VNode {
	return rtui.Element("text").Prop("content", format).Build()
}

// TextWithStyle creates a styled text VNode
func TextWithStyle(content string, s style.Style) VNode {
	return rtui.Element("text").Prop("content", content).Style(s).Build()
}

// TextAlign creates a text VNode with horizontal alignment
// Align is applied when text is stretched in a VStack with Stretch()
// Supported align values: "left", "center", "right"
func TextAlign(content string, align string) VNode {
	return rtui.Element("text").Prop("content", content).Prop("textAlign", align).Build()
}

// TextCenter creates a centered text VNode
func TextCenter(content string) VNode {
	return rtui.Element("text").Prop("content", content).Prop("textAlign", "center").Build()
}

// TextRight creates a right-aligned text VNode
func TextRight(content string) VNode {
	return rtui.Element("text").Prop("content", content).Prop("textAlign", "right").Build()
}

// =============================================================================
// Form Component Shortcuts
// =============================================================================

// Input creates an input field
func Input(placeholder string) VNode {
	return rtui.Element("input").
		Prop("placeholder", placeholder).
		Prop("value", "").
		Build()
}

// InputWithValue creates an input field with initial value
func InputWithValue(placeholder, value string) VNode {
	return rtui.Element("input").
		Prop("placeholder", placeholder).
		Prop("value", value).
		Build()
}

// Textarea creates a textarea field
func Textarea(placeholder string) VNode {
	return rtui.Element("textarea").
		Prop("placeholder", placeholder).
		Prop("value", "").
		Build()
}

// Checkbox creates a checkbox
func Checkbox(label string, checked bool) VNode {
	return rtui.Element("checkbox").
		Prop("label", label).
		Prop("checked", checked).
		Build()
}

// =============================================================================
// Button Component Shortcut
// =============================================================================

// Button creates a button with label and click handler
// Note: For full button functionality (disabled state, styling), use app.Button()
func Button(label string, onClick func()) VNode {
	return rtui.Element("button").
		Prop("label", label).
		Prop("onClick", onClick).
		Build()
}

// =============================================================================
// Feedback Component Shortcuts
// =============================================================================

// Progress creates a progress bar
func Progress(value, max int) VNode {
	return rtui.Element("progress").
		Prop("value", value).
		Prop("max", max).
		Build()
}

// ProgressPercent creates a progress bar with percentage
func ProgressPercent(percent int) VNode {
	return rtui.Element("progress").
		Prop("percent", percent).
		Prop("value", percent).
		Prop("max", 100).
		Build()
}

// =============================================================================
// Style Helpers
// =============================================================================

// Styled applies a style to a VNode (only works for ElementVNode)
func Styled(vnode VNode, s style.Style) VNode {
	if elem, ok := vnode.(*rtui.ElementVNode); ok {
		elem.SetStyle(s)
		return elem
	}
	return vnode
}

// WithStyle is a fluent wrapper for Styled
func WithStyle(s style.Style) func(VNode) VNode {
	return func(vnode VNode) VNode {
		return Styled(vnode, s)
	}
}

// =============================================================================
// Key Helper
// =============================================================================

// WithKey adds a key to a VNode for reconciliation
func WithKey(key string) func(VNode) VNode {
	return func(vnode VNode) VNode {
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
// Props Helpers
// =============================================================================

// WithProp adds a single property to a VNode
func WithProp(key string, value interface{}) func(VNode) VNode {
	return func(vnode VNode) VNode {
		if elem, ok := vnode.(*rtui.ElementVNode); ok {
			elem.Props().Set(key, value)
			return elem
		}
		return vnode
	}
}

// WithProps adds multiple properties to a VNode
func WithProps(props map[string]interface{}) func(VNode) VNode {
	return func(vnode VNode) VNode {
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
// Error Boundary Shortcuts
// =============================================================================

// ErrorBoundary creates a new error boundary wrapper
//   - name: identifier for this boundary (for debugging)
//   - component: the component function to wrap
//   - fallback: the VNode to render when errors occur
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
// Memo Shortcuts
// =============================================================================

// Memo wraps a component to memoize its output, skipping re-renders when props haven't changed
// Similar to React.memo() - uses shallow comparison by default
func Memo(component VNode) VNode {
	return rtui.NewMemo(component)
}

// MemoWithCompare wraps a component with a custom comparison function
// The compare function should return true if props are equal (no update needed)
func MemoWithCompare(component VNode, compare rtui.PropsEqual) VNode {
	return rtui.NewMemoWithCompare(component, compare)
}

// MemoComponent wraps a component function with memoization (convenience wrapper)
func MemoizedComponent(name string, fn rtui.ComponentFunc) VNode {
	return rtui.MemoComponent(name, fn)
}

// ShallowPropsEqual performs shallow comparison of two Props objects
// Returns true if they are equal (no update needed)
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
// This is a convenience alias for MemoComponent, equivalent to React's PureComponent
// Use this for components that render the same output given the same props
//
// Example:
//
//	pureExpensive := ui.PureComponent("ExpensiveCalculation", func() ui.VNode {
//	    result := expensiveCalculation()
//	    return app.Text(result)
//	})
func PureComponent(name string, fn rtui.ComponentFunc) VNode {
	return rtui.MemoComponent(name, fn)
}

// PureComponentWithProps creates a memoized component with props
// Only re-renders when props change (shallow comparison)
//
// Example:
//
//	pureItem := ui.PureComponentWithProps("ListItem", func(props rtui.Props) ui.VNode {
//	    title := props.Get("title").(string)
//	    return app.Text(title)
//	})
func PureComponentWithProps(name string, fn rtui.ComponentFuncWithProps) VNode {
	component := rtui.NewComponentWithProps(name, fn)
	return rtui.NewMemo(component)
}
