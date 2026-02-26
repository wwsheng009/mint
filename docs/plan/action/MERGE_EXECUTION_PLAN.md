# Framework/Action 与 Runtime/Action 合并执行方案

## 文档信息

| 项目 | 内容 |
|------|------|
| **标题** | Framework/Action 与 Runtime/Action 合并执行方案 |
| **版本** | v1.0 |
| **日期** | 2026年2月25日 |
| **状态** | 待执行 |

---

## 一、概述

### 1.1 背景

项目中存在两个 Action 系统，导致功能重复、接口不一致、依赖混乱：

| 系统 | 特点 | 问题 |
|------|------|------|
| `framework/action` | 三阶段路由、中间件、作用域分发、KeyMap | TargetID 用 uint64，缺少复合 Action |
| `runtime/action` | 复合 Action、WorkerPool、错误处理 | Target 用 string，缺少路由和中间件 |

### 1.2 目标

1. **消除重复**：合并两个 Action 系统到 `runtime/action`
2. **功能增强**：保留两者的所有优点，删除不足
3. **向后兼容**：通过适配器支持旧代码
4. **彻底清理**：删除 `framework/action` 目录

### 1.3 合并原则

- **保留最强功能**：三阶段路由 + 复合 Action + 中间件
- **统一接口**：Target 统一使用 string ID，内部维护 uint64 映射
- **渐进迁移**：通过适配器支持过渡期
- **测试驱动**：每个阶段必须有测试验证

---

## 二、最终架构设计

### 2.1 统一的 Action 结构体

```go
package action

type Action struct {
    // 核心字段
    Type      ActionType
    Payload   interface{}
    Source    string

    // 目标标识（支持两种方式）
    Target    string       // 主要: 语义化组件 ID（如 "button.submit"）
    TargetID  uint64       // 可选: 内部 NodeID（用于快速查找）

    // 追踪和调试（来自 framework）
    ID        uint64
    Timestamp time.Time
    Meta      map[string]interface{}
    stopped   bool

    // 内部字段（不要直接修改）
    mu        sync.RWMutex
}
```

### 2.2 统一的 Target 接口体系

```go
// 核心 Target（简化版）
type Target interface {
    ID() string
    HandleAction(a *Action) bool
}

// 能力接口（可选实现）
type Focusable interface {
    Focus() bool
    Blur()
    IsFocused() bool
    IsFocusable() bool
}

type Scrollable interface {
    CanScroll(delta int) bool
    Scroll(delta int) bool
    GetScrollPosition() (current, total, visible int)
}

type Editable interface {
    InsertText(text string) bool
    DeleteText(direction int) bool
    GetText() string
    SetCursorPosition(pos int) bool
}

type Selectable interface {
    Select() bool
    ToggleSelection() bool
    IsSelected() bool
    GetSelectedCount() int
}

type Expandable interface {
    Expand() bool
    Collapse() bool
    Toggle() bool
    IsExpanded() bool
}

type Draggable interface {
    StartDrag(act *Action) bool
    Drag(act *Action) bool
    EndDrag(act *Action) bool
    IsDragging() bool
}

// 组合接口（声明式能力）
type FocusableTarget interface {
    Target
    Focusable
}

type ScrollableTarget interface {
    Target
    Scrollable
}
```

### 2.3 ActionType 常量（合并版）

| 类别 | 常量（合并去重后） |
|------|-------------------|
| **导航** | NavigateFirst, NavigateLast, NavigateNext, NavigatePrev, NavigateUp, NavigateDown, NavigateLeft, NavigateRight, NavigatePageUp, NavigatePageDown, NavigateHome, NavigateEnd, FocusNext, FocusPrev |
| **编辑** | InputChar, InputText, DeleteChar, DeleteWord, DeleteLine, Backspace, CursorHome, CursorEnd, CursorLeft, CursorRight, CursorWordLeft, CursorWordRight, SelectAll, SelectWord, SelectLine |
| **鼠标** | Click, DoubleClick, TripleClick, MousePress, MouseRelease, MouseMotion, MouseDrag, MouseWheel, MouseWheelUp, MouseWheelDown, RightClick, MiddleClick, MiddlePress, Hover |
| **表单** | Submit, Cancel, Validate, Reset, Clear |
| **系统** | Quit, Close, Maximize, Minimize, Fullscreen, Inspect, Refresh, Quit, Copy, Cut, Paste, Undo, Redo, Search, Help |
| **数据** | DataLoad, DataUpdate, DataError |
| **焦点** | Focus, Blur, FocusGained, FocusLost |
| **AI** | AIInspect, AIFind, AIQuery, AIDispatch, AIWait, AIWatch |
| **视图** | Scroll, ScrollUp, ScrollDown, ScrollLeft, ScrollRight, ZoomIn, ZoomOut, ZoomReset, Resize |
| **复合** | Init, Mount, Unmount, Select, Toggle, Expand, Collapse |

### 2.4 文件结构（最终）

```
runtime/action/
├── action.go          # 核心 Action 结构体 + ActionType 常量
├── target.go          # Target 接口体系 + 组合器
├── payload.go         # 强类型 Payload 定义
├── router.go          # Router（三阶段路由 + 中间件）
├── dispatcher.go      # SimpleDispatcher（简化版）
├── middleware.go      # 中间件定义 + 8种内置中间件
├── processor.go       # InputProcessor + KeyMap
├── scope.go           # ScopeDispatcher（作用域分发）
├── composite.go       # 复合 Action（并发/顺序/重试/回退/超时）
├── workerpool.go      # WorkerPool（限制并发数）
├── errors.go          # 结构化错误定义
├── adapters.go        # 向后兼容适配器
├── helpers.go         # 辅助函数和工具
└── README.md          # 更新的文档
```

---

## 三、执行阶段

### 阶段 0: 准备阶段（约 2 小时）

**目标**: 建立基线，分析影响范围

#### 任务列表

| # | 任务 | 验收标准 | 状态 |
|---|------|----------|------|
| 0.1 | 运行所有测试，确保无失败 | 所有测试 PASS | ☐ |
| 0.2 | 列出所有 `framework/action` 的文件 | 记录 13 个文件列表 | ☐ |
| 0.3 | 列出所有 `runtime/action` 的文件 | 记录 10 个文件列表 | ☐ |
| 0.4 | 统计导入 `framework/action` 的文件数 | 记录 131 个文件 | ☐ |
| 0.5 | 统计导入 `runtime/action` 的文件数 | 记录 63 个文件 | ☐ |
| 0.6 | 创建 Git 分支 `merge-action-systems` | 分支创建成功 | ☐ |

---

### 阶段 1: 核心结构合并（约 3 小时）

**目标**: 创建统一的 Action 结构体和常量

#### 1.1 创建 `action.go`（核心）

**文件**: `runtime/action/action.go`

**任务**:
- [ ] 定义统一的 `Action` 结构体
- [ ] 定义合并后的 `ActionType` 常量（去重）
- [ ] 实现 `NewAction`, `NewActionWithPayload`, `NewActionFromKey`, `NewActionFromMouse`
- [ ] 实现 `WithTarget`, `WithPayload`, `WithSource`, `WithMeta`
- [ ] 实现 `Clone` 方法
- [ ] 实现 `StopPropagation`, `IsStopped`
- [ ] 实现 `String` 方法
- [ ] 实现 `ID()` 计算方法和自动赋值
- [ ] 实现 `Timestamp` 自动赋值
- [ ] 实现 `InitializeMeta`
- [ ] 实现 TargetID ↔ String 转换方法
- [ ] 实现 Payload 辅助方法 (`AsString`, `AsInt`, `AsPayload`...)

**验收标准**:
```bash
go test ./runtime/action/... -run TestAction
# 所有测试通过
```

#### 1.2 创建 `target.go`（接口体系）

**文件**: `runtime/action/target.go`

**任务**:
- [ ] 定义核心 `Target` 接口
- [ ] 定义 6 个能力接口 (Focusable, Scrollable, Editable, Selectable, Expandable, Draggable)
- [ ] 定义 2 个组合接口 (FocusableTarget, ScrollableTarget)
- [ ] 实现 `TargetFunc`（函数式适配）
- [ ] 实现 `TargetChain`（责任链）
- [ ] 实现 `BaseActionTarget`（基础实现）
- [ ] 实现 `CompositeActionTarget`（组合器）
- [ ] 实现 `ActionTargetAdapter`（适配器）
- [ ] 实现辅助工具函数 (`GetActionTargets`, `FilterActionTargets`, `FindActionTarget`)

**验收标准**:
```bash
go test ./runtime/action/... -run TestTarget
# 所有测试通过
```

#### 1.3 创建 `payload.go`

**文件**: `runtime/action/payload.go`

**任务**:
- [ ] 定义 `ClickPayload`
- [ ] 定义 `InputPayload`
- [ ] 定义 `ChangePayload`
- [ ] 定义 `KeyPayload`
- [ ] 定义 `FocusPayload`
- [ ] 定义 `SubmitPayload`
- [ ] 定义 `NavigatePayload`
- [ ] 定义 `ResizePayload`
- [ ] 实现所有 Payload 的 `New*` 构造函数
- [ ] 实现 Action 的 `AsXxxPayload()` 方法

**验收标准**:
```bash
go test ./runtime/action/... -run TestPayload
# 所有测试通过
```

---

### 阶段 2: 路由系统迁移（约 4 小时）

**目标**: 迁移 Router 和 Dispatcher 系统

#### 2.1 迁移 `router.go`（三阶段路由）

**文件**: `runtime/action/router.go`

**迁移自**: `framework/action/router.go`

**任务**:
- [ ] 定义 `RouterResult` 结构体
- [ ] 定义 `ActionPhase` 常量
- [ ] 定义 `CaptureActionHandler` 接口
- [ ] 定义 `BubbleActionHandler` 接口
- [ ] 定义 `GlobalActionHandler` 接口
- [ ] 定义 `Router` 结构体
- [ ] 实现 `NewRouter`
- [ ] 实现 `SetMiddleware`, `AddMiddleware`
- [ ] 实现 `AddGlobalHandler`
- [ ] 实现 `AddCaptureHandler` (支持优先级)
- [ ] 实现 `AddBubbleHandler`
- [ ] 实现 `RegisterTarget`, `UnregisterTarget`
- [ ] 实现 `Dispatch` (三阶段逻辑)
- [ ] 实现 `capturePhase`, `targetPhase`, `bubblePhase`
- [ ] 实现 `BuildTargetRegistry` (遍历组件树注册)
- [ ] 实现 `findNodeByID` (递归查找)
- [ ] 实现辅助方法 (`GetRoot`, `SetRoot`, `GetCaptureHandlers`, `GetBubbleHandlers`)

**修改点**:
- `TargetID` 改为支持 `string` 和 `uint64`
- 移除对 `framework/action` 包的依赖
- 使用新的 `Action` 结构体

**验收标准**:
```bash
go test ./runtime/action/... -run TestRouter
# 所有测试通过，包括三阶段传播测试
```

#### 2.2 保留/更新 `dispatcher.go`（简化版）

**文件**: `runtime/action/dispatcher.go`

**任务**:
- [ ] 检查现有 `SimpleDispatcher` 实现
- [ ] 更新 `SimpleDispatcher` 以兼容新的 `Action` 结构体
- [ ] 实现 TargetID ↔ String 转换逻辑
- [ ] 实现 `Register`, `Unregister`
- [ ] 实现 `Subscribe` (全局处理器)
- [ ] 实现 `Dispatch` (全局→目标→默认)
- [ ] 实现 `DispatchToFocus`
- [ ] 实现 `DispatchToTarget`
- [ ] 实现日志功能 (`EnableLog`, `GetLog`, `PrintLog`)
- [ ] 实现统计方法 (`Stats`, `String`)

**验收标准**:
```bash
go test ./runtime/action/... -run TestDispatcher
# 所有测试通过
```

---

### 阶段 3: 中间件系统迁移（约 3 小时）

**目标**: 迁移并扩展中间件

#### 3.1 创建 `middleware.go`

**文件**: `runtime/action/middleware.go`

**迁移自**: `framework/action/middleware.go`

**任务**:
- [ ] 定义 `ActionMiddleware` 接口
- [ ] 定义 `MiddlewareChain` 结构体
- [ ] 实现 `NewMiddlewareChain`
- [ ] 实现 `Before`, `After` 方法
- [ ] 实现 `Add`, `Middlewares` 方法
- [ ] 实现内置 LoggingMiddleware
  - 设置启用/禁用
  - Before 记录开始时间
  - After 计算并记录时长
- [ ] 实现内置 ThrottleMiddleware
  - 配置时间间隔
  - Before 检查是否拦截
- [ ] 实现内置 ValidationMiddleware
  - 注册/注销验证器
  - Before 执行验证
- [ ] 实现内置 MetricsMiddleware
  - Action 计数
  - 处理时长统计
  - 错误计数
  - 格式化统计 (`FormatStats`)
- [ ] 实现内置 RecoveryMiddleware
  - Before 空实现
  - After 捕获 panic
- [ ] **新增** ProfilingMiddleware (性能分析)
- [ ] **新增** CachingMiddleware (结果缓存)
- [ ] **新增** AuditMiddleware (审计日志)
- [ ] 实现预设链 (`DefaultMiddlewareChain`, `DebugMiddlewareChain`, `ProductionMiddlewareChain`)

**验收标准**:
```bash
go test ./runtime/action/... -run TestMiddleware
# 所有中间件测试通过
```

---

### 阶段 4: 输入处理系统迁移（约 3 小时）

**目标**: 迁移 InputProcessor 和 KeyMap

#### 4.1 迁移 `processor.go`

**文件**: `runtime/action/processor.go`

**迁移自**: `framework/action/processor.go`

**任务**:
- [ ] 定义 `InputProcessor` 结构体
- [ ] 实现 `NewInputProcessor`
- [ ] 实现 `SetKeyMap`, `GetKeyMap`
- [ ] 实现 `ProcessMsg` (处理 KeyMsg/MouseMsg)
- [ ] 实现 `processKeyMsg` (优先 KeyMap，然后默认规则)
- [ ] 实现 `processMouseMsg` (转换鼠标事件)
- [ ] 实现 `applyDefaultKeyMapping` (默认转换规则)
- [ ] 移除对 `framework/action` 的依赖
- [ ] 使用新的 Action 结构体

**验收标准**:
```bash
go test ./runtime/action/... -run TestInputProcessor
# 测试键盘和鼠标转换
```

#### 4.2 迁移 `keymap.go`（重命名为 `keymap.go`）

**文件**: `runtime/action/keymap.go`

**迁移自**: `framework/action/keymap.go`

**任务**:
- [ ] 定义 `KeyMap` 结构体
- [ ] 定义 `KeySignature` 结构体
- [ ] 定义 `Modifier` 常量
- [ ] 实现 `NewKeyMap`
- [ ] 实现 `Bind` (全局绑定)
- [ ] 实现 `BindWithContext` (上下文绑定)
- [ ] 实现 `LookupKeyMsg` (从 KeyMsg 查找)
- [ ] 实现 `Lookup` (从 keySpec 查找)
- [ ] 实现上下文管理 (`PushContext`, `PopContext`, `SetCurrentContext`)
- [ ] 实现 `Unbind`, `UnbindWithContext`
- [ ] 实现 `Clear`, `Size`
- [ ] 实现 `Dump` (调试输出)
- [ ] 实现 `DefaultKeyMap` (预设映射)
- [ ] 实现按键解析 (`parseKeySpec`)
- [ ] 实现特殊键判断 (`isSpecialKey`)

**验收标准**:
```bash
go test ./runtime/action/... -run TestKeyMap
# 测试按键绑定和查找
```

---

### 阶段 5: 高级功能迁移（约 2 小时）

**目标**: 迁移作用域分发和保留复合 Action

#### 5.1 迁移 `scope.go`（作用域分发）

**文件**: `runtime/action/scope.go`

**迁移自**: `framework/action/scope_dispatcher.go`

**任务**:
- [ ] 定义 `ScopeDispatcher` 结构体
- [ ] 定义 `ScopeActionHandler` 类型
- [ ] 实现 `NewScopeDispatcher`, `NewScopeDispatcherWithName`
- [ ] 实现 `Register`, `Unregister`
- [ ] 实现 `Dispatch` (支持冒泡)
- [ ] 实现 `DispatchByID`
- [ ] 实现父子关系 (`GetParent`, `SetParent`)
- [ ] 实现全局作用域管理 (`SetCurrentScopeDispatcher`, `GetCurrentScopeDispatcher`, `WithScopeDispatcher`)
- [ ] 实现闭包注册 (`RegisterScopeClosure`, `RegisterScopeClosureWithAction`, `RegisterScopeClosureToDispatcher`)
- [ ] 实现 `GenerateScopeActionID`, `GenerateScopeActionIDWithPrefix`
- [ ] 更新 `String` 方法

**验收标准**:
```bash
go test ./runtime/action/... -run TestScopeDispatcher
# 测试作用域隔离和冒泡
```

#### 5.2 验证 `composite.go`（复合 Action）

**文件**: `runtime/action/composite.go`（已存在）

**任务**:
- [ ] 检查现有实现完整
- [ ] 确认兼容新的 `Action` 结构体
- [ ] 确认 `Execute` 方法返回 `ActionResult`
- [ ] 确认 `Batch`, `Sequence` 等便捷函数
- [ ] 确认 `WorkerPool` 实现
- [ ] 确认 `RetryAction`, `TimeoutAction`, `FallbackAction` 实现
- [ ] 如有冲突，更新以使用新的 Action

**验收标准**:
```bash
go test ./runtime/action/... -run TestComposite
# 测试并发和顺序执行
```

#### 5.3 验证 `workerpool.go`

**文件**: `runtime/action/workerpool.go`（可能在 composite.go 中）

**任务**:
- [ ] 确认 `WorkerPool` 实现
- [ ] 确认 `NewWorkerPool`, `Start`, `Stop`
- [ ] 确认 `Submit`, `SubmitWithTimeout`
- [ ] 确认 `worker` 协程逻辑

**验收标准**:
```bash
go test ./runtime/action/... -run TestWorkerPool
# 测试并发限制
```

---

### 阶段 6: 错误处理（约 1 小时）

**目标**: 确认并增强错误处理

#### 6.1 验证 `errors.go`

**文件**: `runtime/action/errors.go`（已存在）

**任务**:
- [ ] 检查 `ErrorType` 常量
- [ ] 检查 `Error` 结构体
- [ ] 检查预定义错误构造器
- [ ] 确认兼容新的 `Action` 结构体
- [ ] 确认 Payload 验证函数

**验收标准**:
```bash
go test ./runtime/action/... -run TestErrors
# 所有错误测试通过
```

---

### 阶段 7: 适配器层（约 2 小时）

**目标**: 创建向后兼容的适配器

#### 7.1 创建 `adapters.go`

**文件**: `runtime/action/adapters.go`

**任务**:
- [ ] 实现 `TargetIDConverter` (uint64 ↔ string)
  - `StringToTargetID(source string) uint64`
  - `TargetIDToString(id uint64) string`
- [ ] 实现 `MsgToActionConverter` (Msg → Action)
  - `ConvertKeyMsgToAction(msg *KeyMsg) *Action`
  - `ConvertMouseMsgToAction(msg *MouseMsg) *Action`
- [ ] 实现 `LegacyActionTargetAdapter` (适配旧的 `framework/action.ActionTarget`)
  - `AdaptActionTarget(old ActionTarget) Target`
- [ ] 实现 `LegacyRouterAdapter` (适配旧的 `framework/action.Router`)
- [ ] 实现 `LegacyDispatcherAdapter` (适配旧的 `framework/action` Dispatcher)
- [ ] 提供迁移指南注释

**验收标准**:
```bash
go test ./runtime/action/... -run TestAdapters
# 适配器测试通过
```

#### 7.2 创建 `helpers.go`

**文件**: `runtime/action/helpers.go`

**任务**:
- [ ] 实现 Action 分类方法 (`IsNavigation`, `IsEditing`, `IsMouse`,...)
- [ ] 实现 Payload 辅助方法 (`GetPayloadString`, `GetPayloadInt`, `GetPayloadPoint`,...)
- [ ] 实现 Meta 操作 (`GetMeta`, `SetMeta`)
- [ ] 实现调试工具 (`DumpAction`, `FormatAction`)
- [ ] 实现统计工具 (`CountActionsByType`)

**验收标准**:
```bash
go test ./runtime/action/... -run TestHelpers
# 辅助函数测试通过
```

---

### 阶段 8: 测试和文档（约 2 小时）

**目标**: 确保完整测试和文档

#### 8.1 编写集成测试

**文件**: `runtime/action/integration_test.go`

**任务**:
- [ ] 测试完整流程: Msg → Processor → Action → Router → Target
- [ ] 测试三阶段路由 (Capture → Target → Bubble)
- [ ] 测试中间件链 (Before/After)
- [ ] 测试 KeyMap 查找
- [ ] 测试 ScopeDispatcher 冒泡
- [ ] 测试 Composite Action 并发
- [ ] 测试 WorkerPool 限制
- [ ] 测试 TargetID 转换

#### 8.2 更新文档

**文件**: `runtime/action/README.md`

**任务**:
- [ ] 更新架构图
- [ ] 更新 API 文档
- [ ] 更新使用示例
- [ ] 添加迁移指南 (`MIGRATION.md`)
- [ ] 添加 CHANGELOG 记录变更

---

### 阶段 9: 逐步迁移使用方（约 4 小时）

**目标**: 更新所有导入 `framework/action` 的文件

#### 9.1 迁移策略

**步骤**: 按依赖深度从低到高迁移

| 层级 | 目录 | 优先级 | 预计时间 |
|------|------|--------|----------|
| 1 | `runtime/action` 测试文件 | P0 | 0.5h |
| 2 | `framework/action` 测试文件 | P0 | 0.5h |
| 3 | `components/*` 基础组件 | P1 | 1h |
| 4 | `components/form/*` 表单组件 | P1 | 1h |
| 5 | `ui/components/*` UI 组件 | P1 | 0.5h |
| 6 | `framework/*` 框架内部 | P2 | 0.5h |

#### 9.2 批量替换脚本

创建 `scripts/replace_action_imports.sh`:

```bash
#!/bin/bash
# 批量替换 framework/action 为 runtime/action

find . -name "*.go" -type f -exec sed -i 's|github.com/wwsheng009/mint/framework/action|github.com/wwsheng009/mint/runtime/action|g' {} \;
```

**任务**:
- [ ] 创建替换脚本
- [ ] 执行脚本
- [ ] 检查替换结果
- [ ] 手动修复特殊情况
- [ ] 运行 `go mod tidy`

#### 9.3 逐层验证

**任务**:
- [ ] 迁移 `components/button/*`，测试通过
- [ ] 迁移 `components/input/*`，测试通过
- [ ] 迁移 `components/textarea/*`，测试通过
- [ ] 迁移 `components/select/*`，测试通过
- [ ] 迁移 `components/treeview/*`，测试通过
- [ ] 迁移所有其他组件
- [ ] 迁移 `framework/*` 内部文件

---

### 阶段 10: 清理和验证（约 2 小时）

**目标**: 删除旧代码并最终验证

#### 10.1 删除 `framework/action`

**任务**:
```bash
# 确认没有导入 framework/action
grep -r "framework/action" . --include="*.go"  # 应该返回空

# 备份（可选）
# git mv framework/action framework/action.backup

# 删除目录
rm -rf framework/action
```

#### 10.2 最终验证

**任务**:
- [ ] 运行所有测试: `go test ./...`
- [ ] 检查编译错误: `go build ./...`
- [ ] 运行静态检查: `go vet ./...`
- [ ] 运行 linter: `golangci-lint run ./...`
- [ ] 手动测试示例应用

#### 10.3 提交变更

```bash
git add .
git commit -m "Merge framework/action into runtime/action

- Unified Action struct with string Target and optional uint64 TargetID
- Merged Router, InputProcessor, KeyMap from framework
- Merged Middleware system (8 built-in middleware)
- Retained Composite Action, WorkerPool from runtime
- Added backward compatibility adapters
- Deleted framework/action directory

Breaking changes: All imports from framework/action must use runtime/action"
```

---

## 四、详细 Checklist

### 4.1 准备阶段 Checklist

| ID | 任务 | 负责人 | 状态 | 备注 |
|----|------|--------|------|------|
| P-01 | 创建分支 `merge-action-systems` | | ☐ | |
| P-02 | 运行所有测试确保基线 | | ☐ | 记录结果 |
| P-03 | 列出 framework/action 所有文件 | | ☐ | 13 个文件 |
| P-04 | 列出 runtime/action 所有文件 | | ☐ | 10 个文件 |
| P-05 | 统计导入 framework/action 的文件 | | ☐ | 131 个 |
| P-06 | 统计导入 runtime/action 的文件 | | ☐ | 63 个 |
| P-07 | 记录所有测试文件位置 | | ☐ | |

### 4.2 阶段 1: 核心结构 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 1-01 | 定义 Action 结构体 | action.go | ☐ |
| 1-02 | 定义 ActionType 常量（合并） | action.go | ☐ |
| 1-03 | 实现 NewAction 函数 | action.go | ☐ |
| 1-04 | 实现 WithTarget, WithPayload, WithSource | action.go | ☐ |
| 1-05 | 实现 Clone 方法 | action.go | ☐ |
| 1-06 | 实现 StopPropagation | action.go | ☐ |
| 1-07 | 实现 String 方法 | action.go | ☐ |
| 1-08 | 实现 TargetID ↔ String 转换 | action.go | ☐ |
| 1-09 | 实现 Payload 辅助方法 | action.go | ☐ |
| 1-10 | 编写 action.go 测试 | action_test.go | ☐ |
| 1-11 | 定义 Target 接口 | target.go | ☐ |
| 1-12 | 定义能力接口 (Focusable 等) | target.go | ☐ |
| 1-13 | 实现 TargetFunc | target.go | ☐ |
| 1-14 | 实现 TargetChain | target.go | ☐ |
| 1-15 | 实现 BaseActionTarget | target.go | ☐ |
| 1-16 | 实现 CompositeActionTarget | target.go | ☐ |
| 1-17 | 编写 target.go 测试 | target_test.go | ☐ |
| 1-18 | 定义 Payload 类型 | payload.go | ☐ |
| 1-19 | 实现 New* 构造函数 | payload.go | ☐ |
| 1-20 | 实现 AsXxxPayload 方法 | payload.go | ☐ |
| 1-21 | 编写 payload.go 测试 | payload_test.go | ☐ |

### 4.3 阶段 2: 路由系统 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 2-01 | 定义 RouterResult, ActionPhase | router.go | ☐ |
| 2-02 | 定义接口 (CaptureActionHandler 等) | router.go | ☐ |
| 2-03 | 定义 Router 结构体 | router.go | ☐ |
| 2-04 | 实现 NewRouter | router.go | ☐ |
| 2-05 | 实现 SetMiddleware, AddMiddleware | router.go | ☐ |
| 2-06 | 实现 AddGlobalHandler | router.go | ☐ |
| 2-07 | 实现 AddCaptureHandler | router.go | ☐ |
| 2-08 | 实现 AddBubbleHandler | router.go | ☐ |
| 2-09 | 实现 RegisterTarget, UnregisterTarget | router.go | ☐ |
| 2-10 | 实现 Dispatch (三阶段) | router.go | ☐ |
| 2-11 | 实现 capturePhase | router.go | ☐ |
| 2-12 | 实现 targetPhase | router.go | ☐ |
| 2-13 | 实现 bubblePhase | router.go | ☐ |
| 2-14 | 实现 BuildTargetRegistry | router.go | ☐ |
| 2-15 | 更新 SimpleDispatcher | dispatcher.go | ☐ |
| 2-16 | 实现 TargetID 转换 | dispatcher.go | ☐ |
| 2-17 | 编写 router 测试 | router_test.go | ☐ |
| 2-18 | 编写 dispatcher 测试 | dispatcher_test.go | ☐ |

### 4.4 阶段 3: 中间件 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 3-01 | 定义 ActionMiddleware 接口 | middleware.go | ☐ |
| 3-02 | 定义 MiddlewareChain | middleware.go | ☐ |
| 3-03 | 实现 LoggingMiddleware | middleware.go | ☐ |
| 3-04 | 实现 ThrottleMiddleware | middleware.go | ☐ |
| 3-05 | 实现 ValidationMiddleware | middleware.go | ☐ |
| 3-06 | 实现 MetricsMiddleware | middleware.go | ☐ |
| 3-07 | 实现 RecoveryMiddleware | middleware.go | ☐ |
| 3-08 | 实现 ProfilingMiddleware (新增) | middleware.go | ☐ |
| 3-09 | 实现 CachingMiddleware (新增) | middleware.go | ☐ |
| 3-10 | 实现 AuditMiddleware (新增) | middleware.go | ☐ |
| 3-11 | 实现预设链 | middleware.go | ☐ |
| 3-12 | 编写 middleware 测试 | middleware_test.go | ☐ |

### 4.5 阶段 4: 输入处理 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 4-01 | 定义 InputProcessor | processor.go | ☐ |
| 4-02 | 实现 ProcessMsg (KeyMsg/MouseMsg) | processor.go | ☐ |
| 4-03 | 实现 processKeyMsg | processor.go | ☐ |
| 4-04 | 实现 processMouseMsg | processor.go | ☐ |
| 4-05 | 实现 applyDefaultKeyMapping | processor.go | ☐ |
| 4-06 | 定义 KeyMap | keymap.go | ☐ |
| 4-07 | 实现 Bind, BindWithContext | keymap.go | ☐ |
| 4-08 | 实现 LookupKeyMsg | keymap.go | ☐ |
| 4-09 | 实现上下文管理 | keymap.go | ☐ |
| 4-10 | 实现 DefaultKeyMap | keymap.go | ☐ |
| 4-11 | 编写 processor 测试 | processor_test.go | ☐ |
| 4-12 | 编写 keymap 测试 | keymap_test.go | ☐ |

### 4.6 阶段 5: 高级功能 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 5-01 | 定义 ScopeDispatcher | scope.go | ☐ |
| 5-02 | 实现 Register, Unregister | scope.go | ☐ |
| 5-03 | 实现 Dispatch (冒泡) | scope.go | ☐ |
| 5-04 | 实现全局作用域管理 | scope.go | ☐ |
| 5-05 | 实现闭包注册 | scope.go | ☐ |
| 5-06 | 编写 scope 测试 | scope_test.go | ☐ |
| 5-07 | 验证 composite.go 兼容性 | composite.go | ☐ |
| 5-08 | 验证 WorkerPool | composite.go | ☐ |
| 5-09 | 编写 composite 测试 | composite_test.go | ☐ |

### 4.7 阶段 6: 错误处理 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 6-01 | 验证 ErrorType 常量 | errors.go | ☐ |
| 6-02 | 验证 Error 结构体 | errors.go | ☐ |
| 6-03 | 验证预定义错误 | errors.go | ☐ |
| 6-04 | 验证 Payload 验证函数 | errors.go | ☐ |
| 6-05 | 编写 errors 测试 | errors_test.go | ☐ |

### 4.8 阶段 7: 适配器 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 7-01 | 实现 TargetIDConverter | adapters.go | ☐ |
| 7-02 | 实现 MsgToActionConverter | adapters.go | ☐ |
| 7-03 | 实现 LegacyActionTargetAdapter | adapters.go | ☐ |
| 7-04 | 实现 LegacyRouterAdapter | adapters.go | ☐ |
| 7-05 | 实现 LegacyDispatcherAdapter | adapters.go | ☐ |
| 7-06 | 编写 adapters 测试 | adapters_test.go | ☐ |
| 7-07 | 实现辅助函数 | helpers.go | ☐ |
| 7-08 | 编写 helpers 测试 | helpers_test.go | ☐ |

### 4.9 阶段 8: 测试和文档 Checklist

| ID | 任务 | 文件 | 状态 |
|----|------|------|------|
| 8-01 | 编写完整流程测试 | integration_test.go | ☐ |
| 8-02 | 编写三阶段路由测试 | integration_test.go | ☐ |
| 8-03 | 编写中间件链测试 | integration_test.go | ☐ |
| 8-04 | 编写 KeyMap 测试 | integration_test.go | ☐ |
| 8-05 | 编写 ScopeDispatcher 测试 | integration_test.go | ☐ |
| 8-06 | 编写 Composite 测试 | integration_test.go | ☐ |
| 8-07 | 编写 WorkerPool 测试 | integration_test.go | ☐ |
| 8-08 | 编写 TargetID 转换测试 | integration_test.go | ☐ |
| 8-09 | 更新 README 架构图 | README.md | ☐ |
| 8-10 | 更新 API 文档 | README.md | ☐ |
| 8-11 | 更新使用示例 | README.md | ☐ |
| 8-12 | 创建 MIGRATION.md | MIGRATION.md | ☐ |
| 8-13 | 更新 CHANGELOG | CHANGELOG.md | ☐ |

### 4.10 阶段 9-10: 迁移和清理 Checklist

| ID | 任务 | 状态 |
|----|------|------|
| 9-01 | 创建批量替换脚本 | ☐ |
| 9-02 | 执行脚本并检查结果 | ☐ |
| 9-03 | 运行 go mod tidy | ☐ |
| 9-04 | 迁移 components/button | ☐ |
| 9-05 | 迁移 components/input | ☐ |
| 9-06 | 迁移 components/textarea | ☐ |
| 9-07 | 迁移 components/select | ☐ |
| 9-08 | 迁移 components/treeview | ☐ |
| 9-09 | 迁移 components/form/* | ☐ |
| 9-10 | 迁移 ui/components/* | ☐ |
| 9-11 | 迁移 framework/* | ☐ |
| 9-12 | 验证所有测试通过 | ☐ |
| 10-01 | 确认没有 framework/action 导入 | ☐ |
| 10-02 | 运行 go test ./... | ☐ |
| 10-03 | 运行 go build ./... | ☐ |
| 10-04 | 运行 go vet ./... | ☐ |
| 10-05 | 运行 gofmt / golangci-lint | ☐ |
| 10-06 | 删除 framework/action 目录 | ☐ |
| 10-07 | 提交变更到 Git | ☐ |

---

## 五、风险评估和回退策略

### 5.1 风险识别

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| TargetID 转换导致性能问题 | 高 | 中 | 缓存转换结果，性能测试 |
| 旧代码无法编译破坏 CI | 高 | 低 | 分阶段迁移，保留适配器 |
| 测试覆盖率不足导致回归 | 高 | 中 | 每个阶段都写完整测试 |
| 中间件顺序导致错误 | 中 | 低 | 详尽的中间件集成测试 |
| 作用域分发逻辑错误 | 中 | 中 | 专门的 ScopeDispatcher 测试 |
| 内存泄漏（未清理 Meta） | 中 | 低 | Runtime 检测工具 |

### 5.2 回退策略

#### 回退触发条件

任何以下情况发生都应考虑回退：

1. **阶段 1-8 期间**: 任何测试失败且无法在 2 小时内修复
2. **阶段 9 期间**: 超过 10% 的组件迁移后无法测试通过
3. **阶段 10 之前**: 任何性能回归超过 20%
4. **阶段 10 期间**: CI/CD 流水线失败

#### 回退步骤

```bash
# 如果在阶段 1-8 期间需要回退
git reset --hard HEAD~1  # 回退当前阶段的提交
git checkout merge-action-systems  # 重新开始

# 如果在阶段 9 之后需要回退
git revert --no-commit
git reset --hard refs/original/refs/heads/merge-action-systems

# 恢复 framework/action（如果已删除）
git checkout HEAD~1 -- framework/action
```

### 5.3 性能基准

迁移前后对比指标：

| 指标 | 迁移前基线 | 迁移后目标 | 容差 |
|------|-----------|-----------|------|
| Action 创建开销 | ~100ns | ≤150ns | +50% |
| Dispatch 开销 | ~500ns | ≤600ns | +20% |
| TargetID 转换 | N/A | ≤50ns | - |
| 内存占用 | 1MB | ≤1.2MB | +20% |
| 全量单测耗时 | 30s | ≤36s | +20% |

---

## 六、时间估算

| 阶段 | 任务 | 预计时间 | 实际时间 |
|------|------|----------|----------|
| 0 | 准备阶段 | 2h | \_\_\_\_h |
| 1 | 核心结构合并 | 3h | \_\_\_\_h |
| 2 | 路由系统迁移 | 4h | \_\_\_\_h |
| 3 | 中间件系统迁移 | 3h | \_\_\_\_h |
| 4 | 输入处理迁移 | 3h | \_\_\_\_h |
| 5 | 高级功能迁移 | 2h | \_\_\_\_h |
| 6 | 错误处理 | 1h | \_\_\_\_h |
| 7 | 适配器层 | 2h | \_\_\_\_h |
| 8 | 测试和文档 | 2h | \_\_\_\_h |
| 9 | 迁移使用方 | 4h | \_\_\_\_h |
| 10 | 清理和验证 | 2h | \_\_\_\_h |
| **总计** | | **28h** | **\_\_\_\_**h |

---

## 七、验收标准

### 7.1 功能验收

- [ ] 所有 131 个导入 `framework/action` 的文件成功迁移到 `runtime/action`
- [ ] 所有现有测试通过（无回归）
- [ ] 新增测试覆盖率 ≥ 90%
- [ ] 所有中间件功能正常工作
- [ ] 三阶段路由传播正确
- [ ] ScopeDispatcher 作用域隔离正确

### 7.2 性能验收

- [ ] Action 创建开销 ≤ 150ns
- [ ] Dispatch 开销 ≤ 600ns
- [ ] TargetID 转换 ≤ 50ns
- [ ] 全量单测耗时 ≤ 36s

### 7.3 代码质量验收

- [ ] 无 `go vet` 警告
- [ ] 无 `golangci-lint` 警告
- [ ] 所有代码通过 `gofmt`
- [ ] 代码覆盖率 ≥ 90%

### 7.4 文档验收

- [ ] README.md 更新完整
- [ ] MIGRATION.md 提供清晰指南
- [ ] CHANGELOG.md 记录所有破坏性变更
- [ ] 所有新 API 有 Go Doc 注释

---

## 八、后续改进建议

### 8.1 短期优化（1-2 周）

1. **TargetID 缓存优化**: 缓存 string → uint64 转换结果
2. **Action 对象池**: 减少内存分配
3. **中间件性能分析**: 识别瓶颈中间件

### 8.2 中期优化（1-2 月）

1. **异步 Action**: 支持异步处理和回调
2. **Action 序列化**: 支持持久化和回放
3. **分布式分发**: 支持跨进程/跨机器 Action 分发

### 8.3 长期方向（3+ 月）

1. **Action DSL**: 定义领域特定的 Action 流程语言
2. **Action 可视化**: 调试工具可视化 Action 流转
3. **Action Profiling**: 集成性能分析工具

---

## 九、附录

### 9.1 相关文档

- [原始分析报告](./MERGE_ANALYSIS.md)
- [Action 架构设计](../action/README.md)
- [迁移指南](./MIGRATION_TEMPLATES.md)

### 9.2 命令参考

#### 测试命令

```bash
# 单元测试
go test ./runtime/action/... -v

# 基准测试
go test ./runtime/action/... -bench=. -benchmem

# 集成测试
go test ./runtime/action/... -run TestIntegration

# 覆盖率
go test ./runtime/action/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### 代码检查命令

```bash
go vet ./runtime/action/...
gofmt -s -w runtime/action/
golangci-lint run runtime/action/
```

### 9.3 人员分配

| 角色 | 姓名/团队 | 职责 |
|------|----------|------|
| 主程 | | 架构设计、核心代码实现 |
| 测试 | | 测试编写、问题验证 |
| DevOps | | CI/CD 流水线更新 |
| 文档 | | 文档更新、迁移指南 |

---

**文档版本**: v1.0
**最后更新**: 2026年2月25日
**状态**: 待批准
