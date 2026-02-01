# Mint UI 全面重构计划

> **创建日期**: 2026-02-01
> **状态**: 📋 规划中
> **目标**: 构建符合设计愿景的 Terminal UI Runtime Platform
> **参考文档**: framework/docs/ui/idea/*.md

---

## 目录

1. [愿景与现状对比](#一愿景与现状对比)
2. [架构分层设计](#二架构分层设计)
3. [组件解耦方案](#三组件解耦方案)
4. [核心模块重构](#四核心模块重构)
5. [实施路线图](#五实施路线图)
6. [验证标准](#六验证标准)

---

## 一、愿景与现状对比

### 1.1 设计愿景（来自 idea 文档）

根据 `framework/docs/ui/idea/idea7_final.md`，目标架构是：

```
Application Layer     ← 业务逻辑 / 页面 / 组件组合
    ↓
Declarative UI Layer  ← VNode / Hooks / State / Props
    ↓
Reconciler Layer      ← Diff / Fiber / Scheduler
    ↓
Layout Engine         ← Constraint Layout / Flex / Virtualized
    ↓
Render Engine         ← DrawCmd / Clip / Transform
    ↓
Animation Subsystem   ← Timeline / Easing / Physics
    ↓
Rasterizer            ← DrawCmd → Cells
    ↓
Dirty Region System   ← Diff Cells / Rect Merge
    ↓
Terminal Backend      ← ANSI Driver / Input / Resize
```

### 1.2 当前架构现状

```
ui/app.go (2169行)
    ├── declarativeRoot (单一根组件)
    │   ├── reconciler (Fiber协调器)
    │   ├── instanceManager
    │   ├── buttons/inputs/... (集中收集的交互元素)
    │   └── focusedIndex (全局焦点)
    └── renderVNode() (直接写Buffer)
```

### 1.3 差距分析

| 维度 | 设计愿景 | 当前实现 | 差距 |
|------|---------|---------|------|
| **组件契约** | Render/Measure/Paint 分离 | VNode 没有这些方法 | ⚠️ 需补充 |
| **渲染方式** | DrawCmd → Rasterizer | 直接写 Buffer | ⚠️ 部分实现 |
| **组件组织** | 按功能分类 (basic/form/button) | 全在 ui/ 目录 | ⚠️ 需重构 |
| **状态管理** | Hooks Slot (位置绑定) | 部分在 VNode 字段 | ⚠️ 需统一 |
| **多组件支持** | 可独立渲染 | 单一 declarativeRoot | ⚠️ 需解耦 |
| **动画系统** | 与状态分离 | 未分离 | ⚠️ 需新增 |

---

## 二、架构分层设计

### 2.1 目标目录结构

```
mint/
├── cmd/                            # 可执行程序
│   └── examples/                   # 示例入口
│
├── ui/                             # 公开 API 层 (精简)
│   ├── app.go                      # Run() 入口 ~100行
│   ├── hooks.go                    # Hooks API
│   ├── vnode.go                    # VNode 接口
│   ├── builder.go                  # 组件构造器
│   └── shortcuts.go                # 快捷函数
│
├── components/                     # 声明式组件库 (新增)
│   ├── basic/                      # 基础组件
│   │   ├── text.go                 # Text 组件
│   │   ├── icon.go                 # Icon 组件
│   │   ├── separator.go            # Separator 组件
│   │   └── spacer.go               # Spacer 组件
│   │
│   ├── layout/                     # 布局组件
│   │   ├── box.go                  # Box 组件
│   │   ├── flex.go                 # Flex 组件
│   │   ├── stack.go                # HStack/VStack
│   │   ├── grid.go                 # Grid 组件
│   │   └── overlay.go              # Overlay 组件
│   │
│   ├── form/                       # 表单组件
│   │   ├── input.go                # TextInput 组件
│   │   ├── textarea.go             # TextArea 组件
│   │   ├── checkbox.go             # Checkbox 组件
│   │   ├── select.go               # Select 组件
│   │   └── field.go                # Field 包装器
│   │
│   ├── button/                     # 按钮组件
│   │   ├── button.go               # Button 组件
│   │   ├── icon_button.go          # IconButton
│   │   └── button_group.go         # ButtonGroup
│   │
│   ├── feedback/                   # 反馈组件
│   │   ├── progress.go             # ProgressBar
│   │   ├── spinner.go              # Spinner
│   │   ├── toast.go                # Toast
│   │   ├── alert.go                # Alert
│   │   └── badge.go                # Badge
│   │
│   ├── data/                       # 数据展示
│   │   ├── list.go                 # List 组件
│   │   ├── table.go                # Table 组件
│   │   ├── tree.go                 # Tree 组件
│   │   └── virtuallist.go          # VirtualList
│   │
│   ├── navigation/                 # 导航组件
│   │   ├── tabs.go                 # Tabs
│   │   ├── menu.go                 # Menu
│   │   └── sidebar.go              # Sidebar
│   │
│   ├── overlay/                    # 覆盖层组件
│   │   ├── modal.go                # Modal
│   │   ├── dialog.go               # Dialog
│   │   ├── dropdown.go             # Dropdown
│   │   └── tooltip.go              # Tooltip
│   │
│   └── container/                  # 容器组件
│       ├── panel.go                # Panel
│       ├── split.go                # SplitPane
│       └── scroll.go               # ScrollArea
│
├── internal/                       # 内部实现 (不对外暴露)
│   ├── reconciler/                 # 协调器系统
│   │   ├── fiber.go
│   │   ├── reconciler.go
│   │   ├── begin_work.go
│   │   ├── complete_work.go
│   │   ├── diff.go
│   │   └── public.go               # 公开接口
│   │
│   ├── scheduler/                  # 调度器
│   │   ├── ui_scheduler.go
│   │   └── priority.go
│   │
│   ├── state/                      # 状态系统
│   │   ├── instance.go
│   │   ├── instance_manager.go
│   │   ├── interaction_state.go
│   │   └── public.go
│   │
│   └── render/                     # 渲染引擎
│       ├── rnode.go                # 真实节点树
│       ├── layout_engine.go        # 布局引擎
│       ├── render_tree.go          # 渲染树
│       └── rasterizer.go           # 栅格化器
│
├── framework/                      # 框架层 (保持)
│   ├── app.go
│   ├── component/
│   ├── event/
│   └── ...
│
├── runtime/                        # 运行时层 (保持)
│   ├── paint/                      # 已有 Painter/DrawCmd
│   ├── event/
│   ├── platform/
│   └── ...
│
└── docs/                           # 文档
```

### 2.2 依赖关系图

```
components/ (用户层)
    ↓ 使用
ui/ (公开API层)
    ↓ 使用
internal/ (内部实现层)
    ├─ reconciler/
    ├─ scheduler/
    ├─ state/
    └─ render/
    ↓ 使用
framework/ (框架层)
    ↓ 使用
runtime/ (运行时层)
```

---

## 三、组件解耦方案

### 3.1 组件标准接口

```go
// internal/render/component.go

// Component 组件标准接口
// 参考: framework/docs/ui/idea/idea4_comp.md
type Component interface {
    // 组件标识
    ID() string
    Type() string

    // 生命周期 (参考 idea4_comp.md)
    Mount(ctx Context) error
    Update(ctx Context) error
    Unmount(ctx Context) error

    // 渲染能力 (与 idea3_vnode.md 一致)
    Measure(constraints Constraints) Size
    Paint(ctx PaintContext)
}

// 渲染上下文 - 已在 runtime/paint/context.go 实现
// PaintContext 提供了 Painter 接口，组件通过 Painter 绘制
```

### 3.2 VNode 与 Component 的关系

```go
// VNode 保持为轻量级描述接口
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

// 新增: 可渲染 VNode 接口
type RenderableVNode interface {
    VNode
    Measure(constraints Constraints) Size
    Paint(ctx PaintContext)
}

// ElementVNode 实现 RenderableVNode
func (e *ElementVNode) Measure(constraints Constraints) Size {
    // 根据类型调用对应的组件
}

func (e *ElementVNode) Paint(ctx PaintContext) {
    // 使用 runtime/paint 的 Painter
}
```

### 3.3 组件库结构

```go
// components/form/input.go

package form

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/runtime/style"
)

// TextInput 组件
// 实现内部组件接口，同时提供 Builder 模式
type TextInput struct {
    // 组合而非继承
    vnode ui.VNode

    // 组件状态 (不持久化，每次重建)
    value    string
    placeholder string
    readOnly bool
    disabled bool
    onChange func(string)
}

// New 创建输入组件
func NewInput() *TextInput {
    return &TextInput{
        vnode: ui.NewElement("input"),
    }
}

// Build 构建 VNode (供 ui 包使用)
func (t *TextInput) Build() ui.VNode {
    return t.vnode
}

// Builder 模式
type InputBuilder struct {
    node *TextInput
}

func Input(placeholder string) *InputBuilder {
    return &InputBuilder{
        node: &TextInput{
            vnode:      ui.NewElement("input"),
            placeholder: placeholder,
        },
    }
}

func (b *InputBuilder) Value(v string) *InputBuilder {
    b.node.value = v
    return b
}

func (b *InputBuilder) OnChange(fn func(string)) *InputBuilder {
    b.node.onChange = fn
    return b
}

func (b *InputBuilder) Build() ui.VNode {
    return b.node.Build()
}
```

### 3.4 ui 包作为入口

```go
// ui/shortcuts.go

// 提供便捷的组件构造函数
// 这些函数代理到 components 包

import (
    "github.com/wwsheng009/mint/components/basic"
    "github.com/wwsheng009/mint/components/form"
    "github.com/wwsheng009/mint/components/button"
    "github.com/wwsheng009/mint/components/layout"
)

// Text 创建文本组件
func Text(content string) ui.VNode {
    return basic.NewText(content).Build()
}

// Input 创建输入组件
func Input(placeholder string) *form.InputBuilder {
    return form.Input(placeholder)
}

// Button 创建按钮组件
func Button(label string) *button.ButtonBuilder {
    return button.Button(label)
}

// HStack 创建水平布局
func HStack(children ...ui.VNode) ui.VNode {
    return layout.HStack(children...)
}
```

---

## 三、声明式 API 设计与兼容性

### 3.1 核心承诺：声明式组件功能完全保留

重构的**首要原则**是保持现有代码无需修改。用户编写的声明式组件将在重构后继续工作：

```go
// ✅ 当前写法 (重构后完全支持)
func Counter() ui.VNode {
    count, setCount := ui.UseState(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.HStack(
            ui.Button("-").OnClick(func() { setCount(count - 1) }),
            ui.Text(fmt.Sprintf("%d", count)),
            ui.Button("+").OnClick(func() { setCount(count + 1) }),
        ),
        ui.Input("Name").Value(name),
    )
}
```

### 3.2 三种 API 层次并存

重构后提供三种使用方式，满足不同场景需求：

#### 方式 1: 声明式函数（推荐，最简洁）

```go
// 快速构建 UI，适合大多数场景
func MyApp() ui.VNode {
    return ui.VStack(
        ui.Text("Hello"),
        ui.Button("Click"),
    )
}
```

#### 方式 2: Builder 模式（复杂配置）

```go
// 需要多属性配置时使用
ui.Input("Name").
    Placeholder("Enter name").
    Value(value).
    OnChange(func(s string) { value = s }).
    Style(style.Style{}.Fg(color.Cyan))
```

#### 方式 3: 直接导入（高级用法）

```go
// 需要访问组件特有方法或精细控制时
import "github.com/wwsheng009/mint/components/form"

input := form.NewInput()
// ... 直接操作 input
```

### 3.3 API 桥接层设计

```go
// ui/shortcuts.go - API 桥接层

package ui

// 基础组件快捷函数
func Text(content string) ui.VNode {
    return components.basic.Text(content).Build()
}

func Icon(name string) ui.VNode {
    return components.basic.Icon(name).Build()
}

// 布局组件快捷函数
func HStack(children ...ui.VNode) ui.VNode {
    return components.layout.HStack(children...)
}

func VStack(children ...ui.VNode) ui.VNode {
    return components.layout.VStack(children...)
}

// 表单组件快捷函数 - 返回 Builder 以支持链式调用
func Input(placeholder string) *components.form.InputBuilder {
    return components.form.Input(placeholder)
}

func Checkbox(label string) *components.form.CheckboxBuilder {
    return components.form.Checkbox(label)
}

// 按钮组件快捷函数
func Button(label string) *components.button.ButtonBuilder {
    return components.button.Button(label)
}
```

### 3.4 完整示例：重构前后对比

#### 重构前（当前）

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/ui"
)

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

func main() {
    ui.Run(Counter)
}
```

#### 重构后（保持不变！）

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/ui"  // ← 仅此导入不变
)

func Counter() ui.VNode {  // ← 函数签名不变
    count, setCount := ui.UseState(0)  // ← Hooks 用法不变

    return ui.VStack(  // ← 声明式组件不变
        ui.Text(fmt.Sprintf("Count: %d", count)),  // ← 简单组件不变
        ui.HStack(  // ← 布局组件不变
            ui.Button("-").OnClick(func() { setCount(count - 1) }),  // ← 链式调用不变
            ui.Text(fmt.Sprintf("%d", count)),
            ui.Button("+").OnClick(func() { setCount(count + 1) }),
        ),
    )
}

func main() {
    ui.Run(Counter)  // ← 运行方式不变
}
```

**关键点**：用户代码 **零修改**！

### 3.5 新增能力：直接导入组件

重构后，用户还可以选择直接导入组件：

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/ui"              // 核心 API
    "github.com/wwsheng009/mint/components/form"  // 表单组件
    "github.com/wwsheng009/mint/components/button" // 按钮组件
)

func Counter() ui.VNode {
    count, setCount := ui.UseState(0)

    // 方式 1: 使用 ui 快捷函数
    incrementBtn := ui.Button("+").OnClick(func() { setCount(count + 1) })

    // 方式 2: 直接使用 form.Button (可访问更多方法)
    nameInput := form.Input("Name").
        Placeholder("Enter name").
        MaxLength(20).
        Validator(func(s string) error {
            if len(s) < 3 {
                return fmt.Errorf("name too short")
            }
            return nil
        })

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        incrementBtn,
        nameInput.Build(),  // Builder 需要调用 Build()
    )
}
```

### 3.6 API 迁移路径

```
Phase 0 (当前)          Phase 1 (重构后)       Phase 2 (新增能力)
─────────────────    ───────────────────    ──────────────────
ui.Text("Hello")  →  ui.Text("Hello")   →  三种方式并存
                      (零代码改动)         (新增直接导入)
```

---

## 四、核心模块重构

### 4.1 渲染流程重构

当前问题：`renderVNode` 直接写 Buffer

目标流程：
```
VNode Tree
    ↓ Measure
Layout Tree (with x,y,w,h)
    ↓ Paint (生成 DrawCmd)
Render Tree (DrawCmd[])
    ↓ Rasterize
Buffer
```

### 4.2 DrawCmd 模式 (已部分实现)

`runtime/paint` 已有 `Painter` 和 `PaintContext`，需要明确：

```go
// internal/render/render_tree.go

// DrawCmd 绘制命令接口
type DrawCmd interface {
    // 命令的类型，用于批量优化
    Type() CmdType
    // 执行绘制
    Execute(painter *paint.Painter)
}

type CmdType int

const (
    CmdText CmdType = iota
    CmdFill
    CmdBox
   CmdCustom
)

// TextCmd 文本绘制命令
type TextCmd struct {
    X, Y  int
    Text  string
    Style style.Style
}

func (c *TextCmd) Type() CmdType { return CmdText }
func (c *TextCmd) Execute(painter *paint.Painter) {
    painter.Print(c.X, c.Y, c.Text, c.Style)
}
```

### 4.3 多组件支持

当前：单一 `declarativeRoot`

目标：支持多个独立的声明式组件

```go
// internal/render/declarative_node.go

// DeclarativeNode 可独立渲染的声明式节点
type DeclarativeNode struct {
    component.Component

    // 声明式特有
    componentFn ui.ComponentFunc
    ctx         *ui.ComponentContext
    reconciler   *reconciler.Reconciler
    instanceMgr  *state.InstanceManager

    // 局部渲染
    buffer       *paint.Buffer  // 可选：独立缓冲区
    bounds       paint.Rect      // 渲染边界
}

// RenderTo 渲染到指定区域
func (d *DeclarativeNode) RenderTo(x, y int, buffer *paint.Buffer) {
    ctx := paint.NewPaintContext(buffer, paint.Rect{
        X: x, Y: y,
        Width: d.bounds.Width,
        Height: d.bounds.Height,
    })

    // 执行渲染流程
    d.reconciler.Render(ctx, buffer, d.componentFn)
}
```

### 4.4 焦点系统解耦

当前：全局 `focusedIndex` + 集中收集按钮

目标：每个组件可管理自己的焦点

```go
// internal/state/focus_manager.go

// FocusManager 焦点管理器
type FocusManager struct {
    mu       sync.RWMutex
    focused  string // 当前焦点组件 ID
    root     *FocusScope
}

// FocusScope 焦点作用域
type FocusScope struct {
    id       string
    parent   *FocusScope
    children []*FocusScope
    focusable []string // 可聚焦的组件 ID
    current  int      // 当前焦点索引
}

// NewFocusManager 创建焦点管理器
func NewFocusManager() *FocusManager {
    return &FocusManager{
        root: &FocusScope{id: "root"},
    }
}

// Register 注册可聚焦组件
func (fm *FocusManager) Register(scopeID, componentID string) {
    // ...
}

// Focus 设置焦点
func (fm *FocusManager) Focus(componentID string) bool {
    // ...
}

// Next 下一个焦点
func (fm *FocusManager) Next() {
    // ...
}

// Prev 上一个焦点
func (fm *FocusManager) Prev() {
    // ...
}
```

---

## 五、实施路线图

### Phase 1: 基础重构 (Week 1-2)

**目标**: 目录重组，接口定义

```
┌─────────────────────────────────────────────┐
│ Week 1                                      │
├─────────────────────────────────────────────┤
│ Day 1-2: 创建目录结构                        │
│   ├── 创建 components/ 各子目录             │
│   ├── 创建 internal/ 各子目录               │
│   └── 创建接口定义文件                       │
│                                              │
│ Day 3-4: 迁移核心文件到 internal/            │
│   ├── reconciler/                           │
│   ├── scheduler/                            │
│   └── state/                                │
│                                              │
│ Day 5: 更新导入路径，确保编译通过            │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ Week 2                                      │
├─────────────────────────────────────────────┤
│ Day 1-3: 迁移组件到 components/             │
│   ├── 从 ui/ 提取组件到各分类目录            │
│   ├── 保持 Builder 模式                      │
│   └── 添加 components 包的 public.go         │
│                                              │
│ Day 4-5: 更新 ui/ 作为入口层                 │
│   ├── 精简 ui/app.go                         │
│   ├── 添加 shortcuts.go                      │
│   └── 保持向后兼容的 API                     │
└─────────────────────────────────────────────┘
```

### Phase 2: 渲染重构 (Week 3-4)

**目标**: 实现完整的渲染管线

```
┌─────────────────────────────────────────────┐
│ Week 3                                      │
├─────────────────────────────────────────────┤
│ Day 1-2: RNode 系统实现                     │
│   ├── RNode 数据结构                        │
│   ├── VNode → RNode 转换                    │
│   └── RNode 树遍历                           │
│                                              │
│ Day 3-4: Layout Engine                       │
│   ├── 约束系统                               │
│   ├── 测量传递                               │
│   └── 布局计算                               │
│                                              │
│ Day 5: Render Tree                          │
│   ├── DrawCmd 收集                           │
│   └── 渲染命令优化                           │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ Week 4                                      │
├─────────────────────────────────────────────┤
│ Day 1-2: Rasterizer                         │
│   ├── DrawCmd 执行                           │
│   ├── 裁剪与变换                             │
│   └── Buffer 写入                            │
│                                              │
│ Day 3-4: 集成测试                            │
│   ├── 端到端渲染测试                         │
│   ├── 性能基准测试                           │
│   └── 内存泄漏检测                           │
│                                              │
│ Day 5: 文档更新                              │
└─────────────────────────────────────────────┘
```

### Phase 3: 组件完善 (Week 5-6)

**目标**: 完善组件库

```
┌─────────────────────────────────────────────┐
│ Week 5                                      │
├─────────────────────────────────────────────┤
│ Day 1-2: 补充基础组件                        │
│   ├── Icon, Separator, Spacer              │
│   └── 样式统一                               │
│                                              │
│ Day 3-4: 补充表单组件                        │
│   ├── TextArea 完善                         │
│   ├── Switch, Slider                        │
│   └── Field 包装器                           │
│                                              │
│ Day 5: 补充反馈组件                          │
│   ├── Toast, Alert                          │
│   └── Badge                                  │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ Week 6                                      │
├─────────────────────────────────────────────┤
│ Day 1-2: 补充导航组件                        │
│   ├── Tabs, Menu                            │
│   └── Sidebar                                │
│                                              │
│ Day 3-4: 补充容器组件                        │
│   ├── Panel, SplitPane                      │
│   └── ScrollArea                             │
│                                              │
│ Day 5: 示例更新                              │
└─────────────────────────────────────────────┘
```

### Phase 4: 高级特性 (Week 7+)

**目标**: 动画、并发、DevTools

```
Week 7+: 动画系统 + 并发模式 + DevTools
```

---

## 六、验证标准

### 6.1 功能验证

| 验证项 | 标准 | 测试方法 |
|--------|------|----------|
| API 兼容性 | 所有公开 API 仍然可用 | 运行现有示例 |
| 组件独立 | 组件可独立导入使用 | `import "github.com/wwsheng009/mint/components/form"` |
| 渲染正确 | 与当前实现输出一致 | 视觉对比测试 |
| 性能不退化 | 关键操作性能保持 | 基准测试 |

### 6.2 架构验证

| 验证项 | 标准 |
|--------|------|
| 分层清晰 | ui/ → components/ → internal/ → framework/ |
| 职责单一 | 每个包只有一个明确职责 |
| 依赖单向 | 无循环依赖 |
| 接口稳定 | public.go 定义的接口稳定 |

### 6.3 代码质量

```bash
# 验证命令
go build ./...                    # 编译通过
go test ./... -cover               # 测试覆盖率 > 70%
go vet ./...                       # 静态检查
golangci-lint run                  # 代码质量
```

---

## 附录

### A. 文件迁移清单

| 源文件 | 目标 | 优先级 |
|--------|------|--------|
| ui/fiber.go | internal/reconciler/fiber.go | P0 |
| ui/reconciler.go | internal/reconciler/reconciler.go | P0 |
| ui/diff.go | internal/reconciler/diff.go | P0 |
| ui/begin_work.go | internal/reconciler/begin_work.go | P0 |
| ui/complete_work.go | internal/reconciler/complete_work.go | P0 |
| ui/scheduler.go | internal/scheduler/ui_scheduler.go | P0 |
| ui/instance.go | internal/state/instance.go | P0 |
| ui/instance_manager.go | internal/state/instance_manager.go | P0 |
| ui/interaction_state.go | internal/state/interaction_state.go | P0 |
| ui/button.go | components/button/button.go | P1 |
| ui/input.go | components/form/input.go | P1 |
| ui/checkbox.go | components/form/checkbox.go | P1 |
| ui/select.go | components/form/select.go | P1 |
| ui/textarea.go | components/form/textarea.go | P1 |
| ui/progress.go | components/feedback/progress.go | P1 |
| ui/modal.go | components/overlay/modal.go | P1 |
| ui/tooltip.go | components/overlay/tooltip.go | P1 |
| ui/virtuallist.go | components/data/virtuallist.go | P1 |
| ui/layout.go | components/layout/stack.go | P1 |
| ui/absolute.go | components/layout/absolute.go | P1 |
| ui/grid.go | components/layout/grid.go | P1 |
| ui/text.go | components/basic/text.go | P1 |

### B. 关键设计决策

1. **components/ 对外公开** - 用户可以直接 `import "github.com/wwsheng009/mint/components/form"`
2. **ui/ 作为快捷入口** - 提供 `ui.Input()` 等便捷函数
3. **internal/ 完全隐藏** - 实现细节不对外暴露
4. **向后兼容** - 保留所有现有 API，渐进式迁移

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
