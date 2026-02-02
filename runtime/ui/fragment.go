package ui

import "github.com/wwsheng009/mint/runtime/style"

// FragmentVNode represents a fragment that doesn't add extra nodes
type FragmentVNode struct {
	key      string
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

// SetProps implements VNode
func (f *FragmentVNode) SetProps(p Props) {
	f.props = p
}

// Children implements VNode
func (f *FragmentVNode) Children() []VNode {
	return f.children
}

// SetChildren implements VNode
func (f *FragmentVNode) SetChildren(children []VNode) {
	f.children = children
}

// Key implements VNode
func (f *FragmentVNode) Key() string {
	return f.key
}

// SetKey implements VNode
func (f *FragmentVNode) SetKey(key string) {
	f.key = key
}

// Style implements VNode
func (f *FragmentVNode) Style() style.Style {
	return f.style
}

// SetStyle implements VNode
func (f *FragmentVNode) SetStyle(s style.Style) {
	f.style = s
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
