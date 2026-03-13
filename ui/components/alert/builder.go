package alert

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Alert VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Alert builder with the given message.
func NewBuilder(message string) *Builder {
	b := &Builder{node: New()}
	b.node.SetMessage(message)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Title sets the alert title.
func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

// Message sets the alert message.
func (b *Builder) Message(msg string) *Builder {
	b.node.SetMessage(msg)
	return b
}

// Type sets the alert type.
func (b *Builder) Type(t AlertType) *Builder {
	b.node.SetAlertType(t)
	return b
}

// Info sets the alert type to info.
func (b *Builder) Info() *Builder {
	b.node.Info()
	return b
}

// Success sets the alert type to success.
func (b *Builder) Success() *Builder {
	b.node.Success()
	return b
}

// Warning sets the alert type to warning.
func (b *Builder) Warning() *Builder {
	b.node.Warning()
	return b
}

// Error sets the alert type to error.
func (b *Builder) Error() *Builder {
	b.node.Error()
	return b
}

// Closable sets whether the alert can be closed.
func (b *Builder) Closable(closable bool) *Builder {
	b.node.SetClosable(closable)
	return b
}

// CloseIntent sets the intent to emit when the alert is closed.
func (b *Builder) CloseIntent(ci interface{}) *Builder {
	b.node.SetCloseIntent(ci)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetAlertStyle(s)
	return b
}

// Build returns the constructed Alert VNode.
func (b *Builder) Build() *VNode {
	return b.node
}

// Alert is a shortcut constructor: alert.Alert("message").
func Alert(message string) *VNode {
	return New().SetMessage(message)
}
