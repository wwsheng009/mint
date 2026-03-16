package input

import (
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/cursor"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for building Input VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Input builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(),
	}
}

// Placeholder sets the placeholder text.
func (b *Builder) Placeholder(text string) *Builder {
	b.node.SetPlaceholder(text)
	return b
}

// Value sets the initial value.
func (b *Builder) Value(value string) *Builder {
	b.node.SetValue(value)
	return b
}

// Key sets the key for diffing.
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

// Type sets the input type.
func (b *Builder) Type(t Type) *Builder {
	b.node.SetType(t)
	return b
}

// AllowNegative configures whether TypeNumber accepts a leading minus sign.
func (b *Builder) AllowNegative(v bool) *Builder {
	b.node.SetAllowNegative(v)
	return b
}

// AllowDecimal configures whether TypeNumber accepts a decimal point.
func (b *Builder) AllowDecimal(v bool) *Builder {
	b.node.SetAllowDecimal(v)
	return b
}

// Prefix sets the inner prefix text rendered before the editable value.
func (b *Builder) Prefix(text string) *Builder {
	b.node.SetPrefix(text)
	return b
}

// Suffix sets the inner suffix text rendered after the editable value.
func (b *Builder) Suffix(text string) *Builder {
	b.node.SetSuffix(text)
	return b
}

// AddonBefore sets the outer leading addon text.
func (b *Builder) AddonBefore(text string) *Builder {
	b.node.SetAddonBefore(text)
	return b
}

// AddonAfter sets the outer trailing addon text.
func (b *Builder) AddonAfter(text string) *Builder {
	b.node.SetAddonAfter(text)
	return b
}

// Search enables the search input variant.
func (b *Builder) Search() *Builder {
	b.node.SetSearchVariant(true)
	return b
}

// Password sets the input type to password.
func (b *Builder) Password() *Builder {
	b.node.SetPassword()
	return b
}

// MaxLen sets the maximum length.
func (b *Builder) MaxLen(len int) *Builder {
	b.node.SetMaxLen(len)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

// ReadOnly sets the read-only state.
func (b *Builder) ReadOnly(v bool) *Builder {
	b.node.SetReadOnly(v)
	return b
}

// CursorConfig sets cursor blink/shape config for the embedded caret.
func (b *Builder) CursorConfig(cfg cursor.Config) *Builder {
	b.node.SetCursorConfig(cfg)
	return b
}

// CursorShape sets embedded caret shape.
func (b *Builder) CursorShape(shape cursor.Shape) *Builder {
	b.node.SetCursorShape(shape)
	return b
}

// InsertCursor sets a thin insertion caret.
func (b *Builder) InsertCursor() *Builder {
	b.node.SetInsertCursor()
	return b
}

// BlockCursor sets a block caret.
func (b *Builder) BlockCursor() *Builder {
	b.node.SetBlockCursor()
	return b
}

// UnderlineCursor sets an underline caret.
func (b *Builder) UnderlineCursor() *Builder {
	b.node.SetUnderlineCursor()
	return b
}

// CursorBlink enables/disables caret blink.
func (b *Builder) CursorBlink(enabled bool) *Builder {
	b.node.SetCursorBlink(enabled)
	return b
}

// CursorBlinkInterval sets caret blink interval.
func (b *Builder) CursorBlinkInterval(interval time.Duration) *Builder {
	b.node.SetCursorBlinkInterval(interval)
	return b
}

// Width sets the explicit width.
func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Foreground(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Foreground(color)
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Background(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Background(color)
		b.node.SetStyle(s)
	}
	return b
}

// OnChange sets the change intent.
func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.node.SetChangeIntent(changeIntent)
	return b
}

// ForField sets the field binding for MVP data flow.
// This creates a FieldBinding intent that will be used by the Instance
// to emit FieldChangeIntent when the user types.
//
// Example:
//
//	input.NewBuilder().ForField(intent.BindField("username"))
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.node.SetChangeIntent(binding)
	return b
}

// ForForm sets the form ID for Form integration (Phase 6).
// When combined with ForField(), the component will emit
// FormFieldChangeIntent/FormFieldBlurIntent instead of FieldChangeIntent.
//
// Example:
//
//	input.NewBuilder().
//	    ForField(intent.BindField("username")).
//	    ForForm(intent.BindForm("loginForm"))
func (b *Builder) ForForm(binding intent.FormBinding) *Builder {
	b.node.SetFormID(binding.GetFormID())
	return b
}

// OnSubmit sets the submit intent.
func (b *Builder) OnSubmit(submitIntent intent.Intent) *Builder {
	b.node.SetSubmitIntent(submitIntent)
	return b
}

// OnSearch is an alias of OnSubmit for the Search variant.
func (b *Builder) OnSearch(searchIntent intent.Intent) *Builder {
	b.node.SetSubmitIntent(searchIntent)
	return b
}

// Build returns the VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildTyped returns the typed VNode.
func (b *Builder) BuildTyped() *VNode {
	return b.node
}

// BuildInstance returns the runtime instance.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.node.Props())
}

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// Input creates a new Input VNode (for backward compatibility).
// This matches the old form.Input() API.
func Input() *VNode {
	return New()
}

// NewInput creates a new Input VNode (alias for New, for backward compatibility).
// This matches the old form.NewInput() API.
func NewInput() *VNode {
	return New()
}

// SearchInput creates a new Search input builder.
func SearchInput() *Builder {
	return NewBuilder().Search()
}

// InputBuilder is an alias for Builder (for backward compatibility).
// This matches the old form.InputBuilder type.
type InputBuilder = Builder
