// Package state provides component instance management for persistent state.
// This is a transitional package that re-exports types from runtime/ui.
// New code should use runtime/ui types directly.
package state

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Type Aliases to runtime/ui
// =============================================================================
// These types are now defined in runtime/ui and re-exported here for compatibility.

// VNode is an alias to rtui.VNode
type VNode = rtui.VNode

// ComponentInstance is an alias to rtui.ComponentInstance
type ComponentInstance = rtui.ComponentInstance

// BaseComponentInstance is an alias to rtui.BaseComponentInstance
type BaseComponentInstance = rtui.BaseComponentInstance

// ComponentContext is an alias to rtui.ComponentContext
type ComponentContext = rtui.ComponentContext

// ComponentFunc is an alias to rtui.ComponentFunc
type ComponentFunc = rtui.ComponentFunc

// ComponentFuncWithProps is an alias to rtui.ComponentFuncWithProps
type ComponentFuncWithProps = rtui.ComponentFuncWithProps

// Props is an alias to rtui.Props
type Props = rtui.Props

// =============================================================================
// Re-exported Functions
// =============================================================================

// NewBaseComponentInstance creates a new base component instance
func NewBaseComponentInstance(key string, fn ComponentFunc) *BaseComponentInstance {
	return rtui.NewBaseComponentInstance(key, fn)
}

// NewBaseComponentInstanceWithProps creates a new base component instance with props
func NewBaseComponentInstanceWithProps(key string, fn ComponentFuncWithProps, props Props) *BaseComponentInstance {
	return rtui.NewBaseComponentInstanceWithProps(key, fn, props)
}

// NewComponentContext creates a new component context
func NewComponentContext(key string) *ComponentContext {
	return rtui.NewComponentContext(key)
}

// SetCurrentContext sets the current component context
func SetCurrentContext(ctx *ComponentContext) {
	rtui.SetCurrentContext(ctx)
}

// GetCurrentContext returns the current component context
func GetCurrentContext() *ComponentContext {
	return rtui.GetCurrentContext()
}
