package barchart

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating bar chart VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new bar chart builder.
func NewBuilder(values []float64) *Builder {
	return &Builder{node: New(values)}
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

func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

func (b *Builder) Labels(labels []string) *Builder {
	b.node.SetLabels(labels)
	return b
}

func (b *Builder) Values(values []float64) *Builder {
	b.node.SetValues(values)
	return b
}

func (b *Builder) Series(series ...Series) *Builder {
	b.node.SetSeries(series)
	return b
}

func (b *Builder) Mode(mode Mode) *Builder {
	b.node.SetMode(mode)
	return b
}

func (b *Builder) Orientation(orientation Orientation) *Builder {
	b.node.SetOrientation(orientation)
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

func (b *Builder) ShowAxis(show bool) *Builder {
	b.node.SetShowAxis(show)
	return b
}

func (b *Builder) ShowLegend(show bool) *Builder {
	b.node.SetShowLegend(show)
	return b
}

func (b *Builder) ShowValue(show bool) *Builder {
	b.node.SetShowValue(show)
	return b
}

func (b *Builder) Grouped() *Builder {
	b.node.Grouped()
	return b
}

func (b *Builder) Stacked() *Builder {
	b.node.Stacked()
	return b
}

func (b *Builder) Vertical() *Builder {
	b.node.Vertical()
	return b
}

func (b *Builder) Horizontal() *Builder {
	b.node.Horizontal()
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetChartStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}
