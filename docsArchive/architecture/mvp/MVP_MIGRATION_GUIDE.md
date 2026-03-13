# MVP 架构迁移指南

> ⚠️ **DEPRECATED** - 本文档已归档
>
> **新架构**: **Store + Reducer**
> **新迁移指南**: [`/docs/ui/store/guides/MIGRATION_GUIDE.md`](../../ui/store/guides/MIGRATION_GUIDE.md)
> **迁移进度**: [`/docs/ui/store/status/MIGRATION_PROGRESS.md`](../../ui/store/status/MIGRATION_PROGRESS.md)
>
> 本文档记录了 MVP 模式的早期探索。当前推荐使用 **Store + Reducer 架构**，它是 MVP 的完整进化版。

---

**创建时间**: 2026-02-26
**版本**: Phase 10
**归档时间**: 2026-03-08
**适用版本**: Mint UI v0.10+

---

## 一、概述

### 什么是 MVP 架构？

MVP (Model-View-Presenter) 架构的核心原则是：

1. **State（State + Setter）是单一事实源**
   - 所有状态存储在 State 中
   - 通过 Setter 更新状态，不允许直接修改
   - 触发 Setter 会自动触发重新渲染

2. **Intent 携带最少数据**
   - Intent 包含 Field（字段名）和 Value（运行时值）
   - Instance 不决定状态，只发射带值的 Intent
   - Handler 根据 Intent 更新 State

3. **Instance 是纯缓冲**
   - Instance 只负责渲染和收集用户输入
   - 不维护业务状态，只临时缓存显示值
   - 所有状态变更必须通过 Intent

### 数据流

```
┌─────────┐   FieldChangeIntent   ┌──────────┐   Setter   ┌─────┐
│ Instance│  (Field, Value)        │ Handler  │ ────────► │State│
│ (缓冲)  │ ──────────────────────► │          │           │(事实源)│
└─────────┘                      └──────────┘           └─────┘
     │                                                   │
     │                                                   ▼
     │                                              ┌─────────┐
     │  VNode 渲染同步                              │  VNode  │
     └──────────────────────────────────────────────► │ (描述)  │
                                                   └─────────┘
```

---

## 二、核心概念

### 1. StateKey[T] - 类型安全的字段键

Go 1.18+ 泛型支持编译期类型检查：

```go
// 定义字段键（通常在包级别）
var (
    username = intent.StateKey[string]("username")
    email    = intent.StateKey[string]("email")
    age      = intent.StateKey[int]("age")
    agree    = intent.StateKey[bool]("agree")
)
```

**好处**：
- 编译期类型检查，防止类型错误
- IDE 自动补全支持
- 重构安全（改键名会更新所有引用）

### 2. FieldChangeIntent - 统一的字段变更 Intent

所有表单组件的字段变更都通过同一个 Intent 类型：

```go
type FieldChangeIntent struct {
    Field string  // 字段名（或 StateKey.String()）
    Value string  // 运行时值（转换为字符串）
}
```

### 3. ForField() - 字段绑定 API

Builder API 提供 `ForField()` 方法，自动处理 Intent 发射：

```go
input.NewBuilder().
    ForField(intent.ForField(username)).  // 绑定到 StateKey
    Value(username).                      // 显示值
    Placeholder("Enter username").
    Build()
```

---

## 三、从旧 API 迁移

### 旧 vs 新 API 对照表

| 组件 | 旧 API | 新 API (MVP) |
|------|--------|--------------|
| Input | `OnChange(setUsername)` | `ForField(intent.ForField(username))` |
| Textarea | `OnChange(setBio)` | `ForField(intent.ForField(bio))` |
| Checkbox | `OnToggle(setAgree)` | `ForField(intent.ForField(agree))` |
| Select | `OnChange(setCountry)` | `ForField(intent.ForField(country))` |

---

## 四、组件迁移示例

### 示例 1: Input 组件

#### 旧 API（直接 Setter）

```go
func App() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")

    return ui.VStack(
        app.InputBuilder().
            Value(username).
            OnChange(setUsername).  // ❌ 直接传递 setter
            Placeholder("Username").
            Build(),

        app.InputBuilder().
            Value(email).
            OnChange(setEmail).      // ❌ 直接传递 setter
            Placeholder("Email").
            Build(),
    )
}
```

#### 新 API (MVP)

```go
// 1. 定义 StateKey
var usernameKey = intent.StateKey[string]("username")
var emailKey = intent.StateKey[string]("email")

func App() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")

    // 保存 setters 到 State 供 Handler 使用
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState[usernameKey.String()] = setUsername
        ctx.GlobalState[emailKey.String()] = setEmail
    }

    return ui.VStack(
        app.InputBuilder().
            ForField(intent.ForField(usernameKey)).  // ✅ 使用 ForField
            Value(username).
            Placeholder("Username").
            Build(),

        app.InputBuilder().
            ForField(intent.ForField(emailKey)).      // ✅ 使用 ForField
            Value(email).
            Placeholder("Email").
            Build(),
    )
}

// 2. 注册统一的 Handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
    switch i.Field {
    case usernameKey.String():
        setter, _ := ctx.GetState(usernameKey.String())
        callSetter(setter, i.Value)
    case emailKey.String():
        setter, _ := ctx.GetState(emailKey.String())
        callSetter(setter, i.Value)
    }
    return intent.HandledResult()
})

func callSetter(fn interface{}, arg interface{}) {
    v := reflect.ValueOf(fn)
    if v.Kind() == reflect.Func {
        v.Call([]reflect.Value{reflect.ValueOf(arg)})
    }
}
```

---

### 示例 2: Textarea 组件

#### 旧 API

```go
func App() ui.VNode {
    bio, setBio := ui.UseStateString("")

    return ui.VStack(
        app.NewTextBuilder("Bio:").Build(),
        app.TextareaBuilder().
            ForField(intent.BindField("bio")).  // ❌ 使用旧 API
            OnChange(setBio).                    // ❌ 冗余
            Value(bio).
            Rows(5).
            Cols(40).
            Build(),
    )
}
```

#### 新 API (MVP)

```go
var bioKey = intent.StateKey[string]("bio")

func App() ui.VNode {
    bio, setBio := ui.UseStateString("")

    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState[bioKey.String()] = setBio
    }

    return ui.VStack(
        app.NewTextBuilder("Bio:").Build(),
        app.TextareaBuilder().
            ForField(intent.ForField(bioKey)).  // ✅ 类型安全
            Value(bio).
            Rows(5).
            Cols(40).
            Build(),
    )
}
```

---

### 示例 3: Select 组件

#### 旧 API

```go
func App() ui.VNode {
    countries := []selectcomp.Option{
        {Value: "us", Label: "USA"},
        {Value: "cn", Label: "China"},
    }

    country, setCountry := ui.UseStateInt(0)

    return ui.VStack(
        app.SelectBuilder().
            Options(countries).
            Selected(country).
            OnChange(setCountry).  // ❌ 直接 setter
            Build(),
    )
}
```

#### 新 API (MVP)

```go
var countryKey = intent.StateKey[int]("country")

func App() ui.VNode {
    countries := []selectcomp.Option{
        {Value: "us", Label: "USA"},
        {Value: "cn", Label: "China"},
    }

    country, setCountry, _ := ui.UseStateInt(0)

    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState[countryKey.String()] = setCountry
    }

    return ui.VStack(
        app.SelectBuilder().
            Options(countries).
            Selected(country).
            ForField(intent.ForField(countryKey)).  // ✅ 类型安全
            Build(),
    )
}
```

---

### 示例 4: Checkbox 组件

#### 旧 API

```go
func App() ui.VNode {
    agree, setAgree := ui.UseStateBool(false)

    return ui.VStack(
        app.CheckboxBuilder().
            OnToggle(setAgree).  // ❌ 直接 setter
            Checked(agree).
            Label("I agree").
            Build(),
    )
}
```

#### 新 API (MVP)

```go
var agreeKey = intent.StateKey[bool]("agree")

func App() ui.VNode {
    agree, setAgree := ui.UseStateBool(false)

    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState[agreeKey.String()] = setAgree
    }

    return ui.VStack(
        app.CheckboxBuilder().
            ForField(intent.ForField(agreeKey)).  // ✅ 类型安全
            Checked(agree).
            Label("I agree").
            Build(),
    )
}
```

---

## 五、Handler 模式

### 1. 反射调用 Setter

```go
func callSetter(fn interface{}, arg interface{}) {
    if fn == nil {
        return
    }
    v := reflect.ValueOf(fn)
    if v.Kind() != reflect.Func {
        return
    }
    argV := reflect.ValueOf(arg)
    v.Call([]reflect.Value{argV})
}
```

### 2. 统一的 FieldChangeIntent Handler

```go
ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
    field := i.Field
    value := i.Value

    var setter interface{}

    switch field {
    case usernameKey.String():
        setter, _ = ctx.GetState(usernameKey.String())
        callSetter(setter, value)
    case emailKey.String():
        setter, _ = ctx.GetState(emailKey.String())
        callSetter(setter, value)
    case ageKey.String():
        setter, _ = ctx.GetState(ageKey.String())
        if age, err := strconv.Atoi(value); err == nil {
            callSetter(setter, age)
        }
    case agreeKey.String():
        setter, _ = ctx.GetState(agreeKey.String())
        agreeVal := value == "true"
        callSetter(setter, agreeVal)
    }

    return intent.HandledResult()
})
```

---

## 六、完整示例

参见 `examples/mvp_components_demo/main.go`：

```go
func main() {
    err := ui.Run(App,
        ui.WithInit(func() {
            // 注册统一的 Handler
            ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
                // 根据字段名更新对应的 State
                // ...
                return intent.HandledResult()
            })
        }),
    )
}

func App() ui.VNode {
    // 使用 UseState 获取状态
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")
    bio, setBio := ui.UseStateString("")
    country, setCountry, _ := ui.UseStateInt(0)
    agree, setAgree := ui.UseStateBool(false)

    // 保存 setters 到 State
    // ...

    // 使用 ForField 绑定所有组件
    return ui.VStack(
        // Input
        app.InputBuilder().
            ForField(intent.ForField(usernameKey)).
            Value(username).
            Build(),

        // Textarea
        app.TextareaBuilder().
            ForField(intent.ForField(bioKey)).
            Value(bio).
            Build(),

        // Select
        app.SelectBuilder().
            ForField(intent.ForField(countryKey)).
            Options(countries).
            Selected(country).
            Build(),

        // Checkbox
        app.CheckboxBuilder().
            ForField(intent.ForField(agreeKey)).
            Checked(agree).
            Build(),
    )
}
```

---

## 七、最佳实践

### 1. 字段键定义

```go
// ✅ 推荐：包级别定义 StateKey
var (
    username = intent.StateKey[string]("username")
    email    = intent.StateKey[string]("email")
    age      = intent.StateKey[int]("age")
)

// ❌ 避免：魔数字符串
input.OnChange(setUsername)  // 无类型检查
```

### 2. Setter 管理

```go
// ✅ 推荐：统一初始化时保存所有 setters
if ctx := ui.GetCurrentContext(); ctx != nil {
    ctx.GlobalState[username.String()] = setUsername
    ctx.GlobalState[email.String()] = setEmail
    ctx.GlobalState[age.String()] = setAge
}

// ❌ 避免：在组件函数中分散保存
```

### 3. Handler 组织

```go
// ✅ 推荐：一个统一的 FieldChangeIntent Handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
    switch i.Field {
    case username.String():
        // 更新 username
    case email.String():
        // 更新 email
    }
    return intent.HandledResult()
})

// ❌ 避免：为每个组件分别注册 Handler
```

### 4. 值类型转换

```go
// ✅ 推荐：在 Handler 中统一处理类型转换
case ageKey.String():
    setter, _ := ctx.GetState(ageKey.String())
    if age, err := strconv.Atoi(value); err == nil {
        callSetter(setter, age)
    }

// ❌ 避免：在客户端组件中转换
```

---

## 八、常见问题

### Q1: 为什么要移除直接绑定 setter？

**A**:
1. **可测试性**: Intent 可以独立测试，不需要模拟 State
2. **可预测性**: 所有状态变更都经过唯一入口，便于调试
3. **可扩展性**: 可以轻松添加全局验证、日志、持久化等中间件

### Q2: 为什么要用 StateKey[T]？

**A**:
1. **类型安全**: 编译期检查，避免运行时错误
2. **IDE 支持**: 自动补全和重构
3. **文档性**: 类型本身就是注释

### Q3: Handler 反射调用是否影响性能？

**A**:
1. 影响极小，反射调用通常在微秒级别
2. 如果担心性能，可以使用类型断言替代反射
3. 可以为高频组件使用专门的非反射 Handler

### Q4: 如何迁移现有代码？

**A**:
1. 添加 StateKey 定义
2. 将 `OnChange(setXxx)` 替换为 `ForField(intent.ForField(xxxKey))`
3. 注册统一的 FieldChangeIntent Handler
4. 测试验证

---

## 九、迁移检查清单

- [ ] 所有表单组件使用 `ForField()` 绑定
- [ ] 定义 `StateKey[T]` 类型安全字段键
- [ ] 注册统一的 `FieldChangeIntent` Handler
- [ ] 移除组件上的直接 setter 绑定（`OnChange(Setter)`）
- [ ] 验证 Handler 正确处理所有字段
- [ ] 添加单元测试（可选）

---

## 十、相关资源

- **详细组件检查报告**: `docs/architecture/COMPONENT_INTENT_REVIEW.md`
- **MVP 示例程序**: `examples/mvp_components_demo/main.go`
- **基础 MVP 示例**: `examples/mvp_form_demo/main.go`
- **Intent 定义**: `runtime/intent/field_change.go`

---

## 十一、版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| v0.10 | 2026-02-26 | Phase 10 完成，新增 MVP 组件综合示例 |
| v0.9 | 2026-02-26 | Phase 9 完成，修复 Textarea/Select/Tabs |
| v0.8 | 2026-02-26 | Phase 7 完成，Transition Intents |
| v0.7 | 2026-02-26 | Phase 6 完成，Builder API |
