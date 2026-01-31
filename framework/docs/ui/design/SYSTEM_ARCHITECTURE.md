# Mint UI 系统架构文档

## 📋 文档概述

**Mint UI** 是一套现代的 Go 语言终端用户界面（TUI）框架，采用声明式架构、函数式组件、并发调度等现代前端范式，为终端应用提供完整的运行时平台。

**版本**: v0.1  
**最后更新**: 2026-01-31

---

## 🎯 核心设计理念

### 1. 声明式 UI（Declarative UI）

```go
// 传统命令式 API（避免）
text := NewText("Hello")
text.SetPosition(1, 1)
text.SetColor(FgRed)
container.Add(text)

// 声明式 API（推荐）
func App() VNode {
    return ui.HStack(
        ui.Text("Hello").FgColor(FgRed),
        ui.Text("World").FgColor(FgBlue),
    )
}
```

### 2. 函数组件（Functional Components）

```go
type CounterProps struct {
    Title string
}

func Counter(props CounterProps) VNode {
    count, setCount := useState(0)
    
    return ui.VStack(
        ui.Text(props.Title),
        ui.Button(fmt.Sprintf("Count: %d", count)).OnClick(func() {
            setCount(count + 1)
        }),
    )
}
```

### 3. VNode（虚拟节点）

**核心数据结构**：

```go
type VNode interface {
    // 类型标识
    Type() VNodeType
    
    // 子节点
    Children() []VNode
    
    // Props
    Props() Props
    
    // Key（用于 Diff）
    Key() string
}
```

**VNode 类型**：

| 类型 | 说明 | 示例 |
|------|------|------|
| `VNodeComponent` | 组件节点 | `Counter{...}` |
| `VNodeElement` | 元素节点 | `ui.Text("Hello")` |
| `VNodeText` | 文本节点 | `"Hello World"` |
| `VNodeFragment` | 片段节点 | `ui.Fragment(...)` |

---

## 🏗️ 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                          │
│  (Components, Forms, Styling - High-level APIs)              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      UI Framework                             │
│  (Reconciler, Layout, Render, Event, State, Animation)      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Runtime Core                             │
│  (Paint, Event, Action, Terminal I/O - Pure Go)              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Terminal Driver                           │
│  (ANSI sequences, Input parsing - OS abstraction)            │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 核心子系统

### 1. 声明式 UI 系统

#### 1.1 函数组件

**特性**：
- 无状态函数
- Props 输入
- VNode 输出
- 支持 Hooks

**示例**：

```go
type ButtonProps struct {
    Label string
    OnClick func()
}

func Button(props ButtonProps) VNode {
    return ui.Element("button").
        Prop("label", props.Label).
        OnClick(props.OnClick)
}
```

#### 1.2 Hooks 机制

**内置 Hooks**：

| Hook | 说明 | 示例 |
|------|------|------|
| `useState` | 本地状态 | `count, setCount := useState(0)` |
| `useEffect` | 副作用 | `useEffect(func() {...}, []dependency)` |
| `useContext` | 上下文 | `theme := useContext(ThemeContext)` |
| `useMemo` | 缓存计算 | `value := useMemo(func() {...}, []dependency)` |
| `useCallback` | 缓存函数 | `fn := useCallback(func() {...}, []dependency)` |
| `useRef` | 可变引用 | `ref := useRef(initial)` |

**Hooks 实现原理**：

```go
type Hook struct {
    Type     HookType
    State    interface{}
    Cleanup  func()
    Deps     []interface{}
}

type ComponentContext struct {
    Hooks    []Hook
    HookIndex int
    // ...
}
```

---

### 2. Reconciler 系统

#### 2.1 Diff 算法

**Diff 策略**：

```go
func diff(old, new VNode) Patch {
    switch {
    case old == nil && new != nil:
        return Create(new)
    case old != nil && new == nil:
        return Delete(old)
    case old.Type() != new.Type():
        return Replace(old, new)
    case old.Key() != new.Key():
        return Replace(old, new)
    default:
        return diffChildren(old, new)
    }
}
```

**Diff 规则**：

1. **同层比较**：只在同层级比较
2. **类型判断**：类型不同直接替换
3. **Key 识别**：Key 不同直接替换
4. **子节点 Diff**：使用双指针算法优化

#### 2.2 Fiber 架构

**Fiber 节点**：

```go
type Fiber struct {
    // VNode 关联
    VNode VNode
    
    // 树结构
    Return   *Fiber  // 父节点
    Child    *Fiber  // 第一个子节点
    Sibling  *Fiber  // 下一个兄弟节点
    
    // 工作单元状态
    PendingProps Props
    MemoizedProps Props
    UpdateQueue  *UpdateQueue
    
    // Effect
    EffectTag EffectTag
    NextEffect *Fiber
    
    // 优先级
    Lanes Lanes
}
```

**工作循环**：

```
Work Loop → PerformUnitOfWork → BeginWork → Children
                 ↓
             CompleteWork
                 ↓
             CommitWork
```

#### 2.3 Scheduler（调度器）

**优先级定义**：

```go
type Lane uint64

const (
    SyncLane Lane = 0b00000001  // 同步（输入事件）
    InputLane Lane = 0b00000010  // 输入（按键）
    AnimationLane Lane = 0b00000100  // 动画
    TransitionLane Lane = 0b00001000  // 过渡
    IdleLane Lane = 0b10000000  // 空闲
)
```

**时间切片**：

```go
func workLoop(deadline time.Time) {
    for {
        if time.Now().After(deadline) {
            break  // 时间片用完
        }
        
        performUnitOfWork()
    }
    
    // 请求下一帧
    requestAnimationFrame(workLoop)
}
```

---

### 3. Layout 系统

#### 3.1 约束驱动布局

**核心概念**：

```
Parent Constraint:
  - MinWidth: 0
  - MaxWidth: 80
  - MinHeight: 0
  - MaxHeight: 24

↓

Child Layout:
  - Chooses: Width: 40, Height: 10

↓

Parent Position:
  - Child positioned at: x=0, y=0
```

**布局接口**：

```go
type Measurable interface {
    Measure(constraint Constraint) Size
}

type Positionable interface {
    Position(x, y int)
}
```

#### 3.2 Flexbox 布局

**支持的方向**：

```go
// 水平排列
ui.HStack(
    ui.Text("A"),
    ui.Text("B"),
    ui.Text("C"),
)

// 垂直排列
ui.VStack(
    ui.Text("A"),
    ui.Text("B"),
    ui.Text("C"),
)
```

**Flex 参数**：

```go
ui.HStack(
    ui.Text("Left").Flex(1),      // 比例 1
    ui.Text("Middle").Flex(2),    // 比例 2
    ui.Text("Right").Flex(1),    // 比例 1
)
```

**对齐方式**：

```go
ui.HStack(
    ui.Align(Start),      // 起始对齐
    ui.Align(Center),     // 居中对齐
    ui.Align(End),        // 结束对齐
)
```

#### 3.3 Grid 布局 (新增)

**二维网格布局**：

```go
ui.Grid(
    []ui.Dimension{ui.Fixed(10), ui.Fixed(10), ui.Flex(1)},
    []ui.Dimension{ui.Fixed(5), ui.Flex(1)},
    ui.UICell(0, 0, CpuPanel()),
    ui.UICell(0, 1, MemPanel()),
    ui.UICellSpan(1, 0, 1, 2, LogsPanel()), // 跨列
)
```

**Grid 特性**：

- 固定尺寸：`ui.Fixed(n)`
- 弹性尺寸：`ui.Flex(n)` - 按比例分配
- 自动尺寸：`ui.Auto()` - 由内容决定
- 跨行跨列：`ui.UICellSpan(row, col, rowSpan, colSpan, child)`
- 间距控制：`RowGap`, `ColGap`

详见：[GRID_LAYOUT_DESIGN.md](GRID_LAYOUT_DESIGN.md)

#### 3.4 Absolute 定位 (新增)

**脱离布局流的绝对定位**：

```go
ui.Box().Child(
    ui.Stack(
        ui.Text("Background"),
        ui.Absolute(
            ui.Text("Overlay"),
        ).Top(0).Right(0),
    ),
)
```

**Absolute 特性**：

- 位置控制：`Top()`, `Bottom()`, `Left()`, `Right()`
- 百分比定位：`TopPercent(n)`, `LeftPercent(n)`
- 尺寸控制：`Width()`, `Height()`
- Z-Index：控制层级
- 锚点：`Anchor(Center)`, `Anchor(TopRight)` 等

详见：[ABSOLUTE_POSITIONING_DESIGN.md](ABSOLUTE_POSITIONING_DESIGN.md)

#### 3.5 虚拟化渲染

**Viewport 机制**：

```go
type VirtualList struct {
    Items      []Item
    ItemHeight int  // 可变高度支持
    Visible    []Item
}

func (vl *VirtualList) Measure(constraint Constraint) Size {
    viewportHeight := constraint.MaxHeight
    
    // 计算可见范围
    start := vl.scrollTop / vl.itemHeight
    end := start + viewportHeight/vl.itemHeight + 2
    
    vl.visible = vl.items[start:end]
    
    return constraint.Constrain(Size{
        Width:  constraint.MaxWidth,
        Height: len(vl.visible) * vl.itemHeight,
    })
}
```

---

### 4. 渲染管线

#### 4.1 渲染管线流程

```
VNode → Fiber → Layout → Paint → Rasterize → Buffer → Terminal
```

**详细流程**：

```
1. Render Phase（可中断）
   ├─ BeginWork: 创建/更新 Fiber
   ├─ CompleteWork: 标记 Effect
   └─ 处理所有节点

2. Commit Phase（不可中断）
   ├─ Before Mutation: 执行 getSnapshotBeforeUpdate
   ├─ Mutation: DOM 操作（终端 Buffer）
   └─ Layout: 执行 useEffect

3. Paint Phase
   ├─ Generate DrawCmd
   ├─ Apply Styles
   └─ Rasterize to Cells
```

#### 4.2 DrawCmd（绘制命令）

**DrawCmd 类型**：

```go
type DrawCmd interface {
    Type() DrawCmdType
}

type DrawText struct {
    X, Y    int
    Text    string
    Style   Style
}

type DrawRect struct {
    X, Y, W, H int
    Style      Style
}

type DrawClip struct {
    X, Y, W, H int
}

type DrawTransform struct {
    OffsetX, OffsetY int
}
```

#### 4.3 光栅化（Rasterization）

**Cell 数据结构**：

```go
type Cell struct {
    Char  rune    // 字符
    Fg    Color   // 前景色
    Bg    Color   // 背景色
    Style Style   // 样式（粗体、下划线等）
}

type Buffer struct {
    Width  int
    Height int
    Cells  [][]Cell
}
```

**光栅化过程**：

```go
func rasterize(cmds []DrawCmd, width, height int) *Buffer {
    buffer := NewBuffer(width, height)
    
    for _, cmd := range cmds {
        switch c := cmd.(type) {
        case *DrawText:
            buffer.drawText(c.X, c.Y, c.Text, c.Style)
        case *DrawRect:
            buffer.drawRect(c.X, c.Y, c.W, c.H, c.Style)
        // ...
        }
    }
    
    return buffer
}
```

#### 4.4 Buffer Diff（差分更新）

**Diff 算法**：

```go
func diffBuffer(old, new *Buffer) []CellChange {
    changes := []CellChange{}
    
    for y := 0; y < old.Height; y++ {
        for x := 0; x < old.Width; x++ {
            if old.Cells[y][x] != new.Cells[y][x] {
                changes = append(changes, CellChange{
                    X, Y: x, y,
                    Cell: new.Cells[y][x],
                })
            }
        }
    }
    
    return changes
}
```

**优化策略**：

1. **增量更新**：只更新变化的 Cell
2. **批量渲染**：合并相邻的相同 Style Cell
3. **ANSI 优化**：减少 Style 切换

---

### 5. 状态系统

#### 5.1 状态层次

```
Local State（组件本地）
    ↓
Derived State（派生状态）
    ↓
Global State（全局状态）
```

**Local State**：

```go
func Counter() VNode {
    count, setCount := useState(0)  // 本地状态
    doubled, _ := useMemo(func() int {
        return count * 2  // 派生状态
    }, []interface{}{count})
    
    return ui.Text(fmt.Sprintf("Count: %d, Doubled: %d", count, doubled))
}
```

**Global State**：

```go
// 创建全局 Store
store := createStore(CounterState{Count: 0})

// 在组件中使用
func Counter() VNode {
    state := useSelector(store, func(s CounterState) int {
        return s.Count
    })
    dispatch := useDispatch(store)
    
    return ui.Button(fmt.Sprintf("Count: %d", state)).OnClick(func() {
        dispatch(IncrementAction{})
    })
}
```

#### 5.2 状态一致性

**保证机制**：

1. **单一数据源**：每个状态只有一个真值
2. **不可变更新**：状态更新返回新状态
3. **批量更新**：同一事件循环内的多个 State 更新合并
4. **同步更新**：State 更新同步触发 Re-render

**批量更新示例**：

```go
// 多次 useState 调用
setCount(1)  // 不立即渲染
setCount(2)  // 不立即渲染
setCount(3)  // 不立即渲染

// 事件循环结束时，一次性渲染，count = 3
```

---

### 6. 事件系统

#### 6.1 事件流

```
Terminal → Event Queue → Event Dispatcher → Component
                ↓
            State Update
                ↓
              Re-render
```

**事件类型**：

```go
type Event interface {
    Type() EventType
}

type KeyEvent struct {
    Key   rune
    Mod   KeyMod
}

type MouseEvent struct {
    X, Y  int
    Button MouseButton
}

type ResizeEvent struct {
    Width, Height int
}
```

#### 6.2 事件冒泡与捕获

```
┌─────────────────────────────────────┐
│           Parent                    │ ← Capture
│  ┌───────────────────────────────┐  │
│  │         Child                 │  │ ← Capture
│  │  ┌─────────────────────────┐  │  │
│  │  │   Target                │  │  │ ← Target
│  │  │                         │  │  │
│  │  └─────────────────────────┘  │  │ ← Bubble
│  └───────────────────────────────┘  │ ← Bubble
└─────────────────────────────────────┘ ← Bubble
```

**事件处理**：

```go
ui.Button("Click Me").
    OnClick(func(e Event) {
        fmt.Println("Bubble: Button")
        e.StopPropagation()  // 阻止冒泡
    }).
    OnClickCapture(func(e Event) {
        fmt.Println("Capture: Button")
    })
```

---

### 7. 样式系统

#### 7.1 Design Token

**Token 定义**：

```go
type DesignToken struct {
    Colors  ColorPalette
    Spacing SpacingScale
    Fonts   FontScale
    Effects Effects
}

type ColorPalette struct {
    Primary  Color
    Secondary Color
    Success  Color
    Warning  Color
    Error    Color
    Neutral  NeutralColors
}
```

**使用 Token**：

```go
ui.Text("Primary").Color(tokens.Colors.Primary)
```

#### 7.2 主题系统

**主题切换**：

```go
// 主题定义
var LightTheme = Theme{
    Name: "light",
    Palette: ColorPalette{
        Primary: Color{R: 66, G: 133, B: 244},
        Background: Color{R: 255, G: 255, B: 255},
        Text: Color{R: 0, G: 0, B: 0},
    },
}

var DarkTheme = Theme{
    Name: "dark",
    Palette: ColorPalette{
        Primary: Color{R: 100, G: 149, B: 237},
        Background: Color{R: 30, G: 30, B: 30},
        Text: Color{R: 255, G: 255, B: 255},
    },
}

// 切换主题
func App() VNode {
    theme, setTheme := useContext(ThemeContext)
    
    return ui.Button("Toggle Theme").OnClick(func() {
        if theme.Name == "light" {
            setTheme(DarkTheme)
        } else {
            setTheme(LightTheme)
        }
    })
}
```

#### 7.3 Style Diff 优化

**问题**：频繁切换 ANSI 样式代码会影响性能

**优化**：

```go
// 错误写法（频繁切换）
for _, cell := range cells {
    fmt.Print(ansi.FgColor(cell.Fg))
    fmt.Print(ansi.BgColor(cell.Bg))
    fmt.Print(cell.Char)
}

// 优化写成（合并切换）
currentStyle := Style{}
for _, cell := range cells {
    if cell.Style != currentStyle {
        fmt.Print(ansi.ApplyStyle(cell.Style))
        currentStyle = cell.Style
    }
    fmt.Print(cell.Char)
}
```

---

#### 7.4 Style Diff 优化 (新增)

**问题**: 终端渲染中，样式切换比字符输出更昂贵

**解决方案**: 详见 [STYLE_DIFF_DESIGN.md](STYLE_DIFF_DESIGN.md)

**核心机制**:

```go
// 终端状态追踪
type TerminalState struct {
    FgColor   *color.Color
    BgColor   *color.Color
    Bold      bool
    Italic    bool
    Underline bool
}

// 只输出变化的样式
func DiffStyles(old, new Style, state *TerminalState) []string
```

**性能提升**: 输出量减少 99% (60KB → 1KB)

---

### 8. 动画系统

#### 8.1 动画时间轴

```go
type Timeline struct {
    Duration time.Duration
    Easing   EasingFunction
    Keyframes []Keyframe
}

type Keyframe struct {
    Time    float64  // 0.0 - 1.0
    Props   Props
}

// 创建动画
timeline := NewTimeline(500 * time.Millisecond, EaseInOutQuad).
    AddKeyframe(0.0, Props{"opacity": 0}).
    AddKeyframe(1.0, Props{"opacity": 1})
```

#### 8.2 动画 API

```go
func AnimatedButton() VNode {
    opacity, _ := useAnimation(AnimationConfig{
        Duration: 300 * time.Millisecond,
        From:    0.0,
        To:      1.0,
    })
    
    return ui.Text("Hello").Style(Style{Opacity: opacity})
}
```

#### 8.3 动画自愈

**问题**：动画中断后状态不一致

**解决方案**：

```go
type AnimationState struct {
    Running bool
    Cancelled bool
    FinalValue float64
}

func (a *AnimationState) Cancel() {
    a.Cancelled = true
    a.Running = false
    // 应用最终值
    a.Apply(a.FinalValue)
}
```

---

### 9. 远程渲染系统

#### 9.1 架构设计

```
┌──────────────────┐     DrawCmd Stream    ┌──────────────────┐
│   Server (Go)    │ ────────────────────→  │   Client (JS)    │
│                  │                        │                  │
│  - Reconciler    │                        │  - Terminal Emu  │
│  - Layout        │                        │  - WebSocket     │
│  - Render        │                        │  - Canvas/Divs   │
└──────────────────┘                        └──────────────────┘
```

#### 9.2 DrawCmd Streaming

**协议格式**：

```json
{
  "type": "draw_cmd",
  "data": {
    "commands": [
      {"type": "text", "x": 0, "y": 0, "text": "Hello", "style": {...}},
      {"type": "rect", "x": 0, "y": 0, "w": 80, "h": 24, "style": {...}}
    ]
  }
}
```

**服务端**：

```go
func streamDrawCmds(ws *websocket.Conn, fiberRoot *Fiber) {
    for {
        select {
        case cmds := <-drawCmdChannel:
            ws.WriteJSON(map[string]interface{}{
                "type": "draw_cmd",
                "data": cmds,
            })
        }
    }
}
```

**客户端**：

```javascript
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'draw_cmd') {
    renderDrawCmds(msg.data.commands);
  }
};
```

#### 9.3 职责分离

| 组件 | 职责 | 示例 |
|------|------|------|
| Server | 计算、布局、渲染 | Reconciler、Layout、DrawCmd 生成 |
| Client | 显示、交互 | 终端模拟、事件转发 |

---

### 10. DevTools 系统

#### 10.1 组件树可视化

```go
type ComponentNode struct {
    Name     string
    Type     VNodeType
    Props    Props
    State    interface{}
    Children []ComponentNode
}

func buildComponentTree(fiber *Fiber) ComponentNode {
    node := ComponentNode{
        Name:  fiber.VNode.Name(),
        Type:  fiber.VNode.Type(),
        Props: fiber.MemoizedProps,
        State: fiber.State,
    }
    
    for child := fiber.Child; child != nil; child = child.Sibling {
        node.Children = append(node.Children, buildComponentTree(child))
    }
    
    return node
}
```

#### 10.2 布局调试

**可视化布局边界**：

```go
func (l *LayoutDebugger) Paint(b *Buffer) {
    if !l.Enabled {
        return
    }
    
    // 绘制红色边框
    for _, rect := range l.LayoutRects {
        b.drawRect(rect.X, rect.Y, rect.W, rect.H, Style{
            Fg: Color{R: 255, G: 0, B: 0},
            Border: BorderTypeSingle,
        })
    }
}
```

#### 10.3 性能火焰图

```go
type Profiler struct {
    Samples []ProfileSample
}

type ProfileSample struct {
    Name      string
    Duration  time.Duration
    StartTime time.Time
}

func (p *Profiler) Record(name string, fn func()) {
    start := time.Now()
    fn()
    duration := time.Since(start)
    
    p.Samples = append(p.Samples, ProfileSample{
        Name:     name,
        Duration: duration,
    })
}
```

---

### 11. 并发与调度

#### 11.1 优先级队列

```go
type Scheduler struct {
    TaskQueue []*Task  // 按优先级排序
    Lanes     Lanes   // 当前可处理的 Lanes
}

type Task struct {
    Callback func()
    Lanes    Lanes
}
```

#### 11.2 可中断渲染

```go
func renderRoot(fiberRoot *Fiber, deadline time.Time) {
    workLoop(deadline)
    
    if !isWorkComplete() {
        // 请求下一帧继续
        requestAnimationFrame(func() {
            renderRoot(fiberRoot, time.Now().Add(5*time.Millisecond))
        })
    }
}
```

#### 11.3 时间切片

```go
const TimeSlice = 5 * time.Millisecond

func workLoop(deadline time.Time) {
    for {
        if time.Now().After(deadline) {
            break  // 时间片用完
        }
        
        performUnitOfWork()
    }
}

// 主循环
func mainLoop() {
    for {
        start := time.Now()
        deadline := start.Add(TimeSlice)
        
        workLoop(deadline)
        
        // 下一帧
        requestAnimationFrame(mainLoop)
    }
}
```

---

### 10. Layer 层级系统 (新增)

**目标**: 实现视觉层级管理，支持 Modal、Tooltip 等脱离正常布局流的组件

**详见**: [LAYER_SYSTEM_DESIGN.md](LAYER_SYSTEM_DESIGN.md)

**层级定义**:

```go
type Layer int

const (
    LayerBase Layer = iota      // 基础层（默认内容）
    LayerOverlay                 // 覆盖层（下拉菜单）
    LayerModal                   // 模态框层
    LayerTooltip                 // 提示框层
    LayerNotification            // 通知层
)
```

**使用示例**:

```go
// 显示模态框
ui.Modal("my-modal", ModalContent())

// 显示提示
ui.Tooltip("tip", ui.Text("Help text"))

// 显示通知
ui.Toast("toast", ui.Text("Message"))
```

**特性**:
- Focus Trap（焦点陷阱）
- ESC 自动关闭
- 背景冻结
- 事件阻止

---

### 11. 输入优先级调度 (新增)

**目标**: 输入永远优先于渲染，确保即时响应

**详见**: [INPUT_SCHEDULING.md](INPUT_SCHEDULING.md)

**优先级定义**:

```go
type Priority int

const (
    PriorityImmediate Priority = 3  // 输入事件
    PriorityUserBlock Priority = 2  // 交互事件
    PriorityNormal    Priority = 1  // UI 更新
    PriorityLow       Priority = 0  // 后台任务
)
```

**核心机制**:

```go
// 可中断任务
func (t *InterruptibleTask) Execute(inputQueue *InputQueue) error {
    for {
        if inputQueue.HasPending() {
            return ErrInterrupted  // 立即中断
        }
        performUnitOfWork()
    }
}
```

---

### 12. TextBuffer 文本缓冲 (新增)

**目标**: 文本编辑器级的输入处理

**详见**: [TEXT_BUFFER_DESIGN.md](TEXT_BUFFER_DESIGN.md)

**核心特性**:
- UTF-32 rune 存储（避免中文乱码）
- 光标移动（字符、单词、行）
- 选择区（复制/粘贴）
- 撤销/重做

**API 示例**:

```go
// Input 组件内部使用
buffer := input.NewTextBuffer()
buffer.Insert("你好")  // 正确处理中文
buffer.MoveWordForward()
buffer.DeleteWord()
```

---

### 13. 语法高亮 (新增)

**目标**: 代码编辑器级的语法着色

**详见**: [SYNTAX_HIGHLIGHT_DESIGN.md](SYNTAX_HIGHLIGHT_DESIGN.md)

**核心特性**:
- 增量词法分析（只解析修改行）
- 跨行状态传播（多行注释、字符串）
- Token 缓存（避免重复解析）
- 多语言支持（Go, JavaScript, Python, etc.）

**API 示例**:

```go
// 创建增量词法分析器
lexer := editor.NewIncrementalLexer(buffer)

// 文本变化时标记脏行
buffer.OnChange(func(line int) {
    lexer.MarkDirty(line)
})

// 渲染时获取 Token
tokens := lexer.GetTokens(lineNum)
for _, token := range tokens {
    style := editor.GetStyle(token.Type)
    buffer.Draw(token.Text, style)
}
```

---

### 14. 容错与自愈机制

#### 14.1 渲染级容错

```go
func safeRender(fiber *Fiber) VNode {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Render panic recovered: %v", r)
            // 返回错误边界组件
            return ErrorBoundaryVNode(r)
        }
    }()
    
    return fiber.VNode.Render()
}
```

#### 12.2 组件级容错

```go
func ErrorBoundary(props ErrorBoundaryProps) VNode {
    error, _ := useErrorState()
    
    if error != nil {
        return ui.VStack(
            ui.Text("Error occurred").Color(FgRed),
            ui.Text(error.Error()),
            ui.Button("Retry").OnClick(func() {
                props.Retry()
            }),
        )
    }
    
    return props.Children
}
```

#### 12.3 异常保护

**保护机制**：

| 层级 | 保护机制 | 示例 |
|------|---------|------|
| Render | Recover Panic | `defer recover()` |
| Event | Error Handler | `OnError(func(err error))` |
| Animation | Auto Cancel | 中断时应用最终值 |
| Layout | Constraint | 强制边界约束 |

---

## 🌐 平台化设计

### 13.1 三层架构

```
┌─────────────────────────────────────────────────────────┐
│                    Ecosystem                            │
│  (Templates, Components, Plugins, Examples, Docs)        │
└─────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────┐
│                       SDK                                │
│  (ui.Run, ui.View, ui.HStack, ui.useState, etc.)        │
└─────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────┐
│                  Engine Core                             │
│  (Reconciler, Layout, Render, Terminal I/O)              │
└─────────────────────────────────────────────────────────┘
```

### 13.2 SDK 核心 API

```go
package ui

// 运行应用
func Run(app func() VNode, opts ...Option) error

// 创建视图
func View(name string, render func() VNode) Component

// 布局组件
func HStack(children ...VNode) VNode
func VStack(children ...VNode) VNode

// 基础组件
func Text(text string) VNode
func Button(label string) VNode
func Input(placeholder string) VNode

// Hooks
func useState(initial interface{}) (interface{}, func(interface{}))
func useEffect(effect func(), deps []interface{})
func useContext(ctx Context) interface{}
```

### 13.3 生态系统

**Components（组件库）**：
- Form 表单
- Table 表格
- Menu 菜单
- Modal 弹窗
- Toast 提示

**Plugins（插件）**：
- Logger 日志
- Metrics 指标
- Tracing 追踪
- Profiling 性能分析

**Templates（模板）**：
- Dashboard 仪表板
- Admin 管理后台
- Chat 聊天
- Terminal 终端

---

## 📂 目录结构

### v0.1 目录规划

```
mint/
├── cmd/
│   └── main.go              # 应用入口
│
├── internal/
│   ├── reconciler/         # Reconciler 系统
│   │   ├── diff.go
│   │   ├── fiber.go
│   │   └── scheduler.go
│   │
│   ├── layout/             # Layout 系统
│   │   ├── constraint.go
│   │   ├── flexbox.go
│   │   └── virtual.go
│   │
│   ├── render/             # 渲染管线
│   │   ├── drawcmd.go
│   │   ├── rasterize.go
│   │   ├── buffer.go
│   │   ├── style_diff.go      # Style Diff 优化 (新增)
│   │   └── optimizer.go       # 输出优化器 (新增)
│   │
│   ├── terminal/           # 终端驱动
│   │   ├── ansi.go
│   │   ├── input.go
│   │   └── output.go
│   │
│   ├── state/              # 状态系统
│   │   ├── local.go
│   │   ├── derived.go
│   │   └── global.go
│   │
│   ├── event/              # 事件系统
│   │   ├── queue.go
│   │   ├── dispatcher.go
│   │   └── handler.go
│   │
│   ├── style/              # 样式系统
│   │   ├── token.go
│   │   ├── theme.go
│   │   └── diff.go
│   │
│   ├── animation/          # 动画系统
│   │   ├── timeline.go
│   │   ├── easing.go
│   │   └── hooks.go
│   │
│   ├── layer/              # 层级系统 (新增)
│   │   ├── layer.go
│   │   ├── tree.go
│   │   └── manager.go
│   │
│   ├── input/              # 输入处理 (新增)
│   │   ├── buffer.go
│   │   ├── cursor.go
│   │   └── selection.go
│   │
│   ├── scheduler/          # 调度器 (增强)
│   │   ├── scheduler.go
│   │   ├── priority.go
│   │   └── interruptible.go
│   │
│   ├── remote/             # 远程渲染
│   │   ├── protocol.go
│   │   ├── server.go
│   │   └── client.go
│   │
│   └── devtools/           # DevTools
│       ├── tree.go
│       ├── layout.go
│       └── profiler.go
│
├── ui/                     # 对外 SDK
│   ├── app.go
│   ├── component.go
│   ├── hooks.go
│   └── vnode.go
│
├── examples/               # 示例应用
│   ├── hello/
│   ├── demo/
│   └── theme/
│
└── framework/
    └── docs/               # 文档
        ├── ui/
        │   ├── design/
        │   │   └── SYSTEM_ARCHITECTURE.md  # 本文档
        │   └── idea/
        └── ...
```

---

## 🗺️ 开发路线图

### Phase 1: MVP（最小可行产品）

**目标**：验证核心架构可行性

- [ ] 基础 VNode 系统
- [ ] 简单的 Diff 算法
- [ ] 基础 Layout（HStack、VStack）
- [ ] 终端驱动（ANSI 输出）
- [ ] 基础组件（Text、Button、Input）

### Phase 2: DX（开发者体验）

**目标**：提升开发体验

- [ ] 完整的 Hooks 系统
- [ ] 事件系统
- [ ] 状态系统
- [ ] 样式系统
- [ ] DevTools 基础版

### Phase 3: 平台化

**目标**：构建完整平台

- [ ] SDK 完善
- [ ] 组件库
- [ ] 远程渲染
- [ ] 动画系统
- [ ] 虚拟化渲染

### Phase 4: 商业化

**目标**：生态建设

- [ ] 插件系统
- [ ] 模板市场
- [ ] 云端 DevTools
- [ ] 企业级支持

---

## 📊 性能指标

### 目标性能

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 渲染帧率 | 60 FPS | 流畅动画 |
| 布局计算 | < 1 ms | 快速响应 |
| 组件数量 | 10,000+ | 大型应用 |
| 虚拟滚动 | 100,000+ | 大数据量 |
| 内存占用 | < 50 MB | 轻量级 |

### 优化策略

1. **增量渲染**：只渲染变化的组件
2. **虚拟化**：只渲染可视区域
3. **批量更新**：合并多个状态更新
4. **缓存优化**：Memo、useMemo、useCallback
5. **并发调度**：时间切片、优先级队列

---

## 🔒 稳定性检查清单

### 渲染
- [ ] Panic Recover 机制
- [ ] 错误边界组件
- [ ] 无限循环检测

### 布局
- [ ] 约束边界保护
- [ ] 溢出处理
- [ ] 最小尺寸限制

### 状态
- [ ] 状态一致性保证
- [ ] 批量更新机制
- [ ] 内存泄漏防护

### 调度
- [ ] 优先级队列
- [ ] 时间切片
- [ ] 任务取消

### 输入事件
- [ ] 事件队列
- [ ] 防抖节流
- [ ] 事件冒泡控制

### 资源
- [ ] 内存管理
- [ ] Goroutine 泄漏防护
- [ ] 连接池管理

---

## 📚 相关文档

### 核心设计文档

- **`SYSTEM_ARCHITECTURE.md`** - 本文档，系统架构总览
- **`DIRECTORY_STRUCTURE.md`** - 目录结构设计
- **`COMPONENT_CLASSIFICATION.md`** - 组件分类方案
- **`API_DESIGN.md`** - API 设计文档
- **`MIGRATION_GUIDE.md`** - 迁移指南
- **`BENCHMARK.md`** - 性能基准

### 新增设计文档

#### 渲染优化
- **`STYLE_DIFF_DESIGN.md`** - 终端样式优化设计

#### 布局系统
- **`LAYER_SYSTEM_DESIGN.md`** - 视觉层级系统设计 (Modal/Tooltip/Toast)
- **`GRID_LAYOUT_DESIGN.md`** - 二维网格布局设计
- **`ABSOLUTE_POSITIONING_DESIGN.md`** - 绝对定位设计

#### 输入与编辑
- **`TEXT_BUFFER_DESIGN.md`** - 文本缓冲区设计 (UTF-32)
- **`INPUT_SCHEDULING.md`** - 输入优先级调度设计
- **`SYNTAX_HIGHLIGHT_DESIGN.md`** - 语法高亮设计 (增量 Lexer)

#### 分析文档
- **`IDEA_COVERAGE_ANALYSIS.md`** - Idea 文档覆盖分析
- **`DEMO_COVERAGE_ANALYSIS.md`** - Demo 功能覆盖分析

### Idea 构思文档

- **`idea/idea1.md`** - 声明式架构设计理念
- **`idea/idea2_layout.md`** - 布局引擎设计
- **`idea/idea3_vnode.md`** - VNode 与渲染管线
- **`idea/idea4_comp.md`** - 组件系统规范
- **`idea/idea5_style.md`** - 样式系统设计
- **`idea/idea5.1_style_diff.md`** - 样式 Diff 优化
- **`idea/idea4.3_modal.md`** - Modal 组件与 Layer 系统
- **`idea/idea4.4_input.md`** - Input 组件与 TextBuffer
- **`idea/idea6_remote.md`** - 远程渲染协议
- **`idea/idea8_Concurrent.md`** - 并发调度设计
- **`idea/idea9_dev_tools.md`** - DevTools 设计
- **`idea/idea10_checklist.md`** - 稳定性检查清单
- **`idea/idea11_safe.md`** - 容错与自愈机制
- **`idea/idea12_platform.md`** - 平台化落地设计
- **`idea/idea13_roadmap.md`** - 开发路线图
- **`idea/idea14_sdk.md`** - SDK API 设计
- **`idea/idea15_performance.md`** - 性能优化策略
- **`idea/idea16_product.md`** - 产品化思考
- **`idea/idea17_project.md`** - v0.1 功能清单
- **`idea/idea18_0.1.md`** - v0.1 目录结构
- **`idea/idea19_.md`** - v0.1 开发执行表

---

## 🤝 贡献指南

### 代码规范

- **文件命名**：使用 `snake_case`
- **接口设计**：小而专一的能力接口
- **错误处理**：使用 `fmt.Errorf` 包装错误
- **测试覆盖**：目标 > 80%

### 提交规范

```
feat: 添加新的功能
fix: 修复 bug
docs: 更新文档
style: 代码格式调整
refactor: 重构代码
test: 添加测试
chore: 构建/工具链
```

---

## 📄 许可证

MIT License

---

## 📞 联系方式

- **Issues**: https://github.com/yao/wwsheng009/mint/issues
- **Discussions**: https://github.com/yao/wwsheng009/mint/discussions

---

**最后更新**: 2026-01-31  
**文档版本**: v1.0
