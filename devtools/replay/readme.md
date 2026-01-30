# Replay - 确定性回放模块

> 事件录制、回放引擎、确定性验证

## 功能概述

Replay 模块提供完整的事件录制和回放功能，用于问题复现、Bug 报告和确定性验证。

## 核心组件

### 1. Event Recorder (`recorder.go`)

```go
// EventRecorder 事件录制器
type EventRecorder struct {
    mu          sync.RWMutex
    sessions    map[string]*RecordingSession
    currentSession *string
}

// RecordingSession 录制会话
type RecordingSession struct {
    ID        string
    StartTime time.Time
    EndTime   time.Time
    Events    []RecordedEvent
    Seeds     map[string]uint64  // 随机种子
    Inputs    []InputEvent
}

// RecordedEvent 录制的事件
type RecordedEvent struct {
    FrameID   devtools.FrameID
    Type      string
    NodeID    devtools.NodeID
    Phase     string
    Data      map[string]interface{}
}

// 开始录制
func (r *EventRecorder) StartSession(id string) error

// 记录事件
func (r *EventRecorder) RecordEvent(event *RecordedEvent)

// 结束录制
func (r *EventRecorder) EndSession(id string) (*RecordingSession, error)
```

### 2. Event Replayer (`replayer.go`)

```go
// EventReplayer 事件回放器
type EventReplayer struct {
    mu       sync.RWMutex
    session  *RecordingSession
    position int
    speed    float64  // 回放速度
    paused   bool
}

// 创建回放器
func NewReplayer(session *RecordingSession) *EventReplayer

// 开始回放
func (r *EventReplayer) Start() error

// 下一步
func (r *EventReplayer) Next() error

// 暂停/继续
func (r *EventReplayer) Pause()
func (r *EventReplayer) Resume()

// 设置速度
func (r *EventReplayer) SetSpeed(speed float64)
```

### 3. Determinism Checker (`determinism.go`)

```go
// DeterminismChecker 确定性验证器
type DeterminismChecker struct {
    mu           sync.RWMutex
    original     *RecordingSession
    replay       *RecordingSession
    differences  []Difference
}

// Difference 差异
type Difference struct {
    FrameID   devtools.FrameID
    NodeID    devtools.NodeID
    Path      string
    Original  interface{}
    Replay    interface{}
}

// 比较原始和回放
func (d *DeterminismChecker) Compare() ([]Difference, error)

// 验证确定性
func (d *DeterminismChecker) Verify() bool
```

### 4. Input Capture (`input.go`)

```go
// InputCapture 输入捕获器
type InputCapture struct {
    mu     sync.RWMutex
    inputs []InputEvent
}

// InputEvent 输入事件
type InputEvent struct {
    Timestamp time.Time
    Type      string  // "key", "mouse", "resize"
    Data      map[string]interface{}
}

// 捕获输入
func (ic *InputCapture) Capture(event InputEvent)

// 获取输入序列
func (ic *InputCapture) GetSequence() []InputEvent
```

### 5. Seed Capture (`seed.go`)

```go
// SeedCapture 随机种子捕获器
type SeedCapture struct {
    mu    sync.RWMutex
    seeds map[string]uint64
}

// 捕获种子
func (sc *SeedCapture) Capture(name string, seed uint64)

// 获取种子
func (sc *SeedCapture) Get(name string) (uint64, bool)

// 导出/导入
func (sc *SeedCapture) Export() map[string]uint64
func (sc *SeedCapture) Import(seeds map[string]uint64)
```

## 使用方法

### 录制事件

```go
import "github.com/wwsheng009/mint/devtools/replay"

// 创建录制器
recorder := replay.NewEventRecorder()

// 开始录制会话
sessionID := "bug-report-001"
recorder.StartSession(sessionID)

// 在应用循环中记录事件
recorder.RecordEvent(&replay.RecordedEvent{
    FrameID: 42,
    Type:    "keypress",
    NodeID:  "button-1",
    Phase:   "bubble",
    Data:    map[string]interface{}{"key": "Enter"},
})

// 结束录制
session, _ := recorder.EndSession(sessionID)

// 保存会话
data, _ := json.MarshalIndent(session, "", "  ")
os.WriteFile("bug-report.json", data, 0644)
```

### 回放事件

```go
// 创建回放器
replayer := replay.NewReplayer(session)

// 开始回放
replayer.Start()

// 单步执行
for replayer.Next() == nil {
    // 等待用户输入或自动延迟
    time.Sleep(time.Duration(100 * replayer.Speed()) * time.Millisecond)
}

// 暂停/继续
replayer.Pause()
// ... 分析状态 ...
replayer.Resume()

// 设置回放速度 (0.5x, 1.0x, 2.0x)
replayer.SetSpeed(0.5)
```

### 确定性验证

```go
// 创建验证器
checker := replay.NewDeterminismChecker(originalSession, replaySession)

// 比较差异
differences, _ := checker.Compare()

for _, diff := range differences {
    fmt.Printf("Frame %d, Node %s: %s\n",
        diff.FrameID, diff.NodeID, diff.Path)
    fmt.Printf("  Original: %v\n", diff.Original)
    fmt.Printf("  Replay:   %v\n", diff.Replay)
}

// 验证是否确定
isDeterministic := checker.Verify()
```

### 输入捕获

```go
// 创建输入捕获器
inputCapture := replay.NewInputCapture()

// 捕获键盘输入
inputCapture.Capture(replay.InputEvent{
    Timestamp: time.Now(),
    Type:      "key",
    Data:      map[string]interface{}{
        "key":  "Enter",
        "ctrl": false,
    },
})

// 捕获鼠标输入
inputCapture.Capture(replay.InputEvent{
    Timestamp: time.Now(),
    Type:      "mouse",
    Data:      map[string]interface{}{
        "x": 100,
        "y": 50,
        "button": "left",
    },
})
```

### 随机种子管理

```go
// 创建种子捕获器
seedCapture := replay.NewSeedCapture()

// 捕获随机数生成器种子
rng := rand.New(rand.NewSource(42))
seedCapture.Set("main-rng", rng.Uint64())

// 在回放时恢复种子
seed, _ := seedCapture.Get("main-rng")
replayRng := rand.New(rand.NewSource(int64(seed)))
```

## 宏录制/回放

```go
// 录制宏
recorder.StartSession("macro-submit-form")

// ... 执行操作 ...

session, _ := recorder.EndSession("macro-submit-form")

// 回放宏
replayer := replay.NewReplayer(session)
replayer.SetSpeed(2.0)  // 快速回放
replayer.Start()
for replayer.Next() == nil {}
```

## 相关模块

| 模块 | 关系 |
|------|------|
| `devtools` | 核心功能，录制和回放其事件 |
| `snapshot` | 状态快照，用于验证回放结果 |
| `timetravel` | 时间旅行，与回放协同工作 |

## API 参考

### RecordingSession

```go
type RecordingSession struct {
    ID        string
    StartTime time.Time
    EndTime   time.Time
    Events    []RecordedEvent
    Seeds     map[string]uint64
    Inputs    []InputEvent
    Metadata  map[string]interface{}
}
```

### ReplayerState

```go
type ReplayerState struct {
    Position int
    Paused   bool
    Speed    float64
    Loop     bool
}
```

## 文件列表

- `recorder.go` - 事件录制器
- `replayer.go` - 事件回放器
- `determinism.go` - 确定性验证器
- `input.go` - 输入捕获
- `seed.go` - 随机种子捕获
