package layout

import (
	"sync"
)

// ==============================================================================
// HitMap (V3)
// ==============================================================================
// HitMap 用于快速查找布局树中位于特定位置的节点

// HitMapEntry 命中映射条目
type HitMapEntry struct {
	NodeID string
	Rect   Rect
	ZIndex int
}

const hitMapLayerZStride = 10000

// HitMap 命中映射表
// 用于快速查找给定坐标处的所有节点
type HitMap struct {
	entries map[string]*HitMapEntry
	mutex   sync.RWMutex
}

// NewHitMap 创建新的命中映射表
func NewHitMap() *HitMap {
	return &HitMap{
		entries: make(map[string]*HitMapEntry),
	}
}

// BuildFromLayoutBox 从布局盒子构建命中映射表
func (hm *HitMap) BuildFromLayoutBox(box *LayoutBox) {
	if box == nil {
		return
	}

	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	// 清空现有条目
	hm.entries = make(map[string]*HitMapEntry)

	// 递归构建条目。Hit testing must follow the same effective layer model as
	// painting: children inherit their parent's layer, and non-base layers sort
	// above deeper base-layer content.
	hm.buildFromLayoutBoxRecursive(box, 0, LayerBase)
}

// buildFromLayoutBoxRecursive 递归构建命中映射条目
func (hm *HitMap) buildFromLayoutBoxRecursive(box *LayoutBox, zIndex int, inheritedLayer Layer) {
	if box == nil {
		return
	}

	effectiveLayer := box.Layer
	if effectiveLayer < inheritedLayer {
		effectiveLayer = inheritedLayer
	}

	// Use box.ZIndex if it's explicitly set (non-zero),
	// otherwise use the recursive zIndex parameter
	// This allows Portals to have their own ZIndex range (1000+)
	// while main tree nodes use the recursive depth-based ZIndex
	actualZIndex := zIndex
	if box.ZIndex != 0 {
		actualZIndex = box.ZIndex
	}
	actualZIndex += effectiveLayer.ZIndex() * hitMapLayerZStride

	// 为当前盒子创建条目
	rect := Rect{
		X:      box.X,
		Y:      box.Y,
		Width:  box.Width,
		Height: box.Height,
	}

	hm.entries[box.ID] = &HitMapEntry{
		NodeID: box.ID,
		Rect:   rect,
		ZIndex: actualZIndex,
	}

	// 递归处理子节点
	// Portal nodes (box.ZIndex != 0): each child gets unique local ZIndex
	// Main tree nodes (box.ZIndex == 0): all children get same ZIndex (backward compatible)
	for i, child := range box.Children {
		if box.ZIndex != 0 {
			// Portal: each child gets unique local ZIndex (parent.ZIndex + 1, +2, +3, ...)
			hm.buildFromLayoutBoxRecursive(child, box.ZIndex+1+i, effectiveLayer)
		} else {
			// Main tree: all children get same ZIndex (backward compatible)
			hm.buildFromLayoutBoxRecursive(child, zIndex+1, effectiveLayer)
		}
	}
}

// HitTest 测试点 (x, y) 是否命中任何节点
// 返回第一个命中的节点（Z 顺序最高的）
func (hm *HitMap) HitTest(x, y int) *HitMapEntry {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	var result *HitMapEntry
	maxZIndex := -1

	// 查找所有包含该点的节点
	for _, entry := range hm.entries {
		if entry.Rect.Contains(x, y) {
			// 返回 Z 顺序最高的节点
			if entry.ZIndex > maxZIndex {
				maxZIndex = entry.ZIndex
				result = entry
			}
		}
	}

	return result
}

// HitTestAll 测试点 (x, y) 并返回所有命中的节点
// 按 Z 顺序排序（从低到高）
func (hm *HitMap) HitTestAll(x, y int) []*HitMapEntry {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	var results []*HitMapEntry

	// 收集所有包含该点的节点
	for _, entry := range hm.entries {
		if entry.Rect.Contains(x, y) {
			results = append(results, entry)
		}
	}

	// 按 Z 顺序排序（从低到高）
	sortByZIndex(results)

	return results
}

// Get 获取特定节点的命中映射条目
func (hm *HitMap) Get(nodeID string) *HitMapEntry {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	return hm.entries[nodeID]
}

// GetAll 获取所有命中映射条目
func (hm *HitMap) GetAll() []*HitMapEntry {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	results := make([]*HitMapEntry, 0, len(hm.entries))
	for _, entry := range hm.entries {
		results = append(results, entry)
	}

	return results
}

// Clear 清空命中映射表
func (hm *HitMap) Clear() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.entries = make(map[string]*HitMapEntry)
}

// Size 返回命中映射表中的条目数量
func (hm *HitMap) Size() int {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	return len(hm.entries)
}

// sortByZIndex 按 Z 顺序排序命中映射条目（从低到高）
func sortByZIndex(entries []*HitMapEntry) {
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].ZIndex > entries[j].ZIndex {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}
