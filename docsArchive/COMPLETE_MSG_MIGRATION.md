# Mint TUI 事件系统 - Msg/Cmd 彻底迁移方案

> **版本**: v3.0 (彻底重构版)
> **日期**: 2026-02-10
> **原则**: 完全切换到 Msg/Cmd 系统，移除所有 HandleEvent 兼容代码

---

## 🎯 核心目标

**问题**：当前系统混用 Event 和 Msg 两套接口，架构不统一

**解决方案**：
1. ✅ **完全移除** `HandleEvent(Event)` 接口
2. ✅ **强制使用** `Update(Msg) cmd.Cmd` 接口
3. ✅ **Pump 直接输出** Msg 而不是 Event
4. ✅ **Router 只处理** Msg
5. ✅ **所有组件** 实现 Update(Msg)

---

## 📊 当前架构问题

### 🔴 混乱的现状

```
Pump → Event (MouseEvent/KeyEvent)
  ↓
Router → Dispatch(Event)
  ↓
组件 → HandleEvent(Event) ❌ 老接口
组件 → Update(Msg) ✅ 新接口
两套接口并存！架构混乱！
```

### 🎯 目标架构

```
Pump → Msg (MouseMsg/KeyMsg) ✅ 统一
  ↓
Router → Dispatch(Msg) ✅ 统一
  ↓
组件 → Update(Msg) ✅ 统一
单一接口，架构清晰！
```

---

## 📝 彻底迁移方案

### Phase 1: Pump 直接输出 Msg（1天）

#### 1.1 修改 Pump 创建 Msg

**文件**: `framework/event/pump.go`

```go
// framework/event/pump.go

// ❌ 删除老代码
/*
func (p *Pump) convertToMouseEvent(raw *platform.RawInput) *MouseEvent {
    ev := &MouseEvent{...}
    return ev
}
*/

// ✅ 新代码：直接创建 MouseMsg
func (p *Pump) convertToMouseMsg(raw *platform.RawInput) *msg.MouseMsg {
    // Phase 1-6: HitTest 填充 TargetID/LocalX/LocalY
    p.hitMapMu.RLock()
    hitMap := p.hitMap
    p.hitMapMu.RUnlock()

    var targetID string
    var localX, localY int

    if hitMap != nil {
        entry := hitMap.HitTest(raw.MouseX, raw.MouseY)
        if entry != nil {
            targetID = entry.NodeID
            localX, localY = entry.LocalXY(raw.MouseX, raw.MouseY)
        }
    }

    // 转换按钮和动作
    button := convertMouseButton(raw.MouseButton)
    action := convertMouseAction(raw.MouseAction)

    // ✅ 直接返回 MouseMsg，不是 MouseEvent！
    return &msg.MouseMsg{
        X:        raw.MouseX,
        Y:        raw.MouseY,
        LocalX:   localX,
        LocalY:   localY,
        TargetID: targetID,
        Button:   button,
        Action:   action,
        Delta:    calculateWheelDelta(raw.MouseAction),
    }
}

func (p *Pump) convertToKeyMsg(raw *platform.RawInput) *msg.KeyMsg {
    return &msg.KeyMsg{
        Rune:    rune(raw.Key),
        Special: raw.SpecialKey,
        Mod: msg.Modifiers{
            Alt:   raw.ModAlt,
            Ctrl:  raw.ModCtrl,
            Shift: raw.ModShift,
        },
    }
}

// ✅ Pump.Events() 返回 Msg channel
func (p *Pump) Events() <-chan msg.Msg {
    return p.events
}

// ✅ 内部 channel 改为 Msg
type Pump struct {
    events chan msg.Msg  // 改为 Msg，不是 Event
    // ...
}
```

#### 1.2 删除 Event → Msg 适配器

**文件**: `framework/msg/adapter.go`

```go
// ❌ 删除这个文件中的适配器函数
// 不再需要 Event → Msg 转换！
/*
func MouseEventToMsg(mouseEvent *MouseEvent) Msg { ... }
func KeyEventToMsg(keyEvent *KeyEvent) Msg { ... }
*/
```

**原因**：Pump 直接创建 Msg，不需要适配器。

---

### Phase 2: Router 只处理 Msg（1天）

#### 2.1 修改 Router 接口

**文件**: `framework/event/router.go`

```go
// framework/event/router.go

// ❌ 删除老接口
/*
func (r *Router) Dispatch(ev Event) bool { ... }
*/

// ✅ 新接口：只处理 Msg
func (r *Router) Dispatch(message msg.Msg) bool {
    // 根据 Msg 类型分发
    switch m := message.(type) {
    case *msg.MouseMsg:
        return r.dispatchMouse(m)
    case *msg.KeyMsg:
        return r.dispatchKey(m)
    default:
        return false
    }
}

// dispatchMouse 根据 TargetID 直接分发
func (r *Router) dispatchMouse(mouseMsg *msg.MouseMsg) bool {
    // 如果有 TargetID，直接分发
    if mouseMsg.TargetID != "" {
        return r.dispatchToTarget(mouseMsg.TargetID, mouseMsg)
    }

    // 否则回退到焦点系统
    return r.dispatchToFocused(mouseMsg)
}

// dispatchToTarget 根据 ID 直接分发到组件
func (r *Router) dispatchToTarget(targetID string, message msg.Msg) bool {
    // 从注册表查找组件
    component := r.registry.Lookup(targetID)
    if component == nil {
        return false
    }

    // ✅ 所有组件都必须实现 Update(Msg)
    if updater, ok := component.(component.Updater); ok {
        cmd := updater.Update(message)
        if cmd != nil {
            r.app.Execute(cmd)
        }
        return true
    }

    // ❌ 如果组件没实现 Update(Msg)，这是严重错误！
    panic(fmt.Sprintf("Component %s does not implement Update(Msg) interface", targetID))
}

// dispatchKey 分发键盘消息
func (r *Router) dispatchKey(keyMsg *msg.KeyMsg) bool {
    // 1. 优先检查全局快捷键
    if r.handleGlobalShortcut(keyMsg) {
        return true
    }

    // 2. 分发到焦点组件
    focused := r.focusManager.Focused()
    if focused != nil {
        if updater, ok := focused.(component.Updater); ok {
            cmd := updater.Update(keyMsg)
            if cmd != nil {
                r.app.Execute(cmd)
            }
            return true
        }
    }

    return false
}
```

#### 2.2 组件注册表

```go
// framework/event/registry.go (新建)

// Registry 维护 NodeID → Component 映射
type Registry struct {
    sync.RWMutex
    components map[string]component.Updater
}

func (r *Registry) Register(id string, component component.Updater) {
    r.Lock()
    defer r.Unlock()
    r.components[id] = component
}

func (r *Registry) Lookup(id string) component.Updater {
    r.RLock()
    defer r.RUnlock()
    return r.components[id]
}

func (r *Registry) Unregister(id string) {
    r.Lock()
    defer r.Unlock()
    delete(r.components, id)
}
```

---

### Phase 3: 移除 HandleEvent 接口（1天）

#### 3.1 删除 Component 接口

**文件**: `framework/event/event.go`

```go
// ❌ 删除老接口
/*
type Component interface {
    HandleEvent(Event) bool
}
*/

// ✅ 只保留 Updater 接口（在 component/updater.go）
/*
package component

type Updater interface {
    Update(message msg.Msg) cmd.Cmd
}
*/
```

#### 3.2 删除所有 HandleEvent 实现

**批量删除**：
```bash
# 查找所有 HandleEvent 实现
grep -r "func.*HandleEvent" framework/ components/ internal/

# 删除这些方法
```

**受影响的文件**：
- `components/container/panel.go` - Panel.HandleEvent
- `components/navigation/tabs.go` - Tabs.HandleEvent
- `components/tree/tree.go` - TreeView.HandleEvent
- `internal/inspector/standalone_inspector.go` - Inspector.HandleEvent
- 其他所有实现 HandleEvent 的组件

---

### Phase 4: 所有组件实现 Update(Msg)（2-3天）

#### 4.1 TreeView 实现

**文件**: `components/tree/tree.go`

```go
// components/tree/tree.go

// ✅ 实现 Update(Msg)
func (t *TreeView) Update(message msg.Msg) cmd.Cmd {
    switch m := message.(type) {
    case *msg.MouseMsg:
        return t.handleMouse(m)
    case *msg.KeyMsg:
        return t.handleKey(m)
    default:
        return cmd.None()
    }
}

func (t *TreeView) handleMouse(mouseMsg *msg.MouseMsg) cmd.Cmd {
    switch mouseMsg.Action {
    case msg.MouseActionPress:
        if mouseMsg.Button == msg.MouseLeft {
            return t.handleClick(mouseMsg.LocalX, mouseMsg.LocalY)
        }
    case msg.MouseActionWheel:
        return t.handleScroll(mouseMsg.Delta)
    }
    return cmd.None()
}

func (t *TreeView) handleClick(localX, localY int) cmd.Cmd {
    // ✅ 直接使用 LocalY 选中行
    if localY >= 0 && localY < len(t.visibleNodes) {
        t.selectedIndex = localY
        t.selectedNode = t.visibleNodes[localY]
        t.MarkDirty()

        // 可选：触发自定义事件
        return t.emitSelectionChanged()
    }
    return cmd.None()
}

func (t *TreeView) handleKey(keyMsg *msg.KeyMsg) cmd.Cmd {
    switch {
    case keyMsg.IsUp():
        return t.moveSelection(-1)
    case keyMsg.IsDown():
        return t.moveSelection(1)
    case keyMsg.IsEnter():
        return t.toggleExpanded()
    }
    return cmd.None()
}

// ❌ 删除老的 HandleEvent
/*
func (t *TreeView) HandleEvent(ev event.Event) bool { ... }
*/
```

#### 4.2 Tabs 实现

**文件**: `components/navigation/tabs.go`

```go
// components/navigation/tabs.go

func (t *TabsVNode) Update(message msg.Msg) cmd.Cmd {
    mouseMsg, ok := message.(*msg.MouseMsg)
    if !ok || !mouseMsg.IsClick() {
        return nil
    }

    // ✅ 使用 TargetID 识别点击区域
    switch {
    case mouseMsg.TargetID == t.tabBarID:
        // 点击了 tab 栏
        return t.handleTabBarClick(mouseMsg.LocalX)

    default:
        // Tab 内容区域的事件会直接分发到对应的 tab 内容组件
        // Tabs 不需要处理
        return nil
    }
}

func (t *TabsVNode) handleTabBarClick(localX int) cmd.Cmd {
    // 根据 localX 计算点击了哪个 tab
    clickedIndex := t.findTabByPosition(localX)
    if clickedIndex >= 0 && clickedIndex != t.activeTab {
        t.activeTab = clickedIndex
        t.MarkDirty()
        return t.emitTabChanged()
    }
    return nil
}
```

#### 4.3 Panel 简化

**文件**: `components/container/panel.go`

```go
// components/container/panel.go

// ❌ 删除 HandleEvent
/*
func (p *Panel) HandleEvent(ev event.Event) bool { ... }
*/

// ✅ Panel 不需要 Update，Router 根据 TargetID 直接分发到内容
// Panel 只负责布局和渲染，不处理事件
```

#### 4.4 Inspector 实现

**文件**: `internal/inspector/standalone_inspector.go`

```go
// internal/inspector/standalone_inspector.go

// ❌ 删除所有手动坐标解析代码
/*
func (si *StandaloneInspector) HandleEvent(ev event.Event) bool {
    // 几百行的手动坐标计算和事件转发代码
}
*/

// ✅ 实现 Update(Msg)
func (si *StandaloneInspector) Update(message msg.Msg) cmd.Cmd {
    switch m := message.(type) {
    case *msg.MouseMsg:
        return si.handleMouse(m)
    case *msg.KeyMsg:
        return si.handleKey(m)
    }
    return cmd.None()
}

func (si *StandaloneInspector) handleMouse(mouseMsg *msg.MouseMsg) cmd.Cmd {
    // ✅ Pump 已经设置好 TargetID 和 LocalX/LocalY
    // ✅ Router 根据 TargetID 直接分发到目标组件

    // Inspector 只处理捕获阶段的快捷键
    if mouseMsg.TargetID == "" {
        // 全局快捷键：按 F12 关闭 Inspector
        if keyMsg := si.extractKeyMsg(mouseMsg); keyMsg != nil && keyMsg.IsF12() {
            return si.toggle()
        }
    }

    // 其他事件直接传递，Router 会根据 TargetID 分发到子组件
    return nil
}
```

#### 4.5 Input 组件实现

**文件**: `components/input/input.go`

```go
// components/input/input.go

func (i *Input) Update(message msg.Msg) cmd.Cmd {
    switch m := message.(type) {
    case *msg.MouseMsg:
        if m.IsClick() {
            i.Focus()
            return cmd.None()
        }
    case *msg.KeyMsg:
        if !i.IsFocused() {
            return nil
        }

        switch {
        case keyMsg.IsEnter():
            return i.emitSubmit()
        case keyMsg.IsEscape():
            i.Blur()
            return cmd.None()
        case keyMsg.IsBackspace():
            i.backspace()
            i.MarkDirty()
            return cmd.None()
        case keyMsg.IsPrintable():
            i.insertRune(keyMsg.Rune)
            i.MarkDirty()
            return cmd.None()
        }
    }
    return cmd.None()
}
```

---

### Phase 5: App 主循环更新（1天）

#### 5.1 App.ProcessEvent 改为 ProcessMsg

**文件**: `framework/app.go`

```go
// framework/app.go

// ❌ 删除老方法
/*
func (a *App) ProcessEvent(ev event.Event) { ... }
*/

// ✅ 新方法：处理 Msg
func (a *App) ProcessMsg(message msg.Msg) {
    // Router 分发 Msg
    handled := a.router.Dispatch(message)

    if !handled {
        // 未处理的 Msg
        a.handleUnprocessedMsg(message)
    }
}

func (a *App) mainLoop() {
    for {
        select {
        case message := <-a.pump.Events():  // ✅ 接收 Msg
            a.ProcessMsg(message)

        case cmd := <-a.cmdChan:  // Cmd 执行结果
            a.ProcessCmd(cmd)

        case <-a.quit:
            return
        }
    }
}
```

#### 5.2 Cmd 执行

```go
// framework/app.go

func (a *App) Execute(cmd cmd.Cmd) {
    if cmd == nil {
        return
    }

    // 同步执行 Cmd
    result := cmd.Execute()

    // 如果 Cmd 产生了新 Msg，处理它
    if result != nil {
        a.ProcessMsg(result)
    }
}
```

---

### Phase 6: 单元测试迁移（1天）

#### 6.1 删除 HandleEvent 测试

```bash
# 删除所有测试 HandleEvent 的测试用例
grep -l "HandleEvent" **/*_test.go | xargs rm
```

#### 6.2 添加 Update(Msg) 测试

**文件**: `components/tree/tree_test.go`

```go
func TestTreeView_Update_MouseClick(t *testing.T) {
    tree := NewTreeView()
    tree.SetNodes([]Node{
        {Text: "Node1"},
        {Text: "Node2"},
        {Text: "Node3"},
    })

    // 模拟点击第 2 行
    mouseMsg := &msg.MouseMsg{
        Action:  msg.MouseActionPress,
        Button:  msg.MouseLeft,
        LocalX:  0,
        LocalY:  1,  // 点击第 2 行
    }

    result := tree.Update(mouseMsg)

    // 验证选中了第 2 个节点
    assert.Equal(t, 1, tree.SelectedIndex)
    assert.Equal(t, "Node2", tree.SelectedNode().Text)
    assert.Equal(t, cmd.None(), result)
}

func TestTreeView_Update_KeyNavigation(t *testing.T) {
    tree := NewTreeView()
    tree.SetNodes([]Node{{Text: "A"}, {Text: "B"}, {Text: "C"}})

    // 按向下键
    keyMsg := &msg.KeyMsg{
        Special: runtimeplatform.KeyArrowDown,
    }

    tree.Update(keyMsg)

    assert.Equal(t, 1, tree.SelectedIndex)
}
```

---

### Phase 7: 集成测试（1天）

#### 7.1 Inspector 集成测试

```bash
cd examples/ui_demos/demo2_runtime_internals/inspector_overlay
go run main.go

# 测试场景
# 1. 按 F12 打开 Inspector ✅
# 2. 点击 TreeView 第 3 行 ✅
#    - 验证控制台输出 "Selected row 2: XXX" ✅
#    - 验证 Inspector 显示正确的属性 ✅
# 3. 按 ↓ 键导航到第 4 行 ✅
# 4. 按 Enter 展开节点 ✅
# 5. 点击 Tab 切换到不同面板 ✅
# 6. 在 Input 中输入文字 ✅
# 7. 按 F12 关闭 Inspector ✅
# 8. 验证主应用正常工作 ✅
```

---

## ✅ 验收标准

### 代码清理
- [ ] **完全删除** `HandleEvent(Event)` 接口定义
- [ ] **完全删除** 所有 `HandleEvent` 方法实现
- [ ] **完全删除** `Component` 接口（老的事件接口）
- [ ] **完全删除** Event → Msg 适配器代码
- [ ] **完全删除** 手动坐标计算代码

### 新架构实现
- [ ] Pump 直接输出 `Msg`（MouseMsg/KeyMsg）
- [ ] Router 只处理 `Msg`
- [ ] **所有**组件实现 `Update(Msg) cmd.Cmd`
- [ ] 所有单元测试使用 `Msg`
- [ ] 所有集成测试通过

### 功能验证
- [ ] Inspector TreeView 点击正确
- [ ] 所有组件交互正常
- [ ] 性能无明显下降

---

## 📊 迁移矩阵

| 组件 | 当前接口 | 目标接口 | 状态 |
|------|---------|---------|------|
| Pump | 输出 Event | 输出 Msg | ⏳ 待修改 |
| Router | Dispatch(Event) | Dispatch(Msg) | ⏳ 待修改 |
| TreeView | HandleEvent | Update(Msg) | ⏳ 待修改 |
| Tabs | HandleEvent | Update(Msg) | ⏳ 待修改 |
| Panel | HandleEvent | 无需实现 | ⏳ 待删除 |
| Input | HandleEvent | Update(Msg) | ⏳ 待修改 |
| Button | HandleEvent | Update(Msg) | ⏳ 待修改 |
| Inspector | HandleEvent | Update(Msg) | ⏳ 待修改 |
| ListView | HandleEvent | Update(Msg) | ⏳ 待修改 |

---

## ⚠️ 破坏性变更

### API 变更
- ❌ `event.Component` 接口删除
- ❌ `HandleEvent(Event)` 方法删除
- ✅ `component.Updater` 接口强制实现
- ✅ `Update(Msg) cmd.Cmd` 方法强制实现

### 依赖变更
- ❌ 删除 `framework/msg/adapter.go`（不再需要）
- ✅ Pump 依赖 `framework/msg`（直接创建 Msg）
- ✅ Router 依赖 `framework/msg`
- ✅ 所有组件依赖 `framework/msg` 和 `framework/cmd`

### 测试变更
- ❌ 所有基于 Event 的测试删除
- ✅ 所有测试改为基于 Msg

---

## 🚀 迁移步骤

1. **备份代码**：创建新分支 `feature/complete-msg-migration`
2. **Phase 1**: 修改 Pump（1天）
3. **Phase 2**: 修改 Router（1天）
4. **Phase 3**: 删除老接口（1天）
5. **Phase 4**: 迁移所有组件（2-3天）
6. **Phase 5**: 更新 App 主循环（1天）
7. **Phase 6**: 迁移测试（1天）
8. **Phase 7**: 集成测试（1天）

**总工期**：8-10 天

---

## 🎯 最终架构

```
┌─────────────────────────────────────────────────────────────┐
│                    单一 Msg/Cmd 架构                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  用户输入                                                     │
│      ↓                                                       │
│  Pump (创建 MouseMsg/KeyMsg)                                 │
│      ↓                                                       │
│  Router.Dispatch(Msg) (根据 TargetID 直接分发)               │
│      ↓                                                       │
│  Component.Update(Msg) → Cmd                                 │
│      ↓                                                       │
│  App.Execute(Cmd) → 可能产生新 Msg                            │
│      ↓                                                       │
│  循环继续                                                     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**关键特点**：
- ✅ 单一消息类型：只有 Msg，没有 Event
- ✅ 单一组件接口：只有 Update(Msg)，没有 HandleEvent
- ✅ 直接分发：根据 TargetID 直接分发，无需层次遍历
- ✅ 纯函数风格：Update 返回 Cmd，副作用由 App 执行

---

**状态**: 待审核
**工期**: 8-10 天
**优先级**: 最高（架构统一，技术债务清零）
