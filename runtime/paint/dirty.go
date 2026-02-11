package paint

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/internal/log"
)

// ==============================================================================
// Dirty Region Tracking (V3)
// ==============================================================================
// 脏区域跟踪，用于优化渲染（只重绘变化的部分）

// DirtyTracker 脏区域跟踪器
type DirtyTracker struct {
	// 脏单元格集合
	cells map[cellRef]struct{}

	// 脏矩形列表
	rects []Rect

	// 全局脏标记
	allDirty bool

	// 上次渲染的快照
	prevBuffer *Buffer

	// 变更计数
	changedCells int

	// 复用的 visited 数组（用于 extractDirtyRegions，避免频繁分配）
	visited      [][]bool
	visitedWidth int
	visitedHeight int
}

// cellRef 单元格引用
type cellRef struct {
	x int
	y int
}

// DiffResult 表示两次渲染之间的差异结果
type DiffResult struct {
	DirtyRegions []Rect // 发生变化的区域
	HasChanges   bool   // 是否有任何变化
	ChangedCells int    // 变化的单元格数量
}

// NewDirtyTracker 创建脏区域跟踪器
func NewDirtyTracker() *DirtyTracker {
	return &DirtyTracker{
		cells:      make(map[cellRef]struct{}),
		rects:      make([]Rect, 0),
		allDirty:   false, // 初始状态为非全脏，由调用方决定何时全屏渲染
		prevBuffer: nil,
	}
}

// ==============================================================================
// 基础标记 API
// ==============================================================================

// MarkCell 标记单元格为脏
func (d *DirtyTracker) MarkCell(x, y int) {
	if d.allDirty {
		return
	}
	d.cells[cellRef{x: x, y: y}] = struct{}{}
}

// MarkRect 标记矩形区域为脏
func (d *DirtyTracker) MarkRect(rect Rect) {
	if d.allDirty {
		return
	}
	d.rects = append(d.rects, rect)
}

// MarkAll 标记全部为脏
func (d *DirtyTracker) MarkAll() {
	d.allDirty = true
	d.cells = make(map[cellRef]struct{})
	d.rects = make([]Rect, 0)
}

// IsAllDirty 检查是否全部为脏
func (d *DirtyTracker) IsAllDirty() bool {
	return d.allDirty
}

// GetDirtyRects 获取脏矩形列表（优化后的）
func (d *DirtyTracker) GetDirtyRects() []Rect {
	if d.allDirty {
		if d.prevBuffer != nil {
			return []Rect{{X: 0, Y: 0, Width: d.prevBuffer.Width, Height: d.prevBuffer.Height}}
		}
		return []Rect{{X: 0, Y: 0, Width: 80, Height: 24}} // 默认尺寸
	}
	return d.rects
}

// GetDirtyCells 获取脏单元格集合
func (d *DirtyTracker) GetDirtyCells() map[cellRef]struct{} {
	if d.allDirty {
		return nil
	}
	return d.cells
}

// Clear 清除脏标记
func (d *DirtyTracker) Clear() {
	d.allDirty = false
	d.cells = make(map[cellRef]struct{})
	d.rects = make([]Rect, 0)
	d.changedCells = 0
}

// IsDirtyCell 检查单元格是否为脏
func (d *DirtyTracker) IsDirtyCell(x, y int) bool {
	if d.allDirty {
		return true
	}
	_, ok := d.cells[cellRef{x: x, y: y}]
	if ok {
		return true
	}
	// 检查是否在脏矩形内
	for _, rect := range d.rects {
		if x >= rect.X && x < rect.X+rect.Width &&
			y >= rect.Y && y < rect.Y+rect.Height {
			return true
		}
	}
	return false
}

// ==============================================================================
// Diff 功能 - 主要入口
// ==============================================================================

// Diff 比较两个 Buffer 并返回差异结果
// 这是对外的主要 Diff 入口点
func (d *DirtyTracker) Diff(prev, curr *Buffer) DiffResult {
	result := DiffResult{
		DirtyRegions: []Rect{},
		HasChanges:   false,
			ChangedCells: 0,
	}

	// 边界情况处理
	if prev == nil && curr == nil {
		return result
	}

	if curr == nil {
		// 当前为空，清屏
		result.HasChanges = true
		if prev != nil {
			result.DirtyRegions = append(result.DirtyRegions, Rect{
				X: 0, Y: 0, Width: prev.Width, Height: prev.Height,
			})
		}
		return result
	}

	if prev == nil {
		// 上次为空，全屏绘制
		result.HasChanges = true
		result.DirtyRegions = append(result.DirtyRegions, Rect{
			X: 0, Y: 0, Width: curr.Width, Height: curr.Height,
		})
		// 清除 allDirty 标记
		d.allDirty = false
		return result
	}

	// 尺寸不同，全屏重绘
	if prev.Width != curr.Width || prev.Height != curr.Height {
		d.MarkAll()
		result.HasChanges = true
		result.DirtyRegions = append(result.DirtyRegions, Rect{
			X: 0, Y: 0, Width: curr.Width, Height: curr.Height,
		})
		return result
	}

	// 检查是否标记为全脏（ForceFullRender 调用 MarkAll）
	// 这会强制输出全屏，不管内容是否真的变化
	if d.allDirty {
		result.HasChanges = true
		result.ChangedCells = curr.Width * curr.Height
		result.DirtyRegions = []Rect{{X: 0, Y: 0, Width: curr.Width, Height: curr.Height}}
		// 清除 allDirty 标记，这样下次渲染会使用正常 diff
		d.allDirty = false
		d.cells = make(map[cellRef]struct{})
		d.rects = make([]Rect, 0)
		return result
	}

	// 使用脏网格进行高级 diff 分析
	dirtyGrid := d.compareBuffersWithGrid(prev, curr)
	result.DirtyRegions = d.extractDirtyRegions(dirtyGrid, curr.Width, curr.Height)
	result.ChangedCells = d.changedCells
	result.HasChanges = len(result.DirtyRegions) > 0

	// 保存当前 Buffer 作为下次比较的基准
	d.prevBuffer = prev

	return result
}

// CompareBuffers 比较两个 Buffer，更新内部脏状态
// 保留此方法以保持向后兼容
func (d *DirtyTracker) CompareBuffers(prev, curr *Buffer) {
	_ = d.Diff(prev, curr)
}

// compareBuffersWithGrid 使用网格比较两个 Buffer
// 使用 IsCellChanged 来确保与 renderLine 的逻辑一致
func (d *DirtyTracker) compareBuffersWithGrid(prev, curr *Buffer) *dirtyGrid {
	grid := newDirtyGrid(curr.Width, curr.Height)
	d.changedCells = 0

	// DEBUG: 调试光标闪烁问题
	debugRender := os.Getenv("TUI_RENDER_DEBUG") == "1"
	var firstChangedCell struct{ x, y int }

	for y := 0; y < curr.Height; y++ {
		for x := 0; x < curr.Width; x++ {
			prevCell := prev.Cells[y][x]
			currCell := curr.Cells[y][x]
			// 使用 IsCellChanged 而不是 cellsEqual，确保逻辑一致
			changed := IsCellChanged(currCell, prevCell)
			if changed {
				grid.Mark(x, y)
				d.changedCells++
				if debugRender && d.changedCells == 1 {
					firstChangedCell.x = x
					firstChangedCell.y = y
				}
			}
		}
	}

	if debugRender && d.changedCells > 0 {
		log.RenderLogger.Debug("[compareBuffersWithGrid] ChangedCells=%d, first=(%d,%d)\n",
			d.changedCells, firstChangedCell.x, firstChangedCell.y)
		// 输出第一个变化单元格的详细信息
		if curr.Height > firstChangedCell.y && curr.Width > firstChangedCell.x {
			c := curr.Cells[firstChangedCell.y][firstChangedCell.x]
			p := prev.Cells[firstChangedCell.y][firstChangedCell.x]
			log.RenderLogger.Debug("  curr: Cluster=%q, Width=%d, Style.Reverse=%v\n",
				c.Cluster, c.Width, c.Style.IsReverse())
			log.RenderLogger.Debug("  prev: Cluster=%q, Width=%d, Style.Reverse=%v\n",
				p.Cluster, p.Width, p.Style.IsReverse())
		}
	}

	return grid
}

// extractDirtyRegions 从脏网格中提取脏区域
func (d *DirtyTracker) extractDirtyRegions(grid *dirtyGrid, width, height int) []Rect {
	if grid == nil || width == 0 || height == 0 {
		return []Rect{}
	}

	// 使用或分配复用的 visited 数组
	if d.visited == nil || d.visitedWidth != width || d.visitedHeight != height {
		d.visited = make([][]bool, height)
		for i := range d.visited {
			d.visited[i] = make([]bool, width)
		}
		d.visitedWidth = width
		d.visitedHeight = height
	} else {
		// 清空 visited 数组（比重新分配更快）
		for y := 0; y < height; y++ {
			row := d.visited[y]
			for x := 0; x < width; x++ {
				row[x] = false
			}
		}
	}

	visited := d.visited
	var regions []Rect

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid.IsDirty(x, y) && !visited[y][x] {
				// 找到一个新的未访问脏单元格
				region := d.extractRegion(x, y, grid, visited)
				regions = append(regions, region)
			}
		}
	}

	// DEBUG: 调试区域提取
	if os.Getenv("TUI_RENDER_DEBUG") == "1" {
		log.RenderLogger.Debug("[extractDirtyRegions] found %d regions\n", len(regions))
		for i, r := range regions {
			log.RenderLogger.Debug("  Region[%d]: X=%d, Y=%d, W=%d, H=%d\n", i, r.X, r.Y, r.Width, r.Height)
		}
	}

	// 合并重叠或相邻的区域
	merged := d.mergeDirtyRegions(regions)

	if os.Getenv("TUI_RENDER_DEBUG") == "1" && len(regions) != len(merged) {
		log.RenderLogger.Debug("[extractDirtyRegions] merged from %d to %d regions\n", len(regions), len(merged))
	}

	return merged
}

// extractRegion 使用 flood fill 提取单个连续脏区域
func (d *DirtyTracker) extractRegion(startX, startY int, grid *dirtyGrid, visited [][]bool) Rect {
	minX, minY := startX, startY
	maxX, maxY := startX, startY

	// 使用栈进行 flood fill
	stack := []struct{ x, y int }{{startX, startY}}
	visited[startY][startX] = true

	for len(stack) > 0 {
		// 弹出
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		x, y := pos.x, pos.y

		// 更新边界
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}

		// 检查邻居（4方向）
		neighbors := []struct{ x, y int }{
			{x - 1, y}, // 左
			{x + 1, y}, // 右
			{x, y - 1}, // 上
			{x, y + 1}, // 下
		}

		for _, n := range neighbors {
			nx, ny := n.x, n.y

			// 检查边界
			if nx >= 0 && nx < grid.width && ny >= 0 && ny < grid.height {
				// 检查是否脏且未访问
				if grid.IsDirty(nx, ny) && !visited[ny][nx] {
					visited[ny][nx] = true
					stack = append(stack, struct{ x, y int }{nx, ny})
				}
			}
		}
	}

	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX + 1,
		Height: maxY - minY + 1,
	}
}

// mergeDirtyRegions 合并重叠或相邻的脏区域
func (d *DirtyTracker) mergeDirtyRegions(regions []Rect) []Rect {
	if len(regions) <= 1 {
		return regions
	}

	// 贪心合并：合并彼此靠近的区域
	merged := true
	for merged {
		merged = false
		for i := 0; i < len(regions); i++ {
			for j := i + 1; j < len(regions); j++ {
				if d.shouldMerge(regions[i], regions[j]) {
					// 合并区域 i 和 j
					regions[i] = d.mergeTwoRects(regions[i], regions[j])
					// 移除区域 j
					regions = append(regions[:j], regions[j+1:]...)
					merged = true
					break
				}
			}
			if merged {
				break
			}
		}
	}

	return regions
}

// shouldMerge 检查两个区域是否应该合并
// 合并条件：重叠或相邻（1个单元格内）
func (d *DirtyTracker) shouldMerge(a, b Rect) bool {
	// 检查重叠
	overlapX := a.X < b.X+b.Width && a.X+a.Width > b.X
	overlapY := a.Y < b.Y+b.Height && a.Y+a.Height > b.Y

	if overlapX && overlapY {
		return true
	}

	// 检查相邻（1个单元格内）
	adjacentX := a.X <= b.X+b.Width+1 && a.X+a.Width >= b.X-1
	adjacentY := a.Y <= b.Y+b.Height+1 && a.Y+a.Height >= b.Y-1

	return adjacentX && adjacentY
}

// mergeTwoRects 将两个矩形合并为边界框
func (d *DirtyTracker) mergeTwoRects(a, b Rect) Rect {
	minX := minInt(a.X, b.X)
	minY := minInt(a.Y, b.Y)
	maxX := maxInt(a.X+a.Width, b.X+b.Width)
	maxY := maxInt(a.Y+a.Height, b.Y+b.Height)

	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// ==============================================================================
// 脏网格 - 用于高级 diff 算法
// ==============================================================================

// dirtyGrid 跟踪脏单元格用于区域提取
type dirtyGrid struct {
	width  int
	height int
	cells  [][]bool
}

// newDirtyGrid 创建新的脏网格
func newDirtyGrid(width, height int) *dirtyGrid {
	cells := make([][]bool, height)
	for i := range cells {
		cells[i] = make([]bool, width)
	}

	return &dirtyGrid{
		width:  width,
		height: height,
		cells:  cells,
	}
}

// Mark 标记单元格为脏
func (dg *dirtyGrid) Mark(x, y int) {
	if x >= 0 && x < dg.width && y >= 0 && y < dg.height {
		dg.cells[y][x] = true
	}
}

// IsDirty 检查单元格是否为脏
func (dg *dirtyGrid) IsDirty(x, y int) bool {
	if x >= 0 && x < dg.width && y >= 0 && y < dg.height {
		return dg.cells[y][x]
	}
	return false
}

// ==============================================================================
// 辅助方法
// ==============================================================================

// MergeDirtyCells 合并脏单元格为矩形（简化版本）
// 优化：将相邻的脏单元格合并为更大的矩形以减少渲染调用
func (d *DirtyTracker) MergeDirtyCells() {
	if d.allDirty || len(d.cells) == 0 {
		return
	}

	// 简化实现：将每个单元格转换为1x1矩形
	for cell := range d.cells {
		d.rects = append(d.rects, Rect{
			X:      cell.x,
			Y:      cell.y,
			Width:  1,
			Height: 1,
		})
	}
	d.cells = make(map[cellRef]struct{})

	// 优化矩形
	d.OptimizeRects()
}

// OptimizeRects 优化矩形列表
// 合并重叠或相邻的矩形
func (d *DirtyTracker) OptimizeRects() {
	if len(d.rects) <= 1 {
		return
	}

	// 简化实现：去除完全包含的矩形
	optimized := make([]Rect, 0, len(d.rects))

	for i, r1 := range d.rects {
		contained := false
		for j, r2 := range d.rects {
			if i == j {
				continue
			}
			if rectContains(r2, r1) {
				contained = true
				break
			}
		}
		if !contained {
			optimized = append(optimized, r1)
		}
	}

	d.rects = optimized
}

// cellsEqual 比较两个单元格是否相等
func cellsEqual(a, b Cell) bool {
	return a.Cluster == b.Cluster && a.Style == b.Style && a.Width == b.Width
}

// rectContains 检查 r1 是否包含 r2
func rectContains(r1, r2 Rect) bool {
	return r2.X >= r1.X && r2.Y >= r1.Y &&
		r2.X+r2.Width <= r1.X+r1.Width &&
		r2.Y+r2.Height <= r1.Y+r1.Height
}

// SetPreviousBuffer 设置上次渲染的 Buffer
func (d *DirtyTracker) SetPreviousBuffer(buf *Buffer) {
	d.prevBuffer = buf
}

// GetPreviousBuffer 获取上次渲染的 Buffer
func (d *DirtyTracker) GetPreviousBuffer() *Buffer {
	return d.prevBuffer
}

// GetChangedCells 获取变化的单元格数量
func (d *DirtyTracker) GetChangedCells() int {
	return d.changedCells
}

// String 返回脏区域的字符串表示
func (d *DirtyTracker) String() string {
	if d.allDirty {
		return "DirtyTracker{ALL}"
	}
	return "DirtyTracker{" +
		"cells:" + fmt.Sprint(len(d.cells)) +
		",rects:" + fmt.Sprint(len(d.rects)) +
		",changed:" + fmt.Sprint(d.changedCells) +
		"}"
}

// ==============================================================================
// 实用函数
// ==============================================================================

// minInt 和 maxInt 已在 buffer.go 中定义，此处不再重复
