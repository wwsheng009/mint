package event

import (
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
)

// IncrementalUpdater 支持 HitMap 的增量更新
//
// IncrementalUpdater 可以跟踪组件树的变化，只重新构建受影响的部分，
// 而不是每次都完全重建 HitMap。
type IncrementalUpdater struct {
	hitMap    *HitMap
	root      layout.Node
	version   uint32
	dirtyNodes map[uint64]bool // 脏节点标记
}

// NewIncrementalUpdater 创建增量更新器
func NewIncrementalUpdater(hitMap *HitMap, root layout.Node) *IncrementalUpdater {
	return &IncrementalUpdater{
		hitMap:     hitMap,
		root:       root,
		version:    0,
		dirtyNodes: make(map[uint64]bool),
	}
}

// MarkDirty 标记节点为脏（需要重建）
func (iu *IncrementalUpdater) MarkDirty(nodeID uint64) {
	iu.dirtyNodes[nodeID] = true
}

// MarkSubtreeDirty 标记子树为脏
//
// 从指定节点开始，递归标记所有子节点为脏。
func (iu *IncrementalUpdater) MarkSubtreeDirty(node layout.Node) {
	if node == nil {
		return
	}

	// 标记当前节点
	if node.ID() != "" {
		iu.MarkDirty(stringToNodeID(node.ID()))
	}

	// 递归标记子节点
	for _, child := range node.Children() {
		iu.MarkSubtreeDirty(child)
	}
}

// IncrementalUpdate 执行增量更新
//
// 只重新构建标记为脏的节点及其父链，而不是重建整个 HitMap。
// 这可以显著提高性能，特别是在大型组件树中。
func (iu *IncrementalUpdater) IncrementalUpdate() error {
	if len(iu.dirtyNodes) == 0 {
		// 没有脏节点，无需更新
		return nil
	}

	// 1. 移除所有脏节点的旧条目
	iu.removeDirtyEntries()

	// 2. 重建脏节点的条目
	iu.rebuildDirtyEntries()

	// 3. 清空脏节点标记
	iu.dirtyNodes = make(map[uint64]bool)

	// 4. 增加版本号
	iu.version++

	return nil
}

// removeDirtyEntries 移除脏节点的旧条目
func (iu *IncrementalUpdater) removeDirtyEntries() {
	// 收集要移除的 ID
	dirtyIDs := make(map[uint64]bool)
	for id := range iu.dirtyNodes {
		dirtyIDs[id] = true
	}

	// 过滤掉脏节点的条目
	newEntries := make([]HitMapEntry, 0, len(iu.hitMap.entries))
	for _, entry := range iu.hitMap.entries {
		if !dirtyIDs[entry.NodeID] {
			newEntries = append(newEntries, entry)
		}
	}

	iu.hitMap.entries = newEntries
}

// rebuildDirtyEntries 重建脏节点的条目
func (iu *IncrementalUpdater) rebuildDirtyEntries() {
	// 找到所有脏节点并重建
	for nodeID := range iu.dirtyNodes {
		node := iu.findNode(iu.root, nodeID)
		if node != nil {
			iu.buildNodeEntry(node, 0)
		}
	}

	// 重新排序
	iu.hitMap.sortByZOrder()
}

// findNode 在组件树中查找节点
func (iu *IncrementalUpdater) findNode(root layout.Node, nodeID uint64) layout.Node {
	if root == nil {
		return nil
	}

	if stringToNodeID(root.ID()) == nodeID {
		return root
	}

	for _, child := range root.Children() {
		if found := iu.findNode(child, nodeID); found != nil {
			return found
		}
	}

	return nil
}

// buildNodeEntry 为节点构建 HitMapEntry
func (iu *IncrementalUpdater) buildNodeEntry(node layout.Node, zOrder int) {
	if node == nil {
		return
	}

	// 获取节点位置和尺寸
	x, y := node.GetPosition()
	width, height := node.GetSize()

	if width <= 0 || height <= 0 {
		return
	}

	// 创建条目
	entry := HitMapEntry{
		NodeID: stringToNodeID(node.ID()),
		Node:   node,
		Bounds: layout.Rect{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		},
		LocalXY: func(screenX, screenY int) (int, int) {
			return screenX - x, screenY - y
		},
		ZOrder: zOrder,
	}

	// 添加到 HitMap
	iu.hitMap.entries = append(iu.hitMap.entries, entry)

	// 递归处理子节点（子节点的 Z-order 更高）
	for _, child := range node.Children() {
		iu.buildNodeEntry(child, zOrder+1)
	}
}

// GetVersion 获取当前版本号
func (iu *IncrementalUpdater) GetVersion() uint32 {
	return iu.version
}

// IsDirty 检查是否有脏节点
func (iu *IncrementalUpdater) IsDirty() bool {
	return len(iu.dirtyNodes) > 0
}

// GetDirtyCount 获取脏节点数量
func (iu *IncrementalUpdater) GetDirtyCount() int {
	return len(iu.dirtyNodes)
}

// ClearDirty 清空脏节点标记
func (iu *IncrementalUpdater) ClearDirty() {
	iu.dirtyNodes = make(map[uint64]bool)
}

// UpdateAll 标记所有节点为脏并执行更新
//
// 这相当于完全重建 HitMap，但在某些情况下可能是必要的。
func (iu *IncrementalUpdater) UpdateAll() {
	// 标记所有节点为脏
	iu.markAllDirty(iu.root)

	// 执行增量更新
	iu.IncrementalUpdate()
}

// markAllDirty 递归标记所有节点为脏
func (iu *IncrementalUpdater) markAllDirty(node layout.Node) {
	if node == nil {
		return
	}

	if node.ID() != "" {
		iu.MarkDirty(stringToNodeID(node.ID()))
	}

	for _, child := range node.Children() {
		iu.markAllDirty(child)
	}
}

// ============================================================================
// 优化辅助函数
// ============================================================================

// ShouldUseIncrementalUpdate 判断是否应该使用增量更新
//
// 如果脏节点比例小于阈值，使用增量更新更高效。
func ShouldUseIncrementalUpdate(totalNodes, dirtyNodes int) bool {
	if totalNodes == 0 {
		return false
	}

	dirtyRatio := float64(dirtyNodes) / float64(totalNodes)

	// 如果脏节点少于 30%，使用增量更新
	return dirtyRatio < 0.3
}

// GetDirtyRatio 获取脏节点比例
func (iu *IncrementalUpdater) GetDirtyRatio() float64 {
	if iu.root == nil {
		return 0
	}

	totalNodes := iu.countNodes(iu.root)
	if totalNodes == 0 {
		return 0
	}

	return float64(len(iu.dirtyNodes)) / float64(totalNodes)
}

// countNodes 计算节点总数
func (iu *IncrementalUpdater) countNodes(node layout.Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children() {
		count += iu.countNodes(child)
	}

	return count
}

// OptimizeRebuild 优化重建策略
//
// 根据脏节点数量和位置，选择最优的重建策略。
func (iu *IncrementalUpdater) OptimizeRebuild() {
	if iu.root == nil {
		return
	}

	totalNodes := iu.countNodes(iu.root)
	dirtyNodes := len(iu.dirtyNodes)

	// 如果脏节点很多，直接重建整个 HitMap
	if !ShouldUseIncrementalUpdate(totalNodes, dirtyNodes) {
		// 完全重建
		newHitMap := BuildHitMap(iu.root)
		iu.hitMap.entries = newHitMap.entries
		iu.ClearDirty()
		return
	}

	// 否则使用增量更新
	iu.IncrementalUpdate()
}

// ============================================================================
// 性能监控
// ============================================================================

// PerformanceStats 性能统计
type PerformanceStats struct {
	LastRebuildTime      time.Duration
	LastRebuildMethod    string // "full" or "incremental"
	TotalNodes          int
	DirtyNodes          int
	RebuildCount        int
	AvgRebuildTime      time.Duration
}

// GetPerformanceStats 获取性能统计
func (iu *IncrementalUpdater) GetPerformanceStats() *PerformanceStats {
	stats := &PerformanceStats{
		TotalNodes:   iu.countNodes(iu.root),
		DirtyNodes:   len(iu.dirtyNodes),
		RebuildCount: int(iu.version),
	}

	// TODO: 添加计时功能

	return stats
}
