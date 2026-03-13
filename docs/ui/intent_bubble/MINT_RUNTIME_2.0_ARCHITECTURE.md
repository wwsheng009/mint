# Mint Runtime 2.0 架构升级方案

> **日期**: 2026-03-06
> **版本**: 2.0
> **状态**: 🚧 架构设计完成，待实施

---

## 📋 执行摘要

OptionGroup 组件暴露的问题**不仅仅是 Bug**，而是揭示了 **Mint UI Runtime 的核心架构缺陷**：

> **Fiber-first 架构中缺少正式的"父子组件关系模型"和"组件通信机制"。**

当前 Mint 只有 **Fiber 调度树**，但缺少 **Instance 组件树**、**Context 依赖注入**和 **Intent 事件冒泡**。这导致：

- ❌ 父子组件无法正常通信
- ❌ 组合组件（Form、Menu、Tabs、Tree）难以实现
- ❌ Portal、Dialog、Tooltip 等高级特性受限
- ❌ 频繁依赖 hack（全局注册表、闭包包装）

本方案提出 **Mint Runtime 2.0** 的完整架构升级，借鉴 React / Flutter / SwiftUI 的成熟模式，建立三层通信模型：

```
1️⃣ Instance Tree (结构关系)
2️⃣ Context (依赖注入)
3️⃣ Intent/Event (行为通信)
```

同时保持 **Fiber-first** 和 **声明式组件** 的核心设计，针对 **Go + TUI** runtime 进行优化。

---

## 🔍 一、当前架构缺陷分析

### 1.1 Mint 当前的三层模型

```
VNode (描述层)
    ↓
Fiber (调度层)
    ↓
Instance (组件层)
```

### 1.2 结构性缺陷

| 层级 | 生命周期 | 问题 |
|------|---------|------|
| **VNode** | 每次 render 创建后丢弃 | ❌ 无法后期修改 children props |
| **Fiber** | 每次 render 创建 | ❌ 只有调度树，Instance 无父子关系 |
| **Instance** | 组件生命周期内持久 | ❌ Instance 之间没有通信通道 |

### 1.3 根本问题对比

| 功能 | Mint 当前状态 | React / Flutter / SwiftUI |
|------|-------------|--------------------------|
| **Instance Tree** | ❌ 无 | ✅ 完整（组件逻辑树） |
| **Context** | ❌ 无 | ✅ 完整（依赖注入） |
| **Event Bubble** | ❌ 无 | ✅ 完整（事件冒泡） |
| **Parent Lookup** | ❌ 无（需要 hack） | ✅ 原生支持 |
| **Portal** | ⚠️ 部分实现 | ✅ 完整 |

### 1.4 受影响的组件

当前架构缺陷会影响以下组件的实现：

| 组件类型 | 受影响程度 | 当前状态 |
|---------|-----------|---------|
| **OptionGroup** | 🔴 严重 | 已暴露（子选项无法选中） |
| **Form** | 🔴 严重 | 字段联动和验证困难 |
| **Menu** | 🔴 严重 | 子菜单展开/收起受限 |
| **Tabs** | 🔴 严重 | 动态 Tab 管理困难 |
| **Tree** | 🔴 严重 | 展开/折叠逻辑受限 |
| **Portal** | 🟡 中等 | 已有基础实现，但功能不全 |
| **Dialog** | 🟡 中等 | 缺少 FocusTrap |
| **Tooltip** | 🟡 中等 | 定位受限 |
| **DragDrop** | 🟡 中等 | 事件冒泡缺失 |
| **List** | 🟢 轻微 | 虚拟滚动不受影响 |

---

## 🏗️ 二、Mint Runtime 2.0 总体架构

### 2.1 核心设计原则

1. **Fiber-first**: 保持 Fiber 调度引擎
2. **声明式组件**: VNode 描述 UI
3. **Instance Tree**: 建立组件逻辑树
4. **Context**: 依赖注入机制
5. **Intent Bubble**: 事件冒泡系统
6. **高性能**: 针对 TUI 优化

### 2.2 新架构层次

```
┌─────────────────────────────────────────────────────────┐
│                    Mint Runtime 2.0                     │
├─────────────────────────────────────────────────────────┤
│  VNode (声明式 UI)                                      │
│          ↓                                              │
│  Fiber Runtime (调度)                                   │
│          ↓                                              │
│  Component Instance (组件)                              │
│          ↓                                              │
│  Instance Tree (逻辑组件树) ──────────────────────┐    │
│          ↓                                         │    │
│  ┌─────────────────────────────────────────────┐ │    │
│  │  Context System (依赖注入)                   │ │    │
│  │  Intent/Event System (事件冒泡)              │ │    │
│  │  Layout Tree (渲染树)                        │ │    │
│  │  Portal Layer (门户层)                       │ │    │
│  └─────────────────────────────────────────────┘ │    │
└────────────────────────────────────────────────────┼────┘
                                                     ↓
                                            ┌────────┐
                                            │ Store  │
                                            │Reducer │
                                            └────────┘
```

### 2.3 组件通信模型

| 通信类型 | 用途 | 层级 | 示例 |
|---------|------|------|------|
| **Parent Lookup** | 组合组件 | Instance Tree | Form → Field, Menu → MenuItem |
| **Context** | 依赖注入 | Context System | Theme, Router, Provider |
| **Intent Bubble** | 事件通信 | Intent System | OptionClick → OptionGroup |
| **Store** | 全局状态 | App Level | AppState, Reducer |
| **Props** | 父→子显式传递 | VNode | 静态配置、初始值 |

---

## 🌲 三、第一层：Instance Tree（组件树）

### 3.1 核心概念

Mint 当前只有 **Fiber 调度树**：

```
Fiber Tree (调度)
├─ Fiber (OptionGroup)
│  ├─ Fiber (Option 1)
│  ├─ Fiber (Option 2)
│  └─ Fiber (Option 3)
```

但需要增加 **Instance 逻辑树**：

```
Instance Tree (逻辑)
├─ OptionGroupInstance
│  ├─ OptionInstance 1
│  ├─ OptionInstance 2
│  └─ OptionInstance 3
```

### 3.2 核心接口

```go
package runtime

// ComponentInstance 组件实例接口
type ComponentInstance interface {
    // 基础生命周期
    OnMount()
    OnUpdate()
    OnUnmount()

    // 组件树关系
    Parent() ComponentInstance
    Children() []ComponentInstance

    // 意图处理
    HandleIntent(intent.Intent) bool

    // 渲染
    Layout(ctx LayoutContext)
    Paint(ctx PaintContext)
}

// BaseInstance 基础实例实现
type BaseInstance struct {
    // 树关系
    parent   ComponentInstance
    children []ComponentInstance

    // 生命周期状态
    mounted  bool
    dirty    bool
}

// Parent 获取父组件
func (b *BaseInstance) Parent() ComponentInstance {
    return b.parent
}

// Children 获取子组件列表
func (b *BaseInstance) Children() []ComponentInstance {
    return b.children
}

// AddChild 添加子组件（Fiber 创建时调用）
func (b *BaseInstance) AddChild(c ComponentInstance) {
    b.children = append(b.children, c)
    child, ok := c.(*BaseInstance)
    if ok {
        child.parent = b
    }
}
```

### 3.3 Fiber 创建 Instance

```go
package fiber

// mountInstance 挂载组件实例
func mountInstance(f *Fiber) {
    // 创建实例
    inst := f.VNode.CreateInstance()
    f.Instance = inst

    // 建立父子关系
    if f.Parent != nil && f.Parent.Instance != nil {
        parentInst := f.Parent.Instance

        // 检查父实例是否支持 AddChild
        if parent, ok := parentInst.(interface{ AddChild(ComponentInstance) }); ok {
            parent.AddChild(inst)
        }
    }

    // 触发 OnMount
    inst.OnMount()
}

// unmountInstance 卸载组件实例
func unmountInstance(f *Fiber) {
    if f.Instance != nil {
        // 先卸载所有子组件
        child := f.Child
        for child != nil {
            unmountInstance(child)
            child = child.Sibling
        }

        // 再卸载当前组件
        f.Instance.OnUnmount()
        f.Instance = nil
    }
}
```

### 3.4 OptionGroup 实例示例

```go
type Instance struct {
    runtime.BaseInstance // 继承基础实例

    // 原有字段
    options  []Option
    selected  string
    selecteds []string
    // ...
}

// 不再需要 hack！
// Instance 可以直接访问子实例
func (inst *Instance) UpdateChildrenProps() {
    for _, childInst := range inst.Children() {
        if opt, ok := childInst.(*OptionInstance); ok {
            opt.UpdateFromParent(inst.selected, inst.mode)
        }
    }
}

// 子实例可以直接访父实例
func (inst *OptionInstance) SelectOption() {
    if parent := inst.Parent(); parent != nil {
        if group, ok := parent.(*OptionGroupInstance); ok {
            group.Select(inst.value)
        }
    }
}
```

---

## 🧩 四、第二层：Context（依赖注入）

### 4.1 核心概念

Context 类似 React Context，用于**跨层级依赖注入**，解决：

- ❌ Props Drilling（深层嵌套需要手动传递 Props）
- ❌ 组合组件需要手动回调
- ❌ 全局状态（Theme、Router）难以访问

### 4.2 核心接口

```go
package context

// ContextKey Context 键类型（类型安全）
type ContextKey string

// Context 存储 Fiber 层级的 Context 值
type FiberContext struct {
    values map[ContextKey]any
    parent *FiberContext // 支持嵌套
}

// NewContext 创建新的 Context
func NewContext(parent *FiberContext) *FiberContext {
    return &FiberContext{
        values: make(map[ContextKey]any),
        parent: parent,
    }
}

// Provide 提供一个 Context 值
// 在组件 VNode 或 VNode.Children() 中调用
func Provide(ctx *FiberContext, key ContextKey, value any) {
    if ctx != nil {
        ctx.values[key] = value
    }
}

// UseContext 从 Context 中获取值
// 在组件 Instance 中调用
func UseContext(ctx *FiberContext, key ContextKey) any {
    if ctx == nil {
        return nil
    }

    // 当前层级查找
    if v, ok := ctx.values[key]; ok {
        return v
    }

    // 向上查找
    if ctx.parent != nil {
        return UseContext(ctx.parent, key)
    }

    return nil
}

// UseContextValue 类型安全的 Context 访问
func UseContextValue[T any](ctx *FiberContext, key ContextKey) (T, bool) {
    v := UseContext(ctx, key)
    if v == nil {
        var zero T
        return zero, false
    }
    t, ok := v.(T)
    return t, ok
}
```

### 4.3 Fiber 集成

```go
package fiber

type Fiber struct {
    // ... 原有字段

    // Context 栈
    Context *context.FiberContext
}

// 在 CreateFiber 时创建/继承 Context
func CreateFiber(vnode VNode, parent *Fiber) *Fiber {
    f := &Fiber{
        VNode:   vnode,
        Parent:  parent,
        Context: context.NewContext(parent.Context),
    }
    return f
}
```

### 4.4 OptionGroup 示例

```go
// 全局 Context 键
const OptionGroupContext context.ContextKey = "optionGroup"

// ===== OptionGroup 提供者 =====

func (inst *Instance) OnMount() {
    inst.BaseInstance.OnMount()

    // 获取当前 Fiber 的 Context
    f := GetCurrentFiber(inst)

    // 提供 OptionGroup Instance 到 Context
    context.Provide(f.Context, OptionGroupContext, inst)
}

func (inst *Instance) SelectOption(value string) {
    // 原有逻辑
    // ...
}

// ===== Option 消费者 =====

type OptionInstance struct {
    runtime.BaseInstance
    value string
    // ...
}

func (inst *OptionInstance) HandleAction(act *action.Action) bool {
    if act.Type == action.ActionClick {
        // 从 Context 获取父实例
        f := GetCurrentFiber(inst)
        group, ok := context.UseContextValue[*Instance](
            f.Context,
            OptionGroupContext,
        )

        if ok {
            group.SelectOption(inst.value)
        }
        return true
    }
    return false
}

// 或者在 OnMount 时缓存引用
func (inst *OptionInstance) OnMount() {
    inst.BaseInstance.OnMount()

    f := GetCurrentFiber(inst)
    group, ok := context.UseContextValue[*Instance](
        f.Context,
        OptionGroupContext,
    )

    if ok {
        // 缓存父实例引用（可选）
        inst.group = group
    }
}

// 点击时直接使用缓存
func (inst *OptionInstance) HandleAction(act *action.Action) bool {
    if act.Type == action.ActionClick && inst.group != nil {
        inst.group.SelectOption(inst.value)
        return true
    }
    return false
}
```

### 4.5 Context 用例

| 用例 | Context Key | 提供者 | 消费者 |
|------|-------------|--------|--------|
| **OptionGroup** | `optionGroup` | OptionGroupInstance | OptionInstance |
| **Form** | `form` | FormInstance | FieldInstance |
| **Menu** | `menu` | MenuInstance | MenuItemInstance |
| **Theme** | `theme` | ThemeProvider | 所有组件 |
| **Router** | `router` | RouterProvider | Link, Navigation |

---

## 🎯 五、第三层：Intent Bubble（事件冒泡）

### 5.1 核心概念

Intent Bubble 类似 DOM 事件冒泡，用于**组件间事件通信**。

优势：
- ✅ 解耦父子组件
- ✅ 支持事件拦截
- ✅ 天然支持撤销/重做（Intent 是纯数据）

### 5.2 核心接口

```go
package intent

// Intent 意图接口
type Intent interface {
    Type() string
}

// IntentEmitter Intent 发射器
type IntentEmitter interface {
    Emit(intent.Intent)
}

// HandleIntentIntentHandler Intent 处理器
type IntentHandler interface {
    HandleIntent(Intent) bool // 返回 true 表示已处理
}

// Emit 向上冒泡 Intent
func Emit(inst ComponentInstance, i Intent) {
    if inst == nil {
        return
    }

    // 向上查找处理器
    current := inst
    for current != nil {
        // 检查是否实现了 IntentHandler
        if handler, ok := current.(IntentHandler); ok {
            handled := handler.HandleIntent(i)
            if handled {
                return // 已处理，停止冒泡
            }
        }

        // 继续向上冒泡
        current = current.Parent()
    }

    // 如果没有组件处理，可以传递给全局 Store
    Store.Dispatch(i)
}
```

### 5.3 OptionGroup Intent 示例

```go
// ===== Intent 定义 =====

type OptionSelectIntent struct {
    GroupKey string
    Value    string
}

func (OptionSelectIntent) Type() string {
    return "OptionSelect"
}

// ===== OptionInstance 发射 Intent =====

func (inst *OptionInstance) HandleAction(act *action.Action) bool {
    if act.Type == action.ActionClick {
        // 不再需要持有父实例引用
        // 发射 Intent，向上冒泡
        intent.Emit(inst, OptionSelectIntent{
            GroupKey: inst.groupKey,
            Value:    inst.value,
        })
        return true
    }
    return false
}

// ===== OptionGroupInstance 处理 Intent =====

type Instance struct {
    runtime.BaseInstance
    key string
    // ...
}

// 实现 IntentHandler
func (inst *Instance) HandleIntent(i intent.Intent) bool {
    switch v := i.(type) {
    case OptionSelectIntent:
        // 过滤只处理自己组的 Intent
        if v.GroupKey == inst.key {
            inst.SelectOption(v.Value)
            return true // 已处理
        }
    }
    return false // 未处理，继续冒泡
}
```

### 5.4 Intent 使用模式

| 组件 | 发射的 Intent | 处理的 Intent |
|------|--------------|--------------|
| **Option** | `OptionSelectIntent` | 不处理（纯发射器） |
| **OptionGroup** | 不处理 | `OptionSelectIntent` |
| **Menu Item** | `MenuSelectIntent` | 不处理 |
| **Menu** | `MenuCloseIntent` | `MenuSelectIntent`, `MenuCloseIntent` |
| **Form Field** | `FieldChangeIntent` | 不处理 |
| **Form** | 不处理 | `FieldChangeIntent`, `FormSubmitIntent` |

---

## 🚪 六、Portal（门户层）

### 6.1 核心概念

Portal 允许组件渲染到不同的位置，用于：

- **Dialog**: 模态对话框
- **Tooltip**: 工具提示
- **Dropdown**: 下拉菜单
- **Overlay**: 覆盖层

关键：**逻辑树 ≠ 渲染树**

```
逻辑树（Instance Tree）
├─ App
│  ├─ Content
│  └─ Button (触发 Dialog)
│
渲染树（Layout Tree）
├─ App
│  ├─ Content
│  └─ PortalRoot ← Dialog 渲染到这里
```

### 6.2 Portal VNode

```go
package portal

type VNode struct {
    rtui.ElementVNode

    target  string  // 渲染目标（"dialog", "tooltip" 等）
    child   rtui.VNode
}

func New(target string, child rtui.VNode) rtui.VNode {
    return &VNode{
        ElementVNode: rtui.NewElement("portal"),
        target:       target,
        child:        child,
    }
}

func (p *VNode) Type() string {
    return "portal"
}

func (p *VNode) Target() string {
    return p.target
}

func (p *VNode) Child() rtui.VNode {
    return p.child
}
```

### 6.3 Portal Fiber 处理

```go
package fiber

// reconcilePortals 处理 Portal Fiber
func reconcilePortals(root *Fiber, portalRoot *Fiber) {
    // 从根 Fiber 收集所有 Portal
    collectPortals := func(f *Fiber) []*Fiber {
        var portals []*Fiber
        var traverse func(*Fiber)
        traverse = func(fiber *Fiber) {
            if fiber.VNode.Type() == "portal" {
                portals = append(portals, fiber)
            }
            child := fiber.Child
            for child != nil {
                traverse(child)
                child = child.Sibling
            }
        }
        traverse(f)
        return portals
    }

    portals := collectPortals(root)

    // 将 Portal 爆炸到 portalRoot
    for _, portal := range portals {
        pvNode, ok := portal.VNode.(*portal.VNode)
        if !ok {
            continue
        }

        // 创建 Portal 的子 Fiber
        childFiber := CreateFiber(pvNode.Child(), portalRoot)
        childFiber.PortalParent = portal.Parent // 记录逻辑父级

        // 挂载到 portalRoot
        if portalRoot.Child == nil {
            portalRoot.Child = childFiber
        } else {
            sibling := portalRoot.Child
            for sibling.Sibling != nil {
                sibling = sibling.Sibling
            }
            sibling.Sibling = childFiber
        }
    }
}
```

### 6.4 Portal Instance

```go
package portal

type Instance struct {
    runtime.BaseInstance

    target  string
}

func NewInstance(props rtui.Props) ComponentInstance {
    return &Instance{
        target: props["target"].(string),
    }
}

func (p *Instance) OnMount() {
    p.BaseInstance.OnMount()

    // Portal 本身不渲染任何内容
    // 它的子组件会在 portalRoot 中渲染
}

func (p *Instance) Paint(ctx PaintContext) []paint.DrawCmd {
    // Portal 不绘制
    return nil
}
```

---

## 🎨 七、Layout Tree（渲染树）

### 7.1 核心概念

Layout Tree 独立于 Instance Tree，专注于布局和渲染：

```
Instance Tree = 逻辑树（父子关系、生命周期、Intent）
Layout Tree = 渲染树（位置、大小、可见性）
```

### 7.2 LayoutNode

```go
package layout

type Node struct {
    // 关联的 Instance
    Instance ComponentInstance
    Fiber    *Fiber

    // 布局关系
    Parent   *Node
    Children []*Node

    // 布局结果
    Bounds Rect // x, y, w, h
    Dirty  bool
}

type Rect struct {
    X, Y, W, H int
}
```

### 7.3 Build Layout Tree

```go
package layout

// BuildLayoutTree 从 Fiber 构建渲染树
func BuildLayoutTree(rootFiber *Fiber) *Node {
    var build func(*Fiber, *Node) *Node
    build = func(fiber *Fiber, parent *Node) *Node {
        if fiber == nil || fiber.Instance == nil {
            return nil
        }

        node := &Node{
            Instance: fiber.Instance,
            Fiber:    fiber,
            Parent:   parent,
        }

        // 递归构建子节点
        child := fiber.Child
        for child != nil {
            childNode := build(child, node)
            if childNode != nil {
                node.Children = append(node.Children, childNode)
            }
            child = child.Sibling
        }

        return node
    }

    return build(rootFiber, nil)
}
```

### 7.4 Layout 计算

```go
package layout

// Layout 计算布局
func Layout(root *Node, constraints Constraints) {
    if root == nil {
        return
    }

    // 计算当前节点
    ctx := LayoutContext{
        Bounds: constraints.Bounds,
    }
    root.Instance.Layout(ctx)
    root.Bounds = ctx.Bounds

    // 计算子节点
    for _, child := range root.Children {
        childConstraints := deriveConstraints(root.Bounds, child)
        Layout(child, childConstraints)
    }
}

type LayoutContext struct {
    Bounds Rect
}

type Constraints struct {
    MinWidth, MaxWidth  int
    MinHeight, MaxHeight int
    Bounds Rect
}
```

---

## 📅 八、实施路线图

### 8.1 分阶段实施

| 阶段 | 任务 | 工作量 | 优先级 | 状态 |
|------|------|-------|--------|------|
| **Phase 1** | 建立 Instance Tree | 2天 | 🔴 P0 | 📝 设计中 |
| **Phase 2** | 实现 Context System | 3天 | 🔴 P0 | ⏸️ 待开始 |
| **Phase 3** | 实现 Intent Bubble | 3天 | 🟡 P1 | ⏸️ 待开始 |
| **Phase 4** | 实现 Portal Layer | 2天 | 🟡 P1 | ⏸️ 待开始 |
| **Phase 5** | 重构 OptionGroup | 1天 | 🟢 P2 | ⏸️ 待开始 |
| **Phase 6** | 重构 Form 组件 | 2天 | 🟢 P2 | ⏸️ 待开始 |
| **Phase 7** | 添加 Menu/Tabs | 3天 | 🟢 P2 | ⏸️ 待开始 |
| **Phase 8** | 性能优化和测试 | 2天 | 🟢 P2 | ⏸️ 待开始 |

**总工作量**: 约 18 天

---

### 8.2 Phase 1: Instance Tree

#### 目标
建立组件逻辑树，解决父子关系问题。

#### 实施步骤

**Step 1.1**: 定义 `BaseInstance` 和接口
```go
// runtime/ui/instance.go
package instance

type BaseInstance struct {
    parent   ComponentInstance
    children []ComponentInstance
}

func (b *BaseInstance) Parent() ComponentInstance
func (b *BaseInstance) Children() []ComponentInstance
func (b *BaseInstance) AddChild(c ComponentInstance)
```

**Step 1.2**: 更新 `ComponentInstance` 接口
```go
// runtime/ui/instance.go
type ComponentInstance interface {
    Parent() ComponentInstance
    Children() []ComponentInstance
    OnMount()
    OnUpdate()
    OnUnmount()
}
```

**Step 1.3**: 修改 Fiber 创建逻辑
```go
// runtime/ui/fiber_util.go
func mountInstance(f *Fiber) {
    inst := f.VNode.CreateInstance()
    f.Instance = inst

    if f.Parent != nil && f.Parent.Instance != nil {
        parent := f.Parent.Instance
        parent.(AddChilder).AddChild(inst)
    }

    inst.OnMount()
}
```

**Step 1.4**: 更新现有组件
- `ui/components/panel/`
- `ui/components/optiongroup/`
- `ui/components/text/`

#### 验证
- [ ] 单元测试全部通过
- [ ] OptionGroup 子选项可以通过 `Parent()` 访问父实例
- [ ] 不再需要全局注册表 hack

---

### 8.3 Phase 2: Context System

#### 目标
实现依赖注入机制，解决 Props Drilling。

#### 实施步骤

**Step 2.1**: 定义 Context 接口
```go
// runtime/context/context.go
package context

type ContextKey string
type FiberContext struct {
    values map[ContextKey]any
    parent *FiberContext
}

func NewContext(parent *FiberContext) *FiberContext
func Provide(ctx *FiberContext, key ContextKey, value any)
func UseContext(ctx *FiberContext, key ContextKey) any
func UseContextValue[T any](ctx *FiberContext, key ContextKey) (T, bool)
```

**Step 2.2**: Fiber 集成
```go
// runtime/ui/fiber.go
type Fiber struct {
    Context *context.FiberContext
}
```

**Step 2.3**: 创建 Context Provider 组件
```go
// runtime/ui/context_provider.go
type Provider struct {
    key   context.ContextKey
    value any
}
```

**Step 2.4**: 使用示例
```go
// ui/components/optiongroup/
func (inst *Instance) OnMount() {
    f := GetCurrentFiber(inst)
    context.Provide(f.Context, OptionGroupContext, inst)
}

func (inst *OptionInstance) OnMount() {
    f := GetCurrentFiber(inst)
    group, ok := context.UseContextValue[*Instance](
        f.Context,
        OptionGroupContext,
    )
    // ...
}
```

#### 验证
- [ ] Context 可以跨层级访问
- [ ] Context 支持嵌套覆盖
- [ ] 类型安全的 `UseContextValue` 正常工作

---

### 8.4 Phase 3: Intent Bubble

#### 目标
实现事件冒泡机制，解耦父子组件。

#### 实施步骤

**Step 3.1**: 定义 Intent 接口
```go
// runtime/intent/types.go
package intent

type Intent interface {
    Type() string
}

type IntentEmitter interface {
    Emit(Intent)
}

type IntentHandler interface {
    HandleIntent(Intent) bool
}

func Emit(inst ComponentInstance, i Intent)
```

**Step 3.2**: 定义 OptionGroup Intent
```go
// ui/components/optiongroup/intent.go
type OptionSelectIntent struct {
    GroupKey string
    Value    string
}
```

**Step 3.3**: 修改 OptionInstance
```go
func (inst *OptionInstance) HandleAction(act *action.Action) bool {
    if act.Type == action.ActionClick {
        intent.Emit(inst, OptionSelectIntent{
            GroupKey: inst.groupKey,
            Value:    inst.value,
        })
        return true
    }
    return false
}
```

**Step 3.4**: 修改 OptionGroup
```go
func (inst *Instance) HandleIntent(i Intent) bool {
    switch v := i.(type) {
    case OptionSelectIntent:
        if v.GroupKey == inst.key {
            inst.SelectOption(v.Value)
            return true
        }
    }
    return false
}
```

#### 验证
- [ ] Intent 可以正确向上冒泡
- [ ] 父组件可以拦截 Intent
- [ ] 未处理的 Intent 传递到 Store

---

## 📊 九、新架构能力对比

### 9.1 通信方式对比

| 场景 | 当前方案（Hack） | 新方案（原生） |
|------|----------------|--------------|
| **子→父通信** | 全局注册表 / 闭包 | `Parent()` / Context / Intent |
| **深层组件访问** | Props Drilling | Context |
| **跨层事件** | 手动回调 | Intent Bubble |
| **Portal 渲染** | 部分支持 | 完整支持 |

### 9.2 组件实现对比

| 组件 | 当前难度 | 新难度 | 改善 |
|------|---------|-------|------|
| **OptionGroup** | 🔴 困难（需要 hack） | 🟢 简单 | ✅ 80% |
| **Form** | 🔴 困难 | 🟢 简单 | ✅ 70% |
| **Menu** | 🔴 困难 | 🟢 简单 | ✅ 75% |
| **Tabs** | 🟡 中等 | 🟢 简单 | ✅ 60% |
| **Tree** | 🔴 困难 | 🟡 中等 | ✅ 50% |
| **Portal** | ⚠️ 基础 | 🟢 完整 | ✅ 100% |

---

## 🎯 十、结论与建议

### 10.1 核心结论

OptionGroup 问题**不是 Bug**，而是揭示了 Mint Runtime 缺失的三个核心能力：

1. **Instance Tree** - 组件逻辑树（结构关系）
2. **Context System** - 依赖注入（解决 Props Drilling）
3. **Intent Bubble** - 事件冒泡（解耦父子通信）

建议**不要长期停留在闭包或全局注册表**这些 hack 方案上。

### 10.2 推荐方案

**长期方案**（分三步）：

```
Phase 1 (2天): 建立 Instance Tree
Phase 2 (3天): 实现 Context System
Phase 3 (3天): 实现 Intent Bubble
```

**短期方案**（可选）：

如果需要快速修复 OptionGroup，可以临时使用**闭包方案**（见 `IMPLEMENTATION_GUIDE.md`），但应明确这是**临时解决方案**。

### 10.3 架构升级收益

完成 Mint Runtime 2.0 升级后，系统获得：

- ✅ **原生组件通信** - 无需 hack
- ✅ **可组合组件** - Form、Menu、Tabs、Tree 优雅实现
- ✅ **Portal 支持** - Dialog、Tooltip、Overlay
- ✅ **类型安全** - Context + Go 泛型
- ✅ **高性能** - 针对Go + TUI 优化
- ✅ **可扩展性** - 符合现代 UI Runtime 架构

### 10.4 参考架构

本方案参考以下成熟系统：

- **React Fiber** - 调度引擎、事件冒泡
- **Flutter** - 三树架构（Widget → Element → RenderObject）
- **SwiftUI** - Environment 系统
- **BLoC** - Intent/Event 模式

但针对 **Go + TUI** runtime 进行了大量简化和优化。

---

## 🔗 附录

### A. 目录结构建议

```
mint/
├── runtime/
│   ├── fiber/
│   │   ├── fiber.go
│   │   ├── scheduler.go
│   │   └── lanes.go
│   ├── instance/
│   │   ├── base.go          ← BaseInstance
│   │   └── instance.go      ← 接口
│   ├── context/
│   │   └── context.go       ← Context
│   ├── intent/
│   │   ├── intent.go        ← Intent 接口
│   │   └── bubble.go        ← 冒泡逻辑
│   ├── layout/
│   │   └── tree.go          ← Layout Tree
│   ├── portal/
│   │   └── portal.go        ← Portal
│   └── ui/
│       ├── component.go
│       └── vnode.go
├── ui/
│   └── components/
│       ├── optiongroup/
│       │   ├── instance.go
│       │   ├── option.go
│       │   └── intent.go    ← OptionGroup Intent
│       ├── form/
│       │   ├── form.go
│       │   └── field.go
│       └── menu/
│           └── menu.go
└── docs/
    └── ui/
        └── optiongroup/
            ├── MINT_RUNTIME_2.0_ARCHITECTURE.md  ← 本文档
            ├── INSTANCE_TREE_IMPLEMENTATION.md   ← Phase 1
            ├── CONTEXT_IMPLEMENTATION.md         ← Phase 2
            └── INTENT_BUBBLE_IMPLEMENTATION.md   ← Phase 3
```

### B. 相关文档

- [ARCHITECTURE_ANALYSIS_REPORT.md](/docsArchive/ARCHITECTURE_ANALYSIS_REPORT.md) - 原始 Bug 分析
- [IMPLEMENTATION_GUIDE.md](./INTENT_BUBBLE_IMPLEMENTATION.md) - 临时闭包方案
- [CURRENT_STATUS.md](/docsArchive/CURRENT_STATUS.md) - 当前系统状态
- [review/review1.md](/docsArchive/review/review1.md) - 架构演进建议
- [review/review2.md](./review2.md) - 完整架构设计

---

**文档状态**: ✅ 架构设计完成

**下一步**: 开始实施 Phase 1 - Instance Tree
