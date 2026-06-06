package pageviewport

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propChild         = "child"
	propHeight        = "height"
	propKey           = "key"
	propScrollOffset  = "scrollOffset"
	propShowIndicator = "showIndicator"
	propStyle         = "style"
	propWidth         = "width"
)

// VNode describes an interactive clipped viewport that preserves its child tree.
type VNode struct {
	*rtui.ElementVNode

	key             string
	child           rtui.VNode
	width           int
	height          int
	scrollOffset    int
	scrollOffsetSet bool
	showIndicator   bool
	style           style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a PageViewport VNode.
func New(child rtui.VNode) *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("pageviewport"),
		child:         child,
		showIndicator: true,
	}
}

func (v *VNode) Type() rtui.VNodeType { return rtui.VNodeElement }
func (v *VNode) Key() string          { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "pageviewport" }

func (v *VNode) Style() style.Style { return v.style }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.style = s
	return v
}

func (v *VNode) Children() []rtui.VNode {
	if v.child == nil {
		return nil
	}
	return []rtui.VNode{v.child}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		v.child = children[0]
	} else {
		v.child = nil
	}
	return v
}

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	props := rtui.Props{
		propKey:           v.key,
		propChild:         v.child,
		propWidth:         v.width,
		propHeight:        v.height,
		propShowIndicator: v.showIndicator,
		propStyle:         v.style,
	}
	if v.scrollOffsetSet {
		props[propScrollOffset] = v.scrollOffset
	}
	return props
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p[propKey].(string); ok {
		v.key = val
	}
	if val, ok := p[propChild].(rtui.VNode); ok {
		v.child = val
	}
	if val, ok := p[propWidth].(int); ok {
		v.width = val
	}
	if val, ok := p[propHeight].(int); ok {
		v.height = val
	}
	if val, ok := p[propScrollOffset].(int); ok {
		v.scrollOffset = val
		v.scrollOffsetSet = true
	}
	if val, ok := p[propShowIndicator].(bool); ok {
		v.showIndicator = val
	}
	if val, ok := p[propStyle].(style.Style); ok {
		v.style = val
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetChild(child rtui.VNode) *VNode { v.child = child; return v }
func (v *VNode) SetWidth(width int) *VNode        { v.width = width; return v }
func (v *VNode) SetHeight(height int) *VNode      { v.height = height; return v }
func (v *VNode) SetScrollOffset(offset int) *VNode {
	v.scrollOffset = offset
	v.scrollOffsetSet = true
	return v
}
func (v *VNode) SetShowIndicator(show bool) *VNode { v.showIndicator = show; return v }
