package treeview

import (
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

// ScrollOffset sets the initial scroll offset.
func (b *Builder) ScrollOffset(offset int) *Builder {
	b.vnode.SetScrollOffset(offset)
	return b
}

// SelectedIndex sets the currently selected node index.
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

// AllowExpand enables/disables expand/collapse.
func (b *Builder) AllowExpand(allow bool) *Builder {
	b.vnode.SetAllowExpand(allow)
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
