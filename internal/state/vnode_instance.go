package state

import (
	"github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNodeComponentInstance - Persistent Instance for VNode Struct Components
// =============================================================================
// This implements ComponentInstance for VNode-based components like ButtonVNode.
//
// Architecture:
// - VNode struct components (Button, Text, etc.) are recreated each render
// - But their state and event handlers need to persist
// - VNodeComponentInstance bridges this gap
//
// Usage:
// - Created on first render of a VNode component
// - Reused on subsequent renders when key matches
// - Stores event handlers (onClick, onMouseEnter, etc.)
// - Stores component state (hover, focus, etc.)
// =============================================================================

// VNodeComponentInstance is a persistent instance for VNode struct components
type VNodeComponentInstance struct {
	// Base instance functionality
	*BaseComponentInstance

	// Event handlers (persistent across renders)
	OnClick      func()
	OnMouseEnter func()
	OnMouseLeave func()
	OnKeyPress   func(string)

	// Component state
	State map[string]interface{}

	// Reference to the latest VNode (updated each render) - DEPRECATED
	// Will be removed in Phase 4
	LatestVNode ui.VNode

	// Reference to the Fiber (for Fiber-first architecture)
	Fiber *ui.Fiber
}

// NewVNodeComponentInstance creates a new instance for a VNode component
func NewVNodeComponentInstance(key string, vnode ui.VNode) *VNodeComponentInstance {
	base := NewBaseComponentInstance(key, func() ui.VNode {
		return vnode
	})

	return &VNodeComponentInstance{
		BaseComponentInstance: base,
		State:                 make(map[string]interface{}),
		LatestVNode:           vnode,
	}
}

// NewVNodeComponentInstanceFromFiber creates a new instance from Fiber (Fiber-first)
func NewVNodeComponentInstanceFromFiber(key string, fiber *ui.Fiber) *VNodeComponentInstance {
	base := NewBaseComponentInstance(key, func() ui.VNode {
		return nil // No VNode reference in Fiber-first
	})

	inst := &VNodeComponentInstance{
		BaseComponentInstance: base,
		State:                 make(map[string]interface{}),
		Fiber:                 fiber,
	}

	// Extract event handlers from Fiber props
	inst.extractHandlersFromFiber(fiber)

	return inst
}

// UpdateVNode updates the VNode reference and extracts handlers
func (inst *VNodeComponentInstance) UpdateVNode(vnode ui.VNode) {
	inst.LatestVNode = vnode

	// Extract event handlers from VNode
	inst.extractHandlers(vnode)
}

// UpdateFiber updates the Fiber reference and extracts handlers
func (inst *VNodeComponentInstance) UpdateFiber(fiber *ui.Fiber) {
	inst.Fiber = fiber
	inst.extractHandlersFromFiber(fiber)
}

// extractHandlers extracts event handlers from VNode
func (inst *VNodeComponentInstance) extractHandlers(vnode ui.VNode) {
	// Try to extract onClick
	if clicker, ok := vnode.(interface{ OnClick() func() }); ok {
		inst.OnClick = clicker.OnClick()
	}

	// Try to extract onMouseEnter
	if enterer, ok := vnode.(interface{ OnMouseEnter() func() }); ok {
		inst.OnMouseEnter = enterer.OnMouseEnter()
	}

	// Try to extract onMouseLeave
	if leaver, ok := vnode.(interface{ OnMouseLeave() func() }); ok {
		inst.OnMouseLeave = leaver.OnMouseLeave()
	}

	// Try to extract onKeyPress
	if presser, ok := vnode.(interface{ OnKeyPress() func(string) }); ok {
		inst.OnKeyPress = presser.OnKeyPress()
	}
}

// extractHandlersFromFiber extracts event handlers from Fiber props
func (inst *VNodeComponentInstance) extractHandlersFromFiber(fiber *ui.Fiber) {
	if fiber.Props == nil {
		return
	}

	// Extract onClick from props
	if onClick, ok := fiber.Props["onClick"].(func()); ok {
		inst.OnClick = onClick
	}

	// Extract onMouseEnter from props
	if onMouseEnter, ok := fiber.Props["onMouseEnter"].(func()); ok {
		inst.OnMouseEnter = onMouseEnter
	}

	// Extract onMouseLeave from props
	if onMouseLeave, ok := fiber.Props["onMouseLeave"].(func()); ok {
		inst.OnMouseLeave = onMouseLeave
	}

	// Extract onKeyPress from props
	if onKeyPress, ok := fiber.Props["onKeyPress"].(func(string)); ok {
		inst.OnKeyPress = onKeyPress
	}
}

// GetState returns the component state
func (inst *VNodeComponentInstance) GetState() map[string]interface{} {
	return inst.State
}

// SetState updates a state value
func (inst *VNodeComponentInstance) SetState(key string, value interface{}) {
	inst.State[key] = value
	inst.MarkDirty()
}

// Render implements ComponentInstance
func (inst *VNodeComponentInstance) Render() ui.VNode {
	return inst.LatestVNode
}

// OnUpdate implements ComponentInstance
func (inst *VNodeComponentInstance) OnUpdate(newProps, oldProps ui.Props) bool {
	return true
}

// OnUnmount implements ComponentInstance
func (inst *VNodeComponentInstance) OnUnmount() {
	inst.BaseComponentInstance.OnUnmount()
	inst.OnClick = nil
	inst.OnMouseEnter = nil
	inst.OnMouseLeave = nil
	inst.OnKeyPress = nil
	inst.State = nil
}
