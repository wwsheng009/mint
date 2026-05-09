# Layer 系统完整指南

> 当前 API 提示：Layer 概念仍有参考价值，但部分示例使用旧组件 builder 和回调式处理器。当前交互组件应通过 `OnPress`、`ForField` 或组件专用 Intent builder 方法发出 `intent.Intent`。

## 目录

1. [概述](#一概述)
2. [Layer 运行机制](#二layer-运行机制)
3. [Layer 与 Layout 的配合](#三layer-与-layout-的配合)
4. [常见问题及原因分析](#四常见问题及原因分析)
5. [如何规避问题](#五如何规避问题)
6. [关键代码说明](#六关键代码说明)
7. [Layer 使用指南](#七layer-使用指南)
8. [调试技巧](#八调试技巧)

---

## 一、概述

### 1.1 什么是 Layer 系统

Layer 系统是 Mint TUI 框架中用于管理多层 UI 组件（如模态框、下拉菜单、工具提示）的子系统。它允许开发者声明式地创建浮动在主内容之上的 UI 元素，而无需手动管理坐标和层级关系。

### 1.2 Layer 类型

```go
type Layer int

const (
    LayerBase    Layer = iota  // 0: 基础内容层
    LayerOverlay              // 1: 覆盖层（下拉菜单、弹出框）
    LayerModal                // 2: 模态框层
    LayerTooltip              // 3: 工具提示层
)
```

**层级关系**（从底到顶）：
```
┌─────────────────────────────────────┐
│        LayerTooltip (3)             │  ← 最顶层
├─────────────────────────────────────┤
│        LayerModal (2)               │  ← 模态框
├─────────────────────────────────────┤
│        LayerOverlay (1)             │  ← 下拉菜单
├─────────────────────────────────────┤
│        LayerBase (0)                │  ← 主内容
└─────────────────────────────────────┘
```

---

## 二、Layer 运行机制

### 2.1 核心组件

```
┌─────────────────────────────────────────────────────────────────┐
│                        LayerManager                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐    ┌───────────────┐    ┌────────────────┐  │
│  │  Collector   │───▶│  Layout Tree  │───▶│  LayerLayouts  │  │
│  │              │    │               │    │  (map[Layer]   │  │
│  │ - Collect()  │    │ - baseTree   │    │   →Layout)     │  │
│  │ - StripLayers│    │ - modalNode  │    │                │  │
│  └──────────────┘    └───────────────┘    └────────────────┘  │
│                            │                    │               │
│                            ▼                    ▼               │
│                    ┌──────────────────────────────────────┐   │
│                    │         ComputeEngine                 │   │
│                    │  - Layout baseTree                   │   │
│                    │  - Layout modalNode (独立布局)        │   │
│                    └──────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 工作流程

```go
// 1. 收集阶段 - Collector.Collect()
func (m *Manager) CollectAndLayout(vnode VNode, constraints BoxConstraints, engine *Engine) error {
    // 步骤1: 清空之前的收集结果
    m.layouts = make(LayerLayouts)

    // 步骤2: 遍历 VNode 树，收集 layer 节点
    m.collector.Collect(vnode)
    // 结果: layers[LayerModal] = []*LayerNode{...}

    // 步骤3: 从主树中移除 layer 节点，得到纯净的基础树
    baseTree := m.collector.StripLayers(vnode)

    // 步骤4: 布局基础层
    baseLayout, err := engine.Layout(baseTree, constraints)
    m.layouts[LayerBase] = baseLayout

    // 步骤5: 布局每个收集到的 layer
    for layer, nodes := range m.collector.GetLayers() {
        for _, node := range nodes {
            layerLayout, err := m.layoutLayer(node, layer, constraints, engine)
            m.layouts[layer] = layerLayout
        }
    }

    return nil
}
```

### 2.3 收集机制详解

```go
// Collector.walk() 递归遍历 VNode 树
func (c *Collector) walk(vnode VNode) {
    // 检查节点是否设置了非基础层
    if layer := vnode.GetLayer(); layer != LayerBase && layer.IsValid() {
        // 创建 LayerNode 并添加到对应 layer
        node := &LayerNode{
            Layer:   layer,
            ID:      c.getNodeID(vnode),
            Content: vnode,
            Visible: c.isVisible(vnode),
        }
        c.layers.Add(layer, node)

        // 关键: 不再遍历 layer 节点的子节点
        // 它们将在独立的布局过程中被处理
        return
    }

    // 递归遍历子节点
    for _, child := range vnode.Children() {
        c.walk(child)
    }
}
```

### 2.4 布局机制详解

```go
// 不同 layer 类型有不同的约束策略
func (m *Manager) layoutLayer(node *LayerNode, layer Layer, constraints BoxConstraints, engine *Engine) (*ComputedLayout, error) {
    var layerConstraints BoxConstraints

    switch layer {
    case LayerModal:
        // Modal: 全屏约束，之后会居中
        layerConstraints = BoxConstraints{
            MinWidth:  0,
            MaxWidth:  constraints.MaxWidth,
            MinHeight: 0,
            MaxHeight: constraints.MaxHeight,
        }

    case LayerOverlay:
        // Overlay: 根据锚点定位
        layerConstraints = BoxConstraints{
            MinWidth:  0,
            MaxWidth:  constraints.MaxWidth,
            MinHeight: 0,
            MaxHeight: constraints.MaxHeight,
        }

    case LayerTooltip:
        // Tooltip: 小尺寸，自定位
        layerConstraints = BoxConstraints{
            MinWidth:  0,
            MaxWidth:  constraints.MaxWidth,
            MinHeight: 0,
            MaxHeight: constraints.MaxHeight,
        }

    default:
        layerConstraints = constraints
    }

    // 执行布局
    layout, err := engine.Layout(node.Content, layerConstraints)

    // Modal 特殊处理: 居中
    if layer == LayerModal && layout.Root != nil {
        m.centerModal(layout.Root, constraints)
    }

    return layout, nil
}
```

### 2.5 居中机制

```go
func (m *Manager) centerModal(root *ComputedBox, constraints BoxConstraints) {
    modalWidth := root.Box.Width
    modalHeight := root.Box.Height

    containerWidth := constraints.MaxWidth
    containerHeight := constraints.MaxHeight

    // 计算居中偏移
    offsetX := (containerWidth - modalWidth) / 2
    offsetY := (containerHeight - modalHeight) / 2

    // 递归移动整个布局树的所有节点
    m.shiftPositions(root, offsetX, offsetY)
}

func (m *Manager) shiftPositions(box *ComputedBox, offsetX, offsetY int) {
    box.Box.X += offsetX
    box.Box.Y += offsetY

    for _, child := range box.Children {
        m.shiftPositions(child, offsetX, offsetY)
    }
}
```

---

## 三、Layer 与 Layout 的配合

### 3.1 约束传递流程

```
┌──────────────────────────────────────────────────────────────┐
│                      约束传递流程                              │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  App() VNode Tree                                            │
│      │                                                       │
│      ▼                                                       │
│  ┌─────────────────────────────────────────┐                │
│  │   VStack (基础内容)                      │                │
│  │   └── Text("Background")                │                │
│  │   └── Modal(Bordered(...))  ← LayerModal │                │
│  └─────────────────────────────────────────┘                │
│      │                                                       │
│      ▼                                                       │
│  ┌─────────────────────────────────────────┐                │
│  │   Collector.Collect()                   │                │
│   ──────────────────────────────────────── │                │
│   baseTree:             modalNode:         │                │
│   ┌── VStack           ┌── Bordered       │                │
│   │   └── Text        │       └── VStack  │                │
│   └── (modal removed)           └── ...    │                │
│  └─────────────────────────────────────────┘                │
│      │                    │                                 │
│      ▼                    ▼                                 │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ baseTree Layout │  │ modal Layout    │                   │
│  │ constraints:    │  │ constraints:    │                   │
│  │ {0,80, 0,24}    │  │ {0,80, 0,24}    │                   │
│  └─────────────────┘  └─────────────────┘                   │
│                              │                                │
│                              ▼                                │
│                    ┌─────────────────┐                       │
│                    │  centerModal()  │                       │
│                    │  shiftPosition  │                       │
│                    └─────────────────┘                       │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 3.2 独立布局的必要性

Layer 节点需要独立布局的原因：

1. **约束隔离**：Modal 可能需要不同的约束（如全屏）
2. **位置独立**：Modal 需要居中，与基础内容无关
3. **事件隔离**：Modal 事件不应影响基础内容
4. **渲染顺序**：Layer 需要在基础内容之后渲染

### 3.3 渲染顺序

```go
// PaintEngine.PaintLayers() 按固定顺序渲染
func (e *PaintEngine) PaintLayers(layouts LayerLayouts, buffer Buffer) {
    order := []Layer{
        LayerBase,      // 先渲染基础内容
        LayerOverlay,   // 再渲染覆盖层
        LayerModal,     // 然后渲染模态框
        LayerTooltip,   // 最后渲染工具提示
    }

    for _, layer := range order {
        if layout, ok := layouts[layer]; ok && layout.Root != nil {
            e.Paint(layout, buffer)

            // Modal 特殊处理: 绘制半透明背景
            if layer == LayerModal {
                e.paintModalBackdrop(layout.Root.Box, buffer)
            }
        }
    }
}
```

---

## 四、常见问题及原因分析

### 4.1 问题分类

| 问题类型 | 症状 | 根本原因 |
|---------|------|---------|
| 布局问题 | Modal 内容不居中 | 约束未正确传递到 HStack 子节点 |
| 尺寸问题 | Modal 宽度不正确 | Bordered 节点未创建 tight 约束 |
| 事件问题 | ESC 无法关闭 Modal | 事件处理器未正确集成 |
| 渲染问题 | Modal 内容闪烁 | 多次布局导致状态不一致 |

### 4.2 深度问题分析

#### 问题 1: Modal 内容不居中

**现象**：
```
│*** Title ***                         │  ← 标题偏左
│*[ [ Cancel ] ]                [ [ OK ] ]│  ← 按钮分散
```

**原因链**：
```
1. VStack 测量阶段
   ├─ 正确创建 tight 约束 {MinWidth:38, MaxWidth:38}
   └─ HStack.Measure() 返回 width=38 ✅

2. buildComputedBox() 阶段
   ├─ getChildConstraints() 被调用
   ├─ 对于 VStack 的 HStack 子节点
   ├─ 返回 {MinWidth:0, MaxWidth:38}  ❌ 不是 tight!
   └─ HStack 被重新测量，返回 width=13 ❌

3. layoutHStack() 定位阶段
   ├─ box.Width=13 (错误，应该是 38)
   └─ 居中计算基于错误的宽度
```

**根本原因**：
`getChildConstraints()` 没有为 VStack 中的 HStack 子节点创建 tight 约束，导致 HStack 在第二次测量时使用非 tight 约束，返回错误的宽度。

**修复**：
```go
// runtime/layout/layout_engine.go - getChildConstraints()
case "vstack":
    // ...
    // Non-flex child
    childMinWidth := 0
    if childMaxWidth != Infinity && isHStack(child) {
        childMinWidth = childMaxWidth // HStack 填充 VStack 宽度以支持对齐
    }
    return BoxConstraints{
        MinWidth:  childMinWidth,  // 关键修复
        MaxWidth:  childMaxWidth,
        MinHeight: 0,
        MaxHeight: Infinity,
    }
```

#### 问题 2: Modal 尺寸异常

**现象**：
- Modal 宽度为 0
- 或者 Modal 宽度超出屏幕
- `Width(40)` 不生效，显示为全屏宽度

**原因**：
```go
// 问题1: BorderedBuilder.Width() 设置的位置
func (b *BorderedBuilder) Width(n int) *BorderedBuilder {
    b.node.SetProp("width", n)  // 存储在 Props
    return b
}

// 问题2: BorderedNode.Measure() 检查的位置
func (bn *BorderedNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // ...
    elemStyle := bn.Style()
    if elemStyle.Width > 0 {  // ❌ 只检查 Style，不检查 Props!
        totalWidth = elemStyle.Width
    }
}
```

**修复 (2025-02-07)**：
遵循"layout不参与计算"的设计原则，由Engine负责处理Props并创建约束：

```go
// runtime/layout/layout_engine.go - measureVNode()
func (e *Engine) measureVNode(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
    // SPECIAL CASE: Bordered nodes use Engine's measureBordered
    if isBordered(vnode) {
        return e.measureBordered(vnode, constraints)  // ✅
    }
    // ...
}

// runtime/layout/layout_engine.go - measureBordered()
func (e *Engine) measureBordered(vnode VNode, constraints runtime.BoxConstraints) runtime.Size {
    // 检查 Props["width"]
    if props := vnode.Props(); props != nil {
        if w, ok := props["width"].(int); ok && w > 0 {
            // 创建 tight 约束
            innerConstraints = NewBoxConstraints(
                w-2,  // MinWidth = MaxWidth (tight)
                w-2,
                0,
                innerConstraints.MaxHeight,
            )
        }
    }
    // ...
}
```

**设计原则**：
| 组件 | 职责 |
|------|------|
| Engine | 读取Props，创建tight约束 |
| BorderedNode | 只根据传入的constraints计算尺寸 |

#### 问题 3: 约束传播失败

**现象**：
```
Invalid constraints: MinWidth=40 > MaxWidth=38
```

**原因**：
```go
// SubtractPadding 只减少了 MaxWidth，没减少 MinWidth
func (bc BoxConstraints) SubtractPadding(horizontal, vertical int) BoxConstraints {
    result := bc
    if bc.MaxWidth > horizontal {
        result.MaxWidth = bc.MaxWidth - horizontal  // ✅
    }
    // MinWidth 未被修改 ❌
    return result
}
```

修复后：
```go
func (bc BoxConstraints) SubtractPadding(horizontal, vertical int) BoxConstraints {
    result := bc
    if bc.MaxWidth > horizontal {
        result.MaxWidth = bc.MaxWidth - horizontal
    }
    if bc.MinWidth > horizontal {
        result.MinWidth = bc.MinWidth - horizontal  // ✅ 修复
    } else {
        result.MinWidth = 0
    }
    // 确保 MinWidth <= MaxWidth
    if result.MinWidth > result.MaxWidth {
        result.MinWidth = result.MaxWidth
    }
    return result
}
```

### 4.3 架构层面的问题

#### 问题: centerModal() 的后处理设计

```go
// 当前实现: Layout 后修改坐标
func (m *Manager) layoutLayer(...) {
    layout, _ := engine.Layout(node.Content, layerConstraints)
    if layer == LayerModal {
        m.centerModal(layout.Root, constraints)  // ← 后处理
    }
    return layout
}
```

**问题**：
1. ComputeEngine.Layout() 完成两阶段布局后，坐标已确定
2. centerModal() 直接修改这些坐标
3. 破坏了 Layout 系统的两阶段设计原则

**影响**：
- calculatePositions() 的结果被覆盖
- 子节点的相对坐标可能不一致
- 难以调试和维护

**建议的改进方向**：
1. 在 Layout 阶段前就确定 Modal 位置
2. 让 Layout 系统支持"位置约束"
3. 或者明确 centerModal() 作为 Layout 阶段的一部分

---

## 五、如何规避问题

### 5.1 使用 Layer 的最佳实践

#### 规则 1: 确保 HStack 子节点收到 tight 约束

```go
// ❌ 错误: VStack 直接测量，未考虑 HStack 对齐需求
ui.VStack(
    ui.HStack(  // 这个 HStack 可能收到 {MinWidth:0, MaxWidth:38}
        ui.Text("Title"),
    ).Align(ui.AlignCenter),
)

// ✅ 正确: 使用 Bordered 确保 tight 约束
ui.Bordered().Width(40).Child(
    ui.VStack(
        ui.HStack(  // 现在会收到 {MinWidth:38, MaxWidth:38}
            ui.Text("Title"),
        ).Align(ui.AlignCenter),
    ),
).Build()
```

#### 规则 2: 避免嵌套 Layer 节点

```go
// ❌ 错误: Modal 内嵌套另一个 Modal
ui.Modal(
    ui.VStack(
        ui.Text("First Modal"),
        ui.Modal(  // 嵌套的 Modal
            ui.Text("Second Modal"),
        ).Build(),
    ),
).Build()

// ✅ 正确: 通过状态控制显示哪个 Modal
func App() ui.VNode {
    showModal1, setShowModal1 := ui.UseStateBool(false)
    showModal2, setShowModal2 := ui.UseStateBool(false)

    content := ui.VStack(
        ui.Text("Main Content"),
    )

    if showModal2 {
        content = ui.VStack(content, Modal2(...))
    } else if showModal1 {
        content = ui.VStack(content, Modal1(...))
    }

    return content
}
```

#### 规则 3: 明确设置 Modal 尺寸

```go
// ❌ 模糊: Modal 依赖内容尺寸
ui.Modal(
    ui.VStack(
        ui.Text("Content that might be too wide"),
    ),
)

// ✅ 清晰: 明确设置宽度和最大宽度
ui.Modal(
    ui.Bordered().
        Width(60).           // 固定宽度
        MaxWidth(80).        // 最大宽度限制
        Child(
            ui.VStack(
                ui.Text("Content"),
            ),
        ).
        Build(),
).Build()
```

#### 规则 4: 使用 Align 而非手动添加空格

```go
// ❌ 错误: 手动添加空格来"居中"
ui.HStack(
    ui.Text("      "),
    ui.Text("Title"),
    ui.Text("      "),
)

// ✅ 正确: 使用 Align
ui.HStackBuilder(
    ui.Text("Title"),
).Align(ui.AlignCenter).Build()
```

### 5.2 调试技巧

#### 启用调试输出

```bash
# Layer 系统调试
export TUI_LAYER_DEBUG=true

# Layout 系统调试
export TUI_ALIGN_DEBUG=true

# 运行应用
go run main.go
```

#### 检查约束传播

在关键位置添加断点或日志：

```go
// 在 VStack.Measure() 中
if os.Getenv("TUI_ALIGN_DEBUG") == "true" {
    fmt.Fprintf(os.Stderr, "[VStack.Measure] constraints=%+v, tag=%s\n",
        constraints, l.tag)
}
```

#### 验证布局结果

```go
// 在 layoutHStack() 中
if os.Getenv("TUI_ALIGN_DEBUG") == "true" {
    fmt.Fprintf(os.Stderr, "[layoutHStack] mainAlign=%d, box.Width=%d, totalChildWidth=%d\n",
        mainAlign, box.Box.Width, totalChildWidth)
}
```

### 5.3 常见模式

#### 模式 1: 带 Title 和 Buttons 的 Modal

```go
func ConfirmModal(onConfirm, onCancel func()) ui.VNode {
    return ui.Modal(
        ui.Bordered().
            Width(50).  // 固定宽度
            Color("yellow").
            Child(
                ui.VStackBuilder(
                    // 顶部空行
                    ui.Text(""),

                    // 居中标题
                    ui.HStackBuilder(
                        ui.Text("*** Confirm Action ***"),
                    ).Align(ui.AlignCenter).FgColor("yellow").Build(),

                    ui.Text(""),
                    ui.Text(""),

                    // 居中按钮
                    ui.HStackBuilder(
                        app.ButtonBuilder("[ Cancel ]").
                            OnClick(onCancel).
                            Build(),
                        ui.Text("  "),
                        app.ButtonBuilder("[ Confirm ]").
                            BgColor("red").
                            FgColor("white").
                            OnClick(onConfirm).
                            Build(),
                    ).Align(ui.AlignCenter).Build(),

                    ui.Text(""),
                ).Build(),
            ).
            Build(),
    ).
    CloseOnESC(true).
    CloseOnBackdropClick(true).
    OnClose(onCancel).
    Build()
}
```

#### 模式 2: 条件性显示 Modal

```go
func App() ui.VNode {
    showConfirm, setShowConfirm := ui.UseStateBool(false)

    mainContent := ui.VStack(
        ui.Text("Main Content"),
        app.ButtonBuilder("Delete").
            BgColor("red").
            OnClick(func() {
                setShowConfirm(true)
            }).
            Build(),
    )

    // 如果需要确认，同时渲染主内容和 Modal
    // LayerManager 会将它们分离到不同的层
    if showConfirm {
        return ui.VStack(
            mainContent,
            ConfirmModal(
                func() { /* 确认删除 */ },
                func() { setShowConfirm(false) },
            ),
        )
    }

    return mainContent
}
```

---

## 六、关键代码说明

### 6.1 核心文件结构

```
runtime/layer/
├── manager.go       # LayerManager - 协调层收集和布局
├── collector.go     # Collector - 收集 layer 节点
└── event_handler.go # EventHandler - 处理层特定事件

ui/layer.go          # Layer API - Modal, Overlay, Tooltip 构建器

runtime/layout/layout_engine.go  # ComputeEngine - 布局计算引擎
runtime/ui/layout.go       # LayoutNode - 布局节点实现
```

### 6.2 关键函数

#### Collector.Collect()

```go
// internal/render/rendering_pipeline.go
func (c *Collector) Collect(vnode VNode) {
    c.layers.Clear()      // 清空之前的收集
    c.walk(vnode)         // 递归遍历
}
```

#### Collector.StripLayers()

```go
// internal/render/rendering_pipeline.go
func (c *Collector) StripLayers(vnode VNode) VNode {
    // 返回移除了 layer 节点的基础树
    // 这是纯净的基础内容，不包含任何 layer 节点
}
```

#### LayerManager.layoutLayer()

```go
// runtime/layout/layer_manager.go
func (m *Manager) layoutLayer(
    node *LayerNode,
    layer Layer,
    constraints BoxConstraints,
    engine *Engine,
) (*ComputedLayout, error) {
    // 1. 根据 layer 类型确定约束
    // 2. 调用 engine.Layout()
    // 3. 如果是 modal，调用 centerModal()
    // 4. 返回 ComputedLayout
}
```

#### ComputeEngine.getChildConstraints()

```go
// runtime/layout/layout_engine.go
func (e *Engine) getChildConstraints(
    parent, child VNode,
    parentConstraints BoxConstraints,
    parentSize Size,
) BoxConstraints {
    // 根据 parent 类型计算 child 应该收到的约束
    // 关键修复: VStack 的 HStack 子节点需要 tight 约束
}
```

### 6.3 关键数据结构

```go
// LayerNode - 表示一个 layer 中的节点
type LayerNode struct {
    Layer   Layer          // 层级
    ID      string         // 唯一标识
    Content VNode          // 实际内容
    Visible bool           // 是否可见
    FocusID string         // 焦点组件 ID
}

// LayerLayouts - 每层的布局结果
type LayerLayouts map[Layer]*ComputedLayout

// BoxConstraints - 布局约束
type BoxConstraints struct {
    MinWidth, MaxWidth     int
    MinHeight, MaxHeight   int
}
```

---

## 七、Layer 使用指南

### 7.1 Modal 使用

#### 基础用法

```go
import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/app"
)

func SimpleModal() ui.VNode {
    return ui.Modal(
        ui.Bordered().
            Width(40).
            Child(
                ui.VStack(
                    ui.Text("Modal Content"),
                    app.ButtonBuilder("Close").Build(),
                ),
            ).
            Build(),
    ).Build()
}
```

#### 完整配置

```go
func FullFeaturedModal(onClose func()) ui.VNode {
    return ui.Modal(
        ui.Bordered().
            Width(50).
            Color("cyan").
            Child(
                ui.VStackBuilder(
                    ui.Text(""),
                    ui.HStackBuilder(
                        app.NewTextBuilder("Title").
                            Bold(true).
                            Build(),
                    ).Align(ui.AlignCenter).Build(),
                    ui.Text(""),
                    ui.Text("Modal content goes here..."),
                    ui.Text(""),
                    ui.HStackBuilder(
                        app.ButtonBuilder("[ Cancel ]").
                            OnClick(onClose).
                            Build(),
                        ui.Text(" "),
                        app.ButtonBuilder("[ OK ]").
                            BgColor("green").
                            OnClick(onClose).
                            Build(),
                    ).Align(ui.AlignCenter).Build(),
                    ui.Text(""),
                ).Build(),
            ).
            Build(),
    ).
    OnClose(onClose).              // 关闭回调
    CloseOnESC(true).              // ESC 键关闭
    CloseOnBackdropClick(true).    // 点击背景关闭
    Centered(true).                // 居中显示
    Build()
}
```

### 7.2 Overlay 使用

```go
func DropdownMenu() ui.VNode {
    trigger := app.ButtonBuilder("Menu ▼").Build()

    menu := ui.Overlay(
        ui.Bordered().
            Child(
                ui.VStack(
                    app.ButtonBuilder("Option 1").Build(),
                    app.ButtonBuilder("Option 2").Build(),
                    app.ButtonBuilder("Option 3").Build(),
                ),
            ).
            Build(),
    ).
    Anchor(trigger).          // 锚定到触发按钮
    Position(OverlayPositionBottom).
    Offset(0, 1).
    Build()

    return ui.HStack(trigger, menu)
}
```

### 7.3 Tooltip 使用

```go
func ButtonWithTooltip() ui.VNode {
    button := app.ButtonBuilder("Help").Build()

    tooltip := ui.Toololtip("Press this button for help").
        Anchor(button).
        Delay(500).  // 延迟 500ms 显示
        Build()

    return ui.HStack(button, tooltip)
}
```

### 7.4 状态管理

```go
func AppWithModal() ui.VNode {
    // 使用 State 管理 Modal 显示状态
    showModal, setShowModal := ui.UseStateBool(false)

    // 构建主内容
    mainContent := ui.VStack(
        ui.Text("Main Application"),
        app.ButtonBuilder("Open Modal").
            OnClick(func() {
                setShowModal(true)
            }).
            Build(),
    )

    // 条件渲染: 如果 showModal 为 true，同时渲染主内容和 Modal
    if showModal {
        return ui.VStack(
            mainContent,
            // Modal 会自动被分离到独立的层
            FullFeaturedModal(func() {
                setShowModal(false)
            }),
        )
    }

    return mainContent
}
```

---

## 八、调试技巧

### 8.1 环境变量

| 变量 | 作用 |
|-----|------|
| `TUI_LAYER_DEBUG` | 启用 Layer 系统调试输出 |
| `TUI_ALIGN_DEBUG` | 启用布局对齐调试输出 |
| `TUI_LAYOUT_DEBUG` | 启用通用布局调试输出 |

### 8.2 常见调试场景

#### 场景 1: Modal 不显示

```bash
# 1. 启用 Layer 调试
TUI_LAYER_DEBUG=true go run main.go

# 2. 检查输出中的:
#    - [CollectAndLayout] collected X modal nodes
#    - 如果是 0，说明 Modal 节点未被正确收集
```

#### 场景 2: Modal 内容不居中

```bash
# 1. 启用对齐调试
TUI_ALIGN_DEBUG=true go run main.go

# 2. 检查输出中的:
#    - [HStack.Measure] constraints=..., IsTight=?
#    - [layoutHStack] mainAlign=?, box.Width=?
#    - IsTight 应该为 true
#    - box.Width 应该等于父容器宽度
```

#### 场景 3: 约束传播问题

```go
// 在代码中添加临时调试
func (l *LayoutNode) Measure(constraints BoxConstraints) Size {
    fmt.Fprintf(os.Stderr, "[Measure] tag=%s, constraints=%+v\n",
        l.tag, constraints)
    // ... 原有代码
}
```

### 8.3 测试工具

```go
// 使用 RunTest 进行单元测试
func TestModalCentering(t *testing.T) {
    app := func() ui.VNode {
        return ui.Modal(
            ui.Bordered().Width(40).Child(
                ui.HStack(
                    ui.Text("Title"),
                ).Align(ui.AlignCenter),
            ).Build(),
        ).Build()
    }

    testApp, err := ui.RunTest(app, ui.WithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    rendered := testApp.GetRenderString()
    // 验证渲染结果
}
```

---

## 总结

### 核心要点

1. **Layer 系统通过分离和独立布局实现多层 UI**
   - Collector 收集 layer 节点
   - StripLayers 创建纯净的基础树
   - 每个 layer 独立布局

2. **约束正确传递是居中的关键**
   - VStack 的 HStack 子节点需要 tight 约束
   - getChildConstraints() 需要特殊处理 HStack

3. **遵循最佳实践避免问题**
   - 明确设置 Modal 尺寸
   - 使用 Align 而非手动空格
   - 避免嵌套 Layer 节点

4. **调试工具帮助快速定位问题**
   - 环境变量控制调试输出
   - 检查约束传播路径
   - 验证布局结果

### 架构演进方向

1. **改进 centerModal 机制**
   - 考虑将居中逻辑集成到 Layout 阶段
   - 而非后处理修改坐标

2. **增强约束系统**
   - 支持"位置约束"类型
   - 允许更灵活的布局控制

3. **完善事件处理**
   - 集成 LayerEventHandler 到主循环
   - 实现完整的 focus trap

---

## 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2025-02-06 | 初始版本 |
| 1.1 | 2025-02-07 | 添加 Bordered.Width() 修复说明，明确 Engine/Node 职责分离 |

---

*文档维护者: Claude*
*最后更新: 2025-02-07*
