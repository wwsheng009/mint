# 可测试输入系统设计方案

## 1. 问题分析

### 当前问题
1. 测试依赖于真实的终端输入环境
2. 无法在单元测试中模拟键盘/鼠标事件
3. 示例程序使用 `fmt.Scanln()` 等阻塞调用，不适合自动化测试

### 现有架构
```
┌─────────────────────────────────────────────────────────────────┐
│                         应用层 (ui/)                            │
│  - ui.Run()                                                     │
│  - ui.Button, ui.Input 等组件                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Framework层 (framework/)                   │
│  - framework/app.go: Run(), Init(), handleEvent()               │
│  - event/pump.go: 事件泵                                        │
│  - event/router.go: 事件路由                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Platform层 (runtime/platform/)            │
│  - input_windows.go / input_unix.go / input_darwin.go           │
│  - 直接读取系统输入 (syscalls, console API)                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          操作系统                               │
│  - Windows: Win32 Console API                                  │
│  - Unix: tty, termios                                          │
└─────────────────────────────────────────────────────────────────┘
```

## 2. 设计方案

### 2.1 目标
1. 创建输入抽象层，允许注入模拟事件
2. 保持现有API不变，向后兼容
3. 支持无头模式（Headless Mode）运行
4. 提供测试友好的事件注入API

### 2.2 新架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         应用层 (ui/)                            │
│  - ui.Run() → 新增: ui.RunTest()                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Framework层 (framework/)                   │
│  - app.go: 支持可配置的输入源                                    │
│  - testing/mock_input.go: 新增模拟输入实现                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              输入抽象层 (runtime/input/)                        │
│  - input_source.go: 新增 - 定义输入源接口                       │
│  - real_input_source.go: 新增 - 真实输入源                      │
│  - mock_input_source.go: 新增 - 模拟输入源                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Platform层 (runtime/platform/)            │
│  - input_*.go: 保持现有实现                                     │
└─────────────────────────────────────────────────────────────────┘
```

## 3. 详细设计

### 3.1 输入源接口 (`runtime/input/input_source.go`)

```go
package input

import (
    "github.com/wwsheng009/mint/framework/event"
)

// InputSource 输入源接口
type InputSource interface {
    // Start 启动输入源
    Start() error

    // Stop 停止输入源
    Stop() error

    // Events 返回事件通道
    Events() <-chan event.Event

    // InjectEvent 注入事件（用于测试）
    InjectEvent(ev event.Event) error

    // IsMock 是否为模拟输入源
    IsMock() bool
}
```

### 3.2 真实输入源 (`runtime/input/real_input_source.go`)

```go
package input

import (
    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/framework/event"
    frameworkevent "github.com/wwsheng009/mint/framework/event"
)

// RealInputSource 真实输入源 - 包装现有的平台输入
type RealInputSource struct {
    reader    platform.InputReader
    eventChan chan event.Event
    stopChan  chan struct{}
}

func NewRealInputSource() (*RealInputSource, error) {
    reader, err := platform.NewInputReader()
    if err != nil {
        return nil, err
    }

    return &RealInputSource{
        reader:    reader,
        eventChan: make(chan event.Event, 100),
        stopChan:  make(chan struct{}),
    }, nil
}

func (r *RealInputSource) Start() error {
    return r.reader.Start(func(raw platform.RawInput) {
        ev := convertRawInputToEvent(raw)
        r.eventChan <- ev
    })
}

func (r *RealInputSource) Stop() error {
    close(r.stopChan)
    return r.reader.Stop()
}

func (r *RealInputSource) Events() <-chan event.Event {
    return r.eventChan
}

func (r *RealInputSource) InjectEvent(ev event.Event) error {
    // 真实输入源不支持注入
    return ErrNotSupported
}

func (r *RealInputSource) IsMock() bool {
    return false
}
```

### 3.3 模拟输入源 (`runtime/input/mock_input_source.go`)

```go
package input

import (
    "sync"
    "time"

    "github.com/wwsheng009/mint/framework/event"
)

// MockInputSource 模拟输入源 - 用于测试
type MockInputSource struct {
    eventChan   chan event.Event
    injectChan  chan event.Event
    stopChan    chan struct{}
    started     bool
    mu          sync.Mutex

    // 可选：记录接收到的所有事件
    recordedEvents []event.Event
    recordEnabled  bool
}

func NewMockInputSource(bufferSize int) *MockInputSource {
    return &MockInputSource{
        eventChan:   make(chan event.Event, bufferSize),
        injectChan:  make(chan event.Event, bufferSize),
        stopChan:    make(chan struct{}),
        recordedEvents: make([]event.Event, 0),
    }
}

// Start 启动模拟输入源
func (m *MockInputSource) Start() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.started {
        return nil
    }

    m.started = true

    // 启动事件转发goroutine
    go m.forwardEvents()

    return nil
}

// forwardEvents 将注入的事件转发到事件通道
func (m *MockInputSource) forwardEvents() {
    for {
        select {
        case ev := <-m.injectChan:
            m.mu.Lock()
            if m.recordEnabled {
                m.recordedEvents = append(m.recordedEvents, ev)
            }
            m.mu.Unlock()
            m.eventChan <- ev

        case <-m.stopChan:
            return
        }
    }
}

// Stop 停止模拟输入源
func (m *MockInputSource) Stop() error {
    close(m.stopChan)
    close(m.eventChan)
    close(m.injectChan)
    return nil
}

// Events 返回事件通道
func (m *MockInputSource) Events() <-chan event.Event {
    return m.eventChan
}

// InjectEvent 注入事件
func (m *MockInputSource) InjectEvent(ev event.Event) error {
    select {
    case m.injectChan <- ev:
        return nil
    case <-time.After(5 * time.Second):
        return ErrTimeout
    }
}

// InjectKey 注入按键事件（便捷方法）
func (m *MockInputSource) InjectKey(key rune) error {
    return m.InjectEvent(event.NewKeyEvent(event.Key{Rune: key}))
}

// InjectSpecialKey 注入特殊按键
func (m *MockInputSource) InjectSpecialKey(sk event.SpecialKey) error {
    ev := event.NewKeyEvent(event.Key{Name: sk.String()})
    ev.Special = sk
    return m.InjectEvent(ev)
}

// InjectMouse 注入鼠标事件
func (m *MockInputSource) InjectMouse(x, y int, button event.MouseButton, eventType event.EventType) error {
    return m.InjectEvent(event.NewMouseEvent(x, y, button, eventType))
}

// IsMock 标识为模拟输入源
func (m *MockInputSource) IsMock() bool {
    return true
}

// EnableRecording 启用事件记录
func (m *MockInputSource) EnableRecording(enable bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.recordEnabled = enable
}

// GetRecordedEvents 获取记录的事件
func (m *MockInputSource) GetRecordedEvents() []event.Event {
    m.mu.Lock()
    defer m.mu.Unlock()

    result := make([]event.Event, len(m.recordedEvents))
    copy(result, m.recordedEvents)
    return result
}

// ClearRecordedEvents 清空记录
func (m *MockInputSource) ClearRecordedEvents() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.recordedEvents = m.recordedEvents[:0]
}
```

### 3.4 Framework层改造 (`framework/app.go`)

```go
// App 应用程序
type App struct {
    // ... 现有字段 ...

    // 新增：输入源（可配置）
    inputSource input.InputSource
}

// NewApp 创建新应用
func NewApp() *App {
    return &App{
        // ... 现有初始化 ...
        inputSource: nil, // 延迟初始化，默认使用真实输入
    }
}

// SetInputSource 设置输入源（用于测试）
func (a *App) SetInputSource(src input.InputSource) {
    a.inputSource = src
}

// Init 初始化应用
func (a *App) Init() error {
    // 如果没有设置输入源，使用默认的真实输入
    if a.inputSource == nil {
        realSrc, err := input.NewRealInputSource()
        if err != nil {
            return err
        }
        a.inputSource = realSrc
    }

    // 启动输入源
    if err := a.inputSource.Start(); err != nil {
        return err
    }

    // ... 其余初始化代码 ...
}
```

### 3.5 UI层测试API (`ui/test.go` - 新增)

```go
package ui

import (
    "github.com/wwsheng009/mint/runtime/input"
)

// TestApp 测试应用包装器
type TestApp struct {
    *App
    mockInput *input.MockInputSource
}

// NewTestApp 创建测试应用
func NewTestApp(componentFunc ComponentFunc, opts ...Option) (*TestApp, error) {
    // 创建模拟输入源
    mockInput := input.NewMockInputSource(100)

    options := &Options{
        Width:  80,
        Height: 24,
        Title:  "Test App",
        FPS:    60,
    }

    for _, opt := range opts {
        opt(options)
    }

    // 创建framework app
    fwApp := framework.NewApp()
    fwApp.Resize(options.Width, options.Height)
    fwApp.SetInputSource(mockInput)  // 设置模拟输入

    // 创建declarative root
    declarativeRoot := newDeclarativeRoot(componentFunc, fwApp)
    fwApp.SetRoot(declarativeRoot)

    return &TestApp{
        App:       &App{fwApp: fwApp},
        mockInput: mockInput,
    }, nil
}

// InjectKey 注入按键
func (t *TestApp) InjectKey(key rune) error {
    return t.mockInput.InjectKey(key)
}

// InjectSpecialKey 注入特殊按键
func (t *TestApp) InjectSpecialKey(sk frameworkevent.SpecialKey) error {
    return t.mockInput.InjectSpecialKey(sk)
}

// InjectMouse 注入鼠标事件
func (t *TestApp) InjectMouse(x, y int, button frameworkevent.MouseButton, eventType frameworkevent.EventType) error {
    return t.mockInput.InjectMouse(x, y, button, eventType)
}

// GetRenderOutput 获取渲染输出
func (t *TestApp) GetRenderOutput() string {
    buf := t.fwApp.GetRenderer().GetBackBuffer()
    return bufferToString(buf)
}

// AssertRender 断言渲染内容
func (t *TestApp) AssertRender(contains string) error {
    output := t.GetRenderOutput()
    if !strings.Contains(output, contains) {
        return fmt.Errorf("render output does not contain %q:\n%s", contains, output)
    }
    return nil
}
```

### 3.6 测试示例 (`examples/test_button/main_test.go`)

```go
package main

import (
    "testing"

    "github.com/wwsheng009/mint/ui"
    frameworkevent "github.com/wwsheng009/mint/framework/event"
)

func TestButtonClick(t *testing.T) {
    clickCount := 0

    // 创建测试组件
    testFunc := func() ui.VNode {
        return ui.VStack(
            ui.Text("Button Test"),
            ui.ButtonBuilder("Click Me").
                OnClick(func() {
                    clickCount++
                }).
                Build(),
        )
    }

    // 创建测试应用
    testApp, err := ui.NewTestApp(testFunc,
        ui.WithWidth(30),
        ui.WithHeight(10),
    )
    if err != nil {
        t.Fatal(err)
    }

    // 初始化（不运行主循环）
    if err := testApp.Init(); err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 初始渲染
    testApp.Render()

    // 断言初始状态
    if err := testApp.AssertRender("Button Test"); err != nil {
        t.Error(err)
    }

    // 模拟点击按钮
    // 1. 模拟 Tab 键聚焦到按钮
    testApp.InjectKey('\t')  // Tab
    testApp.ProcessEvents()

    // 2. 模拟 Enter 键触发点击
    testApp.InjectSpecialKey(frameworkevent.KeyEnter)
    testApp.ProcessEvents()

    // 断言按钮被点击
    if clickCount != 1 {
        t.Errorf("expected button to be clicked once, got %d", clickCount)
    }
}

func TestInputField(t *testing.T) {
    var inputValue string

    testFunc := func() ui.VNode {
        return ui.VStack(
            ui.Text("Input Test"),
            ui.InputBuilder().
                Placeholder("Type here...").
                OnChange(func(value string) {
                    inputValue = value
                }).
                Build(),
        )
    }

    testApp, err := ui.NewTestApp(testFunc,
        ui.WithWidth(40),
        ui.WithHeight(10),
    )
    if err != nil {
        t.Fatal(err)
    }

    if err := testApp.Init(); err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    testApp.Render()

    // 模拟输入文本
    testText := "Hello"
    for _, ch := range testText {
        testApp.InjectKey(ch)
        testApp.ProcessEvents()
    }

    // 断言输入值
    if inputValue != testText {
        t.Errorf("expected input value %q, got %q", testText, inputValue)
    }

    // 断言渲染
    if err := testApp.AssertRender(testText); err != nil {
        t.Error(err)
    }
}
```

### 3.7 示例程序改造

原示例程序（阻塞等待）：
```go
func main() {
    app := ui.NewApp(MyComponent, ui.WithWidth(80), ui.WithHeight(24))
    app.Run()
    fmt.Println("Press Enter to exit...")
    fmt.Scanln()  // 阻塞调用
}
```

改造后（支持测试模式）：
```go
func main() {
    // 检测是否在测试模式
    if os.Getenv("MINT_TEST_MODE") == "true" {
        runTestMode()
        return
    }

    // 正常运行模式
    app := ui.NewApp(MyComponent, ui.WithWidth(80), ui.WithHeight(24))
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}

func runTestMode() {
    // 测试模式：不阻塞，可以自动运行
    testApp, _ := ui.NewTestApp(MyComponent)
    testApp.Init()

    // 执行测试场景...
    testApp.InjectKey('a')
    testApp.ProcessEvents()

    testApp.Close()
}
```

## 4. 实现计划

### 阶段1: 核心抽象层
1. 创建 `runtime/input/input_source.go` - 定义输入源接口
2. 创建 `runtime/input/real_input_source.go` - 包装现有平台输入
3. 创建 `runtime/input/mock_input_source.go` - 实现模拟输入

### 阶段2: Framework集成
1. 修改 `framework/app.go` 支持可配置输入源
2. 修改 `framework/event/pump.go` 使用输入源接口

### 阶段3: UI测试API
1. 创建 `ui/test.go` - 提供测试API
2. 创建测试辅助方法：`InjectKey`, `InjectMouse`, `AssertRender` 等

### 阶段4: 测试和文档
1. 为示例程序添加测试
2. 编写测试文档
3. 更新组件开发指南

## 5. 优势

1. **完全解耦**：输入处理与平台无关
2. **向后兼容**：现有代码无需修改
3. **易于测试**：可以模拟任意输入序列
4. **可记录/回放**：MockInputSource支持事件记录
5. **无头模式**：可以在CI/CD中运行GUI测试

## 6. 使用场景

### 单元测试
```go
func TestLoginForm(t *testing.T) {
    // ...
    testApp.InjectKey('u')
    testApp.InjectKey('s')
    testApp.InjectKey('e')
    testApp.InjectKey('r')
    testApp.InjectSpecialKey(KeyTab)
    testApp.InjectKey('p')
    testApp.InjectKey('w')
    testApp.InjectSpecialKey(KeyEnter)
    // 断言登录成功...
}
```

### 事件回放
```go
// 记录真实会话
recorder := NewEventRecorder()
// ... 用户交互 ...
recorder.SaveTo("session.json")

// 测试中回放
events := LoadEvents("session.json")
for _, ev := range events {
    testApp.InjectEvent(ev)
    testApp.ProcessEvents()
}
```

### 无头运行
```bash
# CI环境中运行测试
MINT_TEST_MODE=true go test ./...
```
