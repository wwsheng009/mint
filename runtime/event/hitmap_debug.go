package event

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
)

// DebugHitMap 提供 HitMap 调试和可视化功能
type DebugHitMap struct {
	hitMap *HitMap
}

// NewDebugHitMap 创建 HitMap 调试器
func NewDebugHitMap(hitMap *HitMap) *DebugHitMap {
	return &DebugHitMap{
		hitMap: hitMap,
	}
}

// Dump 将 HitMap 内容转储为字符串
//
// 返回格式化的 HitMap 信息，包括所有节点的边界、ID 和 Z-order。
// 用于调试和日志记录。
func (d *DebugHitMap) Dump() string {
	if d.hitMap == nil {
		return "HitMap: <nil>"
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("HitMap Dump (%d entries):\n", len(d.hitMap.entries)))
	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n")

	for i, entry := range d.hitMap.entries {
		builder.WriteString(fmt.Sprintf("[%d] ID: %s\n", i, entry.NodeID))
		builder.WriteString(fmt.Sprintf("     Bounds: (%d, %d) - (%d, %d)\n",
			entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.X+entry.Bounds.Width,
			entry.Bounds.Y+entry.Bounds.Height))
		builder.WriteString(fmt.Sprintf("     Size: %dx%d\n", entry.Bounds.Width, entry.Bounds.Height))
		builder.WriteString(fmt.Sprintf("     Z-Order: %d\n", entry.ZOrder))
		builder.WriteString("\n")
	}

	return builder.String()
}

// Visualize 创建 HitMap 的可视化字符串
//
// 返回一个 ASCII 艺术风格的 HitMap 可视化，
// 显示所有组件的边界和位置关系。
func (d *DebugHitMap) Visualize() string {
	if d.hitMap == nil || len(d.hitMap.entries) == 0 {
		return "HitMap: <empty>"
	}

	// 计算总边界
	minX, minY, maxX, maxY := d.calculateBounds()
	width := maxX - minX
	height := maxY - minY

	// 创建网格
	grid := make([][]string, height)
	for y := range grid {
		grid[y] = make([]string, width)
		for x := range width {
			grid[y][x] = " "
		}
	}

	// 在网格上绘制 HitMapEntry
	for _, entry := range d.hitMap.entries {
		// 使用 ID 的首字符作为标记
		marker := "?"
		if len(entry.NodeID) > 0 {
			marker = string(entry.NodeID[0])
		}

		// 绘制矩形边界
		for y := entry.Bounds.Y; y < entry.Bounds.Y+entry.Bounds.Height && y < maxY; y++ {
			for x := entry.Bounds.X; x < entry.Bounds.X+entry.Bounds.Width && x < maxX; x++ {
				if y >= 0 && y < height && x >= 0 && x < width {
					// 边界使用 #
					if x == entry.Bounds.X || x == entry.Bounds.X+entry.Bounds.Width-1 ||
						y == entry.Bounds.Y || y == entry.Bounds.Y+entry.Bounds.Height-1 {
						grid[y][x] = "#"
					} else {
						grid[y][x] = marker
					}
				}
			}
		}
	}

	// 转换为字符串
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("HitMap Visualization (%dx%d):\n", width, height))
	builder.WriteString(strings.Repeat("-", width))
	builder.WriteString("\n")

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			builder.WriteString(grid[y][x])
		}
		builder.WriteString("\n")
	}

	builder.WriteString(strings.Repeat("-", width))
	builder.WriteString("\n")

	// 添加图例
	builder.WriteString("\nLegend:\n")
	builder.WriteString("  # - Component boundary\n")
	builder.WriteString("  ? - Unknown component\n")
	seen := make(map[rune]bool)
	for _, entry := range d.hitMap.entries {
		if len(entry.NodeID) > 0 {
			marker := rune(entry.NodeID[0])
			if !seen[marker] {
				builder.WriteString(fmt.Sprintf("  %c - %s\n", marker, entry.NodeID))
				seen[marker] = true
			}
		}
	}

	return builder.String()
}

// calculateBounds 计算 HitMap 的总边界
func (d *DebugHitMap) calculateBounds() (minX, minY, maxX, maxY int) {
	if len(d.hitMap.entries) == 0 {
		return 0, 0, 0, 0
	}

	minX = d.hitMap.entries[0].Bounds.X
	minY = d.hitMap.entries[0].Bounds.Y
	maxX = d.hitMap.entries[0].Bounds.X + d.hitMap.entries[0].Bounds.Width
	maxY = d.hitMap.entries[0].Bounds.Y + d.hitMap.entries[0].Bounds.Height

	for _, entry := range d.hitMap.entries {
		if entry.Bounds.X < minX {
			minX = entry.Bounds.X
		}
		if entry.Bounds.Y < minY {
			minY = entry.Bounds.Y
		}
		if entry.Bounds.X+entry.Bounds.Width > maxX {
			maxX = entry.Bounds.X + entry.Bounds.Width
		}
		if entry.Bounds.Y+entry.Bounds.Height > maxY {
			maxY = entry.Bounds.Y + entry.Bounds.Height
		}
	}

	return minX, minY, maxX, maxY
}

// EnableDebugOutput 启用 HitMap 调试输出
//
// 检查 HitMapLogger 状态，启用调试输出。
func (d *DebugHitMap) EnableDebugOutput() {
	if log.HitMapLogger.Enabled() {
		// 打印 HitMap 转储
		fmt.Fprint(os.Stderr, d.Dump())
		fmt.Fprint(os.Stderr, d.Visualize())
	}
}

// SaveToFile 将 HitMap 转储保存到文件
func (d *DebugHitMap) SaveToFile(filename string) error {
	return os.WriteFile(filename, []byte(d.Dump()), 0644)
}

// GetStats 返回 HitMap 的统计信息
func (d *DebugHitMap) GetStats() map[string]interface{} {
	if d.hitMap == nil {
		return map[string]interface{}{
			"count": 0,
		}
	}

	stats := map[string]interface{}{
		"count": len(d.hitMap.entries),
	}

	// 计算总边界
	minX, minY, maxX, maxY := d.calculateBounds()
	stats["bounds"] = map[string]int{
		"minX":   minX,
		"minY":   minY,
		"maxX":   maxX,
		"maxY":   maxY,
		"width":  maxX - minX,
		"height": maxY - minY,
	}

	// 统计 Z-order
	zOrderCounts := make(map[int]int)
	for _, entry := range d.hitMap.entries {
		zOrderCounts[entry.ZOrder]++
	}
	stats["z_order_counts"] = zOrderCounts

	return stats
}

// Validate 验证 HitMap 的有效性
//
// 检查 HitMap 是否有问题，例如：
// - 重叠的组件
// - 超出边界的组件
// - 零大小的组件
func (d *DebugHitMap) Validate() []string {
	issues := make([]string, 0)

	if d.hitMap == nil {
		return append(issues, "HitMap is nil")
	}

	for i, entry := range d.hitMap.entries {
		// 检查零大小
		if entry.Bounds.Width <= 0 || entry.Bounds.Height <= 0 {
			issues = append(issues, fmt.Sprintf("[%d] %s has zero or negative size: %dx%d",
				i, entry.NodeID, entry.Bounds.Width, entry.Bounds.Height))
		}

		// 检查负坐标
		if entry.Bounds.X < 0 || entry.Bounds.Y < 0 {
			issues = append(issues, fmt.Sprintf("[%d] %s has negative coordinates: (%d, %d)",
				i, entry.NodeID, entry.Bounds.X, entry.Bounds.Y))
		}

		// 检查与其他组件的重叠（相同 Z-order）
		for j := i + 1; j < len(d.hitMap.entries); j++ {
			other := d.hitMap.entries[j]
			if entry.ZOrder == other.ZOrder {
				if d.entriesOverlap(entry, other) {
					issues = append(issues, fmt.Sprintf("[%d] %s overlaps with [%d] %s at Z-order %d",
						i, entry.NodeID, j, other.NodeID, entry.ZOrder))
				}
			}
		}
	}

	return issues
}

// entriesOverlap 检查两个条目是否重叠
func (d *DebugHitMap) entriesOverlap(a, b HitMapEntry) bool {
	// 检查 X 轴重叠
	xOverlap := a.Bounds.X < b.Bounds.X+b.Bounds.Width && a.Bounds.X+a.Bounds.Width > b.Bounds.X

	// 检查 Y 轴重叠
	yOverlap := a.Bounds.Y < b.Bounds.Y+b.Bounds.Height && a.Bounds.Y+a.Bounds.Height > b.Bounds.Y

	return xOverlap && yOverlap
}
