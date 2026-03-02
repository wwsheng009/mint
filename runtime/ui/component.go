package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
)

// ComponentVNode represents a function component
type ComponentVNode struct {
	name       string
	fn         ComponentFunc
	fnWithProps ComponentFuncWithProps
	props      Props
	key        string
	id         string
	style      style.Style
}

// NewComponent creates a new component VNode
func NewComponent(name string, fn ComponentFunc) *ComponentVNode {
	return &ComponentVNode{
		name:  name,
		fn:    fn,
		props: make(Props),
		style: style.Style{},
	}
}

// NewComponentWithProps creates a new component VNode that accepts props
func NewComponentWithProps(name string, fn ComponentFuncWithProps) *ComponentVNode {
	return &ComponentVNode{
		name:       name,
		fnWithProps: fn,
		props:      make(Props),
		style:      style.Style{},
	}
}

// Type implements VNode
func (c *ComponentVNode) Type() VNodeType {
	return VNodeComponent
}

// Props implements VNode
func (c *ComponentVNode) Props() Props {
	return c.props
}

// SetProps implements VNode - returns VNode for chaining
func (c *ComponentVNode) SetProps(p Props) VNode {
	c.props = p
	return c
}

// Children implements VNode (component children are rendered by the function)
func (c *ComponentVNode) Children() []VNode {
	// Components render their children dynamically
	return nil
}

// SetChildren implements VNode - returns VNode for chaining
func (c *ComponentVNode) SetChildren(children []VNode) VNode {
	// Components don't have static children
	return c
}

// Key implements VNode
func (c *ComponentVNode) Key() string {
	return c.key
}

// SetKey implements VNode - returns VNode for chaining
func (c *ComponentVNode) SetKey(key string) VNode {
	c.key = key
	return c
}

// ID implements VNode - returns the business identifier for component reference/positioning
func (c *ComponentVNode) ID() string {
	return c.id
}

// SetID implements VNode - sets the business identifier and returns VNode for chaining
func (c *ComponentVNode) SetID(id string) VNode {
	c.id = id
	return c
}

// Style implements VNode
func (c *ComponentVNode) Style() style.Style {
	return c.style
}

// SetStyle implements VNode - returns VNode for chaining
func (c *ComponentVNode) SetStyle(s style.Style) VNode {
	c.style = s
	return c
}

// Name returns the component name
func (c *ComponentVNode) Name() string {
	return c.name
}

// Tag returns the component name (implements VNode interface)
func (c *ComponentVNode) Tag() string {
	return c.name
}

// GetLayer returns the rendering layer from props
func (c *ComponentVNode) GetLayer() Layer {
	return getNodeLayer(c)
}

// SetLayer sets the rendering layer in props
func (c *ComponentVNode) SetLayer(l Layer) VNode {
	return setNodeLayer(c, l)
}

// =============================================================================
// Portal Methods - Chainable methods for Portal configuration
// =============================================================================

// SetPortalRoot implements VNode - sets the portalRoot property
func (c *ComponentVNode) SetPortalRoot(portalRootID string) VNode {
	if c.props == nil {
		c.props = make(Props)
	}
	c.props["portalRoot"] = portalRootID
	return c
}

// SetAnchorTo implements VNode - sets anchorId and anchor properties
func (c *ComponentVNode) SetAnchorTo(anchorID string, anchor types.Anchor) VNode {
	if c.props == nil {
		c.props = make(Props)
	}
	c.props["anchorId"] = anchorID
	c.props["anchor"] = anchor
	return c
}

// SetPortalPosition implements VNode - sets the position property
func (c *ComponentVNode) SetPortalPosition(position types.PositionType) VNode {
	if c.props == nil {
		c.props = make(Props)
	}
	c.props["position"] = position
	return c
}

// SetPortalPriority implements VNode - sets the priority property
func (c *ComponentVNode) SetPortalPriority(priority int) VNode {
	if c.props == nil {
		c.props = make(Props)
	}
	c.props["priority"] = priority
	return c
}

// SetPortalRootId implements VNode - sets the portalRootId property
func (c *ComponentVNode) SetPortalRootId(portalRootId string) VNode {
	if c.props == nil {
		c.props = make(Props)
	}
	c.props["portalRootId"] = portalRootId
	return c
}

// Render calls the component function to get the rendered VNode
func (c *ComponentVNode) Render() VNode {
	if c.fn != nil {
		return c.fn()
	}
	if c.fnWithProps != nil {
		return c.fnWithProps(c.props)
	}
	return nil
}

// Fn returns the component function
func (c *ComponentVNode) Fn() ComponentFunc {
	return c.fn
}

// FnWithProps returns the component function with props
func (c *ComponentVNode) FnWithProps() ComponentFuncWithProps {
	return c.fnWithProps
}

// ComponentBuilder provides fluent API for building components
type ComponentBuilder struct {
	node *ComponentVNode
}

// Component creates a new component builder
func Component(name string, fn ComponentFunc) *ComponentBuilder {
	return &ComponentBuilder{
		node: NewComponent(name, fn),
	}
}

// ComponentWithProps creates a component builder with props
func ComponentWithProps(name string, fn ComponentFuncWithProps) *ComponentBuilder {
	return &ComponentBuilder{
		node: NewComponentWithProps(name, fn),
	}
}

// Props sets the component props
func (b *ComponentBuilder) Props(p Props) *ComponentBuilder {
	b.node.props = p
	return b
}

// Prop sets a single prop
func (b *ComponentBuilder) Prop(key string, value interface{}) *ComponentBuilder {
	b.node.props[key] = value
	return b
}

// Key sets the key for diffing
func (b *ComponentBuilder) Key(key string) *ComponentBuilder {
	b.node.SetKey(key)
	return b
}

// Build returns the VNode
func (b *ComponentBuilder) Build() VNode {
	return b.node
}
