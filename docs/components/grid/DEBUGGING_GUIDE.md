# Grid 组件调试指南

本文档提供 Grid 组件开发中的实际调试方法和布局数据跟踪技巧。

---

## 目录

1. [快速开始](#1-快速开始)
2. [调试工具](#2-调试工具)
3. [布局数据跟踪](#3-布局数据跟踪)
4. [常见问题定位](#4-常见问题定位)
5. [可视化调试](#5-可视化调试)
6. [实战案例](#6-实战案例)
7. [调试技巧总结](#7-调试技巧总结)

---

## 1. 快速开始

### 1.1 最简单的调试方式

在 `Instance.Measure()` 中添加打印：

```go
// ui/components/grid/instance.go

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    fmt.Printf("[Grid] Measure called: constraints=%v\n", constraints)

    size := inst.computeLayout(constraints)

    fmt.Printf("[Grid] Measure result: size=%v\n", size)
    return size
}
```

### 1.2 使用追踪系统

```go
func runWithTrace() {
    // 启用追踪
    layout.EnableTracer()

    // 运行渲染
    renderGrid()

    // 打印追踪日志
    fmt.Println(layout.DumpTrace())
}
```

---

## 2. 调试工具

### 2.1 列表

| 工具 | 用途 | 使用场景 |
|------|------|-----------|
| **Printf/Fmt** | 快速打印变量值 | 开发阶段快速验证 |
| **Tracer** | 约束传播追踪 | 布局问题诊断 |
| **DebugLog** | 结构化日志 | 生产环境调试 |
| **可视化面板** | 实时查看布局 | UI 渲染问题 |
| **Go Delve** | 断点调试 | 复杂逻辑分析 |

### 2.2 格式化打印辅助函数

```go
// ui/components/grid/debug.go

package grid

import (
    "fmt"
    "log"

    "github.com/wwsheng009/mint/runtime/layout"
)

// DebugMode 控制是否启用调试日志
var DebugMode = false

// DebugPrintf 条件打印（仅在 DebugMode=true 时输出）
func DebugPrintf(format string, args ...interface{}) {
    if DebugMode {
        log.Printf("[Grid-Debug] "+format, args...)
    }
}

// FormatConstraints 格式化约束
func FormatConstraints(c layout.Constraints) string {
    return fmt.Sprintf("W:[%d..%d] H:[%d..%d]",
        c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

// FormatSize 格式化尺寸
func FormatSize(s layout.Size) string {
    return fmt.Sprintf("%dx%d", s.Width, s.Height)
}

// FormatDimensions 格式化 Grid 的列/行尺寸
func FormatDimensions(colWidths, rowHeights []int, colGap, rowGap int) string {
    cols := fmt.Sprintf("Cols(%d): %v (gap=%d)",
        len(colWidths), colWidths, colGap)
    rows := fmt.Sprintf("Rows(%d): %v (gap=%d)",
        len(rowHeights), rowHeights, rowGap)
    return cols + "\n" + rows
}

// PrintMeasurementInfo 打印测量信息
func PrintMeasurementInfo(gridId string, constraints, output layout.Constraints, size layout.Size) {
    DebugPrintf("Grid[%s] Measure:\n", gridId)
    DebugPrintf("  Input:  %s\n", FormatConstraints(constraints))
    DebugPrintf("  Output: %s\n", FormatConstraints(output))
    DebugPrintf("  Size:   %s\n", FormatSize(size))
}
```

---

## 3. 布局数据跟踪

### 3.1 跟踪层级

```
┌─────────────────────────────────────────────────────┐
│ Level 1: 应用层 (入口追踪)                           │
│ - 查看从父组件传递来的约束                           │
├─────────────────────────────────────────────────────┤
│ Level 2: 实例层 (实例级追踪)                         │
│ - 查看 Grid 接收的约束和返回的结果尺寸               │
├─────────────────────────────────────────────────────┤
│ Level 3: 计算层 (列/行追踪)                         │
│ - 查看每列/每行的计算过程                           │
├─────────────────────────────────────────────────────┤
│ Level 4: 子节点层 (子项追踪)                         │
│ - 查看每个子格子的约束传递和测量结果                 │
└─────────────────────────────────────────────────────┘
```

### 3.2 Level 1: 应用层追踪

```go
// 在 fiber-first 渲染流程入口处添加

import (
    "github.com/wwsheng009/mint/runtime/layout"
)

func renderFiberGrid() {
    // 启用追踪
    layout.EnableTracer()
    defer layout.DisableTracer()

    // 定义约束（模拟父容器传递）
    constraints := layout.NewConstraints(0, 120, 0, 40)

    log.Printf("[Render] Starting grid render with constraints: %v", constraints)

    // 执行渲染...

    // 输出追踪
    traceLogs := layout.DumpTrace()
    log.Printf("[Render] \n%s", traceLogs)
}
```

### 3.3 Level 2: 实例层追踪

```go
// ui/components/grid/instance.go

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    gridId := fmt.Sprintf("grid-%d", inst.fiberId)

    DebugPrintf("=== Level 2: Instance Measure ===")
    DebugPrintf("Grid[%s] Input Constraints: %s",
        gridId, FormatConstraints(constraints))

    // 检查约束有效性
    if constraints.MinWidth > constraints.MaxWidth {
        DebugPrintf("⚠️ WARNING: MinWidth(%d) > MaxWidth(%d)",
            constraints.MinWidth, constraints.MaxWidth)
    }

    // ... 测量逻辑 ...

    DebugPrintf("Grid[%s] Output Size: %s",
        gridId, FormatSize(size))

    PrintMeasurementInfo(gridId, constraints, outputConstraints, size)

    return size
}
```

### 3.4 Level 3: 计算层追踪

```go
// runtime/layout/grid.go

func (g *GridLayout) calculateColumnWidths(availableWidth int) []int {
    layout.TraceMeasuring(
        g.id,
        "column-calculator",
        g.id+"/columns",
        layout.Constraints{MinWidth: 0, MaxWidth: availableWidth},
        layout.Constraints{},
        layout.Size{},
        "Column calculation start",
    )

    DebugPrintf("=== Level 3: Column Calculation ===")
    DebugPrintf("Grid[%s] Available width: %d", g.id, availableWidth)

    numCols := len(g.style.Columns)
    widths := make([]int, numCols)

    // Pass 1: 计算固定宽度
    DebugPrintf("--- Pass 1: Fixed widths ---")
    for i, col := range g.style.Columns {
        switch c := col.(type) {
        case GridFixed:
            widths[i] = int(c)
            DebugPrintf("  Col %d: Fixed = %d", i, widths[i])
        case GridMin:
            widths[i] = c.Min
            DebugPrintf("  Col %d: Min = %d", i, widths[i])
        // ... 其他类型
        }
    }

    // Pass 2: 分配 Flex 宽度
    DebugPrintf("--- Pass 2: Flex distribution ---")
    DebugPrintf("  Remaining width: %d", remainingWidth)
    DebugPrintf("  Flex count: %d, Total factor: %d",
        flexCount, flexTotalFactor)

    for i, width := range widths {
        colPath := fmt.Sprintf("%s/col-%d", g.id, i)

        layout.TraceMeasuring(
            fmt.Sprintf("col-%d", i),
            "computed",
            colPath,
            layout.Constraints{MinWidth: 0, MaxWidth: availableWidth},
            layout.Constraints{MinWidth: width, MaxWidth: width},
            layout.Size{Width: width, Height: 0},
            fmt.Sprintf("Column %d: %s", i, g.describeDimension(g.style.Columns[i])),
        )

        DebugPrintf("  Col %d: Final width = %d", i, width)
    }

    return widths
}

// 辅助函数：描述维度类型
func (g *GridLayout) describeDimension(dim GridDimension) string {
    switch d := dim.(type) {
    case GridFixed:
        return fmt.Sprintf("Fixed(%d)", d)
    case GridFlex:
        return fmt.Sprintf("Flex(factor=%d)", d.Factor)
    case GridAuto:
        return "Auto"
    default:
        return "Unknown"
    }
}
```

### 3.5 Level 4: 子节点追踪

```go
// ui/components/grid/instance.go (在遍历格子时)

func (inst *Instance) measureChildren(constraints layout.Constraints) {
    DebugPrintf("=== Level 4: Children Measurement ===")

    for i, cell := range inst.cells {
        cellId := fmt.Sprintf("cell-%d-%d", cell.Row, cell.Col)

        // 计算格子约束
        cellConstraints := inst.calculateCellConstraints(cell)
        DebugPrintf("Cell[%s]: %s", cellId, FormatConstraints(cellConstraints))

        // 测量子节点
        childInst := cell.Child.(*Instance)
        childSize := childInst.Measure(cellConstraints)

        DebugPrintf("Cell[%s] result: %s", cellId, FormatSize(childSize))

        // 记录子节点追踪
        childKey := fmt.Sprintf("grid-%d/%s", inst.fiberId, cellId)
        childOutput := layout.Constraints{
            MinWidth:  childSize.Width,
            MaxWidth:  childSize.Width,
            MinHeight: childSize.Height,
            MaxHeight: childSize.Height,
        }

        layout.TraceMeasuring(
            fmt.Sprintf("grid-%d", inst.fiberId),
            cellId,
            childKey,
            cellConstraints,
            childOutput,
            childSize,
            fmt.Sprintf("Cell [%d,%d] measurement", cell.Row, cell.Col),
        )
    }
}
```

### 3.6 完整追踪流程输出示例

```
[Grid-Debug] === Level 1: Application Entry ===
[Grid-Debug] Parent constraints: W:[0..120] H:[0..40]

[Grid-Debug] === Level 2: Instance Measure ===
[Grid-Debug] Grid[grid-123] Input Constraints: W:[0..120] H:[0..40]

[Grid-Debug] === Level 3: Column Calculation ===
[Grid-Debug] Grid[grid-123] Available width: 118 (after padding:2)
[Grid-Debug] --- Pass 1: Fixed widths ---
[Grid-Debug]   Col 0: Fixed = 20
[Grid-Debug] --- Pass 2: Flex distribution ---
[Grid-Debug]   Remaining width: 98
[Grid-Debug]   Flex count: 2, Total factor: 3
[Grid-Debug]   Col 1: Final width = 33
[Grid-Debug]   Col 2: Final width = 65

[Grid-Debug] === Level 3: Row Calculation ===
[Grid-Debug] Grid[grid-123] Available height: 38 (after padding:2)
[Grid-Debug]   Row 0: final height = 19
[Grid-Debug]   Row 1: final height = 19

[Grid-Debug] === Level 4: Children Measurement ===
[Grid-Debug] Cell[cell-0-0]: W:[20..20] H:[19..19]
[Grid-Debug] Cell[cell-0-0] result: 20x19
[Grid-Debug] Cell[cell-0-1]: W:[33..33] H:[19..19]
[Grid-Debug] Cell[cell-0-1] result: 33x19
[Grid-Debug] Cell[cell-1-0]: W:[20..20] H:[19..19]
[Grid-Debug] Cell[cell-1-0] result: 20x19
[Grid-Debug] Cell[cell-1-1]: W:[65..65] H:[19..19]
[Grid-Debug] Cell[cell-1-1] result: 65x19

[Grid-Debug] Grid[grid-123] Output Size: 120x40
[Grid-Debug] Grid[grid-123] Measure:
[Grid-Debug]   Input:  W:[0..120] H:[0..40]
[Grid-Debug]   Output: W:[120..120] H:[40..40]
[Grid-Debug]   Size:   120x40
```

---

## 4. 常见问题定位

### 4.1 问题分类

| 问题类型 | 症状 | 调试方法 |
|---------|------|-----------|
| **约束传递错误** | Grid 尺寸超出/不符合预期 | 检查 Level 1 → Level 2 约束对比 |
| **列宽计算错误** | Flex 列比例不正确 | 检查 Level 3 列计算日志 |
| **行高计算错误** | Auto 行高度不合理 | 检查 Level 3 行计算日志 + 子节点 Trace |
| **边框占位错误** | 尺寸计算少/多 | 对比有无边框的约束差异 |
| **Gap/Padding 错误** | 间距明显不对 | 检查 totalW/totalH 计算公式 |

### 4.2 约束传递错误调试

**问题场景**：
```
父容器约束: {0..50, 0..30}
Grid 返回尺寸: 80x40  ❌ 超出约束
```

**调试步骤**：

1. 在 `Instance.Measure` 入口添加断点：
```go
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // ⬇️ 断点 1：检查输入约束
    fmt.Printf("[DEBUG] Input constraints: %v\n", constraints)

    // ... 计算 ...

    // ⬇️ 断点 2：检查结果是否超过约束
    if resultSize.Width > constraints.MaxWidth {
        fmt.Printf("[ERROR] Width %d exceeds MaxWidth %d\n",
            resultSize.Width, constraints.MaxWidth)
    }
    if resultSize.Height > constraints.MaxHeight {
        fmt.Printf("[ERROR] Height %d exceeds MaxHeight %d\n",
            resultSize.Height, constraints.MaxHeight)
    }

    return resultSize
}
```

2. 检查返回值，可能的原因：
```go
// 原因 1：忘记应用约束
totalW = constraints.ConstrainWidth(totalW)  // ❌ 缺少这行

// 原因 2：显式尺寸没有受约束限制
if g.style.Width > 0 {
    totalW = g.style.Width  // ❌ 应该: totalW = min(totalW, constraints.MaxWidth)
}

// 原因 3：计算边框占位时出错
cellBorderWidth := len(g.style.Columns) + 1  // ❌ 应该检查 ShowCellBorders 标志
```

### 4.3 Flex 列宽度问题调试

**问题场景**：
```
Grid 设置: cols=[Flex(1), Flex(2)], width=120
预期: 列 0 = 40, 列 1 = 80
实际: 列 0 = 60, 列 1 = 60  ❌ 比例错误
```

**调试步骤**：

1. 检查 Flex 因子：
```go
// 在 calculateColumnWidths 中添加
for i, col := range g.style.Columns {
    if flex, ok := col.(GridFlex); ok {
        fmt.Printf("[DEBUG] Col %d: Flex factor = %d\n", i, flex.Factor)
    }
}
```

2. 检查 Flex 总数：
```go
if flexCount == 0 {
    fmt.Println("[DEBUG] WARNING: No Flex columns found!")
}

if flexTotalFactor == 0 {
    fmt.Println("[DEBUG] WARNING: Flex total factor is 0!")
}
```

3. 检查 Flex 分配逻辑：
```go
// 问题代码示例
widths[i] = remainingWidth / flexTotalFactor * c.Factor  // ❌ 整数除法提前

// 正确代码
widths[i] = (remainingWidth * c.Factor) / flexTotalFactor  // ✅
```

### 4.4 Auto 行高度问题调试

**问题场景**：
```
Grid 设置: rows=[Auto, Auto]
子格子尺寸: cell(0,0)=10, cell(0,1)=20
预期: 行 0 = max(10,20) = 20
实际: 行 0 = 10  ❌ 取了最小值
```

**调试步骤**：

1. 在行计算中打印每个格子的预期尺寸：
```go
func (g *GridLayout) calculateRowHeights(availableHeight int, numCols int) []int {
    DebugPrintf("=== Row Height Calculation ===")
    DebugPrintf("Available height: %d", availableHeight)

    heights := make([]int, numRows)

    for row := 0; row < numRows; row++ {
        maxHeight := 0

        DebugPrintf("--- Row %d ---", row)

        for col := 0; col < numCols; col++ {
            cell := g.findCell(row, col)
            if cell == nil {
                continue
            }

            // 假设我们获取子格子的期望高度
            cellHeight := g.getCellExpectedHeight(row, col)
            DebugPrintf("  Cell(%d,%d) expected height: %d", row, col, cellHeight)

            if cellHeight > maxHeight {
                maxHeight = cellHeight
            }
        }

        heights[row] = maxHeight
        DebugPrintf("Row %d final height: %d", row, maxHeight)
    }

    return heights
}
```

2. 检查是否正确调用子节点 `Measure`：
```go
func (g *GridLayout) getCellExpectedHeight(row, col int) int {
    cell := g.findCell(row, col)
    if cell == nil || cell.Child == nil {
        return 0
    }

    // 计算 Cell 约束
    cellConstraints := g.calculateCellConstraints(cell)

    // ⬇️ 调用子节点 Measure
    expectedHeight := cell.Child.Measure(cellConstraints).Height

    DebugPrintf("Cell(%d,%d) measured height: %d", row, col, expectedHeight)
    return expectedHeight
}
```

### 4.5 边框占位问题调试

**问题场景**：
```
Grid 设置: 禁用边框
实际尺寸: 比预期多 4 字符  ❌ 多算了边框占位
```

**调试步骤**：

```go
// 在 Measure 中打印边框相关计算
DebugPrintf("=== Border Calculation ===")
DebugPrintf("ShowCellBorders: %v", g.style.ShowCellBorders)
DebugPrintf("NumCols: %d, NumRows: %d", numCols, numRows)

cellBorderWidth := 0
cellBorderHeight := 0
if g.style.ShowCellBorders {
    cellBorderWidth = numCols + 1
    cellBorderHeight = numRows + 1
}

DebugPrintf("BorderWidth: %d, BorderHeight: %d",
    cellBorderWidth, cellBorderHeight)

DebugPrintf("AvailableWidth before border: %d", constraints.MaxWidth)
DebugPrintf("AvailableWidth after border: %d", availableWidth)

availableWidth := constraints.MaxWidth - g.style.Padding.Left - g.style.Padding.Right - cellBorderWidth
```

定位问题：
- 如果 `ShowCellBorders=false` 但 `cellBorderWidth` 仍计算导致错误
- 检查 `if g.style.ShowCellBorders` 条件是否生效

### 4.6 Grid 右边框位置问题调试

**问题场景**：
```
Grid 宽度: 79, 3 列, 启用 cellBorders
最后边框位置: x=78  ✓ 正确
---
Grid 宽度: 78, 2 列, 启用 cellBorders
最后边框位置: x=75  ❌ 应该是 x=77
```

**原因分析**：
当 Layout Engine 调用 `SetBounds` 设置新的宽度时，如果该宽度与 `Measure` 返回的理想宽度不同，
列宽 `colWidths` 没有重新计算，导致右边框位置不正确。

**调试步骤**：

1. 在 `SetBounds` 中添加详细日志：
```go
func (inst *Instance) SetBounds(x, y, w, h int) {
    // 计算边框占用
    cellBorderWidth := 0
    if inst.showCellBorders {
        cellBorderWidth = numCols + 1
    }

    // 计算可用宽度
    availableW := w - inst.padding[3] - inst.padding[1] - cellBorderWidth
    if availableW < 0 {
        availableW = 0
    }

    // 打印调试信息
    if inst.showCellBorders {
        fmt.Printf("[DEBUG GRID SETBOUNDS] x=%d, y=%d, w=%d, h=%d\n", x, y, w, h)
        fmt.Printf("[DEBUG GRID SETBOUNDS] numCols=%d, cellBorderWidth=%d, availableW=%d\n",
            numCols, cellBorderWidth, availableW)
        fmt.Printf("[DEBUG GRID SETBOUNDS] oldColWidths=%v\n", inst.colWidths)

        // 重新计算列宽
        inst.colWidths = inst.calculateColumnWidths(availableW)

        fmt.Printf("[DEBUG GRID SETBOUNDS] newColWidths=%v\n", inst.colWidths)

        // 验证最后边框位置
        lastBorderX := 0
        for _, cw := range inst.colWidths {
            lastBorderX += cw + 1
        }
        fmt.Printf("[DEBUG GRID SETBOUNDS] lastBorderX=%d, expected=%d (GridWidth-1)\n",
            lastBorderX, w-1)
    }
}
```

2. 检查边框位置计算：
```go
// 在边框绘制函数中计算边框位置
func (inst *Instance) GenCellBorderDrawCmds(originX, originY int) []paint.DrawCmd {
    numCols := len(inst.colWidths)
    numRows := len(inst.rowHeights)

    // 验证所有边框位置
    fmt.Printf("[DEBUG BORDERS] Border positions:\n")
    for col := 0; col <= numCols; col++ {
        x := 0
        for c := 0; c < col; c++ {
            x += inst.colWidths[c] + 1
        }
        fmt.Printf("  col[%d]: x=%d\n", col, x)
    }
    lastBorderX := 0
    for _, cw := range inst.colWidths {
        lastBorderX += cw + 1
    }
    gridWidth := inst.bounds[2]
    fmt.Printf("  last: x=%d, gridWidth=%d, expected=%d\n",
        lastBorderX, gridWidth, gridWidth-1)

    // ... 继续绘制边框
}
```

3. 验证关键断言：
```go
// 在测试中添加强校验
func TestGridRightBorderPosition(t *testing.T) {
    gridWidth := 79
    numCols := 3
    expectedLastBorderX := gridWidth - 1  // 78

    // ... 创建 Grid ...

    // 计算最后边框
    lastBorderX := 0
    for _, cw := range grid.colWidths {
        lastBorderX += cw + 1
    }

    assert.Equal(t, expectedLastBorderX, lastBorderX,
        "最后边框必须位于 GridWidth-1 位置")
}
```

**常见原因排查**：

| 症状 | 可能原因 | 解决方案 |
|------|---------|---------|
| 最后边框位置 = 0 | colWidths 为空数组 | 检查 SetBounds 是否在 Measure 之后调用 |
| 最后边框位置偏小 | colWidths 没有重新计算 | 在 SetBounds 中添加 calculateColumnWidths 调用 |
| 最后一列内容溢出 | colWidths 计算错误 | 检查 cellBorderWidth 是否正确考虑边框 |
| 边框与内容重叠 | getCellSize 返回值包含边框 | 确保返回纯内容尺寸（不含边框） |

---

## 5. 可视化调试

### 5.1 ASCII 布局视图

```go
// ui/components/grid/debug.go

// PrintLayoutASCII 打印 ASCII 版本的布局视图
func PrintLayoutASCII(gridId string, colWidths, rowHeights []int, colGap, rowGap int) {
    DebugPrintf("=== ASCII Layout View (%s) ===", gridId)

    // 打印列宽度
    colLine := "Col widths: "
    for i, w := range colWidths {
        colLine += fmt.Sprintf("[%d:%d] ", i, w)
    }
    DebugPrintf(colLine)

    // 打印行高度
    rowLine := "Row heights: "
    for i, h := range rowHeights {
        rowLine += fmt.Sprintf("[%d:%d] ", i, h)
    }
    DebugPrintf(rowLine)

    // 打印简单的网格表示
    numCols := len(colWidths)
    numRows := len(rowHeights)

    // 打印顶边框
    topBorder := "┌"
    for i := 0; i < numCols; i++ {
        for j := 0; j < colWidths[i]; j++ {
            topBorder += "─"
        }
        if i < numCols-1 {
            for g := 0; g < colGap; g++ {
                topBorder += " "
            }
            topBorder += "┬"
        }
    }
    topBorder += "┐"
    DebugPrintf(topBorder)

    // 打印每个格子
    for row := 0; row < numRows; row++ {
        contentLine := "│"
        for col := 0; col < numCols; col++ {
            cellWidth := colWidths[col]
            cellHeight := rowHeights[row]
            contentLine += fmt.Sprintf("(%dx%d)", cellWidth, cellHeight)
            // 填充剩余空格
            for w := len(fmt.Sprintf("(%dx%d)", cellWidth, cellHeight)); w < cellWidth; w++ {
                contentLine += " "
            }
            if col < numCols-1 {
                for g := 0; g < colGap; g++ {
                    contentLine += " "
                }
                contentLine += "│"
            }
        }
        contentLine += "│"
        DebugPrintf(contentLine)

        // 打印分隔线（如果不是最后一行）
        if row < numRows-1 {
            sepLine := "├"
            for i := 0; i < numCols; i++ {
                for j := 0; j < colWidths[i]; j++ {
                    sepLine += "─"
                }
                if i < numCols-1 {
                    for g := 0; g < colGap; g++ {
                        sepLine += " "
                    }
                    sepLine += "┼"
                }
            }
            sepLine += "┤"
            DebugPrintf(sepLine)
        }
    }

    // 打印底边框
    bottomBorder := "└"
    for i := 0; i < numCols; i++ {
        for j := 0; j < colWidths[i]; j++ {
            bottomBorder += "─"
        }
        if i < numCols-1 {
            for g := 0; g < colGap; g++ {
                bottomBorder += " "
            }
            bottomBorder += "┴"
        }
    }
    bottomBorder += "┘"
    DebugPrintf(bottomBorder)
}
```

### 5.2 使用 ASCII 布局视图

```go
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // ... 计算逻辑 ...

    // 打印 ASCII 布局视图（调用完成后）
    if DebugMode {
        PrintLayoutASCII(
            fmt.Sprintf("grid-%d", inst.fiberId),
            inst.GetColumnWidths(),
            inst.GetRowHeights(),
            inst.columnGap,
            inst.rowGap,
        )
    }

    return size
}
```

**输出示例**：
```
[Grid-Debug] === ASCII Layout View (grid-123) ===
[Grid-Debug] Col widths: [0:20] [1:33] [2:65]
[Grid-Debug] Row heights: [0:19] [1:19]
[Grid-Debug] ┌──────────────────┬───────────────────────────┬────────────────────────────┐
[Grid-Debug] │(20x19)            │(33x19)                    │(65x19)                      │
[Grid-Debug] ├──────────────────┼───────────────────────────┼────────────────────────────┤
[Grid-Debug] │(20x19)            │(33x19)                    │(65x19)                      │
[Grid-Debug] └──────────────────┴───────────────────────────┴────────────────────────────┘
```

### 5.3 实时可视化面板（建议）

对于复杂的 UI，建议在终端右侧实现一个实时调试面板：

```
┌─────────────────────────────────────────────┬──────────────────────────┐
│   主应用区域 (Grid 渲染结果)                 │   调试面板               │
│   ┌────┬────┬────┐                         │  ┌─────────────────────┐  │
│   │ A1 │ A2 │ A3 │                         │  │ Grid Layout Debug    │  │
│   ├────┼────┼────┤                         │  │ ─────────────────── │  │
│   │ B1 │ B2 │ B3 │                         │  │ Constraints:         │  │
│   └────┴────┴────┘                         │  │   W: [0..120]        │  │
│                                             │  │   H: [0..40]         │  │
│                                             │  │ Result: 120x40       │  │
│                                             │  │                      │  │
│                                             │  │ Columns:             │  │
│                                             │  │   Col 0: 20 (Fixed) │  │
│                                             │  │   Col 1: 33 (Flex)  │  │
│                                             │  │   Col 2: 65 (Flex)  │  │
│                                             │  │                      │  │
│                                             │  │ Rows:                │  │
│                                             │  │   Row 0: 19 (Auto)   │  │
│                                             │  │   Row 1: 19 (Auto)   │  │
│                                             │  └─────────────────────┘  │
└─────────────────────────────────────────────┴──────────────────────────┘
```

---

## 6. 实战案例

### 6.1 案例 1：Grid 尺寸超出约束

**问题描述**：
```
Grid 设置: cols=[Fixed(30), Fixed(30)], width=explicit=80
约束: MaxWidth=70
结果: 返回 80  ❌ 超出约束
```

**调试流程**：

```go
// Step 1: 确认问题
func Test_Grid_ConstraintViolation() {
    g := New().
        SetColumns(Fixed(30), Fixed(30)).
        SetWidth(80)  // 显式宽度 80

    constraints := layout.NewConstraints(0, 70, 0, 50)  // 最大宽度 70

    inst := g.CreateInstance()
    size := inst.Measure(constraints)

    // 确认问题
    if size.Width > constraints.MaxWidth {
        fmt.Printf("❌ PROBLEM: Width %d exceeds constraint %d\n",
            size.Width, constraints.MaxWidth)
    }
}

// Step 2: 在 Measure 中添加调试
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    DebugPrintf("=== Problem Case: Explicit Width Exceeds Constraint ===")
    DebugPrintf("Explicit width: %d", inst.width)
    DebugPrintf("Constraint MaxWidth: %d", constraints.MaxWidth)

    // 计算布局
    totalW := sum(colWidths) + gaps + padding
    if inst.width > 0 {
        totalW = inst.width
        DebugPrintf("Applied explicit width: %d", totalW)
    }

    // ⬇️ 问题：没有检查约束
    // totalW = constraints.ConstrainWidth(totalW)  // 缺少这行

    DebugPrintf("Total width before constraint: %d", totalW)

    return layout.Size{Width: totalW, Height: totalH}
}

// Step 3: 查看日志输出
// [Grid-Debug] Explicit width: 80
// [Grid-Debug] Constraint MaxWidth: 70
// [Grid-Debug] Applied explicit width: 80
// [Grid-Debug] Total width before constraint: 80
// ❌ PROBLEM: Width 80 exceeds constraint 70

// Step 4: 修复问题
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    totalW := ... // 计算结果

    if inst.width > 0 {
        totalW = inst.width
    }

    // ✅ 修复：应用约束
    totalW = constraints.ConstrainWidth(totalW)

    return layout.Size{Width: totalW, ...}
}
```

### 6.2 案例 2：Flex 列宽度分配不等

**问题描述**：
```
Grid 设置: cols=[Flex(1), Flex(2)], width=120
预期: 列 0=40, 列 1=80
实际: 列 0=60, 列 1=60  ❌ 比例错误
```

**调试流程**：

```go
// 在 calculateColumnWidths 中添加详细日志
func (g *GridLayout) calculateColumnWidths(availableWidth int) []int {
    DebugPrintf("=== Flex Width Distribution Problem ===")
    DebugPrintf("Available width: %d", availableWidth)

    // Step 1: 计算固定宽度
    fixedWidth := 0
    for i, col := range g.style.Columns {
        if _, ok := col.(GridFixed); ok {
            fixedWidth += int(col.(GridFixed))
            DebugPrintf("Col %d: Fixed %d, total fixed = %d", i, col.(GridFixed), fixedWidth)
        }
    }

    // Step 2: 计算剩余宽度
    gapWidth := g.style.ColumnGap * (numCols - 1)
    remainingWidth := availableWidth - fixedWidth - gapWidth
    DebugPrintf("Remaining width: %d", remainingWidth)

    // Step 3: 收集 Flex 因子
    flexCols := []int{}
    flexTotalFactor := 0
    for i, col := range g.style.Columns {
        if f, ok := col.(GridFlex); ok {
            flexCols = append(flexCols, i)
            if f.Factor == 0 {
                flexTotalFactor += 1
            } else {
                flexTotalFactor += f.Factor
            }
            DebugPrintf("Col %d: Flex factor %d, running total = %d", i, f.Factor, flexTotalFactor)
        }
    }

    // Step 4: 分配 Flex 宽度
    heights := make([]int, numCols)
    for _, i := range flexCols {
        flex := g.style.Columns[i].(GridFlex)

        // ✅ 正确的分配算法
        width := (remainingWidth * flex.Factor) / flexTotalFactor

        DebugPrintf("Col %d: (%d * %d) / %d = %d", i, remainingWidth, flex.Factor, flexTotalFactor, width)
        heights[i] = width
    }

    return heights
}

// 日志输出：
// [Grid-Debug] Available width: 120
// [Grid-Debug] Col 0: Flex factor 1, running total = 1
// [Grid-Debug] Col 1: Flex factor 2, running total = 3
// [Grid-Debug] Remaining width: 120
// [Grid-Debug] Col 0: (120 * 1) / 3 = 40  ✅ 正确
// [Grid-Debug] Col 1: (120 * 2) / 3 = 80  ✅ 正确
```

### 6.3 案例 3：Auto 行高度计算错误

**问题描述**：
```
Grid 设置: rows=[Auto, Auto], cols=[Fixed(20), Fixed(20)]
子格子: cell(0,0)=高度10, cell(0,1)=高度30
预期: 行 0 = max(10,30) = 30
实际: 行 0 = 10  ❌ 取了最小值
```

**调试流程**：

```go
// 在 calculateRowHeights 中详细追踪
func (g *GridLayout) calculateRowHeights(availableHeight int, numCols int) []int {
    DebugPrintf("=== Auto Row Height Problem ===")

    numRows := g.calculateRowCount()
    heights := make([]int, numRows)

    for row := 0; row < numRows; row++ {
        DebugPrintf("--- calculating row %d ---", row)

        maxHeight := 0

        for col := 0; col < numCols; col++ {
            cell := g.findCell(row, col)
            if cell == nil || cell.Child == nil {
                Debug.Printf("  Col %d: no cell, skip", col)
                continue
            }

            // 计算子格子约束
            cellConstraints := layout.Constraints{
                MinWidth:  g.colWidths[col],
                MaxWidth:  g.colWidths[col],
                MinHeight: 0,
                MaxHeight: availableHeight,
            }

            // 测量子格子高度
            cellHeight := cell.Child.Measure(cellConstraints).Height

            Debug.Printf("  Col %d: cell measured height = %d", col, cellHeight)

            // ⬇️ 关键：取最大值
            if cellHeight > maxHeight {
                maxHeight = cellHeight
                Debug.Printf("  → new max height: %d (from col %d)", maxHeight, col)
            }
        }

        heights[row] = maxHeight
        Debug.Printf("Row %d final height: %d\n", row, maxHeight)
    }

    return heights
}

// 日志输出：
// [Grid-Debug] --- calculating row 0 ---
// [Grid-Debug]   Col 0: cell measured height = 10
// [Grid-Debug]   → new max height: 10 (from col 0)
// [Grid-Debug]   Col 1: cell measured height = 30
// [Grid-Debug]   → new max height: 30 (from col 1)  ✅ 取最大值
// [Grid-Debug] Row 0 final height: 30
```

---

## 7. 调试技巧总结

### 7.1 快速检查清单

| 检查项 | 位置 | 检查方法 |
|--------|------|-----------|
| **约束输入正确** | `Instance.Measure` 入口 | `fmt.Printf("Input: %v", constraints)` |
| **约束输出合规** | `Instance.Measure` 返回前 | `if size >约束 { panic }` |
| **列宽计算正确** | `calculateColumnWidths` | 打印每列的分配过程 |
| **行高计算正确** | `calculateRowHeights` | 打印每行的计算过程 |
| **子节点测量正常** | 遍历格子处 | 打印子节点的 `Measure` 结果 |
| **边框占位正确** | 计算 `cellBorderWidth` | 对比有/无边框的尺寸差异 |

### 7.2 调试命令速查

```bash
# 启用调试模式运行
MINT_DEBUG_GRID=true go run main.go

# 启用追踪系统
MINT_TRACE_LAYOUT=true go run main.go

# 运行带追踪的测试
MINT_TRACE_LAYOUT=true go test -v ./ui/components/grid/...

# 运行特定测试
go test -v -run TestGrid_Flex ./ui/components/grid/...
```

### 7.3 日志级别建议

| 级别 | 内容 | 启用方式 |
|------|------|-----------|
| **ERROR** | 致命错误（约束违规、负值尺寸） | 始终输出 |
| **WARN** | 警告（异常值、可能的问题） | 生产环境启用 |
| **INFO** | 关键节点（Measure 入口/出口、边界计算） | 生产环境启用 |
| **DEBUG** | 详细计算（每列/每行的分配过程） | 开发/调试时启用 |
| **TRACE** | 最详细（子节点测量、中间变量） | 仅深度调试时启用 |

---

## 附录：调试配置示例

```go
// ui/components/grid/config.go

package grid

import (
    "os"
    "strconv"
    "github.com/wwsheng009/mint/runtime/layout"
)

// InitDebugFromEnv 从环境变量初始化调试配置
func InitDebugFromEnv() {
    // 启用 Debug 模式
    if debugStr := os.Getenv("MINT_DEBUG_GRID"); debugStr == "true" || debugStr == "1" {
        DebugMode = true
        println("[Grid Debug] Debug mode enabled")
    }

    // 启用追踪系统
    if traceStr := os.Getenv("MINT_TRACE_LAYOUT"); traceStr == "true" || traceStr == "1" {
        layout.EnableTracer()
        println("[Grid Debug] Tracer enabled")
    }

    // 设置日志级别
    if levelStr := os.Getenv("MINT_LOG_LEVEL"); levelStr != "" {
        ParseLogLevel(levelStr)
    }
}

// ParseLogLevel 解析日志级别
func ParseLogLevel(level string) {
    switch level {
    case "ERROR":
        SetLogLevel(LogLevelError)
    case "WARN":
        SetLogLevel(LogLevelWarn)
    case "INFO":
        SetLogLevel(LogLevelInfo)
    case "DEBUG":
        SetLogLevel(LogLevelDebug)
    case "TRACE":
        SetLogLevel(LogLevelTrace)
    default:
        SetLogLevel(LogLevelInfo)
    }
}

// LogLevel 日志级别
type LogLevel int

const (
    LogLevelError LogLevel = iota
    LogLevelWarn
    LogLevelInfo
    LogLevelDebug
    LogLevelTrace
)

var currentLogLevel = LogLevelInfo

// SetLogLevel 设置日志级别
func SetLogLevel(level LogLevel) {
    currentLogLevel = level
}

// 日志输出函数（带级别过滤）
func logPrint(level LogLevel, format string, args ...interface{}) {
    if level <= currentLogLevel {
        log.Printf("[Grid-%s] "+format, levelToString(level), args...)
    }
}

func levelToString(level LogLevel) string {
    switch level {
    case LogLevelError:
        return "ERROR"
    case LogLevelWarn:
        return "WARN"
    case LogLevelInfo:
        return "INFO"
    case LogLevelDebug:
        return "DEBUG"
    case LogLevelTrace:
        return "TRACE"
    default:
        return "UNKNOWN"
    }
}
```

在应用启动时调用：
```go
import "github.com/wwsheng009/mint/ui/components/grid"

func main() {
    // 初始化 Grid 调试配置
    grid.InitDebugFromEnv()

    // ... 启动应用 ...
}
```
