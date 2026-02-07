# Modal Centering Issue - Root Cause and Solution

## Problem Statement

After fixing the HStack flex layout issue, modal dialogs were no longer centered in the viewport.

## Root Cause Analysis

### The Conflict

Two requirements were in conflict:
1. **HStack flex layout** needs bounded `BoxConstraints` to divide available space
2. **Modal centering** needs the layer manager's `centerModal()` to be called

### Why the First Fix Broke Modal Centering

My initial fix called `RenderingPipeline.Render()` directly:

```go
// ❌ WRONG - bypasses layer system
constraints := runtime.NewBoxConstraints(0, buf.Width, 0, buf.Height)
adapter.GetRenderingPipeline().Render(n.root, constraints, buf)
```

**Problem**: `RenderingPipeline.Render()` does:
```go
// rendering_pipeline.go:54
layout, err := p.layoutEngine.Layout(vnode, constraints)
```

It calls `layoutEngine.Layout()` **directly**, bypassing the layer manager entirely. This means:
- ✅ Flex layout works (constraints are passed)
- ❌ Modal centering doesn't work (`centerModal()` is never called)

### Why `PipelineRenderer.Render()` Works for Both

The correct approach uses `PipelineRenderer.Render()`:

```go
// ✅ CORRECT - uses layer system
adapter.GetPipeline().Render(n.root, 0, 0, buf)
```

**What it does**:
1. Extracts buffer from the interface (pipeline_renderer.go:43-51)
2. Creates constraints from buffer dimensions (line 58-60):
   ```go
   width := buf.Width
   height := buf.Height
   constraints := runtime.NewBoxConstraints(0, width, 0, height)
   ```
3. Detects layer nodes via `hasLayerNodes()` (line 63)
4. If layers exist, calls `pipeline.RenderLayers()` (line 75):
   ```go
   err = r.pipeline.RenderLayers(vnode, constraints, buf)
   ```
5. `RenderLayers()` triggers layer manager's `CollectAndLayout()` (manager.go:47)
6. `CollectAndLayout()` calls `centerModal()` for LayerModal (manager.go:152-153)

## Modal Centering Implementation

### Layer Manager's Centering Logic

Located in `runtime/layer/manager.go:159-207`:

```go
func (m *Manager) centerModal(root *compute.ComputedBox, constraints runtime.BoxConstraints) {
    // Get modal dimensions
    modalWidth := root.Box.Width
    modalHeight := root.Box.Height

    // Get container dimensions from constraints
    containerWidth := constraints.MaxWidth
    containerHeight := constraints.MaxHeight

    // Handle infinite dimensions (shouldn't happen with our fix)
    if containerWidth == runtime.Infinity {
        containerWidth = modalWidth
    }
    if containerHeight == runtime.Infinity {
        containerHeight = modalHeight
    }

    // Calculate centering offset
    offsetX := (containerWidth - modalWidth) / 2
    offsetY := (containerHeight - modalHeight) / 2

    // Shift entire modal tree by offset
    m.shiftPositions(root, offsetX, offsetY)
}
```

### How It Works

1. Modal is initially laid out at position (0, 0)
2. `centerModal()` calculates the offset to center it
3. `shiftPositions()` recursively shifts all modal boxes by the offset
4. Paint engine uses the shifted positions to render

## Comparison Table

| Approach | Flex Layout | Modal Centering | Why |
|----------|-------------|-----------------|-----|
| Original `n.renderer.Render()` | ❌ | ✅ | Uses legacy x,y positioning |
| `RenderingPipeline.Render()` | ✅ | ❌ | Bypasses layer manager |
| `PipelineRenderer.Render()` | ✅ | ✅ | Detects layers + uses constraints |

## Code Flow Diagram

```
DeclarativeNode.Paint()
    │
    ├─> adapter.GetPipeline().Render(n.root, 0, 0, buf)
    │       │
    │       └─> PipelineRenderer.Render(vnode, x, y, buf)
    │               │
    │               ├─> buf := extract buffer
    │               ├─> constraints := NewBoxConstraints(0, buf.Width, 0, buf.Height)
    │               ├─> hasLayers := hasLayerNodes(vnode)
    │               │
    │               └─> if hasLayers:
    │                       └─> pipeline.RenderLayers(vnode, constraints, buf)
    │                               │
    │                               └─> layerManager.CollectAndLayout(...)
    │                                       │
    │                                       ├─> For each layer:
    │                                       │       └─> layoutEngine.Layout(...)
    │                                       │
    │                                       └─> if layer == LayerModal:
    │                                               └─> centerModal(root, constraints)
    │                                                       │
    │                                                       └─> shiftPositions(root, offsetX, offsetY)
    │
    └─> Paint engine uses computed layout with centered modal
```

## Testing

To verify both features work correctly:

```bash
cd examples/ui_demos/demo1_full_featured
go run main.go
```

**Expected behavior**:
1. Main content panels should fill screen width (no gaps on right)
2. Press button to open modal
3. Modal should be centered both horizontally and vertically

## Key Takeaways

1. **Always use `PipelineRenderer.Render()`** for VNode rendering
   - It handles both constraint-based layout and layer-specific logic
   - Never bypass it by calling `RenderingPipeline.Render()` directly

2. **Layer system is critical for modals/overlays/tooltips**
   - These components need special handling (centering, event blocking, z-order)
   - The layer manager provides this functionality

3. **Buffer dimensions are the source of truth for constraints**
   - `PipelineRenderer` creates constraints from buffer dimensions
   - This ensures layout respects the actual terminal size

## Files Modified

- `internal/render/declarative_node.go`: Updated to use `PipelineRenderer.Render()`
- No other files needed modification (the layer system was already correct)
