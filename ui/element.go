package ui

import rtui "github.com/wwsheng009/mint/runtime/ui"

// =============================================================================
// Element VNode (re-exported from runtime/ui)
// =============================================================================

// ElementVNode represents a standard element node
type ElementVNode = rtui.ElementVNode

// NewElement creates a new element VNode
func NewElement(tag string) *ElementVNode {
	return rtui.NewElement(tag)
}

// Element creates a new element builder
func Element(tag string) *rtui.ElementBuilder {
	return rtui.Element(tag)
}
