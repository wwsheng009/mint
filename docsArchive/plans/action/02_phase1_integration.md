# Phase 1: 核心系统集成

## 目标

将 Action 系统集成到 App 主循环中，建立统一的 Action 传播路径。

## 前置条件

- Phase 0 完成架构设计和接口定义
- 适配器已实现并测试通过

## 1. App 结构修改

### 1.1 新增字段

```go
// framework/app.go

type App struct {
    // ===== 现有字段（保留） =====
    root           component.Node
    instanceRoot   *instance.Instance
    pump           *frameworkevent.Pump
    focusManager   *rtui.VNodeFocusManager
    hitMap         *runtimeevent.HitMap
    // ...

    // ===== 新增：Action 系统 =====
    actionRouter   *action.Router           // Action 分发器
    inputProcessor *action.InputProcessor   // Msg → Action 转换器
    actionRegistry *ActionRegistry          // ActionTarget 注册表

    // ===== 过渡期兼容字段 =====
    legacyMode     bool                     // 是否启用兼容模式
    legacyRouter   *frameworkevent.Router   // 旧 Router（过渡期保留）
}

// ActionRegistry 管理 ActionTarget 的注册
type ActionRegistry struct {
    targets map[uint64]action.ActionTarget
    mu      sync.RWMutex
}

func NewActionRegistry() *ActionRegistry {
    return &ActionRegistry{
        targets: make(map[uint64]action.ActionTarget),
    }
}

func (r *ActionRegistry) Register(id uint64, target action.ActionTarget) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.targets[id] = target
}

func (r *ActionRegistry) Unregister(id uint64) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.targets, id)
}

func (r *ActionRegistry) Get(id uint64) action.ActionTarget {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.targets[id]
}

func (r *ActionRegistry) GetAll() map[uint64]action.ActionTarget {
    r.mu.RLock()
    defer r.mu.RUnlock()
    result := make(map[uint64]action.ActionTarget, len(r.targets))
    for k, v := range r.targets {
        result[k] = v
    }
    return result
}
```

### 1.2 初始化修改

```go
// framework/app.go

func NewApp() *App {
    app := &App{
        // ===== 现有初始化 =====
        focusManager: rtui.NewVNodeFocusManager(),
        quit:         make(chan struct{}, 1),
        tickInterval: 16 * time.Millisecond,
        firstRender:  true,
        throttler:    render.NewThrottler(60),
        contextMgr:   core.NewContextManager(context.Background()),
        userData:     make(map[string]interface{}),
        renderer:     paint.NewRenderer(80, 24),

        // ===== 新增：Action 系统初始化 =====
        actionRouter:   action.NewRouter(nil), // 根节点稍后设置
        inputProcessor: action.NewInputProcessor(),
        actionRegistry: NewActionRegistry(),
        legacyMode:     true, // 默认启用兼容模式
    }

    // 设置 KeyMap
    keyMap := action.NewKeyMap()
    app.inputProcessor.SetKeyMap(keyMap)

    // 兼容模式：保留旧 Router
    app.legacyRouter = frameworkevent.NewRouter()

    return app
}
```

## 2. 主循环修改

### 2.1 新的消息处理入口

```go
// framework/app.go

// processMsg 统一的消息处理入口
// 这是新的核心路径，所有消息都通过这里转换为 Action
func (a *App) processMsg(msg runtimemsg.Msg) {
    if msg == nil {
        return
    }

    // 1. 转换为 Action
    act := a.inputProcessor.ProcessMsg(msg)

    // 2. 如果无法转换，尝试兼容路径
    if act == nil {
        if a.legacyMode {
            a.handleLegacyMsg(msg)
        }
        return
    }

    // 3. 设置默认目标（焦点组件）
    if act.TargetID == 0 {
        focused := a.focusManager.GetCurrent()
        if focused != nil {
            act.TargetID = focused.GetNodeID()
        }
    }

    // 4. 分发 Action
    result := a.dispatchAction(act)

    // 5. 处理结果
    if result.Handled {
        a.dirty = true
    }
}

// dispatchAction 分发 Action 到 ActionRouter
func (a *App) dispatchAction(act *action.Action) *action.RouterResult {
    // 确保路由器有根节点
    if a.actionRouter.Root == nil {
        // 尝试从 root 获取 LayoutNode
        if layoutNode, ok := a.root.(layout.Node); ok {
            a.actionRouter.SetRoot(layoutNode.(*runtime.LayoutNode))
        }
    }

    // 构建目标注册表
    a.actionRouter.BuildTargetRegistry()

    // 分发
    return a.actionRouter.Dispatch(act)
}

// handleLegacyMsg 兼容模式：处理无法转换的消息
func (a *App) handleLegacyMsg(msg runtimemsg.Msg) {
    ev := frameworkevent.MsgToEvent(msg)
    if ev == nil {
        return
    }

    // 使用旧的事件处理路径
    if a.legacyRouter != nil {
        a.legacyRouter.Route(ev)
    }

    // 兼容：发送到根组件
    if handler, ok := a.root.(frameworkevent.Component); ok {
        if handler.HandleEvent(ev) {
            a.dirty = true
        }
    }
}
```

### 2.2 修改主循环

```go
// framework/app.go

func (a *App) Run() error {
    if err := a.Init(); err != nil {
        return err
    }
    defer a.Close()

    ticker := time.NewTicker(a.tickInterval)
    defer ticker.Stop()

    eventChan := a.pump.Events()
    quitAppChan := a.pump.QuitAppRequested()

    for a.state == StateRunning {
        select {
        case msg := <-eventChan:
            if msg == nil {
                continue
            }

            // ===== 新的统一入口 =====
            a.processMsg(msg)

        case <-ticker.C:
            a.handleTick()
            if a.dirty && a.throttler.ShouldRender() {
                a.render()
                a.throttler.RecordFrameTime(time.Since(time.Now()))
            }

        case <-quitAppChan:
            a.state = StateStopping
            return nil

        case <-a.quit:
            a.state = StateStopping
            return nil

        case <-a.contextMgr.Context().Done():
            a.state = StateStopping
            return nil
        }
    }

    return nil
}
```

## 3. 注册 ActionTarget

### 3.1 从组件树注册

```go
// framework/app.go

// buildActionRegistry 从组件树构建 ActionTarget 注册表
func (a *App) buildActionRegistry() {
    if a.root == nil {
        return
    }

    // 清空旧注册
    a.actionRegistry = NewActionRegistry()

    // 递归注册
    a.registerActionTargets(a.root)
}

// registerActionTargets 递归注册 ActionTarget
func (a *App) registerActionTargets(node interface{}) {
    if node == nil {
        return
    }

    // 获取节点 ID
    var nodeID uint64
    if idProvider, ok := node.(interface{ GetNodeID() uint64 }); ok {
        nodeID = idProvider.GetNodeID()
    } else if idProvider, ok := node.(interface{ ID() string }); ok {
        nodeID = event.StringToNodeID(idProvider.ID())
    }

    // 检查是否实现 ActionTarget
    if nodeID != 0 {
        if target, ok := node.(action.ActionTarget); ok {
            a.actionRegistry.Register(nodeID, target)
        } else if updater, ok := node.(component.Updater); ok {
            // 使用适配器包装旧接口
            adapter := action.NewUpdaterAdapter(updater, nodeID)
            a.actionRegistry.Register(nodeID, adapter)
        } else if handler, ok := node.(frameworkevent.EventHandler); ok {
            // 使用适配器包装 EventHandler
            adapter := action.NewEventHandlerAdapter(handler, nodeID)
            a.actionRegistry.Register(nodeID, adapter)
        }
    }

    // 递归处理子节点
    if container, ok := node.(interface{ Children() []component.Node }); ok {
        for _, child := range container.Children() {
            a.registerActionTargets(child)
        }
    }
}
```

### 3.2 渲染后更新注册表

```go
// framework/app.go

func (a *App) render() {
    // ... 现有渲染逻辑 ...

    // ===== 新增：渲染后更新注册表 =====
    a.buildActionRegistry()

    // 同步到 ActionRouter
    for id, target := range a.actionRegistry.GetAll() {
        a.actionRouter.RegisterTarget(id, target)
    }

    a.dirty = false
}
```

## 4. 集成测试

### 4.1 测试用例

```go
// framework/app_action_test.go

func TestAppProcessMsg_ActionConversion(t *testing.T) {
    app := NewApp()
    app.SetRoot(NewTestComponent())
    app.Init()

    // 创建键盘消息
    keyMsg := runtimemsg.NewKeyMsg(rune('a'), 0, nil)

    // 处理消息
    app.processMsg(keyMsg)

    // 验证 Action 被正确分发
    // ...
}

func TestAppLegacyMode(t *testing.T) {
    app := NewApp()
    app.legacyMode = true
    app.SetRoot(NewTestComponent())
    app.Init()

    // 处理无法转换的消息
    resizeMsg := runtimemsg.NewResizeMsg(80, 24)
    app.processMsg(resizeMsg)

    // 验证兼容路径被调用
    // ...
}

func TestActionRegistry(t *testing.T) {
    registry := NewActionRegistry()

    target := &MockActionTarget{}
    registry.Register(1, target)

    assert.Equal(t, target, registry.Get(1))
    assert.Equal(t, 1, len(registry.GetAll()))

    registry.Unregister(1)
    assert.Nil(t, registry.Get(1))
}
```

## 5. 验证清单

### Phase 1 完成标准

- [ ] App 结构包含 actionRouter 和 inputProcessor
- [ ] processMsg() 正确转换 Msg 为 Action
- [ ] Action 被正确分发到 ActionTarget
- [ ] 兼容模式下旧代码仍能工作
- [ ] 注册表正确收集 ActionTarget
- [ ] 渲染后注册表更新
- [ ] 单元测试通过
- [ ] 集成测试通过

## 6. 回滚方案

如果 Phase 1 出现问题：

1. 设置 `legacyMode = true`
2. 所有消息走 `handleLegacyMsg()`
3. 禁用 Action 路径

```go
// 回滚配置
app.legacyMode = true
app.inputProcessor = nil  // 禁用 Action 转换
```
