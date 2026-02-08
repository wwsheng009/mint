package display

import (
	"fmt"
	"strings"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TreeView is a component for displaying tree structures with proper formatting
// Unlike the simple text-based approach, this component maintains tree structure
// with proper indentation, line breaks, and visual hierarchy
type TreeView struct {
	*ui.ElementVNode
	lines       []TreeViewLine  // Pre-rendered tree lines
	focusIndex  int              // Currently focused line index
	totalLines  int              // Total number of lines
	expandState map[int]bool     // Expand/collapse state for each node
	selectedIdx int              // Currently selected line index

	// Navigation state
	scrollOffset int             // Current scroll offset
	viewportHeight int           // Visible height for scrolling
	builder     *TreeViewBuilder // Reference to builder for rebuilds
}

// TreeViewLine represents a single line in the tree view
type TreeViewLine struct {
	Indent   int    // Indentation level (0-based)
	Content  string // The line content without prefix
	Prefix   string // Tree prefix (├──, └──, │  , etc.)
	NodeType  string // Node type for icon display
	NodeID    int    // Optional: node ID for selection
}

// TreeViewBuilder builds tree view components
type TreeViewBuilder struct {
	node         *TreeView
	sourceLines  []string // Raw tree lines
	rootNode     rtui.VNode // Optional: root node for auto-generation
	expandLevel  int      // Default expand level (0 = all collapsed, -1 = all expanded)
	showIcons    bool    // Show type icons
	showLineNums  bool    // Show line numbers
	compact      bool    // Use compact display
}

// NewTreeView creates a new tree view builder
func NewTreeView() *TreeViewBuilder {
	return &TreeViewBuilder{
		expandLevel: 1, // Default: expand first level
		showIcons:   true,
		showLineNums: false,
		compact:     false,
		node: &TreeView{
			ElementVNode: ui.NewElement("treeview"),
			lines:        []TreeViewLine{},
			focusIndex:   0,
			totalLines:   0,
			expandState:   make(map[int]bool),
			selectedIdx:  -1,
		},
	}
}

// FromLines creates a tree view from pre-formatted lines
func (b *TreeViewBuilder) FromLines(lines []string) *TreeViewBuilder {
	b.sourceLines = lines
	return b
}

// FromNode creates a tree view from a VNode (auto-generates tree structure)
func (b *TreeViewBuilder) FromNode(root ui.VNode) *TreeViewBuilder {
	b.rootNode = root
	return b
}

// ExpandLevel sets the default expand level
func (b *TreeViewBuilder) ExpandLevel(level int) *TreeViewBuilder {
	b.expandLevel = level
	return b
}

// ShowIcons enables/disables type icons
func (b *TreeViewBuilder) ShowIcons(show bool) *TreeViewBuilder {
	b.showIcons = show
	return b
}

// ShowLineNumbers enables/disables line numbers
func (b *TreeViewBuilder) ShowLineNumbers(show bool) *TreeViewBuilder {
	b.showLineNums = show
	return b
}

// Compact sets compact display mode
func (b *TreeViewBuilder) Compact(compact bool) *TreeViewBuilder {
	b.compact = compact
	return b
}

// Build renders the tree view into VNodes
func (b *TreeViewBuilder) Build() ui.VNode {
	if b.rootNode != nil {
		// Auto-generate tree structure from VNode
		b.generateFromNode(b.rootNode, 0, true)
	} else if len(b.sourceLines) > 0 {
		// Parse pre-formatted lines
		b.parseLines(b.sourceLines)
	}

	b.node.totalLines = len(b.node.lines)
	// Store builder reference for rebuilds
	b.node.builder = b

	// Create text nodes for each line
	var lineNodes []ui.VNode
	for i, line := range b.node.lines {
		// Build the full line with prefix and content
		fullLine := strings.Repeat(" ", line.Indent) + line.Prefix + line.Content
		if b.showIcons && line.NodeType != "" {
			fullLine = b.addIcon(fullLine, line.NodeType)
		}
		if b.showLineNums {
			fullLine = b.formatLineNumber(i+1) + " " + fullLine
		}

		// Apply style based on focus/selection state
		if i == b.node.selectedIdx {
			// Selected line - reverse video
			lineNodes = append(lineNodes, ui.Text(fmt.Sprintf("[reverse]%s[/reverse]", fullLine)))
		} else if i == b.node.focusIndex {
			// Focused line - bright/bold
			lineNodes = append(lineNodes, ui.Text(fmt.Sprintf("[bold]%s[/bold]", fullLine)))
		} else {
			// Normal line
			lineNodes = append(lineNodes, ui.Text(fullLine))
		}
	}

	// Wrap in VStack for proper rendering
	result := ui.VStack(lineNodes...)
	b.node.SetChildren([]ui.VNode{result})

	return b.node
}

// generateFromNode recursively generates tree lines from VNode
func (b *TreeViewBuilder) generateFromNode(node rtui.VNode, depth int, isLast bool) {
	if node == nil {
		return
	}

	// Get node description
	description := b.getNodeDescription(node)

	// Build tree prefix
	prefix := "├── "
	if isLast {
		prefix = "└── "
	}

	// Add line
	treeLine := TreeViewLine{
		Indent:   depth * 2, // 2 spaces per level
		Content:  description,
		Prefix:   prefix,
		NodeType: fmt.Sprintf("%d", node.Type()), // Convert int type to string for icon lookup
		NodeID:   len(b.node.lines),
	}

	b.node.lines = append(b.node.lines, treeLine)

	// Recursively process children
	children := node.Children()
	if len(children) > 0 && (b.expandLevel < 0 || depth < b.expandLevel) {
		for i, child := range children {
			isLastChild := (i == len(children)-1)
			b.generateFromNode(child, depth+1, isLastChild)
		}
	}
}

// parseLines parses pre-formatted tree lines
func (b *TreeViewBuilder) parseLines(lines []string) {
	for i, line := range lines {
		// Calculate indentation
		indent := 0
		for j := 0; j < len(line); j++ {
			if line[j] == ' ' {
				indent++
			} else {
				break
			}
		}

		// Extract prefix (tree structure characters)
		prefix := ""
		remaining := strings.TrimLeft(line, " ")
		for _, ch := range remaining {
			if ch == '│' || ch == '├' || ch == '└' || ch == '─' || ch == '┌' || ch == '┐' {
			prefix += string(ch)
			indent++ // Count prefix chars for indentation
		} else {
			break
			}
		}

		// Content is the rest after prefix
		content := strings.TrimPrefix(remaining, prefix)
		content = strings.TrimLeft(content, " ")

		b.node.lines = append(b.node.lines, TreeViewLine{
			Indent:  indent,
			Content: content,
			Prefix:   prefix,
			NodeType: "",
			NodeID:   i,
		})
	}
}

// getNodeIcon returns an icon for the node type
func (b *TreeViewBuilder) getNodeIcon(nodeType string) string {
	if !b.showIcons {
		return ""
	}

	// Map node types to icons
	icons := map[string]string{
		"vstack":    "📦",
		"hstack":    "📦",
		"text":       "📝",
		"button":     "🔵",
		"bordered":   "🖼️",
		"flex":       "🔧",
		"element":    "📦",
		"component":  "⚙️",
	}

	typeStr := string(nodeType)
	typeStr = strings.ToLower(typeStr)

	if icon, ok := icons[typeStr]; ok {
		return icon + " "
	}

	return ""
}

// getNodeDescription generates a description for the node
func (b *TreeViewBuilder) getNodeDescription(node ui.VNode) string {
	if node == nil {
		return "nil"
	}

	// Try to get text content
	if textNode, ok := node.(interface{ Content() string }); ok {
		content := textNode.Content()
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		return content
	}

	// Use node type
	return string(node.Type())
}

// formatLineNumber formats a line number
func (b *TreeViewBuilder) formatLineNumber(num int) string {
	return fmt.Sprintf("%3d", num)
}

// getLineStyle returns the style for a line
func (b *TreeViewBuilder) getLineStyle(index int) string {
	styles := map[string]string{
		"fg":      "white",
			"bg":      "",
		"bold":    "false",
		"reverse": "false",
	}

	// Highlight selected line
	if index == b.node.selectedIdx {
		styles["bg"] = "blue"
		styles["reverse"] = "true"
	}

	// Highlight focused line
	if index == b.node.focusIndex {
		styles["fg"] = "yellow"
		styles["bold"] = "true"
	}

	return fmt.Sprintf("fg:%s,bg:%s,bold:%s,reverse:%s",
		styles["fg"], styles["bg"], styles["bold"], styles["reverse"])
}

// addIcon adds an icon to a line
func (b *TreeViewBuilder) addIcon(line string, nodeType string) string {
	icon := b.getNodeIcon(nodeType)
	if icon != "" {
		return icon + " " + line
	}
	return line
}

// GetTreeLines returns the tree as formatted lines (for compatibility)
func (t *TreeView) GetTreeLines() ([]string, int) {
	if len(t.lines) == 0 {
		return []string{}, 0
	}

	lines := make([]string, len(t.lines))
	for i, treeLine := range t.lines {
		lines[i] = strings.Repeat(" ", treeLine.Indent) + treeLine.Prefix + treeLine.Content
		// Note: Icons not added in compatibility mode
	}

	return lines, len(lines)
}

// ScrollBy scrolls the tree view by delta lines
func (t *TreeView) ScrollBy(delta int) int {
	newIndex := t.focusIndex + delta

	// Clamp to valid range
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(t.lines) {
		newIndex = len(t.lines) - 1
	}

	t.focusIndex = newIndex
	return newIndex
}

// ScrollTo scrolls to an absolute line
func (t *TreeView) ScrollTo(index int) int {
	if index < 0 {
		index = 0
	}
	if index >= len(t.lines) {
		index = len(t.lines) - 1
	}

	t.focusIndex = index
	return index
}

// SelectLine selects a line (returns its index)
func (t *TreeView) SelectLine(lineNum int) int {
	if lineNum < 0 || lineNum >= len(t.lines) {
		return -1
	}

	t.selectedIdx = lineNum
	return lineNum
}

// GetSelectedLine returns the currently selected line
func (t *TreeView) GetSelectedLine() TreeViewLine {
	if t.selectedIdx < 0 || t.selectedIdx >= len(t.lines) {
		return TreeViewLine{}
	}

	return t.lines[t.selectedIdx]
}

// FocusLine moves focus to a line
func (t *TreeView) FocusLine(lineNum int) int {
	if lineNum < 0 || lineNum >= len(t.lines) {
		return -1
	}

	oldFocus := t.focusIndex
	t.focusIndex = lineNum
	t.ensureVisible()

	// Auto-expand the node if needed
	t.expandState[lineNum] = true

	return oldFocus
}

// CanScrollUp checks if can scroll up
func (t *TreeView) CanScrollUp() bool {
	return t.focusIndex > 0
}

// CanScrollDown checks if can scroll down
func (t *TreeView) CanScrollDown() bool {
	return t.focusIndex < len(t.lines)-1
}

// ToggleExpand toggles expand/collapse for a line
func (t *TreeView) ToggleExpand(lineNum int) bool {
	if lineNum < 0 || lineNum >= len(t.lines) {
		return false
	}

	// Toggle expand state
	if expanded, ok := t.expandState[lineNum]; ok {
		t.expandState[lineNum] = !expanded
		return !expanded
	}

	t.expandState[lineNum] = true
	return true
}

// GetNodeCount returns the total number of nodes
func (t *TreeView) GetNodeCount() int {
	return len(t.lines)
}

// GetLineCount returns the total number of lines
func (t *TreeView) GetLineCount() int {
	return t.totalLines
}

// =============================================================================
// Navigation Methods
// =============================================================================

// HandleKey handles keyboard events for tree navigation
// Returns true if the key was handled, false otherwise
func (t *TreeView) HandleKey(key platform.SpecialKey, r rune) bool {
	switch key {
	case platform.KeyUp:
		t.MoveUp()
		return true
	case platform.KeyDown:
		t.MoveDown()
		return true
	case platform.KeyPageUp:
		t.PageUp()
		return true
	case platform.KeyPageDown:
		t.PageDown()
		return true
	case platform.KeyHome:
		t.Home()
		return true
	case platform.KeyEnd:
		t.End()
		return true
	case platform.KeyEnter:
		t.SelectCurrent()
		return true
	}

	// Handle character keys
	switch r {
	case 'e', 'E':
		t.ToggleExpandCurrent()
		return true
	}

	return false
}

// MoveUp moves focus to the previous line
func (t *TreeView) MoveUp() {
	if t.focusIndex > 0 {
		t.focusIndex--
		t.ensureVisible()
	}
}

// MoveDown moves focus to the next line
func (t *TreeView) MoveDown() {
	if t.focusIndex < len(t.lines)-1 {
		t.focusIndex++
		t.ensureVisible()
	}
}

// PageUp moves focus up by one page
func (t *TreeView) PageUp() {
	pageSize := t.viewportHeight
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	t.focusIndex -= pageSize
	if t.focusIndex < 0 {
		t.focusIndex = 0
	}
	t.ensureVisible()
}

// PageDown moves focus down by one page
func (t *TreeView) PageDown() {
	pageSize := t.viewportHeight
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	t.focusIndex += pageSize
	if t.focusIndex >= len(t.lines) {
		t.focusIndex = len(t.lines) - 1
	}
	t.ensureVisible()
}

// Home moves focus to the first line
func (t *TreeView) Home() {
	t.focusIndex = 0
	t.scrollOffset = 0
}

// End moves focus to the last line
func (t *TreeView) End() {
	t.focusIndex = len(t.lines) - 1
	t.ensureVisible()
}

// ensureVisible ensures the focused line is visible in the viewport
func (t *TreeView) ensureVisible() {
	if t.viewportHeight <= 0 {
		return
	}

	// Scroll down if focus is below viewport
	if t.focusIndex >= t.scrollOffset+t.viewportHeight {
		t.scrollOffset = t.focusIndex - t.viewportHeight + 1
	}

	// Scroll up if focus is above viewport
	if t.focusIndex < t.scrollOffset {
		t.scrollOffset = t.focusIndex
	}
}

// ToggleExpandCurrent toggles expand/collapse for the focused line
func (t *TreeView) ToggleExpandCurrent() {
	if t.focusIndex >= 0 && t.focusIndex < len(t.lines) {
		line := &t.lines[t.focusIndex]
		// Toggle expand state
		if expanded, ok := t.expandState[line.NodeID]; ok {
			t.expandState[line.NodeID] = !expanded
		} else {
			t.expandState[line.NodeID] = true
		}
	}
}

// SelectCurrent selects the currently focused line
func (t *TreeView) SelectCurrent() {
	if t.focusIndex >= 0 && t.focusIndex < len(t.lines) {
		t.selectedIdx = t.focusIndex
	}
}

// GetFocusedLine returns the currently focused line
func (t *TreeView) GetFocusedLine() TreeViewLine {
	if t.focusIndex >= 0 && t.focusIndex < len(t.lines) {
		return t.lines[t.focusIndex]
	}
	return TreeViewLine{}
}

// SetViewportHeight sets the viewport height for scrolling calculations
func (t *TreeView) SetViewportHeight(height int) {
	t.viewportHeight = height
}

// GetScrollOffset returns the current scroll offset
func (t *TreeView) GetScrollOffset() int {
	return t.scrollOffset
}

// SetScrollOffset sets the scroll offset
func (t *TreeView) SetScrollOffset(offset int) {
	maxOffset := len(t.lines) - t.viewportHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	t.scrollOffset = offset
}

// Rebuild rebuilds the tree view with current state
func (t *TreeView) Rebuild() ui.VNode {
	if t.builder == nil {
		return t
	}

	// Re-run Build with current state
	return t.builder.Build()
}

// GetFocusIndex returns the current focus index
func (t *TreeView) GetFocusIndex() int {
	return t.focusIndex
}

// SetFocusIndex sets the focus index
func (t *TreeView) SetFocusIndex(index int) {
	if index >= 0 && index < len(t.lines) {
		t.focusIndex = index
		t.ensureVisible()
	}
}

// IsExpanded returns true if a node is expanded
func (t *TreeView) IsExpanded(nodeID int) bool {
	if expanded, ok := t.expandState[nodeID]; ok {
		return expanded
	}
	return false // Default collapsed
}

// SetExpanded sets the expand state for a node
func (t *TreeView) SetExpanded(nodeID int, expanded bool) {
	t.expandState[nodeID] = expanded
}

// ClearSelection clears the current selection
func (t *TreeView) ClearSelection() {
	t.selectedIdx = -1
}

// HasSelection returns true if there's a selection
func (t *TreeView) HasSelection() bool {
	return t.selectedIdx >= 0 && t.selectedIdx < len(t.lines)
}
