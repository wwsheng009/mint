package textarea

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
	"github.com/wwsheng009/mint/ui/components/form"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Textarea components.
type Instance struct {
	key          string
	placeholder  string
	textareaStyle style.Style
	rows, cols   int
	changeIntent intent.Intent
	changeIntentField intent.FieldIntent  // For FieldChangeIntent extraction
	submitIntent intent.Intent
	formID       string // Form ID for Form integration (Phase 6)
	maxLen       int

	state   control.InteractionState
	value   string
	bounds  [4]int
	dirty   bool

	intentEmitter func(intent.Intent)
	behaviors     *control.BehaviorList
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ rtui.FocusableInstance = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ control.Instance = (*Instance)(nil)
	_ interface{ Measure(layout.Constraints) layout.Size } = (*Instance)(nil)
)

// NewInstance creates a new TextareaInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:               getStringProp(props, "key", ""),
		placeholder:       getStringProp(props, "placeholder", ""),
		textareaStyle:     getStyleProp(props),
		rows:              getIntProp(props, "rows", 3),
		cols:              getIntProp(props, "cols", 40),
		changeIntent:      getIntentProp(props, "changeIntent"),
		changeIntentField: getChangeIntentFieldProp(props, "changeIntent"),
		formID:            getStringProp(props, "formID", ""),
		submitIntent:      getIntentProp(props, "submitIntent"),
		value:             getStringProp(props, "value", ""),
		maxLen:            getIntProp(props, "maxLen", 0),
		dirty:             true,
	}

	inst.state = control.InteractionState{
		Disabled: getBoolProp(props, "disabled", false),
	}

	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)

	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string         { return inst.key }
func (inst *Instance) SetKey(key string)   { inst.key = key }

// Parent implements TreeComponent interface (intent bubble).
// Returns nil as Textarea is a leaf component without parent tracking.
func (inst *Instance) Parent() interface{} {
	return nil
}

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()            { inst.behaviors.OnUnmount(inst) }
func (inst *Instance) OnMount()            { inst.behaviors.OnMount(inst) }
func (inst *Instance) OnUnmount()          { inst.behaviors.OnUnmount(inst) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldValue := inst.value
	oldDisabled := inst.state.Disabled

	inst.placeholder = getStringProp(props, "placeholder", inst.placeholder)
	inst.textareaStyle = getStyleProp(props)
	inst.rows = getIntProp(props, "rows", inst.rows)
	inst.cols = getIntProp(props, "cols", inst.cols)
	inst.changeIntent = getIntentProp(props, "changeIntent")
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
	inst.formID = getStringProp(props, "formID", inst.formID)
	inst.submitIntent = getIntentProp(props, "submitIntent")
	inst.value = getStringProp(props, "value", inst.value)
	inst.maxLen = getIntProp(props, "maxLen", inst.maxLen)

	newDisabled := getBoolProp(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	changed := oldValue != inst.value || oldDisabled != inst.state.Disabled
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":      inst.key,
		"value":    inst.value,
		"disabled": inst.state.Disabled,
	}
}

func (inst *Instance) MarkDirty()           { inst.dirty = true }
func (inst *Instance) IsDirty() bool        { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

// =============================================================================
// PaintableInstance Interface
// =============================================================================

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	content := inst.value
	if content == "" {
		content = inst.placeholder
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	// Calculate dimensions
	width := inst.cols + 4
	height := inst.rows + 2
	if len(lines) > inst.rows {
		height = len(lines) + 2
	}

	// Resolve style
	borderStyle := inst.resolveStyle()

	// Build commands
	var cmds []paint.DrawCmd

	// Top border
	topBorder := "+" + strings.Repeat("-", width-2) + "+"
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y, Text: topBorder, Style: borderStyle})

	// Content lines
	for i, line := range lines {
		if i >= height-2 {
			break
		}
		lineWidth := utf8.RuneCountInString(line)
		padding := width - 2 - lineWidth
		if padding < 0 {
			// Truncate long lines
			runes := []rune(line)
			line = string(runes[:width-2])
			padding = 0
		}
		lineText := "|" + line + strings.Repeat(" ", padding) + "|"
		cmds = append(cmds, paint.DrawCmd{X: x, Y: y + 1 + i, Text: lineText, Style: borderStyle})
	}

	// Fill empty rows
	for i := len(lines); i < inst.rows; i++ {
		emptyLine := "|" + strings.Repeat(" ", width-2) + "|"
		cmds = append(cmds, paint.DrawCmd{X: x, Y: y + 1 + i, Text: emptyLine, Style: borderStyle})
	}

	// Bottom border
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y + height - 1, Text: topBorder, Style: borderStyle})

	return cmds
}

func (inst *Instance) resolveStyle() style.Style {
	s := inst.textareaStyle

	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	if inst.value == "" && inst.placeholder != "" {
		s = s.Foreground(theme.Placeholder())
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

// =============================================================================
// FocusableInstance Interface
// =============================================================================

func (inst *Instance) SetFocus(focused bool) {
	if inst.state.Focused != focused {
		oldState := inst.state
		inst.state.Focused = focused
		inst.dirty = true
		inst.behaviors.OnStateChange(inst, oldState, inst.state)
	}
}

func (inst *Instance) HasFocus() bool   { return inst.state.Focused }
func (inst *Instance) IsDisabled() bool { return inst.state.Disabled }

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) HandleAction(act *action.Action) bool {
	if inst.state.Disabled {
		return false
	}

	switch act.Type {
	case action.ActionInputText:
		if text, ok := act.GetPayloadString(); ok {
			return inst.InsertText(text)
		}
		return false
	case action.ActionSubmit:
		if inst.submitIntent != nil && inst.intentEmitter != nil {
			inst.intentEmitter(inst.submitIntent)
		}
		return true
	}
	return false
}

// =============================================================================
// Textarea-specific Methods
// =============================================================================

func (inst *Instance) InsertText(text string) bool {
	if inst.state.Disabled {
		return false
	}

	if inst.maxLen > 0 {
		currentLen := utf8.RuneCountInString(inst.value)
		textLen := utf8.RuneCountInString(text)
		if currentLen+textLen > inst.maxLen {
			return false
		}
	}

	inst.value += text
	inst.dirty = true

	// ✨ MVP/Phase 6: Emit FieldChangeIntent or FormFieldChangeIntent with runtime value
	// State becomes the single source of truth
	inst.emitFieldValueChanged()

	return true
}

func (inst *Instance) SetValue(value string) {
	if inst.value != value {
		inst.value = value
		inst.dirty = true
	}
}

func (inst *Instance) GetValue() string { return inst.value }

// =============================================================================
// control.Instance Interface
// =============================================================================

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
func (inst *Instance) GetBounds() (x, y, w, h int) { return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3] }
func (inst *Instance) SetBounds(x, y, w, h int)    { inst.bounds = [4]int{x, y, w, h} }
func (inst *Instance) GetStyle() style.Style       { return inst.textareaStyle }
func (inst *Instance) SetStyle(s style.Style)      { inst.textareaStyle = s }
func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case "disabled":
		return inst.state.Disabled, true
	case "value":
		return inst.value, true
	default:
		return nil, false
	}
}
func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case "disabled":
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.dirty = true
		}
	case "value":
		if v, ok := value.(string); ok {
			inst.SetValue(v)
		}
	}
}
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) { inst.intentEmitter = fn }
func (inst *Instance) ClearDirty()                             { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	width := inst.cols + 4
	height := inst.rows + 2

	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// Form Integration Methods (Phase 6)
// =============================================================================

// emitFieldValueChanged emits FieldChangeIntent or FormFieldChangeIntent
// depending on whether formID is set.
func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}

	// Phase 6: If formID is set, use FormFieldChangeIntent
	if inst.formID != "" {
		if inst.changeIntentField != nil {
			formIntent := form.FieldChange(
				inst.formID,
				inst.changeIntentField.GetField(),
				inst.value,
				true, // isDirty
			)
			intent.Emit(inst, formIntent)
		}
		return
	}

	// Original MVP behavior: emit FieldChangeIntent
	if inst.changeIntentField != nil {
		changeIntent := intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: inst.value,
		}
		inst.intentEmitter(changeIntent)
	} else if inst.changeIntent != nil {
		// Fallback: emit the original intent for backward compatibility
		inst.intentEmitter(inst.changeIntent)
	}
}

// emitFieldBlur emits FormFieldBlurIntent to trigger validation (Phase 6)
// Only called when formID is set.
func (inst *Instance) emitFieldBlur() {
	if inst.intentEmitter == nil || inst.formID == "" {
		return
	}

	if inst.changeIntentField != nil {
		blurIntent := form.FieldBlur(
			inst.formID,
			inst.changeIntentField.GetField(),
			inst.value,
		)
		intent.Emit(inst, blurIntent)
	}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["style"]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.Style{}
}

func getIntentProp(props rtui.Props, key string) intent.Intent {
	if v, ok := props[key]; ok {
		if i, ok := v.(intent.Intent); ok {
			return i
		}
	}
	return nil
}

func getChangeIntentFieldProp(props rtui.Props, key string) intent.FieldIntent {
	if v, ok := props[key]; ok {
		if fieldIntent, ok := v.(intent.FieldIntent); ok {
			return fieldIntent
		}
	}
	return nil
}
