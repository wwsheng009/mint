# Inspector Layer Investigation Summary

## Root Cause Found: SetProps/SetLayer Order Bug

**Problem**: `ElementVNode.SetProps()` REPLACES the entire props map instead of merging.

**Evidence**: When code calls `SetLayer()` then `SetProps()`, the layer property is lost:
```go
// ❌ WRONG - loses _layer property
inspectorOverlay.SetLayer(rtui.LayerInspector)
inspectorOverlay.SetProps(rtui.Props{"x": 40, "y": 5})  // Wipes out _layer!
```

**Fix**: Call `SetProps()` BEFORE `SetLayer()`:
```go
// ✅ CORRECT - preserves _layer property
inspectorOverlay.SetProps(rtui.Props{"x": 40, "y": 5})
inspectorOverlay.SetLayer(rtui.LayerInspector)
```

## Test Results

Before fix:
```
Test 2 (with Inspector): child 1: layer=0  ← Lost!
Total layers: 1  ← Only base layer
```

After fix:
```
Test 2 (with Inspector): child 1: layer=4  ← Preserved!
Total layers: 2  ← Base + Inspector layers
```

## Live Demo Debugging

### TUI_INSPECTOR=false (Working)
```
[PipelineRenderer] hasLayers=false
[PaintEngine.Paint] START: box=(0,0,80x19)  ✅ Normal layout size
```

### TUI_INSPECTOR=true (Not Working)
```
[Inspector] Overlay position set to (40, 5), layer=inspector  ✅ Layer set correctly
[PipelineRenderer] hasLayers=true  ✅ Detected layers
[PipelineRenderer] Using RenderLayers for multi-layer rendering  ✅ Multi-layer path
[StripLayers] child 1: type=*ui.BorderedNode, layer=4  ✅ Inspector layer collected
[positionInspector] inspector=(40,5) size=80x5  ✅ Inspector positioned correctly
[PaintEngine.Paint] START: box=(0,0,80x19)  ✅ Base layer painted
[PaintEngine.Paint] START: box=(40,5,80x5)  ✅ Inspector layer painted

BUT:
[Paint.paintNode] Element at (1,12) size 78x1073741823  ❌ Layout overflow!
```

## Current Status

### ✅ What's Working
1. SetProps/SetLayer order bug is fixed in test
2. Inspector layer is correctly detected and collected
3. Multi-layer rendering is triggered correctly
4. Inspector is positioned at (40, 5) as expected
5. Both base and inspector layers are painted

### ❌ What's Still Broken
1. Layout overflow: `height=1073741823` indicates layout calculation error
2. This overflow happens when multi-layer rendering is used
3. Single-layer rendering (TUI_INSPECTOR=false) has normal layout size

## Key Insight

The SetProps/SetLayer bug was in the TEST, not in the actual Inspector code. The real Inspector code already had the correct order.

The real issue is a **layout calculation error** that occurs during multi-layer rendering, causing some nodes to have MAX_INT height.

## Next Steps

1. Investigate why `LayoutNode` gets height=1073741823 in multi-layer mode
2. Check if StripLayers is corrupting the tree structure
3. Verify that CloneVNode is preserving layout properties correctly
4. Compare the tree structure between single-layer and multi-layer paths

## Files Modified

1. ✅ `runtime/layer/modal_vs_inspector_test.go` - Fixed SetProps/SetLayer order
2. ✅ `runtime/layer/collector.go` - Added debug logging (can be removed)
3. ✅ `INSPECTOR_SETLAYER_BUG_FIX.md` - Documented root cause
4. ✅ `INSPECTOR_LAYER_INVESTIGATION_SUMMARY.md` - This file
