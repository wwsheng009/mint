package notification

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Notification VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Notification builder with the given message.
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

// Title sets the notification title.
func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

// Message sets the notification message.
func (b *Builder) Message(msg string) *Builder {
	b.node.SetMessage(msg)
	return b
}

// Type sets the notification type.
func (b *Builder) Type(t NotificationType) *Builder {
	b.node.SetType(t)
	return b
}

// Info sets the notification type to info.
func (b *Builder) Info() *Builder {
	b.node.Info()
	return b
}

// Success sets the notification type to success.
func (b *Builder) Success() *Builder {
	b.node.Success()
	return b
}

// Warning sets the notification type to warning.
func (b *Builder) Warning() *Builder {
	b.node.Warning()
	return b
}

// Error sets the notification type to error.
func (b *Builder) Error() *Builder {
	b.node.Error()
	return b
}

// Closable sets whether the notification can be closed by the user.
func (b *Builder) Closable(closable bool) *Builder {
	b.node.SetClosable(closable)
	return b
}

// CloseIntent sets the intent emitted when the close action is triggered.
func (b *Builder) CloseIntent(intent interface{}) *Builder {
	b.node.SetCloseIntent(intent)
	return b
}

// Duration sets the auto-dismiss duration. 0 means persistent.
func (b *Builder) Duration(d time.Duration) *Builder {
	b.node.SetDuration(d)
	return b
}

// Placement sets where the notification appears.
func (b *Builder) Placement(p Placement) *Builder {
	b.node.SetPlacement(p)
	return b
}

// Style sets a custom base style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() *VNode {
	return b.node
}

// BuildInstance creates and returns a ready-to-use Instance.
func (b *Builder) BuildInstance() *Instance {
	inst := b.node.CreateInstance().(*Instance)
	inst.Show()
	return inst
}
