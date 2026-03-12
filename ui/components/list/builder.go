package list

import (
	"github.com/wwsheng009/mint/runtime/intent"
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

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.vnode.SetID(id)
	return b
}

// ComponentID sets the logical component ID for state-change intents.
func (b *Builder) ComponentID(id string) *Builder {
	b.vnode.SetComponentID(id)
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

// RowStyleFn sets a dynamic style function that determines style per row.
// The function is called with row index and text, and should return the style.
func (b *Builder) RowStyleFn(fn func(int, string) style.Style) *Builder {
	b.vnode.SetRowStyleFn(fn)
	return b
}

// MatchStyle sets the style for rows matched by the active search query.
func (b *Builder) MatchStyle(s style.Style) *Builder {
	b.vnode.SetMatchStyle(s)
	return b
}

// SearchQuery filters visible rows using case-insensitive substring matching.
func (b *Builder) SearchQuery(query string) *Builder {
	b.vnode.SetSearchQuery(query)
	return b
}

// SearchFn sets a custom search function.
func (b *Builder) SearchFn(fn func(string, string) bool) *Builder {
	b.vnode.SetSearchFn(fn)
	return b
}

// ShowSearchStats toggles the search stats row.
func (b *Builder) ShowSearchStats(show bool) *Builder {
	b.vnode.SetShowSearchStats(show)
	return b
}

// SearchStatsStyle sets the style for the search stats row.
func (b *Builder) SearchStatsStyle(s style.Style) *Builder {
	b.vnode.SetSearchStatsStyle(s)
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

// ShowScrollbar controls whether a vertical scrollbar is shown when scrollable.
func (b *Builder) ShowScrollbar(show bool) *Builder {
	b.vnode.SetShowScrollbar(show)
	return b
}

// ScrollbarStyle sets the style for the scrollbar.
func (b *Builder) ScrollbarStyle(s style.Style) *Builder {
	b.vnode.SetScrollbarStyle(s)
	return b
}

// OnChange sets the change intent emitted when selection changes.
func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.vnode.SetChangeIntent(changeIntent)
	return b
}

// OnSelectionChange sets the intent emitted when checkbox selection changes.
func (b *Builder) OnSelectionChange(selectionIntent intent.Intent) *Builder {
	b.vnode.SetSelectionIntent(selectionIntent)
	return b
}

// ScrollOffset sets the scroll offset in controlled mode.
func (b *Builder) ScrollOffset(offset int) *Builder {
	b.vnode.SetScrollOffset(offset)
	return b
}

// ScrollOffsetControlled sets the scroll offset in controlled mode.
func (b *Builder) ScrollOffsetControlled(offset int) *Builder {
	b.vnode.SetScrollOffsetControlled(offset)
	return b
}

// InitialScrollOffset sets the initial scroll offset in uncontrolled mode.
func (b *Builder) InitialScrollOffset(offset int) *Builder {
	b.vnode.SetInitialScrollOffset(offset)
	return b
}

// SelectedIndex sets the selected row index in controlled mode.
func (b *Builder) SelectedIndex(index int) *Builder {
	b.vnode.SetSelectedIndex(index)
	return b
}

// SelectedIndexControlled sets the selected row index in controlled mode.
func (b *Builder) SelectedIndexControlled(index int) *Builder {
	b.vnode.SetSelectedIndexControlled(index)
	return b
}

// InitialSelectedIndex sets the initial selected row index in uncontrolled mode.
func (b *Builder) InitialSelectedIndex(index int) *Builder {
	b.vnode.SetInitialSelectedIndex(index)
	return b
}

// CheckedIndices sets the checked rows in controlled mode.
func (b *Builder) CheckedIndices(indices ...int) *Builder {
	b.vnode.SetCheckedIndices(indices)
	return b
}

// InitialCheckedIndices sets the initial checked rows in uncontrolled mode.
func (b *Builder) InitialCheckedIndices(indices ...int) *Builder {
	b.vnode.SetInitialCheckedIndices(indices)
	return b
}

// SelectionMode sets the checkbox selection mode.
func (b *Builder) SelectionMode(mode SelectionMode) *Builder {
	b.vnode.SetSelectionMode(mode)
	return b
}

// SingleSelect enables single-select checkbox behavior.
func (b *Builder) SingleSelect() *Builder {
	return b.SelectionMode(SelectionSingle)
}

// MultiSelect enables multi-select checkbox behavior.
func (b *Builder) MultiSelect() *Builder {
	return b.SelectionMode(SelectionMultiple)
}

// ViewportHeight sets the visible height for scrolling.
func (b *Builder) ViewportHeight(height int) *Builder {
	b.vnode.SetViewportHeight(height)
	return b
}

// ForField binds the list selection to a state field using FieldBinding.
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.vnode.SetProps(rtui.Props{
		"changeIntent": binding,
	})
	return b
}

// SelectionForField binds checkbox selection changes to a state field.
func (b *Builder) SelectionForField(binding intent.FieldBinding) *Builder {
	b.vnode.SetProps(rtui.Props{
		"selectionIntent": binding,
	})
	return b
}

// ForForm binds the list to a form using FormBinding.
func (b *Builder) ForForm(binding intent.FormBinding) *Builder {
	b.vnode.SetFormID(binding.GetFormID())
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
