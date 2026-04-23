package scatterplot

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating scatter plot VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new scatter plot builder.
func NewBuilder(points []Point) *Builder {
	return &Builder{node: New(points)}
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

func (b *Builder) Points(points []Point) *Builder {
	b.node.SetPoints(points)
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

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Height(height int) *Builder {
	b.node.SetHeight(height)
	return b
}

func (b *Builder) Domain(domain Domain) *Builder {
	b.node.SetDomain(domain)
	return b
}

func (b *Builder) Viewport(viewport Viewport) *Builder {
	b.node.SetViewport(viewport)
	return b
}

func (b *Builder) XDomain(minX, maxX float64) *Builder {
	b.node.SetXDomain(minX, maxX)
	return b
}

func (b *Builder) XViewport(minX, maxX float64) *Builder {
	b.node.SetXViewport(minX, maxX)
	return b
}

func (b *Builder) YDomain(minY, maxY float64) *Builder {
	b.node.SetYDomain(minY, maxY)
	return b
}

func (b *Builder) YViewport(minY, maxY float64) *Builder {
	b.node.SetYViewport(minY, maxY)
	return b
}

func (b *Builder) XReferenceLines(refs ...float64) *Builder {
	b.node.SetXReferenceLines(refs)
	return b
}

func (b *Builder) XReferenceLineDefs(lines ...ReferenceLine) *Builder {
	b.node.SetXReferenceLineDefs(lines)
	return b
}

func (b *Builder) XReferenceLine(value float64) *Builder {
	b.node.SetXReferenceLineDefs([]ReferenceLine{NewReferenceLine(value)})
	return b
}

func (b *Builder) XReferenceLineLabeled(value float64, label string) *Builder {
	b.node.SetXReferenceLineDefs([]ReferenceLine{NewLabeledReferenceLine(value, label)})
	return b
}

func (b *Builder) YReferenceLines(refs ...float64) *Builder {
	b.node.SetYReferenceLines(refs)
	return b
}

func (b *Builder) YReferenceLineDefs(lines ...ReferenceLine) *Builder {
	b.node.SetYReferenceLineDefs(lines)
	return b
}

func (b *Builder) YReferenceLine(value float64) *Builder {
	b.node.SetYReferenceLineDefs([]ReferenceLine{NewReferenceLine(value)})
	return b
}

func (b *Builder) YReferenceLineLabeled(value float64, label string) *Builder {
	b.node.SetYReferenceLineDefs([]ReferenceLine{NewLabeledReferenceLine(value, label)})
	return b
}

func (b *Builder) XReferenceBands(bands ...ReferenceBand) *Builder {
	b.node.SetXReferenceBands(bands)
	return b
}

func (b *Builder) YReferenceBands(bands ...ReferenceBand) *Builder {
	b.node.SetYReferenceBands(bands)
	return b
}

func (b *Builder) XReferenceBand(minX, maxX float64) *Builder {
	b.node.SetXReferenceBands([]ReferenceBand{NewReferenceBand(minX, maxX)})
	return b
}

func (b *Builder) XReferenceBandLabeled(minX, maxX float64, label string) *Builder {
	b.node.SetXReferenceBands([]ReferenceBand{NewLabeledReferenceBand(minX, maxX, label)})
	return b
}

func (b *Builder) YReferenceBand(minY, maxY float64) *Builder {
	b.node.SetYReferenceBands([]ReferenceBand{NewReferenceBand(minY, maxY)})
	return b
}

func (b *Builder) YReferenceBandLabeled(minY, maxY float64, label string) *Builder {
	b.node.SetYReferenceBands([]ReferenceBand{NewLabeledReferenceBand(minY, maxY, label)})
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
