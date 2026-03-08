package statusbar

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type helpLineVNode struct {
	*rtui.ElementVNode
}

func newHelpLineVNode(model *helpModel, lineStyle style.Style) *helpLineVNode {
	node := &helpLineVNode{ElementVNode: rtui.NewElement("statusbar-help")}
	node.SetStyle(lineStyle)
	node.SetProp("style", lineStyle)
	node.SetProp("helpModel", model)
	return node
}

func (v *helpLineVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props["style"] = v.Style()
	return newHelpLineInstance(props)
}

type helpLineInstance struct {
	model     *helpModel
	lineStyle style.Style
	bounds    [4]int
	dirty     bool
}

var (
	_ rtui.ComponentInstance = (*helpLineInstance)(nil)
	_ rtui.PaintableInstance = (*helpLineInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*helpLineInstance)(nil)
)

func newHelpLineInstance(props rtui.Props) *helpLineInstance {
	return &helpLineInstance{
		model:     getHelpModelProp(props),
		lineStyle: getSectionStyleProp(props),
		dirty:     true,
	}
}

func (inst *helpLineInstance) Key() string                        { return "" }
func (inst *helpLineInstance) SetKey(key string)                  {}
func (inst *helpLineInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *helpLineInstance) Destroy()                           {}
func (inst *helpLineInstance) OnMount()                           {}
func (inst *helpLineInstance) OnUnmount()                         {}
func (inst *helpLineInstance) GetContext() *rtui.ComponentContext { return nil }
func (inst *helpLineInstance) MarkDirty()                         { inst.dirty = true }
func (inst *helpLineInstance) IsDirty() bool                      { return inst.dirty }

func (inst *helpLineInstance) SetProps(props rtui.Props) bool {
	oldModel := inst.model
	oldStyle := inst.lineStyle
	inst.model = getHelpModelProp(props)
	inst.lineStyle = getSectionStyleProp(props)
	changed := oldModel != inst.model || oldStyle != inst.lineStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *helpLineInstance) GetProps() rtui.Props {
	return rtui.Props{
		"style":     inst.lineStyle,
		"helpModel": inst.model,
	}
}

func (inst *helpLineInstance) Paint(x, y int) []paint.DrawCmd {
	if inst.model == nil {
		return nil
	}
	text := inst.model.Current()
	if text == "" {
		return nil
	}
	if inst.bounds[2] > 0 {
		text = fitText(text, inst.bounds[2], rtui.AlignStart, OverflowClip)
	}
	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  text,
		Style: inst.lineStyle,
	}}
}

func (inst *helpLineInstance) Measure(constraints layout.Constraints) layout.Size {
	text := ""
	if inst.model != nil {
		text = inst.model.Current()
	}
	if text == "" {
		return layout.Size{Width: 0, Height: 1}
	}
	return layout.Size{Width: paint.StringWidth(text), Height: 1}
}

func (inst *helpLineInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *helpLineInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func getHelpModelProp(props rtui.Props) *helpModel {
	if v, ok := props["helpModel"].(*helpModel); ok {
		return v
	}
	return nil
}
