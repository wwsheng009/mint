# Event System (事件系统)

事件系统是 Mint Runtime 的核心子系统，提供完整的事件处理和分发机制。

## 职责

- **事件类型定义** - 定义标准事件类型和接口
- **三阶段传播** - Capture → Target → Bubble 事件传播模型
- **事件分发** - 高效的事件路由和处理
- **命中测试** - 确定事件目标组件
- **事件过滤** - 支持事件拦截和转换

## 核心概念

### 三阶段传播

事件遵循 W3C 风格的三阶段传播模型：

```
          ┌─────────────────────────────────┐
          │          Root (Capture)         │ ← Phase 1: Capture 阶段
          │    ┌───────────────────────┐    │
          │    │   Parent (Capture)    │    │
          │    │ ┌───────────────────┐ │    │
          │    │ │ Target (Target)   │ │    │ ← Phase 2: Target 阶段
          │    │ └───────────────────┘ │    │
          │    │   Parent (Bubble)    │    │ ← Phase 3: Bubble 阶段
          │    └───────────────────────┘    │
          │          Root (Bubble)          │
          └─────────────────────────────────┘
```

**阶段说明**:

| 阶段 | 方向 | 说明 | 用途 |
|------|------|------|------|
| **Capture** | 从根到目标 | 自上而下 | 用于容器拦截事件 |
| **Target** | 在目标节点上 | 仅目标 | 目标节点处理事件 |
| **Bubble** | 从目标到根 | 自下而上 | 用于事件委托 |

### 事件类型

```go
type EventType string

const (
    // 键盘事件
    EventTypeKeyDown  EventType = "keydown"
    EventTypeKeyUp    EventType = "keyup"
    EventTypeKeyPress EventType = "keypress"

    // 鼠标事件
    EventTypeMouseDown   EventType = "mousedown"
    EventTypeMouseUp     EventType = "mouseup"
    EventTypeMouseMove   EventType = "mousemove"
    EventTypeMouseWheel  EventType = "mousewheel"
    EventTypeMouseEnter  EventType = "mouseenter"
    EventTypeMouseLeave  EventType = "mouseleave"
    EventTypeClick       EventType = "click"
    EventTypeDoubleClick EventType = "dblclick"

    // 焦点事件
    EventTypeFocus  EventType = "focus"
    EventTypeBlur   EventType = "blur"
    EventTypeFocusIn EventType = "focusin"
    EventTypeFocusOut EventType = "focusout"

    // 系统事件
    EventTypeResize EventType = "resize"
    EventTypeClose  EventType = "close"
    EventTypeQuit   EventType = "quit"
)
```

### 事件属性

事件提供丰富的属性和方法：

```go
type Event interface {
    // 基本信息
    Type() EventType
    Phase() EventPhase
    Timestamp() time.Time
    Target() Node
    CurrentTarget() Node

    // 控制传播
    PreventDefault()
    IsDefaultPrevented() bool
    StopPropagation()
    IsPropagationStopped() bool
    StopImmediatePropagation()
    IsImmediatePropagationStopped() bool
}
```

**控制方法说明**:

| 方法 | 效果 | 用途 |
|------|------|------|
| `PreventDefault()` | 阻止默认行为 | 阻止按键输入、链接跳转等 |
| `StopPropagation()` | 停止冒泡传播 | 阻止事件向上传播 |
| `StopImmediatePropagation()` | 停止所有传播 | 立即停止所有监听器 |

## 使用示例

### 基本使用

```go
import "github.com/wwsheng009/mint/runtime/event"

// 创建键盘事件
keyEvent := event.NewBaseEvent(event.EventTypeKeyDown)
keyEvent.SetTarget(buttonNode)
keyEvent.SetPhase(event.Target)

// 处理事件
if keyEvent.Type() == event.EventTypeKeyDown {
    // 处理按键
}
```

### 事件监听

```go
type Button struct {
    onClick []func(event.Event)
}

func (b *Button) OnClick(handler func(event.Event)) {
    b.onClick = append(b.onClick, handler)
}

func (b *Button) HandleEvent(ev event.Event) {
    if ev.Type() == event.EventTypeClick {
        for _, h := range b.onClick {
            h(ev)
        }
    }
}
```

### 事件分发

```go
dispatcher := event.NewDispatcher()

// 注册监听器
dispatcher.AddListener(buttonNode, event.EventTypeClick, func(ev event.Event) {
    fmt.Println("Button clicked!")
})

// 分发事件
dispatcher.Dispatch(clickEvent)
```

### 阻止默认行为

```go
func (b *Button) HandleEvent(ev event.Event) {
    if ev.Type() == event.EventTypeKeyDown {
        keyData := ev.Data().(map[string]interface{})
        if keyData["key"] == "Enter" {
            // 阻止按 Enter 的默认行为（比如重复触发）
            ev.PreventDefault()

            // 执行自己的点击逻辑
            b.DoClick()
        }
    }
}
```

### 停止事件传播

```go
func (m *Modal) HandleEvent(ev event.Event) {
    if ev.Type() == event.EventTypeClick {
        // 停止事件传播，阻止父容器接收
        ev.StopPropagation()

        // 模态框内部处理点击
        m.HandleModalClick(ev)
    }
}
```

### 命中测试

```go
hitTester := event.NewHitTester(layoutBoxes)

// 确定事件目标
mouseEvent := ev.(*event.MouseEvent)
target := hitTester.HitTest(mouseEvent.X, mouseEvent.Y)

if target != nil {
    fmt.Printf("命中: %s\n", target.ID())
}
```

## 核心类型

### BaseEvent

事件基类，提供默认实现：

```go
type BaseEvent struct {
    eventType  EventType
    phase      EventPhase
    timestamp  time.Time
    target     Node
    current    Node
    prevented  bool
    stopped    bool
    stoppedImm bool
}

func NewBaseEvent(eventType EventType) *BaseEvent
```

### Dispatcher

事件分发器：

```go
type Dispatcher struct {
    listeners map[Node]map[EventType][]ListenerFunc
}

func NewDispatcher() *Dispatcher
func (d *Dispatcher) AddListener(node Node, eventType EventType, handler ListenerFunc)
func (d *Dispatcher) RemoveListener(node Node, eventType EventType)
func (d *Dispatcher) Dispatch(event Event)
```

### HitTester

命中测试器：

```go
type HitTester struct {
    boxes []LayoutBox
}

func NewHitTester(boxes []LayoutBox) *HitTester
func (h *HitTester) AddBox(box LayoutBox)
func (h *HitTester) HitTest(x, y int) Node
```

### Filter

事件过滤器：

```go
type Filter func(event.Event) bool

func (d *Dispatcher) SetFilter(filter Filter)
```

## 具体事件类型

### KeyboardEvent

键盘事件：

```go
type KeyboardEvent struct {
    *BaseEvent
    Key      string
    Ctrl     bool
    Shift    bool
    Alt      bool
    Meta     bool
    Code     string
}

func NewKeyboardEvent(key string, ctrl, shift, alt, meta bool) *KeyboardEvent
```

### MouseEvent

鼠标事件：

```go
type MouseEvent struct {
    *BaseEvent
    X        int
    Y        int
    Button   MouseButton
    Buttons  int
    Ctrl     bool
    Shift    bool
    Alt      bool
}

type MouseButton int
const (
    MouseLeft MouseButton = 0
    MouseMiddle
    MouseRight
)
```

### FocusEvent

焦点事件：

```go
type FocusEvent struct {
    *BaseEvent
    RelatedTarget Node // 焦点丢失或获得的相关目标
}
```

## 文件结构

```
event/
├── README.md              # 本文档
├── event.go               # 事件接口和 BaseEvent
├── dispatch.go            # 事件分发器
├── filter.go              # 事件过滤器
├── handler.go             # 事件处理器接口
├── hittest.go             # 命中测试
├── phase.go               # 事件阶段常量
└── types.go               # 具体事件类型
```

## 性能优化

### 事件池

对于高频事件（如鼠标移动），使用事件池避免重复分配：

```go
var mouseEventPool = event.NewPool(func() interface{} {
    return &event.MouseEvent{}
})

// 从池中获取
ev := mouseEventPool.Get().(*event.MouseEvent)
// 设置属性...
ev.X = x
ev.Y = y

// 使用后放回池中
mouseEventPool.Put(ev)
```

### 防抖和节流

```go
// 防抖
var debouncer = event.NewDebouncer(300*time.Millisecond)

func handleMouseMove(ev event.Event) {
    debouncer(func() {
        // 实际处理...
    })
}

// 节流 - 每秒最多处理 10 次
var throttler = event.NewThrottler(time.Second/10)

func handleScroll(ev event.Event) {
    throttler(func() {
        // 实际处理...
    })
}
```

## 最佳实践

### 1. 使用事件委托

```go
// 好：在容器上监听
container.OnClick(func(ev event.Event) {
    target := ev.Target()
    if target.Type() == "button" {
        // 处理按钮点击
    }
})

// 避免：为每个按钮单独监听
for _, btn := range buttons {
    btn.OnClick(handleButtonClick)
}
```

### 2. 及时停止 propagation

```go
func (m *Modal) HandleEvent(ev event.Event) {
    // 模态框应该阻止事件传播到底层内容
    ev.StopPropagation()
    // ...
}
```

### 3. 合理使用 preventDefault

```go
func (i *Input) HandleEvent(ev event.Event) {
    if ev.Type() == event.EventTypeKeyDown {
        key := ev.Data()["key"].(string)
        if key == "Tab" {
            // 阻止 Tab 的默认行为（切换焦点）
            ev.PreventDefault()
            // 自定义 Tab 处理
        }
    }
}
```

### 4. 避免在事件处理器中执行耗时操作

```go
// 避免
button.OnClick(func(ev event.Event) {
    result := timeConsumingOperation() // 阻塞
    updateUI(result)
})

// 好：使用 goroutine 或事件系统
button.OnClick(func(ev event.Event) {
    go func() {
        result := timeConsumingOperation()
        // 通过 Action 或回调更新 UI
        callback(result)
    }()
})
```

## 调试

### 事件日志

```go
// 启用事件日志
dispatcher.EnableLogging(func(ev event.Event, phase EventPhase) {
    fmt.Printf("%s %s: %s\n", phase, ev.Type(), ev.Target().ID())
})
```

### 事件追踪

```go
// 追踪特定事件
var trackedEvents []event.Event

filteredDispatcher := event.NewDispatcher()
filteredDispatcher.SetFilter(func(ev event.Event) bool {
    // 记录所有事件
    trackedEvents = append(trackedEvents, ev)
    return true
})
```

## 常见问题

### Q: 如何区分原始目标和当前目标？

A: `Target()` 是事件最初触发的组件，`CurrentTarget()` 是当前正在处理事件的组件。

### Q: Capture 和 Bubble 阶段有什么区别？

A: Capture 从上到下（外向内），Bubble 从下到上（内向外）。Capture 适合父组件拦截，Bubble 适合事件委托。

### Q: 如何实现自定义事件类型？

A: 定义新的 `EventType` 常量，继承 `BaseEvent` 并添加自定义字段。

### Q: 停止 propagation 后，事件还会继续传播吗？

A: `StopPropagation()` 只停止后续的 Bubble 阶段，`StopImmediatePropagation()` 立即停止当前节点的所有后续监听器。

## 相关文档

- [DOM 事件规范](https://dom.spec.whatwg.org/#events)
- [事件委托模式](../../framework/docs/event/delegation.md)
- [DevTools 事件追踪](../../devtools/docs/)

## 测试

```bash
go test ./runtime/event/... -v
```

## 注意事项

此目录必须保持纯 Go 实现，不能依赖：
- ✅ Bubble Tea
- ✅ DSL 解析器
- ✅ 具体组件
- ✅ lipgloss
