# TUI DevTools 实施计划 TODO LIST

> **项目**: Mint TUI Runtime
> **文档版本**: 1.1
> **创建日期**: 2026-01-30
> **更新日期**: 2026-01-31
> **状态**: 阶段 1、3、4、5 已完成，阶段 2 待实施

> **文档位置**: `devtools/docs/`
>
> **实施基础**:
> - 设计文档: `implementation_v2.md`
> - 原始评审: `../../framework/docs/buffer/idea5_review.md`

---

## 📊 当前状态 (2026-01-30)

### 阶段 1: 增量基础 (P0) - ✅ 已完成

### 阶段 3: 时间旅行系统 (P2) - 🚧 实施

### 阶段 4: 确定性回放 (P2) - 🚧 实施

### 阶段 5: 客户端/远程调试 (P3) - 🚧 实施

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
阶段 2: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  因果链 (P1)   0% ⏭ 跳过
阶段 3: █████████████████████████████████░░░░░░░  时间旅行 (P2) 90% 🚧
阶段 4: █████████████████████████████████░░░░░░░  确定性回放 (P2) 90% 🚧
阶段 5: █████████████████████████████████░░░░░░░  客户端/远程 (P3) 90% 🚧
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
    ├── causal/               # 阶段2: 因果链引擎 (待实施 - 已跳过)
    ├── timetravel/           # 阶段3: 时间旅行 (已实现)
    ├── replay/               # 阶段4: 确定性回放 (已实现)
    └── remote/               # 阶段5: 客户端/远程调试 (已实现)
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
│  阶段 2: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  因果链 (P1)  ⏭ 跳过    │
│    └─ 项目选择跳过此阶段，直接实现了时间旅行功能                 │
│                                                                  │
│  阶段 3: ████████████████████████░░░░░░░░░░░  时间旅行 (P2) 🚧│
│    ├─ 稀疏快照系统 (snapshot.go) ✅                                │
│    ├─ 重放引擎 (replay.go) ✅                                      │
│    ├─ 时间游标 (cursor.go) ✅                                      │
│    └─ 时间轴 UI (diffengine.go, client.go) ✅                       │
│                                                                  │
│  阶段 4: ████████████████████████░░░░░░░░░░░  确定性回放 (P2) 🚧│
│    ├─ Input Recorder (input.go) ✅                                │
│    ├─ Replay 模式 (recorder.go, replayer.go) ✅                   │
│    ├─ 回放比对 (determinism.go, seed.go) ✅                       │
│    └─ 确定性验证 ✅                                                │
│                                                                  │
│  阶段 5: ████████████████████████░░░░░░░░░░░  客户端/远程 (P3) 🚧│
│    ├─ Remote Debugging (chromium.go) ✅                            │
│    ├─ HTTP API (http_server.go) ✅                                 │
│    └─ 协议优化 (protocol/) ✅
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

- [x] **3.1.1 实现 StateSnapshot**
  - [x] 定义 `StateSnapshot` 结构
  - [x] 定义 `StateBlob` 紧凑存储
  - [x] 实现 `Capture()` 捕获状态
  - [x] 实现 `Restore()` 恢复状态
  - [x] 优化存储格式 (避免 map/interface{})
  - **文件**: `devtools/timetravel/snapshot.go` ✅

- [x] **3.1.2 状态回放器**
  - [x] 定义 `StateReplay` 结构 (集成在 replay.go)
  - [x] 实现播放控制 (Play/Pause/Stop)
  - [x] 实现速度控制
  - [x] 实现进度跟踪
  - **文件**: `devtools/timetravel/replay.go` ✅
  - **状态**: 已完成 (快照调度功能已整合)

### 3.2 时间旅行游标

- [x] **3.2.1 实现 TimeTravelCursor**
  - [x] 创建 `TimeTravelCursor` 结构
  - [x] 实现 `MoveTo()` 移动到指定帧
  - [x] 实现 `Next()` / `Prev()` 前进/后退
  - [x] 实现 `First()` / `Last()` 跳转到首尾
  - [x] 实现 `AddBookmark()` / `GoToBookmark()` 书签管理
  - **文件**: `devtools/timetravel/cursor.go` ✅
  - **状态**: 已完成 (集成在 cursor.go 中)

### 3.3 重放引擎

- [x] **3.3.1 实现 Replay Engine**
  - [x] 创建 `ReplayEngine` 结构
  - [x] 实现 `ApplyMutation()` 应用单个变更
  - [x] 实现 `ReplayToFrame()` 重放到指定帧
  - [x] 实现 `StepForward()` 单步前进
  - [x] 实现 `StepBackward()` 单步后退
  - [x] 确保重放时不触发新 mutation 记录
  - **文件**: `devtools/timetravel/replay.go` ✅
  - **状态**: 已完成

### 3.4 差异引擎

- [x] **3.4.1 实现 Diff Engine**
  - [x] 创建 `DiffEngine` 结构
  - [x] 实现 `CompareStates()` 比较两个状态
  - [x] 实现 `DiffPath()` 计算差异路径
  - [x] 生成 StateChange 列表
  - **文件**: `devtools/timetravel/diffengine.go` ✅
  - **状态**: 已完成

- [x] **3.4.2 实现 TimeTravelClient**
  - [x] 创建 `TimeTravelClient` 结构
  - [x] 实现 `Render()` 渲染当前状态
  - [x] 实现 `HandleInput()` 处理键盘输入
  - [x] 集成 TUI 界面
  - **文件**: `devtools/timetravel/client.go` ✅
  - **状态**: 已完成

---

## 六、阶段4: 确定性回放 (P2)

> **目标**: 实现 UI 行为的完全复现，用于 Bug 录制和分享
> **验收标准**:
> - 同一输入序列产生完全相同的 UI 状态
> - 可以导出/导入录制文件

### 4.1 Input Recorder

- [x] **4.1.1 实现 InputRecorder**
  - [x] 定义 `InputEvent` 结构 (确定性)
  - [x] 定义 `InputType` 枚举
  - [x] 实现 `Record()` 记录输入
  - [x] 实现 `GetSequence()` 获取输入序列
  - [x] 确保不记录时间戳 (只记录逻辑输入)
  - **文件**: `devtools/replay/input.go` ✅

### 4.2 确定性前提

### 4.2 确定性前提

- [x] **4.2.1 Determinism Checker**
  - [x] 定义 `DeterminismChecker` 结构
  - [x] 实现 `Compare()` 比较原始和回放
  - [x] 实现 `Verify()` 验证确定性
  - [x] 生成 Difference 报告
  - **文件**: `devtools/replay/determinism.go` ✅

- [x] **4.2.2 Seed Capture**
  - [x] 定义 `SeedCapture` 结构
  - [x] 实现 `Capture()` 捕获种子
  - [x] 实现 `Get()` 获取种子
  - [x] 实现 `Export()/Import()` 导出/导入
  - **文件**: `devtools/replay/seed.go` ✅

### 4.3 Replay 系统

### 4.3 Replay 系统

- [x] **4.3.1 实现 EventRecorder**
  - [x] 创建 `EventRecorder` 结构
  - [x] 实现 `StartSession()` / `EndSession()`
  - [x] 实现 `RecordEvent()` 记录事件
  - [x] 支持 `RecordingSession` 管理
  - **文件**: `devtools/replay/recorder.go` ✅

- [x] **4.3.2 实现 EventReplayer**
  - [x] 创建 `EventReplayer` 结构
  - [x] 实现 `Start()` / `Next()` 开始/步进
  - [x] 实现 `Pause()` / `Resume()` 暂停/继续
  - [x] 实现 `SetSpeed()` 设置速度
  - **文件**: `devtools/replay/replayer.go` ✅

### 4.4 回放对比

### 4.4 回放对比

- [x] **4.4.1 实现回放对比**
  - [x] 实现 `Compare()` 比较两次回放
  - [x] 检测状态差异
  - [x] 生成差异报告
  - **文件**: `devtools/replay/determinism.go` ✅
  - **状态**: 已完成

---

## 七、阶段5: 客户端实现 (P3)

> **目标**: 提供可视化的调试界面
> **验收标准**:
> - TUI Panel 可在终端内显示调试信息
> - Web Dashboard 可远程调试

### 5.1 通信协议

- [x] **5.1.1 定义协议格式**
  - [x] 定义 `DebugMessage` / `Message` 结构
  - [x] 定义消息类型枚举 (握手、快照、差异等)
  - [x] 实现消息编码/解码器
  - [x] 添加版本协商
  - **文件**: `devtools/protocol/message.go` ✅

- [x] **5.1.2 实现流控机制**
  - [x] 定义 `Session` / `Connection` 结构
  - [x] 实现客户端连接管理
  - [x] 实现订阅机制
  - **文件**: `devtools/remote/` ✅

### 5.2 Remote Debugging Server

- [x] **5.2.1 实现 Chromium Bridge**
  - [x] 创建 `ChromiumBridge` 结构
  - [x] 实现 `GetInspectorHTML()` 获取 Inspector 界面
  - [x] 实现 `ExportForChromium()` 导出 CDP 格式
  - [x] 兼容 Chrome DevTools Protocol
  - **文件**: `devtools/remote/chromium.go` ✅

- [x] **5.2.2 实现 HTTP Server**
  - [x] 创建 `HTTPServer` 结构
  - [x] 实现 REST API (/api/snapshots, /api/diff, etc.)
  - [x] 实现健康检查
  - **文件**: `devtools/remote/http_server.go` ✅

- [x] **5.2.3 实现 WebSocket Server**
  - [x] 创建 `WebSocketServer` 结构
  - [x] 实现 WebSocket 连接管理
  - [x] 实现消息推送/命令接收
  - **文件**: `devtools/remote/` ✅
  - **状态**: 已完成

### 5.3 Web Dashboard

- [x] **5.3.1 实现 Web Dashboard**
  - [x] 创建 HTML 基础页面 (Inspector)
  - [x] 实现 Component 树可视化
  - [x] 实现 Layout 视图
  - [x] 实现快照对比
  - **文件**: `devtools/remote/http_server.go` (内置 Inspector HTML) ✅
  - **状态**: 已完成

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
│   ├── snapshot.go            # 快照系统 ✅
│   ├── cursor.go              # 时间游标 ✅
│   ├── replay.go              # 重放引擎 ✅
│   ├── diffengine.go          # 差异引擎 ✅
│   └── client.go              # TUI 客户端 ✅
│
├── replay/
│   ├── input.go               # 输入记录 ✅
│   ├── recorder.go            # 事件录制器 ✅
│   ├── replayer.go            # 事件回放器 ✅
│   ├── determinism.go         # 确定性验证器 ✅
│   └── seed.go                # 随机种子捕获 ✅
│
├── overlay/
│   └── buffer.go              # 独立覆盖层 Buffer
│
├── protocol/
│   └── message.go             # 消息格式 ✅

├── remote/
│   ├── chromium.go            # Chromium 桥接器 ✅
│   └── http_server.go         # HTTP 服务器 ✅
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
| 2026-01-31 | 1.1 | 更新状态：同步 Phase 3、4、5 实际实现状态；Phase 2 标记为跳过 |
