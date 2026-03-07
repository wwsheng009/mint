# Hooks 使用指南

## 概述

Mint UI 的 Hooks 系统借鉴了 React Hooks 的设计理念，提供了一种在函数式组件中管理状态和副作用的方式。本文档介绍如何正确使用 Hooks 以及避免常见的陷阱。

---

## 核心 Hooks

### UseState

`UseState` 用于在组件中声明状态变量。

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/ui"
)

func CounterComponent(ctx *ui.Context) *ui.VNode {
    // 声明状态变量
    count, setCount := ui.UseState(ctx, 0)

    return ui.VNode{
        Type: "div",
        Children: []ui.VNode{
            ui.Text(fmt.Sprintf("Count: %d", count)),
            ui.Button("Increment", func() {
                setCount(count + 1)
            }),
        },
    }
}
```

#### 重要规则

1. **只在组件顶层调用 Hooks**

```go
// ❌ 错误：在循环中调用
for i := 0; i < 3; i++ {
    val, _ := ui.UseState(ctx, i) // 会导致状态错乱
}

// ✅ 正确：在顶层调用
val1, _ := ui.UseState(ctx, 1)
val2, _ := ui.UseState(ctx, 2)
val3, _ := ui.UseState(ctx, 3)
```

2. **不要在条件语句中调用 Hooks**

```go
// ❌ 错误：条件调用
if condition {
    val, _ := ui.UseState(ctx, 0)
}

// ✅ 正确：始终调用，内部处理条件
val, _ := ui.UseState(ctx, 0)
if condition {
    // 使用 val
}
```

---

### UseEffect

`UseEffect` 用于处理副作用，如订阅、定时器、数据获取等。

```go
func TimerComponent(ctx *ui.Context) *ui.VNode {
    count, setCount := ui.UseState(ctx, 0)

    // 设置定时器
    ui.UseEffect(ctx, func() func() {
        ticker := time.NewTicker(time.Second)
        go func() {
            for range ticker.C {
                setCount(count + 1)
            }
        }()

        // 清理函数
        return func() {
            ticker.Stop()
        }
    }, []interface{}{count}) // 依赖数组

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

#### 依赖数组规则

```go
// 每次渲染都执行
ui.UseEffect(ctx, fn, nil)

// 只执行一次（空依赖）
ui.UseEffect(ctx, fn, []interface{}{})

// 当依赖变化时执行
ui.UseEffect(ctx, fn, []interface{}{count, name})
```

---

### UseMemo

`UseMemo` 用于缓存计算结果，避免重复计算。

```go
func ExpensiveComponent(ctx *ui.Context) *ui.VNode {
    items, _ := ui.UseState(ctx, []int{1, 2, 3, 4, 5})

    // 只在 items 变化时重新计算
    sum := ui.UseMemo(ctx, func() interface{} {
        result := 0
        for _, item := range items {
            result += item
        }
        return result
    }, []interface{}{items})

    return ui.Text(fmt.Sprintf("Sum: %d", sum))
}
```

---

### UseCallback

`UseCallback` 用于缓存回调函数，避免不必要的重新渲染。

```go
func ListComponent(ctx *ui.Context) *ui.VNode {
    items, _ := ui.UseState(ctx, []string{"a", "b", "c"})

    // 缓存回调函数
    handleClick := ui.UseCallback(ctx, func(index int) {
        fmt.Println("Clicked:", items[index])
    }, []interface{}{items})

    // 使用缓存的回调
    children := make([]ui.VNode, len(items))
    for i, item := range items {
        children[i] = ui.Button(item, func() {
            handleClick(i)
        })
    }

    return ui.VNode{Type: "div", Children: children}
}
```

---

### UseRef

`UseRef` 用于创建可变的引用对象。

```go
func InputComponent(ctx *ui.Context) *ui.VNode {
    inputRef := ui.UseRef(ctx, "")

    return ui.VNode{
        Type: "input",
        Props: ui.Props{
            "onChange": func(v string) {
                inputRef.Current = v
            },
        },
    }
}
```

---

## 闭包陷阱及解决方案

### 问题：闭包捕获旧状态

```go
// ❌ 错误：事件处理器捕获闭包
func CounterComponent(ctx *ui.Context) *ui.VNode {
    count, setCount := ui.UseState(ctx, 0)

    // 问题：这个闭包捕获的是定义时的 count
    handler := func() {
        fmt.Println("Count:", count) // 可能是旧值！
        setCount(count + 1)          // 可能基于旧值更新！
    }

    return ui.Button("Click", handler)
}
```

### 解决方案 1：使用函数式更新

```go
// ✅ 正确：使用函数式更新
handler := func() {
    setCount(func(prev int) int {
        return prev + 1
    })
}
```

### 解决方案 2：使用 ActionContext

```go
// ✅ 正确：从 Context 读取最新状态
func CounterComponent(ctx *ui.Context) *ui.VNode {
    count, setCount := ui.UseState(ctx, 0)

    handler := func() {
        // 从 ctx 读取最新状态
        currentCount := ctx.GetState("count").(int)
        setCount(currentCount + 1)
    }

    return ui.Button("Click", handler)
}
```

### 解决方案 3：使用 Store + Reducer

```go
// ✅ 最佳：使用 Store + Reducer 模式
func main() {
    store := store.New(CounterState{Count: 0})

    reducer := func(state CounterState, intent intent.Intent) CounterState {
        switch i := intent.(type) {
        case IncrementIntent:
            state.Count++
        }
        return state
    }

    runtime := statemachine.NewAppRuntime(store, reducer)
    // ...
}
```

---

## Hooks 原理

### Fiber 节点中的 Hooks 存储

```
┌─────────────────────────────────────┐
│            FiberNode                │
├─────────────────────────────────────┤
│  Hooks []Hook                       │
│  ├── Hook{State: 0, Queue: nil}    │ <- UseState 第 1 次调用
│  ├── Hook{State: "", Queue: nil}   │ <- UseState 第 2 次调用
│  └── Hook{State: nil, Effect: fn}  │ <- UseEffect 调用
│                                     │
│  hookIndex int                      │ <- 当前 hooks 索引
└─────────────────────────────────────┘
```

### 调用顺序保证

```go
// 渲染函数每次调用时：
func renderRoot(fiber *FiberNode) {
    currentFiber = fiber
    hookIndex = 0  // 重置索引，但不清空 hooks 数组
    // ...渲染...
}

// UseState 内部：
func UseState[T any](ctx *Context, initial T) (T, func(T)) {
    hooks := currentFiber.Hooks

    if hookIndex >= len(hooks) {
        // 首次渲染：创建新 hook
        hooks = append(hooks, Hook{State: initial})
    }

    // 获取对应位置的 hook
    hook := hooks[hookIndex]
    hookIndex++

    return hook.State.(T), func(v T) {
        hook.State = v
        scheduleUpdate(currentFiber)
    }
}
```

---

## 最佳实践

### 1. 保持 Hooks 顺序稳定

```go
// ✅ 好的做法：稳定的 hooks 顺序
func Component(ctx *ui.Context) *ui.VNode {
    name, setName := ui.UseState(ctx, "")
    age, setAge := ui.UseState(ctx, 0)
    active, setActive := ui.UseState(ctx, false)
    // ...
}
```

### 2. 拆分复杂组件

```go
// ❌ 避免：单个组件使用过多 hooks
func BigComponent(ctx *ui.Context) *ui.VNode {
    // 20+ 个 UseState...
}

// ✅ 推荐：拆分为小组件
func ParentComponent(ctx *ui.Context) *ui.VNode {
    return ui.VNode{
        Type: "div",
        Children: []ui.VNode{
            UserComponent(ctx),
            SettingsComponent(ctx),
            ActionsComponent(ctx),
        },
    }
}
```

### 3. 正确设置依赖数组

```go
// ❌ 遗漏依赖
ui.UseEffect(ctx, func() func() {
    fetchData(userId) // userId 应该在依赖数组中
    return nil
}, []interface{}{})

// ✅ 包含所有依赖
ui.UseEffect(ctx, func() func() {
    fetchData(userId)
    return nil
}, []interface{}{userId})
```

### 4. 清理副作用

```go
// ✅ 总是清理资源
ui.UseEffect(ctx, func() func() {
    subscription := subscribe(eventHandler)
    return func() {
        subscription.Unsubscribe() // 清理
    }
}, []interface{}{})
```

---

## 常见错误排查

### 错误 1：状态更新后 UI 不刷新

**原因**：闭包捕获旧状态

**解决**：使用 Store + Reducer 模式，确保从单一来源读取状态

### 错误 2：状态混乱

**原因**：Hooks 调用顺序不一致

**解决**：确保 Hooks 只在顶层调用，不在条件语句或循环中调用

### 错误 3：内存泄漏

**原因**：未清理副作用（订阅、定时器等）

**解决**：在 UseEffect 返回的清理函数中释放资源

### 错误 4：无限循环

**原因**：UseEffect 的依赖数组包含每次渲染都变化的对象

**解决**：使用 UseMemo 缓存对象，或正确设置依赖

```go
// ❌ 无限循环
ui.UseEffect(ctx, func() func() {
    // ...
    return nil
}, []interface{}{make([]int, 0)}) // 每次都是新切片

// ✅ 使用 UseMemo
data := ui.UseMemo(ctx, func() interface{} {
    return make([]int, 0)
}, []interface{}{})

ui.UseEffect(ctx, func() func() {
    // ...
    return nil
}, []interface{}{data})
```

---

## 迁移指南

### 从旧 API 迁移

```go
// 旧 API（闭包问题）
func OldComponent(ctx *ui.Context) *ui.VNode {
    count := 0
    return ui.Button("Click", func() {
        count++ // 闭包问题
    })
}

// 新 API（状态管理）
func NewComponent(ctx *ui.Context) *ui.VNode {
    count, setCount := ui.UseState(ctx, 0)
    return ui.Button("Click", func() {
        setCount(func(prev int) int {
            return prev + 1
        })
    })
}

// 推荐方案（Store + Reducer）
func main() {
    store := store.New(AppState{Count: 0})
    reducer := func(s AppState, i intent.Intent) AppState {
        // 处理状态更新
        return s
    }
    runtime := statemachine.NewAppRuntime(store, reducer)
}
```

---

## 相关文档

- [Fiber 架构](./FIBER_ARCHITECTURE.md)
- [Store + Reducer 指南](./STORE_REDUCER_GUIDE.md)
- [Intent 迁移指南](./INTENT_HANDLER_MIGRATION.md)
- [重构计划](./REFACTOR_PLAN.md)
