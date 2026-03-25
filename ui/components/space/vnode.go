package space

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propAlign     = "align"
	propChildren  = "children"
	propDirection = "direction"
	propKey       = "key"
	propSize      = "size"
	propSplit     = "split"
	propStyle     = "style"
	propWidth     = "width"
	propWrap      = "wrap"
)

// Direction represents the main axis direction for Space.
type Direction string

const (
	DirectionHorizontal Direction = "horizontal"
	DirectionVertical   Direction = "vertical"
)

const (
	SizeSmall  = 1
	SizeMiddle = 2
	SizeLarge  = 4
)

// Align aliases the runtime layout alignment type.
type Align = rtui.Align

const (
	AlignStart  = rtui.AlignStart
	AlignCenter = rtui.AlignCenter
	AlignEnd    = rtui.AlignEnd
)

// VNode is the declarative description of a Space component.
type VNode struct {
	*rtui.ElementVNode

	key       string
	direction Direction
	size      int
	wrap      bool
	width     int
	align     Align
	split     string
	children  []rtui.VNode
	rootStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Space VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("space"),
		direction:    DirectionHorizontal,
		size:         SizeSmall,
		align:        AlignStart,
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

func (v *VNode) Tag() string { return "space" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode {
	return append([]rtui.VNode(nil), v.children...)
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	v.children = append([]rtui.VNode(nil), children...)
	return v
}

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propAlign:     v.align,
		propChildren:  append([]rtui.VNode(nil), v.children...),
		propDirection: v.direction,
		propKey:       v.key,
		propSize:      v.size,
		propSplit:     v.split,
		propStyle:     v.rootStyle,
		propWidth:     v.width,
		propWrap:      v.wrap,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if direction, ok := props[propDirection].(Direction); ok {
		v.direction = direction
	}
	if size, ok := props[propSize].(int); ok {
		v.size = size
	}
	if wrap, ok := props[propWrap].(bool); ok {
		v.wrap = wrap
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if align, ok := props[propAlign].(Align); ok {
		v.align = align
	}
	if split, ok := props[propSplit].(string); ok {
		v.split = split
	}
	if children, ok := props[propChildren].([]rtui.VNode); ok {
		v.children = append([]rtui.VNode(nil), children...)
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetDirection(direction Direction) *VNode {
	v.direction = direction
	return v
}

func (v *VNode) SetSize(size int) *VNode {
	v.size = size
	return v
}

func (v *VNode) SetWrap(wrap bool) *VNode {
	v.wrap = wrap
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetAlign(align Align) *VNode {
	v.align = align
	return v
}

func (v *VNode) SetSplit(split string) *VNode {
	v.split = split
	return v
}

func (v *VNode) SetChildrenList(children []rtui.VNode) *VNode {
	v.children = append([]rtui.VNode(nil), children...)
	return v
}

func (v *VNode) AddChild(child rtui.VNode) *VNode {
	v.children = append(v.children, child)
	return v
}

func (v *VNode) SetStyleProps(s style.Style) *VNode {
	v.rootStyle = s
	return v
}
