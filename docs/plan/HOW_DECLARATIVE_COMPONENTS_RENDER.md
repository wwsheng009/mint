# 声明式组件渲染流程详解

> **目标**: 解释从声明式组件代码到终端显示的完整渲染流程
> **日期**: 2026-02-01

---

## 目录

1. [渲染流程概览](#渲染流程概览)
2. [启动阶段](#启动阶段)
3. [渲染阶段](#渲染阶段)
4. [VNode 遍历与渲染](#vnode-遍历与渲染)
5. [组件示例](#组件示例)
6. [两种渲染模式](#两种渲染模式)

---

## 渲染流程概览

```
┌─────────────────────────────────────────────────────────────┐
│ 1. 用户代码层                                               │
│                                                            │
│   func Counter() ui.VNode {                                │
│       count, setCount := ui.UseState(0)                    │
│       return ui.VStack(                                   │
│           ui.Text("Count: 0"),                          │
│           ui.Button("+"),                                │
│       )                                                   │
│   }                                                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. API 入口层 (ui.Run)                                     │
│                                                            │
│   ui.Run(Counter, ui.WithWidth(40), ui.WithHeight(20))     │
│     ↓                                                      │
│   ├─→ framework.NewApp()                               │
│   ├─→ newDeclarativeRoot(Counter, fwApp)                 │
│   └─→ fwApp.SetRoot(declarativeRoot)                    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. 框架层 (framework.App)                                 │
│                                                            │
│   Event Loop:                                               │
│     │                                                      │
│     ↓                                                      │
│   root.Paint() ───────────────────────────┐                │
│       ↓                                        │                │
│   declarativeRoot.Paint(ctx, buffer)         │                │
└───────────────────────────────────────────┴────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. 声明式渲染层 (declarativeRoot)                        │
│                                                            │
│   ├─→ Fiber 模式?                                         │
│   │   ├─ Yes → paintWithFiber()                         │
│   │   └─→ No  → paintLegacy()                          │
│ │                                                        │
│ └────────────────────────────────────────────────┐          │
│ │                                                    │          │
│ │  paintLegacy() / paintWithFiber() 流程:          │          │
│ │                                                    │          │
│ │  ┌──────────────────────────────────────────┐  │          │
│ │  │ 1. 调用用户函数: vnode := d.appFn()        │  │          │
│ │  └──────────────────────────────────────────┘  │          │
│ │                    ↓                               │          │
│ │  ┌──────────────────────────────────────────┐  │          │
│ │  │ 2. 递归渲染: renderVNode(vnode, x, y)      │  │          │
│  │  │   - switch vnode.Type                      │  │          │
│  │  │   - 调用对应的 render 方法               │  │          │
│ │  │   - 递归处理子节点                       │  │          │
│  │  └──────────────────────────────────────────┘  │          │
│ │                    ↓                               │          │
│ │  ┌──────────────────────────────────────────┐  │          │
│  │  │ 3. 写入 Buffer                          │  │          │
│  │  └──────────────────────────────────────────┘  │          │
│ │                                                    │          │
│ └────────────────────────────────────────────┘  │          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. 运行时层 (runtime/paint)                                  │
│                                                            │
│   buffer.SetCell(x, y, char, style)                      │
│   buffer.SetString(x, y, text, style)                     │
│   buffer.Fill(rect, char, style)                           │
│                                                            │
│   ↓ 最终输出: 终端字符界面                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 启动阶段

### 用户代码

```go
// main.go
package main

import "github.com/wwsheng009/mint/ui"

func Counter() ui.VNode {
    count, setCount := ui.UseState(0)
    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("+").OnClick(func() { setCount(count + 1) }),
    )
}

func main() {
    ui.Run(Counter)
}
```

### ui.Run() 执行流程

```go
// ui/app.go

func Run(app ComponentFunc, opts ...Option) error {
    // 1. 创建配置
    options := &Options{Width: 80, Height: 24}

    // 2. 创建框架应用
    fwApp := framework.NewApp()
    fwApp.Resize(options.Width, options.Height)
    appInstance = fwApp

    // 3. 创建声明式根组件
    declarativeRoot := newDeclarativeRoot(app, fwApp)

    // 4. 设置为根组件
    fwApp.SetRoot(declarativeRoot)

    // 5. 运行应用 (进入事件循环)
    return fwApp.Run()
}
```

---

## 渲染阶段

### framework.App 事件循环

```
framework.App.Run()
    ↓
Event Loop:
    ├─ 等待输入 (键盘/鼠标/窗口)
    ├─ 收到输入 → 分发事件
    ├─ 状态改变 → MarkDirty()
    └─ 需要重绘 → 调用 root.Paint()
```

### declarativeRoot.Paint() 入口

```go
// ui/app.go

func (d *declarativeRoot) Paint(ctx component.PaintContext, buffer *paint.Buffer) {
    // 根据 MINT_USE_FIBER 环境变量选择渲染模式
    if d.reconciler != nil {
        d.paintWithFiber(ctx, buffer)  // Fiber 模式
        return
    }
    d.paintLegacy(ctx, buffer)         // 遗留模式
}
```

---

## VNode 遍历与渲染

### Legacy 模式渲染流程

```go
// ui/app.go - paintLegacy()

func (d *declarativeRoot) paintLegacy(ctx component.PaintContext, buffer *paint.Buffer) {
    // 1. 重置 Hook 索引，设置当前 Context
    d.ctx.resetContext()
    setCurrentContext(d.ctx)

    // 2. 调用用户函数，获取 VNode 树
    vnode := d.appFn()  // ← 这里调用 Counter()

    // 3. 清除 Context
    setCurrentContext(nil)

    // 4. 运行 Effects
    d.ctx.runEffects()

    // 5. 递归渲染 VNode 树到 Buffer
    d.renderVNode(vnode, ctx.X, ctx.Y, buffer)
}
```

### renderVNode 递归渲染

```go
// ui/app.go - renderVNode()

func (d *declarativeRoot) renderVNode(node VNode, x, y int, buffer *paint.Buffer) int {
    if node == nil {
        return 0
    }

    currentY := y

    // 根据节点类型分发渲染
    switch n := node.(type) {
    case *TextVNode:
        // 文本节点：渲染文本并返回高度
        d.renderText(n, x, currentY, buffer)
        currentY += 1

    case *ElementVNode:
        // 元素节点：递归渲染子节点
        for _, child := range n.Children() {
            offsetY := d.renderVNode(child, x, currentY, buffer)
            currentY += offsetY
        }

    case *LayoutNode:
        // 布局节点：计算位置后渲染子节点
        // ... HStack/VStack 的布局逻辑 ...

    case *ButtonVNode:
        // 按钮节点：收集焦点信息并渲染
        if !n.Disabled() {
            n.focusIndex = len(d.buttons)
            d.buttons = append(d.buttons, n)
        }
        d.renderButton(n, x, currentY, buffer)
        currentY += 1

    case *InputVNode:
        // 输入节点：收集焦点信息并渲染
        if !n.Disabled() && !n.ReadOnly() {
            n.focusIndex = len(d.inputs)
            d.inputs = append(d.inputs, n)
        }
        d.renderInput(n, x, currentY, buffer)
        currentY += 1

    // ... 其他节点类型 ...
    }

    return currentY
}
```

### 组件渲染示例

#### TextVNode 渲染

```go
// ui/text.go (简化)

type TextVNode struct {
    *ui.ElementVNode
    content string
}

// renderText 渲染文本到 Buffer
func (d *declarativeRoot) renderText(node *TextVNode, x, y int, buffer *paint.Buffer) {
    painter := paint.NewPainter(
        paint.NewPaintContext(buffer, paint.Rect{
            X: x, Y: y,
            Width:  buffer.Width - x,
            Height: buffer.Height - y,
        }),
    )

    // 使用 Painter 写入文本
    painter.Print(0, 0, node.content, node.Style())
}
```

#### ButtonVNode 渲染

```go
// ui/app.go - renderButton()

func (d *declarativeRoot) renderButton(node *ButtonVNode, x, y int, buffer *paint.Buffer) {
    painter := paint.NewPainter(
        paint.NewPaintContext(buffer, paint.Rect{
            X: x, Y: y,
            Width:  buffer.Width - x,
            Height: buffer.Height - y,
        }),
    )

    // 根据焦点状态选择样式
    style := node.Style()
    if node.isFocused || node.IsFocused() {
        style = style.Reverse(true)
    }

    // 绘制按钮边框和文本
    painter.DrawButton(0, 0, node.label, node.isFocused, style, style)
}
```

---

## 两种渲染模式

### Legacy 模式 (直接递归)

```
VNode Tree
    ↓
renderVNode() 递归遍历
    ↓
┌──────────────────────────────────┐
│ TextVNode → renderText()       │
│ ElementVNode → renderChildren()   │
│ LayoutNode → 计算布局 + render   │
│ ButtonVNode → renderButton()     │
│ InputVNode → renderInput()       │
└──────────────────────────────────┘
    ↓
直接写入 Buffer
```

### Fiber 模式 (可中断渲染)

```
VNode Tree
    ↓
reconciler.Render()
    ↓
┌──────────────────────────────────┐
│ BeginWork 阶段                   │
│ - 构建 Fiber 树                  │
│ - 处理组件 Hooks                 │
│ - 标记脏节点                     │
└──────────────────────────────────┘
    ↓
┌──────────────────────────────────┐
│ WorkLoop 阶段                    │
│ - 遍历 Fiber 树                  │
│ - 对每个节点调用 renderCallback   │
└──────────────────────────────────┘
    ↓
renderVNodeFiber() 单节点渲染
    ↓
写入 Buffer
```

---

## 组件示例：完整的渲染流程

### 示例代码

```go
func Counter() ui.VNode {
    count, setCount := ui.UseState(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.HStack(
            ui.Button("-").OnClick(func() { setCount(count - 1) }),
            ui.Text(fmt.Sprintf("%d", count)),
            ui.Button("+").OnClick(func() { setCount(count + 1) }),
        ),
    )
}
```

### 渲染流程分解

```
1. ui.Run(Counter) 启动
   ↓
2. framework.App 事件循环等待
   ↓
3. 触发首次渲染 → declarativeRoot.Paint()
   ↓
4. paintLegacy() 调用 Counter() 函数
   ↓
5. Counter() 返回 VNode 树:

   VNode (LayoutNode/VStack)
   ├─ VNode (TextVNode) → "Count: 0"
   └─ VNode (LayoutNode/HStack)
       ├─ VNode (ButtonVNode) → "-"
       ├─ VNode (TextVNode) → "0"
       └─ VNode (ButtonVNode) → "+"
   ↓
6. renderVNode() 递归渲染:
   - VStack → 处理子节点
   - TextVNode → renderText() → 写入 "Count: 0"
   - HStack → 计算布局 → 处理子节点
   - ButtonVNode → renderButton() → 写入 "[-]" 或 "[+]"
   - TextVNode → renderText() → 写入 "0"
   ↓
7. 最终 Buffer 内容显示到终端:

   ┌─────────────────────────┐
   │ Count: 0                 │
   │                         │
   │ [─] 0 [+]               │
   └─────────────────────────┘
```

---

## 关键数据结构

### VNode 接口

```go
// ui/vnode.go
type VNode interface {
    Type() VNodeType
    Props() Props
    SetProps(p Props)
    Children() []VNode
    SetChildren(children []VNode)
    Key() string
    SetKey(key string)
    Style() style.Style
    SetStyle(s style.Style)
}
```

### 具体节点类型

```go
// 文本节点
type TextVNode struct {
    *ElementVNode
    content string
}

// 元素节点
type ElementVNode struct {
    nodeType    string
    props       Props
    children    []VNode
    key         string
    style       style.Style
}

// 布局节点
type LayoutNode struct {
    *ElementVNode
    direction  Direction
    align      Alignment
    gap        int
    padding    [4]int
}

// 按钮节点
type ButtonVNode struct {
    *ElementVNode
    label    string
    onClick func()
    disabled bool
    // ...
}
```

---

## 当前架构的限制

### 单一根组件的问题

当前实现使用单一 `declarativeRoot`，存在以下限制：

```
当前架构:
ui.Run(ComponentFunc)
    ↓
framework.App (单一应用实例)
    ↓
declarativeRoot (单一根组件)
    ↓
renderVNode() (统一渲染所有 VNode)
```

**限制**：
1. 只能通过 `ui.Run()` 启动一个声明式应用
2. 所有 VNode 必须在同一个 declarativeRoot 内渲染
3. 声明式组件无法与 imperative 组件混合使用
4. 不支持多个独立的声明式区域

### 与新设计的冲突

根据 `framework/docs/ui/idea/idea4_comp.md` 的设计愿景：

```
目标架构:
- 声明式组件应实现 Component 接口
- 可独立渲染，有自己的 Measure/Paint 方法
- 可与 imperative 组件自由组合
- 支持多个独立的声明式区域
```

**当前架构与此愿景存在冲突**。

---

## 重构后的渲染流程

### DeclarativeNode - 可独立渲染的声明式节点

重构将引入 `DeclarativeNode`，使声明式组件可独立渲染：

```go
// internal/render/declarative_node.go

type DeclarativeNode struct {
    component.Component

    // 声明式特有
    componentFn ui.ComponentFunc
    ctx         *ui.ComponentContext
    reconciler   *reconciler.Reconciler
    instanceMgr  *state.InstanceManager

    // 局部渲染
    bounds      paint.Rect    // 渲染边界
}

// 实现 Component 接口
func (d *DeclarativeNode) Measure(constraints Constraints) Size {
    // 计算组件大小
}

func (d *DeclarativeNode) Paint(ctx PaintContext) {
    // 独立渲染到指定区域
    d.reconciler.Render(ctx, ctx.Buffer(), d.componentFn)
}
```

### 多组件渲染流程

重构后的架构：

```
┌─────────────────────────────────────────────────────────────┐
│ framework.App                                              │
│                                                            │
│   ┌─────────────────┐  ┌─────────────────┐                │
│   │ Imperative      │  │ Declarative     │                │
│   │ Component       │  │ Component       │                │
│   │                 │  │                 │                │
│   │ Paint() {       │  │ Paint() {       │                │
│   │   直接绘制       │  │   调用自己的     │                │
│   │ }               │  │   reconciler    │                │
│   └─────────────────┘  └─────────────────┘                │
│                                                            │
│   两种组件可以混合在同一层级！                              │
└─────────────────────────────────────────────────────────────┘
```

### 混合使用示例

重构后，用户可以这样混合使用：

```go
func App() ui.VNode {
    return ui.VStack(
        ui.Text("Imperative Header"),

        // 声明式组件可以嵌入 imperative 容器
        ui.NewDeclarativeNode(Counter),

        ui.Button("Imperative Button"),
    )
}

// 或者直接使用 imperative 组件框架
func App2() component.Component {
    return &layout.Box{
        Children: []component.Component{
            &basic.Text{Content: "Hello"},
            NewDeclarativeComponent(MyCounter),
            &button.Button{Label: "Click"},
        },
    }
}
```

---

## 迁移路径

### Phase 1: 保持兼容（当前）

```
单一 declarativeRoot
    ↓ 渲染
所有 VNode
```

### Phase 2: 引入 DeclarativeNode

```
framework.App
    ├── imperative components (Component 接口)
    └── DeclarativeNode (声明式适配器)
            ↓ 独立的 reconciler
        VNode Tree
```

### Phase 3: 完全混合

```
任意 Component
    ├── 可能有多个 DeclarativeNode
    ├── 每个 DeclarativeNode 有自己的 reconciler
    └── 统一通过 Component.Paint() 渲染
```

---

## 总结

### 当前架构的核心机制

1. **函数即组件** - 用户函数返回 VNode 树
2. **VNode 描述 UI** - VNode 是轻量级的描述
3. **递归渲染** - renderVNode() 分发到具体渲染方法
4. **单一入口** - 通过 ui.Run() 启动

### 与 React 的对比

| 概念 | React | Mint (当前) | Mint (重构后) |
|------|-------|------------|--------------|
| 组件描述 | JSX/Element | VNode | VNode |
| 渲染入口 | ReactDOM.render | ui.Run() | Component.Paint() |
| 多根支持 | React Portal | ❌ 不支持 | ✅ 支持 |
| 混合模式 | - | ❌ 不支持 | ✅ 支持 |

### 重构目标

- 声明式组件可实现 `Component` 接口
- 可与 imperative 组件混合使用
- 支持多个独立的声明式区域
- 保持现有 API 完全兼容

---

**相关文档**:
- [组件迁移指南](./plan/COMPONENT_MIGRATION_GUIDE.md)
- [全面重构计划](./plan/COMPREHENSIVE_REFACTOR_PLAN.md)
- [Fiber 协调器](../../framework/docs/ui/idea/idea1.md)
- [组件接口设计](../../framework/docs/ui/idea/idea4_comp.md)
