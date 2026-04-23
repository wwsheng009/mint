# Sandbox 模块开发任务清单

> 基于 SANDBOX_DESIGN_V3.md 设计方案的详细实施任务

## 项目概览

| 属性 | 值 |
|------|-----|
| 预计工期 | 9-12 天 |
| 优先级 | 高 |
| 依赖 | runtime/platform, runtime/paint, runtime/event |
| 目标 | 提供可测试的 TUI 沙箱环境 |

---

## 阶段 1: 项目初始化与核心类型 (1 天)

### 1.1 创建目录结构
- [ ] 创建 `sandbox/` 根目录
- [ ] 创建 `sandbox/adapter/` 子目录
- [ ] 创建 `sandbox/mock/` 子目录
- [ ] 创建 `sandbox/real/` 子目录
- [ ] 创建 `sandbox/replay/` 子目录
- [ ] 创建 `sandbox/testing/` 子目录

### 1.2 实现 types.go
- [ ] 定义 `SandboxType` 枚举 (TypeReal, TypeMock, TypeReplay)
- [ ] 实现 `SandboxType.String()` 方法
- [ ] 定义 `State` 枚举 (StateStopped, StateInitialized, StateRunning, StatePaused, StateError)
- [ ] 实现 `State.String()` 方法
- [ ] 定义 `Phase` 枚举 (PhaseBefore, PhaseAfter)
- [ ] 定义 `HookKey` 结构体
- [ ] 定义 `InjectionStrategy` 枚举 (InjectProhibited, InjectAllowed, InjectRecorded)
- [ ] 定义 `EvictPolicy` 枚举 (EvictOldest, EvictByPriority, EvictPersist)
- [ ] 定义 `SnapshotLevel` 枚举 (SnapshotMinimal, SnapshotStandard, SnapshotFull)
- [ ] 定义 `InputEvent` 结构体 (包装 platform.RawInput)
- [ ] 定义 `BufferWrapper` 结构体

### 1.3 实现 errors.go
- [ ] 定义生命周期错误 (ErrInvalidTransition, ErrNotInitialized, ErrAlreadyRunning, ErrNotRunning)
- [ ] 定义事件注入错误 (ErrInjectionNotAllowed, ErrInvalidStrategy, ErrQueueFull, ErrQueueEmpty)
- [ ] 定义快照错误 (ErrSnapshotNotFound, ErrSnapshotCorrupt, ErrRestoreFailed)
- [ ] 定义配置错误 (ErrInvalidConfig)
- [ ] 定义断言错误 (ErrAssertionFailed, ErrTimeout)
- [ ] 实现 `AssertionError` 结构体及 `Error()` 方法

### 1.4 编写单元测试
- [ ] 测试所有枚举的 String() 方法
- [ ] 测试 AssertionError.Error()

**验收标准:**
- `go build ./sandbox/...` 编译通过
- `go test ./sandbox/...` 所有测试通过

---

## 阶段 2: 核心接口与生命周期 (1 天)

### 2.1 实现 sandbox.go (核心接口)
- [ ] 定义 `Sandbox` 接口
  - [ ] Initialize(config *Config) error
  - [ ] Run() error
  - [ ] Pause() error
  - [ ] Resume() error
  - [ ] Close() error
  - [ ] State() State
  - [ ] Type() SandboxType
  - [ ] Config() *Config
  - [ ] Buffer() *paint.Buffer
  - [ ] SetBuffer(buf *paint.Buffer)
  - [ ] Resize(width, height int)
  - [ ] Size() (width, height int)
- [ ] 定义 `EventSource` 接口 (用于真实环境)
  - [ ] Events() <-chan platform.RawInput
  - [ ] Start() error
  - [ ] Stop() error
- [ ] 定义 `EventSink` 接口 (用于测试环境)
  - [ ] Inject(event platform.RawInput) error
  - [ ] InjectKey(key rune) error
  - [ ] InjectSpecialKey(key platform.SpecialKey) error
  - [ ] InjectKeyWithMod(key rune, mod platform.KeyModifier) error
  - [ ] InjectMouse(x, y int, button, action) error
  - [ ] InjectResize(width, height int) error
  - [ ] InjectString(text string) error
  - [ ] ProcessEvents() error
- [ ] 定义 `Snapshotter` 接口
  - [ ] Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)
  - [ ] Restore(snap *Snapshot) error
  - [ ] ListSnapshots() []*SnapshotMetadata
- [ ] 定义 `TestSandbox` 组合接口
- [ ] 定义 `Renderer` 接口 (由 engine 实现，避免循环依赖)
- [ ] 定义 `EventDispatcher` 接口 (由 engine 实现)

### 2.2 实现 lifecycle.go
- [ ] 定义 `validTransitions` 状态转换表
- [ ] 实现 `Lifecycle` 结构体
  - [ ] mu sync.RWMutex
  - [ ] state State
  - [ ] err error
  - [ ] hooks map[HookKey][]HookFunc
- [ ] 定义 `HookFunc` 类型
- [ ] 实现 `NewLifecycle()` 构造函数
- [ ] 实现 `State()` 方法
- [ ] 实现 `Error()` 方法
- [ ] 实现 `Transition(to State)` 方法
- [ ] 实现 `isValidTransition(from, to State)` 方法
- [ ] 实现 `executeHooks(key HookKey)` 方法
- [ ] 实现 `OnTransition(state, phase, fn)` 方法
- [ ] 实现 `Reset()` 方法
- [ ] 实现 `CanTransitionTo(to State)` 方法

### 2.3 编写单元测试
- [ ] 测试有效状态转换
- [ ] 测试无效状态转换 (应返回 ErrInvalidTransition)
- [ ] 测试钩子执行顺序 (Before -> After)
- [ ] 测试钩子错误处理
- [ ] 测试并发安全性

**验收标准:**
- 所有状态转换符合设计
- 钩子机制正常工作
- 线程安全

---

## 阶段 3: 配置系统 (0.5 天)

### 3.1 实现 config.go
- [ ] 定义 `Config` 结构体
  - [ ] Width, Height, Title, FPS
  - [ ] Event EventConfig
  - [ ] Snapshot SnapshotConfig
  - [ ] Performance PerformanceConfig
- [ ] 定义 `EventConfig` 结构体
  - [ ] QueueMaxSize, QueueMaxMemory, EvictPolicy
  - [ ] Strategy, RecordEnabled, RecordMaxLen
- [ ] 定义 `SnapshotConfig` 结构体
  - [ ] AutoSnapshot, Interval, MaxCount, Level, PersistPath
- [ ] 定义 `PerformanceConfig` 结构体
  - [ ] Throttle, MaxFPS, RenderTimeout, Profile
- [ ] 实现 `DefaultConfig()` 函数
- [ ] 实现 `RealConfig()` 函数
- [ ] 实现 `MockConfig()` 函数
- [ ] 实现 `ReplayConfig()` 函数
- [ ] 实现 `Validate()` 方法
- [ ] 实现 `Clone()` 方法

### 3.2 编写单元测试
- [ ] 测试默认配置值
- [ ] 测试配置验证
- [ ] 测试配置克隆

**验收标准:**
- 配置系统完整
- 验证逻辑正确

---

## 阶段 4: 事件适配器层 (1 天)

### 4.1 实现 adapter/input.go
- [ ] 实现 `InputAdapter` 结构体
  - [ ] reader platform.InputReader
  - [ ] eventsCh chan platform.RawInput
  - [ ] stopCh chan struct{}
- [ ] 实现 `NewInputAdapter()` 构造函数
- [ ] 实现 `Start()` 方法
- [ ] 实现 `Stop()` 方法
- [ ] 实现 `Events()` 方法
- [ ] 实现 `ToSandboxEvent()` 辅助函数
- [ ] 实现 `BuildKeyEvent()` 辅助函数
- [ ] 实现 `BuildSpecialKeyEvent()` 辅助函数
- [ ] 实现 `BuildMouseEvent()` 辅助函数
- [ ] 实现 `BuildResizeEvent()` 辅助函数
- [ ] 实现 `BuildPasteEvent()` 辅助函数

### 4.2 编写单元测试
- [ ] 测试事件构建函数
- [ ] 测试适配器启动/停止

**验收标准:**
- 正确桥接 platform.RawInput
- 所有事件类型支持

---

## 阶段 5: 事件注入系统 (1.5 天)

### 5.1 实现 events.go
- [ ] 定义 `EventHandler` 类型
- [ ] 实现 `EventInjector` 结构体
  - [ ] mu sync.RWMutex
  - [ ] strategy InjectionStrategy
  - [ ] handler EventHandler
  - [ ] recorder *EventRecorder
- [ ] 实现 `NewEventInjector()` 构造函数
- [ ] 实现 `SetHandler()` 方法
- [ ] 实现 `SetRecorder()` 方法
- [ ] 实现 `Strategy()` 方法
- [ ] 实现 `SetStrategy()` 方法
- [ ] 实现 `Inject()` 方法 (根据策略分发)
- [ ] 实现 `injectProhibited()` 方法
- [ ] 实现 `injectAllowed()` 方法
- [ ] 实现 `injectRecorded()` 方法

### 5.2 实现 EventRecorder
- [ ] 实现 `EventRecorder` 结构体
  - [ ] mu sync.Mutex
  - [ ] events []platform.RawInput
  - [ ] maxLen int
- [ ] 实现 `NewEventRecorder()` 构造函数
- [ ] 实现 `Record()` 方法 (带淘汰)
- [ ] 实现 `Events()` 方法
- [ ] 实现 `Clear()` 方法
- [ ] 实现 `Len()` 方法

### 5.3 编写单元测试
- [ ] 测试 InjectProhibited 策略
- [ ] 测试 InjectAllowed 策略
- [ ] 测试 InjectRecorded 策略
- [ ] 测试策略动态切换
- [ ] 测试录制器淘汰策略
- [ ] 测试并发安全

**验收标准:**
- 三种注入策略正常工作
- 录制器内存可控

---

## 阶段 6: 有界事件队列 (1 天)

### 6.1 实现 mock/queue.go
- [ ] 定义 `QueueConfig` 结构体
- [ ] 实现 `DefaultQueueConfig()` 函数
- [ ] 实现 `BoundedQueue` 结构体
  - [ ] mu sync.RWMutex
  - [ ] config QueueConfig
  - [ ] events []platform.RawInput
  - [ ] memory int64
  - [ ] evictCount int64
- [ ] 实现 `NewBoundedQueue()` 构造函数
- [ ] 实现 `Push()` 方法 (带内存/容量检查)
- [ ] 实现 `Pop()` 方法
- [ ] 实现 `Peek()` 方法
- [ ] 实现 `Len()` 方法
- [ ] 实现 `IsEmpty()` 方法
- [ ] 实现 `Clear()` 方法
- [ ] 实现 `evictOne()` 方法
- [ ] 定义 `QueueStats` 结构体
- [ ] 实现 `Stats()` 方法
- [ ] 实现 `estimateEventSize()` 辅助函数

### 6.2 编写单元测试
- [ ] 测试基本入队出队
- [ ] 测试容量限制淘汰
- [ ] 测试内存限制淘汰
- [ ] 测试统计信息准确性
- [ ] 测试并发安全
- [ ] 压力测试 (10万事件)

**验收标准:**
- 内存占用可控
- 淘汰策略正确
- 性能达标 (10000+ 事件/秒)

---

## 阶段 7: 快照系统 (1.5 天)

### 7.1 实现 snapshot.go
- [ ] 定义 `Snapshot` 结构体
  - [ ] Metadata SnapshotMetadata
  - [ ] Buffer *paint.Buffer
  - [ ] Events []platform.RawInput
  - [ ] State map[string]interface{}
  - [ ] Checksum string
- [ ] 定义 `SnapshotMetadata` 结构体
- [ ] 定义 `SnapshotStorage` 接口
- [ ] 实现 `SnapshotManager` 结构体
  - [ ] mu sync.RWMutex
  - [ ] snapshots map[string]*Snapshot
  - [ ] order []string
  - [ ] maxCount int
  - [ ] storage SnapshotStorage
- [ ] 实现 `NewSnapshotManager()` 构造函数
- [ ] 实现 `SetStorage()` 方法
- [ ] 实现 `Create()` 方法 (按级别捕获)
- [ ] 实现 `Get()` 方法
- [ ] 实现 `List()` 方法
- [ ] 实现 `Delete()` 方法
- [ ] 实现 `Verify()` 方法

### 7.2 实现辅助函数
- [ ] 实现 `generateSnapshotID()` 函数
- [ ] 实现 `computeChecksum()` 函数
- [ ] 实现 `estimateSnapshotSize()` 函数
- [ ] 实现 `cloneBuffer()` 函数
- [ ] 实现 `cloneEvents()` 函数
- [ ] 实现 `cloneState()` 函数

### 7.3 编写单元测试
- [ ] 测试 Minimal 级别快照
- [ ] 测试 Standard 级别快照
- [ ] 测试 Full 级别快照
- [ ] 测试快照淘汰
- [ ] 测试校验和验证
- [ ] 测试克隆深度

**验收标准:**
- 三级快照正常工作
- 数据完整性保证
- 内存管理正确

---

## 阶段 8: Mock 沙箱实现 (2 天)

### 8.1 实现 mock/sandbox.go
- [ ] 实现 `MockSandbox` 结构体
  - [ ] mu sync.RWMutex
  - [ ] lifecycle *Lifecycle
  - [ ] config *Config
  - [ ] buffer *paint.Buffer
  - [ ] injector *EventInjector
  - [ ] queue *BoundedQueue
  - [ ] recorder *EventRecorder
  - [ ] snapMgr *SnapshotManager
  - [ ] eventHandler EventHandler
- [ ] 实现 `New()` 构造函数
- [ ] 实现 Sandbox 接口所有方法
  - [ ] Initialize()
  - [ ] Run()
  - [ ] Pause()
  - [ ] Resume()
  - [ ] Close()
  - [ ] State()
  - [ ] Type()
  - [ ] Config()
  - [ ] Buffer()
  - [ ] SetBuffer()
  - [ ] Resize()
  - [ ] Size()
- [ ] 实现 EventSink 接口所有方法
  - [ ] SetEventHandler()
  - [ ] Inject()
  - [ ] InjectKey()
  - [ ] InjectSpecialKey()
  - [ ] InjectKeyWithMod()
  - [ ] InjectMouse()
  - [ ] InjectResize()
  - [ ] InjectString()
  - [ ] ProcessEvents()
- [ ] 实现 Snapshotter 接口所有方法
  - [ ] Snapshot()
  - [ ] Restore()
  - [ ] ListSnapshots()
- [ ] 实现 TestSandbox 接口方法
  - [ ] IsMock()
  - [ ] AssertRender()
  - [ ] AssertNotRender()
  - [ ] RenderString()
  - [ ] Helper()
- [ ] 实现 `QueueStats()` 方法

### 8.2 实现 mock/testapi.go
- [ ] 实现 `TestHelper` 结构体
  - [ ] sandbox *MockSandbox
  - [ ] errors []error
- [ ] 实现 `NewTestHelper()` 构造函数
- [ ] 实现错误管理方法
  - [ ] Errors()
  - [ ] HasErrors()
  - [ ] ClearErrors()
- [ ] 实现 Action 方法 (链式调用)
  - [ ] Type()
  - [ ] Press()
  - [ ] PressKey()
  - [ ] Click()
  - [ ] Tab()
  - [ ] Enter()
  - [ ] Escape()
  - [ ] Process()
  - [ ] Wait()
- [ ] 实现 Assertion 方法
  - [ ] AssertRender()
  - [ ] AssertNotRender()
- [ ] 定义 `TestResult` 结构体
- [ ] 实现 `Result()` 方法
- [ ] 实现 `TestResult.OK()` 方法
- [ ] 实现 `TestResult.Error()` 方法

### 8.3 编写单元测试
- [ ] 测试沙箱生命周期
- [ ] 测试事件注入
- [ ] 测试字符串输入
- [ ] 测试快照创建/恢复
- [ ] 测试渲染断言
- [ ] 测试链式 API
- [ ] 测试 TestResult 模式

**验收标准:**
- 完整实现 TestSandbox 接口
- 链式 API 可用
- 所有测试通过

---

## 阶段 9: 真实沙箱实现 (1.5 天)

### 9.1 实现 real/sandbox.go
- [ ] 实现 `RealSandbox` 结构体
  - [ ] mu sync.RWMutex
  - [ ] lifecycle *Lifecycle
  - [ ] config *Config
  - [ ] buffer *paint.Buffer
  - [ ] input *InputAdapter
  - [ ] injector *EventInjector
  - [ ] recorder *EventRecorder
  - [ ] snapMgr *SnapshotManager
  - [ ] stopCh chan struct{}
- [ ] 实现 `New()` 构造函数
- [ ] 实现 Sandbox 接口所有方法
- [ ] 实现 `eventLoop()` 方法
- [ ] 实现 `handleEvent()` 方法
- [ ] 实现 EventSource 接口
  - [ ] Events()
- [ ] 实现 Snapshotter 接口
- [ ] 实现 `RecordedEvents()` 方法

### 9.2 编写单元测试
- [ ] 测试沙箱创建
- [ ] 测试生命周期管理
- [ ] 测试事件录制
- [ ] 测试快照功能

**验收标准:**
- 正确包装 platform.InputReader
- 终端恢复正确

---

## 阶段 10: 回放系统 (2 天)

### 10.1 实现 replay/sandbox.go
- [ ] 实现 `ReplaySandbox` 结构体
- [ ] 实现 Sandbox 接口
- [ ] 实现回放控制方法
  - [ ] SetSpeed()
  - [ ] GetSpeed()
  - [ ] Step()
  - [ ] StepBack()

### 10.2 实现 replay/player.go
- [ ] 实现 `Player` 结构体
  - [ ] events []platform.RawInput
  - [ ] index int
  - [ ] speed float64
  - [ ] playing bool
- [ ] 实现 `NewPlayer()` 构造函数
- [ ] 实现 `Play()` 方法
- [ ] 实现 `Pause()` 方法
- [ ] 实现 `Stop()` 方法
- [ ] 实现 `Seek()` 方法
- [ ] 实现 `Next()` 方法
- [ ] 实现 `Previous()` 方法

### 10.3 实现 replay/recorder.go
- [ ] 实现 `Recorder` 结构体
- [ ] 实现 `Recording` 结构体
  - [ ] Metadata RecordingMetadata
  - [ ] Events []platform.RawInput
  - [ ] Snapshots []*Snapshot
- [ ] 实现 `StartRecording()` 函数
- [ ] 实现 `StopRecording()` 方法
- [ ] 实现 `Save()` 方法
- [ ] 实现 `LoadRecording()` 函数

### 10.4 编写单元测试
- [ ] 测试录制功能
- [ ] 测试回放功能
- [ ] 测试速度控制
- [ ] 测试 Seek 功能

**验收标准:**
- 录制/回放功能完整
- 时间控制准确

---

## 阶段 11: 测试工具 (1 天)

### 11.1 实现 testing/runner.go
- [ ] 定义 `TestSuite` 结构体
- [ ] 定义 `TestCase` 结构体
- [ ] 实现 `Runner` 结构体
- [ ] 实现 `Run()` 方法

### 11.2 实现 testing/reporter.go
- [ ] 定义 `Reporter` 接口
- [ ] 实现 `ConsoleReporter`
- [ ] 实现 `JSONReporter`
- [ ] 实现 `JUnitReporter`

### 11.3 实现 testing/helpers.go
- [ ] 实现测试辅助函数
- [ ] 实现断言辅助函数

**验收标准:**
- 支持多种报告格式
- CI/CD 友好

---

## 阶段 12: UI 层集成 (0.5 天)

### 12.1 实现 ui/test.go
- [ ] 实现 `TestApp` 结构体
- [ ] 实现 `TestRun()` 函数
- [ ] 实现 `TestRunWithConfig()` 函数
- [ ] 实现 `Close()` 方法
- [ ] 实现 `Sandbox()` 方法
- [ ] 实现 `Helper()` 方法
- [ ] 定义 `TestOption` 类型
- [ ] 实现 `WithWidth()` 选项
- [ ] 实现 `WithHeight()` 选项
- [ ] 实现 `WithSize()` 选项

### 12.2 编写集成测试
- [ ] 测试完整用例流程
- [ ] 测试与现有组件集成

**验收标准:**
- UI 层集成完整
- 示例可运行

---

## 阶段 13: 文档与示例 (1 天)

### 13.1 编写文档
- [ ] 更新 README.md
- [ ] 编写 TESTING_GUIDE.md
- [ ] 编写 API 文档
- [ ] 添加代码注释

### 13.2 编写示例
- [ ] 创建 examples/sandbox_demo/
- [ ] 创建 examples/testing_demo/
- [ ] 添加注释说明

**验收标准:**
- 文档完整
- 示例可运行

---

## 验收检查清单

### 功能验收
- [ ] Mock 沙箱事件注入正常工作
- [ ] 真实沙箱终端操作正常
- [ ] 快照创建和恢复功能正常
- [ ] 回放功能可以复现已录制会话
- [ ] 链式测试 API 可用且错误处理正确
- [ ] 内存限制生效

### 兼容性验收
- [ ] 与 `runtime/event` 完全兼容
- [ ] 与 `runtime/platform` 完全兼容
- [ ] 与 `runtime/paint.Buffer` 完全兼容
- [ ] 与 `runtime/scheduler` 无冲突
- [ ] **无循环依赖** (关键)

### 性能验收
- [ ] Mock 沙箱内存占用可控 (< 100MB 默认)
- [ ] 事件队列支持 10000+ 事件/秒
- [ ] 快照操作 < 100ms

### 测试验收
- [ ] 单元测试覆盖率 > 80%
- [ ] 所有示例测试通过
- [ ] `go test ./sandbox/...` 全部通过
- [ ] `go build ./sandbox/...` 编译通过

---

## 依赖关系图

```
sandbox/
├── types.go, errors.go           ← 无依赖 (最先实现)
├── config.go                     ← 依赖 types.go
├── lifecycle.go                  ← 依赖 types.go, errors.go
├── sandbox.go (接口)             ← 依赖 types.go, runtime/paint, runtime/platform
├── events.go                     ← 依赖 types.go, errors.go, runtime/platform
├── snapshot.go                   ← 依赖 types.go, errors.go, runtime/paint, runtime/platform
│
├── adapter/
│   └── input.go                  ← 依赖 runtime/platform, sandbox (types)
│
├── mock/
│   ├── queue.go                  ← 依赖 sandbox (types, errors), runtime/platform
│   ├── sandbox.go                ← 依赖 sandbox/*, adapter, runtime/*
│   └── testapi.go                ← 依赖 mock/sandbox, runtime/platform
│
├── real/
│   └── sandbox.go                ← 依赖 sandbox/*, adapter, runtime/*
│
└── replay/
    ├── sandbox.go                ← 依赖 sandbox/*, runtime/*
    ├── player.go                 ← 依赖 sandbox (types), runtime/platform
    └── recorder.go               ← 依赖 sandbox (types, snapshot), runtime/platform
```

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 循环依赖 | 高 | 严格遵循依赖方向，sandbox 不依赖 engine |
| 内存泄漏 | 中 | 有界队列 + 快照淘汰 + 单元测试 |
| 终端恢复失败 | 高 | 复用现有 platform.RestoreTerminal() |
| 并发问题 | 中 | 所有共享状态使用 sync.RWMutex |
| 性能不达标 | 低 | 早期进行压力测试 |

---

## 进度跟踪

| 阶段 | 状态 | 开始日期 | 完成日期 | 备注 |
|------|------|----------|----------|------|
| 1. 项目初始化 | ✅ 完成 | 2025-02-01 | 2025-02-01 | 目录结构、types.go、errors.go、单元测试全部通过 |
| 2. 核心接口 | ✅ 完成 | 2025-02-01 | 2025-02-01 | sandbox.go 接口定义、lifecycle.go 状态机 |
| 3. 配置系统 | ✅ 完成 | 2025-02-01 | 2025-02-01 | config.go 配置系统 |
| 4. 事件适配器 | ✅ 完成 | 2025-02-01 | 2025-02-01 | adapter/input.go 事件适配器 |
| 5. 事件注入 | ✅ 完成 | 2025-02-01 | 2025-02-01 | events.go 事件注入系统 |
| 6. 有界队列 | ✅ 完成 | 2025-02-01 | 2025-02-01 | mock/queue.go 有界事件队列 |
| 7. 快照系统 | ✅ 完成 | 2025-02-01 | 2025-02-01 | snapshot.go 快照系统 |
| 8. Mock 沙箱 | ✅ 完成 | 2025-02-01 | 2025-02-01 | mock/sandbox.go、mock/testapi.go |
| 9. 真实沙箱 | ✅ 完成 | 2025-02-01 | 2025-02-01 | real/sandbox.go 真实沙箱 |
| 10. 回放系统 | ✅ 完成 | 2025-02-01 | 2025-02-01 | replay/sandbox.go、player.go、recorder.go |
| 11. 测试工具 | ✅ 完成 | 2025-02-01 | 2025-02-01 | testing/runner.go、reporter.go、helpers.go |
| 12. UI 集成 | ✅ 完成 | 2025-02-01 | 2025-02-01 | ui/test.go UI层集成 |
| 13. 文档示例 | ⏳ 进行中 | 2025-02-01 | | 进度更新、示例代码待补充 |
