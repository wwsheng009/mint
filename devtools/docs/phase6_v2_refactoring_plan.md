# Phase 6 V2 重构方案

> **目标**: 按照设计文档将 V1/V2 能力分离
> **状态**: 进行中
> **日期**: 2026-01-30

---

## 一、当前状态分析

### 1.1 当前实现混合了 V1 和 V2 能力

```
当前 observation/ 包结构:
├── hotspot.go      ❌ 混合了统计+判断
├── waste.go        ❌ 混合了统计+Severity分级
├── jitter.go       ❌ 混合了统计+Severity分级
├── profiler.go     ❌ 混合了统计+异常检测
├── baseline.go     ✅ 偏向V2 (对比需要基线)
└── insights.go     ❌ 纯V2能力 (出现在V1)
```

### 1.2 问题

1. **V1 不纯**: 有 Severity、Insight、Suggestion 等判断能力
2. **缺少 PatternDetector**: V2 核心能力未实现
3. **置信度模型简化**: 只有基础4级，缺少5类信号评分

---

## 二、重构目标

### 2.1 V1 - 纯统计层 (无判断)

```
devtools/observation/
├── v1/
│   ├── metrics.go       # MetricsCollector - 纯计数
│   ├── stats.go         # StatsAnalyzer - TopN, 百分位
│   ├── timeseries.go    # TimeSeriesStore - 固定窗口
│   └── level.go         # LevelController - 级别控制
```

**原则**:
- ✅ 统计数据: 计数、百分位、TopN
- ✅ 时序存储: 固定窗口
- ❌ 无 Severity 分级
- ❌ 无 Insight 生成
- ❌ 无 Suggestion

### 2.2 V2 - 模式识别层 (有判断)

```
devtools/observation/
├── v2/
│   ├── pattern.go       # PatternDetector - 模式识别
│   ├── confidence.go    # ConfidenceModel - 5类信号
│   ├── insights.go      # Insights - 置信度+建议
│   ├── hotspot.go       # HotspotAnalyzer - 统计异常标记
│   ├── waste.go         # WasteAnalyzer - 冗余标记
│   └── jitter.go        # JitterAnalyzer - 抖动标记
```

**新增 PatternDetector**:
- OscillationPattern (A→B→A→B)
- SameFieldPattern (同字段连续改)
- CascadeBurstPattern (级联爆发)
- LayoutRevertPattern (Layout后立即反向改)

---

## 三、实施步骤

### Step 1: 创建 V1 纯统计层

| 文件 | 内容 | 移除内容 |
|------|------|----------|
| v1/metrics.go | MetricsCollector, ComponentMetrics | - |
| v1/stats.go | TopNAnalyzer, PercentileCalc | - |
| v1/timeseries.go | TimeSeriesStore, RingBuffer | - |
| v1/level.go | LevelController, AnalysisLevel | - |

### Step 2: 重构现有检测器到 V2

| 原 | 新 | 变化 |
|----|------|------|
| hotspot.go | v2/hotspot.go | 移除 Severity，改用 Confidence |
| waste.go | v2/waste.go | 移除 Severity，改用 Confidence |
| jitter.go | v2/jitter.go | 移除 Severity，改用 Confidence |
| profiler.go | v2/profile.go | 保留 BehaviorProfile |
| baseline.go | v2/baseline.go | 保留 BaselineComparator |
| insights.go | v2/insights.go | 增强 ConfidenceModel |

### Step 3: 实现 PatternDetector

```go
// PatternDetector 模式检测器
type PatternDetector struct {
    mu     sync.RWMutex
    enabled atomic.Bool

    // 模式识别器
    oscillationDetector  *OscillationDetector
    sameFieldDetector    *SameFieldDetector
    cascadeDetector      *CascadeDetector
    layoutRevertDetector *LayoutRevertDetector

    // 模式缓存
    patternCache map[NodeID][]*DetectedPattern
}

type PatternType int
const (
    PatternOscillation PatternType = iota
    PatternSameField
    PatternCascadeBurst
    PatternLayoutRevert
)

type DetectedPattern struct {
    Type      PatternType
    NodeID    NodeID
    Confidence float64
    StartTime time.Time
    EndTime   time.Time
    Evidence  []PatternEvidence
}
```

### Step 4: 增强 ConfidenceModel

```go
// ConfidenceModel 5类信号评分
type ConfidenceModel struct {
    weights ConfidenceWeights
}

type ConfidenceWeights struct {
    Statistical float64  // 0.25
    Pattern      float64  // 0.20
    Causal       float64  // 0.30  // 最高
    Context      float64  // -0.10 // 惩罚
    Historical   float64  // 0.15
}

type SignalScores struct {
    StatScore    float64  // 统计置信
    PatternScore  float64  // 模式置信
    CausalScore   float64  // 因果置信
    ContextScore  float64  // 上下文置信
    HistoricalScore float64 // 历史置信
}
```

---

## 四、文件结构 (重构后)

```
devtools/observation/
├── README.md
│
├── v1/                          # V1: 纯统计层
│   ├── level.go                 # 运行级别定义
│   ├── metrics.go               # 指标收集
│   ├── stats.go                 # 统计分析 (TopN, 百分位)
│   └── timeseries.go            # 时序存储
│
├── v2/                          # V2: 模式识别层
│   ├── pattern.go               # 模式检测器
│   ├── pattern_types.go         # 模式类型定义
│   ├── confidence.go            # 置信度模型
│   ├── insights.go              # Insights 生成
│   ├── hotspot.go               # 热点分析器
│   ├── waste.go                 # 浪费分析器
│   ├── jitter.go                # 抖动分析器
│   ├── profile.go               # 行为画像
│   └── baseline.go              # 基线对比
│
├── layer.go                     # 主入口 (协调V1+V2)
└── observation_test.go          # 测试
```

---

## 五、API 变化

### 5.1 V1 API (纯统计)

```go
// V1 只提供统计数据，不返回 Insight
type V1Layer interface {
    SetLevel(level Level)

    // 纯统计数据
    GetMetrics() *MetricsSnapshot
    GetTopN(metric MetricType, n int) []*ComponentRank
    GetPercentiles(nodeID NodeID) []PercentileValue
    GetTimeSeries(nodeID NodeID) []DataPoint
}
```

### 5.2 V2 API (模式+洞察)

```go
// V2 提供判断能力
type V2Layer interface {
    EnablePatternDetection()

    // 模式检测
    GetPatterns(nodeID NodeID) []*DetectedPattern
    DetectPatterns() []*DetectedPattern

    // 置信度
    CalculateConfidence(signal *Signal) float64

    // Insights (带置信度)
    GetInsights() []Insight
    GetHighConfidenceInsights(threshold float64) []Insight
}
```

---

## 六、迁移检查清单

### V1 纯统计化
- [ ] hotspot.go 移除 SeverityHotspotSeverity*
- [ ] waste.go 移除 WasteSeverity*
- [ ] jitter.go 移除 JitterSeverity*
- [ ] profiler.go 保留纯统计部分
- [ ] 创建 v1/ 目录结构
- [ ] 迁移纯统计代码到 v1/

### V2 模式识别
- [ ] 实现 PatternDetector
- [ ] 实现 OscillationDetector
- [ ] 实现 SameFieldDetector
- [ ] 实现 CascadeDetector
- [ ] 实现 LayoutRevertDetector

### V2 置信度模型
- [ ] 实现 5类信号评分
- [ ] 实现权重配置
- [ ] 实现置信度计算
- [ ] 集成到 Insights

### 测试
- [ ] V1 统计测试
- [ ] V2 模式检测测试
- [ ] 置信度计算测试
- [ ] 端到端测试
