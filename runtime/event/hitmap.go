package event

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
)

// =============================================================================
// HitMap - 统一的命中测试系统
// =============================================================================
// HitMap 提供了从屏幕坐标到布局节点的映射
// 消除了各组件手动实现 containsPoint() 的需求

// =============================================================================
// HitMap - 统一的命中测试系统
// =============================================================================
// HitMap 提供了从屏幕坐标到布局节点的映射
// 消除了各组件手动实现 containsPoint() 的需求

// StringToNodeID 将字符串 ID 转换为 uint64 NodeID
//
// 转换规则（按优先级）：
// 1. 如果输入是数字字符串（如 "123"），直接解析为 uint64
// 2. 否则使用 FNV-1a 哈希算法将字符串 ID 稳定映射到 uint64
//
// 优先解析数字字符串以确保与 Fiber.ActionTargetID 的一致性：
//   - Fiber 创建：ActionTargetID = fmt.Sprintf("%d", NodeID)
//   - Bridge 路由：targetID = StringToNodeID(ActionTargetID)
//   - 通过 ParseUint 确保往返一致性
//
// 导出的版本供外部包使用
func StringToNodeID(id string) uint64 {
	if id == "" {
		return 0
	}

	// 优先尝试解析为数字（适用于 Fiber.ActionTargetID）
	if nodeID, err := strconv.ParseUint(id, 10, 64); err == nil {
		return nodeID
	}

	// 如果不是数字，使用 FNV-1a 哈希（适用于 layout.Node.ID() 等场景）
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	var h uint64 = offset64
	for i := 0; i < len(id); i++ {
		h ^= uint64(id[i])
		h *= prime64
	}
	return h
}

// stringToNodeID 将字符串 ID 转换为 uint64 NodeID
// 使用 FNV-1a 哈希算法确保字符串 ID 能稳定映射到 uint64
// 内部使用的版本（调用导出版本）
func stringToNodeID(id string) uint64 {
	return StringToNodeID(id)
}

// HitMapEntry 命中条目
// 表示布局树中的一个节点及其边界信息
type HitMapEntry struct {
	// NodeID 节点唯一标识（新架构）
	NodeID uint64

	// Node 布局节点引用
	Node layout.Node

	// Bounds 节点的边界矩形（屏幕绝对坐标）
	Bounds layout.Rect

	// LocalXY 屏幕坐标到局部坐标的转换函数
	LocalXY func(screenX, screenY int) (localX, localY int)

	// ZOrder 渲染层级（用于分层渲染和命中测试）
	ZOrder int

	// TargetFiber Fiber 引用 (Fiber-first Action Architecture)
	// 用于通过 ActionBridge 路由事件到组件
	TargetFiber interface {
		GetActionTargetID() string
	}
}

// String 返回调试字符串
func (e *HitMapEntry) String() string {
	return fmt.Sprintf("HitMapEntry{ID:%d, Bounds:%v, Z:%d}",
		e.NodeID, e.Bounds, e.ZOrder)
}

// HitMap 命中映射表
// 从布局树构建，用于快速命中测试
type HitMap struct {
	// entries 所有节点的命中条目
	entries []HitMapEntry

	// root 布局树的根节点
	root layout.Node

	// buildTime 构建时间戳
	buildTime time.Time

	zSorted bool
	lastZ   int
}

// BuildHitMap 从布局树构建 HitMap
// 这是 HitMap 的主要入口点，在每次布局计算后调用
//
// 参数：
//
//	root - 布局树的根节点
//
// 返回：
//
//	*HitMap - 构建好的命中映射表
//
// 示例：
//
//	hitMap := event.BuildHitMap(layoutRoot)
//	entry := hitMap.HitTest(x, y)
//	if entry != nil {
//	    fmt.Printf("Hit node: %s\n", entry.NodeID)
//	}
func BuildHitMap(root layout.Node) *HitMap {
	if root == nil {
		return &HitMap{
			entries:   make([]HitMapEntry, 0),
			root:      root,
			buildTime: time.Now(),
		}
	}

	hm := &HitMap{
		root:      root,
		entries:   make([]HitMapEntry, 0, countHitMapEntries(root)),
		buildTime: time.Now(),
		zSorted:   true,
	}

	// 递归遍历布局树，构建命中条目
	hm.walkAndBuild(root, 0)

	// 按 Z-order 排序（确保上层节点优先）
	if !hm.zSorted {
		hm.sortByZOrder()
	}

	return hm
}

func countHitMapEntries(node layout.Node) int {
	if node == nil {
		return 0
	}

	count := 0
	width, height := node.GetSize()
	if width > 0 && height > 0 {
		count++
	}

	for _, child := range node.Children() {
		count += countHitMapEntries(child)
	}

	return count
}

// walkAndBuild 递归遍历布局树并构建命中条目
//
// 参数：
//
//	node - 当前节点
//	zOrder - 当前节点的 Z 轴层级
func (hm *HitMap) walkAndBuild(node layout.Node, zOrder int) {
	if node == nil {
		return
	}

	// 获取节点的位置和尺寸
	x, y := node.GetPosition()
	width, height := node.GetSize()

	// 跳过无效节点（未布局或不可见）
	if width <= 0 || height <= 0 {
		// 继续处理子节点（可能有效）
		children := node.Children()
		for _, child := range children {
			hm.walkAndBuild(child, zOrder+1)
		}
		return
	}

	// 创建命中条目
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

	if len(hm.entries) > 0 && zOrder < hm.lastZ {
		hm.zSorted = false
	}
	hm.lastZ = zOrder
	hm.entries = append(hm.entries, entry)

	// 递归处理子节点（子节点的 Z-order 更高）
	children := node.Children()
	for _, child := range children {
		hm.walkAndBuild(child, zOrder+1)
	}
}

// sortByZOrder 按 Z-order 排序条目
// 确保 Z-order 较大的节点（上层）在数组末尾
// 这样 HitTest 从后向前遍历时，上层节点优先
func (hm *HitMap) sortByZOrder() {
	sort.Slice(hm.entries, func(i, j int) bool {
		return hm.entries[i].ZOrder < hm.entries[j].ZOrder
	})
}

// HitTest 在给定屏幕坐标执行命中测试
// 返回最上层（Z-order 最大）包含该点的节点
//
// 参数：
//
//	x, y - 屏幕坐标
//
// 返回：
//
//	*HitMapEntry - 命中的节点条目，如果未命中返回 nil
//
// 示例：
//
//	entry := hitMap.HitTest(100, 50)
//	if entry != nil {
//	    localX, localY := entry.LocalXY(100, 50)
//	    fmt.Printf("Hit %s at local (%d, %d)\n", entry.NodeID, localX, localY)
//	}
func (hm *HitMap) HitTest(x, y int) *HitMapEntry {
	// 从后向前遍历（Z-index 降序，上层优先）
	for i := len(hm.entries) - 1; i >= 0; i-- {
		entry := &hm.entries[i]
		if entry.Bounds.Contains(x, y) {
			return entry
		}
	}
	return nil
}

// FindByID 按节点 ID 查找命中条目
// 主要用于测试和调试
//
// 参数：
//
//	id - 节点 ID
//
// 返回：
//
//	*HitMapEntry - 找到的条目，如果未找到返回 nil
func (hm *HitMap) FindByID(id uint64) *HitMapEntry {
	for i := range hm.entries {
		if hm.entries[i].NodeID == id {
			return &hm.entries[i]
		}
	}
	return nil
}

// FindAllAt 在给定坐标点查找所有包含该点的节点（包括被遮挡的）
// 返回按 Z-order 升序排列（从下到上）
//
// 参数：
//
//	x, y - 屏幕坐标
//
// 返回：
//
//	[]*HitMapEntry - 所有包含该点的节点条目
func (hm *HitMap) FindAllAt(x, y int) []*HitMapEntry {
	var results []*HitMapEntry

	for i := range hm.entries {
		entry := &hm.entries[i]
		if entry.Bounds.Contains(x, y) {
			results = append(results, entry)
		}
	}

	return results
}

// Size 返回 HitMap 中的节点数量
func (hm *HitMap) Size() int {
	return len(hm.entries)
}

// IsEmpty 检查 HitMap 是否为空
func (hm *HitMap) IsEmpty() bool {
	return len(hm.entries) == 0
}

// GetBuildTime 返回 HitMap 的构建时间
func (hm *HitMap) GetBuildTime() time.Time {
	return hm.buildTime
}

// GetRoot 返回 HitMap 的根节点
// 用于构建 runtime.LayoutNode 树以支持 ActionRouter 的 Capture/Bubble 阶段
func (hm *HitMap) GetRoot() layout.Node {
	return hm.root
}

// AllEntries returns all HitMap entries
// This is used for merging HitMaps from multiple layers
func (hm *HitMap) AllEntries() []HitMapEntry {
	return hm.entries
}

// NewHitMap creates a new empty HitMap
func NewHitMap() *HitMap {
	return &HitMap{
		entries:   make([]HitMapEntry, 0),
		buildTime: time.Now(),
	}
}

// HitMapEntryInternal is an internal representation for building HitMap from external sources
// This allows external packages to build HitMap without accessing unexported fields
type HitMapEntryInternal struct {
	NodeID      uint64 // ✨ NodeID for stable runtime identity
	Node        layout.Node
	Bounds      layout.Rect
	LocalXY     func(screenX, screenY int) (int, int)
	ZOrder      int
	TargetFiber interface {
		GetActionTargetID() string
	} // Fiber-first Action Architecture
}

// BuildHitMapFromEntries builds a HitMap from a list of internal entries
// This is used by Engine.buildHitMapFromComputedBoxes to create HitMap from ComputedBox tree
func BuildHitMapFromEntries(entries []HitMapEntryInternal) *HitMap {
	hmEntries := make([]HitMapEntry, len(entries))
	for i, e := range entries {
		hmEntries[i] = HitMapEntry{
			NodeID:      e.NodeID,
			Node:        e.Node,
			Bounds:      e.Bounds,
			LocalXY:     e.LocalXY,
			ZOrder:      e.ZOrder,
			TargetFiber: e.TargetFiber,
		}
	}

	hm := &HitMap{
		entries:   hmEntries,
		buildTime: time.Now(),
	}

	// Sort by Z-order descending (higher layers first) for correct HitTest behavior
	hm.sortByZOrder()

	return hm
}

// Dump 调试输出
// 返回 HitMap 的可读字符串表示，用于调试
//
// 输出格式：
//
//	=== HitMap ===
//	[node1] {X:0, Y:0, W:100, H:50} (Z:0)
//	[node2] {X:10, Y:10, W:80, H:30} (Z:1)
//	...
func (hm *HitMap) Dump() string {
	var buf strings.Builder
	buf.WriteString("=== HitMap ===\n")
	buf.WriteString(fmt.Sprintf("Built at: %s\n", hm.buildTime.Format("15:04:05.000")))
	buf.WriteString(fmt.Sprintf("Entries: %d\n\n", len(hm.entries)))

	for _, entry := range hm.entries {
		buf.WriteString(fmt.Sprintf("[%d] %v (Z:%d)\n",
			entry.NodeID, entry.Bounds, entry.ZOrder))
	}

	return buf.String()
}

// DetailedHitTestResult 详细的命中测试结果
// 提供更丰富的命中信息，包括坐标转换
type DetailedHitTestResult struct {
	// Entry 命中的节点条目
	Entry *HitMapEntry

	// ScreenX, ScreenY 屏幕坐标
	ScreenX, ScreenY int

	// LocalX, LocalY 相对于节点的局部坐标
	LocalX, LocalY int

	// Found 是否命中
	Found bool
}

// HitTestDetailed 在给定屏幕坐标执行详细的命中测试
// 返回包含坐标转换信息的完整结果
//
// 参数：
//
//	x, y - 屏幕坐标
//
// 返回：
//
//	*DetailedHitTestResult - 详细的命中测试结果
func (hm *HitMap) HitTestDetailed(x, y int) *DetailedHitTestResult {
	entry := hm.HitTest(x, y)

	if entry == nil {
		return &DetailedHitTestResult{
			ScreenX: x,
			ScreenY: y,
			Found:   false,
		}
	}

	localX, localY := entry.LocalXY(x, y)

	return &DetailedHitTestResult{
		Entry:   entry,
		ScreenX: x,
		ScreenY: y,
		LocalX:  localX,
		LocalY:  localY,
		Found:   true,
	}
}
