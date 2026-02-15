# 正确的渲染流程梳理

## 核心原则

**VNode 与 Fiber 必须完全分离**，不混用：
- VNode：临时表达式，只存在于 render 阶段
- Fiber：持久化树，是 Layout 和 Event 的唯一数据源

## 当前问题

当前 `Engine.Layout(vnode, fiber, constraints)` 混用了 VNode 和 Fiber：
```go
func (e *Engine) Layout(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    // 混用：同时接收 vnode 和 fiber
    root := e.buildComputedBox(vnode, fiber, nil, constraints)  // ← 问题
    ...
}
```

`buildComputedBox` 内部调用 `vnode.Children()`, `vnode.Key()` 等，违反了 fiber-first 原则。

## 正确的双路径架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Rendering Pipeline                   │
└─────────────────────────────────────────────────────────────┘
                          │
         ┌────────────────┴────────────────┐
         │                              │
    ┌────▼─────┐              ┌────▼─────┐
    │ VNode Path │              │ Fiber Path │
    └────┬─────┘              └────┬─────┘
         │                              │
    ┌────▼─────────┐           ┌────▼─────────┐
    │ buildComputedBox │           │ buildComputedBoxFromFiber │
    │ (DEPRECATED)     │           │ (PURE FIBER)    │
    └────┬─────────┘           └────┬─────────┘
         │                              │
         └──────────┬───────────────────┘
                    │
              ┌────▼─────────┐
              │ ComputedBox    │
              └────┬─────────┘
                   │
                ┌────▼─────┐
                │ PaintEngine │
                └────┬─────┘
                     │
                   ┌─▼─┐
                   │Buffer│
                   └────┘
```

## VNode 路径（临时兼容）

### 入口：`Layout(vnode, fiber, constraints)`

```go
func (e *Engine) Layout(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    // 临时兼容：fiber != nil 时使用 VNode 路径
    if fiber == nil {
        return e.layoutVNode(vnode, constraints)
    }
    // Phase 6: Fiber 路径
    return e.LayoutFiber(fiber, constraints)
}
```

### VNode 路径实现（仅用于非 Fiber 模式）

```go
// layoutVNode - 完全独立的 VNode 布局（不访问 Fiber）
func (e *Engine) layoutVNode(vnode VNode, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    root := e.buildComputedBoxFromVNode(vnode, nil, constraints)
    e.calculatePositionsVNode(root, 0, 0)
    root.ClearDirty()
    hitMap := e.buildHitMapFromComputedBoxesVNode(root)
    layout := NewComputedLayout(root)
    layout.HitMap = hitMap
    return layout, nil
}

// buildComputedBoxFromVNode - 只使用 VNode
func (e *Engine) buildComputedBoxFromVNode(vnode VNode, parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox {
    // 只调用 vnode.Method()，不访问 Fiber
    // ... 实现细节
}
```

## Fiber 路径（最终目标）

### 入口：`LayoutFiber(fiber, constraints)`

```go
// LayoutFiber - 纯 Fiber 布局（不访问 VNode）
func (e *Engine) LayoutFiber(root *rtui.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    e.flexCache = make(map[string]*FlexDistributionInfo)
    root := e.buildComputedBoxFromFiber(root, nil, constraints)
    e.calculatePositionsFiber(root, 0, 0)
    root.ClearDirty()
    hitMap := e.buildHitMapFromFibers(root)
    layout := NewComputedLayout(root)
    layout.HitMap = hitMap
    return layout, nil
}
```

### Fiber 路径实现

```go
// buildComputedBoxFromFiber - 只使用 Fiber
func (e *Engine) buildComputedBoxFromFiber(fiber *rtui.Fiber, parent *ComputedBox, constraints runtime.BoxConstraints) *ComputedBox {
    // 只调用 fiber.Method()，不访问 VNode
    box := &ComputedBox{
        Parent: parent,
        NodeID: fiber.NodeID,
        Layer:  fiber.Layer,
        Box:    runtime.Box{X: 0, Y: 0, Width: 0, Height: 0},
    }

    // 测量
    measurement := e.measureFiberLayout(fiber, constraints)
    box.Box.Width = measurement.Size.Width
    box.Box.Height = measurement.Size.Height

    // 构建子节点（只遍历 Fiber.Child → Sibling）
    children := e.getChildFibers(fiber)
    box.Children = make([]*ComputedBox, 0, len(children))

    for i, childFiber := range children {
        childBox := e.buildComputedBoxFromFiber(childFiber, box, measurement.ChildConstraints[i])
        if childBox != nil {
            box.Children = append(box.Children, childBox)
        }
    }

    return box
}

// measureFiberLayout - 从 Fiber 测量
func (e *Engine) measureFiberLayout(fiber *rtui.Fiber, constraints runtime.BoxConstraints) runtime.LayoutMeasurement {
    // 检查 fiber 是否实现了 LayoutMeasurer
    if !fiber.IsLayoutMeasurer() {
        return runtime.LayoutMeasurement{}
    }

    // 调用 fiber.MeasureLayout(measurer, constraints)
    return fiber.MeasureLayout(e, constraints)
}

// calculatePositionsFiber - 使用 Fiber 计算
func (e *Engine) calculatePositionsFiber(box *ComputedBox, x, y int) {
    box.Box.X = x
    box.Box.Y = y

    // 使用 fiber.Tag() 而不是 vnode.Tag()
    tag := box.GetFiber().Tag

    switch tag {
    case "hstack":
        e.layoutHStackFiber(box, x, y)
    case "vstack":
        e.layoutVStackFiber(box, x, y)
    case "bordered":
        e.layoutBorderedFiber(box, x, y)
    default:
        e.layoutDefaultFiber(box, x, y)
    }
}

// getChildFibers - 遍历 Fiber.Child → Sibling
func (e *Engine) getChildFibers(fiber *rtui.Fiber) []*rtui.Fiber {
    var children []*rtui.Fiber
    for child := fiber.Child; child != nil; child = child.Sibling {
        children = append(children, child)
    }
    return children
}
```

## Fiber 必须实现的方法

```go
// 在 rtui.Fiber 上添加：
func (f *Fiber) IsLayoutMeasurer() bool  // 检查是否可布局
func (f *Fiber) MeasureLayout(measurer ChildMeasurer, constraints BoxConstraints) LayoutMeasurement
func (f *Fiber) GetChildFibers() []*Fiber  // 获取子节点列表
func (f *Fiber) GetDirection() Direction  // 布局方向
func (f *Fiber) GetGap() int            // 间距
func (f *Fiber) GetFlex() int           // flex 值
func (f *Fiber) GetAlign() Align         // 主轴对齐
func (f *Fiber) GetCrossAlign() Align    // 交叉轴对齐
func (f *Fiber) GetPadding() [4]int    // padding
func (f *Fiber) Tag() string            // 节点类型
```

## 实施步骤

### 阶段 1：创建 Fiber 独立方法

```bash
# 1. 在 runtime/compute/engine.go 创建纯 Fiber 方法
- LayoutFiber(fiber, constraints) (*ComputedLayout, error)
- buildComputedBoxFromFiber(fiber, parent, constraints) *ComputedBox
- calculatePositionsFiber(box, x, y)
- measureFiberLayout(fiber, constraints) LayoutMeasurement
- getChildFibers(fiber) []*Fiber

# 2. 修改 Layout() 入口进行路由
func (e *Engine) Layout(vnode VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*ComputedLayout, error) {
    if fiber != nil {
        return e.LayoutFiber(fiber, constraints)
    }
    return e.layoutVNode(vnode, constraints)  // 或报错不支持
}
```

### 阶段 2：在 Fiber 上实现 Layout 方法

```bash
# 在 runtime/reconciler/fiber.go 添加：
func (f *Fiber) IsLayoutMeasurer() bool {
    // 检查 fiber.Type 是否需要布局
    switch f.Type {
    case VNodeElement, VNodeComponent:
        return true
    default:
        return false
    }
}
```

### 阶段 3：Bordered 组件 Fiber 化

```bash
# Bordered 需要在 Fiber 上实现：
1. IsLayoutMeasurer() -> true
2. MeasureLayout() -> 完整的布局逻辑

# 当前 Bordered 的 MeasureLayout 在 VNode (runtime/ui/layout.go)
# 需要迁移到 Fiber (runtime/reconciler/fiber.go 或独立的 layout 文件)
```

## 当前架构问题

```
buildComputedBox(vnode, fiber, parent, constraints)
    ├─ 使用 vnode.Children()  ❌ 违反 fiber-first
    ├─ 使用 vnode.Key()      ❌ 违反 fiber-first
    ├─ 使用 vnode.Props()    ❌ 违反 fiber-first
    └─ 混用 fiber 参数     ❌ 不清晰
```

## 目标架构

```
buildComputedBoxFromFiber(fiber, parent, constraints)
    ├─ 使用 fiber.Child → Sibling  ✅ 纯 Fiber
    ├─ 使用 fiber.MeasureLayout()   ✅ 纯 Fiber
    ├─ 使用 fiber.Tag()          ✅ 纯 Fiber
    └─ 不访问 VNode             ✅ 完全分离
```

## 迁移优先级

1. **高优先级**：创建独立的 Fiber 路径方法
   - `LayoutFiber()`
   - `buildComputedBoxFromFiber()`
   - `calculatePositionsFiber()`

2. **中优先级**：在 Fiber 上添加 Layout 方法
   - `IsLayoutMeasurer()`
   - `MeasureLayout()`

3. **低优先级**：删除旧的 VNode 路径
   - 废弃 `buildComputedBox()`
   - 移除 `vnode.Children()` 调用
