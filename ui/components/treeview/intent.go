package treeview

import (
	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// Treeview Intents (Phase 10: Intent Bubble)
// =============================================================================

// NodeSelectIntent is emitted when a tree node is selected
type NodeSelectIntent struct {
	// NodeIndex is the index of the selected node in the nodes array
	NodeIndex int

	// Path is the unique path identifier of the selected node
	Path string

	// NodeID is the optional ID of the node
	NodeID int

	// NodeType is the type of node (folder, file, etc.)
	NodeType string

	// Content is the display content of the node
	Content string

	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

// IntentType implements intent.Intent
func (i NodeSelectIntent) IntentType() string {
	return "treeview.NodeSelectIntent"
}

// Priority implements PriorityAware
func (i NodeSelectIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i NodeSelectIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// NodeSelectIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (i NodeSelectIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i NodeSelectIntent) GetComponentID() string {
	return i.ComponentID
}

// NodeExpandIntent is emitted when a folder node is expanded
type NodeExpandIntent struct {
	// NodeIndex is the index of the expanded node
	NodeIndex int

	// Path is the unique path identifier
	Path string

	// NodeID is the optional ID
	NodeID int

	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

// IntentType implements intent.Intent
func (i NodeExpandIntent) IntentType() string {
	return "treeview.NodeExpandIntent"
}

// Priority implements PriorityAware
func (i NodeExpandIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i NodeExpandIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// NodeExpandIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (i NodeExpandIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i NodeExpandIntent) GetComponentID() string {
	return i.ComponentID
}

// NodeCollapseIntent is emitted when a folder node is collapsed
type NodeCollapseIntent struct {
	// NodeIndex is the index of the collapsed node
	NodeIndex int

	// Path is the unique path identifier
	Path string

	// NodeID is the optional ID
	NodeID int

	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

// IntentType implements intent.Intent
func (i NodeCollapseIntent) IntentType() string {
	return "treeview.NodeCollapseIntent"
}

// Priority implements PriorityAware
func (i NodeCollapseIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i NodeCollapseIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// NodeCollapseIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (i NodeCollapseIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i NodeCollapseIntent) GetComponentID() string {
	return i.ComponentID
}

// NodeReorderIntent is emitted when a node subtree is reordered among siblings.
type NodeReorderIntent struct {
	FromIndex        int
	ToIndex          int
	FromVisibleIndex int
	ToVisibleIndex   int
	Path             string
	NodeID           int
	NodeType         string
	Content          string
	ParentKey        string
	ComponentID      string
}

func (i NodeReorderIntent) IntentType() string {
	return "treeview.NodeReorderIntent"
}

func (i NodeReorderIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i NodeReorderIntent) IsTransition() bool {
	return false
}

func (i NodeReorderIntent) IsGlobal() bool {
	return false
}

func (i NodeReorderIntent) GetComponentID() string {
	return i.ComponentID
}

// NavigationIntent is emitted for navigation events (up/down/pageup/pagedown)
// Useful for parent components or external controllers to monitor selection changes
type NavigationIntent struct {
	// Direction indicates the navigation direction
	Direction string // "up", "down", "pageup", "pagedown", "home", "end"

	// FromIndex is the previous selected index
	FromIndex int

	// ToIndex is the new selected index
	ToIndex int

	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

// IntentType implements intent.Intent
func (i NavigationIntent) IntentType() string {
	return "treeview.NavigationIntent"
}

// Priority implements PriorityAware
func (i NavigationIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition returns false (synchronous intent)
func (i NavigationIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// NavigationIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (i NavigationIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i NavigationIntent) GetComponentID() string {
	return i.ComponentID
}

// ExpandAllIntent expands all nodes in the treeview.
type ExpandAllIntent struct {
	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

func (i ExpandAllIntent) IntentType() string {
	return "treeview.ExpandAllIntent"
}

func (i ExpandAllIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i ExpandAllIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// ExpandAllIntent bubbles locally through Parent() chain.
func (i ExpandAllIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i ExpandAllIntent) GetComponentID() string {
	return i.ComponentID
}

// CollapseAllIntent collapses all nodes in the treeview.
type CollapseAllIntent struct {
	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

func (i CollapseAllIntent) IntentType() string {
	return "treeview.CollapseAllIntent"
}

func (i CollapseAllIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i CollapseAllIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// CollapseAllIntent bubbles locally through Parent() chain.
func (i CollapseAllIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i CollapseAllIntent) GetComponentID() string {
	return i.ComponentID
}

// SearchNextIntent moves selection to the next search match.
type SearchNextIntent struct {
	ComponentID string
}

func (i SearchNextIntent) IntentType() string {
	return "treeview.SearchNextIntent"
}

func (i SearchNextIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i SearchNextIntent) IsTransition() bool {
	return false
}

func (i SearchNextIntent) IsGlobal() bool {
	return false
}

func (i SearchNextIntent) GetComponentID() string {
	return i.ComponentID
}

// SearchPrevIntent moves selection to the previous search match.
type SearchPrevIntent struct {
	ComponentID string
}

func (i SearchPrevIntent) IntentType() string {
	return "treeview.SearchPrevIntent"
}

func (i SearchPrevIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i SearchPrevIntent) IsTransition() bool {
	return false
}

func (i SearchPrevIntent) IsGlobal() bool {
	return false
}

func (i SearchPrevIntent) GetComponentID() string {
	return i.ComponentID
}

// SearchStatsIntent reports current search match stats.
type SearchStatsIntent struct {
	ComponentID string
	Query       string
	Total       int
	Selected    int
}

func (i SearchStatsIntent) IntentType() string {
	return "treeview.SearchStatsIntent"
}

func (i SearchStatsIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i SearchStatsIntent) IsTransition() bool {
	return false
}

func (i SearchStatsIntent) IsGlobal() bool {
	return false
}

func (i SearchStatsIntent) GetComponentID() string {
	return i.ComponentID
}

// ScrollIntent is emitted when the treeview scroll offset changes.
type ScrollIntent struct {
	Offset      int
	Delta       int
	ViewSize    int
	ContentSize int

	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

func (i ScrollIntent) IntentType() string {
	return "treeview.ScrollIntent"
}

func (i ScrollIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i ScrollIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// ScrollIntent bubbles locally through Parent() chain.
func (i ScrollIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i ScrollIntent) GetComponentID() string {
	return i.ComponentID
}

// SelectionChangeIntent is emitted when checkbox-style selection changes.
type SelectionChangeIntent struct {
	ComponentID    string
	SelectionMode  SelectionMode
	CheckedKeys    []string
	CheckedIndices []int
	CheckedPaths   []string
	CheckedNodeIDs []int
}

func (i SelectionChangeIntent) IntentType() string {
	return "treeview.SelectionChangeIntent"
}

func (i SelectionChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i SelectionChangeIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// SelectionChangeIntent bubbles locally through Parent() chain.
func (i SelectionChangeIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i SelectionChangeIntent) GetComponentID() string {
	return i.ComponentID
}

// LazyLoadIntent is emitted when a lazy node is expanded.
type LazyLoadIntent struct {
	// NodeIndex is the index of the expanded node
	NodeIndex int

	// Path is the unique path identifier
	Path string

	// NodeID is the optional ID
	NodeID int

	// ComponentID is the treeview component ID for routing (optional)
	ComponentID string
}

func (i LazyLoadIntent) IntentType() string {
	return "treeview.LazyLoadIntent"
}

func (i LazyLoadIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i LazyLoadIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// LazyLoadIntent bubbles locally through Parent() chain.
func (i LazyLoadIntent) IsGlobal() bool {
	return false
}

// GetComponentID implements intent.GetComponentID for routing.
func (i LazyLoadIntent) GetComponentID() string {
	return i.ComponentID
}

// LazyLoadSuccessIntent applies lazy-loaded children to a node.
type LazyLoadSuccessIntent struct {
	NodeIndex   int
	Path        string
	NodeID      int
	Children    []TreeNode
	Replace     bool
	ComponentID string
}

func (i LazyLoadSuccessIntent) IntentType() string {
	return "treeview.LazyLoadSuccessIntent"
}

func (i LazyLoadSuccessIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i LazyLoadSuccessIntent) IsTransition() bool {
	return false
}

func (i LazyLoadSuccessIntent) IsGlobal() bool {
	return false
}

func (i LazyLoadSuccessIntent) GetComponentID() string {
	return i.ComponentID
}

// LazyLoadFailureIntent records an async lazy-load failure on a node.
type LazyLoadFailureIntent struct {
	NodeIndex   int
	Path        string
	NodeID      int
	Error       string
	ComponentID string
}

func (i LazyLoadFailureIntent) IntentType() string {
	return "treeview.LazyLoadFailureIntent"
}

func (i LazyLoadFailureIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (i LazyLoadFailureIntent) IsTransition() bool {
	return false
}

func (i LazyLoadFailureIntent) IsGlobal() bool {
	return false
}

func (i LazyLoadFailureIntent) GetComponentID() string {
	return i.ComponentID
}

// =============================================================================
// Intent Constructors
// =============================================================================

// NodeSelect creates a NodeSelectIntent
func NodeSelect(nodeIndex int, path string, nodeID int, nodeType, content string) NodeSelectIntent {
	return NodeSelectIntent{
		NodeIndex: nodeIndex,
		Path:      path,
		NodeID:    nodeID,
		NodeType:  nodeType,
		Content:   content,
	}
}

// NodeSelectWithID creates a NodeSelectIntent with component ID
func NodeSelectWithID(componentID string, nodeIndex int, path string, nodeID int, nodeType, content string) NodeSelectIntent {
	return NodeSelectIntent{
		ComponentID: componentID,
		NodeIndex:   nodeIndex,
		Path:        path,
		NodeID:      nodeID,
		NodeType:    nodeType,
		Content:     content,
	}
}

// NodeExpand creates a NodeExpandIntent
func NodeExpand(nodeIndex int, path string, nodeID int) NodeExpandIntent {
	return NodeExpandIntent{
		NodeIndex: nodeIndex,
		Path:      path,
		NodeID:    nodeID,
	}
}

// NodeExpandWithID creates a NodeExpandIntent with component ID
func NodeExpandWithID(componentID string, nodeIndex int, path string, nodeID int) NodeExpandIntent {
	return NodeExpandIntent{
		ComponentID: componentID,
		NodeIndex:   nodeIndex,
		Path:        path,
		NodeID:      nodeID,
	}
}

// NodeCollapse creates a NodeCollapseIntent
func NodeCollapse(nodeIndex int, path string, nodeID int) NodeCollapseIntent {
	return NodeCollapseIntent{
		NodeIndex: nodeIndex,
		Path:      path,
		NodeID:    nodeID,
	}
}

// NodeCollapseWithID creates a NodeCollapseIntent with component ID
func NodeCollapseWithID(componentID string, nodeIndex int, path string, nodeID int) NodeCollapseIntent {
	return NodeCollapseIntent{
		ComponentID: componentID,
		NodeIndex:   nodeIndex,
		Path:        path,
		NodeID:      nodeID,
	}
}

// NodeReorder creates a NodeReorderIntent.
func NodeReorder(fromIndex, toIndex, fromVisibleIndex, toVisibleIndex int, path string, nodeID int, nodeType, content, parentKey string) NodeReorderIntent {
	return NodeReorderIntent{
		FromIndex:        fromIndex,
		ToIndex:          toIndex,
		FromVisibleIndex: fromVisibleIndex,
		ToVisibleIndex:   toVisibleIndex,
		Path:             path,
		NodeID:           nodeID,
		NodeType:         nodeType,
		Content:          content,
		ParentKey:        parentKey,
	}
}

// NodeReorderWithID creates a NodeReorderIntent with component ID.
func NodeReorderWithID(componentID string, fromIndex, toIndex, fromVisibleIndex, toVisibleIndex int, path string, nodeID int, nodeType, content, parentKey string) NodeReorderIntent {
	reorderIntent := NodeReorder(fromIndex, toIndex, fromVisibleIndex, toVisibleIndex, path, nodeID, nodeType, content, parentKey)
	reorderIntent.ComponentID = componentID
	return reorderIntent
}

// Navigation creates a NavigationIntent
func Navigation(direction string, fromIndex, toIndex int) NavigationIntent {
	return NavigationIntent{
		Direction: direction,
		FromIndex: fromIndex,
		ToIndex:   toIndex,
	}
}

// NavigationWithID creates a NavigationIntent with component ID
func NavigationWithID(componentID string, direction string, fromIndex, toIndex int) NavigationIntent {
	return NavigationIntent{
		ComponentID: componentID,
		Direction:   direction,
		FromIndex:   fromIndex,
		ToIndex:     toIndex,
	}
}

// Scroll creates a ScrollIntent.
func Scroll(offset, delta, viewSize, contentSize int) ScrollIntent {
	return ScrollIntent{
		Offset:      offset,
		Delta:       delta,
		ViewSize:    viewSize,
		ContentSize: contentSize,
	}
}

// ScrollWithID creates a ScrollIntent with component ID.
func ScrollWithID(componentID string, offset, delta, viewSize, contentSize int) ScrollIntent {
	return ScrollIntent{
		ComponentID: componentID,
		Offset:      offset,
		Delta:       delta,
		ViewSize:    viewSize,
		ContentSize: contentSize,
	}
}

// SelectionChange creates a SelectionChangeIntent.
func SelectionChange(componentID string, mode SelectionMode, keys []string, indices []int, paths []string, nodeIDs []int) SelectionChangeIntent {
	return SelectionChangeIntent{
		ComponentID:    componentID,
		SelectionMode:  mode,
		CheckedKeys:    append([]string(nil), keys...),
		CheckedIndices: append([]int(nil), indices...),
		CheckedPaths:   append([]string(nil), paths...),
		CheckedNodeIDs: append([]int(nil), nodeIDs...),
	}
}

// LazyLoad creates a LazyLoadIntent.
func LazyLoad(nodeIndex int, path string, nodeID int) LazyLoadIntent {
	return LazyLoadIntent{
		NodeIndex: nodeIndex,
		Path:      path,
		NodeID:    nodeID,
	}
}

// LazyLoadWithID creates a LazyLoadIntent with component ID.
func LazyLoadWithID(componentID string, nodeIndex int, path string, nodeID int) LazyLoadIntent {
	return LazyLoadIntent{
		ComponentID: componentID,
		NodeIndex:   nodeIndex,
		Path:        path,
		NodeID:      nodeID,
	}
}

// LazyLoadSuccess creates a LazyLoadSuccessIntent.
func LazyLoadSuccess(nodeIndex int, path string, nodeID int, children []TreeNode) LazyLoadSuccessIntent {
	return LazyLoadSuccessIntent{
		NodeIndex: nodeIndex,
		Path:      path,
		NodeID:    nodeID,
		Children:  append([]TreeNode(nil), children...),
	}
}

// LazyLoadSuccessWithID creates a LazyLoadSuccessIntent with component ID.
func LazyLoadSuccessWithID(componentID string, nodeIndex int, path string, nodeID int, children []TreeNode) LazyLoadSuccessIntent {
	return LazyLoadSuccessIntent{
		ComponentID: componentID,
		NodeIndex:   nodeIndex,
		Path:        path,
		NodeID:      nodeID,
		Children:    append([]TreeNode(nil), children...),
	}
}

// LazyLoadFailure creates a LazyLoadFailureIntent.
func LazyLoadFailure(nodeIndex int, path string, nodeID int, err string) LazyLoadFailureIntent {
	return LazyLoadFailureIntent{
		NodeIndex: nodeIndex,
		Path:      path,
		NodeID:    nodeID,
		Error:     err,
	}
}

// LazyLoadFailureWithID creates a LazyLoadFailureIntent with component ID.
func LazyLoadFailureWithID(componentID string, nodeIndex int, path string, nodeID int, err string) LazyLoadFailureIntent {
	return LazyLoadFailureIntent{
		ComponentID: componentID,
		NodeIndex:   nodeIndex,
		Path:        path,
		NodeID:      nodeID,
		Error:       err,
	}
}

// ExpandAll creates an ExpandAllIntent.
func ExpandAll() ExpandAllIntent {
	return ExpandAllIntent{}
}

// ExpandAllWithID creates an ExpandAllIntent with component ID.
func ExpandAllWithID(componentID string) ExpandAllIntent {
	return ExpandAllIntent{ComponentID: componentID}
}

// CollapseAll creates a CollapseAllIntent.
func CollapseAll() CollapseAllIntent {
	return CollapseAllIntent{}
}

// CollapseAllWithID creates a CollapseAllIntent with component ID.
func CollapseAllWithID(componentID string) CollapseAllIntent {
	return CollapseAllIntent{ComponentID: componentID}
}

// SearchNext creates a SearchNextIntent.
func SearchNext() SearchNextIntent {
	return SearchNextIntent{}
}

// SearchNextWithID creates a SearchNextIntent with component ID.
func SearchNextWithID(componentID string) SearchNextIntent {
	return SearchNextIntent{ComponentID: componentID}
}

// SearchPrev creates a SearchPrevIntent.
func SearchPrev() SearchPrevIntent {
	return SearchPrevIntent{}
}

// SearchPrevWithID creates a SearchPrevIntent with component ID.
func SearchPrevWithID(componentID string) SearchPrevIntent {
	return SearchPrevIntent{ComponentID: componentID}
}

// SearchStats creates a SearchStatsIntent.
func SearchStats(query string, total, selected int) SearchStatsIntent {
	return SearchStatsIntent{
		Query:    query,
		Total:    total,
		Selected: selected,
	}
}

// SearchStatsWithID creates a SearchStatsIntent with component ID.
func SearchStatsWithID(componentID, query string, total, selected int) SearchStatsIntent {
	return SearchStatsIntent{
		ComponentID: componentID,
		Query:       query,
		Total:       total,
		Selected:    selected,
	}
}
