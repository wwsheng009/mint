package button

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Button VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Button builder.
func NewBuilder(label string) *Builder {
	return &Builder{
		node: New(label),
	}
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

// Label sets the button label.
func (b *Builder) Label(label string) *Builder {
	b.node.SetLabel(label)
	return b
}

// Variant sets the button variant.
func (b *Builder) Variant(v Variant) *Builder {
	b.node.SetVariant(v)
	return b
}

// Primary sets the variant to Primary.
func (b *Builder) Primary() *Builder {
	return b.Variant(VariantPrimary)
}

// Secondary sets the variant to Secondary.
func (b *Builder) Secondary() *Builder {
	return b.Variant(VariantSecondary)
}

// Danger sets the variant to Danger.
func (b *Builder) Danger() *Builder {
	return b.Variant(VariantDanger)
}

// Success sets the variant to Success.
func (b *Builder) Success() *Builder {
	return b.Variant(VariantSuccess)
}

// Size sets the button size.
func (b *Builder) Size(s Size) *Builder {
	b.node.SetSize(s)
	return b
}

// Small sets the size to Small.
func (b *Builder) Small() *Builder {
	return b.Size(SizeSmall)
}

// Medium sets the size to Medium.
func (b *Builder) Medium() *Builder {
	return b.Size(SizeMedium)
}

// Large sets the size to Large.
func (b *Builder) Large() *Builder {
	return b.Size(SizeLarge)
}

// FocusStyle sets the focus style.
func (b *Builder) FocusStyle(fs FocusStyle) *Builder {
	b.node.SetFocusStyle(fs)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(disabled bool) *Builder {
	b.node.SetDisabled(disabled)
	return b
}

// OnPress sets the intent to emit when pressed.
// This is the Fiber-first way to handle button press events.
//
// Example:
//
//	ButtonBuilder("Open").OnPress(intent.NewAction("open_modal", "settings"))
func (b *Builder) OnPress(pressIntent intent.Intent) *Builder {
	b.node.SetIntent(pressIntent)
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

// Padding sets the padding (top, right, bottom, left).
func (b *Builder) Padding(top, right, bottom, left int) *Builder {
	b.node.SetPaddingProps(top, right, bottom, left)
	return b
}

// PaddingH sets horizontal padding (left, right).
func (b *Builder) PaddingH(left, right int) *Builder {
	b.node.SetPaddingProps(0, right, 0, left)
	return b
}

// PaddingV sets vertical padding (top, bottom).
func (b *Builder) PaddingV(top, bottom int) *Builder {
	b.node.SetPaddingProps(top, 0, bottom, 0)
	return b
}

// PaddingAll sets same padding on all sides.
func (b *Builder) PaddingAll(p int) *Builder {
	b.node.SetPaddingProps(p, p, p, p)
	return b
}

// TextAlign sets the text alignment.
func (b *Builder) TextAlign(align rtui.Align) *Builder {
	b.node.SetTextAlignProps(align)
	return b
}

// Flex sets the flex factor for this button in a layout.
// Usage: .Flex(1) to make button take equal space.
func (b *Builder) Flex(factor int) *Builder {
	b.node.flex = factor
	return b
}

// Margin sets the margin (top, right, bottom, left).
func (b *Builder) Margin(top, right, bottom, left int) *Builder {
	b.node.BoxModelMixin.SetMargin(top, right, bottom, left)
	return b
}

// MarginH sets horizontal margin (left, right).
func (b *Builder) MarginH(left, right int) *Builder {
	b.node.BoxModelMixin.SetMargin(0, right, 0, left)
	return b
}

// MarginV sets vertical margin (top, bottom).
func (b *Builder) MarginV(top, bottom int) *Builder {
	b.node.BoxModelMixin.SetMargin(top, 0, bottom, 0)
	return b
}

// MarginAll sets same margin on all sides.
func (b *Builder) MarginAll(m int) *Builder {
	b.node.BoxModelMixin.SetMargin(m, m, m, m)
	return b
}

// Build returns the VNode as rtui.VNode interface.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildInstance creates and returns the Instance directly.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.node.Props())
}

// =============================================================================
// Convenience Functions
// =============================================================================

// B creates a new Button builder.
func B(label string) *Builder {
	return NewBuilder(label)
}

// Button creates a new Button VNode directly.
func Button(label string) *VNode {
	return New(label)
}

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// NewButton creates a new Button VNode (alias for New, for backward compatibility).
// This matches the old button.NewButton() API.
func NewButton(label string) *VNode {
	return New(label)
}

// ButtonBuilder is an alias for Builder (for backward compatibility).
// This matches the old button.ButtonBuilder type.
type ButtonBuilder = Builder
