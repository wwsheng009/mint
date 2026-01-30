# DevTools Demo

A comprehensive example demonstrating how to use the Mint TUI DevTools.

## Overview

This example program demonstrates all major features of the Mint DevTools:

- **Phase 1** - Incremental Basics: Frame recording, event tracking
- **Phase 2** - Causal Chain Engine: Event -> Mutation -> Layout -> Repaint
- **Phase 3** - Time Travel: Navigate and inspect any frame
- **Phase 4** - Deterministic Replay: Replay exact application state
- **Phase 5** - Client Implementation: TUI debug panel, component inspector
- **Phase 6** - Snapshot System: State snapshots, diff, time travel
- **Phase 7** - Memory Optimization: Ring buffers, adaptive sampling, memory monitoring
- **Phase 8** - Cloud Sync: Remote debugging, Chromium DevTools integration

## Building

```bash
cd examples/devtools_demo
go build -o devtools_demo main.go
go build -o panel_demo.exe panel_demo.go
go build -o interactive_panel.exe interactive_panel.go
```

## Running

```bash
# Main demos
./devtools_demo basic       # Basic DevTools usage
./devtools_demo observation # V1 Statistics layer
./devtools_demo patterns    # V2 Pattern detection
./devtools_demo causal      # Causal chain tracking
./devtools_demo insights    # Confidence-based insights
./devtools_demo all         # Run all demos

# Control panel demos
./panel_demo.exe            # API demo (non-interactive)
./interactive_panel.exe     # Full interactive TUI panel
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

### 3. Control Panel

```go
import "github.com/wwsheng009/mint/devtools/client"

// Create control panel
dt := devtools.New()
dt.Enable()

panel := client.NewTuiDebugPanel(dt)
panel.Enable()
panel.SetSize(80, 24)

// Render panel
output := panel.Render()
fmt.Println(output)

// Handle keyboard input
panel.HandleInput('t')  // Toggle timeline
panel.HandleInput('c')  // Toggle causal graph
panel.HandleInput('q')  // Quit

// Component inspection
result := panel.Inspect("button_id")
fmt.Printf("Type: %s, Position: %s\n", result.Type, result.Position)

// Debug overlay
dt.Highlight("button_1", 10, 5, 20, 3)
dt.ClearOverlay()

// Commands
cmdHandler := client.NewCommandHandler(panel)
cmdHandler.Execute("inspect button_id")
cmdHandler.Execute("frame 42")
```

#### Control Panel Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `t` | Toggle Timeline view |
| `c` | Toggle Causal Graph view |
| `s` | Toggle Snapshots view |
| `r` | Toggle Replay view |
| `←` / `→` | Navigate frames |
| `q` | Quit panel |

#### Control Panel Commands

| Command | Description |
|---------|-------------|
| `help` | Show available commands |
| `inspect <node>` | Inspect a component |
| `highlight <node>` | Highlight a component |
| `frame <id>` | Select a frame |
| `stats` | Show statistics |

#### Control Panel Views

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                             DevTools Debug Panel                             │
├──────────────────────────────────────────────────────────────────────────────┤
│ Timeline View                                                                  │
│──────────────────────────────────────────────────────────────────────────────│
│ Current Frame: 9                                                              │
│ Events: 0   Mutations: 0   Layouts: 0   Repaints: 0                            │
├──────────────────────────────────────────────────────────────────────────────┤
│ Causal Graph View                                                              │
│──────────────────────────────────────────────────────────────────────────────│
│ [E]vents → [M]utations → [L]ayouts → [R]epaints                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ [t]imeline [c]ausal [s]napshot [r]eplay [q]uit                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 4. Interactive Panel

The `interactive_panel.exe` provides a fully interactive TUI DevTools panel built with the Mint framework. It features:

- **4 View Tabs**: Timeline, Causal, Stats, Patterns
- **Real-time Updates**: Live frame tracking and statistics
- **Keyboard Navigation**: Arrow keys, Tab, number shortcuts
- **Mouse Support**: Click tabs to switch views

#### Interactive Panel Controls

| Key | Action |
|-----|--------|
| `Tab` | Cycle through tabs |
| `1-4` | Direct tab selection |
| `←` / `→` | Navigate frames (Timeline view) |
| `T` | Timeline tab |
| `C` | Causal tab |
| `S` | Stats tab |
| `P` | Patterns tab |
| `Q` / `Esc` | Quit |

#### Interactive Panel Views

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ DevTools Interactive Panel                                    Frame: 42 | ...│
├──────────────────────────────────────────────────────────────────────────────┤
│ [Timeline] [Causal] [Stats] [Patterns]                                        │
│                                                                              │
│ Timeline View (Last 30 Frames)                                               │
│                                                                              │
│ ████████████████████████████░░░░░░░░░░░░░░░░                                 │
│                                                                              │
│   Selected Frame: 15                                                         │
│   Events: 3                                                                  │
│   Mutations: 2                                                               │
│                                                                              │
│   Use ←/→ to navigate frames                                                 │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ [Tab] Switch  [←/→] Navigate  [Q]uit                                         │
│ Last event: Switched to tab 0                                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 5. Pattern Detection

The V2 layer automatically detects these behavioral patterns:

| Pattern | Description |
|---------|-------------|
| **Oscillation** | A→B→A→B value oscillation |
| **SameField** | Rapid changes to the same field |
| **CascadeBurst** | Cascading burst updates |
| **LayoutRevert** | Layout immediately reverted |
| **HighFrequency** | High-frequency updates (>60/sec) |
| **Burst** | Burst update pattern |

### 6. Confidence Model

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
│                              │                              │
│                              ▼                              │
│                      ┌─────────────┐                       │
│                      │   Panel    │                       │
│                      │   Client   │                       │
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

## Phase 6: Snapshot System

The snapshot system captures complete application state at any frame.

```go
import "github.com/wwsheng009/mint/devtools/snapshot"

// Create snapshot manager
mgr := snapshot.NewManager(1000) // Keep up to 1000 snapshots

// Capture a snapshot
builder := snapshot.NewBuilder("snap-1", devtools.FrameID(42))
builder.SetWindowSize(80, 24)
builder.AddComponent(&snapshot.ComponentState{
    NodeID: "button-1",
    Type: "Button",
    Props: map[string]interface{}{"label": "Click me"},
    State: map[string]interface{}{"clicked": false},
})
snap, _ := mgr.Capture(42, builder)

// Compare snapshots
diff := snapshot.CompareSnapshots(snap1, snap2)
fmt.Println(diff.FormatSummary())

// Time travel through changes
range := snapshot.NewTimeTravelRange(mgr.GetAll())
range.Compute()
history := range.GetChangeHistory("button-1")
```

### Snapshot Features

| Feature | Description |
|---------|-------------|
| **Full State Capture** | Component props, state, bounds, styles |
| **Incremental Diff** | Efficient change detection between frames |
| **Time Travel** | Navigate entire change history |
| **Persistence** | Save/load snapshots to disk |

## Phase 7: Memory Optimization

Memory optimization features keep overhead minimal.

```go
import "github.com/wwsheng009/mint/devtools/memory"

// Ring buffer for circular storage
ring := memory.NewRingBuffer(1000)
ring.Write(devtools.FrameID(1))
ring.Write(devtools.FrameID(2))
frames := ring.GetAll() // [1, 2]

// Adaptive sampling strategy
strategy := memory.NewAdaptiveStrategy(0.1, 1.0) // 10%-100%
shouldSample := strategy.ShouldSample(frameID, context)

// Memory monitoring
monitor := memory.NewMonitor()
monitor.SetSampleInterval(5 * time.Second)
monitor.SetAlertCallback(func(alert memory.MemoryAlert) {
    fmt.Printf("Alert: %s\n", alert.Message)
})
monitor.Start()
```

### Memory Features

| Feature | Description |
|---------|-------------|
| **Ring Buffer** | Circular storage with O(1) operations |
| **Adaptive Sampling** | Auto-adjusts sampling based on memory pressure |
| **Memory Monitor** | Tracks usage, triggers alerts |
| **Frame Window** | Sliding window for recent frames |

## Phase 8: Remote Debugging

Connect your TUI app to browser DevTools for remote inspection.

```go
import "github.com/wwsheng009/mint/devtools/remote"

// Create bridge
bridge := remote.NewChromiumBridge(devtools, snapshotMgr)
bridge.Enable()

// Get inspector URL
fmt.Println("Inspect at:", bridge.GetInspectURL())

// Handle incoming messages
server.SetMessageHandler(func(session *remote.Session, msg *remote.Message) *remote.Message {
    switch msg.Type {
    case "get_snapshot":
        // Return snapshot data
    case "get_diff":
        // Return diff data
    }
    return nil
})
```

### Remote Features

| Feature | Description |
|---------|-------------|
| **WebSocket Protocol** | JSON-based messaging |
| **Chromium Bridge** | Compatible with Chrome DevTools |
| **Inspector UI** | HTML-based inspection page |
| **Breakpoints** | Set conditional breakpoints |
| **Evaluation** | Evaluate expressions in context |

### Protocol Messages

| Type | Direction | Description |
|------|-----------|-------------|
| `handshake` | Client→Server | Initial connection |
| `get_snapshot` | Client→Server | Get frame snapshot |
| `get_diff` | Client→Server | Compare two frames |
| `set_breakpoint` | Client→Server | Set breakpoint |
| `event` | Server→Client | Push events to client |
