package inspector

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/ui"
)

// TreeNode represents a node in the layout tree
type TreeNode struct {
	VNode    ui.VNode
	Info     ElementInfo
	Children []*TreeNode
	Parent   *TreeNode
	Expanded bool
	Level    int    // Depth in tree (0 = root)
	Index    int    // Index among siblings
	Path     string // Full path from root
	UniqueID string // Unique identifier for expand/collapse tracking
}

// TreeView provides tree visualization and traversal
type TreeView struct {
	root             *TreeNode
	expanded         map[string]bool // Track expanded nodes by unique ID (path-based)
	selectedUniqueID string          // Track selected node by unique ID
	showIcons        bool            // Show type icons
	showPaths        bool            // Show paths
	compact          bool            // Use compact display
	maxDepth         int             // Maximum traversal depth
	maxNodes         int             // Maximum nodes to display
	changeCount      int64           // Counter for tree changes (structure or expansion)
	lastRootVNode    ui.VNode        // Last root VNode (to avoid unnecessary rebuilds)
}

// NewTreeView creates a new TreeView instance
func NewTreeView() *TreeView {
	return &TreeView{
		expanded:  make(map[string]bool),
		showIcons: true,
		showPaths: false,
		compact:   false,
		maxDepth:  100, // Effectively unlimited
		maxNodes:  1000,
	}
}

// SetRoot sets the root VNode for the tree
func (tv *TreeView) SetRoot(root ui.VNode) error {
	if root == nil {
		tv.root = nil
		tv.lastRootVNode = nil
		return nil
	}

	// Check if VNode has actually changed (by pointer comparison)
	// This avoids expensive tree rebuilding when the same VNode is passed multiple times
	if tv.lastRootVNode == root {
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			log.UILogger.Debug("[TreeView] SetRoot: VNode unchanged, skipping rebuild\n")
		}
		return nil
	}

	// Build tree structure (root has index 0)
	tv.root = tv.buildTree(root, nil, 0, "", 0)
	tv.lastRootVNode = root
	tv.changeCount++

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		log.UILogger.Debug("[TreeView] SetRoot: VNode changed, rebuilding tree (changeCount=%d)\n", tv.changeCount)
	}

	return nil
}

// buildTree recursively builds the tree structure
func (tv *TreeView) buildTree(vnode ui.VNode, parent *TreeNode, level int, path string, index int) *TreeNode {
	if vnode == nil {
		return nil
	}

	// ✨ NEW: Use VNode's Key if it's a path-based key from Fiber reconciliation
	// Fiber reconciliation sets VNode keys to path-based keys like /root/base[0]/vstack[0]/panel[0]
	var nodePath string
	if keyer, ok := vnode.(interface{ Key() string }); ok {
		vnodeKey := keyer.Key()
		// Check if this is a path-based key (set by Fiber reconciliation)
		if vnodeKey != "" && strings.HasPrefix(vnodeKey, "/root/") {
			// Use the Fiber-generated path as our display path
			// Remove the "/root/" prefix for cleaner display
			// /root/base[0]/vstack[0]/panel[0] → base[0]/vstack[0]/panel[0]
			if len(vnodeKey) > 6 { // "/root/" is 6 characters
				nodePath = vnodeKey[6:] // Skip "/root/" prefix
			} else {
				nodePath = vnodeKey
			}
		}
	}

	// Fallback: Generate path using old logic if no path-based key available
	if nodePath == "" {
		if path == "" {
			nodePath = getSimpleType(vnode)
		} else {
			nodePath = fmt.Sprintf("%s[%d].%s", path, index, getSimpleType(vnode))
		}
	}

	// Extract element info
	info := ExtractElementInfo(vnode)

	// Generate UniqueID following React's philosophy:
	// 1. User-defined key (preferred, like React's key prop)
	// 2. Path + index (stable across rebuilds, like React's index fallback)
	//
	// This matches React's approach where user keys are preferred,
	// and index is used as fallback (index is stable as long as structure doesn't change)
	var uniqueID string

	// Priority 1: User-defined key (most stable, matches React's key prop)
	if keyer, ok := vnode.(interface{ Key() string }); ok {
		if key := keyer.Key(); key != "" {
			uniqueID = fmt.Sprintf("%s[%s]", nodePath, key)
		}
	}

	// Priority 2: Path + index (stable across rebuilds, like React's array index)
	// This is stable as long as the structure (order of children) doesn't change
	// Example: "vstack.text[0]"
	if uniqueID == "" {
		uniqueID = fmt.Sprintf("%s[%d]", nodePath, index)
	}

	// Create tree node
	// Expand by default if no explicit state is set
	expanded, ok := tv.expanded[uniqueID]
	if !ok {
		expanded = true // Expand by default
	}

	node := &TreeNode{
		VNode:    vnode,
		Info:     info,
		Parent:   parent,
		Level:    level,
		Path:     nodePath,
		Expanded: expanded,
		UniqueID: uniqueID,
		Index:    index,
	}

	// Recursively build children
	children := vnode.Children()
	node.Children = make([]*TreeNode, 0, len(children))

	for i, child := range children {
		childNode := tv.buildTree(child, node, level+1, nodePath, i)
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node
}

// FormatTree formats the entire tree as text
func (tv *TreeView) FormatTree() string {
	if tv.root == nil {
		return "No tree to display"
	}

	var lines []string
	lines = append(lines, "┌─ Layout Tree ─────────────────────────────────")

	lines = tv.formatNode(tv.root, lines, true)

	lines = append(lines, "└─────────────────────────────────────────────┘")

	return strings.Join(lines, "\n")
}

// FormatTreePaginated formats the tree with pagination support
// Returns all lines, total count, and allows showing only a range
func (tv *TreeView) FormatTreePaginated() ([]string, int) {
	if tv.root == nil {
		return []string{"No tree to display"}, 1
	}

	var lines []string
	lines = append(lines, "┌─ Layout Tree ─────────────────────────────────")

	lines = tv.formatNode(tv.root, lines, true)

	lines = append(lines, "└─────────────────────────────────────────────┘")

	return lines, len(lines)
}

// GetTreeLines returns all lines and total count for scrolling
func (tv *TreeView) GetTreeLines() ([]string, int) {
	return tv.FormatTreePaginated()
}

// GetUniqueIDForLineIndex finds the unique ID for a given line index by traversing the tree
// This is used when expand/collapse changes to find which node to toggle
func (tv *TreeView) GetUniqueIDForLineIndex(targetIndex int) string {
	if tv.root == nil {
		return ""
	}
	currentIndex := 0
	return tv.findUniqueIDByIndex(tv.root, targetIndex, &currentIndex)
}

// findUniqueIDByIndex recursively searches for the unique ID at a given line index
func (tv *TreeView) findUniqueIDByIndex(node *TreeNode, targetIndex int, currentIndex *int) string {
	if node == nil {
		return ""
	}

	// Check if this is the target index (this node counts as 1 line)
	if *currentIndex == targetIndex {
		return node.UniqueID
	}
	*currentIndex++

	// If expanded, search children
	if node.Expanded && len(node.Children) > 0 {
		for _, child := range node.Children {
			uid := tv.findUniqueIDByIndex(child, targetIndex, currentIndex)
			if uid != "" {
				return uid
			}
		}
	} else if len(node.Children) > 0 {
		// Collapsed node with children - count the collapsed indicator line
		// This line also toggles the parent
		if *currentIndex == targetIndex {
			return node.UniqueID
		}
		*currentIndex++
	}

	return ""
}

// TreeLine represents a single line with its associated path
type TreeLine struct {
	Line string
	Path string
}

// formatNode recursively formats a tree node and returns the updated lines
func (tv *TreeView) formatNode(node *TreeNode, lines []string, isLast bool) []string {
	if node == nil {
		return lines
	}

	// Check depth limit
	if node.Level > tv.maxDepth {
		return lines
	}

	// Check if this node is selected
	isSelected := tv.selectedUniqueID != "" && node.UniqueID == tv.selectedUniqueID

	// Build prefix
	prefix := strings.Repeat("│  ", node.Level)
	var connector string
	if isSelected {
		// Show selection indicator
		connector = "▶ "
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	// Build node label
	icon := ""
	if tv.showIcons {
		icon = getIconForType(node.Info.Type)
	}

	// Add space after icon if it exists
	if icon != "" {
		icon += " "
	}

	label := fmt.Sprintf("%s%s", icon, node.Info.Type)
	if node.Info.Label != "" {
		label += fmt.Sprintf("(%s)", node.Info.Label)
	}

	// Add size info
	sizeInfo := ""
	if node.Info.Size.Width > 0 || node.Info.Size.Height > 0 {
		sizeInfo = fmt.Sprintf(" %dx%d", node.Info.Size.Width, node.Info.Size.Height)
	}

	// Add Key info (show path-based keys from Fiber reconciliation)
	keyInfo := ""
	if node.Info.Key != "" {
		// Format key differently based on whether it's user-provided or auto-generated
		if strings.HasPrefix(node.Info.Key, "/root/") {
			// Auto-generated path key - show shorter version
			keyInfo = fmt.Sprintf(" key:%s", node.Info.Key)
		} else {
			// User-provided key
			keyInfo = fmt.Sprintf(" key:'%s'", node.Info.Key)
		}
	}

	// Add path if enabled (for backward compatibility)
	pathInfo := ""
	if tv.showPaths && node.Path != "" {
		pathInfo = fmt.Sprintf(" [%s]", node.Path)
	}

	line := prefix + connector + label + sizeInfo + keyInfo + pathInfo
	lines = append(lines, line)

	// Format children if expanded
	if node.Expanded && len(node.Children) > 0 {
		for i, child := range node.Children {
			lines = tv.formatNode(child, lines, i == len(node.Children)-1)
		}
	} else if len(node.Children) > 0 {
		// Show collapsed indicator
		childCount := len(node.Children)
		collapsedLine := prefix
		if isLast {
			collapsedLine += "    "
		} else {
			collapsedLine += "│   "
		}
		collapsedLine += fmt.Sprintf("└── (+ %d children)", childCount)
		lines = append(lines, collapsedLine)
	}

	return lines
}

// ToggleNode toggles a node's expanded state
func (tv *TreeView) ToggleNode(uniqueID string) {
	// First, find the node to get its current expansion state
	currentState := false // Default to collapsed
	node := tv.findNodeByUniqueID(tv.root, uniqueID)
	if node != nil {
		currentState = node.Expanded
	}

	// Toggle the state
	newState := !currentState
	tv.expanded[uniqueID] = newState

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		log.UILogger.Debug("[TreeView] ToggleNode: uniqueID=%s, %v -> %v\n", uniqueID, currentState, newState)
	}

	// Update tree if we have a root
	if tv.root != nil {
		tv.updateNodeExpansion(tv.root, uniqueID)
		tv.changeCount++
	}
}

// findNodeByUniqueID finds a tree node by its unique ID
func (tv *TreeView) findNodeByUniqueID(node *TreeNode, uniqueID string) *TreeNode {
	if node == nil {
		return nil
	}

	if node.UniqueID == uniqueID {
		return node
	}

	for _, child := range node.Children {
		if found := tv.findNodeByUniqueID(child, uniqueID); found != nil {
			return found
		}
	}

	return nil
}

// updateNodeExpansion recursively updates expansion state
func (tv *TreeView) updateNodeExpansion(node *TreeNode, uniqueID string) bool {
	if node == nil {
		return false
	}

	if node.UniqueID == uniqueID {
		node.Expanded = tv.expanded[uniqueID]
		return true
	}

	for _, child := range node.Children {
		if tv.updateNodeExpansion(child, uniqueID) {
			return true
		}
	}

	return false
}

// ExpandAll expands all nodes in the tree
func (tv *TreeView) ExpandAll() {
	if tv.root == nil {
		return
	}

	tv.expandAllRecursive(tv.root)
	tv.changeCount++
}

// expandAllRecursive recursively expands all nodes
func (tv *TreeView) expandAllRecursive(node *TreeNode) {
	if node == nil {
		return
	}

	node.Expanded = true
	tv.expanded[node.Path] = true

	for _, child := range node.Children {
		tv.expandAllRecursive(child)
	}
}

// CollapseAll collapses all nodes in the tree
func (tv *TreeView) CollapseAll() {
	if tv.root == nil {
		return
	}

	tv.collapseAllRecursive(tv.root)
	tv.changeCount++
}

// collapseAllRecursive recursively collapses all nodes
func (tv *TreeView) collapseAllRecursive(node *TreeNode) {
	if node == nil {
		return
	}

	node.Expanded = false
	tv.expanded[node.Path] = false

	for _, child := range node.Children {
		tv.collapseAllRecursive(child)
	}
}

// FindNodeByPath finds a node by its path
func (tv *TreeView) FindNodeByPath(path string) *TreeNode {
	if tv.root == nil {
		return nil
	}

	return tv.findNodeByPathRecursive(tv.root, path)
}

// findNodeByPathRecursive recursively searches for a node by path
func (tv *TreeView) findNodeByPathRecursive(node *TreeNode, path string) *TreeNode {
	if node == nil {
		return nil
	}

	if node.Path == path {
		return node
	}

	for _, child := range node.Children {
		result := tv.findNodeByPathRecursive(child, path)
		if result != nil {
			return result
		}
	}

	return nil
}

// FindNodesByType finds all nodes of a specific type
func (tv *TreeView) FindNodesByType(nodeType string) []*TreeNode {
	var results []*TreeNode

	if tv.root == nil {
		return results
	}

	tv.findNodesByTypeRecursive(tv.root, nodeType, &results)
	return results
}

// findNodesByTypeRecursive recursively finds nodes by type
func (tv *TreeView) findNodesByTypeRecursive(node *TreeNode, nodeType string, results *[]*TreeNode) {
	if node == nil {
		return
	}

	if strings.Contains(node.Info.Type, nodeType) {
		*results = append(*results, node)
	}

	for _, child := range node.Children {
		tv.findNodesByTypeRecursive(child, nodeType, results)
	}
}

// FindNodesByLabel finds nodes with matching label
func (tv *TreeView) FindNodesByLabel(label string) []*TreeNode {
	var results []*TreeNode

	if tv.root == nil {
		return results
	}

	tv.findNodesByLabelRecursive(tv.root, label, &results)
	return results
}

// findNodesByLabelRecursive recursively finds nodes by label
func (tv *TreeView) findNodesByLabelRecursive(node *TreeNode, label string, results *[]*TreeNode) {
	if node == nil {
		return
	}

	if strings.Contains(strings.ToLower(node.Info.Label), strings.ToLower(label)) {
		*results = append(*results, node)
	}

	for _, child := range node.Children {
		tv.findNodesByLabelRecursive(child, label, results)
	}
}

// GetTreeStats returns statistics about the tree
func (tv *TreeView) GetTreeStats() TreeStats {
	stats := TreeStats{}

	if tv.root == nil {
		return stats
	}

	tv.calculateStats(tv.root, &stats)
	return stats
}

// calculateStats recursively calculates tree statistics
func (tv *TreeView) calculateStats(node *TreeNode, stats *TreeStats) {
	if node == nil {
		return
	}

	stats.TotalNodes++
	stats.MaxDepth = max(stats.MaxDepth, node.Level)

	if len(node.Children) == 0 {
		stats.LeafNodes++
	} else {
		stats.ParentNodes++
		for _, child := range node.Children {
			tv.calculateStats(child, stats)
		}
	}
}

// GetFlatList returns a flat list of all tree nodes
func (tv *TreeView) GetFlatList() []*TreeNode {
	var nodes []*TreeNode

	if tv.root == nil {
		return nodes
	}

	tv.flattenRecursive(tv.root, &nodes)
	return nodes
}

// flattenRecursive recursively flattens the tree
func (tv *TreeView) flattenRecursive(node *TreeNode, nodes *[]*TreeNode) {
	if node == nil {
		return
	}

	*nodes = append(*nodes, node)

	for _, child := range node.Children {
		tv.flattenRecursive(child, nodes)
	}
}

// SetShowIcons controls whether to show type icons
func (tv *TreeView) SetShowIcons(show bool) {
	tv.showIcons = show
}

// SetShowPaths controls whether to show paths
func (tv *TreeView) SetShowPaths(show bool) {
	tv.showPaths = show
}

// SetCompact controls compact display mode
func (tv *TreeView) SetCompact(compact bool) {
	tv.compact = compact
}

// SetMaxDepth sets the maximum traversal depth
func (tv *TreeView) SetMaxDepth(depth int) {
	tv.maxDepth = depth
}

// SetMaxNodes sets the maximum nodes to display
func (tv *TreeView) SetMaxNodes(max int) {
	tv.maxNodes = max
}

// SetSelectedNode sets the currently selected node by unique ID
func (tv *TreeView) SetSelectedNode(uniqueID string) {
	tv.selectedUniqueID = uniqueID
}

// GetSelectedNode returns the unique ID of the currently selected node
func (tv *TreeView) GetSelectedNode() string {
	return tv.selectedUniqueID
}

// GetSelectedNodeInfo returns information about the selected node
func (tv *TreeView) GetSelectedNodeInfo() *TreeNode {
	if tv.selectedUniqueID == "" {
		return nil
	}
	return tv.findNodeByUniqueID(tv.root, tv.selectedUniqueID)
}

// GetChangeCount returns the current change counter
func (tv *TreeView) GetChangeCount() int64 {
	return tv.changeCount
}

// TreeStats represents tree statistics
type TreeStats struct {
	TotalNodes  int
	LeafNodes   int
	ParentNodes int
	MaxDepth    int
}

// Helper functions

// getSimpleType returns a simple type name for the VNode
func getSimpleType(vnode ui.VNode) string {
	if vnode == nil {
		return "unknown"
	}

	// Try tag first
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		tag := tagger.Tag()
		if tag != "" {
			return tag
		}
	}

	// Fallback to type name
	info := ExtractElementInfo(vnode)
	return info.Type
}

// getIconForType returns an icon for a given element type
// NOTE: Using single-rune icons only to avoid VS16/ZWJ issues in TUI
// See: docs/TUI_BUFFER_SPECIAL_CHARS_ISSUE.md and TUI_BUFFER_FIX.md
func getIconForType(typeName string) string {
	typeLower := strings.ToLower(typeName)

	switch {
	case strings.Contains(typeLower, "button"):
		return "🔵"
	case strings.Contains(typeLower, "text"):
		return "📝"
	case strings.Contains(typeLower, "hstack"):
		return "→"
	case strings.Contains(typeLower, "vstack"):
		return "↓"
	case strings.Contains(typeLower, "box"):
		return "📦"
	case strings.Contains(typeLower, "border"):
		return "🎨" // Use art icon instead of multi-rune "🖼️"
	case strings.Contains(typeLower, "input"):
		return "✏" // Use single-rune pencil
	case strings.Contains(typeLower, "checkbox"):
		return "☑" // Use single-rune checkbox
	case strings.Contains(typeLower, "select"):
		return "📋"
	default:
		return "📦"
	}
}
