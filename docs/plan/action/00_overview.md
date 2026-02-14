# Action 统一消息传播方案 - 总览

## 目标

将当前的三套消息传播机制（Msg/Event/Action）统一到 Action 系统，实现：

- **单一传播路径**：所有用户交互都通过 Action 处理
- **语义化操作**：组件处理的是"用户意图"而非"原始输入"
- **可预测的传播**：三阶段传播（Capture → Target → Bubble）
- **易于扩展**：支持自定义 Action 和异步操作

## 当前状态

```
┌─────────────────────────────────────────────────────────────────────┐
│                      当前消息传播架构（问题）                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  RawInput ─→ Pump ─→ Msg                                            │
│                      │                                              │
│         ┌────────────┼────────────┐                                 │
│         ▼            ▼            ▼                                 │
│    ┌─────────┐  ┌─────────┐  ┌─────────┐                           │
│    │ handleMsg│  │MsgToEvent│ │InputProc│                          │
│    │ (直接路由)│  │(转换器)  │ │(Action) │                          │
│    └─────────┘  └─────────┘  └─────────┘                           │
│         │            │            │                                 │
│         ▼            ▼            ▼                                 │
│    ┌─────────┐  ┌─────────┐  ┌─────────┐                           │
│    │ Instance│  │ Event   │  │ Action  │ ← 未被使用！              │
│    │ .Handle │  │ Router  │  │ Router  │                           │
│    └─────────┘  └─────────┘  └─────────┘                           │
│                      │                                              │
│                      ▼                                              │
│               Component.HandleEvent()                               │
│                                                                     │
│  问题：                                                              │
│  1. 三条路径并存，职责重叠                                           │
│  2. Action Router 已实现但未集成                                     │
│  3. 组件需要实现多个接口（Updater, EventHandler, ActionTarget）       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## 目标架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                      统一后的 Action 架构                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  RawInput ─→ Pump ─→ Msg                                            │
│                      │                                              │
│                      ▼                                              │
│               ┌─────────────┐                                       │
│               │ InputProc   │  ← KeyMap（快捷键映射）                │
│               │ .Process()  │                                       │
│               └─────────────┘                                       │
│                      │                                              │
│                      ▼                                              │
│               ┌─────────────┐                                       │
│               │   Action    │  语义化操作                           │
│               │ Type+Payload│                                       │
│               └─────────────┘                                       │
│                      │                                              │
│                      ▼                                              │
│               ┌─────────────┐                                       │
│               │ActionRouter │                                       │
│               │ .Dispatch() │                                       │
│               └─────────────┘                                       │
│                      │                                              │
│         ┌────────────┼────────────┐                                 │
│         ▼            ▼            ▼                                 │
│     Capture       Target       Bubble                               │
│     (拦截器)      (目标组件)    (冒泡)                               │
│                      │                                              │
│                      ▼                                              │
│               ┌─────────────┐                                       │
│               │ActionTarget │  唯一的组件接口                       │
│               │.HandleAction│                                       │
│               └─────────────┘                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## 迁移计划

| 阶段 | 内容 | 文档 |
|------|------|------|
| Phase 0 | 架构设计与接口定义 | `01_architecture.md` |
| Phase 1 | 核心系统集成 | `02_phase1_integration.md` |
| Phase 2 | 组件迁移 | `03_phase2_migration.md` |
| Phase 3 | 清理旧代码 | `04_phase3_cleanup.md` |
| Phase 4 | 增强功能 | `05_phase4_enhancements.md` |

## 关键决策

### 1. 保留 Msg 作为底层传输格式

- Msg 是 Pump 产生的原始消息格式
- InputProcessor 负责将 Msg 转换为 Action
- 好处：不改动底层输入系统

### 2. 统一组件接口为 ActionTarget

```go
// 旧接口（废弃）
type Updater interface {
    Update(msg Msg) Cmd
}

type EventHandler interface {
    HandleEvent(Event) bool
}

// 新接口（统一）
type ActionTarget interface {
    HandleAction(action *Action) bool
    GetSupportedActions() []ActionType
    CanHandleAction(action *Action) bool
}
```

### 3. 三阶段传播模型

```
Capture Phase: 从根到目标，高优先级处理器可拦截
    ↓
Target Phase: 在目标组件执行 HandleAction
    ↓
Bubble Phase: 从目标到根，父组件可处理未消费的 Action
```

### 4. 兼容性策略

- 提供适配器将旧接口包装为 ActionTarget
- 分阶段迁移，保持系统可运行
- 每个阶段都有回滚点

## 风险与缓解

| 风险 | 缓解措施 |
|------|---------|
| 大量组件需要改动 | 提供通用适配器和基类 |
| 现有行为可能改变 | 保留 Msg→Event 回退路径 |
| 性能影响 | Action 对象池复用 |
| 学习成本 | 提供迁移指南和示例 |

## 成功指标

- [ ] 所有用户交互都通过 Action 传播
- [ ] 组件只需实现 ActionTarget 接口
- [ ] ActionRouter 成为唯一的事件分发器
- [ ] 删除 framework/event 包中的 Router
- [ ] 测试覆盖率不低于当前水平
