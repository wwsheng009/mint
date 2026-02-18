很好，这个问题非常关键。

你现在已经完成：

* Fiber-first
* Action Runtime
* Lane 优先级
* 行为抽象层

接下来最大的挑战是：

> 在不回退到 closure 模式的情况下，如何保持“声明式 Button API”的优雅？

我们来推导一套：

# 🔥 新机制下的声明式 Button API 设计

目标：

* 保持你现在这种优雅写法
* 不让 VNode 持 closure
* 不破坏 Action + Lane 调度
* 不增加心智负担

---

# 一、先看旧模式的问题

你现在可能写的是：

```go
app.ButtonBuilder("[Open Modal]").
    OnClick(func() {
        setShowModal(true)
    }).
    Build()
```

问题：

* closure 被 capture
* 无法映射 Lane
* 不能优先级调度
* 无法中断
* 无法序列化
* 无法做 transition

这必须改。

---

# 二、正确方向：声明“意图”，不是“行为”

Button 不应该持有函数。

Button 应该声明：

> 当被点击时，触发某个 Action

---

# 三、第一种工业级方案（推荐）

## 方案：ActionID + Dispatcher 注册

---

## 1️⃣ 声明式 Button API

```go
app.Button("[Open Modal]").
    OnPress("openModal").
    Variant(app.ButtonVariantPrimary)
```

---

## 2️⃣ 在组件作用域注册 Action

```go
app.RegisterAction("openModal", func(ctx ActionContext) {
    ctx.SetState("showModal", true)
})
```

或者：

```go
useAction("openModal", func(ctx ActionContext) {
    setShowModal(true)
})
```

---

## 3️⃣ ButtonInstance 只发 Action

```go
func (b *ButtonInstance) OnPress() {
    Dispatch(Action{
        Type: ActionPress,
        ID:   b.props.ActionID,
        Priority: PriorityNormal,
    })
}
```

---

## 4️⃣ Action → Lane 映射

```
Action
   ↓
ResolvePriority
   ↓
Lane
   ↓
ScheduleFiber
```

---

# 四、第二种更高级方案（推荐长期演进）

## 🔥 声明式 Intent API

不是：

```
OnPress("openModal")
```

而是：

```go
OnPress(
    ui.Intent("OpenModal"),
)
```

Intent 是结构体：

```go
type Intent struct {
    Type string
    Payload any
    Priority ActionPriority
}
```

Button 发 Intent。

Dispatcher 解释 Intent。

---

# 五、第三种方案（类似 React Hook）

## 🔥 useAction Hook

```go
openModal := useAction(func(ctx ActionContext) {
    setShowModal(true)
})
```

然后：

```go
Button("[Open Modal]").
    OnPress(openModal)
```

这里：

* openModal 不是 closure
* 是 ActionHandle
* 由 Fiber Runtime 管理
* 带 Priority

内部是：

```go
type ActionHandle struct {
    id string
    lane Lane
}
```

---

# 六、真正工业级推荐结构

最终你应该是：

```
Button
   ↓
PressableBehavior
   ↓
Emit Intent
   ↓
Action Runtime
   ↓
Lane
   ↓
Scheduler
   ↓
Fiber Update
```

Button 永远不知道：

* 业务逻辑
* 状态修改
* Lane 细节

---

# 七、完整示例（新机制）

```go
headerContent := ui.HStack(
    ui.Text("TUI Engine Demo"),

    ui.Button("[Open Modal]").
        Intent(ui.Intent{
            Type: "OpenModal",
            Priority: ui.PriorityUserBlocking,
        }).
        Variant(ui.Primary),

    ui.Text(fmt.Sprintf("Clicks: %d", count)),
)
```

Action 注册：

```go
ui.OnIntent("OpenModal", func(ctx ui.ActionContext) {
    ctx.SetState("showModal", true)
})
```

---

# 八、为什么这样是正确的？

因为：

### 1️⃣ 组件声明的是意图

### 2️⃣ 运行期调度可抢占

### 3️⃣ 可以转成 Transition

```go
Intent{
    Type: "OpenModal",
    Priority: PriorityTransition,
}
```

### 4️⃣ 可以做 Suspense

---

# 九、你现在必须避免的

❌ OnClick(func(){})
❌ Button 直接 setState
❌ Button 直接操作 Fiber
❌ Button 直接调 Scheduler

---

# 十、最终 API 形态建议

你现在阶段，建议：

---

## 最实用版本（短期）

```go
Button("Open").
    OnPress("openModal")
```

---

## 中期版本（推荐）

```go
Button("Open").
    Intent("OpenModal")
```

---

## 长期版本（引擎级）

```go
Button("Open").
    Intent(ui.OpenModal()) // typed intent
```

---

# 十一、真正声明式的定义

声明式 ≠ closure

声明式 = 结构描述

---

# 十二、最终结构图

```
VNode
   ↓
Fiber
   ↓
Instance
   ↓
Behavior
   ↓
Intent
   ↓
ActionRuntime
   ↓
Lane
   ↓
Scheduler
```

---

# 关键判断

如果你删除所有 closure：

系统还能正常工作

那你的架构是对的。

---

# 结论

在新机制下：

Button 的声明式 API 应该：

* 不接受函数
* 只声明 Intent / ActionID
* 由 Runtime 统一调度

---

如果你愿意，我可以下一步：

* 🔥 设计完整 Intent 系统（带类型安全）
* 🔥 推导 Transition + Intent 整合
* 🔥 或帮你做一次“声明式 API 易用性优化设计”

你现在已经走到引擎核心阶段了。

很好。

这一步就是从“Action 字符串分发”升级到：

> 🔥 类型安全 Intent Runtime + 可并发调度 + 可 Transition 化

我会给你一个完整、可落地的设计：

1. ✅ 类型安全 Intent 系统
2. 🔥 Intent → Lane → Scheduler 精确链路
3. 🔥 Transition + Intent 整合模型
4. 🔥 声明式 API 易用性优化（不丢优雅）

我们直接进入工程级设计。

---

# 第一部分：🔥 类型安全 Intent 系统

目标：

* 不用字符串 `"openModal"`
* 不用 any
* 不用 closure
* 可静态类型检查
* 可自动推导 Lane

---

# 一、核心思想：Intent 是“命令类型”

不是：

```
ActionID string
```

而是：

```
OpenModalIntent struct{}
IncrementIntent struct{ Step int }
```

---

# 二、定义泛型 Intent 接口（Go 风格）

```go
type Intent interface {
    intentType()
}
```

每个 Intent 实现它：

```go
type OpenModalIntent struct{}

func (OpenModalIntent) intentType() {}
```

带参数：

```go
type IncrementIntent struct {
    Step int
}

func (IncrementIntent) intentType() {}
```

---

# 三、Intent Handler 注册（类型安全）

```go
type IntentHandler[T Intent] func(ctx ActionContext, intent T)
```

注册：

```go
func RegisterIntent[T Intent](handler IntentHandler[T]) {
    // 内部做 reflect.TypeOf((*T)(nil)).Elem() 做映射
}
```

---

# 四、派发

Button 不传 string。

直接传结构体：

```go
Button("Open").
    Intent(OpenModalIntent{})
```

ButtonInstance 内：

```go
DispatchIntent(OpenModalIntent{})
```

---

# 五、调度链路

```
Intent
   ↓
ResolvePriority(intent)
   ↓
Lane
   ↓
ScheduleFiber
   ↓
InvokeHandler[T]
```

---

# 六、Intent → Priority 映射

方式 1：接口声明

```go
type PriorityAware interface {
    Priority() ActionPriority
}
```

Intent 可实现：

```go
func (OpenModalIntent) Priority() ActionPriority {
    return PriorityUserBlocking
}
```

---

方式 2：全局注册映射表

```go
RegisterPriority[OpenModalIntent](PriorityUserBlocking)
```

推荐方式 1（更清晰）。

---

# 第二部分：🔥 Transition + Intent 整合模型

这是你系统成熟的关键。

---

## 一、问题

有些 Intent 需要：

* 延迟执行
* 低优先级
* 可中断
* 显示 loading

例如：

```
LoadDataIntent
```

---

## 二、定义 TransitionIntent

```go
type TransitionIntent interface {
    Intent
    IsTransition() bool
}
```

例如：

```go
type LoadDataIntent struct {
    URL string
}

func (LoadDataIntent) IsTransition() bool { return true }
```

---

## 三、调度升级

调度器看到：

```go
if intent implements TransitionIntent {
    lane = LaneTransition
}
```

---

## 四、Transition 生命周期

```
Dispatch Intent
   ↓
Mark Fiber as PendingTransition
   ↓
Render fallback
   ↓
Async work
   ↓
Commit
```

---

## 五、与 Suspense 整合

Intent Handler 可以返回：

```go
type AsyncResult struct {
    Promise Promise
}
```

Scheduler：

* 挂起 Fiber
* 显示 fallback
* Promise resolve 后恢复

---

# 第三部分：🔥 完整 Intent Runtime 架构

```
Component
   ↓
Emit Intent[T]
   ↓
Action Runtime
   ↓
Resolve Priority
   ↓
Map Lane
   ↓
Mark Fiber
   ↓
Scheduler
   ↓
Render
   ↓
Commit
```

所有 closure 消失。

---

# 第四部分：🔥 声明式 API 易用性优化

类型安全 + 易用 是关键。

---

## 方案 1：Builder 风格

```go
Button("Open").
    Intent(OpenModalIntent{})
```

---

## 方案 2：Typed DSL

```go
Button("Open").
    On(OpenModal())
```

其中：

```go
func OpenModal() OpenModalIntent {
    return OpenModalIntent{}
}
```

---

## 方案 3：链式 Transition 包装

```go
Button("Load").
    Intent(
        Transition(
            LoadDataIntent{URL: "..."},
        ),
    )
```

Transition 是包装器：

```go
func Transition[T Intent](intent T) T {
    // 标记优先级为 Transition
}
```

---

# 第五部分：🔥 工业级最终 API 示例

```go
ui.HStack(
    ui.Button("Open").
        Intent(OpenModal()),

    ui.Button("Load").
        Intent(
            Transition(
                LoadData{URL: "/api/data"},
            ),
        ),
)
```

注册：

```go
RegisterIntent(func(ctx ActionContext, intent OpenModalIntent) {
    ctx.SetState("showModal", true)
})

RegisterIntent(func(ctx ActionContext, intent LoadDataIntent) {
    data := fetch(intent.URL)
    ctx.SetState("data", data)
})
```

---

# 第六部分：为什么这是工业级？

✅ 无字符串
✅ 无 any
✅ 无 closure
✅ 支持优先级
✅ 支持 Transition
✅ 支持 Suspense
✅ 可静态检查
✅ 可自动生成文档
✅ 可 DevTools 追踪

---

# 第七部分：和你当前架构匹配程度

你现在：

* 有 Fiber
* 有 Action Runtime
* 有 Lane

只差：

> 把 Action 升级为 Intent 泛型系统

这一步完成后，

你的 Runtime 就进入：

> 框架级 → 引擎级

---

# 最后一句

Intent 是：

> 结构化的业务意图

不是回调。

---

如果你愿意，我下一步可以：

* 🔥 设计完整 Scheduler 时间切片实现
* 🔥 推导 Suspense 内核
* 🔥 或帮你画完整 Fiber Runtime 工业级架构图

你现在已经在构建一个真正的 UI 引擎内核。
