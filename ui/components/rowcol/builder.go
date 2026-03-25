package rowcol

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// RowBuilder provides a fluent API for creating Row VNodes.
type RowBuilder struct {
	node *RowVNode
}

// ColBuilder provides a fluent API for creating Col VNodes.
type ColBuilder struct {
	node *ColVNode
}

// NewRowBuilder creates a new Row builder.
func NewRowBuilder() *RowBuilder {
	return &RowBuilder{node: NewRow()}
}

// NewColBuilder creates a new Col builder.
func NewColBuilder() *ColBuilder {
	return &ColBuilder{node: NewCol()}
}

func (b *RowBuilder) Key(key string) *RowBuilder {
	b.node.SetKey(key)
	return b
}

func (b *RowBuilder) SetID(id string) *RowBuilder {
	b.node.SetID(id)
	return b
}

func (b *RowBuilder) Children(children ...rtui.VNode) *RowBuilder {
	b.node.SetChildrenList(children)
	return b
}

func (b *RowBuilder) AddChild(child rtui.VNode) *RowBuilder {
	b.node.AddChild(child)
	return b
}

func (b *RowBuilder) Justify(justify rtui.Align) *RowBuilder {
	b.node.SetJustify(justify)
	return b
}

func (b *RowBuilder) Align(align rtui.Align) *RowBuilder {
	b.node.SetAlign(align)
	return b
}

func (b *RowBuilder) Gutter(horizontal int, vertical ...int) *RowBuilder {
	b.node.SetGutter(horizontal)
	if len(vertical) > 0 {
		b.node.SetVerticalGutter(vertical[0])
	}
	return b
}

func (b *RowBuilder) Wrap(enabled bool) *RowBuilder {
	b.node.SetWrap(enabled)
	return b
}

func (b *RowBuilder) Width(width int) *RowBuilder {
	b.node.SetWidth(width)
	return b
}

func (b *RowBuilder) Style(s style.Style) *RowBuilder {
	b.node.SetStyleProps(s)
	return b
}

func (b *RowBuilder) Build() rtui.VNode {
	return b.node
}

func (b *RowBuilder) BuildVNode() *RowVNode {
	return b.node
}

func (b *ColBuilder) Key(key string) *ColBuilder {
	b.node.SetKey(key)
	return b
}

func (b *ColBuilder) SetID(id string) *ColBuilder {
	b.node.SetID(id)
	return b
}

func (b *ColBuilder) Span(span int) *ColBuilder {
	b.node.SetSpan(span)
	return b
}

func (b *ColBuilder) Offset(offset int) *ColBuilder {
	b.node.SetOffset(offset)
	return b
}

func (b *ColBuilder) Children(children ...rtui.VNode) *ColBuilder {
	b.node.SetChildrenList(children)
	return b
}

func (b *ColBuilder) AddChild(child rtui.VNode) *ColBuilder {
	b.node.AddChild(child)
	return b
}

func (b *ColBuilder) Style(s style.Style) *ColBuilder {
	b.node.SetStyleProps(s)
	return b
}

func (b *ColBuilder) Build() rtui.VNode {
	return b.node
}

func (b *ColBuilder) BuildVNode() *ColVNode {
	return b.node
}
