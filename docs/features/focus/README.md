# 焦点管理功能文档

本目录包含 Mint TUI 框架焦点管理系统的实现文档和分析报告。

## 功能概述

焦点管理系统支持多种焦点切换方式：

- ✅ **Tab 键导航** - 在可聚焦组件间循环切换焦点
- ✅ **Shift+Tab** - 反向导航
- ✅ **鼠标点击** - 点击组件自动切换焦点
- ✅ **Modal 焦点捕获** - Modal 打开时焦点限制在 Modal 内

## 文档列表

### 📋 问题分析
- [`mouse_click_focus_issue.md`](./mouse_click_focus_issue.md) - 鼠标点击焦点问题分析
  - 问题描述：Tab 键可以切换焦点，但鼠标点击不能
  - 根本原因：Button.HandleEvent() 只触发 onClick，不请求焦点
  - 解决方案：在事件路由层自动处理焦点切换

### 📖 实现文档
- [`tab_key_focus_implementation.md`](./tab_key_focus_implementation.md) - Tab 键焦点切换实现位置
  - VNodeFocusManager.HandleEvent() 的实现
  - 焦点管理器的完整流程
  - 焦点收集和应用机制

- [`mouse_click_focus_implementation.md`](./mouse_click_focus_implementation.md) - 鼠标点击焦点切换实现
  - 实现方式：在事件处理流程中插入焦点切换
  - Hit Testing 机制：使用 bounds 进行命中检测
  - Modal 焦点捕获：只收集 modal 层的可聚焦节点
  - 事件处理优先级：Layer → FocusManager → MouseFocus → Component

## 关键代码位置

### 焦点管理器
- **文件**: `runtime/ui/focus_manager.go`
- **关键方法**:
  - `HandleEvent()` - 处理 Tab/Shift+Tab 键 (行 206-212)
  - `FocusNext()` - 切换到下一个焦点 (行 95-108)
  - `FocusPrev()` - 切换到上一个焦点 (行 110-123)
  - `DispatchToFocused()` - 分发事件到焦点元素 (行 217-230)

### 事件处理流程
- **文件**: `internal/render/declarative_node.go`
- **关键方法**:
  - `HandleEvent()` - 主事件处理流程 (行 1039-1100)
  - `handleMouseFocus()` - 鼠标焦点切换 (行 1154-1193)
  - `nodeWasClicked()` - Hit Testing (行 1195-1227)

### 组件接口
- **FocusableVNode**: `runtime/ui/vnode.go`
  - `IsFocusable() bool` - 是否可聚焦
  - `SetFocus(bool)` - 设置焦点状态
  - `GetBounds() (x, y, width, height int)` - 获取边界（用于 hit testing）

## 事件处理流程

```
用户输入事件
    ↓
┌─────────────────────────────────────────┐
│ 0. Layer 事件处理                       │
│    - ESC 关闭 modal                     │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ 1. 焦点管理器处理                       │
│    - Tab 键 → FocusNext()               │
│    - Shift+Tab → FocusPrev()            │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ 1.5. 鼠标焦点切换 ✨                    │
│    - 收集 focusable 节点                │
│    - Hit testing 找到被点击节点          │
│    - SetFocusByIndex() 切换焦点          │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ 2. 分发事件到焦点元素                   │
│    - Enter/Space 触发按钮               │
│    - 字符输入到 Input                   │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ 3. 全局事件分发                         │
│    - 其他未处理的事件                   │
└─────────────────────────────────────────┘
```

## 可聚焦组件

以下组件支持焦点管理：

- ✅ **Button** - `components/button/button.go`
- ✅ **Input** - `components/form/input.go`
- ✅ **Textarea** - `components/form/textarea.go`
- ✅ **Select** - `components/form/select.go`
- ✅ **Checkbox** - `components/form/checkbox.go`

## Modal 焦点捕获

当 Modal 打开时，焦点管理系统会自动捕获焦点：

```go
hasModal := rtui.HasModalInTree(n.root)

if hasModal {
    // 只收集 modal 层的 focusable 节点
    focusable = rtui.CollectFocusableInLayer(n.root, rtui.LayerModal)
} else {
    // 收集所有 focusable 节点
    focusable = rtui.CollectFocusable(n.root)
}
```

**效果**：
- Tab 键只在 modal 内导航
- 鼠标点击只能在 modal 内切换焦点
- Modal 外的组件不受影响

## 测试验证

### 功能测试

1. **基本焦点切换**：
   ```bash
   cd examples/ui_demos/demo1_full_featured
   go run main.go
   # 点击不同按钮，观察焦点切换
   # 按 Tab 键，观察焦点导航
   ```

2. **Modal 焦点捕获**：
   ```bash
   # 打开 modal
   # 尝试点击 modal 外的按钮
   # 验证焦点不切换到外部
   ```

3. **键盘导航**：
   ```bash
   # Tab 键 → 下一个焦点
   # Shift+Tab → 上一个焦点
   # Enter/Space → 触发按钮动作
   ```

### 调试模式

启用调试日志查看详细焦点切换过程：

```bash
export TUI_DEBUG_UI=true
go run examples/ui_demos/demo1_full_featured/main.go
```

**日志输出**：
```
FocusNext current=0, len(focusable)=5
FocusNext: now at index 1
handleMouseFocus: switching focus from index 0 to 1
nodeWasClicked: bounds=(2, 7, 11, 1), mouse=(8, 7)
nodeWasClicked: HIT!
```

## 实现亮点

1. **零组件侵入性**
   - 组件无需修改代码
   - 只需实现 SetFocus() 和 GetBounds()（已有）
   - 自动支持所有可聚焦组件

2. **智能焦点管理**
   - 自动检测 modal 状态
   - Modal 打开时捕获焦点
   - Modal 关闭后恢复全局焦点

3. **高效的 Hit Testing**
   - 使用 bounds 而不是递归遍历
   - O(n) 复杂度，n 是 focusable 节点数量
   - 通常 n 很小（< 20），性能可忽略

4. **事件流清晰**
   - 优先级明确
   - 易于理解和调试
   - 不破坏现有功能

## 相关文档

- [布局系统](../../layout/) - 布局引擎和约束系统
- [主题系统](../../theme/) - 焦点样式和主题颜色
- [事件系统](../../sandbox/) - 事件处理和分发机制
