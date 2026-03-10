# shouldUpdate Index Bug Fix

## Problem

When a button was placed with a single `ui.Text(" ")` spacer in HStack, the button would be duplicated in the UI. The same issue did not occur when two `ui.Text(" ")` elements were used.

Example that fails:
```go
ui.HStack(
    app.ButtonBuilder(" - ").Build(),  // index 0
    ui.Text(" "),                      // index 1
    app.ButtonBuilder(" + ").Build(),  // index 2 - DUPLICATED!
)
```

Example that works (two spacers):
```go
ui.HStack(
    app.ButtonBuilder(" - ").Build(),  // index 0
    ui.Text(" "),                      // index 1
    ui.Text(" "),                      // index 2
    app.ButtonBuilder(" + ").Build(),  // index 3 - correctly rendered
)
```

## Root Cause

The bug was in `shouldUpdate()` function in `internal/reconciler/diff.go`.

### Before (Buggy Code):

```go
func shouldUpdate(current *Fiber, vnode rtui.VNode) bool {
    if current == nil || vnode == nil {
        return false
    }

    // ❌ BUG: Using current.SiblingIndex (old tree position) instead of new position
    newDiffKey := normalizeDiffKey(vnode, current.SiblingIndex)

    if current.DiffKey != newDiffKey {
        return false
    }
    // ... rest of function
}
```

### The Problem:

When reconciling children, if the number of children changes (e.g., from 4 to 3):

1. **Old tree**: 4 children at indices 0, 1, 2, 3
   - Fiber(0, Button), Fiber(1, Text), Fiber(2, Text), Fiber(3, Button)
   - DiffKeys: "_idx_0", "_idx_1", "_idx_2", "_idx_3"

2. **New tree**: 3 children at indices 0, 1, 2
   - VNode(0, Button), VNode(1, Text), VNode(2, Button)

3. **Matching process**:
   - New VNode at index 2 (Button) tries to find a match
   - It checks old Fiber at index 0 (Button):
     - `newDiffKey = normalizeDiffKey(buttonVNode, current.SiblingIndex=0)` = "_idx_0"
     - `current.DiffKey` = "_idx_0"
     - **Match!** ❌ This is WRONG!
   - The Fiber at index 0 is cloned and reused for the new button at index 2
   - Meanwhile, the button at index 3 (which would have been at index 2 in the old tree) is deleted
   - Result: The same button Fiber appears twice in the UI

## Solution

Use the **new sibling index** (position in the new tree) instead of `current.SiblingIndex` (position in the old tree) when calculating DiffKey.

### After (Fixed Code):

```go
func shouldUpdate(current *Fiber, vnode rtui.VNode, newSiblingIndex int) bool {
    if current == nil || vnode == nil {
        return false
    }

    // ✅ FIX: Use newSiblingIndex (position in new tree)
    newDiffKey := normalizeDiffKey(vnode, newSiblingIndex)

    if current.DiffKey != newDiffKey {
        return false
    }
    // ... rest of function
}
```

### The Fix:

1. Modified `findMatchingChild()` to accept `newSiblingIndex` parameter
2. Modified `shouldUpdate()` to accept `newSiblingIndex` parameter and use it for DiffKey calculation
3. Updated `reconcileExistingChildren()` to pass `i` (current new child index) to `findMatchingChild()`

Now, when matching:
   - New VNode at index 2 (Button) tries to find a match
   - It checks old Fiber at index 0 (Button):
     - `newDiffKey = normalizeDiffKey(buttonVNode, newSiblingIndex=2)` = "_idx_2"
     - `current.DiffKey` = "_idx_0"
     - **No match!** ✅ Correct!
   - It checks old Fiber at index 3 (Button):
     - `newDiffKey = normalizeDiffKey(buttonVNode, newSiblingIndex=2)` = "_idx_2"
     - `current.DiffKey` = "_idx_3"
     - **No match!** ✅ Correct!
   - Result: A new Fiber is created for the button, or if there were a fiber at index 2 with matching type, it would be reused

## Impact

This fix ensures that:
1. Children are correctly matched based on their position in the **new** tree, not the old tree
2. No duplicates appear when children are reordered or removed
3. The reconciliation algorithm correctly handles dynamic child lists

## Files Changed

- `internal/reconciler/diff.go`:
  - Modified `findMatchingChild()` signature to add `newSiblingIndex int` parameter
  - Modified `shouldUpdate()` signature to add `newSiblingIndex int` parameter
  - Updated `reconcileExistingChildren()` to pass index `i` to `findMatchingChild()`

## Testing

To verify the fix works, the fiber_counter example should correctly display buttons without duplication:

```bash
set MINT_USE_FIBER=true
go run .\examples\fiber_counter\main.go
```

Expected output: Two distinct buttons, one for "+" and one for "-", each responding correctly to clicks.
