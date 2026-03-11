package treeview

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types
// =============================================================================

// TreeNode represents a single node in the tree view
type TreeNode struct {
	Indent    int    // Byte offset where content starts
	Content   string // The actual node content
	Path      string // Unique path identifier (stable across expand/collapse)
	NodeType  string // Node type for icon display (folder, file, etc.)
	NodeID    int    // Optional node ID
	Lazy      bool   // Whether children are loaded lazily
	Loading   bool   // Whether node is currently loading children
	LoadError string // Error message when loading children fails
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
	showBorder    bool       // Show border around tree
	showScrollbar bool       // Show scrollbar indicator

	// === Styles ===
	treeStyle          style.Style // Style for tree items
	selectedStyle      style.Style // Style for selected item
	iconStyle          style.Style // Style for icons
	scrollbarStyle     style.Style // Style for scrollbar
	rowStyleFn         func(int, TreeNode) style.Style
	matchStyle         style.Style // Style for matched search rows
	lazyLoadFn         func(TreeNode)
	lazyLoadChildrenFn func(TreeNode) []TreeNode

	// === State Properties (declarative initial state) ===
	scrollOffset            int  // Initial scroll offset
	scrollOffsetControlled  bool // Whether scrollOffset is externally controlled
	selectedIndex           int  // Currently selected node index
	selectedIndexControlled bool // Whether selectedIndex is externally controlled
	viewportHeight          int  // Visible height for scrolling
	expandedKeys            map[string]bool
	expandedKeysControlled  bool
	searchQuery             string
	searchQueryControlled   bool
	searchFn                func(TreeNode, string) bool
	selectionMode           SelectionMode
	checkedKeys             map[string]bool
	checkedKeysControlled   bool
	selectionIntent         intent.Intent
	selectionIntentField    intent.FieldIntent
	showSearchStats         bool
	searchStatsStyle        style.Style

	// === Interaction ===
	allowScroll bool // Whether scrolling is enabled
	allowExpand bool // Whether expand/collapse is enabled
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
		ElementVNode:       rtui.NewElement("treeview"),
		key:                "",
		componentID:        "",
		nodes:              []TreeNode{},
		expandLevel:        1, // Default: expand first level
		showIcons:          true,
		showLineNums:       false,
		compact:            false,
		showBorder:         true,
		showScrollbar:      true,
		treeStyle:          style.Style{},
		selectedStyle:      style.Style{BG: style.Blue, FG: style.White},
		iconStyle:          style.Style{FG: style.Yellow},
		scrollbarStyle:     style.Style{}.Foreground(style.BrightBlack),
		matchStyle:         style.Style{},
		scrollOffset:       0,
		selectedIndex:      -1,
		viewportHeight:     10,
		expandedKeys:       nil,
		searchQuery:        "",
		searchStatsStyle:   style.Style{},
		selectionMode:      SelectionNone,
		checkedKeys:        nil,
		lazyLoadFn:         nil,
		lazyLoadChildrenFn: nil,
		allowScroll:        true,
		allowExpand:        true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "treeview" }
func (v *VNode) Style() style.Style                           { return v.treeStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.treeStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	props := rtui.Props{
		"key":                     v.key,
		"componentID":             v.componentID,
		"nodes":                   v.nodes,
		"expandLevel":             v.expandLevel,
		"showIcons":               v.showIcons,
		"showLineNums":            v.showLineNums,
		"compact":                 v.compact,
		"showBorder":              v.showBorder,
		"showScrollbar":           v.showScrollbar,
		"treeStyle":               v.treeStyle,
		"selectedStyle":           v.selectedStyle,
		"iconStyle":               v.iconStyle,
		"scrollbarStyle":          v.scrollbarStyle,
		"scrollOffset":            v.scrollOffset,
		"scrollOffsetControlled":  v.scrollOffsetControlled,
		"selectedIndex":           v.selectedIndex,
		"selectedIndexControlled": v.selectedIndexControlled,
		"viewportHeight":          v.viewportHeight,
		"expandedKeysControlled":  v.expandedKeysControlled,
		"searchQuery":             v.searchQuery,
		"searchQueryControlled":   v.searchQueryControlled,
		"matchStyle":              v.matchStyle,
		"selectionMode":           v.selectionMode,
		"checkedKeysControlled":   v.checkedKeysControlled,
		"selectionIntent":         v.selectionIntent,
		"selectionIntentField":    v.selectionIntentField,
		"lazyLoadFn":              v.lazyLoadFn,
		"lazyLoadChildrenFn":      v.lazyLoadChildrenFn,
		"showSearchStats":         v.showSearchStats,
		"searchStatsStyle":        v.searchStatsStyle,
		"allowScroll":             v.allowScroll,
		"allowExpand":             v.allowExpand,
	}
	if v.expandedKeysControlled {
		props["expandedKeys"] = cloneExpandedKeys(v.expandedKeys)
	}
	if v.checkedKeysControlled {
		props["checkedKeys"] = cloneExpandedKeys(v.checkedKeys)
	}
	if v.rowStyleFn != nil {
		props["rowStyleFn"] = v.rowStyleFn
	}
	if v.searchFn != nil {
		props["searchFn"] = v.searchFn
	}
	if v.showSearchStats {
		props["showSearchStats"] = v.showSearchStats
	}
	if v.searchStatsStyle != (style.Style{}) {
		props["searchStatsStyle"] = v.searchStatsStyle
	}
	if v.lazyLoadFn != nil {
		props["lazyLoadFn"] = v.lazyLoadFn
	}
	if v.lazyLoadChildrenFn != nil {
		props["lazyLoadChildrenFn"] = v.lazyLoadChildrenFn
	}
	return props
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
	if showBorder, ok := p["showBorder"].(bool); ok {
		v.showBorder = showBorder
	}
	if showScrollbar, ok := p["showScrollbar"].(bool); ok {
		v.showScrollbar = showScrollbar
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
	if scrollbarStyle, ok := p["scrollbarStyle"].(style.Style); ok {
		v.scrollbarStyle = scrollbarStyle
	}
	if fn, ok := p["rowStyleFn"].(func(int, TreeNode) style.Style); ok {
		v.rowStyleFn = fn
	}
	if matchStyle, ok := p["matchStyle"].(style.Style); ok {
		v.matchStyle = matchStyle
	}
	if scrollOffset, ok := p["scrollOffset"].(int); ok {
		v.scrollOffset = scrollOffset
	}
	if controlled, ok := p["scrollOffsetControlled"].(bool); ok {
		v.scrollOffsetControlled = controlled
	}
	if selectedIndex, ok := p["selectedIndex"].(int); ok {
		v.selectedIndex = selectedIndex
	}
	if controlled, ok := p["selectedIndexControlled"].(bool); ok {
		v.selectedIndexControlled = controlled
	}
	if viewportHeight, ok := p["viewportHeight"].(int); ok {
		v.viewportHeight = viewportHeight
	}
	if expandedKeys, ok := p["expandedKeys"].(map[string]bool); ok {
		v.expandedKeys = cloneExpandedKeys(expandedKeys)
		v.expandedKeysControlled = true
	}
	if controlled, ok := p["expandedKeysControlled"].(bool); ok {
		v.expandedKeysControlled = controlled
	}
	if searchQuery, ok := p["searchQuery"].(string); ok {
		v.searchQuery = searchQuery
	}
	if controlled, ok := p["searchQueryControlled"].(bool); ok {
		v.searchQueryControlled = controlled
	}
	if searchFn, ok := p["searchFn"].(func(TreeNode, string) bool); ok {
		v.searchFn = searchFn
	}
	if selectionMode, ok := p["selectionMode"].(SelectionMode); ok {
		v.selectionMode = selectionMode
	}
	if checkedKeys, ok := p["checkedKeys"].(map[string]bool); ok {
		v.checkedKeys = cloneExpandedKeys(checkedKeys)
		v.checkedKeysControlled = true
	}
	if controlled, ok := p["checkedKeysControlled"].(bool); ok {
		v.checkedKeysControlled = controlled
	}
	if selectionIntent, ok := p["selectionIntent"].(intent.Intent); ok {
		v.selectionIntent = selectionIntent
	}
	if selectionIntentField, ok := p["selectionIntentField"].(intent.FieldIntent); ok {
		v.selectionIntentField = selectionIntentField
	}
	if lazyLoadFn, ok := p["lazyLoadFn"].(func(TreeNode)); ok {
		v.lazyLoadFn = lazyLoadFn
	}
	if lazyLoadChildrenFn, ok := p["lazyLoadChildrenFn"].(func(TreeNode) []TreeNode); ok {
		v.lazyLoadChildrenFn = lazyLoadChildrenFn
	}
	if showSearchStats, ok := p["showSearchStats"].(bool); ok {
		v.showSearchStats = showSearchStats
	}
	if searchStatsStyle, ok := p["searchStatsStyle"].(style.Style); ok {
		v.searchStatsStyle = searchStatsStyle
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

func (v *VNode) SetNodes(nodes []TreeNode) *VNode       { v.nodes = nodes; return v }
func (v *VNode) SetComponentID(id string) *VNode        { v.componentID = id; return v }
func (v *VNode) SetExpandLevel(level int) *VNode        { v.expandLevel = level; return v }
func (v *VNode) SetShowIcons(show bool) *VNode          { v.showIcons = show; return v }
func (v *VNode) SetShowLineNums(show bool) *VNode       { v.showLineNums = show; return v }
func (v *VNode) SetCompact(compact bool) *VNode         { v.compact = compact; return v }
func (v *VNode) SetShowBorder(show bool) *VNode         { v.showBorder = show; return v }
func (v *VNode) SetShowScrollbar(show bool) *VNode      { v.showScrollbar = show; return v }
func (v *VNode) SetTreeStyle(s style.Style) *VNode      { v.treeStyle = s; return v }
func (v *VNode) SetSelectedStyle(s style.Style) *VNode  { v.selectedStyle = s; return v }
func (v *VNode) SetIconStyle(s style.Style) *VNode      { v.iconStyle = s; return v }
func (v *VNode) SetScrollbarStyle(s style.Style) *VNode { v.scrollbarStyle = s; return v }
func (v *VNode) SetRowStyleFn(fn func(int, TreeNode) style.Style) *VNode {
	v.rowStyleFn = fn
	return v
}
func (v *VNode) SetMatchStyle(s style.Style) *VNode { v.matchStyle = s; return v }
func (v *VNode) SetScrollOffset(offset int) *VNode  { v.scrollOffset = offset; return v }
func (v *VNode) SetScrollOffsetControlled(offset int) *VNode {
	v.scrollOffset = offset
	v.scrollOffsetControlled = true
	return v
}
func (v *VNode) SetSelectedIndex(index int) *VNode { v.selectedIndex = index; return v }
func (v *VNode) SetSelectedIndexControlled(index int) *VNode {
	v.selectedIndex = index
	v.selectedIndexControlled = true
	return v
}
func (v *VNode) SetViewportHeight(height int) *VNode { v.viewportHeight = height; return v }
func (v *VNode) SetExpandedKeys(keys map[string]bool) *VNode {
	v.expandedKeys = cloneExpandedKeys(keys)
	v.expandedKeysControlled = true
	return v
}
func (v *VNode) SetExpandedPaths(paths ...string) *VNode {
	if len(paths) == 0 {
		v.expandedKeys = nil
		v.expandedKeysControlled = true
		return v
	}
	keys := make(map[string]bool, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		keys[path] = true
	}
	v.expandedKeys = keys
	v.expandedKeysControlled = true
	return v
}
func (v *VNode) SetSearchQuery(query string) *VNode {
	v.searchQuery = query
	return v
}
func (v *VNode) SetSearchQueryControlled(query string) *VNode {
	v.searchQuery = query
	v.searchQueryControlled = true
	return v
}
func (v *VNode) SetSearchFn(fn func(TreeNode, string) bool) *VNode {
	v.searchFn = fn
	return v
}
func (v *VNode) SetSelectionMode(mode SelectionMode) *VNode { v.selectionMode = mode; return v }
func (v *VNode) SetCheckedKeys(keys map[string]bool) *VNode {
	v.checkedKeys = cloneExpandedKeys(keys)
	v.checkedKeysControlled = true
	return v
}
func (v *VNode) SetCheckedPaths(paths ...string) *VNode {
	if len(paths) == 0 {
		v.checkedKeys = nil
		v.checkedKeysControlled = true
		return v
	}
	keys := make(map[string]bool, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		keys["path:"+path] = true
	}
	v.checkedKeys = keys
	v.checkedKeysControlled = true
	return v
}
func (v *VNode) SetSelectionIntent(selectionIntent intent.Intent) *VNode {
	v.selectionIntent = selectionIntent
	return v
}
func (v *VNode) SetSelectionFieldIntent(selectionIntentField intent.FieldIntent) *VNode {
	v.selectionIntentField = selectionIntentField
	return v
}
func (v *VNode) SetLazyLoadFn(fn func(TreeNode)) *VNode {
	v.lazyLoadFn = fn
	return v
}
func (v *VNode) SetLazyLoadChildrenFn(fn func(TreeNode) []TreeNode) *VNode {
	v.lazyLoadChildrenFn = fn
	return v
}
func (v *VNode) SetShowSearchStats(show bool) *VNode {
	v.showSearchStats = show
	return v
}
func (v *VNode) SetSearchStatsStyle(s style.Style) *VNode {
	v.searchStatsStyle = s
	return v
}
func (v *VNode) SetAllowScroll(allow bool) *VNode { v.allowScroll = allow; return v }
func (v *VNode) SetAllowExpand(allow bool) *VNode { v.allowExpand = allow; return v }

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
	pathStack := make([]string, 0, 8)

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

		depth := indent / 4
		if depth < 0 {
			depth = 0
		}
		if depth > len(pathStack) {
			depth = len(pathStack)
		}
		pathStack = pathStack[:depth]
		segment := strings.TrimSuffix(content, "/")
		if segment == "" {
			segment = content
		}
		pathStack = append(pathStack, segment)
		path := strings.Join(pathStack, "/")

		nodes = append(nodes, TreeNode{
			Indent:   indent,
			Content:  content,
			Path:     path,
			NodeType: nodeType,
			NodeID:   i,
		})
	}

	return nodes
}

func cloneExpandedKeys(keys map[string]bool) map[string]bool {
	if len(keys) == 0 {
		return nil
	}
	cloned := make(map[string]bool, len(keys))
	for key, value := range keys {
		cloned[key] = value
	}
	return cloned
}
