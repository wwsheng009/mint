# Grid 组件设计架构文档

## 0. 实施前提要求

### 0.1 核心概念理解

#### BoxConstraints（盒约束）
```go
// runtime.BoxConstraints
type BoxConstraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}

// Infinity 表示无界约束
const Infinity = 1 << 30
```

**约束类型：**
- **紧约束 (TightConstraints)**: `MinWidth == MaxWidth && MinHeight == MaxHeight`
  - 组件必须使用确切尺寸，没有灵活性
- **松约束 (LooseConstraints)**: `MinWidth = 0, MaxWidth = maxWidth`
  - 组件可以在范围内选择自己的尺寸
- **无界约束**: `MaxWidth = Infinity, MaxHeight = Infinity`
  - 组件根据内容决定尺寸

#### 四个核心阶段

```
VNode → Fiber → LayoutBox → PaintableBox
 (声明)  (调度)   (布局)      (绘制)
```

**各阶段产物：**
| 阶段 | 数据结构 | 职责 | 生命周期 |
|------|----------|------|----------|
| VNode | `grid.VNode` | 声明式描述，无状态 | 每次渲染重新创建 |
| Fiber | `*Fiber` | 调度单元，包含 Instance | 跨渲染持久化 |
| LayoutBox | `layout.LayoutBox` | 布局结果 (x,y,w,h) | 每次布局计算 |
| PaintableBox | `PaintableBox` interface | 绘制接口 | Paint 阶段使用 |

### 0.2 Grid 实施前的依赖检查

| 依赖项 | 要求 | 检查点 |
|--------|------|--------|
| **runtime.BoxConstraints** | 约束传递机制 | MinWidth ≤ Width ≤ MaxWidth |
| **runtime.Measurable** | Instance 实现 Measure() | 接收 BoxConstraints，返回 Size |
| **runtime/layout.Grid** | 布局计算引擎 | 提供列宽/行高计算 |
| **runtime/layout.Node** | Node 接口 | GetID(), GetType(), Children() |
| **rtui.ComponentInstance** | 实例接口 | SetProps(), SetChildBounds() |
| **rtui.PaintableInstance** | 绘制接口 | Paint(x,y) → []DrawCmd |
| **internal/reconciler** | VNode → Fiber 转换 | completeWork 同步属性 |
| **internal/render.FiberToNodeAdapter** | Fiber → layout.Node 转换 | Measure() 调用 Instance.Measure() |

### 0.3 两阶段的约束系统

#### 阶段 1: Instance.Measure() - 内部测量
```go
// grid/Instance.Measure()
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // 输入：来自 Layout Engine 的 layout.Constraints
    // 输出：layout.Size (理想尺寸)

    // 1. 创建 runtime/layout.Grid 实例
    gridStyle := inst.GetGridStyle()
    gridLayout := layout.NewGridLayout("ui-grid", gridStyle)

    // 2. 转换约束（layout.Constraints -> runtime.BoxConstraints）
    //     注意：这里是 layout 包内部的 Constraints，不是 runtime.BoxConstraints
    bc := runtime.BoxConstraints{
        MinWidth:  constraints.MinWidth,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: constraints.MinHeight,
        MaxHeight: constraints.MaxHeight,
    }

    // 3. 调用布局引擎测量
    size := gridLayout.Measure(constraints)

    // 4. 保存计算结果
    inst.colWidths = gridLayout.GetColumnWidths()
    inst.rowHeights = gridLayout.GetRowHeights()

    return size
}
```

#### 阶段 2: Instance.SetBounds() - 外部布局
```go
// grid/Instance.SetBounds()
func (inst *Instance) SetBounds(x, y, w, h int) {
    // 输入：父容器分配的绝对位置和尺寸
    // 输出：设置子节点的 bounds

    inst.bounds = [4]int{x, y, w, h}

    // 计算每个子节点的绝对位置
    for i, cell := range inst.cells {
        // 使用 runtime/layout.Grid.GetPosition() 获取相对位置
        relX, relY := gridLayout.GetPosition(cell.Row, cell.Col)

        // 转换为绝对坐标
        absX := x + padding[3] + relX
        absY := y + padding[0] + relY

        // 设置子节点 bounds
        inst.childBounds[i] = [4]int{absX, absY, cw, ch}
    }
}
```

#### 约束转换关系

```
runtime.BoxConstraints (层级最高)
    ↓
internal.render.LayoutEngine.Layout() 内部
    ↓
layout.Constraints (runtime/layout 包内)
    ↓
runtime.layout.Grid.Measure()
    ↓
layout.Size (返回结果)
    ↓
internal.render.LayoutEngine
    ↓
layout.LayoutBox
    ↓
PaintableBox 获取 bounds
```

---

## 1. 职责分离与架构分层

```
┌─────────────────────────────────────────────────────────────────┐
│                         用户代码层                               │
│                    grid.New().AddCell(...)                      │
└───────────────────────┬─────────────────────────────────────────┘
                        │ (VNode - 声明式描述)
                        ↓
┌─────────────────────────────────────────────────────────────────┐
│                    VNode 层 (vnode.go)                          │
│  - 纯 declarative 描述                                           │
│  - 无状态、无闭包、无 Paint 逻辑                                 │
│  - 包含布局定义：columns, rows, cells                            │
│  - 包含样式配置：padding, gap, alignContent                      │
│  - 包含边框配置：borderStyle, showCellBorders                    │
└───────────────────────┬─────────────────────────────────────────┘
                        │ (Props map)
                        ↓
┌─────────────────────────────────────────────────────────────────┐
│                  Reconciler 层 (协调过程)                       │
│  - VNode → Props 转换                                           │
│  - Instance 创建/更新                                           │
│  - Props 同步到 Instance 字段                                   │
└───────────────────────┬─────────────────────────────────────────┘
                        │
                        ↓
┌─────────────────────────────────────────────────────────────────┐
│              Instance 层 (instance.go)                          │
│  - 运行时实体，跨渲染持久化                                     │
│  - 持有：bounds, colWidths, rowHeights, childBounds             │
│  - 实现：ComponentInstance, PaintableInstance, Measure()        │
│  - 委托布局计算给 runtime/layout/Grid                           │
└────────────┬────────────────────────────────────────────────────┘
             │
             ├───────────────────────┬──────────────────────┐
             ↓                       ↓                      ↓
┌──────────────────────┐  ┌──────────────────┐  ┌─────────────────────┐
│ 布局计算委托          │  │  边框绘制          │  │  状态管理            │
│ runtime/layout/Grid  │  │  cell_borders...  │  │  bounds, dirty       │
└──────────────────────┘  └──────────────────┘  └─────────────────────┘
```

---

## 1.1 各层级职责详解

### VNode 层 (声明层)

**职责：**
- 纯声明式描述，无任何状态
- 嵌入 `*rtui.ElementVNode` (返回 `VNodeElement` 类型)
- 提供 ToProps() 生成 Props map 用于 reconciler

**关键方法：**
```go
// VNode 实现 rtui.InstanceFactory
func (g *VNode) CreateInstance() rtui.ComponentInstance {
    // 将 VNode 的所有布局和样式属性转换为 Props map
    props := rtui.Props{
        "columns": g.columns,
        "rows": g.rows,
        "cells": g.cells,
        "borderStyle": g.borderStyle,
        "showCellBorders": g.showCellBorders,
        // ...
    }
    return Instance.NewInstance(props)
}

// VNode.Type() 返回 VNodeElement (通过嵌入的 ElementVNode)
// Grid 不覆盖 Type() 方法，因此走 Element 路径
```

### Fiber 层 (调度层)

**职责：**
- 跨渲染持久化的树结构
- 持有 Instance 引用（Fiber-first 架构的关键）
- 存储布局属性：`LayoutDirection`, `LayoutPadding`, `BorderStyle` 等
- 标记 dirty flags：`FlagLayoutDirty`, `FlagPaintDirty`

**关键字段：**
```go
type Fiber struct {
    // 树结构
    Return, Child, Sibling *Fiber

    // Instance 引用（运行时实体）
    Instance ComponentInstance  // 持久化，不随 VNode 重建

    // 布局属性（Phase 1 从 VNode 复制）
    LayoutDirection  Direction
    LayoutPadding    [4]int
    LayoutGap        int
    BorderStyle      string   // 容器边框样式

    // Props 测量（来自 VNode.ToProps()）
    Props Props
    MemoizedProps Props
}
```

**属性同步流程：**
```
VNode.ToProps()
    ↓
reconciler: createWorkInProgress()
    ↓
Fiber.Props = node.Props
    ↓
reconciler: completeWorkElement()
    ↓
syncBorderProperties()  // 将 Props["borderStyle"] -> Fiber.BorderStyle
```

### LayoutBox 层 (布局层)

**职责：**
- 布局计算结果，记录 `X, Y, Width, Height`
- 由 Layout Engine 生成
- 相对于父容器的坐标

**数据结构：**
```go
type LayoutBox struct {
    ID string
    X, Y int           // 相对位置
    Width, Height int  // 布局尺寸
    Baseline int       // 基线（文本对齐）
    Layer Layer         // 渲染层
    ZIndex int         // 层内排序
    Border Border       // 边框信息
    Children []*LayoutBox
}
```

### PaintableBox 层 (绘制层)

**职责：**
- 提供 `GetBounds()` 和 `GetChildren()` 方法
- Paint 阶段访问 Instance.Paint() 获取 DrawCmd
- 使用绝对坐标进行绘制

**接口定义：**
```go
// internal/render/layout_switcher.go
type PaintableBox interface {
    GetBounds() (x, y, width, height int)
    GetChildren() []PaintableBox
}
```

---

## 2. 核心数据流

### 2.1 测量流程详解

#### 测量触发路径

```
用户初始化应用 (NewDeclarativeNodeFromFuncWithFiber)
    ↓
Runtime.Render() 或 Scheduler.ScheduleWork()
    ↓
Reconciler.BeginWork() (Fiber 协调)
    ↓
LayoutPhase: MeasurePass
    ↓
LayoutEngine.LayoutFiber(fiber, constraints)  ⚠️ Fiber-first: 不涉及 VNode
    ↓
[关键 FiberToNodeAdapterPure] 将 Fiber 转换为 layout.Node
    ↓
调用 Node.Measure(constraints)
    ↓
FiberToNodeAdapterPure.Measure() 调用:
    1. fiber.GetInstance()
    2. inst.(Measurable).Measure(bc)
    ↓
grid.Instance.Measure(constraints)
```

**重要说明：**

| 模式 | 调用方法 | 参数 | 说明 |
|------|----------|------|------|
| **Fiber-First** (新) | `LayoutFiber(fiber, constraints)` | Fiber + constraints | 纯 Fiber 架构 |
| **Legacy** (旧) | `Layout(vnode, fiber, constraints)` | VNode + Fiber + constraints | 兼容旧代码 |

Grid 组件应该完全基于 **Fiber-First** 架构，只使用 `LayoutFiber(fiber, constraints)`。

#### Instance.Measure() 的完整实现

```go
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
    // === 输入验证 ===
    if inst == nil {
        return layout.Size{}
    }

    // === 第 1 步：构建 GridStyle ===
    gridStyle := inst.GetGridStyle()
    // GridStyle 包含：
    //   - Columns, Rows, Cells
    //   - ColumnGap, RowGap, Padding
    //   - ShowCellBorders (影响尺寸计算)

    // === 第 2 步：创建布局引擎实例 ===
    gridLayout := layout.NewGridLayout("ui-grid", gridStyle)

    // === 第 3 步：执行测量 ===
    // runtime/layout/Grid.Measure(constraints) 内部：
    //   1. 计算边框占用:
    //      cellBorderWidth = numCols + 1 (有边框时)
    //      cellBorderHeight = numRows + 1
    //   2. 计算可用空间:
    //      availableWidth = MaxWidth - paddingLR - cellBorderWidth
    //   3. 计算列宽:
    //      - Fixed: 直接使用
    //      - Flex: 分配剩余空间
    //      - Auto: 根据内容测量（这里简化为最小宽度）
    //   4. 计算行高:
    //      - Auto: 测量子节点内容
    //      - Fixed: 直接使用
    //   5. 返回总尺寸:
    //      Width  = Σ(colWidths) + Σ(colGaps) + cellBorderWidth + paddingLR
    //      Height = Σ(rowHeights) + Σ(rowGaps) + cellBorderHeight + paddingTB

    size := gridLayout.Measure(constraints)

    // === 第 4 步：保存计算结果 ===
    inst.colWidths = gridLayout.GetColumnWidths()
    inst.rowHeights = gridLayout.GetRowHeights()

    return size
}
```

### 2.2 高度宽度限制机制

#### 约束应用顺序

```go
// 1. 父容器传递 BoxConstraints
parentConstraints := runtime.BoxConstraints{
    MinWidth:  0,
    MaxWidth:  100,
    MinHeight: 0,
    MaxHeight: 24,
}

// 2. LayoutEngine 将约束转为 layout.Constraints
layoutConstraints := layout.Constraints{
    MinWidth:  0,
    MaxWidth:  100,
    MinHeight: 0,
    MaxHeight: 24,
}

// 3. Grid 测量时考虑约束
func (g *GridLayout) Measure(constraints Constraints) Size {
    // === 宽度限制 ===
    availableWidth := constraints.MaxWidth - paddingLR - cellBorderWidth

    // Flex 分配剩余空间
    remainingWidth := availableWidth - fixedWidth - gapWidth
    for _, flexCol := range flexColumns {
        // 每个分配：remainingWidth * factor / totalFactor
        flexWidth := (remainingWidth * flexFactor) / totalFlexFactor
        widths[i] = flexWidth
    }

    // Clamp 到约束范围
    totalWidth = Σ(widths) + gaps + cellBorderWidth + paddingLR
    if totalWidth > constraints.MaxWidth {
        totalWidth = constraints.MaxWidth  // 超出则截断
    }
    if totalWidth < constraints.MinWidth {
        totalWidth = constraints.MinWidth  // 不足则扩展
    }

    // === 高度限制 ===
    totalHeight = Σ(rowHeights) + gaps + cellBorderHeight + paddingTB
    if totalHeight > constraints.MaxHeight {
        totalHeight = constraints.MaxHeight
    }
    if totalHeight < constraints.MinHeight {
        totalHeight = constraints.MinHeight
    }

    return Size{Width: totalWidth, Height: totalHeight}
}
```

#### 边框占用空间的限制

```go
// 有 Cell Borders 时，边框占用计算：
if ShowCellBorders {
    // 每条垂直线占 1 字符宽度
    cellBorderWidth = numCols + 1  // 左 + 中间 + 右

    // 每条水平线占 1 字符高度
    cellBorderHeight = numRows + 1 // 上 + 中间 + 下
}

// 示例：3列 x 2行，padding=[1,1,1,1]，gap=0
// colWidths = [20, 20, 20], rowHeights = [5, 5]

totalWidth = 1 (左padding) + 1 (左边框)          // = 2
          + 20 + 1 (右边框)                      // = 23
          + 20 + 1 (右边框)                      // = 44
          + 20 + 1 (右边框)                      // = 65
          + 1 (右padding)                       // = 66

totalHeight = 1 (上padding) + 1 (上边框)        // = 2
           + 5                                 // = 7
           + 1 (中间边框)                       // = 8
           + 5                                 // = 13
           + 1 (下边框)                         // = 14
           + 1 (下padding)                      // = 15
```

### 2.3 VNode → Fiber 转换的完整流程

```go
// internal/reconciler/create_fiber.go

func createWorkInProgress(current *Fiber, pendingProps Props) *Fiber {
    // 1. 创建或复用 Fiber
    var workInProgress *Fiber
    if current != nil {
        // 复用现有 Fiber
        workInProgress = current
    } else {
        // 创建新 Fiber
        workInProgress = &Fiber{}
    }

    // 2. 设置 Props
    workInProgress.Props = pendingProps

    // 3. 复用 Instance（关键：Fiber-first）
    // Instance 是持久化的，不重新创建
    if workInProgress.Instance == nil {
        // 首次创建 Instance
        if factory, ok := node.(InstanceFactory); ok {
            workInProgress.Instance = factory.CreateInstance()
        }
    } else {
        // 复用现有 Instance，调用 SetProps 更新
        workInProgress.Instance.SetProps(pendingProps)
    }

    return workInProgress
}

// internal/reconciler/complete_work.go (Element 路径)

func completeWorkElement(fiber *Fiber) error {
    // 1. 同步布局属性
    fiber.LayoutDirection = getDirectionProp(fiber.Props)
    fiber.LayoutPadding = getPaddingProp(fiber.Props)
    fiber.LayoutGap = getIntProp(fiber.Props, "gap", 0)
    fiber.LayoutFlex = getIntProp(fiber.Props, "flex", 0)

    // 2. 同步边框属性（Grid 走这里）
    syncBorderProperties(fiber)

    return nil
}

func syncBorderProperties(fiber *Fiber) {
    if fiber == nil {
        return
    }

    // 从 Props 读取边框样式
    if style, ok := fiber.Props["borderStyle"].(string); ok {
        fiber.BorderStyle = style
    } else {
        fiber.BorderStyle = "none"  // 默认无边框
    }

    if label, ok := fiber.Props["label"].(string); ok {
        fiber.BorderLabel = label
    }
}
```

### 2.4 声明 → 渲染流程

```
1. 用户声明
   grid := grid.New().
       Columns(grid.Flex{1}, grid.Fixed(20)).
       Rows(grid.Auto{}).
       AddCell(text.V("Hello"), 0, 0).
       ShowCellBorders()
       ↓
2. VNode 结构
   {
     columns: [Flex{1}, Fixed(20)],
     rows: [Auto{}],
     cells: [{Row:0, Col:0, Child: text.V("Hello")}],
     showCellBorders: true,
     cellBorderStyle: "single"
   }
   ↓
3. VNode.ToProps()
   {
     "columns": [...],
     "rows": [...],
     "cells": [...],
     "showCellBorders": true,
     "cellBorderStyle": "single"
   }
   ↓
4. Instance.NewInstance(props)
   - 解析 props 存储到实例字段
   - 初始化：dirty=true
   ↓
5. Layout Engine 调度
   - Instance.Measure(constraints)
   ↓
6. Instance.Measure() {
     // 委托给布局引擎
     gridStyle := inst.GetGridStyle()
     gridLayout := layout.NewGridLayout("ui-grid", gridStyle)
     size := gridLayout.Measure(constraints)

     // 获取计算结果
     inst.colWidths = gridLayout.GetColumnWidths()
     inst.rowHeights = gridLayout.GetRowHeights()
     return size
   }
   ↓
7. runtime/layout/Grid.Measure()
   - 计算列宽、行高
   - 考虑 padding、gap、cellBorders
   - 返回总的 size + 布局结果
   ↓
8. Instance.SetBounds(x, y, w, h)
   - 更新 inst.bounds = [x, y, w, h]
   - 计算子节点位置（通过 runtime/layout.Grid.GetPosition()）
   - 设置子节点的 bounds
   ↓
9. Paint 阶段
   - Instance.Paint(x, y) 返回 DrawCmd 列表
   - 绘制 cell borders (如果启用了 showCellBorders)
```

### 2.2 布局计算细节

```
可用空间 = constraints.MaxWidth - padding 左右
         ↓
┌─────────────────────────────────────────────────────┐
│  runtime/layout/Grid 计算                           │
├─────────────────────────────────────────────────────┤
│  1. 计算列宽                                        │
│     - Fixed：直接使用指定宽度                       │
│     - Flex：分配剩余空间                            │
│     - Auto：根据内容测量                            │
│  2. 计算行高                                        │
│     - Auto：根据子节点 Measure()                    │
│     - Fixed：直接使用                               │
│  3. 计算总尺寸                                      │
│     TotalWidth  = Σ(colWidths) + Σ(gap) + padding   │
│     TotalHeight = Σ(rowHeights) + Σ(gap) + padding  │
│                   + (numRows+1) * borderHeight     │
│  4. 计算总高度（如果有 cellBorders）               │
│     BorderHeight = (numRows + 1) * 1               │
└─────────────────────────────────────────────────────┘
```

---

## 3. 实施检查清单

### 3.1 实施前的依赖验证

在开始实施 Grid 组件之前，请确保以下所有依赖项都已就绪：

#### 基础类型系统

| 检查项 | 验证方法 | 预期结果 |
|--------|----------|----------|
| `runtime.BoxConstraints` | 查看代码 | 定义了 MinWidth, MaxWidth 等 |
| `runtime.Size` | 查看代码 | 定义了 Width, Height |
| `runtime.Infinity`常量 | 查看代码 | 值为 `1 << 30` |
| `layout.Constraints` | 查看代码 | 与 runtime.BoxConstraints 相同结构 |
| `layout.Size` | 查看代码 | 定义了 Width, Height |

#### 接口系统

| 检查项 | 验证方法 | 预期结果 |
|--------|----------|----------|
| `runtime.Measurable` | 查看 code | Measure(BoxConstraints) Size |
| `rtui.ComponentInstance` | 查看 code | SetProps(), SetChildBounds() |
| `rtui.PaintableInstance` | 查看 code | Paint(x,y) []DrawCmd |
| `rtui.InstanceFactory` | 查看 code | CreateInstance() ComponentInstance |
| `layout.Node` | 查看 code | GetID(), GetType(), Children(), Measure() |

#### Fiber 系统

| 检查项 | 验证方法 | 预期结果 |
|--------|----------|----------|
| `Fiber` 类型定义 | 查看 code | 包含 Instance, Props, BorderStyle |
| `Fiber.BorderStyle` | 查看 code | string 类型 |
| `createWorkInProgress()` | 查看 code | 创建/复用 Fiber 和 Instance |
| `completeWorkElement()` | 查看 code | 调用 syncBorderProperties() |
| `syncBorderProperties()` | 查看 code | Props → Fiber.BorderStyle |

#### Layout 系统

| 检查项 | 验证方法 | 预期结果 |
|--------|----------|----------|
| `runtime/layout.Grid` | 查看 code | GridLayout 类型 |
| `GridLayout.Measure()` | 查看 code | 接受 Constraints，返回 Size |
| `GridLayout.GetColumnWidths()` | 查看 code | 返回 []int |
| `GridLayout.GetRowHeights()` | 查看 code | 返回 []int |
| `GridLayout.GetPosition()` | 查看 code | row, col → x, y |
| `GridStyle` | 查看 code | 包含 ShowCellBorders |

#### 渲染系统

| 检查项 | 验证方法 | 预期结果 |
|--------|----------|----------|
| `FiberToNodeAdapter` | 查看 code | Fiber → layout.Node 转换 |
| `FiberToNodeAdapter.Measure()` | 查看 code | 调用 Instance.Measure() |
| `FiberToNodeAdapter.GetBorder()` | 查看 code | 读取 Fiber.BorderStyle |
| `PaintableBox` | 查看 code | GetBounds(), GetChildren() |
| `paint.DrawCmd` | 查看 code | 绘制命令类型 |

### 3.2 约束转换链验证

```
用户代码 → Runtime.Render()
    ↓
Runtime.BoxConstraints {MinWidth:0, MaxWidth:100, ...}
    ↓
LayoutEngine.LayoutFiber(fiber, constraints)  ⚠️ Fiber-first: 不涉及 VNode
    ↓
layout.Constraints {MinWidth:0, MaxWidth:100, ...}
    ↓
FiberToNodeAdapterPure.Measure(constraints)
    ↓
Instance.Measure(constraints)
    ↓
layout.Size {Width:XX, Height:YY}
    ↓
LayoutEngine 返回 LayoutResult
    ↓
PaintableBox.GetBounds()
```

**验证方法：** 在每个转换点添加 `fmt.Printf()` 打印约束值，确保传递正确。

### 3.3 边框数据流验证

#### 容器边框

```
VNode.borderStyle (声明)
    ↓
VNode.ToProps() → Props["borderStyle"]
    ↓
Fiber.Props = Props
    ↓
completeWorkElement() → syncBorderProperties()
    ↓
Fiber.BorderStyle (持久化)
    ↓
FiberToNodeAdapter.GetBorder()
    ↓
绘制阶段读取
```

#### 格子边框

```
VNode.showCellBorders (声明)
    ↓
VNode.ToProps() → Props["showCellBorders"]
    ↓
Instance.NewInstance(props)
    ↓
Instance.showCellBorders (持久化)
    ↓
Instance.GetGridStyle() → GridStyle.ShowCellBorders
    ↓
GridLayout.Measure() 考虑边框占用
    ↓
Instance.Paint() 检查 showCellBorders
    ↓
GenCellBorderDrawCmds() 生成绘制命令
```

### 3.4 关键方法签名清单

#### VNode 层

```go
// 必须实现
func (g *VNode) CreateInstance() rtui.ComponentInstance  // InstanceFactory
func (g *VNode) ToProps() rtui.Props                      // 属性转换
func (g *VNode) SetShowCellBorders(show bool) *VNode     // API 方法
```

#### Instance 层

```go
// 必须实现
func NewInstance(props rtui.Props) *Instance
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size
func (inst *Instance) SetBounds(x, y, w, h int)
func (inst *Instance) Paint(x, y int) []paint.DrawCmd
func (inst *Instance) GetGridStyle() *layout.GridStyle  // GridStyleProvider
func (inst *Instance) SetProps(props rtui.Props)
func (inst *Instance) SetChildBounds(id interface{}, bounds [4]int)
```

#### runtime/layout/Grid

```go
// 必须提供
func (g *GridLayout) Measure(constraints Constraints) Size
func (g *GridLayout) GetColumnWidths() []int
func (g *GridLayout) GetRowHeights() []int
func (g *GridLayout) GetCellPosition(row, col int) [2]int
func (g *GridLayout) GetCellSize(row, col int) [2]int
```

---

## 4. 边框实现详解

### 4.1 两类边框的对比

| 特性 | 容器边框 (Outer Border) | 格子边框 (Cell Borders) |
|------|-------------------------|--------------------------|
| **用途** | 绘制整个 Grid 的外边框 | 绘制格子之间的分隔线 |
| **API** | `SetBorderStyle("single")` | `ShowCellBorders()` |
| **存储位置** | Fiber.BorderStyle | Instance.showCellBorders |
| **同步路径** | reconciler.syncBorderProperties() | VNode.ToProps() → Instance |
| **绘制时机** | container 布局阶段 | Instance.Paint() 阶段 |
| **坐标参考** | Grid 整体 bounds | 基于 colWidths/rowHeights |
| **边框字符** | 标准边框字符 | 标准/双线/轻边框字符 |

### 4.2 容器边框实现

#### 步骤 1: VNode 声明边框

```go
// ui/components/grid/vnode.go

type VNode struct {
    *rtui.ElementVNode
    // 容器边框配置
    borderStyle  string // "none", "single", "double", "rounded", "dashed"
    borderLabel  string // 可选标签
}

func (g *VNode) SetBorderStyle(style string) *VNode {
    g.borderStyle = style
    return g
}
```

#### 步骤 2: Props 转换

```go
// VNode.ToProps()
func (g *VNode) ToProps() rtui.Props {
    return rtui.Props{
        // ... 其他属性
        "borderStyle": g.borderStyle,
        "label":       g.borderLabel,
    }
}
```

#### 步骤 3: Reconciler 同步边框属性

```go
// internal/reconciler/complete_work.go

func syncBorderProperties(fiber *Fiber) {
    // 从 Props 读取
    if style, ok := fiber.Props["borderStyle"].(string); ok {
        fiber.BorderStyle = style  // 赋值到 Fiber.BorderStyle
    } else {
        fiber.BorderStyle = "none"
    }

    if label, ok := fiber.Props["label"].(string); ok {
        fiber.BorderLabel = label
    }
}
```

#### 步骤 4: 渲染层读取边框

```go
// internal/render/fiber_adapter.go

type FiberToNodeAdapter struct {
    fiber *Fiber
}

func (a *FiberToNodeAdapter) GetBorder() Border {
    // Path A: 优先读取 Fiber.BorderStyle
    borderStyle := a.fiber.BorderStyle

    // Path B: Fallback 到 Props["borderStyle"]
    if borderStyle == "" {
        if style, ok := a.fiber.Props["borderStyle"].(string); ok {
            borderStyle = style
        }
    }

    return Border{
        Style: borderStyle,
        Label: a.fiber.BorderLabel,
    }
}
```

#### 步骤 5: Paint 阶段绘制边框

```go
// container 绘制逻辑（在其他组件中）

func drawBorder(x, y, w, h int, border Border) []paint.DrawCmd {
    if border.Style == "none" {
        return nil
    }

    chars := getBorderChars(border.Style) // ┌─┐ 等
    return []paint.DrawCmd{
        // 上边框
        paint.Text(chars.TopLeft, x, y),
        paint.Repeat(chars.Top, w-2, x+1, y),
        paint.Text(chars.TopRight, x+w-1, y),
        // ... 其他边框
    }
}
```

### 4.3 格子边框 (Cell Borders) 实现

#### 步骤 1: VNode 配置 Cell Borders

```go
// ui/components/grid/vnode.go

type VNode struct {
    // Cell Borders 配置
    showCellBorders   bool   // 是否显示
    cellBorderStyle   string // "none", "single", "double", "light"
    cellBorderRounded bool   // 圆角
    cellBorderColor   string // 颜色
}

func (g *VNode) ShowCellBorders() *VNode {
    g.showCellBorders = true
    g.cellBorderStyle = "single"
    return g
}

func (g *VNode) SetCellBorderStyle(style string) *VNode {
    g.cellBorderStyle = style
    return g
}
```

#### 步骤 2: Props 转换

```go
func (g *VNode) ToProps() rtui.Props {
    return rtui.Props{
        // ... 其他属性
        "showCellBorders":   g.showCellBorders,
        "cellBorderStyle":   g.cellBorderStyle,
        "cellBorderRounded": g.cellBorderRounded,
        "cellBorderColor":   g.cellBorderColor,
    }
}
```

#### 步骤 3: Instance 持有边框配置

```go
// ui/components/grid/instance.go

type Instance struct {
    // Cell Borders 配置
    showCellBorders   bool
    cellBorderStyle   string
    cellBorderRounded bool
    cellBorderColor   string

    // 计算结果（用于绘制）
    colWidths  []int
    rowHeights []int
    bounds     [4]int
}

func NewInstance(props rtui.Props) *Instance {
    return &Instance{
        showCellBorders:   getBoolProp(props, "showCellBorders", false),
        cellBorderStyle:   getStringProp(props, "cellBorderStyle", "single"),
        cellBorderRounded: getBoolProp(props, "cellBorderRounded", false),
        cellBorderColor:   getStringProp(props, "cellBorderColor", ""),
    }
}
```

#### 步骤 4: 布局引擎考虑边框

```go
// runtime/layout/grid.go

func (g *GridLayout) Measure(constraints Constraints) Size {
    numCols := len(g.style.Columns)
    numRows := g.calculateRowCount()

    // === 计算边框占用空间 ===
    cellBorderWidth := 0
    cellBorderHeight := 0
    if g.style.ShowCellBorders {
        cellBorderWidth = numCols + 1   // n+1 条垂直线
        cellBorderHeight = numRows + 1  // n+1 条水平线
    }

    // 从约束中减去边框空间
    availableWidth := constraints.MaxWidth
        - g.style.Padding.Left - g.style.Padding.Right
        - cellBorderWidth

    // ... 计算列宽、行高 ...

    // 总尺寸包含边框
    return Size{
        Width:  totalWidth + cellBorderWidth,
        Height: totalHeight + cellBorderHeight,
    }
}
```

#### 步骤 5: Paint 阶段绘制

```go
// ui/components/grid/instance.go

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
    if !inst.showCellBorders {
        return nil
    }

    // 获取当前 bounds
    bx, by, bw, bh := inst.GetBounds()

    // 获取边框字符集
    chars := cellBorderChars[inst.cellBorderStyle]

    // 生成绘制命令
    return inst.GenCellBorderDrawCmds(x, y)
}

// GenCellBorderDrawCmds 实现（伪代码）
func (inst *Instance) GenCellBorderDrawCmds(offsetX, offsetY int) []paint.DrawCmd {
    var cmds []paint.DrawCmd

    // 计算内容区域起点
    contentX := offsetX + inst.padding[3]  // left padding
    contentY := offsetY + inst.padding[0]  // top padding

    // 绘制水平线
    for row := 0; row <= len(inst.rowHeights); row++ {
        // 计算当前水平线的 Y 坐标
        lineY := contentY
        for i := 0; i < row; i++ {
            lineY += inst.rowHeights[i] + 1  // 行高 + 下边框
        }

        // 绘制整条水平线
        for col := 0; col <= len(inst.colWidths); col++ {
            // 计算当前格子的 X 坐标
            lineX := contentX
            for i := 0; i < col; i++ {
                lineX += inst.colWidths[i] + 1  // 列宽 + 右边框
            }

            // 选择正确的字符（交叉点、边角、边线）
            char := selectHorizontalChar(row, col, len(inst.rowHeights), len(inst.colWidths))
            cmds = append(cmds, paint.Text(char, lineX, lineY))
        }
    }

    // 绘制垂直线（类似逻辑）
    // ...

    return cmds
}
```

### 4.4 边框字符集定义

```go
// cell_borders_paint.go

var cellBorderChars = map[string]BorderChars{
    "single": {
        Top:         "─", Bottom: "─",
        Left:        "│", Right: "│",
        TopLeft:     "┌", TopRight: "┐",
        BottomLeft:  "└", BottomRight: "┘",
        Cross:       "┼",
        TopCross:    "┬", BottomCross: "┴",
        LeftCross:   "├", RightCross: "┤"
    },
    "double": {
        Top:         "═", Bottom: "═",
        Left:        "║", Right: "║",
        TopLeft:     "╔", TopRight: "╗",
        BottomLeft:  "╚", BottomRight: "╝",
        Cross:       "╬",
        TopCross:    "╦", BottomCross: "╩",
        LeftCross:   "╠", RightCross: "╣"
    },
    "light": {
        Top:         "─", Bottom: "─",
        Left:        "│", Right: "│",
        TopLeft:     "┌", TopRight: "┐",
        BottomLeft:  "└", BottomRight: "┘",
        Cross:       "┼",
        TopCross:    "┬", BottomCross: "┴",
        LeftCross:   "├", RightCross: "┤"
    },
}
```

---

## 5. Cell 实现的前提条件

### 5.1 Cell 结构定义

```go
ui/components/grid/vnode.go

type Cell struct {
    Child   rtui.VNode
    Row     int  // 0-based 行索引
    Col     int  // 0-based 列索引
    RowSpan int  // 跨行数（默认 1）
    ColSpan int  // 跨列数（默认 1）
}
```

### 5.2 Cell 实现所需的前提

| 前提条件 | 要求 | 验证方法 |
|----------|------|----------|
| **VNode.Child** | 必须实现 rtui.VNode | `child.Type()` 返回有效类型 |
| **Runtime Grid 支持跨单元格** | GridLayout 必须 RowSpan/ColSpan | GetCellPosition(spanX, spanY) 计算 |
| **布局边界计算** | SetChildBounds() 能设置子节点位置 | 子节点使用绝对坐标 |
| **测量约束传递** | 子节点接收正确约束 | constraints 减去 padding/gap |
| **边框坐标对齐** | Cell 内容在边框内绘制 | offset = +1 (跳过边框线) |

### 5.3 跨行跨列 Cell 的处理

#### 布局引擎支持

```go
// runtime/layout/grid.go

func (g *GridLayout) Measure(constraints Constraints) Size {
    // ... 计算基础列宽和行高 ...

    // === 处理跨单元格 ===
    // 找出跨越的 cell，调整对应的 span 区域

    for _, cell := range g.style.Cells {
        if cell.ColSpan > 1 {
            // 跨列 cell 的宽度 = Σ(colWidths[spanStart:spanEnd])
            spannedWidth := 0
            for c := cell.Col; c < cell.Col + cell.ColSpan; c++ {
                spannedWidth += colWidths[c]
            }
            // 加上中间边框
            spannedWidth += (cell.ColSpan - 1) * 1

            // 存储尺寸供 GetCellSize() 使用
            g.cellSpanSizes[cellKey] = [2]int{spannedWidth, spannedHeight}
        }
    }

    // ... 返回总尺寸 ...
}

func (g *GridLayout) GetCellSize(row, col int) [2]int {
    // 检查是否有跨单元格从 (row, col) 开始
    for _, cell := range g.style.Cells {
        if cell.Row == row && cell.Col == col {
            // 返回跨单元格的尺寸
            width := Σ(colWidths[col:col+cell.ColSpan])
            height := Σ(rowHeights[row:row+cell.RowSpan])
            // 加上边框
            return [2]int{width + (cell.ColSpan-1), height + (cell.RowSpan-1)}
        }
    }

    // 单个 cell
    return [2]int{colWidths[col], rowHeights[row]}
}
```

#### Instance 应用边界

```go
func (inst *Instance) SetBounds(x, y, w, h int) {
    inst.bounds = [4]int{x, y, w, h}
    contentX := x + inst.padding[3] + 1  // +1 跳过左边框
    contentY := y + inst.padding[0] + 1  // +1 跳过上边框

    for i, cell := range inst.cells {
        // 获取相对位置（考虑跨列）
        relX, relY := inst.getCellPosition(cell.Row, cell.Col)
        cellW, cellH := inst.getCellSize(cell.Row, cell.Col)

        // 应用到子节点
        inst.childBounds[i] = [4]int{
            contentX + relX,
            contentY + relY,
            cellW,
            cellH,
        }
    }
}
```

### 5.4 Cell 边界情况处理

#### 情况 1: 空 Grid（无 cells）

```go
if len(inst.cells) == 0 {
    // Grid 只有边框，没有内容
    // 尺寸 = padding + 边框
    return Size{
        Width:  paddingLR + cellBorderWidth,
        Height: paddingTB + cellBorderHeight,
    }
}
```

#### 情况 2: 单列单行

```go
// 1 列 1 行: 边框是 2x2 的正方形线
//
// ┌─────┬─────┐ = colWidths[0] + 1 + colWidths[1] + 1
// │ ... │ ... │   ↑ 左边框  ↑ 右边框
// ├─────┼─────┤   中间水平线
// │ ... │ ... │
// └─────┴─────┘
```

#### 情况 3: 边框宽度大于内容宽度

```go
// 如果边框占用超过可用空间，需要优雅降级
if cellBorderWidth > availableWidth {
    // 方案 1: 隐藏边框
    showCellBorders = false

    // 方案 2: 压缩内容
    availableWidth = 0

    // 方案 3: 返回最小尺寸
    return Size{
        Width:  cellBorderWidth + paddingLR,
        Height: cellBorderHeight + paddingTB,
    }
}
```

---

## 5. 坐标系统规范

### 5.1 坐标定义

```
contentX = bounds.x + padding[3] (left)
contentY = bounds.y + padding[0] (top)

对于 cell(row, col)：
- 内部逻辑坐标：runtime/layout/Grid 返回相对位置
- 绝对坐标：contentX + xOffset, contentY + yOffset
```

### 6.1 坐标定义

```
假设：3列 × 2行，colWidths=[10,10,10], rowHeights=[5,5]

垂直线条位置：
  Line 0: contentX
  Line 1: contentX + 10 + 1 (第0列内容宽 + 右边框宽1)
  Line 2: contentX + 10 + 1 + 10 + 1
  Line 3: contentX + 10 + 1 + 10 + 1 + 10 + 1

水平线条位置：
  Line 0: contentY
  Line 1: contentY + 5 + 1
  Line 2: contentY + 5 + 1 + 5 + 1

Cell(0,0) 内容区域：
  X: contentX + 1 到 contentX + 10
  Y: contentY + 1 到 contentY + 5
```

### 6.2 边框占位规则

```
边框宽度：每条线占 1 字符
边框高度：每条线占 1 字符

总宽度计算（有 cellBorders）：
  TotalWidth = padding.left
             + colWidths[0] + 1 (右边框)
             + colWidths[1] + 1
             + ...
             + colWidths[n-1] + 1
             + 1 (最右边框)
             + padding.right
             + Σ(columnGaps)

总高度计算（有 cellBorders）：
  TotalHeight = padding.top
              + (numRows + 1) * 1 (水平边框线)
              + Σ(rowHeights)
              + Σ(rowGaps)
              + padding.bottom
```

---

## 7. 边框实施方案

### 7.1 两类边框

#### 7.1.1 容器边框（Outer Border）
```
用途：对整个 Grid 组件绘制外边框
API：
  - VNode.SetBorderStyle("single")
  - VNode.SetLabel("Grid Title")

实现位置：
  - 由 reconciler 的 syncBorderProperties() 同步
  - 最终由 Fiber.BorderStyle 字段持有
  - 渲染层通过 FiberToNodeAdapter.GetBorder() 读取
  - 在 container 布局时绘制

坐标：
  覆盖整个 Grid 的 bounds 区域
```

#### 7.1.2 格子边框（Cell Borders）
```
用途：在格子之间绘制分隔线
API：
  - VNode.ShowCellBorders()
  - VNode.SetCellBorderStyle("single/double/light")
  - VNode.SetCellBorderRounded(boolean)
  - VNode.SetCellBorderColor("cyan")

实现位置：
  - stored in Instance.showCellBorders
  - 绘制逻辑在 Instance.Paint() → GenCellBorderDrawCmds()

坐标：
  基于计算的 colWidths, rowHeights
  通过 runtime/layout/Grid.GetPosition() 获取参考点
```

### 7.2 边框属性同步路径

```
VNode showCellBorders
         ↓
VNode.ToProps()
         ↓
Props["showCellBorders"]
         ↓
Instance.NewInstance(props)
         ↓
Instance.showCellBorders
         ↓
Instance.GetGridStyle() → GridStyle.ShowCellBorders
         ↓
runtime/layout/Grid.Measure()
         ↓
Instance.Paint() 检查 showCellBorders
```

### 7.3 边框字符集

```go
cellBorderChars = {
  "single": {
    Top:         "─", Bottom:      "─",
    Left:        "│", Right:       "│",
    TopLeft:     "┌", TopRight:    "┐",
    BottomLeft:  "└", BottomRight: "┘",
    Cross:       "┼",
    TopCross:    "┬", BottomCross: "┴",
    LeftCross:   "├", RightCross:  "┤"
  },
  "double": {
    Top:         "═", Bottom:      "═",
    Left:        "║", Right:       "║",
    TopLeft:     "╔", TopRight:    "╗",
    BottomLeft:  "╚", BottomRight: "╝",
    Cross:       "╬",
    TopCross:    "╦", BottomCross: "╩",
    LeftCross:   "╠", RightCross:  "╣"
  },
  "light": {
    Top:         "─", Bottom:      "─",
    Left:        "│", Right:       "│",
    TopLeft:     "┌", TopRight:    "┐",
    BottomLeft:  "└", BottomRight: "┘",
    Cross:       "┼",
    TopCross:    "┬", BottomCross: "┴",
    LeftCross:   "├", RightCross:  "┤"
  }
}
```

---

## 8. 节点 Checklist

### 8.1 VNode 层 checkItems

- [ ] **维度类型定义**
  - [ ] Fixed(int) - 固定尺寸
  - [ ] Flex{Factor int} - 弹性尺寸
  - [ ] Auto{} - 内容自适应
  - [ ] Min{Min int, Content Dimension} - 最小尺寸
  - [ ] Max{Max int, Content Dimension} - 最大尺寸

- [ ] **Cell 结构**
  - [ ] Child rtui.VNode
  - [ ] Row, Col 位置（0-based）
  - [ ] RowSpan, ColSpan 跨度

- [ ] **VNode 字段**
  - [ ] columns []Dimension
  - [ ] rows []Dimension
  - [ ] cells []Cell
  - [ ] columnGap, rowGap int
  - [ ] padding [4]int
  - [ ] alignContent rtui.Align
  - [ ] width, height, flex int
  - [ ] borderStyle, borderLabel (容器边框)
  - [ ] showCellBorders, cellBorderStyle, cellBorderRounded, cellBorderColor (格子边框)
  - [ ] style style.Style

- [ ] **构造函数**
  - [ ] New() → *VNode (默认初始化)
  - [ ] SetColumns(...Dimension) *VNode
  - [ ] SetRows(...Dimension) *VNode
  - [ ] AddCell(child, row, col) *VNode
  - [ ] AddCellWithSpan(child, row, col, rowSpan, colSpan) *VNode
  - [ ] SetGap(col, row int) *VNode
  - [ ] SetPadding(top, right, bottom, left int) *VNode

- [ ] **容器边框 API**
  - [ ] SetBorderStyle(style string) *VNode
  - [ ] SetBorderLabel(label string) *VNode
  - [ ] SingleBorder()
  - [ ] DoubleBorder()
  - [ ] RoundedBorder()
  - [ ] DashedBorder()

- [ ] **Cell 边框 API**
  - [ ] SetShowCellBorders(show bool) *VNode
  - [ ] SetCellBorderStyle(style string) *VNode
  - [ ] SetCellBorderRounded(rounded bool) *VNode
  - [ ] SetCellBorderColor(color string) *VNode
  - [ ] ShowCellBorders()
  - [ ] HideCellBorders()
  - [ ] SingleCellBorders()
  - [ ] DoubleCellBorders()
  - [ ] LightCellBorders()
  - [ ] RoundedCellBorders()

- [ ] **VNode 接口实现**
  - [ ] Key() string
  - [ ] SetKey(key string) rtui.VNode
  - [ ] Type() VNodeType (通过嵌入 *ElementVNode 返回 VNodeElement)
  - [ ] CreateInstance() rtui.ComponentInstance
  - [ ] ToProps() rtui.Props (生成 props 供 reconciler 使用)

---

### 8.2 Instance 层 checkItems

- [ ] **字段定义**
  - [ ] 关键标识：key
  - [ ] 布局配置：columns, rows, cells
  - [ ] 间距：columnGap, rowGap, padding
  - [ ] 对齐：alignContent
  - [ ] 尺寸：width, height, flex
  - [ ] 边框：borderStyle, borderLabel (容器)
  - [ ] cellBorders: showCellBorders, cellBorderStyle, cellBorderRounded, cellBorderColor
  - [ ] 样式：instStyle
  - [ **运行时状态**：bounds, colWidths, rowHeights, childBounds, dirty]

- [ ] **构造函数**
  - [ ] NewInstance(props rtui.Props) *Instance
    - [ ] 解析所有 props 字段
    - [ ] 设置默认值
    - [ ] 初始化 dirty=true

- [ ] **ComponentInstance 接口实现**
  - [ ] Key() string
  - [ ] GetBounds() [4]int
  - [ ] GetStyle() style.Style
  - [ ] SetStyle(s style.Style)
  - [ ] ClearDirty()
  - [ ] GetChildren(parent rtui.ComponentInstance) []rtui.ComponentInstance
  - [ ] SetProps(props rtui.Props)
  - [ ] SetChildBounds(id interface{}, bounds [4]int)
  - [ ] GetChildBounds(id interface{}) [4]int

- [ ] **测量接口**
  - [ ] Measure(constraints layout.Constraints) layout.Size
    - [ ] 创建 runtime/layout.Grid 实例
    - [ ] 调用 gridLayout.Measure(constraints)
    - [ ] 获取 colWidths = gridLayout.GetColumnWidths()
    - [ ] 获取 rowHeights = gridLayout.GetRowHeights()
    - [ ] 返回 size

- [ ] **PaintableInstance 接口实现**
  - [ ] Paint(x, y int) []paint.DrawCmd
    - [ ] 获取当前 bounds
    - [ ] 检查 showCellBorders
    - [ ] 调用 GenCellBorderDrawCmds(x, y)
    - [ ] 返回绘制命令列表

- [ ] **GridStyleProvider 接口**
  - [ ] GetGridStyle() *layout.GridStyle
    - [ ] 构造包含所有布局参数的 GridStyle
    - [ ] 包含 cellBorders 配置

- [ ] **边框绘制**
  - [ ] GenCellBorderDrawCmds(x, y int) []paint.DrawCmd
    - [ ] 获取 cellBorderChars 映射
    - [ ] 计算每个格子区域的 bounding box
    - [ ] 绘制水平和垂直线
    - [ ] 绘制交叉点字符
    - [ ] 处理 roundedCorner 情况
    - [ ] 应用边框颜色

---

### 8.3 runtime/layout/Grid checkItems

- [ ] **类型定义**
  - [ ] GridDimension interface
  - [ ] GridFixed, GridFlex, GridAuto, GridMin, GridMax
  - [ ] GridCell struct
  - [ ] GridStyle struct
    - [ ] Columns, Rows, Cells
    - [ ] ColumnGap, RowGap
    - [ ] Padding
    - [ ] Width, Height
    - [ ] ShowCellBorders, CellBorderWidth, CellBorderHeight
  - [ ] GridStyleProvider interface

- [ ] **GridLayout struct**
  - [ ] id string
  - [ ] style *GridStyle
  - [ ] children []Node (实现 layout.Node 的子节点)
  - [ ] colWidths []int (计算结果)
  - [ ] rowHeights []int (计算结果)

- [ ] **核心方法**
  - [ ] NewGridLayout(id string, style *GridStyle) *GridLayout
  - [ ] Measure(constraints Constraints) Size
    - [ ] 计算列宽
    - [ ] 计算行高
    - [ ] 处理 Flex 分配
    - [ ] 考虑 gap
    - [ ] 考虑 padding
    - [ ] **考虑 cellBorders 的高度**：borderHeight = (numRows + 1) * 1
    - [ ] 返回总尺寸
  - [ ] GetColumnWidths() []int
  - [ ] GetRowHeights() []int
  - [ ] GetCellPosition(row, col int) [2]int (返回相对位置)
  - [ ] GetCellSize(row, col int) [2]int

- [ ] **坐标计算**
  - [ ] getCellPosition(row, col int) [2]int
    - [ ] X = padding.left + Σ(colWidths[0:col]) + col * gap
    - [ ] **加上 cellBorders 的 +1 偏移**：x += 1 (跳过左边框线)
    - [ ] Y = padding.top + Σ(rowHeights[0:row]) + row * gap
    - [ ] **加上 cellBorders 的 +1 偏移**：y += 1 (跳过上边框线)
  - [ ] getCellSize(row, col int) [2]int
    - [ ] 考虑 RowSpan, ColSpan
    - [ ] 对跨行列的 cell 加上边框高度

---

### 8.4 边框属性同步 checkItems

#### VNode → Props

- [ ] VNode.ToProps() 包含所有边框属性
  - [ ] "showCellBorders": bool
  - [ ] "cellBorderStyle": string
  - [ ] "cellBorderRounded": bool
  - [ ] "cellBorderColor": string

#### Props → Instance

- [ ] Instance.NewInstance(props) 解析处理
  - [ ] showCellBorders = getBoolProp(props, "showCellBorders", false)
  - [ ] cellBorderStyle = getStringProp(props, "cellBorderStyle", "single")
  - [ ] cellBorderRounded = getBoolProp(props, "cellBorderRounded", false)
  - [ ] cellBorderColor = getStringProp(props, "cellBorderColor", "")

#### Instance → Layout Engine

- [ ] Instance.GetGridStyle() 返回完整 GridStyle
  - [ ] style.ShowCellBorders = inst.showCellBorders
  - [ ] style.CellBorderWidth = 1 (固定)
  - [ ] style.CellBorderHeight = 1 (固定)

---

### 8.5 Reconciler checkItems

- [ ] **VNode 创建/更新**
  - [ ] VNode.Type() 返回正确的类型
    - Grid 嵌入 *ElementVNode，返回 VNodeElement
  - [ ] 调用 VNode.CreateInstance()
  - [ ] Props 通过 VNode.ToProps() 传递

- [ ] **属性同步**
  - [ ] Instance.SetProps(props)
  - [ ] 对于 Element 类型的 VNode，调用 syncBorderProperties()
    - [ ] 将 Props["borderStyle"] 同步到 Fiber.BorderStyle

- [ ] **Fiber 类型**
  - [ ] Fiber.FontStyle 字段
  - [ ] Fiber.BorderStyle 字段
    - [ ] 用于容器边框渲染
  - [ ] Fiber.Props map
    - [ ] 备用路径存储 borderStyle 属性

---

### 8.6 渲染层 checkItems

- [ ] **FiberToNodeAdapter**
  - [ ] GetBorder() 方法
    - [ ] 优先读取 Fiber.BorderStyle
    - [ ] Fallback 到 Props["borderStyle"]

- [ ] **Paint Engine**
  - [ ] 处理 DrawCmd
  - [ ] 绘制边框字符
  - [ ] 处理重叠和 Z-order

---

### 8.7 单元测试 checkItems

- [ ] **VNode 层**
  - [ ] TestVNode_CellBordersDefaults
  - [ ] TestVNode_SetCellBorderProperties
  - [ ] TestVNode_ShowCellBorders / HideCellBorders
  - [ ] TestVNode_SingleCellBorders / DoubleCellBorders / LightCellBorders
  - [ ] TestVNode_RoundedCellBorders
  - [ ] TestVNode_CellBordersChaining

- [ ] **Instance 层**
  - [ ] TestInstance_CellBordersProps
  - [ ] TestInstance_CellBordersMeasureSize
  - [ ] TestInstance_GetGridStyle

- [ ] **边框绘制**
  - [ ] TestCellBorderChars
  - [ ] Test_GridCellBorders_Verify (验证 DrawCmd 正确性)
  - [ ] Test_GridCellBorders_NoBorders
  - [ ] Test_GridCellBorders_BorderStyles
  - [ ] Test_GridCellBorders_PaintInterface

- [ ] **集成测试**
  - [ ] 实际渲染测试
  - [ ] 复杂布局测试
  - [ ] 边框 + 容器边框混合测试

---

## 9. 关键设计原则

### 9.1 职责单一

- **VNode**: 纯声明，无状态，无逻辑
- **Instance**: 持有运行时状态，管理生命周期
- **runtime/layout/Grid**: 纯布局计算，不涉及渲染
- **cell_borders_*.go**: 纯边框绘制逻辑

### 6.2 委托而非重复

- Instance.Measure() 委托给 runtime/layout/Grid
- 不在 Instance 中重复布局计算逻辑
- runtime/layout/Grid 是可复用的布局引擎

### 6.3 坐标系统一致性

- 明确定义 contentX, contentY 基准点
- 边框位置、内容位置统一基于同一坐标系
- runtime/layout.Grid 返回相对坐标
- Instance.Paint() 转换为绝对坐标

### 6.4 边框占位明确

- 每条边框线占 1 字符
- 总高度/宽度计算中明确包含边框占位
- 内容与边框之间无额外 gap

### 6.5 Props 同步路径清晰

- VNode → Props → Instance → GridStyle
- 实例字段 -> GridStyle 的映射明确
- 边框属性不丢失、不混淆

---

## 7. 当前架构状态

### ✅ 已实现

1. **VNode 层**
   - 完整的维度类型
   - Cell Borders API
   - 容器边框 API
   - ToProps() 转换

2. **Instance 层**
   - 字段完整
   - Measure() 委托给 runtime/layout/Grid
   - GetGridStyle() 实现
   - Paint() 预留 cellBorders 绘制入口

3. **runtime/layout/Grid**
   - 布局计算引擎
   - colWidths/rowHeights 计算
   - CellBorderHeight 处理

4. **边框绘制**
   - cellBorderChars 映射
   - GenCellBorderDrawCmds 方法

### ⚠️ 需要验证/改进

1. **坐标系统**
   - [ ] 验证 getCellPosition 的 +1 偏移是否正确
   - [ ] 验证边框线位置与内容位置的一致性
   - [ ] 编写端到端测试验证渲染结果

2. **边界情况**
   - [ ] 空 Grid（无 cells）
   - [ ] 跨行列 cell 的边框绘制
   - [ ] 单列/单行 Grid
   - [ ] 过小的 Grid（不足以显示边框）

3. **性能**
   - [ ] 避免 Measure() 阶段的重复计算
   - [ ] 缓存 colWidths/rowHeights
   - [ ] dirty 标记的正确使用

---

## 8. 调试指南

### 8.1 Debug 标记

在关键位置添加 Debug 输出：

```go
// Instance.Measure()
fmt.Printf("[GRID MEASURE] numRows=%d, colWidths=%v, rowHeights=%v\n",
  len(inst.rowHeights), inst.colWidths, inst.rowHeights)

// Instance.Paint()
fmt.Printf("[GRID PAINT] bounds=%v, showCellBorders=%v\n",
  inst.bounds, inst.showCellBorders)
fmt.Printf("[GRID PAINT] Generated %d draw commands\n", len(cmds))
```

### 8.2 验证流程

1. **运行 demo**
   ```bash
   go run examples/fiber_firsts/border_as_property/simple_grid_demo.go
   ```

2. **检查输出**
   - 边框字符是否正确显示
   - 内容是否与边框对齐
   - 高度/宽度计算是否正确

3. **对比不同边框样式**
   - single vs double vs light
   - 圆角 vs 直角

---

## 9. 扩展方向

### 9.1 高级边框样式

- 虚线边框
- 彩色边框（每条线不同颜色）
- 自定义边框字符集

### 9.2 交互功能

- 鼠标高亮 cell
- 选中边框样式
- 可折叠行列

### 9.3 性能优化

- 虚拟滚动（仅渲染可见区域）
- 增量渲染（只重绘变化的 cell）

---

## 10. 总结

Grid 组件的架构是清晰且可扩展的：

1. **分层清晰**: VNode → Instance → Layout Engine → Paint
2. **单一职责**: 每层负责特定任务
3. **委托设计**: 布局计算委托给 runtime/layout/Grid
4. **坐标统一**: 基于 contentX/contentY 的统一坐标系
5. **边框明确**: 容器边框与 cellBorders 分离，各自独立

这个架构虽然复杂（因为网格布局本身的复杂性），但不是乱的。每模块职责边界清晰，数据流向明确，易于维护和扩展。
