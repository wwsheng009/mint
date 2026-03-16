package list

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/virtuallist"
)

// VirtualBridge snapshots a List into a VirtualList plus source-row index mapping.
type VirtualBridge struct {
	VNode         *virtuallist.VNode
	SourceIndices []int
}

// ToVirtualList snapshots the declarative List into a VirtualList bridge node.
func (v *VNode) ToVirtualList() *virtuallist.VNode {
	return v.ToVirtualBridge().VNode
}

// ToVirtualBridge snapshots the declarative List into a VirtualList bridge.
func (v *VNode) ToVirtualBridge() *VirtualBridge {
	inst := v.CreateInstance().(*Instance)
	return inst.ToVirtualBridge()
}

// ToVirtualList snapshots the current runtime List state into a VirtualList bridge node.
func (inst *Instance) ToVirtualList() *virtuallist.VNode {
	return inst.ToVirtualBridge().VNode
}

// ToVirtualBridge snapshots the current runtime List state into a VirtualList bridge.
func (inst *Instance) ToVirtualBridge() *VirtualBridge {
	sourceIndices := inst.visibleRowIndices()
	items := inst.virtualBridgeItems(sourceIndices)
	selectedIndex := inst.virtualBridgeSelectedIndex(sourceIndices)
	width := max(4, inst.calculateWidth())
	visibleCount := max(1, inst.visibleHeight())
	height := max(2, visibleCount+2)

	vnode := virtuallist.New().
		SetItems(items).
		SetItemCount(len(items)).
		SetVisibleCount(visibleCount).
		SetHeight(height).
		SetWidth(width).
		SetAllowScroll(inst.allowScroll).
		SetScrollOffset(inst.virtualBridgeScrollOffset(len(items), visibleCount)).
		SetSelectedIndex(selectedIndex).
		SetListStyle(inst.rowStyle).
		SetSelectedStyle(inst.selectedStyle).
		SetItemStyleFn(inst.virtualBridgeItemStyleFn(sourceIndices))

	return &VirtualBridge{
		VNode:         vnode,
		SourceIndices: append([]int(nil), sourceIndices...),
	}
}

// SourceIndex maps a VirtualList item index back to the source List row index.
func (b *VirtualBridge) SourceIndex(virtualIndex int) (int, bool) {
	if b == nil || virtualIndex < 0 || virtualIndex >= len(b.SourceIndices) {
		return -1, false
	}
	return b.SourceIndices[virtualIndex], true
}

// SyncToList applies VirtualList scroll/selection state back onto the List runtime.
func (b *VirtualBridge) SyncToList(inst *Instance, scrollOffset, selectedIndex int) bool {
	if b == nil || inst == nil {
		return false
	}

	targetScroll := clampVirtualBridgeOffset(scrollOffset, len(b.SourceIndices), max(1, inst.visibleHeight()))
	targetSelected := -1
	if selectedIndex >= 0 {
		sourceIndex, ok := b.SourceIndex(selectedIndex)
		if !ok {
			return false
		}
		targetSelected = sourceIndex
	}

	oldScroll := inst.scrollOffset
	oldSelected := inst.selectedIndex
	scrollChanged := false
	selectedChanged := false

	if inst.scrollOffset != targetScroll {
		inst.scrollOffset = targetScroll
		inst.recordPendingScroll()
		inst.dirty = true
		scrollChanged = true
	}

	if inst.selectedIndex != targetSelected {
		inst.selectedIndex = targetSelected
		if targetSelected >= 0 {
			inst.ensureSelectedRowVisible()
		}
		inst.recordPendingSelected()
		inst.dirty = true
		selectedChanged = true
	}

	finalScroll := inst.scrollOffset
	if finalScroll != oldScroll {
		scrollChanged = true
	}
	if !scrollChanged && !selectedChanged {
		return false
	}

	if selectedChanged {
		inst.emitSelectionChanged()
		if inst.selectedIndex >= 0 {
			inst.emitRowSelect(inst.selectedIndex)
		}
		inst.emitSearchStats()
	}
	inst.emitStateChanged()
	if finalScroll != oldScroll {
		inst.emitScrollIntent(finalScroll - oldScroll)
	}
	_ = oldSelected
	return true
}

func (inst *Instance) virtualBridgeItems(sourceIndices []int) []string {
	if len(sourceIndices) == 0 {
		return []string{}
	}

	items := make([]string, 0, len(sourceIndices))
	for _, rowIndex := range sourceIndices {
		rowText := inst.rows[rowIndex]
		items = append(items, inst.rowDisplayText(rowIndex, rowText))
	}
	return items
}

func (inst *Instance) virtualBridgeItemStyleFn(sourceIndices []int) func(int, string) style.Style {
	return func(virtualIndex int, _ string) style.Style {
		if virtualIndex < 0 || virtualIndex >= len(sourceIndices) {
			return inst.rowStyle
		}
		rowIndex := sourceIndices[virtualIndex]
		rowText := inst.rows[rowIndex]
		return inst.virtualBridgeRowStyle(rowIndex, rowText, inst.searchActive())
	}
}

func (inst *Instance) virtualBridgeRowStyle(rowIndex int, rowText string, matched bool) style.Style {
	if matched && inst.matchStyle != (style.Style{}) {
		return inst.matchStyle
	}
	if inst.rowStyleFn != nil {
		return inst.rowStyleFn(rowIndex, rowText)
	}
	return inst.rowStyle
}

func (inst *Instance) virtualBridgeSelectedIndex(sourceIndices []int) int {
	if inst.selectedIndex < 0 || len(sourceIndices) == 0 {
		return -1
	}
	for visibleIndex, rowIndex := range sourceIndices {
		if rowIndex == inst.selectedIndex {
			return visibleIndex
		}
	}
	return -1
}

func (inst *Instance) virtualBridgeScrollOffset(itemCount, visibleCount int) int {
	return clampVirtualBridgeOffset(inst.scrollOffset, itemCount, visibleCount)
}

func clampVirtualBridgeOffset(offset, itemCount, visibleCount int) int {
	if offset < 0 {
		offset = 0
	}
	maxOffset := itemCount - visibleCount
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	return offset
}

func virtualBridgeWidthFromRows(rows []string) int {
	width := 4
	for _, row := range rows {
		width = max(width, paint.StringWidth(strings.TrimSpace(row))+4)
	}
	return width
}
