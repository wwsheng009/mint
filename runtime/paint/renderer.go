package paint

import (
	"bytes"
	"sort"
	"strings"
	"sync"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Renderer 2.0 - Line Diff Based Renderer
// =============================================================================
// 从 cell diff 升级到 line diff 架构
// 核心原则：只要一行任意 cell 改变 → 整行重新渲染
//
// 优势：
// - 解决折叠/展开导致的行位移问题
// - 解决 UI 结构变化导致的残留字符
// - 解决宽字符/continuation cell 问题
// - 自动清理行尾残留
// - O(n) diff 复杂度

// Renderer 双缓冲渲染器
//
// 实现了完整的双缓冲渲染管线：
// - front buffer: 当前屏幕状态
// - back buffer: 新一帧绘制目标
// - line diff: 只输出变化的行
// - run merging: 合并连续相同样式的片段
type Renderer struct {
	mu sync.Mutex

	// 双缓冲
	front *Buffer // 当前屏幕状态
	back  *Buffer // 新一帧绘制目标

	// 渲染状态
	styleState *StyleStateMachine
	cursorX    int
	cursorY    int

	// 输出缓冲
	output bytes.Buffer

	// 渲染模式
	useLineDiff bool // 是否使用行级 diff（默认 true）
	forceFull   bool // 强制下一帧全量渲染

	// 脏区域提示（来自 Fiber/PaintableBox 管线）
	// 注意：这是“提示”，不是强约束；渲染仍以 buffer diff 为准保证正确性。
	dirtyHints []Rect

	// 统计信息
	changedLines int // 最近一次渲染变化的行数
}

// NewRenderer 创建新的双缓冲渲染器
func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		front:       NewBuffer(width, height),
		back:        NewBuffer(width, height),
		styleState:  NewStyleStateMachine(),
		cursorX:     -1,
		cursorY:     -1,
		useLineDiff: true, // 默认使用行级 diff
		dirtyHints:  make([]Rect, 0),
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
// 1. 比较 front 和 back buffer，找出变化的行
// 2. 对每个变化行进行整行渲染（run merging 优化）
// 3. 生成优化的 ANSI 输出
// 4. 交换 front/back buffer
func (r *Renderer) Render() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.output.Reset()

	// 重置内部状态，确保每帧渲染起点确定
	r.ResetState()

	// 计算行 hash（用于快速比较）
	// 每次都需要重新计算，因为 buffer 内容可能已被修改
	r.back.Rehash()
	if r.front != nil {
		r.front.Rehash()
	}

	// 将脏区域提示转换为行内区间（用于 X/Y 双维度最小化重绘）。
	hintRanges := rectsToLineRanges(r.dirtyHints, r.back.Width, r.back.Height)

	// 默认使用行级 diff，确保正确性
	diff := LineDiff(r.front, r.back)

	var fullLines []int
	hasChanges := diff.HasChanges || len(hintRanges) > 0

	// useLineDiff=false 时，回退为保守策略：有变化则整屏行重绘
	if !r.useLineDiff {
		if diff.HasChanges || r.forceFull || len(hintRanges) > 0 {
			fullLines = allLines(r.back.Height)
			hasChanges = len(fullLines) > 0
		}
	} else if r.forceFull {
		// 强制全量渲染
		fullLines = allLines(r.back.Height)
		hasChanges = len(fullLines) > 0
	} else if diff.HasScroll {
		// 滚动优化：只重绘滚动后新增的尾部行（再合并提示行）。
		// 其余行由 ANSI scroll 指令完成位置迁移。
		fullLines = scrollTailLines(r.back.Height, diff.ScrollAmount)
		if len(fullLines) == 0 {
			// 防御回退：若尾部计算异常，退回常规 changed lines，保证正确性。
			fullLines = diff.ChangedLines
		}
		if len(fullLines) > 0 || len(hintRanges) > 0 {
			hasChanges = true
		}
	} else {
		// 常规行级 diff
		fullLines = diff.ChangedLines
		if len(fullLines) > 0 || len(hintRanges) > 0 {
			hasChanges = true
		}
	}

	partialLineRanges := subtractFullLines(hintRanges, fullLines, r.back.Height)
	r.changedLines = countRenderedLines(fullLines, partialLineRanges, r.back.Height)
	log.RenderLogger.Debug("[RENDER] HasChanges=%v, ChangedLines=%d, HasScroll=%v, UseLineDiff=%v, HintRects=%d, ForceFull=%v",
		hasChanges, r.changedLines, diff.HasScroll, r.useLineDiff, len(r.dirtyHints), r.forceFull)

	if !hasChanges {
		r.clearFrameHintsLocked()
		return "" // 无变化，不输出
	}

	// 处理滚动（如果检测到）
	if r.useLineDiff && !r.forceFull && diff.HasScroll {
		r.emitScroll(diff.ScrollAmount)
	}

	fullMarks := make([]bool, r.back.Height)
	for _, y := range fullLines {
		if y >= 0 && y < r.back.Height {
			fullMarks[y] = true
		}
	}

	// 渲染变化内容：优先整行，其次渲染提示区间
	for y := 0; y < r.back.Height; y++ {
		if fullMarks[y] {
			r.renderFullLine(y)
			continue
		}
		if ranges := partialLineRanges[y]; len(ranges) > 0 {
			r.renderLineRanges(y, ranges)
		}
	}

	// 重置样式
	r.output.WriteString("\x1b[0m")
	r.styleState.Reset()

	// 交换缓冲区
	r.swapBuffers()
	r.clearFrameHintsLocked()

	output := r.output.String()
	log.RenderLogger.Debug("[RENDER] output length=%d bytes", len(output))

	return output
}

// renderFullLine 渲染整行
// 这是 Renderer 2.0 的核心方法，替代原来的 renderLine
// 整行渲染解决了以下问题：
// - 行移位导致的残留
// - 跨行结构变化
// - 行尾残留
func (r *Renderer) renderFullLine(y int) {
	if y >= r.back.Height {
		return
	}
	if y >= len(r.back.Cells) {
		return
	}

	row := r.back.Cells[y]
	x := 0

	for x < len(row) {
		cell := row[x]

		// 跳过 continuation 单元格（由主单元格处理）
		if cell.IsContinuation {
			x++
			continue
		}

		// 计算单元格宽度
		width := cell.Width
		if width <= 0 {
			width = 1
		}

		// 获取文本
		text := cell.Cluster
		if text == "" || text == "\x00" {
			text = " "
		}

		// 尝试合并相邻同样式的单元格
		startX := x
		runStyle := cell.Style
		var runText bytes.Buffer
		totalWidth := 0

		// 收集连续同样式的单元格
		for x < len(row) {
			c := row[x]

			// continuation 由主单元格处理
			if c.IsContinuation {
				x++
				continue
			}

			// 样式不同，停止合并
			if c.Style != runStyle {
				break
			}

			// 添加文本
			if c.Cluster == "" || c.Cluster == "\x00" {
				runText.WriteString(" ")
			} else {
				runText.WriteString(c.Cluster)
			}

			cWidth := c.Width
			if cWidth <= 0 {
				cWidth = 1
			}
			totalWidth += cWidth
			x += cWidth
		}

		// 输出合并的 run
		if runText.Len() > 0 {
			r.emitRunWithWidth(startX, y, runStyle, runText.String(), totalWidth)
		}
	}

	// 清理行尾残留（关键修复：使用 ESC[K）
	r.output.WriteString("\x1b[K")
}

// renderLineRanges renders only specified x-ranges of a line.
// Unlike renderFullLine, this method does not emit ESC[K because it must not
// clear line tail outside the hinted dirty region.
func (r *Renderer) renderLineRanges(y int, ranges []lineRange) {
	if y < 0 || y >= r.back.Height || y >= len(r.back.Cells) || len(ranges) == 0 {
		return
	}

	row := r.back.Cells[y]
	rowWidth := len(row)

	for _, rg := range ranges {
		start, end := normalizeRangeForWideCells(row, rg.start, rg.end)
		if start < 0 {
			start = 0
		}
		if end > rowWidth {
			end = rowWidth
		}
		if start >= end {
			continue
		}

		x := start
		for x < end {
			cell := row[x]
			if cell.IsContinuation {
				x++
				continue
			}

			startX := x
			runStyle := cell.Style
			var runText bytes.Buffer
			totalWidth := 0

			for x < end {
				c := row[x]
				if c.IsContinuation {
					x++
					continue
				}
				if c.Style != runStyle {
					break
				}

				cWidth := c.Width
				if cWidth <= 0 {
					cWidth = 1
				}
				if x+cWidth > end {
					break
				}

				if c.Cluster == "" || c.Cluster == "\x00" {
					runText.WriteString(" ")
				} else {
					runText.WriteString(c.Cluster)
				}
				totalWidth += cWidth
				x += cWidth
			}

			if runText.Len() > 0 {
				r.emitRunWithWidth(startX, y, runStyle, runText.String(), totalWidth)
				continue
			}

			// 防御性推进，避免极端情况下死循环。
			x++
		}
	}
}

// emitRunWithWidth 输出一个渲染批次（带宽度参数，用于正确跟踪光标）
// 对于边框字符，使用 cell width 而非 runewidth.StringWidth，避免光标位置错误
func (r *Renderer) emitRunWithWidth(x, y int, runStyle style.Style, text string, textWidth int) {
	if text == "" {
		// 空文本也要更新内部光标状态，避免后续相对移动基准错误。
		r.cursorX = x + textWidth
		r.cursorY = y
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

	// 文本写入后，更新内部光标到 run 末尾
	r.cursorX = x + textWidth
	r.cursorY = y
}

// emitScroll 发送滚动 ANSI 命令
// 正数表示向上滚动，负数表示向下滚动
func (r *Renderer) emitScroll(amount int) {
	if amount == 0 {
		return
	}

	if amount > 0 {
		// 向上滚动: CSI n S
		r.output.WriteString("\x1b[")
		r.output.WriteString(itoa(amount))
		r.output.WriteString("S")
	} else {
		// 向下滚动: CSI n T
		r.output.WriteString("\x1b[")
		r.output.WriteString(itoa(-amount))
		r.output.WriteString("T")
	}

	log.RenderLogger.Debug("[emitScroll] amount=%d", amount)
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

// swapBuffers 更新前后缓冲区状态
// 关键设计：保持 back 指针不变，复制 front 内容到 back
// 这样用户持有的 back 指针始终有效
func (r *Renderer) swapBuffers() {
	// 确保 buffer 存在且大小一致
	if r.front == nil || r.back == nil {
		return
	}

	// 如果大小不一致，调整 front 缓冲区
	if r.front.Width != r.back.Width || r.front.Height != r.back.Height {
		r.front = NewBuffer(r.back.Width, r.back.Height)
	}

	// 将 back 的内容复制到 front，作为当前屏幕状态
	minHeight := minInt(r.front.Height, r.back.Height)
	if len(r.front.Cells) < minHeight {
		minHeight = len(r.front.Cells)
	}
	if len(r.back.Cells) < minHeight {
		minHeight = len(r.back.Cells)
	}

	for y := 0; y < minHeight; y++ {
		if y >= len(r.front.Cells) || y >= len(r.back.Cells) {
			break
		}
		srcRow := r.back.Cells[y]
		dstRow := r.front.Cells[y]
		copyLen := minInt(len(srcRow), len(dstRow))
		copy(dstRow[:copyLen], srcRow[:copyLen])
	}

	// 同步 line hash
	if r.back.LineHash != nil {
		if r.front.LineHash == nil || len(r.front.LineHash) != len(r.back.LineHash) {
			r.front.LineHash = make([]uint64, len(r.back.LineHash))
		}
		copy(r.front.LineHash, r.back.LineHash)
	}

	// 注意：不交换指针！
	// back 保持不变，用户可以继续在同一个 buffer 上绘制
	// front 更新为 back 的内容，用于下一次 diff
}

// Resize 调整渲染器大小
func (r *Renderer) Resize(width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.front = NewBuffer(width, height)
	r.back = NewBuffer(width, height)
	r.clearFrameHintsLocked()
}

// GetFrontBuffer 获取前缓冲区（用于测试）
func (r *Renderer) GetFrontBuffer() *Buffer {
	return r.front
}

// GetRenderSnapshot returns a plain-text snapshot of the current rendered
// content. It locks the renderer so callers never see a partially-reset buffer.
func (r *Renderer) GetRenderSnapshot() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	buf := r.back
	if buf == nil || buf.Height == 0 || buf.Width == 0 {
		buf = r.front
	}
	if buf == nil {
		return ""
	}

	// Check for non-trivial content; fall back to front if back is blank.
	hasContent := false
	for y := 0; y < buf.Height && !hasContent; y++ {
		if y >= len(buf.Cells) {
			break
		}
		for x := 0; x < buf.Width && !hasContent; x++ {
			if x >= len(buf.Cells[y]) {
				break
			}
			c := buf.Cells[y][x].Cluster
			if c != "" && c != " " {
				hasContent = true
			}
		}
	}
	if !hasContent && r.front != nil {
		buf = r.front
	}

	var sb strings.Builder
	for y := 0; y < buf.Height; y++ {
		if y >= len(buf.Cells) {
			break
		}
		for x := 0; x < buf.Width; x++ {
			if x >= len(buf.Cells[y]) {
				break
			}
			cell := buf.Cells[y][x]
			if cell.IsContinuation {
				continue
			}
			if cell.Cluster == "" {
				sb.WriteRune(' ')
			} else {
				sb.WriteString(cell.Cluster)
			}
		}
		if y < buf.Height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// MarkDirty 标记整个缓冲区为脏（兼容 API）
// 在 Renderer 2.0 中，此方法保持 no-op 以兼容历史调用点。
// 强制全量渲染请使用 ForceFullRender().
func (r *Renderer) MarkDirty() {
	// No-op (compatibility)
}

// MarkDirtyRect 标记矩形区域为脏（兼容 API）
// 在 Renderer 2.x 中用于提示局部重绘（支持 X/Y 范围）。
func (r *Renderer) MarkDirtyRect(rect Rect) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rect.Width <= 0 || rect.Height <= 0 {
		return
	}
	r.dirtyHints = append(r.dirtyHints, rect)
}

// ForceFullRender 强制下一帧全量渲染
// 通过重置 front buffer 来实现
func (r *Renderer) ForceFullRender() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.forceFull = true
	r.dirtyHints = r.dirtyHints[:0]

	// 清空 front buffer，强制下一帧全量渲染
	if r.front != nil {
		r.front = NewBuffer(r.front.Width, r.front.Height)
	}
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
		ChangedCells: r.changedLines, // 在 Renderer 2.0 中，这表示变化的行数
		OutputBytes:  r.output.Len(),
	}
}

// Reset 重置渲染器状态
func (r *Renderer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ResetState()
	r.output.Reset()
	r.clearFrameHintsLocked()
}

// UseLineDiff 设置是否使用行级 diff
func (r *Renderer) UseLineDiff(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.useLineDiff = enabled
}

// clearFrameHintsLocked clears one-frame dirty hints.
// Caller must hold r.mu.
func (r *Renderer) clearFrameHintsLocked() {
	r.forceFull = false
	if len(r.dirtyHints) > 0 {
		r.dirtyHints = r.dirtyHints[:0]
	}
}

func allLines(height int) []int {
	if height <= 0 {
		return nil
	}
	lines := make([]int, 0, height)
	for y := 0; y < height; y++ {
		lines = append(lines, y)
	}
	return lines
}

type lineRange struct {
	start int // inclusive
	end   int // exclusive
}

func rectsToLineRanges(rects []Rect, width, height int) map[int][]lineRange {
	if width <= 0 || height <= 0 || len(rects) == 0 {
		return nil
	}

	rangesByLine := make(map[int][]lineRange, len(rects))
	for _, rect := range rects {
		if rect.Width <= 0 || rect.Height <= 0 {
			continue
		}

		startY := rect.Y
		endY := rect.Y + rect.Height
		if startY < 0 {
			startY = 0
		}
		if endY > height {
			endY = height
		}
		if startY >= endY {
			continue
		}

		startX := rect.X
		endX := rect.X + rect.Width
		if startX < 0 {
			startX = 0
		}
		if endX > width {
			endX = width
		}
		if startX >= endX {
			continue
		}

		for y := startY; y < endY; y++ {
			rangesByLine[y] = append(rangesByLine[y], lineRange{start: startX, end: endX})
		}
	}

	for y, ranges := range rangesByLine {
		rangesByLine[y] = mergeLineRanges(ranges)
	}

	return rangesByLine
}

func mergeLineRanges(ranges []lineRange) []lineRange {
	if len(ranges) <= 1 {
		return ranges
	}

	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].start == ranges[j].start {
			return ranges[i].end < ranges[j].end
		}
		return ranges[i].start < ranges[j].start
	})

	merged := make([]lineRange, 0, len(ranges))
	current := ranges[0]
	for i := 1; i < len(ranges); i++ {
		next := ranges[i]
		if next.start <= current.end {
			if next.end > current.end {
				current.end = next.end
			}
			continue
		}
		merged = append(merged, current)
		current = next
	}
	merged = append(merged, current)
	return merged
}

func subtractFullLines(hints map[int][]lineRange, fullLines []int, height int) map[int][]lineRange {
	if len(hints) == 0 || height <= 0 {
		return nil
	}

	fullMarks := make([]bool, height)
	for _, y := range fullLines {
		if y >= 0 && y < height {
			fullMarks[y] = true
		}
	}

	out := make(map[int][]lineRange, len(hints))
	for y, ranges := range hints {
		if y < 0 || y >= height || fullMarks[y] || len(ranges) == 0 {
			continue
		}
		out[y] = ranges
	}
	return out
}

func countRenderedLines(fullLines []int, partial map[int][]lineRange, height int) int {
	if height <= 0 {
		return 0
	}
	marks := make([]bool, height)
	for _, y := range fullLines {
		if y >= 0 && y < height {
			marks[y] = true
		}
	}
	for y, ranges := range partial {
		if y >= 0 && y < height && len(ranges) > 0 {
			marks[y] = true
		}
	}
	count := 0
	for _, marked := range marks {
		if marked {
			count++
		}
	}
	return count
}

func normalizeRangeForWideCells(row []Cell, start, end int) (int, int) {
	if len(row) == 0 {
		return start, end
	}
	if start < 0 {
		start = 0
	}
	if end > len(row) {
		end = len(row)
	}
	if start >= end {
		return start, end
	}

	for start > 0 && start < len(row) && row[start].IsContinuation {
		start--
	}
	for end < len(row) && row[end].IsContinuation {
		end++
	}
	if end > len(row) {
		end = len(row)
	}
	return start, end
}

// scrollTailLines returns the minimal set of lines that must be repainted
// after applying terminal scroll commands.
// amount > 0: scroll up, repaint bottom "amount" lines
// amount < 0: scroll down, repaint top "-amount" lines
func scrollTailLines(height, amount int) []int {
	if height <= 0 || amount == 0 {
		return nil
	}

	if amount > 0 {
		if amount > height {
			amount = height
		}
		start := height - amount
		lines := make([]int, 0, amount)
		for y := start; y < height; y++ {
			lines = append(lines, y)
		}
		return lines
	}

	// amount < 0
	n := -amount
	if n > height {
		n = height
	}
	lines := make([]int, 0, n)
	for y := 0; y < n; y++ {
		lines = append(lines, y)
	}
	return lines
}
