# Wrap FillWidth Fix Summary

## Problem

When using `WrapBuilder(...).FillWidth().Build()`, buttons were not stretching to fill the container width. They remained clustered on the left side instead of distributing evenly.

## Root Cause

The Wrap component was only setting `stretchCross` on the **VStack**, but not on each **HStack** (row). This caused the layout engine to skip the stretching logic for each row.

## Solution

Modified `components/layout/wrap.go` in the `Build()` method to set `stretchCross = true` on **both**:
1. The VStack (container)
2. Each HStack (row)

## Code Changes

**File**: `components/layout/wrap.go`

**Location**: Lines 290-337 (HStack creation loop)

**Change**:
```go
// Before fix: Only VStack had stretchCross
for _, row := range rows {
    hstackBuilder := &LayoutBuilder{...}
    rowNodes = append(rowNodes, hstackBuilder.Build())
}

// After fix: Both VStack and each HStack have stretchCross
fillWidth := false
if props := b.node.Props(); props != nil {
    if fw := props.GetBool("fillWidth"); fw {
        fillWidth = true
    }
}

for _, row := range rows {
    hstackBuilder := &LayoutBuilder{...}

    // Set stretchCross on each HStack when FillWidth is enabled
    if fillWidth {
        hstackBuilder.node.stretchCross = true
        hstackBuilder.node.SetProp("stretchCross", true)
        hstackBuilder.node.SetProp("gap", b.node.gap)
        hstackBuilder.node.SetProp("align", int(align))
    }

    rowNodes = append(rowNodes, hstackBuilder.Build())
}
```

## Technical Details

### Why both levels need stretchCross

1. **VStack's stretchCross**: Tells layout engine to stretch its children (the HStack rows)
2. **HStack's stretchCross**: Marks each row as eligible for stretching

Layout engine code (`runtime/compute/engine.go:992`):
```go
func (e *Engine) layoutVStack(box *ComputedBox, x, y int) {
    layoutInfo := rtui.GetLayoutInfo(box.VNode)
    stretchCross := layoutInfo.StretchCross  // VStack's stretchCross

    for _, child := range box.Children {
        childInfo := rtui.GetLayoutInfo(child.VNode)

        // Child stretches if:
        // 1. Child has flex > 0, OR
        // 2. VStack has stretchCross, OR
        // 3. Child has FillWidth
        if (childInfo.Flex > 0 || stretchCross || childInfo.FillWidth) && box.Box.Width < runtime.Infinity {
            child.Box.Width = box.Box.Width  // Stretch child
        }
    }
}
```

## Verification

### Test Results

```bash
$ go test ./components/layout/... -run TestWrap -v
=== RUN   TestWrap_BasicWrapping
--- PASS: TestWrap_BasicWrapping (0.00s)
=== RUN   TestWrap_FillWidth
--- PASS: TestWrap_FillWidth (0.00s)
...
PASS
ok      github.com/wwsheng009/mint/components/layout
```

### Structure Verification

**Without FillWidth**:
```
VStack StretchCross: false
HStack StretchCross: false
```

**With FillWidth** (after fix):
```
VStack StretchCross: true  ✅
HStack StretchCross: true  ✅
```

## Visual Result

### Before Fix
```
┌─────────────────────────────────────────────────┐
│ [1] Event [2] State [3] Render [4] Paint      │
│ Buttons clustered on left                      │
└─────────────────────────────────────────────────┘
```

### After Fix
```
┌─────────────────────────────────────────────────┐
│ [1] Event     [2] State     [3] Render         │
│ Buttons stretched to fill width                 │
└─────────────────────────────────────────────────┘
```

## Related Files

- `components/layout/wrap.go` - Fix implementation
- `docs/layout/wrap_component.md` - Updated with implementation details
- `FILLWIDTH_TRUE_FIX.md` - Detailed technical analysis
- `examples/wrap_demo/main.go` - Demo showing FillWidth behavior

## Impact

- ✅ All existing tests pass
- ✅ demo2_runtime_internals now correctly stretches buttons
- ✅ No breaking changes to API
- ✅ Documentation updated with technical details

## Status

**Status**: ✅ Fixed and verified
**Date**: 2024
**Tests**: All passing
**Examples**: Working correctly
