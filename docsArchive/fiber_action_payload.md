这个问题非常关键。

你现在的架构是：

* Action 是纯语义
* Fiber 不绑定业务回调
* Dispatcher 调用 Target
* Scheduler 控制更新

那么问题来了：

> 如果组件里需要回调函数参数（比如 onClick(x,y) / onChange(value)），参数如何安全传递？

我们从工业级设计角度推导一套**类型安全 + 并发安全 + 不污染 Fiber 的参数传递模型**。

---

# 一、核心原则

必须满足 4 条：

1. Fiber 不保存回调函数
2. VNode 不参与运行期
3. Action 可以携带参数
4. 参数不能在并发阶段被修改

---

# 二、工业级 Action 设计模型

## 1️⃣ Action 必须支持 Payload

```go
type Action struct {
    Type     ActionType
    TargetID string
    Payload  any
    Meta     ActionMeta
}
```

关键字段：

```go
Payload any
```

---

# 三、参数传递的三种工业方案

我们从简单到高级。

---

# ✅ 方案一：Payload 结构体传递（推荐）

最干净、最可维护。

---

## 示例：Click 事件

定义 payload：

```go
type ClickPayload struct {
    X int
    Y int
    Button int
}
```

EventSystem：

```go
payload := ClickPayload{
    X: x,
    Y: y,
    Button: btn,
}

bridge.DispatchFromFiber(
    fiber,
    ActionClick,
    payload,
)
```

---

组件：

```go
func (b *Button) HandleAction(a *Action) bool {

    if a.Type == ActionClick {

        p := a.Payload.(ClickPayload)

        runtime.Update(func() {
            b.lastX = p.X
        })

        return true
    }

    return false
}
```

优点：

* 类型明确
* 无 Fiber 依赖
* 无 VNode 依赖
* 并发安全（payload 是值拷贝）

这是推荐方案。

---

# ✅ 方案二：泛型 Action（更类型安全）

如果你想更工业级：

```go
type Action[T any] struct {
    Type     ActionType
    TargetID string
    Payload  T
}
```

但：

* Dispatcher 复杂度增加
* 泛型不利于统一处理

一般 UI Runtime 不推荐泛型 Action。

---

# ❌ 方案三：回调函数传入（不推荐）

比如：

```go
Action{
    Type: ActionClick,
    Payload: func() {...}
}
```

问题：

* 破坏并发安全
* 无法序列化
* 无法 replay
* 无法做 time travel

工业级引擎必须避免。

---

# 四、如何支持“组件声明式回调参数”

比如：

```go
Button{
    OnClick: func(value int) {...}
}
```

但你又不能把这个函数塞到 Fiber。

怎么办？

---

## 正确做法：ActionID + 参数分离

VNode 阶段：

```go
Button{
    ActionID: "btn_save",
}
```

Dispatcher 注册：

```go
dispatcher.Register("btn_save", buttonInstance)
```

参数通过 Payload 传递。

Fiber 只存：

```go
ActionTargetID string
```

---

# 五、并发安全推导

⚠ 非常重要：

Action.Payload 必须满足：

* 不可变
* 不引用可变共享状态

建议规则：

| 类型         | 是否安全      |
| ---------- | --------- |
| int/string | ✅         |
| struct (值) | ✅         |
| slice      | ⚠ 必须 copy |
| map        | ⚠ 必须 copy |
| pointer    | ❌ 禁止      |

---

# 六、并发调度下的参数生命周期

完整流程：

```text
Input
  ↓
EventSystem
  ↓
构造 Action + Payload（值）
  ↓
Scheduler 入队
  ↓
稍后执行
  ↓
Target 读取 Payload
```

Payload 必须：

* 在构造时完整
* 不依赖 Fiber
* 不依赖 VNode
* 不引用组件内部字段

---

# 七、工业级升级模型：Immutable Payload

你可以做一个工具：

```go
func FreezePayload[T any](p T) T {
    return deepCopy(p)
}
```

但通常：

> 规范比工具更重要

约定 Payload 必须为值类型。

---

# 八、完整工业级事件数据流图

```text
Input Event
    ↓
HitTest
    ↓
Fiber
    ↓
ActionBridge
    ↓
Action {
    Type
    TargetID
    Payload (immutable)
}
    ↓
Scheduler
    ↓
Dispatcher
    ↓
Target.HandleAction
    ↓
runtime.Update()
```

---

# 九、回答你的核心问题

> 在处理 action 时，如何把回调函数的参数进行传递？

正确答案是：

👉 不传递函数
👉 只传递参数数据（Payload）
👉 函数逻辑存在 Target 中
👉 参数是 Action 的一部分

---

# 十、最推荐的工业级写法总结

```go
// 1. 定义 payload
type InputPayload struct {
    Value string
}

// 2. EventSystem 构造 action
bridge.DispatchFromFiber(
    fiber,
    ActionInputChange,
    InputPayload{Value: newValue},
)

// 3. 组件消费
func (i *Input) HandleAction(a *Action) bool {
    if a.Type == ActionInputChange {
        p := a.Payload.(InputPayload)
        runtime.Update(func() {
            i.value = p.Value
        })
        return true
    }
    return false
}
```

干净、稳定、并发安全。

---

# 十一、如果你愿意继续深入

我们可以下一步推导：

* 🔥 “Action Replay + Time Travel 调试模型”
* 🔥 “跨 Fiber Root 事件分发”
* 🔥 “批量 Action 合并优化”
* 🔥 “Fiber 无锁并发更新模型”

你现在已经在做真正的 UI 引擎核心设计，而不是普通组件框架。
很好，这个问题本质是：

> 在 **Action + Fiber-first 架构**下，如何优雅支持声明式组件里的 `OnClick(func(){})`？

你现在的声明式写法是：

```go
app.ButtonBuilder("[Open Modal]").
    Variant(app.ButtonVariantPrimary).
    OnClick(func() {
        setShowModal(true)
    }).
    FocusStyle(app.FocusStyleBracket).
    Build()
```

⚠ 但我们前面已经确立：

* Fiber 不能保存函数
* Action 不能携带函数
* 运行期不能依赖 VNode

所以我们必须改造 **声明式 API → Action 语义模型**。

---

# 一、核心问题拆解

当前写法的隐藏问题：

```go
OnClick(func() {
    setShowModal(true)
})
```

这里发生了三件事：

1. 捕获了闭包（含外部变量）
2. 把函数塞进 VNode
3. 运行期必须从 VNode 取函数

这会：

* ❌ 破坏 Fiber-first
* ❌ 不可并发
* ❌ 无法 replay
* ❌ 难以调试

---

# 二、工业级改造方向

我们把：

```text
声明式回调函数
```

改成：

```text
声明式 Action 描述
```

---

# 三、工业级方案一（推荐）：OnClick 变成 Action 描述器

## ✅ 新 API

```go
app.ButtonBuilder("[Open Modal]").
    OnClickAction(app.ActionOpenModal).
    Build()
```

---

## 1️⃣ 定义 Action 类型

```go
type UIAction string

const (
    ActionOpenModal UIAction = "open_modal"
)
```

---

## 2️⃣ Builder 存 ActionID，而不是函数

```go
type ButtonBuilder struct {
    clickAction UIAction
}

func (b *ButtonBuilder) OnClickAction(a UIAction) *ButtonBuilder {
    b.clickAction = a
    return b
}
```

---

## 3️⃣ Fiber 保存 ActionTargetID

在构建 Fiber 时：

```go
fiber.ActionTargetID = button.clickAction
```

---

## 4️⃣ Dispatcher 注册

```go
dispatcher.Register(ActionOpenModal, func(a *Action) {
    runtime.Update(func() {
        setShowModal(true)
    })
})
```

---

# 四、改造后的声明式写法

原来：

```go
OnClick(func() {
    setShowModal(true)
})
```

变成：

```go
OnClickAction(ActionOpenModal)
```

逻辑在 Dispatcher 中。

UI 是纯声明。

---

# 五、如果你仍然想支持闭包语法怎么办？

我们可以做一个 **编译期转换层**。

---

# 六、工业级方案二：Closure → Action 自动桥接

允许：

```go
OnClick(func() {
    setShowModal(true)
})
```

但内部实现为：

---

## 1️⃣ Builder 阶段生成唯一 ActionID

```go
actionID := runtime.GenerateActionID()
```

---

## 2️⃣ 注册闭包到 Dispatcher

```go
dispatcher.Register(actionID, func(a *Action) {
    handler()
})
```

---

## 3️⃣ Fiber 只保存 actionID

```go
fiber.ActionTargetID = actionID
```

---

## 4️⃣ ActionBridge 触发 actionID

```go
Action{
    Type: ActionClick,
    TargetID: actionID,
}
```

---

### ✅ 这样：

* Fiber 不存函数
* VNode 运行期无参与
* Dispatcher 持有闭包
* 并发安全

---

# 七、完整数据流

```text
声明式 Button
   ↓
Builder
   ↓
生成 ActionID
   ↓
Dispatcher 注册闭包
   ↓
Fiber 保存 ActionID
   ↓
点击
   ↓
ActionBridge
   ↓
Dispatcher
   ↓
执行闭包
   ↓
runtime.Update()
```

---

# 八、你当前 header 示例改造

原代码：

```go
app.ButtonBuilder("[Open Modal]").
    Variant(app.ButtonVariantPrimary).
    OnClick(func() {
        setShowModal(true)
    }).
    Build()
```

改为：

---

## 推荐（语义 Action）

```go
app.ButtonBuilder("[Open Modal]").
    Variant(app.ButtonVariantPrimary).
    OnClickAction(app.ActionOpenModal).
    Build()
```

Dispatcher：

```go
dispatcher.Register(app.ActionOpenModal, func(a *Action) {
    runtime.Update(func() {
        setShowModal(true)
    })
})
```

---

## 或兼容闭包版本

```go
app.ButtonBuilder("[Open Modal]").
    OnClick(func() {
        setShowModal(true)
    }).
    Build()
```

但内部自动：

* 生成 ActionID
* 注册到 Dispatcher
* Fiber 只存 ID

---

# 九、count 变量问题

你代码里还有：

```go
fmt.Sprintf("Clicks: %d", count)
```

这必须满足：

* count 在组件 state 内
* 不允许直接使用外部全局变量

否则：

> 并发下状态不同步

推荐：

```go
type AppState struct {
    count int
}
```

所有 render 读取 state。

---

# 十、最终工业级组件写法示例

```go
headerContent := ui.HStack(
    app.Text("TUI Engine Demo").
        Style(...),

    app.Spacer(),

    app.Button("[Open Modal]").
        Variant(app.ButtonVariantPrimary).
        OnClickAction(app.ActionOpenModal),

    app.Text(fmt.Sprintf("Clicks: %d", state.count)).
        Style(...),
)
```

UI 只描述：

* 结构
* 样式
* 行为语义

不包含函数逻辑。

---

# 十一、总结

在 Fiber-first 架构下：

| 不允许         | 允许                |
| ----------- | ----------------- |
| Fiber 持函数   | Fiber 持 ActionID  |
| Action 携带函数 | Action 携带 Payload |
| VNode 运行期使用 | Fiber 运行期使用       |
| 组件闭包驱动 UI   | Action 语义驱动 UI    |

---

# 十二、如果你愿意继续进阶

我们可以下一步：

* 🔥 设计“组件作用域 Action 自动注册系统”
* 🔥 设计“局部 Dispatcher（子树隔离）”
* 🔥 推导“Action 冒泡 + 捕获 + 优先级抢占”
* 🔥 或给你画一张完整 Runtime 交互图

你现在已经在做一个真正可工业落地的 TUI 引擎。

🔥 设计“组件作用域 Action 自动注册系统”

🔥 设计“局部 Dispatcher（子树隔离）”

🔥 推导“Action 冒泡 + 捕获 + 优先级抢占”

🔥 或给你画一张完整 Runtime 交互图

很好，这一题我们直接做一套**工业级完整模型**，一次把 4 件事打通：

1. ✅ 参数如何安全传递
2. 🔥 组件作用域 Action 自动注册系统
3. 🔥 局部 Dispatcher（子树隔离）
4. 🔥 Action 冒泡 + 捕获 + 优先级抢占
5. 🔥 完整 Runtime 交互图

你现在做的是引擎级 runtime，这一层设计必须非常干净。

---

# 一、参数如何传递（最终工业解）

核心原则：

> Action 传数据
> Dispatcher 持闭包
> Fiber 不持函数
> Payload 不可变

---

## 1️⃣ Action 结构

```go
type Action struct {
    Type     ActionType
    TargetID string
    Payload  any
    Lane     Lane
}
```

---

## 2️⃣ 参数通过 Payload 传递

例如按钮点击携带值：

```go
type ClickPayload struct {
    Value string
}
```

EventSystem：

```go
bridge.DispatchFromFiber(
    fiber,
    ActionClick,
    ClickPayload{Value: "abc"},
)
```

组件接收：

```go
func (b *Button) HandleAction(a *Action) bool {
    if a.Type == ActionClick {
        p := a.Payload.(ClickPayload)
        runtime.Update(func() {
            b.value = p.Value
        })
        return true
    }
    return false
}
```

⚠ Payload 必须：

* 不含可变引用
* 不引用组件内部状态
* 不传闭包

---

# 二、🔥 组件作用域 Action 自动注册系统

目标：

```go
Button().
    OnClick(func() {
        setShowModal(true)
    })
```

内部自动：

* 生成 ActionID
* 注册到当前组件 scope
* Fiber 只存 ID

---

## 1️⃣ ComponentScope 设计

```go
type ComponentScope struct {
    id        string
    actions   map[string]ActionHandler
}
```

---

## 2️⃣ Builder 阶段自动注册

```go
func (b *ButtonBuilder) OnClick(handler func()) *ButtonBuilder {

    actionID := runtime.GenerateActionID()

    runtime.CurrentScope().Register(actionID, func(a *Action) {
        handler()
    })

    b.actionID = actionID
    return b
}
```

---

## 3️⃣ Fiber 保存

```go
fiber.ActionTargetID = b.actionID
```

Fiber 不知道函数。

---

## 4️⃣ Scope 生命周期

Scope 跟随 Component Fiber：

```go
Fiber {
    Scope *ComponentScope
}
```

卸载 Fiber 时自动清理 Scope。

---

# 三、🔥 局部 Dispatcher（子树隔离）

目标：

> 不同子树拥有独立 Action 空间
> Modal 不影响主树
> 支持 Portal

---

## 1️⃣ Dispatcher 分层结构

```go
type Dispatcher struct {
    parent   *Dispatcher
    handlers map[string]ActionHandler
}
```

---

## 2️⃣ 局部查找逻辑

```go
func (d *Dispatcher) Dispatch(a *Action) bool {

    if handler, ok := d.handlers[a.TargetID]; ok {
        handler(a)
        return true
    }

    if d.parent != nil {
        return d.parent.Dispatch(a)
    }

    return false
}
```

---

## 3️⃣ Fiber 挂载 Dispatcher

```go
Fiber {
    Dispatcher *Dispatcher
}
```

Root 有全局 Dispatcher
Modal 有局部 Dispatcher

形成：

```text
RootDispatcher
    ├── AppDispatcher
    │     ├── ModalDispatcher
    │     └── SidebarDispatcher
```

实现子树隔离。

---

# 四、🔥 Action 冒泡 + 捕获 + 优先级抢占

我们做完整模型。

---

## 1️⃣ 三阶段模型

```text
Capture → Target → Bubble
```

---

## 2️⃣ Fiber 冒泡路径构建

```go
func BuildPath(f *Fiber) []*Fiber {
    var path []*Fiber
    for p := f; p != nil; p = p.Parent {
        path = append(path, p)
    }
    return path
}
```

---

## 3️⃣ 捕获阶段（自上而下）

```go
for i := len(path)-1; i >= 0; i-- {
    path[i].Dispatcher.DispatchCapture(action)
}
```

---

## 4️⃣ 目标阶段

```go
target.Dispatcher.Dispatch(action)
```

---

## 5️⃣ 冒泡阶段

```go
for i := 1; i < len(path); i++ {
    path[i].Dispatcher.DispatchBubble(action)
}
```

---

# 优先级抢占

Action 带 Lane：

```go
Action{
    Lane: LaneSync
}
```

Scheduler 永远执行：

```go
highestPriorityLane()
```

如果：

* 正在执行 Transition
* 来了 Sync

立即中断。

---

# 五、🔥 完整 Runtime 交互图

```text
                ┌─────────────────────┐
                │       Input         │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │    EventSystem      │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │      HitTest        │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │   Fiber Path Build  │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │   ActionBridge      │
                │  (Lane mapping)     │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │     Scheduler       │
                │  (Lane Queue)       │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │   Capture Phase     │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │   Target Phase      │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │   Bubble Phase      │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │   runtime.Update    │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │    Reconcile WIP    │
                └──────────┬──────────┘
                           ▼
                ┌─────────────────────┐
                │      Commit         │
                └─────────────────────┘
```

---

# 六、你现在的架构成熟度评估

如果你实现：

* Fiber-first
* Scope 自动注册
* 子树 Dispatcher
* Lane 抢占
* Capture/Bubble
* Immutable Payload

那你的 TUI Runtime 已经具备：

* 类 React 18 并发模型
* 可扩展插件架构
* Portal 支持
* Modal 隔离
* 可实现 Suspense

这已经是 UI 引擎级核心。

---

# 七、下一步建议

如果继续深入，我们可以做：

* 🔥 无锁 Scheduler
* 🔥 双缓冲 Fiber 内存模型优化
* 🔥 Action replay + 时间旅行调试
* 🔥 多 Root + Portal 完整隔离模型

你现在已经不在做“框架”，而是在做一个真正的 UI runtime 内核。

如果你愿意，我可以下一步给你：

👉 一套“最小可运行工业级 Runtime 核心代码骨架”。
