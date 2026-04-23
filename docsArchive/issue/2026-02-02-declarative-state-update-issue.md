# 声明式UI状态更新与事件分发问题排查与修复

**日期**: 2026-02-02
**问题**: Sandbox 测试中，按钮点击后状态未正确更新，显示 "Count: 0" 而非预期值
**状态**: ✅ 已修复事件分发问题，状态管理正常工作，待实现焦点管理集成

---

## 1. 问题发现

### 1.1 初始现象
在 `TestCounterWithRunTest/IncrementCount` 测试中：
- 预期: 按 Tab → Enter → Count 从 0 变为 1
- 实际: 按键后 Count 仍为 0

### 1.2 调试过程

#### 第一步：确认事件是否到达组件
在 `ButtonVNode.HandleEvent` 和 `DeclarativeNode.HandleEvent` 中添加调试输出：

```go
// Button HandleEvent called: label="  [ - ]  ", disabled=false, event type=*event.KeyEvent
// Button HandleEvent: KeyEvent, Special=2 (enter), Rune=, KeyEnter=2
// Button HandleEvent: triggering onClick for label="  [ - ]  " (Special=2)
```

事件确实到达了按钮组件。

#### 第二步：确认 setState 是否被调用
```go
// setState BEFORE: hookIndex=0, value=0
// setState AFTER: hookIndex=0, value=-1
```

setState 正确修改了 hook 值。

#### 第三步：确认状态是否持久化
```go
// GetOrCreateHook: returning existing hook[0], Type=useState, Value=0, Initialized=true
```

**发现异常**：setState 设置 value=-1 后，下一次渲染时 GetOrCreateHook 返回的仍是 Value=0！

#### 第四步：追踪多个 setState 调用
发现**两个按钮都被点击了**：

```
Button HandleEvent: triggering onClick for label="  [ - ]  " (Special=2)
setState BEFORE: value=0
setState AFTER: value=-1

Button HandleEvent: triggering onClick for label="  [ + ]  " (Special=2)
setState BEFORE: value=-1
setState AFTER: value=0
```

第一次点击 [-]：0 → -1
第二次点击 [+]：-1 → 0

最终结果：0（被第二个按钮抵消了）

---

## 2. 根本原因分析

### 2.1 事件分发问题

**问题代码** (`internal/render/declarative_node.go`):

```go
for _, child := range children {
    if n.distributeEventToVNode(child, ev) {
        handled = true
        // Don't break - allow multiple components to handle the event
    }
}
```

一个 Enter 键事件被分发给所有按钮，导致所有按钮的 onClick 都被触发。

**修复**:

```go
for _, child := range children {
    if n.distributeEventToVNode(child, ev) {
        // Event was handled by a child - stop propagation
        return true
    }
}
```

### 2.2 焦点管理缺失

测试期望：
1. Tab 键将焦点从 [-] 按钮移到 [+] 按钮
2. Enter 键只触发当前焦点按钮

实际行为：
- Tab 键没有被焦点系统处理
- Enter 键被分发给所有按钮（现已修复为只触发一个）
- 默认情况下，[-] 按钮首先接收事件

---

## 3. 状态管理验证

修复事件分发后，验证状态管理：

```
初始状态: Count: 0
点击 [-]: setState(0 → -1) ✓
重新渲染: Count: -1 ✓
```

**结论**: 状态管理系统工作正常！

- `useState` 正确创建/获取 hook
- `setState` 正确修改 hook value
- `DeclarativeNode.Paint` 正确调用 renderFn() 重新创建 VNode 树
- 新的 VNode 树包含更新后的状态值

---

## 4. 现有架构分析

### 4.1 旧架构 FocusManager

位置: `runtime/focus.go`

```go
type FocusManager struct {
    focusable      []*FocusableItem
    currentIndex   int
    onFocusChange  func(focused, previous *FocusableItem)
    // ...
}

type FocusableItem struct {
    ID       string
    Node     *LayoutNode          // ← 旧架构 LayoutNode
    Instance FocusableComponent   // ← 旧架构接口
}
```

依赖：
- `runtime.LayoutNode`
- `runtime.FocusableComponent` 接口

### 4.2 新架构 DeclarativeNode

位置: `internal/render/declarative_node.go`

```go
type DeclarativeNode struct {
    renderFn rtui.ComponentFunc     // 返回 VNode 树
    instance *rtui.ComponentContext   // Hook 管理
    root     rtui.VNode              // VNode 树根节点
}
```

特点：
- 纯声明式 VNode 树
- 组件来自 `components/` 目录（如 `ButtonVNode`）
- 不实现 `runtime.FocusableComponent` 接口

### 4.3 架构不匹配问题

| 方面 | 旧架构 (runtime) | 新架构 (declarative) |
|------|------------------|---------------------|
| 组件表示 | LayoutNode | VNode |
| 组件接口 | Component, FocusableComponent | VNode interface |
| 焦点收集 | 从 LayoutNode 树收集 | 无 |
| 事件分发 | 通过 Runtime.Dispatch | 通过 DeclarativeNode.HandleEvent |

---

## 5. 解决方案

### 5.1 事件分发修复 ✅

已修复：事件被处理后立即停止传播。

### 5.2 焦点管理集成方案

#### 方案 A: 为新组件添加 FocusableComponent 实现

修改组件实现旧接口：

```go
// components/button/button.go
func (b *ButtonVNode) SetFocus(focus bool) {
    b.hasFocus = focus
    // 触发重新渲染以显示焦点状态
}

func (b *ButtonVNode) IsFocusable() bool {
    return !b.disabled
}
```

#### 方案 B: 创建新的焦点系统

创建 `runtime/ui/` 下的新焦点系统：

```go
// runtime/ui/focus_manager.go (新)
type VNodeFocusManager struct {
    focusable []VNode  // 直接使用 VNode
    current   int
}

func (m *VNodeFocusManager) CollectFromVNode(root VNode) {
    // 遍历 VNode 树，收集可聚焦节点
}

func (m *VNodeFocusManager) HandleKeyEvent(ev KeyEvent) bool {
    if ev.Special == KeyTab {
        m.FocusNext()
        return true
    }
    if ev.Special == KeyEnter {
        // 只向当前焦点分发事件
        return m.DispatchToFocused(ev)
    }
    return false
}
```

#### 方案 C: 在 DeclarativeNode 中集成焦点管理

```go
// internal/render/declarative_node.go
type DeclarativeNode struct {
    renderFn rtui.ComponentFunc
    instance *rtui.ComponentContext
    root     rtui.VNode
    focusMgr *VNodeFocusManager  // 新增
}

func (n *DeclarativeNode) HandleEvent(ev frameworkevent.Event) bool {
    // 1. 焦点导航处理
    if n.focusMgr.HandleNavigation(ev) {
        return true
    }

    // 2. 向焦点元素分发事件
    if focused := n.focusMgr.GetCurrent(); focused != nil {
        if n.dispatchToVNode(focused, ev) {
            return true
        }
    }

    // 3. 回退到全局分发
    return n.distributeEventToVNode(n.root, ev)
}
```

---

## 6. 推荐实现路径

### Phase 1: 创建 VNode 可聚焦接口 (短期)

```go
// runtime/ui/focusable.go
type FocusableVNode interface {
    VNode
    SetFocus(bool)
    IsFocusable() bool
    GetFocusID() string
}
```

### Phase 2: 为现有组件添加焦点支持

```go
// components/button/button.go
func (b *ButtonVNode) SetFocus(focus bool) {
    b.hasFocus = focus
}

func (b *ButtonVNode) IsFocusable() bool {
    return !b.disabled
}

func (b *ButtonVNode) GetFocusID() string {
    if b.Key() != "" {
        return b.Key()
    }
    return fmt.Sprintf("button-%p", b)
}
```

### Phase 3: 集成到 DeclarativeNode

在 `Paint()` 中收集可聚焦节点，在 `HandleEvent()` 中处理焦点导航。

---

## 7. 遗留问题

1. **Tab 键导航** - 需要实现焦点在按钮间移动
2. **焦点状态显示** - 当前焦点按钮需要视觉反馈（如下划线或边框）
3. **组件 ID 系统** - 焦点管理需要稳定的组件标识
4. **FocusManager 与声明式架构的桥接** - 需要统一新旧两套系统

---

## 8. 验证测试

### 测试用例

```bash
# 运行计数器测试
TUI_DEBUG_UI=true go test -v ./examples/sandbox/demo/... -run "TestCounterWithRunTest"

# 预期输出（修复后）:
# - Tab 键移动焦点
# - Enter 只触发当前焦点按钮
# - Count 值正确更新
```

### 当前状态

- ✅ 事件分发修复 - 一个事件只触发一个按钮
- ✅ 状态持久化验证通过 - setState → re-render → 新值显示
- ⏳ 焦点管理待集成 - Tab 键需要移动焦点
- ⏳ 视觉焦点状态待实现 - 需要显示当前焦点元素

---

## 9. 相关文件

| 文件 | 修改 |
|------|------|
| `internal/render/declarative_node.go` | 修复事件分发逻辑 |
| `runtime/ui/hooks.go` | 添加调试输出 |
| `ui/hooks.go` | 添加调试输出 |
| `components/button/button.go` | 添加键盘事件处理 |
| `runtime/focus.go` | 现有焦点管理器（待集成） |
| `ui/test.go` | 测试框架 |

---

## 10. 总结

本次调试主要解决了一个**事件分发错误**，这个问题导致所有按钮同时响应键盘事件。

状态管理本身是正常工作的，问题出在事件系统。

下一步重点是**集成焦点管理**到声明式架构中，这需要：
1. 为新组件添加可聚焦接口实现
2. 在 VNode 树中收集可聚焦节点
3. 处理 Tab 键导航
4. 只向焦点元素分发键盘事件
