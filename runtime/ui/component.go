package ui

import "github.com/wwsheng009/mint/runtime/style"

// ComponentVNode represents a function component
type ComponentVNode struct {
	name       string
	fn         ComponentFunc
	fnWithProps ComponentFuncWithProps
	props      Props
	key        string
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
