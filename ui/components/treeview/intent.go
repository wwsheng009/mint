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
