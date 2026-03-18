package skeleton

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Skeleton VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Skeleton builder.
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

func (b *Builder) Content(content rtui.VNode) *Builder {
	b.node.SetContent(content)
	return b
}

func (b *Builder) Loading(loading bool) *Builder {
	b.node.SetLoading(loading)
	return b
}

func (b *Builder) Active(active bool) *Builder {
	b.node.SetActive(active)
	return b
}

func (b *Builder) Avatar(show bool) *Builder {
	b.node.SetAvatar(show)
	return b
}

func (b *Builder) AvatarShape(shape Shape) *Builder {
	b.node.SetAvatarShape(shape)
	return b
}

func (b *Builder) AvatarSize(size int) *Builder {
	b.node.SetAvatarSize(size)
	return b
}

func (b *Builder) Title(show bool) *Builder {
	b.node.SetTitle(show)
	return b
}

func (b *Builder) TitleWidth(width int) *Builder {
	b.node.SetTitleWidth(width)
	return b
}

func (b *Builder) Paragraph(show bool) *Builder {
	b.node.SetParagraph(show)
	return b
}

func (b *Builder) ParagraphRows(rows int) *Builder {
	b.node.SetParagraphRows(rows)
	return b
}

func (b *Builder) ParagraphWidths(widths ...int) *Builder {
	b.node.SetParagraphWidths(widths...)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Gap(gap int) *Builder {
	b.node.SetGap(gap)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) PlaceholderStyle(s style.Style) *Builder {
	b.node.SetPlaceholderStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildVNode() *VNode {
	return b.node
}
