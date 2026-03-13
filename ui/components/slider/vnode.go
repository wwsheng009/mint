package slider

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propKey          = "key"
	propLabel        = "label"
	propValue        = "value"
	propMin          = "min"
	propMax          = "max"
	propStep         = "step"
	propWidth        = "width"
	propDisabled     = "disabled"
	propShowValue    = "showValue"
	propStyle        = "style"
	propChangeIntent = "changeIntent"
	propFormID       = "formID"
)

const (
	defaultWidth = 20
	defaultMax   = 100
	defaultStep  = 1
)

// VNode is the declarative description of a Slider component.
type VNode struct {
	*rtui.ElementVNode

	key   string
	label string
	style style.Style

	changeIntent intent.Intent
	formID       string

	value     int
	min       int
	max       int
	step      int
	width     int
	disabled  bool
	showValue bool

	rtui.BoxModelMixin
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// New creates a new Slider VNode with defaults.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("slider"),
		max:          defaultMax,
		step:         defaultStep,
		width:        defaultWidth,
		showValue:    true,
	}
}

// Key returns the component key.
func (s *VNode) Key() string { return s.key }

// SetKey sets the component key.
func (s *VNode) SetKey(key string) rtui.VNode { s.key = key; return s }

// Tag returns the tag name.
func (s *VNode) Tag() string { return "slider" }

// Style returns the visual style.
func (s *VNode) Style() style.Style { return s.style }

// SetStyle sets the visual style.
func (s *VNode) SetStyle(st style.Style) rtui.VNode { s.style = st; return s }

// Children returns nil (no children).
func (s *VNode) Children() []rtui.VNode { return nil }

// SetChildren is a no-op for slider.
func (s *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return s }

// GetLayer returns the rendering layer.
func (s *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

// SetLayer sets the rendering layer.
func (s *VNode) SetLayer(l rtui.Layer) rtui.VNode { return s }

// Props returns the node properties.
func (s *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:          s.key,
		propLabel:        s.label,
		propValue:        s.value,
		propMin:          s.min,
		propMax:          s.max,
		propStep:         s.step,
		propWidth:        s.width,
		propDisabled:     s.disabled,
		propShowValue:    s.showValue,
		propStyle:        s.style,
		propChangeIntent: s.changeIntent,
		propFormID:       s.formID,
	}
}

// SetProps sets the node properties.
func (s *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		s.key = v
	}
	if v, ok := p[propLabel].(string); ok {
		s.label = v
	}
	if v, ok := p[propValue].(int); ok {
		s.value = v
	}
	if v, ok := p[propMin].(int); ok {
		s.min = v
	}
	if v, ok := p[propMax].(int); ok {
		s.max = v
	}
	if v, ok := p[propStep].(int); ok {
		s.step = v
	}
	if v, ok := p[propWidth].(int); ok {
		s.width = v
	}
	if v, ok := p[propDisabled].(bool); ok {
		s.disabled = v
	}
	if v, ok := p[propShowValue].(bool); ok {
		s.showValue = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		s.style = v
	}
	if v, ok := p[propChangeIntent].(intent.Intent); ok {
		s.changeIntent = v
	}
	if v, ok := p[propFormID].(string); ok {
		s.formID = v
	}
	return s
}

// CreateInstance creates a new Slider instance from this VNode.
func (s *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(s.Props())
}

// --- Builder methods on VNode ---

func (s *VNode) SetLabel(label string) *VNode        { s.label = label; return s }
func (s *VNode) SetValue(value int) *VNode           { s.value = value; return s }
func (s *VNode) SetMin(min int) *VNode               { s.min = min; return s }
func (s *VNode) SetMax(max int) *VNode               { s.max = max; return s }
func (s *VNode) SetStep(step int) *VNode             { s.step = step; return s }
func (s *VNode) SetWidth(width int) *VNode           { s.width = width; return s }
func (s *VNode) SetDisabled(disabled bool) *VNode    { s.disabled = disabled; return s }
func (s *VNode) SetShowValue(show bool) *VNode       { s.showValue = show; return s }
func (s *VNode) SetStyleProps(st style.Style) *VNode { s.style = st; return s }
func (s *VNode) SetIntent(i intent.Intent) *VNode    { s.changeIntent = i; return s }
func (s *VNode) SetFormID(formID string) *VNode      { s.formID = formID; return s }

// GetBoxModel returns the box model for the Slider VNode.
func (s *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   s.BoxModelMixin.Padding()[3],
			Right:  s.BoxModelMixin.Padding()[1],
			Top:    s.BoxModelMixin.Padding()[0],
			Bottom: s.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   s.BoxModelMixin.Margin()[3],
			Right:  s.BoxModelMixin.Margin()[1],
			Top:    s.BoxModelMixin.Margin()[0],
			Bottom: s.BoxModelMixin.Margin()[2],
		},
		Border: layout.Border{Style: layout.BorderNone},
	}
}
