# Fiber-First 优化渲染管线设计

## 目录

1. [当前系统问题分析](#当前系统问题分析)
2. [Fiber-First 优化目标](#fiber-first-优化目标)
3. [优化后的渲染流程](#优化后的渲染流程)
4. [核心架构变更](#核心架构变更)
5. [实施方案](#实施方案)
6. [迁移路径](#迁移路径)

---

## 当前系统问题分析

### 现有渲染流程

```
用户代码 → renderFn() → VNode 树
                          ↓
                  Fiber Reconciler (协调 + Diff)
                          ↓
                    Fiber 树 + VNode 树
                          ↓
                  Layout 阶段 (依赖 VNode)
                          ↓
                    ComputedBox 树
                          ↓
                  Paint 阶段 (依赖 VNode + ComputedBox)
                          ↓
                    Buffer (屏幕缓冲)
```

### 核心问题

#### 1. 双重依赖问题

```go
// ❌ 当前：Layout 和 Paint 都依赖 VNode
func Layout(vnode VNode, fiber *Fiber) ComputedBox {
    style := vnode.Props().Style  // 依赖 VNode
    // ...
}

func Paint(vnode VNode, computedBox ComputedBox, buf *Buffer) {
    text := vnode.Text()  // 依赖 VNode
    // ...
}
```

**问题**：
- VNode 生命周期不明确
- 并发渲染时可能访问过期的 VNode
- 无法支持 Suspense 和时间切片
- 内存无法及时释放

#### 2. 三层结构混乱

```
VNode (声明) + Fiber (结构) + ComputedBox (布局结果)
```

**问题**：
- ComputedBox 持有 VNode 引用
- Layout 阶段需要访问 VNode
- Paint 阶段需要访问 VNode
- 模块边界不清晰

#### 3. 渲染管线复杂

```go
// declarative_node.go:240
func (n *DeclarativeNode) Paint(ctx PaintContext, buf *Buffer) {
    // Phase 1: 获取 VNode 树
    if useFiber {
        n.root = n.renderWithFiberContext()  // 生成 VNode
    } else {
        n.root = n.nonFiberRender()          // 生成 VNode
    }
    
    // Phase 2: 渲染 VNode
    pipeline.RenderWithConstraints(n.root, ...)  // 使用 VNode
}
```

**问题**：
- VNode 在每次 Paint 时重新创建
- 无法实现真正的增量更新
- 性能浪费在 VNode 创建上

---

## Fiber-First 优化目标

### 核心原则

> **Fiber 是唯一的运行时实体**
> **VNode 只在 Reconcile 阶段存在**
> **paint.PaintableBox 是渲染单元**

### 优化目标

1. **消除 VNode 运行时依赖**
   - Layout 只读 Fiber
   - Paint 只用 paint.PaintableBox
   - VNode 在 commit 后丢弃

2. **简化渲染管线**
   - 从 3 层结构简化为 2 层
   - 减少数据转换
   - 提高性能

3. **支持并发特性**
   - 可中断渲染
   - 可时间切片
   - 可 Suspense

---

## 优化后的渲染流程

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                   Fiber-First 渲染架构                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  用户代码 → renderFn() → VNode (临时)                       │
│                           ↓                                 │
│                   Reconciler (协调)                         │
│                           ↓                                 │
│                   Fiber 树 (持久化)                         │
│                    ↓         ↓                              │
│              Instance    Instance  → paint.PaintableBox     │
│                           ↓                                 │
│                   Layout Engine                             │
│                           ↓                                 │
│              paint.PaintableBox (布局结果)                  │
│                           ↓                                 │
│                   Paint Engine                              │
│                           ↓                                 │
│                   Buffer (屏幕缓冲)                         │
└─────────────────────────────────────────────────────────────┘
```

### 详细流程图

```
┌──────────────────────────────────────────────────────────────────┐
│                         App.render()                             │
│                      (framework/app.go)                          │
├──────────────────────────────────────────────────────────────────┤
│ Phase 1: Fiber Reconciliation (协调阶段)                        │
│ ┌────────────────────────────────────────────────────────────┐  │
│ │ reconciler.Reconcile(renderFn)                             │  │
│ │     ↓                                                       │  │
│ │ 生成临时 VNode 树                                           │  │
│ │     ↓                                                       │  │
│ │ Diff + 更新 Fiber 树                                        │  │
│ │     ↓                                                       │  │
│ │ 创建/复用 paint.PaintableBox                               │  │
│ │     ↓                                                       │  │
│ │ Commit (VNode 被丢弃)                                       │  │
│ └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│ Phase 2: Layout (布局阶段)                                       │
│ ┌────────────────────────────────────────────────────────────┐  │
│ │ FiberTree.Layout(constraints)                              │  │
│ │     ↓                                                       │  │
│ │ 遍历 Fiber 树 (只读)                                        │  │
│ │     ↓                                                       │  │
│ │ 计算 paint.PaintableBox 布局                               │  │
│ │     ↓                                                       │  │
│ │ 返回 LayoutResult (paint.PaintableBox 树)                  │  │
│ └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│ Phase 3: Paint (绘制阶段)                                        │
│ ┌────────────────────────────────────────────────────────────┐  │
│ │ PaintEngine.Paint(layoutResult, buf)                       │  │
│ │     ↓                                                       │  │
│ │ 遍历 paint.PaintableBox 树                                 │  │
│ │     ↓                                                       │  │
│ │ 调用 instance.Paint(x, y)                                   │  │
│ │     ↓                                                       │  │
│ │ 生成 DrawCommand → Buffer                                   │  │
│ └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

### 核心数据流

```
┌─────────────┐
│ 用户代码     │
└──────┬──────┘
       │ renderFn()
       ▼
┌─────────────┐
│ VNode 树     │ ← 临时，Reconcile 后丢弃
└──────┬──────┘
       │ Reconcile
       ▼
┌─────────────┐
│ Fiber 树     │ ← 持久化，运行时唯一实体
└──────┬──────┘
       │ 持有
       ▼
┌─────────────────────┐
│ paint.PaintableBox  │ ← 持久化，组件实例
└──────┬──────────────┘
       │ Layout
       ▼
┌─────────────────────┐
│ LayoutResult        │ ← paint.PaintableBox + 布局信息
└──────┬──────────────┘
       │ Paint
       ▼
┌─────────────┐
│ Buffer      │ ← 屏幕缓冲
└─────────────┘
```

---


### 多层渲染架构

Fiber-First 架构通过 Layer 系统支持多层渲染：

```
┌──────────────────────────────────────────────────────────────────┐
│                    多层渲染架构                                   │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Fiber 树 (持久化)                                               │
│       ↓                                                          │
│  FiberToNodeAdapter (接口适配，位于 internal/render)             │
│       ↓                                                          │
│  runtime/layout.Engine                                           │
│       ↓                                                          │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              layout.LayeredLayoutResult                     │ │
│  ├─────────────────────────────────────────────────────────────┤ │
│  │  LayerBase     → Box1, Box2, Box3    (z-index: 0-999)      │ │
│  │  LayerDropdown → Box4                 (z-index: 1000-1999)  │ │
│  │  LayerFixed    → Box5                 (z-index: 3000-3999)  │ │
│  │  LayerModal    → Box6, Box7           (z-index: 5000-5999)  │ │
│  │  LayerTooltip  → Box8                 (z-index: 7000-7999)  │ │
│  └─────────────────────────────────────────────────────────────┘ │
│       ↓                                                          │
│  SortByZIndex() 排序                                             │
│       ↓                                                          │
│  PaintMultiLayer (从低到高绘制)                                   │
│       ↓                                                          │
│  Buffer (屏幕缓冲)                                               │
└──────────────────────────────────────────────────────────────────┘
```

#### Layer 类型定义

| Layer | 用途 | Z-Index 范围 |
|-------|------|-------------|
| LayerBase | 默认层，普通内容 | 0-999 |
| LayerDropdown | 下拉菜单 | 1000-1999 |
| LayerSticky | 粘性定位 | 2000-2999 |
| LayerFixed | 固定定位 | 3000-3999 |
| LayerModalBackdrop | 模态背景 | 4000-4999 |
| LayerModal | 模态对话框 | 5000-5999 |
| LayerPopover | 弹出框 | 6000-6999 |
| LayerTooltip | 工具提示 | 7000-7999 |

#### 单层 vs 多层渲染

```go
// 单层渲染 (所有内容在 LayerBase):
layoutResult := layoutEngine.Layout(node, constraints)
paintEngine.PaintSingleLayer(layoutResult, buf)

// 多层渲染 (包含 Modal、Dropdown 等):
layoutResult := layoutEngine.Layout(node, constraints)
layeredResult := layoutEngine.BuildLayeredResult(layoutResult)
paintEngine.PaintMultiLayer(layeredResult, buf)
```

### 与 runtime/layout 接口集成

> **架构约束**：`runtime/layout` 不依赖 Fiber/VNode，只定义抽象接口。
> 
> 适配器 `FiberToNodeAdapter` 位于 `internal/render/fiber_adapter.go`。

Fiber 通过适配器模式与 runtime/layout 解耦集成：

```
┌─────────────────────────────────────────────────────────────────┐
│                    接口适配架构                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Fiber (运行时实体)                          │   │
│  │  - Style (布局输入)                                      │   │
│  │  - Instance (paint.PaintableBox)                        │   │
│  │  - Props, State                                         │   │
│  └────────────────────────┬────────────────────────────────┘   │
│                           │                                     │
│                           ↓                                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │     FiberToNodeAdapter (internal/render/fiber_adapter.go)│   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  实现 layout.Node      - ID(), Type(), Children()   │ │   │
│  │  │  实现 layout.Layered   - GetLayer(), GetZIndex()    │ │   │
│  │  │  实现 layout.Measurable - Measure(constraints)      │ │   │
│  │  │  实现 layout.Dirtyable  - IsLayoutDirty()           │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  └────────────────────────┬────────────────────────────────┘   │
│                           │                                     │
│                           ↓                                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │          runtime/layout.Engine                           │   │
│  │  - Layout(node, constraints) → LayoutResult             │   │
│  │  - 支持增量布局 (Dirtyable)                              │   │
│  │  - 支持多层布局 (Layered)                                │   │
│  │  - ❌ 不依赖 Fiber/VNode                                 │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

> **注意**：`FiberToNodeAdapter` 已在 `internal/render/fiber_adapter.go` 中实现，
> 不在 `runtime/layout/` 目录中。这确保了 `runtime/layout` 保持为纯布局引擎。

#### 核心接口映射

| Fiber 字段 | layout 接口 | 说明 |
|-----------|------------|------|
| fiber.ActionTargetID / fiber.Key | Node.ID() | 节点唯一标识 |
| fiber.Style.Layer | Layered.GetLayer() | 渲染层 |
| fiber.Style.ZIndex | Layered.GetZIndex() | 层内排序 |
| fiber.Instance.GetSize() | Measurable.Measure() | 尺寸计算 |
| fiber.Flags & FlagLayoutDirty | Dirtyable.IsLayoutDirty() | 增量布局 |

---

## 核心架构变更

### 1. Fiber 结构变更

#### 旧结构
```go
type Fiber struct {
    Type      ElementType
    Key       string
    VNode     VNode           // ❌ 持有 VNode 引用
    LayoutBox ComputedBox     // ❌ 布局结果混入 Fiber
    // ...
}
```

#### 新结构
```go
type Fiber struct {
    // Identity
    Type ElementType
    Key  string
    
    // Tree Structure
    Parent    *Fiber
    Child     *Fiber
    Sibling   *Fiber
    Alternate *Fiber
    
    // Runtime Entity
    Instance  paint.PaintableBox  // ✅ 持有运行时实例
    
    // Layout Input (只读)
    Style     Style               // ✅ 从 VNode 提取的样式
    Props     MemoizedProps       // ✅ 从 VNode 提取的属性
    
    // Scheduling
    Flags     Flags
    Lanes     Lane
    
    // ❌ 不再持有 VNode
    // ❌ 不再持有 LayoutBox
}
```

### 2. Layout 阶段变更

#### 旧实现
```go
// ❌ 依赖 VNode
func (e *LayoutEngine) Layout(vnode VNode, fiber *Fiber) ComputedBox {
    style := vnode.Props().Style
    children := vnode.Children()
    // ...
}
```

#### 新实现
```go
// ✅ 纯 Fiber 驱动
func (e *LayoutEngine) LayoutFiber(root *Fiber, constraints Constraints) *LayoutResult {
    return e.layoutFiberNode(root, constraints)
}

func (e *LayoutEngine) layoutFiberNode(fiber *Fiber, constraints Constraints) *LayoutNode {
    // 1. 从 Fiber 读取样式 (只读)
    style := fiber.Style
    
    // 2. 布局当前节点
    layoutNode := &LayoutNode{
        Fiber:    fiber,
        Instance: fiber.Instance,  // paint.PaintableBox
        Box:      e.calculateBox(style, constraints),
    }
    
    // 3. 递归布局子节点
    for child := fiber.Child; child != nil; child = child.Sibling {
        childConstraints := e.getChildConstraints(layoutNode.Box, child.Style)
        childLayoutNode := e.layoutFiberNode(child, childConstraints)
        layoutNode.Children = append(layoutNode.Children, childLayoutNode)
    }
    
    return layoutNode
}
```

### 3. Paint 阶段变更

#### 旧实现
```go
// ❌ 依赖 VNode 和 ComputedBox
func (e *PaintEngine) Paint(vnode VNode, computedBox ComputedBox, buf *Buffer) {
    switch vnode.Type() {
    case VNodeText:
        text := vnode.Text()
        buf.DrawText(text, computedBox.X, computedBox.Y)
    case VNodeElement:
        // ...
    }
}
```

#### 新实现
```go
// ✅ 纯 paint.PaintableBox 驱动
func (e *PaintEngine) PaintLayout(layoutResult *LayoutResult, buf *paint.Buffer) {
    e.paintLayoutNode(layoutResult.Root, buf)
}

func (e *PaintEngine) paintLayoutNode(node *LayoutNode, buf *paint.Buffer) {
    if node.Instance == nil {
        // 递归子节点
        for _, child := range node.Children {
            e.paintLayoutNode(child, buf)
        }
        return
    }
    
    // 调用 paint.PaintableBox.Paint()
    drawCommands := node.Instance.Paint(node.Box.X, node.Box.Y)
    
    // 执行绘制命令
    for _, cmd := range drawCommands {
        e.executeDrawCommand(cmd, buf)
    }
    
    // 递归绘制子节点
    for _, child := range node.Children {
        e.paintLayoutNode(child, buf)
    }
}
```

### 4. 渲染管线变更

#### 旧管线
```go
func (n *DeclarativeNode) Paint(ctx PaintContext, buf *Buffer) {
    // Phase 1: 创建 VNode 树
    n.root = n.renderFn()
    
    // Phase 2: 使用 VNode 渲染
    pipeline.RenderWithConstraints(n.root, ...)
}
```

#### 新管线
```go
func (n *DeclarativeNode) Paint(ctx PaintContext, buf *Buffer) {
    // Phase 1: Fiber Reconciliation (VNode 临时存在)
    n.reconciler.Reconcile(n.renderFn)
    // VNode 在这里被丢弃
    
    // Phase 2: Fiber-based Layout
    layoutResult := n.layoutEngine.LayoutFiber(n.fiberRoot, ctx.Constraints)
    
    // Phase 3: Paint PaintableBox
    n.paintEngine.PaintLayout(layoutResult, buf)
}
```

---

## 实施方案

### Phase 1: Fiber 结构优化 (2-3天)

#### 任务
1. **删除 Fiber.VNode 字段**
   ```go
   // 删除
   VNode VNode
   ```

2. **添加 Fiber.Instance 字段**
   ```go
   Instance paint.PaintableBox
   ```

3. **添加 Fiber.Style 字段**
   ```go
   Style Style
   ```

#### 文件修改
- `runtime/ui/fiber.go`
- `runtime/ui/fiber_util.go`

#### 验证
- [ ] Fiber 不再持有 VNode 引用
- [ ] Instance 在 CloneFiber 时被复用
- [ ] Style 在 completeWork 中被填充

### Phase 2: Layout 引擎优化 (3-5天)

#### 任务
1. **创建 Fiber-based Layout 接口**
   ```go
   type FiberLayoutEngine interface {
       LayoutFiber(root *Fiber, constraints Constraints) *LayoutResult
   }
   ```

2. **实现 LayoutFiber 方法**
   - 替换 `Layout(vnode VNode, ...)`
   - 只读 Fiber 树
   - 返回 LayoutResult (paint.PaintableBox + Box)

3. **创建 LayoutNode 结构**
   ```go
   type LayoutNode struct {
       Fiber    *Fiber
       Instance paint.PaintableBox
       Box      Box
       Children []*LayoutNode
   }
   ```

#### 文件修改
- `runtime/layout/layout_engine.go`
- `runtime/layout/layout_result.go`
- `internal/render/layout_switcher.go`

#### 验证
- [ ] Layout 不访问 VNode
- [ ] Layout 只读 Fiber
- [ ] LayoutResult 包含 paint.PaintableBox

### Phase 3: Paint 引擎优化 (3-5天)

#### 任务
1. **创建 PaintLayout 接口**
   ```go
   type PaintEngine interface {
       PaintLayout(layoutResult *LayoutResult, buf *paint.Buffer)
   }
   ```

2. **实现 PaintLayout 方法**
   - 调用 `instance.Paint(x, y)`
   - 执行 DrawCommand

3. **优化 paint.PaintableBox 接口**
   ```go
   type PaintableBox interface {
       Paint(x, y int) []paint.DrawCmd
   }
   ```

#### 文件修改
- `runtime/paint/renderer.go`
- `internal/render/paint_engine.go`

#### 验证
- [ ] Paint 不访问 VNode
- [ ] Paint 只用 paint.PaintableBox
- [ ] 所有组件实现 PaintableBox

### Phase 4: 渲染管线集成 (5-7天)

#### 任务
1. **优化 DeclarativeNode.Paint()**
   ```go
   func (n *DeclarativeNode) Paint(ctx PaintContext, buf *Buffer) {
       // 1. Reconcile (生成 VNode → 更新 Fiber → 丢弃 VNode)
       n.reconciler.Reconcile(n.renderFn)
       
       // 2. Layout (Fiber → LayoutResult)
       layoutResult := n.layoutEngine.LayoutFiber(n.fiberRoot, ctx.Constraints)
       
       // 3. Paint (LayoutResult → Buffer)
       n.paintEngine.PaintLayout(layoutResult, buf)
   }
   ```

2. **删除 Legacy 渲染路径**
   - 删除 `PaintVNode()`
   - 删除 `nonFiberRender()`

3. **优化 RenderingPipeline**
   - 简化 Layer 处理
   - 移除 VNode 依赖

#### 文件修改
- `internal/render/declarative_node.go`
- `internal/render/rendering_pipeline.go`
- `internal/render/pipeline_renderer.go`

#### 验证
- [ ] 渲染管线不依赖 VNode
- [ ] 所有组件正常渲染
- [ ] 性能提升明显

### Phase 5: 组件迁移 (7-10天)

#### 任务
1. **迁移基础组件**
   - Text
   - VStack
   - HStack
   - Spacer

2. **迁移交互组件**
   - Button
   - Input
   - Checkbox

3. **迁移高级组件**
   - Table
   - Modal
   - List

#### 文件修改
- `components/*/vnode.go`
- `components/*/instance.go`

#### 验证
- [ ] 所有组件实现 paint.PaintableBox
- [ ] 所有组件测试通过
- [ ] 示例应用正常运行

---

## 迁移路径

### 向后兼容策略

#### 阶段 1: 双轨运行 (1-2周)
```go
func (n *DeclarativeNode) Paint(ctx PaintContext, buf *Buffer) {
    if n.useFiberFirst {
        // 新渲染路径
        n.fiberFirstPaint(ctx, buf)
    } else {
        // 旧渲染路径 (兼容)
        n.legacyPaint(ctx, buf)
    }
}
```

#### 阶段 2: 逐步迁移 (2-4周)
- 先迁移简单组件
- 再迁移复杂组件
- 最后删除旧路径

#### 阶段 3: 完全切换 (1周)
- 删除 Legacy 代码
- 删除 VNode 运行时依赖
- 清理代码

### 性能对比指标

| 指标 | 旧架构 | 新架构 | 提升 |
|------|--------|--------|------|
| 内存占用 | VNode + Fiber + ComputedBox | Fiber + paint.PaintableBox | ~30% ↓ |
| 渲染时间 | VNode 创建 + Layout + Paint | Layout + Paint | ~20% ↓ |
| GC 压力 | 每帧创建 VNode | 仅更新 Fiber | ~40% ↓ |
| 并发能力 | ❌ | ✅ | ∞ |

### 风险控制

#### 高风险点
1. **组件迁移** - 需要所有组件实现 paint.PaintableBox
2. **布局兼容性** - 确保新 Layout 引擎结果一致
3. **性能回退** - 密切监控性能指标

#### 缓解措施
1. **渐进式迁移** - 保持双轨运行
2. **充分测试** - 每个阶段都有完整测试
3. **性能监控** - 实时监控关键指标
4. **回滚机制** - 随时可以切换回旧路径

---

## 成功标准

### 技术标准
- [ ] VNode 在 commit 后被完全丢弃
- [ ] Layout 只读 Fiber
- [ ] Paint 只用 paint.PaintableBox
- [ ] 所有组件实现 paint.PaintableBox
- [ ] 性能提升 > 15%

### 功能标准
- [ ] 所有现有功能正常
- [ ] 所有测试通过
- [ ] 示例应用正常运行
- [ ] 无内存泄漏

### 架构标准
- [ ] 模块边界清晰
- [ ] 无循环依赖
- [ ] 代码可维护性提高
- [ ] 支持未来并发特性

---


## 组件模板参考

### ui/components/button 目录结构

Button 组件是 Fiber-first 架构的完整参考实现：

```
ui/components/button/
├── vnode.go        # VNode 描述（纯声明）
├── instance.go     # Instance 运行期实体（状态 + 渲染）
├── builder.go      # Builder 流式 API
├── button_test.go  # 单元测试
└── README.md       # 组件文档
```

### 核心设计模式

#### 1. VNode = 纯描述

```go
type VNode struct {
    // Props only - 无状态、无闭包、无 Paint
    label       string
    variant     Variant
    size        Size
    pressIntent intent.Intent  // 替代 func()
}
```

#### 2. Instance = 运行期实体

```go
type Instance struct {
    // Props from VNode
    label   string
    
    // Runtime State
    state  control.InteractionState
    
    // Behaviors
    behaviors *control.BehaviorList
}
```

#### 3. Behavior 组合

```go
inst.behaviors = control.NewBehaviorList(
    &control.FocusableBehavior{},
    &control.PressableBehavior{},
    &control.HoverableBehavior{},
    &control.DisableableBehavior{},
)
```

#### 4. Intent 替代闭包

```go
// ❌ 旧方式
button.OnClick(func() { showModal() })

// ✅ 新方式
button.B("Open").OnPress(intent.OpenModal("settings")).Build()
```

### 组件迁移目标

所有组件应迁移到 `ui/components/` 目录：

| 组件 | 状态 | 说明 |
|------|------|------|
| button | ✅ 完成 | 参考模板 |
| text | 待迁移 | 基础组件 |
| stack | 待迁移 | VStack/HStack |
| input | 待迁移 | 交互组件 |
| table | 待迁移 | 复杂组件 |

---

## 总结

Fiber-First 优化后的渲染管线将：

1. **消除 VNode 运行时依赖** - VNode 只在 Reconcile 阶段存在
2. **简化渲染流程** - 从 3 层简化为 2 层
3. **提高性能** - 减少内存占用和 GC 压力
4. **支持并发** - 为未来并发特性打下基础

这是一次架构级的优化，将使系统更加清晰、高效、可扩展。

---

**下一步**: 开始实施 [Phase 1: Fiber 结构优化](#phase-1-fiber-结构优化-2-3天)
