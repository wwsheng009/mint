# API 封装改进对比

## 旧 API（反直觉）

```go
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

func SimpleCounter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)
    countSetterKey := intent.StateKey[func(int)]("countSetter")

    // ❌ 需要手动保存 setter
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState[countSetterKey.String()] = setCount
    }

    return ui.VStack(...)
}

func main() {
    err := ui.Run(SimpleCounter,
        ui.WithInit(func() {
            // ❌ 需要在 WithInit 中手动注册处理器
            countSetterKey := intent.StateKey[func(int)]("countSetter")

            ui.RegisterIntent(func(ctx *intent.ActionContext, i IncrementIntent) intent.IntentResult {
                setter, _ := ctx.GetState(countSetterKey.String())
                callSetter(setter, func(c int) int { return c + 1 })  // ❌ 需要反射
                return intent.HandledResult()
            })
        }),
    )
}

// ❌ 还需要一个反射辅助函数
func callSetter(fn interface{}, arg interface{}) {
    if fn == nil { return }
    v := reflect.ValueOf(fn)
    if v.Kind() != reflect.Func { return }
    argV := reflect.ValueOf(arg)
    v.Call([]reflect.Value{argV})
}
```

**缺点：**
- ❌ 代码分散：setter 定义、保存、注册在三个不同地方
- ❌ 需要反射：`callSetter` 使用反射不够类型安全
- ❌ 冗余代码：`countSetterKey` 在多处重复定义
- ❌ 不直观：开发者需要理解 GlobalState 机制

---

## 新 API（简洁直观）

```go
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

// ✅ 简单的 On() 辅助函数
func On[T interface{ IntentType() string; StayPressed() bool }](intentType T, handler func()) {
    ctx := ui.GetCurrentContext()
    if ctx == nil { return }

    // 使用标签确保只注册一次
    registryKey := "__on_handler_registered_" + intentType.IntentType()
    if _, exists := ctx.GlobalState[registryKey]; exists {
        return
    }

    ctx.GlobalState[registryKey] = true
    ui.RegisterIntent(func(ctx *intent.ActionContext, i T) intent.IntentResult {
        handler()
        return intent.HandledResult()
    })
}

func SimpleCounter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    // ✅ 简洁：直接使用 On 注册处理器
    On(IncrementIntent{}, func() {
        setCount(count + 1)  // 闭包直接访问局部变量
    })

    On(DecrementIntent{}, func() {
        setCount(count - 1)
    })

    return ui.VStack(...)
}

// ✅ 不需要 WithInit
func main() {
    err := ui.Run(SimpleCounter,
        ui.WithWidth(40),
        ui.WithHeight(10),
        ui.WithTitle("Fiber Counter (Improved API)"),
    )
}
```

**优点：**
- ✅ 代码集中：状态和处理在同一个函数内
- ✅ 类型安全：闭包直接访问，无需反射
- ✅ 自动注册：`On()` 内部处理去重
- ✅ 直观易用：类似事件监听器的常见模式

---

## 代码行数对比

| 项目 | 旧 API | 新 API | 改进 |
|------|--------|--------|------|
| main 函数 | ~30 行 | ~15 行 | -50% |
| 组件函数 | ~15 行 | ~10 行 | -33% |
| 反射辅助 | ~10 行 | 0 行 | -100% |
| **总计** | ~55 行 | ~25 行 | **-55%** |

---

## 适用场景

### 旧 API 仍然适用的场景

当需要在多个组件间共享状态时：

```go
// 在 WithInit 中绑定全局 setter
ui.BindField(usernameKey, setUsername)  // 声明式绑定

// 在任何组件中都可以接收更新
func ComponentA() ui.VNode {
    return app.InputBuilder().ForField(usernameKey).Build()
}

func ComponentB() ui.VNode {
    username := ui.GetStringState(usernameKey.String(), "")
    return app.Text("Hello " + username)
}
```

### 新 API 最适合的场景

单一组件内状态的简单操作：

```go
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    On(IncrementIntent{}, func() { setCount(count + 1) })
    On(DecrementIntent{}, func() { setCount(count - 1) })

    return ...
}
```

---

## 建议

1. **新增示例优先使用新 API**：`On()` 简洁直观
2. **全局状态使用 `BindField` API**：跨组件共享
3. **两种 API 可以共存**：根据场景选择
