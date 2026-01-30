# Mint TUI DevTools 文档

> **Mint TUI Runtime** 调试与可观测性系统文档

## 文档目录

### 📋 实施规划

| 文档 | 说明 | 状态 |
|------|------|------|
| [implementation_todo.md](./implementation_todo.md) | 分阶段实施计划 TODO LIST | 进行中 |
| [implementation_v2.md](./implementation_v2.md) | V2.0 架构设计文档 | 已完成 |

### 📊 阶段总结

| 文档 | 说明 | 状态 |
|------|------|------|
| [phase1_summary.md](./phase1_summary.md) | 阶段1: 增量基础实施总结 | ✅ 已完成 |
| [phase2_summary.md](./phase2_summary.md) | 阶段2: 因果链引擎实施总结 | ✅ 已完成 |
| [phase3_summary.md](./phase3_summary.md) | 阶段3: 时间旅行实施总结 | ✅ 已完成 |
| [phase4_summary.md](./phase4_summary.md) | 阶段4: 确定性回放实施总结 | ✅ 已完成 |
| [phase5_summary.md](./phase5_summary.md) | 阶段5: 客户端实施总结 | ✅ 已完成 |
| [phase6_summary.md](./phase6_summary.md) | 阶段6: 快照系统实施总结 | ✅ 已完成 |
| [phase7_summary.md](./phase7_summary.md) | 阶段7: 内存优化实施总结 | ✅ 已完成 |
| [phase8_summary.md](./phase8_summary.md) | 阶段8: 远程调试实施总结 | ✅ 已完成 |

### 🧪 问题分析

| 文档 | 说明 | 状态 |
|------|------|------|
| [phase1_5_issues_analysis.md](./phase1_5_issues_analysis.md) | 阶段1-5潜在问题分析报告 | ✅ 已完成 |

### 📐 设计文档

| 文档 | 说明 | 状态 |
|------|------|------|
| [phase6_design.md](./phase6_design.md) | 阶段6: 高级功能详细设计方案 (初稿) |
| [phase6_design_v2.md](./phase6_design_v2.md) | 阶段6: 智能分析层设计方案 V2.0 (修订版) | ✅ 已完成 |

### 📝 原始设计文档

这些文档位于 `framework/docs/buffer/` 目录，是设计过程中的原始讨论和评审：

| 文档 | 说明 |
|------|------|
| [idea5_review.md](../../framework/docs/buffer/idea5_review.md) | 架构审查文档 |
| [idea5_devtools_implementation_v2.md](../../framework/docs/buffer/idea5_devtools_implementation_v2.md) | 实施文档 V2.0 |

### 📦 模块文档

每个模块都有独立的文档说明其功能和使用方法：

| 模块 | 说明 | 文档 |
|------|------|------|
| [client](../client/readme.md) | 调试客户端、协议处理、可视化 | TUI 面板、WebSocket 协议 |
| [memory](../memory/readme.md) | 内存优化 | 环形缓冲区、自适应采样、内存监控 |
| [observation](../observation/readme.md) | 观察层 | 数据收集、统计分析、模式检测 |
| [protocol](../protocol/readme.md) | 类型定义、常量 | 核心类型系统 |
| [remote](../remote/readme.md) | 远程调试 | WebSocket、HTTP API、Chromium 集成 |
| [replay](../replay/readme.md) | 确定性回放 | 事件录制、回放引擎 |
| [snapshot](../snapshot/readme.md) | 快照系统 | 状态捕获、差异比较 |
| [testing](../testing/readme.md) | 测试工具 | Mock、Fixture、断言 |
| [timetravel](../timetravel/readme.md) | 时间旅行 | 帧快照、状态导航 |

---

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Mint TUI Application                          │
│                    (用户界面 + 业务逻辑)                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │ 观察者模式
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        DevTools 调试系统                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                   Core (核心层)                                  │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │    │
│  │  │ DevTools    │  │ EventBus    │  │ Collector   │             │    │
│  │  │             │  │             │  │             │             │    │
│  │  │ - BeginFrame │  │ - LockFree  │  │ - LayoutDelta│             │    │
│  │  │ - EndFrame   │  │ - AsyncChan  │  │ - EventDelta │             │    │
│  │  │ - RecordEvent│  │ - Broadcast │  │ - Repaint    │             │    │
│  │  └─────────────┘  └─────────────┘  │             │             │    │
│  └─────────────────────────────────│             │─────────────┘             │    │
│                                     ▼                                  │    │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                   Causal (因果层)                               │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │    │
│  │  │ CausalGraph │  │CausalBuild │  │CausalQuery │             │    │
│  │  │             │  │             │  │             │             │    │
│  │  │ - Nodes      │  │ - Link      │  │ - Path      │             │    │
│  │  │ - Edges      │  │ - Event→Mut │  │ - RootCause │             │    │
│  │  │ - Timestamp  │  │ - Mut→Lay  │  │ - Affected   │             │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                     ▼                                  │    │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                 Observation (观察层)                           │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │    │
│  │  │    Layer    │  │     V1      │  │     V2      │             │    │
│  │  │             │  │ Statistics  │  │ Patterns    │             │    │
│  │  │ - Record    │  │ - Counts    │  │ - Detect    │             │    │
│  │  │ - GetMetrics│  │ - TopN      │  │ - Insights  │             │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                     ▼                                  │    │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                  Analysis (分析层)                              │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │    │
│  │  │ Timeline    │  │  Snapshot   │  │  Memory     │             │    │
│  │  │             │  │             │  │             │             │    │
│  │  │ - FrameTime  │  │ - Capture   │  │ - RingBuffer │             │    │
│  │  │ - Duration   │  │ - Compare   │  │ - Sampling  │             │    │
│  │  │ - Serialize  │  │ - Persist   │  │ - Monitor   │             │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                     ▼                                  │    │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                  Replay (回放层)                                │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │    │
│  │  │ Recorder    │  │ Replayer    │  │ Determinism  │             │    │
│  │  │             │  │             │  │             │             │    │
│  │  │ - Events    │  │ - Playback   │  │ - Verify    │             │    │
│  │  │ - Inputs    │  │ - Speed     │  │ - Diff       │             │    │
│  │  │ - Seeds     │  │ - Pause     │  │             │             │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                     ▼                                  │    │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                  Client (客户端层)                              │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │    │
│  │  │ TUI Panel   │  │   Protocol  │  │  Visualizer  │             │    │
│  │  │             │  │             │  │             │             │    │
│  │  │ - Render    │  │ - WebSocket│  │ - GraphViz    │             │    │
│  │  │ - Input     │  │ - HTTP      │  │ - Charts    │             │    │
│  │  │ - Navigate  │  │ - Message   │  │             │             │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘             │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                     ▼                                  │    │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                  Remote (远程层)                                │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │    │
│  │  │   Protocol  │  │   HTTP      │  │   WebSocket│             │    │
│  │  │             │  │             │  │             │             │    │
│  │  │ - Messages  │  │ - /debug    │  │ - /ws       │             │    │
│  │  │ - Session   │  │ - /api/*    │  │ - Binary    │             │    │
│  │  │ - Subscribe │  │ - Export    │  │ - Broadcast │             │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘             │    │
│  │  ┌─────────────┐                                                    │    │
│  │  │ChromiumBridge│                                                   │    │
│  │  │ - CDP       │                                                    │    │
│  │  │ - Inspector │                                                    │    │
│  │  └─────────────┘                                                    │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 数据流图

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Runtime    │───▶│   DevTools   │───▶│   Observer   │
│              │    │              │    │              │
│  - Layout     │    │  - Events    │    │  - Stats     │
│  - Events    │    │  - Deltas    │    │  - Patterns  │
│  - Repaint   │    │              │    │  - Insights  │
└──────────────┘    └──────────────┘    └──────────────┘
                            │                    │
                            ▼                    ▼
                   ┌──────────────┐    ┌──────────────┐
                   │   Snapshot   │    │   Memory     │
                   │              │    │              │
                   │  - Capture   │    │  - RingBuf    │
                   │  - Compare    │    │  - Sampling   │
                   └──────────────┘    └──────────────┘
                            │                    │
                            ▼                    ▼
                   ┌──────────────────────────────────────┐
                   │            Client Output            │
                   │  ┌──────────────┐  ┌─────────────┐     │
                   │  │ TUI Panel    │  │ Web Dashboard│     │
                   │  │              │  │              │     │
                   │  │ - Timeline   │  │ - Charts     │     │
                   │  │ - Causal     │  │ - Graphs     │     │
                   │  │ - Stats      │  │ - Heatmap    │     │
                   │  └──────────────┘  └─────────────┘     │
                   └──────────────────────────────────────┘
                            │
                            ▼
                   ┌──────────────────────────────────────┐
                   │         Remote (WebSocket/HTTP)       │
                   │  ┌──────────────┐  ┌─────────────┐     │
                   │  │ Inspector UI │  │ REST API     │     │
                   │  │              │  │              │     │
                   │  │ - Browser    │  │ - JSON       │     │
                   │  └──────────────┘  └─────────────┘     │
                   └──────────────────────────────────────┘
```

### 模块依赖关系

```
                    ┌─────────────────────┐
                    │     devtools       │
                    │   (核心类型定义)     │
                    └────────▲────────────┘
                           │
        ┌───────────────────┼───────────────────┬─────────────────┐
        │                   │                   │             │
        ▼                   ▼                   ▼             ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  snapshot   │   │   memory     │   │ observation │   │   remote     │
│              │   │              │   │              │   │              │
│ ├─> Manager   │   ├─> RingBuffer │   ├─> Layer      │   ├─> Protocol   │
│ ├─> Builder   │   ├─> Sampling   │   ├─> V1 Stats   │   ├─> WebSocket │
│ └─> Differ    │   └─> Monitor    │   └─> V2 Detect  │   └─> HTTP       │
└──────────────┘                  │                   │   └──────────────┘
                                   │                   │
                    ┌──────────────┐   ┌──────────────┐
                    │  timetravel  │   │   replay     │
                    │              │   │              │
                    ├─> Cursor     │   ├─> Recorder  │
                    └─> Replay     │   └─> Replayer  │
                    └──────────────┘   └──────────────┘
                                   │
                    ┌──────────────┐   ┌──────────────┐
                    │   client     │   │   testing    │
                    │              │   │              │
                    ├─> Panel      │   ├─> Mock       │
                    ├─> Protocol   │   ├─> Fixture    │
                    └─> Visualizer │   └─> Assert     │
                    └──────────────┘   └──────────────┘
```

### 核心数据结构

```
┌─────────────────────────────────────────────────────────────┐
│                     DevTools 核心类型                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  NodeID        string    # 组件唯一标识 (预分配)              │
│  FrameID       int       # 帧编号                             │
│  MutationID    uint64    # 变更唯一标识                        │
│                                                             │
│  EventDelta                                      │
│  └── Type, Count, FirstTime, LastTime                   │
│                                                             │
│  LayoutDelta                                     │
│  └── Component, OldBounds, NewBounds, Reason          │
│                                                             │
│  RepaintDelta                                    │
│  └── DirtyRegions, ChangedCells, TotalCells           │
│                                                             │
│  CausalGraph                                    │
│  ├── Nodes: map[NodeID]*CausalNode                  │
│  └── Edges: []*CausalEdge                          │
│                                                             │
│  Snapshot                                              │
│  ├── ID, FrameID, Timestamp                         │
│  ├── States: map[NodeID]*ComponentState          │
│  └── Global: WindowSize, Cursor, FocusedNode        │
│                                                             │
│  Session (Remote Debugging)                        │
│  ├── id, clientID, createdAt                        │
│  └── subs: map[string]bool (订阅类型)              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 设计模式

| 模式 | 应用场景 | 实现位置 |
|------|---------|----------|
| **观察者模式** | DevTools 观察 Runtime | `devtools/*.go` |
| **构建器模式** | 快照构造 | `snapshot/builder.go` |
| **策略模式** | 采样策略 | `memory/sampling.go` |
| **工厂模式** | 创建组件 | `client/panel.go` |
| **单例模式** | 全局会话 | `remote/server.go` |
| **命令模式** | 处理消息 | `remote/protocol.go` |

### 性能考虑

| 策略 | 实现 | 目标 |
|------|------|------|
| **零拷贝** | 只记录变化的字段 | 最小化开销 |
| **异步处理** | goroutine 分析 | 不阻塞渲染 |
| **环形缓冲** | 固定内存上限 | O(1) 操作 |
| **自适应采样** | 根据内存压力调整 | 动态平衡 |
| **懒加载** | 按需计算 | 延迟执行 |

---

## 快速导航

### 🚀 快速开始

```go
import "github.com/wwsheng009/mint/devtools"

// 初始化
dt := devtools.New()
dt.Enable()

// 在主循环中
dt.BeginFrame()
dt.CollectLayout(layoutResult)
dt.CollectRepaint(dirtyRegions, changedCells, totalCells)
dt.EndFrame()

// 禁用
dt.Disable()
```

### 📸 快照系统

```go
import "github.com/wwsheng009/mint/devtools/snapshot"

// 创建快照管理器
mgr := snapshot.NewManager(1000)

// 捕获快照
builder := snapshot.NewBuilder("snap-1", devtools.FrameID(42))
builder.SetWindowSize(80, 24)
builder.AddComponent(&snapshot.ComponentState{
    NodeID: "button-1",
    Type: "Button",
    Props: map[string]interface{}{"label": "Click me"},
})
snap, _ := mgr.Capture(42, builder)

// 比较快照
differ := snapshot.NewDiffer()
diff := differ.Compare(snap1, snap2)
fmt.Printf("Changes: %d\n", len(diff.Changes))
```

### 💾 内存优化

```go
import "github.com/wwsheng009/mint/devtools/memory"

// 环形缓冲区 (最近 1000 帧)
ring := memory.NewRingBuffer(1000)
ring.Write(devtools.FrameID(1))
frames := ring.GetAll()

// 自适应采样 (10%-100%)
sampler := memory.NewAdaptiveStrategy(0.1, 1.0)
if sampler.ShouldSample(frameID, context) {
    captureFullMetrics()
}

// 内存监控
monitor := memory.NewMonitor()
monitor.SetThresholds(0.7, 0.9)  // 警告/严重阈值
monitor.SetAlertCallback(func(alert memory.MemoryAlert) {
    log.Printf("Alert: %s\n", alert.Message)
})
monitor.Start()
```

### 🌐 远程调试

```go
import "github.com/wwsheng009/mint/devtools/remote"

// 创建服务器
server := remote.NewDevToolsServer(9222, dt, snapshotMgr)
go server.Start()

// 服务器运行在:
// - http://localhost:9222/debug    (Inspector UI)
// - ws://localhost:9222/ws          (WebSocket)
// - http://localhost:9222/api/*    (REST API)
```

**WebSocket 客户端示例:**

```go
ws, _ := websocket.Dial("ws://localhost:9222/ws", "", "http://localhost")

// 发送请求
websocket.JSON.Send(ws, map[string]interface{}{
    "version": "1.0.0",
    "type": "get_range",
    "id": "req-1",
    "payload": map[string]int{"from": 0, "to": 100},
})

// 接收响应
var msg map[string]interface{}
websocket.JSON.Receive(ws, &msg)
```

### 📖 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                    DevTools 架构                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   runtime (零依赖)                                                │
│      │                                                           │
│      └──> devtools (观察者)                                     │
│              │                                                   │
│              ├──> EventBus (异步处理)                            │
│              ├──> MutationTap (变更捕获)                         │
│              ├──> Collectors (增量收集)                          │
│              │                                                   │
│              ├──> snapshot (快照系统)                            │
│              │    ├── Manager (生命周期管理)                    │
│              │    ├── Differ (差异比较)                         │
│              │    └── TimeTravelRange (时间旅行)                │
│              │                                                   │
│              ├──> memory (内存优化)                             │
│              │    ├── RingBuffer (环形缓冲区)                   │
│              │    ├── Sampling (自适应采样)                     │
│              │    └── Monitor (内存监控)                        │
│              │                                                   │
│              └──> remote (远程调试)                             │
│                   ├── Protocol (CDP 协议)                       │
│                   ├── WebSocket (实时通信)                      │
│                   ├── HTTP (REST API)                          │
│                   └── Inspector (Web UI)                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 📚 实施阶段

```
阶段 1: ███████████████████████████████████  增量基础 (P0) ✅
阶段 2: ███████████████████████████████████  因果链引擎 (P1) ✅
阶段 3: ███████████████████████████████████  时间旅行 (P2) ✅
阶段 4: ███████████████████████████████████  确定性回放 (P2) ✅
阶段 5: ███████████████████████████████████  客户端实施 (P3) ✅
阶段 6: ███████████████████████████████████  快照系统 (P4) ✅
阶段 7: ███████████████████████████████████  内存优化 (P4) ✅
阶段 8: ███████████████████████████████████  远程调试 (P5) ✅
```

---

## 架构原则

| 原则 | 说明 |
|------|------|
| **单向依赖** | `devtools` → `runtime`，runtime 不依赖 devtools |
| **观察者模式** | devtools 观察 runtime 的状态变化 |
| **增量模型** | 只记录变化，而非完整快照 |
| **异步处理** | Render Thread 记录，Debug Goroutine 分析 |
| **零侵入** | 使用预分配 ID，避免字符串操作 |

---

## 当前状态

### ✅ 已完成 (阶段 1-8)

**阶段 1: 增量基础 (P0)**
- [x] Runtime LayoutVersion 支持
- [x] Debug ID 注册表系统
- [x] 核心类型定义
- [x] Lock-Free 事件总线
- [x] Mutation Tap (变更捕获)
- [x] Layout/Event 增量收集器
- [x] 异步收集器协调

**阶段 2: 因果链引擎 (P1)**
- [x] CausalGraph 数据结构
- [x] CausalBuilder 因果链构建器
- [x] FrameTimeline 帧时间线
- [x] Causal Query API
- [x] Component 状态变更 Hook
- [x] 根本原因分析
- [x] 影响链分析
- [x] 因果路径追踪

**阶段 3: 时间旅行 (P2)**
- [x] FrameSnapshot 完整状态快照
- [x] TimeTravelCursor 时间游标
- [x] StateReplay 状态回放
- [x] DiffEngine 快照差异引擎
- [x] TimeTravelClient TUI 客户端
- [x] 书签管理
- [x] 回放速度控制
- [x] JSON 导出/导入

**阶段 4: 确定性回放 (P2)**
- [x] EventRecorder 事件录制
- [x] EventReplayer 事件回放
- [x] DeterminismChecker 确定性验证
- [x] RandomSeedCapture 随机种子捕获
- [x] InputCapture 输入捕获
- [x] 回放进度跟踪
- [x] 断点功能
- [x] 宏录制/回放

**阶段 5: 客户端实施 (P3)**
- [x] TUI 调试面板 (TuiDebugPanel)
- [x] WebSocket 协议 (WebSocketHandler)
- [x] Web Dashboard
- [x] 远程调试会话 (RemoteDebugSession)
- [x] API 处理器 (APIHandler)
- [x] 调试覆盖层 (DebugOverlay)
- [x] 性能分析器 (Profiler)

**阶段 6: 快照系统 (P4)**
- [x] Snapshot 数据结构
- [x] Builder 构建器模式
- [x] Manager 快照管理器
- [x] Differ 差异引擎
- [x] TimeTravelRange 时间旅行范围
- [x] 持久化支持 (Save/Load)
- [x] 单元测试

**阶段 7: 内存优化 (P4)**
- [x] RingBuffer 环形缓冲区
- [x] FrameWindow 滑动窗口
- [x] AdaptiveStrategy 自适应采样
- [x] FixedRateStrategy 固定速率采样
- [x] PriorityStrategy 优先级采样
- [x] Monitor 内存监控器
- [x] Alert 告警系统

**阶段 8: 远程调试 (P5)**
- [x] Protocol 消息协议定义
- [x] ChromiumBridge CDP 桥接器
- [x] WebSocketServer WebSocket 服务器
- [x] HTTPServer REST API 服务器
- [x] Inspector UI (内嵌 HTML)
- [x] 错误处理 (JSON 错误响应)
- [x] 测试客户端

**测试状态**
- [x] 编译通过
- [x] 单元测试通过

---

## 性能目标

| 指标 | 目标 | 状态 |
|------|------|------|
| Debug 关闭时开销 | < 0.1% | 待验证 |
| Debug 开启时开销 | < 5% | 待验证 |
| 内存占用 (1000帧) | < 10MB | 待验证 |
| 大型 UI 支持 | 1000+ 组件 | 待验证 |

---

## 相关链接

- [GoDoc](https://pkg.go.dev/github.com/wwsheng009/mint/devtools)
- [项目主 README](../../README.md)
- [Runtime 文档](../../runtime/)
