package checkbox

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Checkbox components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label        string
	checkboxStyle style.Style
	toggleIntent  intent.Intent

	// === Runtime State (managed by instance) ===
	state   control.InteractionState
	checked bool
	bounds  [4]int // x, y, w, h
	dirty   bool

	// === Intent Emitter ===
	intentEmitter func(intent.Intent)

	// === Behaviors ===
	behaviors *control.BehaviorList
}

// Ensure Instance implements required interfaces
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

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new CheckboxInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:           getStringProp(props, "key", ""),
		label:         getStringProp(props, "label", ""),
		checkboxStyle: getStyleProp(props),
		toggleIntent:  getIntentProp(props),
		checked:       getBoolProp(props, "checked", false),
		dirty:         true,
	}

	// Initialize state
	inst.state = control.InteractionState{
		Disabled: getBoolProp(props, "disabled", false),
	}

	// Initialize behaviors
	inst.initBehaviors()

	return inst
}

// initBehaviors initializes the behavior composition.
func (inst *Instance) initBehaviors() {
	// Compose behaviors - checkbox uses standard behaviors
	// Toggle logic is handled directly in HandleAction
	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		&control.HoverableBehavior{},
		&control.DisableableBehavior{},
	)
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// Key implements ComponentInstance.
func (inst *Instance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *Instance) SetKey(key string) {
	inst.key = key
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
	oldIntent := inst.toggleIntent

	inst.label = getStringProp(props, "label", inst.label)
	inst.checkboxStyle = getStyleProp(props)
	inst.toggleIntent = getIntentProp(props)
	inst.checked = getBoolProp(props, "checked", inst.checked)

	newDisabled := getBoolProp(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	// Check if props changed
	changed := oldLabel != inst.label ||
		oldDisabled != inst.state.Disabled ||
		oldChecked != inst.checked ||
		oldIntent != inst.toggleIntent

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":      inst.key,
		"label":    inst.label,
		"disabled": inst.state.Disabled,
		"checked":  inst.checked,
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

// GetContext implements ComponentInstance (no hooks for Checkbox).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// Build checkbox indicator: [X] or [ ]
	var indicator string
	if inst.checked {
		indicator = "[X]"
	} else {
		indicator = "[ ]"
	}

	// Build checkbox display: indicator + gap + label
	var displayText string
	if inst.label != "" {
		displayText = indicator + " " + inst.label
	} else {
		displayText = indicator
	}

	// Apply styling
	checkboxStyle := inst.resolveStyle()

	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  displayText,
		Style: checkboxStyle,
	}}
}

// resolveStyle resolves the visual style based on state.
func (inst *Instance) resolveStyle() style.Style {
	s := inst.checkboxStyle

	// Apply default colors if not set
	if s.FG == "" {
		s = s.Foreground(theme.Text())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}

	// Checked state: make it bold
	if inst.checked && !inst.state.Disabled {
		s = s.Bold(true)
	}

	// State priority: Disabled > Focused > Hovered > Normal
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

// SetFocus implements FocusableInstance.
func (inst *Instance) SetFocus(focused bool) {
	if inst.state.Focused != focused {
		oldState := inst.state
		inst.state.Focused = focused
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

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

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

// =============================================================================
// Checkbox-specific Methods
// =============================================================================

// Toggle toggles the checked state and returns the new value.
func (inst *Instance) Toggle() bool {
	inst.checked = !inst.checked
	inst.dirty = true

	// ✨ MVP: Emit FieldChangeIntent with runtime value
	// State becomes the single source of truth
	// Convert boolean to string for FieldChangeIntent
	value := "false"
	if inst.checked {
		value = "true"
	}

	if inst.intentEmitter != nil {
		if fieldIntent, ok := inst.toggleIntent.(intent.FieldIntent); ok {
			changeIntent := intent.FieldChangeIntent{
				Field: fieldIntent.GetField(),
				Value: value,
			}
			inst.intentEmitter(changeIntent)
		} else if inst.toggleIntent != nil {
			// Fallback: emit the original intent for backward compatibility
			inst.intentEmitter(inst.toggleIntent)
		}
	}

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

// Label returns the label text.
func (inst *Instance) Label() string {
	return inst.label
}

// =============================================================================
// control.Instance Interface (for Behaviors)
// =============================================================================

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
	return inst.checkboxStyle
}

// SetStyle sets the visual style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.checkboxStyle = s
}

// GetProp returns a prop value.
func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case "disabled":
		return inst.state.Disabled, true
	case "checked":
		return inst.checked, true
	case "label":
		return inst.label, true
	case "toggleIntent":
		return inst.toggleIntent, true
	default:
		return nil, false
	}
}

// SetProp sets a prop value.
func (inst *Instance) SetProp(key string, value interface{}) {
	switch key {
	case "disabled":
		if v, ok := value.(bool); ok {
			inst.state.Disabled = v
			inst.dirty = true
		}
	case "checked":
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

// =============================================================================
// Measurable Interface (Two-Pass Layout)
// =============================================================================

// Measure implements layout.Measurable interface.
// Calculates the checkbox's ideal size given the constraints.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Width: "[X]" (3) + " " (1) + label
	// Per Ant Design spec: checkbox-width=3, gap=1
	contentWidth := 4 + utf8.RuneCountInString(inst.label)
	contentHeight := 1

	// Apply constraints
	width := constraints.ConstrainWidth(contentWidth)
	height := constraints.ConstrainHeight(contentHeight)

	// Apply explicit style dimensions if set
	if inst.checkboxStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.checkboxStyle.Width)
	}
	if inst.checkboxStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.checkboxStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// GetNaturalSize returns the natural (unconstrained) size.
func (inst *Instance) GetNaturalSize() (width, height int) {
	width = 4 + utf8.RuneCountInString(inst.label)
	height = 1
	return width, height
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

func getIntentProp(props rtui.Props) intent.Intent {
	if v, ok := props["toggleIntent"]; ok {
		if i, ok := v.(intent.Intent); ok {
			return i
		}
	}
	return nil
}
