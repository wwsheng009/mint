# 组件开发指南

本指南介绍如何使用 mint 的声明式 UI 框架开发具有持久化状态的组件。

## 目录

- [组件基础](#组件基础)
- [Hooks 状态管理](#hooks-状态管理)
- [组件实例](#组件实例)
- [最佳实践](#最佳实践)
- [内存安全](#内存安全)
- [示例](#示例)

## 组件基础

### 函数组件

mint 使用函数组件来构建 UI。一个函数组件是一个返回 VNode 的函数：

```go
func MyComponent() ui.VNode {
    return ui.Text("Hello, World!")
}
```

### 带 Props 的组件

使用 `ComponentWithProps` 创建接受参数的组件：

```go
type GreetingProps struct {
    Name string
}

func Greeting(props ui.Props) ui.VNode {
    name := props.GetString("name")
    if name == "" {
        name = "World"
    }
    return ui.Text(fmt.Sprintf("Hello, %s!", name))
}

// 使用
ui.ComponentWithProps("Greeting", Greeting).
    Prop("name", "Alice").
    Build()
```

## Hooks 状态管理

Hooks 是在函数组件中管理状态和副作用的机制。

### useState - 基础状态

`useState` 用于在组件中存储和更新状态：

```go
func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.Row(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("[+]").OnClick(func() {
            setCount(count + 1)
        }),
    )
}
```

**重要**: `useState` 必须在组件函数的顶层调用，不能在条件语句或循环中调用。

### useRef - 持久化引用

`useRef` 创建一个在渲染之间保持不变的引用：

```go
func TimerComponent() ui.VNode {
    countRef := ui.UseRef(0)

    // 更新 ref 不会触发重新渲染
    ui.UseEffect(func() func() {
        ticker := time.NewTicker(time.Second)
        go func() {
            for range ticker.C {
                val := countRef.Value.(int) + 1
                countRef.Value = val
            }
        }()
        return func() { ticker.Stop() }
    }, nil) // nil = 只运行一次

    return ui.Text(fmt.Sprintf("Ticks: %v", countRef.Value))
}
```

### useHoverState - 悬停状态

`useHoverState` 专门用于跟踪鼠标悬停状态：

```go
func HoverButton() ui.VNode {
    isHovered, setHovered := ui.UseHoverState()

    var btnStyle ui.StyleBuilder
    if isHovered() {
        btnStyle = ui.NewStyle().Background("cyan").Foreground("black")
    } else {
        btnStyle = ui.NewStyle().Background("blue").Foreground("white")
    }

    return ui.Button("Hover me").
        Style(btnStyle.Build()).
        OnMouseEnter(func() {
            setHovered(true)
        }).
        OnMouseLeave(func() {
            setHovered(false)
        })
}
```

### useEffect - 副作用

`useEffect` 用于处理副作用（网络请求、订阅、定时器等）：

```go
func DataFetcher() ui.VNode {
    data, setData := ui.UseStateString("")
    loading, setLoading := ui.UseStateBool(false)

    ui.UseEffect(func() func() {
        setLoading(true)

        // 模拟数据获取
        go func() {
            time.Sleep(time.Second)
            setData("Fetched data!")
            setLoading(false)
        }()

        // 清理函数
        return func() {
            fmt.Println("Cleanup")
        }
    }, nil) // nil deps = 只在挂载时运行

    if loading {
        return ui.Text("Loading...")
    }
    return ui.Text(data)
}
```

### useMemo 和 useCallback - 性能优化

```go
func ExpensiveComponent(a, b int) ui.VNode {
    // 只在 a 或 b 变化时重新计算
    result := ui.UseMemo(func() interface{} {
        return computeExpensive(a, b)
    }, []interface{}{a, b})

    // 只在 count 变化时创建新函数
    memoizedCallback := ui.UseCallback(func() {
        fmt.Println("Clicked")
    }, []interface{}{count})

    return ui.Text(fmt.Sprintf("Result: %v", result))
}
```

## 组件实例

### Key 属性

为组件设置 key 以确保状态在重新渲染时保持：

```go
func UserList(users []User) ui.VNode {
    children := make([]ui.VNode, len(users))
    for i, user := range users {
        children[i] = ui.ComponentWithProps("UserItem", UserItem).
            Prop("name", user.Name).
            Key(fmt.Sprintf("user-%d", user.ID)). // 重要：设置唯一 key
            Build()
    }
    return ui.Column(children...)
}
```

### 实例生命周期

每个组件实例都经历以下生命周期：

1. **挂载 (Mount)**: 组件首次渲染，`OnMount()` 被调用
2. **更新 (Update)**: props 变化时，`OnUpdate()` 被调用
3. **卸载 (Unmount)**: 组件被移除，`OnUnmount()` 被调用（清理 hooks）

```go
// BaseComponentInstance 默认实现
func (b *BaseComponentInstance) OnMount() {
    // 初始化逻辑
}

func (b *BaseComponentInstance) OnUpdate(newProps, oldProps Props) bool {
    // 返回 false 可以取消更新
    return true
}

func (b *BaseComponentInstance) OnUnmount() {
    // 所有 useEffect 的清理函数会被自动调用
}
```

## 最佳实践

### 1. Hooks 调用规则

- ✅ 只在组件函数的顶层调用 hooks
- ✅ 只在函数组件中调用 hooks
- ❌ 不要在循环、条件或嵌套函数中调用 hooks

```go
// 正确
func GoodComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)
    name, setName := ui.UseStateString("")

    return ui.Text(fmt.Sprintf("%s: %d", name, count))
}

// 错误
func BadComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    if count > 0 {
        // ❌ 不要在条件语句中调用 hook
        name, setName := ui.UseStateString("")
    }
    return ui.Text("...")
}
```

### 2. Key 策略

- 为动态列表中的每个项提供稳定的 key
- 使用唯一标识符（如 ID）而非数组索引
- 确保同一组件在不同渲染中使用相同的 key

```go
// 好：使用稳定的 ID
for _, user := range users {
    ui.ComponentWithProps("User", User).
        Key(fmt.Sprintf("user-%s", user.ID)). // 稳定的 key
        Build()
}

// 不好：使用索引（当列表重排时会出问题）
for i, user := range users {
    ui.ComponentWithProps("User", User).
        Key(fmt.Sprintf("user-%d", i)). // 不稳定的 key
        Build()
}
```

### 3. 状态管理

- 将相关状态分组到一个对象中
- 使用 reducer 模式处理复杂状态逻辑
- 避免派生状态（从 props 计算出的值不应该存为 state）

```go
// 不好的做法：冗余状态
func Bad() ui.VNode {
    firstName, _ := ui.UseStateString("")
    lastName, _ := ui.UseStateString("")
    fullName, _ := ui.UseStateString("") // 派生状态！

    // 每次 firstName/lastName 变化都要更新 fullName
    return ui.Text(fullName)
}

// 好的做法：按需计算
func Good() ui.VNode {
    firstName, _ := ui.UseStateString("")
    lastName, _ := ui.UseStateString("")

    fullName := fmt.Sprintf("%s %s", firstName, lastName)
    return ui.Text(fullName)
}
```

### 4. 性能优化

```go
func OptimizedList(items []Item) ui.VNode {
    // 使用 useMemo 缓存计算结果
    sortedItems := ui.UseMemo(func() interface{} {
        return sortItems(items)
    }, []interface{}{items})

    // 使用 useCallback 缓存回调
    handleClick := ui.UseCallback(func() {
        fmt.Println("Clicked")
    }, nil) // 空依赖 = 永不重新创建

    children := make([]ui.VNode, len(sortedItems.([]Item)))
    for i, item := range sortedItems.([]Item) {
        children[i] = ui.ComponentWithProps("Item", Item).
            Prop("item", item).
            Prop("onClick", handleClick). // 使用缓存的回调
            Key(fmt.Sprintf("item-%d", item.ID)).
            Build()
    }

    return ui.Column(children...)
}
```

## 内存安全

正确管理资源对于防止内存泄露至关重要。以下是 mint 提供的内存安全工具。

### Goroutine 管理

在 `useEffect` 中创建 goroutine 时，必须确保它们能在组件卸载时正确停止。

❌ **危险做法**：

```go
func LeakyTimer() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    ui.UseEffect(func() ui.CleanupFunc {
        ticker := time.NewTicker(time.Second)
        go func() {
            for range ticker.C {
                setCount(count + 1) // count 是闭包捕获的旧值
            }
        }()
        return func() {
            ticker.Stop() // ticker 停止了，但 goroutine 还在运行！
        }
    }, nil)

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

✅ **安全做法**：

```go
func SafeTimer() ui.VNode {
    count, setCount, getCount := ui.UseStateInt(0)
    gr := ui.UseGoRoutine()

    ui.UseEffect(func() ui.CleanupFunc {
        gr.Go(func(ctx <-chan struct{}) {
            ticker := time.NewTicker(time.Second)
            defer ticker.Stop()

            for {
                select {
                case <-ticker.C:
                    // 使用 getter 获取最新值
                    current := getCount()
                    setCount(current + 1)
                case <-ctx:
                    return // 收到停止信号，安全退出
                }
            }
        })

        return func() { gr.Stop() } // 通知 goroutine 停止
    }, nil)

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

### 订阅管理

使用 `UseSubscription` 管理外部订阅，确保组件卸载时自动取消。

```go
func MessageComponent() ui.VNode {
    messages, setMessages := ui.UseStateString("")

    ui.UseSubscription(func() *ui.Subscription {
        return messageHub.Subscribe(func(msg string) {
            setMessages(msg)
        })
    })

    return ui.Text(messages)
}
```

### 定时器管理

使用 `SafeTimer` 和 `SafeTicker` 避免定时器泄露。

```go
func PollingComponent() ui.VNode {
    data, setData := ui.UseStateString("")
    ticker := ui.NewSafeTicker(5 * time.Second)

    ui.UseEffect(func() ui.CleanupFunc {
        go func() {
            for range ticker.Channel() {
                // 轮询数据
                resp, _ := http.Get("https://api.example.com/data")
                setData(resp.Body)
            }
        }()

        return func() { ticker.Stop() }
    }, nil)

    return ui.Text(data)
}
```

### 并发限制

使用 `ResourcePool` 限制并发操作数量。

```go
func BatchFetch() ui.VNode {
    pool := ui.NewResourcePool(10) // 最多10个并发请求

    ui.UseEffect(func() ui.CleanupFunc {
        urls := []string{"url1", "url2", "url3"}

        for _, url := range urls {
            pool.Go(func() {
                fetch(url)
            })
        }

        return func() { pool.Close() }
    }, nil)

    return ui.Text("Fetching...")
}
```

### 内存安全检查清单

- [ ] useEffect 中的 goroutine 使用 `UseGoRoutine()`
- [ ] 外部订阅使用 `UseSubscription()`
- [ ] 定时器使用 `SafeTimer`/`SafeTicker`
- [ ] 闭包捕获变量使用正确方式（通过参数或 getter）
- [ ] 状态更新使用 getter 获取最新值
- [ ] 组件卸载时清理所有资源

详细 API 请参考 [内存安全工具 API](../api/memory-safety.md)。

## 示例

### 完整的计数器组件

```go
package main

import (
    "fmt"

    "github.com/wwsheng009/mint/ui"
)

func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)
    isHovered, setHovered := ui.UseHoverState()

    // 样式根据状态变化
    var style ui.StyleBuilder
    if isHovered() {
        style = ui.NewStyle().Background("cyan").Foreground("black")
    } else {
        style = ui.NewStyle().Background("blue").Foreground("white")
    }

    return ui.Row(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Text(" "),
        ui.Button("[-]").
            Style(style.Build()).
            OnClick(func() {
                setCount(count - 1)
            }).
            OnMouseEnter(func() { setHovered(true) }).
            OnMouseLeave(func() { setHovered(false) }),
        ui.Text(" "),
        ui.Button("[+]").
            Style(style.Build()).
            OnClick(func() {
                setCount(count + 1)
            }).
            OnMouseEnter(func() { setHovered(true) }).
            OnMouseLeave(func() { setHovered(false) }),
    )
}

func main() {
    ui.Run(func() ui.VNode {
        return ui.Column(
            ui.Text("Counter Example").Style(
                ui.NewStyle().Bold(true).Foreground("cyan"),
            ),
            ui.Text(""),
            Counter(),
        )
    })
}
```

### 带数据获取的列表组件

```go
func UserList() ui.VNode {
    users, setUsers := ui.UseStateInt(0) // 实际应用中可以使用自定义类型
    loading, setLoading := ui.UseStateBool(false)

    // 获取数据
    ui.UseEffect(func() func() {
        setLoading(true)

        go func() {
            // 模拟 API 调用
            time.Sleep(time.Second)
            // setUsers(newUsers)
            setLoading(false)
        }()

        return nil // 无清理
    }, nil)

    if loading {
        return ui.Text("Loading users...")
    }

    // 渲染列表
    children := make([]ui.VNode, users)
    for i := 0; i < users; i++ {
        children[i] = ui.Text(fmt.Sprintf("User %d", i+1))
    }

    return ui.Column(children...)
}
```

### 自定义 Hook

创建可重用的状态逻辑：

```go
// useCounter 自定义 hook
func useCounter(initialValue int) (int, func(int), func(), func()) {
    count, setCount := ui.UseStateInt(initialValue)

    increment := func() {
        setCount(count + 1)
    }

    decrement := func() {
        setCount(count - 1)
    }

    reset := func() {
        setCount(initialValue)
    }

    return count, increment, decrement, reset
}

// 使用自定义 hook
func CounterWithCustomHook() ui.VNode {
    count, increment, decrement, reset := useCounter(0)

    return ui.Row(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("[-]").OnClick(decrement),
        ui.Button("[+]").OnClick(increment),
        ui.Button("[Reset]").OnClick(reset),
    )
}
```

## 参考资料

- [Hooks API 参考](../api/hooks.md)
- [组件规范](../api/component.md)
- [内存安全工具 API](../api/memory-safety.md)
- [迁移指南](./migration-guide.md)
- [内存泄露分析](../issue/2026-02-01-memory-leak-analysis.md)
