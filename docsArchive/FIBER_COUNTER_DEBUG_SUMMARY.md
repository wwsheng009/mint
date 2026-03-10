# Fiber Counter 调试总结

## 问题

fiber_counter 示例在点击几次后计数器停止更新。

## 已知情况

1. `mvp_components_demo` 和 `mvp_form_demo` 能通过 `ui.Run` 正常运行 ✅
2. 这说明 **MVP 模式本身工作正常**

## 测试框架状态

- `ui.RunTest` 函数支持自动化测试
- 通过 `InjectSpecialKey` 模拟用户输入
- **尚未验证测试框架是否与当前 MVP 模式正确集成** ⚠️

## 修复尝试

### 1. 添加调试输出到 fiber_counter
- SimpleCounter 中添加了 GlobalState 保存时的调试日志
- Intent handler 添加了类型检查和调试输出

### 2. 修复 GlobalState 共享问题
- 在 `internal/reconciler/begin_work.go` 中
- 让所有组件的实例共享 `reconciler.ctx.GlobalState`
- 这样 Intent Handler 就能访问组件保存的 setter

### 3. 修复 ui.RunTest 缺失的 Intent Runtime
- 在 `ui/run_test.go` 中添加了 Intent Runtime 初始化
- 在 `options.InitFunc()` 调用之前设置全局 Intent Runtime
- 调用 `render.SetDeclarativeNodeIntentRuntime` 连接 Intent Runtime

## 下一步

由于 `mvp_components_demo` 能工作，我们需要：

1. **手动测试 fiber_counter**
   - 编译并运行 `fiber_counter.exe`
   - 手动点击按钮看是否能正常工作
   - 观察调试输出

2. **确认测试框架问题**
   - 如果手动测试能工作，说明问题在测试框架
   - 如果手动测试不能工作，说明问题在代码本身

3. **对比可工作的示例**
   - 比较 `mvp_components_demo` 和 `fiber_counter` 的实现差异
   - 找出不同的地方

## 关键发现

| 组件 | 工作状态 | 说明 |
|------|---------|------|
| mvp_components_demo | ✅ 正常 | UseState, WithInit, ForField 模式 |
| mvp_form_demo | ✅ 正常 | 同上 |
| fiber_counter | ❌ 有问题 | UseStateInt, WithInit, OnPress 模式 |

## 可能的原因

1. **测试框架尚未验证**: `ui.RunTest` 可能还没有被充分测试
2. **初始化顺序问题**: Intent Runtime 在创建 Fiber 之前可能没有正确初始化
3. **事件处理差异**: Button 的 OnPress 模式 vs Input 的 ForField 模式可能有不同
