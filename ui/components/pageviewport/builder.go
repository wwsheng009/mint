package pageviewport

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for PageViewport.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a PageViewport builder.
func NewBuilder() *Builder {
	return &Builder{vnode: New(nil)}
}

func (b *Builder) Key(key string) *Builder          { b.vnode.SetKey(key); return b }
func (b *Builder) SetID(id string) *Builder         { b.vnode.SetID(id); return b }
func (b *Builder) Child(child rtui.VNode) *Builder  { b.vnode.SetChild(child); return b }
func (b *Builder) Width(width int) *Builder         { b.vnode.SetWidth(width); return b }
func (b *Builder) Height(height int) *Builder       { b.vnode.SetHeight(height); return b }
func (b *Builder) Size(width, height int) *Builder  { return b.Width(width).Height(height) }
func (b *Builder) ScrollOffset(offset int) *Builder { b.vnode.SetScrollOffset(offset); return b }
func (b *Builder) ShowIndicator(show bool) *Builder { b.vnode.SetShowIndicator(show); return b }
func (b *Builder) Style(s style.Style) *Builder     { b.vnode.SetStyle(s); return b }

// Build returns the configured PageViewport VNode.
func (b *Builder) Build() rtui.VNode { return b.vnode }

// BuildVNode returns the concrete VNode.
func (b *Builder) BuildVNode() *VNode { return b.vnode }

// Of creates a PageViewport around child.
func Of(child rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Child(child).Size(width, height).Build()
}
