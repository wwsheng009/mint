// Package types provides core type definitions for the declarative UI framework.
// These types are shared between the UI layer and the internal reconciler.
package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
)

// =============================================================================
// Layer Type - 使用 types.Layer 统一类型
// =============================================================================

// Layer 是 types.Layer 的类型别名，保持向后兼容
type Layer = types.Layer

// 层级常量 - 引用 types 包的统一常量
const (
	LayerBase      = types.LayerBase
	LayerOverlay   = types.LayerOverlay
	LayerModal     = types.LayerModal
	LayerTooltip   = types.LayerTooltip
	LayerInspector = types.LayerInspector
)

// =============================================================================
// VNode Interface
// =============================================================================

// VNode is the virtual node interface - the core of the declarative UI system.
// It represents a lightweight description of what should be rendered.
type VNode interface {
	// Type returns the node type for diff algorithm
	Type() VNodeType

	// Props returns the node properties
	Props() Props
	// SetProps sets the node properties and returns the VNode for chaining
	SetProps(p Props) VNode

	// Children returns the child nodes
	Children() []VNode
	// SetChildren sets child nodes and returns the VNode for chaining
	SetChildren(children []VNode) VNode

	// Key returns the key for list diffing
	Key() string
	// SetKey sets the key for diffing and returns the VNode for chaining
	SetKey(key string) VNode

	// ID returns the business identifier for element reference/positioning
	// This is separate from Key which is used for diffing lists
	ID() string
	// SetID sets the business identifier and returns the VNode for chaining
	SetID(id string) VNode

	// Style returns the visual style
	Style() style.Style
	// SetStyle sets the visual style and returns the VNode for chaining
	SetStyle(s style.Style) VNode

	// Tag returns the tag/name of this node for identification
	// For elements: the HTML-like tag (e.g., "div", "button")
	// For components: the component name (e.g., "MyComponent")
	// For fragments: "fragment"
	// For text/layout nodes: the type identifier (e.g., "text", "hstack")
	Tag() string

	// GetLayer returns the rendering layer for this node
	// Returns LayerBase if no layer is explicitly set
	GetLayer() Layer

	// SetLayer sets the rendering layer for this node and returns the VNode for chaining
	SetLayer(l Layer) VNode

	// =============================================================================
	// Portal Methods - Chainable methods for Portal configuration
	// =============================================================================

	// SetPortalRoot sets the portalRoot property for Portal components
	// portalRoot specifies which PortalRoot this Portal should be attached to
	SetPortalRoot(portalRootID string) VNode

	// SetAnchorTo sets the anchorId and anchor properties for Portal positioning
	// anchorId: the ID of the element to anchor to (must use SetID() on the target element)
	// anchor: how to align the portal relative to the anchor element
	SetAnchorTo(anchorID string, anchor types.Anchor) VNode

	// SetPortalPosition sets the position property for Portal positioning
	// position: the positioning scheme (relative, absolute, fixed)
	SetPortalPosition(position types.PositionType) VNode

	// SetPortalPriority sets the portal priority (z-index offset)
	// Higher priority Portals render above lower priority ones
	// Portal Z-index = 1000 + priority
	SetPortalPriority(priority int) VNode

	// SetPortalRootId sets the portalRootId property
	// This marks a node as a PortalRoot (a mounting target for Portals)
	SetPortalRootId(portalRootId string) VNode
}

// VNodeType represents the type of VNode
type VNodeType int

const (
	// VNodeElement is a standard element node (div, span, etc.)
	VNodeElement VNodeType = iota

	// VNodeText is a text node with content
	VNodeText

	// VNodeComponent is a function component
	VNodeComponent

	// VNodeFragment is a fragment that doesn't add extra DOM nodes
	VNodeFragment
)

// String returns the string representation of VNodeType
func (t VNodeType) String() string {
	switch t {
	case VNodeElement:
		return "Element"
	case VNodeText:
		return "Text"
	case VNodeComponent:
		return "Component"
	case VNodeFragment:
		return "Fragment"
	default:
		return "Unknown"
	}
}

// Props represents a map of properties for a VNode
type Props map[string]interface{}

// Get returns a property value
func (p Props) Get(key string) interface{} {
	if p == nil {
		return nil
	}
	return p[key]
}

// Set sets a property value
func (p Props) Set(key string, value interface{}) Props {
	if p == nil {
		p = make(Props)
	}
	p[key] = value
	return p
}

// GetString returns a string property
func (p Props) GetString(key string) string {
	if v, ok := p.Get(key).(string); ok {
		return v
	}
	return ""
}

// GetInt returns an int property
func (p Props) GetInt(key string) int {
	if v, ok := p.Get(key).(int); ok {
		return v
	}
	return 0
}

// GetBool returns a bool property
func (p Props) GetBool(key string) bool {
	if v, ok := p.Get(key).(bool); ok {
		return v
	}
	return false
}

// GetFunc returns a function property
func (p Props) GetFunc(key string) func() {
	if v, ok := p.Get(key).(func()); ok {
		return v
	}
	return nil
}

// Merge merges another Props into this one
func (p Props) Merge(other Props) Props {
	result := make(Props)
	for k, v := range p {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
}

// Clone creates a copy of the Props
func (p Props) Clone() Props {
	result := make(Props, len(p))
	for k, v := range p {
		result[k] = v
	}
	return result
}

// ComponentFunc represents a function component that returns a VNode
type ComponentFunc func() VNode

// ComponentFuncWithProps represents a component that accepts props
type ComponentFuncWithProps func(Props) VNode

// Buildable is an interface for VNode builders that need to be finalized
// before being used by the reconciler. Build() is called automatically
// by CreateFiber() to ensure the VNode is in its final form.
type Buildable interface {
	VNode
	// Build finalizes the builder and returns the constructed VNode
	Build() VNode
}

// =============================================================================
// Layer Methods (default implementations)
// =============================================================================

// GetLayer returns the rendering layer from props
// Default implementation for nodes that don't override it
func getNodeLayer(vnode VNode) Layer {
	if vnode == nil {
		return LayerBase
	}
	props := vnode.Props()
	if props == nil {
		return LayerBase
	}
	if l, ok := props["_layer"].(Layer); ok {
		return l
	}
	return LayerBase
}

// SetLayer sets the rendering layer in props
func setNodeLayer(vnode VNode, l Layer) VNode {
	if vnode == nil {
		return nil
	}
	if vnode.Props() == nil {
		vnode.SetProps(make(Props))
	}
	vnode.Props().Set("_layer", l)
	return vnode
}

// GetLayer for TextVNode
func (n *TextVNode) GetLayer() Layer {
	return getNodeLayer(n)
}

// SetLayer for TextVNode
func (n *TextVNode) SetLayer(l Layer) VNode {
	return setNodeLayer(n, l)
}
