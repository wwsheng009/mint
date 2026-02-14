# Phase 0: 架构设计与接口定义

## 1. 核心类型定义

### 1.1 Action 结构（已有，需增强）

```go
// framework/action/action.go

type Action struct {
    Type     ActionType      // 语义化操作类型
    Payload  interface{}     // 操作数据
    Source   string          // 触发源：keyboard, mouse, system, custom
    TargetID uint64          // 目标节点 ID

    // ===== 新增字段 =====
    Timestamp time.Time      // 创建时间
    ID        uint64         // 唯一标识符（用于追踪和调试）

    // 传播控制
    stopped   bool           // 内部：是否停止传播

    // 元数据
    Meta      map[string]interface{} // 扩展元数据
}

// 新增方法
func (a *Action) StopPropagation()    { a.stopped = true }
func (a *Action) IsStopped() bool     { return a.stopped }
func (a *Action) Clone() *Action      { ... }
```

### 1.2 ActionType 扩展（已有，补充）

```go
// framework/action/action.go

const (
    // ... 现有类型 ...

    // ===== 新增：系统级 Action =====
    ActionInit       ActionType = "init"        // 组件初始化
    ActionMount      ActionType = "mount"       // 组件挂载
    ActionUnmount    ActionType = "unmount"     // 组件卸载
    ActionResize     ActionType = "resize"      // 窗口调整（Payload: Size{W, H}）

    // ===== 新增：焦点 Action =====
    ActionFocusGained  ActionType = "focus_gained"   // 获得焦点
    ActionFocusLost    ActionType = "focus_lost"     // 失去焦点
    ActionFocusNext    ActionType = "focus_next"     // 焦点后移
    ActionFocusPrev    ActionType = "focus_prev"     // 焦点前移

    // ===== 新增：数据 Action =====
    ActionDataLoad   ActionType = "data_load"   // 数据加载
    ActionDataUpdate ActionType = "data_update" // 数据更新
    ActionDataError  ActionType = "data_error"  // 数据错误

    // ===== 新增：撤销/重做 =====
    ActionUndo ActionType = "undo"              // 撤销
    ActionRedo ActionType = "redo"              // 重做
)
```

### 1.3 ActionTarget 接口（已有，保持不变）

```go
// framework/action/actiontarget.go

type ActionTarget interface {
    // HandleAction 处理 Action
    // 返回 true 表示已处理，停止传播
    // 返回 false 表示未处理，继续传播
    HandleAction(action *Action) bool

    // GetSupportedActions 返回支持的 Action 类型
    GetSupportedActions() []ActionType

    // CanHandleAction 预检查（不修改状态）
    CanHandleAction(action *Action) bool
}
```

### 1.4 ActionRouter 增强

```go
// framework/action/router.go

type Router struct {
    Root            *runtime.LayoutNode

    // 捕获阶段处理器（按优先级排序）
    CaptureHandlers []*CaptureHandlerEntry

    // 冒泡阶段处理器
    BubbleHandlers []*BubbleHandlerEntry

    // 目标处理器映射（按 targetID 索引）
    TargetHandlers map[uint64]*TargetHandlerEntry

    // ===== 新增 =====

    // 全局处理器（无目标 Action）
    GlobalHandlers []GlobalActionHandler

    // Action 中间件链
    Middleware []ActionMiddleware

    // Action 对象池
    actionPool sync.Pool
}

// 全局处理器接口
type GlobalActionHandler interface {
    HandleGlobalAction(action *Action) bool
    Priority() int
}

// 中间件接口
type ActionMiddleware interface {
    // Before 在 Action 分发前调用
    // 返回修改后的 Action 或 nil（拦截）
    Before(action *Action) *Action

    // After 在 Action 分发后调用
    After(action *Action, handled bool)
}
```

### 1.5 InputProcessor 增强

```go
// framework/action/processor.go

type InputProcessor struct {
    keyMap *KeyMap

    // ===== 新增 =====

    // 自定义转换器
    customConverters []MsgConverter

    // 默认目标（用于无目标的键盘事件）
    defaultTarget uint64
}

// MsgConverter 自定义消息转换器接口
type MsgConverter interface {
    CanConvert(msg runtimemsg.Msg) bool
    Convert(msg runtimemsg.Msg) *Action
}

// 新增方法
func (p *InputProcessor) AddConverter(converter MsgConverter)
func (p *InputProcessor) SetDefaultTarget(targetID uint64)
func (p *InputProcessor) ProcessMsgWithTarget(msg runtimemsg.Msg, targetID uint64) *Action
```

## 2. 适配器定义

### 2.1 Updater → ActionTarget 适配器

```go
// framework/action/adapters.go

// UpdaterAdapter 将旧 Updater 接口适配为 ActionTarget
type UpdaterAdapter struct {
    updater component.Updater
    id      uint64
}

func NewUpdaterAdapter(updater component.Updater, id uint64) *UpdaterAdapter {
    return &UpdaterAdapter{updater: updater, id: id}
}

func (a *UpdaterAdapter) HandleAction(action *Action) bool {
    // 将 Action 转换回 Msg（兼容旧代码）
    msg := ActionToMsg(action)
    if msg == nil {
        return false
    }

    cmd := a.updater.Update(msg)
    // TODO: 执行 Cmd
    return cmd != nil
}

func (a *UpdaterAdapter) GetSupportedActions() []ActionType {
    // 根据 Updater 的能力推断支持的 Action
    return []ActionType{
        ActionClick, ActionInputText, ActionNavigateUp, ActionNavigateDown,
    }
}

func (a *UpdaterAdapter) CanHandleAction(action *Action) bool {
    return true // 保守估计
}
```

### 2.2 EventHandler → ActionTarget 适配器

```go
// EventHandlerAdapter 将旧 EventHandler 接口适配为 ActionTarget
type EventHandlerAdapter struct {
    handler frameworkevent.EventHandler
    id      uint64
}

func NewEventHandlerAdapter(handler frameworkevent.EventHandler, id uint64) *EventHandlerAdapter {
    return &EventHandlerAdapter{handler: handler, id: id}
}

func (a *EventHandlerAdapter) HandleAction(action *Action) bool {
    // 将 Action 转换为 Event
    event := ActionToEvent(action)
    if event == nil {
        return false
    }

    return a.handler.HandleEvent(event)
}

func (a *EventHandlerAdapter) GetSupportedActions() []ActionType {
    return []ActionType{ActionClick, ActionKeyPress}
}

func (a *EventHandlerAdapter) CanHandleAction(action *Action) bool {
    return true
}
```

## 3. App 集成设计

### 3.1 App 结构变更

```go
// framework/app.go

type App struct {
    // ... 现有字段 ...

    // ===== 新增：Action 系统 =====
    actionRouter    *action.Router
    inputProcessor  *action.InputProcessor
    actionTargets   map[uint64]action.ActionTarget  // 注册的 ActionTarget

    // ===== 废弃字段（保留兼容） =====
    // router       *frameworkevent.Router      // 废弃，用 actionRouter 替代
    // keyMap       *frameworkevent.KeyMap      // 废弃，合并到 InputProcessor
    // pump         *frameworkevent.Pump        // 保留，但输出 Action
}
```

### 3.2 主循环变更

```go
// framework/app.go

func (a *App) Run() error {
    // ... 初始化 ...

    for a.state == StateRunning {
        select {
        case msg := <-eventChan:
            // ===== 新的 Action 统一路径 =====
            a.processMsg(msg)

        case <-ticker.C:
            a.handleTick()
            if a.dirty && a.throttler.ShouldRender() {
                a.render()
            }

        // ... 其他 case ...
        }
    }
    return nil
}

// processMsg 统一的消息处理入口
func (a *App) processMsg(msg runtimemsg.Msg) {
    // 1. 转换为 Action
    act := a.inputProcessor.ProcessMsg(msg)
    if act == nil {
        // 无法转换的消息，尝试旧路径（兼容）
        a.handleEvent(frameworkevent.MsgToEvent(msg))
        return
    }

    // 2. 应用中间件
    for _, mw := range a.actionRouter.Middleware {
        act = mw.Before(act)
        if act == nil {
            return // 被中间件拦截
        }
    }

    // 3. 分发 Action
    result := a.actionRouter.Dispatch(act)

    // 4. 后处理
    for _, mw := range a.actionRouter.Middleware {
        mw.After(act, result.Handled)
    }

    // 5. 标记脏
    if result.Handled {
        a.dirty = true
    }
}
```

## 4. 文件结构

```
framework/
├── action/
│   ├── action.go           # Action 定义（增强）
│   ├── actiontarget.go     # ActionTarget 接口（保持）
│   ├── router.go           # ActionRouter（增强）
│   ├── processor.go        # InputProcessor（增强）
│   ├── keymap.go           # KeyMap（保持）
│   ├── adapters.go         # 新增：旧接口适配器
│   ├── converters.go       # 新增：Msg→Action 转换器
│   ├── middleware.go       # 新增：中间件
│   └── pool.go             # 新增：Action 对象池
├── app.go                  # App 主逻辑（修改）
└── component/
    └── component.go        # 组件定义（修改接口）

# 废弃/删除
framework/event/
├── handler.go              # 废弃 Router
└── msg_adapter.go          # 废弃 MsgToEvent
```

## 5. 接口兼容矩阵

| 旧接口 | 新接口 | 适配器 |
|--------|--------|--------|
| `Updater.Update(Msg) Cmd` | `ActionTarget.HandleAction(*Action) bool` | `UpdaterAdapter` |
| `EventHandler.HandleEvent(Event) bool` | `ActionTarget.HandleAction(*Action) bool` | `EventHandlerAdapter` |
| `Component.HandleEvent(Event) bool` | `ActionTarget.HandleAction(*Action) bool` | `EventHandlerAdapter` |
| `MsgHandler.Handle(interface{}) interface{}` | `ActionTarget.HandleAction(*Action) bool` | 直接迁移 |

## 6. 迁移检查清单

### Phase 0 完成标准

- [ ] Action 结构增加新字段
- [ ] ActionType 补充新类型
- [ ] ActionRouter 增加中间件支持
- [ ] InputProcessor 增加扩展点
- [ ] 编写适配器实现
- [ ] 编写单元测试
- [ ] 更新文档
