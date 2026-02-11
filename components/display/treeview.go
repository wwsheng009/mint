package display

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/cmd"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/platform"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

// Interface implementation assertions
var _ runtime.Measurable = (*TreeView)(nil)
var _ frameworkevent.Component = (*TreeView)(nil)
var _ component.Updater = (*TreeView)(nil) // Phase 3: Msg/Cmd support
var _ action.ActionTarget = (*TreeView)(nil)
var _ action.FocusableActionTarget = (*TreeView)(nil)
var _ action.ScrollableActionTarget = (*TreeView)(nil)
var _ action.SelectableActionTarget = (*TreeView)(nil)
// Note: TreeView cannot implement ExpandableActionTarget due to IsExpanded() vs IsExpanded(int) conflict

// TreeView is a component for displaying tree structures with proper formatting

// Unlike the simple text-based approach, this component maintains tree structure
// with proper indentation, line breaks, and visual hierarchy
type TreeView struct {
	*ui.ElementVNode
	lines       []TreeViewLine // Pre-rendered tree lines
	focusIndex  int            // Currently focused line index
	totalLines  int            // Total number of lines
	expandState map[int]bool   // Expand/collapse state for each node
	selectedIdx int            // Currently selected line index

	// Navigation state
	scrollOffset         int              // Current scroll offset
	viewportHeight       int              // Visible height for scrolling
	builder              *TreeViewBuilder // Reference to builder for rebuilds
	currentRender        ui.VNode         // Latest render result (for GetRender())
	expandStateChanged   bool             // True when expand/collapse state changed
	expandStateLineIndex int              // Which line index needs expand toggle
	lastMeasuredWidth    int              // Last measured width (for constraining content)

	// Layout bounds (for mouse hit testing)
	boundsX int
	boundsY int
	boundsW int
	boundsH int

	// ActionTarget support
	supportedActions []action.ActionType // Supported action types
}

// TreeViewLine represents a single line in the tree view
type TreeViewLine struct {
	Indent   int    // Byte offset where content starts (or space indentation)
	Content  string // The actual content (node name, etc.)
	RawLine  string // Original full line (for accurate rendering)
	Prefix   string // Tree prefix (├──, └──, │  , etc.) - deprecated
	NodeType string // Node type for icon display
	NodeID   int    // Optional: node ID for selection
	Path     string // Unique path identifier for this node (stable across expand/collapse)
}

// TreeViewBuilder builds tree view components
type TreeViewBuilder struct {
	node         *TreeView
	sourceLines  []string   // Raw tree lines
	rootNode     rtui.VNode // Optional: root node for auto-generation
	expandLevel  int        // Default expand level (0 = all collapsed, -1 = all expanded)
	showIcons    bool       // Show type icons
	showLineNums bool       // Show line numbers
	compact      bool       // Use compact display
}

// NewTreeView creates a new tree view builder
func NewTreeView() *TreeViewBuilder {
	return &TreeViewBuilder{
		expandLevel:  1, // Default: expand first level
		showIcons:    true,
		showLineNums: false,
		compact:      false,
		node: &TreeView{
			ElementVNode: ui.NewElement("treeview"),
			lines:         []TreeViewLine{},
			focusIndex:    0,
			totalLines:    0,
			expandState:   make(map[int]bool),
			selectedIdx:   -1,
			supportedActions: []action.ActionType{
				action.ActionNavigateUp,
				action.ActionNavigateDown,
				action.ActionNavigateLeft,
				action.ActionNavigateRight,
				action.ActionNavigatePageUp,
				action.ActionNavigatePageDown,
				action.ActionNavigateHome,
				action.ActionNavigateEnd,
				action.ActionSelect,
				action.ActionToggle,
				action.ActionExpand,
				action.ActionCollapse,
				action.ActionScroll,
				action.ActionClick,
			},
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

	// Call regenerateDisplay to handle virtual scrolling
	// This ensures only visible lines are rendered based on viewportHeight
	b.node.regenerateDisplay()

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
	// Clear existing lines before parsing to avoid duplicates
	b.node.lines = b.node.lines[:0]

	for i, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// SIMPLE APPROACH: Store the original line and try to identify where content starts
		// Find the position where actual content begins (after tree connectors)
		contentStart := 0
		seenContent := false

		// Convert to runes for proper Unicode character handling
		runes := []rune(line)
		for j, ch := range runes {
			// Tree connector characters (as runes)
			if ch == ' ' || ch == '│' || ch == '├' || ch == '└' || ch == '─' || ch == '┌' || ch == '┐' {
				continue
			}
			// Found content start (non-space, non-connector)
			contentStart = j
			seenContent = true
			break
		}

		var rawLine, content string
		if seenContent {
			rawLine = line
			content = string(runes[contentStart:])
		} else {
			// All connectors, no content
			rawLine = line
			content = ""
		}

		b.node.lines = append(b.node.lines, TreeViewLine{
			Indent:   contentStart, // Store as rune offset
			Content:  content,
			RawLine:  rawLine, // Store original line for rendering
			Prefix:   "",
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
	// NOTE: Using single-rune icons only to avoid VS16/ZWJ issues in TUI
	// See: docs/TUI_BUFFER_SPECIAL_CHARS_ISSUE.md and TUI_BUFFER_FIX.md
	icons := map[string]string{
		"vstack":    "📦",
		"hstack":    "📦",
		"text":      "📝",
		"button":    "🔵",
		"bordered":  "🎨", // Use art icon instead of multi-rune "🖼️"
		"flex":      "🔧",
		"element":   "📦",
		"component": "⚙", // Use single-rune gear
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

	// Use node type (convert int to string properly)
	return fmt.Sprintf("%d", node.Type())
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

// SetBounds stores layout bounds for mouse hit testing.
func (t *TreeView) SetBounds(x, y, width, height int) {
	t.boundsX = x
	t.boundsY = y
	t.boundsW = width
	t.boundsH = height
}

// GetBounds returns stored bounds (tuple form for hit testing).
func (t *TreeView) GetBounds() (int, int, int, int) {
	return t.boundsX, t.boundsY, t.boundsW, t.boundsH
}

// Bounds returns bounds as array (used by some inspector utilities).
func (t *TreeView) Bounds() [4]int {
	return [4]int{t.boundsX, t.boundsY, t.boundsW, t.boundsH}
}

// containsPoint reports whether screen coordinates are inside the tree view.
func (t *TreeView) containsPoint(x, y int) bool {
	return x >= t.boundsX && x < t.boundsX+t.boundsW &&
		y >= t.boundsY && y < t.boundsY+t.boundsH
}

// HandleEvent adds basic mouse interaction: click to focus/select, wheel to scroll.
func (t *TreeView) HandleEvent(ev frameworkevent.Event) bool {
	me, ok := ev.(*frameworkevent.MouseEvent)
	if !ok {
		return false
	}

	// Ignore events outside our bounds.
	if !t.containsPoint(me.X, me.Y) {
		return false
	}

	switch ev.Type() {
	case frameworkevent.EventMousePress:
		localY := me.Y - t.boundsY
		if localY < 0 {
			return false
		}

		target := t.scrollOffset + localY
		if target < 0 || target >= len(t.lines) {
			return false
		}

		t.SetFocusIndex(target)
		t.SelectLine(target)
		return true

	case frameworkevent.EventMouseWheel:
		// Wheel events lack direction info in current event model; ignore for now.
		return false

	case frameworkevent.EventMouseMove:
		// Hover focus (non-selecting) for visual cue.
		localY := me.Y - t.boundsY
		if localY < 0 {
			return false
		}
		target := t.scrollOffset + localY
		if target < 0 || target >= len(t.lines) {
			return false
		}
		if target != t.focusIndex {
			t.SetFocusIndex(target)
			return true
		}
	}

	return false
}

// =============================================================================
// Msg/Cmd Architecture Support (Phase 3)
// =============================================================================

// Update implements component.Updater interface for Msg/Cmd architecture
//
// Handles:
// - MouseMsg: Tree node selection, hover focus
// - KeyMsg: Keyboard navigation (when focused)
func (t *TreeView) Update(message runtimemsg.Msg) cmd.Cmd {
	switch msg := message.(type) {
	case *runtimemsg.MouseMsg:
		return t.updateMouse(msg)
	case *runtimemsg.KeyMsg:
		return t.updateKey(msg)
	}
	return nil
}

// updateMouse handles mouse messages for tree node interaction
func (t *TreeView) updateMouse(mouseMsg *runtimemsg.MouseMsg) cmd.Cmd {
	switch mouseMsg.Action {
	case runtimemsg.MouseActionPress:
		// Calculate which line was clicked based on LocalY
		// LocalY is relative to the TreeView component
		localY := mouseMsg.LocalY
		if localY < 0 {
			return nil
		}

		// Calculate target line (account for scroll offset)
		target := t.scrollOffset + localY
		if target < 0 || target >= len(t.lines) {
			return nil
		}

		// Set focus and select the line
		t.SetFocusIndex(target)
		t.SelectLine(target)
		return nil // TODO: Return Cmd to trigger re-render

	case runtimemsg.MouseActionMove:
		// Hover focus (non-selecting) for visual cue
		localY := mouseMsg.LocalY
		if localY < 0 {
			return nil
		}
		target := t.scrollOffset + localY
		if target < 0 || target >= len(t.lines) {
			return nil
		}
		if target != t.focusIndex {
			t.SetFocusIndex(target)
			return nil
		}
	}

	return nil
}

// updateKey handles keyboard messages for navigation (when focused)
func (t *TreeView) updateKey(keyMsg *runtimemsg.KeyMsg) cmd.Cmd {
	switch keyMsg.Special {
	case runtimeplatform.KeyUp:
		// Move focus up
		if t.focusIndex > 0 {
			t.SetFocusIndex(t.focusIndex - 1)
		}
		return nil

	case runtimeplatform.KeyDown:
		// Move focus down
		if t.focusIndex < len(t.lines)-1 {
			t.SetFocusIndex(t.focusIndex + 1)
		}
		return nil

	case runtimeplatform.KeyEnter:
		// Select current focused line
		if t.focusIndex >= 0 && t.focusIndex < len(t.lines) {
			t.SelectLine(t.focusIndex)
		}
		return nil

	case runtimeplatform.KeyHome:
		// Move to first line
		if len(t.lines) > 0 {
			t.SetFocusIndex(0)
		}
		return nil

	case runtimeplatform.KeyEnd:
		// Move to last line
		if len(t.lines) > 0 {
			t.SetFocusIndex(len(t.lines) - 1)
		}
		return nil

	case runtimeplatform.KeyPageUp:
		// Move up by viewport height
		t.MoveUpN(t.viewportHeight)
		return nil

	case runtimeplatform.KeyPageDown:
		// Move down by viewport height
		t.MoveDownN(t.viewportHeight)
		return nil
	}

	return nil
}

// MoveUpN moves focus up by n lines.
func (t *TreeView) MoveUpN(n int) {
	if n <= 0 {
		return
	}
	t.focusIndex -= n
	if t.focusIndex < 0 {
		t.focusIndex = 0
	}
	t.ensureVisible()
	t.regenerateDisplay()
}

// MoveDownN moves focus down by n lines.
func (t *TreeView) MoveDownN(n int) {
	if n <= 0 {
		return
	}
	t.focusIndex += n
	if t.focusIndex >= len(t.lines) {
		t.focusIndex = len(t.lines) - 1
	}
	t.ensureVisible()
	t.regenerateDisplay()
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

// SelectLine selects a line and regenerates display
func (t *TreeView) SelectLine(lineNum int) int {
	if lineNum < 0 || lineNum >= len(t.lines) {
		return -1
	}

	t.selectedIdx = lineNum
	t.regenerateDisplay() // Regenerate text nodes with new selection state
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
		t.regenerateDisplay()
	}
}

// MoveDown moves focus to the next line
func (t *TreeView) MoveDown() {
	if t.focusIndex < len(t.lines)-1 {
		t.focusIndex++
		t.ensureVisible()
		t.regenerateDisplay()
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
	t.regenerateDisplay()
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
	t.regenerateDisplay()
}

// Home moves focus to the first line
func (t *TreeView) Home() {
	t.focusIndex = 0
	t.scrollOffset = 0
	t.regenerateDisplay()
}

// End moves focus to the last line
func (t *TreeView) End() {
	t.focusIndex = len(t.lines) - 1
	t.ensureVisible()
	t.regenerateDisplay()
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

		// Mark that expand state changed (Inspector will rebuild tree)
		t.expandStateChanged = true
		t.expandStateLineIndex = t.focusIndex // Store which line was toggled
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			log.UILogger.Debug("[TreeView] Requesting expand toggle for line #%d (NodeID: %d)\n", t.focusIndex, line.NodeID)
		}
	}
}

// nodeHasChildren checks if a tree node can have children
func (t *TreeView) nodeHasChildren(line *TreeViewLine) bool {
	// Simple heuristic: if it's a container node (LayoutNode, ElementVNode with children)
	// In a real implementation, this would check the actual VNode structure
	return line.NodeType != "" && (line.NodeType == "LayoutNode" ||
		line.NodeType == "ElementVNode" ||
		line.NodeType == "ComponentNode")
}

// rebuildTreeFromSource rebuilds the tree lines based on current expand state
func (t *TreeView) rebuildTreeFromSource() {
	if t.builder == nil || len(t.builder.sourceLines) == 0 {
		return
	}

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		log.UILogger.Debug("[TreeView] Rebuilding tree with %d source lines\n", len(t.builder.sourceLines))
	}

	// Re-parse lines with current expand state
	t.builder.parseLines(t.builder.sourceLines)
	t.lines = t.builder.node.lines
	t.totalLines = len(t.lines)

	// Adjust focus index if it's now out of bounds
	if t.focusIndex >= t.totalLines {
		t.focusIndex = t.totalLines - 1
	}
	if t.focusIndex < 0 {
		t.focusIndex = 0
	}

	// Update display with new tree
	t.regenerateDisplay()
}

// SelectCurrent selects the currently focused line
func (t *TreeView) SelectCurrent() {
	if t.focusIndex >= 0 && t.focusIndex < len(t.lines) {
		t.selectedIdx = t.focusIndex
		t.regenerateDisplay()
	}
}

// GetFocusedLine returns the currently focused line
func (t *TreeView) GetFocusedLine() TreeViewLine {
	if t.focusIndex >= 0 && t.focusIndex < len(t.lines) {
		return t.lines[t.focusIndex]
	}
	return TreeViewLine{}
}

// GetLines returns all lines in the tree view
func (t *TreeView) GetLines() []TreeViewLine {
	return t.lines
}

// GetRender returns the current render result with latest focus/selection highlighting
// This should be used instead of Children() to get the latest styled VNode tree
func (t *TreeView) GetRender() ui.VNode {
	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		if t.currentRender == nil {
			log.UILogger.Debug("[TreeView] GetRender() returning nil!\n")
		} else {
			log.UILogger.Debug("[TreeView] GetRender() returning valid render, type=%T\n", t.currentRender)
		}
	}
	return t.currentRender
}

// SetViewportHeight sets the viewport height for scrolling calculations
func (t *TreeView) SetViewportHeight(height int) {
	if t.viewportHeight != height {
		t.viewportHeight = height
		// Trigger re-render when viewport height changes
		t.regenerateDisplay()
	}
}

// UpdateLines updates the tree lines without creating a new TreeView instance
// This preserves the viewportHeight that was set by the layout engine
func (t *TreeView) UpdateLines(lines []string) {
	if t.builder == nil {
		return
	}

	// Parse new lines (this now clears old lines before parsing)
	t.builder.sourceLines = lines
	t.builder.parseLines(lines)

	// Update lines from the builder's node
	t.lines = t.builder.node.lines
	t.totalLines = len(t.lines)

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		log.UILogger.Debug("[TreeView] UpdateLines: updated %d lines\n", len(t.lines))
	}

	// Re-render with the new lines (preserving viewportHeight)
	t.regenerateDisplay()
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
	if t.scrollOffset != offset {
		t.scrollOffset = offset
		// Trigger re-render when scroll offset changes
		t.regenerateDisplay()
	}
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

// SetFocusIndex sets the focus index and regenerates display
func (t *TreeView) SetFocusIndex(index int) {
	if index >= 0 && index < len(t.lines) {
		t.focusIndex = index
		t.ensureVisible()
		t.regenerateDisplay() // Regenerate text nodes with new focus state
	}
}

// ExpandStateChanged returns true if expand state changed since last check
// This is used by the Inspector to know when to rebuild the tree
func (t *TreeView) ExpandStateChanged() bool {
	return t.expandStateChanged
}

// GetExpandStateLineIndex returns the line index that needs expand toggle
func (t *TreeView) GetExpandStateLineIndex() int {
	return t.expandStateLineIndex
}

// ClearExpandStateChanged clears the expand state changed flag
func (t *TreeView) ClearExpandStateChanged() {
	t.expandStateChanged = false
	t.expandStateLineIndex = -1
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

// regenerateDisplay recreates the children VNodes based on current state
// This is called when focusIndex or selectedIdx changes to update the visual display
// Implements virtual scrolling: only renders visible lines based on viewportHeight
func (t *TreeView) regenerateDisplay() {
	if t.builder == nil {
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			log.UILogger.Debug("[TreeView] regenerateDisplay: builder is nil!\n")
		}
		return
	}

	if os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
		log.UILogger.Debug("[TreeView.regenerateDisplay] viewportHeight=%d, totalLines=%d\n", t.viewportHeight, len(t.lines))
	}

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		log.UILogger.Debug("[TreeView] regenerateDisplay: focus=%d, selected=%d, total lines=%d, viewportHeight=%d, scrollOffset=%d\n",
			t.focusIndex, t.selectedIdx, len(t.lines), t.viewportHeight, t.scrollOffset)
	}

	// Calculate visible range for virtual scrolling
	// Only render lines that are visible in the viewport
	totalLines := len(t.lines)
	var startLine, endLine int

	// If viewportHeight is not set (0 or negative), render all lines (fallback to old behavior)
	if t.viewportHeight <= 0 {
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			log.UILogger.Debug("[TreeView] Viewport height not set (%d), rendering all %d lines\n",
				t.viewportHeight, totalLines)
		}
		startLine = 0
		endLine = totalLines
	} else {
		// Virtual scrolling: only render visible lines
		startLine = t.scrollOffset
		endLine = startLine + t.viewportHeight

		// Clamp to valid range
		if startLine < 0 {
			startLine = 0
		}
		if endLine > totalLines {
			endLine = totalLines
		}
		if startLine >= totalLines {
			startLine = totalLines
			endLine = totalLines
		}

		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			log.UILogger.Debug("[TreeView] Virtual scroll: rendering lines [%d:%d] of %d total lines\n",
				startLine, endLine, totalLines)
		}
	}

	// Recreate text nodes for each VISIBLE line with current focus/selection state
	var lineNodes []ui.VNode
	for i := startLine; i < endLine; i++ {
		line := t.lines[i]

		// Use RawLine which has the original correctly formatted tree structure
		// If RawLine is empty (old code path), fall back to reconstruction
		var fullLine string
		if line.RawLine != "" {
			fullLine = line.RawLine

			// Insert icon after tree connectors (at Indent position)
			if t.builder.showIcons && line.NodeType != "" {
				icon := t.builder.getNodeIcon(line.NodeType)
				if icon != "" {
					// Insert icon at Indent position (which is where content starts)
					runes := []rune(fullLine)
					if line.Indent < len(runes) {
						// Insert icon before content
						prefix := string(runes[:line.Indent])
						suffix := string(runes[line.Indent:])
						fullLine = prefix + icon + suffix
					}
				}
			}
		} else {
			// Fallback for old code path
			fullLine = strings.Repeat(" ", line.Indent) + line.Prefix + line.Content
		}

		// Apply style based on focus/selection state using actual style attributes
		// NOTE: We don't add prefixes (>, *, spaces) to avoid breaking tree connector alignment
		if i == t.selectedIdx {
			// Selected line - use REVERSE video + BOLD (most visible)
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				log.UILogger.Debug("[TreeView] Line %d: SELECTED (REVERSE + BOLD)\n", i)
			}
			lineNodes = append(lineNodes, app.NewTextBuilder(fullLine).
				Style(style.NewStyle().
					Reverse(true).
					Bold(true)).
				Build())
		} else if i == t.focusIndex {
			// Focused line - use REVERSE video + BOLD
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				log.UILogger.Debug("[TreeView] Line %d: FOCUSED (REVERSE + BOLD)\n", i)
			}
			lineNodes = append(lineNodes, app.NewTextBuilder(fullLine).
				Style(style.NewStyle().
					Reverse(true).
					Bold(true)).
				Build())
		} else {
			// Normal line - white text
			lineNodes = append(lineNodes, app.NewTextBuilder(fullLine).
				Style(style.NewStyle().
					Foreground(style.White)).
				Build())
		}
	}

	// Add placeholder if no lines are visible (edge case)
	if len(lineNodes) == 0 && totalLines > 0 {
		lineNodes = append(lineNodes, app.NewTextBuilder("...").Build())
	}

	// Wrap in VStack for proper rendering
	// IMPORTANT: If viewportHeight is set, constrain the VStack height
	// This ensures the Layout engine measures VStack with bounded constraints
	vstackBuilder := ui.VStackBuilder(lineNodes...)
	if t.viewportHeight > 0 {
		vstackBuilder.Height(t.viewportHeight)
	}

	// IMPORTANT: Also constrain width to ensure it doesn't overflow parent
	// This prevents tree content from extending beyond the expected area
	contentWidth := t.lastMeasuredWidth
	if contentWidth <= 0 {
		// Fallback: get width from props or use default
		if w, ok := t.Props()["width"].(int); ok && w > 0 {
			contentWidth = w
		} else {
			// Use a reasonable default width
			contentWidth = 76 // Common inspector width minus padding
		}
	}

	if contentWidth > 0 {
		vstackBuilder.Width(contentWidth)

		// Truncate long lines to fit within width to prevent overflow
		// This is important for ensuring tree content doesn't extend beyond borders
		truncatedLineNodes := make([]ui.VNode, 0, len(lineNodes))
		for _, node := range lineNodes {
			// Try to get the text content from the node
			var text string
			switch n := node.(type) {
			case *rtui.TextVNode:
				text = n.Content()
			case interface{ String() string }:
				text = n.String()
			}

			// IMPORTANT: Truncate by number of runes (each rune = 1 cell in TUI)
			// This correctly handles multi-rune emojis like 🖼️ (2 runes = 2 cells)
			if text != "" && len(text) > contentWidth {
				runes := []rune(text)
				if contentWidth > 3 {
					truncated := string(runes[:contentWidth-3]) + "..."
					truncatedLineNodes = append(truncatedLineNodes, app.NewTextBuilder(truncated).Build())
				} else {
					truncatedLineNodes = append(truncatedLineNodes, node)
				}
			} else {
				truncatedLineNodes = append(truncatedLineNodes, node)
			}

			if text != "" && len(text) > contentWidth {
				// Truncate to fit within width
				runes := []rune(text)
				if contentWidth > 3 {
					truncated := string(runes[:contentWidth-3]) + "..."
					truncatedLineNodes = append(truncatedLineNodes, app.NewTextBuilder(truncated).Build())
				} else {
					truncatedLineNodes = append(truncatedLineNodes, node)
				}
			} else {
				truncatedLineNodes = append(truncatedLineNodes, node)
			}
		}
		lineNodes = truncatedLineNodes
	}

	result := vstackBuilder.Build()

	// IMPORTANT: Store this as the current render result for GetRender()
	t.currentRender = result

	t.SetChildren([]ui.VNode{result})

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		log.UILogger.Debug("[TreeView] regenerateDisplay: Rendered %d lines (visible range [%d:%d]), Set %d children\n",
			len(lineNodes), startLine, endLine, len(result.Children()))
	}
}

// Measure implements runtime.Measurable interface
// This allows TreeView to respond to layout constraints and adjust its viewport
func (t *TreeView) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if t == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// FORCE PANIC TO VERIFY CALL
	if os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
		log.UILogger.Debug("PANIC CHECK: TreeView.Measure called!\n")
	}

	if os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
		log.UILogger.Debug("[TreeView.Measure] constraints=%v\n", constraints)
	}

	totalLines := len(t.lines)

	// Default size
	width := 80
	height := totalLines

	// Respect width constraint if bounded
	if constraints.HasBoundedWidth() {
		width = constraints.MaxWidth
		if width == 0 || width > 200 {
			width = 80
		}
		// Store measured width for constraining internal VStack
		t.lastMeasuredWidth = width
	}

	// If we have bounded height, use it for viewport
	if constraints.HasBoundedHeight() {
		viewportHeight := constraints.MaxHeight

		// Check if viewportHeight changed, and if so, re-render
		oldViewportHeight := t.viewportHeight
		t.SetViewportHeight(viewportHeight)
		height = viewportHeight

		// Re-render if viewport height changed (to apply virtual scrolling)
		if oldViewportHeight != viewportHeight && t.builder != nil {
			if os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
				log.UILogger.Debug("[TreeView.Measure] Bounded height: %d, triggering regenerateDisplay\n", viewportHeight)
			}
			t.regenerateDisplay()
		} else {
			if os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
				log.UILogger.Debug("[TreeView.Measure] Viewport height unchanged: %d\n", viewportHeight)
			}
		}
	} else {
		// No bounded height - render all lines
		t.SetViewportHeight(0) // 0 means render all
		height = totalLines

		if os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
			log.UILogger.Debug("[TreeView.Measure] No bounded height, rendering all %d lines\n", totalLines)
		}
	}

	return runtime.Size{Width: width, Height: height}
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction implements ActionTarget interface
func (t *TreeView) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	// Handle action based on type
	switch act.Type {
	// Navigation actions
	case action.ActionNavigateUp:
		t.MoveUp()
		return true
	case action.ActionNavigateDown:
		t.MoveDown()
		return true
	case action.ActionNavigateLeft:
		// Left arrow: collapse current node
		t.ToggleExpandCurrent()
		return true
	case action.ActionNavigateRight:
		// Right arrow: expand current node
		t.ToggleExpandCurrent()
		return true
	case action.ActionNavigatePageUp:
		t.PageUp()
		return true
	case action.ActionNavigatePageDown:
		t.PageDown()
		return true
	case action.ActionNavigateHome:
		t.Home()
		return true
	case action.ActionNavigateEnd:
		t.End()
		return true

	// Selection actions
	case action.ActionSelect, action.ActionEnter:
		t.SelectCurrent()
		return true

	// Expand/Collapse actions
	case action.ActionToggle, action.ActionExpand, action.ActionCollapse:
		t.ToggleExpandCurrent()
		return true

	// Scroll action
	case action.ActionScroll:
		if delta, ok := act.GetPayloadInt(); ok {
			if delta > 0 && t.CanScrollDown() {
				t.MoveDown()
				return true
			} else if delta < 0 && t.CanScrollUp() {
				t.MoveUp()
				return true
			}
		}
		return false

	// Mouse click
	case action.ActionClick:
		// Click action already handled by HandleEvent
		// But we can update selection if needed
		if t.focusIndex >= 0 && t.focusIndex < len(t.lines) {
			t.SelectLine(t.focusIndex)
			return true
		}
		return false
	}

	return false
}

// GetSupportedActions implements ActionTarget interface
func (t *TreeView) GetSupportedActions() []action.ActionType {
	if t.supportedActions == nil {
		return []action.ActionType{
			action.ActionNavigateUp,
			action.ActionNavigateDown,
			action.ActionNavigateLeft,
			action.ActionNavigateRight,
			action.ActionNavigatePageUp,
			action.ActionNavigatePageDown,
			action.ActionNavigateHome,
			action.ActionNavigateEnd,
			action.ActionSelect,
			action.ActionToggle,
			action.ActionExpand,
			action.ActionCollapse,
			action.ActionScroll,
			action.ActionClick,
		}
	}
	return t.supportedActions
}

// CanHandleAction implements ActionTarget interface
func (t *TreeView) CanHandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	// Check if action type is supported
	supported := t.GetSupportedActions()
	for _, supportedType := range supported {
		if supportedType == act.Type {
			return true
		}
	}

	return false
}

// ============================================================================
// FocusableActionTarget 接口实现
// ============================================================================

// Focus implements FocusableActionTarget interface
func (t *TreeView) Focus() bool {
	if len(t.lines) == 0 {
		return false
	}
	// Focus on first line if no focus
	if t.focusIndex < 0 {
		t.focusIndex = 0
	}
	t.regenerateDisplay()
	return true
}

// Blur implements FocusableActionTarget interface
func (t *TreeView) Blur() {
	// Clear visual focus indication
	t.focusIndex = -1
	t.regenerateDisplay()
}

// IsFocused implements FocusableActionTarget interface
func (t *TreeView) IsFocused() bool {
	return t.focusIndex >= 0 && t.focusIndex < len(t.lines)
}

// IsFocusable implements FocusableActionTarget interface
func (t *TreeView) IsFocusable() bool {
	return len(t.lines) > 0
}

// ============================================================================
// ScrollableActionTarget 接口实现
// ============================================================================

// CanScroll implements ScrollableActionTarget interface
func (t *TreeView) CanScroll(delta int) bool {
	if delta > 0 {
		return t.CanScrollDown()
	} else if delta < 0 {
		return t.CanScrollUp()
	}
	return false
}

// Scroll implements ScrollableActionTarget interface
func (t *TreeView) Scroll(delta int) bool {
	if !t.CanScroll(delta) {
		return false
	}
	t.ScrollBy(delta)
	return true
}

// GetScrollPosition implements ScrollableActionTarget interface
func (t *TreeView) GetScrollPosition() (int, int, int) {
	current := t.focusIndex
	total := len(t.lines)
	visible := t.viewportHeight
	if visible <= 0 {
		visible = total
	}
	return current, total, visible
}

// ============================================================================
// SelectableActionTarget 接口实现
// ============================================================================

// Select implements SelectableActionTarget interface
func (t *TreeView) Select() bool {
	if !t.IsFocusable() {
		return false
	}
	t.SelectCurrent()
	return true
}

// IsSelected implements SelectableActionTarget interface
func (t *TreeView) IsSelected() bool {
	return t.HasSelection()
}

// ToggleSelection implements SelectableActionTarget interface
func (t *TreeView) ToggleSelection() bool {
	if t.HasSelection() {
		t.ClearSelection()
		return false
	}
	t.SelectCurrent()
	return true
}

// GetSelectedCount implements SelectableActionTarget interface
func (t *TreeView) GetSelectedCount() int {
	if t.HasSelection() {
		return 1
	}
	return 0
}

// Note: TreeView cannot fully implement ExpandableActionTarget interface
// because it has IsExpanded(int) method which conflicts with IsExpanded().
// The expand/collapse functionality is available through:
// - ToggleExpandCurrent()
// - SetExpanded(nodeID, expanded)
// - IsExpanded(nodeID)

