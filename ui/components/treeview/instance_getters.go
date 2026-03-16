package treeview

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

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

func (inst *Instance) emitNodeReorder(result dragReorderResult) {
	reorderIntent := NodeReorder(
		result.fromIndex,
		result.toIndex,
		result.fromVisibleIndex,
		result.toVisibleIndex,
		result.entry.Node.Path,
		result.entry.Node.NodeID,
		result.entry.Node.NodeType,
		result.entry.Node.Content,
		result.parentKey,
	)
	if inst.componentID != "" {
		reorderIntent = NodeReorderWithID(
			inst.componentID,
			result.fromIndex,
			result.toIndex,
			result.fromVisibleIndex,
			result.toVisibleIndex,
			result.entry.Node.Path,
			result.entry.Node.NodeID,
			result.entry.Node.NodeType,
			result.entry.Node.Content,
			result.parentKey,
		)
	}
	inst.emitOptionalGlobalIntent(reorderIntent)
	if inst.intentEmitter != nil && inst.reorderIntent != nil {
		inst.intentEmitter(inst.reorderIntent)
	}
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
