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

	// Reference to the latest VNode (updated each render)
	LatestVNode ui.VNode
}

// NewVNodeComponentInstance creates a new instance for a VNode component
func NewVNodeComponentInstance(key string, vnode ui.VNode) *VNodeComponentInstance {
	base := NewBaseComponentInstance(key, func() ui.VNode {
		// Return the latest VNode on render
		return vnode
	})

	return &VNodeComponentInstance{
		BaseComponentInstance: base,
		State:                 make(map[string]interface{}),
		LatestVNode:            vnode,
	}
}

// UpdateVNode updates the VNode reference and extracts handlers
func (inst *VNodeComponentInstance) UpdateVNode(vnode ui.VNode) {
	inst.LatestVNode = vnode

	// Extract event handlers from VNode
	inst.extractHandlers(vnode)

	// NOTE: Don't call SetProps() because props may contain functions
	// (like onClick), which are not comparable and cause panic in propsEqual
	// The handlers are already extracted above, so we don't need to update props
}

// extractHandlers extracts event handlers from VNode
// This bridges the gap between VNode (ephemeral) and Instance (persistent)
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
// Returns the latest VNode
func (inst *VNodeComponentInstance) Render() ui.VNode {
	return inst.LatestVNode
}

// OnUpdate implements ComponentInstance
// Called when props change
func (inst *VNodeComponentInstance) OnUpdate(newProps, oldProps ui.Props) bool {
	// Allow update by default
	return true
}

// OnUnmount implements ComponentInstance
// Cleanup handlers
func (inst *VNodeComponentInstance) OnUnmount() {
	inst.BaseComponentInstance.OnUnmount()
	// Clear handlers
	inst.OnClick = nil
	inst.OnMouseEnter = nil
	inst.OnMouseLeave = nil
	inst.OnKeyPress = nil
	// Clear state
	inst.State = nil
}
