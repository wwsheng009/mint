# DevTools 阶段6: 智能分析层设计方案 V2.0

> **项目**: Mint TUI Runtime
> **文档版本**: 2.0 (根据评审修订)
> **创建日期**: 2026-01-30
> **修订日期**: 2026-01-30
> **状态**: 设计中
> **依赖**: 阶段1-5 已完成

> **V2 修订说明**: 根据阶段6评审文档，从"一次性上线所有功能"改为"按智能成熟度模型分阶段演进"，加入运行级别、置信度模型、行为画像等核心机制。

---

## 目录

1. [概述](#一概述)
2. [V2 核心改进](#二v2-核心改进)
3. [分阶段演进路线](#三分阶段演进路线)
4. [运行级别系统](#四运行级别系统)
5. [置信度模型](#五置信度模型)
6. [行为画像系统](#六行为画像系统)
7. [模块设计 (V1)](#七模块设计-v1)
8. [文件结构](#八文件结构)
9. [API 设计](#九api-设计)

---

## 一、概述

### 1.1 设计理念转变

V1 设计的问题：试图一次性上线完整的智能分析系统

V2 设计的核心：**按智能成熟度模型分阶段演进**

```
┌─────────────────────────────────────────────────────────────────┐
│                    设计理念转变                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  V1 设计 (❌ 风险过高):                                           │
│    一次性上线: Analyzer → Pattern → Rule → Suggestion → AutoFix  │
│    问题: 信号不稳定、建议误报多、自动修复不敢开                    │
│                                                                  │
│  V2 设计 (✅ 可演进):                                             │
│    分阶段上线: V1观测 → V2识别 → V3建议 → V4修复 → V5自适应       │
│    优势: 每阶段验证后再进入下一阶段，风险可控                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 系统定位

这不是 DevTools，而是：

> **UI Runtime 行为智能分析引擎**

```
数据采集 (阶段1-5) → 行为分析 → 模式识别 → 置信度评估 → 智能决策
```

### 1.3 能力层级

```
┌─────────────────────────────────────────────────────────────────┐
│                    智能能力层级图                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  L1: 数据采集         ✅ 阶段1-5 已完成                           │
│  L2: 统计分析         🎯 V1 目标                                 │
│  L3: 模式识别         🎯 V2 目标                                 │
│  L4: 置信度评估       🎯 V3 目标                                 │
│  L5: 行为画像         🎯 V3 目标                                 │
│  L6: 模式学习         🎯 V4 目标                                 │
│  L7: 自适应规则       🎯 V5 目标                                 │
│  L8: 运行时调节       🔮 未来                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 二、V2 核心改进

### 2.1 关键问题修复

| 问题 | V1 设计 | V2 修正 |
|------|---------|---------|
| 运行时开销 | 承诺"零开销"（不实际） | 分级运行，Level 1 才开启分析 |
| O(n²) 复杂度 | 历史窗口无限增长 | 固定 Ring Buffer (30帧) |
| CodeRewriter | Runtime 直接改文件 | 移到 IDE 插件层 |
| React 术语 | useMemo, setState | 抽象为通用 TUI 术语 |
| 一次性上线 | 所有功能同时开发 | V1-V5 分阶段演进 |

### 2.2 新增核心机制

```
┌─────────────────────────────────────────────────────────────────┐
│                    V2 新增核心机制                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. 运行级别系统 (Level 0-3)                                     │
│     • Level 0: 关闭 - 零开销                                    │
│     • Level 1: 轻量统计 - 只计数                                │
│     • Level 2: 深度分析 - 含历史窗口                             │
│     • Level 3: 全开 - DevTools 完整功能                         │
│                                                                  │
│  2. 置信度模型 (Confidence Model)                                │
│     • 统计置信 (Statistical) - 值是否异常                        │
│     • 模式置信 (Pattern) - 形态是否像 bug                        │
│     • 因果置信 (Causal) - 是否导致性能损失                      │
│     • 上下文置信 (Context) - 场景是否合理                        │
│     • 历史置信 (Historical) - 是否持续异常                       │
│                                                                  │
│  3. 行为画像系统 (Behavior Profile)                              │
│     • 每个组件的长期行为特征向量                                 │
│     • EWMA 滑动学习                                             │
│     • 从"绝对异常"变为"相对异常"                                 │
│                                                                  │
│  4. Insight 失效机制 (TTL)                                       │
│     • 自动清理过期洞察                                           │
│     • 防止内存泄漏                                              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 术语修正

| V1 (React 风格) | V2 (TUI 通用) |
|-----------------|---------------|
| useMemo | Computation Caching |
| setState batching | State Coalescing |
| state lifting | State Ownership Reassignment |
| shouldComponentUpdate | Render Guard |
| debounce/throttle | Event Rate Limiting |

---

## 三、分阶段演进路线

### 3.1 演进总览

```
┌─────────────────────────────────────────────────────────────────┐
│                    5 阶段演进路线图                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  V1 ──── V2 ──── V3 ──── V4 ──── V5                             │
│  观测    识别    建议    修复    自适应                           │
│  ↓       ↓       ↓       ↓       ↓                              │
│  2周     2周     4周     4周     未来                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 V1 - 观测增强层 (安全上线版本)

**目标**: 绝不做判断，只做"统计型洞察"

**上线时间**: 2 周

**核心能力**:

| 模块 | 保留内容 | 禁止内容 |
|------|----------|----------|
| MetricsCollector | 频率统计、Top N 组件 | 不做"问题判定" |
| HotspotAnalyzer | 计数、百分位 | 不判断是好是坏 |
| WasteDetector | 计数 | 不判断 NoOp 是否错误 |
| JitterDetector | 变化次数统计 | 不说"抖动" |
| Insight | 纯客观数据 | 没有 Suggestion |

**输出示例**:

```text
Component: button_submit
  Mutation Rate: 145/sec
  Layout Rate: 70/sec
  Repaint Rate: 68/sec
  Affected Nodes (avg): 23
  Percentile: 98th
```

❗ **不出现"优化建议"、"问题"、"警告"等字眼**

**验收标准**:
- [ ] 系统稳定运行 2 周无崩溃
- [ ] 收集到足够数据分布
- [ ] 确定什么是"正常值"

### 3.3 V2 - 模式识别层 (开始有"判断")

**目标**: 标记"异常值"而不是"问题"

**上线条件**: V1 稳定运行 2 周

**核心能力**:

| 模块 | 新增能力 |
|------|----------|
| HotspotAnalyzer | 标记"统计异常值" |
| WasteDetector | 标记"可能冗余" |
| JitterDetector | 标记"高频变化模式" |
| PatternDetector | 只识别模式，不给建议 |
| ConfidenceModel | 基础置信度计算 |

**输出示例**:

```text
⚠ Observed Anomaly:
Component: button_submit
  Mutation rate is in top 2% of all components
  Statistical confidence: 0.76
  Pattern: burst_updates
```

❌ **仍然不给建议**

**验收标准**:
- [ ] 误报率 < 20%
- [ ] 用户不反感提示
- [ ] 数据分布趋于稳定

### 3.4 V3 - 建议生成层 (但不自动修复)

**目标**: 高置信度模式给出优化建议

**上线条件**: V2 误报率可接受

**核心能力**:

| 模块 | 条件 |
|------|------|
| OptimizationEngine | 仅针对 Confidence > 0.8 |
| Suggestion | 必须带置信度 |
| RuleEngine | 只跑 P0 规则 |
| BehaviorProfile | 参与判断 |

**输出示例**:

```text
Detected: High mutation burst pattern
Confidence: 0.86
Suggestion: Consider coalescing state updates
  Instead of:
    setState({field: value1})
    setState({field: value2})
  Use:
    setState({field: value1, field2: value2})
```

⚠️ **仍然不能 AutoFix**

**验收标准**:
- [ ] 建议采纳率 > 30%
- [ ] 用户不觉得烦
- [ ] 高置信度建议准确

### 3.5 V4 - 半自动修复层 (需要人工确认)

**目标**: IDE 插件辅助应用修复

**上线条件**: V3 建议质量稳定

**核心能力**:

| 模块 | 形态 |
|------|------|
| CodeRewriter | 移出 Runtime → IDE 插件 |
| PatchGenerator | 生成 AST 级 Patch |
| IDEPlugin | 用户确认后应用 |

**流程**:

```
Runtime → Suggestion → IDE Plugin → AST Patch → User Confirm
```

**验收标准**:
- [ ] 自动修复准确率 > 90%
- [ ] 用户愿意使用
- [ ] 无破坏性修改

### 3.6 V5 - 自适应优化系统 (终极形态)

**目标**: 系统根据历史学习阈值

**上线条件**: V4 稳定运行

**核心能力**:

| 能力 | 说明 |
|------|------|
| 自适应阈值 | 系统自己学正常分布 |
| 误报自学习 | 用户忽略的建议自动降权 |
| 组件画像 | 每个组件有行为模型 |
| 自动规则生成 | 系统发现新模式 |

---

## 四、运行级别系统

### 4.1 级别定义

```go
// AnalysisLevel 分析级别
type AnalysisLevel int

const (
    LevelNone AnalysisLevel = iota // 0: 完全关闭
    LevelLight                      // 1: 轻量统计
    LevelDeep                       // 2: 深度分析
    LevelFull                       // 3: 完全开启
)

// LevelConfig 级别配置
type LevelConfig struct {
    Level AnalysisLevel

    // Level 0: 零开销
    // Level 1: 只计数，无锁，无历史
    // Level 2: 有历史窗口，有锁
    // Level 3: 全部功能
}
```

### 4.2 各级别能力

```
┌─────────────────────────────────────────────────────────────────┐
│                    运行级别能力矩阵                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  能力                  │ L0 │ L1 │ L2 │ L3                     │
│  ────────────────────────┼────┼────┼────┼────                    │
│  帧计数                  │ ❌ │ ✅ │ ✅ │ ✅                      │
│  组件计数                │ ❌ │ ✅ │ ✅ │ ✅                      │
│  频率统计                │ ❌ │ ✅ │ ✅ │ ✅                      │
│  Top N 排序              │ ❌ │ ❌ │ ✅ │ ✅                      │
│  历史窗口 (30帧)         │ ❌ │ ❌ │ ✅ │ ✅                      │
│  模式识别                │ ❌ │ ❌ │ ❌ │ ✅                      │
│  置信度计算              │ ❌ │ ❌ │ ❌ │ ✅                      │
│  优化建议                │ ❌ │ ❌ │ ❌ │ ✅                      │
│  行为画像                │ ❌ │ ❌ │ ❌ │ ✅                      │
│                                                                  │
│  预期开销                │ 0% │<1% │<3% │<5%                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 4.3 实现

```go
// AnalysisEngine 分析引擎
type AnalysisEngine struct {
    mu sync.RWMutex

    level      AnalysisLevel
    lightStats *LightStats     // Level 1: 原子计数
    deepAnalyzer *DeepAnalyzer  // Level 2-3: 完整分析
}

type LightStats struct {
    // 原子计数，无锁
    totalFrames     atomic.Int64
    totalMutations  atomic.Int64
    totalLayouts    atomic.Int64
    totalRepaints   atomic.Int64
}

func (ae *AnalysisEngine) SetLevel(level AnalysisLevel) {
    ae.mu.Lock()
    defer ae.mu.Unlock()

    ae.level = level

    switch level {
    case LevelNone:
        // 完全关闭，清理资源
        ae.lightStats = nil
        ae.deepAnalyzer = nil

    case LevelLight:
        // 只启动轻量统计
        ae.lightStats = &LightStats{}
        ae.deepAnalyzer = nil

    case LevelDeep, LevelFull:
        // 启动完整分析
        ae.lightStats = &LightStats{}
        ae.deepAnalyzer = NewDeepAnalyzer(level)
    }
}

func (ae *AnalysisEngine) AnalyzeFrame(frame *FrameRecord) {
    switch ae.level {
    case LevelNone:
        return // 零开销

    case LevelLight:
        // 只计数，无锁，极快
        ae.lightStats.totalFrames.Add(1)
        ae.lightStats.totalMutations.Add(int64(len(frame.Mutations)))
        ae.lightStats.totalLayouts.Add(int64(len(frame.Layouts)))
        ae.lightStats.totalRepaints.Add(int64(len(frame.Repaints)))

    case LevelDeep, LevelFull:
        // 完整分析
        ae.deepAnalyzer.Analyze(frame)
    }
}
```

---

## 五、置信度模型

### 5.1 模型概述

> **Suggestion 必须是概率判断，而不是逻辑判断**

```text
Confidence = w1*StatScore + w2*PatternScore + w3*CausalScore + w4*ContextScore + w5*HistoricalScore

范围: 0 ~ 1
```

### 5.2 五类信号

#### 5.2.1 统计置信 (Statistical Confidence)

**判断**: 这个值是否异常，而不是大

```go
// 计算百分位排名
StatScore = percentileRank(value, distribution)

// 示例:
// Mutation rate = 120/s
// 系统平均 = 15/s, P95 = 40/s, P99 = 80/s
// → StatScore = 0.98 (异常尾部)
```

#### 5.2.2 模式置信 (Pattern Confidence)

**判断**: 形态是否像 bug

```go
// 模式特征加分
PatternScore = baseScore + patternBonus

// 模式加分表:
var patternBonuses = []PatternBonus{
    {Pattern: "oscillation_ab",      Bonus: 0.30}, // A→B→A→B
    {Pattern: "same_field_5x",       Bonus: 0.20}, // 同字段连续改5次
    {Pattern: "layout_revert",       Bonus: 0.30}, // Layout后立即被反向改
    {Pattern: "cascade_burst",       Bonus: 0.15}, // 级联爆发
}
```

#### 5.2.3 因果置信 (Causal Confidence) 🔥 最重要

**判断**: 是否真的导致性能损失

这是 DevTools 的独特优势（有 CausalGraph）

```go
// 检查因果链
CausalScore = 0.0
if mutation.LayoutImpact > 0 {
    CausalScore += 0.3
}
if mutation.RepaintImpact > 0 {
    CausalScore += 0.3
}
if causedFrameTimeSpike() {
    CausalScore += 0.4
}

// 只有真正导致性能问题，置信度才高
```

#### 5.2.4 上下文置信 (Context Confidence)

**判断**: 当前场景是否合理

```go
// 场景降权
ContextScore = baseScore - contextPenalty

var contextPenalties = []ContextPenalty{
    {Scenario: "animation",    Penalty: 0.30}, // 动画中高频正常
    {Scenario: "user_input",   Penalty: 0.20}, // 用户交互中
    {Scenario: "loading_state", Penalty: 0.20}, // loading 状态
    {Scenario: "drag_drop",    Penalty: 0.25}, // 拖拽操作
}
```

#### 5.2.5 历史置信 (Historical Confidence)

**判断**: 是否持续异常

```go
// 长期稳定性
HistoricalScore = 0.0
if anomalyDuration > 5*time.Minute {
    HistoricalScore += 0.4  // 持续异常 = 结构问题
}
if isRegressionFromBaseline() {
    HistoricalScore += 0.3  // 相比历史退化
}
if firstTimeAnomaly() {
    HistoricalScore -= 0.2  // 首次出现，可能是瞬态
}
```

### 5.3 权重配置

```go
// ConfidenceWeights 置信度权重
type ConfidenceWeights struct {
    Statistical float64  // 默认 0.25
    Pattern      float64  // 默认 0.20
    Causal       float64  // 默认 0.30  ← 最高
    Context      float64  // 默认 -0.10 (惩罚)
    Historical   float64  // 默认 0.15
}

// 计算最终置信度
func (cm *ConfidenceModel) Calculate(signal *Signal) float64 {
    weights := cm.GetWeights()  // V5 可自适应调整

    confidence :=
        weights.Statistical * signal.StatScore +
        weights.Pattern * signal.PatternScore +
        weights.Causal * signal.CausalScore +
        weights.Historical * signal.HistoricalScore -
        weights.Context * signal.ContextPenalty

    return clamp(confidence, 0.0, 1.0)
}
```

### 5.4 置信度阈值

```go
// 根据置信度决定行为
type ConfidenceLevel int

const (
    ConfidenceNone ConfidenceLevel = iota  // < 0.4: 只显示数据
    ConfidenceLow                            // 0.4-0.6: 标记异常
    ConfidenceMedium                          // 0.6-0.8: 弱建议
    ConfidenceHigh                            // 0.8-0.9: 强建议
    ConfidenceVeryHigh                        // > 0.9: AutoFix 候选
)
```

---

## 六、行为画像系统

### 6.1 核心思想

> **真正的异常不是"高"，而是"偏离自己的历史行为"**

### 6.2 Behavior Profile 结构

```go
// BehaviorProfile 组件行为画像
type BehaviorProfile struct {
    NodeID NodeID

    // 频率特征 (EWMA 学习)
    AvgMutationRate   float64
    P95MutationRate   float64
    P99MutationRate   float64
    StdDevMutation    float64
    AvgLayoutRate     float64
    AvgRepaintRate    float64

    // 变化模式特征
    JitterProbability float64  // 抖动概率
    BurstProbability  float64  // 爆发概率
    CascadeProbability float64 // 级联概率

    // 影响特征
    AvgAffectedNodes  float64
    AvgFrameCostImpact float64

    // 时间稳定性
    VarianceScore     float64
    StabilityScore    float64

    // EWMA 参数
    Alpha float64  // 学习率，默认 0.1

    LastUpdated time.Time
    SampleCount int64
}
```

### 6.3 EWMA 学习

```go
// UpdateProfile 更新画像
func (bp *BehaviorProfile) Update(mutationRate float64) {
    // 指数滑动平均
    bp.AvgMutationRate = bp.Alpha * mutationRate +
        (1 - bp.Alpha) * bp.AvgMutationRate

    // 增量更新标准差
    delta := mutationRate - bp.AvgMutationRate
    bp.VarianceScore = bp.Alpha * delta * delta +
        (1 - bp.Alpha) * bp.VarianceScore
    bp.StdDevMutation = math.Sqrt(bp.VarianceScore)

    bp.SampleCount++
    bp.LastUpdated = time.Now()
}

// IsAbnormal 检测是否异常 (相对判断)
func (bp *BehaviorProfile) IsAbnormal(currentRate float64) bool {
    // 不是绝对阈值，而是偏离自身历史
    threshold := bp.P95MutationRate * 1.8
    return currentRate > threshold
}
```

### 6.4 画像参与置信度

```go
// 计算行为偏离度
func (bp *BehaviorProfile) DeviationScore(currentRate float64) float64 {
    if bp.StdDevMutation == 0 {
        return 0
    }

    // Z-score: 当前值偏离均值多少个标准差
    zScore := (currentRate - bp.AvgMutationRate) / bp.StdDevMutation

    // 转换为 0-1 分数
    return 1.0 / (1.0 + math.Exp(-zScore))
}
```

---

## 七、模块设计 (V1)

### 7.1 V1 范围

**V1 只实现观测能力，不做任何判断**

```
┌─────────────────────────────────────────────────────────────────┐
│                    V1 模块范围                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ✅ MetricsCollector   - 频率统计                                 │
│  ✅ TopNAnalyzer      - Top N 组件                               │
│  ✅ PercentileCalc    - 百分位计算                               │
│  ✅ TimeSeriesStore   - 时序存储 (固定窗口)                       │
│                                                                  │
│  ❌ PatternDetector   - 模式识别 (V2)                            │
│  ❌ ConfidenceModel   - 置信度 (V2)                              │
│  ❌ OptimizationEngine - 建议 (V3)                               │
│  ❌ CodeRewriter      - 自动修复 (V4/IDE)                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 7.2 MetricsCollector

```go
// MetricsCollector 指标收集器 (V1)
type MetricsCollector struct {
    mu sync.RWMutex

    level AnalysisLevel

    // Level 1: 原子计数
    frameCount     atomic.Int64
    mutationCount  atomic.Int64
    layoutCount    atomic.Int64
    repaintCount   atomic.Int64

    // Level 1+: 组件级别统计
    componentStats map[NodeID]*ComponentMetrics
}

// ComponentMetrics 组件指标 (纯统计)
type ComponentMetrics struct {
    NodeID        NodeID

    // 计数
    MutationCount int64
    LayoutCount   int64
    RepaintCount  int64

    // 时间
    FirstSeen     time.Time
    LastSeen      time.Time
}

// RecordMutation 记录突变 (极快)
func (mc *MetricsCollector) RecordMutation(nodeID NodeID) {
    if mc.level == LevelNone {
        return
    }

    mc.mutationCount.Add(1)

    if mc.level >= LevelLight {
        mc.mu.Lock()
        stat := mc.getOrCreateStat(nodeID)
        stat.MutationCount++
        mc.mu.Unlock()
    }
}

// GetSnapshot 获取快照 (纯客观数据)
func (mc *MetricsCollector) GetSnapshot() *MetricsSnapshot {
    return &MetricsSnapshot{
        TotalFrames:     mc.frameCount.Load(),
        TotalMutations:  mc.mutationCount.Load(),
        TotalLayouts:    mc.layoutCount.Load(),
        TotalRepaints:   mc.repaintCount.Load(),
        ComponentStats:  mc.getComponentStatsCopy(),
    }
}
```

### 7.3 TopNAnalyzer

```go
// TopNAnalyzer Top N 分析器 (V1)
type TopNAnalyzer struct {
    mu sync.RWMutex

    metrics *MetricsCollector
    topN    int  // 默认 10
}

// TopComponents 获取 Top N 组件
func (ta *TopNAnalyzer) TopComponents(metric MetricType) []*ComponentRank {
    ta.mu.RLock()
    defer ta.mu.RUnlock()

    snapshot := ta.metrics.GetSnapshot()
    ranks := make([]*ComponentRank, 0, len(snapshot.ComponentStats))

    for _, stat := range snapshot.ComponentStats {
        var value int64
        switch metric {
        case MetricMutation:
            value = stat.MutationCount
        case MetricLayout:
            value = stat.LayoutCount
        case MetricRepaint:
            value = stat.RepaintCount
        }

        ranks = append(ranks, &ComponentRank{
            NodeID: stat.NodeID,
            Value:  value,
        })
    }

    // 排序
    sort.Slice(ranks, func(i, j int) bool {
        return ranks[i].Value > ranks[j].Value
    })

    // Top N
    if len(ranks) > ta.topN {
        ranks = ranks[:ta.topN]
    }

    return ranks
}
```

### 7.4 PercentileCalc

```go
// PercentileCalc 百分位计算 (V1)
type PercentileCalc struct {
    mu sync.RWMutex

    samples map[NodeID]*RingBuffer  // 固定窗口，防止 O(n²)
}

// RingBuffer 固定大小环形缓冲区
type RingBuffer struct {
    data     []float64
    size     int
    writePos int
    count    int
}

func NewRingBuffer(size int) *RingBuffer {
    return &RingBuffer{
        data: make([]float64, size),
        size: size,
    }
}

func (rb *RingBuffer) Add(value float64) {
    rb.data[rb.writePos] = value
    rb.writePos = (rb.writePos + 1) % rb.size
    if rb.count < rb.size {
        rb.count++
    }
}

func (rb *RingBuffer) Percentile(p float64) float64 {
    if rb.count == 0 {
        return 0
    }

    // 复制数据进行排序
    sorted := make([]float64, rb.count)
    copy(sorted, rb.data[:rb.count])
    sort.Float64s(sorted)

    k := int(float64(len(sorted)-1) * p)
    return sorted[k]
}
```

### 7.5 TimeSeriesStore

```go
// TimeSeriesStore 时序存储 (V1)
type TimeSeriesStore struct {
    mu sync.RWMutex

    // 固定窗口，防止内存泄漏
    windowSize int  // 默认 30 帧
    series     map[NodeID]*TimeSeries

    // TTL 自动清理
    ttl time.Duration
}

// TimeSeries 单个组件时序
type TimeSeries struct {
    NodeID  NodeID
    Points  []DataPoint  // 固定长度
    writePos int
}

type DataPoint struct {
    FrameID   FrameID
    Timestamp time.Time
    Value     float64
}

func (ts *TimeSeriesStore) AddPoint(nodeID NodeID, value float64) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    series := ts.getOrCreate(nodeID)

    // 覆盖旧数据 (固定窗口)
    series.Points[series.writePos] = DataPoint{
        FrameID:   currentFrameID,
        Timestamp: time.Now(),
        Value:     value,
    }
    series.writePos = (series.writePos + 1) % len(series.Points)
}

// CleanupExpired 清理过期数据
func (ts *TimeSeriesStore) CleanupExpired() {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    now := time.Now()
    for nodeID, series := range ts.series {
        // 检查最后更新时间
        if now.Sub(series.Points[series.writePos-1].Timestamp) > ts.ttl {
            delete(ts.series, nodeID)
        }
    }
}
```

---

## 八、文件结构

### 8.1 V1 文件结构

```
mint/devtools/analysis/              # 阶段6: 分析引擎
├── analyzer.go                      # 分析器接口
├── level.go                         # 运行级别系统
├── coordinator.go                   # 分析器协调器
│
├── metrics/                         # V1: 指标收集
│   ├── collector.go                 # 指标收集器
│   ├── snapshot.go                  # 快照
│   └── types.go                     # 类型定义
│
├── stats/                           # V1: 统计分析
│   ├── topn.go                      # Top N 分析
│   ├── percentile.go                # 百分位计算
│   └── distribution.go              # 分布统计
│
├── timeseries/                      # V1: 时序存储
│   ├── store.go                     # 时序存储
│   ├── ring_buffer.go               # 环形缓冲区
│   └── ttl.go                       # TTL 清理
│
├── hotspots/                        # V2: 性能热区分析
│   ├── analyzer.go
│   ├── component.go
│   └── config.go
│
├── waste/                           # V2: 无效刷新检测
│   ├── detector.go
│   ├── noop.go
│   └── redundant.go
│
├── jitter/                          # V2: 布局抖动检测
│   ├── detector.go
│   ├── pattern.go
│   └── history.go                   # 固定窗口历史
│
├── confidence/                       # V2: 置信度模型
│   ├── model.go                     # 置信度模型
│   ├── statistical.go               # 统计置信
│   ├── pattern.go                   # 模式置信
│   ├── causal.go                    # 因果置信
│   ├── context.go                   # 上下文置信
│   └── historical.go                # 历史置信
│
├── profile/                         # V3: 行为画像
│   ├── profile.go                   # 行为画像
│   ├── ewma.go                      # EWMA 学习
│   └── baseline.go                  # 基线学习
│
├── pattern/                         # V4: 模式学习
│   ├── detector.go
│   ├── mining.go                    # 模式挖掘
│   └── learned.go                   # 学习到的模式
│
├── optimization/                    # V3: 优化建议
│   ├── engine.go
│   ├── rules.go
│   └── suggestions.go
│
├── rewriter/                        # V4: IDE 层代码重写
│   └── (移到 IDE 插件)
│
└── docs/
    └── phase6_design_v2.md          # 本文档
```

### 8.2 模块与版本对应

```
┌─────────────────────────────────────────────────────────────────┐
│                    模块版本矩阵                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  模块                │ V1 │ V2 │ V3 │ V4 │ V5                   │
│  ────────────────────┼────┼────┼────┼────┼────                    │
│  metrics/           │ ✅ │ ✅ │ ✅ │ ✅ │ ✅                      │
│  stats/             │ ✅ │ ✅ │ ✅ │ ✅ │ ✅                      │
│  timeseries/        │ ✅ │ ✅ │ ✅ │ ✅ │ ✅                      │
│  hotspots/          │    │ ✅ │ ✅ │ ✅ │ ✅                      │
│  waste/             │    │ ✅ │ ✅ │ ✅ │ ✅                      │
│  jitter/            │    │ ✅ │ ✅ │ ✅ │ ✅                      │
│  confidence/        │    │ ✅ │ ✅ │ ✅ │ ✅                      │
│  profile/           │    │    │ ✅ │ ✅ │ ✅                      │
│  pattern/           │    │    │    │ ✅ │ ✅                      │
│  optimization/      │    │    │ ✅ │ ✅ │ ✅                      │
│  adaptive/          │    │    │    │    │ ✅                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 九、API 设计

### 9.1 V1 API

```go
// AnalysisEngine 分析引擎主入口
type AnalysisEngine struct {
    level    AnalysisLevel
    metrics  *MetricsCollector
    stats    *StatsAnalyzer
    store    *TimeSeriesStore
}

// SetLevel 设置分析级别 (关键 API)
func (ae *AnalysisEngine) SetLevel(level AnalysisLevel)

// GetMetrics 获取当前指标
func (ae *AnalysisEngine) GetMetrics() *MetricsSnapshot

// GetTopN 获取 Top N 组件
func (ae *AnalysisEngine) GetTopN(metric MetricType, n int) []*ComponentRank

// GetDistribution 获取分布
func (ae *AnalysisEngine) GetDistribution(nodeID NodeID) *Distribution

// GetTimeSeries 获取时序数据
func (ae *AnalysisEngine) GetTimeSeries(nodeID NodeID) []DataPoint

// V1 不提供的 API (V2+)
// - GetInsights()     // V2
// - GetSuggestions()  // V3
// - ApplyFix()        // V4/IDE
```

### 9.2 客户端 API (V1)

```go
// TuiDebugPanel 扩展
type TuiDebugPanel struct {
    // ... 现有字段

    analysisEngine *AnalysisEngine
}

// ShowMetrics 显示指标 (V1)
func (p *TuiDebugPanel) ShowMetrics() {
    metrics := p.analysisEngine.GetMetrics()
    // 显示纯统计数据
}

// ShowTopN 显示 Top N (V1)
func (p *TuiDebugPanel) ShowTopN(metric MetricType) {
    topN := p.analysisEngine.GetTopN(metric, 10)
    // 显示排行榜
}

// ShowInsights 显示洞察 (V2)
func (p *TuiDebugPanel) ShowInsights() {
    // V2 才实现
}

// ShowSuggestions 显示建议 (V3)
func (p *TuiDebugPanel) ShowSuggestions() {
    // V3 才实现
}
```

### 9.3 配置 API

```go
// AnalysisConfig 分析配置
type AnalysisConfig struct {
    // 运行级别
    Level AnalysisLevel

    // 窗口大小
    TimeSeriesWindow int    // 默认 30
    RingBufferSize   int    // 默认 100

    // TTL
    InsightTTL   time.Duration
    SeriesTTL    time.Duration

    // Top N
    DefaultTopN  int  // 默认 10

    // 百分位
    Percentiles  []float64  // [0.5, 0.95, 0.99]
}

// DefaultAnalysisConfig V1 默认配置
func DefaultAnalysisConfig() *AnalysisConfig {
    return &AnalysisConfig{
        Level:           LevelLight,  // V1 默认轻量模式
        TimeSeriesWindow: 30,
        RingBufferSize:   100,
        InsightTTL:       5 * time.Minute,
        SeriesTTL:        10 * time.Minute,
        DefaultTopN:      10,
        Percentiles:      []float64{0.5, 0.95, 0.99},
    }
}
```

---

## 附录

### A. V1 验收标准

- [ ] Level 0 开启时零开销 (< 0.01%)
- [ ] Level 1 开启时 < 1% 开销
- [ ] 所有统计数据准确
- [ ] 内存占用可控 (固定窗口)
- [ ] 无内存泄漏 (TTL 生效)
- [ ] 稳定运行 2 周无崩溃

### B. 性能目标

| 级别 | 目标开销 |
|------|----------|
| Level 0 | < 0.01% (分支预测) |
| Level 1 | < 1% (原子计数) |
| Level 2 | < 3% (有锁，固定窗口) |
| Level 3 | < 5% (全功能) |

### C. 与 V1 设计对比

| 方面 | V1 设计 | V2 设计 |
|------|--------|--------|
| 运行开销 | 承诺"零开销" | 分级，Level 0 才零开销 |
| 历史窗口 | 无限增长 | 固定 Ring Buffer |
| 内存管理 | 无 TTL | 自动清理过期数据 |
| 术语 | React 风格 | TUI 通用 |
| 上线策略 | 一次性全上 | V1-V5 分阶段 |
| 判断逻辑 | 规则=建议 | 置信度模型 |
| 异常判断 | 绝对阈值 | 相对画像 |
| 自动修复 | Runtime 改代码 | IDE 插件 |

---

## 总结

V2 设计的核心转变：

1. **从"智能工具"到"可演进系统"**
2. **从"绝对判断"到"概率评估"**
3. **从"固定规则"到"行为画像"**
4. **从"Runtime 改代码"到"IDE 插件"**

V1 的成功标准不是"修复了问题"，而是：

> **"我们终于知道什么是正常"**
