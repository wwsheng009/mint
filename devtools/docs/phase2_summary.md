# DevTools 阶段2 实施总结

> 实施日期: 2026-01-30
> 状态: 已完成
> 阶段: 因果链引擎 (P1)
> 编译: ✅ 通过
> 测试: ✅ 16/16 通过

## 已完成的工作

### 1. Causal Graph 数据结构

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/causal.go` | `CausalGraph` 核心数据结构 | ✅ |
| `devtools/causal.go` | `CausalEvent` 事件节点 | ✅ |
| `devtools/causal.go` | `CausalMutation` 变更节点 | ✅ |
| `devtools/causal.go` | `CausalLayout` 布局变更节点 | ✅ |
| `devtools/causal.go` | `CausalRepaint` 重绘节点 | ✅ |
| `devtools/causal.go` | `CausalEdge` 因果边 | ✅ |
| `devtools/causal.go` | `FrameSummary` 帧摘要 | ✅ |
| `devtools/types.go` | `LayoutID`, `RepaintID` 类型 | ✅ |

### 2. CausalBuilder 因果链构建器

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/causal_builder.go` | `CausalBuilder` 主构建器 | ✅ |
| `devtools/causal_builder.go` | `RecordEvent()` 记录事件 | ✅ |
| `devtools/causal_builder.go` | `RecordMutation()` 记录变更 | ✅ |
| `devtools/causal_builder.go` | `RecordLayoutChange()` 记录布局 | ✅ |
| `devtools/causal_builder.go` | `RecordRepaint()` 记录重绘 | ✅ |
| `devtools/causal_builder.go` | 图历史管理 (100帧) | ✅ |
| `devtools/causal_builder.go` | `GetStats()` 统计信息 | ✅ |

### 3. FrameTimeline 模型

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/timeline.go` | `FrameTimeline` 时间线 | ✅ |
| `devtools/timeline.go` | `FrameEntry` 帧条目 | ✅ |
| `devtools/timeline.go` | `BeginFrame()` / `EndFrame()` | ✅ |
| `devtools/timeline.go` | `AttachGraph()` 关联因果图 | ✅ |
| `devtools/timeline.go` | `GetSlowFrames()` 慢帧分析 | ✅ |
| `devtools/timeline.go` | `GetStats()` 统计信息 | ✅ |
| `devtools/timeline.go` | FPS 计算 | ✅ |

### 4. Causal Query API

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/causal_query.go` | `CausalQuery` 查询器 | ✅ |
| `devtools/causal_query.go` | `FindRootCauses()` 根本原因分析 | ✅ |
| `devtools/causal_query.go` | `FindEffects()` 影响链分析 | ✅ |
| `devtools/causal_query.go` | `GetCausalPath()` 因果路径 | ✅ |
| `devtools/causal_query.go` | `GetCriticalPath()` 关键路径 | ✅ |
| `devtools/causal_query.go` | `TraceFromEvent()` 完整追踪 | ✅ |
| `devtools/causal_query.go` | `EffectChain` 效果链 | ✅ |

### 5. Component 状态变更 Hook

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/component_hook.go` | `ComponentHookManager` Hook 管理器 | ✅ |
| `devtools/component_hook.go` | `ComponentHooks` 组件 Hook | ✅ |
| `devtools/component_hook.go` | `SimpleHookManager` 简化 Hook | ✅ |
| `devtools/component_hook.go` | 状态变更自动记录 | ✅ |
| `devtools/component_hook.go` | 全局 Hook 实例 | ✅ |

---

## 架构设计

### 因果链数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                        因果链数据流                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Render Thread                           Debug Goroutine         │
│  ─────────────                           ────────────────       │
│                                                                  │
│  1. Event Phase                                                   │
│  ┌─────────────┐    RecordEvent()        ┌──────────────────┐   │
│  │ Event Input │ ─────────────────────►  │ CausalGraph      │   │
│  └─────────────┘                         │ • Add Event      │   │
│                                          │ • Set Context    │   │
│                                          └──────────────────┘   │
│                                                  │               │
│  2. Mutation Phase                               ▼               │
│  ┌─────────────┐    RecordMutation()     ┌──────────────────┐   │
│  │ State Change│ ─────────────────────►  │ CausalGraph      │   │
│  └─────────────┘                         │ • Add Mutation   │   │
│                                          │ • Link to Event  │   │
│                                          └──────────────────┘   │
│                                                  │               │
│  3. Layout Phase                                 ▼               │
│  ┌─────────────┐    RecordLayoutChange()  ┌──────────────────┐   │
│  │ Layout Calc │ ─────────────────────►  │ CausalGraph      │   │
│  └─────────────┘                         │ • Add Layout     │   │
│                                          │ • Link to Mutations│  │
│                                          └──────────────────┘   │
│                                                  │               │
│  4. Paint Phase                                  ▼               │
│  ┌─────────────┐    RecordRepaint()       ┌──────────────────┐   │
│  │ Paint       │ ─────────────────────►  │ CausalGraph      │   │
│  └─────────────┘                         │ • Add Repaint    │   │
│                                          │ • Link to Layout │   │
│                                          └──────────────────┘   │
│                                                                  │
│  5. Analysis Phase                                               │
│                                          ┌──────────────────┐   │
│                                          │ CausalQuery      │   │
│                                          │ • FindRootCauses │   │
│                                          │ • FindEffects    │   │
│                                          │ • GetCausalPath  │   │
│                                          └──────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### CausalGraph 结构

```
┌─────────────────────────────────────────────────────────────────┐
│                     CausalGraph 结构                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    CausalGraph                          │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │  FrameID     FrameID                                    │    │
│  │  StartTime   time.Time                                  │    │
│  │  EndTime     time.Time                                  │    │
│  │                                                          │    │
│  │  Nodes:                                                  │    │
│  │    ├── Events      []*CausalEvent    (输入)             │    │
│  │    ├── Mutations   []*CausalMutation (状态变更)          │    │
│  │    ├── Layouts     []*CausalLayout   (布局变更)          │    │
│  │    └── Repaints    []*CausalRepaint  (重绘)             │    │
│  │                                                          │    │
│  │  Edges:                                                  │    │
│  │    └── Edges      []*CausalEdge     (因果链)            │    │
│  │                                                          │    │
│  │  Indexes (O(1) 查找):                                    │    │
│  │    ├── eventIndex    map[EventID]int                    │    │
│  │    ├── mutationIndex map[MutationID]int                 │    │
│  │    ├── layoutIndex   map[NodeID]int                     │    │
│  │    └── repaintIndex  map[RepaintID]int                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  因果链类型:                                                      │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  EdgeEventToMutation    Event → Mutation (事件导致变更)   │     │
│  │  EdgeMutationToLayout   Mutation → Layout (变更导致布局)  │     │
│  │  EdgeLayoutToRepaint    Layout → Repaint (布局导致重绘)   │     │
│  └────────────────────────────────────────────────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### CausalQuery API

```
┌─────────────────────────────────────────────────────────────────┐
│                    CausalQuery API                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  根本原因分析:                                                   │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  FindRootCauses(repaintID) → []*CausalEvent             │    │
│  │                                                          │    │
│  │  "导致这次重绘的所有原始事件是什么？"                      │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  影响链分析:                                                     │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  FindEffects(eventID) → EffectChain                     │    │
│  │                                                          │    │
│  │  "这个事件导致了哪些变更、布局和重绘？"                     │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  因果路径分析:                                                   │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  GetCausalPath(eventID, repaintID) → CausalPath         │    │
│  │                                                          │    │
│  │  "从事件到重绘的完整因果路径是什么？"                      │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  关键路径分析:                                                   │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  GetCriticalPath() → CausalPath                         │    │
│  │                                                          │    │
│  │  "帧内最长的因果链是什么？（性能瓶颈）"                    │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  完整追踪:                                                       │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  TraceFromEvent(eventID) → TraceResult                  │    │
│  │                                                          │    │
│  │  "这个事件的完整影响，包括受影响节点和脏区域"               │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 文件结构

```
mint/
└── devtools/
    ├── types.go              # 核心类型 (+ LayoutID, RepaintID)
    ├── bus.go                # 异步事件总线
    ├── tap.go                # Mutation Tap
    ├── collector.go          # Delta 收集器
    ├── async_collector.go    # 异步收集器协调
    ├── devtools.go           # 主入口
    ├── devtools_test.go      # 单元测试
    │
    ├── causal.go             # ✨ 阶段2: Causal Graph
    ├── causal_builder.go     # ✨ 阶段2: CausalBuilder
    ├── timeline.go           # ✨ 阶段2: FrameTimeline
    ├── causal_query.go       # ✨ 阶段2: Query API
    ├── component_hook.go     # ✨ 阶段2: Component Hooks
    │
    ├── timetravel/           # 阶段3: 时间旅行 (计划中)
    ├── replay/               # 阶段4: 确定性回放 (计划中)
    └── client/               # 阶段5: 客户端 (计划中)
```

---

## 使用示例

### 1. 基本因果追踪

```go
import "github.com/wwsheng009/mint/devtools"

// 初始化
builder := devtools.NewCausalBuilder()
builder.Enable()

// 在帧循环中
func FrameLoop() {
    frameID := devtools.FrameID(frameCounter)

    // 开始帧
    builder.BeginFrame(frameID)

    // 处理事件
    for _, event := range inputEvents {
        builder.RecordEvent(event.Type, event.TargetID, event.Phase)

        // 事件处理会触发状态变更
        // 组件内部通过 Hook 自动记录变更
    }

    // 布局计算
    for _, node := range layoutNodes {
        if node.Invalidated {
            oldRect := node.OldRect()
            newRect := node.CalculateLayout()

            builder.RecordLayoutChange(
                node.ID,
                node.ChangeMask,
                &oldRect,
                &newRect,
            )
        }
    }

    // 渲染
    dirtyRegions := renderer.GetDirtyRegions()
    builder.RecordRepaint(
        dirtyRegions,
        renderer.ChangedCells(),
        renderer.TotalCells(),
    )

    // 结束帧
    builder.EndFrame()
}
```

### 2. 根本原因分析

```go
// 获取当前帧的因果图
graph := builder.GetCurrentGraph()
if graph != nil {
    // 创建查询器
    query := devtools.NewCausalQuery(graph)

    // 找到导致重绘的根本原因
    repaints := graph.Repaints
    if len(repaints) > 0 {
        rootCauses := query.FindRootCauses(repaints[0].ID)

        fmt.Printf("重绘由以下事件触发:\n")
        for _, event := range rootCauses {
            fmt.Printf("  - %s on %s (%s)\n",
                event.Type, event.TargetID, event.Phase)
        }
    }
}
```

### 3. 影响链分析

```go
// 分析某个事件的影响
eventID := lastEventID
query := devtools.NewCausalQuery(graph)

effects := query.FindEffects(eventID)

fmt.Printf("事件 %s 的影响:\n", eventID)
fmt.Printf("  变更: %d\n", len(effects.Mutations))
fmt.Printf("  布局: %d\n", len(effects.Layouts))
fmt.Printf("  重绘: %d\n", len(effects.Repaints))
```

### 4. Component Hook

```go
// 创建 Hook 管理器
hookMgr := devtools.NewComponentHookManager(builder)
hookMgr.Enable()

// 为组件注册 Hook
hookMgr.RegisterHooks(componentID, "MyComponent", &devtools.ComponentHooks{
    BeforeStateChange: func(field string, oldValue, newValue interface{}) bool {
        fmt.Printf("状态变更: %s.%s: %v → %v\n", "MyComponent", field, oldValue, newValue)
        return true // 允许变更
    },
    AfterStateChange: func(field string, oldValue, newValue interface{}) {
        // 自动记录到因果图 (由 HookManager 处理)
    },
})

// 简化的 Hook 接口
simpleHookMgr := devtools.GetGlobalSimpleHookManager(builder)
simpleHookMgr.Enable()

simpleHookMgr.OnStateChange(func(compID uint32, field string, oldValue, newValue interface{}) {
    fmt.Printf("State changed: %s: %v → %v\n", field, oldValue, newValue)
})

// 在状态变更处调用
simpleHookMgr.RecordStateChange("MyComponent", "count", oldValue, newValue)
```

---

## 设计特点

1. **完整的因果链**: Event → Mutation → Layout → Repaint
2. **双向查询**: 支持根本原因分析和影响链分析
3. **路径追踪**: 支持查找任意两点间的因果路径
4. **性能分析**: 关键路径分析找出性能瓶颈
5. **自动化 Hook**: 组件状态变更自动记录到因果图
6. **历史管理**: 保存最近 100 帧的因果图

---

## 性能特性

| 特性 | 实现 |
|------|------|
| O(1) 节点查找 | 索引 map |
| O(E) 路径查找 | BFS 算法 |
| 无锁写入 | sync.RWMutex 细粒度锁 |
| 内存控制 | 固定 100 帧历史 |
| 异步处理 | Debug Goroutine 分析 |

---

## 下一步 (阶段3: 时间旅行)

- [ ] FrameSnapshot 完整状态快照
- [ ] TimeTravelCursor 时间游标
- [ ] StateReplay 状态回放
- [ ] DiffEngine 快照差异引擎
- [ ] TimeTravelClient 时间旅行 UI

---

## 验收检查清单

### 编译与测试
- [x] devtools 包编译通过
- [x] 16/16 单元测试通过
- [x] 无循环依赖

### 功能实现
- [x] CausalGraph 数据结构已实现
- [x] CausalBuilder 因果链构建器已实现
- [x] FrameTimeline 模型已实现
- [x] Causal Query API 已实现
- [x] Component 状态变更 Hook 已实现

### 集成
- [x] 与阶段1模块无缝集成
- [x] 支持与 runtime LayoutVersion 集成
- [x] Hook 自动记录状态变更

---

## 总结

阶段2 完成了因果链引擎的核心实现，提供了:

1. **完整的因果追踪**: 从事件到重绘的完整因果链
2. **强大的查询能力**: 支持根本原因、影响链、路径追踪等分析
3. **自动化 Hook**: 组件状态变更自动集成到因果图
4. **性能分析工具**: 关键路径、慢帧分析等

这些功能为后续的时间旅行和确定性回放奠定了基础。
