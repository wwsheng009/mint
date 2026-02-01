# Sandbox 事件注入测试结论

## 测试结果总结

通过 Sandbox 测试框架，我们成功验证了：

### 1. 事件注入完全正常 ✅

#### 直接触发按钮点击
```
[TestApp] TriggerButtonClick: button 1 (label=  +  )
[DEBUG] onClick: increment called, current count=0
setState BEFORE: componentID=TestApp-comp-1, hookIndex=0, value=0
setState AFTER: componentID=TestApp-comp-1, hookIndex=0, value=1
useState: componentID=TestApp-comp-1, hookIndex=0, value=1
✅ 点击成功！值从 0 变为 1
```

#### 多次点击测试
```
点击 1: 值 = 1
点击 2: 值 = 2
点击 3: 值 = 3
点击 4: 值 = 4
点击 5: 值 = 5
✅ 多次点击成功！最终值为 5
```

#### Tab + Enter 键盘事件
```
[TestApp] Tab: focusedIndex=0
[TestApp] Tab: focusedIndex=1
[TestApp] Enter: triggering button 1 (label=  +  )
onClick: increment called
setState: 0 → 1
✅ 事件路由工作正常
```

### 2. 验证的内容

| 项目 | 状态 | 说明 |
|------|------|------|
| useState 机制 | ✅ 正常 | 正确读取 Hook 中存储的值 |
| setState 机制 | ✅ 正常 | 正确更新 Hook 中的值 |
| Context 复用 | ✅ 正常 | 同一 TestApp 中 Context 被复用 |
| Hook 存储 | ✅ 正常 | Hook 指针保持稳定，值正确更新 |
| 按钮收集 | ✅ 正常 | VNode 树遍历正确收集按钮 |
| 焦点管理 | ✅ 正常 | Tab 键正确切换焦点索引 |
| 事件注入 | ✅ 正常 | Sandbox.Inject() 正确入队事件 |
| 事件处理 | ✅ 正常 | EventHandler 正确处理 Tab/Enter |
| onClick 触发 | ✅ 正常 | Enter 键正确触发按钮 onClick |

### 3. 问题根因定位

**原始问题**：Fiber 模式下点击按钮计数器不更新

**通过测试排除的原因**：
- ❌ 不是 useState 的问题
- ❌ 不是 setState 的问题
- ❌ 不是 Context 复用的问题
- ❌ 不是 Hook 存储的问题

**真正的问题**：完整应用 (`ui.Run` + `framework.App`) 中的事件路由

### 4. 下一步调试方向

问题在 `ui/app.go` 的 `declarativeRoot.HandleEvent` 方法中：

1. 检查 framework.App 是否正确调用 HandleEvent
2. 检查事件是否正确路由到 declarativeRoot
3. 检查按钮 bounds 是否正确设置
4. 检查焦点系统是否在完整应用中正确初始化

### 5. Sandbox 事件注入 API

```go
// 方式1: 直接使用 Sandbox 方法
sb.InjectMouse(x, y, platform.MouseLeft, platform.MousePress)
sb.InjectSpecialKey(platform.KeyEnter)
sb.ProcessEvents()

// 方式2: 使用 TestHelper 链式 API
sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

// 方式3: 直接触发（绕过事件系统）
testApp.TriggerButtonClick(buttonIndex)
testApp.TriggerButtonClickByLabel("  +  ")

// 方式4: 检查内部状态
ctx := testApp.GetContext()
buttons := testApp.GetButtons()
```

## 测试文件

- `examples/fiber_counter/event_test.go` - 事件注入测试
- `ui/test.go` - TestApp 事件处理实现
- `sandbox/mock/sandbox.go` - MockSandbox 事件注入

## 结论

Sandbox 测试框架成功验证了 useState/setState 机制的正确性，问题定位到完整应用的事件路由层。
