# Sandbox机制设计方案 V3 (系统兼容优化版)

> 基于 V2 方案的审查，结合现有系统实际情况进行优化

## 🔴 重要：循环引用问题及解决方案

### 问题分析

当前 runtime 包的依赖结构是**严格分层**的：

```
                    ┌─────────────────────┐
                    │   runtime/engine    │  ◄── 顶层：整合所有子系统
                    └─────────────────────┘
                              │
          ┌───────────┬───────┼───────┬───────────┐
          ▼           ▼       ▼       ▼           ▼
     ┌────────┐  ┌────────┐ ┌─────┐ ┌──────┐ ┌──────────┐
     │ runtime│  │  event │ │focus│ │paint │ │ platform │
     └────────┘  └────────┘ └─────┘ └──────┘ └──────────┘
          │                           │
          └───────────────────────────┼───────► style (最底层)
```

如果 `sandbox/` 作为独立模块：

```
❌ 错误设计：
sandbox ──► runtime/engine ──► runtime/platform
                              ──► runtime/paint
                              
如果 engine 也需要 sandbox（用于测试模式）：
runtime/engine ──► sandbox  ══► 循环引用！
```

### 解决方案：接口隔离 + 依赖注入

**核心原则：sandbox 只依赖 runtime 的底层包，不依赖 engine**

```
✅ 正确设计：

sandbox/
├── 依赖 runtime/platform  ✅ (底层，无循环)
├── 依赖 runtime/paint     ✅ (底层，无循环)
├── 依赖 runtime/event     ✅ (底层，无循环)
└── 不依赖 runtime/engine  ✅ (避免循环)

runtime/engine/
└── 依赖 sandbox           ✅ (engine 在上层，可以依赖 sandbox)
```

### 依赖方向图（最终设计）

```
                    ┌─────────────────────┐
                    │        ui/          │  ◄── 应用层
                    └─────────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │   runtime/engine    │  ◄── 可选依赖 sandbox (测试模式)
                    └─────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
     ┌─────────┐        ┌──────────┐        ┌─────────┐
     │ sandbox │        │ runtime/ │        │  focus  │
     │         │        │ (其他)   │        │         │
     └─────────┘        └──────────┘        └─────────┘
          │                   │
          └───────┬───────────┘
                  ▼
     ┌────────┬──────────┬──────────┐
     │ event  │ platform │  paint   │  ◄── 底层包 (无依赖)
     └────────┴──────────┴──────────┘
                  │
                  ▼
             ┌────────┐
             │ style  │  ◄── 最底层
             └────────┘
```

### 关键设计决策

| 决策 | 说明 |
|------|------|
| sandbox 只依赖底层包 | `platform`, `paint`, `event`, `style` |
| sandbox 不依赖 engine | 避免循环引用 |
| engine 可选依赖 sandbox | 通过接口注入，用于测试模式 |
| 使用接口而非具体类型 | `EventSource`, `EventSink` 接口解耦 |

### 接口隔离示例

```go
// sandbox/interfaces.go - sandbox 定义接口

// Renderer 渲染器接口 (由 engine 实现)
type Renderer interface {
    Render(buf *paint.Buffer) error
}

// EventDispatcher 事件分发接口 (由 engine 实现)  
type EventDispatcher interface {
    Dispatch(event platform.RawInput) error
}
```

```go
// runtime/engine/engine.go - engine 实现接口

func (e *Engine) SetSandbox(sb sandbox.Sandbox) {
    e.sandbox = sb
}

// Engine 实现 sandbox.Renderer 接口
func (e *Engine) Render(buf *paint.Buffer) error {
    // ...
}
```

这样 sandbox 定义接口，engine 实现接口，依赖方向始终是 engine → sandbox，不会产生循环。

---

## 变更摘要

### 相对于 V2 的主要改进

| 问题 | V2 设计 | V3 优化 |
|------|---------|---------|
| 包路径错误 | `framework/event` | 正确使用 `runtime/event` |
| 调度器重复 | 新建 `sandbox/scheduler.go` | 复用 `runtime/scheduler` |
| 钩子类型缺失 | `hookKey` 未定义 | 完整定义 `Phase` 和 `HookKey` |
| 事件类型不匹配 | 自定义事件类型 | 适配 `platform.RawInput` |
| 链式API错误处理 | 返回类型混乱 | 统一 Result 模式 |
| Buffer接口 | 自定义 Buffer | 复用 `runtime/paint.Buffer` |

## 目录结构

```
mint/
├── sandbox/                      # 独立沙箱模块
│   ├── sandbox.go                # 核心接口定义
│   ├── lifecycle.go              # 生命周期状态机
│   ├── events.go                 # 事件注入系统
│   ├── buffer.go                 # 缓冲区管理 (包装 paint.Buffer)
│   ├── snapshot.go               # 快照系统
│   ├── config.go                 # 配置系统
│   ├── types.go                  # 公共类型定义
│   ├── errors.go                 # 错误定义
│   │
│   ├── adapter/                  # 适配器层 (桥接现有系统)
│   │   ├── input.go              # platform.RawInput 适配
│   │   ├── scheduler.go          # runtime/scheduler 适配
│   │   └── event.go              # runtime/event 适配
│   │
│   ├── real/                     # 真实环境实现
│   │   ├── sandbox.go
│   │   ├── input.go              # 包装 platform.InputReader
│   │   └── terminal.go
│   │
│   ├── mock/                     # 模拟环境实现
│   │   ├── sandbox.go
│   │   ├── queue.go              # 有界事件队列
│   │   ├── assertions.go
│   │   └── testapi.go            # 测试辅助API
│   │
│   ├── replay/                   # 回放环境实现
│   │   ├── sandbox.go
│   │   ├── player.go
│   │   └── recorder.go
│   │
│   └── testing/                  # 测试工具
│       ├── runner.go
│       ├── reporter.go
│       └── helpers.go
│
├── runtime/                      # 现有运行时 (保持不变)
│   ├── event/
│   ├── platform/
│   ├── scheduler/
│   └── paint/
│
└── ui/                           # UI层集成
    └── test.go
```

## 1. 核心类型定义

### 1.1 公共类型 (types.go)

```go
// sandbox/types.go

package sandbox

import (
    "time"

    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/platform"
)

// SandboxType 沙箱类型
type SandboxType int

const (
    TypeReal   SandboxType = iota // 真实终端环境
    TypeMock                      // 模拟测试环境
    TypeReplay                    // 回放环境
)

func (t SandboxType) String() string {
    switch t {
    case TypeReal:
        return "real"
    case TypeMock:
        return "mock"
    case TypeReplay:
        return "replay"
    default:
        return "unknown"
    }
}

// State 沙箱状态
type State int

const (
    StateStopped State = iota
    StateInitialized
    StateRunning
    StatePaused
    StateError
)

func (s State) String() string {
    switch s {
    case StateStopped:
        return "stopped"
    case StateInitialized:
        return "initialized"
    case StateRunning:
        return "running"
    case StatePaused:
        return "paused"
    case StateError:
        return "error"
    default:
        return "unknown"
    }
}

// Phase 生命周期阶段
type Phase int

const (
    PhaseBefore Phase = iota // 状态转换前
    PhaseAfter               // 状态转换后
)

// HookKey 钩子键
type HookKey struct {
    State State
    Phase Phase
}

// InjectionStrategy 事件注入策略
type InjectionStrategy int

const (
    InjectProhibited InjectionStrategy = iota // 禁止注入 (真实环境)
    InjectAllowed                              // 允许注入 (测试环境)
    InjectRecorded                             // 仅录制 (录制模式)
)

// EvictPolicy 事件淘汰策略
type EvictPolicy int

const (
    EvictOldest     EvictPolicy = iota // 淘汰最旧的
    EvictByPriority                    // 按优先级淘汰
    EvictPersist                       // 持久化到磁盘
)

// SnapshotLevel 快照级别
type SnapshotLevel int

const (
    SnapshotMinimal  SnapshotLevel = iota // 仅渲染缓冲区
    SnapshotStandard                      // 缓冲区+事件历史
    SnapshotFull                          // 包括应用状态
)

// InputEvent 统一输入事件 (包装 platform.RawInput)
type InputEvent struct {
    Raw       platform.RawInput
    Injected  bool      // 是否为注入事件
    Timestamp time.Time
}

// BufferWrapper 缓冲区包装器 (复用 paint.Buffer)
type BufferWrapper struct {
    *paint.Buffer
    history []*paint.Buffer // 历史快照
    maxHistory int
}
```

### 1.2 错误定义 (errors.go)

```go
// sandbox/errors.go

package sandbox

import "errors"

var (
    // 生命周期错误
    ErrInvalidTransition = errors.New("sandbox: invalid state transition")
    ErrNotInitialized    = errors.New("sandbox: not initialized")
    ErrAlreadyRunning    = errors.New("sandbox: already running")
    ErrNotRunning        = errors.New("sandbox: not running")

    // 事件注入错误
    ErrInjectionNotAllowed = errors.New("sandbox: event injection not allowed")
    ErrInvalidStrategy     = errors.New("sandbox: invalid injection strategy")
    ErrQueueFull           = errors.New("sandbox: event queue full")
    ErrQueueEmpty          = errors.New("sandbox: event queue empty")

    // 快照错误
    ErrSnapshotNotFound = errors.New("sandbox: snapshot not found")
    ErrSnapshotCorrupt  = errors.New("sandbox: snapshot data corrupted")
    ErrRestoreFailed    = errors.New("sandbox: restore failed")

    // 配置错误
    ErrInvalidConfig = errors.New("sandbox: invalid configuration")

    // 断言错误
    ErrAssertionFailed = errors.New("sandbox: assertion failed")
    ErrTimeout         = errors.New("sandbox: operation timeout")
)

// AssertionError 断言错误详情
type AssertionError struct {
    Message  string
    Expected interface{}
    Actual   interface{}
    Selector string
}

func (e *AssertionError) Error() string {
    return e.Message
}
```

## 2. 核心接口

### 2.1 Sandbox 接口 (sandbox.go)

```go
// sandbox/sandbox.go

package sandbox

import (
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/platform"
)

// Sandbox 沙箱核心接口
type Sandbox interface {
    // ========================================================================
    // 生命周期
    // ========================================================================

    // Initialize 初始化沙箱
    Initialize(config *Config) error

    // Run 运行沙箱主循环
    Run() error

    // Pause 暂停沙箱
    Pause() error

    // Resume 恢复沙箱
    Resume() error

    // Close 关闭沙箱并释放资源
    Close() error

    // ========================================================================
    // 状态查询
    // ========================================================================

    // State 获取当前状态
    State() State

    // Type 获取沙箱类型
    Type() SandboxType

    // Config 获取配置
    Config() *Config

    // ========================================================================
    // 缓冲区操作
    // ========================================================================

    // Buffer 获取渲染缓冲区
    Buffer() *paint.Buffer

    // SetBuffer 设置渲染缓冲区
    SetBuffer(buf *paint.Buffer)

    // Resize 调整缓冲区大小
    Resize(width, height int)

    // Size 获取当前尺寸
    Size() (width, height int)
}

// EventSource 事件源接口 (用于真实环境)
type EventSource interface {
    // Events 返回事件通道
    Events() <-chan platform.RawInput

    // Start 启动事件读取
    Start() error

    // Stop 停止事件读取
    Stop() error
}

// EventSink 事件注入接口 (用于测试环境)
type EventSink interface {
    // Inject 注入单个事件
    Inject(event platform.RawInput) error

    // InjectKey 注入按键事件
    InjectKey(key rune) error

    // InjectSpecialKey 注入特殊按键
    InjectSpecialKey(key platform.SpecialKey) error

    // InjectKeyWithMod 注入带修饰符的按键
    InjectKeyWithMod(key rune, mod platform.KeyModifier) error

    // InjectMouse 注入鼠标事件
    InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error

    // InjectResize 注入窗口调整事件
    InjectResize(width, height int) error

    // InjectString 注入字符串 (转换为按键序列)
    InjectString(text string) error

    // ProcessEvents 处理所有待处理事件
    ProcessEvents() error
}

// Snapshotter 快照接口
type Snapshotter interface {
    // Snapshot 创建快照
    Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)

    // Restore 恢复快照
    Restore(snap *Snapshot) error

    // ListSnapshots 列出所有快照
    ListSnapshots() []*SnapshotMetadata
}

// TestSandbox 测试沙箱接口 (组合接口)
type TestSandbox interface {
    Sandbox
    EventSink
    Snapshotter

    // IsMock 是否为模拟沙箱
    IsMock() bool

    // AssertRender 断言渲染输出包含指定文本
    AssertRender(text string) error

    // AssertNotRender 断言渲染输出不包含指定文本
    AssertNotRender(text string) error

    // RenderString 获取渲染输出字符串
    RenderString() string

    // Helper 获取测试辅助器
    Helper() *TestHelper
}
```

## 3. 生命周期管理

### 3.1 状态机 (lifecycle.go)

```go
// sandbox/lifecycle.go

package sandbox

import (
    "fmt"
    "sync"
)

// 合法状态转换表
var validTransitions = map[State][]State{
    StateStopped:     {StateInitialized},
    StateInitialized: {StateRunning, StateStopped},
    StateRunning:     {StatePaused, StateStopped},
    StatePaused:      {StateRunning, StateStopped},
    StateError:       {StateStopped},
}

// Lifecycle 生命周期管理器
type Lifecycle struct {
    mu     sync.RWMutex
    state  State
    err    error
    hooks  map[HookKey][]HookFunc
}

// HookFunc 钩子函数类型
type HookFunc func() error

// NewLifecycle 创建生命周期管理器
func NewLifecycle() *Lifecycle {
    return &Lifecycle{
        state: StateStopped,
        hooks: make(map[HookKey][]HookFunc),
    }
}

// State 获取当前状态
func (l *Lifecycle) State() State {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.state
}

// Error 获取错误状态
func (l *Lifecycle) Error() error {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.err
}

// Transition 执行状态转移
func (l *Lifecycle) Transition(to State) error {
    l.mu.Lock()
    defer l.mu.Unlock()

    from := l.state

    // 验证状态转移是否合法
    if !l.isValidTransition(from, to) {
        return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
    }

    // 执行前置钩子
    if err := l.executeHooks(HookKey{to, PhaseBefore}); err != nil {
        l.state = StateError
        l.err = err
        return err
    }

    // 更新状态
    l.state = to

    // 执行后置钩子
    if err := l.executeHooks(HookKey{to, PhaseAfter}); err != nil {
        l.state = StateError
        l.err = err
        return err
    }

    return nil
}

// isValidTransition 检查状态转移是否合法
func (l *Lifecycle) isValidTransition(from, to State) bool {
    allowed, ok := validTransitions[from]
    if !ok {
        return false
    }
    for _, s := range allowed {
        if s == to {
            return true
        }
    }
    return false
}

// executeHooks 执行指定钩子
func (l *Lifecycle) executeHooks(key HookKey) error {
    hooks, ok := l.hooks[key]
    if !ok {
        return nil
    }
    for _, hook := range hooks {
        if err := hook(); err != nil {
            return err
        }
    }
    return nil
}

// OnTransition 注册状态转移钩子
func (l *Lifecycle) OnTransition(state State, phase Phase, fn HookFunc) {
    l.mu.Lock()
    defer l.mu.Unlock()

    key := HookKey{state, phase}
    l.hooks[key] = append(l.hooks[key], fn)
}

// Reset 重置生命周期
func (l *Lifecycle) Reset() {
    l.mu.Lock()
    defer l.mu.Unlock()

    l.state = StateStopped
    l.err = nil
}

// CanTransitionTo 检查是否可以转移到目标状态
func (l *Lifecycle) CanTransitionTo(to State) bool {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.isValidTransition(l.state, to)
}
```

## 4. 事件系统

### 4.1 事件适配器 (adapter/input.go)

```go
// sandbox/adapter/input.go

package adapter

import (
    "time"

    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/sandbox"
)

// InputAdapter 输入适配器 - 桥接 platform.RawInput 和 sandbox
type InputAdapter struct {
    reader   platform.InputReader
    eventsCh chan platform.RawInput
    stopCh   chan struct{}
}

// NewInputAdapter 创建输入适配器
func NewInputAdapter() (*InputAdapter, error) {
    reader, err := platform.NewInputReader()
    if err != nil {
        return nil, err
    }
    return &InputAdapter{
        reader:   reader,
        eventsCh: make(chan platform.RawInput, 100),
        stopCh:   make(chan struct{}),
    }, nil
}

// Start 启动输入读取
func (a *InputAdapter) Start() error {
    return a.reader.Start(a.eventsCh)
}

// Stop 停止输入读取
func (a *InputAdapter) Stop() error {
    close(a.stopCh)
    return a.reader.Stop()
}

// Events 返回事件通道
func (a *InputAdapter) Events() <-chan platform.RawInput {
    return a.eventsCh
}

// ToSandboxEvent 转换为沙箱事件
func ToSandboxEvent(raw platform.RawInput, injected bool) sandbox.InputEvent {
    return sandbox.InputEvent{
        Raw:       raw,
        Injected:  injected,
        Timestamp: time.Now(),
    }
}

// BuildKeyEvent 构建按键事件
func BuildKeyEvent(key rune) platform.RawInput {
    return platform.RawInput{
        Type:      platform.InputKeyPress,
        Key:       key,
        Timestamp: time.Now(),
    }
}

// BuildSpecialKeyEvent 构建特殊按键事件
func BuildSpecialKeyEvent(key platform.SpecialKey, mods ...platform.KeyModifier) platform.RawInput {
    var mod platform.KeyModifier
    for _, m := range mods {
        mod |= m
    }
    return platform.RawInput{
        Type:      platform.InputKeyPress,
        Special:   key,
        Modifiers: mod,
        Timestamp: time.Now(),
    }
}

// BuildMouseEvent 构建鼠标事件
func BuildMouseEvent(x, y int, button platform.MouseButton, action platform.MouseAction) platform.RawInput {
    return platform.RawInput{
        Type:        platform.InputMouse,
        MouseX:      x,
        MouseY:      y,
        MouseButton: button,
        MouseAction: action,
        Timestamp:   time.Now(),
    }
}

// BuildResizeEvent 构建窗口调整事件
func BuildResizeEvent(width, height int) platform.RawInput {
    return platform.RawInput{
        Type:      platform.InputResize,
        Width:     width,
        Height:    height,
        Timestamp: time.Now(),
    }
}

// BuildPasteEvent 构建粘贴事件
func BuildPasteEvent(text string) platform.RawInput {
    return platform.RawInput{
        Type:      platform.InputPaste,
        Data:      []byte(text),
        Timestamp: time.Now(),
    }
}
```

### 4.2 事件注入器 (events.go)

```go
// sandbox/events.go

package sandbox

import (
    "sync"

    "github.com/wwsheng009/mint/runtime/platform"
)

// EventInjector 事件注入器
type EventInjector struct {
    mu       sync.RWMutex
    strategy InjectionStrategy
    handler  EventHandler
    recorder *EventRecorder
}

// EventHandler 事件处理函数
type EventHandler func(event platform.RawInput) error

// NewEventInjector 创建事件注入器
func NewEventInjector(strategy InjectionStrategy) *EventInjector {
    return &EventInjector{
        strategy: strategy,
    }
}

// SetHandler 设置事件处理器
func (ei *EventInjector) SetHandler(handler EventHandler) {
    ei.mu.Lock()
    defer ei.mu.Unlock()
    ei.handler = handler
}

// SetRecorder 设置事件录制器
func (ei *EventInjector) SetRecorder(recorder *EventRecorder) {
    ei.mu.Lock()
    defer ei.mu.Unlock()
    ei.recorder = recorder
}

// Strategy 获取当前策略
func (ei *EventInjector) Strategy() InjectionStrategy {
    ei.mu.RLock()
    defer ei.mu.RUnlock()
    return ei.strategy
}

// SetStrategy 动态切换策略
func (ei *EventInjector) SetStrategy(strategy InjectionStrategy) {
    ei.mu.Lock()
    defer ei.mu.Unlock()
    ei.strategy = strategy
}

// Inject 注入事件 (根据策略)
func (ei *EventInjector) Inject(event platform.RawInput) error {
    ei.mu.RLock()
    strategy := ei.strategy
    handler := ei.handler
    recorder := ei.recorder
    ei.mu.RUnlock()

    switch strategy {
    case InjectProhibited:
        return ei.injectProhibited(event, recorder)

    case InjectAllowed:
        return ei.injectAllowed(event, handler, recorder)

    case InjectRecorded:
        return ei.injectRecorded(event, recorder)

    default:
        return ErrInvalidStrategy
    }
}

func (ei *EventInjector) injectProhibited(event platform.RawInput, recorder *EventRecorder) error {
    // 真实环境：仅记录，不注入
    if recorder != nil {
        recorder.Record(event)
    }
    return ErrInjectionNotAllowed
}

func (ei *EventInjector) injectAllowed(event platform.RawInput, handler EventHandler, recorder *EventRecorder) error {
    // 测试环境：记录并注入
    if recorder != nil {
        recorder.Record(event)
    }
    if handler != nil {
        return handler(event)
    }
    return nil
}

func (ei *EventInjector) injectRecorded(event platform.RawInput, recorder *EventRecorder) error {
    // 录制模式：只记录不注入
    if recorder != nil {
        return recorder.Record(event)
    }
    return nil
}

// EventRecorder 事件录制器
type EventRecorder struct {
    mu     sync.Mutex
    events []platform.RawInput
    maxLen int
}

// NewEventRecorder 创建事件录制器
func NewEventRecorder(maxLen int) *EventRecorder {
    return &EventRecorder{
        events: make([]platform.RawInput, 0, maxLen),
        maxLen: maxLen,
    }
}

// Record 录制事件
func (r *EventRecorder) Record(event platform.RawInput) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if len(r.events) >= r.maxLen {
        // 淘汰最旧的
        r.events = r.events[1:]
    }
    r.events = append(r.events, event)
    return nil
}

// Events 获取所有录制的事件
func (r *EventRecorder) Events() []platform.RawInput {
    r.mu.Lock()
    defer r.mu.Unlock()

    result := make([]platform.RawInput, len(r.events))
    copy(result, r.events)
    return result
}

// Clear 清空录制
func (r *EventRecorder) Clear() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.events = r.events[:0]
}

// Len 返回事件数量
func (r *EventRecorder) Len() int {
    r.mu.Lock()
    defer r.mu.Unlock()
    return len(r.events)
}
```

## 5. 有界事件队列

### 5.1 内存安全的事件队列 (mock/queue.go)

```go
// sandbox/mock/queue.go

package mock

import (
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/sandbox"
)

// QueueConfig 队列配置
type QueueConfig struct {
    MaxSize     int           // 最大队列长度 (默认 10000)
    MaxMemory   int64         // 最大内存占用 (字节，默认 100MB)
    EvictPolicy sandbox.EvictPolicy
}

// DefaultQueueConfig 默认配置
func DefaultQueueConfig() QueueConfig {
    return QueueConfig{
        MaxSize:     10000,
        MaxMemory:   100 * 1024 * 1024, // 100MB
        EvictPolicy: sandbox.EvictOldest,
    }
}

// BoundedQueue 有界事件队列
type BoundedQueue struct {
    mu          sync.RWMutex
    config      QueueConfig
    events      []platform.RawInput
    memory      int64
    evictCount  int64 // 已淘汰事件数
}

// NewBoundedQueue 创建有界队列
func NewBoundedQueue(config QueueConfig) *BoundedQueue {
    if config.MaxSize <= 0 {
        config.MaxSize = 10000
    }
    return &BoundedQueue{
        config: config,
        events: make([]platform.RawInput, 0, min(config.MaxSize, 1000)),
    }
}

// Push 添加事件
func (q *BoundedQueue) Push(event platform.RawInput) error {
    q.mu.Lock()
    defer q.mu.Unlock()

    eventSize := estimateEventSize(event)

    // 检查内存限制
    for q.config.MaxMemory > 0 && q.memory+eventSize > q.config.MaxMemory && len(q.events) > 0 {
        if err := q.evictOne(); err != nil {
            return err
        }
    }

    // 检查容量限制
    for len(q.events) >= q.config.MaxSize {
        if err := q.evictOne(); err != nil {
            return err
        }
    }

    q.events = append(q.events, event)
    q.memory += eventSize

    return nil
}

// Pop 取出最旧的事件
func (q *BoundedQueue) Pop() (platform.RawInput, error) {
    q.mu.Lock()
    defer q.mu.Unlock()

    if len(q.events) == 0 {
        return platform.RawInput{}, sandbox.ErrQueueEmpty
    }

    event := q.events[0]
    q.events = q.events[1:]
    q.memory -= estimateEventSize(event)

    return event, nil
}

// Peek 查看最旧的事件 (不移除)
func (q *BoundedQueue) Peek() (platform.RawInput, error) {
    q.mu.RLock()
    defer q.mu.RUnlock()

    if len(q.events) == 0 {
        return platform.RawInput{}, sandbox.ErrQueueEmpty
    }

    return q.events[0], nil
}

// Len 返回队列长度
func (q *BoundedQueue) Len() int {
    q.mu.RLock()
    defer q.mu.RUnlock()
    return len(q.events)
}

// IsEmpty 检查队列是否为空
func (q *BoundedQueue) IsEmpty() bool {
    q.mu.RLock()
    defer q.mu.RUnlock()
    return len(q.events) == 0
}

// Clear 清空队列
func (q *BoundedQueue) Clear() {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.events = q.events[:0]
    q.memory = 0
}

// evictOne 淘汰一个事件
func (q *BoundedQueue) evictOne() error {
    if len(q.events) == 0 {
        return nil
    }

    event := q.events[0]
    q.events = q.events[1:]
    q.memory -= estimateEventSize(event)
    q.evictCount++

    return nil
}

// Stats 队列统计
type QueueStats struct {
    Length      int
    MemoryUsed  int64
    MemoryLimit int64
    EvictCount  int64
}

// Stats 获取队列统计
func (q *BoundedQueue) Stats() QueueStats {
    q.mu.RLock()
    defer q.mu.RUnlock()

    return QueueStats{
        Length:      len(q.events),
        MemoryUsed:  q.memory,
        MemoryLimit: q.config.MaxMemory,
        EvictCount:  q.evictCount,
    }
}

// estimateEventSize 估算事件内存占用
func estimateEventSize(event platform.RawInput) int64 {
    // 基础结构大小
    size := int64(64) // RawInput 结构体大小估算

    // Data 字段
    if event.Data != nil {
        size += int64(len(event.Data))
    }

    return size
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## 6. 快照系统

### 6.1 快照管理 (snapshot.go)

```go
// sandbox/snapshot.go

package sandbox

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/platform"
)

// Snapshot 快照
type Snapshot struct {
    Metadata SnapshotMetadata
    Buffer   *paint.Buffer
    Events   []platform.RawInput
    State    map[string]interface{}
    Checksum string
}

// SnapshotMetadata 快照元数据
type SnapshotMetadata struct {
    ID        string
    Timestamp time.Time
    Level     SnapshotLevel
    Tags      []string
    Size      int64
}

// SnapshotManager 快照管理器
type SnapshotManager struct {
    mu        sync.RWMutex
    snapshots map[string]*Snapshot
    order     []string // 按时间顺序
    maxCount  int
    storage   SnapshotStorage
}

// SnapshotStorage 快照存储接口
type SnapshotStorage interface {
    Save(snap *Snapshot) error
    Load(id string) (*Snapshot, error)
    Delete(id string) error
    List() ([]*SnapshotMetadata, error)
}

// NewSnapshotManager 创建快照管理器
func NewSnapshotManager(maxCount int) *SnapshotManager {
    if maxCount <= 0 {
        maxCount = 100
    }
    return &SnapshotManager{
        snapshots: make(map[string]*Snapshot),
        order:     make([]string, 0),
        maxCount:  maxCount,
    }
}

// SetStorage 设置持久化存储
func (sm *SnapshotManager) SetStorage(storage SnapshotStorage) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    sm.storage = storage
}

// Create 创建快照
func (sm *SnapshotManager) Create(level SnapshotLevel, buffer *paint.Buffer, events []platform.RawInput, state map[string]interface{}, tags ...string) (*Snapshot, error) {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    snap := &Snapshot{
        Metadata: SnapshotMetadata{
            ID:        generateSnapshotID(),
            Timestamp: time.Now(),
            Level:     level,
            Tags:      tags,
        },
    }

    // 根据级别捕获不同层次的数据
    switch level {
    case SnapshotMinimal:
        snap.Buffer = cloneBuffer(buffer)

    case SnapshotStandard:
        snap.Buffer = cloneBuffer(buffer)
        snap.Events = cloneEvents(events)

    case SnapshotFull:
        snap.Buffer = cloneBuffer(buffer)
        snap.Events = cloneEvents(events)
        snap.State = cloneState(state)
    }

    // 计算校验和
    snap.Checksum = computeChecksum(snap)
    snap.Metadata.Size = estimateSnapshotSize(snap)

    // 存储到内存
    sm.snapshots[snap.Metadata.ID] = snap
    sm.order = append(sm.order, snap.Metadata.ID)

    // 淘汰旧快照
    for len(sm.order) > sm.maxCount {
        oldID := sm.order[0]
        sm.order = sm.order[1:]
        delete(sm.snapshots, oldID)
    }

    // 持久化 (如果配置了)
    if sm.storage != nil {
        if err := sm.storage.Save(snap); err != nil {
            return nil, err
        }
    }

    return snap, nil
}

// Get 获取快照
func (sm *SnapshotManager) Get(id string) (*Snapshot, error) {
    sm.mu.RLock()
    snap, ok := sm.snapshots[id]
    sm.mu.RUnlock()

    if ok {
        return snap, nil
    }

    // 尝试从存储加载
    if sm.storage != nil {
        return sm.storage.Load(id)
    }

    return nil, ErrSnapshotNotFound
}

// List 列出所有快照元数据
func (sm *SnapshotManager) List() []*SnapshotMetadata {
    sm.mu.RLock()
    defer sm.mu.RUnlock()

    result := make([]*SnapshotMetadata, 0, len(sm.snapshots))
    for _, id := range sm.order {
        if snap, ok := sm.snapshots[id]; ok {
            meta := snap.Metadata
            result = append(result, &meta)
        }
    }
    return result
}

// Delete 删除快照
func (sm *SnapshotManager) Delete(id string) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    if _, ok := sm.snapshots[id]; !ok {
        return ErrSnapshotNotFound
    }

    delete(sm.snapshots, id)

    // 从顺序列表中移除
    for i, oid := range sm.order {
        if oid == id {
            sm.order = append(sm.order[:i], sm.order[i+1:]...)
            break
        }
    }

    // 从存储删除
    if sm.storage != nil {
        return sm.storage.Delete(id)
    }

    return nil
}

// Verify 验证快照完整性
func (sm *SnapshotManager) Verify(snap *Snapshot) bool {
    return snap.Checksum == computeChecksum(snap)
}

// ==============================================================================
// Helper Functions
// ==============================================================================

func generateSnapshotID() string {
    data := time.Now().UnixNano()
    hash := sha256.Sum256([]byte(string(rune(data))))
    return hex.EncodeToString(hash[:8])
}

func computeChecksum(snap *Snapshot) string {
    data, _ := json.Marshal(struct {
        Level  SnapshotLevel
        Events int
        State  int
    }{
        Level:  snap.Metadata.Level,
        Events: len(snap.Events),
        State:  len(snap.State),
    })
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:])
}

func estimateSnapshotSize(snap *Snapshot) int64 {
    var size int64 = 100 // 基础元数据

    if snap.Buffer != nil {
        size += int64(snap.Buffer.Width * snap.Buffer.Height * 16) // 每个 Cell 约 16 字节
    }

    size += int64(len(snap.Events) * 64) // 每个事件约 64 字节

    return size
}

func cloneBuffer(buf *paint.Buffer) *paint.Buffer {
    if buf == nil {
        return nil
    }

    clone := paint.NewBuffer(buf.Width, buf.Height)
    for y := 0; y < buf.Height; y++ {
        copy(clone.Cells[y], buf.Cells[y])
    }
    return clone
}

func cloneEvents(events []platform.RawInput) []platform.RawInput {
    if events == nil {
        return nil
    }
    clone := make([]platform.RawInput, len(events))
    copy(clone, events)
    return clone
}

func cloneState(state map[string]interface{}) map[string]interface{} {
    if state == nil {
        return nil
    }
    clone := make(map[string]interface{}, len(state))
    for k, v := range state {
        clone[k] = v
    }
    return clone
}
```

## 7. 配置系统

### 7.1 沙箱配置 (config.go)

```go
// sandbox/config.go

package sandbox

import "time"

// Config 沙箱配置
type Config struct {
    // 基础配置
    Width   int
    Height  int
    Title   string
    FPS     int

    // 事件配置
    Event EventConfig

    // 快照配置
    Snapshot SnapshotConfig

    // 性能配置
    Performance PerformanceConfig
}

// EventConfig 事件配置
type EventConfig struct {
    QueueMaxSize   int              // 最大队列长度 (默认 10000)
    QueueMaxMemory int64            // 最大内存占用 (默认 100MB)
    EvictPolicy    EvictPolicy      // 淘汰策略
    Strategy       InjectionStrategy // 注入策略
    RecordEnabled  bool             // 是否启用录制
    RecordMaxLen   int              // 录制最大长度
}

// SnapshotConfig 快照配置
type SnapshotConfig struct {
    AutoSnapshot bool          // 自动快照
    Interval     time.Duration // 快照间隔
    MaxCount     int           // 最大快照数
    Level        SnapshotLevel // 默认快照级别
    PersistPath  string        // 持久化路径
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
    Throttle      bool          // 节流
    MaxFPS        int           // 最大帧率
    RenderTimeout time.Duration // 渲染超时
    Profile       bool          // 性能分析
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
    return &Config{
        Width:  80,
        Height: 24,
        FPS:    60,

        Event: EventConfig{
            QueueMaxSize:   10000,
            QueueMaxMemory: 100 * 1024 * 1024, // 100MB
            EvictPolicy:    EvictOldest,
            Strategy:       InjectAllowed,
            RecordEnabled:  false,
            RecordMaxLen:   10000,
        },

        Snapshot: SnapshotConfig{
            AutoSnapshot: false,
            MaxCount:     100,
            Level:        SnapshotStandard,
        },

        Performance: PerformanceConfig{
            Throttle:      true,
            MaxFPS:        60,
            RenderTimeout: 100 * time.Millisecond,
            Profile:       false,
        },
    }
}

// RealConfig 真实环境配置
func RealConfig() *Config {
    cfg := DefaultConfig()
    cfg.Event.Strategy = InjectProhibited
    cfg.Event.RecordEnabled = true
    return cfg
}

// MockConfig 模拟环境配置
func MockConfig() *Config {
    cfg := DefaultConfig()
    cfg.Event.Strategy = InjectAllowed
    cfg.Performance.Throttle = false // 测试时不节流
    return cfg
}

// ReplayConfig 回放环境配置
func ReplayConfig() *Config {
    cfg := DefaultConfig()
    cfg.Event.Strategy = InjectRecorded
    return cfg
}

// Validate 验证配置
func (c *Config) Validate() error {
    if c.Width <= 0 {
        c.Width = 80
    }
    if c.Height <= 0 {
        c.Height = 24
    }
    if c.FPS <= 0 {
        c.FPS = 60
    }
    if c.Event.QueueMaxSize <= 0 {
        c.Event.QueueMaxSize = 10000
    }
    return nil
}

// Clone 克隆配置
func (c *Config) Clone() *Config {
    clone := *c
    return &clone
}
```

## 8. Mock 沙箱实现

### 8.1 模拟沙箱 (mock/sandbox.go)

```go
// sandbox/mock/sandbox.go

package mock

import (
    "strings"
    "sync"
    "time"

    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/sandbox"
    "github.com/wwsheng009/mint/sandbox/adapter"
)

// MockSandbox 模拟沙箱
type MockSandbox struct {
    mu sync.RWMutex

    lifecycle *sandbox.Lifecycle
    config    *sandbox.Config
    buffer    *paint.Buffer

    // 事件系统
    injector *sandbox.EventInjector
    queue    *BoundedQueue
    recorder *sandbox.EventRecorder

    // 快照
    snapMgr *sandbox.SnapshotManager

    // 事件处理
    eventHandler sandbox.EventHandler
}

// New 创建模拟沙箱
func New(width, height int) *MockSandbox {
    config := sandbox.MockConfig()
    config.Width = width
    config.Height = height

    ms := &MockSandbox{
        lifecycle: sandbox.NewLifecycle(),
        config:    config,
        buffer:    paint.NewBuffer(width, height),
        injector:  sandbox.NewEventInjector(sandbox.InjectAllowed),
        queue:     NewBoundedQueue(DefaultQueueConfig()),
        recorder:  sandbox.NewEventRecorder(config.Event.RecordMaxLen),
        snapMgr:   sandbox.NewSnapshotManager(config.Snapshot.MaxCount),
    }

    ms.injector.SetRecorder(ms.recorder)

    return ms
}

// ==============================================================================
// Sandbox Interface
// ==============================================================================

func (ms *MockSandbox) Initialize(config *sandbox.Config) error {
    ms.mu.Lock()
    defer ms.mu.Unlock()

    if config != nil {
        ms.config = config
        ms.buffer = paint.NewBuffer(config.Width, config.Height)
    }

    return ms.lifecycle.Transition(sandbox.StateInitialized)
}

func (ms *MockSandbox) Run() error {
    return ms.lifecycle.Transition(sandbox.StateRunning)
}

func (ms *MockSandbox) Pause() error {
    return ms.lifecycle.Transition(sandbox.StatePaused)
}

func (ms *MockSandbox) Resume() error {
    return ms.lifecycle.Transition(sandbox.StateRunning)
}

func (ms *MockSandbox) Close() error {
    return ms.lifecycle.Transition(sandbox.StateStopped)
}

func (ms *MockSandbox) State() sandbox.State {
    return ms.lifecycle.State()
}

func (ms *MockSandbox) Type() sandbox.SandboxType {
    return sandbox.TypeMock
}

func (ms *MockSandbox) Config() *sandbox.Config {
    ms.mu.RLock()
    defer ms.mu.RUnlock()
    return ms.config
}

func (ms *MockSandbox) Buffer() *paint.Buffer {
    ms.mu.RLock()
    defer ms.mu.RUnlock()
    return ms.buffer
}

func (ms *MockSandbox) SetBuffer(buf *paint.Buffer) {
    ms.mu.Lock()
    defer ms.mu.Unlock()
    ms.buffer = buf
}

func (ms *MockSandbox) Resize(width, height int) {
    ms.mu.Lock()
    defer ms.mu.Unlock()
    ms.buffer = paint.NewBuffer(width, height)
    ms.config.Width = width
    ms.config.Height = height
}

func (ms *MockSandbox) Size() (int, int) {
    ms.mu.RLock()
    defer ms.mu.RUnlock()
    return ms.config.Width, ms.config.Height
}

// ==============================================================================
// EventSink Interface
// ==============================================================================

func (ms *MockSandbox) SetEventHandler(handler sandbox.EventHandler) {
    ms.mu.Lock()
    defer ms.mu.Unlock()
    ms.eventHandler = handler
    ms.injector.SetHandler(handler)
}

func (ms *MockSandbox) Inject(event platform.RawInput) error {
    if err := ms.queue.Push(event); err != nil {
        return err
    }
    return ms.injector.Inject(event)
}

func (ms *MockSandbox) InjectKey(key rune) error {
    event := adapter.BuildKeyEvent(key)
    return ms.Inject(event)
}

func (ms *MockSandbox) InjectSpecialKey(key platform.SpecialKey) error {
    event := adapter.BuildSpecialKeyEvent(key)
    return ms.Inject(event)
}

func (ms *MockSandbox) InjectKeyWithMod(key rune, mod platform.KeyModifier) error {
    event := platform.RawInput{
        Type:      platform.InputKeyPress,
        Key:       key,
        Modifiers: mod,
        Timestamp: time.Now(),
    }
    return ms.Inject(event)
}

func (ms *MockSandbox) InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error {
    event := adapter.BuildMouseEvent(x, y, button, action)
    return ms.Inject(event)
}

func (ms *MockSandbox) InjectResize(width, height int) error {
    event := adapter.BuildResizeEvent(width, height)
    return ms.Inject(event)
}

func (ms *MockSandbox) InjectString(text string) error {
    for _, r := range text {
        if err := ms.InjectKey(r); err != nil {
            return err
        }
    }
    return nil
}

func (ms *MockSandbox) ProcessEvents() error {
    for !ms.queue.IsEmpty() {
        event, err := ms.queue.Pop()
        if err != nil {
            break
        }
        if ms.eventHandler != nil {
            if err := ms.eventHandler(event); err != nil {
                return err
            }
        }
    }
    return nil
}

// ==============================================================================
// Snapshotter Interface
// ==============================================================================

func (ms *MockSandbox) Snapshot(level sandbox.SnapshotLevel, tags ...string) (*sandbox.Snapshot, error) {
    ms.mu.RLock()
    defer ms.mu.RUnlock()

    return ms.snapMgr.Create(level, ms.buffer, ms.recorder.Events(), nil, tags...)
}

func (ms *MockSandbox) Restore(snap *sandbox.Snapshot) error {
    ms.mu.Lock()
    defer ms.mu.Unlock()

    if snap.Buffer != nil {
        ms.buffer = paint.NewBuffer(snap.Buffer.Width, snap.Buffer.Height)
        for y := 0; y < snap.Buffer.Height; y++ {
            copy(ms.buffer.Cells[y], snap.Buffer.Cells[y])
        }
    }

    return nil
}

func (ms *MockSandbox) ListSnapshots() []*sandbox.SnapshotMetadata {
    return ms.snapMgr.List()
}

// ==============================================================================
// TestSandbox Interface
// ==============================================================================

func (ms *MockSandbox) IsMock() bool {
    return true
}

func (ms *MockSandbox) AssertRender(text string) error {
    rendered := ms.RenderString()
    if !strings.Contains(rendered, text) {
        return &sandbox.AssertionError{
            Message:  "render does not contain expected text",
            Expected: text,
            Actual:   rendered,
        }
    }
    return nil
}

func (ms *MockSandbox) AssertNotRender(text string) error {
    rendered := ms.RenderString()
    if strings.Contains(rendered, text) {
        return &sandbox.AssertionError{
            Message:  "render contains unexpected text",
            Expected: "not " + text,
            Actual:   rendered,
        }
    }
    return nil
}

func (ms *MockSandbox) RenderString() string {
    ms.mu.RLock()
    defer ms.mu.RUnlock()

    if ms.buffer == nil {
        return ""
    }

    var sb strings.Builder
    for y := 0; y < ms.buffer.Height; y++ {
        for x := 0; x < ms.buffer.Width; x++ {
            cell := ms.buffer.Cells[y][x]
            if cell.IsContinuation {
                continue
            }
            if cell.Cluster == "" {
                sb.WriteRune(' ')
            } else {
                sb.WriteString(cell.Cluster)
            }
        }
        if y < ms.buffer.Height-1 {
            sb.WriteRune('\n')
        }
    }
    return sb.String()
}

func (ms *MockSandbox) Helper() *TestHelper {
    return NewTestHelper(ms)
}

// QueueStats 返回队列统计
func (ms *MockSandbox) QueueStats() QueueStats {
    return ms.queue.Stats()
}
```

## 9. 测试辅助API

### 9.1 测试辅助器 (mock/testapi.go)

```go
// sandbox/mock/testapi.go

package mock

import (
    "time"

    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/sandbox"
)

// TestHelper 测试辅助器
type TestHelper struct {
    sandbox *MockSandbox
    errors  []error
}

// NewTestHelper 创建测试辅助器
func NewTestHelper(sb *MockSandbox) *TestHelper {
    return &TestHelper{
        sandbox: sb,
        errors:  make([]error, 0),
    }
}

// Errors 返回所有错误
func (th *TestHelper) Errors() []error {
    return th.errors
}

// HasErrors 检查是否有错误
func (th *TestHelper) HasErrors() bool {
    return len(th.errors) > 0
}

// ClearErrors 清除错误
func (th *TestHelper) ClearErrors() {
    th.errors = th.errors[:0]
}

// ==============================================================================
// Action Methods (链式调用，返回 *TestHelper)
// ==============================================================================

// Type 输入文本
func (th *TestHelper) Type(text string) *TestHelper {
    if err := th.sandbox.InjectString(text); err != nil {
        th.errors = append(th.errors, err)
    }
    return th
}

// Press 按下按键
func (th *TestHelper) Press(key platform.SpecialKey) *TestHelper {
    if err := th.sandbox.InjectSpecialKey(key); err != nil {
        th.errors = append(th.errors, err)
    }
    return th
}

// PressKey 按下字符键
func (th *TestHelper) PressKey(key rune) *TestHelper {
    if err := th.sandbox.InjectKey(key); err != nil {
        th.errors = append(th.errors, err)
    }
    return th
}

// Click 点击
func (th *TestHelper) Click(x, y int) *TestHelper {
    if err := th.sandbox.InjectMouse(x, y, platform.MouseLeft, platform.MousePress); err != nil {
        th.errors = append(th.errors, err)
    }
    return th
}

// Tab 按 Tab 键
func (th *TestHelper) Tab() *TestHelper {
    return th.Press(platform.KeyTab)
}

// Enter 按 Enter 键
func (th *TestHelper) Enter() *TestHelper {
    return th.Press(platform.KeyEnter)
}

// Escape 按 Escape 键
func (th *TestHelper) Escape() *TestHelper {
    return th.Press(platform.KeyEscape)
}

// Process 处理所有事件
func (th *TestHelper) Process() *TestHelper {
    if err := th.sandbox.ProcessEvents(); err != nil {
        th.errors = append(th.errors, err)
    }
    return th
}

// Wait 等待一段时间
func (th *TestHelper) Wait(d time.Duration) *TestHelper {
    time.Sleep(d)
    return th
}

// ==============================================================================
// Assertion Methods
// ==============================================================================

// AssertRender 断言渲染包含文本
func (th *TestHelper) AssertRender(text string) *TestHelper {
    if err := th.sandbox.AssertRender(text); err != nil {
        th.errors = append(th.errors, err)
    }
    return th
}

// AssertNotRender 断言渲染不包含文本
func (th *TestHelper) AssertNotRender(text string) *TestHelper {
    if err := th.sandbox.AssertNotRender(text); err != nil {
        th.errors = append(th.errors, err)
    }
    return th
}

// ==============================================================================
// Result Method
// ==============================================================================

// Result 返回测试结果
type TestResult struct {
    Errors []error
}

// Result 完成链式调用并返回结果
func (th *TestHelper) Result() TestResult {
    return TestResult{
        Errors: th.errors,
    }
}

// OK 检查是否成功 (无错误)
func (r TestResult) OK() bool {
    return len(r.Errors) == 0
}

// Error 返回第一个错误
func (r TestResult) Error() error {
    if len(r.Errors) == 0 {
        return nil
    }
    return r.Errors[0]
}
```

## 10. 真实沙箱实现

### 10.1 真实环境沙箱 (real/sandbox.go)

```go
// sandbox/real/sandbox.go

package real

import (
    "sync"

    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/sandbox"
    "github.com/wwsheng009/mint/sandbox/adapter"
)

// RealSandbox 真实终端沙箱
type RealSandbox struct {
    mu sync.RWMutex

    lifecycle *sandbox.Lifecycle
    config    *sandbox.Config
    buffer    *paint.Buffer

    // 输入适配器
    input *adapter.InputAdapter

    // 事件系统
    injector *sandbox.EventInjector
    recorder *sandbox.EventRecorder

    // 快照
    snapMgr *sandbox.SnapshotManager

    // 停止信号
    stopCh chan struct{}
}

// New 创建真实沙箱
func New(width, height int) (*RealSandbox, error) {
    config := sandbox.RealConfig()
    config.Width = width
    config.Height = height

    input, err := adapter.NewInputAdapter()
    if err != nil {
        return nil, err
    }

    rs := &RealSandbox{
        lifecycle: sandbox.NewLifecycle(),
        config:    config,
        buffer:    paint.NewBuffer(width, height),
        input:     input,
        injector:  sandbox.NewEventInjector(sandbox.InjectProhibited),
        recorder:  sandbox.NewEventRecorder(config.Event.RecordMaxLen),
        snapMgr:   sandbox.NewSnapshotManager(config.Snapshot.MaxCount),
        stopCh:    make(chan struct{}),
    }

    rs.injector.SetRecorder(rs.recorder)

    return rs, nil
}

// ==============================================================================
// Sandbox Interface
// ==============================================================================

func (rs *RealSandbox) Initialize(config *sandbox.Config) error {
    rs.mu.Lock()
    defer rs.mu.Unlock()

    if config != nil {
        rs.config = config
        rs.buffer = paint.NewBuffer(config.Width, config.Height)
    }

    return rs.lifecycle.Transition(sandbox.StateInitialized)
}

func (rs *RealSandbox) Run() error {
    if err := rs.lifecycle.Transition(sandbox.StateRunning); err != nil {
        return err
    }

    // 启动输入读取
    if err := rs.input.Start(); err != nil {
        return err
    }

    // 事件循环
    go rs.eventLoop()

    return nil
}

func (rs *RealSandbox) eventLoop() {
    for {
        select {
        case <-rs.stopCh:
            return
        case event := <-rs.input.Events():
            rs.handleEvent(event)
        }
    }
}

func (rs *RealSandbox) handleEvent(event platform.RawInput) {
    // 录制事件
    rs.recorder.Record(event)

    // 处理窗口调整
    if event.Type == platform.InputResize {
        rs.Resize(event.Width, event.Height)
    }
}

func (rs *RealSandbox) Pause() error {
    return rs.lifecycle.Transition(sandbox.StatePaused)
}

func (rs *RealSandbox) Resume() error {
    return rs.lifecycle.Transition(sandbox.StateRunning)
}

func (rs *RealSandbox) Close() error {
    close(rs.stopCh)
    rs.input.Stop()
    platform.RestoreTerminal()
    return rs.lifecycle.Transition(sandbox.StateStopped)
}

func (rs *RealSandbox) State() sandbox.State {
    return rs.lifecycle.State()
}

func (rs *RealSandbox) Type() sandbox.SandboxType {
    return sandbox.TypeReal
}

func (rs *RealSandbox) Config() *sandbox.Config {
    rs.mu.RLock()
    defer rs.mu.RUnlock()
    return rs.config
}

func (rs *RealSandbox) Buffer() *paint.Buffer {
    rs.mu.RLock()
    defer rs.mu.RUnlock()
    return rs.buffer
}

func (rs *RealSandbox) SetBuffer(buf *paint.Buffer) {
    rs.mu.Lock()
    defer rs.mu.Unlock()
    rs.buffer = buf
}

func (rs *RealSandbox) Resize(width, height int) {
    rs.mu.Lock()
    defer rs.mu.Unlock()
    rs.buffer = paint.NewBuffer(width, height)
    rs.config.Width = width
    rs.config.Height = height
}

func (rs *RealSandbox) Size() (int, int) {
    rs.mu.RLock()
    defer rs.mu.RUnlock()
    return rs.config.Width, rs.config.Height
}

// ==============================================================================
// EventSource Interface
// ==============================================================================

func (rs *RealSandbox) Events() <-chan platform.RawInput {
    return rs.input.Events()
}

// ==============================================================================
// Snapshotter Interface
// ==============================================================================

func (rs *RealSandbox) Snapshot(level sandbox.SnapshotLevel, tags ...string) (*sandbox.Snapshot, error) {
    rs.mu.RLock()
    defer rs.mu.RUnlock()

    return rs.snapMgr.Create(level, rs.buffer, rs.recorder.Events(), nil, tags...)
}

func (rs *RealSandbox) Restore(snap *sandbox.Snapshot) error {
    rs.mu.Lock()
    defer rs.mu.Unlock()

    // 真实环境软恢复：重新渲染
    if snap.Buffer != nil {
        rs.buffer = paint.NewBuffer(snap.Buffer.Width, snap.Buffer.Height)
        for y := 0; y < snap.Buffer.Height; y++ {
            copy(rs.buffer.Cells[y], snap.Buffer.Cells[y])
        }
    }

    return nil
}

func (rs *RealSandbox) ListSnapshots() []*sandbox.SnapshotMetadata {
    return rs.snapMgr.List()
}

// RecordedEvents 获取录制的事件
func (rs *RealSandbox) RecordedEvents() []platform.RawInput {
    return rs.recorder.Events()
}
```

## 11. UI 层集成

### 11.1 测试入口 (ui/test.go)

```go
// ui/test.go

package ui

import (
    "github.com/wwsheng009/mint/sandbox"
    "github.com/wwsheng009/mint/sandbox/mock"
)

// TestApp 测试应用包装器
type TestApp struct {
    sandbox *mock.MockSandbox
    app     interface{} // 实际应用
}

// TestRun 运行测试应用
func TestRun(app interface{}, opts ...TestOption) (*TestApp, error) {
    config := &testConfig{
        width:  80,
        height: 24,
    }

    for _, opt := range opts {
        opt(config)
    }

    sb := mock.New(config.width, config.height)

    if err := sb.Initialize(nil); err != nil {
        return nil, err
    }

    testApp := &TestApp{
        sandbox: sb,
        app:     app,
    }

    return testApp, nil
}

// TestRunWithConfig 使用自定义配置运行测试应用
func TestRunWithConfig(app interface{}, config *sandbox.Config) (*TestApp, error) {
    sb := mock.New(config.Width, config.Height)

    if err := sb.Initialize(config); err != nil {
        return nil, err
    }

    return &TestApp{
        sandbox: sb,
        app:     app,
    }, nil
}

// Close 关闭测试应用
func (ta *TestApp) Close() error {
    return ta.sandbox.Close()
}

// Sandbox 获取沙箱
func (ta *TestApp) Sandbox() sandbox.TestSandbox {
    return ta.sandbox
}

// Helper 获取测试辅助器
func (ta *TestApp) Helper() *mock.TestHelper {
    return ta.sandbox.Helper()
}

// ==============================================================================
// Test Options
// ==============================================================================

type testConfig struct {
    width  int
    height int
}

type TestOption func(*testConfig)

func WithWidth(w int) TestOption {
    return func(c *testConfig) {
        c.width = w
    }
}

func WithHeight(h int) TestOption {
    return func(c *testConfig) {
        c.height = h
    }
}

func WithSize(w, h int) TestOption {
    return func(c *testConfig) {
        c.width = w
        c.height = h
    }
}
```

## 12. 测试示例

### 12.1 基础测试

```go
package ui_test

import (
    "testing"

    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/sandbox"
    "github.com/wwsheng009/mint/ui"
)

func TestLoginForm(t *testing.T) {
    testApp, err := ui.TestRun(LoginForm, ui.WithSize(80, 24))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    helper := testApp.Helper()

    // 链式测试
    result := helper.
        Type("user@example.com").
        Tab().
        Type("secret").
        Enter().
        Process().
        AssertRender("Welcome").
        Result()

    if !result.OK() {
        t.Error(result.Error())
    }
}

func TestWithSnapshot(t *testing.T) {
    testApp, _ := ui.TestRun(MyComponent)
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 创建快照
    snap, err := sb.Snapshot(sandbox.SnapshotStandard)
    if err != nil {
        t.Fatal(err)
    }

    // 执行操作
    sb.InjectString("test")
    sb.ProcessEvents()

    // 恢复快照
    if err := sb.Restore(snap); err != nil {
        t.Error(err)
    }
}

func TestMemoryConstrained(t *testing.T) {
    config := sandbox.MockConfig()
    config.Event.QueueMaxMemory = 10 * 1024 * 1024 // 10MB

    testApp, _ := ui.TestRunWithConfig(MyComponent, config)
    defer testApp.Close()

    sb := testApp.Sandbox().(*mock.MockSandbox)

    // 注入大量事件
    for i := 0; i < 100000; i++ {
        sb.InjectKey('a')
    }

    // 检查队列统计
    stats := sb.QueueStats()
    if stats.MemoryUsed > stats.MemoryLimit {
        t.Errorf("memory limit exceeded: %d > %d",
            stats.MemoryUsed, stats.MemoryLimit)
    }
}
```

## 13. 实施计划

### 阶段1: 核心类型和接口 (1天)
- [ ] 创建 `sandbox/` 目录结构
- [ ] 实现 `types.go`, `errors.go`
- [ ] 实现 `sandbox.go` 核心接口
- [ ] 实现 `lifecycle.go`
- [ ] 实现 `config.go`

### 阶段2: 事件系统 (1-2天)
- [ ] 实现 `adapter/input.go`
- [ ] 实现 `events.go` 事件注入器
- [ ] 实现 `mock/queue.go` 有界队列
- [ ] 添加单元测试

### 阶段3: Mock 沙箱 (2天)
- [ ] 实现 `mock/sandbox.go`
- [ ] 实现 `mock/testapi.go`
- [ ] 添加测试

### 阶段4: 快照系统 (1-2天)
- [ ] 实现 `snapshot.go`
- [ ] 添加测试

### 阶段5: 真实沙箱 (1-2天)
- [ ] 实现 `real/sandbox.go`
- [ ] 与现有 Engine 集成

### 阶段6: UI 层集成 (1天)
- [ ] 实现 `ui/test.go`
- [ ] 添加示例测试

### 阶段7: 回放系统 (2天)
- [ ] 实现 `replay/sandbox.go`
- [ ] 实现 `replay/player.go`
- [ ] 实现 `replay/recorder.go`

**总计: 9-12天**

## 14. 验收标准

### 功能验收
- [ ] Mock 沙箱事件注入正常
- [ ] 快照创建和恢复正常
- [ ] 内存限制生效
- [ ] 链式测试 API 可用

### 兼容性验收
- [ ] 与 `runtime/event` 完全兼容
- [ ] 与 `runtime/platform` 完全兼容
- [ ] 与 `runtime/paint.Buffer` 完全兼容
- [ ] 与 `runtime/scheduler` 无冲突

### 性能验收
- [ ] Mock 沙箱内存占用可控
- [ ] 事件队列支持 10000+ 事件
- [ ] 快照操作 < 100ms

### 测试验收
- [ ] 单元测试覆盖率 > 80%
- [ ] 所有示例测试通过
