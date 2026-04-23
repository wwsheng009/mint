package bulletchart

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey               = "key"
	propLabel             = "label"
	propValue             = "value"
	propTarget            = "target"
	propMax               = "max"
	propWidth             = "width"
	propShowTarget        = "showTarget"
	propShowValueText     = "showValueText"
	propValueLabelMode    = "valueLabelMode"
	propDirection         = "direction"
	propQualitativeRanges = "qualitativeRanges"
	propTargetMarkerRune  = "targetMarkerRune"
	propTargetMarkerStyle = "targetMarkerStyle"
	propStyle             = "style"
)

// VNode is the declarative description of a bullet chart component.
type VNode struct {
	*rtui.ElementVNode

	key               string
	label             string
	value             int
	target            int
	max               int
	width             int
	showTarget        bool
	showValueText     bool
	valueLabelMode    ValueLabelMode
	direction         Direction
	qualitativeRanges []QualitativeRange
	targetMarkerRune  rune
	targetMarkerStyle style.Style
	chartStyle        style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new bullet chart VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:     rtui.NewElement("bulletchart"),
		max:              100,
		width:            20,
		showTarget:       true,
		showValueText:    true,
		valueLabelMode:   ValueLabelModeAuto,
		direction:        DirectionNeutral,
		targetMarkerRune: '│',
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "bulletchart" }
func (v *VNode) Style() style.Style                           { return v.chartStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.chartStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:               v.key,
		propLabel:             v.label,
		propValue:             v.value,
		propTarget:            v.target,
		propMax:               v.max,
		propWidth:             v.width,
		propShowTarget:        v.showTarget,
		propShowValueText:     v.showValueText,
		propValueLabelMode:    v.valueLabelMode,
		propDirection:         v.direction,
		propQualitativeRanges: copyQualitativeRangeSlice(v.qualitativeRanges, v.max),
		propTargetMarkerRune:  v.targetMarkerRune,
		propTargetMarkerStyle: v.targetMarkerStyle,
		propStyle:             v.chartStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if label, ok := props[propLabel].(string); ok {
		v.label = label
	}
	if value, ok := props[propValue].(int); ok {
		v.value = value
	}
	if target, ok := props[propTarget].(int); ok {
		v.target = target
	}
	if max, ok := props[propMax].(int); ok {
		v.max = max
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if showTarget, ok := props[propShowTarget].(bool); ok {
		v.showTarget = showTarget
	}
	if showValueText, ok := props[propShowValueText].(bool); ok {
		v.showValueText = showValueText
	}
	if valueLabelMode, ok := props[propValueLabelMode].(ValueLabelMode); ok {
		v.valueLabelMode = valueLabelMode
	}
	if direction, ok := props[propDirection].(Direction); ok {
		v.direction = direction
	}
	if ranges, ok := props[propQualitativeRanges].([]QualitativeRange); ok {
		v.qualitativeRanges = copyQualitativeRangeSlice(ranges, v.max)
	}
	if targetMarkerRune, ok := props[propTargetMarkerRune].(rune); ok {
		v.targetMarkerRune = targetMarkerRune
	}
	if targetMarkerStyle, ok := props[propTargetMarkerStyle].(style.Style); ok {
		v.targetMarkerStyle = targetMarkerStyle
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.chartStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetLabel(label string) *VNode {
	v.label = label
	return v
}

func (v *VNode) SetValue(value int) *VNode {
	v.value = value
	return v
}

func (v *VNode) SetTarget(target int) *VNode {
	v.target = target
	return v
}

func (v *VNode) SetMax(max int) *VNode {
	v.max = max
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetShowTarget(show bool) *VNode {
	v.showTarget = show
	return v
}

func (v *VNode) SetShowValueText(show bool) *VNode {
	v.showValueText = show
	return v
}

func (v *VNode) SetValueLabelMode(mode ValueLabelMode) *VNode {
	v.valueLabelMode = mode
	return v
}

func (v *VNode) SetDirection(direction Direction) *VNode {
	v.direction = direction
	return v
}

func (v *VNode) SetQualitativeRanges(ranges []QualitativeRange) *VNode {
	v.qualitativeRanges = copyQualitativeRangeSlice(ranges, v.max)
	return v
}

func (v *VNode) SetTargetMarkerRune(marker rune) *VNode {
	if marker == 0 {
		marker = '│'
	}
	v.targetMarkerRune = marker
	return v
}

func (v *VNode) SetTargetMarkerStyle(s style.Style) *VNode {
	v.targetMarkerStyle = s
	return v
}

func (v *VNode) SetChartStyle(s style.Style) *VNode {
	v.chartStyle = s
	return v
}

func (v *VNode) Label() string                  { return v.label }
func (v *VNode) Value() int                     { return v.value }
func (v *VNode) Target() int                    { return v.target }
func (v *VNode) Max() int                       { return v.max }
func (v *VNode) Width() int                     { return v.width }
func (v *VNode) ShowTarget() bool               { return v.showTarget }
func (v *VNode) ShowValueText() bool            { return v.showValueText }
func (v *VNode) ValueLabelMode() ValueLabelMode { return v.valueLabelMode }
func (v *VNode) Direction() Direction           { return v.direction }
func (v *VNode) QualitativeRanges() []QualitativeRange {
	return copyQualitativeRangeSlice(v.qualitativeRanges, v.max)
}
func (v *VNode) TargetMarkerRune() rune         { return v.targetMarkerRune }
func (v *VNode) TargetMarkerStyle() style.Style { return v.targetMarkerStyle }
