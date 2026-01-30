# DevTools Demo

A comprehensive example demonstrating how to use the Mint TUI DevTools.

## Overview

This example program demonstrates all major features of the Mint DevTools:

- **Basic Usage** - Enable, collect data, and disable DevTools
- **V1 Statistics** - Pure statistics: counts, TopN, percentiles, distributions
- **V2 Pattern Detection** - Behavioral pattern recognition with confidence scoring
- **Causal Chain Tracking** - Event -> Mutation -> Layout -> Repaint tracing
- **Insights Generation** - Confidence-based optimization suggestions

## Building

```bash
cd examples/devtools_demo
go build -o devtools_demo main.go
```

## Running

Run individual demos:

```bash
./devtools_demo basic       # Basic DevTools usage
./devtools_demo observation # V1 Statistics layer
./devtools_demo patterns    # V2 Pattern detection
./devtools_demo causal      # Causal chain tracking
./devtools_demo insights    # Confidence-based insights
./devtools_demo all         # Run all demos
```

## Examples

### 1. Basic Usage

```go
import "github.com/wwsheng009/mint/devtools"

// Create and enable DevTools
dt := devtools.New()
dt.Enable()

// Record frames
dt.BeginFrame()
dt.RecordEvent("keypress", "node-id", "bubble", data)
dt.EndFrame()

// Clean shutdown
dt.Disable()
dt.Shutdown()
```

### 2. Observation Layer (V1 + V2)

```go
import "github.com/wwsheng009/mint/devtools/observation"
import "github.com/wwsheng009/mint/devtools/observation/v1"

// Create observation layer
cfg := observation.DefaultConfig()
layer := observation.NewLayer(cfg)
layer.Enable(v1.LevelAdvanced)

// Record mutations
layer.RecordMutation(nodeID, fieldType, fieldValue)

// Get statistics
metrics := layer.GetMetrics()
topN := layer.GetTopN(v1.MetricMutations, 10)
dist := layer.GetDistribution(v1.MetricMutations)

// Get detected patterns
patterns := layer.GetPatterns(nodeID)
```

### 3. Pattern Detection

The V2 layer automatically detects these behavioral patterns:

| Pattern | Description |
|---------|-------------|
| **Oscillation** | A→B→A→B value oscillation |
| **SameField** | Rapid changes to the same field |
| **CascadeBurst** | Cascading burst updates |
| **LayoutRevert** | Layout immediately reverted |
| **HighFrequency** | High-frequency updates (>60/sec) |
| **Burst** | Burst update pattern |

### 4. Confidence Model

DevTools uses a 5-signal confidence model:

| Signal | Weight | Description |
|--------|--------|-------------|
| Statistical | 0.25 | Data distribution fit |
| Pattern | 0.20 | Pattern match strength |
| **Causal** | **0.30** | Causal link confidence (highest) |
| Context | -0.10 | Context penalty factor |
| Historical | 0.15 | Historical baseline |

## API Reference

### Observation Levels

| Level | Value | Overhead | Features |
|-------|-------|----------|----------|
| LevelNone | 0 | 0% | Completely disabled |
| LevelBasic | 1 | <1% | Counts only |
| LevelEnhanced | 2 | <3% | + TopN, percentiles |
| LevelAdvanced | 3 | <5% | + Pattern detection, insights |

### Key Types

```go
// Node identifiers
type NodeID string
type FrameID int
type MutationID uint64

// Metrics
type MetricsSnapshot struct {
    TotalFrames     uint64
    TotalMutations  uint64
    TotalLayouts    uint64
    TotalRepaints   uint64
}

// Distribution
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

// Pattern
type DetectedPattern struct {
    ID         string
    Type       PatternType
    NodeID     NodeID
    Confidence float64
    Severity   PatternSeverity
    Evidence   []PatternEvidence
}

// Insight
type Insight struct {
    ID          string
    Type        InsightType
    Confidence  float64
    Level       ConfidenceLevel
    Severity    InsightSeverity
    Suggestions []Suggestion
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      DevTools Layer                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐     ┌─────────────────────────────────┐   │
│  │    V1       │     │              V2                 │   │
│  │  Statistics │     │    Pattern Detection            │   │
│  │             │     │    + Confidence Model           │   │
│  │ - Counts    │     │    + Insights Generator         │   │
│  │ - Top N     │     │                                 │   │
│  │ - Percentile│     │  - Oscillation                  │   │
│  │ - Distribution│     │  - Same Field                   │   │
│  └─────────────┘     │  - High Frequency               │   │
│                      │  - Burst                        │   │
│                      └─────────────────────────────────┘   │
│                              │                              │
│                              ▼                              │
│                      ┌─────────────┐                       │
│                      │  Causal    │                       │
│                      │    Chain    │                       │
│                      │  Tracking  │                       │
│                      └─────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

## Use Cases

1. **Debug Performance Issues**
   - Identify components with high mutation rates
   - Detect layout thrashing
   - Find repaint hotspots

2. **Understand Render Behavior**
   - Trace causal chains from events to renders
   - See why components re-render
   - Identify unnecessary updates

3. **Optimization Guidance**
   - Get confidence-based suggestions
   - Pattern-based recommendations
   - Historical comparison

## License

Same as Mint TUI Runtime.
