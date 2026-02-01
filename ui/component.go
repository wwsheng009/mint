package ui

import "github.com/wwsheng009/mint/runtime/style"

// ComponentVNode represents a function component
type ComponentVNode struct {
	name      string
	fn        ComponentFunc
	fnWithProps ComponentFuncWithProps
	props     Props
	key       string
	style     style.Style
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

// SetProps implements VNode
func (c *ComponentVNode) SetProps(p Props) {
	c.props = p
}

// Children implements VNode (component children are rendered by the function)
func (c *ComponentVNode) Children() []VNode {
	// Components render their children dynamically
	return nil
}

// SetChildren implements VNode
func (c *ComponentVNode) SetChildren(children []VNode) {
	// Components don't have static children
}

// Key implements VNode
func (c *ComponentVNode) Key() string {
	return c.key
}

// SetKey implements VNode
func (c *ComponentVNode) SetKey(key string) {
	c.key = key
}

// Style implements VNode
func (c *ComponentVNode) Style() style.Style {
	return c.style
}

// SetStyle implements VNode
func (c *ComponentVNode) SetStyle(s style.Style) {
	c.style = s
}

// Name returns the component name
func (c *ComponentVNode) Name() string {
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

// Fn returns the component function (exported for Fiber)
func (c *ComponentVNode) Fn() ComponentFunc {
	return c.fn
}

// FnWithProps returns the component function with props (exported for Fiber)
func (c *ComponentVNode) FnWithProps() ComponentFuncWithProps {
	return c.fnWithProps
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

// ComponentBuilder provides fluent API for building components
type ComponentBuilder struct {
	node *ComponentVNode
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
