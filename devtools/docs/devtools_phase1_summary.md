# DevTools 阶段1 实施总结

> 实施日期: 2026-01-30
> 状态: 已完成
> 阶段: 增量基础 (P0)
> 编译: ✅ 通过
> 测试: ✅ 16/16 通过

## 已完成的工作

### 1. Runtime 侧修改

| 文件 | 修改内容 | 状态 |
|------|----------|------|
| `runtime/node.go` | 添加 `LayoutVersion uint32` 字段 | ✅ |
| `runtime/node.go` | 添加 `InvalidateLayout()` 方法 | ✅ |
| `runtime/node.go` | 添加 `SetPosition()` 方法 | ✅ |
| `runtime/node.go` | 添加 `SetSize()` 方法 | ✅ |
| `runtime/node.go` | 添加 `GetLayoutVersion()` 方法 | ✅ |
| `runtime/node.go` | 修改 `MarkLayoutDirty()` 自动递增版本 | ✅ |
| `runtime/debug_id.go` | 创建 Debug ID 注册表系统 | ✅ |

### 2. DevTools 核心模块

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/types.go` | 核心类型定义 (NodeID, MutationID, FrameID, etc.) | ✅ |
| `devtools/types.go` | Delta 类型 (LayoutDelta, RepaintDelta, EventDelta) | ✅ |
| `devtools/types.go` | Debug 消息类型 (DebugMessage, MessageType) | ✅ |
| `devtools/types.go` | 配置类型 (Config, DefaultConfig) | ✅ |

### 3. 异步事件总线

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/bus.go` | Lock-Free Ring Buffer 实现 | ✅ |
| `devtools/bus.go` | `Emit()` 无锁写入 | ✅ |
| `devtools/bus.go` | `Subscribe()` 订阅机制 | ✅ |
| `devtools/bus.go` | `dispatchLoop()` 异步分发 | ✅ |

### 4. Mutation Tap

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/tap.go` | 全局 mutationTap Ring Buffer (16K) | ✅ |
| `devtools/tap.go` | `RecordMutation()` 极速写入 | ✅ |
| `devtools/tap.go` | `PollMutations()` 消费接口 | ✅ |
| `devtools/tap.go` | `EnableMutationTap()` / `DisableMutationTap()` | ✅ |

### 5. Delta 收集器

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/collector.go` | `LayoutCollector` 布局增量收集 | ✅ |
| `devtools/collector.go` | `EventCollector` 事件增量收集 | ✅ |
| `devtools/collector.go` | 背压处理 (chan 阻塞时丢弃) | ✅ |

### 6. 异步收集器协调

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/async_collector.go` | `AsyncCollector` 主协调器 | ✅ |
| `devtools/async_collector.go` | `Start()` 启动处理 goroutines | ✅ |
| `devtools/async_collector.go` | `Stop()` 优雅关闭 | ✅ |
| `devtools/async_collector.go` | `BeginFrame()` / `EndFrame()` 帧管理 | ✅ |

### 7. DevTools 主入口

| 文件 | 功能 | 状态 |
|------|------|------|
| `devtools/devtools.go` | `DevTools` 主结构 | ✅ |
| `devtools/devtools.go` | `Enable()` / `Disable()` | ✅ |
| `devtools/devtools.go` | `CollectLayout()` 收集布局 | ✅ |
| `devtools/devtools.go` | `CollectRepaint()` 收集重绘 | ✅ |
| `devtools/devtools.go` | `RecordEvent()` 记录事件 | ✅ |
| `devtools/devtools.go` | `Highlight()` 高亮组件 | ✅ |

### 8. 测试

| 文件 | 覆盖范围 | 状态 |
|------|----------|------|
| `devtools/devtools_test.go` | 16 个单元测试 | ✅ 全部通过 |

---

## 架构设计

### 依赖关系

```
┌─────────────────────────────────────────────────────────────────┐
│                        依赖关系图                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│    ┌─────────┐                                                  │
│    │  App    │                                                  │
│    │  Layer  │                                                  │
│    └────┬────┘                                                  │
│         │                                                       │
│         ├──> runtime     (核心运行时，零依赖)                    │
│         │    └── node.go       (LayoutVersion支持)               │
│         │    └── debug_id.go   (ID注册表)                       │
│         │                                                       │
│         └──> devtools    (调试工具，观察runtime)                  │
│              ├── types.go        (核心类型)                      │
│              ├── bus.go          (事件总线)                      │
│              ├── tap.go          (变更捕获)                      │
│              ├── collector.go    (增量收集)                      │
│              ├── async_*.go      (异步协调)                      │
│              └── devtools.go     (主入口)                        │
│                                                                  │
│    runtime  ───────>  devtools  (单向依赖，无循环)                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**关键原则**:

| 原则 | 说明 |
|------|------|
| 单向依赖 | `devtools` → `runtime`，runtime 不依赖 devtools |
| 观察者模式 | devtools 观察 runtime 的状态变化 |
| 依赖注入 | App 层负责协调，不破坏 runtime 的独立性 |
| 零侵入 | runtime 提供接口，devtools 消费数据 |

### 模块架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    DevTools 模块架构                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                    Runtime Layer                         │    │
│  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐       │    │
│  │  │ Layout │  │ Paint  │  │ Event  │  │ Focus  │       │    │
│  │  │ Engine │  │ Buffer │  │ System │  │Manager │       │    │
│  │  └────┬───┘  └────┬───┘  └────┬───┘  └────┬───┘       │    │
│  │       │            │            │            │           │    │
│  │       └────────────┴────────────┴────────────┘           │    │
│  │                      │                                   │    │
│  │           ┌──────────▼───────────┐                      │    │
│  │           │   LayoutNode (+ID)    │                      │    │
│  │           │   LayoutVersion       │                      │    │
│  │           │   GetLayoutVersion()  │                      │    │
│  │           └───────────────────────┘                      │    │
│  └────────────────────────────────────────────────────────┘    │
│                           │                                    │
│                           ▼                                    │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                   DevTools Layer                        │    │
│  │  ┌─────────────────────────────────────────────────┐  │    │
│  │  │              AsyncCollector                      │  │    │
│  │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐       │  │    │
│  │  │  │  Layout  │  │  Event   │  │  Frame   │       │  │    │
│  │  │  │ Collector│  │ Collector│  │Timeline  │       │  │    │
│  │  │  └─────┬────┘  └─────┬────┘  └─────┬────┘       │  │    │
│  │  │        │             │             │              │  │    │
│  │  └────────┼─────────────┼─────────────┼──────────────┘  │    │
│  │           │             │             │                   │    │
│  │  ┌────────▼─────────┐   │        ┌────▼────────┐       │    │
│  │  │    EventBus      │   │        │ MutationTap │       │    │
│  │  │ (Lock-Free Ring) │   │        │ (16K Buffer) │       │    │
│  │  └──────────────────┘   │        └─────────────┘       │    │
│  │                         │                                │    │
│  │           ┌─────────────▼───────────────┐             │    │
│  │           │    DebugMessage Channel     │             │    │
│  │           └─────────────┬───────────────┘             │    │
│  └──────────────────────────┼─────────────────────────────┘    │
│                             │                                   │
│                             ▼                                   │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                   Client Layer (阶段5)                  │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐     │    │
│  │  │ TUI      │  │ Web      │  │ Protocol         │     │    │
│  │  │ Panel    │  │ Dashboard│  │ (WebSocket/IPC)  │     │    │
│  │  └──────────┘  └──────────┘  └──────────────────┘     │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                        数据流向图                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Render Thread                           Debug Goroutine         │
│  ─────────────                           ────────────────       │
│                                                                  │
│  1. Layout Phase                                                  │
│  ┌─────────────┐                                                   │
│  │ LayoutEngine │                                                   │
│  └──────┬──────┘                                                   │
│         │                                                          │
│         │  LayoutResult                                          │
│         │  (with LayoutVersion)                                    │
│         │                                                          │
│         ▼                                                          │
│  ┌─────────────┐     Emit(LayoutEvent)                           │
│  │LayoutCollector│ ─────────────────────────────────────────►   │
│  └─────────────┘                                                   │
│                                                                  │
│  2. Paint Phase                                                   │
│  ┌─────────────┐                                                   │
│  │Paint Engine  │                                                   │
│  └──────┬──────┘                                                   │
│         │                                                          │
│         │  DirtyRegions                                           │
│         │                                                          │
│         ▼                                                          │
│  ┌─────────────┐     Emit(RepaintEvent)                          │
│  │RepaintCollector│ ─────────────────────────────────────────►  │
│  └─────────────┘                                                   │
│                                                                  │
│  3. Event Phase                                                   │
│  ┌─────────────┐                                                   │
│  │ Event Dispatch│                                                  │
│  └──────┬──────┘                                                   │
│         │                                                          │
│         │  InputEvent                                            │
│         │                                                          │
│         ▼                                                          │
│  ┌─────────────┐     Emit(InputEvent)                            │
│  │EventCollector │ ─────────────────────────────────────────►    │
│  └─────────────┘                                                   │
│                                                                  │
│                                                                 │
│                           ┌─────────────────────────────────┐   │
│                           │    Process in Debug Goroutine    │   │
│                           │                                   │   │
│                           │  • Build FrameTimeline         │   │
│                           │  • Assemble Deltas             │   │
│                           │  • Send to DebugMessage Channel│   │
│                           └───────────────┬─────────────────┘   │
│                                           │                       │
│                                           ▼                       │
│                           ┌─────────────────────────────────┐   │
│                           │      Client Layer               │   │
│                           │  • TUI Panel                     │   │
│                           │  • Web Dashboard                │   │
│                           └─────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 性能特性

```
┌─────────────────────────────────────────────────────────────────┐
│                      性能设计特性                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  零开销设计 (Debug 关闭时):                                        │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  if atomic.LoadUint32(&enabled) == 0 { return }         │     │
│  │                                                          │     │
│  │  • 单条指令                                            │     │
│  │  • 分支预测成功                                         │     │
│  │  • 无函数调用                                           │     │
│  │  → 开销 < 0.1%                                         │     │
│  └────────────────────────────────────────────────────────┘     │
│                                                                  │
│  低开销设计 (Debug 开启时):                                        │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  • 增量收集: 只记录变化节点                               │     │
│  │  • Lock-Free: 无锁写入                                  │     │
│  │  • 异步处理: 不阻塞主循环                               │     │
│  │  • 背压丢弃: 防止内存堆积                               │     │
│  │  → 开销 < 5%                                          │     │
│  └────────────────────────────────────────────────────────┘     │
│                                                                  │
│  内存特性:                                                       │
│  ┌────────────────────────────────────────────────────────┐     │
│  │  • Ring Buffer: 固定大小，无 GC 压力                    │     │
│  │  • Delta 模型: 只存变化，内存占用可控                   │     │
│  │  • 预分配 ID: 避免字符串分配                           │     │
│  │  → 内存占用 < 10MB / 1000 帧                            │     │
│  └────────────────────────────────────────────────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 文件结构

```
mint/
├── runtime/
│   ├── node.go          # 修改: 添加 LayoutVersion
│   └── debug_id.go      # 新建: ID 注册表
│
└── devtools/
    ├── types.go              # 核心类型
    ├── bus.go                # 异步事件总线
    ├── tap.go                # Lock-Free Ring Buffer Tap
    ├── collector.go          # Delta 收集器
    ├── async_collector.go    # 异步收集器协调
    ├── devtools.go           # 主入口
    ├── devtools_test.go      # 单元测试
    │
    ├── causal/               # 阶段2: 因果链引擎
    ├── timetravel/           # 阶段3: 时间旅行
    ├── replay/               # 阶段4: 确定性回放
    └── client/               # 阶段5: 客户端
```

---

## 设计特点

1. **增量模型**: 只记录变化，而非完整快照
2. **异步处理**: Render Thread 只记录，Debug Goroutine 分析
3. **零侵入 Hook**: 使用预分配 ID，避免字符串操作
4. **背压机制**: Channel 阻塞时丢弃数据，不影响主循环
5. **Lock-Free**: 使用原子操作和 Ring Buffer
6. **单向依赖**: 避免循环依赖，runtime 保持独立

---

## 使用示例

```go
// 初始化 DevTools
dt := devtools.New()
dt.Enable()

// 在主循环中
for {
    // 开始帧
    dt.BeginFrame()

    // 执行布局
    layoutResult := runtime.Layout(...)
    dt.CollectLayout(layoutResult)

    // 执行渲染
    diffResult := renderer.Paint(...)
    dt.CollectRepaint(...)

    // 记录事件
    dt.RecordEvent("click", "button_1", "bubble", nil)

    // 结束帧
    dt.EndFrame()
}

// 禁用 DevTools
dt.Disable()
```

---

## 下一步 (阶段2: 因果链引擎)

- [ ] 实现 Causal Graph 数据结构
- [ ] 添加 Component 状态变更 Hook
- [ ] 实现 CausalBuilder
- [ ] 实现 FrameTimeline 模型
- [ ] 实现 Causal Query API

---

## 验收检查清单

### 编译与测试
- [x] devtools 包编译通过
- [x] runtime 包编译通过
- [x] 整个项目编译通过
- [x] 16/16 单元测试通过
- [x] 无循环依赖

### 功能实现
- [x] Runtime 侧 LayoutVersion 字段已添加
- [x] Debug ID 注册表系统已实现
- [x] 核心类型定义已完成
- [x] 异步事件总线已实现
- [x] Mutation Tap 已实现
- [x] Layout Delta Collector 已实现
- [x] Event Delta Collector 已实现
- [x] 异步收集器协调器已实现
- [x] DevTools 主入口已实现

### 性能目标 (待验证)
- [ ] Debug 关闭时开销 < 0.1% (需要性能基准测试)
- [ ] Debug 开启时开销 < 5% (需要性能基准测试)
- [ ] 大型 UI (1000+ 组件) 可用 (需要稳定性测试)
- [ ] 内存占用 < 10MB / 1000 帧 (需要内存分析)

---

## 测试结果

```
=== RUN   TestNodeID                  ✅ PASS
=== RUN   TestMutationID              ✅ PASS
=== RUN   TestFrameID                 ✅ PASS
=== RUN   TestChangeMask              ✅ PASS
=== RUN   TestRect                    ✅ PASS
=== RUN   TestNodeDelta               ✅ PASS
=== RUN   TestLayoutDelta             ✅ PASS
=== RUN   TestEventEntry              ✅ PASS
=== RUN   TestConfig                  ✅ PASS
=== RUN   TestDevToolsEnableDisable    ✅ PASS
=== RUN   TestDevToolsRecordEvent      ✅ PASS
=== RUN   TestDebugOverlay            ✅ PASS
=== RUN   TestEventBus                ✅ PASS
=== RUN   TestMutationTap             ✅ PASS
=== RUN   TestAtomicOperations        ✅ PASS
=== RUN   TestNextPowerOfTwo          ✅ PASS

PASS  (16/16 tests, 1.989s)
```

---

## 注意事项

1. **依赖关系**: `devtools` → `runtime` 为单向依赖，无循环依赖问题

2. **集成方式**: 使用依赖注入模式，在 App 层协调 runtime 和 devtools

3. **后续集成**: 需要将 DevTools 集成到 Runtime 的主循环中，在适当的时机调用收集方法

4. **性能测试**: 需要运行性能基准测试，确保满足性能目标
