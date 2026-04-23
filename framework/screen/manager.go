package screen

import (
	"io"
	"os"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"golang.org/x/term"
)

// Manager 屏幕管理器
type Manager struct {
	// 终端操作通过低级系统调用实现
	width  int
	height int

	// 双缓冲
	front *Buffer
	back  *Buffer

	// 光标
	cursor        Cursor
	cursorVisible bool

	out      io.Writer
	stdin    *os.File
	rawState *term.State
}

// NewManager 创建屏幕管理器
func NewManager(width, height int) *Manager {
	return &Manager{
		width:  width,
		height: height,
		cursor: Cursor{},
		out:    os.Stdout,
		stdin:  os.Stdin,
	}
}

// Init 初始化屏幕
func (m *Manager) Init() error {
	m.ensureBuffers(m.width, m.height)

	// 进入备用屏幕
	if err := m.writeString("\x1b[?1049h"); err != nil {
		return err
	}

	// 启用原始模式
	if err := m.enableRawMode(); err != nil {
		_ = m.writeString("\x1b[?1049l")
		return err
	}

	// 隐藏光标
	if err := m.hideCursor(); err != nil {
		_ = m.disableRawMode()
		_ = m.writeString("\x1b[?1049l")
		return err
	}
	m.cursorVisible = false

	return nil
}

// Close 关闭屏幕
func (m *Manager) Close() error {
	var firstErr error

	// 显示光标
	if err := m.showCursor(); err != nil && firstErr == nil {
		firstErr = err
	}
	m.cursorVisible = true

	// 退出原始模式
	if err := m.disableRawMode(); err != nil && firstErr == nil {
		firstErr = err
	}

	// 退出备用屏幕
	if err := m.writeString("\x1b[?1049l"); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

// Render 渲染缓冲区
func (m *Manager) Render(buf *Buffer) error {
	if buf == nil {
		return nil
	}

	m.ensureBuffers(buf.width, buf.height)
	m.back.CopyFrom(buf)

	// 计算差异
	diff := m.diff(m.front, m.back)

	// 输出变更
	if err := m.drawChanges(diff); err != nil {
		return err
	}

	// 更新前缓冲
	m.front, m.back = m.back, m.front

	return nil
}

// diff 计算缓冲区差异
func (m *Manager) diff(old, new *Buffer) []Change {
	var changes []Change

	minW := minInt(old.width, new.width)
	minH := minInt(old.height, new.height)

	for y := 0; y < minH; y++ {
		for x := 0; x < minW; x++ {
			oldCell := old.cells[y][x]
			newCell := new.cells[y][x]

			if oldCell.Cluster != newCell.Cluster || oldCell.Style != newCell.Style {
				changes = append(changes, Change{
					X:       x,
					Y:       y,
					Cluster: newCell.Cluster,
					Style:   newCell.Style,
				})
			}
		}
	}

	return changes
}

// drawChanges 绘制变更
func (m *Manager) drawChanges(changes []Change) error {
	if len(changes) == 0 {
		return nil
	}

	batch := paint.NewCommandBatch()
	for _, c := range changes {
		text := c.Cluster
		if text == "" || text == " " {
			text = " "
		}
		batch.Add(c.X, c.Y, text, c.Style)
	}

	return m.writeString(batch.Flush())
}

// GetSize 获取屏幕尺寸
func (m *Manager) GetSize() (width, height int) {
	return m.width, m.height
}

// SetSize 设置屏幕尺寸
func (m *Manager) SetSize(width, height int) {
	m.ensureBuffers(width, height)
}

// SetCursor 设置光标位置
func (m *Manager) SetCursor(x, y int) {
	m.cursor.X = x
	m.cursor.Y = y
	if m.cursorVisible {
		_ = m.moveCursor(x, y)
	}
}

// GetCursor 获取光标位置
func (m *Manager) GetCursor() (x, y int) {
	return m.cursor.X, m.cursor.Y
}

// SetCursorVisible 设置光标可见性
func (m *Manager) SetCursorVisible(visible bool) {
	if visible != m.cursorVisible {
		if visible {
			_ = m.showCursor()
		} else {
			_ = m.hideCursor()
		}
		m.cursorVisible = visible
	}
}

// hideCursor 隐藏光标
func (m *Manager) hideCursor() error {
	return m.writeString("\x1b[?25l")
}

// showCursor 显示光标
func (m *Manager) showCursor() error {
	return m.writeString("\x1b[?25h")
}

// moveCursor 移动光标
func (m *Manager) moveCursor(x, y int) error {
	return m.writeString("\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H")
}

func (m *Manager) writeString(text string) error {
	if text == "" {
		return nil
	}
	if m.out == nil {
		m.out = os.Stdout
	}
	_, err := io.WriteString(m.out, text)
	return err
}

func (m *Manager) ensureBuffers(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	m.width = width
	m.height = height

	if m.front == nil || m.front.width != width || m.front.height != height {
		m.front = NewBuffer(width, height)
	}
	if m.back == nil || m.back.width != width || m.back.height != height {
		m.back = NewBuffer(width, height)
	}
}

func (m *Manager) enableRawMode() error {
	if m.rawState != nil || m.stdin == nil {
		return nil
	}

	fd := int(m.stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	m.rawState = state
	return nil
}

func (m *Manager) disableRawMode() error {
	if m.rawState == nil || m.stdin == nil {
		return nil
	}

	err := term.Restore(int(m.stdin.Fd()), m.rawState)
	m.rawState = nil
	return err
}

// Cursor 光标结构
type Cursor struct {
	X int
	Y int
}

// Change 缓冲区变更
type Change struct {
	X       int
	Y       int
	Cluster string
	Style   style.Style
}

// Buffer 渲染缓冲区
type Buffer struct {
	width  int
	height int
	cells  [][]Cell
}

// Cell 缓冲区单元格
type Cell struct {
	Cluster string
	Style   style.Style
}

// NewBuffer 创建缓冲区
func NewBuffer(width, height int) *Buffer {
	b := &Buffer{
		width:  width,
		height: height,
		cells:  make([][]Cell, height),
	}

	for y := 0; y < height; y++ {
		b.cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			b.cells[y][x] = Cell{Cluster: " "}
		}
	}

	return b
}

// GetSize 获取缓冲区尺寸
func (b *Buffer) GetSize() (width, height int) {
	return b.width, b.height
}

// SetCell 设置单元格
func (b *Buffer) SetCell(x, y int, char rune, s style.Style) {
	if x >= 0 && x < b.width && y >= 0 && y < b.height {
		b.cells[y][x] = Cell{Cluster: string(char), Style: s}
	}
}

// SetLine 设置一行
func (b *Buffer) SetLine(y int, text string, s style.Style) {
	if y < 0 || y >= b.height {
		return
	}

	for x, r := range text {
		if x < b.width {
			b.cells[y][x] = Cell{Cluster: string(r), Style: s}
		}
	}
}

// Fill 填充区域
func (b *Buffer) Fill(x, y, width, height int, char rune, s style.Style) {
	for py := y; py < y+height && py < b.height; py++ {
		for px := x; px < x+width && px < b.width; px++ {
			b.cells[py][px] = Cell{Cluster: string(char), Style: s}
		}
	}
}

// Clear 清空缓冲区
func (b *Buffer) Clear() {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			b.cells[y][x] = Cell{Cluster: " "}
		}
	}
}

// CopyFrom 复制另一个缓冲区的内容。
func (b *Buffer) CopyFrom(src *Buffer) {
	if b == nil || src == nil {
		return
	}

	if b.width != src.width || b.height != src.height {
		*b = *NewBuffer(src.width, src.height)
	}

	for y := 0; y < src.height; y++ {
		copy(b.cells[y], src.cells[y])
	}
}

// GetCell 获取单元格
func (b *Buffer) GetCell(x, y int) Cell {
	if x >= 0 && x < b.width && y >= 0 && y < b.height {
		return b.cells[y][x]
	}
	return Cell{}
}

// String 返回缓冲区的字符串表示
func (b *Buffer) String() string {
	var lines []string
	for y := 0; y < b.height; y++ {
		var line string
		currentStyle := style.Style{}
		for x := 0; x < b.width; x++ {
			cell := b.cells[y][x]
			if cell.Style != currentStyle {
				if currentStyle != (style.Style{}) {
					line += "\x1b[0m"
				}
				line += cell.Style.ToANSI()
				currentStyle = cell.Style
			}
			if cell.Cluster == "" || cell.Cluster == " " {
				line += " "
			} else {
				line += cell.Cluster
			}
		}
		if currentStyle != (style.Style{}) {
			line += "\x1b[0m"
		}
		lines = append(lines, line)
	}
	return joinLines(lines)
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\r\n"
		}
		result += line
	}
	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 11)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
