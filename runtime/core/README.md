# Core

Runtime 核心引擎和基础设施。提供运行时引擎、上下文管理、panic 恢复等核心功能。

## 职责

- **Runtime 引擎**: 整合布局、绘制、状态、焦点、输入等子系统
- **上下文管理**: 提供优雅关闭和超时控制（ContextManager）
- **Panic 恢复**: 确保应用在发生不可恢复错误时能够正确清理资源
- **Goroutine 管理**: 安全的 goroutine 启动和生命周期管理

## 核心概念

### 1. Runtime 引擎

Runtime 是 Mint TUI 的核心引擎，整合所有子系统：

```
┌─────────────────────────────────────────┐
│              Runtime                    │
├─────────────────────────────────────────┤
│  Platform     Input    KeyMap          │
│  (屏幕/输入)  → RawInput → Action      │
│                                          │
│  FocusManager  LayoutEngine             │
│  (焦点管理)    (布局计算)               │
│                                          │
│  ActionDispatcher  StateTracker         │
│  (Action 分发)   (状态追踪)             │
│                                          │
│  PaintBuffer   DirtyTracker             │
│  (绘制缓冲)    (脏区域跟踪)             │
└─────────────────────────────────────────┘
```

**核心子系统**:
- **Platform**: 提供屏幕、输入、信号等平台抽象
- **LayoutEngine**: 布局引擎，计算组件位置和尺寸
- **FocusManager**: 焦点管理器，处理焦点导航
- **StateTracker**: 状态追踪器，支持 Undo/Redo
- **ActionDispatcher**: Action 分发器，将 Action 分发到目标组件
- **KeyMap**: 输入映射，将 RawInput 转换为 Action
- **Buffer**: 绘制缓冲区
- **DirtyTracker**: 脏区域跟踪器

**主循环流程**:
1. **Input**: 从 Platform 读取 RawInput
2. **Map**: KeyMap 将 RawInput 转换为 Action
3. **Dispatch**: ActionDispatcher 将 Action 分发到目标组件
4. **Layout**: LayoutEngine 重新布局（如果需要）
5. **Paint**: 组件绘制到 Buffer
6. **Sync**: Buffer 同步到 Screen

### 2. ContextManager（上下文管理器）

ContextManager 提供优雅关闭和超时控制：

**特性**:
- 关闭时自动取消所有 goroutine
- 支持超时关闭
- 传递上下文值
- WaitGroup 集成

**主要方法**:
- `Go(f func(context.Context) error)`: 在上下文中启动 goroutine
- `Shutdown(timeout...)`: 优雅关闭
- `WithTimeout(timeout)`: 创建子上下文带超时
- `WithValue(key, value)`: 设置上下文值

### 3. Recovery（Panic 恢复系统）

Recovery 提供全面的 panic 恢复能力：

**特**性:
- 恢复终端状态（退出备用屏幕、显示光标等）
- 记录 panic 到日志和文件
- 支持多个 Handlers
- 提供 SafeRunner 安全运行器

**内置 Handlers**:
- `LoggingPanicHandler`: 输出 panic 到日志
- `MetricsPanicHandler`: 统计 panic 次数和记录
- `CrashReportPanicHandler`: 生成崩溃报告文件
- `NotificationPanicHandler`: 自定义通知处理

**终端恢复步骤**:
1. 设置正常模式
2. 显示光标
3. 退出备用屏幕
4. 启用回显
5. 刷新输出
6. 关闭终端

## 使用示例

### 创建和启动 Runtime

```go
// 创建 Runtime（传入 Platform 实现）
runtime := core.NewRuntime(platform)

// 设置根节点
runtime.SetRoot(rootNode)

// 启动 Runtime
if err := runtime.Start(); err != nil {
    log.Fatalf("Failed to start runtime: %v", err)
}

// 主循环
runtime.Run()
```

### Action 处理注册

```go
// 注册 Action Target
runtime.RegisterActionTarget(myComponent)

// 订阅全局 Action
unsubscribe := runtime.SubscribeGlobalAction(action.ActionQuit, func(act *action.Action) {
    fmt.Println("Quitting...")
    runtime.Shutdown()
})

// 设置默认 Action Handler
runtime.SetDefaultActionHandler(func(act *action.Action) {
    fmt.Printf("Unhandled action: %s\n", act.Type)
})

// 取消订阅
unsubscribe()
```

### 焦点管理

```go
// 移动焦点到下一个
nextID, ok := runtime.FocusNext()

// 移动焦点到上一个
prevID, ok := runtime.FocusPrev()

// 设置焦点到指定组件
ok := runtime.FocusSpecific("button-submit")

// 获取当前焦点
focusID, ok := runtime.GetFocused()

// 推入焦点域（用于模态对话框）
modalScope := focus.NewScope("modal", "modal-dialog")
runtime.PushFocusScope(modalScope)

// 弹出焦点域
runtime.PopFocusScope()
```

### 状态管理

```go
// 获取当前状态
snapshot := runtime.GetState()

// StateTracker 支持 Undo/Redo
tracker := runtime.stateTracker // 内部使用

// 在 Action 处理前后记录状态
before := tracker.BeforeAction()
// ... 处理 Action ...
tracker.AfterAction(before)
```

### 手动更新和渲染

```go
// 更新状态
if err := runtime.Update(); err != nil {
    log.Printf("Update failed: %v", err)
}

// 渲染到屏幕
if err := runtime.Render(); err != nil {
    log.Printf("Render failed: %v", err)
}
```

### Runtime 注销

```go
// 停止 Runtime
if err := runtime.Stop(); err != nil {
    log.Printf("Stop failed: %v", err)
}

// 优雅关闭（带超时）
if err := runtime.Shutdown(5 * time.Second); err != nil {
    log.Printf("Shutdown timeout: %v", err)
}
```

### ContextManager 使用

```go
// 创建上下文管理器
ctxMgr := core.NewContextManager(context.Background())

// 设置上下文值
ctxMgr.WithValue(core.KeyUser, "john")
ctxMgr.WithValue(core.KeySessionID, "session-123")

// 在上下文中启动 goroutine
ctxMgr.Go(func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            // 上下文已取消，清理资源
            return nil
        default:
            // 正常工作
            time.Sleep(100 * time.Millisecond)
        }
    }
})

// 优雅关闭
err := ctxMgr.Shutdown(3 * time.Second)
if err != nil {
    log.Printf("Shutdown timeout: %v", err)
}
```

### Panic 恢复

```go
// 创建 Recovery
recovery := core.NewRecovery(terminal)

// 添加 Handlers
recovery.AddHandler(core.NewLoggingPanicHandler(os.Stderr))
recovery.AddHandler(core.NewCrashReportPanicHandler("/tmp/crashes"))
recovery.AddHandler(core.NewMetricsPanicHandler(10))

// 启用 panic 日志文件
recovery.EnablePanicLog("panic.log")

// 在主函数中使用
func main() {
    defer recovery.Handle(recover())

    // 启动 Runtime
    runtime := core.NewRuntime(platform)
    runtime.Run()
}
```

### SafeRunner 安全运行

```go
recovery := core.NewRecovery(terminal)

// 创建 SafeRunner
runner := core.NewSafeRunner(recovery)

// 安全运行函数
err := runner.Run(func() error {
    // 可能发生 panic 的代码
    doSomething()
    return nil
})

if err != nil {
    log.Printf("Function failed: %v", err)
}

// 在上下文中安全运行
err = runner.RunWithContext(context.Background(), func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // 处理任务
        }
    }
})
```

### 便捷 Panic 处理

```go
// 简单恢复 panic 并退出
defer core.RecoverPanic(terminal)

// 在函数中添加 panic 恢复
core.WithRecovery(terminal, func() {
    // 可能发生 panic 的代码
})

// 安全启动 goroutine
core.SafeGo(terminal, func() {
    // goroutine 代码
})

// 安全启动带上下文的 goroutine
core.SafeGoWithContext(ctx, terminal, func(ctx context.Context) {
    // 带上下文的 goroutine 代码
})
```

## 核心类型

### Runtime

```go
type Runtime struct {
    platform          platform.RuntimePlatform
    layoutEngine      *layout.Engine
    focusManager      *focus.ManagerV3
    stateTracker      *state.Tracker
    actionDispatcher  *action.Dispatcher
    keyMap            *input.KeyMap
    contextManager    *ContextManager
    root              layout.Node
    buffer            *paint.Buffer
    dirtyTracker      *paint.DirtyTracker
}

func NewRuntime(pf platform.RuntimePlatform) *Runtime
func (r *Runtime) Start() error
func (r *Runtime) Stop() error
func (r *Runtime) Run() error
func (r *Runtime) Update() error
func (r *Runtime) Render() error
func (r *Runtime) ProcessInput() error
func (r *Runtime) Shutdown(timeout...time.Duration) error
```

### ContextManager

```go
type ContextManager struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func NewContextManager(parent context.Context) *ContextManager
func (m *ContextManager) Context() context.Context
func (m *ContextManager) Go(f func(context.Context) error)
func (m *ContextManager) Shutdown(timeout...time.Duration) error
func (m *ContextManager) WithValue(key ContextKey, value interface{})
func (m *ContextManager) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc)
```

### Recovery

```go
type Recovery struct {
    handlers     []PanicHandler
    terminal     Terminal
    panicLogFile *os.File
    logWriter    io.Writer
}

func NewRecovery(terminal Terminal) *Recovery
func (r *Recovery) AddHandler(h PanicHandler)
func (r *Recovery) Handle(panicValue interface{})
func (r *Recovery) EnablePanicLog(filename string) error
func (r *Recovery) Close() error
```

### SafeRunner

```go
type SafeRunner struct {
    recovery *Recovery
    onPanic  func(interface{})
}

func NewSafeRunner(recovery *Recovery) *SafeRunner
func (s *SafeRunner) Run(fn SafeFunc) error
func (s *SafeRunner) RunWithContext(ctx context.Context, fn func(context.Context) error) error
```

## 文件结构

- `runtime.go` - Runtime 核心引擎
- `context.go` - 上下文管理器
- `recovery.go` - Panic 恢复系统

## 依赖

**可以依赖**:
- `runtime/platform` - 平台抽象
- `runtime/action` - Action 系统
- `runtime/focus` - 焦点管理
- `runtime/input` - 输入处理
- `runtime/layout` - 布局引擎
- `runtime/paint` - 绘制系统
- `runtime/state` - 状态追踪
- 标准库: `context`, `sync`, `time`, `runtime`, `debug`

**不能依赖**:
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 与其他模块集成

Runtime 是所有子系统的协调者：

```
Platform → RawInput → KeyMap → Action
                          ↓
                    (ActionDispatcher)
                          ↓
        ┌─────────────────┴─────────────────┐
        ↓                                   ↓
   (FocusManager)                    (Component)
        ↓                                   ↓
   (LayoutEngine) ←─────────────────────────┘
        ↓
   (PaintBuffer) → (DirtyTracker) → Screen
        ↓
   (StateTracker)
```

## 最佳实践

### 1. 总是使用 ContextManager 管理 goroutine

```go
// 推荐：使用 ContextManager.Go
runtime.Go(func(ctx context.Context) error {
    // 处理任务，自动处理取消
    return nil
})

// 不推荐：直接 go
go someFunction()
```

### 2. 使用 Recovery 处理 panic

```go
defer func() {
    if r := recover(); r != nil {
        recovery.Handle(r)
    }
}()
```

### 3. 注册所有 Action Target

```go
// 在应用启动时注册所有组件
runtime.RegisterActionTarget(button1)
runtime.RegisterActionTarget(textField1)
runtime.RegisterActionTarget(listView1)
```

### 4. 正确使用焦点域

```go
// 推入焦点域
runtime.PushFocusScope(modalScope)
defer runtime.PopFocusScope()

// 在域内操作
runtime.FocusSpecific("modal-button-1")
```

### 5. 设置适当的帧率

使用 `render` 模块的 Throttler 控制帧率，避免过度渲染。

## 常见问题

### Q: Runtime 和 Framework 有什么区别？

A:
- **Runtime**: 纯 Go 的核心引擎，提供 TUI 运行所需的所有功能，不依赖外部库
- **Framework**: 应用层，提供组件、表单、样式等高级 API，可以依赖 lipgloss 等库

### Q: 为什么需要 ContextManager？

A: TUI 应用通常有多个 goroutine（例如：输入读取、定时器、动画等）。ContextManager 确保在应用关闭时，所有 goroutine 都能优雅退出，避免资源泄漏。

### Q: Panic 恢复为什么重要？

A: TUI 应用会修改终端模式（如备用屏幕、隐藏光标）。panic 时如果不恢复终端状态，用户会看到终端状态混乱。Recovery 确保无论何时 panic，都能正确恢复终端。

### Q: 如何调试 Runtime 启动问题？

A: 检查以下几点：
1. Platform 实现是否正确
2. 根节点是否正确设置
3. 焦点域是否正确初始化
4. 查看日志输出

### Q: StateTracker 如何支持时间旅行调试？

A: DevTools 可以利用 StateTracker 记录的所有状态快照，实现 Undo/Redo 操作，甚至"跳转到任意状态"。这对于调试复杂交互非常有用。
