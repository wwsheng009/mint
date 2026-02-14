# Phase 2: 组件迁移

## 目标

将现有组件从旧接口（Updater/EventHandler）迁移到 ActionTarget 接口。

## 迁移策略

### 分批迁移

1. **基础组件**：Button, Text, TextInput
2. **容器组件**：Container, VStack, HStack, List
3. **高级组件**：Table, Tree, Form
4. **内部组件**：Inspector, 调试工具

### 迁移模式

```go
// 模式 A：直接实现 ActionTarget（推荐）
type Button struct {
    // ...
}

func (b *Button) HandleAction(action *Action) bool {
    switch action.Type {
    case ActionClick:
        if b.onClick != nil {
            b.onClick()
        }
        return true
    }
    return false
}

func (b *Button) GetSupportedActions() []ActionType {
    return []ActionType{ActionClick, ActionHover}
}

func (b *Button) CanHandleAction(action *Action) bool {
    return action.Type == ActionClick || action.Type == ActionHover
}

// 模式 B：嵌入 BaseActionTarget（简化版）
type Text struct {
    action.BaseActionTarget  // 嵌入基类
    // ...
}

func (t *Text) HandleAction(action *Action) bool {
    // 只需要实现 HandleAction
    return false  // Text 不处理任何 Action
}

// 模式 C：使用适配器（过渡期）
// 参见 Phase 1 的适配器定义
```

## 1. 基础组件迁移

### 1.1 Button 组件

```go
// framework/components/button.go

// 旧代码
type Button struct {
    text    string
    onClick func()
}

func (b *Button) HandleEvent(ev frameworkevent.Event) bool {
    if ev.Type() == frameworkevent.EventClick {
        if b.onClick != nil {
            b.onClick()
        }
        return true
    }
    return false
}

// 新代码
type Button struct {
    text    string
    onClick func()
    nodeID  uint64
}

func (b *Button) HandleAction(action *action.Action) bool {
    // 检查目标
    if action.TargetID != 0 && action.TargetID != b.nodeID {
        return false
    }

    switch action.Type {
    case action.ActionClick:
        if b.onClick != nil {
            b.onClick()
        }
        return true

    case action.ActionHover:
        // 处理悬停状态
        b.hovered = true
        return true

    case action.ActionKeyPress:
        // 支持 Enter/Space 触发点击
        if payload, ok := action.GetPayloadString(); ok {
            if payload == "enter" || payload == " " {
                if b.onClick != nil {
                    b.onClick()
                }
                return true
            }
        }
    }

    return false
}

func (b *Button) GetSupportedActions() []action.ActionType {
    return []action.ActionType{
        action.ActionClick,
        action.ActionHover,
        action.ActionKeyPress,
    }
}

func (b *Button) CanHandleAction(action *action.Action) bool {
    switch action.Type {
    case action.ActionClick, action.ActionHover:
        return true
    case action.ActionKeyPress:
        return true
    }
    return false
}
```

### 1.2 TextInput 组件

```go
// framework/components/textinput.go

type TextInput struct {
    value       string
    cursor      int
    placeholder string
    focused     bool
    nodeID      uint64

    onChange    func(string)
    onSubmit    func(string)
}

func (t *TextInput) HandleAction(action *action.Action) bool {
    if action.TargetID != 0 && action.TargetID != t.nodeID {
        return false
    }

    switch action.Type {
    case action.ActionInputText:
        // 输入字符
        if payload, ok := action.GetPayloadString(); ok {
            t.insertChar(payload)
            t.notifyChange()
        }
        return true

    case action.ActionBackspace:
        t.deleteChar(-1)
        t.notifyChange()
        return true

    case action.ActionDeleteChar:
        t.deleteChar(1)
        t.notifyChange()
        return true

    case action.ActionNavigateLeft:
        if t.cursor > 0 {
            t.cursor--
        }
        return true

    case action.ActionNavigateRight:
        if t.cursor < len(t.value) {
            t.cursor++
        }
        return true

    case action.ActionNavigateHome:
        t.cursor = 0
        return true

    case action.ActionNavigateEnd:
        t.cursor = len(t.value)
        return true

    case action.ActionEnter:
        if t.onSubmit != nil {
            t.onSubmit(t.value)
        }
        return true

    case action.ActionFocus:
        t.focused = true
        return true

    case action.ActionBlur:
        t.focused = false
        return true

    case action.ActionClick:
        // 点击设置光标位置
        if x, _, ok := action.GetPayloadPoint(); ok {
            t.cursor = x
            t.focused = true
        }
        return true
    }

    return false
}

func (t *TextInput) GetSupportedActions() []action.ActionType {
    return []action.ActionType{
        action.ActionInputText,
        action.ActionBackspace,
        action.ActionDeleteChar,
        action.ActionNavigateLeft,
        action.ActionNavigateRight,
        action.ActionNavigateHome,
        action.ActionNavigateEnd,
        action.ActionEnter,
        action.ActionFocus,
        action.ActionBlur,
        action.ActionClick,
    }
}

func (t *TextInput) CanHandleAction(action *action.Action) bool {
    return t.focused || action.Type == action.ActionClick || action.Type == action.ActionFocus
}

// 辅助方法
func (t *TextInput) insertChar(char string) {
    t.value = t.value[:t.cursor] + char + t.value[t.cursor:]
    t.cursor += len(char)
}

func (t *TextInput) deleteChar(dir int) {
    if dir < 0 && t.cursor > 0 {
        t.value = t.value[:t.cursor-1] + t.value[t.cursor:]
        t.cursor--
    } else if dir > 0 && t.cursor < len(t.value) {
        t.value = t.value[:t.cursor] + t.value[t.cursor+1:]
    }
}

func (t *TextInput) notifyChange() {
    if t.onChange != nil {
        t.onChange(t.value)
    }
}
```

### 1.3 List 组件

```go
// framework/components/list.go

type List struct {
    items      []ListItem
    selected   int
    scrollTop  int
    nodeID     uint64

    onSelect   func(int)
}

func (l *List) HandleAction(action *action.Action) bool {
    switch action.Type {
    case action.ActionNavigateUp:
        if l.selected > 0 {
            l.selected--
            l.ensureVisible()
        }
        return true

    case action.ActionNavigateDown:
        if l.selected < len(l.items)-1 {
            l.selected++
            l.ensureVisible()
        }
        return true

    case action.ActionNavigatePageUp:
        l.selected -= 10
        if l.selected < 0 {
            l.selected = 0
        }
        l.ensureVisible()
        return true

    case action.ActionNavigatePageDown:
        l.selected += 10
        if l.selected >= len(l.items) {
            l.selected = len(l.items) - 1
        }
        l.ensureVisible()
        return true

    case action.ActionNavigateHome:
        l.selected = 0
        l.scrollTop = 0
        return true

    case action.ActionNavigateEnd:
        l.selected = len(l.items) - 1
        l.ensureVisible()
        return true

    case action.ActionSelect:
        if l.onSelect != nil {
            l.onSelect(l.selected)
        }
        return true

    case action.ActionClick:
        // 点击选择项
        if y, _, ok := action.GetPayloadPoint(); ok {
            idx := l.scrollTop + y
            if idx >= 0 && idx < len(l.items) {
                l.selected = idx
                if l.onSelect != nil {
                    l.onSelect(l.selected)
                }
            }
        }
        return true

    case action.ActionScroll:
        if delta, ok := action.GetPayloadInt(); ok {
            l.scrollTop += delta
            l.clampScroll()
        }
        return true
    }

    return false
}

func (l *List) GetSupportedActions() []action.ActionType {
    return []action.ActionType{
        action.ActionNavigateUp,
        action.ActionNavigateDown,
        action.ActionNavigatePageUp,
        action.ActionNavigatePageDown,
        action.ActionNavigateHome,
        action.ActionNavigateEnd,
        action.ActionSelect,
        action.ActionClick,
        action.ActionScroll,
    }
}

func (l *List) CanHandleAction(action *action.Action) bool {
    return true  // List 始终可以处理导航和选择
}

func (l *List) ensureVisible() {
    // 确保选中项可见
    // ...
}
```

## 2. 容器组件迁移

### 2.1 Container 组件

```go
// framework/components/container.go

type Container struct {
    children []component.Node
    nodeID   uint64

    // 容器通常不处理 Action，只传递给子组件
}

func (c *Container) HandleAction(action *action.Action) bool {
    // 容器本身不处理 Action
    // Action 由 Router 通过 Target Phase 传递给正确的子组件
    return false
}

func (c *Container) GetSupportedActions() []action.ActionType {
    return nil  // 容器不处理任何 Action
}

func (c *Container) CanHandleAction(action *action.Action) bool {
    return false
}
```

### 2.2 FocusGroup 组件

```go
// framework/components/focusgroup.go

// FocusGroup 管理一组可聚焦的子组件
type FocusGroup struct {
    children    []component.Node
    focusIndex  int
    nodeID      uint64
}

func (f *FocusGroup) HandleAction(action *action.Action) bool {
    switch action.Type {
    case action.ActionFocusNext:
    case action.ActionNavigateNext:
    case action.ActionTab:  // 如果有这个 Action
        f.focusNext()
        return true

    case action.ActionFocusPrev:
    case action.ActionNavigatePrev:
        f.focusPrev()
        return true
    }

    return false
}

func (f *FocusGroup) GetSupportedActions() []action.ActionType {
    return []action.ActionType{
        action.ActionFocusNext,
        action.ActionFocusPrev,
    }
}

func (f *FocusGroup) CanHandleAction(action *action.Action) bool {
    return action.Type == action.ActionFocusNext || action.Type == action.ActionFocusPrev
}

func (f *FocusGroup) focusNext() {
    if len(f.children) == 0 {
        return
    }

    // 移除当前焦点
    f.blurCurrent()

    // 移动到下一个
    f.focusIndex = (f.focusIndex + 1) % len(f.children)

    // 设置新焦点
    f.focusCurrent()
}

func (f *FocusGroup) focusPrev() {
    if len(f.children) == 0 {
        return
    }

    f.blurCurrent()
    f.focusIndex--
    if f.focusIndex < 0 {
        f.focusIndex = len(f.children) - 1
    }
    f.focusCurrent()
}

func (f *FocusGroup) blurCurrent() {
    if focusable, ok := f.children[f.focusIndex].(action.FocusableActionTarget); ok {
        focusable.Blur()
    }
}

func (f *FocusGroup) focusCurrent() {
    if focusable, ok := f.children[f.focusIndex].(action.FocusableActionTarget); ok {
        focusable.Focus()
    }
}
```

## 3. 迁移检查清单

### 每个组件迁移步骤

1. [ ] 添加 `nodeID` 字段（如果还没有）
2. [ ] 实现 `HandleAction(action *Action) bool`
3. [ ] 实现 `GetSupportedActions() []ActionType`
4. [ ] 实现 `CanHandleAction(action *Action) bool`
5. [ ] 删除旧的 `HandleEvent` 或 `Update` 方法
6. [ ] 更新测试用例
7. [ ] 更新文档/注释

### 组件迁移优先级

| 优先级 | 组件 | 状态 |
|--------|------|------|
| P0 | Button | [ ] 待迁移 |
| P0 | TextInput | [ ] 待迁移 |
| P0 | List | [ ] 待迁移 |
| P1 | Checkbox | [ ] 待迁移 |
| P1 | Select | [ ] 待迁移 |
| P1 | Table | [ ] 待迁移 |
| P2 | Tree | [ ] 待迁移 |
| P2 | Form | [ ] 待迁移 |
| P3 | Inspector | [ ] 待迁移 |

## 4. 测试策略

### 4.1 单元测试模板

```go
// framework/components/button_test.go

func TestButtonHandleAction_Click(t *testing.T) {
    clicked := false
    btn := &Button{
        text: "Click Me",
        onClick: func() {
            clicked = true
        },
    }

    action := action.NewAction(action.ActionClick)
    handled := btn.HandleAction(action)

    assert.True(t, handled)
    assert.True(t, clicked)
}

func TestButtonHandleAction_Hover(t *testing.T) {
    btn := &Button{text: "Hover Me"}

    action := action.NewAction(action.ActionHover)
    handled := btn.HandleAction(action)

    assert.True(t, handled)
    assert.True(t, btn.hovered)
}

func TestButtonHandleAction_IgnoreOthers(t *testing.T) {
    btn := &Button{text: "Test"}

    action := action.NewAction(action.ActionNavigateUp)
    handled := btn.HandleAction(action)

    assert.False(t, handled)
}

func TestButtonGetSupportedActions(t *testing.T) {
    btn := &Button{text: "Test"}
    actions := btn.GetSupportedActions()

    assert.Contains(t, actions, action.ActionClick)
    assert.Contains(t, actions, action.ActionHover)
}

func TestButtonCanHandleAction(t *testing.T) {
    btn := &Button{text: "Test"}

    assert.True(t, btn.CanHandleAction(action.NewAction(action.ActionClick)))
    assert.True(t, btn.CanHandleAction(action.NewAction(action.ActionHover)))
    assert.False(t, btn.CanHandleAction(action.NewAction(action.ActionNavigateUp)))
}
```

### 4.2 集成测试

```go
// framework/integration/action_test.go

func TestActionEndToEnd(t *testing.T) {
    // 创建应用
    app := NewApp()
    app.legacyMode = false  // 禁用兼容模式

    // 创建组件树
    btn := NewButton("Click", func() {
        // 点击回调
    })
    list := NewList([]string{"A", "B", "C"})
    container := NewContainer(btn, list)
    app.SetRoot(container)

    // 初始化
    app.Init()

    // 模拟键盘事件
    keyMsg := runtimemsg.NewKeyMsg(rune('a'), 0, nil)
    app.processMsg(keyMsg)

    // 验证 Action 被正确处理
    // ...

    // 模拟鼠标点击
    mouseMsg := runtimemsg.NewMouseMsg(
        runtimemsg.MouseActionPress,
        runtimemsg.MouseLeft,
        10, 5,  // 坐标
        0,      // delta
    )
    app.processMsg(mouseMsg)

    // 验证点击被处理
    // ...
}
```

## 5. 回滚方案

如果某个组件迁移出现问题：

1. 临时使用适配器包装
2. 设置 `legacyMode = true` 回退到旧路径
3. 修复问题后重新迁移

```go
// 临时回滚：使用适配器
func (b *Button) HandleEvent(ev frameworkevent.Event) bool {
    // 保留旧的实现作为备份
}

// 在注册时使用适配器
adapter := action.NewEventHandlerAdapter(btn, btn.nodeID)
app.actionRegistry.Register(btn.nodeID, adapter)
```

## 6. Phase 2 完成标准

- [ ] 所有 P0 组件迁移完成
- [ ] 所有 P1 组件迁移完成
- [ ] 单元测试覆盖率 >= 80%
- [ ] 集成测试通过
- [ ] 手动测试通过
- [ ] 文档更新完成
