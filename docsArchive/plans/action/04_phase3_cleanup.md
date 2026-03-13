# Phase 3: 清理旧代码

## 目标

移除废弃的 Event 系统和适配器代码，简化代码库。

## 前置条件

- Phase 2 所有组件迁移完成
- 所有测试通过
- 线上运行稳定（建议观察 1-2 周）

## 1. 废弃代码识别

### 1.1 可删除的文件/代码

```
framework/
├── event/
│   ├── handler.go          # 删除：Router, KeyMap
│   ├── msg_adapter.go      # 删除：MsgToEvent 转换器
│   ├── pump.go             # 保留：Pump 仍用于原始输入
│   └── event.go            # 精简：保留 BaseEvent 用于内部
│
├── app.go
│   ├── legacyRouter        # 删除
│   ├── legacyMode          # 删除
│   ├── handleLegacyMsg()   # 删除
│   └── handleEvent()       # 删除（或精简为只处理系统事件）
│
└── action/
    └── adapters.go         # 删除：迁移完成后不再需要
```

### 1.2 可删除的接口

```go
// 删除这些接口（组件不再需要实现）
type Updater interface {
    Update(msg Msg) Cmd
}

type EventHandler interface {
    HandleEvent(Event) bool
}

type Component interface {
    HandleEvent(Event) bool
}
```

## 2. 清理步骤

### 2.1 删除 Legacy 字段

```go
// framework/app.go

type App struct {
    // ===== 删除这些字段 =====
    // legacyMode     bool                     // 删除
    // legacyRouter   *frameworkevent.Router   // 删除
    // router         *frameworkevent.Router   // 删除（如果是独立的）
    // keyMap         *frameworkevent.KeyMap   // 删除（已合并到 InputProcessor）

    // ===== 保留字段 =====
    actionRouter   *action.Router
    inputProcessor *action.InputProcessor
    actionRegistry *ActionRegistry
    pump           *frameworkevent.Pump  // 保留，仍需要
}
```

### 2.2 简化消息处理

```go
// framework/app.go

// processMsg 简化后的版本
func (a *App) processMsg(msg runtimemsg.Msg) {
    if msg == nil {
        return
    }

    // 转换为 Action
    act := a.inputProcessor.ProcessMsg(msg)
    if act == nil {
        // 系统消息（如 Resize）特殊处理
        if resizeMsg, ok := msg.(*runtimemsg.ResizeMsg); ok {
            a.Resize(resizeMsg.Width, resizeMsg.Height)
        }
        return
    }

    // 设置默认目标
    if act.TargetID == 0 {
        if focused := a.focusManager.GetCurrent(); focused != nil {
            act.TargetID = focused.GetNodeID()
        }
    }

    // 分发
    result := a.actionRouter.Dispatch(act)
    if result.Handled {
        a.dirty = true
    }
}

// 删除 handleLegacyMsg()
// 删除 handleEvent()（或简化为只处理系统事件）
```

### 2.3 清理 event 包

```go
// framework/event/event.go

// 保留：基础事件类型（用于内部）
type BaseEvent struct {
    eventType EventType
    timestamp time.Time
}

// 保留：事件类型常量（部分）
const (
    EventResize EventType = iota
    EventQuit
    // 删除：EventKeyPress, EventMousePress 等（已被 Action 替代）
)

// 删除：Router 相关代码 → 移到 action/router.go
// 删除：KeyMap 相关代码 → 移到 action/keymap.go
// 删除：MsgToEvent 转换器 → 不再需要
```

### 2.4 删除适配器

```go
// 删除整个文件：framework/action/adapters.go
// 或者只保留文档说明

// adapters.go - DEPRECATED
// 此文件中的适配器已完成使命，不再需要。
// 所有组件现在直接实现 ActionTarget 接口。
//
// 保留此文件仅供参考，将在下个版本删除。
```

## 3. 代码差异对比

### 3.1 App 结构

```go
// ===== 清理前 =====
type App struct {
    router       *frameworkevent.Router
    keyMap       *frameworkevent.KeyMap
    pump         *frameworkevent.Pump
    legacyMode   bool
    legacyRouter *frameworkevent.Router
    // ...
}

// ===== 清理后 =====
type App struct {
    actionRouter   *action.Router
    inputProcessor *action.InputProcessor
    actionRegistry *ActionRegistry
    pump           *frameworkevent.Pump
    // ...
}
```

### 3.2 组件接口

```go
// ===== 清理前 =====
type Button struct {
    text    string
    onClick func()
}

func (b *Button) HandleEvent(ev frameworkevent.Event) bool { ... }
func (b *Button) Update(msg runtimemsg.Msg) cmd.Cmd { ... }

// ===== 清理后 =====
type Button struct {
    text    string
    onClick func()
    nodeID  uint64
}

func (b *Button) HandleAction(action *action.Action) bool { ... }
func (b *Button) GetSupportedActions() []action.ActionType { ... }
func (b *Button) CanHandleAction(action *action.Action) bool { ... }
```

### 3.3 消息处理流程

```
===== 清理前 =====
RawInput → Pump → Msg
                    ├→ handleMsg() → Instance.Handle()
                    └→ MsgToEvent() → handleEvent() → Router → Component

===== 清理后 =====
RawInput → Pump → Msg → InputProcessor → Action → ActionRouter → ActionTarget
```

## 4. 测试清理

### 4.1 删除的测试

```
framework/event/
├── handler_test.go       # 删除
├── router_test.go        # 删除
├── msg_adapter_test.go   # 删除

framework/app/
├── app_legacy_test.go    # 删除
├── app_event_test.go     # 删除（或精简）
```

### 4.2 保留/更新的测试

```
framework/action/
├── action_test.go        # 保留
├── router_test.go        # 保留
├── processor_test.go     # 保留

framework/app/
├── app_action_test.go    # 保留/更新
├── app_test.go           # 更新
```

### 4.3 测试覆盖率要求

清理后保持测试覆盖率：

| 包 | 清理前 | 清理后目标 |
|----|--------|-----------|
| framework | 75% | >= 75% |
| framework/action | 80% | >= 80% |
| framework/components | 70% | >= 70% |

## 5. 文档更新

### 5.1 需要更新的文档

```
docs/
├── getting-started.md    # 更新：使用 ActionTarget
├── components.md         # 更新：组件接口变更
├── events.md             # 删除/重写：改为 actions.md
├── migration-guide.md    # 新增：旧版本迁移指南
└── api/
    └── reference.md      # 更新：API 变更
```

### 5.2 迁移指南示例

```markdown
# 从 Event 系统迁移到 Action 系统

## 概述

从 v0.x 版本开始，框架使用 Action 系统替代 Event 系统。
本文档帮助你迁移现有代码。

## 接口变更

### 旧接口（已废弃）

\`\`\`go
type EventHandler interface {
    HandleEvent(Event) bool
}
\`\`\`

### 新接口

\`\`\`go
type ActionTarget interface {
    HandleAction(action *Action) bool
    GetSupportedActions() []ActionType
    CanHandleAction(action *Action) bool
}
\`\`\`

## 迁移示例

### Button 组件

**旧代码：**
\`\`\`go
func (b *Button) HandleEvent(ev frameworkevent.Event) bool {
    if ev.Type() == frameworkevent.EventClick {
        b.onClick()
        return true
    }
    return false
}
\`\`\`

**新代码：**
\`\`\`go
func (b *Button) HandleAction(action *action.Action) bool {
    if action.Type == action.ActionClick {
        b.onClick()
        return true
    }
    return false
}
\`\`\`
```

## 6. 发布计划

### 6.1 版本号

- **v0.x → v1.0**：重大变更，移除旧系统

### 6.2 发布说明模板

```markdown
# Release v1.0.0

## 重大变更

### 统一 Action 系统

所有用户交互现在通过 Action 系统处理，移除了旧的 Event 系统。

**变更摘要：**
- 组件接口从 `EventHandler` 改为 `ActionTarget`
- 消息传播使用 `ActionRouter` 的三阶段模型
- 移除了 `framework/event` 包中的 `Router` 和 `KeyMap`

**迁移指南：** 参见 `docs/migration-guide.md`

## 新功能

- 语义化 Action 类型（54+ 种）
- 三阶段传播模型（Capture → Target → Bubble）
- Action 中间件支持
- Action 对象池（性能优化）

## Bug 修复

- ...

## 贡献者

- ...
```

## 7. 回滚方案

如果清理后发现问题：

1. **短期回滚**：恢复到 v0.x 版本
2. **热修复**：在 v1.0.x 中修复问题
3. **兼容层**：临时添加适配器作为补丁

```go
// 紧急补丁：临时兼容层
func (a *App) processMsgWithFallback(msg runtimemsg.Msg) {
    // 尝试 Action 路径
    act := a.inputProcessor.ProcessMsg(msg)
    if act != nil {
        if result := a.actionRouter.Dispatch(act); result.Handled {
            a.dirty = true
            return
        }
    }

    // 回退到旧路径（临时）
    // ... 需要保留部分旧代码
}
```

## 8. Phase 3 完成标准

- [ ] 删除所有废弃代码
- [ ] 更新所有文档
- [ ] 测试覆盖率达标
- [ ] 发布 v1.0.0
- [ ] 更新 CHANGELOG
- [ ] 发布迁移指南
