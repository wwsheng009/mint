package ui

import rtui "github.com/wwsheng009/mint/runtime/ui"

// =============================================================================
// Component VNode (re-exported from runtime/ui)
// =============================================================================

// ComponentVNode represents a function component
type ComponentVNode = rtui.ComponentVNode

// NewComponent creates a new component VNode
func NewComponent(name string, fn ComponentFunc) *ComponentVNode {
	return rtui.NewComponent(name, fn)
}

// NewComponentWithProps creates a new component VNode that accepts props
func NewComponentWithProps(name string, fn ComponentFuncWithProps) *ComponentVNode {
	return rtui.NewComponentWithProps(name, fn)
}

// Component creates a new component builder
func Component(name string, fn ComponentFunc) *rtui.ComponentBuilder {
	return rtui.Component(name, fn)
}

// ComponentWithProps creates a component builder with props
func ComponentWithProps(name string, fn ComponentFuncWithProps) *rtui.ComponentBuilder {
	return rtui.ComponentWithProps(name, fn)
}
