# Hooks API 参考

本文档介绍 mint UI 框架中所有 Hooks 的 API。

## 目录

- [状态 Hooks](#状态-hooks)
- [Effect Hooks](#effect-hooks)
- [Ref Hooks](#ref-hooks)
- [性能优化 Hooks](#性能优化-hooks)
- [悬停状态 Hooks](#悬停状态-hooks)
- [Hook 验证](#hook-验证)

---

## 状态 Hooks

### UseStateInt

创建一个整型状态。

```go
func UseStateInt(initial int) (current int, setter func(interface{}), getter func() int)
```

**参数**:
- `initial`: 初始值

**返回值**:
- `current`: 当前状态值
- `setter`: 设置状态的函数，接受 `int` 或 `func(int) int`
- `getter`: 获取当前状态值的函数

**示例**:

```go
func Counter() ui.VNode {
    count, setCount, getCount := ui.UseStateInt(0)

    return ui.Row(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("[+]").OnClick(func() {
            setCount(getCount() + 1) // 使用 getter 获取最新值
        }),
    )
}
```

### UseStateString

创建一个字符串状态。

```go
func UseStateString(initial string) (current string, setter func(string))
```

### UseStateBool

创建一个布尔状态。

```go
func UseStateBool(initial bool) (current bool, setter func(bool))
```

---

## Effect Hooks

### UseEffect

执行副作用，可选择返回清理函数。

```go
func UseEffect(callback EffectCallback, deps []interface{})
```

**类型**:

```go
type EffectCallback func() CleanupFunc
type CleanupFunc func()
```

**依赖参数说明**:
- `nil` - 只在组件挂载时运行一次
- `[]` - 每次渲染后都运行
- `[values]` - 仅当依赖值变化时运行

**示例**:

```go
// 只运行一次
ui.UseEffect(func() ui.CleanupFunc {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            // 处理
        }
    }()
    return func() { ticker.Stop() } // 清理
}, nil)

// 依赖变化时运行
ui.UseEffect(func() ui.CleanupFunc {
    fetchData(userID)
    return nil
}, []interface{}{userID})
```

---

## Ref Hooks

### UseRef

创建一个在渲染之间保持不变的引用。

```go
func UseRef(initial interface{}) *Ref
```

**类型**:

```go
type Ref struct {
    Value interface{}
}
```

**示例**:

```go
func TimerComponent() ui.VNode {
    tickCount := ui.UseRef(0)

    ui.UseEffect(func() ui.CleanupFunc {
        ticker := time.NewTicker(time.Second)
        go func() {
            for range ticker.C {
                tickCount.Value = tickCount.Value.(int) + 1
            }
        }()
        return func() { ticker.Stop() }
    }, nil)

    return ui.Text(fmt.Sprintf("Ticks: %v", tickCount.Value))
}
```

---

## 性能优化 Hooks

### UseMemo

缓存计算结果，仅在依赖变化时重新计算。

```go
func UseMemo(compute func() interface{}, deps []interface{}) interface{}
```

**示例**:

```go
func ExpensiveComponent(a, b int) ui.VNode {
    result := ui.UseMemo(func() interface{} {
        return expensiveComputation(a, b)
    }, []interface{}{a, b})

    return ui.Text(fmt.Sprintf("Result: %v", result))
}
```

### UseCallback

缓存回调函数，仅在依赖变化时重新创建。

```go
func UseCallback(callback func(), deps []interface{}) func()
```

**示例**:

```go
func ButtonComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    handleClick := ui.UseCallback(func() {
        setCount(count + 1)
    }, []interface{}{count}) // count 变化时创建新回调

    return ui.Button("Click").OnClick(handleClick)
}
```

---

## 悬停状态 Hooks

### UseHoverState

创建一个跨渲染持久化的悬停状态。

```go
func UseHoverState() (getter func() bool, setter func(bool))
```

**返回值**:
- `getter`: 返回当前悬停状态的函数
- `setter`: 设置悬停状态的函数

**示例**:

```go
func HoverButton() ui.VNode {
    isHovered, setHovered := ui.UseHoverState()

    var style ui.StyleBuilder
    if isHovered() {
        style = ui.NewStyle().Background("cyan").Foreground("black")
    } else {
        style = ui.NewStyle().Background("blue").Foreground("white")
    }

    return ui.Button("Hover me").
        Style(style.Build()).
        OnMouseEnter(func() { setHovered(true) }).
        OnMouseLeave(func() { setHovered(false) })
}
```

---

## Hook 验证

框架会自动验证 Hooks 的调用顺序和条件。违反规则会导致 panic。

### 验证规则

1. **只能在组件顶层调用 Hooks**
   ```go
   // ✅ 正确
   func Good() ui.VNode {
       count := ui.UseStateInt(0)
       name := ui.UseStateString("")
       return ui.Text(name)
   }

   // ❌ 错误
   func Bad() ui.VNode {
       count := ui.UseStateInt(0)
       if count > 0 {
           name := ui.UseStateString("") // 在条件中调用
       }
       return ui.Text("")
   }
   ```

2. **每次渲染 Hooks 顺序必须一致**
   ```go
   // 第一次渲染
   count := ui.UseStateInt(0)
   name := ui.UseStateString("")

   // 第二次渲染必须是相同顺序
   count := ui.UseStateInt(0)
   name := ui.UseStateString("")
   ```

3. **必须在组件内调用**
   ```go
   // ❌ 错误
   count := ui.UseStateInt(0) // 不在组件内，会 panic
   ```

### Hook 类型

| HookType | 常量 | Hook 函数 |
|----------|------|---------|
| State | `HookState` | `UseStateInt`, `UseStateString`, `UseStateBool` |
| Effect | `HookEffect` | `UseEffect` |
| Context | `HookContext` | `UseContext` (未实现) |
| Memo | `HookMemo` | `UseMemo`, `UseCallback` |
| Ref | `HookRef` | `UseRef` |

---

## 完整 API 列表

| 函数 | 签名 | 描述 |
|------|------|------|
| `UseStateInt` | `func(int) (int, func(interface{}), func() int)` | 整型状态 |
| `UseStateString` | `func(string) (string, func(string))` | 字符串状态 |
| `UseStateBool` | `func(bool) (bool, func(bool))` | 布尔状态 |
| `UseEffect` | `func(EffectCallback, []interface{})` | 副作用 |
| `UseRef` | `func(interface{}) *Ref` | 持久化引用 |
| `UseMemo` | `func(func() interface{}, []interface{}) interface{}` | 缓存值 |
| `UseCallback` | `func(func(), []interface{}) func()` | 缓存函数 |
| `UseHoverState` | `func() (func() bool, func(bool))` | 悬停状态 |

---

## 相关文档

- [组件开发指南](../guide/component-development-guide.md)
- [内存安全工具](./memory-safety.md)
- [组件状态持久化](/docsArchive/issue/2026-02-01-component-state-persistence.md)
