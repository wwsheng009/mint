package textarea

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for building Textarea VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Textarea builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Placeholder(text string) *Builder {
	b.node.SetPlaceholder(text)
	return b
}

func (b *Builder) Value(value string) *Builder {
	b.node.SetValue(value)
	return b
}

func (b *Builder) Rows(rows int) *Builder {
	b.node.SetRows(rows)
	return b
}

func (b *Builder) Cols(cols int) *Builder {
	b.node.SetCols(cols)
	return b
}

func (b *Builder) MaxLen(len int) *Builder {
	b.node.SetMaxLen(len)
	return b
}

func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) OnChange(i intent.Intent) *Builder {
	b.node.SetChangeIntent(i)
	return b
}

func (b *Builder) OnSubmit(i intent.Intent) *Builder {
	b.node.SetSubmitIntent(i)
	return b
}

// ForField binds the textarea to a state field using FieldBinding.
// When the textarea value changes, it emits a FieldChangeIntent with the current text.
// Example:
//
//	var message = intent.StateKey[string]("message")
//	textarea.NewBuilder().
//	    ForField(intent.ForField(message)).
//	    Build()
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	// Set the FieldIntent as the changeIntentField for MVP mode
	b.node.SetProps(rtui.Props{
		"changeIntent": binding, // binding implements both FieldIntent and Intent
	})
	return b
}

// ForForm binds the textarea to a form using FormBinding (Phase 6).
// When the textarea value changes, it emits a FormFieldChangeIntent.
// Example:
//
//	formBinding := form.ForForm("myForm")
//	var message = intent.StateKey[string]("message")
//	textarea.NewBuilder().
//	    ForField(intent.ForField(message)).
//	    ForForm(formBinding).
//	    Build()
func (b *Builder) ForForm(binding intent.FormBinding) *Builder {
	b.node.SetFormID(binding.GetFormID())
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// Textarea creates a new Textarea VNode (for backward compatibility).
// This matches the old form.Textarea() API.
func Textarea() *VNode {
	return New()
}

// NewTextarea creates a new Textarea VNode (alias for New, for backward compatibility).
// This matches the old form.NewTextarea() API.
func NewTextarea() *VNode {
	return New()
}

// TextareaBuilder is an alias for Builder (for backward compatibility).
// This matches the old form.TextareaBuilder type.
type TextareaBuilder = Builder
