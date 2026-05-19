# Sandbox机制设计方案

> 当前状态：本文是 Sandbox 机制的历史设计说明。当前可执行测试入口优先使用 `ui.RunTest(...)`、`ui.RunTestWithSandbox(...)` 和 `sandbox/testing`；旧示例里的 `OnClick(func)`、`ui.ButtonBuilder` 需要按当前 Intent API 改写。

## 1. 核心概念

### 设计原则
```
┌─────────────────────────────────────────────────────────────────────────┐
│                          测试工具 / 真实环境                             │
└─────────────────────────────────────────────────────────────────────────┘
                                   │
                    ┌──────────────┼──────────────┐
                    ▼              ▼              ▼
           ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
           │ RealSandbox │ │ MockSandbox │ │ ReplaySandbox│
           │  (真实环境)  │ │  (测试环境)  │ │  (回放模式)  │
           └─────────────┘ └─────────────┘ └─────────────┘
                    │              │              │
                    └──────────────┼──────────────┘
                                   ▼
                    ┌──────────────────────────────┐
                    │       Sandbox 接口            │
                    │  - Input()                   │
                    │  - Output()                  │
                    │  - Events()                  │
                    │  - Inject(event)             │
                    └──────────────────────────────┘
                                   │
                                   ▼
                    ┌──────────────────────────────┐
                    │         应用                 │
                    │  (不知道运行在哪个Sandbox)   │
                    └──────────────────────────────┘
```

### 核心思想
1. **应用永远运行在Sandbox中**
2. **Sandbox提供统一的I/O抽象**
3. **测试时只需替换Sandbox实现**
4. **应用代码无需修改**

## 2. Sandbox接口定义

### 2.1 核心接口 (`sandbox/sandbox.go`)

```go
package sandbox

import (
    "io"
    "time"

    "github.com/wwsheng009/mint/framework/event"
    "github.com/wwsheng009/mint/runtime/paint"
)

// Sandbox 运行时沙箱接口
// 所有应用都运行在Sandbox中，通过Sandbox与外部交互
type Sandbox interface {
    // ========================================================================
    // 生命周期
    // ========================================================================

    // Init 初始化沙箱
    Init() error

    // Start 启动沙箱
    Start() error

    // Stop 停止沙箱
    Stop() error

    // Close 关闭沙箱并清理资源
    Close() error

    // ========================================================================
    // 输入
    // ========================================================================

    // Input 返回输入读取器
    // 应用通过Input读取用户输入
    Input() io.Reader

    // Events 返回事件通道
    // 框架通过Events接收键盘/鼠标事件
    Events() <-chan event.Event

    // InjectEvent 注入事件（测试用）
    // 测试工具通过InjectEvent模拟用户输入
    InjectEvent(ev event.Event) error

    // InjectKey 注入按键（便捷方法）
    InjectKey(key rune) error

    // InjectSpecialKey 注入特殊按键
    InjectSpecialKey(sk event.SpecialKey) error

    // InjectMouse 注入鼠标事件
    InjectMouse(x, y int, button event.MouseButton, eventType event.EventType) error

    // ========================================================================
    // 输出
    // ========================================================================

    // Output 返回输出写入器
    // 应用渲染的内容写入Output
    Output() io.Writer

    // GetRenderBuffer 获取渲染缓冲区
    // 测试时用于断言渲染结果
    GetRenderBuffer() *paint.Buffer

    // SetRenderBuffer 设置渲染缓冲区
    SetRenderBuffer(*paint.Buffer)

    // ========================================================================
    // 终端控制
    // ========================================================================

    // Size 返回终端尺寸
    Size() (width, height int)

    // SetSize 设置终端尺寸（测试用）
    SetSize(width, height int)

    // Clear 清屏
    Clear() error

    // ShowCursor / HideCursor 光标控制
    ShowCursor() error
    HideCursor() error

    // ========================================================================
    // 状态查询
    // ========================================================================

    // IsMock 是否为模拟沙箱
    IsMock() bool

    // Type 返回沙箱类型
    Type() SandboxType

    // ========================================================================
    // 快照功能（高级特性）
    // ========================================================================

    // Snapshot 保存当前状态快照
    Snapshot() (*Snapshot, error)

    // Restore 恢复快照
    Restore(*Snapshot) error

    // ========================================================================
    // 事件记录（用于回放）
    // ========================================================================

    // RecordEvents 启用/禁用事件记录
    RecordEvents(enable bool)

    // GetRecordedEvents 获取记录的事件
    GetRecordedEvents() []event.Event

    // ExportEvents 导出事件记录
    ExportEvents() ([]byte, error)

    // ImportEvents 导入事件记录（用于回放）
    ImportEvents(data []byte) error
}

// SandboxType 沙箱类型
type SandboxType string

const (
    SandboxTypeReal   SandboxType = "real"    // 真实环境
    SandboxTypeMock   SandboxType = "mock"    // 模拟环境（测试）
    SandboxTypeReplay SandboxType = "replay"  // 回放环境
    SandboxTypeRecord SandboxType = "record"  // 记录环境
)

// Snapshot 状态快照
type Snapshot struct {
    Time      time.Time
    Buffer    *paint.Buffer
    Events    []event.Event
    State     map[string]interface{}
    Metadata  map[string]string
}
```

### 2.2 Sandbox 核心 (`sandbox/sandbox.go`)

```go
package sandbox

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/wwsheng009/mint/framework/component"
    "github.com/wwsheng009/mint/runtime/paint"
)

// Container 应用容器
// 应用运行在Container中，Container使用Sandbox处理I/O
type Container struct {
    sb        Sandbox
    root      component.Node
    ctx       context.Context
    cancel    context.CancelFunc
    running   bool
}

// NewContainer 创建应用容器
func NewContainer(root component.Node, sb Sandbox) *Container {
    ctx, cancel := context.WithCancel(context.Background())
    return &Container{
        sb:     sb,
        root:   root,
        ctx:    ctx,
        cancel: cancel,
    }
}

// Run 运行应用（阻塞）
func (c *Container) Run() error {
    if err := c.sb.Init(); err != nil {
        return fmt.Errorf("sandbox init failed: %w", err)
    }
    if err := c.sb.Start(); err != nil {
        return fmt.Errorf("sandbox start failed: %w", err)
    }
    defer c.sb.Close()

    c.running = true

    // 主循环
    ticker := time.NewTicker(16 * time.Millisecond)
    defer ticker.Stop()

    for c.running {
        select {
        case ev := <-c.sb.Events():
            c.handleEvent(ev)

        case <-ticker.C:
            c.render()

        case <-c.ctx.Done():
            c.running = false
        }
    }

    return nil
}

// ProcessEvents 处理单个事件（测试用，非阻塞）
func (c *Container) ProcessEvents(timeout time.Duration) int {
    count := 0
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        select {
        case ev := <-c.sb.Events():
            c.handleEvent(ev)
            count++
        case <-time.After(10 * time.Millisecond):
            return count
        }
    }
    return count
}

// render 渲染界面
func (c *Container) render() {
    if paintable, ok := c.root.(component.Paintable); ok {
        buf := c.sb.GetRenderBuffer()
        w, h := c.sb.Size()
        buf.Reset(w, h)

        paintCtx := component.PaintContext{
            AvailableWidth:  w,
            AvailableHeight: h,
            X:               0,
            Y:               0,
        }

        paintable.Paint(paintCtx, buf)
    }
}

// handleEvent 处理事件
func (c *Container) handleEvent(ev event.Event) {
    if handler, ok := c.root.(event.Component); ok {
        handler.HandleEvent(ev)
    }
}

// Stop 停止应用
func (c *Container) Stop() {
    c.running = false
    c.cancel()
}

// GetSandbox 获取沙箱
func (c *Container) GetSandbox() Sandbox {
    return c.sb
}

// GetRoot 获取根组件
func (c *Container) GetRoot() component.Node {
    return c.root
}
```

## 3. Sandbox实现

### 3.1 RealSandbox (`sandbox/real/sandbox.go`)

```go
package sandbox

import (
    "fmt"
    "io"
    "os"
    "sync"

    "github.com/wwsheng009/mint/framework/event"
    "github.com/wwsheng009/mint/runtime/paint"
    "github.com/wwsheng009/mint/runtime/platform"
)

// RealSandbox 真实环境沙箱
type RealSandbox struct {
    mu            sync.RWMutex
    started       bool

    // 平台输入读取器
    inputReader   platform.InputReader

    // 事件通道
    eventChan     chan event.Event

    // 渲染缓冲区
    buffer        *paint.Buffer

    // 终端尺寸
    width         int
    height        int

    // 事件记录
    recordEnabled bool
    recordedEvents []event.Event
}

// NewRealSandbox 创建真实环境沙箱
func NewRealSandbox(width, height int) *RealSandbox {
    return &RealSandbox{
        eventChan:      make(chan event.Event, 100),
        buffer:         paint.NewBuffer(width, height),
        width:          width,
        height:         height,
        recordedEvents: make([]event.Event, 0),
    }
}

func (r *RealSandbox) Init() error {
    reader, err := platform.NewInputReader()
    if err != nil {
        return fmt.Errorf("failed to create input reader: %w", err)
    }
    r.inputReader = reader
    return nil
}

func (r *RealSandbox) Start() error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if r.started {
        return nil
    }

    if err := r.inputReader.Start(r.handleRawInput); err != nil {
        return err
    }

    // 初始化终端
    fmt.Print("\x1b[2J")      // 清屏
    fmt.Print("\x1b[?25l")    // 隐藏光标

    r.started = true
    return nil
}

func (r *RealSandbox) Stop() error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if !r.started {
        return nil
    }

    // 恢复终端
    fmt.Print("\x1b[?25h")    // 显示光标
    fmt.Print("\x1b[2J")      // 清屏
    fmt.Print("\x1b[H")       // 移动光标到左上角

    r.started = false
    return r.inputReader.Stop()
}

func (r *RealSandbox) Close() error {
    return r.Stop()
}

func (r *RealSandbox) Input() io.Reader {
    return os.Stdin
}

func (r *RealSandbox) Events() <-chan event.Event {
    return r.eventChan
}

func (r *RealSandbox) handleRawInput(raw platform.RawInput) {
    ev := convertRawToEvent(raw)
    r.mu.Lock()
    if r.recordEnabled {
        r.recordedEvents = append(r.recordedEvents, ev)
    }
    r.mu.Unlock()
    r.eventChan <- ev
}

func (r *RealSandbox) InjectEvent(ev event.Event) error {
    // 真实沙箱不支持注入
    return ErrNotSupported
}

func (r *RealSandbox) InjectKey(key rune) error {
    return ErrNotSupported
}

func (r *RealSandbox) InjectSpecialKey(sk event.SpecialKey) error {
    return ErrNotSupported
}

func (r *RealSandbox) InjectMouse(x, y int, button event.MouseButton, eventType event.EventType) error {
    return ErrNotSupported
}

func (r *RealSandbox) Output() io.Writer {
    return os.Stdout
}

func (r *RealSandbox) GetRenderBuffer() *paint.Buffer {
    return r.buffer
}

func (r *RealSandbox) SetRenderBuffer(buf *paint.Buffer) {
    r.buffer = buf
}

func (r *RealSandbox) Size() (int, int) {
    return r.width, r.height
}

func (r *RealSandbox) SetSize(width, height int) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.width = width
    r.height = height
    r.buffer.Resize(width, height)
}

func (r *RealSandbox) Clear() error {
    fmt.Print("\x1b[2J")
    fmt.Print("\x1b[H")
    return nil
}

func (r *RealSandbox) ShowCursor() error {
    fmt.Print("\x1b[?25h")
    return nil
}

func (r *RealSandbox) HideCursor() error {
    fmt.Print("\x1b[?25l")
    return nil
}

func (r *RealSandbox) IsMock() bool {
    return false
}

func (r *RealSandbox) Type() SandboxType {
    return SandboxTypeReal
}

func (r *RealSandbox) Snapshot() (*Snapshot, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    // 深拷贝缓冲区
    bufCopy := r.buffer.Clone()

    return &Snapshot{
        Time:     time.Now(),
        Buffer:   bufCopy,
        Events:   append([]event.Event{}, r.recordedEvents...),
        Metadata: map[string]string{"type": string(SandboxTypeReal)},
    }, nil
}

func (r *RealSandbox) Restore(snap *Snapshot) error {
    // 真实沙箱不支持恢复
    return ErrNotSupported
}

func (r *RealSandbox) RecordEvents(enable bool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.recordEnabled = enable
    if !enable {
        r.recordedEvents = r.recordedEvents[:0]
    }
}

func (r *RealSandbox) GetRecordedEvents() []event.Event {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return append([]event.Event{}, r.recordedEvents...)
}

func (r *RealSandbox) ExportEvents() ([]byte, error) {
    events := r.GetRecordedEvents()
    return json.Marshal(events)
}

func (r *RealSandbox) ImportEvents(data []byte) error {
    // 真实沙箱不支持导入
    return ErrNotSupported
}
```

### 3.2 MockSandbox (`sandbox/mock/sandbox.go`)

```go
package sandbox

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "sync"
    "time"

    "github.com/wwsheng009/mint/framework/event"
    "github.com/wwsheng009/mint/runtime/paint"
)

// MockSandbox 模拟沙箱（用于测试）
type MockSandbox struct {
    mu            sync.RWMutex
    started       bool

    // 事件通道
    eventChan     chan event.Event
    injectChan    chan event.Event

    // 渲染缓冲区
    buffer        *paint.Buffer

    // 输出缓冲区（捕获所有输出）
    output        *bytes.Buffer

    // 终端尺寸
    width         int
    height        int

    // 事件记录
    recordEnabled bool
    recordedEvents []event.Event

    // 快照历史
    snapshots     []*Snapshot

    // 配置选项
    autoProcess   bool  // 是否自动处理注入的事件
}

// NewMockSandbox 创建模拟沙箱
func NewMockSandbox(width, height int) *MockSandbox {
    return &MockSandbox{
        eventChan:      make(chan event.Event, 100),
        injectChan:     make(chan event.Event, 100),
        buffer:         paint.NewBuffer(width, height),
        output:         &bytes.Buffer{},
        width:          width,
        height:         height,
        recordedEvents: make([]event.Event, 0),
        snapshots:      make([]*Snapshot, 0),
        autoProcess:    true,  // 默认自动处理
    }
}

func (m *MockSandbox) Init() error {
    return nil
}

func (m *MockSandbox) Start() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.started {
        return nil
    }

    // 启动事件转发
    go m.forwardEvents()

    m.started = true
    return nil
}

func (m *MockSandbox) Stop() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.started = false
    return nil
}

func (m *MockSandbox) Close() error {
    m.Stop()
    m.mu.Lock()
    defer m.mu.Unlock()

    close(m.eventChan)
    close(m.injectChan)
    return nil
}

// forwardEvents 将注入的事件转发到事件通道
func (m *MockSandbox) forwardEvents() {
    for {
        select {
        case ev, ok := <-m.injectChan:
            if !ok {
                return
            }
            m.mu.Lock()
            if m.recordEnabled {
                m.recordedEvents = append(m.recordedEvents, ev)
            }
            m.mu.Unlock()
            m.eventChan <- ev
        }
    }
}

func (m *MockSandbox) Input() io.Reader {
    return &bytes.Buffer{}  // 模拟输入返回空读取器
}

func (m *MockSandbox) Events() <-chan event.Event {
    return m.eventChan
}

func (m *MockSandbox) InjectEvent(ev event.Event) error {
    m.mu.RLock()
    started := m.started
    m.mu.RUnlock()

    if !started {
        return ErrNotStarted
    }

    select {
    case m.injectChan <- ev:
        if m.autoProcess {
            // 短暂等待事件被处理
            time.Sleep(10 * time.Millisecond)
        }
        return nil
    case <-time.After(time.Second):
        return ErrTimeout
    }
}

func (m *MockSandbox) InjectKey(key rune) error {
    return m.InjectEvent(event.NewKeyEvent(event.Key{Rune: key}))
}

func (m *MockSandbox) InjectSpecialKey(sk event.SpecialKey) error {
    ev := event.NewKeyEvent(event.Key{Name: sk.String()})
    ev.Special = sk
    return m.InjectEvent(ev)
}

func (m *MockSandbox) InjectMouse(x, y int, button event.MouseButton, eventType event.EventType) error {
    return m.InjectEvent(event.NewMouseEvent(x, y, button, eventType))
}

// InjectSequence 注入事件序列（便捷方法）
func (m *MockSandbox) InjectSequence(events ...event.Event) error {
    for _, ev := range events {
        if err := m.InjectEvent(ev); err != nil {
            return err
        }
    }
    return nil
}

// InjectString 注入字符串（便捷方法）
func (m *MockSandbox) InjectString(s string) error {
    for _, ch := range s {
        if err := m.InjectKey(ch); err != nil {
            return err
        }
    }
    return nil
}

func (m *MockSandbox) Output() io.Writer {
    return m.output
}

func (m *MockSandbox) GetRenderBuffer() *paint.Buffer {
    return m.buffer
}

func (m *MockSandbox) SetRenderBuffer(buf *paint.Buffer) {
    m.buffer = buf
}

func (m *MockSandbox) Size() (int, int) {
    return m.width, m.height
}

func (m *MockSandbox) SetSize(width, height int) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.width = width
    m.height = height
    m.buffer = paint.NewBuffer(width, height)
}

func (m *MockSandbox) Clear() error {
    m.buffer.Reset(m.width, m.height)
    return nil
}

func (m *MockSandbox) ShowCursor() error {
    return nil  // 模拟环境不需要实际操作
}

func (m *MockSandbox) HideCursor() error {
    return nil
}

func (m *MockSandbox) IsMock() bool {
    return true
}

func (m *MockSandbox) Type() SandboxType {
    return SandboxTypeMock
}

// GetOutput 获取捕获的输出
func (m *MockSandbox) GetOutput() string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.output.String()
}

// AssertOutput 断言输出包含指定文本
func (m *MockSandbox) AssertOutput(contains string) error {
    output := m.GetOutput()
    if !contains(output, contains) {
        return fmt.Errorf("output does not contain %q\nActual:\n%s", contains, output)
    }
    return nil
}

// AssertRender 断言渲染包含指定文本
func (m *MockSandbox) AssertRender(contains string) error {
    renderOutput := m.buffer.String()
    if !contains(renderOutput, contains) {
        return fmt.Errorf("render does not contain %q\nActual:\n%s", contains, renderOutput)
    }
    return nil
}

// GetRenderText 获取渲染的文本内容
func (m *MockSandbox) GetRenderText() string {
    return m.buffer.String()
}

// Snapshot 创建快照
func (m *MockSandbox) Snapshot() (*Snapshot, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    snap := &Snapshot{
        Time:     time.Now(),
        Buffer:   m.buffer.Clone(),
        Events:   append([]event.Event{}, m.recordedEvents...),
        State:    make(map[string]interface{}),
        Metadata: map[string]string{"type": string(SandboxTypeMock)},
    }

    m.snapshots = append(m.snapshots, snap)
    return snap, nil
}

// Restore 恢复快照
func (m *MockSandbox) Restore(snap *Snapshot) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.buffer = snap.Buffer.Clone()
    m.recordedEvents = append([]event.Event{}, snap.Events...)
    return nil
}

// GetSnapshots 获取所有快照
func (m *MockSandbox) GetSnapshots() []*Snapshot {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return append([]*Snapshot{}, m.snapshots...)
}

// ClearSnapshots 清空快照
func (m *MockSandbox) ClearSnapshots() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.snapshots = m.snapshots[:0]
}

// RecordEvents 启用/禁用事件记录
func (m *MockSandbox) RecordEvents(enable bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.recordEnabled = enable
    if !enable {
        m.recordedEvents = m.recordedEvents[:0]
    }
}

// GetRecordedEvents 获取记录的事件
func (m *MockSandbox) GetRecordedEvents() []event.Event {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return append([]event.Event{}, m.recordedEvents...)
}

// ExportEvents 导出事件记录
func (m *MockSandbox) ExportEvents() ([]byte, error) {
    return json.Marshal(m.GetRecordedEvents())
}

// ImportEvents 导入事件记录
func (m *MockSandbox) ImportEvents(data []byte) error {
    var events []event.Event
    if err := json.Unmarshal(data, &events); err != nil {
        return err
    }

    m.mu.Lock()
    defer m.mu.Unlock()
    m.recordedEvents = events
    return nil
}

// SetAutoProcess 设置是否自动处理注入的事件
func (m *MockSandbox) SetAutoProcess(auto bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.autoProcess = auto
}

// DrainEvents 排空事件通道（用于测试同步）
func (m *MockSandbox) DrainEvents(timeout time.Duration) int {
    count := 0
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        select {
        case <-m.eventChan:
            count++
        case <-time.After(10 * time.Millisecond):
            return count
        }
    }
    return count
}
```

### 3.3 ReplaySandbox (`sandbox/replay/sandbox.go`)

```go
package sandbox

import (
    "encoding/json"
    "time"

    "github.com/wwsheng009/mint/framework/event"
)

// ReplaySandbox 回放沙箱
// 用于回放之前记录的事件序列
type ReplaySandbox struct {
    *MockSandbox
    events       []event.Event
    currentIndex int
    speed        float64  // 回放速度（1.0 = 正常速度）
    autoReplay   bool     // 是否自动回放
}

// NewReplaySandbox 创建回放沙箱
func NewReplaySandbox(width, height int, events []event.Event) *ReplaySandbox {
    return &ReplaySandbox{
        MockSandbox:   NewMockSandbox(width, height),
        events:        events,
        currentIndex:  0,
        speed:         1.0,
        autoReplay:    false,
    }
}

// NewReplaySandboxFromJSON 从JSON创建回放沙箱
func NewReplaySandboxFromJSON(width, height int, data []byte) (*ReplaySandbox, error) {
    var events []event.Event
    if err := json.Unmarshal(data, &events); err != nil {
        return nil, err
    }
    return NewReplaySandbox(width, height, events), nil
}

// SetSpeed 设置回放速度
func (r *ReplaySandbox) SetSpeed(speed float64) {
    r.speed = speed
}

// SetAutoReplay 设置是否自动回放
func (r *ReplaySandbox) SetAutoReplay(auto bool) {
    r.autoReplay = auto
    if auto {
        go r.autoReplayEvents()
    }
}

// autoReplayEvents 自动回放事件
func (r *ReplaySandbox) autoReplayEvents() {
    var lastTime time.Time

    for r.currentIndex < len(r.events) {
        ev := r.events[r.currentIndex]

        // 计算延迟（如果事件有时间戳）
        delay := time.Duration(0)
        if !lastTime.IsZero() {
            // 这里假设事件有时间戳信息
            // 实际实现需要在Event中添加Timestamp字段
        }

        if delay > 0 && r.speed > 0 {
            time.Sleep(time.Duration(float64(delay) / r.speed))
        }

        r.InjectEvent(ev)
        r.currentIndex++
        lastTime = time.Now()
    }
}

// ReplayNext 回放下一个事件
func (r *ReplaySandbox) ReplayNext() error {
    if r.currentIndex >= len(r.events) {
        return ErrNoMoreEvents
    }

    ev := r.events[r.currentIndex]
    r.currentIndex++
    return r.InjectEvent(ev)
}

// ReplayAll 回放所有事件
func (r *ReplaySandbox) ReplayAll() error {
    for r.currentIndex < len(r.events) {
        if err := r.ReplayNext(); err != nil {
            return err
        }
    }
    return nil
}

// Reset 重置回放位置
func (r *ReplaySandbox) Reset() {
    r.currentIndex = 0
}

// Progress 获取回放进度
func (r *ReplaySandbox) Progress() (current, total int) {
    return r.currentIndex, len(r.events)
}

func (r *ReplaySandbox) Type() SandboxType {
    return SandboxTypeReplay
}
```

## 4. 应用改造

### 4.1 UI层 (`ui/app.go` - 改造)

```go
package ui

import (
    "github.com/wwsheng009/mint/sandbox"
)

// App 应用程序
type App struct {
    container *sandbox.Container
    sandbox   sandbox.Sandbox
}

// Run 运行应用（使用RealSandbox）
func Run(app ComponentFunc, opts ...Option) error {
    options := &Options{
        Width:  80,
        Height: 24,
        Title:  "Mint UI App",
        FPS:    60,
    }
    for _, opt := range opts {
        opt(options)
    }

    // 创建真实沙箱
    sb := sandbox.NewRealSandbox(options.Width, options.Height)

    // 创建根组件
    root := createRootComponent(app)

    // 创建容器并运行
    container := sandbox.NewContainer(root, sb)
    return container.Run()
}

// TestRun 测试运行（使用MockSandbox）
func TestRun(app ComponentFunc, opts ...Option) (*TestApp, error) {
    options := &Options{
        Width:  80,
        Height: 24,
    }
    for _, opt := range opts {
        opt(options)
    }

    // 创建模拟沙箱
    sb := sandbox.NewMockSandbox(options.Width, options.Height)

    // 创建根组件
    root := createRootComponent(app)

    // 创建容器
    container := sandbox.NewContainer(root, sb)

    return &TestApp{
        container: container,
        sandbox:   sb.(*sandbox.MockSandbox),
    }, nil
}

// TestApp 测试应用
type TestApp struct {
    container *sandbox.Container
    sandbox   *sandbox.MockSandbox
}

// InjectKey 注入按键
func (t *TestApp) InjectKey(key rune) error {
    return t.sandbox.InjectKey(key)
}

// InjectString 注入字符串
func (t *TestApp) InjectString(s string) error {
    return t.sandbox.InjectString(s)
}

// InjectSpecialKey 注入特殊按键
func (t *TestApp) InjectSpecialKey(sk event.SpecialKey) error {
    return t.sandbox.InjectSpecialKey(sk)
}

// InjectMouse 注入鼠标事件
func (t *TestApp) InjectMouse(x, y int, button event.MouseButton, eventType event.EventType) error {
    return t.sandbox.InjectMouse(x, y, button, eventType)
}

// ProcessEvents 处理事件
func (t *TestApp) ProcessEvents() {
    t.container.ProcessEvents(100 * time.Millisecond)
}

// Render 渲染界面
func (t *TestApp) Render() {
    t.container.render()
}

// AssertRender 断言渲染内容
func (t *TestApp) AssertRender(contains string) error {
    return t.sandbox.AssertRender(contains)
}

// GetRenderText 获取渲染文本
func (t *TestApp) GetRenderText() string {
    return t.sandbox.GetRenderText()
}

// Snapshot 创建快照
func (t *TestApp) Snapshot() (*sandbox.Snapshot, error) {
    return t.sandbox.Snapshot()
}

// Restore 恢复快照
func (t *TestApp) Restore(snap *sandbox.Snapshot) error {
    return t.sandbox.Restore(snap)
}

// Close 关闭应用
func (t *TestApp) Close() error {
    return t.container.Stop()
}
```

## 5. 测试示例

### 5.1 基础测试

```go
package ui_test

import (
    "testing"

    "github.com/wwsheng009/mint/ui"
    frameworkevent "github.com/wwsheng009/mint/framework/event"
)

func TestButtonClick(t *testing.T) {
    clickCount := 0

    // 创建测试组件
    appFunc := func() ui.VNode {
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
    testApp, err := ui.TestRun(appFunc,
        ui.WithWidth(30),
        ui.WithHeight(10),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    // 初始渲染
    testApp.Render()

    // 断言初始状态
    if err := testApp.AssertRender("Button Test"); err != nil {
        t.Error(err)
    }

    // 模拟点击
    testApp.InjectKey('\t')  // Tab - 聚焦按钮
    testApp.ProcessEvents()

    testApp.InjectSpecialKey(frameworkevent.KeyEnter)  // Enter - 点击
    testApp.ProcessEvents()

    // 断言
    if clickCount != 1 {
        t.Errorf("expected 1 click, got %d", clickCount)
    }
}

func TestInputField(t *testing.T) {
    var inputValue string

    appFunc := func() ui.VNode {
        return ui.VStack(
            ui.Text("Input Test"),
            ui.InputBuilder().
                Placeholder("Type here...").
                OnChange(func(v string) {
                    inputValue = v
                }).
                Build(),
        )
    }

    testApp, err := ui.TestRun(appFunc, ui.WithWidth(40), ui.WithHeight(10))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    testApp.Render()

    // 输入文本
    testApp.InjectString("Hello")
    testApp.ProcessEvents()

    // 断言
    if inputValue != "Hello" {
        t.Errorf("expected 'Hello', got '%s'", inputValue)
    }

    if err := testApp.AssertRender("Hello"); err != nil {
        t.Error(err)
    }
}
```

### 5.2 快照测试

```go
func TestSnapshot(t *testing.T) {
    appFunc := func() ui.VNode {
        return ui.Text("Hello World")
    }

    testApp, _ := ui.TestRun(appFunc)
    defer testApp.Close()

    // 初始状态
    testApp.Render()
    snap1, _ := testApp.Snapshot()

    // 修改后
    testApp.InjectKey('X')
    testApp.ProcessEvents()

    // 恢复快照
    testApp.Restore(snap1)

    // 断言恢复正确
    renderText := testApp.GetRenderText()
    if !contains(renderText, "Hello World") {
        t.Error("snapshot restore failed")
    }
}
```

### 5.3 记录和回放

```go
// 记录真实会话
func RecordSession() ([]byte, error) {
    sb := sandbox.NewRealSandbox(80, 24)
    sb.RecordEvents(true)

    // ... 运行应用 ...

    return sb.ExportEvents()
}

// 回放测试
func TestReplay(t *testing.T) {
    // 加载记录的事件
    events, _ := loadRecordedEvents()

    sb := sandbox.NewReplaySandbox(80, 24, events)
    sb.SetSpeed(10.0)  // 10倍速回放

    container := sandbox.NewContainer(createApp(), sb)
    container.Run()
}
```

## 6. 目录结构

```
mint/
├── runtime/
│   └── sandbox/
│       ├── sandbox.go          # 接口定义
│       ├── container.go        # 应用容器
│       ├── real.go             # 真实环境实现
│       ├── mock.go             # 模拟环境实现
│       ├── replay.go           # 回放实现
│       ├── snapshot.go         # 快照功能
│       └── errors.go           # 错误定义
├── ui/
│   ├── app.go                  # 改造：使用Sandbox
│   └── test.go                 # 测试API
├── examples/
│   ├── test_button/
│   │   ├── main.go
│   │   └── main_test.go        # 新增测试
│   └── ...
└── docs/
    ├── SANDBOX_DESIGN.md       # 本文档
    └── TESTING_GUIDE.md        # 测试指南
```

## 7. 优势总结

| 特性 | 原方案 | Sandbox方案 |
|------|--------|-------------|
| 隔离性 | 部分 | 完全隔离 |
| 测试复杂度 | 需要修改代码 | 无需修改 |
| 快照功能 | 难以实现 | 原生支持 |
| 事件记录/回放 | 需要额外工具 | 内置支持 |
| CI/CD友好 | 依赖终端环境 | 完全无头运行 |
| 调试能力 | 有限 | 强大（快照+回放） |
| 扩展性 | 中等 | 高（易于添加新Sandbox类型） |

## 8. 使用场景

### 场景1：CI/CD测试
```bash
# 无需真实终端环境
go test ./... -short
```

### 场景2：交互式调试
```go
// 在真实会话中记录
sb.RecordEvents(true)
// ... 用户操作 ...
events := sb.GetRecordedEvents()

// 在调试器中回放
debugSandbox := sandbox.NewReplaySandbox(80, 24, events)
```

### 场景3：演示/录制
```go
// 录制演示
recordSandbox := sandbox.NewRecordSandbox(80, 24)
// ... 运行演示 ...

// 导出为视频脚本
script := recordSandbox.ExportScript()
```

### 场景4：自动化测试
```go
// API测试
testApp, _ := ui.TestRun(api)
testApp.InjectString("search query")
testApp.InjectSpecialKey(KeyEnter)
testApp.ProcessEvents()
testApp.AssertRender("results found")
```

---

# 实施计划

## 阶段1: 创建Sandbox核心接口

**文件列表:**
1. `sandbox/sandbox.go` - Sandbox接口定义
2. `sandbox/errors.go` - 错误定义
3. `sandbox/types.go` - 类型定义

**任务:**
- 定义 `Sandbox` 接口
- 定义 `SandboxType` 枚举
- 定义 `Snapshot` 结构体
- 定义错误类型 (`ErrNotSupported`, `ErrTimeout`, `ErrNotStarted` 等)

## 阶段2: 实现Sandbox类型

**文件列表:**
1. `sandbox/sandbox.go` - 应用容器
2. `sandbox/real/sandbox.go` - RealSandbox实现
3. `sandbox/mock/sandbox.go` - MockSandbox实现
4. `sandbox/replay/sandbox.go` - ReplaySandbox实现
5. `sandbox/snapshot.go` - 快照功能

**任务:**
- `Container`: 封装应用和Sandbox，提供主循环
- `RealSandbox`: 包装现有的 `platform.InputReader`
- `MockSandbox`: 实现内存模拟，支持事件注入和断言
- `ReplaySandbox`: 继承MockSandbox，添加事件回放功能
- 快照功能: Buffer克隆、状态保存/恢复

## 阶段3: 改造UI层使用Sandbox

**文件修改:**
1. `ui/app.go` - 添加 `TestRun()` 函数

**新增API:**
```go
// 正常运行（使用RealSandbox）
func Run(app ComponentFunc, opts ...Option) error

// 测试运行（使用MockSandbox）
func TestRun(app ComponentFunc, opts ...Option) (*TestApp, error)

// TestApp 测试包装器
type TestApp struct {
    container *sandbox.Container
    sandbox   *sandbox.MockSandbox
}
```

## 阶段4: 添加测试

**文件列表:**
1. `ui/components/button/button_test.go` - 按钮测试
2. `ui/layout_test.go` - 核心功能测试

**测试场景:**
- 按钮点击测试
- 输入框测试
- 键盘导航测试
- 鼠标交互测试
- 快照/恢复测试

## 阶段5: 文档

**文件列表:**
1. `docs/testing/TESTING_TOOL.md` - 测试指南（新增）

## 验证方式

### 1. 单元测试
```bash
go test ./sandbox/... -v
go test ./ui/... -v
```

### 2. 示例测试
```bash
cd examples/test_button
go test -v
```

### 3. 手动测试
```go
// 创建测试应用
testApp, _ := ui.TestRun(MyComponent)

// 注入事件
testApp.InjectString("Hello")
testApp.ProcessEvents()

// 断言
testApp.AssertRender("Hello")
```

## 关键文件路径

| 文件 | 操作 |
|------|------|
| `sandbox/sandbox.go` | 新增 |
| `sandbox/sandbox.go` | 新增 |
| `sandbox/real/sandbox.go` | 新增 |
| `sandbox/mock/sandbox.go` | 新增 |
| `sandbox/replay/sandbox.go` | 新增 |
| `sandbox/errors.go` | 新增 |
| `sandbox/types.go` | 新增 |
| `ui/app.go` | 修改（添加TestRun） |
| `ui/components/button/button_test.go` | 新增 |
| `ui/layout_test.go` | 新增 |
| `docs/testing/TESTING_TOOL.md` | 新增 |

