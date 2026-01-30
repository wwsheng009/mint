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
│              └──> Client (调试界面)                              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 📚 实施阶段

```
阶段 1: ███████████████████████████████████  增量基础 (P0) ✅
阶段 2: ███████████████████████████████████  因果链 (P1) ✅
阶段 3: ███████████████████████████████████  时间旅行 (P2) ✅
阶段 4: ███████████████████████████████████  确定性回放 (P2) ✅
阶段 5: ███████████████████████████████████  客户端 (P3) ✅
阶段 6: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  高级功能 (未来)
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

### ✅ 已完成 (阶段 1 + 阶段 2 + 阶段 3)

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

**阶段 5: 客户端 (P3)**
- [x] TUI 调试面板 (TuiDebugPanel)
- [x] WebSocket 协议 (WebSocketHandler)
- [x] Web Dashboard
- [x] 远程调试会话 (RemoteDebugSession)
- [x] API 处理器 (APIHandler)
- [x] 调试覆盖层 (DebugOverlay)
- [x] 性能分析器 (Profiler)

**测试状态**
- [x] 编译通过
- [x] 16/16 单元测试通过

### 📋 计划中 (阶段 6)

**阶段 6: 高级功能**
- [ ] 性能分析 AI
- [ ] 自动异常检测

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
