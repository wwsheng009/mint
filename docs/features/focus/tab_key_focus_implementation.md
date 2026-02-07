# Tab 键切换焦点的实现位置

## 发现

Tab 键切换焦点功能**已实现**，位于 `runtime/ui/focus_manager.go`。

---

## 实现位置

### 1. VNodeFocusManager.HandleEvent() - Tab 键处理

**文件**: `runtime/ui/focus_manager.go:206-212`

```go
// HandleEvent processes keyboard events for focus navigation.
// Returns (handled, shouldRender) tuple.
func (m *VNodeFocusManager) HandleEvent(ev event.Event) (bool, bool) {
    // ... 特殊键处理 ...

    // Tab - navigate to next  ← Tab 键处理
    if keyEvent.Special == event.KeyTab {
        if keyEvent.Modifiers == event.ModShift {
            return m.FocusPrev(), true  // Shift+Tab: 上一个
        }
        return m.FocusNext(), true        // Tab: 下一个
    }

    return false, false
}
```

**关键点**：
- ✅ 检测 `KeyTab`
- ✅ Shift+Tab 调用 `FocusPrev()`
- ✅ Tab 调用 `FocusNext()`
- ✅ 返回 `(handled, true)` 表示需要重新渲染

---

### 2. DeclarativeNode.HandleEvent() - 调用焦点管理器

**文件**: `internal/render/declarative_node.go:1039-1072`

```go
func (n *DeclarativeNode) HandleEvent(ev frameworkevent.Event) bool {
    // ... ESC 键处理 ...

    // 1. Let focus manager handle navigation (Tab, Shift+Tab)  ← 关键步骤
    if focusMgr != nil {
        handled, shouldRender := focusMgr.HandleEvent(ev)
        if handled {
            // 焦点已切换，请求重新渲染
            if shouldRender {
                n.requestRender(useFiber, reconciler)
            }
            return true  // ← 事件已处理，停止传播
        }

        // 2. Try to dispatch to the focused element first
        if focusMgr.DispatchToFocused(ev) {
            return true  // ← 焦点元素处理了事件（如 Enter 触发按钮）
        }
    }

    // 3. Fall back to global event distribution
    handled := n.distributeEventToVNode(root, ev)
    if handled {
        n.requestRender(useFiber, reconciler)
    }
    return handled
}
```

**事件处理优先级**：
1. **焦点管理器处理 Tab/Shift+Tab** → 焦点切换
2. **分发到焦点元素** → Enter/Space 触发按钮
3. **全局事件分发** → 其他事件

---

### 3. 焦点切换的完整流程

```
Tab 键按下
    ↓
framework/app.go:handleEvent()
    ├→ router.Route(ev)  // 事件路由
    └→ root.HandleEvent(ev)
        ↓
DeclarativeNode.HandleEvent(ev)
    ├→ focusMgr.HandleEvent(ev)  ← 调用焦点管理器
    │   ↓
    │   VNodeFocusManager.HandleEvent()
    │   ├→ 检测到 KeyTab
    │   ├→ 调用 FocusNext()
    │   └→ 返回 (true, true)
    │
    ├→ requestRender()  ← 请求重新渲染
    └→ return true
        ↓
App 主循环
    └→ render()  ← 重新绘制，显示新的焦点状态
```

---

## VNodeFocusManager 的完整实现

### FocusNext() - 下一个焦点

**文件**: `runtime/ui/focus_manager.go:95-108`

```go
func (m *VNodeFocusManager) FocusNext() bool {
    log.FocusLogger.Debug("FocusNext current=%d, len(focusable)=%d", m.current, len(m.focusable))
    if len(m.focusable) == 0 {
        return false
    }

    // 清除当前焦点
    if m.current >= 0 && m.current < len(m.focusable) {
        m.focusable[m.current].SetFocus(false)
    }

    // 移动到下一个
    m.current = (m.current + 1) % len(m.focusable)
    m.focusable[m.current].SetFocus(true)

    // 触发回调
    if m.onNavigate != nil && m.current > 0 {
        oldIndex := (m.current - 1 + len(m.focusable)) % len(m.focusable)
        m.onNavigate(m.focusable[oldIndex], m.focusable[m.current])
    }

    log.FocusLogger.Debug("FocusNext: now at index %d", m.current)
    return true
}
```

**逻辑**：
1. 检查是否有 focusable 节点
2. 清除当前节点的焦点状态（`SetFocus(false)`）
3. 将焦点索引移动到下一个（循环）
4. 设置新节点的焦点状态（`SetFocus(true)`）
5. 触发导航回调（如果有）

---

### FocusPrev() - 上一个焦点

**文件**: `runtime/ui/focus_manager.go:110-123`

```go
func (m *VNodeFocusManager) FocusPrev() bool {
    if len(m.focusable) == 0 {
        return false
    }

    // 清除当前焦点
    if m.current >= 0 && m.current < len(m.focusable) {
        m.focusable[m.current].SetFocus(false)
    }

    // 移动到上一个（循环）
    m.current = (m.current - 1 + len(m.focusable)) % len(m.focusable)
    m.focusable[m.current].SetFocus(true)

    // 触发回调
    if m.onNavigate != nil {
        oldIndex := (m.current + 1) % len(m.focusable)
        m.onNavigate(m.focusable[oldIndex], m.focusable[m.current])
    }

    return true
}
```

---

## 焦点收集机制

### CollectFocusable() - 收集可聚焦节点

**文件**: `runtime/ui/focus_manager.go:278-281`

```go
func CollectFocusable(root VNode) []FocusableVNode {
    return collectFocusableRecursive(root)
}
```

**收集逻辑**：
1. 递归遍历 VNode 树
2. 检查每个节点是否实现 `FocusableVNode` 接口
3. 检查 `IsFocusable()` 方法返回 true
4. 返回所有可聚焦的节点列表

**哪些组件是 focusable 的**：
- ✅ Button (`components/button/button.go`)
- ✅ Input (`components/form/input.go`)
- ✅ Textarea (`components/form/textarea.go`)
- ✅ Select (`components/form/select.go`)
- ✅ Checkbox (`components/form/checkbox.go`)

---

## 焦点状态应用

### applyFocusState() - 应用焦点状态

**文件**: `internal/render/declarative_node.go:294-335`

```go
func (n *DeclarativeNode) applyFocusState() {
    if n.focusMgr == nil || n.root == nil {
        return
    }

    var focusable []rtui.FocusableVNode

    // 检查是否有 modal 打开 - 如果有，捕获焦点在 modal 内
    hasModal := rtui.HasModalInTree(n.root)

    if hasModal {
        // 焦点捕获：只从 modal 层收集可聚焦节点
        focusable = rtui.CollectFocusableInLayer(n.root, rtui.LayerModal)
    } else {
        // 没有 modal：收集所有可聚焦节点
        focusable = rtui.CollectFocusable(n.root)
    }

    n.focusMgr.UpdateFocusableList(focusable)

    // 限制焦点索引范围
    currentIndex := n.focusMgr.CurrentIndex()
    if currentIndex >= len(focusable) {
        currentIndex = len(focusable) - 1
    }
    if currentIndex < 0 && len(focusable) > 0 {
        currentIndex = 0
    }
    if currentIndex >= 0 {
        n.focusMgr.SetFocusByIndex(currentIndex)
    }

    // 应用焦点状态
    n.applyFocus(focusable)
}
```

**逻辑**：
1. 检查是否有 modal 打开
2. 如果有 modal，只收集 modal 内的 focusable 节点（焦点捕获）
3. 如果没有 modal，收集所有 focusable 节点
4. 更新焦点管理器的列表
5. 设置焦点索引
6. 应用焦点状态（调用每个节点的 `SetFocus()`）

---

## DispatchToFocused() - 分发事件到焦点元素

**文件**: `runtime/ui/focus_manager.go:217-230`

```go
func (m *VNodeFocusManager) DispatchToFocused(ev event.Event) bool {
    if m.current < 0 || m.current >= len(m.focusable) {
        return false
    }

    // 尝试分发到焦点组件（如果它实现了 HandleEvent）
    if component, ok := m.focusable[m.current].(interface{ HandleEvent(event.Event) bool }); ok {
        return component.HandleEvent(ev)
    }

    return false
}
```

**作用**：
- 将键盘事件（Enter、Space、方向键等）发送给当前焦点元素
- 焦点元素可以响应用户输入（如 Input 接收字符、Button 响应 Enter）

---

## 总结

### Tab 键焦点切换的完整流程

```
1. 用户按下 Tab 键
   ↓
2. 框架层捕获键盘事件
   ↓
3. DeclarativeNode.HandleEvent() 接收事件
   ↓
4. 调用 VNodeFocusManager.HandleEvent()
   ↓
5. 检测到 KeyTab，调用 FocusNext()
   ├→ 清除当前焦点：oldFocusable.SetFocus(false)
   ├→ 移动焦点索引：current = (current + 1) % len(focusable)
   └→ 设置新焦点：newFocusable.SetFocus(true)
   ↓
6. 返回 (handled: true, shouldRender: true)
   ↓
7. DeclarativeNode 请求重新渲染
   ↓
8. App.render() 重新绘制
   ↓
9. 新的焦点状态显示在屏幕上
```

### 关键文件和代码位置

| 功能 | 文件 | 行号 |
|------|------|------|
| Tab 键检测 | `runtime/ui/focus_manager.go` | 206-212 |
| FocusNext() | `runtime/ui/focus_manager.go` | 95-108 |
| FocusPrev() | `runtime/ui/focus_manager.go` | 110-123 |
| 调用焦点管理器 | `internal/render/declarative_node.go` | 1039-1072 |
| 焦点收集 | `runtime/ui/focus_manager.go` | 278-287 |
| 焦点状态应用 | `internal/render/declarative_node.go` | 294-335 |
| 事件分发到焦点元素 | `runtime/ui/focus_manager.go` | 217-230 |

### 为什么 Tab 可以切换焦点？

✅ **已实现**：
1. VNodeFocusManager.HandleEvent() 检测 Tab 键
2. 调用 FocusNext()/FocusPrev() 切换焦点索引
3. 更新焦点状态（SetFocus(true/false)）
4. 请求重新渲染显示新的焦点状态

### 为什么鼠标点击不切换焦点？

❌ **未实现**：
1. Button.HandleEvent() 处理鼠标点击时没有请求焦点
2. DeclarativeNode 没有拦截鼠标点击来切换焦点
3. 缺少"点击 → 切换焦点 → 分发事件"的流程
