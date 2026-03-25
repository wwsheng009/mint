package layout

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Layout VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Layout builder.
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

func (b *Builder) Header(header rtui.VNode) *Builder {
	b.node.SetHeader(header)
	return b
}

func (b *Builder) Sider(sider rtui.VNode) *Builder {
	b.node.SetSider(sider)
	return b
}

func (b *Builder) LeftSider(sider rtui.VNode) *Builder {
	b.node.SetLeftSider(sider)
	return b
}

func (b *Builder) Content(content rtui.VNode) *Builder {
	b.node.SetContent(content)
	return b
}

func (b *Builder) RightSider(sider rtui.VNode) *Builder {
	b.node.SetRightSider(sider)
	return b
}

func (b *Builder) Footer(footer rtui.VNode) *Builder {
	b.node.SetFooter(footer)
	return b
}

func (b *Builder) Gap(gap int) *Builder {
	b.node.SetGap(gap)
	return b
}

func (b *Builder) BodyGap(gap int) *Builder {
	b.node.SetBodyGap(gap)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Height(height int) *Builder {
	b.node.SetHeight(height)
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
