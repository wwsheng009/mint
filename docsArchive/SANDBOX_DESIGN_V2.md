# Sandbox机制设计方案 V2 (优化版)

## 目录结构变化

```
mint/
├── sandbox/              # 独立的沙箱模块 (新增)
│   ├── sandbox.go        # 核心接口
│   ├── lifecycle.go      # 生命周期管理 (简化)
│   ├── events.go         # 事件系统 (统一事件注入)
│   ├── buffer.go         # 渲染缓冲区管理 (内存优化)
│   ├── snapshot.go       # 快照系统 (增强)
│   ├── types.go          # 类型定义
│   ├── errors.go         # 错误定义
│   │
│   ├── real/             # 真实环境实现
│   │   ├── sandbox.go
│   │   └── input.go
│   │
│   ├── mock/             # 模拟环境实现
│   │   ├── sandbox.go
│   │   ├── events.go     # 事件队列管理
│   │   └── assertions.go # 断言功能
│   │
│   └── replay/           # 回放环境实现
│       ├── sandbox.go
│       ├── player.go     # 回放控制器
│       └── recorder.go   # 录制控制器
│
├── runtime/              # 运行时 (保持不变)
├── ui/                   # UI层 (添加测试支持)
└── docs/
```

## 1. 简化的生命周期管理

### 1.1 统一状态机

```go
// sandbox/lifecycle.go

package sandbox

import "sync"

// State 沙箱状态
type State int

const (
    StateStopped State = iota
    StateInitialized
    StateRunning
    StatePaused
    StateError
)

// Lifecycle 生命周期管理器
type Lifecycle struct {
    mu     sync.RWMutex
    state  State
    err    error
    hooks map[State][]func() error
}

// Transition 状态转移
func (l *Lifecycle) Transition(to State) error {
    l.mu.Lock()
    defer l.mu.Unlock()

    // 验证状态转移是否合法
    if !l.isValidTransition(l.state, to) {
        return ErrInvalidTransition
    }

    // 执行前置钩子
    if err := l.executeHooks(to, true); err != nil {
        l.state = StateError
        l.err = err
        return err
    }

    from := l.state
    l.state = to

    // 执行后置钩子
    if err := l.executeHooks(to, false); err != nil {
        l.state = StateError
        l.err = err
        return err
    }

    return nil
}

// isValidTransition 验证状态转移
func (l *Lifecycle) isValidTransition(from, to State) bool {
    validTransitions := map[State][]State{
        StateStopped:      {StateInitialized},
        StateInitialized:  {StateRunning, StateStopped},
        StateRunning:      {StatePaused, StateStopped},
        StatePaused:       {StateRunning, StateStopped},
        StateError:        {StateStopped},
    }
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

// OnRegister 注册状态钩子
func (l *Lifecycle) OnRegister(state State, phase Phase, fn func() error) {
    if l.hooks == nil {
        l.hooks = make(map[State][]func() error)
    }
    key := hookKey{state, phase}
    l.hooks[key] = append(l.hooks[key], fn)
}
```

### 1.2 简化的Sandbox接口

```go
// sandbox/sandbox.go

package sandbox

// Sandbox 沙箱接口 (简化版)
type Sandbox interface {
    // ========================================================================
    // 核心方法 (简化)
    // ========================================================================

    // Initialize 初始化沙箱 (合并Init+配置)
    Initialize(config *Config) error

    // Run 运行沙箱 (合并Start+主循环)
    Run() error

    // Pause 暂停沙箱
    Pause() error

    // Resume 恢复沙箱
    Resume() error

    // Close 关闭沙箱 (合并Stop+清理)
    Close() error

    // ========================================================================
    // 状态查询
    // ========================================================================

    // State 获取当前状态
    State() State

    // Type 获取沙箱类型
    Type() SandboxType

    // IsMock 是否为模拟沙箱
    IsMock() bool

    // Config 获取配置
    Config() *Config
}
```

## 2. 统一的事件注入系统

### 2.1 事件注入策略

```go
// sandbox/events.go

package sandbox

import (
    "github.com/wwsheng009/mint/framework/event"
)

// InjectionStrategy 事件注入策略
type InjectionStrategy int

const (
    // InjectStrategyProhibited 禁止注入 (真实环境默认)
    InjectStrategyProhibited InjectionStrategy = iota

    // InjectStrategyAllowed 允许注入 (测试环境)
    InjectStrategyAllowed

    // InjectStrategyRecorded 仅记录不注入 (录制模式)
    InjectStrategyRecorded
)

// EventInjector 事件注入器
type EventInjector struct {
    strategy InjectionStrategy
    handler  EventHandler
    recorder *EventRecorder
}

// EventHandler 处理注入的事件
type EventHandler func(ev event.Event) error

// Inject 注入事件 (根据策略)
func (ei *EventInjector) Inject(ev event.Event) error {
    switch ei.strategy {
    case InjectStrategyProhibited:
        return ei.injectProhibited(ev)

    case InjectStrategyAllowed:
        return ei.injectAllowed(ev)

    case InjectStrategyRecorded:
        return ei.injectRecorded(ev)

    default:
        return ErrInvalidStrategy
    }
}

func (ei *EventInjector) injectProhibited(ev event.Event) error {
    // 真实环境：记录但不注入
    if ei.recorder != nil {
        ei.recorder.Record(ev)
    }
    return ErrInjectionNotAllowed
}

func (ei *injector) injectAllowed(ev event.Event) error {
    // 测试环境：直接注入
    if ei.recorder != nil {
        ei.recorder.Record(ev)
    }
    if ei.handler != nil {
        return ei.handler(ev)
    }
    return nil
}

func (ei *injector) injectRecorded(ev event.Event) error {
    // 录制模式：只记录
    if ei.recorder != nil {
        return ei.recorder.Record(ev)
    }
    return nil
}

// SetStrategy 动态切换策略
func (ei *EventInjector) SetStrategy(strategy InjectionStrategy) {
    ei.strategy = strategy
}
```

### 2.2 增强的事件类型支持

```go
// sandbox/events.go (续)

// EventBuilder 事件构建器
type EventBuilder struct {
    sandbox Sandbox
}

// KeyPress 构建按键事件
func (eb *EventBuilder) KeyPress(key rune, mods ...Modifier) event.Event {
    ev := event.NewKeyEvent(event.Key{Rune: key})
    ev.Modifiers = combineModifiers(mods)
    return ev
}

// KeyCombo 构建组合键事件
func (eb *EventBuilder) KeyCombo(keys string) event.Event {
    // 解析 "Ctrl+C", "Alt+Shift+Delete" 等
    return parseKeyCombo(keys)
}

// MouseClick 构建鼠标点击事件
func (eb *EventBuilder) MouseClick(x, y int, button MouseButton, clicks int) event.Event {
    ev := event.NewMouseEvent(x, y, button, event.EventClick)
    ev.ClickCount = clicks
    return ev
}

// MouseScroll 构建鼠标滚轮事件
func (eb *EventBuilder) MouseScroll(x, y int, delta int) event.Event {
    ev := event.NewMouseEvent(x, y, 0, event.EventMouseWheel)
    ev.Delta = delta
    return ev
}

// MouseDrag 构建鼠标拖拽事件
func (eb *EventBuilder) MouseDrag(path []Point, button MouseButton) []event.Event {
    events := make([]event.Event, len(path))
    for i, p := range path {
        etype := event.EventMouseMove
        if i == 0 {
            etype = event.EventMousePress
        } else if i == len(path)-1 {
            etype = event.EventMouseRelease
        }
        events[i] = event.NewMouseEvent(p.X, p.Y, button, etype)
    }
    return events
}

// PasteText 构建粘贴事件
func (eb *EventBuilder) PasteText(text string) event.Event {
    return event.NewPasteEvent(text)
}

// ResizeWindow 构建窗口调整事件
func (eb *EventBuilder) ResizeWindow(width, height int) event.Event {
    return event.NewResizeEvent(width, height)
}
```

## 3. 内存优化的事件队列

### 3.1 有界事件队列

```go
// sandbox/mock/events.go

package mock

import (
    "container/ring"
    "sync"

    "github.com/wwsheng009/mint/framework/event"
)

// QueueConfig 队列配置
type QueueConfig struct {
    MaxSize      int           // 最大队列长度
    MaxMemory    int64         // 最大内存占用 (字节)
    EvictPolicy  EvictPolicy   // 淘汰策略
    PersistPath  string        // 持久化路径 (可选)
}

// EvictPolicy 淘汰策略
type EvictPolicy int

const (
    EvictOldest EvictPolicy = iota  // 淘汰最旧的
    EvictByPriority                  // 按优先级淘汰
    EvictPersist                     // 持久化到磁盘
)

// BoundedEventQueue 有界事件队列
type BoundedEventQueue struct {
    mu          sync.RWMutex
    config      QueueConfig
    ring        *ring.Ring          // 环形缓冲区
    size        int                 // 当前大小
    memory      int64               // 当前内存占用
    persisted   int                 // 已持久化的事件数
    storage     EventStorage        // 持久化存储 (可选)
}

// EventStorage 事件存储接口
type EventStorage interface {
    Append(events []event.Event) error
    Load(start, count int) ([]event.Event, error)
    Size() int
}

// Push 添加事件
func (q *BoundedEventQueue) Push(ev event.Event) error {
    q.mu.Lock()
    defer q.mu.Unlock()

    estimatedSize := estimateEventSize(ev)

    // 检查内存限制
    if q.config.MaxMemory > 0 && q.memory+estimatedSize > q.config.MaxMemory {
        if err := q.evict(q.config.EvictPolicy); err != nil {
            return err
        }
    }

    // 检查容量限制
    if q.config.MaxSize > 0 && q.size >= q.config.MaxSize {
        if err := q.evict(q.config.EvictPolicy); err != nil {
            return err
        }
    }

    // 添加到队列
    if q.ring == nil {
        q.ring = ring.New(q.config.MaxSize)
    }
    q.ring.Value = ev
    q.ring = q.ring.Next()
    q.size++
    q.memory += estimatedSize

    return nil
}

// evict 淘汰事件
func (q *BoundedEventQueue) evict(policy EvictPolicy) error {
    switch policy {
    case EvictOldest:
        return q.evictOldest()

    case EvictByPriority:
        return q.evictByPriority()

    case EvictPersist:
        return q.evictPersist()

    default:
        return ErrInvalidPolicy
    }
}

// evictOldest 淘汰最旧的事件
func (q *BoundedEventQueue) evictOldest() error {
    if q.size == 0 {
        return nil
    }

    // 找到最旧的事件并移除
    oldest := q.ring.Move(-q.size)
    if oldest.Value != nil {
        q.memory -= estimateEventSize(oldest.Value.(event.Event))
        oldest.Value = nil
        q.size--
    }

    return nil
}

// evictPersist 持久化到磁盘
func (q *BoundedEventQueue) evictPersist() error {
    if q.storage == nil {
        // 回退到淘汰最旧
        return q.evictOldest()
    }

    // 批量持久化旧事件
    batch := make([]event.Event, 0, 100)
    for i := 0; i < 100 && q.size > 0; i++ {
        oldest := q.ring.Move(-q.size)
        if oldest.Value != nil {
            batch = append(batch, oldest.Value.(event.Event))
            oldest.Value = nil
            q.size--
        }
    }

    if len(batch) > 0 {
        return q.storage.Append(batch)
    }

    return nil
}

// MemoryReport 内存报告
type MemoryReport struct {
    EventCount    int
    MemoryUsed    int64
    MemoryLimit   int64
    PersistedCount int
    OldestEvent   time.Time
    NewestEvent   time.Time
}

// ReportMemory 生成内存报告
func (q *BoundedEventQueue) ReportMemory() *MemoryReport {
    q.mu.RLock()
    defer q.mu.RUnlock()

    return &MemoryReport{
        EventCount:     q.size,
        MemoryUsed:     q.memory,
        MemoryLimit:    q.config.MaxMemory,
        PersistedCount: q.persisted,
    }
}
```

## 4. 增强的快照系统

### 4.1 分层快照

```go
// sandbox/snapshot.go

package sandbox

import (
    "encoding/json"
    "time"

    "github.com/wwsheng009/mint/runtime/paint"
)

// SnapshotLevel 快照级别
type SnapshotLevel int

const (
    // SnapshotLevelMinimal 最小快照 (仅渲染缓冲区)
    SnapshotLevelMinimal SnapshotLevel = iota

    // SnapshotLevelStandard 标准快照 (缓冲区+事件历史)
    SnapshotLevelStandard

    // SnapshotLevelFull 完整快照 (包括应用状态)
    SnapshotLevelFull
)

// Snapshot 快照 (增强版)
type Snapshot struct {
    Metadata  SnapshotMetadata
    Buffer    *paint.Buffer
    Events    []event.Event
    State     StateSnapshot
    Checksum  string
    compressed bool
}

// SnapshotMetadata 快照元数据
type SnapshotMetadata struct {
    ID        string
    Timestamp time.Time
    Level     SnapshotLevel
    Tags      []string
    Size      int64
}

// StateSnapshot 状态快照
type StateSnapshot struct {
    Components map[string]ComponentState
    Global     map[string]interface{}
}

// ComponentState 组件状态
type ComponentState struct {
    ID       string
    Type     string
    Props    map[string]interface{}
    Internal map[string]interface{}
}

// SnapshotManager 快照管理器
type SnapshotManager struct {
    snapshots []*Snapshot
    index     map[string]*Snapshot
    storage   SnapshotStorage
}

// SnapshotStorage 快照存储接口
type SnapshotStorage interface {
    Save(id string, snap *Snapshot) error
    Load(id string) (*Snapshot, error)
    Delete(id string) error
    List() []*SnapshotMetadata
}

// Create 创建快照
func (sm *SnapshotManager) Create(level SnapshotLevel, tags ...string) (*Snapshot, error) {
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
    case SnapshotLevelMinimal:
        sm.captureMinimal(snap)

    case SnapshotLevelStandard:
        sm.captureMinimal(snap)
        sm.captureEvents(snap)

    case SnapshotLevelFull:
        sm.captureMinimal(snap)
        sm.captureEvents(snap)
        sm.captureState(snap)
    }

    // 计算校验和
    snap.Checksum = computeChecksum(snap)

    // 存储
    sm.snapshots = append(sm.snapshots, snap)
    sm.index[snap.Metadata.ID] = snap

    if sm.storage != nil {
        if err := sm.storage.Save(snap.Metadata.ID, snap); err != nil {
            return nil, err
        }
    }

    return snap, nil
}

// Diff 快照差异
func (sm *SnapshotManager) Diff(from, to *Snapshot) (*SnapshotDiff, error) {
    return &SnapshotDiff{
        BufferChanged: !compareBuffers(from.Buffer, to.Buffer),
        EventsAdded:   diffEvents(from.Events, to.Events),
        StateChanges:  diffStates(from.State, to.State),
    }, nil
}

// SnapshotDiff 快照差异
type SnapshotDiff struct {
    BufferChanged bool
    EventsAdded   []event.Event
    EventsRemoved []event.Event
    StateChanges  StateChange
}

// StateChange 状态变化
type StateChange struct {
    Added     []string
    Modified  []string
    Deleted   []string
}
```

### 4.2 真实环境的快照支持

```go
// sandbox/real/snapshot.go

package real

import (
    "os"
    "path/filepath"

    "github.com/wwsheng009/mint/sandbox"
)

// FileSnapshotStorage 文件快照存储
type FileSnapshotStorage struct {
    dir string
}

func NewFileSnapshotStorage(dir string) (*FileSnapshotStorage, error) {
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, err
    }
    return &FileSnapshotStorage{dir: dir}, nil
}

func (fs *FileSnapshotStorage) Save(id string, snap *sandbox.Snapshot) error {
    path := filepath.Join(fs.dir, id+".json")
    data, err := json.Marshal(snap)
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

func (fs *FileSnapshotStorage) Load(id string) (*sandbox.Snapshot, error) {
    path := filepath.Join(fs.dir, id+".json")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var snap sandbox.Snapshot
    if err := json.Unmarshal(data, &snap); err != nil {
        return nil, err
    }
    return &snap, nil
}

// RealSandbox 的 Restore 实现 (软恢复)
func (rs *RealSandbox) Restore(snap *sandbox.Snapshot) error {
    // 真实环境支持软恢复：
    // 1. 恢复终端尺寸
    if snap.Buffer != nil {
        rs.SetSize(snap.Buffer.Width, snap.Buffer.Height)
    }

    // 2. 重新渲染
    if snap.Buffer != nil {
        rs.SetRenderBuffer(snap.Buffer)
        rs.render()
    }

    // 3. 应用状态恢复 (通过事件重放)
    for _, ev := range snap.Events {
        rs.handleEvent(ev)
    }

    return nil
}
```

## 5. 高性能事件调度

### 5.1 事件调度器

```go
// sandbox/scheduler.go

package sandbox

import (
    "time"

    "github.com/wwsheng009/mint/framework/event"
)

// Scheduler 事件调度器
type Scheduler struct {
    queue       *PriorityQueue
    timer       *time.Ticker
    dispatchCh  chan event.Event
    speed       float64  // 回放速度倍数
    realtime    bool     // 是否实时模式
}

// PriorityEvent 带优先级的事件
type PriorityEvent struct {
    Event     event.Event
    Priority  int
    Timestamp time.Time
    Delay     time.Duration
}

// NewScheduler 创建调度器
func NewScheduler(config *SchedulerConfig) *Scheduler {
    return &Scheduler{
        queue:      NewPriorityQueue(config.MaxSize),
        timer:      time.NewTicker(config.TickInterval),
        dispatchCh: make(chan event.Event, config.BufferSize),
        speed:      1.0,
        realtime:   true,
    }
}

// Schedule 调度事件
func (s *Scheduler) Schedule(ev event.Event, opts ...ScheduleOption) error {
    pe := &PriorityEvent{
        Event:     ev,
        Priority:  PriorityNormal,
        Timestamp: time.Now(),
    }

    // 应用选项
    for _, opt := range opts {
        opt(pe)
    }

    return s.queue.Push(pe)
}

// Run 运行调度器
func (s *Scheduler) Run() error {
    for {
        select {
        case <-s.timer.C:
            s.dispatchDueEvents()

        case ev := <-s.dispatchCh:
            s.handleDispatch(ev)
        }
    }
}

// dispatchDueEvents 分发到期事件
func (s *Scheduler) dispatchDueEvents() {
    now := time.Now()

    for !s.queue.Empty() {
        pe := s.queue.Peek().(*PriorityEvent)

        // 计算触发时间
        triggerTime := pe.Timestamp
        if s.realtime && pe.Delay > 0 {
            triggerTime = triggerTime.Add(pe.Delay)
        }

        if now.Before(triggerTime) {
            break
        }

        // 回放模式：考虑速度倍数
        if !s.realtime && s.speed != 1.0 {
            adjustedDelay := time.Duration(float64(pe.Delay) / s.speed)
            if time.Since(triggerTime) < adjustedDelay {
                break
            }
        }

        // 从队列移除并分发
        s.queue.Pop()
        s.dispatchCh <- pe.Event
    }
}

// ScheduleOption 调度选项
type ScheduleOption func(*PriorityEvent)

// WithDelay 设置延迟
func WithDelay(d time.Duration) ScheduleOption {
    return func(pe *PriorityEvent) {
        pe.Delay = d
    }
}

// WithPriority 设置优先级
func WithPriority(p int) ScheduleOption {
    return func(pe *PriorityEvent) {
        pe.Priority = p
    }
}

// WithTimestamp 设置时间戳
func WithTimestamp(t time.Time) ScheduleOption {
    return func(pe *PriorityEvent) {
        pe.Timestamp = t
    }
}
```

## 6. 增强的测试API

### 6.1 流式测试API

```go
// sandbox/mock/test_api.go

package mock

import (
    "github.com/wwsheng009/mint/sandbox"
    frameworkevent "github.com/wwsheng009/mint/framework/event"
)

// TestHelper 测试辅助器
type TestHelper struct {
    sandbox *MockSandbox
}

// When 创建条件触发器
func (th *TestHelper) When() *ConditionBuilder {
    return &ConditionBuilder{th: th}
}

// Assert 创建断言器
func (th *TestHelper) Assert() *AssertionBuilder {
    return &AssertionBuilder{th: th}
}

// Wait 创建等待器
func (th *TestHelper) Wait() *WaitBuilder {
    return &WaitBuilder{th: th}
}

// ConditionBuilder 条件构建器
type ConditionBuilder struct {
    th *TestHelper
}

// Focus 聚焦元素
func (cb *ConditionBuilder) Focus(selector string) *ActionBuilder {
    return &ActionBuilder{th: cb.th, action: "focus", target: selector}
}

// Hover 悬停在元素上
func (cb *ConditionBuilder) Hover(selector string) *ActionBuilder {
    return &ActionBuilder{th: cb.th, action: "hover", target: selector}
}

// ActionBuilder 动作构建器
type ActionBuilder struct {
    th     *TestHelper
    action string
    target string
}

// Click 点击
func (ab *ActionBuilder) Click() error {
    // 查找元素位置并注入点击事件
    x, y := ab.th.findElement(ab.target)
    return ab.th.sandbox.InjectMouse(x, y, MouseButtonLeft, EventClick)
}

// Type 输入文本
func (ab *ActionBuilder) Type(text string) error {
    // 先聚焦
    ab.th.sandbox.InjectSpecialKey(frameworkevent.KeyTab)
    ab.th.ProcessEvents()

    // 输入文本
    return ab.th.sandbox.InjectString(text)
}

// PressKey 按键
func (ab *ActionBuilder) PressKey(key string) error {
    return ab.th.sandbox.InjectSpecialKey(parseSpecialKey(key))
}

// AssertionBuilder 断言构建器
type AssertionBuilder struct {
    th *TestHelper
}

// RenderContains 断言渲染包含
func (as *AssertionBuilder) RenderContains(text string) error {
    return as.th.sandbox.AssertRender(text)
}

// ElementVisible 断言元素可见
func (as *AssertionBuilder) ElementVisible(selector string) error {
    visible := as.th.isElementVisible(selector)
    if !visible {
        return &AssertionError{
            Msg:      "element not visible",
            Selector: selector,
        }
    }
    return nil
}

// ValueEquals 断言值相等
func (as *AssertionBuilder) ValueEquals(selector, expected string) error {
    actual := as.th.getElementValue(selector)
    if actual != expected {
        return &AssertionError{
            Msg:      "value mismatch",
            Selector: selector,
            Expected: expected,
            Actual:   actual,
        }
    }
    return nil
}

// WaitBuilder 等待构建器
type WaitBuilder struct {
    th *TestHelper
}

// ForElement 等待元素出现
func (wb *WaitBuilder) ForElement(selector string) *WaitCondition {
    return &WaitCondition{th: wb.th, condition: "element", target: selector}
}

// WaitForValue 等待值变化
func (wb *WaitBuilder) ForValue(selector, value string) error {
    timeout := time.After(5 * time.Second)
    ticker := time.NewTicker(50 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-timeout:
            return ErrTimeout
        case <-ticker.C:
            if wb.th.getElementValue(selector) == value {
                return nil
            }
        }
    }
}

// WaitCondition 等待条件
type WaitCondition struct {
    th       *TestHelper
    condition string
    target   string
}

// ToBeVisible 等待直到可见
func (wc *WaitCondition) ToBeVisible() error {
    return wc.th.Wait().ForElement(wc.target).toBeVisible()
}

// 流式API使用示例
func TestFlow(t *testing.T) {
    testApp, _ := ui.TestRun(MyApp)
    defer testApp.Close()

    helper := testApp.Helper()

    // 流式测试
    helper.
        When().
        Focus("#username").
        Type("user@example.com").
        PressKey("Tab").
        Focus("#password").
        Type("secret").
        PressKey("Enter").
        Assert().
        RenderContains("Welcome").
        Assert().
        ElementVisible("#dashboard").
        Wait().
        ForElement("#loading").
        ToBeHidden()
}
```

## 7. CI/CD集成

### 7.1 测试运行器

```go
// sandbox/testing/runner.go

package testing

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/wwsheng009/mint/sandbox"
)

// TestSuite 测试套件
type TestSuite struct {
    Name      string
    Setup     func(*sandbox.MockSandbox) error
    Teardown  func(*sandbox.MockSandbox) error
    Tests     []TestCase
}

// TestCase 测试用例
type TestCase struct {
    Name      string
    Steps     []TestStep
    Assertions []Assertion
}

// Runner 测试运行器
type Runner struct {
    suites    []*TestSuite
    reporters []Reporter
    config    *RunnerConfig
}

// Reporter 报告器接口
type Reporter interface {
    OnSuiteStart(suite *TestSuite)
    OnSuiteEnd(suite *TestSuite, result *SuiteResult)
    OnTestStart(test *TestCase)
    OnTestEnd(test *TestCase, result *TestResult)
    OnFailure(failure *Failure)
}

// Run 运行测试
func (r *Runner) Run() *Report {
    report := &Report{
        StartTime: time.Now(),
    }

    for _, suite := range r.suites {
        r.runSuite(suite, report)
    }

    report.EndTime = time.Now()

    // 生成报告
    for _, reporter := range r.reporters {
        reporter.Generate(report)
    }

    return report
}

// JSONReporter JSON报告器
type JSONReporter struct {
    Output string
}

func (jr *JSONReporter) Generate(report *Report) error {
    data, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(jr.Output, data, 0644)
}

// JUnitReporter JUnit XML报告器
type JUnitReporter struct {
    Output string
}

func (jr *JUnitReporter) Generate(report *Report) error {
    // 生成JUnit XML格式报告
    return jr.writeJUnitXML(report)
}
```

### 7.2 录制回放集成

```go
// sandbox/testing/replay.go

package testing

import (
    "os"

    "github.com/wwsheng009/mint/sandbox"
)

// SessionRecorder 会话记录器
type SessionRecorder struct {
    sb       *sandbox.RealSandbox
    recording *Recording
}

// Recording 录制数据
type Recording struct {
    Metadata  RecordingMetadata
    Events    []event.Event
    Snapshots []*sandbox.Snapshot
}

// RecordingMetadata 录制元数据
type RecordingMetadata struct {
    ID        string
    StartTime time.Time
    EndTime   time.Time
    AppVersion string
    Platform    string
}

// StartRecording 开始录制
func StartRecording(app ComponentFunc, opts ...Option) (*SessionRecorder, error) {
    sb := sandbox.NewRealSandbox(80, 24)
    sb.RecordEvents(true)

    // ... 运行应用 ...

    return &SessionRecorder{
        sb: sb,
        recording: &Recording{
            Metadata: RecordingMetadata{
                ID:        generateID(),
                StartTime: time.Now(),
            },
            Events: sb.GetRecordedEvents(),
        },
    }, nil
}

// StopRecording 停止录制
func (sr *SessionRecorder) StopRecording() error {
    sr.recording.Metadata.EndTime = time.Now()
    return sr.sb.Close()
}

// Save 保存录制
func (sr *SessionRecorder) Save(path string) error {
    return os.WriteFile(path, sr.recording.Marshal(), 0644)
}

// LoadRecording 加载录制
func LoadRecording(path string) (*Recording, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    return UnmarshalRecording(data)
}

// ReplayTest 回放测试
func ReplayTest(t *testing.T, recordingPath string) {
    recording, err := LoadRecording(recordingPath)
    if err != nil {
        t.Fatal(err)
    }

    sb := sandbox.NewReplaySandbox(80, 24, recording.Events)
    container := sandbox.NewContainer(createApp(), sb)

    if err := container.Run(); err != nil {
        t.Errorf("replay failed: %v", err)
    }
}
```

## 8. 配置系统

### 8.1 沙箱配置

```go
// sandbox/config.go

package sandbox

import "time"

// Config 沙箱配置
type Config struct {
    // 基础配置
    Width      int
    Height     int
    Title      string
    FPS        int

    // 事件配置
    EventQueue *EventQueueConfig
    Injection  *InjectionConfig

    // 缓冲区配置
    Buffer     *BufferConfig

    // 快照配置
    Snapshot   *SnapshotConfig

    // 性能配置
    Performance *PerformanceConfig
}

// EventQueueConfig 事件队列配置
type EventQueueConfig struct {
    MaxSize     int           // 默认 10000
    MaxMemory   int64         // 默认 100MB
    EvictPolicy EvictPolicy   // 默认 EvictOldest
    PersistPath string        // 可选
}

// InjectionConfig 注入配置
type InjectionConfig struct {
    Strategy    InjectionStrategy
    Async       bool          // 异步注入
    BufferSize  int           // 注入缓冲区大小
}

// BufferConfig 缓冲区配置
type BufferConfig struct {
    DoubleBuffer   bool        // 双缓冲
    Compression    bool        // 压缩
    MaxHistory     int         // 历史记录数
}

// SnapshotConfig 快照配置
type SnapshotConfig struct {
    AutoSnapshot   bool        // 自动快照
    Interval       time.Duration // 快照间隔
    Storage        SnapshotStorage
    Level          SnapshotLevel // 默认快照级别
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
    Throttle       bool          // 节流
    MaxFPS         int           // 最大帧率
    Profile        bool          // 性能分析
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
    return &Config{
        Width:  80,
        Height: 24,
        FPS:    60,

        EventQueue: &EventQueueConfig{
            MaxSize:     10000,
            MaxMemory:   100 * 1024 * 1024, // 100MB
            EvictPolicy: EvictOldest,
        },

        Injection: &InjectionConfig{
            Strategy:   InjectStrategyAllowed,
            Async:      true,
            BufferSize: 100,
        },

        Buffer: &BufferConfig{
            DoubleBuffer: true,
            Compression:  false,
            MaxHistory:   10,
        },

        Snapshot: &SnapshotConfig{
            AutoSnapshot: false,
            Level:        SnapshotLevelStandard,
        },

        Performance: &PerformanceConfig{
            Throttle: true,
            MaxFPS:   60,
            Profile:  false,
        },
    }
}
```

## 9. 更新后的文件结构

```
mint/
├── sandbox/                      # 独立沙箱模块
│   ├── sandbox.go                # 核心接口
│   ├── lifecycle.go              # 生命周期管理
│   ├── events.go                 # 事件系统
│   ├── scheduler.go              # 事件调度器
│   ├── buffer.go                 # 渲染缓冲区管理
│   ├── snapshot.go               # 快照系统
│   ├── config.go                 # 配置系统
│   ├── types.go                  # 类型定义
│   ├── errors.go                 # 错误定义
│   │
│   ├── real/                     # 真实环境
│   │   ├── sandbox.go
│   │   ├── input.go
│   │   ├── terminal.go           # 终端控制
│   │   └── snapshot.go           # 文件快照存储
│   │
│   ├── mock/                     # 模拟环境
│   │   ├── sandbox.go
│   │   ├── events.go             # 有界事件队列
│   │   ├── assertions.go         # 断言功能
│   │   ├── test_api.go           # 流式测试API
│   │   └── selectors.go          # 元素选择器
│   │
│   ├── replay/                   # 回放环境
│   │   ├── sandbox.go
│   │   ├── player.go             # 回放控制器
│   │   └── recorder.go           # 录制控制器
│   │
│   └── testing/                  # 测试工具
│       ├── runner.go             # 测试运行器
│       ├── reporter.go            # 报告器
│       ├── replay.go              # 录制回放
│       └── junit.go               # JUnit报告器
│
├── runtime/                      # 运行时 (不变)
│   ├── paint/
│   ├── platform/
│   └── ...
│
├── ui/                           # UI层
│   ├── app.go                    # 添加测试支持
│   └── test.go                   # 测试包装器
│
├── examples/
│   └── test_button/
│       ├── main.go
│       └── main_test.go
│
└── docs/
    ├── SANDBOX_DESIGN.md         # 设计文档
    ├── TESTING_GUIDE.md          # 测试指南
    └── SANDBOX_API.md            # API参考
```

## 10. 测试示例 (更新后)

```go
package ui_test

import (
    "testing"
    "time"

    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/sandbox"
)

func TestLoginForm(t *testing.T) {
    testApp, err := ui.TestRun(LoginForm,
        ui.WithWidth(80),
        ui.WithHeight(24),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    helper := testApp.Helper()

    // 使用流式API
    helper.
        When().Focus("#username").Type("user@example.com").
        PressKey("Tab").
        Focus("#password").Type("secret").
        PressKey("Enter").
        Assert().RenderContains("Welcome").
        Assert().ElementVisible("#dashboard")

    // 或使用传统API
    testApp.InjectString("user@example.com")
    testApp.InjectSpecialKey(sandbox.KeyTab)
    testApp.InjectString("secret")
    testApp.InjectSpecialKey(sandbox.KeyEnter)
    testApp.ProcessEvents()

    if err := testApp.AssertRender("Welcome"); err != nil {
        t.Error(err)
    }
}

func TestWithSnapshot(t *testing.T) {
    testApp, _ := ui.TestRun(MyComponent)
    defer testApp.Close()

    // 创建快照
    snap, _ := testApp.Snapshot(sandbox.SnapshotLevelFull)

    // 执行操作
    testApp.InjectString("test")
    testApp.ProcessEvents()

    // 恢复快照
    testApp.Restore(snap)

    // 验证恢复正确
    testApp.Assert().RenderContains("initial state")
}

func TestMemoryConstrained(t *testing.T) {
    // 配置内存限制
    config := sandbox.DefaultConfig()
    config.EventQueue.MaxMemory = 10 * 1024 * 1024 // 10MB

    testApp, _ := ui.TestRunWithConfig(MyComponent, config)
    defer testApp.Close()

    // 注入大量事件
    for i := 0; i < 100000; i++ {
        testApp.InjectKey('a')
    }

    // 检查内存报告
    report := testApp.MemoryReport()
    if report.MemoryUsed > report.MemoryLimit {
        t.Errorf("memory limit exceeded: %d > %d",
            report.MemoryUsed, report.MemoryLimit)
    }
}
```

## 11. 实施计划 (更新)

### 阶段1: 核心接口 (1-2天)
- [ ] 创建 `sandbox/` 目录
- [ ] 实现 `sandbox.go` 核心接口
- [ ] 实现 `lifecycle.go` 生命周期
- [ ] 实现 `types.go`, `errors.go`
- [ ] 实现 `config.go` 配置系统

### 阶段2: 事件系统 (2-3天)
- [ ] 实现 `events.go` 事件注入策略
- [ ] 实现 `scheduler.go` 事件调度器
- [ ] 实现 `mock/events.go` 有界队列
- [ ] 添加事件测试

### 阶段3: 真实环境实现 (2-3天)
- [ ] 实现 `real/sandbox.go`
- [ ] 实现 `real/input.go` 包装platform
- [ ] 实现 `real/terminal.go`
- [ ] 实现 `real/snapshot.go` 文件存储

### 阶段4: 模拟环境实现 (3-4天)
- [ ] 实现 `mock/sandbox.go`
- [ ] 实现 `mock/assertions.go`
- [ ] 实现 `mock/test_api.go` 流式API
- [ ] 实现 `mock/selectors.go` 元素选择器

### 阶段5: 回放系统 (2-3天)
- [ ] 实现 `replay/sandbox.go`
- [ ] 实现 `replay/player.go`
- [ ] 实现 `replay/recorder.go`
- [ ] 添加录制回放测试

### 阶段6: 测试工具 (2-3天)
- [ ] 实现 `testing/runner.go`
- [ ] 实现 `testing/reporter.go`
- [ ] 实现 `testing/replay.go`
- [ ] 实现 `testing/junit.go`

### 阶段7: UI层集成 (1-2天)
- [ ] 修改 `ui/app.go` 添加测试支持
- [ ] 实现 `ui/test.go`
- [ ] 添加示例测试

### 阶段8: 文档 (1-2天)
- [ ] 编写 `docs/TESTING_GUIDE.md`
- [ ] 编写 `docs/SANDBOX_API.md`
- [ ] 添加示例代码注释

## 12. 验收标准

### 功能验收
- [ ] 所有沙箱类型可正常创建和运行
- [ ] 事件注入在MockSandbox中工作正常
- [ ] 快照创建和恢复功能正常
- [ ] 回放功能可以复现已录制会话
- [ ] 流式测试API可用

### 性能验收
- [ ] MockSandbox内存占用可控
- [ ] 事件调度器支持10000+事件/秒
- [ ] 快照压缩率 > 50%

### 测试验收
- [ ] 单元测试覆盖率 > 80%
- [ ] 所有示例可通过测试
- [ ] CI/CD集成测试通过
