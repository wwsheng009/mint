package slider

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

const (
	filledChar   = "█"
	unfilledChar = "░"
)

// Instance is the runtime entity for Slider components.
type Instance struct {
	key          string
	label        string
	sliderStyle  style.Style
	changeIntent intent.Intent
	formID       string

	value int
	min   int
	max   int
	step  int
	width int

	showValue bool
	state     control.InteractionState
	bounds    [4]int
	dirty     bool

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

// NewInstance creates a new Slider instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          proputil.GetString(props, propKey, ""),
		label:        proputil.GetString(props, propLabel, ""),
		sliderStyle:  proputil.GetStyle(props, propStyle, style.Style{}),
		changeIntent: proputil.GetIntent(props, propChangeIntent, nil),
		formID:       proputil.GetString(props, propFormID, ""),
		value:        proputil.GetInt(props, propValue, 0),
		min:          proputil.GetInt(props, propMin, 0),
		max:          proputil.GetInt(props, propMax, defaultMax),
		step:         proputil.GetInt(props, propStep, defaultStep),
		width:        proputil.GetInt(props, propWidth, defaultWidth),
		showValue:    proputil.GetBool(props, propShowValue, true),
		dirty:        true,
	}

	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, propDisabled, false),
	}

	inst.value = clampValue(inst.value, inst.min, inst.max)
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

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Parent() interface{} { return nil }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.behaviors.OnUnmount(inst) }
func (inst *Instance) OnMount()              { inst.behaviors.OnMount(inst) }
func (inst *Instance) OnUnmount()            { inst.behaviors.OnUnmount(inst) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldValue := inst.value
	oldMin := inst.min
	oldMax := inst.max
	oldStep := inst.step
	oldWidth := inst.width
	oldLabel := inst.label
	oldShowValue := inst.showValue
	oldDisabled := inst.state.Disabled

	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.sliderStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, nil)
	inst.formID = proputil.GetString(props, propFormID, inst.formID)
	inst.value = proputil.GetInt(props, propValue, inst.value)
	inst.min = proputil.GetInt(props, propMin, inst.min)
	inst.max = proputil.GetInt(props, propMax, inst.max)
	inst.step = proputil.GetInt(props, propStep, inst.step)
	inst.width = proputil.GetInt(props, propWidth, inst.width)
	inst.showValue = proputil.GetBool(props, propShowValue, inst.showValue)

	newDisabled := proputil.GetBool(props, propDisabled, inst.state.Disabled)
	inst.state.Disabled = newDisabled

	inst.value = clampValue(inst.value, inst.min, inst.max)

	changed := oldValue != inst.value ||
		oldMin != inst.min ||
		oldMax != inst.max ||
		oldStep != inst.step ||
		oldWidth != inst.width ||
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
		propKey:       inst.key,
		propValue:     inst.value,
		propMin:       inst.min,
		propMax:       inst.max,
		propDisabled:  inst.state.Disabled,
		propShowValue: inst.showValue,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext  { return nil }

// --- PaintableInstance ---

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	track := inst.renderTrack()
	trackStyle := inst.resolveTrackStyle()

	cmds := []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  track,
		Style: trackStyle,
	}}

	offset := paint.StringWidth(track)

	if inst.label != "" {
		cmds = append([]paint.DrawCmd{{
			X:     x,
			Y:     y,
			Text:  inst.label + ": ",
			Style: inst.resolveLabelStyle(),
		}}, cmds...)
		// shift track right after label
		cmds[1].X = x + paint.StringWidth(inst.label+": ")
		offset = cmds[1].X + paint.StringWidth(track)
	}

	if inst.showValue {
		cmds = append(cmds, paint.DrawCmd{
			X:     offset + 1,
			Y:     y,
			Text:  fmt.Sprintf("%d", inst.value),
			Style: inst.resolveLabelStyle(),
		})
	}

	return cmds
}

func (inst *Instance) renderTrack() string {
	w := inst.width
	if w < 2 {
		w = 2
	}

	filled := 0
	span := inst.max - inst.min
	if span > 0 {
		filled = ((inst.value - inst.min) * w) / span
	}
	if filled > w {
		filled = w
	}

	var sb strings.Builder
	for i := 0; i < w; i++ {
		if i < filled {
			sb.WriteString(filledChar)
		} else {
			sb.WriteString(unfilledChar)
		}
	}
	return sb.String()
}

func (inst *Instance) resolveTrackStyle() style.Style {
	s := inst.sliderStyle

	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	if s.FG == "" {
		s = s.Foreground(theme.Primary())
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
	s := inst.sliderStyle

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
		inst.changeBy(-inst.step)
		return true
	case action.ActionCursorRight:
		inst.changeBy(inst.step)
		return true
	case action.ActionCursorHome:
		inst.setValue(inst.min)
		return true
	case action.ActionCursorEnd:
		inst.setValue(inst.max)
		return true
	}
	return false
}

// changeBy adjusts value by delta (snapped to step, clamped to [min,max]).
func (inst *Instance) changeBy(delta int) {
	inst.setValue(inst.value + delta)
}

// setValue sets the value, clamping and emitting intent.
func (inst *Instance) setValue(v int) {
	v = clampValue(v, inst.min, inst.max)
	if v != inst.value {
		inst.value = v
		inst.dirty = true
		inst.emitFieldValueChanged()
	}
}

// GetValue returns the current slider value.
func (inst *Instance) GetValue() int { return inst.value }

// SetValue sets the slider value (public, for direct use).
func (inst *Instance) SetValue(v int) {
	inst.setValue(v)
}

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

func (inst *Instance) GetStyle() style.Style { return inst.sliderStyle }
func (inst *Instance) SetStyle(st style.Style) { inst.sliderStyle = st }

func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case propDisabled:
		return inst.state.Disabled, true
	case propValue:
		return inst.value, true
	case propMin:
		return inst.min, true
	case propMax:
		return inst.max, true
	case propStep:
		return inst.step, true
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

	w := inst.width
	if w < 2 {
		w = 2
	}

	contentWidth := w
	if inst.label != "" {
		contentWidth += paint.StringWidth(inst.label+": ")
	}
	if inst.showValue {
		contentWidth += 1 + len(fmt.Sprintf("%d", inst.max))
	}

	width := constraints.ConstrainWidth(contentWidth)
	height := constraints.ConstrainHeight(1)

	if inst.sliderStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.sliderStyle.Width)
	}
	if inst.sliderStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.sliderStyle.Height)
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

	value := fmt.Sprintf("%d", inst.value)

	if fieldIntent, ok := inst.changeIntent.(intent.FieldIntent); ok {
		intent.Emit(inst, form.FieldBlur(
			inst.formID,
			fieldIntent.GetField(),
			value,
		))
	}
}

// --- helpers ---

func clampValue(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
