package rate

import (
	"fmt"
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

const emptyCharacter = "☆"

// Instance is the runtime entity for Rate components.
type Instance struct {
	key          string
	label        string
	rateStyle    style.Style
	changeIntent intent.Intent
	formID       string

	value      int
	count      int
	allowClear bool
	character  string
	showValue  bool

	state  control.InteractionState
	bounds [4]int
	dirty  bool

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

// NewInstance creates a new Rate instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          proputil.GetString(props, propKey, ""),
		label:        proputil.GetString(props, propLabel, ""),
		rateStyle:    proputil.GetStyle(props, propStyle, style.Style{}),
		changeIntent: proputil.GetIntent(props, propChangeIntent, nil),
		formID:       proputil.GetString(props, propFormID, ""),
		value:        proputil.GetInt(props, propValue, 0),
		count:        proputil.GetInt(props, propCount, defaultCount),
		allowClear:   proputil.GetBool(props, propAllowClear, true),
		character:    normalizeCharacter(proputil.GetString(props, propCharacter, defaultCharacter)),
		showValue:    proputil.GetBool(props, propShowValue, false),
		dirty:        true,
	}

	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, propDisabled, false),
	}

	inst.count = clampValue(inst.count, 1, 99)
	inst.value = clampValue(inst.value, 0, inst.count)
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

// --- ComponentInstance ---

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Parent() interface{}   { return nil }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.behaviors.OnUnmount(inst) }
func (inst *Instance) OnMount()              { inst.behaviors.OnMount(inst) }
func (inst *Instance) OnUnmount()            { inst.behaviors.OnUnmount(inst) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldValue := inst.value
	oldCount := inst.count
	oldAllowClear := inst.allowClear
	oldCharacter := inst.character
	oldLabel := inst.label
	oldShowValue := inst.showValue
	oldDisabled := inst.state.Disabled

	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.rateStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, nil)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)
	inst.value = proputil.GetInt(props, propValue, inst.value)
	inst.count = proputil.GetInt(props, propCount, inst.count)
	inst.allowClear = proputil.GetBool(props, propAllowClear, inst.allowClear)
	inst.character = normalizeCharacter(proputil.GetString(props, propCharacter, inst.character))
	inst.showValue = proputil.GetBool(props, propShowValue, inst.showValue)

	newDisabled := proputil.GetBool(props, propDisabled, inst.state.Disabled)
	inst.state.Disabled = newDisabled

	inst.count = clampValue(inst.count, 1, 99)
	inst.value = clampValue(inst.value, 0, inst.count)

	changed := oldValue != inst.value ||
		oldCount != inst.count ||
		oldAllowClear != inst.allowClear ||
		oldCharacter != inst.character ||
		oldLabel != inst.label ||
		oldShowValue != inst.showValue ||
		oldDisabled != inst.state.Disabled

	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:        inst.key,
		propValue:      inst.value,
		propCount:      inst.count,
		propAllowClear: inst.allowClear,
		propCharacter:  inst.character,
		propDisabled:   inst.state.Disabled,
		propShowValue:  inst.showValue,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// --- PaintableInstance ---

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	stars := inst.renderStars()
	startX := x
	cmds := make([]paint.DrawCmd, 0, 3)

	if inst.label != "" {
		labelText := inst.label + ": "
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     y,
			Text:  labelText,
			Style: inst.resolveLabelStyle(),
		})
		startX += paint.StringWidth(labelText)
	}

	cmds = append(cmds, paint.DrawCmd{
		X:     startX,
		Y:     y,
		Text:  stars,
		Style: inst.resolveStarsStyle(),
	})

	if inst.showValue {
		cmds = append(cmds, paint.DrawCmd{
			X:     startX + paint.StringWidth(stars) + 1,
			Y:     y,
			Text:  fmt.Sprintf("%d/%d", inst.value, inst.count),
			Style: inst.resolveLabelStyle(),
		})
	}

	return cmds
}

func (inst *Instance) renderStars() string {
	filled := clampValue(inst.value, 0, inst.count)
	if inst.count <= 0 {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < inst.count; i++ {
		if i < filled {
			sb.WriteString(inst.character)
		} else {
			sb.WriteString(emptyCharacter)
		}
	}
	return sb.String()
}

func (inst *Instance) resolveStarsStyle() style.Style {
	s := inst.rateStyle

	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	if s.FG == "" {
		s = s.Foreground(theme.Warning())
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
	s := inst.rateStyle

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
	}

	return s
}

// --- FocusableInstance ---

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

func (inst *Instance) HasFocus() bool   { return inst.state.Focused }
func (inst *Instance) IsDisabled() bool { return inst.state.Disabled }

// --- ActionHandlerInstance ---

func (inst *Instance) HandleAction(act *action.Action) bool {
	if inst.state.Disabled {
		return false
	}

	switch act.Type {
	case action.ActionCursorLeft:
		inst.setValue(inst.value - 1)
		return true
	case action.ActionCursorRight:
		inst.setValue(inst.value + 1)
		return true
	case action.ActionCursorHome:
		inst.setValue(0)
		return true
	case action.ActionCursorEnd:
		inst.setValue(inst.count)
		return true
	case action.ActionSelectItem:
		if v, ok := act.GetPayloadInt(); ok {
			inst.setValue(v)
			return true
		}
	case action.ActionToggle, action.ActionEnter, action.ActionClick, action.ActionSelect:
		inst.advance()
		return true
	}

	return false
}

func (inst *Instance) advance() {
	if inst.value < inst.count {
		inst.setValue(inst.value + 1)
		return
	}
	if inst.allowClear {
		inst.setValue(0)
	}
}

func (inst *Instance) setValue(v int) {
	v = clampValue(v, 0, inst.count)
	if v != inst.value {
		inst.value = v
		inst.dirty = true
		inst.emitFieldValueChanged()
	}
}

// GetValue returns the current rate value.
func (inst *Instance) GetValue() int { return inst.value }

// SetValue sets the current rate value.
func (inst *Instance) SetValue(v int) { inst.setValue(v) }

// GetCount returns the total number of stars.
func (inst *Instance) GetCount() int { return inst.count }

// --- control.Instance ---

func (inst *Instance) GetState() *control.InteractionState { return &inst.state }

func (inst *Instance) SetState(state control.InteractionState) {
	oldState := inst.state
	inst.state = state
	inst.behaviors.OnStateChange(inst, oldState, inst.state)
}

func (inst *Instance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) GetStyle() style.Style   { return inst.rateStyle }
func (inst *Instance) SetStyle(st style.Style) { inst.rateStyle = st }

func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case propDisabled:
		return inst.state.Disabled, true
	case propValue:
		return inst.value, true
	case propCount:
		return inst.count, true
	case propAllowClear:
		return inst.allowClear, true
	case propCharacter:
		return inst.character, true
	case propLabel:
		return inst.label, true
	default:
		return nil, false
	}
}

func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case propDisabled:
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.dirty = true
		}
	case propValue:
		if v, ok := value.(int); ok {
			inst.setValue(v)
		}
	case propCount:
		if v, ok := value.(int); ok {
			inst.count = clampValue(v, 1, 99)
			inst.value = clampValue(inst.value, 0, inst.count)
			inst.dirty = true
		}
	}
}

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *Instance) ClearDirty() { inst.dirty = false }

// --- Measure ---

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	contentWidth := paint.StringWidth(inst.renderStars())
	if inst.label != "" {
		contentWidth += paint.StringWidth(inst.label + ": ")
	}
	if inst.showValue {
		contentWidth += 1 + len(fmt.Sprintf("%d/%d", inst.count, inst.count))
	}

	width := constraints.ConstrainWidth(contentWidth)
	height := constraints.ConstrainHeight(1)

	if inst.rateStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.rateStyle.Width)
	}
	if inst.rateStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.rateStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// --- intent emission ---

func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}

	value := fmt.Sprintf("%d", inst.value)

	if inst.formID != "" {
		if fieldIntent, ok := inst.changeIntent.(intent.FieldIntent); ok {
			intent.Emit(inst, form.FieldChange(
				inst.formID,
				fieldIntent.GetField(),
				value,
				true,
			))
		}
		return
	}

	if fieldIntent, ok := inst.changeIntent.(intent.FieldIntent); ok {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: fieldIntent.GetField(),
			Value: value,
		})
	} else if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) emitFieldBlur() {
	if inst.intentEmitter == nil || inst.formID == "" {
		return
	}

	if fieldIntent, ok := inst.changeIntent.(intent.FieldIntent); ok {
		intent.Emit(inst, form.FieldBlur(
			inst.formID,
			fieldIntent.GetField(),
			fmt.Sprintf("%d", inst.value),
		))
	}
}

// --- helpers ---

func normalizeCharacter(ch string) string {
	if strings.TrimSpace(ch) == "" {
		return defaultCharacter
	}
	return ch
}

func clampValue(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
