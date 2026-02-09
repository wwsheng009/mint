# Inspector 数字键响应问题 - 分析与解决方案

## 问题确认

根据测试验证：
- ✅ 数字键 **确实被处理了**
- ✅ `activeTab` **确实切换到了 Layout tab**
- ✅ `RenderContent()` **可以生成正确的内容**
- ❌ **UI 没有立即更新**

## 根本原因

Inspector 通过 **render hook** 系统注入到应用中：

```
用户按数字键 "5"
    ↓
HandleKeyEvent("5") 被调用
    ↓
activeTab 改变为 TabLayout ✅
    ↓
返回 true（阻止事件传播）
    ↓
应用不知道需要重新渲染 ❌
    ↓
Render hook 不会被调用
    ↓
UI 不更新 ❌
```

## 当前行为

### Tab 键能切换的原因

按 **Tab** 键后，可能：
1. 返回 false（让事件传播）
2. 或者触发了某种"默认按钮焦点变化"机制
3. 应用接收事件后触发重新渲染
4. Render hook 被调用，UI 更新

### 数字键没有反馈的原因

按 **数字键** 后：
1. HandleKeyEvent 返回 true（事件被阻止）
2. 没有触发应用重新渲染的机制
3. Inspector 自己不控制渲染循环
4. 必须等待下一个事件才能看到变化

## 解决方案

### 方案 1：使用 Tab 键切换（立即可用）

**当前可用的方式**：
```
1 → 切换到 Elements tab
Tab → 切换到下一个 tab（Console）
Tab → 切换到下一个 tab（Performance）
Tab → 切换到下一个 tab（Diagnostics）
Tab → 切换到下一个 tab（Layout）← 多按几次到达
Tab → 切换到下一个 tab（Network）
```

### 方案 2：触发其他事件来强制重新渲染

按数字键后，再按：
- **任意其他键**（如方向键）
- 或 **鼠标移动**
- 或 **Enter** 键

这会触发新的事件循环，应用会重新渲染，Inspector UI 会更新。

### 方案 3：让数字键事件传播（推荐）

修改代码，让数字键返回 false 而不是 true：

```go
// Tab switching
if key == "1" {
    si.activeTab = TabElements
    return false  // 改为 false，让事件传播
}
```

**优点**：事件会传播到应用，触发重新渲染
**缺点**：可能会影响其他组件（但不应该，因为数字键不是标准快捷键）

### 方案 4：添加显式刷新机制

需要框架支持"请求重新渲染"的 API。这是长期解决方案。

## 推荐的临时解决方案

**使用 Tab 键循环切换**：

```
F12          → 打开 Inspector
Tab          → 切换到 Console
Tab          → 切换到 Performance
Tab          → 切换到 Diagnostics
Tab          → 切换到 Layout ← 到达！
```

或者按一下数字键后，再按一下**方向键**或**Enter**来触发刷新。

## 确认可用的功能

根据测试，这些功能是确认可用的：

1. ✅ **F12** - 打开/关闭 Inspector
2. ✅ **Tab / Shift+Tab** - 循环切换 tabs（有视觉反馈）
3. ✅ **↑↓ Enter** - 在 Elements tab 中导航和选择节点
4. ✅ **Alt+H/J/K/L** - 移动 Inspector 窗口

## 总结

当前 Inspector 的设计限制：
- Inspector 通过 hook 系统注入，由应用渲染循环控制
- Inspector 处理按键后，应用不会自动重新渲染
- 需要下一个事件才能看到 UI 更新

**立即可用的解决方案**：
- 使用 **Tab 键**切换到 Layout tab（多按几次到达）
- 或者按数字键后，再按一下方向键触发刷新

长期解决方案需要框架支持"请求渲染"机制。
