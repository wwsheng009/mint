# Layer 系统分析与实现计划

> **版本**: 1.0
> **创建日期**: 2025-02-06
> **状态**: 分析完成，待实施

---

## 一、设计文档回顾

### 1.1 参考文档

| 文档 | 路径 | 核心内容 |
|------|------|----------|
| Modal 设计思路 | `framework/docs/ui/idea/idea4.3_modal.md` | Layer 系统核心理念 |
| Layer 系统设计 | `framework/docs/ui/design/LAYER_SYSTEM_DESIGN.md` | 详细技术方案 |

### 1.2 设计目标

Layer 系统要解决的核心问题：

1. **Z 轴覆盖** - Modal 浮在所有内容之上
2. **事件阻断** - Modal 打开时背景不响应
3. **焦点陷阱** - Tab 键只在 Modal 内循环
4. **独立布局** - Modal 不参与正常布局流
5. **ESC 关闭** - 统一的关闭机制

---

## 二、当前系统架构分析

### 2.1 渲染管线

```
┌─────────────────────────────────────────────────────────────────┐
│                    渲染管线 (Rendering Pipeline)                 │
├─────────────────────────────────────────────────────────────────┤
│  VNode Tree                                                      │
│      ↓                                                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Fiber Reconciler (协调器)                                  │ │
│  │  - 创建/更新 Fiber 树                                       │ │
│  │  - 处理 State 更新                                          │ │
│  │  - 标记 Dirty 节点                                          │ │
│  └────────────────────────────────────────────────────────────┘ │
│      ↓                                                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Layout Phase (布局阶段)                                    │ │
│  │  - compute.Engine.Layout()                                 │ │
│  │  - 约束驱动的布局计算                                       │ │
│  │  - 输出: ComputedLayout { Root *ComputedBox }             │ │
│  └────────────────────────────────────────────────────────────┘ │
│      ↓                                                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Paint Phase (绘制阶段)                                     │ │
│  │  - PaintEngine.Paint()                                     │ │
│  │  - 按 ComputedBox 位置渲染                                  │ │
│  │  - 输出: buffer                                             │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 当前文件结构

| 文件 | 职责 | Layer 相关 |
|------|------|-----------|
| `runtime/ui/fiber.go` | Fiber 树结构 | ❌ 无 Layer 支持 |
| `internal/reconciler/fiber.go` | 协调器 | ❌ 无 Layer 支持 |
| `runtime/compute/engine.go` | 布局引擎 | ❌ 单一布局树 |
| `internal/render/paint_engine.go` | 绘制引擎 | ❌ 单层绘制 |
| `internal/render/rendering_pipeline.go` | 渲染管线 | ❌ 无 Layer 阶段 |
| `runtime/paint/layer.go` | Paint Layer | ⚠️ 仅有类型定义 |
| `runtime/focus/trap.go` | 焦点陷阱 | ✅ 已实现 |
| `components/overlay/modal.go` | Modal 组件 | ❌ 普通组件 |

### 2.3 当前 Modal 实现（demo1_full_featured）

```go
// 当前实现方式：条件渲染
func App() ui.VNode {
    showModal, setShowModal := ui.UseStateBool(false)

    mainContent := ui.VStackBuilder(
        Header(...),
        MainBody(...),
    ).Stretch().Build()

    // ❌ Modal 作为 VStack 子元素，参与正常布局
    if showModal {
        return ui.VStackBuilder(
            mainContent,      // 背景内容
            ConfirmModal(...), // Modal
        ).Stretch().Build()
    }

    return mainContent
}
```

**问题分析**:
1. Modal 是 VStack 的子元素，受父容器约束
2. Modal 会压缩主内容，不是真正的"覆盖"
3. 背景组件仍可交互（焦点问题）
4. 无 ESC 键关闭机制
5. 无遮罩（dim background）

---

## 三、现有能力盘点

### 3.1 已实现功能 ✅

| 功能 | 位置 | 状态 |
|------|------|------|
| Focus Trap | `runtime/focus/trap.go` | ✅ 完整 |
| TrapManager (栈管理) | `runtime/focus/trap.go` | ✅ 完整 |
| PushTrap/PopTrap | `runtime/focus/manager.go` | ✅ 已集成 |
| Paint Layer 类型 | `runtime/paint/layer.go` | ⚠️ 仅定义 |
| Fiber 树结构 | `runtime/ui/fiber.go` | ✅ 完整 |
| 两阶段渲染 | `rendering_pipeline.go` | ✅ 完整 |

### 3.2 缺失功能 ❌

| 功能 | 位置 | 状态 |
|------|------|------|
| LayerManager | - | ❌ 无 |
| LayerTree | - | ❌ 无 |
| Layer 收集机制 | - | ❌ 无 |
| Overlay Layout Pass | - | ❌ 无 |
| 多层 Paint | - | ❌ 无 |
| 事件阻断机制 | - | ❌ 无 |
| ESC 键处理 | - | ❌ 无 |
| ui.Modal() API | - | ❌ 无 |
| 动画系统 | - | ❌ 无 |

---

## 四、Layer 系统设计方案

### 4.1 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                      Layer 系统架构                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  VNode Tree                                                      │
│      ↓                                                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Fiber Reconciler                                           │ │
│  │                                                              │ │
│  │  ┌────────────────────────────────────────────────────────┐ │ │
│  │  │  LayerCollector (新增)                                  │ │ │
│  │  │  - 扫描 Fiber 树                                      │ │ │
│  │  │  - 识别 Modal/Overlay/Tooltip 节点                     │ │ │
│  │  │  - 提取到 LayerTree                                    │ │ │
│  │  └────────────────────────────────────────────────────────┘ │ │
│  │           ↓                                                  │ │
│  │  ┌────────────────────────────────────────────────────────┐ │ │
│  │  │  Main Fiber Tree (无 Modal)                            │ │ │
│  │  └────────────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────┘ │
│      ↓                                                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  LayerManager (新增)                                        │ │
│  │  - 管理 4 个 Layer                                         │ │
│  │  - 每个独立 Layout                                         │ │
│  │  - 维护 Layer 栈                                           │ │
│  └────────────────────────────────────────────────────────────┘ │
│      ↓                                                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Multi-Layer Layout Phase (扩展)                           │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │ │
│  │  │ Base Layer  │ │ Overlay     │ │ Modal        │        │ │
│  │  │ Layout      │ │ Layout      │ │ Layout (居中) │        │ │
│  │  └─────────────┘ └─────────────┘ └─────────────┘        │ │
│  └────────────────────────────────────────────────────────────┘ │
│      ↓                                                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Multi-Layer Paint Phase (扩展)                            │ │
│  │  for layer := Base → Modal:                                │ │
│  │      Paint(layer)                                          │ │
│  └────────────────────────────────────────────────────────────┘ │
│      ↓                                                           │
│  Final Buffer                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 Layer 定义

```go
// runtime/layer/types.go

package layer

// Layer 视觉层级
type Layer int

const (
    LayerBase Layer = iota   // 基础层 (默认内容)
    LayerOverlay             // 覆盖层 (Dropdown, Popover)
    LayerModal               // 模态框层
    LayerTooltip             // 提示框层
)

// String 返回层级名称
func (l Layer) String() string {
    switch l {
    case LayerBase:
        return "base"
    case LayerOverlay:
        return "overlay"
    case LayerModal:
        return "modal"
    case LayerTooltip:
        return "tooltip"
    default:
        return "unknown"
    }
}
```

### 4.3 VNode 扩展

```go
// runtime/ui/vnode.go

// Layer 属性
func (v *VNode) SetLayer(layer layer.Layer) {
    if v.Props == nil {
        v.Props = make(Props)
    }
    v.Props["_layer"] = layer
}

func (v *VNode) GetLayer() layer.Layer {
    if v.Props == nil {
        return layer.LayerBase
    }
    if l, ok := v.Props["_layer"].(layer.Layer); ok {
        return l
    }
    return layer.LayerBase
}
```

### 4.4 Layer 收集器

```go
// runtime/layer/collector.go

package layer

import (
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// LayerNode 层级节点
type LayerNode struct {
    Layer   Layer
    ID      string
    Content rtui.VNode
    Visible bool
}

// Collector 收集 Layer 节点
type Collector struct {
    layers map[Layer][]*LayerNode
}

// NewCollector 创建收集器
func NewCollector() *Collector {
    return &Collector{
        layers: make(map[Layer][]*LayerNode),
    }
}

// Collect 从 VNode 树中收集 Layer 节点
func (c *Collector) Collect(vnode rtui.VNode) {
    c.walk(vnode)
}

// walk 遍历 VNode 树
func (c *Collector) walk(vnode rtui.VNode) {
    if vnode == nil {
        return
    }

    // 检查是否有 Layer 属性
    if layer := vnode.GetLayer(); layer != LayerBase {
        // 添加到对应 Layer
        node := &LayerNode{
            Layer:   layer,
            ID:      vnode.Key(),
            Content: vnode,
            Visible: true,
        }
        c.layers[layer] = append(c.layers[layer], node)

        // 不继续遍历 Layer 节点的子节点
        // 因为它们会独立布局
        return
    }

    // 递归遍历子节点
    for _, child := range vnode.Children() {
        c.walk(child)
    }
}

// GetLayers 返回所有 Layer
func (c *Collector) GetLayers() map[Layer][]*LayerNode {
    return c.layers
}

// HasModal 检查是否有 Modal
func (c *Collector) HasModal() bool {
    return len(c.layers[LayerModal]) > 0
}
```

### 4.5 LayerManager

```go
// runtime/layer/manager.go

package layer

import (
    "github.com/wwsheng009/mint/runtime"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Manager 管理多个 Layer 的布局和渲染
type Manager struct {
    collector *Collector
    layouts   map[Layer]*runtime.ComputedLayout
}

// NewManager 创建 Manager
func NewManager() *Manager {
    return &Manager{
        collector: NewCollector(),
        layouts:   make(map[Layer]*runtime.ComputedLayout),
    }
}

// CollectAndLayout 收集 Layer 并执行布局
func (m *Manager) CollectAndLayout(
    vnode rtui.VNode,
    constraints runtime.BoxConstraints,
    engine Engine,
) error {
    // 1. 收集 Layer 节点
    m.collector.Collect(vnode)

    // 2. 移除主树中的 Layer 节点，得到纯净的 Base 树
    baseTree := m.stripLayers(vnode)

    // 3. 布局 Base Layer
    baseLayout, err := engine.Layout(baseTree, constraints)
    if err != nil {
        return err
    }
    m.layouts[LayerBase] = baseLayout

    // 4. 布局其他 Layer (独立约束)
    for layer, nodes := range m.collector.GetLayers() {
        for _, node := range nodes {
            if !node.Visible {
                continue
            }

            // Modal: 居中布局
            if layer == LayerModal {
                modalLayout := m.layoutModal(node, constraints, engine)
                m.layouts[layer] = modalLayout
            }
        }
    }

    return nil
}

// layoutModal 布局 Modal (居中)
func (m *Manager) layoutModal(
    node *LayerNode,
    constraints runtime.BoxConstraints,
    engine Engine,
) *runtime.ComputedLayout {
    // 使用全屏约束布局 Modal
    modalConstraints := runtime.BoxConstraints{
        MinWidth:  0,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: 0,
        MaxHeight: constraints.MaxHeight,
    }

    layout, _ := engine.Layout(node.Content, modalConstraints)

    // 计算居中位置
    if layout.Root != nil {
        rootBox := layout.Root.Box
        offsetX := (constraints.MaxWidth - rootBox.Width) / 2
        offsetY := (constraints.MaxHeight - rootBox.Height) / 2

        // 重新定位 Modal 及其子节点
        m.shiftPositions(layout.Root, offsetX, offsetY)
    }

    return layout
}

// GetLayouts 返回所有 Layer 的布局
func (m *Manager) GetLayouts() map[Layer]*runtime.ComputedLayout {
    return m.layouts
}

// HasModal 检查是否有 Modal
func (m *Manager) HasModal() bool {
    return m.collector.HasModal()
}
```

### 4.6 PaintEngine 扩展

```go
// internal/render/paint_engine.go

// PaintLayers 渲染多个 Layer
func (e *PaintEngine) PaintLayers(
    layouts map[layer.Layer]*compute.ComputedLayout,
    buffer *paint.Buffer,
) error {
    // 按层级顺序渲染
    for l := layer.LayerBase; l <= layer.LayerTooltip; l++ {
        layout, ok := layouts[l]
        if !ok || layout.Root == nil {
            continue
        }

        // 渲染该层
        if err := e.Paint(layout, buffer); err != nil {
            return err
        }

        // Modal 层需要绘制遮罩
        if l == layer.LayerModal {
            e.paintModalBackdrop(layout.Root, buffer)
        }
    }
    return nil
}

// paintModalBackdrop 绘制 Modal 背景遮罩
func (e *PaintEngine) paintModalBackdrop(
    root *compute.ComputedBox,
    buffer *paint.Buffer,
) {
    // 在 Modal 后面绘制半透明遮罩
    // 由于 TUI 限制，使用字符阴影
    for y := root.Box.Y; y < root.Box.Y+root.Box.Height; y++ {
        for x := root.Box.X; x < root.Box.X+root.Box.Width; x++ {
            cell := buffer.GetCell(x, y)
            if cell == nil || cell.Rune == ' ' {
                buffer.SetCell(x, y, paint.Cell{
                    Rune:  ' ',
                    Style: style.Style{BG: style.Color("black")},
                })
            }
        }
    }
}
```

### 4.7 API 设计

```go
// ui/layer.go

package ui

import (
    "github.com/wwsheng009/mint/runtime/layer"
)

// Layer 指定内容的渲染层级
func Layer(content VNode, l layer.Layer) VNode {
    if elem, ok := content.(*ElementVNode); ok {
        elem.SetLayer(l)
        return content
    }
    // 对于其他类型，使用 props
    return Element("layer").
        Prop("_layer", l).
        Child(content).
        Build()
}

// Modal 创建模态框
func Modal(id string, content VNode) VNode {
    return Layer(content, layer.LayerModal)
}

// Overlay 创建覆盖层
func Overlay(id string, content VNode) VNode {
    return Layer(content, layer.LayerOverlay)
}

// Tooltip 创建提示框
func Tooltip(id string, content VNode) VNode {
    return Layer(content, layer.LayerTooltip)
}

// ModalBuilder 模态框构建器
type ModalBuilder struct {
    node    VNode
    onClose func()
}

// ModalBuilder 创建模态框构建器
func ModalBuilder() *ModalBuilder {
    return &ModalBuilder{
        node: NewElement("modal"),
    }
}

func (b *ModalBuilder) Title(title string) *ModalBuilder {
    b.node.SetProp("title", title)
    return b
}

func (b *ModalBuilder) Content(content VNode) *ModalBuilder {
    b.node.SetChildren([]VNode{content})
    return b
}

func (b *ModalBuilder) OnClose(fn func()) *ModalBuilder {
    b.node.SetProp("onClose", fn)
    return b
}

func (b *ModalBuilder) Build() VNode {
    return Layer(b.node, layer.LayerModal)
}
```

---

## 五、事件系统扩展

### 5.1 事件阻断

```go
// runtime/event/dispatch.go

// Dispatcher 事件分发器
type Dispatcher struct {
    layerManager *layer.Manager
}

// DispatchEvent 分发事件
func (d *Dispatcher) DispatchEvent(e Event) bool {
    // 检查是否有 Modal 激活
    if d.layerManager.HasModal() {
        // 只分发事件到 Modal 层
        return d.dispatchToLayer(e, layer.LayerModal)
    }

    // 正常事件分发流程
    return d.dispatchToLayer(e, layer.LayerBase)
}

func (d *Dispatcher) dispatchToLayer(e Event, l Layer) bool {
    // 实现层级内事件分发...
    return true
}
```

### 5.2 ESC 键处理

```go
// framework/app.go

// handleKeyEvent 处理键盘事件
func (a *App) handleKeyEvent(e event.Event) bool {
    ke, ok := e.(*event.EventStruct).Key
    if !ok {
        return false
    }

    // ESC 键处理
    if ke.Key == '\x1b' { // ESC
        // 检查是否有 Modal
        if a.layerManager.HasModal() {
            // 关闭顶层 Modal
            a.closeTopModal()
            return true // 阻止事件继续传播
        }
    }

    // Tab 键处理（尊重 FocusTrap）
    if ke.Key == '\t' {
        if a.focusManager.HasActiveFocusTrap() {
            if ke.Shift {
                a.focusManager.FocusPrev()
            } else {
                a.focusManager.FocusNext()
            }
            return true
        }
    }

    return false
}
```

---

## 六、实施计划

### 阶段 1: 基础设施 (1-2天)

- [ ] 创建 `runtime/layer/types.go` - Layer 类型定义
- [ ] 创建 `runtime/layer/collector.go` - Layer 收集器
- [ ] 扩展 `runtime/ui/vnode.go` - 添加 SetLayer/GetLayer

### 阶段 2: Manager 实现 (2-3天)

- [ ] 创建 `runtime/layer/manager.go` - LayerManager
- [ ] 扩展 `runtime/compute/engine.go` - 支持多 Layer 布局
- [ ] 扩展 `internal/render/paint_engine.go` - 多 Layer 绘制

### 阶段 3: Pipeline 集成 (2-3天)

- [ ] 修改 `internal/render/rendering_pipeline.go`
- [ ] 在 Reconciler 中集成 LayerCollector
- [ ] 更新 Layout 阶段支持多 Layer
- [ ] 更新 Paint 阶段支持多 Layer

### 阶段 4: API 实现 (1-2天)

- [ ] 创建 `ui/layer.go` - 公开 API
- [ ] 实现 `ui.Modal()`
- [ ] 实现 `ui.Overlay()`
- [ ] 实现 `ui.Tooltip()`
- [ ] 实现 `ui.ModalBuilder()`

### 阶段 5: 事件系统 (1-2天)

- [ ] 实现 Modal 事件阻断
- [ ] 实现 ESC 键关闭
- [ ] 集成 FocusTrap 到 Layer 系统

### 阶段 6: 动画系统 (2-3天，可选)

- [ ] 设计动画状态机
- [ ] 实现淡入淡出效果
- [ ] 实现 Modal 缩放动画

---

## 七、临时修复方案

在完整 Layer 系统实现之前，使用现有能力修复 demo Modal：

### 7.1 当前问题

```go
// 问题代码
if showModal {
    return ui.VStackBuilder(
        mainContent,      // ❌ 被压缩
        ConfirmModal(...),
    ).Stretch().Build()
}
```

### 7.2 修复方案

```go
// 方案 1: 使用绝对定位 (推荐)
// 利用现有布局系统，让 Modal 脱离文档流

func App() ui.VNode {
    showModal, setShowModal := ui.UseStateBool(false)

    // 主内容始终独立
    mainContent := ui.VStackBuilder(
        Header(...),
        MainBody(...),
    ).Stretch().Build()

    // Modal 作为独立的返回值，不混合
    if showModal {
        return ModalContent(setShowModal)
    }

    return mainContent
}

func ModalContent(setShowModal func(bool)) ui.VNode {
    // Modal 全屏渲染，内部居中
    return ui.Box().
        Width(80).
        Height(24).
        Child(
            // 背景遮罩
            ui.Box().BgColor("black").Child(
                // 居中的 Modal 框
                ui.VStack(
                    ui.Text(""),
                    ui.Text("       *** Are you sure? ***"),
                    ui.Text(""),
                    ui.HStack(
                        ui.Text("       "),
                        ui.Button("[Cancel]").OnClick(func() { setShowModal(false) }),
                        ui.Text(" "),
                        ui.Button("[OK]").OnClick(func() { setShowModal(false) }),
                    ),
                    ui.Text(""),
                    ui.Text("Press ESC to close").FgColor("gray"),
                ),
            ),
        ).
        Build()
}
```

### 7.3 需要添加的功能

```go
// runtime/focus/manager.go

// SetModalMode 设置/取消 Modal 模式
// 当 Modal 模式激活时：
// - Tab 键只在 Modal 内循环
// - 背景组件无法获得焦点
func (fm *Manager) SetModalMode(modalNode *runtime.LayoutNode, active bool) {
    if active {
        // 创建 FocusTrap
        trap := focus.NewFocusTrap("modal", focus.TrapModal, modalNode)
        fm.PushFocusTrap(trap)
    } else {
        // 移除 FocusTrap
        fm.PopFocusTrap()
    }
}
```

---

## 八、测试策略

### 8.1 单元测试

```go
// runtime/layer/collector_test.go

func TestCollectorCollectModal(t *testing.T) {
    vnode := ui.VStack(
        ui.Text("Background"),
        ui.Layer(ui.Text("Modal"), layer.LayerModal),
    )

    collector := layer.NewCollector()
    collector.Collect(vnode)

    assert.True(t, collector.HasModal())
}

func TestLayerStrip(t *testing.T) {
    vnode := ui.VStack(
        ui.Text("Base"),
        ui.Layer(ui.Text("Overlay"), layer.LayerOverlay),
    )

    manager := layer.NewManager()
    baseTree := manager.stripLayers(vnode)

    // Base 树不应该包含 Overlay 节点
    assert.Equal(t, 1, len(baseTree.Children()))
}
```

### 8.2 集成测试

```go
// internal/render/pipeline_test.go

func TestLayerRendering(t *testing.T) {
    app := ui.VStack(
        ui.Text("Main Content"),
        ui.Modal("test", ui.Text("Hello Modal")),
    )

    pipeline := render.NewRenderingPipeline()
    buffer := paint.NewBuffer(80, 24)

    err := pipeline.Render(app, runtime.BoxConstraints{
        MinWidth: 0, MaxWidth: 80,
        MinHeight: 0, MaxHeight: 24,
    }, buffer)

    assert.NoError(t, err)
    // 验证 Modal 在正确位置
}
```

---

## 九、优先级排序

| 阶段 | 优先级 | 预计时间 | 价值 |
|------|--------|----------|------|
| 临时修复 (demo1) | P0 | 1小时 | 立即改善 |
| 基础设施 | P1 | 1-2天 | 必需 |
| Manager | P1 | 2-3天 | 核心 |
| Pipeline 集成 | P1 | 2-3天 | 必需 |
| API | P2 | 1-2天 | 易用性 |
| 事件系统 | P2 | 1-2天 | 完整性 |
| 动画 | P3 | 2-3天 | 增强体验 |

---

## 十、风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Fiber 兼容性 | 高 | 充分测试 Fiber 遍历 |
| 性能回归 | 中 | 缓存 Layer 结果 |
| 布局复杂性 | 高 | 分阶段实现，先简单后复杂 |
| 事件系统耦合 | 中 | 使用接口隔离 |

---

## 十一、相关文档

- [Flex Layout](../layout/flex_layout.md)
- [Stretch Layout](../layout/stretch_layout.md)
- [Layout Refactor](../layout/layout_refactor.md)
- [Rendering Pipeline](../layout/LAYOUT_RENDERING_REFACTOR.md)

---

**文档版本**: 1.0
**最后更新**: 2025-02-06
