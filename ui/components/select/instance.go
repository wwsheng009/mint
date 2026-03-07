package selectcomp

import (
	"fmt"
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

// Instance is the runtime entity for Select components.
type Instance struct {
	// === Identification ===
	key         string
	componentID string // Component ID for Intent routing (Phase 10)

	// === Instance Tree (Phase 2 fix: Parent/Child tracking for Intent Bubble) ===
	parent rtui.ComponentInstance // Parent component for intent bubbling

	// === Props (from VNode) ===
	options      []Option
	selectStyle  style.Style
	width        int
	changeIntent intent.Intent
	changeIntentField intent.FieldIntent  // For FieldChangeIntent extraction
	formID       string // Form ID for Form integration (Phase 6)

	// === Runtime State ===
	state        control.InteractionState
	selectedIndex int
	bounds       [4]int
	dirty        bool

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

// NewInstance creates a new SelectInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:                getStringProp(props, "key", ""),
		componentID:        getStringProp(props, "componentID", ""),
		options:            getOptionsProp(props),
		selectStyle:        getStyleProp(props),
		width:              getIntProp(props, "width", 0),
		changeIntent:       getIntentProp(props, "changeIntent"),
		changeIntentField:  getChangeIntentFieldProp(props, "changeIntent"),
		formID:             getStringProp(props, "formID", ""),
		selectedIndex:      getIntProp(props, "selectedIndex", -1),
		dirty:              true,
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

// Parent implements TreeComponent interface (intent bubble).
// Returns the parent component for intent bubbling.
// Phase 2 fix: Now properly returns parent reference (was always nil before).
func (inst *Instance) Parent() interface{} {
	return inst.parent
}

// SetParent sets the parent component for intent bubbling.
// Phase 2 fix: Added method for components to set parent reference.
func (inst *Instance) SetParent(parent rtui.ComponentInstance) {
	inst.parent = parent
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
	oldOptions := inst.options
	oldSelected := inst.selectedIndex
	oldDisabled := inst.state.Disabled

	inst.componentID = getStringProp(props, "componentID", inst.componentID)
	inst.options = getOptionsProp(props)
	inst.selectStyle = getStyleProp(props)
	inst.width = getIntProp(props, "width", inst.width)
	inst.changeIntent = getIntentProp(props, "changeIntent")
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
	inst.formID = getStringProp(props, "formID", inst.formID)
	inst.selectedIndex = getIntProp(props, "selectedIndex", inst.selectedIndex)

	newDisabled := getBoolProp(props, "disabled", inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	changed := len(oldOptions) != len(inst.options) ||
		oldSelected != inst.selectedIndex ||
		oldDisabled != inst.state.Disabled

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":          inst.key,
		"selectedIndex": inst.selectedIndex,
		"disabled":     inst.state.Disabled,
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

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// Get the selected label
	displayLabel := inst.SelectedLabel()
	if displayLabel == "" && len(inst.options) > 0 {
		displayLabel = inst.options[0].Label
	}
	if displayLabel == "" {
		displayLabel = "..."
	}

	// Build select display: < label >
	selectDisplay := "< " + displayLabel + " >"

	// Resolve style
	selectStyle := inst.resolveStyle()

	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  selectDisplay,
		Style: selectStyle,
	}}
}

// resolveStyle resolves the visual style based on state.
func (inst *Instance) resolveStyle() style.Style {
	s := inst.selectStyle

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
	if inst.state.Disabled || len(inst.options) == 0 {
		return false
	}

	switch act.Type {
	case action.ActionSelect, action.ActionClick, action.ActionEnter,
		action.ActionNavigateDown:
		inst.SelectNext()
		return true
	case action.ActionNavigateUp:
		inst.SelectPrev()
		return true
	}
	return false
}

// =============================================================================
// Select-specific Methods
// =============================================================================

// SelectNext selects the next option.
func (inst *Instance) SelectNext() {
	if len(inst.options) == 0 {
		return
	}
	inst.selectedIndex = (inst.selectedIndex + 1) % len(inst.options)
	inst.dirty = true
	inst.emitChange()
	inst.emitSelectChange() // Phase 10: Emit SelectChangeIntent
}

// SelectPrev selects the previous option.
func (inst *Instance) SelectPrev() {
	if len(inst.options) == 0 {
		return
	}
	inst.selectedIndex--
	if inst.selectedIndex < 0 {
		inst.selectedIndex = len(inst.options) - 1
	}
	inst.dirty = true
	inst.emitChange()
	inst.emitSelectChange() // Phase 10: Emit SelectChangeIntent
}

// SetSelectedIndex sets the selected index.
func (inst *Instance) SetSelectedIndex(idx int) {
	if idx >= -1 && idx < len(inst.options) && inst.selectedIndex != idx {
		inst.selectedIndex = idx
		inst.dirty = true
		inst.emitSelectChange() // Phase 10: Emit SelectChangeIntent
	}
}

// SelectedIndex returns the selected index.
func (inst *Instance) SelectedIndex() int {
	return inst.selectedIndex
}

// SelectedValue returns the selected value.
func (inst *Instance) SelectedValue() string {
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(inst.options) {
		return inst.options[inst.selectedIndex].Value
	}
	return ""
}

// SelectedLabel returns the selected label.
func (inst *Instance) SelectedLabel() string {
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(inst.options) {
		return inst.options[inst.selectedIndex].Label
	}
	return ""
}

// emitChange emits the change intent.
func (inst *Instance) emitChange() {
	// ✨ MVP/Phase 6: Emit FieldChangeIntent or FormFieldChangeIntent with runtime value
	// State becomes the single source of truth
	inst.emitFieldValueChanged()
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

// HandleIntent implements intent.IntentHandler to handle select-specific intents.
// This allows external components or controllers to control the select via intents.
// Phase 2 fix: Uses unified component ID routing (P1-4: INTENT_BUBBLE_AUDIT_REPORT.md)
func (inst *Instance) HandleIntent(i intent.Intent) bool {
	// Check if this intent is for this component using unified routing
	if !intent.ShouldHandleIntentWithID(inst.componentID, i) {
		// Intent is for a different component, ignore it
		return false
	}

	// Process the intent
	switch v := i.(type) {
	case SelectNextIntent:
		// Handle next selection by external request
		if len(inst.options) == 0 {
			return false
		}
		inst.SelectNext()
		return true

	case SelectPrevIntent:
		// Handle previous selection by external request
		if len(inst.options) == 0 {
			return false
		}
		inst.SelectPrev()
		return true

	case SelectByIndexIntent:
		// Handle selection by index
		if v.Index >= -1 && v.Index < len(inst.options) {
			inst.SetSelectedIndex(v.Index)
			return true
		}

	case SelectByValueIntent:
		// Handle selection by value
		for idx, opt := range inst.options {
			if opt.Value == v.Value {
				inst.SetSelectedIndex(idx)
				return true
			}
		}
	}

	return false
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
	return inst.selectStyle
}

// SetStyle sets the visual style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.selectStyle = s
}

// GetProp returns a prop value.
func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case "disabled":
		return inst.state.Disabled, true
	case "selectedIndex":
		return inst.selectedIndex, true
	case "options":
		return inst.options, true
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
	case "selectedIndex":
		if v, ok := value.(int); ok {
			inst.SetSelectedIndex(v)
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
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Find the longest option label
	maxWidth := 0
	for _, opt := range inst.options {
		labelWidth := utf8.RuneCountInString(opt.Label)
		if labelWidth > maxWidth {
			maxWidth = labelWidth
		}
	}

	// Minimum width
	if maxWidth < 10 {
		maxWidth = 10
	}

	// Width: longest label + 4 for "< " and " >"
	width := maxWidth + 4
	height := 1

	// Apply explicit width
	if inst.width > 0 {
		width = inst.width
	}

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// Form Integration Methods (Phase 6)
// =============================================================================

// emitSelectChange emits SelectChangeIntent (Phase 10).
// Phase 11 fix: Uses intentEmitter to trigger FiberUtil's smart routing.
func (inst *Instance) emitSelectChange() {
	if inst.intentEmitter == nil {
		return
	}

	var selectIntent SelectChangeIntent
	selectedValue := inst.SelectedValue()
	selectedLabel := inst.SelectedLabel()

	if inst.componentID != "" {
		selectIntent = SelectChangeWithID(inst.componentID, inst.selectedIndex, selectedValue, selectedLabel)
	} else {
		selectIntent = SelectChange(inst.selectedIndex, selectedValue, selectedLabel)
	}

	// Emit via intentEmitter (FiberUtil's smart router decides global vs local)
	inst.intentEmitter(selectIntent)
}

// emitFieldValueChanged emits FieldChangeIntent or FormFieldChangeIntent
// depending on whether formID is set.
func (inst *Instance) emitFieldValueChanged() {
	if inst.intentEmitter == nil {
		return
	}

	// Convert selected index to string for FieldChangeIntent
	value := fmt.Sprintf("%d", inst.selectedIndex)

	// Phase 6: If formID is set, use FormFieldChangeIntent
	if inst.formID != "" {
		if inst.changeIntentField != nil {
			formIntent := form.FieldChange(
				inst.formID,
				inst.changeIntentField.GetField(),
				value,
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
			Value: value,
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

	// Convert selected index to string for FieldBlurIntent
	value := fmt.Sprintf("%d", inst.selectedIndex)

	if inst.changeIntentField != nil {
		blurIntent := form.FieldBlur(
			inst.formID,
			inst.changeIntentField.GetField(),
			value,
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

func getOptionsProp(props rtui.Props) []Option {
	if v, ok := props["options"]; ok {
		if opts, ok := v.([]Option); ok {
			return opts
		}
	}
	return nil
}
