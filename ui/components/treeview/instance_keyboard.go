package treeview

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

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

