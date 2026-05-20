# Phase 6: 快照系统实施总结

> **Snapshot System** - 完整状态捕获、差异比较、时间旅行

## 概述

Phase 6 实现了快照系统，允许在任意帧捕获完整的 TUI 应用状态，并支持高效的状态差异比较和时间旅行导航。

## 核心组件

### 1. 快照数据结构 (`snapshot/snapshot.go`)

```go
// Snapshot 完整的 TUI 状态快照
type Snapshot struct {
    ID        SnapshotID          // 唯一标识符
    FrameID   devtools.FrameID    // 关联的帧ID
    Timestamp time.Time           // 捕获时间
    Metadata  SnapshotMetadata    // 元数据
    States    map[devtools.NodeID]*ComponentState  // 组件状态
    Global    GlobalState         // 全局状态
}

// ComponentState 组件状态
type ComponentState struct {
    NodeID   devtools.NodeID              // 组件ID
    Type     string                       // 组件类型
    Props    map[string]interface{}       // 组件属性
    State    map[string]interface{}       // 组件内部状态
    Bounds   Rect                         // 位置和大小
    Children []devtools.NodeID            // 子组件
    Visible  bool                         // 是否可见
    Focused  bool                         // 是否聚焦
}

// SnapshotDiff 两个快照之间的差异
type SnapshotDiff struct {
    FromID    SnapshotID      // 起始快照ID
    ToID      SnapshotID      // 目标快照ID
    Timestamp time.Time       // 差异计算时间
    Changes   []StateChange   // 变更列表
    Summary   DiffSummary     // 差异摘要
}
```

### 2. 快照管理器 (`snapshot/manager.go`)

```go
// Manager 快照生命周期管理
type Manager struct {
    mu           sync.RWMutex
    snapshots    map[devtools.FrameID]*Snapshot
    maxSnapshots int                    // 最大快照数量
    persistDir   string                 // 持久化目录
}

// 主要方法
func (m *Manager) Capture(frameID devtools.FrameID, builder *Builder) (*Snapshot, error)
func (m *Manager) Get(frameID devtools.FrameID) (*Snapshot, bool)
func (m *Manager) GetRange(from, to devtools.FrameID) []*Snapshot
func (m *Manager) GetAll() []*Snapshot
func (m *Manager) Delete(id SnapshotID) bool
func (m *Manager) Clear()
func (m *Manager) Save(path string) error
func (m *Manager) Load(path string) error
```

### 3. 差异引擎 (`snapshot/diff.go`)

```go
// Differ 快照比较器
type Differ struct {
    ignoreProps   []string
    compareStyle  bool
    compareBounds bool
}

// StateChange 状态变更
type StateChange struct {
    NodeID     devtools.NodeID
    ChangeType ChangeType     // Added, Removed, Modified
    Path       string         // 变更路径 (如 "props.label")
    OldValue   interface{}
    NewValue   interface{}
}

// 时间旅行范围
type TimeTravelRange struct {
    snapshots    []*Snapshot
    changesByNode map[devtools.NodeID][]*StateChange
}

// 主要方法
func (d *Differ) Compare(from, to *Snapshot) *SnapshotDiff
func (d *Differ) ComputeStateChanges(old, new map[devtools.NodeID]*ComponentState) []StateChange
```

## 使用示例

### 基本快照操作

```go
import "github.com/wwsheng009/mint/devtools/snapshot"

// 创建管理器 (最多保留1000个快照)
mgr := snapshot.NewManager(1000)

// 捕获快照
builder := snapshot.NewBuilder("snap-1", devtools.FrameID(42))
builder.SetWindowSize(80, 24)
builder.SetLabel("session", "user-123")

// 添加组件状态
builder.AddComponent(&snapshot.ComponentState{
    NodeID:  "button-submit",
    Type:    "Button",
    Visible: true,
    Focused: true,
    Props: map[string]interface{}{
        "label": "Submit",
        "disabled": false,
    },
    State: map[string]interface{}{
        "hovered": true,
    },
    Bounds: snapshot.Rect{X: 10, Y: 5, Width: 15, Height: 1},
})

snap, err := mgr.Capture(devtools.FrameID(42), builder)
```

### 差异比较

```go
// 获取两个快照
snap1, _ := mgr.Get(devtools.FrameID(0))
snap9, _ := mgr.Get(devtools.FrameID(9))

// 比较差异
differ := snapshot.NewDiffer()
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
    fmt.Printf("Frame %d: %s = %v\n", event.FrameID, event.Path, event.Value)
}

// 查找特定状态
frames := range.FindState("button-submit", "props.label", "Submit")
```

### 快照持久化

```go
// 保存到文件
mgr.Save("/tmp/snapshots.json")

// 从文件加载
mgr.Load("/tmp/snapshots.json")
```

## 数据结构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Snapshot                               │
├─────────────────────────────────────────────────────────────┤
│  ID: "snap-1"                                               │
│  FrameID: 42                                                │
│  Timestamp: 2024-01-30T10:00:00Z                            │
├─────────────────────────────────────────────────────────────┤
│  States: map[NodeID]*ComponentState                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ "button-submit" → ComponentState                     │   │
│  │   - Type: "Button"                                  │   │
│  │   - Props: {label: "Submit", disabled: false}       │   │
│  │   - State: {hovered: true}                          │   │
│  │   - Bounds: {X:10, Y:5, W:15, H:1}                 │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ "input-username" → ComponentState                    │   │
│  │   - Type: "Input"                                   │   │
│  │   - Props: {value: "user@example.com"}              │   │
│  │   - State: {focused: true}                          │   │
│  └─────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│  Global: GlobalState                                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │   WindowSize: {Width: 80, Height: 24}               │   │
│  │   Cursor: {X: 25, Y: 5}                             │   │
│  │   Focused: "button-submit"                          │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    SnapshotDiff                              │
├─────────────────────────────────────────────────────────────┤
│  FromID: "snap-0"                                            │
│  ToID: "snap-9"                                              │
├─────────────────────────────────────────────────────────────┤
│  Changes:                                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ [0] NodeID: "button-submit"                         │   │
│  │     Type: Modified                                  │   │
│  │     Path: "state.clicked"                           │   │
│  │     OldValue: false                                 │   │
│  │     NewValue: true                                  │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │ [1] NodeID: "button-submit"                         │   │
│  │     Type: Modified                                  │   │
│  │     Path: "bounds.y"                                │   │
│  │     OldValue: 5                                     │   │
│  │     NewValue: 9                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 性能特性

| 特性 | 实现方式 |
|------|---------|
| **增量捕获** | 只记录变化的组件 |
| **内存池** | Builder 使用对象池减少分配 |
| **懒加载** | 大型数据按需加载 |
| **压缩存储** | 可选的 JSON 压缩 |

## API 参考

### Builder API

```go
// 创建 Builder
builder := snapshot.NewBuilder(id, frameID)

// 设置全局状态
builder.SetWindowSize(width, height)
builder.SetCursor(x, y)
builder.SetFocused(nodeID)
builder.SetLabel(key, value)

// 添加组件
builder.AddComponent(state)
builder.AddComponents([]*ComponentState{...})

// 构建快照
snap := builder.Build()
```

### Manager API

```go
// 创建管理器
mgr := snapshot.NewManager(maxSnapshots)
mgr.SetPersistDir(dir)

// 快照生命周期
snap, err := mgr.Capture(frameID, builder)
snap, ok := mgr.Get(frameID)
snapshots := mgr.GetRange(from, to)
all := mgr.GetAll()

// 删除操作
deleted := mgr.Delete(id)
mgr.Clear()

// 持久化
mgr.Save(path)
mgr.Load(path)

// 统计
stats := mgr.GetStats()
// Stats.TotalSnapshots, Stats.MaxSnapshots, Stats.MemoryUsage
```

### Differ API

```go
// 创建 Differ
differ := snapshot.NewDiffer()
differ.SetIgnoreProps([]string{"__internal"})
differ.SetCompareStyle(true)
differ.SetCompareBounds(true)

// 比较快照
diff := differ.Compare(from, to)

// 结果分析
changes := diff.GetChangesByNode(nodeID)
added := diff.GetAdded()
removed := diff.GetRemoved()
modified := diff.GetModified()
```

## 测试

```bash
cd devtools/snapshot
go test -v
```

测试覆盖：
- [x] 快照构建
- [x] 组件状态比较
- [x] 差异计算
- [x] 时间旅行范围
- [x] 管理器 CRUD

## 集成示例

```go
// 在 TUI 应用中集成
type App struct {
    dt      *devtools.DevTools
    snapMgr *snapshot.Manager
}

func (a *App) Update() {
    a.dt.BeginFrame()
    defer a.dt.EndFrame()

    // ... 应用逻辑 ...

    // 每10帧捕获一次快照
    if a.dt.CurrentFrame() % 10 == 0 {
        a.captureSnapshot()
    }
}

func (a *App) captureSnapshot() {
    builder := snapshot.NewBuilder(
        snapshot.SnapshotID(fmt.Sprintf("frame-%d", a.dt.CurrentFrame())),
        a.dt.CurrentFrame(),
    )

    // 遍历所有可见组件
    for _, comp := range a.visibleComponents() {
        builder.AddComponent(&snapshot.ComponentState{
            NodeID:  comp.ID(),
            Type:    comp.Type(),
            Props:   comp.Props(),
            State:   comp.State(),
            Bounds:  comp.Bounds(),
            Visible: comp.IsVisible(),
            Focused: comp.IsFocused(),
        })
    }

    a.snapMgr.Capture(a.dt.CurrentFrame(), builder)
}
```

## 状态

✅ **已完成**
- [x] 核心数据结构
- [x] Builder 模式
- [x] Manager 生命周期管理
- [x] Differ 差异引擎
- [x] TimeTravelRange
- [x] 持久化支持
- [x] 单元测试
