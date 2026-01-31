# Mint Runtime

Mint Runtime 是一个纯 Go 实现的 TUI 运行时内核，提供布局、事件、渲染等核心功能。

## 目录

- [概述](#概述)
- [设计原则](#设计原则)
- [核心模块](#核心模块)
- [架构](#架构)
- [使用指南](#使用指南)
- [扩展开发](#扩展开发)

## 概述

Mint Runtime 是 Mint 框架的运行时层，负责：

- **布局计算** - Flexbox 布局算法，支持绝对/相对定位
- **事件处理** - 三阶段事件传播（捕获→目标→冒泡）
- **焦点管理** - 几何焦点导航和焦点陷阱
- **渲染引擎** - Z-Index 渲染、Diff 优化、宽字符支持
- **调试支持** - Debug ID 系统、布局版本追踪

### 纯 Go 约束

Runtime 层**必须保持纯 Go 实现**，不能依赖外部库：

- ❌ Bubble Tea
- ❌ lipgloss（仅 `runtime/render/` 可用）
- ❌ DSL 解析器
- ❌ 具体组件

此约束确保运行时可独立使用，便于测试和优化。

## 设计原则

### 1. 纯内核架构

Runtime 是纯内核，专注于基础功能：

```
┌─────────────────────────────────────────────────────┐
│                   Framework Layer                    │
│  (app.go, Components, Themes, DSL)                   │
└─────────────────────────────────────────────────────┘
                        │
                        └── ─── 使用 ─── ┤
                                       ▼
┌─────────────────────────────────────────────────────┐
│                   Runtime Layer                      │
│  (布局、事件、渲染、焦点 - 纯 Go)                      │
└─────────────────────────────────────────────────────┘
```

### 2. 能力接口

使用小而专一的接口（基于能力），而非大型接口：

```go
// 好：小接口，职责单一
type Measurable interface {
    Measure(constraints BoxConstraints) Size
}

type Layoutable interface {
    Layout(constraints BoxConstraints)
}

type Paintable interface {
    Paint(buffer *Buffer, x, y int)
}
```

### 3. 类型安全

使用命名类型而非字符串常量：

```go
// 好：类型安全
type NodeType string
const (
    NodeTypeText NodeType = "text"
    NodeTypeBox  NodeType = "box"
)

// 避免：字符串比较
if nodeType == "text" { ... }
```

### 4. 无外部依赖

Runtime 层不依赖任何第三方库，确保：

- 可独立测试
- 易于集成到其他项目
- 性能可控
- 无版本冲突

## 核心模块

### 1. Layout (布局系统)

**位置**: `runtime/layout/`

**职责**:
- Flexbox 布局算法
- 约束计算和传递
- 尺寸测量和分配

**核心类型**:

```go
type Engine struct {
    cache *Cache
    stats *Stats
}

type Constraints struct {
    MinWidth, MaxWidth   int
    MinHeight, MaxHeight int
}

type LayoutResult struct {
    Boxes []LayoutBox
}
```

**关键文件**:
- `engine.go` - 布局引擎主逻辑
- `flex.go` - Flexbox 算法实现
- `cache.go` - 布局结果缓存
- `types.go` - 布局相关类型定义

### 2. Event (事件系统)

**位置**: `runtime/event/`

**职责**:
- 三阶段事件传播
- 事件过滤和拦截
- 命中测试

**核心类型**:

```go
type Event interface {
    Type() EventType
    Phase() EventPhase
    Target() Node
    CurrentTarget() Node
    PreventDefault()
    StopPropagation()
}

type EventPhase int
const (
    PhaseNone EventPhase = iota
    PhaseCapture
    PhaseTarget
    PhaseBubble
)
```

**关键文件**:
- `event.go` - 事件接口和 BaseEvent
- `dispatch.go` - 事件分发器
- `filter.go` - 事件过滤器
- `hittest.go` - 命中测试

### 3. Focus (焦点管理)

**位置**: `runtime/focus/`

**职责**:
- 焦点栈管理
- 焦点导航（方向键、Tab）
- 焦点陷阱

**核心类型**:

```go
type Manager struct {
    stack      []*Focusable
    root       *LayoutNode
    navigation Navigation
}

type Focusable interface {
    SetFocus(bool)
    IsFocusable() bool
}
```

**关键文件**:
- `manager.go` - 焦点管理器
- `geometric.go` - 几何焦点导航
- `trap.go` - 焦点陷阱

### 4. Paint (绘制系统)

**位置**: `runtime/paint/`

**职责**:
- CellBuffer 管理
- Z-Index 渲染顺序
- Diff 优化
- 宽字符支持

**核心类型**:

```go
type Buffer struct {
    Width  int
    Height int
    Cells  [][]Cell
}

type Cell struct {
    Cluster string
    Style   Style
    ZIndex  int
    NodeID  string
}
```

**关键文件**:
- `buffer.go` - CellBuffer 实现
- `cell.go` - Cell 类型定义
- `renderer.go` - 渲染器
- `diff.go` - Diff 算法
- `wide_char_*.go` - 宽字符支持

**文档**:
- `WIDE_CHAR_GUIDE.md` - 宽字符处理指南
- `BUGFIX_diff_rendering.md` - Diff 渲染修复说明

### 5. State (状态管理)

**位置**: `runtime/state/`

**职责**:
- 状态定义和序列化
- 状态差异计算
- 状态快照

**核心类型**:

```go
type State interface{}

type Tracker struct {
    previous State
    changes  []StateChange
}

type StateChange struct {
    Field    string
    OldValue interface{}
    NewValue interface{}
}
```

**关键文件**:
- `tracker.go` - 状态追踪器
- `diff.go` - 状态差异计算
- `serialize.go` - 状态序列化
- `snapshot.go` - 状态快照

### 6. Action (Action 处理)

**位置**: `runtime/action/`

**职责**:
- Action 接口定义
- Action 路由和分发
- Composite Action（组合操作）

**核心类型**:

```go
type Action interface {
    Type() string
    Payload() interface{}
}

type Interface interface {
    HandleAction(action Action) *Result
}
```

**关键文件**:
- `action.go` - Action 接口定义
- `dispatcher.go` - Action 分发器
- `composite.go` - Composite Action

### 7. Animation (动画)

**位置**: `runtime/animation/`

**职责**:
- 动画状态管理
- 缓动函数 (Easing)
- 帧计算

**核心类型**:

```go
type Manager struct {
    animations []*Animation
}

type EasingFunc func(t float64) float64
```

**关键文件**:
- `manager.go` - 动画管理器
- `easing.go` - 缓动函数集合
- `types.go` - 动画类型定义
- `builders.go` - 动画构建器

### 8. Input (输入处理)

**位置**: `runtime/input/`

**职责**:
- 输入事件抽象
- 键盘映射
- 鼠标追踪

**核心类型**:

```go
type InputReader interface {
    ReadInput() ([]byte, error)
}

type KeyMap struct {
    bindings map[string]KeyBinding
}
```

**关键文件**:
- `reader.go` - 输入读取器
- `keymap.go` - 键盘映射
- `mouse_tracker.go` - 鼠标追踪

### 9. Platform (平台抽象)

**位置**: `runtime/platform/`

**职责**:
- 终端控制
- 屏幕管理
- 光标控制
- 信号处理

**核心类型**:

```go
type Platform interface {
    Init() error
    Close() error
    Size() (w, h int)
    SetCursor(x, y int)
    ClearScreen()
    ListenSignals() <-chan Signal
}
```

**关键文件**:
- `platform.go` - 平台接口定义
- `terminal.go` - 终端控制
- `cursor.go` - 光标控制
- `signal.go` - 信号处理

### 10. Scheduler (调度器)

**位置**: `runtime/scheduler/`

**职责**:
- 更新批处理
- 优先级处理
- 时间切片渲染

**核心类型**:

```go
type Scheduler struct {
    dirtyNodes map[priority.DirtyLevel][]*DirtyNode
    batch      *UpdateBatch
}

type DirtyLevel int
const (
    DirtyHigh DirtyLevel = iota
    DirtyNormal
    DirtyLow
)
```

**关键文件**:
- `scheduler.go` - 调度器主逻辑
- 文档: `SCHEDULER_INTEGRATION.md` - 集成指南

### 11. AI (AI 集成)

**位置**: `runtime/ai/`

**职责**:
- AI Controller 接口
- AI 组件交互
- 智能辅助功能

**核心类型**:

```go
type Controller interface {
    SuggestAction(context Context) Action
    ExecuteCommand(command string) Result
}
```

**关键文件**:
- `controller.go` - AI 控制器接口
- `operations.go` - AI 操作定义
- `runtime_controller.go` - Runtime 控制器

## 架构

### 数据流

```
┌────────────────────────────────────────────────────────────┐
│                        Input                                │
│                   (键盘/鼠标/信号)                           │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                      Event Dispatch                         │
│             (三阶段传播: Capture → Target → Bubble)          │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                      Action Handling                        │
│              (Action 路由和组件更新)                          │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                        Layout                               │
│              (测量 → 布局 → 缓存 → 收集结果)                  │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                        Paint                                │
│              (Z-Index 渲染 → Diff → 优化)                    │
└────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────┐
│                     Output                                  │
│                 (终端更新/导出)                              │
└────────────────────────────────────────────────────────────┘
```

### 关键类型

#### LayoutNode
UI 中间表示 (IR)，包含布局和渲染所需的所有信息：

```go
type LayoutNode struct {
    ID       string
    Type     NodeType
    Style    Style
    Props    map[string]interface{}
    Position Position

    Component *ComponentRef

    Parent   *LayoutNode
    Children []*LayoutNode

    // Runtime 字段（仅供 Runtime 写）
    X, Y                     int
    MeasuredWidth, MeasuredHeight int
    LayoutVersion           uint32
    layoutDirty, paintDirty bool
}
```

**关键原则**:
- DSL/Builder 只能填写 Type, Style, Props
- Runtime 是**唯一**允许写入 X, Y, MeasuredWidth, MeasuredHeight 的实体
- 组件不能通过 LayoutNode 反向修改布局

#### Debug ID 系统

为避免热路径中的字符串操作，Runtime 提供了 Debug ID 注册表：

```go
// 注册组件类型
id := RegisterComponent("button") // 返回 uint32

// 注册字段名
fieldID := RegisterField("text") // 返回 uint16

// 逆向查询
name := GetComponentName(id)
```

## 使用指南

### 基本示例

```go
package main

import (
    "github.com/wwsheng009/mint/runtime"
    "github.com/wwsheng009/mint/runtime/layout"
    "github.com/wwsheng009/mint/runtime/paint"
)

func main() {
    // 1. 创建布局引擎
    engine := layout.NewEngine()

    // 2. 创建节点树
    root := runtime.NewLayoutNode("root", runtime.NodeTypeBox, runtime.Style{})
    child := runtime.NewLayoutNode("text", runtime.NodeTypeText, runtime.Style{})
    root.AddChild(child)

    // 3. 计算布局
    constraints := runtime.NewBoxConstraints(0, 80, 0, 24)
    result := engine.Layout([]runtime.LayoutNode{*root}, constraints)

    // 4. 渲染
    buffer := paint.NewBuffer(80, 24)
    for _, box := range result.Boxes {
        // 渲染每个 box...
    }

    // 5. 输出
    fmt.Println(buffer.String())
}
```

### DevTools 集成

Runtime 内置对 DevTools 的支持：

```go
//注册组件类型（用于调试）
componentID := runtime.RegisterComponent("button")

// 注册字段
fieldID := runtime.RegisterField("label")

// 使用 Debug ID 进行高效追踪
if dt.IsEnabled() {
    dt.RecordMutation(componentID, fieldID, oldValue, newValue)
}
```

## 扩展开发

### 实现自定义可测量组件

```go
type MyComponent struct {
    content string
}

func (c *MyComponent) Measure(constraints runtime.BoxConstraints) runtime.Size {
    // 计算内容尺寸
    width := len(c.content)
    height := 1
    return runtime.Size{Width: width, Height: height}
}

func (c *MyComponent) Layout(constraints runtime.BoxConstraints) {
    // 布局逻辑...
}

func (c *MyComponent) Paint(buffer *runtime.CellBuffer, x, y int) {
    // 渲染逻辑...
    buffer.SetContent(x, y, 0, rune(c.content[0]), ...)
}
```

### 实现自定义事件处理器

```go
type CustomHandler struct{}

func (h *CustomHandler) HandleEvent(ev event.Event) *event.Result {
    switch ev.Type() {
    case event.EventTypeKeyDown:
        // 处理键盘事件
    case event.EventTypeMouseDown:
        // 处理鼠标事件
    }
    return event.Result{Stop: false}
}

// 注册处理器
dispatcher.RegisterHandler(handler)
```

## 相关文档

- [DevTools 集成指南](../devtools/docs/)
- [布局系统文档](layout/)
- [事件系统文档](event/)
- [绘制系统文档](paint/)
- [选择系统使用指南](SELECTION_USAGE.md)
- [应用引擎对比](engine/APP_ENGINE_COMPARISON.md)

## 测试

```bash
# 运行所有测试
go test ./runtime/...

# 运行特定包测试
go test ./runtime/layout -v
go test ./runtime/event -v
go test ./runtime/paint -v
```

## 贡献

Runtime 遵循严格的纯 Go 约束，确保：

1. 无外部依赖
2. 接口基于能力
3. 类型安全
4. 高性能

## 许可证

MIT License - 详见项目根目录 LICENSE 文件
