# TUI DevTools 实施计划 TODO LIST

> **项目**: Mint TUI Runtime
> **文档版本**: 1.1
> **创建日期**: 2026-01-30
> **更新日期**: 2026-01-30
> **状态**: 阶段1 已完成

> **文档位置**: `devtools/docs/`
>
> **实施基础**:
> - 设计文档: `implementation_v2.md`
> - 原始评审: `../../framework/docs/buffer/idea5_review.md`

---

## 📊 当前状态 (2026-01-30)

### 阶段 1: 增量基础 (P0) - ✅ 已完成

```
┌─────────────────────────────────────────────────────────────────┐
│                    阶段1 完成状态                                 │
├─────────────────────────────────────────────────────────────────┤
│  编译状态    │  ✅ 通过                                          │
│  测试状态    │  ✅ 16/16 通过                                    │
│  循环依赖    │  ✅ 无 (devtools → runtime 单向依赖)               │
│                                                                  │
│  已实现模块:                                                       │
│  ✅ Runtime LayoutVersion 支持                                    │
│  ✅ Debug ID 注册表系统                                          │
│  ✅ 核心类型定义 (types.go)                                     │
│  ✅ Lock-Free 事件总线 (bus.go)                                │
│  ✅ Mutation Tap (tap.go)                                      │
│  ✅ Layout/Event 增量收集器 (collector.go)                       │
│  ✅ 异步收集器协调 (async_collector.go)                         │
│  ✅ DevTools 主入口 (devtools.go)                              │
│  ✅ 单元测试 (devtools_test.go)                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 实施进度

```
阶段 1: ████████████████████████████████████████  增量基础 (P0) ✅ 100%
阶段 2: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  因果链 (P1)   0%
阶段 3: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  时间旅行 (P2) 0%
阶段 4: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  确定性回放 (P2) 0%
阶段 5: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  客户端 (P3)   0%
阶段 6: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  高级功能 (未来) 0%
```

---

## 目录

1. [系统架构分析总结](#一系统架构分析总结)
2. [实施阶段总览](#二实施阶段总览)
3. [阶段1: 增量基础 (P0)](#三阶段1-增量基础-p0)
4. [阶段2: 因果链引擎 (P1)](#四阶段2-因果链引擎-p1)
5. [阶段3: 时间旅行系统 (P2)](#五阶段3-时间旅行系统-p2)
6. [阶段4: 确定性回放 (P2)](#六阶段4-确定性回放-p2)
7. [阶段5: 客户端实现 (P3)](#七阶段5-客户端实现-p3)
8. [阶段6: 高级功能 (未来)](#八阶段6-高级功能-未来)

---

## 一、系统架构分析总结

### 1.1 现有基础设施 (已更新)

| 模块 | 位置 | 状态 | DevTools 相关性 |
|------|------|------|----------------|
| LayoutNode | `runtime/node.go` | ✅ 已更新 | ✅ 添加 `LayoutVersion` 字段 |
| DebugID | `runtime/debug_id.go` | ✅ 新建 | ✅ ID 注册表系统 |
| Buffer | `runtime/paint/buffer.go` | ✅ 完善 | 支持宽字符和 diff 模式 |
| Layout Engine | `runtime/layout/` | ✅ 完善 | 需暴露 `DebugView` 接口 |
| Event System | `runtime/event/` | ✅ 完善 | 需添加事件记录 Hook |
| RenderDebug | `runtime/debug.go` | ✅ 存在 | 快照式，需升级为增量 |
| DevTools | `devtools/` | ✅ 已实现 | 核心模块已完成 |

### 1.2 已实现的文件结构

```
mint/
├── runtime/
│   ├── node.go              # 修改: 添加 LayoutVersion 和辅助方法
│   └── debug_id.go          # 新建: Debug ID 注册表
│
└── devtools/
    ├── types.go              # 核心类型定义
    ├── bus.go                # Lock-Free 事件总线
    ├── tap.go                # Mutation Tap
    ├── collector.go          # 增量收集器
    ├── async_collector.go    # 异步协调器
    ├── devtools.go           # 主入口 API
    ├── devtools_test.go      # 单元测试
    │
    ├── docs/                 # 📚 文档目录
    │   ├── README.md          # 文档索引
    │   ├── implementation_todo.md  # 本文档
    │   ├── implementation_v2.md    # 架构设计 V2.0
    │   └── phase1_summary.md       # 阶段1 完成总结
    │
    ├── causal/               # 阶段2: 因果链引擎 (待实施)
    ├── timetravel/           # 阶段3: 时间旅行 (待实施)
    ├── replay/               # 阶段4: 确定性回放 (待实施)
    └── client/               # 阶段5: 客户端 (待实施)
```

### 1.3 架构风险解决状态

| 风险 | 级别 | 状态 | 解决方案 |
|------|------|------|----------|
| 每帧全量快照 | P0 | ✅ 已解决 | 增量 Delta 模型 |
| 同步观察者 | P0 | ✅ 已解决 | 异步事件总线 |
| LayoutCollector 复制布局 | P0 | ✅ 已解决 | 直接读取 LayoutVersion |
| DebugOverlay 污染渲染 | P1 | ✅ 已解决 | 独立 DebugOverlay |
| Event Trace 缺时间线 | P1 | 🚧 计划中 | FrameTimeline 模型 (阶段2) |
| 协议层缺流控 | P2 | ✅ 已解决 | Channel 背压处理 |

---

## 二、实施阶段总览

```
┌─────────────────────────────────────────────────────────────────┐
│                    TUI DevTools 实施路线图                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  阶段 1: ████████████████████████████████████████  增量基础 (P0) ✅│
│    ├─ Layout Delta Collector                                     │
│    ├─ Mutation Tap (Lock-Free Ring Buffer)                       │
│    ├─ 异步事件总线                                                │
│    └─ 独立 Overlay Buffer                                         │
│                                                                  │
│  阶段 2: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  因果链 (P1)         │
│    ├─ Event → Mutation 关联                                       │
│    ├─ Mutation → Layout 关联                                      │
│    ├─ Layout → Repaint 关联                                       │
│    └─ FrameTimeline 模型                                          │
│                                                                  │
│  阶段 3: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  时间旅行 (P2)       │
│    ├─ 稀疏快照系统                                                │
│    ├─ 重放引擎                                                    │
│    └─ 时间轴 UI                                                  │
│                                                                  │
│  阶段 4: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  确定性回放 (P2)     │
│    ├─ Input Recorder                                             │
│    ├─ Replay 模式                                                 │
│    └─ 回放比对                                                    │
│                                                                  │
│  阶段 5: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  客户端 (P3)         │
│    ├─ TUI Panel                                                   │
│    ├─ Web Dashboard                                              │
│    └─ 协议优化                                                    │
│                                                                  │
│  阶段 6: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  高级功能 (未来)     │
│    ├─ 性能分析 AI                                                 │
│    ├─ 自动优化建议                                                │
│    └─ 代码重写                                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 优先级说明

| 优先级 | 说明 | 预估工作量 |
|--------|------|------------|
| P0 | 必须优先，否则 DevTools 会严重影响 Runtime 性能 | 40-50 小时 |
| P1 | 高优先级，核心功能缺失 | 30-40 小时 |
| P2 | 中优先级，高级调试能力 | 40-50 小时 |
| P3 | 低优先级，用户体验优化 | 20-30 小时 |

---

## 三、阶段1: 增量基础 (P0)

> **目标**: 解决一级风险，建立零侵入的增量数据收集基础
> **验收标准**:
> - Debug 关闭时性能影响 < 0.1%
> - Debug 开启时性能影响 < 5%
> - 大型 UI (1000+ 组件) 可正常运行

### 1.1 Runtime 侧修改

- [ ] **1.1.1 添加 LayoutVersion 字段**
  - [ ] 在 `LayoutNode` 结构体添加 `LayoutVersion uint32`
  - [ ] 实现 `InvalidateLayout()` 方法自动递增版本
  - [ ] 修改 `SetPosition()` 自动递增版本
  - [ ] 修改 `MarkLayoutDirty()` 自动递增版本
  - [ ] 添加单元测试验证版本递增逻辑
  - **文件**: `runtime/node.go`
  - **预估**: 2 小时

- [ ] **1.1.2 添加 Debug ID 系统**
  - [ ] 创建全局 `debugIDRegistry` 管理 Component ID
  - [ ] 为 Component 添加 `debugID uint32` 字段
  - [ ] 创建 `fieldIDRegistry` 管理 State Field ID
  - [ ] 实现预分配 ID 避免字符串使用
  - [ ] 添加 ID 分配和回收测试
  - **文件**: `runtime/debug_id.go` (新建)
  - **预估**: 3 小时

- [ ] **1.1.3 暴露 Layout DebugView 接口**
  - [ ] 定义 `LayoutDebugView` 接口
  - [ ] 实现 `ForEachBox()` 方法
  - [ ] 实现 `GetBoxInfo()` 方法返回调试信息
  - [ ] 集成到 LayoutEngine
  - **文件**: `runtime/layout/debug_view.go` (新建)
  - **预估**: 2 小时

### 1.2 核心类型定义

- [ ] **1.2.1 创建 DevTypes 核心类型**
  - [ ] 定义 `NodeID` 类型
  - [ ] 定义 `MutationID` 类型
  - [ ] 定义 `ChangeMask` 枚举
  - [ ] 定义 `LayoutDelta` 结构
  - [ ] 定义 `NodeDelta` 结构
  - [ ] 定义 `RepaintDelta` 结构
  - [ ] 定义 `EventDelta` 结构
  - [ ] 添加类型序列化方法
  - **文件**: `devtools/types.go` (新建)
  - **预估**: 3 小时

### 1.3 异步事件总线

- [ ] **1.3.1 实现 EventBus**
  - [ ] 定义 `DebugEvent` 极轻量结构
  - [ ] 定义 `DebugEventType` 枚举
  - [ ] 实现 Lock-Free Ring Buffer
  - [ ] 实现 `Emit()` 无锁写入
  - [ ] 实现 `Subscribe()` 订阅机制
  - [ ] 实现 `dispatchLoop()` 异步分发
  - [ ] 实现 `Enable()/Disable()` 控制开关
  - [ ] 添加并发安全测试
  - **文件**: `devtools/bus.go` (新建)
  - **预估**: 5 小时

- [ ] **1.3.2 实现 Mutation Tap**
  - [ ] 定义 `MutationRecord` 结构（无字符串/map）
  - [ ] 创建全局 mutationTap Ring Buffer (16K)
  - [ ] 实现 `recordMutation()` 极速写入
  - [ ] 实现 `PollMutations()` 消费接口
  - [ ] 添加启用/禁用控制
  - [ ] 性能测试: 确保单次写入 < 50ns
  - **文件**: `devtools/tap.go` (新建)
  - **预估**: 4 小时

### 1.4 Delta 收集器

- [ ] **1.4.1 实现 Layout Delta Collector**
  - [ ] 创建 `LayoutCollector` 结构
  - [ ] 实现 `lastVersion` 追踪 map
  - [ ] 实现 `nodeRegistry` 缓存
  - [ ] 实现 `Collect()` 增量检测逻辑
  - [ ] 实现 `buildNodeDelta()` 变化提取
  - [ ] 添加变化检测: Added/Removed/Changed
  - [ ] 实现背压处理 (chan 阻塞时丢弃)
  - [ ] 添加单元测试覆盖各种变化场景
  - **文件**: `devtools/delta/layout.go` (新建)
  - **预估**: 6 小时

- [ ] **1.4.2 实现 Repaint Delta Collector**
  - [ ] 创建 `RepaintCollector` 结构
  - [ ] 集成现有 `DiffResult` 数据
  - [ ] 提取 DirtyRegions 信息
  - [ ] 统计 ChangedCells/TotalCells
  - [ ] 实现背压处理
  - **文件**: `devtools/delta/repaint.go` (新建)
  - **预估**: 2 小时

- [ ] **1.4.3 实现 Event Delta Collector**
  - [ ] 创建 `EventCollector` 结构
  - [ ] 定义 `EventEntry` 结构
  - [ ] 集成到 Event Dispatch 机制
  - [ ] 记录事件: Type/Target/Phase/Time
  - [ ] 实现帧级别的 Event 聚合
  - **文件**: `devtools/delta/event.go` (新建)
  - **预估**: 3 小时

### 1.5 独立 Overlay Buffer

- [ ] **1.5.1 实现 Overlay 系统**
  - [ ] 创建 `Overlay` 独立 Buffer
  - [ ] 实现 `Highlight()` 绘制调试边框
  - [ ] 实现 `Clear()` 清空覆盖层
  - [ ] 实现 `Compose()` 合成到主输出
  - [ ] 确保 Overlay 不影响 Diff 算法
  - [ ] 实现 `drawBoxBorder()` 绘制边框
  - **文件**: `devtools/overlay/buffer.go` (新建)
  - **预估**: 3 小时

### 1.6 异步收集器协调

- [ ] **1.6.1 实现 AsyncCollector**
  - [ ] 创建 `AsyncCollector` 主结构
  - [ ] 实现 `Start()` 启动处理 goroutines
  - [ ] 实现 `processLayoutDeltas()` 处理循环
  - [ ] 实现 `processRepaintDeltas()` 处理循环
  - [ ] 实现 `processEventDeltas()` 处理循环
  - [ ] 实现 `processFrameTimeline()` 帧时间线
  - [ ] 添加优雅关闭机制
  - **文件**: `devtools/async_collector.go` (新建)
  - **预估**: 4 小时

### 1.7 Runtime 集成

- [ ] **1.7.1 创建 DevTools 主入口**
  - [ ] 实现 `DevTools` 主结构
  - [ ] 实现 `Initialize()` 初始化
  - [ ] 实现 `Enable()/Disable()` 开关
  - [ ] 实现 `IsEnabled()` 快速检查
  - [ ] 集成所有收集器
  - [ ] 集成 EventBus
  - [ ] 添加配置选项
  - **文件**: `devtools/devtools.go` (新建)
  - **预估**: 3 小时

- [ ] **1.7.2 集成到 Runtime 主循环**
  - [ ] 在 `Runtime` 添加 `devTools *DevTools` 字段
  - [ ] 在 Layout 后调用 LayoutCollector
  - [ ] 在 Paint 后调用 RepaintCollector
  - [ ] 在 Event Dispatch 后调用 EventCollector
  - [ ] 确保 Debug 关闭时零开销 (分支预测)
  - **文件**: `runtime/runtime.go`
  - **预估**: 2 小时

### 1.8 性能验证

- [ ] **1.8.1 性能基准测试**
  - [ ] 创建 benchmark 测试套件
  - [ ] 测试: Debug 关闭时开销 < 0.1%
  - [ ] 测试: Debug 开启时开销 < 5%
  - [ ] 测试: 大型 UI (1000+ 组件) 稳定性
  - [ ] 测试: GC 压力测试
  - [ ] 测试: 内存泄漏检测 (长时间运行)
  - **文件**: `devtools/benchmark_test.go` (新建)
  - **预估**: 4 小时

---

## 四、阶段2: 因果链引擎 (P1)

> **目标**: 从"事件日志"升级为"事件因果链"，回答"为什么会发生"
> **验收标准**:
> - 可追踪: Event → Mutation → Layout → Repaint 完整链路
> - 可回答: "为什么这个按钮闪了一下？"

### 2.1 因果链数据结构

- [ ] **2.1.1 创建 Causal Graph 类型**
  - [ ] 定义 `FrameRecord` 结构
  - [ ] 定义 `EventNode` 结构
  - [ ] 定义 `MutationNode` 结构
  - [ ] 定义 `MutationKind` 枚举
  - [ ] 定义 `Edge` 结构 (因果边)
  - [ ] 定义 `EdgeType` 枚举
  - [ ] 添加序列化支持
  - **文件**: `devtools/causal/graph.go` (新建)
  - **预估**: 3 小时

### 2.2 Mutation 捕获集成

- [ ] **2.2.1 Component 状态变更 Hook**
  - [ ] 在 `Component.SetState()` 添加 mutation 记录
  - [ ] 在 `Component.SetProp()` 添加 mutation 记录
  - [ ] 在 `Component.SetStyle()` 添加 mutation 记录
  - [ ] 实现 `recordMutation()` 调用
  - [ ] 确保 Record 使用预分配 ID
  - [ ] 添加单元测试
  - **文件**: `runtime/component.go`, `framework/component/base.go`
  - **预估**: 4 小时

- [ ] **2.2.2 Focus 变更 Hook**
  - [ ] 在 FocusManager 中添加 mutation 记录
  - [ ] 记录焦点变化事件
  - [ ] 关联到触发的 Event
  - **文件**: `runtime/focus/manager.go`
  - **预估**: 2 小时

- [ ] **2.2.3 Layout 变更 Hook**
  - [ ] 在 LayoutEngine 计算完成后记录变化
  - [ ] 记录哪些 Node 的 LayoutVersion 变化
  - [ ] 关联到触发的 Mutation
  - **文件**: `runtime/layout/engine.go`
  - **预估**: 2 小时

### 2.3 因果链构建

- [ ] **2.3.1 实现 CausalBuilder**
  - [ ] 创建 `CausalBuilder` 结构
  - [ ] 实现 `BeginFrame()` 开始新帧
  - [ ] 实现 `AddEvent()` 添加事件
  - [ ] 实现 `AddMutation()` 添加状态变更
  - [ ] 实现 `Build()` 构建完整因果链
  - [ ] 实现 `linkMutationsToLayout()` 关联
  - [ ] 实现 `linkLayoutToRepaint()` 关联
  - **文件**: `devtools/causal/builder.go` (新建)
  - **预估**: 5 小时

### 2.4 FrameTimeline 模型

- [ ] **2.4.1 实现 FrameTimeline**
  - [ ] 定义 `FrameTimeline` 结构
  - [ ] 包含: FrameID, StartTime, Events
  - [ ] 包含: Mutations, LayoutDelta, RepaintDelta
  - [ ] 实现 `AddEvent()` 记录
  - [ ] 实现 `Finalize()` 完成帧记录
  - [ ] 实现 `GetCausalChain()` 获取因果链
  - **文件**: `devtools/causal/timeline.go` (新建)
  - **预估**: 3 小时

### 2.5 因果链查询

- [ ] **2.5.1 实现 Causal Query API**
  - [ ] 实现 `FindCausedBy(EventID)` 查询
  - [ ] 实现 `FindCaused(NodeID)` 查询
  - [ ] 实现 `GetFullChain(NodeID)` 获取完整链
  - [ ] 实现 `GetFrameSummary()` 帧摘要
  - [ ] 添加查询测试
  - **文件**: `devtools/causal/query.go` (新建)
  - **预估**: 4 小时

---

## 五、阶段3: 时间旅行系统 (P2)

> **目标**: 实现状态回溯能力，可以"回到任意帧"
> **验收标准**:
> - 可以回溯到历史任意帧
> - 回溯后状态完全一致
> - 内存占用可控

### 3.1 稀疏快照系统

- [ ] **3.1.1 实现 StateSnapshot**
  - [ ] 定义 `StateSnapshot` 结构
  - [ ] 定义 `StateBlob` 紧凑存储
  - [ ] 实现 `Capture()` 捕获状态
  - [ ] 实现 `Restore()` 恢复状态
  - [ ] 优化存储格式 (避免 map/interface{})
  - **文件**: `devtools/timetravel/snapshot.go` (新建)
  - **预估**: 5 小时

- [ ] **3.1.2 实现 Snapshot 调度**
  - [ ] 定义快照间隔 (默认 120 帧)
  - [ ] 实现自动快照触发
  - [ ] 实现快照压缩
  - [ ] 实现旧快照淘汰策略
  - [ ] 添加内存使用监控
  - **文件**: `devtools/timetravel/scheduler.go` (新建)
  - **预估**: 3 小时

### 3.2 时间旅行存储

- [ ] **3.2.1 实现 TimeTravelStore**
  - [ ] 创建 `TimeTravelStore` 结构
  - [ ] 实现 snapshots map 存储
  - [ ] 实现 mutations log 存储
  - [ ] 实现 `SaveFrame()` 保存帧
  - [ ] 实现 `Rewind()` 回溯算法
  - [ ] 实现 `findNearestSnapshot()` 查找最近快照
  - [ ] 实现 `replay()` 从快照重放
  - **文件**: `devtools/timetravel/store.go` (新建)
  - **预估**: 6 小时

### 3.3 重放引擎

- [ ] **3.3.1 实现 Replay Engine**
  - [ ] 创建 `ReplayEngine` 结构
  - [ ] 实现 `ApplyMutation()` 应用单个变更
  - [ ] 实现 `ReplayToFrame()` 重放到指定帧
  - [ ] 实现 `StepForward()` 单步前进
  - [ ] 实现 `StepBackward()` 单步后退
  - [ ] 确保重放时不触发新 mutation 记录
  - **文件**: `devtools/timetravel/replay.go` (新建)
  - **预估**: 5 小时

### 3.4 Runtime 重放模式

- [ ] **3.4.1 添加 Replay Mode 支持**
  - [ ] 在 Runtime 添加 `replayMode` 标志
  - [ ] 在 Replay 模式下跳过真实输入
  - [ ] 在 Replay 模式下跳过 mutation 记录
  - [ ] 在 Replay 模式下跳过真实渲染
  - [ ] 添加 `SetReplayMode()` 接口
  - **文件**: `runtime/runtime.go`
  - **预估**: 2 小时

---

## 六、阶段4: 确定性回放 (P2)

> **目标**: 实现 UI 行为的完全复现，用于 Bug 录制和分享
> **验收标准**:
> - 同一输入序列产生完全相同的 UI 状态
> - 可以导出/导入录制文件

### 4.1 Input Recorder

- [ ] **4.1.1 实现 InputRecorder**
  - [ ] 定义 `InputEvent` 结构 (确定性)
  - [ ] 定义 `InputType` 枚举
  - [ ] 实现 `Record()` 记录输入
  - [ ] 实现 `Save()` 保存到文件
  - [ ] 实现 `Load()` 从文件加载
  - [ ] 确保不记录时间戳 (只记录逻辑输入)
  - **文件**: `devtools/replay/input.go` (新建)
  - **预估**: 4 小时

- [ ] **4.1.2 集成到 Input 处理**
  - [ ] 在 `Runtime.HandleInput()` 添加记录
  - [ ] 拦截所有键盘事件
  - [ ] 拦截所有鼠标事件
  - [ ] 确保输入记录不影响性能
  - **文件**: `runtime/input/handler.go`
  - **预估**: 2 小时

### 4.2 确定性前提

- [ ] **4.2.1 审查时间使用**
  - [ ] 审查所有 `time.Now()` 使用
  - [ ] 替换为 `FrameTime`
  - [ ] 添加静态检查规则
  - **预估**: 3 小时

- [ ] **4.2.2 审查随机数使用**
  - [ ] 审查所有随机数使用
  - [ ] 确保使用固定种子 PRNG
  - [ ] 添加确定性测试
  - **预估**: 2 小时

- [ ] **4.2.3 审查并发使用**
  - [ ] 确保 UI 主线程单线程模型
  - [ ] 审查所有 goroutine 使用
  - [ ] 确保无竞态条件
  - **预估**: 2 小时

### 4.3 Replay 系统

- [ ] **4.3.1 实现 Replay 系统**
  - [ ] 实现 `Replay()` 重放输入
  - [ ] 实现 `LoadSnapshot()` 加载快照
  - [ ] 实现 `LoadInputStream()` 加载输入流
  - [ ] 实现逐帧重放逻辑
  - **文件**: `devtools/replay/replay.go` (新建)
  - **预估**: 4 小时

### 4.4 回放对比

- [ ] **4.4.1 实现回放对比**
  - [ ] 实现 `CompareReplay()` 对比两次回放
  - [ ] 检测状态差异
  - [ ] 检测布局差异
  - [ ] 检测渲染差异
  - [ ] 生成差异报告
  - **文件**: `devtools/replay/compare.go` (新建)
  - **预估**: 3 小时

---

## 七、阶段5: 客户端实现 (P3)

> **目标**: 提供可视化的调试界面
> **验收标准**:
> - TUI Panel 可在终端内显示调试信息
> - Web Dashboard 可远程调试

### 5.1 通信协议

- [ ] **5.1.1 定义协议格式**
  - [ ] 定义 `DebugMessage` 结构
  - [ ] 定义消息类型枚举
  - [ ] 实现消息编码器
  - [ ] 实现消息解码器
  - [ ] 添加版本协商
  - **文件**: `devtools/protocol/message.go`, `encode.go`, `decode.go` (新建)
  - **预估**: 4 小时

- [ ] **5.1.2 实现流控机制**
  - [ ] 定义 `Transport` 接口
  - [ ] 实现 `Send()` 背压处理
  - [ ] 实现 `IsBackpressured()` 检测
  - [ ] 实现丢数据策略
  - **预估**: 2 小时

### 5.2 TUI Panel

- [ ] **5.2.1 实现 TUI 调试面板**
  - [ ] 创建独立 TUI 窗口
  - [ ] 实现 Layout Inspector 视图
  - [ ] 实现 Repaint Debug 视图
  - [ ] 实现 Event Trace 视图
  - [ ] 实现 Component 树显示
  - [ ] 实现键盘导航
  - **文件**: `devtools/client/tui/panel.go` (新建)
  - **预估**: 8 小时

### 5.3 Web Dashboard

- [ ] **5.3.1 实现 WebSocket 服务器**
  - [ ] 创建 WebSocket Server
  - [ ] 实现客户端连接管理
  - [ ] 实现消息推送
  - [ ] 实现命令接收
  - **文件**: `devtools/client/web/server.go` (新建)
  - **预估**: 4 小时

- [ ] **5.3.2 实现 Web 前端**
  - [ ] 创建 HTML 基础页面
  - [ ] 实现 Component 树可视化
  - [ ] 实现 Layout 视图
  - [ ] 实现 Event 时间线
  - [ ] 实现 Repaint 热力图
  - **文件**: `devtools/client/web/frontend/` (新建目录)
  - **预估**: 12 小时

---

## 八、阶段6: 高级功能 (未来)

> **目标**: 基于 DevTools 数据的智能分析能力
> **状态**: 规划中

### 6.1 性能分析 AI

- [ ] **6.1.1 实现性能热区分析**
  - [ ] 分析 Mutation 频率
  - [ ] 分析 Layout 触发频率
  - [ ] 分析 Repaint 区域
  - [ ] 生成性能报告

- [ ] **6.1.2 实现无效刷新检测**
  - [ ] 检测无视觉变化的 setState
  - [ ] 检测无实际变化的 Layout
  - [ ] 检测过度重绘

- [ ] **6.1.3 实现布局抖动检测**
  - [ ] 检测 Layout 版本反复变化
  - [ ] 检测 Repaint 区域来回变化

### 6.2 自动优化建议

- [ ] **6.2.1 实现优化建议引擎**
  - [ ] 分析状态粒度
  - [ ] 建议状态合并
  - [ ] 建议状态提升
  - [ ] 建议布局优化

- [ ] **6.2.2 实现代码重写**
  - [ ] 实现自动批量 setState
  - [ ] 实现 shouldRepaint 优化
  - [ ] 实现 Layout 缓存插入

---

## 附录

### A. 文件结构总览

```
runtime/
├── node.go                    # 修改: 添加 LayoutVersion
├── runtime.go                 # 修改: 集成 DevTools
├── component.go               # 修改: 添加 mutation hook
├── debug.go                   # 保留: 快照调试
└── debug_id.go                # 新建: ID 注册表

runtime/layout/
├── debug_view.go              # 新建: 调试视图接口
└── engine.go                  # 修改: 添加 layout hook

runtime/focus/
└── manager.go                 # 修改: 添加 focus hook

devtools/
├── devtools.go                # 主入口
├── types.go                   # 核心类型
├── bus.go                     # 异步事件总线
├── tap.go                     # Lock-Free Ring Buffer Tap
├── async_collector.go         # 异步收集器
│
├── delta/
│   ├── layout.go              # 布局增量收集
│   ├── repaint.go             # 重绘增量收集
│   └── event.go               # 事件增量收集
│
├── causal/
│   ├── graph.go               # 因果图
│   ├── builder.go             # 构建器
│   ├── timeline.go            # 帧时间线
│   └── query.go               # 查询 API
│
├── timetravel/
│   ├── snapshot.go            # 快照系统
│   ├── scheduler.go           # 快照调度
│   ├── store.go               # 时间旅行存储
│   └── replay.go              # 重放引擎
│
├── replay/
│   ├── input.go               # 输入记录
│   ├── replay.go              # 回放系统
│   └── compare.go             # 差异对比
│
├── overlay/
│   └── buffer.go              # 独立覆盖层 Buffer
│
├── protocol/
│   ├── message.go             # 消息格式
│   ├── encode.go              # 编码器
│   └── decode.go              # 解码器
│
└── client/
    ├── tui/
    │   └── panel.go           # TUI 调试面板
    └── web/
        ├── server.go          # WebSocket 服务
        └── frontend/          # 前端资源
```

### B. 验收检查清单

#### 阶段 1 验收

- [ ] Debug 关闭时性能无影响 (分支预测成功)
- [ ] Debug 开启时帧率稳定 (开销 < 5%)
- [ ] 大型 UI (1000+ 组件) 可用
- [ ] 长时间运行无内存泄漏
- [ ] 增量收集正确 (无遗漏/重复)
- [ ] 背压机制有效 (无阻塞/堆积)

#### 阶段 2 验收

- [ ] 因果链完整无丢失
- [ ] Event → Mutation 关联正确
- [ ] Mutation → Layout 关联正确
- [ ] Layout → Repaint 关联正确
- [ ] 查询 API 性能良好

#### 阶段 3 验收

- [ ] 快照功能正常
- [ ] 回溯功能正确
- [ ] 重放结果一致
- [ ] 内存占用可控 (< 10MB / 1000 帧)

#### 阶段 4 验收

- [ ] 输入记录完整
- [ ] 回放结果完全一致
- [ ] 对比功能正常

### C. 风险和缓解措施

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 性能回归 | 高 | 中 | 严格的性能测试，异步处理 |
| 内存泄漏 | 高 | 低 | 定期压力测试，Ring Buffer |
| 状态不一致 | 高 | 中 | 完整的单元测试 |
| 协议兼容性 | 中 | 低 | 版本协商机制 |

### D. 参考资料

- Chrome DevTools Protocol
- Flutter Observatory
- React DevTools (Fiber)
- Browser Performance Timeline

---

## 更新日志

| 日期 | 版本 | 变更内容 |
|------|------|----------|
| 2026-01-30 | 1.0 | 初始版本，基于架构审查文档创建 |
