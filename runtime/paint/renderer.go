package paint

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/wwsheng009/mint/runtime/style"
)

// Renderer 双缓冲渲染器
//
// 实现了完整的双缓冲渲染管线：
// - front buffer: 当前屏幕状态
// - back buffer: 新一帧绘制目标
// - diff: 只输出变化部分
// - run merging: 合并连续相同样式的片段
type Renderer struct {
	mu sync.Mutex

	// 双缓冲
	front *Buffer // 当前屏幕状态
	back  *Buffer // 新一帧绘制目标

	// 脏区域跟踪
	dirtyTracker *DirtyTracker

	// 渲染状态
	styleState *StyleStateMachine
	cursorX    int
	cursorY    int

	// 输出缓冲
	output bytes.Buffer
}

// NewRenderer 创建新的双缓冲渲染器
func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		front:        NewBuffer(width, height),
		back:         NewBuffer(width, height),
		dirtyTracker: NewDirtyTracker(),
		styleState:   NewStyleStateMachine(),
		cursorX:      -1,
		cursorY:      -1,
	}
}

// GetBackBuffer 获取用于绘制的后缓冲区
func (r *Renderer) GetBackBuffer() *Buffer {
	return r.back
}
// ResetState resets the internal state machine (cursor, style)
// This should be called before starting a new frame rendering
func (r *Renderer) ResetState() {
	r.styleState.Reset()
	r.cursorX = -1
	r.cursorY = -1
}

// Render 对比 front/back 并生成差异输出
//
// 这是核心渲染方法，执行以下步骤：
// 1. 比较 front 和 back buffer，找出变化区域
// 2. 对每个脏区域进行 run merging 渲染
// 3. 生成优化的 ANSI 输出
// 4. 交换 front/back buffer
func (r *Renderer) Render() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.output.Reset()

	// 重置内部状态，确保每帧渲染起点确定
	r.ResetState()

	// 执行 diff，找出变化区域
	diff := r.dirtyTracker.Diff(r.front, r.back)

	// DEBUG: 输出 diff 信息
	if os.Getenv("TUI_RENDER_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[RENDER] HasChanges=%v, ChangedCells=%d, Regions=%d\n",
			diff.HasChanges, diff.ChangedCells, len(diff.DirtyRegions))
	}

	if !diff.HasChanges {
		return "" // 无变化，不输出
	}

	// 渲染每个脏区域
	for _, region := range diff.DirtyRegions {
		r.renderRegion(region)
	}

	// 重置样式
	r.output.WriteString("\x1b[0m")

	// 交换缓冲区
	r.swapBuffers()

	return r.output.String()
}

// renderRegion 渲染单个脏区域
func (r *Renderer) renderRegion(region Rect) {
	if os.Getenv("TUI_RENDER_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[renderRegion] region={X:%d, Y:%d, W:%d, H:%d}, back.H=%d\n",
			region.X, region.Y, region.Width, region.Height, r.back.Height)
	}

	for y := region.Y; y < region.Y+region.Height; y++ {
		if y >= r.back.Height {
			if os.Getenv("TUI_RENDER_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[renderRegion] y=%d >= back.Height=%d, break\n", y, r.back.Height)
			}
			break
		}
		r.renderLine(y, region)
	}
}

// renderLine 渲染单行，使用 run merging 优化
func (r *Renderer) renderLine(y int, region Rect) {
	debugRender := os.Getenv("TUI_RENDER_DEBUG") == "1"

	if debugRender {
		fmt.Fprintf(os.Stderr, "[renderLine] y=%d, region.X=%d, region.W=%d, back.W=%d\n",
			y, region.X, region.Width, r.back.Width)
	}

	x := region.X
	// 确保 endX 不超过 back 和 front 缓冲区的宽度
	endX := minInt(region.X+region.Width, r.back.Width)
	if r.front != nil && r.front.Width < endX {
		endX = r.front.Width
	}

	// 确保 x 不超出范围
	if x >= endX || x < 0 {
		if debugRender {
			fmt.Fprintf(os.Stderr, "[renderLine] x=%d >= endX=%d or x<0, no render!\n", x, endX)
		}
		return
	}

	// 确保 y 在有效范围内
	if y >= len(r.back.Cells) || (r.front != nil && y >= len(r.front.Cells)) {
		if debugRender {
			fmt.Fprintf(os.Stderr, "[renderLine] y=%d out of bounds, no render!\n", y)
		}
		return
	}

	runCount := 0
	for x < endX {
		// 边界检查
		if x >= len(r.back.Cells[y]) || (r.front != nil && x >= len(r.front.Cells[y])) {
			if debugRender {
				fmt.Fprintf(os.Stderr, "[renderLine] x=%d out of row bounds, break\n", x)
			}
			break
		}

		cell := r.back.Cells[y][x]
		prevCell := r.front.Cells[y][x]

		// 跳过未变化的单元格
		if !IsCellChanged(cell, prevCell) {
			x++
			continue
		}

		if debugRender {
			fmt.Fprintf(os.Stderr, "[renderLine] changed at x=%d: cell.Cluster=%q, prev.Cluster=%q, IsContinuation=%v\n",
				x, cell.Cluster, prevCell.Cluster, cell.IsContinuation)
		}

		// 如果是延续单元格，跳过（由主单元格处理）
		if cell.IsContinuation {
			x++
			continue
		}

		// 跳过空 cluster（避免无限循环和无效输出）
		if cell.Cluster == "" || cell.Cluster == "\x00" {
			x++
			continue
		}

		// 开始一个 run
		startX := x
		runStyle := cell.Style
		var runText bytes.Buffer

		// 只收集当前单元格（因为 run merging 要求相邻单元格样式相同）
		runText.WriteString(cell.Cluster)
		// 确保 x 至少前进 1，避免 Width=0 导致无限循环
		width := cell.Width
		if width <= 0 {
			width = 1
		}
		x += width

		runCount++
		// 输出这个 run，传入实际的 cell width 用于光标跟踪
		r.emitRunWithWidth(startX, y, runStyle, runText.String(), width)
	}

	if debugRender {
		fmt.Fprintf(os.Stderr, "[renderLine] emitted %d runs\n", runCount)
	}
}

// emitRunWithWidth 输出一个渲染批次（带宽度参数，用于正确跟踪光标）
// 对于边框字符，使用 cell width 而非 runewidth.StringWidth，避免光标位置错误
func (r *Renderer) emitRunWithWidth(x, y int, runStyle style.Style, text string, textWidth int) {
	if text == "" {
		return
	}

	// 移动光标
	cursorCmd := r.moveCursorOptimized(x, y)
	if cursorCmd != "" {
		r.output.WriteString(cursorCmd)
	}

	// 设置样式（只输出变化部分）
	if r.styleState.NeedsUpdate(runStyle) {
		r.output.WriteString(r.styleState.Update(runStyle))
	}

	// 输出文本
	r.output.WriteString(text)

	// 更新光标位置 - 使用实际的 cell width 而非 runewidth.StringWidth
	// 这修复了边框字符被 runewidth 认为宽度 2 导致光标跳过的问题
	r.cursorX = x + textWidth
	r.cursorY = y
}

// emitRun 输出一个渲染批次
//
// 这是输出优化的核心：
// 1. 最小化光标移动
// 2. 只输出变化的样式
// 3. 批量输出文本
func (r *Renderer) emitRun(x, y int, runStyle style.Style, text string) {
	if text == "" {
		return
	}

	// 移动光标
	cursorCmd := r.moveCursorOptimized(x, y)
	if cursorCmd != "" {
		r.output.WriteString(cursorCmd)
	}

	// 设置样式（只输出变化部分）
	if r.styleState.NeedsUpdate(runStyle) {
		r.output.WriteString(r.styleState.Update(runStyle))
	}

	// 输出文本
	r.output.WriteString(text)

	// 更新光标位置 - 使用正确计算的文本宽度
	r.cursorX = x + r.getTextWidth(text)
	r.cursorY = y
}

// moveCursorOptimized 优化的光标移动
func (r *Renderer) moveCursorOptimized(x, y int) string {
	// 未知位置，使用绝对定位
	if r.cursorX < 0 || r.cursorY < 0 {
		r.cursorX, r.cursorY = x, y
		return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
	}

	// 相同位置，无需移动
	if r.cursorX == x && r.cursorY == y {
		return ""
	}

	// 同行小步右移
	if r.cursorY == y && x > r.cursorX {
		delta := x - r.cursorX
		if delta <= 5 {
			r.cursorX = x
			return "\x1b[" + itoa(delta) + "C"
		}
	}

	// 默认绝对定位
	r.cursorX, r.cursorY = x, y
	return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
}

// getTextWidth 计算文本的显示宽度
// 对于边框字符（U+2500-U+257F Unicode Box Drawing block），返回 1 而非 runewidth.StringWidth 的 2
// 这确保光标跟踪正确，不会跳过字符
func (r *Renderer) getTextWidth(text string) int {
	return StringWidth(text)
}

// swapBuffers 交换前后缓冲区
// 交换后，front 指向新渲染的内容（成为当前屏幕状态），back 指向旧内容
// 交换后，back 会被复制 front 的内容作为下一帧的基准，确保未绘制的区域不会产生误判的 diff
func (r *Renderer) swapBuffers() {
	r.front, r.back = r.back, r.front

	// 确保两个缓冲区大小一致
	if r.front == nil || r.back == nil {
		return
	}

	// 如果大小不一致，调整 back 缓冲区
	if r.back.Width != r.front.Width || r.back.Height != r.front.Height {
		r.back = NewBuffer(r.front.Width, r.front.Height)
	}

	// 将 front 的内容复制到 back，作为下一帧绘制的基准
	// 这样确保 diff 算法只检测真正变化的部分
	minHeight := r.front.Height
	if len(r.front.Cells) < minHeight {
		minHeight = len(r.front.Cells)
	}
	if len(r.back.Cells) < minHeight {
		minHeight = len(r.back.Cells)
	}

	for y := 0; y < minHeight; y++ {
		// 确保行存在且长度足够
		if y >= len(r.front.Cells) || y >= len(r.back.Cells) {
			break
		}
		srcRow := r.front.Cells[y]
		dstRow := r.back.Cells[y]
		copyLen := len(srcRow)
		if len(dstRow) < copyLen {
			copyLen = len(dstRow)
		}
		copy(dstRow[:copyLen], srcRow[:copyLen])
	}
}

// Resize 调整渲染器大小
func (r *Renderer) Resize(width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.front = NewBuffer(width, height)
	r.back = NewBuffer(width, height)
	r.dirtyTracker.MarkAll()
}

// GetFrontBuffer 获取前缓冲区（用于测试）
func (r *Renderer) GetFrontBuffer() *Buffer {
	return r.front
}

// MarkDirty 标记整个缓冲区为脏
func (r *Renderer) MarkDirty() {
	r.dirtyTracker.MarkAll()
}

// MarkDirtyRect 标记矩形区域为脏
func (r *Renderer) MarkDirtyRect(rect Rect) {
	r.dirtyTracker.MarkRect(rect)
}

// ForceFullRender 强制下一帧全量渲染
func (r *Renderer) ForceFullRender() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.dirtyTracker.MarkAll()
}

// GetStats 获取渲染统计信息
type RenderStats struct {
	HasChanges   bool
	ChangedCells int
	DirtyRegions int
	OutputBytes  int
}

// GetStats 获取最近一次渲染的统计信息
func (r *Renderer) GetStats() RenderStats {
	return RenderStats{
		ChangedCells: r.dirtyTracker.GetChangedCells(),
		OutputBytes:  r.output.Len(),
	}
}

// Reset 重置渲染器状态
func (r *Renderer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ResetState()
	r.output.Reset()
}
