# 内存安全工具 API 参考

本文档介绍了 mint UI 框架中用于防止内存泄露的工具类和函数。

## 目录

- [Goroutine 管理](#goroutine-管理)
- [订阅管理](#订阅管理)
- [定时器管理](#定时器管理)
- [资源池](#资源池)
- [检测工具](#检测工具)
- [最佳实践](#最佳实践)

---

## Goroutine 管理

### UseGoRoutine

创建一个受管理的 goroutine，在组件卸载时自动停止。

```go
func UseGoRoutine() *GoRoutine
```

**返回值**: `*GoRoutine` - 受管理的 goroutine 控制器

**示例**:

```go
func TimerComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)
    gr := ui.UseGoRoutine()

    ui.UseEffect(func() ui.CleanupFunc {
        // 启动 goroutine
        gr.Go(func(ctx <-chan struct{}) {
            ticker := time.NewTicker(time.Second)
            defer ticker.Stop()

            for {
                select {
                case <-ticker.C:
                    setCount(func(c int) int { return c + 1 })
                case <-ctx:
                    return // 收到停止信号，安全退出
                }
            }
        })

        // 清理函数：停止 goroutine
        return func() { gr.Stop() }
    }, nil)

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

### GoRoutine 方法

#### Go(channel) <-chan struct{}

返回 done channel，用于监听停止信号。

```go
done := gr.Done()
<-done  // 阻塞直到 goroutine 被停止
```

#### Stop()

停止 goroutine。可多次调用（幂等）。

```go
gr.Stop()
```

#### Go(func(<-chan struct{}))

启动一个 goroutine，接收 done channel 作为参数。

```go
gr.Go(func(ctx <-chan struct{}) {
    for {
        select {
        case <-ctx:
            return
        case <-time.After(time.Second):
            // 做一些工作
        }
    }
})
```

---

## 订阅管理

### UseSubscription

管理一个外部订阅（如事件流、数据源等），组件卸载时自动取消订阅。

```go
func UseSubscription(createSub func() *Subscription) *Subscription
```

**参数**:
- `createSub`: 创建订阅的函数

**返回值**: `*Subscription` - 订阅对象

**示例**:

```go
func MessageComponent() ui.VNode {
    messages, setMessages := ui.UseStateString("")

    sub := ui.UseSubscription(func() *ui.Subscription {
        return messageSource.Subscribe(func(msg string) {
            setMessages(msg)
        })
    })

    // sub 会在组件卸载时自动 Unsubscribe
    return ui.Text(messages)
}
```

### Subscription 方法

#### Unsubscribe()

取消订阅。可多次调用（幂等）。

```go
sub.Unsubscribe()
```

#### Done() <-chan struct{}

返回一个 channel，当订阅取消时关闭。

```go
select {
case <-sub.Done():
    fmt.Println("Subscription cancelled")
case <-time.After(time.Second):
    fmt.Println("Timeout")
}
```

---

## 定时器管理

### SafeTimer

一个安全的定时器包装器，防止 goroutine 泄露。

```go
func NewSafeTimer(d time.Duration, fn func()) *SafeTimer
```

**参数**:
- `d`: 定时器持续时间
- `fn`: 定时器触发时执行的函数

**方法**:

| 方法 | 说明 |
|------|------|
| `Start()` | 启动定时器 |
| `Reset(d)` | 重置定时器为新的持续时间 |
| `Stop()` | 停止定时器 |

**示例**:

```go
func DebouncedInput() ui.VNode {
    value, setValue := ui.UseStateString("")

    st := ui.NewSafeTimer(500*time.Millisecond, func() {
        // 防抖：用户停止输入 500ms 后执行
        fmt.Printf("Final value: %s\n", value)
    })

    ui.UseEffect(func() ui.CleanupFunc {
        st.Start()
        return func() { st.Stop() }
    }, nil)

    return ui.Input("Type...", value, 20).OnChange(func(newValue string) {
        setValue(newValue)
        st.Reset(500 * time.Millisecond) // 重置定时器
    })
}
```

### SafeTicker

一个安全的 Ticker 包装器。

```go
func NewSafeTicker(d time.Duration) *SafeTicker
```

**方法**:

| 方法 | 说明 |
|------|------|
| `Channel() <-chan time.Time` | 返回 ticker channel |
| `Stop()` | 停止 ticker |
| `Done() <-chan struct{}` | 返回停止通知 channel |

**示例**:

```go
func ClockComponent() ui.VNode {
    st := ui.NewSafeTicker(time.Second)

    ui.UseEffect(func() ui.CleanupFunc {
        return func() { st.Stop() }
    }, nil)

    return ui.Text("Tick...").OnRender(func() {
        select {
        case <-st.Channel():
            // 每秒触发一次
        case <-st.Done():
            // 组件卸载
        }
    })
}
```

---

## 资源池

### ResourcePool

限制并发资源数量的资源池。

```go
func NewResourcePool(maxSize int) *ResourcePool
```

**方法**:

| 方法 | 说明 |
|------|------|
| `Acquire() bool` | 获取一个资源，成功返回 true |
| `Release()` | 释放资源回池 |
| `Go(fn func()) bool` | 在池中运行函数 |
| `Close()` | 关闭资源池 |

**示例**:

```go
func ConcurrentFetch() ui.VNode {
    pool := ui.NewResourcePool(5) // 最多5个并发

    ui.UseEffect(func() ui.CleanupFunc {
        urls := []string{"url1", "url2", "url3"}

        for _, url := range urls {
            pool.Go(func() {
                resp := fetch(url)
                process(resp)
            })
        }

        return func() { pool.Close() }
    }, nil)

    return ui.Text("Fetching...")
}
```

---

## 检测工具

### GoroutineTracker

追踪 goroutine 数量变化，检测泄露。

```go
func NewGoroutineTracker(threshold int) *GoroutineTracker
```

**方法**:

| 方法 | 说明 |
|------|------|
| `Update()` | 更新当前 goroutine 计数 |
| `CheckForLeaks() error` | 检测是否超过阈值 |
| `Count() int` | 获取当前 goroutine 数 |

**示例**:

```go
func TestComponent_NoLeaks(t *testing.T) {
    tracker := ui.NewGoroutineTracker(2) // 允许2个新goroutine

    // 运行组件
    ui.Run(func() ui.VNode {
        return MyComponent()
    })

    // 检查泄露
    tracker.Update()
    if err := tracker.CheckForLeaks(); err != nil {
        t.Errorf("Goroutine leak detected: %v", err)
    }
}
```

### MemStats

追踪内存分配统计。

```go
func NewMemStats() *MemStats
```

**方法**:

| 方法 | 说明 |
|------|------|
| `CheckAlloc() uint64` | 检查自上次以来的内存分配增量 |
| `TotalAlloc() uint64` | 获取总分配量 |
| `Reset()` | 重置统计 |

**示例**:

```go
func TestMemoryUsage(t *testing.T) {
    stats := ui.NewMemStats()

    // 运行测试
    runComponent()

    delta := stats.CheckAlloc()
    t.Logf("Allocated %d bytes during test", delta)
}
```

---

## 最佳实践

### 1. useEffect 中创建 Goroutine

❌ **错误做法**:

```go
ui.UseEffect(func() ui.CleanupFunc {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            setCount(count + 1)
        }
    }()
    return func() { ticker.Stop() }  // ticker 停止了，goroutine 还在运行！
}, nil)
```

✅ **正确做法**:

```go
gr := ui.UseGoRoutine()

ui.UseEffect(func() ui.CleanupFunc {
    gr.Go(func(ctx <-chan struct{}) {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                setCount(count + 1)
            case <-ctx:
                return // 安全退出
            }
        }
    })

    return func() { gr.Stop() }
}, nil)
```

### 2. 事件订阅

❌ **错误做法**:

```go
func MyComponent() ui.VNode {
    // 订阅但没有取消
    dataSource.Subscribe(func(msg string) {
        setData(msg)
    })
    return ui.Text(data)
}
```

✅ **正确做法**:

```go
func MyComponent() ui.VNode {
    data, setData := ui.UseStateString("")

    ui.UseSubscription(func() *ui.Subscription {
        return dataSource.Subscribe(func(msg string) {
            setData(msg)
        })
    })

    return ui.Text(data)
}
```

### 3. 定时器使用

❌ **错误做法**:

```go
func MyComponent() ui.VNode {
    ticker := time.NewTicker(time.Second)

    ui.UseEffect(func() ui.CleanupFunc {
        return func() { ticker.Stop() }
    }, nil)

    // ticker 在 useEffect 外部创建，可能泄露
    return ui.Text("...")
}
```

✅ **正确做法**:

```go
func MyComponent() ui.VNode {
    ticker := ui.NewSafeTicker(time.Second)

    ui.UseEffect(func() ui.CleanupFunc {
        // 使用 ticker
        go func() {
            for range ticker.Channel() {
                // 处理 tick
            }
        }()

        return func() { ticker.Stop() }
    }, nil)

    return ui.Text("...")
}
```

### 4. 防止闭包捕获

❌ **问题代码**:

```go
for i := 0; i < 5; i++ {
    gr.Go(func(ctx <-chan struct{}) {
        fmt.Println(i) // 可能打印错误的值
    })
}
```

✅ **正确做法**:

```go
for i := 0; i < 5; i++ {
    i := i // 捕获循环变量
    gr.Go(func(ctx <-chan struct{}) {
        fmt.Println(i)
    })
}
```

### 5. 状态更新的闭包问题

❌ **问题代码**:

```go
count, setCount := ui.UseStateInt(0)

gr.Go(func(ctx <-chan struct{}) {
    for range time.NewTicker(time.Second).C {
        setCount(count + 1) // count 是旧值！
    }
})
```

✅ **正确做法**:

```go
count, setCount, getCount := ui.UseStateInt(0)

gr.Go(func(ctx <-chan struct{}) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            current := getCount() // 获取最新值
            setCount(current + 1)
        case <-ctx:
            return
        }
    }
})
```

---

## API 快速参考

| 工具 | 用途 | 生命周期 |
|------|------|---------|
| `UseGoRoutine()` | 管理 goroutine | 组件卸载 |
| `UseSubscription()` | 管理订阅 | 组件卸载 |
| `NewSafeTimer()` | 安全定时器 | 手动 Stop |
| `NewSafeTicker()` | 安全 Ticker | 手动 Stop |
| `NewResourcePool()` | 并发限制 | 手动 Close |
| `NewGoroutineTracker()` | 泄露检测 | 测试 |
| `NewMemStats()` | 内存统计 | 测试 |
