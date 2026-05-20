package paint

import (
	"bytes"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/wwsheng009/mint/runtime/style"
)

// Buffer represents a grid of cells that components paint into.
// It acts as the "canvas" for the TUI rendering engine.
type Buffer struct {
	// Width and Height of the buffer
	Width  int
	Height int

	// Cells stores the grid content.
	// Access via GetCell/SetCell.
	Cells [][]Cell

	// LineHash stores hash values for each line for O(1) line comparison.
	// Used by the line-diff renderer for efficient change detection.
	LineHash []uint64
}

// NewBuffer creates a new buffer with the specified dimensions.
func NewBuffer(width, height int) *Buffer {
	b := &Buffer{
		Width:  width,
		Height: height,
		Cells:  make([][]Cell, height),
	}

	for y := 0; y < height; y++ {
		b.Cells[y] = make([]Cell, width)
		// 初始化为空格（与 Reset 保持一致）
		// IMPORTANT: 必须设置 Width=1，否则 Renderer 会跳过这些单元格
		// 导致未渲染区域在终端上显示为"白色格子"
		for x := 0; x < width; x++ {
			b.Cells[y][x] = Cell{Cluster: " ", Width: 1}
		}
	}

	return b
}

// SetCell sets the character and style at the given coordinates.
// It handles boundary checks safely and marks continuation cells for wide characters.
// This is a convenience method that converts a rune to a string cluster.
// For complex grapheme clusters, use SetString instead.
func (b *Buffer) SetCell(x, y int, char rune, s style.Style) {
	b.setCluster(x, y, string(char), getRuneWidth(char), s)
}

// getRuneWidth returns the display width for a rune.
// Border drawing characters are treated as width 1 to avoid conflicts.
func getRuneWidth(char rune) int {
	// Check if this is a border drawing character
	// These characters should be treated as single-width for TUI borders
	switch char {
	case '┌', '┐', '└', '┘', // Corners
		'─', '│', // Lines (including vertical for TreeView)
		'╔', '╗', '╚', '╝', // Double corners
		'═', '║', // Double lines
		'╭', '╮', '╰', '╯', // Rounded corners
		'+', '|', // ASCII style
		'├', '┤', '┬', '┴', // Tree connectors (important for TreeView)
		'┼': // Cross connector
		return 1
	default:
		return RuneWidth(char)
	}
}

// getClusterWidth returns the display width for a grapheme cluster.
// Border drawing characters are treated as width 1 to avoid conflicts.
func getClusterWidth(cluster string) int {
	return StringWidth(cluster)
}

// setCluster sets a grapheme cluster at the given coordinates.
// This is the low-level method that all writing operations should use.
func (b *Buffer) setCluster(x, y int, cluster string, width int, s style.Style) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}

	// 宽字符需要足够空间：位置 x 和 x+1 都必须在 buffer 范围内
	if width == 2 && x+1 >= b.Width {
		// 没有足够空间写入宽字符，不写入
		return
	}

	// 清除当前位置及其关联的宽字符单元格
	b.clearCellAt(x, y)

	// 设置当前单元格
	b.Cells[y][x] = Cell{
		Cluster:        cluster,
		Style:          s,
		Width:          width,
		IsContinuation: false,
	}

	// 对于宽字符，标记下一个单元格为延续
	if width == 2 {
		b.Cells[y][x+1] = Cell{
			IsContinuation: true,
		}
	}
}

// SetString writes a string starting at (x, y) with the given style.
// This method properly handles grapheme clusters (emoji ZWJ sequences,
// combining characters, flag emojis, etc.) using the uniseg library.
func (b *Buffer) SetString(x, y int, text string, s style.Style) {
	b.writeString(x, y, text, s, 0)
}

// SetStringAligned writes a string and pads the remaining cells in the row up to maxWidth.
// This implements row-level full coverage rendering (TUI_BUFFER_FIX2.md) to prevent
// leftover characters from previous renders (e.g., wide emoji continuation cells).
func (b *Buffer) SetStringAligned(x, y int, text string, s style.Style, maxWidth int) {
	b.writeString(x, y, text, s, maxWidth)
}

// writeString is the internal implementation that optionally pads the row.
// maxWidth: if > 0, pad from end of text to maxWidth with spaces
func (b *Buffer) writeString(x, y int, text string, s style.Style, maxWidth int) {
	if y < 0 || y >= b.Height {
		return
	}

	// 统一做终端安全过滤，避免 VS16/ZWJ/组合符导致的 cell 错位。
	text = SanitizeForTerminal(text)

	col := x
	g := uniseg.NewGraphemes(text)

	for g.Next() {
		cluster := g.Str()                // 完整字形簇
		width := getClusterWidth(cluster) // 使用正确的宽度计算（边框字符为宽度1）
		if width <= 0 {
			continue
		}

		// maxWidth 为绝对终止列（exclusive）。
		// 例如 x=10, maxWidth=17 时，只允许写入 [10,16]。
		if maxWidth > 0 {
			if col >= maxWidth {
				break
			}
			if col+width > maxWidth {
				break
			}
		}

		// 边界检查
		if col >= b.Width {
			break
		}
		if width == 2 && col+1 >= b.Width {
			break
		}

		b.setCluster(col, y, cluster, width, s)
		col += width
	}

	// Pad remaining cells to maxWidth (TUI_BUFFER_FIX2.md - border collapse fix)
	if maxWidth > 0 && col < maxWidth {
		endX := col + maxWidth - col
		if endX > b.Width {
			endX = b.Width
		}
		for i := col; i < endX; i++ {
			b.setCluster(i, y, " ", 1, s)
		}
	}
}

// Fill fills a rectangular area with a character and style.
func (b *Buffer) Fill(rect Rect, char rune, s style.Style) {
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			b.SetCell(x, y, char, s)
		}
	}
}

// GetContent returns the cell at the given position.
// This provides compatibility with runtime.CellBuffer interface.
func (b *Buffer) GetContent(x, y int) Cell {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return Cell{}
	}
	return b.Cells[y][x]
}

// SetContent sets a cell at the given position with Z-Index checking.
// If the new Z-Index is greater than or equal to the existing cell's Z-Index,
// the cell is overwritten. Otherwise, the operation is ignored.
// This provides compatibility with runtime.CellBuffer interface.
func (b *Buffer) SetContent(x, y, z int, char rune, s style.Style, nodeID string) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}

	// Check Z-Index - only overwrite if new Z is >= existing Z
	if z < b.Cells[y][x].ZIndex {
		return
	}

	b.Cells[y][x] = Cell{
		Cluster: string(char),
		Style:   s,
		ZIndex:  z,
		NodeID:  nodeID,
	}
}

// SetContentDirect sets a cell at the given position without Z-Index checking.
// This directly overwrites the cell regardless of Z-Index.
func (b *Buffer) SetContentDirect(x, y int, char rune, s style.Style, zIndex int) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}

	b.Cells[y][x] = Cell{
		Cluster: string(char),
		Style:   s,
		ZIndex:  zIndex,
	}
}

// ==============================================================================
// Wide Character Helper Functions
// ==============================================================================

// IsCellChanged 比较两个单元格是否不同，正确处理宽字符。
//
// 变化检测规则：
// - continuation → continuation: ❌ 不刷新（由主单元格处理）
// - continuation → non-continuation: ✅ 需要刷新（prevCell 需要被擦除）
// - head → continuation: ✅ 需要刷新（宽字符被覆盖）
// - continuation → head: ✅ 需要刷新（宽字符位置现在有内容）
// - 正常单元格: 比较 Cluster、Style 和 Selected
func IsCellChanged(cell, prevCell Cell) bool {
	// 如果当前单元格是延续单元格
	if cell.IsContinuation {
		// 如果 prevCell 也是 continuation，不刷新（由主单元格处理）
		// 如果 prevCell 不是 continuation，需要刷新（prevCell 的内容需要被擦除）
		return !prevCell.IsContinuation
	}

	// 如果前一个单元格是 continuation，当前是 head → 需要刷新
	// 这表示一个宽字符被新字符覆盖，必须刷新
	if prevCell.IsContinuation {
		return true
	}

	// 如果前一个是宽字符头，当前是空/continuation → 需要刷新
	// 这表示宽字符被部分覆盖
	if prevCell.Width == 2 && cell.Width == 0 {
		return true
	}

	// 正常比较 Cluster、Style 和 Selected（文本选择高亮）
	return cell.Cluster != prevCell.Cluster ||
		cell.Style != prevCell.Style ||
		cell.Selected != prevCell.Selected
}

// GetCellWidth 获取单元格的显示宽度
// 如果是延续单元格，返回 0
func GetCellWidth(cell Cell) int {
	if cell.IsContinuation {
		return 0
	}
	return cell.Width
}

// ShouldSkipCell 判断是否应该跳过输出此单元格
func ShouldSkipCell(cell Cell) bool {
	return cell.IsContinuation
}

// clearCellAt 清除指定位置的单元格，并正确处理宽字符的关联清理。
// 这是设计文档中的关键修复：
// 1. 如果是 continuation，往左清除 head
// 2. 如果是宽字符头，清除右侧 continuation
// 3. 然后清除当前位置
func (b *Buffer) clearCellAt(x, y int) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}

	cell := b.Cells[y][x]

	// 如果当前位置是 continuation，需要往左找到 head 并清除
	// 但要检查 head 是否与当前位置宽度相同（防止误清除）
	if cell.IsContinuation && x > 0 {
		head := b.Cells[y][x-1]
		// 只清除 head 当它是宽字符且宽度大于0
		if head.Width == 2 && head.Cluster != "" {
			b.Cells[y][x-1] = Cell{}
		}
	}

	// 如果当前位置是宽字符头，需要清除右侧的 continuation
	if cell.Width == 2 && x+1 < b.Width {
		b.Cells[y][x+1] = Cell{}
	}

	// 清除当前位置
	b.Cells[y][x] = Cell{}
}

// ClearWideChar 清除从 (x,y) 开始的宽字符
// 如果该位置的字符是宽字符，会同时清除其延续单元格
// 保留此方法以保持向后兼容，内部调用 clearCellAt
func (b *Buffer) ClearWideChar(x, y int) {
	b.clearCellAt(x, y)
}

// Rect represents a rectangular area.
// We duplicate this simple struct here or share it.
// For now, let's define it here to make paint package self-contained.
type Rect struct {
	X, Y, Width, Height int
}

// Intersect calculates the intersection of two rectangles.
// Returns nil if there is no intersection.
func (r Rect) Intersect(other *Rect) *Rect {
	if other == nil {
		return &r
	}

	x1 := maxInt(r.X, other.X)
	y1 := maxInt(r.Y, other.Y)
	x2 := minInt(r.X+r.Width, other.X+other.Width)
	y2 := minInt(r.Y+r.Height, other.Y+other.Height)

	if x1 >= x2 || y1 >= y2 {
		return nil
	}

	return &Rect{
		X:      x1,
		Y:      y1,
		Width:  x2 - x1,
		Height: y2 - y1,
	}
}

// Contains checks if a point is inside the rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width &&
		y >= r.Y && y < r.Y+r.Height
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// Buffer Output and Selection Methods
// =============================================================================

// String returns the buffer as a string with ANSI escape codes for styling.
// This outputs the buffer with proper terminal styling support.
// Applies selection highlighting (reverse video) to selected cells.
func (b *Buffer) String() string {
	w, h := b.Width, b.Height
	if h == 0 {
		return ""
	}

	lines := make([]string, h)
	for y := 0; y < h; y++ {
		var lineBuilder strings.Builder

		for x := 0; x < w; x++ {
			cell := b.Cells[y][x]

			// Skip continuation cells (wide characters)
			if cell.IsContinuation {
				continue
			}

			// Apply reverse video if selected
			if cell.Selected {
				lineBuilder.WriteString("\x1b[7m")
			}

			// Apply style if present
			if cell.Style != (style.Style{}) {
				lineBuilder.WriteString(cell.Style.ToANSI())
			}

			if cell.Cluster == "" || cell.Cluster == "\x00" {
				lineBuilder.WriteRune(' ')
			} else {
				lineBuilder.WriteString(cell.Cluster)
			}

			// Reset style after each cell
			if cell.Style != (style.Style{}) || cell.Selected {
				lineBuilder.WriteString("\x1b[0m")
			}
		}

		lines[y] = lineBuilder.String()
	}

	return strings.Join(lines, "\r\n")
}

// =============================================================================
// Optimized Rendering Methods
// =============================================================================

// run represents a sequence of identical styles for optimization
type run struct {
	text  string
	style style.Style
	start int
	width int
}

// StringOptimized returns the buffer as an optimized string with ANSI codes.
// This version uses run-merging and style state machine to reduce output size.
// Significant performance improvement for buffers with large continuous regions.
func (b *Buffer) StringOptimized() string {
	w, h := b.Width, b.Height
	if h == 0 {
		return ""
	}

	var output bytes.Buffer
	output.Grow(h*(w+2) + len("\x1b[0m"))
	styleMachine := NewStyleStateMachine()

	for y := 0; y < h; y++ {
		b.emitRowOptimized(&output, styleMachine, y, w)
		output.WriteString("\r\n")
	}

	// Reset final style
	output.WriteString("\x1b[0m")

	return output.String()
}

func (b *Buffer) emitRowOptimized(out *bytes.Buffer, styleMachine *StyleStateMachine, y int, width int) {
	if y >= len(b.Cells) {
		return
	}

	x := 0
	for x < width {
		cell := b.Cells[y][x]
		if cell.IsContinuation {
			x++
			continue
		}

		runStyle := cell.Style
		if styleMachine.NeedsUpdate(runStyle) {
			out.WriteString(styleMachine.Update(runStyle))
		}

		for {
			if cell.Cluster == "" || cell.Cluster == "\x00" {
				out.WriteByte(' ')
			} else {
				out.WriteString(cell.Cluster)
			}

			x++
			if x >= width {
				break
			}

			cell = b.Cells[y][x]
			if cell.IsContinuation || cell.Style != runStyle {
				break
			}
		}
	}
}

// encodeRuns encodes a buffer row into runs of identical styles
func (b *Buffer) encodeRuns(y int, width int) []run {
	if y >= len(b.Cells) {
		return nil
	}

	runs := make([]run, 0, 8)
	x := 0

	for x < width {
		cell := b.Cells[y][x]

		// Skip continuation cells
		if cell.IsContinuation {
			x++
			continue
		}

		// Start a new run
		startX := x
		runStyle := cell.Style
		runWidth := 0
		var runText strings.Builder
		runText.Grow(width - x)

		// Build the run
		if cell.Cluster == "" || cell.Cluster == "\x00" {
			runText.WriteByte(' ')
		} else {
			runText.WriteString(cell.Cluster)
		}
		if cell.Width > 0 {
			runWidth += cell.Width
		} else {
			runWidth++
		}

		// Check if we should merge with next cells
		for {
			x++
			if x >= width {
				break
			}

			nextCell := b.Cells[y][x]

			// Stop on continuation cells
			if nextCell.IsContinuation {
				break
			}

			// Stop if style differs
			if nextCell.Style != runStyle {
				break
			}

			// Add to run text
			if nextCell.Cluster == "" || nextCell.Cluster == "\x00" {
				runText.WriteByte(' ')
			} else {
				runText.WriteString(nextCell.Cluster)
			}
			if nextCell.Width > 0 {
				runWidth += nextCell.Width
			} else {
				runWidth++
			}
		}

		runs = append(runs, run{
			text:  runText.String(),
			style: runStyle,
			start: startX,
			width: runWidth,
		})
	}

	return runs
}

// emitRunOptimized emits a run with style state machine optimization
func (b *Buffer) emitRunOptimized(out *bytes.Buffer, styleMachine *StyleStateMachine, run run) {
	// Set style only if it changed
	if styleMachine.NeedsUpdate(run.style) {
		out.WriteString(styleMachine.Update(run.style))
	}

	// Skip selected handling in optimized version for performance
	// (selection is typically handled by Renderer with dirty tracking)

	// Output text
	out.WriteString(run.text)
}

// SetSelected sets the Selected flag for a cell at the given position.
func (b *Buffer) SetSelected(x, y int, selected bool) {
	w, h := b.Width, b.Height
	if x < 0 || x >= w || y < 0 || y >= h {
		return
	}
	b.Cells[y][x].Selected = selected
}

// ClearSelection clears the selection flag for all cells.
func (b *Buffer) ClearSelection() {
	w, h := b.Width, b.Height
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b.Cells[y][x].Selected = false
		}
	}
}

// Clear clears the entire buffer.
func (b *Buffer) Clear() {
	w, h := b.Width, b.Height
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b.Cells[y][x] = Cell{
				Cluster: " ",
				Style:   style.Style{},
				ZIndex:  0,
			}
		}
	}
}

// Reset resets the buffer to the given dimensions.
// This is used by the pool to reuse buffers.
func (b *Buffer) Reset(width, height int) {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Check if we need to allocate new cells
	if b.Cells == nil || len(b.Cells) != height || (len(b.Cells) > 0 && len(b.Cells[0]) != width) {
		b.Cells = make([][]Cell, height)
		for y := 0; y < height; y++ {
			b.Cells[y] = make([]Cell, width)
		}
	}

	// Clear all cells
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			b.Cells[y][x] = Cell{
				Cluster: " ",
				Style:   style.Style{},
				ZIndex:  0,
			}
		}
	}

	b.Width = width
	b.Height = height

	// Reset line hash
	b.LineHash = nil
}

// =============================================================================
// Line Hash Methods (Renderer 2.0)
// =============================================================================

// HashLine computes a hash for a single line using FNV-1a algorithm.
// This enables O(1) line comparison for efficient diff.
func HashLine(row []Cell) uint64 {
	// FNV-1a offset basis
	var h uint64 = 1469598103934665603

	for _, c := range row {
		// Hash cluster content
		for _, r := range c.Cluster {
			h ^= uint64(r)
			h *= 1099511628211
		}

		// Hash style (for style changes detection)
		h ^= uint64(c.Style.Hash())
		h *= 1099511628211

		// Hash width
		h ^= uint64(c.Width)
		h *= 1099511628211

		// Hash continuation flag
		if c.IsContinuation {
			h ^= 1
			h *= 1099511628211
		}
	}

	return h
}

// Rehash recomputes all line hashes for the buffer.
// Should be called after buffer content changes before diff.
func (b *Buffer) Rehash() {
	if b.LineHash == nil || len(b.LineHash) != b.Height {
		b.LineHash = make([]uint64, b.Height)
	}

	for y := 0; y < b.Height; y++ {
		if y < len(b.Cells) {
			b.LineHash[y] = HashLine(b.Cells[y])
		}
	}
}

// EqualLine checks if a specific line equals between two buffers.
// Falls back to cell-by-cell comparison if hashes don't match.
func (b *Buffer) EqualLine(other *Buffer, y int) bool {
	// Bounds check
	if y >= b.Height || y >= other.Height {
		return false
	}
	if y >= len(b.Cells) || y >= len(other.Cells) {
		return false
	}

	// Fast path: compare hashes
	if b.LineHash != nil && other.LineHash != nil {
		if y < len(b.LineHash) && y < len(other.LineHash) {
			if b.LineHash[y] != other.LineHash[y] {
				return false
			}
			// Hashes match, lines are equal
			return true
		}
	}

	// Slow path: cell-by-cell comparison
	row1 := b.Cells[y]
	row2 := other.Cells[y]
	width := minInt(len(row1), len(row2))

	for x := 0; x < width; x++ {
		if !IsCellEqual(row1[x], row2[x]) {
			return false
		}
	}

	return true
}

// IsCellEqual compares two cells for equality.
// Used for cell-by-cell comparison when hash is not available.
func IsCellEqual(a, b Cell) bool {
	return a.Cluster == b.Cluster &&
		a.Style == b.Style &&
		a.Width == b.Width &&
		a.IsContinuation == b.IsContinuation
}
