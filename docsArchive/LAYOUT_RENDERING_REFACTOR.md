# 布局与渲染系统重构实施计划

> **版本**: 1.0
> **日期**: 2026-02-05
> **目标**: 重构布局系统以支持 Fiber 和非 Fiber 模式，实现约束驱动布局

---

## 一、背景与问题分析

### 1.1 当前架构的问题

基于 `internal/render/declarative_node.go` 的分析，当前实现存在以下问题：

| 问题 | 描述 | 影响 |
|------|------|------|
| **分散的测量逻辑** | 按钮、边框、表格等组件的测量散落在不同地方，需要特殊处理 | 维护困难，新增组件容易遗漏 |
| **缺少约束系统** | 没有 `Constraints` 结构，组件无法响应父容器约束 | 无法实现自适应布局 |
| **边框耦合测量** | 边框宽度与内容宽度耦合计算 | 边框应作为装饰层，不应影响内容测量 |
| **组件不自测量** | 测量由外部遍历树完成，而非组件自己实现 `Measure` 接口 | 不符合设计文档的组件模型 |
| **无脏标记优化** | 没有布局脏标记，每次都全量计算 | 性能问题 |
| **Fiber/非 Fiber 双路径** | 两套独立的渲染逻辑，代码重复 | 维护成本高 |

### 1.2 设计文档要求

根据 `framework/docs/ui/idea/idea2*.md`：

1. **布局必须从组件行为中剥离** (`idea2_layout.md`)
2. **组件只提供 Measure 和 Paint 能力**
3. **约束驱动布局**: Parent 给约束 → Child 返回尺寸 → Parent 排布
4. **布局是纯计算阶段，不可有副作用**
5. **增量布局优化**: 只重新计算脏子树
6. **增量 Reconcile**: 跳过未变化组件子树

### 1.3 重构目标

1. 实现 `Measurable` 接口，所有组件自测量
2. 引入 `Constraints` 系统，支持约束驱动布局
3. 边框作为装饰层，独立于布局计算
4. 支持 Fiber 和非 Fiber 双模式
5. 实现脏标记优化

---

## 二、核心架构设计

### 2.1 约束系统

```go
// Constraints 定义父容器对子组件的尺寸约束
type Constraints struct {
    MinWidth  int
    MaxWidth  int
    MinHeight int
    MaxHeight int
}

// 特殊值
const (
    Infinity = 1<<30 - 1  // 表示无限制
)

// 创建常用约束的辅助函数
func TightConstraints(width, height int) Constraints {
    return Constraints{
        MinWidth: width, MaxWidth: width,
        MinHeight: height, MaxHeight: height,
    }
}

func LooseConstraints(maxWidth, maxHeight int) Constraints {
    return Constraints{
        MinWidth: 0, MaxWidth: maxWidth,
        MinHeight: 0, MaxHeight: maxHeight,
    }
}

func UnboundedConstraints() Constraints {
    return Constraints{
        MinWidth: 0, MaxWidth: Infinity,
        MinHeight: 0, MaxHeight: Infinity,
    }
}

// Constrain 将尺寸约束到指定范围内
func (c Constraints) Constrain(width, height int) (int, int) {
    w := width
    if w < c.MinWidth {
        w = c.MinWidth
    } else if w > c.MaxWidth {
        w = c.MaxWidth
    }
    h := height
    if h < c.MinHeight {
        h = c.MinHeight
    } else if h > c.MaxHeight {
        h = c.MaxHeight
    }
    return w, h
}

// IsTight 检查约束是否为固定尺寸
func (c Constraints) IsTight() bool {
    return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
}

// IsBounded 检查是否有上界
func (c Constraints) IsBounded() bool {
    return c.MaxWidth < Infinity || c.MaxHeight < Infinity
}
```

### 2.2 Measurable 接口

```go
// Size 表示组件的测量尺寸
type Size struct {
    Width  int
    Height int
}

// Measurable 组件可测量接口
type Measurable interface {
    Measure(constraints Constraints) Size
}
```

### 2.3 布局节点结构

```go
// LayoutNode 布局节点
type LayoutNode struct {
    ID           string
    VNode        ui.VNode
    Parent       *LayoutNode
    Children     []*LayoutNode

    // 布局属性
    Constraints  Constraints
    MeasuredSize Size
    Box          Rect  // 最终位置和尺寸

    // 脏标记
    LayoutDirty  bool
    PaintDirty   bool

    // 缓存
    CachedConstraints Constraints
    CachedSize       Size
    CacheValid       bool
}
```

### 2.4 渲染模式抽象

```go
// RenderMode 渲染模式
type RenderMode int

const (
    RenderModeNonFiber RenderMode = iota
    RenderModeFiber
)

// Renderer 渲染器接口（双模式支持）
type Renderer interface {
    // Setup 设置渲染环境
    Setup() error

    // Render 渲染 VNode 树
    Render(vnode ui.VNode) error

    // Measure 测量 VNode
    Measure(vnode ui.VNode, constraints Constraints) Size

    // Cleanup 清理资源
    Cleanup() error

    // Mode 获取当前渲染模式
    Mode() RenderMode
}
```

---

## 三、组件测量实现

### 3.1 基础组件测量

```go
// Text 测量
func (t *Text) Measure(constraints Constraints) Size {
    textWidth := len(t.content)
    textHeight := 1

    // 如果有多行文本
    if constraints.MaxWidth < textWidth {
        // 计算换行后的高度
        lines := (textWidth + constraints.MaxWidth - 1) / constraints.MaxWidth
        textHeight = lines
        textWidth = constraints.MaxWidth
    }

    return constraints.Constrain(textWidth, textHeight)
}

// Button 测量
func (b *Button) Measure(constraints Constraints) Size {
    // 按钮格式: "[ label ]"
    contentWidth := len(b.label) + 4  // 2个空格 + 2个括号
    contentHeight := 1

    return constraints.Constrain(contentWidth, contentHeight)
}

// Input 测量
func (i *Input) Measure(constraints Constraints) Size {
    // 输入框格式: "[ content.... ]"
    placeholderWidth := len(i.placeholder)
    contentWidth := placeholderWidth + 4
    if len(i.value) > placeholderWidth {
        contentWidth = len(i.value) + 4
    }

    // 父容器可能限制宽度
    if constraints.MaxWidth < contentWidth {
        contentWidth = constraints.MaxWidth
    }

    return constraints.Constrain(contentWidth, 1)
}
```

### 3.2 容器组件测量

```go
// VStack 测量（垂直堆叠）
func (v *VStack) Measure(constraints Constraints) Size {
    totalHeight := 0
    maxWidth := 0

    for _, child := range v.children {
        // 给子组件的约束：宽度受限，高度不限
        childConstraints := Constraints{
            MinWidth: 0,
            MaxWidth: constraints.MaxWidth,
            MinHeight: 0,
            MaxHeight: constraints.MaxHeight - totalHeight,
        }

        childSize := measureChild(child, childConstraints)

        totalHeight += childSize.Height
        if childSize.Width > maxWidth {
            maxWidth = childSize.Width
        }
    }

    return constraints.Constrain(maxWidth, totalHeight)
}

// HStack 测量（水平排列）
func (h *HStack) Measure(constraints Constraints) Size {
    totalWidth := 0
    maxHeight := 0

    for _, child := range h.children {
        // 给子组件的约束：宽度不限（先测量），高度受限
        childConstraints := Constraints{
            MinWidth: 0,
            MaxWidth: Infinity,
            MinHeight: 0,
            MaxHeight: constraints.MaxHeight,
        }

        childSize := measureChild(child, childConstraints)

        totalWidth += childSize.Width
        if childSize.Height > maxHeight {
            maxHeight = childSize.Height
        }
    }

    // 如果总宽度超过约束，需要按 flex-grow 重新分配
    if totalWidth > constraints.MaxWidth {
        // TODO: 实现 flex shrink 逻辑
        totalWidth = constraints.MaxWidth
    }

    return constraints.Constrain(totalWidth, maxHeight)
}
```

### 3.3 边框装饰测量

边框不参与内容测量，只在布局后装饰：

```go
// BorderedNode 测量
func (b *BorderedNode) Measure(constraints Constraints) Size {
    if b.child == nil {
        return Size{Width: 2, Height: 2}  // 只有边框
    }

    // 内容在约束内测量
    innerConstraints := Constraints{
        MinWidth: max(0, constraints.MinWidth - 2),
        MaxWidth: max(0, constraints.MaxWidth - 2),
        MinHeight: max(0, constraints.MinHeight - 2),
        MaxHeight: max(0, constraints.MaxHeight - 2),
    }

    contentSize := measureChild(b.child, innerConstraints)

    // 加上边框宽度
    return Size{
        Width: contentSize.Width + 2,
        Height: contentSize.Height + 2,
    }
}
```

---

## 四、布局引擎实现

### 4.1 布局引擎接口

```go
// LayoutEngine 布局引擎
type LayoutEngine struct {
    root     *LayoutNode
    cache    map[string]*LayoutCacheEntry
    mode     RenderMode
}

// LayoutCacheEntry 布局缓存条目
type LayoutCacheEntry struct {
    Constraints Constraints
    Size        Size
    Valid       bool
}
```

### 4.2 布局计算

```go
// Layout 执行布局计算
func (e *LayoutEngine) Layout(constraints Constraints) {
    e.layoutNode(e.root, constraints)
}

// layoutNode 递归布局节点
func (e *LayoutEngine) layoutNode(node *LayoutNode, constraints Constraints) Size {
    // 脏标记优化
    if !node.LayoutDirty && node.CacheValid {
        return node.CachedSize
    }

    // 缓存检查
    cacheKey := node.ID + hashConstraints(constraints)
    if entry, ok := e.cache[cacheKey]; ok && entry.Valid {
        node.MeasuredSize = entry.Size
        node.CacheValid = true
        return entry.Size
    }

    // 测量阶段
    var size Size
    if measurable, ok := node.VNode.(Measurable); ok {
        size = measurable.Measure(constraints)
    } else {
        size = e.defaultMeasure(node, constraints)
    }

    node.MeasuredSize = size

    // 布局子节点
    e.layoutChildren(node, size, constraints)

    // 计算最终 Box
    node.Box = e.computeBox(node, size)

    // 更新缓存
    node.CachedSize = size
    node.CachedConstraints = constraints
    node.CacheValid = true
    node.LayoutDirty = false

    if e.cache != nil {
        e.cache[cacheKey] = &LayoutCacheEntry{
            Constraints: constraints,
            Size:        size,
            Valid:       true,
        }
    }

    return size
}

// layoutChildren 布局子节点
func (e *LayoutEngine) layoutChildren(node *LayoutNode, size Size, parentConstraints Constraints) {
    if len(node.Children) == 0 {
        return
    }

    // 根据布局类型排列子节点
    layoutInfo := ui.GetLayoutInfo(node.VNode)

    if layoutInfo.IsHorizontal {
        e.layoutHorizontal(node, size)
    } else {
        e.layoutVertical(node, size)
    }
}

// layoutVertical 垂直布局子节点
func (e *LayoutEngine) layoutVertical(node *LayoutNode, parentSize Size) {
    y := node.Box.Y

    for _, child := range node.Children {
        childConstraints := Constraints{
            MinWidth: 0,
            MaxWidth: parentSize.Width,
            MinHeight: 0,
            MaxHeight: node.Box.Height - (y - node.Box.Y),
        }

        childSize := e.layoutNode(child, childConstraints)

        child.Box.X = node.Box.X
        child.Box.Y = y
        child.Box.Width = childSize.Width
        child.Box.Height = childSize.Height

        y += childSize.Height
    }
}

// layoutHorizontal 水平布局子节点
func (e *LayoutEngine) layoutHorizontal(node *LayoutNode, parentSize Size) {
    x := node.Box.X

    for _, child := range node.Children {
        childConstraints := Constraints{
            MinWidth: 0,
            MaxWidth: node.Box.Width - (x - node.Box.X),
            MinHeight: 0,
            MaxHeight: parentSize.Height,
        }

        childSize := e.layoutNode(child, childConstraints)

        child.Box.X = x
        child.Box.Y = node.Box.Y
        child.Box.Width = childSize.Width
        child.Box.Height = childSize.Height

        x += childSize.Width
    }
}
```

### 4.3 脏标记管理

```go
// MarkLayoutDirty 标记节点及其祖先为布局脏
func (e *LayoutEngine) MarkLayoutDirty(node *LayoutNode) {
    for node != nil {
        if node.LayoutDirty {
            return  // 已经标记过
        }
        node.LayoutDirty = true
        node.CacheValid = false
        node = node.Parent
    }
}

// MarkPaintDirty 标记节点为绘制脏
func (e *LayoutEngine) MarkPaintDirty(node *LayoutNode) {
    node.PaintDirty = true
    // 绘制脏不传播给父节点
}

// ShouldLayout 判断是否需要重新布局
func (e *LayoutEngine) ShouldLayout(node *LayoutNode) bool {
    return node.LayoutDirty || !node.CacheValid
}
```

---

## 五、Fiber 与非 Fiber 双模式支持

### 5.1 渲染器工厂

```go
// NewRenderer 创建渲染器
func NewRenderer(mode RenderMode) Renderer {
    switch mode {
    case RenderModeFiber:
        return NewFiberRenderer()
    case RenderModeNonFiber:
        return NewNonFiberRenderer()
    default:
        return NewNonFiberRenderer()
    }
}

// DetectRenderMode 检测渲染模式
func DetectRenderMode() RenderMode {
    if os.Getenv("MINT_USE_FIBER") == "true" {
        return RenderModeFiber
    }
    return RenderModeNonFiber
}
```

### 5.2 Fiber 渲染器

```go
// FiberRenderer Fiber 模式渲染器
type FiberRenderer struct {
    reconciler *reconciler.Reconciler
    layout     *LayoutEngine
}

func NewFiberRenderer() *FiberRenderer {
    return &FiberRenderer{
        layout: NewLayoutEngine(RenderModeFiber),
    }
}

func (r *FiberRenderer) Setup() error {
    // 初始化 Fiber reconciler
    return nil
}

func (r *FiberRenderer) Render(vnode ui.VNode) error {
    // Fiber 模式：通过 reconciler 处理
    // reconciler 会处理组件 diff 和更新
    return nil
}

func (r *FiberRenderer) Measure(vnode ui.VNode, constraints Constraints) Size {
    // 使用布局引擎测量
    if measurable, ok := vnode.(Measurable); ok {
        return measurable.Measure(constraints)
    }
    return r.layout.Measure(vnode, constraints)
}
```

### 5.3 非 Fiber 渲染器

```go
// NonFiberRenderer 非 Fiber 模式渲染器
type NonFiberRenderer struct {
    layout     *LayoutEngine
    focusMgr   *ui.VNodeFocusManager
}

func NewNonFiberRenderer() *NonFiberRenderer {
    return &NonFiberRenderer{
        layout: NewLayoutEngine(RenderModeNonFiber),
    }
}

func (r *NonFiberRenderer) Setup() error {
    return nil
}

func (r *NonFiberRenderer) Render(vnode ui.VNode) error {
    // 非 Fiber 模式：直接遍历渲染
    return r.renderVNode(vnode, 0, 0)
}

func (r *NonFiberRenderer) Measure(vnode ui.VNode, constraints Constraints) Size {
    if measurable, ok := vnode.(Measurable); ok {
        return measurable.Measure(constraints)
    }
    return r.layout.Measure(vnode, constraints)
}
```

### 5.4 统一渲染接口

```go
// DeclarativeNode 统一接口
type DeclarativeNode struct {
    renderer Renderer
    mode     RenderMode
}

func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 统一入口，根据模式分发
    if err := n.renderer.Setup(); err != nil {
        // 处理错误
    }
    defer n.renderer.Cleanup()

    // 使用布局引擎计算布局
    constraints := LooseConstraints(ctx.Bounds.Width, ctx.Bounds.Height)
    n.renderer.Measure(n.root, constraints)

    // 渲染
    n.renderer.Render(n.root)
}
```

---

## 六、实施阶段

### 阶段 1: 核心类型定义（1-2 天）

**文件**: `internal/layout/constraints.go`, `internal/layout/size.go`, `internal/layout/measurable.go`

**任务**:
1. 定义 `Constraints` 结构和辅助方法
2. 定义 `Size` 结构
3. 定义 `Measurable` 接口
4. 编写单元测试

### 阶段 2: 组件测量实现（3-4 天）

**文件**: `runtime/ui/components_measure.go`

**任务**:
1. 为 Text 实现 `Measure`
2. 为 Button 实现 `Measure`
3. 为 Input 实现 `Measure`
4. 为 BorderedNode 实现 `Measure`
5. 编写测量测试

### 阶段 3: 容器布局实现（3-4 天）

**文件**: `internal/layout/container.go`

**任务**:
1. 实现 VStack 测量和布局
2. 实现 HStack 测量和布局
3. 实现 Table 布局
4. 支持Gap属性
5. 编写布局测试

### 阶段 4: 布局引擎实现（3-4 天）

**文件**: `internal/layout/engine.go`

**任务**:
1. 实现 `LayoutEngine` 核心逻辑
2. 实现脏标记管理
3. 实现布局缓存
4. 实现增量布局
5. 编写引擎测试

### 阶段 5: 渲染器双模式实现（2-3 天）

**文件**: `internal/render/renderer.go`, `internal/render/fiber_renderer.go`, `internal/render/nonfiber_renderer.go`

**任务**:
1. 定义 `Renderer` 接口
2. 实现 `FiberRenderer`
3. 实现 `NonFiberRenderer`
4. 实现渲染器工厂
5. 编写渲染器测试

### 阶段 6: 集成到现有系统（2-3 天）

**文件**: `internal/render/declarative_node.go`

**任务**:
1. 修改 `DeclarativeNode` 使用新的布局引擎
2. 保持向后兼容
3. 迁移现有组件到 `Measurable` 接口
4. 集成测试

### 阶段 7: 优化与测试（2-3 天）

**任务**:
1. 性能基准测试
2. 内存使用优化
3. 缓存策略优化
4. 文档更新
5. 示例更新

---

## 七、文件结构

```
internal/
├── layout/
│   ├── constraints.go       # 约束系统
│   ├── size.go              # 尺寸结构
│   ├── measurable.go        # 可测量接口
│   ├── engine.go            # 布局引擎
│   ├── container.go         # 容器布局
│   └── cache.go             # 布局缓存
├── render/
│   ├── renderer.go          # 渲染器接口
│   ├── fiber_renderer.go    # Fiber 渲染器
│   ├── nonfiber_renderer.go # 非 Fiber 渲染器
│   └── declarative_node.go  # 声明式节点（修改）
runtime/
└── ui/
    └── components_measure.go # 组件测量实现
```

---

## 八、测试策略

### 8.1 单元测试

- `Constraints` 操作测试
- 各组件 `Measure` 方法测试
- 布局引擎测试

### 8.2 集成测试

- 完整布局流程测试
- Fiber/非 Fiber 模式切换测试
- 现有示例回归测试

### 8.3 性能测试

- 布局缓存命中率测试
- 脏标记优化效果测试
- 大规模树性能测试

---

## 九、风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 破坏现有功能 | 高 | 充分的回归测试，逐步迁移 |
| 性能下降 | 中 | 基准测试，性能对比 |
| Fiber 兼容性问题 | 中 | 统一接口，充分测试 |
| 实施周期过长 | 低 | 分阶段交付，优先核心功能 |

---

## 十、成功标准

1. 所有组件实现 `Measurable` 接口
2. 布局计算正确，现有示例通过测试
3. Fiber 和非 Fiber 模式都能正常工作
4. 性能不低于现有实现
5. 脏标记优化有效（ measurable 布局减少）

---

## 附录 A: 设计文档参考

- `framework/docs/ui/idea/idea2_layout.md` - 布局系统设计
- `framework/docs/ui/idea/idea2.1_layout_detail.md` - Flex/Grid 布局细节
- `framework/docs/ui/idea/idea2.2_layout_diff.md` - 增量布局优化
- `framework/docs/ui/idea/idea2.3_layout_reconcile.md` - 增量 Reconcile

## 附录 B: 当前实现问题清单

1. `MeasureVNodeWidth/Height` 分散处理，需要统一
2. 按钮宽度计算 `len(label) + 4` 应该在 `Measure` 中
3. 边框宽度计算应该独立于内容
4. VStack/HStack 测量需要基于约束
5. 缺少布局缓存和脏标记

## 附录 C: 迁移检查清单

- [ ] 定义核心类型
- [ ] 实现组件测量
- [ ] 实现容器布局
- [ ] 实现布局引擎
- [ ] 实现双模式渲染器
- [ ] 集成到现有系统
- [ ] 所有测试通过
- [ ] 性能达标
- [ ] 文档更新完成
