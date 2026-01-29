# TUI 宽字符处理架构指南

## 问题描述

在终端 UI 中，某些字符（CJK、Emoji 等）的显示宽度为 2，称为**宽字符**。如果处理不当，会导致：

1. **跳空输入**：使用 `for i, r := range text` 循环时，`i` 是 rune 索引而非显示列位置
2. **跳空输出**：输出时没有正确跳过延续单元格
3. **输出不完整**：宽字符的延续单元格被覆盖

## 错误示例

```go
// ❌ 错误：使用 rune 索引作为列位置
for i, r := range text {
    buf.SetCell(x+i, y, r, style)  // 宽字符会导致位置错乱
}

// ❌ 错误：输出时没有跳过延续单元格
for x := 0; x < width; x++ {
    cell := buf.Cells[y][x]
    fmt.Print(string(cell.Char))  // 会重复输出延续单元格
    x++  // 应该根据字符宽度递增
}
```

## 正确做法

### 1. 写入文本：使用 SetString

```go
// ✅ 正确：使用 SetString
buf.SetString(x, y, "Scheduler演示", style)

// ✅ 或使用 PaintContext.SetString
ctx.SetString(0, 0, "=== 日志面板 ===", style)
```

### 2. 必须使用 SetCell 时，使用辅助函数

```go
// ✅ 正确：按字符宽度递增
col := x
for _, r := range text {
    buf.SetCell(col, y, r, style)
    col += runeWidth(r)  // 使用字符宽度递增
}
```

### 3. 输出缓冲区时跳过延续单元格

```go
// ✅ 正确：跳过延续单元格
for x := 0; x < width; {
    cell := buf.Cells[y][x]

    // 跳过延续单元格
    if cell.IsContinuation {
        x++
        continue
    }

    // 输出字符
    if cell.Char != 0 {
        fmt.Print(string(cell.Char))
    }

    // 按字符宽度递增
    if cell.Width > 0 {
        x += cell.Width
    } else {
        x++
    }
}
```

## 架构设计

### Cell 结构

```go
type Cell struct {
    Char           rune     // 字符
    Style          Style    // 样式
    Width          int      // 显示宽度 (1 或 2)
    IsContinuation bool     // 是否为宽字符的延续单元格
}
```

### 写入层 (Buffer)

```go
// SetCell 自动处理延续标记
func (b *Buffer) SetCell(x, y int, char rune, s style.Style) {
    width := runeWidth(char)
    b.Cells[y][x] = Cell{Char: char, Style: s, Width: width, IsContinuation: false}
    if width == 2 && x+1 < b.Width {
        b.Cells[y][x+1] = Cell{IsContinuation: true}  // 标记延续
    }
}

// SetString 按宽度递增位置
func (b *Buffer) SetString(x, y int, text string, s style.Style) {
    col := x
    for _, char := range text {
        if b.Cells[y][col].IsContinuation {
            b.Cells[y][col] = Cell{}  // 清除旧延续
        }
        b.SetCell(col, y, char, s)
        col += runeWidth(char)  // 按宽度递增
    }
}
```

### 输出层 (Render)

```go
// 辅助函数
func IsCellChanged(cell, prevCell Cell) bool {
    if cell.IsContinuation || prevCell.IsContinuation {
        return false  // 延续单元格不比较
    }
    return cell.Char != prevCell.Char || cell.Style != prevCell.Style
}

func GetCellWidth(cell Cell) int {
    if cell.IsContinuation {
        return 0  // 延续单元格宽度为 0
    }
    return cell.Width
}

func ShouldSkipCell(cell Cell) bool {
    return cell.IsContinuation
}
```

## API 使用指南

### 组件开发者

| 场景 | 推荐方法 | 避免使用 |
|------|----------|----------|
| 写入字符串 | `buf.SetString()` | `for i, r := range text` + `SetCell` |
| 写入单个字符 | `buf.SetCell()` | - |
| 居中文本 | `ctx.DrawString(x, y, text, AlignCenter, style)` | 手动计算位置 |
| 输出缓冲区 | 使用 `GetCellWidth()` 递增 | `x++` 固定递增 |

### 最佳实践

1. **优先使用高层 API**：`SetString`、`DrawText`、`Painter`
2. **避免手动位置计算**：让框架处理宽度计算
3. **使用辅助函数**：`runeWidth()`、`textDisplayWidth()`
4. **测试宽字符**：包含中文、Emoji 的测试用例

## 扩展：Painter 抽象

对于更复杂的绘制场景，使用 `Painter` 包装器：

```go
painter := paint.NewPainter(&ctx)

// 自动处理宽度
painter.Print(0, 0, "Hello 世界", style)
painter.Printf(0, 1, "计数: %d", 123, style)

// 绘制边框
painter.DrawBorder(0, 0, width, height, style)
```

## 相关文件

- `runtime/paint/cell.go` - Cell 结构定义
- `runtime/paint/buffer.go` - Buffer 和写入逻辑
- `runtime/paint/context.go` - PaintContext 和绘制方法
- `framework/display/text.go` - Text 组件（正确处理宽字符）
