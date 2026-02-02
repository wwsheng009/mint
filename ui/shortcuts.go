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
