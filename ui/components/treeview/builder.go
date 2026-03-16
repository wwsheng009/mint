package treeview

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing TreeView VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new TreeView builder.
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

// ComponentID sets the logical component ID for intent routing.
func (b *Builder) ComponentID(id string) *Builder {
	b.vnode.SetComponentID(id)
	return b
}

// Nodes sets the tree nodes.
func (b *Builder) Nodes(nodes []TreeNode) *Builder {
	b.vnode.SetNodes(nodes)
	return b
}

// ExpandLevel sets the default expand level.
func (b *Builder) ExpandLevel(level int) *Builder {
	b.vnode.SetExpandLevel(level)
	return b
}

// ShowIcons enables/disables node icons.
func (b *Builder) ShowIcons(show bool) *Builder {
	b.vnode.SetShowIcons(show)
	return b
}

// ShowLineNumbers enables/disables line numbers.
func (b *Builder) ShowLineNumbers(show bool) *Builder {
	b.vnode.SetShowLineNums(show)
	return b
}

// Compact sets compact display mode.
func (b *Builder) Compact(compact bool) *Builder {
	b.vnode.SetCompact(compact)
	return b
}

// ShowBorder toggles the outer border.
func (b *Builder) ShowBorder(show bool) *Builder {
	b.vnode.SetShowBorder(show)
	return b
}

// ShowScrollbar toggles the vertical scrollbar indicator.
func (b *Builder) ShowScrollbar(show bool) *Builder {
	b.vnode.SetShowScrollbar(show)
	return b
}

// TreeStyle sets the style for tree items.
func (b *Builder) TreeStyle(s style.Style) *Builder {
	b.vnode.SetTreeStyle(s)
	return b
}

// SelectedStyle sets the style for the selected item.
func (b *Builder) SelectedStyle(s style.Style) *Builder {
	b.vnode.SetSelectedStyle(s)
	return b
}

// IconStyle sets the style for icons.
func (b *Builder) IconStyle(s style.Style) *Builder {
	b.vnode.SetIconStyle(s)
	return b
}

// RowStyleFn sets the style function for rows.
func (b *Builder) RowStyleFn(fn func(int, TreeNode) style.Style) *Builder {
	b.vnode.SetRowStyleFn(fn)
	return b
}

// MatchStyle sets the style for search matches.
func (b *Builder) MatchStyle(s style.Style) *Builder {
	b.vnode.SetMatchStyle(s)
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

// ScrollbarStyle sets the style for the scrollbar.
func (b *Builder) ScrollbarStyle(s style.Style) *Builder {
	b.vnode.SetScrollbarStyle(s)
	return b
}

// ScrollOffset sets the initial scroll offset.
func (b *Builder) ScrollOffset(offset int) *Builder {
	b.vnode.SetScrollOffset(offset)
	return b
}

// ScrollOffsetControlled sets a controlled scroll offset.
func (b *Builder) ScrollOffsetControlled(offset int) *Builder {
	b.vnode.SetScrollOffsetControlled(offset)
	return b
}

// SelectedIndex sets the currently selected node index.
func (b *Builder) SelectedIndex(index int) *Builder {
	b.vnode.SetSelectedIndex(index)
	return b
}

// SelectedIndexControlled sets a controlled selected index.
func (b *Builder) SelectedIndexControlled(index int) *Builder {
	b.vnode.SetSelectedIndexControlled(index)
	return b
}

// ViewportHeight sets the visible height for scrolling.
func (b *Builder) ViewportHeight(height int) *Builder {
	b.vnode.SetViewportHeight(height)
	return b
}

// SearchQuery sets the filter query for nodes.
func (b *Builder) SearchQuery(query string) *Builder {
	b.vnode.SetSearchQuery(query)
	return b
}

// SearchQueryControlled sets a controlled filter query for nodes.
func (b *Builder) SearchQueryControlled(query string) *Builder {
	b.vnode.SetSearchQueryControlled(query)
	return b
}

// SearchFn sets the custom search function.
func (b *Builder) SearchFn(fn func(TreeNode, string) bool) *Builder {
	b.vnode.SetSearchFn(fn)
	return b
}

// SelectionMode sets the selection mode for checkbox-style selection.
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

// CheckedKeys sets checked nodes by key (controlled).
func (b *Builder) CheckedKeys(keys map[string]bool) *Builder {
	b.vnode.SetCheckedKeys(keys)
	return b
}

// CheckedPaths sets checked nodes by path (controlled).
func (b *Builder) CheckedPaths(paths ...string) *Builder {
	b.vnode.SetCheckedPaths(paths...)
	return b
}

// OnSelectionChange sets the intent emitted when checkbox selection changes.
func (b *Builder) OnSelectionChange(selectionIntent intent.Intent) *Builder {
	b.vnode.SetSelectionIntent(selectionIntent)
	return b
}

// OnReorder sets the intent emitted when sibling drag reorder changes node order.
func (b *Builder) OnReorder(reorderIntent intent.Intent) *Builder {
	b.vnode.SetReorderIntent(reorderIntent)
	return b
}

// SelectionForField binds checked nodes to a field.
func (b *Builder) SelectionForField(binding intent.FieldIntent) *Builder {
	b.vnode.SetSelectionFieldIntent(binding)
	return b
}

// OnLazyLoad sets a synchronous lazy-load hook.
func (b *Builder) OnLazyLoad(fn func(TreeNode)) *Builder {
	b.vnode.SetLazyLoadFn(fn)
	return b
}

// OnLazyLoadChildren sets a lazy-load hook that returns children to insert.
func (b *Builder) OnLazyLoadChildren(fn func(TreeNode) []TreeNode) *Builder {
	b.vnode.SetLazyLoadChildrenFn(fn)
	return b
}

// ExpandedKeys sets the expanded state map (controlled).
func (b *Builder) ExpandedKeys(keys map[string]bool) *Builder {
	b.vnode.SetExpandedKeys(keys)
	return b
}

// ExpandedPaths marks the given paths as expanded (controlled).
func (b *Builder) ExpandedPaths(paths ...string) *Builder {
	b.vnode.SetExpandedPaths(paths...)
	return b
}

// AllowScroll enables/disables scrolling.
func (b *Builder) AllowScroll(allow bool) *Builder {
	b.vnode.SetAllowScroll(allow)
	return b
}

// AllowExpand enables/disables expand/collapse.
func (b *Builder) AllowExpand(allow bool) *Builder {
	b.vnode.SetAllowExpand(allow)
	return b
}

// Reorderable enables/disables sibling drag reorder.
func (b *Builder) Reorderable(reorderable bool) *Builder {
	b.vnode.SetReorderable(reorderable)
	return b
}

// AddNode adds a single node to the tree.
func (b *Builder) AddNode(node TreeNode) *Builder {
	b.vnode.AddNode(node)
	return b
}

// FromLines creates tree nodes from pre-formatted lines.
func (b *Builder) FromLines(lines []string) *Builder {
	b.vnode.FromLines(lines)
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

// Of creates a TreeView with the given nodes.
func Of(nodes []TreeNode) rtui.VNode {
	return NewBuilder().Nodes(nodes).Build()
}

// OfLines creates a TreeView from pre-formatted lines.
func OfLines(lines []string) rtui.VNode {
	return NewBuilder().FromLines(lines).Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// TreeView creates a new TreeView builder.
func TreeView() *Builder {
	return NewBuilder()
}

// WithNodes creates a builder with pre-configured nodes.
func WithNodes(nodes []TreeNode) *Builder {
	return NewBuilder().Nodes(nodes)
}

// WithLines creates a builder with pre-formatted lines.
func WithLines(lines []string) *Builder {
	return NewBuilder().FromLines(lines)
}
