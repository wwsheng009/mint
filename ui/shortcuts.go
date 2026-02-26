// Package ui provides convenient shortcuts for common UI components
//
// These shortcuts redirect to Fiber-first components in ui/components/.
// For full-featured Builder patterns, use the component packages directly:
//   - ui/components/text.New(content)
//   - ui/components/button.New(label)
//   - ui/components/input.New()
//
// Or use the app package for convenience re-exports:
//   - app.Text(content) / app.Button(label)
//   - app.Input() / app.Checkbox(label)
package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/textarea"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/progress"
)

// =============================================================================
// Text Shortcuts - Redirect to ui/components/text
// =============================================================================

// Text creates a simple text VNode
// Shortcut for text.New(content) from ui/components/text
func Text(content string) VNode {
	return text.New(content)
}

// Textf creates a formatted text VNode
// Shortcut for text.New(fmt.Sprintf(format, args...))
func Textf(format string, args ...interface{}) VNode {
	// Note: actual formatting should be done by caller with fmt.Sprintf
	return text.New(format)
}

// TextWithStyle creates a styled text VNode
// Shortcut for text.New(content).SetStyle(s)
func TextWithStyle(content string, s style.Style) VNode {
	return text.New(content).SetStyle(s)
}

// TextAlign creates a text VNode with horizontal alignment
// Shortcut for text.New(content).SetTextAlignProps(align)
// Supported align values: "left", "center", "right"
func TextAlign(content string, align string) VNode {
	var a rtui.Align
	switch align {
	case "center":
		a = rtui.AlignCenter
	case "right":
		a = rtui.AlignEnd
	default:
		a = rtui.AlignStart
	}
	return text.New(content).SetTextAlignProps(a)
}

// TextCenter creates a centered text VNode
// Shortcut for text.New(content).SetTextAlignProps(AlignCenter)
func TextCenter(content string) VNode {
	return text.New(content).SetTextAlignProps(rtui.AlignCenter)
}

// TextRight creates a right-aligned text VNode
// Shortcut for text.New(content).SetTextAlignProps(AlignEnd)
func TextRight(content string) VNode {
	return text.New(content).SetTextAlignProps(rtui.AlignEnd)
}

// =============================================================================
// Form Component Shortcuts - Redirect to ui/components/*
// =============================================================================

// Input creates an input field
// Shortcut for input.New().SetPlaceholder(placeholder) from ui/components/input
func Input(placeholder string) VNode {
	return input.New().SetPlaceholder(placeholder)
}

// InputWithValue creates an input field with initial value
// Shortcut for input.New().SetValue(value).SetPlaceholder(placeholder)
func InputWithValue(placeholder, value string) VNode {
	return input.New().SetValue(value).SetPlaceholder(placeholder)
}

// Textarea creates a textarea field
// Shortcut for textarea.New().SetPlaceholder(placeholder) from ui/components/textarea
func Textarea(placeholder string) VNode {
	return textarea.New().SetPlaceholder(placeholder)
}

// Checkbox creates a checkbox
// Shortcut for checkbox.New(label).SetChecked(checked) from ui/components/checkbox
func Checkbox(label string, checked bool) VNode {
	return checkbox.New(label).SetChecked(checked)
}

// =============================================================================
// Button Shortcuts - Redirect to ui/components/button (Intent-based)
// =============================================================================

// Button creates a button with label (no click handler)
// Shortcut for button.New(label) from ui/components/button
// Note: This version does not support onClick. Use ButtonWithIntent for actions.
func Button(label string) VNode {
	return button.New(label)
}

// ButtonWithIntent creates a button with Intent (Fiber-first pattern)
// Shortcut for button.New(label).SetIntent(intent)
// This is the recommended way to create buttons with actions.
func ButtonWithIntent(label string, pressIntent intent.Intent) VNode {
	return button.New(label).SetIntent(pressIntent)
}

// =============================================================================
// Progress Shortcuts - Redirect to ui/components/progress
// =============================================================================

// Progress creates a progress bar
// Shortcut for progress.New().SetValue(value).SetMax(max) from ui/components/progress
func Progress(value, max int) VNode {
	return progress.New().SetValue(value).SetMax(max)
}

// ProgressPercent creates a progress bar with percentage
// Shortcut for progress.New().SetValue(percent).SetMax(100)
func ProgressPercent(percent int) VNode {
	return progress.New().SetValue(percent).SetMax(100)
}

// =============================================================================
// Style Helpers - Still relevant for VNode manipulation
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
// Error Boundary Shortcuts - Runtime feature, still relevant
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
// Memo Shortcuts - Runtime optimization, still relevant
// =============================================================================

// Memo wraps a component to memoize its output
func Memo(component VNode) VNode {
	return rtui.NewMemo(component)
}

// MemoWithCompare wraps a component with a custom comparison function
func MemoWithCompare(component VNode, compare rtui.PropsEqual) VNode {
	return rtui.NewMemoWithCompare(component, compare)
}

// MemoComponent wraps a component function with memoization
func MemoizedComponent(name string, fn rtui.ComponentFunc) VNode {
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
func PureComponent(name string, fn rtui.ComponentFunc) VNode {
	return rtui.MemoComponent(name, fn)
}

// PureComponentWithProps creates a memoized component with props
func PureComponentWithProps(name string, fn rtui.ComponentFuncWithProps) VNode {
	component := rtui.NewComponentWithProps(name, fn)
	return rtui.NewMemo(component)
}
