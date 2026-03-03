# Intent Handler 迁移指南

**版本**: v2.0
**创建日期**: 2026-03-04
**状态**: 活跃

---

## 概述

本文档描述如何将组件从旧的闭包式 `On` API 迁移到新的 `On` API（接受 `*ActionContext` 参数）。

### 变更摘要

| 项目 | 旧 API | 新 API |
|------|--------|--------|
| 函数签名 | `On(T, func())` | `On(T, func(*ActionContext))` |
| 类型约束 | `StayPressed() bool` 必需 | 仅需 `IntentType() string` |
| 状态访问 | 闭包捕获（有风险） | 从 ActionContext 读取（安全） |

---

## 迁移步骤

### 步骤 1: 更新 Intent 类型定义

**旧代码**：必须实现 `StayPressed()`

```go
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }  // 必需
```

**新代码**：`StayPressed()` 变为可选

```go
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
// StayPressed() 可选：仅当需要控制按钮视觉反馈时实现
func (IncrementIntent) StayPressed() bool  { return true }
```

---

### 步骤 2: 更新 Handler 注册

**旧代码**：闭包捕获状态

```go
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    ui.On(IncrementIntent{}, func() {
        setCount(func(c int) int { return c + 1 })
    })

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

**新代码**：从 ActionContext 读取状态

```go
import "github.com/wwsheng009/mint/runtime/intent"

func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    // 1. 保存 setter 到 Context
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.SetState("setCount", setCount)
    }

    // 2. Handler 从 Context 读取
    ui.On(IncrementIntent{}, func(ctx *intent.ActionContext) {
        if fn, ok := ctx.GetState("setCount"); ok {
            if setter, ok := fn.(func(func(int) int)); ok {
                setter(func(c int) int { return c + 1 })
            }
        }
    })

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

---

### 步骤 3: 使用 StateKey 实现类型安全（推荐）

```go
import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui"
)

// 定义类型安全的状态键
var CountKey = ui.NewStateKey[int]("count")
var SetCountKey = ui.NewStateKey[func(func(int) int)]("setCount")

func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    ctx := ui.GetCurrentContext()
    if ctx != nil {
        CountKey.Set(ctx, count)
        SetCountKey.Set(ctx, setCount)
    }

    ui.On(IncrementIntent{}, func(ctx *intent.ActionContext) {
        if setter := SetCountKey.Get(ctx, nil); setter != nil {
            setter(func(c int) int { return c + 1 })
        }
    })

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

---

## 常见模式迁移

### 模式 1: 简单计数器

**旧代码**：
```go
ui.On(IncrementIntent{}, func() {
    setCount(count + 1)  // ⚠️ 闭包捕获 count，可能有陈旧值
})
```

**新代码**：
```go
ctx := ui.GetCurrentContext()
if ctx != nil {
    ctx.SetState("setCount", setCount)
}

ui.On(IncrementIntent{}, func(ctx *intent.ActionContext) {
    // ✅ 使用函数式更新，无需读取当前值
    if fn, ok := ctx.GetState("setCount"); ok {
        if setter, ok := fn.(func(func(int) int)); ok {
            setter(func(c int) int { return c + 1 })
        }
    }
})
```

### 模式 2: 多状态更新

**旧代码**：
```go
ui.On(ResetIntent{}, func() {
    setUsername("")
    setEmail("")
    setAge(0)
})
```

**新代码**：
```go
ctx := ui.GetCurrentContext()
if ctx != nil {
    ctx.SetState("setUsername", setUsername)
    ctx.SetState("setEmail", setEmail)
    ctx.SetState("setAge", setAge)
}

ui.On(ResetIntent{}, func(ctx *intent.ActionContext) {
    if fn, ok := ctx.GetState("setUsername"); ok {
        if setter, ok := fn.(func(string)); ok {
            setter("")
        }
    }
    if fn, ok := ctx.GetState("setEmail"); ok {
        if setter, ok := fn.(func(string)); ok {
            setter("")
        }
    }
    if fn, ok := ctx.GetState("setAge"); ok {
        if setter, ok := fn.(func(int)); ok {
            setter(0)
        }
    }
})
```

### 模式 3: 条件逻辑

**旧代码**：
```go
ui.On(SubmitIntent{}, func() {
    if len(username) >= 3 {
        submitForm(username, email)
    }
})
```

**新代码**：
```go
ctx := ui.GetCurrentContext()
if ctx != nil {
    ctx.SetState("username", username)
    ctx.SetState("email", email)
    ctx.SetState("submitForm", submitForm)
}

ui.On(SubmitIntent{}, func(ctx *intent.ActionContext) {
    username := ctx.GetStringState("username", "")
    email := ctx.GetStringState("email", "")
    
    if len(username) >= 3 {
        if fn, ok := ctx.GetState("submitForm"); ok {
            if submit, ok := fn.(func(string, string)); ok {
                submit(username, email)
            }
        }
    }
})
```

---

## Setter 类型参考

| Hook | Setter 类型 |
|------|-------------|
| `UseState[T]` | `func(T)` 或 `func(func(T) T)` |
| `UseStateInt` | `func(int)` 或 `func(func(int) int)` |
| `UseStateString` | `func(string)` 或 `func(func(string) string)` |
| `UseStateBool` | `func(bool)` 或 `func(func(bool) bool)` |

### 类型断言示例

```go
// 直接设置值
if setter, ok := fn.(func(int)); ok {
    setter(10)
}

// 函数式更新
if setter, ok := fn.(func(func(int) int)); ok {
    setter(func(c int) int { return c + 1 })
}
```

---

## 最佳实践

### 1. 优先使用函数式更新

函数式更新不依赖闭包中的当前值，更加安全：

```go
// ✅ 推荐：函数式更新
setter(func(c int) int { return c + 1 })

// ❌ 避免：直接使用闭包值
setter(count + 1)  // count 可能是陈旧值
```

### 2. 状态键命名约定

```go
// 状态值
ctx.SetState("form_username", username)

// Setter 函数
ctx.SetState("form_setUsername", setUsername)
```

### 3. 使用 StateKey 获得类型安全

```go
var UsernameKey = ui.NewStateKey[string]("username")

// 类型安全的读取
username := UsernameKey.Get(ctx, "")

// 类型安全的写入
UsernameKey.Set(ctx, "new_value")
```

---

## 迁移检查清单

- [ ] 更新 Intent 类型定义（移除不必要的 `StayPressed()`）
- [ ] 导入 `runtime/intent` 包
- [ ] 在组件函数开头保存状态到 Context
- [ ] 更新 `On` 调用，添加 `*ActionContext` 参数
- [ ] 从 Context 读取状态，避免闭包捕获
- [ ] 编译并测试

---

## 常见问题

### Q: 为什么要迁移？

A: 旧 API 存在闭包陈旧值问题。当组件重新渲染时，handler 中捕获的变量可能是旧值，导致状态更新不正确。

### Q: `StayPressed()` 还有用吗？

A: 有用，但是可选的。它控制 TUI 按钮按下后的视觉反馈状态：
- `StayPressed() == true`: 保持按下状态（导航、状态更新）
- `StayPressed() == false`: 立即重置（Quit、Delete）

### Q: 所有组件都需要迁移吗？

A: 建议迁移所有使用 `ui.On` 的组件。对于简单的函数式更新，迁移风险较低。

---

## 相关文档

- [REFACTOR_PLAN.md](./REFACTOR_PLAN.md) - 完整重构计划
- [store.md](./store.md) - Store + Reducer 设计
- [MVP_MIGRATION_GUIDE.md](../mvp/MVP_MIGRATION_GUIDE.md) - MVP 基础迁移

---

**最后更新**: 2026-03-04
