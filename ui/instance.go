package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Component Instance (re-exported from runtime/types)
// =============================================================================

// ComponentInstance represents a persistent component instance that maintains
// state across renders.
type ComponentInstance = rtui.ComponentInstance

// BaseComponentInstance provides a base implementation of ComponentInstance
// that can be embedded by specific component instances.
type BaseComponentInstance = rtui.BaseComponentInstance

// NewBaseComponentInstance creates a new base component instance
func NewBaseComponentInstance(key string, fn ComponentFunc) *BaseComponentInstance {
	return rtui.NewBaseComponentInstance(key, fn)
}

// NewBaseComponentInstanceWithProps creates a new base component instance with props
func NewBaseComponentInstanceWithProps(key string, fn ComponentFuncWithProps, props Props) *BaseComponentInstance {
	return rtui.NewBaseComponentInstanceWithProps(key, fn, props)
}
