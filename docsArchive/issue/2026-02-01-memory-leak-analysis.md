# 内存泄露风险分析

本文档分析 mint 声明式 UI 框架中潜在的内存泄露风险。

## 风险总结

| 风险类别 | 严重性 | 状态 |
|---------|--------|------|
| Goroutine 泄露 | 🔴 高 | 需要用户正确处理 |
| 闭包循环引用 | 🟡 中 | 已有保护机制 |
| Props 闭包引用 | 🟡 中 | 需要注意 |
| Hooks 数组永不收缩 | 🟢 低 | 设计特性 |
| 实例无上限 | 🟢 低 | 已有 LRU 限制 |

## 详细分析

### 1. Goroutine 泄露 🔴 高风险

**问题**: useEffect 中创建的 goroutine 如果不正确停止会导致泄露

```go
// 危险示例：goroutine 泄露
func BadComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    ui.UseEffect(func() func() {
        ticker := time.NewTicker(time.Second)
        go func() {
            for range ticker.C {
                // 问题：如果组件卸载，goroutine 仍在运行
                setCount(count + 1)
            }
        }()
        return func() { ticker.Stop() } // ticker 停止了，但 goroutine 可能阻塞
    }, nil)

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

**修复方案**:

```go
// 正确示例：使用 done 通道
func GoodComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)
    doneRef := ui.UseRef(make(chan struct{}))

    ui.UseEffect(func() func() {
        // 获取新的 done 通道
        done := make(chan struct{})
        doneRef.Value = done

        ticker := time.NewTicker(time.Second)
        go func() {
            for {
                select {
                case <-ticker.C:
                    setCount(func(c int) int { return c + 1 })
                case <-done:
                    ticker.Stop()
                    return // 退出 goroutine
                }
            }
        }()
        return func() {
            close(done) // 通知 goroutine 退出
        }
    }, nil)

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

### 2. 闭包循环引用 🟡 中风险

**问题**: 闭包捕获外部变量可能导致循环引用

```go
// 潜在问题：闭包捕获了组件上下文
func ClosureComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    // 这个闭包捕获了 count 和 setCount
    // setCount 内部可能持有组件上下文的引用
    handler := func() {
        setCount(count + 1)
    }

    // 如果 handler 被存储到全局或长时间存活的对象中
    // 组件实例将无法被 GC 回收
    return ui.Button("Click").OnClick(handler)
}
```

**保护机制**:

- 组件实例通过 key 匹配，不是通过引用
- 每次渲染创建新的闭包（除非使用 useCallback）
- 实例管理器有 LRU 上限

**建议**:

```go
// 使用 useCallback 缓存回调，但要小心依赖
func OptimizedComponent() ui.VNode {
    count, setCount, getCount := ui.UseStateInt(0)

    // 使用 getter 而不是直接捕获 count
    handler := ui.UseCallback(func() {
        currentCount := getCount() // 获取最新值
        setCount(currentCount + 1)
    }, nil) // 空依赖 = 永不重新创建

    return ui.Button("Click").OnClick(handler)
}
```

### 3. Props 闭包引用 🟡 中风险

**问题**: Props 传递闭包时可能引用父组件的状态

```go
func ParentComponent() ui.VNode {
    data, setData := ui.UseStateString("")

    // 这个闭包捕获了 data 的引用
    return ui.ComponentWithProps("Child", ChildComponent).
        Prop("onChange", func(newValue string) {
            data = newValue // 问题：直接修改，不会触发重新渲染
            // 正确做法：使用 setData
        }).
        Prop("data", data).
        Build()
}
```

### 4. Hooks 数组永不收缩 🟢 低风险

**问题**: ComponentContext.Hooks 数组只会增长，不会收缩

```go
type ComponentContext struct {
    ComponentID string
    Hooks       []Hook  // 只会增长，永不收缩
    HookIndex   int
    Validator   *HookValidator
    RenderCount int
}
```

**影响**: 对于固定数量的 hooks，这是设计特性，不是 bug。但如果条件性地使用 hooks 会导致 panic（已被 HookValidator 保护）。

### 5. 实例无上限 🟢 低风险

**问题**: 如果没有设置 maxInstances，组件实例会无限增长

**保护机制**:

```go
// InstanceManager 默认限制
func NewInstanceManager() *InstanceManager {
    return &InstanceManager{
        instances:     make(map[string]ComponentInstance),
        instanceOrder: make([]string, 0),
        lastAccess:    make(map[string]time.Time),
        maxInstances:  1000, // 默认上限 1000
    }
}

// cleanupLRU 自动清理最久未使用的实例
func (m *InstanceManager) cleanupLRU() {
    for len(m.instances) > m.maxInstances && len(m.instanceOrder) > 0 {
        oldestKey := m.instanceOrder[0]
        inst := m.instances[oldestKey]
        inst.OnUnmount() // 调用清理
        delete(m.instances, oldestKey)
        delete(m.lastAccess, oldestKey)
        m.instanceOrder = m.instanceOrder[1:]
    }
}
```

## 已有的保护机制

### 1. OnUnmount 清理

```go
// BaseComponentInstance.OnUnmount
func (b *BaseComponentInstance) OnUnmount() {
    b.context.cleanupAll() // 清理所有 effect 的 cleanup 函数
    b.mounted = false
}
```

### 2. Cleanup 调用链

```
InstanceManager.Cleanup()
    → inst.OnUnmount()
        → context.cleanupAll()
            → hook.Cleanup()  // 每个 effect 的清理函数
```

### 3. LRU 淘汰

```go
// InstanceManager 自动清理超过上限的实例
func (m *InstanceManager) GetOrCreate(key string, creator func() ComponentInstance) ComponentInstance {
    // ...
    m.cleanupLRU() // 在创建新实例后自动检查
    return inst
}
```

## 建议的改进

### 1. 添加 context done 通道辅助

```go
// 添加到 hooks.go
func UseDoneChannel() <-chan struct{} {
    ctx := getCurrentContext()
    if ctx == nil {
        panic("UseDoneChannel must be called within a component")
    }

    ref := UseRef(make(chan struct{}))

    // useEffect 确保卸载时关闭通道
    UseEffect(func() func() {
        done := make(chan struct{})
        ref.Value = done
        return func() { close(done) }
    }, nil)

    return ref.Value.(<-chan struct{})
}
```

### 2. 添加 Goroutine 辅助函数

```go
// Go 是一个在组件卸载时自动取消的 goroutine 启动器
func Go(f func(<-chan struct{})) func() {
    ctx := getCurrentContext()
    doneChan := make(chan struct{})

    go func() {
        <-ctx.done // 等待组件卸载信号
        close(doneChan)
    }()

    go func() {
        f(doneChan)
    }()

    return func() {
        close(ctx.done)
    }
}
```

### 3. 检测泄露的测试

```go
func TestNoGoroutineLeak(t *testing.T) {
    initial := runtime.NumGoroutine()

    // 运行组件
    ui.Run(func() ui.VNode {
        return TestComponent()
    })

    // 等待清理
    time.Sleep(100 * time.Millisecond)

    final := runtime.NumGoroutine()
    if final > initial+2 { // 允许少量误差
        t.Errorf("possible goroutine leak: %d -> %d", initial, final)
    }
}
```

## 最佳实践

### 1. useEffect 中创建 goroutine 的模式

```go
func SafeTimer() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    ui.UseEffect(func() func() {
        done := make(chan struct{})
        ticker := time.NewTicker(time.Second)

        go func() {
            for {
                select {
                case <-ticker.C:
                    setCount(func(c int) int { return c + 1 })
                case <-done:
                    ticker.Stop()
                    return
                }
            }
        }()

        return func() { close(done) }
    }, nil)

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

### 2. 避免在 Props 中传递闭包

```go
// 不好：每次创建新闭包
func BadParent() ui.VNode {
    return ui.ComponentWithProps("Child", Child).
        Prop("onClick", func() { /* ... */ }). // 每次都是新函数
        Build()
}

// 好：使用 useCallback 或稳定的函数
func GoodParent() ui.VNode {
    handler := ui.UseCallback(func() { /* ... */ }, nil)
    return ui.ComponentWithProps("Child", Child).
        Prop("onClick", handler).
        Build()
}
```

### 3. 及时取消订阅

```go
func SubscriptionComponent() ui.VNode {
    data, setData := ui.UseStateString("")

    ui.UseEffect(func() func() {
        sub := subscribeToEvents(func(msg string) {
            setData(msg)
        })

        return func() {
            sub.Unsubscribe() // 确保取消订阅
        }
    }, nil)

    return ui.Text(data)
}
```

## 总结

| 风险 | 缓解措施 |
|------|---------|
| Goroutine 泄露 | 使用 done 通道模式 |
| 闭包引用 | 使用 getter/getter 模式 |
| 实例过多 | LRU 自动清理 (maxInstances=1000) |
| Effect 清理 | OnUnmount → cleanupAll() |

框架已提供基础保护机制，但用户需要注意：
1. 在 useEffect 中创建 goroutine 时使用 done 模式
2. 避免在闭包中直接捕获状态
3. 为长期运行的订阅提供清理函数
