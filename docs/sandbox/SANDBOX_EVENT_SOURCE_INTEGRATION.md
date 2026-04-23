# SandboxEventSource 集成指南

> 版本: 1.1 - 已完成集成
>
> `SandboxEventSource` 现已完全集成到测试框架中

## 1. 当前状态

`SandboxEventSource` 已实现并完全集成。它作为适配器连接：
- **MockSandbox** (回调模式) → **SandboxEventSource** (通道模式) → **Pump**

```go
// ui/sandbox_source.go (已实现)

type SandboxEventSource struct {
    sandbox    *mock.MockSandbox
    rawInputs  chan platform.RawInput
}

func (s *SandboxEventSource) Start() (<-chan platform.RawInput, error) {
    // 设置 EventHandler：当 MockSandbox.Inject() 被调用时
    // 事件会被转发到 rawInputs 通道
    s.sandbox.SetEventHandler(func(raw platform.RawInput) error {
        select {
        case s.rawInputs <- raw:
        default:
        }
        return nil
    })

    return s.rawInputs, nil
}
```

---

## 2. 集成架构

### 当前 (未使用 SandboxEventSource)

```
ui.RunTest()
  → framework.App
    → Pump (使用 PlatformEventSource)
      → platform.InputReader
      → Inject() 直接注入
```

### 使用 SandboxEventSource 后

```
ui.RunTestWithSandbox()
  → framework.App
    → Pump (使用 SandboxEventSource)
      → MockSandbox
        → EventHandler 回调
          → rawInputs 通道
            → Pump 读取
```

---

## 3. 实现方案

### 方案 A: 添加 RunTestWithSandbox 函数

```go
// ui/app.go

// RunTestWithSandbox 使用 MockSandbox 作为事件源进行测试
// 这允许使用 MockSandbox 的丰富功能，如事件录制、回放等
func RunTestWithSandbox(app ComponentFunc, opts ...Option) (*TestableApp, error) {
    options := &Options{
        Width:  80,
        Height: 24,
        Title:  "Mint UI Test (Sandbox)",
        FPS:    60,
    }

    for _, opt := range opts {
        opt(options)
    }

    // 创建 MockSandbox
    sb := mock.New(options.Width, options.Height)
    if err := sb.Initialize(nil); err != nil {
        return nil, fmt.Errorf("sandbox init failed: %w", err)
    }

    // 创建 SandboxEventSource
    source := NewSandboxEventSource(sb)

    // 创建使用自定义 EventSource 的 framework.App
    fwApp := framework.NewAppWithSource(source)
    fwApp.Resize(options.Width, options.Height)
    fwApp.InitTheme("dark")

    // 创建声明式根组件
    declarativeNode := newDeclarativeRoot(app, fwApp)
    declarativeRoot, ok := declarativeNode.(*declarativeRoot)
    if !ok {
        return nil, fmt.Errorf("failed to get declarativeRoot")
    }

    fwApp.SetRoot(declarativeRoot)

    // 在后台运行
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
        fwApp:   fwApp,
        root:    declarativeRoot,
        opts:    options,
        sandbox: sb, // 保存 MockSandbox 引用
    }, nil
}
```

### 方案 B: 扩展 TestableApp 添加 Sandbox 支持

```go
// ui/app.go

// TestableApp 可测试的应用包装器
type TestableApp struct {
    fwApp   *framework.App
    root    *declarativeRoot
    opts    *Options
    sandbox *mock.MockSandbox // 可选：使用 Sandbox 模式
}

// InjectViaSandbox 通过 MockSandbox 注入事件
// 仅当使用 RunTestWithSandbox 创建时可用
func (ta *TestableApp) InjectViaSandbox(key rune) error {
    if ta.sandbox == nil {
        return errors.New("sandbox not initialized, use RunTestWithSandbox")
    }
    return ta.sandbox.InjectKey(key)
}

// InjectSpecialKeyViaSandbox 通过 MockSandbox 注入特殊键
func (ta *TestableApp) InjectSpecialKeyViaSandbox(key platform.SpecialKey) error {
    if ta.sandbox == nil {
        return errors.New("sandbox not initialized, use RunTestWithSandbox")
    }
    return ta.sandbox.InjectSpecialKey(key)
}

// GetSandbox 获取 MockSandbox (用于事件录制/回放)
func (ta *TestableApp) GetSandbox() *mock.MockSandbox {
    return ta.sandbox
}
```

---

## 4. 使用示例

### 基本使用

```go
func TestWithSandbox(t *testing.T) {
    testApp, err := ui.RunTestWithSandbox(MyComponent,
        ui.WithWidth(40),
        ui.WithHeight(12),
    )
    if err != nil {
        t.Fatalf("RunTestWithSandbox failed: %v", err)
    }
    defer testApp.Close()

    // 两种注入方式都可以工作
    testApp.InjectKey('a')           // 直接注入到 Pump
    testApp.InjectViaSandbox('b')    // 通过 Sandbox 注入

    time.Sleep(50 * time.Millisecond)

    rendered := testApp.GetRenderString()
    // ...
}
```

### 事件录制与回放

```go
func TestRecordAndReplay(t *testing.T) {
    testApp, _ := ui.RunTestWithSandbox(MyComponent)
    defer testApp.Close()

    // 获取 MockSandbox
    sb := testApp.GetSandbox()

    // 创建录制器
    recorder := sandbox.NewEventRecorder()
    sb.SetRecorder(recorder)

    // 执行操作...
    testApp.InjectViaSandbox('a')
    testApp.InjectViaSandbox('b')

    // 保存录制
    events := recorder.GetEvents()

    // 回放
    sb2 := mock.New(40, 12)
    sb2.SetReplayer(sandbox.NewEventReplayer(events))
    // ...
}
```

---

## 5. 已完成的文件修改

### 5.1 framework/app.go - 添加 NewAppWithSource (✅ 完成)

```go
// framework/app.go

// App 结构体添加 customSource 字段
type App struct {
    // ... 其他字段
    customSource frameworkevent.EventSource // 自定义事件源（测试时使用）
}

// NewAppWithSource 创建使用自定义 EventSource 的应用
func NewAppWithSource(source frameworkevent.EventSource) *App {
    return &App{
        // ... 标准初始化
        customSource: source, // 使用自定义事件源
    }
}

// Init() 方法支持自定义 EventSource
func (a *App) Init() error {
    // ...
    if a.customSource != nil {
        a.pump = frameworkevent.NewPumpWithSource(a.customSource)
    } else {
        inputReader, err := platform.NewInputReader()
        // ...
        a.pump = frameworkevent.NewPump(inputReader)
    }
    // ...
}
```

### 5.2 ui/test.go - 添加 RunTestWithSandbox (✅ 完成)

```go
// ui/test.go

import (
    "github.com/wwsheng009/mint/sandbox/mock"
)

// TestableApp 结构体添加 sandbox 字段
type TestableApp struct {
    fwApp   *framework.App
    root    *declarativeRoot
    opts    *Options
    sandbox *mock.MockSandbox // 可选：使用 MockSandbox 作为事件源
}

// RunTestWithSandbox 使用 MockSandbox 作为事件源进行测试
func RunTestWithSandbox(app ComponentFunc, opts ...Option) (*TestableApp, error) {
    options := &Options{Width: 80, Height: 24, Title: "Mint UI Test (Sandbox)", FPS: 60}

    // 创建 MockSandbox
    sb := mock.New(options.Width, options.Height)
    if err := sb.Initialize(nil); err != nil {
        return nil, fmt.Errorf("sandbox init failed: %w", err)
    }

    // 创建 SandboxEventSource
    source := NewSandboxEventSource(sb)

    // 创建使用自定义 EventSource 的 framework.App
    fwApp := framework.NewAppWithSource(source)
    fwApp.Resize(options.Width, options.Height)
    fwApp.InitTheme("dark")

    // 创建声明式根组件并运行
    // ...

    return &TestableApp{
        fwApp:   fwApp,
        root:    declarativeRoot,
        opts:    options,
        sandbox: sb, // 保存 MockSandbox 引用
    }, nil
}

// GetSandbox 获取 MockSandbox
func (ta *TestableApp) GetSandbox() *mock.MockSandbox {
    return ta.sandbox
}
```

### 5.3 framework/event/pump.go - 并发安全修复 (✅ 完成)

```go
// framework/event/pump.go

type Pump struct {
    source EventSource
    events chan Event
    quit   chan struct{}
    running bool
    mu     sync.RWMutex // 保护 events 通道
}

// convertLoop 发送前获取读锁
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

// Stop 关闭前获取写锁
func (p *Pump) Stop() {
    // ...
    p.mu.Lock()
    close(p.events)
    p.mu.Unlock()
}
```

---

## 6. 优势对比

| 特性 | 直接注入 (当前) | SandboxEventSource |
|------|-----------------|-------------------|
| 事件注入 | ✅ Pump.Inject() | ✅ MockSandbox.Inject() |
| 事件录制 | ❌ | ✅ MockSandbox.Recorder |
| 事件回放 | ❌ | ✅ MockSandbox.Replayer |
| 事件队列 | ❌ | ✅ MockSandbox.Queue |
| 调试追踪 | 基础 | 丰富 |
| 复杂度 | 简单 | 较复杂 |

---

## 7. 推荐使用场景

### 使用 RunTest (推荐 - 默认)
- 简单的集成测试
- 按钮点击、文本输入
- 不需要高级功能

```go
testApp, err := ui.RunTest(MyComponent, ui.WithWidth(80), ui.WithHeight(24))
defer testApp.Close()
```

### 使用 RunTestWithSandbox (可选 - 高级功能)
- 需要事件录制/回放
- 需要事件队列调试
- 需要复杂的交互场景测试
- 测试事件时序问题

```go
testApp, err := ui.RunTestWithSandbox(MyComponent, ui.WithWidth(80), ui.WithHeight(24))
defer testApp.Close()

sb := testApp.GetSandbox()
stats := sb.QueueStats() // 访问队列统计
```

---

## 8. 实现步骤 (已完成)

### ✅ 第一步：添加 framework.NewAppWithSource (已完成)

### ✅ 第二步：实现 RunTestWithSandbox (已完成)

### ✅ 第三步：扩展 TestableApp (已完成)

### ✅ 第四步：编写测试 (已完成)

---

## 9. 测试验证 (已完成)
    fwApp := framework.NewAppWithSource(source)

    // ... 后续代码与 RunTest 相同
}
```

### 第三步：扩展 TestableApp

```go
// 添加 sandbox 字段
type TestableApp struct {
    // ... 现有字段
    sandbox *mock.MockSandbox
}

// 添加 Sandbox 相关方法
func (ta *TestableApp) InjectViaSandbox(key rune) error { ... }
func (ta *TestableApp) GetSandbox() *mock.MockSandbox { ... }
```

### 第四步：编写测试

```go
func TestSandboxIntegration(t *testing.T) {
    testApp, _ := ui.RunTestWithSandbox(MyComponent)
    defer testApp.Close()

    // 使用 Sandbox 注入
    sb := testApp.GetSandbox()
    sb.InjectKey('a')
    sb.InjectSpecialKey(platform.KeyEnter)

    // 验证
    time.Sleep(100 * time.Millisecond)
    rendered := testApp.GetRenderString()
    // ...
}
```

---

## 9. 数据流图

```
┌─────────────────────────────────────────────────────────────────┐
│                      Test                                   │
└────────────────────────────┬────────────────────────────────┘
                             │
        ┌────────────────────┴──────────────────────┐
        │                                         │
        ▼                                         ▼
┌──────────────────┐                   ┌──────────────────────┐
│  RunTest          │                   │  RunTestWithSandbox  │
└─────────┬─────────┘                   └──────────┬───────────┘
          │                                         │
          ▼                                         ▼
   ┌────────────────┐                       ┌─────────────────────┐
   │ framework.App  │                       │  framework.App       │
   │ (默认 Pump)    │                       │  (自定义 Pump)       │
   └────────┬────────┘                       └──────────┬───────────┘
            │                                         │
            ▼                                         ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │                      Pump                                         │
   │  ┌────────────────────┬─────────────────────┬─────────────────┐   │
   │  │  PlatformEvent    │  ChannelEvent        │  SandboxEvent   │   │
   │  │  Source           │  Source               │  Source          │   │
   │  └──────────┬─────────┘  └──────┬──────────────┘  └────┬────────────┘   │
   │             │                    │                     │              │
   │             ▼                    ▼                     ▼              │
   │    platform.InputReader   (测试通道)          MockSandbox          │
   │    (真实终端)                                 (事件录制/回放)        │
   └──────────────────────────────────────────────────────────────────┘
```

---

## 10. 总结

**SandboxEventSource 的价值**：
1. 提供了统一的事件源接口
2. 支持 MockSandbox 的高级功能（录制、回放）
3. 保持了架构的清晰性

**是否需要集成**：
- 如果只需要基本的测试功能 → 当前 `RunTest` 已足够
- 如果需要事件录制/回放 → 集成 `SandboxEventSource` 有价值
- 可以作为可选功能，不强制使用


---

**集成状态**: ✅ 完成 (v1.1)

**完成日期**: 2026-02-01

**相关文件**:
- framework/app.go - 添加 NewAppWithSource()
- ui/test.go - 添加 RunTestWithSandbox() 和 GetSandbox()
- framework/event/pump.go - 并发安全修复
- sandbox/events_test.go - 测试验证
