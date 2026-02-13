package ui

import "github.com/wwsheng009/mint/runtime/style"

// ElementVNode represents a standard element node
type ElementVNode struct {
	vnodeType VNodeType
	tag       string
	key       string
	props     Props
	children  []VNode
	style     style.Style
	// Layout bounds (set by layout engine for hit testing)
	bounds [4]int // [x, y, width, height]
}

// NewElement creates a new element VNode
func NewElement(tag string) *ElementVNode {
	return &ElementVNode{
		vnodeType: VNodeElement,
		tag:       tag,
		props:     make(Props),
		children:  make([]VNode, 0),
		style:     style.Style{},
	}
}

// Type implements VNode
func (e *ElementVNode) Type() VNodeType {
	return e.vnodeType
}

// Props implements VNode
func (e *ElementVNode) Props() Props {
	return e.props
}

// SetProps implements VNode
func (e *ElementVNode) SetProps(p Props) {
	e.props = p
}

// Children implements VNode
func (e *ElementVNode) Children() []VNode {
	return e.children
}

// SetChildren implements VNode
func (e *ElementVNode) SetChildren(children []VNode) {
	e.children = children
}

// Key implements VNode
func (e *ElementVNode) Key() string {
	return e.key
}

// SetKey implements VNode
// 由用户设置或是fiber回调更新
func (e *ElementVNode) SetKey(key string) {
	e.key = key
}

// Style implements VNode
func (e *ElementVNode) Style() style.Style {
	return e.style
}

// SetStyle implements VNode
func (e *ElementVNode) SetStyle(s style.Style) {
	e.style = s
}

// Tag returns the element tag name
func (e *ElementVNode) Tag() string {
	return e.tag
}

// SetBounds sets the layout bounds (called by layout engine)
// This enables hit testing for inspector hover detection
func (e *ElementVNode) SetBounds(x, y, width, height int) {
	e.bounds = [4]int{x, y, width, height}
}

// GetBounds returns the layout bounds [x, y, width, height]
func (e *ElementVNode) GetBounds() [4]int {
	return e.bounds
}

// AddChild adds a single child
func (e *ElementVNode) AddChild(child VNode) *ElementVNode {
	e.children = append(e.children, child)
	return e
}

// AddChildren adds multiple children
func (e *ElementVNode) AddChildren(children ...VNode) *ElementVNode {
	e.children = append(e.children, children...)
	return e
}

// SetProp sets a single property
func (e *ElementVNode) SetProp(key string, value interface{}) *ElementVNode {
	if e.props == nil {
		e.props = make(Props)
	}
	e.props[key] = value
	return e
}

// ElementBuilder provides fluent API for building elements
type ElementBuilder struct {
	node *ElementVNode
}

// Element creates a new element builder
func Element(tag string) *ElementBuilder {
	return &ElementBuilder{
		node: NewElement(tag),
	}
}

// Prop sets a property
func (b *ElementBuilder) Prop(key string, value interface{}) *ElementBuilder {
	b.node.SetProp(key, value)
	return b
}

// Props sets multiple properties
func (b *ElementBuilder) Props(p Props) *ElementBuilder {
	b.node.props = b.node.props.Merge(p)
	return b
}

// Key sets the key for diffing
func (b *ElementBuilder) Key(key string) *ElementBuilder {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *ElementBuilder) Style(s style.Style) *ElementBuilder {
	b.node.SetStyle(s)
	return b
}

// Children sets the child nodes
func (b *ElementBuilder) Children(children ...VNode) *ElementBuilder {
	b.node.SetChildren(children)
	return b
}

// Child adds a single child
func (b *ElementBuilder) Child(child VNode) *ElementBuilder {
	b.node.AddChild(child)
	return b
}

// Build returns the VNode
func (b *ElementBuilder) Build() VNode {
	return b.node
}
