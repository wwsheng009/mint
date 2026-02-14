# Phase 4.5: Transition State Documentation

## Overview

**Status**: Phase 4 is in a transition state between old and new architecture
**Date**: 2026-02-14
**Context**: Fiber Unified Architecture Refactor

---

## Current Implementation Status

### Completed Phases (1-3)

| Phase | Name | Status | Key Deliverables |
|-------|------|--------|------------------|
| Phase 1 | Infrastructure | ✅ Complete | Fiber.Layer, ComputedBox fields added |
| Phase 2 | Layout Refactor | ✅ Complete | layoutFiber(), measureFiber() implemented |
| Phase 3 | RenderPlane Introduction | ✅ Complete | RenderPlanes implemented and integrated |
| Phase 3.x | Test Coverage | ✅ Complete | Unit/integration/benchmark tests |

### Phase 4: Deprecation (In Transition)

**Goal**: Mark old APIs as Deprecated but keep them functional until transforms move

#### Deprecated APIs

1. **Collector.StripLayers()** (`runtime/layer/collector.go:212-229`)
   - Status: Marked Deprecated
   - Reason: Creates VNode clones, violates single data source principle
   - Still used by: `CollectAndLayout()` (line 75)

2. **Collector.cloneWithoutLayers()** (`runtime/layer/collector.go:232-301`)
   - Status: Marked Deprecated
   - Reason: Internal helper for StripLayers
   - Still used by: `StripLayers()`

3. **LayerManager.CollectAndLayout()** (`runtime/layer/manager.go:44-122`)
   - Status: Marked Deprecated
   - Reason: Old layer-based layout with transforms
   - Still used by: `RenderingPipeline.RenderLayers()` (lines 223, 283)

---

## Phase 4.5 Architecture: Transition State

### Why Can't We Remove CollectAndLayout Yet?

The old `CollectAndLayout()` API is still needed because it provides **critical post-layout transforms**:

1. **Modal Centering** (`runtime/layer/manager.go:189-192```go
// Post-process layout for modal (center it)
if layer == rtui.LayerModal && layout.Root != nil {
    m.centerModal(layout.Root, constraints)
}
```

   - `centerModal()` (lines 216-258): Computes centering offset and shifts modal position
   - Happens AFTER `engine.Layout()` computes initial positions
   - Happens BEFORE building HitMap (to capture final positions)

2. **Inspector Positioning** (`runtime/layer/manager.go:194-197```go
// Post-process layout for inspector (position it at specified coordinates)
if layer == rtui.LayerInspector && layout.Root != nil {
    m.positionInspector(node, layout.Root)
}
```

   - `positionInspector()` (lines 282-329): Reads `x, y` props and shifts inspector position
   - Happens AFTER `engine.Layout()` computes initial positions
   - Happens BEFORE building HitMap (to capture final positions)

3. **HitMap Rebuild** (`runtime/layer/manager.go:199-211```go
// IMPORTANT: Rebuild HitMap AFTER post-processing (centering, etc.)
// The HitMap built in Engine.Layout() was before centering, so it has wrong positions
if layout.Root != nil {
    layout.HitMap = m.buildHitMapFromComputedBox(layout.Root)
}
```

   - Rebuilds HitMap with FINAL positions after transforms
   - Ensures event hit testing uses correct coordinates

### Current RenderLayers Architecture

```
RenderingPipeline.RenderLayers()
  │
  ├─→ layerMgr.CollectAndLayout(vnode, fiber, constraints, engine)
  │     │
  │     ├─→ collector.Collect(vnode)              // Extract layer nodes
  │     ├─→ baseTree := collector.StripLayers(vnode) // ❌ Old API
  │     ├─→ engine.Layout(baseTree, fiber, constraints)
  │     ├─→ for layer, nodes := range collector.GetLayers() {
  │     │     layoutLayer(node, layer, constraints, engine, fiber)
  │     │        │
  │     │        ├─→ engine.Layout(node.Content, fiber, layerConstraints)
  │     │        ├─→ if layer == LayerModal { centerModal(...) }     // ✅ TRANSFORM
  │     │        ├─→ if layer == LayerInspector { positionInspector(...) } // ✅ TRANSFORM
  │     │        └─→ layout.HitMap = buildHitMapFromComputedBox(...) // ✅ REBUILD
  │     │     }
  │     ├─→ m.renderPlanes = BuildRenderPlanesFromLayouts(m.layouts)
  │     │                                                │
  │     │                                                └─► Builds RenderPlanes from FINAL positions
  │     │                                                   (after centering, positioning)
  │     └─→ return nil
  │
  ├─→ renderPlanes := layerMgr.BuildRenderPlanes(fiber)
  │     └─► NEW API (for Phase 5+)
  │
  ├─→ p.paintEngine.PaintRenderPlanes(renderPlanes, buffer)
  │     │
  │     └─► Paint from RenderPlanes (NEW architecture)
  │
  └─→ p.lastHitMap = event.BuildHitMapFromFiber(fiber)
      │
      └─► Build HitMap from Fiber (NEW architecture)
```

---

## Key Insight: Transition Strategy

### "Build RenderPlanes From Layouts" Pattern

The current implementation uses a clever transition pattern:

```go
// Phase 4.5: BuildRenderPlanesFromLayouts captures final transformed positions
m.renderPlanes = BuildRenderPlanesFromLayouts(m.layouts)
```

**How it works:**
1. Call `CollectAndLayout()` (old API) to get transforms applied
2. Build `RenderPlanes` from the **transformed** layouts (new data structure)
3. Paint from `RenderPlanes` (new painting API)
4. Build `HitMap` from `Fiber` (new HitMap API)

**Why this works:**
- `RenderPlanes` becomes the **source of truth** for rendering
- Old `CollectAndLayout()` becomes a **helper** for applying transforms
- When transforms move to Layout Engine, we can:
  - Call `engine.Layout()` directly on Fiber
  - Build `RenderPlanes` from `Fiber` (via `BuildFromFiber()`)
  - Remove `CollectAndLayout()` entirely

---

## Next Steps: Phase 5 - Move Transforms to Layout Engine

### Current Transform Locations

| Transform | Current Location | Target Location (Phase 5) |
|-----------|------------------|---------------------------|
| Modal Centering | `LayerManager.centerModal()` | `Engine.layoutFiber()` or `Engine.applyTransforms()` |
| Inspector Positioning | `LayerManager.positionInspector()` | `Engine.layoutFiber()` or `Engine.applyTransforms()` |
| HitMap Rebuild | `LayerManager.buildHitMapFromComputedBox()` | `Engine.buildHitMapFromFiber()` |

### Phase 5 Implementation Options

#### Option A: Apply Transforms Inside layoutFiber()

**Approach**: Apply transforms during `layoutFiber()` for each node

```go
func (e *Engine) layoutFiber(
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
    depth int,
) *ComputedBox {
    // 1. Create/Compute box
    box := e.buildComputedBox(fiber, ...)

    // 2. Recursively layout children
    for _, childFiber := range fiber.Children {
        childBox := e.layoutFiber(childFiber, childConstraints, depth+1)
        box.Children = append(box.Children, childBox)
    }

    // 3. Apply transforms based on Layer
    if fiber.Layer == rtui.LayerModal {
        e.centerModal(box, constraints)
    } else if fiber.Layer == rtui.LayerInspector {
        e.positionInspector(fiber, box)
    }

    return box
}
```

**Pros:**
- Clean: Transforms happen during layout, in the Layout Engine
- Single-pass: No need for post-processing
- No duplicate HitMap: Use positions from `layoutFiber()`

**Cons:**
- Layout Engine needs to know about layer-specific transforms
- Increases Layout Engine responsibilities

#### Option B: Apply Transforms After Layout, Before RenderPlanes

**Approach**: Keep `Engine.Layout()` clean, apply transforms in `RenderingPipeline`

```go
func (p *RenderingPipeline) RenderLayers(...) error {
    // 1. Layout entire Fiber tree (no layer separation)
    layout := p.layoutEngine.Layout(vnode, fiber, constraints)

    // 2. Apply layer transforms (move from LayerManager to RenderingPipeline)
    p.applyLayerTransforms(fiber, constraints)

    // 3. Build RenderPlanes from Fiber (with final positions)
    renderPlanes := layer.BuildFromFiber(fiber)

    // 4. Paint
    p.paintEngine.PaintRenderPlanes(renderPlanes, buffer)

    // 5. Build HitMap
    p.lastHitMap = event.BuildHitMapFromFiber(fiber)
}
```

**Pros:**
- Layout Engine stays clean (pure layout)
- Transforms are clearly separated from layout calculation
- Easy to test transforms independently

**Cons:**
- Two-pass: Layout → Transform → RenderPlanes
- Need to walk Fiber tree again for transforms

### Recommended Approach: Option A

**Rationale:**
1. Matches the design document: "Phase 5: Layout 基于 Fiber，Layer 作为渲染维度"
2. Layout Engine is the right place to compute final positions
3. Aligns with "Single Fiber Tree" principle: All transforms happen on Fiber

---

## Phase 5 Task Breakdown

### 5.1 Prepare Layout Engine

**File**: `runtime/compute/engine.go`

**Tasks:**
- [ ] Add `centerModal(box *ComputedBox, constraints BoxConstraints)` method
- [ ] Add `positionInspector(fiber *Fiber, box *ComputedBox)` method
- [ ] Update `layoutFiber()` to call these transforms
- [ ] Add tests for modal centering in Layout Engine
- [ ] Add tests for inspector positioning in Layout Engine

### 5.2 Update RenderingPipeline

**File**: `internal/render/rendering_pipeline.go`

**Tasks:**
- [ ] Modify `RenderLayers()` to:
  - Call `layoutEngine.Layout()` directly (no `CollectAndLayout`)
  - Get renderLayers from Fiber via `BuildFromFiber()`
  - No calling `centerModal` or `positionInspector` directly
- [ ] Update comment: "Phase 5: Using new Layout Engine with layer transforms"
- [ ] Keep old `CollectAndLayout()` for backward compatibility (can be removed in Phase 7)

### 5.3 Update LayerManager

**File**: `runtime/layer/manager.go`

**Tasks:**
- [ ] Keep `centerModal()` and `positionInspector()` but mark them Deprecated
- [ ] Update comment: "Moved to Layout Engine in Phase 5"
- [ ] Keep `CollectAndLayout()` for backward compatibility
- [ ] Add migration note in comments

### 5.4 Tests

**Tasks:**
- [ ] Test modal centering with new Layout Engine
- [ ] Test inspector positioning with new Layout Engine
- [ ] Test RenderLayers with direct Layout Engine call
- [ ] Run all existing tests to ensure no regressions
- [ ] Add integration test for full flow with modals/inspectors

### 5.5 Documentation

**Tasks:**
- [ ] Update `docs/plan/fiber/TODO_LIST.md` - mark Phase 4 complete, Phase 5 in progress
- [ ] Update AGENTS.md with new architecture
- [ ] Create migration guide from old to new API
- [ ] Update commit message template for Phase 5

---

## Verification Checklist

Before marking Phase 5 complete:

- [ ] All RenderPlanes tests pass
- [ ] Modal centering works correctly (visual check)
- [ ] Inspector positioning works correctly (visual check)
- [ ] HitMap positions match actual rendered positions
- [ ] Event hit testing works correctly
- [ ] No deprecation warnings for new code paths
- [ ] Old `CollectAndLayout()` still works (backward compatibility)
- [ ] Performance: No measurable degradation vs Phase 4

---

## Open Questions

1. **Layer Constraints**: Should each layer have different constraints during layout?
   - Current: Layout Manager applies different constraints per layer
   - Future: Layout Engine might need a `LayerConstraints` map

2. **HitMap Timing**: When should HitMap be built?
   - Option A: During `layoutFiber()` (add entries as we go)
   - Option B: After all transforms complete (walk Fiber tree once)
   - Decision: Option B (simpler, single pass)

3. **Backward Compatibility**: How long to keep deprecated APIs?
   - Phase 5-6: Keep for safety
   - Phase 7: Remove deprecated APIs

---

## Related Documents

- [Main Refactor Plan](./UNIFIED_FIBER_ARCHITECTURE_REFACTOR.md)
- [Phase TODO List](./TODO_LIST.md)
- [Phase Checklist](./CHECK_LIST.md)
- [Test Strategy](./TESTING_VERIFICATION.md)

---

**Last Updated**: 2026-02-14
**Next Review**: After Phase 5 completion
