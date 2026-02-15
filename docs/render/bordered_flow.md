# Bordered 组件渲染流程

## 组件结构

```
Bordered (边框容器)
├── BorderNode (runtime/ui/layout.go)
│   ├── BorderStyle: 边框样式 (Single/Double/Rounded/Dashed)
│   ├── BorderColor: 边框颜色
│   ├── BorderLabel: 边框标签
│   └── Children[0]: 子内容 (VStack/Text/etc.)
```

## 测量阶段

### BorderNode.Measure()

位置: `runtime/ui/layout.go:1151-1276`

```go
func (bn *BorderedNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // 1. 计算边框宽度 (左右各1个字符)
    borderWidth := 0
    borderHeight := 0
    if bn.borderStyle != BorderNone {
        borderWidth = 2  // 左边1个字符 + 右边1个字符
        borderHeight = 2  // 上边1个字符 + 下边1个字符
    }

    // 2. 测量子内容尺寸
    var contentWidth, contentHeight int
    children := bn.Children()

    if len(children) > 0 {
        child := children[0]

        // 2a. 如果子节点实现了 Measurable 接口
        if measurable, ok := child.(interface {
            Measure(constraints runtime.BoxConstraints) runtime.Size
        }); ok {
            // 使用子节点的 Measure 方法
            innerConstraints := constraints.SubtractPadding(borderWidth, borderHeight)
            contentSize := measurable.Measure(innerConstraints)
            contentWidth = contentSize.Width
            contentHeight = contentSize.Height
        } else {
            // 2b. Fallback: 估计子节点大小
            contentWidth = 10  // 默认最小宽度
            contentHeight = 1  // 默认最小高度
        }
    }

    // 3. 计算总尺寸 = 内容 + 边框
    innerWidth := contentWidth
    if labelWidth > innerWidth {
        innerWidth = labelWidth
    }

    totalWidth := innerWidth + borderWidth
    totalHeight := contentHeight + borderHeight

    return runtime.Size{Width: totalWidth, Height: totalHeight}
}
```

### 关键问题

**问题**: 当子节点是 VStack 时，fallback 逻辑返回 `contentHeight = 1`，而不是实际的 VStack 高度（如3）。

**原因**:
- VStack 是 LayoutNode，实现了 `Measure(constraints BoxConstraints) Size` 接口
- 但 fallback 路径检查的是 `Measurable` 接口
- LayoutNode 虽然实现了 Measurable，但其 Measure 方法对 VStack 返回的是内容高度，不包括边框

**影响**:
- `contentHeight = 1` 导致边框中间行不渲染
- 边框只有上下边框线，左右竖线缺失
- 子内容（VStack 的多行文本）被压缩到一行

## 绘制阶段

### paintBordered()

位置: `internal/render/declarative_node.go:656-722`

```go
func (n *DeclarativeNode) paintBordered(
    vnode rtui.VNode,
    _ interface{ RenderBorder(int, int) []rtui.VNode },
    x, y int,
    buf *paint.Buffer,
) {
    children := vnode.Children()
    if len(children) == 0 {
        return
    }

    child := children[0]
    if child == nil {
        return
    }

    // 测量内容尺寸
    contentWidth := n.MeasureVNodeWidth(child)
    contentHeight := n.MeasureVNodeHeight(child)

    // 获取边框配置
    config := border.DefaultConfig()

    // 设置边框样式
    if bn, ok := vnode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
        config.Style = border.Style(bn.GetBorderStyle())
    }
    if bn, ok := vnode.(interface{ GetBorderColor() string }); ok {
        config.Color = bn.GetBorderColor()
    }
    if bn, ok := vnode.(interface{ GetBorderLabel() string }); ok {
        config.Label = bn.GetBorderLabel()
    }

    // 创建边框渲染器
    renderer := border.WithConfig(config)

    // 绘制边框单元格
    renderer.Paint(x, y, contentWidth, contentHeight, func(px, py int, ch rune, s style.Style) {
        buf.SetCell(px, py, ch, s)
    })

    // 绘制边框内的内容
    offsetX, offsetY := renderer.GetContentOffset()
    n.PaintVNode(child, x+offsetX, y+offsetY, buf)
}
```

### MeasureVNodeHeight(child)

位置: `internal/render/declarative_node.go:830-883`

```go
func (n *DeclarativeNode) MeasureVNodeHeight(vnode rtui.VNode) int {
    if vnode == nil {
        return 0
    }

    // PRIORITY 1: 使用 Measurable 接口（如果可用）
    if measurable, ok := vnode.(interface {
        Measure(constraints runtime.BoxConstraints) runtime.Size
    }); ok {
        size := measurable.Measure(runtime.UnboundedConstraints())
        return size.Height
    }

    // PRIORITY 2: 检查 VStack（垂直堆叠）- 求总高度
    if tagger, ok := vnode.(interface{ Tag() string }); ok {
        tag = tagger.Tag()
    }
    if tag == "vstack" {
        totalHeight := 0
        children := vnode.Children()
        for _, child := range children {
            totalHeight += n.MeasureVNodeHeight(child)
        }
        return totalHeight
    }

    // PRIORITY 3: 检查 HStack（水平堆叠）- 取最大高度
    if tag == "hstack" {
        maxHeight := 0
        children := vnode.Children()
        for _, child := range children {
            childHeight := n.MeasureVNodeHeight(child)
            if childHeight > maxHeight {
                maxHeight = childHeight
            }
        }
        return maxHeight
    }

    // ... 其他特殊情况处理 ...

    // Fallback: 默认高度
    return 1
}
```

## 边框渲染器

### Renderer.Render()

位置: `runtime/border/border.go:150-220`

```go
func (r *Renderer) Render(x, y int, contentWidth, contentHeight int) []Cell {
    var cells []Cell

    // === 上边框 ===
    // 左上角
    cells = append(cells, Cell{cornerTL, x, y, borderStyle})
    // 上边框线
    for i := 0; i < innerWidth; i++ {
        cells = append(cells, Cell{horizontal, x + 1 + i, y, borderStyle})
    }
    // 右上角
    cells = append(cells, Cell{cornerTR, x + innerWidth + 1, y, borderStyle})

    // === 中间行（左边框 + 内容区域 + 右边框）===
    // 这个循环根据 contentHeight 迭行
    for row := 0; row < contentHeight; row++ {
        rowY := y + 1 + row

        // 左边框
        cells = append(cells, Cell{vertical, x, rowY, borderStyle})

        // 右边框 (at: x + innerWidth + 1, rowY, borderStyle)
        cells = append(cells, Cell{vertical, x + innerWidth + 1, rowY, borderStyle})
    }

    // === 下边框 ===
    bottomY := y + 1 + contentHeight
    cells = append(cells, Cell{cornerBL, x, bottomY, borderStyle})
    for i := 0; i < innerWidth; i++ {
        cells = append(cells, Cell{horizontal, x + 1 + i, bottomY, borderStyle})
    }
    cells = append(cells, Cell{cornerBR, x + innerWidth + 1, bottomY, borderStyle})

    return cells
}
```

## 渲染流程图

```
Bordered(VStack("Line1", "Line2", "Line3"))
    │
    ├─ Measure()
    │   ├─ 子节点 Measure() → VStack.Measure()
    │   │   ├─ VStack 高度 = 3 (三行文本)
    │   │   └─ 返回 Size{Width: 6, Height: 3}
    │   ├─ 加上边框: Width = 6+2 = 8, Height = 3+2 = 5
    │   └─ 返回 Size{Width: 8, Height: 5}
    │
    ├─ Paint()
    │   ├─ paintBordered()
    │   │   ├─ MeasureVNodeWidth(VStack) → 6
    │   │   ├─ MeasureVNodeHeight(VStack)
    │   │   │   ├─ 如果 tag == "vstack": 累加子节点高度
    │   │   │   ├─ Line1: 1 + Line2: 1 + Line3: 1 = 3
    │   │   │   └─ 返回 3
    │   │   ├─ Border.Render(0, 0, 6, 3)
    │   │   │   ├─ 上边框: ┌───┐ (y=0)
    │   │   │   ├─ 中间行: │ 内容 │ (y=1, 2, 3)
    │   │   │   │   ├─ 左竖线 │
    │   │   │   │   └─ 右竖线 │
    │   │   │   └─ 下边框: └───┘ (y=4)
    │   │   │
    │   │   └─ PaintVNode(VStack, 1, 1, buf)
    │   │      └─ VStack 逐行绘制 "Line1", "Line2", "Line3"
    │   └─ 输出: 完整边框 + 内容
```

## 问题分析

### 问题表现

测试 `TestBorderedSimple` 的输出：
```
┌───┐
└───┘
```

- ✓ 上边框正确 (┌───┐)
- ✓ 下边框正确 (└───┘)
- ✗ 中间行缺失 (没有 │ Hello │)
- ✗ 内容缺失 ("Hello" 没有显示)

### 根本原因

1. **MeasureVNodeHeight 对 VStack 返回 1**
   - `MeasureVNodeHeight` 检查 `tag == "vstack"` 的逻辑缺失
   - 导致 `contentHeight = 1`
   - 边框渲染器只创建 1 行中间单元格（或不创建）

2. **fallback 路径返回默认高度**
   - `BorderNode.Measure` 的 fallback 返回 `contentHeight = 1`
   - 即使 VStack 的 Measure 返回 3，也没被正确传递

### 修复方案

**方案 1**: 在 `MeasureVNodeHeight` 中添加 VStack 特殊处理

```go
// PRIORITY 2: 检查 VStack（垂直堆叠）- 求总高度
if tagger, ok := vnode.(interface{ Tag() string }); ok {
    tag = tagger.Tag()
}
if tag == "vstack" {
    totalHeight := 0
    children := vnode.Children()
    for _, child := range children {
        totalHeight += n.MeasureVNodeHeight(child)
    }
    return totalHeight
}
```

**方案 2**: 在 `BorderNode.Measure` 中使用 LayoutNode 的 MeasureLayout 方法

```go
// 检查子节点是否是 LayoutNode
if layoutNode, ok := child.(interface {
    MeasureLayout(runtime.ChildMeasurer, runtime.BoxConstraints) runtime.LayoutMeasurement
}); ok {
    measurement := layoutNode.MeasureLayout(nil, constraints)
    contentWidth = measurement.Size.Width
    contentHeight = measurement.Size.Height
}
```

## 修复位置

需要修改的文件：
1. `internal/render/declarative_node.go` - `MeasureVNodeHeight` 方法
2. `runtime/ui/layout.go` - `BorderedNode.Measure` 方法（可选）
