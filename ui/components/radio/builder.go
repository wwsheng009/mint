package radio

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for building Radio VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Radio builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(""),
	}
}

// Label sets the label text.
func (b *Builder) Label(text string) *Builder {
	b.node.SetLabel(text)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

// Checked sets the checked state.
func (b *Builder) Checked(v bool) *Builder {
	b.node.SetChecked(v)
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

// Bold sets the bold attribute.
func (b *Builder) Bold(v bool) *Builder {
	s := b.node.Style()
	s = s.Bold(v)
	b.node.SetStyle(s)
	return b
}

// OnSelect sets the select intent.
func (b *Builder) OnSelect(selectIntent intent.Intent) *Builder {
	b.node.SetIntent(selectIntent)
	return b
}

// ForField sets the field binding for MVP data flow.
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.node.SetIntent(binding)
	return b
}

// ForForm sets the form ID for Form integration.
func (b *Builder) ForForm(binding intent.FormBinding) *Builder {
	b.node.SetFormID(binding.GetFormID())
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

// GroupBuilder provides a fluent API for building RadioGroup VNodes.
type GroupBuilder struct {
	node *GroupVNode
}

// NewGroupBuilder creates a new RadioGroup builder.
func NewGroupBuilder(options []Option) *GroupBuilder {
	return &GroupBuilder{
		node: NewGroup(options),
	}
}

// Label sets the component label text.
func (b *GroupBuilder) Label(text string) *GroupBuilder {
	b.node.SetLabel(text)
	return b
}

// Key sets the key for diffing.
func (b *GroupBuilder) Key(key string) *GroupBuilder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
func (b *GroupBuilder) SetID(id string) *GroupBuilder {
	b.node.SetID(id)
	return b
}

// Disabled sets the disabled state.
func (b *GroupBuilder) Disabled(v bool) *GroupBuilder {
	b.node.SetDisabled(v)
	return b
}

// Options replaces the group options.
func (b *GroupBuilder) Options(options []Option) *GroupBuilder {
	b.node.SetOptions(options)
	return b
}

// Selected sets the selected value.
func (b *GroupBuilder) Selected(selected string) *GroupBuilder {
	b.node.SetSelected(selected)
	return b
}

// Style sets the visual style.
func (b *GroupBuilder) Style(s style.Style) *GroupBuilder {
	b.node.SetStyleProps(s)
	return b
}

// FgColor sets the foreground color.
func (b *GroupBuilder) FgColor(c interface{}) *GroupBuilder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Foreground(style.Color(colorStr))
		b.node.SetStyleProps(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Foreground(color)
		b.node.SetStyleProps(s)
	}
	return b
}

// BgColor sets the background color.
func (b *GroupBuilder) BgColor(c interface{}) *GroupBuilder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Background(style.Color(colorStr))
		b.node.SetStyleProps(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Background(color)
		b.node.SetStyleProps(s)
	}
	return b
}

// Bold sets the bold attribute.
func (b *GroupBuilder) Bold(v bool) *GroupBuilder {
	s := b.node.Style()
	s = s.Bold(v)
	b.node.SetStyleProps(s)
	return b
}

// Orientation sets the layout orientation.
func (b *GroupBuilder) Orientation(orientation Orientation) *GroupBuilder {
	b.node.SetOrientation(orientation)
	return b
}

// Spacing sets the gap between options.
func (b *GroupBuilder) Spacing(spacing int) *GroupBuilder {
	b.node.SetSpacing(spacing)
	return b
}

// OnSelect sets the select intent.
func (b *GroupBuilder) OnSelect(selectIntent intent.Intent) *GroupBuilder {
	b.node.SetIntent(selectIntent)
	return b
}

// ForField binds the selected value to FieldChangeIntent.
func (b *GroupBuilder) ForField(binding intent.FieldBinding) *GroupBuilder {
	b.node.SetIntent(binding)
	return b
}

// Vertical sets orientation to vertical.
func (b *GroupBuilder) Vertical() *GroupBuilder {
	b.node.Vertical()
	return b
}

// Horizontal sets orientation to horizontal.
func (b *GroupBuilder) Horizontal() *GroupBuilder {
	b.node.Horizontal()
	return b
}

// Build returns the VNode.
func (b *GroupBuilder) Build() rtui.VNode {
	return b.node
}

// BuildTyped returns the typed VNode.
func (b *GroupBuilder) BuildTyped() *GroupVNode {
	return b.node
}

// BuildInstance returns the runtime instance.
func (b *GroupBuilder) BuildInstance() *GroupInstance {
	return b.node.CreateInstance().(*GroupInstance)
}

// Radio creates a new Radio VNode.
func Radio(label string) *VNode {
	return New(label)
}

// NewRadio creates a new Radio VNode.
func NewRadio(label string) *VNode {
	return New(label)
}

// RadioGroup creates a new RadioGroup builder.
func RadioGroup(options []Option) *GroupBuilder {
	return NewGroupBuilder(options)
}
