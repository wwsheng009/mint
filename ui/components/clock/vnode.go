package clock

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey             = "key"
	propShape           = "shape"
	propRadius          = "radius"
	propRadiusX         = "radiusX"
	propRadiusY         = "radiusY"
	propCellAspectX     = "cellAspectX"
	propLive            = "live"
	propTime            = "time"
	propLocation        = "location"
	propShowSecondHand  = "showSecondHand"
	propSmoothSecond    = "smoothSecond"
	propShowDigital     = "showDigital"
	propNumericTicks    = "numericTicks"
	propPreset          = "preset"
	propHandStyle       = "handStyle"
	propDialStyle       = "dialStyle"
	propTickStyle       = "tickStyle"
	propCenterStyle     = "centerStyle"
	propDigitalStyle    = "digitalStyle"
	propHourHandStyle   = "hourHandStyle"
	propMinuteHandStyle = "minuteHandStyle"
	propSecondHandStyle = "secondHandStyle"
	propStyle           = "style"
)

// HandRenderStyle controls how clock hands are drawn.
type HandRenderStyle int

const (
	HandRenderStyleASCII HandRenderStyle = iota
	HandRenderStyleUnicode
)

// VNode is the immutable description of a Clock component.
type VNode struct {
	*rtui.ElementVNode

	key             string
	shape           DialShape
	radius          int
	radiusY         int
	cellAspectX     float64
	live            bool
	timeValue       time.Time
	location        *time.Location
	showSecondHand  bool
	smoothSecond    bool
	showDigital     bool
	numericTicks    bool
	preset          Preset
	handStyle       HandRenderStyle
	dialStyle       style.Style
	tickStyle       style.Style
	centerStyle     style.Style
	digitalStyle    style.Style
	hourHandStyle   style.Style
	minuteHandStyle style.Style
	secondHandStyle style.Style
	clockStyle      style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Clock VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:   rtui.NewElement("clock"),
		shape:          DialShapeCircle,
		radius:         5,
		radiusY:        5,
		cellAspectX:    DefaultCellAspectX,
		live:           true,
		showSecondHand: true,
		smoothSecond:   true,
		showDigital:    true,
		preset:         PresetNone,
		handStyle:      HandRenderStyleASCII,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "clock" }
func (v *VNode) Style() style.Style                           { return v.clockStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.clockStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:             v.key,
		propShape:           v.shape,
		propRadius:          v.radius,
		propRadiusX:         v.radius,
		propRadiusY:         v.radiusY,
		propCellAspectX:     v.cellAspectX,
		propLive:            v.live,
		propTime:            v.timeValue,
		propLocation:        v.location,
		propShowSecondHand:  v.showSecondHand,
		propSmoothSecond:    v.smoothSecond,
		propShowDigital:     v.showDigital,
		propNumericTicks:    v.numericTicks,
		propPreset:          v.preset,
		propHandStyle:       v.handStyle,
		propDialStyle:       v.dialStyle,
		propTickStyle:       v.tickStyle,
		propCenterStyle:     v.centerStyle,
		propDigitalStyle:    v.digitalStyle,
		propHourHandStyle:   v.hourHandStyle,
		propMinuteHandStyle: v.minuteHandStyle,
		propSecondHandStyle: v.secondHandStyle,
		propStyle:           v.clockStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if radius, ok := props[propRadius].(int); ok {
		v.radius = radius
		v.radiusY = radius
	}
	if shape, ok := props[propShape].(DialShape); ok {
		v.shape = shape
	}
	if radiusX, ok := props[propRadiusX].(int); ok {
		v.radius = radiusX
	}
	if radiusY, ok := props[propRadiusY].(int); ok {
		v.radiusY = radiusY
	}
	if cellAspectX, ok := props[propCellAspectX].(float64); ok {
		v.cellAspectX = normalizeCellAspectX(cellAspectX)
	}
	if live, ok := props[propLive].(bool); ok {
		v.live = live
	}
	if timeValue, ok := props[propTime].(time.Time); ok {
		v.timeValue = timeValue
	}
	if location, ok := props[propLocation].(*time.Location); ok {
		v.location = location
	}
	if showSecondHand, ok := props[propShowSecondHand].(bool); ok {
		v.showSecondHand = showSecondHand
	}
	if smoothSecond, ok := props[propSmoothSecond].(bool); ok {
		v.smoothSecond = smoothSecond
	}
	if showDigital, ok := props[propShowDigital].(bool); ok {
		v.showDigital = showDigital
	}
	if numericTicks, ok := props[propNumericTicks].(bool); ok {
		v.numericTicks = numericTicks
	}
	if preset, ok := props[propPreset].(Preset); ok {
		v.preset = preset
	}
	if handStyle, ok := props[propHandStyle].(HandRenderStyle); ok {
		v.handStyle = handStyle
	}
	if dialStyle, ok := props[propDialStyle].(style.Style); ok {
		v.dialStyle = dialStyle
	}
	if tickStyle, ok := props[propTickStyle].(style.Style); ok {
		v.tickStyle = tickStyle
	}
	if centerStyle, ok := props[propCenterStyle].(style.Style); ok {
		v.centerStyle = centerStyle
	}
	if digitalStyle, ok := props[propDigitalStyle].(style.Style); ok {
		v.digitalStyle = digitalStyle
	}
	if hourHandStyle, ok := props[propHourHandStyle].(style.Style); ok {
		v.hourHandStyle = hourHandStyle
	}
	if minuteHandStyle, ok := props[propMinuteHandStyle].(style.Style); ok {
		v.minuteHandStyle = minuteHandStyle
	}
	if secondHandStyle, ok := props[propSecondHandStyle].(style.Style); ok {
		v.secondHandStyle = secondHandStyle
	}
	if clockStyle, ok := props[propStyle].(style.Style); ok {
		v.clockStyle = clockStyle
	}
	if v.shape == DialShapeCircle {
		v.radiusY = v.radius
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetShape(shape DialShape) *VNode {
	v.shape = shape
	if shape == DialShapeCircle {
		v.radiusY = v.radius
	}
	return v
}
func (v *VNode) SetRadius(radius int) *VNode { v.radius = radius; v.radiusY = radius; return v }
func (v *VNode) SetRadiusX(radius int) *VNode {
	v.radius = radius
	if v.shape == DialShapeCircle {
		v.radiusY = radius
	}
	return v
}
func (v *VNode) SetRadiusY(radius int) *VNode {
	v.radiusY = radius
	if v.shape == DialShapeCircle {
		v.radius = radius
	}
	return v
}
func (v *VNode) SetRadii(radiusX, radiusY int) *VNode {
	v.radius = radiusX
	v.radiusY = radiusY
	if v.shape == DialShapeCircle {
		v.radiusY = radiusX
	}
	return v
}
func (v *VNode) SetCellAspectX(cellAspectX float64) *VNode {
	v.cellAspectX = normalizeCellAspectX(cellAspectX)
	return v
}
func (v *VNode) SetLive(live bool) *VNode                   { v.live = live; return v }
func (v *VNode) SetTimeValue(timeValue time.Time) *VNode    { v.timeValue = timeValue; return v }
func (v *VNode) SetLocation(location *time.Location) *VNode { v.location = location; return v }
func (v *VNode) SetShowSecondHand(show bool) *VNode         { v.showSecondHand = show; return v }
func (v *VNode) SetSmoothSecond(smooth bool) *VNode         { v.smoothSecond = smooth; return v }
func (v *VNode) SetShowDigital(show bool) *VNode            { v.showDigital = show; return v }
func (v *VNode) SetNumericTicks(show bool) *VNode          { v.numericTicks = show; return v }
func (v *VNode) SetPreset(preset Preset) *VNode             { v.preset = preset; return v }
func (v *VNode) SetTheme(theme Theme) *VNode {
	v.clockStyle = theme.BaseStyle
	v.dialStyle = theme.DialStyle
	v.tickStyle = theme.TickStyle
	v.centerStyle = theme.CenterStyle
	v.digitalStyle = theme.DigitalStyle
	v.hourHandStyle = theme.HourHandStyle
	v.minuteHandStyle = theme.MinuteHandStyle
	v.secondHandStyle = theme.SecondHandStyle
	return v
}
func (v *VNode) SetHandRenderStyle(handStyle HandRenderStyle) *VNode {
	v.handStyle = handStyle
	return v
}
func (v *VNode) SetDialStyle(s style.Style) *VNode {
	v.dialStyle = s
	return v
}
func (v *VNode) SetTickStyle(s style.Style) *VNode {
	v.tickStyle = s
	return v
}
func (v *VNode) SetCenterStyle(s style.Style) *VNode {
	v.centerStyle = s
	return v
}
func (v *VNode) SetDigitalStyle(s style.Style) *VNode {
	v.digitalStyle = s
	return v
}
func (v *VNode) SetHourHandStyle(handStyle style.Style) *VNode {
	v.hourHandStyle = handStyle
	return v
}
func (v *VNode) SetMinuteHandStyle(handStyle style.Style) *VNode {
	v.minuteHandStyle = handStyle
	return v
}
func (v *VNode) SetSecondHandStyle(handStyle style.Style) *VNode {
	v.secondHandStyle = handStyle
	return v
}
func (v *VNode) SetClockStyle(s style.Style) *VNode { v.clockStyle = s; return v }
func (v *VNode) Shape() DialShape                   { return v.shape }
func (v *VNode) Radius() int                        { return v.radius }
func (v *VNode) RadiusX() int                       { return v.radius }
func (v *VNode) RadiusY() int                       { return v.radiusY }
func (v *VNode) CellAspectX() float64               { return v.cellAspectX }
func (v *VNode) Live() bool                         { return v.live }
func (v *VNode) TimeValue() time.Time               { return v.timeValue }
func (v *VNode) Location() *time.Location           { return v.location }
func (v *VNode) ShowSecondHand() bool               { return v.showSecondHand }
func (v *VNode) SmoothSecond() bool                 { return v.smoothSecond }
func (v *VNode) ShowDigital() bool                  { return v.showDigital }
func (v *VNode) Preset() Preset                     { return v.preset }
func (v *VNode) Theme() Theme {
	return Theme{
		BaseStyle:       v.clockStyle,
		DialStyle:       v.dialStyle,
		TickStyle:       v.tickStyle,
		CenterStyle:     v.centerStyle,
		DigitalStyle:    v.digitalStyle,
		HourHandStyle:   v.hourHandStyle,
		MinuteHandStyle: v.minuteHandStyle,
		SecondHandStyle: v.secondHandStyle,
	}
}
func (v *VNode) HandRenderStyle() HandRenderStyle { return v.handStyle }
func (v *VNode) DialStyle() style.Style           { return v.dialStyle }
func (v *VNode) TickStyle() style.Style           { return v.tickStyle }
func (v *VNode) CenterStyle() style.Style         { return v.centerStyle }
func (v *VNode) DigitalStyle() style.Style        { return v.digitalStyle }
func (v *VNode) HourHandStyle() style.Style       { return v.hourHandStyle }
func (v *VNode) MinuteHandStyle() style.Style     { return v.minuteHandStyle }
func (v *VNode) SecondHandStyle() style.Style     { return v.secondHandStyle }
func (v *VNode) ClockStyle() style.Style          { return v.clockStyle }
func (v *VNode) StaticTime(timeValue time.Time) *VNode {
	v.live = false
	v.timeValue = timeValue
	return v
}
func (v *VNode) Realtime() *VNode      { v.live = true; return v }
func (v *VNode) HideSeconds() *VNode   { v.showSecondHand = false; return v }
func (v *VNode) HideDigital() *VNode   { v.showDigital = false; return v }
func (v *VNode) Circle() *VNode        { v.shape = DialShapeCircle; v.radiusY = v.radius; return v }
func (v *VNode) Ellipse() *VNode       { v.shape = DialShapeEllipse; return v }
func (v *VNode) NoPreset() *VNode      { v.preset = PresetNone; return v }
func (v *VNode) ClassicPreset() *VNode { v.preset = PresetClassic; return v }
func (v *VNode) NeonPreset() *VNode    { v.preset = PresetNeon; return v }
func (v *VNode) MinimalPreset() *VNode { v.preset = PresetMinimal; return v }
func (v *VNode) AlertPreset() *VNode   { v.preset = PresetAlert; return v }
func (v *VNode) ASCIIHands() *VNode    { v.handStyle = HandRenderStyleASCII; return v }
func (v *VNode) UnicodeHands() *VNode  { v.handStyle = HandRenderStyleUnicode; return v }
