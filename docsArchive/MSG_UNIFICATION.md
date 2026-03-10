# Msg 统一设计与 Sandbox 集成

> **版本**: v1.0
> **日期**: 2025-02-10
> **状态**: 设计阶段

---

## 📋 设计目标

### 核心原则（来自 long_term_event_architecture.md）

1. **统一 Msg 层**: 一切输入/异步产出皆 Msg
2. **纯函数 Update**: 状态演进在 `Update(Msg)`，输出 Cmd 副作用
3. **渲染分离**: 渲染只读 Model（VNode 树）
4. **可组合 Cmd**: 异步、定时、IO 均以 Cmd 表达

### 需要处理的事件类型

| 事件类型 | 来源 | 处理方式 |
|---------|------|---------|
| **键盘事件** | Terminal (ANSI) | KeyMsg → Action |
| **鼠标事件** | Terminal (ANSI) | MouseMsg → Action |
| **自定义事件** | 组件内部 | CustomMsg |
| **Sandbox 注入** | 测试/调试 | SandboxMsg |

---

## 🎨 Msg 类型定义

### 核心接口

**文件**: `framework/msg/msg.go`

```go
package msg

// Msg 是所有消息的统一类型
type Msg interface {
    // Type 返回消息类型的唯一标识
    Type() MsgType

    // String 返回调试字符串
    String() string
}

// MsgType 消息类型
type MsgType string

const (
    KeyMsgType    MsgType = "key"
    MouseMsgType  MsgType = "mouse"
    ActionMsgType MsgType = "action"
    ResizeMsgType MsgType = "resize"
    TickMsgType   MsgType = "tick"
    CustomMsgType MsgType = "custom"
    SandboxMsgType MsgType = "sandbox"  // Sandbox 注入
)
```

### 1. KeyMsg（键盘消息）

```go
// KeyMsg 键盘消息
type KeyMsg struct {
    // 按键信息
    Rune     rune   // 字符键 (如 'a', 'A', '1')
    Special  SpecialKey  // 特殊键 (如 Enter, Escape)

    // 修饰键
    Alt   bool
    Ctrl  bool
    Shift bool

    // 重复标志
    Repeat bool
}

func (k *KeyMsg) Type() MsgType { return KeyMsgType }

func (k *KeyMsg) String() string {
    if k.Special != KeyUnknown {
        return fmt.Sprintf("KeyMsg{%s, Alt:%v, Ctrl:%v, Shift:%v}",
            k.Special, k.Alt, k.Ctrl, k.Shift)
    }
    return fmt.Sprintf("KeyMsg{'%c', Alt:%v, Ctrl:%v, Shift:%v}",
        k.Rune, k.Alt, k.Ctrl, k.Shift)
}

// IsPrintable 是否为可打印字符
func (k *KeyMsg) IsPrintable() bool {
    return k.Rune >= 32 && k.Rune <= 126
}

// HasModifier 是否有修饰键
func (k *KeyMsg) HasModifier() bool {
    return k.Alt || k.Ctrl || k.Shift
}
```

### 2. MouseMsg（鼠标消息）

```go
// MouseMsg 鼠标消息
type MouseMsg struct {
    // 屏幕坐标
    X, Y int

    // 命中信息（由 Pump 填充）
    TargetID string     // 目标节点 ID
    LocalX   int        // 相对于目标的坐标
    LocalY   int

    // 鼠标按钮
    Button MouseButton

    // 鼠标动作
    Action MouseAction

    // 滚轮增量
    Delta int  // +1/-1

    // 修饰键
    Alt   bool
    Ctrl  bool
    Shift bool
}

func (m *MouseMsg) Type() MsgType { return MouseMsgType }

func (m *MouseMsg) String() string {
    return fmt.Sprintf("MouseMsg{%s at (%d,%d), Target:%s, Local:(%d,%d)}",
        m.Action, m.X, m.Y, m.TargetID, m.LocalX, m.LocalY)
}

type MouseButton int

const (
    MouseNone MouseButton = iota
    MouseLeft
    MouseMiddle
    MouseRight
)

type MouseAction int

const (
    MousePress MouseAction = iota
    MouseRelease
    MouseMove
    MouseWheel
)

// IsClick 是否为点击
func (m *MouseMsg) IsClick() bool {
    return m.Action == MousePress && m.Button == MouseLeft
}

// IsScroll 是否为滚动
func (m *MouseMsg) IsScroll() bool {
    return m.Action == MouseWheel
}
```

### 3. ActionMsg（语义化消息）

```go
// ActionMsg Action 消息（语义化）
type ActionMsg struct {
    Action *action.Action
}

func (a *ActionMsg) Type() MsgType { return ActionMsgType }

func (a *ActionMsg) String() string {
    return fmt.Sprintf("ActionMsg{%s, Target:%s}",
        a.Action.Type, a.Action.TargetID)
}
```

### 4. CustomMsg（自定义消息）

```go
// CustomMsg 自定义消息
type CustomMsg struct {
    TypeStr string       // 自定义类型标识
    Data    interface{} // 任意数据
}

func (c *CustomMsg) Type() MsgType { return CustomMsgType }

func (c *CustomMsg) String() string {
    return fmt.Sprintf("CustomMsg{%s, Data:%v}", c.TypeStr, c.Data)
}

// NewCustomMsg 创建自定义消息
func NewCustomMsg(typeStr string, data interface{}) *CustomMsg {
    return &CustomMsg{
        TypeStr: typeStr,
        Data:    data,
    }
}
```

### 5. TickMsg（定时器消息）

```go
// TickMsg 定时器消息
type TickMsg struct {
    Time time.Time
    Tag  string  // 用于区分不同定时器
}

func (t *TickMsg) Type() MsgType { return TickMsgType }

func (t *TickMsg) String() string {
    return fmt.Sprintf("TickMsg{Tag:%s, Time:%v}", t.Tag, t.Time)
}
```

### 6. ResizeMsg（窗口大小变化）

```go
// ResizeMsg 窗口大小变化消息
type ResizeMsg struct {
    Width  int
    Height int
}

func (r *ResizeMsg) Type() MsgType { return ResizeMsgType }

func (r *ResizeMsg) String() string {
    return fmt.Sprintf("ResizeMsg{%dx%d}", r.Width, r.Height)
}
```

### 7. SandboxMsg（Sandbox 注入消息）

```go
// SandboxMsg Sandbox 注入消息（用于测试和调试）
type SandboxMsg struct {
    // 注入类型
    InjectType SandboxInjectType

    // 目标节点
    TargetID string

    // 注入数据
    Data interface{}

    // 来源标识
    Source string
}

func (s *SandboxMsg) Type() MsgType { return SandboxMsgType }

func (s *SandboxMsg) String() string {
    return fmt.Sprintf("SandboxMsg{%s, Target:%s, Source:%s}",
        s.InjectType, s.TargetID, s.Source)
}

// SandboxInjectType Sandbox 注入类型
type SandboxInjectType string

const (
    // 键盘注入
    SandboxInjectKeyPress   SandboxInjectType = "inject_key_press"
    SandboxInjectKeyRelease SandboxInjectType = "inject_key_release"
    SandboxInjectKeySequence SandboxInjectType = "inject_key_sequence"  // 序列

    // 鼠标注入
    SandboxInjectMouseClick    SandboxInjectType = "inject_mouse_click"
    SandboxInjectMouseDoubleClick SandboxInjectType = "inject_mouse_double_click"
    SandboxInjectMouseMove    SandboxInjectType = "inject_mouse_move"
    SandboxInjectMouseScroll   SandboxInjectType = "inject_mouse_scroll"
    SandboxInjectMouseHover    SandboxInjectType = "inject_mouse_hover"

    // Action 注入（直接注入语义化 Action）
    SandboxInjectAction        SandboxInjectType = "inject_action"

    // 状态注入（直接修改组件状态）
    SandboxInjectSetState      SandboxInjectType = "inject_set_state"
    SandboxInjectSetProp       SandboxInjectType = "inject_set_prop"

    // 事件序列注入
    SandboxInjectEventSequence SandboxInjectType = "inject_event_sequence"  // 批量事件
    SandboxInjectScript        SandboxInjectType = "inject_script"  // 脚本

    // 时间控制
    SandboxInjectTimeTravel    SandboxInjectType = "inject_time_travel"  // 时间旅行
    SandboxInjectFastForward   SandboxInjectType = "inject_fast_forward"  // 快进
    SandboxInjectPause         SandboxInjectType = "inject_pause"  // 暂停
)

// IsInput 是否为输入注入
func (s *SandboxMsg) IsInput() bool {
    switch s.InjectType {
    case SandboxInjectKeyPress, SandboxInjectKeyRelease,
         SandboxInjectMouseClick, SandboxInjectMouseMove,
         SandboxInjectMouseScroll:
        return true
    }
    return false
}

// IsDirectAction 是否为直接 Action 注入
func (s *SandboxMsg) IsDirectAction() bool {
    return s.InjectType == SandboxInjectAction
}

// IsStateMutation 是否为状态变更注入
func (s *SandboxMsg) IsStateMutation() bool {
    switch s.InjectType {
    case SandboxInjectSetState, SandboxInjectSetProp:
        return true
    }
    return false
}
```

---

## 🔄 Msg 转换层

### Event → Msg 适配器

**文件**: `framework/msg/adapter.go`

```go
package msg

// ToMsg 将 Event 转换为 Msg
func ToMsg(ev event.Event) Msg {
    switch e := ev.(type) {
    case *event.KeyEvent:
        return &KeyMsg{
            Rune:    rune(e.Key),
            Special: e.Special,
            Alt:     e.Modifiers.Has(event.ModAlt),
            Ctrl:    e.Modifiers.Has(event.ModCtrl),
            Shift:   e.Modifiers.Has(event.ModShift),
            Repeat:  e.Repeat,
        }

    case *event.MouseEvent:
        return &MouseMsg{
            X:        e.X,
            Y:        e.Y,
            TargetID: e.TargetID,
            LocalX:   e.LocalX,
            LocalY:   e.LocalY,
            Button:   MouseButton(e.Button),
            Action:   MouseAction(e.Type()),
            Delta:    e.Delta,
            Alt:      e.Modifiers.Has(event.ModAlt),
            Ctrl:     e.Modifiers.Has(event.ModCtrl),
            Shift:    e.Modifiers.Has(event.ModShift),
        }

    case *event.ResizeEvent:
        return &ResizeMsg{
            Width:  e.Width,
            Height: e.Height,
        }

    default:
        return nil
    }
}
```

### Msg → Action 转换

**文件**: `framework/msg/processor.go`

```go
package msg

// MsgProcessor 处理 Msg，输出 Action
type MsgProcessor struct {
    keyMap *action.KeyMap
}

// Process 处理 Msg
func (p *MsgProcessor) Process(msg Msg) []*action.Action {
    switch m := msg.(type) {
    case *KeyMsg:
        return p.processKeyMsg(m)
    case *MouseMsg:
        return p.processMouseMsg(m)
    case *SandboxMsg:
        return p.processSandboxMsg(m)
    default:
        return nil
    }
}

// processKeyMsg 处理键盘消息
func (p *MsgProcessor) processKeyMsg(msg *KeyMsg) []*action.Action {
    // 构建 KeyEvent
    keyEv := &event.KeyEvent{
        Key:      msg.Rune,
        Special:  msg.Special,
        Modifiers: p.buildModifiers(msg),
        Repeat:   msg.Repeat,
    }

    // 使用 KeyMap 转换
    if act := p.keyMap.Lookup(keyEv); act != nil {
        return []*action.Action{act}
    }

    // 默认转换规则
    return []*action.Action{p.defaultKeyConversion(msg)}
}

// processMouseMsg 处理鼠标消息
func (p *MsgProcessor) processMouseMsg(msg *MouseMsg) []*action.Action {
    var act *action.Action

    switch msg.Action {
    case MousePress:
        if msg.Button == MouseLeft {
            act = &action.Action{
                Type:     action.ActionClick,
                TargetID: msg.TargetID,
                Payload:  struct{ X, Y int }{msg.LocalX, msg.LocalY},
            }
        } else if msg.Button == MouseRight {
            act = &action.Action{
                Type:     action.ActionRightClick,
                TargetID: msg.TargetID,
            }
        }
    case MouseWheel:
        act = &action.Action{
            Type:     action.ActionScroll,
            TargetID: msg.TargetID,
            Payload:  msg.Delta,
        }
    case MouseMove:
        act = &action.Action{
            Type:     action.ActionHover,
            TargetID: msg.TargetID,
            Payload:  struct{ X, Y int }{msg.LocalX, msg.LocalY},
        }
    }

    if act != nil {
        return []*action.Action{act}
    }
    return nil
}

// processSandboxMsg 处理 Sandbox 注入消息
func (p *MsgProcessor) processSandboxMsg(msg *SandboxMsg) []*action.Action {
    switch msg.InjectType {
    case SandboxInjectAction:
        // 直接注入 Action
        if act, ok := msg.Data.(*action.Action); ok {
            act.TargetID = msg.TargetID
            return []*action.Action{act}
        }

    case SandboxInjectKeyPress:
        // 注入按键
        if keyData, ok := msg.Data.(KeyData); ok {
            return []*action.Action{
                Type:    action.ActionInputText,
                Payload: string(keyData.Rune),
            }}
        }

    case SandboxInjectMouseClick:
        // 注入点击
        return []*action.Action{
            Type:     action.ActionClick,
            TargetID: msg.TargetID,
        }}

    case SandboxInjectEventSequence:
        // 注入事件序列
        if sequence, ok := msg.Data.([]Msg); ok {
            var actions []*action.Action
            for _, seqMsg := range sequence {
                actions = append(actions, p.Process(seqMsg)...)
            }
            return actions
        }

    default:
        // 其他注入类型...
    }

    return nil
}
```

---

## 🎮 Cmd 副作用系统

### Cmd 定义

**文件**: `framework/cmd/cmd.go`

```go
package cmd

// Cmd 副作用命令
type Cmd interface {
    // Execute 执行命令，返回产生的 Msg
    Execute() Msg
}
```

### 标准 Cmd

```go
// Batch 组合多个 Cmd
type Batch []Cmd

func (b Batch) Execute() Msg {
    // 并发执行，返回第一个非 nil msg
    ch := make(chan Msg, len(b))
    for _, c := range b {
        go func(cmd Cmd) {
            ch <- cmd.Execute()
        }(c)
    }
    return <-ch
}

// After 延迟执行
type After struct {
    Duration time.Duration
    Msg      Msg
}

func (a *After) Execute() Msg {
    time.Sleep(a.Duration)
    return a.Msg
}

// Tick 定时触发
type Tick struct {
    Interval time.Duration
    Tag      string
}

func (t *Tick) Execute() Msg {
    for {
        time.Sleep(t.Interval)
        return &msg.TickMsg{Time: time.Now(), Tag: t.Tag}
    }
}

// IO 执行 IO 操作
type IO struct {
    Fn func() Msg
}

func (i *IO) Execute() Msg {
    return i.Fn()
}
```

---

## 🧪 Sandbox 集成

### Sandbox 注入 API

**文件**: `framework/sandbox/injector.go`

```go
package sandbox

// Injector Sandbox 注入器（用于测试和调试）
type Injector struct {
    app *framework.App
}

// NewInjector 创建注入器
func NewInjector(app *framework.App) *Injector {
    return &Injector{app: app}
}

// InjectKey 注入按键
func (inj *Injector) InjectKey(key rune, special SpecialKey, modifiers Modifier) {
    msg := &msg.SandboxMsg{
        InjectType: msg.SandboxInjectKeyPress,
        Data: KeyData{
            Rune:      key,
            Special:   special,
            Modifiers: modifiers,
        },
        Source: "sandbox",
    }
    inj.app.ProcessMsg(msg)
}

// InjectKeySequence 注入按键序列
func (inj *Injector) InjectKeySequence(keys string) {
    var msgs []msg.Msg
    for _, r := range keys {
        msgs = append(msgs, &msg.SandboxMsg{
            InjectType: msg.SandboxInjectKeyPress,
            Data: KeyData{Rune: r},
            Source: "sandbox",
        })
    }

    // 注入序列
    inj.app.ProcessMsg(&msg.SandboxMsg{
        InjectType: msg.SandboxInjectEventSequence,
        Data: msgs,
        Source: "sandbox",
    })
}

// InjectMouseClick 注入鼠标点击
func (inj *Injector) InjectMouseClick(targetID string) {
    // 从 HitMap 查找节点坐标
    hitMap := inj.app.GetHitMap()
    if hitMap == nil {
        return
    }

    entry := hitMap.FindByID(targetID)
    if entry == nil {
        return
    }

    // 计算中心点
    centerX := entry.Bounds.X + entry.Bounds.Width/2
    centerY := entry.Bounds.Y + entry.Bounds.Height/2

    msg := &msg.SandboxMsg{
        InjectType: msg.SandboxInjectMouseClick,
        TargetID:   targetID,
        Data: MouseData{
            X: centerX,
            Y: centerY,
            LocalX: entry.Bounds.Width/2,
            LocalY: entry.Bounds.Height/2,
        },
        Source: "sandbox",
    }
    inj.app.ProcessMsg(msg)
}

// InjectAction 直接注入 Action
func (inj *Injector) InjectAction(act *action.Action) {
    msg := &msg.SandboxMsg{
        InjectType: msg.SandboxInjectAction,
        Data:       act,
        Source:     "sandbox",
    }
    inj.app.ProcessMsg(msg)
}

// InjectSetState 注入状态变更
func (inj *Injector) InjectSetState(targetID string, state interface{}) {
    msg := &msg.SandboxMsg{
        InjectType: msg.SandboxInjectSetState,
        TargetID:   targetID,
        Data:       state,
        Source:     "sandbox",
    }
    inj.app.ProcessMsg(msg)
}
```

### Sandbox 测试辅助

**文件**: `framework/sandbox/testable_app.go`

```go
package sandbox

// TestableApp 可测试的应用
type TestableApp struct {
    *framework.App
    t *testing.T
    injector *Injector
}

// NewTestableApp 创建可测试应用
func NewTestableApp(t *testing.T) *TestableApp {
    app := framework.NewApp()
    return &TestableApp{
        App:      app,
        t:        t,
        injector: NewInjector(app),
    }
}

// InjectMouseByID 按节点 ID 注入鼠标点击
func (ta *TestableApp) InjectMouseByID(nodeID string) {
    ta.injector.InjectMouseClick(nodeID)
}

// InjectKeys 注入按键序列
func (ta *TestableApp) InjectKeys(keys string) {
    ta.injector.InjectKeySequence(keys)
}

// AssertHovered 断言某个节点被 hover
func (ta *TestableApp) AssertHovered(nodeID string) {
    if ta.hoveredNodeID != nodeID {
        ta.t.Errorf("Expected %s to be hovered, got %s", nodeID, ta.hoveredNodeID)
    }
}

// AssertFocused 断言某个节点获得焦点
func (ta *TestableApp) AssertFocused(nodeID string) {
    if ta.focusedNodeID != nodeID {
        ta.t.Errorf("Expected %s to be focused, got %s", nodeID, ta.focusedNodeID)
    }
}
```

---

## 📐 组件 Update 接口

### 可选的 Update API

**文件**: `framework/component/updater.go`

```go
package component

// Updater 可更新组件（可选实现）
type Updater interface {
    // Update 处理 Msg，返回 (副作用, 是否消费)
    Update(msg.Msg) (cmd.Cmd, bool)
}

// App 支持 Update 循环
func (a *App) processMsg(msg msg.Msg) {
    // 检查根组件是否实现 Update
    if updater, ok := a.root.(Updater); ok {
        cmd, handled := updater.Update(msg)
        if handled {
            a.dirty = true
            if cmd != nil {
                a.StartCmd(cmd)
            }
            return
        }
    }

    // 回退到旧的事件系统
    ev := msgToEvent(msg)
    a.dispatchEvent(ev)
}
```

### 示例：TreeView 实现 Update

```go
// Update 处理消息
func (t *TreeView) Update(msg msg.Msg) (cmd.Cmd, bool) {
    switch m := msg.(type) {
    case *msg.ActionMsg:
        return nil, t.HandleAction(m.Action)

    case *msg.MouseMsg:
        // 转换为 Action
        act := t.mouseToAction(m)
        if act != nil {
            return nil, t.HandleAction(act)
        }

    case *msg.SandboxMsg:
        // Sandbox 注入
        return t.handleSandbox(m)

    default:
        return nil, false
    }

    return nil, false
}
```

---

## ✅ 验收标准

### 功能完整性
- [ ] KeyMsg/MouseMsg/CustomMsg/SandboxMsg 定义完整
- [ ] Event → Msg 转换正确
- [ ] Msg → Action 转换正确
- [ ] Sandbox 注入所有类型可用
- [ ] 组件可选择性实现 Update

### 测试工具
- [ ] TestableApp 支持按 ID 注入
- [ ] Injector 支持 Keys/Mouse/Action 注入
- [ ] 断言助手完整

### 文档
- [ ] Msg API 文档
- [ ] Sandbox 使用指南
- [ ] Update 迁移指南

---

**下一步**: 查看 [任务列表](./TASKS.md) 开始实施
