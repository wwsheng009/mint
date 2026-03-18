package skeleton

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propActive           = "active"
	propAvatar           = "avatar"
	propAvatarShape      = "avatarShape"
	propAvatarSize       = "avatarSize"
	propContent          = "content"
	propGap              = "gap"
	propKey              = "key"
	propLoading          = "loading"
	propParagraph        = "paragraph"
	propParagraphRows    = "paragraphRows"
	propParagraphWidths  = "paragraphWidths"
	propPlaceholderStyle = "placeholderStyle"
	propStyle            = "style"
	propTitle            = "title"
	propTitleWidth       = "titleWidth"
	propWidth            = "width"
)

// Shape controls how avatar placeholders are rendered.
type Shape int

const (
	ShapeSquare Shape = iota
	ShapeRound
)

// VNode is the declarative description of a Skeleton component.
type VNode struct {
	*rtui.ElementVNode

	key              string
	content          rtui.VNode
	loading          bool
	active           bool
	showAvatar       bool
	avatarShape      Shape
	avatarSize       int
	showTitle        bool
	titleWidth       int
	showParagraph    bool
	paragraphRows    int
	paragraphWidths  []int
	width            int
	gap              int
	rootStyle        style.Style
	placeholderStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Skeleton VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:  rtui.NewElement("skeleton"),
		loading:       true,
		avatarShape:   ShapeSquare,
		avatarSize:    4,
		showTitle:     true,
		showParagraph: true,
		paragraphRows: 3,
		gap:           1,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "skeleton" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		v.content = children[0]
	} else {
		v.content = nil
	}
	return v
}

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propActive:           v.active,
		propAvatar:           v.showAvatar,
		propAvatarShape:      v.avatarShape,
		propAvatarSize:       v.avatarSize,
		propContent:          v.content,
		propGap:              v.gap,
		propKey:              v.key,
		propLoading:          v.loading,
		propParagraph:        v.showParagraph,
		propParagraphRows:    v.paragraphRows,
		propParagraphWidths:  cloneInts(v.paragraphWidths),
		propPlaceholderStyle: v.placeholderStyle,
		propStyle:            v.rootStyle,
		propTitle:            v.showTitle,
		propTitleWidth:       v.titleWidth,
		propWidth:            v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if content, ok := props[propContent].(rtui.VNode); ok {
		v.content = content
	}
	if loading, ok := props[propLoading].(bool); ok {
		v.loading = loading
	}
	if active, ok := props[propActive].(bool); ok {
		v.active = active
	}
	if showAvatar, ok := props[propAvatar].(bool); ok {
		v.showAvatar = showAvatar
	}
	if shape, ok := props[propAvatarShape].(Shape); ok {
		v.avatarShape = shape
	}
	if size, ok := props[propAvatarSize].(int); ok {
		v.avatarSize = size
	}
	if showTitle, ok := props[propTitle].(bool); ok {
		v.showTitle = showTitle
	}
	if titleWidth, ok := props[propTitleWidth].(int); ok {
		v.titleWidth = titleWidth
	}
	if showParagraph, ok := props[propParagraph].(bool); ok {
		v.showParagraph = showParagraph
	}
	if rows, ok := props[propParagraphRows].(int); ok {
		v.paragraphRows = rows
	}
	if widths, ok := props[propParagraphWidths].([]int); ok {
		v.paragraphWidths = cloneInts(widths)
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if gap, ok := props[propGap].(int); ok {
		v.gap = gap
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propPlaceholderStyle].(style.Style); ok {
		v.placeholderStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetContent(content rtui.VNode) *VNode {
	v.content = content
	return v
}

func (v *VNode) SetLoading(loading bool) *VNode {
	v.loading = loading
	return v
}

func (v *VNode) SetActive(active bool) *VNode {
	v.active = active
	return v
}

func (v *VNode) SetAvatar(show bool) *VNode {
	v.showAvatar = show
	return v
}

func (v *VNode) SetAvatarShape(shape Shape) *VNode {
	v.avatarShape = shape
	return v
}

func (v *VNode) SetAvatarSize(size int) *VNode {
	v.avatarSize = size
	return v
}

func (v *VNode) SetTitle(show bool) *VNode {
	v.showTitle = show
	return v
}

func (v *VNode) SetTitleWidth(width int) *VNode {
	v.titleWidth = width
	if width > 0 {
		v.showTitle = true
	}
	return v
}

func (v *VNode) SetParagraph(show bool) *VNode {
	v.showParagraph = show
	return v
}

func (v *VNode) SetParagraphRows(rows int) *VNode {
	v.paragraphRows = rows
	v.showParagraph = rows > 0
	return v
}

func (v *VNode) SetParagraphWidths(widths ...int) *VNode {
	v.paragraphWidths = cloneInts(widths)
	if len(widths) > 0 {
		v.showParagraph = true
	}
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetGap(gap int) *VNode {
	v.gap = gap
	return v
}

func (v *VNode) SetPlaceholderStyle(s style.Style) *VNode {
	v.placeholderStyle = s
	return v
}

func cloneInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}
