# FocusManager 集成到声明式架构方案

**日期**: 2026-02-02
**目标**: 将现有 `runtime.FocusManager` 集成到新的声明式 VNode 架构

---

## 1. 现状分析

### 1.1 现有 FocusManager

**位置**: `runtime/focus.go`

**接口依赖**:
```go
type FocusableItem struct {
    ID       string
    Node     *LayoutNode          // ← 问题：依赖旧 LayoutNode
    Instance FocusableComponent   // ← 问题：依赖旧 Component 接口
}
```

**工作流程**:
1. 从 `LayoutResult` 收集可聚焦组件
2. 维护焦点索引
3. 处理 `FocusNext()`, `FocusPrev()` 等操作
4. 调用组件的 `SetFocus(bool)` 方法

### 1.2 新架构组件

**位置**: `components/` 目录

**特点**:
- 实现 `VNode` 接口
- 实现 `Measurable` 和 `Paintable` 接口
- **不实现** `runtime.FocusableComponent` 接口
- 不在 LayoutNode 树中，而在 VNode 树中

---

## 2. 集成方案

### 2.1 创建新的可聚焦接口

```go
// runtime/ui/focusable.go
package ui

// FocusableVNode 是 VNode 的可聚焦扩展接口
type FocusableVNode interface {
    VNode

    // SetFocus 设置焦点状态
    SetFocus(hasFocus bool)

    // IsFocusable 返回是否可以接收焦点
    IsFocusable() bool

    // GetFocusID 返回焦点标识符（用于焦点持久化）
    GetFocusID() string
}
```

### 2.2 让组件实现新接口

```go
// components/button/button.go
func (b *ButtonVNode) SetFocus(hasFocus bool) *ButtonVNode {
    b.hasFocus = hasFocus
    // 触发重新渲染
    return b
}

func (b *ButtonVNode) IsFocusable() bool {
    return !b.disabled
}

func (b *ButtonVNode) GetFocusID() string {
    if b.Key() != "" {
        return b.Key()
    }
    // 为没有 key 的按钮生成稳定 ID
    return fmt.Sprintf("btn-%s-%p", b.label, b)
}
```

### 2.3 创建 VNode 焦点收集器

```go
// runtime/ui/vnode_focus_collector.go
package ui

// CollectFocusable 从 VNode 树中收集所有可聚焦节点
func CollectFocusable(root VNode) []FocusableVNode {
    return collectFocusableRecursive(root, nil)
}

func collectFocusableRecursive(node VNode, parent VNode) []FocusableVNode {
    var result []FocusableVNode

    if node == nil {
        return result
    }

    // 检查当前节点是否可聚焦
    if focusable, ok := node.(FocusableVNode); ok {
        if focusable.IsFocusable() {
            result = append(result, focusable)
        }
    }

    // 递归检查子节点
    for _, child := range node.Children() {
        childFocusable := collectFocusableRecursive(child, node)
        result = append(result, childFocusable...)
    }

    return result
}
```

### 2.4 创建声明式焦点管理器

```go
// runtime/ui/focus_manager.go
package ui

import frameworkevent "github.com/wwsheng009/mint/framework/event"

// VNodeFocusManager 管理 VNode 树中的焦点
type VNodeFocusManager struct {
    focusable   []FocusableVNode
    current     int
    onNavigate  func(from, to FocusableVNode) // 焦点变化回调
}

// NewVNodeFocusManager 创建焦点管理器
func NewVNodeFocusManager() *VNodeFocusManager {
    return &VNodeFocusManager{
        focusable: []FocusableVNode{},
        current:   -1,
    }
}

// SetFocusable 设置可聚焦节点列表
func (m *VNodeFocusManager) SetFocusable(nodes []FocusableVNode) {
    // 保存当前焦点 ID
    currentID := ""
    if m.current >= 0 && m.current < len(m.focusable) {
        currentID = m.focusable[m.current].GetFocusID()
    }

    m.focusable = nodes
    m.current = -1

    // 尝试恢复焦点
    if currentID != "" {
        for i, node := range m.focusable {
            if node.GetFocusID() == currentID {
                m.current = i
                break
            }
        }
    }

    // 如果没有焦点且存在可聚焦节点，聚焦第一个
    if m.current < 0 && len(m.focusable) > 0 {
        m.FocusFirst()
    }
}

// FocusNext 聚焦下一个节点
func (m *VNodeFocusManager) FocusNext() bool {
    if len(m.focusable) == 0 {
        return false
    }

    old := m.current
    m.current = (m.current + 1) % len(m.focusable)

    // 更新焦点状态
    if old >= 0 {
        m.focusable[old].SetFocus(false)
    }
    m.focusable[m.current].SetFocus(true)

    if m.onNavigate != nil {
        m.onNavigate(
            getOrNil(m.focusable, old),
            m.focusable[m.current],
        )
    }

    return true
}

// FocusPrev 聚焦上一个节点
func (m *VNodeFocusManager) FocusPrev() bool {
    if len(m.focusable) == 0 {
        return false
    }

    old := m.current
    m.current = m.current - 1
    if m.current < 0 {
        m.current = len(m.focusable) - 1
    }

    // 更新焦点状态
    if old >= 0 {
        m.focusable[old].SetFocus(false)
    }
    m.focusable[m.current].SetFocus(true)

    if m.onNavigate != nil {
        m.onNavigate(
            getOrNil(m.focusable, old),
            m.focusable[m.current],
        )
    }

    return true
}

// HandleEvent 处理键盘事件
func (m *VNodeFocusManager) HandleEvent(ev frameworkevent.Event) (handled bool, shouldRender bool) {
    // 检查是否是 Tab 键（焦点导航）
    if keyEvent, ok := ev.(*event.KeyEvent); ok {
        if keyEvent.Special == event.KeyTab {
            if m.FocusNext() {
                return true, true  // 处理了事件，需要重新渲染
            }
        }
    }

    // 检查是否是 Enter/Space（向焦点元素分发）
    if m.current >= 0 && m.current < len(m.focusable) {
        // 让焦点元素自己处理
        if component, ok := m.focusable[m.current].(frameworkevent.Component); ok {
            if component.HandleEvent(ev) {
                return true, true
            }
        }
    }

    return false, false
}

// GetCurrent 返回当前焦点节点
func (m *VNodeFocusManager) GetCurrent() FocusableVNode {
    if m.current < 0 || m.current >= len(m.focusable) {
        return nil
    }
    return m.focusable[m.current]
}

// FocusFirst 聚焦第一个节点
func (m *VNodeFocusManager) FocusFirst() bool {
    if len(m.focusable) == 0 {
        return false
    }

    m.current = 0
    m.focusable[0].SetFocus(true)

    if m.onNavigate != nil {
        m.onNavigate(nil, m.focusable[0])
    }

    return true
}

func getOrNil[T any](slice []T, index int) T {
    if index < 0 || index >= len(slice) {
        var zero T
        return zero
    }
    return slice[index]
}
```

### 2.5 集成到 DeclarativeNode

```go
// internal/render/declarative_node.go
type DeclarativeNode struct {
    mu        sync.RWMutex
    root      rtui.VNode
    renderFn  rtui.ComponentFunc
    instance  *rtui.ComponentContext
    focusMgr  *ui.VNodeFocusManager  // 新增：VNode 焦点管理器
}

func NewDeclarativeNodeFromFunc(fn rtui.ComponentFunc) *DeclarativeNode {
    return &DeclarativeNode{
        renderFn: fn,
        instance: rtui.NewComponentContextForRoot(),
        focusMgr: ui.NewVNodeFocusManager(),
    }
}

func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    n.mu.Lock()
    defer n.mu.Unlock()

    // ... 现有的 renderFn 调用 ...

    // 收集可聚焦节点
    if n.root != nil {
        focusable := ui.CollectFocusable(n.root)
        n.focusMgr.SetFocusable(focusable)
    }

    // ... 现有的 paintVNode 调用 ...
}

func (n *DeclarativeNode) HandleEvent(ev frameworkevent.Event) bool {
    n.mu.RLock()
    defer n.mu.RUnlock()

    // 1. 让焦点管理器先处理（Tab 导航、Enter 分发）
    handled, shouldRender := n.focusMgr.HandleEvent(ev)
    if handled {
        // 如果需要重新渲染（焦点变化），标记 dirty
        if shouldRender {
            // TODO: 触发重新渲染
        }
        return true
    }

    // 2. 回退到全局事件分发
    return n.distributeEventToVNode(n.root, ev)
}
```

---

## 3. 实现优先级

### P0 - 核心功能
1. 创建 `FocusableVNode` 接口
2. Button 组件实现该接口
3. 创建 `VNodeFocusManager`
4. 实现 Tab 键导航
5. 实现 Enter 键只触发焦点按钮

### P1 - 完善功能
1. 其他组件实现可聚焦接口
2. 焦点状态视觉反馈
3. Shift+Tab 反向导航
4. 焦点持久化（re-render 后保持）

### P2 - 高级功能
1. 可配置的焦点导航策略
2. 自定义焦点顺序
3. 焦点陷阱/逃逸
4. 多区域焦点管理

---

## 4. 组件修改清单

| 组件 | 需要添加的方法 | 说明 |
|------|----------------|------|
| ButtonVNode | SetFocus, IsFocusable, GetFocusID | 已有 disabled, label |
| InputVNode | SetFocus, IsFocusable, GetFocusID | 需要实现 |
| CheckboxVNode | SetFocus, IsFocusable, GetFocusID | 需要实现 |
| SelectVNode | SetFocus, IsFocusable, GetFocusID | 需要实现 |
| TextVNode | - | 文本不可聚焦 |

---

## 5. 测试计划

```go
func TestFocusNavigation(t *testing.T) {
    testApp, _ := ui.RunTest(Counter, ui.WithSize(40, 18))
    defer testApp.Close()

    // 初始状态：第一个按钮应有焦点
    // ...

    // 按 Tab：焦点移动到第二个按钮
    testApp.InjectSpecialKey(platform.KeyTab)
    // ...

    // 按 Enter：只触发第二个按钮
    testApp.InjectSpecialKey(platform.KeyEnter)
    // ...
}
```

---

## 6. 与旧系统的兼容性

### 保留旧 FocusManager

对于仍在使用旧 `LayoutNode` 的代码，保留 `runtime/focus.go`：

```go
// runtime/focus.go 保留不变
// 用于 LayoutNode 架构的组件
```

### 创建新 VNode 焦点系统

```go
// runtime/ui/focus_manager.go (新建)
// 用于 VNode 架构的组件
```

### 两者共存

```go
// framework/app.go 或相应位置
type HybridFocusManager struct {
    layoutFocus *runtime.FocusManager   // 旧系统
    vnodeFocus  *ui.VNodeFocusManager   // 新系统
}
```

---

## 7. API 设计示例

```go
// 使用示例
func Counter() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.HStack(
            button.New("-").
                OnClick(func() { setCount(-1) }).
                FocusKey("btn-decrement"),  // 可选：指定焦点 key
            button.New("+").
                OnClick(func() { setCount(1) }).
                FocusKey("btn-increment"),
        ),
    )
}
```

---

## 8. 注意事项

1. **线程安全**: 焦点状态变化可能需要加锁保护
2. **性能**: 每次渲染都收集可聚焦节点可能影响性能
3. **ID 唯一性**: 确保焦点 ID 在整个应用中唯一
4. **边界情况**: 禁用按钮不应获得焦点
5. **回退机制**: 当没有可聚焦节点时的行为
