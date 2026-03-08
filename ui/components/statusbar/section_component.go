package statusbar

import (
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
)

type sectionVNode struct {
	*rtui.ElementVNode
}

func newSectionVNode(key, text string, sectionStyle style.Style, pressIntent intent.Intent, disabled bool, helpText, helpKey string, helpOrder int, helpModel *helpModel, hoverStyle, focusStyle, pressedStyle, disabledStyle style.Style) *sectionVNode {
	node := &sectionVNode{ElementVNode: rtui.NewElement("statusbar-section")}
	if key != "" {
		node.SetKey(key)
	}
	node.SetStyle(sectionStyle)
	node.SetProp("style", sectionStyle)
	node.SetProp("text", text)
	node.SetProp("pressIntent", pressIntent)
	node.SetProp("disabled", disabled)
	node.SetProp("helpText", helpText)
	node.SetProp("helpKey", helpKey)
	node.SetProp("helpOrder", helpOrder)
	node.SetProp("helpModel", helpModel)
	node.SetProp("hoverStyle", hoverStyle)
	node.SetProp("focusStyle", focusStyle)
	node.SetProp("pressedStyle", pressedStyle)
	node.SetProp("disabledStyle", disabledStyle)
	return node
}

func (v *sectionVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props["key"] = v.Key()
	props["style"] = v.Style()
	return newSectionInstance(props)
}

type sectionInstance struct {
	key           string
	text          string
	sectionStyle  style.Style
	hoverStyle    style.Style
	focusStyle    style.Style
	pressedStyle  style.Style
	disabledStyle style.Style
	pressIntent   intent.Intent
	helpText      string
	helpKey       string
	helpOrder     int
	helpModel     *helpModel
	state         control.InteractionState
	bounds        [4]int
	dirty         bool
	behaviors     *control.BehaviorList
	intentEmitter func(intent.Intent)
}

var (
	_ rtui.ComponentInstance     = (*sectionInstance)(nil)
	_ rtui.PaintableInstance     = (*sectionInstance)(nil)
	_ rtui.ActionHandlerInstance = (*sectionInstance)(nil)
	_ rtui.FocusableInstance     = (*sectionInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*sectionInstance)(nil)
)

func newSectionInstance(props rtui.Props) *sectionInstance {
	inst := &sectionInstance{
		key:           getSectionStringProp(props, "key", ""),
		text:          getSectionStringProp(props, "text", ""),
		sectionStyle:  getSectionStyleProp(props),
		hoverStyle:    getSectionStylePropKey(props, "hoverStyle"),
		focusStyle:    getSectionStylePropKey(props, "focusStyle"),
		pressedStyle:  getSectionStylePropKey(props, "pressedStyle"),
		disabledStyle: getSectionStylePropKey(props, "disabledStyle"),
		pressIntent:   getSectionIntentProp(props),
		helpText:      getSectionStringProp(props, "helpText", ""),
		helpKey:       getSectionStringProp(props, "helpKey", ""),
		helpOrder:     getSectionIntProp(props, "helpOrder", 0),
		helpModel:     getHelpModelProp(props),
		dirty:         true,
	}
	inst.state.Disabled = getSectionBoolProp(props, "disabled", false)
	inst.initBehaviors()
	inst.syncHelpModel()
	return inst
}

func (inst *sectionInstance) initBehaviors() {
	var behaviors []control.Behavior
	if inst.pressIntent != nil {
		behaviors = append(behaviors, &control.FocusableBehavior{})
	}
	if inst.pressIntent != nil || inst.helpText != "" {
		behaviors = append(behaviors, &control.HoverableBehavior{})
	}
	if inst.pressIntent != nil {
		behaviors = append(behaviors, control.NewPressableBehavior(inst.pressIntent))
	}
	inst.behaviors = control.NewBehaviorList(behaviors...)
}

func (inst *sectionInstance) syncHelpModel() {
	if inst.helpModel == nil || inst.helpKey == "" {
		return
	}
	if inst.helpText == "" || inst.state.Disabled {
		inst.helpModel.Remove(inst.helpKey)
		return
	}
	inst.helpModel.Update(inst.helpKey, inst.helpOrder, inst.helpText, inst.state.Hovered, inst.state.Focused)
}

func (inst *sectionInstance) clearHelpModel() {
	if inst.helpModel != nil && inst.helpKey != "" {
		inst.helpModel.Remove(inst.helpKey)
	}
}

func (inst *sectionInstance) Key() string { return inst.key }

func (inst *sectionInstance) SetKey(key string) { inst.key = key }

func (inst *sectionInstance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *sectionInstance) Destroy() {
	inst.clearHelpModel()
	if inst.behaviors != nil {
		inst.behaviors.OnUnmount(inst)
	}
}

func (inst *sectionInstance) OnMount() {
	if inst.behaviors != nil {
		inst.behaviors.OnMount(inst)
	}
}

func (inst *sectionInstance) OnUnmount() {
	inst.clearHelpModel()
	if inst.behaviors != nil {
		inst.behaviors.OnUnmount(inst)
	}
}

func (inst *sectionInstance) SetProps(props rtui.Props) bool {
	oldText := inst.text
	oldStyle := inst.sectionStyle
	oldHoverStyle := inst.hoverStyle
	oldFocusStyle := inst.focusStyle
	oldPressedStyle := inst.pressedStyle
	oldDisabledStyle := inst.disabledStyle
	oldIntent := inst.pressIntent
	oldDisabled := inst.state.Disabled
	oldHelpModel := inst.helpModel
	oldHelpKey := inst.helpKey
	oldHelpText := inst.helpText

	inst.key = getSectionStringProp(props, "key", inst.key)
	inst.text = getSectionStringProp(props, "text", inst.text)
	inst.sectionStyle = getSectionStyleProp(props)
	inst.hoverStyle = getSectionStylePropKey(props, "hoverStyle")
	inst.focusStyle = getSectionStylePropKey(props, "focusStyle")
	inst.pressedStyle = getSectionStylePropKey(props, "pressedStyle")
	inst.disabledStyle = getSectionStylePropKey(props, "disabledStyle")
	inst.pressIntent = getSectionIntentProp(props)
	inst.state.Disabled = getSectionBoolProp(props, "disabled", inst.state.Disabled)
	inst.helpText = getSectionStringProp(props, "helpText", inst.helpText)
	inst.helpKey = getSectionStringProp(props, "helpKey", inst.helpKey)
	inst.helpOrder = getSectionIntProp(props, "helpOrder", inst.helpOrder)
	inst.helpModel = getHelpModelProp(props)

	if inst.behaviors == nil || oldIntent != inst.pressIntent || oldHelpText != inst.helpText {
		inst.initBehaviors()
	}
	if pressable := inst.behaviors.Get("Pressable"); pressable != nil {
		if behavior, ok := pressable.(*control.PressableBehavior); ok {
			behavior.SetIntent(inst.pressIntent)
		}
	}
	if oldHelpModel != nil && (oldHelpModel != inst.helpModel || oldHelpKey != inst.helpKey) {
		oldHelpModel.Remove(oldHelpKey)
	}
	inst.syncHelpModel()

	changed := oldText != inst.text ||
		oldStyle != inst.sectionStyle ||
		oldHoverStyle != inst.hoverStyle ||
		oldFocusStyle != inst.focusStyle ||
		oldPressedStyle != inst.pressedStyle ||
		oldDisabledStyle != inst.disabledStyle ||
		oldIntent != inst.pressIntent ||
		oldDisabled != inst.state.Disabled ||
		oldHelpText != inst.helpText ||
		oldHelpModel != inst.helpModel ||
		oldHelpKey != inst.helpKey
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *sectionInstance) GetProps() rtui.Props {
	return rtui.Props{
		"key":           inst.key,
		"text":          inst.text,
		"style":         inst.sectionStyle,
		"hoverStyle":    inst.hoverStyle,
		"focusStyle":    inst.focusStyle,
		"pressedStyle":  inst.pressedStyle,
		"disabledStyle": inst.disabledStyle,
		"pressIntent":   inst.pressIntent,
		"disabled":      inst.state.Disabled,
		"helpText":      inst.helpText,
		"helpKey":       inst.helpKey,
		"helpOrder":     inst.helpOrder,
		"helpModel":     inst.helpModel,
	}
}

func (inst *sectionInstance) MarkDirty() {
	inst.syncHelpModel()
	inst.dirty = true
}

func (inst *sectionInstance) IsDirty() bool { return inst.dirty }

func (inst *sectionInstance) GetContext() *rtui.ComponentContext { return nil }

func (inst *sectionInstance) Paint(x, y int) []paint.DrawCmd {
	resolvedStyle := inst.sectionStyle
	if inst.state.Disabled {
		resolvedStyle = resolvedStyle.Merge(inst.disabledStyle)
	} else {
		if inst.state.Hovered {
			resolvedStyle = resolvedStyle.Merge(inst.hoverStyle)
		}
		if inst.state.Focused {
			resolvedStyle = resolvedStyle.Merge(inst.focusStyle)
		}
		if inst.state.Pressed {
			resolvedStyle = resolvedStyle.Merge(inst.pressedStyle)
		}
	}
	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  inst.text,
		Style: resolvedStyle,
	}}
}

func (inst *sectionInstance) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{Width: paint.StringWidth(inst.text), Height: 1}
}

func (inst *sectionInstance) HandleAction(act *action.Action) bool {
	if inst.behaviors == nil {
		return false
	}
	return inst.behaviors.OnAction(inst, act)
}

func (inst *sectionInstance) SetFocus(focused bool) {
	if inst.state.Focused == focused {
		return
	}
	oldState := inst.state
	inst.state.Focused = focused
	inst.dirty = true
	if inst.behaviors != nil {
		inst.behaviors.OnStateChange(inst, oldState, inst.state)
	}
	inst.syncHelpModel()
}

func (inst *sectionInstance) HasFocus() bool {
	return inst.state.Focused
}

func (inst *sectionInstance) IsDisabled() bool {
	return inst.state.Disabled
}

func (inst *sectionInstance) IsFocusable() bool {
	return inst.pressIntent != nil && !inst.state.Disabled
}

func (inst *sectionInstance) GetState() *control.InteractionState { return &inst.state }

func (inst *sectionInstance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	if inst.behaviors != nil {
		inst.behaviors.OnStateChange(inst, oldState, inst.state)
	}
	inst.syncHelpModel()
}

func (inst *sectionInstance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

func (inst *sectionInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *sectionInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *sectionInstance) GetStyle() style.Style { return inst.sectionStyle }

func (inst *sectionInstance) SetStyle(s style.Style) { inst.sectionStyle = s }

func (inst *sectionInstance) GetProp(key string) (interface{}, bool) {
	switch key {
	case "text":
		return inst.text, true
	case "pressIntent":
		return inst.pressIntent, true
	case "disabled":
		return inst.state.Disabled, true
	case "helpText":
		return inst.helpText, true
	default:
		return nil, false
	}
}

func (inst *sectionInstance) SetProp(key string, value interface{}) {
	switch key {
	case "text":
		if v, ok := value.(string); ok {
			inst.text = v
		}
	case "pressIntent":
		if v, ok := value.(intent.Intent); ok {
			inst.pressIntent = v
			inst.initBehaviors()
		}
	case "disabled":
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
		}
	case "helpText":
		if v, ok := value.(string); ok {
			inst.helpText = v
		}
	}
	inst.syncHelpModel()
}

func (inst *sectionInstance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func getSectionStringProp(props rtui.Props, key, fallback string) string {
	if v, ok := props[key].(string); ok {
		return v
	}
	return fallback
}

func getSectionBoolProp(props rtui.Props, key string, fallback bool) bool {
	if v, ok := props[key].(bool); ok {
		return v
	}
	return fallback
}

func getSectionIntProp(props rtui.Props, key string, fallback int) int {
	if v, ok := props[key].(int); ok {
		return v
	}
	return fallback
}

func getSectionStyleProp(props rtui.Props) style.Style {
	return getSectionStylePropKey(props, "style")
}

func getSectionStylePropKey(props rtui.Props, key string) style.Style {
	if v, ok := props[key].(style.Style); ok {
		return v
	}
	return style.Style{}
}

func getSectionIntentProp(props rtui.Props) intent.Intent {
	if v, ok := props["pressIntent"].(intent.Intent); ok {
		return v
	}
	return nil
}
