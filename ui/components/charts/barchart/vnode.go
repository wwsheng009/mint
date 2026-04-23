package barchart

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey         = "key"
	propTitle       = "title"
	propLabels      = "labels"
	propValues      = "values"
	propSeries      = "series"
	propMode        = "mode"
	propOrientation = "orientation"
	propWidth       = "width"
	propHeight      = "height"
	propShowAxis    = "showAxis"
	propShowLegend  = "showLegend"
	propShowValue   = "showValue"
	propStyle       = "style"
)

// Mode controls how multiple series are laid out within each category.
type Mode int

const (
	ModeGrouped Mode = iota
	ModeStacked
)

// Orientation controls whether bars grow vertically or horizontally.
type Orientation int

const (
	OrientationVertical Orientation = iota
	OrientationHorizontal
)

// VNode is the declarative description of a bar chart component.
type VNode struct {
	*rtui.ElementVNode

	key         string
	title       string
	labels      []string
	values      []float64
	series      []Series
	mode        Mode
	orientation Orientation
	width       int
	height      int
	showAxis    bool
	showLegend  bool
	showValue   bool
	chartStyle  style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new bar chart VNode.
func New(values []float64) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("barchart"),
		values:       copyFloat64Slice(values),
		mode:         ModeGrouped,
		orientation:  OrientationVertical,
		showAxis:     true,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "barchart" }
func (v *VNode) Style() style.Style                           { return v.chartStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.chartStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:         v.key,
		propTitle:       v.title,
		propLabels:      copyStringSlice(v.labels),
		propValues:      copyFloat64Slice(v.values),
		propSeries:      copySeriesSlice(v.series),
		propMode:        v.mode,
		propOrientation: v.orientation,
		propWidth:       v.width,
		propHeight:      v.height,
		propShowAxis:    v.showAxis,
		propShowLegend:  v.showLegend,
		propShowValue:   v.showValue,
		propStyle:       v.chartStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if labels, ok := props[propLabels].([]string); ok {
		v.labels = copyStringSlice(labels)
	}
	if values, ok := props[propValues].([]float64); ok {
		v.values = copyFloat64Slice(values)
	}
	if series, ok := props[propSeries].([]Series); ok {
		v.series = copySeriesSlice(series)
	}
	if mode, ok := props[propMode].(Mode); ok {
		v.mode = mode
	}
	if orientation, ok := props[propOrientation].(Orientation); ok {
		v.orientation = orientation
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
	if showLegend, ok := props[propShowLegend].(bool); ok {
		v.showLegend = showLegend
	}
	if showValue, ok := props[propShowValue].(bool); ok {
		v.showValue = showValue
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

func (v *VNode) SetLabels(labels []string) *VNode {
	v.labels = copyStringSlice(labels)
	return v
}

func (v *VNode) SetValues(values []float64) *VNode {
	v.values = copyFloat64Slice(values)
	return v
}

func (v *VNode) SetSeries(series []Series) *VNode {
	v.series = copySeriesSlice(series)
	return v
}

func (v *VNode) SetMode(mode Mode) *VNode {
	v.mode = mode
	return v
}

func (v *VNode) SetOrientation(orientation Orientation) *VNode {
	v.orientation = orientation
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

func (v *VNode) SetShowLegend(show bool) *VNode {
	v.showLegend = show
	return v
}

func (v *VNode) SetShowValue(show bool) *VNode {
	v.showValue = show
	return v
}

func (v *VNode) Grouped() *VNode {
	v.mode = ModeGrouped
	return v
}

func (v *VNode) Stacked() *VNode {
	v.mode = ModeStacked
	return v
}

func (v *VNode) Vertical() *VNode {
	v.orientation = OrientationVertical
	return v
}

func (v *VNode) Horizontal() *VNode {
	v.orientation = OrientationHorizontal
	return v
}

func (v *VNode) SetChartStyle(s style.Style) *VNode {
	v.chartStyle = s
	return v
}

func (v *VNode) Title() string            { return v.title }
func (v *VNode) Labels() []string         { return copyStringSlice(v.labels) }
func (v *VNode) Values() []float64        { return copyFloat64Slice(v.values) }
func (v *VNode) Series() []Series         { return copySeriesSlice(v.series) }
func (v *VNode) Mode() Mode               { return v.mode }
func (v *VNode) Orientation() Orientation { return v.orientation }
func (v *VNode) Width() int               { return v.width }
func (v *VNode) Height() int              { return v.height }
func (v *VNode) ShowAxis() bool           { return v.showAxis }
func (v *VNode) ShowLegend() bool         { return v.showLegend }
func (v *VNode) ShowValue() bool          { return v.showValue }

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
