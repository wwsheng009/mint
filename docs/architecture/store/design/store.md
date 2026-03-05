你这个问题本质上不是 Go 的“闭包 bug”，而是 **TUI 声明式运行时 + 事件注册生命周期不清晰** 导致的“闭包捕获时机错误”。

我直接给你拆透，不讲表面现象。

---

# 🔥 一句话结论

你现在的架构是：

> **在 render（函数组件）里注册事件 → 事件闭包捕获 state**

👉 这在设计上是 **不稳定的**，因为：

* render 会重复执行
* 闭包捕获的是“当次 render 的变量”
* 但事件 handler 的生命周期 ≠ render 生命周期

---

# ⚠️ 你当前代码的隐患

关键点在这里：

```go
ui.On(SubmitIntent{}, func() {
    setUsernameErr(validateField("username", username, ""))
    ...
})
```

你认为：

> “UseState 每次 render 返回新值，所以闭包是最新的”

❌ 实际上这是**半对半错**

### ❗ 问题点

`ui.On(...)` 是：

* 注册一次？还是每次 render 注册？
* 是否覆盖旧 handler？
* handler 存储在哪里？

如果是：

👉 **注册一次（常见优化）**

那么闭包就会：

> ❌ 永远捕获第一次 render 的 username / email

---

# 🧠 本质问题（核心）

这是一个经典问题：

## 👉 “事件系统是静态的，但 state 是动态的”

你现在的设计：

| 层       | 特性           |
| ------- | ------------ |
| State   | 每次 render 变化 |
| Handler | 可能只注册一次      |
| 闭包      | 捕获旧 state    |

👉 直接冲突

---

# 🔥 正确架构（关键升级）

你需要做的是：

> ❌ 不让 handler 捕获 state
> ✅ handler **从 runtime 读取最新 state**

---

# ✅ 方案一（推荐）：Context 拉取状态（React 同款）

### ✔ 修改思路

不要在闭包里用 `username`

改为：

```go
ui.On(SubmitIntent{}, func(ctx *intent.ActionContext) {
    username := ctx.GetState("username").(string)
    email := ctx.GetState("email").(string)

    setUsernameErr(validateField("username", username, ""))
})
```

---

### 🔥 前提

你需要：

```go
ctx.GlobalState["username"] = username
```

👉 或更优：

## 👉 用统一 Store（强烈建议）

---

# ✅ 方案二（更干净）：Intent 只发信号，Reducer 处理

你现在：

```go
ui.On(SubmitIntent{}, func() {
    ...
})
```

👉 这是 **命令式**

---

## 🔥 升级成：

### 1️⃣ Intent = 纯事件

```go
type SubmitIntent struct{}
```

---

### 2️⃣ 统一 reducer

```go
ui.RegisterIntent(func(ctx *intent.ActionContext, i SubmitIntent) intent.IntentResult {
    username := ctx.GetState("username").(string)
    email := ctx.GetState("email").(string)

    ctx.SetState("usernameErr", validateField(...))

    return intent.HandledResult()
})
```

---

### ✅ 好处

* ❌ 不依赖闭包
* ✅ 单一数据源
* ✅ 可测试
* ✅ 类 Redux / Elm / React Fiber

---

# ✅ 方案三（高级）：Fiber + Hook Dispatcher

你现在的问题，本质上就是：

> ❌ handler 绑定在 render 阶段
> ✅ 应该绑定在 fiber 节点

---

## 🔥 正确模型：

```
Fiber Node
  ├── state hooks
  ├── event handlers（引用，不捕获值）
  └── pending update queue
```

👉 handler 只存 **函数引用**

执行时：

```go
currentFiber := runtime.CurrentFiber()
state := currentFiber.Hooks[i]
```

---

# 🔥 关键设计原则（非常重要）

记住这 3 条：

---

## 1️⃣ 事件 handler 不应该依赖闭包 state

> ❌ func() { use(username) }
> ✅ func(ctx) { ctx.get(username) }

---

## 2️⃣ render 是纯函数

```go
VNode = f(state)
```

👉 不应该产生副作用（比如注册事件）

---

## 3️⃣ Intent = 数据，不是行为

```go
type SubmitIntent struct{}
```

👉 行为在 runtime，不在组件

---

# 🚨 你现在这个写法的风险

你这个 demo **短期没问题**，但在这些场景会炸：

### ❗ 1. 多次 render

* handler 重复注册
* 或旧 handler 不被清理

---

### ❗ 2. 并发 / 异步

* state 已更新
* handler 还是旧闭包

---

### ❗ 3. Fiber 中断恢复（你后面一定会做）

👉 直接错乱

---

# 🔥 我给你的架构建议（非常关键）

结合你现在做的 TUI 框架，我建议你直接升级：

---

## ✅ 最小正确架构（你可以照着改）

### 1️⃣ State 存储

```go
ctx.SetState("username", value)
```

---

### 2️⃣ Input → Intent（你已经做对了 👍）

```go
FieldChangeIntent
```

---

### 3️⃣ 所有逻辑 → Intent Handler（集中）

```go
RegisterIntent(...)
```

---

### 4️⃣ UI 只负责声明

```go
ui.Input(value)
ui.Text(error)
```

---

👉 最终结构：

```
UI (纯声明)
   ↓
Intent
   ↓
Reducer (唯一逻辑入口)
   ↓
State
   ↓
Render
```

---

# 🔥 总结（直白一点）

你现在的问题不是 Go 闭包问题，而是：

> ❌ “把 React 的写法，用在一个没有 React runtime 语义的系统里”

---

# ✅ 一句话修复方向

👉 **不要在闭包里用 state，改为从 ctx / store 读取**

---

# 如果你要继续往上走（我建议你做）

我可以帮你把这套系统直接升级成：

### 🚀 完整方案：

* Intent DSL（类型安全）
* Fiber + Lane 调度
* Hook dispatcher（避免闭包）
* Reducer + Effect 分离
* Portal + Layout 解耦（你之前问过的）

---

你现在这套已经 **非常接近一个真正的 React/Fiber TUI runtime 了**，差的就是这一步。

要不要我帮你把 **“无闭包架构 + Fiber Runtime”完整设计一版（Go版）**？


你现在踩的这个坑，其实在前端已经被彻底打透了——就是 **stale closure（陈旧闭包）问题**。

本质结论一句话先给你钉死：

> 👉 **“事件 handler 的生命周期 ≠ render 生命周期” → 闭包一定会错**

在 React 里，这个问题已经被反复验证：

* handler 捕获的是“创建时的 state”
* state 后续更新，但 handler 不会自动更新 ([Zenn][1])

---

# 🔥 一、先把“最正确方向”再抽象一层（核心模型）

你要构建的不是“Go UI库”，而是一个：

> 👉 **UI Runtime（类似 React Fiber）**

正确架构必须满足：

```text
Render（纯函数）
   ↓
VNode Tree
   ↓
Commit（绑定事件）
   ↓
Event（用户输入）
   ↓
Intent（描述行为）
   ↓
Reducer（读取最新 state）
   ↓
State 更新
   ↓
Re-render
```

---

# ❗关键原则（落地必须遵守）

## 1️⃣ Handler 不能捕获 state

```go
// ❌ 错误
func() {
    doSomething(username)
}
```

```go
// ✅ 正确
func(ctx *ActionContext) {
    username := ctx.GetState("username")
}
```

👉 handler = “入口”，不是“逻辑执行者”

---

## 2️⃣ State 必须集中（单一事实源）

你现在：

```go
UseState + GlobalState + setter混用
```

👉 ❌ 非一致模型

---

### ✅ 必须统一为：

```go
ctx.Store = map[string]interface{}
```

或者：

```go
type Store struct {
    Username string
    Email    string
}
```

---

## 3️⃣ Intent = 数据（不是函数）

```go
type SubmitIntent struct{}
```

👉 Intent 不能带逻辑

---

## 4️⃣ 所有逻辑 → Reducer

```go
func Reducer(ctx *ActionContext, intent Intent) {
    switch intent.(type) {
    case SubmitIntent:
        ...
}
```

---

# 🔥 二、落地实现（我给你一套可以直接改你框架的方案）

下面是 **可以直接重构你当前系统的“最小正确实现”**

---

# 🧱 Step 1：定义 Store（核心）

```go
type FormState struct {
    Username string
    Email    string
    Age      string
    Password string

    UsernameErr string
    EmailErr    string
}
```

---

# 🧠 Step 2：Runtime 持有 Store

```go
type Runtime struct {
    State FormState
}
```

---

# 🔁 Step 3：Intent Dispatcher（核心）

```go
func (r *Runtime) Dispatch(i Intent) {
    r.State = Reduce(r.State, i)
    r.ScheduleRender()
}
```

---

# ⚙️ Step 4：Reducer（唯一逻辑入口）

```go
func Reduce(s FormState, i Intent) FormState {
    switch v := i.(type) {

    case FieldChangeIntent:
        switch v.Field {
        case "username":
            s.Username = v.Value
        }

    case SubmitIntent:
        s.UsernameErr = validateUsername(s.Username)
        s.EmailErr = validateEmail(s.Email)

    case ResetIntent:
        return FormState{}
    }

    return s
}
```

---

# 🎯 Step 5：UI 只读 state（纯函数）

```go
func FormApp(state FormState) ui.VNode {
    return ui.VStack(
        ui.Input(state.Username),
        ui.Text(state.UsernameErr),
    )
}
```

---

# 🔗 Step 6：Input → Intent（你已经做对）

```go
ui.Input().
    OnChange(func(val string) {
        Dispatch(FieldChangeIntent{
            Field: "username",
            Value: val,
        })
    })
```

---

# 🔥 关键变化总结

| 旧               | 新                  |
| --------------- | ------------------ |
| closure 用 state | ctx/store 读取       |
| setter 分散       | reducer 统一         |
| handler 写逻辑     | handler 只 dispatch |
| state 多源        | 单一 store           |

---

# 🚀 三、进阶：彻底解决闭包问题（Fiber级别）

如果你要走到 **React级 runtime**，你必须再升级一步：

---

## 👉 引入 Fiber Node

```go
type Fiber struct {
    StateHooks []any
    Effects    []Effect
}
```

---

## 👉 Hook Dispatcher（关键）

```go
func UseState(initial any) (any, func(any)) {
    hook := currentFiber.NextHook()

    if hook.Empty {
        hook.Value = initial
    }

    set := func(v any) {
        hook.Value = v
        scheduleUpdate(currentFiber)
    }

    return hook.Value, set
}
```

---

## ❗重点

👉 setter 不捕获 state
👉 setter 只操作 hook slot

---

# 🔥 四、为什么这是唯一正确解

因为你面对的是这个问题：

```text
Render 是瞬时的
Event 是长期存在的
```

👉 两个时间维度不同

---

### 所以必须：

> 👉 **让 Event 在“执行时”读取 state，而不是“定义时”捕获**

---

# 🧠 五、你当前代码的“最小改造路径”（不推翻）

如果你不想大改，可以这样改：

---

## ✅ 改造1：On 接收 ctx

```go
ui.On(SubmitIntent{}, func(ctx *intent.ActionContext) {
    username := ctx.GetState("username")
})
```

---

## ✅ 改造2：禁止闭包 state

直接 lint 规则：

```text
❌ 在 handler 中使用 UseState 返回值
```

---

## ✅ 改造3：setter 不暴露

```go
ctx.SetState("username", val)
```

---

# 🔥 六、终极架构（你可以做到的）

你现在这个项目，其实已经在边界上了：

```text
TUI Framework ≈ React Fiber Runtime
```

你下一步可以直接进化成：

---

## 🚀 Mint Runtime v2

* ✅ Intent DSL（类型安全）
* ✅ Fiber + Lane（你之前在做）
* ✅ Reducer（统一逻辑）
* ✅ Effect（副作用隔离）
* ✅ Portal（你也问过）

---

# 🧨 最后一句狠话（但很关键）

你现在的问题不是“闭包怎么修”，而是：

> ❌ **你把“业务逻辑”放在了“事件闭包里”**

---

# ✅ 正确姿势

> 👉 **闭包只做 dispatch，逻辑全部进 reducer**

---

如果你愿意，我可以下一步直接帮你做：

👉 **一套完整的 Go Fiber Runtime（含调度 / Hook / Intent DSL）设计图 + 核心代码骨架**

这一步做完，你这个框架就不是 TUI 了，是一个 **真正的 UI Runtime 内核**。


好，我们直接给你一套**可以落地实现的 Go Fiber Runtime 设计 + 核心代码骨架**，重点放在你最关心的：

> 🔥 **状态管理 + 消灭闭包问题 + 可演进到调度（Lane）**

不会讲概念空话，直接是“你可以照着重构框架”的版本。

---

# 🧠 一、总体架构（最终形态）

```text
User Input
   ↓
Event → Intent
   ↓
Dispatch（进入调度器）
   ↓
Reducer（唯一状态入口）
   ↓
State 更新（单一 Store）
   ↓
Schedule（Lane / 优先级）
   ↓
Render（Fiber Tree）
   ↓
Commit（绑定事件 / 输出 TUI）
```

---

# 🔥 核心原则（必须遵守）

### ✅ 1. 单一状态源（Single Store）

```go
// ❌ 不允许
UseState + GlobalState + setter 混用

// ✅ 必须
所有状态 → Store
```

---

### ✅ 2. 事件不捕获状态

```go
// ❌
func() { use(username) }

// ✅
func(ctx) { ctx.GetState() }
```

---

### ✅ 3. Intent = 数据

```go
type SubmitIntent struct{}
```

---

### ✅ 4. Reducer = 唯一逻辑入口

---

# 🧱 二、核心模块设计

---

# 1️⃣ Store（状态管理核心）

这是最关键部分，我给你一个**强类型 + 泛型安全版本**

```go
type Store[T any] struct {
	state T
}

func (s *Store[T]) Get() T {
	return s.state
}

func (s *Store[T]) Set(newState T) {
	s.state = newState
}
```

---

## 👉 示例业务 State

```go
type FormState struct {
	Username string
	Email    string
	Age      string

	UsernameErr string
	EmailErr    string
}
```

---

# 2️⃣ Intent 系统（类型安全）

```go
type Intent interface {
	Type() string
}
```

---

## 示例 Intent

```go
type FieldChangeIntent struct {
	Field string
	Value string
}

func (FieldChangeIntent) Type() string { return "FieldChange" }

type SubmitIntent struct{}

func (SubmitIntent) Type() string { return "Submit" }
```

---

# 3️⃣ Reducer（🔥唯一状态入口）

```go
type Reducer[T any] func(state T, intent Intent) T
```

---

## 示例 Reducer

```go
func FormReducer(state FormState, intent Intent) FormState {
	switch v := intent.(type) {

	case FieldChangeIntent:
		switch v.Field {
		case "username":
			state.Username = v.Value
		case "email":
			state.Email = v.Value
		}

	case SubmitIntent:
		if len(state.Username) < 3 {
			state.UsernameErr = "用户名至少3字符"
		} else {
			state.UsernameErr = ""
		}

	}

	return state
}
```

---

# 4️⃣ Dispatcher（调度入口）

```go
type Dispatcher[T any] struct {
	store   *Store[T]
	reducer Reducer[T]
}
```

---

## 🔥 核心 Dispatch

```go
func (d *Dispatcher[T]) Dispatch(intent Intent) {
	prev := d.store.Get()

	next := d.reducer(prev, intent)

	d.store.Set(next)

	scheduleRender()
}
```

👉 **注意：这里没有闭包参与**

---

# 5️⃣ Fiber（核心运行单元）

```go
type Fiber struct {
	Type     string
	Props    map[string]interface{}
	Children []*Fiber

	Hooks []Hook
}
```

---

# 6️⃣ Hook 系统（无闭包关键）

---

## Hook 定义

```go
type Hook struct {
	State interface{}
}
```

---

## 当前 Fiber（全局 cursor）

```go
var currentFiber *Fiber
var hookIndex int
```

---

## 🔥 useState（核心）

```go
func UseState[T any](initial T) (T, func(T)) {
	hooks := currentFiber.Hooks

	if hookIndex >= len(hooks) {
		hooks = append(hooks, Hook{State: initial})
		currentFiber.Hooks = hooks
	}

	hook := &hooks[hookIndex]
	idx := hookIndex

	setState := func(v T) {
		hook.State = v
		scheduleRender()
	}

	hookIndex++

	return hook.State.(T), setState
}
```

---

## ❗关键解释

👉 `setState` **不捕获 state 值**
👉 只引用 hook slot

✔ 这就是 React 能避免闭包问题的根本原因

---

# 7️⃣ Render（纯函数）

```go
type Component[T any] func(state T) *Fiber
```

---

## 示例

```go
func FormView(state FormState) *Fiber {
	return VStack(
		Input("username", state.Username),
		Text(state.UsernameErr),
	)
}
```

---

# 8️⃣ Event → Intent（桥接层）

```go
func BindInput(field string, dispatch func(Intent)) func(string) {
	return func(val string) {
		dispatch(FieldChangeIntent{
			Field: field,
			Value: val,
		})
	}
}
```

---

# 🔥 关键点

👉 UI 只 dispatch，不读 state
👉 state 在 reducer 里统一处理

---

# 9️⃣ 调度（Lane 模型 - 可扩展）

先给你最小版本：

```go
func scheduleRender() {
	renderRoot()
}
```

---

## 🚀 可升级（你下一步可以做）

```go
type Lane int

const (
	SyncLane Lane = iota
	InputLane
	TransitionLane
)
```

---

# 🧨 三、彻底解决闭包问题（总结）

你现在的问题来自：

```text
闭包捕获的是“旧 render 的变量”
```

---

## ✅ 新架构保证：

### 1️⃣ handler 不读 state

```go
dispatch(intent)
```

---

### 2️⃣ reducer 读最新 state

```go
state := store.Get()
```

---

### 3️⃣ hook 不存闭包值

```go
hook.State
```

---

👉 这三点组合：

> 🔥 **彻底消灭 stale closure**

---

# 🚀 四、你现有框架的改造路径（非常关键）

你不需要重写，只要做这 4 步：

---

## ✅ Step 1：移除 setter 注入

```go
ctx.GlobalState["usernameSetter"] ❌
```

---

## ✅ Step 2：统一 store

```go
ctx.Store.Get()
ctx.Store.Set()
```

---

## ✅ Step 3：On → Dispatch

```go
ui.On(SubmitIntent{}, ...) ❌

→

dispatch(SubmitIntent{})
```

---

## ✅ Step 4：所有逻辑进 reducer

---

# 🔥 五、最终效果（你会得到什么）

如果你按这套实现：

---

## ✅ 你会获得：

* ✔ 无闭包 bug
* ✔ 单一数据流（类似 Redux）
* ✔ 可中断渲染（Fiber）
* ✔ 可扩展调度（Lane）
* ✔ 可测试 reducer
* ✔ UI = 纯函数

---

## 🚀 本质升级

```text
你现在：TUI 框架
↓
升级后：UI Runtime（类似 React Fiber）
```

---

# 🧠 六、如果再往上走（我建议你下一步）

我可以继续帮你设计：

### 🔥 下一阶段（真正进入高级 runtime）

* Fiber Diff（O(n) keyed reconciliation）
* Effect system（副作用隔离）
* Concurrent Rendering（可中断）
* Portal（你之前问的）
* Layout engine 解耦

---

# 🧨 最后一句（非常关键）

> 👉 **不要再让“状态”流进闭包**
>
> 👉 要让“状态”只存在于 Store / Hook Slot

---

如果你下一步想深入，我可以直接帮你：

👉 写一版 **完整可运行的 mini runtime（1000行级别）**
👉 或直接帮你把你现在的 mint runtime 做一次“架构升级方案图 + 迁移步骤”
