# runtime/layout 功能补全计划

## 设计原则

### 核心原则：保持 layout 独立性

`runtime/layout` 必须保持为一个**纯布局库**，不依赖任何 Mint 特定类型：

```
✅ 允许                          ❌ 禁止
─────────────────────────────────────────────
runtime/layout.Node 接口        runtime/ui.VNode
runtime/layout.Constraints      runtime/ui.Fiber
基本类型 (int, string)          reconciler.Fiber
标准库 (sync, fmt)              任何 Mint UI 类型
```

### 适配器模式

所有与 VNode/Fiber 的集成必须通过**适配器**在 `internal/render` 中实现：

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  runtime/ui     │     │  internal/render │     │ runtime/layout  │
│  VNode/Fiber    │────▶│  FiberAdapter    │────▶│ Node interface  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                              (适配层)
```

---

## 功能差距与实现计划

### Phase 6: Margin 支持 (优先级: 高)

**目标**: 在 `runtime/layout` 中添加 margin 支持

**修改文件**:

| 文件 | 修改内容 |
|------|----------|
| `runtime/layout/types.go` | 添加 `Margin` 结构体 |
| `runtime/layout/flex.go` | `FlexStyle` 添加 `Margin` 字段 |
| `runtime/layout/flex.go` | 布局计算时处理 margin |
| `runtime/layout/margin_test.go` | 新建测试文件 |

**实现要点**:
```go
// runtime/layout/types.go
type Margin struct {
    Top    int
    Right  int
    Bottom int
    Left   int
}

// runtime/layout/flex.go - FlexStyle 添加
type FlexStyle struct {
    // ... existing fields
    Margin Margin  // 新增
}
```

**适配器扩展**:
```go
// internal/render/fiber_adapter.go - extractFlexStyle 添加
style.Margin = layout.Margin{
    Top:    layoutInfo.Margin[0],
    Right:  layoutInfo.Margin[1],
    Bottom: layoutInfo.Margin[2],
    Left:   layoutInfo.Margin[3],
}
```

**预计工作量**: 2-3 天

---

### Phase 7: Absolute Positioning (优先级: 高)

**目标**: 支持绝对定位布局

**修改文件**:

| 文件 | 修改内容 |
|------|----------|
| `runtime/layout/position.go` | 新建 - PositionType, Position 结构体 |
| `runtime/layout/types.go` | Node 接口添加 `GetPositionType()` (可选接口) |
| `runtime/layout/position.go` | AbsolutePositioner 实现 |
| `runtime/layout/position_test.go` | 新建测试文件 |

**实现要点**:
```go
// runtime/layout/position.go (新文件)

type PositionType int

const (
    PositionRelative PositionType = iota
    PositionAbsolute
)

type Position struct {
    Type   PositionType
    Top    *int  // nil = auto
    Left   *int
    Right  *int
    Bottom *int
}

// 可选接口 - 节点可实现
type Positionable interface {
    Node
    GetPosition() Position
}

// AbsolutePositioner 处理绝对定位
type AbsolutePositioner struct {
    parentSize Size
}

func (ap *AbsolutePositioner) Layout(node Node, constraints Constraints) *LayoutBox {
    // 1. 正常布局 relative 子节点
    // 2. 跳过 absolute 子节点
    // 3. 二次遍历计算 absolute 位置
}
```

**适配器扩展**:
```go
// internal/render/fiber_adapter.go
type FiberToNodeAdapter struct {
    // ... existing fields
    position layout.Position
}

func (a *FiberToNodeAdapter) GetPosition() layout.Position {
    // 从 Fiber.LayoutPosition 或 VNode props 提取
}
```

**预计工作量**: 3-4 天

---

### Phase 8: Border 容器 (优先级: 高)

**目标**: 支持带边框的容器布局

**修改文件**:

| 文件 | 修改内容 |
|------|----------|
| `runtime/layout/border.go` | 新建 - BorderStyle, BorderedNode |
| `runtime/layout/border_test.go` | 新建测试文件 |

**实现要点**:
```go
// runtime/layout/border.go (新文件)

type BorderStyle int

const (
    BorderNone BorderStyle = iota
    BorderSingle
    BorderDouble
    BorderRounded
)

type Border struct {
    Style BorderStyle
    Width int  // 边框占用的宽度 (通常为 2)
}

// BorderedNode 包装节点添加边框
type BorderedNode struct {
    child    Node
    border   Border
    label    string
}

func NewBorderedNode(child Node, border Border) *BorderedNode

// 实现 Node 接口
func (b *BorderedNode) Children() []Node {
    return []Node{b.child}
}

// 实现 Measurable 接口
func (b *BorderedNode) Measure(constraints Constraints) Size {
    // 1. 减去边框空间
    innerConstraints := constraints.Subtract(b.border.Width, b.border.Width)
    // 2. 测量子节点
    childSize := b.child.Measure(innerConstraints)
    // 3. 加上边框空间返回
    return Size{
        Width:  childSize.Width + b.border.Width,
        Height: childSize.Height + b.border.Width,
    }
}
```

**注意**: 边框绘制逻辑仍由 `PaintEngine` 处理，`runtime/layout` 只计算布局空间。

**预计工作量**: 2-3 天

---

### Phase 9: Table 布局 (优先级: 中)

**目标**: 支持表格布局

**修改文件**:

| 文件 | 修改内容 |
|------|----------|
| `runtime/layout/table.go` | 新建 - TableNode, TableCell |
| `runtime/layout/table_test.go` | 新建测试文件 |

**实现要点**:
```go
// runtime/layout/table.go (新文件)

type TableNode struct {
    rows     [][]Node
    colWidths []int  // 缓存列宽
}

func NewTableNode(rows [][]Node) *TableNode

func (t *TableNode) Measure(constraints Constraints) Size {
    // 1. 第一遍：计算每列最大宽度
    // 2. 第二遍：应用列宽约束
    // 3. 计算总尺寸
}

func (t *TableNode) Layout(constraints Constraints) *LayoutBox {
    // 为每个单元格生成 LayoutBox
}
```

**预计工作量**: 2-3 天

---

### Phase 10: Multi-layer / RenderPlanes 支持 (优先级: 中)

**目标**: 支持分层渲染 (modal, overlay, tooltip)

**修改文件**:

| 文件 | 修改内容 |
|------|----------|
| `runtime/layout/layer.go` | 新建 - Layer, LayeredNode |
| `runtime/layout/types.go` | LayoutBox 添加 Layer/ZIndex 字段 |
| `runtime/layout/layer_test.go` | 新建测试文件 |

**实现要点**:
```go
// runtime/layout/layer.go (新文件)

type Layer int

const (
    LayerBase Layer = iota
    LayerModal
    LayerOverlay
    LayerTooltip
)

// Layered 接口 - 节点可实现
type Layered interface {
    Node
    GetLayer() Layer
    GetZIndex() int
}

// LayoutBox 扩展
type LayoutBox struct {
    // ... existing fields
    Layer  Layer  // 新增
    ZIndex int    // 新增
}

// LayeredLayoutResult 分层结果
type LayeredLayoutResult struct {
    Layers map[Layer][]*LayoutBox
    Root   *LayoutBox
}
```

**适配器扩展**:
```go
// internal/render/fiber_adapter.go
func (a *FiberToNodeAdapter) GetLayer() layout.Layer {
    if a.fiber == nil {
        return layout.LayerBase
    }
    // 从 Fiber.Layer 转换
    switch a.fiber.Layer {
    case rtui.LayerModal:
        return layout.LayerModal
    // ...
    }
    return layout.LayerBase
}
```

**预计工作量**: 2-3 天

---

### Phase 11: FlexShrink 计算 (优先级: 低)

**目标**: 实现 flexShrink 实际计算逻辑

**当前状态**: `Flex.Shrrink` 字段已定义但未使用

**修改文件**:

| 文件 | 修改内容 |
|------|----------|
| `runtime/layout/flex.go` | 实现 shrink 算法 |

**算法**:
```go
// 当容器空间不足时，按 shrink 比例缩小子节点
func distributeShrink(availableSpace int, children []Node, styles []*FlexStyle) []int {
    totalShrink := 0
    for i, child := range children {
        if styles[i].Flex.Shrink > 0 {
            totalShrink += styles[i].Flex.Shrink
        }
    }
    
    if totalShrink == 0 {
        return nil  // 不收缩
    }
    
    // 按 shrink 比例分配收缩量
    sizes := make([]int, len(children))
    for i, child := range children {
        shrinkRatio := float64(styles[i].Flex.Shrink) / float64(totalShrink)
        sizes[i] = originalSize[i] - int(float64(deficit) * shrinkRatio)
    }
    return sizes
}
```

**预计工作量**: 1-2 天

---

## 实现顺序

```
Week 1: Phase 6 (Margin) + Phase 7 (Absolute Position)
Week 2: Phase 8 (Border) + Phase 9 (Table)
Week 3: Phase 10 (Layer) + Phase 11 (FlexShrink)
Week 4: 集成测试 + 性能优化
```

---

## 适配器同步更新清单

每完成一个 Phase，需要同步更新：

| 适配器文件 | 更新内容 |
|------------|----------|
| `internal/render/fiber_adapter.go` | 添加新接口实现 |
| `internal/render/layout_switcher.go` | 更新 LayoutResult 转换 |
| `docs/layout/PARALLEL_RENDERING_PIPELINE.md` | 更新功能对比表 |

---

## 测试策略

### 单元测试 (runtime/layout/*_test.go)
- 纯布局逻辑测试
- 不依赖 VNode/Fiber
- 使用 mock 节点

### 集成测试 (internal/render/*_test.go)
- 适配器功能测试
- 两个引擎结果对比

### 基准测试
- 性能对比
- 缓存命中率

---

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| 破坏现有功能 | 严格保持接口兼容 |
| 性能回退 | 每个阶段运行基准测试 |
| 适配器复杂度增长 | 提取公共转换逻辑 |

---

## 验收标准

每个 Phase 完成标准：

1. ✅ 所有单元测试通过
2. ✅ 适配器更新完成
3. ✅ 文档更新
4. ✅ 无循环导入
5. ✅ `runtime/layout` 不导入 `runtime/ui` 或 `internal/reconciler`
