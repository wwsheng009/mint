# Inspector Layer Bug - Root Cause & Fix

## Problem
When `TUI_INSPECTOR=true`, the Inspector overlay was not being collected as a separate layer, causing multi-layer rendering to fail.

## Root Cause
`ElementVNode.SetProps()` **replaces** the entire props map instead of merging:

```go
// runtime/ui/element.go:38
func (e *ElementVNode) SetProps(p Props) {
    e.props = p  // ← REPLACES, not merges!
}
```

When code calls `SetLayer()` then `SetProps()`, the `SetProps()` call **wipes out the `_layer` property** that `SetLayer()` just set:

```go
// ❌ WRONG ORDER - loses the layer property!
inspectorOverlay.SetLayer(rtui.LayerInspector)     // Sets _layer in props
inspectorOverlay.SetProps(rtui.Props{"x": 40, "y": 5}) // REPLACES props, _layer is lost!
```

## The Fix
Call `SetProps()` **BEFORE** `SetLayer()`:

```go
// ✅ CORRECT ORDER - preserves the layer property
inspectorOverlay.SetProps(rtui.Props{"x": 40, "y": 5}) // Set props first
inspectorOverlay.SetLayer(rtui.LayerInspector)        // Then set layer (modifies existing props)
```

## Proof

### Before Fix
```
Test 2 (with Inspector):
  [StripLayers]   child 1: type=*ui.BorderedNode, layer=0  ← ❌ Lost!
  Total layers: 1  ← Only base layer
  hasInspector=<nil>  ← Inspector not collected
```

### After Fix
```
Test 2 (with Inspector):
  [StripLayers]   child 1: type=*ui.BorderedNode, layer=4  ← ✅ Preserved!
  Total layers: 2  ← Base + Inspector layers
  hasInspector=&{(40,5,13×3)}  ← Inspector collected successfully
```

## Files Already Fixed
✅ `internal/inspector/standalone_inspector.go` - Already has correct order (lines 283-289)
✅ `runtime/layer/modal_vs_inspector_test.go` - Fixed in this commit

## Why Modal Worked But Inspector Didn't
The Modal test didn't call `SetProps()` after `SetLayer()`, so the layer property was preserved:

```go
modalOverlay.SetLayer(rtui.LayerModal)  // No SetProps() after, so layer is preserved
```

The Inspector test called `SetProps()` AFTER `SetLayer()`, wiping out the layer:

```go
inspectorOverlay.SetLayer(rtui.LayerInspector)  // Sets _layer
inspectorOverlay.SetProps(rtui.Props{...})      // ❌ Wipes out _layer!
```

## Architectural Issue
The real problem is that `SetProps()` replaces instead of merges. Consider:
1. Change `SetProps()` to merge with existing props
2. Or add a `SetPropsMerging()` method that preserves existing keys
3. Or document clearly that `SetProps()` replaces all props

For now, the workaround is simple: **call `SetProps()` before `SetLayer()`**.

## Test Results
```bash
cd runtime/layer
go test -v -run TestModalVsInspector

=== Test 1 (with Modal): hasModal=&{(54,18,12×3)}, hasInspector=<nil>
=== Test 2 (with Inspector): hasModal=<nil>, hasInspector=&{(40,5,13×3)}
--- PASS: TestModalVsInspector (0.00s)
```

Both Modal and Inspector are now correctly collected as separate layers.
