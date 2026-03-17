package badge

import "github.com/wwsheng009/mint/runtime/style"

// Builder provides a fluent API for creating Badge VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Badge builder.
func NewBuilder(label string) *Builder {
	return &Builder{node: New(label)}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Label sets the anchor/label text.
func (b *Builder) Label(label string) *Builder {
	b.node.SetLabel(label)
	return b
}

// Count sets the numeric badge count.
func (b *Builder) Count(count int) *Builder {
	b.node.SetCount(count)
	return b
}

// Text sets a custom badge text.
func (b *Builder) Text(text string) *Builder {
	b.node.SetText(text)
	return b
}

// Dot toggles dot mode.
func (b *Builder) Dot(dot bool) *Builder {
	b.node.SetDot(dot)
	return b
}

// ShowZero controls whether zero count remains visible.
func (b *Builder) ShowZero(show bool) *Builder {
	b.node.SetShowZero(show)
	return b
}

// OverflowCount sets the max visible numeric count before `+`.
func (b *Builder) OverflowCount(max int) *Builder {
	b.node.SetOverflowCount(max)
	return b
}

// Status sets the badge status variant.
func (b *Builder) Status(status Status) *Builder {
	b.node.SetStatus(status)
	return b
}

// Style sets the base style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// LabelStyle sets the label style.
func (b *Builder) LabelStyle(s style.Style) *Builder {
	b.node.SetLabelStyle(s)
	return b
}

// BadgeStyle sets the badge style override.
func (b *Builder) BadgeStyle(s style.Style) *Builder {
	b.node.SetBadgeStyle(s)
	return b
}

// Default sets default badge styling.
func (b *Builder) Default() *Builder {
	b.node.Default()
	return b
}

// Primary sets primary badge styling.
func (b *Builder) Primary() *Builder {
	b.node.Primary()
	return b
}

// Success sets success badge styling.
func (b *Builder) Success() *Builder {
	b.node.Success()
	return b
}

// Warning sets warning badge styling.
func (b *Builder) Warning() *Builder {
	b.node.Warning()
	return b
}

// Error sets error badge styling.
func (b *Builder) Error() *Builder {
	b.node.Error()
	return b
}

// Processing sets processing badge styling.
func (b *Builder) Processing() *Builder {
	b.node.Processing()
	return b
}

// Build returns the configured Badge VNode.
func (b *Builder) Build() *VNode {
	return b.node
}

// Badge creates a Badge node for the given label.
func Badge(label string) *VNode {
	return New(label)
}
