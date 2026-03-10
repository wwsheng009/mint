# Inspector Investigation - Complete Findings

## Summary

I've identified and fixed a critical bug, but found that the Inspector demo still has display issues due to a separate layout problem.

## Fixed: SetProps/SetLayer Order Bug

**Root Cause**: `ElementVNode.SetProps()` **replaces** the entire props map, wiping out the `_layer` property.

**The Bug**:
```go
// ❌ WRONG - loses _layer property
inspectorOverlay.SetLayer(rtui.LayerInspector)     // Sets _layer in props
inspectorOverlay.SetProps(rtui.Props{"x": 40, "y": 5}) // REPLACES props, _layer is lost!
```

**The Fix**:
```go
// ✅ CORRECT - preserves _layer property
inspectorOverlay.SetProps(rtui.Props{"x": 40, "y": 5}) // Set props first
inspectorOverlay.SetLayer(rtui.LayerInspector)        // Then set layer
```

**Proof**:
```bash
cd runtime/layer
go test -v -run TestModalVsInspector

=== Test 2 (with Inspector): hasInspector=&{(40,5,13×3)} ✅
--- PASS: TestModalVsInspector
```

## Current Issue: Layout Overflow in Multi-Layer Rendering

**Symptom**: When `TUI_INSPECTOR=true`, the interface doesn't display properly.

**Root Cause**: Inspector's Tabs component gets `height=1073741823` (MAX_INT) during multi-layer layout.

**Debug Output**:
```
TUI_INSPECTOR=false (Working):
  [PaintEngine.Paint] box=(0,0,80x19)  ✅ Normal

TUI_INSPECTOR=true (Broken):
  [PaintEngine.Paint] START: box=(40,5,80x5)  ← Inspector container
  [Paint.paintNode] Element at (41,8) size 78x1073741823  ← Tabs overflow!
```

**Analysis**:
- Inspector panel has `.Height(25)` - correct
- VStack inside should distribute 23 lines among children
- But Tabs gets MAX_INT instead of constrained height
- Only happens in multi-layer rendering, not single-layer

**Hypothesis**: The layout engine passes different constraints during multi-layer vs single-layer rendering, causing VStack to give unbounded height to flexible children.

## Status

### ✅ Fixed
- SetProps/SetLayer order in `runtime/layer/modal_vs_inspector_test.go`
- Added debug logging to `runtime/layer/collector.go`
- Layer collection now works correctly for both Modal and Inspector

### ❌ Still Broken
- Inspector demo doesn't display when `TUI_INSPECTOR=true`
- Layout overflow: Tabs height = 1073741823
- Likely issue: constraint propagation in multi-layer layout

## Next Steps

1. **Investigate constraint propagation**: Compare how `compute.Engine.Layout()` passes constraints to children in single-layer vs multi-layer mode

2. **Check VStack flex layout**: See if VStack is correctly applying parent height constraints to flexible children

3. **Workaround**: Set explicit height on Tabs component to avoid flex expansion

4. **Alternative**: Use ScrollView with fixed height for tab content

## Files Modified

1. ✅ `runtime/layer/modal_vs_inspector_test.go` - Fixed SetProps/SetLayer order
2. ✅ `runtime/layer/collector.go` - Added debug logging (lines 220-253)
3. ✅ `runtime/layer/collector.go` - Added missing imports (fmt, os)
4. ✅ `INSPECTOR_SETLAYER_BUG_FIX.md` - Root cause documentation
5. ✅ `INSPECTOR_LAYER_INVESTIGATION_SUMMARY.md` - Investigation summary
6. ✅ `INSPECTOR_INVESTIGATION_COMPLETE.md` - This file

## Recommendation

The SetProps/SetLayer bug is fixed in the test. However, the real issue is a layout engine problem where height constraints aren't being properly propagated during multi-layer rendering. This requires deeper investigation into the compute.Engine layout logic.
