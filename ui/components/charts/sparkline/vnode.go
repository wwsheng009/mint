package sparkline

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey             = "key"
	propData            = "data"
	propTitle           = "title"
	propWidth           = "width"
	propHeight          = "height"
	propInlineLabel     = "inlineLabel"
	propHighlightMinMax = "highlightMinMax"
	propAutoHeight      = "autoHeight"
	propRenderMode      = "renderMode"
	propStyle           = "style"
)

// RenderMode controls the preferred sparkline rendering mode.
type RenderMode int

const (
	RenderModeAuto RenderMode = iota
	RenderModeBraille
	RenderModeBlock
	RenderModeASCII
)

// VNode is the declarative description of a sparkline component.
type VNode struct {
	*rtui.ElementVNode

	key             string
	data            []float64
	title           string
	width           int
	height          int
	inlineLabel     string
	highlightMinMax bool
	autoHeight      bool
	renderMode      RenderMode
	sparkStyle      style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new sparkline VNode.
func New(data []float64) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("sparkline"),
		data:         copyFloat64Slice(data),
		renderMode:   RenderModeAuto,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "sparkline" }
func (v *VNode) Style() style.Style                           { return v.sparkStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.sparkStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:             v.key,
		propData:            copyFloat64Slice(v.data),
		propTitle:           v.title,
		propWidth:           v.width,
		propHeight:          v.height,
		propInlineLabel:     v.inlineLabel,
		propHighlightMinMax: v.highlightMinMax,
		propAutoHeight:      v.autoHeight,
		propRenderMode:      v.renderMode,
		propStyle:           v.sparkStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if data, ok := props[propData].([]float64); ok {
		v.data = copyFloat64Slice(data)
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if height, ok := props[propHeight].(int); ok {
		v.height = height
	}
	if inlineLabel, ok := props[propInlineLabel].(string); ok {
		v.inlineLabel = inlineLabel
	}
	if highlightMinMax, ok := props[propHighlightMinMax].(bool); ok {
		v.highlightMinMax = highlightMinMax
	}
	if autoHeight, ok := props[propAutoHeight].(bool); ok {
		v.autoHeight = autoHeight
	}
	if mode, ok := props[propRenderMode].(RenderMode); ok {
		v.renderMode = mode
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.sparkStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetData(data []float64) *VNode {
	v.data = copyFloat64Slice(data)
	return v
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetHeight(height int) *VNode {
	v.height = height
	v.autoHeight = false
	return v
}

func (v *VNode) SetAutoHeight(autoHeight bool) *VNode {
	v.autoHeight = autoHeight
	if autoHeight {
		v.height = 0
	}
	return v
}

func (v *VNode) SetInlineLabel(label string) *VNode {
	v.inlineLabel = label
	return v
}

func (v *VNode) SetHighlightMinMax(highlight bool) *VNode {
	v.highlightMinMax = highlight
	return v
}

func (v *VNode) SetRenderMode(mode RenderMode) *VNode {
	v.renderMode = mode
	return v
}

func (v *VNode) SetSparkStyle(s style.Style) *VNode {
	v.sparkStyle = s
	return v
}

func (v *VNode) Auto() *VNode {
	v.renderMode = RenderModeAuto
	return v
}

func (v *VNode) Braille() *VNode {
	v.renderMode = RenderModeBraille
	return v
}

func (v *VNode) Block() *VNode {
	v.renderMode = RenderModeBlock
	return v
}

func (v *VNode) ASCII() *VNode {
	v.renderMode = RenderModeASCII
	return v
}

func (v *VNode) Data() []float64         { return copyFloat64Slice(v.data) }
func (v *VNode) Title() string           { return v.title }
func (v *VNode) Width() int              { return v.width }
func (v *VNode) Height() int             { return v.height }
func (v *VNode) InlineLabel() string     { return v.inlineLabel }
func (v *VNode) HighlightMinMax() bool   { return v.highlightMinMax }
func (v *VNode) AutoHeightEnabled() bool { return v.autoHeight }
func (v *VNode) Mode() RenderMode        { return v.renderMode }
