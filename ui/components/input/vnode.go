package input

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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
	changeIntent  intent.Intent // emitted when value changes
	submitIntent  intent.Intent // emitted on Enter

	// === State Props (declarative, actual state managed by Instance) ===
	value    string
	maxLen   int
	disabled bool
	readOnly bool

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
		"key":          i.key,
		"placeholder":  i.placeholder,
		"inputType":    i.inputType,
		"style":        i.style,
		"width":        i.width,
		"borderStyle":  i.borderStyle,
		"changeIntent": i.changeIntent,
		"submitIntent": i.submitIntent,
		"value":        i.value,
		"maxLen":       i.maxLen,
		"disabled":     i.disabled,
		"readOnly":     i.readOnly,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (i *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		i.key = v
	}
	if v, ok := p["placeholder"].(string); ok {
		i.placeholder = v
	}
	if v, ok := p["inputType"].(Type); ok {
		i.inputType = v
	}
	if v, ok := p["style"].(style.Style); ok {
		i.style = v
	}
	if v, ok := p["width"].(int); ok {
		i.width = v
	}
	if v, ok := p["borderStyle"].(layout.BorderStyle); ok {
		i.borderStyle = v
	}
	if v, ok := p["changeIntent"].(intent.Intent); ok {
		i.changeIntent = v
	}
	if v, ok := p["submitIntent"].(intent.Intent); ok {
		i.submitIntent = v
	}
	if v, ok := p["value"].(string); ok {
		i.value = v
	}
	if v, ok := p["maxLen"].(int); ok {
		i.maxLen = v
	}
	if v, ok := p["disabled"].(bool); ok {
		i.disabled = v
	}
	if v, ok := p["readOnly"].(bool); ok {
		i.readOnly = v
	}
	return i
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new InputInstance from this VNode description.
func (i *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":          i.key,
		"placeholder":  i.placeholder,
		"inputType":    i.inputType,
		"style":        i.style,
		"width":        i.width,
		"borderStyle":  i.borderStyle,
		"changeIntent": i.changeIntent,
		"submitIntent": i.submitIntent,
		"value":        i.value,
		"maxLen":       i.maxLen,
		"disabled":     i.disabled,
		"readOnly":     i.readOnly,
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
