package button

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
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

// Instance is the runtime entity for Button components.
// It persists across renders and holds all state.
//
// Architecture (from fiber_paint.md):
//
//	ButtonVNode (description) ──CreateInstance()──→ Instance
//	                                                 ↓
//	                                           Fiber.Instance = inst
//	                                                 ↓
//	                                           Persists across renders
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label       string
	variant     Variant
	size        Size
	focusStyle  FocusStyle
	buttonStyle style.Style
	pressIntent intent.Intent
	padding     [4]int
	textAlign   rtui.Align
	flex        int // flex grow factor

	// === Runtime State (managed by instance) ===
	state  control.InteractionState
	bounds [4]int // x, y, w, h
	dirty  bool

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

// NewInstance creates a new ButtonInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:         proputil.GetString(props, propKey, ""),
		label:       proputil.GetString(props, propLabel, ""),
		variant:     getVariantProp(props, VariantDefault),
		size:        getSizeProp(props, SizeMedium),
		focusStyle:  getFocusStyleProp(props, FocusStyleReverse),
		buttonStyle: proputil.GetStyle(props, propStyle, style.Style{}),
		pressIntent: proputil.GetIntent(props, propPressIntent, nil),
		padding:     getPaddingProp(props),
		textAlign:   getTextAlignProp(props, rtui.AlignStart),
		flex:        proputil.GetInt(props, propFlex, 0),
		dirty:       true,
	}

	// Initialize state
	inst.state = control.InteractionState{
		Disabled: proputil.GetBool(props, propDisabled, false),
	}

	// Initialize behaviors
	inst.initBehaviors()

	return inst
}

// initBehaviors initializes the behavior composition.
func (inst *Instance) initBehaviors() {
	// Create PressableBehavior with the press intent
	pressable := control.NewPressableBehavior(inst.pressIntent)

	// Compose behaviors
	inst.behaviors = control.NewBehaviorList(
		&control.FocusableBehavior{},
		pressable,
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
	// Notify behaviors
	inst.behaviors.OnUnmount(inst)
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {
	// Notify behaviors
	inst.behaviors.OnMount(inst)
}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {
	inst.behaviors.OnUnmount(inst)
}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldLabel := inst.label
	oldVariant := inst.variant
	oldSize := inst.size
	oldDisabled := inst.state.Disabled
	oldFocusStyle := inst.focusStyle
	oldIntent := inst.pressIntent

	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.variant = getVariantProp(props, inst.variant)
	inst.size = getSizeProp(props, inst.size)
	inst.focusStyle = getFocusStyleProp(props, inst.focusStyle)
	inst.buttonStyle = proputil.GetStyle(props, propStyle, style.Style{})
	inst.pressIntent = proputil.GetIntent(props, propPressIntent, nil)
	inst.padding = getPaddingProp(props)
	inst.textAlign = getTextAlignProp(props, inst.textAlign)

	newFlex := proputil.GetInt(props, propFlex, inst.flex)
	if newFlex != inst.flex {
		inst.flex = newFlex
	}

	newDisabled := proputil.GetBool(props, propDisabled, inst.state.Disabled)
	if newDisabled != inst.state.Disabled {
		inst.state.Disabled = newDisabled
	}

	// Update pressable behavior intent
	if inst.pressIntent != oldIntent {
		if pressable := inst.behaviors.Get("Pressable"); pressable != nil {
			if p, ok := pressable.(*control.PressableBehavior); ok {
				p.SetIntent(inst.pressIntent)
			}
		}
	}

	// Check if props changed
	changed := oldLabel != inst.label ||
		oldVariant != inst.variant ||
		oldSize != inst.size ||
		oldDisabled != inst.state.Disabled ||
		oldFocusStyle != inst.focusStyle ||
		oldIntent != inst.pressIntent

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":        inst.key,
		propLabel:      inst.label,
		"variant":    inst.variant,
		"size":       inst.size,
		"focusStyle": inst.focusStyle,
		propDisabled:   inst.state.Disabled,
		propFlex:       inst.flex,
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

// GetContext implements ComponentInstance (no hooks for Button).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// Build button label with brackets
	displayLabel := inst.label
	if displayLabel == "" {
		displayLabel = " "
	}

	// Format: [label]
	var labelText string
	switch inst.size {
	case SizeSmall:
		labelText = "[" + displayLabel + "]"
	case SizeMedium:
		labelText = "[ " + displayLabel + " ]"
	case SizeLarge:
		labelText = "[  " + displayLabel + "  ]"
	default:
		labelText = "[" + displayLabel + "]"
	}

	// Apply variant-based styling
	buttonStyle := inst.resolveStyle()

	// Build button text
	buttonText := inst.buildButtonText(labelText, buttonStyle)

	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  buttonText,
		Style: buttonStyle,
	}}
}

// resolveStyle resolves the visual style based on state.
func (inst *Instance) resolveStyle() style.Style {
	s := inst.buttonStyle

	// Apply variant-based styling if not explicitly set
	if s.FG == "" && s.BG == "" {
		switch inst.variant {
		case VariantPrimary:
			s = s.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
		case VariantSecondary:
			s = s.Foreground(theme.Text()).Background(theme.Surface())
		case VariantDanger:
			s = s.Foreground(theme.BG()).Background(theme.Error()).Bold(true)
		case VariantSuccess:
			s = s.Foreground(theme.BG()).Background(theme.Success()).Bold(true)
		case VariantDefault:
			s = s.Foreground(theme.Text()).Background(theme.Surface())
		}
	}

	// Apply state-based styling
	// Priority: Disabled > Focused > Hovered > Normal
	if inst.state.Disabled {
		s = s.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	} else if inst.state.Focused {
		switch inst.focusStyle {
		case FocusStyleUnderline:
			s = s.Foreground(theme.FocusBright()).Underline(true).Bold(true)
		case FocusStyleBracket:
			customBG := s.BG
			s = s.Foreground(theme.FocusBright()).Bold(true)
			if customBG != "" {
				s = s.Background(customBG)
			}
		case FocusStyleBold:
			s = s.Bold(true)
		case FocusStyleReverse:
			s = s.Foreground(theme.Foreground()).Background(theme.Focus()).Bold(true)
		}
	} else if inst.state.Hovered {
		s = s.Underline(true)
	}

	return s
}

// buildButtonText builds the button text with focus indicator and padding.
func (inst *Instance) buildButtonText(labelText string, buttonStyle style.Style) string {
	// Add focus indicator
	var focusIndicator string
	if inst.state.Focused && !inst.state.Disabled {
		if inst.focusStyle == FocusStyleReverse {
			focusIndicator = "*"
		} else if inst.focusStyle == FocusStyleUnderline || inst.focusStyle == FocusStyleBracket {
			focusIndicator = ">"
		}
	}
	if focusIndicator == "" {
		focusIndicator = " "
	}

	buttonText := focusIndicator + labelText
	contentWidth := paint.StringWidth(buttonText)

	// Get padding
	paddingLeft := inst.padding[3]
	paddingRight := inst.padding[1]

	naturalWidth := contentWidth
	layoutWidth := naturalWidth
	if inst.bounds[2] > 0 {
		layoutWidth = inst.bounds[2]
	}

	// Apply text alignment if button is stretched
	if layoutWidth > naturalWidth {
		availableSpace := layoutWidth - naturalWidth
		switch inst.textAlign {
		case rtui.AlignCenter:
			leftSpace := paddingLeft + availableSpace/2
			rightSpace := paddingRight + (availableSpace - availableSpace/2)
			buttonText = strings.Repeat(" ", leftSpace) + buttonText + strings.Repeat(" ", rightSpace)
		case rtui.AlignEnd:
			leftSpace := paddingLeft + availableSpace
			buttonText = strings.Repeat(" ", leftSpace) + buttonText + strings.Repeat(" ", paddingRight)
		default:
			buttonText = strings.Repeat(" ", paddingLeft) + buttonText + strings.Repeat(" ", paddingRight+availableSpace)
		}
	} else {
		buttonText = strings.Repeat(" ", paddingLeft) + buttonText + strings.Repeat(" ", paddingRight)
	}

	return buttonText
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
	return inst.behaviors.OnAction(inst, act)
}

// ResetPressed resets the pressed state of the button.
// Called by InteractionContext when new keyboard input is detected.
func (inst *Instance) ResetPressed() {
	if pressable := inst.behaviors.Get("Pressable"); pressable != nil {
		if p, ok := pressable.(*control.PressableBehavior); ok {
			p.ResetPressedWithInstance(inst)
		}
	}
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
	return inst.buttonStyle
}

// SetStyle sets the visual style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.buttonStyle = s
}

// GetProp returns a prop value.
func (inst *Instance) GetProp(key string) (interface{}, bool) {
	switch key {
	case propDisabled:
		return inst.state.Disabled, true
	case propLabel:
		return inst.label, true
	case "variant":
		return inst.variant, true
	case "size":
		return inst.size, true
	case "focusStyle":
		return inst.focusStyle, true
	case propPressIntent:
		return inst.pressIntent, true
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
// Calculates the button's ideal size given the constraints.
// This is Phase 1 of two-pass layout: measure natural size without position.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Calculate button content width: label + brackets + focus indicator
	label := inst.label
	if label == "" {
		label = " " // Empty button still has minimal width
	}

	// Width: label length + 2 for brackets "[]" + 1 for focus indicator
	contentWidth := paint.StringWidth(label) + 3

	// Height is always 1 for single-line button
	contentHeight := 1

	// Apply size modifiers
	switch inst.size {
	case SizeSmall:
		// Small button: no extra padding
	case SizeMedium:
		// Medium button: +1 padding on each side
		contentWidth += 2
	case SizeLarge:
		// Large button: +2 padding on each side
		contentWidth += 4
	}

	// Apply user-specified padding
	horizontalPadding := inst.padding[1] + inst.padding[3] // right + left
	verticalPadding := inst.padding[0] + inst.padding[2]   // top + bottom

	width := contentWidth + horizontalPadding
	height := contentHeight + verticalPadding

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	// Apply explicit style dimensions if set
	if inst.buttonStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.buttonStyle.Width)
	}
	if inst.buttonStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.buttonStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// GetNaturalSize returns the natural (unconstrained) size of the button.
// This is used for alignment calculations when the button is stretched.
func (inst *Instance) GetNaturalSize() (width, height int) {
	label := inst.label
	if label == "" {
		label = " "
	}

	width = paint.StringWidth(label) + 3 // label + brackets + focus
	height = 1

	switch inst.size {
	case SizeSmall:
		// no extra padding
	case SizeMedium:
		width += 2
	case SizeLarge:
		width += 4
	}

	return width, height
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getVariantProp(props rtui.Props, def Variant) Variant {
	if v, ok := props["variant"]; ok {
		if variant, ok := v.(Variant); ok {
			return variant
		}
	}
	return def
}

func getSizeProp(props rtui.Props, def Size) Size {
	if v, ok := props["size"]; ok {
		if size, ok := v.(Size); ok {
			return size
		}
	}
	return def
}

func getFocusStyleProp(props rtui.Props, def FocusStyle) FocusStyle {
	if v, ok := props["focusStyle"]; ok {
		if fs, ok := v.(FocusStyle); ok {
			return fs
		}
	}
	return def
}

func getPaddingProp(props rtui.Props) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return [4]int{}
}

func getTextAlignProp(props rtui.Props, def rtui.Align) rtui.Align {
	if v, ok := props["textAlign"]; ok {
		if a, ok := v.(rtui.Align); ok {
			return a
		}
	}
	return def
}
