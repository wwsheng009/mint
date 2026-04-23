# Engine 与 Sandbox 集成方案

## 1. 当前架构分析

### 1.1 完整应用运行流程 (ui.Run)

```
ui.Run()
  → framework.NewApp()
    → framework/event.Pump
      → platform.InputReader (真实终端)
        → RawInput 通道
```

**说明**: `ui.Run()` 用于生产环境，从真实终端读取输入。

### 1.2 测试应用运行流程 (ui.RunTest)

```
ui.RunTest()
  → framework.NewApp()
    → framework/event.Pump
      → EventSource 接口 (可插拔事件源)
        → ChannelEventSource / PlatformEventSource
          → RawInput 通道
            → Inject() 方法注入事件
```

**说明**: `ui.RunTest()` 用于测试环境，支持事件注入。

### 1.3 依赖关系图

```
┌─────────────────────────────────────────────────────────────────┐
│                        应用层 (ui/)                              │
├──────────────────────────────┬──────────────────────────────────┤
│         ui.Run               │         ui.RunTest              │
│   (完整应用，阻塞)            │      (测试模式，可注入)           │
└──────────────┬───────────────┴──────────────┬───────────────────┘
               │                              │
               ▼                              ▼
┌──────────────────────────┐     ┌──────────────────────────┐
│     framework.App        │     │      TestableApp         │
│  (使用 Pump 读取终端)      │     │  (使用 Pump + Inject)    │
└──────────────┬───────────┘     └──────────────┬───────────┘
               │                               │
               ▼                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                    framework/event/Pump                            │
│                                                                    │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    EventSource 接口                          ││
│  │  Start() (<-chan RawInput, error)                           ││
│  │  Stop() error                                               ││
│  └─────────────────────────────────────────────────────────────┘│
│           │                           │                           │
│  ┌────────┴────────┐        ┌────────┴────────┐        ┌───────┴───────┐
│  │  PlatformEvent  │        │  ChannelEvent  │        │  SandboxEvent │
│  │  Source         │        │  Source         │        │  Source       │
│  │  (生产环境)      │        │  (测试环境)      │        │  (未使用)      │
│  └─────────────────┘        └─────────────────┘        └───────────────┘
│         │                           │                           │
│         ▼                           ▼                           ▼
│  platform.InputReader         测试通道                    MockSandbox
│  (真实终端输入)               Inject()                    (未集成)
└───────────────────────────────────────────────────────────────────┘
```

---

## 2. EventSource 接口模式

### 2.1 接口定义

```go
// framework/event/pump.go

// EventSource 事件源接口
// Pump 从 EventSource 读取原始输入，转换为框架事件
type EventSource interface {
    // Start 启动事件源，返回事件通道
    Start() (<-chan platform.RawInput, error)

    // Stop 停止事件源
    Stop() error
}
```

### 2.2 Pump 结构

```go
type Pump struct {
    source EventSource  // 可插拔事件源 (而非硬编码 InputReader)
    events chan Event
    quit   chan struct{}
    running bool
}
```

### 2.3 创建 Pump

```go
// 方式 1: 使用 platform.InputReader (生产环境)
func NewPump(reader platform.InputReader) *Pump {
    return &Pump{
        source: &PlatformEventSource{reader: reader},
        events: make(chan Event, 100),
        quit:   make(chan struct{}),
    }
}

// 方式 2: 使用自定义 EventSource (测试/扩展)
func NewPumpWithSource(source EventSource) *Pump {
    return &Pump{
        source: source,
        events: make(chan Event, 100),
        quit:   make(chan struct{}),
    }
}
```

---

## 3. EventSource 实现

### 3.1 PlatformEventSource (生产环境)

```go
// PlatformEventSource 包装 platform.InputReader 为 EventSource
type PlatformEventSource struct {
    reader     platform.InputReader
    rawInputs  chan platform.RawInput
}

func (s *PlatformEventSource) Start() (<-chan platform.RawInput, error) {
    s.rawInputs = make(chan platform.RawInput, 50)
    if err := s.reader.Start(s.rawInputs); err != nil {
        return nil, err
    }
    return s.rawInputs, nil
}

func (s *PlatformEventSource) Stop() error {
    if s.reader != nil {
        return s.reader.Stop()
    }
    return nil
}
```

### 3.2 ChannelEventSource (测试环境)

```go
// ChannelEventSource 直接从通道读取事件 (最简单的 EventSource)
type ChannelEventSource struct {
    ch chan platform.RawInput
}

func NewChannelEventSource(ch chan platform.RawInput) *ChannelEventSource {
    return &ChannelEventSource{ch: ch}
}

func (s *ChannelEventSource) Start() (<-chan platform.RawInput, error) {
    return s.ch, nil
}

func (s *ChannelEventSource) Stop() error {
    return nil  // 无操作
}
```

### 3.3 SandboxEventSource (未使用)

```go
// SandboxEventSource 将 MockSandbox 适配为 EventSource
type SandboxEventSource struct {
    sandbox    *mock.MockSandbox
    rawInputs  chan platform.RawInput
}

func NewSandboxEventSource(sb *mock.MockSandbox) *SandboxEventSource {
    return &SandboxEventSource{
        sandbox:   sb,
        rawInputs: make(chan platform.RawInput, 100),
    }
}

func (s *SandboxEventSource) Start() (<-chan platform.RawInput, error) {
    // 设置 MockSandbox 的事件处理器
    s.sandbox.SetEventHandler(func(raw platform.RawInput) error {
        select {
        case s.rawInputs <- raw:
        default:
        }
        return nil
    })
    return s.rawInputs, nil
}

func (s *SandboxEventSource) Stop() error {
    return nil
}
```

**状态**: `NewSandboxEventSource` 已实现但未使用。当前 `ui.RunTest()` 直接使用 `Pump.Inject()` 方法。

---

## 4. 事件注入机制

### 4.1 Pump.Inject() - 直接注入

```go
// framework/event/pump.go

// Inject 用于测试时直接注入事件到 Pump
// 注意：此方法仅用于测试，不应用于生产代码
func (p *Pump) Inject(raw platform.RawInput) {
    if !p.running {
        if os.Getenv("TUI_DEBUG_UI") == "true" {
            fmt.Fprintf(os.Stderr, "[PUMP] Inject: pump not running!\n")
        }
        return
    }
    ev := p.convertToEvent(raw)
    if ev != nil {
        if os.Getenv("TUI_DEBUG_UI") == "true" {
            fmt.Fprintf(os.Stderr, "[PUMP] Injecting event: Type=%v\n", ev.Type())
        }
        select {
        case p.events <- ev:
            if os.Getenv("TUI_DEBUG_UI") == "true" {
                fmt.Fprintf(os.Stderr, "[PUMP] Event sent to channel\n")
            }
        case <-p.quit:
        }
    }
}
```

### 4.2 framework.App.InjectEvent()

```go
// framework/app.go

// InjectEvent 用于测试时注入事件
// 注意：此方法仅用于测试，不应用于生产代码
func (a *App) InjectEvent(raw platform.RawInput) error {
    if a.pump == nil {
        return errors.New("event pump not initialized")
    }
    if !a.pump.IsRunning() {
        return errors.New("event pump not running")
    }
    a.pump.Inject(raw)
    return nil
}
```

### 4.3 TestableApp 事件注入接口

```go
// ui/app.go

// InjectKey 注入字符键
func (ta *TestableApp) InjectKey(key rune) error {
    raw := platform.RawInput{
        Type: platform.InputKeyPress,
        Key:  key,
    }
    return ta.fwApp.InjectEvent(raw)
}

// InjectSpecialKey 注入特殊键
func (ta *TestableApp) InjectSpecialKey(key platform.SpecialKey) error {
    raw := platform.RawInput{
        Type:    platform.InputKeyPress,
        Special: key,
    }
    return ta.fwApp.InjectEvent(raw)
}
```

---

## 5. RunTest 实现

### 5.1 TestableApp 结构

```go
// ui/app.go

type TestableApp struct {
    fwApp *framework.App   // 框架应用
    root  *declarativeRoot // 根组件
    opts  *Options          // 配置
}
```

### 5.2 RunTest 函数

```go
// RunTest 运行可测试的应用
func RunTest(app ComponentFunc, opts ...Option) (*TestableApp, error) {
    // 应用默认配置
    options := &Options{
        Width:  80,
        Height: 24,
        Title:  "Mint UI Test",
        FPS:    60,
    }

    // 应用用户配置
    for _, opt := range opts {
        opt(options)
    }

    // 创建框架应用
    fwApp := framework.NewApp()
    fwApp.Resize(options.Width, options.Height)
    fwApp.InitTheme("dark")

    // 创建声明式根组件
    declarativeNode := newDeclarativeRoot(app, fwApp)
    declarativeRoot, ok := declarativeNode.(*declarativeRoot)
    if !ok {
        return nil, fmt.Errorf("failed to get declarativeRoot")
    }

    fwApp.SetRoot(declarativeNode)

    // 在后台运行 (Run 会调用 Init)
    go func() {
        fwApp.Run()
    }()

    // 等待应用启动
    for i := 0; i < 100; i++ {
        if fwApp.GetState() == framework.StateRunning {
            break
        }
        time.Sleep(10 * time.Millisecond)
    }

    return &TestableApp{
        fwApp: fwApp,
        root:  declarativeRoot,
        opts:  options,
    }, nil
}
```

---

## 6. 使用示例

### 6.1 基本测试

```go
func TestBasic(t *testing.T) {
    testApp, err := ui.RunTest(MyComponent,
        ui.WithWidth(40),
        ui.WithHeight(12),
    )
    if err != nil {
        t.Fatalf("RunTest failed: %v", err)
    }
    defer testApp.Close()

    // 等待初始化
    time.Sleep(100 * time.Millisecond)

    // 获取渲染结果
    rendered := testApp.GetRenderString()
    t.Logf("Rendered:\n%s", rendered)
}
```

### 6.2 事件注入测试

```go
func TestButtonClick(t *testing.T) {
    testApp, _ := ui.RunTest(CounterComponent)
    defer testApp.Close()

    time.Sleep(100 * time.Millisecond)

    // 按 Tab 切换焦点到按钮
    testApp.InjectSpecialKey(platform.KeyTab)
    time.Sleep(50 * time.Millisecond)

    // 按 Enter 触发点击
    testApp.InjectSpecialKey(platform.KeyEnter)
    time.Sleep(100 * time.Millisecond)

    // 验证状态更新
    rendered := testApp.GetRenderString()
    if strings.Contains(rendered, "Count: 1") {
        t.Log("✅ Button click works!")
    }
}
```

### 6.3 文本输入测试

```go
func TestTextInput(t *testing.T) {
    testApp, _ := ui.RunTest(FormComponent)
    defer testApp.Close()

    // 切换到输入框
    testApp.InjectSpecialKey(platform.KeyTab)
    time.Sleep(50 * time.Millisecond)

    // 输入文本
    testApp.InjectKey('H')
    testApp.InjectKey('e')
    testApp.InjectKey('l')
    testApp.InjectKey('l')
    testApp.InjectKey('o')
    time.Sleep(50 * time.Millisecond)

    // 验证
    rendered := testApp.GetRenderString()
    if strings.Contains(rendered, "Hello") {
        t.Log("✅ Text input works!")
    }
}
```

---

## 7. 架构设计原则

### 7.1 依赖方向

```
✅ 正确:
sandbox → runtime/platform (底层包)
ui → framework (上层框架)
ui → sandbox (可选，用于测试)

❌ 错误:
sandbox → engine (会产生循环引用)
framework → sandbox (保持框架纯净)
```

### 7.2 接口隔离

```go
// framework 定义 EventSource 接口
type EventSource interface {
    Start() (<-chan platform.RawInput, error)
    Stop() error
}

// 不同环境提供不同实现
// - 生产环境: PlatformEventSource
// - 测试环境: ChannelEventSource / SandboxEventSource
```

### 7.3 测试隔离

```go
// 测试专用方法，明确标记
func (a *App) InjectEvent(raw platform.RawInput) error {
    // 注意：此方法仅用于测试
    // ...
}

func (p *Pump) Inject(raw platform.RawInput) {
    // 注意：此方法仅用于测试
    // ...
}
```

---

## 8. 当前状态

| 组件 | 状态 | 说明 |
|------|------|------|
| EventSource 接口 | ✅ 完成 | 支持可插拔事件源 |
| PlatformEventSource | ✅ 完成 | 生产环境实现 |
| ChannelEventSource | ✅ 完成 | 测试环境实现 |
| SandboxEventSource | ✅ 完成 | 已集成，支持事件录制/回放 |
| Pump.Inject() | ✅ 完成 | 直接注入事件，带并发安全保护 |
| framework.App.InjectEvent() | ✅ 完成 | 应用层注入接口 |
| framework.NewAppWithSource() | ✅ 完成 | 支持自定义 EventSource |
| ui.RunTest() | ✅ 完成 | 测试模式入口（直接注入） |
| ui.RunTestWithSandbox() | ✅ 完成 | 测试模式入口（Sandbox 模式） |
| TestableApp.GetSandbox() | ✅ 完成 | 获取 MockSandbox 实例 |
| 集成测试 | ✅ 完成 | fiber_counter 测试通过 |
| Fiber 集成 | ✅ 完成 | Fiber 模式事件处理正常 |
| Pump 并发安全 | ✅ 完成 | 添加 RWMutex 保护 events 通道 |

### 8.1 并发安全修复 (v1.1)

**问题**: Pump 关闭时，convertLoop 可能仍在尝试向 events 通道发送，导致 "send on closed channel" panic。

**解决方案**: 添加 `sync.RWMutex` 保护 events 通道的并发访问：

```go
// framework/event/pump.go
type Pump struct {
    source EventSource
    events chan Event
    quit   chan struct{}
    running bool
    mu     sync.RWMutex // 保护 events 通道
}

// convertLoop - 获取读锁后发送
func (p *Pump) convertLoop(rawInputs <-chan platform.RawInput) {
    // ...
    p.mu.RLock()
    select {
    case p.events <- ev:
        p.mu.RUnlock()
    case <-p.quit:
        p.mu.RUnlock()
        return
    }
}

// Stop - 获取写锁后关闭
func (p *Pump) Stop() {
    // ...
    p.mu.Lock()
    close(p.events)
    p.mu.Unlock()
}

// Inject - 获取读锁后发送
func (p *Pump) Inject(raw platform.RawInput) {
    // ...
    p.mu.RLock()
    select {
    case p.events <- ev:
        p.mu.RUnlock()
    case <-p.quit:
        p.mu.RUnlock()
    }
}
```

---

## 9. SandboxEventSource 集成 (v1.1)

**状态**: ✅ 完成

`SandboxEventSource` 现已完全集成，支持通过 MockSandbox 进行事件注入。

### 9.1 使用方式

```go
// 创建使用 Sandbox 事件源的测试应用
testApp, err := ui.RunTestWithSandbox(MyComponent,
    ui.WithWidth(80),
    ui.WithHeight(24),
)
defer testApp.Close()

// 获取 MockSandbox
sb := testApp.GetSandbox()

// 通过 Sandbox 注入事件（会转发到 Pump）
sb.InjectSpecialKey(platform.KeyTab)
sb.InjectSpecialKey(platform.KeyEnter)

// 直接注入也支持
testApp.InjectKey('a')
```

### 9.2 数据流

```
TestableApp.GetSandbox()
    ↓
mock.MockSandbox.InjectSpecialKey()
    ↓
MockSandbox.EventHandler 回调
    ↓
SandboxEventSource.rawInputs 通道
    ↓
Pump.convertLoop()
    ↓
Pump.events 通道
    ↓
framework.App 事件处理
```

---

## 10. 待改进项

```go
// 利用 MockSandbox 的录制功能
func TestRecordAndReplay(t *testing.T) {
    // 录制
    recorder := sandbox.NewEventRecorder()
    sb := mock.NewWithRecorder(width, height, recorder)

    // 用户操作...
    // 保存录制

    // 回放
    replay := sandbox.NewEventReplayer(recordedEvents)
    sb := mock.NewWithReplayer(width, height, replay)
    // ...
}
```

---

## 11. 参考资料

- [Fiber 集成修复报告](./FIBER_INTEGRATION_FIX_REPORT.md)
- [Sandbox 调试技巧指南](./SANDBOX_DEBUG_GUIDE.md)
- [Sandbox 设计文档](./SANDBOX_DESIGN_V3.md)
- [SandboxEventSource 集成指南](./SANDBOX_EVENT_SOURCE_INTEGRATION.md)
