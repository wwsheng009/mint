package toast

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Toast Builder - Fluent API
// =============================================================================

// ToastBuilder provides a fluent API for creating Toast VNodes.
type ToastBuilder struct {
	node *ToastVNode
}

// NewToastBuilder creates a new Toast builder.
func NewToastBuilder(message string) *ToastBuilder {
	return &ToastBuilder{
		node: NewToast(message),
	}
}

// Key sets the key for diffing.
func (b *ToastBuilder) Key(key string) *ToastBuilder {
	b.node.SetKey(key)
	return b
}

// Title sets the toast title.
func (b *ToastBuilder) Title(title string) *ToastBuilder {
	b.node.SetTitle(title)
	return b
}

// Message sets the toast message.
func (b *ToastBuilder) Message(msg string) *ToastBuilder {
	b.node.SetMessage(msg)
	return b
}

// Type sets the toast type.
func (b *ToastBuilder) Type(toastType ToastType) *ToastBuilder {
	b.node.SetType(toastType)
	return b
}

// Info sets toast type to info.
func (b *ToastBuilder) Info() *ToastBuilder {
	b.node.Info()
	return b
}

// Success sets toast type to success.
func (b *ToastBuilder) Success() *ToastBuilder {
	b.node.Success()
	return b
}

// Warning sets toast type to warning.
func (b *ToastBuilder) Warning() *ToastBuilder {
	b.node.Warning()
	return b
}

// Error sets toast type to error.
func (b *ToastBuilder) Error() *ToastBuilder {
	b.node.Error()
	return b
}

// Duration sets how long the toast is visible.
func (b *ToastBuilder) Duration(d time.Duration) *ToastBuilder {
	b.node.SetDuration(d)
	return b
}

// CloseIntent sets the intent to emit when the toast is closed.
func (b *ToastBuilder) CloseIntent(closeIntent interface{}) *ToastBuilder {
	b.node.SetCloseIntent(closeIntent)
	return b
}

// Style sets the visual style.
func (b *ToastBuilder) Style(s style.Style) *ToastBuilder {
	b.node.SetStyleProps(s)
	return b
}

// FgColor sets the foreground color.
func (b *ToastBuilder) FgColor(c interface{}) *ToastBuilder {
	s := b.node.Style()
	switch v := c.(type) {
	case string:
		s.FG = style.Color(v)
	case style.Color:
		s.FG = v
	}
	b.node.SetStyleProps(s)
	return b
}

// BgColor sets the background color.
func (b *ToastBuilder) BgColor(c interface{}) *ToastBuilder {
	s := b.node.Style()
	switch v := c.(type) {
	case string:
		s.BG = style.Color(v)
	case style.Color:
		s.BG = v
	}
	b.node.SetStyleProps(s)
	return b
}

// Padding sets the padding (top, right, bottom, left).
func (b *ToastBuilder) Padding(top, right, bottom, left int) *ToastBuilder {
	b.node.SetPaddingProps(top, right, bottom, left)
	return b
}

// PaddingAll sets same padding on all sides.
func (b *ToastBuilder) PaddingAll(p int) *ToastBuilder {
	b.node.SetPaddingProps(p, p, p, p)
	return b
}

// Build returns the Toast as rtui.VNode interface.
func (b *ToastBuilder) Build() rtui.VNode {
	return b.node
}

// Layer sets the rendering layer for the toast.
func (b *ToastBuilder) Layer(l rtui.Layer) *ToastBuilder {
	b.node.SetLayer(l)
	return b
}

// SetRenderLayer is an alias for Layer for backward compatibility.
func (b *ToastBuilder) SetRenderLayer(l rtui.Layer) *ToastBuilder {
	return b.Layer(l)
}

// BaseLayer sets the toast to the base layer.
func (b *ToastBuilder) BaseLayer() *ToastBuilder {
	return b.Layer(rtui.LayerBase)
}

// OverlayLayer sets the toast to the overlay layer (default).
func (b *ToastBuilder) OverlayLayer() *ToastBuilder {
	return b.Layer(rtui.LayerOverlay)
}

// ModalLayer sets the toast to the modal layer.
func (b *ToastBuilder) ModalLayer() *ToastBuilder {
	return b.Layer(rtui.LayerModal)
}

// TooltipLayer sets the toast to the tooltip layer.
func (b *ToastBuilder) TooltipLayer() *ToastBuilder {
	return b.Layer(rtui.LayerTooltip)
}

// InspectorLayer sets the toast to the inspector layer.
func (b *ToastBuilder) InspectorLayer() *ToastBuilder {
	return b.Layer(rtui.LayerInspector)
}

// BuildInstance creates and returns the ToastInstance directly.
func (b *ToastBuilder) BuildInstance() *ToastInstance {
	return NewToastInstance(b.node.Props())
}


// Info creates an info toast.
func Info(message string) *ToastVNode {
	return NewToastBuilder(message).Info().Build().(*ToastVNode)
}

// Success creates a success toast.
func Success(message string) *ToastVNode {
	return NewToastBuilder(message).Success().Build().(*ToastVNode)
}

// Warning creates a warning toast.
func Warning(message string) *ToastVNode {
	return NewToastBuilder(message).Warning().Build().(*ToastVNode)
}

// Error creates an error toast.
func Error(message string) *ToastVNode {
	return NewToastBuilder(message).Error().Build().(*ToastVNode)
}
