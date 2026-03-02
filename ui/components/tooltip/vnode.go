package tooltip

import (
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Tooltip Position
// =============================================================================

// Position defines where the tooltip appears relative to its anchor.
type Position int

const (
	PositionTop    Position = iota // Tooltip appears above anchor
	PositionBottom                 // Tooltip appears below anchor
	PositionLeft                   // Tooltip appears to the left of anchor
	PositionRight                  // Tooltip appears to the right of anchor
	PositionAuto                   // Position is automatically determined
)

// =============================================================================
// Toast Type
// =============================================================================

// ToastType defines the type of toast notification.
type ToastType int

const (
	ToastInfo    ToastType = iota // Info toast
	ToastSuccess                  // Success toast
	ToastWarning                  // Warning toast
	ToastError                    // Error toast
)

// =============================================================================
// Tooltip VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the tooltip description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	text     string
	style    style.Style
	position Position
	delay    time.Duration

	// === Rendering Layer ===
	layer rtui.Layer // Layer for Z-order: Base, Overlay, Modal, Tooltip, Inspector

	// === Content Props ===
	content rtui.VNode // Child content that triggers tooltip
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Tooltip VNode wrapping content.
func New(content rtui.VNode, text string) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("tooltip"),
		content:      content,
		text:         text,
		position:     PositionAuto,
		delay:        500 * time.Millisecond,
		layer:        rtui.LayerTooltip, // Default to Tooltip layer
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (t *VNode) Key() string {
	return t.key
}

// SetKey sets the component key - returns VNode for chaining.
func (t *VNode) SetKey(key string) rtui.VNode {
	t.key = key
	return t
}

// Tag returns the tag name.
func (t *VNode) Tag() string {
	return "tooltip"
}

// Style returns the visual style.
func (t *VNode) Style() style.Style {
	return t.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (t *VNode) SetStyle(s style.Style) rtui.VNode {
	t.style = s
	return t
}

// Children returns the wrapped content as a child.
func (t *VNode) Children() []rtui.VNode {
	if t.content != nil {
		return []rtui.VNode{t.content}
	}
	return nil
}

// SetChildren sets the wrapped content - returns VNode for chaining.
func (t *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		t.content = children[0]
	}
	return t
}

// GetLayer returns the rendering layer (tooltips appear above content).
func (t *VNode) GetLayer() rtui.Layer {
	return t.layer
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	t.layer = l
	return t
}

// Props returns the node properties.
func (t *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":      t.key,
		"text":     t.text,
		"position": t.position,
		"delay":    t.delay,
		"style":    t.style,
		"layer":    t.layer,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (t *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		t.key = v
	}
	if v, ok := p["text"].(string); ok {
		t.text = v
	}
	if v, ok := p["position"].(Position); ok {
		t.position = v
	}
	if v, ok := p["delay"].(time.Duration); ok {
		t.delay = v
	}
	if v, ok := p["style"].(style.Style); ok {
		t.style = v
	}
	if v, ok := p["layer"].(rtui.Layer); ok {
		t.layer = v
	}
	return t
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new Instance from this VNode description.
func (t *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":      t.key,
		"text":     t.text,
		"position": t.position,
		"delay":    t.delay,
		"style":    t.style,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API
// =============================================================================

// SetText sets the tooltip text.
func (t *VNode) SetText(text string) *VNode {
	t.text = text
	return t
}

// SetPosition sets the tooltip position.
func (t *VNode) SetPosition(position Position) *VNode {
	t.position = position
	return t
}

// SetDelay sets the delay before showing the tooltip.
func (t *VNode) SetDelay(delay time.Duration) *VNode {
	t.delay = delay
	return t
}

// SetStyleProps sets the visual style.
func (t *VNode) SetStyleProps(s style.Style) *VNode {
	t.style = s
	return t
}

// =============================================================================
// Props Accessors
// =============================================================================

// Text returns the tooltip text.
func (t *VNode) Text() string {
	return t.text
}

// Position returns the tooltip position.
func (t *VNode) Position() Position {
	return t.position
}

// Delay returns the delay before showing.
func (t *VNode) Delay() time.Duration {
	return t.delay
}

// Content returns the wrapped content node.
func (t *VNode) Content() rtui.VNode {
	return t.content
}

// =============================================================================
// Toast VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// ToastVNode is the toast notification description.
type ToastVNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	title     string
	message   string
	toastType ToastType
	style     style.Style

	// === Rendering Layer ===
	layer rtui.Layer // Layer for Z-order: Base, Overlay, Modal, Tooltip, Inspector

	// === Timing Props ===
	duration time.Duration

	// === Intent Props (no closures!) ===
	closeIntent interface{} // Intent to emit when closed

	// === Box Model ===
	rtui.BoxModelMixin
}

// Ensure ToastVNode implements required interfaces
var (
	_ rtui.VNode           = (*ToastVNode)(nil)
	_ rtui.InstanceFactory = (*ToastVNode)(nil)
	_ rtui.BoxModel        = (*ToastVNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewToast creates a new Toast VNode.
func NewToast(message string) *ToastVNode {
	return &ToastVNode{
		ElementVNode: rtui.NewElement("toast"),
		message:      message,
		toastType:    ToastInfo,
		duration:     3000 * time.Millisecond,
		layer:        rtui.LayerOverlay, // Default to Overlay layer
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (t *ToastVNode) Key() string {
	return t.key
}

// SetKey sets the component key - returns ToastVNode for chaining.
func (t *ToastVNode) SetKey(key string) rtui.VNode {
	t.key = key
	return t
}

// Tag returns the tag name.
func (t *ToastVNode) Tag() string {
	return "toast"
}

// Style returns the visual style.
func (t *ToastVNode) Style() style.Style {
	return t.style
}

// SetStyle sets the visual style - returns ToastVNode for chaining.
func (t *ToastVNode) SetStyle(s style.Style) rtui.VNode {
	t.style = s
	return t
}

// Children returns child nodes (toast has no children).
func (t *ToastVNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for toast - returns ToastVNode for chaining.
func (t *ToastVNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Toast has no children
	return t
}

// GetLayer returns the rendering layer (toasts appear above content).
func (t *ToastVNode) GetLayer() rtui.Layer {
	return t.layer
}

// SetLayer sets the rendering layer - returns ToastVNode for chaining.
func (t *ToastVNode) SetLayer(l rtui.Layer) rtui.VNode {
	t.layer = l
	return t
}

// Props returns the node properties.
func (t *ToastVNode) Props() rtui.Props {
	return rtui.Props{
		"key":         t.key,
		"title":       t.title,
		"message":     t.message,
		"toastType":   t.toastType,
		"duration":    t.duration,
		"closeIntent": t.closeIntent,
		"style":       t.style,
		"padding":     t.Padding(),
		"layer":       t.layer,
	}
}

// SetProps sets the node properties - returns ToastVNode for chaining.
func (t *ToastVNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		t.key = v
	}
	if v, ok := p["title"].(string); ok {
		t.title = v
	}
	if v, ok := p["message"].(string); ok {
		t.message = v
	}
	if v, ok := p["toastType"].(ToastType); ok {
		t.toastType = v
	}
	if v, ok := p["duration"].(time.Duration); ok {
		t.duration = v
	}
	if v, ok := p["closeIntent"].(interface{}); ok {
		t.closeIntent = v
	}
	if v, ok := p["style"].(style.Style); ok {
		t.style = v
	}
	if v, ok := p["layer"].(rtui.Layer); ok {
		t.layer = v
	}
	return t
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new ToastInstance from this ToastVNode description.
func (t *ToastVNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		"key":         t.key,
		"title":       t.title,
		"message":     t.message,
		"toastType":   t.toastType,
		"duration":    t.duration,
		"closeIntent": t.closeIntent,
		"style":       t.style,
		"padding":     t.Padding(),
	}
	return NewToastInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API
// =============================================================================

// SetTitle sets the toast title.
func (t *ToastVNode) SetTitle(title string) *ToastVNode {
	t.title = title
	return t
}

// SetMessage sets the toast message.
func (t *ToastVNode) SetMessage(message string) *ToastVNode {
	t.message = message
	return t
}

// SetType sets the toast type.
func (t *ToastVNode) SetType(toastType ToastType) *ToastVNode {
	t.toastType = toastType
	return t
}

// SetDuration sets how long the toast is visible.
func (t *ToastVNode) SetDuration(duration time.Duration) *ToastVNode {
	t.duration = duration
	return t
}

// SetCloseIntent sets the intent to emit when the toast is closed.
func (t *ToastVNode) SetCloseIntent(closeIntent interface{}) *ToastVNode {
	t.closeIntent = closeIntent
	return t
}

// SetStyleProps sets the visual style.
func (t *ToastVNode) SetStyleProps(s style.Style) *ToastVNode {
	t.style = s
	return t
}

// SetPaddingProps sets the padding (top, right, bottom, left).
func (t *ToastVNode) SetPaddingProps(top, right, bottom, left int) *ToastVNode {
	t.BoxModelMixin.SetPadding(top, right, bottom, left)
	return t
}

// Info sets toast type to info.
func (t *ToastVNode) Info() *ToastVNode {
	t.toastType = ToastInfo
	return t
}

// Success sets toast type to success.
func (t *ToastVNode) Success() *ToastVNode {
	t.toastType = ToastSuccess
	return t
}

// Warning sets toast type to warning.
func (t *ToastVNode) Warning() *ToastVNode {
	t.toastType = ToastWarning
	return t
}

// Error sets toast type to error.
func (t *ToastVNode) Error() *ToastVNode {
	t.toastType = ToastError
	return t
}

// =============================================================================
// Props Accessors
// =============================================================================

// Title returns the toast title.
func (t *ToastVNode) Title() string {
	return t.title
}

// Message returns the toast message.
func (t *ToastVNode) Message() string {
	return t.message
}

// ToastType returns the toast type.
func (t *ToastVNode) ToastType() ToastType {
	return t.toastType
}

// Duration returns the toast duration.
func (t *ToastVNode) Duration() time.Duration {
	return t.duration
}

// CloseIntent returns the close intent.
func (t *ToastVNode) CloseIntent() interface{} {
	return t.closeIntent
}

// Toast is a convenience function that creates a new toast VNode.
// This is the public API for creating toasts.
func Toast(message string) *ToastVNode {
	return NewToast(message)
}

// =============================================================================
// layout.BoxModelProvider Implementation for ToastVNode
// =============================================================================

// GetBoxModel returns the box model for the ToastVNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: ToastVNode uses BoxModelMixin for padding/margin, and has no border.
func (t *ToastVNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   t.BoxModelMixin.Padding()[3],
			Right:  t.BoxModelMixin.Padding()[1],
			Top:    t.BoxModelMixin.Padding()[0],
			Bottom: t.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   t.BoxModelMixin.Margin()[3],
			Right:  t.BoxModelMixin.Margin()[1],
			Top:    t.BoxModelMixin.Margin()[0],
			Bottom: t.BoxModelMixin.Margin()[2],
		},
		// Toast typically doesn't have a border
		Border: layout.Border{Style: layout.BorderNone},
	}
}
