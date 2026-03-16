package breadcrumb

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// Builder provides a fluent API for creating Breadcrumb VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Breadcrumb builder.
func NewBuilder() *Builder {
	return &Builder{node: New(nil)}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for portal anchoring and lookup.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Items replaces the breadcrumb items.
func (b *Builder) Items(items []Item) *Builder {
	b.node.SetItems(items)
	return b
}

// Labels replaces the breadcrumb items using plain labels.
func (b *Builder) Labels(labels ...string) *Builder {
	items := make([]Item, 0, len(labels))
	for _, label := range labels {
		items = append(items, Crumb(label))
	}
	b.node.SetItems(items)
	return b
}

// Item appends a breadcrumb item.
func (b *Builder) Item(item Item) *Builder {
	b.node.AddItem(item)
	return b
}

// Separator sets the separator text.
func (b *Builder) Separator(separator string) *Builder {
	b.node.SetSeparator(separator)
	return b
}

// MaxWidth sets the preferred maximum width.
func (b *Builder) MaxWidth(width int) *Builder {
	b.node.SetMaxWidth(width)
	return b
}

// Style sets the base breadcrumb style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// ItemStyle sets the style for non-current items.
func (b *Builder) ItemStyle(s style.Style) *Builder {
	b.node.SetItemStyle(s)
	return b
}

// CurrentStyle sets the style for the active item.
func (b *Builder) CurrentStyle(s style.Style) *Builder {
	b.node.SetCurrentStyle(s)
	return b
}

// SeparatorStyle sets the style for separator text.
func (b *Builder) SeparatorStyle(s style.Style) *Builder {
	b.node.SetSeparatorStyle(s)
	return b
}

// Build returns the constructed breadcrumb VNode.
func (b *Builder) Build() *VNode {
	return b.node
}

// Of creates a breadcrumb node from the provided items.
func Of(items []Item) *VNode {
	return New(items)
}
