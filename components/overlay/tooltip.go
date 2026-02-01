package overlay

// =============================================================================
// Tooltip Component
// =============================================================================

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Tooltip Component
// =============================================================================

// TooltipPosition defines where the tooltip appears
type TooltipPosition int

const (
	TooltipTop TooltipPosition = iota
	TooltipBottom
	TooltipLeft
	TooltipRight
	TooltipAuto
)

// TooltipVNode represents a tooltip component
type TooltipVNode struct {
	*ui.ElementVNode
	content  ui.VNode
	text     string
	position TooltipPosition
	delay    time.Duration
	visible  bool
}

// NewTooltip creates a new tooltip
func NewTooltip() *TooltipVNode {
	return &TooltipVNode{
		ElementVNode: ui.NewElement("tooltip"),
		content:      nil,
		text:         "",
		position:     TooltipAuto,
		delay:        500 * time.Millisecond,
		visible:      false,
	}
}

// Tooltip creates a tooltip wrapping content
func Tooltip(content ui.VNode, text string) ui.VNode {
	return &TooltipVNode{
		ElementVNode: ui.NewElement("tooltip"),
		content:      content,
		text:         text,
		position:     TooltipAuto,
		delay:        500 * time.Millisecond,
		visible:      false,
	}
}

// TooltipBuilder creates a tooltip builder
func TooltipBuilder(content ui.VNode) *TooltipBuilderType {
	return &TooltipBuilderType{
		node: &TooltipVNode{
			ElementVNode: ui.NewElement("tooltip"),
			content:      content,
			text:         "",
			position:     TooltipAuto,
			delay:        500 * time.Millisecond,
			visible:      false,
		},
	}
}

// TooltipBuilderType provides fluent API for tooltips
type TooltipBuilderType struct {
	node *TooltipVNode
}

// Text sets the tooltip text
func (b *TooltipBuilderType) Text(text string) *TooltipBuilderType {
	b.node.text = text
	return b
}

// Position sets the tooltip position
func (b *TooltipBuilderType) Position(pos TooltipPosition) *TooltipBuilderType {
	b.node.position = pos
	return b
}

// Delay sets the show delay
func (b *TooltipBuilderType) Delay(d time.Duration) *TooltipBuilderType {
	b.node.delay = d
	return b
}

// Visible sets initial visibility
func (b *TooltipBuilderType) Visible(v bool) *TooltipBuilderType {
	b.node.visible = v
	return b
}

// Style sets the visual style
func (b *TooltipBuilderType) Style(s style.Style) *TooltipBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *TooltipBuilderType) FgColor(c interface{}) *TooltipBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *TooltipBuilderType) BgColor(c interface{}) *TooltipBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Key sets the key for diffing
func (b *TooltipBuilderType) Key(key string) *TooltipBuilderType {
	b.node.SetKey(key)
	return b
}

// Build returns the tooltip ui.VNode
func (b *TooltipBuilderType) Build() ui.VNode {
	return b.node
}

// Getters
func (t *TooltipVNode) Content() ui.VNode        { return t.content }
func (t *TooltipVNode) Text() string             { return t.text }
func (t *TooltipVNode) Position() TooltipPosition { return t.position }
func (t *TooltipVNode) Delay() time.Duration     { return t.delay }
func (t *TooltipVNode) IsVisible() bool           { return t.visible }

// =============================================================================
// Toast Component
// =============================================================================

// ToastType defines the type of toast notification
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastVNode represents a toast notification
type ToastVNode struct {
	*ui.ElementVNode
	title      string
	message    string
	toastType  ToastType
	duration   time.Duration
	visible    bool
	onClose    func()
}

// NewToast creates a new toast
func NewToast() *ToastVNode {
	return &ToastVNode{
		ElementVNode: ui.NewElement("toast"),
		title:        "",
		message:      "",
		toastType:    ToastInfo,
		duration:     3000 * time.Millisecond,
		visible:      false,
		onClose:      nil,
	}
}

// Toast creates a toast notification node
func Toast(message string) ui.VNode {
	return &ToastVNode{
		ElementVNode: ui.NewElement("toast"),
		title:        "",
		message:      message,
		toastType:    ToastInfo,
		duration:     3000 * time.Millisecond,
		visible:      true,
	}
}

// ToastBuilder creates a toast builder
func ToastBuilder() *ToastBuilderType {
	return &ToastBuilderType{node: NewToast()}
}

// ToastBuilderType provides fluent API for toasts
type ToastBuilderType struct {
	node *ToastVNode
}

// Title sets the toast title
func (b *ToastBuilderType) Title(title string) *ToastBuilderType {
	b.node.title = title
	return b
}

// Message sets the toast message
func (b *ToastBuilderType) Message(msg string) *ToastBuilderType {
	b.node.message = msg
	return b
}

// Type sets the toast type
func (b *ToastBuilderType) Type(t ToastType) *ToastBuilderType {
	b.node.toastType = t
	return b
}

// Info sets toast type to info
func (b *ToastBuilderType) Info() *ToastBuilderType {
	b.node.toastType = ToastInfo
	return b
}

// Success sets toast type to success
func (b *ToastBuilderType) Success() *ToastBuilderType {
	b.node.toastType = ToastSuccess
	return b
}

// Warning sets toast type to warning
func (b *ToastBuilderType) Warning() *ToastBuilderType {
	b.node.toastType = ToastWarning
	return b
}

// Error sets toast type to error
func (b *ToastBuilderType) Error() *ToastBuilderType {
	b.node.toastType = ToastError
	return b
}

// Duration sets how long the toast is visible
func (b *ToastBuilderType) Duration(d time.Duration) *ToastBuilderType {
	b.node.duration = d
	return b
}

// Visible sets the visibility
func (b *ToastBuilderType) Visible(v bool) *ToastBuilderType {
	b.node.visible = v
	return b
}

// OnClose sets the close callback
func (b *ToastBuilderType) OnClose(fn func()) *ToastBuilderType {
	b.node.onClose = fn
	return b
}

// Style sets the visual style
func (b *ToastBuilderType) Style(s style.Style) *ToastBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *ToastBuilderType) FgColor(c interface{}) *ToastBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *ToastBuilderType) BgColor(c interface{}) *ToastBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Key sets the key for diffing
func (b *ToastBuilderType) Key(key string) *ToastBuilderType {
	b.node.SetKey(key)
	return b
}

// Build returns the toast ui.VNode
func (b *ToastBuilderType) Build() ui.VNode {
	return b.node
}

// Getters
func (t *ToastVNode) Title() string         { return t.title }
func (t *ToastVNode) Message() string       { return t.message }
func (t *ToastVNode) ToastType() ToastType  { return t.toastType }
func (t *ToastVNode) Duration() time.Duration { return t.duration }
func (t *ToastVNode) IsVisible() bool       { return t.visible }
func (t *ToastVNode) OnClose() func()       { return t.onClose }

// =============================================================================
// Toast Manager for handling multiple toasts
// =============================================================================

// ToastManager manages toast notifications
type ToastManager struct {
	toasts []ui.VNode
}

// NewToastManager creates a new toast manager
func NewToastManager() *ToastManager {
	return &ToastManager{
		toasts: make([]ui.VNode, 0),
	}
}

// Show adds a toast to be displayed
func (tm *ToastManager) Show(toast ui.VNode) {
	if t, ok := toast.(*ToastVNode); ok {
		t.visible = true
	}
	tm.toasts = append(tm.toasts, toast)
}

// Info shows an info toast
func (tm *ToastManager) Info(message string) {
	tm.Show(ToastBuilder().Message(message).Info().Visible(true).Build())
}

// Success shows a success toast
func (tm *ToastManager) Success(message string) {
	tm.Show(ToastBuilder().Message(message).Success().Visible(true).Build())
}

// Warning shows a warning toast
func (tm *ToastManager) Warning(message string) {
	tm.Show(ToastBuilder().Message(message).Warning().Visible(true).Build())
}

// Error shows an error toast
func (tm *ToastManager) Error(message string) {
	tm.Show(ToastBuilder().Message(message).Error().Visible(true).Build())
}

// Remove removes a toast
func (tm *ToastManager) Remove(toast ui.VNode) {
	for i, t := range tm.toasts {
		if t == toast {
			tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
			break
		}
	}
}

// Clear removes all toasts
func (tm *ToastManager) Clear() {
	tm.toasts = tm.toasts[:0]
}

// GetToasts returns all active toasts
func (tm *ToastManager) GetToasts() []ui.VNode {
	return tm.toasts
}
