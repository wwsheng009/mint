package list

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing List VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new List builder.
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

// Header sets the column header text.
func (b *Builder) Header(h string) *Builder {
	b.vnode.SetHeader(h)
	return b
}

// Rows sets all data rows at once.
func (b *Builder) Rows(rows []string) *Builder {
	b.vnode.SetRows(rows)
	return b
}

// AddRow appends a single row.
func (b *Builder) AddRow(row string) *Builder {
	b.vnode.AddRow(row)
	return b
}

// AddRows appends multiple rows.
func (b *Builder) AddRows(rows ...string) *Builder {
	b.vnode.AddRows(rows...)
	return b
}

// EmptyText sets the text displayed when there are no rows.
func (b *Builder) EmptyText(text string) *Builder {
	b.vnode.SetEmptyText(text)
	return b
}

// MaxRows limits the number of visible rows (0 = show all).
func (b *Builder) MaxRows(n int) *Builder {
	b.vnode.SetMaxRows(n)
	return b
}

// ShowBorder controls whether a border is drawn around the list.
func (b *Builder) ShowBorder(show bool) *Builder {
	b.vnode.SetShowBorder(show)
	return b
}

// ShowSeparator controls whether a separator line is drawn between header and rows.
func (b *Builder) ShowSeparator(show bool) *Builder {
	b.vnode.SetShowSeparator(show)
	return b
}

// SepChar sets the separator character (default '─').
func (b *Builder) SepChar(ch rune) *Builder {
	b.vnode.SetSeparatorChar(ch)
	return b
}

// HeaderStyle sets the style for the header row.
func (b *Builder) HeaderStyle(s style.Style) *Builder {
	b.vnode.SetHeaderStyle(s)
	return b
}

// RowStyle sets the default style for data rows.
func (b *Builder) RowStyle(s style.Style) *Builder {
	b.vnode.SetRowStyle(s)
	return b
}

// SelectedStyle sets the style for the selected row.
func (b *Builder) SelectedStyle(s style.Style) *Builder {
	b.vnode.SetSelectedStyle(s)
	return b
}

// BorderStyle sets the style for the border.
func (b *Builder) BorderStyle(s style.Style) *Builder {
	b.vnode.SetBorderStyle(s)
	return b
}

// ScrollOffset sets the initial scroll offset.
func (b *Builder) ScrollOffset(offset int) *Builder {
	b.vnode.SetScrollOffset(offset)
	return b
}

// SelectedIndex sets the currently selected row index.
func (b *Builder) SelectedIndex(index int) *Builder {
	b.vnode.SetSelectedIndex(index)
	return b
}

// ViewportHeight sets the visible height for scrolling.
func (b *Builder) ViewportHeight(height int) *Builder {
	b.vnode.SetViewportHeight(height)
	return b
}

// AllowScroll enables/disables scrolling.
func (b *Builder) AllowScroll(allow bool) *Builder {
	b.vnode.SetAllowScroll(allow)
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

// Of creates a List with the given rows.
func Of(rows []string) rtui.VNode {
	return NewBuilder().Rows(rows).Build()
}

// OfRows creates a List with multiple rows.
func OfRows(header string, rows ...string) rtui.VNode {
	return NewBuilder().Header(header).Rows(rows).Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// List creates a new List builder.
func List() *Builder {
	return NewBuilder()
}

// WithRows creates a builder with pre-configured rows.
func WithRows(rows []string) *Builder {
	return NewBuilder().Rows(rows)
}

// WithHeader creates a builder with a header.
func WithHeader(header string) *Builder {
	return NewBuilder().Header(header)
}
