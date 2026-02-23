package tooltip

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Tooltip Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Tooltip VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Tooltip builder.
func NewBuilder(content rtui.VNode, text string) *Builder {
	return &Builder{
		node: New(content, text),
	}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Text sets the tooltip text.
func (b *Builder) Text(text string) *Builder {
	b.node.SetText(text)
	return b
}

// Position sets the tooltip position.
func (b *Builder) Position(position Position) *Builder {
	b.node.SetPosition(position)
	return b
}

// Top sets position to top.
func (b *Builder) Top() *Builder {
	return b.Position(PositionTop)
}

// Bottom sets position to bottom.
func (b *Builder) Bottom() *Builder {
	return b.Position(PositionBottom)
}

// Left sets position to left.
func (b *Builder) Left() *Builder {
	return b.Position(PositionLeft)
}

// Right sets position to right.
func (b *Builder) Right() *Builder {
	return b.Position(PositionRight)
}

// Auto sets position to auto.
func (b *Builder) Auto() *Builder {
	return b.Position(PositionAuto)
}

// Delay sets the delay before showing the tooltip.
func (b *Builder) Delay(delay time.Duration) *Builder {
	b.node.SetDelay(delay)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyleProps(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
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
func (b *Builder) BgColor(c interface{}) *Builder {
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

// Build returns the VNode as rtui.VNode interface.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// Layer sets the rendering layer for the tooltip.
func (b *Builder) Layer(l rtui.Layer) *Builder {
	b.node.SetLayer(l)
	return b
}

// SetRenderLayer is an alias for Layer for backward compatibility.
func (b *Builder) SetRenderLayer(l rtui.Layer) *Builder {
	return b.Layer(l)
}

// BaseLayer sets the tooltip to the base layer.
func (b *Builder) BaseLayer() *Builder {
	return b.Layer(rtui.LayerBase)
}

// OverlayLayer sets the tooltip to the overlay layer.
func (b *Builder) OverlayLayer() *Builder {
	return b.Layer(rtui.LayerOverlay)
}

// ModalLayer sets the tooltip to the modal layer.
func (b *Builder) ModalLayer() *Builder {
	return b.Layer(rtui.LayerModal)
}

// TooltipLayer sets the tooltip to the tooltip layer (default).
func (b *Builder) TooltipLayer() *Builder {
	return b.Layer(rtui.LayerTooltip)
}

// InspectorLayer sets the tooltip to the inspector layer.
func (b *Builder) InspectorLayer() *Builder {
	return b.Layer(rtui.LayerInspector)
}

// BuildInstance creates and returns the Instance directly.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.node.Props())
}

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

// =============================================================================
// Convenience Functions
// =============================================================================

// T creates a new Tooltip builder.
func T(content rtui.VNode, text string) *Builder {
	return NewBuilder(content, text)
}

// Tooltip creates a new Tooltip VNode directly.
func Tooltip(content rtui.VNode, text string) *VNode {
	return New(content, text)
}

// Note: The Toast() convenience function is defined in vnode.go to avoid circular imports.
// The following functions use the builder pattern for additional type-specific conveniences.

// Info creates an info toast using the builder.
func Info(message string) *ToastVNode {
	return NewToastBuilder(message).Info().Build().(*ToastVNode)
}

// Success creates a success toast using the builder.
func Success(message string) *ToastVNode {
	return NewToastBuilder(message).Success().Build().(*ToastVNode)
}

// Warning creates a warning toast using the builder.
func Warning(message string) *ToastVNode {
	return NewToastBuilder(message).Warning().Build().(*ToastVNode)
}

// Error creates an error toast using the builder.
func Error(message string) *ToastVNode {
	return NewToastBuilder(message).Error().Build().(*ToastVNode)
}
