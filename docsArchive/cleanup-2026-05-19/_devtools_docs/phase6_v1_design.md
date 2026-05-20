# DevTools 阶段6 V1 - 观测增强层

> **版本**: V1.0
> **日期**: 2026-01-30
> **状态**: 设计阶段

---

## 一、概述

阶段6 V1 实现轻量级观测增强层，不依赖代码重写，仅基于现有 DevTools 数据进行智能分析。

### 1.1 目标

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 热点检测 | 识别性能瓶颈（慢帧、慢组件） | P0 |
| 浪费检测 | 识别无效渲染（无变化的布局/重绘） | P1 |
| 抖动检测 | 识别帧时间不稳定 | P1 |
| 行为画像 | 组件行为模式分析 | P2 |
| 基准对比 | 与历史数据对比 | P2 |

### 1.2 设计原则

1. **零侵入**: 不修改应用代码
2. **低开销**: 分析开销 < 5% DevTools 总开销
3. **渐进增强**: Level 0-3 运行级别

---

## 二、架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    DevTools Core                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ EventBus │ │ Timeline │ │CausalGraph│ │Collector │       │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘       │
└───────┼──────────┼──────────┼──────────┼──────────────────┘
        │          │          │          │
        └──────────┴──────────┴──────────┴─────┐
                                                 ▼
┌────────────────────────────────────────────────────────────┐
│              Observation Layer (Phase 6 V1)                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐       │
│  │HotspotDetector│ │ WasteDetector│ │JitterDetector│      │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘       │
│         │                │                │               │
│  ┌──────┴───────┐ ┌──────┴───────┐ ┌──────┴───────┐       │
│  │BehaviorProfiler│ │BaselineComparator│ │ConfidenceModel││
│  └──────┬───────┘ └──────┴───────┘ └──────┬───────┘       │
└─────────┼──────────────────────────────────┼───────────────┘
          │                                  │
          └──────────────┬───────────────────┘
                         ▼
┌────────────────────────────────────────────────────────────┐
│                   Insights & Suggestions                    │
│  • 性能热点报告     • 浪费渲染警告     • 优化建议           │
└────────────────────────────────────────────────────────────┘
```

---

## 三、核心组件

### 3.1 HotspotDetector (热点检测器)

```go
type HotspotDetector struct {
    mu              sync.RWMutex
    enabled         atomic.Bool
    frameBuffer     *RingBuffer[FrameMetrics]  // 保留最近 N 帧
    componentStats  map[NodeID]*ComponentHotspot

    // 配置
    slowFrameThreshold  time.Duration  // 慢帧阈值 (默认 16.67ms)
    hotComponentThreshold time.Duration // 热组件阈值
    sampleRate        float64         // 采样率 (0.0-1.0)
}

type FrameMetrics struct {
    FrameID      devtools.FrameID
    Duration     time.Duration
    LayoutTime   time.Duration
    PaintTime    time.Duration
    SlowNodes    []NodeID
}

type ComponentHotspot struct {
    NodeID       NodeID
    FrameCount   uint64
    TotalTime    time.Duration
    AvgTime      time.Duration
    MaxTime      time.Duration
    LastSlowTime time.Time
    Severity     HotspotSeverity
}

type HotspotSeverity int
const (
    SeverityNone HotspotSeverity = iota
    SeverityWarning  // > 16.67ms
    SeverityCritical // > 33.33ms
)
```

**检测逻辑**:
1. 每帧结束后，收集帧度量
2. 识别慢帧（超过阈值）
3. 追踪慢组件
4. 使用 EWMA 更新平均时间

### 3.2 WasteDetector (浪费检测器)

```go
type WasteDetector struct {
    mu              sync.RWMutex
    enabled         atomic.Bool
    frameBuffer     *RingBuffer[WasteFrame]

    // 组件状态追踪
    lastComponentState map[NodeID]ComponentState

    // 配置
    minWasteThreshold int  // 最小浪费帧数（连续）
}

type WasteFrame struct {
    FrameID       devtools.FrameID
    WastedLayouts int
    WastedPaints  int
    WastedNodes   []NodeID
}

type ComponentState struct {
    LayoutVersion uint32
    ContentHash   uint64
    LastChanged   time.Time
}

type WasteReport struct {
    NodeID        NodeID
    WasteCount    int
    WasteRate     float64  // 浪费率 (浪费次数/总次数)
    LastWasteTime time.Time
    Severity      WasteSeverity
}

type WasteSeverity int
const (
    WasteNone WasteSeverity = iota
    WasteLow    // 10-30% 浪费
    WasteMedium // 30-50% 浪费
    WasteHigh   // > 50% 浪费
)
```

**检测逻辑**:
1. 追踪每个组件的状态（布局版本、内容哈希）
2. 检测无变化的布局/重绘
3. 计算浪费率
4. 报告高浪费组件

### 3.3 JitterDetector (抖动检测器)

```go
type JitterDetector struct {
    mu              sync.RWMutex
    enabled         atomic.Bool
    frameBuffer     *RingBuffer[time.Duration]  // 最近帧时间

    // 统计
    meanDuration    float64
    variance        float64
    stdDev          float64

    // 配置
    jitterThreshold float64  // 抖动阈值 (标准差倍数)
    windowSize      int      // 统计窗口大小
}

type JitterReport struct {
    CurrentJitter    float64  // 当前抖动值
    MeanFrameTime    float64  // 平均帧时间
    StdDev           float64  // 标准差
    JitterFrames     []devtools.FrameID  // 抖动帧
    Severity         JitterSeverity
}

type JitterSeverity int
const (
    JitterNone JitterSeverity = iota
    JitterLow     // 轻微抖动 (CV < 0.2)
    JitterMedium  // 中度抖动 (CV 0.2-0.5)
    JitterHigh    // 高度抖动 (CV > 0.5)
)
```

**检测逻辑**:
1. 收集最近 N 帧的持续时间
2. 计算均值和方差
3. 使用变异系数 (CV) 评估抖动
4. 识别异常帧

### 3.4 BehaviorProfiler (行为画像)

```go
type BehaviorProfiler struct {
    mu              sync.RWMutex
    enabled         atomic.Bool

    // 组件行为画像
    profiles        map[NodeID]*BehaviorProfile
    eventPatterns   map[EventType]*EventPattern

    // 学习状态
    learningMode    bool
    minSamples      int
}

type BehaviorProfile struct {
    NodeID           NodeID
    SampleCount      uint64

    // 更新频率
    UpdateFrequency  Frequency  // High/Medium/Low

    // 触发模式
    TriggerEvents    []EventType
    TriggeredBy      map[EventType]int

    // 性能特征
    AvgLayoutTime    time.Duration
    AvgPaintTime     time.Duration
    Percentile90     time.Duration
    Percentile99     time.Duration

    // 异常检测
    Baseline         *PerformanceBaseline
    AnomalyCount     int
    LastAnomalyTime  time.Time
}

type Frequency int
const (
    FrequencyHigh Frequency = iota    // > 60 次/秒
    FrequencyMedium                   // 10-60 次/秒
    FrequencyLow                      // < 10 次/秒
)

type PerformanceBaseline struct {
    Mean       float64
    StdDev     float64
    Min        time.Duration
    Max        time.Duration
    UpdatedAt  time.Time
}

type EventPattern struct {
    EventType   EventType
    Frequency   float64
    AvgInterval time.Duration
    LastSeen    time.Time
}
```

### 3.5 BaselineComparator (基准对比)

```go
type BaselineComparator struct {
    mu              sync.RWMutex
    enabled         atomic.Bool

    // 历史基准
    baselines       map[string]*PerformanceBaseline

    // 快照存储
    snapshots       map[time.Time]*FrameSnapshot
}

type FrameSnapshot struct {
    Timestamp       time.Time
    FrameCount      int
    AvgDuration     time.Duration
    P95Duration     time.Duration
    P99Duration     time.Duration
    TotalLayouts    int
    TotalPaints     int
    Hotspots        []ComponentHotspot
}

type ComparisonResult struct {
    Metric          string
    Current         float64
    Baseline        float64
    ChangePercent   float64
    Direction       TrendDirection
    Significance    bool  // 是否显著变化
}

type TrendDirection int
const (
    TrendNone TrendDirection = iota
    TrendImproved
    TrendDegraded
    TrendUnchanged
)
```

---

## 四、置信度模型

```go
type ConfidenceLevel int

const (
    ConfidenceLow ConfidenceLevel = iota    // < 0.5
    ConfidenceMedium                        // 0.5-0.7
    ConfidenceHigh                          // 0.7-0.9
    ConfidenceVeryHigh                      // > 0.9
)

type Insight struct {
    Type         InsightType
    Title        string
    Description  string
    Confidence   ConfidenceLevel
    Evidence     []Evidence
    Suggestions  []Suggestion
    Severity     Severity
}

type InsightType int
const (
    InsightHotspot InsightType = iota
    InsightWaste
    InsightJitter
    InsightAnomaly
    InsightRegression
)

type Evidence struct {
    Source    string  // "HotspotDetector", "WasteDetector", etc.
    Metric    string
    Value     float64
    Expected  float64
}

type Suggestion struct {
    Priority  int
    Action    string
    Reason    string
    ExpectedImpact string
}

type Severity int
const (
    SeverityInfo Severity = iota
    SeverityWarning
    SeverityCritical
)
```

### 置信度计算

```go
func CalculateInsightConfidence(insight *Insight) ConfidenceLevel {
    var score float64

    // 1. 样本量评分 (0-0.3)
    sampleScore := calculateSampleScore(insight.Evidence)
    score += sampleScore * 0.3

    // 2. 一致性评分 (0-0.3)
    consistencyScore := calculateConsistencyScore(insight.Evidence)
    score += consistencyScore * 0.3

    // 3. 因果强度评分 (0-0.4)
    causalScore := calculateCausalScore(insight.Evidence)
    score += causalScore * 0.4

    return mapScoreToLevel(score)
}
```

---

## 五、运行级别

### Level 0: 禁用
- 所有检测器关闭
- 零开销

### Level 1: 基础监控
- 仅启用 HotspotDetector
- 开销: ~50 ns/op

### Level 2: 增强监控
- HotspotDetector + WasteDetector
- 开销: ~200 ns/op

### Level 3: 完整分析
- 所有检测器 + BehaviorProfiler + BaselineComparator
- 开销: ~500 ns/op

---

## 六、API 设计

```go
// 启用观测层
func (dt *DevTools) EnableObservation(level Level) error

// 获取热点报告
func (dt *DevTools) GetHotspots() []ComponentHotspot

// 获取浪费报告
func (dt *DevTools) GetWasteReport() []WasteReport

// 获取抖动报告
func (dt *DevTools) GetJitterReport() *JitterReport

// 获取组件行为画像
func (dt *DevTools) GetBehaviorProfile(nodeID NodeID) *BehaviorProfile

// 创建基准快照
func (dt *DevTools) CreateSnapshot(name string) error

// 对比当前状态与基准
func (dt *DevTools) CompareWithBaseline(name string) *ComparisonResult

// 获取所有洞察
func (dt *DevTools) GetInsights() []Insight
```

---

## 七、可视化集成

```
┌─────────────────────────────────────────────────────────────┐
│                    TUI Debug Panel                           │
├─────────────────────────────────────────────────────────────┤
│  [Timeline] [Causal] [Observation] [Insights]                │
├─────────────────────────────────────────────────────────────┤
│  Observation View:                                           │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ 🔴 Hotspots (3)                                      │    │
│  │   • Button#123 - 45ms/frame (Critical)              │    │
│  │   • List#456 - 28ms/frame (Warning)                 │    │
│  │   • Container#789 - 22ms/frame (Warning)            │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ ⚠️  Waste (2)                                        │    │
│  │   • Header#111 - 67% wasted renders                 │    │
│  │   • Status#222 - 54% wasted renders                 │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │ 📊 Frame Jitter: Medium (CV: 0.35)                  │    │
│  │   Avg: 14.2ms, StdDev: 5.0ms                        │    │
│  └─────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│  Insights:                                                   │
│  [1] High hotspot risk in Button#123 (Confidence: 0.92)     │
│      → Consider using shouldComponentUpdate                  │
│  [2] Wasted renders in Header#111 (Confidence: 0.85)        │
│      → Content unchanged for 67% of renders                  │
└─────────────────────────────────────────────────────────────┘
```

---

## 八、实施计划

### Phase 6.1: HotspotDetector (P0)
- [ ] 实现基础热点检测
- [ ] 帧度量和组件追踪
- [ ] EWMA 平滑算法
- [ ] TUI 集成

### Phase 6.2: WasteDetector (P1)
- [ ] 组件状态追踪
- [ ] 浪费检测算法
- [ ] 浪费率计算

### Phase 6.3: JitterDetector (P1)
- [ ] 帧时间统计
- [ ] 变异系数计算
- [ ] 异常帧识别

### Phase 6.4: BehaviorProfiler (P2)
- [ ] 行为模式学习
- [ ] 异常检测
- [ ] 画像生成

### Phase 6.5: BaselineComparator (P2)
- [ ] 快照创建/加载
- [ ] 基准对比算法
- [ ] 趋势分析

---

## 九、性能预算

| 组件 | 预算 | 测量方法 |
|------|------|----------|
| HotspotDetector | < 100 ns/op | Benchmark |
| WasteDetector | < 150 ns/op | Benchmark |
| JitterDetector | < 50 ns/op | Benchmark |
| BehaviorProfiler | < 200 ns/op | Benchmark |
| **总计 (Level 3)** | < 500 ns/op | End-to-end |

---

## 十、测试策略

### 单元测试
- 每个检测器的独立测试
- 边界条件测试
- 并发安全测试

### 集成测试
- 与现有 DevTools 集成
- 数据流完整性
- TUI 集成

### 性能测试
- 基准测试
- 内存泄漏检测
- CPU 使用率

### 真实场景测试
- 模拟应用场景
- 长时间运行稳定性
