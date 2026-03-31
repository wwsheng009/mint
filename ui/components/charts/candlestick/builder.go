package candlestick

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating candlestick VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new candlestick builder.
func NewBuilder(candles []Candle) *Builder {
	return &Builder{node: New(candles)}
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

func (b *Builder) Candles(candles []Candle) *Builder {
	b.node.SetCandles(candles)
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

func (b *Builder) ShowGrid(show bool) *Builder {
	b.node.SetShowGrid(show)
	return b
}

func (b *Builder) ShowLegend(show bool) *Builder {
	b.node.SetShowLegend(show)
	return b
}

func (b *Builder) ShowVolume(show bool) *Builder {
	b.node.SetShowVolume(show)
	return b
}

func (b *Builder) VolumeHeight(height int) *Builder {
	b.node.SetVolumeHeight(height)
	return b
}

func (b *Builder) UpStyle(s style.Style) *Builder {
	b.node.SetUpStyle(s)
	return b
}

func (b *Builder) DownStyle(s style.Style) *Builder {
	b.node.SetDownStyle(s)
	return b
}

func (b *Builder) FlatStyle(s style.Style) *Builder {
	b.node.SetFlatStyle(s)
	return b
}

func (b *Builder) WickStyle(s style.Style) *Builder {
	b.node.SetWickStyle(s)
	return b
}

func (b *Builder) VolumeStyle(s style.Style) *Builder {
	b.node.SetVolumeStyle(s)
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
