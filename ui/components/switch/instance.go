package switchcomp

import (
	"strings"

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
)

// Instance is the runtime entity for Switch components.
type Instance struct {
	key string

	label          string
	switchStyle    style.Style
	checkedLabel   string
	uncheckedLabel string
	toggleIntent   intent.Intent
	formID         string

	state   control.InteractionState
	checked bool
	bounds  [4]int
	dirty   bool

	intentEmitter func(intent.Intent)
	behaviors     *control.BehaviorList
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

// NewInstance creates a new Switch instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:            proputil.GetString(props, propKey, ""),
		label:          proputil.GetString(props, propLabel, ""),
		switchStyle:    proputil.GetStyle(props, propStyle, style.Style{}),
		checkedLabel:   normalizeStateLabel(proputil.GetString(props, propCheckedLabel, defaultCheckedLabel), defaultCheckedLabel),
		uncheckedLabel: normalizeStateLabel(proputil.GetString(props, propUncheckedLabel, defaultUncheckedLabel), defaultUncheckedLabel),
		toggleIntent:   proputil.GetIntent(props, propToggleIntent, nil),
		formID:         proputil.GetString(props, propFormID, ""),
		checked:        proputil.GetBool(props, propChecked, false),
		dirty:          true,
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
	oldToggleIntent := inst.toggleIntent
	oldCheckedLabel := inst.checkedLabel
	oldUncheckedLabel := inst.uncheckedLabel

	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.switchStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.checkedLabel = normalizeStateLabel(
		proputil.GetString(props, propCheckedLabel, inst.checkedLabel),
		defaultCheckedLabel,
	)
	inst.uncheckedLabel = normalizeStateLabel(
		proputil.GetString(props, propUncheckedLabel, inst.uncheckedLabel),
		defaultUncheckedLabel,
	)
	inst.toggleIntent = proputil.GetIntent(props, propToggleIntent, nil)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)
	inst.checked = proputil.GetBool(props, propChecked, inst.checked)

	newDisabled := proputil.GetBool(props, propDisabled, inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	changed := oldLabel != inst.label ||
		oldDisabled != inst.state.Disabled ||
		oldChecked != inst.checked ||
		oldToggleIntent != inst.toggleIntent ||
		oldCheckedLabel != inst.checkedLabel ||
		oldUncheckedLabel != inst.uncheckedLabel

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propChecked:        inst.checked,
		propCheckedLabel:   inst.checkedLabel,
		propDisabled:       inst.state.Disabled,
		propKey:            inst.key,
		propLabel:          inst.label,
		propUncheckedLabel: inst.uncheckedLabel,
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
	track := inst.renderTrack()
	cmds := []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  track,
		Style: inst.resolveTrackStyle(),
	}}

	if inst.label != "" {
		cmds = append(cmds, paint.DrawCmd{
			X:     x + paint.StringWidth(track) + 1,
			Y:     y,
			Text:  inst.label,
			Style: inst.resolveLabelStyle(),
		})
	}

	return cmds
}

func (inst *Instance) resolveTrackStyle() style.Style {
	s := inst.switchStyle

	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	if s.FG == "" {
		if inst.checked {
			s = s.Foreground(theme.Success())
		} else {
			s = s.Foreground(theme.Muted())
		}
	}

	if inst.checked && !inst.state.Disabled {
		s = s.Bold(true)
	}

	if inst.state.Disabled {
		s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	} else if inst.state.Focused {
		s = s.Foreground(theme.Focus()).Bold(true).Underline(true)
	} else if inst.state.Hovered {
		s = s.Underline(true)
	}

	return s
}

func (inst *Instance) resolveLabelStyle() style.Style {
	s := inst.switchStyle

	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
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

func (inst *Instance) renderTrack() string {
	checkedLabel := normalizeStateLabel(inst.checkedLabel, defaultCheckedLabel)
	uncheckedLabel := normalizeStateLabel(inst.uncheckedLabel, defaultUncheckedLabel)

	labelWidth := paint.StringWidth(checkedLabel)
	if altWidth := paint.StringWidth(uncheckedLabel); altWidth > labelWidth {
		labelWidth = altWidth
	}
	if labelWidth < 2 {
		labelWidth = 2
	}

	innerWidth := labelWidth + 2

	var content string
	if inst.checked {
		content = checkedLabel + strings.Repeat(" ", innerWidth-paint.StringWidth(checkedLabel))
	} else {
		content = strings.Repeat(" ", innerWidth-paint.StringWidth(uncheckedLabel)) + uncheckedLabel
	}

	return "[" + content + "]"
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
		inst.Toggle()
		return true
	}
	return false
}

// Toggle toggles the checked state and returns the new value.
func (inst *Instance) Toggle() bool {
	inst.checked = !inst.checked
	inst.dirty = true
	inst.emitFieldValueChanged()
	return inst.checked
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

// Label returns the switch label text.
func (inst *Instance) Label() string {
	return inst.label
}

// CheckedLabel returns the checked-state label.
func (inst *Instance) CheckedLabel() string {
	return inst.checkedLabel
}

// UncheckedLabel returns the unchecked-state label.
func (inst *Instance) UncheckedLabel() string {
	return inst.uncheckedLabel
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
	return inst.switchStyle
}

// SetStyle sets the visual style.
func (inst *Instance) SetStyle(st style.Style) {
	inst.switchStyle = st
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
	case propToggleIntent:
		return inst.toggleIntent, true
	case propCheckedLabel:
		return inst.checkedLabel, true
	case propUncheckedLabel:
		return inst.uncheckedLabel, true
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
	case propCheckedLabel:
		if v, ok := value.(string); ok {
			inst.checkedLabel = normalizeStateLabel(v, defaultCheckedLabel)
			inst.dirty = true
		}
	case propUncheckedLabel:
		if v, ok := value.(string); ok {
			inst.uncheckedLabel = normalizeStateLabel(v, defaultUncheckedLabel)
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

	contentWidth := paint.StringWidth(inst.renderTrack())
	if inst.label != "" {
		contentWidth += 1 + paint.StringWidth(inst.label)
	}

	width := constraints.ConstrainWidth(contentWidth)
	height := constraints.ConstrainHeight(1)

	if inst.switchStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.switchStyle.Width)
	}
	if inst.switchStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.switchStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// GetNaturalSize returns the natural size.
func (inst *Instance) GetNaturalSize() (width, height int) {
	width = paint.StringWidth(inst.renderTrack())
	if inst.label != "" {
		width += 1 + paint.StringWidth(inst.label)
	}
	return width, 1
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
		if fieldIntent, ok := inst.toggleIntent.(intent.FieldIntent); ok {
			intent.Emit(inst, form.FieldChange(
				inst.formID,
				fieldIntent.GetField(),
				value,
				true,
			))
		}
		return
	}

	if fieldIntent, ok := inst.toggleIntent.(intent.FieldIntent); ok {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: fieldIntent.GetField(),
			Value: value,
		})
	} else if inst.toggleIntent != nil {
		inst.intentEmitter(inst.toggleIntent)
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

	if fieldIntent, ok := inst.toggleIntent.(intent.FieldIntent); ok {
		intent.Emit(inst, form.FieldBlur(
			inst.formID,
			fieldIntent.GetField(),
			value,
		))
	}
}

func normalizeStateLabel(label, fallback string) string {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
