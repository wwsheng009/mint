package treeview

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

const indentStep = 4

type nodeEntry struct {
	Index          int
	Node           TreeNode
	Depth          int
	HasChildren    bool
	HasDescendants bool
	Key            string
	Match          bool
}

// Instance is the runtime entity for TreeView components
type Instance struct {
	// === Identification ===
	key         string
	componentID string // Component ID for Intent routing (Phase 10)

	// === Props (from VNode, may change each render) ===
	nodes                   []TreeNode
	expandLevel             int
	showIcons               bool
	showLineNums            bool
	compact                 bool
	showBorder              bool
	showScrollbar           bool
	treeStyle               style.Style
	selectedStyle           style.Style
	iconStyle               style.Style
	scrollbarStyle          style.Style
	rowStyleFn              func(int, TreeNode) style.Style
	matchStyle              style.Style
	searchQuery             string
	searchQueryControlled   bool
	searchFn                func(TreeNode, string) bool
	selectionMode           SelectionMode
	checkedKeys             map[string]bool
	checkedKeysControlled   bool
	selectionIntent         intent.Intent
	selectionIntentField    intent.FieldIntent
	lazyLoadFn              func(TreeNode)
	lazyLoadChildrenFn      func(TreeNode) []TreeNode
	showSearchStats         bool
	searchStatsStyle        style.Style
	scrollOffset            int
	scrollOffsetControlled  bool
	selectedIndex           int
	selectedIndexControlled bool
	viewportHeight          int
	expandedKeys            map[string]bool
	expandedKeysControlled  bool
	allowScroll             bool
	allowExpand             bool

	// === Runtime State ===
	focused                  bool
	expandState              map[string]bool // Expand/collapse state for each node (keyed by nodeKey)
	lazyRequested            map[string]bool
	lazyInsertions           map[string][]TreeNode
	lastSearchTotal          int
	lastSearchSelected       int
	lastSearchQuery          string
	bounds                   [4]int // x, y, w, h
	dirty                    bool
	cacheDirty               bool
	entryCache               []nodeEntry
	visibleCache             []nodeEntry
	visibleIndexCache        []int
	scrollOffsetInitialized  bool
	selectedIndexInitialized bool
	checkedKeysInitialized   bool
	autoSelectMatch          bool
	searchQueryInitialized   bool
	selectionAnchorKey       string

	// === Intent Support (Phase 10) ===
	intentEmitter func(intent.Intent) // Intent emitter for bubbling
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
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
		key:                     getStringProp(props, "key", ""),
		componentID:             getStringProp(props, "componentID", ""),
		nodes:                   getNodesProp(props, []TreeNode{}),
		expandLevel:             getIntProp(props, "expandLevel", 1),
		showIcons:               getBoolProp(props, "showIcons", true),
		showLineNums:            getBoolProp(props, "showLineNums", false),
		compact:                 getBoolProp(props, "compact", false),
		showBorder:              getBoolProp(props, "showBorder", true),
		showScrollbar:           getBoolProp(props, "showScrollbar", true),
		treeStyle:               getStyleProp(props, "treeStyle"),
		selectedStyle:           getStyleProp(props, "selectedStyle"),
		iconStyle:               getStyleProp(props, "iconStyle"),
		scrollbarStyle:          getStyleProp(props, "scrollbarStyle"),
		rowStyleFn:              getRowStyleFn(props),
		matchStyle:              getStyleProp(props, "matchStyle"),
		searchQuery:             getStringProp(props, "searchQuery", ""),
		searchQueryControlled:   getBoolProp(props, "searchQueryControlled", false),
		searchFn:                getSearchFn(props),
		selectionMode:           getSelectionModeProp(props, "selectionMode", SelectionNone),
		checkedKeys:             normalizeNodeKeys(getCheckedKeysProp(props)),
		checkedKeysControlled:   getBoolProp(props, "checkedKeysControlled", false),
		selectionIntent:         getIntentProp(props, "selectionIntent"),
		selectionIntentField:    getFieldIntentProp(props, "selectionIntentField"),
		lazyLoadFn:              getLazyLoadFn(props),
		lazyLoadChildrenFn:      getLazyLoadChildrenFn(props),
		showSearchStats:         getBoolProp(props, "showSearchStats", false),
		searchStatsStyle:        getStyleProp(props, "searchStatsStyle"),
		scrollOffset:            getIntProp(props, "scrollOffset", 0),
		scrollOffsetControlled:  getBoolProp(props, "scrollOffsetControlled", false),
		selectedIndex:           getIntProp(props, "selectedIndex", -1),
		selectedIndexControlled: getBoolProp(props, "selectedIndexControlled", false),
		viewportHeight:          getIntProp(props, "viewportHeight", 10),
		expandedKeys:            getExpandedKeysProp(props),
		expandedKeysControlled:  getBoolProp(props, "expandedKeysControlled", false),
		allowScroll:             getBoolProp(props, "allowScroll", true),
		allowExpand:             getBoolProp(props, "allowExpand", true),
		expandState:             make(map[string]bool),
		lazyRequested:           make(map[string]bool),
		lazyInsertions:          make(map[string][]TreeNode),
		dirty:                   true,
		cacheDirty:              true,
	}

	if inst.expandedKeys != nil && len(inst.expandedKeys) > 0 {
		inst.expandedKeysControlled = true
	}
	if inst.checkedKeys != nil && len(inst.checkedKeys) > 0 {
		inst.checkedKeysControlled = true
	}
	inst.scrollOffsetInitialized = inst.scrollOffsetControlled || hasProp(props, "scrollOffset")
	inst.selectedIndexInitialized = inst.selectedIndexControlled || hasProp(props, "selectedIndex")
	inst.checkedKeysInitialized = inst.checkedKeysControlled || hasProp(props, "checkedKeys")
	inst.searchQueryInitialized = inst.searchQueryControlled || hasProp(props, "searchQuery")

	// Initialize expand state based on expandLevel or controlled keys
	if inst.expandedKeysControlled {
		inst.applyExpandedKeys(inst.expandedKeys, true)
	} else {
		inst.syncExpandState(true)
	}

	inst.normalizeCheckedKeys()

	// Clamp selection and scroll to visible range
	inst.normalizeSelectionAndScroll()

	return inst
}

// syncExpandState initializes or updates expand states based on expandLevel.
// If reset is true, all nodes are reset to defaults. Otherwise, existing states are preserved.
func (inst *Instance) syncExpandState(reset bool) {
	next := make(map[string]bool, len(inst.nodes))
	for i, node := range inst.nodes {
		key := nodeKey(node, i)
		if !reset {
			if existing, ok := inst.expandState[key]; ok {
				next[key] = existing
				continue
			}
		}

		depth := nodeDepth(node)
		if inst.expandLevel < 0 {
			next[key] = true
		} else if depth < inst.expandLevel {
			next[key] = true
		} else {
			next[key] = false
		}
	}
	inst.expandState = next
	inst.invalidateCache()
}

func (inst *Instance) invalidateCache() {
	inst.cacheDirty = true
}

func (inst *Instance) applyExpandedKeys(keys map[string]bool, reset bool) {
	normalized := normalizeExpandedKeys(keys)
	if inst.expandedKeysControlled {
		inst.expandedKeys = cloneExpandedKeys(normalized)
	}
	next := make(map[string]bool, len(inst.nodes))
	for i, node := range inst.nodes {
		key := nodeKey(node, i)
		if value, ok := normalized[key]; ok {
			next[key] = value
			continue
		}
		if !reset {
			if existing, ok := inst.expandState[key]; ok {
				next[key] = existing
				continue
			}
		}

		depth := nodeDepth(node)
		if inst.expandLevel < 0 {
			next[key] = true
		} else if depth < inst.expandLevel {
			next[key] = true
		} else {
			next[key] = false
		}
	}
	inst.expandState = next
	inst.invalidateCache()
}

func (inst *Instance) setAllExpanded(expand bool) {
	entries := inst.buildNodeEntries()
	next := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.HasChildren {
			next[entry.Key] = expand
		} else {
			next[entry.Key] = false
		}
	}
	inst.expandState = next
	if expand {
		for _, entry := range entries {
			inst.maybeEmitLazyLoad(entry)
		}
	}
	if inst.expandedKeysControlled {
		inst.expandedKeys = cloneExpandedKeys(next)
	}
	inst.invalidateCache()
	inst.normalizeSelectionAndScroll()
	inst.dirty = true
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
	oldComponentID := inst.componentID
	oldNodes := append([]TreeNode(nil), inst.nodes...)
	oldExpandLevel := inst.expandLevel
	oldShowIcons := inst.showIcons
	oldShowLineNums := inst.showLineNums
	oldCompact := inst.compact
	oldShowBorder := inst.showBorder
	oldShowScrollbar := inst.showScrollbar
	oldTreeStyle := inst.treeStyle
	oldSelectedStyle := inst.selectedStyle
	oldIconStyle := inst.iconStyle
	oldScrollbarStyle := inst.scrollbarStyle
	oldRowStyleFn := inst.rowStyleFn
	oldMatchStyle := inst.matchStyle
	oldSearchQuery := inst.searchQuery
	oldSearchFn := inst.searchFn
	oldSelectionMode := inst.selectionMode
	oldCheckedKeys := cloneExpandedKeys(inst.checkedKeys)
	oldCheckedKeysControlled := inst.checkedKeysControlled
	oldSelectionIntent := inst.selectionIntent
	oldSelectionIntentField := inst.selectionIntentField
	oldLazyLoadFn := inst.lazyLoadFn
	oldLazyLoadChildrenFn := inst.lazyLoadChildrenFn
	oldSearchQueryControlled := inst.searchQueryControlled
	oldShowSearchStats := inst.showSearchStats
	oldSearchStatsStyle := inst.searchStatsStyle
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	oldScrollControlled := inst.scrollOffsetControlled
	oldSelectedControlled := inst.selectedIndexControlled
	oldViewportHeight := inst.viewportHeight
	oldExpandedKeys := cloneExpandedKeys(inst.expandedKeys)
	oldExpandedKeysControlled := inst.expandedKeysControlled
	oldAllowScroll := inst.allowScroll
	oldAllowExpand := inst.allowExpand

	inst.componentID = getStringProp(props, "componentID", inst.componentID)
	inst.nodes = getNodesProp(props, inst.nodes)
	inst.expandLevel = getIntProp(props, "expandLevel", inst.expandLevel)
	inst.showIcons = getBoolProp(props, "showIcons", inst.showIcons)
	inst.showLineNums = getBoolProp(props, "showLineNums", inst.showLineNums)
	inst.compact = getBoolProp(props, "compact", inst.compact)
	inst.showBorder = getBoolProp(props, "showBorder", inst.showBorder)
	inst.showScrollbar = getBoolProp(props, "showScrollbar", inst.showScrollbar)
	inst.treeStyle = getStyleProp(props, "treeStyle")
	inst.selectedStyle = getStyleProp(props, "selectedStyle")
	inst.iconStyle = getStyleProp(props, "iconStyle")
	inst.scrollbarStyle = getStyleProp(props, "scrollbarStyle")
	if fn, ok := props["rowStyleFn"].(func(int, TreeNode) style.Style); ok {
		inst.rowStyleFn = fn
	} else if _, exists := props["rowStyleFn"]; exists {
		inst.rowStyleFn = nil
	}
	inst.matchStyle = getStyleProp(props, "matchStyle")
	if controlled, ok := props["searchQueryControlled"].(bool); ok {
		inst.searchQueryControlled = controlled
	}
	if inst.searchQueryControlled {
		inst.searchQuery = getStringProp(props, "searchQuery", inst.searchQuery)
		inst.searchQueryInitialized = true
	} else if query, ok := props["searchQuery"].(string); ok && !inst.searchQueryInitialized {
		inst.searchQuery = query
		inst.searchQueryInitialized = true
	}
	if fn, ok := props["searchFn"].(func(TreeNode, string) bool); ok {
		inst.searchFn = fn
	} else if _, exists := props["searchFn"]; exists {
		inst.searchFn = nil
	}
	inst.selectionMode = getSelectionModeProp(props, "selectionMode", inst.selectionMode)
	inst.selectionIntent = getIntentProp(props, "selectionIntent")
	inst.selectionIntentField = getFieldIntentProp(props, "selectionIntentField")
	inst.showSearchStats = getBoolProp(props, "showSearchStats", inst.showSearchStats)
	inst.searchStatsStyle = getStyleProp(props, "searchStatsStyle")
	if fn, ok := props["lazyLoadFn"].(func(TreeNode)); ok {
		inst.lazyLoadFn = fn
	} else if _, exists := props["lazyLoadFn"]; exists {
		inst.lazyLoadFn = nil
	}
	if fn, ok := props["lazyLoadChildrenFn"].(func(TreeNode) []TreeNode); ok {
		inst.lazyLoadChildrenFn = fn
	} else if _, exists := props["lazyLoadChildrenFn"]; exists {
		inst.lazyLoadChildrenFn = nil
	}
	if controlled, ok := props["scrollOffsetControlled"].(bool); ok {
		inst.scrollOffsetControlled = controlled
	}
	if inst.scrollOffsetControlled {
		inst.scrollOffset = getIntProp(props, "scrollOffset", inst.scrollOffset)
		inst.scrollOffsetInitialized = true
	} else if offset, ok := props["scrollOffset"].(int); ok && !inst.scrollOffsetInitialized {
		inst.scrollOffset = offset
		inst.scrollOffsetInitialized = true
	}
	if controlled, ok := props["selectedIndexControlled"].(bool); ok {
		inst.selectedIndexControlled = controlled
	}
	if inst.selectedIndexControlled {
		inst.selectedIndex = getIntProp(props, "selectedIndex", inst.selectedIndex)
		inst.selectedIndexInitialized = true
	} else if index, ok := props["selectedIndex"].(int); ok && !inst.selectedIndexInitialized {
		inst.selectedIndex = index
		inst.selectedIndexInitialized = true
	}
	if controlled, ok := props["checkedKeysControlled"].(bool); ok {
		inst.checkedKeysControlled = controlled
	}
	if inst.checkedKeysControlled {
		inst.checkedKeys = normalizeNodeKeys(getCheckedKeysProp(props))
		inst.checkedKeysInitialized = true
	} else if _, ok := props["checkedKeys"].(map[string]bool); ok && !inst.checkedKeysInitialized {
		inst.checkedKeys = normalizeNodeKeys(getCheckedKeysProp(props))
		inst.checkedKeysInitialized = true
	}
	inst.viewportHeight = getIntProp(props, "viewportHeight", inst.viewportHeight)
	if expandedKeys, ok := props["expandedKeys"].(map[string]bool); ok {
		inst.expandedKeys = cloneExpandedKeys(expandedKeys)
		inst.expandedKeysControlled = true
	}
	if controlled, ok := props["expandedKeysControlled"].(bool); ok {
		inst.expandedKeysControlled = controlled
	}
	inst.allowScroll = getBoolProp(props, "allowScroll", inst.allowScroll)
	inst.allowExpand = getBoolProp(props, "allowExpand", inst.allowExpand)

	lazyInserted := inst.reapplyLazyInsertions()
	nodesChanged := !nodesEqual(oldNodes, inst.nodes) || lazyInserted
	expandLevelChanged := oldExpandLevel != inst.expandLevel
	expandedKeysChanged := !equalExpandedKeys(oldExpandedKeys, inst.expandedKeys) || oldExpandedKeysControlled != inst.expandedKeysControlled
	checkedKeysChanged := !equalExpandedKeys(oldCheckedKeys, inst.checkedKeys) || oldCheckedKeysControlled != inst.checkedKeysControlled
	searchChanged := oldSearchQuery != inst.searchQuery || !sameSearchFn(oldSearchFn, inst.searchFn)
	if searchChanged && !inst.selectedIndexControlled {
		inst.autoSelectMatch = true
	}
	if inst.expandedKeysControlled {
		if nodesChanged || expandLevelChanged || expandedKeysChanged {
			inst.applyExpandedKeys(inst.expandedKeys, true)
		}
	} else if nodesChanged || expandLevelChanged {
		inst.syncExpandState(expandLevelChanged)
	}
	if nodesChanged || expandLevelChanged || expandedKeysChanged || searchChanged {
		inst.invalidateCache()
	}

	inst.normalizeCheckedKeys()
	inst.normalizeSelectionAndScroll()

	changed := oldComponentID != inst.componentID ||
		nodesChanged ||
		oldExpandLevel != inst.expandLevel ||
		oldShowIcons != inst.showIcons ||
		oldShowLineNums != inst.showLineNums ||
		oldCompact != inst.compact ||
		oldShowBorder != inst.showBorder ||
		oldShowScrollbar != inst.showScrollbar ||
		oldTreeStyle != inst.treeStyle ||
		oldSelectedStyle != inst.selectedStyle ||
		oldIconStyle != inst.iconStyle ||
		oldScrollbarStyle != inst.scrollbarStyle ||
		!sameRowStyleFn(oldRowStyleFn, inst.rowStyleFn) ||
		oldMatchStyle != inst.matchStyle ||
		searchChanged ||
		oldSelectionMode != inst.selectionMode ||
		checkedKeysChanged ||
		!sameIntent(oldSelectionIntent, inst.selectionIntent) ||
		!sameFieldIntent(oldSelectionIntentField, inst.selectionIntentField) ||
		!sameLazyLoadFn(oldLazyLoadFn, inst.lazyLoadFn) ||
		!sameLazyLoadChildrenFn(oldLazyLoadChildrenFn, inst.lazyLoadChildrenFn) ||
		oldSearchQueryControlled != inst.searchQueryControlled ||
		oldShowSearchStats != inst.showSearchStats ||
		oldSearchStatsStyle != inst.searchStatsStyle ||
		oldSelected != inst.selectedIndex ||
		oldScroll != inst.scrollOffset ||
		oldScrollControlled != inst.scrollOffsetControlled ||
		oldSelectedControlled != inst.selectedIndexControlled ||
		oldViewportHeight != inst.viewportHeight ||
		expandedKeysChanged ||
		oldAllowScroll != inst.allowScroll ||
		oldAllowExpand != inst.allowExpand
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	props := rtui.Props{
		"key":                     inst.key,
		"componentID":             inst.componentID,
		"nodes":                   inst.nodes,
		"expandLevel":             inst.expandLevel,
		"showIcons":               inst.showIcons,
		"showLineNums":            inst.showLineNums,
		"compact":                 inst.compact,
		"showBorder":              inst.showBorder,
		"showScrollbar":           inst.showScrollbar,
		"matchStyle":              inst.matchStyle,
		"searchQuery":             inst.searchQuery,
		"searchQueryControlled":   inst.searchQueryControlled,
		"selectionMode":           inst.selectionMode,
		"checkedKeysControlled":   inst.checkedKeysControlled,
		"selectionIntent":         inst.selectionIntent,
		"selectionIntentField":    inst.selectionIntentField,
		"lazyLoadFn":              inst.lazyLoadFn,
		"lazyLoadChildrenFn":      inst.lazyLoadChildrenFn,
		"showSearchStats":         inst.showSearchStats,
		"searchStatsStyle":        inst.searchStatsStyle,
		"scrollOffsetControlled":  inst.scrollOffsetControlled,
		"selectedIndexControlled": inst.selectedIndexControlled,
		"scrollOffset":            inst.scrollOffset,
		"selectedIndex":           inst.selectedIndex,
		"viewportHeight":          inst.viewportHeight,
		"expandedKeysControlled":  inst.expandedKeysControlled,
		"allowScroll":             inst.allowScroll,
		"allowExpand":             inst.allowExpand,
	}
	if inst.expandedKeysControlled {
		props["expandedKeys"] = cloneExpandedKeys(inst.expandedKeys)
	}
	if inst.checkedKeysControlled {
		props["checkedKeys"] = cloneExpandedKeys(inst.checkedKeys)
	}
	return props
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
	visible, _ := inst.visibleEntries()
	contentWidth := inst.calculateContentWidth(visible)
	if inst.showSearchStats {
		statsWidth := paint.StringWidth(inst.searchStatsLine())
		if statsWidth > contentWidth {
			contentWidth = statsWidth
		}
	}
	width := contentWidth
	scrollbarVisible := inst.showScrollbar && len(visible) > inst.desiredViewportHeight(len(visible))
	if inst.showBorder {
		width += 4
	} else if scrollbarVisible {
		width += 1
	}
	height := inst.calculateHeight(len(visible))

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// calculateHeight calculates the total height of the tree
func (inst *Instance) calculateHeight(visibleCount int) int {
	height := inst.desiredViewportHeight(visibleCount)
	if inst.showBorder {
		height += 2
	}
	height += inst.statsHeight()
	if height < 1 {
		height = 1
	}
	return height
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the tree view
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	visible, _ := inst.visibleEntries()
	visibleCount := len(visible)
	viewSize := inst.effectiveViewportHeight(visibleCount)
	viewport := scrollutil.NewVerticalViewport(visibleCount, viewSize, inst.scrollOffset)
	scrollbarVisible := inst.showScrollbar && viewport.IsScrollable()
	startLine, endLine := viewport.VisibleRange()

	contentWidth := inst.calculateContentWidth(visible)
	statsLine := inst.searchStatsLine()
	if inst.showSearchStats {
		statsWidth := paint.StringWidth(statsLine)
		if statsWidth > contentWidth {
			contentWidth = statsWidth
		}
	}
	if contentWidth < 1 {
		contentWidth = 1
	}
	width := contentWidth
	if inst.showBorder {
		width += 4
	} else if scrollbarVisible {
		width += 1
	}

	cmds := make([]paint.DrawCmd, 0, viewSize+2)
	borderStyle := inst.treeStyle
	borderHorizontal := "─"
	borderVertical := "│"
	borderTopLeft := "┌"
	borderTopRight := "┐"
	borderBottomLeft := "└"
	borderBottomRight := "┘"
	focusLabel := ""
	if inst.focused {
		borderStyle = borderStyle.Bold(true)
		borderHorizontal = "═"
		borderVertical = "║"
		borderTopLeft = "╔"
		borderTopRight = "╗"
		borderBottomLeft = "╚"
		borderBottomRight = "╝"
		focusLabel = " [FOCUS] "
	}

	if inst.showBorder {
		topBorder := borderTopLeft + inst.borderInnerText(width-2, borderHorizontal, focusLabel) + borderTopRight
		cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))
	}

	rowOffset := 0
	if inst.showBorder {
		rowOffset = 1
	}
	if inst.showSearchStats {
		statsText := padRightToWidth(truncateText(statsLine, contentWidth), contentWidth)
		statsStyle := inst.searchStatsStyle
		if statsStyle == (style.Style{}) {
			statsStyle = inst.treeStyle
		}
		if inst.showBorder {
			statsText = borderVertical + " " + statsText + " " + borderVertical
		}
		cmds = append(cmds, paint.NewTextCmd(x, y+rowOffset, statsText, statsStyle))
		rowOffset++
	}

	for i := startLine; i < endLine; i++ {
		entry := visible[i]
		rowY := y + rowOffset + (i - startLine)
		selected := i == inst.selectedIndex

		line, prefixWidth, icon, iconWidth, content := inst.composeLine(entry, contentWidth)
		rowStyle := inst.treeStyle
		if inst.rowStyleFn != nil {
			if override := inst.rowStyleFn(entry.Index, entry.Node); override != (style.Style{}) {
				rowStyle = override
			}
		}
		if entry.Match && inst.matchStyle != (style.Style{}) {
			rowStyle = inst.matchStyle
		}
		if selected {
			rowStyle = inst.selectedStyle
		}

		if inst.showBorder {
			line = borderVertical + " " + line + " " + borderVertical
		}
		cmds = append(cmds, paint.NewTextCmd(x, rowY, line, rowStyle))

		if entry.Match && inst.matchStyle != (style.Style{}) && inst.searchFn == nil {
			if highlightStart, highlightText, ok := inst.matchHighlight(content); ok {
				highlightX := x + prefixWidth + iconWidth + highlightStart
				if inst.showBorder {
					highlightX += 2
				}
				cmds = append(cmds, paint.NewTextCmd(highlightX, rowY, highlightText, inst.matchStyle))
			}
		}

		if !selected && inst.showIcons && iconWidth > 0 && inst.iconStyle != (style.Style{}) {
			iconX := x + prefixWidth
			if inst.showBorder {
				iconX += 2
			}
			// Only overlay icon if it fully fits in the content width.
			if contentWidth >= prefixWidth+iconWidth {
				cmds = append(cmds, paint.NewTextCmd(iconX, rowY, icon, inst.iconStyle))
			}
		}
	}

	// Fill remaining viewport rows (to clear previous content)
	for i := endLine; i < startLine+viewSize; i++ {
		rowY := y + rowOffset + (i - startLine)
		blank := strings.Repeat(" ", contentWidth)
		if inst.showBorder {
			blank = borderVertical + " " + blank + " " + borderVertical
		}
		cmds = append(cmds, paint.NewTextCmd(x, rowY, blank, inst.treeStyle))
	}

	if inst.showBorder {
		bottomBorder := borderBottomLeft + strings.Repeat(borderHorizontal, max(1, width-2)) + borderBottomRight
		cmds = append(cmds, paint.NewTextCmd(x, y+rowOffset+viewSize, bottomBorder, borderStyle))
	}

	if inst.showScrollbar {
		scrollbarStyle := inst.scrollbarStyle
		if scrollbarStyle == (style.Style{}) {
			scrollbarStyle = borderStyle
		} else if scrollbarStyle.FG == "" {
			scrollbarStyle = scrollbarStyle.Foreground(borderStyle.FG)
		}
		scrollbarX := x + max(1, width) - 1
		scrollbarY := y
		if inst.showBorder {
			scrollbarY++
		}
		scrollbarY += inst.statsHeight()
		cmds = append(cmds, scrollutil.DrawVerticalScrollbar(
			scrollbarX,
			scrollbarY,
			viewSize,
			viewport,
			scrollbarStyle,
			scrollutil.DefaultVerticalScrollbarConfig(),
		)...)
	}

	return cmds
}

func (inst *Instance) lineParts(entry nodeEntry) (prefix, icon, content string) {
	prefix = inst.indentPrefix(entry.Depth)
	prefix += inst.selectionMarker(entry)
	icon = inst.iconFor(entry)
	content = entry.Node.Content
	if inst.showLineNums {
		content += fmt.Sprintf(" [%d]", entry.Node.NodeID)
	}
	if suffix := inst.statusSuffix(entry); suffix != "" {
		content += suffix
	}
	return prefix, icon, content
}

func (inst *Instance) indentPrefix(depth int) string {
	if depth <= 0 {
		return ""
	}
	if inst.compact {
		return strings.Repeat("  ", depth)
	}
	return strings.Repeat("│   ", depth)
}

func (inst *Instance) iconFor(entry nodeEntry) string {
	if !inst.showIcons {
		return "  "
	}
	isFolder := entry.HasChildren || entry.Node.NodeType == "folder"
	if isFolder {
		if inst.isExpanded(entry.Index) {
			return "📂 "
		}
		return "📁 "
	}
	return "📄 "
}

func (inst *Instance) statusSuffix(entry nodeEntry) string {
	node := entry.Node
	if node.LoadError != "" {
		msg := trimToWidth(node.LoadError, 24)
		if msg != "" {
			return " [error: " + msg + "] [retry:R]"
		}
		return " [error] [retry:R]"
	}
	if node.Loading {
		return " [loading]"
	}
	if node.Lazy && !entry.HasDescendants {
		return " [load:R]"
	}
	return ""
}

func (inst *Instance) matchHighlight(content string) (int, string, bool) {
	query := strings.TrimSpace(inst.searchQuery)
	if query == "" || content == "" {
		return 0, "", false
	}

	searchable := content
	if strings.HasSuffix(searchable, "...") {
		searchable = strings.TrimSuffix(searchable, "...")
	}
	start, length, ok := matchSpanRunes(searchable, query)
	if !ok || length <= 0 {
		return 0, "", false
	}
	runes := []rune(searchable)
	if start < 0 || start+length > len(runes) {
		return 0, "", false
	}
	prefix := string(runes[:start])
	match := string(runes[start : start+length])
	return paint.StringWidth(prefix), match, true
}

func matchSpanRunes(content, query string) (int, int, bool) {
	contentRunes := []rune(content)
	queryRunes := []rune(query)
	if len(queryRunes) == 0 || len(contentRunes) == 0 || len(queryRunes) > len(contentRunes) {
		return 0, 0, false
	}
	for i := 0; i <= len(contentRunes)-len(queryRunes); i++ {
		match := true
		for j := range queryRunes {
			if unicode.ToLower(contentRunes[i+j]) != unicode.ToLower(queryRunes[j]) {
				match = false
				break
			}
		}
		if match {
			return i, len(queryRunes), true
		}
	}
	return 0, 0, false
}

func (inst *Instance) selectionMarker(entry nodeEntry) string {
	if inst.selectionMode == SelectionNone {
		return ""
	}
	if inst.isChecked(entry) {
		return "[x] "
	}
	return "[ ] "
}

func (inst *Instance) selectionMarkerWidth() int {
	if inst.selectionMode == SelectionNone {
		return 0
	}
	return paint.StringWidth("[ ] ")
}

func (inst *Instance) isChecked(entry nodeEntry) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	return inst.checkedKeys != nil && inst.checkedKeys[entry.Key]
}

func (inst *Instance) composeLine(entry nodeEntry, maxWidth int) (line string, prefixWidth int, icon string, iconWidth int, content string) {
	prefix, icon, content := inst.lineParts(entry)
	prefixWidth = paint.StringWidth(prefix)
	iconWidth = paint.StringWidth(icon)

	available := maxWidth - prefixWidth - iconWidth
	if available < 0 {
		trimmed := trimToWidth(prefix+icon, maxWidth)
		return padRightToWidth(trimmed, maxWidth), prefixWidth, icon, iconWidth, ""
	}

	content = truncateText(content, available)
	line = prefix + icon + content
	line = padRightToWidth(line, maxWidth)
	return line, prefixWidth, icon, iconWidth, content
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) HandleAction(act *action.Action) bool {
	switch act.Type {
	case action.ActionScroll:
		if !inst.allowScroll {
			return false
		}
		if delta, ok := scrollutil.DeltaFromAction(act); ok {
			return inst.scrollBy(delta)
		}
		return false
	case action.ActionScrollUp:
		if !inst.allowScroll {
			return false
		}
		return inst.scrollBy(-1)
	case action.ActionScrollDown:
		if !inst.allowScroll {
			return false
		}
		return inst.scrollBy(1)
	case action.ActionClick:
		return inst.handleClick(act)
	case action.ActionDoubleClick:
		return inst.handleDoubleClick(act)
	case action.ActionNavigateUp:
		return inst.navigateUp()
	case action.ActionNavigateDown:
		return inst.navigateDown()
	case action.ActionNavigateLeft:
		return inst.navigateLeft()
	case action.ActionNavigateRight:
		return inst.navigateRight()
	case action.ActionNavigatePrev:
		return inst.navigateUp()
	case action.ActionNavigateNext:
		return inst.navigateDown()
	case action.ActionNavigateFirst:
		return inst.navigateHome()
	case action.ActionNavigateLast:
		return inst.navigateEnd()
	case action.ActionNavigateHome:
		return inst.navigateHome()
	case action.ActionNavigateEnd:
		return inst.navigateEnd()
	case action.ActionNavigatePageUp:
		return inst.pageUp()
	case action.ActionNavigatePageDown:
		return inst.pageDown()
	case action.ActionToggle:
		return inst.toggleExpand()
	case action.ActionSearch:
		if dir, ok := inst.searchDirectionFromAction(act); ok {
			return inst.navigateMatch(dir)
		}
		return false
	case action.ActionRefresh:
		return inst.refreshLazyByAction(act)
	case action.ActionInputText:
		return inst.handleInputShortcut(act)
	case action.ActionToggleSelect:
		return inst.toggleCheckedSelection(act)
	case action.ActionDeselectItem:
		return inst.deselectByAction(act)
	case action.ActionSelectRange:
		return inst.selectRangeByAction(act)
	case action.ActionSelectAll:
		return inst.selectAllByAction(act)
	case action.ActionClear:
		return inst.clearCheckedSelectionByAction(act)
	case action.ActionSelectItem:
		if index, ok := act.GetPayloadInt(); ok {
			visible, _ := inst.visibleEntries()
			return inst.selectVisibleIndex(index, visible, true)
		}
		return false
	case action.ActionSelect, action.ActionEnter:
		return inst.handleActivate()
	}
	return false
}

func (inst *Instance) borderInnerText(width int, fill, label string) string {
	if width <= 0 {
		return ""
	}
	if label == "" {
		return strings.Repeat(fill, width)
	}
	labelWidth := paint.StringWidth(label)
	if labelWidth >= width {
		return trimToWidth(label, width)
	}
	return label + strings.Repeat(fill, width-labelWidth)
}

// =============================================================================
// FocusableInstance Interface
// =============================================================================

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool { return inst.focused }

func (inst *Instance) IsDisabled() bool { return false }

// =============================================================================
// Navigation Methods
// =============================================================================

func (inst *Instance) navigateUp() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	fromIndex := inst.selectedIndex
	target := inst.selectedIndex
	if target < 0 {
		target = 0
	} else if target > 0 {
		target--
	} else {
		return false
	}
	if inst.selectVisibleIndex(target, visible, true) {
		inst.emitNavigation("up", fromIndex, inst.selectedIndex)
		return true
	}
	return false
}

func (inst *Instance) navigateDown() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	fromIndex := inst.selectedIndex
	target := inst.selectedIndex
	if target < 0 {
		target = 0
	} else if target < len(visible)-1 {
		target++
	} else {
		return false
	}
	if inst.selectVisibleIndex(target, visible, true) {
		inst.emitNavigation("down", fromIndex, inst.selectedIndex)
		return true
	}
	return false
}

func (inst *Instance) navigateLeft() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
		return inst.selectVisibleIndex(0, visible, true)
	}

	entry := visible[inst.selectedIndex]
	if entry.HasChildren && inst.expandState[entry.Key] {
		inst.expandState[entry.Key] = false
		inst.invalidateCache()
		if inst.expandedKeysControlled {
			inst.setExpandedKey(entry.Key, false)
		}
		inst.normalizeSelectionAndScroll()
		inst.dirty = true
		inst.emitNodeCollapse(entry.Index, entry.Node.Path, entry.Node.NodeID)
		return true
	}

	parentIndex := inst.parentVisibleIndex(visible, inst.selectedIndex)
	if parentIndex >= 0 {
		fromIndex := inst.selectedIndex
		if inst.selectVisibleIndex(parentIndex, visible, true) {
			inst.emitNavigation("left", fromIndex, inst.selectedIndex)
			return true
		}
	}

	return false
}

func (inst *Instance) navigateRight() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
		return inst.selectVisibleIndex(0, visible, true)
	}

	entry := visible[inst.selectedIndex]
	if entry.HasChildren {
		if !inst.expandState[entry.Key] {
			inst.expandState[entry.Key] = true
			inst.invalidateCache()
			if inst.expandedKeysControlled {
				inst.setExpandedKey(entry.Key, true)
			}
			inst.normalizeSelectionAndScroll()
			inst.dirty = true
			inst.emitNodeExpand(entry.Index, entry.Node.Path, entry.Node.NodeID)
			inst.maybeEmitLazyLoad(entry)
			return true
		}
		childIndex := inst.firstChildVisibleIndex(visible, inst.selectedIndex)
		if childIndex >= 0 {
			fromIndex := inst.selectedIndex
			if inst.selectVisibleIndex(childIndex, visible, true) {
				inst.emitNavigation("right", fromIndex, inst.selectedIndex)
				return true
			}
		}
	}
	return false
}

func (inst *Instance) navigateHome() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	fromIndex := inst.selectedIndex
	if inst.selectedIndex == 0 && inst.scrollOffset == 0 {
		return false
	}
	oldScroll := inst.scrollOffset
	viewport := inst.visibleViewport(len(visible))
	viewport.ScrollTo(0)
	inst.scrollOffset = viewport.Offset
	if inst.selectVisibleIndex(0, visible, true) {
		inst.emitNavigation("home", fromIndex, inst.selectedIndex)
		if inst.scrollOffset != oldScroll {
			inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
		}
		return true
	}
	return false
}

func (inst *Instance) navigateEnd() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	fromIndex := inst.selectedIndex
	viewport := inst.visibleViewport(len(visible))
	oldScroll := inst.scrollOffset
	viewport.ScrollTo(viewport.MaxOffset())
	inst.scrollOffset = viewport.Offset
	if inst.selectVisibleIndex(len(visible)-1, visible, true) {
		inst.emitNavigation("end", fromIndex, inst.selectedIndex)
		if inst.scrollOffset != oldScroll {
			inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
		}
		return true
	}
	return false
}

func (inst *Instance) pageUp() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	fromIndex := inst.selectedIndex
	oldScroll := inst.scrollOffset

	viewport := inst.visibleViewport(len(visible))
	viewport.PageUp()
	inst.scrollOffset = viewport.Offset

	target := inst.selectedIndex
	if target < 0 {
		target = 0
	} else {
		target = max(0, target-viewport.ViewSize)
	}

	changed := inst.selectVisibleIndex(target, visible, true)
	if changed || oldScroll != inst.scrollOffset {
		inst.dirty = true
		inst.emitNavigation("pageup", fromIndex, inst.selectedIndex)
		if inst.scrollOffset != oldScroll {
			inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
		}
		return true
	}
	return false
}

func (inst *Instance) pageDown() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	fromIndex := inst.selectedIndex
	oldScroll := inst.scrollOffset

	viewport := inst.visibleViewport(len(visible))
	viewport.PageDown()
	inst.scrollOffset = viewport.Offset

	target := inst.selectedIndex
	if target < 0 {
		target = min(len(visible)-1, max(0, viewport.ViewSize-1))
	} else {
		target = min(len(visible)-1, target+viewport.ViewSize)
	}

	changed := inst.selectVisibleIndex(target, visible, true)
	if changed || oldScroll != inst.scrollOffset {
		inst.dirty = true
		inst.emitNavigation("pagedown", fromIndex, inst.selectedIndex)
		if inst.scrollOffset != oldScroll {
			inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
		}
		return true
	}
	return false
}

func (inst *Instance) toggleExpand() bool {
	if !inst.allowExpand {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
		return inst.selectVisibleIndex(0, visible, true)
	}

	entry := visible[inst.selectedIndex]
	if !entry.HasChildren && entry.Node.NodeType != "folder" {
		return false
	}

	wasExpanded := inst.expandState[entry.Key]
	inst.expandState[entry.Key] = !wasExpanded
	nowExpanded := inst.expandState[entry.Key]
	inst.invalidateCache()
	if inst.expandedKeysControlled {
		inst.setExpandedKey(entry.Key, nowExpanded)
	}

	inst.normalizeSelectionAndScroll()
	inst.dirty = true

	// Emit Expand/Collapse Intent (Phase 10)
	if nowExpanded {
		inst.emitNodeExpand(entry.Index, entry.Node.Path, entry.Node.NodeID)
		inst.maybeEmitLazyLoad(entry)
	} else {
		inst.emitNodeCollapse(entry.Index, entry.Node.Path, entry.Node.NodeID)
	}

	return nowExpanded
}

func (inst *Instance) maybeEmitLazyLoad(entry nodeEntry) {
	inst.requestLazyLoad(entry, false)
}

func (inst *Instance) requestLazyLoad(entry nodeEntry, force bool) bool {
	if !force {
		if !entry.Node.Lazy {
			return false
		}
		if entry.HasDescendants {
			return false
		}
		if inst.lazyRequested[entry.Key] {
			return false
		}
	}
	inst.lazyRequested[entry.Key] = true
	if entry.Index >= 0 && entry.Index < len(inst.nodes) {
		inst.nodes[entry.Index].Loading = true
		inst.nodes[entry.Index].LoadError = ""
	}
	loaded := false
	if inst.lazyLoadChildrenFn != nil {
		children := inst.lazyLoadChildrenFn(entry.Node)
		if inst.insertLazyChildren(entry, children, false) {
			loaded = true
		}
	}
	if loaded {
		if entry.Index >= 0 && entry.Index < len(inst.nodes) {
			inst.nodes[entry.Index].Loading = false
			inst.nodes[entry.Index].Lazy = false
		}
		inst.normalizeSelectionAndScroll()
		inst.dirty = true
	}
	if inst.lazyLoadFn != nil {
		inst.lazyLoadFn(entry.Node)
	}
	inst.emitLazyLoad(entry.Index, entry.Node.Path, entry.Node.NodeID)
	return true
}

// =============================================================================
// Helper Methods
// =============================================================================

// getVisibleNodes returns nodes that should be visible based on expand state.
func (inst *Instance) getVisibleNodes() []TreeNode {
	visible, _ := inst.visibleEntries()
	nodes := make([]TreeNode, 0, len(visible))
	for _, entry := range visible {
		nodes = append(nodes, entry.Node)
	}
	return nodes
}

// visibleEntries returns visible nodes with metadata plus a map of node index -> visible index.
func (inst *Instance) visibleEntries() ([]nodeEntry, []int) {
	if !inst.cacheDirty && inst.visibleCache != nil {
		return inst.visibleCache, inst.visibleIndexCache
	}

	entries := inst.buildNodeEntries()
	visible := make([]nodeEntry, 0, len(entries))
	visibleIndex := make([]int, len(entries))
	for i := range visibleIndex {
		visibleIndex[i] = -1
	}

	query := strings.TrimSpace(inst.searchQuery)
	filterActive := query != ""
	match := make([]bool, len(entries))
	include := make([]bool, len(entries))
	if filterActive {
		stack := make([]int, 0, 8)
		for _, entry := range entries {
			for len(stack) > 0 && entries[stack[len(stack)-1]].Depth >= entry.Depth {
				stack = stack[:len(stack)-1]
			}
			if inst.nodeMatches(entry.Node, query) {
				match[entry.Index] = true
				include[entry.Index] = true
				for _, ancestor := range stack {
					include[ancestor] = true
				}
			}
			if entry.HasChildren {
				stack = append(stack, entry.Index)
			}
		}
	}

	type stackEntry struct {
		depth    int
		expanded bool
		visible  bool
	}

	stack := make([]stackEntry, 0, 8)
	for _, entry := range entries {
		for len(stack) > 0 && stack[len(stack)-1].depth >= entry.Depth {
			stack = stack[:len(stack)-1]
		}

		isVisible := true
		if len(stack) > 0 {
			top := stack[len(stack)-1]
			isVisible = top.visible && top.expanded
		}

		if filterActive && !include[entry.Index] {
			isVisible = false
		}

		if isVisible {
			entry.Match = filterActive && match[entry.Index]
			visibleIndex[entry.Index] = len(visible)
			visible = append(visible, entry)
		}

		if entry.HasChildren {
			expanded := inst.expandState[entry.Key]
			if filterActive && include[entry.Index] {
				expanded = true
			}
			stack = append(stack, stackEntry{
				depth:    entry.Depth,
				expanded: expanded,
				visible:  isVisible,
			})
		}
	}

	inst.entryCache = entries
	inst.visibleCache = visible
	inst.visibleIndexCache = visibleIndex
	inst.cacheDirty = false
	return visible, visibleIndex
}

// buildNodeEntries prepares metadata for each node.
func (inst *Instance) buildNodeEntries() []nodeEntry {
	entries := make([]nodeEntry, len(inst.nodes))
	for i, node := range inst.nodes {
		depth := nodeDepth(node)
		hasDescendants := false
		if i+1 < len(inst.nodes) {
			nextDepth := nodeDepth(inst.nodes[i+1])
			if nextDepth > depth {
				hasDescendants = true
			}
		}
		hasChildren := hasDescendants
		if node.Lazy {
			hasChildren = true
		}
		key := nodeKey(node, i)
		if hasDescendants || !node.Lazy {
			delete(inst.lazyRequested, key)
		}
		entries[i] = nodeEntry{
			Index:          i,
			Node:           node,
			Depth:          depth,
			HasChildren:    hasChildren,
			HasDescendants: hasDescendants,
			Key:            key,
		}
	}
	return entries
}

func (inst *Instance) insertLazyChildren(parent nodeEntry, children []TreeNode, replace bool) bool {
	if parent.Index < 0 || parent.Index >= len(inst.nodes) {
		return false
	}
	if replace {
		inst.removeDescendantsAt(parent.Index)
	} else if inst.hasDescendantsAt(parent.Index) {
		return false
	}

	normalized := inst.normalizeLazyChildren(parent, children)
	insertAt := parent.Index + 1
	for insertAt < len(inst.nodes) && nodeDepth(inst.nodes[insertAt]) > parent.Depth {
		insertAt++
	}
	if len(normalized) > 0 {
		inst.nodes = append(inst.nodes[:insertAt], append(normalized, inst.nodes[insertAt:]...)...)
	}
	changed := len(normalized) > 0 || replace || inst.nodes[parent.Index].Lazy || inst.nodes[parent.Index].Loading || inst.nodes[parent.Index].LoadError != ""
	inst.nodes[parent.Index].Lazy = false
	inst.nodes[parent.Index].Loading = false
	inst.nodes[parent.Index].LoadError = ""
	if inst.lazyInsertions != nil {
		if len(normalized) == 0 {
			delete(inst.lazyInsertions, parent.Key)
		} else {
			inst.lazyInsertions[parent.Key] = normalized
		}
	}
	if changed {
		inst.invalidateCache()
	}
	return changed
}

func (inst *Instance) normalizeLazyChildren(parent nodeEntry, children []TreeNode) []TreeNode {
	if len(children) == 0 {
		return nil
	}
	baseIndent := (parent.Depth + 1) * indentStep
	parentIndent := parent.Depth * indentStep
	normalized := make([]TreeNode, 0, len(children))
	for _, child := range children {
		c := child
		if c.Indent <= parentIndent {
			c.Indent = baseIndent + c.Indent
		} else if c.Indent < baseIndent {
			c.Indent = baseIndent
		}
		if c.Path == "" {
			segment := strings.TrimSuffix(c.Content, "/")
			if segment == "" {
				segment = c.Content
			}
			if parent.Node.Path != "" {
				c.Path = parent.Node.Path + "/" + segment
			} else {
				c.Path = segment
			}
		}
		normalized = append(normalized, c)
	}
	return normalized
}

func (inst *Instance) reapplyLazyInsertions() bool {
	if len(inst.lazyInsertions) == 0 {
		return false
	}
	changed := false
	for key, children := range inst.lazyInsertions {
		parentIndex := inst.findNodeIndexByKey(key)
		if parentIndex < 0 || parentIndex >= len(inst.nodes) {
			continue
		}
		if inst.hasDescendantsAt(parentIndex) {
			continue
		}
		parent := nodeEntry{
			Index: parentIndex,
			Node:  inst.nodes[parentIndex],
			Depth: nodeDepth(inst.nodes[parentIndex]),
			Key:   key,
		}
		if inst.insertLazyChildren(parent, children, false) {
			changed = true
		}
	}
	return changed
}

func (inst *Instance) removeDescendantsAt(index int) bool {
	if index < 0 || index >= len(inst.nodes)-1 {
		return false
	}
	depth := nodeDepth(inst.nodes[index])
	end := index + 1
	for end < len(inst.nodes) && nodeDepth(inst.nodes[end]) > depth {
		end++
	}
	if end == index+1 {
		return false
	}
	inst.nodes = append(inst.nodes[:index+1], inst.nodes[end:]...)
	return true
}

func (inst *Instance) hasDescendantsAt(index int) bool {
	if index < 0 || index >= len(inst.nodes)-1 {
		return false
	}
	return nodeDepth(inst.nodes[index+1]) > nodeDepth(inst.nodes[index])
}

func (inst *Instance) findNodeIndexByKey(key string) int {
	for i, node := range inst.nodes {
		if nodeKey(node, i) == key {
			return i
		}
	}
	return -1
}

func (inst *Instance) resolveNodeIndex(nodeIndex int, path string, nodeID int) int {
	if nodeIndex >= 0 && nodeIndex < len(inst.nodes) {
		return nodeIndex
	}
	if path != "" {
		if idx := inst.findNodeIndexByPath(path); idx >= 0 {
			return idx
		}
	}
	if nodeID != 0 {
		if idx := inst.findNodeIndexByID(nodeID); idx >= 0 {
			return idx
		}
	}
	return -1
}

func (inst *Instance) applyLazyLoadResult(nodeIndex int, children []TreeNode, replace bool) bool {
	if nodeIndex < 0 || nodeIndex >= len(inst.nodes) {
		return false
	}
	parent := nodeEntry{
		Index: nodeIndex,
		Node:  inst.nodes[nodeIndex],
		Depth: nodeDepth(inst.nodes[nodeIndex]),
		Key:   nodeKey(inst.nodes[nodeIndex], nodeIndex),
	}
	if !inst.insertLazyChildren(parent, children, replace) {
		return false
	}
	inst.normalizeSelectionAndScroll()
	inst.dirty = true
	return true
}

func (inst *Instance) applyLazyLoadFailure(nodeIndex int, err string) bool {
	if nodeIndex < 0 || nodeIndex >= len(inst.nodes) {
		return false
	}
	if !inst.nodes[nodeIndex].Loading && inst.nodes[nodeIndex].LoadError == err {
		return false
	}
	inst.nodes[nodeIndex].Loading = false
	inst.nodes[nodeIndex].LoadError = err
	inst.invalidateCache()
	inst.normalizeSelectionAndScroll()
	inst.dirty = true
	return true
}

func (inst *Instance) nodeMatches(node TreeNode, query string) bool {
	if query == "" {
		return false
	}
	if inst.searchFn != nil {
		return inst.searchFn(node, query)
	}
	lower := strings.ToLower(query)
	content := strings.ToLower(node.Content)
	path := strings.ToLower(node.Path)
	if strings.Contains(content, lower) || (path != "" && strings.Contains(path, lower)) {
		return true
	}
	if node.LoadError != "" && strings.Contains(strings.ToLower(node.LoadError), lower) {
		return true
	}
	return false
}

func (inst *Instance) parentVisibleIndex(visible []nodeEntry, index int) int {
	if index <= 0 || index >= len(visible) {
		return -1
	}
	currentDepth := visible[index].Depth
	if currentDepth == 0 {
		return -1
	}
	for i := index - 1; i >= 0; i-- {
		if visible[i].Depth < currentDepth {
			return i
		}
	}
	return -1
}

func (inst *Instance) firstChildVisibleIndex(visible []nodeEntry, index int) int {
	if index < 0 || index >= len(visible)-1 {
		return -1
	}
	currentDepth := visible[index].Depth
	next := visible[index+1]
	if next.Depth > currentDepth {
		return index + 1
	}
	return -1
}

func (inst *Instance) findNodeIndexByPath(path string) int {
	if path == "" {
		return -1
	}
	for i, node := range inst.nodes {
		if node.Path == path {
			return i
		}
	}
	return -1
}

func (inst *Instance) findNodeIndexByID(id int) int {
	if id == 0 {
		return -1
	}
	for i, node := range inst.nodes {
		if node.NodeID == id {
			return i
		}
	}
	return -1
}

// isExpanded checks if a node is expanded.
func (inst *Instance) isExpanded(nodeIndex int) bool {
	if nodeIndex < 0 || nodeIndex >= len(inst.nodes) {
		return false
	}
	key := nodeKey(inst.nodes[nodeIndex], nodeIndex)
	return inst.expandState[key]
}

func (inst *Instance) calculateContentWidth(visible []nodeEntry) int {
	maxWidth := 1
	for _, entry := range visible {
		prefix, icon, content := inst.lineParts(entry)
		lineWidth := paint.StringWidth(prefix + icon + content)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	return maxWidth
}

func (inst *Instance) chromeHeight() int {
	if inst.showBorder {
		return 2 + inst.statsHeight()
	}
	return inst.statsHeight()
}

func (inst *Instance) statsHeight() int {
	if inst.showSearchStats {
		return 1
	}
	return 0
}

func (inst *Instance) searchStatsLine() string {
	query := strings.TrimSpace(inst.searchQuery)
	total, selected := inst.GetMatchStats()
	if query == "" {
		return "Search: --"
	}
	return fmt.Sprintf("Search: %q %d/%d", query, selected, total)
}

func (inst *Instance) desiredViewportHeight(visibleCount int) int {
	viewSize := inst.viewportHeight
	if viewSize < 1 {
		viewSize = visibleCount
	}
	if visibleCount > 0 && viewSize > visibleCount {
		viewSize = visibleCount
	}
	if viewSize < 1 {
		viewSize = 1
	}
	return viewSize
}

func (inst *Instance) effectiveViewportHeight(visibleCount int) int {
	viewSize := inst.desiredViewportHeight(visibleCount)
	if inst.bounds[3] > 0 {
		available := inst.bounds[3] - inst.chromeHeight()
		if available < 1 {
			available = 1
		}
		if viewSize > available {
			viewSize = available
		}
	}
	if visibleCount > 0 && viewSize > visibleCount {
		viewSize = visibleCount
	}
	if viewSize < 1 {
		viewSize = 1
	}
	return viewSize
}

func (inst *Instance) visibleViewport(visibleCount int) scrollutil.VerticalViewport {
	viewSize := inst.effectiveViewportHeight(visibleCount)
	return scrollutil.NewVerticalViewport(visibleCount, viewSize, inst.scrollOffset)
}

func (inst *Instance) normalizeSelectionAndScroll() {
	visible, _ := inst.visibleEntries()
	visibleCount := len(visible)
	if visibleCount == 0 {
		inst.selectedIndex = -1
		inst.scrollOffset = 0
		return
	}

	query := strings.TrimSpace(inst.searchQuery)
	if query == "" {
		inst.autoSelectMatch = false
	}
	if query != "" && !inst.selectedIndexControlled {
		firstMatch := -1
		for i, entry := range visible {
			if entry.Match {
				firstMatch = i
				break
			}
		}
		if firstMatch >= 0 {
			if inst.autoSelectMatch || inst.selectedIndex < 0 || inst.selectedIndex >= visibleCount || !visible[inst.selectedIndex].Match {
				inst.selectedIndex = firstMatch
			}
		} else if inst.autoSelectMatch {
			inst.selectedIndex = -1
		}
		inst.autoSelectMatch = false
	}
	if inst.selectedIndex >= visibleCount {
		inst.selectedIndex = visibleCount - 1
	}
	if inst.selectedIndex < -1 {
		inst.selectedIndex = -1
	}
	viewport := inst.visibleViewport(visibleCount)
	if inst.selectedIndex >= 0 {
		if viewport.EnsureVisible(inst.selectedIndex) {
			inst.scrollOffset = viewport.Offset
		}
	} else {
		inst.scrollOffset = viewport.Offset
	}

	inst.updateSearchStats()
}

func (inst *Instance) normalizeCheckedKeys() {
	if inst.selectionMode == SelectionNone || len(inst.nodes) == 0 {
		inst.checkedKeys = nil
		inst.selectionAnchorKey = ""
		return
	}
	normalized := make(map[string]bool)
	for i, node := range inst.nodes {
		key := nodeKey(node, i)
		if inst.checkedKeys != nil && inst.checkedKeys[key] {
			normalized[key] = true
		}
	}
	if inst.selectionMode == SelectionSingle && len(normalized) > 1 {
		trimmed := make(map[string]bool, 1)
		for i, node := range inst.nodes {
			key := nodeKey(node, i)
			if normalized[key] {
				trimmed[key] = true
				break
			}
		}
		normalized = trimmed
	}
	inst.checkedKeys = normalized
	if inst.selectionAnchorKey != "" && inst.findNodeIndexByKey(inst.selectionAnchorKey) < 0 {
		inst.selectionAnchorKey = ""
	}
}

func (inst *Instance) emitCheckedSelectionChanged() {
	if inst.intentEmitter == nil {
		return
	}

	keys, indices, paths, nodeIDs := inst.checkedSnapshot()
	if inst.componentID != "" {
		inst.emitOptionalGlobalIntent(SelectionChange(inst.componentID, inst.selectionMode, keys, indices, paths, nodeIDs))
	}

	if inst.selectionIntentField != nil {
		value := strings.Join(paths, ",")
		if value == "" {
			value = strings.Join(keys, ",")
		}
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.selectionIntentField.GetField(),
			Value: value,
		})
		return
	}
	if inst.selectionIntent != nil {
		inst.intentEmitter(inst.selectionIntent)
	}
}

func (inst *Instance) checkedSnapshot() (keys []string, indices []int, paths []string, nodeIDs []int) {
	if inst.checkedKeys == nil {
		return nil, nil, nil, nil
	}
	keys = make([]string, 0, len(inst.checkedKeys))
	indices = make([]int, 0, len(inst.checkedKeys))
	paths = make([]string, 0, len(inst.checkedKeys))
	nodeIDs = make([]int, 0, len(inst.checkedKeys))
	for i, node := range inst.nodes {
		key := nodeKey(node, i)
		if inst.checkedKeys[key] {
			keys = append(keys, key)
			indices = append(indices, i)
			if node.Path != "" {
				paths = append(paths, node.Path)
			}
			if node.NodeID != 0 {
				nodeIDs = append(nodeIDs, node.NodeID)
			}
		}
	}
	return keys, indices, paths, nodeIDs
}

func (inst *Instance) selectVisibleIndex(target int, visible []nodeEntry, emitSelect bool) bool {
	if len(visible) == 0 {
		inst.selectedIndex = -1
		inst.scrollOffset = 0
		return false
	}
	if target < 0 {
		target = 0
	}
	if target >= len(visible) {
		target = len(visible) - 1
	}

	viewport := inst.visibleViewport(len(visible))
	oldScroll := inst.scrollOffset
	changed := inst.selectedIndex != target
	inst.selectedIndex = target
	if viewport.EnsureVisible(target) {
		if inst.scrollOffset != viewport.Offset {
			changed = true
		}
		inst.scrollOffset = viewport.Offset
	} else if inst.scrollOffset != viewport.Offset {
		changed = true
		inst.scrollOffset = viewport.Offset
	}

	if changed {
		inst.dirty = true
		if emitSelect {
			inst.emitNodeSelect(visible[target].Index)
		}
		if inst.scrollOffset != oldScroll {
			inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
		}
	}
	return changed
}

func (inst *Instance) scrollBy(delta int) bool {
	visible, _ := inst.visibleEntries()
	viewport := inst.visibleViewport(len(visible))
	oldScroll := inst.scrollOffset
	if !viewport.ScrollBy(delta) {
		return false
	}
	inst.scrollOffset = viewport.Offset
	inst.dirty = true
	inst.emitScroll(inst.scrollOffset-oldScroll, viewport.ViewSize, len(visible))
	return true
}

func (inst *Instance) toggleChecked(entry nodeEntry) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	changed := false
	switch inst.selectionMode {
	case SelectionSingle:
		changed = !inst.isChecked(entry) || len(inst.checkedKeys) != 1
		inst.checkedKeys = map[string]bool{entry.Key: true}
	case SelectionMultiple:
		if inst.checkedKeys == nil {
			inst.checkedKeys = make(map[string]bool)
		}
		targetChecked := !inst.checkedKeys[entry.Key]
		changed = inst.setCheckedSubtree(entry.Index, targetChecked)
	}

	if changed {
		inst.selectionAnchorKey = entry.Key
		inst.normalizeCheckedKeys()
		inst.dirty = true
		inst.emitCheckedSelectionChanged()
	}
	return changed
}

func (inst *Instance) setCheckedSubtree(index int, checked bool) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	if index < 0 || index >= len(inst.nodes) {
		return false
	}
	if inst.checkedKeys == nil {
		inst.checkedKeys = make(map[string]bool)
	}
	changed := false
	depth := nodeDepth(inst.nodes[index])
	for i := index; i < len(inst.nodes); i++ {
		if i > index && nodeDepth(inst.nodes[i]) <= depth {
			break
		}
		key := nodeKey(inst.nodes[i], i)
		if checked {
			if !inst.checkedKeys[key] {
				inst.checkedKeys[key] = true
				changed = true
			}
			continue
		}
		if inst.checkedKeys[key] {
			delete(inst.checkedKeys, key)
			changed = true
		}
	}
	return changed
}

func (inst *Instance) toggleCheckedSelection(act *action.Action) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	if _, matchesOnly, ok := inst.selectionInvertFromAction(act); ok {
		return inst.invertCheckedSelection(matchesOnly)
	}
	visible, visibleIndex := inst.visibleEntries()
	entry, ok := inst.entryFromAction(act, visible, visibleIndex)
	if !ok {
		if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
			if len(visible) == 0 {
				return false
			}
			entry = visible[0]
		} else {
			entry = visible[inst.selectedIndex]
		}
	}
	return inst.toggleChecked(entry)
}

func (inst *Instance) deselectByAction(act *action.Action) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	visible, visibleIndex := inst.visibleEntries()
	entry, ok := inst.entryFromAction(act, visible, visibleIndex)
	if !ok {
		if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
			return false
		}
		entry = visible[inst.selectedIndex]
	}
	if inst.checkedKeys == nil || !inst.checkedKeys[entry.Key] {
		return false
	}
	delete(inst.checkedKeys, entry.Key)
	inst.selectionAnchorKey = entry.Key
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) selectRangeByAction(act *action.Action) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	start, end, ok := inst.rangeFromAction(act, len(visible))
	if !ok {
		anchor, anchorOK := inst.selectionAnchorVisibleIndex(visible)
		if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
			return false
		}
		if !anchorOK {
			anchor = inst.selectedIndex
			inst.selectionAnchorKey = visible[anchor].Key
		}
		start, end, ok = anchor, inst.selectedIndex, true
	}
	if start > end {
		start, end = end, start
	}
	if inst.checkedKeys == nil {
		inst.checkedKeys = make(map[string]bool)
	}
	for i := start; i <= end; i++ {
		inst.checkedKeys[visible[i].Key] = true
	}
	inst.selectionAnchorKey = visible[start].Key
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) selectAllByAction(act *action.Action) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	filterActive := inst.selectionScopeMatchesOnly(act, strings.TrimSpace(inst.searchQuery) != "")
	checked := make(map[string]bool, len(visible))
	for _, entry := range visible {
		if filterActive && !entry.Match {
			continue
		}
		checked[entry.Key] = true
	}
	inst.checkedKeys = checked
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(visible) {
		inst.selectionAnchorKey = visible[inst.selectedIndex].Key
	}
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) clearCheckedSelectionByAction(act *action.Action) bool {
	if len(inst.checkedKeys) == 0 {
		return false
	}
	if act != nil {
		if scoped, matchesOnly, ok := inst.selectionClearScopeFromAction(act); ok {
			if scoped {
				return inst.clearCheckedScope(matchesOnly)
			}
		}
	}
	inst.checkedKeys = nil
	inst.selectionAnchorKey = ""
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) clearCheckedScope(matchesOnly bool) bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	changed := false
	for _, entry := range visible {
		if matchesOnly && !entry.Match {
			continue
		}
		if inst.checkedKeys[entry.Key] {
			delete(inst.checkedKeys, entry.Key)
			changed = true
		}
	}
	if !changed {
		return false
	}
	inst.normalizeCheckedKeys()
	if len(inst.checkedKeys) == 0 {
		inst.selectionAnchorKey = ""
	}
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) invertCheckedSelection(matchesOnly bool) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	if inst.checkedKeys == nil {
		inst.checkedKeys = make(map[string]bool)
	}
	changed := false
	for _, entry := range visible {
		if matchesOnly && !entry.Match {
			continue
		}
		if inst.checkedKeys[entry.Key] {
			delete(inst.checkedKeys, entry.Key)
		} else {
			inst.checkedKeys[entry.Key] = true
		}
		changed = true
	}
	if !changed {
		return false
	}
	if inst.selectedIndex >= 0 && inst.selectedIndex < len(visible) {
		inst.selectionAnchorKey = visible[inst.selectedIndex].Key
	}
	inst.normalizeCheckedKeys()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) selectionAnchorVisibleIndex(visible []nodeEntry) (int, bool) {
	if inst.selectionAnchorKey == "" {
		return -1, false
	}
	for i, entry := range visible {
		if entry.Key == inst.selectionAnchorKey {
			return i, true
		}
	}
	return -1, false
}

func (inst *Instance) selectionInvertFromAction(act *action.Action) (invert bool, matchesOnly bool, ok bool) {
	if act == nil {
		return false, false, false
	}
	def := strings.TrimSpace(inst.searchQuery) != ""
	switch payload := act.Payload.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(payload)) {
		case "invert":
			return true, def, true
		case "invert_visible":
			return true, false, true
		case "invert_matches":
			return true, true, true
		}
	case map[string]string:
		mode := strings.ToLower(strings.TrimSpace(payload["mode"]))
		if mode == "" {
			mode = strings.ToLower(strings.TrimSpace(payload["op"]))
		}
		if mode == "invert" {
			return true, inst.selectionScopeMatchesOnly(act, def), true
		}
	case map[string]interface{}:
		mode := ""
		if raw, exists := payload["mode"]; exists {
			if value, ok := raw.(string); ok {
				mode = strings.ToLower(strings.TrimSpace(value))
			}
		}
		if mode == "" {
			if raw, exists := payload["op"]; exists {
				if value, ok := raw.(string); ok {
					mode = strings.ToLower(strings.TrimSpace(value))
				}
			}
		}
		if mode == "invert" {
			return true, inst.selectionScopeMatchesOnly(act, def), true
		}
	}
	return false, false, false
}

func (inst *Instance) selectionScopeMatchesOnly(act *action.Action, def bool) bool {
	if act == nil {
		return def
	}
	switch payload := act.Payload.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(payload)) {
		case "visible", "all", "all_visible", "invert_visible", "clear_visible":
			return false
		case "matches", "matched", "search", "search_matches", "invert_matches", "clear_matches":
			return true
		}
	case map[string]string:
		if scope := strings.ToLower(strings.TrimSpace(payload["scope"])); scope != "" {
			return scope == "matches" || scope == "matched" || scope == "search"
		}
	case map[string]interface{}:
		if raw, exists := payload["scope"]; exists {
			if scope, ok := raw.(string); ok {
				scope = strings.ToLower(strings.TrimSpace(scope))
				return scope == "matches" || scope == "matched" || scope == "search"
			}
		}
	}
	return def
}

func (inst *Instance) selectionClearScopeFromAction(act *action.Action) (scoped bool, matchesOnly bool, ok bool) {
	if act == nil {
		return false, false, false
	}
	switch payload := act.Payload.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(payload)) {
		case "visible", "clear_visible":
			return true, false, true
		case "matches", "matched", "clear_matches":
			return true, true, true
		}
	case map[string]string:
		mode := strings.ToLower(strings.TrimSpace(payload["mode"]))
		if mode == "" {
			mode = strings.ToLower(strings.TrimSpace(payload["op"]))
		}
		if mode == "clear" {
			return true, inst.selectionScopeMatchesOnly(act, false), true
		}
	case map[string]interface{}:
		mode := ""
		if raw, exists := payload["mode"]; exists {
			if value, ok := raw.(string); ok {
				mode = strings.ToLower(strings.TrimSpace(value))
			}
		}
		if mode == "" {
			if raw, exists := payload["op"]; exists {
				if value, ok := raw.(string); ok {
					mode = strings.ToLower(strings.TrimSpace(value))
				}
			}
		}
		if mode == "clear" {
			return true, inst.selectionScopeMatchesOnly(act, false), true
		}
	}
	return false, false, false
}

func (inst *Instance) handleActivate() bool {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
		return inst.selectVisibleIndex(0, visible, true)
	}

	entry := visible[inst.selectedIndex]
	if inst.selectionMode != SelectionNone {
		return inst.toggleChecked(entry)
	}
	if inst.allowExpand && (entry.HasChildren || entry.Node.NodeType == "folder") {
		return inst.toggleExpand()
	}

	inst.emitNodeSelect(entry.Index)
	return true
}

func (inst *Instance) handleClick(act *action.Action) bool {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil {
		return inst.handleActivate()
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	viewSize := inst.effectiveViewportHeight(len(visible))
	rowIndex, ok := inst.rowIndexAtLocalY(mouseMsg.LocalY, viewSize)
	if !ok {
		return false
	}
	target := inst.scrollOffset + rowIndex
	if target < 0 || target >= len(visible) {
		return false
	}
	// Select the row first.
	inst.selectVisibleIndex(target, visible, true)

	entry := visible[target]
	if inst.allowExpand && entry.HasChildren && inst.clickOnExpander(entry, mouseMsg.LocalX) {
		inst.toggleExpand()
		return true
	}
	if inst.selectionMode != SelectionNone {
		return inst.toggleChecked(entry)
	}
	return true
}

func (inst *Instance) handleDoubleClick(act *action.Action) bool {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil {
		return inst.handleActivate()
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	viewSize := inst.effectiveViewportHeight(len(visible))
	rowIndex, ok := inst.rowIndexAtLocalY(mouseMsg.LocalY, viewSize)
	if !ok {
		return false
	}
	target := inst.scrollOffset + rowIndex
	if target < 0 || target >= len(visible) {
		return false
	}
	inst.selectVisibleIndex(target, visible, true)
	entry := visible[target]
	if inst.allowExpand && entry.HasChildren {
		inst.toggleExpand()
		return true
	}
	if inst.selectionMode != SelectionNone {
		return inst.toggleChecked(entry)
	}
	return true
}

func (inst *Instance) rowIndexAtLocalY(localY, viewSize int) (int, bool) {
	offset := 0
	if inst.showBorder {
		offset = 1
	}
	offset += inst.statsHeight()
	row := localY - offset
	if row < 0 || row >= viewSize {
		return -1, false
	}
	return row, true
}

func (inst *Instance) handleInputShortcut(act *action.Action) bool {
	if act == nil {
		return false
	}
	input, ok := act.GetPayloadString()
	if !ok {
		return false
	}
	runes := []rune(input)
	if len(runes) != 1 {
		return false
	}
	switch unicode.ToLower(runes[0]) {
	case 'r':
		return inst.refreshSelectedLazy()
	}
	return false
}

func (inst *Instance) clickOnExpander(entry nodeEntry, localX int) bool {
	if !inst.showIcons {
		return false
	}
	adjusted := localX
	if inst.showBorder {
		adjusted -= 2
	}
	if adjusted < 0 {
		return false
	}
	prefixWidth := paint.StringWidth(inst.indentPrefix(entry.Depth)) + inst.selectionMarkerWidth()
	iconWidth := paint.StringWidth(inst.iconFor(entry))
	if iconWidth <= 0 {
		return false
	}
	iconStart := prefixWidth
	iconEnd := iconStart + iconWidth
	return adjusted >= iconStart && adjusted < iconEnd
}

func (inst *Instance) searchDirectionFromAction(act *action.Action) (int, bool) {
	if act == nil {
		return 0, false
	}
	if dir, ok := act.GetPayloadInt(); ok {
		if dir < 0 {
			return -1, true
		}
		if dir > 0 {
			return 1, true
		}
		return 0, false
	}
	switch payload := act.Payload.(type) {
	case string:
		switch strings.ToLower(payload) {
		case "next", "forward", "down":
			return 1, true
		case "prev", "previous", "back", "up":
			return -1, true
		}
	case map[string]string:
		if dir, ok := payload["dir"]; ok {
			return inst.searchDirectionFromAction(action.NewAction(action.ActionSearch).WithPayload(dir))
		}
	case map[string]interface{}:
		if raw, ok := payload["dir"]; ok {
			if dir, ok := raw.(string); ok {
				return inst.searchDirectionFromAction(action.NewAction(action.ActionSearch).WithPayload(dir))
			}
		}
	}
	return 0, false
}

func (inst *Instance) navigateMatch(direction int) bool {
	if direction == 0 {
		return false
	}
	if strings.TrimSpace(inst.searchQuery) == "" {
		return false
	}
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	start := inst.selectedIndex
	if start < 0 || start >= len(visible) {
		start = 0
	}
	for step := 1; step <= len(visible); step++ {
		idx := start + step*direction
		for idx < 0 {
			idx += len(visible)
		}
		idx = idx % len(visible)
		if visible[idx].Match {
			fromIndex := inst.selectedIndex
			if inst.selectVisibleIndex(idx, visible, true) {
				dirLabel := "search_next"
				if direction < 0 {
					dirLabel = "search_prev"
				}
				inst.emitNavigation(dirLabel, fromIndex, inst.selectedIndex)
				return true
			}
			return true
		}
	}
	return false
}

func (inst *Instance) refreshSelectedLazy() bool {
	return inst.refreshLazyByAction(nil)
}

func (inst *Instance) refreshLazyByAction(act *action.Action) bool {
	visible, visibleIndex := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	var entry nodeEntry
	if act != nil {
		if visibleEntry, ok := inst.entryFromAction(act, visible, visibleIndex); ok {
			entry = visibleEntry
		}
	}
	if entry.Key == "" {
		if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
			return false
		}
		entry = visible[inst.selectedIndex]
	}
	if !entry.Node.Lazy && entry.Node.LoadError == "" {
		return false
	}
	return inst.requestLazyLoad(entry, true)
}

func (inst *Instance) entryFromAction(act *action.Action, visible []nodeEntry, visibleIndex []int) (nodeEntry, bool) {
	index, ok := inst.visibleIndexFromAction(act, visible, visibleIndex)
	if !ok || index < 0 || index >= len(visible) {
		return nodeEntry{}, false
	}
	return visible[index], true
}

func (inst *Instance) visibleIndexFromAction(act *action.Action, visible []nodeEntry, visibleIndex []int) (int, bool) {
	if idx, ok := act.GetPayloadInt(); ok {
		if idx >= 0 && idx < len(visible) {
			return idx, true
		}
		return -1, false
	}

	switch payload := act.Payload.(type) {
	case string:
		return inst.visibleIndexForKeyOrPath(payload, visibleIndex)
	case map[string]int:
		if idx, ok := payload["index"]; ok {
			if idx >= 0 && idx < len(visible) {
				return idx, true
			}
			return -1, false
		}
		if nodeID, ok := payload["nodeID"]; ok {
			return inst.visibleIndexForNodeID(nodeID, visibleIndex)
		}
	case map[string]string:
		if key, ok := payload["key"]; ok {
			return inst.visibleIndexForKeyOrPath(key, visibleIndex)
		}
		if path, ok := payload["path"]; ok {
			return inst.visibleIndexForKeyOrPath(path, visibleIndex)
		}
	case map[string]interface{}:
		if raw, ok := payload["index"]; ok {
			if idx, ok := raw.(int); ok {
				if idx >= 0 && idx < len(visible) {
					return idx, true
				}
				return -1, false
			}
		}
		if raw, ok := payload["nodeID"]; ok {
			if nodeID, ok := raw.(int); ok {
				return inst.visibleIndexForNodeID(nodeID, visibleIndex)
			}
		}
		if raw, ok := payload["path"]; ok {
			if path, ok := raw.(string); ok {
				return inst.visibleIndexForKeyOrPath(path, visibleIndex)
			}
		}
		if raw, ok := payload["key"]; ok {
			if key, ok := raw.(string); ok {
				return inst.visibleIndexForKeyOrPath(key, visibleIndex)
			}
		}
	}
	return -1, false
}

func (inst *Instance) visibleIndexForNodeID(nodeID int, visibleIndex []int) (int, bool) {
	if nodeID == 0 {
		return -1, false
	}
	nodeIndex := inst.findNodeIndexByID(nodeID)
	if nodeIndex < 0 || nodeIndex >= len(visibleIndex) {
		return -1, false
	}
	visible := visibleIndex[nodeIndex]
	if visible < 0 {
		return -1, false
	}
	return visible, true
}

func (inst *Instance) visibleIndexForKeyOrPath(value string, visibleIndex []int) (int, bool) {
	nodeIndex := inst.resolveNodeIndexFromKeyOrPath(value)
	if nodeIndex < 0 || nodeIndex >= len(visibleIndex) {
		return -1, false
	}
	visible := visibleIndex[nodeIndex]
	if visible < 0 {
		return -1, false
	}
	return visible, true
}

func (inst *Instance) resolveNodeIndexFromKeyOrPath(value string) int {
	if value == "" {
		return -1
	}
	if strings.HasPrefix(value, "path:") {
		return inst.findNodeIndexByPath(strings.TrimPrefix(value, "path:"))
	}
	if strings.HasPrefix(value, "id:") {
		if id, err := strconv.Atoi(strings.TrimPrefix(value, "id:")); err == nil {
			return inst.findNodeIndexByID(id)
		}
		return -1
	}
	if strings.HasPrefix(value, "idx:") {
		if idx, err := strconv.Atoi(strings.TrimPrefix(value, "idx:")); err == nil {
			if idx >= 0 && idx < len(inst.nodes) {
				return idx
			}
		}
		return -1
	}

	if idx := inst.findNodeIndexByPath(value); idx >= 0 {
		return idx
	}
	for i, node := range inst.nodes {
		if nodeKey(node, i) == value {
			return i
		}
	}
	return -1
}

func (inst *Instance) rangeFromAction(act *action.Action, max int) (int, int, bool) {
	if max <= 0 {
		return 0, 0, false
	}
	switch payload := act.Payload.(type) {
	case []int:
		if len(payload) >= 2 {
			return clampIndex(payload[0], max), clampIndex(payload[1], max), true
		}
	case [2]int:
		return clampIndex(payload[0], max), clampIndex(payload[1], max), true
	case map[string]int:
		start, okStart := payload["start"]
		end, okEnd := payload["end"]
		if okStart && okEnd {
			return clampIndex(start, max), clampIndex(end, max), true
		}
	case map[string]interface{}:
		rawStart, okStart := payload["start"]
		rawEnd, okEnd := payload["end"]
		if okStart && okEnd {
			if start, ok := rawStart.(int); ok {
				if end, ok := rawEnd.(int); ok {
					return clampIndex(start, max), clampIndex(end, max), true
				}
			}
		}
	}
	return 0, 0, false
}

func clampIndex(value, max int) int {
	if value < 0 {
		return 0
	}
	if value >= max {
		return max - 1
	}
	return value
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
func (inst *Instance) GetMatchStats() (total int, selected int) {
	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return 0, 0
	}
	count := 0
	selectedMatch := 0
	for i, entry := range visible {
		if entry.Match {
			count++
			if i == inst.selectedIndex {
				selectedMatch = count
			}
		}
	}
	return count, selectedMatch
}

func (inst *Instance) updateSearchStats() {
	total, selected := inst.GetMatchStats()
	if inst.lastSearchQuery == inst.searchQuery &&
		inst.lastSearchTotal == total &&
		inst.lastSearchSelected == selected {
		return
	}
	inst.lastSearchQuery = inst.searchQuery
	inst.lastSearchTotal = total
	inst.lastSearchSelected = selected
	if inst.intentEmitter != nil {
		inst.emitSearchStats(total, selected)
	}
}
func (inst *Instance) GetCheckedKeys() []string {
	keys, _, _, _ := inst.checkedSnapshot()
	return append([]string(nil), keys...)
}
func (inst *Instance) GetCheckedIndices() []int {
	_, indices, _, _ := inst.checkedSnapshot()
	return append([]int(nil), indices...)
}
func (inst *Instance) GetCheckedNodes() []TreeNode {
	checked := []TreeNode{}
	if inst.checkedKeys == nil {
		return checked
	}
	for i, node := range inst.nodes {
		key := nodeKey(node, i)
		if inst.checkedKeys[key] {
			checked = append(checked, node)
		}
	}
	return checked
}
func (inst *Instance) SelectIndex(index int) bool {
	visible, _ := inst.visibleEntries()
	if index < 0 || index >= len(visible) {
		return false
	}
	return inst.selectVisibleIndex(index, visible, true)
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
			visible, visibleIndex := inst.visibleEntries()
			if v.NodeIndex < len(visibleIndex) && visibleIndex[v.NodeIndex] >= 0 {
				return inst.selectVisibleIndex(visibleIndex[v.NodeIndex], visible, false)
			}
		}
		if v.Path != "" {
			if idx := inst.findNodeIndexByPath(v.Path); idx >= 0 {
				visible, visibleIndex := inst.visibleEntries()
				if idx < len(visibleIndex) && visibleIndex[idx] >= 0 {
					return inst.selectVisibleIndex(visibleIndex[idx], visible, false)
				}
			}
		}
		if v.NodeID != 0 {
			if idx := inst.findNodeIndexByID(v.NodeID); idx >= 0 {
				visible, visibleIndex := inst.visibleEntries()
				if idx < len(visibleIndex) && visibleIndex[idx] >= 0 {
					return inst.selectVisibleIndex(visibleIndex[idx], visible, false)
				}
			}
		}

	case NodeExpandIntent:
		// Handle expand by external request
		nodeIndex := inst.resolveNodeIndex(v.NodeIndex, v.Path, v.NodeID)
		if nodeIndex >= 0 && nodeIndex < len(inst.nodes) {
			key := nodeKey(inst.nodes[nodeIndex], nodeIndex)
			inst.expandState[key] = true
			inst.invalidateCache()
			if inst.expandedKeysControlled {
				inst.setExpandedKey(key, true)
			}
			entries := inst.buildNodeEntries()
			if nodeIndex < len(entries) {
				inst.maybeEmitLazyLoad(entries[nodeIndex])
			}
			inst.normalizeSelectionAndScroll()
			inst.dirty = true
			return true
		}

	case NodeCollapseIntent:
		// Handle collapse by external request
		nodeIndex := inst.resolveNodeIndex(v.NodeIndex, v.Path, v.NodeID)
		if nodeIndex >= 0 && nodeIndex < len(inst.nodes) {
			key := nodeKey(inst.nodes[nodeIndex], nodeIndex)
			inst.expandState[key] = false
			inst.invalidateCache()
			if inst.expandedKeysControlled {
				inst.setExpandedKey(key, false)
			}
			inst.normalizeSelectionAndScroll()
			inst.dirty = true
			return true
		}
	case ExpandAllIntent:
		inst.setAllExpanded(true)
		return true
	case CollapseAllIntent:
		inst.setAllExpanded(false)
		return true
	case SearchNextIntent:
		return inst.navigateMatch(1)
	case SearchPrevIntent:
		return inst.navigateMatch(-1)
	case LazyLoadSuccessIntent:
		nodeIndex := inst.resolveNodeIndex(v.NodeIndex, v.Path, v.NodeID)
		return inst.applyLazyLoadResult(nodeIndex, v.Children, v.Replace)
	case LazyLoadFailureIntent:
		nodeIndex := inst.resolveNodeIndex(v.NodeIndex, v.Path, v.NodeID)
		return inst.applyLazyLoadFailure(nodeIndex, v.Error)
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
	inst.emitOptionalGlobalIntent(nodeSelect)
}

// emitNodeExpand emits a NodeExpandIntent when a folder is expanded.
func (inst *Instance) emitNodeExpand(nodeIndex int, path string, nodeID int) {
	var expandIntent NodeExpandIntent
	if inst.componentID != "" {
		expandIntent = NodeExpandWithID(inst.componentID, nodeIndex, path, nodeID)
	} else {
		expandIntent = NodeExpand(nodeIndex, path, nodeID)
	}
	inst.emitOptionalGlobalIntent(expandIntent)
}

// emitNodeCollapse emits a NodeCollapseIntent when a folder is collapsed.
func (inst *Instance) emitNodeCollapse(nodeIndex int, path string, nodeID int) {
	var collapseIntent NodeCollapseIntent
	if inst.componentID != "" {
		collapseIntent = NodeCollapseWithID(inst.componentID, nodeIndex, path, nodeID)
	} else {
		collapseIntent = NodeCollapse(nodeIndex, path, nodeID)
	}
	inst.emitOptionalGlobalIntent(collapseIntent)
}

// emitNavigation emits a NavigationIntent when the selection changes via navigation.
func (inst *Instance) emitNavigation(direction string, fromIndex, toIndex int) {
	var navIntent NavigationIntent
	if inst.componentID != "" {
		navIntent = NavigationWithID(inst.componentID, direction, fromIndex, toIndex)
	} else {
		navIntent = Navigation(direction, fromIndex, toIndex)
	}
	inst.emitOptionalGlobalIntent(navIntent)
}

func (inst *Instance) emitScroll(delta, viewSize, contentSize int) {
	var scrollIntent ScrollIntent
	if inst.componentID != "" {
		scrollIntent = ScrollWithID(inst.componentID, inst.scrollOffset, delta, viewSize, contentSize)
	} else {
		scrollIntent = Scroll(inst.scrollOffset, delta, viewSize, contentSize)
	}
	inst.emitOptionalGlobalIntent(scrollIntent)
}

func (inst *Instance) emitSearchStats(total, selected int) {
	var statsIntent SearchStatsIntent
	if inst.componentID != "" {
		statsIntent = SearchStatsWithID(inst.componentID, inst.searchQuery, total, selected)
	} else {
		statsIntent = SearchStats(inst.searchQuery, total, selected)
	}
	inst.emitOptionalGlobalIntent(statsIntent)
}

func (inst *Instance) emitLazyLoad(nodeIndex int, path string, nodeID int) {
	var lazyIntent LazyLoadIntent
	if inst.componentID != "" {
		lazyIntent = LazyLoadWithID(inst.componentID, nodeIndex, path, nodeID)
	} else {
		lazyIntent = LazyLoad(nodeIndex, path, nodeID)
	}
	inst.emitOptionalGlobalIntent(lazyIntent)
}

func (inst *Instance) emitOptionalGlobalIntent(i intent.Intent) {
	if i == nil {
		return
	}
	runtime := rtui.GetGlobalIntentRuntime()
	if runtime == nil || runtime.Registry == nil {
		return
	}
	intentType := i.IntentType()
	if !runtime.Registry.HasHandler(intentType) && !runtime.Registry.HasFallback() {
		return
	}
	rtui.EmitIntentGlobal(i)
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

func (inst *Instance) setExpandedKey(key string, value bool) {
	if inst.expandedKeys == nil {
		inst.expandedKeys = make(map[string]bool)
	}
	inst.expandedKeys[key] = value
}

func normalizeExpandedKeys(keys map[string]bool) map[string]bool {
	if len(keys) == 0 {
		return nil
	}
	normalized := make(map[string]bool, len(keys))
	for key, value := range keys {
		if key == "" {
			continue
		}
		if strings.HasPrefix(key, "path:") || strings.HasPrefix(key, "id:") || strings.HasPrefix(key, "idx:") {
			normalized[key] = value
			continue
		}
		normalized["path:"+key] = value
	}
	return normalized
}

func normalizeNodeKeys(keys map[string]bool) map[string]bool {
	return normalizeExpandedKeys(keys)
}

func equalExpandedKeys(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if other, ok := right[key]; !ok || other != value {
			return false
		}
	}
	return true
}

func getExpandedKeysProp(props rtui.Props) map[string]bool {
	if value, ok := props["expandedKeys"]; ok {
		if keys, ok := value.(map[string]bool); ok {
			return cloneExpandedKeys(keys)
		}
	}
	return nil
}

func getCheckedKeysProp(props rtui.Props) map[string]bool {
	if value, ok := props["checkedKeys"]; ok {
		if keys, ok := value.(map[string]bool); ok {
			return cloneExpandedKeys(keys)
		}
	}
	return nil
}

func getSelectionModeProp(props rtui.Props, key string, def SelectionMode) SelectionMode {
	if value, ok := props[key]; ok {
		if mode, ok := value.(SelectionMode); ok {
			return mode
		}
	}
	return def
}

func getIntentProp(props rtui.Props, key string) intent.Intent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.Intent); ok {
			return result
		}
	}
	return nil
}

func getFieldIntentProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	if value, ok := props[key+"Field"]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	return nil
}

func getSearchFn(props rtui.Props) func(TreeNode, string) bool {
	if value, ok := props["searchFn"]; ok {
		if fn, ok := value.(func(TreeNode, string) bool); ok {
			return fn
		}
	}
	return nil
}

func getLazyLoadFn(props rtui.Props) func(TreeNode) {
	if value, ok := props["lazyLoadFn"]; ok {
		if fn, ok := value.(func(TreeNode)); ok {
			return fn
		}
	}
	return nil
}

func getLazyLoadChildrenFn(props rtui.Props) func(TreeNode) []TreeNode {
	if value, ok := props["lazyLoadChildrenFn"]; ok {
		if fn, ok := value.(func(TreeNode) []TreeNode); ok {
			return fn
		}
	}
	return nil
}

func sameSearchFn(left, right func(TreeNode, string) bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return reflect.ValueOf(left).Pointer() == reflect.ValueOf(right).Pointer()
}

func sameLazyLoadFn(left, right func(TreeNode)) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return reflect.ValueOf(left).Pointer() == reflect.ValueOf(right).Pointer()
}

func sameLazyLoadChildrenFn(left, right func(TreeNode) []TreeNode) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return reflect.ValueOf(left).Pointer() == reflect.ValueOf(right).Pointer()
}

func sameIntent(left, right intent.Intent) bool {
	return reflect.DeepEqual(left, right)
}

func sameFieldIntent(left, right intent.FieldIntent) bool {
	return reflect.DeepEqual(left, right)
}

func hasProp(props rtui.Props, key string) bool {
	_, ok := props[key]
	return ok
}

func nodeKey(node TreeNode, index int) string {
	if node.Path != "" {
		return "path:" + node.Path
	}
	if node.NodeID != 0 {
		return fmt.Sprintf("id:%d", node.NodeID)
	}
	return fmt.Sprintf("idx:%d", index)
}

func nodeDepth(node TreeNode) int {
	if node.Indent <= 0 {
		return 0
	}
	if indentStep <= 0 {
		return 0
	}
	depth := node.Indent / indentStep
	if depth < 0 {
		return 0
	}
	return depth
}

func nodesEqual(left, right []TreeNode) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].Indent != right[i].Indent ||
			left[i].Content != right[i].Content ||
			left[i].Path != right[i].Path ||
			left[i].NodeType != right[i].NodeType ||
			left[i].NodeID != right[i].NodeID ||
			left[i].Lazy != right[i].Lazy ||
			left[i].Loading != right[i].Loading ||
			left[i].LoadError != right[i].LoadError {
			return false
		}
	}
	return true
}

func truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if paint.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return trimToWidth(text, maxWidth)
	}
	runes := []rune(text)
	for end := len(runes); end > 0; end-- {
		candidate := string(runes[:end])
		if paint.StringWidth(candidate)+3 <= maxWidth {
			return candidate + "..."
		}
	}
	return trimToWidth(text, maxWidth)
}

func trimToWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	currentWidth := 0
	for _, character := range text {
		charWidth := paint.RuneWidth(character)
		if currentWidth+charWidth > maxWidth {
			break
		}
		builder.WriteRune(character)
		currentWidth += charWidth
	}
	return builder.String()
}

func padRightToWidth(text string, width int) string {
	textWidth := paint.StringWidth(text)
	if textWidth >= width {
		return trimToWidth(text, width)
	}
	return text + strings.Repeat(" ", width-textWidth)
}

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

func getRowStyleFn(props rtui.Props) func(int, TreeNode) style.Style {
	if v, ok := props["rowStyleFn"]; ok {
		if fn, ok := v.(func(int, TreeNode) style.Style); ok {
			return fn
		}
	}
	return nil
}

func sameRowStyleFn(left, right func(int, TreeNode) style.Style) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return reflect.ValueOf(left).Pointer() == reflect.ValueOf(right).Pointer()
}
