# TimeTravel - 时间旅行模块

> 帧快照、状态回放、时间导航、书签管理

## 功能概述

TimeTravel 模块提供时间旅行功能，允许开发者导航到任意历史帧、查看应用状态、理解状态演变过程。

## 核心组件

### 1. Frame Snapshot (`snapshot.go`)

```go
// FrameSnapshot 帧快照
type FrameSnapshot struct {
    FrameID   FrameID
    Timestamp time.Time
    Duration  time.Duration
    State     StateSnapshot
    Events    []EventData
    Metadata  SnapshotMetadata
}

// StateSnapshot 状态快照
type StateSnapshot struct {
    Components map[NodeID]*ComponentState
    Global     *GlobalState
}

// 创建帧快照
func NewFrameSnapshot(frameID FrameID) *FrameSnapshot

// 捕获状态
func (fs *FrameSnapshot) Capture(dt *DevTools) error
```

### 2. Time Travel Cursor (`cursor.go`)

```go
// TimeTravelCursor 时间游标
type TimeTravelCursor struct {
    mu        sync.RWMutex
    snapshots []*FrameSnapshot
    position  int
    bookmarks map[string]int
}

// 创建游标
func NewTimeTravelCursor() *TimeTravelCursor

// 添加快照
func (tc *TimeTravelCursor) AddSnapshot(snap *FrameSnapshot)

// 移动到指定帧
func (tc *TimeTravelCursor) MoveTo(frameID FrameID) error

// 前进/后退
func (tc *TimeTravelCursor) Next() error
func (tc *TimeTravelCursor) Prev() error

// 跳转到开头/结尾
func (tc *TimeTravelCursor) First() error
func (tc *TimeTravelCursor) Last() error

// 添加书签
func (tc *TimeTravelCursor) AddBookmark(name string) error

// 跳转到书签
func (tc *TimeTravelCursor) GoToBookmark(name string) error

// 获取当前位置
func (tc *TimeTravelCursor) Current() (*FrameSnapshot, error)
```

### 3. State Replay (`replay.go`)

```go
// StateReplay 状态回放器
type StateReplay struct {
    mu      sync.RWMutex
    cursor  *TimeTravelCursor
    speed   float64  // 回放速度
    playing bool
}

// 创建回放器
func NewStateReplay(cursor *TimeTravelCursor) *StateReplay

// 播放
func (sr *StateReplay) Play()

// 暂停
func (sr *StateReplay) Pause()

// 停止
func (sr *StateReplay) Stop()

// 设置速度
func (sr *StateReplay) SetSpeed(speed float64)

// 获取进度
func (sr *StateReplay) Progress() float64  // 0.0 - 1.0
```

### 4. Diff Engine (`diffengine.go`)

```go
// DiffEngine 差异引擎
type DiffEngine struct {
    ignoreProps []string
}

// 创建差异引擎
func NewDiffEngine() *DiffEngine

// 比较两个状态
func (de *DiffEngine) CompareStates(old, new *StateSnapshot) []StateChange

// 计算差异路径
func (de *DiffEngine) DiffPath(from, to FrameID) []StateChange
```

### 5. Time Travel Client (`client.go`)

```go
// TimeTravelClient TUI 客户端
type TimeTravelClient struct {
    cursor   *TimeTravelCursor
    replay   *StateReplay
    engine   *DiffEngine
}

// 创建客户端
func NewTimeTravelClient() *TimeTravelClient

// 渲染当前状态
func (c *TimeTravelClient) Render() string

// 处理输入
func (c *TimeTravelClient) HandleInput(key rune) bool
```

## 使用方法

### 基本时间旅行

```go
import "github.com/wwsheng009/mint/devtools/timetravel"

// 创建游标
cursor := timetravel.NewTimeTravelCursor()

// 在应用循环中捕获快照
for frame := 0; frame < 100; frame++ {
    dt.BeginFrame()
    // ... 应用逻辑 ...
    dt.EndFrame()

    // 每 10 帧捕获一次
    if frame % 10 == 0 {
        snap := timetravel.NewFrameSnapshot(devtools.FrameID(frame))
        snap.Capture(dt)
        cursor.AddSnapshot(snap)
    }
}

// 导航到特定帧
cursor.MoveTo(devtools.FrameID(50))
current, _ := cursor.Current()
fmt.Printf("Frame %d has %d components\n",
    current.FrameID, len(current.State.Components))
```

### 书签管理

```go
// 添加书签
cursor.AddBookmark("bug-state")
cursor.AddBookmark("before-crash")

// 跳转到书签
cursor.GoToBookmark("bug-state")

// 列出所有书签
for name := range cursor.GetBookmarks() {
    fmt.Println(name)
}
```

### 状态回放

```go
// 创建回放器
replay := timetravel.NewStateReplay(cursor)

// 播放所有帧
replay.Play()

// 控制播放
replay.SetSpeed(2.0)  // 2倍速
replay.Pause()
replay.Play()

// 获取进度
progress := replay.Progress()
fmt.Printf("Progress: %.1f%%\n", progress*100)
```

### 状态差异

```go
// 创建差异引擎
engine := timetravel.NewDiffEngine()

// 比较两个帧
oldSnap, _ := cursor.GetSnapshot(0)
newSnap, _ := cursor.GetSnapshot(50)

changes := engine.CompareStates(oldSnap.State, newSnap.State)

for _, change := range changes {
    fmt.Printf("%s: %s %s\n", change.NodeID, change.ChangeType, change.Path)
}
```

### TUI 客户端

```go
// 创建时间旅行客户端
client := timetravel.NewTimeTravelClient()
client.SetCursor(cursor)

// 渲染
fmt.Println(client.Render())

// 处理键盘输入
client.HandleInput('n')  // 下一帧
client.HandleInput('p')  // 上一帧
client.HandleInput('f')  // 第一帧
client.HandleInput('l')  // 最后一帧
client.HandleInput('b')  // 添加书签
client.HandleInput('g')  // 跳转到书签
client.HandleInput(' ')  // 播放/暂停
client.HandleInput('+')  // 加速
client.HandleInput('-')  // 减速
```

## TUI 界面

```
┌─────────────────────────────────────────────────────────────────┐
│                      Time Travel Client                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Frame: 50 / 100 [▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░]                │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Component: button-submit                                │   │
│  │ Type: Button                                            │   │
│  │ Props: {label: "Submit", disabled: false}               │   │
│  │ State: {hovered: true, clicked: false}                  │   │
│  │ Bounds: {x: 10, y: 5, w: 15, h: 1}                      │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Controls: [n]ext [p]rev [f]irst [l]ast [b]ookmark [g]oto      │
│  Playback: [ ] play/pause [+] speed [-] slow                 │
│                                                                  │
│  Bookmarks: bug-state, before-crash, test-result             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 键盘快捷键

| 按键 | 功能 |
|------|------|
| `n` | 下一帧 |
| `p` | 上一帧 |
| `f` | 跳转到第一帧 |
| `l` | 跳转到最后一帧 |
| `b` | 添加书签 |
| `g` | 跳转到书签 |
| `Space` | 播放/暂停 |
| `+` | 加速 |
| `-` | 减速 |
| `q` | 退出 |

## 相关模块

| 模块 | 关系 |
|------|------|
| `devtools` | 核心功能，捕获其状态 |
| `snapshot` | 快照系统，用于存储和比较 |
| `replay` | 事件回放，与时间旅行协同 |
| `client` | TUI 客户端，共享 UI 组件 |

## API 参考

### TimeTravelPosition

```go
type TimeTravelPosition struct {
    Current int
    Total   int
    Offset  int
}
```

### Bookmark

```go
type Bookmark struct {
    Name     string
    FrameID  FrameID
    Created  time.Time
}
```

### ReplayOptions

```go
type ReplayOptions struct {
    Speed      float64
    Loop       bool
    AutoStop   bool
    StopFrame  FrameID
}
```

## 文件列表

- `snapshot.go` - 帧快照
- `cursor.go` - 时间游标
- `replay.go` - 状态回放
- `diffengine.go` - 差异引擎
- `client.go` - TUI 客户端
