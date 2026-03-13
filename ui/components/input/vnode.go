package input

import (
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/cursor"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propBorderStyle = "borderStyle"
	propChangeIntent = "changeIntent"
	propCursorConfig = "cursorConfig"
	propDisabled = "disabled"
	propFormID = "formID"
	propInputType = "inputType"
	propKey = "key"
	propMaxLen = "maxLen"
	propPlaceholder = "placeholder"
	propReadOnly = "readOnly"
	propStyle = "style"
	propSubmitIntent = "submitIntent"
	propValue = "value"
	propWidth = "width"
)

// =============================================================================
// Input Types
// =============================================================================

// Type represents the type of input.
type Type int

const (
	TypeText Type = iota
	TypePassword
	TypeNumber
	TypeEmail
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the input description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	placeholder string
	inputType   Type
	style       style.Style

	// === Layout Props ===
	width int // explicit width (0 = auto)

	// === Border Props ===
	borderStyle layout.BorderStyle

	// === Intent Props (no closures!) ===
	changeIntent intent.Intent // emitted when value changes
	submitIntent intent.Intent // emitted on Enter
	formID       string        // Form ID for Form integration (Phase 6)

	// === State Props (declarative, actual state managed by Instance) ===
	value        string
	maxLen       int
	disabled     bool
	readOnly     bool
	cursorConfig cursor.Config

	// === Box Model (via interface) ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Input VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("input"),
		inputType:    TypeText,
		borderStyle:  layout.BorderSingle, // Default border
		cursorConfig: cursor.DefaultConfig(),
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (i *VNode) Key() string {
	return i.key
}

// SetKey sets the component key - returns VNode for chaining.
func (i *VNode) SetKey(key string) rtui.VNode {
	i.key = key
	return i
}

// Tag returns the tag name.
func (i *VNode) Tag() string {
	return "input"
}

// Style returns the visual style.
func (i *VNode) Style() style.Style {
	return i.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (i *VNode) SetStyle(s style.Style) rtui.VNode {
	i.style = s
	return i
}

// Children returns child nodes (input has no children).
func (i *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for input - returns VNode for chaining.
func (i *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	return i
}

// GetLayer returns the rendering layer.
func (i *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (i *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return i
}

// Props returns the node properties.
func (i *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:          i.key,
		propPlaceholder:  i.placeholder,
		propInputType:    i.inputType,
		propStyle:        i.style,
		propWidth:        i.width,
		propBorderStyle:  i.borderStyle,
		propChangeIntent: i.changeIntent,
		propSubmitIntent: i.submitIntent,
		propFormID:       i.formID,
		propValue:        i.value,
		propMaxLen:       i.maxLen,
		propDisabled:     i.disabled,
		propReadOnly:     i.readOnly,
		propCursorConfig: i.cursorConfig,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (i *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		i.key = v
	}
	if v, ok := p[propPlaceholder].(string); ok {
		i.placeholder = v
	}
	if v, ok := p[propInputType].(Type); ok {
		i.inputType = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		i.style = v
	}
	if v, ok := p[propWidth].(int); ok {
		i.width = v
	}
	if v, ok := p[propBorderStyle].(layout.BorderStyle); ok {
		i.borderStyle = v
	}
	if v, ok := p[propChangeIntent].(intent.Intent); ok {
		i.changeIntent = v
	}
	if v, ok := p[propSubmitIntent].(intent.Intent); ok {
		i.submitIntent = v
	}
	if v, ok := p[propFormID].(string); ok {
		i.formID = v
	}
	if v, ok := p[propValue].(string); ok {
		i.value = v
	}
	if v, ok := p[propMaxLen].(int); ok {
		i.maxLen = v
	}
	if v, ok := p[propDisabled].(bool); ok {
		i.disabled = v
	}
	if v, ok := p[propReadOnly].(bool); ok {
		i.readOnly = v
	}
	if v, ok := p[propCursorConfig].(cursor.Config); ok {
		i.cursorConfig = cursor.NormalizeConfig(v)
	}
	return i
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new InputInstance from this VNode description.
func (i *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		propKey:          i.key,
		propPlaceholder:  i.placeholder,
		propInputType:    i.inputType,
		propStyle:        i.style,
		propWidth:        i.width,
		propBorderStyle:  i.borderStyle,
		propChangeIntent: i.changeIntent,
		propSubmitIntent: i.submitIntent,
		propFormID:       i.formID,
		propValue:        i.value,
		propMaxLen:       i.maxLen,
		propDisabled:     i.disabled,
		propReadOnly:     i.readOnly,
		propCursorConfig: i.cursorConfig,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetPlaceholder sets the placeholder text.
func (i *VNode) SetPlaceholder(text string) *VNode {
	i.placeholder = text
	return i
}

// SetValue sets the initial value.
func (i *VNode) SetValue(value string) *VNode {
	i.value = value
	return i
}

// SetType sets the input type.
func (i *VNode) SetType(t Type) *VNode {
	i.inputType = t
	return i
}

// SetPassword sets the input type to password.
func (i *VNode) SetPassword() *VNode {
	i.inputType = TypePassword
	return i
}

// SetMaxLen sets the maximum length (0 = no limit).
func (i *VNode) SetMaxLen(len int) *VNode {
	i.maxLen = len
	return i
}

// SetDisabled sets the disabled state.
func (i *VNode) SetDisabled(disabled bool) *VNode {
	i.disabled = disabled
	return i
}

// SetReadOnly sets the read-only state.
func (i *VNode) SetReadOnly(readOnly bool) *VNode {
	i.readOnly = readOnly
	return i
}

// SetCursorConfig sets cursor blink/shape config for the embedded caret.
func (i *VNode) SetCursorConfig(cfg cursor.Config) *VNode {
	i.cursorConfig = cursor.NormalizeConfig(cfg)
	return i
}

// SetCursorShape sets the embedded caret shape.
func (i *VNode) SetCursorShape(shape cursor.Shape) *VNode {
	i.cursorConfig.Shape = shape
	return i
}

// SetInsertCursor configures a thin insertion caret.
func (i *VNode) SetInsertCursor() *VNode {
	i.cursorConfig.Shape = cursor.ShapeBar
	i.cursorConfig.Glyph = "|"
	return i
}

// SetBlockCursor configures a block caret.
func (i *VNode) SetBlockCursor() *VNode {
	i.cursorConfig.Shape = cursor.ShapeBlock
	i.cursorConfig.Glyph = ""
	return i
}

// SetUnderlineCursor configures an underline caret.
func (i *VNode) SetUnderlineCursor() *VNode {
	i.cursorConfig.Shape = cursor.ShapeUnderline
	i.cursorConfig.Glyph = ""
	return i
}

// SetCursorBlink enables or disables caret blink.
func (i *VNode) SetCursorBlink(enabled bool) *VNode {
	i.cursorConfig.Blink = enabled
	if i.cursorConfig.BlinkInterval <= 0 {
		i.cursorConfig.BlinkInterval = cursor.NormalBlinkInterval
	}
	return i
}

// SetCursorBlinkInterval sets caret blink interval.
func (i *VNode) SetCursorBlinkInterval(interval time.Duration) *VNode {
	i.cursorConfig.BlinkInterval = interval
	return i
}

// SetWidth sets the explicit width.
func (i *VNode) SetWidth(width int) *VNode {
	i.width = width
	return i
}

// SetBorderStyle sets the border style.
func (i *VNode) SetBorderStyle(style layout.BorderStyle) *VNode {
	i.borderStyle = style
	return i
}

// SetNoBorder removes the border.
func (i *VNode) SetNoBorder() *VNode {
	i.borderStyle = layout.BorderNone
	return i
}

// SetChangeIntent sets the change intent.
func (i *VNode) SetChangeIntent(changeIntent intent.Intent) *VNode {
	i.changeIntent = changeIntent
	return i
}

// SetSubmitIntent sets the submit intent.
func (i *VNode) SetSubmitIntent(submitIntent intent.Intent) *VNode {
	i.submitIntent = submitIntent
	return i
}

// SetFormID sets the form ID for Form integration.
// When set, the component will emit FormFieldChangeIntent/FormFieldBlurIntent.
func (i *VNode) SetFormID(formID string) *VNode {
	i.formID = formID
	return i
}

// SetStyleProps sets the visual style.
func (i *VNode) SetStyleProps(s style.Style) *VNode {
	i.style = s
	return i
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Placeholder returns the placeholder text.
func (i *VNode) Placeholder() string {
	return i.placeholder
}

// InputType returns the input type.
func (i *VNode) InputType() Type {
	return i.inputType
}

// Value returns the initial value.
func (i *VNode) Value() string {
	return i.value
}

// MaxLen returns the maximum length.
func (i *VNode) MaxLen() int {
	return i.maxLen
}

// Disabled returns the disabled state.
func (i *VNode) Disabled() bool {
	return i.disabled
}

// ReadOnly returns the read-only state.
func (i *VNode) ReadOnly() bool {
	return i.readOnly
}

// Width returns the explicit width.
func (i *VNode) Width() int {
	return i.width
}

// BorderStyle returns the border style.
func (i *VNode) BorderStyle() layout.BorderStyle {
	return i.borderStyle
}

// ChangeIntent returns the change intent.
func (i *VNode) ChangeIntent() intent.Intent {
	return i.changeIntent
}

// SubmitIntent returns the submit intent.
func (i *VNode) SubmitIntent() intent.Intent {
	return i.submitIntent
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Input VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Input uses BoxModelMixin for padding/margin, and borderStyle property.
func (i *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   i.BoxModelMixin.Padding()[3],
			Right:  i.BoxModelMixin.Padding()[1],
			Top:    i.BoxModelMixin.Padding()[0],
			Bottom: i.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   i.BoxModelMixin.Margin()[3],
			Right:  i.BoxModelMixin.Margin()[1],
			Top:    i.BoxModelMixin.Margin()[0],
			Bottom: i.BoxModelMixin.Margin()[2],
		},
		Border: layout.NewBorder(i.borderStyle),
	}
}
