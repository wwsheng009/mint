# Mint UI 架构重构方案

**创建时间**: 2026-03-03
**版本**: v1.0
**目标**: 构建无闭包依赖、状态单一、基于 Reducer 的声明式 UI Runtime

---

## 📋 目录

- [一、执行摘要](#一执行摘要)
- [二、当前问题分析](#二当前问题分析)
- [三、设计原则](#三设计原则)
- [四、重构目标架构](#四重构目标架构)
- [五、阶段划分](#五阶段划分)
  - [Phase 1: 消除闭包状态依赖](#phase-1-消除闭包状态依赖)
  - [Phase 2: Store + Reducer 统一](#phase-2-store--reducer-统一)
  - [Phase 3: Fiber Runtime 完善](#phase-3-fiber-runtime-完善)
  - [Phase 4: 类型安全 Intent DSL](#phase-4-类型安全-intent-dsl)
  - [Phase 5: Lane 调度](#phase-5-lane-调度)
- [六、详细技术方案](#六详细技术方案)
- [七、迁移指南](#七迁移指南)
- [八、风险评估](#八风险评估)

---

## 一、执行摘要

### 1.1 问题本质

当前 Mint UI 系统面临的核心问题不是"Go 闭包 bug"，而是 **"事件系统生命周期与渲染生命周期不匹配"** 导致的架构不稳定：

```text
❌ 当前模型（不稳定）：
  Render 函数组件 → 多次执行 → 闭包捕获旧 state
      ↓
  事件 handler（可能只注册一次）→ 引用过期闭包
```

导致后果：
- 状态更新后 UI 不刷新（闭包引用过期）
- 多源状态冲突（UseState + GlobalState + setter）
- 无法支持 Fiber 中断渲染、时间旅行调试等高级特性

---

### 1.2 设计愿景

将 Mint UI 从"组件驱动的 TUI 库"升级为"**状态机驱动的 UI Runtime**"：

```text
✅ 目标模型（稳定）：
  State (单一 Store) → VNode (纯声明) → Render
       ↑                       ↓
    Reducer (逻辑集中) ← Intent (纯数据)
       ↑
  Dispatcher
       ↑
  Event → Instance (只缓存输入)
```

核心转变：
- **从**: 事件 handler 捕获 state 执行逻辑
- **到**: 事件只发 Intent，Redurcer 读取最新 state 处理逻辑

---

### 1.3 关键收益

| 收益项 | 当前状态 | 目标状态 |
|--------|---------|---------|
| 状态一致性 | 多源冲突 | 单一 Store |
| 闭包依赖 | 存在过期闭包风险 | 完全无依赖 |
| 可测试性 | 需要闭包模拟 | Intent 独立测试 |
| 可调试性 | 状态分散 | 可时间旅行 |
| 可扩展性 | 难以支持 Lane | 天然支持调度 |

---

## 二、当前问题分析

### 2.1 代码现状统计

```bash
# 大量示例仍使用闭包模式
$ grep -r "OnClick(func()" examples/
208 matches  # ← 闭包 API 滥用

$ grep -r "OnPress(" examples/
3 matches   # ← Intent API 使用不足
```

---

### 2.2 核心问题清单

#### ❌ 问题 1: 闭包捕获过期变量

```go
// ❌ 当前代码（不稳定）
func App() ui.VNode {
    username, setUsername := ui.UseStateString("")  // 每次渲染是新变量

    ui.On(SubmitIntent{}, func() {
        // 👉 问题：如果 handler 只在首次注册，这里闭包永远捕获首次的 username
        setUsernameErr(validate(username))
    })

    // ❌ 如果 ui.On 有去重机制，后续渲染不会更新闭包引用
    return ui.VStack(...)
}
```

**根本原因**：
- 渲染循环每次执行 → UseState 返回新闭包引用
- 事件系统可能只注册一次 → handler 持有首次渲染的闭包
- 用户输入 → 触发旧 handler → 调用旧 setter → 状态更新但 UI 不刷新

---

#### ❌ 问题 2: 多源状态冲突

```go
// ❌ 当前代码（状态分散）
func App() ui.VNode {
    // 源1：useState 返回的 setter
    username, setUsername := ui.UseStateString("")

    // 源2：GlobalState
    ctx := ui.GetCurrentContext()
    ctx.GlobalState["username"] = username  // 用于 handler 赋值

    // 源3：ctx.SetState
    ctx.SetState("username", value)

    // 👉 问题：三个来源互不感知，易冲突
}
```

---

#### ❌ 问题 3: Handler 逻辑分散

```go
// ❌ 当前代码（逻辑分散）
ui.On(SubmitIntent{}, func() {
    // 验证逻辑在这里
    if len(username) < 3 {
        setUsernameErr("用户名至少3字符")
    }
})

ui.On(ResetIntent{}, func() {
    // 清空逻辑在这里
    setUsername("")
    setEmail("")
})

// 👉 问题：逻辑分散在各个 handler，难以统一管理和测试
```

---

#### ❌ 问题 4: 组件内 RegisterIntent

```go
// ❌ 当前代码（错误做法）
func App() ui.VNode {
    username, setUsername := ui.UseStateString("")

    // ❌ 每次渲染都注册新 handler（或重复注册）
    ui.RegisterIntent(func(ctx *intent.ActionContext, i FieldChangeIntent) intent.IntentResult {
        setUsername(i.Value)  // 闭包引用旧 setter
        return intent.HandledResult()
    })

    return ui.VStack(...)
}
```

**问题**：
- RegisterIntent 注册的是持久 handler
- 闭包捕获的是首次渲染时的 setter
- 后续渲染 setter 已更新，但 handler 不知道

---

#### ❌ 问题 5: StateKey[T] 与字符串键混用

```go
// ❌ 当前代码（类型不一致）
var username = intent.StateKey[string]("username")

func App() ui.VNode {
    // 使用 StateKey
    input.ForField(intent.ForField(username))

    // ❌ 但 handler 仍用字符串
    ui.RegisterIntent(func(ctx *intent.ActionContext, i FieldChangeIntent) intent.IntentResult {
        switch i.Field {
        case "username":  // 字符串！
            // ❌ 拼写错误，编译器无法检查
            setter, _ := ctx.GetState("username")
        }
    })
}
```

---

### 2.3 影响范围评估

| 影响层级 | 影响程度 | 说明 |
|---------|---------|------|
| **核心架构** | 🔴 高 | 需要重构状态管理机制 |
| **组件 API** | 🟡 中 | 需要调整 Builder API |
| **示例代码** | 🔴 高 | 大量示例需要迁移 |
| **文档** | 🟡 中 | 需要更新最佳实践 |
| **测试** | 🟢 低 | 有组件测试作为安全网 |

---

## 三、设计原则

### 3.1 核心设计原则（不可违背）

#### 🔴 原则 1: 事件 handler 不得捕获闭包状态

```go
// ❌ 禁止
func() {
    doSomething(username)  // 闭包捕获 username
}

// ✅ 必须从 Context/Store 读取
func(ctx *ActionContext) {
    username := ctx.GetState("username")
}
```

**原理**：
- 事件生命周期 ≠ 渲染生命周期
- 闭包捕获的是"定义时的值"，不是"执行时的值"
- 必须让事件"执行时"动态读取最新状态

---

#### 🔴 原则 2: State 单一真相源

```go
// ❌ 禁止多源状态
UseState + GlobalState + setter + ctx.SetState 混用

// ✅ 必须统一为单源
type Store[T any] struct { state T }
```

**原理**：
- 多源状态 → 无法预测哪个来源更新生效
- 单一来源 → 可预测、可调试、可回溯

---

#### 🔴 原则 3: Intent = 纯数据

```go
// ❌ 禁止 Intent 携带逻辑
type SubmitIntent struct {
    Handler func() error  // 不要放函数
}

// ✅ Intent 只描述意图
type SubmitIntent struct{}
```

**原理**：
- Intent 应该是可序列化、可传输的消息
- 逻辑在 Reducer，不在 Intent

---

#### 🔴 原则 4: Reducer = 唯一逻辑入口

```go
// ❌ 禁止逻辑分散
ui.On(A, func() { 处理 })
ui.On(B, func() { 处理 })
ui.On(C, func() { 处理 })

// ✅ 统一 Reducer
func Reduce(state, intent) newState {
    switch intent {
    case A: 处理
    case B: 处理
    case C: 处理
    }
}
```

**原理**：
- 所有状态变更集中处理
- 易于测试、调试、插件化

---

### 3.2 实施原则（可调整）

#### 🟡 原则 5: 向后兼容（短期）

- MVP 组件 API（ForField）保持不变
- FieldChangeIntent 保持不变
- 仅迁移示例代码和 handler 模式

---

#### 🟡 原则 6: 渐进式重构（中期）

- 先修复核心问题（闭包）
- 再引入架构升级（Store + Reducer）
- 最后完善高级特性（Lane）

---

#### 🟢 原则 7: 类型安全优先（长期）

- 全面采用 StateKey[T]
- 消除字符串键
- 编译期检查状态访问

---

## 四、重构目标架构

### 4.1 最终数据流

```text
┌─────────────────────────────────────────────────────────────────────┐
│                        用户交互层 (Event)                            │
│  Instance (临时缓存输入) → Intent (携带数据)                         │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      调度层 (Dispatcher)                              │
│  Dispatch(intent) → Lane 排队 → Scheduler                           │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      逻辑层 (Reducer)                                │
│  Reducer(state, intent) → newState                                 │
│  📍 唯一状态变更入口：验证、业务逻辑、副作用管理                      │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      状态层 (Store)                                  │
│  Store[T] <- 单一真相源                                              │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      渲染层 (Fiber)                                  │
│  Fiber Tree (VNode diff + minimal patch)                            │
│  Hook State (无闭包依赖)                                             │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      输出层 (Commit)                                 │
│  TUI 屏幕更新                                                         │
└─────────────────────────────────────────────────────────────────────┘
```

---

### 4.2 核心模块职责

| 模块 | 职责 | 关键原则 |
|------|------|---------|
| **Event** | 收集用户输入，发射 Intent | 不保存状态 |
| **Dispatcher** | Intent 队列、Lane 调度 | 优先级管理 |
| **Reducer** | 状态变更、业务逻辑 | 纯函数（无副作用） |
| **Store** | 持有唯一状态 | 单一真相源 |
| **Fiber** | 虚拟 DOM、Hook 管理 | 无闭包依赖 |
| **Component** | 声明式 UI | 纯函数 |

---

### 4.3 模块间依赖关系

```text
Event → Dispatcher (单向)
Dispatcher → Reducer (单向)
Reducer → Store (单向)
Store → Fiber → Component (单向渲染)
Component (只读) → 无写入依赖
```

**关键特性**：
- 单向数据流
- 无循环依赖
- 易于测试和并行

---

## 五、阶段划分

### Phase 1: 消除闭包状态依赖 🎯 高优先级

**目标**: 修复当前最严重的闭包过期问题，无需大改架构

---

#### 1.1 核心改动

**改动 A: ui.On 支持 Context 参数**

```go
// ui/intent.go

// ❌ 旧 API（捕获闭包）
func On[T interface{ IntentType() string; StayPressed() bool }](
    intentType T,
    handler func(),
) { ... }

// ✅ 新 API（从 Context 读取）
func On[T interface{ IntentType() string; StayPressed() bool }](
    intentType T,
    handler func(*intent.ActionContext),
) {
    key := intentType.IntentType()
    if _, loaded := registeredHandlers.LoadOrStore(key, true); loaded {
        return
    }
    rtui.RegisterIntent(func(ctx *intent.ActionContext, i T) intent.IntentResult {
        handler(ctx)  // ← 传递 Context
        return intent.HandledResult()
    })
}
```

---

**改动 B: ActionContext 扩展**

```go
// runtime/intent/action_context.go

type ActionContext struct {
    context.Context
    Source      string      // 组件 key
    Timestamp   time.Time

    // ✅ 新增：访问 GlobalState 的便捷方法
    GlobalStore map[string]interface{}
}

// ✅ 新增类型安全的 State 访问
func (ctx *ActionContext) GetStringState(key string, defaultValue string) string {
    if v, ok := ctx.GlobalStore[key]; ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return defaultValue
}

func (ctx *ActionContext) GetIntState(key string, defaultValue int) int {
    if v, ok := ctx.GlobalStore[key]; ok {
        if i, ok := v.(int); ok {
            return i
        }
    }
    return defaultValue
}
```

---

#### 1.2 示例迁移

```go
// ❌ 旧写法（闭包捕获）
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    ui.On(SimpleIncrementIntent{}, func() {
        count, _, _ = ui.UseStateInt(0)  // ❌ 重复调用，顺序不可预测
        setCount(count + 1)
    })

    count, _, _ = ui.UseStateInt(0)
    return ui.Text(fmt.Sprintf("Count: %d", count))
}

// ✅ 新写法（从 Context 读取）
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    // ✅ 保存到 GlobalStore 供 handler 读取
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalStore["counter_count"] = count
        ctx.GlobalStore["counter_setCount"] = setCount
    }

    ui.On(SimpleIncrementIntent{}, func(ctx *intent.ActionContext) {
        // ✅ 从 Context 读取最新值
        currentCount := ctx.GetIntState("counter_count", 0)
        setCountFn, _ := ctx.GlobalStore["counter_setCount"].(func(func(int) int))

        if setCountFn != nil {
            setCountFn(func(c int) int { return c + 1 })
        }
    })

    count, _, _ = ui.UseStateInt(0)
    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

---

**更简洁写法（推荐）**：

```go
// ✅ 简化：直接通过 StateKey 访问
var CounterCount = intent.StateKey[int]("counter_count")

func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    // ✅ 使用 StateKey 保存
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalStore[CounterCount.String()] = count
        ctx.GlobalStore[CounterCount.String()+"_setter"] = setCount
    }

    ui.On(SimpleIncrementIntent{}, func(ctx *intent.ActionContext) {
        // ✅ 类型安全
        setCountFn, _ := ctx.GlobalStore[CounterCount.String()+"_setter"].(func(func(int) int))
        if setCountFn != nil {
            setCountFn(func(c int) int { return c + 1 })
        }
    })

    count, _, _ = ui.UseStateInt(0)
    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

---

#### 1.3 状态总结

| 指标 | 状态 |
|------|------|
| 核心代码改动 | 🟢 小（仅需扩展接口） |
| 组件 API 兼容性 | 🟢 完全兼容（旧 API 仍可用） |
| 示例代码迁移 | 🟡 中（约 50+ 文件需更新） |
| 架构复杂度 | 🟢 不变 |
| **预期效果** | **解决闭包过期问题** |

---

### Phase 2: Store + Reducer 统一 🎯 核心重构

**目标**: 引入单一状态源和统一逻辑入口，消除多源状态冲突

---

#### 2.1 核心模块设计

**模块 A: 泛型 Store**

```go
// runtime/store/store.go

package store

import "github.com/wwsheng009/mint/runtime/intent"

// Store[T] 类型安全的状态容器
type Store[T any] struct {
    state    T
    onChange func(T)
}

func NewStore[T any](initial T) *Store[T] {
    return &Store[T]{
        state: initial,
    }
}

func (s *Store[T]) Get() T {
    return s.state
}

func (s *Store[T]) Set(next T) {
    s.state = next
    if s.onChange != nil {
        s.onChange(next)
    }
}

// Subscribe 订阅状态变更
func (s *Store[T]) Subscribe(callback func(T)) {
    s.onChange = callback
}
```

---

**模块 B: Reducer 定义**

```go
// runtime/reducer/reducer.go

package reducer

import (
    "github.com/wwsheng009/mint/runtime/intent"
)

// Reducer[T] 状态变更函数
type Reducer[T any] func(state T, intent intent.Intent) T

// Compose 组合多个 Reducer
func Compose[T any](reducers ...Reducer[T]) Reducer[T] {
    return func(state T, intentVal intent.Intent) T {
        for _, r := range reducers {
            state = r(state, intentVal)
        }
        return state
    }
}
```

---

**模块 C: Dispatcher**

```go
// runtime/dispatcher/dispatcher.go

package dispatcher

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/runtime/scheduler"
)

type Dispatcher[T any] struct {
    store    *store.Store[T]
    reducer  reducer.Reducer[T]
    scheduler *scheduler.Scheduler
}

func NewDispatcher[T any](
    st *store.Store[T],
    red reducer.Reducer[T],
    sched *scheduler.Scheduler,
) *Dispatcher[T] {
    return &Dispatcher[T]{
        store:    st,
        reducer:  red,
        scheduler: sched,
    }
}

func (d *Dispatcher[T]) Dispatch(intentVal intent.Intent) {
    prev := d.store.Get()
    next := d.reducer(prev, intentVal)
    d.store.Set(next)

    // 触发调度
    d.scheduler.ScheduleRender()
}
```

---

**模块 D: 应用状态示例**

```go
// examples/store_demo/app_state.go

package main

import "github.com/wwsheng009/mint/runtime/intent"

type FormState struct {
    Username   string
    Email      string
    Age        int
    UsernameErr string
    EmailErr    string
}

// FormReducer 表单逻辑集中处理
var FormReducer = func(state FormState, intentVal intent.Intent) FormState {
    switch v := intentVal.(type) {
    case intent.FieldChangeIntent:
        switch v.Field {
        case "username":
            state.Username = v.Value
            // 实时验证
            if len(state.Username) < 3 {
                state.UsernameErr = "用户名至少3字符"
            } else {
                state.UsernameErr = ""
            }
        case "email":
            state.Email = v.Value
        case "age":
            if age, err := strconv.Atoi(v.Value); err == nil {
                state.Age = age
            }
        }

    case SubmitIntent:
        // 提交时验证所有字段
        state.UsernameErr = validateUsername(state.Username)
        state.EmailErr = validateEmail(state.Email)

    case ResetIntent:
        return FormState{}
    }

    return state
}
```

---

#### 2.2 Runtime 集成

```go
// runtime/runtime.go

type Runtime[T any] struct {
    store      *store.Store[T]
    dispatcher *dispatcher.Dispatcher[T]
    rootFiber  *Fiber
    view       func(T) ui.VNode  // 纯函数
}

func NewRuntime[T any](initial T, view func(T) ui.VNode) *Runtime[T] {
    st := store.NewStore(initial)

    red := FormReducer  // 应用自定义 Reducer

    sched := scheduler.NewScheduler()

    disp := dispatcher.NewDispatcher(st, red, sched)

    rt := &Runtime[T]{
        store:      st,
        dispatcher: disp,
        rootFiber:  &Fiber{},
        view:       view,
    }

    // 订阅状态变更，触发渲染
    st.Subscribe(func(state T) {
        renderRoot(rt, state)
    })

    return rt
}

func (rt *Runtime[T]) Dispatch(intentVal intent.Intent) {
    rt.dispatcher.Dispatch(intentVal)
}

func (rt *Runtime[T]) GetState() T {
    return rt.store.Get()
}

// renderRoot 渲染 Fiber 树
func renderRoot[T any](rt *Runtime[T], state T) {
    // 清空 hooks
    rt.rootFiber.Hooks = nil
    currentFiber = rt.rootFiber
    hookIndex = 0

    // 执行视图函数
    vnode := rt.view(state)

    // diff & patch
    reconcile(rt.rootFiber, vnode)
}
```

---

#### 2.3 组件迁移示例

**旧写法（逻辑分散）**：

```go
func OldForm() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")
    usernameErr, setUsernameErr := ui.UseStateString("")

    // ❌ 验证逻辑分散在 handler
    ui.On(SubmitIntent{}, func() {
        if len(username) < 3 {
            setUsernameErr("用户名至少3字符")
        }
        if !validateEmail(email) {
            setEmailErr("邮箱格式错误")
        }
    })

    ui.On(ResetIntent{}, func() {
        setUsername("")
        setEmail("")
    })

    return ui.VStack(
        input.ForField("username").Value(username),
        text.Text(usernameErr),
        input.ForField("email").Value(email),
        text.Text(usernameErr),
    )
}
```

**新写法（逻辑集中）**：

```go
// ✅ 定义状态
var formState = NewStore(FormState{})

// ✅ 定义组件函数（纯函数，无状态）
func NewForm(state FormState) ui.VNode {
    return ui.VStack(
        input.ForField("username").Value(state.Username),
        text.Text(state.UsernameErr),
        input.ForField("email").Value(state.Email),
        text.Text(state.EmailErr),
    )
}

// ✅ 初始化 Runtime
func main() {
    runtime := NewRuntime(FormState{}, NewForm)

    // Runtime 自动处理所有状态更新和渲染
    ui.RunApp(runtime)
}
```

---

#### 2.4 状态总结

| 指标 | 状态 |
|------|------|
| 核心代码改动 | 🟡 中（新增模块，大改调用方式） |
| 组件 API 兼容性 | 🟡 部分兼容（UseState 仍可用，但推荐 Store） |
| 示例代码迁移 | 🟡 中（复杂组件需重写） |
| 架构复杂度 | 🟡 增加（引入 Store/Reducer） |
| **预期效果** | **消除多源冲突、逻辑集中、可测试性提升** |

---

### Phase 3: Fiber Runtime 完善 🎯 高级特性

**目标**: 完善 Fiber、Hook 机制，支持可中断渲染和时间旅行

---

#### 3.1 核心模块

**模块 A: Fiber 节点**

```go
// runtime/fiber/fiber.go

type FiberNode struct {
    // 基础信息
    Type     string
    Key      string
    Props    map[string]interface{}

    // 树结构
   Parent    *FiberNode
    Child     *FiberNode
    Sibling   *FiberNode
    Alternate *FiberNode  // 双缓存

    // Hooks
    Hooks     []Hook
    Effect    []Effect

    // 更新队列
    PendingUpdateQueue []Update
}

type Hook struct {
    State interface{}
}

type Effect struct {
    Create    func()
    Destroy   func()
    Deps      []interface{}
}

type Update struct {
    Intent   intent.Intent
    Lane     Lane
}
```

---

**模块 B: UseState 无闭包实现**

```go
// runtime/hooks/state.go

var currentFiber *FiberNode
var hookIndex int

func UseState[T any](initial T) (T, func(T)) {
    if currentFiber == nil {
        panic("UseState can only be called during render")
    }

    hooks := currentFiber.Hooks

    // 初始化 hook
    if hookIndex >= len(hooks) {
        hooks = append(hooks, Hook{State: initial})
        currentFiber.Hooks = hooks
    }

    hook := &hooks[hookIndex]
    idx := hookIndex  // 捕获索引，不是值

    // ✅ 关键：setState 不捕获 state 值，只引用 hook slot
    setState := func(v T) {
        hook.State = v
        scheduleUpdate(currentFiber)
    }

    hookIndex++
    return hook.State.(T), setState
}
```

---

**模块 C: VNode Diff 算法**

```go
// runtime/reconciler/diff.go

func Reconcile(current *FiberNode, vnode ui.VNode) *FiberNode {
    if current == nil {
        return mountFiber(vnode)
    }

    if !isSameType(current, vnode) {
        // 类型不同，整树替换
        unmountFiber(current)
        return mountFiber(vnode)
    }

    // 类型相同，复用节点，diff props
    current.Props = vnode.Props()
    diffProps(currentProps, newProps)

    // Diff 子节点
    reconcileChildren(current, vnode.Children())

    return current
}

func isSameType(fiber *FiberNode, vnode ui.VNode) bool {
    return fiber.Type == vnode.Type()
}
```

---

**模块 D: 时间旅行调试**

```go
// runtime/debug/timetravel.go

type TimeTravelDebugger[T any] struct {
    history   []T
    currentIndex int
    store     *store.Store[T]
}

func NewTimeTravelDebugger[T any](st *store.Store[T]) *TimeTravelDebugger[T] {
    dbg := &TimeTravelDebugger[T]{
        history:   []T{st.Get()},
        store:     st,
    }
    st.Subscribe(func(state T) {
        dbg.history = append(dbg.history, state)
    })
    return dbg
}

func (dbg *TimeTravelDebugger[T]) JumpTo(index int) {
    if index >= 0 && index < len(dbg.history) {
        state := dbg.history[index]
        dbg.currentIndex = index
        dbg.store.Set(state)  // 回滚状态
    }
}

func (dbg *TimeTravelDebugger[T]) Undo() {
    if dbg.currentIndex > 0 {
        dbg.JumpTo(dbg.currentIndex - 1)
    }
}

func (dbg *TimeTravelDebugger[T]) Redo() {
    if dbg.currentIndex < len(dbg.history)-1 {
        dbg.JumpTo(dbg.currentIndex + 1)
    }
}
```

---

#### 3.2 状态总结

| 指标 | 状态 |
|------|------|
| 核心代码改动 | 🔴 高（Fiber 完整实现） |
| 组件 API 兼容性 | 🟢 完全兼容 |
| 示例代码迁移 | 🟢 无需改动 |
| 架构复杂度 | 🔴 增加（引入 Fiber/Diff） |
| **预期效果** | **支持可中断渲染、时间旅行、性能优化** |

---

### Phase 4: 类型安全 Intent DSL 🎯 编译期安全

**目标**: 消除字符串键，全面采用类型安全的 StateKey[T]

---

#### 4.1 核心设计

**模块 A: 状态键定义**

```go
// runtime/intent/state_key.go

package intent

type StateKey[T any] struct {
    name string
}

func (k StateKey[T]) String() string {
    return k.name
}

// 应用状态键（集中的地方）
var (
    Username = StateKey[string]("username")
    Email    = StateKey[string]("email")
    Age      = StateKey[int]("age")
    Agree    = StateKey[bool]("agree")
)
```

---

**模块 B: 类型安全的 FieldChangeIntent**

```go
// runtime/intent/typed_field_change.go

package intent

type TypedFieldChange[T any] struct {
    Key   StateKey[T]
    Value T
}

func (TypedFieldChange[T]) IntentType() string {
    return "FieldChange"
}

// 辅助函数
func SetField[T any](key StateKey[T], value T) Intent {
    return TypedFieldChange[T]{Key: key, Value: value}
}
```

---

**模块 C: 类型安全的 Reducer**

```go
// 示例：类型安全的表单 Reducer

var TypedFormReducer = func(state FormState, intentVal intent.Intent) FormState {
    switch v := intentVal.(type) {
    case TypedFieldChange[string]:
        switch v.Key {
        case Username:
            state.Username = v.Value
            if len(state.Username) < 3 {
                state.UsernameErr = "用户名至少3字符"
            } else {
                state.UsernameErr = ""
            }
        case Email:
            state.Email = v.Value
        }

    case TypedFieldChange[int]:
        if v.Key == Age {
            state.Age = v.Value
        }

    case TypedFieldChange[bool]:
        if v.Key == Agree {
            state.Agree = v.Value
        }

    case SubmitIntent:
        state.UsernameErr = validateUsername(state.Username)
        state.EmailErr = validateEmail(state.Email)

    case ResetIntent:
        return FormState{}
    }

    return state
}
```

---

**模块 D: 组件使用**

```go
// 使用类型安全的 Intent
func TypedForm(state FormState) ui.VNode {
    return ui.VStack(
        input.ForField(Username).Value(state.Username),
        text.Text(state.UsernameErr),

        input.ForField(Email).Value(state.Email),
        text.Text(state.EmailErr),

        input.ForField(Age).Value(fmt.Sprintf("%d", state.Age)),
        checkbox.ForField(Agree).Checked(state.Agree),
    )
}

// 发射 Intent
func OnUsernameChange(value string) Intent {
    return SetField(Username, value)  // ✅ 类型安全，编译期检查
}
```

---

#### 4.2 状态总结

| 指标 | 状态 |
|------|------|
| 核心代码改动 | 🟡 中（扩展 Intent 类型） |
| 组件 API 兼容性 | 🟢 完全兼容（旧 API 仍可用） |
| 示例代码迁移 | 🟡 中（推荐迁移） |
| 架构复杂度 | 🟢 不变 |
| **预期效果** | **编译期类型安全、IDE 支持优化、重构安全** |

---

### Phase 5: Lane 调度 🎯 性能优化

**目标**: 支持优先级调度和可中断渲染

---

#### 5.1 核心设计

**模块 A: Lane 定义**

```go
// runtime/scheduler/lane.go

type Lane uint32

const (
    SyncLane Lane = 1 << iota  // 同步（立即渲染）
    InputLane                   // 输入（高优先级）
    TransitionLane              // 过渡（低优先级，可中断）
    IdleLane                   // 空闲（最低优先级）
)

func MergeLanes(lanes ...Lane) Lane {
    var result Lane
    for _, l := range lanes {
        result |= l
    }
    return result
}

func PickLowestPriorityLane(lanes Lane) Lane {
    // 返回最低优先级的 Lane
    if lanes&IdleLane != 0 {
        return IdleLane
    }
    if lanes&TransitionLane != 0 {
        return TransitionLane
    }
    if lanes&InputLane != 0 {
        return InputLane
    }
    return SyncLane
}
```

---

**模块 B: Intent 分配 Lane**

```go
// 发射 Intent 时指定 Lane
type IntentWithLane struct {
    Intent
    Lane Lane
}

func UserInput(field string, value string) IntentWithLane {
    return IntentWithLane{
        Intent: SetField(field, value),
        Lane:   InputLane,
    }
}

func DataFetch(url string) IntentWithLane {
    return IntentWithLane{
        Intent: FetchDataIntent{URL: url},
        Lane:   TransitionLane,
    }
}
```

---

**模块 C: Scheduler 实现**

```go
// runtime/scheduler/scheduler.go

type Scheduler struct {
    workInProgress *FiberNode
    pendingLanes   Lane
}

func (s *Scheduler) ScheduleRender() {
    if s.pendingLanes == 0 {
        return
    }

    // 选择优先级最高的 Lane
    lane := PickLowestPriorityLane(s.pendingLanes)

    // 渲染
    s.workOnRenderLoop(lane)
}

func (s *Scheduler) workOnRenderLoop(lane Lane) {
    for {
        // 检查是否应该中断（低优先级任务）
        if lane == TransitionLane && shouldYield() {
            return  // 中断，让位给高优先级任务
        }

        // 执行渲染工作
        performUnitOfWork(s.workInProgress)

        // 检查是否完成
        if s.workInProgress == nil {
            break
        }
    }

    // 提交
    commitRoot()
}
```

---

#### 5.2 状态总结

| 指标 | 状态 |
|------|------|
| 核心代码改动 | 🔴 高（完整调度系统） |
| 组件 API 兼容性 | 🟢 完全兼容 |
| 示例代码迁移 | 🟢 无需改动 |
| 架构复杂度 | 🔴 增加 |
| **预期效果** | **优先级调度、避免阻塞、大列表性能提升** |

---

## 六、详细技术方案

### 6.1 数据流详解

#### 完整数据流（Phase 2 之后）

```go
// 步骤 1: 用户输入
user types "a"
→ InputInstance.InsertText("a")
→ inst.value = "a" (临时缓存)

// 步骤 2: Instance 发射 Intent
→ emitter(FieldChangeIntent{Field: "username", Value: "a"})

// 步骤 3: Dispatch 处理
→ dispatcher.Dispatch(FieldChangeIntent{...})
→ prevState := store.Get()
→ nextState := reducer(prevState, intent)  // nextState.Username = "a"
→ store.Set(nextState)

// 步骤 4: 状态变更触发订阅
→ store.onChange(nextState)
→ runtime.renderRoot(nextState)

// 步骤 5: 重新渲染
→ resetHooks(rootFiber)
→ vnode := view(nextState)  // 纯函数
→ reconcile(rootFiber, vnode)
→ diff & patch

// 步骤 6: 同步到 Instance
→ InputInstance.SetProps({value: "a"})
→ inst.value = "a" (State → Instance 单向同步)
```

---

### 6.2 关键难点解决方案

#### 难点 1: Hooks 顺序保持

**问题**: React 要求 hooks 调用顺序必须一致，否则状态错乱

**解决方案**:

```go
// runtime/fiber/fiber.go

type FiberNode struct {
    Hooks []Hook  // 每个 Fiber 维护独立的 hooks 数组
}

var currentFiber *FiberNode
var hookIndex int

// 每次渲染前重置
func renderRoot(fiber *FiberNode) {
    currentFiber = fiber
    hookIndex = 0
    // 不清空 fiber.Hooks！保持顺序
}

// UseState 严格按顺序访问
func UseState[T any](initial T) (T, func(T)) {
    hooks := currentFiber.Hooks

    // 第一次渲染：扩展数组
    if hookIndex >= len(hooks) {
        hooks = append(hooks, Hook{State: initial})
        currentFiber.Hooks = hooks
    }

    hook := &hooks[hookIndex]  // ✅ 严格按顺序访问
    idx := hookIndex

    setState := func(v T) {
        hook.State = v
        scheduleUpdate(currentFiber)
    }

    hookIndex++
    return hook.State.(T), setState
}
```

**强制规则**:

```go
// ❌ 错误：条件调用 hooks
if condition {
    UseState(...)  // 破坏顺序
}

// ✅ 正确：始终调用
UseState(...)  // 即使不用也调用
if condition {
    // 使用逻辑
}
```

---

#### 难点 2: 双缓存 Fiber

**问题**: 避免 DOM 频繁闪烁，实现平滑更新

**解决方案**:

```go
// type FiberNode struct {
//     Alternate *FiberNode  // 当前节点的副本（上一帧）
// }

// 双缓存机制
func commitRoot() {
    // workInProgress 是当前帧
    // current 是上一帧

    // 交换引用
    current = workInProgress
    workInProgress = current.Alternate  // 下一帧复用

    // 提交到屏幕
    commitWork(current)
}

function reconcile(current *FiberNode, vnode VNode) *FiberNode {
    if current == nil {
        return mountFiber(vnode)
    }

    if !isSameType(current, vnode) {
        // 类型不同，整树替换
        unmountFiber(current)
        return mountFiber(vnode)
    }

    // 复用当前节点，克隆到 workInProgress
    workInProgress = cloneFiber(current)
    workInProgress.Alternate = current

    // diff props 和 children
    diffProps(current.Props, vnode.Props())
    reconcileChildren(workInProgress, vnode.Children())

    return workInProgress
}
```

---

#### 难点 3: Effect 依赖追踪

**问题**: 避免不必要的副作用执行

**解决方案**:

```go
type Effect struct {
    Create  func()
    Destroy func()
    Deps    []interface{}
}

func UseEffect(create func(), deps []interface{}) {
    if hookIndex >= len(currentFiber.Hooks) {
        // 第一次渲染，记录 effect
        currentFiber.Hooks = append(currentFiber.Hooks, Hook{
            Effect: &Effect{Create: create, Deps: deps},
        })
        create()
    } else {
        // 后续渲染，比较依赖
        oldEffect := currentFiber.Hooks[hookIndex].Effect
        if !eqDeps(oldEffect.Deps, deps) {
            // 依赖变化，执行销毁 + 创建
            if oldEffect.Destroy != nil {
                oldEffect.Destroy()
            }
            oldEffect.Create = create
            oldEffect.Deps = deps
            create()
        }
    }
}

// 提交阶段执行 effect
func commitEffects(fiber *FiberNode) {
    for _, hook := range fiber.Hooks {
        if hook.Effect != nil && hook.Effect.Create != nil {
            fiber.Effects = append(fiber.Effects, *hook.Effect)
        }
    }
}
```

---

### 6.3 错误处理与边界情况

#### 情况 1: 用户快速连续输入

```go
// 用户快速输入 "ab" → 发射两个 Intent
InputChange("a") → FieldChangeIntent{Field: "username", Value: "a"}
InputChange("b") → FieldChangeIntent{Field: "username", Value: "ab"}

// ✅ 解决方案：Intent 队列 + Lane 优先级
// 两个 Intent 都入队，合并处理后渲染一次
```

---

#### 情况 2: 异步操作

```go
// 数据加载异步完成后更新状态
go func() {
    data := fetchData()
    dispatch(LoadDataSuccess{Data: data})  // ✅ Intent 仍然是纯数据
}()

// ✅ 解决方案：Dispatcher 串行处理
// 每个 Intent 都会触发一次完整的状态更新和渲染
```

---

#### 情况 3: 错误边界

```go
// 提供错误边界组件
type ErrorBoundaryState struct {
    Error error
}

func ErrorBoundary(child ui.VNode) ui.VNode {
    errorState, setError := ui.UseState(ErrorBoundaryState{})

    // 捕获错误
    defer func() {
        if r := recover(); r != nil {
            setError(ErrorBoundaryState{Error: r.(error)})
        }
    }()

    if errorState.Error != nil {
        return text.Text(fmt.Sprintf("Error: %v", errorState.Error))
    }

    return child
}
```

---

## 七、迁移指南

### 7.1 迁移检查清单

#### Phase 1: 消除闭包依赖

- [x] 扩展 `ui.On` 支持 `*ActionContext` 参数
- [x] 扩展 `ActionContext` 增加 `GlobalStore` 访问方法
- [x] 创建迁移文档 `INTENT_HANDLER_MIGRATION.md`
- [x] 移除 `On` 函数的 `StayPressed()` 强制约束（改为可选）
- [x] 迁移示例代码
  - [x] `examples/validation_demo/` - 已迁移使用新 API
  - [x] `examples/test_fiber/` - 已迁移
  - [ ] 其他示例待迁移（参考迁移文档）
- [ ] 添加 E2E 测试，验证闭包问题修复
- [ ] 发布 v1.1.0

---

#### Phase 2: Store + Reducer

- [x] 实现 `runtime/store/store.go`
- [x] 实现 `runtime/reducer/reducer.go`
- [x] 实现 `runtime/statemachine/runtime.go` - AppRuntime 集成 Store + Reducer
- [x] 更新文档
  - [x] 创建 `STORE_REDUCER_GUIDE.md`
  - [ ] 更新 `MVP_MIGRATION_GUIDE.md`
- [x] 创建示例
  - [x] `examples/store_reducer_demo/` - Store + Reducer 完整示例
  - [ ] `examples/store_demo/` - 基础 Store 示例
  - [ ] `examples/reducer_demo/` - Reducer 示例
- [ ] 添加性能基准测试
- [ ] 发布 v2.0.0

---

#### Phase 3: Fiber Runtime

- [x] 完善 FiberNode 结构（双缓存、Effect）- 已有实现
- [x] 实现 UseState 无闭包版本 - 已有实现
- [x] 实现 VNode Diff 算法 - 已有实现
- [x] 实现 Reconciler - 已有实现
- [x] 添加时间旅行调试器
- [x] 更新文档
  - [x] 创建 `FIBER_ARCHITECTURE.md`
  - [x] 创建 `HOOK_USAGE_GUIDE.md`
- [x] 创建示例
  - [x] `examples/fiber_demo/` - Fiber 基础示例
  - [x] `examples/timetravel_demo/` - 时间旅行示例
- [ ] 发布 v2.1.0

---

#### Phase 4: 类型安全 DSL

- [x] 实现 `StateKey[T]` 类型
- [x] 实现 `TypedFieldChange[T]` Intent
- [ ] 更新 Reducer 支持类型安全
- [x] 更新文档
  - [x] 创建 `TYPE_SAFE_INTENT.md`
  - [ ] 更新 `MVP_MIGRATION_GUIDE.md`
- [x] 迁移示例代码
  - [x] `examples/typed_intent_demo/` - 类型安全 Intent 示例
  - [ ] `examples/mvp_components_demo/` 使用类型安全
- [ ] 发布 v2.2.0

---

#### Phase 5: Lane 调度

- [x] 实现 Lane 类型定义
- [x] 实现 Scheduler
- [x] 实现 IntentWithLane
- [ ] 添加性能测试
  - [ ] 大列表渲染
  - [ ] 快速连续输入
- [x] 更新文档
  - [x] 创建 `LANE_SCHEDULING.md`
- [x] 创建示例
  - [x] `examples/lane_demo/` - Lane 优先级示例
  - [x] `examples/interruptible_demo/` - 可中断渲染示例
- [ ] 发布 v2.3.0

---

### 7.2 迁移策略

#### 新项目

```go
// ✅ 直接使用新架构
func main() {
    // 定义状态
    type AppState struct {
        Count int
    }

    // 定义 Reducer
    reducer := func(state AppState, intent Intent) AppState {
        switch intent.(type) {
        case IncrementIntent:
            state.Count++
        }
        return state
    }

    // 创建 Runtime
    runtime := NewRuntime(AppState{Count: 0}, AppView, reducer)

    // 运行
    ui.RunApp(runtime)
}
```

---

#### 现有项目

```go
// ✅ 渐进式迁移

// 阶段 1: 仅修复闭包问题
// 修改 handler 从 Context 读取状态

// 阶段 2: 引入 Store（可选，与 UseState 共存）
// 逐步将复杂组件迁移到 Store

// 阶段 3: 完全迁移
// 移除所有 UseState，使用 Store
```

---

## 八、风险评估

### 8.1 技术风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| Hook 顺序错误 | 🟡 中 | 🔴 高 | 编译期检查 + 运行时断言 |
| Fiber Diff 性能问题 | 🟡 中 | 🟡 中 | 基准测试 + 优化热点 |
| 类型转换错误 | 🟡 中 | 🟡 中 | 类型安全 DSL |
| 内存泄漏（Hooks 未清空） | 🟢 低 | 🔴 高 | 自动检测 + GC |
| 并发竞争 | 🟢 低 | 🔴 高 | 单线程调度 + 队列 |

---

### 8.2 兼容性风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 现有示例代码破坏 | 🟡 中 | 🟡 中 | 保持旧 API 废弃但不删除 |
| 性能回退 | 🟡 中 | 🟡 中 | 基准测试对比 |
| 用户迁移成本 | 🔴 高 | 🟡 中 | 详细文档 + 迁移工具 |

---

### 8.3 业务风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 开发周期延长 | 🟡 中 | 🟡 中 | 分阶段发布 |
| 团队学习成本 | 🟡 中 | 🟡 中 | 培训 + 示例 |
| 功能回归 | 🟢 低 | 🔴 高 | 自动化测试 |

---

## 九、版本规划

### 9.1 版本路线图

```text
v1.0 (当前) ──────► v1.1 (Phase 1)
   │                      │
   └─ MVP 组件 ✅         ├─ 消除闭包依赖
                          ├─ Context 参数支持
                          └─ 50+ 示例迁移

v1.1 ────────────► v2.0 (Phase 2)
   │                      │
   └─ 稳定运行             ├─ Store + Reducer
                          ├─ 单一状态源
                          └─ 代码示例重写

v2.0 ────────────► v2.1 (Phase 3)
   │                      │
   └─ 架构升级             ├─ Fiber Runtime 完善
                          ├─ 时间旅行
                          └─ 性能优化

v2.1 ────────────► v2.2 (Phase 4)
   │                      │
   └─ 高级特性             ├─ 类型安全 DSL
                          ├─ StateKey[T]
                          └─ 编译期检查

v2.2 ────────────► v2.3 (Phase 5)
   │                      │
   └─ 生产就绪             ├─ Lane 调度
                          ├─ 可中断渲染
                          └─ 性能基准
```

---

### 9.2 时间估算

| 阶段 | 工作量 | 并行度 | 净周期 |
|------|--------|--------|--------|
| Phase 1 | 3 天 | 🟡 中 | 2 周 |
| Phase 2 | 5 天 | 🟡 中 | 2 周 |
| Phase 3 | 7 天 | 🔴 低（串行） | 3 周 |
| Phase 4 | 4 天 | 🟢 高 | 1 周 |
| Phase 5 | 6 天 | 🔴 低 | 2 周 |
| **总计** | **25 天** | - | **10 周** |

---

## 十、结论

### 10.1 核心价值

通过本次重构，Mint UI 将实现：

1. **架构正确性**: 消除闭包依赖，建立稳定的数据流
2. **可维护性**: 状态集中，逻辑集中，易于调试
3. **可扩展性**: 支持 Fiber、Lane、时间旅行等高级特性
4. **类型安全**: 编译期检查，减少运行时错误
5. **性能提升**: Lane 调度、可中断渲染、双缓存优化

---

### 10.2 技术定位

从 **"组件驱动的 TUI 库"** 升级为 **"状态机驱动的 UI Runtime"**

```text
Before: Mint UI ≈ Bootstrap for TUI
After:  Mint UI ≈ React Fiber (Go版)
```

---

### 10.3 下一步行动

**立即开始**: Phase 1 - 消除闭包依赖

1. ✅ 创建 `runtime/intent/action_context_ext.go` - 扩展 ActionContext
2. ✅ 更新 `ui/intent.go` - 增加 Context 参数支持
3. ✅ 更新 `examples/ant_design_demo/` - 迁移示例代码
4. ✅ 添加 E2E 测试 - 验证闭包问题修复

**预计收益**:
- 解决当前最严重的问题
- 为后续阶段奠定基础
- 2 周内交付 v1.1.0

---

**文档维护**: 如有问题或建议，请提交 Issue 或 PR。

**相关文档**:
- [mini_demo.md](./mini_demo.md) - Mini Fiber Runtime 参考实现
- [store.md](./store.md) - Store + Reducer 详细设计
- [MVP_MIGRATION_GUIDE.md](../mvp/MVP_MIGRATION_GUIDE.md) - MVP 基础迁移指南
- [INTENT_MANAGEMENT_PATTERNS.md](../mvp/INTENT_MANAGEMENT_PATTERNS.md) - Intent 管理模式

---

**最后更新**: 2026-03-03
**版本**: v1.0
**状态**: 🟢 草稿待审
