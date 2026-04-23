package badge

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propBadgeStyle    = "badgeStyle"
	propCount         = "count"
	propDot           = "dot"
	propKey           = "key"
	propLabel         = "label"
	propLabelStyle    = "labelStyle"
	propOverflowCount = "overflowCount"
	propShowZero      = "showZero"
	propStatus        = "status"
	propStyle         = "style"
	propText          = "text"
)

// Status controls badge coloring.
type Status int

const (
	StatusDefault Status = iota
	StatusPrimary
	StatusSuccess
	StatusWarning
	StatusError
	StatusProcessing
)

// VNode is the immutable description of a Badge component.
type VNode struct {
	*rtui.ElementVNode

	key           string
	label         string
	count         int
	text          string
	dot           bool
	showZero      bool
	overflowCount int
	status        Status
	baseStyle     style.Style
	labelStyle    style.Style
	badgeStyle    style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Badge VNode.
func New(label string) *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("badge"),
		label:         label,
		overflowCount: 99,
		status:        StatusError,
		showZero:      false,
	}
}

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "badge" }
func (v *VNode) Style() style.Style                           { return v.baseStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.baseStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propBadgeStyle:    v.badgeStyle,
		propCount:         v.count,
		propDot:           v.dot,
		propKey:           v.key,
		propLabel:         v.label,
		propLabelStyle:    v.labelStyle,
		propOverflowCount: v.overflowCount,
		propShowZero:      v.showZero,
		propStatus:        v.status,
		propStyle:         v.baseStyle,
		propText:          v.text,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if value, ok := props[propKey].(string); ok {
		v.key = value
	}
	if value, ok := props[propLabel].(string); ok {
		v.label = value
	}
	if value, ok := props[propCount].(int); ok {
		v.count = value
	}
	if value, ok := props[propText].(string); ok {
		v.text = value
	}
	if value, ok := props[propDot].(bool); ok {
		v.dot = value
	}
	if value, ok := props[propShowZero].(bool); ok {
		v.showZero = value
	}
	if value, ok := props[propOverflowCount].(int); ok {
		v.overflowCount = value
	}
	if value, ok := props[propStatus].(Status); ok {
		v.status = value
	}
	if value, ok := props[propStyle].(style.Style); ok {
		v.baseStyle = value
	}
	if value, ok := props[propLabelStyle].(style.Style); ok {
		v.labelStyle = value
	}
	if value, ok := props[propBadgeStyle].(style.Style); ok {
		v.badgeStyle = value
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetLabel(label string) *VNode       { v.label = label; return v }
func (v *VNode) SetCount(count int) *VNode          { v.count = count; return v }
func (v *VNode) SetText(text string) *VNode         { v.text = text; return v }
func (v *VNode) SetDot(dot bool) *VNode             { v.dot = dot; return v }
func (v *VNode) SetShowZero(show bool) *VNode       { v.showZero = show; return v }
func (v *VNode) SetOverflowCount(max int) *VNode    { v.overflowCount = max; return v }
func (v *VNode) SetStatus(status Status) *VNode     { v.status = status; return v }
func (v *VNode) SetLabelStyle(s style.Style) *VNode { v.labelStyle = s; return v }
func (v *VNode) SetBadgeStyle(s style.Style) *VNode { v.badgeStyle = s; return v }
func (v *VNode) Default() *VNode                    { v.status = StatusDefault; return v }
func (v *VNode) Primary() *VNode                    { v.status = StatusPrimary; return v }
func (v *VNode) Success() *VNode                    { v.status = StatusSuccess; return v }
func (v *VNode) Warning() *VNode                    { v.status = StatusWarning; return v }
func (v *VNode) Error() *VNode                      { v.status = StatusError; return v }
func (v *VNode) Processing() *VNode                 { v.status = StatusProcessing; return v }
