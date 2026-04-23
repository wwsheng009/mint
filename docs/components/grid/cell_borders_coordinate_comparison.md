# 边框绘制与布局引擎坐标对比分析

## 1. 边框绘制的 X 坐标计算

```go
// ui/components/grid/cell_borders_paint.go
for row := 0; row <= numRows; row++ {
    for col := 0; col <= numCols; col++ {
        x := contentX
        for c := 0; c < col; c++ {
            x += inst.colWidths[c] + 1  // 内容宽度 + 右边框
            if c < col-1 {
                x += inst.columnGap
            }
        }

        // 边框字符位置
        // col=0: x = contentX + 0
        // col=1: x = contentX + colWidths[0] + 1
        // col=2: x = contentX + colWidths[0] + 1 + colWidths[1] + 1
    }
}
```

## 2. 布局引擎的 getCellPosition X 坐标计算

```go
// runtime/layout/grid.go
func (g *GridLayout) getCellPosition(row, col int) (x, y int) {
    if g.style.ShowCellBorders {
        x = 0
        for c := 0; c < col; c++ {
            x += g.colWidths[c] + 1  // 内容宽度 + 右边框
            if c < col-1 {
                x += g.style.ColumnGap  // 列间距
            }
        }
        x += 1  // ✅ 跳过上边框！
    }
    return x
}
```

## 3. 坐标对比（以 Demo 9 为例）

**配置：**
- padding: [0, 0, 0, 0]
- originX, originY: (0, 0)
- contentX = 0, contentY = 0
- 列宽：每列 10 字符，共 3 列
- 行高：每行 1 字符，共 2 行

### 3.1 边框绘制位置（边框字符）

```go
// Row 0（上边框线）
col=0: x = 0      → ┌ (左上角)          ← 边框字符在 y=0
col=1: x = 0 + 10 + 1 = 11  → ┬        ← 边框字符在 y=0
col=2: x = 0 + 10 + 1 + 10 + 1 = 22  → ┬ ← 边框字符在 y=0
col=3: x = 0 + 10 + 1 + 10 + 1 + 10 = 32  → ┐ (右上角) ← 边框字符在 y=0
```

**注意：** 边框字符在第 0 行（`y = contentY + 0 = 0`）

### 3.2 内容区域位置（子节点）

```go
// Row 0 的内容区域
Grid bounds[0] = 0, Grid padding[3] = 0
子节点 (0, 0):
    相对 X = 0 + 1 = 1  (从 getCellPosition 获得)
    绝对 X = 0 + 0 + 1 = 1  ← 子节点文本从 x=1 开始

    相对 Y = 0 + 1 = 1  (从 getCellPosition 获得)
    绝对 Y = 0 + 0 + 1 = 1  ← 子节点文本从 y=1 开始

子节点 (0, 1):
    相对 X = 1 + 10 + 1 = 12  (从 getCellPosition 获得)
    绝对 X = 0 + 0 + 12 = 12
```

## 4. 问题发现：Y 坐标冲突！

### 4.1 边框在第 0 行

```go
// 边框绘制的第一行（row=0，即上边框）
y := contentY  // = 0
// ┌────┬────┬────┐  ← y=0
// 在 y=0 上绘制边框字符
```

### 4.2 内容区域从第 1 行开始

```go
// 子节点从 getCellPosition(0, 0) 获得 Y
y = 0 + 1  // = 1
//      ID     Name     Role  ← y=1
// 子节点文本从 y=1 开始渲染
```

### 4.3 问题场景

**如果子节点的绝对坐标计算少了 +1：**

```go
// 错误的坐标转换（少了 padding[0] 的 +1？
错误的绝对 Y = Grid.bounds[1] + relX  // 应该是 + relY
              = 0 + 1  // 如果 relX=1（错误！）则 Y=1

// 但如果：
错误的绝对 Y = Grid.bounds[1] + 0  // 跳过了 Y 偏移
              = 0 + 0 = 0  ← ❌ 子节点在 y=0 上绘制！
```

**这就是问题所在！**

如果子节点的绝对坐标计算时，没有正确使用 `getCellPosition()` 返回的**已经加过 +1 的相对坐标**，或者直接使用了布局引擎返回的 LayoutBox 的 X 和 Y（这些是相对于内容区域的），那么：

1. 子节点会从 `y=0` 开始渲染
2. 但边框字符也在 `y=0` 上
3. 导致子节点内容覆盖边框字符

## 5. 布局引擎的 LayoutBox 存的是什么？

```go
// runtime/layout/grid.go
func (g *GridLayout) LayoutChildren(width, height int) []LayoutBox {
    boxes := make([]LayoutBox, 0)
    for i, child := range g.children {
        x, y := g.getCellPosition(row, col)  // ✅ 返回的是跳过边框的相对坐标

        box := LayoutBox{
            ID:     child.ID(),
            X:      x,  // ✅ 已跳过上边框（例如 0+1=1）
            Y:      y,  // ✅ 已跳过左边框（例如 0+1=1）
            Width:  w,
            Height: h,
        }
        boxes = append(boxes, box)

        child.SetPosition(x, y)  // ✅ 设置相对坐标（已跳过边框）
    }
    return boxes
}
```

**LayoutBox.X 和 LayoutBox.Y 已经是跳过边框的相对坐标！**

## 6. 渲染引擎应该如何转换？

```go
// ✅ 正确的转换逻辑
绝对 X = Grid.bounds[0] + Grid.padding[3] + LayoutBox.X
绝对 Y = Grid.bounds[1] + Grid.padding[0] + LayoutBox.Y

// 例如：
// Grid.bounds = [0, 13, 34, 5]
// Grid.padding = [0, 0, 0, 0]
// LayoutBox.X = 1  (从 getCellPosition(0, 0) 获得，已经 +1)
// LayoutBox.Y = 1  (从 getCellPosition(0, 0) 获得，已经 +1)
//
// 绝对 X = 0 + 0 + 1 = 1 ✅
// 绝对 Y = 13 + 0 + 1 = 14 ✅
```

```go
// ❌ 错误的转换逻辑（缺少 LayoutBox.X, Y）
绝对 X = Grid.bounds[0] + Grid.padding[3]
绝对 Y = Grid.bounds[1] + Grid.padding[0]

// 例如：
// 绝对 X = 0 + 0 = 0  ← ❌ 少了 +1，子节点会在 y=0 上绘制
// 绝对 Y = 13 + 0 = 13  ← ❌ 少了 +1，子节点会在边框上绘制
```

## 7. 从 debug 输出验证

从 `grid_borders_output.txt` 的 Demo 9 第 13 行起：

```
[DEBUG GRID SETBOUNDS] x=0, y=13, w=0, h=0
[DEBUG GRID SETBOUNDS] numRows=2, cellBorderHeight=3, availableH=0, oldRowHeights=[1 1], newRowHeights=[1 1]
[DEBUG GRID SETBOUNDS] x=0, y=0, w=34, h=5
[DEBUG GRID SETBOUNDS] numRows=2, cellBorderHeight=3, availableH=2, oldRowHeights=[1 1], newRowHeights=[1 1]
```

注意：`Grid.bounds[1] = 13`，这意味着：
- Grid 组件在 y=13 处
- Grid 的内容区域起点在 `y = 13 + 0 + 0 = 13`
- 边框字符在 `y = 13 + 0 = 13`（上边框）
- 子节点应该在 `y = 13 + 0 + 1 = 14`（内容区域起点）

但如果子节点的绝对 Y 被错误计算为 13，就会覆盖边框。

## 8. 关键问题：谁调用 child.SetPosition()？

```go
// runtime/layout/grid.go
child.SetPosition(x, y)  // 调用 layout.Node.SetPosition()
```

这会调用到：
```go
// internal/render/fiber_adapter.go
func (a *FiberToNodeAdapter) SetPosition(x, y int) {
    // Store in fiber.ComputedBox
    a.fiber.ComputedBox = &compute.ComputedBox{
        Box: compute.Box{X: x, Y: y},  // ✅ 存储的已经是正确的相对坐标（已跳过边框）
    }

    // ✅ FIX: Also sync to Instance.bounds
    if a.fiber.Instance != nil {
        // ...
        if boundsHaver, ok := a.fiber.Instance.(interface{ SetBounds(x, y, w, h int) }); ok {
            boundsHaver.SetBounds(x, y, w, h)  // ⚠️ 这里使用的是相对坐标 x, y
        }
    }
}
```

**问题：** `Instance.SetBounds()` 接收的是**相对坐标**还是**绝对坐标**？

如果 `SetBounds()` 期望的是绝对坐标，但收到的是相对坐标，就会有问题。

但实际上，从 `FiberToNodeAdapter.SetPosition()` 的代码来看：
- `ComputedBox.Box.X` 存储的是相对坐标
- `Instance.SetBounds()` 也接收的是相对坐标
- 子节点最终绘制时，需要自己知道自己在父组件中的相对位置

**真正的问题在于：子节点的 `Paint()` 方法。**

让我们检查 Text 组件的 `Paint()` 方法。

## 9. 结论

**问题根源：**

1. ✅ 布局引擎正确计算了子节点的相对坐标（已跳过边框）
2. ✅ 边框绘制正确计算了边框字符的位置（在边框线上）
3. ⚠️ 子节点的 `Paint()` 方法接收的 `(x, y)` 参数没有正确转换为绝对坐标

**具体场景：**
- 边框绘制在 `y = originY + contentY + row累积 = 13 + 0 + 0 = 13`
- 子节点应该绘制在 `y = 13 + 0 + 1 = 14`（相对于 Grid 边界）
- 但如果子节点的 `Paint(x, y)` 接收的 `y = 13`，就会在边框上绘制

**需要检查：**
1. Text 组件的 `Paint()` 方法接收的 x, y 参数是什么
2. 渲染引擎如何将 Grid 的相对坐标转换为子节点的绝对坐标
3. 这个转换逻辑是否正确
