package linechart

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey           = "key"
	propTitle         = "title"
	propData          = "data"
	propSeries        = "series"
	propSeriesName    = "seriesName"
	propLabels        = "labels"
	propAxisLabelMode = "axisLabelMode"
	propWidth         = "width"
	propHeight        = "height"
	propShowAxis      = "showAxis"
	propShowGrid      = "showGrid"
	propShowLegend    = "showLegend"
	propShowPoints    = "showPoints"
	propStyle         = "style"
)

// AxisLabelMode controls how x-axis labels are sampled when labels are present.
type AxisLabelMode int

const (
	AxisLabelModeAuto AxisLabelMode = iota
	AxisLabelModeDense
	AxisLabelModeSparse
)

// VNode is the declarative description of a line chart component.
type VNode struct {
	*rtui.ElementVNode

	key           string
	title         string
	data          []float64
	series        []Series
	seriesName    string
	labels        []string
	axisLabelMode AxisLabelMode
	renderBackend RenderBackend
	width         int
	height        int
	showAxis      bool
	showGrid      bool
	showLegend    bool
	showPoints    bool
	chartStyle    style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new line chart VNode.
func New(data []float64) *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("linechart"),
		data:          copyFloat64Slice(data),
		axisLabelMode: AxisLabelModeAuto,
		renderBackend: RenderBackendText,
		showAxis:      true,
		showPoints:    true,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "linechart" }
func (v *VNode) Style() style.Style                           { return v.chartStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.chartStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:           v.key,
		propTitle:         v.title,
		propData:          copyFloat64Slice(v.data),
		propSeries:        copySeriesSlice(v.series),
		propSeriesName:    v.seriesName,
		propLabels:        copyStringSlice(v.labels),
		propAxisLabelMode: v.axisLabelMode,
		propRenderBackend: v.renderBackend,
		propWidth:         v.width,
		propHeight:        v.height,
		propShowAxis:      v.showAxis,
		propShowGrid:      v.showGrid,
		propShowLegend:    v.showLegend,
		propShowPoints:    v.showPoints,
		propStyle:         v.chartStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if data, ok := props[propData].([]float64); ok {
		v.data = copyFloat64Slice(data)
	}
	if series, ok := props[propSeries].([]Series); ok {
		v.series = copySeriesSlice(series)
	}
	if seriesName, ok := props[propSeriesName].(string); ok {
		v.seriesName = seriesName
	}
	if labels, ok := props[propLabels].([]string); ok {
		v.labels = copyStringSlice(labels)
	}
	if axisLabelMode, ok := props[propAxisLabelMode].(AxisLabelMode); ok {
		v.axisLabelMode = axisLabelMode
	}
	if renderBackend, ok := props[propRenderBackend].(RenderBackend); ok {
		v.renderBackend = normalizeRenderBackend(renderBackend)
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if height, ok := props[propHeight].(int); ok {
		v.height = height
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
	if showPoints, ok := props[propShowPoints].(bool); ok {
		v.showPoints = showPoints
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

func (v *VNode) SetData(data []float64) *VNode {
	v.data = copyFloat64Slice(data)
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

func (v *VNode) SetLabels(labels []string) *VNode {
	v.labels = copyStringSlice(labels)
	return v
}

func (v *VNode) SetAxisLabelMode(mode AxisLabelMode) *VNode {
	v.axisLabelMode = mode
	return v
}

func (v *VNode) SetRenderBackend(backend RenderBackend) *VNode {
	v.renderBackend = normalizeRenderBackend(backend)
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

func (v *VNode) SetShowPoints(show bool) *VNode {
	v.showPoints = show
	return v
}

func (v *VNode) SetChartStyle(s style.Style) *VNode {
	v.chartStyle = s
	return v
}

func (v *VNode) Title() string                { return v.title }
func (v *VNode) Data() []float64              { return copyFloat64Slice(v.data) }
func (v *VNode) Series() []Series             { return copySeriesSlice(v.series) }
func (v *VNode) SeriesName() string           { return v.seriesName }
func (v *VNode) Labels() []string             { return copyStringSlice(v.labels) }
func (v *VNode) AxisLabelMode() AxisLabelMode { return v.axisLabelMode }
func (v *VNode) RenderBackend() RenderBackend { return v.renderBackend }
func (v *VNode) Width() int                   { return v.width }
func (v *VNode) Height() int                  { return v.height }
func (v *VNode) ShowAxis() bool               { return v.showAxis }
func (v *VNode) ShowGrid() bool               { return v.showGrid }
func (v *VNode) ShowLegend() bool             { return v.showLegend }
func (v *VNode) ShowPoints() bool             { return v.showPoints }
func (v *VNode) ChartStyle() style.Style      { return v.chartStyle }

func copyFloat64Slice(src []float64) []float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
}

func copyStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
