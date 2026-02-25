# Grid 追踪系统集成指南

本文档说明如何将 Grid 组件与 `runtime/layout/tracer.go` 追踪系统集成。

---

## 目录

1. [概述](#1-概述)
2. [追踪系统简介](#2-追踪系统简介)
3. [集成策略](#3-集成策略)
4. [实现步骤](#4-实现步骤)
5. [代码示例](#5-代码示例)
6. [追踪输出分析](#6-追踪输出分析)
7. [调试技巧](#7-调试技巧)

---

## 1. 概述

### 1.1 集成目标

为 Grid 组件添加完整的约束传播追踪能力，帮助开发者：

- 调试布局计算问题
- 理解约束在组件树中的传递
- 定位尺寸计算不正确的原因
- 优化布局性能

### 1.2 追踪范围

| 追踪点 | 位置 | 追踪内容 |
|--------|------|-----------|
| Grid 入口 | `ui/components/grid/instance.go:Measure()` | 输入约束 |
| 列计算 | `runtime/layout/grid.go:calculateColumnWidths()` | 列宽约束传递 |
| 行计算 | `runtime/layout/grid.go:calculateRowHeights()` | 行高约束传递 |
| 子节点测量 | 遍历格子时的子节点测量 | 子节点约束 + 结果 |
| Grid 出口 | `ui/components/grid/instance.go:Measure()` | 最终尺寸 + 原因 |

---

## 2. 追踪系统简介

### 2.1 核心接口

```go
// 启用全局追踪器
func EnableTracer()

// 禁用全局追踪器
func DisableTracer()

// 检查追踪器是否启用
func IsTracerEnabled() bool

// 追踪测量过程
func TraceMeasuring(
    from string,           // 来源组件 ID
    to string,             // 目标组件 ID
    path string,           // 完整路径
    input Constraints,     // 输入约束
    output Constraints,    // 输出约束
    resultSize Size,       // 测量结果
    reason string,         // 约束修改原因
)

// 输出追踪日志
func DumpTrace() string

// 清除追踪数据
func ClearTrace()

// 获取追踪条目（用于测试）
func GetTraceEntries() []TraceEntry
```

### 2.2 追踪条目结构

```go
type TraceEntry struct {
    Seq        int           // 序列号
    Timestamp  time.Time     // 时间戳
    From       string        // 来源组件 ID
    To         string        // 目标组件 ID
    Path       string        // 完整路径
    Input      Constraints   // 输入约束
    Output     Constraints   // 输出约束
    Dimension  Size          // 测量结果
    Reason     string        // 约束修改原因
}
```

---

## 3. 集成策略

### 3.1 两层集成

Grid 组件的追踪需要在两个层次进行：

```
┌─────────────────────────────────────────────────┐
│  ui/components/grid/instance.go                  │
│  - Instance.Measure() 入口追踪                  │
│  - 委托 runtime/layout/Grid                      │
│  - Instance.Measure() 出口追踪                  │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│  runtime/layout/grid.go                          │
│  - calculateColumnWidths() 追踪列计算           │
│  - calculateRowHeights() 追踪行计算             │
│  - 子节点测量追踪（若有）                       │
└─────────────────────────────────────────────────┘
```

### 3.2 路径命名规范

```
root/grids/{grid-key}              - Grid 根节点
root/grids/{grid-key}/col-{index}  - 某一列
root/grids/{grid-key}/row-{index}  - 某一行
root/grids/{grid-key}/cell-{r}-{c} - 某一格子
```

### 3.3 ID 命名规范

```
grid-{NodeID}       - Grid 节点（使用 Fiber.NodeID）
col-{index}         - 列标识
row-{index}         - 行标识
cell-{row}-{col}    - 格子标识
```

---

## 4. 实现步骤

### 步骤 1：在 Instance.Measure 入口添加追踪

修改 `ui/components/grid/instance.go`：

```go
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 生成路径
    path := fmt.Sprintf("root/grids/%s", inst.key)

    // 记录入口
    layout.TraceMeasuring(
        "parent",                  // from
        fmt.Sprintf("grid-%d", inst.fiberId),  // to
        path,                      // path
        constraints,               // input
        layout.Constraints{},      // output (待填充)
        layout.Size{},             // dimension (待填充)
        "Grid.Measure entrance",   // reason
    )

    // ... 原有测量逻辑 ...
}
```

### 步骤 2：在 computeColumnWidths 添加追踪

修改 `runtime/layout/grid.go:calculateColumnWidths()`：

```go
func (g *GridLayout) calculateColumnWidths(availableWidth int) []int {
    numCols := len(g.style.Columns)
    if numCols == 0 {
        return []int{availableWidth}
    }

    // 追踪可用宽度
    layout.TraceMeasuring(
        g.id,
        "grid-layout-columns",
        g.id+"/columns",
        layout.Constraints{MinWidth: 0, MaxWidth: availableWidth},
        layout.Constraints{},
        layout.Size{},
        "Column calculation start",
    )

    // ... 原有计算逻辑 ...

    // 追踪每列计算结果
    for i, width := range widths {
        colPath := fmt.Sprintf("%s/col-%d", g.id, i)
        layout.TraceMeasuring(
            fmt.Sprintf("col-%d", i),
            "computed",
            colPath,
            layout.Constraints{MinWidth: 0, MaxWidth: availableWidth},
            layout.Constraints{MinWidth: width, MaxWidth: width},
            layout.Size{Width: width, Height: 0},
            fmt.Sprintf("Column %d width calculated", i),
        )
    }

    return widths
}
```

### 步骤 3：在 computeRowHeights 添加追踪

类似步骤 2，添加行计算的追踪。

### 步骤 4：在 Instance.Measure 出口添加追踪

```go
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // ... 测量逻辑 ...

    // 计算输出约束（基于结果尺寸）
    outputConstraints := layout.Constraints{
        MinWidth:  resultSize.Width,
        MaxWidth:  resultSize.Width,
        MinHeight: resultSize.Height,
        MaxHeight: resultSize.Height,
    }

    // 记录出口
    layout.TraceMeasuring(
        fmt.Sprintf("grid-%d", inst.fiberId),
        "parent",
        path,
        constraints,
        outputConstraints,
        resultSize,
        "Grid.Measure complete",
    )

    return resultSize
}
```

---

## 5. 代码示例

### 5.1 完整的 Instance.Measure 实现

```go
// ui/components/grid/instance.go

package grid

import (
    "fmt"

    "github.com/wwsheng009/mint/runtime/layout"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Instance represents a Grid component instance
type Instance struct {
    // ... 现有字段 ...
    fiberId int   // 添加：Fiber 节点 ID
    key     string // 添加：组件 Key
}

// Measure 计算布局尺寸
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 生成路径
    path := fmt.Sprintf("root/grids/%s", inst.key)

    // 记录入口
    layout.TraceMeasuring(
        "parent",
        fmt.Sprintf("grid-%d", inst.fiberId),
        path,
        constraints,
        layout.Constraints{},
        layout.Size{},
        "Grid.Measure entrance",
    )

    // 委托给 runtime/layout/Grid 计算布局
    gridLayout := layout.NewGridLayout(
        fmt.Sprintf("grid-%d", inst.fiberId),
        &layout.GridStyle{
            Columns:          inst.convertDimensions(inst.columns),
            Rows:             inst.convertDimensions(inst.rows),
            Cells:            inst.convertCells(),
            ColumnGap:        inst.columnGap,
            RowGap:           inst.rowGap,
            Padding:          layout.Padding{
                Top:    inst.padding[0],
                Right:  inst.padding[1],
                Bottom: inst.padding[2],
                Left:   inst.padding[3],
            },
            Width:            inst.width,
            Height:           inst.height,
            ShowCellBorders:  inst.showCellBorders,
            CellBorderWidth:  inst.cellBorderWidth,
            CellBorderHeight: inst.cellBorderHeight,
        },
    )

    // 执行测量（内部会追踪列/行计算）
    resultSize := gridLayout.Measure(constraints)

    // 计算输出约束
    outputConstraints := layout.Constraints{
        MinWidth:  resultSize.Width,
        MaxWidth:  resultSize.Width,
        MinHeight: resultSize.Height,
        MaxHeight: resultSize.Height,
    }

    // 记录出口
    layout.TraceMeasuring(
        fmt.Sprintf("grid-%d", inst.fiberId),
        "parent",
        path,
        constraints,
        outputConstraints,
        resultSize,
        fmt.Sprintf(
            "Grid.Measure complete: %d×%d (cols=%d, rows=%d, gap=%d×%d)",
            resultSize.Width,
            resultSize.Height,
            len(inst.columns),
            len(inst.rows),
            inst.columnGap,
            inst.rowGap,
        ),
    )

    return resultSize
}
```

### 5.2 列计算追踪示例

```go
// runtime/layout/grid.go

func (g *GridLayout) calculateColumnWidths(availableWidth int) []int {
    numCols := len(g.style.Columns)
    if numCols == 0 {
        return []int{availableWidth}
    }

    colPath := fmt.Sprintf("%s/columns", g.id)

    // 追踪：列计算开始
    layout.TraceMeasuring(
        g.id,
        "columns-layout",
        colPath,
        layout.Constraints{MinWidth: 0, MaxWidth: availableWidth},
        layout.Constraints{},
        layout.Size{},
        "Column widths calculation start",
    )

    // ... 原有计算逻辑 ...

    // 追踪：汇总结果
    totalFixed := 0
    for _, w := range widths {
        totalFixed += w
    }

    layout.TraceMeasuring(
        "columns-layout",
        "computed",
        colPath+"/result",
        layout.Constraints{MinWidth: 0, MaxWidth: availableWidth},
        layout.Constraints{MinWidth: totalFixed, MaxWidth: totalFixed},
        layout.Size{Width: totalFixed, Height: 0},
        fmt.Sprintf("Columns calculated: fixed=%d, flex=%d, remaining=%d",
            totalFixed, flexCount, remainingWidth),
    )

    return widths
}
```

### 5.3 测试中使用追踪

```go
// ui/components/grid/tracer_test.go

package grid

import (
    "strings"
    "testing"

    "github.com/wwsheng009/mint/runtime/layout"
)

func TestGrid_Tracing(t *testing.T) {
    // 启用追踪
    layout.EnableTracer()
    defer layout.DisableTracer()
    defer layout.ClearTrace()

    // 创建 Grid
    g := New().
        SetColumns(Flex{Factor: 1}, Flex{Factor: 2}).
        SetRows(Auto{}).
        ShowCellBorders().
        SetChildrenAuto([]rtui.VNode{
            text.New("A"), text.New("B"),
        })

    // 执行测量
    inst := g.CreateInstance()
    constraints := layout.NewConstraints(0, 100, 0, 50)
    inst.Measure(constraints)

    // 获取追踪条目
    entries := layout.GetTraceEntries()

    // 验证追踪条目
    if len(entries) == 0 {
        t.Fatal("No trace entries recorded")
    }

    // 检查入口追踪
    foundEntrance := false
    for _, entry := range entries {
        if entry.Reason == "Grid.Measure entrance" {
            foundEntrance = true
            if entry.Input.MaxWidth != 100 {
                t.Errorf("Input MaxWidth: expected 100, got %d", entry.Input.MaxWidth)
            }
        }
    }

    if !foundEntrance {
        t.Error("Grid.Measure entrance trace not found")
    }

    // 检查列计算追踪
    foundColumns := false
    for _, entry := range entries {
        if strings.Contains(entry.Path, "/columns") {
            foundColumns = true
            break
        }
    }

    if !foundColumns {
        t.Error("Column calculation trace not found")
    }

    // 输出追踪日志（用于调试）
    t.Log("Trace log:")
    t.Log(layout.DumpTrace())
}
```

---

## 6. 追踪输出分析

### 6.1 典型输出示例

```
╔══════════════════════════════════════════════════════════════════╗
║                    Constraint Propagation Trace               ║
╚══════════════════════════════════════════════════════════════════╝

Step 0
  Path: root/grids/my-table
  parent → grid-123
  Input:    {0..100} × {0..50}
  Output:   {60..60} × {30..30}
  Dimension: 60w × 30h
  Reason:   Grid.Measure entrance

Step 1
  Path: root/grids/my-table/columns
  grid-123 → columns-layout
  Input:    {0..94} × {0..50}
  Output:   {0..0} × {0..0}
  Reason:   Column widths calculation start

Step 2
  Path: root/grids/my-table/columns/col-0
  col-0 → computed
  Input:    {0..94} × {0..50}
  Output:   {31..31} × {0..0}
  Dimension: 31w × 0h
  Reason:   Column 0 width calculated

Step 3
  Path: root/grids/my-table/columns/col-1
  col-1 → computed
  Input:    {0..94} × {0..50}
  Output:   {62..62} × {0..0}
  Dimension: 62w × 0h
  Reason:   Column 1 width calculated

Step 4
  Path: root/grids/my-table/columns/result
  columns-layout → computed
  Input:    {0..94} × {0..50}
  Output:   {93..93} × {0..0}
  Dimension: 93w × 0h
  Reason:   Columns calculated: fixed=0, flex=2, remaining=94

Step 5
  Path: root/grids/my-table
  grid-123 → parent
  Input:    {0..100} × {0..50}
  Output:   {60..60} × {30..30}
  Dimension: 60w × 30h
  Reason:   Grid.Measure complete: 60×30 (cols=2, rows=1, gap=1×1)
```

### 6.2 分析技巧

#### 识别约束问题

| 问题 | 追踪特征 |
|------|-----------|
| **约束传递错误** | Step N 的 Input ≠ Step (N-1) 的 Output |
| **尺寸计算错误** | Dimension 异常（如负数、过大） |
| **Flex 分配错误** | Flex 列宽度比例不符合预期 |
| **边框占位错误** | 输入约束没有正确减去边框宽度 |

#### 示例分析

```
问题：Grid 返回宽度超出约束

Step 0
  Input:    {0..50} × {0..30}
  ...
  Reason:   Grid.Measure entrance

Step 5
  Input:    {0..50}
  Output:   {60..60}
  Dimension: 60w × 30h
  Reason:   Grid.Measure complete: 60×30

分析：
- Input MaxWidth=50, Output Width=60, 超出约束
- 检查：是否忘记应用 constraints.ConstrainWidth()

解决：在 Measure 返回前添加：
  totalW = constraints.ConstrainWidth(totalW)
```

---

## 7. 调试技巧

### 7.1 启用追踪的多种方式

#### 方式 1：代码中启用

```go
import "github.com/wwsheng009/mint/runtime/layout"

func runGrid() {
    layout.EnableTracer()
    defer layout DisableTracer()

    // ... 运行 Grid ...
    println(layout.DumpTrace())
}
```

#### 方式 2：环境变量控制

```go
// 在应用初始化时
import "os"

func main() {
    if os.Getenv("MINT_TRACE_LAYOUT") == "true" {
        layout.EnableTracer()
    }

    // ... 运行应用 ...
}
```

#### 方式 3：测试中启用

```go
func TestGrid_LayoutWithTrace(t *testing.T) {
    layout.EnableTracer()
    defer layout.DisableTracer()
    defer layout.ClearTrace()

    // ... 测试代码 ...

    t.Logf("\n%s", layout.DumpTrace())
}
```

### 7.2 查看特定节点的追踪

```go
func findNodeTrace(entries []layout.TraceEntry, pathPattern string) []layout.TraceEntry {
    var result []layout.TraceEntry
    for _, entry := range entries {
        if strings.Contains(entry.Path, pathPattern) {
            result = append(result, entry)
        }
    }
    return result
}

// 使用示例
gridTraces := findNodeTrace(layout.GetTraceEntries(), "grids/my-table")
for _, entry := range gridTraces {
    fmt.Println(entry.Reason, entry.Dimension)
}
```

### 7.3 性能分析

```go
func analyzeTracePerformance(entries []layout.TraceEntry) {
    for i := 1; i < len(entries); i++ {
        duration := entries[i].Timestamp.Sub(entries[i-1].Timestamp)
        fmt.Printf("Step %d → %d: %v\n", i-1, i, duration)
    }
}
```

---

## 附录：追踪系统集成检查清单

- [ ] `ui/components/grid/instance.go:Measure()` 入口添加追踪
- [ ] `ui/components/grid/instance.go:Measure()` 出口添加追踪
- [ ] `runtime/layout/grid.go:calculateColumnWidths()` 添加列计算追踪
- [ ] `runtime/layout/grid.go:calculateRowHeights()` 添加行计算追踪
- [ ] 添加追踪测试用例 (`tracer_test.go`)
- [ ] 验证追踪输出格式正确
- [ ] 验证禁用追踪后无性能影响
- [ ] 更新 ARCHITECTURE.md 说明追踪集成
