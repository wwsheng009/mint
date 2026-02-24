package table

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing Table VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new Table builder.
func NewBuilder() *Builder {
	return &Builder{
		vnode: New(),
	}
}

// Key sets the component key.
func (b *Builder) Key(key string) *Builder {
	b.vnode.SetKey(key)
	return b
}

// Columns sets the columns.
func (b *Builder) Columns(cols []TableColumn) *Builder {
	b.vnode.SetColumns(cols)
	return b
}

// Rows sets the rows.
func (b *Builder) Rows(rows [][]string) *Builder {
	b.vnode.SetRows(rows)
	return b
}

// AddRow adds a single row.
func (b *Builder) AddRow(row ...string) *Builder {
	b.vnode.AddRow(row...)
	return b
}

// HeaderStyle sets the header style.
func (b *Builder) HeaderStyle(s style.Style) *Builder {
	b.vnode.SetHeaderStyle(s)
	return b
}

// TableStyle sets the table style.
func (b *Builder) TableStyle(s style.Style) *Builder {
	b.vnode.SetStyle(s)
	return b
}

// Gap sets the gap between header and rows.
func (b *Builder) Gap(gap int) *Builder {
	b.vnode.SetGap(gap)
	return b
}

// Style sets the visual style (shortcut for TableStyle).
func (b *Builder) Style(s style.Style) *Builder {
	b.vnode.SetStyle(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c style.Color) *Builder {
	s := b.vnode.Style()
	s.FG = c
	b.vnode.SetStyle(s)
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c style.Color) *Builder {
	s := b.vnode.Style()
	s.BG = c
	b.vnode.SetStyle(s)
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.vnode
}

// BuildVNode returns the configured *VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.vnode
}

// BuildInstance directly creates an Instance.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.vnode.Props())
}

// =============================================================================
// Convenience Functions
// =============================================================================

// Of creates a Table with the given columns and rows.
func Of(cols []TableColumn, rows [][]string) rtui.VNode {
	return NewBuilder().Columns(cols).Rows(rows).Build()
}

// OfColumns creates a Table with the given columns.
func OfColumns(cols []TableColumn) rtui.VNode {
	return NewBuilder().Columns(cols).Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// Table creates a new Table builder.
func Table() *Builder {
	return NewBuilder()
}

// WithHeaders creates a Table builder with columns.
func WithHeaders(cols []TableColumn) *Builder {
	return NewBuilder().Columns(cols)
}
