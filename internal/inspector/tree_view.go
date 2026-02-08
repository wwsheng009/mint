package inspector

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/ui"
)

// TreeNode represents a node in the layout tree
type TreeNode struct {
	VNode     ui.VNode
	Info      ElementInfo
	Children  []*TreeNode
	Parent    *TreeNode
	Expanded  bool
	Level     int  // Depth in tree (0 = root)
	Index     int  // Index among siblings
	Path      string // Full path from root
}

// TreeView provides tree visualization and traversal
type TreeView struct {
	root        *TreeNode
	expanded    map[string]bool // Track expanded nodes by path
	showIcons   bool            // Show type icons
	showPaths   bool            // Show paths
	compact     bool            // Use compact display
	maxDepth    int             // Maximum traversal depth
	maxNodes   int             // Maximum nodes to display
}

// NewTreeView creates a new TreeView instance
func NewTreeView() *TreeView {
	return &TreeView{
		expanded:  make(map[string]bool),
		showIcons: true,
		showPaths: false,
		compact: false,
		maxDepth: 100, // Effectively unlimited
		maxNodes: 1000,
	}
}

// SetRoot sets the root VNode for the tree
func (tv *TreeView) SetRoot(root ui.VNode) error {
	if root == nil {
		tv.root = nil
		return nil
	}

	// Build tree structure
	tv.root = tv.buildTree(root, nil, 0, "")
	return nil
}

// buildTree recursively builds the tree structure
func (tv *TreeView) buildTree(vnode ui.VNode, parent *TreeNode, level int, path string) *TreeNode {
	if vnode == nil {
		return nil
	}

	// Generate path for this node
	var nodePath string
	if path == "" {
		nodePath = getSimpleType(vnode)
	} else {
		nodePath = path + "." + getSimpleType(vnode)
	}

	// Extract element info
	info := ExtractElementInfo(vnode)

	// Create tree node
	// Expand by default if no explicit state is set
	expanded, ok := tv.expanded[nodePath]
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
	}

	// Recursively build children
	children := vnode.Children()
	node.Children = make([]*TreeNode, 0, len(children))

	for i, child := range children {
		childNode := tv.buildTree(child, node, level+1, nodePath)
		if childNode != nil {
			childNode.Index = i
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

// formatNode recursively formats a tree node and returns the updated lines
func (tv *TreeView) formatNode(node *TreeNode, lines []string, isLast bool) []string {
	if node == nil {
		return lines
	}

	// Check depth limit
	if node.Level > tv.maxDepth {
		return lines
	}

	// Build prefix
	prefix := strings.Repeat("│  ", node.Level)
	var connector string
	if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	// Build node label
	icon := ""
	if tv.showIcons {
		icon = getIconForType(node.Info.Type)
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

	// Add path if enabled
	pathInfo := ""
	if tv.showPaths && node.Path != "" {
		pathInfo = fmt.Sprintf(" [%s]", node.Path)
	}

	line := prefix + connector + label + sizeInfo + pathInfo
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
func (tv *TreeView) ToggleNode(path string) {
	// If key doesn't exist, it means it's expanded by default (true)
	// So we need to toggle it to false
	currentState, exists := tv.expanded[path]
	if !exists {
		currentState = true // Default expanded state
	}
	tv.expanded[path] = !currentState

	// Update tree if we have a root
	if tv.root != nil {
		tv.updateNodeExpansion(tv.root, path)
	}
}

// updateNodeExpansion recursively updates expansion state
func (tv *TreeView) updateNodeExpansion(node *TreeNode, path string) bool {
	if node == nil {
		return false
	}

	if node.Path == path {
		node.Expanded = tv.expanded[path]
		return true
	}

	for _, child := range node.Children {
		if tv.updateNodeExpansion(child, path) {
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

// TreeStats represents tree statistics
type TreeStats struct {
	TotalNodes int
	LeafNodes  int
	ParentNodes int
	MaxDepth   int
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
		return "🖼️"
	case strings.Contains(typeLower, "input"):
		return "✏️"
	case strings.Contains(typeLower, "checkbox"):
		return "☑️"
	case strings.Contains(typeLower, "select"):
		return "📋"
	default:
		return "📦"
	}
}
