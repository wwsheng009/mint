package layout

import (
	"reflect"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Instance is the runtime entity for Layout components.
type Instance struct {
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
	dirty      bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a new Layout instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:        getStringProp(props, propKey, ""),
		header:     getVNodeProp(props, propHeader),
		leftSider:  getVNodeProp(props, propLeftSider),
		content:    getVNodeProp(props, propContent),
		rightSider: getVNodeProp(props, propRightSider),
		footer:     getVNodeProp(props, propFooter),
		sectionGap: getIntProp(props, propSectionGap, 0),
		bodyGap:    getIntProp(props, propBodyGap, 0),
		width:      getIntProp(props, propWidth, 0),
		height:     getIntProp(props, propHeight, 0),
		rootStyle:  getStyleProp(props, propStyle, style.Style{}),
		dirty:      true,
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
	old := *inst

	inst.key = getStringProp(props, propKey, inst.key)
	inst.header = getVNodePropWithDefault(props, propHeader, inst.header)
	inst.leftSider = getVNodePropWithDefault(props, propLeftSider, inst.leftSider)
	inst.content = getVNodePropWithDefault(props, propContent, inst.content)
	inst.rightSider = getVNodePropWithDefault(props, propRightSider, inst.rightSider)
	inst.footer = getVNodePropWithDefault(props, propFooter, inst.footer)
	inst.sectionGap = getIntProp(props, propSectionGap, inst.sectionGap)
	inst.bodyGap = getIntProp(props, propBodyGap, inst.bodyGap)
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.height = getIntProp(props, propHeight, inst.height)
	inst.rootStyle = getStyleProp(props, propStyle, inst.rootStyle)
	inst.normalize()

	changed := old.key != inst.key ||
		!reflect.DeepEqual(old.header, inst.header) ||
		!reflect.DeepEqual(old.leftSider, inst.leftSider) ||
		!reflect.DeepEqual(old.content, inst.content) ||
		!reflect.DeepEqual(old.rightSider, inst.rightSider) ||
		!reflect.DeepEqual(old.footer, inst.footer) ||
		old.sectionGap != inst.sectionGap ||
		old.bodyGap != inst.bodyGap ||
		old.width != inst.width ||
		old.height != inst.height ||
		old.rootStyle != inst.rootStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propBodyGap:    inst.bodyGap,
		propContent:    inst.content,
		propFooter:     inst.footer,
		propHeader:     inst.header,
		propHeight:     inst.height,
		propKey:        inst.key,
		propLeftSider:  inst.leftSider,
		propRightSider: inst.rightSider,
		propStyle:      inst.rootStyle,
		propSectionGap: inst.sectionGap,
		propWidth:      inst.width,
	}
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	rootChildren := make([]rtui.VNode, 0, 3)

	if inst.header != nil {
		rootChildren = append(rootChildren, inst.header)
	}

	body := inst.bodyNode()
	if body != nil {
		rootChildren = append(rootChildren, rtui.Flex(body, 1))
	}

	if inst.footer != nil {
		rootChildren = append(rootChildren, inst.footer)
	}

	if len(rootChildren) == 0 {
		return nil
	}

	root := rtui.VStackBuilder(rootChildren...).Gap(inst.sectionGap).AlignCross(rtui.AlignStart).Stretch()
	if inst.width > 0 {
		root.Width(inst.width)
	}
	if inst.height > 0 {
		root.Height(inst.height)
	}
	if !inst.rootStyle.IsEmpty() {
		root.SetStyleProps(inst.rootStyle)
	}

	node := root.Build()
	node.SetKey(inst.rootKey())
	return []rtui.VNode{node}
}

func (inst *Instance) bodyNode() rtui.VNode {
	bodyChildren := make([]rtui.VNode, 0, 3)
	if inst.leftSider != nil {
		bodyChildren = append(bodyChildren, inst.leftSider)
	}
	if inst.content != nil {
		bodyChildren = append(bodyChildren, inst.contentNode())
	}
	if inst.rightSider != nil {
		bodyChildren = append(bodyChildren, inst.rightSider)
	}
	if len(bodyChildren) == 0 {
		return nil
	}

	body := rtui.HStackBuilder(bodyChildren...).Gap(inst.bodyGap).AlignCross(rtui.AlignStart).Stretch()
	node := body.Build()
	node.SetKey(inst.bodyKey())
	return node
}

func (inst *Instance) contentNode() rtui.VNode {
	content := rtui.VStackBuilder(inst.content).Gap(0).Stretch().Flex(1)
	node := content.Build()
	node.SetKey(inst.contentKey())
	return node
}

func (inst *Instance) rootKey() string {
	if inst.key == "" {
		return "layout-root"
	}
	return inst.key + "-root"
}

func (inst *Instance) bodyKey() string {
	if inst.key == "" {
		return "layout-body"
	}
	return inst.key + "-body"
}

func (inst *Instance) contentKey() string {
	if inst.key == "" {
		return "layout-content"
	}
	return inst.key + "-content"
}

func (inst *Instance) normalize() {
	if inst.sectionGap < 0 {
		inst.sectionGap = 0
	}
	if inst.bodyGap < 0 {
		inst.bodyGap = 0
	}
	if inst.width < 0 {
		inst.width = 0
	}
	if inst.height < 0 {
		inst.height = 0
	}
}

func getVNodeProp(props rtui.Props, key string) rtui.VNode {
	if value, ok := props[key]; ok {
		if node, ok := value.(rtui.VNode); ok {
			return node
		}
	}
	return nil
}

func getVNodePropWithDefault(props rtui.Props, key string, def rtui.VNode) rtui.VNode {
	if node := getVNodeProp(props, key); node != nil {
		return node
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

func getStyleProp(props rtui.Props, key string, def style.Style) style.Style {
	if value, ok := props[key]; ok {
		if s, ok := value.(style.Style); ok {
			return s
		}
	}
	return def
}
