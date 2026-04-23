package treeview

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

type dragReorderResult struct {
	fromIndex        int
	toIndex          int
	fromVisibleIndex int
	toVisibleIndex   int
	entry            nodeEntry
	parentKey        string
}

func (inst *Instance) handlePress(act *action.Action) bool {
	mouseMsg, ok := inst.mouseMsgFromAction(act)
	if !ok {
		return inst.handleActivate()
	}

	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		return false
	}
	rowIndex, ok := inst.rowIndexAtLocalY(mouseMsg.LocalY, inst.effectiveViewportHeight(len(visible)))
	if !ok {
		return false
	}
	target := inst.scrollOffset + rowIndex
	if target < 0 || target >= len(visible) {
		return false
	}
	entry := visible[target]

	if inst.allowExpand && entry.HasChildren && inst.clickOnExpander(entry, mouseMsg.LocalX) {
		inst.clearDragState()
		inst.selectVisibleIndex(target, visible, true)
		inst.toggleExpand()
		return true
	}

	if !inst.canStartReorder(entry) {
		return inst.activateVisibleEntry(target, visible)
	}

	inst.startDrag(entry, target)
	return true
}

func (inst *Instance) handleDragMove(act *action.Action) bool {
	if !inst.dragging {
		return false
	}

	mouseMsg, ok := inst.mouseMsgFromAction(act)
	if !ok {
		return false
	}

	visible, _ := inst.visibleEntries()
	if len(visible) == 0 {
		inst.clearDragState()
		return false
	}

	sourceVisible, sourceEntry, ok := inst.findVisibleEntryByKey(visible, inst.dragSourceKey)
	if !ok {
		inst.clearDragState()
		return false
	}

	rowIndex, ok := inst.rowIndexAtLocalY(mouseMsg.LocalY, inst.effectiveViewportHeight(len(visible)))
	if !ok {
		return true
	}
	targetVisible := inst.scrollOffset + rowIndex
	if targetVisible < 0 || targetVisible >= len(visible) || targetVisible == sourceVisible {
		return true
	}
	targetEntry := visible[targetVisible]
	if !inst.canDropReorder(sourceEntry, targetEntry) {
		return true
	}

	_, moved := inst.reorderDraggedSubtree(sourceEntry, targetEntry, sourceVisible, targetVisible)
	if !moved {
		return true
	}

	inst.dragMoved = true
	inst.dirty = true
	return true
}

func (inst *Instance) handleDragRelease(act *action.Action) bool {
	if !inst.dragging {
		return false
	}

	pendingSelect := inst.dragPendingSelect
	visible, _ := inst.visibleEntries()
	sourceVisible, _, ok := inst.findVisibleEntryByKey(visible, inst.dragSourceKey)

	result, emit := inst.finishDrag(visible)
	if !emit && pendingSelect && ok {
		return inst.activateVisibleEntry(sourceVisible, visible)
	}
	if emit {
		inst.emitNodeReorder(result)
		return true
	}
	return ok
}

func (inst *Instance) activateVisibleEntry(target int, visible []nodeEntry) bool {
	if target < 0 || target >= len(visible) {
		return false
	}
	inst.selectVisibleIndex(target, visible, true)
	entry := visible[target]
	if inst.selectionMode != SelectionNone {
		return inst.toggleChecked(entry)
	}
	return true
}

func (inst *Instance) canStartReorder(entry nodeEntry) bool {
	if !inst.reorderable || inst.searchActive() {
		return false
	}
	if strings.HasPrefix(entry.Key, "idx:") {
		return false
	}
	return true
}

func (inst *Instance) canDropReorder(source, target nodeEntry) bool {
	if source.Depth != target.Depth {
		return false
	}
	if strings.HasPrefix(target.Key, "idx:") {
		return false
	}
	return inst.parentKeyForIndex(source.Index) == inst.parentKeyForIndex(target.Index)
}

func (inst *Instance) startDrag(entry nodeEntry, visibleIndex int) {
	inst.dragging = true
	inst.dragMoved = false
	inst.dragSourceKey = entry.Key
	inst.dragSourceParentKey = inst.parentKeyForIndex(entry.Index)
	inst.dragSourceNodeIndex = entry.Index
	inst.dragSourceVisibleIndex = visibleIndex
	inst.dragPendingSelect = inst.selectedIndex != visibleIndex
	inst.dirty = true
}

func (inst *Instance) finishDrag(visible []nodeEntry) (dragReorderResult, bool) {
	if !inst.dragging {
		return dragReorderResult{}, false
	}
	sourceVisible, sourceEntry, ok := inst.findVisibleEntryByKey(visible, inst.dragSourceKey)
	if !ok {
		inst.clearDragState()
		return dragReorderResult{}, false
	}

	result := dragReorderResult{
		fromVisibleIndex: inst.dragSourceVisibleIndex,
		toVisibleIndex:   sourceVisible,
		fromIndex:        inst.dragSourceNodeIndex,
		toIndex:          sourceEntry.Index,
		entry:            sourceEntry,
		parentKey:        inst.dragSourceParentKey,
	}
	emit := inst.dragMoved && inst.dragSourceVisibleIndex != sourceVisible
	inst.clearDragState()
	return result, emit
}

func (inst *Instance) clearDragState() {
	if !inst.dragging && !inst.dragMoved && inst.dragSourceKey == "" {
		return
	}
	inst.dragging = false
	inst.dragMoved = false
	inst.dragSourceKey = ""
	inst.dragSourceParentKey = ""
	inst.dragSourceNodeIndex = -1
	inst.dragSourceVisibleIndex = -1
	inst.dragPendingSelect = false
	inst.dirty = true
}

func (inst *Instance) mouseMsgFromAction(act *action.Action) (*runtimemsg.MouseMsg, bool) {
	if act == nil {
		return nil, false
	}
	switch payload := act.Payload.(type) {
	case *runtimemsg.MouseMsg:
		if payload == nil {
			return nil, false
		}
		return payload, true
	case runtimemsg.MouseMsg:
		return &payload, true
	default:
		return nil, false
	}
}

func (inst *Instance) parentKeyForIndex(index int) string {
	if index <= 0 || index >= len(inst.nodes) {
		return ""
	}
	depth := nodeDepth(inst.nodes[index])
	if depth == 0 {
		return ""
	}
	for i := index - 1; i >= 0; i-- {
		if nodeDepth(inst.nodes[i]) == depth-1 {
			return nodeKey(inst.nodes[i], i)
		}
	}
	return ""
}

func (inst *Instance) subtreeRange(index int) (int, int) {
	if index < 0 || index >= len(inst.nodes) {
		return -1, -1
	}
	start := index
	end := index + 1
	depth := nodeDepth(inst.nodes[index])
	for end < len(inst.nodes) && nodeDepth(inst.nodes[end]) > depth {
		end++
	}
	return start, end
}

func (inst *Instance) reorderDraggedSubtree(source, target nodeEntry, sourceVisible, targetVisible int) (dragReorderResult, bool) {
	fromStart, fromEnd := inst.subtreeRange(source.Index)
	toStart, toEnd := inst.subtreeRange(target.Index)
	if fromStart < 0 || toStart < 0 {
		return dragReorderResult{}, false
	}

	insertIndex := toStart
	if targetVisible > sourceVisible {
		insertIndex = toEnd
	}
	blockLen := fromEnd - fromStart
	if fromStart < insertIndex {
		insertIndex -= blockLen
	}
	if insertIndex == fromStart {
		return dragReorderResult{}, false
	}

	selectedKey := inst.selectedEntryKey()
	block := append([]TreeNode(nil), inst.nodes[fromStart:fromEnd]...)
	next := append([]TreeNode(nil), inst.nodes[:fromStart]...)
	next = append(next, inst.nodes[fromEnd:]...)
	if insertIndex >= len(next) {
		next = append(next, block...)
	} else {
		next = append(next[:insertIndex], append(block, next[insertIndex:]...)...)
	}
	inst.nodes = next
	inst.invalidateCache()
	inst.restoreSelectedEntryKey(selectedKey)
	inst.dirty = true

	visible, _ := inst.visibleEntries()
	newVisible, entry, ok := inst.findVisibleEntryByKey(visible, source.Key)
	if !ok {
		return dragReorderResult{}, false
	}
	return dragReorderResult{
		fromIndex:        fromStart,
		toIndex:          entry.Index,
		fromVisibleIndex: sourceVisible,
		toVisibleIndex:   newVisible,
		entry:            entry,
		parentKey:        inst.parentKeyForIndex(entry.Index),
	}, true
}

func (inst *Instance) selectedEntryKey() string {
	visible, _ := inst.visibleEntries()
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(visible) {
		return ""
	}
	key := visible[inst.selectedIndex].Key
	if strings.HasPrefix(key, "idx:") {
		return ""
	}
	return key
}

func (inst *Instance) restoreSelectedEntryKey(key string) {
	visible, _ := inst.visibleEntries()
	if key == "" {
		inst.normalizeSelectionAndScroll()
		return
	}
	for i, entry := range visible {
		if entry.Key == key {
			inst.selectedIndex = i
			inst.normalizeSelectionAndScroll()
			return
		}
	}
	inst.normalizeSelectionAndScroll()
}

func (inst *Instance) findVisibleEntryByKey(visible []nodeEntry, key string) (int, nodeEntry, bool) {
	for i, entry := range visible {
		if entry.Key == key {
			return i, entry, true
		}
	}
	return -1, nodeEntry{}, false
}

func (inst *Instance) searchActive() bool {
	return strings.TrimSpace(inst.searchQuery) != ""
}
