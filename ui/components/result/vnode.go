package result

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propBordered      = "bordered"
	propExtra         = "extra"
	propIcon          = "icon"
	propIconStyle     = "iconStyle"
	propKey           = "key"
	propResultStyle   = "style"
	propStatus        = "status"
	propSubtitle      = "subtitle"
	propSubtitleStyle = "subtitleStyle"
	propTitle         = "title"
	propTitleStyle    = "titleStyle"
	propWidth         = "width"
)

// Status controls preset iconography and color.
type Status int

const (
	StatusInfo Status = iota
	StatusSuccess
	StatusWarning
	StatusError
	Status403
	Status404
	Status500
)

// VNode is the declarative description of a Result component.
type VNode struct {
	*rtui.ElementVNode

	key           string
	status        Status
	icon          string
	title         string
	subtitle      string
	extra         rtui.VNode
	bordered      bool
	width         int
	rootStyle     style.Style
	iconStyle     style.Style
	titleStyle    style.Style
	subtitleStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Result VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("result"),
		status:       StatusInfo,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "result" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propBordered:      v.bordered,
		propExtra:         v.extra,
		propIcon:          v.icon,
		propIconStyle:     v.iconStyle,
		propKey:           v.key,
		propResultStyle:   v.rootStyle,
		propStatus:        v.status,
		propSubtitle:      v.subtitle,
		propSubtitleStyle: v.subtitleStyle,
		propTitle:         v.title,
		propTitleStyle:    v.titleStyle,
		propWidth:         v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if status, ok := props[propStatus].(Status); ok {
		v.status = status
	}
	if icon, ok := props[propIcon].(string); ok {
		v.icon = icon
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if subtitle, ok := props[propSubtitle].(string); ok {
		v.subtitle = subtitle
	}
	if extra, ok := props[propExtra].(rtui.VNode); ok {
		v.extra = extra
	}
	if bordered, ok := props[propBordered].(bool); ok {
		v.bordered = bordered
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if s, ok := props[propResultStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propIconStyle].(style.Style); ok {
		v.iconStyle = s
	}
	if s, ok := props[propTitleStyle].(style.Style); ok {
		v.titleStyle = s
	}
	if s, ok := props[propSubtitleStyle].(style.Style); ok {
		v.subtitleStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetStatus(status Status) *VNode {
	v.status = status
	return v
}

func (v *VNode) SetIcon(icon string) *VNode {
	v.icon = icon
	return v
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetSubtitle(subtitle string) *VNode {
	v.subtitle = subtitle
	return v
}

func (v *VNode) SetExtra(extra rtui.VNode) *VNode {
	v.extra = extra
	return v
}

func (v *VNode) SetBordered(bordered bool) *VNode {
	v.bordered = bordered
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetIconStyle(s style.Style) *VNode {
	v.iconStyle = s
	return v
}

func (v *VNode) SetTitleStyle(s style.Style) *VNode {
	v.titleStyle = s
	return v
}

func (v *VNode) SetSubtitleStyle(s style.Style) *VNode {
	v.subtitleStyle = s
	return v
}
