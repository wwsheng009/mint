// Package splitpane provides a Fiber-first split pane layout component.
package splitpane

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propAlign          = "align"
	propDirection      = "direction"
	propGap            = "gap"
	propHeight         = "height"
	propKey            = "key"
	propPrimary        = "primary"
	propPrimaryFlex    = "primaryFlex"
	propPrimarySize    = "primarySize"
	propSecondary      = "secondary"
	propSecondaryFlex  = "secondaryFlex"
	propSecondarySize  = "secondarySize"
	propSeparator      = "separator"
	propSeparatorGlyph = "separatorGlyph"
	propSeparatorStyle = "separatorStyle"
	propStyle          = "style"
	propWidth          = "width"
)

// Direction controls whether panes are arranged left/right or top/bottom.
type Direction string

const (
	// DirectionHorizontal arranges primary and secondary panes left to right.
	DirectionHorizontal Direction = "horizontal"
	// DirectionVertical arranges primary and secondary panes top to bottom.
	DirectionVertical Direction = "vertical"
)

// VNode is the declarative description of a SplitPane component.
type VNode struct {
	*rtui.ElementVNode

	key            string
	direction      Direction
	primary        rtui.VNode
	secondary      rtui.VNode
	primarySize    int
	secondarySize  int
	primaryFlex    int
	secondaryFlex  int
	gap            int
	separator      bool
	separatorGlyph string
	width          int
	height         int
	align          rtui.Align
	rootStyle      style.Style
	separatorStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a SplitPane VNode with master/detail-friendly defaults.
func New() *VNode {
	return &VNode{
		ElementVNode:   rtui.NewElement("splitpane"),
		direction:      DirectionHorizontal,
		secondaryFlex:  1,
		gap:            1,
		separator:      true,
		separatorGlyph: "│",
		align:          rtui.AlignStart,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to the key.
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

func (v *VNode) Tag() string { return "splitpane" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode {
	return nil
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		v.primary = children[0]
	}
	if len(children) > 1 {
		v.secondary = children[1]
	}
	return v
}

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propAlign:          v.align,
		propDirection:      v.direction,
		propGap:            v.gap,
		propHeight:         v.height,
		propKey:            v.key,
		propPrimary:        v.primary,
		propPrimaryFlex:    v.primaryFlex,
		propPrimarySize:    v.primarySize,
		propSecondary:      v.secondary,
		propSecondaryFlex:  v.secondaryFlex,
		propSecondarySize:  v.secondarySize,
		propSeparator:      v.separator,
		propSeparatorGlyph: v.separatorGlyph,
		propSeparatorStyle: v.separatorStyle,
		propStyle:          v.rootStyle,
		propWidth:          v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	v.key = getStringProp(props, propKey, v.key)
	v.direction = getDirectionProp(props, v.direction)
	v.primary = getVNodeProp(props, propPrimary, v.primary)
	v.secondary = getVNodeProp(props, propSecondary, v.secondary)
	v.primarySize = getIntProp(props, propPrimarySize, v.primarySize)
	v.secondarySize = getIntProp(props, propSecondarySize, v.secondarySize)
	v.primaryFlex = getIntProp(props, propPrimaryFlex, v.primaryFlex)
	v.secondaryFlex = getIntProp(props, propSecondaryFlex, v.secondaryFlex)
	v.gap = getIntProp(props, propGap, v.gap)
	v.separator = getBoolProp(props, propSeparator, v.separator)
	v.separatorGlyph = getStringProp(props, propSeparatorGlyph, v.separatorGlyph)
	v.width = getIntProp(props, propWidth, v.width)
	v.height = getIntProp(props, propHeight, v.height)
	v.align = getAlignProp(props, propAlign, v.align)
	v.rootStyle = getStyleProp(props, propStyle, v.rootStyle)
	v.separatorStyle = getStyleProp(props, propSeparatorStyle, v.separatorStyle)
	v.normalize()
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetDirection(direction Direction) *VNode {
	oldDirection := v.direction
	oldGlyph := v.separatorGlyph
	v.direction = direction
	v.normalize()
	if oldDirection != v.direction && isDefaultGlyphForDirection(oldDirection, oldGlyph) {
		v.separatorGlyph = defaultGlyphForDirection(v.direction)
	}
	return v
}

func (v *VNode) SetPrimary(primary rtui.VNode) *VNode {
	v.primary = primary
	return v
}

func (v *VNode) SetSecondary(secondary rtui.VNode) *VNode {
	v.secondary = secondary
	return v
}

func (v *VNode) SetPrimarySize(size int) *VNode {
	v.primarySize = size
	v.normalize()
	return v
}

func (v *VNode) SetSecondarySize(size int) *VNode {
	v.secondarySize = size
	v.normalize()
	return v
}

func (v *VNode) SetPrimaryFlex(flex int) *VNode {
	v.primaryFlex = flex
	v.normalize()
	return v
}

func (v *VNode) SetSecondaryFlex(flex int) *VNode {
	v.secondaryFlex = flex
	v.normalize()
	return v
}

func (v *VNode) SetGap(gap int) *VNode {
	v.gap = gap
	v.normalize()
	return v
}

func (v *VNode) SetSeparator(enabled bool) *VNode {
	v.separator = enabled
	return v
}

func (v *VNode) SetSeparatorGlyph(glyph string) *VNode {
	v.separatorGlyph = glyph
	v.normalize()
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	v.normalize()
	return v
}

func (v *VNode) SetHeight(height int) *VNode {
	v.height = height
	v.normalize()
	return v
}

func (v *VNode) SetAlign(align rtui.Align) *VNode {
	v.align = align
	return v
}

func (v *VNode) SetRootStyle(s style.Style) *VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) SetSeparatorStyle(s style.Style) *VNode {
	v.separatorStyle = s
	return v
}

func (v *VNode) normalize() {
	if v.direction != DirectionVertical {
		v.direction = DirectionHorizontal
	}
	if v.primarySize < 0 {
		v.primarySize = 0
	}
	if v.secondarySize < 0 {
		v.secondarySize = 0
	}
	if v.primaryFlex < 0 {
		v.primaryFlex = 0
	}
	if v.secondaryFlex < 0 {
		v.secondaryFlex = 0
	}
	if v.gap < 0 {
		v.gap = 0
	}
	if v.width < 0 {
		v.width = 0
	}
	if v.height < 0 {
		v.height = 0
	}
	if v.separatorGlyph == "" {
		v.separatorGlyph = defaultGlyphForDirection(v.direction)
	}
}

func defaultGlyphForDirection(direction Direction) string {
	if direction == DirectionVertical {
		return "─"
	}
	return "│"
}

func isDefaultGlyphForDirection(direction Direction, glyph string) bool {
	if glyph == "" {
		return true
	}
	return glyph == defaultGlyphForDirection(direction)
}
