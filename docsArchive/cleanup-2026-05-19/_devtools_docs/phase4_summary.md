# DevTools 阶段4 实施总结

> 实施日期: 2026-01-30
> 状态: 已完成
> 阶段: 确定性回放 (P2)
> 编译: ✅ 通过
> 测试: ✅ 16/16 通过

## 已完成的工作

### 1. EventRecorder 事件录制

| 文件 | 功能 | 状态 |
|------|------|------|
| `replay/recorder.go` | `EventRecorder` 事件录制器 | ✅ |
| `replay/recorder.go` | `RecordingSession` 录制会话 | ✅ |
| `replay/recorder.go` | `RecordedEvent` 录制事件 | ✅ |
| `replay/recorder.go` | `InputEvent` 输入事件 | ✅ |
| `replay/recorder.go` | `InputBuilder` 输入构建器 | ✅ |
| `replay/recorder.go` | JSON 导出/导入 | ✅ |
| `replay/recorder.go` | 会话分割/合并 | ✅ |

### 2. EventReplayer 事件回放

| 文件 | 功能 | 状态 |
|------|------|------|
| `replay/replayer.go` | `EventReplayer` 事件回放器 | ✅ |
| `replay/replayer.go` | `ReplayProgress` 回放进度 | ✅ |
| `replay/replayer.go` | 回放速度控制 | ✅ |
| `replay/replayer.go` | 实时模式/快进模式 | ✅ |
| `replay/replayer.go` | 暂停/恢复/单步执行 | ✅ |
| `replay/replayer.go` | `ReplaySession` 回放会话 | ✅ |
| `replay/replayer.go` | 断点功能 | ✅ |
| `replay/replayer.go` | 验证和比较 | ✅ |

### 3. DeterminismChecker 确定性验证

| 文件 | 功能 | 状态 |
|------|------|------|
| `replay/determinism.go` | `DeterminismChecker` 确定性检查器 | ✅ |
| `replay/determinism.go` | `Checkpoint` 状态检查点 | ✅ |
| `replay/determinism.go` | `VerifyCheckpoint` 验证检查点 | ✅ |
| `replay/determinism.go` | `VerifyFull` 完整验证 | ✅ |
| `replay/determinism.go` | `CompareSessions` 会话比较 | ✅ |
| `replay/determinism.go` | 状态哈希验证 | ✅ |
| `replay/determinism.go` | `GenerateReport` 生成报告 | ✅ |

### 4. RandomSeedCapture 随机种子捕获

| 文件 | 功能 | 状态 |
|------|------|------|
| `replay/seed.go` | `SeedTracker` 种子跟踪器 | ✅ |
| `replay/seed.go` | `SeedSnapshot` 种子快照 | ✅ |
| `replay/seed.go` | `SeedHistory` 种子历史 | ✅ |
| `replay/seed.go` | `DeterministicSeed` 确定性种子生成器 | ✅ |
| `replay/seed.go` | 多源种子支持 | ✅ |

### 5. InputCapture 输入捕获

| 文件 | 功能 | 状态 |
|------|------|------|
| `replay/input.go` | `InputCapture` 输入捕获器 | ✅ |
| `replay/input.go` | `InputFilter` 输入过滤器 | ✅ |
| `replay/input.go` | 键盘/鼠标事件捕获 | ✅ |
| `replay/input.go` | `InputSequence` 输入序列 | ✅ |
| `replay/input.go` | `Macro` 宏录制 | ✅ |
| `replay/input.go` | `MacroRegistry` 宏注册表 | ✅ |

---

## 架构设计

### 确定性回放系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    确定性回放系统架构                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Recording Phase                       │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │  │ EventRecorder│  │ InputCapture │  │ SeedTracker  │   │    │
│  │  │              │  │              │  │              │   │    │
│  │  │ • RecordEvent│  │ • KeyPress   │  │ • Capture    │   │    │
│  │  │ • RecordInput│  │ • MouseMove  │  │ • SeedValue  │   │    │
│  │  │ • Seeds      │  │ • Filter     │  │ • History    │   │    │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │    │
│  │         │                  │                  │            │    │
│  │         └──────────────────┼──────────────────┘            │    │
│  │                            ▼                               │    │
│  │                 ┌────────────────────┐                    │    │
│  │                 │ RecordingSession  │                    │    │
│  │                 │ • Events          │                    │    │
│  │                 │ • Inputs          │                    │    │
│  │                 │ • Seeds           │                    │    │
│  │                 │ • Metadata        │                    │    │
│  │                 └────────┬───────────┘                    │    │
│  └────────────────────────────┼───────────────────────────────┘    │
│                              │                                     │
│                              ▼                                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                      Storage                            │    │
│  │  • JSON File                                           │    │
│  │  • Binary Format                                       │    │
│  │  • Memory                                             │    │
│  └────────────────────────────┬────────────────────────────┘    │
│                              │                                     │
│                              ▼                                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Replay Phase                         │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │  │EventReplayer │  │DeterminismChk│  │ReplaySession │   │    │
│  │  │              │  │              │  │              │   │    │
│  │  │ • ReplayFrom │  │ • Verify     │  │ • Breakpoints│   │    │
│  │  │ • SpeedCtrl  │  │ • Compare    │  │ • StepFrame  │   │    │
│  │  │ • RealTime   │  │ • Checkpoint │  │ • Progress   │   │    │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │    │
│  │         │                  │                  │            │    │
│  │         └──────────────────┼──────────────────┘            │    │
│  │                            ▼                               │    │
│  │                 ┌────────────────────┐                    │    │
│  │                 │  VerificationReport│                   │    │
│  │                 │  • Match Rate      │                    │    │
│  │                 │  • Mismatches      │                    │    │
│  │                 │  • Issues         │                    │    │
│  │                 └────────────────────┘                    │    │
│  └────────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 录制会话数据结构

```
┌─────────────────────────────────────────────────────────────────┐
│                   RecordingSession 结构                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  metadata:                                                       │
│    SessionID   string                                            │
│    StartTime   time.Time                                         │
│    EndTime     time.Time                                         │
│    Metadata    map[string]interface{}                            │
│                                                                  │
│  events: []RecordedEvent                                         │
│    ┌────────────────────────────────────────────────────┐       │
│    │ Seq         int64                                     │       │
│    │ Timestamp   time.Time                                 │       │
│    │ FrameID     devtools.FrameID                          │       │
│    │ Type        string                                    │       │
│    │ Data        map[string]interface{}                    │       │
│    │ CausalID    uint64                                    │       │
│    └────────────────────────────────────────────────────┘       │
│                                                                  │
│  inputs: []InputEvent                                            │
│    ┌────────────────────────────────────────────────────┐       │
│    │ Timestamp   time.Time                                 │       │
│    │ Type        InputType (KeyPress/KeyRelease/...)       │       │
│    │ Key         rune                                      │       │
│    │ MouseButton MouseButton                               │       │
│    │ Position    {X, Y}                                    │       │
│    │ Modifiers   {Ctrl, Alt, Shift, Meta}                 │       │
│    └────────────────────────────────────────────────────┘       │
│                                                                  │
│  seeds: []SeedSnapshot                                           │
│    ┌────────────────────────────────────────────────────┐       │
│    │ FrameID     devtools.FrameID                          │       │
│    │ Timestamp   time.Time                                 │       │
│    │ Source      string (math/rand/crypto/rand/...)       │       │
│    │ Value       int64                                      │       │
│    │ State       []byte (optional full state)              │       │
│    └────────────────────────────────────────────────────┘       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 文件结构

```
mint/
└── devtools/
    ├── types.go              # 核心类型
    ├── causal.go             # 阶段2: Causal Graph
    ├── causal_builder.go     # 阶段2: CausalBuilder
    ├── timeline.go           # 阶段2: FrameTimeline
    ├── causal_query.go       # 阶段2: Query API
    ├── component_hook.go     # 阶段2: Component Hooks
    │
    ├── timetravel/           # 阶段3: 时间旅行
    │   ├── snapshot.go       # 快照管理
    │   ├── cursor.go         # 时间游标
    │   ├── replay.go         # 状态回放
    │   ├── diffengine.go     # 差异引擎
    │   └── client.go         # TUI 客户端
    │
    └── replay/               # ✨ 阶段4: 确定性回放
        ├── recorder.go       # 事件录制
        ├── replayer.go       # 事件回放
        ├── determinism.go    # 确定性验证
        ├── seed.go           # 种子跟踪
        └── input.go          # 输入捕获
```

---

## 使用示例

### 1. 事件录制

```go
import "github.com/wwsheng009/mint/devtools/replay"

// 创建录制器
recorder := replay.NewEventRecorder()

// 开始录制
recorder.Start("session_001")

// 在主循环中
func FrameLoop() {
    frameID := devtools.FrameID(frameCounter)

    // 记录事件
    for _, event := range inputEvents {
        recorder.RecordEvent(frameID, event.Type, event.Data, causalID)
    }

    // 记录用户输入
    recorder.RecordInput(replay.NewInputBuilder().
        KeyPress('a').
        Modifiers(false, false, false, false).
        Build())
}

// 停止录制
session, err := recorder.Stop()
if err != nil {
    log.Fatal(err)
}

// 保存到文件
session.Save("session_001.json")
```

### 2. 事件回放

```go
// 加载录制会话
session, err := replay.Load("session_001.json")
if err != nil {
    log.Fatal(err)
}

// 创建回放器
replayer := replay.NewEventReplayer(session)

// 设置回调
replayer.SetEventCallback(func(event replay.RecordedEvent) {
    fmt.Printf("Event: %s at frame %d\n", event.Type, event.FrameID)
})

replayer.SetInputCallback(func(input replay.InputEvent) {
    // 重放输入事件
})

replayer.SetFrameStartCallback(func(frameID devtools.FrameID) {
    fmt.Printf("Frame %d started\n", frameID)
})

// 开始回放
replayer.Start()

// 控制回放
replayer.SetReplaySpeed(2.0)  // 2倍速
replayer.Pause()
replayer.Resume()
replayer.Stop()

// 单步执行
for i := 0; i < 10; i++ {
    replayer.StepFrame()
}
```

### 3. 确定性验证

```go
// 创建确定性检查器
checker := replay.NewDeterminismChecker(originalSession)

// 启用检查
checker.Enable()

// 记录检查点
for _, frameID := range frameIDs {
    stateData := captureState()
    checker.RecordCheckpoint(frameID, stateData)
}

// 验证回放
replayer := replay.NewEventReplayer(originalSession)
replayer.Start()
replayer.WaitUntilComplete()

// 生成验证报告
report := checker.GenerateReport()
fmt.Printf("Determinism rate: %.2f%%\n", report.DeterminismRate)

for _, issue := range report.Issues {
    fmt.Printf("Issue at frame %d: %s\n", issue.FrameID, issue.Description)
}
```

### 4. 种子跟踪

```go
// 创建种子跟踪器
seedTracker := replay.NewSeedTracker()

// 捕获种子
seedTracker.Capture(frameID, "math/rand", rand.Int63())

// 或者带完整状态
seedTracker.CaptureWithState(frameID, "math/rand", seedValue, stateBytes)

// 获取帧的种子
if seed := seedTracker.GetSeedForFrame(frameID); seed != nil {
    fmt.Printf("Seed for frame %d: %d from %s\n",
        frameID, seed.Value, seed.Source)
}

// 保存到会话
seedTracker.SaveToSession(session)
```

### 5. 输入捕获

```go
// 创建输入捕获器
inputCapture := replay.NewInputCapture(1024)
inputCapture.Enable()

// 捕获键盘事件
inputCapture.CaptureKeyPress('a', replay.KeyModifier{
    Ctrl: true,
})

// 捕获鼠标事件
inputCapture.CaptureMousePress(replay.MouseLeft, 100, 200)
inputCapture.CaptureMouseMove(105, 205)
inputCapture.CaptureMouseRelease(replay.MouseLeft, 105, 205)

// 获取捕获的输入
inputs := inputCapture.GetCapturedInputs()

// 使用输入序列构建器
sequence := replay.NewInputSequence()
sequence.AddKeyPress('h').
    AddKeyPress('e').
    AddKeyPress('l').
    AddKeyPress('l').
    AddKeyPress('o').
    AddText(" world").
    AddMouseClick(replay.MouseLeft, 10, 10)

// 使用宏
macro := replay.NewMacro("test_macro")
macro.Record(sequence.GetInputs())
macro.Play(inputCapture)
```

### 6. 回放会话与断点

```go
// 创建回放会话
session := replay.NewReplaySession(recordingSession)

// 添加断点
session.AddBreakpoint(devtools.FrameID(100))
session.AddBreakpoint(devtools.FrameID(200))

// 开始回放
session.Replayer.Start()

// 检查断点
for session.Replayer.IsRunning() {
    currentFrame := session.CurrentFrame
    if session.HasBreakpoint(currentFrame) {
        fmt.Printf("Hit breakpoint at frame %d\n", currentFrame)
        session.Replayer.Pause()
        // 等待用户输入
        waitForUser()
        session.Replayer.Resume()
    }
    time.Sleep(100 * time.Millisecond)
}

// 获取进度
progress := session.Replayer.GetProgress()
fmt.Printf("Progress: %.1f%%\n", progress.PercentComplete)
```

---

## API 快速参考

### EventRecorder

```go
NewEventRecorder() *EventRecorder
Start(sessionID string) error
Stop() (*RecordingSession, error)
IsRecording() bool
RecordEvent(frameID FrameID, eventType string, data map[string]interface{}, causalID uint64)
RecordInput(input InputEvent)
GetSession() *RecordingSession
GetSeedTracker() *SeedTracker
```

### EventReplayer

```go
NewEventReplayer(session *RecordingSession) *EventReplayer
Load(session *RecordingSession)
Start() error
Stop()
Pause()
Resume()
IsRunning() bool
IsPaused() bool
SetReplaySpeed(speed float64)
SetRealTimeMode(enabled bool)
GetProgress() *ReplayProgress
StepFrame() error
JumpToFrame(frameID FrameID) error
```

### DeterminismChecker

```go
NewDeterminismChecker(original *RecordingSession) *DeterminismChecker
Enable()
Disable()
IsEnabled() bool
RecordCheckpoint(frameID FrameID, stateData []byte)
VerifyCheckpoint(frameID FrameID, stateData []byte) *VerificationResult
VerifyFull(replaySession *RecordingSession) *FullVerificationReport
GenerateReport() *DeterminismReport
```

### SeedTracker

```go
NewSeedTracker() *SeedTracker
Capture(frameID FrameID, source string, seed int64)
CaptureWithState(frameID FrameID, source string, seed int64, state []byte)
GetSeedForFrame(frameID FrameID) *SeedSnapshot
ApplySeed(source string, seed int64)
GetAllSeeds() []*SeedSnapshot
SaveToSession(session *RecordingSession)
LoadFromSession(session *RecordingSession)
```

### InputCapture

```go
NewInputCapture(bufferSize int) *InputCapture
Enable()
Disable()
IsEnabled() bool
SetFilter(filter InputFilter)
Capture(input InputEvent) bool
CaptureKeyPress(key rune, modifiers KeyModifier) bool
CaptureKeyRelease(key rune, modifiers KeyModifier) bool
CaptureMousePress(button MouseButton, x, y int) bool
CaptureMouseRelease(button MouseButton, x, y int) bool
CaptureMouseMove(x, y int) bool
GetCapturedInputs() []InputEvent
GetStats() *InputCaptureStats
```

---

## 设计特点

1. **完整录制**: 事件、输入、随机种子完整捕获
2. **确定性回放**: 支持种子恢复确保可重现性
3. **灵活控制**: 暂停、恢复、变速、单步执行
4. **断点调试**: 在指定帧暂停检查状态
5. **验证机制**: 状态哈希验证回放正确性
6. **输入过滤**: 可配置的输入捕获过滤器
7. **宏支持**: 输入序列录制和重放

---

## 下一步 (阶段5: 客户端集成)

- [ ] TUI 调试面板集成
- [ ] Web Dashboard
- [ ] WebSocket 协议
- [ ] 远程调试支持

---

## 验收检查清单

### 编译与测试
- [x] replay 包编译通过
- [x] 整个项目编译通过
- [x] 16/16 单元测试通过
- [x] 无循环依赖

### 功能实现
- [x] EventRecorder 事件录制已实现
- [x] EventReplayer 事件回放已实现
- [x] DeterminismChecker 确定性验证已实现
- [x] RandomSeedCapture 随机种子捕获已实现
- [x] InputCapture 输入捕获已实现

### 特性
- [x] JSON 序列化支持
- [x] 会话分割/合并
- [x] 回放进度跟踪
- [x] 断点功能
- [x] 宏录制/回放
- [x] 多源种子支持

---

## 总结

阶段4 完成了确定性回放系统的核心实现，提供了:

1. **完整录制**: 捕获所有影响应用状态的因素（事件、输入、随机）
2. **精确回放**: 恢复所有录制的内容，确保100%可重现
3. **验证机制**: 通过状态哈希验证回放的正确性
4. **调试支持**: 断点、单步、变速等调试功能
5. **扩展性**: 支持宏、输入过滤等高级功能

这些功能为自动化测试、bug复现、性能分析提供了强大的基础。
