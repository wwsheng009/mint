# Absolute 示例中 zIndex 的作用

## 示例代码

```go
app.VStack(
    app.Text("Background layer"),
    app.HStack(
        app.Text("Middle layer"),
        app.AbsoluteBuilder(
            app.NewTextBuilder("OVERLAY").FgColor("white").BgColor("red").Build(),
        ).
            Left(absolute.AbsolutePos(10)).
            Top(absolute.AbsolutePos(5)).
            ZIndex(10).  // ← 关键： zIndex = 10
            Build(),
    ),
)
```

## zIndex 作用

`zIndex` 控制元素的**渲染顺序**，数值越大的元素会覆盖数值较小的元素。

### 工作原理

#### 1. Cell 级别的 Z-Index 控制

在 `runtime/paint/buffer.go` 中：

```go
func (b *Buffer) SetContent(x, y, z int, char rune, s style.Style, nodeID string) {
    // Check Z-Index - only overwrite if new Z is >= existing Z
    if z < b.Cells[y][x].ZIndex {
        return  // ZIndex 低，不覆盖
    }

    b.Cells[y][x] = Cell{
        Cluster: string(char),
        Style:   s,
        ZIndex:  z,  // 使用 ZIndex
        NodeID:  nodeID,
    }
}
```

**规则**：只有 `newZIndex >= existingZIndex` 时才会覆盖已存在的内容。

#### 2. 计算机显示位置

在终端中，"Middle layer" 和 "OVERLAY" 的位置有重叠：

```
Terminal Layout:
┌──────────────────────────────────────┐
│  Background layer                    │  ← Line 0
│                                      │
│  Middle layer OVERLAY                │  ← Line 2 (HStack)
│  ↑↑↑↑↑↑↑↑↑  ↑↑↑↑↑↑↑↑  ↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑  │
│   0123456789   10      17...       │
│              相对偏移 10            │
└──────────────────────────────────────┘
```

- "Middle layer" 从 x=0 开始，长度约 13 个字符
- "OVERLAY" 从 HStack 的 x=0 + 偏移 10 开始
- **位置有重叠**，都是 x=10 左右的区域

### 场景分析

#### 场景 1：不设置 ZIndex（默认值）

```go
// 不设置 ZIndex，默认为 0
app.AbsoluteBuilder(...).
    Left(absolute.AbsolutePos(10)).
    Top(absolute.AbsolutePos(5)).
    Build(),  // 没有 ZIndex 调用
```

**结果**：
- "Middle layer" → zIndex = 0
- "OVERLAY" → zIndex = 0

在渲染时：
1. "Middle layer" 先渲染到 Buffer，设置 Cells 的 ZIndex = 0
2. "OVERLAY" 后渲染，设置 Cells 的 ZIndex = 0
3. `SetContent()` 检查：`0 < 0` 为 false
4. 等价于 `SetContent(x, y, 0, ...)` 覆盖 `SetContent(x, y, 0, ...)`
5. **OVERLAY 会覆盖 "Middle layer"**（因为 0 >= 0）

#### 场景 2：设置 ZIndex(10)

```go
app.AbsoluteBuilder(...).
    Left(absolute.AbsolutePos(10)).
    Top(absolute.AbsolutePos(5)).
    ZIndex(10).  // ← 设置为 10
    Build(),
```

**结果**：
- "Middle layer" → zIndex = 0
- "OVERLAY" → zIndex = 10

在渲染时：
1. "Middle layer" 先渲染到 Buffer，设置 Cells 的 ZIndex = 0
2. "OVERLAY" 后渲染，设置 Cells 的 ZIndex = 10
3. `SetContent()` 检查：`10 >= 0` 为 true
4. **OVERLAY 会覆盖 "Middle layer"**（因为 10 >= 0）

#### 场景 3：如果 "Middle layer" 也有 ZIndex

```go
app.HStack(
    app.Text("Middle layer").ZIndex(5),  // 假设有这个方法
    app.AbsoluteBuilder(...).
        ZIndex(10).
        Build(),
)
```

**结果**：
- "Middle layer" → zIndex = 5
- "OVERLAY" → zIndex = 10
- `10 >= 5` → OVERLAY 仍然覆盖（因为 10 >= 5）

如果反过来：

```go
app.HStack(
    app.Text("Middle layer").ZIndex(15),
    app.AbsoluteBuilder(...).
        ZIndex(10).
        Build(),
)
```

**结果**：
- "Middle layer" → zIndex = 15
- "OVERLAY" → zIndex = 10
- `10 < 15` → OVERLAY **不会**覆盖 "Middle layer"

### 在这个例子中为什么需要 ZIndex(10)？

实际上，在这个简单的例子中：

```go
app.VStack(
    app.Text("Background layer"),
    app.HStack(
        app.Text("Middle layer"),
        app.AbsoluteBuilder(
            app.NewTextBuilder("OVERLAY").Build(),
        ).
            Left(absolute.AbsolutePos(10)).
            Top(absolute.AbsolutePos(5)).
            ZIndex(10).  // ← 这里
            Build(),
    ),
)
```

**如果 ZIndex 默认为 0**，"OVERLAY" 仍然会覆盖 "Middle layer"（因为先渲染 Middle，后渲染 OVERLAY）。

**设置 ZIndex(10) 的意义**：
1. **明确意图**：明确表示这是一个覆盖层
2. **多层叠加**：如果有多个覆盖层，可以控制它们的显示顺序
3. **层级关系**：建立清晰的层级语义

#### 更复杂的场景示例

```go
app.HStack(
    // 底层
    absolute.NewBuilder(app.Text("Base")).
        Left(0).Top(0).
        ZIndex(0).Build(),

    // 中间层
    absolute.NewBuilder(app.Text("Middle")).
        Left(5).Top(5).
        ZIndex(5).Build(),

    // 顶层
    absolute.NewBuilder(app.Text("Top")).
        Left(10).Top(10).
        ZIndex(10).Build(),
)
```

**显示顺序**：Top (10) > Middle (5) > Base (0)

### ZIndex 值的建议

| 值 | 用途 |
|----|------|
| 0 | 默认元素（文本、按钮等） |
| 1-9 | 次级覆盖层（提示、标签） |
| 10-99 | 主要覆盖层（弹窗、Overlay） |
| 100+ | 顶级覆盖层（模态框、通知） |

### 与 CSS 的对比

Web/CSS 中的 `z-index` 与 TUI 中的 `zIndex` 非常相似：

```css
/* CSS */
.overlay {
    position: absolute;
    left: 10px;
    top: 5px;
    z-index: 10;  /* ← 相同 */
}
```

```go
/* Go TUI */
absolute.NewBuilder(...).
    Left(absolute.AbsolutePos(10)).
    Top(absolute.AbsolutePos(5)).
    ZIndex(10).  /* ← 相同 */
    Build()
```

## 总结

在 `examples/absolute/main.go` 中：

1. **ZIndex(10)** 确保 "OVERLAY" 渲染在其他元素之上
2. 如果不设置或设置为 0，行为仍然类似（因为有渲染顺序）
3. ZIndex 的真正作用是在**多层叠加**时控制层级关系
4. 对于单个覆盖层示例，ZIndex 主要起到"明确意图"的作用
