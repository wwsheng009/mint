package button

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Button Types
// =============================================================================

// Variant represents button style variants.
type Variant int

const (
	VariantDefault Variant = iota
	VariantPrimary
	VariantSecondary
	VariantDanger
	VariantSuccess
)

// Size represents button sizes.
type Size int

const (
	SizeSmall Size = iota
	SizeMedium
	SizeLarge
)

// FocusStyle defines how a button displays focus state.
type FocusStyle int

const (
	// FocusStyleReverse uses reversed colors (default).
	FocusStyleReverse FocusStyle = iota
	// FocusStyleUnderline uses underline only (preserves background).
	FocusStyleUnderline
	// FocusStyleBracket uses brackets around the label.
	FocusStyleBracket
	// FocusStyleBold uses bold text only.
	FocusStyleBold
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the button description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	label      string
	variant    Variant
	size       Size
	focusStyle FocusStyle
	style      style.Style

	// === Layout Props ===
	padding   [4]int // top, right, bottom, left
	textAlign rtui.Align

	// === Intent Props (no closures!) ===
	pressIntent intent.Intent // Structured intent instead of func()

	// === State Props ===
	disabled bool

	// === Focus state (transient, for rendering) ===
	hasFocus bool

	// === Box Model (via interface) ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.FocusableVNode  = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Button VNode.
func New(label string) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("button"),
		label:        label,
		variant:      VariantDefault,
		size:         SizeMedium,
		focusStyle:   FocusStyleReverse,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (b *VNode) Key() string {
	return b.key
}

// SetKey sets the component key.
func (b *VNode) SetKey(key string) {
	b.key = key
}

// Tag returns the tag name.
func (b *VNode) Tag() string {
	return "button"
}

// Style returns the visual style.
func (b *VNode) Style() style.Style {
	return b.style
}

// SetStyle sets the visual style.
func (b *VNode) SetStyle(s style.Style) {
	b.style = s
}

// Children returns child nodes (button has no children).
func (b *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for button.
func (b *VNode) SetChildren(children []rtui.VNode) {
	// Button has no children
}

// GetLayer returns the rendering layer.
func (b *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer.
func (b *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return b
}

// Props returns the node properties.
func (b *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":         b.key,
		"label":       b.label,
		"variant":     b.variant,
		"size":        b.size,
		"focusStyle":  b.focusStyle,
		"style":       b.style,
		"pressIntent": b.pressIntent,
		"disabled":    b.disabled,
		"padding":     b.padding,
		"textAlign":   b.textAlign,
	}
}

// SetProps sets the node properties.
func (b *VNode) SetProps(p rtui.Props) {
	if v, ok := p["key"].(string); ok {
		b.key = v
	}
	if v, ok := p["label"].(string); ok {
		b.label = v
	}
	if v, ok := p["variant"].(Variant); ok {
		b.variant = v
	}
	if v, ok := p["size"].(Size); ok {
		b.size = v
	}
	if v, ok := p["focusStyle"].(FocusStyle); ok {
		b.focusStyle = v
	}
	if v, ok := p["style"].(style.Style); ok {
		b.style = v
	}
	if v, ok := p["pressIntent"].(intent.Intent); ok {
		b.pressIntent = v
	}
	if v, ok := p["disabled"].(bool); ok {
		b.disabled = v
	}
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new ButtonInstance from this VNode description.
func (b *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":         b.key,
		"label":       b.label,
		"variant":     b.variant,
		"size":        b.size,
		"focusStyle":  b.focusStyle,
		"style":       b.style,
		"pressIntent": b.pressIntent,
		"disabled":    b.disabled,
		"padding":     b.Padding(),
		"textAlign":   b.TextAlign(),
	}
	return NewInstance(props)
}

// =============================================================================
// FocusableVNode Interface
// =============================================================================

// SetFocus sets the focus state for rendering purposes.
func (b *VNode) SetFocus(hasFocus bool) {
	b.hasFocus = hasFocus
}

// IsFocusable returns true if button can receive focus.
func (b *VNode) IsFocusable() bool {
	return !b.disabled
}

// GetFocusID returns a unique identifier for focus.
func (b *VNode) GetFocusID() string {
	if b.key != "" {
		return "button:" + b.key
	}
	return "button:" + b.label
}

// Label returns the button label.
func (b *VNode) Label() string {
	return b.label
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetLabel sets the button label.
func (b *VNode) SetLabel(label string) *VNode {
	b.label = label
	return b
}

// SetVariant sets the button variant.
func (b *VNode) SetVariant(v Variant) *VNode {
	b.variant = v
	return b
}

// SetSize sets the button size.
func (b *VNode) SetSize(s Size) *VNode {
	b.size = s
	return b
}

// SetFocusStyle sets the focus style.
func (b *VNode) SetFocusStyle(fs FocusStyle) *VNode {
	b.focusStyle = fs
	return b
}

// SetDisabled sets the disabled state.
func (b *VNode) SetDisabled(disabled bool) *VNode {
	b.disabled = disabled
	return b
}

// SetIntent sets the press intent (replaces OnClick closure).
func (b *VNode) SetIntent(pressIntent intent.Intent) *VNode {
	b.pressIntent = pressIntent
	return b
}

// SetStyleProps sets the visual style.
func (b *VNode) SetStyleProps(s style.Style) *VNode {
	b.style = s
	return b
}

// SetPaddingProps sets the padding (top, right, bottom, left).
func (b *VNode) SetPaddingProps(top, right, bottom, left int) *VNode {
	b.BoxModelMixin.SetPadding(top, right, bottom, left)
	return b
}

// SetTextAlignProps sets the text alignment.
func (b *VNode) SetTextAlignProps(align rtui.Align) *VNode {
	b.BoxModelMixin.SetTextAlign(align)
	return b
}

// =============================================================================
// Intent Methods (replacing closures)
// =============================================================================

// OnPress sets the intent to emit when pressed.
// This replaces the old OnClick(func()) closure pattern.
//
// Example:
//
//	button.OnPress(intent.OpenModal("settings"))
func (b *VNode) OnPress(pressIntent intent.Intent) *VNode {
	return b.SetIntent(pressIntent)
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Variant returns the button variant.
func (b *VNode) Variant() Variant {
	return b.variant
}

// Size returns the button size.
func (b *VNode) Size() Size {
	return b.size
}

// FocusStyle returns the focus style.
func (b *VNode) FocusStyle() FocusStyle {
	return b.focusStyle
}

// Disabled returns the disabled state.
func (b *VNode) Disabled() bool {
	return b.disabled
}

// PressIntent returns the press intent.
func (b *VNode) PressIntent() intent.Intent {
	return b.pressIntent
}

// HasFocus returns the focus state.
func (b *VNode) HasFocus() bool {
	return b.hasFocus
}
