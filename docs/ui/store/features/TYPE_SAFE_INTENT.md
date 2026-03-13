# 类型安全 Intent DSL

**版本**: v1.0
**创建时间**: 2026-03-04
**状态**: ✅ 已实现

---

## 概述

`TypedFieldChange[T]` 和 `StateKey[T]` 提供编译期类型安全的状态访问，消除字符串键带来的运行时错误风险。

### 核心价值

| 特性 | 字符串键 | StateKey[T] |
|------|---------|-------------|
| 类型安全 | ❌ 运行时检查 | ✅ 编译期检查 |
| IDE 支持 | ❌ 无自动补全 | ✅ 完整支持 |
| 重构安全 | ❌ 字符串难追踪 | ✅ 编译器验证 |
| 文档化 | ❌ 分散 | ✅ 集中定义 |

---

## 快速开始

### 1. 定义状态键

```go
// app/keys.go
package app

import "github.com/wwsheng009/mint/runtime/intent"

var (
    // 表单字段
    Username = intent.NewStateKey[string]("username")
    Email    = intent.NewStateKey[string]("email")
    Age      = intent.NewStateKey[int]("age")
    Active   = intent.NewStateKey[bool]("active")

    // 复杂类型
    User     = intent.NewStateKey[User]("user")
    Settings = intent.NewStateKey[Settings]("settings")
)
```

### 2. 在组件中使用

```go
func Form(state AppState) ui.VNode {
    ctx := ui.GetCurrentContext()

    // 保存状态供 handler 访问
    Username.Set(ctx, state.Username)
    Email.Set(ctx, state.Email)

    // 类型安全地读取
    username := Username.Get(ctx, "")

    return ui.VStack(
        input.ForField(Username.String()).Value(state.Username),
        input.ForField(Email.String()).Value(state.Email),
    )
}
```

### 3. 发射类型安全 Intent

```go
// 方式 1: 使用 Change 方法
intent := Username.Change("alice")

// 方式 2: 使用 SetField 函数
intent := intent.SetField(Username, "alice")

// 方式 3: 在组件中使用
input.ForField(Username.String()).OnChange(func(v string) {
    dispatcher.Dispatch(Username.Change(v))
})
```

### 4. 在 Reducer 中处理

```go
func AppReducer(state AppState, i intent.Intent) AppState {
    switch v := i.(type) {
    case intent.TypedFieldChange[string]:
        switch v.Key.String() {
        case Username.String():
            state.Username = v.Value
            // 实时验证
            if len(state.Username) < 3 {
                state.UsernameErr = "用户名至少3字符"
            } else {
                state.UsernameErr = ""
            }
        case Email.String():
            state.Email = v.Value
        }

    case intent.TypedFieldChange[int]:
        if v.Key.String() == Age.String() {
            state.Age = v.Value
        }

    case intent.TypedFieldChange[bool]:
        if v.Key.String() == Active.String() {
            state.Active = v.Value
        }

    case SubmitIntent:
        // 提交验证
        state.UsernameErr = validateUsername(state.Username)
        state.EmailErr = validateEmail(state.Email)
    }

    return state
}
```

---

## API 参考

### StateKey[T]

```go
// 创建状态键
var Username = intent.NewStateKey[string]("username")

// 方法
Username.String()              // "username" - 字符串表示
Username.Name()                // "username" - 键名
Username.Get(ctx, "")          // 从 Context 读取
Username.Set(ctx, "alice")     // 写入 Context
Username.Change("alice")       // 创建 TypedFieldChange intent
```

### TypedFieldChange[T]

```go
// 创建
intent := intent.TypedFieldChange[string]{
    Key:   Username,
    Value: "alice",
}

// 便捷函数
intent := intent.SetField(Username, "alice")
intent := intent.UpdateField[string]("username", "alice")

// 实现 Intent 接口
intent.IntentType() // "TypedFieldChange"
```

### MultiFieldChange

```go
// 批量更新
changes := intent.MultiFieldChange{
    Username.Change("alice"),
    Age.Change(25),
    Active.Change(true),
}

// 链式添加
changes := intent.MultiFieldChange{}.
    Add(Username.Change("alice"))
changes = intent.AddTyped(changes, Age, 25)
```

---

## 完整示例

### 计数器应用

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui"
)

// 状态键定义
var Count = intent.NewStateKey[int]("count")

// 应用状态
type AppState struct {
    Count int
}

// Intent 类型
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }

type DecrementIntent struct{}
func (DecrementIntent) IntentType() string { return "Decrement" }

// Reducer
func reducer(state AppState, i intent.Intent) AppState {
    switch i.(type) {
    case IncrementIntent:
        state.Count++
    case DecrementIntent:
        state.Count--
    case intent.TypedFieldChange[int]:
        if tfc, ok := i.(intent.TypedFieldChange[int]); ok {
            if tfc.Key.String() == Count.String() {
                state.Count = tfc.Value
            }
        }
    }
    return state
}

// 视图
func Counter(state AppState) ui.VNode {
    ctx := ui.GetCurrentContext()
    Count.Set(ctx, state.Count)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", state.Count)),
        ui.HStack(
            ui.NewButtonBuilder("-").OnPress(DecrementIntent{}).Build(),
            ui.NewButtonBuilder("+").OnPress(IncrementIntent{}).Build(),
        ),
    )
}
```

### 表单验证

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui"
)

// 状态键
var (
    Username   = intent.NewStateKey[string]("username")
    Email      = intent.NewStateKey[string]("email")
    UsernameErr = intent.NewStateKey[string]("username_err")
    EmailErr    = intent.NewStateKey[string]("email_err")
)

// 应用状态
type FormState struct {
    Username    string
    Email       string
    UsernameErr string
    EmailErr    string
}

// Intent 类型
type SubmitIntent struct{}
func (SubmitIntent) IntentType() string { return "Submit" }

type ResetIntent struct{}
func (ResetIntent) IntentType() string { return "Reset" }

// Reducer - 逻辑集中
func formReducer(state FormState, i intent.Intent) FormState {
    switch v := i.(type) {
    case intent.TypedFieldChange[string]:
        switch v.Key.String() {
        case Username.String():
            state.Username = v.Value
            // 实时验证
            if len(state.Username) < 3 {
                state.UsernameErr = "用户名至少3字符"
            } else {
                state.UsernameErr = ""
            }
        case Email.String():
            state.Email = v.Value
            // 邮箱格式验证
            if !isValidEmail(state.Email) {
                state.EmailErr = "邮箱格式错误"
            } else {
                state.EmailErr = ""
            }
        }

    case SubmitIntent:
        // 提交时验证所有
        if len(state.Username) < 3 {
            state.UsernameErr = "用户名至少3字符"
        }
        if !isValidEmail(state.Email) {
            state.EmailErr = "邮箱格式错误"
        }
        if state.UsernameErr == "" && state.EmailErr == "" {
            // 提交成功
            fmt.Printf("提交: %+v\n", state)
        }

    case ResetIntent:
        return FormState{}
    }

    return state
}

// 视图 - 纯声明
func Form(state FormState) ui.VNode {
    return ui.VStack(
        ui.Text("注册表单"),
        input.ForField(Username.String()).
            Placeholder("用户名").
            Value(state.Username),
        ui.Text(state.UsernameErr).Color("red"),
        input.ForField(Email.String()).
            Placeholder("邮箱").
            Value(state.Email),
        ui.Text(state.EmailErr).Color("red"),
        ui.HStack(
            ui.NewButtonBuilder("重置").OnPress(ResetIntent{}).Build(),
            ui.NewButtonBuilder("提交").OnPress(SubmitIntent{}).Build(),
        ),
    )
}
```

---

## 迁移指南

### 从字符串键迁移

```go
// ❌ 旧代码（字符串键）
func handler(ctx *intent.ActionContext) {
    username := ctx.GetStringState("username", "")
    // 拼写错误风险："usrename"
}

// ✅ 新代码（类型安全）
var Username = intent.NewStateKey[string]("username")

func handler(ctx *intent.ActionContext) {
    username := Username.Get(ctx, "")
    // 编译期检查，拼写错误会报错
}
```

### 从 FieldChangeIntent 迁移

```go
// ❌ 旧代码
type FieldChangeIntent struct {
    Field string  // 字符串，易出错
    Value string  // 无类型
}

ui.On(FieldChangeIntent{}, func(ctx *intent.ActionContext) {
    // 需要手动类型转换
    if i.Field == "age" {
        age, _ := strconv.Atoi(i.Value)
    }
})

// ✅ 新代码
ui.On(intent.TypedFieldChange[int]{}, func(ctx *intent.ActionContext) {
    // 类型已经确定，无需转换
})

// 发射
dispatcher.Dispatch(intent.SetField(Age, 25))  // 类型安全
```

---

## 最佳实践

### 1. 集中定义状态键

```go
// ✅ 好的做法 - 集中在 keys.go
var (
    Username = intent.NewStateKey[string]("username")
    Email    = intent.NewStateKey[string]("email")
)

// ❌ 避免 - 分散定义
func ComponentA() {
    username := intent.NewStateKey[string]("username")  // 重复定义
}
```

### 2. 使用有意义的前缀

```go
// ✅ 好的做法 - 语义化前缀
var (
    FormUsername = intent.NewStateKey[string]("form.username")
    FormEmail    = intent.NewStateKey[string]("form.email")
    ModalVisible = intent.NewStateKey[bool]("modal.visible")
)
```

### 3. 类型一致性

```go
// ✅ 好的做法 - 类型一致
var Age = intent.NewStateKey[int]("age")
// 读取和设置都使用 int
Age.Set(ctx, 25)
age := Age.Get(ctx, 0)

// ❌ 避免 - 类型不一致
ctx.SetState("age", "25")  // 字符串
Age.Get(ctx, 0)            // 期望 int
```

---

## 设计决策

### 为什么 TypedFieldChange[T] 是泛型？

1. **类型安全**: 编译期验证类型匹配
2. **IDE 支持**: 自动补全和类型推断
3. **无运行时开销**: 泛型在编译期展开

### 为什么保留字符串键？

1. **向后兼容**: 现有代码无需立即迁移
2. **动态场景**: 某些场景需要动态键名
3. **渐进迁移**: 可以逐步迁移到类型安全

---

## 相关文档

- [INTENT_HANDLER_MIGRATION.md](../migration/INTENT_HANDLER_MIGRATION.md) - Intent Handler 迁移指南
- [STORE_REDUCER_GUIDE.md](../guides/STORE_REDUCER_GUIDE.md) - Store + Reducer 完整指南
- [REFACTOR_PLAN.md](/docsArchive/REFACTOR_PLAN.md) - 架构重构计划

---

**最后更新**: 2026-03-04
