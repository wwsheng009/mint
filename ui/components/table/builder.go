package table

import (
	"github.com/wwsheng009/mint/runtime/intent"
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

// EmptyText sets the text shown when no rows match.
func (b *Builder) EmptyText(text string) *Builder {
	b.vnode.SetEmptyText(text)
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

// ShowBorder toggles the outer table border.
func (b *Builder) ShowBorder(show bool) *Builder {
	b.vnode.SetShowBorder(show)
	return b
}

// BorderStyle sets the border style.
func (b *Builder) BorderStyle(s style.Style) *Builder {
	b.vnode.SetBorderStyle(s)
	return b
}

// SelectedStyle sets the selected row style.
func (b *Builder) SelectedStyle(s style.Style) *Builder {
	b.vnode.SetSelectedStyle(s)
	return b
}

// StatusStyle sets the footer/status line style.
func (b *Builder) StatusStyle(s style.Style) *Builder {
	b.vnode.SetStatusStyle(s)
	return b
}

// StatusText overrides the footer/status line text.
func (b *Builder) StatusText(text string) *Builder {
	b.vnode.SetStatusText(text)
	return b
}

// FilterStyle sets the active filter row style.
func (b *Builder) FilterStyle(s style.Style) *Builder {
	b.vnode.SetFilterStyle(s)
	return b
}

// ShowFooter toggles the status/footer line.
func (b *Builder) ShowFooter(show bool) *Builder {
	b.vnode.SetShowFooter(show)
	return b
}

// ShowScrollbar toggles the vertical scrollbar indicator.
func (b *Builder) ShowScrollbar(show bool) *Builder {
	b.vnode.SetShowScrollbar(show)
	return b
}

// ScrollbarStyle sets the scrollbar style.
func (b *Builder) ScrollbarStyle(s style.Style) *Builder {
	b.vnode.SetScrollbarStyle(s)
	return b
}

// PageSize enables pagination with the given page size.
func (b *Builder) PageSize(size int) *Builder {
	b.vnode.SetPageSize(size)
	return b
}

// CurrentPage sets the current page in controlled mode.
func (b *Builder) CurrentPage(page int) *Builder {
	b.vnode.SetCurrentPage(page)
	return b
}

// SearchQuery applies case-insensitive search across all columns.
func (b *Builder) SearchQuery(query string) *Builder {
	b.vnode.SetSearchQuery(query)
	return b
}

// Filters applies case-insensitive column filters.
func (b *Builder) Filters(filters map[int]string) *Builder {
	b.vnode.SetFilters(filters)
	return b
}

// ExpandedContent sets the expandable detail content by source row index.
func (b *Builder) ExpandedContent(content map[int]string) *Builder {
	b.vnode.SetExpandedContent(content)
	return b
}

// TreeParents sets the tree parent mapping by source row index.
func (b *Builder) TreeParents(parents map[int]int) *Builder {
	b.vnode.SetTreeParents(parents)
	return b
}

// TreeParent sets the tree parent for a single source row.
func (b *Builder) TreeParent(index int, parentIndex int) *Builder {
	b.vnode.SetTreeParent(index, parentIndex)
	return b
}

// ExpandedRow sets or clears expandable detail content for a single source row.
func (b *Builder) ExpandedRow(index int, content string) *Builder {
	b.vnode.SetExpandedRow(index, content)
	return b
}

// ExpandedIndices sets the expanded source row indices in controlled mode.
func (b *Builder) ExpandedIndices(indices ...int) *Builder {
	b.vnode.SetExpandedIndices(indices)
	return b
}

// Filter applies or clears a filter for a single column.
func (b *Builder) Filter(columnIndex int, value string) *Builder {
	b.vnode.SetFilter(columnIndex, value)
	return b
}

// SortBy sets the sort state in controlled mode.
func (b *Builder) SortBy(columnIndex int, descending bool) *Builder {
	b.vnode.SetSortBy(columnIndex, descending)
	return b
}

// SelectedIndex sets the selected row index in controlled mode.
func (b *Builder) SelectedIndex(index int) *Builder {
	b.vnode.SetSelectedIndex(index)
	return b
}

// OnSelectionChange sets the intent emitted when checkbox selection changes.
func (b *Builder) OnSelectionChange(selectionIntent intent.Intent) *Builder {
	b.vnode.SetSelectionIntent(selectionIntent)
	return b
}

// CheckedIndices sets the checked source row indices in controlled mode.
func (b *Builder) CheckedIndices(indices ...int) *Builder {
	b.vnode.SetCheckedIndices(indices)
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

// OnChange sets a fallback change intent emitted on interactive state changes.
func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.vnode.SetIntent(changeIntent)
	return b
}

// ForField binds the selected source row index to FieldChangeIntent.
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.vnode.SetFieldIntent(binding)
	return b
}

// SelectionForField binds checked source row indices to a field.
func (b *Builder) SelectionForField(binding intent.FieldBinding) *Builder {
	b.vnode.SetSelectionFieldIntent(binding)
	return b
}

// PageForField binds the current page index to a field.
func (b *Builder) PageForField(binding intent.FieldBinding) *Builder {
	b.vnode.SetPageFieldIntent(binding)
	return b
}

// OnExpand sets an intent emitted when expanded row state changes.
func (b *Builder) OnExpand(expandIntent intent.Intent) *Builder {
	b.vnode.SetExpandIntent(expandIntent)
	return b
}

// ExpandForField binds expanded source row indices to FieldChangeIntent.
func (b *Builder) ExpandForField(binding intent.FieldBinding) *Builder {
	b.vnode.SetExpandFieldIntent(binding)
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

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// NewTable creates a new Table VNode (alias for New, for backward compatibility).
// This matches the old data.NewTable() API.
func NewTable() *VNode {
	return New()
}

// TableBuilder is an alias for Builder (for backward compatibility).
// This matches the old data.TableBuilder type.
type TableBuilder = Builder

// TableColumn is already defined in vnode.go, re-exported for convenience.
