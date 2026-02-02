package ui

import rtui "github.com/wwsheng009/mint/runtime/ui"

// =============================================================================
// Fragment VNode (re-exported from runtime/ui)
// =============================================================================

// FragmentVNode represents a fragment that doesn't add extra nodes
type FragmentVNode = rtui.FragmentVNode

// NewFragment creates a new fragment VNode
func NewFragment(children ...VNode) *FragmentVNode {
	return rtui.NewFragment(children...)
}

// Fragment creates a new fragment with the given children
func Fragment(children ...VNode) VNode {
	return rtui.Fragment(children...)
}
