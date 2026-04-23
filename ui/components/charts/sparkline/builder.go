package sparkline

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating sparkline VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new sparkline builder.
func NewBuilder(data []float64) *Builder {
	return &Builder{node: New(data)}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Data(data []float64) *Builder {
	b.node.SetData(data)
	return b
}

func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
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

func (b *Builder) AutoHeight() *Builder {
	b.node.SetAutoHeight(true)
	return b
}

func (b *Builder) InlineLabel(label string) *Builder {
	b.node.SetInlineLabel(label)
	return b
}

func (b *Builder) HighlightMinMax(highlight bool) *Builder {
	b.node.SetHighlightMinMax(highlight)
	return b
}

func (b *Builder) Mode(mode RenderMode) *Builder {
	b.node.SetRenderMode(mode)
	return b
}

func (b *Builder) Auto() *Builder {
	b.node.Auto()
	return b
}

func (b *Builder) Braille() *Builder {
	b.node.Braille()
	return b
}

func (b *Builder) Block() *Builder {
	b.node.Block()
	return b
}

func (b *Builder) ASCII() *Builder {
	b.node.ASCII()
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetSparkStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}
