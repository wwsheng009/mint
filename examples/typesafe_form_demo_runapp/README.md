# Type-Safe Form Demo (RunApp + FieldMap)

这是一个展示 Mint UI 最新特性的完整表单示例，使用了 Store + Reducer 架构的所有最佳实践。

**版本**: 2.0
**最后更新**: 2025-03-05 (已实现全部6个高级特性)
**运行命令**: `go run ./examples/typesafe_form_demo_runapp/`

---

## 📋 核心特性清单

本示例展示了 Mint UI 以下核心和高级特性：

### ✅ 已实现的核心特性

| 特性 | 描述 | 代码行号 |
|------|------|---------|
| **ui.RunApp[T]** | 最新应用启动方式，自动状态管理 | main.go:635 |
| **AppRuntime[T]** | Store + Reducer + View 统一运行时 | main.go:635 |
| **Store[T]** | 泛型状态容器，单一真相源 | 已实现 |
| **Reducer[T]** | 纯函数状态转换器，Builder 模式 | 已实现 |
| **FieldMap** | 消除 switch-case 的字段处理器优化 | main.go:66-93 |
| **BindFieldMap** | Map 方式的字段映射，无需类型断言 | main.go:67-93 |
| **ForField 绑定** | 自动字段双向绑定，更新自动同步 | main.go:132-159 |
| **Intent 处理** | 类型安全的用户意图和处理器 | main.go:96-113 |
| **自动重新渲染** | 状态变化自动触发 UI 更新 | 已自动 |
| **UI 组件库** | Input, Checkbox, Button 等组件 | main.go:130-171 |
| **样式系统** | FgColor, Bold 等文本样式 | main.go:175-230 |
| **类型安全 View** | 包装函数模式避免循环依赖 | main.go:317-322 |

### ✅ 已实现的高级特性

本示例已实现以下6个高级特性，展示 Mint UI 的完整能力：

| 特性 | 描述 | 代码行号 | 实现方式 |
|------|------|---------|----------|
| **Lane Scheduler** | 优先级渲染调度，通过Lane优化渲染性能 | main.go:712, 721 | `ui.WithLaneScheduler(ui.LaneBackground)` |
| **WithDefaultLane** | 设置默认渲染优先级 | main.go:716 | `ui.WithDefaultLane(ui.LaneNormal)` |
| **PluginSetup** | 插件系统（如Modal插件处理） | main.go:695 | `ui.WithPluginSetup()` |
| **Time Travel Debugging** | 状态历史记录、撤销和重做 | main.go:84-88, 720 | `rt.Undo()`, `rt.Redo()`, `rt.JumpTo()` |
| **Computed/Selector** | 计算值优化，避免不必要的重渲染 | main.go:110-122 | `store.NewComputed()` |
| **Middleware** | Reducer中间件（日志、调试等） | main.go:310-311 | `reducer.WithMiddleware()` |

**重要更新**: Time Travel Debugging 现在包含完整的 Undo/Redo 功能，使用 `skipHistory` 机制避免无限循环；历史修复记录见 `docsArchive/cleanup-2026-05-19/docs/ui/store/fixes/TIMETRAVEL_FIX.md`。

---

## 🚀 快速开始

```bash
# 运行示例
go run ./examples/typesafe_form_demo_runapp/

# 或编译后运行
go build -o form_demo ./examples/typesafe_form_demo_runapp/
./form_demo
```

---

## 📚 特性详解

### 1️⃣ ui.RunApp[T] - 现代化应用启动

#### 功能描述

`ui.RunApp[T]` 是 Mint UI 推荐的应用启动方式，自动管理 Store、Reducer 和 View 的生命周期。

#### 优势

| 对比项 | 旧方式 (ui.Run) | 新方式 (ui.RunApp) |
|--------|---------------|----------------|
| **全局 Store** | 需要 ❌ | 不需要 ✅ |
| **手动读取状态** | 需要 ❌ | 自动 ✅ |
| **手动触发渲染** | 需要 ❌ | 自动 ✅ |
| **代码行数** | ~184 行 | ~286 行（包含所有状态显示） |
| **时间旅行调试** | 手动实现 | 内置 ✅ |

#### 使用方法

```go
func main() {
    rt := statemachine.NewAppRuntime(
        AppState{},
        AppView,
        AppReducer,
        statemachine.WithMaxHistory(100),  // 可选：历史记录
    )
    
    ui.RunApp(rt,
        ui.WithWidth(60),
        ui.WithHeight(30),
        ui.WithTitle("My Form"),
        ui.WithInit(func() {
            appReducerBuilder.RegisterToGlobal(rt.GetStore())
        }),
    )
}
```

#### 注意事项

- ⚠️ **必须注册 Intent handlers**：使用 `ui.WithInit` 和 `RegisterToGlobal()`
- ⚠️ **保持 Builder 引用**：使用 `fieldMapBuilder` 而不是构建后的 `AppReducer` 进行扩展

---

### 2️⃣ FieldMap - 消除 switch-case 硬编码 ⚡

#### 功能描述

`reducer.BindField()` + `BindFieldMap()` 提供了一种声明式的字段处理方式，自动处理所有表单字段更新，无需编写switch-case语句。

#### 优势

| 指标 | 传统 switch-case | FieldMap |
|------|----------------|----------|
| 代码行数 | ~40 行 | ~20 行 **(-50%)** |
| 类型断言 | 需要（运行时） | ❌ 不需要 |
| 编译时检查 | ❌ 无 | ✅ 有 |
| 添加新字段 | 修改 3+ 处 | **只添加 1 行** |
| 测试难度 | 困难 | 容易 |

#### 代码对比

**传统方式**（复杂）：
```go
reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            switch fieldChange.Field {
            case "Username":
                s.Username = fieldChange.Value
            case "Email":
                s.Email = fieldChange.Value
            case "Age":
                if v, err := strconv.Atoi(fieldChange.Value); err == nil {
                    s.Age = v
                } else {
                    return s  // 错误处理
                }
            case "AcceptTerms":
                s.AcceptTerms = fieldChange.Value == "true"
            case "Subscribe":
                s.Subscribe = fieldChange.Value == "true"
            }
        }
        return s
    })
```

**FieldMap 方式**（简洁）：
```go
var fieldMapBuilder = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "Username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "Email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
        "Age": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Age = v
            }
            return s
        },
        "AcceptTerms": func(s AppState, val string) AppState {
            s.AcceptTerms = val == "true"
            return s
        },
        "Subscribe": func(s AppState, val string) AppState {
            s.Subscribe = val == "true"
            return s
        },
    })
```

#### 添加新字段

```go
// 步骤 1: 在 AppState 中添加字段
type AppState struct {
    Username string
    Email    string
    Phone    string  // ← 新增
}

// 步骤 2: 在 FieldMap 中添加一行
fieldMapBuilder.BindFieldMap(map[string]func(AppState, string) AppState{
    "Username": func(s AppState, val string) AppState {
        s.Username = val
        return s
    },
    "Email": func(s AppState, val string) AppState {
        s.Email = val
        return s
    },
    "Phone": func(s AppState, val string) AppState {  // ← 新增：仅此一行！
        s.Phone = val
        return s
    },
})

// 步骤 3: 在 View 中添加 UI 组件
phoneInput := input.NewBuilder().
    ForField(intent.BindField("Phone")).
    Value(state.Phone).
    Width(20).
    Build()
```

---

### 3️⃣ ForField 绑定 - 自动字段更新

#### 功能描述

`intent.BindField("username")` 创建一个字段绑定对象，UI 组件通过 `ForField()` 设置后，会自动与状态字段双向同步。

#### 工作原理

```
用户输入 "John"
    ↓
Input 组件发出 Intent
    ↓
FieldChangeIntent{Field: "Username", Value: "John"}
    ↓
FieldMap 自动处理
    ↓
AppState.Username = "John"
    ↓
Store.Set(newState)
    ↓
AppRuntime.OnStateChange()
    ↓
UI 自动重新渲染
    ↓
Input 组件显示 "John"  （自动更新）
```

#### 使用方法

```go
// 1. 定义状态字段
type AppState struct {
    Username string
}

// 2. 创建带绑定的输入框
usernameInput := input.NewBuilder().
    ForField(intent.BindField("Username")).  // ← 字段绑定
    Value(state.Username).
    Placeholder("Enter username").
    Width(30).
    Build()
```

#### 可用类型

| 显示类型 | Go 类型 | 自动转换 |
|---------|---------|---------|
| Text | string | 无需转换 |
| Number | int, int64, float64 | `reducer.FormatInt()` / `reducer.FormatFloat()` |
| Boolean | bool | `"true"` / `"false"` 自动转换 |

---

### 4️⃣ 类型安全的 View 包装

#### 功能描述

使用包装函数避免 `runtime` 和 `ui` 包之间的循环依赖，同时提供完整的编译时类型安全。

#### 原因

```
ui 包
  └─→ ui.RunApp[T](statemachine.AppRuntime[T])
      └→ ViewFunction[T] func(T) any  ← 使用 any 避免循环依赖
          
statemachine 包
  └──→ ui.VNode (不能直接引用)
```

#### 使用方法

```go
// AppView 返回 any（避免循环依赖）
func AppView(state AppState) any {
    return renderAppView(state)
}

// renderAppView 返回 ui.VNode（完全类型安全）
func renderAppView(state AppState) ui.VNode {
    return ui.VStack(
        ui.Text("Hello"),
        ui.NewButtonBuilder("Click").OnPress(MyIntent{}).Build(),
    )
}
```

#### 优势

- ✅ IDE 智能提示和类型检查
- ✅ 编译器会检查 `renderAppView` 的返回类型
- ✅ 零运行时开销（编译器内联）

---

### 5️⃣ UI 组件系统

#### 可用组件

| 组件 | Builder 模式 | 主要属性/方法 |
|------|-------------|----------------|
| **Input** | `input.NewBuilder()` | `ForField()`, `Value()`, `Width()`, `Placeholder()` |
| **Checkbox** | `checkbox.NewBuilder()` | `ForField()`, `Checked()`, `Label()` |
| **Button** | `button.NewBuilder()` | `OnPress()`, `Variant()` |
| **Text** | `NewTextBuilder()` | `FgColor()`, `Bold()`, `Underline()` |
| **VStack** | `ui.VStack()` | 垂直堆叠容器 |
| **HStack** | `ui.HStack()` | 水平堆叠容器 |

#### 使用示例

```go
// 文本输入框
input.NewBuilder().
    ForField(intent.BindField("Username")).
    Value(state.Username).
    Placeholder("Enter username").
    Width(30).
    Build()

// 复选框
checkbox.NewBuilder().
    ForField(intent.BindField("AcceptTerms")).
    Checked(state.AcceptTerms).
    Label("I accept terms").
    Build()

// 按钮
button.NewBuilder("Submit").
    OnPress(SubmitIntent{}).
    Variant(button.VariantPrimary).
    Build()

// 文本样式
ui.NewTextBuilder("Success Message").
    FgColor("green").
    Bold(true).
    Underline(true).
    Build()
```

---

## ⚠️ 重要注意事项

### 1. 避免在 Reducer 中使用 fmt.Println

```go
// ❌ 错误：输出会被 UI 渲染覆盖或污染
On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
    fmt.Println("Form submitted")  // ← 输出会异常
    return s
})

// ✅ 正确：设置状态，在 UI 中显示
type AppState struct {
    Submitted bool
}

On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
    s.Submitted = true
    return s
})

// 在 View 中显示
if state.Submitted {
    ui.NewTextBuilder("✅ Form Submitted").
        FgColor("green").
        Build()
}
```

**原因**：`os.Stdout` 被 UI 渲染系统占用（每秒60次写入），`fmt.Println` 输出会被控制字符污染。历史问题记录见 `docsArchive/cleanup-2026-05-19/docs/ui/store/issues/FMT_PRINT_ISSUE.md`。

### 2. 必须注册 Intent handlers

```go
ui.RunApp(rt,
    ui.WithInit(func() {
        // ⚠️ 必须注册，否则按钮点击会报错
        appReducerBuilder.RegisterToGlobal(rt.GetStore())
    }),
)
```

### 3. 保持 Builder 引用

```go
// ✅ 正确：使用 Builder 引用进行注册
var appReducerBuilder = fieldMapBuilder.GetBuilder().
    On(SubmitIntent{}, handleSubmit).
    On(ResetIntent{}, handleReset)

ui.RunApp(rt, ui.WithInit(func() {
    appReducerBuilder.RegisterToGlobal(rt.GetStore())  // ✅ Builder 引用
}))

// ❌ 错误：使用已构建的 Reducer
var AppReducer = appReducerBuilder.Build()  // 字段处理器已丢失！

ui.RunApp(rt, WithInit(func() {
    AppReducer.RegisterToGlobal(rt.GetStore())  // ❌ 无法注册字段处理器！
}))
```

### 4. View 函数必须返回 any

```go
// ✅ 正确
func AppView(state AppState) any {
    return renderAppView(state)
}

func renderAppView(state AppState) ui.VNode {
    // UI 逻辑
    return ui.VStack(...)
}

// ❌ 错误（会导致循环依赖编译错误）
func AppView(state AppState) ui.VNode {
    return ui.VStack(...)
}
```

---

### 🎯 高级特性详解

本示例实现了以下6个高级特性，展示 Mint UI 的完整能力：

#### 1️⃣ Lane Scheduler & WithDefaultLane - 优先级渲染调度

##### 功能描述

Lane Scheduler 提供基于优先级的渲染调度机制，可以控制UI更新的执行顺序，优化性能和响应速度。

##### 使用方法

```go
ui.RunApp(rt,
    ui.WithLaneScheduler(ui.LaneBackground),  // 配置 Lane Scheduler
    ui.WithDefaultLane(ui.LaneNormal),         // 设置默认优先级
    // ... 其他选项
)
```

##### Lane 优先级

| Lane类型 | 说明 | 使用场景 |
|---------|------|---------|
| `LaneImmediate` | 立即执行 | 用户交互、关键反馈 |
| `LaneNormal` | 正常优先级 | 默认UI更新 |
| `LaneLow` | 低优先级 | 后台任务、日志 |
| `LaneBackground` | 后台执行 | 非关键更新 |

##### 示例中的实现

- **main.go:712** - `ui.WithLaneScheduler(ui.LaneBackground)` 配置后台任务
- **main.go:716** - `ui.WithDefaultLane(ui.LaneNormal)` 设置默认优先级
- **main.go:721** - 后台任务定期更新时间戳

##### 效果

- UI更新根据优先级排队，避免阻塞用户交互
- 后台任务在低优先级Lane中执行，不影响主流程
- 提供可预测的渲染行为

---

#### 2️⃣ PluginSetup - 插件系统

##### 功能描述

PluginSetup 允许注册和管理UI插件，例如Modal、Toast等系统级组件，提供统一的插件生命周期管理。

##### 使用方法

```go
ui.RunApp(rt,
    ui.WithPluginSetup(),  // 启用插件系统
    // ... 其他选项
)
```

##### 示例中的实现

- **main.go:695** - 使用 `ui.WithPluginSetup()` 启用插件系统
- **main.go:657-672** - 覆盖内置的 `CloseModalIntent` handler，使用Store状态管理
- **main.go:681-686** - ShowModalIntent 添加到state

##### 关键点

1. **覆盖内置Handler**: 使用 `RegisterTypedWithOpts(..., WithOverridable(true))`
2. **状态管理**: Modal状态保存在Store中，而非全局
3. **条件渲染**: 根据state.ShowModal显示Modal组件

##### 效果

- Modal组件与Store+Reducer架构无缝集成
- ESC键和点击外部区域可以关闭Modal
- 插件系统提供统一的UI扩展机制

---

#### 3️⃣ Time Travel Debugging - 时间旅行调试

##### 功能描述

Time Travel Debugging 提供完整的状态历史记录和撤销/重做功能，可以跳转到任意历史状态，是调试和错误恢复的重要工具。

##### 核心功能

- **Undo()**: 撤销到上一个状态
- **Redo()**: 重做到下一个状态
- **JumpTo()**: 跳转到指定历史状态
- **CanUndo() / CanRedo()**: 检查是否可以撤销/重做
- **History() / HistoryIndex()**: 获取历史记录和当前索引

##### 重要修复

Time Travel功能的实现使用了 `skipHistory` 机制来避免无限循环问题（历史记录见 `docsArchive/cleanup-2026-05-19/docs/ui/store/fixes/TIMETRAVEL_FIX.md`）

##### 使用方法

```go
// 创建 Runtime 时启用历史记录
rt := statemachine.NewAppRuntime(
    AppState{},
    AppView,
    AppReducer,
    statemachine.WithMaxHistory(100),  // 保存100个历史状态
)

// 在 WithInit 中注册 Undo/Redo handlers
ui.WithInit(func() {
    // Undo Handler
    intent.RegisterTypedWithOpts(
        intent.DefaultRegistry(),
        func(ctx *intent.ActionContext, i UndoIntent) intent.IntentResult {
            if !rt.CanUndo() {
                return intent.HandledResult()
            }
            rt.Undo()
            return intent.HandledResult()
        },
        intent.WithOverridable(true),
    )

    // Redo Handler
    intent.RegisterTypedWithOpts(
        intent.DefaultRegistry(),
        func(ctx *intent.ActionContext, i RedoIntent) intent.IntentResult {
            if !rt.CanRedo() {
                return intent.HandledResult()
            }
            rt.Redo()
            return intent.HandledResult()
        },
        intent.WithOverridable(true),
    )
})

// 在 View 中使用
button.NewBuilder("Undo").OnPress(UndoIntent{}).Build()
button.NewBuilder("Redo").OnPress(RedoIntent{}).Build()

// 显示历史信息
ui.NewTextBuilder(fmt.Sprintf("History: %d/100", rt.HistoryIndex()+1))
    .FgColor("cyan")
    .Build()
```

##### 示例中的实现

- **main.go:84-88** - 定义 UndoIntent 和 RedoIntent
- **main.go:676-694** - UndoIntent handler 实现
- **main.go:696-712** - RedoIntent handler 实现
- **main.go:720** - statemachine.WithMaxHistory(100) 启用历史记录

##### 历史截断机制

当在历史中间执行新操作时，会自动截断后续历史：

```
History: [0, 1, 2, 3, 4]  currentIndex = 2, State = 2
         ↓ Undo × 2
History: [0, 1, 2, 3, 4]  currentIndex = 0, State = 0
         ↓ New action (Value = 10)
History: [0, 10]           currentIndex = 1, State = 10
         ↑ 旧历史 (1, 2, 3, 4) 被截断
```

##### 效果

- 完整的撤销/重做功能，不会再出现无限冻结
- 用户可以自由探索历史状态
- 新操作会截断旧历史，避免状态混乱

---

#### 4️⃣ Computed/Selector - 计算值优化

##### 功能描述

Computed/Selector 提供计算值机制，避免在每次状态变化时都重新计算，减少不必要的重渲染，提升性能。

##### 基本概念

- **Computed**: 从状态派生的计算值，只在相关状态变化时更新
- **Selector**: 从 Computed 或状态中提取特定值

##### 使用方法

```go
// 1. 创建 Computed
emailValidComputed := store.NewComputed(
    rt.GetStore(),
    func(s AppState) bool {
        return containsAt(s.Email)
    },
    func(s AppState) bool {
        return s.Email != "" && len(s.Email) >= 6
    },
)

// 2. 使用 Computed 值
isValid := emailValidComputed.Get()

// 3. 在 state 中缓存计算结果
type AppState struct {
    Email     string
    EmailValid bool  // 缓存 Computed 结果
}
```

##### 示例中的实现

- **main.go:110-113** - emailValidComputed: 验证邮箱格式
- **main.go:115-118** - formCompleteComputed: 检查表单完整性
- **main.go:120-122** - formStrengthComputed: 计算表单强度
- **main.go:295-306** - UpdateComputedIntent handler 更新缓存
- **main.go:392-399** - 在 View 中显示计算值

##### 常见使用场景

```go
// 场景1: 验证状态
emailValid := NewComputed(store, func(s) bool {
    return isValidEmail(s.Email)
})

// 场景2: 数据转换
formattedDate := NewComputed(store, func(s) string {
    return s.CreatedAt.Format("2006-01-02")
})

// 场景3: 组合数据
fullName := NewComputed(store, func(s) string {
    return s.FirstName + " " + s.LastName
})

// 场景4: 复杂逻辑
orderTotal := NewComputed(store, func(s) float64 {
    total := 0
    for _, item := range s.Items {
        total += item.Price * item.Quantity
    }
    return total * (1 - s.Discount)
})
```

##### 效果

- 减少重复计算，提升性能
- 代码更清晰，关注点分离
- 自动跟踪依赖，无需手动管理更新

---

#### 5️⃣ Middleware - Reducer中间件

##### 功能描述

Middleware 提供在 Reducer 执行前后插入自定义逻辑的能力，常用于日志记录、异常处理、性能监控等。

##### 基本概念

Middleware是一个高阶函数，接收当前Reducer并返回一个增强后的Reducer：

```
middleware(next) reducer.Reducer
  调用 orderTotal -> return enhancedReducer
```

##### 使用方法

```go
// 1. 定义 Middleware
func loggingMiddleware(next reducer.Reducer[AppState]) reducer.Reducer[AppState] {
    return func(state AppState, i intent.Intent) AppState {
        // 前置逻辑
        fmt.Printf("Before: %s\n", i.IntentType())

        // 执行下一个 Reducer/Middleware
        newState := next(state, i)

        // 后置逻辑
        fmt.Printf("After: %+v\n", newState)

        return newState
    }
}

// 2. 应用 Middleware
var AppReducer = reducer.WithMiddleware(
    appReducerBuilder.Build(),  // 基础 Reducer
    loggingMiddleware,          // 第一个 Middleware
    performanceMiddleware,      // 第二个 Middleware
)

// 3. 中间件链执行顺序
State = loggingMiddleware(
    performanceMiddleware(
        appReducerBuilder.Build()
    )
)(state, intent)
```

##### 示例中的实现

- **main.go:231-255** - loggingMiddleware 实现
  - 记录动作类型和时间戳
  - 维护 actionLog 用于UI显示
  - 更新 LaneStatus 状态

- **main.go:310-311** - 应用 Middleware
  ```go
  var AppReducer = reducer.WithMiddleware(
      appReducerBuilder.Build(),
      loggingMiddleware,
  )
  ```

- **main.go:567-571** - 显示 action 日志

##### 常见使用场景

```go
// 场景1: 日志记录
func loggingMiddleware(next Reducer) Reducer {
    return func(state, intent) State {
        logger.Infof("Action: %s", intent.Type())
        return next(state, intent)
    }
}

// 场景2: 性能监控
func timingMiddleware(next Reducer) Reducer {
    return func(state, intent) State {
        start := time.Now()
        result := next(state, intent)
        duration := time.Since(start)
        metrics.Record("reducer.duration", duration)
        return result
    }
}

// 场景3: 异常处理
func errorHandlingMiddleware(next Reducer) Reducer {
    return func(state, intent) State {
        defer func() {
            if r := recover(); r != nil {
                logger.Errorf("Reducer panic: %v", r)
                sentry.CaptureException(r)
            }
        }()
        return next(state, intent)
    }
}

// 场景4: 数据验证
func validationMiddleware(next Reducer) Reducer {
    return func(state, intent) State {
        if !validate(state) {
            return state  // 拒绝无效状态
        }
        return next(state, intent)
    }
}

// 场景5: 状态快照
func snapshotMiddleware(next Reducer) Reducer {
    return func(state, intent) State {
        before := state
        after := next(state, intent)
        if before != after {
            saveSnapshot(after)
        }
        return after
    }
}
```

##### 中间件组合示例

```go
// 多个 Middleware 组合
var AppReducer = reducer.WithMiddleware(
    appReducerBuilder.Build(),
    loggingMiddleware,      // 外层 - 首先执行
    errorHandlingMiddleware, // 中层
    performanceMiddleware,   // 内层 - 最后执行
)

// 执行顺序：
// 1. loggingMiddleware 前置逻辑
// 2. errorHandlingMiddleware 前置逻辑
// 3. performanceMiddleware 前置逻辑
// 4. appReducerBuilder.Build()
// 5. performanceMiddleware 后置逻辑
// 6. errorHandlingMiddleware 后置逻辑
// 7. loggingMiddleware 后置逻辑
```

##### 效果

- 集中管理横切关注点（日志、监控、验证）
- 不修改核心业务逻辑
- 易于扩展和测试
- 代码复用性高

---

### 高级特性总结

| 特性 | 核心价值 | 使用门槛 | 推荐场景 |
|------|---------|---------|---------|
| **Lane Scheduler** | 性能优化 | 低 | 复杂UI、多任务应用 |
| **WithDefaultLane** | 配置简化 | 低 | 所有应用 |
| **PluginSetup** | 扩展性 | 中 | 需要插件的应用 |
| **Time Travel** | 调试/恢复 | 低 | 所有应用（推荐） |
| **Computed/Selector** | 性能+可维护性 | 中 | 复杂计算、大数据 |
| **Middleware** | 关注点分离 | 低 | 所有应用（推荐） |



## 🔧 迁移指南

### 从 ui.Run + 全局 Store 迁移

#### 步骤 1：移除全局 Store

```diff
- var appStore *store.Store[AppState]
```

#### 步骤 2：修改 View 函数签名

```diff
- func App() ui.VNode {
-     state := appStore.Get()
-     return renderForm(state)
+ func AppView(state AppState) any {
+     return renderAppView(state)
+ }
+
+ func renderAppView(state AppState) ui.VNode {
      return renderForm(state)
  }
```

#### 步骤 3：修改 main 函数

```diff
- func main() {
-     fieldMap.RegisterToGlobal(appStore)
-     ui.Run(App, ...)
- }
+ func main() {
+     rt := statemachine.NewAppRuntime(
+         AppState{},
+         AppView,
+         AppReducer,
+     )
+     ui.RunApp(rt, WithInit(func() {
+         fieldMapBuilder.RegisterToGlobal(rt.GetStore())
+     }))
+ }
```

#### 步骤 4：移除所有 fmt.Println

```diff
- On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
-     fmt.Println("Form submitted")
-     return s
- })
+ type AppState struct {
+     Submitted bool
+ }
+
+ On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
+     s.Submitted = true
+     return s
+ })
+ 
+ // 在 View 中显示
+ if state.Submitted {
+     ui.NewTextBuilder("✅ Form Submitted").
+         FgColor("green").
+         Build()
+ }
```

---

## 📊 特性对比

### ui.Run vs ui.RunApp[T]

| 功能 | ui.Run | ui.RunApp[T] |
|------|--------|--------------|
| **全局 Store** | 需要 | ❌ 不需要 |
| **手动读取状态** | 需要 | ❌ 不需要 |
| **自动状态订阅** | ❌ | ✅ 自动 |
| **自动重新渲染** | ❌ | ✅ 自动 |
| **时间旅行调试** | 手动实现 | ✅ 内置 (WithMaxHistory) |
| **代码复杂度** | 中等 | 低 |
| **最佳适用场景** | 灵活需求、过渡期 | 新项目、推荐 |

### FieldMap vs switch-case

| 指标 | switch-case | FieldMap |
|------|-----------|----------|
| **代码行数** | ~40 行 | ~20 行 **(-50%)** |
| **类型断言** | 需要（运行时） | ❌ 不需要 |
| **编译时检查** | ❌ 无 | ✅ 有 |
| **扩展新字段** | 修改 3+ 处 | **仅添加 1 行** |
| **维护难度** | 中等 | 低 |

---

## 🎯 最佳实践

### 表单状态设计

```go
type AppState struct {
    // 表单字段
    Username     string
    Email        string
    Phone        string
    Age          int
    City         string
    
    // 布尔字段
    AcceptTerms  bool
    Subscribe    bool
    
    // 表单状态
    Submitted    bool
    SubmittedAt  time.Time
    
    // 验证错误
    Errors       map[string]string
}
```

### 验证模式

```go
// 在 Reducer 中验证
var appReducerBuilder = fieldMapBuilder.GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        errors := make(map[string]string)
        
        if len(s.Username) < 3 {
            errors["Username"] = "用户名至少3个字符"
        }
        
        if s.Email == "" || !containsAt(s.Email) {
            errors["Email"] = "邮箱格式不正确"
        }
        
        if s.AcceptTerms == false {
            errors["AcceptTerms"] = "必须接受条款"
        }
        
        if len(errors) > 0 {
            // 有错误，不提交
        } else {
            s.Submitted = true
            s.SubmittedAt = time.Now()
        }
        
        return s
    })
```

### 错误显示

```go
func renderErrors(errors map[string]string) ui.VNode {
    var lines []ui.VNode
    
    for field, errMsg := range errors {
        lines = append(lines,
            ui.NewTextBuilder("⚠ " + errMsg).
                FgColor("red").
                Build(),
        )
    }
    
    return ui.VStack(lines...)
}
```

---

## 📚 扩展阅读

### 核心文档

- **RunApp 指南**：`docs/architecture/store/RUNAPP_GUIDE.md`
- **FieldMap 优化**：`docs/architecture/store/FIELD_BINDING_OPTIMIZATION.md`
- **类型优化历史**：`docsArchive/cleanup-2026-05-19/docs/ui/store/optimization/APPVIEW_TYPE_OPTIMIZATION.md`
- **fmt.Println 问题历史**：`docsArchive/cleanup-2026-05-19/docs/ui/store/issues/FMT_PRINT_ISSUE.md`
- **时间旅行调试修复历史**：`docsArchive/cleanup-2026-05-19/docs/ui/store/fixes/TIMETRAVEL_FIX.md`

### 相关示例

| 示例 | 位置 | 说明 |
|------|------|------|
| **本示例** | `examples/typesafe_form_demo_runapp/main.go` | RunApp + FieldMap 完整示例 |
| **基础 RunApp** | `examples/runapp_demo/main.go` | RunApp 基础示例（计数器） |
| **Store + Reducer** | `examples/store_reducer_demo/main.go` | 传统全局 Store 模式 |

---

## 🎉 总结

本示例展示了 Mint UI 的完整最佳实践：

### 核心架构（7个）

1. ✅ **ui.RunApp[T]** - 现代化的应用启动方式
2. ✅ **AppRuntime[T]** - 完整的状态管理运行时
3. ✅ **FieldMap** - 消除 switch-case 的字段处理优化（**核心特性**）
4. ✅ **ForField 绑定** - 自动化字段双向绑定
5. ✅ **类型安全 View 包装** - 避免循环依赖的同时保持类型检查
6. ✅ **UI 组件系统** - 丰富的表单组件
7. ✅ **正确的状态显示** - 在 UI 组件中显示而不是使用 fmt.Println

### 高级特性（6个）

8. ✅ **Lane Scheduler** - 优先级渲染调度，优化性能
9. ✅ **WithDefaultLane** - 设置默认渲染优先级
10. ✅ **PluginSetup** - 插件系统，支持Modal等扩展
11. ✅ **Time Travel Debugging** - 完整的Undo/Redo功能，支持状态历史回溯
12. ✅ **Computed/Selector** - 计算值优化，减少不必要的重渲染
13. ✅ **Middleware** - Reducer中间件，支持日志、监控等横切关注点

**这是 Mint UI 推荐的新项目开发模式！** 🚀

---

## 📥 快速迁移检查清单

### 基础架构迁移

从其他示例迁移到本示例的模式：

- [ ] 移除全局 `appStore`
- [ ] 创建 `AppRuntime` 实例
- [ ] 将 `App()` 改为 `AppView(state AppState) any`
- [ ] 添加 `renderAppView(state AppState) ui.VNode` 内部函数
- [ ] 使用 `fieldMapBuilder` 而不是 `fieldMap`
- [ ] 在 `ui.WithInit` 中调用 `RegisterToGlobal(rt.GetStore())`
- [ ] 移除所有 `fmt.Println` 改为状态字段
- [ ] 使用 `ui.NewTextBuilder()` 添加样式
- [ ] 使用 `FgColor()` 设置颜色

### 高级特性启用（可选）

如果想启用高级特性：

- [ ] **Lane Scheduler**: 使用 `ui.WithLaneScheduler()` 配置
- [ ] **WithDefaultLane**: 使用 `ui.WithDefaultLane()` 设置默认优先级
- [ ] **PluginSetup**: 使用 `ui.WithPluginSetup()` 启用插件系统
- [ ] **Time Travel**: 使用 `statemachine.WithMaxHistory(100)` 启用历史记录
- [ ] **Computed**: 使用 `store.NewComputed()` 创建计算值
- [ ] **Middleware**: 使用 `reducer.WithMiddleware()` 添加中间件

**开始使用**：`go run ./examples/typesafe_form_demo_runapp/`
