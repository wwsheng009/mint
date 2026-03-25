package space

import (
	"reflect"
	"strconv"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
	wrapcomp "github.com/wwsheng009/mint/ui/components/wrap"
)

// Instance is the runtime entity for Space components.
type Instance struct {
	key       string
	direction Direction
	size      int
	wrap      bool
	width     int
	align     Align
	split     string
	children  []rtui.VNode
	rootStyle style.Style
	dirty     bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a new Space instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:       getStringProp(props, propKey, ""),
		direction: getDirectionProp(props, DirectionHorizontal),
		size:      getIntProp(props, propSize, SizeSmall),
		wrap:      getBoolProp(props, propWrap, false),
		width:     getIntProp(props, propWidth, 0),
		align:     getAlignProp(props, AlignStart),
		split:     getStringProp(props, propSplit, ""),
		children:  getChildrenProp(props),
		rootStyle: getStyleProp(props, propStyle, style.Style{}),
		dirty:     true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              {}
func (inst *Instance) OnMount()              {}
func (inst *Instance) OnUnmount()            {}
func (inst *Instance) MarkDirty()            { inst.dirty = true }
func (inst *Instance) IsDirty() bool         { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldChildren := append([]rtui.VNode(nil), inst.children...)
	oldDirection := inst.direction
	oldSize := inst.size
	oldWrap := inst.wrap
	oldWidth := inst.width
	oldAlign := inst.align
	oldSplit := inst.split
	oldStyle := inst.rootStyle

	inst.key = getStringProp(props, propKey, inst.key)
	inst.direction = getDirectionPropWithDefault(props, inst.direction)
	inst.size = getIntProp(props, propSize, inst.size)
	inst.wrap = getBoolProp(props, propWrap, inst.wrap)
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.align = getAlignPropWithDefault(props, inst.align)
	inst.split = getStringProp(props, propSplit, inst.split)
	inst.children = getChildrenPropWithDefault(props, inst.children)
	inst.rootStyle = getStyleProp(props, propStyle, inst.rootStyle)
	inst.normalize()

	changed := oldDirection != inst.direction ||
		oldSize != inst.size ||
		oldWrap != inst.wrap ||
		oldWidth != inst.width ||
		oldAlign != inst.align ||
		oldSplit != inst.split ||
		oldStyle != inst.rootStyle ||
		!reflect.DeepEqual(oldChildren, inst.children)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propAlign:     inst.align,
		propChildren:  append([]rtui.VNode(nil), inst.children...),
		propDirection: inst.direction,
		propKey:       inst.key,
		propSize:      inst.size,
		propSplit:     inst.split,
		propStyle:     inst.rootStyle,
		propWidth:     inst.width,
		propWrap:      inst.wrap,
	}
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	content := inst.runtimeChildren()
	if len(content) == 0 {
		return nil
	}

	var node rtui.VNode
	switch {
	case inst.direction == DirectionVertical:
		builder := rtui.VStackBuilder(content...).Gap(inst.size).AlignCross(inst.align)
		if inst.width > 0 {
			builder.Width(inst.width)
		}
		if !inst.rootStyle.IsEmpty() {
			builder.SetStyleProps(inst.rootStyle)
		}
		node = builder.Build()
	case inst.wrap:
		builder := wrapcomp.NewBuilder().Gap(inst.size).Children(content...).Align(inst.align)
		if inst.width > 0 {
			builder.Width(inst.width)
		}
		if !inst.rootStyle.IsEmpty() {
			builder.Style(inst.rootStyle)
		}
		node = builder.Build()
	default:
		builder := rtui.HStackBuilder(content...).Gap(inst.size).AlignCross(inst.align)
		if inst.width > 0 {
			builder.Width(inst.width)
		}
		if !inst.rootStyle.IsEmpty() {
			builder.SetStyleProps(inst.rootStyle)
		}
		node = builder.Build()
	}

	node.SetKey(inst.rootKey())
	return []rtui.VNode{node}
}

func (inst *Instance) runtimeChildren() []rtui.VNode {
	filtered := make([]rtui.VNode, 0, len(inst.children))
	for _, child := range inst.children {
		if child != nil {
			filtered = append(filtered, child)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	if inst.split == "" {
		return filtered
	}

	result := make([]rtui.VNode, 0, len(filtered)*2-1)
	for index, child := range filtered {
		if index > 0 {
			separator := textcomp.New(inst.split)
			separator.SetKey(inst.key + "-split-" + strconv.Itoa(index-1))
			result = append(result, separator)
		}
		result = append(result, child)
	}
	return result
}

func (inst *Instance) rootKey() string {
	if inst.key == "" {
		return "space-root"
	}
	return inst.key + "-root"
}

func (inst *Instance) normalize() {
	if inst.size < 0 {
		inst.size = 0
	}
	if inst.direction != DirectionVertical {
		inst.direction = DirectionHorizontal
	}
}

func getDirectionProp(props rtui.Props, def Direction) Direction {
	if value, ok := props[propDirection]; ok {
		if direction, ok := value.(Direction); ok {
			return direction
		}
	}
	return def
}

func getDirectionPropWithDefault(props rtui.Props, def Direction) Direction {
	return getDirectionProp(props, def)
}

func getAlignProp(props rtui.Props, def Align) Align {
	if value, ok := props[propAlign]; ok {
		if align, ok := value.(Align); ok {
			return align
		}
	}
	return def
}

func getAlignPropWithDefault(props rtui.Props, def Align) Align {
	return getAlignProp(props, def)
}

func getChildrenProp(props rtui.Props) []rtui.VNode {
	if value, ok := props[propChildren]; ok {
		if children, ok := value.([]rtui.VNode); ok {
			return append([]rtui.VNode(nil), children...)
		}
	}
	return nil
}

func getChildrenPropWithDefault(props rtui.Props, def []rtui.VNode) []rtui.VNode {
	if children := getChildrenProp(props); children != nil {
		return children
	}
	return append([]rtui.VNode(nil), def...)
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

func getStyleProp(props rtui.Props, key string, def style.Style) style.Style {
	if value, ok := props[key]; ok {
		if s, ok := value.(style.Style); ok {
			return s
		}
	}
	return def
}
