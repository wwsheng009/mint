# 鼠标点击与焦点管理关系分析报告

## 问题描述

**现象**：
- ✅ Tab 键可以切换焦点
- ❌ 鼠标点击按钮时，焦点不会切换到该按钮

## 根本原因

### 1. Button 组件的 HandleEvent 实现

**文件**: `components/button/button.go:318-417`

```go
func (b *ButtonVNode) HandleEvent(e event.Event) bool {
    // 键盘事件处理 - 只有在获得焦点时才响应
    if keyEvent, ok := e.(*event.KeyEvent); ok {
        if !b.hasFocus {
            return false  // ❌ 没有焦点时不处理键盘事件
        }
        // 处理 Enter/Space 键...
    }

    // 鼠标事件处理
    switch mouseEvent.Type() {
    case event.EventMousePress:
        if b.ContainsPoint(mouseEvent.X, mouseEvent.Y) && mouseEvent.Button == event.MouseLeft {
            // ❌ 只调用 onClick，没有设置焦点
            if b.onClick != nil {
                b.onClick()
            }
            return true
        }

    case event.EventClick:
        if b.isHovered && mouseEvent.Button == event.MouseLeft {
            // ❌ 只调用 onClick，没有设置焦点
            if b.onClick != nil {
                b.onClick()
            }
            return true
        }
    }
}
```

**问题**：鼠标点击时，Button 组件只触发 `onClick()` 回调，**没有请求焦点**。

---

### 2. VNodeFocusManager 有焦点切换方法，但没被使用

**文件**: `runtime/ui/focus_manager.go:95-132`

```go
// FocusNext moves focus to the next focusable node
func (m *VNodeFocusManager) FocusNext() bool {
    if len(m.focusable) == 0 {
        return false
    }

    // Clear current focus
    if m.current >= 0 && m.current < len(m.focusable) {
        m.focusable[m.current].SetFocus(false)
    }

    // Move to next
    m.current = (m.current + 1) % len(m.focusable)
    m.focusable[m.current].SetFocus(true)

    // Trigger callback
    if m.onNavigate != nil {
        m.onNavigate(m.focusable[(m.current-1+len(m.focusable))%len(m.focusable)], m.focusable[m.current])
    }

    return true
}
```

**问题**：`FocusNext()` / `FocusPrev()` 方法存在，但是：
- ❌ 没有键盘事件处理器调用它们
- ❌ 鼠标点击时也不会调用它们
- ❌ Tab 键可能是在组件级别手动处理的（而不是通过焦点管理器）

---

### 3. Tab 键切换的真相

**Tab 键可能工作**的原因：

1. **可能在 app 层有默认的键盘映射**
   ```go
   // framework/app.go 可能注册了 Tab 键的默认处理
   ```

2. **或者在 DeclarativeNode 层有处理**
   ```go
   // internal/render/declarative_node.go 可能有 Tab 键处理
   ```

3. **或者根本没实现 Tab 导航**
   - 用户看到的 Tab 切换可能只是**单个组件内部**的焦点切换
   - 而不是**跨组件**的焦点切换

让我验证一下...

---

## 验证实验

### 实验 1：查找 Tab 键处理器

```bash
grep -rn "KeyTab" --include="*.go" | grep -i "focus"
```

**预期结果**：
- 如果找到：说明 Tab 键调用了焦点管理器
- 如果没找到：说明 Tab 键切换没有实现，或者是在组件级别处理的

---

### 实验 2：查看 demo1 的焦点切换

**测试步骤**：
1. 启动 demo1
2. 按 Tab 键多次
3. 观察焦点是否在多个按钮间切换

**预期结果**：
- 如果焦点只在 modal 内的按钮间切换：说明 modal 有自己的焦点管理
- 如果焦点在整个应用的按钮间切换：说明有全局焦点管理
- 如果焦点不切换：说明 Tab 键处理根本没实现

---

## 解决方案

### 方案 1：在 Button 组件中添加焦点请求（推荐）

**修改**: `components/button/button.go`

```go
func (b *ButtonVNode) HandleEvent(e event.Event) bool {
    // ... 现有代码 ...

    switch mouseEvent.Type() {
    case event.EventMousePress:
        if b.ContainsPoint(mouseEvent.X, mouseEvent.Y) && mouseEvent.Button == event.MouseLeft {
            // ✅ 添加：请求焦点
            if !b.hasFocus {
                b.RequestFocus()
            }

            // 触发点击事件
            if b.onClick != nil {
                b.onClick()
            }
            return true
        }

    case event.EventClick:
        if b.isHovered && mouseEvent.Button == event.MouseLeft {
            // ✅ 添加：请求焦点
            if !b.hasFocus {
                b.RequestFocus()
            }

            // 触发点击事件
            if b.onClick != nil {
                b.onClick()
            }
            return true
        }
    }
}
```

**需要实现**：`RequestFocus()` 方法

```go
// RequestFocus requests focus for this button
func (b *ButtonVNode) RequestFocus() bool {
    // 通过 FocusManager 请求焦点
    // 这需要访问到 VNodeFocusManager
}
```

---

### 方案 2：在事件路由层自动处理焦点切换

**修改**: `internal/render/declarative_node.go`

```go
func (n *DeclarativeNode) HandleEvent(ev event.Event) bool {
    // ... 现有代码 ...

    // ✅ 添加：鼠标点击时自动切换焦点
    if ev.Type() == event.EventMousePress || ev.Type() == event.EventClick {
        if mouseEv, ok := ev.(*event.MouseEvent); ok {
            // 通过 hit testing 找到被点击的节点
            target := n.hitTest(mouseEv.X, mouseEv.Y)

            // 如果目标节点是 focusable，切换焦点到它
            if focusable, ok := target.(rtui.FocusableVNode); ok {
                n.focusMgr.SetFocusByID(focusable.GetFocusID())
                return true
            }
        }
    }

    // ... 其他代码 ...
}
```

---

### 方案 3：实现全局键盘导航处理

**新建**: `internal/render/focus_navigation.go`

```go
package render

import (
    "github.com/wwsheng009/mint/framework/event"
)

// setupKeyboardNavigation 注册全局键盘导航快捷键
func (n *DeclarativeNode) setupKeyboardNavigation() {
    // Tab 键：下一个焦点
    n.router.Register(event.EventKeyPress, event.EventHandlerFunc(func(ev event.Event) bool {
        if keyEv, ok := ev.(*event.KeyEvent); ok && keyEv.Special == event.KeyTab {
            // Shift+Tab：上一个焦点，Tab：下一个焦点
            if keyEv.Modifiers&event.ModShift != 0 {
                return n.focusMgr.FocusPrev()
            }
            return n.focusMgr.FocusNext()
        }
        return false
    }))
}
```

---

## 当前架构问题

### 问题 1：焦点管理不统一

```
┌─────────────────────────────────────┐
│ App Layer (framework/app.go)        │
│ - 处理事件路由                        │
│ - ❌ 没有全局焦点管理                 │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│ DeclarativeNode                     │
│ - 有 VNodeFocusManager               │
│ - ❌ 没有注册 Tab 键处理              │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│ Components (Button, Input, etc.)    │
│ - 有 SetFocus() 方法                 │
│ - ❌ 没有请求焦点的方法                │
│ - ❌ 鼠标点击时不请求焦点             │
└─────────────────────────────────────┘
```

### 问题 2：事件流向

```
鼠标点击流程：

1. 鼠标点击 → framework/app.go:handleEvent()
2. handleEvent() → root.HandleEvent(ev)
3. root.HandleEvent() → DeclarativeNode.HandleEvent()
4. DeclarativeNode.HandleEvent() → 分发到子组件
5. Button.HandleEvent() → onClick() ❌ 没有请求焦点

Tab 键流程（未知）：

1. Tab 键 → framework/app.go:handleEvent()
2. handleEvent() → keyMap.Lookup()
3. ❌ 没有找到 Tab 键的默认处理
4. handleEvent() → root.HandleEvent(ev)
5. ❌ DeclarativeNode 可能没有处理 Tab
6. Button.HandleEvent() → 忽略（因为不是 Enter/Space）
```

---

## 推荐实现步骤

### 阶段 1：添加 RequestFocus 支持

1. **VNodeFocusManager 添加公共方法**：
   ```go
   func (m *VNodeFocusManager) RequestFocus(id string) bool {
       for i, node := range m.focusable {
           if node.GetFocusID() == id {
               return m.SetFocusByIndex(i)
           }
       }
       return false
   }
   ```

2. **DeclarativeNode 暴露焦点管理器**：
   ```go
   func (n *DeclarativeNode) GetFocusManager() *rtui.VNodeFocusManager {
       return n.focusMgr
   }
   ```

3. **Button 组件请求焦点**：
   ```go
   func (b *ButtonVNode) requestFocusThroughParent() {
       // 通过某种方式访问父节点的焦点管理器
   }
   ```

### 阶段 2：实现全局键盘导航

1. **在 DeclarativeNode.Init() 中注册 Tab 键处理**：
   ```go
   func (n *DeclarativeNode) Init() error {
       // ... 现有代码 ...

       // 注册全局导航快捷键
       n.registerNavigationShortcuts()

       return nil
   }
   ```

2. **实现 registerNavigationShortcuts()**：
   ```go
   func (n *DeclarativeNode) registerNavigationShortcuts() {
       n.keyMap.Register(event.KeyEvent{
           Special: event.KeyTab,
       }, event.HandlerFunc{
           F: func(ev event.Event) bool {
               return n.focusMgr.FocusNext()
           },
       })
   }
   ```

### 阶段 3：鼠标点击自动切换焦点

1. **在 DeclarativeNode.HandleEvent() 中拦截鼠标事件**：
   ```go
   func (n *DeclarativeNode) HandleEvent(ev event.Event) bool {
       // 鼠标点击事件：先切换焦点，再分发到子组件
       if ev.Type() == event.EventMousePress || ev.Type() == event.EventClick {
           if mouseEv, ok := ev.(*event.MouseEvent); ok {
               if n.handleMouseFocus(mouseEv) {
                   // 焦点已切换，继续处理事件
               }
           }
       }

       // 继续正常的事件分发
       return n.handleEventDefault(ev)
   }
   ```

2. **实现 handleMouseFocus()**：
   ```go
   func (n *DeclarativeNode) handleMouseFocus(ev *event.MouseEvent) bool {
       // Hit testing 找到被点击的焦点节点
       target := n.hitTest(ev.X, ev.Y)

       // 如果是 focusable 节点，切换焦点
       if focusable, ok := target.(rtui.FocusableVNode); ok {
           return n.focusMgr.SetFocusByID(focusable.GetFocusID())
       }

       return false
   }
   ```

---

## 总结

### 当前状态

| 功能 | 状态 | 说明 |
|------|------|------|
| VNodeFocusManager | ✅ 已实现 | 有 FocusNext/FocusPrev 方法 |
| 焦点列表收集 | ✅ 已实现 | CollectFocusable() 正常工作 |
| 焦点状态应用 | ✅ 已实现 | SetFocus() 正常工作 |
| Tab 键导航 | ❓ 未知 | 可能没有全局实现 |
| 鼠标点击焦点 | ❌ 未实现 | Button 不请求焦点 |

### 为什么鼠标点击没有焦点切换？

1. **Button 组件不请求焦点**：
   - 鼠标点击时只调用 `onClick()`
   - 没有调用焦点管理器的 `SetFocusByID()`

2. **没有自动焦点切换机制**：
   - DeclarativeNode 层没有拦截鼠标点击事件
   - 事件直接分发到子组件，不经过焦点管理

3. **Tab 键可能也没正确实现**：
   - 没有找到 Tab 键调用 FocusNext() 的代码
   - 可能用户看到的 Tab 切换是单个组件内部行为

### 下一步

需要实现：
1. ✅ Button 组件在鼠标点击时请求焦点
2. ✅ 全局 Tab 键导航处理
3. ✅ DeclarativeNode 拦截鼠标事件自动切换焦点
