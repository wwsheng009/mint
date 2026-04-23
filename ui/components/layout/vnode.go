package layout

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propBodyGap    = "bodyGap"
	propContent    = "content"
	propFooter     = "footer"
	propHeader     = "header"
	propHeight     = "height"
	propKey        = "key"
	propLeftSider  = "leftSider"
	propRightSider = "rightSider"
	propStyle      = "style"
	propSectionGap = "sectionGap"
	propWidth      = "width"
)

// VNode is the declarative description of a Layout component.
type VNode struct {
	*rtui.ElementVNode

	key        string
	header     rtui.VNode
	leftSider  rtui.VNode
	content    rtui.VNode
	rightSider rtui.VNode
	footer     rtui.VNode
	sectionGap int
	bodyGap    int
	width      int
	height     int
	rootStyle  style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Layout VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("layout"),
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to key.
func (v *VNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	return v.key
}

func (v *VNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *VNode) Tag() string { return "layout" }

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
		propBodyGap:    v.bodyGap,
		propContent:    v.content,
		propFooter:     v.footer,
		propHeader:     v.header,
		propHeight:     v.height,
		propKey:        v.key,
		propLeftSider:  v.leftSider,
		propRightSider: v.rightSider,
		propStyle:      v.rootStyle,
		propSectionGap: v.sectionGap,
		propWidth:      v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if node, ok := props[propHeader].(rtui.VNode); ok {
		v.header = node
	}
	if node, ok := props[propLeftSider].(rtui.VNode); ok {
		v.leftSider = node
	}
	if node, ok := props[propContent].(rtui.VNode); ok {
		v.content = node
	}
	if node, ok := props[propRightSider].(rtui.VNode); ok {
		v.rightSider = node
	}
	if node, ok := props[propFooter].(rtui.VNode); ok {
		v.footer = node
	}
	if gap, ok := props[propSectionGap].(int); ok {
		v.sectionGap = gap
	}
	if gap, ok := props[propBodyGap].(int); ok {
		v.bodyGap = gap
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if height, ok := props[propHeight].(int); ok {
		v.height = height
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetHeader(header rtui.VNode) *VNode {
	v.header = header
	return v
}

func (v *VNode) SetSider(sider rtui.VNode) *VNode {
	v.leftSider = sider
	return v
}

func (v *VNode) SetLeftSider(sider rtui.VNode) *VNode {
	v.leftSider = sider
	return v
}

func (v *VNode) SetContent(content rtui.VNode) *VNode {
	v.content = content
	return v
}

func (v *VNode) SetRightSider(sider rtui.VNode) *VNode {
	v.rightSider = sider
	return v
}

func (v *VNode) SetFooter(footer rtui.VNode) *VNode {
	v.footer = footer
	return v
}

func (v *VNode) SetGap(gap int) *VNode {
	v.sectionGap = gap
	return v
}

func (v *VNode) SetBodyGap(gap int) *VNode {
	v.bodyGap = gap
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

func (v *VNode) SetStyleProps(s style.Style) *VNode {
	v.rootStyle = s
	return v
}
