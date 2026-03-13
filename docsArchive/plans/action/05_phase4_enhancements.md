# Phase 4: 增强功能

## 目标

在基础 Action 系统稳定后，添加高级功能以提升开发体验和系统能力。

## 1. 异步 Action 支持

### 1.1 问题

当前 Action 是同步的，无法直接处理：
- 网络请求
- 文件 I/O
- 长时间计算

### 1.2 解决方案：AsyncAction

```go
// framework/action/async.go

// AsyncAction 异步 Action
type AsyncAction struct {
    *Action
    async     bool
    onResult  func(result interface{}, err error)
    timeout   time.Duration
    cancel    context.CancelFunc
}

// NewAsyncAction 创建异步 Action
func NewAsyncAction(actionType ActionType) *AsyncAction {
    return &AsyncAction{
        Action: NewAction(actionType),
        async:  true,
    }
}

func (a *AsyncAction) WithTimeout(d time.Duration) *AsyncAction {
    a.timeout = d
    return a
}

func (a *AsyncAction) OnResult(callback func(interface{}, error)) *AsyncAction {
    a.onResult = callback
    return a
}

func (a *AsyncAction) IsAsync() bool {
    return a.async
}

// AsyncActionTarget 异步 Action 处理器接口
type AsyncActionTarget interface {
    ActionTarget
    HandleActionAsync(action *AsyncAction) <-chan AsyncResult
}

// AsyncResult 异步结果
type AsyncResult struct {
    Value interface{}
    Error error
}
```

### 1.3 App 集成

```go
// framework/app.go

func (a *App) processMsg(msg runtimemsg.Msg) {
    act := a.inputProcessor.ProcessMsg(msg)
    if act == nil {
        return
    }

    // 检查是否异步 Action
    if asyncAct, ok := act.(*action.AsyncAction); ok && asyncAct.IsAsync() {
        go a.dispatchAsyncAction(asyncAct)
        return
    }

    // 同步分发
    result := a.actionRouter.Dispatch(act)
    if result.Handled {
        a.dirty = true
    }
}

func (a *App) dispatchAsyncAction(act *action.AsyncAction) {
    ctx := context.Background()
    if act.timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, act.timeout)
        defer cancel()
        act.cancel = cancel
    }

    // 查找目标
    target := a.actionRegistry.Get(act.TargetID)
    if target == nil {
        if act.onResult != nil {
            act.onResult(nil, errors.New("target not found"))
        }
        return
    }

    // 检查是否支持异步
    if asyncTarget, ok := target.(action.AsyncActionTarget); ok {
        resultChan := asyncTarget.HandleActionAsync(act)
        select {
        case result := <-resultChan:
            if act.onResult != nil {
                act.onResult(result.Value, result.Error)
            }
            a.dirty = true
            a.requestRender()
        case <-ctx.Done():
            if act.onResult != nil {
                act.onResult(nil, ctx.Err())
            }
        }
    } else {
        // 降级为同步
        handled := target.HandleAction(act.Action)
        if act.onResult != nil {
            act.onResult(handled, nil)
        }
    }
}

func (a *App) requestRender() {
    // 通知主循环进行渲染
    select {
    case a.renderChan <- struct{}{}:
    default:
    }
}
```

### 1.4 使用示例

```go
// 异步加载数据
func (c *DataList) HandleActionAsync(act *action.AsyncAction) <-chan action.AsyncResult {
    resultChan := make(chan action.AsyncResult, 1)

    go func() {
        defer close(resultChan)

        // 模拟网络请求
        data, err := fetchDataFromAPI()
        if err != nil {
            resultChan <- action.AsyncResult{Error: err}
            return
        }

        c.items = data
        resultChan <- action.AsyncResult{Value: data}
    }()

    return resultChan
}

// 触发异步 Action
act := action.NewAsyncAction(action.ActionDataLoad).
    WithTimeout(5 * time.Second).
    OnResult(func(result interface{}, err error) {
        if err != nil {
            log.Printf("Load failed: %v", err)
            return
        }
        log.Printf("Loaded %d items", len(result.([]Item)))
    })

app.dispatchAsyncAction(act)
```

## 2. Action 中间件系统

### 2.1 中间件接口

```go
// framework/action/middleware.go

// ActionMiddleware 中间件接口
type ActionMiddleware interface {
    // Name 中间件名称
    Name() string

    // Before 在 Action 分发前调用
    // 返回 nil 表示拦截该 Action
    // 返回修改后的 Action 继续传播
    Before(action *Action) *Action

    // After 在 Action 分发后调用
    After(action *Action, result *RouterResult)
}

// MiddlewareChain 中间件链
type MiddlewareChain struct {
    middlewares []ActionMiddleware
}

func NewMiddlewareChain(middlewares ...ActionMiddleware) *MiddlewareChain {
    return &MiddlewareChain{middlewares: middlewares}
}

func (c *MiddlewareChain) Before(action *Action) *Action {
    for _, mw := range c.middlewares {
        action = mw.Before(action)
        if action == nil {
            return nil // 被拦截
        }
    }
    return action
}

func (c *MiddlewareChain) After(action *Action, result *RouterResult) {
    // 逆序调用 After
    for i := len(c.middlewares) - 1; i >= 0; i-- {
        c.middlewares[i].After(action, result)
    }
}
```

### 2.2 内置中间件

```go
// framework/action/middleware_logging.go

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
    logger *log.Logger
    level  string
}

func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
    return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Name() string { return "logging" }

func (m *LoggingMiddleware) Before(action *Action) *Action {
    m.logger.Debug("[Action] Dispatching: %s", action)
    return action
}

func (m *LoggingMiddleware) After(action *Action, result *RouterResult) {
    m.logger.Debug("[Action] Completed: %s, Handled=%v, Phase=%s",
        action, result.Handled, result.Phase)
}

// framework/action/middleware_throttle.go

// ThrottleMiddleware 节流中间件
type ThrottleMiddleware struct {
    interval   time.Duration
    lastAction map[ActionType]time.Time
    mu         sync.Mutex
}

func NewThrottleMiddleware(interval time.Duration) *ThrottleMiddleware {
    return &ThrottleMiddleware{
        interval:   interval,
        lastAction: make(map[ActionType]time.Time),
    }
}

func (m *ThrottleMiddleware) Name() string { return "throttle" }

func (m *ThrottleMiddleware) Before(action *Action) *Action {
    m.mu.Lock()
    defer m.mu.Unlock()

    now := time.Now()
    last, exists := m.lastAction[action.Type]

    if exists && now.Sub(last) < m.interval {
        return nil // 拦截：触发太频繁
    }

    m.lastAction[action.Type] = now
    return action
}

func (m *ThrottleMiddleware) After(action *Action, result *RouterResult) {}

// framework/action/middleware_validation.go

// ValidationMiddleware 验证中间件
type ValidationMiddleware struct {
    validators map[ActionType]ActionValidator
}

type ActionValidator func(action *Action) error

func NewValidationMiddleware() *ValidationMiddleware {
    return &ValidationMiddleware{
        validators: make(map[ActionType]ActionValidator),
    }
}

func (m *ValidationMiddleware) Name() string { return "validation" }

func (m *ValidationMiddleware) RegisterValidator(actionType ActionType, validator ActionValidator) {
    m.validators[actionType] = validator
}

func (m *ValidationMiddleware) Before(action *Action) *Action {
    if validator, exists := m.validators[action.Type]; exists {
        if err := validator(action); err != nil {
            log.Printf("[Validation] Action %s rejected: %v", action.Type, err)
            return nil
        }
    }
    return action
}

func (m *ValidationMiddleware) After(action *Action, result *RouterResult) {}
```

### 2.3 使用中间件

```go
// framework/app.go

func (a *App) setupMiddlewares() {
    chain := action.NewMiddlewareChain(
        action.NewLoggingMiddleware(a.logger),
        action.NewThrottleMiddleware(50*time.Millisecond),
        action.NewValidationMiddleware(),
    )

    a.actionRouter.SetMiddleware(chain)
}
```

## 3. Action 撤销/重做

### 3.1 撤销系统

```go
// framework/action/undo.go

// UndoableAction 可撤销的 Action
type UndoableAction struct {
    *Action
    undo    func() error
    redo    func() error
    description string
}

// UndoManager 撤销管理器
type UndoManager struct {
    undoStack []*UndoableAction
    redoStack []*UndoableAction
    maxSize   int
    mu        sync.Mutex
}

func NewUndoManager(maxSize int) *UndoManager {
    return &UndoManager{
        maxSize: maxSize,
    }
}

// Push 压入可撤销操作
func (m *UndoManager) Push(action *UndoableAction) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.undoStack = append(m.undoStack, action)

    // 限制栈大小
    if len(m.undoStack) > m.maxSize {
        m.undoStack = m.undoStack[1:]
    }

    // 清空 redo 栈
    m.redoStack = nil
}

// Undo 撤销
func (m *UndoManager) Undo() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if len(m.undoStack) == 0 {
        return errors.New("nothing to undo")
    }

    action := m.undoStack[len(m.undoStack)-1]
    m.undoStack = m.undoStack[:len(m.undoStack)-1]

    if err := action.undo(); err != nil {
        // 撤销失败，放回栈中
        m.undoStack = append(m.undoStack, action)
        return err
    }

    m.redoStack = append(m.redoStack, action)
    return nil
}

// Redo 重做
func (m *UndoManager) Redo() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if len(m.redoStack) == 0 {
        return errors.New("nothing to redo")
    }

    action := m.redoStack[len(m.redoStack)-1]
    m.redoStack = m.redoStack[:len(m.redoStack)-1]

    if err := action.redo(); err != nil {
        m.redoStack = append(m.redoStack, action)
        return err
    }

    m.undoStack = append(m.undoStack, action)
    return nil
}

// CanUndo 是否可撤销
func (m *UndoManager) CanUndo() bool {
    return len(m.undoStack) > 0
}

// CanRedo 是否可重做
func (m *UndoManager) CanRedo() bool {
    return len(m.redoStack) > 0
}

// GetUndoDescription 获取撤销描述
func (m *UndoManager) GetUndoDescription() string {
    if len(m.undoStack) == 0 {
        return ""
    }
    return m.undoStack[len(m.undoStack)-1].description
}
```

### 3.2 集成到 ActionRouter

```go
// framework/action/router.go

func (r *Router) Dispatch(act *Action) *RouterResult {
    // 特殊处理撤销/重做
    if act.Type == ActionUndo {
        if err := r.undoManager.Undo(); err != nil {
            log.Printf("Undo failed: %v", err)
        }
        return &RouterResult{Handled: true}
    }

    if act.Type == ActionRedo {
        if err := r.undoManager.Redo(); err != nil {
            log.Printf("Redo failed: %v", err)
        }
        return &RouterResult{Handled: true}
    }

    // 正常分发
    result := r.dispatchNormal(act)

    // 如果 Action 实现了 UndoableAction，注册到撤销管理器
    if undoable, ok := act.(*UndoableAction); ok && result.Handled {
        r.undoManager.Push(undoable)
    }

    return result
}
```

### 3.3 使用示例

```go
// 文本编辑器支持撤销
type TextEditor struct {
    content    string
    undoManager *action.UndoManager
}

func (e *TextEditor) HandleAction(act *action.Action) bool {
    switch act.Type {
    case action.ActionInputText:
        char, _ := act.GetPayloadString()

        // 保存旧状态
        oldContent := e.content
        cursor := e.cursor

        // 执行操作
        e.insertChar(char)

        // 创建可撤销 Action
        undoable := &action.UndoableAction{
            Action: act,
            undo: func() error {
                e.content = oldContent
                e.cursor = cursor
                return nil
            },
            redo: func() error {
                e.insertChar(char)
                return nil
            },
            description: "Insert '" + char + "'",
        }

        // 触发撤销注册（通过返回特殊 Action）
        return true
    }

    return false
}
```

## 4. Action 对象池

### 4.1 实现

```go
// framework/action/pool.go

// ActionPool Action 对象池
type ActionPool struct {
    pool sync.Pool
}

func NewActionPool() *ActionPool {
    return &ActionPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &Action{
                    Meta: make(map[string]interface{}),
                }
            },
        },
    }
}

// Get 从池中获取 Action
func (p *ActionPool) Get() *Action {
    action := p.pool.Get().(*Action)
    action.reset()
    return action
}

// Put 归还 Action 到池中
func (p *ActionPool) Put(action *Action) {
    if action == nil {
        return
    }
    // 清理引用
    action.Payload = nil
    action.Meta = make(map[string]interface{})
    p.pool.Put(action)
}

// reset 重置 Action 状态
func (a *Action) reset() {
    a.Type = ""
    a.Payload = nil
    a.Source = ""
    a.TargetID = 0
    a.stopped = false
    for k := range a.Meta {
        delete(a.Meta, k)
    }
}
```

### 4.2 使用

```go
// framework/app.go

var actionPool = action.NewActionPool()

func (a *App) processMsg(msg runtimemsg.Msg) {
    act := actionPool.Get()
    defer actionPool.Put(act)

    // 填充 Action
    if !a.inputProcessor.ProcessMsgInto(msg, act) {
        return
    }

    // 分发
    result := a.actionRouter.Dispatch(act)
    // ...
}
```

## 5. Action 调试工具

### 5.1 Action Inspector

```go
// framework/debug/action_inspector.go

// ActionInspector Action 调试检查器
type ActionInspector struct {
    history    []*ActionRecord
    maxHistory int
    enabled    bool
}

type ActionRecord struct {
    Action    *Action
    Timestamp time.Time
    Result    *RouterResult
    Duration  time.Duration
}

func NewActionInspector(maxHistory int) *ActionInspector {
    return &ActionInspector{
        maxHistory: maxHistory,
        history:    make([]*ActionRecord, 0, maxHistory),
    }
}

func (i *ActionInspector) Name() string { return "inspector" }

func (i *ActionInspector) Before(action *Action) *Action {
    if !i.enabled {
        return action
    }

    // 记录开始时间
    action.Meta["_start"] = time.Now()
    return action
}

func (i *ActionInspector) After(action *Action, result *RouterResult) {
    if !i.enabled {
        return
    }

    start, _ := action.Meta["_start"].(time.Time)
    record := &ActionRecord{
        Action:    action.Clone(),
        Timestamp: start,
        Result:    result,
        Duration:  time.Since(start),
    }

    i.history = append(i.history, record)
    if len(i.history) > i.maxHistory {
        i.history = i.history[1:]
    }
}

func (i *ActionInspector) GetHistory() []*ActionRecord {
    return i.history
}

func (i *ActionInspector) ClearHistory() {
    i.history = nil
}

func (i *ActionInspector) Enable(enabled bool) {
    i.enabled = enabled
}

// FormatHistory 格式化历史记录
func (i *ActionInspector) FormatHistory() string {
    var sb strings.Builder
    for _, record := range i.history {
        sb.WriteString(fmt.Sprintf("[%s] %s -> %v (%v)\n",
            record.Timestamp.Format("15:04:05.000"),
            record.Action,
            record.Result.Handled,
            record.Duration,
        ))
    }
    return sb.String()
}
```

### 5.2 UI 集成

在现有的 Inspector 中添加 Action 历史面板：

```go
// inspector/tabs/action_tab.go

type ActionTab struct {
    inspector *ActionInspector
    selected  int
}

func (t *ActionTab) Render() string {
    history := t.inspector.GetHistory()

    var sb strings.Builder
    sb.WriteString("Action History:\n")
    sb.WriteString("─────────────────────────────\n")

    for i, record := range history {
        prefix := "  "
        if i == t.selected {
            prefix = "> "
        }

        handled := "✓"
        if !record.Result.Handled {
            handled = "✗"
        }

        sb.WriteString(fmt.Sprintf("%s%s %s %v\n",
            prefix,
            handled,
            record.Action.Type,
            record.Duration,
        ))
    }

    return sb.String()
}
```

## 6. Phase 4 完成标准

- [ ] 异步 Action 支持完成
- [ ] 中间件系统完成
- [ ] 撤销/重做系统完成
- [ ] Action 对象池完成
- [ ] 调试工具集成完成
- [ ] 文档更新完成
- [ ] 性能测试通过
