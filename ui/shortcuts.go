// Package ui provides basic shortcuts for common VNode patterns
// For full component functionality, use app package which re-exports all components
package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Basic Text VNode Creation
// =============================================================================
// Note: For full-featured components, use app package:
// - app.Text() / app.NewText()
// - app.Button() / app.NewButton()
// - app.Input() / app.NewInput()
// - app.HStack() / app.VStack()

// Text creates a simple text VNode
func Text(content string) VNode {
	return rtui.Element("text").Prop("content", content).Build()
}

// TextWithStyle creates a styled text VNode
func TextWithStyle(content string, s style.Style) VNode {
	return rtui.Element("text").Prop("content", content).Style(s).Build()
}

// =============================================================================
// Style Helper
// =============================================================================

// Styled applies a style to a VNode (only works for ElementVNode)
func Styled(vnode VNode, s style.Style) VNode {
	if elem, ok := vnode.(*rtui.ElementVNode); ok {
		elem.SetStyle(s)
		return elem
	}
	return vnode
}
