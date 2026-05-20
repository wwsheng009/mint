# DevTools 阶段3 实施总结

> 实施日期: 2026-01-30
> 状态: 已完成
> 阶段: 时间旅行 (P2)
> 编译: ✅ 通过
> 测试: ✅ 16/16 通过

## 已完成的工作

### 1. FrameSnapshot 完整状态快照

| 文件 | 功能 | 状态 |
|------|------|------|
| `timetravel/snapshot.go` | `SnapshotManager` 快照管理器 | ✅ |
| `timetravel/snapshot.go` | `FrameSnapshot` 帧快照结构 | ✅ |
| `timetravel/snapshot.go` | `ComponentState` 组件状态 | ✅ |
| `timetravel/snapshot.go` | `LayoutSnapshot` 布局快照 | ✅ |
| `timetravel/snapshot.go` | `RepaintSnapshot` 重绘快照 | ✅ |
| `timetravel/snapshot.go` | `SnapshotBuilder` 快照构建器 | ✅ |
| `timetravel/snapshot.go` | `SnapshotDiff` 快照差异 | ✅ |
| `timetravel/snapshot.go` | JSON 序列化支持 | ✅ |

### 2. TimeTravelCursor 时间游标

| 文件 | 功能 | 状态 |
|------|------|------|
| `timetravel/cursor.go` | `TimeTravelCursor` 游标 | ✅ |
| `timetravel/cursor.go` | `MoveToFrame()` 移动到指定帧 | ✅ |
| `timetravel/cursor.go` | `MoveNext()` / `MovePrev()` 前后移动 | ✅ |
| `timetravel/cursor.go` | `MoveBySteps()` 按步数移动 | ✅ |
| `timetravel/cursor.go` | `JumpToEvent()` 跳转到事件 | ✅ |
| `timetravel/cursor.go` | `Bookmark()` 书签功能 | ✅ |
| `timetravel/cursor.go` | `GetDiffToNext()` 获取下一帧差异 | ✅ |

### 3. StateReplay 状态回放

| 文件 | 功能 | 状态 |
|------|------|------|
| `timetravel/replay.go` | `ReplayEngine` 回放引擎 | ✅ |
| `timetravel/replay.go` | `ReplayFrom()` 从指定帧回放 | ✅ |
| `timetravel/replay.go` | `ReplayRange()` 范围回放 | ✅ |
| `timetravel/replay.go` | `StateApplier` 状态应用接口 | ✅ |
| `timetravel/replay.go` | `ReplaySession` 回放会话 | ✅ |
| `timetravel/replay.go` | 回放速度控制 | ✅ |
| `timetravel/replay.go` | JSON 导出/导入 | ✅ |

### 4. DiffEngine 快照差异引擎

| 文件 | 功能 | 状态 |
|------|------|------|
| `timetravel/diffengine.go` | `DiffEngine` 差异引擎 | ✅ |
| `timetravel/diffengine.go` | `DeltaSet` 增量集合 | ✅ |
| `timetravel/diffengine.go` | `ComponentDelta` 组件差异 | ✅ |
| `timetravel/diffengine.go` | `BufferDiff` 缓冲区差异 | ✅ |
| `timetravel/diffengine.go` | `JSONDiff` JSON 差异 | ✅ |
| `timetravel/diffengine.go` | `Patch()` 增量应用 | ✅ |
| `timetravel/diffengine.go` | `VisualDiff` 可视化差异 | ✅ |

### 5. TimeTravelClient 时间旅行 UI

| 文件 | 功能 | 状态 |
|------|------|------|
| `timetravel/client.go` | `TimeTravelClient` TUI 客户端 | ✅ |
| `timetravel/client.go` | `ViewTimeline` 时间线视图 | ✅ |
| `timetravel/client.go` | `ViewSnapshot` 快照视图 | ✅ |
| `timetravel/client.go` | `ViewDiff` 差异视图 | ✅ |
| `timetravel/client.go` | `ViewReplay` 回放控制视图 | ✅ |
| `timetravel/client.go` | 键盘输入处理 | ✅ |
| `timetravel/client.go` | 状态导出/导入 | ✅ |

---

## 架构设计

### 时间旅行系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    时间旅行系统架构                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                    Application                          │    │
│  └────────────────────────┬───────────────────────────────┘    │
│                           │                                     │
│                           ▼                                     │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                TimeTravelClient (TUI)                   │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ │    │
│  │  │ Timeline │  │ Snapshot │  │   Diff   │  │ Replay │ │    │
│  │  │  View    │  │  View    │  │  View    │  │  View  │ │    │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └───┬────┘ │    │
│  └───────┼────────────┼────────────┼──────────────┼────────┘    │
│          │            │            │              │             │
│  ┌───────┴────────────┴────────────┴──────────────┴────────┐    │
│  │                  TimeTravelCursor                        │    │
│  │  • Navigate frames (next, prev, jump)                   │    │
│  │  • Bookmark management                                  │    │
│  │  • Diff queries                                         │    │
│  └──────────────────────┬───────────────────────────────────┘    │
│                         │                                         │
│  ┌──────────────────────┴───────────────────────────────────┐    │
│  │                  SnapshotManager                          │    │
│  │  ┌──────────────────────────────────────────────────┐   │    │
│  │  │         FrameSnapshot[] (Ring Buffer)             │   │    │
│  │  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐     │   │    │
│  │  │  │Frame 1 │ →│Frame 2 │ →│Frame 3 │ →│Frame N │     │   │    │
│  │  │  └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘     │   │    │
│  │  └──────┼──────────┼──────────┼──────────┼──────────┘   │    │
│  │         │          │          │          │              │    │
│  │  PrevFrame ────────┘          │          │              │    │
│  │  ComponentStates              │          │              │    │
│  │  LayoutState                  │          │              │    │
│  │  RepaintState                 │          │              │    │
│  │  CausalGraph ──────────────────┘          │              │    │
│  │  Events                                   │              │    │
│  └───────────────────────────────────────────────┘              │
│                         │                                         │
│  ┌──────────────────────┴───────────────────────────────────┐    │
│  │                    ReplayEngine                           │    │
│  │  • ReplayFrom() - 从指定帧回放                            │    │
│  │  • ReplayRange() - 范围回放                               │    │
│  │  • SetReplaySpeed() - 速度控制                            │    │
│  │  • StateApplier - 状态应用                                │    │
│  └───────────────────────────────────────────────────────────┘    │
│                         │                                         │
│  ┌──────────────────────┴───────────────────────────────────┐    │
│  │                     DiffEngine                            │    │
│  │  • Compute() - 计算快照差异                               │    │
│  │  • DeltaSet - 增量集合                                    │    │
│  │  • BufferDiff - 缓冲区差异                                │    │
│  │  • VisualDiff - 可视化差异                                │    │
│  └───────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 快照数据结构

```
┌─────────────────────────────────────────────────────────────────┐
│                   FrameSnapshot 结构                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  metadata:                                                       │
│    FrameID     int                                              │
│    Timestamp   time.Time                                        │
│    PrevFrame   *FrameSnapshot  ← Linked list for traversal     │
│                                                                  │
│  causal_data:                                                    │
│    CausalGraph *CausalGraph                                     │
│    Events []CausalEvent                                         │
│                                                                  │
│  component_state:                                                │
│    ComponentStates map[uint32]*ComponentState                   │
│      └── [component_id] → ComponentState                        │
│            ├── ComponentID   uint32                             │
│            ├── ComponentName string                             │
│            ├── State         map[string]interface{}             │
│            ├── Props         map[string]interface{}             │
│            ├── Style         map[string]interface{}             │
│            └── Children      []uint32                           │
│                                                                  │
│  layout_state:                                                   │
│    LayoutState *LayoutSnapshot                                  │
│      └── Nodes map[string]*NodeLayout                          │
│            ├── [node_id] → NodeLayout                           │
│            │     ├── ID      string                             │
│            │     ├── X, Y    int                                │
│            │     ├── Width, Height int                          │
│            │     └── ...                                        │
│                                                                  │
│  repaint_state:                                                  │
│    RepaintState *RepaintSnapshot                                │
│      ├── DirtyRegions []Rect                                    │
│      ├── ChangedCells int                                       │
│      └── Buffer []byte  (可选: 完整缓冲区快照)                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 文件结构

```
mint/
└── devtools/
    ├── types.go              # 核心类型
    ├── causal.go             # 阶段2: Causal Graph
    ├── causal_builder.go     # 阶段2: CausalBuilder
    ├── timeline.go           # 阶段2: FrameTimeline
    ├── causal_query.go       # 阶段2: Query API
    ├── component_hook.go     # 阶段2: Component Hooks
    │
    └── timetravel/           # ✨ 阶段3: 时间旅行
        ├── snapshot.go       # 快照管理
        ├── cursor.go         # 时间游标
        ├── replay.go         # 状态回放
        ├── diffengine.go     # 差异引擎
        └── client.go         # TUI 客户端
```

---

## 使用示例

### 1. 创建快照

```go
import "github.com/wwsheng009/mint/devtools/timetravel"

// 创建快照管理器
mgr := timetravel.NewSnapshotManager(100) // 保留100帧

// 使用构建器创建快照
builder := timetravel.NewSnapshotBuilder(mgr)

builder.BeginSnapshot(devtools.FrameID(1)).
    WithCausalGraph(currentGraph).
    WithLayoutState(layoutSnapshot).
    WithRepaintState(repaintSnapshot).
    Build()
```

### 2. 时间游标导航

```go
// 创建时间游标
cursor := timetravel.NewTimeTravelCursor(mgr)

// 导航操作
cursor.MoveToFirst()
cursor.MoveNext()
cursor.MovePrev()
cursor.MoveToFrame(devtools.FrameID(10))
cursor.MoveBySteps(5) // 前进5帧

// 跳转到特定事件
cursor.JumpToEvent(devtools.EventID(5))
cursor.JumpToComponentChange(componentID)

// 书签
cursor.Bookmark("bug-point")
cursor.GotoBookmark("bug-point")
```

### 3. 状态回放

```go
// 创建回放引擎
engine := timetravel.NewReplayEngine(mgr, cursor)

// 设置回放速度
engine.SetReplaySpeed(2.0) // 2倍速

// 从指定帧回放
engine.ReplayFrom(devtools.FrameID(5))

// 回放范围
engine.ReplayRange(devtools.FrameID(10), devtools.FrameID(20))

// 单步控制
engine.StepForward()
engine.StepBackward()

// 停止回放
engine.Stop()
```

### 4. 差异分析

```go
// 创建差异引擎
diffEngine := timetravel.NewDiffEngine()

// 计算两个快照的差异
snapshot1 := mgr.GetSnapshot(devtools.FrameID(1))
snapshot2 := mgr.GetSnapshot(devtools.FrameID(2))
diff := diffEngine.Compute(snapshot1, snapshot2)

// 查看组件变更
for _, compDiff := range diff.ChangedComponents {
    fmt.Printf("Component %d changed:\n", compDiff.ComponentID)
    for field, change := range compDiff.Changes.Modified {
        fmt.Printf("  %s: %v -> %v\n", field, change.OldValue, change.NewValue)
    }
}

// 获取增量集合
deltaSet := diffEngine.ComputeDeltaSet(diff)
```

### 5. TUI 客户端

```go
// 创建时间旅行客户端
client := timetravel.NewTimeTravelClient(mgr)

// 渲染当前视图
output := client.Render(80, 24)
fmt.Println(output)

// 处理用户输入
client.HandleInput("n")    // 下一帧
client.HandleInput("p")    // 上一帧
client.HandleInput("s")    // 切换到快照视图
client.HandleInput("d")    // 切换到差异视图
client.HandleInput("r")    // 切换到回放视图

// 控制显示选项
client.ToggleLayout()
client.ToggleState()
client.ToggleCausal()
client.ToggleDiff()

// 导出状态
state := client.ExportState()
```

---

## API 快速参考

### SnapshotManager

```go
NewSnapshotManager(maxCount int) *SnapshotManager
AddSnapshot(snapshot *FrameSnapshot)
GetSnapshot(frameID FrameID) *FrameSnapshot
GetAllSnapshots() []*FrameSnapshot
Clear()
Count() int
```

### TimeTravelCursor

```go
NewTimeTravelCursor(mgr *SnapshotManager) *TimeTravelCursor
MoveToFrame(frameID FrameID) bool
MoveNext() bool
MovePrev() bool
MoveBySteps(steps int) bool
JumpToEvent(eventID EventID) bool
Bookmark(name string) bool
GotoBookmark(name string) bool
GetDiffToNext() *SnapshotDiff
GetInfo() *CursorInfo
```

### ReplayEngine

```go
NewReplayEngine(mgr *SnapshotManager, cursor *TimeTravelCursor) *ReplayEngine
SetReplaySpeed(speed float64)
ReplayFrom(frameID FrameID) error
ReplayRange(startID, endID FrameID) error
Stop()
StepForward() bool
StepBackward() bool
```

### DiffEngine

```go
NewDiffEngine() *DiffEngine
Compute(from, to *FrameSnapshot) *SnapshotDiff
ComputeDeltaSet(diff *SnapshotDiff) *DeltaSet
ClearCache()
```

### TimeTravelClient

```go
NewTimeTravelClient(mgr *SnapshotManager) *TimeTravelClient
Render(width, height int) string
HandleInput(input string) bool
SetView(view ViewMode)
ToggleLayout()
ToggleState()
ExportState() map[string]interface{}
```

---

## 设计特点

1. **完整状态捕获**: 组件、布局、重绘、因果图全部保存
2. **高效差异计算**: 增量模型，只存储变化
3. **灵活导航**: 前后移动、跳转、书签
4. **可控回放**: 速度控制、范围回放、单步执行
5. **TUI 界面**: 时间线、快照、差异、回放四种视图
6. **JSON 序列化**: 支持快照导出/导入

---

## 下一步 (阶段4: 确定性回放)

- [ ] EventRecorder 事件录制
- [ ] EventReplayer 事件回放
- [ ] DeterminismChecker 确定性验证
- [ ] RandomSeedCapture 随机种子捕获
- [ ] InputCapture 输入捕获

---

## 验收检查清单

### 编译与测试
- [x] timetravel 包编译通过
- [x] 整个项目编译通过
- [x] 16/16 单元测试通过
- [x] 无循环依赖

### 功能实现
- [x] FrameSnapshot 完整状态快照已实现
- [x] TimeTravelCursor 时间游标已实现
- [x] StateReplay 状态回放已实现
- [x] DiffEngine 快照差异引擎已实现
- [x] TimeTravelClient TUI 客户端已实现

### 特性
- [x] 支持快照链式遍历
- [x] 支持书签管理
- [x] 支持差异可视化
- [x] 支持回放速度控制
- [x] 支持状态导出/导入

---

## 总结

阶段3 完成了时间旅行系统的核心实现，提供了:

1. **完整快照**: 捕获每帧的完整状态
2. **自由导航**: 在帧之间自由移动和跳转
3. **智能回放**: 支持范围回放和速度控制
4. **精确差异**: 高效的差异计算和可视化
5. **友好界面**: TUI 客户端提供直观的操作体验

这些功能为调试提供了强大的时间旅行能力，开发者可以：
- 回退到任意帧查看状态
- 重放问题发生过程
- 对比不同帧之间的差异
- 通过书签快速定位关键帧
