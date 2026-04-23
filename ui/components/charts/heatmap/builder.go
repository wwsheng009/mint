package heatmap

import (
	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating heatmap VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new heatmap builder.
func NewBuilder(values [][]float64) *Builder {
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

func (b *Builder) RowLabels(labels []string) *Builder {
	b.node.SetRowLabels(labels)
	return b
}

func (b *Builder) ColLabels(labels []string) *Builder {
	b.node.SetColLabels(labels)
	return b
}

func (b *Builder) Values(values [][]float64) *Builder {
	b.node.SetValues(values)
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

func (b *Builder) ShowSummary(show bool) *Builder {
	b.node.SetShowSummary(show)
	return b
}

func (b *Builder) SummaryMode(mode SummaryMode) *Builder {
	b.node.SetSummaryMode(mode)
	return b
}

func (b *Builder) CompactSummary() *Builder {
	b.node.SetSummaryMode(SummaryModeCompact)
	return b
}

func (b *Builder) DetailedSummary() *Builder {
	b.node.SetSummaryMode(SummaryModeDetailed)
	return b
}

func (b *Builder) LegendMode(mode LegendMode) *Builder {
	b.node.SetLegendMode(mode)
	return b
}

func (b *Builder) FullLegend() *Builder {
	b.node.SetLegendMode(LegendModeFull)
	return b
}

func (b *Builder) CompactLegend() *Builder {
	b.node.SetLegendMode(LegendModeCompact)
	return b
}

func (b *Builder) ColorMode(mode fwtheme.ColorMode) *Builder {
	b.node.SetColorMode(mode)
	return b
}

func (b *Builder) ScaleMode(mode ScaleMode) *Builder {
	b.node.SetScaleMode(mode)
	return b
}

func (b *Builder) GlobalScale() *Builder {
	b.node.SetScaleMode(ScaleModeGlobal)
	return b
}

func (b *Builder) ViewportScale() *Builder {
	b.node.SetScaleMode(ScaleModeViewport)
	return b
}

func (b *Builder) AutoScale() *Builder {
	b.node.SetScaleMode(ScaleModeAuto)
	return b
}

func (b *Builder) Viewport(viewport Viewport) *Builder {
	b.node.SetViewport(viewport)
	return b
}

func (b *Builder) RowWindow(start, count int) *Builder {
	viewport := b.node.Viewport()
	viewport.RowStart = start
	viewport.RowCount = count
	b.node.SetViewport(viewport)
	return b
}

func (b *Builder) ColWindow(start, count int) *Builder {
	viewport := b.node.Viewport()
	viewport.ColStart = start
	viewport.ColCount = count
	b.node.SetViewport(viewport)
	return b
}

func (b *Builder) MaxRowLabelWidth(width int) *Builder {
	b.node.SetMaxRowLabelWidth(width)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetHeatmapStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}
