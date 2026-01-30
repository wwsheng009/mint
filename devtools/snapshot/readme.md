# Snapshot - 快照系统模块

> 状态捕获、差异比较、时间旅行导航

## 功能概述

Snapshot 模块提供完整的 TUI 应用状态捕获功能，支持高效的状态差异比较和时间旅行导航。

## 核心组件

### 1. Snapshot (`snapshot.go`)

```go
// Snapshot 完整的 TUI 状态快照
type Snapshot struct {
    ID        SnapshotID
    FrameID   devtools.FrameID
    Timestamp time.Time
    Metadata  SnapshotMetadata
    States    map[devtools.NodeID]*ComponentState
    Global    GlobalState
}

// ComponentState 组件状态
type ComponentState struct {
    NodeID   devtools.NodeID
    Type     string
    Props    map[string]interface{}
    State    map[string]interface{}
    Bounds   Rect
    Children []devtools.NodeID
    Visible  bool
    Focused  bool
}

// GlobalState 全局状态
type GlobalState struct {
    WindowSize   Size
    Cursor       Position
    FocusedNode  devtools.NodeID
}

// Builder 快照构建器
type Builder struct {
    snapshot *Snapshot
}

// 创建构建器
func NewBuilder(id SnapshotID, frameID devtools.FrameID) *Builder

// 添加组件
func (b *Builder) AddComponent(state *ComponentState) *Builder

// 设置全局状态
func (b *Builder) SetWindowSize(w, h int) *Builder

// 构建快照
func (b *Builder) Build() *Snapshot
```

### 2. Manager (`manager.go`)

```go
// Manager 快照管理器
type Manager struct {
    mu           sync.RWMutex
    snapshots    map[devtools.FrameID]*Snapshot
    maxSnapshots int
    persistDir   string
}

// 创建管理器
func NewManager(maxSnapshots int) *Manager

// 捕获快照
func (m *Manager) Capture(frameID devtools.FrameID, builder *Builder) (*Snapshot, error)

// 获取快照
func (m *Manager) Get(frameID devtools.FrameID) (*Snapshot, bool)

// 获取范围
func (m *Manager) GetRange(from, to devtools.FrameID) []*Snapshot

// 获取所有快照
func (m *Manager) GetAll() []*Snapshot

// 删除快照
func (m *Manager) Delete(id SnapshotID) bool

// 清空所有快照
func (m *Manager) Clear()

// 保存到文件
func (m *Manager) Save(path string) error

// 从文件加载
func (m *Manager) Load(path string) error

// 获取统计
func (m *Manager) GetStats() ManagerStats
```

### 3. Differ (`diff.go`)

```go
// Differ 快照比较器
type Differ struct {
    ignoreProps   []string
    compareStyle  bool
    compareBounds bool
}

// 创建比较器
func NewDiffer() *Differ

// 比较两个快照
func (d *Differ) Compare(from, to *Snapshot) *SnapshotDiff

// StateChange 状态变更
type StateChange struct {
    NodeID     devtools.NodeID
    ChangeType ChangeType
    Path       string
    OldValue   interface{}
    NewValue   interface{}
}

// ChangeType 变更类型
type ChangeType int

const (
    ChangeTypeAdded ChangeType = iota
    ChangeTypeRemoved
    ChangeTypeModified
)

// SnapshotDiff 快照差异
type SnapshotDiff struct {
    FromID    SnapshotID
    ToID      SnapshotID
    Timestamp time.Time
    Changes   []StateChange
    Summary   DiffSummary
}
```

### 4. Time Travel Range (`diff.go`)

```go
// TimeTravelRange 时间旅行范围
type TimeTravelRange struct {
    snapshots    []*Snapshot
    changesByNode map[devtools.NodeID][]*StateChange
}

// 创建时间旅行范围
func NewTimeTravelRange(snapshots []*Snapshot) *TimeTravelRange

// 计算变更
func (r *TimeTravelRange) Compute() error

// 获取组件的变更历史
func (r *TimeTravelRange) GetChangeHistory(nodeID devtools.NodeID) []*StateChange

// 查找特定状态
func (r *TimeTravelRange) FindState(nodeID devtools.NodeID, path, value interface{}) []devtools.FrameID
```

## 使用方法

### 捕获快照

```go
import "github.com/wwsheng009/mint/devtools/snapshot"

// 创建管理器
mgr := snapshot.NewManager(1000)

// 创建快照
builder := snapshot.NewBuilder("snap-1", devtools.FrameID(42))
builder.SetWindowSize(80, 24)
builder.SetCursor(10, 5)

// 添加组件状态
builder.AddComponent(&snapshot.ComponentState{
    NodeID:  "button-submit",
    Type:    "Button",
    Props: map[string]interface{}{
        "label": "Submit",
    },
    State: map[string]interface{}{
        "hovered": true,
    },
    Bounds: snapshot.Rect{X: 10, Y: 5, Width: 15, Height: 1},
})

// 捕获
snap, err := mgr.Capture(devtools.FrameID(42), builder)
```

### 比较快照

```go
// 获取两个快照
snap1, _ := mgr.Get(devtools.FrameID(0))
snap9, _ := mgr.Get(devtools.FrameID(9))

// 创建比较器
differ := snapshot.NewDiffer()

// 比较差异
diff := differ.Compare(snap1, snap9)

// 查看结果
fmt.Printf("Changes: %d\n", len(diff.Changes))
fmt.Printf("Summary: %s\n", diff.FormatSummary())

// 遍历变更
for _, change := range diff.Changes {
    fmt.Printf("%s: %s %s\n", change.NodeID, change.ChangeType, change.Path)
}
```

### 时间旅行

```go
// 创建时间旅行范围
range := snapshot.NewTimeTravelRange(mgr.GetAll())
range.Compute()

// 查看组件的完整变更历史
history := range.GetChangeHistory("button-submit")
for _, event := range history {
    fmt.Printf("Frame %d: %s = %v\n",
        event.FrameID, event.Path, event.NewValue)
}

// 查找特定状态
frames := range.FindState("button-submit", "props.label", "Submit")
fmt.Printf("Button had 'Submit' label in frames: %v\n", frames)
```

### 持久化

```go
// 保存到文件
mgr.Save("/tmp/snapshots.json")

// 从文件加载
mgr.Load("/tmp/snapshots.json")
```

## 数据结构图

```
Snapshot
├── ID: "snap-1"
├── FrameID: 42
├── Timestamp: 2024-01-30T10:00:00Z
├── States: map[NodeID]*ComponentState
│   ├── "button-1" → ComponentState
│   └── "input-1" → ComponentState
└── Global: GlobalState
    ├── WindowSize: {80, 24}
    ├── Cursor: {10, 5}
    └── FocusedNode: "button-1"
```

## 相关模块

| 模块 | 关系 |
|------|------|
| `devtools` | 核心功能，快照捕获其状态 |
| `memory` | 内存优化，管理快照存储 |
| `remote` | 远程调试，通过 API 提供快照 |
| `replay` | 事件回放，使用快照验证 |

## API 参考

### ManagerStats

```go
type ManagerStats struct {
    TotalSnapshots int
    MaxSnapshots   int
    MemoryUsage    int64
}
```

### DiffSummary

```go
type DiffSummary struct {
    Added     int
    Removed   int
    Modified  int
    Props     int
    State     int
    Bounds    int
}
```

### SnapshotMetadata

```go
type SnapshotMetadata struct {
    Labels            map[string]string
    FramesSinceLast   int
    MutationsCount    int
    LayoutsCount      int
    RepaintsCount     int
}
```

## 文件列表

- `snapshot.go` - 快照数据结构
- `manager.go` - 快照管理器
- `diff.go` - 差异引擎和时间旅行
- `snapshot_test.go` - 单元测试
