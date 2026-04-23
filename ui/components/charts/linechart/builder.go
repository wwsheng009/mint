package linechart

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating line chart VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new line chart builder.
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

func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

func (b *Builder) Data(data []float64) *Builder {
	b.node.SetData(data)
	return b
}

func (b *Builder) Series(series ...Series) *Builder {
	b.node.SetSeries(series)
	return b
}

func (b *Builder) SeriesName(name string) *Builder {
	b.node.SetSeriesName(name)
	return b
}

func (b *Builder) Labels(labels []string) *Builder {
	b.node.SetLabels(labels)
	return b
}

func (b *Builder) AxisLabelMode(mode AxisLabelMode) *Builder {
	b.node.SetAxisLabelMode(mode)
	return b
}

func (b *Builder) RenderBackend(backend RenderBackend) *Builder {
	b.node.SetRenderBackend(backend)
	return b
}

func (b *Builder) TextBackend() *Builder {
	b.node.SetRenderBackend(RenderBackendText)
	return b
}

func (b *Builder) ImagePlotBackend() *Builder {
	// Chart image-plot integration is temporarily paused; keep the builder
	// API stable while normalizing runtime behavior back to text rendering.
	b.node.SetRenderBackend(RenderBackendImagePlot)
	return b
}

func (b *Builder) DenseAxisLabels() *Builder {
	b.node.SetAxisLabelMode(AxisLabelModeDense)
	return b
}

func (b *Builder) SparseAxisLabels() *Builder {
	b.node.SetAxisLabelMode(AxisLabelModeSparse)
	return b
}

func (b *Builder) AutoAxisLabels() *Builder {
	b.node.SetAxisLabelMode(AxisLabelModeAuto)
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

func (b *Builder) ShowPoints(show bool) *Builder {
	b.node.SetShowPoints(show)
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
