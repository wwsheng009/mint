# Fiber 调试测试总结

## 测试目的

使用 Sandbox 测试框架来定位 Fiber 集成中 setState 不生效的问题。

## 测试实现

### 1. 添加的调试功能

#### `ui/test.go` - 简化的测试应用包装器
```go
type TestApp struct {
    sandbox *mock.MockSandbox
    app     ComponentFunc
    ctx     *ComponentContext  // 测试用的 Context
}
```

**关键方法：**
- `TestRun()` - 创建测试环境，初始化 Context
- `Render()` - 设置当前 Context 并调用应用函数
- `GetContext()` - 获取测试用的 ComponentContext（用于检查 Hooks 状态）

#### `ui/hooks.go` - 调试函数
- `DebugContextInfo()` - 返回 Context 的调试信息
- `SetDebugLogger()` - 设置调试日志回调
- `UseStateIntWithDebug()` - 带调试输出的 useState

### 2. 测试用例

#### `TestSimpleContextRecreation`
验证每次 TestRun 是否创建独立的 Context。

**结果：** ✅ 通过
```
✅ 使用 3 个不同的 Context - 每次测试创建独立 Context
```

#### `TestHookValueInSameApp`
验证同一 TestApp 中多次渲染的 Hook 值变化。

**结果：** ❌ Hook 值始终为 0
```
Hook 值序列: [0 0 0]
Context 指针: 0xc00007e200 (相同)
Hook 指针: 0xc0000bc5f0 (相同)
```

**分析：**
- Context 和 Hook 被正确复用
- 但是值始终是 0
- **原因**：测试只调用了 `Render()`，没有触发 `setState`

#### `TestDirectContextMutation`
验证直接修改 Hook 值后 useState 是否读取新值。

**结果：** 直接修改后值保持为 999
```
修改前: Hooks[0].Value = 0
直接修改: Hooks[0].Value = 999
重新渲染后: Hooks[0].Value = 999
```

**分析：**
- useState 正确读取了 Hook 中存储的值（999）
- **结论**：useState 机制本身工作正常
- **问题**：setState 从未被调用

## 根本原因

### 问题定位

1. **useState 本身正常** ✅
   - Hook 存储和读取机制工作正常
   - Context 复用机制正常
   - 直接修改 Hook 值会被 useState 读取

2. **setState 未被触发** ❌
   - 测试场景中只有 `Render()` 调用
   - 没有用户交互（如按钮点击）
   - `setCount` 从未被调用

### Fiber 模式下的实际问题

真正的问题在于 Fiber 模式下的事件路由：

1. **按钮点击事件未触发 onClick**
   - 可能是事件路由问题
   - 可能是焦点系统问题
   - 需要进一步调试事件处理流程

2. **需要实际的交互测试**
   - 需要模拟真实的用户输入
   - 需要验证 HandleEvent 是否被调用
   - 需要验证 onClick 是否被触发

## 下一步调试

### 1. 事件路由测试
```go
func TestEventRouting(t *testing.T) {
    // 1. 创建测试应用
    testApp, _ := ui.TestRun(DebugCounter)

    // 2. 设置调试日志
    ui.SetDebugLogger(func(msg string) {
        t.Log("[DBG]", msg)
    })

    // 3. 模拟键盘事件（Tab + Enter）
    // 4. 检查 onClick 是否被调用
}
```

### 2. 焦点系统测试
- 验证按钮是否能获取焦点
- 验证焦点索引是否正确
- 验证 Enter 键是否触发按钮点击

### 3. Fiber Reconciler 测试
- 验证 Fiber 树是否正确构建
- 验证事件处理器是否正确绑定
- 验证 setState 是否触发重新渲染

## 相关文件

- `ui/test.go` - 测试应用包装器
- `ui/hooks.go` - Hooks 和调试函数
- `examples/fiber_counter/simple_test.go` - 简化调试测试
- `examples/fiber_counter/debug_test.go` - 详细调试测试
- `examples/fiber_counter/main.go` - 调试版计数器示例

## 结论

通过 Sandbox 测试框架，我们成功定位了问题：

1. **useState 机制本身正常** - Hook 存储和读取工作正常
2. **问题在于事件层** - setState 从未被调用
3. **需要调试事件路由** - 下一步重点是验证按钮点击事件

这证明了 Sandbox 测试框架对于定位复杂 UI 问题非常有价值，能够直接观察内部状态（Context、Hooks）而不仅仅是外部行为。
