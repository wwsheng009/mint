# Fiber Counter 修复说明

## 问题

参考 `mvp_components_demo` 与 `fiber_counter` 的实现差异，发现以下问题：

1. **StateKey 类型不匹配**
   - 原: `intent.StateKey[func(int)]("countSetter")`
   - 修复: `intent.StateKey[func(interface{})]("countSetter")`
   - 原因: `UseStateInt` 返回的 setter 签名是 `func(interface{})`

2. **不必要的类型断言**
   - 原代码尝试直接类型断言并调用
   - 修复: 使用 `callSetter` 统一处理

3. **调试输出过多**
   - 移除了大量的 printf 调试，改用 log.UILogger

## 修复内容

### 1. internal/reconciler/begin_work.go

让所有组件实例共享 `reconciler.ctx.GlobalState`，确保 Intent Handler 能访问组件保存的 setter。

```go
// Share the GlobalState map and its mutex from root context
instanceCtx.GlobalState = sharedCtx.GlobalState
instanceCtx.StateMu = sharedCtx.StateMu
```

### 2. ui/test.go

添加 Intent Runtime 初始化到 `RunTest` 和 `RunTestWithSandbox`，防止测试场景下 Intent Runtime 为 nil。

```go
// Initialize Intent Runtime (required for event handling)
intentRuntime := intent.NewRuntime()
intent.SetupBuiltinHandlers(intentRuntime)
rtui.SetGlobalIntentRuntime(intentRuntime)

// Call initialization function if provided
if options.InitFunc != nil {
    options.InitFunc()
}
```

### 3. examples/fiber_counter/main.go

完全重写，参考 `mvp_components_demo` 的正确模式：

```go
// StateKey 使用正确的类型
var countSetterKey = intent.StateKey[func(interface{})]("countSetter")

// 使用 callSetter 统一处理
ui.RegisterIntent(func(ctx *intent.ActionContext, i IncrementIntent) intent.IntentResult {
    setter, _ := ctx.GetState(countSetterKey.String())
    callSetter(setter, func(c int) int { return c + 1 })
    return intent.HandledResult()
})
```

## 测试建议

1. 手动测试:
   ```bash
   set MINT_USE_FIBER=true
   go run examples/fiber_counter/main.go
   ```

2. 按空格键移动焦点，按 Enter 点击按钮

3. 观察计数器是否正常递增
