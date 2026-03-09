package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
)

// FragmentVNode represents a fragment that doesn't add extra nodes
type FragmentVNode struct {
	key      string
	id       string
	children []VNode
	props    Props
	style    style.Style
}

// NewFragment creates a new fragment VNode
func NewFragment(children ...VNode) *FragmentVNode {
	return &FragmentVNode{
		children: children,
		props:    make(Props),
		style:    style.Style{},
	}
}

// Type implements VNode
func (f *FragmentVNode) Type() VNodeType {
	return VNodeFragment
}

// Props implements VNode
func (f *FragmentVNode) Props() Props {
	return f.props
}

// SetProps implements VNode - returns VNode for chaining
func (f *FragmentVNode) SetProps(p Props) VNode {
	f.props = p
	return f
}

// SetProp sets a single property - returns VNode for chaining (implements VNode interface)
func (f *FragmentVNode) SetProp(key string, value interface{}) VNode {
	if f.props == nil {
		f.props = make(Props)
	}
	f.props[key] = value
	return f
}

// Children implements VNode
func (f *FragmentVNode) Children() []VNode {
	return f.children
}

// SetChildren implements VNode - returns VNode for chaining
func (f *FragmentVNode) SetChildren(children []VNode) VNode {
	f.children = children
	return f
}

// Key implements VNode
func (f *FragmentVNode) Key() string {
	return f.key
}

// SetKey implements VNode - returns VNode for chaining
func (f *FragmentVNode) SetKey(key string) VNode {
	f.key = key
	return f
}

// ID implements VNode - returns the business identifier for fragment reference/positioning
func (f *FragmentVNode) ID() string {
	return f.id
}

// SetID implements VNode - sets the business identifier and returns VNode for chaining
func (f *FragmentVNode) SetID(id string) VNode {
	f.id = id
	return f
}

// Style implements VNode
func (f *FragmentVNode) Style() style.Style {
	return f.style
}

// SetStyle implements VNode - returns VNode for chaining
func (f *FragmentVNode) SetStyle(s style.Style) VNode {
	f.style = s
	return f
}

// Tag returns "fragment" (implements VNode interface)
func (f *FragmentVNode) Tag() string {
	return "fragment"
}

// GetLayer returns the rendering layer
func (f *FragmentVNode) GetLayer() Layer {
	return getNodeLayer(f)
}

// SetLayer sets the rendering layer - returns VNode for chaining
func (f *FragmentVNode) SetLayer(l Layer) VNode {
	return setNodeLayer(f, l)
}

// =============================================================================
// Portal Methods - Chainable methods for Portal configuration
// =============================================================================

// SetPortalRoot implements VNode - sets the portalRoot property
func (f *FragmentVNode) SetPortalRoot(portalRootID string) VNode {
	if f.props == nil {
		f.props = make(Props)
	}
	f.props["portalRoot"] = portalRootID
	return f
}

// SetAnchorTo implements VNode - sets anchorId and anchor properties
func (f *FragmentVNode) SetAnchorTo(anchorID string, anchor types.Anchor) VNode {
	if f.props == nil {
		f.props = make(Props)
	}
	f.props["anchorId"] = anchorID
	f.props["anchor"] = anchor
	return f
}

// SetPortalPosition implements VNode - sets the position property
func (f *FragmentVNode) SetPortalPosition(position types.PositionType) VNode {
	if f.props == nil {
		f.props = make(Props)
	}
	f.props["position"] = position
	return f
}

// SetPortalPriority implements VNode - sets the priority property
func (f *FragmentVNode) SetPortalPriority(priority int) VNode {
	if f.props == nil {
		f.props = make(Props)
	}
	f.props["priority"] = priority
	return f
}

// SetPortalRootId implements VNode - sets the portalRootId property
func (f *FragmentVNode) SetPortalRootId(portalRootId string) VNode {
	if f.props == nil {
		f.props = make(Props)
	}
	f.props["portalRootId"] = portalRootId
	return f
}

// AddChild adds a single child
func (f *FragmentVNode) AddChild(child VNode) *FragmentVNode {
	f.children = append(f.children, child)
	return f
}

// AddChildren adds multiple children
func (f *FragmentVNode) AddChildren(children ...VNode) *FragmentVNode {
	f.children = append(f.children, children...)
	return f
}

// Fragment creates a new fragment with the given children
func Fragment(children ...VNode) VNode {
	return NewFragment(children...)
}
