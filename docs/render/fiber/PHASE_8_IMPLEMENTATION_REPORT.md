# Phase 8 Implementation Report: Fiber to Layout Engine NodeID Propagation

**Date**: 2026-02-13
**Status**: Complete
**Author**: wwsheng009

## Overview

Phase 8 completes the Fiber to Layout Engine NodeID propagation chain, ensuring that NodeIDs assigned during reconciliation are properly propagated through the entire rendering pipeline to ComputedBox nodes, enabling stable runtime identity independent of VNode keys and paths.

## Implementation Principle

> **运行时身份只认 NodeID，不认 key，不认 path**
> Runtime identity only recognizes NodeID, not key, not path

The core identity principle requires that all runtime components use NodeID as the sole stable identifier for component identity.

## Complete Propagation Chain

```
Reconciler (Fiber Tree Generation)
  │
  ├─ CommitRoot() assigns NodeID during reconciliation
  │
  └─ GetFiberRoot() → SetFiber() on renderer
       │
       └─ PipelineRenderer.fiber (stored reference)
            │
            ├─ Render() / RenderWithConstraints() / RenderWithFiber()
            │    └─ pipeline.Render(vnode, r.fiber, constraints, buf)
            │         │
            │         └─ RenderingPipeline
            │              ├─ Render() → engine.Layout(vnode, fiber, constraints)
            │              └─ RenderLayers() → layerMgr.CollectAndLayout(vnode, fiber, constraints, engine)
            │                   │
            │                   └─ LayerManager
            │                        ├─ CollectAndLayout()
            │                        │    └─ engine.Layout(baseTree, fiber, constraints)
            │                        │
            │                        └─ layoutLayer(node, layer, constraints, engine, fiber)
            │                             └─ engine.Layout(node.Content, fiber, layerConstraints)
            │                                  │
            │                                  └─ ComputeEngine.Layout()
            │                                       └─ box.NodeID = fiber.NodeID
            │                                            │
            │                                            └─ ComputedBox.NodeID
            │                                                 │
            │                                                 └─ HitMap.Entry.NodeID
            │                                                      │
            │                                                       └─ Event Routing (NodeID priority)
```

## Files Modified

### 1. internal/render/rendering_pipeline.go

**Updated methods:**

- `Render(vnode, fiber, constraints, buffer)` - Added fiber parameter
- `RenderLayers(vnode, fiber, constraints, buffer)` - Added fiber parameter
- `RenderWithFiber(vnode, fiber, constraints, buffer)` - Explicit fiber rendering
- `ComputeLayout(vnode, fiber, constraints)` - Added fiber parameter
- `HasModalChecks(vnode, fiber, constraints)` - Added fiber parameter

**Key changes:**
```go
// Before: Pass nil for Fiber (non-Fiber mode, backward compatible)
layout, err := p.layoutEngine.Layout(vnode, nil, constraints)

// After: Pass Fiber for NodeID propagation
layout, err := p.layoutEngine.Layout(vnode, fiber, constraints)
```

### 2. runtime/layer/manager.go

**Updated methods:**

- `CollectAndLayout(vnode, fiber, constraints, engine)` - Added fiber parameter
- `layoutLayer(node, layer, constraints, engine, fiber)` - Added fiber parameter

**Key changes:**
```go
// Before:
layerLayout, err := m.layoutLayer(node, layer, constraints, engine)
// layoutLayer passed nil to engine.Layout

// After:
layerLayout, err := m.layoutLayer(node, layer, constraints, engine, fiber)
// layoutLayer passes fiber to engine.Layout
```

### 3. internal/render/pipeline_renderer.go

**Updated methods:**

- `Measure(vnode, maxWidth, maxHeight)` - Now uses `r.fiber` for Layout
- `Render()` - Passes `r.fiber` to pipeline
- `RenderWithConstraints()` - Passes `r.fiber` to pipeline
- `RenderWithFiber()` - Explicit fiber-aware rendering

**Key changes:**
```go
// Before:
layout, err := r.pipeline.GetLayoutEngine().Layout(vnode, nil, constraints)

// After:
layout, err := r.pipeline.GetLayoutEngine().Layout(vnode, r.fiber, constraints)
```

### 4. internal/render/vnode_renderer.go

**Updated methods:**

- `SetFiber(fiber *reconciler.Fiber)` - Sets fiber on PipelineRenderer

**Key changes:**
```go
func (r *PipelineRendererAdapter) SetFiber(fiber *reconciler.Fiber) {
	r.pipeline.fiber = fiber
}
```

### 5. internal/reconciler/reconciler.go

**Updated fields:**

- Added `renderer rtui.VNodeRenderer` field

**Updated methods:**

- `SetRenderer(renderer rtui.VNodeRenderer)` - Sets renderer for SetFiber call
- `CommitRoot()` - Calls SetFiber after reconciliation

**Key changes in CommitRoot():**
```go
// Phase 8: Set Fiber on renderer for NodeID propagation before layout
if r.renderer != nil {
	if adapter, ok := r.renderer.(interface{ SetFiber(*Fiber) }); ok {
		adapter.SetFiber(r.root)
	}
}
```

### 6. internal/reconciler/get_fiber_root.go

**New file:**

- `GetFiberRootForRendering()` - Returns Fiber root for rendering pipeline

### 7. internal/render/declarative_node.go

**Updated methods:**

- `NewDeclarativeNodeFromFuncWithFiber()` - Calls `adapter.SetRenderer(renderer)`
- Added `fiberReconcilerAdapter.SetRenderer()` method
- Added `GetFiberRoot() *reconciler.Fiber` method
- Added `RenderWithFiber(buffer *paint.Buffer)` method
- `SetReconciler()` - Also calls `SetRenderer()` when reconciler is set

**Key changes:**
```go
// In NewDeclarativeNodeFromFuncWithFiber():
renderer := NewPipelineRendererAdapter()

// Phase 8: Set renderer on reconciler for NodeID propagation
if adapter, ok := r.(*fiberReconcilerAdapter); ok {
	adapter.SetRenderer(renderer)
}

// In fiberReconcilerAdapter:
func (a *fiberReconcilerAdapter) SetRenderer(renderer rtui.VNodeRenderer) {
	a.r.SetRenderer(renderer)
}
```

## Integration Point

The framework integration requires two connections:

1. **Reconciler to Renderer** - In `NewDeclarativeNodeFromFuncWithFiber()`:
```go
renderer := NewPipelineRendererAdapter()
if adapter, ok := r.(*fiberReconcilerAdapter); ok {
	adapter.SetRenderer(renderer)
}
```

2. **Reconciler to Renderer (Fiber propagation)** - In `Reconciler.CommitRoot()`:
```go
if r.renderer != nil {
	if adapter, ok := r.renderer.(interface{ SetFiber(*Fiber) }); ok {
		adapter.SetFiber(r.root)
	}
}
```

This ensures the rendering pipeline has access to the Fiber tree before layout computation begins.

## Backward Compatibility

The following methods intentionally pass `nil` for fiber for backward compatibility or test isolation:

1. **ComputeLayout()** - Was updated to accept fiber, but callers may pass nil
2. **Measure()** - Now uses `r.fiber` from PipelineRenderer field
3. **Test files** - Intentionally use nil for test isolation
4. **Diagnostic tools** - Inspector and layout debug tools run without reconciliation

## Testing

Run the following to verify the implementation:

```bash
go test ./internal/reconciler/... -run TestHitMap
go test ./internal/render/... -run TestRender
go test ./runtime/layer/... -run TestModal
go test ./runtime/event/... -run TestNodeID
```

## Verification Checklist

- [x] `PipelineRenderer` has `fiber *reconciler.Fiber` field
- [x] `SetFiber()` method on `PipelineRendererAdapter`
- [x] `Reconciler.SetRenderer()` for renderer registration
- [x] `Reconciler.CommitRoot()` calls `SetFiber()` after reconciliation
- [x] `fiberReconcilerAdapter.SetRenderer()` adapter method
- [x] `NewDeclarativeNodeFromFuncWithFiber()` calls `SetRenderer()`
- [x] `RenderingPipeline.Render()` passes fiber to LayoutEngine
- [x] `RenderingPipeline.RenderLayers()` passes fiber to LayerManager
- [x] `LayerManager.CollectAndLayout()` passes fiber to LayoutEngine
- [x] `LayerManager.layoutLayer()` passes fiber to LayoutEngine
- [x] `ComputeEngine.Layout()` sets `box.NodeID = fiber.NodeID`
- [x] `PipelineRenderer.Measure()` uses `r.fiber` for Layout
- [x] `ComputeLayout()` accepts fiber parameter
- [x] `HasModalChecks()` accepts fiber parameter

## Related Documentation

- [IDENTITY_REFACTORING_PLAN.md](IDENTITY_REFACTORING_PLAN.md) - Overall refactoring plan
- [Phase 1-7](../../plan/fiber/IDENTITY_REFACTORING_PLAN.md) - Previous phases
- [InstanceRegistry](../../../inspector/identity/INSTANCE_REGISTRY.md) - Instance management with NodeID

## Next Steps

Phase 8 completes the core infrastructure. Future enhancements may include:

1. Enhanced testing for Fiber tree integrity
2. Performance optimization for large Fiber trees
3. Tooling for Fiber tree debugging and inspection
4. Documentation updates for Fiber-based component development
