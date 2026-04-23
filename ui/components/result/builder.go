package result

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Result VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Result builder.
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

func (b *Builder) Status(status Status) *Builder {
	b.node.SetStatus(status)
	return b
}

func (b *Builder) Icon(icon string) *Builder {
	b.node.SetIcon(icon)
	return b
}

func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

func (b *Builder) Subtitle(subtitle string) *Builder {
	b.node.SetSubtitle(subtitle)
	return b
}

func (b *Builder) Extra(extra rtui.VNode) *Builder {
	b.node.SetExtra(extra)
	return b
}

func (b *Builder) Bordered(bordered bool) *Builder {
	b.node.SetBordered(bordered)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) IconStyle(s style.Style) *Builder {
	b.node.SetIconStyle(s)
	return b
}

func (b *Builder) TitleStyle(s style.Style) *Builder {
	b.node.SetTitleStyle(s)
	return b
}

func (b *Builder) SubtitleStyle(s style.Style) *Builder {
	b.node.SetSubtitleStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildVNode() *VNode {
	return b.node
}
