package treeview

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for TreeView components
type Instance struct {
	// === Identification ===
	key         string
	componentID string // Component ID for Intent routing (Phase 10)

	// === Props (from VNode, may change each render) ===
	nodes          []TreeNode
	expandLevel    int
	showIcons      bool
	showLineNums   bool
	compact        bool
	treeStyle      style.Style
	selectedStyle  style.Style
	iconStyle      style.Style
	scrollOffset   int
	selectedIndex  int
	viewportHeight int
	allowScroll    bool
	allowExpand    bool

	// === Runtime State ===
	expandState map[int]bool // Expand/collapse state for each node
	bounds      [4]int       // x, y, w, h
	dirty       bool

	// === Intent Support (Phase 10) ===
	intentEmitter func(intent.Intent) // Intent emitter for bubbling
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new TreeViewInstance from props
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:            getStringProp(props, "key", ""),
		componentID:    getStringProp(props, "componentID", ""),
		nodes:          getNodesProp(props, []TreeNode{}),
		expandLevel:    getIntProp(props, "expandLevel", 1),
		showIcons:      getBoolProp(props, "showIcons", true),
		showLineNums:   getBoolProp(props, "showLineNums", false),
		compact:        getBoolProp(props, "compact", false),
		treeStyle:      getStyleProp(props, "treeStyle"),
		selectedStyle:  getStyleProp(props, "selectedStyle"),
		iconStyle:      getStyleProp(props, "iconStyle"),
		scrollOffset:   getIntProp(props, "scrollOffset", 0),
		selectedIndex:  getIntProp(props, "selectedIndex", -1),
		viewportHeight: getIntProp(props, "viewportHeight", 10),
		allowScroll:    getBoolProp(props, "allowScroll", true),
		allowExpand:    getBoolProp(props, "allowExpand", true),
		expandState:    make(map[int]bool),
		dirty:          true,
	}

	// Initialize expand state based on expandLevel
	inst.initExpandState()

	return inst
}

// initExpandState initializes expand states based on expandLevel
func (inst *Instance) initExpandState() {
	for i, node := range inst.nodes {
		// Calculate node depth based on indent
		depth := node.Indent / 4 // Assume 4 spaces per level

		if inst.expandLevel < 0 {
			// All expanded
			inst.expandState[i] = true
		} else if depth < inst.expandLevel {
			inst.expandState[i] = true
		}
	}
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }

// Parent implements TreeComponent interface (intent bubble).
// Returns nil as TreeView is currently a leaf component without parent tracking.
// Can be extended in the future to support nested tree structures.
func (inst *Instance) Parent() interface{} { return nil }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.expandState = nil }
func (inst *Instance) OnMount()              { inst.dirty = true }
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset

	inst.componentID = getStringProp(props, "componentID", inst.componentID)
	inst.nodes = getNodesProp(props, inst.nodes)
	inst.expandLevel = getIntProp(props, "expandLevel", inst.expandLevel)
	inst.showIcons = getBoolProp(props, "showIcons", inst.showIcons)
	inst.showLineNums = getBoolProp(props, "showLineNums", inst.showLineNums)
	inst.compact = getBoolProp(props, "compact", inst.compact)
	inst.treeStyle = getStyleProp(props, "treeStyle")
	inst.selectedStyle = getStyleProp(props, "selectedStyle")
	inst.iconStyle = getStyleProp(props, "iconStyle")
	inst.scrollOffset = getIntProp(props, "scrollOffset", inst.scrollOffset)
	inst.selectedIndex = getIntProp(props, "selectedIndex", inst.selectedIndex)
	inst.viewportHeight = getIntProp(props, "viewportHeight", inst.viewportHeight)
	inst.allowScroll = getBoolProp(props, "allowScroll", inst.allowScroll)
	inst.allowExpand = getBoolProp(props, "allowExpand", inst.allowExpand)

	inst.initExpandState()

	changed := oldSelected != inst.selectedIndex || oldScroll != inst.scrollOffset
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":            inst.key,
		"nodes":          inst.nodes,
		"expandLevel":    inst.expandLevel,
		"showIcons":      inst.showIcons,
		"showLineNums":   inst.showLineNums,
		"compact":        inst.compact,
		"scrollOffset":   inst.scrollOffset,
		"selectedIndex":  inst.selectedIndex,
		"viewportHeight": inst.viewportHeight,
		"allowScroll":    inst.allowScroll,
		"allowExpand":    inst.allowExpand,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := 60 // Default width
	height := inst.calculateHeight()

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return layout.Size{Width: width, Height: height}
}

// calculateHeight calculates the total height of the tree
func (inst *Instance) calculateHeight() int {
	visibleNodes := inst.getVisibleNodes()
	return len(visibleNodes) + 2 // +2 for border
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the tree view
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd
	treeStyle := inst.treeStyle

	// Get visible nodes based on expand state
	visibleNodes := inst.getVisibleNodes()

	// Calculate width
	width := inst.calculateWidth()

	// Draw top border
	topBorder := "┌" + strings.Repeat("─", width-2) + "┐"
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, treeStyle))

	// Draw visible tree nodes
	startLine := inst.scrollOffset
	for i := startLine; i < len(visibleNodes) && i-startLine < inst.viewportHeight; i++ {
		node := visibleNodes[i]
		rowY := y + 1 + (i - startLine)

		// Build line
		line := inst.buildTreeLine(node, i == inst.selectedIndex)

		// Add padding
		if len(line) < width-2 {
			line += strings.Repeat(" ", width-2-len(line))
		}
		line = "│ " + line + " │"

		// Determine style
		style := treeStyle
		if i == inst.selectedIndex {
			style = inst.selectedStyle
		}

		cmds = append(cmds, paint.NewTextCmd(x, rowY, line, style))
	}

	// Fill remaining space
	visibleCount := len(visibleNodes)
	rowsShown := min(inst.viewportHeight, visibleCount-startLine)
	for i := rowsShown; i < inst.viewportHeight; i++ {
		rowY := y + 1 + i
		emptyRow := "│" + strings.Repeat(" ", width-2) + "│"
		cmds = append(cmds, paint.NewTextCmd(x, rowY, emptyRow, treeStyle))
	}

	// Draw bottom border
	bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"
	cmds = append(cmds, paint.NewTextCmd(x, y+inst.viewportHeight+1, bottomBorder, treeStyle))

	return cmds
}

// buildTreeLine builds a single tree line
func (inst *Instance) buildTreeLine(node TreeNode, isSelected bool) string {
	var line strings.Builder

	// Add indentation
	for i := 0; i < node.Indent; i += 4 {
		line.WriteString("│   ")
	}

	// Add icon if enabled
	if inst.showIcons {
		icon := "  "
		if node.NodeType == "folder" {
			if inst.isExpanded(node.NodeID) {
				icon = "📂 "
			} else {
				icon = "📁 "
			}
		} else {
			icon = "📄 "
		}
		line.WriteString(icon)
	} else {
		line.WriteString("  ")
	}

	// Add content
	line.WriteString(node.Content)

	// Add line number if enabled
	if inst.showLineNums {
		line.WriteString(fmt.Sprintf(" [%d]", node.NodeID))
	}

	return line.String()
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) HandleAction(act *action.Action) bool {
	if !inst.allowScroll {
		return false
	}

	switch act.Type {
	case action.ActionNavigateUp:
		if inst.selectedIndex > 0 {
			return inst.navigateUp()
		}
		return false
	case action.ActionNavigateDown:
		visibleNodes := inst.getVisibleNodes()
		if inst.selectedIndex < len(visibleNodes)-1 {
			return inst.navigateDown()
		}
		return false
	case action.ActionNavigateHome:
		if inst.scrollOffset > 0 {
			return inst.navigateHome()
		}
		return false
	case action.ActionNavigateEnd:
		visibleNodes := inst.getVisibleNodes()
		if inst.scrollOffset < len(visibleNodes)-inst.viewportHeight {
			return inst.navigateEnd()
		}
		return false
	case action.ActionNavigatePageUp:
		if inst.scrollOffset > 0 {
			_ = inst.pageUp()
			return true
		}
		return false
	case action.ActionNavigatePageDown:
		visibleNodes := inst.getVisibleNodes()
		if inst.scrollOffset < len(visibleNodes)-inst.viewportHeight {
			_ = inst.pageDown()
			return true
		}
		return false
	case action.ActionSelect:
		return inst.toggleExpand()
	}
	return false
}

// =============================================================================
// Navigation Methods
// =============================================================================

func (inst *Instance) navigateUp() bool {
	if inst.selectedIndex > 0 {
		fromIndex := inst.selectedIndex
		inst.selectedIndex--
		inst.dirty = true
		// Emit Navigation Intent (Phase 10)
		inst.emitNavigation("up", fromIndex, inst.selectedIndex)
		return true
	}
	return false
}

func (inst *Instance) navigateDown() bool {
	visibleNodes := inst.getVisibleNodes()
	if inst.selectedIndex < len(visibleNodes)-1 {
		fromIndex := inst.selectedIndex
		inst.selectedIndex++
		inst.dirty = true
		// Emit Navigation Intent (Phase 10)
		inst.emitNavigation("down", fromIndex, inst.selectedIndex)
		return true
	}
	return false
}

func (inst *Instance) navigateHome() bool {
	inst.scrollOffset = 0
	inst.selectedIndex = 0
	inst.dirty = true
	return true
}

func (inst *Instance) navigateEnd() bool {
	visibleNodes := inst.getVisibleNodes()
	inst.scrollOffset = max(0, len(visibleNodes)-inst.viewportHeight)
	inst.selectedIndex = len(visibleNodes) - 1
	inst.dirty = true
	return true
}

func (inst *Instance) pageUp() int {
	inst.scrollOffset = max(0, inst.scrollOffset-inst.viewportHeight)
	inst.selectedIndex = max(0, inst.selectedIndex-inst.viewportHeight)
	inst.dirty = true
	return inst.scrollOffset
}

func (inst *Instance) pageDown() int {
	visibleNodes := inst.getVisibleNodes()
	maxOffset := max(0, len(visibleNodes)-inst.viewportHeight)
	inst.scrollOffset = min(maxOffset, inst.scrollOffset+inst.viewportHeight)
	inst.selectedIndex = min(len(visibleNodes)-1, inst.selectedIndex+inst.viewportHeight)
	inst.dirty = true
	return inst.scrollOffset
}

func (inst *Instance) toggleExpand() bool {
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(inst.nodes) {
		return false
	}

	node := inst.nodes[inst.selectedIndex]
	wasExpanded := inst.expandState[inst.selectedIndex]
	inst.expandState[inst.selectedIndex] = !wasExpanded
	nowExpanded := inst.expandState[inst.selectedIndex]
	inst.dirty = true

	// Emit Expand/Collapse Intent (Phase 10)
	if nowExpanded {
		inst.emitNodeExpand(inst.selectedIndex, node.Path, node.NodeID)
	} else {
		inst.emitNodeCollapse(inst.selectedIndex, node.Path, node.NodeID)
	}

	return nowExpanded
}

// =============================================================================
// Helper Methods
// =============================================================================

// getVisibleNodes returns nodes that should be visible based on expand state
func (inst *Instance) getVisibleNodes() []TreeNode {
	visible := make([]TreeNode, 0, len(inst.nodes))

	// Track parent depth to determine visibility
	stack := []int{} // Stack of node indices

	for i, node := range inst.nodes {
		depth := node.Indent / 4

		// Pop nodes from stack until we find parent
		for len(stack) > 0 && stack[len(stack)-1] >= depth {
			stack = stack[:len(stack)-1]
		}

		// Check if parent is expanded
		isVisible := true
		if len(stack) > 0 {
			parentIndex := stack[len(stack)-1]
			if !inst.expandState[parentIndex] {
				isVisible = false
			}
		}

		if isVisible {
			visible = append(visible, node)
		}

		// Push current node to stack if it's a folder and visible
		if node.NodeType == "folder" && isVisible {
			stack = append(stack, i)
		}
	}

	return visible
}

// isExpanded checks if a node is expanded
func (inst *Instance) isExpanded(nodeID int) bool {
	return inst.expandState[nodeID]
}

// calculateWidth calculates the maximum width needed
func (inst *Instance) calculateWidth() int {
	maxWidth := 40 // Minimum width

	for _, node := range inst.nodes {
		line := inst.buildTreeLine(node, false)
		if len(line)+4 > maxWidth {
			maxWidth = len(line) + 4
		}
	}

	return maxWidth
}

// =============================================================================
// Getters
// =============================================================================

func (inst *Instance) GetScrollOffset() int   { return inst.scrollOffset }
func (inst *Instance) GetSelectedIndex() int  { return inst.selectedIndex }
func (inst *Instance) GetViewportHeight() int { return inst.viewportHeight }
func (inst *Instance) GetComponentID() string { return inst.componentID }
func (inst *Instance) GetNodes() []TreeNode {
	return append([]TreeNode(nil), inst.nodes...)
}
func (inst *Instance) GetVisibleNodes() []TreeNode {
	return append([]TreeNode(nil), inst.getVisibleNodes()...)
}
func (inst *Instance) GetSelectedNode() (TreeNode, bool) {
	visibleNodes := inst.getVisibleNodes()
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(visibleNodes) {
		return TreeNode{}, false
	}
	return visibleNodes[inst.selectedIndex], true
}
func (inst *Instance) SelectIndex(index int) bool {
	visibleNodes := inst.getVisibleNodes()
	if index < 0 || index >= len(visibleNodes) {
		return false
	}
	if inst.selectedIndex == index {
		return false
	}
	inst.selectedIndex = index
	inst.dirty = true
	return true
}

// =============================================================================
// Intent Bubble Support (Phase 10)
// =============================================================================

// EmitIntent emits an intent through the bubble system.
func (inst *Instance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil {
		inst.intentEmitter(i)
	}
}

// SetIntentEmitter sets the intent emitter function for bubbling.
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// HandleIntent implements intent.IntentHandler to handle treeview-specific intents.
// This allows external components or controllers to control the treeview via intents.
func (inst *Instance) HandleIntent(i intent.Intent) bool {
	// Only handle intents for this treeview (if componentID is set)
	if inst.componentID != "" {
		if id, ok := i.(interface{ GetComponentID() string }); ok {
			if id.GetComponentID() != "" && id.GetComponentID() != inst.componentID {
				// Intent is for a different treeview, ignore
				return false
			}
		}
	}

	switch v := i.(type) {
	case NodeSelectIntent:
		// Handle selection by external request
		if v.NodeIndex >= 0 && v.NodeIndex < len(inst.nodes) {
			inst.selectedIndex = v.NodeIndex
			inst.dirty = true
			return true
		}

	case NodeExpandIntent:
		// Handle expand by external request
		if v.NodeIndex >= 0 && v.NodeIndex < len(inst.nodes) {
			inst.expandState[v.NodeIndex] = true
			inst.dirty = true
			return true
		}

	case NodeCollapseIntent:
		// Handle collapse by external request
		if v.NodeIndex >= 0 && v.NodeIndex < len(inst.nodes) {
			inst.expandState[v.NodeIndex] = false
			inst.dirty = true
			return true
		}
	}

	return false
}

// emitNodeSelect emits a NodeSelectIntent when a node is selected.
func (inst *Instance) emitNodeSelect(nodeIndex int) {
	if nodeIndex < 0 || nodeIndex >= len(inst.nodes) {
		return
	}

	node := inst.nodes[nodeIndex]
	nodeSelect := NodeSelect(nodeIndex, node.Path, node.NodeID, node.NodeType, node.Content)
	if inst.componentID != "" {
		nodeSelect = NodeSelectWithID(inst.componentID, nodeIndex, node.Path, node.NodeID, node.NodeType, node.Content)
	}

	intent.Emit(inst, nodeSelect)
}

// emitNodeExpand emits a NodeExpandIntent when a folder is expanded.
func (inst *Instance) emitNodeExpand(nodeIndex int, path string, nodeID int) {
	var expandIntent NodeExpandIntent
	if inst.componentID != "" {
		expandIntent = NodeExpandWithID(inst.componentID, nodeIndex, path, nodeID)
	} else {
		expandIntent = NodeExpand(nodeIndex, path, nodeID)
	}

	intent.Emit(inst, expandIntent)
}

// emitNodeCollapse emits a NodeCollapseIntent when a folder is collapsed.
func (inst *Instance) emitNodeCollapse(nodeIndex int, path string, nodeID int) {
	var collapseIntent NodeCollapseIntent
	if inst.componentID != "" {
		collapseIntent = NodeCollapseWithID(inst.componentID, nodeIndex, path, nodeID)
	} else {
		collapseIntent = NodeCollapse(nodeIndex, path, nodeID)
	}

	intent.Emit(inst, collapseIntent)
}

// emitNavigation emits a NavigationIntent when the selection changes via navigation.
func (inst *Instance) emitNavigation(direction string, fromIndex, toIndex int) {
	var navIntent NavigationIntent
	if inst.componentID != "" {
		navIntent = NavigationWithID(inst.componentID, direction, fromIndex, toIndex)
	} else {
		navIntent = Navigation(direction, fromIndex, toIndex)
	}

	intent.Emit(inst, navIntent)
}

// =============================================================================
// Bounds Support
// =============================================================================

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string) style.Style {
	v, ok := props[key]
	if !ok {
		return style.Style{}
	}
	if s, ok := v.(style.Style); ok {
		return s
	}
	return style.Style{}
}

func getNodesProp(props rtui.Props, def []TreeNode) []TreeNode {
	v, ok := props["nodes"]
	if !ok {
		return def
	}
	if nodes, ok := v.([]TreeNode); ok {
		return nodes
	}
	return def
}
