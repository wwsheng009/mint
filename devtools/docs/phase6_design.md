# DevTools 阶段6: 高级功能详细实施方案

> **项目**: Mint TUI Runtime
> **文档版本**: 1.0
> **创建日期**: 2026-01-30
> **状态**: 设计中
> **依赖**: 阶段1-5 已完成

---

## 目录

1. [概述](#一概述)
2. [架构设计](#二架构设计)
3. [模块6.1: 性能分析引擎](#三模块61-性能分析引擎)
4. [模块6.2: 自动优化建议](#四模块62-自动优化建议)
5. [模块6.3: 规则引擎](#五模块63-规则引擎)
6. [数据流设计](#六数据流设计)
7. [文件结构](#七文件结构)
8. [实施计划](#八实施计划)
9. [API 设计](#九api-设计)

---

## 一、概述

### 1.1 目标

阶段6 在阶段1-5 已有的数据收集能力基础上，增加**智能分析层**：

```
┌─────────────────────────────────────────────────────────────────┐
│                   阶段6: 智能分析层                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  已有能力 (阶段1-5):         新增能力 (阶段6):                    │
│  ┌─────────────────┐        ┌─────────────────────────────┐    │
│  │ 数据收集         │   →    │ 智能分析                     │    │
│  │ • Event Trace   │        │ • 性能热区检测               │    │
│  │ • Mutation Log  │        │ • 无效刷新检测               │    │
│  │ • Layout Delta  │        │ • 布局抖动检测               │    │
│  │ • Repaint Delta │        │ • 异常模式识别               │    │
│  │ • Causal Graph  │        └─────────────────────────────┘    │
│  └─────────────────┘                ↓                          │
│           │                  ┌─────────────────────────────┐    │
│           │                  │ 优化建议                     │    │
│           │                  │ • 状态粒度分析               │    │
│           ▼                  │ • 代码重构建议               │    │
│  ┌─────────────────┐        │ • 自动修复                   │    │
│  │ 调试界面         │        └─────────────────────────────┘    │
│  │ • TUI Panel     │                                           │
│  │ • Web Dashboard │                                           │
│  └─────────────────┘                                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 核心能力矩阵

| 能力 | 描述 | 优先级 | 复杂度 |
|------|------|--------|--------|
| 性能热区分析 | 识别高频变化的组件和操作 | P0 | 中 |
| 无效刷新检测 | 检测无视觉效果的状态更新 | P0 | 中 |
| 布局抖动检测 | 检测反复变化的布局 | P1 | 高 |
| 状态粒度分析 | 分析状态划分是否合理 | P1 | 高 |
| 优化建议生成 | 基于规则生成优化建议 | P1 | 中 |
| 自动修复 | 自动应用某些优化 | P2 | 高 |

### 1.3 设计原则

1. **零运行时开销**: 分析在后台异步进行，不影响主线程
2. **可配置规则**: 支持自定义分析规则和阈值
3. **增量分析**: 只分析新增数据，避免重复计算
4. **可操作建议**: 生成具体可执行的优化建议

---

## 二、架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        阶段6 分析引擎架构                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                     Runtime (阶段1-5)                            │   │
│   │  EventBus → CausalGraph → FrameTimeline → TimeTravelStore       │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                              ↓ 数据流                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│  │                   Analysis Engine (阶段6)                         │   │
│   │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐        │   │
│   │  │HotspotAnalyzer│  │WasteDetector  │  │JitterDetector │        │   │
│   │  │               │  │               │  │               │        │   │
│   │  │• MutationFreq │  │• NoOpSetState │  │• LayoutJitter │        │   │
│   │  │• LayoutFreq   │  │• NoOpLayout   │  │• PaintFlicker │        │   │
│   │  │• PaintFreq    │  │• OverPaint    │  │• Oscillation  │        │   │
│   │  └───────────────┘  └───────────────┘  └───────────────┘        │   │
│   │                              ↓                                   │   │
│   │  ┌─────────────────────────────────────────────────────────┐    │   │
│   │  │                PatternDetector                          │    │   │
│   │  │  • CascadeUpdate (级联更新)                              │    │   │
│   │  │  • Thrashing (颠簸)                                      │    │   │
│   │  │  • MemoryLeak (内存泄漏)                                 │    │   │
│   │  └─────────────────────────────────────────────────────────┘    │   │
│   │                              ↓                                   │   │
│   │  ┌─────────────────────────────────────────────────────────┐    │   │
│   │  │                OptimizationEngine                        │    │   │
│   │  │  • StateGranularity (状态粒度)                            │    │   │
│   │  │  • MergeStrategy (合并策略)                              │    │   │
│   │  │  • LiftStrategy (提升策略)                                │    │   │
│   │  └─────────────────────────────────────────────────────────┘    │   │
│   │                              ↓                                   │   │
│   │  ┌─────────────────────────────────────────────────────────┐    │   │
│   │  │                RuleEngine                                │    │   │
│   │  │  • PerformanceRules (性能规则)                            │    │   │
│   │  │  • AntipatternRules (反模式规则)                          │    │   │
│   │  │  • CustomRules (自定义规则)                               │    │   │
│   │  └─────────────────────────────────────────────────────────┘    │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                              ↓ 输出                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                     Insights Store                              │   │
│   │  • Hotspots (热区)    • Wastes (浪费)    • Jitters (抖动)        │   │
│   │  • Patterns (模式)    • Suggestions (建议)                       │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流

```
FrameRecord (阶段5) → AnalysisQueue → Analyzer → Insight → Storage → Client
                                                      ↓
                                                 Rule Engine
                                                      ↓
                                                 Suggestion
```

### 2.3 核心接口

```go
// 分析器接口
type Analyzer interface {
    // 名称
    Name() string

    // 分析帧数据
    Analyze(frame *FrameRecord) *Insight

    // 分析时间窗口
    AnalyzeWindow(frames []*FrameRecord) []*Insight

    // 重置状态
    Reset()
}

// 洞察接口
type Insight struct {
    ID         InsightID
    Type       InsightType
    Severity   Severity
    Title      string
    Description string
    Location   Location
    Evidence   []Evidence
    Suggestion *Suggestion
    Timestamp  time.Time
}

// 建议接口
type Suggestion struct {
    ID         SuggestionID
    Type       SuggestionType
    Title      string
    Description string
    CodeDelta  *CodeDelta
    Confidence float32
    AutoFix    bool
}
```

---

## 三、模块6.1: 性能分析引擎

### 3.1 性能热区分析 (HotspotAnalyzer)

#### 3.1.1 检测指标

| 指标 | 计算方式 | 阈值 |
|------|----------|------|
| Mutation Frequency | mutations / second | > 100/s |
| Layout Frequency | layouts / second | > 60/s |
| Repaint Frequency | repaints / second | > 60/s |
| Component Impact | mutations × affected_nodes | Top 10% |
| Frame Time Contribution | duration / total_frame_time | > 20% |

#### 3.1.2 数据结构

```go
// HotspotAnalyzer 性能热区分析器
type HotspotAnalyzer struct {
    mu sync.RWMutex

    // 窗口统计
    windowSize int              // 分析窗口大小（帧数）
    minSamples int              // 最小样本数

    // 组件统计
    componentStats map[NodeID]*ComponentHotspot

    // 全局统计
    globalStats *GlobalHotspotStats

    // 配置
    config HotspotConfig
}

// HotspotConfig 热区分析配置
type HotspotConfig struct {
    // 阈值
    MutationFreqThreshold    float64  // Mutation 频率阈值 (次/秒)
    LayoutFreqThreshold      float64  // 布局频率阈值 (次/秒)
    RepaintFreqThreshold    float64  // 重绘频率阈值 (次/秒)

    // 时间窗口
    AnalysisWindowSeconds   int      // 分析窗口时长 (秒)

    // 百分比
    TopPercentage          float64  // Top N 百分比
    ImpactThreshold        float64  // 影响阈值
}

// ComponentHotspot 组件热区统计
type ComponentHotspot struct {
    NodeID       NodeID
    ComponentID  string

    // Mutation 统计
    MutationCount   int64
    MutationRate    float64  // mutations per second

    // Layout 统计
    LayoutCount     int64
    LayoutRate      float64

    // Repaint 统计
    RepaintCount    int64
    RepaintRate     float64

    // 影响范围
    AffectedNodes   int      // 影响的子节点数
    ImpactScore     float64  // 综合影响分数

    // 时间
    FirstSeen       time.Time
    LastSeen        time.Time
}

// GlobalHotspotStats 全局热区统计
type GlobalHotspotStats struct {
    TotalFrames     int64
    TotalMutations  int64
    TotalLayouts    int64
    TotalRepaints   int64

    WindowDuration  time.Duration
    Samples         int64
}
```

#### 3.1.3 分析算法

```go
// Analyze 分析单帧
func (ha *HotspotAnalyzer) Analyze(frame *FrameRecord) *HotspotInsight {
    ha.mu.Lock()
    defer ha.mu.Unlock()

    // 更新组件统计
    for _, mutation := range frame.Mutations {
        ha.updateMutationStats(mutation)
    }

    for _, layout := range frame.Layouts {
        ha.updateLayoutStats(layout)
    }

    // 检查是否达到阈值
    insights := ha.checkThresholds()

    return insights
}

// GetHotspots 获取当前热区
func (ha *HotspotAnalyzer) GetHotspots(limit int) []*ComponentHotspot {
    ha.mu.RLock()
    defer ha.mu.RUnlock()

    hotspots := make([]*ComponentHotspot, 0, len(ha.componentStats))
    for _, stat := range ha.componentStats {
        hotspots = append(hotspots, stat)
    }

    // 按影响分数排序
    sort.Slice(hotspots, func(i, j int) bool {
        return hotspots[i].ImpactScore > hotspots[j].ImpactScore
    })

    if limit > 0 && len(hotspots) > limit {
        hotspots = hotspots[:limit]
    }

    return hotspots
}

// checkThresholds 检查阈值
func (ha *HotspotAnalyzer) checkThresholds() *HotspotInsight {
    var violations []*HotspotViolation

    for _, stat := range ha.componentStats {
        // 检查 Mutation 频率
        if stat.MutationRate > ha.config.MutationFreqThreshold {
            violations = append(violations, &HotspotViolation{
                NodeID:    stat.NodeID,
                Type:      ViolationMutation,
                Actual:    stat.MutationRate,
                Threshold: ha.config.MutationFreqThreshold,
            })
        }

        // 检查 Layout 频率
        if stat.LayoutRate > ha.config.LayoutFreqThreshold {
            violations = append(violations, &HotspotViolation{
                NodeID:    stat.NodeID,
                Type:      ViolationLayout,
                Actual:    stat.LayoutRate,
                Threshold: ha.config.LayoutFreqThreshold,
            })
        }

        // 检查 Repaint 频率
        if stat.RepaintRate > ha.config.RepaintFreqThreshold {
            violations = append(violations, &HotspotViolation{
                NodeID:    stat.NodeID,
                Type:      ViolationRepaint,
                Actual:    stat.RepaintRate,
                Threshold: ha.config.RepaintFreqThreshold,
            })
        }
    }

    if len(violations) == 0 {
        return nil
    }

    return &HotspotInsight{
        ID:         generateInsightID(),
        Type:       InsightHotspot,
        Severity:   SeverityWarning,
        Title:      "Performance Hotspot Detected",
        Violations: violations,
        Timestamp:  time.Now(),
    }
}
```

### 3.2 无效刷新检测 (WasteDetector)

#### 3.2.1 检测类型

| 类型 | 描述 | 检测方法 |
|------|------|----------|
| NoOp SetState | setState 后值不变 | oldValue == newValue |
| NoOp Layout | 布局计算后结果相同 | rect 无变化 |
| OverPaint | 重绘区域无实际变化 | dirty 区域内容相同 |
| Redundant Update | 同一帧内多次更新同一字段 | 重复 mutation |

#### 3.2.2 数据结构

```go
// WasteDetector 无效刷新检测器
type WasteDetector struct {
    mu sync.RWMutex

    // 帧内状态跟踪
    currentFrameMutations map[NodeID]map[string]int  // nodeID -> field -> count

    // 累计统计
    wasteStats map[NodeID]*ComponentWaste

    // 配置
    config WasteConfig
}

// WasteConfig 浪费检测配置
type WasteConfig struct {
    EnableNoOpSetState   bool
    EnableNoOpLayout     bool
    EnableOverPaint      bool
    EnableRedundantUpdate bool
}

// ComponentWaste 组件浪费统计
type ComponentWaste struct {
    NodeID      NodeID

    // NoOp 统计
    NoOpMutations   int64
    NoOpLayouts     int64

    // 冗余更新
    RedundantUpdates int64

    // 过度重绘
    OverPaintCells   int64

    // 浪费分数
    WasteScore       float64
}

// WasteInsight 浪费洞察
type WasteInsight struct {
    ID          InsightID
    Type        InsightType
    Severity    Severity

    // 浪费详情
    NoOpMutations  []*NoOpMutation
    NoOpLayouts    []*NoOpLayout
    RedundantUpdates []*RedundantUpdate
    OverPaints     []*OverPaint

    // 建议
    Suggestions []*Suggestion
}

// NoOpMutation 无效状态变更
type NoOpMutation struct {
    NodeID     NodeID
    Field      string
    OldValue   interface{}
    NewValue   interface{}
    FrameID    FrameID
    StackTrace string  // 调用栈
}

// NoOpLayout 无效布局
type NoOpLayout struct {
    NodeID    NodeID
    OldRect   *Rect
    NewRect   *Rect
    FrameID   FrameID
}
```

#### 3.2.3 检测算法

```go
// DetectNoOpMutation 检测无效状态变更
func (wd *WasteDetector) DetectNoOpMutation(mutation *MutationNode) *NoOpMutation {
    // 检查值是否实际变化
    if isEqual(mutation.OldValue, mutation.NewValue) {
        return &NoOpMutation{
            NodeID:    NodeID(mutation.Component),
            Field:     mutation.Field,
            OldValue:  mutation.OldValue,
            NewValue:  mutation.NewValue,
            FrameID:   currentFrameID,
        }
    }
    return nil
}

// DetectRedundantUpdate 检测冗余更新
func (wd *WasteDetector) DetectRedundantUpdate(mutation *MutationNode) bool {
    wd.mu.Lock()
    defer wd.mu.Unlock()

    nodeID := NodeID(mutation.Component)
    field := mutation.Field

    // 检查本帧是否已更新过该字段
    if fields, ok := wd.currentFrameMutations[nodeID]; ok {
        if count, exists := fields[field]; exists && count > 0 {
            return true
        }
    }

    // 记录本次更新
    if _, ok := wd.currentFrameMutations[nodeID]; !ok {
        wd.currentFrameMutations[nodeID] = make(map[string]int)
    }
    wd.currentFrameMutations[nodeID][field]++

    return false
}

// DetectOverPaint 检测过度重绘
func (wd *WasteDetector) DetectOverPaint(oldContent, newContent []byte) bool {
    // 简单比较：内容相同则无需重绘
    return bytes.Equal(oldContent, newContent)
}
```

### 3.3 布局抖动检测 (JitterDetector)

#### 3.3.1 检测类型

| 类型 | 描述 | 检测方法 |
|------|------|----------|
| Layout Jitter | 布局反复变化 | 短时间内多次变化 |
| Paint Flicker | 重绘区域闪烁 | dirty 区域快速切换 |
| Size Oscillation | 尺寸振荡 | 宽高来回变化 |
| Position Drift | 位置漂移 | 坐标微小变化 |

#### 3.3.2 数据结构

```go
// JitterDetector 布局抖动检测器
type JitterDetector struct {
    mu sync.RWMutex

    // 历史跟踪
    layoutHistory map[NodeID]*LayoutHistory
    paintHistory  map[NodeID]*PaintHistory

    // 配置
    config JitterConfig
}

// JitterConfig 抖动检测配置
type JitterConfig struct {
    // 时间窗口
    JitterWindowMs        int64     // 抖动检测窗口 (毫秒)
    MinJitterCount        int       // 最小抖动次数

    // 尺寸变化阈值
    SizeDeltaThreshold    int       // 尺寸变化阈值 (像素)

    // 位置变化阈值
    PositionDeltaThreshold int      // 位置变化阈值 (像素)
}

// LayoutHistory 布局历史
type LayoutHistory struct {
    NodeID     NodeID
    Snapshots  []*LayoutSnapshot
    JitterCount int
}

// LayoutSnapshot 布局快照
type LayoutSnapshot struct {
    Rect      *Rect
    Timestamp time.Time
    FrameID   FrameID
}

// JitterInsight 抖动洞察
type JitterInsight struct {
    ID           InsightID
    Type         InsightType
    Severity     Severity

    // 抖动详情
    LayoutJitters []*LayoutJitter
    PaintFlickers []*PaintFlicker

    // 分析结果
    Pattern       JitterPattern
    Frequency     float64
    Amplitude     int
}

// LayoutJitter 布局抖动
type LayoutJitter struct {
    NodeID       NodeID
    Snapshots    []*LayoutSnapshot
    JitterType   JitterType
    Frequency    float64  // 抖动频率
}

type JitterType int

const (
    JitterSize JitterType = iota
    JitterPosition
    JitterOscillation
)

type JitterPattern int

const (
    PatternUnknown JitterPattern = iota
    PatternSawtooth      // 锯齿波
    PatternSquare        // 方波
    PatternRandom        // 随机
    PatternDecay         // 衰减
)
```

#### 3.3.3 检测算法

```go
// DetectLayoutJitter 检测布局抖动
func (jd *JitterDetector) DetectLayoutJitter(nodeID NodeID, rect *Rect, frameID FrameID) *LayoutJitter {
    jd.mu.Lock()
    defer jd.mu.Unlock()

    // 获取或创建历史
    history, ok := jd.layoutHistory[nodeID]
    if !ok {
        history = &LayoutHistory{
            NodeID:    nodeID,
            Snapshots: make([]*LayoutSnapshot, 0, jd.config.MinJitterCount*2),
        }
        jd.layoutHistory[nodeID] = history
    }

    // 添加当前快照
    snapshot := &LayoutSnapshot{
        Rect:      rect,
        Timestamp: time.Now(),
        FrameID:   frameID,
    }
    history.Snapshots = append(history.Snapshots, snapshot)

    // 清理旧快照
    jd.cleanupOldSnapshots(history)

    // 检查是否有抖动
    if len(history.Snapshots) < jd.config.MinJitterCount {
        return nil
    }

    // 分析抖动模式
    jitter := jd.analyzeJitterPattern(history)
    if jitter != nil {
        history.JitterCount++
        return jitter
    }

    return nil
}

// analyzeJitterPattern 分析抖动模式
func (jd *JitterDetector) analyzeJitterPattern(history *LayoutHistory) *LayoutJitter {
    snapshots := history.Snapshots
    n := len(snapshots)

    // 检测尺寸振荡
    sizeChanges := 0
    for i := 1; i < n; i++ {
        if jd.sizeChanged(snapshots[i-1].Rect, snapshots[i].Rect) {
            sizeChanges++
        }
    }

    // 检测位置振荡
    posChanges := 0
    for i := 1; i < n; i++ {
        if jd.positionChanged(snapshots[i-1].Rect, snapshots[i].Rect) {
            posChanges++
        }
    }

    // 判断抖动类型
    if sizeChanges >= jd.config.MinJitterCount {
        return &LayoutJitter{
            NodeID:     history.NodeID,
            Snapshots:  snapshots,
            JitterType: JitterSize,
            Frequency:  jd.calculateFrequency(snapshots),
        }
    }

    if posChanges >= jd.config.MinJitterCount {
        return &LayoutJitter{
            NodeID:     history.NodeID,
            Snapshots:  snapshots,
            JitterType: JitterPosition,
            Frequency:  jd.calculateFrequency(snapshots),
        }
    }

    return nil
}

// calculateFrequency 计算抖动频率
func (jd *JitterDetector) calculateFrequency(snapshots []*LayoutSnapshot) float64 {
    if len(snapshots) < 2 {
        return 0
    }

    duration := snapshots[len(snapshots)-1].Timestamp.Sub(snapshots[0].Timestamp)
    if duration <= 0 {
        return 0
    }

    return float64(len(snapshots)) / duration.Seconds()
}
```

---

## 四、模块6.2: 自动优化建议

### 4.1 优化建议引擎 (OptimizationEngine)

#### 4.1.1 建议类型

| 类型 | 描述 | 优先级 |
|------|------|--------|
| 批量更新 | 合并多次 setState | P0 |
| 记忆化 | 添加 useMemo 缓存 | P1 |
| 状态提升 | 将状态移到父组件 | P1 |
| 条件渲染 | 添加条件判断避免无效渲染 | P1 |
| 防抖/节流 | 对高频事件添加防抖 | P2 |

#### 4.1.2 数据结构

```go
// OptimizationEngine 优化建议引擎
type OptimizationEngine struct {
    mu sync.RWMutex

    // 规则集
    rules []OptimizationRule

    // 建议缓存
    suggestions map[SuggestionID]*Suggestion

    // 配置
    config OptimizationConfig
}

// OptimizationRule 优化规则接口
type OptimizationRule interface {
    // 规则名称
    Name() string

    // 检查是否适用
    Check(insight Insight) bool

    // 生成建议
    Generate(insight Insight) *Suggestion
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
    EnableBatchUpdate      bool
    EnableMemoization      bool
    EnableStateLifting     bool
    EnableConditionalRender bool
    EnableDebounce         bool

    // 置信度阈值
    MinConfidence          float32
}

// Suggestion 优化建议
type Suggestion struct {
    ID           SuggestionID
    Type         SuggestionType
    Title        string
    Description  string
    Location     Location
    Confidence   float32  // 0-1
    Impact       ImpactLevel
    AutoFix      bool
    CodeDelta    *CodeDelta
    Instructions string
}

type SuggestionType int

const (
    SuggestionBatchUpdate SuggestionType = iota
    SuggestionMemoization
    SuggestionStateLifting
    SuggestionConditionalRender
    SuggestionDebounce
)

type ImpactLevel int

const (
    ImpactLow ImpactLevel = iota
    ImpactMedium
    ImpactHigh
    ImpactCritical
)

// CodeDelta 代码变更
type CodeDelta struct {
    FilePath    string
    OldLine     int
    NewLine     int
    OldCode     string
    NewCode     string
    Explanation string
}
```

#### 4.1.3 规则实现

```go
// BatchUpdateRule 批量更新规则
type BatchUpdateRule struct {
    config BatchUpdateConfig
}

type BatchUpdateConfig struct {
    MinMutations    int     // 最小突变次数
    MaxIntervalMs   int64   // 最大间隔 (毫秒)
}

func (r *BatchUpdateRule) Check(insight Insight) bool {
    hotspot, ok := insight.(*HotspotInsight)
    if !ok {
        return false
    }

    // 检查是否有高频 setState
    for _, v := range hotspot.Violations {
        if v.Type == ViolationMutation &&
            v.Actual > float64(r.config.MinMutations) {
            return true
        }
    }

    return false
}

func (r *BatchUpdateRule) Generate(insight Insight) *Suggestion {
    hotspot := insight.(*HotspotInsight)

    // 找到违反规则的组件
    var nodeID NodeID
    for _, v := range hotspot.Violations {
        if v.Type == ViolationMutation {
            nodeID = v.NodeID
            break
        }
    }

    return &Suggestion{
        Type:        SuggestionBatchUpdate,
        Title:       "Batch State Updates",
        Description: fmt.Sprintf("Component %s has %d mutations in a short time. Consider batching updates.",
            nodeID, r.config.MinMutations),
        Confidence:  0.9,
        Impact:      ImpactHigh,
        AutoFix:     false,
        Instructions: `Instead of:
    setState({foo: value1})
    setState({bar: value2})

Use:
    setState({foo: value1, bar: value2})`,
    }
}

// MemoizationRule 记忆化规则
type MemoizationRule struct {
    config MemoizationConfig
}

type MemoizationConfig struct {
    ExpensiveOperations int      // 昂贵操作次数
    UnchangedInputs     float64  // 未变化输入比例
}

func (r *MemoizationRule) Check(insight Insight) bool {
    waste, ok := insight.(*WasteInsight)
    if !ok {
        return false
    }

    // 检查是否有重复计算
    return len(waste.NoOpMutations) > 0
}

func (r *MemoizationRule) Generate(insight Insight) *Suggestion {
    waste := insight.(*WasteInsight)

    return &Suggestion{
        Type:        SuggestionMemoization,
        Title:       "Add Memoization",
        Description: fmt.Sprintf("Component has %d redundant updates. Consider memoizing.",
            len(waste.NoOpMutations)),
        Confidence:  0.8,
        Impact:      ImpactMedium,
        AutoFix:     false,
        Instructions: `Use useMemo to cache expensive computations:
    const memoizedValue = useMemo(() => {
        return expensiveComputation(props.data)
    }, [props.data])`,
    }
}

// StateLiftingRule 状态提升规则
type StateLiftingRule struct {
    config StateLiftingConfig
}

type StateLiftingConfig struct {
    MinChildren     int
    PropDrillDepth  int
}

func (r *StateLiftingRule) Check(insight Insight) bool {
    // 检查级联更新模式
    return true  // 简化
}

func (r *StateLiftingRule) Generate(insight Insight) *Suggestion {
    return &Suggestion{
        Type:        SuggestionStateLifting,
        Title:       "Lift State Up",
        Description: "Multiple components update due to shared state. Consider lifting state.",
        Confidence:  0.7,
        Impact:      ImpactMedium,
        AutoFix:     false,
        Instructions: `Move shared state to the nearest common ancestor:
    // Before
    <Parent>
      <ChildA state={localState} />
      <ChildB state={localState} />
    </Parent>

    // After
    <Parent state={sharedState}>
      <ChildA />
      <ChildB />
    </Parent>`,
    }
}
```

### 4.2 代码重写 (CodeRewriter)

```go
// CodeRewriter 代码重写器
type CodeRewriter struct {
    mu sync.RWMutex

    // 分析器
    analyzers []Analyzer

    // 规则引擎
    ruleEngine *RuleEngine

    // 配置
    config RewriterConfig
}

// RewriterConfig 重写器配置
type RewriterConfig struct {
    EnableAutoFix      bool
    RequireConfirmation bool
    BackupOriginal     bool
}

// GeneratePatch 生成补丁
func (cr *CodeRewriter) GeneratePatch(suggestion *Suggestion) (*Patch, error) {
    if suggestion.CodeDelta == nil {
        return nil, fmt.Errorf("no code delta available")
    }

    return &Patch{
        ID:          generatePatchID(),
        SuggestionID: suggestion.ID,
        FilePath:    suggestion.CodeDelta.FilePath,
        OldLine:     suggestion.CodeDelta.OldLine,
        NewLine:     suggestion.CodeDelta.NewLine,
        OldCode:     suggestion.CodeDelta.OldCode,
        NewCode:     suggestion.CodeDelta.NewCode,
        CreatedAt:   time.Now(),
    }, nil
}

// ApplyPatch 应用补丁
func (cr *CodeRewriter) ApplyPatch(patch *Patch) error {
    if !cr.config.EnableAutoFix {
        return fmt.Errorf("auto-fix is disabled")
    }

    // 备份原文件
    if cr.config.BackupOriginal {
        if err := cr.backupFile(patch.FilePath); err != nil {
            return err
        }
    }

    // 应用补丁
    return cr.applyPatchToFile(patch)
}

// ApplySuggestions 批量应用建议
func (cr *CodeRewriter) ApplySuggestions(suggestions []*Suggestion) (*ApplyResult, error) {
    result := &ApplyResult{
        Applied:  make([]SuggestionID, 0),
        Failed:   make([]SuggestionID, 0),
        Skipped:  make([]SuggestionID, 0),
    }

    for _, suggestion := range suggestions {
        if !suggestion.AutoFix {
            result.Skipped = append(result.Skipped, suggestion.ID)
            continue
        }

        patch, err := cr.GeneratePatch(suggestion)
        if err != nil {
            result.Failed = append(result.Failed, suggestion.ID)
            continue
        }

        if err := cr.ApplyPatch(patch); err != nil {
            result.Failed = append(result.Failed, suggestion.ID)
            continue
        }

        result.Applied = append(result.Applied, suggestion.ID)
    }

    return result, nil
}

// Patch 补丁
type Patch struct {
    ID           PatchID
    SuggestionID SuggestionID
    FilePath     string
    OldLine      int
    NewLine      int
    OldCode      string
    NewCode      string
    CreatedAt    time.Time
    AppliedAt    time.Time
}

// ApplyResult 应用结果
type ApplyResult struct {
    Applied  []SuggestionID
    Failed   []SuggestionID
    Skipped  []SuggestionID
}
```

---

## 五、模块6.3: 规则引擎

### 5.1 规则引擎架构

```go
// RuleEngine 规则引擎
type RuleEngine struct {
    mu sync.RWMutex

    // 规则集
    performanceRules []Rule
    antipatternRules []Rule
    customRules      []Rule

    // 规则索引
    ruleIndex map[RuleType][]Rule

    // 配置
    config RuleEngineConfig
}

// Rule 规则接口
type Rule interface {
    // 规则 ID
    ID() RuleID

    // 规则类型
    Type() RuleType

    // 规则名称
    Name() string

    // 规则描述
    Description() string

    // 严重程度
    Severity() Severity

    // 检查条件
    Check(ctx *RuleContext) bool

    // 执行动作
    Execute(ctx *RuleContext) *RuleResult
}

// RuleType 规则类型
type RuleType int

const (
    RuleTypePerformance RuleType = iota
    RuleTypeAntipattern
    RuleTypeCustom
)

// RuleContext 规则上下文
type RuleContext struct {
    Frame      *FrameRecord
    Insights   []Insight
    Stats      *AnalysisStats
    Parameters map[string]interface{}
}

// RuleResult 规则结果
type RuleResult struct {
    RuleID      RuleID
    Passed      bool
    Message     string
    Suggestion  *Suggestion
    Metrics     map[string]interface{}
}

// AnalysisStats 分析统计
type AnalysisStats struct {
    TotalFrames      int64
    TotalMutations   int64
    TotalLayouts     int64
    TotalRepaints    int64
    AverageFrameTime time.Duration
    P95FrameTime     time.Duration
    P99FrameTime     time.Duration
}
```

### 5.2 内置规则

```go
// HighMutationRateRule 高突变率规则
type HighMutationRateRule struct {
    config HighMutationRateConfig
}

type HighMutationRateConfig struct {
    Threshold    float64  // mutations per second
    WindowFrames int
}

func (r *HighMutationRateRule) ID() RuleID { return "high-mutation-rate" }
func (r *HighMutationRateRule) Type() RuleType { return RuleTypePerformance }
func (r *HighMutationRateRule) Name() string { return "High Mutation Rate" }
func (r *HighMutationRateRule) Description() string {
    return "Detects components with unusually high mutation rates"
}
func (r *HighMutationRateRule) Severity() Severity { return SeverityWarning }

func (r *HighMutationRateRule) Check(ctx *RuleContext) bool {
    if ctx.Stats == nil {
        return false
    }

    // 计算平均突变率
    rate := float64(ctx.Stats.TotalMutations) /
        ctx.Stats.TotalFrames * 60  // per second at 60fps

    return rate > r.config.Threshold
}

func (r *HighMutationRateRule) Execute(ctx *RuleContext) *RuleResult {
    return &RuleResult{
        RuleID: r.ID(),
        Passed: false,
        Message: fmt.Sprintf("Mutation rate %.2f/s exceeds threshold %.2f/s",
            float64(ctx.Stats.TotalMutations)/ctx.Stats.TotalFrames*60,
            r.config.Threshold),
        Suggestion: &Suggestion{
            Type:       SuggestionBatchUpdate,
            Title:      "Reduce Mutation Rate",
            Confidence: 0.85,
        },
    }
}

// LayoutThrashingRule 布局颠簸规则
type LayoutThrashingRule struct {
    config LayoutThrashingConfig
}

type LayoutThrashingConfig struct {
    Threshold   int
    WindowMs    int64
}

func (r *LayoutThrashingRule) ID() RuleID { return "layout-thrashing" }
func (r *LayoutThrashingRule) Type() RuleType { return RuleTypeAntipattern }
func (r *LayoutThrashingRule) Name() string { return "Layout Thrashing" }
func (r *LayoutThrashingRule) Description() string {
    return "Detects layout thrashing patterns"
}
func (r *LayoutThrashingRule) Severity() Severity { return SeverityError }

func (r *LayoutThrashingRule) Check(ctx *RuleContext) bool {
    // 检查是否有组件在短时间内多次布局
    for _, frame := range ctx.Frames {
        layoutCount := len(frame.Layouts)
        if layoutCount > r.config.Threshold {
            return true
        }
    }
    return false
}

func (r *LayoutThrashingRule) Execute(ctx *RuleContext) *RuleResult {
    return &RuleResult{
        RuleID: r.ID(),
        Passed: false,
        Message: fmt.Sprintf("Layout thrashing detected: %d layouts exceed threshold %d",
            len(ctx.Frame.Layouts), r.config.Threshold),
        Suggestion: &Suggestion{
            Type:       SuggestionMemoization,
            Title:      "Fix Layout Thrashing",
            Confidence: 0.9,
        },
    }
}

// PaintFlashingRule 重绘闪烁规则
type PaintFlashingRule struct {
    config PaintFlashingConfig
}

type PaintFlashingConfig struct {
    FlickerThreshold int
    WindowMs        int64
}

func (r *PaintFlashingRule) ID() RuleID { return "paint-flashing" }
func (r *PaintFlashingRule) Type() RuleType { return RuleTypeAntipattern }
func (r *PaintFlashingRule) Name() string { return "Paint Flashing" }
func (r *PaintFlashingRule) Description() string {
    return "Detects paint flashing patterns"
}
func (r *PaintFlashingRule) Severity() Severity { return SeverityWarning }

func (r *PaintFlashingRule) Check(ctx *RuleContext) bool {
    // 检查是否有组件快速重绘
    return len(ctx.Frame.Repaints) > r.config.FlickerThreshold
}

func (r *PaintFlashingRule) Execute(ctx *RuleContext) *RuleResult {
    return &RuleResult{
        RuleID: r.ID(),
        Passed: false,
        Message: fmt.Sprintf("Paint flashing detected: %d repaints",
            len(ctx.Frame.Repaints)),
        Suggestion: &Suggestion{
            Type:       SuggestionConditionalRender,
            Title:      "Reduce Paint Frequency",
            Confidence: 0.8,
        },
    }
}
```

### 5.3 自定义规则

```go
// CustomRuleBuilder 自定义规则构建器
type CustomRuleBuilder struct {
    id          RuleID
    ruleType    RuleType
    name        string
    description string
    severity    Severity
    checkFunc   func(*RuleContext) bool
    executeFunc func(*RuleContext) *RuleResult
}

func NewCustomRule(id RuleID) *CustomRuleBuilder {
    return &CustomRuleBuilder{
        id:       id,
        severity: SeverityInfo,
    }
}

func (b *CustomRuleBuilder) Type(t RuleType) *CustomRuleBuilder {
    b.ruleType = t
    return b
}

func (b *CustomRuleBuilder) Name(name string) *CustomRuleBuilder {
    b.name = name
    return b
}

func (b *CustomRuleBuilder) Description(desc string) *CustomRuleBuilder {
    b.description = desc
    return b
}

func (b *CustomRuleBuilder) Severity(s Severity) *CustomRuleBuilder {
    b.severity = s
    return b
}

func (b *CustomRuleBuilder) Check(fn func(*RuleContext) bool) *CustomRuleBuilder {
    b.checkFunc = fn
    return b
}

func (b *CustomRuleBuilder) Execute(fn func(*RuleContext) *RuleResult) *CustomRuleBuilder {
    b.executeFunc = fn
    return b
}

func (b *CustomRuleBuilder) Build() Rule {
    return &customRule{builder: b}
}

type customRule struct {
    builder *CustomRuleBuilder
}

func (r *customRule) ID() RuleID { return r.builder.id }
func (r *customRule) Type() RuleType { return r.builder.ruleType }
func (r *customRule) Name() string { return r.builder.name }
func (r *customRule) Description() string { return r.builder.description }
func (r *customRule) Severity() Severity { return r.builder.severity }
func (r *customRule) Check(ctx *RuleContext) bool {
    if r.builder.checkFunc != nil {
        return r.builder.checkFunc(ctx)
    }
    return false
}
func (r *customRule) Execute(ctx *RuleContext) *RuleResult {
    if r.builder.executeFunc != nil {
        return r.builder.executeFunc(ctx)
    }
    return &RuleResult{RuleID: r.ID(), Passed: true}
}

// 使用示例
func ExampleCustomRule() {
    rule := NewCustomRule("my-custom-rule").
        Type(RuleTypePerformance).
        Name("My Custom Rule").
        Description("Checks for specific pattern").
        Severity(SeverityWarning).
        Check(func(ctx *RuleContext) bool {
            return ctx.Stats.TotalMutations > 1000
        }).
        Execute(func(ctx *RuleContext) *RuleResult {
            return &RuleResult{
                RuleID:  "my-custom-rule",
                Passed:  false,
                Message: "Custom rule violation",
            }
        }).
        Build()

    engine := NewRuleEngine()
    engine.AddRule(rule)
}
```

---

## 六、数据流设计

### 6.1 分析流水线

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        分析流水线                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  FrameRecord (阶段1-5)                                                   │
│       ↓                                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    AnalysisQueue                                 │    │
│  │  • 无锁队列                                                      │    │
│  │  • 背压处理                                                      │    │
│  │  • 批量消费                                                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│       ↓                                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    AnalyzerCoordinator                          │    │
│  │  • 协调多个分析器                                                │    │
│  │  • 并行分析                                                      │    │
│  │  • 结果聚合                                                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│       ↓                                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    Analyzers                                     │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │    │
│  │  │Hotspot       │  │Waste         │  │Jitter        │          │    │
│  │  │Analyzer      │  │Detector      │  │Detector      │          │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│       ↓                                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    InsightAggregator                             │    │
│  │  • 聚合相关洞察                                                  │    │
│  │  • 去重                                                          │    │
│  │  • 排序                                                          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│       ↓                                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    RuleEngine                                   │    │
│  │  • 规则匹配                                                      │    │
│  │  • 建议生成                                                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│       ↓                                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    InsightStore                                  │    │
│  │  • 持久化存储                                                    │    │
│  │  • 索引                                                          │    │
│  │  • 查询                                                          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│       ↓                                                                  │
│  Client (TUI / Web Dashboard)                                            │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.2 数据结构

```go
// AnalysisQueue 分析队列
type AnalysisQueue struct {
    frames    chan *FrameRecord
    batchSize int
    timeout   time.Duration
}

// AnalyzerCoordinator 分析器协调器
type AnalyzerCoordinator struct {
    analyzers []Analyzer
    queue     *AnalysisQueue
    workers   int
}

// InsightAggregator 洞察聚合器
type InsightAggregator struct {
    insights  map[InsightID]*Insight
    relations map[InsightID][]InsightID  // 相关洞察
}

// InsightStore 洞察存储
type InsightStore struct {
    mu       sync.RWMutex
    insights map[InsightID]*Insight

    // 索引
    byType     map[InsightType][]InsightID
    bySeverity map[Severity][]InsightID
    byNode     map[NodeID][]InsightID
    byTime     []InsightID  // 时间排序

    // 配置
    maxInsights int
    ttl         time.Duration
}
```

---

## 七、文件结构

```
mint/devtools/
├── analysis/                    # ✨ 阶段6: 分析引擎 (新建)
│   ├── analyzer.go              # 分析器接口和基础实现
│   ├── coordinator.go           # 分析器协调器
│   ├── queue.go                 # 分析队列
│   ├── aggregator.go            # 洞察聚合器
│   ├── store.go                 # 洞察存储
│   │
│   ├── hotspots/                # 性能热区分析
│   │   ├── analyzer.go          # 热区分析器
│   │   ├── component.go         # 组件热区统计
│   │   └── config.go            # 配置
│   │
│   ├── waste/                   # 无效刷新检测
│   │   ├── detector.go          # 浪费检测器
│   │   ├── noop.go              # NoOp 检测
│   │   ├── redundant.go         # 冗余检测
│   │   └── overpaint.go         # 过度重绘检测
│   │
│   ├── jitter/                  # 布局抖动检测
│   │   ├── detector.go          # 抖动检测器
│   │   ├── layout.go            # 布局抖动
│   │   ├── paint.go             # 重绘闪烁
│   │   └── pattern.go           # 模式识别
│   │
│   ├── pattern/                 # 模式检测
│   │   ├── detector.go          # 模式检测器
│   │   ├── cascade.go           # 级联更新
│   │   ├── thrash.go            # 颠簸检测
│   │   └── memory.go            # 内存泄漏检测
│   │
│   ├── optimization/            # 优化建议
│   │   ├── engine.go            # 优化引擎
│   │   ├── rules.go             # 优化规则
│   │   ├── batch.go             # 批量更新
│   │   ├── memo.go              # 记忆化
│   │   └── lifting.go           # 状态提升
│   │
│   ├── rewriter/                # 代码重写
│   │   ├── rewriter.go          # 重写器
│   │   ├── patch.go             # 补丁生成
│   │   └── apply.go             # 补丁应用
│   │
│   └── rules/                   # 规则引擎
│       ├── engine.go            # 规则引擎
│       ├── rule.go              # 规则接口
│       ├── builtins.go          # 内置规则
│       ├── custom.go            # 自定义规则
│       └── context.go           # 规则上下文
│
├── docs/                        # 文档
│   └── phase6_design.md         # 本文档
│
└── (现有文件: 阶段1-5)           # 已有实现
```

---

## 八、实施计划

### 8.1 实施阶段

| 阶段 | 内容 | 预估时间 | 依赖 |
|------|------|----------|------|
| 6.1 | 分析引擎基础 | 4 小时 | 阶段1-5 |
| 6.2 | 性能热区分析 | 6 小时 | 6.1 |
| 6.3 | 无效刷新检测 | 5 小时 | 6.1 |
| 6.4 | 布局抖动检测 | 6 小时 | 6.1 |
| 6.5 | 模式检测 | 4 小时 | 6.1 |
| 6.6 | 优化建议引擎 | 5 小时 | 6.2-6.5 |
| 6.7 | 代码重写 | 6 小时 | 6.6 |
| 6.8 | 规则引擎 | 5 小时 | 6.1 |
| 6.9 | 集成测试 | 4 小时 | 6.1-6.8 |

**总计**: ~45 小时

### 8.2 详细任务清单

#### 6.1 分析引擎基础 (4 小时)

- [ ] 创建 `analysis/analyzer.go`
  - [ ] 定义 `Analyzer` 接口
  - [ ] 定义 `Insight` 基础结构
  - [ ] 定义 `InsightType` 枚举
  - [ ] 定义 `Severity` 枚举
  - [ ] 定义 `Location` 结构

- [ ] 创建 `analysis/coordinator.go`
  - [ ] 实现 `AnalyzerCoordinator`
  - [ ] 实现并行分析
  - [ ] 实现结果聚合

- [ ] 创建 `analysis/queue.go`
  - [ ] 实现 `AnalysisQueue`
  - [ ] 实现背压处理
  - [ ] 实现批量消费

- [ ] 创建 `analysis/store.go`
  - [ ] 实现 `InsightStore`
  - [ ] 实现索引
  - [ ] 实现查询

- [ ] 创建单元测试

#### 6.2 性能热区分析 (6 小时)

- [ ] 创建 `analysis/hotspots/analyzer.go`
  - [ ] 实现 `HotspotAnalyzer`
  - [ ] 实现 `Analyze()` 方法
  - [ ] 实现 `GetHotspots()` 方法

- [ ] 创建 `analysis/hotspots/component.go`
  - [ ] 实现 `ComponentHotspot`
  - [ ] 实现统计更新
  - [ ] 实现影响分数计算

- [ ] 创建 `analysis/hotspots/config.go`
  - [ ] 实现 `HotspotConfig`
  - [ ] 实现默认配置

- [ ] 创建单元测试

#### 6.3 无效刷新检测 (5 小时)

- [ ] 创建 `analysis/waste/detector.go`
  - [ ] 实现 `WasteDetector`
  - [ ] 实现 `DetectNoOpMutation()`
  - [ ] 实现 `DetectRedundantUpdate()`
  - [ ] 实现 `DetectOverPaint()`

- [ ] 创建 `analysis/waste/noop.go`
  - [ ] NoOp 检测详细实现

- [ ] 创建 `analysis/waste/redundant.go`
  - [ ] 冗余检测详细实现

- [ ] 创建 `analysis/waste/overpaint.go`
  - [ ] 过度重绘检测详细实现

- [ ] 创建单元测试

#### 6.4 布局抖动检测 (6 小时)

- [ ] 创建 `analysis/jitter/detector.go`
  - [ ] 实现 `JitterDetector`
  - [ ] 实现 `DetectLayoutJitter()`
  - [ ] 实现 `DetectPaintFlicker()`

- [ ] 创建 `analysis/jitter/layout.go`
  - [ ] 布局抖动检测详细实现

- [ ] 创建 `analysis/jitter/paint.go`
  - [ ] 重绘闪烁检测详细实现

- [ ] 创建 `analysis/jitter/pattern.go`
  - [ ] 模式识别实现

- [ ] 创建单元测试

#### 6.5 模式检测 (4 小时)

- [ ] 创建 `analysis/pattern/detector.go`
  - [ ] 实现 `PatternDetector`

- [ ] 创建 `analysis/pattern/cascade.go`
  - [ ] 级联更新检测

- [ ] 创建 `analysis/pattern/thrash.go`
  - [ ] 颠簸检测

- [ ] 创建 `analysis/pattern/memory.go`
  - [ ] 内存泄漏检测

- [ ] 创建单元测试

#### 6.6 优化建议引擎 (5 小时)

- [ ] 创建 `analysis/optimization/engine.go`
  - [ ] 实现 `OptimizationEngine`
  - [ ] 实现 `Analyze()` 方法
  - [ ] 实现 `GenerateSuggestions()` 方法

- [ ] 创建 `analysis/optimization/rules.go`
  - [ ] 实现 `OptimizationRule` 接口
  - [ ] 实现内置规则

- [ ] 创建 `analysis/optimization/batch.go`
  - [ ] 批量更新规则

- [ ] 创建 `analysis/optimization/memo.go`
  - [ ] 记忆化规则

- [ ] 创建 `analysis/optimization/lifting.go`
  - [ ] 状态提升规则

- [ ] 创建单元测试

#### 6.7 代码重写 (6 小时)

- [ ] 创建 `analysis/rewriter/rewriter.go`
  - [ ] 实现 `CodeRewriter`

- [ ] 创建 `analysis/rewriter/patch.go`
  - [ ] 实现 `GeneratePatch()`
  - [ ] 实现 `Patch` 结构

- [ ] 创建 `analysis/rewriter/apply.go`
  - [ ] 实现 `ApplyPatch()`
  - [ ] 实现 `ApplySuggestions()`

- [ ] 创建单元测试

#### 6.8 规则引擎 (5 小时)

- [ ] 创建 `analysis/rules/engine.go`
  - [ ] 实现 `RuleEngine`
  - [ ] 实现 `AddRule()`
  - [ ] 实现 `ExecuteRules()`

- [ ] 创建 `analysis/rules/rule.go`
  - [ ] 实现 `Rule` 接口
  - [ ] 实现 `RuleContext`
  - [ ] 实现 `RuleResult`

- [ ] 创建 `analysis/rules/builtins.go`
  - [ ] 实现 `HighMutationRateRule`
  - [ ] 实现 `LayoutThrashingRule`
  - [ ] 实现 `PaintFlashingRule`

- [ ] 创建 `analysis/rules/custom.go`
  - [ ] 实现 `CustomRuleBuilder`

- [ ] 创建 `analysis/rules/context.go`
  - [ ] 实现 `AnalysisStats`

- [ ] 创建单元测试

#### 6.9 集成测试 (4 小时)

- [ ] 创建 `analysis/integration_test.go`
- [ ] 端到端测试
- [ ] 性能测试
- [ ] 压力测试

---

## 九、API 设计

### 9.1 分析 API

```go
// AnalysisEngine 分析引擎主入口
type AnalysisEngine struct {
    coordinator *AnalyzerCoordinator
    store       *InsightStore
    ruleEngine  *RuleEngine
}

// NewAnalysisEngine 创建分析引擎
func NewAnalysisEngine(config *AnalysisConfig) *AnalysisEngine

// Start 启动分析引擎
func (ae *AnalysisEngine) Start() error

// Stop 停止分析引擎
func (ae *AnalysisEngine) Stop() error

// AnalyzeFrame 分析单帧
func (ae *AnalysisEngine) AnalyzeFrame(frame *FrameRecord) error

// GetInsights 获取所有洞察
func (ae *AnalysisEngine) GetInsights(filter *InsightFilter) []*Insight

// GetInsightsByType 按类型获取洞察
func (ae *AnalysisEngine) GetInsightsByType(insightType InsightType) []*Insight

// GetInsightsBySeverity 按严重程度获取洞察
func (ae *AnalysisEngine) GetInsightsBySeverity(severity Severity) []*Insight

// GetInsightsByNode 按组件获取洞察
func (ae *AnalysisEngine) GetInsightsByNode(nodeID NodeID) []*Insight

// GetSuggestions 获取优化建议
func (ae *AnalysisEngine) GetSuggestions() []*Suggestion

// ExecuteRules 执行规则
func (ae *AnalysisEngine) ExecuteRules() []*RuleResult
```

### 9.2 客户端 API

```go
// 扩展 TuiDebugPanel
type TuiDebugPanel struct {
    // ... 现有字段

    analysisEngine *AnalysisEngine
}

// ShowInsights 显示洞察
func (p *TuiDebugPanel) ShowInsights()

// ShowSuggestions 显示建议
func (p *TuiDebugPanel) ShowSuggestions()

// ApplySuggestion 应用建议
func (p *TuiDebugPanel) ApplySuggestion(id SuggestionID) error

// 扩展 WebDashboard
type WebDashboard struct {
    // ... 现有字段

    analysisEngine *AnalysisEngine
}

// GetAnalysisAPI 获取分析 API
func (wd *WebDashboard) GetAnalysisAPI() *AnalysisAPI

// AnalysisAPI 分析 API
type AnalysisAPI struct{}

// GetInsights GET /api/analysis/insights
func (api *AnalysisAPI) GetInsights(w http.ResponseWriter, r *http.Request)

// GetSuggestions GET /api/analysis/suggestions
func (api *AnalysisAPI) GetSuggestions(w http.ResponseWriter, r *http.Request)

// ApplySuggestion POST /api/analysis/apply
func (api *AnalysisAPI) ApplySuggestion(w http.ResponseWriter, r *http.Request)
```

### 9.3 配置 API

```go
// AnalysisConfig 分析配置
type AnalysisConfig struct {
    // 队列配置
    QueueSize    int
    BatchSize    int
    BatchTimeout time.Duration

    // Worker 配置
    Workers      int

    // 存储配置
    MaxInsights  int
    InsightTTL   time.Duration

    // 分析器配置
    Hotspot      *HotspotConfig
    Waste        *WasteConfig
    Jitter       *JitterConfig

    // 优化配置
    Optimization *OptimizationConfig

    // 规则配置
    Rules        *RuleEngineConfig
}

// DefaultAnalysisConfig 默认配置
func DefaultAnalysisConfig() *AnalysisConfig
```

---

## 附录

### A. 验收标准

#### 6.1 分析引擎基础

- [ ] 分析队列吞吐量 > 1000 帧/秒
- [ ] 并行分析无竞态条件
- [ ] 内存占用可控

#### 6.2 性能热区分析

- [ ] 能准确识别高频组件
- [ ] 影响分数计算正确
- [ ] 误报率 < 5%

#### 6.3 无效刷新检测

- [ ] NoOp 检测准确率 > 95%
- [ ] 冗余更新检测准确率 > 90%
- [ ] 过度重绘检测准确率 > 85%

#### 6.4 布局抖动检测

- [ ] 能检测振荡模式
- [ ] 频率计算准确
- [ ] 模式识别正确

#### 6.5 优化建议

- [ ] 建议可执行
- [ ] 置信度评估准确
- [ ] 自动修复安全

### B. 性能目标

| 指标 | 目标 |
|------|------|
| 分析延迟 | < 10ms/帧 |
| 内存占用 | < 50MB |
| CPU 开销 | < 5% |
| 误报率 | < 5% |
| 漏报率 | < 10% |

### C. 参考资料

- Chrome DevTools Performance Panel
- React DevTools Profiler
- Flutter DevTools Performance
- Firefox Performance Tools

---

**文档状态**: 待审核
**下一步**: 创建实施 TODO 并开始编码
