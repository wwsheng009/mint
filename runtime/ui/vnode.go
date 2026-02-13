// Package types provides core type definitions for the declarative UI framework.
// These types are shared between the UI layer and the internal reconciler.
package ui

import "github.com/wwsheng009/mint/runtime/style"

// =============================================================================
// Layer Type
// =============================================================================

// Layer represents a visual rendering layer for overlay components
type Layer int

const (
	// LayerBase is the default layer for normal UI content
	LayerBase Layer = iota

	// LayerOverlay is for dropdown menus, popovers, and similar components
	LayerOverlay

	// LayerModal is for modal dialogs that require user attention
	LayerModal

	// LayerTooltip is for tooltips and hints
	LayerTooltip

	// LayerInspector is for the UI Inspector debugging overlay
	LayerInspector
)

// String returns the string representation of the layer
func (l Layer) String() string {
	switch l {
	case LayerBase:
		return "base"
	case LayerOverlay:
		return "overlay"
	case LayerModal:
		return "modal"
	case LayerTooltip:
		return "tooltip"
	case LayerInspector:
		return "inspector"
	default:
		return "unknown"
	}
}

// ZIndex returns the z-index value for this layer (higher = rendered on top)
func (l Layer) ZIndex() int {
	return int(l)
}

// IsValid checks if the layer value is valid
func (l Layer) IsValid() bool {
	return l >= LayerBase && l <= LayerInspector
}

// IsModal checks if this layer is the modal layer
func (l Layer) IsModal() bool {
	return l == LayerModal
}

// IsOverlay checks if this layer is any overlay type (Overlay, Modal, or Tooltip)
func (l Layer) IsOverlay() bool {
	return l >= LayerOverlay
}

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
	SetProps(p Props)

	// Children returns the child nodes
	Children() []VNode
	SetChildren(children []VNode)

	// Key returns the key for list diffing
	Key() string
	// SetKey sets the key for Component
	// fiber reconciler 调用 vnode.SetKey(fiber.Path)
	SetKey(key string)

	// Style returns the visual style
	Style() style.Style
	SetStyle(s style.Style)

	// Tag returns the tag/name of this node for identification
	// For elements: the HTML-like tag (e.g., "div", "button")
	// For components: the component name (e.g., "MyComponent")
	// For fragments: "fragment"
	// For text/layout nodes: the type identifier (e.g., "text", "hstack")
	Tag() string

	// GetLayer returns the rendering layer for this node
	// Returns LayerBase if no layer is explicitly set
	GetLayer() Layer

	// SetLayer sets the rendering layer for this node
	SetLayer(l Layer) VNode
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

// GetLayer for ElementVNode
func (n *ElementVNode) GetLayer() Layer {
	return getNodeLayer(n)
}

// SetLayer for ElementVNode
func (n *ElementVNode) SetLayer(l Layer) VNode {
	return setNodeLayer(n, l)
}

// GetLayer for TextVNode
func (n *TextVNode) GetLayer() Layer {
	return getNodeLayer(n)
}

// SetLayer for TextVNode
func (n *TextVNode) SetLayer(l Layer) VNode {
	return setNodeLayer(n, l)
}

// GetLayer for ComponentVNode
func (n *ComponentVNode) GetLayer() Layer {
	return getNodeLayer(n)
}

// SetLayer for ComponentVNode
func (n *ComponentVNode) SetLayer(l Layer) VNode {
	return setNodeLayer(n, l)
}

// GetLayer for FragmentVNode
func (n *FragmentVNode) GetLayer() Layer {
	return getNodeLayer(n)
}

// SetLayer for FragmentVNode
func (n *FragmentVNode) SetLayer(l Layer) VNode {
	return setNodeLayer(n, l)
}
