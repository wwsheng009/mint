package scatterplot

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey        = "key"
	propTitle      = "title"
	propPoints     = "points"
	propSeries     = "series"
	propSeriesName = "seriesName"
	propWidth      = "width"
	propHeight     = "height"
	propDomain     = "domain"
	propViewport   = "viewport"
	propXRefs      = "xReferenceLines"
	propYRefs      = "yReferenceLines"
	propXRefDefs   = "xReferenceLineDefs"
	propYRefDefs   = "yReferenceLineDefs"
	propXBands     = "xReferenceBands"
	propYBands     = "yReferenceBands"
	propShowAxis   = "showAxis"
	propShowGrid   = "showGrid"
	propShowLegend = "showLegend"
	propStyle      = "style"
)

// VNode is the declarative description of a scatter plot component.
type VNode struct {
	*rtui.ElementVNode

	key        string
	title      string
	points     []Point
	series     []Series
	seriesName string
	width      int
	height     int
	domain     Domain
	viewport   Viewport
	xRefs      []ReferenceLine
	yRefs      []ReferenceLine
	xBands     []ReferenceBand
	yBands     []ReferenceBand
	showAxis   bool
	showGrid   bool
	showLegend bool
	chartStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new scatter plot VNode.
func New(points []Point) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("scatterplot"),
		points:       copyPointSlice(points),
		showAxis:     true,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "scatterplot" }
func (v *VNode) Style() style.Style                           { return v.chartStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.chartStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:        v.key,
		propTitle:      v.title,
		propPoints:     copyPointSlice(v.points),
		propSeries:     copySeriesSlice(v.series),
		propSeriesName: v.seriesName,
		propWidth:      v.width,
		propHeight:     v.height,
		propDomain:     v.domain,
		propViewport:   v.viewport,
		propXRefs:      referenceLineValues(v.xRefs),
		propYRefs:      referenceLineValues(v.yRefs),
		propXRefDefs:   copyReferenceLineSlice(v.xRefs),
		propYRefDefs:   copyReferenceLineSlice(v.yRefs),
		propXBands:     copyReferenceBandSlice(v.xBands),
		propYBands:     copyReferenceBandSlice(v.yBands),
		propShowAxis:   v.showAxis,
		propShowGrid:   v.showGrid,
		propShowLegend: v.showLegend,
		propStyle:      v.chartStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if points, ok := props[propPoints].([]Point); ok {
		v.points = copyPointSlice(points)
	}
	if series, ok := props[propSeries].([]Series); ok {
		v.series = copySeriesSlice(series)
	}
	if seriesName, ok := props[propSeriesName].(string); ok {
		v.seriesName = seriesName
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if height, ok := props[propHeight].(int); ok {
		v.height = height
	}
	if domain, ok := props[propDomain].(Domain); ok {
		v.domain = normalizeDomainSpec(domain)
	}
	if viewport, ok := props[propViewport].(Viewport); ok {
		v.viewport = normalizeViewportSpec(viewport)
	}
	if refs, ok := props[propXRefs].([]float64); ok {
		v.xRefs = referenceLinesFromValues(refs)
	}
	if refs, ok := props[propYRefs].([]float64); ok {
		v.yRefs = referenceLinesFromValues(refs)
	}
	if refs, ok := props[propXRefDefs].([]ReferenceLine); ok {
		v.xRefs = copyReferenceLineSlice(refs)
	}
	if refs, ok := props[propYRefDefs].([]ReferenceLine); ok {
		v.yRefs = copyReferenceLineSlice(refs)
	}
	if bands, ok := props[propXBands].([]ReferenceBand); ok {
		v.xBands = copyReferenceBandSlice(bands)
	}
	if bands, ok := props[propYBands].([]ReferenceBand); ok {
		v.yBands = copyReferenceBandSlice(bands)
	}
	if showAxis, ok := props[propShowAxis].(bool); ok {
		v.showAxis = showAxis
	}
	if showGrid, ok := props[propShowGrid].(bool); ok {
		v.showGrid = showGrid
	}
	if showLegend, ok := props[propShowLegend].(bool); ok {
		v.showLegend = showLegend
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.chartStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetPoints(points []Point) *VNode {
	v.points = copyPointSlice(points)
	return v
}

func (v *VNode) SetSeries(series []Series) *VNode {
	v.series = copySeriesSlice(series)
	return v
}

func (v *VNode) SetSeriesName(name string) *VNode {
	v.seriesName = name
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetHeight(height int) *VNode {
	v.height = height
	return v
}

func (v *VNode) SetDomain(domain Domain) *VNode {
	v.domain = normalizeDomainSpec(domain)
	return v
}

func (v *VNode) SetViewport(viewport Viewport) *VNode {
	v.viewport = normalizeViewportSpec(viewport)
	return v
}

func (v *VNode) SetXDomain(minX, maxX float64) *VNode {
	domain := normalizeDomainSpec(v.domain)
	domain.MinX = minX
	domain.MaxX = maxX
	domain.HasX = true
	v.domain = normalizeDomainSpec(domain)
	return v
}

func (v *VNode) SetXViewport(minX, maxX float64) *VNode {
	viewport := normalizeViewportSpec(v.viewport)
	viewport.MinX = minX
	viewport.MaxX = maxX
	viewport.HasX = true
	v.viewport = normalizeViewportSpec(viewport)
	return v
}

func (v *VNode) SetYDomain(minY, maxY float64) *VNode {
	domain := normalizeDomainSpec(v.domain)
	domain.MinY = minY
	domain.MaxY = maxY
	domain.HasY = true
	v.domain = normalizeDomainSpec(domain)
	return v
}

func (v *VNode) SetYViewport(minY, maxY float64) *VNode {
	viewport := normalizeViewportSpec(v.viewport)
	viewport.MinY = minY
	viewport.MaxY = maxY
	viewport.HasY = true
	v.viewport = normalizeViewportSpec(viewport)
	return v
}

func (v *VNode) SetXReferenceLines(refs []float64) *VNode {
	v.xRefs = referenceLinesFromValues(refs)
	return v
}

func (v *VNode) SetYReferenceLines(refs []float64) *VNode {
	v.yRefs = referenceLinesFromValues(refs)
	return v
}

func (v *VNode) SetXReferenceLineDefs(refs []ReferenceLine) *VNode {
	v.xRefs = copyReferenceLineSlice(refs)
	return v
}

func (v *VNode) SetYReferenceLineDefs(refs []ReferenceLine) *VNode {
	v.yRefs = copyReferenceLineSlice(refs)
	return v
}

func (v *VNode) SetXReferenceBands(bands []ReferenceBand) *VNode {
	v.xBands = copyReferenceBandSlice(bands)
	return v
}

func (v *VNode) SetYReferenceBands(bands []ReferenceBand) *VNode {
	v.yBands = copyReferenceBandSlice(bands)
	return v
}

func (v *VNode) SetShowAxis(show bool) *VNode {
	v.showAxis = show
	return v
}

func (v *VNode) SetShowGrid(show bool) *VNode {
	v.showGrid = show
	return v
}

func (v *VNode) SetShowLegend(show bool) *VNode {
	v.showLegend = show
	return v
}

func (v *VNode) SetChartStyle(s style.Style) *VNode {
	v.chartStyle = s
	return v
}

func (v *VNode) Title() string              { return v.title }
func (v *VNode) Points() []Point            { return copyPointSlice(v.points) }
func (v *VNode) Series() []Series           { return copySeriesSlice(v.series) }
func (v *VNode) SeriesName() string         { return v.seriesName }
func (v *VNode) Width() int                 { return v.width }
func (v *VNode) Height() int                { return v.height }
func (v *VNode) Domain() Domain             { return v.domain }
func (v *VNode) Viewport() Viewport         { return v.viewport }
func (v *VNode) XReferenceLines() []float64 { return referenceLineValues(v.xRefs) }
func (v *VNode) YReferenceLines() []float64 { return referenceLineValues(v.yRefs) }
func (v *VNode) XReferenceLineDefs() []ReferenceLine {
	return copyReferenceLineSlice(v.xRefs)
}
func (v *VNode) YReferenceLineDefs() []ReferenceLine {
	return copyReferenceLineSlice(v.yRefs)
}
func (v *VNode) XReferenceBands() []ReferenceBand {
	return copyReferenceBandSlice(v.xBands)
}
func (v *VNode) YReferenceBands() []ReferenceBand {
	return copyReferenceBandSlice(v.yBands)
}
func (v *VNode) ShowAxis() bool          { return v.showAxis }
func (v *VNode) ShowGrid() bool          { return v.showGrid }
func (v *VNode) ShowLegend() bool        { return v.showLegend }
func (v *VNode) ChartStyle() style.Style { return v.chartStyle }
