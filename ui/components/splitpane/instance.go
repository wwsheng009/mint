package splitpane

import (
	"reflect"
	"strconv"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

// Instance is the runtime entity for SplitPane.
type Instance struct {
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
	dirty          bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a SplitPane instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:            getStringProp(props, propKey, ""),
		direction:      getDirectionProp(props, DirectionHorizontal),
		primary:        getVNodeProp(props, propPrimary, nil),
		secondary:      getVNodeProp(props, propSecondary, nil),
		primarySize:    getIntProp(props, propPrimarySize, 0),
		secondarySize:  getIntProp(props, propSecondarySize, 0),
		primaryFlex:    getIntProp(props, propPrimaryFlex, 0),
		secondaryFlex:  getIntProp(props, propSecondaryFlex, 1),
		gap:            getIntProp(props, propGap, 1),
		separator:      getBoolProp(props, propSeparator, true),
		separatorGlyph: getStringProp(props, propSeparatorGlyph, "│"),
		width:          getIntProp(props, propWidth, 0),
		height:         getIntProp(props, propHeight, 0),
		align:          getAlignProp(props, propAlign, rtui.AlignStart),
		rootStyle:      getStyleProp(props, propStyle, style.Style{}),
		separatorStyle: getStyleProp(props, propSeparatorStyle, style.Style{}),
		dirty:          true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}
func (inst *Instance) Destroy()   {}
func (inst *Instance) OnMount()   {}
func (inst *Instance) OnUnmount() {}
func (inst *Instance) MarkDirty() { inst.dirty = true }
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := inst.snapshot()

	inst.key = getStringProp(props, propKey, inst.key)
	inst.direction = getDirectionProp(props, inst.direction)
	inst.primary = getVNodeProp(props, propPrimary, inst.primary)
	inst.secondary = getVNodeProp(props, propSecondary, inst.secondary)
	inst.primarySize = getIntProp(props, propPrimarySize, inst.primarySize)
	inst.secondarySize = getIntProp(props, propSecondarySize, inst.secondarySize)
	inst.primaryFlex = getIntProp(props, propPrimaryFlex, inst.primaryFlex)
	inst.secondaryFlex = getIntProp(props, propSecondaryFlex, inst.secondaryFlex)
	inst.gap = getIntProp(props, propGap, inst.gap)
	inst.separator = getBoolProp(props, propSeparator, inst.separator)
	inst.separatorGlyph = getStringProp(props, propSeparatorGlyph, inst.separatorGlyph)
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.height = getIntProp(props, propHeight, inst.height)
	inst.align = getAlignProp(props, propAlign, inst.align)
	inst.rootStyle = getStyleProp(props, propStyle, inst.rootStyle)
	inst.separatorStyle = getStyleProp(props, propSeparatorStyle, inst.separatorStyle)
	inst.normalize()

	changed := !reflect.DeepEqual(old, inst.snapshot())
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propAlign:          inst.align,
		propDirection:      inst.direction,
		propGap:            inst.gap,
		propHeight:         inst.height,
		propKey:            inst.key,
		propPrimary:        inst.primary,
		propPrimaryFlex:    inst.primaryFlex,
		propPrimarySize:    inst.primarySize,
		propSecondary:      inst.secondary,
		propSecondaryFlex:  inst.secondaryFlex,
		propSecondarySize:  inst.secondarySize,
		propSeparator:      inst.separator,
		propSeparatorGlyph: inst.separatorGlyph,
		propSeparatorStyle: inst.separatorStyle,
		propStyle:          inst.rootStyle,
		propWidth:          inst.width,
	}
}

// RuntimeChildren synthesizes the concrete layout subtree used by Fiber.
func (inst *Instance) RuntimeChildren() []rtui.VNode {
	children := make([]rtui.VNode, 0, 3)
	if inst.primary != nil {
		children = append(children, inst.wrapPane(inst.primary, true))
	}
	if inst.separator && inst.primary != nil && inst.secondary != nil {
		children = append(children, inst.separatorNode())
	}
	if inst.secondary != nil {
		children = append(children, inst.wrapPane(inst.secondary, false))
	}
	if len(children) == 0 {
		return nil
	}

	var root *rtui.LayoutBuilder
	if inst.direction == DirectionVertical {
		root = rtui.VStackBuilder(children...).Gap(inst.gap).AlignCross(inst.align)
	} else {
		root = rtui.HStackBuilder(children...).Gap(inst.gap).AlignCross(inst.align)
	}
	if inst.width > 0 {
		root.Width(inst.width)
	}
	if inst.height > 0 {
		root.Height(inst.height)
	}
	if !inst.rootStyle.IsEmpty() {
		root.SetStyleProps(inst.rootStyle)
	}
	root.SetKey(inst.rootKey())
	return []rtui.VNode{root.Build()}
}

func (inst *Instance) wrapPane(child rtui.VNode, primary bool) rtui.VNode {
	box := rtui.Box().Child(child)
	if primary {
		if inst.direction == DirectionVertical {
			if inst.primarySize > 0 {
				box.Height(inst.primarySize)
			}
		} else if inst.primarySize > 0 {
			box.Width(inst.primarySize)
		}
		if inst.primaryFlex > 0 {
			box.Flex(inst.primaryFlex)
		}
	} else {
		if inst.direction == DirectionVertical {
			if inst.secondarySize > 0 {
				box.Height(inst.secondarySize)
			}
		} else if inst.secondarySize > 0 {
			box.Width(inst.secondarySize)
		}
		if inst.secondaryFlex > 0 {
			box.Flex(inst.secondaryFlex)
		}
	}
	if inst.direction == DirectionVertical {
		if inst.width > 0 {
			box.FillWidth()
		}
	} else if inst.height > 0 {
		box.FillHeight()
	}
	node := box.Build()
	if primary {
		node.SetKey(inst.childKey("primary"))
	} else {
		node.SetKey(inst.childKey("secondary"))
	}
	return node
}

func (inst *Instance) separatorNode() rtui.VNode {
	if inst.direction == DirectionVertical {
		width := inst.width
		if width <= 0 {
			width = 1
		}
		content := repeatGlyph(inst.separatorGlyph, width)
		node := text.NewBuilder(content).Key(inst.childKey("separator")).Style(inst.separatorStyle).Build()
		return node
	}

	height := inst.height
	if height <= 0 {
		height = 1
	}
	lines := make([]rtui.VNode, 0, height)
	for i := 0; i < height; i++ {
		lines = append(lines, text.NewBuilder(inst.separatorGlyph).
			Key(inst.childKey("separator-line-"+strconv.Itoa(i))).
			Style(inst.separatorStyle).
			Build())
	}
	return rtui.VStackBuilder(lines...).Gap(0).Build().SetKey(inst.childKey("separator"))
}

func (inst *Instance) rootKey() string {
	if inst.key == "" {
		return "splitpane-root"
	}
	return inst.key + "-root"
}

func (inst *Instance) childKey(suffix string) string {
	if inst.key == "" {
		return "splitpane-" + suffix
	}
	return inst.key + "-" + suffix
}

func (inst *Instance) normalize() {
	if inst.direction != DirectionVertical {
		inst.direction = DirectionHorizontal
	}
	if inst.primarySize < 0 {
		inst.primarySize = 0
	}
	if inst.secondarySize < 0 {
		inst.secondarySize = 0
	}
	if inst.primaryFlex < 0 {
		inst.primaryFlex = 0
	}
	if inst.secondaryFlex < 0 {
		inst.secondaryFlex = 0
	}
	if inst.gap < 0 {
		inst.gap = 0
	}
	if inst.width < 0 {
		inst.width = 0
	}
	if inst.height < 0 {
		inst.height = 0
	}
	if inst.separatorGlyph == "" {
		inst.separatorGlyph = defaultGlyphForDirection(inst.direction)
	}
}

type instanceSnapshot struct {
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

func (inst *Instance) snapshot() instanceSnapshot {
	return instanceSnapshot{
		key:            inst.key,
		direction:      inst.direction,
		primary:        inst.primary,
		secondary:      inst.secondary,
		primarySize:    inst.primarySize,
		secondarySize:  inst.secondarySize,
		primaryFlex:    inst.primaryFlex,
		secondaryFlex:  inst.secondaryFlex,
		gap:            inst.gap,
		separator:      inst.separator,
		separatorGlyph: inst.separatorGlyph,
		width:          inst.width,
		height:         inst.height,
		align:          inst.align,
		rootStyle:      inst.rootStyle,
		separatorStyle: inst.separatorStyle,
	}
}

func repeatGlyph(glyph string, width int) string {
	if width <= 0 {
		return ""
	}
	if glyph == "" {
		glyph = "─"
	}
	glyphRunes := []rune(glyph)
	if len(glyphRunes) == 0 {
		glyphRunes = []rune("─")
	}
	runes := make([]rune, 0, width)
	for len(runes) < width {
		runes = append(runes, glyphRunes...)
	}
	return string(runes[:width])
}

func getDirectionProp(props rtui.Props, def Direction) Direction {
	if value, ok := props[propDirection]; ok {
		if direction, ok := value.(Direction); ok {
			return direction
		}
	}
	return def
}

func getVNodeProp(props rtui.Props, key string, def rtui.VNode) rtui.VNode {
	if value, ok := props[key]; ok {
		if node, ok := value.(rtui.VNode); ok {
			return node
		}
	}
	return def
}

func getStringProp(props rtui.Props, key, def string) string {
	if value, ok := props[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if value, ok := props[key]; ok {
		if number, ok := value.(int); ok {
			return number
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if value, ok := props[key]; ok {
		if flag, ok := value.(bool); ok {
			return flag
		}
	}
	return def
}

func getAlignProp(props rtui.Props, key string, def rtui.Align) rtui.Align {
	if value, ok := props[key]; ok {
		if align, ok := value.(rtui.Align); ok {
			return align
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string, def style.Style) style.Style {
	if value, ok := props[key]; ok {
		if s, ok := value.(style.Style); ok {
			return s
		}
	}
	return def
}
