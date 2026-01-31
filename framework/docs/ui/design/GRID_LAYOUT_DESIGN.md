# Grid 布局设计文档

**版本**: v1.0
**日期**: 2026-01-31
**来源**: demo4_layout.md, demo5_ide.md
**状态**: 🟡 中优先级

---

## 一、概述

### 1.1 设计目标

实现二维网格布局系统，支持 Dashboard、IDE 等复杂界面场景。

### 1.2 使用场景

```
┌─────────┬─────────┬─────────┐
│ CPU 32% │ RAM 68% │ Net 120 │
├─────────┴─────────┴─────────┤
│ Logs (auto scroll)            │
└───────────────────────────────┘
```

### 1.3 与 Flex 的区别

| 特性 | Flex | Grid |
|------|------|------|
| 维度 | 一维（行或列） | 二维（行和列） |
| 适用场景 | 线性布局 | 区域型布局 |
| 对齐 | 主轴/交叉轴 | 行/列独立控制 |
| 跨度 | 无 | 跨行/跨列 |

---

## 二、Grid 类型定义

### 2.1 Dimension 类型

```go
// framework/layout/dimension.go

package layout

// Dimension 尺寸类型
type Dimension interface {
    isDimension()
}

// Fixed 固定尺寸
type Fixed int

func (f Fixed) isDimension() {}

// Flex 弹性尺寸
type Flex struct {
    Grow int // 比例因子
}

func (f Flex) isDimension() {}

// Auto 自动尺寸（由内容决定）
type Auto struct{}

func (a Auto) isDimension() ()

// Min 最小尺寸约束
type Min struct {
    Min  int
    Rest Dimension
}

// Max 最大尺寸约束
type Max struct {
    Max  int
    Rest Dimension
}
```

### 2.2 Grid Props

```go
// framework/layout/grid.go

package layout

// GridProps Grid 布局属性
type GridProps struct {
    // 行定义
    RowSizes []Dimension
    // 列定义
    ColSizes []Dimension
    // 行间距
    RowGap int
    // 列间距
    ColGap int
    // 对齐方式
    AlignMain    AlignType // 主轴对齐
    AlignCross   AlignType // 交叉轴对齐
}

// CellProps 单元格属性
type CellProps struct {
    Row     int // 起始行（从 0 开始）
    Col     int // 起始列
    RowSpan int // 跨行数（默认 1）
    ColSpan int // 跨列数（默认 1）
}
```

---

## 三、Grid API 设计

### 3.1 声明式 API

```go
// framework/layout/grid.go

// Grid 创建网格布局
func Grid(props GridProps, children ...VNode) VNode {
    return &VNodeElement{
        Type:     "Grid",
        Props:    props,
        Children: children,
    }
}

// Cell 创建网格单元
func Cell(props CellProps, child VNode) VNode {
    return &VNodeElement{
        Type:     "Cell",
        Props:    props,
        Children: []VNode{child},
    }
}
```

### 3.2 便捷函数

```go
// ui/grid.go

package ui

// 预定义尺寸
func Fixed(n int) layout.Fixed     { return layout.Fixed(n) }
func Flex(n int) layout.Flex       { return layout.Flex{Grow: n} }
func Auto() layout.Auto            { return layout.Auto{} }
func Min(n int, rest layout.Dimension) layout.Min {
    return layout.Min{Min: n, Rest: rest}
}
func Max(n int, rest layout.Dimension) layout.Max {
    return layout.Max{Max: n, Rest: rest}
}

// Grid 快捷函数
func UIGrid(rowSizes, colSizes []Dimension, children ...VNode) VNode {
    return layout.Grid(layout.GridProps{
        RowSizes: rowSizes,
        ColSizes: colSizes,
    }, children...)
}

// Cell 快捷函数
func UICell(row, col int, child VNode) VNode {
    return layout.Cell(layout.CellProps{
        Row: row,
        Col: col,
    }, child)
}

// 跨列单元格
func UICellSpan(row, col, rowSpan, colSpan int, child VNode) VNode {
    return layout.Cell(layout.CellProps{
        Row:     row,
        Col:     col,
        RowSpan: rowSpan,
        ColSpan: colSpan,
    }, child)
}
```

---

## 四、布局算法

### 4.1 约束传递

```go
// framework/layout/grid_algorithm.go

package layout

// GridLayout 网格布局计算
type GridLayout struct {
    Props    GridProps
    Children []VNode
    Constraints
}

// Measure 测量阶段
func (g *GridLayout) Measure(constraint Constraints) Size {
    // 1. 计算列宽
    colWidths := g.computeColumnWidths(constraint.MaxW)

    // 2. 计算行高
    rowHeights := g.computeRowHeights(constraint.MaxH, colWidths)

    // 3. 计算总尺寸
    totalW := sum(colWidths) + (len(colWidths)-1)*g.Props.ColGap
    totalH := sum(rowHeights) + (len(rowHeights)-1)*g.Props.RowGap

    return Size{W: totalW, H: totalH}
}

// computeColumnWidths 计算列宽
func (g *GridLayout) computeColumnWidths(availableW int) []int {
    n := len(g.Props.ColSizes)
    widths := make([]int, n)

    // 1. 计算固定尺寸
    fixedTotal := 0
    flexIndices := []int{}

    for i, dim := range g.Props.ColSizes {
        switch d := dim.(type) {
        case Fixed:
            widths[i] = int(d)
            fixedTotal += int(d)
        case Flex:
            flexIndices = append(flexIndices, i)
        case Auto:
            // 测量该列所有子元素
            widths[i] = g.measureColAuto(i)
            fixedTotal += widths[i]
        }
    }

    // 2. 分配剩余空间给 Flex
    remaining := max(0, availableW - fixedTotal)
    if len(flexIndices) > 0 {
        totalFlexGrow := 0
        for _, i := range flexIndices {
            if fg, ok := g.Props.ColSizes[i].(Flex); ok {
                totalFlexGrow += fg.Grow
            }
        }

        perFlex := remaining / totalFlexGrow
        for _, i := range flexIndices {
            widths[i] = perFlex * g.Props.ColSizes[i].(Flex).Grow
        }
    }

    return widths
}

// computeRowHeights 计算行高
func (g *GridLayout) computeRowHeights(availableH int, colWidths []int) []int {
    // 类似 computeColumnWidths
    // ...
}

// Layout 布局阶段
func (g *GridLayout) Layout(x, y int, colWidths, rowHeights []int) {
    for _, child := range g.Children {
        cell := child.Props.(CellProps)

        // 计算单元格位置和尺寸
        cellX := x
        for c := 0; c < cell.Col; c++ {
            cellX += colWidths[c] + g.Props.ColGap
        }

        cellY := y
        for r := 0; r < cell.Row; r++ {
            cellY += rowHeights[r] + g.Props.RowGap
        }

        cellW := 0
        for c := cell.Col; c < cell.Col+cell.ColSpan; c++ {
            cellW += colWidths[c]
            if c < cell.Col+cell.ColSpan-1 {
                cellW += g.Props.ColGap
            }
        }

        cellH := 0
        for r := cell.Row; r < cell.Row+cell.RowSpan; r++ {
            cellH += rowHeights[r]
            if r < cell.Row+cell.RowSpan-1 {
                cellH += g.Props.RowGap
            }
        }

        // 递归布局子元素
        child.Layout(cellX, cellY, cellW, cellH)
    }
}
```

---

## 五、使用示例

### 5.1 Dashboard 示例

```go
// 示例：仪表盘布局
func Dashboard() VNode {
    return ui.Grid(
        []ui.Dimension{
            ui.Fixed(10), // CPU
            ui.Fixed(10), // Memory
            ui.Fixed(10), // Network
        },
        []ui.Dimension{
            ui.Fixed(5),
            ui.Fixed(5),
            ui.Flex(1), // Logs
        },
        ui.UICell(0, 0, CpuPanel()),
        ui.UICell(0, 1, MemoryPanel()),
        ui.UICell(0, 2, NetworkPanel()),
        ui.UICellSpan(1, 0, 1, 3, LogsPanel()), // 跨 3 列
    )
}

func CpuPanel() VNode {
    return ui.Box().
        Border(true).
        Padding(1).
        Child(ui.Text("CPU: 32%"))
}

func LogsPanel() VNode {
    return ui.Box().
        Flex(1).
        Border(true).
        Padding(1).
        Child(ui.VirtualList(...))
}
```

### 5.2 IDE 主布局

```go
// 示例：IDE 主布局
func IDELayout() VNode {
    return ui.Grid(
        []ui.Dimension{
            ui.Fixed(3),   // Header
            ui.Flex(1),    // Main
            ui.Fixed(1),   // StatusBar
        },
        []ui.Dimension{
            ui.Flex(1),    // 全宽
        },
        ui.UICell(0, 0, Header()),
        ui.UICell(1, 0, MainArea()),
        ui.UICell(2, 0, StatusBar()),
    )
}

func MainArea() VNode {
    return ui.Grid(
        []ui.Dimension{
            ui.Flex(1),    // 完整高度
        },
        []ui.Dimension{
            ui.Fixed(24),  // Sidebar
            ui.Flex(1),    // Content
        },
        ui.UICell(0, 0, Sidebar()),
        ui.UICell(0, 1, ContentArea()),
    )
}
```

### 5.3 自适应布局

```go
// 示例：最小尺寸约束
func ResponsiveGrid() VNode {
    return ui.Grid(
        []ui.Dimension{
            ui.Min(3, ui.Flex(1)), // 最小 3，其余弹性
            ui.Max(10, ui.Flex(2)), // 最大 10，其余弹性
        },
        []ui.Dimension{
            ui.Fixed(20),
            ui.Flex(1),
        },
        // ...
    )
}
```

---

## 六、性能优化

### 6.1 缓存策略

```go
// framework/layout/grid_cache.go

type GridLayoutCache struct {
    // 缓存计算结果
    lastConstraints Constraints
    lastColWidths   []int
    lastRowHeights  []int
    dirty           bool
}

func (c *GridLayoutCache) Invalidate() {
    c.dirty = true
}
```

### 6.2 增量更新

只有以下情况需要重新计算：

1. 行/列定义变化
2. 约束变化
3. 子元素尺寸变化
4. Gap 变化

---

## 七、实施计划

### 阶段 1: 基础实现

- [ ] 实现 Dimension 类型
- [ ] 实现 GridProps/CellProps
- [ ] 实现基础 API

### 阶段 2: 布局算法

- [ ] 实现列宽计算
- [ ] 实现行高计算
- [ ] 实现跨度支持
- [ ] 实现 Gap 支持

### 阶段 3: 集成测试

- [ ] 编写单元测试
- [ ] 创建 Dashboard 示例
- [ ] 创建 IDE 示例
- [ ] 性能基准测试

---

## 八、测试策略

```go
// framework/layout/grid_test.go

func TestGridFixed(t *testing.T) {
    grid := ui.Grid(
        []ui.Dimension{ui.Fixed(10), ui.Fixed(20)},
        []ui.Dimension{ui.Fixed(5), ui.Fixed(15)},
        ui.UICell(0, 0, ui.Box()),
        ui.UICell(0, 1, ui.Box()),
        ui.UICell(1, 0, ui.Box()),
        ui.UICell(1, 1, ui.Box()),
    )

    // 验证布局结果
    assert.Equal(t, 4, len(grid.Children))
}

func TestGridFlex(t *testing.T) {
    grid := ui.Grid(
        []ui.Dimension{ui.Flex(1), ui.Flex(2)}, // 1:2 比例
        []ui.Dimension{ui.Fixed(50)},
        // ...
    )

    // 验证弹性分配
    // 在 150px 宽度中，应该是 50px 和 100px
}

func TestGridSpan(t *testing.T) {
    // 测试跨行跨列
}
```

---

## 九、边界情况处理

### 9.1 空单元格

```go
// 可以留空，Grid 会自动处理
ui.Grid(
    rowSizes, colSizes,
    ui.UICell(0, 0, content1),
    // (0,1) 空白
    ui.UICell(1, 1, content2),
)
```

### 9.2 越界处理

```go
// Row/Col 超出定义范围时的处理
func (g *GridLayout) validateCell(cell CellProps) error {
    if cell.Row >= len(g.Props.RowSizes) {
        return fmt.Errorf("row %d out of bounds", cell.Row)
    }
    if cell.Col >= len(g.Props.ColSizes) {
        return fmt.Errorf("col %d out of bounds", cell.Col)
    }
    // ...
}
```

### 9.3 重叠单元格

```go
// 检测并处理重叠
func (g *GridLayout) checkOverlap() error {
    occupied := make(map[string]bool)
    for _, child := range g.Children {
        cell := child.Props.(CellProps)
        for r := cell.Row; r < cell.Row+cell.RowSpan; r++ {
            for c := cell.Col; c < cell.Col+cell.ColSpan; c++ {
                key := fmt.Sprintf("%d,%d", r, c)
                if occupied[key] {
                    return fmt.Errorf("overlap at %s", key)
                }
                occupied[key] = true
            }
        }
    }
    return nil
}
```

---

## 十、与其他布局的组合

### 10.1 Grid + Flex

```go
// Grid 单元格内使用 Flex
ui.UICell(0, 0,
    ui.VStack(  // Flex 列布局
        ui.Text("Title"),
        ui.Text("Content"),
    ),
)
```

### 10.2 Grid + Scroll

```go
// Grid 单元格内使用 Scroll
ui.UICell(1, 0,
    ui.ScrollY(
        ui.VirtualList(...),
    ),
)
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-31
