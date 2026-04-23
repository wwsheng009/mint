package steps

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
)

// Builder provides a fluent API for creating Steps VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Steps builder.
func NewBuilder() *Builder {
	return &Builder{node: New(nil)}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for lookup.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// ComponentID sets the event component identifier.
func (b *Builder) ComponentID(id string) *Builder {
	b.node.SetComponentID(id)
	return b
}

// Items replaces the step items.
func (b *Builder) Items(items []Item) *Builder {
	b.node.SetItems(items)
	return b
}

// Titles replaces the steps using plain titles.
func (b *Builder) Titles(titles ...string) *Builder {
	items := make([]Item, 0, len(titles))
	for _, title := range titles {
		items = append(items, Step(title))
	}
	b.node.SetItems(items)
	return b
}

// Item appends a step item.
func (b *Builder) Item(item Item) *Builder {
	b.node.AddItem(item)
	return b
}

// Current sets the current active step index.
func (b *Builder) Current(index int) *Builder {
	b.node.SetCurrent(index)
	return b
}

// InitialCurrent sets the initial active step index in uncontrolled mode.
func (b *Builder) InitialCurrent(index int) *Builder {
	b.node.SetInitialCurrent(index)
	return b
}

// Direction sets the render direction.
func (b *Builder) Direction(direction Direction) *Builder {
	b.node.SetDirection(direction)
	return b
}

// Horizontal renders steps in a single line.
func (b *Builder) Horizontal() *Builder {
	b.node.SetDirection(DirectionHorizontal)
	return b
}

// Vertical renders steps as stacked lines.
func (b *Builder) Vertical() *Builder {
	b.node.SetDirection(DirectionVertical)
	return b
}

// Disabled toggles interactivity.
func (b *Builder) Disabled(disabled bool) *Builder {
	b.node.SetDisabled(disabled)
	return b
}

// ProgressDot toggles dot-style indicators.
func (b *Builder) ProgressDot(enabled bool) *Builder {
	b.node.SetProgressDot(enabled)
	return b
}

// Percent sets the current process step progress percentage.
func (b *Builder) Percent(percent int) *Builder {
	b.node.SetPercent(percent)
	return b
}

// Style sets the base style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// TitleStyle sets the title style.
func (b *Builder) TitleStyle(s style.Style) *Builder {
	b.node.SetTitleStyle(s)
	return b
}

// DescriptionStyle sets the description style.
func (b *Builder) DescriptionStyle(s style.Style) *Builder {
	b.node.SetDescriptionStyle(s)
	return b
}

// SeparatorStyle sets the connector/separator style.
func (b *Builder) SeparatorStyle(s style.Style) *Builder {
	b.node.SetSeparatorStyle(s)
	return b
}

// WaitStyle sets the wait-step style override.
func (b *Builder) WaitStyle(s style.Style) *Builder {
	b.node.SetWaitStyle(s)
	return b
}

// ProcessStyle sets the current-step style override.
func (b *Builder) ProcessStyle(s style.Style) *Builder {
	b.node.SetProcessStyle(s)
	return b
}

// FinishStyle sets the finished-step style override.
func (b *Builder) FinishStyle(s style.Style) *Builder {
	b.node.SetFinishStyle(s)
	return b
}

// ErrorStyle sets the error-step style override.
func (b *Builder) ErrorStyle(s style.Style) *Builder {
	b.node.SetErrorStyle(s)
	return b
}

// OnChange sets the intent emitted when current step changes.
func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.node.SetCurrentIntent(changeIntent)
	return b
}

// CurrentForField binds the current step index to a field.
func (b *Builder) CurrentForField(binding intent.FieldBinding) *Builder {
	b.node.SetCurrentIntentField(binding)
	return b
}

// Build returns the configured Steps VNode.
func (b *Builder) Build() *VNode {
	return b.node
}

// Of creates a steps node from items.
func Of(items []Item) *VNode {
	return New(items)
}
