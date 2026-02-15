# 布局系统差异分析报告

## 执行摘要

本报告分析了 Mint TUI 框架中两个布局系统的差异：

1. **runtime/compute/engine.go** - 基于 VNode 的约束驱动布局引擎（生产环境）
2. **runtime/layout/** - 通用的 Flexbox 布局库（实验性/未使用）

**关键发现**：
- `compute/` 是实际使用的生产系统，功能完整且高度优化
- `layout/` 是独立的通用库，未与 Mint 的 VNode/Fiber 架构集成
- 两者设计目标不同，不应直接比较，而是互补关系

---

## 1. 架构设计对比

### 1.1 设计目标

| 维度 | compute/ | layout/ |
|------|----------|---------|
| **目标** | 为 Mint TUI 提供生产级布局引擎 | 提供通用 Flexbox 布局算法 |
| **抽象层次** | 高度集成 VNode/Fiber 架构 | 独立的抽象节点接口 |
| **依赖关系** | 依赖 runtime/ui、reconciler、event | 零依赖（纯 Go） |
| **使用状态** | ✅ 生产环境使用 | ❌ 未集成（实验性） |

### 1.2 节点抽象

#### compute/ - VNode 作为输入
```go
// 基于 VNode 的布局
type VNode = rtui.VNode

type ComputedBox struct {
    VNode     VNode        // 引用原始 VNode
    NodeID    uint64       // Fiber 节点 ID（稳定标识）
    Layer     rtui.Layer   // 渲染层
    Box       runtime.Box  // 计算后的位置和尺寸
    Children  []*ComputedBox
    // ...
}
```

**特点**：
- 直接操作 VNode 树
- 通过 Fiber.NodeID 实现稳定标识
- 支持 VNode 的所有属性（Props、Key、Type 等）
- 与 Diff 算法紧密集成

#### layout/ - 通用 Node 接口
```go
// 抽象节点接口
type Node interface {
    ID() string
    Type() string
    Children() []Node
    GetPosition() (x, y int)
    SetPosition(x, y int)
    GetSize() (width, height int)
    SetSize(width, height int)
}

type LayoutBox struct {
    ID       string
    X, Y     int
    Width, Height int
    Children []*LayoutBox
}
```

**特点**：
- 独立的抽象接口，不依赖 VNode
- 节点必须实现接口方法
- 更通用，可适配任何树形结构
- 缺少 Mint 特定的功能（如 Key、Props）

---

## 2. 布局算法对比

### 2.1 约束系统

#### compute/ - BoxConstraints
```go
type BoxConstraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}

// 辅助方法
func (c BoxConstraints) HasBoundedWidth() bool
func (c BoxConstraints) HasBoundedHeight() bool
func (c BoxConstraints) SubtractPadding(h, v int) BoxConstraints
```

**特点**：
- 支持 Infinity 表示无界
- 提供丰富的辅助方法
- 与 Paint 系统共享约束定义
- 支持边距和内边距计算

#### layout/ - Constraints
```go
type Constraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}

const MaxInt = 1 << 30  // 使用大整数表示无界

// 辅助方法
func (c Constraints) IsTight() bool
func (c Constraints) IsBounded() bool
func (c Constraints) Constrain(width, height int) (int, int)
```

**特点**：
- 使用 `MaxInt = 1<<30` 代替 Infinity
- 提供约束检查方法
- 独立的约束系统
- 更简单的方法集

### 2.2 布局流程

#### compute/ - 两阶段布局

**阶段 1：测量（自底向上）**
```go
func (e *Engine) buildComputedBox(vnode VNode, fiber *reconciler.Fiber,
    parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox

// 支持单次测量（如果实现 LayoutMeasurer）
measurement := e.TryMeasureLayout(vnode, constraints)
if len(measurement.ChildConstraints) > 0 {
    // 单次路径：同时计算尺寸和子节点约束
    box.Box.Width = measurement.Size.Width
    box.Box.Height = measurement.Size.Height
    // 使用预计算的约束构建子节点
}
```

**阶段 2：定位（自顶向下）**
```go
func (e *Engine) calculatePositions(box *ComputedBox, x, y int) {
    box.Box.X = x
    box.Box.Y = y

    // 根据节点类型布局子节点
    switch tagger.Tag() {
    case "hstack":
        e.layoutHStack(box, x, y)
    case "vstack":
        e.layoutVStack(box, x, y)
    case "bordered":
        e.layoutBordered(box, x, y)
    }
}
```

**关键特性**：
- 单次测量优化（LayoutMeasurer 接口）
- 支持 Fiber 树遍历
- 区分主轴和交叉轴对齐
- 支持文本对齐和填充

#### layout/ - 简化的两阶段布局

**阶段 1：测量**
```go
func (f *FlexLayout) Measure(constraints Constraints) Size {
    // Phase 1: 测量所有子节点
    childSizes := make([]Size, len(f.children))
    totalMainSize := 0
    maxCrossSize := 0

    for i, child := range f.children {
        childSizes[i] = measurable.Measure(childConstraints)
        // 累加主轴尺寸，取最大交叉轴尺寸
    }

    return Size{Width: width, Height: height}
}
```

**阶段 2：布局子节点**
```go
func (f *FlexLayout) LayoutChildren(width, height int) []LayoutBox {
    // Phase 1: 测量所有子节点
    // Phase 2: 计算剩余空间
    remainingSpace := availableSize - fixedTotal - totalGap

    // Phase 3: 分配剩余空间给可伸缩节点
    for i, flex := range flexibleChildren {
        extra := (remainingSpace * flex.Grow) / flexGrowTotal
        finalSizes[i].Width += extra
    }

    // Phase 4: 计算位置
    return boxes
}
```

**关键特性**：
- 标准的 Flexbox 算法
- 支持 flex-grow、flex-shrink、flex-basis
- 支持主轴和交叉轴对齐
- 更简单，缺少优化

---

## 3. 核心功能对比

### 3.1 Flexbox 支持

#### compute/ - HStack/VStack

**实现位置**：`measureLayoutChildren()` (604-914行)

```go
// HStack: 水平布局
if layoutInfo.IsHorizontal {
    // 识别 flex 和固定子节点
    for i, child := range children {
        childInfo := rtui.GetLayoutInfo(child)
        if childInfo.Flex > 0 {
            flexChildren = append(flexChildren, ...)
        } else {
            // 测量固定子节点
            fixedWidth += childSize.Width
        }
    }

    // 分配剩余空间
    if constraints.HasBoundedWidth() {
        availableWidth := constraints.MaxWidth - paddingWidth
        remainingSpace := availableWidth - fixedWidth
        for _, fc := range flexChildren {
            flexWidth := (remainingSpace * fc.factor) / flexTotalFactor
        }
    }
}
```

**特性**：
- ✅ Flex 布局（flex factor）
- ✅ 主轴对齐（Start/Center/End/SpaceBetween/Around）
- ✅ 交叉轴对齐（Start/Center/End/Stretch）
- ❌ 不支持 flex-shrink 和 flex-basis
- ❌ 不支持反向排列（RowReverse/ColumnReverse）

#### layout/ - FlexLayout

**实现位置**：`flex.go` (108-573行)

```go
type FlexLayout struct {
    style *FlexStyle
}

type FlexStyle struct {
    Direction          FlexDirection
    MainAxis           MainAxisAlignment
    CrossAxis          CrossAxisAlignment
    Gap, CrossGap      int
    Padding            Padding
    FlexibleChildren   map[int]*Flex
}

type Flex struct {
    Grow   int  // 放大比例
    Shrink int  // 缩小比例
    Basis  int  // 基础尺寸
}
```

**特性**：
- ✅ 完整的 Flexbox 支持
- ✅ Flex-grow、flex-shrink、flex-basis
- ✅ 反向排列（RowReverse/ColumnReverse）
- ✅ 交叉轴间距（CrossGap）
- ✅ 更标准的实现

**优势**：
- 更接近 CSS Flexbox 规范
- 支持更复杂的布局场景
- 更可预测的行为

### 3.2 缓存机制

#### compute/ - LayoutCache

```go
type LayoutCache struct {
    entries map[LayoutCacheKey]*LayoutCacheEntry
}

type LayoutCacheKey struct {
    VNodeType   string
    VNodeKey    string
    Constraints BoxConstraints
    ContentHash uint64
    PropsHash   uint64
}

// 缓存策略：仅缓存叶子节点
if isLeaf || vnode.Key() != "" {
    if cached, ok := e.cache.Get(cacheKey); ok {
        box.Box = cached.Box
        return box
    }
}
```

**特点**：
- 基于节点类型、key、约束和内容哈希
- 仅缓存叶子节点（避免子树复杂性）
- 支持 dirty tracking
- 统计缓存命中率

#### layout/ - Cache

```go
type Cache struct {
    entries map[string]*CachedLayout
    maxSize int
}

type CachedLayout struct {
    Result    *LayoutResult
    Timestamp time.Time
    HitCount  int
}

func (c *Cache) makeKey(node Node, constraints Constraints) string {
    constraintKey := c.constraintsKey(constraints)
    nodesHash := c.nodesHash(node)
    return constraintKey + ":" + nodesHash
}
```

**特点**：
- 基于节点树 SHA256 哈希
- 缓存完整布局结果
- LRU 淘汰策略
- 基于时间戳驱逐

**对比**：
- `compute/` 更精细（叶子节点）
- `layout/` 更通用（完整树）
- `compute/` 更高效（SHA256 vs 简单哈希）

### 3.3 Fiber 集成

#### compute/ - Fiber 支持

```go
func (e *Engine) Layout(vnode VNode, fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints) (*ComputedLayout, error) {

    // 传递 Fiber 到 buildComputedBox
    root := e.buildComputedBox(vnode, fiber, nil, constraints)

    // 设置 NodeID
    if fiber != nil {
        box.NodeID = fiber.NodeID
        box.ChildFiber = fiber
    }

    // 构建 HitMap
    hitMap := e.buildHitMapFromComputedBoxes(root)
}

func (e *Engine) LayoutFiber(root *rtui.Fiber,
    constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    // 纯 Fiber 布局（Phase 2+）
}
```

**特性**：
- ✅ 完整的 Fiber 集成
- ✅ NodeID 传播（稳定标识）
- ✅ DiffKey 匹配（处理 VNode 重排序）
- ✅ Layer 支持（渲染层）
- ✅ HitMap 构建（事件处理）

#### layout/ - 无 Fiber 支持

```go
// 不支持 Fiber
// 只能处理通用 Node 接口
```

**影响**：
- ❌ 无法利用 Fiber 的优化
- ❌ 无法与 Diff 算法集成
- ❌ 缺少稳定标识（NodeID）
- ❌ 无事件处理支持

---

## 4. 性能优化

### 4.1 compute/ 优化

1. **单次测量** (TryMeasureLayout)
   - 避免 O(N²) 重复测量
   - 预计算子节点约束

2. **Flex 分布缓存** (FlexDistributionInfo)
   - 缓存 flex 子节点分布
   - 避免重复计算

3. **叶子节点缓存**
   - 仅缓存叶子节点
   - 减少内存开销

4. **Dirty Tracking**
   - 仅重新布局脏节点
   - 支持增量更新

5. **Fiber 匹配优化**
   - 使用 DiffKey 而非索引
   - 处理动态列表

### 4.2 layout/ 优化

1. **完整结果缓存**
   - 缓存整个布局树
   - LRU 淘汰

2. **节点哈希**
   - SHA256 树哈希
   - 精确的缓存键

3. **布局统计**
   - 缓存命中率
   - 性能监控

**对比**：
- `compute/` 更精细的优化策略
- `layout/` 更简单但通用

---

## 5. 代码质量与维护

### 5.1 代码规模

| 文件 | 行数 | 职责 |
|------|------|------|
| `compute/engine.go` | 2000+ | 主布局引擎 |
| `compute/types.go` | 243 | 类型定义 |
| `compute/cache.go` | 200+ | 缓存实现 |
| `layout/types.go` | 434 | 引擎+类型 |
| `layout/flex.go` | 591 | Flexbox 算法 |
| `layout/cache.go` | 171 | 缓存实现 |

### 5.2 测试覆盖

#### compute/
```
engine_nodeid_test.go        # NodeID 测试
fiber_bench_test.go          # 性能基准测试
fiber_integration_test.go    # Fiber 集成测试
fiber_layout_test.go         # Fiber 布局测试
```

#### layout/
```
flex_test.go                 # Flexbox 测试
cache_test.go                # 缓存测试
constraints_test.go          # 约束测试
measure_test.go              # 测量测试
complex_test.go              # 复杂布局测试
```

### 5.3 文档

#### compute/
- ✅ 内联注释详细
- ✅ 设计文档引用
- ❌ 缺少 README

#### layout/
- ✅ 完整的 README.md
- ✅ 使用示例
- ✅ API 文档
- ✅ 性能优化说明

---

## 6. 使用建议

### 6.1 当前状态

**compute/** - ✅ 生产就绪
- 与 VNode/Fiber 深度集成
- 完整的功能集
- 性能优化到位
- 实际项目中使用

**layout/** - ❌ 未集成
- 独立的通用库
- 未与 VNode 集成
- 缺少 Mint 特定功能
- 仅作为参考实现

### 6.2 未来方向

#### 选项 A：保持现状
- 继续使用 `compute/` 作为主布局引擎
- 保留 `layout/` 作为通用 Flexbox 库
- 必要时从 `layout/` 移植算法到 `compute/`

#### 选项 B：整合到 compute/
- 将 `layout/` 的 Flexbox 实现移植到 `compute/`
- 添加缺失功能（flex-shrink、flex-basis、反向排列）
- 保持与 VNode/Fiber 的集成

#### 选项 C：创建适配层
- 在 `layout/` 中创建 VNode 适配器
- 使 `layout/` 可以处理 VNode 树
- 逐步迁移到 `layout/`

### 6.3 推荐方案

**推荐选项 B**：从 `layout/` 移植增强功能到 `compute/`

**理由**：
1. `compute/` 已经与 VNode/Fiber 深度集成
2. `layout/` 的 Flexbox 实现更完整
3. 移植成本低于创建适配层
4. 保持现有架构的稳定性

**具体步骤**：
1. 增强 `compute/` 的 measureLayoutChildren：
   - 添加 flex-shrink 支持
   - 添加 flex-basis 支持
   - 添加反向排列（RowReverse/ColumnReverse）

2. 优化 `compute/` 的缓存策略：
   - 参考 `layout/` 的树哈希
   - 改进缓存键生成

3. 添加 `compute/` 的文档：
   - 创建 README.md
   - 添加使用示例
   - 记录 API 设计

---

## 7. 功能差异总结

### 7.1 compute/ 独有功能

| 功能 | 描述 |
|------|------|
| Fiber 集成 | 完整的 Fiber 树支持 |
| NodeID 传播 | 稳定的节点标识 |
| DiffKey 匹配 | 处理动态列表重排序 |
| Layer 支持 | 多层渲染（Modal/Tooltip 等） |
| HitMap 构建 | 事件处理支持 |
| 单次测量优化 | 避免重复测量 |
| Flex 分布缓存 | O(N²) → O(N) |
| Dirty Tracking | 增量布局 |
| 文本对齐 | TextAlign 支持 |
| 边框处理 | Bordered 节点特殊处理 |
| 表格布局 | Table 节点支持 |

### 7.2 layout/ 独有功能

| 功能 | 描述 |
|------|------|
| Flex-shrink | 子节点缩小比例 |
| Flex-basis | 子节点基础尺寸 |
| 反向排列 | RowReverse/ColumnReverse |
| 交叉轴间距 | CrossGap 支持 |
| 标准 Flexbox | 更接近 CSS 规范 |
| 通用接口 | 可适配任何树形结构 |
| 完整文档 | README + 示例 |

### 7.3 共同功能

| 功能 | compute/ | layout/ |
|------|----------|---------|
| Flex-grow | ✅ | ✅ |
| 主轴对齐 | ✅ | ✅ |
| 交叉轴对齐 | ✅ | ✅ |
| Gap 间距 | ✅ | ✅ |
| Padding | ✅ | ✅ |
| 缓存 | ✅ | ✅ |
| 约束系统 | ✅ | ✅ |
| 两阶段布局 | ✅ | ✅ |

---

## 8. 性能对比

### 8.1 理论分析

| 指标 | compute/ | layout/ |
|------|----------|---------|
| **时间复杂度** | O(N) | O(N) |
| **空间复杂度** | O(N) | O(N) |
| **缓存命中率** | 高（叶子节点） | 中（完整树） |
| **增量更新** | ✅ Dirty tracking | ❌ 全量重新计算 |
| **单次测量** | ✅ | ❌ |

### 8.2 实际性能

**compute/** 优势：
- Fiber 匹配避免 VNode 重排序问题
- 单次测量减少递归深度
- Dirty tracking 减少计算量

**layout/** 优势：
- 更简单的实现可能更快（小规模）
- SHA256 哈希更精确但更慢

**结论**：`compute/` 在大规模动态 UI 中性能更优

---

## 9. 代码示例对比

### 9.1 HStack 布局

#### compute/ 使用方式
```go
// 创建 HStack
hstack := rtui.HStack(
    rtui.Text("Hello"),
    rtui.Text("World"),
).Props(
    rtui.Gap(1),
    rtui.Align(rtui.AlignCenter),
    rtui.CrossAlign(rtui.AlignCenter),
)

// 布局
engine := compute.NewEngine()
layout, err := engine.Layout(hstack, nil, constraints)

// 访问结果
for _, box := range layout.Root.Children {
    fmt.Printf("Node %s: x=%d, y=%d, w=%d, h=%d\n",
        box.VNode.Key(), box.Box.X, box.Box.Y,
        box.Box.Width, box.Box.Height)
}
```

#### layout/ 使用方式
```go
// 创建 FlexLayout
container := layout.NewFlexLayout("container", []layout.Node{
    &TextNode{id: "1", text: "Hello"},
    &TextNode{id: "2", text: "World"},
})
container.SetDirection(layout.FlexRow)
container.SetGap(1)
container.SetMainAxis(layout.Center)
container.SetCrossAxis(layout.CrossCenter)

// 布局
engine := layout.NewEngine()
result := engine.Layout(container, constraints)

// 访问结果
for _, box := range result.Boxes {
    fmt.Printf("Node %s: x=%d, y=%d, w=%d, h=%d\n",
        box.ID, box.X, box.Y, box.Width, box.Height)
}
```

**差异**：
- `compute/` 使用 VNode 构建（更简洁）
- `layout/` 需要手动创建节点（更通用）

### 9.2 Flex 布局

#### compute/
```go
vstack := rtui.VStack(
    rtui.Text("Header").Props(rtui.Flex(0)),
    rtui.Text("Content").Props(rtui.Flex(1)),
    rtui.Text("Footer").Props(rtui.Flex(0)),
)
```

#### layout/
```go
container := layout.NewFlexLayout("container", children)
container.SetDirection(layout.FlexColumn)
container.SetFlex(0, 0, 0)  // Header: 不伸缩
container.SetFlex(1, 1, 0)  // Content: 伸缩
container.SetFlex(2, 0, 0)  // Footer: 不伸缩
```

**差异**：
- `compute/` 更简洁（Props 方式）
- `layout/` 更明确（方法调用）

---

## 10. 结论

### 10.1 主要差异总结

1. **架构定位**
   - `compute/`: Mint 专用的生产级布局引擎
   - `layout/`: 通用 Flexbox 算法库

2. **集成程度**
   - `compute/`: 深度集成 VNode/Fiber
   - `layout/`: 独立实现，无集成

3. **功能完整性**
   - `compute/`: 缺少部分 Flexbox 特性
   - `layout/`: 更接近标准 Flexbox

4. **性能优化**
   - `compute/`: 更多优化（单次测量、缓存、dirty tracking）
   - `layout/`: 基础优化

5. **代码质量**
   - `compute/`: 生产级别，测试完善
   - `layout/`: 实验级别，文档完善

### 10.2 建议

1. **短期**：保持现状
   - 继续使用 `compute/`
   - 参考 `layout/` 优化算法

2. **中期**：功能移植
   - 将 `layout/` 的完整 Flexbox 支持移植到 `compute/`
   - 添加 flex-shrink、flex-basis、反向排列

3. **长期**：统一架构
   - 考虑是否需要通用化 `compute/`
   - 或创建适配层集成 `layout/`

### 10.3 最终评价

**compute/** 是当前 Mint TUI 框架的正确选择，因为它：
- ✅ 与现有架构深度集成
- ✅ 性能优化到位
- ✅ 功能满足需求
- ✅ 生产环境验证

**layout/** 是一个有价值的参考实现，可以：
- 📚 作为 Flexbox 算法的学习资源
- 🔧 为 `compute/` 提供功能增强灵感
- 🧪 用于算法验证和测试

---

## 附录

### A. 文件映射

| 功能 | compute/ | layout/ |
|------|----------|---------|
| 主引擎 | engine.go | types.go (Engine) |
| 类型定义 | types.go | types.go |
| Flexbox | measureLayoutChildren() | flex.go |
| 缓存 | cache.go | cache.go |
| 测试 | *_test.go | *_test.go |

### B. 关键方法映射

| compute/ | layout/ | 功能 |
|----------|---------|------|
| Layout() | Layout() | 执行布局 |
| buildComputedBox() | layoutNode() | 构建布局盒子 |
| calculatePositions() | LayoutChildren() | 计算位置 |
| measureVNode() | Measure() | 测量尺寸 |
| measureLayoutChildren() | Measure() | Flexbox 测量 |

### C. 类型映射

| compute/ | layout/ | 说明 |
|----------|---------|------|
| VNode | Node | 节点抽象 |
| BoxConstraints | Constraints | 约束系统 |
| ComputedBox | LayoutBox | 布局结果 |
| ComputedLayout | LayoutResult | 布局输出 |
| Size | Size | 尺寸 |

---

**报告生成时间**: 2025-02-15
**分析版本**: fiber_first_phase_two 分支
**分析者**: AI Assistant (Crush)
