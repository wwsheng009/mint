# TextBuffer 设计文档

**版本**: v1.0
**日期**: 2026-01-31
**来源**: idea/idea4.4_input.md
**状态**: 🟡 中优先级

---

## 一、概述

### 1.1 设计目标

实现一个**文本编辑器级的输入缓冲区**，用于支持复杂的文本输入场景：

- UTF-8 多字节字符（中文、日文、emoji）
- 光标移动（字符、单词、行）
- 文本选择和复制粘贴
- 撤销/重做
- 水平滚动

### 1.2 核心问题

**问题 1**: Go 的 string 是 byte 数组，直接操作会导致中文乱码

```go
// ❌ 错误：按字节操作
s := "你好"
s = s[:1] // 会截断 UTF-8 序列，导致乱码

// ✅ 正确：按 rune 操作
runes := []rune("你好")
s = string(runes[:1]) // "你"
```

**问题 2**: 光标位置需要是"逻辑位置"而非"物理位置"

```go
// 逻辑光标：第 N 个字符（rune 索引）
// 物理光标：屏幕上的绝对坐标
```

---

## 二、核心数据结构

### 2.1 TextBuffer

```go
// input/buffer.go

package input

import (
    "sync"
    "unicode/utf8"
)

// TextBuffer 文本缓冲区
type TextBuffer struct {
    mu       sync.RWMutex
    runes    []rune      // UTF-32 存储，避免中文问题
    cursor   int         // 光标位置（rune 索引）
    anchor   int         // 选择锚点
    scroll   int         // 水平滚动偏移
    maxLen   int         // 最大长度（0 表示无限制）
    history  *History    // 撤销/重做历史
}

// NewTextBuffer 创建文本缓冲区
func NewTextBuffer() *TextBuffer {
    return &TextBuffer{
        runes:   make([]rune, 0),
        cursor:  0,
        anchor:  -1, // -1 表示无选择
        scroll:  0,
        maxLen:  0,
        history: NewHistory(),
    }
}

// NewTextBufferWithInitial 创建带初始内容的缓冲区
func NewTextBufferWithInitial(text string) *TextBuffer {
    tb := NewTextBuffer()
    tb.SetText(text)
    return tb
}

// Length 返回文本长度（rune 数量）
func (b *TextBuffer) Length() int {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return len(b.runes)
}

// String 返回文本内容
func (b *TextBuffer) String() string {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return string(b.runes)
}

// Runes 返回 rune 切片（副本）
func (b *TextBuffer) Runes() []rune {
    b.mu.RLock()
    defer b.mu.RUnlock()
    result := make([]rune, len(b.runes))
    copy(result, b.runes)
    return result
}
```

### 2.2 Selection 选择区

```go
// input/selection.go

package input

// Selection 选择区
type Selection struct {
    Start int // 起始位置（rune 索引）
    End   int // 结束位置
}

// IsEmpty 检查是否为空选择
func (s Selection) IsEmpty() bool {
    return s.Start == s.End
}

// Normalized 返回规范化的选择（Start <= End）
func (s Selection) Normalized() Selection {
    if s.Start <= s.End {
        return s
    }
    return Selection{Start: s.End, End: s.Start}
}

// Length 返回选择长度
func (s Selection) Length() int {
    return s.End - s.Start
}

// Contains 检查是否包含指定位置
func (s Selection) Contains(pos int) bool {
    norm := s.Normalized()
    return pos >= norm.Start && pos < norm.End
}

// Text 返回选中的文本
func (s Selection) Text(buffer []rune) string {
    norm := s.Normalized()
    if norm.Start < 0 || norm.End > len(buffer) {
        return ""
    }
    return string(buffer[norm.Start:norm.End])
}

// GetSelection 获取当前选择
func (b *TextBuffer) GetSelection() Selection {
    b.mu.RLock()
    defer b.mu.RUnlock()

    if b.anchor < 0 {
        return Selection{Start: b.cursor, End: b.cursor}
    }

    return Selection{Start: b.anchor, End: b.cursor}
}

// SetSelection 设置选择
func (b *TextBuffer) SetSelection(start, end int) {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 边界检查
    if start < 0 {
        start = 0
    }
    if start > len(b.runes) {
        start = len(b.runes)
    }
    if end < 0 {
        end = 0
    }
    if end > len(b.runes) {
        end = len(b.runes)
    }

    b.anchor = start
    b.cursor = end
}

// ClearSelection 清除选择
func (b *TextBuffer) ClearSelection() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.anchor = -1
}

// HasSelection 检查是否有选择
func (b *TextBuffer) HasSelection() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.anchor >= 0 && b.anchor != b.cursor
}
```

---

## 三、文本编辑操作

### 3.1 插入和删除

```go
// input/edit.go

package input

// Insert 在光标位置插入文本
func (b *TextBuffer) Insert(text string) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 保存状态用于撤销
    b.history.Save(b)

    // 解码为 rune
    runes := []rune(text)

    // 检查长度限制
    if b.maxLen > 0 && len(b.runes)+len(runes) > b.maxLen {
        return ErrMaxLength
    }

    // 删除选中的文本
    if b.HasSelection() {
        b.deleteSelection()
    }

    // 插入 rune
    b.runes = insertRunes(b.runes, b.cursor, runes)
    b.cursor += len(runes)

    return nil
}

// Delete 删除光标后的字符
func (b *TextBuffer) Delete(count int) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 如果有选择，删除选择内容
    if b.HasSelection() {
        b.history.Save(b)
        b.deleteSelection()
        return nil
    }

    if count <= 0 {
        return nil
    }

    if b.cursor >= len(b.runes) {
        return nil
    }

    b.history.Save(b)

    // 计算实际删除数量
    end := b.cursor + count
    if end > len(b.runes) {
        end = len(b.runes)
    }

    // 删除 rune
    b.runes = append(b.runes[:b.cursor], b.runes[end:]...)

    return nil
}

// Backspace 删除光标前的字符
func (b *TextBuffer) Backspace(count int) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 如果有选择，删除选择内容
    if b.HasSelection() {
        b.history.Save(b)
        b.deleteSelection()
        return nil
    }

    if count <= 0 || b.cursor == 0 {
        return nil
    }

    b.history.Save(b)

    // 计算实际删除数量
    start := b.cursor - count
    if start < 0 {
        start = 0
    }

    // 删除 rune
    b.runes = append(b.runes[:start], b.runes[b.cursor:]...)
    b.cursor = start

    return nil
}

// deleteSelection 删除选中的文本
func (b *TextBuffer) deleteSelection() {
    sel := b.GetSelection().Normalized()
    b.runes = append(b.runes[:sel.Start], b.runes[sel.End:]...)
    b.cursor = sel.Start
    b.anchor = -1
}

// insertRunes 在指定位置插入 rune
func insertRunes(runes []rune, pos int, insert []rune) []rune {
    result := make([]rune, 0, len(runes)+len(insert))
    result = append(result, runes[:pos]...)
    result = append(result, insert...)
    result = append(result, runes[pos:]...)
    return result
}
```

### 3.2 文本设置

```go
// SetText 设置文本内容
func (b *TextBuffer) SetText(text string) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.history.Save(b)

    b.runes = []rune(text)
    b.cursor = 0
    b.anchor = -1
    b.scroll = 0
}

// Clear 清空内容
func (b *TextBuffer) Clear() {
    b.SetText("")
}
```

---

## 四、光标移动

### 4.1 基础移动

```go
// input/cursor.go

package input

// MoveCursor 移动光标
func (b *TextBuffer) MoveCursor(delta int) bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    newCursor := b.cursor + delta

    // 边界检查
    if newCursor < 0 {
        newCursor = 0
    }
    if newCursor > len(b.runes) {
        newCursor = len(b.runes)
    }

    b.cursor = newCursor
    b.anchor = -1 // 移动光标时清除选择

    return true
}

// MoveToStart 移动到开头
func (b *TextBuffer) MoveToStart() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.cursor = 0
    b.anchor = -1
}

// MoveToEnd 移动到结尾
func (b *TextBuffer) MoveToEnd() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.cursor = len(b.runes)
    b.anchor = -1
}

// CursorPos 返回光标位置
func (b *TextBuffer) CursorPos() int {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.cursor
}
```

### 4.2 单词移动

```go
// input/word.go

package input

import "unicode"

// MoveWordForward 移动到下一个单词
func (b *TextBuffer) MoveWordForward() bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.cursor >= len(b.runes) {
        return false
    }

    // 跳过当前单词
    for b.cursor < len(b.runes) && !unicode.IsSpace(b.runes[b.cursor]) {
        b.cursor++
    }

    // 跳过空格
    for b.cursor < len(b.runes) && unicode.IsSpace(b.runes[b.cursor]) {
        b.cursor++
    }

    b.anchor = -1
    return true
}

// MoveWordBackward 移动到上一个单词
func (b *TextBuffer) MoveWordBackward() bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.cursor <= 0 {
        return false
    }

    b.cursor--

    // 跳过空格
    for b.cursor > 0 && unicode.IsSpace(b.runes[b.cursor]) {
        b.cursor--
    }

    // 跳过当前单词
    for b.cursor > 0 && !unicode.IsSpace(b.runes[b.cursor-1]) {
        b.cursor--
    }

    b.anchor = -1
    return true
}

// DeleteWord 删除到下一个单词
func (b *TextBuffer) DeleteWord() error {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.HasSelection() {
        return b.Delete(1)
    }

    start := b.cursor
    end := b.cursor

    // 找到单词结束
    for end < len(b.runes) && !unicode.IsSpace(b.runes[end]) {
        end++
    }

    // 找到单词开始（对于 Backspace）
    for start > 0 && unicode.IsSpace(b.runes[start-1]) {
        start--
    }

    b.history.Save(b)
    b.runes = append(b.runes[:start], b.runes[end:]...)
    b.cursor = start

    return nil
}
```

### 4.3 行移动（多行支持）

```go
// input/line.go

package input

// MoveLineUp 移动到上一行
func (b *TextBuffer) MoveLineUp(visibleWidth int) bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 计算当前行的起始位置
    lineStart := b.findLineStart(b.cursor)
    if lineStart == 0 {
        return false // 已经在第一行
    }

    // 计算上一行的起始位置
    prevLineStart := b.findLineStart(lineStart - 1)

    // 计算列偏移
    col := b.cursor - lineStart

    // 移动到上一行的相同列
    b.cursor = prevLineStart + col
    if b.cursor > len(b.runes) {
        b.cursor = len(b.runes)
    }

    b.anchor = -1
    return true
}

// MoveLineDown 移动到下一行
func (b *TextBuffer) MoveLineDown(visibleWidth int) bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 找到当前行结束
    lineEnd := b.findLineEnd(b.cursor)
    if lineEnd >= len(b.runes) {
        return false // 已经在最后一行
    }

    // 计算下一行结束
    nextLineEnd := b.findLineEnd(lineEnd + 1)

    // 计算列偏移
    lineStart := b.findLineStart(b.cursor)
    col := b.cursor - lineStart

    // 移动到下一行的相同列
    b.cursor = lineEnd + 1 + col
    if b.cursor > nextLineEnd {
        b.cursor = nextLineEnd
    }

    b.anchor = -1
    return true
}

// findLineStart 找到行起始位置
func (b *TextBuffer) findLineStart(pos int) int {
    for pos > 0 && b.runes[pos-1] != '\n' {
        pos--
    }
    return pos
}

// findLineEnd 找到行结束位置
func (b *TextBuffer) findLineEnd(pos int) int {
    for pos < len(b.runes) && b.runes[pos] != '\n' {
        pos++
    }
    return pos
}
```

---

## 五、水平滚动

### 5.1 滚动计算

```go
// input/scroll.go

package input

// ScrollOffset 返回水平滚动偏移
func (b *TextBuffer) ScrollOffset() int {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.scroll
}

// EnsureCursorVisible 确保光标可见
func (b *TextBuffer) EnsureCursorVisible(visibleWidth int) {
    b.mu.Lock()
    defer b.mu.Unlock()

    if visibleWidth <= 0 {
        return
    }

    // 计算光标相对于滚动位置的偏移
    relPos := b.cursor - b.scroll

    // 如果光标在可见区域左侧
    if relPos < 0 {
        b.scroll = b.cursor
        return
    }

    // 如果光标在可见区域右侧
    if relPos >= visibleWidth {
        b.scroll = b.cursor - visibleWidth + 1
        return
    }
}

// SetScrollOffset 设置滚动偏移
func (b *TextBuffer) SetScrollOffset(offset int) {
    b.mu.Lock()
    defer b.mu.Unlock()

    if offset < 0 {
        offset = 0
    }

    b.scroll = offset
}

// ScrollLeft 向左滚动
func (b *TextBuffer) ScrollLeft(amount int) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.scroll -= amount
    if b.scroll < 0 {
        b.scroll = 0
    }
}

// ScrollRight 向右滚动
func (b *TextBuffer) ScrollRight(amount int) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.scroll += amount
}
```

---

## 六、撤销/重做

### 6.1 History

```go
// input/history.go

package input

import (
    "sync"
)

// History 撤销/重做历史
type History struct {
    mu     sync.RWMutex
    past   []string // 历史记录
    future []string // 未来记录（用于 redo）
    maxSize int     // 最大历史记录数
}

// NewHistory 创建历史记录
func NewHistory() *History {
    return &History{
        past:     make([]string, 0),
        future:   make([]string, 0),
        maxSize: 100,
    }
}

// Save 保存当前状态
func (h *History) Save(buffer *TextBuffer) {
    h.mu.Lock()
    defer h.mu.Unlock()

    current := buffer.String()

    // 避免重复保存
    if len(h.past) > 0 && h.past[len(h.past)-1] == current {
        return
    }

    h.past = append(h.past, current)
    h.future = h.future[:0] // 清空 future

    // 限制大小
    if len(h.past) > h.maxSize {
        h.past = h.past[1:]
    }
}

// Undo 撤销
func (h *History) Undo(buffer *TextBuffer) bool {
    h.mu.Lock()
    defer h.mu.Unlock()

    if len(h.past) <= 1 {
        return false
    }

    // 保存当前状态到 future
    current := buffer.String()
    h.future = append(h.future, current)

    // 恢复上一个状态
    h.past = h.past[:len(h.past)-1]
    buffer.SetText(h.past[len(h.past)-1])

    return true
}

// Redo 重做
func (h *History) Redo(buffer *TextBuffer) bool {
    h.mu.Lock()
    defer h.mu.Unlock()

    if len(h.future) == 0 {
        return false
    }

    // 恢复 future 状态
    state := h.future[len(h.future)-1]
    h.future = h.future[:len(h.future)-1]
    h.past = append(h.past, state)

    buffer.SetText(state)
    return true
}

// CanUndo 是否可以撤销
func (h *History) CanUndo() bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return len(h.past) > 1
}

// CanRedo 是否可以重做
func (h *History) CanRedo() bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return len(h.future) > 0
}

// Clear 清空历史
func (h *History) Clear() {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.past = h.past[:0]
    h.future = h.future[:0]
}
```

---

## 七、UI 集成

### 7.1 Input 组件集成

```go
// components/form/input.go

package form

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/framework/input"
)

// InputState 输入组件状态
type InputState struct {
    buffer      *input.TextBuffer
    placeholder string
    password    bool
    focused     bool
}

// NewInputState 创建输入状态
func NewInputState() *InputState {
    return &InputState{
        buffer: input.NewTextBuffer(),
    }
}

// HandleInput 处理输入事件
func (s *InputState) HandleInput(e ui.KeyEvent) bool {
    switch {
    case e.Key == '\x1b': // ESC
        // 取消
        return true

    case e.Key == '\r': // Enter
        // 提交
        return true

    case e.Key == '\t': // Tab
        // 切换焦点
        return false

    case e.Mod == ui.ModCtrl && e.Key == 'u':
        // Ctrl+U: 清空
        s.buffer.Clear()
        return true

    case e.Mod == ui.ModCtrl && e.Key == 'w':
        // Ctrl+W: 删除单词
        s.buffer.DeleteWord()
        return true

    case e.Key == 127 || e.Key == 8: // Backspace
        s.buffer.Backspace(1)
        return true

    case e.Key >= 32 && e.Key <= 126: // 可打印字符
        s.buffer.Insert(string(rune(e.Key)))
        return true

    case e.Key >= 0xD800 && e.Key <= 0xDFFF: // Emoji 等多字节字符
        // 等待完整序列
        return true

    default:
        return false
    }
}
```

---

## 八、实施计划

### 阶段 1: 基础实现

- [ ] 实现 TextBuffer 基础结构
- [ ] 实现 Insert/Delete 操作
- [ ] 实现光标移动

### 阶段 2: 高级功能

- [ ] 实现选择区
- [ ] 实现单词操作
- [ ] 实现行操作

### 阶段 3: 撤销重做

- [ ] 实现 History
- [ ] 实现 Undo/Redo
- [ ] 集成快捷键

### 阶段 4: UI 集成

- [ ] 集成到 Input 组件
- [ ] 实现水平滚动
- [ ] 实现密码模式

---

## 九、测试策略

```go
// input/buffer_test.go

func TestTextBufferInsert(t *testing.T) {
    tb := NewTextBuffer()

    tb.Insert("Hello")
    assert.Equal(t, "Hello", tb.String())
    assert.Equal(t, 5, tb.CursorPos())

    tb.Insert(" World")
    assert.Equal(t, "Hello World", tb.String())
}

func TestTextBufferChinese(t *testing.T) {
    tb := NewTextBuffer()

    tb.Insert("你好世界")
    assert.Equal(t, "你好世界", tb.String())
    assert.Equal(t, 4, tb.CursorPos()) // 4 个 rune

    tb.Backspace(1)
    assert.Equal(t, "你好世", tb.String())
}

func TestTextBufferSelection(t *testing.T) {
    tb := NewTextBufferWithInitial("Hello World")

    tb.SetSelection(0, 5)
    assert.True(t, tb.HasSelection())
    assert.Equal(t, "Hello", tb.GetSelection().Text(tb.Runes()))

    tb.Delete(1)
    assert.Equal(t, " World", tb.String())
}

func TestTextBufferHistory(t *testing.T) {
    tb := NewTextBuffer()

    tb.Insert("Hello")
    tb.Insert(" World")

    tb.Undo()
    assert.Equal(t, "Hello", tb.String())

    tb.Redo()
    assert.Equal(t, "Hello World", tb.String())
}
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-31
