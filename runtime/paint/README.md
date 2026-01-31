# Paint System

绘制系统核心实现。

## 职责

- **CellBuffer 管理**：创建、管理、重用绘制缓冲区
- **宽字符支持**：正确处理 CJK 字符、Emoji、组合字符等
- **Z-index 渲染顺序**：支持多层渲染和 Z-Index 控制
- **绘制命令批处理**：合并相邻命令以优化终端 IO
- **图层合成**：支持多图层合成（背景、前景、覆盖等）
- **脏标记优化**：只重绘变化的区域

## 纯 Go 约束

此目录保持纯 Go 实现，外部依赖：
- `github.com/rivo/uniseg`：字形簇处理（emoji ZWJ 序列）
- `github.com/mattn/go-runewidth`：宽度计算

**但是**：此目录不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss（样式由外部传入 style.Style）

## 核心概念

### Cell 和 CellBuffer

`Cell` 是绘制系统的最小单位，代表终端上的一个单元格：

```go
type Cell struct {
    Cluster        string      // 字形簇（可能是 emoji 组合）
    Style          style.Style // 样式信息（颜色、粗体等）
    Width          int         // 显示宽度（1 或 2）
    IsContinuation bool        // 是否为宽字符的延续单元格
    Selected       bool        // 是否被选中（反色显示）
    ZIndex         int         // 层级
    NodeID         string      // 所属组件 ID（用于调试）
}
```

`Buffer` 是 Cell 的二维网格：

```go
type Buffer struct {
    Width  int
    Height int
    Cells  [][]Cell
}
```

### 宽字符处理

**问题**：CJK 字符和某些 emoji 宽度为 2，占用 2 个单元格。需要特别注意：

1. **主单元格**：`IsContinuation=false`，`Width=2`
2. **延续单元格**：`IsContinuation=true`，`Width=0`，不应该被单独绘制

**修复策略**（见 `BUGFIX_diff_rendering.md`）：

- 写入时：设置主单元格，同时标记下一个单元格为 `IsContinuation`
- 清除时：调用 `clearCellAt`，它会同时清除主单元格和延续单元格
- 比较时：使用 `IsCellChanged`，跳过 continuation 单元格的比较

```go
// 绘制宽字符（例如：'中'）
buffer.SetCell(0, 0, '中', style) 
// 自动处理：
//   cells[0][0] = {Cluster: "中", Width: 2, IsContinuation: false}
//   cells[0][1] = {IsContinuation: true}
```

### 字形簇

使用 `uniseg` 库正确处理复杂的字形簇：

- Emoji ZWJ 序列：👨‍👩‍👧（家庭 emoji）
- 组合字符：é（e +  ́）
- 字符序列：🇺🇸（国旗 emoji）

```go
buffer.SetString(x, y, "👨‍👩‍👧", style) // 完整绘制为 1 个字形簇
```

### Z-Index 和图层

`Layer` 支持 Z-Index 排序的合成渲染：

```go
type Layer struct {
    ID      string
    Typ     LayerType  // LayerNormal, LayerStream, LayerOverlay
    ZIndex  int         // 渲染顺序（小的在下面）
    Rect    Rect        // 层级边界
    Buffer  *Buffer     // 层级内容
    Enabled bool
    Visible bool
    dirty   bool        // 脏标记
}
```

### 绘制命令批处理 (CommandBatch)

`CommandBatch` 将多个绘制命令合并以减少终端 IO：

1. **合并相邻命令**：同行、相邻位置、相同样式的命令会被合并
2. **样式状态机**：只在样式变化时发送 ANSI 代码
3. **光标优化**：使用相对移动（`\x1b[nC`）而非绝对定位

```go
batch := paint.NewCommandBatch()
batch.Add(0, 0, "Hello", style1)
batch.Add(5, 0, " ", style1)      // 会被合并为 "Hello "
batch.Add(6, 0, "World", style2)
output := batch.Flush() // 生成优化的 ANSI 输出
```

### 样式状态机 (StyleStateMachine)

减少 ANSI 样式切换的开销：

```go
// 只在样式变化时发送代码
if ssm.NeedsUpdate(newStyle) {
    output.WriteString(ssm.Update(newStyle))
}
```

## 使用示例

### 基本 Buffer 操作

```go
import "github.com/wwsheng009/mint/runtime/paint"

// 创建 buffer
buf := paint.NewBuffer(80, 24)

// 设置单个字符
buf.SetCell(0, 0, 'H', style)

// 设置字符串（支持宽字符）
buf.SetString(0, 0, "你好，世界！", style)

// 绘制 emoji
buf.SetString(20, 0, "👨‍👩‍👧", style)

// 填充矩形
buf.Fill(paint.Rect{0, 0, 10, 5}, ' ', backgroundStyle)
```

### 获取 Buffer 输出

```go
// 获取 ANSI 输出（用于终端渲染）
output := buf.String() // 包含 ANSI 样式代码

// 直接输出到终端
fmt.Print(output)
```

### 使用 CommandBatch

```go
batch := paint.NewCommandBatch()

// 添加绘制命令
batch.Add(0, 0, "标题", titleStyle)
batch.Add(0, 1, "行1", normalStyle)
batch.Add(0, 2, "行2", normalStyle)

// 获取优化后的输出
output := batch.Flush()
```

### 图层合成

```go
compositor := paint.NewCompositor(80, 24)

// 添加图层
compositor.AddLayer(&paint.Layer{
    ID:     "background",
    Type:   paint.LayerNormal,
    ZIndex: 0,
    Rect:   paint.Rect{0, 0, 80, 24},
    Buffer: bgBuffer,
})

compositor.AddLayer(&paint.Layer{
    ID:     "foreground",
    Type:   paint.LayerOverlay,
    ZIndex: 1,
    Rect:   paint.Rect{10, 5, 60, 15},
    Buffer: fgBuffer,
})

// 合成所有图层
finalBuffer := compositor.Composite()
output := finalBuffer.String()
```

### 选择高亮

```go
// 设置选中单元格
buf.SetSelected(5, 5, true)
buf.SetSelected(6, 5, true)
buf.SetSelected(7, 5, true)

// 输出时会应用反色（\x1b[7m）
output := buf.String()

// 清除选择
buf.ClearSelection()
```

### 缓冲池（性能优化）

```go
// 重用 buffer 对象
buf := paint.NewBuffer(80, 24)

// 绘制
for _, item := range items {
    buf.SetString(x, y, item.Text, style)
}

// 输出
fmt.Print(buf.String())

// 重置（而非 NewBuffer，避免分配）
buf.Reset(80, 24)
```

## 核心类型

| 类型 | 说明 |
|------|------|
| `Buffer` | 绘制缓冲区，Cell 的二维网格 |
| `Cell` | 单个单元格，包含内容、样式和元数据 |
| `Rect` | 矩形区域（X, Y, Width, Height） |
| `CommandBatch` | 绘制命令批处理器 |
| `DrawCmd` | 单个绘制命令 |
| `Compositor` | 图层合成器 |
| `Layer` | 绘制层级 |
| `StyleStateMachine` | 样式状态机 |

## 文件结构

| 文件 | 说明 |
|------|------|
| `buffer.go` | Buffer 核心实现，宽字符处理 |
| `cell.go` | Cell 类型定义和管理 |
| `batch.go` | CommandBatch 实现，命令批处理 |
| `compositor.go` | 图层合成和混合 |
| `layer.go` | Layer 类型和管理 |
| `painter.go` | Painter 接口定义 |
| `context.go` | 绘制上下文（如果存在） |
| `dirty.go` | 脏标记和差异检测 |
| `remote.go` | 远程绘制支持（DevTools 集成） |
| `style_state.go` | 样式状态机实现 |
| `renderer.go` | 渲染器实现（如果存在） |

## 最佳实践

### 1. 使用 SetString 而非循环 SetCell

```go
// 推荐（高效）：
buf.SetString(0, 0, "你好世界", style)

// 不推荐（低效）：
for _, ch := range "你好世界" {
    buf.SetCell(x++, 0, ch, style) // 不正确处理宽字符
}
```

### 2. 性能优化：重用 Buffer

```go
// 正确：重用 buffer
buf.Reset(width, height)

// 错误：每次分配
buf = paint.NewBuffer(width, height) // 产生垃圾回收
```

### 3. 使用 CommandBatch 减少终端 IO

```go
// 对于大量小绘制命令
batch := paint.NewCommandBatch()
for _, ch := range characters {
    batch.Add(x, y, string(ch), style)
}
fmt.Print(batch.Flush()) // 优化后输出
```

### 4. 宽字符边界检查

```go
// 检查宽字符是否会超出边界
width := runewidth.StringWidth(text)
if x+width > buffer.Width {
    // 处理边界情况
    break
}
```

### 5. 脏标记优化

只重绘变化的区域：

```go
if !layer.IsDirty() {
    continue // 跳过干净的图层
}
// 绘制图层
layer.ClearDirty()
```

## 常见问题

### Q: 宽字符显示错乱？

确保：
1. 使用 `SetString` 而非 `SetCell` 绘制宽字符
2. 不要手动设置 `IsContinuation` 单元格
3. 清除时使用 `ClearWideChar` 或让 `clearCellAt` 处理

### Q: Emoji 显示为问号或分成多个字符？

使用 `SetString`，它会正确处理字形簇：

```go
buf.SetString(0, 0, "👨‍👩‍👧", style) // 正确
buf.SetCell(0, 0, '👨', style) // 错误
```

### Q: ANSI 样式代码太多，性能差？

使用 `CommandBatch` 进行命令批处理和样式去重：

```go
batch := paint.NewCommandBatch()
// 添加命令
output := batch.Flush() // 已优化
```

### Q: 如何实现透明背景？

使用 Z-Index 和图层合成：

```go
// 下层：背景
compositor.AddLayer(bgLayer)

// 上层：只绘制内容，不绘制背景
compositor.AddLayer(contentLayer)
```

## 相关文档

- `BUGFIX_diff_rendering.md`：宽字符差异检测修复详解
- `WIDE_CHAR_GUIDE.md`：宽字符处理完整指南
