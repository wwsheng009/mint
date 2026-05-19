# Count 只能增加一次的问题修复

## 问题描述

`examples/absolute/main.go` 中的计数器点击按钮后只能增加一次（0 → 1），之后点击不再增加。

## 根本原因

### 原因 1：闭包捕获旧值

在原来的代码中：

```go
count, setCount, _ := ui.UseStateInt(0)

ui.On(IncrementIntent{}, func() {
    setCount(count + 1)  // ❌ 问题：闭包捕获 count 的值
})
```

**执行流程**：

1. **第一次渲染**：
   ```go
   count = 0
   setCount = func(int) {}
   
   // 注册 handler（只在第一次注册）
   handler = func() {
       setCount(0 + 1)  // 闭包捕获 count = 0
   }
   ```

2. **第一次点击**：
   ```go
   handler() → setCount(0 + 1) → count = 1
   // 组件重新渲染
   ```

3. **第二次渲染**：
   ```go
   count = 1
   setCount = func(int) {}
   
   // ui.On 检测到 handler 已注册，跳过注册
   // handler 仍然是第一次的闭包，捕获的是 count = 0
   ```

4. **第二次点击**：
   ```go
   handler() → setCount(0 + 1) → count = 1  // 没有增加！
   ```

**问题核心**：
- `ui.On` 使用全局 `registeredHandlers` (sync.Map) 避免重复注册
- handler 只注册一次，闭包捕获的是**第一次渲染时的 count 值**
- 即使重新渲染 count 变了，handler 中的 `count` 仍然是旧值
- Go 的整型是值类型，闭包捕获值拷贝，不会更新

### 原因 2：Go闭包机制

```go
x := 1
f := func() int {
    return x + 1  // 捕获 x 的值
}
x = 10
println(f())  // 输出 2（不是 11），因为 f 捕获的是 x 的快照
```

## 解决方案：使用函数式更新

### 修复后的代码

```go
count, setCount, _ := ui.UseStateInt(0)

ui.On(IncrementIntent{}, func() {
    setCount(func(c int) int {  // ✅ 使用函数式更新
        return c + 1
    })
})
```

**为什么这样能工作**：

1. **第一次渲染**：
   ```go
   count = 0
   setCount = func(interface{}) {}
   
   func setCount(fn func(int) int) {
       current := getValue()  // 从 Hooks[hookIndex] 获取最新值
       setValue(fn(current))  // Apply function to current value
   }
   
   handler = func() {
       setCount(func(c int) int {
           return c + 1  // c 是参数，不是闭包捕获
       })
   }
   ```

2. **第一次点击**：
   ```go
   handler() → setCount(fn) → current=0 → fn(0)=1 → count = 1
   // 组件重新渲染
   ```

3. **第二次点击**：
   ```go
   handler() → setCount(fn) → current=1 → fn(1)=2 → count = 2  // ✅ 正确！
   ```

**关键点**：
- `setCount(func(c int) int { return c + 1 })` 中的 `c` 是**参数**
- 每次 handler 被调用时，`fn(c)` 中的 `c` 会从最新状态获取
- 没有闭包捕获问题

## 其他注意事项

### 1. UseStateInt 的返回值

```go
func UseStateInt(initial int) (int, func(interface{}), func() int) {
    // count       当前值
    // setCount    设置器（支持值或函数）
    // getValue    获取最新值（可通过闭包捕获保持更新）
    
    value, setValue := useState(initial)
    getValue := func() int {
        // 使用 hookIndex 获取最新值
        return ctx.Hooks[hookIndex].Value.(int)
    }
    
    setInt := func(newValue interface{}) {
        switch v := newValue.(type) {
        case int:
            setValue(v)
        case func(int) int:
            current := getValue()
            setValue(v(current))  // 函数式更新
        }
    }
    
    return value, setInt, getValue
}
```

### 2. ui.On 的去重机制

```go
func On[T ...](intentType T, handler func()) {
    key := intentType.IntentType()
    
    // 包级 map，跨组件共享
    if _, loaded := registeredHandlers.LoadOrStore(key, true); loaded {
        return  // 已注册，跳过
    }
    
    rtui.RegisterIntent(func(ctx *intent.ActionContext, i T) intent.IntentResult {
        handler()  // 调用闭包
        return intent.HandledResult()
    })
}
```

**影响**：
- handler 只注册一次
- 如果 handler 闭包捕获了组件内的局部变量，变量值不会更新
- **必须使用函数式更新**避免这个问题

### 3. 正确的使用示例

```go
// ❌ 错误：闭包捕获旧值
count, setCount, _ := ui.UseStateInt(0)
ui.On(IncrementIntent{}, func() {
    setCount(count + 1)  // count 是 0
})

// ✅ 正确：函数式更新
count, setCount, _ := ui.UseStateInt(0)
ui.On(IncrementIntent{}, func() {
    setCount(func(c int) int { return c + 1 })
})

// ✅ 也可以：使用 getValue（较少见）
count, setCount, getValue := ui.UseStateInt(0)
ui.On(IncrementIntent{}, func() {
    setCount(getValue() + 1)  // getValue() 总是返回最新值
})
```

## 总结

| 方式 | 特点 | 适用场景 |
|------|------|---------|
| `setCount(newVal)` | 直接设置值 | 用于外部计算的新值 |
| `setCount(func(int) int)` | 函数式更新 | **推荐：避免闭包捕获问题** |
| `getValue() + 1` | 手动获取最新值 | 需要同时使用 `getValue` 时 |

**最佳实践**：
- 与 `ui.On` 配合使用时，**总是使用函数式更新**
- 模式：`setCount(func(c int) int { return c + 1 })`
- 参考：`ui/memory_safety.go` 中的示例代码
