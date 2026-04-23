package candlestick

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey          = "key"
	propTitle        = "title"
	propCandles      = "candles"
	propWidth        = "width"
	propHeight       = "height"
	propShowAxis     = "showAxis"
	propShowGrid     = "showGrid"
	propShowLegend   = "showLegend"
	propShowVolume   = "showVolume"
	propVolumeHeight = "volumeHeight"
	propUpStyle      = "upStyle"
	propDownStyle    = "downStyle"
	propFlatStyle    = "flatStyle"
	propWickStyle    = "wickStyle"
	propVolumeStyle  = "volumeStyle"
	propStyle        = "style"
)

// VNode is the declarative description of a candlestick chart component.
type VNode struct {
	*rtui.ElementVNode

	key          string
	title        string
	candles      []Candle
	width        int
	height       int
	showAxis     bool
	showGrid     bool
	showLegend   bool
	showVolume   bool
	volumeHeight int
	upStyle      style.Style
	downStyle    style.Style
	flatStyle    style.Style
	wickStyle    style.Style
	volumeStyle  style.Style
	chartStyle   style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new candlestick VNode.
func New(candles []Candle) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("candlestick"),
		candles:      copyCandleSlice(candles),
		showAxis:     true,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "candlestick" }
func (v *VNode) Style() style.Style                           { return v.chartStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.chartStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:          v.key,
		propTitle:        v.title,
		propCandles:      copyCandleSlice(v.candles),
		propWidth:        v.width,
		propHeight:       v.height,
		propShowAxis:     v.showAxis,
		propShowGrid:     v.showGrid,
		propShowLegend:   v.showLegend,
		propShowVolume:   v.showVolume,
		propVolumeHeight: v.volumeHeight,
		propUpStyle:      v.upStyle,
		propDownStyle:    v.downStyle,
		propFlatStyle:    v.flatStyle,
		propWickStyle:    v.wickStyle,
		propVolumeStyle:  v.volumeStyle,
		propStyle:        v.chartStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if candles, ok := props[propCandles].([]Candle); ok {
		v.candles = copyCandleSlice(candles)
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
	if showVolume, ok := props[propShowVolume].(bool); ok {
		v.showVolume = showVolume
	}
	if volumeHeight, ok := props[propVolumeHeight].(int); ok {
		v.volumeHeight = volumeHeight
	}
	if s, ok := props[propUpStyle].(style.Style); ok {
		v.upStyle = s
	}
	if s, ok := props[propDownStyle].(style.Style); ok {
		v.downStyle = s
	}
	if s, ok := props[propFlatStyle].(style.Style); ok {
		v.flatStyle = s
	}
	if s, ok := props[propWickStyle].(style.Style); ok {
		v.wickStyle = s
	}
	if s, ok := props[propVolumeStyle].(style.Style); ok {
		v.volumeStyle = s
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

func (v *VNode) SetCandles(candles []Candle) *VNode {
	v.candles = copyCandleSlice(candles)
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

func (v *VNode) SetShowVolume(show bool) *VNode {
	v.showVolume = show
	return v
}

func (v *VNode) SetVolumeHeight(height int) *VNode {
	v.volumeHeight = height
	return v
}

func (v *VNode) SetUpStyle(s style.Style) *VNode {
	v.upStyle = s
	return v
}

func (v *VNode) SetDownStyle(s style.Style) *VNode {
	v.downStyle = s
	return v
}

func (v *VNode) SetFlatStyle(s style.Style) *VNode {
	v.flatStyle = s
	return v
}

func (v *VNode) SetWickStyle(s style.Style) *VNode {
	v.wickStyle = s
	return v
}

func (v *VNode) SetVolumeStyle(s style.Style) *VNode {
	v.volumeStyle = s
	return v
}

func (v *VNode) SetChartStyle(s style.Style) *VNode {
	v.chartStyle = s
	return v
}

func (v *VNode) Title() string            { return v.title }
func (v *VNode) Candles() []Candle        { return copyCandleSlice(v.candles) }
func (v *VNode) Width() int               { return v.width }
func (v *VNode) Height() int              { return v.height }
func (v *VNode) ShowAxis() bool           { return v.showAxis }
func (v *VNode) ShowGrid() bool           { return v.showGrid }
func (v *VNode) ShowLegend() bool         { return v.showLegend }
func (v *VNode) ShowVolume() bool         { return v.showVolume }
func (v *VNode) VolumeHeight() int        { return v.volumeHeight }
func (v *VNode) UpStyle() style.Style     { return v.upStyle }
func (v *VNode) DownStyle() style.Style   { return v.downStyle }
func (v *VNode) FlatStyle() style.Style   { return v.flatStyle }
func (v *VNode) WickStyle() style.Style   { return v.wickStyle }
func (v *VNode) VolumeStyle() style.Style { return v.volumeStyle }
func (v *VNode) ChartStyle() style.Style  { return v.chartStyle }
