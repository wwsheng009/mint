# Parallel Rendering Pipeline

## Overview

This document describes the parallel rendering path implementation that allows switching between `runtime/compute` (stable) and `runtime/layout` (experimental) layout engines.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  ParallelRenderingPipeline                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                  LayoutSwitcher                       │    │
│  │  ┌─────────────────┐  ┌─────────────────────────┐   │    │
│  │  │ ComputeEngine   │  │ NewLayoutEngine         │   │    │
│  │  │ Adapter         │  │ Adapter                 │   │    │
│  │  │ (stable)        │  │ (experimental)          │   │    │
│  │  └────────┬────────┘  └───────────┬─────────────┘   │    │
│  │           │                        │                  │    │
│  │           └──────────┬─────────────┘                  │    │
│  │                      │                                │    │
│  │           ┌──────────▼──────────┐                    │    │
│  │           │ LayoutResult        │                    │    │
│  │           │ (unified interface) │                    │    │
│  │           └──────────┬──────────┘                    │    │
│  └──────────────────────┼───────────────────────────────┘    │
│                         │                                    │
│           ┌─────────────▼─────────────┐                     │
│           │       PaintEngine         │                     │
│           │   (shared paint logic)    │                     │
│           └───────────────────────────┘                     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. LayoutEngineType

Defines which layout engine to use:

```go
const (
    LayoutEngineCompute LayoutEngineType = iota  // stable, production
    LayoutEngineNew                               // experimental
    LayoutEngineBoth                              // run both in parallel
)
```

### 2. LayoutSwitcher

Manages switching between layout engines:

```go
switcher := NewLayoutSwitcher()

// Switch engines
switcher.SetEngineType(LayoutEngineNew)

// Enable comparison (for Both mode)
switcher.SetCompareResults(true)
switcher.SetTolerance(1.0) // 1% tolerance for size differences
```

### 3. ParallelRenderingPipeline

Main pipeline with switch capability:

```go
pipeline := NewParallelRenderingPipeline()

// Use compute engine (default)
pipeline.SetLayoutEngineType(LayoutEngineCompute)

// Use new layout engine
pipeline.SetLayoutEngineType(LayoutEngineNew)

// Run both engines in parallel for comparison
pipeline.SetLayoutEngineType(LayoutEngineBoth)
pipeline.SetCompareResults(true)

// Render
err := pipeline.Render(vnode, fiber, constraints, buffer)
```

### 4. Fiber Adapter

Adapts Fiber/VNode trees to `layout.Node` interface:

```go
// Adapt Fiber tree
adapter := NewFiberToNodeAdapter(fiber, vnode)

// Now can be used with runtime/layout
result := layoutEngine.Layout(adapter, constraints)
```

## Usage

### Environment Variable

Set the layout engine via environment variable:

```bash
# Use compute engine (default)
export MINT_LAYOUT_ENGINE=compute

# Use new layout engine
export MINT_LAYOUT_ENGINE=layout

# Run both engines
export MINT_LAYOUT_ENGINE=both
```

### Programmatic Usage

```go
package main

import (
    "github.com/wwsheng009/mint/internal/render"
    "github.com/wwsheng009/mint/runtime"
)

func main() {
    // Create pipeline
    pipeline := render.NewParallelRenderingPipeline()
    
    // Switch to experimental engine
    pipeline.SetLayoutEngineType(render.LayoutEngineNew)
    
    // Enable debug output
    pipeline.SetLayoutDebug(true)
    
    // Render
    constraints := runtime.NewBoxConstraints(0, 80, 0, 24)
    err := pipeline.Render(vnode, fiber, constraints, buffer)
    if err != nil {
        panic(err)
    }
    
    // Get statistics
    computeStats, newStats, switcherStats := pipeline.GetStats()
    fmt.Printf("Compute: hits=%d, misses=%d\n", computeStats.Hits, computeStats.Misses)
    fmt.Printf("New: hits=%d, misses=%d\n", newStats.Hits, newStats.Misses)
    fmt.Printf("Total renders: %d, Differences: %d\n", 
        switcherStats.TotalRenders, switcherStats.Differences)
}
```

### Benchmark Mode

Compare performance between engines:

```go
pipeline := NewParallelRenderingPipeline()

// Benchmark both engines
computeResult, newResult := pipeline.BenchmarkBoth(vnode, fiber, constraints, buffer)

fmt.Printf("Compute: layout=%v, paint=%v, total=%v\n",
    computeResult.LayoutDuration,
    computeResult.PaintDuration,
    computeResult.TotalDuration)

fmt.Printf("New: layout=%v, total=%v\n",
    newResult.LayoutDuration,
    newResult.TotalDuration)
```

## Feature Comparison

| Feature | runtime/compute | runtime/layout |
|---------|-----------------|----------------|
| Flexbox Layout | ✅ | ✅ |
| Flex Grow | ✅ | ✅ |
| Flex Shrink | ✅ | ⚠️ (defined but not calculated) |
| Gap | ✅ | ✅ |
| Padding | ✅ | ✅ |
| Margin | ✅ | ❌ (TODO) |
| Absolute Position | ✅ | ❌ (TODO) |
| Border Container | ✅ | ❌ (TODO) |
| Table Layout | ✅ | ❌ (TODO) |
| Multi-layer | ✅ | ❌ (TODO) |
| Cache | ✅ | ✅ |
| Dirty Tracking | ✅ | ✅ |
| HitMap | ✅ | ✅ |
| Validation | ✅ | ✅ |

## Migration Path

### Phase 1: Parallel Testing
- Use `LayoutEngineBoth` mode
- Enable comparison logging
- Monitor differences

### Phase 2: Feature Parity
- Implement missing features in `runtime/layout`
- Run comparison tests
- Fix discrepancies

### Phase 3: Gradual Rollout
- Enable `LayoutEngineNew` for specific components
- Monitor performance and correctness
- Expand coverage

### Phase 4: Full Migration
- Switch default to `LayoutEngineNew`
- Deprecate `LayoutEngineCompute`

## Files

| File | Purpose |
|------|---------|
| `layout_switcher.go` | Core switching logic, adapters, pipeline |
| `fiber_adapter.go` | Fiber/VNode to layout.Node conversion |
| `layout_switcher_test.go` | Unit tests |

## Statistics Tracking

```go
type SwitcherStats struct {
    TotalRenders     int64  // Total render calls
    ComputeRenders   int64  // Renders using compute engine
    NewEngineRenders int64  // Renders using new engine
    BothRenders      int64  // Renders using both engines
    Differences      int64  // Number of layout differences found
    Errors           int64  // Number of errors encountered
}
```

Access via:
```go
_, _, stats := pipeline.GetStats()
```
