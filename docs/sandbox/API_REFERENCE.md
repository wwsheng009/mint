# Sandbox API 参考手册

> Mint TUI 框架沙箱测试环境 API 完整参考
>
> 版本: 1.0
> 更新日期: 2026-02-01

---

## 目录

1. [核心类型](#核心类型)
2. [核心接口](#核心接口)
3. [生命周期](#生命周期)
4. [配置系统](#配置系统)
5. [事件系统](#事件系统)
6. [快照系统](#快照系统)
7. [Mock 沙箱](#mock-沙箱)
8. [真实沙箱](#真实沙箱)
9. [回放沙箱](#回放沙箱)
10. [测试辅助器](#测试辅助器)
11. [错误处理](#错误处理)
12. [UI 集成](#ui-集成)

---

## 核心类型

### SandboxType

沙箱类型枚举。

```go
type SandboxType int

const (
    TypeReal   SandboxType = iota  // 真实终端环境
    TypeMock                       // 模拟测试环境
    TypeReplay                     // 回放环境
)
```

#### 方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `String()` | `string` | 返回类型的字符串表示 |

#### 示例

```go
sb := mock.New(80, 24)
fmt.Println(sb.Type()) // 输出: mock
```

---

### State

沙箱状态枚举。

```go
type State int

const (
    StateStopped     State = iota  // 已停止
    StateInitialized               // 已初始化
    StateRunning                   // 运行中
    StatePaused                    // 已暂停
    StateError                     // 错误状态
)
```

#### 方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `String()` | `string` | 返回状态的字符串表示 |

#### 状态转换图

```
Stopped → Initialized → Running → Paused
    ↓         ↓             ↓         ↓
    └─────────┴─────────────┴─────────→ Error
                                  ↓
                              Stopped
```

---

### Phase

生命周期阶段枚举。

```go
type Phase int

const (
    PhaseBefore Phase = iota  // 状态转换前
    PhaseAfter               // 状态转换后
)
```

---

### HookKey

钩子键，用于注册状态转换钩子。

```go
type HookKey struct {
    State State
    Phase Phase
}
```

#### 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `State` | `State` | 状态 |
| `Phase` | `Phase` | 阶段 (Before/After) |

---

### InjectionStrategy

事件注入策略枚举。

```go
type InjectionStrategy int

const (
    InjectProhibited InjectionStrategy = iota  // 禁止注入 (真实环境)
    InjectAllowed                               // 允许注入 (测试环境)
    InjectRecorded                              // 仅录制 (录制模式)
)
```

#### 方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `String()` | `string` | 返回策略的字符串表示 |

---

### EvictPolicy

事件淘汰策略枚举。

```go
type EvictPolicy int

const (
    EvictOldest     EvictPolicy = iota  // 淘汰最旧的
    EvictByPriority                    // 按优先级淘汰
    EvictPersist                       // 持久化到磁盘
)
```

#### 方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `String()` | `string` | 返回策略的字符串表示 |

---

### SnapshotLevel

快照级别枚举。

```go
type SnapshotLevel int

const (
    SnapshotMinimal  SnapshotLevel = iota  // 仅渲染缓冲区
    SnapshotStandard                      // 缓冲区+事件历史
    SnapshotFull                          // 包括应用状态
)
```

#### 方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `String()` | `string` | 返回级别的字符串表示 |

---

### InputEvent

统一输入事件，包装 `platform.RawInput`。

```go
type InputEvent struct {
    Raw       platform.RawInput  // 原始事件
    Injected  bool               // 是否为注入事件
    Timestamp time.Time          // 时间戳
}
```

---

### BufferWrapper

缓冲区包装器，支持历史快照。

```go
type BufferWrapper struct {
    *paint.Buffer           // 嵌入的 Buffer
    history    []*paint.Buffer  // 历史快照
    maxHistory int              // 最大历史数
}
```

#### 方法

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `NewBufferWrapper(buf *paint.Buffer, maxHistory int)` | `*BufferWrapper` | 创建包装器 |
| `SaveSnapshot()` | - | 保存当前缓冲区到历史 |
| `History()` | `[]*paint.Buffer` | 返回历史快照 |
| `ClearHistory()` | - | 清空历史 |

---

## 核心接口

### Sandbox

沙箱核心接口，所有沙箱类型必须实现。

```go
type Sandbox interface {
    // ========================================================================
    // 生命周期
    // ========================================================================

    Initialize(config *Config) error
    Run() error
    Pause() error
    Resume() error
    Close() error

    // ========================================================================
    // 状态查询
    // ========================================================================

    State() State
    Type() SandboxType
    Config() *Config

    // ========================================================================
    // 缓冲区操作
    // ========================================================================

    Buffer() *paint.Buffer
    SetBuffer(buf *paint.Buffer)
    Resize(width, height int)
    Size() (width, height int)
}
```

#### 方法详情

##### Initialize

初始化沙箱。

```go
Initialize(config *Config) error
```

**参数：**
- `config` - 沙箱配置，nil 表示使用默认配置

**返回：**
- `error` - 初始化错误

**状态转换：** `Stopped → Initialized`

---

##### Run

运行沙箱主循环。

```go
Run() error
```

**返回：**
- `error` - 运行错误

**状态转换：** `Initialized → Running`

---

##### Pause

暂停沙箱。

```go
Pause() error
```

**返回：**
- `error` - 暂停错误

**状态转换：** `Running → Paused`

---

##### Resume

恢复沙箱运行。

```go
Resume() error
```

**返回：**
- `error` - 恢复错误

**状态转换：** `Paused → Running`

---

##### Close

关闭沙箱并释放资源。

```go
Close() error
```

**返回：**
- `error` - 关闭错误

**状态转换：** `* → Stopped`

---

##### State

获取当前状态。

```go
State() State
```

**返回：**
- `State` - 当前状态

---

##### Type

获取沙箱类型。

```go
Type() SandboxType
```

**返回：**
- `SandboxType` - 沙箱类型 (TypeReal/TypeMock/TypeReplay)

---

##### Config

获取配置。

```go
Config() *Config
```

**返回：**
- `*Config` - 配置指针（只读）

---

##### Buffer

获取渲染缓冲区。

```go
Buffer() *paint.Buffer
```

**返回：**
- `*paint.Buffer` - 渲染缓冲区

---

##### SetBuffer

设置渲染缓冲区。

```go
SetBuffer(buf *paint.Buffer)
```

**参数：**
- `buf` - 新的渲染缓冲区

---

##### Resize

调整缓冲区大小。

```go
Resize(width, height int)
```

**参数：**
- `width` - 新宽度
- `height` - 新高度

---

##### Size

获取当前尺寸。

```go
Size() (width, height int)
```

**返回：**
- `width` - 宽度
- `height` - 高度

---

### EventSource

事件源接口，用于真实环境。

```go
type EventSource interface {
    Events() <-chan platform.RawInput
    Start() error
    Stop() error
}
```

#### 方法详情

##### Events

返回事件通道。

```go
Events() <-chan platform.RawInput
```

**返回：**
- `<-chan platform.RawInput` - 只读事件通道

---

##### Start

启动事件读取。

```go
Start() error
```

**返回：**
- `error` - 启动错误

---

##### Stop

停止事件读取。

```go
Stop() error
```

**返回：**
- `error` - 停止错误

---

### EventSink

事件注入接口，用于测试环境。

```go
type EventSink interface {
    SetEventHandler(handler EventHandler)
    Inject(event platform.RawInput) error
    InjectKey(key rune) error
    InjectSpecialKey(key platform.SpecialKey) error
    InjectKeyWithMod(key rune, mod platform.KeyModifier) error
    InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error
    InjectResize(width, height int) error
    InjectString(text string) error
    ProcessEvents() error
}
```

#### 方法详情

##### SetEventHandler

设置事件处理器。

```go
SetEventHandler(handler EventHandler)
```

**参数：**
- `handler` - 事件处理函数

**事件处理函数类型：**
```go
type EventHandler func(event platform.RawInput) error
```

---

##### Inject

注入单个事件。

```go
Inject(event platform.RawInput) error
```

**参数：**
- `event` - 要注入的事件

**返回：**
- `error` - 注入错误

---

##### InjectKey

注入按键事件。

```go
InjectKey(key rune) error
```

**参数：**
- `key` - 字符键（如 'a', '1'）

**返回：**
- `error` - 注入错误

---

##### InjectSpecialKey

注入特殊按键。

```go
InjectSpecialKey(key platform.SpecialKey) error
```

**参数：**
- `key` - 特殊按键（如 `platform.KeyEnter`, `platform.KeyTab`）

**返回：**
- `error` - 注入错误

**可用的特殊按键：**
- `platform.KeyEnter`
- `platform.KeyTab`
- `platform.KeyEscape`
- `platform.KeySpace`
- `platform.KeyBackspace`
- `platform.KeyDelete`
- `platform.KeyUp`
- `platform.KeyDown`
- `platform.KeyLeft`
- `platform.KeyRight`

---

##### InjectKeyWithMod

注入带修饰符的按键。

```go
InjectKeyWithMod(key rune, mod platform.KeyModifier) error
```

**参数：**
- `key` - 字符键
- `mod` - 修饰符（如 `platform.KeyModCtrl`, `platform.KeyModShift`）

**返回：**
- `error` - 注入错误

**修饰符类型：**
- `platform.KeyModCtrl` - Ctrl
- `platform.KeyModAlt` - Alt
- `platform.KeyModShift` - Shift

---

##### InjectMouse

注入鼠标事件。

```go
InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error
```

**参数：**
- `x` - X 坐标
- `y` - Y 坐标
- `button` - 鼠标按钮（`platform.MouseLeft/Middle/Right`）
- `action` - 鼠标动作（`platform.MousePress/Release/Scroll`）

**返回：**
- `error` - 注入错误

---

##### InjectResize

注入窗口调整事件。

```go
InjectResize(width, height int) error
```

**参数：**
- `width` - 新宽度
- `height` - 新高度

**返回：**
- `error` - 注入错误

---

##### InjectString

注入字符串（转换为按键序列）。

```go
InjectString(text string) error
```

**参数：**
- `text` - 要输入的字符串

**返回：**
- `error` - 注入错误

**示例：**
```go
sb.InjectString("hello world")
// 等价于：
// sb.InjectKey('h')
// sb.InjectKey('e')
// sb.InjectKey('l')
// sb.InjectKey('l')
// sb.InjectKey('o')
// ...
```

---

##### ProcessEvents

处理所有待处理事件。

```go
ProcessEvents() error
```

**返回：**
- `error` - 处理错误

**说明：** 从事件队列中取出所有事件并调用事件处理器。

---

### Snapshotter

快照接口。

```go
type Snapshotter interface {
    Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)
    Restore(snap *Snapshot) error
    ListSnapshots() []*SnapshotMetadata
}
```

#### 方法详情

##### Snapshot

创建快照。

```go
Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)
```

**参数：**
- `level` - 快照级别（`SnapshotMinimal/Standard/Full`）
- `tags` - 可选的自定义标签

**返回：**
- `*Snapshot` - 快照对象
- `error` - 错误

---

##### Restore

恢复快照。

```go
Restore(snap *Snapshot) error
```

**参数：**
- `snap` - 要恢复的快照

**返回：**
- `error` - 错误

---

##### ListSnapshots

列出所有快照元数据。

```go
ListSnapshots() []*SnapshotMetadata
```

**返回：**
- `[]*SnapshotMetadata` - 快照元数据列表

---

### TestSandbox

测试沙箱接口，组合了 Sandbox、EventSink、Snapshotter。

```go
type TestSandbox interface {
    Sandbox
    EventSink
    Snapshotter

    IsMock() bool
    AssertRender(text string) error
    AssertNotRender(text string) error
    RenderString() string
    Helper() interface{}
}
```

#### 方法详情

##### IsMock

是否为模拟沙箱。

```go
IsMock() bool
```

**返回：**
- `bool` - true 表示 Mock 沙箱

---

##### AssertRender

断言渲染输出包含指定文本。

```go
AssertRender(text string) error
```

**参数：**
- `text` - 期望的文本

**返回：**
- `error` - 如果不包含文本则返回错误

---

##### AssertNotRender

断言渲染输出不包含指定文本。

```go
AssertNotRender(text string) error
```

**参数：**
- `text` - 不期望的文本

**返回：**
- `error` - 如果包含文本则返回错误

---

##### RenderString

获取渲染输出字符串。

```go
RenderString() string
```

**返回：**
- `string` - 渲染输出的字符串表示

---

##### Helper

获取测试辅助器。

```go
Helper() interface{}
```

**返回：**
- `interface{}` - 测试辅助器（在 Mock 沙箱中为 `*mock.TestHelper`）

---

## 生命周期

### Lifecycle

生命周期管理器。

```go
type Lifecycle struct {
    mu     sync.RWMutex
    state  State
    err    error
    hooks  map[HookKey][]HookFunc
}
```

#### 方法

##### NewLifecycle

创建生命周期管理器。

```go
func NewLifecycle() *Lifecycle
```

**返回：**
- `*Lifecycle` - 新的生命周期管理器

---

##### State

获取当前状态。

```go
func (l *Lifecycle) State() State
```

**返回：**
- `State` - 当前状态

---

##### Error

获取错误状态。

```go
func (l *Lifecycle) Error() error
```

**返回：**
- `error` - 错误（如果有）

---

##### Transition

执行状态转移。

```go
func (l *Lifecycle) Transition(to State) error
```

**参数：**
- `to` - 目标状态

**返回：**
- `error` - 转移错误

---

##### OnTransition

注册状态转移钩子。

```go
func (l *Lifecycle) OnTransition(state State, phase Phase, fn HookFunc)
```

**参数：**
- `state` - 状态
- `phase` - 阶段（Before/After）
- `fn` - 钩子函数

**钩子函数类型：**
```go
type HookFunc func() error
```

---

##### Reset

重置生命周期。

```go
func (l *Lifecycle) Reset()
```

---

##### CanTransitionTo

检查是否可以转移到目标状态。

```go
func (l *Lifecycle) CanTransitionTo(to State) bool
```

**参数：**
- `to` - 目标状态

**返回：**
- `bool` - 是否可以转移

---

## 配置系统

### Config

沙箱配置。

```go
type Config struct {
    // 基础配置
    Width  int
    Height int
    Title  string
    FPS    int

    // 事件配置
    Event EventConfig

    // 快照配置
    Snapshot SnapshotConfig

    // 性能配置
    Performance PerformanceConfig
}
```

#### 字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Width` | `int` | 80 | 缓冲区宽度 |
| `Height` | `int` | 24 | 缓冲区高度 |
| `Title` | `string` | "" | 标题 |
| `FPS` | `int` | 60 | 帧率 |

#### 方法

##### Validate

验证配置。

```go
func (c *Config) Validate() error
```

**返回：**
- `error` - 验证错误

---

##### Clone

克隆配置。

```go
func (c *Config) Clone() *Config
```

**返回：**
- `*Config` - 配置副本

---

### EventConfig

事件配置。

```go
type EventConfig struct {
    QueueMaxSize   int               // 最大队列长度 (默认 10000)
    QueueMaxMemory int64             // 最大内存占用 (默认 100MB)
    EvictPolicy    EvictPolicy       // 淘汰策略
    Strategy       InjectionStrategy // 注入策略
    RecordEnabled  bool              // 是否启用录制
    RecordMaxLen   int               // 录制最大长度
}
```

---

### SnapshotConfig

快照配置。

```go
type SnapshotConfig struct {
    AutoSnapshot bool           // 自动快照
    Interval     time.Duration  // 快照间隔
    MaxCount     int            // 最大快照数
    Level        SnapshotLevel  // 默认快照级别
    PersistPath  string         // 持久化路径
}
```

---

### PerformanceConfig

性能配置。

```go
type PerformanceConfig struct {
    Throttle      bool          // 节流
    MaxFPS        int           // 最大帧率
    RenderTimeout time.Duration // 渲染超时
    Profile       bool          // 性能分析
}
```

---

#### 配置工厂函数

##### DefaultConfig

返回默认配置。

```go
func DefaultConfig() *Config
```

**返回：**
- `*Config` - 默认配置

**默认值：**
```go
&Config{
    Width:  80,
    Height: 24,
    FPS:    60,
    Event: EventConfig{
        QueueMaxSize:   10000,
        QueueMaxMemory: 100 * 1024 * 1024,
        EvictPolicy:    EvictOldest,
        Strategy:       InjectAllowed,
        RecordMaxLen:   10000,
    },
    Snapshot: SnapshotConfig{
        MaxCount: 100,
        Level:    SnapshotStandard,
    },
    Performance: PerformanceConfig{
        Throttle:      true,
        MaxFPS:        60,
        RenderTimeout: 100 * time.Millisecond,
    },
}
```

---

##### RealConfig

返回真实环境配置。

```go
func RealConfig() *Config
```

**返回：**
- `*Config` - 真实环境配置

**特点：**
- `Event.Strategy = InjectProhibited`
- `Event.RecordEnabled = true`

---

##### MockConfig

返回模拟环境配置。

```go
func MockConfig() *Config
```

**返回：**
- `*Config` - 模拟环境配置

**特点：**
- `Event.Strategy = InjectAllowed`
- `Performance.Throttle = false`

---

##### ReplayConfig

返回回放环境配置。

```go
func ReplayConfig() *Config
```

**返回：**
- `*Config` - 回放环境配置

**特点：**
- `Event.Strategy = InjectRecorded`

---

## 事件系统

### EventInjector

事件注入器。

```go
type EventInjector struct {
    mu       sync.RWMutex
    strategy InjectionStrategy
    handler  EventHandler
    recorder *EventRecorder
}
```

#### 方法

##### NewEventInjector

创建事件注入器。

```go
func NewEventInjector(strategy InjectionStrategy) *EventInjector
```

**参数：**
- `strategy` - 注入策略

**返回：**
- `*EventInjector` - 事件注入器

---

##### SetHandler

设置事件处理器。

```go
func (ei *EventInjector) SetHandler(handler EventHandler)
```

---

##### SetRecorder

设置事件录制器。

```go
func (ei *EventInjector) SetRecorder(recorder *EventRecorder)
```

---

##### Strategy

获取当前策略。

```go
func (ei *EventInjector) Strategy() InjectionStrategy
```

**返回：**
- `InjectionStrategy` - 当前策略

---

##### SetStrategy

设置策略。

```go
func (ei *EventInjector) SetStrategy(strategy InjectionStrategy)
```

---

##### Inject

注入事件。

```go
func (ei *EventInjector) Inject(event platform.RawInput) error
```

**参数：**
- `event` - 要注入的事件

**返回：**
- `error` - 注入错误

---

### EventRecorder

事件录制器。

```go
type EventRecorder struct {
    mu     sync.Mutex
    events []platform.RawInput
    maxLen int
}
```

#### 方法

##### NewEventRecorder

创建事件录制器。

```go
func NewEventRecorder(maxLen int) *EventRecorder
```

**参数：**
- `maxLen` - 最大录制长度

**返回：**
- `*EventRecorder` - 事件录制器

---

##### Record

录制事件。

```go
func (r *EventRecorder) Record(event platform.RawInput) error
```

**参数：**
- `event` - 要录制的事件

**返回：**
- `error` - 错误

**说明：** 如果超过最大长度，会自动淘汰最旧的事件。

---

##### Events

获取所有录制的事件。

```go
func (r *EventRecorder) Events() []platform.RawInput
```

**返回：**
- `[]platform.RawInput` - 事件列表

---

##### Clear

清空录制。

```go
func (r *EventRecorder) Clear()
```

---

##### Len

返回事件数量。

```go
func (r *EventRecorder) Len() int
```

**返回：**
- `int` - 事件数量

---

## 快照系统

### Snapshot

快照。

```go
type Snapshot struct {
    Metadata SnapshotMetadata          // 元数据
    Buffer   *paint.Buffer             // 渲染缓冲区
    Events   []platform.RawInput        // 事件历史
    State    map[string]interface{}     // 应用状态
    Checksum string                     // 校验和
}
```

---

### SnapshotMetadata

快照元数据。

```go
type SnapshotMetadata struct {
    ID        string        // 唯一标识
    Timestamp time.Time     // 创建时间
    Level     SnapshotLevel // 快照级别
    Tags      []string      // 自定义标签
    Size      int64         // 大小（字节）
}
```

---

### SnapshotManager

快照管理器。

```go
type SnapshotManager struct {
    mu        sync.RWMutex
    snapshots map[string]*Snapshot
    order     []string
    maxCount  int
    storage   SnapshotStorage
}
```

#### 方法

##### NewSnapshotManager

创建快照管理器。

```go
func NewSnapshotManager(maxCount int) *SnapshotManager
```

**参数：**
- `maxCount` - 最大快照数量

**返回：**
- `*SnapshotManager` - 快照管理器

---

##### SetStorage

设置持久化存储。

```go
func (sm *SnapshotManager) SetStorage(storage SnapshotStorage)
```

---

##### Create

创建快照。

```go
func (sm *SnapshotManager) Create(level SnapshotLevel, buffer *paint.Buffer, events []platform.RawInput, state map[string]interface{}, tags ...string) (*Snapshot, error)
```

**参数：**
- `level` - 快照级别
- `buffer` - 渲染缓冲区
- `events` - 事件历史
- `state` - 应用状态
- `tags` - 可选标签

**返回：**
- `*Snapshot` - 快照
- `error` - 错误

---

##### Get

获取快照。

```go
func (sm *SnapshotManager) Get(id string) (*Snapshot, error)
```

**参数：**
- `id` - 快照 ID

**返回：**
- `*Snapshot` - 快照
- `error` - 错误

---

##### Delete

删除快照。

```go
func (sm *SnapshotManager) Delete(id string) error
```

**参数：**
- `id` - 快照 ID

**返回：**
- `error` - 错误

---

##### Verify

验证快照完整性。

```go
func (sm *SnapshotManager) Verify(snap *Snapshot) bool
```

**参数：**
- `snap` - 要验证的快照

**返回：**
- `bool` - 是否有效

---

## Mock 沙箱

### MockSandbox

模拟沙箱实现。

```go
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
```

#### 构造函数

##### New

创建模拟沙箱。

```go
func New(width, height int) *MockSandbox
```

**参数：**
- `width` - 缓冲区宽度
- `height` - 缓冲区高度

**返回：**
- `*MockSandbox` - 模拟沙箱

---

#### 方法

MockSandbox 实现了所有 Sandbox、EventSink、Snapshotter、TestSandbox 接口的方法。

**Sandbox 接口：**
- `Initialize(config *Config) error`
- `Run() error`
- `Pause() error`
- `Resume() error`
- `Close() error`
- `State() State`
- `Type() SandboxType`
- `Config() *Config`
- `Buffer() *paint.Buffer`
- `SetBuffer(buf *paint.Buffer)`
- `Resize(width, height int)`
- `Size() (int, int)`

**EventSink 接口：**
- `SetEventHandler(handler EventHandler)`
- `Inject(event platform.RawInput) error`
- `InjectKey(key rune) error`
- `InjectSpecialKey(key platform.SpecialKey) error`
- `InjectKeyWithMod(key rune, mod platform.KeyModifier) error`
- `InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error`
- `InjectResize(width, height int) error`
- `InjectString(text string) error`
- `ProcessEvents() error`

**Snapshotter 接口：**
- `Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)`
- `Restore(snap *Snapshot) error`
- `ListSnapshots() []*SnapshotMetadata`

**TestSandbox 接口：**
- `IsMock() bool`
- `AssertRender(text string) error`
- `AssertNotRender(text string) error`
- `RenderString() string`
- `Helper() *TestHelper`

**额外方法：**
- `QueueStats() QueueStats` - 获取队列统计

---

### BoundedQueue

有界事件队列。

```go
type BoundedQueue struct {
    mu          sync.RWMutex
    config      QueueConfig
    events      []platform.RawInput
    memory      int64
    evictCount  int64
}
```

#### 构造函数

##### NewBoundedQueue

创建有界队列。

```go
func NewBoundedQueue(config QueueConfig) *BoundedQueue
```

**参数：**
- `config` - 队列配置

**返回：**
- `*BoundedQueue` - 有界队列

---

##### DefaultQueueConfig

返回默认队列配置。

```go
func DefaultQueueConfig() QueueConfig
```

**返回：**
- `QueueConfig` - 默认配置

**默认值：**
```go
QueueConfig{
    MaxSize:     10000,
    MaxMemory:   100 * 1024 * 1024,  // 100MB
    EvictPolicy: EvictOldest,
}
```

---

#### 方法

##### Push

添加事件。

```go
func (q *BoundedQueue) Push(event platform.RawInput) error
```

**参数：**
- `event` - 要添加的事件

**返回：**
- `error` - 错误

---

##### Pop

取出最旧的事件。

```go
func (q *BoundedQueue) Pop() (platform.RawInput, error)
```

**返回：**
- `platform.RawInput` - 事件
- `error` - 错误

---

##### Peek

查看最旧的事件（不移除）。

```go
func (q *BoundedQueue) Peek() (platform.RawInput, error)
```

**返回：**
- `platform.RawInput` - 事件
- `error` - 错误

---

##### Len

返回队列长度。

```go
func (q *BoundedQueue) Len() int
```

**返回：**
- `int` - 队列长度

---

##### IsEmpty

检查队列是否为空。

```go
func (q *BoundedQueue) IsEmpty() bool
```

**返回：**
- `bool` - 是否为空

---

##### Clear

清空队列。

```go
func (q *BoundedQueue) Clear()
```

---

##### Stats

获取队列统计。

```go
func (q *BoundedQueue) Stats() QueueStats
```

**返回：**
- `QueueStats` - 统计信息

---

### QueueStats

队列统计信息。

```go
type QueueStats struct {
    Length      int
    MemoryUsed  int64
    MemoryLimit int64
    EvictCount  int64
}
```

---

### TestHelper

测试辅助器，提供链式 API。

```go
type TestHelper struct {
    sandbox *MockSandbox
    errors  []error
}
```

#### 构造函数

##### NewTestHelper

创建测试辅助器。

```go
func NewTestHelper(sb *MockSandbox) *TestHelper
```

**参数：**
- `sb` - Mock 沙箱

**返回：**
- `*TestHelper` - 测试辅助器

---

#### 方法

##### Errors

返回所有错误。

```go
func (th *TestHelper) Errors() []error
```

**返回：**
- `[]error` - 错误列表

---

##### HasErrors

检查是否有错误。

```go
func (th *TestHelper) HasErrors() bool
```

**返回：**
- `bool` - 是否有错误

---

##### ClearErrors

清除错误。

```go
func (th *TestHelper) ClearErrors()
```

---

#### 链式动作方法

所有链式方法都返回 `*TestHelper`，支持链式调用。

##### Type

输入文本。

```go
func (th *TestHelper) Type(text string) *TestHelper
```

---

##### Press

按下特殊按键。

```go
func (th *TestHelper) Press(key platform.SpecialKey) *TestHelper
```

---

##### PressKey

按下字符键。

```go
func (th *TestHelper) PressKey(key rune) *TestHelper
```

---

##### Click

点击。

```go
func (th *TestHelper) Click(x, y int) *TestHelper
```

---

##### Tab

按 Tab 键。

```go
func (th *TestHelper) Tab() *TestHelper
```

---

##### Enter

按 Enter 键。

```go
func (th *TestHelper) Enter() *TestHelper
```

---

##### Escape

按 Escape 键。

```go
func (th *TestHelper) Escape() *TestHelper
```

---

##### Process

处理所有事件。

```go
func (th *TestHelper) Process() *TestHelper
```

---

##### Wait

等待一段时间。

```go
func (th *TestHelper) Wait(d time.Duration) *TestHelper
```

---

#### 链式断言方法

##### AssertRender

断言渲染包含文本。

```go
func (th *TestHelper) AssertRender(text string) *TestHelper
```

---

##### AssertNotRender

断言渲染不包含文本。

```go
func (th *TestHelper) AssertNotRender(text string) *TestHelper
```

---

##### Result

完成链式调用并返回结果。

```go
func (th *TestHelper) Result() TestResult
```

**返回：**
- `TestResult` - 测试结果

---

### TestResult

测试结果。

```go
type TestResult struct {
    Errors []error
}
```

#### 方法

##### OK

检查是否成功（无错误）。

```go
func (r TestResult) OK() bool
```

**返回：**
- `bool` - 是否成功

---

##### Error

返回第一个错误。

```go
func (r TestResult) Error() error
```

**返回：**
- `error` - 第一个错误（如果没有则为 nil）

---

## 真实沙箱

### RealSandbox

真实终端沙箱实现。

```go
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
```

#### 构造函数

##### New

创建真实沙箱。

```go
func New(width, height int) (*RealSandbox, error)
```

**参数：**
- `width` - 缓冲区宽度
- `height` - 缓冲区高度

**返回：**
- `*RealSandbox` - 真实沙箱
- `error` - 错误

---

#### 方法

RealSandbox 实现了 Sandbox 和 EventSource 接口的所有方法，以及额外的录制功能。

**Sandbox 接口：**
- `Initialize(config *Config) error`
- `Run() error`
- `Pause() error`
- `Resume() error`
- `Close() error`
- `State() State`
- `Type() SandboxType`
- `Config() *Config`
- `Buffer() *paint.Buffer`
- `SetBuffer(buf *paint.Buffer)`
- `Resize(width, height int)`
- `Size() (int, int)`

**EventSource 接口：**
- `Events() <-chan platform.RawInput`
- （Start/Stop 由内部管理）

**Snapshotter 接口：**
- `Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)`
- `Restore(snap *Snapshot) error`
- `ListSnapshots() []*SnapshotMetadata`

**额外方法：**
- `RecordedEvents() []platform.RawInput` - 获取录制的事件

---

## 回放沙箱

### ReplaySandbox

回放沙箱实现。

```go
type ReplaySandbox struct {
    mu     sync.RWMutex

    player *Player
    buffer *paint.Buffer
    config *sandbox.Config
}
```

#### 构造函数

##### New

创建回放沙箱。

```go
func New(events []platform.RawInput, width, height int) *ReplaySandbox
```

**参数：**
- `events` - 要回放的事件列表
- `width` - 缓冲区宽度
- `height` - 缓冲区高度

**返回：**
- `*ReplaySandbox` - 回放沙箱

---

#### 方法

**Sandbox 接口：**
- `Initialize(config *Config) error`
- `Run() error`
- `Pause() error`
- `Resume() error`
- `Close() error`
- `State() State`
- `Type() SandboxType`
- `Config() *Config`
- `Buffer() *paint.Buffer`
- `SetBuffer(buf *paint.Buffer)`
- `Resize(width, height int)`
- `Size() (int, int)`

**Snapshotter 接口：**
- `Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)`
- `Restore(snap *Snapshot) error`
- `ListSnapshots() []*SnapshotMetadata`

**额外方法：**
- `SetSpeed(speed float64)` - 设置回放速度
- `GetSpeed() float64` - 获取回放速度
- `Step() (platform.RawInput, error)` - 前进一步
- `StepBack() (platform.RawInput, error)` - 后退一步

---

### Player

事件回放器。

```go
type Player struct {
    mu      sync.RWMutex
    events  []platform.RawInput
    index   int
    speed   float64
    playing bool
}
```

#### 构造函数

##### NewPlayer

创建回放器。

```go
func NewPlayer(events []platform.RawInput) *Player
```

**参数：**
- `events` - 要回放的事件列表

**返回：**
- `*Player` - 回放器

---

#### 方法

##### Play

开始回放。

```go
func (p *Player) Play() error
```

**返回：**
- `error` - 错误

---

##### Pause

暂停回放。

```go
func (p *Player) Pause() error
```

**返回：**
- `error` - 错误

---

##### Stop

停止回放。

```go
func (p *Player) Stop() error
```

**返回：**
- `error` - 错误

---

##### Seek

跳转到指定索引。

```go
func (p *Player) Seek(index int) error
```

**参数：**
- `index` - 目标索引

**返回：**
- `error` - 错误

---

##### Next

下一个事件。

```go
func (p *Player) Next() (platform.RawInput, error)
```

**返回：**
- `platform.RawInput` - 事件
- `error` - 错误

---

##### Previous

上一个事件。

```go
func (p *Player) Previous() (platform.RawInput, error)
```

**返回：**
- `platform.RawInput` - 事件
- `error` - 错误

---

##### Current

当前事件。

```go
func (p *Player) Current() (platform.RawInput, error)
```

**返回：**
- `platform.RawInput` - 事件
- `error` - 错误

---

##### SetSpeed

设置回放速度。

```go
func (p *Player) SetSpeed(speed float64)
```

**参数：**
- `speed` - 速度倍数（1.0 = 正常）

---

##### Speed

获取回放速度。

```go
func (p *Player) Speed() float64
```

**返回：**
- `float64` - 速度倍数

---

##### IsPlaying

是否正在播放。

```go
func (p *Player) IsPlaying() bool
```

**返回：**
- `bool` - 是否正在播放

---

##### Index

当前索引。

```go
func (p *Player) Index() int
```

**返回：**
- `int` - 当前索引

---

##### Length

事件总数。

```go
func (p *Player) Length() int
```

**返回：**
- `int` - 事件总数

---

##### HasNext

是否有下一个事件。

```go
func (p *Player) HasNext() bool
```

**返回：**
- `bool` - 是否有下一个事件

---

##### HasPrevious

是否有上一个事件。

```go
func (p *Player) HasPrevious() bool
```

**返回：**
- `bool` - 是否有上一个事件

---

##### Reset

重置到开始。

```go
func (p *Player) Reset()
```

---

## 错误处理

### 预定义错误

```go
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
```

### AssertionError

断言错误详情。

```go
type AssertionError struct {
    Message  string
    Expected interface{}
    Actual   interface{}
    Selector string
}
```

#### 方法

##### Error

返回错误信息。

```go
func (e *AssertionError) Error() string
```

**返回：**
- `string` - 错误信息

---

## UI 集成

### TestApp

测试应用包装器。

```go
type TestApp struct {
    sandbox *mock.MockSandbox
    app     interface{}  // 实际应用
}
```

#### 构造函数

##### TestRun

运行测试应用。

```go
func TestRun(app interface{}, opts ...TestOption) (*TestApp, error)
```

**参数：**
- `app` - 应用实例
- `opts` - 测试选项

**返回：**
- `*TestApp` - 测试应用
- `error` - 错误

---

##### TestRunWithConfig

使用自定义配置运行测试应用。

```go
func TestRunWithConfig(app interface{}, config *sandbox.Config) (*TestApp, error)
```

**参数：**
- `app` - 应用实例
- `config` - 沙箱配置

**返回：**
- `*TestApp` - 测试应用
- `error` - 错误

---

#### 方法

##### Close

关闭测试应用。

```go
func (ta *TestApp) Close() error
```

**返回：**
- `error` - 错误

---

##### Sandbox

获取沙箱。

```go
func (ta *TestApp) Sandbox() *mock.MockSandbox
```

**返回：**
- `*mock.MockSandbox` - Mock 沙箱

---

##### Helper

获取测试辅助器。

```go
func (ta *TestApp) Helper() *mock.TestHelper
```

**返回：**
- `*mock.TestHelper` - 测试辅助器

---

### TestOption

测试选项类型。

```go
type TestOption func(*testConfig)
```

#### 可用选项

##### TestWithWidth

设置测试宽度。

```go
func TestWithWidth(w int) TestOption
```

**参数：**
- `w` - 宽度

**返回：**
- `TestOption` - 测试选项

---

##### TestWithHeight

设置测试高度。

```go
func TestWithHeight(h int) TestOption
```

**参数：**
- `h` - 高度

**返回：**
- `TestOption` - 测试选项

---

##### TestWithSize

设置测试尺寸。

```go
func TestWithSize(w, h int) TestOption
```

**参数：**
- `w` - 宽度
- `h` - 高度

**返回：**
- `TestOption` - 测试选项

---

## 附录

### A. 平台常量

#### 特殊按键

```go
const (
    KeyTab       platform.SpecialKey = iota
    KeyEnter
    KeyEscape
    KeySpace
    KeyBackspace
    KeyDelete
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    // ... 更多按键
)
```

#### 修饰符

```go
const (
    KeyModShift platform.KeyModifier = 1 << iota
    KeyModCtrl
    KeyModAlt
    KeyModMeta
)
```

#### 鼠标按钮

```go
const (
    MouseLeft   platform.MouseButton = iota
    MouseMiddle
    MouseRight
)
```

#### 鼠标动作

```go
const (
    MousePress   platform.MouseAction = iota
    MouseRelease
    MouseScroll
)
```

---

### B. 类型转换

```go
// SandboxType -> string
TypeReal.String()    // "real"
TypeMock.String()    // "mock"
TypeReplay.String()  // "replay"

// State -> string
StateStopped.String()       // "stopped"
StateInitialized.String()   // "initialized"
StateRunning.String()       // "running"
StatePaused.String()        // "paused"
StateError.String()         // "error"

// InjectionStrategy -> string
InjectProhibited.String()  // "prohibited"
InjectAllowed.String()     // "allowed"
InjectRecorded.String()    // "recorded"

// EvictPolicy -> string
EvictOldest.String()     // "oldest"
EvictByPriority.String() // "priority"
EvictPersist.String()    // "persist"

// SnapshotLevel -> string
SnapshotMinimal.String()   // "minimal"
SnapshotStandard.String()  // "standard"
SnapshotFull.String()      // "full"
```

---

### C. 依赖关系

```
sandbox/
├── runtime/platform  ✅ 依赖（RawInput, InputReader 等）
├── runtime/paint     ✅ 依赖（Buffer 等）
├── runtime/event     ❌ 不依赖（避免循环）
└── runtime/engine    ❌ 不依赖（避免循环）

runtime/engine/
└── sandbox            ✅ 可选依赖（测试模式）
```

---

**API 参考手册结束**

如有问题，请参考使用手册或提交 Issue。
