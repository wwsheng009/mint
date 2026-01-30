# Memory - 内存优化模块

> 环形缓冲区、自适应采样、内存监控

## 功能概述

Memory 模块提供内存优化功能，确保 DevTools 在长时间运行时保持低内存占用，同时不牺牲调试能力。

## 核心组件

### 1. Ring Buffer (`ringbuffer.go`)

```go
// RingBuffer 固定大小的循环缓冲区
type RingBuffer struct {
    buffer   []devtools.FrameID
    capacity int
    head     int
    tail     int
    full     bool
}

// 创建环形缓冲区
func NewRingBuffer(capacity int) *RingBuffer

// 写入帧 ID
func (rb *RingBuffer) Write(frameID devtools.FrameID) error

// 读取一个帧
func (rb *RingBuffer) Read() (devtools.FrameID, error)

// 获取所有帧
func (rb *RingBuffer) GetAll() []devtools.FrameID

// 调整大小
func (rb *RingBuffer) Resize(newCapacity int)
```

**特性：**
- O(1) 写入和读取
- 固定内存占用
- 自动覆盖最旧数据
- 线程安全

### 2. Frame Window (`ringbuffer.go`)

```go
// FrameWindow 滑动窗口
type FrameWindow struct {
    ring    *RingBuffer
    maxSize int
}

// 创建窗口
func NewFrameWindow(maxSize int) *FrameWindow

// 添加帧
func (w *FrameWindow) Add(frameID devtools.FrameID)

// 获取窗口内的帧
func (w *FrameWindow) GetFrames() []devtools.FrameID

// 检查帧是否在窗口内
func (w *FrameWindow) Contains(frameID devtools.FrameID) bool
```

### 3. Sampling Strategies (`sampling.go`)

```go
// SamplingStrategy 采样策略接口
type SamplingStrategy interface {
    ShouldSample(frameID devtools.FrameID, context *SamplingContext) bool
    GetSamplingRate() float64
    SetSamplingRate(rate float64)
}

// AdaptiveStrategy 自适应采样
func NewAdaptiveStrategy(minRate, maxRate float64) *AdaptiveStrategy

// FixedRateStrategy 固定速率采样
func NewFixedRateStrategy(rate float64) *FixedRateStrategy

// PriorityStrategy 优先级采样
func NewPriorityStrategy(defaultRate float64) *PriorityStrategy
```

**策略对比：**

| 策略 | 适用场景 | 内存节省 |
|------|---------|---------|
| FixedRate | 稳定环境 | 固定比例 |
| Adaptive | 动态负载 | 根据压力调整 |
| Priority | 关键组件 | 关键数据不丢失 |

### 4. Memory Monitor (`monitor.go`)

```go
// Monitor 内存使用监控器
type Monitor struct {
    mu                sync.RWMutex
    sampleInterval    time.Duration
    alertCallback     AlertCallback
    alertThreshold    float64
    warningThreshold  float64
    criticalThreshold float64
}

// 创建监控器
func NewMonitor() *Monitor

// 启动监控
func (m *Monitor) Start()

// 停止监控
func (m *Monitor) Stop()

// 设置告警回调
func (m *Monitor) SetAlertCallback(callback AlertCallback)

// 获取当前使用率
func (m *Monitor) GetCurrentUsage() float64
```

## 使用方法

### 环形缓冲区

```go
import "github.com/wwsheng009/mint/devtools/memory"

// 创建容量为 1000 的环形缓冲区
ring := memory.NewRingBuffer(1000)

// 写入帧 ID
for i := 0; i < 1500; i++ {
    ring.Write(devtools.FrameID(i))
}

// 读取所有帧 (最多 1000 个)
frames := ring.GetAll()
fmt.Printf("Stored %d frames\n", len(frames))

// 检查状态
fmt.Printf("Full: %v, Size: %d\n", ring.IsFull(), ring.Size())

// 清空
ring.Clear()
```

### 滑动窗口

```go
// 创建最近 100 帧的窗口
window := memory.NewFrameWindow(100)

// 添加帧
window.Add(devtools.FrameID(1))
window.Add(devtools.FrameID(2))

// 获取窗口内的所有帧
frames := window.GetFrames()

// 检查帧是否在窗口内
exists := window.Contains(devtools.FrameID(1))
```

### 自适应采样

```go
// 创建自适应策略 (10%-100% 采样率)
strategy := memory.NewAdaptiveStrategy(0.1, 1.0)

// 检查是否应该采样
context := &memory.SamplingContext{
    FrameID:        42,
    EventCount:     5,
    MutationCount:  3,
    LayoutCount:    1,
    MemoryPressure: 0.7,
}

if strategy.ShouldSample(devtools.FrameID(42), context) {
    // 捕获详细数据
    captureDetailedMetrics()
} else {
    // 只记录基本信息
    captureBasicMetrics()
}
```

### 内存监控

```go
// 创建监控器
monitor := memory.NewMonitor()
monitor.SetSampleInterval(5 * time.Second)
monitor.SetThresholds(0.7, 0.9) // 70% 警告, 90% 严重

// 设置告警回调
monitor.SetAlertCallback(func(alert memory.MemoryAlert) {
    log.Printf("[%s] Memory: %.1f%% - %s",
        alert.Level, alert.Usage*100, alert.Message)
})

// 启动监控
monitor.Start()
defer monitor.Stop()
```

## 内存优化策略

### 分层存储

```
┌─────────────────────────────────────────┐
│  热数据 (内存) - 最近 100 帧            │
│  • 完整组件状态                         │
│  • 快速访问                             │
├─────────────────────────────────────────┤
│  温数据 (压缩) - 100-1000 帧           │
│  • 增量存储                             │
├─────────────────────────────────────────┤
│  冷数据 (磁盘) - 1000+ 帧              │
│  • 持久化存储                           │
└─────────────────────────────────────────┘
```

### 自适应采样级别

| 内存压力 | 采样率 | 策略 |
|---------|-------|------|
| < 50% | 100% | 全量捕获 |
| 50-70% | 50% | 隔帧捕获 |
| 70-90% | 25% | 关键帧 + 变更组件 |
| > 90% | 10% | 仅关键帧 |

## 告警级别

```go
type AlertLevel int

const (
    AlertLevelInfo AlertLevel = iota
    AlertLevelWarning
    AlertLevelCritical
)
```

## 相关模块

| 模块 | 关系 |
|------|------|
| `devtools` | 核心功能，Memory 模块优化其内存使用 |
| `snapshot` | 快照存储，使用环形缓冲区管理 |
| `observation` | 采样数据，受采样策略控制 |

## API 参考

### SamplingContext

```go
type SamplingContext struct {
    FrameID        devtools.FrameID
    EventCount     int
    MutationCount  int
    LayoutCount    int
    MemoryPressure float64
}
```

### MemoryAlert

```go
type MemoryAlert struct {
    Level       AlertLevel
    Usage       float64
    Message     string
    Timestamp   time.Time
    Suggestions []string
}
```

## 文件列表

- `ringbuffer.go` - 环形缓冲区和滑动窗口
- `sampling.go` - 采样策略
- `monitor.go` - 内存监控器
