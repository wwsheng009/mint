# Mint TUI Layer 系统完整实施指南

> 基于 `docs/layer/layer_design.md` 设计意图的完整实施计划
>
> **目标**: 实现符合设计原则的单树 + Layer 属性架构，支持 Modal 居中、AbsX/AbsY、Position:Fixed 和 Portal 系统

---

## 目录

1. [设计原则回顾](#设计原则回顾)
2. [现状分析](#现状分析)
3. [实施路线图](#实施路线图)
4. [Phase 1: 快速修复 (1-2天)](#phase-1-快速修复-1-2天)
5. [Phase 2: 中期优化 (3-5天)](#phase-2-中期优化-3-5天)
6. [Phase 3: 长期特性 (1-2周)](#phase-3-长期特性-1-2周)
7. [验收标准](#验收标准)
8. [风险与缓解](#风险与缓解)

---

## 设计原则回顾

根据 `layer_design.md`，正确架构必须满足：

| 原则 | 说明 | 状态 |
|-----|------|------|
| ✅ 一棵树结构 | 所有层的节点在同一棵布局树中 | ✅ 已实现 |
| ✅ 父→子坐标累加 | 每个节点只存相对坐标，全局坐标通过累积计算 | ⚠️ 部分实现 |
| ✅ Z 不参与布局 | Layer/ZIndex 仅用于渲染排序和事件命中 | ✅ 已实现 |
| ✅ 渲染前 Flatten | 将树扁平化后按 Z 排序绘制 | ✅ 已实现 |
| ✅ 事件从 top → down | 按 Z 逆序命中和路由事件 | ✅ 已实现 |
| ✅ Modal 使用 PositionFixed | Modal 应脱离父布局流，以 Root 为参考系 | ❌ 未实现 |
| ✅ Layout 阶段预计算 AbsX | 在布局阶段计算全局坐标，避免反复递归 | ❌ 未实现 |

---

## 现状分析

### ✅ 已实现的正确架构

1. **单树共享布局**
   - 位置: `declarative_node.go:444`
   - 注释: `// 方案A: 单树共享布局 - 所有layer在一个布局树中计算`
   - 所有层的节点在同一个 `LayoutBox` 树中

2. **Layer 属性传播链**
   ```
   Fiber.Layer → LayoutBox.Layer → PaintableBox.Layer → PaintablePlanes
   ```

3. **分层绘制**
   - `PaintablePlanes` 按层组织 `PaintableBox`
   - 绘制引擎按 Z-order 从高到低绘制

### ⚠️ 需要修复的问题

| 问题 | 当前实现 | 应该的实现 | 优先级 |
|-----|---------|-----------|--------|
| Modal 居中 | `rendering_pipeline.go:applyLayerTransformsToPaintable()` (Paint 阶段) | `types.go:layoutNodeWithDepth()` (Layout 阶段) | 🔴 高 |
| AbsX/AbsY 缺失 | 无字段，需要向上遍历计算 | 在布局阶段预计算 | 🔴 高 |

### ❌ 计划中的功能

| 功能 | 设计要求 | 当前状态 | 优先级 |
|-----|---------|---------|--------|
| Position:Fixed | Fiber 需要字段，布局引擎逻辑支持 | 类型已定义但未实现 | 🟠 中 |
| Anchor 系统 | 支持 9 种锚点类型 | 类型已定义但未实现 | 🟠 中 |
| Portal 系统 | Fiber.PortalRoot + OverlayManager | 未实现 | 🟢 低 |

---

## 实施路线图

```
Phase 1: 快速修复 (1-2天)
├── 1.1 将 Modal 居中从 Paint 阶段移到 Layout 阶段
├── 1.2 在 LayoutBox 中添加 AbsX/AbsY 字段
└── 1.3 在 layoutNodeWithDepth 中预计算全局坐标

Phase 2: 中期优化 (3-5天)
├── 2.1 在 Fiber 中添加 Position 和 Anchor 字段
├── 2.2 在 completeWork 中填充 Position/Anchor
├── 2.3 在布局引擎中实现 PositionFixed 处理
└── 2.4 实现 Anchor 居中计算逻辑

Phase 3: 长期特性 (1-2周) [可选]
├── 3.1 在 Fiber 中添加 PortalRoot 字段
├── 3.2 实现 OverlayManager 管理浮层
├── 3.3 实现 Portal 跨树挂载
└── 3.4 集成 Focus Trap for Modal
```

---

## Phase 1: 快速修复 (1-2天)

### 1.1 将 Modal 居中从 Paint 阶段移到 Layout 阶段

#### 问题分析

**当前错误实现**: `internal/render/rendering_pipeline.go`

```go
// ❌ 错误: Modal 居中逻辑在 Paint 阶段
func applyLayerTransformsToPaintable(pbox *paint.PaintableBox, constraints runtime.BoxConstraints) {
    if pbox.Layer == int(runtime.LayerModal) {
        // 在 Paint 阶段计算居中
        if pbox.X == 0 && pbox.Y == 0 {
            pbox.X = (constraints.MaxWidth - pbox.Width) / 2
            pbox.Y = (constraints.MaxHeight - pbox.Height) / 2
        }
    }
}
```

**问题**:
- 违反了设计原则：计算应该在 Layout 阶段完成
- 只在 Paint 时才知道约束太晚
- 无法复用于事件命中

#### 正确实现方案

##### Step 1: 在 LayoutBox 中添加中心化标记

```go
// runtime/layout/types.go
type LayoutBox struct {
    ID       string
    X, Y     int           // 相对坐标
    AbsX, AbsY int         // ✨ 全局坐标 (Phase 1.2 添加)
    Width, Height int
    Layer    Layer
    ZIndex   int
    Border   Border

    // ✨ 添加：是否需要居中（用于 Modal）
    // 设置方法: 检测到 Modal 且未设置明确位置时为 true
    ShouldCenter bool        // Phase 1.1 添加

    Children []*LayoutBox
}
```

##### Step 2: 在布局引擎中实现居中逻辑

**文件**: `runtime/layout/types.go`

```go
// layoutNodeWithDepth 中添加居中检查
func (e *Engine) layoutNodeWithDepth(
    node Node,
    constraints Constraints,
    x, y int,
    depth int,
    visited map[string]bool,
) *LayoutBox {
    if node == nil {
        return nil
    }

    // ... 现有代码 ...

    // 获取节点尺寸
    width, height := node.GetSize()
    if measurable, ok := node.(Measurable); ok {
        size := measurable.Measure(constraints)
        width, height = size.Width, size.Height
    }

    // ✨ Phase 1.1: Modal 居中逻辑
    layer := GetLayerFromNode(node)
    zIndex := GetZIndexFromNode(node)

    // 检测是否需要居中
    shouldCenter := false
    if layer == LayerModal && width > 0 && height > 0 {
        // 检查是否有明确的定位 (通过 AbsoluteStyleProvider)
        if absProvider, ok := node.(AbsoluteStyleProvider); ok {
            absStyle := absProvider.GetAbsoluteStyle()
            // 如果未设置明确的 left/top/right/bottom，则需要居中
            shouldCenter = absStyle.ShouldCenter()
        } else {
            // 默认 Modal 居中
            shouldCenter = true
        }
    }

    // 计算居中坐标
    if shouldCenter {
        x = (constraints.MaxWidth - width) / 2
        y = (constraints.MaxHeight - height) / 2
    }

    box := &LayoutBox{
        ID:       node.ID(),
        X:        x,
        Y:        y,
        Width:    width,
        Height:   height,
        Layer:    layer,
        ZIndex:   zIndex,
        Border:   nodeBorder,
        ShouldCenter: shouldCenter,  // ✨ 保存标记
        Children: make([]*LayoutBox, 0),
    }

    // ... 后续代码（布局子节点） ...

    return box
}
```

##### Step 3: 修改 AbsoluteStyle 接口

**文件**: `runtime/layout/position.go`

```go
// 在 AbsoluteStyle 接口中添加 ShouldCenter 方法
type AbsoluteStyle interface {
    // Left 相对容器左边的距离（0~1 表示百分比）
    Left() float64
    Right() float64
    Top() float64
    Bottom() float64
    Width() int
    Height() int
    Anchor() Anchor

    // ✨ Phase 1.1: 判断是否需要居中
    ShouldCenter() bool
}

// Concrete implementation
type AbsoluteStyleConcrete struct {
    left, right float64
    top, bottom float64
    width, height int
    anchor Anchor
}

func (s *AbsoluteStyleConcrete) ShouldCenter() bool {
    // 如果所有定位属性都是默认值（0），则需要居中
    return s.Left() == 0 && s.Right() == 0 && s.Top() == 0 && s.Bottom() == 0 &&
           s.Anchor() == AnchorTopLeft &&
           s.Width() == 0 && s.Height() == 0
}
```

##### Step 4: 移除 Paint 阶段的居中逻辑

**文件**: `internal/render/rendering_pipeline.go`

```go
// ❌ 注释掉或删除 applyLayerTransformsToPaintable 中的居中逻辑
// 在 Phase 1.1 完成后，这个函数不再需要调用

func (e *PaintEngine) PaintPaintablePlanes(planes *PaintablePlanes, buf *Buffer) error {
    // 从高 layer 到低 layer 绘制
    for layer := LayerMax; layer >= LayerBase; layer-- {
        boxes := planes.GetLayer(layer)

        // 直接绘制，不再应用 transform
        for _, box := range boxes {
            if err := e.paintBox(box, buf); err != nil {
                return err
            }
        }
    }

    return nil
}
```

#### 验收标准

- [ ] Modal 默认显示在屏幕中央
- [ ] Modal 居中逻辑在 Layout 阶段完成（查看日志 `layoutNodeWithDepth`）
- [ ] Paint 阶段不再修改坐标
- [ ] 测试用例覆盖：
  - Modal 在居中位置
  - Modal 在指定位置（使用 left/top）

---

### 1.2 在 LayoutBox 中添加 AbsX/AbsY 字段

#### 实现方案

**文件**: `runtime/layout/types.go`

```go
type LayoutBox struct {
    ID       string
    X, Y     int           // 相对坐标（相对于父节点）
    AbsX, AbsY int         // ✨ 全局坐标（相对于屏幕/Root，Phase 1.2 添加）
    Width, Height int
    Baseline int
    Layer    Layer
    ZIndex   int
    Border   Border
    ShouldCenter bool      // Phase 1.1 添加

    Children []*LayoutBox
}
```

#### 同时在 PaintableBox 中添加

**文件**: `runtime/paint/types.go`（如果需要）

```go
type PaintableBox struct {
    X, Y, Width, Height int
    AbsX, AbsY int      // ✨ 全局坐标（可选，如果事件命中需要）
    Layer, ZIndex int
    ...
}
```

---

### 1.3 在布局阶段预计算 AbsX/AbsY

#### 实现方案

**文件**: `runtime/layout/types.go`

```go
// layoutNodeWithDepth 中预计算全局坐标
func (e *Engine) layoutNodeWithDepth(
    node Node,
    constraints Constraints,
    x, y int,
    depth int,
    visited map[string]bool,
) *LayoutBox {
    if node == nil {
        return nil
    }

    // ... 现有的居中逻辑 (Phase 1.1) ...

    // ✨ Phase 1.3: 设置全局坐标
    // x, y 已经是传入的全局坐标（由父节点累积）
    absX, absY := x, y

    // ... 获取尺寸、检测居中 (Phase 1.1) ...

    box := &LayoutBox{
        ID:       node.ID(),
        X:        x,           // 保存全局坐标（相对 Root）
        Y:        y,
        AbsX:     absX,        // ✨ 明确保存全局坐标
        AbsY:     absY,
        Width:    width,
        Height:   height,
        Layer:    layer,
        ZIndex:   zIndex,
        Border:   nodeBorder,
        ShouldCenter: shouldCenter,
        Children: make([]*LayoutBox, 0),
    }

    // 设置节点位置和尺寸（通知 adapter）
    node.SetPosition(x, y)
    node.SetSize(width, height)

    // 计算边框偏移
    borderOffsetX, borderOffsetY := nodeBorder.ContentOffset()

    // ... Flex / Grid / Wrap 布局逻辑 ...

    // ✨ Phase 1.3: 递归布局子节点时传递全局坐标
    // 默认垂直布局（如果未实现 Flex/Grid）
    childAbsX := absX + borderOffsetX
    childAbsY := absY + borderOffsetY

    for _, child := range node.Children() {
        childBox := e.layoutNodeWithDepth(child, childConstraints, childAbsX, childAbsY, depth+1, visited)
        if childBox != nil {
            box.Children = append(box.Children, childBox)
            childAbsY += childBox.Height  // 累积 Y 坐标
        }
    }

    return box
}
```

#### 对于 Flex布局子节点

```go
// Flex 布局子节点时
for iChild, childBox := range childBoxes {
    child := node.Children()[iChild]
    if child != nil {
        // ✨ Phase 1.3: 计算子节点的全局坐标
        childAbsX := x + childBox.X + borderOffsetX
        childAbsY := y + childBox.Y + borderOffsetY

        childConstraints := Constraints{
            MinWidth:  childBox.Width,
            MaxWidth:  childBox.Width,
            MinHeight: childBox.Height,
            MaxHeight: childBox.Height,
        }

        subBox := e.layoutNodeWithDepth(child, childConstraints, childAbsX, childAbsY, depth+1, visited)
        if subBox != nil {
            subBox.X = childAbsX  // 确保使用全局坐标
            subBox.Y = childAbsY
            box.Children = append(box.Children, subBox)
        }
    }
}
```

#### 验收标准

- [ ] 所有 `LayoutBox` 都有正确的 `AbsX` 和 `AbsY`
- [ ] 根节点的 `AbsX, AbsY == 0, 0`
- [ ] 子节点的全局坐标是父节点 + 偏移
- [ ] 调试日志显示正确的全局坐标

---

## Phase 2: 中期优化 (3-5天)

### 2.1 在 Fiber 中添加 Position 和 Anchor 字段

#### 实现方案

**文件**: `runtime/ui/fiber.go`

```go
type Fiber struct {
    // === Tree Structure ===
    Return  *Fiber
    Child   *Fiber
    Sibling *Fiber
    Alternate *Fiber

    // ... 现有字段 ...

    // ✨ Phase 2.1: Position 语义
    // Position: Relative(默认) / Absolute / Fixed
    // Fixed 用于 Modal/Tooltip 等脱离父布局流的节点
    Position PositionType

    // ✨ Phase 2.1: Anchor 语义
    // 用于 Fixed/Absolute 定位的锚点
    Anchor AnchorType

    // ... 现有字段 ...
}
```

#### 同时添加类型定义（可能已存在）

**文件**: `runtime/layout/position.go`

```go
// PositionType 定位类型
type PositionType int

const (
    PositionRelative PositionType = iota // 默认：相对父节点定位
    PositionAbsolute                     // 绝对定位：相对父节点
    PositionFixed                        // 固定定位：相对 Root/视口
)
```

---

### 2.2 在 completeWork 中填充 Position/Anchor

#### 实现方案

**文件**: `internal/reconciler/complete_work.go`

```go
func completeWork(
    unitOfWork *reconcilerUnitOfWork,
    fiber *Fiber,
) *Fiber {
    // ... 现有代码 ...

    switch fiber.Type {
    case rtui.VNodeComponent:
        fiber = completeWorkComponent(unitOfWork, fiber)
    case rtui.VNodeElement:
        fiber = completeWorkElement(unitOfWork, fiber)
    case rtui.VNodeText:
        fiber = completeWorkText(unitOfWork, fiber)
    }

    // ... 现有代码 ...

    return fiber
}

func completeWorkElement(unitOfWork *reconcilerUnitOfWork, fiber *Fiber) *Fiber {
    // ... 现有代码 (填充 LayoutDirection, LayoutAlign 等) ...

    // ✨ Phase 2.2: 填充 Position
    if fiber.Props != nil {
        if pos, ok := fiber.Props["position"].(string); ok {
            switch pos {
            case "absolute":
                fiber.Position = PositionAbsolute
            case "fixed":
                fiber.Position = PositionFixed
            default:
                fiber.Position = PositionRelative
            }
        }

        // ✨ Phase 2.2: 填充 Anchor
        if anchor, ok := fiber.Props["anchor"].(string); ok {
            fiber.Anchor = ParseAnchorType(anchor)
        } else {
            fiber.Anchor = AnchorTopLeft  // 默认
        }
    }

    return fiber
}

// ParseAnchorType 解析锚点类型
func ParseAnchorType(s string) AnchorType {
    switch s {
    case "top", "topleft":
        return AnchorTopLeft
    case "topcenter":
        return AnchorTop
    case "topright":
        return AnchorTopRight
    case "center", "centerleft":
        return AnchorLeft
    case "centercenter":
        return AnchorCenter
    case "centerright":
        return AnchorRight
    case "bottom", "bottomleft":
        return AnchorBottomLeft
    case "bottomcenter":
        return AnchorBottom
    case "bottomright":
        return AnchorBottomRight
    default:
        return AnchorTopLeft
    }
}
```

---

### 2.3 在布局引擎中实现 PositionFixed 处理

#### 实现方案

**修改**: `FiberToNodeAdapter` 实现 `PositionProvider` 接口

**文件**: `internal/render/fiber_adapter.go`

```go
// 定义 PositionProvider 接口（在 position.go 中）
type PositionProvider interface {
    GetPosition() PositionType
    GetAnchor() AnchorType
}

// FiberToNodeAdapterPure 实现 PositionProvider
func (a *FiberToNodeAdapterPure) GetPosition() PositionType {
    if a.fiber == nil {
        return PositionRelative
    }
    return a.fiber.Position
}

func (a *FiberToNodeAdapterPure) GetAnchor() AnchorType {
    if a.fiber == nil {
        return AnchorTopLeft
    }
    return a.fiber.Anchor
}
```

#### 修改布局引擎

**文件**: `runtime/layout/types.go`

```go
// layoutNodeWithDepth 中添加 Fixed 定位处理
func (e *Engine) layoutNodeWithDepth(
    node Node,
    constraints Constraints,
    x, y int,
    depth int,
    visited map[string]bool,
) *LayoutBox {
    if node == nil {
        return nil
    }

    // ... 现有代码 (获取尺寸、Layer、ZIndex) ...

    // ✨ Phase 1.1: Modal 居中逻辑 (保持不变)
    shouldCenter := ...
    if shouldCenter {
        x = (constraints.MaxWidth - width) / 2
        y = (constraints.MaxHeight - height) / 2
    }

    // ✨ Phase 2.3: PositionFixed 处理
    // 如果节点使用了 fixed 定位，直接使用 Root 坐标系
    position := PositionRelative
    anchor := AnchorTopLeft

    if posProvider, ok := node.(PositionProvider); ok {
        position = posProvider.GetPosition()
        anchor = posProvider.GetAnchor()
    }

    if position == PositionFixed {
        // ✨ Fixed 定位：以 Root 为参考系
        // 计算锚点位置
        rootW := constraints.MaxWidth
        rootH := constraints.MaxHeight

        switch anchor {
        case AnchorTopLeft:
            x, y = 0, 0
        case AnchorTop:
            x = (rootW - width) / 2
            y = 0
        case AnchorTopRight:
            x = rootW - width
            y = 0
        case AnchorLeft:
            x = 0
            y = (rootH - height) / 2
        case AnchorCenter:
            x = (rootW - width) / 2
            y = (rootH - height) / 2
        case AnchorRight:
            x = rootW - width
            y = (rootH - height) / 2
        case AnchorBottomLeft:
            x = 0
            y = rootH - height
        case AnchorBottom:
            x = (rootW - width) / 2
            y = rootH - height
        case AnchorBottomRight:
            x = rootW - width
            y = rootH - height
        }
    }

    // ... 现有代码 (创建 LayoutBox, 设置 AbsX/AbsY) ...

    return box
}
```

---

### 2.4 实现 Anchor 居中计算逻辑

Anchor 逻辑已在 Phase 2.3 中集成到 `PositionFixed` 处理中。这里补充测试用例。

#### 测试用例

**文件**: `runtime/layout/position_test.go`

```go
func TestEngine_PositionFixed_Anchors(t *testing.T) {
    tests := []struct {
        name     string
        anchor   AnchorType
        nodeW    int
        nodeH    int
        rootW    int
        rootH    int
        expectX  int
        expectY  int
    }{
        {"TopLeft", AnchorTopLeft, 10, 5, 80, 24, 0, 0},
        {"TopCenter", AnchorTop, 10, 5, 80, 24, 35, 0},      // (80-10)/2
        {"TopRight", AnchorTopRight, 10, 5, 80, 24, 70, 0},    // 80-10
        {"CenterLeft", AnchorLeft, 10, 5, 80, 24, 0, 9},       // (24-5)/2
        {"Center", AnchorCenter, 10, 5, 80, 24, 35, 9},        // (80-10)/2, (24-5)/2
        {"CenterRight", AnchorRight, 10, 5, 80, 24, 70, 9},    // 80-10, (24-5)/2
        {"BottomLeft", AnchorBottomLeft, 10, 5, 80, 24, 0, 19}, // 24-5
        {"BottomCenter", AnchorBottom, 10, 5, 80, 24, 35, 19},  // (80-10)/2, 24-5
        {"BottomRight", AnchorBottomRight, 10, 5, 80, 24, 70, 19}, // 80-10, 24-5
    }

    engine := NewEngine()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 创建一个固定定位的节点
            node := &mockFixedNode{
                width:  tt.nodeW,
                height: tt.nodeH,
                position: PositionFixed,
                anchor: tt.anchor,
            }

            constraints := Constraints{
                MinWidth:  0,
                MaxWidth:  tt.rootW,
                MinHeight: 0,
                MaxHeight: tt.rootH,
            }

            box := engine.Layout(node, constraints)

            if box.Root.X != tt.expectX || box.Root.Y != tt.expectY {
                t.Errorf("Anchor %s: expect (%d, %d), got (%d, %d)",
                    tt.name, tt.expectX, tt.expectY, box.Root.X, box.Root.Y)
            }
        })
    }
}

// 测试辅助类型
type mockFixedNode struct {
    width, height int
    position     PositionType
    anchor       AnchorType
}

func (m *mockFixedNode) ID() string                        { return "fixed-node" }
func (m *mockFixedNode) GetSize() (int, int)               { return m.width, m.height }
func (m *mockFixedNode) Children() []Node                  { return nil }
func (m *mockFixedNode) SetPosition(x, y int)              {}
func (m *mockFixedNode) SetSize(w, h int)                  {}
func (m *mockFixedNode) GetPosition() PositionType         { return m.position }
func (m *mockFixedNode) GetAnchor() AnchorType             { return m.anchor }
```

#### 验收标准

- [ ] 所有 9 种 Anchor 定位都正确
- [ ] PositionFixed 节点不受父布局影响
- [ ] 坐标以 Root 为参考系
- [ ] 测试用例全部通过

---

## Phase 3: 长期特性 (1-2周) [可选]

### 3.1 在 Fiber 中添加 PortalRoot 字段

#### 实现方案

**文件**: `runtime/ui/fiber.go`

```go
type Fiber struct {
    // === Tree Structure ===
    Return  *Fiber
    Child   *Fiber
    Sibling *Fiber
    Alternate *Fiber

    // ... 现有字段 ...

    // ✨ Phase 3.1: Portal 支持
    // PortalRoot 指定该节点在布局/渲染时挂载到的目标
    // 常用于 Modal/Tooltip 等需要跨树渲染的组件
    // nil 表示正常挂载在父节点下
    PortalRoot *Fiber

    // ... 现有字段 ...
}
```

### 3.2 实现 OverlayManager

#### 设计

```go
// runtime/layout/overlay.go

// OverlayManager 管理所有浮层节点
type OverlayManager struct {
    mu    sync.RWMutex
    stack []*LayoutBox  // 按优先级排序的浮层
}

// Push 添加一个新的浮层到栈顶
func (o *OverlayManager) Push(box *LayoutBox) {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.stack = append(o.stack, box)
}

// Pop 移除栈顶的浮层
func (o *OverlayManager) Pop() *LayoutBox {
    o.mu.Lock()
    defer o.mu.Unlock()
    if len(o.stack) == 0 {
        return nil
    }
    top := o.stack[len(o.stack)-1]
    o.stack = o.stack[:len(o.stack)-1]
    return top
}

// Top 返回栈顶的浮层
func (o *OverlayManager) Top() *LayoutBox {
    o.mu.RLock()
    defer o.mu.RUnlock()
    if len(o.stack) == 0 {
        return nil
    }
    return o.stack[len(o.stack)-1]
}

// GetAll 返回所有浮层（按 Z 逆序，用于渲染/事件）
func (o *OverlayManager) GetAll() []*LayoutBox {
    o.mu.RLock()
    defer o.mu.RUnlock()
    // 返回副本，避免外部修改
    result := make([]*LayoutBox, len(o.stack))
    copy(result, o.stack)
    return result
}
```

#### 集成到 Layout

**文件**: `runtime/layout/types.go`

```go
type Engine struct {
    cache        *LayoutCache
    hitMap       *HitMap
    stats        LayoutStats
    overlayMgr   *OverlayManager  // ✨ Phase 3.2
}

func NewEngine() *Engine {
    return &Engine{
        cache:      NewLayoutCache(),
        hitMap:     NewHitMap(),
        overlayMgr: NewOverlayManager(),  // ✨ 初始化
    }
}
```

### 3.3 实现 Portal 跨树挂载

#### 需要修改的工作流

1. **Reconciliation**: 检测 Fiber.PortalRoot
2. **Layout**: 将 Portal 节点挂载到指定目标
3. **Render**: Portal 节点跟随目标渲染

**Phase 3 暂不展开详细实现，等待 Phase 1-2 完成后的需求评估。**

---

## 验收标准

### Phase 1 验收

| 标准 | 验证方法 |
|-----|---------|
| Modal 居中在 Layout 阶段完成 | 日志显示 `layoutNodeWithDepth` 计算了居中坐标 |
| LayoutBox 有 AbsX/AbsY | 检查调试输出或断言 |
| 全局坐标正确递归 | Root.AbsX=0, Child.AbsX=Parent.AbsX+Offset |
| Modal 默认居中 | 运行 Demo，Modal 显示在屏幕中央 |
| Paint 阶段不修改坐标 | 查看 `PaintPaintablePlanes` 日志 |

### Phase 2 验收

| 标准 | 验证方法 |
|-----|---------|
| Fiber 有 Position/Anchor 字段 | 检查 `fiber.go` 结构体定义 |
| completeWork 填充 Position/Anchor | 单元测试 |
| PositionFixed 正确计算坐标 | 运行 `position_test.go` |
| 9 种 Anchor 都支持 | 运行测试用例 |
| Fixed 节点不受父布局影响 | 创建嵌套结构的测试用例 |

### Phase 3 验收（如需实现）

| 标准 | 验证方法 |
|-----|---------|
| Fiber 有 PortalRoot 字段 | 检查 `fiber.go` |
| OverlayManager 正确管理浮层 | 单元测试 |
| Modal 跨树渲染 | 端到端测试 |
| Focus Trap for Modal | 交互测试 |

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| Modal 居中位置计算错误，与现有使用冲突 | 高 | 先运行所有现有测试，确保功能不变 |
| AbsX/AbsY 计算错误导致渲染错乱 | 高 | 增加单元测试和可视化调试工具 |
| PositionFixed 破坏现有布局 | 中 | 默认 PositionRelative，需要显式设置才启用 |
| 性能问题（递归计算开销） | 中 | 使用缓存和增量布局 |
| API 变更导致用户代码失效 | 低 | 保留旧 API（标记 deprecated），添加迁移指南 |

---

## 参考资料

- [Fiber-first 渲染路径完整分析](../fiber/FIBER_FIRST_PAINT_PATH.md)
- [Layer 系统设计文档](./layer_design.md)
- [Layer 架构文档](./LAYER_SYSTEM_ARCHITECTURE.md)
- [Modal 定位实现指南](./POSITIONING.md)
- [布局引擎文档](../layout/README.md)

---

## 附录: 快速评估检查清单

在开始实施前，请确认：

- [ ] 我已阅读并理解 `layer_design.md` 的设计原则
- [ ] 我了解 Current 实现和设计意图的差异
- [ ] 我确认 Phase 1-3 的优先级符合实际需求
- [ ] 我已准备好运行测试确保向后兼容
- [ ] 我有权限修改相应文件（`runtime/layout`, `internal/render`, `runtime/ui`）

---

**文档版本**: v1.0
**创建日期**: 2026-03-01
**最后更新**: 2026-03-01
**作者**: Qwen Code
