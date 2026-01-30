# Observation - 观察层模块

> 数据收集、统计分析、模式检测、洞察生成

## 功能概述

Observation 模块提供分层的数据收集和分析功能，从基础的计数统计到高级的模式检测和 AI 洞察生成。

## 模块结构

```
observation/
├── layer.go           # 核心观察层
├── v1/                # V1 统计层
│   ├── level.go       # 观察级别
│   ├── metrics.go     # 指标收集
│   ├── stats.go       # 统计数据
│   └── timeseries.go  # 时间序列
└── v2/                # V2 智能层
    ├── confidence.go  # 置信度模型
    ├── insights.go    # 洞察生成
    ├── pattern_detector.go  # 模式检测
    └── pattern_types.go      # 模式类型
```

## 核心组件

### 1. Observation Layer (`layer.go`)

```go
// Layer 观察层
type Layer struct {
    mu            sync.RWMutex
    config        *Config
    enabled       bool
    level         Level
    v1Metrics     *v1.Metrics
    v2Detector    *v2.PatternDetector
    v2Insights    *v2.InsightGenerator
}

// 创建观察层
func NewLayer(cfg *Config) *Layer

// 启用观察层
func (l *Layer) Enable(level Level)

// 记录变更
func (l *Layer) RecordMutation(nodeID devtools.NodeID, field string, value interface{})

// 获取指标
func (l *Layer) GetMetrics() *v1.MetricsSnapshot

// 获取模式
func (l *Layer) GetPatterns(nodeID devtools.NodeID) []v2.DetectedPattern

// 获取洞察
func (l *Layer) GetInsights(nodeID devtools.NodeID) []v2.Insight
```

### 2. V1 统计层

```go
// 观察级别
const (
    LevelNone      Level = 0  // 完全禁用
    LevelBasic     Level = 1  // 基础计数
    LevelEnhanced  Level = 2  // 增强统计
    LevelAdvanced  Level = 3  // 高级分析
)

// 指标类型
const (
    MetricFrames      MetricType = "frames"
    MetricMutations   MetricType = "mutations"
    MetricLayouts     MetricType = "layouts"
    MetricRepaints    MetricType = "repaints"
)

// 获取 Top N
func (m *Metrics) GetTopN(metric MetricType, n int) []NodeRank

// 获取分布
func (m *Metrics) GetDistribution(metric MetricType) *Distribution
```

### 3. V2 智能层

```go
// PatternDetector 模式检测器
type PatternDetector struct {
    mu            sync.RWMutex
    enabled       bool
    detectors     map[PatternType]DetectorFunc
    confidence    *ConfidenceModel
}

// 检测模式
func (pd *PatternDetector) Detect(nodeID NodeID) []DetectedPattern

// ConfidenceModel 置信度模型
type ConfidenceModel struct {
    weights map[string]float64
}

// 计算置信度
func (cm *ConfidenceModel) Calculate(evidence *Evidence) float64

// InsightGenerator 洞察生成器
type InsightGenerator struct {
    detector  *PatternDetector
    model     *ConfidenceModel
}

// 生成洞察
func (ig *InsightGenerator) Generate(nodeID NodeID) []Insight
```

## 模式类型

| 模式 | 描述 | 严重性 |
|------|------|--------|
| Oscillation | A→B→A→B 值振荡 | High |
| SameField | 快速修改同一字段 | Medium |
| CascadeBurst | 级联爆发更新 | High |
| LayoutRevert | 布局立即回滚 | Medium |
| HighFrequency | 高频更新 (>60/sec) | High |
| Burst | 突发更新模式 | Low |

## 使用方法

### 基础使用

```go
import "github.com/wwsheng009/mint/devtools/observation"
import v1 "github.com/wwsheng009/mint/devtools/observation/v1"

// 创建观察层
cfg := observation.DefaultConfig()
layer := observation.NewLayer(cfg)
layer.Enable(v1.LevelAdvanced)

// 记录状态变更
layer.RecordMutation("button-1", "clicked", true)
layer.RecordMutation("input-1", "value", "hello")

// 获取统计指标
metrics := layer.GetMetrics()
fmt.Printf("Mutations: %d\n", metrics.TotalMutations)

// 获取 Top N 组件
topMutators := metrics.GetTopN(v1.MetricMutations, 10)
for _, item := range topMutators {
    fmt.Printf("%s: %d mutations\n", item.NodeID, item.Count)
}
```

### 模式检测

```go
// 获取检测到的模式
patterns := layer.GetPatterns("button-1")

for _, pattern := range patterns {
    fmt.Printf("Pattern: %s (Confidence: %.2f)\n",
        pattern.Type, pattern.Confidence)

    for _, evidence := range pattern.Evidence {
        fmt.Printf("  - %s\n", evidence.Description)
    }
}
```

### 洞察生成

```go
// 获取优化建议
insights := layer.GetInsights("button-1")

for _, insight := range insights {
    fmt.Printf("[%s] %s\n", insight.Severity, insight.Type)
    fmt.Printf("Confidence: %.1f%%\n", insight.Confidence*100)

    for _, suggestion := range insight.Suggestions {
        fmt.Printf("  → %s\n", suggestion)
    }
}
```

### 置信度评分

```go
import "github.com/wwsheng009/mint/devtools/observation/v2"

// 5 信号置信度模型
model := v2.NewConfidenceModel()

// 设置权重
model.SetWeight("statistical", 0.25)
model.SetWeight("pattern", 0.20)
model.SetWeight("causal", 0.30)
model.SetWeight("context", -0.10)
model.SetWeight("historical", 0.15)

// 计算置信度
evidence := &v2.Evidence{
    Statistical: 0.8,
    Pattern:     0.9,
    Causal:      0.7,
    Context:     0.5,
    Historical:  0.6,
}

confidence := model.Calculate(evidence)
fmt.Printf("Confidence: %.2f\n", confidence)  // ~0.73
```

## 观察级别开销

| 级别 | 开销 | 功能 |
|------|------|------|
| LevelNone | 0% | 完全禁用 |
| LevelBasic | <1% | 计数统计 |
| LevelEnhanced | <3% | TopN、百分位 |
| LevelAdvanced | <5% | 模式检测、洞察 |

## 相关模块

| 模块 | 关系 |
|------|------|
| `devtools` | 核心功能，观察层记录其事件 |
| `causal` | 因果链数据，用于置信度计算 |
| `client` | 调试面板，展示统计数据和模式 |

## API 参考

### MetricsSnapshot

```go
type MetricsSnapshot struct {
    TotalFrames    uint64
    TotalMutations uint64
    TotalLayouts   uint64
    TotalRepaints  uint64
    ByNode         map[NodeID]*NodeMetrics
}
```

### Distribution

```go
type Distribution struct {
    Count  int
    Min    uint64
    Max    uint64
    Mean   float64
    Median uint64
    P90    uint64
    P95    uint64
    P99    uint64
    StdDev float64
}
```

### DetectedPattern

```go
type DetectedPattern struct {
    ID         string
    Type       PatternType
    NodeID     NodeID
    Confidence float64
    Severity   PatternSeverity
    Evidence   []PatternEvidence
}
```

## 文件列表

- `layer.go` - 核心观察层
- `v1/level.go` - 观察级别
- `v1/metrics.go` - 指标收集
- `v1/stats.go` - 统计数据
- `v1/timeseries.go` - 时间序列
- `v2/confidence.go` - 置信度模型
- `v2/insights.go` - 洞察生成
- `v2/pattern_detector.go` - 模式检测
- `v2/pattern_types.go` - 模式类型定义
