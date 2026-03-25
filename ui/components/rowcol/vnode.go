package rowcol

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propAlign          = "align"
	propChildren       = "children"
	propColOffset      = "offset"
	propColSpan        = "span"
	propGutter         = "gutter"
	propJustify        = "justify"
	propKey            = "key"
	propRowStyle       = "style"
	propVerticalGutter = "verticalGutter"
	propWidth          = "width"
	propWrap           = "wrap"
)

// RowVNode is the declarative description of a Row component.
type RowVNode struct {
	*rtui.ElementVNode

	key            string
	children       []rtui.VNode
	justify        rtui.Align
	align          rtui.Align
	gutter         int
	verticalGutter int
	wrap           bool
	width          int
	rootStyle      style.Style
}

// ColVNode is the declarative description of a Col component.
type ColVNode struct {
	*rtui.ElementVNode

	key       string
	span      int
	offset    int
	children  []rtui.VNode
	rootStyle style.Style
}

var (
	_ rtui.VNode           = (*RowVNode)(nil)
	_ rtui.InstanceFactory = (*RowVNode)(nil)
	_ rtui.VNode           = (*ColVNode)(nil)
	_ rtui.InstanceFactory = (*ColVNode)(nil)
)

// NewRow creates a new Row VNode.
func NewRow() *RowVNode {
	return &RowVNode{
		ElementVNode: rtui.NewElement("row"),
		justify:      rtui.AlignStart,
		align:        rtui.AlignStart,
		wrap:         true,
	}
}

// NewCol creates a new Col VNode.
func NewCol() *ColVNode {
	return &ColVNode{
		ElementVNode: rtui.NewElement("col"),
	}
}

func (v *RowVNode) Key() string { return v.key }

func (v *RowVNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to key.
func (v *RowVNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	return v.key
}

func (v *RowVNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *RowVNode) Tag() string { return "row" }

func (v *RowVNode) Style() style.Style { return v.rootStyle }

func (v *RowVNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *RowVNode) Children() []rtui.VNode {
	return append([]rtui.VNode(nil), v.children...)
}

func (v *RowVNode) SetChildren(children []rtui.VNode) rtui.VNode {
	v.children = append([]rtui.VNode(nil), children...)
	return v
}

func (v *RowVNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *RowVNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *RowVNode) Props() rtui.Props {
	return rtui.Props{
		propAlign:          v.align,
		propChildren:       append([]rtui.VNode(nil), v.children...),
		propGutter:         v.gutter,
		propJustify:        v.justify,
		propKey:            v.key,
		propRowStyle:       v.rootStyle,
		propVerticalGutter: v.verticalGutter,
		propWidth:          v.width,
		propWrap:           v.wrap,
	}
}

func (v *RowVNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if children, ok := props[propChildren].([]rtui.VNode); ok {
		v.children = append([]rtui.VNode(nil), children...)
	}
	if justify, ok := props[propJustify].(rtui.Align); ok {
		v.justify = justify
	}
	if align, ok := props[propAlign].(rtui.Align); ok {
		v.align = align
	}
	if gutter, ok := props[propGutter].(int); ok {
		v.gutter = gutter
	}
	if gutter, ok := props[propVerticalGutter].(int); ok {
		v.verticalGutter = gutter
	}
	if wrap, ok := props[propWrap].(bool); ok {
		v.wrap = wrap
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if s, ok := props[propRowStyle].(style.Style); ok {
		v.rootStyle = s
	}
	return v
}

func (v *RowVNode) CreateInstance() rtui.ComponentInstance {
	return NewRowInstance(v.Props())
}

func (v *RowVNode) SetJustify(justify rtui.Align) *RowVNode {
	v.justify = justify
	return v
}

func (v *RowVNode) SetAlign(align rtui.Align) *RowVNode {
	v.align = align
	return v
}

func (v *RowVNode) SetGutter(gutter int) *RowVNode {
	v.gutter = gutter
	return v
}

func (v *RowVNode) SetVerticalGutter(gutter int) *RowVNode {
	v.verticalGutter = gutter
	return v
}

func (v *RowVNode) SetWrap(wrap bool) *RowVNode {
	v.wrap = wrap
	return v
}

func (v *RowVNode) SetWidth(width int) *RowVNode {
	v.width = width
	return v
}

func (v *RowVNode) SetChildrenList(children []rtui.VNode) *RowVNode {
	v.children = append([]rtui.VNode(nil), children...)
	return v
}

func (v *RowVNode) AddChild(child rtui.VNode) *RowVNode {
	v.children = append(v.children, child)
	return v
}

func (v *RowVNode) SetStyleProps(s style.Style) *RowVNode {
	v.rootStyle = s
	return v
}

func (v *ColVNode) Key() string { return v.key }

func (v *ColVNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to key.
func (v *ColVNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	return v.key
}

func (v *ColVNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *ColVNode) Tag() string { return "col" }

func (v *ColVNode) Style() style.Style { return v.rootStyle }

func (v *ColVNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *ColVNode) Children() []rtui.VNode {
	return append([]rtui.VNode(nil), v.children...)
}

func (v *ColVNode) SetChildren(children []rtui.VNode) rtui.VNode {
	v.children = append([]rtui.VNode(nil), children...)
	return v
}

func (v *ColVNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *ColVNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *ColVNode) Props() rtui.Props {
	return rtui.Props{
		propChildren:  append([]rtui.VNode(nil), v.children...),
		propColOffset: v.offset,
		propColSpan:   v.span,
		propKey:       v.key,
		propRowStyle:  v.rootStyle,
	}
}

func (v *ColVNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if children, ok := props[propChildren].([]rtui.VNode); ok {
		v.children = append([]rtui.VNode(nil), children...)
	}
	if offset, ok := props[propColOffset].(int); ok {
		v.offset = offset
	}
	if span, ok := props[propColSpan].(int); ok {
		v.span = span
	}
	if s, ok := props[propRowStyle].(style.Style); ok {
		v.rootStyle = s
	}
	return v
}

func (v *ColVNode) CreateInstance() rtui.ComponentInstance {
	return NewColInstance(v.Props())
}

func (v *ColVNode) SetSpan(span int) *ColVNode {
	v.span = span
	return v
}

func (v *ColVNode) SetOffset(offset int) *ColVNode {
	v.offset = offset
	return v
}

func (v *ColVNode) SetChildrenList(children []rtui.VNode) *ColVNode {
	v.children = append([]rtui.VNode(nil), children...)
	return v
}

func (v *ColVNode) AddChild(child rtui.VNode) *ColVNode {
	v.children = append(v.children, child)
	return v
}

func (v *ColVNode) SetStyleProps(s style.Style) *ColVNode {
	v.rootStyle = s
	return v
}
