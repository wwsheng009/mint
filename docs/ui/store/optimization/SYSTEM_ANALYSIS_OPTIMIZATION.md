# Mint UI Store + Intent 系统分析与优化方案

**创建时间**: 2026-03-04
**版本**: v1.0
**状态**: 📋 草稿

---

## 📋 目录

- [一、当前系统存在的问题](#一当前系统存在的问题)
  - [1.1 Intent Handler 注册顺序问题](#11-intent-handler-注册顺序问题)
  - [1.2 内置 FieldChangeIntent Handler 冲突](#12-内置-fieldchangeintent-handler-冲突)
  - [1.3 Intent 类型命名冲突](#13-intent-类型命名冲突)
  - [1.4 日志与错误管理不足](#14-日志与错误管理不足)
  - [1.5 类型安全问题](#15-类型安全问题)
- [二、根因分析](#二根因分析)
- [三、优化方案](#三优化方案)
  - [3.1 方案 1: 生命周期钩子（推荐）](#31-方案-1-生命周期钩子推荐)
  - [3.2 方案 2: Handler 继承模式](#32-方案-2-handler-继承模式)
  - [3.3 方案 3: 命名空间隔离](#33-方案-3-命名空间隔离)
- [四、统一日志与错误管理](#四统一日志与错误管理)
- [五、类型安全增强](#五类型安全增强)
- [六、实施计划](#六实施计划)
- [七、总结](#七总结)

---

## 一、当前系统存在的问题

### 1.1 Intent Handler 注册顺序问题

**问题描述**

用户在使用 Store + Reducer 模式时，Intent handler 的注册顺序导致预期行为失效。

**示例问题代码**

```go
// ❌ 错误：Handler 被覆盖

func initStore() {
    appStore = store.NewStore(AppState{...})

    // 1. 用户注册 FieldChangeIntent handler
    reducerBuilder.RegisterToGlobal(appStore)  
    // -> 注册了 "FieldChange" -> 用户 reducer + appStore.Set()
}

func main() {
    ui.Run(App,
        // ❌ 没有使用 WithInit()
    )
}
```

**运行时序**

```
1. main() 执行
   ↓
2. initStore() → 注册 "FieldChange" handler
   ↓
3. ui.Run() → SetupBuiltinHandlers()
   ↓
4. SetupBuiltinHandlers() → 又注册 "FieldChange" handler
   ↓
5. 🔥 用户的 handler 被覆盖！
```

**内置 handler 行为**

```go
// runtime/intent/builtin.go

func handleFieldChange(ctx *ActionContext, i FieldChangeIntent) IntentResult {
    // ✨ MVP: State is the single source of truth
    // The value from Instance (user input) becomes the new state value
    ctx.SetState(i.Field, i.Value)  // ← 使用 SimpleStore，不是 appStore！
    ctx.ScheduleUpdate()
    return HandledResult()
}
```

**影响**

- 输入框输入正确
- 但状态没有更新到用户的 `appStore`
- UI 不会重新渲染（因为用户 reducer 没有被调用）

---

### 1.2 内置 FieldChangeIntent Handler 冲突

**问题描述**

内置 `handleFieldChange` 使用 `ActionContext.SimpleStore`，而 Store + Reducer 模式使用强类型 `Store[AppState]`。

**内置 Store 实现**

```go
// runtime/intent/helper.go

type SimpleStore struct {
    mu    sync.RWMutex
    state map[string]interface{}  // ← 字符串键 + interface{} 值
    dirty bool
}

func (s *SimpleStore) SetState(key string, value interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.state[key] = value
    s.dirty = true
}
```

**用户 Store 实现**

```go
// examples/store_reducer_demo/main.go

type AppState struct {
    Username   string
    Email      string
    UsernameErr string
    EmailErr    string
}

var appStore *store.Store[AppState]
```

**冲突分析**

| 特性 | 内置 SimpleStore | 用户 Store[AppState] |
|------|-----------------|---------------------|
| 类型 | `map[string]interface{}` | 强类型 `AppState` |
| 键 | 字符串（易拼写错误） | 结构体字段 |
| 访问方式 | `ctx.SetState("username", val)` | `state.Username = val` |
| 类型安全 | ❌ 运行时断言 | ✅ 编译期检查 |
| 验证逻辑 | ❌ 无 | ✅ 在 Reducer 中 |

---

### 1.3 Intent 类型命名冲突

**问题描述**

用户自定义 Intent 类型可能与内置 Intent 冲突。

**已知冲突案例**

```go
// ❌ 用户定义（与内置冲突）
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }

// 内置定义
// runtime/intent/builtin.go
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
```

**其他潜在冲突**

| 用户 Intent 类型 | 内置 Intent 类型 | 冲突后果 |
|----------------|----------------|---------|
| `SubmitFormIntent` | `SubmitFormIntent` | 提交表单行为异常 |
| `IncrementIntent` | `IncrementIntent` | 计数器逻辑失效 |
| `DecrementIntent` | `DecrementIntent` | 计数器逻辑失效 |
| `ChangeTabIntent` | N/A | 可能未来冲突 |

**当前临时解决**

```go
// 用户被迫使用前缀避免冲突
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "DemoIncrement" }

type SubmitFormIntent struct{}
func (SubmitFormIntent) IntentType() string { return "DemoSubmitForm" }
```

**问题**

- 临时方案：用户需要记住添加前缀
- 不优雅：意图类型名称冗余
- 易遗漏：新用户可能不知道冲突

---

### 1.4 日志与错误管理不足

**问题描述**

系统的日志和错误处理分散，缺乏统一的调试工具。

**当前日志状态**

| 位置 | 日志方式 | 问题 |
|------|---------|------|
| `dispatcher.go` | `if d.log { ... }` | 默认关闭，无结构化日志 |
| `fiber_util.go` | `fmt.Printf("❌ ...")` | 标准输出，不可控 |
| `app.go` | `log.UILogger.Debug(...)` | 仅框架层日志 |
| 用户代码 | 无标准方式 | 用户需自己实现 |

**错误处理问题**

```go
// ❌ 当前：部分代码不检查 error

// runtime/ui/fiber_util.go
func (f IntentFactory) EmitIntent(i Intent) {
    if f.intentEmitter != nil {
        f.intentEmitter(i)  // ❌ 不检查返回的 IntentResult
    }
}

// examples/store_reducer_demo/main.go
ui.On(IncrementIntent{}, func() {  // ❌ 无错误处理
    state.Count++
})
```

**影响**

- Intent 失败时静默失败，难以调试
- 日志分散，无法统一查询
- 性能分析困难（没有 Intent 处理时间统计）

---

### 1.5 类型安全问题

**问题描述**

字符串键容易拼写错误，IDE 无法提供自动补全。

**问题示例**

```go
// ❌ 字符串键
input.ForField("username")       // ✅ 正确
input.ForField("ussrename")      // ❌ 拼写错误，编译期无法检测！

// ❌ Handler 中字符串匹配
if fieldChange.Field == "username" {
    // ...
}
```

**影响**

- 拼写错误 → 运行时才发现
- 无法重构（IDE 无法追踪字符串）
- 无类型推导，文档分散

**已有改进但未充分利用**

```go
// ✅ StateKey[T] 提供类型安全
var Username = intent.NewStateKey[string]("username")

Username.Set(ctx, "alice")
username := Username.Get(ctx, "")
input.ForField(Username.String())  // "username"
```

**问题**

- `StateKey[T]` 已实现但示例代码未广泛使用
- `TypedFieldChange[T]` 存在但未被充分利用
- 迁移指南缺乏

---

## 二、根因分析

### 2.1 架构设计：生命周期不匹配

| 层 | 生命周期 | 注册时机 |
|---|---------|---------|
| **用户 Handler** | 应用级 | `init()` 或 `main()` 开始 |
| **内置 Handler** | 框架级 | `ui.Run()` 内部 |
| **组件渲染** | 请求级 | 每次 `ui.Run()` 调用 |

```
用户 handler 注册 → [用户代码]
  ↓
内置 handler 注册 → [框架代码，覆盖用户！]
  ↓
应用启动
```

**根本原因**

框架假设"内置 handler 优先"，但 Store + Reducer 模式需要"用户 handler 优先"。

---

### 2.2 职责模糊：StateStore vs SimpleStore

```text
ActionContext
  ├─ GlobalState (map[string]interface{})
  └─ StateSetter (SetState → SimpleStore)

Store[T]
  └─ 强类型状态

问题：
  1. 内置 handler 使用 SimpleStore
  2. 用户代码使用 Store[T]
  3. 两者无法互通
```

---

### 2.3 缺乏命名空间

所有 Intent 类型使用扁平命名：
- "FieldChange"
- "Increment"
- "SubmitForm"

没有前缀或命名空间隔离：
- 框架 Intent: `framework.Increment`
- 用户 Intent: `app.Increment`

---

## 三、优化方案

### 3.1 方案 1: 生命周期钩子（推荐）⭐

**设计目标**

- 用户通过 `WithInit()` 在内置 handler 之后注册
- 内置 handler 检测是否有用户 handler，若有则跳过

**实现步骤**

#### 步骤 1: 内置 handler 支持检测用户注册

```go
// runtime/intent/builtin.go

func SetupBuiltinHandlers(rt *Runtime) {
    registry := rt.Dispatcher.GetRegistry()
    
    // ✨ 标记为"可覆盖"
    registry.Register(IncrementIntent{}.IntentType(), 
        intent.HandlerFunc(handleIncrement),
        intent.WithOverridable(true))  // ← 新选项
}
```

#### 步骤 2: Registry 支持覆盖标记

```go
// runtime/intent/registry.go

type HandlerRegistration struct {
    Handler      Handler
    Overridable  bool  // 新增：是否允许被覆盖
    Priority     ActionPriority
}

func (r *Registry) Register(intentType string, handler Handler, opts ...RegisterOption) {
    reg := HandlerRegistration{
        Handler: handler,
        // ...
    }
    
    // 应用选项
    for _, opt := range opts {
        opt(&reg)
    }
    
    r.handlers[intentType] = reg
}

// ✨ 注册选项
type RegisterOption func(*HandlerRegistration)

func WithOverridable(overridable bool) RegisterOption {
    return func(reg *HandlerRegistration) {
        reg.Overridable = overridable
    }
}
```

#### 步骤 3: `WithInit()` 确保覆盖执行

```go
func (r *Registry) Register(intentType string, handler Handler, opts ...RegisterOption) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    existing, ok := r.handlers[intentType]
    
    // ✨ 如果存在且可覆盖，替换
    if ok && existing.Overridable {
        log.Debug("Overriding overridable handler for %s", intentType)
        r.handlers[intentType] = handler
        return
    }
    
    // 禁止覆盖保护
    if ok && !existing.Overridable {
        log.Warn("Cannot override protected handler for %s", intentType)
        return
    }
    
    // 正常注册
    r.handlers[intentType] = handler
}
```

#### 使用示例

```go
func main() {
    initStore()
    
    ui.Run(App,
        ui.WithInit(func() {
            // ✅ 内置 handlers 之后执行，安全覆盖
            registerHandlers()
        }),
    )
}

func registerHandlers() {
    reducerBuilder.RegisterToGlobal(appStore)
    
    // ✨ 声明式覆盖：不再需要手动 EmitIntent
    // 内置 handleFieldChange 会被检测并跳过
}
```

---

### 3.2 方案 2: Handler 继承模式

**设计目标**

用户可以继承内置 handler，选择性覆盖逻辑。

```go
// runtime/intent/registry.go

// HandlerChain 允许链式调用多个 handler
type HandlerChain struct {
    handlers []Handler
}

func (c *HandlerChain) Handle(ctx *ActionContext, i Intent) IntentResult {
    for _, h := range c.handlers {
        result := h.Handle(ctx, i)
        if !result.Handled {
            continue  // chain to next
        }
        if result.StopPropagation {
            return result  // stop chain
        }
    }
    return HandledResult()
}

// ✨ 扩展 IntentResult
type IntentResult struct {
    Handled           bool
    StopPropagation   bool  // 新增：停止传播
    Error             error
}
```

**使用示例**

```go
func registerHandlers() {
    // ✨ 链式：先执行内置，再执行用户逻辑
    existingHandler, _ := intent.DefaultRegistry().GetHandler("FieldChange")
    
    intent.DefaultRegistry().Register("FieldChange", 
        &intent.HandlerChain{
            handlers: []intent.Handler{
                existingHandler,         // 内置：更新 SimpleStore
                intent.HandlerFunc(func(ctx *intent.ActionContext, i Intent) IntentResult {
                    // 用户：更新 appStore
                    if fc, ok := i.(intent.FieldChangeIntent); ok {
                        newState := reducerBuilder.Build().Reduce(appStore.Get(), fc)
                        appStore.Set(newState)
                    }
                    return intent.HandledResult()
                }),
            },
        })
}
```

**优势**

- 内置逻辑保留，用户逻辑追加
- 灵活：可以完全覆盖或部分扩展

**劣势**

- 复杂度增加
- 双重更新（SimpleStore + appStore）

---

### 3.3 方案 3: 命名空间隔离

**设计目标**

为 Intent 类型添加命名空间，避免冲突。

```go
// runtime/intent/types.go

type Intent interface {
    // IntentType 返回命名空间:类型
    // 示例: "framework:Increment", "app:Increment"
    IntentType() string
    
    // Namespace 返回命名空间（可选）
    Namespace() string
}

// ✨ 便捷函数
func NamespacedType(namespace, typ string) string {
    return fmt.Sprintf("%s:%s", namespace, typ)
}
```

**内置 Intent 使用框架命名空间**

```go
// runtime/intent/builtin.go

const FrameworkNamespace = "framework"

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string {
    return NamespacedType(FrameworkNamespace, "Increment")
}

func (IncrementIntent) Namespace() string {
    return FrameworkNamespace
}
```

**用户 Intent 使用应用命名空间**

```go
// examples/store_reducer_demo/main.go

const AppNamespace = "demo"

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string {
    return NamespacedType(AppNamespace, "Increment")  // "demo:Increment"
}

// ✅ 无冲突！
// 内置: framework:Increment
// 用户: demo:Increment
```

**Registry 支持命名空间过滤**

```go
// runtime/intent/registry.go

// GetHandlerWithNamespace 在指定命名空间查找 handler
func (r *Registry) GetHandlerWithNamespace(intentType string, ns string) (Handler, bool) {
    // 1. 先查找确切的 namespace:type
    if h, ok := r.handlers[intentType]; ok {
        return h, true
    }
    
    // 2. 再查找全局的 type（无 namespace）
    if withoutNs := strings.TrimPrefix(intentType, ns+":"); withoutNs != intentType {
        if h, ok := r.handlers[withoutNs]; ok {
            log.Warn("Falling back to global handler for %s", intentType)
            return h, true
        }
    }
    
    return nil, false
}
```

**优势和权衡**

| 特性 | 命名空间隔离 | 前缀方案 |
|------|-------------|---------|
| 冲突避免 | ✅ 完全隔离 | ⚠️ 需要手动记忆 |
| 类型安全 | ✅ 编译期检查 | ⚠️ 字符串 |
| 向后兼容 | ❌ 需要迁移 Intent | ✅ 完全兼容 |
| 可读性 | ✅ 清晰表达来源 | ⚠️ 前缀冗余 |

---

## 四、统一日志与错误管理

### 4.1 结构化日志系统

**设计目标**

- 统一的日志接口
- 结构化日志（JSON 或键值对）
- 可配置的日志级别
- Intent 处理全程追踪

```go
// internal/log/logger.go

type LogLevel int

const (
    LevelDebug LogLevel = iota
    LevelInfo
    LevelWarn
    LevelError
)

type LogEntry struct {
    Level     LogLevel
    Timestamp time.Time
    Component string  // "dispatcher", "reducer", "store"
    Message   string
    Fields    map[string]interface{}
    
    // Intent 专用字段
    IntentType string
    IntentData Intent
    Duration   time.Duration
    Error      error
}

type Logger interface {
    Debug(msg string, fields map[string]interface{})
    Info(msg string, fields map[string]interface{})
    Warn(msg string, fields map[string]interface{})
    Error(msg string, err error, fields map[string]interface{})
    
    // Intent 专用
    LogIntentDispatch(intentType string, intent Intent, priority ActionPriority)
    LogIntentResult(intentType string, result IntentResult, duration time.Duration)
}
```

### 4.2 Intent 结果处理

**设计目标**

- 强制检查 Intent 结果
- 统一错误处理策略
- 可配置的错误回调

```go
// runtime/intent/dispatcher.go

type ErrorHandlingStrategy int

const (
    LogAndIgnore  ErrorHandlingStrategy = iota  // 仅记录，忽略
    LogAndPanic                               // 记录并 panic
    LogAndRetry                               // 记录并重试（最多 N 次）
    CustomCallback                            // 调用用户自定义回调
)

type DispatcherConfig struct {
    ErrorStrategy ErrorHandlingStrategy
    ErrorHandler  func(intent Intent, err error)
    MaxRetry      int
}

func (d *Dispatcher) DispatchWithResult(intent Intent) (IntentResult, error) {
    start := time.Now()
    
    result := d.dispatchInternal(intent)
    
    // ✨ 强制错误检查
    if result.Error != nil {
        switch d.config.ErrorStrategy {
        case LogAndIgnore:
            d.logger.Error("Intent failed", result.Error, map[string]interface{}{
                "intentType": intent.IntentType(),
            })
        case LogAndPanic:
            panic(fmt.Sprintf("Intent %s failed: %v", intent.IntentType(), result.Error))
        case LogAndRetry:
            return d.retryDispatch(intent, result)
        case CustomCallback:
            if d.config.ErrorHandler != nil {
                d.config.ErrorHandler(intent, result.Error)
            }
        }
    }
    
    d.logger.LogIntentResult(intent.IntentType(), result, time.Since(start))
    
    return result, result.Error
}
```

### 4.3 使用示例

```go
// 配置日志和错误处理

logger := logging.NewConsoleLogger(logging.LevelDebug)

dispatcher := intent.NewDispatcher(registry,
    intent.WithLogger(logger),
    intent.WithErrorStrategy(intent.LogAndIgnore),
)

// 发射 Intent 并处理结果
result, err := dispatcher.DispatchWithResult(IncrementIntent{})
if err != nil {
    // 用户自定义错误处理
    showErrorModal(fmt.Sprintf("操作失败: %v", err))
}

// 自定义错误回调
dispatcher.SetErrorHandler(func(intent Intent, err error) {
    // 发送到错误追踪服务
    sentry.CaptureException(err, map[string]interface{}{
        "intentType": intent.IntentType(),
    })
})
```

---

## 五、类型安全增强

### 5.1 强类型 Intent DSL

**目标**

- 编译期类型检查
- 消除字符串键
- IDE 自动补全和重构支持

**定义类型安全的 Actions**

```go
// app/actions.go

package app

import "github.com/wwsheng009/mint/runtime/intent"

// ✨ Intent 构建器函数（类型安全）
type Actions struct{}

// IncrementCounter 创建 Increment Intent
func (Actions) IncrementCounter() intent.Intent {
    return IncrementIntent{
        Namespace: AppNamespace,
    }
}

// SetUsername 更新用户名
func (Actions) SetUsername(value string) intent.Intent {
    return intent.TypedFieldChange[string]{
        Key:   usernameKey,
        Value: value,
    }
}

// SubmitForm 提交表单
func (Actions) SubmitForm() intent.Intent {
    return SubmitFormIntent{
        Namespace: AppNamespace,
    }
}

// ✨ 导出单例
var AppActions = Actions{}
```

**使用示例**

```go
// 组件中使用类型安全的 Actions

import "github.com/wwsheng009/mint/app"

func Counter(state AppState) ui.VNode {
    return ui.HStack(
        ui.NewButtonBuilder("-").
            OnPress(app.AppActions.DecrementCounter()).  // ✅ 类型安全！
            Build(),
        ui.Text(strconv.Itoa(state.Count)),
        ui.NewButtonBuilder("+").
            OnPress(app.AppActions.IncrementCounter()).
            Build(),
    )
}
```

**Reduicer 使用类型断言**

```go
func appReducer(state AppState, i intent.Intent) AppState {
    switch v := i.(type) {
    case app.IncrementIntent:
        state.Count++
        
    case app.TypedFieldChange[string]:
        if v.Key == usernameKey {
            state.Username = v.Value
        }
    }
    return state
}
```

---

## 六、实施计划

### Phase 1: 紧急修复（1-2天）

| 任务 | 优先级 | 预估时间 |
|------|-------|---------|
| 修复 FieldChangeIntent handler 覆盖问题 | 🔴 P0 | 2h |
| 增强 `WithInit()` 文档和示例 | 🔴 P0 | 1h |
| 添加 Intent 结果检查到关键路径 | 🟡 P1 | 3h |
| 更新 `store_reducer_demo` 使用最佳实践 | 🟡 P1 | 1h |

**交付**
- `examples/store_reducer_demo` 正常工作
- 迁移到 `WithInit()` 模式

---

### Phase 2: 日志和错误管理（3-5天）

| 任务 | 优先级 | 预估时间 |
|------|-------|---------|
| 设计并实现统一 Logger 接口 | 🟡 P1 | 4h |
| 集成到 Dispatcher | 🟡 P1 | 3h |
| 设计错误处理策略 | 🟡 P1 | 2h |
| 更新文档和示例 | 🟢 P2 | 3h |
| 编写单元测试 | 🟢 P2 | 4h |

**交付**
- `runtime/logging` 包
- Intent 错误追踪能力
- 调试工具（`mint debug intents`）

---

### Phase 3: 类型安全增强（5-7天）

| 任务 | 优先级 | 预估时间 |
|------|-------|---------|
| 实现 Actions DSL 生成器（代码生成） | 🟢 P2 | 6h |
| 更新示例到类型安全模式 | 🟢 P2 | 4h |
| 编写迁移指南 | 🟢 P2 | 3h |
| IDE 插件（可选） | 🟢 P3 | 8h |

**交付**
- `mint/actions` 工具（代码生成）
- 类型安全的 Intent DSL
- 迁移指南

---

### Phase 4: 高级特性（可选，7-14天）

| 任务 | 优先级 | 预估时间 |
|------|-------|---------|
| 命名空间隔离 | 🟢 P3 | 6h |
| Handler 链式扩展 | 🟢 P3 | 4h |
| 性能监控和追踪 | 🟢 P3 | 8h |
| 时间旅行调试完善 | 🟢 P3 | 10h |

---

## 七、总结

### 7.1 核心

当前 Mint UI 的 Store + Intent 系统面临几个关键问题：

1. **Handler 注册顺序冲突** - 用户 handler 被内置 handler 覆盖
2. **类型冲突** - SimpleStore vs Store[T]
3. **命名冲突** - Intent 类型名称冲突
4. **日志缺失** - 无统一的调试和错误追踪

### 7.2 推荐方案

**短期（紧急）**：
1. ✅ 使用 `WithInit()` 确保用户 handler 在内置之后注册
2. ✅ 显式覆盖 `FieldChangeIntent` handler
3. ✅ 添加 Intent 结果检查

**中期（优化）**：
1. 统一日志系统（`runtime/logging`）
2. 错误处理策略（`LogAndIgnore`, `LogAndPanic` 等）
3. 类型安全 Intent DSL（`StateKey[T]`, `TypedFieldChange[T]`）

**长期（演进）**：
1. 命名空间隔离
2. Handler 链式继承
3. 代码生成的 Actions

### 7.3 设计原则

| 原则 | 适用场景 | 实现位置 |
|------|---------|---------|
| 优先级：用户 > 内置 | Store + Reducer 模式 | `WithInit()` + overridable 标记 |
| 类型安全优先 | 新项目 | `StateKey[T]` + `TypedFieldChange[T]` |
| 向后兼容 | 所有阶段 | 保留字符串键，渐进迁移 |
| 可观测性 | 生产环境 | 统一日志 + 错误处理 |

### 7.4 后续行动

- [x] **修复 `store_reducer_demo`** - 已完成
- [ ] 设计 `WithOverridable()` API
- [ ] 实现统一 Logger
- [ ] 编写 `actions` 代码生成工具
- [ ] 更新迁移指南文档

---

## 附录

### A. 当前可行的变通方案

```go
// ✅ 立即可用的修复方案

func main() {
    initStore()
    
    ui.Run(App,
        ui.WithInit(registerHandlers),  // 关键：在内置 handler 之后注册
    )
}

func registerHandlers() {
    // 1. 先注册用户 handlers（可能被覆盖）
    reducerBuilder.RegisterToGlobal(appStore)
    
    // 2. 显式重新注册 FieldChangeIntent handler（覆盖内置）
    intent.DefaultRegistry().Register(
        intent.FieldChangeIntent{}.IntentType(),
        intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
            if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
                newState := reducerBuilder.Build().Reduce(appStore.Get(), fieldChange)
                appStore.Set(newState)
                ctx.ScheduleUpdate()
            }
            return intent.HandledResult()
        }))
}
```

### B. 相关文档

- [STORE_REDUCER_GUIDE.md](../guides/STORE_REDUCER_GUIDE.md) - Store + Reducer 完整指南
- [TYPE_SAFE_INTENT.md](../features/TYPE_SAFE_INTENT.md) - 类型安全 Intent DSL
- [INTENT_HANDLER_MIGRATION.md](../migration/INTENT_HANDLER_MIGRATION.md) - Intent Handler 迁移指南
- [REFACTOR_PLAN.md](/docsArchive/REFACTOR_PLAN.md) - 完整架构重构计划

---

**文档版本**: v1.0  
**最后更新**: 2026-03-04  
**维护者**: Mint UI 团队
