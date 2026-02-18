package button

import (
	"strings"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// ButtonInstance - Runtime Instance for Button
// =============================================================================
// ButtonInstance is the runtime entity for Button components.
// It persists across renders and holds all state (focus, hover, etc.)
//
// Architecture (from fiber_paint.md):
//
//	ButtonVNode (description) ──CreateInstance()──→ ButtonInstance
//	                                                 ↓
//	                                           Fiber.Instance = instance
//	                                                 ↓
//	                                           Persists across renders
//
// Key Points:
//   - ButtonVNode is created every render (description only)
//   - ButtonInstance is created once and persists (stateful)
//   - Fiber.Instance holds ButtonInstance reference
//   - During commit phase: instance.Paint(x, y)

// Ensure ButtonInstance implements required interfaces
var (
	_ rtui.ComponentInstance   = (*ButtonInstance)(nil)
	_ rtui.PaintableInstance   = (*ButtonInstance)(nil)
	_ rtui.FocusableInstance   = (*ButtonInstance)(nil)
	_ rtui.ActionHandlerInstance = (*ButtonInstance)(nil)
)

// ButtonInstance is the runtime instance for Button components.
type ButtonInstance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label       string
	variant     ButtonVariant
	size        ButtonSize
	focusStyle  ButtonFocusStyle
	buttonStyle style.Style
	onClick     func()
	clickAction action.ActionType
	padding     [4]int
	textAlign   rtui.Align

	// === Runtime State (managed by instance) ===
	hasFocus  bool
	isHovered bool
	disabled  bool
	bounds    [4]int

	// === Dirty flag ===
	dirty bool
}

// NewButtonInstance creates a new ButtonInstance from props
func NewButtonInstance(props rtui.Props) *ButtonInstance {
	inst := &ButtonInstance{
		key:         getStringProp(props, "key", ""),
		label:       getStringProp(props, "label", ""),
		variant:     getVariantProp(props, ButtonVariantDefault),
		size:        getSizeProp(props, ButtonSizeMedium),
		focusStyle:  getFocusStyleProp(props, FocusStyleReverse),
		buttonStyle: getStyleProp(props),
		onClick:     getOnClickProp(props),
		clickAction: getClickActionProp(props),
		padding:     getPaddingProp(props),
		textAlign:   getTextAlignProp(props, rtui.AlignStart),
		disabled:    getBoolProp(props, "disabled", false),
		dirty:       true,
	}
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// Key implements ComponentInstance
func (inst *ButtonInstance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance
func (inst *ButtonInstance) SetKey(key string) {
	inst.key = key
}

// Init implements ComponentInstance
func (inst *ButtonInstance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy implements ComponentInstance
func (inst *ButtonInstance) Destroy() {
	// Clean up any resources
	inst.onClick = nil
}

// OnMount implements ComponentInstance
func (inst *ButtonInstance) OnMount() {
	// Called when mounted
}

// OnUnmount implements ComponentInstance
func (inst *ButtonInstance) OnUnmount() {
	// Called when unmounted
}

// SetProps implements ComponentInstance
func (inst *ButtonInstance) SetProps(props rtui.Props) bool {
	oldLabel := inst.label
	oldVariant := inst.variant
	oldSize := inst.size
	oldDisabled := inst.disabled
	oldFocusStyle := inst.focusStyle

	inst.label = getStringProp(props, "label", inst.label)
	inst.variant = getVariantProp(props, inst.variant)
	inst.size = getSizeProp(props, inst.size)
	inst.focusStyle = getFocusStyleProp(props, inst.focusStyle)
	inst.buttonStyle = getStyleProp(props)
	inst.onClick = getOnClickProp(props)
	inst.clickAction = getClickActionProp(props)
	inst.padding = getPaddingProp(props)
	inst.textAlign = getTextAlignProp(props, inst.textAlign)
	inst.disabled = getBoolProp(props, "disabled", inst.disabled)

	// Check if props changed
	changed := oldLabel != inst.label ||
		oldVariant != inst.variant ||
		oldSize != inst.size ||
		oldDisabled != inst.disabled ||
		oldFocusStyle != inst.focusStyle

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance
func (inst *ButtonInstance) GetProps() rtui.Props {
	return rtui.Props{
		"key":        inst.key,
		"label":      inst.label,
		"variant":    inst.variant,
		"size":       inst.size,
		"focusStyle": inst.focusStyle,
		"disabled":   inst.disabled,
	}
}

// MarkDirty implements ComponentInstance
func (inst *ButtonInstance) MarkDirty() {
	inst.dirty = true
}

// IsDirty implements ComponentInstance
func (inst *ButtonInstance) IsDirty() bool {
	return inst.dirty
}

// GetContext implements ComponentInstance (no hooks for Button)
func (inst *ButtonInstance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance
func (inst *ButtonInstance) Paint(x, y int) []paint.DrawCmd {
	// Build button label with brackets
	displayLabel := inst.label
	if displayLabel == "" {
		displayLabel = " "
	}

	// Format: [label]
	var labelText string
	switch inst.size {
	case ButtonSizeSmall:
		labelText = "[" + displayLabel + "]"
	case ButtonSizeMedium:
		labelText = "[ " + displayLabel + " ]"
	case ButtonSizeLarge:
		labelText = "[  " + displayLabel + "  ]"
	default:
		labelText = "[" + displayLabel + "]"
	}

	// Apply variant-based styling if not explicitly set
	buttonStyle := inst.buttonStyle
	if buttonStyle.FG == "" && buttonStyle.BG == "" {
		switch inst.variant {
		case ButtonVariantPrimary:
			buttonStyle = buttonStyle.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
		case ButtonVariantSecondary:
			buttonStyle = buttonStyle.Foreground(theme.Text()).Background(theme.Surface())
		case ButtonVariantDanger:
			buttonStyle = buttonStyle.Foreground(theme.BG()).Background(theme.Error()).Bold(true)
		case ButtonVariantSuccess:
			buttonStyle = buttonStyle.Foreground(theme.BG()).Background(theme.Success()).Bold(true)
		case ButtonVariantDefault:
			buttonStyle = buttonStyle.Foreground(theme.Text()).Background(theme.Surface())
		}
	}

	// Apply disabled state
	if inst.disabled {
		buttonStyle = buttonStyle.Foreground(theme.DisabledFG()).Background(theme.DisabledBG())
	}

	// State priority: Focused > Hovered > Normal
	if inst.hasFocus && !inst.disabled {
		switch inst.focusStyle {
		case FocusStyleUnderline:
			buttonStyle = buttonStyle.
				Foreground(theme.FocusBright()).
				Underline(true).
				Bold(true)
		case FocusStyleBracket:
			customBG := buttonStyle.BG
			buttonStyle = buttonStyle.
				Foreground(theme.FocusBright()).
				Bold(true)
			if customBG != "" {
				buttonStyle = buttonStyle.Background(customBG)
			}
		case FocusStyleBold:
			buttonStyle = buttonStyle.Bold(true)
		case FocusStyleReverse:
			buttonStyle = buttonStyle.Foreground(theme.Foreground()).Background(theme.Focus()).Bold(true)
		}
	} else if inst.isHovered && !inst.disabled {
		buttonStyle = buttonStyle.Underline(true)
	}

	// Add focus indicator
	var focusIndicator string
	if inst.hasFocus && !inst.disabled {
		if inst.focusStyle == FocusStyleReverse {
			focusIndicator = "*"
		} else if inst.focusStyle == FocusStyleUnderline || inst.focusStyle == FocusStyleBracket {
			focusIndicator = ">"
		} else {
			focusIndicator = " "
		}
	} else {
		focusIndicator = " "
	}

	// Build button text
	buttonText := focusIndicator + labelText
	contentWidth := len(buttonText)

	// Get padding
	paddingLeft := inst.padding[3]
	paddingRight := inst.padding[1]

	naturalWidth := contentWidth
	layoutWidth := naturalWidth

	// Apply text alignment if button is stretched
	if layoutWidth > naturalWidth {
		availableSpace := layoutWidth - naturalWidth
		switch inst.textAlign {
		case rtui.AlignCenter:
			leftSpace := paddingLeft + availableSpace/2
			rightSpace := paddingRight + (availableSpace - availableSpace/2)
			buttonText = strings.Repeat(" ", leftSpace) + buttonText +
				strings.Repeat(" ", rightSpace)
		case rtui.AlignEnd:
			leftSpace := paddingLeft + availableSpace
			buttonText = strings.Repeat(" ", leftSpace) + buttonText +
				strings.Repeat(" ", paddingRight)
		default:
			buttonText = strings.Repeat(" ", paddingLeft) + buttonText +
				strings.Repeat(" ", paddingRight+availableSpace)
		}
	} else {
		buttonText = strings.Repeat(" ", paddingLeft) + buttonText +
			strings.Repeat(" ", paddingRight)
	}

	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  buttonText,
		Style: buttonStyle,
	}}
}

// =============================================================================
// FocusableInstance Interface
// =============================================================================

// SetFocus implements FocusableInstance
func (inst *ButtonInstance) SetFocus(focused bool) {
	if inst.hasFocus != focused {
		inst.hasFocus = focused
		inst.dirty = true
	}
}

// HasFocus implements FocusableInstance
func (inst *ButtonInstance) HasFocus() bool {
	return inst.hasFocus
}

// IsDisabled implements FocusableInstance
func (inst *ButtonInstance) IsDisabled() bool {
	return inst.disabled
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

// CanHandleAction implements ActionHandlerInstance
func (inst *ButtonInstance) CanHandleAction(actionType string) bool {
	if inst.disabled {
		return false
	}
	return actionType == string(action.ActionClick) ||
		actionType == string(action.ActionEnter)
}

// HandleAction implements ActionHandlerInstance
func (inst *ButtonInstance) HandleAction(actionType string, payload interface{}) bool {
	if inst.disabled {
		return false
	}

	if actionType == string(action.ActionClick) || actionType == string(action.ActionEnter) {
		if inst.onClick != nil {
			inst.onClick()
			return true
		}
	}
	return false
}

// =============================================================================
// Additional Methods
// =============================================================================

// SetHover sets the hover state
func (inst *ButtonInstance) SetHover(hovered bool) {
	if inst.isHovered != hovered {
		inst.isHovered = hovered
		inst.dirty = true
	}
}

// SetBounds sets the layout bounds
func (inst *ButtonInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// ClearDirty clears the dirty flag
func (inst *ButtonInstance) ClearDirty() {
	inst.dirty = false
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

func getVariantProp(props rtui.Props, def ButtonVariant) ButtonVariant {
	if v, ok := props["variant"]; ok {
		if variant, ok := v.(ButtonVariant); ok {
			return variant
		}
	}
	return def
}

func getSizeProp(props rtui.Props, def ButtonSize) ButtonSize {
	if v, ok := props["size"]; ok {
		if size, ok := v.(ButtonSize); ok {
			return size
		}
	}
	return def
}

func getFocusStyleProp(props rtui.Props, def ButtonFocusStyle) ButtonFocusStyle {
	if v, ok := props["focusStyle"]; ok {
		if fs, ok := v.(ButtonFocusStyle); ok {
			return fs
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

func getOnClickProp(props rtui.Props) func() {
	if v, ok := props["onClick"]; ok {
		if fn, ok := v.(func()); ok {
			return fn
		}
	}
	return nil
}

func getClickActionProp(props rtui.Props) action.ActionType {
	if v, ok := props["clickAction"]; ok {
		if a, ok := v.(action.ActionType); ok {
			return a
		}
	}
	return ""
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
