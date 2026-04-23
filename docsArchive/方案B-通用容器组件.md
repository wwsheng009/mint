# 方案 B：通用 Container 组件 - 实施计划

## 概述

引入一个超级灵活的通用容器组件，统一所有布局模式，通过配置支持 Flex、Grid、Wrap、Absolute 等布局方式。

## 核心理念

一个容器组件，通过配置选择布局模式，减少组件数量，降低学习成本。

## 目标

1. 统一所有布局容器为一个 Container 组件
2. 通过配置参数灵活切换布局模式
3. 极简化组件层次
4. 提供极致的开发体验

---

## 设计概览

### Container 组件核心 API

```go
package container

// LayoutMode 定义布局模式
type LayoutMode int

const (
    LayoutFlex    LayoutMode = iota  // 弹性布局
    LayoutGrid                      // 网格布局
    LayoutWrap                      // 换行布局
    LayoutAbsolute                  // 绝对定位
    LayoutNone                      // 无特殊布局（垂直堆叠）
)

// VNode 通用容器
type VNode struct {
    *rtui.ElementVNode

    // === 布局配置 ===
    layoutMode   LayoutMode

    // === Flex 属性（LayoutFlex 模式）===
    direction    rtui.Direction
    align        rtui.Align
    crossAlign   rtui.Align
    gap          int

    // === Grid 属性（LayoutGrid 模式）===
    columns      []Dimension
    rows         []Dimension
    columnGap    int
    rowGap       int
    gridCells    []Cell

    // === Wrap 属性（LayoutWrap 模式）===
    wrapAlign    rtui.Align

    // === Absolute 属性（LayoutAbsolute 模式）===
    left         PositionValue
    top          PositionValue
    right        PositionValue
    bottom       PositionValue
    anchor       layout.Anchor

    // === 通用属性 ===
    padding      [4]int
    width        int
    height       int
    flex         int

    // === 边框属性（方案 A 集成）===
    borderStyle  layout.BorderStyle
    borderLabel  string

    // === 内容 ===
    children     []rtui.VNode
    style        style.Style
}
```

---

## 实施步骤

### Phase 1: Container 组件基础框架

#### 1.1 创建 container 包
```
ui/components/container/
├── vnode.go       # VNode 定义
├── dimension.go   # Grid/LocPosition 类型定义
├── instance.go    # Instance（可选）
└── container_test.go
```

#### 1.2 定义 LayoutMode 和相关枚举
```go
// ui/components/container/vnode.go

// LayoutMode 定义布局模式
type LayoutMode int

const (
    LayoutNone     LayoutMode = iota  // 垂直堆叠（默认）
    LayoutFlex                       // 弹性布局
    LayoutGrid                       // 网格布局
    LayoutWrap                       // 换行布局
    LayoutAbsolute                   // 绝对定位
)

// String 实现
func (m LayoutMode) String() string {
    switch m {
    case LayoutFlex:
        return "flex"
    case LayoutGrid:
        return "grid"
    case LayoutWrap:
        return "wrap"
    case LayoutAbsolute:
        return "absolute"
    case LayoutNone:
        return "none"
    default:
        return "unknown"
    }
}
```

#### 1.3 定义 Dimension 类型（Grid 用）
```go
// ui/components/container/dimension.go

// Dimension represents a column or row dimension in the grid.
type Dimension interface {
    isGridDimension()
}

// Fixed creates a fixed-size dimension.
type Fixed int

func (f Fixed) isGridDimension() {}

// Flex creates a flexible dimension that takes remaining space.
type Flex struct {
    Factor int // Flex factor, defaults to 1
}

func (f Flex) isGridDimension() {}

// Auto creates a dimension that sizes to content.
type Auto struct{}

func (a Auto) isGridDimension() {}

// Min creates a dimension with minimum size.
type Min struct {
    Min     int
    Content Dimension
}

func (m Min) isGridDimension() {}

// Max creates a dimension with maximum size.
type Max struct {
    Max     int
    Content Dimension
}

func (m Max) isGridDimension() {}

// PositionValue represents a position value (absolute or relative).
type PositionValue interface {
    isPositionValue()
    Resolve(containerSize int) int
}

// AbsolutePos is a fixed position in cells.
type AbsolutePos int

func (a AbsolutePos) isPositionValue() {}
func (a AbsolutePos) Resolve(_ int) int { return int(a) }

// RelativePos is a percentage (0-100).
type RelativePos int

func (r RelativePos) isPositionValue() {}
func (r RelativePos) Resolve(containerSize int) int {
    return containerSize * int(r) / 100
}

// Anchor is an alias for layout.Anchor
type Anchor = layout.Anchor
```

---

### Phase 2: Container VNode 实现

#### 2.1 VNode 结构定义
```go
// ui/components/container/vnode.go

type VNode struct {
    *rtui.ElementVNode

    // === 标识 ===
    key  string

    // === 布局模式 ===
    layoutMode LayoutMode

    // === Flex 属性 ===
    direction    rtui.Direction
    align        rtui.Align
    crossAlign   rtui.Align
    gap          int

    // === Grid 属性 ===
    columns      []Dimension
    rows         []Dimension
    columnGap    int
    rowGap       int
    gridCells    []Cell

    // === Wrap 属性 ===
    wrapAlign    rtui.Align

    // === Absolute 属性 ===
    left         PositionValue
    top          PositionValue
    right        PositionValue
    bottom       PositionValue
    anchor       Anchor

    // === 通用属性 ===
    padding      [4]int
    width        int
    height       int
    flex         int

    // === 边框属性 ===
    borderStyle  layout.BorderStyle
    borderLabel  string

    // === 内容 ===
    children     []rtui.VNode
    style        style.Style
}

type Cell struct {
    Child   rtui.VNode
    Row     int
    Col     int
    RowSpan int
    ColSpan int
}
```

#### 2.2 构造函数
```go
// New creates a new Container VNode with default layout.
func New() *VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("container"),
        layoutMode:   LayoutNone,  // 默认垂直堆叠
        padding:      [4]int{0, 0, 0, 0},
        anchor:       layout.AnchorTopLeft,
        children:     make([]rtui.VNode, 0),
    }
}

// NewFlex creates a Flex container.
func NewFlex() *VNode {
    return New().LayoutMode(LayoutFlex)
}

// NewGrid creates a Grid container.
func NewGrid() *VNode {
    return New().LayoutMode(LayoutGrid)
}

// NewWrap creates a Wrap container.
func NewWrap() *VNode {
    return New().LayoutMode(LayoutWrap)
}

// NewAbsolute creates an Absolute container.
func NewAbsolute() *VNode {
    return New().LayoutMode(LayoutAbsolute)
}
```

---

### Phase 3: 布局模式配置方法

#### 3.1 基础布局方法
```go
// LayoutMode sets the layout mode.
func (c *VNode) LayoutMode(mode LayoutMode) *VNode {
    c.layoutMode = mode
    return c
}

// Padding sets the padding.
func (c *VNode) Padding(top, right, bottom, left int) *VNode {
    c.padding = [4]int{top, right, bottom, left}
    return c
}

// PaddingAll sets same padding on all sides.
func (c *VNode) PaddingAll(v int) *VNode {
    return c.Padding(v, v, v, v)
}

// Size sets width and height.
func (c *VNode) Size(width, height int) *VNode {
    c.width = width
    c.height = height
    return c
}

// Flex sets the flex factor.
func (c *VNode) Flex(factor int) *VNode {
    c.flex = factor
    return c
}
```

#### 3.2 Flex 模式方法
```go
// Direction sets the flex direction.
func (c *VNode) Direction(dir rtui.Direction) *VNode {
    c.direction = dir
    return c
}

// Row sets horizontal layout.
func (c *VNode) Row() *VNode {
    return c.Direction(rtui.DirectionRow)
}

// Column sets vertical layout.
func (c *VNode) Column() *VNode {
    return c.Direction(rtui.DirectionColumn)
}

// Align sets the main axis alignment.
func (c *VNode) Align(align rtui.Align) *VNode {
    c.align = align
    return c
}

// CrossAlign sets the cross axis alignment.
func (c *VNode) CrossAlign(align rtui.Align) *VNode {
    c.crossAlign = align
    return c
}

// Gap sets the gap between children.
func (c *VNode) Gap(g int) *VNode {
    c.gap = g
    return c
}
```

#### 3.3 Grid 模式方法
```go
// Columns sets the grid columns.
func (c *VNode) Columns(dims ...Dimension) *VNode {
    c.columns = dims
    return c
}

// Rows sets the grid rows.
func (c *VNode) Rows(dims ...Dimension) *VNode {
    c.rows = dims
    return c
}

// ColumnGap sets the gap between columns.
func (c *VNode) ColumnGap(g int) *VNode {
    c.columnGap = g
    return c
}

// RowGap sets the gap between rows.
func (c *VNode) RowGap(g int) *VNode {
    c.rowGap = g
    return c
}

// AddCell adds a cell to the grid.
func (c *VNode) AddCell(child rtui.VNode, row, col int) *VNode {
    c.gridCells = append(c.gridCells, Cell{
        Child:   child,
        Row:     row,
        Col:     col,
        RowSpan: 1,
        ColSpan: 1,
    })
    return c
}

// AddCellSpan adds a cell with custom span.
func (c *VNode) AddCellSpan(child rtui.VNode, row, col, rowSpan, colSpan int) *VNode {
    c.gridCells = append(c.gridCells, Cell{
        Child:   child,
        Row:     row,
        Col:     col,
        RowSpan: rowSpan,
        ColSpan: colSpan,
    })
    return c
}
```

#### 3.4 Wrap 模式方法
```go
// WrapAlign sets the alignment in wrap layout.
func (c *VNode) WrapAlign(align rtui.Align) *VNode {
    c.wrapAlign = align
    return c
}
```

#### 3.5 Absolute 模式方法
```go
// Position sets left and top positions.
func (c *VNode) Position(left, top PositionValue) *VNode {
    c.left = left
    c.top = top
    return c
}

// SetLeft sets the left position.
func (c *VNode) SetLeft(pos PositionValue) *VNode {
    c.left = pos
    return c
}

// SetTop sets the top position.
func (c *VNode) SetTop(pos PositionValue) *VNode {
    c.top = pos
    return c
}

// SetRight sets the right position.
func (c *VNode) SetRight(pos PositionValue) *VNode {
    c.right = pos
    return c
}

// SetBottom sets the bottom position.
func (c *VNode) SetBottom(pos PositionValue) *VNode {
    c.bottom = pos
    return c
}

// SetAnchor sets the anchor point.
func (c *VNode) SetAnchor(anchor Anchor) *VNode {
    c.anchor = anchor
    return c
}
```

#### 3.6 边框属性方法
```go
// Border sets the border style and label.
func (c *VNode) Border(style layout.BorderStyle, label string) *VNode {
    c.borderStyle = style
    c.borderLabel = label
    return c
}

// Bordered sets border with specified style.
func (c *VNode) Bordered(style layout.BorderStyle) *VNode {
    return c.Border(style, "")
}

// SingleBorder sets single line border.
func (c *VNode) SingleBorder(label string) *VNode {
    return c.Border(layout.BorderSingle, label)
}

// DoubleBorder sets double line border.
func (c *VNode) DoubleBorder(label string) *VNode {
    return c.Border(layout.BorderDouble, label)
}
```

---

### Phase 4: 内容和属性管理

#### 4.1 子节点管理
```go
// Children returns all children.
func (c *VNode) Children() []rtui.VNode {
    return c.children
}

// SetChildren sets the children.
func (c *VNode) SetChildren(children []rtui.VNode) *VNode {
    c.children = children
    return c
}

// Add adds a child.
func (c *VNode) Add(child rtui.VNode) *VNode {
    c.children = append(c.children, child)
    return c
}

// AddMultiple adds multiple children.
func (c *VNode) AddMultiple(children ...rtui.VNode) *VNode {
    c.children = append(c.children, children...)
    return c
}
```

#### 4.2 VNode 接口实现
```go
func (c *VNode) Key() string {
    return c.key
}

func (c *VNode) SetKey(key string) rtui.VNode {
    c.key = key
    return c
}

func (c *VNode) Tag() string {
    return "container"
}

func (c *VNode) Style() style.Style {
    return c.style
}

func (c *VNode) SetStyle(st style.Style) rtui.VNode {
    c.style = st
    return c
}

func (c *VNode) GetLayer() rtui.Layer {
    return rtui.LayerBase
}

func (c *VNode) SetLayer(l rtui.Layer) rtui.VNode {
    return c
}
```

#### 4.3 Props 实现
```go
func (c *VNode) Props() rtui.Props {
    return rtui.Props{
        "key":          c.key,
        "layoutMode":   c.layoutMode,
        "direction":    c.direction,
        "align":        c.align,
        "crossAlign":   c.crossAlign,
        "gap":          c.gap,
        "padding":      c.padding,
        "width":        c.width,
        "height":       c.height,
        "flex":         c.flex,
        "borderStyle":  c.borderStyle,
        "borderLabel":  c.borderLabel,
        "wrapAlign":    c.wrapAlign,
        "left":         c.left,
        "top":          c.top,
        "right":        c.right,
        "bottom":       c.bottom,
        "anchor":       c.anchor,
        "style":        c.style,
    }
}

func (c *VNode) SetProps(p rtui.Props) rtui.VNode {
    // 实现属性设置逻辑
    return c
}
```

#### 4.4 CreateInstance（可选）
```go
func (c *VNode) CreateInstance() rtui.ComponentInstance {
    // Container 使用 Fiber-first 布局，不需要特殊 Instance
    return nil
}
```

---

### Phase 5: Fiber 层支持

#### 5.1 Fiber 添加 LayoutMode 属性
```go
// runtime/ui/fiber.go
type Fiber struct {
    // ... 现有字段 ...

    // ✨ 通用容器属性（方案 B 新增）
    LayoutMode    LayoutMode  // 布局模式

    // Flex 属性
    LayoutDirection  rtui.Direction
    LayoutAlign      rtui.Align
    LayoutCrossAlign rtui.Align
    LayoutGap        int

    // Grid 属性
    LayoutColumns    []interface{}  // Dimension 序列化
    LayoutRows       []interface{}
    LayoutColumnGap  int
    LayoutRowGap     int
    LayoutGridCells  []interface{}  // Cell 序列化

    // Wrap 属性
    LayoutWrapAlign  rtui.Align

    // Absolute 属性
    LayoutLeft       interface{}  // PositionValue 序列化
    LayoutTop        interface{}
    LayoutRight      interface{}
    LayoutBottom     interface{}
    LayoutAnchor     layout.Anchor

    // 边框属性（方案 A 集成）
    BorderStyle  layout.BorderStyle
    BorderLabel  string
}
```

#### 5.2 completeWork 同步 Container 属性
```go
// internal/render/reconciler.go
func completeWorkContainer(fiber *Fiber) {
    // 同步布局模式
    if mode, ok := fiber.Props["layoutMode"].(LayoutMode); ok {
        fiber.LayoutMode = mode
    }

    // 根据模式同步不同属性
    switch fiber.LayoutMode {
    case LayoutFlex:
        syncFlexProperties(fiber)
    case LayoutGrid:
        syncGridProperties(fiber)
    case LayoutWrap:
        syncWrapProperties(fiber)
    case LayoutAbsolute:
        syncAbsoluteProperties(fiber)
    }
}
```

---

### Phase 6: FiberToNodeAdapter 支持 Container

#### 6.1 扩展 GetFlexStyle
```go
func (a *FiberToNodeAdapter) GetFlexStyle() *layout.FlexStyle {
    if a.fiber == nil {
        return nil
    }

    // 支持多种 tag
    switch a.fiber.Tag {
    case "vstack", "hstack", "stack":
        // 原有逻辑...
        return style

    case "container":
        // ✨ 支持 Container 标签
        if a.fiber.LayoutMode != LayoutFlex {
            return nil
        }
        // 应用 Flex 属性...
        return style
    }

    return nil
}
```

#### 6.2 扩展 GetGridStyle
```go
func (a *FiberToNodeAdapter) GetGridStyle() *layout.GridStyle {
    if a.fiber == nil {
        return nil
    }

    switch a.fiber.Tag {
    case "grid":
        // 原有逻辑...
        return gridStyle

    case "container":
        // ✨ 支持 Container Grid 模式
        if a.fiber.LayoutMode != LayoutGrid {
            return nil
        }
        // 应用 Grid 属性...
        return gridStyle
    }

    return nil
}
```

#### 6.3 扩展 GetWrapStyle
```go
func (a *FiberToNodeAdapter) GetWrapStyle() *layout.WrapStyle {
    if a.fiber == nil {
        return nil
    }

    switch a.fiber.Tag {
    case "wrap":
        // 原有逻辑...

    case "container":
        // ✨ 支持 Container Wrap 模式
        if a.fiber.LayoutMode != LayoutWrap {
            return nil
        }
        // 应用 Wrap 属性...
    }
}
```

#### 6.4 扩展 GetAbsoluteStyle
```go
func (a *FiberToNodeAdapter) GetAbsoluteStyle() *layout.AbsoluteStyle {
    if a.fiber == nil {
        return nil
    }

    switch a.fiber.Tag {
    case "absolute":
        // 原有逻辑...

    case "container":
        // ✨ 支持 Container Absolute 模式
        if a.fiber.LayoutMode != LayoutAbsolute {
            return nil
        }
        // 应用 Absolute 属性...
    }
}
```

---

### Phase 7: 向后兼容层

#### 7.1 保留现有组件的接口
```go
// Stack 保留，内部委托给 Container
type Stack struct {
    container *container.VNode
}

func NewVStack() *Stack {
    return &Stack{
        container: container.NewFlex().Column(),
    }
}

// 实现所有原有方法...
```

#### 7.2 或者现有组件逐步废弃
- 保持现有组件 API 不变
- 内部改用 Container 实现
- 文档标注为 Deprecated
- 逐渐引导用户迁移到 Container

---

### Phase 8: 测试和验证

#### 8.1 单元测试
- [ ] VNode 所有方法
- [ ] Props 序列化/反序列化
- [ ] 各布局模式配置

#### 8.2 集成测试
- [ ] Flex 模式渲染
- [ ] Grid 模式渲染
- [ ] Wrap 模式渲染
- [ ] Absolute 模式渲染
- [ ] 边框渲染
- [ ] 模式切换

#### 8.3 性能测试
- [ ] 对比旧组件的性能
- [ ] 内存占用测试

---

## API 变更示例

### 多种布局方式对比

#### 旧 API: VStack
```go
vstack.New().
    Gap(1).
    SetChildren(
        text.New("A"),
        text.New("B"),
    )
```

#### 新 API: Container (Flex)
```go
container.NewFlex().
    Column().
    Gap(1).
    Add(
        text.New("A"),
        text.New("B"),
    )
```

---

#### 旧 API: HStack
```go
hstack.New().
    Align(AlignCenter).
    SetChildren(...)
```

#### 新 API: Container (Flex Row)
```go
container.NewFlex().
    Row().
    Align(AlignCenter).
    Add(...)
```

---

#### 旧 API: Grid
```go
grid.New().
    Columns(Fixed(10), Flex{1}).
    Rows(Auto{}).
    AddCell(...).
    AddCell(...)
```

#### 新 API: Container (Grid)
```go
container.NewGrid().
    Columns(Fixed(10), Flex{1}).
    Rows(Auto{}).
    AddCell(...).
    AddCell(...)
```

---

#### 旧 API: Wrap
```go
wrap.New().
    Gap(0).
    SetChildren(...)
```

#### 新 API: Container (Wrap)
```go
container.NewWrap().
    Gap(0).
    Add(...)
```

---

#### 旧 API: Absolute
```go
absolute.New(child).
    Left(AbsolutePos(5)).
    Top(AbsolutePos(3))
```

#### 新 API: Container (Absolute)
```go
container.NewAbsolute().
    SetLeft(AbsolutePos(5)).
    SetTop(AbsolutePos(3)).
    Add(child)
```

---

### 边框对比

#### 旧方式（需要组合）
```go
border.New(layout.BorderDouble).SetLabel("标题").SetChildren(
    vstack.New().Gap(1).SetChildren(...)
)
```

#### 新方式（原生支持）
```go
container.NewFlex().
    Border(layout.BorderDouble, "标题").
    Gap(1).
    Add(...)
```

---

## 简化后的组件生态

### 组件数量对比

| 方案 | 组件数 | 说明 |
|------|--------|------|
| **当前** | 6 个 | Stack, Grid, Wrap, Absolute, Border, 其他 |
| **方案 A** | 6 个 | Stack, Grid, Wrap, Absolute, Border (兼容), 其他 |
| **方案 B** | 1 个 | Container (统一所有) |

### 复杂度对比

| 方案 | 学习成本 | API 复杂度 | 灵活性 |
|------|----------|-----------|--------|
| **当前** | 高（多个组件） | 中 | 低（需组合） |
| **方案 A** | 中（熟悉 Stack 即可） | 低 | 中 |
| **方案 B** | 低（只学 Container） | 高（配置多） | 高 |

---

## 迁移策略

### 阶段 1：并置运行
- 新增 Container 组件
- 保留所有现有组件
- 文档标注 Container 为推荐方式

### 阶段 2：文档引导
- 新示例优先使用 Container
- 旧组件标注为 Legacy

### 阶段 3：逐步废弃
- 发布废弃通知
- 提供迁移工具/脚本
- 收集用户反馈

### 阶段 4：移除（可选）
- 两个大版本后移除旧组件
- 保留兼容层（可选）

---

## 风险评估

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|----------|
| 用户不适应 | 中 | 高 | 充分文档引导，保留旧组件 |
| 性能下降 | 低 | 低 | 容器只是配置分发，无额外开销 |
| 配置过于复杂 | 高 | 中 | 提供预设方法（NewFlex, NewGrid 等） |
| 破坏现有代码 | 低 | 高 | 保留旧 API，内部改用 Container 实现 |
| 测试覆盖不足 | 中 | 高 | 充分的单元测试和集成测试 |

---

## 时间估算

| Phase | 任务 | 预计时间 |
|-------|------|----------|
| 1 | 基础框架 | 1 天 |
| 2 | VNode 实现 | 2 天 |
| 3 | 布局方法 | 2 天 |
| 4 | 内容管理 | 1 天 |
| 5 | Fiber 支持 | 1.5 天 |
| 6 | Adapter 扩展 | 1.5 天 |
| 7 | 兼容层 | 1 天 |
| 8 | 测试验证 | 2 天 |
| **合计** | | **12 天** |

---

## 总结

### 优势
- ✅ 一个组件解决所有布局需求
- ✅ 组件数量大幅减少
- ✅ API 高度统一
- ✅ 易于扩展新布局模式

### 劣势
- ⚠️ 配置较多，学习曲线
- ⚠️ 对现有用户改动较大
- ⚠️ 需要长时间迁移

### 适用场景
- 新项目可以采用 Container
- 复杂布局项目简化组件层次
- 团队可以统一组件使用规范
