# Protocol - 协议模块

> 类型定义、常量、通用接口

## 功能概述

Protocol 模块定义 DevTools 的核心类型系统和常量，是整个 DevTools 的基础模块。

## 核心类型

### 节点标识符

```go
// NodeID 节点唯一标识符
type NodeID string

// FrameID 帧标识符
type FrameID int

// MutationID 变更标识符
type MutationID uint64
```

### Delta 类型

```go
// EventDelta 事件增量
type EventDelta struct {
    Type      string
    Count     int
    FirstTime time.Time
    LastTime  time.Time
}

// LayoutDelta 布局增量
type LayoutDelta struct {
    Component NodeID
    OldBounds Rect
    NewBounds Rect
    Reason    string
}

// RepaintDelta 重绘增量
type RepaintDelta struct {
    DirtyRegions []Rect
    ChangedCells int
    TotalCells   int
}
```

### 矩形和边界

```go
// Rect 矩形区域
type Rect struct {
    X      int
    Y      int
    Width  int
    Height int
}

// Bounds 边界框
type Bounds struct {
    Left   int
    Top    int
    Right  int
    Bottom int
}
```

### 组件状态

```go
// ComponentState 组件状态
type ComponentState struct {
    NodeID   NodeID
    Type     string
    Props    map[string]interface{}
    State    map[string]interface{}
    Bounds   Rect
    Children []NodeID
    Visible  bool
    Focused  bool
}
```

## 常量定义

### 事件类型

```go
const (
    EventTypeKeypress    = "keypress"
    EventTypeMouse       = "mouse"
    EventTypeResize      = "resize"
    EventTypeFocus       = "focus"
    EventTypeBlur        = "blur"
    EventTypeCustom      = "custom"
)
```

### 变更类型

```go
const (
    ChangeTypeAdded    = "added"
    ChangeTypeRemoved  = "removed"
    ChangeTypeModified = "modified"
)
```

### 观察级别

```go
const (
    LevelNone      = 0
    LevelBasic     = 1
    LevelEnhanced  = 2
    LevelAdvanced  = 3
)
```

## 辅助函数

### 节点 ID 生成

```go
// GenerateNodeID 生成节点 ID
func GenerateNodeID(prefix string) NodeID {
    return NodeID(fmt.Sprintf("%s-%d", prefix, atomic.AddUint64(&nodeCounter, 1)))
}
```

### 矩形操作

```go
// Contains 检查点是否在矩形内
func (r Rect) Contains(x, y int) bool

// Intersects 检查两个矩形是否相交
func (r Rect) Intersects(other Rect) bool

// Union 计算两个矩形的并集
func (r Rect) Union(other Rect) Rect

// Area 计算矩形面积
func (r Rect) Area() int
```

### 边界操作

```go
// ToRect 转换为矩形
func (b Bounds) ToRect() Rect

// Contains 检查点是否在边界内
func (b Bounds) Contains(x, y int) bool

// Expand 扩展边界
func (b *Bounds) Expand(margin int)
```

## 类型转换

```go
// NodeID 到字符串
func (n NodeID) String() string

// FrameID 到整数
func (f FrameID) Int() int

// 整数到 FrameID
func ToFrameID(i int) FrameID
```

## 使用示例

### 创建节点 ID

```go
import "github.com/wwsheng009/mint/devtools"

// 生成唯一节点 ID
buttonID := devtools.GenerateNodeID("button")
// button-1, button-2, ...

inputID := devtools.GenerateNodeID("input")
// input-1, input-2, ...
```

### 使用矩形

```go
// 创建矩形
rect := devtools.Rect{
    X: 10, Y: 5,
    Width: 20, Height: 1,
}

// 检查点是否在矩形内
if rect.Contains(15, 5) {
    fmt.Println("Point is inside")
}

// 检查相交
other := devtools.Rect{X: 5, Y: 3, Width: 10, Height: 5}
if rect.Intersects(other) {
    fmt.Println("Rectangles intersect")
}

// 计算面积
area := rect.Area()  // 20
```

### 组件状态

```go
state := &devtools.ComponentState{
    NodeID: "button-1",
    Type: "Button",
    Props: map[string]interface{}{
        "label": "Click me",
        "disabled": false,
    },
    State: map[string]interface{}{
        "hovered": true,
    },
    Bounds: devtools.Rect{X: 10, Y: 5, Width: 15, Height: 1},
    Visible: true,
    Focused: false,
}
```

## 相关模块

| 模块 | 关系 |
|------|------|
| **所有模块** | Protocol 是基础模块，所有模块都依赖其类型定义 |
| `devtools` | 使用 NodeID、FrameID 等核心类型 |
| `snapshot` | 使用 ComponentState、Rect 等结构 |
| `causal` | 使用事件类型常量 |
| `client` | 使用所有类型进行展示 |

## API 参考

### 时间戳

```go
type Timestamp struct {
    FrameID   FrameID
    Time      time.Time
    Duration  time.Duration
}
```

### 事件数据

```go
type EventData struct {
    Type      string
    NodeID    NodeID
    Phase     string  // "capture", "bubble", "target"
    Timestamp time.Time
    Data      map[string]interface{}
}
```

## 文件列表

- `types.go` - 核心类型定义
- `tap.go` - Mutation Tap 接口
- `runtime_adapter.go` - Runtime 适配器
