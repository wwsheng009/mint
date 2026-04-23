# Layer 系统架构深度分析

## 用户提出的四个核心问题

### 1. Inspector 过多干涉渲染层，需要显式调用 setLayer

### 2. 渲染引擎多层渲染机制，Inspector 位置为什么不是 (0, 0)

### 3. 架构优化需求：一个 root 节点下多个节点同时渲染

### 4. Inspector 如何覆盖显示，为什么是 (40, 5)

---

## 问题 1: Inspector 与 Layer 系统的耦合

### 当前设计的问题

```go
// internal/inspector/standalone_inspector.go:283
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    content := si.buildOverlayContent()

    // Inspector 必须知道 Layer 系统
    content.SetLayer(rtui.LayerInspector)  // ← 耦合点

    return content
}
```

**问题**:
- Inspector 组件需要知道 `rtui.LayerInspector` 的存在
- Inspector 需要知道如何使用 `SetLayer()` API
- 违反了单一职责原则：Inspector 应该只负责 UI 内容

### 更好的设计

#### 方案 A: 外部标记 Layer

```go
// Inspector 不需要知道 Layer 系统
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
    return si.buildOverlayContent()
}

// 应用层负责标记 Layer
func RuntimeDemo() ui.VNode {
    inspectorContent := globalInspector.RenderOverlay()

    // 由应用层决定 Inspector 是什么 layer
    inspectorContent.SetLayer(rtui.LayerInspector)

    return ui.Fragment(appContent, inspectorContent)
}
```

**优势**:
- ✅ Inspector 组件不需要知道 Layer 系统
- ✅ 应用层可以灵活控制 Inspector 的 layer 类型
- ✅ 可以根据需要将 Inspector 放在不同 layer

#### 方案 B: 自动 Layer 检测

```go
// 渲染系统自动识别 Inspector
func (r *PipelineRenderer) hasOverlayComponents(vnode rtui.VNode) bool {
    // 检测 Inspector 组件类型
    if _, ok := vnode.(*inspector.StandaloneInspector); ok {
        return true
    }
    // 递归检测
    for _, child := range vnode.Children() {
        if r.hasOverlayComponents(child) {
            return true
        }
    }
    return false
}

// 自动分配 Layer
func (r *PipelineRenderer) autoAssignLayers(vnode rtui.VNode) {
    if inspector, ok := vnode.(*inspector.StandaloneInspector); ok {
        vnode.SetLayer(rtui.LayerInspector)
    }
    // ...
}
```

**优势**:
- ✅ 组件完全不需要知道 Layer 系统
- ✅ 渲染系统自动识别特殊组件
- ✅ 更符合面向对象设计原则

---

## 问题 2: 渲染引擎的多层渲染机制

### PaintEngine.PaintLayers() 实现

```go
// internal/render/paint_engine.go:329-376
func (e *PaintEngine) PaintLayers(
    layouts LayerLayouts,
    buffer *paint.Buffer,
) error {
    // 定义渲染顺序 (z-order)
    renderOrder := []rtui.Layer{
        rtui.LayerBase,       // 0: 底层
        rtui.LayerOverlay,    // 1: 覆盖层
        rtui.LayerModal,      // 2: 模态框
        rtui.LayerTooltip,    // 3: 提示
        rtui.LayerInspector,  // 4: Inspector (最高层)
    }

    // 按顺序渲染每个 layer
    for _, layer := range renderOrder {
        layout, ok := layouts[layer]
        if !ok || layout.Root == nil {
            continue
        }

        // 渲染这个 layer
        e.Paint(layout, buffer)

        // 特殊处理：Modal 需要绘制背景遮罩
        if layer == rtui.LayerModal {
            e.paintModalBackdrop(layout.Root, buffer)
        }
    }

    return nil
}
```

### 多层渲染的关键特性

#### 1. Z-Order 渲染

```
渲染顺序 (从底到顶):
┌─────────────────────────────────────────────────────┐
│ Layer 4: Inspector      (最上层，覆盖所有内容)        │
├─────────────────────────────────────────────────────┤
│ Layer 3: Tooltip        (提示框)                     │
├─────────────────────────────────────────────────────┤
│ Layer 2: Modal          (模态框 + 背景遮罩)          │
├─────────────────────────────────────────────────────┤
│ Layer 1: Overlay        (覆盖层)                     │
├─────────────────────────────────────────────────────┤
│ Layer 0: Base           (应用内容，最底层)           │
└─────────────────────────────────────────────────────┘
```

#### 2. 独立布局

每个 layer **独立布局**，使用自己的坐标系统：

```go
// LayerManager.CollectAndLayout()
for layer, nodes := range collector.GetLayers() {
    for _, node := range nodes {
        // 每个 layer 独立布局
        layerLayout := engine.Layout(node.Content, constraints)

        // 每个 layer 可以有自己的位置
        if layer == rtui.LayerInspector {
            m.positionInspector(node, layerLayout.Root)
        }

        layouts[layer] = layerLayout
    }
}
```

#### 3. 位置计算

```go
// runtime/layer/manager.go:232-273
func (m *Manager) positionInspector(node *LayerNode, root *compute.ComputedBox) {
    // 获取 Inspector 期望的位置
    props := node.Content.Props()
    targetX, targetY := 0, 0

    if x, ok := props["x"].(int); ok {
        targetX = x
    }
    if y, ok := props["y"].(int); ok {
        targetY = y
    }

    // 计算偏移并移动所有节点
    offsetX := targetX - root.Box.X
    offsetY := targetY - root.Box.Y
    m.shiftPositions(root, offsetX, offsetY)
}
```

---

## 问题 3: Inspector 位置为什么是 (40, 5)？

### 位置配置

```go
// internal/inspector/standalone_inspector.go:115-131
func NewStandaloneInspector() *StandaloneInspector() {
    // 计算默认位置以适应典型屏幕 (120x40)
    // Inspector 宽度是 80，所以定位在 x=40 以保持在屏幕边界内
    defaultX := 40  // 屏幕宽度 120 - Inspector 宽度 80 = 40
    defaultY := 5   // 顶部边距

    return &StandaloneInspector{
        overlayWidth:  80,
        overlayHeight: 25,
        floatX:        defaultX,  // ← X 位置
        floatY:        defaultY,  // ← Y 位置
    }
}
```

### 位置计算逻辑

```
屏幕宽度: 120 像素
Inspector 宽度: 80 像素

计算: Inspector X = 屏幕宽度 - Inspector 宽度
                    = 120 - 80
                    = 40

屏幕布局:
┌──────────────────────────────────────────────────────────┐
│  (0,0)                                                  │
│  ┌────────────────────┐  ┌────────────────────────┐    │
│  │                    │  │  Inspector            │    │
│  │  App Content       │  │  (40, 5)              │    │
│  │  (0, 0)            │  │  Width: 80            │    │
│  │  Width: 40         │  │  Height: 25           │    │
│  │                    │  │                       │    │
│  └────────────────────┘  └────────────────────────┘    │
│                                                          │
└──────────────────────────────────────────────────────────┘
     0  ..........  40  ..........  120
     └─ App Area ─┘  └── Inspector Area ──┘
```

### 为什么不是 (0, 0)？

**因为 (0, 0) 会导致**:
1. Inspector 覆盖整个应用内容
2. 用户无法看到应用界面
3. 违背了 Inspector 作为"开发工具"的设计初衷

**设计意图**:
- Inspector 应该悬浮在应用内容旁边
- 应用内容和 Inspector 同时可见
- 用户可以同时操作两者

---

## 问题 4: Inspector 如何覆盖显示？

### 绝对定位系统

Layer 系统使用**绝对定位**，每个 layer 有独立的坐标系统：

```
LayerBase (appContent):
  └─ 布局位置: (0, 0)
      └─ 占据: (0, 0) 到 (40, 40)

LayerInspector (inspectorOverlay):
  └─ 布局位置: (0, 0)  ← 初始布局位置
      └─ 定位偏移: (40, 5)  ← positionInspector 添加的偏移
      └─ 最终位置: (40, 5)
          └─ 占据: (40, 5) 到 (120, 30)
```

### 覆盖显示原理

#### 步骤 1: 独立布局

```go
// 每个 layer 先在 (0, 0) 位置布局
baseLayout = engine.Layout(appContent, constraints)
  → baseLayout.Root.Box = (0, 0, 40, 40)  // 假设

inspectorLayout = engine.Layout(inspectorOverlay, constraints)
  → inspectorLayout.Root.Box = (0, 0, 80, 25)
```

#### 步骤 2: 位置调整

```go
// LayerManager.positionInspector()
// 计算偏移
offsetX := 40 - 0 = 40  // 目标 X - 当前 X
offsetY := 5 - 0 = 5    // 目标 Y - 当前 Y

// 应用偏移到所有节点
shiftPositions(inspectorLayout.Root, 40, 5)
  → inspectorLayout.Root.Box = (40, 5, 80, 25)
```

#### 步骤 3: 顺序绘制

```go
// PaintEngine.PaintLayers()
// 按顺序绘制到同一个 buffer

// 1. 先绘制 Base layer
Paint(baseLayout, buffer)
  → 绘制到 (0, 0) 区域

// 2. 再绘制 Inspector layer (会覆盖上面的内容)
Paint(inspectorLayout, buffer)
  → 绘制到 (40, 5) 区域
```

### Buffer 的最终状态

```
Buffer (120x40):

┌─────────────────────────────────────────────────────────┐
│ (0,0)                            Inspector (40,5)        │
│ ┌──────────────────┐  ┌───────────────────────────┐  │
│ │ App Content      │  │ INSPECTOR                │  │
│ │ Runtime Pipeline │  │ Elements | Console | ...  │  │
│ │ Statistics       │  │ ┌─────────────────────┐  │  │
│ │ Control Panel    │  │ │ TreeView            │  │  │
│ │                  │  │ │ - VStack            │  │  │
│ │                  │  │ │   - Bordered        │  │  │
│ │                  │  │ │     - VStack        │  │  │
│ └──────────────────┘  │ └─────────────────────┘  │  │
│                      └───────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**关键点**:
- Buffer 是一个二维数组 `buffer.Cells[y][x]`
- 后绘制的 layer 会覆盖先绘制的 layer 的像素
- 不同 layer 在不同区域绘制，互不干扰

---

## 问题 5: 架构优化建议

### 当前架构的问题

```go
// 用户需要显式管理 Layer
func RuntimeDemo() ui.VNode {
    inspectorOverlay := globalInspector.RenderOverlay()
    inspectorOverlay.SetLayer(rtui.LayerInspector)  // ← 用户手动设置

    modal := someModal
    modal.SetLayer(rtui.LayerModal)  // ← 用户手动设置

    return ui.Fragment(appContent, inspectorOverlay, modal)
}
```

**问题**:
1. ❌ 每个组件都需要知道 Layer 系统
2. ❌ 容易忘记设置 Layer
3. ❌ Layer 管理分散在各个地方
4. ❌ 难以统一管理和调试

### 优化方案 A: 声明式 Layer 系统

```go
// 定义新的 Layer 组件
func Layer(layerType rtui.Layer, children ...ui.VNode) ui.VNode {
    node := ui.Fragment(children...)
    node.SetLayer(layerType)
    return node
}

// 使用
func RuntimeDemo() ui.VNode {
    return ui.VStack(
        appContent,
        Layer(rtui.LayerInspector,
            globalInspector.RenderOverlay(),  // 不需要 SetLayer
        ),
        Layer(rtui.LayerModal,
            someModal,  // 不需要 SetLayer
        ),
    )
}
```

**优势**:
- ✅ 组件不需要知道 `SetLayer()`
- ✅ Layer 声明更加明确
- ✅ 更容易理解和管理

### 优化方案 B: 自动 Layer 推断

```go
// PipelineRenderer 自动识别组件类型
func (r *PipelineRenderer) detectLayer(vnode rtui.VNode) rtui.Layer {
    switch vnode.(type) {
    case *inspector.StandaloneInspector:
        return rtui.LayerInspector
    case *ModalComponent:
        return rtui.LayerModal
    case *TooltipComponent:
        return rtui.LayerTooltip
    default:
        return rtui.LayerBase
    }
}

// 自动分配 Layer
func (r *PipelineRenderer) autoAssignLayers(vnode rtui.VNode) {
    layer := r.detectLayer(vnode)
    if layer != rtui.LayerBase {
        vnode.SetLayer(layer)
    }

    for _, child := range vnode.Children() {
        r.autoAssignLayers(child)
    }
}
```

**优势**:
- ✅ 完全自动，用户无需关心
- ✅ 基于 Go 类型系统，类型安全
- ✅ 易于扩展新的 layer 类型

### 优化方案 C: Layer 容器组件

```go
// 定义 Layer 容器
type LayerContainer struct {
    layerType rtui.Layer
    children  []ui.VNode
}

func NewInspectorLayer(children ...ui.VNode) *LayerContainer {
    return &LayerContainer{
        layerType: rtui.LayerInspector,
        children:  children,
    }
}

func (lc *LayerContainer) GetLayer() rtui.Layer {
    return lc.layerType
}

func (lc *LayerContainer) Children() []ui.VNode {
    return lc.children
}

// 使用
func RuntimeDemo() ui.VNode {
    return ui.VStack(
        appContent,
        NewInspectorLayer(  // 自动管理 Layer
            globalInspector.RenderOverlay(),
        ),
    )
}
```

**优势**:
- ✅ 类型安全
- ✅ 明确的语义
- ✅ 易于维护和测试

---

## 总结与建议

### 当前实现

| 特性 | 实现方式 | 问题 |
|------|---------|------|
| Layer 设置 | 组件调用 `SetLayer()` | 组件需要知道 Layer 系统 |
| 位置管理 | Props `{"x": 40, "y": 5}` | 魔法数字，硬编码 |
| 多层渲染 | `PaintLayers()` 按顺序绘制 | 工作正常 ✅ |
| 覆盖显示 | 绝对定位 + 顺序绘制 | 工作正常 ✅ |

### 建议的改进

1. **解耦 Layer 系统和组件**
   - 组件不应该知道 `SetLayer()` 的存在
   - 由外部系统或容器组件管理 Layer

2. **改进位置管理**
   - 使用枚举或常量代替魔法数字
   - 例如: `PositionTopRight`, `PositionCenter`

3. **完善架构**
   - 引入 `LayerContainer` 概念
   - 或实现自动 Layer 推断

4. **改进测试**
   - 添加多层渲染的集成测试
   - 测试覆盖、定位、交互等场景

### 当前可用的方案

**立即可用** (当前代码已支持):

```go
func RuntimeDemo() ui.VNode {
    inspectorVisible := globalInspector.IsVisible()

    if inspectorVisible {
        inspectorOverlay := globalInspector.RenderOverlay()
        // ✅ 使用 Fragment 避免嵌套布局
        return ui.Fragment(
            appContent,
            inspectorOverlay,
        )
    }

    return appContent
}
```

**工作原理**:
1. Fragment 不创建额外布局节点
2. StripLayers 正确分离两个 layer
3. 两个 layer 独立布局和绘制
4. Inspector 通过绝对定位显示在 (40, 5)
5. 多层渲染按 z-order 顺序绘制到 buffer
