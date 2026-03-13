# Grid 组件验收标准

本文档定义了 Grid 组件的完整验收标准，包括功能验收、渲染效果验收、性能验收和集成验收。

---

## 目录

1. [概述](#1-概述)
2. [VNode API 验收](#2-vnode-api-验收)
3. [Instance 测量验收](#3-instance-测量验收)
4. [渲染效果验收](#4-渲染效果验收)
5. [边框渲染验收](#5-边框渲染验收)
6. [坐标与定位验收](#6-坐标与定位验收)
7. [集成测试验收](#7-集成测试验收)
8. [性能验收](#8-性能验收)
9. [追踪系统集成](#9-追踪系统集成)
10. [验收检查清单](#10-验收检查清单)

---

## 1. 概述

### 1.1 组件目标

Grid 组件是一个基于 Fiber-first 架构的网格布局容器，提供类似 CSS Grid 的布局能力，支持：

- 灵活的列/行定义（Fixed、Flex、Auto、Min、Max）
- 自动或显式的格子放置
- 间距控制（Gap、Padding）
- 容器边框和 Cell 边框双重边框支持
- 约束驱动的响应式布局

### 1.2 架构要求

| 要求 | 说明 |
|------|------|
| **Fiber-first** | 完全基于 Fiber 树进行布局，不依赖 VNode |
| **职责分离** | 布局计算委托给 `runtime/layout/Grid` |
| **两阶段约束** | Measure（内部测量） + SetBounds（外部定位） |
| **统一坐标** | 基于 `contentX/contentY = bounds + padding` 统一基准 |

---

## 2. VNode API 验收

### 2.1 基本结构验收

| API | 验收标准 | 测试方法 |
|-----|-----------|-----------|
| `New()` | 返回非 nil VNode，Tag 为 "grid" | `TestVNode_New` |
| `Columns()` | 返回列定义数组 | 检查返回值类型和数量 |
| `Rows()` | 返回行定义数组 | 检查返回值类型和数量 |
| `Cells()` | 返回格子数组 | 检查格子数量和位置 |

### 2.2 列/行定义验收

| API | 验收标准 | 边界条件 |
|-----|-----------|-----------|
| `SetColumns(Fixed(10), Flex{Factor: 1}, Auto{})` | 正确设置 3 列，类型为 Fixed、Flex、Auto | 空、超大数量、无效值 |
| `SetRows(Fixed(3), Flex{Factor: 2})` | 正确设置 2 行，类型为 Fixed、Flex | 空、超大数量、无效值 |
| `Min{Min: 5, Content: Fixed(10)}` | 最小高度 5，内容为 Fixed(10) | Min > Content |
| `Max{Max: 20, Content: Flex{Factor: 1}}` | 最大宽度 20，内容为 Flex | Max < Content |

### 2.3 格子放置验收

| API | 验收标准 | 预期行为 |
|-----|-----------|-----------|
| `AddCell(row, col, node)` | node 放置在指定位置 | 允许跳过格子 |
| `SetChildrenAuto([]node)` | 按 row-major 顺序自动填充 | 2 列时：A(0,0) B(0,1) C(1,0) |

### 2.4 间距与尺寸验收

| API | 验收标准 | 链式调用 |
|-----|-----------|-----------|
| `SetGap(colGap, rowGap)` | 独立设置列间距和行间距 | ✅ 支持 |
| `SetPadding(top, right, bottom, left)` | 独立设置四边内边距 | ✅ 支持 |
| `SetWidth(w)` / `SetHeight(h)` | 设置显式宽高 | ✅ 支持 |
| `SetFlex(f)` | 设置弹性因子 | ✅ 支持 |

### 2.5 边框 API 验收

| API | 验收标准 | 渲染效果 |
|-----|-----------|-----------|
| `Border()` | 启用容器单线边框 | `┌─┐│└─┘` |
| `SingleBorder("title")` | 容器单线边框 + 标题 | 顶部标题栏 |
| `ShowCellBorders()` | 启用 Cell 单线边框 | 格子之间线条 |
| `DoubleCellBorders()` | Cell 双线边框 | `╔═╗║╚═╝` |
| `LightCellBorders()` | Cell 轻边框 | `│─` |
| `RoundedCellBorders()` | Cell 圆角 | `╭╮╰╯` |
| `SetCellBorderColor("cyan")` | Cell 边框颜色 | 指定 ANSI 颜色 |

### 2.6 链式调用验收

```go
// 验收示例：完整链式调用
grid.New().
    SetColumns(Flex{Factor: 1}, Flex{Factor: 1}).
    SetRows(Flex{Factor: 1}, Flex{Factor: 1}).
    SingleCellBorders().
    SetGap(1, 1).
    SetPadding(1, 1, 1, 1).
    SetChildrenAuto([]ui.VNode{
        text.New("A1"), text.New("A2"),
        text.New("B1"), text.New("B2"),
    })

// 验收标准：所有调用正确执行，无 nil 断点
```

---

## 3. Instance 测量验收

### 3.1 基本测量验收

| 场景 | 输入 | 验收标准 |
|------|------|-----------|
| **空网格** | `cols=[Flex(1)], rows=[Auto]` | 返回最小且非负尺寸 |
| **固定尺寸** | `cols=[Fixed(10), Fixed(20)], rows=[Fixed(2), Fixed(3)]` | Width=30, Height=5 |
| **Flex 尺寸** | `constraints={0..60, 0..20}`, `cols=[Flex(1), Flex(2)]` | Width=60, 列宽=[20, 40] |

### 3.2 约束传播验收

| 约束类型 | 验收标准 | 输出行为 |
|----------|-----------|-----------|
| **紧约束** | `TightConstraints(30, 10)` | 结果不超过 30×10 |
| **最大约束** | `Constraints{0..100, 0..50}` | Flex 按比例填充 |
| **最小约束** | `Constraints{50..200, 30..100}` | 结果 >= MinWidth/MinHeight |

### 3.3 尺寸计算验收

| 项目 | 计算公式 | 验收方法 |
|------|-----------|-----------|
| **总宽度** | `Sum(colWidths) + (n-1)*columnGap + padding[1] + padding[3] + borderWidth` | 对比 Measure 返回值 |
| **总高度** | `Sum(rowHeights) + (n-1)*rowGap + padding[0] + padding[2] + borderWidth` | 对比 Measure 返回值 |
| **边框宽度** | `cellBorders ? (cols+rows) : 0` | 单边框占 1 字符 |

### 3.4 显式尺寸验收

| 设置 | 验收标准 | 优先级 |
|------|-----------|--------|
| `width=50, height=20` | Measure 返回 50×20 | 最高 |
| `width=50, 约束 max=30` | 返回 30（受约束限制） | 约束优先 |
| `无显式尺寸` | 返回计算尺寸 | 默认 |

---

## 4. 渲染效果验收

### 4.1 基本网格渲染

**验收标准示例：2×2 网格**

```
预期无边框:
┌────┬────┐
│ A1 │ A2 │
├────┼────┤
│ B1 │ B2 │
└────┴────┘

验收: 子节点正确分布在四格，内容居中或对齐
```

### 4.2 Gap 渲染验收

```
预期 (gap=1, 无边框):
A1 │ A2
───┼───
B1 │ B2

验收: 格子之间有 1 字符间隔
```

### 4.3 Flex 布局验收

```
预期 (cols=[Flex(1), Flex(2)], width=30):
┌────┬──────┐
│ 10 │  20  │
└────┴──────┘

验收: 第二列宽度是第一列的 2 倍
```

---

## 5. 边框渲染验收

### 5.1 容器边框验收

| 样式 | 预期字符 | 验收标准 |
|------|-----------|-----------|
| **Single** | `┌─┐│└─┘├─┤┬┴┼` | 正确显示所有边框字符 |
| **Double** | `╔═╗║╚═╝╠═╣╦╩╬` | 正确显示双线字符 |
| **带标题** | 顶部显示标题文本 | 标题居中对齐 |

### 5.2 Cell 边框验收

| 样式 | 预期字符 | 验收标准 |
|------|-----------|-----------|
| **Single** | `│─┌┐└┘┼┬┴├┤` | 单线格子边框 |
| **Double** | `║═╔╗╚╝╬╦╩╠╣` | 双线格子边框 |
| **Light** | `│─` | 轻边框字符 |
| **Rounded** | `╭╮╰╯` | 圆角字符用于四角 |

### 5.3 混合边框验收

```
预期 (容器单线 + Cell 双线):
┌────────────────────────┐
│        Table           │  ← 容器边框 (Single)
│┌───────┬───────┐       │
││ Name  │ Age   │       │  ← Cell 边框 (Double)
│├───────┼───────┤       │
││ Alice │ 30    │       │
│└───────┴───────┘       │
└────────────────────────┘
```

**验收标准：**
- 容器边框在外层，Cell 边框在内层
- 两者互不干扰，独立渲染
- 交点字符正确，无重叠或错位

### 5.4 边框尺寸验收

| 场景 | 验收标准 | 测试方法 |
|------|-----------|-----------|
| **无边框** | 尺寸 = 内容 | - |
| **单 Cell 边框** | 尺寸 = 内容 + 边框占位 | Measure 返回值 +1 |
| **容器边框** | 尺寸 = 内容 + 外边框占位 | Measure 返回值 +2 |
| **混合边框** | 尺寸 = 内容 + 两层占位 | Measure 返回值检查 |

---

## 6. 坐标与定位验收

### 6.1 坐标系统验收

| 坐标项 | 计算公式 | 验收标准 |
|--------|-----------|-----------|
| **contentX** | `bounds.X + padding[3]` | contentX >= bounds.X |
| **contentY** | `bounds.Y + padding[0]` | contentY >= bounds.Y |
| **格子位置** | `contentX + Σ(colWidths + gaps) + borderOffset` | 正确放置格子 |

### 6.2 SetBounds 验收

| 调用 | 验收标准 |
|------|-----------|
| `SetBounds(5, 10, 100, 50)` | bounds 准确设置 |
| 后续 `GetBounds()` | 返回 (5, 10, 100, 50) |
| 子节点位置 | 基于内内容区域正确计算 |
| **colWidths 重新计算** | 根据 SetBounds 的宽度重新分配列宽（考虑 cellBorders） |
| **右边框位置正确** | 最后边框位置 = GridWidth - 1 |

**关键场景 - 宽度变化时 colWidths 重新分配**：
```
Grid 初始: width=79, cols=[Flex(1), Flex(1), Flex(1)], 启用 cellBorders
Measure 返回: colWidths=[25, 25, 25]

SetBounds(79, ...) ✅
→ colWidths=[25, 25, 25] (与 Measure 相同)
→ 最后边框 x=78 = GridWidth-1 ✓

SetBounds(80, ...) ✅
→ colWidths=[26, 26, 26] (重新分配: (80-4)/3 = 25.33 → 25+1=26, 25, 25)
→ 最后边框 x=79 = GridWidth-1 ✓

SetBounds(60, ...) ✅
→ colWidths=[19, 19, 18] (重新分配: (60-4)/3 = 18.66 → 19, 19, 18)
→ 最后边框 x=59 = GridWidth-1 ✓
```

### 6.3 边框占位验收

```
预期布局 (边框占 1 字符):

[paddingTop]
[b] [格子内容区] [b]
[paddingBottom]

验收:
- 边框字符位于格子边界
- 内容不与边框重叠
- 边框占位包含在总尺寸中
```

---

## 7. 集成测试验收

### 7.1 Fiber-first 集成验收

| 测试场景 | 验收标准 |
|----------|-----------|
| **完整渲染流程** | 通过 `NewDeclarativeNodeFromFuncWithFiber` 渲染成功 |
| **约束传播** | 约束正确从 LayoutEngine → Fiber → Instance |
| **布局计算** | `runtime/layout/Grid` 正确执行计算 |
| **绘制输出** | `PaintEngine` 正确绘制到屏幕 |

### 7.2 复杂布局验收

| 场景 | 验收标准 |
|------|-----------|
| **嵌套 Grid** | Grid 内嵌 Grid，约束正确传递 |
| **混合组件** | Grid 包含 Text、Button 等组件 |
| **动态更新** | 修改 Grid 属性后正确更新 Fiber 树 |

---

## 8. 性能验收

### 8.1 测量性能

| 指标 | 目标 | 测量方法 |
|------|------|-----------|
| **Measure 时间** | < 100ms (100 格子) | benchmark 测试 |
| **约束传递开销** | 可忽略 (< 1ms) | 对比有无追踪 |

### 8.2 内存验收

| 指标 | 目标 |
|------|------|
| **VNode 内存** | 符合典型组件大小 |
| **Instance 内存** | 无内存泄漏 |
| **Fiber 节点** | 正确复用，无循环引用 |

---

## 9. 追踪系统集成

### 9.1 集成要求

Grid 组件应与 `runtime/layout/tracer.go` 集成，提供完整的约束传播追踪能力。

### 9.2 追踪点设计

| 追踪点 | 位置 | 追踪内容 |
|--------|------|-----------|
| **Instance.Measure 入口** | `grid.Instance.Measure(constraints)` | 输入约束 |
| **计算列宽度** | 委托 `layout.Grid.ComputeColumnWidths` | 约束传递 + 结果 |
| **计算行高度** | 委托 `layout.Grid.ComputeRowHeights` | 约束传递 + 结果 |
| **测量子节点** | 遍历格子测量每个子项 | 子节点约束 + 尺寸 |
| **Instance.Measure 出口** | 返回 `size` | 最终尺寸 + 原因 |

### 9.3 追踪集成代码示例

```go
// ui/components/grid/instance.go

package grid

import (
    "github.com/wwsheng009/mint/runtime/layout"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 记录入口
    layout.TraceMeasuring(
        "parent",                    // from
        inst.fiberId,                 // to (grid 节点 ID)
        "root/grids/"+inst.key,      // path
        constraints,                 // input
        layout.Constraints{},        // output (待填充)
        layout.Size{},               // dimension (待填充)
        "Grid.Measure entrance",     // reason
    )

    // 委托给 layout.Grid 计算布局
    resultSize := layout.ComputeGridLayout(
        layout.GridConfig{
            Columns:  inst.columns,
            Rows:     inst.rows,
            Cells:    inst.cells,
            Gap:      layout.Gap{Column: inst.columnGap, Row: inst.rowGap},
            Padding:  inst.padding,
            Borders: layout.BorderConfig{
                ShowContainerBorder: inst.border != nil,
                ShowCellBorders:    inst.showCellBorders,
            },
        },
        constraints,
    )

    // 记录出口
    layout.TraceMeasuring(
        inst.fiberId,                // from
        "parent",                    // to
        "root/grids/"+inst.key,      // path
        constraints,                 // input
        resultSize.OutputConstraints, // output
        resultSize.Dimension,         // dimension
        "Grid.Measure complete",     // reason
    )

    return resultSize.Dimension
}
```

### 9.4 追踪输出示例

```
╔══════════════════════════════════════════════════════════════════╗
║                    Constraint Propagation Trace               ║
╚══════════════════════════════════════════════════════════════════╝

Step 0
  Path: root/grids/my-table
  parent → grid-123
  Input:    {0..100} × {0..50}
  Output:   {20..60} × {10..30}
  Dimension: 60w × 30h
  Reason:   Grid.Measure entrance

Step 1
  Path: root/grids/my-table/col-0
  grid-123 → flex-adapter
  Input:    {0..100} × {0..50}
  Output:   {20..20} × {0..50}
  Dimension: 20w × 10h
  Reason:   Column 0 (Flex factor 1)

Step 2
  Path: root/grids/my-table/col-1
  grid-123 → flex-adapter
  Input:    {0..100} × {0..50}
  Output:   {40..40} × {0..50}
  Dimension: 40w × 10h
  Reason:   Column 1 (Flex factor 2)

...
```

### 9.5 调试验收

| 场景 | 验收标准 |
|------|-----------|
| **启用追踪** | `layout.EnableTracer()` 后记录约束传播 |
| **禁用追踪** | `layout.DisableTracer()` 后无性能影响 |
| **追踪输出** | `layout.DumpTrace()` 返回完整的约束传播日志 |
| **测试验证** | 单元测试可读取 `layout.GetTraceEntries()` 验证 |

---

## 10. 验收检查清单

### 10.1 开发阶段检查清单

- [ ] VNode API 实现完整（列/行/格子/间距/边框）
- [ ] Instance 测量逻辑正确（Fixed/Flex/Auto）
- [ ] 委托 `runtime/layout/Grid` 计算布局
- [ ] 边框数据流正确（VNode → Instance → Paint）
- [ ] 坐标系统基于 contentX/contentY 统一

### 10.2 测试阶段检查清单

- [ ] 单元测试覆盖率 >= 80%
- [ ] VNode API 测试全部通过
- [ ] Instance 测量测试全部通过
- [ ]边框渲染测试全部通过
- [ ] 追踪集成测试通过

### 10.3 集成测试检查清单

- [ ] Fiber-first 完整渲染流程通过
- [ ] 约束传播追踪正确
- [ ] 嵌套 Grid 布局正确
- [ ] 动态更新正确刷新
- [ ] 绘制输出可视化正确

### 10.4 性能验收检查清单

- [ ] Measure 性能达标 (< 100ms / 100 格子)
- [ ] 无内存泄漏
- [ ] 禁用追踪后无性能损失

### 10.5 文档验收检查清单

- [ ] 架构文档 (`ARCHITECTURE.md`) 完整
- [ ] 设计文档 (`cell_borders_design.md`) 完整
- [ ] 验收标准 (`ACCEPTANCE.md`) 完整
- [ ] 代码注释清晰

---

## 附录：验收命令

### 启用追踪

```go
import "github.com/wwsheng009/mint/runtime/layout"

// 启用追踪
layout.EnableTracer()

// 运行渲染
renderGrid()

// 输出追踪日志
println(layout.DumpTrace())

// 清除追踪数据
layout.ClearTrace()
```

### 单元测试

```bash
# 运行 Grid 全部测试
go test ./ui/components/grid/... -v

# 运行边框测试
go test ./ui/components/grid/... -run TestVNode_CellBorders -v

# 运行测量测试
go test ./ui/components/grid/... -run TestInstance_Measure -v
```

### 性能测试

```bash
# 运行 benchmark
go test ./ui/components/grid/... -bench=. -benchmem
```
