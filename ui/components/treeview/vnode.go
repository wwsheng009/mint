package treeview

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types
// =============================================================================

// TreeNode represents a single node in the tree view
type TreeNode struct {
	Indent   int    // Byte offset where content starts
	Content  string // The actual node content
	Path     string // Unique path identifier (stable across expand/collapse)
	NodeType string // Node type for icon display (folder, file, etc.)
	NodeID   int    // Optional node ID
}

// =============================================================================
// VNode - Pure Description
// =============================================================================

// VNode is the pure description for the TreeView component
// Contains only declarative properties - no state, no closures, no Paint
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key         string
	componentID string // Component ID for Intent routing (Phase 10)

	// === Visual Properties ===
	nodes         []TreeNode // Tree nodes to display
	expandLevel   int        // Default expand level (0 = collapsed, -1 = all expanded)
	showIcons     bool       // Show node icons
	showLineNums  bool       // Show line numbers
	compact       bool       // Use compact display

	// === Styles ===
	treeStyle    style.Style // Style for tree items
	selectedStyle style.Style // Style for selected item
	iconStyle    style.Style // Style for icons

	// === State Properties (declarative initial state) ===
	scrollOffset   int // Initial scroll offset
	selectedIndex  int // Currently selected node index
	viewportHeight int // Visible height for scrolling

	// === Interaction ===
	allowScroll    bool // Whether scrolling is enabled
	allowExpand    bool // Whether expand/collapse is enabled
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new TreeView VNode
func New() *VNode {
	return &VNode{
		ElementVNode:   rtui.NewElement("treeview"),
		key:            "",
		componentID:    "",
		nodes:          []TreeNode{},
		expandLevel:    1, // Default: expand first level
		showIcons:      true,
		showLineNums:   false,
		compact:        false,
		treeStyle:      style.Style{},
		selectedStyle:  style.Style{BG: style.Blue, FG: style.White},
		iconStyle:      style.Style{FG: style.Yellow},
		scrollOffset:   0,
		selectedIndex:  -1,
		viewportHeight: 10,
		allowScroll:    true,
		allowExpand:    true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                   { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode  { v.key = key; return v }
func (v *VNode) Tag() string                   { return "treeview" }
func (v *VNode) Style() style.Style            { return v.treeStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.treeStyle = s; return v }
func (v *VNode) Children() []rtui.VNode        { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer          { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":            v.key,
		"componentID":    v.componentID,
		"nodes":          v.nodes,
		"expandLevel":    v.expandLevel,
		"showIcons":      v.showIcons,
		"showLineNums":   v.showLineNums,
		"compact":        v.compact,
		"treeStyle":      v.treeStyle,
		"selectedStyle":  v.selectedStyle,
		"iconStyle":      v.iconStyle,
		"scrollOffset":   v.scrollOffset,
		"selectedIndex":  v.selectedIndex,
		"viewportHeight": v.viewportHeight,
		"allowScroll":    v.allowScroll,
		"allowExpand":    v.allowExpand,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if key, ok := p["key"].(string); ok {
		v.key = key
	}
	if componentID, ok := p["componentID"].(string); ok {
		v.componentID = componentID
	}
	if nodes, ok := p["nodes"].([]TreeNode); ok {
		v.nodes = nodes
	}
	if expandLevel, ok := p["expandLevel"].(int); ok {
		v.expandLevel = expandLevel
	}
	if showIcons, ok := p["showIcons"].(bool); ok {
		v.showIcons = showIcons
	}
	if showLineNums, ok := p["showLineNums"].(bool); ok {
		v.showLineNums = showLineNums
	}
	if compact, ok := p["compact"].(bool); ok {
		v.compact = compact
	}
	if treeStyle, ok := p["treeStyle"].(style.Style); ok {
		v.treeStyle = treeStyle
	}
	if selectedStyle, ok := p["selectedStyle"].(style.Style); ok {
		v.selectedStyle = selectedStyle
	}
	if iconStyle, ok := p["iconStyle"].(style.Style); ok {
		v.iconStyle = iconStyle
	}
	if scrollOffset, ok := p["scrollOffset"].(int); ok {
		v.scrollOffset = scrollOffset
	}
	if selectedIndex, ok := p["selectedIndex"].(int); ok {
		v.selectedIndex = selectedIndex
	}
	if viewportHeight, ok := p["viewportHeight"].(int); ok {
		v.viewportHeight = viewportHeight
	}
	if allowScroll, ok := p["allowScroll"].(bool); ok {
		v.allowScroll = allowScroll
	}
	if allowExpand, ok := p["allowExpand"].(bool); ok {
		v.allowExpand = allowExpand
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// =============================================================================
// Setter Methods (Fluent API)
// =============================================================================

func (v *VNode) SetNodes(nodes []TreeNode) *VNode { v.nodes = nodes; return v }
func (v *VNode) SetComponentID(id string) *VNode  { v.componentID = id; return v }
func (v *VNode) SetExpandLevel(level int) *VNode  { v.expandLevel = level; return v }
func (v *VNode) SetShowIcons(show bool) *VNode    { v.showIcons = show; return v }
func (v *VNode) SetShowLineNums(show bool) *VNode { v.showLineNums = show; return v }
func (v *VNode) SetCompact(compact bool) *VNode   { v.compact = compact; return v }
func (v *VNode) SetTreeStyle(s style.Style) *VNode { v.treeStyle = s; return v }
func (v *VNode) SetSelectedStyle(s style.Style) *VNode { v.selectedStyle = s; return v }
func (v *VNode) SetIconStyle(s style.Style) *VNode { v.iconStyle = s; return v }
func (v *VNode) SetScrollOffset(offset int) *VNode { v.scrollOffset = offset; return v }
func (v *VNode) SetSelectedIndex(index int) *VNode { v.selectedIndex = index; return v }
func (v *VNode) SetViewportHeight(height int) *VNode { v.viewportHeight = height; return v }
func (v *VNode) SetAllowScroll(allow bool) *VNode  { v.allowScroll = allow; return v }
func (v *VNode) SetAllowExpand(allow bool) *VNode  { v.allowExpand = allow; return v }

// =============================================================================
// Getter Methods
// =============================================================================

// GetSelectedIndex returns the currently selected node index
func (v *VNode) GetSelectedIndex() int { return v.selectedIndex }

// GetComponentID returns the component ID
func (v *VNode) GetComponentID() string { return v.componentID }

// =============================================================================
// Convenience Methods
// =============================================================================

// AddNode adds a single node to the tree
func (v *VNode) AddNode(node TreeNode) *VNode {
	v.nodes = append(v.nodes, node)
	return v
}

// FromLines creates tree nodes from pre-formatted lines
func (v *VNode) FromLines(lines []string) *VNode {
	v.nodes = parseLines(lines)
	return v
}

// =============================================================================
// Helper Functions
// =============================================================================

// parseLines converts pre-formatted tree lines to TreeNode structures
func parseLines(lines []string) []TreeNode {
	nodes := make([]TreeNode, 0, len(lines))

	for i, line := range lines {
		// Calculate indent by counting leading spaces
		indent := 0
		for _, ch := range line {
			if ch == ' ' {
				indent++
			} else if ch == '\t' {
				indent += 4 // Assume 4 spaces per tab
			} else {
				break
			}
		}

		// Trim leading spaces to get content
		content := strings.TrimSpace(line)

		// Detect node type from content
		nodeType := "file"
		if strings.Contains(content, "/") || !strings.Contains(content, ".") {
			nodeType = "folder"
		}

		nodes = append(nodes, TreeNode{
			Indent:   indent,
			Content:  content,
			Path:     content, // Simplified: use content as path
			NodeType: nodeType,
			NodeID:   i,
		})
	}

	return nodes
}
