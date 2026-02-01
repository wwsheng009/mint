# Fiber 问题排查指南

## 当前状态

### ✅ 已验证正常（通过 Sandbox 测试）
- useState/setState 机制完全正常
- Context 复用正常
- Hook 存储和读取正常
- 按钮收集和焦点管理正常
- onClick 事件处理正常

### ❌ 问题所在
**完整应用 (`ui.Run`) 中的事件路由**

## 调试步骤

### 1. 启动应用并观察调试输出

```bash
cd E:\projects\yao\wwsheng009\mint
set MINT_USE_FIBER=true
set TUI_DEBUG_UI=true
go run ./examples/fiber_counter/main.go
```

### 2. 观察启动时的日志

应用启动时应该看到：
```
[paintWithFiber] Reset interactive elements, starting Fiber render
[renderVNodeFiber] Collected button 0: label=  -
[renderVNodeFiber] Collected button 1: label=  +
After render: buttons=2, totalElements=2, focusedIndex=0
```

**检查点**：
- [ ] 是否看到 `paintWithFiber` 消息？
- [ ] 是否收集到了 2 个按钮？
- [ ] focusedIndex 是否为 0？

### 3. 按 Tab 键并观察

按 Tab 键后应该看到：
```
[HandleEvent] Called: Type=1, Special=3
[HandleEvent] State: focusedIndex=0, totalElements=2, buttons=2
[HandleEvent] KeyTab: focusedIndex=0 -> 1
```

**检查点**：
- [ ] `HandleEvent` 是否被调用？
- [ ] Special=3 是否正确（KeyTab）？
- [ ] focusedIndex 是否从 0 变为 1？

### 4. 按 Enter 键并观察

按 Enter 键后应该看到：
```
[HandleEvent] Called: Type=1, Special=2
[HandleEvent] State: focusedIndex=1, totalElements=2, buttons=2
[HandleEvent] KeyEnter: focusedIndex=1, totalElements=2, buttons=2
[HandleEvent] getElementByIndex(1) -> elemType=0
[HandleEvent] Triggering button: label=  +  , hasOnClick=true
[DEBUG] onClick: increment called
```

**检查点**：
- [ ] `elemType` 是否为 0（Button）？
- [ ] `hasOnClick` 是否为 true？
- [ ] `onClick` 是否被调用？

## 可能的问题点

### 问题 1: 按钮未被收集
**症状**：`buttons=0`

**原因**：
- Fiber 模式未启用（检查 `MINT_USE_FIBER=true`）
- `renderVNodeFiber` 未被调用
- 按钮被标记为 Disabled

**解决**：检查 `paintWithFiber` 是否被调用

### 问题 2: HandleEvent 未被调用
**症状**：按 Tab/Enter 无任何日志

**原因**：
- framework.App 未正确转发事件到 declarativeRoot
- 事件被其他组件拦截

**解决**：检查 framework.App 的事件处理流程

### 问题 3: 按钮收集了但 focusedIndex 错误
**症状**：`buttons=2` 但 `focusedIndex=-1` 或超出范围

**原因**：
- `getTotalFocusableCount()` 计算错误
- `focusedIndex` 未正确初始化

**解决**：检查 `resetInteractiveElements()` 和 `getTotalFocusableCount()`

### 问题 4: elemType 错误
**症状**：`getElementByIndex` 返回错误的类型

**原因**：
- 按钮和输入元素混合计数导致索引错位
- `getFirstElementType()` 返回错误的类型

**解决**：检查 `getElementByIndex()` 的逻辑

## 快速诊断脚本

创建一个快速测试来验证核心功能：

```bash
# 运行 Sandbox 测试（应该通过）
go test ./examples/fiber_counter/... -v -run TestDirectButtonClick

# 运行多次点击测试
go test ./examples/fiber_counter/... -v -run TestMultipleButton
```

## 下一步

根据调试输出确定问题点后：

1. **如果按钮未收集** → 检查 Fiber reconciler
2. **如果 HandleEvent 未调用** → 检查 framework.App
3. **如果 elemType 错误** → 修复 `getTotalFocusableCount()`
4. **如果 onClick 未调用** → 检查按钮的 OnClick 绑定
