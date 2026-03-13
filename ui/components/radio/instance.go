package radio

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"github.com/wwsheng009/mint/ui/components/optiongroup"
)

// Instance is the runtime entity for Radio components.
type Instance struct {
	key string

	label        string
	radioStyle   style.Style
	selectIntent intent.Intent
	formID       string

	state   control.InteractionState
	checked bool
	bounds  [4]int
	dirty   bool

	intentEmitter func(intent.Intent)

	behaviors *control.BehaviorList
}

var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ control.Instance           = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// GroupInstance is the RadioGroup runtime entity.
type GroupInstance struct {
	*optiongroup.Instance
}

// NewInstance creates a new Radio instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          proputil.GetString(props, propKey, ""),
		label:        proputil.GetString(props, propLabel, ""),
		radioStyle:   proputil.GetStyle(props, propStyle, style.Style{}),
		selectIntent: proputil.GetIntent(props, propSelectIntent, nil),
		formID:       proputil.GetString(props, propFormID, ""),
		checked:      proputil.GetBool(props, propChecked, false),
		dirty:        true,
	}

	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, propDisabled, false),
	}

	inst.initBehaviors()

	return inst
}

func (inst *Instance) initBehaviors() {
	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
}

// Key implements ComponentInstance.
func (inst *Instance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *Instance) SetKey(key string) {
	inst.key = key
}

// Parent implements TreeComponent.
func (inst *Instance) Parent() interface{} {
	return nil
}

// Init implements ComponentInstance.
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy implements ComponentInstance.
func (inst *Instance) Destroy() {
	inst.behaviors.OnUnmount(inst)
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {
	inst.behaviors.OnMount(inst)
}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {
	inst.behaviors.OnUnmount(inst)
}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldLabel := inst.label
	oldDisabled := inst.state.Disabled
	oldChecked := inst.checked
	oldIntent := inst.selectIntent

	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.radioStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.selectIntent = proputil.GetIntent(props, propSelectIntent, nil)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)
	inst.checked = proputil.GetBool(props, propChecked, inst.checked)

	newDisabled := proputil.GetBool(props, propDisabled, inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	changed := oldLabel != inst.label ||
		oldDisabled != inst.state.Disabled ||
		oldChecked != inst.checked ||
		oldIntent != inst.selectIntent

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:      inst.key,
		propLabel:    inst.label,
		propDisabled: inst.state.Disabled,
		propChecked:  inst.checked,
	}
}

// MarkDirty implements ComponentInstance.
func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

// IsDirty implements ComponentInstance.
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

// GetContext implements ComponentInstance.
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	indicator := "( )"
	if inst.checked {
		indicator = "(*)"
	}

	displayText := indicator
	if inst.label != "" {
		displayText += " " + inst.label
	}

	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  displayText,
		Style: inst.resolveStyle(),
	}}
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.radioStyle

	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	if inst.checked && !inst.state.Disabled {
		s = s.Bold(true)
	}

	if inst.state.Disabled {
		s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	} else if inst.state.Focused {
		s = s.Foreground(theme.Focus()).Bold(true)
	} else if inst.state.Hovered {
		s = s.Underline(true)
	}

	return s
}

// SetFocus implements FocusableInstance.
func (inst *Instance) SetFocus(focused bool) {
	if inst.state.Focused != focused {
		oldState := inst.state
		wasFocused := inst.state.Focused
		inst.state.Focused = focused
		if wasFocused && !focused {
			inst.emitFieldBlur()
		}
		inst.dirty = true
		inst.behaviors.OnStateChange(inst, oldState, inst.state)
	}
}

// HasFocus implements FocusableInstance.
func (inst *Instance) HasFocus() bool {
	return inst.state.Focused
}

// IsDisabled implements FocusableInstance.
func (inst *Instance) IsDisabled() bool {
	return inst.state.Disabled
}

// HandleAction implements ActionHandlerInstance.
func (inst *Instance) HandleAction(act *action.Action) bool {
	if inst.state.Disabled {
		return false
	}

	switch act.Type {
	case action.ActionToggle, action.ActionClick, action.ActionEnter:
		inst.Select()
		return true
	}
	return false
}

// Select marks the radio as selected.
func (inst *Instance) Select() bool {
	if inst.checked {
		return true
	}

	inst.checked = true
	inst.dirty = true
	inst.emitFieldValueChanged()

	return true
}

// SetChecked sets the checked state.
func (inst *Instance) SetChecked(checked bool) {
	if inst.checked != checked {
		inst.checked = checked
		inst.dirty = true
	}
}

// IsChecked returns the checked state.
func (inst *Instance) IsChecked() bool {
	return inst.checked
}

// Label returns the label text.
func (inst *Instance) Label() string {
	return inst.label
}

// GetState returns the interaction state.
func (inst *Instance) GetState() *control.InteractionState {
	return &inst.state
}

// SetState sets the interaction state.
func (inst *Instance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

// EmitIntent emits an intent.
func (inst *Instance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// GetStyle returns the visual style.
func (inst *Instance) GetStyle() style.Style {
	return inst.radioStyle
}

// SetStyle sets the visual style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.radioStyle = s
}

// GetProp returns a prop value.
func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case propDisabled:
		return inst.state.Disabled, true
	case propChecked:
		return inst.checked, true
	case propLabel:
		return inst.label, true
	case propSelectIntent:
		return inst.selectIntent, true
	default:
		return nil, false
	}
}

// SetProp sets a prop value.
func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case propDisabled:
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.dirty = true
		}
	case propChecked:
		if v, ok := value.(bool); ok {
			inst.checked = v
			inst.dirty = true
		}
	}
}

// SetIntentEmitter sets the intent emitter function.
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

// Measure implements layout.Measurable.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	contentWidth := 4 + utf8.RuneCountInString(inst.label)
	contentHeight := 1

	width := constraints.ConstrainWidth(contentWidth)
	height := constraints.ConstrainHeight(contentHeight)

	if inst.radioStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.radioStyle.Width)
	}
	if inst.radioStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.radioStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// GetNaturalSize returns the natural size.
func (inst *Instance) GetNaturalSize() (width, height int) {
	return 4 + utf8.RuneCountInString(inst.label), 1
}

func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}

	value := "false"
	if inst.checked {
		value = "true"
	}

	if inst.formID != "" {
		if fieldIntent, ok := inst.selectIntent.(intent.FieldIntent); ok {
			formIntent := form.FieldChange(
				inst.formID,
				fieldIntent.GetField(),
				value,
				true,
			)
			intent.Emit(inst, formIntent)
		}
		return
	}

	if fieldIntent, ok := inst.selectIntent.(intent.FieldIntent); ok {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: fieldIntent.GetField(),
			Value: value,
		})
	} else if inst.selectIntent != nil {
		inst.intentEmitter(inst.selectIntent)
	}
}

func (inst *Instance) emitFieldBlur() {
	if inst.intentEmitter == nil || inst.formID == "" {
		return
	}

	value := "false"
	if inst.checked {
		value = "true"
	}

	if fieldIntent, ok := inst.selectIntent.(intent.FieldIntent); ok {
		intent.Emit(inst, form.FieldBlur(
			inst.formID,
			fieldIntent.GetField(),
			value,
		))
	}
}
