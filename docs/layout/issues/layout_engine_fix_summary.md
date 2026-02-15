# Layout Engine Fix - HStack Width Constraint Issue

## Problem

The demo1_full_featured application had a layout issue where the HStack panels in the main body didn't fill the screen width properly. Each panel was 44 characters wide (total 88), but the screen was only 80-81 characters wide, causing the panels to extend beyond the screen boundary.

### Root Cause

The `DeclarativeNode.Paint` method was calling the old `VNodeRenderer.Render(vnode, x, y, buffer)` interface which only passed x, y coordinates to the renderer, NOT the width/height constraints.

While the new `RenderingPipeline` supports BoxConstraints for proper flex layout, the old interface path bypassed this functionality.

## Solution

Modified three files to use the constraint-based rendering pipeline:

### 1. `internal/render/declarative_node.go`

Updated the `Paint` method to:
- Detect when the renderer is a `PipelineRendererAdapter`
- Create proper `BoxConstraints` from the buffer dimensions
- Call `RenderingPipeline.Render(vnode, constraints, buf)` directly with constraints
- Fallback to legacy rendering if the pipeline fails

### 2. `internal/render/pipeline_renderer.go`

Added `GetRenderingPipeline()` method to expose the inner `RenderingPipeline` for constraint-based rendering.

### 3. `internal/render/vnode_renderer.go`

Added two methods to `PipelineRendererAdapter`:
- `GetPipeline()` - returns the outer `PipelineRenderer`
- `GetRenderingPipeline()` - returns the inner `RenderingPipeline` for constraint-based rendering

## Technical Details

### Before (Old Path)

```go
// Old interface - only x, y coordinates, no constraints
n.renderer.Render(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
```

This caused the layout engine to measure HStack children with **unbounded width constraints**, so flex children used their natural sizes instead of dividing the available space.

### After (New Path)

```go
// New interface - proper BoxConstraints from buffer dimensions
constraints := runtime.NewBoxConstraints(0, buf.Width, 0, buf.Height)
adapter.GetRenderingPipeline().Render(n.root, constraints, buf)
```

Now the layout engine receives **bounded width constraints** (MaxWidth = buffer width), enabling proper flex space distribution.

### Layout Flow

1. **framework/app.go** calls `paintable.Paint(ctx, buf)` with buffer sized to `terminalWidth` × `terminalHeight`
2. **DeclarativeNode.Paint** creates `BoxConstraints(0, buf.Width, 0, buf.Height)`
3. **RenderingPipeline.Render** passes constraints to `computeEngine.Layout()`
4. **compute.Engine** calls `LayoutNode.MeasureLayout()` with bounded constraints
5. **HStack.measureHStackLayout()** (line 112-140 of `layout_measurement.go`):
   - Detects `constraints.HasBoundedWidth() == true`
   - Calculates available width: `MaxWidth - padding - gaps`
   - Distributes space to flex children proportionally
   - Each flex child gets exact width constraints

## Result

With the fix:
- VStack receives bounded width (80 chars)
- HStack children (flex=1 each) receive constraints: `MinWidth=40, MaxWidth=40`
- Each panel is exactly 40 chars wide
- Total: 40 + 40 = 80 chars (fits screen perfectly)

## Testing

Run the demo to verify:
```bash
cd examples/ui_demos/demo1_full_featured
go run main.go
```

Expected behavior:
- Left sidebar (menu) should be ~40 characters wide
- Right content area should be ~40 characters wide
- Both panels should touch the screen edges with no gaps

## Files Modified

1. `internal/render/declarative_node.go` - Updated Paint() to use constraint-based rendering
2. `internal/render/pipeline_renderer.go` - Added GetRenderingPipeline() method
3. `internal/render/vnode_renderer.go` - Added GetPipeline() and GetRenderingPipeline() methods

## Backward Compatibility

The changes are backward compatible:
- Falls back to legacy rendering if the pipeline fails
- Old renderer implementations still work via the `VNodeRenderer` interface
- Only `PipelineRendererAdapter` gets the new constraint-based behavior
