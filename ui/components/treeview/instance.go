package treeview

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
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
	nodes                    []TreeNode
	expandLevel              int
	showIcons                bool
	showLineNums             bool
	compact                  bool
	showBorder               bool
	showScrollbar            bool
	treeStyle                style.Style
	selectedStyle            style.Style
	iconStyle                style.Style
	scrollbarStyle           style.Style
	rowStyleFn               func(int, TreeNode) style.Style
	matchStyle               style.Style
	searchMatches            map[string]bool
	searchMatchesControlled  bool
	searchPending            bool
	searchPageSize           int
	searchQuery              string
	searchQueryControlled    bool
	searchFn                 func(TreeNode, string) bool
	selectionMode            SelectionMode
	checkedKeys              map[string]bool
	checkedKeysControlled    bool
	selectionIntent          intent.Intent
	selectionIntentField     intent.FieldIntent
	reorderIntent            intent.Intent
	lazyLoadFn               func(TreeNode)
	lazyLoadChildrenFn       func(TreeNode) []TreeNode
	showSearchStats          bool
	searchStatsStyle         style.Style
	scrollOffset             int
	scrollOffsetControlled   bool
	selectedIndex            int
	selectedIndexControlled  bool
	viewportHeight           int
	expandedKeys             map[string]bool
	expandedKeysControlled   bool
	lastExternalExpandedKeys map[string]bool
	lastExternalNodes        []TreeNode
	allowScroll              bool
	allowExpand              bool
	reorderable              bool

	// === Runtime State ===
	focused                  bool
	expandState              map[string]bool // Expand/collapse state for each node (keyed by nodeKey)
	lazyRequested            map[string]bool
	lazyInsertions           map[string][]TreeNode
	lastSearchTotal          int
	lastSearchSelected       int
	lastSearchQuery          string
	lastSearchPending        bool
	lastSearchPage           int
	lastSearchPageCount      int
	lastSearchPageSize       int
	lastSearchResultsDigest  string
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
	dragging                 bool
	dragMoved                bool
	dragSourceKey            string
	dragSourceParentKey      string
	dragSourceNodeIndex      int
	dragSourceVisibleIndex   int
	dragPendingSelect        bool

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
		key:                     proputil.GetString(props, "key", ""),
		componentID:             proputil.GetString(props, "componentID", ""),
		nodes:                   getNodesProp(props, []TreeNode{}),
		expandLevel:             proputil.GetInt(props, "expandLevel", 1),
		showIcons:               proputil.GetBool(props, "showIcons", true),
		showLineNums:            proputil.GetBool(props, "showLineNums", false),
		compact:                 proputil.GetBool(props, "compact", false),
		showBorder:              proputil.GetBool(props, "showBorder", true),
		showScrollbar:           proputil.GetBool(props, "showScrollbar", true),
		treeStyle:               proputil.GetStyle(props, "treeStyle", style.Style{}),
		selectedStyle:           proputil.GetStyle(props, "selectedStyle", style.Style{}),
		iconStyle:               proputil.GetStyle(props, "iconStyle", style.Style{}),
		scrollbarStyle:          proputil.GetStyle(props, "scrollbarStyle", style.Style{}),
		rowStyleFn:              getRowStyleFn(props),
		matchStyle:              proputil.GetStyle(props, "matchStyle", style.Style{}),
		searchMatches:           normalizeNodeKeys(getSearchMatchesProp(props)),
		searchMatchesControlled: proputil.GetBool(props, "searchMatchesControlled", false),
		searchPending:           proputil.GetBool(props, "searchPending", false),
		searchPageSize:          max(0, proputil.GetInt(props, "searchPageSize", 0)),
		searchQuery:             proputil.GetString(props, "searchQuery", ""),
		searchQueryControlled:   proputil.GetBool(props, "searchQueryControlled", false),
		searchFn:                getSearchFn(props),
		selectionMode:           getSelectionModeProp(props, "selectionMode", SelectionNone),
		checkedKeys:             normalizeNodeKeys(getCheckedKeysProp(props)),
		checkedKeysControlled:   proputil.GetBool(props, "checkedKeysControlled", false),
		selectionIntent:         proputil.GetIntent(props, "selectionIntent", nil),
		selectionIntentField:    getFieldIntentProp(props, "selectionIntentField"),
		reorderIntent:           proputil.GetIntent(props, propReorderIntent, nil),
		lazyLoadFn:              getLazyLoadFn(props),
		lazyLoadChildrenFn:      getLazyLoadChildrenFn(props),
		showSearchStats:         proputil.GetBool(props, "showSearchStats", false),
		searchStatsStyle:        proputil.GetStyle(props, "searchStatsStyle", style.Style{}),
		scrollOffset:            proputil.GetInt(props, "scrollOffset", 0),
		scrollOffsetControlled:  proputil.GetBool(props, "scrollOffsetControlled", false),
		selectedIndex:           proputil.GetInt(props, "selectedIndex", -1),
		selectedIndexControlled: proputil.GetBool(props, "selectedIndexControlled", false),
		viewportHeight:          proputil.GetInt(props, "viewportHeight", 10),
		expandedKeys:            getExpandedKeysProp(props),
		expandedKeysControlled:  proputil.GetBool(props, "expandedKeysControlled", false),
		allowScroll:             proputil.GetBool(props, "allowScroll", true),
		allowExpand:             proputil.GetBool(props, "allowExpand", true),
		reorderable:             proputil.GetBool(props, propReorderable, false),
		expandState:             make(map[string]bool),
		lazyRequested:           make(map[string]bool),
		lazyInsertions:          make(map[string][]TreeNode),
		dirty:                   true,
		cacheDirty:              true,
	}

	if inst.expandedKeys != nil && len(inst.expandedKeys) > 0 {
		inst.expandedKeysControlled = true
	}
	if hasProp(props, propSearchMatches) {
		inst.searchMatchesControlled = true
	}
	if inst.checkedKeys != nil && len(inst.checkedKeys) > 0 {
		inst.checkedKeysControlled = true
	}
	inst.scrollOffsetInitialized = inst.scrollOffsetControlled || hasProp(props, "scrollOffset")
	inst.selectedIndexInitialized = inst.selectedIndexControlled || hasProp(props, "selectedIndex")
	inst.checkedKeysInitialized = inst.checkedKeysControlled || hasProp(props, "checkedKeys")
	inst.searchQueryInitialized = inst.searchQueryControlled || hasProp(props, "searchQuery")

	// Initialize lastExternalExpandedKeys so that SetProps can correctly detect
	// changes when the initial expandedKeys is non-empty (e.g. {"root": true})
	// and a subsequent update passes an empty map.
	if hasProp(props, "expandedKeys") {
		inst.lastExternalExpandedKeys = cloneExpandedKeys(normalizeExpandedKeys(inst.expandedKeys))
	}
	if hasProp(props, "nodes") {
		inst.lastExternalNodes = append([]TreeNode(nil), inst.nodes...)
	}

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

		// In controlled mode, only use keys explicitly provided by external source.
		// Do not fall back to expandLevel defaults.
		if inst.expandedKeysControlled {
			next[key] = false
			continue
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
	oldSearchMatches := cloneExpandedKeys(inst.searchMatches)
	oldSearchMatchesControlled := inst.searchMatchesControlled
	oldSearchPending := inst.searchPending
	oldSearchPageSize := inst.searchPageSize
	oldSearchQuery := inst.searchQuery
	oldSearchFn := inst.searchFn
	oldSelectionMode := inst.selectionMode
	oldCheckedKeys := cloneExpandedKeys(inst.checkedKeys)
	oldCheckedKeysControlled := inst.checkedKeysControlled
	oldSelectionIntent := inst.selectionIntent
	oldSelectionIntentField := inst.selectionIntentField
	oldReorderIntent := inst.reorderIntent
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
	oldReorderable := inst.reorderable

	inst.componentID = proputil.GetString(props, "componentID", inst.componentID)
	externalNodesChanged := false
	if hasProp(props, "nodes") {
		incomingNodes := getNodesProp(props, inst.nodes)
		if !nodesEqual(incomingNodes, inst.lastExternalNodes) {
			inst.lastExternalNodes = append([]TreeNode(nil), incomingNodes...)
			inst.nodes = incomingNodes
			externalNodesChanged = true
		}
	}
	inst.expandLevel = proputil.GetInt(props, "expandLevel", inst.expandLevel)
	inst.showIcons = proputil.GetBool(props, "showIcons", inst.showIcons)
	inst.showLineNums = proputil.GetBool(props, "showLineNums", inst.showLineNums)
	inst.compact = proputil.GetBool(props, "compact", inst.compact)
	inst.showBorder = proputil.GetBool(props, "showBorder", inst.showBorder)
	inst.showScrollbar = proputil.GetBool(props, "showScrollbar", inst.showScrollbar)
	inst.treeStyle = proputil.GetStyle(props, "treeStyle", style.Style{})
	inst.selectedStyle = proputil.GetStyle(props, "selectedStyle", style.Style{})
	inst.iconStyle = proputil.GetStyle(props, "iconStyle", style.Style{})
	inst.scrollbarStyle = proputil.GetStyle(props, "scrollbarStyle", style.Style{})
	if fn, ok := props[propRowStyleFn].(func(int, TreeNode) style.Style); ok {
		inst.rowStyleFn = fn
	} else if _, exists := props[propRowStyleFn]; exists {
		inst.rowStyleFn = nil
	}
	inst.matchStyle = proputil.GetStyle(props, "matchStyle", style.Style{})
	if searchMatches, ok := props[propSearchMatches].(map[string]bool); ok {
		inst.searchMatches = normalizeNodeKeys(searchMatches)
		inst.searchMatchesControlled = true
	}
	if controlled, ok := props[propSearchMatchesControlled].(bool); ok {
		inst.searchMatchesControlled = controlled
	}
	if inst.searchMatchesControlled && hasProp(props, propSearchMatches) {
		inst.searchMatches = normalizeNodeKeys(getSearchMatchesProp(props))
	}
	inst.searchPending = proputil.GetBool(props, "searchPending", inst.searchPending)
	inst.searchPageSize = max(0, proputil.GetInt(props, "searchPageSize", inst.searchPageSize))
	if controlled, ok := props[propSearchQueryControlled].(bool); ok {
		inst.searchQueryControlled = controlled
	}
	if inst.searchQueryControlled {
		inst.searchQuery = proputil.GetString(props, "searchQuery", inst.searchQuery)
		inst.searchQueryInitialized = true
	} else if query, ok := props[propSearchQuery].(string); ok && !inst.searchQueryInitialized {
		inst.searchQuery = query
		inst.searchQueryInitialized = true
	}
	if fn, ok := props[propSearchFn].(func(TreeNode, string) bool); ok {
		inst.searchFn = fn
	} else if _, exists := props[propSearchFn]; exists {
		inst.searchFn = nil
	}
	inst.selectionMode = getSelectionModeProp(props, "selectionMode", inst.selectionMode)
	inst.selectionIntent = proputil.GetIntent(props, "selectionIntent", nil)
	inst.selectionIntentField = getFieldIntentProp(props, "selectionIntentField")
	inst.reorderIntent = proputil.GetIntent(props, propReorderIntent, inst.reorderIntent)
	inst.showSearchStats = proputil.GetBool(props, "showSearchStats", inst.showSearchStats)
	inst.searchStatsStyle = proputil.GetStyle(props, "searchStatsStyle", style.Style{})
	if fn, ok := props[propLazyLoadFn].(func(TreeNode)); ok {
		inst.lazyLoadFn = fn
	} else if _, exists := props[propLazyLoadFn]; exists {
		inst.lazyLoadFn = nil
	}
	if fn, ok := props[propLazyLoadChildrenFn].(func(TreeNode) []TreeNode); ok {
		inst.lazyLoadChildrenFn = fn
	} else if _, exists := props[propLazyLoadChildrenFn]; exists {
		inst.lazyLoadChildrenFn = nil
	}
	if controlled, ok := props[propScrollOffsetControlled].(bool); ok {
		inst.scrollOffsetControlled = controlled
	}
	if inst.scrollOffsetControlled {
		inst.scrollOffset = proputil.GetInt(props, "scrollOffset", inst.scrollOffset)
		inst.scrollOffsetInitialized = true
	} else if offset, ok := props[propScrollOffset].(int); ok && !inst.scrollOffsetInitialized {
		inst.scrollOffset = offset
		inst.scrollOffsetInitialized = true
	}
	if controlled, ok := props[propSelectedIndexControlled].(bool); ok {
		inst.selectedIndexControlled = controlled
	}
	if inst.selectedIndexControlled {
		inst.selectedIndex = proputil.GetInt(props, "selectedIndex", inst.selectedIndex)
		inst.selectedIndexInitialized = true
	} else if index, ok := props[propSelectedIndex].(int); ok && !inst.selectedIndexInitialized {
		inst.selectedIndex = index
		inst.selectedIndexInitialized = true
	}
	if controlled, ok := props[propCheckedKeysControlled].(bool); ok {
		inst.checkedKeysControlled = controlled
	}
	if inst.checkedKeysControlled {
		inst.checkedKeys = normalizeNodeKeys(getCheckedKeysProp(props))
		inst.checkedKeysInitialized = true
	} else if _, ok := props[propCheckedKeys].(map[string]bool); ok && !inst.checkedKeysInitialized {
		inst.checkedKeys = normalizeNodeKeys(getCheckedKeysProp(props))
		inst.checkedKeysInitialized = true
	}
	inst.viewportHeight = proputil.GetInt(props, "viewportHeight", inst.viewportHeight)
	externalExpandedKeysChanged := false
	if expandedKeys, ok := props[propExpandedKeys].(map[string]bool); ok {
		// Only overwrite internal expandedKeys when the external value has actually
		// changed since the last SetProps call. This prevents a same-frame
		// SetProps (triggered by dirty=true before the intent is processed) from
		// undoing an expand/collapse that navigateLeft/Right just performed.
		normalized := normalizeExpandedKeys(expandedKeys)
		if !equalExpandedKeys(normalized, inst.lastExternalExpandedKeys) {
			inst.lastExternalExpandedKeys = cloneExpandedKeys(normalized)
			inst.expandedKeys = cloneExpandedKeys(normalized)
			externalExpandedKeysChanged = true
		}
		inst.expandedKeysControlled = true
	}
	if controlled, ok := props[propExpandedKeysControlled].(bool); ok {
		inst.expandedKeysControlled = controlled
	}
	inst.allowScroll = proputil.GetBool(props, "allowScroll", inst.allowScroll)
	inst.allowExpand = proputil.GetBool(props, "allowExpand", inst.allowExpand)
	inst.reorderable = proputil.GetBool(props, propReorderable, inst.reorderable)

	lazyInserted := inst.reapplyLazyInsertions()
	nodesChanged := externalNodesChanged || !nodesEqual(oldNodes, inst.nodes) || lazyInserted
	expandLevelChanged := oldExpandLevel != inst.expandLevel
	expandedKeysChanged := !equalExpandedKeys(oldExpandedKeys, inst.expandedKeys) || oldExpandedKeysControlled != inst.expandedKeysControlled
	checkedKeysChanged := !equalExpandedKeys(oldCheckedKeys, inst.checkedKeys) || oldCheckedKeysControlled != inst.checkedKeysControlled
	searchMatchesChanged := !equalExpandedKeys(oldSearchMatches, inst.searchMatches) || oldSearchMatchesControlled != inst.searchMatchesControlled
	searchChanged := oldSearchQuery != inst.searchQuery ||
		!sameSearchFn(oldSearchFn, inst.searchFn) ||
		searchMatchesChanged ||
		oldSearchPending != inst.searchPending
	searchPresentationChanged := searchChanged || oldSearchPageSize != inst.searchPageSize
	if searchChanged && !inst.selectedIndexControlled {
		inst.autoSelectMatch = true
	}
	if !inst.reorderable || nodesChanged || searchChanged {
		inst.clearDragState()
	}
	if inst.expandedKeysControlled {
		if nodesChanged || expandLevelChanged || externalExpandedKeysChanged {
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
		oldSearchPageSize != inst.searchPageSize ||
		oldSelectionMode != inst.selectionMode ||
		checkedKeysChanged ||
		!sameIntent(oldSelectionIntent, inst.selectionIntent) ||
		!sameFieldIntent(oldSelectionIntentField, inst.selectionIntentField) ||
		!sameIntent(oldReorderIntent, inst.reorderIntent) ||
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
		oldAllowExpand != inst.allowExpand ||
		oldReorderable != inst.reorderable
	if searchPresentationChanged {
		changed = true
	}
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	props := rtui.Props{
		propKey:                     inst.key,
		propComponentID:             inst.componentID,
		propNodes:                   inst.nodes,
		propExpandLevel:             inst.expandLevel,
		propShowIcons:               inst.showIcons,
		propShowLineNums:            inst.showLineNums,
		propCompact:                 inst.compact,
		propShowBorder:              inst.showBorder,
		propShowScrollbar:           inst.showScrollbar,
		propMatchStyle:              inst.matchStyle,
		propSearchQuery:             inst.searchQuery,
		propSearchQueryControlled:   inst.searchQueryControlled,
		propSelectionMode:           inst.selectionMode,
		propCheckedKeysControlled:   inst.checkedKeysControlled,
		propSelectionIntent:         inst.selectionIntent,
		propSelectionIntentField:    inst.selectionIntentField,
		propReorderIntent:           inst.reorderIntent,
		propLazyLoadFn:              inst.lazyLoadFn,
		propLazyLoadChildrenFn:      inst.lazyLoadChildrenFn,
		propShowSearchStats:         inst.showSearchStats,
		propSearchStatsStyle:        inst.searchStatsStyle,
		propScrollOffsetControlled:  inst.scrollOffsetControlled,
		propSelectedIndexControlled: inst.selectedIndexControlled,
		propScrollOffset:            inst.scrollOffset,
		propSelectedIndex:           inst.selectedIndex,
		propViewportHeight:          inst.viewportHeight,
		propExpandedKeysControlled:  inst.expandedKeysControlled,
		propSearchMatchesControlled: inst.searchMatchesControlled,
		propSearchPending:           inst.searchPending,
		propSearchPageSize:          inst.searchPageSize,
		propAllowScroll:             inst.allowScroll,
		propAllowExpand:             inst.allowExpand,
		propReorderable:             inst.reorderable,
	}
	if inst.expandedKeysControlled {
		props[propExpandedKeys] = cloneExpandedKeys(inst.expandedKeys)
	}
	if inst.searchMatchesControlled {
		props[propSearchMatches] = cloneExpandedKeys(inst.searchMatches)
	}
	if inst.checkedKeysControlled {
		props[propCheckedKeys] = cloneExpandedKeys(inst.checkedKeys)
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
	if value {
		inst.expandedKeys[key] = true
	} else {
		delete(inst.expandedKeys, key)
	}
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
	if value, ok := props[propExpandedKeys]; ok {
		if keys, ok := value.(map[string]bool); ok {
			return cloneExpandedKeys(keys)
		}
	}
	return nil
}

func getSearchMatchesProp(props rtui.Props) map[string]bool {
	if value, ok := props[propSearchMatches]; ok {
		if keys, ok := value.(map[string]bool); ok {
			return cloneExpandedKeys(keys)
		}
	}
	return nil
}

func getCheckedKeysProp(props rtui.Props) map[string]bool {
	if value, ok := props[propCheckedKeys]; ok {
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
	if value, ok := props[propSearchFn]; ok {
		if fn, ok := value.(func(TreeNode, string) bool); ok {
			return fn
		}
	}
	return nil
}

func getLazyLoadFn(props rtui.Props) func(TreeNode) {
	if value, ok := props[propLazyLoadFn]; ok {
		if fn, ok := value.(func(TreeNode)); ok {
			return fn
		}
	}
	return nil
}

func getLazyLoadChildrenFn(props rtui.Props) func(TreeNode) []TreeNode {
	if value, ok := props[propLazyLoadChildrenFn]; ok {
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

func getNodesProp(props rtui.Props, def []TreeNode) []TreeNode {
	v, ok := props[propNodes]
	if !ok {
		return def
	}
	if nodes, ok := v.([]TreeNode); ok {
		return nodes
	}
	return def
}

func getRowStyleFn(props rtui.Props) func(int, TreeNode) style.Style {
	if v, ok := props[propRowStyleFn]; ok {
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
