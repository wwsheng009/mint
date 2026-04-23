package collapse

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Collapse VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Collapse builder.
func NewBuilder() *Builder {
	return &Builder{node: New(nil)}
}

// Key sets the diff key.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// ComponentID sets the local intent routing ID.
func (b *Builder) ComponentID(id string) *Builder {
	b.node.SetComponentID(id)
	return b
}

// Items replaces all panels.
func (b *Builder) Items(items []Item) *Builder {
	b.node.SetItems(items)
	return b
}

// Item appends a panel.
func (b *Builder) Item(item Item) *Builder {
	b.node.AddItem(item)
	return b
}

// Accordion toggles accordion mode.
func (b *Builder) Accordion(accordion bool) *Builder {
	b.node.SetAccordion(accordion)
	return b
}

// AccordionMode enables accordion mode.
func (b *Builder) AccordionMode() *Builder {
	return b.Accordion(true)
}

// Multiple enables multi-expand mode.
func (b *Builder) Multiple() *Builder {
	return b.Accordion(false)
}

// ActiveKeys sets controlled active keys.
func (b *Builder) ActiveKeys(keys ...string) *Builder {
	b.node.SetActiveKeys(keys)
	return b
}

// InitialActiveKeys sets uncontrolled initial active keys.
func (b *Builder) InitialActiveKeys(keys ...string) *Builder {
	b.node.SetInitialActiveKeys(keys)
	return b
}

// Disabled disables the whole component.
func (b *Builder) Disabled(disabled bool) *Builder {
	b.node.SetDisabled(disabled)
	return b
}

// Bordered toggles borders.
func (b *Builder) Bordered(bordered bool) *Builder {
	b.node.SetBordered(bordered)
	return b
}

// Ghost toggles ghost style.
func (b *Builder) Ghost(ghost bool) *Builder {
	b.node.SetGhost(ghost)
	return b
}

// Width sets panel width.
func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

// Style sets base style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// HeaderStyle sets header button style.
func (b *Builder) HeaderStyle(s style.Style) *Builder {
	b.node.SetHeaderStyle(s)
	return b
}

// ActiveHeaderStyle sets expanded header style.
func (b *Builder) ActiveHeaderStyle(s style.Style) *Builder {
	b.node.SetActiveHeaderStyle(s)
	return b
}

// ContentStyle sets content wrapper style.
func (b *Builder) ContentStyle(s style.Style) *Builder {
	b.node.SetContentStyle(s)
	return b
}

// OnChange sets the optional custom change intent.
func (b *Builder) OnChange(i intent.Intent) *Builder {
	b.node.SetChangeIntent(i)
	return b
}

// ActiveKeysForField binds active keys to a field.
func (b *Builder) ActiveKeysForField(binding intent.FieldBinding) *Builder {
	b.node.SetChangeIntentField(binding)
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildVNode returns the concrete VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.node
}

// Of creates a collapse vnode from items.
func Of(items []Item) *VNode {
	return New(items)
}
