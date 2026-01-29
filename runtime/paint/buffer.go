package paint

import (
	"github.com/mattn/go-runewidth"
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
		// Initialize with empty cells if needed, or rely on zero value
	}

	return b
}

// SetCell sets the character and style at the given coordinates.
// It handles boundary checks safely and marks continuation cells for wide characters.
// This is a convenience method that converts a rune to a string cluster.
// For complex grapheme clusters, use SetString instead.
func (b *Buffer) SetCell(x, y int, char rune, s style.Style) {
	b.setCluster(x, y, string(char), runeWidth(char), s)
}

// setCluster sets a grapheme cluster at the given coordinates.
// This is the low-level method that all writing operations should use.
func (b *Buffer) setCluster(x, y int, cluster string, width int, s style.Style) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
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
	if width == 2 && x+1 < b.Width {
		b.Cells[y][x+1] = Cell{
			IsContinuation: true,
		}
	}
}

// runeWidth 返回字符的显示宽度 (1 或 2)
func runeWidth(r rune) int {
	// CJK 字符范围 (中文、日文、韩文等)
	if r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a ||
		(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe19) ||
		(r >= 0xfe30 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x20000 && r <= 0x2fffd) ||
		(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	// Emoji 和其他符号
	if r >= 0x1f300 && r <= 0x1f9f0 {
		return 2
	}
	return 1
}

// SetString writes a string starting at (x, y) with the given style.
// This method properly handles grapheme clusters (emoji ZWJ sequences,
// combining characters, flag emojis, etc.) using the uniseg library.
func (b *Buffer) SetString(x, y int, text string, s style.Style) {
	if y < 0 || y >= b.Height {
		return
	}

	col := x
	g := uniseg.NewGraphemes(text)

	for g.Next() {
		cluster := g.Str()                       // 完整字形簇
		width := runewidth.StringWidth(cluster)  // 使用标准库计算宽度

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
}

// Fill fills a rectangular area with a character and style.
func (b *Buffer) Fill(rect Rect, char rune, s style.Style) {
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			b.SetCell(x, y, char, s)
		}
	}
}

// ==============================================================================
// Wide Character Helper Functions
// ==============================================================================

// IsCellChanged 比较两个单元格是否不同，正确处理宽字符。
// IsCellChanged 检查单元格是否有变化
// 延续单元格始终返回 false（被其主单元格处理）
func IsCellChanged(cell, prevCell Cell) bool {
	// 如果当前单元格是延续单元格，跳过（由主单元格处理）
	if cell.IsContinuation {
		return false
	}

	// 如果前一个单元格是延续单元格，忽略其 Cluster，只比较 Style
	// 因为延续单元格的 Cluster 是无效的
	if prevCell.IsContinuation {
		return cell.Style != prevCell.Style
	}

	// 正常比较 Cluster 和 Style
	return cell.Cluster != prevCell.Cluster || cell.Style != prevCell.Style
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
	if cell.IsContinuation && x > 0 {
		head := b.Cells[y][x-1]
		if head.Width == 2 {
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
