package space

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Space VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Space builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Direction(direction Direction) *Builder {
	b.node.SetDirection(direction)
	return b
}

func (b *Builder) Horizontal() *Builder {
	b.node.SetDirection(DirectionHorizontal)
	return b
}

func (b *Builder) Vertical() *Builder {
	b.node.SetDirection(DirectionVertical)
	return b
}

func (b *Builder) Size(size int) *Builder {
	b.node.SetSize(size)
	return b
}

func (b *Builder) Small() *Builder {
	return b.Size(SizeSmall)
}

func (b *Builder) Middle() *Builder {
	return b.Size(SizeMiddle)
}

func (b *Builder) Large() *Builder {
	return b.Size(SizeLarge)
}

func (b *Builder) Wrap(enabled bool) *Builder {
	b.node.SetWrap(enabled)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Align(align Align) *Builder {
	b.node.SetAlign(align)
	return b
}

func (b *Builder) Split(split string) *Builder {
	b.node.SetSplit(split)
	return b
}

func (b *Builder) Children(children ...rtui.VNode) *Builder {
	b.node.SetChildrenList(children)
	return b
}

func (b *Builder) AddChild(child rtui.VNode) *Builder {
	b.node.AddChild(child)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyleProps(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildVNode() *VNode {
	return b.node
}

// Of creates a horizontal Space directly from children.
func Of(children ...rtui.VNode) *VNode {
	return New().SetChildrenList(children)
}
